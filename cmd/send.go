package cmd

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/spf13/cobra"
	"github.com/tadeaspetak/santa/cmd/cmdData"
	"github.com/tadeaspetak/santa/internal/app"
	"github.com/tadeaspetak/santa/internal/app/mailer"
	"github.com/tadeaspetak/santa/internal/data"
	"github.com/tadeaspetak/santa/internal/validation"
)

var isDebugFlagName = "debug"
var alwaysSendToFlagName = "alwaysSendTo"
var shouldPrintPdfFlagName = "printPdf"

var translations map[string]string = map[string]string{
	"smtp-required_without":    "`smtp` or `mailgun` are required.",
	"mailgun-required_without": "`mailgun` or `smtp` are required.",
}

func validate(dat data.Data) error {
	// validation
	removeData := regexp.MustCompile(`data.(.*)`)
	removeArray := regexp.MustCompile(`(participant|extra)s\[\d+\](.*)`)
	err := validation.Validate.Struct(dat)
	if err != nil {
		errs := []string{}
		for _, err := range err.(validator.ValidationErrors) {
			// e.g. `cmddata.data.participants[1].email`, or `cmddata.data.template.subject`
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

var sendCmd = &cobra.Command{
	Use:   "send",
	Short: "send the email",
	Run: func(cmd *cobra.Command, args []string) {
		dat := (&cmdData.CmdData{}).Load(cmd)

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
		if err = validate(dat.Data); err != nil {
			log.Fatalf("Invalid data!\n\n%s\n\n", err)
		}

		// mailer
		var mlr app.Mailer
		// note: checking the domain is enough, the rest has been validated above
		if dat.Mailgun.Domain != "" {
			mlr = mailer.NewMailgunMailer(dat.Mailgun.Domain, dat.Mailgun.APIKey)
		} else {
			mlr = mailer.NewSmtpMailer(dat.Smtp.Host, 587, dat.Smtp.User, dat.Smtp.Pass)
		}

		err = app.Send(
			mlr,
			app.Pair(dat.Participants, dat.Extras),
			dat.Data.Template,
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
