package config

import (
	"fmt"
	"os"
	"path"

	"github.com/spf13/afero"
	"gopkg.in/yaml.v2"
)

type Config struct {
	CurrentContext string              `yaml:"current_context"`
	Contexts       map[string]*Context `yaml:"contexts"`
	dirty          bool
}

type Context struct {
	URL    string `yaml:"url"`
	APIKey string `yaml:"api_key"`
}

const (
	applandFilename    string = ".appland"
	defaultContextName string = "default"
)

var (
	config         *Config
	configPath     string
	currentContext *Context
	defaultContext = Context{
		URL:    "https://app.land",
		APIKey: "",
	}
)

func makeDefault() *Config {
	return &Config{
		CurrentContext: defaultContextName,
		Contexts: map[string]*Context{
			defaultContextName: &defaultContext,
		},
	}
}

func makeDirty() {
	if config == nil {
		return
	}

	config.dirty = true
}

func loadCLIConfig(path string) bool {
	info, err := getFS().Stat(path)
	if os.IsNotExist(err) {
		return false
	} else if err != nil {
		fmt.Fprintf(os.Stderr, "warn: %s\n", err)
		return false
	}

	if info.IsDir() {
		fmt.Fprintf(os.Stderr, "warn: %s is a directory\n", path)
		return false
	}

	data, err := afero.ReadFile(getFS(), path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warn: %s\v", err)
		return false
	}

	c := &Config{}
	err = yaml.Unmarshal(data, c)
	if err != nil {
		fmt.Fprintf(os.Stderr, "warn: %s\n", err)
		return false
	}

	config = c
	configPath = path

	if len(config.Contexts) == 0 {
		config = makeDefault()
	}

	return true
}

func LoadCLIConfig() {
	envPath := os.Getenv("APPLAND_CONFIG")
	if envPath != "" && loadCLIConfig(envPath) {
		return
	}

	currentDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "warn: %s\n", err)
	} else {
		if loadCLIConfig(path.Join(currentDir, applandFilename)) {
			return
		}
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintf(os.Stderr, "warn: %s\n", err)
	} else {
		if loadCLIConfig(path.Join(homeDir, applandFilename)) {
			return
		}
	}

	config = makeDefault()
	configPath = path.Join(homeDir, applandFilename)
}

func WriteCLIConfig() error {
	if config == nil {
		return fmt.Errorf("no config present")
	}

	if !config.dirty {
		return nil
	}

	if configPath == "" {
		return fmt.Errorf("no config path is set")
	}

	data, err := yaml.Marshal(config)
	if err != nil {
		return err
	}

	return afero.WriteFile(getFS(), configPath, data, 0600)
}

func GetAPIKey() string {
	if currentContext == nil {
		return ""
	}

	return ResolveValue(currentContext.APIKey)
}

func SetCurrentContext(name string) error {
	if _, err := GetContext(name); err != nil {
		return err
	}

	config.CurrentContext = name

	makeDirty()

	return nil
}

func GetContext(name string) (*Context, error) {
	if name == "" {
		return nil, fmt.Errorf("cannot retrieve unnamed context")
	}

	if context, ok := config.Contexts[name]; ok {
		return context, nil
	}

	return nil, fmt.Errorf("context '%s' does not exist", name)
}

func GetCurrentContext() (*Context, error) {
	return GetContext(config.CurrentContext)
}

func GetCurrentContextName() string {
	return config.CurrentContext
}

func RenameContext(old string, new string) {
	if old == new {
		return
	}

	config.Contexts[new] = config.Contexts[old]
	delete(config.Contexts, old)

	if config.CurrentContext == old {
		config.CurrentContext = new
	}

	makeDirty()
}

func MakeContext(name string, url string) error {
	if existingContext, _ := GetContext(name); existingContext != nil {
		return fmt.Errorf("a context named '%s' already exists", name)
	}

	config.Contexts[name] = &Context{
		URL: url,
	}

	makeDirty()

	return nil
}

func GetCLIConfig() *Config {
	return config
}

func (context *Context) GetAPIKey() string {
	if apiKey := os.Getenv("APPLAND_API_KEY"); apiKey != "" {
		return apiKey
	}

	return ResolveValue(context.APIKey)
}

func (context *Context) GetURL() string {
	if url := os.Getenv("APPLAND_URL"); url != "" {
		return url
	}

	return ResolveValue(context.URL)
}

func (context *Context) SetAPIKey(apiKey string) {
	if IsEnvironmentVariable(context.APIKey) {
		return
	}

	context.APIKey = apiKey

	makeDirty()
}

func (context *Context) SetURL(url string) {
	if IsEnvironmentVariable(context.URL) {
		return
	}

	context.URL = url

	makeDirty()
}

func (context *Context) GetName() (string, error) {
	if config == nil {
		return "", fmt.Errorf("no config present")
	}

	for name, c := range config.Contexts {
		if c == context {
			return name, nil
		}
	}

	return "", fmt.Errorf("context not found")
}

func (context *Context) SetVariable(key, value string) error {
	switch key {
	case "url":
		context.SetURL(value)
	case "api_key":
		context.SetAPIKey(key)
	case "name":
		name, err := context.GetName()
		if err != nil {
			return err
		}
		RenameContext(name, value)
	default:
		return fmt.Errorf("unknown key '%s'", key)
	}

	return nil
}
