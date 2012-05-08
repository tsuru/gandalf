package api

import (
	"io/ioutil"
	. "launchpad.net/gocheck"
	"net/http"
	"net/http/httptest"
	"strings"
)

func (s *S) TestCreateProjectHandlerShouldCreateProject(c *C) {
	b := strings.NewReader(`{"name": "some_project"}`)
	request, err := http.NewRequest("POST", "/projects", b)
	c.Assert(err, IsNil)

	request.Header.Set("Content-Type", "application/json")
	recorder := httptest.NewRecorder()

	err = CreateProjectHandler(recorder, request)
	c.Assert(err, IsNil)

	body, err := ioutil.ReadAll(recorder.Body)
	c.Assert(err, IsNil)

	c.Assert(string(body), Equals, "success")
}
