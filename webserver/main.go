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

const version = "0.5.1"

func main() {
	dry := flag.Bool("dry", false, "dry-run: does not start the server (for testing purpose)")
	configFile := flag.String("config", "/etc/gandalf.conf", "Gandalf configuration file")
	gVersion := flag.Bool("version", false, "Print version and exit")
	flag.Parse()
	if *gVersion {
		fmt.Printf("gandalf-webserver version %s\n", version)
		return
	}
	err := config.ReadAndWatchConfigFile(*configFile)
	if err != nil {
		msg := `Could not find gandalf config file. Searched on %s.
For an example conf check gandalf/etc/gandalf.conf file.\n %s`
		log.Panicf(msg, *configFile, err)
	}
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
