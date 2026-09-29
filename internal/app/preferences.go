package app

import (
	"errors"
	"fmt"
	"slices"

	"github.com/tadeaspetak/santa/internal/data"
)

var errUnknownPredestinedRecipient = errors.New("predestined recipient does not exist")
var errSelfPredestinedRecipient = errors.New("participant cannot be their own predestined recipient")
var errDuplicatePredestinedRecipient = errors.New("participant is a predestined recipient of multiple givers")
var errExcludedPredestinedRecipient = errors.New("predestined recipient is also excluded")
var errUnknownExcludedRecipient = errors.New("excluded recipient does not exist")
var errUnknownExcludedGiver = errors.New("excluded giver does not exist")

// validatePreferences makes sure the preferences (predestined recipients, excluded recipients
// and excluded givers) reference existing participants and don't contradict each other.
// All the problems are reported at once so that they can be fixed in one go.
func validatePreferences(participants []data.Participant, extras []data.Extra) error {
	participantsByEmail := getParticipantsByEmail(participants)
	errs := []error{}

	predestinedBy := map[string]string{}
	for _, p := range participants {
		for _, excluded := range p.ExcludedRecipients {
			if _, ok := participantsByEmail[excluded]; !ok {
				errs = append(errs, fmt.Errorf("%w: %s (excluded by %s)", errUnknownExcludedRecipient, excluded, p.Email))
			}
		}

		predestined := p.PredestinedRecipient
		if predestined == "" {
			continue
		}
		if _, ok := participantsByEmail[predestined]; !ok {
			errs = append(errs, fmt.Errorf("%w: %s (predestined for %s)", errUnknownPredestinedRecipient, predestined, p.Email))
		}
		if predestined == p.Email {
			errs = append(errs, fmt.Errorf("%w: %s", errSelfPredestinedRecipient, predestined))
		}
		if slices.Contains(p.ExcludedRecipients, predestined) {
			errs = append(errs, fmt.Errorf("%w: %s (by %s)", errExcludedPredestinedRecipient, predestined, p.Email))
		}
		if other, ok := predestinedBy[predestined]; ok {
			errs = append(errs, fmt.Errorf("%w: %s (predestined for %s and %s)", errDuplicatePredestinedRecipient, predestined, other, p.Email))
		} else {
			predestinedBy[predestined] = p.Email
		}
	}

	for _, e := range extras {
		for _, excluded := range e.ExcludedGivers {
			if _, ok := participantsByEmail[excluded]; !ok {
				errs = append(errs, fmt.Errorf("%w: %s (excluded for %s)", errUnknownExcludedGiver, excluded, e.Salutation))
			}
		}
	}

	return errors.Join(errs...)
}
