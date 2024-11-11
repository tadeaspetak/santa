package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tadeaspetak/secret-reindeer/cmd/version"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "app version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version.Version)
	},
}
