package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/globocom/gandalf/repository"
	"github.com/globocom/gandalf/user"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
)

func GrantAccess(w http.ResponseWriter, r *http.Request) {
	rName := r.URL.Query().Get(":name")
	uName := r.URL.Query().Get(":username")
	if err := repository.GrantAccess(rName, uName); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	fmt.Fprintf(w, "Successfuly granted access to user \"%s\" into repository \"%s\"", uName, rName)
}

func RevokeAccess(w http.ResponseWriter, r *http.Request) {
	rName := r.URL.Query().Get(":name")
	uName := r.URL.Query().Get(":username")
	if err := repository.RevokeAccess(rName, uName); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest) // should return 404 when not found and 412 when cannot remove
		return
	}
	fmt.Fprintf(w, "Successfuly revoked access to user \"%s\" into repository \"%s\"", uName, rName)
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
	k := map[string]string{params["name"]: params["key"]}
	if err := user.AddKey(uName, k); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	fmt.Fprintf(w, "Key \"%s\" successfuly created", params["key"])
}

func RemoveKey(w http.ResponseWriter, r *http.Request) {
	uName := r.URL.Query().Get(":username")
	kName := r.URL.Query().Get(":keyname")
	if err := user.RemoveKey(uName, kName); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	fmt.Fprintf(w, "Key \"%s\" successfuly removed", kName)
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
