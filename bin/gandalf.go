package main

import (
	"fmt"
	"github.com/globocom/gandalf/db"
	"github.com/globocom/gandalf/repository"
	"github.com/globocom/gandalf/user"
	"labix.org/v2/mgo/bson"
	"os"
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

func requestedRepository() (repository.Repository, error) {
    strings.Split(os.Getenv("SSH_ORIGINAL_COMMAND"), " ")[1]
}

func main() {
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
