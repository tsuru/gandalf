package gandalf

import (
	"io"
	"io/ioutil"
	"launchpad.net/mgo/bson"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func request(url string, b io.Reader, t *testing.T) (*httptest.ResponseRecorder, *http.Request) {
	request, err := http.NewRequest("POST", url, b)
	if err != nil {
		t.Errorf("Error when creating new request: %s", err)
		t.FailNow()
	}
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	return recorder, request
}

func readBody(b io.Reader, t *testing.T) string {
	body, err := ioutil.ReadAll(b)
	if err != nil {
		t.Errorf("Error when reading body: %s", err)
		t.FailNow()
	}
	return string(body)
}

func TestCreateUser(t *testing.T) {
}

func TestCreateProject(t *testing.T) {
	c := session.DB("gandalf").C("project")
	defer c.Remove(bson.M{"name": "some_project"})
	b := strings.NewReader(`{"name": "some_project"}`)
	recorder, request := request("/projects", b, t)
	CreateProject(recorder, request)
	got := readBody(recorder.Body, t)
	expected := "Project some_project successfuly created"
	if got != expected {
		t.Errorf(`Expected body to be "%s", got: "%s"`, expected, got)
	}
}

func TestCreateProjectShouldSaveInDB(t *testing.T) {
	b := strings.NewReader(`{"name": "myProject"}`)
	recorder, request := request("/projects", b, t)
	CreateProject(recorder, request)
	c := session.DB("gandalf").C("project")
	defer c.Remove(bson.M{"name": "myProject"})
	var p project
	err := c.Find(bson.M{"name": "myProject"}).One(&p)
	if err != nil {
		t.Errorf(`There was an error while retrieving project: "%s"`, err.Error())
	}
}

func TestCreateProjectShouldReturnErrorWhenNoParametersArePassed(t *testing.T) {
	b := strings.NewReader("{}")
	recorder, request := request("/projects", b, t)
	CreateProject(recorder, request)
	if recorder.Code != 400 {
		t.Errorf(`Expected code to be "400", got "%d"`, recorder.Code)
	}
	body := readBody(recorder.Body, t)
	expected := "Project needs a name"
	got := strings.Replace(body, "\n", "", -1)
	if got != expected {
		t.Errorf(`Expected body to matches: "%s", got: "%s"`, expected, got)
	}
}

func TestCreateProjectShouldReturnErrorWhenJsonIsBroken(t *testing.T) {
	b := strings.NewReader("{]ja9aW}")
	recorder, request := request("/projects", b, t)
	CreateProject(recorder, request)
	if recorder.Code != 400 {
		t.Errorf(`Expected code to be "400", got "%d"`, recorder.Code)
	}
	body := readBody(recorder.Body, t)
	expected := "Could not parse json: invalid character ']' looking for beginning of object key string"
	got := strings.Replace(body, "\n", "", -1)
	if got != expected {
		t.Errorf(`Expected body to matches: "%s", got: "%s"`, expected, got)
	}
}

func TestCreateProjectShouldReturnErrorWhenBodyIsEmpty(t *testing.T) {
	b := strings.NewReader("")
	recorder, request := request("/projects", b, t)
	CreateProject(recorder, request)
	if recorder.Code != 400 {
		t.Errorf(`Expected code to be "400", got "%d"`, recorder.Code)
	}
}
