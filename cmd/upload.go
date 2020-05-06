package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/applandinc/appland-cli/internal/config"
	"github.com/applandinc/appland-cli/internal/metadata"
	"github.com/pkg/browser"
	progressbar "github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

func init() {
	var (
		organization    string
		dontOpenBrowser bool

		uploadCmd = &cobra.Command{
			Use:   "upload [files]",
			Short: "Upload AppMap files to AppLand",
			Args:  cobra.MinimumNArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				appmapConfig, err := config.LoadAppmapConfig("", args[0])
				if err != nil {
					fail(err)
				}

				scenarios := make([]string, len(args))
				progressBar := progressbar.New(len(args) + 1)
				progressBar.RenderBlank()

				for i, path := range args {
					file, err := os.Open(path)
					if err != nil {
						fail(err)
					}

					data, err := ioutil.ReadAll(file)
					if err != nil {
						fail(err)
					}

					gitPatch, err := metadata.GetGitMetadata(path)
					if err != nil {
						progressBar.Clear()
						warn(fmt.Errorf("could not collect git metadata (%w)", err))
						progressBar.RenderBlank()
					} else {
						data, err = gitPatch.Apply(data)
						if err != nil {
							fail(err)
						}
					}

					resp, err := api.CreateScenario(organization, bytes.NewReader(data))
					if err != nil {
						fail(err)
					}

					scenarios[i] = resp.UUID
					progressBar.Add(1)
				}

				mapSet, err := api.CreateMapSet(appmapConfig.Application, organization, scenarios)
				if err != nil {
					fail(err)
				}

				progressBar.Finish()

				fmt.Printf("\n\nSuccess! %s has been updated with %d scenarios.\n", appmapConfig.Application, len(args))

				url := api.BuildUrl("applications", fmt.Sprintf("%d?mapset=%d", mapSet.AppID, mapSet.ID))
				if dontOpenBrowser {
					fmt.Println(url)
				} else {
					browser.OpenURL(url)
				}
			},
		}
	)

	uploadCmd.Flags().BoolVar(&dontOpenBrowser, "no-open", false, "Do not open the browser after a successful upload")
	uploadCmd.Flags().StringVarP(&organization, "org", "o", "", "Override the owning organization")
	rootCmd.AddCommand(uploadCmd)
}
