package gandalf

import "labix.org/v2/mgo"

type session struct {
	conn *mgo.Session
}

var Session = session{}

func init() {
	var err error
	Session.conn, err = mgo.Dial("localhost:27017")
	if err != nil {
		panic(err)
	}
}

func (ssn *session) Repository() *mgo.Collection {
	return ssn.conn.DB("gandalf").C("repository")
}

func (ssn *session) User() *mgo.Collection {
	return ssn.conn.DB("gandalf").C("user")
}
