package repository

import (
	"errors"
	"github.com/timeredbull/gandalf/db"
	"github.com/timeredbull/gandalf/fs"
	"regexp"
)

var fsystem fs.Fs

type Repository struct {
	Name     string `bson:"_id"`
	Users    []string
	IsPublic bool
}

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
