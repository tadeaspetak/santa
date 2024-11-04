package cmd

import (
	"log"

	"github.com/spf13/cobra"
	"github.com/tadeaspetak/secret-reindeer/app"
	"github.com/tadeaspetak/secret-reindeer/internal/data"
	"github.com/tadeaspetak/secret-reindeer/internal/validation"
)

var isDebugFlagName = "isDebug"
var fixedRecipientFlagName = "fixedRecipient"

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
			log.Fatalf("Unable to get the %s flag: %w", isDebugFlagName, err)
		}

		fixedRecipient, err := cmd.Flags().GetString(fixedRecipientFlagName)
		if err != nil {
			log.Fatalf("Unable to get the %s flag: %w", fixedRecipientFlagName, err)
		}

		app.Send(cmdData.Data, isDebug, fixedRecipient)
	},
}

func init() {
	sendCmd.Flags().BoolP(isDebugFlagName, "d", false, "turn the debug mode on")
	sendCmd.Flags().StringP(fixedRecipientFlagName, "f", "", "set persistent recipient for debugging purposes")
}
