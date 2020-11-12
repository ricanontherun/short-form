package utils

import (
	"os"
	"strings"
)

// Create an entire file path, including any parent directories.
// Returns if the path existed and any errors
func EnsureFilePath(filePath string) (bool, error) {
	var exists = true

	// Does the file already exists?
	_, statErr := os.Stat(filePath)
	if statErr == nil {
		return true, nil
	}

	exists = false
	levels := strings.Split(filePath, string(os.PathSeparator))
	lastLevel := len(levels) - 1

	if err := os.MkdirAll(strings.Join(levels[0:lastLevel], string(os.PathSeparator)), os.ModePerm); err != nil {
		return exists, err
	}

	if _, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, os.ModePerm); err != nil {
		return exists, err
	}

	return exists, nil
}
