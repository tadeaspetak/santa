package prompt

import (
	"fmt"
	"html/template"
	"log"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/tadeaspetak/secret-reindeer/internal/data"
	"github.com/tadeaspetak/secret-reindeer/internal/utils"
)

func PromptStringNew(label string) string {
	prompt := promptui.Prompt{
		Label: label,
	}

	value, err := prompt.Run()
	if err != nil {
		log.Fatalf("Prompt failed %v.\n", err)
	}

	return value
}

func PromptStringEdit(label string, currentValue string) string {
	prompt := promptui.Prompt{
		Label:     label,
		Default:   currentValue,
		AllowEdit: true,
	}

	nextValue, err := prompt.Run()
	if err != nil {
		log.Fatalf("Prompt failed %v\n", err)
	}

	return nextValue
}

func PromptSelectParticipant(participants []data.Participant, selectedLabel string) int {
	// add a function to the template FuncMap, and manually add the color functions
	// https://github.com/manifoldco/promptui/blob/c2e487d3597f59bcf76b24c9e80679740a72212b/prompt.go#L101
	funcMap := template.FuncMap{
		"stringsJoin": func(slice []string) string { return strings.Join(slice, ",") },
	}
	for k, v := range promptui.FuncMap {
		funcMap[k] = v
	}

	templates := &promptui.SelectTemplates{
		Label:    "{{ . }}",
		Active:   "👉 {{ .Email | cyan }} ({{ .Salutation }}, Excluded: {{ .ExcludedRecipients | stringsJoin }}, Predestined: {{ .PredestinedRecipient }})",
		Inactive: "   {{ .Email | cyan }} ({{ .Salutation }}, Excluded: {{ .ExcludedRecipients | stringsJoin }}, Predestined: {{ .PredestinedRecipient }})",
		Selected: fmt.Sprintf("%s {{ .Email }}", selectedLabel),
		FuncMap:  funcMap,
	}

	searcher := func(input string, index int) bool {
		participant := participants[index]
		email := strings.Replace(strings.ToLower(participant.Email), " ", "", -1)
		name := strings.Replace(strings.ToLower(participant.Salutation), " ", "", -1)
		input = strings.Replace(strings.ToLower(input), " ", "", -1)

		return strings.Contains(name, input) || strings.Contains(email, input)
	}

	prompt := promptui.Select{
		Label:     "Which participant would you like to manage?",
		Items:     participants,
		Templates: templates,
		Size:      5,
		Searcher:  searcher,
	}

	index, _, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return -1
	}

	return index
}

type PromptMultiSelectItem struct {
	ID         string
	IsSelected bool
}

func PromptMultiSelect(items []PromptMultiSelectItem, selectedPosition int, label string) ([]PromptMultiSelectItem, error) {
	// if the `done` item doesn't exist yet, prepend it
	const doneID = "Done"
	if len(items) > 0 && items[0].ID != doneID {
		items = append([]PromptMultiSelectItem{{ID: doneID}}, items...)
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
		return PromptMultiSelect(items, index, label)
	}

	selectedItems := utils.Filter(items, func(i PromptMultiSelectItem, _ int) bool { return i.IsSelected })
	return selectedItems, nil
}
