package mailgun

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tadeaspetak/secret-reindeer/internal/data"
	"github.com/tadeaspetak/secret-reindeer/internal/prompt"
)

var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "edit mailgun config",
	Run: func(cmd *cobra.Command, args []string) {
		cmdData := data.LoadCmdData(cmd)
		fmt.Println("Edit the mailgun config:\n")

		cmdData.Mailgun.Domain = strings.TrimSpace(prompt.PromptStringEdit("Mailgun domain", cmdData.Mailgun.Domain))
		cmdData.Mailgun.APIKey = strings.TrimSpace(prompt.PromptStringEdit("Mailgun API key", cmdData.Mailgun.APIKey))

		data.SaveCmdData(cmd, cmdData)
	},
}
