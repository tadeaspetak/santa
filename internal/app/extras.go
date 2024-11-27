package app

import (
	"log"
	"slices"

	"github.com/tadeaspetak/santa/internal/data"
)

type extraWithGiver struct {
	data.Extra
	giver data.Participant
}

func getParticipantsByEmail(participants []data.Participant) map[string]data.Participant {
	byEmail := map[string]data.Participant{}
	for _, p := range participants {
		byEmail[p.Email] = p
	}
	return byEmail
}

func getPotentialGivers(extra data.Extra, givers map[string]data.Participant) []data.Participant {
	potentialGivers := make([]data.Participant, 0)
	for _, g := range givers {
		if !slices.Contains(extra.ExcludedGivers, g.Email) {
			potentialGivers = append(potentialGivers, g)
		}
	}
	return potentialGivers
}

func pairExtras(extras []data.Extra, givers []data.Participant) []extraWithGiver {
	extrasWithGivers := make([]extraWithGiver, len(extras))

	remainingGivers := getParticipantsByEmail(givers)

	// sort the extras so that the ones with the most `excludedGivers` are at the top
	// this is to improve our chances of assigning the extras more uniformly across the givers
	slices.SortFunc(extras, func(a data.Extra, b data.Extra) int {
		return len(b.ExcludedGivers) - len(a.ExcludedGivers)
	})

	for i, extra := range extras {
		potentialGivers := getPotentialGivers(extra, remainingGivers)

		// no more givers, let's do a new round
		if len(potentialGivers) == 0 {
			remainingGivers := getParticipantsByEmail(givers)
			potentialGivers = getPotentialGivers(extra, remainingGivers)
			// still no givers after the restart? let's throw
			if len(potentialGivers) == 0 {
				log.Fatalf(
					"Too many failed attempts to assign extra recipients! Try again or adjust your data, such as excluded givers, to make assigning extra recipients possible.",
				)
			}
		}

		giver := potentialGivers[getRandomIndexInArray(potentialGivers)]
		extrasWithGivers[i] = extraWithGiver{Extra: extra, giver: giver}
		delete(remainingGivers, giver.Email)
	}

	return extrasWithGivers
}
