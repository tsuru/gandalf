package user

import (
	"errors"
	"github.com/globocom/gandalf/db"
	"github.com/globocom/gandalf/key"
	"github.com/globocom/tsuru/fs"
	"regexp"
)

type User struct {
	Name string `bson:"_id"`
	Keys []string
}

func New(name string, keys []string) (*User, error) {
	u := &User{Name: name, Keys: keys}
	if v, err := u.isValid(); !v {
		return u, err
	}
	if err := db.Session.User().Insert(&u); err != nil {
		return u, err
	}
	return u, key.BulkAdd(keys, filesystem())
}

func (u *User) isValid() (isValid bool, err error) {
	m, err := regexp.Match(`\s|[^aA-zZ0-9\.@]|(^$)`, []byte(u.Name))
	if err != nil {
		panic(err)
	}
	if m {
		return false, errors.New("Validation Error: user name is not valid")
	}
	return true, nil
}

func Remove(u *User) error {
	err := db.Session.User().RemoveId(u.Name)
	if err != nil {
		return err
	}
	return key.BulkRemove(u.Keys, filesystem())
}

var fsystem fs.Fs

func filesystem() fs.Fs {
	if fsystem == nil {
		return fs.OsFs{}
	}
	return fsystem
}
