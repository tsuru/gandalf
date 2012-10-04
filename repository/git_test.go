package repository

import (
	"github.com/globocom/commandmocker"
	"github.com/globocom/config"
	"github.com/timeredbull/gandalf/fs/testing"
	. "launchpad.net/gocheck"
	"path"
)

func (s *S) TestBareLocationValuShouldComeFromGandalfConf(c *C) {
	bare = ""
	config.Set("bare-location", "/home/gandalf")
	l := bareLocation()
	c.Assert(l, Equals, "/home/gandalf")
}

func (s *S) TestBareLocationShouldResetBareValue(c *C) {
	l := bareLocation()
	config.Set("bare-location", "fooo/baaar")
	c.Assert(bareLocation(), Equals, l)
}

func (s *S) TestNewBareShouldCreateADir(c *C) {
	dir, err := commandmocker.Add("git", "$*")
	c.Check(err, IsNil)
	defer commandmocker.Remove(dir)
	err = newBare("myBare")
	c.Assert(err, IsNil)
	c.Assert(commandmocker.Ran(dir), Equals, true)
}

func (s *S) TestNewBareShouldReturnMeaningfullErrorWhenBareCreationFails(c *C) {
	dir, err := commandmocker.Error("git", "ooooi", 1)
	c.Assert(err, IsNil)
	defer commandmocker.Remove(dir)
	err = newBare("foo")
	c.Check(err, NotNil)
	got := err.Error()
	expected := "Could not create git bare repository: exit status 1"
	c.Assert(got, Equals, expected)
}

func (s *S) TestRemoveBareShouldRemoveBareDirFromFileSystem(c *C) {
	rfs := &testing.RecordingFs{FileContent: "foo"}
	fsystem = rfs
	defer func() { fsystem = nil }()
	err := removeBare("myBare")
	c.Assert(err, IsNil)
	action := "removeall " + path.Join(bareLocation(), "myBare")
	c.Assert(rfs.HasAction(action), Equals, true)
}

func (s *S) TestRemoveBareShouldReturnDescriptiveErrorWhenRemovalFails(c *C) {
	rfs := &testing.RecordingFs{FileContent: "foo"}
	fsystem = &testing.FailureFs{RecordingFs: *rfs}
	defer func() { fsystem = nil }()
	err := removeBare("fooo")
	c.Assert(err, ErrorMatches, "^Could not remove git bare repository: .*")
}
