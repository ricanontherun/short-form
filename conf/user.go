package conf

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ricanontherun/short-form/utils"
	"io/ioutil"
	"os"
	"os/user"
	"path"
)

type userConfig struct {
	DatabasePath string `json:"database_path"`

	user *user.User
}

type Config interface {
	GetDatabasePath() string
	SetDatabasePath(path string) error
	Save() error
}

// Save the current state of the config to disk.
func (config *userConfig) Save() error {
	if userConfigBytes, err := json.Marshal(config); err != nil {
		return err
	} else {
		if err := ioutil.WriteFile(path.Join(config.user.HomeDir, shortFormConfigurationPath), userConfigBytes, 0644); err != nil {
			return errors.New(fmt.Sprintf("failed to save user config to disk, %s", err.Error()))
		}
	}

	return nil
}

func (config *userConfig) GetDatabasePath() string {
	return config.DatabasePath
}

func (config *userConfig) SetDatabasePath(path string) error {
	// Validate the path as being legit.
	// If the file at path is non-empty ... should we warn the user?
	if len(path) == 0 {
		return errors.New("cannot set database path, empty string")
	}

	// The path should at least be present, not non-empty necessarily.

	config.DatabasePath = path
	return nil
}

func newUserConfig(user *user.User) Config {
	return &userConfig{
		DatabasePath: path.Join(user.HomeDir, shortFormDefaultDatabasePath),
		user:         user,
	}
}

func ReadUserConfig() (Config, error) {
	systemUser, err := user.Current()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("failed to read user's home directory, %s", err.Error()))
	}

	configFilePath := path.Join(systemUser.HomeDir, shortFormConfigurationPath)
	_, err = os.Stat(configFilePath)

	existed, err := utils.EnsureFilePath(configFilePath)
	if err != nil {
		return nil, err
	}

	if !existed {
		var userConfig = newUserConfig(systemUser)

		if userConfigBytes, err := json.Marshal(userConfig); err != nil {
			return nil, errors.New(fmt.Sprintf("failed to marshal default config, %s", err.Error()))
		} else { // Write default config to disk.
			if err := ioutil.WriteFile(configFilePath, userConfigBytes, 0644); err != nil {
				return nil, err
			}

			return userConfig, nil
		}
	}

	configBytes, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("failed to open config file (%s), %s", configFilePath, err.Error()))
	}

	var config userConfig
	if err = json.Unmarshal(configBytes, &config); err != nil {
		return nil, errors.New(fmt.Sprintf("failed to parse config file (%s), %s", configFilePath, err.Error()))
	}

	config.user = systemUser
	return &config, nil
}
