package prompt

import (
	"fmt"
	"html/template"
	"log"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/tadeaspetak/secret-reindeer/internal/data"
)

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
		Active:   "👉 {{ .Email | cyan }} ({{ .Salutation }}, exc: {{ .ExcludedRecipients | stringsJoin }}, pre: {{ .PredestinedRecipient }})",
		Inactive: "   {{ .Email | cyan }} ({{ .Salutation }}, exc: {{ .ExcludedRecipients | stringsJoin }}, pre: {{ .PredestinedRecipient }})",
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
		Searcher:  searcher,
	}

	index, _, err := prompt.Run()

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return -1
	}

	return index
}
