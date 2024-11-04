package edit

import (
	"fmt"

	"github.com/manifoldco/promptui"
	"github.com/tadeaspetak/secret-reindeer/internal/data"
	"github.com/tadeaspetak/secret-reindeer/internal/utils"
)

type promptMultiSelectItem struct {
	ID         string
	IsSelected bool
}

func promptMultiSelect(items []promptMultiSelectItem, selectedPosition int, label string) ([]promptMultiSelectItem, error) {
	// if the `done` item doesn't exist yet, prepend it
	const doneID = "Done"
	if len(items) > 0 && items[0].ID != doneID {
		items = append([]promptMultiSelectItem{{ID: doneID}}, items...)
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
		return nil, fmt.Errorf("prompt failed: %w", err)
	}

	selectedItem := &items[index]
	if selectedItem.ID != doneID {
		// unless `done` has been selected, toggle the current item and display again
		selectedItem.IsSelected = !selectedItem.IsSelected
		return promptMultiSelect(items, index, label)
	}

	selectedItems := utils.Filter(items, func(i promptMultiSelectItem, _ int) bool { return i.IsSelected })
	return selectedItems, nil
}

func editExcludedRecipients(participants []data.Participant, participantIndex int) []data.Participant {
	participant := &participants[participantIndex]

	// create a map for convenience & performance reasons
	currentExcludedRecipients := map[string]bool{}
	for _, email := range participant.ExcludedRecipients {
		currentExcludedRecipients[email] = true
	}

	// create items for the multi-select
	multiSelectParticipants := []promptMultiSelectItem{}
	for _, p := range participants {
		// skip the participant themself
		if p.Email == participant.Email {
			continue
		}
		multiSelectParticipants = append(multiSelectParticipants, promptMultiSelectItem{
			ID:         p.Email,
			IsSelected: currentExcludedRecipients[p.Email],
		})
	}
	excludedParticipants, _ := promptMultiSelect(multiSelectParticipants, 1, "Exclude the following participants")

	// update the excluded recipients
	participant.ExcludedRecipients = make([]string, len(excludedParticipants))
	for i, item := range excludedParticipants {
		participant.ExcludedRecipients[i] = item.ID
	}

	return participants
}
