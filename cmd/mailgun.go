package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tadeaspetak/santa/cmd/cmdData"
	"github.com/tadeaspetak/santa/internal/prompt"
)

var mailgunCmd = &cobra.Command{
	Use:   "mailgun",
	Short: "edit mailgun config",
	Run: func(cmd *cobra.Command, args []string) {
		dat := (&cmdData.CmdData{}).Load(cmd)

		fmt.Print(`
You need Mailgun to be set up for sending emails to your Secret Santa participants.
See the README for info on how to create a Mailgun account and get the required info.

`)

		dat.Mailgun.Domain = strings.TrimSpace(prompt.PromptStringEdit("Mailgun domain", dat.Mailgun.Domain))
		dat.Mailgun.APIKey = strings.TrimSpace(prompt.PromptStringEdit("Mailgun API key", dat.Mailgun.APIKey))

		dat.Save()
		fmt.Print("\nSaved the Mailgun settings.")
	},
}
