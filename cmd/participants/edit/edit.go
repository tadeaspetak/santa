package edit

import (
	"fmt"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/tadeaspetak/secret-reindeer/internal/data"
	"github.com/tadeaspetak/secret-reindeer/internal/prompt"
)

type editActionID int

const (
	EditEmail = iota
	EditSalutation
	EditExcludedRecipients
	EditPredestinedRecipient
)

type editAction struct {
	ID    editActionID
	Label string
}

func selectEditAction() (editActionID, error) {
	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "👉 {{ .Label | cyan }}",
		Inactive: "   {{ .Label | cyan }}",
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
		fmt.Printf("Prompt failed %v\n", err)
		return -1, err
	}

	return editActions[index].ID, nil
}

var EditParticipantCmd = &cobra.Command{
	Use:   "edit",
	Short: "edit a participant",
	Run: func(cmd *cobra.Command, args []string) {
		cmdData := (&data.CmdData{}).Load(cmd)
		fmt.Println("Edit a aprticipant:\n")

		for {
			// select a participant
			editedParticipantIndex := prompt.PromptSelectParticipant(cmdData.Participants, "Editing")
			if editedParticipantIndex < 0 {
				fmt.Printf("Failed to select a participant")
				return
			}

			// choose an action
			actionIndex, _ := selectEditAction()
			if actionIndex < 0 {
				fmt.Printf("Failed to select an action")
				return
			}

			// make the update
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

			// save the data
			cmdData.Save()
		}
	},
}
