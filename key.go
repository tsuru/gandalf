package gandalf

import (
	"os"
)

var authKeys string = "authorized_keys"

func addKey(key string) error {
	file, err := os.OpenFile(authKeys, os.O_WRONLY, 0755)
	defer file.Close()
	if err != nil {
		return err
	}
	_, err = file.Write([]byte(key))
	if err != nil {
		return err
	}
	return nil
}
