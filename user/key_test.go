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
	c.Assert(authKey, Equals, expected)
}

func (s *S) TestShouldAddKeyWithoutError(c *C) {
	key := &Key{Content: "somekey blaaaaaaa r2d2@host", Name: "somekey"}
	err := addKey(key, "someuser")
	c.Assert(err, IsNil)
}

func (s *S) TestShouldWriteKeyInFile(c *C) {
	key := &Key{Content: "somekey blaaaaaaa r2d2@host", Name: "somekey"}
	err := addKey(key, "someuser")
	c.Assert(err, IsNil)
	f, err := s.rfs.OpenFile(authKey, os.O_RDWR, 0755)
	b, err := ioutil.ReadAll(f)
	c.Assert(err, IsNil)
	got := string(b)
	c.Assert(got, Equals, formatKey(key.Content, "someuser"))
}

func (s *S) TestShouldAppendKeyInFile(c *C) {
	key1 := &Key{Content: "somekey blaaaaaaa r2d2@host", Name: "somekey"}
	err := addKey(key1, "someuser")
	c.Assert(err, IsNil)
	key2 := &Key{Content: "somekey fooo r2d2@host", Name: "somekey"}
	err = addKey(key2, "someuser")
	c.Assert(err, IsNil)
	f, err := s.rfs.OpenFile(authKey, os.O_RDWR, 0755)
	c.Assert(err, IsNil)
	b, err := ioutil.ReadAll(f)
	c.Assert(err, IsNil)
	got := string(b)
	expected := fmt.Sprintf(".*%s\n.*%s", key1.Content, key2.Content)
	c.Assert(got, Matches, expected)
}

func (s *S) TestAddShouldWrapKeyWithRestrictions(c *C) {
	key := &Key{Content: "somekey bleeeerh r2d2@host", Name: "somekey"}
	expected := fmt.Sprintf("no-port-forwarding,no-X11-forwarding,no-agent-forwarding,no-pty,command=.* %s", key.Content)
	err := addKey(key, "someuser")
	c.Assert(err, IsNil)
	f, err := s.rfs.OpenFile(authKey, os.O_RDWR, 0755)
	c.Assert(err, IsNil)
	b, err := ioutil.ReadAll(f)
	c.Assert(err, IsNil)
	got := string(b)
	c.Assert(got, Matches, expected)
}

func (s *S) TestaddKeysShouldWriteToAuthorizedKeysFile(c *C) {
	key := Key{Content: "ssh-rsa mykey pippin@nowhere", Name: "somekey"}
	err := addKeys([]Key{key}, "someuser")
	c.Assert(err, IsNil)
	keys := s.authKeysContent(c)
	c.Assert(keys, Matches, ".*ssh-rsa mykey pippin@nowhere")
}

func (s *S) TestremoveKeysShouldRemoveKeysFromAuthorizedKeys(c *C) {
	key := Key{Content: "ssh-rsa mykey pippin@nowhere", Name: "somekey"}
	err := removeKeys([]Key{key}, "someuser")
	c.Assert(err, IsNil)
	keys := s.authKeysContent(c)
	c.Assert(keys, Equals, "")
}

func (s *S) TestRemoveKey(c *C) {
	key1 := &Key{Content: "somekey blaaaaaaa r2d2@host", Name: "somekey"}
	err := addKey(key1, "someuser")
	c.Assert(err, IsNil)
	key2 := &Key{Content: "someotherkey fooo r2d2@host", Name: "somekey"}
	err = addKey(key2, "someuser")
	c.Assert(err, IsNil)
	err = removeKey(key1.Content, "someuser")
	f, err := s.rfs.OpenFile(authKey, os.O_RDWR, 0755)
	c.Assert(err, IsNil)
	b, err := ioutil.ReadAll(f)
	c.Assert(err, IsNil)
	got := string(b)
	expected := formatKey(key2.Content, "someuser")
	c.Assert(got, Matches, expected)
	expected = formatKey(key1.Content, "someuser")
	c.Assert(got, Not(Matches), expected)
}

func (s *S) TestRemoveWhenKeyDoesNotExists(c *C) {
	err := removeKey("somekey blaaaaaaa r2d2@host", "anotheruser")
	c.Assert(err, IsNil)
	f, err := s.rfs.OpenFile(authKey, os.O_RDWR, 0755)
	c.Assert(err, IsNil)
	b, err := ioutil.ReadAll(f)
	c.Assert(err, IsNil)
	got := string(b)
	c.Assert(got, Equals, "")
}

func (s *S) TestRemoveWhenExistsOnlyOneKey(c *C) {
	key := &Key{Content: "somekey blaaaaaaa r2d2@host", Name: "somekey"}
	err := addKey(key, "someuser")
	c.Assert(err, IsNil)
	err = removeKey(key.Content, "someuser")
	c.Assert(err, IsNil)
	f, err := s.rfs.OpenFile(authKey, os.O_RDWR, 0755)
	c.Assert(err, IsNil)
	b, err := ioutil.ReadAll(f)
	c.Assert(err, IsNil)
	got := string(b)
	c.Assert(got, Equals, "")
}

func (s *S) TestFormatKeyShouldAddSshLoginRestrictionsAtBegining(c *C) {
	key := &Key{Content: "somekeeey fooo bar@bar.com", Name: "somekey"}
	got := formatKey(key.Content, "someuser")
	expected := fmt.Sprintf("no-port-forwarding,no-X11-forwarding,no-agent-forwarding,no-pty,command=.* %s", key.Content)
	c.Assert(got, Matches, expected)
}

func (s *S) TestFormatKeyShouldAddCommandAfterSshRestrictions(c *C) {
	key := &Key{Content: "somekeyyyy fooow bar@bar.com", Name: "somekey"}
	got := formatKey(key.Content, "brain")
	p, err := config.GetString("bin-path")
	c.Assert(err, IsNil)
	expected := fmt.Sprintf(`no-port-forwarding,no-X11-forwarding,no-agent-forwarding,no-pty,command="%s brain" %s`, p, key.Content)
	c.Assert(got, Equals, expected)
}

func (s *S) TestFormatKeyShouldGetCommandPathFromGandalfConf(c *C) {
	oldConf, err := config.GetString("bin-path")
	c.Assert(err, IsNil)
	config.Set("bin-path", "/foo/bar/hi.go")
	defer config.Set("bin-path", oldConf)
	key := &Key{Content: "lol loool bar@bar.com", Name: "test"}
	got := formatKey(key.Content, "dash")
	expected := fmt.Sprintf(`no-port-forwarding,no-X11-forwarding,no-agent-forwarding,no-pty,command="/foo/bar/hi.go dash" %s`, key.Content)
	c.Assert(got, Equals, expected)
}

func (s *S) TestFormatKeyShouldAppendUserNameAsCommandParameter(c *C) {
	key := &Key{Content: "ssh-rsa fueeel bar@bar.com", Name: "somekey"}
	p, err := config.GetString("bin-path")
	c.Assert(err, IsNil)
	got := formatKey(key.Content, "someuser")
	expected := fmt.Sprintf(`no-port-forwarding,no-X11-forwarding,no-agent-forwarding,no-pty,command="%s someuser" %s`, p, key.Content)
	c.Assert(got, Equals, expected)
}
