package key

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
)

// file to write user's keys
var authKey string = path.Join(os.Getenv("HOME"), "authorized_keys")

// writes a key in authKey file
func addKey(key string) error {
	file, err := os.OpenFile(authKey, os.O_WRONLY, 0755)
	defer file.Close()
	if err != nil {
		return err
	}
	keys, err := ioutil.ReadFile(authKey)
	if err != nil {
		return err
	}
	content := key
	if len(keys) != 0 {
		content = fmt.Sprintf("%s\n%s", keys, key)
	}
	_, err = file.Write([]byte(content))
	if err != nil {
		return err
	}
	return nil
}
