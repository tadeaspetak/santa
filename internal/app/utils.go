package app

import (
	"math/rand"

	"github.com/tadeaspetak/santa/internal/data"
)

func getRandomIndexInArray[T any](arr []T) int {
	return rand.Intn(len(arr))
}

func shuffle[T any](arr []T) {
	rand.Shuffle(len(arr), func(i, j int) { arr[i], arr[j] = arr[j], arr[i] })
}

func getParticipantsByEmail(participants []data.Participant) map[string]data.Participant {
	byEmail := map[string]data.Participant{}
	for _, p := range participants {
		byEmail[p.Email] = p
	}
	return byEmail
}
