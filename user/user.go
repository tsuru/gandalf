package user

import (
	"errors"
	"fmt"
	"github.com/globocom/gandalf/db"
	"github.com/globocom/gandalf/key"
	"github.com/globocom/gandalf/repository"
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
// Also removes it's associated keys from authorized_keys and repositories
// It handles user with repositories specially:
// - if a user has at least one repository:
//     - if he/she is the only one with access to the repository, the removal will stop and return an error
//     - if there are more than one user, gandalf will first revoke user's access to the user and then remove it permanently
// - if a user has no repositories, gandalf will simply remove the user
func Remove(name string) error {
	var u *User
	err := db.Session.User().Find(bson.M{"_id": name}).One(&u)
	if err != nil {
		return fmt.Errorf("Could not remove user: %s", err)
	}
	// find associated repos
	var repos []repository.Repository
	err = db.Session.Repository().Find(bson.M{"users": u.Name}).All(&repos)
	if err != nil {
		return err
	}
	// check repositories association
	for _, r := range repos {
		if len(r.Users) == 1 {
			return errors.New("Could not remove user: user is the only one with access to at least one of it's repositories")
		}
	}
	for _, r := range repos {
		for i, v := range r.Users {
			if v == u.Name {
				r.Users[i], r.Users = r.Users[len(r.Users)-1], r.Users[:len(r.Users)-1]
				err = db.Session.Repository().Update(bson.M{"_id": r.Name}, r)
				if err != nil {
					return err
				}
				break
			}
		}
	}
	err = db.Session.User().RemoveId(u.Name)
	if err != nil {
		return fmt.Errorf("Could not remove user: %s", err.Error())
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
