// Copyright 2013 gandalf authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package user

import (
	"fmt"
	"github.com/globocom/config"
	"io/ioutil"
	. "launchpad.net/gocheck"
	"os"
	"path"
)

func (s *S) TestAuthKeysShouldBeAbsolutePathToUsersAuthorizedKeysByDefault(c *C) {
	home := os.Getenv("HOME")
	expected := path.Join(home, ".ssh", "authorized_keys")
	c.Assert(authKey(), Equals, expected)
}

func (s *S) TestShouldAddKeyWithoutError(c *C) {
	err := addKey("somekey blaaaaaaa r2d2@host", "someuser")
	c.Assert(err, IsNil)
}

func (s *S) TestShouldWriteKeyInFile(c *C) {
	key := "somekey blaaaaaaa r2d2@host"
	err := addKey(key, "someuser")
	c.Assert(err, IsNil)
	f, err := s.rfs.OpenFile(authKey(), os.O_RDWR, 0755)
	b, err := ioutil.ReadAll(f)
	c.Assert(err, IsNil)
	got := string(b)
	c.Assert(got, Equals, formatKey(key, "someuser"))
}

func (s *S) TestShouldAppendKeyInFile(c *C) {
	key1 := "somekey blaaaaaaa r2d2@host"
	err := addKey(key1, "someuser")
	c.Assert(err, IsNil)
	key2 := "somekey foo r2d2@host"
	err = addKey(key2, "someuser")
	c.Assert(err, IsNil)
	f, err := s.rfs.OpenFile(authKey(), os.O_RDWR, 0755)
	c.Assert(err, IsNil)
	b, err := ioutil.ReadAll(f)
	c.Assert(err, IsNil)
	got := string(b)
	expected := fmt.Sprintf(".*%s\n.*%s", key1, key2)
	c.Assert(got, Matches, expected)
}

func (s *S) TestAddShouldWrapKeyWithRestrictions(c *C) {
	key := "somekey bleeeerh r2d2@host"
	expected := fmt.Sprintf("no-port-forwarding,no-X11-forwarding,no-agent-forwarding,no-pty,command=.* %s", key)
	err := addKey(key, "someuser")
	c.Assert(err, IsNil)
	f, err := s.rfs.OpenFile(authKey(), os.O_RDWR, 0755)
	c.Assert(err, IsNil)
	b, err := ioutil.ReadAll(f)
	c.Assert(err, IsNil)
	got := string(b)
	c.Assert(got, Matches, expected)
}

func (s *S) TestaddKeysShouldWriteToAuthorizedKeysFile(c *C) {
	key := map[string]string{"somekey": "ssh-rsa mykey pippin@nowhere"}
	err := addKeys(key, "someuser")
	c.Assert(err, IsNil)
	keys := s.authKeysContent(c)
	c.Assert(keys, Matches, ".*ssh-rsa mykey pippin@nowhere")
}

func (s *S) TestaddDuplicateKeyShouldReturnError(c *C) {
	keys := map[string]string{
		"somekey": "ssh-rsa mykey pippin@nowhere",
		"double":  "ssh-rsa mykey pippin@nowhere",
	}
	err := addKeys(keys, "someuser")
	c.Assert(err, ErrorMatches, "Key already exists.")
}

func (s *S) TestremoveKeysShouldRemoveKeysFromAuthorizedKeys(c *C) {
	key := map[string]string{"somekey": "ssh-rsa mykey pippin@nowhere"}
	err := removeKeys(key, "someuser")
	c.Assert(err, IsNil)
	keys := s.authKeysContent(c)
	c.Assert(keys, Equals, "")
}

func (s *S) TestRemoveKey(c *C) {
	key1 := "somekey blaaaaaaa r2d2@host"
	err := addKey(key1, "someuser")
	c.Assert(err, IsNil)
	key2 := "someotherkey fooo r2d2@host"
	err = addKey(key2, "someuser")
	c.Assert(err, IsNil)
	err = removeKey(key1, "someuser")
	f, err := s.rfs.OpenFile(authKey(), os.O_RDWR, 0755)
	c.Assert(err, IsNil)
	b, err := ioutil.ReadAll(f)
	c.Assert(err, IsNil)
	got := string(b)
	expected := formatKey(key2, "someuser")
	c.Assert(got, Matches, expected)
	expected = formatKey(key1, "someuser")
	c.Assert(got, Not(Matches), expected)
}

func (s *S) TestRemoveWhenKeyDoesNotExists(c *C) {
	err := removeKey("somekey blaaaaaaa r2d2@host", "anotheruser")
	c.Assert(err, IsNil)
	f, err := s.rfs.OpenFile(authKey(), os.O_RDWR, 0755)
	c.Assert(err, IsNil)
	b, err := ioutil.ReadAll(f)
	c.Assert(err, IsNil)
	got := string(b)
	c.Assert(got, Equals, "")
}

func (s *S) TestRemoveWhenExistsOnlyOneKey(c *C) {
	key := "somekey blaaaaaaa r2d2@host"
	err := addKey(key, "someuser")
	c.Assert(err, IsNil)
	err = removeKey(key, "someuser")
	c.Assert(err, IsNil)
	f, err := s.rfs.OpenFile(authKey(), os.O_RDWR, 0755)
	c.Assert(err, IsNil)
	b, err := ioutil.ReadAll(f)
	c.Assert(err, IsNil)
	got := string(b)
	c.Assert(got, Equals, "")
}

func (s *S) TestFormatKeyShouldAddSshLoginRestrictionsAtBegining(c *C) {
	key := "somekey fooo bar@bar.com"
	got := formatKey(key, "someuser")
	expected := fmt.Sprintf("no-port-forwarding,no-X11-forwarding,no-agent-forwarding,no-pty,command=.* %s", key)
	c.Assert(got, Matches, expected)
}

func (s *S) TestFormatKeyShouldAddCommandAfterSshRestrictions(c *C) {
	key := "somekeyyyy fooow bar@bar.com"
	got := formatKey(key, "brain")
	p, err := config.GetString("bin-path")
	c.Assert(err, IsNil)
	expected := fmt.Sprintf(`no-port-forwarding,no-X11-forwarding,no-agent-forwarding,no-pty,command="%s brain" %s`, p, key)
	c.Assert(got, Equals, expected)
}

func (s *S) TestFormatKeyShouldGetCommandPathFromGandalfConf(c *C) {
	oldConf, err := config.GetString("bin-path")
	c.Assert(err, IsNil)
	config.Set("bin-path", "/foo/bar/hi.go")
	defer config.Set("bin-path", oldConf)
	key := "lol loool bar@bar.com"
	got := formatKey(key, "dash")
	expected := fmt.Sprintf(`no-port-forwarding,no-X11-forwarding,no-agent-forwarding,no-pty,command="/foo/bar/hi.go dash" %s`, key)
	c.Assert(got, Equals, expected)
}

func (s *S) TestFormatKeyShouldAppendUserNameAsCommandParameter(c *C) {
	key := "ssh-rsa fueeel bar@bar.com"
	p, err := config.GetString("bin-path")
	c.Assert(err, IsNil)
	got := formatKey(key, "someuser")
	expected := fmt.Sprintf(`no-port-forwarding,no-X11-forwarding,no-agent-forwarding,no-pty,command="%s someuser" %s`, p, key)
	c.Assert(got, Equals, expected)
}

func (s *S) TestMergeMaps(c *C) {
	m1 := map[string]string{"foo": "bar"}
	m2 := map[string]string{"bar": "foo"}
	m3 := mergeMaps(m1, m2)
	c.Assert(m3, DeepEquals, map[string]string{"foo": "bar", "bar": "foo"})
}
