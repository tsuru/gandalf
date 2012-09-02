package gandalf

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"labix.org/v2/mgo/bson"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type bufferCloser struct {
	*bytes.Buffer
}

func (b bufferCloser) Close() error { return nil }

func post(url string, b io.Reader, t *testing.T) (*httptest.ResponseRecorder, *http.Request) {
	request, err := http.NewRequest("POST", url, b)
	if err != nil {
		t.Errorf("Error when creating new request: %s", err)
		t.FailNow()
	}
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	return recorder, request
}

func createUser(name string) (u user, err error) {
	u = user{Name: "pippin"}
	c := session.DB("gandalf").C("user")
	err = c.Insert(&u)
	return
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
	b := strings.NewReader(`{"name": "brain", "key": ["some id_rsa.pub key.. use your imagination!"]}`)
	recorder, request := post("/user", b, t)
	CreateUser(recorder, request)
	c := session.DB("gandalf").C("user")
	defer c.Remove(bson.M{"_id": "brain"})
	if recorder.Code != 200 {
		t.Errorf(`Failed to create user, expected "%d" status code, got: "%d"`, 200, recorder.Code)
	}
}

func TestCreateUserShouldSaveInDB(t *testing.T) {
	b := strings.NewReader(`{"name": "brain", "key": ["some id_rsa.pub key.. use your imagination!"]}`)
	recorder, request := post("/user", b, t)
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
	recorder, request := post("/user", b, t)
	CreateUser(recorder, request)
	body := readBody(recorder.Body, t)
	expected := "Could not parse json: invalid character ']' looking for beginning of object key string"
	got := strings.Replace(body, "\n", "", -1)
	if got != expected {
		t.Errorf(`Expected error to matches: "%s", got: "%s"`, expected, got)
	}
}

func TestCreateUserShouldRequireUserName(t *testing.T) {
	b := strings.NewReader(`{"name": ""}`)
	recorder, request := post("/user", b, t)
	CreateUser(recorder, request)
	body := readBody(recorder.Body, t)
	expected := "User needs a name"
	got := strings.Replace(body, "\n", "", -1)
	if got != expected {
		t.Errorf(`Expected error to matches "%s", got: "%s"`, expected, got)
	}
}

func TestCreateUserWihoutKey(t *testing.T) {
	b := strings.NewReader(`{"name": "brain"}`)
	recorder, request := post("/user", b, t)
	CreateUser(recorder, request)
	c := session.DB("gandalf").C("user")
	defer c.Remove(bson.M{"_id": "brain"})
	if recorder.Code != 200 {
		t.Errorf(`Failed to create user, expected "%d" status code, got: "%d"`, 200, recorder.Code)
	}
}

func TestCreateRepository(t *testing.T) {
	c := session.DB("gandalf").C("repository")
	defer c.Remove(bson.M{"_id": "some_repository"})
	b := strings.NewReader(`{"name": "some_repository", "users": ["r2d2"]}`)
	recorder, request := post("/repository", b, t)
	CreateRepository(recorder, request)
	got := readBody(recorder.Body, t)
	expected := "Repository some_repository successfuly created"
	if got != expected {
		t.Errorf(`Expected body to be "%s", got: "%s"`, expected, got)
	}
}

func TestCreateRepositoryShouldSaveInDB(t *testing.T) {
	b := strings.NewReader(`{"name": "myRepository", "users": ["r2d2"]}`)
	recorder, request := post("/repository", b, t)
	CreateRepository(recorder, request)
	c := session.DB("gandalf").C("repository")
	defer c.Remove(bson.M{"_id": "myRepository"})
	var p repository
	err := c.Find(bson.M{"_id": "myRepository"}).One(&p)
	if err != nil {
		t.Errorf(`There was an error while retrieving repository: "%s"`, err.Error())
	}
}

func TestCreateRepositoryShouldSaveUserIdInRepository(t *testing.T) {
	b := strings.NewReader(`{"name": "myRepository", "users": ["r2d2", "brain"]}`)
	recorder, request := post("/repository", b, t)
	CreateRepository(recorder, request)
	c := session.DB("gandalf").C("repository")
	defer c.Remove(bson.M{"_id": "myRepository"})
	var p repository
	err := c.Find(bson.M{"_id": "myRepository"}).One(&p)
	if err != nil {
		t.Errorf(`There was an error while retrieving repository: "%s"`, err.Error())
	}
	if len(p.Users) == 0 {
		t.Errorf(`Expected user to be %s and %s, got empty.`, "r2d2", "brain")
	}
}

func TestCreateRepositoryShouldReturnErrorWhenNoUserIsPassed(t *testing.T) {
	b := strings.NewReader(`{"name": "myRepository"}`)
	recorder, request := post("/repository", b, t)
	CreateRepository(recorder, request)
	if recorder.Code != 400 {
		t.Errorf(`Expected code to be "400", got "%d"`, recorder.Code)
	}
	body := readBody(recorder.Body, t)
	expected := "Repository needs a user"
	got := strings.Replace(body, "\n", "", -1)
	if got != expected {
		t.Errorf(`Expected body to matches: "%s", got: "%s"`, expected, got)
	}
}

func TestCreateRepositoryShouldReturnErrorWhenNoParametersArePassed(t *testing.T) {
	b := strings.NewReader("{}")
	recorder, request := post("/repository", b, t)
	CreateRepository(recorder, request)
	if recorder.Code != 400 {
		t.Errorf(`Expected code to be "400", got "%d"`, recorder.Code)
	}
	body := readBody(recorder.Body, t)
	expected := "Repository needs a name"
	got := strings.Replace(body, "\n", "", -1)
	if got != expected {
		t.Errorf(`Expected body to matches: "%s", got: "%s"`, expected, got)
	}
}

func TestParseBodyShouldMapBodyJsonToGivenStruct(t *testing.T) {
	var p repository
	b := bufferCloser{bytes.NewBufferString(`{"name": "Dummy Repository"}`)}
	err := parseBody(b, &p)
	if err != nil {
		t.Errorf(`Expecting err to be nil, got: "%s"`, err.Error())
	}
	expected := "Dummy Repository"
	if p.Name != expected {
		t.Errorf(`Expecting err to be "%s", got: "%s"`, expected, p.Name)
	}
}

func TestParseBodyShouldReturnErrorWhenJsonIsInvalid(t *testing.T) {
	var p repository
	b := bufferCloser{bytes.NewBufferString("{]ja9aW}")}
	err := parseBody(b, &p)
	expected := "Could not parse json: invalid character ']' looking for beginning of object key string"
	if err.Error() != expected {
		t.Errorf(`Expected error to matches: "%s", got: "%s"`, expected, err.Error())
	}
}

func TestParseBodyShouldReturnErrorWhenBodyIsEmpty(t *testing.T) {
	var p repository
	b := bufferCloser{bytes.NewBufferString("")}
	err := parseBody(b, &p)
	expected := "Could not parse json: unexpected end of JSON input"
	if err.Error() != expected {
		t.Errorf(`Expected error to matches "%s", got: "%s"`, expected, err.Error())
	}
}

func TestParseBodyShouldReturnErrorWhenResultParamIsNotAPointer(t *testing.T) {
	var p repository
	b := bufferCloser{bytes.NewBufferString(`{"name": "something"}`)}
	err := parseBody(b, p)
	expected := "parseBody function cannot deal with struct. Use pointer"
	if err.Error() != expected {
		t.Errorf(`Expected error to matches "%s", got: "%s"`, expected, err.Error())
	}
}

func TestCreateRepositoryShouldReturnErrorWhenBodyIsEmpty(t *testing.T) {
	b := strings.NewReader("")
	recorder, request := post("/repository", b, t)
	CreateRepository(recorder, request)
	if recorder.Code != 400 {
		t.Errorf(`Expected code to be "400", got "%d"`, recorder.Code)
	}
}

func TestGrantAccess(t *testing.T) {
	u, err := createUser("pippin")
	c := session.DB("gandalf").C("user")
	defer c.Remove(bson.M{"_id": "pippin"})
	r := repository{Name: "repo"}
	c = Session.Repository()
	err = c.Insert(&r)
	if err != nil {
		t.Errorf(`Expected error to be nil, got %s`, err.Error())
	}
	defer c.Remove(bson.M{"_id": "repo"})
	b := strings.NewReader(`{"users": ["pippin"]}`)
	url := fmt.Sprintf("/repository/%s/grant?:name=%s", r.Name, r.Name)
	rec, req := post(url, b, t)
	GrantAccess(rec, req)
	c.Find(bson.M{"_id": "repo"}).One(&r)
	if len(r.Users) == 0 {
		t.Errorf(`Expected repository to have one user, got 0`)
		t.FailNow()
	}
	if r.Users[0] != u.Name {
		t.Errorf(`Expected repository's user to be %s, got %s`, u.Name, r.Users[0])
	}
}

func TestAddKey(t *testing.T) {
	user, err := createUser("Frodo")
	if err != nil {
		t.Errorf("Error while creating user: %s", err.Error())
		t.FailNow()
	}
	defer Session.User().Remove(bson.M{"_id": "Frodo"})
	b := strings.NewReader(`{"key": "a public key"}`)
	recorder, request := post(fmt.Sprintf("/user/%s/key?:name=%s", user.Name, user.Name), b, t)
	AddKey(recorder, request)
	got := readBody(recorder.Body, t)
	expected := "Key \"a public key\" successfuly created"
	if got != expected {
		t.Errorf(`Expected body to be "%s", got: "%s"`, expected, got)
	}
}

func TestAddKeyShouldReturnErorWhenUserDoesNotExists(t *testing.T) {
	b := strings.NewReader(`{"key": "a public key"}`)
	recorder, request := post("/user/Frodo/key?:name=Frodo", b, t)
	AddKey(recorder, request)
	if recorder.Code != 404 {
		t.Errorf(`Expected code to be "404", got "%d"`, recorder.Code)
	}
}

func TestAddKeyShouldRequireKey(t *testing.T) {
	user := user{Name: "Frodo"}
	c := session.DB("gandalf").C("user")
	c.Insert(&user)
	defer c.Remove(bson.M{"_id": "Frodo"})
	b := strings.NewReader(`{"key": ""}`)
	recorder, request := post("/user/Frodo/key?:name=Frodo", b, t)
	AddKey(recorder, request)
	body := readBody(recorder.Body, t)
	expected := "A key is needed"
	got := strings.Replace(body, "\n", "", -1)
	if got != expected {
		t.Errorf(`Expected error to matches "%s", got: "%s"`, expected, got)
	}
}
