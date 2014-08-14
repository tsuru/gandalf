// Copyright 2014 gandalf authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package multipartzip

import (
	"archive/zip"
	"bytes"
	"io"
	"io/ioutil"
	"launchpad.net/gocheck"
	"mime/multipart"
	"os"
	"path"
	"testing"
)

func Test(t *testing.T) { gocheck.TestingT(t) }

type S struct {
	tmpdir string
}

var _ = gocheck.Suite(&S{})

func (s *S) TestCopyZipFile(c *gocheck.C) {
	tempDir, err := ioutil.TempDir("", "TestCopyZipFileDir")
	defer func() {
		os.RemoveAll(tempDir)
	}()
	c.Assert(err, gocheck.IsNil)
	var files = []File{
		{"doge.txt", "Much doge"},
		{"much.txt", "Much mucho"},
		{"WOW/WOW.WOW1", "WOW\nWOW"},
		{"WOW/WOW.WOW2", "WOW\nWOW"},
		{"/usr/WOW/WOW.WOW3", "WOW\nWOW"},
		{"/usr/WOW/WOW.WOW4", "WOW\nWOW"},
	}
	buf, err := CreateZipBuffer(files)
	c.Assert(err, gocheck.IsNil)
	c.Assert(buf, gocheck.NotNil)
	r, err := zip.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
	for _, f := range r.File {
		err = CopyZipFile(f, tempDir, f.Name)
		c.Assert(err, gocheck.IsNil)
		fstat, errStat := os.Stat(path.Join(tempDir, f.Name))
		c.Assert(errStat, gocheck.IsNil)
		c.Assert(fstat.IsDir(), gocheck.Equals, false)
	}
}

func (s *S) TestExtractZip(c *gocheck.C) {
	boundary := "muchBOUNDARY"
	params := map[string]string{}
	var files = []File{
		{"doge.txt", "Much doge"},
		{"much.txt", "Much mucho"},
		{"WOW/WOW.WOW1", "WOW\nWOW"},
		{"WOW/WOW.WOW2", "WOW\nWOW"},
		{"/usr/WOW/WOW.WOW3", "WOW\nWOW"},
		{"/usr/WOW/WOW.WOW4", "WOW\nWOW"},
	}
	buf, err := CreateZipBuffer(files)
	c.Assert(err, gocheck.IsNil)
	reader, writer := io.Pipe()
	go StreamWriteMultipartForm(params, "zipfile", "scaffold.zip", boundary, writer, buf)
	mpr := multipart.NewReader(reader, boundary)
	form, err := mpr.ReadForm(0)
	c.Assert(err, gocheck.IsNil)
	formfile := form.File["zipfile"][0]
	tempDir, err := ioutil.TempDir("", "TestCopyZipFileDir")
	defer func() {
		os.RemoveAll(tempDir)
	}()
	c.Assert(err, gocheck.IsNil)
	ExtractZip(formfile, tempDir)
	for _, file := range files {
		body, err := ioutil.ReadFile(path.Join(tempDir, file.Name))
		c.Assert(err, gocheck.IsNil)
		c.Assert(string(body), gocheck.Equals, file.Body)
	}
}

func (s *S) TestValueField(c *gocheck.C) {
	boundary := "muchBOUNDARY"
	params := map[string]string{
		"committername":  "Barking Doge",
		"committerEmail": "bark@much.com",
		"authorName":     "Doge Dog",
		"authorEmail":    "doge@much.com",
		"message":        "Repository scaffold",
		"branch":         "master",
	}
	reader, writer := io.Pipe()
	go StreamWriteMultipartForm(params, "", "", boundary, writer, nil)
	mpr := multipart.NewReader(reader, boundary)
	form, err := mpr.ReadForm(0)
	c.Assert(err, gocheck.IsNil)
	value, err := ValueField(form, "branch")
	c.Assert(err, gocheck.IsNil)
	c.Assert(value, gocheck.Equals, "master")
}

func (s *S) TestValueFieldWhenFieldInvalid(c *gocheck.C) {
	boundary := "muchBOUNDARY"
	params := map[string]string{}
	reader, writer := io.Pipe()
	go StreamWriteMultipartForm(params, "", "", boundary, writer, nil)
	mpr := multipart.NewReader(reader, boundary)
	form, err := mpr.ReadForm(0)
	c.Assert(err, gocheck.IsNil)
	_, err = ValueField(form, "dleif_dilavni")
	c.Assert(err, gocheck.NotNil)
	c.Assert(err.Error(), gocheck.Equals, "Invalid value field \"dleif_dilavni\"")
}

func (s *S) TestValueFieldWhenFieldEmpty(c *gocheck.C) {
	boundary := "muchBOUNDARY"
	params := map[string]string{
		"branch": "",
	}
	reader, writer := io.Pipe()
	go StreamWriteMultipartForm(params, "", "", boundary, writer, nil)
	mpr := multipart.NewReader(reader, boundary)
	form, err := mpr.ReadForm(0)
	c.Assert(err, gocheck.IsNil)
	_, err = ValueField(form, "branch")
	c.Assert(err, gocheck.NotNil)
	c.Assert(err.Error(), gocheck.Equals, "Empty value \"branch\"")
}

func (s *S) TestFileField(c *gocheck.C) {
	boundary := "muchBOUNDARY"
	params := map[string]string{}
	var files = []File{
		{"doge.txt", "Much doge"},
		{"much.txt", "Much mucho"},
		{"WOW/WOW.WOW1", "WOW\nWOW"},
		{"WOW/WOW.WOW2", "WOW\nWOW"},
		{"/usr/WOW/WOW.WOW3", "WOW\nWOW"},
		{"/usr/WOW/WOW.WOW4", "WOW\nWOW"},
	}
	buf, err := CreateZipBuffer(files)
	c.Assert(err, gocheck.IsNil)
	reader, writer := io.Pipe()
	go StreamWriteMultipartForm(params, "muchfile", "muchfile.zip", boundary, writer, buf)
	mpr := multipart.NewReader(reader, boundary)
	form, err := mpr.ReadForm(0)
	c.Assert(err, gocheck.IsNil)
	file, err := FileField(form, "muchfile")
	c.Assert(err, gocheck.IsNil)
	c.Assert(file.Filename, gocheck.Equals, "muchfile.zip")
}

func (s *S) TestFileFieldWhenFieldInvalid(c *gocheck.C) {
	boundary := "muchBOUNDARY"
	params := map[string]string{
		"dleif_dilavni": "dleif_dilavni",
	}
	reader, writer := io.Pipe()
	go StreamWriteMultipartForm(params, "", "", boundary, writer, nil)
	mpr := multipart.NewReader(reader, boundary)
	form, err := mpr.ReadForm(0)
	c.Assert(err, gocheck.IsNil)
	_, err = FileField(form, "dleif_dilavni")
	c.Assert(err, gocheck.NotNil)
	c.Assert(err.Error(), gocheck.Equals, "Invalid file field \"dleif_dilavni\"")
}

func (s *S) TestFileFieldWhenFieldEmpty(c *gocheck.C) {
	boundary := "muchBOUNDARY"
	params := map[string]string{}
	reader, writer := io.Pipe()
	go StreamWriteMultipartForm(params, "muchfile", "muchfile.zip", boundary, writer, nil)
	mpr := multipart.NewReader(reader, boundary)
	form, err := mpr.ReadForm(0)
	c.Assert(err, gocheck.IsNil)
	file, err := FileField(form, "muchfile")
	c.Assert(err, gocheck.IsNil)
	c.Assert(file.Filename, gocheck.Equals, "muchfile.zip")
	fp, err := file.Open()
	c.Assert(err, gocheck.IsNil)
	fs, err := fp.Seek(0, 2)
	c.Assert(err, gocheck.IsNil)
	c.Assert(fs, gocheck.Equals, int64(0))
}
