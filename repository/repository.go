package repository

import (
	"errors"
	"fmt"
	"github.com/timeredbull/config"
	"github.com/timeredbull/gandalf/db"
	"github.com/timeredbull/gandalf/fs"
	"labix.org/v2/mgo/bson"
	"regexp"
)

func init() {
	err := config.ReadConfigFile("/etc/gandalf.conf")
	if err != nil {
		msg := `Could not find gandalf config file. Searched on /etc/gandalf.conf.
For an example conf check gandalf/etc/gandalf.conf file.`
		panic(msg)
	}
}

var fsystem fs.Fs

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
	v, err := r.isValid()
	if !v {
		return r, err
	}
	err = newBare(name)
	if err != nil {
		return r, err
	}
	err = db.Session.Repository().Insert(&r)
	if err != nil {
		return r, err
	}
	return r, nil
}

// Removes a repository representation
// Deletes the repository from the database and
// removes it's bare repository
func Remove(r *Repository) error {
	err := removeBare(r.Name)
	if err != nil {
		return err
	}
	err = db.Session.Repository().Remove(bson.M{"_id": r.Name})
	if err != nil {
		return fmt.Errorf("Could not remove repository: %s", err.Error())
	}
	return nil
}

// Validates a repository
// A valid repository must have:
//  - a name without any special chars only alphanumeric and underlines are allowed.
//  - at least one user in users array
func (r *Repository) isValid() (v bool, err error) {
	v = true
	m, e := regexp.Match(`(^$)|\W+|\s+`, []byte(r.Name))
	if e != nil {
		panic(e)
	}
	if m {
		v = false
		err = errors.New("Validation Error: repository name is not valid")
		return
	}
	if len(r.Users) == 0 {
		v = false
		err = errors.New("Validation Error: repository should have at least one user")
	}
	return
}

func filesystem() fs.Fs {
	if fsystem == nil {
		return fs.OsFs{}
	}
	return fsystem
}
