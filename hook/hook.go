// Copyright 2014 gandalf authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package hook

import (
	"os"
	"strings"

	"github.com/tsuru/config"
	"github.com/tsuru/gandalf/fs"
)

func createHookFile(path string, content []byte) error {
	file, err := fs.Filesystem().OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(content)
	if err != nil {
		return err
	}
	return nil
}

// Adds a hook script.
func Add(name string, repos []string, content []byte) error {
	configParam := "git:bare:template"
	if len(repos) > 0 {
		configParam = "git:bare:location"
	}
	path, err := config.GetString(configParam)
	if err != nil {
		return err
	}
	if len(repos) > 0 {
		for _, repo := range repos {
			repo += ".git"
			s := []string{path, repo, "hooks"}
			dirPath := strings.Join(s, "/")
			err = fs.Filesystem().MkdirAll(dirPath, 0755)
			if err != nil {
				return err
			}
			s = []string{dirPath, name}
			scriptPath := strings.Join(s, "/")
			err = createHookFile(scriptPath, content)
			if err != nil {
				return err
			}
		}
	} else {
		s := []string{path, "hooks", name}
		scriptPath := strings.Join(s, "/")
		return createHookFile(scriptPath, content)
	}
	return nil
}
