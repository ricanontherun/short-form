package main

import (
	"fmt"
	"github.com/ricanontherun/short-form/data"
	"github.com/ricanontherun/short-form/utils"
	"github.com/urfave/cli"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Return a cleaned array of tags provided as --tags=t1,t2,t3, as ['t1', 't2', 't3']
func getTagsFromContext(c *cli.Context) []string {
	var tags []string

	for _, tag := range strings.Split(c.String("tags"), ",") {
		trimmed := strings.TrimSpace(tag)

		if len(trimmed) > 0 {
			tags = append(tags, strings.ToLower(trimmed))
		}
	}

	return tags
}

// Print notes, sorted by timestamp.
func printNotes(notes map[string]*data.Note) {
	if len(notes) <= 0 {
		return
	}

	// Sort by date
	dates := make([]string, 0, len(notes))
	notesByTimestamp := make(map[string]*data.Note)

	for _, note := range notes {
		notesByTimestamp[note.Timestamp] = note
		dates = append(dates, note.Timestamp)
	}
	sort.Strings(dates)

	// Print
	for _, date := range dates {
		printNote(notesByTimestamp[date])
	}
}

func printNote(note *data.Note) {
	num, err := strconv.ParseInt(note.Timestamp, 10, 64)
	if err != nil {
		return
	}

	topLine := time.Unix(num, 0).Format("January 02, 2006 3:04 PM")
	//topLine += ", ID: " + note.ID

	if len(note.Tags) > 0 {
		topLine += ", tags: " + strings.Join(note.Tags, ", ")
	}

	fmt.Println(topLine)
	fmt.Println(note.Content)
	fmt.Println()
}

func main() {
	repository, err := data.NewRepository()
	if err != nil {
		log.Fatalf("Failed to open database: %s", err.Error())
	}
	defer repository.Close()

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
							searchFilters := data.Filters{
								DateRange: &data.DateRange{
									From: time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()),
									To:   now,
								},
								Tags: getTagsFromContext(c),
							}

							if notes, err := repository.SearchNotes(searchFilters); err != nil {
								return err
							} else {
								printNotes(notes)
							}

							return nil
						},
					},

					// Search against yesterday's notes.
					{
						Name:    "yesterday",
						Aliases: []string{"y"},
						Action: func(c *cli.Context) error {
							yesterday := time.Now().AddDate(0, 0, -1)
							yesterdayStart := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, yesterday.Location())
							yesterdayEnding := time.Date(yesterdayStart.Year(), yesterdayStart.Month(), yesterdayStart.Day(), 23, 59, 59, 0, yesterday.Location())

							searchFilters := data.Filters{
								DateRange: &data.DateRange{
									From: yesterdayStart,
									To:   yesterdayEnding,
								},
								Tags: getTagsFromContext(c),
							}

							if notes, err := repository.SearchNotes(searchFilters); err != nil {
								return err
							} else {
								printNotes(notes)
							}

							return nil
						},
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
