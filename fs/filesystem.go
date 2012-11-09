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
