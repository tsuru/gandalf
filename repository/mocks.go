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

func CreateTestRepository(tmp_path string, repo string, file string, content string) func() {
	gitPath, _ := exec.LookPath("git")
	testPath := path.Join(tmp_path, repo+".git")
	exec.Command("mkdir", "-p", testPath).Output()

	cmd := exec.Command(gitPath, "init")
	cmd.Dir = testPath
	cmd.Output()

	err := ioutil.WriteFile(path.Join(testPath, file), []byte(content), 0644)
	if err != nil {
		panic(err)
	}

	cmd = exec.Command(gitPath, "add", ".")
	cmd.Dir = testPath
	cmd.Output()

	cmd = exec.Command(gitPath, "commit", "-m", content)
	cmd.Dir = testPath
	cmd.Output()

	return func() {
		exec.Command("rm", "-rf", testPath).Output()
	}
}
