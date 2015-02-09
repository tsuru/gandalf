// Copyright 2015 gandalf authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gandalftest

import (
	"net"
	"testing"

	"launchpad.net/gocheck"
)

func Test(t *testing.T) { gocheck.TestingT(t) }

type S struct{}

var _ = gocheck.Suite(&S{})

func (s *S) TestNewServerFreePort(c *gocheck.C) {
	server, err := NewServer("127.0.0.1:0")
	c.Assert(err, gocheck.IsNil)
	defer server.Stop()
	conn, err := net.Dial("tcp", server.listener.Addr().String())
	c.Assert(err, gocheck.IsNil)
	c.Assert(conn.Close(), gocheck.IsNil)
}

func (s *S) TestNewServerSpecificPort(c *gocheck.C) {
	server, err := NewServer("127.0.0.1:8599")
	c.Assert(err, gocheck.IsNil)
	defer server.Stop()
	conn, err := net.Dial("tcp", server.listener.Addr().String())
	c.Assert(err, gocheck.IsNil)
	c.Assert(conn.Close(), gocheck.IsNil)
}

func (s *S) TestNewServerListenError(c *gocheck.C) {
	listen, err := net.Listen("tcp", "127.0.0.1:0")
	c.Assert(err, gocheck.IsNil)
	defer listen.Close()
	server, err := NewServer(listen.Addr().String())
	c.Assert(err, gocheck.ErrorMatches, `^.*bind: address already in use$`)
	c.Assert(server, gocheck.IsNil)
}

func (s *S) TestServerStop(c *gocheck.C) {
	server, err := NewServer("127.0.0.1:0")
	c.Assert(err, gocheck.IsNil)
	err = server.Stop()
	c.Assert(err, gocheck.IsNil)
	_, err = net.Dial("tcp", server.listener.Addr().String())
	c.Assert(err, gocheck.ErrorMatches, `^.*connection refused$`)
}

func (s *S) TestURL(c *gocheck.C) {
	server, err := NewServer("127.0.0.1:0")
	c.Assert(err, gocheck.IsNil)
	defer server.Stop()
	expected := "http://" + server.listener.Addr().String() + "/"
	c.Assert(server.URL(), gocheck.Equals, expected)
}
