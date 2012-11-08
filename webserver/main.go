package main

import (
	"flag"
	"fmt"
	"github.com/bmizerany/pat"
	"github.com/globocom/config"
	"github.com/globocom/gandalf/api"
	"log"
	"net/http"
	"os/exec"
)

func startGitDaemon() error {
	bLocation, err := config.GetString("bare-location")
	if err != nil {
		return err
	}
	basePath := fmt.Sprintf("--base-path=%s", bLocation)
	return exec.Command("git", "daemon", basePath, "--syslog").Run()
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
	router := pat.New()
	router.Post("/user", http.HandlerFunc(api.NewUser))
	router.Del("/user/:name", http.HandlerFunc(api.RemoveUser))
	router.Post("/repository", http.HandlerFunc(api.NewRepository))
	router.Del("/repository/:name", http.HandlerFunc(api.RemoveRepository))

	if !*dry {
		startGitDaemon()
		log.Fatal(http.ListenAndServe(":8080", router))
	}
}
