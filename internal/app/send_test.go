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
	p1 := data.Participant{Person: data.Person{Salutation: "s1"}, Email: "p1"}
	p2 := data.Participant{Person: data.Person{Salutation: "s2"}, Email: "p2"}
	p3 := data.Participant{Person: data.Person{Salutation: "s3"}, Email: "p3"}

	e4 := data.Extra{Person: data.Person{Salutation: "e4"}}
	e5 := data.Extra{Person: data.Person{Salutation: "e5"}}

	mailer := mockMailer{}
	Send(
		&mailer,
		[]giverWithRecipients{
			{giver: p1, recipients: []data.Person{p2.Person, e4.Person}},
			{giver: p2, recipients: []data.Person{p3.Person, e5.Person}},
			{giver: p3, recipients: []data.Person{p1.Person}},
		},
		data.Template{
			Subject:             "sub %{recipientSalutation}",
			Body:                "bod %{recipientSalutation}",
			Sender:              "sender",
			RecipientsSeparator: ", ",
		},
		SendOpts{},
	)
	expected := []mail{
		{"sender", "sub s2, e4", "<html><body>bod s2, e4</body></html>", "p1", "p1"},
		{"sender", "sub s3, e5", "<html><body>bod s3, e5</body></html>", "p2", "p2"},
		{"sender", "sub s1", "<html><body>bod s1</body></html>", "p3", "p3"},
	}
	if !reflect.DeepEqual(mailer.mails, expected) {
		t.Fatalf(`
TestSend failed
Expected: %v
Got: %v
    `, expected, mailer.mails)
	}

}
