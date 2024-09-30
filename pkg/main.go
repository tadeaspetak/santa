package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/mailgun/mailgun-go/v4"
)

var IS_DEBUG = true
var validate *validator.Validate

type pair struct {
	giver    Participant
	receiver Participant
}

func raffle(people []Participant) []pair {
	potentialReceivers := make([]Participant, len(people))
	copy(potentialReceivers, people)

	raffled := make([]pair, len(people))
	for i, giver := range people {
		attemptCount := 0
		receiverIndex := getRandomIndexInArray(potentialReceivers)
		receiver := potentialReceivers[receiverIndex]

		// ensure the receiver is not the giver, and also not in the excluded people for the giver
		for giver.Id == receiver.Id || contains(giver.ExcludedPersonIds, receiver.Id) {
			if attemptCount >= 5 {
				if IS_DEBUG {
					log.Println("Too many failed attempts to pick a receiver, let's start anew.")
				}
				return raffle(people)
			}
			attemptCount += 1

			receiverIndex = getRandomIndexInArray(potentialReceivers)
			receiver = potentialReceivers[receiverIndex]
		}

		raffled[i] = pair{giver: giver, receiver: receiver}

		// remove the receiver from the potential ones
		potentialReceivers = append(potentialReceivers[:receiverIndex], potentialReceivers[receiverIndex+1:]...)

	}

	return raffled
}

func main() {
	validate = validator.New()

	config := loadConfig()
	mg := mailgun.NewMailgun(config.Mailgun.Domain, config.Mailgun.ApiKey)
	mg.SetAPIBase("https://api.eu.mailgun.net/v3")

	raffled := raffle(config.Participants)
	for _, pair := range raffled {
		// prefer the email provided in the config (for testing purposes)
		recipient := config.Email.Recipient
		if recipient == "" {
			recipient = pair.giver.Email
		}

		message := mg.NewMessage(config.Email.Sender, config.Email.Subject, "", recipient)

		// construct the message body
		replacer := strings.NewReplacer("%{receiverSalutation}", pair.receiver.Salutation, "%{recipientEmail}", pair.giver.Email)
		body := replacer.Replace(fmt.Sprintf(`<html><body>%s</body></html>`, config.Email.Body))
		message.SetHtml(body)

		if IS_DEBUG {
			fmt.Printf("%s -> %s, recipient: %s\n", pair.giver.Email, pair.receiver.Salutation, recipient)
			fmt.Printf("%s\n\n", body)

			// don't send anything when debugging
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		defer cancel()

		resp, id, err := mg.Send(ctx, message)

		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("ID: %s, email: %s, Resp: %s\n", id, recipient, resp)
	}
}
