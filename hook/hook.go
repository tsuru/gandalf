// Copyright 2014 gandalf authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package hook

import (
	"github.com/tsuru/config"
	"github.com/tsuru/gandalf/fs"
	"io"
	"os"
	"strings"
)

// Adds a hook script.
func Add(name string, body io.Reader) error {
	path, err := config.GetString("git:bare:template")
	if err != nil {
		return err
	}
	s := []string{path, "hooks", name}
	scriptPath := strings.Join(s, "/")
	file, err := fs.Filesystem().OpenFile(scriptPath, os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = io.Copy(file, body)
	if err != nil {
		return err
	}
	return nil
}
