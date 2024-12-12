package cmd

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
)

var dataPathFlagName = "data"

func getDataPath(cmd *cobra.Command) string {
	dataPath, err := cmd.Flags().GetString(dataPathFlagName)

	if err != nil {
		log.Fatalf("Could not get data-file path %v\n", err)
	}

	return dataPath
}

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

	RootCmd.PersistentFlags().String(dataPathFlagName, "data.json", "data file path")

	RootCmd.AddCommand(sendCmd)
	RootCmd.AddCommand(versionCmd)
	RootCmd.AddCommand(initCmd)

}
