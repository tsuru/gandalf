// Copyright 2013 gandalf authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package repository

import (
	"fmt"
	"github.com/globocom/commandmocker"
	"github.com/globocom/config"
	"github.com/globocom/gandalf/db"
	"github.com/globocom/gandalf/fs"
	fstesting "github.com/globocom/tsuru/fs/testing"
	"labix.org/v2/mgo/bson"
	. "launchpad.net/gocheck"
	"path"
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
	config.Set("database:url", "127.0.0.1:27017")
	config.Set("database:name", "gandalf_repository_tests")
	db.Connect()
}

func (s *S) TearDownSuite(c *C) {
	db.Session.DB.DropDatabase()
}

func (s *S) TestNewShouldCreateANewRepository(c *C) {
	tmpdir, err := commandmocker.Add("git", "$*")
	c.Assert(err, IsNil)
	defer commandmocker.Remove(tmpdir)
	users := []string{"smeagol", "saruman"}
	r, err := New("myRepo", users, true)
	c.Assert(err, IsNil)
	defer db.Session.Repository().Remove(bson.M{"_id": "myRepo"})
	c.Assert(r.Name, Equals, "myRepo")
	c.Assert(r.Users, DeepEquals, users)
	c.Assert(r.IsPublic, Equals, true)
}

func (s *S) TestNewShouldRecordItOnDatabase(c *C) {
	tmpdir, err := commandmocker.Add("git", "$*")
	c.Assert(err, IsNil)
	defer commandmocker.Remove(tmpdir)
	r, err := New("someRepo", []string{"smeagol"}, true)
	defer db.Session.Repository().Remove(bson.M{"_id": "someRepo"})
	c.Assert(err, IsNil)
	err = db.Session.Repository().Find(bson.M{"_id": "someRepo"}).One(&r)
	c.Assert(err, IsNil)
	c.Assert(r.Name, Equals, "someRepo")
	c.Assert(r.Users, DeepEquals, []string{"smeagol"})
	c.Assert(r.IsPublic, Equals, true)
}

func (s *S) TestNewBreaksOnValidationError(c *C) {
	_, err := New("", []string{"smeagol"}, false)
	c.Check(err, NotNil)
	expected := "Validation Error: repository name is not valid"
	got := err.Error()
	c.Assert(got, Equals, expected)
}

func (s *S) TestRepositoryIsNotValidWithoutAName(c *C) {
	r := Repository{Users: []string{"gollum"}, IsPublic: true}
	v, err := r.isValid()
	c.Assert(v, Equals, false)
	c.Check(err, NotNil)
	got := err.Error()
	expected := "Validation Error: repository name is not valid"
	c.Assert(got, Equals, expected)
}

func (s *S) TestRepositoryIsNotValidWithInvalidName(c *C) {
	r := Repository{Name: "foo bar", Users: []string{"gollum"}, IsPublic: true}
	v, err := r.isValid()
	c.Assert(v, Equals, false)
	c.Check(err, NotNil)
	got := err.Error()
	expected := "Validation Error: repository name is not valid"
	c.Assert(got, Equals, expected)
}

func (s *S) TestRepositoryShoudBeInvalidWIthoutAnyUsers(c *C) {
	r := Repository{Name: "foo_bar", Users: []string{}, IsPublic: true}
	v, err := r.isValid()
	c.Assert(v, Equals, false)
	c.Assert(err, NotNil)
	got := err.Error()
	expected := "Validation Error: repository should have at least one user"
	c.Assert(got, Equals, expected)
}

func (s *S) TestRepositoryShouldBeValidWithoutIsPublic(c *C) {
	r := Repository{Name: "someName", Users: []string{"smeagol"}}
	v, _ := r.isValid()
	c.Assert(v, Equals, true)
}

func (s *S) TestNewShouldCreateNewGitBareRepository(c *C) {
	tmpdir, err := commandmocker.Add("git", "$*")
	c.Assert(err, IsNil)
	defer commandmocker.Remove(tmpdir)
	_, err = New("myRepo", []string{"pumpkin"}, true)
	c.Assert(err, IsNil)
	defer db.Session.Repository().Remove(bson.M{"_id": "myRepo"})
	c.Assert(commandmocker.Ran(tmpdir), Equals, true)
}

func (s *S) TestNewShouldNotStoreRepoInDbWhenBareCreationFails(c *C) {
	dir, err := commandmocker.Error("git", "", 1)
	c.Check(err, IsNil)
	defer commandmocker.Remove(dir)
	r, err := New("myRepo", []string{"pumpkin"}, true)
	c.Check(err, NotNil)
	err = db.Session.Repository().Find(bson.M{"_id": r.Name}).One(&r)
	c.Assert(err, ErrorMatches, "^not found$")
}

func (s *S) TestRemoveShouldRemoveBareRepositoryFromFileSystem(c *C) {
	tmpdir, err := commandmocker.Add("git", "$*")
	c.Assert(err, IsNil)
	defer commandmocker.Remove(tmpdir)
	rfs := &fstesting.RecordingFs{FileContent: "foo"}
	fs.Fsystem = rfs
	defer func() { fs.Fsystem = nil }()
	r, err := New("myRepo", []string{"pumpkin"}, false)
	c.Assert(err, IsNil)
	err = Remove(r)
	c.Assert(err, IsNil)
	action := "removeall " + path.Join(bareLocation(), "myRepo.git")
	c.Assert(rfs.HasAction(action), Equals, true)
}

func (s *S) TestRemoveShouldRemoveRepositoryFromDatabase(c *C) {
	tmpdir, err := commandmocker.Add("git", "$*")
	c.Assert(err, IsNil)
	defer commandmocker.Remove(tmpdir)
	rfs := &fstesting.RecordingFs{FileContent: "foo"}
	fs.Fsystem = rfs
	defer func() { fs.Fsystem = nil }()
	r, err := New("myRepo", []string{"pumpkin"}, false)
	c.Assert(err, IsNil)
	err = Remove(r)
	c.Assert(err, IsNil)
	err = db.Session.Repository().Find(bson.M{"_id": r.Name}).One(&r)
	c.Assert(err, ErrorMatches, "^not found$")
}

func (s *S) TestRemoveShouldReturnMeaningfulErrorWhenRepositoryDoesNotExistsInDatabase(c *C) {
	rfs := &fstesting.RecordingFs{FileContent: "foo"}
	fs.Fsystem = rfs
	defer func() { fs.Fsystem = nil }()
	r := &Repository{Name: "fooBar"}
	err := Remove(r)
	c.Assert(err, ErrorMatches, "^Could not remove repository: not found$")
}

func (s *S) TestRemoteShouldFormatAndReturnTheGitRemote(c *C) {
	host, err := config.GetString("host")
	c.Assert(err, IsNil)
	remote := (&Repository{Name: "lol"}).Remote()
	c.Assert(remote, Equals, fmt.Sprintf("git@%s:lol.git", host))
}

func (s *S) TestRemoteShouldUseUidFromConfigFile(c *C) {
	uid, err := config.GetString("uid")
	c.Assert(err, IsNil)
	host, err := config.GetString("host")
	c.Assert(err, IsNil)
	config.Set("uid", "test")
	defer config.Set("uid", uid)
	remote := (&Repository{Name: "f#"}).Remote()
	c.Assert(remote, Equals, fmt.Sprintf("test@%s:f#.git", host))
}

func (s *S) TestGrantAccessShouldAddUserToListOfRepositories(c *C) {
	tmpdir, err := commandmocker.Add("git", "$*")
	c.Assert(err, IsNil)
	defer commandmocker.Remove(tmpdir)
	r, err := New("proj1", []string{"someuser"}, true)
	c.Assert(err, IsNil)
	defer db.Session.Repository().RemoveId(r.Name)
	r2, err := New("proj2", []string{"otheruser"}, true)
	c.Assert(err, IsNil)
	defer db.Session.Repository().RemoveId(r2.Name)
	u := struct {
		Name string `bson:"_id"`
	}{Name: "lolcat"}
	err = db.Session.User().Insert(&u)
	c.Assert(err, IsNil)
	defer db.Session.User().RemoveId(u.Name)
	err = GrantAccess([]string{r.Name, r2.Name}, []string{u.Name})
	c.Assert(err, IsNil)
	err = db.Session.Repository().FindId(r.Name).One(&r)
	c.Assert(err, IsNil)
	err = db.Session.Repository().FindId(r2.Name).One(&r2)
	c.Assert(err, IsNil)
	c.Assert(r.Users, DeepEquals, []string{"someuser", u.Name})
	c.Assert(r2.Users, DeepEquals, []string{"otheruser", u.Name})
}

func (s *S) TestGrantAccessShouldAddFirstUserIntoRepositoryDocument(c *C) {
	r := Repository{Name: "proj1"}
	err := db.Session.Repository().Insert(&r)
	c.Assert(err, IsNil)
	defer db.Session.Repository().RemoveId(r.Name)
	r2 := Repository{Name: "proj2"}
	err = db.Session.Repository().Insert(&r2)
	c.Assert(err, IsNil)
	defer db.Session.Repository().RemoveId(r2.Name)
	err = GrantAccess([]string{r.Name, r2.Name}, []string{"Umi"})
	c.Assert(err, IsNil)
	err = db.Session.Repository().FindId(r.Name).One(&r)
	c.Assert(err, IsNil)
	err = db.Session.Repository().FindId(r.Name).One(&r2)
	c.Assert(err, IsNil)
	c.Assert(r.Users, DeepEquals, []string{"Umi"})
	c.Assert(r2.Users, DeepEquals, []string{"Umi"})
}

func (s *S) TestGrantAccessShouldSkipDuplicatedUsers(c *C) {
	r := Repository{Name: "proj1", Users: []string{"umi", "luke", "pade"}}
	err := db.Session.Repository().Insert(&r)
	c.Assert(err, IsNil)
	defer db.Session.Repository().RemoveId(r.Name)
	err = GrantAccess([]string{r.Name}, []string{"pade"})
	c.Assert(err, IsNil)
	err = db.Session.Repository().FindId(r.Name).One(&r)
	c.Assert(err, IsNil)
	c.Assert(r.Users, DeepEquals, []string{"umi", "luke", "pade"})
}

func (s *S) TestRevokeAccessShouldRemoveUserFromAllRepositories(c *C) {
	tmpdir, err := commandmocker.Add("git", "$*")
	c.Assert(err, IsNil)
	defer commandmocker.Remove(tmpdir)
	r, err := New("proj1", []string{"someuser", "umi"}, true)
	c.Assert(err, IsNil)
	defer db.Session.Repository().RemoveId(r.Name)
	r2, err := New("proj2", []string{"otheruser", "umi"}, true)
	c.Assert(err, IsNil)
	defer db.Session.Repository().RemoveId(r2.Name)
	err = RevokeAccess([]string{r.Name, r2.Name}, []string{"umi"})
	c.Assert(err, IsNil)
	err = db.Session.Repository().FindId(r.Name).One(&r)
	c.Assert(err, IsNil)
	err = db.Session.Repository().FindId(r2.Name).One(&r2)
	c.Assert(err, IsNil)
	c.Assert(r.Users, DeepEquals, []string{"someuser"})
	c.Assert(r2.Users, DeepEquals, []string{"otheruser"})
}

func (s *S) TestConflictingRepositoryNameShouldReturnExplicitError(c *C) {
	tmpdir, err := commandmocker.Add("git", "$*")
	c.Assert(err, IsNil)
	defer commandmocker.Remove(tmpdir)
	_, err = New("someRepo", []string{"gollum"}, true)
	defer db.Session.Repository().Remove(bson.M{"_id": "someRepo"})
	c.Assert(err, IsNil)
	_, err = New("someRepo", []string{"gollum"}, true)
	c.Assert(err, ErrorMatches, "A repository with this name already exists.")
}
