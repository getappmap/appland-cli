package cmd

import (
	"fmt"

	"github.com/applandinc/appland-cli/internal/config"
	"github.com/spf13/cobra"
)

func init() {
	var (
		logoutCmd = &cobra.Command{
			Use:   "logout",
			Short: "Log out of AppLand",
			Run: func(cmd *cobra.Command, args []string) {
				context := config.GetCurrentContext()

				if context.APIKey == "" {
					fail(fmt.Errorf("not logged in to %s", context.URL))
				}

				err := api.DeleteAPIKey()
				if err != nil {
					fail(err)
				}
				fmt.Printf("logged out of %s\n", context.URL)
			},
		}
	)

	rootCmd.AddCommand(logoutCmd)
}
