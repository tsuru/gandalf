package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/globocom/gandalf/db"
	"github.com/globocom/gandalf/repository"
	"github.com/globocom/gandalf/user"
	"io"
	"io/ioutil"
	"labix.org/v2/mgo/bson"
	"net/http"
	"reflect"
)

func GrantAccess(w http.ResponseWriter, r *http.Request) {
	// it's need a intermediary method to grant access to a user into a repository
	// something equivalent to what we have in NewUser handler
	repo := repository.Repository{Name: r.URL.Query().Get(":name")}
	c := db.Session.Repository()
	c.Find(bson.M{"_id": repo.Name}).One(&repo)
	req := map[string][]string{}
	err := parseBody(r.Body, &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for _, u := range req["users"] {
		// TODO (flaviamissi): query all only once, then iterate over them?
		_, err = getUserOr404(u)
		if err != nil {
			if len(req["users"]) == 1 {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			} else {
				// #TODO (flaviamissi): log a warning saying the user "u" was not found and skip it
				continue
			}
		}
		repo.Users = append(repo.Users, u)
	}
	err = c.Update(bson.M{"_id": repo.Name}, &repo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func AddKey(w http.ResponseWriter, r *http.Request) {
	params := map[string]string{}
	if err := parseBody(r.Body, &params); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if params["key"] == "" {
		http.Error(w, "A key is needed", http.StatusBadRequest)
		return
	}
	uName := r.URL.Query().Get(":name")
	if err := user.AddKey(uName, params["key"]); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	fmt.Fprintf(w, "Key \"%s\" successfuly created", params["key"])
}

func NewUser(w http.ResponseWriter, r *http.Request) {
	// I need some attention, somebody give me some love!
	var usr user.User
	err := parseBody(r.Body, &usr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	u, err := user.New(usr.Name, usr.Keys)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Fprintf(w, "User \"%s\" successfuly created\n", u.Name)
}

func RemoveUser(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get(":name")
	err := user.Remove(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Fprintf(w, "User \"%s\" successfuly removed\n", name)
}

func NewRepository(w http.ResponseWriter, r *http.Request) {
	var repo repository.Repository
	err := parseBody(r.Body, &repo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	rep, err := repository.New(repo.Name, repo.Users, repo.IsPublic)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Fprintf(w, "Repository \"%s\" successfuly created\n", rep.Name)
}

func RemoveRepository(w http.ResponseWriter, r *http.Request) {
	repo := &repository.Repository{Name: r.URL.Query().Get(":name")}
	err := repository.Remove(repo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Fprintf(w, "Repository \"%s\" successfuly removed\n", repo.Name)
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
