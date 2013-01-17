// Copyright 2013 gandalf authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package db

import (
	"github.com/globocom/config"
	"labix.org/v2/mgo"
)

type session struct {
	DB *mgo.Database
}

var Session = session{}

// Connect uses database:url and database:name settings in config file and
// connects to the database. If it cannot connect or these settings are not
// defined, it will panic.
func Connect() {
	url, err := config.GetString("database:url")
	if err != nil {
		panic(err)
	}
	name, err := config.GetString("database:name")
	if err != nil {
		panic(err)
	}
	s, err := mgo.Dial(url)
	if err != nil {
		panic(err)
	}
	Session.DB = s.DB(name)
}

func (s *session) Repository() *mgo.Collection {
	return s.DB.C("repository")
}

func (s *session) User() *mgo.Collection {
	return s.DB.C("user")
}
