package user

import (
	"errors"
	"github.com/timeredbull/gandalf/db"
	"regexp"
)

type User struct {
	Name string `bson:"_id"`
	Keys []string
}

func New(name string, keys []string) (u *User, err error) {
	u = &User{Name: name, Keys: keys}
	v, err := u.isValid()
	if !v {
		return
	}
	err = db.Session.User().Insert(&u)
	return
}

func (u *User) isValid() (v bool, err error) {
	v = true
	m, e := regexp.Match(`\s|[^aA-zZ0-9\.@]|(^$)`, []byte(u.Name))
	if e != nil {
		panic(e)
	}
	if m {
		v = false
		err = errors.New("Validation Error: user name is not valid")
	}
	return
}

func Remove(u *User) error {
	return db.Session.User().RemoveId(u.Name)
}
