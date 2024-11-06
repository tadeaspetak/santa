package app

import (
	"fmt"
	"log"
	"strings"

	"github.com/tadeaspetak/secret-reindeer/internal/data"
)

func Send(mlr mailer, pairs []participantPair, template data.Template, isDebug bool, alwaysSendTo string) {
	for _, pair := range pairs {
		// prefer the email provided via the flag for testing purposes
		recipient := alwaysSendTo
		if recipient == "" {
			recipient = pair.giver.Email
		}

		replacer := strings.NewReplacer("%{recipientSalutation}", pair.recipient.Salutation)
		subject := replacer.Replace(template.Subject)
		body := replacer.Replace(fmt.Sprintf(`<html><body>%s</body></html>`, template.Body))

		// don't send anything when debugging
		if isDebug {
			fmt.Printf("%s -> %s, email recipient: %s\n", pair.giver.Email, pair.recipient.Salutation, recipient)
			fmt.Printf("%s\n\n", body)

			continue
		}

		err := mlr.send(
			template.Sender,
			subject,
			body,
			recipient,
			pair.giver.Email, // even when a fixed recipient is present, set the reply-to to the actual giver's email to make debugging easy
		)

		if err != nil {
			log.Fatalf("Could not send email to %v: %v\n", recipient, err)
		}

		fmt.Printf("Email to %s sent successfully\n", recipient)
	}
}
