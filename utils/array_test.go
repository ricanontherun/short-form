package utils

import "testing"

func TestInArray(t *testing.T) {
	array := []string{
		"one",
		"two",
		"three",
	}

	tests := []struct {
		elem        string
		shouldExist bool
	}{
		{"one", true},
		{"two", true},
		{"three", true},
		{"1", false},
		{"no", false},
		{"", false},
	}

	for _, test := range tests {
		if InArray(test.elem, array) != test.shouldExist {
			if test.shouldExist {
				t.Fatalf("'%s' should exist in the array, but it does not", test.elem)
			} else {
				t.Fatalf("'%s' should not exist in the array, but it does", test.elem)
			}
		}
	}
}
