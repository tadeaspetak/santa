package app

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/tadeaspetak/santa/internal/data"
)

func Send(mlr mailer, pairs []participantPair, template data.Template, isDebug bool, alwaysSendTo string) error {
	batchDate := time.Now().Local().Format("20060102-150405")

	for _, pair := range pairs {
		// prefer the email provided via the `alwaysSendTo` flag for testing purposes
		recipient := alwaysSendTo
		if recipient == "" {
			recipient = pair.giver.Email
		}

		replacer := strings.NewReplacer("%{recipientSalutation}", pair.recipient.Salutation)
		subject := replacer.Replace(template.Subject)
		body := replacer.Replace(fmt.Sprintf(`<html><body>%s</body></html>`, template.Body))

		// write a history batch file
		err := os.WriteFile(
			fmt.Sprintf("santa-batch-%s-%s.txt", batchDate, pair.giver.Email),
			[]byte(fmt.Sprintf("Sent to: %s\nSubject: %s\nBody: %s", pair.giver.Email, subject, body)),
			0644,
		)
		if err != nil {
			return fmt.Errorf("unable to write a history batch file %v", err)
		}

		// don't send anything when debugging
		if isDebug {
			continue
		}

		err = mlr.send(
			template.Sender,
			subject,
			body,
			recipient,
			pair.giver.Email, // even when a fixed recipient is present, set the reply-to to the actual giver's email to make debugging easy
		)

		if err != nil {
			// Note: It's questionable what the best course of action is here. Is it better to continue
			// with the current batch or return an error even though some emails may already have been sent out?
			return fmt.Errorf("could not send email to %v: %v", recipient, err)
		}

		fmt.Printf("Email to %s sent successfully\n", recipient)

	}

	return nil
}
