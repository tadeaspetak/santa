package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tadeaspetak/secret-reindeer/internal/data"
	"github.com/tadeaspetak/secret-reindeer/internal/prompt"
)

var mailgunCmd = &cobra.Command{
	Use:   "mailgun",
	Short: "edit mailgun config",
	Run: func(cmd *cobra.Command, args []string) {
		cmdData := (&data.CmdData{}).Load(cmd)

		fmt.Println("Edit the mailgun config:\n")

		cmdData.Mailgun.Domain = strings.TrimSpace(prompt.PromptStringEdit("Mailgun domain", cmdData.Mailgun.Domain))
		cmdData.Mailgun.APIKey = strings.TrimSpace(prompt.PromptStringEdit("Mailgun API key", cmdData.Mailgun.APIKey))

		cmdData.Save()
	},
}
