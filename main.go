package main

import (
	"fmt"
	"github.com/urfave/cli"
	"log"
	"os"
	"short-form/data"
	"short-form/search"
	"short-form/utils"
	"strconv"
	"strings"
	"time"
)

func getTagsFromContext(c *cli.Context) []string {
	var tags []string

	for _, tag := range strings.Split(c.String("tags"), ",") {
		tags = append(tags, strings.TrimSpace(tag))
	}

	return tags
}

func printNote(note data.Note) {
	num, err := strconv.ParseInt(note.Timestamp, 10, 64)
	if err != nil {
		return
	}
	prettyDate := time.Unix(num, 0)
	fmt.Println("ID", note.ID)
	fmt.Println("This is the next content")
	fmt.Println(prettyDate, "health, test")
}

func main() {
	repository, err := data.NewRepository()
	if err != nil {
		log.Fatalf("Failed to open database: %s", err.Error())
	}
	defer repository.Close()

	app := cli.App{
		Commands: []*cli.Command{
			{
				Name:    "configure",
				Aliases: []string{"c"},
				Usage:   "configure",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "path",
						Usage:       "--path",
						Value:       "",
						DefaultText: "uh",
						Aliases:     []string{"p"},
					},
				},
			},
			{
				Name:    "write",
				Aliases: []string{"w"},
				Usage:   "write",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "tags",
						Usage:       "--tags music,health",
						Value:       "",
						DefaultText: "random",
						Aliases:     []string{"t"},
					},
				},
				Action: func(c *cli.Context) error {
					note := data.Note{
						ID:        utils.MakeUUID(),
						Timestamp: utils.CurrentUnixTimestamp(),
						Tags:      getTagsFromContext(c),
						Content:   strings.Join(c.Args().Slice(), " "),
					}

					if err := repository.WriteNote(note); err == nil {
						log.Println(note.ID)
						return nil
					} else {
						return err
					}
				},
			},
			{
				Name:    "search",
				Aliases: []string{"s"},
				Subcommands: []*cli.Command{
					// Search against today's notes.
					{
						Name:    "today",
						Aliases: []string{"t"},
						Action: func(c *cli.Context) error {
							now := time.Now()
							searchFilters := search.Filters{
								DateRange: &search.DateRange{
									From: time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()),
								},
								Tags: getTagsFromContext(c),
							}

							if notes, err := repository.SearchNotes(searchFilters); err != nil {
								return err
							} else {
								for _, note := range notes {
									printNote(note)
								}
							}

							return nil
						},
					},

					// Search against yesterday's notes.
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
					searchContext := data.SearchContext{}

					sinceString := c.String("since")
					if len(sinceString) > 0 {
						sinceString = strings.Replace(sinceString, "-", "", 1)

						if since, err := time.ParseDuration("-" + sinceString); err != nil {
							// Parse manually.
							log.Fatalf("Failed to parse --since of '%s', %s", sinceString, err.Error())
						} else {
							searchContext.From = time.Now().Add(since)
						}
					}

					fmt.Println(searchContext.From.String())
					fmt.Println(getTagsFromContext(c))
					return nil
				},
			},
		},
	}

	if err = app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
