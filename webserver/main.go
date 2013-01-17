// Copyright 2013 gandalf authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"github.com/bmizerany/pat"
	"github.com/globocom/config"
	"github.com/globocom/gandalf/api"
	"github.com/globocom/gandalf/db"
	"log"
	"net/http"
	"os/exec"
)

func startGitDaemon() error {
	bLocation, err := config.GetString("git:bare:location")
	if err != nil {
		return err
	}
	args := []string{"daemon", fmt.Sprintf("--base-path=%s", bLocation), "--syslog"}
	if exportAll, err := config.GetBool("git:daemon:export-all"); err == nil && exportAll {
		args = append(args, "--export-all")
	}
	return exec.Command("git", args...).Run()
}

func main() {
	dry := flag.Bool("dry", false, "dry-run: does not start the server and git daemon (for testing purpose)")
	configFile := flag.String("config", "/etc/gandalf.conf", "Gandalf configuration file")
	flag.Parse()

	err := config.ReadConfigFile(*configFile)
	if err != nil {
		msg := `Could not find gandalf config file. Searched on %s.
For an example conf check gandalf/etc/gandalf.conf file.`
		log.Panicf(msg, *configFile)
	}
	db.Connect()
	router := pat.New()
	router.Post("/user/:name/key", http.HandlerFunc(api.AddKey))
	router.Del("/user/:name/key/:keyname", http.HandlerFunc(api.RemoveKey))
	router.Post("/user", http.HandlerFunc(api.NewUser))
	router.Del("/user/:name", http.HandlerFunc(api.RemoveUser))
	router.Post("/repository", http.HandlerFunc(api.NewRepository))
	router.Post("/repository/grant", http.HandlerFunc(api.GrantAccess))
	router.Del("/repository/revoke", http.HandlerFunc(api.RevokeAccess))
	router.Del("/repository/:name", http.HandlerFunc(api.RemoveRepository))

	port, err := config.GetString("webserver:port")
	if err != nil {
		panic(err)
	}
	if !*dry {
		startGitDaemon()
		log.Fatal(http.ListenAndServe(port, router))
	}
}
