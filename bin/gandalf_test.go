package main

import (
	"bytes"
	"github.com/globocom/commandmocker"
	"github.com/globocom/config"
	"github.com/globocom/gandalf/db"
	"github.com/globocom/gandalf/repository"
	"github.com/globocom/gandalf/user"
	"labix.org/v2/mgo/bson"
	. "launchpad.net/gocheck"
	"log/syslog"
	"os"
	"path"
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
	log, err = syslog.New(syslog.LOG_INFO, "gandalf-listener")
	c.Check(err, IsNil)
	s.user, err = user.New("testuser", []string{})
	c.Check(err, IsNil)
	// does not uses repository.New to avoid creation of bare git repo
	s.repo = &repository.Repository{Name: "myapp", Users: []string{s.user.Name}}
	err = db.Session.Repository().Insert(s.repo)
	c.Check(err, IsNil)
	err = config.ReadConfigFile("../etc/gandalf.conf")
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

func (s *S) TestActionShouldReturnTheCommandBeingExecutedBySSH_ORIGINAL_COMMANDEnvVar(c *C) {
	os.Setenv("SSH_ORIGINAL_COMMAND", "test-cmd")
	defer os.Setenv("SSH_ORIGINAL_COMMAND", "")
	cmd := action()
	c.Assert(cmd, Equals, "test-cmd")
}

func (s *S) TestActionShouldReturnEmptyWhenEnvVarIsNotSet(c *C) {
	cmd := action()
	c.Assert(cmd, Equals, "")
}

func (s *S) TestRequestedRepositoryShouldGetArgumentInSSH_ORIGINAL_COMMANDAndRetrieveTheEquivalentDatabaseRepository(c *C) {
	r := repository.Repository{Name: "foo"}
	err := db.Session.Repository().Insert(&r)
	c.Assert(err, IsNil)
	defer db.Session.Repository().Remove(bson.M{"_id": r.Name})
	os.Setenv("SSH_ORIGINAL_COMMAND", "git-receive-pack 'foo.git'")
	defer os.Setenv("SSH_ORIGINAL_COMMAND", "")
	repo, err := requestedRepository()
	c.Assert(err, IsNil)
	c.Assert(repo.Name, Equals, r.Name)
}

func (s *S) TestRequestedRepositoryShouldDeduceCorrectlyRepositoryNameWithDash(c *C) {
	r := repository.Repository{Name: "foo-bar"}
	err := db.Session.Repository().Insert(&r)
	c.Assert(err, IsNil)
	defer db.Session.Repository().Remove(bson.M{"_id": r.Name})
	os.Setenv("SSH_ORIGINAL_COMMAND", "git-receive-pack 'foo-bar.git'")
	defer os.Setenv("SSH_ORIGINAL_COMMAND", "")
	repo, err := requestedRepository()
	c.Assert(err, IsNil)
	c.Assert(repo.Name, Equals, r.Name)
}

func (s *S) TestRequestedRepositoryShouldReturnErrorWhenCommandDoesNotPassesWhatIsExpected(c *C) {
	os.Setenv("SSH_ORIGINAL_COMMAND", "rm -rf /")
	defer os.Setenv("SSH_ORIGINAL_COMMAND", "")
	_, err := requestedRepository()
	c.Assert(err, ErrorMatches, "^Cannot deduce repository name from command. You are probably trying to do something you shouldn't$")
}

func (s *S) TestRequestedRepositoryShouldReturnErrorWhenThereIsNoCommandPassedToSSH_ORIGINAL_COMMAND(c *C) {
	_, err := requestedRepository()
	c.Assert(err, ErrorMatches, "^Cannot deduce repository name from command. You are probably trying to do something you shouldn't$")
}

func (s *S) TestRequestedRepositoryShouldReturnFormatedErrorWhenRepositoryDoesNotExists(c *C) {
	os.Setenv("SSH_ORIGINAL_COMMAND", "git-receive-pack 'inexistent-repo.git'")
	defer os.Setenv("SSH_ORIGINAL_COMMAND", "")
	_, err := requestedRepository()
	c.Assert(err, ErrorMatches, "^Repository not found$")
}

func (s *S) TestRequestedRepositoryShouldReturnEmptyRepositoryStructOnError(c *C) {
	repo, err := requestedRepository()
	c.Assert(err, NotNil)
	c.Assert(repo.Name, Equals, "")
}

func (s *S) TestRequestedRepositoryName(c *C) {
	os.Setenv("SSH_ORIGINAL_COMMAND", "git-receive-pack 'foobar.git'")
	defer os.Setenv("SSH_ORIGINAL_COMMAND", "")
	name, err := requestedRepositoryName()
	c.Assert(err, IsNil)
	c.Assert(name, Equals, "foobar")
}

func (s *S) TestrequestedRepositoryNameShouldReturnErrorWhenTheresNoMatch(c *C) {
	os.Setenv("SSH_ORIGINAL_COMMAND", "git-receive-pack foobar")
	defer os.Setenv("SSH_ORIGINAL_COMMAND", "")
	name, err := requestedRepositoryName()
	c.Assert(err, ErrorMatches, "Cannot deduce repository name from command. You are probably trying to do something nasty")
	c.Assert(name, Equals, "")
}

func (s *S) TestValidateCmdReturnsErrorWhenSSH_ORIGINAL_COMMANDIsNotAGitCommand(c *C) {
	os.Setenv("SSH_ORIGINAL_COMMAND", "rm -rf /")
	defer os.Setenv("SSH_ORIGINAL_COMMAND", "")
	err := validateCmd()
	c.Assert(err, ErrorMatches, "^You've tried to execute some weird command, I'm deliberately denying you to do that, get over it.$")
}

func (s *S) TestValidateCmdDoNotReturnsErrorWhenSSH_ORIGINAL_COMMANDIsAValidGitCommand(c *C) {
	os.Setenv("SSH_ORIGINAL_COMMAND", "git-receive-pack my-repo.git")
	defer os.Setenv("SSH_ORIGINAL_COMMAND", "")
	err := validateCmd()
	c.Assert(err, IsNil)
}

func (s *S) TestExecuteActionShouldExecuteGitReceivePackWhenUserHasWritePermission(c *C) {
	dir, err := commandmocker.Add("git-receive-pack", "$*")
	c.Check(err, IsNil)
	defer commandmocker.Remove(dir)
	os.Args = []string{"gandalf", s.user.Name}
	os.Setenv("SSH_ORIGINAL_COMMAND", "git-receive-pack 'myapp.git'")
	defer func() {
		os.Args = []string{}
		os.Setenv("SSH_ORIGINAL_COMMAND", "")
	}()
	stdout := &bytes.Buffer{}
	executeAction(hasWritePermission, "You don't have access to write in this repository.", stdout)
	c.Assert(commandmocker.Ran(dir), Equals, true)
	p, err := config.GetString("bare-location")
	c.Assert(err, IsNil)
	expected := path.Join(p, "myapp.git")
	c.Assert(stdout.String(), Equals, expected)
}

func (s *S) TestExecuteActionShouldNotCallSSH_ORIGINAL_COMMANDWhenUserDoesNotExists(c *C) {
	dir, err := commandmocker.Add("git-receive-pack", "$*")
	c.Check(err, IsNil)
	defer commandmocker.Remove(dir)
	os.Args = []string{"gandalf", "god"}
	os.Setenv("SSH_ORIGINAL_COMMAND", "git-receive-pack 'myapp.git'")
	defer func() {
		os.Args = []string{}
		os.Setenv("SSH_ORIGINAL_COMMAND", "")
	}()
	stdout := new(bytes.Buffer)
	errorMsg := "You don't have access to write in this repository."
	executeAction(hasWritePermission, errorMsg, stdout)
	c.Assert(commandmocker.Ran(dir), Equals, false)
}

func (s *S) TestExecuteActionShouldNotCallSSH_ORIGINAL_COMMANDWhenRepositoryDoesNotExists(c *C) {
	dir, err := commandmocker.Add("git-receive-pack", "$*")
	c.Check(err, IsNil)
	defer commandmocker.Remove(dir)
	os.Args = []string{"gandalf", s.user.Name}
	os.Setenv("SSH_ORIGINAL_COMMAND", "git-receive-pack 'ghostapp.git'")
	defer func() {
		os.Args = []string{}
		os.Setenv("SSH_ORIGINAL_COMMAND", "")
	}()
	stdout := &bytes.Buffer{}
	errorMsg := "You don't have access to write in this repository."
	executeAction(hasWritePermission, errorMsg, stdout)
	c.Assert(commandmocker.Ran(dir), Equals, false)
}

func (s *S) TestFormatCommandShouldReceiveAGitCommandAndCanonizalizeTheRepositoryPath(c *C) {
	os.Setenv("SSH_ORIGINAL_COMMAND", "git-receive-pack 'myproject.git'")
	defer os.Setenv("SSH_ORIGINAL_COMMAND", "")
	cmd, err := formatCommand()
	c.Assert(err, IsNil)
	p, err := config.GetString("bare-location")
	c.Assert(err, IsNil)
	expected := path.Join(p, "myproject.git")
	c.Assert(cmd, DeepEquals, []string{"git-receive-pack", expected})
}
