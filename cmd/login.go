package cmd

import (
	"bufio"
	"fmt"
	"os"

	"github.com/applandinc/appland-cli/internal/config"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
)

func init() {
	var (
		loginCmd = &cobra.Command{
			Use:   "login",
			Short: "Login to AppLand",
			Args:  cobra.MaximumNArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				reader := bufio.NewReader(os.Stdin)
				context := config.GetCurrentContext()

				fmt.Printf("logging into %s\n\n", context.URL)
				fmt.Printf("login: ")
				_, err := reader.ReadString('\n')
				if err != nil {
					fail(err)
				}

				fmt.Printf("password: ")
				_, err = terminal.ReadPassword(0)
				if err != nil {
					fail(err)
				}
				fmt.Printf("\n\nlogged in.\n")

				context.APIKey = "TODO"
			},
		}
	)

	rootCmd.AddCommand(loginCmd)
}
