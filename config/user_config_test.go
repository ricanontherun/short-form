package config

import (
	"encoding/json"
	"fmt"
	"github.com/ricanontherun/short-form/utils"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"path"
	"testing"
)

const (
	basePath    = "./test_data"
	baseTmpPath = basePath + "/tmp"
)

func TestReadUserConfig_NewConfig(t *testing.T) {
	configPath := path.Join(baseTmpPath, configurationPath)
	_, statErr := os.Stat(configPath)
	assert.True(t, os.IsNotExist(statErr))

	config, readConfigErr := ReadUserConfig(baseTmpPath)
	assert.Nil(t, readConfigErr)
	assert.EqualValues(t, config.GetDatabasePath(), path.Join(baseTmpPath, defaultDatabasePath))

	// Assert the file written to disk is valid JSON.
	configBytes, readErr := ioutil.ReadFile(configPath)
	assert.Nil(t, readErr)
	var testConfig userConfig
	unmarshalErr := json.Unmarshal(configBytes, &testConfig)
	assert.Nil(t, unmarshalErr)

	assert.EqualValues(t, testConfig.DatabasePath, path.Join(baseTmpPath, defaultDatabasePath))
}

func TestReadUserConfig_ExistingConfig(t *testing.T) {
	config, readConfigErr := ReadUserConfig(basePath)

	assert.Nil(t, readConfigErr)
	assert.NotNil(t, config)
	assert.EqualValues(t, "/path/to/my/database.db", config.GetDatabasePath())
}

func TestReadUserConfig_UpdateConfig(t *testing.T) {
	config, readConfigErr := ReadUserConfig(baseTmpPath)
	assert.Nil(t, readConfigErr)
	assert.EqualValues(t, path.Join(baseTmpPath, defaultDatabasePath), config.GetDatabasePath())

	assert.Nil(t, config.SetDatabasePath("/new/path/to/database.db"))
	assert.Nil(t, config.Save())

	updatedConfig, readConfigErr := ReadUserConfig(baseTmpPath)
	assert.Nil(t, readConfigErr)
	assert.EqualValues(t, "/new/path/to/database.db", updatedConfig.GetDatabasePath())
}

func TestMain(m *testing.M) {
	if err := utils.CleanTestDir(baseTmpPath); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	code := m.Run()
	if err := utils.CleanTestDir(baseTmpPath); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	os.Exit(code)
}
