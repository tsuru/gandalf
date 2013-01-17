// Copyright 2013 gandalf authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package db

import (
	"github.com/globocom/config"
	"labix.org/v2/mgo"
	. "launchpad.net/gocheck"
	"testing"
)

func Test(t *testing.T) { TestingT(t) }

type S struct{}

var _ = Suite(&S{})

func (s *S) SetUpSuite(c *C) {
	config.Set("database:url", "127.0.0.1:27017")
	config.Set("database:name", "gandalf_tests")
	Connect()
}

func (s *S) TestSessionRepositoryShouldReturnAMongoCollection(c *C) {
	var rep *mgo.Collection
	rep = Session.Repository()
	cRep := Session.DB.C("repository")
	c.Assert(rep, DeepEquals, cRep)
}

func (s *S) TestSessionUserShouldReturnAMongoCollection(c *C) {
	var usr *mgo.Collection
	usr = Session.User()
	cUsr := Session.DB.C("user")
	c.Assert(usr, DeepEquals, cUsr)
}

func (s *S) TestConnect(c *C) {
	Connect()
	c.Assert(Session.DB.Name, Equals, "gandalf_tests")
	err := Session.DB.Session.Ping()
	c.Assert(err, IsNil)
}
