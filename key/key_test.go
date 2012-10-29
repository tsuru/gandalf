package key

import (
	"fmt"
	"github.com/globocom/config"
	"github.com/globocom/tsuru/fs"
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

func (s *S) SetUpSuite(c *C) {
	s.origKeyFile = authKey
	err := config.ReadConfigFile("../etc/gandalf.conf")
	c.Check(err, IsNil)
	s.rfs = &fstesting.RecordingFs{}
	fsystem = s.rfs
}

func (s *S) TearDownSuite(c *C) {
	fsystem = nil
}

func (s *S) SetUpTest(c *C) {
	changeAuthKey()
}

func (s *S) TearDownTest(c *C) {
	ok := clearAuthKeyFile()
	c.Assert(ok, Equals, true)
}

func changeAuthKey() {
	authKey = "testdata/authorized_keys"
}

func clearAuthKeyFile() bool {
	f, err := filesystem().OpenFile(authKey, os.O_RDWR, 0755)
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
	expected := path.Join(home, "authorized_keys")
	c.Assert(s.origKeyFile, Equals, expected)
}

func (s *S) TestShouldAddKeyWithoutError(c *C) {
	key := "somekey blaaaaaaa r2d2@host"
	err := Add(key)
	c.Assert(err, IsNil)
}

func (s *S) TestShouldWriteKeyInFile(c *C) {
	key := "somekey blaaaaaaa r2d2@host"
	err := Add(key)
	c.Assert(err, IsNil)
	f, err := filesystem().OpenFile(authKey, os.O_RDWR, 0755)
	b, err := ioutil.ReadAll(f)
	c.Assert(err, IsNil)
	got := string(b)
	c.Assert(got, Equals, formatKey(key))
}

func (s *S) TestShouldAppendKeyInFile(c *C) {
	key1 := "somekey blaaaaaaa r2d2@host"
	err := Add(key1)
	c.Assert(err, IsNil)
	key2 := "someotherkey fooo r2d2@host"
	err = Add(key2)
	c.Assert(err, IsNil)
	f, err := filesystem().OpenFile(authKey, os.O_RDWR, 0755)
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
	err := Add(key)
	c.Assert(err, IsNil)
	f, err := filesystem().OpenFile(authKey, os.O_RDWR, 0755)
	c.Assert(err, IsNil)
	b, err := ioutil.ReadAll(f)
	c.Assert(err, IsNil)
	got := string(b)
	c.Assert(got, Matches, expected)
}

func (s *S) TestRemoveKey(c *C) {
	key1 := "somekey blaaaaaaa r2d2@host"
	err := Add(key1)
	c.Assert(err, IsNil)
	key2 := "someotherkey fooo r2d2@host"
	err = Add(key2)
	c.Assert(err, IsNil)
	err = Remove(key1)
	f, err := filesystem().OpenFile(authKey, os.O_RDWR, 0755)
	c.Assert(err, IsNil)
	b, err := ioutil.ReadAll(f)
	c.Assert(err, IsNil)
	got := string(b)
	expected := fmt.Sprintf(".*%s", key2)
	c.Assert(got, Matches, expected)
}

func (s *S) TestRemoveWhenKeyDoesNotExists(c *C) {
	key1 := "somekey blaaaaaaa r2d2@host"
	err := Remove(key1)
	c.Assert(err, IsNil)
	f, err := filesystem().OpenFile(authKey, os.O_RDWR, 0755)
	c.Assert(err, IsNil)
	b, err := ioutil.ReadAll(f)
	c.Assert(err, IsNil)
	got := string(b)
	c.Assert(got, Equals, "")
}

func (s *S) TestRemoveWhenExistsOnlyOneKey(c *C) {
	key1 := "somekey blaaaaaaa r2d2@host"
	err := Add(key1)
	c.Assert(err, IsNil)
	err = Remove(key1)
	c.Assert(err, IsNil)
	f, err := filesystem().OpenFile(authKey, os.O_RDWR, 0755)
	c.Assert(err, IsNil)
	b, err := ioutil.ReadAll(f)
	c.Assert(err, IsNil)
	got := string(b)
	c.Assert(got, Equals, "")
}

func (s *S) TestFormatKeyShouldAddSshLoginRestrictionsAtBegining(c *C) {
	key := "somekeeey fooo bar@bar.com"
	got := formatKey(key)
	expected := fmt.Sprintf("no-port-forwarding,no-X11-forwarding,no-agent-forwarding,no-pty,command=.* %s", key)
	c.Assert(got, Matches, expected)
}

func (s *S) TestFormatKeyShouldAddCommandAfterSshRestrictions(c *C) {
	key := "somekeyyyy fooow bar@bar.com"
	got := formatKey(key)
	expected := fmt.Sprintf(`no-port-forwarding,no-X11-forwarding,no-agent-forwarding,no-pty,command="/usr/local/bin/gandalf.go" %s`, key)
	c.Assert(got, Equals, expected)
}

func (s *S) TestFormatKeyShouldGetCommandPathFromGandalfConf(c *C) {
	config.Set("bin-path", "/foo/bar/hi.go")
	key := "lol loool bar@bar.com"
	got := formatKey(key)
	expected := fmt.Sprintf(`no-port-forwarding,no-X11-forwarding,no-agent-forwarding,no-pty,command="/foo/bar/hi.go" %s`, key)
	c.Assert(got, Equals, expected)
}

func (s *S) TestFsystemShouldSetGlobalFsystemWhenItsNil(c *C) {
	oldFs := fsystem
	fsystem = nil
	defer func() { fsystem = oldFs }()
	fsys := filesystem()
	_, ok := fsys.(fs.Fs)
	c.Assert(ok, Equals, true)
}
