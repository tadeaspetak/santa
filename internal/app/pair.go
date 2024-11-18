package app

import (
	"log"
	"slices"

	"math/rand"

	"github.com/tadeaspetak/santa/internal/data"
)

func getRandomIndexInArray[T any](arr []T) int {
	return rand.Intn(len(arr))
}

type participantPair struct {
	giver     data.Participant
	recipient data.Participant
}

func PairParticipants(participants []data.Participant, maxAttemptCount int) []participantPair {
	if maxAttemptCount <= 0 {
		log.Fatalf(
			"Too many failed attempts to pair participants! Try again or adjust your data, such as predestined or excluded recipients, to make pairing participants possible.",
		)
	}

	participantsByMail := map[string]data.Participant{}
	predestinedEmails := make([]string, 0)
	for _, p := range participants {
		participantsByMail[p.Email] = p
		if p.PredestinedRecipient != "" {
			predestinedEmails = append(predestinedEmails, p.PredestinedRecipient)
		}
	}

	// potential recipients start as everyone except the predestined ones
	remainingPotentialRecipients := map[string]data.Participant{}
	for _, p := range participants {
		participantsByMail[p.Email] = p
		if slices.Index(predestinedEmails, p.Email) == -1 {
			remainingPotentialRecipients[p.Email] = p
		}
	}

	paired := make([]participantPair, len(participants))
	for i, giver := range participants {
		// the recipient is predestined, just assign them
		if giver.PredestinedRecipient != "" {
			recipient, ok := participantsByMail[giver.PredestinedRecipient]
			if !ok {
				log.Fatalf("Predestined recipient %s does not exist.", giver.PredestinedRecipient)
			}
			paired[i] = participantPair{giver: giver, recipient: recipient}
			continue
		}

		actualPotentialRecipients := make([]data.Participant, 0)
		for _, p := range remainingPotentialRecipients {
			if giver.Email != p.Email && slices.Index(giver.ExcludedRecipients, p.Email) == -1 {
				actualPotentialRecipients = append(actualPotentialRecipients, p)
			}
		}

		if len(actualPotentialRecipients) == 0 {
			return PairParticipants(participants, maxAttemptCount-1)
		}

		// get a random recipient, add to the paried and remove from potential
		actualRecipient := actualPotentialRecipients[getRandomIndexInArray(actualPotentialRecipients)]
		paired[i] = participantPair{giver: giver, recipient: actualRecipient}
		delete(remainingPotentialRecipients, actualRecipient.Email)
	}

	return paired
}
