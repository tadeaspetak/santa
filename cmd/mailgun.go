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

		fmt.Print("Edit the mailgun config. See the README for info on how to get create a Mailgun account.\n\n")

		cmdData.Mailgun.Domain = strings.TrimSpace(prompt.PromptStringEdit("Mailgun domain", cmdData.Mailgun.Domain))
		cmdData.Mailgun.APIKey = strings.TrimSpace(prompt.PromptStringEdit("Mailgun API key", cmdData.Mailgun.APIKey))

		cmdData.Save()
		fmt.Println("Saved the changes.")
	},
}
