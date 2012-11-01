package key

import (
	"fmt"
	"github.com/globocom/config"
	fstesting "github.com/globocom/tsuru/fs/testing"
	"io/ioutil"
	. "launchpad.net/gocheck"
	"os"
	"path"
	"strings"
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
	fsystem = s.rfs
}

func (s *S) TearDownSuite(c *C) {
	fsystem = nil
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
	key := "somekey blaaaaaaa r2d2@host"
	err := Add(key, "someuser", s.rfs)
	c.Assert(err, IsNil)
}

func (s *S) TestShouldWriteKeyInFile(c *C) {
	key := "somekey blaaaaaaa r2d2@host"
	err := Add(key, "someuser", s.rfs)
	c.Assert(err, IsNil)
	f, err := s.rfs.OpenFile(authKey, os.O_RDWR, 0755)
	b, err := ioutil.ReadAll(f)
	c.Assert(err, IsNil)
	got := string(b)
	c.Assert(got, Equals, formatKey(key, "someuser"))
}

func (s *S) TestShouldAppendKeyInFile(c *C) {
	key1 := "somekey blaaaaaaa r2d2@host"
	err := Add(key1, "someuser", s.rfs)
	c.Assert(err, IsNil)
	key2 := "someotherkey fooo r2d2@host"
	err = Add(key2, "someuser", s.rfs)
	c.Assert(err, IsNil)
	f, err := s.rfs.OpenFile(authKey, os.O_RDWR, 0755)
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
	err := Add(key, "someuser", s.rfs)
	c.Assert(err, IsNil)
	f, err := s.rfs.OpenFile(authKey, os.O_RDWR, 0755)
	c.Assert(err, IsNil)
	b, err := ioutil.ReadAll(f)
	c.Assert(err, IsNil)
	got := string(b)
	c.Assert(got, Matches, expected)
}

func (s *S) TestBulkActionAddKeys(c *C) {
	key1 := "ssh-rsa mykey pippin@nowhere"
	key2 := "ssh-rsa myotherkey pippin@somewhere"
	keys := []string{key1, key2}
	err := bulkAction(Add, keys, "someuser", s.rfs)
	c.Assert(err, IsNil)
	got := strings.Replace(s.authKeysContent(c), "\n", " ", -1)
	c.Assert(got, Matches, ".*"+key1+".*")
	c.Assert(got, Matches, ".*"+key2+".*")
}

func (s *S) TestBulkActionRemoveKeys(c *C) {
	key1 := "ssh-rsa mykey pippin@nowhere"
	key2 := "ssh-rsa myotherkey pippin@somewhere"
	keys := []string{key1, key2}
	err := bulkAction(Add, keys, "someuser", s.rfs)
	got := strings.Replace(s.authKeysContent(c), "\n", " ", -1)
	c.Assert(got, Matches, ".*"+key1+".*")
	c.Assert(got, Matches, ".*"+key2+".*")
	err = bulkAction(Remove, keys, "someuser", s.rfs)
	c.Assert(err, IsNil)
	got = strings.Replace(s.authKeysContent(c), "\n", " ", -1)
	c.Assert(got, Not(Matches), ".*"+key1+".*")
	c.Assert(got, Not(Matches), ".*"+key2+".*")
}

func (s *S) TestBulkAddShouldWriteToAuthorizedKeysFile(c *C) {
	err := BulkAdd([]string{"ssh-rsa mykey pippin@nowhere"}, "someuser", s.rfs)
	c.Assert(err, IsNil)
	keys := s.authKeysContent(c)
	c.Assert(keys, Matches, ".*ssh-rsa mykey pippin@nowhere")
}

func (s *S) TestBulkRemoveShouldRemoveKeysFromAuthorizedKeys(c *C) {
	key := "ssh-rsa mykey pippin@nowhere"
	err := BulkRemove([]string{key}, "someuser", s.rfs)
	c.Assert(err, IsNil)
	keys := s.authKeysContent(c)
	c.Assert(keys, Equals, "")
}

func (s *S) TestRemoveKey(c *C) {
	key1 := "somekey blaaaaaaa r2d2@host"
	err := Add(key1, "someuser", s.rfs)
	c.Assert(err, IsNil)
	key2 := "someotherkey fooo r2d2@host"
	err = Add(key2, "someuser", s.rfs)
	c.Assert(err, IsNil)
	err = Remove(key1, "someuser", s.rfs)
	f, err := s.rfs.OpenFile(authKey, os.O_RDWR, 0755)
	c.Assert(err, IsNil)
	b, err := ioutil.ReadAll(f)
	c.Assert(err, IsNil)
	got := string(b)
	expected := fmt.Sprintf(".*%s", key2)
	c.Assert(got, Matches, expected)
}

func (s *S) TestRemoveWhenKeyDoesNotExists(c *C) {
	key1 := "somekey blaaaaaaa r2d2@host"
	err := Remove(key1, "anotheruser", s.rfs)
	c.Assert(err, IsNil)
	f, err := s.rfs.OpenFile(authKey, os.O_RDWR, 0755)
	c.Assert(err, IsNil)
	b, err := ioutil.ReadAll(f)
	c.Assert(err, IsNil)
	got := string(b)
	c.Assert(got, Equals, "")
}

func (s *S) TestRemoveWhenExistsOnlyOneKey(c *C) {
	key1 := "somekey blaaaaaaa r2d2@host"
	err := Add(key1, "someuser", s.rfs)
	c.Assert(err, IsNil)
	err = Remove(key1, "someuser", s.rfs)
	c.Assert(err, IsNil)
	f, err := s.rfs.OpenFile(authKey, os.O_RDWR, 0755)
	c.Assert(err, IsNil)
	b, err := ioutil.ReadAll(f)
	c.Assert(err, IsNil)
	got := string(b)
	c.Assert(got, Equals, "")
}

func (s *S) TestFormatKeyShouldAddSshLoginRestrictionsAtBegining(c *C) {
	key := "somekeeey fooo bar@bar.com"
	got := formatKey(key, "brain")
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
