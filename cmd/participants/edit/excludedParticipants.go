package edit

import (
	"log"

	"github.com/manifoldco/promptui"
	"github.com/tadeaspetak/secret-reindeer/internal/data"
)

type promptMultiSelectItem struct {
	ID         string
	IsSelected bool
}

func promptMultiSelect(items []promptMultiSelectItem, selectedPosition int, label string) []promptMultiSelectItem {
	if len(items) < 1 {
		log.Fatalf("No participants to be selected.\n")
	}

	// if the `done` item doesn't exist yet, append it
	const doneID = "Done"
	if items[len(items)-1].ID != doneID {
		items = append(items, promptMultiSelectItem{ID: doneID, IsSelected: false})
	}

	templates := &promptui.SelectTemplates{
		Label:    `{{if .IsSelected}}✔ {{end}} {{ .ID }} - label`,
		Active:   "→ {{if .IsSelected}}✔ {{end}}{{ .ID | cyan }}",
		Inactive: "{{if .IsSelected}}✔ {{end}}{{ .ID | cyan }}",
	}

	prompt := promptui.Select{
		Label:        label,
		Items:        items,
		Templates:    templates,
		CursorPos:    selectedPosition, // place the cursor at the currently selected position
		HideSelected: true,
	}

	index, _, err := prompt.Run()
	if err != nil {
		log.Fatalf("Prompt failed %v.\n", err)
	}

	if index != len(items)-1 {
		// unless `done` has been selected, toggle the current item and display again
		selectedItem := &items[index]
		selectedItem.IsSelected = !selectedItem.IsSelected
		return promptMultiSelect(items, index, label)
	}

	selectedItems := make([]promptMultiSelectItem, 0)
	for _, i := range items {
		if i.IsSelected {
			selectedItems = append(selectedItems, i)
		}
	}
	return selectedItems
}

func editExcludedRecipients(participants []data.Participant, participantIndex int) []data.Participant {
	participant := &participants[participantIndex]

	// create a map for convenience & performance reasons
	currentExcludedRecipients := map[string]bool{}
	for _, email := range participant.ExcludedRecipients {
		currentExcludedRecipients[email] = true
	}

	// create items for the multi-select
	multiSelectExcludedParticipants := []promptMultiSelectItem{}
	for _, p := range participants {
		// skip the participant themself
		if p.Email == participant.Email {
			continue
		}
		multiSelectExcludedParticipants = append(multiSelectExcludedParticipants, promptMultiSelectItem{
			ID:         p.Email,
			IsSelected: currentExcludedRecipients[p.Email],
		})
	}
	excludedParticipants := promptMultiSelect(multiSelectExcludedParticipants, 1, "Exclude the following participants")

	// update the excluded recipients
	participant.ExcludedRecipients = make([]string, len(excludedParticipants))
	for i, item := range excludedParticipants {
		participant.ExcludedRecipients[i] = item.ID
	}

	return participants
}
