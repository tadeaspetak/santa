package main

import "math/rand"

func getRandomIndexInArray[K any](arr []K) int {
	return rand.Intn(len(arr))
}

func contains[K comparable](haystack []K, needle K) bool {
	for _, value := range haystack {
		if value == needle {
			return true
		}
	}
	return false
}
