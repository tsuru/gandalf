package api

import (
	"github.com/globocom/commandmocker"
	"github.com/globocom/config"
	"github.com/globocom/gandalf/db"
	"github.com/globocom/gandalf/user"
	"labix.org/v2/mgo/bson"
	. "launchpad.net/gocheck"
	"testing"
)

func Test(t *testing.T) { TestingT(t) }

type S struct {
	tmpdir string
}

var _ = Suite(&S{})

func (s *S) SetUpSuite(c *C) {
	err := config.ReadConfigFile("../etc/gandalf.conf")
	c.Assert(err, IsNil)
	s.tmpdir, err = commandmocker.Add("git", "")
	c.Assert(err, IsNil)
}

func (s *S) TearDownSuite(c *C) {
	commandmocker.Remove(s.tmpdir)
}

func (s *S) TestGetUserOr404(c *C) {
	u := user.User{Name: "umi"}
	err := db.Session.User().Insert(&u)
	c.Assert(err, IsNil)
	defer db.Session.User().Remove(bson.M{"_id": u.Name})
	rUser, err := getUserOr404("umi")
	c.Assert(err, IsNil)
	c.Assert(rUser.Name, Equals, "umi")
}

func (s *S) TestGetUserOr404ShouldReturn404WhenUserDoesntExists(c *C) {
	_, e := getUserOr404("umi")
	expected := "User umi not found"
	got := e.Error()
	c.Assert(got, Equals, expected)
}
