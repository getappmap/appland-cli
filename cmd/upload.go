package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/applandinc/appland-cli/internal/appland"
	"github.com/applandinc/appland-cli/internal/config"
	"github.com/applandinc/appland-cli/internal/metadata"
	"github.com/pkg/browser"
	progressbar "github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

func init() {
	var (
		environment     string
		organization    string
		version         string
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

				// TODO
				// In the future, perhaps we do something a little more graceful than to
				// use the git metadata from the last file uploaded. It seems we
				// shouldn't ever be uploading files from multiple repositories under a
				// single mapset.
				var git *metadata.GitMetadata

				for i, path := range args {
					file, err := os.Open(path)
					if err != nil {
						fail(err)
					}

					data, err := ioutil.ReadAll(file)
					if err != nil {
						fail(err)
					}

					git, err = metadata.GetGitMetadata(path)
					if err != nil {
						progressBar.Clear()
						warn(fmt.Errorf("Could not collect git metadata (%w)", err))
						progressBar.RenderBlank()
					} else {
						gitPatch, err := git.AsPatch()
						if err != nil {
							fail(err)
						}

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

				mapSet := appland.BuildMapSet(appmapConfig.Application, scenarios).
					SetOrganization(organization).
					SetVersion(version).
					SetEnvironment(environment).
					WithGitMetadata(git)

				res, err := api.CreateMapSet(mapSet)
				if err != nil {
					fail(err)
				}

				progressBar.Finish()

				fmt.Printf("\n\nSuccess! %s has been updated with %d scenarios.\n", appmapConfig.Application, len(args))

				url := api.BuildUrl("applications", fmt.Sprintf("%d?mapset=%d", res.AppID, res.ID))
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
	uploadCmd.Flags().StringVarP(&version, "version", "v", "", "Set the MapSet version")
	uploadCmd.Flags().StringVarP(&environment, "environment", "e", "", "Set the MapSet environment")
	rootCmd.AddCommand(uploadCmd)
}
