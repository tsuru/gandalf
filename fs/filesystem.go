// fs is just a filesystem wrapper.
// It makes use of tsuru/fs pkg.
package fs

import (
	"github.com/globocom/tsuru/fs"
)

var Fsystem fs.Fs

func Filesystem() fs.Fs {
	if Fsystem == nil {
		return fs.OsFs{}
	}
	return Fsystem
}
