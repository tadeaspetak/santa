package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tadeaspetak/santa/internal/app"
)

var dryRunFlagName = "dry-run"
var toFlagName = "to"
var pdfFlagName = "pdf"
var yesFlagName = "yes"

// confirm asks a yes/no question, defaulting to no.
func confirm(in io.Reader, out io.Writer, question string) bool {
	fmt.Fprintf(out, "%s [y/N] ", question)
	answer, _ := bufio.NewReader(in).ReadString('\n')
	switch strings.ToLower(strings.TrimSpace(answer)) {
	case "y", "yes":
		return true
	default:
		return false
	}
}

// newOutputDir returns a name for a new output directory, making sure it doesn't exist yet
// so that runs within the same second don't mix their files.
func newOutputDir(now time.Time) string {
	base := fmt.Sprintf("santa-batch-%s", now.Local().Format("20060102-150405"))
	dir := base
	for i := 2; ; i++ {
		if _, err := os.Stat(dir); errors.Is(err, os.ErrNotExist) {
			return dir
		}
		dir = fmt.Sprintf("%s-%d", base, i)
	}
}

func newSendCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "send",
		Short: "Draw the pairings and email every participant",
		Long: `Draw the pairings and email every participant who they're giving a gift to.

Every run creates a santa-batch-<datetime> folder with a copy of each email, named after
the participant it's for (plus emails.pdf with --pdf). Opening those reveals the pairings!`,
		Example: `  santa send --dry-run          # nothing is sent, check the output folder
  santa send --dry-run --pdf    # ...and generate a printable PDF
  santa send --to you@x.com     # send all the emails to yourself
  santa send                    # send the emails to the participants`,
		Args: cobra.NoArgs,
		RunE: runSend,
	}

	cmd.Flags().BoolP(dryRunFlagName, "d", false, "do everything except sending the emails")
	cmd.Flags().StringP(toFlagName, "t", "", "send all the emails to this address instead of the participants")
	cmd.Flags().BoolP(pdfFlagName, "p", false, "also generate a printable PDF with all the emails")
	cmd.Flags().BoolP(yesFlagName, "y", false, "don't ask for confirmation")

	return cmd
}

func runSend(cmd *cobra.Command, args []string) error {
	flags := cmd.Flags()
	isDryRun, _ := flags.GetBool(dryRunFlagName)
	to, _ := flags.GetString(toFlagName)
	shouldPrintPdf, _ := flags.GetBool(pdfFlagName)
	yes, _ := flags.GetBool(yesFlagName)

	if to != "" {
		if err := validateEmail(to); err != nil {
			return fmt.Errorf("'%s' is not a valid email address", to)
		}
	}

	dat, _, err := loadData(cmd)
	if err != nil {
		return err
	}
	if err := validate(dat, !isDryRun); err != nil {
		return fmt.Errorf("invalid data!\n\n%s\n\nRun `santa check` for more details", err)
	}

	// pair before asking anything so that any problems surface straight away
	pairs, err := app.Pair(dat.Participants, dat.Extras)
	if err != nil {
		return fmt.Errorf("unable to pair the participants:\n%w", err)
	}

	var mlr app.Mailer
	if !isDryRun {
		var description string
		mlr, description = newMailer(dat)
		if err := mlr.Verify(cmd.Context()); err != nil {
			return fmt.Errorf("unable to log in to %s: %w\n\nRun `santa check` once you've fixed it", description, err)
		}
	}

	in, out := cmd.InOrStdin(), cmd.OutOrStdout()
	switch {
	case isDryRun:
		fmt.Fprintln(out, "Dry run, nothing will be sent.")
	case to != "":
		if !yes && !confirm(in, out, fmt.Sprintf("Send %d emails to %s (instead of the participants)?", len(pairs), to)) {
			fmt.Fprintln(out, "Nothing sent.")
			return nil
		}
		fmt.Fprintf(out, "Sending %d emails to %s:\n", len(pairs), to)
	default:
		if !yes && !confirm(in, out, fmt.Sprintf("Send %d emails to the participants? This can't be undone.", len(pairs))) {
			fmt.Fprintln(out, "Nothing sent.")
			return nil
		}
		fmt.Fprintf(out, "Sending %d emails:\n", len(pairs))
	}

	outputDir := newOutputDir(time.Now())
	err = app.Send(
		mlr,
		pairs,
		*dat.Template,
		app.SendOpts{
			To:             to,
			IsDryRun:       isDryRun,
			ShouldPrintPdf: shouldPrintPdf,
			OutputDir:      outputDir,
			Out:            out,
		},
	)
	if err != nil {
		return fmt.Errorf("there has been an error sending your emails: %w\n\nThe emails are in %s/", err, outputDir)
	}

	switch {
	case isDryRun:
		fmt.Fprintf(out, "\nDone. The emails are in %s/ (opening them reveals the pairings).\n", outputDir)
	case to != "":
		fmt.Fprintf(out, "\nDone! Each email's Reply-To is the participant it was meant for. Copies are in %s/.\n", outputDir)
	default:
		fmt.Fprintf(out, "\nDone, all emails sent 🎅 Copies are in %s/, don't peek!\n", outputDir)
	}
	return nil
}
