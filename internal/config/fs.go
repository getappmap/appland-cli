package config

import (
	"github.com/spf13/afero"
)

var fs = afero.NewOsFs()

func SetFileSystem(filesystem afero.Fs) {
	fs = filesystem
}

func GetFS() afero.Fs {
	return fs
}
