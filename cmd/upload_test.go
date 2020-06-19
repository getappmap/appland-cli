package cmd

import (
	"bytes"
	"io"
	"testing"

	"github.com/applandinc/appland-cli/internal/appland"
	"github.com/applandinc/appland-cli/internal/config"
	"github.com/applandinc/appland-cli/internal/metadata"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	appmapYml   = `name: myorg/myapp`
	validAppmap = `
	{
		"metadata": {},
		"events": [],
		"classMap": []
	}
	`
	validAppmapWithMetadata       = `{"classMap":[],"events":[],"metadata":{"git":{"annotated_tag":"0.0.0","branch":"master","commit":"76c0ae55fff17ae52ab67a0ff61e1af3d1157555","repository":"repo.git"}}}`
	validAppmapWithBranchOverride = `{"classMap":[],"events":[],"metadata":{"git":{"annotated_tag":"0.0.0","branch":"my-branch","commit":"76c0ae55fff17ae52ab67a0ff61e1af3d1157555","repository":"repo.git"}}}`
)

type MockGitProvider struct {
	mock.Mock
	metadata.GitProvider
}

func (m *MockGitProvider) Get(path string) (metadata.Metadata, error) {
	args := m.Called(path)
	data, _ := args.Get(0).(metadata.Metadata)
	return data, args.Error(1)
}

func (m *MockClient) CreateMapSet(mapset *appland.MapSet) (*appland.CreateMapSetResponse, error) {
	args := m.Called(mapset)
	resp, _ := args.Get(0).(*appland.CreateMapSetResponse)
	return resp, args.Error(1)
}

func (m *MockClient) CreateScenario(app string, scenarioData io.Reader) (*appland.ScenarioResponse, error) {
	args := m.Called(app, scenarioData)
	resp, _ := args.Get(0).(*appland.ScenarioResponse)
	return resp, args.Error(1)
}

func (m *MockClient) BuildUrl(paths ...interface{}) string {
	args := m.Called(paths)
	return args.String(0)
}

func (m *MockClient) Context() *config.Context {
	args := m.Called()
	context, _ := args.Get(0).(*config.Context)
	return context
}

func TestUploadSingleAppMap(t *testing.T) {
	fs := afero.NewMemMapFs()
	config.SetFileSystem(fs)

	fileName := "example.appmap.json"
	afero.WriteFile(fs, fileName, []byte(validAppmap), 0755)
	afero.WriteFile(fs, "appmap.yml", []byte(appmapYml), 0755)

	mockClient := &MockClient{}
	api = mockClient

	mockClient.
		On("CreateScenario", "myorg/myapp", bytes.NewReader([]byte(validAppmap))).
		Return(&appland.ScenarioResponse{UUID: "uuid"}, nil)

	mockClient.
		On("CreateMapSet", &appland.MapSet{Application: "myorg/myapp", Scenarios: []string{"uuid"}}).
		Return(&appland.CreateMapSetResponse{ID: 1, AppID: 1}, nil)

	mockClient.
		On("BuildUrl", []interface{}{"applications", "1?mapset=1"}).
		Return("http://example/applications/1?mapset=1")

	cmd := NewUploadCommand(&UploadOptions{appmapPath: "appmap.yml", dontOpenBrowser: true}, []metadata.Provider{})
	assert.Nil(t, cmd.RunE(cmd, []string{fileName}))
}

func TestUploadWithGitMetadata(t *testing.T) {
	fs := afero.NewMemMapFs()
	config.SetFileSystem(fs)

	fileName := "example.appmap.json"
	afero.WriteFile(fs, fileName, []byte(validAppmap), 0755)
	afero.WriteFile(fs, "appmap.yml", []byte(appmapYml), 0755)

	gitMetadata := &metadata.Git{
		Commit:     "76c0ae55fff17ae52ab67a0ff61e1af3d1157555",
		Branch:     "master",
		Tag:        "0.0.0",
		Repository: "repo.git",
	}

	mockGitProvider := &MockGitProvider{}
	mockGitProvider.
		On("Get", fileName).
		Return(gitMetadata, nil)

	mockClient := &MockClient{}
	mockClient.
		On("CreateScenario", "myorg/myapp", bytes.NewReader([]byte(validAppmapWithMetadata))).
		Return(&appland.ScenarioResponse{UUID: "uuid"}, nil)

	mockClient.
		On("CreateMapSet", &appland.MapSet{
			Application: "myorg/myapp",
			Scenarios:   []string{"uuid"},
			Commit:      gitMetadata.Commit,
			Branch:      gitMetadata.Branch,
		}).
		Return(&appland.CreateMapSetResponse{ID: 1, AppID: 1}, nil)

	mockClient.
		On("BuildUrl", []interface{}{"applications", "1?mapset=1"}).
		Return("http://example/applications/1?mapset=1")

	api = mockClient

	providers := []metadata.Provider{mockGitProvider}
	cmd := NewUploadCommand(&UploadOptions{appmapPath: "appmap.yml", dontOpenBrowser: true}, providers)
	assert.Nil(t, cmd.RunE(cmd, []string{fileName}))
}

func TestUploadWithBranchOverride(t *testing.T) {
	fs := afero.NewMemMapFs()
	config.SetFileSystem(fs)

	fileName := "example.appmap.json"
	afero.WriteFile(fs, fileName, []byte(validAppmap), 0755)
	afero.WriteFile(fs, "appmap.yml", []byte(appmapYml), 0755)

	gitMetadata := &metadata.Git{
		Commit:     "76c0ae55fff17ae52ab67a0ff61e1af3d1157555",
		Branch:     "master",
		Tag:        "0.0.0",
		Repository: "repo.git",
	}

	mockGitProvider := &MockGitProvider{}
	mockGitProvider.
		On("Get", fileName).
		Return(gitMetadata, nil)

	mockClient := &MockClient{}
	mockClient.
		On("CreateScenario", "myorg/myapp", bytes.NewReader([]byte(validAppmapWithBranchOverride))).
		Return(&appland.ScenarioResponse{UUID: "uuid"}, nil)

	branchOverride := "my-branch"
	mockClient.
		On("CreateMapSet", &appland.MapSet{
			Application: "myorg/myapp",
			Scenarios:   []string{"uuid"},
			Commit:      gitMetadata.Commit,
			Branch:      branchOverride,
		}).
		Return(&appland.CreateMapSetResponse{ID: 1, AppID: 1}, nil)

	mockClient.
		On("BuildUrl", []interface{}{"applications", "1?mapset=1"}).
		Return("http://example/applications/1?mapset=1")

	api = mockClient

	providers := []metadata.Provider{mockGitProvider}
	options := &UploadOptions{
		appmapPath:      "appmap.yml",
		dontOpenBrowser: true,
		branch:          branchOverride,
	}
	cmd := NewUploadCommand(options, providers)
	assert.Nil(t, cmd.RunE(cmd, []string{fileName}))
}
