// Copyright 2013 gandalf authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/globocom/gandalf/repository"
	"github.com/globocom/gandalf/user"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"regexp"
)

var re = regexp.MustCompile(`^User ".*" not found$`)

func accessParameters(body io.ReadCloser) (repositories, users []string, err error) {
	var params map[string][]string
	if err := parseBody(body, &params); err != nil {
		return []string{}, []string{}, err
	}
	users, ok := params["users"]
	if !ok {
		return []string{}, []string{}, errors.New("It is need a user list")
	}
	repositories, ok = params["repositories"]
	if !ok {
		return []string{}, []string{}, errors.New("It is need a repository list")
	}
	return repositories, users, nil
}

func GrantAccess(w http.ResponseWriter, r *http.Request) {
	// TODO: update README
	repositories, users, err := accessParameters(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := repository.GrantAccess(repositories, users); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	fmt.Fprintf(w, "Successfully granted access to users \"%s\" into repository \"%s\"", users, repositories)
}

func RevokeAccess(w http.ResponseWriter, r *http.Request) {
	repositories, users, err := accessParameters(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := repository.RevokeAccess(repositories, users); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "Successfully revoked access to users \"%s\" into repositories \"%s\"", users, repositories)
}

func AddKey(w http.ResponseWriter, r *http.Request) {
	keys := map[string]string{}
	if err := parseBody(r.Body, &keys); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(keys) == 0 {
		http.Error(w, "A key is needed", http.StatusBadRequest)
		return
	}
	uName := r.URL.Query().Get(":name")
	if err := user.AddKey(uName, keys); err != nil {
		status := http.StatusNotFound
		if !re.MatchString(err.Error()) {
			status = http.StatusConflict
		}
		http.Error(w, err.Error(), status)
		return
	}
	fmt.Fprint(w, "Key(s) successfully created")
}

func RemoveKey(w http.ResponseWriter, r *http.Request) {
	uName := r.URL.Query().Get(":name")
	kName := r.URL.Query().Get(":keyname")
	if err := user.RemoveKey(uName, kName); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	fmt.Fprintf(w, "Key \"%s\" successfully removed", kName)
}

func NewUser(w http.ResponseWriter, r *http.Request) {
	var usr user.User
	if err := parseBody(r.Body, &usr); err != nil {
		http.Error(w, "Got error while parsing body: "+err.Error(), http.StatusBadRequest)
		return
	}
	u, err := user.New(usr.Name, usr.Keys)
	if err != nil {
		http.Error(w, "Got error while creating user: "+err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Fprintf(w, "User \"%s\" successfully created\n", u.Name)
}

func RemoveUser(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get(":name")
	if err := user.Remove(name); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Fprintf(w, "User \"%s\" successfully removed\n", name)
}

func NewRepository(w http.ResponseWriter, r *http.Request) {
	var repo repository.Repository
	if err := parseBody(r.Body, &repo); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	rep, err := repository.New(repo.Name, repo.Users, repo.IsPublic)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Fprintf(w, "Repository \"%s\" successfully created\n", rep.Name)
}

func GetRepository(w http.ResponseWriter, r *http.Request) {
	repo, _ := repository.Get(r.URL.Query().Get(":name"))
	out, _ := json.Marshal(&repo)
	w.Write(out)
}

func RemoveRepository(w http.ResponseWriter, r *http.Request) {
	repo := &repository.Repository{Name: r.URL.Query().Get(":name")}
	if err := repository.Remove(repo); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Fprintf(w, "Repository \"%s\" successfully removed\n", repo.Name)
}

func parseBody(body io.ReadCloser, result interface{}) error {
	if reflect.ValueOf(result).Kind() == reflect.Struct {
		return errors.New("parseBody function cannot deal with struct. Use pointer")
	}
	b, err := ioutil.ReadAll(body)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, &result); err != nil {
		return errors.New(fmt.Sprintf("Could not parse json: %s", err.Error()))
	}
	return nil
}
