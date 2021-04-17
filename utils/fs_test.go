package utils

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)


func cleanTestDir(dir string) error {
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


func removeFileIfExists(path string) {
	_, statErr := os.Stat(path)
	if statErr != nil || os.IsNotExist(statErr) {
		return
	}

	if removeErr := os.Remove(path); removeErr != nil {
		panic(removeErr)
	}
}

func TestEnsureFilePath_ExistingFile(t *testing.T) {
	path := "./fs_test.go"
	exists, err := EnsureFilePath(path)

	assert.True(t, exists)
	assert.Nil(t, err)
}

// write test data to /tmp
func TestEnsureFilePath_SingleFile(t *testing.T) {
	testFile := "./test_data/single.txt"
	exists, err := EnsureFilePath(testFile)

	assert.Nil(t, err)
	assert.False(t, exists)

	// Stat for good measure.
	_, statErr := os.Stat(testFile)
	assert.Nil(t, statErr)
}

func TestEnsureFilePath_MultiLevelFile(t *testing.T) {
	testFile := "./test_data/foo/bar/bazz.txt"
	exists, err := EnsureFilePath(testFile)

	assert.False(t, exists)
	assert.Nil(t, err)

	_, statErr := os.Stat(testFile)
	assert.Nil(t, statErr)
}

func TestEnsureFilePath_NoDirectory(t *testing.T) {
	testFile := "./test.db"
	exists, err := EnsureFilePath(testFile)

	assert.False(t, exists)
	assert.Nil(t, err)

	_, statErr := os.Stat(testFile)
	assert.Nil(t, statErr)
}

func TestMain(m *testing.M) {
	removeFileIfExists("./test.db")
	if err := cleanTestDir("./test_data"); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	code := m.Run()

	removeFileIfExists("./test.db")
	if err := cleanTestDir("./test_data"); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	os.Exit(code)
}
