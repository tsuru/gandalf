package key

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
func Add(key string, username string) error {
	file, err := fs.Filesystem().OpenFile(authKey, os.O_RDWR|os.O_EXCL, 0755)
	defer file.Close()
	if err != nil {
		return err
	}
	keys, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}
	content := formatKey(key, username)
	if len(keys) != 0 {
		content = fmt.Sprintf("%s\n%s", keys, formatKey(key, username))
	}
	if _, err := file.Seek(0, 0); err != nil {
		return err
	}
	if _, err := file.WriteString(content); err != nil {
		return err
	}
	return nil
}

func BulkAdd(keys []string, username string) error {
	return bulkAction(Add, keys, username)
}

func BulkRemove(keys []string, username string) error {
	return bulkAction(Remove, keys, username)
}

// applies `action` into a bulk of keys
// this method does len(keys) io actions but we do not expect the user to have
// a LOT of keys, thus for now it is not a problem to do this extra io ops
func bulkAction(action func(string, string) error, keys []string, username string) error {
	for _, k := range keys {
		err := action(k, username)
		if err != nil {
			return err
		}
	}
	return nil
}

// Remove a key from auhtKey file
func Remove(key, username string) error {
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
