package participants

import (
	"fmt"
	"log"
	"slices"

	"github.com/spf13/cobra"
	"github.com/tadeaspetak/secret-reindeer/internal/data"
	"github.com/tadeaspetak/secret-reindeer/internal/prompt"
	"github.com/tadeaspetak/secret-reindeer/internal/validation"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "add a participant",
	Long:  `Add a participant.`,
	Run: func(cmd *cobra.Command, args []string) {
		cmdData := data.LoadCmdData(cmd)
		fmt.Println("Let's add a new participant!\n")

		// collect the data
		email := validation.SanitizeEmail(prompt.PromptStringNew("Email address"))
		salutation := prompt.PromptStringNew("Salutation")

		// ensure the email is unique
		if index := slices.IndexFunc(cmdData.Participants, func(participant data.Participant) bool {
			return participant.Email == email
		}); index > -1 {
			log.Fatalf("A participant with an email '%v' already exists.", email)
		}

		// add a participant
		cmdData.Participants = append(cmdData.Participants, data.Participant{
			Email:      email,
			Salutation: salutation,
		})
		data.SaveCmdData(cmd, cmdData)
		fmt.Printf("A new participant with an email '%v' and a salutation '%v' has been added.\n", email, salutation)
	},
}
