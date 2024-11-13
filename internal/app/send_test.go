package app

import (
	"reflect"
	"testing"

	"github.com/tadeaspetak/santa/internal/data"
)

type mail struct {
	sender    string
	subject   string
	body      string
	recipient string
	replyTo   string
}

type mockMailer struct {
	mails []mail
}

func (m *mockMailer) send(sender, subject, body, recipient, replyTo string) error {
	m.mails = append(
		m.mails,
		mail{
			sender:    sender,
			subject:   subject,
			body:      body,
			recipient: recipient,
			replyTo:   replyTo},
	)

	return nil
}

func TestSend(t *testing.T) {
	p1 := data.Participant{Email: "p1", Salutation: "s1"}
	p2 := data.Participant{Email: "p2", Salutation: "s2"}
	p3 := data.Participant{Email: "p3", Salutation: "s3"}

	mailer := mockMailer{}
	Send(
		&mailer,
		[]participantPair{
			participantPair{giver: p1, recipient: p2},
			participantPair{giver: p2, recipient: p3},
			participantPair{giver: p3, recipient: p1},
		},
		data.Template{
			Subject: "sub %{recipientSalutation}",
			Body:    "bod %{recipientSalutation}",
			Sender:  "sender",
		},
		false,
		"",
	)
	expected := []mail{
		mail{"sender", "sub s2", "<html><body>bod s2</body></html>", "p1", "p1"},
		mail{"sender", "sub s3", "<html><body>bod s3</body></html>", "p2", "p2"},
		mail{"sender", "sub s1", "<html><body>bod s1</body></html>", "p3", "p3"},
	}
	if !reflect.DeepEqual(mailer.mails, expected) {
		t.Fatalf(`
TestSend failed
Expected: %v
Got: %v
    `, expected, mailer.mails)
	}

}
