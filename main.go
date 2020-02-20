package main

import (
	"encoding/json"
	"fmt"
	"github.com/ricanontherun/short-form/command"
	"github.com/ricanontherun/short-form/conf"
	"github.com/ricanontherun/short-form/database"
	"github.com/ricanontherun/short-form/repository"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"os/signal"
	"syscall"
)

var (
	tagFlag = &cli.StringFlag{
		Name:    "tags",
		Aliases: []string{"t"},
		Usage:   "comma,separated,list of tags to filter on.",
		Value:   "",
	}

	appVersion = "1.4.0"
)

var searchFlags = []cli.Flag{
	tagFlag,
	&cli.StringFlag{
		Name:    "content",
		Usage:   "Search by note content",
		Aliases: []string{"c"},
		Value:   "",
	},
	&cli.StringFlag{
		Name:    "age",
		Usage:   "Search by age of note, e.g 2d for 2 days old",
		Aliases: []string{"a"},
		Value:   "",
	},
	&cli.BoolFlag{
		Name:    "detailed",
		Aliases: []string{"d"},
		Usage:   "Display detailed note information",
		Value:   false,
	},
}

func dd(message string) {
	fmt.Println(message)
	os.Exit(1)
}

func startupError(err error) {
	dd("Failed to start sf: " + err.Error())
}

// Setup signal handlers so that users can back out of multi-step operations.
func setupSignalHandlers() {
	signalChan := make(chan os.Signal, 2)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-signalChan
		os.Exit(1)
	}()
}

func main() {
	userConfig, err := conf.ReadUserConfig()
	if err != nil {
		log.Fatalln(err)
	}

	db := database.NewDatabase(userConfig.GetDatabasePath())

	repo, err := repository.NewSqlRepository(db)
	if err != nil {
		log.Fatalf("Failed to open database: %s", err.Error())
	}

	handle := command.NewHandlerBuilder(repo).Build()

	setupSignalHandlers()

	app := cli.App{
		Name:        "sf",
		Usage:       "A command-line journal for bite sized thoughts",
		Description: "short-form allows you to write, tag and search for short notes via the command line.",
		Version:     appVersion,
		Flags: []cli.Flag{
			&cli.BoolFlag{
				Name:    "pretty",
				Aliases: []string{"p"},
				Value:   false,
			},
		},
		Commands: []*cli.Command{
			{
				Name:    "write",
				Aliases: []string{"w"},
				Usage:   "Write a new note",
				Flags: []cli.Flag{
					tagFlag,
				},
				Action: handle.WriteNote,
			},
			{
				Name:    "delete",
				Aliases: []string{"d"},
				Usage:   "Delete a note",
				Flags: []cli.Flag{
					&cli.BoolFlag{
						Name:  "no-confirm",
						Usage: "Don't prompt for confirmation",
						Value: false,
					},
				},
				Action: handle.DeleteNote,
			},
			{
				Name:    "edit",
				Aliases: []string{"e"},
				Usage:   "Edit a note's content",
				Action:  handle.EditNote,
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
						Flags:   searchFlags,
						Action:  handle.SearchToday,
					},

					// Search against yesterday's notes.
					{
						Name:    "yesterday",
						Usage:   "Search for notes written yesterday",
						Aliases: []string{"y"},
						Flags:   searchFlags,
						Action:  handle.SearchYesterday,
					},
				},
				Flags:  searchFlags,
				Action: handle.SearchNotes,
			},
			{
				Name:    "configure",
				Usage: "Configure short-form",
				Aliases: []string{"c"},
				Subcommands: []*cli.Command{
					{
						Name:    "read",
						Aliases: []string{"r"},
						Usage: "Display current configure",
						Action: func(ctx *cli.Context) error {
							if pretty, err := json.MarshalIndent(userConfig, "", "	"); err != nil {
								return err
							} else {
								fmt.Println(string(pretty))
								return nil
							}
						},
					},
					{
						Name:    "database",
						Usage: "Configure database properties",
						Aliases: []string{"d"},
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:        "path",
								Aliases:     []string{"p"},
								Usage:       "Path to database file",
								Required:    true,
								Value:       "",
								DefaultText: "",
							},
						},
						Action: func(ctx *cli.Context) error {
							return handle.ConfigureDatabase(ctx, userConfig)
						},
					},
				},
			},
		},
	}

	if err = app.Run(os.Args); err != nil {
		dd(err.Error())
	}
}
