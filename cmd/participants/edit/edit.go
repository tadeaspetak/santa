package edit

import (
	"fmt"
	"log"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/tadeaspetak/santa/cmd/cmdData"
	"github.com/tadeaspetak/santa/internal/prompt"
	"github.com/tadeaspetak/santa/internal/validation"
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
		dat := (&cmdData.CmdData{}).Load(cmd)
		fmt.Print("Edit a aprticipant:\n\n")

		// repeat forever to make editing multiple participants smoother
		for {
			// select participant and action
			editedParticipantIndex := prompt.PromptSelectParticipant(dat.Participants, "Editing")
			actionIndex := selectEditAction()

			switch editActionID(actionIndex) {
			case EditEmail:
				curr := dat.Participants[editedParticipantIndex].Email
				next := validation.SanitizeEmail(prompt.PromptStringEdit("Edit email", curr))
				err := dat.UpdateParticipantEmail(editedParticipantIndex, curr, next)
				if err != nil {
					log.Fatalf("Failed to update participant email: %v", err)
				}
			case EditSalutation:
				participant := &dat.Participants[editedParticipantIndex]
				salutation := prompt.PromptStringEdit("Edit salutation", participant.Salutation)
				participant.Salutation = salutation
			case EditExcludedRecipients:
				editExcludedRecipients(dat.Participants, editedParticipantIndex)
			case EditPredestinedRecipient:
				editPredestinedParticipant(dat.Participants, editedParticipantIndex)
			}

			// save the data
			dat.Save()
		}
	},
}
