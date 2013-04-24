// Copyright 2013 gandalf authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package user

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/globocom/config"
	"github.com/globocom/gandalf/db"
	"io"
	"io/ioutil"
	"labix.org/v2/mgo/bson"
	. "launchpad.net/gocheck"
	"os"
	"path"
)

type shortWriter struct{}

func (shortWriter) Write(p []byte) (int, error) {
	return len(p) / 2, nil
}

type failWriter struct{}

func (failWriter) Write(p []byte) (int, error) {
	return 0, errors.New("Failed")
}

const rawKey = "ssh-dss AAAAB3NzaC1kc3MAAACBAIHfSDLpSCfIIVEJ/Is3RFMQhsCi7WZtFQeeyfi+DzVP0NGX4j/rMoQEHgXgNlOKVCJvPk5e00tukSv6iVzJPFcozArvVaoCc5jCoDi5Ef8k3Jil4Q7qNjcoRDDyqjqLcaviJEz5GrtmqAyXEIzJ447BxeEdw3Z7UrIWYcw2YyArAAAAFQD7wiOGZIoxu4XIOoeEe5aToTxN1QAAAIAZNAbJyOnNceGcgRRgBUPfY5ChX+9A29n2MGnyJ/Cxrhuh8d7B0J8UkvEBlfgQICq1UDZbC9q5NQprwD47cGwTjUZ0Z6hGpRmEEZdzsoj9T6vkLiteKH3qLo7IPVx4mV6TTF6PWQbQMUsuxjuDErwS9nhtTM4nkxYSmUbnWb6wfwAAAIB2qm/1J6Jl8bByBaMQ/ptbm4wQCvJ9Ll9u6qtKy18D4ldoXM0E9a1q49swml5CPFGyU+cgPRhEjN5oUr5psdtaY8CHa2WKuyIVH3B8UhNzqkjpdTFSpHs6tGluNVC+SQg1MVwfG2wsZUdkUGyn+6j8ZZarUfpAmbb5qJJpgMFEKQ== f@xikinbook.local"
const body = "ssh-dss AAAAB3NzaC1kc3MAAACBAIHfSDLpSCfIIVEJ/Is3RFMQhsCi7WZtFQeeyfi+DzVP0NGX4j/rMoQEHgXgNlOKVCJvPk5e00tukSv6iVzJPFcozArvVaoCc5jCoDi5Ef8k3Jil4Q7qNjcoRDDyqjqLcaviJEz5GrtmqAyXEIzJ447BxeEdw3Z7UrIWYcw2YyArAAAAFQD7wiOGZIoxu4XIOoeEe5aToTxN1QAAAIAZNAbJyOnNceGcgRRgBUPfY5ChX+9A29n2MGnyJ/Cxrhuh8d7B0J8UkvEBlfgQICq1UDZbC9q5NQprwD47cGwTjUZ0Z6hGpRmEEZdzsoj9T6vkLiteKH3qLo7IPVx4mV6TTF6PWQbQMUsuxjuDErwS9nhtTM4nkxYSmUbnWb6wfwAAAIB2qm/1J6Jl8bByBaMQ/ptbm4wQCvJ9Ll9u6qtKy18D4ldoXM0E9a1q49swml5CPFGyU+cgPRhEjN5oUr5psdtaY8CHa2WKuyIVH3B8UhNzqkjpdTFSpHs6tGluNVC+SQg1MVwfG2wsZUdkUGyn+6j8ZZarUfpAmbb5qJJpgMFEKQ==\n"
const comment = "f@xikinbook.local"
const otherKey = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCaNZSIEyP6FSdCX0WHDcUFTvebNbvqKiiLEiC7NTGvKrT15r2MtCDi4EPi4Ul+UyxWqb2D7FBnK1UmIcEFHd/ZCnBod2/FSplGOIbIb2UVVbqPX5Alv7IBCMyZJD14ex5cFh16zoqOsPOkOD803LMIlNvXPDDwKjY4TVOQV1JtA2tbZXvYUchqhTcKPxt5BDBZbeQkMMgUgHIEz6IueglFB3+dIZfrzlmM8CVSElKZOpucnJ5JOpGh3paSO/px2ZEcvY8WvjFdipvAWsis75GG/04F641I6XmYlo9fib/YytBXS23szqmvOqEqAopFnnGkDEo+LWI0+FXgPE8lc5BD"

func (s *S) TestNewKey(c *C) {
	k, err := newKey("key1", "me@tsuru.io", rawKey)
	c.Assert(err, IsNil)
	c.Assert(k.Name, Equals, "key1")
	c.Assert(k.Body, Equals, body)
	c.Assert(k.Comment, Equals, comment)
	c.Assert(k.UserName, Equals, "me@tsuru.io")
}

func (s *S) TestNewKeyInvalidKey(c *C) {
	raw := "ssh-dss ASCCDD== invalid@tsuru.io"
	k, err := newKey("key1", "me@tsuru.io", raw)
	c.Assert(k, IsNil)
	c.Assert(err, Equals, ErrInvalidKey)
}

func (s *S) TestKeyString(c *C) {
	k := Key{Body: "ssh-dss not-secret", Comment: "me@host"}
	c.Assert(k.Body+" "+k.Comment, Equals, k.String())
}

func (s *S) TestKeyStringNewLine(c *C) {
	k := Key{Body: "ssh-dss not-secret\n", Comment: "me@host"}
	c.Assert("ssh-dss not-secret me@host", Equals, k.String())
}

func (s *S) TestFormatKeyShouldAddSshLoginRestrictionsAtBegining(c *C) {
	key := Key{
		Name:     "my-key",
		Body:     "somekey\n",
		Comment:  "me@host",
		UserName: "brain",
	}
	got := key.format()
	expected := fmt.Sprintf("no-port-forwarding,no-X11-forwarding,no-agent-forwarding,no-pty,command=.* %s\n", &key)
	c.Assert(got, Matches, expected)
}

func (s *S) TestFormatKeyShouldAddCommandAfterSshRestrictions(c *C) {
	key := Key{
		Name:     "my-key",
		Body:     "somekey\n",
		Comment:  "me@host",
		UserName: "brain",
	}
	got := key.format()
	p, err := config.GetString("bin-path")
	c.Assert(err, IsNil)
	expected := fmt.Sprintf(`no-port-forwarding,no-X11-forwarding,no-agent-forwarding,no-pty,command="%s brain" %s`+"\n", p, &key)
	c.Assert(got, Equals, expected)
}

func (s *S) TestFormatKeyShouldGetCommandPathFromGandalfConf(c *C) {
	oldConf, err := config.GetString("bin-path")
	c.Assert(err, IsNil)
	config.Set("bin-path", "/foo/bar/hi.go")
	defer config.Set("bin-path", oldConf)
	key := Key{
		Name:     "my-key",
		Body:     "somekey\n",
		Comment:  "me@host",
		UserName: "dash",
	}
	got := key.format()
	expected := fmt.Sprintf(`no-port-forwarding,no-X11-forwarding,no-agent-forwarding,no-pty,command="/foo/bar/hi.go dash" %s`+"\n", &key)
	c.Assert(got, Equals, expected)
}

func (s *S) TestFormatKeyShouldAppendUserNameAsCommandParameter(c *C) {
	p, err := config.GetString("bin-path")
	c.Assert(err, IsNil)
	key := Key{
		Name:     "my-key",
		Body:     "somekey\n",
		Comment:  "me@host",
		UserName: "someuser",
	}
	got := key.format()
	expected := fmt.Sprintf(`no-port-forwarding,no-X11-forwarding,no-agent-forwarding,no-pty,command="%s someuser" %s`+"\n", p, &key)
	c.Assert(got, Equals, expected)
}

func (s *S) TestDump(c *C) {
	var buf bytes.Buffer
	key := Key{
		Name:     "my-key",
		Body:     "somekey\n",
		Comment:  "me@host",
		UserName: "someuser",
	}
	err := key.dump(&buf)
	c.Assert(err, IsNil)
	c.Assert(buf.String(), Equals, key.format())
}

func (s *S) TestDumpShortWrite(c *C) {
	key := Key{
		Name:     "my-key",
		Body:     "somekey\n",
		Comment:  "me@host",
		UserName: "someuser",
	}
	err := key.dump(shortWriter{})
	c.Assert(err, Equals, io.ErrShortWrite)
}

func (s *S) TestDumpWriteFailure(c *C) {
	key := Key{
		Name:     "my-key",
		Body:     "somekey\n",
		Comment:  "me@host",
		UserName: "someuser",
	}
	err := key.dump(failWriter{})
	c.Assert(err, NotNil)
}

func (s *S) TestAuthKeysShouldBeAbsolutePathToUsersAuthorizedKeysByDefault(c *C) {
	home := os.Getenv("HOME")
	expected := path.Join(home, ".ssh", "authorized_keys")
	c.Assert(authKey(), Equals, expected)
}

func (s *S) TestWriteKey(c *C) {
	key, err := newKey("my-key", "me@tsuru.io", rawKey)
	c.Assert(err, IsNil)
	writeKey(key)
	f, err := s.rfs.Open(authKey())
	c.Assert(err, IsNil)
	defer f.Close()
	b, err := ioutil.ReadAll(f)
	c.Assert(err, IsNil)
	got := string(b)
	c.Assert(got, Equals, key.format())
}

func (s *S) TestWriteTwoKeys(c *C) {
	key1 := Key{
		Name:     "my-key",
		Body:     "ssh-dss mykeys-not-secret",
		Comment:  "me@machine",
		UserName: "gopher",
	}
	key2 := Key{
		Name:     "your-key",
		Body:     "ssh-dss yourkeys-not-secret",
		Comment:  "me@machine",
		UserName: "glenda",
	}
	writeKey(&key1)
	writeKey(&key2)
	expected := key1.format() + key2.format()
	f, err := s.rfs.Open(authKey())
	c.Assert(err, IsNil)
	defer f.Close()
	b, err := ioutil.ReadAll(f)
	c.Assert(err, IsNil)
	got := string(b)
	c.Assert(got, Equals, expected)
}

func (s *S) TestAddKeyStoresKeyInTheDatabase(c *C) {
	err := addKey("key1", rawKey, "gopher")
	c.Assert(err, IsNil)
	var k Key
	err = db.Session.Key().Find(bson.M{"name": "key1"}).One(&k)
	c.Assert(err, IsNil)
	defer db.Session.Key().Remove(bson.M{"name": "key1"})
	c.Assert(k.Name, Equals, "key1")
	c.Assert(k.UserName, Equals, "gopher")
	c.Assert(k.Comment, Equals, comment)
	c.Assert(k.Body, Equals, body)
}

func (s *S) TestAddKeyShouldSaveTheKeyInTheAuthorizedKeys(c *C) {
	err := addKey("key1", rawKey, "gopher")
	c.Assert(err, IsNil)
	defer db.Session.Key().Remove(bson.M{"name": "key1"})
	var k Key
	err = db.Session.Key().Find(bson.M{"name": "key1"}).One(&k)
	c.Assert(err, IsNil)
	f, err := s.rfs.Open(authKey())
	c.Assert(err, IsNil)
	defer f.Close()
	b, err := ioutil.ReadAll(f)
	c.Assert(err, IsNil)
	c.Assert(string(b), Equals, k.format())
}

func (s *S) TestAddKeyDuplicate(c *C) {
	err := addKey("key1", rawKey, "gopher")
	c.Assert(err, IsNil)
	defer db.Session.Key().Remove(bson.M{"name": "key1"})
	err = addKey("key2", rawKey, "gopher")
	c.Assert(err, Equals, ErrDuplicateKey)
}

func (s *S) TestAddKeyInvalidKey(c *C) {
	err := addKey("key1", "something-invalid", "gopher")
	c.Assert(err, Equals, ErrInvalidKey)
}

func (s *S) TestRemoveKeyDeletesFromDB(c *C) {
	err := addKey("key1", rawKey, "gopher")
	c.Assert(err, IsNil)
	err = removeKey("key1", "gopher")
	c.Assert(err, IsNil)
	count, err := db.Session.Key().Find(bson.M{"name": "key1"}).Count()
	c.Assert(err, IsNil)
	c.Assert(count, Equals, 0)
}

func (s *S) TestRemoveKeyDeletesOnlyTheRightKey(c *C) {
	err := addKey("key1", rawKey, "gopher")
	c.Assert(err, IsNil)
	defer removeKey("key1", "gopher")
	err = addKey("key1", otherKey, "glenda")
	c.Assert(err, IsNil)
	err = removeKey("key1", "glenda")
	c.Assert(err, IsNil)
	count, err := db.Session.Key().Find(bson.M{"name": "key1", "username": "gopher"}).Count()
	c.Assert(err, IsNil)
	c.Assert(count, Equals, 1)
}

func (s *S) TestRemoveUnknownKey(c *C) {
	err := removeKey("wut", "glenda")
	c.Assert(err, Equals, ErrKeyNotFound)
}

func (s *S) TestRemoveKeyRemovesFromAuthorizedKeys(c *C) {
	err := addKey("key1", rawKey, "gopher")
	c.Assert(err, IsNil)
	err = removeKey("key1", "gopher")
	f, err := s.rfs.Open(authKey())
	c.Assert(err, IsNil)
	defer f.Close()
	b, err := ioutil.ReadAll(f)
	c.Assert(err, IsNil)
	got := string(b)
	c.Assert(got, Equals, "")
}

func (s *S) TestRemoveKeyKeepOtherKeys(c *C) {
	err := addKey("key1", rawKey, "gopher")
	c.Assert(err, IsNil)
	defer removeKey("key1", "gopher")
	err = addKey("key2", otherKey, "gopher")
	c.Assert(err, IsNil)
	err = removeKey("key2", "gopher")
	c.Assert(err, IsNil)
	var key Key
	err = db.Session.Key().Find(bson.M{"name": "key1"}).One(&key)
	c.Assert(err, IsNil)
	f, err := s.rfs.Open(authKey())
	c.Assert(err, IsNil)
	defer f.Close()
	b, err := ioutil.ReadAll(f)
	c.Assert(err, IsNil)
	got := string(b)
	c.Assert(got, Equals, key.format())
}

func (s *S) TestRemoveUserKeys(c *C) {
	err := addKey("key1", rawKey, "gopher")
	c.Assert(err, IsNil)
	defer removeKey("key1", "gopher")
	err = addKey("key1", otherKey, "glenda")
	c.Assert(err, IsNil)
	err = removeUserKeys("glenda")
	c.Assert(err, IsNil)
	var key Key
	err = db.Session.Key().Find(bson.M{"name": "key1"}).One(&key)
	c.Assert(err, IsNil)
	f, err := s.rfs.Open(authKey())
	c.Assert(err, IsNil)
	defer f.Close()
	b, err := ioutil.ReadAll(f)
	c.Assert(err, IsNil)
	got := string(b)
	c.Assert(got, Equals, key.format())
}

func (s *S) TestRemoveUserMultipleKeys(c *C) {
	err := addKey("key1", rawKey, "glenda")
	c.Assert(err, IsNil)
	err = addKey("key1", otherKey, "glenda")
	c.Assert(err, IsNil)
	err = removeUserKeys("glenda")
	c.Assert(err, IsNil)
	count, err := db.Session.Key().Find(nil).Count()
	c.Assert(err, IsNil)
	c.Assert(count, Equals, 0)
	f, err := s.rfs.Open(authKey())
	c.Assert(err, IsNil)
	defer f.Close()
	b, err := ioutil.ReadAll(f)
	c.Assert(err, IsNil)
	got := string(b)
	c.Assert(got, Equals, "")
}

func (s *S) TestKeyListJSON(c *C) {
	keys := []Key{
		{Name: "key1", Body: "ssh-dss not-secret", Comment: "me@host1"},
		{Name: "key2", Body: "ssh-dss not-secret1", Comment: "me@host2"},
		{Name: "another-key", Body: "ssh-rsa not-secret", Comment: "me@work"},
	}
	expected := map[string]string{
		keys[0].Name: keys[0].String(),
		keys[1].Name: keys[1].String(),
		keys[2].Name: keys[2].String(),
	}
	var got map[string]string
	b, err := KeyList(keys).MarshalJSON()
	c.Assert(err, IsNil)
	err = json.Unmarshal(b, &got)
	c.Assert(err, IsNil)
	c.Assert(got, DeepEquals, expected)
}

func (s *S) TestListKeys(c *C) {
	user := map[string]string{"_id": "glenda"}
	err := db.Session.User().Insert(user)
	c.Assert(err, IsNil)
	defer db.Session.User().Remove(user)
	err = addKey("key1", rawKey, "glenda")
	c.Assert(err, IsNil)
	err = addKey("key2", otherKey, "glenda")
	c.Assert(err, IsNil)
	defer removeUserKeys("glenda")
	var expected []Key
	err = db.Session.Key().Find(nil).All(&expected)
	c.Assert(err, IsNil)
	got, err := ListKeys("glenda")
	c.Assert(err, IsNil)
	c.Assert(got, DeepEquals, KeyList(expected))
}

func (s *S) TestListKeysUnknownUser(c *C) {
	got, err := ListKeys("glenda")
	c.Assert(got, IsNil)
	c.Assert(err, Equals, ErrUserNotFound)
}

func (s *S) TestListKeysEmpty(c *C) {
	user := map[string]string{"_id": "gopher"}
	err := db.Session.User().Insert(user)
	c.Assert(err, IsNil)
	defer db.Session.User().Remove(user)
	got, err := ListKeys("gopher")
	c.Assert(err, IsNil)
	c.Assert(got, HasLen, 0)
}

func (s *S) TestListKeysFromTheUserOnly(c *C) {
	user := map[string]string{"_id": "gopher"}
	err := db.Session.User().Insert(user)
	c.Assert(err, IsNil)
	defer db.Session.User().Remove(user)
	user2 := map[string]string{"_id": "glenda"}
	err = db.Session.User().Insert(user2)
	c.Assert(err, IsNil)
	defer db.Session.User().Remove(user2)
	err = addKey("key1", rawKey, "glenda")
	c.Assert(err, IsNil)
	err = addKey("key1", otherKey, "gopher")
	c.Assert(err, IsNil)
	defer removeUserKeys("glenda")
	defer removeUserKeys("gopher")
	var expected []Key
	err = db.Session.Key().Find(bson.M{"username": "gopher"}).All(&expected)
	c.Assert(err, IsNil)
	got, err := ListKeys("gopher")
	c.Assert(err, IsNil)
	c.Assert(got, DeepEquals, KeyList(expected))
}
