package app

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/mailgun/mailgun-go/v4"
	"github.com/tadeaspetak/secret-reindeer/internal/data"
)

type mailer interface {
	// TODO (ask): would it be better to use an anonymous struct here?
	send(sender, subject, body, recipient, replyTo string) error
}

type mailgunMailer struct {
	gun *mailgun.MailgunImpl
}

func NewMailgunMailer(domain string, apiKey string) mailgunMailer {
	gun := mailgun.NewMailgun(domain, apiKey)
	gun.SetAPIBase("https://api.eu.mailgun.net/v3")
	return mailgunMailer{gun: gun}
}

func (m mailgunMailer) send(sender, subject, body, recipient, replyTo string) error {
	message := m.gun.NewMessage(sender, subject, "", recipient)
	message.SetReplyTo(replyTo)
	message.SetHtml(body)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	resp, id, err := m.gun.Send(ctx, message)

	if err != nil {
		return err
	}

	fmt.Printf("ID: %s, email: %s, Resp: %s\n", id, recipient, resp)
	return nil

}

func Send(mlr mailer, pairs []raffledPair, template data.Template, isDebug bool, fixedRecipient string) {
	for _, pair := range pairs {
		// prefer the email provided via the flag for testing purposes
		recipient := fixedRecipient
		if recipient == "" {
			recipient = pair.giver.Email
		}

		replacer := strings.NewReplacer("%{recipientSalutation}", pair.recipient.Salutation)
		subject := replacer.Replace(template.Subject)
		body := replacer.Replace(fmt.Sprintf(`<html><body>%s</body></html>`, template.Body))

		if isDebug {
			fmt.Printf("%s -> %s, email recipient: %s\n", pair.giver.Email, pair.recipient.Salutation, recipient)
			fmt.Printf("%s\n\n", body)

			// don't send anything when debugging
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
