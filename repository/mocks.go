// Copyright 2014 gandalf authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package repository

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
