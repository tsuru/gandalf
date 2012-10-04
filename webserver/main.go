package main

import (
	"flag"
	"github.com/bmizerany/pat"
	"github.com/globocom/config"
	"github.com/timeredbull/gandalf/api"
	"log"
	"net/http"
)

func main() {
	dry := flag.Bool("dry", false, "dry-run: does not start the server (for testing purpose)")
	flag.Parse()

	err := config.ReadConfigFile("/etc/gandalf.conf")
	if err != nil {
		msg := `Could not find gandalf config file. Searched on /etc/gandalf.conf.
For an example conf check gandalf/etc/gandalf.conf file.`
		panic(msg)
	}
	router := pat.New()
	router.Post("/user", http.HandlerFunc(api.NewUser))
	router.Post("/repository", http.HandlerFunc(api.NewRepository))

	if !*dry {
		log.Fatal(http.ListenAndServe(":8080", router))
	}
}
