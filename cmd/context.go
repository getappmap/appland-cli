package cmd

import (
	"fmt"

	"github.com/applandinc/appland-cli/internal/config"
	"github.com/spf13/cobra"
)

func init() {
	var (
		contextName string

		contextCmd = &cobra.Command{
			Use:   "context",
			Short: "manage AppLand contexts",
			Run: func(cmd *cobra.Command, args []string) {
				cmd.Usage()
			},
		}

		contextAddCmd = &cobra.Command{
			Use:   "add",
			Short: "add a new AppLand context",
			Args:  cobra.ExactArgs(2),
			Run: func(cmd *cobra.Command, args []string) {
				name := args[0]
				url := args[1]

				if err := config.MakeContext(name, url); err != nil {
					fail(err)
				}
			},
		}

		contextCurrentCmd = &cobra.Command{
			Use:   "current",
			Short: "show the current AppLand context",
			Run: func(cmd *cobra.Command, args []string) {
				name := config.GetCurrentContextName()
				context := config.GetCurrentContext()
				fmt.Printf("%s: %s\n", name, context.URL)
			},
		}

		contextSetCmd = &cobra.Command{
			Use:   "set",
			Short: "set a context value",
			Args:  cobra.ExactArgs(2),
			Run: func(cmd *cobra.Command, args []string) {
				key := args[0]
				value := args[1]
				var context *config.Context
				if contextName != "" {
					context = config.GetContext(contextName)
					if context == nil {
						fail(fmt.Errorf("context '%s' not found", contextName))
					}
				} else {
					context = config.GetCurrentContext()
					contextName = config.GetCurrentContextName()
				}

				if context == nil {
					fail(fmt.Errorf("no context could be resolved"))
				}

				switch key {
				case "url":
					context.URL = value
				case "api_key":
					context.APIKey = value
				case "name":
					config.RenameContext(contextName, value)
				default:
					fail(fmt.Errorf("unknown key '%s'", key))
				}

				fmt.Printf("'%s' set for '%s'\n", key, contextName)
			},
		}

		contextUseCmd = &cobra.Command{
			Use:   "use",
			Short: "change the current AppLand context",
			Args:  cobra.ExactArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				name := args[0]

				if err := config.SetCurrentContext(name); err != nil {
					fail(err)
				}

				fmt.Printf("current context set to '%s'\n", name)
			},
		}

		contextListCmd = &cobra.Command{
			Use:   "list",
			Short: "list all AppLand contexts",
			Run: func(cmd *cobra.Command, args []string) {
				configuration := config.GetCLIConfig()
				for name, context := range configuration.Contexts {
					fmt.Printf("%s: %s\n", name, context.URL)
				}
			},
		}
	)

	rootCmd.AddCommand(contextCmd)
	contextCmd.AddCommand(contextAddCmd)
	contextCmd.AddCommand(contextSetCmd)
	contextCmd.AddCommand(contextCurrentCmd)
	contextCmd.AddCommand(contextUseCmd)
	contextCmd.AddCommand(contextListCmd)

	contextSetCmd.Flags().StringVarP(&contextName, "context", "c", "", "name of a context")
}
