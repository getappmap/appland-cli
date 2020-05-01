package config

import (
	"fmt"
	"os"
	"path"

	util "github.com/applandinc/appland-cli/internal/util"
	"github.com/spf13/afero"
	"gopkg.in/yaml.v2"
)

const appmapYaml = "appmap.yml"

type AppMapConfig struct {
	Application string          `yaml:"name"`
	Packages    []AppMapPackage `yaml:"packages"`
}

type AppMapPackage struct {
	Path    string   `yaml:"path"`
	Exclude []string `yaml:"exclude"`
}

func loadAppmapConfig(path string) (*AppMapConfig, error) {
	data, err := afero.ReadFile(getFS(), path)
	if err != nil {
		return nil, err
	}

	appmapConfig := &AppMapConfig{}
	if err := yaml.Unmarshal(data, appmapConfig); err != nil {
		return nil, err
	}

	return appmapConfig, nil
}

func LoadAppmapConfig(overridePath string, fallbackPath string) (*AppMapConfig, error) {
	if overridePath != "" {
		return loadAppmapConfig(overridePath)
	}

	currentDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	appmapPath := path.Join(currentDir, appmapYaml)
	if ok, _ := afero.Exists(getFS(), appmapPath); ok {
		return loadAppmapConfig(appmapPath)
	}

	if repoInfo, err := util.GetRepository(currentDir); err != nil {
		appmapPath = path.Join(repoInfo.Path, appmapYaml)
		if ok, _ := afero.Exists(getFS(), appmapPath); ok {
			return loadAppmapConfig(appmapPath)
		}
	}

	if fallbackPath != "" {
		if repoInfo, err := util.GetRepository(fallbackPath); err == nil {
			appmapPath = path.Join(repoInfo.Path, appmapYaml)
			if ok, _ := afero.Exists(getFS(), appmapPath); ok {
				return loadAppmapConfig(appmapPath)
			}
		}
	}

	return nil, fmt.Errorf("could not locate %s", appmapYaml)
}
