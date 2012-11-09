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
