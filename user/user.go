package user

import (
	"errors"
	"github.com/globocom/gandalf/db"
	"regexp"
)

type User struct {
	Name string `bson:"_id"`
	Keys []string
}

func New(name string, keys []string) (*User, error) {
	u := &User{Name: name, Keys: keys}
	v, err := u.isValid()
	if !v {
		return u, err
	}
	err = db.Session.User().Insert(&u)
	return u, err
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
	return db.Session.User().RemoveId(u.Name)
}
