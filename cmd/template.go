package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tadeaspetak/santa/cmd/cmdData"
	"github.com/tadeaspetak/santa/internal/prompt"
)

var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "edit email template config",
	Run: func(cmd *cobra.Command, args []string) {
		dat := (&cmdData.CmdData{}).Load(cmd)

		fmt.Print(`
Edit your email template. This is used to construct emails sent to your participants.
Use '%{recipientSalutation}' (without the aposotrophes) to substitute the recipient's salutation in the body and subject.

`)

		dat.Template.Body = strings.TrimSpace(prompt.PromptStringEdit("Body", dat.Template.Body))
		dat.Template.Subject = strings.TrimSpace(prompt.PromptStringEdit("Subject", dat.Template.Subject))
		dat.Template.Sender = strings.TrimSpace(prompt.PromptStringEdit("Sender", dat.Template.Sender))
		// note: do not trim spaces here, they are most likely intentional
		dat.Template.RecipientsSeparator =
			prompt.PromptStringEdit("RecipientsSeparator", dat.Template.RecipientsSeparator)

		dat.Save()
		fmt.Print("\nSaved the template.\n\n")
	},
}
