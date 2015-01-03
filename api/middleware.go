// Copyright 2014 gandalf authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/codegangsta/negroni"
)

type loggerMiddleware struct {
	logger *log.Logger
}

func NewLoggerMiddleware() *loggerMiddleware {
	return &loggerMiddleware{
		logger: log.New(os.Stdout, "", 0),
	}
}

func (l *loggerMiddleware) ServeHTTP(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	start := time.Now()
	next(rw, r)
	duration := time.Since(start)
	res := rw.(negroni.ResponseWriter)
	nowFormatted := time.Now().Format(time.RFC3339Nano)
	l.logger.Printf("%s %s %s %d in %0.6fms", nowFormatted, r.Method, r.URL.Path, res.Status(), float64(duration)/float64(time.Millisecond))
}
