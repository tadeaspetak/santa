package mailer

import (
	"gopkg.in/gomail.v2"
)

type smtpMailer struct {
	dialer *gomail.Dialer
}

func NewSmtpMailer(host string, port int, user, pass string) smtpMailer {
	return smtpMailer{dialer: gomail.NewDialer(host, port, user, pass)}
}

func (m smtpMailer) Send(sender, subject, body, recipient, replyTo string) error {
	message := gomail.NewMessage()
	message.SetHeader("From", sender)
	message.SetHeader("To", recipient)
	message.SetHeader("Subject", subject)
	message.SetBody("text/html", body)
	message.SetHeader("Reply-To", replyTo)

	if err := m.dialer.DialAndSend(message); err != nil {
		return err
	}

	return nil
}
