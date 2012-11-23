package user

import (
	"fmt"
	"github.com/globocom/config"
	"github.com/globocom/gandalf/fs"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

// file to write user's keys
var authKey string = path.Join(os.Getenv("HOME"), ".ssh", "authorized_keys")

// Writes `key` in authorized_keys file (from current user)
// It does not writes in the database, there is no need for that since the key
// object is embedded on the user's document
func addKey(k, username string) error {
	file, err := fs.Filesystem().OpenFile(authKey, os.O_RDWR|os.O_EXCL, 0755)
	defer file.Close()
	if err != nil {
		return err
	}
	keys, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}
	content := formatKey(k, username)
	if len(keys) != 0 {
		content = fmt.Sprintf("%s\n%s", keys, content)
	}
	if _, err := file.Seek(0, 0); err != nil {
		return err
	}
	if _, err := file.WriteString(content); err != nil {
		return err
	}
	return nil
}

func addKeys(keys map[string]string, username string) error {
	for _, k := range keys {
		err := addKey(k, username)
		if err != nil {
			return err
		}
	}
	return nil
}

func removeKeys(keys map[string]string, username string) error {
	for _, k := range keys {
		err := removeKey(k, username)
		if err != nil {
			return err
		}
	}
	return nil
}

// removes a key from auhtKey file
func removeKey(key, username string) error {
	file, err := fs.Filesystem().OpenFile(authKey, os.O_RDWR|os.O_EXCL, 0755)
	defer file.Close()
	if err != nil {
		return err
	}
	keys, err := ioutil.ReadAll(file)
	key = formatKey(key, username)
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

func formatKey(key, username string) string {
	binPath, err := config.GetString("bin-path")
	if err != nil {
		panic(err)
	}
	keyTmpl := `no-port-forwarding,no-X11-forwarding,no-agent-forwarding,no-pty,command="%s %s" %s`
	return fmt.Sprintf(keyTmpl, binPath, username, key)
}

func mergeMaps(x, y map[string]string) map[string]string {
	for k, v := range y {
		x[k] = v
	}
	return x
}
