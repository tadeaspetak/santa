package template

import (
	"github.com/spf13/cobra"
)

var TemplateCmd = &cobra.Command{
	Use:   "template",
	Short: "template config",
}

func init() {
	TemplateCmd.AddCommand(editCmd)
}
