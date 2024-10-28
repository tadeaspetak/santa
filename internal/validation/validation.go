package validation

import (
	"strings"

	"github.com/go-playground/validator/v10"
)

var Validate *validator.Validate

func init() {
	Validate = validator.New(validator.WithRequiredStructEnabled())
}

func SanitizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
