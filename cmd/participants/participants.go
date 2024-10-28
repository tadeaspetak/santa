package participants

import (
	"github.com/spf13/cobra"
)

var ParticipantsCmd = &cobra.Command{
	Use:   "participants",
	Short: "Secret reindeer participants.",
}

func init() {
	ParticipantsCmd.AddCommand(addCmd)
	ParticipantsCmd.AddCommand(deleteCmd)
	ParticipantsCmd.AddCommand(editCmd)
}
