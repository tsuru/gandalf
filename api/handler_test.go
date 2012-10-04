package api

import (
	"bytes"
	"fmt"
	"github.com/globocom/gandalf/db"
	"github.com/globocom/gandalf/repository"
	"github.com/globocom/gandalf/user"
	"io"
	"io/ioutil"
	"labix.org/v2/mgo/bson"
	. "launchpad.net/gocheck"
	"net/http"
	"net/http/httptest"
	"strings"
)

type bufferCloser struct {
	*bytes.Buffer
}

func (b bufferCloser) Close() error { return nil }

func post(url string, b io.Reader, c *C) (*httptest.ResponseRecorder, *http.Request) {
	request, err := http.NewRequest("POST", url, b)
	c.Assert(err, IsNil)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	return recorder, request
}

func readBody(b io.Reader, c *C) string {
	body, err := ioutil.ReadAll(b)
	c.Assert(err, IsNil)
	return string(body)
}

func (s *S) TestNewUser(c *C) {
	b := strings.NewReader(`{"name": "brain", "key": ["some id_rsa.pub key.. use your imagination!"]}`)
	recorder, request := post("/user", b, c)
	NewUser(recorder, request)
	defer db.Session.User().Remove(bson.M{"_id": "brain"})
	c.Assert(recorder.Code, Equals, 200)
}

func (s *S) TestNewUserShouldSaveInDB(c *C) {
	b := strings.NewReader(`{"name": "brain", "key": ["some id_rsa.pub key.. use your imagination!"]}`)
	recorder, request := post("/user", b, c)
	NewUser(recorder, request)
	collection := db.Session.User()
	var u user.User
	err := collection.Find(bson.M{"_id": "brain"}).One(&u)
	defer collection.Remove(bson.M{"_id": "brain"})
	c.Assert(err, IsNil)
}

func (s *S) TestNewUserShouldRepassParseBodyErrors(c *C) {
	b := strings.NewReader("{]9afe}")
	recorder, request := post("/user", b, c)
	NewUser(recorder, request)
	body := readBody(recorder.Body, c)
	expected := "Could not parse json: invalid character ']' looking for beginning of object key string"
	got := strings.Replace(body, "\n", "", -1)
	c.Assert(got, Equals, expected)
}

func (s *S) TestNewUserShouldRequireUserName(c *C) {
	b := strings.NewReader(`{"name": ""}`)
	recorder, request := post("/user", b, c)
	NewUser(recorder, request)
	body := readBody(recorder.Body, c)
	expected := "Validation Error: user name is not valid"
	got := strings.Replace(body, "\n", "", -1)
	c.Assert(got, Equals, expected)
}

func (s *S) TestNewUserWihoutKey(c *C) {
	b := strings.NewReader(`{"name": "brain"}`)
	recorder, request := post("/user", b, c)
	NewUser(recorder, request)
	defer db.Session.User().Remove(bson.M{"_id": "brain"})
	c.Assert(recorder.Code, Equals, 200)
}

func (s *S) TestNewRepository(c *C) {
	defer db.Session.Repository().Remove(bson.M{"_id": "some_repository"})
	b := strings.NewReader(`{"name": "some_repository", "users": ["r2d2"]}`)
	recorder, request := post("/repository", b, c)
	NewRepository(recorder, request)
	got := readBody(recorder.Body, c)
	expected := "Repository some_repository successfuly created"
	c.Assert(got, Equals, expected)
}

func (s *S) TestNewRepositoryShouldSaveInDB(c *C) {
	b := strings.NewReader(`{"name": "myRepository", "users": ["r2d2"]}`)
	recorder, request := post("/repository", b, c)
	NewRepository(recorder, request)
	collection := db.Session.Repository()
	defer collection.Remove(bson.M{"_id": "myRepository"})
	var p repository.Repository
	err := collection.Find(bson.M{"_id": "myRepository"}).One(&p)
	c.Assert(err, IsNil)
}

func (s *S) TestNewRepositoryShouldSaveUserIdInRepository(c *C) {
	b := strings.NewReader(`{"name": "myRepository", "users": ["r2d2", "brain"]}`)
	recorder, request := post("/repository", b, c)
	NewRepository(recorder, request)
	collection := db.Session.Repository()
	defer collection.Remove(bson.M{"_id": "myRepository"})
	var p repository.Repository
	err := collection.Find(bson.M{"_id": "myRepository"}).One(&p)
	c.Assert(err, IsNil)
	c.Assert(len(p.Users), Not(Equals), 0)
}

func (s *S) TestNewRepositoryShouldReturnErrorWhenNoUserIsPassed(c *C) {
	b := strings.NewReader(`{"name": "myRepository"}`)
	recorder, request := post("/repository", b, c)
	NewRepository(recorder, request)
	c.Assert(recorder.Code, Equals, 400)
	body := readBody(recorder.Body, c)
	expected := "Validation Error: repository should have at least one user"
	got := strings.Replace(body, "\n", "", -1)
	c.Assert(got, Equals, expected)
}

func (s *S) TestNewRepositoryShouldReturnErrorWhenNoParametersArePassed(c *C) {
	b := strings.NewReader("{}")
	recorder, request := post("/repository", b, c)
	NewRepository(recorder, request)
	c.Assert(recorder.Code, Equals, 400)
	body := readBody(recorder.Body, c)
	expected := "Validation Error: repository name is not valid"
	got := strings.Replace(body, "\n", "", -1)
	c.Assert(got, Equals, expected)
}

func (s *S) TestParseBodyShouldMapBodyJsonToGivenStruct(c *C) {
	var p repository.Repository
	b := bufferCloser{bytes.NewBufferString(`{"name": "Dummy Repository"}`)}
	err := parseBody(b, &p)
	c.Assert(err, IsNil)
	expected := "Dummy Repository"
	c.Assert(p.Name, Equals, expected)
}

func (s *S) TestParseBodyShouldReturnErrorWhenJsonIsInvalid(c *C) {
	var p repository.Repository
	b := bufferCloser{bytes.NewBufferString("{]ja9aW}")}
	err := parseBody(b, &p)
	c.Assert(err, NotNil)
	expected := "Could not parse json: invalid character ']' looking for beginning of object key string"
	c.Assert(err.Error(), Equals, expected)
}

func (s *S) TestParseBodyShouldReturnErrorWhenBodyIsEmpty(c *C) {
	var p repository.Repository
	b := bufferCloser{bytes.NewBufferString("")}
	err := parseBody(b, &p)
	c.Assert(err, NotNil)
	expected := "Could not parse json: unexpected end of JSON input"
	c.Assert(err.Error(), Equals, expected)
}

func (s *S) TestParseBodyShouldReturnErrorWhenResultParamIsNotAPointer(c *C) {
	var p repository.Repository
	b := bufferCloser{bytes.NewBufferString(`{"name": "something"}`)}
	err := parseBody(b, p)
	c.Assert(err, NotNil)
	expected := "parseBody function cannot deal with struct. Use pointer"
	c.Assert(err.Error(), Equals, expected)
}

func (s *S) TestNewRepositoryShouldReturnErrorWhenBodyIsEmpty(c *C) {
	b := strings.NewReader("")
	recorder, request := post("/repository", b, c)
	NewRepository(recorder, request)
	c.Assert(recorder.Code, Equals, 400)
}

func (s *S) TestGrantAccess(c *C) {
	u, err := user.New("pippin", []string{})
	defer db.Session.User().Remove(bson.M{"_id": "pippin"})
	c.Assert(err, IsNil)
	r := repository.Repository{Name: "repo"}
	collection := db.Session.Repository()
	err = collection.Insert(&r)
	c.Assert(err, IsNil)
	defer collection.Remove(bson.M{"_id": "repo"})
	b := strings.NewReader(`{"users": ["pippin"]}`)
	url := fmt.Sprintf("/repository/%s/grant?:name=%s", r.Name, r.Name)
	rec, req := post(url, b, c)
	GrantAccess(rec, req)
	collection.Find(bson.M{"_id": "repo"}).One(&r)
	c.Assert(len(r.Users), Not(Equals), 0)
	c.Assert(r.Users[0], Equals, u.Name)
}

func (s *S) TestGrantAccessShouldReturn404WhenSingleUserDoesntExists(c *C) {
	r := repository.Repository{Name: "repo"}
	collection := db.Session.Repository()
	collection.Insert(&r)
	defer collection.Remove(bson.M{"_id": "repo"})
	url := fmt.Sprintf("/repository/%s/grant?:name=%s", r.Name, r.Name)
	b := strings.NewReader(`{"users": ["gandalf"]}`)
	rec, req := post(url, b, c)
	GrantAccess(rec, req)
	c.Assert(rec.Code, Equals, 404)
}

func (s *S) TestGrantAccessShouldNotInsertInexistentSingleUser(c *C) {
	r := repository.Repository{Name: "repo"}
	collection := db.Session.Repository()
	err := collection.Insert(&r)
	c.Assert(err, IsNil)
	defer collection.Remove(bson.M{"_id": "repo"})
	url := fmt.Sprintf("/repository/%s/grant?:name=%s", r.Name, r.Name)
	b := strings.NewReader(`{"users": ["gandalf"]}`)
	rec, req := post(url, b, c)
	GrantAccess(rec, req)
	err = db.Session.Repository().Find(bson.M{"_id": r.Name}).One(&r)
	c.Assert(err, IsNil)
	c.Assert(len(r.Users), Equals, 0)
}

func (s *S) TestGrantAccessShouldSkipUserGrantWhenMultipleUsersArePassed(c *C) {
	r := repository.Repository{Name: "somerepo"}
	err := db.Session.Repository().Insert(&r)
	defer db.Session.Repository().Remove(bson.M{"_id": r.Name})
	c.Assert(err, IsNil)
	u, err := user.New("gandalf", []string{})
	c.Assert(err, IsNil)
	defer db.Session.User().Remove(bson.M{"_id": u.Name})
	url := fmt.Sprintf("/repository/%s/grant?:name=%s", r.Name, r.Name)
	b := strings.NewReader(`{"users": ["gandalf", "frodo"]}`)
	rec, req := post(url, b, c)
	GrantAccess(rec, req)
	err = db.Session.Repository().Find(bson.M{"_id": r.Name}).One(&r)
	c.Assert(err, IsNil)
	c.Assert(len(r.Users), Equals, 1)
}

func (s *S) TestAddKey(c *C) {
	user, err := user.New("Frodo", []string{})
	c.Assert(err, IsNil)
	defer db.Session.User().Remove(bson.M{"_id": "Frodo"})
	b := strings.NewReader(`{"key": "a public key"}`)
	recorder, request := post(fmt.Sprintf("/user/%s/key?:name=%s", user.Name, user.Name), b, c)
	AddKey(recorder, request)
	got := readBody(recorder.Body, c)
	expected := "Key \"a public key\" successfuly created"
	c.Assert(got, Equals, expected)
}

func (s *S) TestAddKeyShouldReturnErorWhenUserDoesNotExists(c *C) {
	b := strings.NewReader(`{"key": "a public key"}`)
	recorder, request := post("/user/Frodo/key?:name=Frodo", b, c)
	AddKey(recorder, request)
	c.Assert(recorder.Code, Equals, 404)
}

func (s *S) TestAddKeyShouldRequireKey(c *C) {
	u := user.User{Name: "Frodo"}
	collection := db.Session.User()
	collection.Insert(&u)
	defer collection.Remove(bson.M{"_id": "Frodo"})
	b := strings.NewReader(`{"key": ""}`)
	recorder, request := post("/user/Frodo/key?:name=Frodo", b, c)
	AddKey(recorder, request)
	body := readBody(recorder.Body, c)
	expected := "A key is needed"
	got := strings.Replace(body, "\n", "", -1)
	c.Assert(got, Equals, expected)
}

func (s *S) TestRemoveUser(c *C) {
	u, err := user.New("username", []string{})
	c.Assert(err, IsNil)
	url := fmt.Sprintf("/user/%s/?:name=%s", u.Name, u.Name)
	request, err := http.NewRequest("DELETE", url, nil)
	c.Assert(err, IsNil)
	recorder := httptest.NewRecorder()
	RemoveUser(recorder, request)
	c.Assert(recorder.Code, Equals, 200)
}

func (s *S) TestRemoveUserShouldRemoveFromDB(c *C) {
	u, err := user.New("anuser", []string{})
	c.Assert(err, IsNil)
	url := fmt.Sprintf("/user/%s/?:name=%s", u.Name, u.Name)
	request, err := http.NewRequest("DELETE", url, nil)
	c.Assert(err, IsNil)
	recorder := httptest.NewRecorder()
	RemoveUser(recorder, request)
	collection := db.Session.User()
	lenght, err := collection.Find(bson.M{"_id": u.Name}).Count()
	c.Assert(err, IsNil)
	c.Assert(lenght, Equals, 0)
}

func (s *S) TestRemoveRepositoryShouldRemoveFromDB(c *C) {
	r, err := repository.New("myRepo", []string{"pippin"}, true)
	c.Assert(err, IsNil)
	url := fmt.Sprintf("repository/%s/?:name=%s", r.Name, r.Name)
	request, err := http.NewRequest("DELETE", url, nil)
	c.Assert(err, IsNil)
	recorder := httptest.NewRecorder()
	RemoveRepository(recorder, request)
	err = db.Session.Repository().Find(bson.M{"_id": r.Name}).One(&r)
	c.Assert(err, ErrorMatches, "^not found$")
}

func (s *S) TestRemoveRepositoryShouldReturn400OnFailure(c *C) {
	url := fmt.Sprintf("repository/%s/?:name=%s", "foo", "foo")
	request, err := http.NewRequest("DELETE", url, nil)
	c.Assert(err, IsNil)
	recorder := httptest.NewRecorder()
	RemoveRepository(recorder, request)
	c.Assert(recorder.Code, Equals, 400)
}
