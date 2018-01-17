// Copyright 2014 gandalf authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

import (
	"fmt"

	"github.com/globalsign/mgo/bson"
	"github.com/tsuru/gandalf/db"
	"github.com/tsuru/gandalf/user"
)

func getUserOr404(name string) (user.User, error) {
	var u user.User
	conn, err := db.Conn()
	if err != nil {
		return u, err
	}
	defer conn.Close()
	if err := conn.User().Find(bson.M{"_id": name}).One(&u); err != nil && err.Error() == "not found" {
		return u, fmt.Errorf("User %s not found", name)
	}
	return u, nil
}
