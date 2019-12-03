package handler

import (
	"errors"
	"fmt"
	"github.com/ricanontherun/short-form/data"
	"github.com/ricanontherun/short-form/utils"
	"github.com/urfave/cli/v2"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	ErrEmptyContent = errors.New("empty content")
)

type Handler struct {
	Repository data.Repository
}

// Return a cleaned array of tags provided as --tags=t1,t2,t3, as ['t1', 't2', 't3']
func getTagsFromContext(c *cli.Context) []string {
	tags := utils.NewSet()

	for _, tag := range strings.Split(c.String("tags"), ",") {
		trimmed := strings.TrimSpace(tag)

		if len(trimmed) > 0 {
			tags.Add(strings.ToLower(trimmed))
		}
	}

	return tags.Entries()
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

	fmt.Println()
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
	topLine += " | " + note.ID
	if note.Secure {
		topLine += " | secure"
	} else {
		topLine += " | insecure"
	}

	if len(note.Tags) > 0 {
		topLine += " | " + strings.Join(note.Tags, ", ")
	}

	fmt.Println(topLine)
	fmt.Println(note.Content)
	fmt.Println()
}

func (handler Handler) WriteNote(ctx *cli.Context) error {
	note := data.Note{
		ID:        utils.MakeUUID(),
		Timestamp: utils.CurrentUnixTimestamp(),
		Tags:      getTagsFromContext(ctx),
		Content:   strings.Join(ctx.Args().Slice(), " "),
	}

	if len(note.Content) <= 0 {
		return ErrEmptyContent
	}

	if err := handler.Repository.WriteNote(note, false); err == nil {
		fmt.Println(note.ID)
		return nil
	} else {
		return err
	}
}

func (handler Handler) WriteSecureNote(ctx *cli.Context) error {
	note := data.Note{
		ID:        utils.MakeUUID(),
		Timestamp: utils.CurrentUnixTimestamp(),
		Tags:      getTagsFromContext(ctx),
		Content:   strings.Join(ctx.Args().Slice(), " "),
	}

	if len(note.Content) <= 0 {
		return ErrEmptyContent
	}

	if err := handler.Repository.WriteNote(note, true); err == nil {
		fmt.Println(note.ID)
		return nil
	} else {
		return err
	}
}

func (handler Handler) SearchTodayNote(ctx *cli.Context) error {
	now := time.Now()
	searchFilters := data.Filters{
		DateRange: &data.DateRange{
			From: time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()),
			To:   now,
		},
		Tags: getTagsFromContext(ctx),
	}

	if notes, err := handler.Repository.SearchNotes(searchFilters); err != nil {
		return err
	} else {
		printNotes(notes)
	}

	return nil
}

func (handler Handler) SearchYesterdayNote(ctx *cli.Context) error {
	yesterday := time.Now().AddDate(0, 0, -1)
	yesterdayStart := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, yesterday.Location())
	yesterdayEnding := time.Date(yesterdayStart.Year(), yesterdayStart.Month(), yesterdayStart.Day(), 23, 59, 59, 0, yesterday.Location())

	searchFilters := data.Filters{
		DateRange: &data.DateRange{
			From: yesterdayStart,
			To:   yesterdayEnding,
		},
		Tags: getTagsFromContext(ctx),
	}

	if notes, err := handler.Repository.SearchNotes(searchFilters); err != nil {
		return err
	} else {
		printNotes(notes)
	}

	return nil
}

func (handler Handler) SearchNotes(ctx *cli.Context) error {
	contextTags := getTagsFromContext(ctx)

	if len(contextTags) == 0 {
		return errors.New("invalid search")
	}

	searchFilters := data.Filters{
		Tags: contextTags,
	}

	if notes, err := handler.Repository.SearchNotes(searchFilters); err != nil {
		return err
	} else {
		printNotes(notes)
	}

	return nil
}
