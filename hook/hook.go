// Copyright 2014 gandalf authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package hook

import (
	"github.com/globocom/config"
	"github.com/globocom/gandalf/fs"
	"os"
	"strings"
	"syscall"
)

type JsonHookScript struct {
	Name    string
	Content string
}

// Adds a hook script.
func Add(hookScript JsonHookScript) error {
	path, _ := config.GetString("git:bare:template")
	s := []string{path, "hooks", hookScript.Name}
	scriptPath := strings.Join(s, "/")
	file, err := fs.Filesystem().OpenFile(scriptPath, os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		return err
	}
	defer file.Close()
	syscall.Flock(int(file.Fd()), syscall.LOCK_EX)
	defer syscall.Flock(int(file.Fd()), syscall.LOCK_UN)
	_, err = file.WriteString(hookScript.Content)
	if err != nil {
		return err
	}
	return nil
}
