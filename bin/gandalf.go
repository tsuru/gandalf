package main

import (
	"github.com/globocom/gandalf/repository"
	"github.com/globocom/gandalf/user"
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
