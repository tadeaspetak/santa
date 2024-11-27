package participants

import (
	"fmt"
	"log"
	"slices"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/tadeaspetak/santa/cmd/cmdData"
	"github.com/tadeaspetak/santa/internal/data"
	"github.com/tadeaspetak/santa/internal/validation"
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

var AddCmd = &cobra.Command{
	Use:   "add",
	Short: "add a participant",
	Run: func(cmd *cobra.Command, args []string) {
		dat := (&cmdData.CmdData{}).Load(cmd)
		fmt.Print("\nAdd a new participant!\n\n")

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
		dat.Participants = append(
			dat.Participants,
			data.Participant{Person: data.Person{Salutation: salutation}, Email: email},
		)
		dat.Save()
		fmt.Printf("\nA new participant with an email '%v' and a salutation '%v' has been added.", email, salutation)
	},
}
