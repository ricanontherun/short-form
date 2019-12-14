package utils

import "testing"

func TestInArray(t *testing.T) {
	array := []string{
		"one",
		"two",
		"three",
	}

	for _, elem := range array {
		if !InArray(elem, array) {
			t.Errorf("%s should be in array", elem)
		}
	}

	for _, badElem := range []string{"four", "five", "six"} {
		if InArray(badElem, array) {
			t.Errorf("%s should not be array", badElem)
		}
	}
}
