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
	c.Check(err, IsNil)
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
	remote := (&Repository{Name: "lol"}).Remote()
	c.Assert(remote, Equals, "git@gandalfhost.com:lol.git")
}

func (s *S) TestGrantAccessShouldAddUserToRepositoryDocument(c *C) {
	tmpdir, err := commandmocker.Add("git", "$*")
	c.Assert(err, IsNil)
	defer commandmocker.Remove(tmpdir)
	r, err := New("myproj", []string{"someuser"}, true)
	c.Assert(err, IsNil)
	defer db.Session.Repository().RemoveId(r.Name)
	u := struct {
		Name string `bson:"_id"`
	}{Name: "lolcat"}
	err = db.Session.User().Insert(&u)
	c.Assert(err, IsNil)
	defer db.Session.User().RemoveId(u.Name)
	err = GrantAccess(r.Name, u.Name)
	c.Assert(err, IsNil)
	err = db.Session.Repository().FindId(r.Name).One(&r)
	c.Assert(err, IsNil)
	c.Assert(r.Users, DeepEquals, []string{"someuser", u.Name})
}

func (s *S) TestGrantAccessShouldReturnFormatedErrorWhenRepositoryDoesNotExists(c *C) {
	u := struct {
		Name string `bson:"_id"`
	}{Name: "lolcat"}
	err := db.Session.User().Insert(&u)
	c.Assert(err, IsNil)
	defer db.Session.User().RemoveId(u.Name)
	err = GrantAccess("absentrepo", "someuser")
	c.Assert(err, ErrorMatches, "^Repository \"absentrepo\" does not exists$")
}

func (s *S) TestGrantAccessShouldReturnFormatedErrorWhenUserDoesNotExists(c *C) {
	tmpdir, err := commandmocker.Add("git", "$*")
	c.Assert(err, IsNil)
	defer commandmocker.Remove(tmpdir)
	r, err := New("myproj", []string{"someuser"}, true)
	c.Assert(err, IsNil)
	defer db.Session.Repository().RemoveId(r.Name)
	err = GrantAccess(r.Name, "absentuser")
	c.Assert(err, ErrorMatches, "^User \"absentuser\" does not exists$")
}

func (s *S) TestBulkGrantAccessShouldAddUserToListOfRepositories(c *C) {
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
	err = BulkGrantAccess(u.Name, []string{r.Name, r2.Name})
	c.Assert(err, IsNil)
	err = db.Session.Repository().FindId(r.Name).One(&r)
	c.Assert(err, IsNil)
	err = db.Session.Repository().FindId(r2.Name).One(&r2)
	c.Assert(err, IsNil)
	c.Assert(r.Users, DeepEquals, []string{"someuser", u.Name})
	c.Assert(r2.Users, DeepEquals, []string{"otheruser", u.Name})
}

func (s *S) TestBulkGrantAccessShouldAddFirstUserIntoRepositoryDocument(c *C) {
	r := Repository{Name: "proj1"}
	err := db.Session.Repository().Insert(&r)
	c.Assert(err, IsNil)
	defer db.Session.Repository().RemoveId(r.Name)
	r2 := Repository{Name: "proj2"}
	err = db.Session.Repository().Insert(&r2)
	c.Assert(err, IsNil)
	defer db.Session.Repository().RemoveId(r2.Name)
	err = BulkGrantAccess("Umi", []string{r.Name, r2.Name})
	c.Assert(err, IsNil)
	err = db.Session.Repository().FindId(r.Name).One(&r)
	c.Assert(err, IsNil)
	err = db.Session.Repository().FindId(r.Name).One(&r2)
	c.Assert(err, IsNil)
	c.Assert(r.Users, DeepEquals, []string{"Umi"})
	c.Assert(r2.Users, DeepEquals, []string{"Umi"})
}

func (s *S) TestBulkRevokeAccessShouldRemoveUserFromAllRepositories(c *C) {
	tmpdir, err := commandmocker.Add("git", "$*")
	c.Assert(err, IsNil)
	defer commandmocker.Remove(tmpdir)
	r, err := New("proj1", []string{"someuser", "umi"}, true)
	c.Assert(err, IsNil)
	defer db.Session.Repository().RemoveId(r.Name)
	r2, err := New("proj2", []string{"otheruser", "umi"}, true)
	c.Assert(err, IsNil)
	defer db.Session.Repository().RemoveId(r2.Name)
	err = BulkRevokeAccess("umi", []string{r.Name, r2.Name})
	c.Assert(err, IsNil)
	err = db.Session.Repository().FindId(r.Name).One(&r)
	c.Assert(err, IsNil)
	err = db.Session.Repository().FindId(r2.Name).One(&r2)
	c.Assert(err, IsNil)
	c.Assert(r.Users, DeepEquals, []string{"someuser"})
	c.Assert(r2.Users, DeepEquals, []string{"otheruser"})
}

func (s *S) TestRevokeAccessShouldRemoveUserFromRepositoryDocument(c *C) {
	tmpdir, err := commandmocker.Add("git", "$*")
	c.Assert(err, IsNil)
	defer commandmocker.Remove(tmpdir)
	u := struct {
		Name string `bson:"_id"`
	}{Name: "lolcat"}
	err = db.Session.User().Insert(&u)
	c.Assert(err, IsNil)
	defer db.Session.User().RemoveId(u.Name)
	r, err := New("myproj", []string{u.Name, "zezinho"}, true)
	c.Assert(err, IsNil)
	defer db.Session.Repository().RemoveId(r.Name)
	err = RevokeAccess(r.Name, u.Name)
	c.Assert(err, IsNil)
	err = db.Session.Repository().FindId(r.Name).One(&r)
	c.Assert(err, IsNil)
	c.Assert(r.Users, DeepEquals, []string{"zezinho"})
}

func (s *S) TestRevokeAccessShouldReturnFormatedErrorWhenUserHasNotAccessToRepository(c *C) {
	tmpdir, err := commandmocker.Add("git", "$*")
	c.Assert(err, IsNil)
	defer commandmocker.Remove(tmpdir)
	r, err := New("myproj", []string{"zezinho", "luizinho"}, true)
	c.Assert(err, IsNil)
	defer db.Session.Repository().RemoveId(r.Name)
	err = RevokeAccess(r.Name, "gandalf")
	expected := fmt.Sprintf("^User \"gandalf\" does not have access to repository \"%s\"$", r.Name)
	c.Assert(err, ErrorMatches, expected)
}

func (s *S) TestRevokeAccessShouldReturnFormatedErrorWhenRepositoryDoesNotExists(c *C) {
	err := RevokeAccess("absentrepo", "gandalf")
	c.Assert(err, ErrorMatches, "^Repository \"absentrepo\" does not exists$")
}

func (s *S) TestRevokeAccessShouldReturnErrorIfUserBeingRevokedIsTheOnlyOneWithAccessIntoRepository(c *C) {
	tmpdir, err := commandmocker.Add("git", "$*")
	c.Assert(err, IsNil)
	defer commandmocker.Remove(tmpdir)
	r, err := New("myproj", []string{"luizinho"}, true)
	c.Assert(err, IsNil)
	defer db.Session.Repository().RemoveId(r.Name)
	err = RevokeAccess(r.Name, "luizinho")
	c.Assert(err, ErrorMatches, "^Cannot revoke access to only user that has access into repository \"myproj\"$")
}
