package main

import (
	"encoding/json"
	"io"
	"log"
	"os"
)

type Config struct {
	Email        Email         `json:"email"`
	Mailgun      Mailgun       `json:"mailgun"`
	Participants []Participant `json:"participants" validate:"min=2,dive"`
}

type Email struct {
	Body      string `json:"body" validate:"required"`
	Recipient string `json:"recipient"`
	Sender    string `json:"sender" validate:"required"`
	Subject   string `json:"subject" validate:"required"`
}

type Mailgun struct {
	Domain string `json:"domain" validate:"required"`
	ApiKey string `json:"apiKey" validate:"required"`
}

type Participant struct {
	Id                int    `json:"id" validate:"required"`
	Email             string `json:"email" validate:"required"`
	Salutation        string `json:"salutation" validate:"required"`
	ExcludedPersonIds []int  `json:"excludedPersonIds" validate:"required"`
}

func loadConfig() Config {
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
