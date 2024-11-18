package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tadeaspetak/santa/cmd/cmdData"
	"github.com/tadeaspetak/santa/cmd/participants"
)

var RootCmd = &cobra.Command{
	Short: "generate your secret santa pairings easily",
	Long: `Santa is a simple CLI app that makes drawing your Secret Santa pairings a breeze.

For more info, head over to the README at https://github.com/tadeaspetak/santa/blob/main/README.md`,
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// disable the default completion command
	RootCmd.CompletionOptions.DisableDefaultCmd = true

	RootCmd.PersistentFlags().String(cmdData.DataPathFlagName, "data.json", "data file path")

	RootCmd.AddCommand(participants.ParticipantsCmd)
	RootCmd.AddCommand(mailgunCmd)
	RootCmd.AddCommand(templateCmd)
	RootCmd.AddCommand(sendCmd)
	RootCmd.AddCommand(versionCmd)
	RootCmd.AddCommand(initCmd)

}
