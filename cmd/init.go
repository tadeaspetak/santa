package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/tadeaspetak/santa/internal/data"
)

func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Create a sample data file to start from",
		Args:  cobra.NoArgs,
		RunE:  runInit,
	}
}

func runInit(cmd *cobra.Command, args []string) error {
	dataPath, err := getDataPath(cmd)
	if err != nil {
		return err
	}
	if _, err := os.Stat(dataPath); err == nil {
		return fmt.Errorf(
			"file '%s' already exists, if you'd like to start fresh, rename or remove this file first",
			dataPath,
		)
	}

	dat := data.Data{
		Schema: "https://raw.githubusercontent.com/tadeaspetak/santa/refs/heads/main/schema.json",
		Smtp: &data.Smtp{
			Host: "smtp.gmail.com",
			Pass: "xxx yyyy zzz",
			User: "your.email@gmail.com",
		},
		Template: &data.Template{
			Subject:             "🎄 Find a gift for %{recipientSalutation}",
			Body:                "<p>Hi</p><p>Come up with something lovely for %{recipientSalutation}.</p><p>Happy hunting,<br/>Your Santa 🎅</p>",
			Sender:              "your.email+santa@gmail.com",
			RecipientsSeparator: " and ",
		},
		Participants: []data.Participant{
			{
				Email:              "mom@family.com",
				Person:             data.Person{Salutation: "Mom"},
				ExcludedRecipients: []string{"dad@family.com"},
			}, {
				Email:              "dad@family.com",
				Person:             data.Person{Salutation: "Dad"},
				ExcludedRecipients: []string{"mom@family.com"},
			}, {
				Email:  "auntie@family.com",
				Person: data.Person{Salutation: "Auntie"},
			}, {
				Email:                "emily@family.com",
				Person:               data.Person{Salutation: "Emily"},
				PredestinedRecipient: "auntie@family.com",
			}, {
				Email:  "jake@family.com",
				Person: data.Person{Salutation: "Jake"},
			}},
	}

	if err := data.SaveData(dataPath, dat); err != nil {
		return fmt.Errorf("could not create a data file: %w", err)
	}

	out := cmd.OutOrStdout()
	fmt.Fprintf(out, "Your data file '%s' has been created 🎉\n\n", dataPath)
	fmt.Fprintln(out, "Next, edit it (participants, email template and mailer) and run `santa check`.")
	return nil
}
