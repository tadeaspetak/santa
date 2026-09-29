package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tadeaspetak/santa/internal/app"
	"github.com/tadeaspetak/santa/internal/app/mailer"
	"github.com/tadeaspetak/santa/internal/data"
)

var dataPathFlagName = "data"

// mailgunAPIBase is a variable so that tests can point it to a fake server.
var mailgunAPIBase = mailer.MailgunEUAPIBase

func getDataPath(cmd *cobra.Command) (string, error) {
	return cmd.Flags().GetString(dataPathFlagName)
}

// loadData loads the data file, returning a helpful error if that's not possible.
func loadData(cmd *cobra.Command) (data.Data, string, error) {
	dataPath, err := getDataPath(cmd)
	if err != nil {
		return data.Data{}, "", err
	}

	dat, err := data.LoadData(dataPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return data.Data{}, dataPath, fmt.Errorf("data file '%s' does not exist, run `santa init` to create one", dataPath)
		}
		return data.Data{}, dataPath, fmt.Errorf("unable to load the data file '%s': %w", dataPath, err)
	}
	return dat, dataPath, nil
}

// newMailer creates the mailer configured in the data. Mailgun wins if both are configured.
// The data must have been validated with a mailer required.
func newMailer(dat data.Data) (mlr app.Mailer, description string) {
	// note: checking the domain is enough, the rest has been validated
	if dat.Mailgun != nil && dat.Mailgun.Domain != "" {
		return mailer.NewMailgunMailer(dat.Mailgun.Domain, dat.Mailgun.APIKey, mailgunAPIBase),
			fmt.Sprintf("Mailgun (%s)", dat.Mailgun.Domain)
	}

	port := dat.Smtp.Port
	if port == 0 {
		port = 587
	}
	return mailer.NewSmtpMailer(dat.Smtp.Host, port, dat.Smtp.User, dat.Smtp.Pass),
		fmt.Sprintf("SMTP (%s:%d as %s)", dat.Smtp.Host, port, dat.Smtp.User)
}

// NewRootCmd creates the whole command tree. A new tree is created every time so that
// flag values don't leak between runs, e.g. in tests.
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "santa",
		Short: "Draw your Secret Santa pairings and email them to everyone",
		Long: `Santa draws your Secret Santa pairings and emails every participant who they're giving a gift to.

A typical run:
  santa init                   create a sample data.json, then edit it
  santa check                  validate the data and the mailer login (sends nothing)
  santa preview --to you@x.com send yourself a single sample email
  santa send --dry-run         do everything except sending, see the output folder
  santa send --to you@x.com    send all emails to yourself instead of the participants
  santa send                   the real thing 🎅

For more info, head over to the README at https://github.com/tadeaspetak/santa/blob/main/README.md`,
		// errors are printed by `Execute`, and usage only helps with invalid flags or arguments
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	// disable the default completion command
	root.CompletionOptions.DisableDefaultCmd = true

	root.PersistentFlags().String(dataPathFlagName, "data.json", "data file path")

	root.AddCommand(newInitCmd())
	root.AddCommand(newCheckCmd())
	root.AddCommand(newPreviewCmd())
	root.AddCommand(newSendCmd())
	root.AddCommand(newVersionCmd())

	return root
}

func Execute() {
	if err := NewRootCmd().Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
