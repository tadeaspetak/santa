package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/mailgun/mailgun-go/v4"
	"github.com/tadeaspetak/secret-santa-go/internal/config"
	"github.com/tadeaspetak/secret-santa-go/internal/utils"
)

type pair struct {
	giver    config.Participant
	receiver config.Participant
}

func raffle(participants []config.Participant, isDebug bool) []pair {
	potentialReceivers := make([]config.Participant, len(participants))
	copy(potentialReceivers, participants)

	raffled := make([]pair, len(participants))
	for i, giver := range participants {
		attemptCount := 0
		receiverIndex := utils.GetRandomIndexInArray(potentialReceivers)
		receiver := potentialReceivers[receiverIndex]

		// ensure the receiver is not the giver, and also not in the excluded people for the giver
		for giver.ID == receiver.ID || utils.Contains(giver.ExcludedPersonIds, receiver.ID) {
			if attemptCount >= 5 {
				if isDebug {
					log.Println("Too many failed attempts to pick a receiver, let's start anew.")
				}
				return raffle(participants, isDebug)
			}
			attemptCount++

			receiverIndex = utils.GetRandomIndexInArray(potentialReceivers)
			receiver = potentialReceivers[receiverIndex]
		}

		raffled[i] = pair{giver: giver, receiver: receiver}

		// remove the receiver from the potential ones
		potentialReceivers = append(potentialReceivers[:receiverIndex], potentialReceivers[receiverIndex+1:]...)

	}

	return raffled
}

func main() {

	config := config.LoadConfig()
	mg := mailgun.NewMailgun(config.Mailgun.Domain, config.Mailgun.APIKey)
	mg.SetAPIBase("https://api.eu.mailgun.net/v3")

	raffled := raffle(config.Participants, config.IsDebug)
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

		if config.IsDebug {
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
