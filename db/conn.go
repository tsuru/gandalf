// Copyright 2013 gandalf authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package db

import "labix.org/v2/mgo"

type session struct {
	DB *mgo.Database
}

var Session = session{}

func init() {
	s, err := mgo.Dial("localhost:27017")
	if err != nil {
		panic(err)
	}
	Session.DB = s.DB("gandalf")
}

func (s *session) Repository() *mgo.Collection {
	return s.DB.C("repository")
}

func (s *session) User() *mgo.Collection {
	return s.DB.C("user")
}
