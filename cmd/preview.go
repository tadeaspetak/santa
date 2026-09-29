package cmd

import (
	"fmt"
	"log"
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

var previewCmd = &cobra.Command{
	Use:   "preview",
	Short: "Show a sample email, or send it to yourself with --to",
	Long: `Render a single sample email from your template. It uses a made-up recipient ("Rudolph")
instead of a real one, so it doesn't reveal any pairings.

Without flags, the email is printed in the terminal. With --to, it's sent to that address only,
so you can see what it looks like in a real mailbox and make sure your mailer works.`,
	Example: `  santa preview
  santa preview --to you@example.com`,
	Run: func(cmd *cobra.Command, args []string) {
		to := getFlag(cmd, previewToFlagName, cmd.Flags().GetString)
		if to != "" {
			if err := validateEmail(to); err != nil {
				log.Fatalf("'%s' is not a valid email address.", to)
			}
		}

		dat := loadData(cmd)
		if err := validate(dat, to != ""); err != nil {
			log.Fatalf("Invalid data!\n\n%s\n\nRun `santa check` for more details.", err)
		}

		subject, body := app.Render(*dat.Template, sampleRecipients(dat))

		if to == "" {
			fmt.Printf("From:    %s\n", dat.Template.Sender)
			fmt.Printf("To:      each participant\n")
			fmt.Printf("Subject: %s\n", subject)
			fmt.Printf("%s\n%s\n%s\n", strings.Repeat("─", 60), strings.TrimSpace(app.HTMLToText(body)), strings.Repeat("─", 60))
			fmt.Println("This is a sample, the real emails name each giver's actual recipient(s).")
			fmt.Println("To see it in a real mailbox, run `santa preview --to <your email>`.")
			return
		}

		mlr, description := newMailer(dat)
		if err := mlr.Send(dat.Template.Sender, subject, body, to, to); err != nil {
			log.Fatalf("Unable to send the sample email via %s: %v", description, err)
		}
		fmt.Printf("✓ Sample email sent to %s via %s\n", to, description)
		fmt.Println("If it looks good, try `santa send --dry-run` next.")
	},
}

func init() {
	previewCmd.Flags().StringP(previewToFlagName, "t", "", "send the sample email to this address")
}
