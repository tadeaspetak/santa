package app

import (
	"errors"
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

// inTempDir runs the test in a temporary working directory so that the batch files
// and PDFs `Send` generates don't pollute the repo. Returns the directory.
func inTempDir(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(wd); err != nil {
			t.Fatal(err)
		}
	})

	return dir
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
			opts:     SendOpts{AlwaysSendTo: "me"},
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
			opts:     SendOpts{IsDebug: true},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inTempDir(t)

			mailer := mockMailer{}
			if err := Send(&mailer, testPairs, tt.template, tt.opts); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(mailer.mails, tt.expected) {
				t.Fatalf("\nexpected: %v\ngot: %v", tt.expected, mailer.mails)
			}
		})
	}
}

func TestSendWritesBatchFiles(t *testing.T) {
	for _, isDebug := range []bool{false, true} {
		t.Run(map[bool]string{false: "sending", true: "debug"}[isDebug], func(t *testing.T) {
			dir := inTempDir(t)

			if err := Send(&mockMailer{}, testPairs, testTemplate, SendOpts{IsDebug: isDebug}); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			for _, pair := range testPairs {
				matches, err := filepath.Glob(filepath.Join(dir, "santa-batch-*-"+pair.giver.Email+".txt"))
				if err != nil {
					t.Fatal(err)
				}
				if len(matches) != 1 {
					t.Fatalf("expected a single batch file for %s, got %v", pair.giver.Email, matches)
				}

				content, err := os.ReadFile(matches[0])
				if err != nil {
					t.Fatal(err)
				}
				if !strings.HasPrefix(string(content), "Sent to: "+pair.giver.Email+"\n") {
					t.Fatalf("unexpected batch file content for %s: %s", pair.giver.Email, content)
				}
			}

			if pdfs, _ := filepath.Glob(filepath.Join(dir, "*.pdf")); len(pdfs) != 0 {
				t.Fatalf("no PDF expected without the print flag, got %v", pdfs)
			}
		})
	}
}

func TestSendPrintsPdf(t *testing.T) {
	dir := inTempDir(t)

	template := testTemplate
	// symbols such as emojis break fpdf, make sure they're handled
	template.Subject = "🎄 " + template.Subject

	if err := Send(&mockMailer{}, testPairs, template, SendOpts{IsDebug: true, ShouldPrintPdf: true}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	pdfs, err := filepath.Glob(filepath.Join(dir, "santa-batch-*.pdf"))
	if err != nil {
		t.Fatal(err)
	}
	if len(pdfs) != 1 {
		t.Fatalf("expected a single PDF, got %v", pdfs)
	}

	content, err := os.ReadFile(pdfs[0])
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(content), "%PDF") {
		t.Fatalf("the generated file is not a PDF")
	}
}

func TestSendStopsOnMailerError(t *testing.T) {
	inTempDir(t)

	mailer := mockMailer{failFor: "p2"}
	err := Send(&mailer, testPairs, testTemplate, SendOpts{})
	if !errors.Is(err, errMockSend) {
		t.Fatalf("expected the mailer error, got %v", err)
	}

	// the first email went out, the failing one and everything after it didn't
	if len(mailer.mails) != 1 || mailer.mails[0].recipient != "p1" {
		t.Fatalf("expected only the email to p1 to be sent, got %v", mailer.mails)
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
