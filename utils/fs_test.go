package utils

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

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

func TestMain(m *testing.M) {
	if err := CleanTestDir("./test_data"); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	code := m.Run()
	if err := CleanTestDir("./test_data"); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	os.Exit(code)
}
