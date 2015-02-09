// Copyright 2015 gandalf authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

import (
	"testing"

	"github.com/gorilla/pat"
	"github.com/tsuru/commandmocker"
	"github.com/tsuru/config"
	"github.com/tsuru/gandalf/db"
	"github.com/tsuru/gandalf/fs"
	"github.com/tsuru/gandalf/user"
	"github.com/tsuru/tsuru/fs/fstest"
	"gopkg.in/check.v1"
	"gopkg.in/mgo.v2/bson"
)

func Test(t *testing.T) { check.TestingT(t) }

type S struct {
	tmpdir string
	rfs    *fstest.RecordingFs
	router *pat.Router
}

var _ = check.Suite(&S{})

func (s *S) SetUpSuite(c *check.C) {
	err := config.ReadConfigFile("../etc/gandalf.conf")
	c.Assert(err, check.IsNil)
	config.Set("database:url", "127.0.0.1:27017")
	config.Set("database:name", "gandalf_api_tests")
	s.tmpdir, err = commandmocker.Add("git", "")
	c.Assert(err, check.IsNil)
	s.router = SetupRouter()
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

func (s *S) TestGetUserOr404(c *check.C) {
	u := user.User{Name: "umi"}
	conn, err := db.Conn()
	c.Assert(err, check.IsNil)
	defer conn.Close()
	err = conn.User().Insert(&u)
	c.Assert(err, check.IsNil)
	defer conn.User().Remove(bson.M{"_id": u.Name})
	rUser, err := getUserOr404("umi")
	c.Assert(err, check.IsNil)
	c.Assert(rUser.Name, check.Equals, "umi")
}

func (s *S) TestGetUserOr404ShouldReturn404WhenUserDoesntExist(c *check.C) {
	_, e := getUserOr404("umi")
	expected := "User umi not found"
	got := e.Error()
	c.Assert(got, check.Equals, expected)
}
