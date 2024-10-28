package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tadeaspetak/secret-reindeer/cmd/participants"
)

var dataPath string

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
	RootCmd.PersistentFlags().StringVarP(&dataPath, "data", "d", "data.json", "data path")
	RootCmd.AddCommand(participants.ParticipantsCmd)

}
