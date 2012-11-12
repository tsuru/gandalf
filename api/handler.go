package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/globocom/gandalf/db"
	"github.com/globocom/gandalf/key"
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
	if err := parseBody(r.Body, &req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	for _, u := range req["users"] {
		// TODO (flaviamissi): query all only once, then iterate over them?
		if _, err := getUserOr404(u); err != nil {
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
	if err := c.Update(bson.M{"_id": repo.Name}, &repo); err != nil {
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
	k := key.Key{Name: params["name"], Content: params["key"]}
	if err := user.AddKey(uName, &k); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	fmt.Fprintf(w, "Key \"%s\" successfuly created", params["key"])
}

func NewUser(w http.ResponseWriter, r *http.Request) {
	var usr user.User
	if err := parseBody(r.Body, &usr); err != nil {
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
	if err := user.Remove(name); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Fprintf(w, "User \"%s\" successfuly removed\n", name)
}

func NewRepository(w http.ResponseWriter, r *http.Request) {
	var repo repository.Repository
	if err := parseBody(r.Body, &repo); err != nil {
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
	if err := repository.Remove(repo); err != nil {
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
	if err := json.Unmarshal(b, &result); err != nil {
		return errors.New(fmt.Sprintf("Could not parse json: %s", err.Error()))
	}
	return nil
}
