package gandalf

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func CreateProject(w http.ResponseWriter, r *http.Request) {
	var p project
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = json.Unmarshal(b, &p)
	if err != nil {
		e := fmt.Sprintf("Could not parse json: %s", err.Error())
		http.Error(w, e, http.StatusBadRequest)
		return
	}
	if p.Name == "" {
		http.Error(w, "Project needs a name", http.StatusBadRequest)
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
