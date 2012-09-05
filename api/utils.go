package api

import (
	"fmt"
	"github.com/timeredbull/gandalf/db"
	"github.com/timeredbull/gandalf/user"
	"labix.org/v2/mgo/bson"
)

func getUserOr404(name string) (u user.User, e error) {
	e = db.Session.User().Find(bson.M{"_id": name}).One(&u)
	if e != nil && e.Error() == "not found" {
		return u, fmt.Errorf("User %s not found", name)
	}
	return
}
