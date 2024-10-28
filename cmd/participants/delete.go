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
	Long:  `Delete a participant.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmdData := data.LoadCmdData(cmd)

		// select a participant
		participantIndex := prompt.PromptSelectParticipant(cmdData.Participants, "Deleting")
		if participantIndex < 0 {
			fmt.Printf("Failed to select a participant")
			return
		}

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

		cmdData.RemoveParticipant(participantIndex)
		data.SaveCmdData(cmd, cmdData)
		fmt.Printf("Successfully deleted participant %v.", email)

	},
}
