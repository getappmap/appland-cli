package appland

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
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
	resp, err := client.httpClient.Post(url, "application/json", reader)
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
