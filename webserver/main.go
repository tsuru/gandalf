package main

import (
	"flag"
	"github.com/bmizerany/pat"
	"github.com/timeredbull/gandalf/api"
	"log"
	"net/http"
)

func main() {
	dry := flag.Bool("dry", false, "dry-run: does not start the server (for testing purpose)")
	flag.Parse()

	router := pat.New()
	router.Post("/user", http.HandlerFunc(api.CreateUser))
	router.Post("/repository", http.HandlerFunc(api.CreateRepository))

	if !*dry {
		log.Fatal(http.ListenAndServe(":8080", router))
	}
}
