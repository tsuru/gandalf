package repository

import (
	"github.com/timeredbull/gandalf/db"
	"labix.org/v2/mgo/bson"
	"testing"
)

func TestNewShouldCreateANew(t *testing.T) {
	r, err := New("myRepo", []string{"smeagol", "saruman"}, true)
	defer db.Session.Repository().Remove(bson.M{"_id": "myRepo"})
	if err != nil {
		t.Errorf("Unexpected error while creating new repository: %s", err.Error())
	}
	if r.Name != "myRepo" {
		t.Errorf(`Expected repository name to be "myRepo", got "%s"`, r.Name)
	}
	if r.Users == nil {
		t.Errorf(`Expected repository users to be {"smeagol", "saruman"}, got "%s"`, r.Users)
	}
	if !r.IsPublic {
		t.Errorf(`Expected repository to be public`)
	}
}

func TestNewShouldRecordItOnDatabase(t *testing.T) {
	r, err := New("someRepo", []string{"smeagol"}, true)
	defer db.Session.Repository().Remove(bson.M{"_id": "someRepo"})
	if err != nil {
		t.Errorf("Unexpected error while creating new repository: %s", err.Error())
	}
	err = db.Session.Repository().Find(bson.M{"_id": "someRepo"}).One(&r)
	if err != nil {
		t.Errorf("Unexpected error while retrieving new repository: %s", err.Error())
	}
	if r.Name != "someRepo" || r.Users == nil || !r.IsPublic {
		t.Errorf("Retrieved parameters from database doens't match")
	}
}

func TestNewBreaksOnValidationError(t *testing.T) {
	_, err := New("", []string{"smeagol"}, false)
	if err == nil {
		t.Errorf("Expecting an error, got nil")
	}
	expected := "Validation Error: repository needs a valid name"
	got := err.Error()
	if got != expected {
		t.Errorf(`Expected error to be "%s", got "%s"`, expected, got)
	}
}

func TestRepositoryIsNotValidWithoutAName(t *testing.T) {
	r := Repository{Users: []string{"gollum"}, IsPublic: true}
	if r.isValid() {
		t.Errorf("Expecting repository not to be valid")
	}
}

func TestRepositoryIsNotValidWithInvalidName(t *testing.T) {
	r := Repository{Name: "foo bar", Users: []string{"gollum"}, IsPublic: true}
	if r.isValid() {
		t.Errorf("Expecting repository not to be valid")
	}
}

func TestRepositoryShoudBeInvalidWIthoutAnyUsers(t *testing.T) {
	r := Repository{Name: "foo_bar", Users: []string{}, IsPublic: true}
	if r.isValid() {
		t.Errorf("Expecting repository not to be valid")
	}
}

func TestRepositoryShouldBeValidWithoutIsPublic(t *testing.T) {
	r := Repository{Name: "someName", Users: []string{"smeagol"}}
	if !r.isValid() {
		t.Errorf("Expecting repository to be valid")
	}
}
