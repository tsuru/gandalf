// Copyright 2013 gandalf authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fs

import (
	tsurufs "github.com/globocom/tsuru/fs"
	. "launchpad.net/gocheck"
	"testing"
)

func Test(t *testing.T) { TestingT(t) }

type S struct{}

var _ = Suite(&S{})

func (s *S) TestFsystemShouldSetGlobalFsystemWhenItsNil(c *C) {
	Fsystem = nil
	fsys := Filesystem()
	_, ok := fsys.(tsurufs.Fs)
	c.Assert(ok, Equals, true)
}
