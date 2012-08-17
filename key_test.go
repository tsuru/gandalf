package gandalf

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

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

func TestAuthKeysShouldBeAbsolutePathToUsersAuthorizedKeys(t *testing.T) {
	home := os.Getenv("HOME")
	expected := path.Join(home, "authorized_keys")
	if authKey != expected {
		t.Errorf(`expected authKey to be %s, got: %s`, expected, authKey)
	}
}

func TestShouldAddKeyWithoutError(t *testing.T) {
	changeAuthKey()
	key := "somekey blaaaaaaa r2d2@host"
	err := addKey(key)
	if err != nil {
		t.Errorf(`Expecting err to be nil, got: %s`, err.Error())
	}
	ok := clearAuthKeyFile()
	if !ok {
		t.Errorf("Could not truncate file... :/")
		t.FailNow()
	}
}

func TestShouldWriteKeyInFile(t *testing.T) {
	changeAuthKey()
	key := "somekey blaaaaaaa r2d2@host"
	err := addKey(key)
	if err != nil {
		t.Errorf(`Expecting err to be nil, got: %s`, err.Error())
		t.FailNow()
	}
	b, err := ioutil.ReadFile(authKey)
	if err != nil {
		t.Errorf(`Expecting err to be nil, got: %s`, err.Error())
		t.FailNow()
	}
	got := string(b)
	if got != key {
		t.Errorf(`Expecing authorized_keys to be "%s", got "%s"`, key, got)
	}
	ok := clearAuthKeyFile()
	if !ok {
		t.Errorf("Could not truncate file... :/")
		t.FailNow()
	}
}

func TestShouldAppendKeyInFile(t *testing.T) {
	changeAuthKey()
	key1 := "somekey blaaaaaaa r2d2@host"
	err := addKey(key1)
	if err != nil {
		t.Errorf(`Expecting err to be nil, got: %s`, err.Error())
		t.FailNow()
	}
	key2 := "someotherkey fooo r2d2@host"
	err = addKey(key2)
	if err != nil {
		t.Errorf(`Expecting err to be nil, got: %s`, err.Error())
		t.FailNow()
	}
	b, err := ioutil.ReadFile(authKey)
	if err != nil {
		t.Errorf(`Expecting err to be nil, got: %s`, err.Error())
		t.FailNow()
	}
	got := string(b)
	expected := fmt.Sprintf("%s\n%s", key1, key2)
	if got != expected {
		t.Errorf(`Expecing authorized_keys to be "%s", got "%s"`, expected, got)
	}
	ok := clearAuthKeyFile()
	if !ok {
		t.Errorf("Could not truncate file... :/")
		t.FailNow()
	}
}
