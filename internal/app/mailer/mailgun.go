package mailer

import (
	"context"
	"time"

	"github.com/mailgun/mailgun-go/v4"
)

type mailgunMailer struct {
	gun *mailgun.MailgunImpl
}

// MailgunEUAPIBase is the API base of Mailgun's EU region, the only one supported for now.
const MailgunEUAPIBase = "https://api.eu.mailgun.net/v3"

func NewMailgunMailer(domain, apiKey, apiBase string) mailgunMailer {
	gun := mailgun.NewMailgun(domain, apiKey)
	gun.SetAPIBase(apiBase)
	return mailgunMailer{gun: gun}
}

func (m mailgunMailer) Send(sender, subject, body, recipient, replyTo string) error {
	message := m.gun.NewMessage(sender, subject, "", recipient)
	message.SetReplyTo(replyTo)
	message.SetHtml(body)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	_, _, err := m.gun.Send(ctx, message)
	return err
}

// Verify checks the API key and the sending domain without sending anything.
func (m mailgunMailer) Verify(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	_, err := m.gun.GetDomain(ctx, m.gun.Domain())
	return err
}
