package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

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
				context, err := config.GetCurrentContext()
				if err != nil {
					fail(err)
				}

				fmt.Printf("logging into %s\n\n", context.GetURL())
				fmt.Printf("login: ")
				login, err := reader.ReadString('\n')
				if err != nil {
					fail(err)
				}

				fmt.Printf("password: ")
				password, err := terminal.ReadPassword(0)
				if err != nil {
					fail(err)
				}
				fmt.Printf("\n\nlogged in.\n")

				err = api.Login(strings.TrimSpace(login), strings.TrimSpace(string(password)))
				if err != nil {
					fail(err)
				}
			},
		}
	)

	rootCmd.AddCommand(loginCmd)
}
