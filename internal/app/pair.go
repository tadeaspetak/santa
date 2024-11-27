package app

import (
	"log"
	"slices"

	"github.com/tadeaspetak/santa/internal/data"
)

type giverWithRecipients struct {
	giver      data.Participant
	recipients []data.Person
}

func pairParticipants(participants []data.Participant, maxAttemptCount int) []giverWithRecipients {
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

	// recipients start as everyone except the predestined ones
	remainingRecipients := map[string]data.Participant{}
	for _, p := range participants {
		if slices.Index(predestinedEmails, p.Email) == -1 {
			remainingRecipients[p.Email] = p
		}
	}

	paired := make([]giverWithRecipients, len(participants))
	for i, giver := range participants {
		// the recipient is predestined, just assign them
		if giver.PredestinedRecipient != "" {
			recipient, ok := participantsByMail[giver.PredestinedRecipient]
			if !ok {
				log.Fatalf("Predestined recipient %s does not exist.", giver.PredestinedRecipient)
			}
			paired[i] = giverWithRecipients{giver: giver, recipients: []data.Person{recipient.Person}}
			continue
		}

		potentialRecipients := make([]data.Participant, 0)
		for _, p := range remainingRecipients {
			if giver.Email != p.Email && slices.Index(giver.ExcludedRecipients, p.Email) == -1 {
				potentialRecipients = append(potentialRecipients, p)
			}
		}

		if len(potentialRecipients) == 0 {
			return pairParticipants(participants, maxAttemptCount-1)
		}

		// get a random recipient, add to the paried and remove from potential
		recipient := potentialRecipients[getRandomIndexInArray(potentialRecipients)]
		paired[i] = giverWithRecipients{giver: giver, recipients: []data.Person{recipient.Person}}
		delete(remainingRecipients, recipient.Email)
	}

	return paired
}

func Pair(participants []data.Participant, extraRecipients []data.Extra) []giverWithRecipients {
	paired := pairParticipants(participants, 5)

	if len(extraRecipients) > 0 {
		extras := pairExtras(extraRecipients, participants)
		for _, e := range extras {
			i := slices.IndexFunc(paired, func(g giverWithRecipients) bool { return g.giver.Email == e.giver.Email })
			paired[i].recipients = append(paired[i].recipients, e.Person)
		}
	}

	return paired
}
