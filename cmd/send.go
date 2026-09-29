package cmd

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/tadeaspetak/santa/internal/app"
)

var dryRunFlagName = "dry-run"
var toFlagName = "to"
var pdfFlagName = "pdf"
var yesFlagName = "yes"

// confirm asks a yes/no question, defaulting to no.
func confirm(question string) bool {
	prompt := promptui.Prompt{
		Label:     question,
		IsConfirm: true,
	}
	_, err := prompt.Run()
	if err != nil {
		if errors.Is(err, promptui.ErrAbort) {
			return false
		}

		log.Fatalf("Prompt failed %v\n", err)
	}

	return true
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

var sendCmd = &cobra.Command{
	Use:   "send",
	Short: "Draw the pairings and email every participant",
	Long: `Draw the pairings and email every participant who they're giving a gift to.

Every run creates a santa-batch-<datetime> folder with a copy of each email, named after
the participant it's for (plus emails.pdf with --pdf). Opening those reveals the pairings!`,
	Example: `  santa send --dry-run          # nothing is sent, check the output folder
  santa send --dry-run --pdf    # ...and generate a printable PDF
  santa send --to you@x.com     # send all the emails to yourself
  santa send                    # send the emails to the participants`,
	Run: func(cmd *cobra.Command, args []string) {
		isDryRun := getFlag(cmd, dryRunFlagName, cmd.Flags().GetBool)
		to := getFlag(cmd, toFlagName, cmd.Flags().GetString)
		shouldPrintPdf := getFlag(cmd, pdfFlagName, cmd.Flags().GetBool)
		yes := getFlag(cmd, yesFlagName, cmd.Flags().GetBool)

		if to != "" {
			if err := validateEmail(to); err != nil {
				log.Fatalf("'%s' is not a valid email address.", to)
			}
		}

		dat := loadData(cmd)
		if err := validate(dat, !isDryRun); err != nil {
			log.Fatalf("Invalid data!\n\n%s\n\nRun `santa check` for more details.", err)
		}

		// pair before asking anything so that any problems surface straight away
		pairs, err := app.Pair(dat.Participants, dat.Extras)
		if err != nil {
			log.Fatalf("Unable to pair the participants:\n%v", err)
		}

		var mlr app.Mailer
		if !isDryRun {
			var description string
			mlr, description = newMailer(dat)
			if err := mlr.Verify(context.Background()); err != nil {
				log.Fatalf("Unable to log in to %s: %v\n\nRun `santa check` once you've fixed it.", description, err)
			}
		}

		switch {
		case isDryRun:
			fmt.Printf("Dry run, nothing will be sent.\n")
		case to != "":
			if !yes && !confirm(fmt.Sprintf("Send %d emails to %s (instead of the participants)", len(pairs), to)) {
				fmt.Println("Nothing sent.")
				return
			}
			fmt.Printf("Sending %d emails to %s:\n", len(pairs), to)
		default:
			if !yes && !confirm(fmt.Sprintf("Send %d emails to the participants? This can't be undone", len(pairs))) {
				fmt.Println("Nothing sent.")
				return
			}
			fmt.Printf("Sending %d emails:\n", len(pairs))
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
			},
		)
		if err != nil {
			log.Fatalf("\nThere has been an error sending your emails: %v\n\nThe emails are in %s/.", err, outputDir)
		}

		switch {
		case isDryRun:
			fmt.Printf("\nDone. The emails are in %s/ (opening them reveals the pairings).\n", outputDir)
		case to != "":
			fmt.Printf("\nDone! Each email's Reply-To is the participant it was meant for. Copies are in %s/.\n", outputDir)
		default:
			fmt.Printf("\nDone, all emails sent 🎅 Copies are in %s/, don't peek!\n", outputDir)
		}
	},
}

func init() {
	sendCmd.Flags().BoolP(dryRunFlagName, "d", false, "do everything except sending the emails")
	sendCmd.Flags().StringP(toFlagName, "t", "", "send all the emails to this address instead of the participants")
	sendCmd.Flags().BoolP(pdfFlagName, "p", false, "also generate a printable PDF with all the emails")
	sendCmd.Flags().BoolP(yesFlagName, "y", false, "don't ask for confirmation")
}
