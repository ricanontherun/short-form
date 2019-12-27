package utils

import "strings"

func InArray(elem string, array []string) bool {
	for _, a := range array {
		if a == elem {
			return true
		}
	}

	return false
}

func ToCommaSeparatedString(array []string) string {
	return strings.Join(array, ", ")
}
