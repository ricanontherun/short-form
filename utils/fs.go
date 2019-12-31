package utils

import (
	"os"
	"strings"
)

func EnsureFilePath(filePath string) error {
	levels := strings.Split(filePath, string(os.PathSeparator))
	lastLevel := len(levels) - 1

	if err := os.MkdirAll(strings.Join(levels[0:lastLevel], string(os.PathSeparator)), os.ModePerm); err != nil {
		return nil
	}

	if _, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, os.ModePerm); err != nil {
		return err
	}

	return nil
}
