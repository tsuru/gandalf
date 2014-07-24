// Copyright 2014 gandalf authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package repository

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/tsuru/config"
	"github.com/tsuru/gandalf/db"
	"github.com/tsuru/gandalf/fs"
	"github.com/tsuru/tsuru/log"
	"io/ioutil"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// Repository represents a Git repository. A Git repository is a record in the
// database and a directory in the filesystem (the bare repository).
type Repository struct {
	Name     string `bson:"_id"`
	Users    []string
	IsPublic bool
}

// exists returns whether the given file or directory exists or not
func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// MarshalJSON marshals the Repository in json format.
func (r *Repository) MarshalJSON() ([]byte, error) {
	data := map[string]interface{}{
		"name":    r.Name,
		"public":  r.IsPublic,
		"ssh_url": r.ReadWriteURL(),
		"git_url": r.ReadOnlyURL(),
	}
	return json.Marshal(&data)
}

// New creates a representation of a git repository. It creates a Git
// repository using the "bare-dir" setting and saves repository's meta data in
// the database.
func New(name string, users []string, isPublic bool) (*Repository, error) {
	log.Debugf("Creating repository %q", name)
	r := &Repository{Name: name, Users: users, IsPublic: isPublic}
	if v, err := r.isValid(); !v {
		log.Errorf("repository.New: Invalid repository %q: %s", name, err)
		return r, err
	}
	if err := newBare(name); err != nil {
		log.Errorf("repository.New: Error creating bare repository for %q: %s", name, err)
		return r, err
	}
	barePath := barePath(name)
	if barePath != "" && isPublic {
		ioutil.WriteFile(barePath+"/git-daemon-export-ok", []byte(""), 0644)
		if f, err := fs.Filesystem().Create(barePath + "/git-daemon-export-ok"); err == nil {
			f.Close()
		}
	}
	conn, err := db.Conn()
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	err = conn.Repository().Insert(&r)
	if mgo.IsDup(err) {
		log.Errorf("repository.New: Duplicate repository %q", name)
		return r, fmt.Errorf("A repository with this name already exists.")
	}
	return r, err
}

// Get find a repository by name.
func Get(name string) (Repository, error) {
	var r Repository
	conn, err := db.Conn()
	if err != nil {
		return r, err
	}
	defer conn.Close()
	err = conn.Repository().FindId(name).One(&r)
	return r, err
}

// Remove deletes the repository from the database and removes it's bare Git
// repository.
func Remove(name string) error {
	log.Debugf("Removing repository %q", name)
	if err := removeBare(name); err != nil {
		log.Errorf("repository.Remove: Error removing bare repository %q: %s", name, err)
		return err
	}
	conn, err := db.Conn()
	if err != nil {
		return err
	}
	defer conn.Close()
	if err := conn.Repository().RemoveId(name); err != nil {
		log.Errorf("repository.Remove: Error removing repository %q from db: %s", name, err)
		return fmt.Errorf("Could not remove repository: %s", err)
	}
	return nil
}

// Rename renames a repository.
func Rename(oldName, newName string) error {
	log.Debugf("Renaming repository %q to %q", oldName, newName)
	repo, err := Get(oldName)
	if err != nil {
		log.Errorf("repository.Rename: Repository %q not found: %s", oldName, err)
		return err
	}
	newRepo := repo
	newRepo.Name = newName
	conn, err := db.Conn()
	if err != nil {
		return err
	}
	defer conn.Close()
	err = conn.Repository().Insert(newRepo)
	if err != nil {
		log.Errorf("repository.Rename: Error adding new repository %q: %s", newName, err)
		return err
	}
	err = conn.Repository().RemoveId(oldName)
	if err != nil {
		log.Errorf("repository.Rename: Error removing old repository %q: %s", oldName, err)
		return err
	}
	return fs.Filesystem().Rename(barePath(oldName), barePath(newName))
}

// ReadWriteURL formats the git ssh url and return it. If no remote is configured in
// gandalf.conf, this method panics.
func (r *Repository) ReadWriteURL() string {
	uid, err := config.GetString("uid")
	if err != nil {
		panic(err.Error())
	}
	remote := uid + "@%s:%s.git"
	if useSSH, _ := config.GetBool("git:ssh:use"); useSSH {
		port, err := config.GetString("git:ssh:port")
		if err == nil {
			remote = "ssh://" + uid + "@%s:" + port + "/%s.git"
		} else {
			remote = "ssh://" + uid + "@%s/%s.git"
		}
	}
	host, err := config.GetString("host")
	if err != nil {
		panic(err.Error())
	}
	return fmt.Sprintf(remote, host, r.Name)
}

// ReadOnly formats the git url and return it. If no host is configured in
// gandalf.conf, this method panics.
func (r *Repository) ReadOnlyURL() string {
	remote := "git://%s/%s.git"
	if useSSH, _ := config.GetBool("git:ssh:use"); useSSH {
		uid, err := config.GetString("uid")
		if err != nil {
			panic(err.Error())
		}
		port, err := config.GetString("git:ssh:port")
		if err == nil {
			remote = "ssh://" + uid + "@%s:" + port + "/%s.git"
		} else {
			remote = "ssh://" + uid + "@%s/%s.git"
		}
	}
	host, err := config.GetString("readonly-host")
	if err != nil {
		host, err = config.GetString("host")
		if err != nil {
			panic(err)
		}
	}
	return fmt.Sprintf(remote, host, r.Name)
}

// Validates a repository
// A valid repository must have:
//  - a name without any special chars only alphanumeric and underlines are allowed.
//  - at least one user in users array
func (r *Repository) isValid() (bool, error) {
	m, e := regexp.Match(`^[\w-]+$`, []byte(r.Name))
	if e != nil {
		panic(e)
	}
	if !m {
		return false, errors.New("Validation Error: repository name is not valid")
	}
	if len(r.Users) == 0 {
		return false, errors.New("Validation Error: repository should have at least one user")
	}
	return true, nil
}

// GrantAccess gives write permission for users in all specified repositories.
// If any of the repositories/users do not exists, GrantAccess just skips it.
func GrantAccess(rNames, uNames []string) error {
	conn, err := db.Conn()
	if err != nil {
		return err
	}
	defer conn.Close()
	_, err = conn.Repository().UpdateAll(bson.M{"_id": bson.M{"$in": rNames}}, bson.M{"$addToSet": bson.M{"users": bson.M{"$each": uNames}}})
	return err
}

// RevokeAccess revokes write permission from users in all specified
// repositories.
func RevokeAccess(rNames, uNames []string) error {
	conn, err := db.Conn()
	if err != nil {
		return err
	}
	defer conn.Close()
	_, err = conn.Repository().UpdateAll(bson.M{"_id": bson.M{"$in": rNames}}, bson.M{"$pullAll": bson.M{"users": uNames}})
	return err
}

type ArchiveFormat int

const (
	Zip ArchiveFormat = iota
	Tar
	TarGz
)

type ContentRetriever interface {
	GetContents(repo, ref, path string) ([]byte, error)
	GetArchive(repo, ref string, format ArchiveFormat) ([]byte, error)
	GetTree(repo, ref, path string) ([]map[string]string, error)
	GetForEachRef(repo, pattern string) ([]map[string]string, error)
	GetBranch(repo string) ([]map[string]string, error)
}

var Retriever ContentRetriever

type GitContentRetriever struct{}

func (*GitContentRetriever) GetContents(repo, ref, path string) ([]byte, error) {
	gitPath, err := exec.LookPath("git")
	if err != nil {
		return nil, fmt.Errorf("Error when trying to obtain file %s on ref %s of repository %s (%s).", path, ref, repo, err)
	}
	cwd := barePath(repo)
	repoExists, err := exists(cwd)
	if err != nil || !repoExists {
		return nil, fmt.Errorf("Error when trying to obtain file %s on ref %s of repository %s (Repository does not exist).", path, ref, repo)
	}
	cmd := exec.Command(gitPath, "show", fmt.Sprintf("%s:%s", ref, path))
	cmd.Dir = cwd
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("Error when trying to obtain file %s on ref %s of repository %s (%s).", path, ref, repo, err)
	}
	return out, nil
}

func (*GitContentRetriever) GetArchive(repo, ref string, format ArchiveFormat) ([]byte, error) {
	gitPath, err := exec.LookPath("git")
	if err != nil {
		return nil, fmt.Errorf("Error when trying to obtain archive for ref %s of repository %s (%s).", ref, repo, err)
	}
	var archiveFormat string
	switch format {
	case Tar:
		archiveFormat = "--format=tar"
	case TarGz:
		archiveFormat = "--format=tar.gz"
	default:
		archiveFormat = "--format=zip"
	}
	prefix := fmt.Sprintf("--prefix=%s-%s/", repo, ref)
	cwd := barePath(repo)
	repoExists, err := exists(cwd)
	if err != nil || !repoExists {
		return nil, fmt.Errorf("Error when trying to obtain archive for ref %s of repository %s (Repository does not exist).", ref, repo)
	}
	cmd := exec.Command(gitPath, "archive", ref, prefix, archiveFormat)
	cmd.Dir = cwd
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("Error when trying to obtain archive for ref %s of repository %s (%s).", ref, repo, err)
	}
	return out, nil
}

func (*GitContentRetriever) GetTree(repo, ref, path string) ([]map[string]string, error) {
	gitPath, err := exec.LookPath("git")
	if err != nil {
		return nil, fmt.Errorf("Error when trying to obtain tree %s on ref %s of repository %s (%s).", path, ref, repo, err)
	}
	cwd := barePath(repo)
	repoExists, err := exists(cwd)
	if err != nil || !repoExists {
		return nil, fmt.Errorf("Error when trying to obtain tree %s on ref %s of repository %s (Repository does not exist).", path, ref, repo)
	}
	cmd := exec.Command(gitPath, "ls-tree", "-r", ref, path)
	cmd.Dir = cwd
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("Error when trying to obtain tree %s on ref %s of repository %s (%s).", path, ref, repo, err)
	}
	lines := strings.Split(string(out), "\n")
	objectCount := 0
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		objectCount++
	}
	objects := make([]map[string]string, objectCount)
	objectCount = 0
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		tabbed := strings.Split(line, "\t")
		meta, filepath := tabbed[0], tabbed[1]
		meta_parts := strings.Split(meta, " ")
		permission, filetype, hash := meta_parts[0], meta_parts[1], meta_parts[2]
		object := make(map[string]string)
		object["permission"] = permission
		object["filetype"] = filetype
		object["hash"] = hash
		object["path"] = strings.TrimSpace(strings.Trim(filepath, "\""))
		object["rawPath"] = filepath
		objects[objectCount] = object
		objectCount++
	}
	return objects, nil
}

func (*GitContentRetriever) GetForEachRef(repo, pattern string) ([]map[string]string, error) {
	var ref, name, commiterName, commiterEmail, commiterDate, authorName, authorEmail, authorDate, subject string
	gitPath, err := exec.LookPath("git")
	if err != nil {
		return nil, fmt.Errorf("Error when trying to obtain the refs of repository %s (%s).", repo, err)
	}
	cwd := barePath(repo)
	repoExists, err := exists(cwd)
	if err != nil || !repoExists {
		return nil, fmt.Errorf("Error when trying to obtain the refs of repository %s (Repository does not exist).", repo)
	}
	cmd := exec.Command(gitPath, "for-each-ref", "--sort=-committerdate", "--format", "%(objectname)%09%(refname:short)%09%(committername)%09%(committeremail)%09%(committerdate)%09%(authorname)%09%(authoremail)%09%(authordate)%09%(contents:subject)", pattern)
	cmd.Dir = cwd
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("Error when trying to obtain the refs of repository %s (%s).", repo, err)
	}
	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	// When the separator does not separate the string, Split returns an array
	// with one element: the string itself. (golang.org/pkg/strings/#Split)
	if len(lines) == 1 && len(lines[0]) == 0 {
		return nil, nil
	}
	objectCount := len(lines)
	objects := make([]map[string]string, objectCount)
	objectCount = 0
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) > 4 { // let there be commits with empty subject
			ref = fields[0]
			name = fields[1]
			commiterName = fields[2]
			commiterEmail = fields[3]
			commiterDate = fields[4]
			authorName = fields[5]
			authorEmail = fields[6]
			authorDate = fields[7]
			subject = strings.Join(fields[8:], "\t") // let there be subjects with \t
		} else {
			return nil, fmt.Errorf("Error when trying to obtain the refs of repository %s (Invalid git for-each-ref output [%s]).", repo, out)
		}
		object := make(map[string]string)
		object["ref"] = ref
		object["name"] = name
		object["commiterName"] = commiterName
		object["commiterEmail"] = commiterEmail
		object["commiterDate"] = commiterDate
		object["authorName"] = authorName
		object["authorEmail"] = authorEmail
		object["authorDate"] = authorDate
		object["subject"] = subject
		objects[objectCount] = object
		objectCount++
	}
	return objects, nil
}

func (*GitContentRetriever) GetBranch(repo string) ([]map[string]string, error) {
	branches, err := retriever().GetForEachRef(repo, "refs/heads/")
	return branches, err
}

func retriever() ContentRetriever {
	if Retriever == nil {
		Retriever = &GitContentRetriever{}
	}
	return Retriever
}

// GetFileContents returns the contents for a given file
// in a given ref for the specified repository
func GetFileContents(repo, ref, path string) ([]byte, error) {
	return retriever().GetContents(repo, ref, path)
}

// GetArchive returns the contents for a given file
// in a given ref for the specified repository
func GetArchive(repo, ref string, format ArchiveFormat) ([]byte, error) {
	return retriever().GetArchive(repo, ref, format)
}

func GetTree(repo, ref, path string) ([]map[string]string, error) {
	return retriever().GetTree(repo, ref, path)
}

func GetForEachRef(repo, pattern string) ([]map[string]string, error) {
	return retriever().GetForEachRef(repo, pattern)
}

func GetBranch(repo string) ([]map[string]string, error) {
	return retriever().GetBranch(repo)
}
