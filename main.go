package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ricanontherun/short-form/conf"
	"github.com/ricanontherun/short-form/data"
	"github.com/ricanontherun/short-form/handler"
	"github.com/urfave/cli"
	"io/ioutil"
	"log"
	"os"
)

var (
	ErrFailedToReadConfigFile         = errors.New("failed to read configuration file contents")
	ErrFailedToParseConfigurationFile = errors.New("failed to parse configuration file contents")
)

func getDefaultConfiguration() conf.ShortFormConfig {
	return conf.ShortFormConfig{Secret: "", Secure: false}
}

func setSecret(config conf.ShortFormConfig, secret string) (conf.ShortFormConfig, error) {
	config.Secret = secret
	config.Secure = true

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
	configFilePath := conf.ResolveConfigurationFilePath()

	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		log.Printf("Creating default configuration at %s\n", configFilePath)

		if file, err := os.Create(configFilePath); err != nil {

			return nil, err
		} else {
			defer file.Close()

			if jsonBytes, err := json.Marshal(getDefaultConfiguration()); err != nil {
				return nil, err
			} else {
				file.WriteString(string(jsonBytes))
			}
		}
	} else if err != nil {
		return nil, err
	}

	fileContents, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return nil, ErrFailedToReadConfigFile
	}

	var c conf.ShortFormConfig
	if err := json.Unmarshal(fileContents, &c); err != nil {
		return nil, ErrFailedToParseConfigurationFile
	}

	return &c, nil
}

func main() {
	config, err := getConfiguration()
	if err != nil {
		log.Fatalf("Failed to start app: %s\n", err.Error())
	}

	fmt.Println(config)

	repository, err := data.NewRepository(*config)
	if err != nil {
		log.Fatalf("Failed to open database: %s", err.Error())
	}
	defer repository.Close()

	handler := handler.Handler{Repository: repository}

	app := cli.App{
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
					&cli.StringFlag{
						Name:        "tags",
						Usage:       "Comma separated list of tags",
						Value:       "",
						DefaultText: "",
						Aliases:     []string{"t"},
					},
				},
				Action: handler.WriteNote,
			},
			{
				Name:    "search",
				Aliases: []string{"s"},
				Subcommands: []*cli.Command{
					// Search against today's notes.
					{
						Name:    "today",
						Aliases: []string{"t"},
						Action:  handler.SearchTodayNote,
					},

					// Search against yesterday's notes.
					{
						Name:    "yesterday",
						Aliases: []string{"y"},
						Action:  handler.SearchYesterdayNote,
					},
				},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:    "since",
						Usage:   "-s 5d",
						Value:   "",
						Aliases: []string{"s"},
					},
					&cli.StringFlag{
						Name:    "tags",
						Usage:   "-t music",
						Value:   "",
						Aliases: []string{"t"},
					},
				},
				Action: func(c *cli.Context) error {
					return nil
				},
			},
		},
	}

	if err = app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
