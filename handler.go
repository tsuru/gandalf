package gandalf

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"labix.org/v2/mgo/bson"
	"net/http"
	"reflect"
)

func AddKey(w http.ResponseWriter, r *http.Request) {
	u := user{Name: r.URL.Query().Get(":name")}
	c := session.DB("gandalf").C("user")
	err := c.Find(bson.M{"_id": u.Name}).One(&u)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	params := map[string]string{}
	err = parseBody(r.Body, &params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	u.Keys = append(u.Keys, params["key"])
	err = c.Update(bson.M{"_id": u.Name}, u)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "Key \"%s\" successfuly created", params["key"])
}

func CreateUser(w http.ResponseWriter, r *http.Request) {
	var u user
	err := parseBody(r.Body, &u)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if u.Name == "" {
		http.Error(w, "User needs a name", http.StatusBadRequest)
		return
	}
	c := session.DB("gandalf").C("user")
	err = c.Insert(&u)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "User %s successfuly created", u.Name)
}

func CreateRepository(w http.ResponseWriter, r *http.Request) {
	var p repository
	err := parseBody(r.Body, &p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if p.Name == "" {
		http.Error(w, "Repository needs a name", http.StatusBadRequest)
		return
	}
	if len(p.Users) == 0 {
		http.Error(w, "Repository needs a user", http.StatusBadRequest)
		return
	}
	c := session.DB("gandalf").C("repository")
	err = c.Insert(&p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "Repository %s successfuly created", p.Name)
}

func parseBody(body io.ReadCloser, result interface{}) error {
	if reflect.ValueOf(result).Kind() == reflect.Struct {
		return errors.New("parseBody function cannot deal with struct. Use pointer")
	}
	b, err := ioutil.ReadAll(body)
	if err != nil {
		return err
	}
	err = json.Unmarshal(b, &result)
	if err != nil {
		e := fmt.Sprintf("Could not parse json: %s", err.Error())
		return errors.New(e)
	}
	return nil
}
