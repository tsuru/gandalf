package repository

import (
	"github.com/timeredbull/commandmocker"
    "github.com/timeredbull/config"
	. "launchpad.net/gocheck"
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

func (s *S) TestCreateBareShouldCreateADir(c *C) {
	dir, err := commandmocker.Add("git", "$*")
	c.Check(err, IsNil)
	defer commandmocker.Remove(dir)
	err = newBare("myBare")
	c.Assert(err, IsNil)
	c.Assert(commandmocker.Ran(dir), Equals, true)
}

func (s *S) TestCreateBareShouldReturnMeaningfullErrorWhenBareCreationFails(c *C) {
	dir, err := commandmocker.Error("git", "ooooi", 1)
	c.Assert(err, IsNil)
	defer commandmocker.Remove(dir)
	err = newBare("foo")
	c.Check(err, NotNil)
	got := err.Error()
	expected := "Could not create git bare repository: exit status 1"
	c.Assert(got, Equals, expected)
}
