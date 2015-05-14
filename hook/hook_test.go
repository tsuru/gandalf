// Copyright 2015 gandalf authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package hook

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/tsuru/commandmocker"
	"github.com/tsuru/config"
	"github.com/tsuru/gandalf/db"
	"github.com/tsuru/gandalf/fs"
	"github.com/tsuru/tsuru/fs/fstest"
	"gopkg.in/check.v1"
)

func Test(t *testing.T) { check.TestingT(t) }

type S struct {
	tmpdir string
	rfs    *fstest.RecordingFs
}

var _ = check.Suite(&S{})

func (s *S) SetUpSuite(c *check.C) {
	err := config.ReadConfigFile("../etc/gandalf.conf")
	c.Assert(err, check.IsNil)
	config.Set("database:url", "127.0.0.1:27017")
	config.Set("database:name", "gandalf_hooks_tests")
	s.tmpdir, err = commandmocker.Add("git", "")
	c.Assert(err, check.IsNil)
}

func (s *S) SetUpTest(c *check.C) {
	s.rfs = &fstest.RecordingFs{}
	fs.Fsystem = s.rfs
	bareTemplate, _ := config.GetString("git:bare:template")
	fs.Fsystem.MkdirAll(bareTemplate+"/hooks", 0755)
}

func (s *S) TearDownTest(c *check.C) {
	fs.Fsystem = nil
}

func (s *S) TearDownSuite(c *check.C) {
	commandmocker.Remove(s.tmpdir)
	conn, err := db.Conn()
	c.Assert(err, check.IsNil)
	defer conn.Close()
	conn.User().Database.DropDatabase()
}

func (s *S) TestCanCreateHookFile(c *check.C) {
	hookContent := []byte("some content")
	err := createHookFile("/tmp/repositories/some-repo.git/hooks/test-can-create-hook-file", hookContent)
	c.Assert(err, check.IsNil)
	file, err := fs.Filesystem().OpenFile("/tmp/repositories/some-repo.git/hooks/test-can-create-hook-file", os.O_RDONLY, 0755)
	defer file.Close()
	content, err := ioutil.ReadAll(file)
	c.Assert(err, check.IsNil)
	c.Assert(string(content), check.Equals, "some content")
}

func (s *S) TestCanAddNewHook(c *check.C) {
	hookContent := []byte("some content")
	err := Add("test-can-add-new-hook", []string{}, hookContent)
	c.Assert(err, check.IsNil)
	file, err := fs.Filesystem().OpenFile("/home/git/bare-template/hooks/test-can-add-new-hook", os.O_RDONLY, 0755)
	defer file.Close()
	content, err := ioutil.ReadAll(file)
	c.Assert(err, check.IsNil)
	c.Assert(string(content), check.Equals, "some content")
}

func (s *S) TestCanAddNewHookInOldRepository(c *check.C) {
	s.rfs = &fstest.RecordingFs{}
	fs.Fsystem = s.rfs
	bareTemplate, _ := config.GetString("git:bare:template")
	err := fs.Fsystem.RemoveAll(bareTemplate + "/hooks")
	c.Assert(err, check.IsNil)
	hookContent := []byte("some content")
	err = Add("test-can-add-new-hook", []string{}, hookContent)
	c.Assert(err, check.IsNil)
	file, err := fs.Filesystem().OpenFile("/home/git/bare-template/hooks/test-can-add-new-hook", os.O_RDONLY, 0755)
	defer file.Close()
	content, err := ioutil.ReadAll(file)
	c.Assert(err, check.IsNil)
	c.Assert(string(content), check.Equals, "some content")
}

func (s *S) TestCanAddNewRepository(c *check.C) {
	hookContent := []byte("some content")
	err := Add("test-can-add-new-repository-hook", []string{"some-repo"}, hookContent)
	c.Assert(err, check.IsNil)
	file, err := fs.Filesystem().OpenFile("/var/lib/gandalf/repositories/some-repo.git/hooks/test-can-add-new-repository-hook", os.O_RDONLY, 0755)
	c.Assert(err, check.IsNil)
	defer file.Close()
	content, err := ioutil.ReadAll(file)
	c.Assert(err, check.IsNil)
	c.Assert(string(content), check.Equals, "some content")
}
