// Copyright 2013 gandalf authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package repository

import (
	"fmt"
	"github.com/globocom/commandmocker"
	"github.com/globocom/config"
	"github.com/globocom/gandalf/fs"
	"github.com/globocom/tsuru/fs/testing"
	. "launchpad.net/gocheck"
	"path"
)

func (s *S) TestBareLocationValuShouldComeFromGandalfConf(c *C) {
	bare = ""
	config.Set("git:bare:location", "/home/gandalf")
	l := bareLocation()
	c.Assert(l, Equals, "/home/gandalf")
}

func (s *S) TestBareLocationShouldResetBareValue(c *C) {
	l := bareLocation()
	config.Set("git:bare:location", "fooo/baaar")
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

func (s *S) TestNewBareShouldPassTemplateOptionWhenItExistsOnConfig(c *C) {
	bareTemplate := "/var/templates"
	bareLocation, err := config.GetString("git:bare:location")
	config.Set("git:bare:template", bareTemplate)
	defer config.Unset("git:bare:template")
	barePath := path.Join(bareLocation, "foo.git")
	dir, err := commandmocker.Add("git", "$*")
	c.Assert(err, IsNil)
	defer commandmocker.Remove(dir)
	err = newBare("foo")
	c.Assert(err, IsNil)
	c.Assert(commandmocker.Ran(dir), Equals, true)
	expected := fmt.Sprintf("init %s --bare --template=%s", barePath, bareTemplate)
	c.Assert(commandmocker.Output(dir), Equals, expected)
}

func (s *S) TestNewBareShouldNotPassTemplateOptionWhenItsNotSetInConfig(c *C) {
	config.Unset("git:bare:template")
	bareLocation, err := config.GetString("git:bare:location")
	c.Assert(err, IsNil)
	barePath := path.Join(bareLocation, "foo.git")
	dir, err := commandmocker.Add("git", "$*")
	c.Assert(err, IsNil)
	defer commandmocker.Remove(dir)
	err = newBare("foo")
	c.Assert(err, IsNil)
	c.Assert(commandmocker.Ran(dir), Equals, true)
	expected := fmt.Sprintf("init %s --bare", barePath)
	c.Assert(commandmocker.Output(dir), Equals, expected)
}

func (s *S) TestRemoveBareShouldRemoveBareDirFromFileSystem(c *C) {
	rfs := &testing.RecordingFs{FileContent: "foo"}
	fs.Fsystem = rfs
	defer func() { fs.Fsystem = nil }()
	err := removeBare("myBare")
	c.Assert(err, IsNil)
	action := "removeall " + path.Join(bareLocation(), "myBare.git")
	c.Assert(rfs.HasAction(action), Equals, true)
}

func (s *S) TestRemoveBareShouldReturnDescriptiveErrorWhenRemovalFails(c *C) {
	rfs := &testing.RecordingFs{FileContent: "foo"}
	fs.Fsystem = &testing.FailureFs{RecordingFs: *rfs}
	defer func() { fs.Fsystem = nil }()
	err := removeBare("fooo")
	c.Assert(err, ErrorMatches, "^Could not remove git bare repository: .*")
}

func (s *S) TestFormatNameShouldAppendDotGitInTheEndOfTheRepoName(c *C) {
	rName := formatName("myrepo")
	c.Assert(rName, Equals, "myrepo.git")
}
