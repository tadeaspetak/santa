package participants

import (
	"github.com/spf13/cobra"
	"github.com/tadeaspetak/secret-reindeer/cmd/participants/edit"
)

var ParticipantsCmd = &cobra.Command{
	Use:   "participants",
	Short: "Secret reindeer participants.",
	Long:  "Create, edit and delete secret reindeer participants.",
}

func init() {
	ParticipantsCmd.AddCommand(addCmd)
	ParticipantsCmd.AddCommand(deleteCmd)
	ParticipantsCmd.AddCommand(edit.EditParticipantCmd)
}
