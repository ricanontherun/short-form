package utils

// Check if a string exists in a slice.
func InArray(elem string, array []string) bool {
	for _, a := range array {
		if a == elem {
			return true
		}
	}

	return false
}
