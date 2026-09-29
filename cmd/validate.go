package cmd

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/tadeaspetak/santa/internal/data"
)

var translations map[string]string = map[string]string{
	"smtp-required_without":    "`smtp` or `mailgun` are required.",
	"mailgun-required_without": "`mailgun` or `smtp` are required.",
}

// validate validates the data. The mailer config (`smtp` or `mailgun`) is only required
// when `requireMailer` is set, so that e.g. dry runs work without any credentials.
func validate(dat data.Data, requireMailer bool) error {
	validate := validator.New(validator.WithRequiredStructEnabled())
	err := validate.Struct(dat)

	removeData := regexp.MustCompile(`data.(.*)`)
	removeArray := regexp.MustCompile(`(participant|extra)s\[\d+\](.*)`)
	if err != nil {
		errs := []string{}
		for _, err := range err.(validator.ValidationErrors) {
			// e.g. `required_without` or
			tag := strings.ToLower(err.ActualTag())
			if !requireMailer && tag == "required_without" {
				continue
			}

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

			// compose the key so it ends up being e.g. `participant.email-required_without`
			key := fmt.Sprintf("%s-%s", simplifiedNamespace, tag)

			if v, ok := translations[key]; ok {
				errs = append(errs, fmt.Sprintf("%s: %s", namespace, v))
			} else {
				errs = append(errs, err.Error())
			}
		}

		if len(errs) > 0 {
			return fmt.Errorf("%s", strings.Join(errs, "\n"))
		}
	}

	return nil
}

// validateEmail validates a single email address, e.g. one passed via a flag.
func validateEmail(email string) error {
	return validator.New().Var(email, "email")
}
