package participants

import (
	"fmt"
	"log"
	"slices"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/tadeaspetak/secret-reindeer/cmd/cmdData"
	"github.com/tadeaspetak/secret-reindeer/internal/data"
	"github.com/tadeaspetak/secret-reindeer/internal/validation"
)

func promptStringNew(label string) string {
	prompt := promptui.Prompt{
		Label: label,
	}

	value, err := prompt.Run()
	if err != nil {
		log.Fatalf("Prompt failed %v.\n", err)
	}

	return value
}

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "add a participant",
	Run: func(cmd *cobra.Command, args []string) {
		dat := (&cmdData.CmdData{}).Load(cmd)
		fmt.Print("Add a new participant!\n\n")

		// collect the data
		email := validation.SanitizeEmail(promptStringNew("Email address"))
		salutation := strings.TrimSpace(promptStringNew("Salutation"))

		// ensure the email is unique
		if index := slices.IndexFunc(dat.Participants, func(participant data.Participant) bool {
			return participant.Email == email
		}); index > -1 {
			log.Fatalf("A participant with an email '%v' already exists.", email)
		}

		// add a participant
		dat.Participants = append(dat.Participants, data.Participant{
			Email:      email,
			Salutation: salutation,
		})
		dat.Save()
		fmt.Printf("A new participant with an email '%v' and a salutation '%v' has been added.\n", email, salutation)
	},
}
