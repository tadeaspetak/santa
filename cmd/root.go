package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tadeaspetak/secret-reindeer/cmd/cmdData"
	"github.com/tadeaspetak/secret-reindeer/cmd/participants"
)

var RootCmd = &cobra.Command{
	Short: "TODO",
	Long: `TODO: Long
            desc `,
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	RootCmd.PersistentFlags().String(cmdData.DataPathFlagName, "data/data.json", "data file path")

	RootCmd.AddCommand(participants.ParticipantsCmd)
	RootCmd.AddCommand(mailgunCmd)
	RootCmd.AddCommand(templateCmd)
	RootCmd.AddCommand(sendCmd)
	RootCmd.AddCommand(versionCmd)
	RootCmd.AddCommand(initCmd)

}
