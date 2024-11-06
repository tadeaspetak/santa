package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tadeaspetak/secret-reindeer/cmd/cmdData"
	"github.com/tadeaspetak/secret-reindeer/internal/prompt"
)

var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "edit template config",
	Run: func(cmd *cobra.Command, args []string) {
		dat := (&cmdData.CmdData{}).Load(cmd)

		fmt.Print(
			"Edit the template. Use '%{recipientSalutation}' (without the aposotrophes) to substitute the recipient's salutation.\n\n",
		)

		dat.Template.Body = strings.TrimSpace(prompt.PromptStringEdit("Body", dat.Template.Body))
		dat.Template.Subject = strings.TrimSpace(prompt.PromptStringEdit("Subject", dat.Template.Subject))

		dat.Save()
		fmt.Println("Saved the changes.")
	},
}
