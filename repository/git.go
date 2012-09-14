package repository

import (
    "fmt"
    "os/exec"
    "path"
)

var bareLocation = "/var/repositories"

func newBare(name string) error {
    cmd := exec.Command("git", "init", "--bare", path.Join(bareLocation, name))
    _, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("Could not create git bare repository: %s", err.Error())
    }
    return nil
}
