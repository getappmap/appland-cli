package appland

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/applandinc/appland-cli/internal/config"
)

type Client struct {
	context    *config.Context
	httpClient *http.Client
}

type CreateMapSetResponse struct {
	ID    uint32 `json:"id"`
	AppID uint32 `json:"app_id"`
}

type createMapSetRequest struct {
	Organization string   `json:"org"`
	Application  string   `json:"app"`
	Scenarios    []string `json:"scenarios"`
}

type createScenarioRequest struct {
	Organization string `json:"org"`
	Data         string `json:"data"`
}

type CreateScenarioResponse struct {
	UUID string
}

func (client *Client) newAuthRequest(method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", client.context.GetAPIKey()))
	req.Header.Add("Content-Type", "application/json")

	return req, nil
}

func (client *Client) post(url string, body io.Reader) (*http.Response, error) {
	req, err := client.newAuthRequest(http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}

	return client.httpClient.Do(req)
}

func (client *Client) get(url string, body io.Reader) (*http.Response, error) {
	req, err := client.newAuthRequest(http.MethodGet, url, body)
	if err != nil {
		return nil, err
	}

	return client.httpClient.Do(req)
}

func (client *Client) delete(url string, body io.Reader) (*http.Response, error) {
	req, err := client.newAuthRequest(http.MethodDelete, url, body)
	if err != nil {
		return nil, err
	}

	return client.httpClient.Do(req)
}

func MakeClient(context *config.Context) *Client {
	return &Client{
		context:    context,
		httpClient: http.DefaultClient,
	}
}

func (client *Client) BuildUrl(paths ...interface{}) string {
	numPaths := len(paths)
	path := client.context.GetURL()
	for i := 0; i < numPaths; i++ {
		path = path + fmt.Sprintf("/%v", paths[i])
	}
	return path
}

func (client *Client) CreateMapSet(app, org string, scenarios []string) (*CreateMapSetResponse, error) {
	requestObj := &createMapSetRequest{
		Organization: org,
		Application:  app,
		Scenarios:    scenarios,
	}

	data, err := json.Marshal(requestObj)
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

func (client *Client) CreateScenario(org string, scenarioData io.Reader) (*CreateScenarioResponse, error) {
	scenarioBytes, err := ioutil.ReadAll(scenarioData)
	if err != nil {
		return nil, err
	}

	requestObj := &createScenarioRequest{
		Organization: org,
		Data:         string(scenarioBytes),
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

	responseObj := &CreateScenarioResponse{}
	if err := json.Unmarshal(body, responseObj); err != nil {
		return nil, err
	}

	return responseObj, nil
}

func (client *Client) Login(login string, password string) error {
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

func (client *Client) DeleteAPIKey() error {
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
