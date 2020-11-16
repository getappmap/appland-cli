package cmd

import (
	"github.com/applandinc/appland-cli/internal/recording"
	"github.com/spf13/cobra"
)

func init() {
	var (
		recordingCmd = &cobra.Command{
			Use:   "recording",
			Short: "Manage AppMap recordings",
		}
		recordingStartCmd = &cobra.Command{
			Use:   "start [url]",
			Short: "Start a new AppMap recording session",
			Args:  cobra.ExactArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				recording.StartRecording(args[0])
			},
		}
		recordingStopCmd = &cobra.Command{
			Use:   "stop [url]",
			Short: "Stop an existing AppMap recording session",
			Args:  cobra.ExactArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				recording.StopRecording(args[0])
			},
		}
		recordingCheckCmd = &cobra.Command{
			Use:   "check [url]",
			Short: "Check the current AppMap recording status",
			Args:  cobra.ExactArgs(1),
			Run: func(cmd *cobra.Command, args []string) {
				recording.CheckRecording(args[0])
			},
		}
	)
	rootCmd.AddCommand(recordingCmd)
	recordingCmd.AddCommand(recordingStartCmd)
	recordingCmd.AddCommand(recordingStopCmd)
	recordingCmd.AddCommand(recordingCheckCmd)
}
