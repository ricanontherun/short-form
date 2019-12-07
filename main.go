package main

import (
	"encoding/json"
	"errors"
	"github.com/ricanontherun/short-form/conf"
	"github.com/ricanontherun/short-form/data"
	"github.com/ricanontherun/short-form/handler"
	"github.com/ricanontherun/short-form/utils"
	"github.com/urfave/cli/v2"
	"io/ioutil"
	"log"
	"os"
)

var (
	ErrFailedToReadConfigFile         = errors.New("failed to read configuration file contents")
	ErrFailedToParseConfigurationFile = errors.New("failed to parse configuration file contents")
)

func getDefaultConfiguration() conf.ShortFormConfig {
	return conf.ShortFormConfig{Secret: ""}
}

var (
	FlagTags = &cli.StringFlag{
		Name:    "tags",
		Aliases: []string{"t"},
		Usage:   "comma,separated,list of tags to filter on.",
		Value:   "",
	}
)

func setSecret(config conf.ShortFormConfig, secret string) (conf.ShortFormConfig, error) {
	config.Secret = secret

	configFilePath := conf.ResolveConfigurationFilePath()
	if file, err := os.Open(configFilePath); err != nil {
		return config, err
	} else {
		defer file.Close()

		if configBytes, err := json.Marshal(config); err != nil {
			return config, err
		} else {
			if _, err := file.Write(configBytes); err != nil {
				return config, err
			}
		}
	}

	return config, nil
}

func getConfiguration() (*conf.ShortFormConfig, error) {
	// Make sure all known files are created.
	configFilePath := conf.ResolveConfigurationFilePath()
	if err := utils.EnsureFilePaths(
		configFilePath, conf.ResolveDatabaseFilePath()); err != nil {
		return nil, err
	}

	fileContents, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return nil, err
	}

	if len(fileContents) == 0 { // Create default configuration
		defaultConfig := getDefaultConfiguration()
		jsonBytes, err := json.Marshal(defaultConfig)

		if err != nil {
			return nil, err
		}

		if err := ioutil.WriteFile(configFilePath, jsonBytes, os.ModePerm); err != nil {
			return nil, errors.New("failed to write default configuration file to disk: " + err.Error())
		}

		return &defaultConfig, nil
	} else { // Attempt to parse existing config file.
		var c conf.ShortFormConfig
		if err := json.Unmarshal(fileContents, &c); err != nil {
			return nil, ErrFailedToParseConfigurationFile
		}

		return &c, nil
	}
}

func main() {
	config, err := getConfiguration()
	if err != nil {
		log.Fatalln(err)
	}

	encryptor := utils.MakeEncryptor(config.Secret)

	repository, err := data.NewSqlRepository(encryptor)
	if err != nil {
		log.Fatalf("Failed to open database: %s", err.Error())
	}
	defer repository.Close()

	handler := handler.NewHandler(repository, encryptor)

	app := cli.App{
		Name:        "short-form",
		Usage:       "A command-line journal for bite sized thoughts",
		Description: "short-form is a privacy focused, command-line journal.",
		Version:     "1.0.0",
		Commands: []*cli.Command{
			//{
			//	Name:    "configure",
			//	Aliases: []string{"c"},
			//	Usage:   "configure",
			//	Flags: []cli.Flag{
			//		&cli.StringFlag{
			//			Name:        "path",
			//			Usage:       "--path",
			//			Value:       "",
			//			DefaultText: "uh",
			//			Aliases:     []string{"p"},
			//		},
			//	},
			//},
			{
				Name:    "write",
				Aliases: []string{"w"},
				Usage:   "Write a new note",
				Flags: []cli.Flag{
					FlagTags,
				},
				Action: handler.WriteInsecureNote,
			},
			{
				Name:    "write-secure",
				Aliases: []string{"ws"},
				Usage:   "Write a new secure (encrypted) note",
				Flags: []cli.Flag{
					FlagTags,
				},
				Action: handler.WriteSecureNote,
			},
			{
				Name:    "search",
				Usage:   "Search for notes by tag, date",
				Aliases: []string{"s"},
				Subcommands: []*cli.Command{
					// Search against today's notes.
					{
						Name:    "today",
						Usage:   "Search for notes written today",
						Aliases: []string{"t"},
						Action:  handler.SearchTodayNote,
					},

					// Search against yesterday's notes.
					{
						Name:    "yesterday",
						Usage:   "Search for notes written yesterday",
						Aliases: []string{"y"},
						Action:  handler.SearchYesterdayNote,
					},
				},
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:    "insecure",
						Usage:   "Search in insecure mode, encrypted notes will be decrypted when displayed",
						Aliases: []string{"i"},
						Value:   false,
					},
					FlagTags,
				},
				Action: handler.SearchNotes,
			},
		},
	}

	if err = app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
