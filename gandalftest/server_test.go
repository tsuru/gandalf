// Copyright 2015 gandalf authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gandalftest

import (
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/tsuru/gandalf/repository"
	"gopkg.in/check.v1"
)

func Test(t *testing.T) { check.TestingT(t) }

type S struct{}

var _ = check.Suite(&S{})

func (s *S) TestNewServerFreePort(c *check.C) {
	server, err := NewServer("127.0.0.1:0")
	c.Assert(err, check.IsNil)
	defer server.Stop()
	conn, err := net.Dial("tcp", server.listener.Addr().String())
	c.Assert(err, check.IsNil)
	c.Assert(conn.Close(), check.IsNil)
}

func (s *S) TestNewServerSpecificPort(c *check.C) {
	server, err := NewServer("127.0.0.1:8599")
	c.Assert(err, check.IsNil)
	defer server.Stop()
	conn, err := net.Dial("tcp", server.listener.Addr().String())
	c.Assert(err, check.IsNil)
	c.Assert(conn.Close(), check.IsNil)
}

func (s *S) TestNewServerListenError(c *check.C) {
	listen, err := net.Listen("tcp", "127.0.0.1:0")
	c.Assert(err, check.IsNil)
	defer listen.Close()
	server, err := NewServer(listen.Addr().String())
	c.Assert(err, check.ErrorMatches, `^.*bind: address already in use$`)
	c.Assert(server, check.IsNil)
}

func (s *S) TestServerStop(c *check.C) {
	server, err := NewServer("127.0.0.1:0")
	c.Assert(err, check.IsNil)
	err = server.Stop()
	c.Assert(err, check.IsNil)
	_, err = net.Dial("tcp", server.listener.Addr().String())
	c.Assert(err, check.ErrorMatches, `^.*connection refused$`)
}

func (s *S) TestURL(c *check.C) {
	server, err := NewServer("127.0.0.1:0")
	c.Assert(err, check.IsNil)
	defer server.Stop()
	expected := "http://" + server.listener.Addr().String() + "/"
	c.Assert(server.URL(), check.Equals, expected)
}

func (s *S) TestCreateUser(c *check.C) {
	server, err := NewServer("127.0.0.1:0")
	c.Assert(err, check.IsNil)
	defer server.Stop()
	recorder := httptest.NewRecorder()
	body := strings.NewReader(`{"Name":"someuser","Keys":{"rsa":"mykeyrsa"}}`)
	request, _ := http.NewRequest("POST", "/user", body)
	server.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusOK)
	c.Assert(server.users, check.DeepEquals, []string{"someuser"})
	c.Assert(server.keys, check.DeepEquals, map[string][]key{"someuser": {{Name: "rsa", Body: "mykeyrsa"}}})
}

func (s *S) TestCreateDuplicateUser(c *check.C) {
	server, err := NewServer("127.0.0.1:0")
	c.Assert(err, check.IsNil)
	defer server.Stop()
	recorder := httptest.NewRecorder()
	body := strings.NewReader(`{"Name":"someuser","Keys":{"rsa":"mykeyrsa"}}`)
	request, _ := http.NewRequest("POST", "/user", body)
	server.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusOK)
	recorder = httptest.NewRecorder()
	body = strings.NewReader(`{"Name":"someuser","Keys":{"rsa":"mykeyrsa"}}`)
	request, _ = http.NewRequest("POST", "/user", body)
	server.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusConflict)
	c.Assert(recorder.Body.String(), check.Equals, "user already exists\n")
}

func (s *S) TestRemoveUser(c *check.C) {
	server, err := NewServer("127.0.0.1:0")
	c.Assert(err, check.IsNil)
	defer server.Stop()
	recorder := httptest.NewRecorder()
	body := strings.NewReader(`{"Name":"someuser","Keys":{"rsa":"mykeyrsa"}}`)
	request, _ := http.NewRequest("POST", "/user", body)
	server.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusOK)
	recorder = httptest.NewRecorder()
	request, _ = http.NewRequest("DELETE", "/user/someuser", nil)
	server.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusOK)
	c.Assert(server.users, check.HasLen, 0)
	c.Assert(server.keys, check.HasLen, 0)
}

func (s *S) TestRemoveUserNotFound(c *check.C) {
	server, err := NewServer("127.0.0.1:0")
	c.Assert(err, check.IsNil)
	defer server.Stop()
	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("DELETE", "/user/someuser", nil)
	server.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusNotFound)
	c.Assert(recorder.Body.String(), check.Equals, "user not found\n")
}

func (s *S) TestCreateRepository(c *check.C) {
	server, err := NewServer("127.0.0.1:0")
	c.Assert(err, check.IsNil)
	defer server.Stop()
	server.users = []string{"user1", "user2", "user3"}
	recorder := httptest.NewRecorder()
	body := strings.NewReader(`{"Name":"myrepo","Users":["user1","user2"],"ReadOnlyUsers":["user3"],"IsPublic":true}`)
	request, _ := http.NewRequest("POST", "/repository", body)
	server.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusOK)
	c.Assert(server.repos, check.HasLen, 1)
	c.Assert(server.repos[0].Name, check.Equals, "myrepo")
	c.Assert(server.repos[0].Users, check.DeepEquals, []string{"user1", "user2"})
	c.Assert(server.repos[0].ReadOnlyUsers, check.DeepEquals, []string{"user3"})
	c.Assert(server.repos[0].IsPublic, check.Equals, true)
}

func (s *S) TestCreateRepositoryDuplicateName(c *check.C) {
	server, err := NewServer("127.0.0.1:0")
	c.Assert(err, check.IsNil)
	defer server.Stop()
	server.repos = []repository.Repository{{Name: "somerepo"}}
	recorder := httptest.NewRecorder()
	body := strings.NewReader(`{"Name":"somerepo","IsPublic":false}`)
	request, _ := http.NewRequest("POST", "/repository", body)
	server.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusConflict)
	c.Assert(recorder.Body.String(), check.Equals, "repository already exists\n")
}

func (s *S) TestCreateRepositoryUserNotFound(c *check.C) {
	server, err := NewServer("127.0.0.1:0")
	c.Assert(err, check.IsNil)
	defer server.Stop()
	recorder := httptest.NewRecorder()
	body := strings.NewReader(`{"Name":"myrepo","Users":["user1","user2"],"ReadOnlyUsers":["user3"],"IsPublic":true}`)
	request, _ := http.NewRequest("POST", "/repository", body)
	server.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusBadRequest)
	c.Assert(recorder.Body.String(), check.Equals, `user "user1" not found`+"\n")
}

func (s *S) TestPrepareFailure(c *check.C) {
	server, err := NewServer("127.0.0.1:0")
	c.Assert(err, check.IsNil)
	defer server.Stop()
	server.PrepareFailure(Failure{Method: "POST", Path: "/users", Response: "fatal error"})
	recorder := httptest.NewRecorder()
	body := strings.NewReader(`{"Name":"someuser","Keys":{"rsa":"mykeyrsa"}}`)
	request, _ := http.NewRequest("POST", "/users", body)
	server.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusInternalServerError)
	c.Assert(recorder.Body.String(), check.Equals, "fatal error\n")
}

func (s *S) TestPrepareFailureNotMatching(c *check.C) {
	server, err := NewServer("127.0.0.1:0")
	c.Assert(err, check.IsNil)
	defer server.Stop()
	server.PrepareFailure(Failure{Method: "GET", Path: "/users", Response: "fatal error"})
	recorder := httptest.NewRecorder()
	body := strings.NewReader(`{"Name":"someuser","Keys":{"rsa":"mykeyrsa"}}`)
	request, _ := http.NewRequest("POST", "/users", body)
	server.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusOK)
}
