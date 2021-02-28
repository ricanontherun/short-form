package main

import (
	"encoding/json"
	"fmt"
	"github.com/ricanontherun/short-form/command"
	"github.com/ricanontherun/short-form/config"
	"github.com/ricanontherun/short-form/database"
	"github.com/ricanontherun/short-form/dto"
	"github.com/ricanontherun/short-form/logging"
	"github.com/ricanontherun/short-form/output"
	"github.com/ricanontherun/short-form/query"
	"github.com/ricanontherun/short-form/repository"
	"github.com/ricanontherun/short-form/user_input"
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

	// initialize singleton database instance.
	database.InitializeDatabaseSingleton(userConfig.GetDatabasePath())
	db := database.GetInstance()

	repo := repository.NewSqlRepository(db)
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
				database.InitializeDatabaseSingleton(inlineDatabasePath)
				repo := repository.NewSqlRepository(database.GetInstance())
				logging.Debug(fmt.Sprintf("--database-path provided, overriding with %s", inlineDatabasePath))
				handler = command.NewHandlerBuilder(repo).Build()
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
					if note, err := command.NewNoteDTOFromContext(context); err != nil {
						return err
					} else {
						writeErr := command.NewWriteCommand(note).Execute()
						if writeErr == nil {
							fmt.Println("note saved")
						}
						return writeErr
					}
				},
			},
			{
				Name:    "delete",
				Aliases: []string{"d"},
				Flags:   []cli.Flag{tagsFlag},
				Action: func(ctx *cli.Context) error {
					noteId := user_input.GetNoteIdFromContext(ctx)

					if len(noteId) != 0 {
						logging.Debug("deleting by note ID: " + noteId)
						return deleteByNoteId(noteId)
					} else { // try deleting by tag
						tags := user_input.GetTagsFromContext(ctx)
						if len(tags) == 0 {
							logging.Debug("no tags present in input")
							return cli.ShowAppHelp(ctx)
						}

						// Fetch the number of notes which would be effected by this action.
						// Force the user to confirm.
						if numNotes, err := query.NewTagsCountQuery(tags).Run(); err != nil {
							return err
						} else {
							numNotesInt := numNotes.(uint64)
							logging.Debug(fmt.Sprintf("found %d notes", numNotesInt))

							switch numNotesInt {
							case 0:
								fmt.Println("no notes found")
								return nil
							default:
								// prompt the user to make a decision.
								message := ""
								if numNotesInt == 1 {
									message = "1 note found, continue?"
								} else {
									message = fmt.Sprintf("%d notes found, continue?", numNotesInt)
								}

								inputController := command.NewUserInputController()
								if inputController.ConfirmAction(message) {
									logging.Debug("deleting by tags")
									if deleteErr := command.NewDeleteByTagsCommand(tags).Execute(); deleteErr != nil {
										return nil
									} else {
										fmt.Println("note(s) deleted")
									}
								} else {
									logging.Debug("delete by tag denied")
								}
							}
						}

						return nil
					}
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

	if err := app.Run(os.Args); err != nil {
		dd(err.Error())
	}
}

func deleteByNoteId(noteId string) error {
	if !dto.IsValidId(noteId) {
		return fmt.Errorf("'%s' is not a valid note ID", noteId)
	} else { // try perform a delete by note ID command.
		deleteByNoteIdErr := command.NewDeleteCommand(&command.DeleteByNoteID{NoteID: noteId}).Execute()
		if deleteByNoteIdErr == nil {
			fmt.Println("note(s) deleted")
			return nil
		} else {
			switch deleteByNoteIdErr.(type) {
			case *command.NoteCollisionError:
				// Short ID collision.
				// Display relevant notes and force user to delete by full UUID.
				output.PrintNoteSummary(deleteByNoteIdErr.(*command.NoteCollisionError).GetNotes())
			}

			return deleteByNoteIdErr
		}
	}
}

func deleteByTags(tags []string) error {
	// A SearchByTagsQuery would be useful here.
	return nil
}
