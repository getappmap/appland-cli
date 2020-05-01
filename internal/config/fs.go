package config

import (
	"github.com/spf13/afero"
)

var fs = afero.NewOsFs()

func setFileSystem(filesystem afero.Fs) {
	fs = filesystem
}

func getFS() afero.Fs {
	return fs
}
