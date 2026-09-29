package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tadeaspetak/santa/internal/app"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check the data file, the pairing and the mailer login without sending anything",
	Long: `Check that everything is ready for sending, without sending anything or revealing any pairings:

  1. the data file is valid,
  2. the preferences (predestined & excluded recipients, excluded givers) allow for a valid pairing,
  3. the mailer (SMTP or Mailgun) accepts your credentials.`,
	Run: func(cmd *cobra.Command, args []string) {
		dataPath := getDataPath(cmd)
		dat := loadData(cmd)
		ok := true

		if err := validate(dat, false); err != nil {
			fmt.Printf("✗ The data file '%s' is invalid:\n\n%s\n", dataPath, err)
			os.Exit(1)
		}
		fmt.Printf(
			"✓ The data file '%s' is valid (%d participants, %d extra recipients)\n",
			dataPath, len(dat.Participants), len(dat.Extras),
		)

		if _, err := app.Pair(dat.Participants, dat.Extras); err != nil {
			fmt.Printf("✗ Unable to pair the participants:\n%v\n", err)
			ok = false
		} else {
			fmt.Println("✓ The participants can be paired (the pairing is not shown)")
		}

		if err := validate(dat, true); err != nil {
			fmt.Println("✗ No mailer configured, add `smtp` or `mailgun` to your data file (dry runs work without one)")
			ok = false
		} else {
			mlr, description := newMailer(dat)
			if err := mlr.Verify(context.Background()); err != nil {
				fmt.Printf("✗ Unable to log in to %s:\n%v\n", description, err)
				ok = false
			} else {
				fmt.Printf("✓ Logged in to %s\n", description)
			}
		}

		if !ok {
			os.Exit(1)
		}

		fmt.Println("\nAll good! Next, try `santa preview --to <your email>` to see what the emails look like.")
	},
}
