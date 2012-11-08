package repository

import (
	"errors"
	"fmt"
	"github.com/globocom/config"
	"github.com/globocom/gandalf/db"
	"github.com/globocom/tsuru/fs"
	"labix.org/v2/mgo/bson"
	"regexp"
)

type Repository struct {
	Name     string `bson:"_id"`
	Users    []string
	IsPublic bool
}

// Creates a representation of a git repository
// This function creates a git repository using the "bare-dir" config
// and saves repository's meta data in the database
func New(name string, users []string, isPublic bool) (*Repository, error) {
	r := &Repository{Name: name, Users: users, IsPublic: isPublic}
	if v, err := r.isValid(); !v {
		return r, err
	}
	if err := newBare(name); err != nil {
		return r, err
	}
	return r, db.Session.Repository().Insert(&r)
}

// Deletes the repository from the database and
// removes it's bare git repository
func Remove(r *Repository) error {
	// maybe it should receive only a name, to standardize the api (user.Remove already does that)
	if err := removeBare(r.Name); err != nil {
		return err
	}
	if err := db.Session.Repository().Remove(bson.M{"_id": r.Name}); err != nil {
		return fmt.Errorf("Could not remove repository: %s", err)
	}
	return nil
}

// Format the git remote url and return it
// If no remote is configured in gandalf.conf Remote will panic
func (r *Repository) Remote() string {
	host, err := config.GetString("host")
	if err != nil {
		panic(err.Error())
	}
	return fmt.Sprintf("git@%s:%s", host, formatName(r.Name))
}

// Validates a repository
// A valid repository must have:
//  - a name without any special chars only alphanumeric and underlines are allowed.
//  - at least one user in users array
func (r *Repository) isValid() (bool, error) {
	m, e := regexp.Match(`(^$)|\W+|\s+`, []byte(r.Name))
	if e != nil {
		panic(e)
	}
	if m {
		return false, errors.New("Validation Error: repository name is not valid")
	}
	if len(r.Users) == 0 {
		return false, errors.New("Validation Error: repository should have at least one user")
	}
	return true, nil
}

var fsystem fs.Fs

func filesystem() fs.Fs {
	if fsystem == nil {
		return fs.OsFs{}
	}
	return fsystem
}
