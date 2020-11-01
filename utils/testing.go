package utils

import (
	"fmt"
	"os"
)

func CleanTestDir(dir string) error {
	removeErr := os.RemoveAll(dir)
	if removeErr != nil {
		return fmt.Errorf("failed to delete %s, %s", dir, removeErr.Error())
	}

	mkdirErr := os.Mkdir(dir, os.ModePerm)
	if mkdirErr != nil {
		return fmt.Errorf("failed to recreate %s, %s", dir, mkdirErr.Error())
	}

	return nil
}
