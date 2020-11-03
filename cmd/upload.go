package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/applandinc/appland-cli/internal/appland"
	"github.com/applandinc/appland-cli/internal/config"
	"github.com/applandinc/appland-cli/internal/files"
	"github.com/applandinc/appland-cli/internal/metadata"
	"github.com/applandinc/appland-cli/internal/util"
	"github.com/pkg/browser"
	progressbar "github.com/schollz/progressbar/v3"
	"github.com/spf13/cobra"
)

const fileSizeLimit = 1024 * 1024 * 2

func checkSize(info os.FileInfo) error {
	if info.Size() > fileSizeLimit {
		return fmt.Errorf("file %s size is %d KiB, which is greater than the size limit of %d KiB, use --force if you want to upload it anyway", info.Name(), info.Size()/1024, fileSizeLimit/1024)
	}
	return nil
}

type UploadOptions struct {
	bench           bool
	branch          string
	environment     string
	application     string
	appmapPath      string
	force           bool
	version         string
	dontOpenBrowser bool
}

func NewUploadCommand(options *UploadOptions, metadataProviders []metadata.Provider) *cobra.Command {
	return &cobra.Command{
		Use:   "upload [files, directories]",
		Short: "Upload AppMap files to AppLand",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			application := options.application
			if application == "" {
				appmapConfig, err := config.LoadAppmapConfig(options.appmapPath, args[0])
				if err != nil {
					return fmt.Errorf("an appmap.yml should exist in the target repository or the --app / -a flag specified")
				}

				application = appmapConfig.Application
			}

			// If we encounter a usage related error later on we can disable this flag
			// We don't want to report the usage for errors which are unrelated to
			// command line arguments
			cmd.SilenceUsage = true

			// TODO
			// In the future, perhaps we do something a little more graceful than to
			// use the git metadata from the last file uploaded. It seems we
			// shouldn't ever be uploading files from multiple repositories under a
			// single mapset.
			var git *metadata.Git

			validator := func(fi os.FileInfo) bool {
				if !options.force {
					if err := checkSize(fi); err != nil {
						warn(err)
						return false
					}
				}
				return true
			}

			scenarioFiles, err := files.FindAppMaps(args, validator)
			if err != nil {
				return fmt.Errorf("failed finding AppMaps: %w", err)
			}

			scenarioUUIDs := make([]string, 0, len(scenarioFiles))
			progressBar := progressbar.New(len(scenarioFiles) + 1)
			progressBar.RenderBlank()

			timing := util.NewTiming("total")

			for _, scenarioFile := range scenarioFiles {
				fileTiming := timing.Start(scenarioFile)

				file, err := config.GetFS().Open(scenarioFile)
				if err != nil {
					return fmt.Errorf("failed opening %s: %w", scenarioFile, err)
				}

				fileTiming.Start("reading")

				data, err := ioutil.ReadAll(file)
				if err != nil {
					return fmt.Errorf("failed reading %s: %w", scenarioFile, err)
				}
				file.Close()

				fileTiming.Start("patching")
				for _, provider := range metadataProviders {
					m, err := provider.Get(scenarioFile)
					if err != nil {
						util.Debugf("%w", err)
					}

					if gitMetadata, ok := m.(*metadata.Git); ok {
						if options.branch != "" {
							gitMetadata.Branch = options.branch
						}

						git = gitMetadata
					}

					if err == nil && m.IsValid() {
						patch, err := m.AsPatch()
						if err != nil {
							return fmt.Errorf("failed patching %s: %w", scenarioFile, err)
						}

						data, err = patch.Apply(data)
						if err != nil {
							return fmt.Errorf("failed patching %s: %w", scenarioFile, err)
						}
					}
				}

				fileTiming.Start("uploading")
				resp, err := api.CreateScenario(application, bytes.NewReader(data))
				if err != nil {
					fmt.Fprintf(os.Stderr, "warning, failed uploading %s: %s\n", scenarioFile, err)
				} else {
					scenarioUUIDs = append(scenarioUUIDs, resp.UUID)
				}
				progressBar.Add(1)

				fileTiming.Finish()
			}

			timing.Finish()

			if len(scenarioFiles) == 0 {
				return fmt.Errorf("no valid appmaps to upload")
			}

			// either both commit and branch are specified or both are unspecified
			// fail otherwise
			commitProvided := bool(git != nil && git.Commit != "")
			branchProvided := bool((git != nil && git.Branch != "") || options.branch != "")
			if commitProvided != branchProvided {
				cmd.SilenceUsage = false
				progressBar.Clear()
				if commitProvided {
					return fmt.Errorf("Git branch could not be resolved\nRun again with the --branch or -b flag specified")
				}
				return fmt.Errorf("The --branch or -b flag can only be provided when uploading appmaps from within a Git repository")
			}

			mapSet := appland.BuildMapSet(application, scenarioUUIDs).
				SetVersion(options.version).
				SetEnvironment(options.environment).
				WithGitMetadata(git).
				SetBranch(options.branch)

			res, err := api.CreateMapSet(mapSet)
			if err != nil {
				return fmt.Errorf("Failed creating mapset, %w", err)
			}

			progressBar.Finish()

			if options.bench {
				fmt.Println()
				timing.Print()
			}

			fmt.Printf("\n\nSuccess! %s has been updated with %d scenarios.\n", application, len(scenarioUUIDs))

			url := api.BuildUrl("applications", fmt.Sprintf("%d?mapset=%d", res.AppID, res.ID))
			if options.dontOpenBrowser {
				fmt.Println(url)
			} else {
				browser.OpenURL(url)
			}

			return nil
		},
	}
}

func init() {
	var (
		options   = &UploadOptions{}
		providers = []metadata.Provider{
			metadata.NewGitProvider(),
		}
		uploadCmd = NewUploadCommand(options, providers)
	)

	f := uploadCmd.Flags()
	f.BoolVar(&options.dontOpenBrowser, "no-open", false, "Do not open the browser after a successful upload")
	f.BoolVarP(&options.force, "force", "f", false, "Force uploading a file over size limit")
	f.BoolVarP(&options.bench, "bench", "", false, "Show a detailed breakdown of time spent uploading")
	f.StringVarP(&options.application, "app", "a", "", "Override the owning application")
	f.StringVar(&options.appmapPath, "f", "", "Specify an appmap.yml path")
	f.StringVarP(&options.branch, "branch", "b", "", "Set the mapset branch if it's otherwise unavailable from Git")
	f.StringVarP(&options.version, "version", "v", "", "Set the mapset version")
	f.StringVarP(&options.environment, "environment", "e", "", "Set the mapset environment")
	rootCmd.AddCommand(uploadCmd)
}
