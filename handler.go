package gandalf

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
)

func CreateUser(w http.ResponseWriter, r *http.Request) {
	var u user
	err := parseBody(r.Body, &u)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if u.Key == "" {
		http.Error(w, "User needs a key", http.StatusBadRequest)
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

func CreateProject(w http.ResponseWriter, r *http.Request) {
	var p project
	err := parseBody(r.Body, &p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if p.Name == "" {
		http.Error(w, "Project needs a name", http.StatusBadRequest)
		return
	}
	if len(p.User) == 0 {
		http.Error(w, "Project needs a user", http.StatusBadRequest)
		return
	}
	c := session.DB("gandalf").C("project")
	err = c.Insert(&p)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "Project %s successfuly created", p.Name)
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
