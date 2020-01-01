package main

import (
	"fmt"
	"github.com/ricanontherun/short-form/command"
	"github.com/ricanontherun/short-form/conf"
	"github.com/ricanontherun/short-form/database"
	"github.com/ricanontherun/short-form/repository"
	"github.com/ricanontherun/short-form/utils"
	"github.com/urfave/cli/v2"
	"log"
	"os"
)

var (
	tagFlag = &cli.StringFlag{
		Name:    "tags",
		Aliases: []string{"t"},
		Usage:   "comma,separated,list of tags to filter on.",
		Value:   "",
	}
)

func dd(message string) {
	fmt.Println(message)
	os.Exit(1)
}

func startupError(err error) {
	dd("Failed to start sf: " + err.Error())
}

func main() {
	if err := utils.EnsureFilePath(conf.ResolveDatabaseFilePath()); err != nil {
		startupError(err)
	}

	databaseConnection, err := database.NewDatabaseConnection()
	if err != nil {
		startupError(err)
	}
	defer databaseConnection.Close()

	repo, err := repository.NewSqlRepository(databaseConnection)
	if err != nil {
		log.Fatalf("Failed to open database: %s", err.Error())
	}

	handle := command.NewHandlerBuilder(repo).Build()

	app := cli.App{
		Name:        "sf",
		Usage:       "A command-line journal for bite sized thoughts",
		Description: "sf is a command-line journal for simple note writing",
		Version:     "1.0.0",
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
						Action:  handle.SearchToday,
					},

					// Search against yesterday's notes.
					{
						Name:    "yesterday",
						Usage:   "Search for notes written yesterday",
						Aliases: []string{"y"},
						Action:  handle.SearchYesterday,
					},
				},
				Flags: []cli.Flag{
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
				},
				Action: handle.SearchNotes,
			},
		},
	}

	if err = app.Run(os.Args); err != nil {
		dd(err.Error())
	}
}
