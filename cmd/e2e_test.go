package cmd

// End-to-end tests running whole `santa` commands in a temporary directory,
// sending real emails to the fake SMTP & Mailgun servers from `testutil`.

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/tadeaspetak/santa/cmd/version"
	"github.com/tadeaspetak/santa/internal/data"
	"github.com/tadeaspetak/santa/internal/testutil"
)

const subjectPrefix = "🎄 Find a gift for "

// run executes the santa command with the given args, feeding it the given stdin.
func run(t *testing.T, stdin string, args ...string) (string, error) {
	t.Helper()

	root := NewRootCmd()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetIn(strings.NewReader(stdin))
	root.SetArgs(args)

	err := root.Execute()
	return out.String(), err
}

// setup writes the example data into `data.json` in a new temporary working directory, with
// SMTP pointing at a fresh fake server. `modify` can adjust the data before it's written.
func setup(t *testing.T, modify func(d *data.Data, smtp *testutil.SMTPServer)) (data.Data, *testutil.SMTPServer) {
	t.Helper()

	dat, err := data.LoadData("../data.example.json")
	if err != nil {
		t.Fatal(err)
	}

	server := testutil.NewSMTPServer(t, "user", "pass")
	dat.Smtp = &data.Smtp{Host: server.Host, Port: server.Port, User: "user", Pass: "pass"}
	if modify != nil {
		modify(&dat, server)
	}

	testutil.Chdir(t, t.TempDir())
	if err := data.SaveData("data.json", dat); err != nil {
		t.Fatal(err)
	}

	return dat, server
}

func outputDirs(t *testing.T) []string {
	t.Helper()
	dirs, err := filepath.Glob("santa-batch-*")
	if err != nil {
		t.Fatal(err)
	}
	return dirs
}

func assertContains(t *testing.T, s string, expected ...string) {
	t.Helper()
	for _, e := range expected {
		if !strings.Contains(s, e) {
			t.Errorf("expected the output to contain %q, got:\n%s", e, s)
		}
	}
}

func assertNotContains(t *testing.T, s string, unexpected ...string) {
	t.Helper()
	for _, u := range unexpected {
		if strings.Contains(s, u) {
			t.Errorf("expected the output not to contain %q, got:\n%s", u, s)
		}
	}
}

// assertValidPairingEmails checks that the emails, as received by the mail server, make up a valid
// pairing for the data: every giver got exactly one email, every participant and extra is someone's
// recipient exactly once, and all the preferences are honoured. `giverOf` tells who the email is
// for, i.e. `To`, or `Reply-To` when all the emails were sent to a single address.
func assertValidPairingEmails(
	t *testing.T,
	dat data.Data,
	messages []testutil.Message,
	giverOf func(testutil.Message) string,
) {
	t.Helper()

	if len(messages) != len(dat.Participants) {
		t.Fatalf("expected %d emails, got %d: %+v", len(dat.Participants), len(messages), messages)
	}

	participantsBySalutation := map[string]data.Participant{}
	for _, p := range dat.Participants {
		participantsBySalutation[p.Salutation] = p
	}
	extrasBySalutation := map[string]data.Extra{}
	for _, e := range dat.Extras {
		extrasBySalutation[e.Salutation] = e
	}

	gifts := map[string]int{}
	givers := []string{}
	for _, m := range messages {
		giver := giverOf(m)
		givers = append(givers, giver)

		if m.From != dat.Template.Sender {
			t.Errorf("unexpected sender %s", m.From)
		}
		if !strings.HasPrefix(m.Subject, subjectPrefix) {
			t.Fatalf("unexpected subject %q", m.Subject)
		}

		recipients := strings.Split(strings.TrimPrefix(m.Subject, subjectPrefix), dat.Template.RecipientsSeparator)
		for _, r := range recipients {
			if !strings.Contains(m.HTML, r) {
				t.Errorf("expected the body for %s to mention %s: %s", giver, r, m.HTML)
			}
		}

		// the first recipient is always a participant, the rest are extras
		recipient, ok := participantsBySalutation[recipients[0]]
		if !ok {
			t.Fatalf("%s's first recipient %q is not a participant", giver, recipients[0])
		}
		giverParticipant := dat.Participants[slices.IndexFunc(dat.Participants, func(p data.Participant) bool {
			return p.Email == giver
		})]
		if recipient.Email == giver ||
			slices.Contains(giverParticipant.ExcludedRecipients, recipient.Email) ||
			(giverParticipant.PredestinedRecipient != "" && giverParticipant.PredestinedRecipient != recipient.Email) {
			t.Errorf("%s got an invalid recipient %s", giver, recipient.Email)
		}
		gifts[recipient.Salutation]++

		for _, r := range recipients[1:] {
			e, ok := extrasBySalutation[r]
			if !ok {
				t.Fatalf("%s's recipient %q is neither a participant nor an extra", giver, r)
			}
			if slices.Contains(e.ExcludedGivers, giver) {
				t.Errorf("extra %s assigned to excluded giver %s", r, giver)
			}
			gifts[r]++
		}
	}

	expectedGivers := []string{}
	for _, p := range dat.Participants {
		expectedGivers = append(expectedGivers, p.Email)
		if gifts[p.Salutation] != 1 {
			t.Errorf("%s receives %d gifts", p.Salutation, gifts[p.Salutation])
		}
	}
	for _, e := range dat.Extras {
		if gifts[e.Salutation] != 1 {
			t.Errorf("%s receives %d gifts", e.Salutation, gifts[e.Salutation])
		}
	}
	slices.Sort(givers)
	slices.Sort(expectedGivers)
	if !slices.Equal(givers, expectedGivers) {
		t.Errorf("expected emails for %v, got %v", expectedGivers, givers)
	}
}

// assertBatchFilesMatch checks that the output folder has a copy of each sent email.
func assertBatchFilesMatch(t *testing.T, dir string, messages []testutil.Message, giverOf func(testutil.Message) string) {
	t.Helper()
	for _, m := range messages {
		content, err := os.ReadFile(filepath.Join(dir, giverOf(m)+".txt"))
		if err != nil {
			t.Fatalf("expected a batch file for %s: %v", giverOf(m), err)
		}
		if !strings.Contains(string(content), "Subject: "+m.Subject+"\n") {
			t.Errorf("the batch file for %s doesn't match the sent email %q:\n%s", giverOf(m), m.Subject, content)
		}
	}
}

func to(m testutil.Message) string      { return m.To }
func replyTo(m testutil.Message) string { return m.ReplyTo }

func TestE2EGettingStarted(t *testing.T) {
	testutil.Chdir(t, t.TempDir())

	_, err := run(t, "", "send")
	if err == nil || !strings.Contains(err.Error(), "santa init") {
		t.Fatalf("expected a hint to run `santa init`, got %v", err)
	}

	out, err := run(t, "", "init")
	if err != nil {
		t.Fatalf("init failed: %v", err)
	}
	assertContains(t, out, "data.json", "santa check")

	if _, err := run(t, "", "init"); err == nil || !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("expected init to refuse overwriting the data file, got %v", err)
	}

	// everything below works offline, the sample SMTP config is never used
	out, err = run(t, "", "preview")
	if err != nil {
		t.Fatalf("preview failed: %v", err)
	}
	assertContains(t, out, "Subject: "+subjectPrefix+"Rudolph", "santa preview --to")

	out, err = run(t, "", "send", "--dry-run", "--pdf")
	if err != nil {
		t.Fatalf("dry run failed: %v", err)
	}
	assertContains(t, out, "Dry run, nothing will be sent.", "mom@family.com (not sent, dry run)")

	dirs := outputDirs(t)
	if len(dirs) != 1 {
		t.Fatalf("expected a single output folder, got %v", dirs)
	}
	for _, f := range []string{"mom@family.com.txt", "jake@family.com.txt", "emails.pdf"} {
		if _, err := os.Stat(filepath.Join(dirs[0], f)); err != nil {
			t.Errorf("expected %s in the output folder: %v", f, err)
		}
	}
}

func TestE2ECheck(t *testing.T) {
	tests := []struct {
		name        string
		modify      func(d *data.Data, smtp *testutil.SMTPServer)
		expectedOut []string
		absentOut   []string
		ok          bool
	}{
		{
			name: "all good",
			expectedOut: []string{
				"✓ The data file 'data.json' is valid (5 participants, 2 extra recipients)",
				"✓ The participants can be paired",
				"✓ Logged in to SMTP (127.0.0.1:",
				"All good!",
			},
			ok: true,
		},
		{
			name:        "wrong password",
			modify:      func(d *data.Data, _ *testutil.SMTPServer) { d.Smtp.Pass = "wrong" },
			expectedOut: []string{"✓ The participants can be paired", "✗ Unable to log in to SMTP", "Username and Password not accepted"},
			absentOut:   []string{"All good!"},
		},
		{
			name:        "no mailer",
			modify:      func(d *data.Data, _ *testutil.SMTPServer) { d.Smtp = nil },
			expectedOut: []string{"✓ The participants can be paired", "✗ No mailer configured"},
		},
		{
			name: "impossible preferences",
			modify: func(d *data.Data, _ *testutil.SMTPServer) {
				d.Participants[0].ExcludedRecipients = []string{"dda@family.com"}
			},
			expectedOut: []string{
				"✗ Unable to pair the participants",
				"excluded recipient does not exist: dda@family.com (excluded by mom@family.com)",
				// the remaining checks still run
				"✓ Logged in to SMTP",
			},
		},
		{
			name:        "invalid data",
			modify:      func(d *data.Data, _ *testutil.SMTPServer) { d.Template.Sender = "santa" },
			expectedOut: []string{"✗ The data file 'data.json' is invalid", "Data.Template.Sender"},
			absentOut:   []string{"paired", "Logged in"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, server := setup(t, tt.modify)

			out, err := run(t, "", "check")
			if tt.ok && err != nil {
				t.Fatalf("unexpected error: %v\n%s", err, out)
			}
			if !tt.ok && !errors.Is(err, errChecksFailed) {
				t.Fatalf("expected the checks to fail, got %v\n%s", err, out)
			}
			assertContains(t, out, tt.expectedOut...)
			assertNotContains(t, out, tt.absentOut...)

			if len(server.Messages()) != 0 || len(outputDirs(t)) != 0 {
				t.Fatalf("check must not send or write anything")
			}
		})
	}
}

func TestE2EPreview(t *testing.T) {
	t.Run("print", func(t *testing.T) {
		// printing doesn't need a mailer
		_, server := setup(t, func(d *data.Data, _ *testutil.SMTPServer) { d.Smtp = nil })

		out, err := run(t, "", "preview")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// the example has extras, so the sample shows two recipients
		assertContains(t, out, "Subject: "+subjectPrefix+"Rudolph and Blitzen", "Come up with something lovely for Rudolph and Blitzen.")
		assertNotContains(t, out, "Mom", "Dad", "Auntie", "Emily", "Jake")

		if len(server.Messages()) != 0 {
			t.Fatalf("printing a preview must not send anything")
		}
	})

	t.Run("send to", func(t *testing.T) {
		dat, server := setup(t, nil)

		out, err := run(t, "", "preview", "--to", "me@example.com")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assertContains(t, out, "✓ Sample email sent to me@example.com")

		messages := server.Messages()
		if len(messages) != 1 {
			t.Fatalf("expected a single email, got %+v", messages)
		}
		m := messages[0]
		if m.To != "me@example.com" || m.From != dat.Template.Sender || m.Subject != subjectPrefix+"Rudolph and Blitzen" {
			t.Fatalf("unexpected email %+v", m)
		}
	})

	t.Run("invalid address", func(t *testing.T) {
		setup(t, nil)
		if _, err := run(t, "", "preview", "-t", "me"); err == nil || !strings.Contains(err.Error(), "not a valid email") {
			t.Fatalf("expected an invalid email error, got %v", err)
		}
	})
}

func TestE2ESend(t *testing.T) {
	t.Run("cancelled by default", func(t *testing.T) {
		_, server := setup(t, nil)

		out, err := run(t, "\n", "send")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		assertContains(t, out, "Send 5 emails to the participants? This can't be undone. [y/N]", "Nothing sent.")

		if len(server.Messages()) != 0 || len(outputDirs(t)) != 0 {
			t.Fatalf("cancelling must not send or write anything")
		}
	})

	t.Run("confirmed", func(t *testing.T) {
		dat, server := setup(t, nil)

		out, err := run(t, "y\n", "send")
		if err != nil {
			t.Fatalf("unexpected error: %v\n%s", err, out)
		}
		assertContains(t, out, "Sending 5 emails:", "  ✓ mom@family.com", "Done, all emails sent")

		messages := server.Messages()
		assertValidPairingEmails(t, dat, messages, to)
		for _, m := range messages {
			if m.ReplyTo != m.To {
				t.Errorf("expected the reply-to to be the giver, got %+v", m)
			}
		}
		assertBatchFilesMatch(t, outputDirs(t)[0], messages, to)
	})

	t.Run("yes", func(t *testing.T) {
		dat, server := setup(t, nil)

		out, err := run(t, "", "send", "-y")
		if err != nil {
			t.Fatalf("unexpected error: %v\n%s", err, out)
		}
		assertNotContains(t, out, "[y/N]")
		assertValidPairingEmails(t, dat, server.Messages(), to)
	})

	t.Run("to a single address", func(t *testing.T) {
		dat, server := setup(t, nil)

		out, err := run(t, "y\n", "send", "--to", "me@example.com")
		if err != nil {
			t.Fatalf("unexpected error: %v\n%s", err, out)
		}
		assertContains(t, out,
			"Send 5 emails to me@example.com (instead of the participants)? [y/N]",
			"  ✓ me@example.com (instead of mom@family.com)",
		)

		messages := server.Messages()
		for _, m := range messages {
			if m.To != "me@example.com" {
				t.Errorf("expected all emails to go to me@example.com, got %+v", m)
			}
		}
		// the reply-to tells who each email was meant for
		assertValidPairingEmails(t, dat, messages, replyTo)
		assertBatchFilesMatch(t, outputDirs(t)[0], messages, replyTo)
	})

	t.Run("dry run without a mailer", func(t *testing.T) {
		dat, server := setup(t, func(d *data.Data, _ *testutil.SMTPServer) { d.Smtp = nil })

		out, err := run(t, "", "send", "-d", "-p")
		if err != nil {
			t.Fatalf("unexpected error: %v\n%s", err, out)
		}
		assertNotContains(t, out, "[y/N]")

		if len(server.Messages()) != 0 {
			t.Fatalf("a dry run must not send anything")
		}
		entries, err := os.ReadDir(outputDirs(t)[0])
		if err != nil {
			t.Fatal(err)
		}
		if len(entries) != len(dat.Participants)+1 {
			t.Fatalf("expected a file per participant plus the PDF, got %v", entries)
		}
	})

	t.Run("wrong password fails before the prompt", func(t *testing.T) {
		_, server := setup(t, func(d *data.Data, _ *testutil.SMTPServer) { d.Smtp.Pass = "wrong" })

		out, err := run(t, "y\n", "send")
		if err == nil || !strings.Contains(err.Error(), "unable to log in to SMTP") {
			t.Fatalf("expected a login error, got %v", err)
		}
		assertNotContains(t, out, "[y/N]")
		if len(server.Messages()) != 0 || len(outputDirs(t)) != 0 {
			t.Fatalf("nothing must be sent or written")
		}
	})

	t.Run("impossible preferences fail before the prompt", func(t *testing.T) {
		_, server := setup(t, func(d *data.Data, _ *testutil.SMTPServer) {
			d.Participants[3].ExcludedRecipients = []string{d.Participants[3].PredestinedRecipient}
		})

		out, err := run(t, "y\n", "send")
		if err == nil || !strings.Contains(err.Error(), "predestined recipient is also excluded") {
			t.Fatalf("expected a pairing error, got %v", err)
		}
		assertNotContains(t, out, "[y/N]")
		if len(server.Messages()) != 0 || len(outputDirs(t)) != 0 {
			t.Fatalf("nothing must be sent or written")
		}
	})

	t.Run("failure halfway through", func(t *testing.T) {
		_, server := setup(t, func(_ *data.Data, smtp *testutil.SMTPServer) { smtp.FailFor("auntie@family.com") })

		_, err := run(t, "", "send", "-y")
		if err == nil {
			t.Fatalf("expected an error")
		}
		// auntie is the third participant
		dirs := outputDirs(t)
		assertContains(t, err.Error(), "auntie@family.com", "2 of 5 sent", "The emails are in "+dirs[0])
		if len(server.Messages()) != 2 {
			t.Fatalf("expected the first two emails to be sent, got %+v", server.Messages())
		}
	})
}

func TestE2EMailgun(t *testing.T) {
	setupMailgun := func(t *testing.T, apiKey string) (data.Data, *testutil.MailgunServer) {
		server := testutil.NewMailgunServer(t, "mg.example.com", "key")
		previous := mailgunAPIBase
		mailgunAPIBase = server.APIBase
		t.Cleanup(func() { mailgunAPIBase = previous })

		dat, _ := setup(t, func(d *data.Data, _ *testutil.SMTPServer) {
			d.Smtp = nil
			d.Mailgun = &data.Mailgun{Domain: "mg.example.com", APIKey: apiKey}
		})
		return dat, server
	}

	t.Run("check and send", func(t *testing.T) {
		dat, server := setupMailgun(t, "key")

		out, err := run(t, "", "check")
		if err != nil {
			t.Fatalf("unexpected error: %v\n%s", err, out)
		}
		assertContains(t, out, "✓ Logged in to Mailgun (mg.example.com)")

		if out, err := run(t, "", "send", "-y"); err != nil {
			t.Fatalf("unexpected error: %v\n%s", err, out)
		}
		assertValidPairingEmails(t, dat, server.Messages(), to)
	})

	t.Run("wrong API key", func(t *testing.T) {
		_, server := setupMailgun(t, "wrong")

		out, err := run(t, "", "check")
		if !errors.Is(err, errChecksFailed) {
			t.Fatalf("expected the checks to fail, got %v", err)
		}
		assertContains(t, out, "✗ Unable to log in to Mailgun (mg.example.com)")

		if _, err := run(t, "y\n", "send"); err == nil {
			t.Fatalf("expected sending to fail")
		}
		if len(server.Messages()) != 0 {
			t.Fatalf("nothing must be sent")
		}
	})
}

func TestE2EVersion(t *testing.T) {
	previous := version.Version
	version.Version = "v1.2.3"
	t.Cleanup(func() { version.Version = previous })

	out, err := run(t, "", "version")
	if err != nil || out != "v1.2.3\n" {
		t.Fatalf("unexpected version output %q (error: %v)", out, err)
	}
}
