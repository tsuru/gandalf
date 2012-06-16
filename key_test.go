package gandalf

import (
	"os"
	"testing"
)

func changeAuthKeys() {
	authKeys = "testdata/authorized_keys"
}

func TestShouldAddKeyWithoutError(t *testing.T) {
	changeAuthKeys()
	key := "somekey blaaaaaaa r2d2@host"
	err := addKey(key)
	if err != nil {
		t.Errorf(`Expecting err to be nil, got: %s`, err.Error())
	}
}

func TestShouldWriteKeyInFile(t *testing.T) {
	changeAuthKeys()
	key := "somekey blaaaaaaa r2d2@host"
	err := addKey(key)
	if err != nil {
		t.Errorf(`Expecting err to be nil, got: %s`, err.Error())
		t.FailNow()
	}
	file, err := os.Open(authKeys)
	if err != nil {
		t.Errorf(`Expecting err to be nil, got: %s`, err.Error())
		t.FailNow()
	}
	b := make([]byte, len(key))
	_, err = file.Read(b)
	if err != nil {
		t.Errorf(`Expecting err to be nil, got: %s`, err.Error())
		t.FailNow()
	}
	got := string(b)
	if got != key {
		t.Errorf(`Expecing authorized_keys to be "%s", got "%s"`, key, got)
	}
}
