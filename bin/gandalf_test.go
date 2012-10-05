package main

import (
	"github.com/globocom/gandalf/db"
	"github.com/globocom/gandalf/repository"
	"github.com/globocom/gandalf/user"
	"labix.org/v2/mgo/bson"
	. "launchpad.net/gocheck"
	"testing"
)

func Test(t *testing.T) { TestingT(t) }

type S struct {
	user *user.User
	repo *repository.Repository
}

var _ = Suite(&S{})

func (s *S) SetUpSuite(c *C) {
	var err error
	s.user, err = user.New("testuser", []string{})
	c.Check(err, IsNil)
	// does not uses repository.New to avoid creation of bare git repo
	s.repo = &repository.Repository{Name: "myapp", Users: []string{s.user.Name}}
	err = db.Session.Repository().Insert(s.repo)
	c.Check(err, IsNil)
}

func (s *S) TearDownSuite(c *C) {
	db.Session.User().Remove(bson.M{"_id": s.user.Name})
	db.Session.Repository().Remove(bson.M{"_id": s.repo.Name})
}

func (s *S) TestHasWritePermissionSholdReturnTrueWhenUserCanWriteInRepo(c *C) {
	allowed := hasWritePermission(s.user, s.repo)
	c.Assert(allowed, Equals, true)
}

func (s *S) TestHasWritePermissionShouldReturnFalseWhenUserCannotWriteinRepo(c *C) {
	r := &repository.Repository{Name: "myotherapp"}
	db.Session.Repository().Insert(&r)
	defer db.Session.Repository().Remove(bson.M{"_id": r.Name})
	allowed := hasWritePermission(s.user, r)
	c.Assert(allowed, Equals, false)
}

func (s *S) TestHasReadPermissionShouldReturnTrueWhenRepositoryIsPublic(c *C) {
	r := &repository.Repository{Name: "myotherapp", IsPublic: true}
	db.Session.Repository().Insert(&r)
	defer db.Session.Repository().Remove(bson.M{"_id": r.Name})
	allowed := hasReadPermission(s.user, r)
	c.Assert(allowed, Equals, true)
}

func (s *S) TestHasReadPermissionShouldReturnTrueWhenRepositoryIsNotPublicAndUserHasPermissionToReadAndWrite(c *C) {
	allowed := hasReadPermission(s.user, s.repo)
	c.Assert(allowed, Equals, true)
}

func (s *S) TestHasReadPermissionShouldReturnFalseWhenUserDoesNotHavePermissionToReadWriteAndRepoIsNotPublic(c *C) {
	r := &repository.Repository{Name: "myotherapp", IsPublic: false}
	db.Session.Repository().Insert(&r)
	defer db.Session.Repository().Remove(bson.M{"_id": r.Name})
	allowed := hasReadPermission(s.user, r)
	c.Assert(allowed, Equals, false)
}
