package app

import (
	"context"
	"fmt"
	"time"

	"github.com/mailgun/mailgun-go/v4"
)

type mailer interface {
	// TODO (ask): would it be better to use an anonymous struct here
	// instead of ordered strings which are very easy to misplace?
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
