package key

import (
	"fmt"
	"io/ioutil"
	. "launchpad.net/gocheck"
	"os"
	"path"
	"testing"
)

func Test(t *testing.T) { TestingT(t) }

type S struct{}

var _ = Suite(&S{})

func changeAuthKey() {
	authKey = "testdata/authorized_keys"
}

func clearAuthKeyFile() bool {
	err := os.Truncate(authKey, 0)
	if err != nil {
		return false
	}
	return true
}

func (s *S) TestAuthKeysShouldBeAbsolutePathToUsersAuthorizedKeys(c *C) {
	home := os.Getenv("HOME")
	expected := path.Join(home, "authorized_keys")
	c.Assert(authKey, Equals, expected)
}

func (s *S) TestShouldAddKeyWithoutError(c *C) {
	changeAuthKey()
	key := "somekey blaaaaaaa r2d2@host"
	err := Add(key)
	c.Assert(err, IsNil)
	ok := clearAuthKeyFile()
	c.Assert(ok, Equals, true)
}

func (s *S) TestShouldWriteKeyInFile(c *C) {
	changeAuthKey()
	key := "somekey blaaaaaaa r2d2@host"
	err := Add(key)
	c.Assert(err, IsNil)
	b, err := ioutil.ReadFile(authKey)
	c.Assert(err, IsNil)
	got := string(b)
	c.Assert(got, Equals, key)
	ok := clearAuthKeyFile()
	c.Assert(ok, Equals, true)
}

func (s *S) TestShouldAppendKeyInFile(c *C) {
	changeAuthKey()
	key1 := "somekey blaaaaaaa r2d2@host"
	err := Add(key1)
	c.Assert(err, IsNil)
	key2 := "someotherkey fooo r2d2@host"
	err = Add(key2)
	c.Assert(err, IsNil)
	b, err := ioutil.ReadFile(authKey)
	c.Assert(err, IsNil)
	got := string(b)
	expected := fmt.Sprintf("%s\n%s", key1, key2)
	c.Assert(got, Equals, expected)
	ok := clearAuthKeyFile()
	c.Assert(ok, Equals, true)
}
