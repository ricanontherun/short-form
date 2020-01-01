package utils

// SliceContainsElement checks if an element is in a slice.
func SliceContainsElement(elem string, array []string) bool {
	for _, a := range array {
		if a == elem {
			return true
		}
	}

	return false
}
