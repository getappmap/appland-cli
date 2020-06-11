package appland

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/applandinc/appland-cli/internal/config"
	"github.com/applandinc/appland-cli/internal/metadata"
	jsonpatch "github.com/evanphx/json-patch"
)

type HttpError struct {
	Status int
}

func (e *HttpError) Error() string { return http.StatusText(e.Status) }
func (e *HttpError) Is(target error) bool {
	t, ok := target.(*HttpError)
	if !ok {
		return false
	}
	return e.Status == t.Status
}

type Client interface {
	BuildUrl(paths ...interface{}) string
	Context() *config.Context
	CreateMapSet(mapset *MapSet) (*CreateMapSetResponse, error)
	CreateScenario(org string, scenarioData io.Reader) (*ScenarioResponse, error)
	GetScenario(id int) (*ScenarioResponse, error)
	DeleteAPIKey() error
	Login(login string, password string) error
	TestAPIKey(apiKey string) (bool, error)
}

type clientImpl struct {
	context    *config.Context
	httpClient *http.Client
}

func (client *clientImpl) Context() *config.Context {
	return client.context
}

type CreateMapSetResponse struct {
	ID    uint32 `json:"id"`
	AppID uint32 `json:"app_id"`
}

type MapSet struct {
	Application string   `json:"app,omitempty"`
	Commit      string   `json:"commit,omitempty"`
	Branch      string   `json:"branch,omitempty"`
	Version     string   `json:"version,omitempty"`
	Environment string   `json:"environment,omitempty"`
	Scenarios   []string `json:"scenarios,omitempty"`
}

func BuildMapSet(application string, scenarios []string) *MapSet {
	return &MapSet{
		Application: application,
		Scenarios:   scenarios,
	}
}

func (mapset *MapSet) SetEnvironment(environment string) *MapSet {
	mapset.Environment = environment
	return mapset
}

func (mapset *MapSet) SetVersion(version string) *MapSet {
	mapset.Version = version
	return mapset
}

func (mapset *MapSet) SetBranch(branch string) *MapSet {
	if branch == "" {
		return mapset
	}

	if mapset.Branch != "" && mapset.Branch != branch {
		fmt.Fprintf(os.Stderr, "warn: current branch differs from override (%s != %s)", mapset.Branch, branch)
	}

	mapset.Branch = branch
	return mapset
}

func (mapset *MapSet) WithGitMetadata(git *metadata.Git) *MapSet {
	if git != nil {
		mapset.Branch = git.Branch
		mapset.Commit = git.Commit
	}
	return mapset
}

type createScenarioRequest struct {
	Data string `json:"data,omitempty"`
}

type ScenarioResponse struct {
	UUID string
}

func (client *clientImpl) newAuthRequest(method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", client.context.GetAPIKey()))
	req.Header.Add("Content-Type", "application/json")

	return req, nil
}

func (client *clientImpl) post(url string, body io.Reader) (*http.Response, error) {
	req, err := client.newAuthRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}

	return client.httpClient.Do(req)
}

func (client *clientImpl) get(url string, body io.Reader) (*http.Response, error) {
	req, err := client.newAuthRequest(http.MethodGet, url, body)
	if err != nil {
		return nil, err
	}

	return client.httpClient.Do(req)
}

func (client *clientImpl) delete(url string, body io.Reader) (*http.Response, error) {
	req, err := client.newAuthRequest(http.MethodDelete, url, body)
	if err != nil {
		return nil, err
	}

	return client.httpClient.Do(req)
}

func MakeClient(context *config.Context) Client {
	return &clientImpl{
		context:    context,
		httpClient: http.DefaultClient,
	}
}

func (client *clientImpl) BuildUrl(paths ...interface{}) string {
	numPaths := len(paths)
	path := client.context.GetURL()
	for i := 0; i < numPaths; i++ {
		path = path + fmt.Sprintf("/%v", paths[i])
	}
	return path
}

func (client *clientImpl) CreateMapSet(mapset *MapSet) (*CreateMapSetResponse, error) {
	data, err := json.Marshal(mapset)
	if err != nil {
		return nil, err
	}

	url := client.BuildUrl("api", "mapsets")
	resp, err := client.post(url, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("got status %d:\n%s", resp.StatusCode, string(body))
	}

	responseObj := &CreateMapSetResponse{}
	if err = json.Unmarshal(body, responseObj); err != nil {
		return nil, err
	}

	return responseObj, nil
}

func (client *clientImpl) CreateScenario(app string, scenarioData io.Reader) (*ScenarioResponse, error) {
	scenarioBytes, err := ioutil.ReadAll(scenarioData)
	if err != nil {
		return nil, err
	}

	appPatch := []byte(fmt.Sprintf(`{"metadata": { "app": "%s" }}`, app))
	scenarioBytes, err = jsonpatch.MergePatch(scenarioBytes, appPatch)
	if err != nil {
		return nil, err
	}

	requestObj := &createScenarioRequest{
		Data: string(scenarioBytes),
	}

	data, err := json.Marshal(requestObj)
	if err != nil {
		return nil, err
	}

	url := client.BuildUrl("api", "scenarios")
	resp, err := client.post(url, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("got status %d:\n%s", resp.StatusCode, string(body))
	}

	responseObj := &ScenarioResponse{}
	if err := json.Unmarshal(body, responseObj); err != nil {
		return nil, err
	}

	return responseObj, nil
}

func (client *clientImpl) GetScenario(id int) (*ScenarioResponse, error) {
	url := client.BuildUrl("api", "scenarios", "0")
	resp, err := client.get(url, nil)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		httpError := &HttpError{resp.StatusCode}
		return nil, fmt.Errorf("GetScenario failed, %w", httpError)
	}

	responseObj := &ScenarioResponse{}
	if err := json.Unmarshal(body, responseObj); err != nil {
		return nil, err
	}

	return responseObj, nil

}

func (client *clientImpl) Login(login string, password string) error {
	url := client.BuildUrl("api", "api_keys")

	hostname, err := os.Hostname()
	if err != nil {
		hostname = "appland CLI"
	}

	requestFormat := struct {
		Description string `json:"description"`
	}{
		Description: hostname,
	}

	requestData, err := json.Marshal(&requestFormat)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(requestData))
	if err != nil {
		return err
	}

	req.SetBasicAuth(login, password)
	req.Header.Add("Content-Type", "application/json")

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return err
	}

	responseFormat := struct {
		APIKey string `json:"api_key"`
	}{}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf(string(data))
	}

	err = json.Unmarshal(data, &responseFormat)
	if err != nil {
		return err
	}

	client.context.SetAPIKey(responseFormat.APIKey)
	return nil
}

func (client *clientImpl) DeleteAPIKey() error {
	url := client.BuildUrl("api", "api_keys")
	resp, err := client.delete(url, nil)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		msg, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		return fmt.Errorf(string(msg))
	}

	client.context.SetAPIKey("")
	return nil
}

func (client *clientImpl) TestAPIKey(apiKey string) (bool, error) {
	// Test the api key by making a request with an invalid scenario
	// id. If the response is NotFound, the API key is valid. If the
	// response is Unauthorized, the API key is invalid. Any other
	// response is an error.
	testContext := &config.Context{client.context.URL, apiKey}
	testApi := MakeClient(testContext)

	_, err := testApi.GetScenario(0)
	if err == nil {
		// Shouldn't ever actually find the scenario, though.
		panic(fmt.Sprintf("Found scenario with id 0?"))
	}
	if errors.Is(err, &HttpError{Status: http.StatusNotFound}) {
		return true, nil
	} else if errors.Is(err, &HttpError{Status: http.StatusUnauthorized}) {
		return false, nil
	}

	return false, err
}
