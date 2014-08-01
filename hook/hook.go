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

func createHookFile(path string, body io.Reader) error {
	file, err := fs.Filesystem().OpenFile(path, os.O_WRONLY|os.O_CREATE, 0755)
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

// Adds a hook script.
func Add(name string, body io.Reader) error {
	path, err := config.GetString("git:bare:template")
	if err != nil {
		return err
	}
	s := []string{path, "hooks", name}
	scriptPath := strings.Join(s, "/")
	return createHookFile(scriptPath, body)
}

// Adds a hook script for a repository
func AddRepository(name string, repos []string, body io.Reader) error {
	path, err := config.GetString("git:bare:location")
	if err != nil {
		return err
	}
	for _, repo := range repos {
		repo += ".git"
		s := []string{path, repo, "hooks", name}
		scriptPath := strings.Join(s, "/")
		err := createHookFile(scriptPath, body)
		if err != nil {
			return err
		}
	}
	return nil
}
