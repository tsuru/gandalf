package repository

import (
    "github.com/timeredbull/commandmocker"
    "testing"
    "os"
)

func TestCreateBareShouldCreateADir(t *testing.T) {
    dir, err := commandmocker.Add("git", "$*")
    if err != nil {
        t.Errorf(`Unpexpected error while mocking git cmd: %s`, err.Error())
        t.FailNow()
    }
    defer commandmocker.Remove(dir)
    err = newBare("myBare")
    if err != nil {
        t.Errorf(`Unexpected error while creating bare: %s`, err.Error())
    }
    if !commandmocker.Ran(dir) {
        t.Errorf("Expected newBare to call git")
    }
}

func TestCreateBareShouldReturnMeaningfullErrorWhenBareCreationFails(t *testing.T) {
    dir, err := commandmocker.Error("git", "ooooi", 1)
    if err != nil {
        t.Errorf(`Unexpected error while mocking git cmd`)
    }
    defer commandmocker.Remove(dir)
    err = newBare("foo")
    if err == nil {
        t.Errorf(`Expected error on git bare creation`)
    }
    got := err.Error()
    expected := "Could not create git bare repository: exit status 1"
    if got != expected {
        t.Errorf(`Expected error to be "%s", got "%s"`, expected, got)
    }
}
