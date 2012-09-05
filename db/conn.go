package db

import "labix.org/v2/mgo"

type session struct {
	DB *mgo.Database
}

var Session = session{}

func init() {
	var err error
	var s *mgo.Session
	s, err = mgo.Dial("localhost:27017")
	Session.DB = s.DB("gandalf")
	if err != nil {
		panic(err)
	}
}

func (s *session) Repository() *mgo.Collection {
	return s.DB.C("repository")
}

func (s *session) User() *mgo.Collection {
	return s.DB.C("user")
}
