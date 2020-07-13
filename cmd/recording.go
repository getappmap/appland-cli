package cmd

import (
	"fmt"
	"io/ioutil"
  	"encoding/json"
  	"net/http"
	"os"

  	"github.com/spf13/cobra"
)

var path string = "/_appmap/record"

func init() {
	var (
		recordingCmd = &cobra.Command {
			Use:   "appmap",
      		Short: "Manage AppMap recordings",
      		Run: func(cmd *cobra.Command, args []string) {
				cmd.Usage()
			},
		}
		// Starts a new AppMap recording
		// Status 200: new recording session was started successfully
		// Status 409: an exisiting recording session is already in progress
		// Returns no body
		recordingStartCmd = &cobra.Command {
			Use:   "start [url]",
			Short: "Start a new AppMap recording session",
			Args:  cobra.ExactArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				url := args[0] + path
                response, err := http.Post(url,"",nil)
                if err != nil {
					fail(err)
                }
				// best practice to defer close response.body
				// what if no body exists?
				defer response.Body.Close()

                switch response.StatusCode {
                case 200:
                	fmt.Println("A new recording session has started")
                case 409:
                	fmt.Println("An existing recording session is already in progress")
                default:
					fail(fmt.Errorf("An unexpected error occured, status code: %v", response.StatusCode))
				}
			},
		}
		// Stops an active Appmap recording session and returns the Appmap json to stdout
		// Status 200: recording session was stopped successfully and response body contains Appmap json
		// Status 404 if there was no active recording session to be stopped
		recordingStopCmd = &cobra.Command {
			Use:   "stop [url]",
			Short: "Stop an existing AppMap recording session",
			Args:  cobra.ExactArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
                url := args[0] + path
				client := &http.Client{}

                request, err := http.NewRequest("DELETE", url, nil)
                if err != nil {
					fail(err)
                }
                response, err := client.Do(request)
                if err != nil {
					fail(err)
                }
                defer response.Body.Close()

                switch response.StatusCode {
                case 200:
					fmt.Fprintln(os.Stderr, "Current recording session has stopped")
                case 404:
					fmt.Fprintln(os.Stderr, "No active recording session to stop")
                default:
					fail(fmt.Errorf("An unexpected error occured, Status Code: %v", response.StatusCode))
                }
                body, err := ioutil.ReadAll(response.Body)
                if err != nil {
					fail(err)
                }
                fmt.Println(string(body))
			},
		}
		// Retrieves the current Appmap recording status (enabled or disabled)
		// Returns a json body {"enabled": bool}
		recordingCheckCmd = &cobra.Command {
			Use:   "check [url]",
      		Short: "Check the current AppMap recording status",
      		Args:  cobra.ExactArgs(1),
      		Run: func(cmd *cobra.Command, args []string) {
				url := args[0] + path
                response, err := http.Get(url)
                if err != nil {
					fail(err)
                }
				defer response.Body.Close()

                if response.StatusCode != 200 {
					fail(fmt.Errorf("An unexpected error occured, Status Code: %v", response.StatusCode))
                }
                body, err := ioutil.ReadAll(response.Body)
                if err != nil {
					fail(err)
                }
				var jsonBody map[string]bool
                err = json.Unmarshal(body, &jsonBody)
                if err != nil {
					fail(err)
                }
                if jsonBody["enabled"] {
                	fmt.Println("Appmap recording is currently enabled")
                } else {
                	fmt.Println("Appmap recording is currently disabled")
                }
			},
		}
	)
	rootCmd.AddCommand(recordingCmd)
  	recordingCmd.AddCommand(recordingStartCmd)
  	recordingCmd.AddCommand(recordingStopCmd)
  	recordingCmd.AddCommand(recordingCheckCmd)
}
