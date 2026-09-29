package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tadeaspetak/santa/cmd/version"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the app version",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintln(cmd.OutOrStdout(), version.Version)
		},
	}
}
