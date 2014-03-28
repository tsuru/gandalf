// Copyright 2014 gandalf authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/globocom/tsuru/log"
	"github.com/tsuru/config"
	"github.com/tsuru/gandalf/db"
	"github.com/tsuru/gandalf/fs"
	"io/ioutil"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"regexp"
)

// Repository represents a Git repository. A Git repository is a record in the
// database and a directory in the filesystem (the bare repository).
type Repository struct {
	Name     string `bson:"_id"`
	Users    []string
	IsPublic bool
}

// MarshalJSON marshals the Repository in json format.
func (r *Repository) MarshalJSON() ([]byte, error) {
	data := map[string]interface{}{
		"name":    r.Name,
		"public":  r.IsPublic,
		"ssh_url": r.ReadWriteURL(),
		"git_url": r.ReadOnlyURL(),
	}
	return json.Marshal(&data)
}

// New creates a representation of a git repository. It creates a Git
// repository using the "bare-dir" setting and saves repository's meta data in
// the database.
func New(name string, users []string, isPublic bool) (*Repository, error) {
	log.Debugf(`Creating repository "%s"`, name)
	r := &Repository{Name: name, Users: users, IsPublic: isPublic}
	if v, err := r.isValid(); !v {
		log.Errorf(`repository.New: Invalid repository "%s": %s`, name, err.Error())
		return r, err
	}
	if err := newBare(name); err != nil {
		log.Errorf(`repository.New: Error creating bare repository for "%s": %s`, name, err.Error())
		return r, err
	}
	barePath := barePath(name)
	if barePath != "" && isPublic {
		ioutil.WriteFile(barePath+"/git-daemon-export-ok", []byte(""), 0644)
		if f, err := fs.Filesystem().Create(barePath + "/git-daemon-export-ok"); err == nil {
			f.Close()
		}
	}
	conn, err := db.Conn()
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	err = conn.Repository().Insert(&r)
	if mgo.IsDup(err) {
		log.Errorf(`repository.New: Duplicate repository "%s"`, name)
		return r, fmt.Errorf("A repository with this name already exists.")
	}
	return r, err
}

// Get find a repository by name.
func Get(name string) (Repository, error) {
	var r Repository
	conn, err := db.Conn()
	if err != nil {
		return r, err
	}
	defer conn.Close()
	err = conn.Repository().FindId(name).One(&r)
	return r, err
}

// Remove deletes the repository from the database and removes it's bare Git
// repository.
func Remove(name string) error {
	log.Debugf(`Removing repository "%s"`, name)
	if err := removeBare(name); err != nil {
		log.Errorf(`repository.Remove: Error removing bare repository "%s": %s`, name, err.Error())
		return err
	}
	conn, err := db.Conn()
	if err != nil {
		return err
	}
	defer conn.Close()
	if err := conn.Repository().RemoveId(name); err != nil {
		log.Errorf(`repository.Remove: Error removing repository "%s" from db: %s`, name, err.Error())
		return fmt.Errorf("Could not remove repository: %s", err)
	}
	return nil
}

// Rename renames a repository.
func Rename(oldName, newName string) error {
	log.Debugf(`Renaming repository "%s" to "%s"`, oldName, newName)
	repo, err := Get(oldName)
	if err != nil {
		log.Errorf(`repository.Rename: Repository "%s" not found: %s`, oldName, err)
		return err
	}
	newRepo := repo
	newRepo.Name = newName
	conn, err := db.Conn()
	if err != nil {
		return err
	}
	defer conn.Close()
	err = conn.Repository().Insert(newRepo)
	if err != nil {
		log.Errorf("repository.Rename: Error adding new repository %q: %s", newName, err)
		return err
	}
	err = conn.Repository().RemoveId(oldName)
	if err != nil {
		log.Errorf(`repository.Rename: Error removing old repository "%s": %s`, oldName, err)
		return err
	}
	return fs.Filesystem().Rename(barePath(oldName), barePath(newName))
}

// ReadWriteURL formats the git ssh url and return it. If no remote is configured in
// gandalf.conf, this method panics.
func (r *Repository) ReadWriteURL() string {
	uid, err := config.GetString("uid")
	if err != nil {
		panic(err.Error())
	}
	remote := uid + "@%s:%s.git"
	if useSSH, _ := config.GetBool("git:ssh:use"); useSSH {
		port, err := config.GetString("git:ssh:port")
		if err == nil {
			remote = "ssh://" + uid + "@%s:" + port + "/%s.git"
		} else {
			remote = "ssh://" + uid + "@%s/%s.git"
		}
	}
	host, err := config.GetString("host")
	if err != nil {
		panic(err.Error())
	}
	return fmt.Sprintf(remote, host, r.Name)
}

// ReadOnly formats the git url and return it. If no host is configured in
// gandalf.conf, this method panics.
func (r *Repository) ReadOnlyURL() string {
	remote := "git://%s/%s.git"
	if useSSH, _ := config.GetBool("git:ssh:use"); useSSH {
		uid, err := config.GetString("uid")
		if err != nil {
			panic(err.Error())
		}
		port, err := config.GetString("git:ssh:port")
		if err == nil {
			remote = "ssh://" + uid + "@%s:" + port + "/%s.git"
		} else {
			remote = "ssh://" + uid + "@%s/%s.git"
		}
	}
	host, err := config.GetString("host")
	if err != nil {
		panic(err.Error())
	}
	return fmt.Sprintf(remote, host, r.Name)
}

// Validates a repository
// A valid repository must have:
//  - a name without any special chars only alphanumeric and underlines are allowed.
//  - at least one user in users array
func (r *Repository) isValid() (bool, error) {
	m, e := regexp.Match(`^[\w-]+$`, []byte(r.Name))
	if e != nil {
		panic(e)
	}
	if !m {
		return false, errors.New("Validation Error: repository name is not valid")
	}
	if len(r.Users) == 0 {
		return false, errors.New("Validation Error: repository should have at least one user")
	}
	return true, nil
}

// GrantAccess gives write permission for users in all specified repositories.
// If any of the repositories/users do not exists, GrantAccess just skips it.
func GrantAccess(rNames, uNames []string) error {
	conn, err := db.Conn()
	if err != nil {
		return err
	}
	defer conn.Close()
	_, err = conn.Repository().UpdateAll(bson.M{"_id": bson.M{"$in": rNames}}, bson.M{"$addToSet": bson.M{"users": bson.M{"$each": uNames}}})
	return err
}

// RevokeAccess revokes write permission from users in all specified
// repositories.
func RevokeAccess(rNames, uNames []string) error {
	conn, err := db.Conn()
	if err != nil {
		return err
	}
	defer conn.Close()
	_, err = conn.Repository().UpdateAll(bson.M{"_id": bson.M{"$in": rNames}}, bson.M{"$pullAll": bson.M{"users": uNames}})
	return err
}
