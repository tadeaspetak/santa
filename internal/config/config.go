package config

import (
	"encoding/json"
	"io"
	"log"
	"os"

	"github.com/go-playground/validator"
)

// Config for the app
type Config struct {
	IsDebug      bool          `json:"isDebug"`
	Email        Email         `json:"email"`
	Mailgun      Mailgun       `json:"mailgun"`
	Participants []Participant `json:"participants" validate:"min=2,dive"`
}

// Email props for the email to be sent out
type Email struct {
	Body      string `json:"body" validate:"required"`
	Recipient string `json:"recipient"`
	Sender    string `json:"sender" validate:"required"`
	Subject   string `json:"subject" validate:"required"`
}

// Mailgun config
type Mailgun struct {
	Domain string `json:"domain" validate:"required"`
	APIKey string `json:"apiKey" validate:"required"`
}

// Participant definition
type Participant struct {
	ID                int    `json:"id" validate:"required"`
	Email             string `json:"email" validate:"required"`
	Salutation        string `json:"salutation" validate:"required"`
	ExcludedPersonIds []int  `json:"excludedPersonIds" validate:"required"`
}

// LoadConfig load configuration from a JSON file
func LoadConfig() Config {
	validate := validator.New()

	jsonFile, err := os.Open("config.json")
	if err != nil {
		log.Fatal("Error loading config", err)
	}

	defer jsonFile.Close()

	byteValue, _ := io.ReadAll(jsonFile)
	var config Config
	err = json.Unmarshal(byteValue, &config)
	if err != nil {
		log.Fatal(err)
	}

	err = validate.Struct(config)
	if err != nil {
		log.Fatal(config, err)
	}

	return config
}
