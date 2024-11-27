package app

import "math/rand"

func getRandomIndexInArray[T any](arr []T) int {
	return rand.Intn(len(arr))
}
