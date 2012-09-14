package repository

import (
	"fmt"
    "github.com/timeredbull/config"
	"os/exec"
	"path"
)

var bare string

func newBare(name string) error {
	cmd := exec.Command("git", "init", "--bare", path.Join(bareLocation(), name))
	_, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("Could not create git bare repository: %s", err.Error())
	}
	return nil
}

func bareLocation() string {
    if bare != "" {
        return bare
    }
    var err error
    bare, err = config.GetString("bare-location")
    if err != nil {
        panic("You should configure a bare-location for gandalf.")
    }
    return bare
}
