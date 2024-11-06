package cmd

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/tadeaspetak/secret-reindeer/app"
	"github.com/tadeaspetak/secret-reindeer/internal/data"
	"github.com/tadeaspetak/secret-reindeer/internal/validation"
)

var isDebugFlagName = "debug"
var alwaysSendToFlagName = "alwaysSendTo"

var sendCmd = &cobra.Command{
	Use:   "send",
	Short: "send the email",
	Run: func(cmd *cobra.Command, args []string) {
		cmdData := (&data.CmdData{}).Load(cmd)

		err := validation.Validate.Struct(cmdData)
		if err != nil {
			log.Fatalf("Invalid data: %v", err)
		}

		isDebug, err := cmd.Flags().GetBool(isDebugFlagName)
		if err != nil {
			log.Fatalf("Unable to get the %s flag: %v", isDebugFlagName, err)
		}

		fixedRecipient, err := cmd.Flags().GetString(alwaysSendToFlagName)
		if err != nil {
			log.Fatalf("Unable to get the %s flag: %v", alwaysSendToFlagName, err)
		}

		app.Send(
			app.NewMailgunMailer(cmdData.Mailgun.Domain, cmdData.Mailgun.APIKey),
			app.Raffle(cmdData.Participants, 5),
			cmdData.Data.Template,
			isDebug,
			fixedRecipient,
		)
	},
}

func init() {
	sendCmd.Flags().BoolP(isDebugFlagName, "d", false, "turn the debug mode on (won't send emails)")
	sendCmd.Flags().
		StringP(alwaysSendToFlagName, "a", "", "send all emails to the given address (for debugging purposes)")
}
