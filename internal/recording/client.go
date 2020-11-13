package recording

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

var (
	Client   HTTPClient
	endpoint string = "/_appmap/record"
)

func init() {
	Client = &http.Client{}
}

// Request - interacts with the AppMap Recording API
// url [string] url of website being recorded
// method [string] can be "POST", "DELETE", or "GET", otherwise get an error
// Returns the http response
func recordingRequest(url string, method string) (*http.Response, error) {
	path := url + endpoint
	request, err := http.NewRequest(method, path, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating request: %v\n", err)
		return nil, err
	}
	return Client.Do(request)
}

// Starts a new AppMap recording
// url [string] url of website being recorded
// Returns true if recording has started, false otherwise
func StartRecording(url string) (bool, error) {
	response, err := recordingRequest(url, "POST")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error handling request: %v\n", err)
		return false, err
	}
	defer response.Body.Close()
	switch response.StatusCode {
	case 200:
		fmt.Println("A new recording session has started")
		return true, nil
	case 409:
		fmt.Println("An existing recording session is already in progress")
		return false, nil
	default:
		fmt.Fprintf(os.Stderr, "Error: unexpected status code %d: %v", response.StatusCode, err)
		return false, err
	}
}

// Stops an active Appmap recording session
// url [string] url of website being recorded
// Returns true if recording has stopped, false otherwise
func StopRecording(url string) (bool, error) {
	response, err := recordingRequest(url, "DELETE")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error handling request: %v\n", err)
		return false, err
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading response body: %v\n", err)
		return false, err
	}
	switch response.StatusCode {
	case 200:
		fmt.Println("Current recording session has stopped")
		fmt.Println(string(body))
		return true, nil
	case 404:
		fmt.Println("No active recording session to stop")
		return false, nil
	default:
		fmt.Fprintf(os.Stderr, "Error: unexpected status code %d: %v", response.StatusCode, err)
		return false, err
	}
}

// Checks for an active recording session
// url [string] url of website being recorded
// returns the recording status
func CheckRecording(url string) (bool, error) {
	response, err := recordingRequest(url, "GET")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error handling request: %v\n", err)
		return false, err
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		fmt.Fprintf(os.Stderr, "Error: unexpected status code %d: %v", response.StatusCode, err)
		return false, err
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading response body: %v\n", err)
		return false, err
	}
	var jsonBody map[string]bool
	err = json.Unmarshal(body, &jsonBody)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error demarshalling json: %v\n", err)
		return false, err
	}
	if jsonBody["enabled"] {
		fmt.Println("Appmap recording is currently enabled")
		return true, nil
	} else {
		fmt.Println("Appmap recording is currently disabled")
		return false, nil
	}
}
