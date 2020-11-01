package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ricanontherun/short-form/logging"
	"github.com/ricanontherun/short-form/utils"
	"io/ioutil"
	"path"
)

// configuration read from the user's ~/.sf directory.
// It's structure is as follows:
/**
{
  "database_path": "/path/to/database/path"
}
*/
type userConfig struct {
	DatabasePath string `json:"database_path"`

	basePath string
}

// Attempt to read and parse the configuration files present at ~/.sf
func ReadUserConfig(basePath string) (Config, error) {
	var configFilePath = path.Join(basePath, configurationPath)
	logging.Debug("reading configuration at " + configFilePath)

	existed, ensureFilePathErr := utils.EnsureFilePath(configFilePath)
	if ensureFilePathErr != nil {
		return nil, ensureFilePathErr
	}

	// Setup initial configuration if it doesn't already exist.
	if !existed {
		var userConfig = newUserConfig(basePath)
		if userConfigBytes, err := json.Marshal(userConfig); err != nil {
			return nil, errors.New(fmt.Sprintf("failed to marshal default config, %s", err.Error()))
		} else { // Write default config to disk.
			if err := ioutil.WriteFile(configFilePath, userConfigBytes, 0644); err != nil {
				return nil, err
			}
			return userConfig, nil
		}
	} else {
		configBytes, err := ioutil.ReadFile(configFilePath)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("failed to open config file (%s), %s", configFilePath, err.Error()))
		}

		var config userConfig
		if err = json.Unmarshal(configBytes, &config); err != nil {
			return nil, errors.New(fmt.Sprintf("failed to parse config file (%s), %s", configFilePath, err.Error()))
		}

		config.basePath = basePath
		return &config, nil
	}
}

func newUserConfig(basePath string) Config {
	return &userConfig{
		DatabasePath: path.Join(basePath, defaultDatabasePath),
		basePath:     basePath,
	}
}

// Save the current state of the config to disk.
func (config *userConfig) Save() error {
	if userConfigBytes, err := json.Marshal(config); err == nil {
		if err := ioutil.WriteFile(path.Join(config.basePath, configurationPath), userConfigBytes, 0644); err != nil {
			return errors.New(fmt.Sprintf("failed to save user config to disk, %s", err.Error()))
		}
	} else {
		return err
	}
	return nil
}

func (config *userConfig) GetDatabasePath() string {
	return config.DatabasePath
}

func (config *userConfig) SetDatabasePath(path string) error {
	if len(path) == 0 {
		return errors.New("database path should be a non-empty string")
	}
	config.DatabasePath = path
	return nil
}
