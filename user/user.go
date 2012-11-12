package user

import (
	"errors"
	"fmt"
	"github.com/globocom/gandalf/db"
	"github.com/globocom/gandalf/key"
	"github.com/globocom/gandalf/repository"
	"labix.org/v2/mgo/bson"
	"regexp"
)

type User struct {
	Name string `bson:"_id"`
	Keys []key.Key
}

func New(name string, keys []key.Key) (*User, error) {
	u := &User{Name: name, Keys: keys}
	if v, err := u.isValid(); !v {
		return u, err
	}
	if err := db.Session.User().Insert(&u); err != nil {
		return u, err
	}
	return u, key.BulkAdd(keys, name)
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
	if err := db.Session.User().Find(bson.M{"_id": name}).One(&u); err != nil {
		return fmt.Errorf("Could not remove user: %s", err)
	}
	if err := u.handleAssociatedRepositories(); err != nil {
		return err
	}
	if err := db.Session.User().RemoveId(u.Name); err != nil {
		return fmt.Errorf("Could not remove user: %s", err.Error())
	}
	return key.BulkRemove(u.Keys, u.Name)
}

func (u *User) handleAssociatedRepositories() error {
	var repos []repository.Repository
	if err := db.Session.Repository().Find(bson.M{"users": u.Name}).All(&repos); err != nil {
		return err
	}
	for _, r := range repos {
		if len(r.Users) == 1 {
			return errors.New("Could not remove user: user is the only one with access to at least one of it's repositories")
		}
	}
	for _, r := range repos {
		for i, v := range r.Users {
			if v == u.Name {
				r.Users[i], r.Users = r.Users[len(r.Users)-1], r.Users[:len(r.Users)-1]
				if err := db.Session.Repository().Update(bson.M{"_id": r.Name}, r); err != nil {
					return err
				}
				break
			}
		}
	}
	return nil
}

func AddKey(uName string, k *key.Key) error {
	var u User
	if err := db.Session.User().FindId(uName).One(&u); err != nil {
		return fmt.Errorf(`User "%s" not found`, uName)
	}
	u.Keys = append(u.Keys, *k)
	if err := db.Session.User().UpdateId(u.Name, u); err != nil {
		return err
	}
	return key.Add(k, u.Name)
}

// RemoveKey removes the key from the user's document and from authorized_keys file
// If the user or the key is not found, returns an error
func RemoveKey(uName, kName string) error {
	var u User
	if err := db.Session.User().FindId(uName).One(&u); err != nil {
		return fmt.Errorf(`User "%s" does not exists`, uName)
	}
	var kContent string
	kNums := len(u.Keys)
	for i, v := range u.Keys {
		if v.Name == kName {
			u.Keys[i], u.Keys = u.Keys[len(u.Keys)-1], u.Keys[:len(u.Keys)-1]
			kContent = v.Content
			break
		}
	}
	if kNums == len(u.Keys) {
		return fmt.Errorf(`Key "%s" for user "%s" does not exists`, kName, uName)
	}
	if err := db.Session.User().UpdateId(uName, u); err != nil {
		return err
	}
	return key.Remove(kContent, uName)
}
