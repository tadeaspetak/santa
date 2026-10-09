package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tadeaspetak/santa/internal/app"
	"github.com/tadeaspetak/santa/internal/data"
)

var previewToFlagName = "to"

// sampleRecipients are used instead of the real recipients so that previews don't reveal anything.
func sampleRecipients(dat data.Data) []data.Person {
	recipients := []data.Person{{Salutation: "Rudolph"}}
	// show what multiple recipients look like when there are extras
	if len(dat.Extras) > 0 {
		recipients = append(recipients, data.Person{Salutation: "Blitzen"})
	}
	return recipients
}

func newPreviewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "preview",
		Short: "Show a sample email, or send it to yourself with --to",
		Long: `Render a single sample email from your template. It uses a made-up recipient ("Rudolph")
instead of a real one, so it doesn't reveal any pairings.

Without flags, the email is printed in the terminal. With --to, it's sent to that address only,
so you can see what it looks like in a real mailbox and make sure your mailer works.`,
		Example: `  santa preview
  santa preview --to you@example.com`,
		Args: cobra.NoArgs,
		RunE: runPreview,
	}

	cmd.Flags().StringP(previewToFlagName, "t", "", "send the sample email to this address")

	return cmd
}

func runPreview(cmd *cobra.Command, args []string) error {
	to, err := cmd.Flags().GetString(previewToFlagName)
	if err != nil {
		return err
	}
	if to != "" {
		if err := validateEmail(to); err != nil {
			return fmt.Errorf("'%s' is not a valid email address", to)
		}
	}

	dat, _, err := loadData(cmd)
	if err != nil {
		return err
	}
	if err := validate(dat, to != ""); err != nil {
		return fmt.Errorf("invalid data!\n\n%s\n\nRun `santa check` for more details", err)
	}

	out := cmd.OutOrStdout()
	subject, body := app.Render(*dat.Template, sampleRecipients(dat))

	if to == "" {
		line := strings.Repeat("─", 60)
		from := dat.Template.Sender
		if dat.Template.SenderName != "" {
			from = fmt.Sprintf("%s <%s>", dat.Template.SenderName, dat.Template.Sender)
		}
		fmt.Fprintf(out, "From:    %s\n", from)
		fmt.Fprintf(out, "To:      each participant\n")
		fmt.Fprintf(out, "Subject: %s\n", subject)
		fmt.Fprintf(out, "%s\n%s\n%s\n", line, strings.TrimSpace(app.HTMLToText(body)), line)
		fmt.Fprintln(out, "This is a sample, the real emails name each giver's actual recipient(s).")
		fmt.Fprintln(out, "To see it in a real mailbox, run `santa preview --to <your email>`.")
		return nil
	}

	mlr, description := newMailer(dat)
	if err := mlr.Send(dat.Template.From(), subject, body, to, to); err != nil {
		return fmt.Errorf("unable to send the sample email via %s: %w", description, err)
	}
	fmt.Fprintf(out, "✓ Sample email sent to %s via %s\n", to, description)
	fmt.Fprintln(out, "If it looks good, try `santa send --dry-run` next.")
	return nil
}
