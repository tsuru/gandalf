// Copyright 2014 gandalf authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/tsuru/config"
	"github.com/tsuru/gandalf/api"
)

const version = "0.5.2"

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
	bind, err := config.GetString("bind")
	if err != nil {
		var perr error
		bind, perr = config.GetString("webserver:port")
		if perr != nil {
			panic(err)
		}
	}
	if !*dry {
		log.Fatal(http.ListenAndServe(bind, router))
	}
}
