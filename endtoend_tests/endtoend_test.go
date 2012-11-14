package endtoend_test

import (
	"flag"
	. "launchpad.net/gocheck"
	"testing"
)

func Test(t *testing.T) { TestingT(t) }

type S struct{}

var _ = Suite(&S{})

var flagDesc = "enable end-to-end tests that hits gandalf's server to try it's api, it's needed to configure the GANDALF_SERVER environment variable"
var enableSuite = flag.Bool("endtoend", false, flagDesc)

func (s *S) SetUpSuite(c *C) {
	if !*enableSuite {
		c.Skip("skipping end-to-end suite, use -endtoend to enable")
	}
}

func (s *S) TestCreatesUser(c *C) {
	c.Assert(true, Equals, true)
}
