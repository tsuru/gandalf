package key

import (
	"fmt"
	"github.com/globocom/config"
	"github.com/globocom/tsuru/fs"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

// file to write user's keys
var authKey string = path.Join(os.Getenv("HOME"), ".ssh", "authorized_keys")
var fsystem fs.Fs

// Writes `key` in authorized_keys file (from current user)
// It does not writes in the database, there is no need for that since the key
// object is embedded on the user's document
// should the fsystem abstraction be passed here as an argument?
// maybe it's not a good idea for api direct usage
func Add(key string, fsystem fs.Fs) error {
	file, err := fsystem.OpenFile(authKey, os.O_RDWR, 0755)
	defer file.Close()
	if err != nil {
		return err
	}
	keys, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}
	content := formatKey(key)
	if len(keys) != 0 {
		content = fmt.Sprintf("%s\n%s", keys, formatKey(key))
	}
	if _, err := file.Seek(0, 0); err != nil {
		return err
	}
	if _, err := file.WriteString(content); err != nil {
		return err
	}
	return nil
}

func BulkAdd(keys []string, fsystem fs.Fs) error {
	return bulkAction(Add, keys, fsystem)
}

func BulkRemove(keys []string, fsystem fs.Fs) error {
	return bulkAction(Remove, keys, fsystem)
}

// applies `action` into a bulk of keys
// this method does len(keys) io actions but we do not expect the user to have
// a LOT of keys, thus for now it is not a problem to do this extra io ops
func bulkAction(action func(string, fs.Fs) error, keys []string, fsystem fs.Fs) error {
	for _, k := range keys {
		err := action(k, fsystem)
		if err != nil {
			return err
		}
	}
	return nil
}

// Remove a key from auhtKey file
func Remove(key string, fsystem fs.Fs) error {
	file, err := fsystem.OpenFile(authKey, os.O_RDWR, 0755)
	defer file.Close()
	if err != nil {
		return err
	}
	keys, err := ioutil.ReadAll(file)
	key = formatKey(key)
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

func formatKey(key string) string {
	binPath, err := config.GetString("bin-path")
	if err != nil {
		panic(err)
	}
	keyTmpl := `no-port-forwarding,no-X11-forwarding,no-agent-forwarding,no-pty,command="%s" %s`
	return fmt.Sprintf(keyTmpl, binPath, key)
}
