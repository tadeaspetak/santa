package app

import (
	"errors"
	"fmt"
	"slices"

	"github.com/tadeaspetak/santa/internal/data"
)

var errNoGiverForExtra = errors.New(
	"unable to assign an extra recipient, adjust your data, such as excluded givers, to make assigning extra recipients possible",
)

type extraWithGiver struct {
	data.Extra
	giver data.Participant
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

func pairExtras(extras []data.Extra, givers []data.Participant) ([]extraWithGiver, error) {
	extrasWithGivers := make([]extraWithGiver, len(extras))

	remainingGivers := getParticipantsByEmail(givers)

	// sort the extras so that the ones with the most `excludedGivers` are at the top
	// this is to improve our chances of assigning the extras more uniformly across the givers
	// note: sort a copy so that the caller's slice is left untouched
	sortedExtras := slices.Clone(extras)
	slices.SortStableFunc(sortedExtras, func(a data.Extra, b data.Extra) int {
		return len(b.ExcludedGivers) - len(a.ExcludedGivers)
	})

	for i, extra := range sortedExtras {
		potentialGivers := getPotentialGivers(extra, remainingGivers)

		// no more givers, let's do a new round
		if len(potentialGivers) == 0 {
			remainingGivers = getParticipantsByEmail(givers)
			potentialGivers = getPotentialGivers(extra, remainingGivers)
			// still no givers after the restart? let's throw
			if len(potentialGivers) == 0 {
				return nil, fmt.Errorf("%w: %s", errNoGiverForExtra, extra.Salutation)
			}
		}

		giver := potentialGivers[getRandomIndexInArray(potentialGivers)]
		extrasWithGivers[i] = extraWithGiver{Extra: extra, giver: giver}
		delete(remainingGivers, giver.Email)
	}

	return extrasWithGivers, nil
}
