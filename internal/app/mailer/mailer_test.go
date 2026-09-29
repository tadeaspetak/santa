package mailer

import (
	"context"
	"reflect"
	"testing"

	"github.com/tadeaspetak/santa/internal/app"
	"github.com/tadeaspetak/santa/internal/testutil"
)

var (
	_ app.Mailer = smtpMailer{}
	_ app.Mailer = mailgunMailer{}
)

// mailerCase describes a mailer to test with a server behind it.
type mailerCase struct {
	name string
	// setup starts a fake server and returns a mailer with valid & invalid credentials for it,
	// plus a function returning the messages the server received
	setup func(t *testing.T) (valid, invalid app.Mailer, messages func() []testutil.Message)
}

var mailerCases = []mailerCase{
	{
		name: "smtp",
		setup: func(t *testing.T) (app.Mailer, app.Mailer, func() []testutil.Message) {
			server := testutil.NewSMTPServer(t, "user", "pass")
			return NewSmtpMailer(server.Host, server.Port, "user", "pass"),
				NewSmtpMailer(server.Host, server.Port, "user", "wrong"),
				server.Messages
		},
	},
	{
		name: "mailgun",
		setup: func(t *testing.T) (app.Mailer, app.Mailer, func() []testutil.Message) {
			server := testutil.NewMailgunServer(t, "mg.example.com", "key")
			return NewMailgunMailer("mg.example.com", "key", server.APIBase),
				NewMailgunMailer("mg.example.com", "wrong", server.APIBase),
				server.Messages
		},
	},
}

func TestMailerVerify(t *testing.T) {
	for _, tt := range mailerCases {
		t.Run(tt.name, func(t *testing.T) {
			valid, invalid, messages := tt.setup(t)

			if err := valid.Verify(context.Background()); err != nil {
				t.Errorf("expected valid credentials to verify, got %v", err)
			}
			if err := invalid.Verify(context.Background()); err == nil {
				t.Errorf("expected invalid credentials to fail verification")
			}
			if len(messages()) != 0 {
				t.Errorf("verifying must not send anything, got %v", messages())
			}
		})
	}
}

func TestMailerSend(t *testing.T) {
	for _, tt := range mailerCases {
		t.Run(tt.name, func(t *testing.T) {
			valid, invalid, messages := tt.setup(t)

			err := valid.Send("santa@example.com", "🎄 Gift for Mom", "<p>Mom</p>", "dad@example.com", "reply@example.com")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			expected := []testutil.Message{{
				From:    "santa@example.com",
				To:      "dad@example.com",
				ReplyTo: "reply@example.com",
				Subject: "🎄 Gift for Mom",
				HTML:    "<p>Mom</p>",
			}}
			if !reflect.DeepEqual(messages(), expected) {
				t.Fatalf("\nexpected: %+v\ngot: %+v", expected, messages())
			}

			if err := invalid.Send("santa@example.com", "s", "b", "dad@example.com", "r@example.com"); err == nil {
				t.Fatalf("expected sending with invalid credentials to fail")
			}
			if len(messages()) != 1 {
				t.Fatalf("expected nothing else to be sent, got %+v", messages())
			}
		})
	}
}
