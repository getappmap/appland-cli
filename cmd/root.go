package cmd

import (
	"fmt"
	"os"

	"github.com/applandinc/appland-cli/internal/appland"
	"github.com/applandinc/appland-cli/internal/build"
	"github.com/applandinc/appland-cli/internal/config"
	"github.com/spf13/cobra"
)

var (
	api     appland.Client
	rootCmd = &cobra.Command{
		Use:     "appland",
		Short:   "Manage AppLand resources",
		Version: build.Version,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			config.LoadCLIConfig()

			api = DefaultConnecter()
		},
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Usage()
		},
	}
)

func fail(err error) {
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
	os.Exit(1)
}

func warn(err error) {
	fmt.Fprintf(os.Stderr, "warn: %v\n", err)
}

type Connecter func() appland.Client

var DefaultConnecter = func() appland.Client {
	context, err := config.GetCurrentContext()
	if err != nil {
		fail(err)
	}

	return appland.MakeClient(context)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fail(err)
	}

	if err := config.WriteCLIConfig(); err != nil {
		fail(err)
	}
}
