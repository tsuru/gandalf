package user

import (
	"github.com/globocom/config"
	"github.com/globocom/gandalf/db"
	"github.com/globocom/gandalf/repository"
	"github.com/globocom/tsuru/fs"
	fstesting "github.com/globocom/tsuru/fs/testing"
	"io/ioutil"
	"labix.org/v2/mgo/bson"
	. "launchpad.net/gocheck"
	"os"
	"path"
	"testing"
)

func Test(t *testing.T) { TestingT(t) }

type S struct {
	rfs *fstesting.RecordingFs
}

var _ = Suite(&S{})

func (s *S) authKeysContent(c *C) string {
	authFile := path.Join(os.Getenv("HOME"), ".ssh", "authorized_keys")
	f, err := filesystem().OpenFile(authFile, os.O_RDWR, 0755)
	c.Assert(err, IsNil)
	b, err := ioutil.ReadAll(f)
	c.Assert(err, IsNil)
	return string(b)
}

func (s *S) SetUpSuite(c *C) {
	err := config.ReadConfigFile("../etc/gandalf.conf")
	c.Check(err, IsNil)
}

func (s *S) SetUpTest(c *C) {
	s.rfs = &fstesting.RecordingFs{}
	fsystem = s.rfs
}

func (s *S) TearDownSuite(c *C) {
	fsystem = nil
}

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

func (s *S) TestNewWritesKeyInAuthorizedKeys(c *C) {
	u, err := New("piccolo", []string{"idrsakey piccolo@myhost"})
	c.Assert(err, IsNil)
	defer db.Session.User().Remove(bson.M{"_id": u.Name})
	keys := s.authKeysContent(c)
	c.Assert(keys, Matches, ".*idrsakey piccolo@myhost")
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
	err = Remove(u.Name)
	c.Assert(err, IsNil)
	lenght, err := db.Session.User().Find(bson.M{"_id": u.Name}).Count()
	c.Assert(err, IsNil)
	c.Assert(lenght, Equals, 0)
}

func (s *S) TestRemoveRemovesKeyFromAuthorizedKeysFile(c *C) {
	u, err := New("gandalf", []string{"gandalfkey gandalf@mordor"})
	c.Assert(err, IsNil)
	err = Remove(u.Name)
	c.Assert(err, IsNil)
	got := s.authKeysContent(c)
	c.Assert(got, Not(Matches), ".*gandalfkey gandalf@mordor")
}

func (s *S) TestRemoveInexistentUserReturnsDescriptiveMessage(c *C) {
	err := Remove("otheruser")
	c.Assert(err, ErrorMatches, "Could not remove user: not found")
}

func (s *S) TestRemoveDoesNotRemovesUserWhenUserIsTheOnlyOneAssciatedWithOneRepository(c *C) {
	u, err := New("silver", []string{})
	c.Assert(err, IsNil)
	r := s.createRepo("run", []string{u.Name}, c)
	defer db.Session.Repository().Remove(bson.M{"_id": r.Name})
	defer db.Session.User().Remove(bson.M{"_id": u.Name})
	err = Remove(u.Name)
	c.Assert(err, ErrorMatches, "^Could not remove user: user is the only one with access to at least one of it's repositories$")
}

func (s *S) TestRemoveRevokesAccessToReposWithMoreThanOneUserAssociated(c *C) {
	u, r, r2 := s.userPlusRepos(c)
	defer db.Session.Repository().Remove(bson.M{"_id": r.Name})
	defer db.Session.Repository().Remove(bson.M{"_id": r2.Name})
	defer db.Session.User().Remove(bson.M{"_id": u.Name})
	err := Remove(u.Name)
	c.Assert(err, IsNil)
	s.retrieveRepos(r, r2, c)
	c.Assert(r.Users, DeepEquals, []string{"slot"})
	c.Assert(r2.Users, DeepEquals, []string{"cnot"})
}

func (s *S) retrieveRepos(r, r2 *repository.Repository, c *C) {
	err := db.Session.Repository().Find(bson.M{"_id": r.Name}).One(&r)
	c.Assert(err, IsNil)
	err = db.Session.Repository().Find(bson.M{"_id": r2.Name}).One(&r2)
	c.Assert(err, IsNil)
}

func (s *S) userPlusRepos(c *C) (*User, *repository.Repository, *repository.Repository) {
	u, err := New("silver", []string{})
	c.Assert(err, IsNil)
	r := s.createRepo("run", []string{u.Name, "slot"}, c)
	r2 := s.createRepo("stay", []string{u.Name, "cnot"}, c)
	return u, &r, &r2
}

func (s *S) createRepo(name string, users []string, c *C) repository.Repository {
	r := repository.Repository{Name: name, Users: users}
	err := db.Session.Repository().Insert(&r)
	c.Assert(err, IsNil)
	return r
}

func (s *S) TestHandleAssociatedRepositoriesShouldRevokeAccessToRepoWithMoreThanOneUserAssociated(c *C) {
	u, r, r2 := s.userPlusRepos(c)
	defer db.Session.Repository().RemoveId(r.Name)
	defer db.Session.Repository().RemoveId(r2.Name)
	defer db.Session.User().RemoveId(u.Name)
	err := u.handleAssociatedRepositories()
	c.Assert(err, IsNil)
	s.retrieveRepos(r, r2, c)
	c.Assert(r.Users, DeepEquals, []string{"slot"})
	c.Assert(r2.Users, DeepEquals, []string{"cnot"})
}

func (s *S) TestHandleAssociateRepositoriesReturnsErrorWhenUserIsOnlyOneWithAccessToAtLeastOneRepo(c *C) {
	u, err := New("umi", []string{})
	c.Assert(err, IsNil)
	r := s.createRepo("proj1", []string{"umi"}, c)
	defer db.Session.User().RemoveId(u.Name)
	defer db.Session.Repository().RemoveId(r.Name)
	err = u.handleAssociatedRepositories()
	expected := "^Could not remove user: user is the only one with access to at least one of it's repositories$"
	c.Assert(err, ErrorMatches, expected)
}

func (s *S) TestFsystemShouldSetGlobalFsystemWhenItsNil(c *C) {
	fsystem = nil
	fsys := filesystem()
	_, ok := fsys.(fs.Fs)
	c.Assert(ok, Equals, true)
}
