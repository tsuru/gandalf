// Copyright 2014 gandalf authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fs

import (
	"testing"

	tsurufs "github.com/tsuru/tsuru/fs"
	"launchpad.net/gocheck"
)

func Test(t *testing.T) { gocheck.TestingT(t) }

type S struct{}

var _ = gocheck.Suite(&S{})

func (s *S) TestFsystemShouldSetGlobalFsystemWhenItsNil(c *gocheck.C) {
	Fsystem = nil
	fsys := Filesystem()
	_, ok := fsys.(tsurufs.Fs)
	c.Assert(ok, gocheck.Equals, true)
}
