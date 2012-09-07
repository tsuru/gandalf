package user

import (
	"github.com/timeredbull/gandalf/db"
	"labix.org/v2/mgo/bson"
	"testing"
)

func TestNewUserReturnsAStructFilled(t *testing.T) {
	u, err := New("someuser", []string{"id_rsa someKeyChars"})
	if err != nil {
		t.Errorf(`Got error while creating user: "%s"`, err.Error())
	}
	defer db.Session.User().Remove(bson.M{"_id": u.Name})
	if u.Name != "someuser" {
		t.Errorf(`Expected user name to be "someuser", got "%s"`, u.Name)
	}
	if len(u.Keys) == 0 {
		t.Errorf(`Expected user to have 1 key, got %d`, len(u.Keys))
	}
}

func TestNewUserShouldStoreUserInDatabase(t *testing.T) {
	u, err := New("someuser", []string{"id_rsa someKeyChars"})
	if err != nil {
		t.Errorf(`Got error while creating user: "%s"`, err.Error())
	}
	defer db.Session.User().Remove(bson.M{"_id": u.Name})
	err = db.Session.User().Find(bson.M{"_id": u.Name}).One(&u)
	if err != nil {
		t.Errorf(`Got error while creating user: "%s"`, err.Error())
	}
	if u.Name != "someuser" {
		t.Errorf(`Expected user name to be "someuser", got "%s"`, u.Name)
	}
	if len(u.Keys) == 0 {
		t.Errorf(`Expected user to have 1 key, got %d`, len(u.Keys))
	}
}

func TestNewChecksIfUserIsValidBeforeStoring(t *testing.T) {
	_, err := New("", []string{})
	if err == nil {
		t.Errorf("Expected err not to be nil")
		t.FailNow()
	}
	got := err.Error()
	expected := "Validation Error: user name is not valid"
	if got != expected {
		t.Errorf(`Expected error to be "%s", got "%s"`, expected, got)
	}
}

func TestIsValidReturnsErrorWhenUserDoesNotHaveAName(t *testing.T) {
	u := User{Keys: []string{"id_rsa foooBar"}}
	v, err := u.isValid()
	if v {
		t.Errorf(`Expected user to be invalid`)
	}
	if err == nil {
		t.Errorf(`Expected error not to be nil`)
		t.FailNow()
	}
	expected := "Validation Error: user name is not valid"
	got := err.Error()
	if got != expected {
		t.Errorf(`Expected error to be "%s", got "%s"`, expected, got)
	}
}

func TestIsValidShouldNotAcceptEmptyUserName(t *testing.T) {
	u := User{Keys: []string{"id_rsa foooBar"}}
	v, err := u.isValid()
	if v {
		t.Errorf(`Expected user to be invalid`)
	}
	if err == nil {
		t.Errorf(`Expected error not to be nil`)
		t.FailNow()
	}
	expected := "Validation Error: user name is not valid"
	got := err.Error()
	if got != expected {
		t.Errorf(`Expected error to be "%s", got "%s"`, expected, got)
	}
}

func TestIsValidShouldAcceptEmailsAsUserName(t *testing.T) {
	u := User{Name: "r2d2@gmail.com", Keys: []string{"id_rsa foooBar"}}
	v, _ := u.isValid()
	if !v {
		t.Errorf(`Expected user to be valid`)
	}
}
