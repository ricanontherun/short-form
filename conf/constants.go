package conf

import (
	"log"
	"os/user"
	"path"
	"sync"
)

const (
	shortFormDirectory = ".sf"
	dataDirectory      = "data"
	dataFile           = "data.db"
)

var shortFormDirectoryPath string
var shortFormDataDirectoryPath string
var shortFormDatabasePath string

var once sync.Once

func initializePaths() {
	once.Do(func() {
		u, err := user.Current()
		if err != nil {
			log.Fatalf("Failed to start the short-form: %s\n", err.Error())
		}

		homeDirectory := u.HomeDir
		shortFormDirectoryPath = path.Join(homeDirectory, shortFormDirectory)
		shortFormDataDirectoryPath = path.Join(shortFormDirectoryPath, dataDirectory)
		shortFormDatabasePath = path.Join(shortFormDataDirectoryPath, dataFile)
	})
}

// ResolveDatabaseFilePath returns the absolute path to the user's database file.
func ResolveDatabaseFilePath() string {
	initializePaths()

	return shortFormDatabasePath
}
