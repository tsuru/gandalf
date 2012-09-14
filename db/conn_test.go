package db

import (
	"labix.org/v2/mgo"
	. "launchpad.net/gocheck"
	"testing"
)

func Test(t *testing.T) { TestingT(t) }

type S struct{}

var _ = Suite(&S{})

func (s *S) TestSessionRepositoryShouldReturnAMongoCollection(c *C) {
	var rep interface{}
	rep = Session.Repository()
	_, ok := rep.(*mgo.Collection)
	c.Assert(ok, Equals, true)
}

func (s *S) TestSessionUserShouldReturnAMongoCollection(c *C) {
	var usr interface{}
	usr = Session.User()
	_, ok := usr.(*mgo.Collection)
	c.Assert(ok, Equals, true)
}
