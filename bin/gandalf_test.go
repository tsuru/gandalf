// Copyright 2015 gandalf authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"log/syslog"
	"os"
	"path"
	"testing"

	"github.com/tsuru/commandmocker"
	"github.com/tsuru/config"
	"github.com/tsuru/gandalf/db"
	"github.com/tsuru/gandalf/repository"
	"github.com/tsuru/gandalf/user"
	"gopkg.in/check.v1"
	"gopkg.in/mgo.v2/bson"
)

func Test(t *testing.T) { check.TestingT(t) }

type S struct {
	user *user.User
	repo *repository.Repository
}

var _ = check.Suite(&S{})

func (s *S) SetUpSuite(c *check.C) {
	var err error
	log, err = syslog.New(syslog.LOG_INFO, "gandalf-listener")
	c.Check(err, check.IsNil)
	err = config.ReadConfigFile("../etc/gandalf.conf")
	c.Check(err, check.IsNil)
	config.Set("database:name", "gandalf_bin_tests")
	s.user, err = user.New("testuser", map[string]string{})
	c.Check(err, check.IsNil)
	// does not uses repository.New to avoid creation of bare git repo
	s.repo = &repository.Repository{Name: "myapp", Users: []string{s.user.Name}}
	conn, err := db.Conn()
	c.Assert(err, check.IsNil)
	defer conn.Close()
	err = conn.Repository().Insert(s.repo)
	c.Check(err, check.IsNil)
}

func (s *S) TearDownSuite(c *check.C) {
	conn, err := db.Conn()
	c.Assert(err, check.IsNil)
	defer conn.Close()
	conn.User().Database.DropDatabase()
}

func (s *S) TestHasWritePermissionSholdReturnTrueWhenUserCanWriteInRepo(c *check.C) {
	allowed := hasWritePermission(s.user, s.repo)
	c.Assert(allowed, check.Equals, true)
}

func (s *S) TestHasWritePermissionShouldReturnFalseWhenUserCannotWriteinRepo(c *check.C) {
	r := &repository.Repository{Name: "myotherapp"}
	conn, err := db.Conn()
	c.Assert(err, check.IsNil)
	defer conn.Close()
	conn.Repository().Insert(&r)
	defer conn.Repository().Remove(bson.M{"_id": r.Name})
	allowed := hasWritePermission(s.user, r)
	c.Assert(allowed, check.Equals, false)
}

func (s *S) TestHasReadPermissionShouldReturnTrueWhenRepositoryIsPublic(c *check.C) {
	r := &repository.Repository{Name: "myotherapp", IsPublic: true}
	conn, err := db.Conn()
	c.Assert(err, check.IsNil)
	defer conn.Close()
	conn.Repository().Insert(&r)
	defer conn.Repository().Remove(bson.M{"_id": r.Name})
	allowed := hasReadPermission(s.user, r)
	c.Assert(allowed, check.Equals, true)
}

func (s *S) TestHasReadPermissionShouldReturnTrueWhenRepositoryIsNotPublicAndUserHasPermissionToRead(c *check.C) {
	user, err := user.New("readonlyuser", map[string]string{})
	c.Check(err, check.IsNil)
	repo := &repository.Repository{
		Name:          "otherapp",
		Users:         []string{s.user.Name},
		ReadOnlyUsers: []string{user.Name},
	}
	allowed := hasReadPermission(user, repo)
	c.Assert(allowed, check.Equals, true)
}

func (s *S) TestHasReadPermissionShouldReturnTrueWhenRepositoryIsNotPublicAndUserHasPermissionToReadAndWrite(c *check.C) {
	allowed := hasReadPermission(s.user, s.repo)
	c.Assert(allowed, check.Equals, true)
}

func (s *S) TestHasReadPermissionShouldReturnFalseWhenUserDoesNotHavePermissionToReadWriteAndRepoIsNotPublic(c *check.C) {
	r := &repository.Repository{Name: "myotherapp", IsPublic: false}
	conn, err := db.Conn()
	c.Assert(err, check.IsNil)
	defer conn.Close()
	conn.Repository().Insert(&r)
	defer conn.Repository().Remove(bson.M{"_id": r.Name})
	allowed := hasReadPermission(s.user, r)
	c.Assert(allowed, check.Equals, false)
}

func (s *S) TestActionShouldReturnTheCommandBeingExecutedBySSH_ORIGINAL_COMMANDEnvVar(c *check.C) {
	os.Setenv("SSH_ORIGINAL_COMMAND", "test-cmd")
	defer os.Setenv("SSH_ORIGINAL_COMMAND", "")
	cmd := action()
	c.Assert(cmd, check.Equals, "test-cmd")
}

func (s *S) TestActionShouldReturnEmptyWhenEnvVarIsNotSet(c *check.C) {
	cmd := action()
	c.Assert(cmd, check.Equals, "")
}

func (s *S) TestRequestedRepositoryShouldGetArgumentInSSH_ORIGINAL_COMMANDAndRetrieveTheEquivalentDatabaseRepository(c *check.C) {
	r := repository.Repository{Name: "foo"}
	conn, err := db.Conn()
	c.Assert(err, check.IsNil)
	defer conn.Close()
	err = conn.Repository().Insert(&r)
	c.Assert(err, check.IsNil)
	defer conn.Repository().Remove(bson.M{"_id": r.Name})
	os.Setenv("SSH_ORIGINAL_COMMAND", "git-receive-pack 'foo.git'")
	defer os.Setenv("SSH_ORIGINAL_COMMAND", "")
	repo, err := requestedRepository()
	c.Assert(err, check.IsNil)
	c.Assert(repo.Name, check.Equals, r.Name)
}

func (s *S) TestRequestedRepositoryShouldDeduceCorrectlyRepositoryNameWithDash(c *check.C) {
	r := repository.Repository{Name: "foo-bar"}
	conn, err := db.Conn()
	c.Assert(err, check.IsNil)
	defer conn.Close()
	err = conn.Repository().Insert(&r)
	c.Assert(err, check.IsNil)
	defer conn.Repository().Remove(bson.M{"_id": r.Name})
	os.Setenv("SSH_ORIGINAL_COMMAND", "git-receive-pack 'foo-bar.git'")
	defer os.Setenv("SSH_ORIGINAL_COMMAND", "")
	repo, err := requestedRepository()
	c.Assert(err, check.IsNil)
	c.Assert(repo.Name, check.Equals, r.Name)
}

func (s *S) TestRequestedRepositoryShouldReturnErrorWhenCommandDoesNotPassesWhatIsExpected(c *check.C) {
	os.Setenv("SSH_ORIGINAL_COMMAND", "rm -rf /")
	defer os.Setenv("SSH_ORIGINAL_COMMAND", "")
	_, err := requestedRepository()
	c.Assert(err, check.ErrorMatches, "^You've tried to execute some weird command, I'm deliberately denying you to do that, get over it.$")
}

func (s *S) TestRequestedRepositoryShouldReturnErrorWhenThereIsNoCommandPassedToSSH_ORIGINAL_COMMAND(c *check.C) {
	_, err := requestedRepository()
	c.Assert(err, check.ErrorMatches, "^You've tried to execute some weird command, I'm deliberately denying you to do that, get over it.$")
}

func (s *S) TestRequestedRepositoryShouldReturnFormatedErrorWhenRepositoryDoesNotExist(c *check.C) {
	os.Setenv("SSH_ORIGINAL_COMMAND", "git-receive-pack 'inexistent-repo.git'")
	defer os.Setenv("SSH_ORIGINAL_COMMAND", "")
	_, err := requestedRepository()
	c.Assert(err, check.ErrorMatches, "^Repository not found$")
}

func (s *S) TestRequestedRepositoryShouldReturnEmptyRepositoryStructOnError(c *check.C) {
	repo, err := requestedRepository()
	c.Assert(err, check.NotNil)
	c.Assert(repo.Name, check.Equals, "")
}

func (s *S) TestParseGitCommand(c *check.C) {
	os.Setenv("SSH_ORIGINAL_COMMAND", "git-receive-pack 'foobar.git'")
	defer os.Setenv("SSH_ORIGINAL_COMMAND", "")
	_, name, err := parseGitCommand()
	c.Assert(err, check.IsNil)
	c.Assert(name, check.Equals, "foobar")
}

func (s *S) TestParseGitCommandWithSlash(c *check.C) {
	os.Setenv("SSH_ORIGINAL_COMMAND", "git-receive-pack '/foobar.git'")
	defer os.Setenv("SSH_ORIGINAL_COMMAND", "")
	_, name, err := parseGitCommand()
	c.Assert(err, check.IsNil)
	c.Assert(name, check.Equals, "foobar")
}

func (s *S) TestParseGitCommandShouldReturnErrorWhenTheresNoMatch(c *check.C) {
	defer os.Setenv("SSH_ORIGINAL_COMMAND", "")
	os.Setenv("SSH_ORIGINAL_COMMAND", "git-receive-pack foobar")
	_, name, err := parseGitCommand()
	c.Assert(err, check.ErrorMatches, "You've tried to execute some weird command, I'm deliberately denying you to do that, get over it.")
	c.Assert(name, check.Equals, "")
	os.Setenv("SSH_ORIGINAL_COMMAND", "git-receive-pack ../foobar")
	_, name, err = parseGitCommand()
	c.Assert(err, check.ErrorMatches, "You've tried to execute some weird command, I'm deliberately denying you to do that, get over it.")
	c.Assert(name, check.Equals, "")
	os.Setenv("SSH_ORIGINAL_COMMAND", "git-receive-pack /etc")
	_, name, err = parseGitCommand()
	c.Assert(err, check.ErrorMatches, "You've tried to execute some weird command, I'm deliberately denying you to do that, get over it.")
	c.Assert(name, check.Equals, "")
}

func (s *S) TestParseGitCommandReturnsErrorWhenSSH_ORIGINAL_COMMANDIsNotAGitCommand(c *check.C) {
	os.Setenv("SSH_ORIGINAL_COMMAND", "rm -rf /")
	defer os.Setenv("SSH_ORIGINAL_COMMAND", "")
	_, _, err := parseGitCommand()
	c.Assert(err, check.ErrorMatches, "^You've tried to execute some weird command, I'm deliberately denying you to do that, get over it.$")
}

func (s *S) TestParseGitCommandDoNotReturnsErrorWhenSSH_ORIGINAL_COMMANDIsAValidGitCommand(c *check.C) {
	os.Setenv("SSH_ORIGINAL_COMMAND", "git-receive-pack 'my-repo.git'")
	defer os.Setenv("SSH_ORIGINAL_COMMAND", "")
	_, _, err := parseGitCommand()
	c.Assert(err, check.IsNil)
}

func (s *S) TestParseGitCommandDoNotReturnsErrorWhenSSH_ORIGINAL_COMMANDIsAValidGitCommandWithDashInName(c *check.C) {
	os.Setenv("SSH_ORIGINAL_COMMAND", "git-receive-pack '/my-repo.git'")
	defer os.Setenv("SSH_ORIGINAL_COMMAND", "")
	_, _, err := parseGitCommand()
	c.Assert(err, check.IsNil)
}

func (s *S) TestExecuteActionShouldExecuteGitReceivePackWhenUserHasWritePermission(c *check.C) {
	dir, err := commandmocker.Add("git-receive-pack", "$*")
	c.Check(err, check.IsNil)
	defer commandmocker.Remove(dir)
	os.Args = []string{"gandalf", s.user.Name}
	os.Setenv("SSH_ORIGINAL_COMMAND", "git-receive-pack 'myapp.git'")
	defer func() {
		os.Args = []string{}
		os.Setenv("SSH_ORIGINAL_COMMAND", "")
	}()
	stdout := &bytes.Buffer{}
	executeAction(hasWritePermission, "You don't have access to write in this repository.", stdout)
	c.Assert(commandmocker.Ran(dir), check.Equals, true)
	p, err := config.GetString("git:bare:location")
	c.Assert(err, check.IsNil)
	expected := path.Join(p, "myapp.git")
	c.Assert(stdout.String(), check.Equals, expected)
	c.Assert(commandmocker.Envs(dir), check.Matches, `(?s).*TSURU_USER=testuser.*`)
}

func (s *S) TestExecuteActionShouldNotCallSSH_ORIGINAL_COMMANDWhenUserDoesNotExist(c *check.C) {
	dir, err := commandmocker.Add("git-receive-pack", "$*")
	c.Check(err, check.IsNil)
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
	c.Assert(commandmocker.Ran(dir), check.Equals, false)
}

func (s *S) TestExecuteActionShouldNotCallSSH_ORIGINAL_COMMANDWhenRepositoryDoesNotExist(c *check.C) {
	dir, err := commandmocker.Add("git-receive-pack", "$*")
	c.Check(err, check.IsNil)
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
	c.Assert(commandmocker.Ran(dir), check.Equals, false)
}

func (s *S) TestFormatCommandShouldReceiveAGitCommandAndCanonizalizeTheRepositoryPath(c *check.C) {
	os.Setenv("SSH_ORIGINAL_COMMAND", "git-receive-pack 'myproject.git'")
	defer os.Setenv("SSH_ORIGINAL_COMMAND", "")
	cmd, err := formatCommand()
	c.Assert(err, check.IsNil)
	p, err := config.GetString("git:bare:location")
	c.Assert(err, check.IsNil)
	expected := path.Join(p, "myproject.git")
	c.Assert(cmd, check.DeepEquals, []string{"git-receive-pack", expected})
}

func (s *S) TestFormatCommandShouldReceiveAGitCommandAndCanonizalizeTheRepositoryPathWithNamespace(c *check.C) {
	os.Setenv("SSH_ORIGINAL_COMMAND", "git-receive-pack 'me/myproject.git'")
	defer os.Setenv("SSH_ORIGINAL_COMMAND", "")
	cmd, err := formatCommand()
	c.Assert(err, check.IsNil)
	p, err := config.GetString("git:bare:location")
	c.Assert(err, check.IsNil)
	expected := path.Join(p, "me/myproject.git")
	c.Assert(cmd, check.DeepEquals, []string{"git-receive-pack", expected})
}

func (s *S) TestFormatCommandShouldReceiveAGitCommandProjectWithDash(c *check.C) {
	os.Setenv("SSH_ORIGINAL_COMMAND", "git-receive-pack '/myproject.git'")
	defer os.Setenv("SSH_ORIGINAL_COMMAND", "")
	cmd, err := formatCommand()
	c.Assert(err, check.IsNil)
	p, err := config.GetString("git:bare:location")
	c.Assert(err, check.IsNil)
	expected := path.Join(p, "myproject.git")
	c.Assert(cmd, check.DeepEquals, []string{"git-receive-pack", expected})
}
