package participants

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

func editExcludedRecipients(participants []data.Participant, participantIndex int) []data.Participant {
	participant := &participants[participantIndex]

	// create a map for convenience & performance reasons
	currentExcludedRecipients := map[string]bool{}
	for _, email := range participant.ExcludedRecipients {
		currentExcludedRecipients[email] = true
	}

	// create items for the multi-select
	multiSelectParticipants := []prompt.PromptMultiSelectItem{}
	for _, p := range participants {
		// skip the participant themself
		if p.Email == participant.Email {
			continue
		}
		multiSelectParticipants = append(multiSelectParticipants, prompt.PromptMultiSelectItem{
			ID:         p.Email,
			IsSelected: currentExcludedRecipients[p.Email],
		})
	}
	excludedParticipants, _ := prompt.PromptMultiSelect(multiSelectParticipants, 1, "Exclude the following participants")

	// update the excluded recipients
	participant.ExcludedRecipients = make([]string, len(excludedParticipants))
	for i, item := range excludedParticipants {
		participant.ExcludedRecipients[i] = item.ID
	}

	return participants
}

type predestinedItem struct {
	Label      string
	IsSelected bool
}

func editPredestined(emails []string, selected string) int {
	items := make([]predestinedItem, len(emails)+1)
	for i, email := range emails {
		items[i] = predestinedItem{Label: email, IsSelected: selected == email}
	}
	items[len(items)-1] = predestinedItem{Label: "Remove", IsSelected: false}

	templates := &promptui.SelectTemplates{
		Label:    `{{if .IsSelected}}✔ {{end}} {{ .Label }} - label`,
		Active:   "→ {{if .IsSelected}}✔ {{end}}{{ .Label | cyan }}",
		Inactive: "{{if .IsSelected}}✔ {{end}}{{ .Label | cyan }}",
	}

	prompt := promptui.Select{
		Label:     "Select",
		Items:     items,
		Templates: templates,
		Size:      5,
	}

	index, _, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return -2
	}

	if index == len(items)-1 {
		return -1
	}

	return index
}

var editCmd = &cobra.Command{
	Use:   "edit",
	Short: "edit a participant",
	Run: func(cmd *cobra.Command, args []string) {
		cmdData := data.LoadCmdData(cmd)
		fmt.Println("Edit a aprticipant:\n")

		for {
			// select a participant
			participantIndex := prompt.PromptSelectParticipant(cmdData.Participants, "Editing")
			if participantIndex < 0 {
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
				prev := cmdData.Participants[participantIndex].Email
				next := prompt.PromptStringEdit("Edit email", prev)
				cmdData.UpdateParticipantEmail(participantIndex, prev, next)
			case EditSalutation:
				participant := &cmdData.Participants[participantIndex]
				salutation := prompt.PromptStringEdit("Edit salutation", participant.Salutation)
				participant.Salutation = salutation
			case EditExcludedRecipients:
				editExcludedRecipients(cmdData.Participants, participantIndex)
			case EditPredestinedRecipient:
				part := &cmdData.Participants[participantIndex]
				available := make([]string, 0)
				for i, p := range cmdData.Participants {
					if i != participantIndex {
						available = append(available, p.Email)
					}
				}
				predestinedindex := editPredestined(available, part.PredestinedRecipient)
				if predestinedindex == -1 {
					part.PredestinedRecipient = ""
				} else {
					part.PredestinedRecipient = available[predestinedindex]
				}
			}

			// save the data
			data.SaveCmdData(cmd, cmdData)
		}
	},
}
