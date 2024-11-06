package participants

import (
	"errors"
	"fmt"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/tadeaspetak/secret-reindeer/internal/data"
	"github.com/tadeaspetak/secret-reindeer/internal/prompt"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "delete a participant",
	Run: func(cmd *cobra.Command, args []string) {
		cmdData := (&data.CmdData{}).Load(cmd)

		fmt.Print("Delete a participant:\n\n")
		participantIndex := prompt.PromptSelectParticipant(cmdData.Participants, "Deleting")

		// confirm
		email := cmdData.Participants[participantIndex].Email
		prompt := promptui.Prompt{
			Label:     fmt.Sprintf("Are you sure you want to delete %s?", email),
			IsConfirm: true,
			Default:   "y",
		}
		_, err := prompt.Run()
		if err != nil {
			if errors.Is(err, promptui.ErrAbort) {
				fmt.Printf("Doing nothing then.")
				return
			}

			fmt.Printf("Prompt failed %v\n", err)
			return
		}

		// actually remove
		cmdData.RemoveParticipant(participantIndex)
		cmdData.Save()
		fmt.Printf("Successfully deleted the participant %v.", email)

	},
}
