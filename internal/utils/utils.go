package utils

import "math/rand"

// GetRandomIndexInArray gets a random index within the given array
func GetRandomIndexInArray[K any](arr []K) int {
	return rand.Intn(len(arr))
}

// Contains attempts to find a needle in the haystack
func Contains[K comparable](haystack []K, needle K) bool {
	for _, value := range haystack {
		if value == needle {
			return true
		}
	}
	return false
}
