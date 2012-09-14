package repository

import (
	"github.com/timeredbull/commandmocker"
	"github.com/timeredbull/gandalf/db"
	"github.com/timeredbull/gandalf/fs"
	"labix.org/v2/mgo/bson"
	"testing"
)

func TestNewShouldCreateANewRepository(t *testing.T) {
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
		t.FailNow()
	}
	expected := "Validation Error: repository name is not valid"
	got := err.Error()
	if got != expected {
		t.Errorf(`Expected error to be "%s", got "%s"`, expected, got)
	}
}

func TestRepositoryIsNotValidWithoutAName(t *testing.T) {
	r := Repository{Users: []string{"gollum"}, IsPublic: true}
	v, err := r.isValid()
	if v {
		t.Errorf("Expecting repository not to be valid")
	}
	if err == nil {
		t.Errorf("Expecting error not to be nil")
		t.FailNow()
	}
	got := err.Error()
	expected := "Validation Error: repository name is not valid"
	if got != expected {
		t.Errorf(`Expecting error to be "%s", got "%s"`, expected, got)
	}
}

func TestRepositoryIsNotValidWithInvalidName(t *testing.T) {
	r := Repository{Name: "foo bar", Users: []string{"gollum"}, IsPublic: true}
	v, err := r.isValid()
	if v {
		t.Errorf("Expecting repository not to be valid")
	}
	if err == nil {
		t.Errorf("Expecting error not to be nil")
		t.FailNow()
	}
	got := err.Error()
	expected := "Validation Error: repository name is not valid"
	if got != expected {
		t.Errorf(`Expecting error to be "%s", got "%s"`, expected, got)
	}
}

func TestRepositoryShoudBeInvalidWIthoutAnyUsers(t *testing.T) {
	r := Repository{Name: "foo_bar", Users: []string{}, IsPublic: true}
	v, err := r.isValid()
	if v {
		t.Errorf("Expecting repository not to be valid")
	}
	if err == nil {
		t.Errorf("Expecting error not to be nil")
		t.FailNow()
	}
	got := err.Error()
	expected := "Validation Error: repository should have at least one user"
	if got != expected {
		t.Errorf(`Expecting error to be "%s", got "%s"`, expected, got)
	}
}

func TestRepositoryShouldBeValidWithoutIsPublic(t *testing.T) {
	r := Repository{Name: "someName", Users: []string{"smeagol"}}
	v, _ := r.isValid()
	if !v {
		t.Errorf("Expecting repository to be valid")
	}
}

func TestNewShouldCreateNewGitBareRepository(t *testing.T) {
    dir, err := commandmocker.Add("git", "$*")
    if err != nil {
        t.Errorf(`Unpexpected error while mocking git cmd: %s`, err.Error())
        t.FailNow()
    }
    defer commandmocker.Remove(dir)
	_, err = New("myRepo", []string{"pumpkin"}, true)
    if err != nil {
        t.Errorf(`Unexpected error while creating repository: %s`, err.Error())
    }
    defer db.Session.Repository().Remove(bson.M{"_id": "myRepo"})
    if !commandmocker.Ran(dir) {
        t.Errorf("Expected New to create git bare repo")
    }
}

func TestFsystemShouldSetGlobalFsystemWhenItsNil(t *testing.T) {
    fsystem = nil
    fsys := filesystem()
    _, ok := fsys.(fs.Fs)
    if !ok {
        t.Errorf("Expected filesystem function to return a fs.Fs")
    }
}
