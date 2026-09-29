package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/tadeaspetak/santa/internal/app"
)

var errChecksFailed = errors.New("some checks failed, see above")

func newCheckCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "check",
		Short: "Check the data file, the pairing and the mailer login without sending anything",
		Long: `Check that everything is ready for sending, without sending anything or revealing any pairings:

  1. the data file is valid,
  2. the preferences (predestined & excluded recipients, excluded givers) allow for a valid pairing,
  3. the mailer (SMTP or Mailgun) accepts your credentials.`,
		Args: cobra.NoArgs,
		RunE: runCheck,
	}
}

func runCheck(cmd *cobra.Command, args []string) error {
	dat, dataPath, err := loadData(cmd)
	if err != nil {
		return err
	}

	out := cmd.OutOrStdout()
	ok := true

	if err := validate(dat, false); err != nil {
		fmt.Fprintf(out, "✗ The data file '%s' is invalid:\n\n%s\n\n", dataPath, err)
		return errChecksFailed
	}
	fmt.Fprintf(
		out, "✓ The data file '%s' is valid (%d participants, %d extra recipients)\n",
		dataPath, len(dat.Participants), len(dat.Extras),
	)

	if _, err := app.Pair(dat.Participants, dat.Extras); err != nil {
		fmt.Fprintf(out, "✗ Unable to pair the participants:\n%v\n", err)
		ok = false
	} else {
		fmt.Fprintln(out, "✓ The participants can be paired (the pairing is not shown)")
	}

	if err := validate(dat, true); err != nil {
		fmt.Fprintln(out, "✗ No mailer configured, add `smtp` or `mailgun` to your data file (dry runs work without one)")
		ok = false
	} else {
		mlr, description := newMailer(dat)
		if err := mlr.Verify(cmd.Context()); err != nil {
			fmt.Fprintf(out, "✗ Unable to log in to %s:\n%v\n", description, err)
			ok = false
		} else {
			fmt.Fprintf(out, "✓ Logged in to %s\n", description)
		}
	}

	if !ok {
		fmt.Fprintln(out)
		return errChecksFailed
	}

	fmt.Fprintln(out, "\nAll good! Next, try `santa preview --to <your email>` to see what the emails look like.")
	return nil
}
