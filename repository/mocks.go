// Copyright 2014 gandalf authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package repository

type MockContentRetriever struct {
	LastFormat     ArchiveFormat
	LastRef        string
	LastPath       string
	ResultContents []byte
	Tree           []map[string]string
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
