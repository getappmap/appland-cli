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
				context, err := config.GetCurrentContext()
				if err != nil {
					fail(err)
				}

				if context.GetAPIKey() == "" {
					fail(fmt.Errorf("Not logged in to %s", context.GetURL()))
				}

				err = api.DeleteAPIKey()
				if err != nil {
					fail(err)
				}
				fmt.Printf("Logged out of %s\n", context.GetURL())
			},
		}
	)

	rootCmd.AddCommand(logoutCmd)
}
