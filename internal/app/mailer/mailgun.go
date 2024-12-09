package mailer

import (
	"context"
	"fmt"
	"time"

	"github.com/mailgun/mailgun-go/v4"
)

type mailgunMailer struct {
	gun *mailgun.MailgunImpl
}

func NewMailgunMailer(domain string, apiKey string) mailgunMailer {
	gun := mailgun.NewMailgun(domain, apiKey)
	gun.SetAPIBase("https://api.eu.mailgun.net/v3")
	return mailgunMailer{gun: gun}
}

func (m mailgunMailer) Send(sender, subject, body, recipient, replyTo string) error {
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
