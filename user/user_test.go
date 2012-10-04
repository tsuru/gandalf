package user

import (
	"github.com/globocom/gandalf/db"
	"labix.org/v2/mgo/bson"
	. "launchpad.net/gocheck"
	"testing"
)

func Test(t *testing.T) { TestingT(t) }

type S struct{}

var _ = Suite(&S{})

func (s *S) TestNewUserReturnsAStructFilled(c *C) {
	u, err := New("someuser", []string{"id_rsa someKeyChars"})
	c.Assert(err, IsNil)
	defer db.Session.User().Remove(bson.M{"_id": u.Name})
	c.Assert(u.Name, Equals, "someuser")
	c.Assert(len(u.Keys), Not(Equals), 0)
}

func (s *S) TestNewUserShouldStoreUserInDatabase(c *C) {
	u, err := New("someuser", []string{"id_rsa someKeyChars"})
	c.Assert(err, IsNil)
	defer db.Session.User().Remove(bson.M{"_id": u.Name})
	err = db.Session.User().Find(bson.M{"_id": u.Name}).One(&u)
	c.Assert(err, IsNil)
	c.Assert(u.Name, Equals, "someuser")
	c.Assert(len(u.Keys), Not(Equals), 0)
}

func (s *S) TestNewChecksIfUserIsValidBeforeStoring(c *C) {
	_, err := New("", []string{})
	c.Assert(err, NotNil)
	got := err.Error()
	expected := "Validation Error: user name is not valid"
	c.Assert(got, Equals, expected)
}

func (s *S) TestIsValidReturnsErrorWhenUserDoesNotHaveAName(c *C) {
	u := User{Keys: []string{"id_rsa foooBar"}}
	v, err := u.isValid()
	c.Assert(v, Equals, false)
	c.Assert(err, NotNil)
	expected := "Validation Error: user name is not valid"
	got := err.Error()
	c.Assert(got, Equals, expected)
}

func (s *S) TestIsValidShouldNotAcceptEmptyUserName(c *C) {
	u := User{Keys: []string{"id_rsa foooBar"}}
	v, err := u.isValid()
	c.Assert(v, Equals, false)
	c.Assert(err, NotNil)
	expected := "Validation Error: user name is not valid"
	got := err.Error()
	c.Assert(got, Equals, expected)
}

func (s *S) TestIsValidShouldAcceptEmailsAsUserName(c *C) {
	u := User{Name: "r2d2@gmail.com", Keys: []string{"id_rsa foooBar"}}
	v, err := u.isValid()
	c.Assert(err, IsNil)
	c.Assert(v, Equals, true)
}

func (s *S) TestRemove(c *C) {
	u, err := New("someuser", []string{})
	c.Assert(err, IsNil)
	err = Remove(u)
	c.Assert(err, IsNil)
	lenght, err := db.Session.User().Find(bson.M{"_id": u.Name}).Count()
	c.Assert(err, IsNil)
	c.Assert(lenght, Equals, 0)
}
