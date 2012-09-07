package repository

import (
	"errors"
	"github.com/timeredbull/gandalf/db"
	"regexp"
)

type Repository struct {
	Name     string `bson:"_id"`
	Users    []string
	IsPublic bool
}

func New(name string, users []string, isPublic bool) (r *Repository, err error) {
	r = &Repository{Name: name, Users: users, IsPublic: isPublic}
	var v bool
	v, err = r.isValid()
	if !v {
		return
	}
	err = db.Session.Repository().Insert(&r)
	return
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
