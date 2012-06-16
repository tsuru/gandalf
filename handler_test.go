package gandalf

import (
	"bytes"
	"io"
	"io/ioutil"
	"launchpad.net/mgo/bson"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type bufferCloser struct {
	*bytes.Buffer
}

func (b bufferCloser) Close() error { return nil }

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
	b := strings.NewReader(`{"name": "brain", "key": "some id_rsa.pub key.. use your imagination!"}`)
	recorder, request := request("/user", b, t)
	CreateUser(recorder, request)
	c := session.DB("gandalf").C("user")
	defer c.Remove(bson.M{"_id": "brain"})
	if recorder.Code != 200 {
		t.Errorf(`Failed to create user, expected "%d" status code, got: "%d"`, 200, recorder.Code)
	}
}

func TestCreateUserShouldSaveInDB(t *testing.T) {
	b := strings.NewReader(`{"name": "brain", "key": "some id_rsa.pub key.. use your imagination!"}`)
	recorder, request := request("/user", b, t)
	CreateUser(recorder, request)
	c := session.DB("gandalf").C("user")
	var u user
	err := c.Find(bson.M{"_id": "brain"}).One(&u)
	defer c.Remove(bson.M{"_id": "brain"})
	if err != nil {
		t.Errorf(`Error when searching for user: "%s"`, err.Error())
	}
}

func TestCreateUserShouldRepassParseBodyErrors(t *testing.T) {
	b := strings.NewReader("{]9afe}")
	recorder, request := request("/user", b, t)
	CreateUser(recorder, request)
	body := readBody(recorder.Body, t)
	expected := "Could not parse json: invalid character ']' looking for beginning of object key string"
	got := strings.Replace(body, "\n", "", -1)
	if got != expected {
		t.Errorf(`Expected error to matches: "%s", got: "%s"`, expected, got)
	}
}

func TestCreateUserShouldRequireUserName(t *testing.T) {
}

func TestCreateUserShouldRequireUserKey(t *testing.T) {
	b := strings.NewReader(`{"name": "brain"}`)
	recorder, request := request("/user", b, t)
	CreateUser(recorder, request)
	body := readBody(recorder.Body, t)
	expected := "User needs a key"
	got := strings.Replace(body, "\n", "", -1)
	if got != expected {
		t.Errorf(`Expected error to matches "%s", got: "%s"`, expected, got)
	}
}

func TestCreateProject(t *testing.T) {
	c := session.DB("gandalf").C("project")
	defer c.Remove(bson.M{"name": "some_project"})
	b := strings.NewReader(`{"name": "some_project", "user": ["r2d2"]}`)
	recorder, request := request("/projects", b, t)
	CreateProject(recorder, request)
	got := readBody(recorder.Body, t)
	expected := "Project some_project successfuly created"
	if got != expected {
		t.Errorf(`Expected body to be "%s", got: "%s"`, expected, got)
	}
}

func TestCreateProjectShouldSaveInDB(t *testing.T) {
	b := strings.NewReader(`{"name": "myProject", "user": ["r2d2"]}`)
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

func TestCreateProjectShouldSaveUserIdInProject(t *testing.T) {
	b := strings.NewReader(`{"name": "myProject", "user": ["r2d2", "brain"]}`)
	recorder, request := request("/projects", b, t)
	CreateProject(recorder, request)
	c := session.DB("gandalf").C("project")
	defer c.Remove(bson.M{"name": "myProject"})
	var p project
	err := c.Find(bson.M{"name": "myProject"}).One(&p)
	if err != nil {
		t.Errorf(`There was an error while retrieving project: "%s"`, err.Error())
	}
	if len(p.User) == 0 {
		t.Errorf(`Expected user to be %s and %s, got empty.`, "r2d2", "brain")
	}
}

func TestCreateProjectShouldReturnErrorWhenNoUserIsPassed(t *testing.T) {
	b := strings.NewReader(`{"name": "myProject"}`)
	recorder, request := request("/projects", b, t)
	CreateProject(recorder, request)
	if recorder.Code != 400 {
		t.Errorf(`Expected code to be "400", got "%d"`, recorder.Code)
	}
	body := readBody(recorder.Body, t)
	expected := "Project needs a user"
	got := strings.Replace(body, "\n", "", -1)
	if got != expected {
		t.Errorf(`Expected body to matches: "%s", got: "%s"`, expected, got)
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

func TestParseBodyShouldMapBodyJsonToGivenStruct(t *testing.T) {
	var p project
	b := bufferCloser{bytes.NewBufferString(`{"name": "Dummy Project"}`)}
	err := parseBody(b, &p)
	if err != nil {
		t.Errorf(`Expecting err to be nil, got: "%s"`, err.Error())
	}
	expected := "Dummy Project"
	if p.Name != expected {
		t.Errorf(`Expecting err to be "%s", got: "%s"`, expected, p.Name)
	}
}

func TestParseBodyShouldReturnErrorWhenJsonIsInvalid(t *testing.T) {
	var p project
	b := bufferCloser{bytes.NewBufferString("{]ja9aW}")}
	err := parseBody(b, &p)
	expected := "Could not parse json: invalid character ']' looking for beginning of object key string"
	if err.Error() != expected {
		t.Errorf(`Expected error to matches: "%s", got: "%s"`, expected, err.Error())
	}
}

func TestParseBodyShouldReturnErrorWhenBodyIsEmpty(t *testing.T) {
	var p project
	b := bufferCloser{bytes.NewBufferString("")}
	err := parseBody(b, &p)
	expected := "Could not parse json: unexpected end of JSON input"
	if err.Error() != expected {
		t.Errorf(`Expected error to matches "%s", got: "%s"`, expected, err.Error())
	}
}

func TestParseBodyShouldReturnErrorWhenResultParamIsNotAPointer(t *testing.T) {
	var p project
	b := bufferCloser{bytes.NewBufferString(`{"name": "something"}`)}
	err := parseBody(b, p)
	expected := "parseBody function cannot deal with struct. Use pointer"
	if err.Error() != expected {
		t.Errorf(`Expected error to matches "%s", got: "%s"`, expected, err.Error())
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
