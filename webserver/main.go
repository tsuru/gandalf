// Copyright 2015 gandalf authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/codegangsta/negroni"
	"github.com/google/gops/agent"

	"github.com/tsuru/config"
	"github.com/tsuru/gandalf/api"
	"github.com/tsuru/tsuru/log"
)

const version = "0.7.3"

func main() {
	dry := flag.Bool("dry", false, "dry-run: does not start the server (for testing purpose)")
	configFile := flag.String("config", "/etc/gandalf.conf", "Gandalf configuration file")
	gVersion := flag.Bool("version", false, "Print version and exit")
	diagnostic := flag.Bool("diagnostic", false, "Start diagnostics agent with github.com/google/gops. Ignored when running with -dry")
	flag.Parse()
	if *gVersion {
		fmt.Printf("gandalf-webserver version %s\n", version)
		return
	}
	log.Debugf("Opening config file: %s ...\n", *configFile)
	err := config.ReadAndWatchConfigFile(*configFile)
	if err != nil {
		msg := `Could not open gandalf config file at %s (%s).
  For an example, see: gandalf/etc/gandalf.conf
  Note that you can specify a different config file with the --config option -- e.g.: --config=./etc/gandalf.conf`
		log.Fatalf(msg, *configFile, err)
	}
	log.Init()
	log.Debugf("Successfully read config file: %s\n", *configFile)
	router := api.SetupRouter()
	n := negroni.New()
	n.Use(api.NewLoggerMiddleware())
	n.Use(api.NewResponseHeaderMiddleware("Server", "gandalf-webserver/"+version))
	n.Use(api.NewResponseHeaderMiddleware("Cache-Control", "private, max-age=0"))
	n.Use(api.NewResponseHeaderMiddleware("Expires", "-1"))
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

		if *diagnostic {
			if err := agent.Listen(nil); err != nil {
				log.Fatal(err.Error())
			}
			fmt.Println("Diagnostics agent started")
		}

		fmt.Printf("Repository location: %s\n", bareLocation)
		fmt.Printf("gandalf-webserver %s listening on %s\n", version, bind)
		http.ListenAndServe(bind, router)
	}
}
