// Copyright 2015 gandalf authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"net/http/httptest"
	"os"
	"path"
	"strings"

	"github.com/tsuru/config"
	"github.com/tsuru/gandalf/db"
	"github.com/tsuru/gandalf/fs"
	"github.com/tsuru/gandalf/multipartzip"
	"github.com/tsuru/gandalf/repository"
	"github.com/tsuru/gandalf/user"
	"gopkg.in/check.v1"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type bufferCloser struct {
	*bytes.Buffer
}

func (b bufferCloser) Close() error { return nil }

func get(url string, b io.Reader, c *check.C) (*httptest.ResponseRecorder, *http.Request) {
	return request("GET", url, b, c)
}

func post(url string, b io.Reader, c *check.C) (*httptest.ResponseRecorder, *http.Request) {
	return request("POST", url, b, c)
}

func put(url string, b io.Reader, c *check.C) (*httptest.ResponseRecorder, *http.Request) {
	return request("PUT", url, b, c)
}

func del(url string, b io.Reader, c *check.C) (*httptest.ResponseRecorder, *http.Request) {
	return request("DELETE", url, b, c)
}

func request(method, url string, b io.Reader, c *check.C) (*httptest.ResponseRecorder, *http.Request) {
	request, err := http.NewRequest(method, url, b)
	c.Assert(err, check.IsNil)
	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()
	return recorder, request
}

func readBody(b io.Reader, c *check.C) string {
	body, err := ioutil.ReadAll(b)
	c.Assert(err, check.IsNil)
	return string(body)
}

func (s *S) authKeysContent(c *check.C) string {
	authKeysPath := path.Join(os.Getenv("HOME"), ".ssh", "authorized_keys")
	f, err := fs.Filesystem().OpenFile(authKeysPath, os.O_RDWR|os.O_EXCL, 0755)
	c.Assert(err, check.IsNil)
	content, err := ioutil.ReadAll(f)
	c.Assert(err, check.IsNil)
	return string(content)
}

func (s *S) TestMaxMemoryValueShouldComeFromGandalfConf(c *check.C) {
	config.Set("api:request:maxMemory", 1024)
	oldMaxMemory := maxMemory
	maxMemory = 0
	defer func() {
		maxMemory = oldMaxMemory
	}()
	c.Assert(maxMemoryValue(), check.Equals, uint(1024))
}

func (s *S) TestMaxMemoryValueDontResetMaxMemory(c *check.C) {
	config.Set("api:request:maxMemory", 1024)
	oldMaxMemory := maxMemory
	maxMemory = 359
	defer func() {
		maxMemory = oldMaxMemory
	}()
	c.Assert(maxMemoryValue(), check.Equals, uint(359))
}

func (s *S) TestAccessParametersShouldReturnErrorWhenInvalidJSONInput(c *check.C) {
	b := bufferCloser{bytes.NewBufferString(``)}
	_, _, err := accessParameters(b)
	c.Assert(err, check.ErrorMatches, `^Could not parse json: .+$`)
	b = bufferCloser{bytes.NewBufferString(`{`)}
	_, _, err = accessParameters(b)
	c.Assert(err, check.ErrorMatches, `^Could not parse json: .+$`)
	b = bufferCloser{bytes.NewBufferString(`bang`)}
	_, _, err = accessParameters(b)
	c.Assert(err, check.ErrorMatches, `^Could not parse json: .+$`)
	b = bufferCloser{bytes.NewBufferString(` `)}
	_, _, err = accessParameters(b)
	c.Assert(err, check.ErrorMatches, `^Could not parse json: .+$`)
}

func (s *S) TestAccessParametersShouldReturnErrorWhenNoUserListProvided(c *check.C) {
	b := bufferCloser{bytes.NewBufferString(`{"users": "oneuser"}`)}
	_, _, err := accessParameters(b)
	c.Assert(err, check.ErrorMatches, `^Could not parse json: json: cannot unmarshal string into Go value of type \[\]string$`)
	b = bufferCloser{bytes.NewBufferString(`{"repositories": ["barad-dur"]}`)}
	_, _, err = accessParameters(b)
	c.Assert(err, check.ErrorMatches, `^It is need a user list$`)
}

func (s *S) TestAccessParametersShouldReturnErrorWhenNoRepositoryListProvided(c *check.C) {
	b := bufferCloser{bytes.NewBufferString(`{"users": ["nazgul"]}`)}
	_, _, err := accessParameters(b)
	c.Assert(err, check.ErrorMatches, `^It is need a repository list$`)
}

func (s *S) TestNewUser(c *check.C) {
	b := strings.NewReader(fmt.Sprintf(`{"name": "brain", "keys": {"keyname": %q}}`, rawKey))
	recorder, request := post("/user", b, c)
	s.router.ServeHTTP(recorder, request)
	conn, err := db.Conn()
	c.Assert(err, check.IsNil)
	defer conn.Close()
	defer conn.User().Remove(bson.M{"_id": "brain"})
	defer conn.Key().Remove(bson.M{"username": "brain"})
	body, err := ioutil.ReadAll(recorder.Body)
	c.Assert(err, check.IsNil)
	c.Assert(string(body), check.Equals, "User \"brain\" successfully created\n")
	c.Assert(recorder.Code, check.Equals, 200)
}

func (s *S) TestNewUserShouldSaveInDB(c *check.C) {
	b := strings.NewReader(`{"name": "brain", "keys": {"content": "some id_rsa.pub key.. use your imagination!", "name": "somekey"}}`)
	recorder, request := post("/user", b, c)
	s.router.ServeHTTP(recorder, request)
	conn, err := db.Conn()
	c.Assert(err, check.IsNil)
	defer conn.Close()
	defer conn.User().Remove(bson.M{"_id": "brain"})
	defer conn.Key().Remove(bson.M{"username": "brain"})
	var u user.User
	err = conn.User().Find(bson.M{"_id": "brain"}).One(&u)
	c.Assert(err, check.IsNil)
	c.Assert(u.Name, check.Equals, "brain")
}

func (s *S) TestNewUserShouldRepassParseBodyErrors(c *check.C) {
	b := strings.NewReader("{]9afe}")
	recorder, request := post("/user", b, c)
	s.router.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusBadRequest)
	body := readBody(recorder.Body, c)
	expected := "Got error while parsing body: Could not parse json: invalid character ']' looking for beginning of object key string"
	got := strings.Replace(body, "\n", "", -1)
	c.Assert(got, check.Equals, expected)
}

func (s *S) TestNewUserShouldRequireUserName(c *check.C) {
	b := strings.NewReader(`{"name": ""}`)
	recorder, request := post("/user", b, c)
	s.router.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusBadRequest)
	body := readBody(recorder.Body, c)
	expected := "username is not valid"
	got := strings.Replace(body, "\n", "", -1)
	c.Assert(got, check.Equals, expected)
}

func (s *S) TestNewUserWihoutKeys(c *check.C) {
	b := strings.NewReader(`{"name": "brain"}`)
	recorder, request := post("/user", b, c)
	s.router.ServeHTTP(recorder, request)
	conn, err := db.Conn()
	c.Assert(err, check.IsNil)
	defer conn.Close()
	defer conn.User().Remove(bson.M{"_id": "brain"})
	c.Assert(recorder.Code, check.Equals, 200)
}

func (s *S) TestNewUserDuplicate(c *check.C) {
	b := strings.NewReader(fmt.Sprintf(`{"name": "brain", "keys": {"keyname": %q}}`, rawKey))
	recorder, request := post("/user", b, c)
	s.router.ServeHTTP(recorder, request)
	conn, err := db.Conn()
	c.Assert(err, check.IsNil)
	defer conn.Close()
	defer conn.User().Remove(bson.M{"_id": "brain"})
	defer conn.Key().Remove(bson.M{"username": "brain"})
	b = strings.NewReader(`{"name":"brain"}`)
	recorder, request = post("/user", b, c)
	s.router.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusConflict)
	c.Assert(recorder.Body.String(), check.Equals, "user already exists\n")
}

func (s *S) TestGetRepository(c *check.C) {
	r := repository.Repository{Name: "onerepo"}
	conn, err := db.Conn()
	c.Assert(err, check.IsNil)
	defer conn.Close()
	err = conn.Repository().Insert(&r)
	c.Assert(err, check.IsNil)
	defer conn.Repository().Remove(bson.M{"_id": r.Name})
	recorder, request := get("/repository/onerepo", nil, c)
	s.router.ServeHTTP(recorder, request)
	body, err := ioutil.ReadAll(recorder.Body)
	c.Assert(err, check.IsNil)
	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	c.Assert(err, check.IsNil)
	expected := map[string]interface{}{
		"name":    r.Name,
		"public":  r.IsPublic,
		"ssh_url": r.ReadWriteURL(),
		"git_url": r.ReadOnlyURL(),
	}
	c.Assert(data, check.DeepEquals, expected)
}

func (s *S) TestGetRepositoryWithNamespace(c *check.C) {
	r := repository.Repository{Name: "onenamespace/onerepo"}
	conn, err := db.Conn()
	c.Assert(err, check.IsNil)
	defer conn.Close()
	err = conn.Repository().Insert(&r)
	c.Assert(err, check.IsNil)
	defer conn.Repository().Remove(bson.M{"_id": r.Name})
	recorder, request := get("/repository/onenamespace/onerepo", nil, c)
	s.router.ServeHTTP(recorder, request)
	body, err := ioutil.ReadAll(recorder.Body)
	c.Assert(err, check.IsNil)
	var data map[string]interface{}
	err = json.Unmarshal(body, &data)
	c.Assert(err, check.IsNil)
	expected := map[string]interface{}{
		"name":    r.Name,
		"public":  r.IsPublic,
		"ssh_url": r.ReadWriteURL(),
		"git_url": r.ReadOnlyURL(),
	}
	c.Assert(data, check.DeepEquals, expected)
}

func (s *S) TestGetRepositoryDoesNotExist(c *check.C) {
	recorder, request := get("/repository/doesnotexist", nil, c)
	s.router.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusNotFound)
}

func (s *S) TestNewRepository(c *check.C) {
	conn, err := db.Conn()
	c.Assert(err, check.IsNil)
	defer conn.Close()
	defer conn.Repository().Remove(bson.M{"_id": "some_repository"})
	b := strings.NewReader(`{"name": "some_repository", "users": ["r2d2"]}`)
	recorder, request := post("/repository", b, c)
	s.router.ServeHTTP(recorder, request)
	got := readBody(recorder.Body, c)
	expected := "Repository \"some_repository\" successfully created\n"
	c.Assert(got, check.Equals, expected)
}

func (s *S) TestNewRepositoryShouldSaveInDB(c *check.C) {
	b := strings.NewReader(`{"name": "myRepository", "users": ["r2d2"]}`)
	recorder, request := post("/repository", b, c)
	s.router.ServeHTTP(recorder, request)
	conn, err := db.Conn()
	c.Assert(err, check.IsNil)
	defer conn.Close()
	collection := conn.Repository()
	defer collection.Remove(bson.M{"_id": "myRepository"})
	var p repository.Repository
	err = collection.Find(bson.M{"_id": "myRepository"}).One(&p)
	c.Assert(err, check.IsNil)
}

func (s *S) TestNewRepositoryShouldSaveUserIdInRepository(c *check.C) {
	b := strings.NewReader(`{"name": "myRepository", "users": ["r2d2", "brain"]}`)
	recorder, request := post("/repository", b, c)
	s.router.ServeHTTP(recorder, request)
	conn, err := db.Conn()
	c.Assert(err, check.IsNil)
	defer conn.Close()
	collection := conn.Repository()
	defer collection.Remove(bson.M{"_id": "myRepository"})
	var p repository.Repository
	err = collection.Find(bson.M{"_id": "myRepository"}).One(&p)
	c.Assert(err, check.IsNil)
	c.Assert(len(p.Users), check.Not(check.Equals), 0)
}

func (s *S) TestNewRepositoryDuplicate(c *check.C) {
	b := strings.NewReader(`{"name": "myRepository", "users": ["r2d2"]}`)
	recorder, request := post("/repository", b, c)
	s.router.ServeHTTP(recorder, request)
	conn, err := db.Conn()
	c.Assert(err, check.IsNil)
	defer conn.Close()
	collection := conn.Repository()
	defer collection.Remove(bson.M{"_id": "myRepository"})
	b = strings.NewReader(`{"name": "myRepository", "users": ["r2d2"]}`)
	recorder, request = post("/repository", b, c)
	s.router.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusConflict)
	c.Assert(recorder.Body.String(), check.Equals, "repository already exists\n")
}

func (s *S) TestNewRepositoryShouldReturnErrorWhenNoUserIsPassed(c *check.C) {
	b := strings.NewReader(`{"name": "myRepository"}`)
	recorder, request := post("/repository", b, c)
	s.router.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, 400)
	body := readBody(recorder.Body, c)
	expected := "repository should have at least one user"
	got := strings.Replace(body, "\n", "", -1)
	c.Assert(got, check.Equals, expected)
}

func (s *S) TestNewRepositoryShouldReturnErrorWhenNoParametersArePassed(c *check.C) {
	b := strings.NewReader("{}")
	recorder, request := post("/repository", b, c)
	s.router.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, 400)
	body := readBody(recorder.Body, c)
	expected := "repository name is not valid"
	got := strings.Replace(body, "\n", "", -1)
	c.Assert(got, check.Equals, expected)
}

func (s *S) TestParseBodyShouldMapBodyJsonToGivenStruct(c *check.C) {
	var p repository.Repository
	b := bufferCloser{bytes.NewBufferString(`{"name": "Dummy Repository"}`)}
	err := parseBody(b, &p)
	c.Assert(err, check.IsNil)
	expected := "Dummy Repository"
	c.Assert(p.Name, check.Equals, expected)
}

func (s *S) TestParseBodyShouldMapBodyEmptyJsonToADict(c *check.C) {
	dict := make(map[string]interface{})
	b := bufferCloser{bytes.NewBufferString(`{"name": "Test", "isPublic": false, "users": []}`)}
	err := parseBody(b, &dict)
	c.Assert(err, check.IsNil)
	expected := map[string]interface{}{
		"name":     "Test",
		"isPublic": false,
		"users":    []interface{}{},
	}
	c.Assert(dict, check.DeepEquals, expected)
}

func (s *S) TestParseBodyShouldMapBodyJsonAndUpdateMap(c *check.C) {
	dict := map[string]interface{}{
		"isPublic":      false,
		"users":         []string{"merry"},
		"readonlyusers": []string{"pippin"},
	}
	b := bufferCloser{bytes.NewBufferString(`{"name": "Test", "users": []}`)}
	err := parseBody(b, &dict)
	c.Assert(err, check.IsNil)
	expected := map[string]interface{}{
		"name":          "Test",
		"isPublic":      false,
		"users":         []interface{}{},
		"readonlyusers": []string{"pippin"},
	}
	c.Assert(dict, check.DeepEquals, expected)
}

func (s *S) TestParseBodyShouldReturnErrorWhenJsonIsInvalid(c *check.C) {
	var p repository.Repository
	b := bufferCloser{bytes.NewBufferString("{]ja9aW}")}
	err := parseBody(b, &p)
	c.Assert(err, check.NotNil)
	expected := "Could not parse json: invalid character ']' looking for beginning of object key string"
	c.Assert(err.Error(), check.Equals, expected)
}

func (s *S) TestParseBodyShouldReturnErrorWhenBodyIsEmpty(c *check.C) {
	var p repository.Repository
	b := bufferCloser{bytes.NewBufferString("")}
	err := parseBody(b, &p)
	c.Assert(err, check.NotNil)
	c.Assert(err, check.ErrorMatches, `^Could not parse json:.*$`)
}

func (s *S) TestParseBodyShouldReturnErrorWhenResultParamIsNotAPointer(c *check.C) {
	var p repository.Repository
	b := bufferCloser{bytes.NewBufferString(`{"name": "something"}`)}
	err := parseBody(b, p)
	c.Assert(err, check.NotNil)
	expected := "parseBody function cannot deal with struct. Use pointer"
	c.Assert(err.Error(), check.Equals, expected)
}

func (s *S) TestNewRepositoryShouldReturnErrorWhenBodyIsEmpty(c *check.C) {
	b := strings.NewReader("")
	recorder, request := post("/repository", b, c)
	s.router.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, 400)
}

func (s *S) TestGrantAccessUpdatesReposDocument(c *check.C) {
	u, err := user.New("pippin", map[string]string{})
	conn, err := db.Conn()
	c.Assert(err, check.IsNil)
	defer conn.Close()
	defer conn.User().Remove(bson.M{"_id": "pippin"})
	c.Assert(err, check.IsNil)
	r := repository.Repository{Name: "onerepo"}
	err = conn.Repository().Insert(&r)
	c.Assert(err, check.IsNil)
	defer conn.Repository().Remove(bson.M{"_id": r.Name})
	r2 := repository.Repository{Name: "otherepo"}
	err = conn.Repository().Insert(&r2)
	c.Assert(err, check.IsNil)
	defer conn.Repository().Remove(bson.M{"_id": r2.Name})
	b := bytes.NewBufferString(fmt.Sprintf(`{"repositories": ["%s", "%s"], "users": ["%s"]}`, r.Name, r2.Name, u.Name))
	rec, req := post("/repository/grant", b, c)
	s.router.ServeHTTP(rec, req)
	var repos []repository.Repository
	err = conn.Repository().Find(bson.M{"_id": bson.M{"$in": []string{r.Name, r2.Name}}}).All(&repos)
	c.Assert(err, check.IsNil)
	c.Assert(rec.Code, check.Equals, 200)
	for _, repo := range repos {
		c.Assert(repo.Users, check.DeepEquals, []string{u.Name})
	}
}

func (s *S) TestGrantAccessReadOnlyUpdatesReposDocument(c *check.C) {
	u, err := user.New("pippin", map[string]string{})
	conn, err := db.Conn()
	c.Assert(err, check.IsNil)
	defer conn.Close()
	defer conn.User().Remove(bson.M{"_id": "pippin"})
	c.Assert(err, check.IsNil)
	r := repository.Repository{Name: "onerepo"}
	err = conn.Repository().Insert(&r)
	c.Assert(err, check.IsNil)
	defer conn.Repository().Remove(bson.M{"_id": r.Name})
	r2 := repository.Repository{Name: "otherepo"}
	err = conn.Repository().Insert(&r2)
	c.Assert(err, check.IsNil)
	defer conn.Repository().Remove(bson.M{"_id": r2.Name})
	b := bytes.NewBufferString(fmt.Sprintf(`{"repositories": ["%s", "%s"], "users": ["%s"]}`, r.Name, r2.Name, u.Name))
	rec, req := post("/repository/grant?readonly=yes", b, c)
	s.router.ServeHTTP(rec, req)
	var repos []repository.Repository
	err = conn.Repository().Find(bson.M{"_id": bson.M{"$in": []string{r.Name, r2.Name}}}).All(&repos)
	c.Assert(err, check.IsNil)
	c.Assert(rec.Code, check.Equals, 200)
	for _, repo := range repos {
		c.Assert(repo.ReadOnlyUsers, check.DeepEquals, []string{u.Name})
	}
}

func (s *S) TestGrantAccessRepositoryNotFound(c *check.C) {
	b := bytes.NewBufferString(`{"repositories":["super-repo"],"users":["someuser"]}`)
	rec, req := post("/repository/grant?readonly=yes", b, c)
	s.router.ServeHTTP(rec, req)
	c.Assert(rec.Code, check.Equals, http.StatusNotFound)
}

func (s *S) TestRevokeAccessUpdatesReposDocument(c *check.C) {
	r := repository.Repository{Name: "onerepo", Users: []string{"Umi", "Luke"}}
	conn, err := db.Conn()
	c.Assert(err, check.IsNil)
	defer conn.Close()
	err = conn.Repository().Insert(&r)
	c.Assert(err, check.IsNil)
	defer conn.Repository().Remove(bson.M{"_id": r.Name})
	r2 := repository.Repository{Name: "otherepo", Users: []string{"Umi", "Luke"}}
	err = conn.Repository().Insert(&r2)
	c.Assert(err, check.IsNil)
	defer conn.Repository().Remove(bson.M{"_id": r2.Name})
	b := bytes.NewBufferString(fmt.Sprintf(`{"repositories": ["%s", "%s"], "users": ["Umi"]}`, r.Name, r2.Name))
	rec, req := del("/repository/revoke", b, c)
	s.router.ServeHTTP(rec, req)
	var repos []repository.Repository
	err = conn.Repository().Find(bson.M{"_id": bson.M{"$in": []string{r.Name, r2.Name}}}).All(&repos)
	c.Assert(err, check.IsNil)
	for _, repo := range repos {
		c.Assert(repo.Users, check.DeepEquals, []string{"Luke"})
	}
}

func (s *S) TestRevokeAccessReadOnlyUpdatesReposDocument(c *check.C) {
	r := repository.Repository{Name: "onerepo", ReadOnlyUsers: []string{"Umi", "Luke"}}
	conn, err := db.Conn()
	c.Assert(err, check.IsNil)
	defer conn.Close()
	err = conn.Repository().Insert(&r)
	c.Assert(err, check.IsNil)
	defer conn.Repository().Remove(bson.M{"_id": r.Name})
	r2 := repository.Repository{Name: "otherepo", ReadOnlyUsers: []string{"Umi", "Luke"}}
	err = conn.Repository().Insert(&r2)
	c.Assert(err, check.IsNil)
	defer conn.Repository().Remove(bson.M{"_id": r2.Name})
	b := bytes.NewBufferString(fmt.Sprintf(`{"repositories": ["%s", "%s"], "users": ["Umi"]}`, r.Name, r2.Name))
	rec, req := del("/repository/revoke", b, c)
	s.router.ServeHTTP(rec, req)
	var repos []repository.Repository
	err = conn.Repository().Find(bson.M{"_id": bson.M{"$in": []string{r.Name, r2.Name}}}).All(&repos)
	c.Assert(err, check.IsNil)
	for _, repo := range repos {
		c.Assert(repo.ReadOnlyUsers, check.DeepEquals, []string{"Luke"})
	}
}

func (s *S) TestRevokeAccessRepositoryNotFound(c *check.C) {
	b := bytes.NewBufferString(`{"repositories":["super-repo"],"users":["someuser"]}`)
	rec, req := del("/repository/revoke", b, c)
	s.router.ServeHTTP(rec, req)
	c.Assert(rec.Code, check.Equals, http.StatusNotFound)
}

func (s *S) TestAddKey(c *check.C) {
	usr, err := user.New("Frodo", map[string]string{})
	c.Assert(err, check.IsNil)
	defer user.Remove(usr.Name)
	b := strings.NewReader(fmt.Sprintf(`{"keyname": %q}`, rawKey))
	recorder, request := post(fmt.Sprintf("/user/%s/key", usr.Name), b, c)
	s.router.ServeHTTP(recorder, request)
	got := readBody(recorder.Body, c)
	expected := "Key(s) successfully created"
	c.Assert(got, check.Equals, expected)
	c.Assert(recorder.Code, check.Equals, 200)
	var k user.Key
	conn, err := db.Conn()
	c.Assert(err, check.IsNil)
	defer conn.Close()
	err = conn.Key().Find(bson.M{"name": "keyname", "username": usr.Name}).One(&k)
	c.Assert(err, check.IsNil)
	c.Assert(k.Body, check.Equals, keyBody)
	c.Assert(k.Comment, check.Equals, keyComment)
}

func (s *S) TestUpdateKey(c *check.C) {
	usr, err := user.New("Frodo", map[string]string{})
	c.Assert(err, check.IsNil)
	defer user.Remove(usr.Name)
	err = user.AddKey(usr.Name, map[string]string{"keyname": rawKey})
	c.Assert(err, check.IsNil)
	defer user.RemoveKey(usr.Name, "keyname")
	b := strings.NewReader(otherKey)
	recorder, request := put(fmt.Sprintf("/user/%s/key/keyname", usr.Name), b, c)
	s.router.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, 200)
	c.Assert(recorder.Body.String(), check.Equals, `Key "keyname" successfully updated!`)
	var k user.Key
	conn, err := db.Conn()
	c.Assert(err, check.IsNil)
	defer conn.Close()
	err = conn.Key().Find(bson.M{"name": "keyname", "username": usr.Name}).One(&k)
	c.Assert(err, check.IsNil)
	c.Assert(k.Body, check.Equals, otherKey+"\n")
}

func (s *S) TestUpdateKeyUserNotFound(c *check.C) {
	b := strings.NewReader(rawKey)
	recorder, request := put("/user/frodo/key/keyname", b, c)
	s.router.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusNotFound)
	c.Assert(recorder.Body.String(), check.Equals, user.ErrUserNotFound.Error()+"\n")
}

func (s *S) TestUpdateKeyNotFound(c *check.C) {
	usr, err := user.New("Frodo", map[string]string{})
	c.Assert(err, check.IsNil)
	defer user.Remove(usr.Name)
	b := strings.NewReader(rawKey)
	recorder, request := put(fmt.Sprintf("/user/%s/key/keyname", usr.Name), b, c)
	s.router.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusNotFound)
	c.Assert(recorder.Body.String(), check.Equals, user.ErrKeyNotFound.Error()+"\n")
}

func (s *S) TestUpdateKeyInvalidKey(c *check.C) {
	usr, err := user.New("Frodo", map[string]string{})
	c.Assert(err, check.IsNil)
	defer user.Remove(usr.Name)
	err = user.AddKey(usr.Name, map[string]string{"keyname": rawKey})
	c.Assert(err, check.IsNil)
	defer user.RemoveKey(usr.Name, "keyname")
	b := strings.NewReader("invalid-key")
	recorder, request := put(fmt.Sprintf("/user/%s/key/keyname", usr.Name), b, c)
	s.router.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusBadRequest)
	c.Assert(recorder.Body.String(), check.Equals, user.ErrInvalidKey.Error()+"\n")
}

func (s *S) TestAddPostReceiveHookRepository(c *check.C) {
	b := strings.NewReader(`{"repositories": ["some-repo"], "content": "some content"}`)
	recorder, request := post("/hook/post-receive", b, c)
	s.router.ServeHTTP(recorder, request)
	got := readBody(recorder.Body, c)
	expected := "hook post-receive successfully created for [some-repo]\n"
	c.Assert(got, check.Equals, expected)
	c.Assert(recorder.Code, check.Equals, 200)
	file, err := fs.Filesystem().OpenFile("/var/lib/gandalf/repositories/some-repo.git/hooks/post-receive", os.O_RDONLY, 0755)
	defer file.Close()
	content, err := ioutil.ReadAll(file)
	c.Assert(err, check.IsNil)
	c.Assert(string(content), check.Equals, "some content")
}

func (s *S) TestAddPreReceiveHookRepository(c *check.C) {
	b := strings.NewReader(`{"repositories": ["some-repo"], "content": "some content"}`)
	recorder, request := post("/hook/pre-receive", b, c)
	s.router.ServeHTTP(recorder, request)
	got := readBody(recorder.Body, c)
	expected := "hook pre-receive successfully created for [some-repo]\n"
	c.Assert(got, check.Equals, expected)
	c.Assert(recorder.Code, check.Equals, 200)
	file, err := fs.Filesystem().OpenFile("/var/lib/gandalf/repositories/some-repo.git/hooks/pre-receive", os.O_RDONLY, 0755)
	defer file.Close()
	content, err := ioutil.ReadAll(file)
	c.Assert(err, check.IsNil)
	c.Assert(string(content), check.Equals, "some content")
}

func (s *S) TestAddUpdateReceiveHookRepository(c *check.C) {
	b := strings.NewReader(`{"repositories": ["some-repo"], "content": "some content"}`)
	recorder, request := post("/hook/update", b, c)
	s.router.ServeHTTP(recorder, request)
	got := readBody(recorder.Body, c)
	expected := "hook update successfully created for [some-repo]\n"
	c.Assert(got, check.Equals, expected)
	c.Assert(recorder.Code, check.Equals, 200)
	file, err := fs.Filesystem().OpenFile("/var/lib/gandalf/repositories/some-repo.git/hooks/update", os.O_RDONLY, 0755)
	defer file.Close()
	content, err := ioutil.ReadAll(file)
	c.Assert(err, check.IsNil)
	c.Assert(string(content), check.Equals, "some content")
}

func (s *S) TestAddInvalidHookRepository(c *check.C) {
	b := strings.NewReader(`{"repositories": ["some-repo"], "content": "some content"}`)
	recorder, request := post("/hook/invalid-hook", b, c)
	s.router.ServeHTTP(recorder, request)
	got := readBody(recorder.Body, c)
	expected := "Unsupported hook, valid options are: post-receive, pre-receive or update\n"
	c.Assert(got, check.Equals, expected)
	c.Assert(recorder.Code, check.Equals, 400)
}

func (s *S) TestAddPostReceiveHook(c *check.C) {
	b := strings.NewReader(`{"content": "some content"}`)
	recorder, request := post("/hook/post-receive", b, c)
	s.router.ServeHTTP(recorder, request)
	got := readBody(recorder.Body, c)
	expected := "hook post-receive successfully created\n"
	c.Assert(got, check.Equals, expected)
	c.Assert(recorder.Code, check.Equals, 200)
	file, err := fs.Filesystem().OpenFile("/home/git/bare-template/hooks/post-receive", os.O_RDONLY, 0755)
	defer file.Close()
	content, err := ioutil.ReadAll(file)
	c.Assert(err, check.IsNil)
	c.Assert(string(content), check.Equals, "some content")
}

func (s *S) TestAddPreReceiveHook(c *check.C) {
	b := strings.NewReader(`{"content": "some content"}`)
	recorder, request := post("/hook/pre-receive", b, c)
	s.router.ServeHTTP(recorder, request)
	got := readBody(recorder.Body, c)
	expected := "hook pre-receive successfully created\n"
	c.Assert(got, check.Equals, expected)
	c.Assert(recorder.Code, check.Equals, 200)
	file, err := fs.Filesystem().OpenFile("/home/git/bare-template/hooks/pre-receive", os.O_RDONLY, 0755)
	defer file.Close()
	content, err := ioutil.ReadAll(file)
	c.Assert(err, check.IsNil)
	c.Assert(string(content), check.Equals, "some content")
}

func (s *S) TestAddUpdateHook(c *check.C) {
	b := strings.NewReader(`{"content": "some content"}`)
	recorder, request := post("/hook/update", b, c)
	s.router.ServeHTTP(recorder, request)
	got := readBody(recorder.Body, c)
	expected := "hook update successfully created\n"
	c.Assert(got, check.Equals, expected)
	c.Assert(recorder.Code, check.Equals, 200)
	file, err := fs.Filesystem().OpenFile("/home/git/bare-template/hooks/update", os.O_RDONLY, 0755)
	defer file.Close()
	content, err := ioutil.ReadAll(file)
	c.Assert(err, check.IsNil)
	c.Assert(string(content), check.Equals, "some content")
}

func (s *S) TestAddInvalidHook(c *check.C) {
	b := strings.NewReader(`{"content": "some content"}`)
	recorder, request := post("/hook/invalid-hook", b, c)
	s.router.ServeHTTP(recorder, request)
	got := readBody(recorder.Body, c)
	expected := "Unsupported hook, valid options are: post-receive, pre-receive or update\n"
	c.Assert(got, check.Equals, expected)
	c.Assert(recorder.Code, check.Equals, 400)
}

func (s *S) TestAddPostReceiveOldFormatHook(c *check.C) {
	b := strings.NewReader("some content")
	recorder, request := post("/hook/post-receive", b, c)
	s.router.ServeHTTP(recorder, request)
	got := readBody(recorder.Body, c)
	expected := "hook post-receive successfully created\n"
	c.Assert(got, check.Equals, expected)
	c.Assert(recorder.Code, check.Equals, 200)
	file, err := fs.Filesystem().OpenFile("/home/git/bare-template/hooks/post-receive", os.O_RDONLY, 0755)
	defer file.Close()
	content, err := ioutil.ReadAll(file)
	c.Assert(err, check.IsNil)
	c.Assert(string(content), check.Equals, "some content")
}

func (s *S) TestAddPreReceiveOldFormatHook(c *check.C) {
	b := strings.NewReader("some content")
	recorder, request := post("/hook/pre-receive", b, c)
	s.router.ServeHTTP(recorder, request)
	got := readBody(recorder.Body, c)
	expected := "hook pre-receive successfully created\n"
	c.Assert(got, check.Equals, expected)
	c.Assert(recorder.Code, check.Equals, 200)
	file, err := fs.Filesystem().OpenFile("/home/git/bare-template/hooks/pre-receive", os.O_RDONLY, 0755)
	defer file.Close()
	content, err := ioutil.ReadAll(file)
	c.Assert(err, check.IsNil)
	c.Assert(string(content), check.Equals, "some content")
}

func (s *S) TestAddUpdateOldFormatHook(c *check.C) {
	b := strings.NewReader("some content")
	recorder, request := post("/hook/update", b, c)
	s.router.ServeHTTP(recorder, request)
	got := readBody(recorder.Body, c)
	expected := "hook update successfully created\n"
	c.Assert(got, check.Equals, expected)
	c.Assert(recorder.Code, check.Equals, 200)
	file, err := fs.Filesystem().OpenFile("/home/git/bare-template/hooks/update", os.O_RDONLY, 0755)
	defer file.Close()
	content, err := ioutil.ReadAll(file)
	c.Assert(err, check.IsNil)
	c.Assert(string(content), check.Equals, "some content")
}

func (s *S) TestAddInvalidOldFormatHook(c *check.C) {
	b := strings.NewReader("some content")
	recorder, request := post("/hook/invalid-hook", b, c)
	s.router.ServeHTTP(recorder, request)
	got := readBody(recorder.Body, c)
	expected := "Unsupported hook, valid options are: post-receive, pre-receive or update\n"
	c.Assert(got, check.Equals, expected)
	c.Assert(recorder.Code, check.Equals, 400)
}

func (s *S) TestAddKeyShouldReturnErrorWhenUserDoesNotExist(c *check.C) {
	b := strings.NewReader(`{"key": "a public key"}`)
	recorder, request := post("/user/Frodo/key", b, c)
	s.router.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusNotFound)
	body, err := ioutil.ReadAll(recorder.Body)
	c.Assert(err, check.IsNil)
	c.Assert(string(body), check.Equals, "user not found\n")
}

func (s *S) TestAddKeyShouldReturnProperStatusCodeWhenKeyAlreadyExists(c *check.C) {
	usr, err := user.New("Frodo", map[string]string{"keyname": rawKey})
	c.Assert(err, check.IsNil)
	defer user.Remove(usr.Name)
	b := strings.NewReader(fmt.Sprintf(`{"keyname": %q}`, rawKey))
	recorder, request := post(fmt.Sprintf("/user/%s/key", usr.Name), b, c)
	s.router.ServeHTTP(recorder, request)
	got := readBody(recorder.Body, c)
	expected := "Key already exists.\n"
	c.Assert(got, check.Equals, expected)
	c.Assert(recorder.Code, check.Equals, http.StatusConflict)
}

func (s *S) TestAddKeyShouldNotAcceptRepeatedKeysForDifferentUsers(c *check.C) {
	usr, err := user.New("Frodo", map[string]string{"keyname": rawKey})
	c.Assert(err, check.IsNil)
	defer user.Remove(usr.Name)
	usr2, err := user.New("tempo", nil)
	c.Assert(err, check.IsNil)
	defer user.Remove(usr2.Name)
	b := strings.NewReader(fmt.Sprintf(`{"keyname": %q}`, rawKey))
	recorder, request := post(fmt.Sprintf("/user/%s/key", usr2.Name), b, c)
	s.router.ServeHTTP(recorder, request)
	got := readBody(recorder.Body, c)
	expected := "Key already exists.\n"
	c.Assert(got, check.Equals, expected)
	c.Assert(recorder.Code, check.Equals, http.StatusConflict)
}

func (s *S) TestAddKeyInvalidKey(c *check.C) {
	u := user.User{Name: "Frodo"}
	conn, err := db.Conn()
	c.Assert(err, check.IsNil)
	defer conn.Close()
	err = conn.User().Insert(&u)
	c.Assert(err, check.IsNil)
	defer conn.User().Remove(bson.M{"_id": "Frodo"})
	b := strings.NewReader(`{"keyname":"invalid-rsa"}`)
	recorder, request := post(fmt.Sprintf("/user/%s/key", u.Name), b, c)
	s.router.ServeHTTP(recorder, request)
	got := readBody(recorder.Body, c)
	expected := "Invalid key\n"
	c.Assert(got, check.Equals, expected)
	c.Assert(recorder.Code, check.Equals, http.StatusBadRequest)
}

func (s *S) TestAddKeyShouldRequireKey(c *check.C) {
	u := user.User{Name: "Frodo"}
	conn, err := db.Conn()
	c.Assert(err, check.IsNil)
	defer conn.Close()
	err = conn.User().Insert(&u)
	c.Assert(err, check.IsNil)
	defer conn.User().Remove(bson.M{"_id": "Frodo"})
	b := strings.NewReader(`{}`)
	recorder, request := post("/user/Frodo/key", b, c)
	s.router.ServeHTTP(recorder, request)
	body := readBody(recorder.Body, c)
	expected := "A key is needed"
	got := strings.Replace(body, "\n", "", -1)
	c.Assert(got, check.Equals, expected)
}

func (s *S) TestAddKeyShouldWriteKeyInAuthorizedKeysFile(c *check.C) {
	u := user.User{Name: "Frodo"}
	conn, err := db.Conn()
	c.Assert(err, check.IsNil)
	defer conn.Close()
	err = conn.User().Insert(&u)
	c.Assert(err, check.IsNil)
	defer conn.User().RemoveId("Frodo")
	b := strings.NewReader(fmt.Sprintf(`{"key": "%s"}`, rawKey))
	recorder, request := post("/user/Frodo/key", b, c)
	s.router.ServeHTTP(recorder, request)
	defer conn.Key().Remove(bson.M{"name": "key", "username": u.Name})
	c.Assert(recorder.Code, check.Equals, 200)
	content := s.authKeysContent(c)
	c.Assert(strings.HasSuffix(strings.TrimSpace(content), rawKey), check.Equals, true)
}

func (s *S) TestRemoveKeyGivesExpectedSuccessResponse(c *check.C) {
	u, err := user.New("Gandalf", map[string]string{"keyname": rawKey})
	c.Assert(err, check.IsNil)
	defer user.Remove(u.Name)
	url := "/user/Gandalf/key/keyname"
	recorder, request := del(url, nil, c)
	s.router.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, 200)
	b := readBody(recorder.Body, c)
	c.Assert(b, check.Equals, `Key "keyname" successfully removed`)
}

func (s *S) TestRemoveKeyRemovesKeyFromDatabase(c *check.C) {
	u, err := user.New("Gandalf", map[string]string{"keyname": rawKey})
	c.Assert(err, check.IsNil)
	defer user.Remove(u.Name)
	url := "/user/Gandalf/key/keyname"
	recorder, request := del(url, nil, c)
	s.router.ServeHTTP(recorder, request)
	conn, err := db.Conn()
	c.Assert(err, check.IsNil)
	defer conn.Close()
	count, err := conn.Key().Find(bson.M{"name": "keyname", "username": "Gandalf"}).Count()
	c.Assert(err, check.IsNil)
	c.Assert(count, check.Equals, 0)
}

func (s *S) TestRemoveKeyShouldRemoveKeyFromAuthorizedKeysFile(c *check.C) {
	u, err := user.New("Gandalf", map[string]string{"keyname": rawKey})
	c.Assert(err, check.IsNil)
	defer user.Remove(u.Name)
	url := "/user/Gandalf/key/keyname"
	recorder, request := del(url, nil, c)
	s.router.ServeHTTP(recorder, request)
	content := s.authKeysContent(c)
	c.Assert(content, check.Equals, "")
}

func (s *S) TestRemoveKeyShouldReturnErrorWithLineBreakAtEnd(c *check.C) {
	url := "/user/idiocracy/key/keyname"
	recorder, request := del(url, nil, c)
	s.router.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusNotFound)
	b := readBody(recorder.Body, c)
	c.Assert(b, check.Equals, "user not found\n")
}

func (s *S) TestListKeysGivesExpectedSuccessResponse(c *check.C) {
	keys := map[string]string{"key1": rawKey, "key2": otherKey}
	u, err := user.New("Gandalf", keys)
	c.Assert(err, check.IsNil)
	defer user.Remove(u.Name)
	url := "/user/Gandalf/keys"
	request, err := http.NewRequest("GET", url, nil)
	c.Assert(err, check.IsNil)
	recorder := httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, 200)
	body, err := ioutil.ReadAll(recorder.Body)
	c.Assert(err, check.IsNil)
	var data map[string]string
	err = json.Unmarshal(body, &data)
	c.Assert(err, check.IsNil)
	c.Assert(data, check.DeepEquals, keys)
}

func (s *S) TestListKeysWithoutKeysGivesEmptyJSON(c *check.C) {
	u, err := user.New("Gandalf", map[string]string{})
	c.Assert(err, check.IsNil)
	defer user.Remove(u.Name)
	url := "/user/Gandalf/keys"
	request, err := http.NewRequest("GET", url, nil)
	c.Assert(err, check.IsNil)
	recorder := httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, 200)
	b := readBody(recorder.Body, c)
	c.Assert(b, check.Equals, "{}")
}

func (s *S) TestListKeysWithInvalidUserReturnsNotFound(c *check.C) {
	url := "/user/no-Gandalf/keys"
	request, err := http.NewRequest("GET", url, nil)
	c.Assert(err, check.IsNil)
	recorder := httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusNotFound)
	b := readBody(recorder.Body, c)
	c.Assert(b, check.Equals, "user not found\n")
}

func (s *S) TestRemoveUser(c *check.C) {
	u, err := user.New("username", map[string]string{})
	c.Assert(err, check.IsNil)
	url := fmt.Sprintf("/user/%s", u.Name)
	request, err := http.NewRequest("DELETE", url, nil)
	c.Assert(err, check.IsNil)
	recorder := httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, 200)
	b, err := ioutil.ReadAll(recorder.Body)
	c.Assert(err, check.IsNil)
	c.Assert(string(b), check.Equals, "User \"username\" successfully removed\n")
}

func (s *S) TestRemoveUserShouldRemoveFromDB(c *check.C) {
	u, err := user.New("anuser", map[string]string{})
	c.Assert(err, check.IsNil)
	url := fmt.Sprintf("/user/%s", u.Name)
	request, err := http.NewRequest("DELETE", url, nil)
	c.Assert(err, check.IsNil)
	recorder := httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	conn, err := db.Conn()
	c.Assert(err, check.IsNil)
	defer conn.Close()
	collection := conn.User()
	lenght, err := collection.Find(bson.M{"_id": u.Name}).Count()
	c.Assert(err, check.IsNil)
	c.Assert(lenght, check.Equals, 0)
}

func (s *S) TestRemoveUserNotFound(c *check.C) {
	request, err := http.NewRequest("DELETE", "/user/unknown-user", nil)
	c.Assert(err, check.IsNil)
	recorder := httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusNotFound)
	c.Assert(recorder.Body.String(), check.Equals, "user not found\n")
}

func (s *S) TestRemoveRepository(c *check.C) {
	r, err := repository.New("myRepo", []string{"pippin"}, []string{""}, true)
	c.Assert(err, check.IsNil)
	url := fmt.Sprintf("/repository/%s", r.Name)
	request, err := http.NewRequest("DELETE", url, nil)
	c.Assert(err, check.IsNil)
	recorder := httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, 200)
	b, err := ioutil.ReadAll(recorder.Body)
	c.Assert(err, check.IsNil)
	c.Assert(string(b), check.Equals, "Repository \"myRepo\" successfully removed\n")
}

func (s *S) TestRemoveRepositoryShouldRemoveFromDB(c *check.C) {
	r, err := repository.New("myRepo", []string{"pippin"}, []string{""}, true)
	c.Assert(err, check.IsNil)
	url := fmt.Sprintf("/repository/%s", r.Name)
	request, err := http.NewRequest("DELETE", url, nil)
	c.Assert(err, check.IsNil)
	recorder := httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	conn, err := db.Conn()
	c.Assert(err, check.IsNil)
	defer conn.Close()
	err = conn.Repository().Find(bson.M{"_id": r.Name}).One(&r)
	c.Assert(err, check.Equals, mgo.ErrNotFound)
}

func (s *S) TestRemoveRepositoryNotFound(c *check.C) {
	url := "/repository/foo"
	request, err := http.NewRequest("DELETE", url, nil)
	c.Assert(err, check.IsNil)
	recorder := httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusNotFound)
	b, err := ioutil.ReadAll(recorder.Body)
	c.Assert(err, check.IsNil)
	c.Assert(string(b), check.Equals, "repository not found\n")
}

func (s *S) TestUpdateRepositoryShouldReturnErrorWhenBodyIsEmpty(c *check.C) {
	r, err := repository.New("something", []string{"guardian@what.com"}, []string{""}, true)
	c.Assert(err, check.IsNil)
	conn, err := db.Conn()
	c.Assert(err, check.IsNil)
	defer conn.Close()
	defer conn.Repository().RemoveId(r.Name)
	b := strings.NewReader("")
	recorder, request := put("/repository/something", b, c)
	s.router.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, 400)
}

func (s *S) TestUpdateRepositoryData(c *check.C) {
	r, err := repository.New("something", []string{"guardian@what.com"}, []string{""}, true)
	c.Assert(err, check.IsNil)
	conn, err := db.Conn()
	c.Assert(err, check.IsNil)
	defer conn.Close()
	defer conn.Repository().RemoveId(r.Name)
	url := fmt.Sprintf("/repository/%s", r.Name)
	body := strings.NewReader(`{"users": ["b"], "readonlyusers": ["a"], "ispublic": false}`)
	request, err := http.NewRequest("PUT", url, body)
	c.Assert(err, check.IsNil)
	recorder := httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusOK)
	r.Users = []string{"b"}
	r.ReadOnlyUsers = []string{"a"}
	r.IsPublic = false
	repo, err := repository.Get("something")
	c.Assert(err, check.IsNil)
	c.Assert(repo, check.DeepEquals, *r)
}

func (s *S) TestUpdateRepositoryDataPartial(c *check.C) {
	r, err := repository.New("something", []string{"pippin"}, []string{"merry"}, true)
	c.Assert(err, check.IsNil)
	conn, err := db.Conn()
	c.Assert(err, check.IsNil)
	defer conn.Close()
	defer conn.Repository().RemoveId(r.Name)
	url := fmt.Sprintf("/repository/%s", r.Name)
	body := strings.NewReader(`{"readonlyusers": ["a", "b"]}`)
	request, err := http.NewRequest("PUT", url, body)
	c.Assert(err, check.IsNil)
	recorder := httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusOK)
	r.Users = []string{"pippin"}
	r.ReadOnlyUsers = []string{"a", "b"}
	r.IsPublic = true
	repo, err := repository.Get("something")
	c.Assert(err, check.IsNil)
	c.Assert(repo, check.DeepEquals, *r)
}

func (s *S) TestUpdateRepositoryNotFound(c *check.C) {
	url := "/repository/foo"
	body := strings.NewReader(`{"ispublic":true}`)
	request, err := http.NewRequest("PUT", url, body)
	c.Assert(err, check.IsNil)
	recorder := httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusNotFound)
}

func (s *S) TestUpdateRepositoryInvalidJSON(c *check.C) {
	r, err := repository.New("bar", []string{"guardian@what.com"}, []string{""}, true)
	c.Assert(err, check.IsNil)
	conn, err := db.Conn()
	c.Assert(err, check.IsNil)
	defer conn.Close()
	defer conn.Repository().RemoveId(r.Name)
	url := "/repository/bar"
	body := strings.NewReader(`{"name""`)
	request, err := http.NewRequest("PUT", url, body)
	c.Assert(err, check.IsNil)
	recorder := httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusBadRequest)
}

func (s *S) TestRenameRepositoryWithNamespace(c *check.C) {
	r, err := repository.New("lift/raising", []string{"guardian@what.com"}, []string{}, true)
	c.Assert(err, check.IsNil)
	conn, err := db.Conn()
	c.Assert(err, check.IsNil)
	defer conn.Close()
	defer conn.Repository().RemoveId(r.Name)
	url := fmt.Sprintf("/repository/%s/", r.Name)
	body := strings.NewReader(`{"name":"norestraint/freedom"}`)
	request, err := http.NewRequest("PUT", url, body)
	c.Assert(err, check.IsNil)
	recorder := httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusOK)
	_, err = repository.Get("raising")
	c.Assert(err, check.NotNil)
	r.Name = "norestraint/freedom"
	repo, err := repository.Get("norestraint/freedom")
	c.Assert(err, check.IsNil)
	c.Assert(repo, check.DeepEquals, *r)
}

func (s *S) TestRenameRepositoryInvalidJSON(c *check.C) {
	r, err := repository.New("foo", []string{"guardian@what.com"}, []string{}, true)
	conn, err := db.Conn()
	c.Assert(err, check.IsNil)
	defer conn.Close()
	defer conn.Repository().RemoveId(r.Name)
	c.Assert(err, check.IsNil)
	url := "/repository/foo"
	body := strings.NewReader(`{"name""`)
	request, err := http.NewRequest("PUT", url, body)
	c.Assert(err, check.IsNil)
	recorder := httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusBadRequest)
}

func (s *S) TestRenameRepositoryNotfound(c *check.C) {
	url := "/repository/foo"
	body := strings.NewReader(`{"name":"freedom"}`)
	request, err := http.NewRequest("PUT", url, body)
	c.Assert(err, check.IsNil)
	recorder := httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusNotFound)
}

func (s *S) TestHealthcheck(c *check.C) {
	url := "/healthcheck"
	request, err := http.NewRequest("GET", url, nil)
	c.Assert(err, check.IsNil)
	recorder := httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusOK)
	c.Assert(recorder.Body.String(), check.Equals, "WORKING")
}

func (s *S) TestGetFileContents(c *check.C) {
	url := "/repository/repo/contents?path=README.txt"
	expected := "result"
	repository.Retriever = &repository.MockContentRetriever{
		ResultContents: []byte(expected),
	}
	defer func() {
		repository.Retriever = nil
	}()
	request, err := http.NewRequest("GET", url, nil)
	c.Assert(err, check.IsNil)
	recorder := httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusOK)
	c.Assert(recorder.Body.String(), check.Equals, expected)
	c.Assert(recorder.Header()["Content-Type"][0], check.Equals, "text/plain; charset=utf-8")
	c.Assert(recorder.Header()["Content-Length"][0], check.Equals, "6")
}

func (s *S) TestGetFileContentsWithoutExtension(c *check.C) {
	url := "/repository/repo/contents?path=README"
	expected := "result"
	repository.Retriever = &repository.MockContentRetriever{
		ResultContents: []byte(expected),
	}
	defer func() {
		repository.Retriever = nil
	}()
	request, err := http.NewRequest("GET", url, nil)
	c.Assert(err, check.IsNil)
	recorder := httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusOK)
	c.Assert(recorder.Body.String(), check.Equals, expected)
	c.Assert(recorder.Header()["Content-Type"][0], check.Equals, "text/plain; charset=utf-8")
	c.Assert(recorder.Header()["Content-Length"][0], check.Equals, "6")
}

func (s *S) TestGetBinaryFileContentsWithoutExtension(c *check.C) {
	url := "/repository/repo/contents?path=my-binary-file"
	expected := new(bytes.Buffer)
	expected.Write([]byte{10, 20, 30, 0, 9, 200})
	repository.Retriever = &repository.MockContentRetriever{
		ResultContents: expected.Bytes(),
	}
	defer func() {
		repository.Retriever = nil
	}()
	request, err := http.NewRequest("GET", url, nil)
	c.Assert(err, check.IsNil)
	recorder := httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusOK)
	c.Assert(recorder.Body, check.DeepEquals, expected)
	c.Assert(recorder.Header()["Content-Type"][0], check.Equals, "application/octet-stream")
	c.Assert(recorder.Header()["Content-Length"][0], check.Equals, "6")
}

func (s *S) TestGetFileContentsWithRef(c *check.C) {
	url := "/repository/repo/contents?path=README.txt&ref=other"
	expected := "result"
	mockRetriever := repository.MockContentRetriever{
		ResultContents: []byte(expected),
	}
	repository.Retriever = &mockRetriever
	defer func() {
		repository.Retriever = nil
	}()
	request, err := http.NewRequest("GET", url, nil)
	c.Assert(err, check.IsNil)
	recorder := httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusOK)
	c.Assert(recorder.Body.String(), check.Equals, expected)
	c.Assert(recorder.Header()["Content-Type"][0], check.Equals, "text/plain; charset=utf-8")
	c.Assert(recorder.Header()["Content-Length"][0], check.Equals, "6")
	c.Assert(mockRetriever.LastRef, check.Equals, "other")
}

func (s *S) TestGetFileContentsWhenCommandFails(c *check.C) {
	url := "/repository/repo/contents?path=README.txt&ref=other"
	outputError := fmt.Errorf("command error")
	repository.Retriever = &repository.MockContentRetriever{
		OutputError: outputError,
	}
	defer func() {
		repository.Retriever = nil
	}()
	request, err := http.NewRequest("GET", url, nil)
	c.Assert(err, check.IsNil)
	recorder := httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusNotFound)
	c.Assert(recorder.Body.String(), check.Equals, "command error\n")
}

func (s *S) TestGetFileContentsWhenNoPath(c *check.C) {
	url := "/repository/repo/contents?&ref=other"
	request, err := http.NewRequest("GET", url, nil)
	c.Assert(err, check.IsNil)
	recorder := httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusBadRequest)
	expected := "Error when trying to obtain an uknown file on ref other of repository repo (path is required).\n"
	c.Assert(recorder.Body.String(), check.Equals, expected)
}

func (s *S) TestGetArchiveWhenNoRef(c *check.C) {
	url := "/repository/repo/archive?ref=&format=zip"
	request, err := http.NewRequest("GET", url, nil)
	c.Assert(err, check.IsNil)
	recorder := httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusBadRequest)
	expected := "Error when trying to obtain archive for ref '' (format: zip) of repository 'repo' (ref and format are required).\n"
	c.Assert(recorder.Body.String(), check.Equals, expected)
}

func (s *S) TestGetArchiveWhenNoFormat(c *check.C) {
	url := "/repository/repo/archive?ref=master&format="
	request, err := http.NewRequest("GET", url, nil)
	c.Assert(err, check.IsNil)
	recorder := httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusBadRequest)
	expected := "Error when trying to obtain archive for ref 'master' (format: ) of repository 'repo' (ref and format are required).\n"
	c.Assert(recorder.Body.String(), check.Equals, expected)
}

func (s *S) TestGetArchiveWhenCommandFails(c *check.C) {
	url := "/repository/repo/archive?ref=master&format=zip"
	expected := fmt.Errorf("output error")
	mockRetriever := repository.MockContentRetriever{
		OutputError: expected,
	}
	repository.Retriever = &mockRetriever
	defer func() {
		repository.Retriever = nil
	}()
	request, err := http.NewRequest("GET", url, nil)
	c.Assert(err, check.IsNil)
	recorder := httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusNotFound)
	c.Assert(recorder.Body.String(), check.Equals, "output error\n")
}

func (s *S) TestGetArchive(c *check.C) {
	url := "/repository/repo/archive?ref=master&format=zip"
	expected := "result123"
	mockRetriever := repository.MockContentRetriever{
		ResultContents: []byte(expected),
	}
	repository.Retriever = &mockRetriever
	defer func() {
		repository.Retriever = nil
	}()
	request, err := http.NewRequest("GET", url, nil)
	c.Assert(err, check.IsNil)
	recorder := httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusOK)
	c.Assert(recorder.Body.String(), check.Equals, expected)
	c.Assert(mockRetriever.LastFormat, check.Equals, repository.Zip)
	c.Assert(recorder.Header()["Content-Type"][0], check.Equals, "application/octet-stream")
	c.Assert(recorder.Header()["Content-Disposition"][0], check.Equals, "attachment; filename=\"repo_master.zip\"")
	c.Assert(recorder.Header()["Content-Transfer-Encoding"][0], check.Equals, "binary")
	c.Assert(recorder.Header()["Accept-Ranges"][0], check.Equals, "bytes")
	c.Assert(recorder.Header()["Content-Length"][0], check.Equals, "9")
	c.Assert(recorder.Header()["Cache-Control"][0], check.Equals, "private")
	c.Assert(recorder.Header()["Pragma"][0], check.Equals, "private")
	c.Assert(recorder.Header()["Expires"][0], check.Equals, "Mon, 26 Jul 1997 05:00:00 GMT")
}

func (s *S) TestGetTreeWithDefaultValues(c *check.C) {
	url := "/repository/repo/tree"
	tree := make([]map[string]string, 1)
	tree[0] = make(map[string]string)
	tree[0]["permission"] = "333"
	tree[0]["filetype"] = "blob"
	tree[0]["hash"] = "123456"
	tree[0]["path"] = "filename.txt"
	tree[0]["rawPath"] = "raw/filename.txt"
	mockRetriever := repository.MockContentRetriever{
		Tree: tree,
	}
	repository.Retriever = &mockRetriever
	defer func() {
		repository.Retriever = nil
	}()
	request, err := http.NewRequest("GET", url, nil)
	c.Assert(err, check.IsNil)
	recorder := httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusOK)
	var obj []map[string]string
	json.Unmarshal(recorder.Body.Bytes(), &obj)
	c.Assert(len(obj), check.Equals, 1)
	c.Assert(obj[0]["permission"], check.Equals, tree[0]["permission"])
	c.Assert(obj[0]["filetype"], check.Equals, tree[0]["filetype"])
	c.Assert(obj[0]["hash"], check.Equals, tree[0]["hash"])
	c.Assert(obj[0]["path"], check.Equals, tree[0]["path"])
	c.Assert(obj[0]["rawPath"], check.Equals, tree[0]["rawPath"])
	c.Assert(mockRetriever.LastRef, check.Equals, "master")
	c.Assert(mockRetriever.LastPath, check.Equals, ".")
}

func (s *S) TestGetTreeWithSpecificPath(c *check.C) {
	url := "/repository/repo/tree?path=/test"
	tree := make([]map[string]string, 1)
	tree[0] = make(map[string]string)
	tree[0]["permission"] = "333"
	tree[0]["filetype"] = "blob"
	tree[0]["hash"] = "123456"
	tree[0]["path"] = "/test/filename.txt"
	tree[0]["rawPath"] = "/test/raw/filename.txt"
	mockRetriever := repository.MockContentRetriever{
		Tree: tree,
	}
	repository.Retriever = &mockRetriever
	defer func() {
		repository.Retriever = nil
	}()
	request, err := http.NewRequest("GET", url, nil)
	c.Assert(err, check.IsNil)
	recorder := httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusOK)
	var obj []map[string]string
	json.Unmarshal(recorder.Body.Bytes(), &obj)
	c.Assert(len(obj), check.Equals, 1)
	c.Assert(obj[0]["permission"], check.Equals, tree[0]["permission"])
	c.Assert(obj[0]["filetype"], check.Equals, tree[0]["filetype"])
	c.Assert(obj[0]["hash"], check.Equals, tree[0]["hash"])
	c.Assert(obj[0]["path"], check.Equals, tree[0]["path"])
	c.Assert(obj[0]["rawPath"], check.Equals, tree[0]["rawPath"])
	c.Assert(mockRetriever.LastRef, check.Equals, "master")
	c.Assert(mockRetriever.LastPath, check.Equals, "/test")
}

func (s *S) TestGetTreeWithSpecificRef(c *check.C) {
	url := "/repository/repo/tree?path=/test&ref=1.1.1"
	tree := make([]map[string]string, 1)
	tree[0] = make(map[string]string)
	tree[0]["permission"] = "333"
	tree[0]["filetype"] = "blob"
	tree[0]["hash"] = "123456"
	tree[0]["path"] = "/test/filename.txt"
	tree[0]["rawPath"] = "/test/raw/filename.txt"
	mockRetriever := repository.MockContentRetriever{
		Tree: tree,
	}
	repository.Retriever = &mockRetriever
	defer func() {
		repository.Retriever = nil
	}()
	request, err := http.NewRequest("GET", url, nil)
	c.Assert(err, check.IsNil)
	recorder := httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusOK)
	var obj []map[string]string
	json.Unmarshal(recorder.Body.Bytes(), &obj)
	c.Assert(len(obj), check.Equals, 1)
	c.Assert(obj[0]["permission"], check.Equals, tree[0]["permission"])
	c.Assert(obj[0]["filetype"], check.Equals, tree[0]["filetype"])
	c.Assert(obj[0]["hash"], check.Equals, tree[0]["hash"])
	c.Assert(obj[0]["path"], check.Equals, tree[0]["path"])
	c.Assert(obj[0]["rawPath"], check.Equals, tree[0]["rawPath"])
	c.Assert(mockRetriever.LastRef, check.Equals, "1.1.1")
	c.Assert(mockRetriever.LastPath, check.Equals, "/test")
}

func (s *S) TestGetTreeWhenCommandFails(c *check.C) {
	url := "/repository/repo/tree/?ref=master&path=/test"
	expected := fmt.Errorf("output error")
	mockRetriever := repository.MockContentRetriever{
		OutputError: expected,
	}
	repository.Retriever = &mockRetriever
	defer func() {
		repository.Retriever = nil
	}()
	request, err := http.NewRequest("GET", url, nil)
	c.Assert(err, check.IsNil)
	recorder := httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusBadRequest)
	c.Assert(recorder.Body.String(), check.Equals, "Error when trying to obtain tree for path /test on ref master of repository repo (output error).\n")
}

func (s *S) TestGetBranches(c *check.C) {
	url := "/repository/repo/branches"
	refs := make([]repository.Ref, 1)
	refs[0] = repository.Ref{
		Ref:       "a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9",
		Name:      "doge_barks",
		CreatedAt: "Mon Jul 28 10:13:27 2014 -0300",
		Committer: &repository.GitUser{
			Name:  "doge",
			Email: "<much@email.com>",
		},
		Author: &repository.GitUser{
			Name:  "doge",
			Email: "<much@email.com>",
		},
		Subject: "will bark",
		Links: &repository.Links{
			ZipArchive: repository.GetArchiveUrl("repo", "doge_barks", "zip"),
			TarArchive: repository.GetArchiveUrl("repo", "doge_barks", "tar.gz"),
		},
	}
	mockRetriever := repository.MockContentRetriever{
		Refs: refs,
	}
	repository.Retriever = &mockRetriever
	defer func() {
		repository.Retriever = nil
	}()
	request, err := http.NewRequest("GET", url, nil)
	c.Assert(err, check.IsNil)
	recorder := httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusOK)
	var obj []repository.Ref
	json.Unmarshal(recorder.Body.Bytes(), &obj)
	c.Assert(obj, check.HasLen, 1)
	c.Assert(obj[0], check.DeepEquals, refs[0])
}

func (s *S) TestGetBranchesWhenRepoNonExistent(c *check.C) {
	url := "/repository/repo/branches"
	request, err := http.NewRequest("GET", url, nil)
	c.Assert(err, check.IsNil)
	recorder := httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusBadRequest)
	expected := "Error when trying to obtain the branches of repository repo (Error when trying to obtain the refs of repository repo (Repository does not exist).).\n"
	c.Assert(recorder.Body.String(), check.Equals, expected)
}

func (s *S) TestGetBranchesWhenCommandFails(c *check.C) {
	url := "/repository/repo/branches"
	expected := fmt.Errorf("output error")
	mockRetriever := repository.MockContentRetriever{
		OutputError: expected,
	}
	repository.Retriever = &mockRetriever
	defer func() {
		repository.Retriever = nil
	}()
	request, err := http.NewRequest("GET", url, nil)
	c.Assert(err, check.IsNil)
	recorder := httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusBadRequest)
	c.Assert(recorder.Body.String(), check.Equals, "Error when trying to obtain the branches of repository repo (output error).\n")
}

func (s *S) TestGetTags(c *check.C) {
	url := "/repository/repo/tags"
	refs := make([]repository.Ref, 1)
	refs[0] = repository.Ref{
		Ref:       "a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9",
		Name:      "doge_barks",
		CreatedAt: "Mon Jul 28 10:13:27 2014 -0300",
		Committer: &repository.GitUser{
			Name:  "doge",
			Email: "<much@email.com>",
		},
		Author: &repository.GitUser{
			Name:  "doge",
			Email: "<much@email.com>",
		},
		Subject: "will bark",
		Links: &repository.Links{
			ZipArchive: repository.GetArchiveUrl("repo", "doge_barks", "zip"),
			TarArchive: repository.GetArchiveUrl("repo", "doge_barks", "tar.gz"),
		},
	}
	mockRetriever := repository.MockContentRetriever{
		Refs: refs,
	}
	repository.Retriever = &mockRetriever
	defer func() {
		repository.Retriever = nil
	}()
	request, err := http.NewRequest("GET", url, nil)
	c.Assert(err, check.IsNil)
	recorder := httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusOK)
	var obj []repository.Ref
	json.Unmarshal(recorder.Body.Bytes(), &obj)
	c.Assert(obj, check.HasLen, 1)
	c.Assert(obj[0], check.DeepEquals, refs[0])
}

func (s *S) TestGetDiff(c *check.C) {
	url := "/repository/repo/diff/commits?previous_commit=1b970b076bbb30d708e262b402d4e31910e1dc10&last_commit=545b1904af34458704e2aa06ff1aaffad5289f8f"
	expected := "test_diff"
	repository.Retriever = &repository.MockContentRetriever{
		ResultContents: []byte(expected),
	}
	defer func() {
		repository.Retriever = nil
	}()
	request, err := http.NewRequest("GET", url, nil)
	c.Assert(err, check.IsNil)
	recorder := httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusOK)
	c.Assert(recorder.Body.String(), check.Equals, expected)
}

func (s *S) TestGetDiffWhenCommandFails(c *check.C) {
	url := "/repository/repo/diff/commits?previous_commit=1b970b076bbb30d708e262b402d4e31910e1dc10&last_commit=545b1904af34458704e2aa06ff1aaffad5289f8f"
	outputError := fmt.Errorf("command error")
	repository.Retriever = &repository.MockContentRetriever{
		OutputError: outputError,
	}
	defer func() {
		repository.Retriever = nil
	}()
	request, err := http.NewRequest("GET", url, nil)
	c.Assert(err, check.IsNil)
	recorder := httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusNotFound)
	c.Assert(recorder.Body.String(), check.Equals, "command error\n")
}

func (s *S) TestGetDiffWhenNoCommits(c *check.C) {
	url := "/repository/repo/diff/commits?previous_commit=&last_commit="
	request, err := http.NewRequest("GET", url, nil)
	c.Assert(err, check.IsNil)
	recorder := httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	expected := "Error when trying to obtain diff between hash commits of repository repo (Hash Commit(s) are required).\n"
	c.Assert(recorder.Code, check.Equals, http.StatusBadRequest)
	c.Assert(recorder.Body.String(), check.Equals, expected)
}

func (s *S) TestPostNewCommit(c *check.C) {
	url := "/repository/repo/commit"
	params := map[string]string{
		"message":         "Repository scaffold",
		"author-name":     "Doge Dog",
		"author-email":    "doge@much.com",
		"committer-name":  "Doge Dog",
		"committer-email": "doge@much.com",
		"branch":          "master",
	}
	var files = []multipartzip.File{
		{"doge.txt", "Much doge"},
		{"much.txt", "Much mucho"},
		{"WOW/WOW.WOW", "WOW\nWOW"},
	}
	buf, err := multipartzip.CreateZipBuffer(files)
	c.Assert(err, check.IsNil)
	reader, writer := io.Pipe()
	go multipartzip.StreamWriteMultipartForm(params, "zipfile", "scaffold.zip", "muchBOUNDARY", writer, buf)
	mockRetriever := repository.MockContentRetriever{
		Ref: repository.Ref{
			Ref:       "some-random-ref",
			Name:      "master",
			CreatedAt: "Mon Jul 28 10:13:27 2014 -0300",
			Author: &repository.GitUser{
				Name:  params["author-name"],
				Email: params["author-email"],
			},
			Committer: &repository.GitUser{
				Name:  params["committer-name"],
				Email: params["committer-email"],
			},
			Subject: params["message"],
			Links: &repository.Links{
				ZipArchive: repository.GetArchiveUrl("repo", "master", "zip"),
				TarArchive: repository.GetArchiveUrl("repo", "master", "tar.gz"),
			},
		},
	}
	repository.Retriever = &mockRetriever
	defer func() {
		repository.Retriever = nil
	}()
	request, err := http.NewRequest("POST", url, reader)
	request.Header.Set("Content-Type", "multipart/form-data;boundary=muchBOUNDARY")
	c.Assert(err, check.IsNil)
	recorder := httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusOK)
	var data map[string]interface{}
	body, err := ioutil.ReadAll(recorder.Body)
	err = json.Unmarshal(body, &data)
	c.Assert(err, check.IsNil)
	expected := map[string]interface{}{
		"ref":  "some-random-ref",
		"name": "master",
		"author": map[string]interface{}{
			"name":  "Doge Dog",
			"email": "doge@much.com",
			"date":  "",
		},
		"committer": map[string]interface{}{
			"name":  "Doge Dog",
			"email": "doge@much.com",
			"date":  "",
		},
		"tagger": nil,
		"_links": map[string]interface{}{
			"tarArchive": "/repository/repo/archive?ref=master\u0026format=tar.gz",
			"zipArchive": "/repository/repo/archive?ref=master\u0026format=zip",
		},
		"subject":   "Repository scaffold",
		"createdAt": "Mon Jul 28 10:13:27 2014 -0300",
	}
	c.Assert(data, check.DeepEquals, expected)
}

func (s *S) TestPostNewCommitWithoutBranch(c *check.C) {
	url := "/repository/repo/commit"
	params := map[string]string{
		"message":         "Repository scaffold",
		"author-name":     "Doge Dog",
		"author-email":    "doge@much.com",
		"committer-name":  "Doge Dog",
		"committer-email": "doge@much.com",
	}
	var files = []multipartzip.File{
		{"doge.txt", "Much doge"},
		{"much.txt", "Much mucho"},
		{"WOW/WOW.WOW", "WOW\nWOW"},
	}
	buf, err := multipartzip.CreateZipBuffer(files)
	c.Assert(err, check.IsNil)
	reader, writer := io.Pipe()
	go multipartzip.StreamWriteMultipartForm(params, "zipfile", "scaffold.zip", "muchBOUNDARY", writer, buf)
	repository.Retriever = &repository.MockContentRetriever{}
	defer func() {
		repository.Retriever = nil
	}()
	request, err := http.NewRequest("POST", url, reader)
	request.Header.Set("Content-Type", "multipart/form-data;boundary=muchBOUNDARY")
	c.Assert(err, check.IsNil)
	recorder := httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusBadRequest)
}

func (s *S) TestPostNewCommitWithEmptyBranch(c *check.C) {
	url := "/repository/repo/commit"
	params := map[string]string{
		"message":         "Repository scaffold",
		"author-name":     "Doge Dog",
		"author-email":    "doge@much.com",
		"committer-name":  "Doge Dog",
		"committer-email": "doge@much.com",
		"branch":          "",
	}
	var files = []multipartzip.File{
		{"doge.txt", "Much doge"},
		{"much.txt", "Much mucho"},
		{"WOW/WOW.WOW", "WOW\nWOW"},
	}
	buf, err := multipartzip.CreateZipBuffer(files)
	c.Assert(err, check.IsNil)
	reader, writer := io.Pipe()
	go multipartzip.StreamWriteMultipartForm(params, "zipfile", "scaffold.zip", "muchBOUNDARY", writer, buf)
	repository.Retriever = &repository.MockContentRetriever{}
	defer func() {
		repository.Retriever = nil
	}()
	request, err := http.NewRequest("POST", url, reader)
	request.Header.Set("Content-Type", "multipart/form-data;boundary=muchBOUNDARY")
	c.Assert(err, check.IsNil)
	recorder := httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusBadRequest)
}

func (s *S) TestLogs(c *check.C) {
	url := "/repository/repo/logs?ref=HEAD&total=1"
	objects := repository.GitHistory{}
	parent := make([]string, 2)
	parent[0] = "a367b5de5943632e47cb6f8bf5b2147bc0be5cf8"
	parent[1] = "b267b5de5943632e47cb6f8bf5b2147bc0be5cf2"
	commits := make([]repository.GitLog, 1)
	commits[0] = repository.GitLog{
		Ref:       "a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9",
		CreatedAt: "Mon Jul 28 10:13:27 2014 -0300",
		Committer: &repository.GitUser{
			Name:  "doge",
			Email: "much@email.com",
		},
		Author: &repository.GitUser{
			Name:  "doge",
			Email: "much@email.com",
		},
		Subject: "will bark",
		Parent:  parent,
	}
	objects.Commits = commits
	objects.Next = "b231c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9"
	mockRetriever := repository.MockContentRetriever{
		History: objects,
	}
	repository.Retriever = &mockRetriever
	defer func() {
		repository.Retriever = nil
	}()
	request, err := http.NewRequest("GET", url, nil)
	c.Assert(err, check.IsNil)
	recorder := httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusOK)
	var obj repository.GitHistory
	json.Unmarshal(recorder.Body.Bytes(), &obj)
	c.Assert(obj.Next, check.Equals, "b231c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9")
	c.Assert(obj.Commits, check.HasLen, 1)
	c.Assert(obj.Commits[0], check.DeepEquals, commits[0])
}

func (s *S) TestLogsWithPath(c *check.C) {
	url := "/repository/repo/logs?ref=HEAD&total=1&path=README.txt"
	objects := repository.GitHistory{}
	parent := make([]string, 2)
	parent[0] = "a367b5de5943632e47cb6f8bf5b2147bc0be5cf8"
	parent[1] = "b267b5de5943632e47cb6f8bf5b2147bc0be5cf2"
	commits := make([]repository.GitLog, 1)
	commits[0] = repository.GitLog{
		Ref:       "a0b1c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9",
		CreatedAt: "Mon Jul 28 10:13:27 2014 -0300",
		Committer: &repository.GitUser{
			Name:  "doge",
			Email: "much@email.com",
		},
		Author: &repository.GitUser{
			Name:  "doge",
			Email: "much@email.com",
		},
		Subject: "will bark",
		Parent:  parent,
	}
	objects.Commits = commits
	objects.Next = "b231c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9"
	mockRetriever := repository.MockContentRetriever{
		History: objects,
	}
	repository.Retriever = &mockRetriever
	defer func() {
		repository.Retriever = nil
	}()
	request, err := http.NewRequest("GET", url, nil)
	c.Assert(err, check.IsNil)
	recorder := httptest.NewRecorder()
	s.router.ServeHTTP(recorder, request)
	c.Assert(recorder.Code, check.Equals, http.StatusOK)
	var obj repository.GitHistory
	json.Unmarshal(recorder.Body.Bytes(), &obj)
	c.Assert(obj.Next, check.Equals, "b231c2d3e4f5a6b7c8d9e0f1a2b3c4d5e6f7a8b9")
	c.Assert(obj.Commits, check.HasLen, 1)
	c.Assert(obj.Commits[0], check.DeepEquals, commits[0])
}

func (s *S) TestGetMimeTypeFromExtension(c *check.C) {
	path := "my-text-file.txt"
	content := new(bytes.Buffer)
	content.WriteString("")
	c.Assert(getMimeType(path, content.Bytes()), check.Equals, "text/plain; charset=utf-8")
	path = "my-text-file.sh"
	content = new(bytes.Buffer)
	content.WriteString("")
	expected := mime.TypeByExtension(".sh")
	c.Assert(getMimeType(path, content.Bytes()), check.Equals, expected)
}

func (s *S) TestGetMimeTypeFromContent(c *check.C) {
	path := "README"
	content := new(bytes.Buffer)
	content.WriteString("thou shalt not pass")
	c.Assert(getMimeType(path, content.Bytes()), check.Equals, "text/plain; charset=utf-8")
	path = "my-binary-file"
	content = new(bytes.Buffer)
	content.Write([]byte{10, 20, 30, 0, 9, 200})
	c.Assert(getMimeType(path, content.Bytes()), check.Equals, "application/octet-stream")
}
