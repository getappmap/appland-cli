package appland

import (
	"encoding/json"
	"os"
	"fmt"
	"io"
	"io/ioutil"
	"bytes"
	"net/http"

	"github.com/applandinc/appland-cli/internal/config"
)

type Client struct {
	context    *config.Context
	httpClient *http.Client
}

type createScenarioAPIResponse struct {
	uuid string
}

type CreateScenarioResponse struct {
	BatchID string
	UUID    string
}

func (client *Client) newAuthRequest(method, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", client.context.APIKey))
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
	path := client.context.URL
	for i := 0; i < numPaths; i++ {
		path = path + fmt.Sprintf("/%v", paths[i])
	}
	return path
}

func (client *Client) CreateScenario(reader io.Reader, batchID string) (*CreateScenarioResponse, error) {
	url := client.BuildUrl("api", "scenarios")
	resp, err := client.post(url, reader)
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

	jsonResponse := &createScenarioAPIResponse{}
	if err := json.Unmarshal(body, jsonResponse); err != nil {
		return nil, err
	}

	returnedBatchID := batchID
	if returnedBatchID == "" {
		returnedBatchID = resp.Header.Get("AppLand-Scenario-Batch")
	}

	return &CreateScenarioResponse{
		BatchID: returnedBatchID,
		UUID:    jsonResponse.uuid,
	}, nil
}

func (client *Client) Login(login string, password string) error {
	url := client.BuildUrl("api", "api_key")

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

	client.context.APIKey = responseFormat.APIKey
	return nil
}

func (client *Client) DeleteAPIKey() error {
	url := client.BuildUrl("api", "api_key")
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

	client.context.APIKey = ""
	return nil
}
