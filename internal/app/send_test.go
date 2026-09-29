package app

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
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
	// failFor makes sending to the given recipient fail
	failFor string
}

var errMockSend = errors.New("mock send failure")

func (m *mockMailer) Send(sender, subject, body, recipient, replyTo string) error {
	if m.failFor != "" && recipient == m.failFor {
		return errMockSend
	}

	m.mails = append(m.mails, mail{sender, subject, body, recipient, replyTo})
	return nil
}

func (m *mockMailer) Verify(_ context.Context) error {
	return nil
}

// testOpts returns options writing into a temporary directory and discarding the output.
func testOpts(t *testing.T, opts SendOpts) SendOpts {
	t.Helper()
	opts.OutputDir = filepath.Join(t.TempDir(), "batch")
	opts.Out = io.Discard
	return opts
}

var (
	p1 = participant("p1")
	p2 = participant("p2")
	p3 = participant("p3")
	e4 = extra("e4")
	e5 = extra("e5")

	testPairs = []giverWithRecipients{
		{giver: p1, recipients: []data.Person{p2.Person, e4.Person}},
		{giver: p2, recipients: []data.Person{p3.Person, e5.Person}},
		{giver: p3, recipients: []data.Person{p1.Person}},
	}

	testTemplate = data.Template{
		Subject:             "sub %{recipientSalutation}",
		Body:                "bod %{recipientSalutation}",
		Sender:              "sender",
		RecipientsSeparator: ", ",
	}
)

func TestSend(t *testing.T) {
	defaultSeparatorTemplate := testTemplate
	defaultSeparatorTemplate.RecipientsSeparator = ""

	tests := []struct {
		name     string
		template data.Template
		opts     SendOpts
		expected []mail
	}{
		{
			name:     "basic",
			template: testTemplate,
			expected: []mail{
				{"sender", "sub p2, e4", "<html><body>bod p2, e4</body></html>", "p1", "p1"},
				{"sender", "sub p3, e5", "<html><body>bod p3, e5</body></html>", "p2", "p2"},
				{"sender", "sub p1", "<html><body>bod p1</body></html>", "p3", "p3"},
			},
		},
		{
			name:     "default separator",
			template: defaultSeparatorTemplate,
			expected: []mail{
				{"sender", "sub p2 and e4", "<html><body>bod p2 and e4</body></html>", "p1", "p1"},
				{"sender", "sub p3 and e5", "<html><body>bod p3 and e5</body></html>", "p2", "p2"},
				{"sender", "sub p1", "<html><body>bod p1</body></html>", "p3", "p3"},
			},
		},
		{
			name:     "always send to",
			template: testTemplate,
			opts:     SendOpts{To: "me"},
			// the reply-to still points to the actual giver
			expected: []mail{
				{"sender", "sub p2, e4", "<html><body>bod p2, e4</body></html>", "me", "p1"},
				{"sender", "sub p3, e5", "<html><body>bod p3, e5</body></html>", "me", "p2"},
				{"sender", "sub p1", "<html><body>bod p1</body></html>", "me", "p3"},
			},
		},
		{
			name:     "debug sends nothing",
			template: testTemplate,
			opts:     SendOpts{IsDryRun: true},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mailer := mockMailer{}
			if err := Send(&mailer, testPairs, tt.template, testOpts(t, tt.opts)); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(mailer.mails, tt.expected) {
				t.Fatalf("\nexpected: %v\ngot: %v", tt.expected, mailer.mails)
			}
		})
	}
}

func TestSendWritesBatchFiles(t *testing.T) {
	for _, isDryRun := range []bool{false, true} {
		t.Run(map[bool]string{false: "sending", true: "dry run"}[isDryRun], func(t *testing.T) {
			opts := testOpts(t, SendOpts{IsDryRun: isDryRun})

			if err := Send(&mockMailer{}, testPairs, testTemplate, opts); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			entries, err := os.ReadDir(opts.OutputDir)
			if err != nil {
				t.Fatalf("expected the output directory to be created: %v", err)
			}
			if len(entries) != len(testPairs) {
				t.Fatalf("expected %d files (and no PDF), got %v", len(testPairs), entries)
			}

			for _, pair := range testPairs {
				content, err := os.ReadFile(filepath.Join(opts.OutputDir, pair.giver.Email+".txt"))
				if err != nil {
					t.Fatalf("expected a batch file for %s: %v", pair.giver.Email, err)
				}
				if !strings.HasPrefix(string(content), "Sent to: "+pair.giver.Email+"\n") {
					t.Fatalf("unexpected batch file content for %s: %s", pair.giver.Email, content)
				}
			}
		})
	}
}

func TestSendPrintsPdf(t *testing.T) {
	template := testTemplate
	// symbols such as emojis break fpdf, make sure they're handled
	template.Subject = "🎄 " + template.Subject

	opts := testOpts(t, SendOpts{IsDryRun: true, ShouldPrintPdf: true})
	if err := Send(&mockMailer{}, testPairs, template, opts); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(opts.OutputDir, "emails.pdf"))
	if err != nil {
		t.Fatalf("expected a PDF: %v", err)
	}
	if !strings.HasPrefix(string(content), "%PDF") {
		t.Fatalf("the generated file is not a PDF")
	}
}

func TestSendStopsOnMailerError(t *testing.T) {
	mailer := mockMailer{failFor: "p2"}
	err := Send(&mailer, testPairs, testTemplate, testOpts(t, SendOpts{}))
	if !errors.Is(err, errMockSend) {
		t.Fatalf("expected the mailer error, got %v", err)
	}
	if !strings.Contains(err.Error(), "1 of 3 sent") {
		t.Fatalf("expected the error to say how many emails went out, got %v", err)
	}

	// the first email went out, the failing one and everything after it didn't
	if len(mailer.mails) != 1 || mailer.mails[0].recipient != "p1" {
		t.Fatalf("expected only the email to p1 to be sent, got %v", mailer.mails)
	}
}

func TestSendOutput(t *testing.T) {
	tests := []struct {
		name     string
		opts     SendOpts
		expected string
	}{
		{
			name:     "sending",
			expected: "  ✓ p1\n  ✓ p2\n  ✓ p3\n",
		},
		{
			name:     "to",
			opts:     SendOpts{To: "me"},
			expected: "  ✓ me (instead of p1)\n  ✓ me (instead of p2)\n  ✓ me (instead of p3)\n",
		},
		{
			name:     "dry run",
			opts:     SendOpts{IsDryRun: true},
			expected: "  - p1 (not sent, dry run)\n  - p2 (not sent, dry run)\n  - p3 (not sent, dry run)\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			opts := testOpts(t, tt.opts)
			opts.Out = &out

			if err := Send(&mockMailer{}, testPairs, testTemplate, opts); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if out.String() != tt.expected {
				t.Fatalf("\nexpected:\n%s\ngot:\n%s", tt.expected, out.String())
			}
		})
	}
}

func TestRender(t *testing.T) {
	subject, body := Render(testTemplate, []data.Person{{Salutation: "Rudolph"}, {Salutation: "Blitzen"}})
	if subject != "sub Rudolph, Blitzen" {
		t.Errorf("unexpected subject: %s", subject)
	}
	if body != "<html><body>bod Rudolph, Blitzen</body></html>" {
		t.Errorf("unexpected body: %s", body)
	}
}

func TestRemoveSymbols(t *testing.T) {
	tests := []struct {
		in       string
		expected string
	}{
		{"plain text", "plain text"},
		{"🎄 Find a gift", "Find a gift"},
		{"Your Santa 🎅", "Your Santa"},
		{"snowman ⛄ inside", "snowman  inside"},
		{"Příliš žluťoučký kůň", "Příliš žluťoučký kůň"},
		{"", ""},
	}

	for _, tt := range tests {
		if got := removeSymbols(tt.in); got != tt.expected {
			t.Errorf("removeSymbols(%q) = %q, expected %q", tt.in, got, tt.expected)
		}
	}
}
