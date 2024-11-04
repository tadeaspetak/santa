package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tadeaspetak/secret-reindeer/internal/data"
	"github.com/tadeaspetak/secret-reindeer/internal/prompt"
)

var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "edit template config",
	Run: func(cmd *cobra.Command, args []string) {
		cmdData := (&data.CmdData{}).Load(cmd)

		// TODO: explain available variables
		fmt.Println("Edit the template:\n")

		cmdData.Template.Body = strings.TrimSpace(prompt.PromptStringEdit("Body", cmdData.Template.Body))
		cmdData.Template.Subject = strings.TrimSpace(prompt.PromptStringEdit("Subject", cmdData.Template.Subject))

		cmdData.Save()
	},
}
