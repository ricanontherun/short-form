package utils

import (
	"os"
	"strings"
)

// Create an entire file path, including any parent directories.
func EnsureFilePath(filePath string) error {
	if _, err := os.Stat(filePath); err != nil {
		if os.IsNotExist(err) {
			levels := strings.Split(filePath, string(os.PathSeparator))
			lastLevel := len(levels) - 1

			if err := os.MkdirAll(strings.Join(levels[0:lastLevel], string(os.PathSeparator)), os.ModePerm); err != nil {
				return nil
			}

			if _, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, os.ModePerm); err != nil {
				return err
			}
		} else {
			return err
		}
	}

	return nil
}
