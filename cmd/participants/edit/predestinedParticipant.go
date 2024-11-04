package edit

import (
	"fmt"
	"slices"

	"github.com/manifoldco/promptui"
	"github.com/tadeaspetak/secret-reindeer/internal/data"
)

func editPredestinedParticipant(participants []data.Participant, editedParticipantIndex int) {
	// get already taken participants
	taken := make([]string, 0)
	for _, p := range participants {
		if p.PredestinedRecipient != "" {
			taken = append(taken, p.PredestinedRecipient)
		}
	}

	// get the available choices, i.e. not themself and not already taken by someone else
	available := make([]string, 0)
	for i, p := range participants {
		if i != editedParticipantIndex && slices.Index(taken, p.Email) == -1 {
			available = append(available, p.Email)
		}
	}

	editedParticipant := &participants[editedParticipantIndex]
	predestinedindex := promptPredestinedSelect(available, editedParticipant.PredestinedRecipient)
	if predestinedindex == -1 {
		editedParticipant.PredestinedRecipient = ""
	} else {
		editedParticipant.PredestinedRecipient = available[predestinedindex]
	}
}

type predestinedItem struct {
	Label      string
	IsSelected bool
}

func promptPredestinedSelect(emails []string, selected string) int {
	removeLabel := "Remove"

	items := make([]predestinedItem, len(emails)+1)
	for i, email := range emails {
		items[i] = predestinedItem{Label: email, IsSelected: selected == email}
	}
	items[len(items)-1] = predestinedItem{Label: removeLabel, IsSelected: false}

	templates := &promptui.SelectTemplates{
		Label:    `{{if .IsSelected}}✔ {{end}} {{ .Label }}`,
		Active:   "→ {{if .IsSelected}}✔ {{end}}{{ .Label | cyan }}",
		Inactive: "{{if .IsSelected}}✔ {{end}}{{ .Label | cyan }}",
	}

	prompt := promptui.Select{
		Label:     "Select",
		Items:     items,
		Templates: templates,
	}

	index, _, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		// TODO: return error
		return -2
	}

	if index == len(items)-1 {
		return -1
	}

	return index
}
