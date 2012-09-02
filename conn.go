package gandalf

import "labix.org/v2/mgo"

type s struct {
    conn *mgo.Session
}
var Session = s{}
var session *mgo.Session

func init() {
	var err error
	session, err = mgo.Dial("localhost:27017")
	if err != nil {
		panic(err)
	}
    Session.conn, err = mgo.Dial("localhost:27017")
	if err != nil {
		panic(err)
	}
}

func (ssn *s) Repository() *mgo.Collection {
    return ssn.conn.DB("gandalf").C("repository")
}

func (ssn *s) User() *mgo.Collection {
    return ssn.conn.DB("gandalf").C("user")
}
