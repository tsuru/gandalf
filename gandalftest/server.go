// Copyright 2015 gandalf authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package gandalftest provides a fake implementation of the Gandalf API.
package gandalftest

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"

	"github.com/gorilla/pat"
	"github.com/tsuru/gandalf/repository"
	"golang.org/x/crypto/ssh"
)

type user struct {
	Name string
	Keys map[string]string
}

type key struct {
	Name string
	Body string
}

// Failure represents a prepared failure, that is used in the PrepareFailure
// method.
type Failure struct {
	Method   string
	Path     string
	Response string
}

// GandalfServer is a fake gandalf server. An instance of the client can be
// pointed to the address generated for this server
type GandalfServer struct {
	listener  net.Listener
	muxer     *pat.Router
	users     []string
	keys      map[string][]key
	repos     []repository.Repository
	usersLock sync.RWMutex
	repoLock  sync.RWMutex
	failures  chan Failure
}

// NewServer returns an instance of the test server, bound to the specified
// address. To get a random port, users can specify the :0 port.
//
// Examples:
//
//     server, err := NewServer("127.0.0.1:8080") // will bind on port 8080
//     server, err := NewServer("127.0.0.1:0") // will get a random available port
func NewServer(bind string) (*GandalfServer, error) {
	listener, err := net.Listen("tcp", bind)
	if err != nil {
		return nil, err
	}
	server := GandalfServer{
		listener: listener,
		keys:     make(map[string][]key),
		failures: make(chan Failure, 1),
	}
	server.buildMuxer()
	go http.Serve(listener, &server)
	return &server, nil
}

// Stop stops the server, cleaning the internal listener and freeing the
// allocated port.
func (s *GandalfServer) Stop() error {
	return s.listener.Close()
}

// URL returns the URL of the server, in the format "http://<host>:<port>/".
func (s *GandalfServer) URL() string {
	return fmt.Sprintf("http://%s/", s.listener.Addr())
}

// PrepareFailure prepares a failure in the server. The next request matching
// the given URL and request path will fail with a 500 code and the provided
// response in the body.
func (s *GandalfServer) PrepareFailure(failure Failure) {
	s.failures <- failure
}

// ServeHTTP handler HTTP requests, dealing with prepared failures before
// dispatching the request to the proper internal handler.
func (s *GandalfServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if failure, ok := s.getFailure(r.Method, r.URL.Path); ok {
		http.Error(w, failure.Response, http.StatusInternalServerError)
		return
	}
	s.muxer.ServeHTTP(w, r)
}

func (s *GandalfServer) buildMuxer() {
	s.muxer = pat.New()
	s.muxer.Post("/user/{name}/key", http.HandlerFunc(s.addKeys))
	s.muxer.Post("/user", http.HandlerFunc(s.createUser))
	s.muxer.Delete("/user/{name}", http.HandlerFunc(s.removeUser))
	s.muxer.Post("/repository", http.HandlerFunc(s.createRepository))
	s.muxer.Delete("/repository/{name}", http.HandlerFunc(s.removeRepository))
	s.muxer.Get("/repository/{name}", http.HandlerFunc(s.getRepository))
}

func (s *GandalfServer) createUser(w http.ResponseWriter, r *http.Request) {
	s.usersLock.Lock()
	defer s.usersLock.Unlock()
	defer r.Body.Close()
	var usr user
	err := json.NewDecoder(r.Body).Decode(&usr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	s.users = append(s.users, usr.Name)
	if _, ok := s.keys[usr.Name]; ok {
		http.Error(w, "user already exists", http.StatusConflict)
		return
	}
	keys := make([]key, 0, len(usr.Keys))
	for name, body := range usr.Keys {
		keys = append(keys, key{Name: name, Body: body})
	}
	s.keys[usr.Name] = keys
}

func (s *GandalfServer) removeUser(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get(":name")
	_, index := s.findUser(username)
	if index < 0 {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	s.usersLock.Lock()
	defer s.usersLock.Unlock()
	last := len(s.users) - 1
	s.users[index] = s.users[last]
	s.users = s.users[:last]
	delete(s.keys, username)
}

func (s *GandalfServer) createRepository(w http.ResponseWriter, r *http.Request) {
	var repo repository.Repository
	defer r.Body.Close()
	err := json.NewDecoder(r.Body).Decode(&repo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	users := append(repo.Users, repo.ReadOnlyUsers...)
	for _, username := range users {
		_, index := s.findUser(username)
		if index < 0 {
			http.Error(w, fmt.Sprintf("user %q not found", username), http.StatusBadRequest)
			return
		}
	}
	s.repoLock.Lock()
	defer s.repoLock.Unlock()
	for _, r := range s.repos {
		if r.Name == repo.Name {
			http.Error(w, "repository already exists", http.StatusConflict)
			return
		}
	}
	s.repos = append(s.repos, repo)
}

func (s *GandalfServer) removeRepository(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get(":name")
	_, index := s.findRepository(name)
	if index < 0 {
		http.Error(w, "repository not found", http.StatusNotFound)
		return
	}
	s.repoLock.Lock()
	defer s.repoLock.Unlock()
	last := len(s.repos) - 1
	s.repos[index] = s.repos[last]
	s.repos = s.repos[:last]
}

func (s *GandalfServer) getRepository(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get(":name")
	repo, index := s.findRepository(name)
	if index < 0 {
		http.Error(w, "repository not found", http.StatusNotFound)
		return
	}
	err := json.NewEncoder(w).Encode(repo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *GandalfServer) addKeys(w http.ResponseWriter, r *http.Request) {
	username := r.URL.Query().Get(":name")
	var keys map[string]string
	defer r.Body.Close()
	err := json.NewDecoder(r.Body).Decode(&keys)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	s.usersLock.Lock()
	defer s.usersLock.Unlock()
	userKeys, ok := s.keys[username]
	if !ok {
		http.Error(w, "user not found", http.StatusNotFound)
		return
	}
	for name, body := range keys {
		if _, _, _, _, err := ssh.ParseAuthorizedKey([]byte(body)); err != nil {
			http.Error(w, fmt.Sprintf("key %q is not valid", name), http.StatusBadRequest)
			return
		}
		for _, userKey := range userKeys {
			if name == userKey.Name {
				http.Error(w, fmt.Sprintf("key %q already exists", name), http.StatusConflict)
				return
			}
		}
	}
	for name, body := range keys {
		userKeys = append(userKeys, key{Name: name, Body: body})
	}
	s.keys[username] = userKeys
}

func (s *GandalfServer) findUser(name string) (username string, index int) {
	s.usersLock.RLock()
	defer s.usersLock.RUnlock()
	for i, user := range s.users {
		if user == name {
			return user, i
		}
	}
	return "", -1
}

func (s *GandalfServer) findRepository(name string) (repository.Repository, int) {
	s.repoLock.RLock()
	defer s.repoLock.RUnlock()
	for i, repo := range s.repos {
		if repo.Name == name {
			return repo, i
		}
	}
	return repository.Repository{}, -1
}

func (s *GandalfServer) getFailure(method, path string) (Failure, bool) {
	var f Failure
	select {
	case f = <-s.failures:
		if f.Method == method && f.Path == path {
			return f, true
		}
		s.failures <- f
		return f, false
	default:
		return f, false
	}
}
