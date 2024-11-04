package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tadeaspetak/secret-reindeer/cmd/participants"
	"github.com/tadeaspetak/secret-reindeer/internal/data"
)

var RootCmd = &cobra.Command{
	Use:   "reindeer",
	Short: "short desc",
	Long: `Long
            desc `,
}

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	RootCmd.PersistentFlags().String(data.DataPathFlagName, "data/data.json", "data path")
	RootCmd.AddCommand(participants.ParticipantsCmd)
	RootCmd.AddCommand(mailgunCmd)
	RootCmd.AddCommand(templateCmd)
	RootCmd.AddCommand(sendCmd)

}
