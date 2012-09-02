package gandalf

import (
    "labix.org/v2/mgo"
	"testing"
)

func TestSessionRepositoryShouldReturnAMongoCollection(t *testing.T) {
    var rep interface{}
    rep = Session.Repository()
    _, ok := rep.(*mgo.Collection)
    if !ok {
        t.Errorf("Expected rep to be a collection")
    }
}
