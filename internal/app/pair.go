package app

import (
	"errors"
	"slices"

	"github.com/tadeaspetak/santa/internal/data"
)

var errNoValidPairing = errors.New(
	"no valid pairing exists, adjust your data, such as predestined or excluded recipients, to make pairing participants possible",
)

type giverWithRecipients struct {
	giver      data.Participant
	recipients []data.Person
}

// pairParticipants assigns a recipient to each participant. It goes through the givers in order,
// trying their potential recipients in random order and backtracking whenever it reaches a dead
// end, so a pairing is always found if one exists.
//
// It expects the preferences to have been checked by `validatePreferences`.
func pairParticipants(participants []data.Participant) ([]giverWithRecipients, error) {
	participantsByEmail := getParticipantsByEmail(participants)
	predestinedEmails := make([]string, 0)
	for _, p := range participants {
		if p.PredestinedRecipient != "" {
			predestinedEmails = append(predestinedEmails, p.PredestinedRecipient)
		}
	}

	// recipients start as everyone except the predestined ones
	remainingRecipients := map[string]data.Participant{}
	for _, p := range participants {
		if !slices.Contains(predestinedEmails, p.Email) {
			remainingRecipients[p.Email] = p
		}
	}

	paired := make([]giverWithRecipients, len(participants))

	var assign func(i int) bool
	assign = func(i int) bool {
		if i == len(participants) {
			return true
		}

		giver := participants[i]

		// the recipient is predestined, just assign them
		if giver.PredestinedRecipient != "" {
			recipient := participantsByEmail[giver.PredestinedRecipient]
			paired[i] = giverWithRecipients{giver: giver, recipients: []data.Person{recipient.Person}}
			return assign(i + 1)
		}

		potentialRecipients := make([]data.Participant, 0)
		for _, p := range remainingRecipients {
			if giver.Email != p.Email && !slices.Contains(giver.ExcludedRecipients, p.Email) {
				potentialRecipients = append(potentialRecipients, p)
			}
		}
		shuffle(potentialRecipients)

		for _, recipient := range potentialRecipients {
			paired[i] = giverWithRecipients{giver: giver, recipients: []data.Person{recipient.Person}}
			delete(remainingRecipients, recipient.Email)
			if assign(i + 1) {
				return true
			}
			// dead end, put the recipient back and try the next one
			remainingRecipients[recipient.Email] = recipient
		}

		return false
	}

	if !assign(0) {
		return nil, errNoValidPairing
	}

	return paired, nil
}

func Pair(participants []data.Participant, extraRecipients []data.Extra) ([]giverWithRecipients, error) {
	if err := validatePreferences(participants, extraRecipients); err != nil {
		return nil, err
	}

	paired, err := pairParticipants(participants)
	if err != nil {
		return nil, err
	}

	if len(extraRecipients) > 0 {
		extras, err := pairExtras(extraRecipients, participants)
		if err != nil {
			return nil, err
		}
		for _, e := range extras {
			i := slices.IndexFunc(paired, func(g giverWithRecipients) bool { return g.giver.Email == e.giver.Email })
			paired[i].recipients = append(paired[i].recipients, e.Person)
		}
	}

	return paired, nil
}
