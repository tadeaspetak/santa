package cmd

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/tadeaspetak/santa/cmd/cmdData"
	"github.com/tadeaspetak/santa/internal/app"
	"github.com/tadeaspetak/santa/internal/validation"
)

var isDebugFlagName = "debug"
var alwaysSendToFlagName = "alwaysSendTo"

var sendCmd = &cobra.Command{
	Use:   "send",
	Short: "send the email",
	Run: func(cmd *cobra.Command, args []string) {
		dat := (&cmdData.CmdData{}).Load(cmd)

		err := validation.Validate.Struct(dat)
		if err != nil {
			log.Fatalf("Invalid data: %v", err)
		}

		isDebug, err := cmd.Flags().GetBool(isDebugFlagName)
		if err != nil {
			log.Fatalf("Unable to get the %s flag: %v", isDebugFlagName, err)
		}

		alwaysSendTo, err := cmd.Flags().GetString(alwaysSendToFlagName)
		if err != nil {
			log.Fatalf("Unable to get the %s flag: %v", alwaysSendToFlagName, err)
		}

		app.Send(
			app.NewMailgunMailer(dat.Mailgun.Domain, dat.Mailgun.APIKey),
			app.PairParticipants(dat.Participants, 5),
			dat.Data.Template,
			isDebug,
			alwaysSendTo,
		)
	},
}

func init() {
	sendCmd.Flags().BoolP(isDebugFlagName, "d", false, "turn the debug mode on (won't send emails)")
	sendCmd.Flags().
		StringP(alwaysSendToFlagName, "a", "", "send all emails to the given address (for testing purposes)")
}
