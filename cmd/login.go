package cmd

import (
	"bufio"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/applandinc/appland-cli/internal/config"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh/terminal"
)

type PasswordReader func() ([]byte, error)

var stdinPasswordReader = func() ([]byte, error) {
	return terminal.ReadPassword(0)
}

func NewLoginCommand(connecter Connecter, stdin io.Reader, passwordReader PasswordReader) *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Login to AppLand",
		Args:  cobra.MaximumNArgs(1),
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			config.LoadCLIConfig()
			// defer connecting to the server until Run (below)
		},
		Run: func(cmd *cobra.Command, args []string) {
			reader := bufio.NewReader(stdin)

			api := connecter()

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
			login = strings.TrimSpace(login)

			validAPIKey := false
			data, err := base64.RawURLEncoding.DecodeString(login)
			if err == nil && len(strings.Split(string(data), ":")) == 2 {
				// Looks like an API key, try it out
				validAPIKey, err = api.TestAPIKey(login)
				if err != nil {
					fail(err)
				}
				if !validAPIKey {
					fail(fmt.Errorf("Invalid API key"))
				}
			}

			if validAPIKey {
				context.SetAPIKey(login)
			} else {
				fmt.Printf("password: ")
				bytes, err := passwordReader()
				if err != nil {
					fail(err)
				}

				fmt.Println("logging in....")
				err = api.Login(login, strings.TrimSpace(string(bytes)))
				if err != nil {
					fail(err)
				}
			}

			fmt.Printf("\n\nlogged in.\n")
		},
	}
}

func init() {
	rootCmd.AddCommand(NewLoginCommand(DefaultConnecter, os.Stdin, stdinPasswordReader))
}
