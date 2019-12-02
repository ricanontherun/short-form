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
	configurationFile  = "c"
	logDirectory       = "log"
)

var shortFormDirectoryPath string
var shortFormDataDirectoryPath string
var configFilePath string

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
		configFilePath = path.Join(shortFormDirectoryPath, configurationFile)
	})
}

func ResolveDataDirectory() string {
	initializePaths()

	return shortFormDataDirectoryPath
}

func ResolveConfigurationFilePath() string {
	initializePaths()

	return configFilePath
}

func ResolveHomeDirectory() string {
	initializePaths()

	return shortFormDataDirectoryPath
}
