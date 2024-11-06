package edit

import (
	"fmt"
	"log"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/tadeaspetak/secret-reindeer/internal/data"
	"github.com/tadeaspetak/secret-reindeer/internal/prompt"
)

type editActionID int

type editAction struct {
	ID    editActionID
	Label string
}

const (
	EditEmail = iota
	EditSalutation
	EditExcludedRecipients
	EditPredestinedRecipient
)

func selectEditAction() editActionID {
	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "→ {{ .Label | cyan }}",
		Inactive: "  {{ .Label | cyan }}",
		Selected: "Editing {{ .Label }}",
	}

	editActions := []editAction{
		{ID: EditEmail, Label: "Email"},
		{ID: EditSalutation, Label: "Salutation"},
		{ID: EditExcludedRecipients, Label: "Excluded participants"},
		{ID: EditPredestinedRecipient, Label: "Predestined recipient"},
	}

	prompt := promptui.Select{
		Label:     "What would you like to edit?",
		Items:     editActions,
		Templates: templates,
	}

	index, _, err := prompt.Run()

	if err != nil {
		log.Fatalf("Prompt failed %v\n", err)
	}

	return editActions[index].ID
}

var EditParticipantCmd = &cobra.Command{
	Use:   "edit",
	Short: "edit a participant",
	Run: func(cmd *cobra.Command, args []string) {
		cmdData := (&data.CmdData{}).Load(cmd)
		fmt.Print("Edit a aprticipant:\n\n")

		// repeat forever to make editing multiple participants smoother
		for {
			// select participant and action
			editedParticipantIndex := prompt.PromptSelectParticipant(cmdData.Participants, "Editing")
			actionIndex := selectEditAction()

			switch editActionID(actionIndex) {
			case EditEmail:
				curr := cmdData.Participants[editedParticipantIndex].Email
				next := prompt.PromptStringEdit("Edit email", curr)
				cmdData.UpdateParticipantEmail(editedParticipantIndex, curr, next)
			case EditSalutation:
				participant := &cmdData.Participants[editedParticipantIndex]
				salutation := prompt.PromptStringEdit("Edit salutation", participant.Salutation)
				participant.Salutation = salutation
			case EditExcludedRecipients:
				editExcludedRecipients(cmdData.Participants, editedParticipantIndex)
			case EditPredestinedRecipient:
				editPredestinedParticipant(cmdData.Participants, editedParticipantIndex)
			}

			fmt.Printf("%s")

			// save the data
			cmdData.Save()
		}
	},
}
