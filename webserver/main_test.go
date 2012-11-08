package main

import (
	"fmt"
	"github.com/globocom/commandmocker"
	"github.com/globocom/config"
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
	s.tmpdir, err = commandmocker.Add("git", "$*")
	c.Assert(err, IsNil)
}

func (s *S) TearDownSuite(c *C) {
	commandmocker.Remove(s.tmpdir)
}

func (s *S) TestStartGitDaemonShouldCallGitDaemonCmd(c *C) {
	err := startGitDaemon()
	c.Assert(err, IsNil)
	c.Assert(commandmocker.Ran(s.tmpdir), Equals, true)
	barePath, err := config.GetString("bare-location")
	c.Assert(err, IsNil)
	expected := fmt.Sprintf("daemon --base-path=%s --syslog", barePath)
	c.Assert(commandmocker.Output(s.tmpdir), Equals, expected)
}
