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
	if !r.isValid() {
		err = errors.New("Validation Error: repository needs a valid name")
		return
	}
	err = db.Session.Repository().Insert(&r)
	return
}

func (r *Repository) isValid() bool {
	m, e := regexp.Match(`(^$)|\W+|\s+`, []byte(r.Name))
	if e != nil {
		panic(e)
	}
	return !m
}
