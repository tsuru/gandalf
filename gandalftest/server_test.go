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
	request, _ := http.NewRequest("POST", "/users", body)
	server.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusOK)
	c.Assert(server.users, check.DeepEquals, []string{"someuser"})
	c.Assert(server.keys, check.DeepEquals, map[string][]key{"someuser": []key{{Name: "rsa", Body: "mykeyrsa"}}})
}
