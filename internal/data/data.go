package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"slices"
	"strconv"
	"strings"
)

// Data for the app
type Data struct {
	Template     Template      `json:"template"`
	Mailgun      Mailgun       `json:"mailgun"`
	Participants []Participant `json:"participants" validate:"min=2,dive"`
}

// Template props for the email to be sent out
type Template struct {
	Body    string `json:"body"    validate:"required"`
	Subject string `json:"subject" validate:"required"`
	Sender  string `json:"sender"  validate:"required,email"`
}

// Mailgun config
type Mailgun struct {
	Domain string `json:"domain" validate:"required"`
	APIKey string `json:"apiKey" validate:"required"`
}

// Participant definition
type Participant struct {
	Email                string   `json:"email"                          validate:"required,email"`
	Salutation           string   `json:"salutation"                     validate:"required"`
	ExcludedRecipients   []string `json:"excludedRecipients,omitempty"`
	PredestinedRecipient string   `json:"predestinedRecipient,omitempty"`
}

// TODO (ask): should this be a pointer or not? since the struct contains a slice, the copy of the struct
// will also contain a copy of that pointer (address); modifying that will modify the original data
// but modifying other regular fields would **not** modify the original data; for clarity, I reckon
// this should be a pointer
func (d *Data) UpdateParticipantEmail(participantIndex int, curr string, next string) error {
	if participantIndex >= len(d.Participants) {
		return fmt.Errorf("Participant with index %v does not exist", participantIndex)
	}

	(&d.Participants[participantIndex]).Email = next

	// ensure the email is also replaced in all excluded recipients
	for index := range d.Participants {
		participant := &d.Participants[index]
		if emailIndex := slices.Index(participant.ExcludedRecipients, curr); emailIndex > -1 {
			participant.ExcludedRecipients[emailIndex] = next
		}
	}

	return nil
}

// remove an element at a given index from a slice
// while preserving order (https://stackoverflow.com/a/37335777/3844098).
func removeFromSlice[K any](slice []K, index int) []K {
	return append(slice[:index], slice[index+1:]...)
}

func (d *Data) RemoveParticipant(participantIndex int) error {
	if participantIndex >= len(d.Participants) {
		return fmt.Errorf("Participant with index %v does not exist", participantIndex)
	}
	removedParticipantEmail := d.Participants[participantIndex].Email

	d.Participants = removeFromSlice(d.Participants, participantIndex)

	// ensure the email is also removed in all excluded recipients
	for index := range d.Participants {
		participant := &d.Participants[index]
		if emailIndex := slices.Index(participant.ExcludedRecipients, removedParticipantEmail); emailIndex > -1 {
			participant.ExcludedRecipients = removeFromSlice(participant.ExcludedRecipients, emailIndex)
		}
	}

	return nil
}

// LoadData load data from a JSON file
func LoadData(filePath string) Data {
	var data Data

	jsonFile, err := os.Open(filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return data
		}

		log.Fatal("Error loading config", err)
	}

	defer jsonFile.Close()

	byteValue, _ := io.ReadAll(jsonFile)
	err = json.Unmarshal(byteValue, &data)
	if err != nil {
		log.Fatal(err)
	}

	return data
}

// TODO: comment
func unescapeUnicodeCharactersInJSON(_jsonRaw json.RawMessage) (json.RawMessage, error) {
	str, err := strconv.Unquote(strings.Replace(strconv.Quote(string(_jsonRaw)), `\\u`, `\u`, -1))
	if err != nil {
		return nil, err
	}
	return []byte(str), nil
}

// TODO (ask): should this be a method on `Data`? or is it better to stick to utils methods
func SaveData(filePath string, data Data) {
	dataJson, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	unescapedDataJson, err := unescapeUnicodeCharactersInJSON(dataJson)
	if err != nil {
		log.Fatal(err)
	}

	err = os.WriteFile(filePath, unescapedDataJson, 0644)

	if err != nil {
		log.Fatal("Error writing data", err)
	}
}
