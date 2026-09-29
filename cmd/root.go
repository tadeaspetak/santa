package cmd

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"
	"github.com/tadeaspetak/santa/internal/app"
	"github.com/tadeaspetak/santa/internal/app/mailer"
	"github.com/tadeaspetak/santa/internal/data"
)

var dataPathFlagName = "data"

func getDataPath(cmd *cobra.Command) string {
	dataPath, err := cmd.Flags().GetString(dataPathFlagName)

	if err != nil {
		log.Fatalf("Could not get data-file path %v\n", err)
	}

	return dataPath
}

// loadData loads the data file, exiting with a helpful message if that's not possible.
func loadData(cmd *cobra.Command) data.Data {
	dataPath := getDataPath(cmd)
	dat, err := data.LoadData(dataPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			log.Fatalf("Data file '%s' does not exist. Run `santa init` to create one.", dataPath)
		}
		log.Fatalf("Unable to load the data file '%s': %v", dataPath, err)
	}
	return dat
}

// newMailer creates the mailer configured in the data. Mailgun wins if both are configured.
// The data must have been validated with a mailer required.
func newMailer(dat data.Data) (mlr app.Mailer, description string) {
	// note: checking the domain is enough, the rest has been validated
	if dat.Mailgun != nil && dat.Mailgun.Domain != "" {
		return mailer.NewMailgunMailer(dat.Mailgun.Domain, dat.Mailgun.APIKey),
			fmt.Sprintf("Mailgun (%s)", dat.Mailgun.Domain)
	}
	port := dat.Smtp.Port
	if port == 0 {
		port = 587
	}
	return mailer.NewSmtpMailer(dat.Smtp.Host, port, dat.Smtp.User, dat.Smtp.Pass),
		fmt.Sprintf("SMTP (%s:%d as %s)", dat.Smtp.Host, port, dat.Smtp.User)
}

func getFlag[T any](cmd *cobra.Command, name string, get func(string) (T, error)) T {
	value, err := get(name)
	if err != nil {
		log.Fatalf("Unable to get the %s flag: %v", name, err)
	}
	return value
}

var RootCmd = &cobra.Command{
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
}

func Execute() {
	// no timestamps in error messages, this is a CLI
	log.SetFlags(0)

	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// disable the default completion command
	RootCmd.CompletionOptions.DisableDefaultCmd = true

	RootCmd.PersistentFlags().String(dataPathFlagName, "data.json", "data file path")

	RootCmd.AddCommand(initCmd)
	RootCmd.AddCommand(checkCmd)
	RootCmd.AddCommand(previewCmd)
	RootCmd.AddCommand(sendCmd)
	RootCmd.AddCommand(versionCmd)
}
