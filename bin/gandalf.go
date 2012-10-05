package main

import (
	"errors"
	"fmt"
	"github.com/globocom/gandalf/db"
	"github.com/globocom/gandalf/repository"
	"github.com/globocom/gandalf/user"
	"labix.org/v2/mgo/bson"
	"os"
	"regexp"
	"strings"
)

// TODO: receive argument with user name

func hasWritePermission(u *user.User, r *repository.Repository) (allowed bool) {
	for _, userName := range r.Users {
		if u.Name == userName {
			return true
		}
	}
	return false
}

func hasReadPermission(u *user.User, r *repository.Repository) (allowed bool) {
	if r.IsPublic {
		return true
	}
	for _, userName := range r.Users {
		if u.Name == userName {
			return true
		}
	}
	return false
}

// Returns the command being executed by ssh.
// When a user runs `$ git push` from his/her machine, the server
// receives a ssh command, identified by this user (by the ssh key).
// The command and it's parameters are available through the SSH_ORIGINAL_COMMAND
// environment variable. In the git push example, it would have the following value:
// SSH_ORIGINAL_COMMAND=git-receive-pack 'foo.git'
// This function is responsible for retrieving the `git-receive-pack` part of SSH_ORIGINAL_COMMAND
func action() string {
	return strings.Split(os.Getenv("SSH_ORIGINAL_COMMAND"), " ")[0]
}

// Get the repository name requested in SSH_ORIGINAL_COMMAND and gets
// the related document on the database and returns it.
// this function does two distinct things (maybe it should'n), it
// parses the SSH_ORIGINAL_COMMAND and returns a "validation" error if it doesn't
// matches the expected format and gets the repository from the database based on the info
// obtained by the SSH_ORIGINAL_COMMAND parse.
func requestedRepository() (repository.Repository, error) {
	r, err := regexp.Compile(`[\w-]+ '([\w-]+)\.git'`)
	if err != nil {
		panic(err)
	}
	m := r.FindStringSubmatch(os.Getenv("SSH_ORIGINAL_COMMAND"))
	if len(m) < 2 {
		return repository.Repository{}, errors.New("Cannot deduce repository name from command. You are probably trying to do something you shouldn't")
	}
	repoName := m[1]
	var repo repository.Repository
	if err = db.Session.Repository().Find(bson.M{"_id": repoName}).One(&repo); err != nil {
		return repository.Repository{}, err
	}
	return repo, nil
}

func main() {
	// (flaviamissi): should we call a validate function before anything? (I think so)
	a := action()
	var u user.User
	err := db.Session.User().Find(bson.M{"_id": os.Args[1]}).One(&u)
	if err != nil {
		fmt.Println("Error obtaining user. Gandalf database is probably in an inconsistent state.")
	}
	// user is trying to write into repository
	// see man git-receive-pack
	if a == "git-receive-pack" {
	}
}
