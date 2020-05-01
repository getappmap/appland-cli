package config

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

var sampleConfigData = []byte(`---
current_context: test
contexts:
  test:
    url: http://localhost:3000
    api_key: MY_API_KEY
`)

func TestWriteCLIConfig(t *testing.T) {
	setFileSystem(afero.NewMemMapFs())

	configPath = ".appland"
	config = &Config{
		CurrentContext: "default",
		Contexts: map[string]*Context{
			"default": &Context{
				URL:    "hostname.com",
				APIKey: "MY_API_KEY",
			},
		},
	}

	err := WriteCLIConfig()
	require.Nil(t, err)

	data, err := afero.ReadFile(fs, configPath)
	require.Nil(t, err)

	c := &Config{}
	err = yaml.Unmarshal(data, c)
	require.Nil(t, err)

	assert := assert.New(t)
	assert.Equal(config.CurrentContext, c.CurrentContext)
	assert.Equal(len(config.Contexts), len(c.Contexts))
	assert.Equal(config.Contexts["default"].URL, c.Contexts["default"].URL)
	assert.Equal(config.Contexts["default"].APIKey, c.Contexts["default"].APIKey)
}

func TestLoadCLIConfig(t *testing.T) {
	setFileSystem(afero.NewMemMapFs())

	afero.WriteFile(fs, ".appland", sampleConfigData, 0600)

	require.True(t, loadCLIConfig(".appland"))

	assert := assert.New(t)
	assert.Equal(GetCurrentContextName(), "test")
	assert.Equal(GetCurrentContext().APIKey, "MY_API_KEY")
	assert.Equal(GetCurrentContext().URL, "http://localhost:3000")
}

func TestMakeContext(t *testing.T) {
	setFileSystem(afero.NewMemMapFs())

	afero.WriteFile(fs, ".appland", sampleConfigData, 0600)

	require.True(t, loadCLIConfig(".appland"))
	require.Nil(t, MakeContext("new", "hostname.com"))
	require.Nil(t, SetCurrentContext("new"))
	assert.Equal(t, GetCurrentContext().URL, "hostname.com")
}
