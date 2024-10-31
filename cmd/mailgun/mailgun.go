package mailgun

import (
	"github.com/spf13/cobra"
)

var MailgunCmd = &cobra.Command{
	Use:   "mailgun",
	Short: "mailgun config",
}

func init() {
	MailgunCmd.AddCommand(editCmd)
}
