// Copyright 2014 gandalf authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package repository

import (
	"io/ioutil"
	"os/exec"
	"path"
)

type MockContentRetriever struct {
	LastFormat     ArchiveFormat
	LastRef        string
	ResultContents []byte
	LookPathError  error
	OutputError    error
}

func (r *MockContentRetriever) GetContents(repo, ref, path string) ([]byte, error) {
	if r.LookPathError != nil {
		return nil, r.LookPathError
	}

	if r.OutputError != nil {
		return nil, r.OutputError
	}

	r.LastRef = ref
	return r.ResultContents, nil
}

func (r *MockContentRetriever) GetArchive(repo, ref string, format ArchiveFormat) ([]byte, error) {
	if r.LookPathError != nil {
		return nil, r.LookPathError
	}

	if r.OutputError != nil {
		return nil, r.OutputError
	}

	r.LastRef = ref
	r.LastFormat = format
	return r.ResultContents, nil
}

func CreateTestRepository(tmp_path string, repo string, file string, content string) (func(), error) {
	testPath := path.Join(tmp_path, repo+".git")
	cleanup := func() {
		exec.Command("rm", "-rf", testPath).Output()
	}
	gitPath, err := exec.LookPath("git")
	if err != nil {
		return cleanup, err
	}
	out, err := exec.Command("mkdir", "-p", testPath).Output()
	if err != nil {
		return cleanup, err
	}
	cmd := exec.Command(gitPath, "init")
	cmd.Dir = testPath
	out, err = cmd.Output()
	if err != nil {
		return cleanup, err
	}
	err = ioutil.WriteFile(path.Join(testPath, file), []byte(content), 0644)
	if err != nil {
		return cleanup, err
	}
	cmd = exec.Command(gitPath, "add", file)
	cmd.Dir = testPath
	out, err = cmd.Output()
	if err != nil {
		return cleanup, err
	}
	cmd = exec.Command(gitPath, "config", "user.email", "much@email.com")
	cmd.Dir = testPath
	out, err = cmd.Output()
	if err != nil {
		return cleanup, err
	}
	cmd = exec.Command(gitPath, "config", "user.name", "doge")
	cmd.Dir = testPath
	out, err = cmd.Output()
	if err != nil {
		return cleanup, err
	}
	cmd = exec.Command(gitPath, "commit", "-m", content)
	cmd.Dir = testPath
	out, err = cmd.Output()
	return cleanup, err
}
