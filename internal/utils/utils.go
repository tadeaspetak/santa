package utils

// remove an element at a given index from a slice
// while preserving order (https://stackoverflow.com/a/37335777/3844098).
func RemoveFromSlice[K any](slice []K, index int) []K {
	return append(slice[:index], slice[index+1:]...)
}

func Filter[K any](slice []K, predicate func(item K) bool) []K {
	result := []K{}
	for _, item := range slice {
		if predicate(item) {
			result = append(result, item)

		}
	}
	return result
}
