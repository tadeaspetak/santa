package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tadeaspetak/secret-reindeer/cmd/cmdData"
	"github.com/tadeaspetak/secret-reindeer/internal/prompt"
)

var mailgunCmd = &cobra.Command{
	Use:   "mailgun",
	Short: "edit mailgun config",
	Run: func(cmd *cobra.Command, args []string) {
		dat := (&cmdData.CmdData{}).Load(cmd)

		fmt.Print("Edit the mailgun config. See the README for info on how to get create a Mailgun account.\n\n")

		dat.Mailgun.Domain = strings.TrimSpace(prompt.PromptStringEdit("Mailgun domain", dat.Mailgun.Domain))
		dat.Mailgun.APIKey = strings.TrimSpace(prompt.PromptStringEdit("Mailgun API key", dat.Mailgun.APIKey))

		dat.Save()
		fmt.Println("Saved the changes.")
	},
}
