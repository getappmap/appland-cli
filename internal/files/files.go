package files

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/applandinc/appland-cli/internal/config"
	"github.com/spf13/afero"
)

type Validator func(fi os.FileInfo) bool

func validateFile(fi os.FileInfo, validators []Validator) (valid bool) {
	valid = true
	for _, v := range validators {
		if !v(fi) {
			valid = false
		}
	}
	return
}

func loadDirectory(dirName string, scenarioFiles []string, validators []Validator) ([]string, error) {
	files, err := afero.ReadDir(config.GetFS(), dirName)
	if err != nil {
		return nil, err
	}

	for _, fi := range files {
		if !fi.Mode().IsRegular() {
			continue
		}
		if !strings.HasSuffix(fi.Name(), ".appmap.json") {
			continue
		}

		if !validateFile(fi, validators) {
			continue
		}

		scenarioFiles = append(scenarioFiles, filepath.Join(dirName, fi.Name()))
	}
	return scenarioFiles, nil
}

func FindAppMaps(paths []string, validators ...Validator) ([]string, error) {
	scenarioFiles := make([]string, 0, 10)
	for _, path := range paths {
		fi, err := config.GetFS().Stat(path)
		if err != nil {
			return nil, err
		}

		switch mode := fi.Mode(); {
		case mode.IsDir():
			scenarioFiles, err = loadDirectory(path, scenarioFiles, validators)
			if err != nil {
				return nil, err
			}
		case mode.IsRegular():
			if !validateFile(fi, validators) {
				continue
			}

			scenarioFiles = append(scenarioFiles, path)
		}
	}
	return scenarioFiles, nil
}
