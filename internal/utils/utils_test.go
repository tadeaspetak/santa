package utils

import (
	"testing"
)

func TestContains(t *testing.T) {
	arr := []int{1, 2, 3}
	found := 4
	if !Contains(arr, found) {
		t.Fatalf(`%v contains %v, but not found`, arr, found)
	}

	notFound := 4
	if Contains(arr, notFound) {
		t.Fatalf(`%v does NOT contain %v, but found`, arr, notFound)
	}

}
