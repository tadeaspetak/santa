package participants

import (
	"github.com/spf13/cobra"
	"github.com/tadeaspetak/secret-reindeer/cmd/participants/edit"
)

var ParticipantsCmd = &cobra.Command{
	Use:   "participants",
	Short: "manage participants",
	Long:  "Create, edit and delete participants.",
}

func init() {
	ParticipantsCmd.AddCommand(AddCmd)
	ParticipantsCmd.AddCommand(deleteCmd)
	ParticipantsCmd.AddCommand(edit.EditParticipantCmd)
}
