package key

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

// file to write user's keys
var authKey string = path.Join(os.Getenv("HOME"), "authorized_keys")

// Add writes a key in authKey file
func Add(key string) error {
	file, err := os.OpenFile(authKey, os.O_RDWR, 0755)
	defer file.Close()
	if err != nil {
		return err
	}
	keys, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}
	content := key
	if len(keys) != 0 {
		content = fmt.Sprintf("%s\n%s", keys, key)
	}
	_, err = file.Seek(0, 0)
	_, err = file.WriteString(content)
	if err != nil {
		return err
	}
	return nil
}

// Remove a key from auhtKey file
func Remove(key string) error {
	file, err := os.OpenFile(authKey, os.O_RDWR, 0755)
	defer file.Close()
	if err != nil {
		return err
	}
	keys, err := ioutil.ReadAll(file)
	content := strings.Replace(string(keys), key+"\n", "", -1)
	content = strings.Replace(content, key, "", -1)
	err = file.Truncate(0)
	_, err = file.Seek(0, 0)
	_, err = file.WriteString(content)
	if err != nil {
		return err
	}
	return nil
}
