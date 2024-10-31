package template

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tadeaspetak/secret-reindeer/internal/data"
	"github.com/tadeaspetak/secret-reindeer/internal/prompt"
)

var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "edit template config",
	Run: func(cmd *cobra.Command, args []string) {
		cmdData := data.LoadCmdData(cmd)
		fmt.Println("Edit the template:\n")

		cmdData.Email.Body = strings.TrimSpace(prompt.PromptStringEdit("Body", cmdData.Email.Body))
		cmdData.Email.Sender = strings.TrimSpace(prompt.PromptStringEdit("Sender", cmdData.Email.Sender))
		cmdData.Email.Subject = strings.TrimSpace(prompt.PromptStringEdit("Subject", cmdData.Email.Subject))

		data.SaveCmdData(cmd, cmdData)
	},
}
