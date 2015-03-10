// Copyright 2015 gandalf authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/codegangsta/negroni"
	"github.com/tsuru/config"
	"github.com/tsuru/gandalf/api"
)

const version = "0.6.0"

func main() {
	dry := flag.Bool("dry", false, "dry-run: does not start the server (for testing purpose)")
	configFile := flag.String("config", "/etc/gandalf.conf", "Gandalf configuration file")
	gVersion := flag.Bool("version", false, "Print version and exit")
	flag.Parse()
	if *gVersion {
		fmt.Printf("gandalf-webserver version %s\n", version)
		return
	}
	log.Printf("Opening config file: %s ...\n", *configFile)
	err := config.ReadAndWatchConfigFile(*configFile)
	if err != nil {
		msg := `Could not open gandalf config file at %s (%s).
  For an example, see: gandalf/etc/gandalf.conf
  Note that you can specify a different config file with the --config option -- e.g.: --config=./etc/gandalf.conf`
		log.Fatalf(msg, *configFile, err)
	}
	log.Printf("Successfully read config file: %s\n", *configFile)
	router := api.SetupRouter()
	n := negroni.New()
	n.Use(api.NewLoggerMiddleware())
	n.UseHandler(router)
	bind, err := config.GetString("bind")
	if err != nil {
		var perr error
		bind, perr = config.GetString("webserver:port")
		if perr != nil {
			panic(err)
		}
	}
	if !*dry {
		bareLocation, err := config.GetString("git:bare:location")
		if err != nil {
			panic("You should configure a git:bare:location for gandalf.")
		}
		log.Printf("Repository location: %s\n", bareLocation)
		log.Printf("gandalf-webserver %s listening on %s\n", version, bind)
		log.Fatal(http.ListenAndServe(bind, router))
	}
}
