package main

import (
	"errors"
	"fmt"
	"github.com/globocom/gandalf/db"
	"github.com/globocom/gandalf/repository"
	"github.com/globocom/gandalf/user"
	"io"
	"labix.org/v2/mgo/bson"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

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
		return repository.Repository{}, errors.New("Repository not found")
	}
	return repo, nil
}

// Checks whether a command is a valid git command
// The following format is allowed:
//   git-([\w-]+) '([\w-]+)\.git'
func validateCmd() error {
	r, err := regexp.Compile(`git-([\w-]+) '([\w-]+)\.git'`)
	if err != nil {
		panic(err)
	}
	if m := r.FindStringSubmatch(os.Getenv("SSH_ORIGINAL_COMMAND")); len(m) < 3 {
		return errors.New("You've tried to execute some weird command, I'm deliberately denying you to execute that, get over it.")
	}
	return nil
}

// Executes the SSH_ORIGINAL_COMMAND based on the condition
// defined by the `f` parameter.
// Also receives a custom error message to print to the end user and a
// stdout object, where the SSH_ORIGINAL_COMMAND output is going to be written
func executeAction(f func(*user.User, *repository.Repository) bool, errMsg string, stdout io.Writer) {
	var u user.User
	if err := db.Session.User().Find(bson.M{"_id": os.Args[1]}).One(&u); err != nil {
		fmt.Println("Error obtaining user. Gandalf database is probably in an inconsistent state.")
		return
	}
	repo, err := requestedRepository()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	if f(&u, &repo) {
		cmdStr := strings.Split(os.Getenv("SSH_ORIGINAL_COMMAND"), " ")
		cmd := exec.Command(cmdStr[0], cmdStr[1:]...)
		cmd.Stdout = stdout
		err = cmd.Run()
		if err != nil {
			fmt.Println("Got error while executing command:")
			fmt.Println(err.Error())
		}
		return
	}
	fmt.Println("Permission denied.")
	fmt.Println(errMsg)
}

func main() {
	err := validateCmd()
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	a := action()
	if a == "git-receive-pack" {
		executeAction(hasWritePermission, "You don't have access to write in this repository.", os.Stdout)
		return
	}
	if a == "git-upload-pack" {
		executeAction(hasReadPermission, "You don't have access to read this repository.", os.Stdout)
		return
	}
}
