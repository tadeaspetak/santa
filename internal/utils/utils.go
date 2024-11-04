package utils

import (
	"math/rand"
)

// remove an element at a given index from a slice
// while preserving order (https://stackoverflow.com/a/37335777/3844098).
func RemoveFromSlice[K any](slice []K, index int) []K {
	return append(slice[:index], slice[index+1:]...)
}

func Filter[K any](slice []K, predicate func(item K, i int) bool) []K {
	result := []K{}
	for i, item := range slice {
		if predicate(item, i) {
			result = append(result, item)

		}
	}
	return result
}

// GetRandomIndexInArray gets a random index within a given array
func GetRandomIndexInArray[T any](arr []T) int {
	return rand.Intn(len(arr))
}
