package app

import (
	"context"
	"fmt"
	"log"
	"slices"
	"strings"
	"time"

	"github.com/mailgun/mailgun-go/v4"
	"github.com/tadeaspetak/secret-reindeer/internal/data"
	"github.com/tadeaspetak/secret-reindeer/internal/utils"
)

type raffledPair struct {
	giver     data.Participant
	recipient data.Participant
}

func raffle(participants []data.Participant, maxAttemptCount int) []raffledPair {
	if maxAttemptCount <= 0 {
		log.Fatalf("Too many failed attempts, try again or adjust the settings.")
	}

	// all participants by mail for convenience
	participantsByMail := map[string]data.Participant{}
	for _, p := range participants {
		participantsByMail[p.Email] = p
	}

	// collect the predestined participants
	predestined := make([]string, 0)
	for _, p := range participants {
		if p.PredestinedRecipient != "" {
			predestined = append(predestined, p.PredestinedRecipient)
		}
	}

	// potential recipients start as everyone except the predestined ones
	potentialRecipients := make([]data.Participant, 0)
	for _, p := range participants {
		if slices.Index(predestined, p.Email) == -1 {
			potentialRecipients = append(potentialRecipients, p)
		}
	}

	raffled := make([]raffledPair, len(participants))
	for i, giver := range participants {
		// the recipient is predestined
		if giver.PredestinedRecipient != "" {
			recipient, ok := participantsByMail[giver.PredestinedRecipient]
			if !ok {
				log.Fatalf("Predestined participant %s does not exist.", giver.PredestinedRecipient)
			}
			raffled[i] = raffledPair{giver: giver, recipient: recipient}
			continue
		}

		actualPotentialRecipients := make([]data.Participant, 0)
		for _, potentialRecipient := range potentialRecipients {
			if giver.Email != potentialRecipient.Email && slices.Index(giver.ExcludedRecipients, potentialRecipient.Email) == -1 {
				actualPotentialRecipients = append(actualPotentialRecipients, potentialRecipient)
			}
		}

		if len(actualPotentialRecipients) == 0 {
			fmt.Printf("No recipients for %s, let's try again.\n\n\n\n\n", giver.Email)
			return raffle(participants, maxAttemptCount-1)
		}

		// get a random recipient, add to the raffled and remove from potential
		actualRecipientIndex := utils.GetRandomIndexInArray(actualPotentialRecipients)
		actualRecipient := actualPotentialRecipients[actualRecipientIndex]
		raffled[i] = raffledPair{giver: giver, recipient: actualRecipient}

		potentialRecipientIndex := slices.IndexFunc(potentialRecipients, func(p data.Participant) bool { return p.Email == actualRecipient.Email })
		potentialRecipients = append(potentialRecipients[:potentialRecipientIndex], potentialRecipients[potentialRecipientIndex+1:]...)

	}

	return raffled
}

func Send(cmdData data.Data, isDebug bool, fixedRecipient string) {
	// set up mailgun
	mg := mailgun.NewMailgun(cmdData.Mailgun.Domain, cmdData.Mailgun.APIKey)
	mg.SetAPIBase("https://api.eu.mailgun.net/v3")

	raffled := raffle(cmdData.Participants, 5)
	for _, pair := range raffled {
		// prefer the email provided in the config (for testing purposes)
		recipient := fixedRecipient
		if recipient == "" {
			recipient = pair.giver.Email
		}

		message := mg.NewMessage(cmdData.Mailgun.Sender, cmdData.Template.Subject, "", recipient)
		message.SetReplyTo(pair.giver.Email) // set the reply-to to the actual giver email to make debugging easy

		// construct the message body
		replacer := strings.NewReplacer("%{recipientSalutation}", pair.recipient.Salutation, "%{recipientEmail}", pair.giver.Email)
		body := replacer.Replace(fmt.Sprintf(`<html><body>%s</body></html>`, cmdData.Template.Body))
		message.SetHtml(body)

		if isDebug {
			fmt.Printf("%s -> %s, email recipient: %s\n", pair.giver.Email, pair.recipient.Salutation, recipient)
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
