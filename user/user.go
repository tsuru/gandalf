package user

import (
	"errors"
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
	return u, u.writeKeys(keys)
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

func (u *User) AddKeys(keys []string) error {
	// the key is saved in the database without any the needed formats (like command and no-pty)
	u.Keys = append(u.Keys, keys...)
	err := db.Session.User().Update(bson.M{"_id": u.Name}, u)
	if err != nil {
		return err
	}
	return u.writeKeys(keys)
}

func (u *User) writeKeys(keys []string) error {
	for _, k := range keys {
		err := key.Add(k, filesystem())
		if err != nil {
			return err
		}
	}
	return nil
}

func Remove(u *User) error {
	//extract
	fSystem := filesystem()
	for _, k := range u.Keys {
		err := key.Remove(k, fSystem)
		if err != nil {
			return err
		}
	}
	return db.Session.User().RemoveId(u.Name)
}

var fsystem fs.Fs

func filesystem() fs.Fs {
	if fsystem == nil {
		return fs.OsFs{}
	}
	return fsystem
}
