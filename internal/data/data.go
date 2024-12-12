package data

import (
	"encoding/json"
	"io"
	"os"
	"strconv"
	"strings"
)

// Data is the representation of the data for the app.
type Data struct {
	Schema       string        `json:"$schema"` // JSON schema to help with structure & validation
	Template     *Template     `json:"template"                  validate:"required"`
	Mailgun      *Mailgun      `json:"mailgun,omitempty"         validate:"required_without=Smtp,omitempty"`
	Smtp         *Smtp         `json:"smtp,omitempty"            validate:"required_without=Mailgun,omitempty"`
	Participants []Participant `json:"participants,omitempty"    validate:"required,min=2,dive"`
	Extras       []Extra       `json:"extraRecipients,omitempty" validate:"dive"`
}

// Template defines the email(s) to be sent out to participants.
type Template struct {
	Body                string `json:"body,omitempty"                validate:"required"`
	Subject             string `json:"subject,omitempty"             validate:"required"`
	Sender              string `json:"sender,omitempty"              validate:"required,email"`
	RecipientsSeparator string `json:"recipientsSeparator,omitempty"`
}

// Mailgun contains the Mailgun-related config.
type Mailgun struct {
	Domain string `json:"domain,omitempty" validate:"required"`
	APIKey string `json:"apiKey,omitempty" validate:"required"`
}

// Smtp contains the required config for sending via SMTP.
type Smtp struct {
	Host string `json:"host,omitempty" validate:"required"`
	User string `json:"user,omitempty" validate:"required"`
	Pass string `json:"pass,omitempty" validate:"required"`
}

// Person is a generic person in the Santa context, i.e. a participant or an extra.
type Person struct {
	Salutation string `json:"salutation" validate:"required"`
}

// Participant is the definition of each participant.
type Participant struct {
	Person
	Email                string   `json:"email"                          validate:"required,email"`
	ExcludedRecipients   []string `json:"excludedRecipients,omitempty"`
	PredestinedRecipient string   `json:"predestinedRecipient,omitempty"`
}

// Extra is the definition of each extra recipient.
type Extra struct {
	Person
	ExcludedGivers []string `json:"excludedGivers,omitempty"`
}

// LoadData loads data from a JSON file.
func LoadData(filePath string) (Data, error) {
	var data Data

	jsonFile, err := os.Open(filePath)
	if err != nil {
		return Data{}, err
	}

	defer jsonFile.Close()

	byteValue, _ := io.ReadAll(jsonFile)
	err = json.Unmarshal(byteValue, &data)
	if err != nil {
		return Data{}, err
	}

	return data, nil
}

// unescape JSON entities, making the result file human readable (and editable)
func unescapeUnicodeCharactersInJSON(_jsonRaw json.RawMessage) (json.RawMessage, error) {
	str, err := strconv.Unquote(strings.Replace(strconv.Quote(string(_jsonRaw)), `\\u`, `\u`, -1))
	if err != nil {
		return nil, err
	}
	return []byte(str), nil
}

// SaveData saves the given data into the JSON file at `filePath`.
func SaveData(filePath string, data Data) error {
	dataJson, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	unescapedDataJson, err := unescapeUnicodeCharactersInJSON(dataJson)
	if err != nil {
		return err
	}

	err = os.WriteFile(filePath, unescapedDataJson, 0644)
	if err != nil {
		return err
	}

	return nil
}
