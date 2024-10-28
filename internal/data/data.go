package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"slices"

	"github.com/tadeaspetak/secret-reindeer/internal/utils"
	"github.com/tadeaspetak/secret-reindeer/internal/validation"
)

// Data for the app
type Data struct {
	IsDebug      bool          `json:"isDebug"`
	Email        Email         `json:"email"`
	Mailgun      Mailgun       `json:"mailgun"`
	Participants []Participant `json:"participants" validate:"min=2,dive"`
}

// Email props for the email to be sent out
type Email struct {
	Body      string `json:"body" validate:"required"`
	Recipient string `json:"recipient,email"` // TODO: explain this
	Sender    string `json:"sender" validate:"required,email"`
	Subject   string `json:"subject" validate:"required"`
}

// Mailgun config
type Mailgun struct {
	Domain string `json:"domain" validate:"required"`
	APIKey string `json:"apiKey" validate:"required"`
}

// Participant definition
type Participant struct {
	Email              string   `json:"email" validate:"required,email"`
	Salutation         string   `json:"salutation" validate:"required"`
	ExcludedRecipients []string `json:"excludedRecipients,omitempty"`
}

// TODO (ask): should this be a pointer or not? since the struct contains a slice, the copy of the struct
// will also contain a copy of that pointer (address); modifying that will modify the original data
// but modifying a other regular fields would **not** modify the original data; for clarity, I think
// this should be a pointer
func (d *Data) UpdateParticipantEmail(participantIndex int, curr string, next string) error {
	if participantIndex >= len(d.Participants) {
		return errors.New(fmt.Sprintf("Participant with index %v does not exist.", participantIndex))
	}

	(&d.Participants[participantIndex]).Email = next

	// ensure the email is also replaced in all excluded recipients
	for index, _ := range d.Participants {
		participant := &d.Participants[index]
		if emailIndex := slices.Index(participant.ExcludedRecipients, curr); emailIndex > -1 {
			participant.ExcludedRecipients[emailIndex] = next
		}
	}

	return nil
}

func (d *Data) RemoveParticipant(participantIndex int) error {
	if participantIndex >= len(d.Participants) {
		return errors.New(fmt.Sprintf("Participant with index %v does not exist.", participantIndex))
	}
	removedParticipantEmail := d.Participants[participantIndex].Email

	d.Participants = utils.RemoveFromSlice(d.Participants, participantIndex)

	// ensure the email is also removed in all excluded recipients
	for index, _ := range d.Participants {
		participant := &d.Participants[index]
		if emailIndex := slices.Index(participant.ExcludedRecipients, removedParticipantEmail); emailIndex > -1 {
			participant.ExcludedRecipients = utils.RemoveFromSlice(participant.ExcludedRecipients, emailIndex)
		}
	}

	return nil
}

// LoadData load data from a JSON file
func LoadData(filePath string) Data {
	jsonFile, err := os.Open(filePath)
	if err != nil {
		log.Fatal("Error loading config", err)
	}

	defer jsonFile.Close()

	byteValue, _ := io.ReadAll(jsonFile)
	var data Data
	err = json.Unmarshal(byteValue, &data)
	if err != nil {
		log.Fatal(err)
	}

	// TODO: only validate on demqnd
	err = validation.Validate.Struct(data)
	if err != nil {
		log.Fatal(data, err)
	}

	return data
}

// TODO (ask): should this be a method on `Data`? or is it better to stick to utils methods
func SaveData(filePath string, data Data) {
	// TODO: only validate on demand / when loading before sending (to allow partial data to exist)
	err := validation.Validate.Struct(data)
	if err != nil {
		log.Fatal(data, err)
	}

	dataJson, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	err = os.WriteFile(filePath, dataJson, 0644)

	if err != nil {
		log.Fatal("Error writing data", err)
	}
}
