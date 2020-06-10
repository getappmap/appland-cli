package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/applandinc/appland-cli/internal/appland"
	"github.com/applandinc/appland-cli/internal/config"
	"github.com/applandinc/appland-cli/internal/metadata"
	"github.com/applandinc/appland-cli/internal/util"
	"github.com/pkg/browser"
	progressbar "github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

func loadDirectory(dirName string, scenarioFiles []string) []string {
	files, err := ioutil.ReadDir(dirName)
	if err != nil {
		fail(err)
	}

	for _, fi := range files {
		if !fi.Mode().IsRegular() {
			continue
		}
		if !strings.HasSuffix(fi.Name(), ".appmap.json") {
			continue
		}

		scenarioFiles = append(scenarioFiles, filepath.Join(dirName, fi.Name()))
	}
	return scenarioFiles
}

func init() {
	var (
		branch          string
		environment     string
		application     string
		appmapPath      string
		version         string
		dontOpenBrowser bool

		uploadCmd = &cobra.Command{
			Use:   "upload [files, directories]",
			Short: "Upload AppMap files to AppLand",
			Args:  cobra.MinimumNArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				if application == "" {
					appmapConfig, err := config.LoadAppmapConfig(appmapPath, args[0])
					if err != nil {
						fail(fmt.Errorf("an appmap.yml should exist in the target repository or the --app / -a flag specified"))
					}

					application = appmapConfig.Application
				}

				// TODO
				// In the future, perhaps we do something a little more graceful than to
				// use the git metadata from the last file uploaded. It seems we
				// shouldn't ever be uploading files from multiple repositories under a
				// single mapset.
				var git *metadata.GitMetadata

				scenarioFiles := make([]string, 0, 10)
				for _, path := range args {
					fi, err := os.Stat(path)
					if err != nil {
						fail(err)
						return
					}

					switch mode := fi.Mode(); {
					case mode.IsDir():
						scenarioFiles = loadDirectory(path, scenarioFiles)
					case mode.IsRegular():
						scenarioFiles = append(scenarioFiles, path)
					}
				}

				scenarioUUIDs := make([]string, 0, len(scenarioFiles))
				progressBar := progressbar.New(len(scenarioFiles) + 1)
				progressBar.RenderBlank()

				for _, scenarioFile := range scenarioFiles {
					file, err := os.Open(scenarioFile)
					if err != nil {
						fail(err)
					}

					data, err := ioutil.ReadAll(file)
					if err != nil {
						fail(err)
					}

					git, err = metadata.GetGitMetadata(scenarioFile)
					if err != nil {
						util.Debugf("%w", err)
					}

					if err == nil && !git.IsEmpty() {
						gitPatch, err := git.AsPatch()
						if err != nil {
							fail(err)
						}

						data, err = gitPatch.Apply(data)
						if err != nil {
							fail(err)
						}
					}

					resp, err := api.CreateScenario(application, bytes.NewReader(data))
					if err != nil {
						fail(err)
					}

					scenarioUUIDs = append(scenarioUUIDs, resp.UUID)
					progressBar.Add(1)
				}

				// either both commit and branch are specified or both are unspecified
				// fail otherwise
				commitProvided := bool(git != nil && git.Commit != "")
				branchProvided := bool((git != nil && git.Branch != "") || branch != "")
				if commitProvided != branchProvided {
					progressBar.Clear()
					if commitProvided {
						fail(fmt.Errorf("Git branch could not be resolved\nRun again with the --branch or -b flag specified"))
					} else {
						fail(fmt.Errorf("The --branch or -b flag can only be provided when uploading appmaps from within a Git repository"))
					}
				}

				mapSet := appland.BuildMapSet(application, scenarioUUIDs).
					SetVersion(version).
					SetEnvironment(environment).
					WithGitMetadata(git).
					SetBranch(branch)

				res, err := api.CreateMapSet(mapSet)
				if err != nil {
					fail(err)
				}

				progressBar.Finish()

				fmt.Printf("\n\nSuccess! %s has been updated with %d scenarios.\n", application, len(scenarioUUIDs))

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
	uploadCmd.Flags().StringVarP(&application, "app", "a", "", "Override the owning application")
	uploadCmd.Flags().StringVar(&appmapPath, "f", "", "Specify an appmap.yml path")
	uploadCmd.Flags().StringVarP(&branch, "branch", "b", "", "Set the MapSet branch if it's otherwise unavailable from Git")
	uploadCmd.Flags().StringVarP(&version, "version", "v", "", "Set the MapSet version")
	uploadCmd.Flags().StringVarP(&environment, "environment", "e", "", "Set the MapSet environment")
	rootCmd.AddCommand(uploadCmd)
}
