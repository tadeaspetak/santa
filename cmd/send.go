package cmd

import (
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/tadeaspetak/santa/internal/app"
	"github.com/tadeaspetak/santa/internal/app/mailer"
	"github.com/tadeaspetak/santa/internal/data"
)

var Validate *validator.Validate

var isDebugFlagName = "debug"
var alwaysSendToFlagName = "alwaysSendTo"
var shouldPrintPdfFlagName = "printPdf"

var translations map[string]string = map[string]string{
	"smtp-required_without":    "`smtp` or `mailgun` are required.",
	"mailgun-required_without": "`mailgun` or `smtp` are required.",
}

func validate(dat data.Data) error {
	validate := validator.New(validator.WithRequiredStructEnabled())
	err := validate.Struct(dat)

	removeData := regexp.MustCompile(`data.(.*)`)
	removeArray := regexp.MustCompile(`(participant|extra)s\[\d+\](.*)`)
	if err != nil {
		errs := []string{}
		for _, err := range err.(validator.ValidationErrors) {
			// e.g. `data.participants[1].email`, or `data.template.subject`
			namespace := err.StructNamespace()

			simplifiedNamespace :=
				// change the array syntax (`participants[1].email`) to a simpler one (`participant.email`)
				removeArray.ReplaceAllString(
					// remove the initial `data.`
					removeData.ReplaceAllString(
						// lowercase
						strings.ToLower(namespace),
						"$1"),
					"$1$2",
				)

			// e.g. `required_without` or
			tag := strings.ToLower(err.ActualTag())
			// compose the key so it ends up being e.g. `participant.email-required_without`
			key := fmt.Sprintf("%s-%s", simplifiedNamespace, tag)

			if v, ok := translations[key]; ok {
				errs = append(errs, fmt.Sprintf("%s: %s", namespace, v))
			} else {
				errs = append(errs, err.Error())
			}
		}

		return fmt.Errorf("%s", strings.Join(errs, "\n"))
	}

	return nil
}

func ask(question string) bool {
	prompt := promptui.Prompt{
		Label:     fmt.Sprint(question),
		IsConfirm: true,
		Default:   "y",
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

var sendCmd = &cobra.Command{
	Use:   "send",
	Short: "send the email",
	Run: func(cmd *cobra.Command, args []string) {
		dat, err := data.LoadData(getDataPath(cmd))
		if err != nil {
			if errors.Is(err, os.ErrNotExist) {
				log.Fatalf("Data file does not exist.")
			}
			log.Fatalf("Error loading config: %v", err)
		}

		// flags
		isDebug, err := cmd.Flags().GetBool(isDebugFlagName)
		if err != nil {
			log.Fatalf("Unable to get the %s flag: %v", isDebugFlagName, err)
		}

		alwaysSendTo, err := cmd.Flags().GetString(alwaysSendToFlagName)
		if err != nil {
			log.Fatalf("Unable to get the %s flag: %v", alwaysSendToFlagName, err)
		}

		shouldPrintPdf, err := cmd.Flags().GetBool(shouldPrintPdfFlagName)
		if err != nil {
			log.Fatalf("Unable to get the %s flag: %v", shouldPrintPdfFlagName, err)
		}

		// validate
		if err = validate(dat); err != nil {
			log.Fatalf("Invalid data!\n\n%s\n\n", err)
		}

		// mailer
		var mlr app.Mailer
		// note: checking the domain is enough, the rest has been validated above
		if dat.Mailgun != nil && dat.Mailgun.Domain != "" {
			mlr = mailer.NewMailgunMailer(dat.Mailgun.Domain, dat.Mailgun.APIKey)
		} else {
			mlr = mailer.NewSmtpMailer(dat.Smtp.Host, 587, dat.Smtp.User, dat.Smtp.Pass)
		}

		if !ask("We are ready to randomise and send the emails. Are YOU ready") {
			fmt.Println("Not sending then.")
			return
		}

		err = app.Send(
			mlr,
			app.Pair(dat.Participants, dat.Extras),
			*dat.Template,
			app.SendOpts{
				AlwaysSendTo:   alwaysSendTo,
				IsDebug:        isDebug,
				ShouldPrintPdf: shouldPrintPdf,
			},
		)

		if err != nil {
			log.Fatalf(`There has been an error sending your emails: %v`, err)
		}

		fmt.Println(`Your emails have been sent out successfully!`)
	},
}

func init() {
	sendCmd.Flags().BoolP(isDebugFlagName, "d", false, "turn the debug mode on (won't send emails)")
	sendCmd.Flags().
		StringP(alwaysSendToFlagName, "a", "", "send all emails to the given email address (for testing purposes)")
	sendCmd.Flags().
		BoolP(shouldPrintPdfFlagName, "p", false, "generate a printable pdf with the pairings")
}
