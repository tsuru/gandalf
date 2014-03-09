// Copyright 2014 gandalf authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package hook

import (
	"github.com/globocom/config"
	"strings"
	"io"
	"io/ioutil"
)

// Adds a hook script.
func Add(name string, body io.ReadCloser) error {
	path, _ := config.GetString("git:bare:template")
	s := []string{path, "hooks", name}
	scriptPath := strings.Join(s, "/")
	b, err := ioutil.ReadAll(body)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(scriptPath, b, 0755)
	if err != nil {
		return err
	}
	return nil
}
