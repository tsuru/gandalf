package gandalf

import "labix.org/v2/mgo"

var session *mgo.Session

func init() {
	var err error
	session, err = mgo.Dial("localhost:27017")
	if err != nil {
		panic(err)
	}
}
