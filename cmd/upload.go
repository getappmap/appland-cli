package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/applandinc/appland-cli/internal/appland"
	"github.com/applandinc/appland-cli/internal/config"
	"github.com/spf13/cobra"
)

func init() {
	var (
		user         string
		organization string

		uploadCmd = &cobra.Command{
			Use:   "upload [files]",
			Short: "Upload AppMap files to AppLand",
			Args:  cobra.MinimumNArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				api := appland.MakeClient(config.GetCurrentContext())
				batchID := ""

				for _, path := range args {
					file, err := os.Open(path)
					if err != nil {
						fail(err)
					}

					data, err := ioutil.ReadAll(file)
					if err != nil {
						fail(err)
					}

					obj := struct {
						User string `json:"user"`
						Org  string `json:"org"`
						Data string `json:"data"`
					}{
						User: user,
						Org:  organization,
						Data: string(data),
					}

					jsonData, err := json.Marshal(&obj)
					if err != nil {
						fail(err)
					}

					resp, err := api.CreateScenario(bytes.NewReader(jsonData), batchID)
					if err != nil {
						fail(err)
					}

					if batchID == "" {
						batchID = resp.BatchID
					}
				}

				fmt.Printf("uploaded %d scenarios\n", len(args))
				fmt.Printf("view the batch: %s\n", api.BuildUrl("scenario_batches", batchID))
			},
		}
	)

	uploadCmd.Flags().StringVarP(&user, "user", "u", "", "[soon to be deprecated] specify a user to own the AppMaps")
	uploadCmd.Flags().StringVarP(&organization, "organization", "o", "", "override the owning organization")
	rootCmd.AddCommand(uploadCmd)
}
