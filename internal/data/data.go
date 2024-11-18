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

// Data is the representation of the data for the app.
type Data struct {
	Template     Template      `json:"template"`
	Mailgun      Mailgun       `json:"mailgun"`
	Participants []Participant `json:"participants,omitempty" validate:"min=2,dive"`
}

// Template defines the email(s) to be sent out to participants.
type Template struct {
	Body    string `json:"body,omitempty"    validate:"required"`
	Subject string `json:"subject,omitempty" validate:"required"`
	Sender  string `json:"sender,omitempty"  validate:"required,email"`
}

// Mailgun contains the Mailgun-related config.
type Mailgun struct {
	Domain string `json:"domain,omitempty" validate:"required"`
	APIKey string `json:"apiKey,omitempty" validate:"required"`
}

// Participant is the definition of each participant.
type Participant struct {
	Email                string   `json:"email"                          validate:"required,email"`
	Salutation           string   `json:"salutation"                     validate:"required"`
	ExcludedRecipients   []string `json:"excludedRecipients,omitempty"`
	PredestinedRecipient string   `json:"predestinedRecipient,omitempty"`
}

// TODO (ask): Should the received by a pointer? Since the struct contains a slice, the copy of the struct
// will contain a copy of the slice address. Modifying that will modify the original data
// but modifying other regular fields would **not** modify the original data. For clarity, I reckon
// this should be a pointer.

// UpdateParticipantEmail updates the email address of the participant at the given index.
func (d *Data) UpdateParticipantEmail(participantIndex int, curr string, next string) error {
	if participantIndex >= len(d.Participants) {
		return fmt.Errorf("participant with index %v does not exist", participantIndex)
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

// RemoveParticipant removes a participant at the given index.
func (d *Data) RemoveParticipant(participantIndex int) error {
	if participantIndex >= len(d.Participants) {
		return fmt.Errorf("participant with index %v does not exist", participantIndex)
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

// LoadData loads data from a JSON file.
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

// unescape JSON entities, making the result file human readable (and editable)
func unescapeUnicodeCharactersInJSON(_jsonRaw json.RawMessage) (json.RawMessage, error) {
	str, err := strconv.Unquote(strings.Replace(strconv.Quote(string(_jsonRaw)), `\\u`, `\u`, -1))
	if err != nil {
		return nil, err
	}
	return []byte(str), nil
}

// TODO (ask): should this be a method on `Data`? or is it better to stick to utils methods
// SaveData saves the given data into the JSON file at `filePath`.
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
