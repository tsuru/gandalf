package user

import (
	"errors"
	"fmt"
	"github.com/globocom/gandalf/db"
	"github.com/globocom/gandalf/key"
	"github.com/globocom/tsuru/fs"
	"labix.org/v2/mgo/bson"
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
	return u, key.BulkAdd(keys, name, filesystem())
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

// Remove a user
// Also removes it's associated keys from authorized_keys
// Does not checks for relations with repositories (maybe it should)
func Remove(name string) error {
	var u *User
	err := db.Session.User().Find(bson.M{"_id": name}).One(&u)
	if err != nil {
		return fmt.Errorf("Could not remove user: %s", err)
	}
	err = db.Session.User().RemoveId(u.Name)
	if err != nil {
		return fmt.Errorf("Could not remove user: %s", err)
	}
	return key.BulkRemove(u.Keys, u.Name, filesystem())
}

var fsystem fs.Fs

func filesystem() fs.Fs {
	if fsystem == nil {
		return fs.OsFs{}
	}
	return fsystem
}
