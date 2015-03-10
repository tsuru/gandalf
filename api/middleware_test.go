// Copyright 2015 gandalf authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/codegangsta/negroni"
	"gopkg.in/check.v1"
)

func (s *S) TestLoggerMiddleware(c *check.C) {
	recorder := httptest.NewRecorder()
	request, err := http.NewRequest("PUT", "/my/path", nil)
	c.Assert(err, check.IsNil)
	var out bytes.Buffer
	middle := loggerMiddleware{
		logger: log.New(&out, "", 0),
	}
	middle.ServeHTTP(negroni.NewResponseWriter(recorder), request, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello"))
	}))
	timePart := time.Now().Format(time.RFC3339Nano)[:19]
	c.Assert(out.String(), check.Matches, fmt.Sprintf(`(?m)%s\..+? PUT /my/path 200 in [\d.]+ms$`, timePart))
}

func (s *S) TestResponseHeaderMiddleware(c *check.C) {
	recorder := httptest.NewRecorder()
	request, err := http.NewRequest("PUT", "/my/path", nil)
	c.Assert(err, check.IsNil)
	middle := responseHeaderMiddleware{name: "Server", value: "super-server/0.1"}
	middle.ServeHTTP(negroni.NewResponseWriter(recorder), request, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("hello"))
	}))
	c.Assert(recorder.Header().Get("Server"), check.Equals, "super-server/0.1")
}

func (s *S) TestResponseHeaderMiddlewareDuplicate(c *check.C) {
	recorder := httptest.NewRecorder()
	request, err := http.NewRequest("PUT", "/my/path", nil)
	c.Assert(err, check.IsNil)
	middle := responseHeaderMiddleware{name: "Server", value: "super-server/0.1"}
	middle.ServeHTTP(negroni.NewResponseWriter(recorder), request, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Server", "mini-server/0.1")
		w.Write([]byte("hello"))
	}))
	c.Assert(recorder.Header().Get("Server"), check.Equals, "mini-server/0.1")
}
