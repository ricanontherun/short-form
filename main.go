package main

import (
	"encoding/json"
	"fmt"
	"github.com/ricanontherun/short-form/command"
	"github.com/ricanontherun/short-form/config"
	"github.com/ricanontherun/short-form/database"
	"github.com/ricanontherun/short-form/logging"
	"github.com/ricanontherun/short-form/repository"
	"github.com/urfave/cli/v2"
	"log"
	"os"
	"os/signal"
	"os/user"
	"syscall"
)

var (
	tagFlag = &cli.StringFlag{
		Name:    "tag",
		Aliases: []string{"t"},
		Usage:   "single tag",
		Value:   "",
	}

	tagsFlag = &cli.StringFlag{
		Name:    "tags",
		Aliases: []string{"t"},
		Usage:   "comma,separated,list of tags to filter on.",
		Value:   "",
	}

	confirmFlag = &cli.BoolFlag{
		Name:    "no-confirm",
		Aliases: []string{"n"},
		Usage:   "Don't prompt for confirmation",
		Value:   false,
	}

	appVersion = "3.0.0"
)

var searchFlags = []cli.Flag{
	tagsFlag,
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
}

func dd(message string) {
	fmt.Println(message)
	os.Exit(1)
}

// Setup support for ctrl-c interrupt signals which are ignored whilst
// waiting for user input by default.
func setupSignalHandlers() {
	signalChan := make(chan os.Signal, 2)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-signalChan
		os.Exit(1)
	}()
}

func main() {
	systemUser, systemUserErr := user.Current()
	if systemUserErr != nil {
		log.Fatalf("failed to read user's home directory: %s\n", systemUserErr.Error())
	}

	userConfig, readConfigErr := config.ReadUserConfig(systemUser.HomeDir)
	if readConfigErr != nil {
		log.Fatalln(readConfigErr)
	}

	db := database.NewDatabase(userConfig.GetDatabasePath())

	repo, err := repository.NewSqlRepository(db)
	if err != nil {
		log.Fatalf("Failed to open database: %s\n", err.Error())
	}

	handler := command.NewHandlerBuilder(repo).Build()

	setupSignalHandlers()

	app := cli.App{
		Name:        "sf",
		Usage:       "A command-line journal for bite sized thoughts",
		Description: "short-form allows you to write, tag and search for short notes via the command line.",
		Version:     appVersion,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "database-path",
				Usage:   "/path/to/database to use for this command",
				Aliases: []string{"d"},
			},
		},
		Before: func(context *cli.Context) error {
			inlineDatabasePath := context.String("database-path")
			if len(inlineDatabasePath) != 0 {
				if repo, repoErr := repository.NewSqlRepository(database.NewDatabase(inlineDatabasePath)); repoErr != nil {
					return repoErr
				} else {
					logging.Debug(fmt.Sprintf("--database-path provided, overriding with %s", inlineDatabasePath))
					handler = command.NewHandlerBuilder(repo).Build()
				}
			}

			return nil
		},
		Commands: []*cli.Command{
			{
				Name:    "write",
				Aliases: []string{"w"},
				Usage:   "Write a new note",
				Flags: []cli.Flag{
					tagsFlag,
					confirmFlag,
				},
				Action: func(context *cli.Context) error {
					return handler.WriteNote(context)
				},
			},
			{
				Name:    "delete",
				Aliases: []string{"d"},
				Usage:   "Delete a note",
				Flags:   []cli.Flag{tagsFlag},
				Action: func(context *cli.Context) error {
					return handler.DeleteNote(context)
				},
			},
			{
				Name:    "edit",
				Aliases: []string{"e"},
				Usage:   "Edit a note's content",
				Action: func(context *cli.Context) error {
					return handler.EditNote(context)
				},
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
						Action: func(context *cli.Context) error {
							return handler.SearchToday(context)
						},
					},

					// Search against yesterday's notes.
					{
						Name:    "yesterday",
						Usage:   "Search for notes written yesterday",
						Aliases: []string{"y"},
						Flags:   searchFlags,
						Action: func(context *cli.Context) error {
							return handler.SearchYesterday(context)
						},
					},
				},
				Flags: searchFlags,
				Action: func(ctx *cli.Context) error {
					return handler.SearchNotes(ctx)
				},
			},
			{
				Name:    "configure",
				Usage:   "Configure short-form",
				Aliases: []string{"c"},
				Subcommands: []*cli.Command{
					{
						Name:    "read",
						Aliases: []string{"r"},
						Usage:   "Display current configure",
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
						Usage:   "Configure database properties",
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
							return handler.ConfigureDatabase(ctx, userConfig)
						},
					},
				},
			},
			{
				Name:    "stream",
				Usage:   "Stream notes",
				Aliases: []string{"st"},
				Action: func(context *cli.Context) error {
					return handler.StreamNotes(context)
				},
				Flags: []cli.Flag{
					tagsFlag,
				},
			},
		},
	}

	if err = app.Run(os.Args); err != nil {
		dd(err.Error())
	}
}
