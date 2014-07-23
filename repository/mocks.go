// Copyright 2014 gandalf authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package repository

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path"
)

type MockContentRetriever struct {
	LastFormat     ArchiveFormat
	LastRef        string
	LastPath       string
	ResultContents []byte
	Tree           []map[string]string
	Refs           []map[string]string
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

func CreateTestRepository(tmp_path string, repo string, file string, content string, folders ...string) (func(), error) {
	testPath := path.Join(tmp_path, repo+".git")
	cleanup := func() {
		os.RemoveAll(testPath)
	}
	gitPath, err := exec.LookPath("git")
	if err != nil {
		return cleanup, err
	}
	err = os.MkdirAll(testPath, 0777)
	if err != nil {
		return cleanup, err
	}
	cmd := exec.Command(gitPath, "init")
	cmd.Dir = testPath
	err = cmd.Run()
	if err != nil {
		return cleanup, err
	}
	err = ioutil.WriteFile(path.Join(testPath, file), []byte(content), 0644)
	if err != nil {
		return cleanup, err
	}
	for _, folder := range folders {
		folderPath := path.Join(testPath, folder)
		err = os.MkdirAll(folderPath, 0777)
		if err != nil {
			return cleanup, err
		}
		err = ioutil.WriteFile(path.Join(folderPath, file), []byte(content), 0644)
		if err != nil {
			return cleanup, err
		}
	}
	cmd = exec.Command(gitPath, "add", ".")
	cmd.Dir = testPath
	err = cmd.Run()
	if err != nil {
		return cleanup, err
	}
	cmd = exec.Command(gitPath, "config", "user.email", "much@email.com")
	cmd.Dir = testPath
	err = cmd.Run()
	if err != nil {
		return cleanup, err
	}
	cmd = exec.Command(gitPath, "config", "user.name", "doge")
	cmd.Dir = testPath
	err = cmd.Run()
	if err != nil {
		return cleanup, err
	}
	cmd = exec.Command(gitPath, "commit", "-m", content, "--allow-empty-message")
	cmd.Dir = testPath
	err = cmd.Run()
	return cleanup, err
}

func CreateBranchesOnTestRepository(tmp_path string, repo string, file string, content string, branches ...string) error {
	testPath := path.Join(tmp_path, repo+".git")
	gitPath, err := exec.LookPath("git")
	if err != nil {
		return err
	}
	cmd := exec.Command(gitPath, "status")
	cmd.Dir = testPath
	err = cmd.Run()
	if err != nil {
		return err
	}
	for _, branch := range branches {
		fp, err := os.OpenFile(path.Join(testPath, file), os.O_APPEND|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		defer fp.Close()
		_, err = fp.WriteString("such string")
		if err != nil {
			return err
		}
		cmd = exec.Command(gitPath, "checkout", "-b", branch)
		cmd.Dir = testPath
		err = cmd.Run()
		if err != nil {
			return err
		}
		cmd = exec.Command(gitPath, "add", ".")
		cmd.Dir = testPath
		err = cmd.Run()
		if err != nil {
			return err
		}
		if len(content) > 0 {
			cmd = exec.Command(gitPath, "commit", "-m", content+" on "+branch)
		} else {
			cmd = exec.Command(gitPath, "commit", "-m", "", "--allow-empty-message")
		}
		cmd.Dir = testPath
		err = cmd.Run()
		if err != nil {
			return err
		}
	}
	return err
}

func (r *MockContentRetriever) GetTree(repo, ref, path string) ([]map[string]string, error) {
	if r.LookPathError != nil {
		return nil, r.LookPathError
	}
	if r.OutputError != nil {
		return nil, r.OutputError
	}
	r.LastRef = ref
	r.LastPath = path
	return r.Tree, nil
}

func (r *MockContentRetriever) GetForEachRef(repo, pattern string) ([]map[string]string, error) {
	if r.LookPathError != nil {
		return nil, r.LookPathError
	}
	if r.OutputError != nil {
		return nil, r.OutputError
	}
	return r.Refs, nil
}

func (r *MockContentRetriever) GetBranch(repo string) ([]map[string]string, error) {
	if r.LookPathError != nil {
		return nil, r.LookPathError
	}
	if r.OutputError != nil {
		return nil, r.OutputError
	}
	return r.Refs, nil
}
