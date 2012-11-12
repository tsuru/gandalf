package key

import (
	"fmt"
	"github.com/globocom/config"
	"github.com/globocom/gandalf/fs"
	fstesting "github.com/globocom/tsuru/fs/testing"
	"io/ioutil"
	. "launchpad.net/gocheck"
	"os"
	"path"
	"testing"
)

func Test(t *testing.T) { TestingT(t) }

type S struct {
	origKeyFile string
	rfs         *fstesting.RecordingFs
}

var _ = Suite(&S{})

func (s *S) authKeysContent(c *C) string {
	f, err := s.rfs.OpenFile(authKey, os.O_RDWR, 0755)
	c.Assert(err, IsNil)
	b, err := ioutil.ReadAll(f)
	c.Assert(err, IsNil)
	return string(b)
}

func (s *S) SetUpSuite(c *C) {
	err := config.ReadConfigFile("../etc/gandalf.conf")
	c.Check(err, IsNil)
}

func (s *S) SetUpTest(c *C) {
	s.rfs = &fstesting.RecordingFs{}
	fs.Fsystem = s.rfs
}

func (s *S) TearDownSuite(c *C) {
	fs.Fsystem = nil
}

func (s *S) TearDownTest(c *C) {
	ok := s.clearAuthKeyFile()
	c.Assert(ok, Equals, true)
}

func (s *S) clearAuthKeyFile() bool {
	f, err := s.rfs.OpenFile(authKey, os.O_RDWR, 0755)
	if err != nil {
		return false
	}
	if err := f.Truncate(0); err != nil {
		return false
	}
	return true
}

func (s *S) TestAuthKeysShouldBeAbsolutePathToUsersAuthorizedKeysByDefault(c *C) {
	home := os.Getenv("HOME")
	expected := path.Join(home, ".ssh", "authorized_keys")
	c.Assert(authKey, Equals, expected)
}

func (s *S) TestShouldAddKeyWithoutError(c *C) {
	key := &Key{Content: "somekey blaaaaaaa r2d2@host", User: "someuser", Name: "somekey"}
	err := Add(key)
	c.Assert(err, IsNil)
}

func (s *S) TestShouldWriteKeyInFile(c *C) {
	key := &Key{Content: "somekey blaaaaaaa r2d2@host", User: "someuser", Name: "somekey"}
	err := Add(key)
	c.Assert(err, IsNil)
	f, err := s.rfs.OpenFile(authKey, os.O_RDWR, 0755)
	b, err := ioutil.ReadAll(f)
	c.Assert(err, IsNil)
	got := string(b)
	c.Assert(got, Equals, formatKey(key))
}

func (s *S) TestShouldAppendKeyInFile(c *C) {
	key1 := &Key{Content: "somekey blaaaaaaa r2d2@host", User: "someuser", Name: "somekey"}
	err := Add(key1)
	c.Assert(err, IsNil)
	key2 := &Key{Content: "somekey fooo r2d2@host", User: "someuser", Name: "somekey"}
	err = Add(key2)
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
	key := &Key{Content: "somekey bleeeerh r2d2@host", User: "someuser", Name: "somekey"}
	expected := fmt.Sprintf("no-port-forwarding,no-X11-forwarding,no-agent-forwarding,no-pty,command=.* %s", key.Content)
	err := Add(key)
	c.Assert(err, IsNil)
	f, err := s.rfs.OpenFile(authKey, os.O_RDWR, 0755)
	c.Assert(err, IsNil)
	b, err := ioutil.ReadAll(f)
	c.Assert(err, IsNil)
	got := string(b)
	c.Assert(got, Matches, expected)
}

func (s *S) TestBulkAddShouldWriteToAuthorizedKeysFile(c *C) {
	key := &Key{Content: "ssh-rsa mykey pippin@nowhere", User: "someuser", Name: "somekey"}
	err := BulkAdd([]*Key{key})
	c.Assert(err, IsNil)
	keys := s.authKeysContent(c)
	c.Assert(keys, Matches, ".*ssh-rsa mykey pippin@nowhere")
}

func (s *S) TestBulkRemoveShouldRemoveKeysFromAuthorizedKeys(c *C) {
	key := &Key{Content: "ssh-rsa mykey pippin@nowhere", User: "someuser", Name: "somekey"}
	err := BulkRemove([]*Key{key})
	c.Assert(err, IsNil)
	keys := s.authKeysContent(c)
	c.Assert(keys, Equals, "")
}

func (s *S) TestRemoveKey(c *C) {
	key1 := &Key{Content: "somekey blaaaaaaa r2d2@host", User: "someuser", Name: "somekey"}
	err := Add(key1)
	c.Assert(err, IsNil)
	key2 := &Key{Content: "someotherkey fooo r2d2@host", User: "someuser", Name: "somekey"}
	err = Add(key2)
	c.Assert(err, IsNil)
	err = Remove(key1.Content, key1.User)
	f, err := s.rfs.OpenFile(authKey, os.O_RDWR, 0755)
	c.Assert(err, IsNil)
	b, err := ioutil.ReadAll(f)
	c.Assert(err, IsNil)
	got := string(b)
	expected := formatKey(key2)
	c.Assert(got, Matches, expected)
	expected = formatKey(key1)
	c.Assert(got, Not(Matches), expected)
}

func (s *S) TestRemoveWhenKeyDoesNotExists(c *C) {
	err := Remove("somekey blaaaaaaa r2d2@host", "anotheruser")
	c.Assert(err, IsNil)
	f, err := s.rfs.OpenFile(authKey, os.O_RDWR, 0755)
	c.Assert(err, IsNil)
	b, err := ioutil.ReadAll(f)
	c.Assert(err, IsNil)
	got := string(b)
	c.Assert(got, Equals, "")
}

func (s *S) TestRemoveWhenExistsOnlyOneKey(c *C) {
	key := &Key{Content: "somekey blaaaaaaa r2d2@host", User: "someuser", Name: "somekey"}
	err := Add(key)
	c.Assert(err, IsNil)
	err = Remove(key.Content, "someuser")
	c.Assert(err, IsNil)
	f, err := s.rfs.OpenFile(authKey, os.O_RDWR, 0755)
	c.Assert(err, IsNil)
	b, err := ioutil.ReadAll(f)
	c.Assert(err, IsNil)
	got := string(b)
	c.Assert(got, Equals, "")
}

func (s *S) TestFormatKeyShouldAddSshLoginRestrictionsAtBegining(c *C) {
	key := &Key{Content: "somekeeey fooo bar@bar.com", Name: "somekey", User: "someuser"}
	got := formatKey(key)
	expected := fmt.Sprintf("no-port-forwarding,no-X11-forwarding,no-agent-forwarding,no-pty,command=.* %s", key.Content)
	c.Assert(got, Matches, expected)
}

func (s *S) TestFormatKeyShouldAddCommandAfterSshRestrictions(c *C) {
	key := &Key{Content: "somekeyyyy fooow bar@bar.com", Name: "somekey", User: "brain"}
	got := formatKey(key)
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
	key := &Key{Content: "lol loool bar@bar.com", User: "dash", Name: "test"}
	got := formatKey(key)
	expected := fmt.Sprintf(`no-port-forwarding,no-X11-forwarding,no-agent-forwarding,no-pty,command="/foo/bar/hi.go dash" %s`, key.Content)
	c.Assert(got, Equals, expected)
}

func (s *S) TestFormatKeyShouldAppendUserNameAsCommandParameter(c *C) {
	key := &Key{Content: "ssh-rsa fueeel bar@bar.com", User: "someuser", Name: "somekey"}
	p, err := config.GetString("bin-path")
	c.Assert(err, IsNil)
	got := formatKey(key)
	expected := fmt.Sprintf(`no-port-forwarding,no-X11-forwarding,no-agent-forwarding,no-pty,command="%s someuser" %s`, p, key.Content)
	c.Assert(got, Equals, expected)
}
