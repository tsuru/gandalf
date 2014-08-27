// Copyright 2014 gandalf authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package testing

import (
	"net/http"
	"testing"

	"launchpad.net/gocheck"
)

func Test(t *testing.T) { gocheck.TestingT(t) }

type S struct{}

var _ = gocheck.Suite(&S{})

func (s *S) TestGandalfServerShouldRespondeToCalls(c *gocheck.C) {
	h := TestHandler{}
	ts := TestServer(&h)
	defer ts.Close()
	_, err := http.Get(ts.URL + "/test-server")
	c.Assert(err, gocheck.IsNil)
	c.Assert(h.Url, gocheck.Equals, "/test-server")
}
