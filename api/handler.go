// Copyright 2015 gandalf authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"path/filepath"
	"reflect"
	"strconv"

	"github.com/gorilla/pat"
	"github.com/tsuru/config"
	"github.com/tsuru/gandalf/db"
	"github.com/tsuru/gandalf/hook"
	"github.com/tsuru/gandalf/multipartzip"
	"github.com/tsuru/gandalf/repository"
	"github.com/tsuru/gandalf/user"
)

var maxMemory uint

func maxMemoryValue() uint {
	if maxMemory > 0 {
		return maxMemory
	}
	var err error
	maxMemory, err = config.GetUint("api:request:maxMemory")
	if err != nil {
		panic("You should configure a api:request:maxMemory for gandalf.")
	}
	return maxMemory
}

func accessParameters(body io.ReadCloser) (repositories, users []string, err error) {
	var params map[string][]string
	if err := parseBody(body, &params); err != nil {
		return []string{}, []string{}, err
	}
	users, ok := params["users"]
	if !ok {
		return []string{}, []string{}, errors.New("It is need a user list")
	}
	repositories, ok = params["repositories"]
	if !ok {
		return []string{}, []string{}, errors.New("It is need a repository list")
	}
	return repositories, users, nil
}

func SetupRouter() *pat.Router {
	router := pat.New()
	router.Post("/user/{name}/key", http.HandlerFunc(addKey))
	router.Delete("/user/{name}/key/{keyname}", http.HandlerFunc(removeKey))
	router.Put("/user/{name}/key/{keyname}", http.HandlerFunc(updateKey))
	router.Get("/user/{name}/keys", http.HandlerFunc(listKeys))
	router.Post("/user", http.HandlerFunc(newUser))
	router.Delete("/user/{name}", http.HandlerFunc(removeUser))
	router.Delete("/repository/revoke", http.HandlerFunc(revokeAccess))
	router.Get("/repository/{name:[^/]*/?[^/]+}/archive", http.HandlerFunc(getArchive))
	router.Get("/repository/{name:[^/]*/?[^/]+}/contents", http.HandlerFunc(getFileContents))
	router.Get("/repository/{name:[^/]*/?[^/]+}/tree", http.HandlerFunc(getTree))
	router.Get("/repository/{name:[^/]*/?[^/]+}/branches", http.HandlerFunc(getBranches))
	router.Get("/repository/{name:[^/]*/?[^/]+}/tags", http.HandlerFunc(getTags))
	router.Get("/repository/{name:[^/]*/?[^/]+}/diff/commits", http.HandlerFunc(getDiff))
	router.Post("/repository/{name:[^/]*/?[^/]+}/commit", http.HandlerFunc(commit))
	router.Get("/repository/{name:[^/]*/?[^/]+}/logs", http.HandlerFunc(getLogs))
	router.Post("/repository/grant", http.HandlerFunc(grantAccess))
	router.Post("/repository", http.HandlerFunc(newRepository))
	router.Get("/repository/{name:[^/]*/?[^/]+}", http.HandlerFunc(getRepository))
	router.Delete("/repository/{name:[^/]*/?[^/]+}", http.HandlerFunc(removeRepository))
	router.Put("/repository/{name:[^/]*/?[^/]+}", http.HandlerFunc(updateRepository))
	router.Get("/healthcheck", http.HandlerFunc(healthCheck))
	router.Post("/hook/{name}", http.HandlerFunc(addHook))
	return router
}

func grantAccess(w http.ResponseWriter, r *http.Request) {
	repositories, users, err := accessParameters(r.Body)
	readOnly := r.URL.Query().Get("readonly") == "yes"
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := repository.GrantAccess(repositories, users, readOnly); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	if readOnly {
		fmt.Fprintf(w, "Successfully granted read-only access to users \"%s\" into repository \"%s\"", users, repositories)
	} else {
		fmt.Fprintf(w, "Successfully granted full access to users \"%s\" into repository \"%s\"", users, repositories)
	}
}

func revokeAccess(w http.ResponseWriter, r *http.Request) {
	repositories, users, err := accessParameters(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := repository.RevokeAccess(repositories, users, true); err != nil {
		status := http.StatusInternalServerError
		if err == repository.ErrRepositoryNotFound {
			status = http.StatusNotFound
		}
		http.Error(w, err.Error(), status)
		return
	}
	if err := repository.RevokeAccess(repositories, users, false); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "Successfully revoked access to users \"%s\" into repositories \"%s\"", users, repositories)
}

func addKey(w http.ResponseWriter, r *http.Request) {
	keys := map[string]string{}
	if err := parseBody(r.Body, &keys); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(keys) == 0 {
		http.Error(w, "A key is needed", http.StatusBadRequest)
		return
	}
	uName := r.URL.Query().Get(":name")
	if err := user.AddKey(uName, keys); err != nil {
		switch err {
		case user.ErrInvalidKey:
			http.Error(w, err.Error(), http.StatusBadRequest)
		case user.ErrDuplicateKey:
			http.Error(w, "Key already exists.", http.StatusConflict)
		case user.ErrUserNotFound:
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	fmt.Fprint(w, "Key(s) successfully created")
}

func updateKey(w http.ResponseWriter, r *http.Request) {
	uName := r.URL.Query().Get(":name")
	kName := r.URL.Query().Get(":keyname")
	defer r.Body.Close()
	content, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	key := user.Key{Name: kName, Body: string(content)}
	if err := user.UpdateKey(uName, key); err != nil {
		switch err {
		case user.ErrInvalidKey:
			http.Error(w, err.Error(), http.StatusBadRequest)
		case user.ErrUserNotFound, user.ErrKeyNotFound:
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	fmt.Fprintf(w, "Key %q successfully updated!", kName)
}

func removeKey(w http.ResponseWriter, r *http.Request) {
	uName := r.URL.Query().Get(":name")
	kName := r.URL.Query().Get(":keyname")
	if err := user.RemoveKey(uName, kName); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	fmt.Fprintf(w, "Key \"%s\" successfully removed", kName)
}

func listKeys(w http.ResponseWriter, r *http.Request) {
	uName := r.URL.Query().Get(":name")
	keys, err := user.ListKeys(uName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	out, err := json.Marshal(&keys)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(out)
}

type jsonUser struct {
	Name string
	Keys map[string]string
}

func newUser(w http.ResponseWriter, r *http.Request) {
	var usr jsonUser
	if err := parseBody(r.Body, &usr); err != nil {
		http.Error(w, "Got error while parsing body: "+err.Error(), http.StatusBadRequest)
		return
	}
	u, err := user.New(usr.Name, usr.Keys)
	if err != nil {
		status := http.StatusInternalServerError
		if err == user.ErrUserAlreadyExists {
			status = http.StatusConflict
		}
		if _, ok := err.(*user.InvalidUserError); ok {
			status = http.StatusBadRequest
		}
		http.Error(w, err.Error(), status)
		return
	}
	fmt.Fprintf(w, "User \"%s\" successfully created\n", u.Name)
}

func removeUser(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get(":name")
	if err := user.Remove(name); err != nil {
		status := http.StatusInternalServerError
		if err == user.ErrUserNotFound {
			status = http.StatusNotFound
		}
		http.Error(w, err.Error(), status)
		return
	}
	fmt.Fprintf(w, "User \"%s\" successfully removed\n", name)
}

func newRepository(w http.ResponseWriter, r *http.Request) {
	var repo repository.Repository
	if err := parseBody(r.Body, &repo); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	_, err := repository.New(repo.Name, repo.Users, repo.ReadOnlyUsers, repo.IsPublic)
	if err != nil {
		status := http.StatusInternalServerError
		if err == repository.ErrRepositoryAlreadyExists {
			status = http.StatusConflict
		}
		if _, ok := err.(*repository.InvalidRepositoryError); ok {
			status = http.StatusBadRequest
		}
		http.Error(w, err.Error(), status)
		return
	}
	fmt.Fprintf(w, "Repository \"%s\" successfully created\n", repo.Name)
}

func getRepository(w http.ResponseWriter, r *http.Request) {
	repo, err := repository.Get(r.URL.Query().Get(":name"))
	if err != nil {
		status := http.StatusInternalServerError
		if err == repository.ErrRepositoryNotFound {
			status = http.StatusNotFound
		}
		http.Error(w, err.Error(), status)
		return
	}
	out, err := json.Marshal(&repo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(out)
}

func removeRepository(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get(":name")
	if err := repository.Remove(name); err != nil {
		status := http.StatusBadRequest
		if err == repository.ErrRepositoryNotFound {
			status = http.StatusNotFound
		}
		http.Error(w, err.Error(), status)
		return
	}
	fmt.Fprintf(w, "Repository \"%s\" successfully removed\n", name)
}

func updateRepository(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get(":name")
	repo, err := repository.Get(name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	defer r.Body.Close()
	err = parseBody(r.Body, &repo)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = repository.Update(name, repo)
	if err != nil && err == repository.ErrRepositoryNotFound {
		http.Error(w, err.Error(), http.StatusNotFound)
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

type repositoryHook struct {
	Repositories []string
	Content      string
}

func addHook(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get(":name")
	if name != "post-receive" && name != "pre-receive" && name != "update" {
		http.Error(w,
			"Unsupported hook, valid options are: post-receive, pre-receive or update",
			http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	var params repositoryHook
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	repos := []string{}
	if err := json.Unmarshal(body, &params); err != nil {
		if err := hook.Add(name, repos, body); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	} else {
		repos = params.Repositories
		if err := hook.Add(name, repos, []byte(params.Content)); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	if len(repos) > 0 {
		fmt.Fprint(w, "hook ", name, " successfully created for ", repos, "\n")
	} else {
		fmt.Fprint(w, "hook ", name, " successfully created\n")
	}
}

func parseBody(body io.ReadCloser, result interface{}) error {
	if reflect.ValueOf(result).Kind() == reflect.Struct {
		return errors.New("parseBody function cannot deal with struct. Use pointer")
	}
	b, err := ioutil.ReadAll(body)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(b, &result); err != nil {
		return errors.New(fmt.Sprintf("Could not parse json: %s", err.Error()))
	}
	return nil
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	conn, err := db.Conn()
	if err != nil {
		return
	}
	defer conn.Close()
	if err := conn.User().Database.Session.Ping(); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Failed to ping the database: %s\n", err)
		return
	}
	w.Write([]byte("WORKING"))
}

func getMimeType(path string, content []byte) string {
	extension := filepath.Ext(path)
	mimeType := mime.TypeByExtension(extension)
	if mimeType == "" {
		mimeType = http.DetectContentType(content)
	}
	return mimeType
}

func getFileContents(w http.ResponseWriter, r *http.Request) {
	repo := r.URL.Query().Get(":name")
	path := r.URL.Query().Get("path")
	ref := r.URL.Query().Get("ref")
	if ref == "" {
		ref = "master"
	}
	if path == "" {
		err := fmt.Errorf("Error when trying to obtain an uknown file on ref %s of repository %s (path is required).", ref, repo)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	contents, err := repository.GetFileContents(repo, ref, path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", getMimeType(path, contents))
	w.Header().Set("Content-Length", strconv.Itoa(len(contents)))
	w.Write(contents)
}

func getArchive(w http.ResponseWriter, r *http.Request) {
	repo := r.URL.Query().Get(":name")
	ref := r.URL.Query().Get("ref")
	format := r.URL.Query().Get("format")
	if ref == "" || format == "" {
		err := fmt.Errorf("Error when trying to obtain archive for ref '%s' (format: %s) of repository '%s' (ref and format are required).", ref, format, repo)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var archiveFormat repository.ArchiveFormat
	switch {
	case format == "tar":
		archiveFormat = repository.Tar
	case format == "tar.gz":
		archiveFormat = repository.TarGz
	default:
		archiveFormat = repository.Zip
	}
	contents, err := repository.GetArchive(repo, ref, archiveFormat)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	// Default headers
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s_%s.%s\"", repo, ref, format))
	w.Header().Set("Content-Transfer-Encoding", "binary")
	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("Content-Length", strconv.Itoa(len(contents)))
	// Prevent Caching of File
	w.Header().Set("Cache-Control", "private")
	w.Header().Set("Pragma", "private")
	w.Header().Set("Expires", "Mon, 26 Jul 1997 05:00:00 GMT")
	w.Write(contents)
}

func getTree(w http.ResponseWriter, r *http.Request) {
	repo := r.URL.Query().Get(":name")
	path := r.URL.Query().Get("path")
	ref := r.URL.Query().Get("ref")
	if ref == "" {
		ref = "master"
	}
	if path == "" {
		path = "."
	}
	tree, err := repository.GetTree(repo, ref, path)
	if err != nil {
		err = fmt.Errorf("Error when trying to obtain tree for path %s on ref %s of repository %s (%s).", path, ref, repo, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	b, err := json.Marshal(tree)
	if err != nil {
		err = fmt.Errorf("Error when trying to obtain tree for path %s on ref %s of repository %s (%s).", path, ref, repo, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(b)
}

func getBranches(w http.ResponseWriter, r *http.Request) {
	repo := r.URL.Query().Get(":name")
	branches, err := repository.GetBranches(repo)
	if err != nil {
		err = fmt.Errorf("Error when trying to obtain the branches of repository %s (%s).", repo, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	b, err := json.Marshal(branches)
	if err != nil {
		err = fmt.Errorf("Error when trying to obtain the branches of repository %s (%s).", repo, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(b)
}

func getTags(w http.ResponseWriter, r *http.Request) {
	repo := r.URL.Query().Get(":name")
	ref := r.URL.Query().Get("ref")
	tags, err := repository.GetTags(repo)
	if err != nil {
		err = fmt.Errorf("Error when trying to obtain tags on ref %s of repository %s (%s).", ref, repo, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	b, err := json.Marshal(tags)
	if err != nil {
		err = fmt.Errorf("Error when trying to obtain tags on ref %s of repository %s (%s).", ref, repo, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(b)
}

func getDiff(w http.ResponseWriter, r *http.Request) {
	repo := r.URL.Query().Get(":name")
	previousCommit := r.URL.Query().Get("previous_commit")
	lastCommit := r.URL.Query().Get("last_commit")
	if previousCommit == "" || lastCommit == "" {
		err := fmt.Errorf("Error when trying to obtain diff between hash commits of repository %s (Hash Commit(s) are required).", repo)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	diff, err := repository.GetDiff(repo, previousCommit, lastCommit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Length", strconv.Itoa(len(diff)))
	w.Write(diff)
}

func commit(w http.ResponseWriter, r *http.Request) {
	repo := r.URL.Query().Get(":name")
	err := r.ParseMultipartForm(int64(maxMemoryValue()))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	form := r.MultipartForm
	data := map[string]string{
		"branch":          "",
		"message":         "",
		"author-name":     "",
		"author-email":    "",
		"committer-name":  "",
		"committer-email": "",
	}
	for key := range data {
		data[key], err = multipartzip.ValueField(form, key)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	commit := repository.GitCommit{
		Branch:  data["branch"],
		Message: data["message"],
		Author: repository.GitUser{
			Name:  data["author-name"],
			Email: data["author-email"],
		},
		Committer: repository.GitUser{
			Name:  data["committer-name"],
			Email: data["committer-email"],
		},
	}
	ref, err := repository.CommitZip(repo, r.MultipartForm.File["zipfile"][0], commit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	b, err := json.Marshal(ref)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(b)
}

func getLogs(w http.ResponseWriter, r *http.Request) {
	repo := r.URL.Query().Get(":name")
	ref := r.URL.Query().Get("ref")
	path := r.URL.Query().Get("path")
	total, err := strconv.Atoi(r.URL.Query().Get("total"))
	if err != nil {
		err = fmt.Errorf("Error when trying to obtain logs for ref %s of repository %s (%s).", ref, repo, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	logs, err := repository.GetLogs(repo, ref, total, path)
	if err != nil {
		err = fmt.Errorf("Error when trying to obtain logs for ref %s of repository %s (%s).", ref, repo, err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	b, err := json.Marshal(logs)
	if err != nil {
		err = fmt.Errorf("Error when trying to obtain logs for ref %s of repository %s (%s).", ref, repo, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(b)
}
