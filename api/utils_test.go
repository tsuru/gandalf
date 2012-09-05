package api

import (
	"github.com/timeredbull/gandalf/db"
	"github.com/timeredbull/gandalf/user"
	"labix.org/v2/mgo/bson"
	"testing"
)

func TestGetUserOr404(t *testing.T) {
	u := user.User{Name: "umi"}
	err := db.Session.User().Insert(&u)
	if err != nil {
		t.Errorf("Got error while creating user: %s", err.Error())
	}
	defer db.Session.User().Remove(bson.M{"_id": u.Name})
	rUser, err := getUserOr404("umi")
	if err != nil {
		t.Errorf("Got error while creating user: %s", err.Error())
	}
	if rUser.Name != "umi" {
		t.Errorf(`Expected retieved user's name to be umi, got: %s`, rUser.Name)
	}
}

func TestGetUserOr404ShouldReturn404WhenUserDoesntExists(t *testing.T) {
	_, e := getUserOr404("umi")
	expected := "User umi not found"
	got := e.Error()
	if got != expected {
		t.Errorf(`Expected error to be "%s", got: %s`, expected, got)
	}
}
