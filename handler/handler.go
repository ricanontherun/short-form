package handler

import (
	"errors"
	"fmt"
	"github.com/fatih/color"
	"github.com/ricanontherun/short-form/data"
	"github.com/ricanontherun/short-form/utils"
	uuid "github.com/satori/go.uuid"
	"github.com/urfave/cli/v2"
	"log"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	ErrEmptyContent    = errors.New("empty content")
	ErrMissingNoteId   = errors.New("missing note id")
	ErrMalformedNoteId = errors.New("malformed note id")
)

type Handler struct {
	Repository data.Repository
	encryptor  utils.Encryptor
}

func NewHandler(repository data.Repository, encryptor utils.Encryptor) Handler {
	return Handler{repository, encryptor}
}

type parsedInput struct {
	content string
	tags    []string
}

func getInputFromContext(ctx *cli.Context) parsedInput {
	return parsedInput{
		content: strings.Join(ctx.Args().Slice(), " "),
		tags:    getTagsFromContext(ctx),
	}
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

func getSearchFiltersFromContext(c *cli.Context) data.Filters {
	return data.Filters{
		Tags:    getTagsFromContext(c),
		Content: strings.ToLower(c.String("content")),
	}
}

func (handler Handler) printNotes(notes []data.Note, insecure bool) {
	fmt.Println(fmt.Sprintf("%d note(s) found", len(notes)))

	if len(notes) <= 0 {
		return
	}

	fmt.Println()

	for _, note := range notes {
		handler.printNote(note, insecure)
	}
}

func (handler Handler) printNote(note data.Note, insecure bool) {
	bits := make([]string, 0, 4)

	bits = append(bits, color.BlueString(note.Timestamp.Format("January 02, 2006 03:04 PM")))
	bits = append(bits, color.MagentaString(note.ID))

	if note.Secure {
		bits = append(bits, color.GreenString("secure"))
	} else {
		bits = append(bits, color.RedString("insecure"))
	}

	if len(note.Tags) > 0 {
		bits = append(bits, strings.Join(note.Tags, ", "))
	}

	fmt.Println(strings.Join(bits, " | "))

	if note.Secure {
		if insecure {
			if insecureContent, err := handler.encryptor.Decrypt([]byte(note.Content)); err != nil {
				log.Fatalln(err)
			} else {
				fmt.Println(string(insecureContent))
			}
		} else {
			fmt.Println("*****************")
		}
	} else {
		fmt.Println(note.Content)
	}

	fmt.Println()
}

func (handler Handler) WriteInsecureNote(ctx *cli.Context) error {
	input := getInputFromContext(ctx)

	note := data.NewInsecureNote(input.tags, input.content)

	if err := handler.writeNote(note); err != nil {
		return err
	}

	fmt.Println(note.ID)

	return nil
}

func (handler Handler) WriteSecureNote(ctx *cli.Context) error {
	input := getInputFromContext(ctx)

	note := data.NewSecureNote(input.tags, input.content)

	if err := handler.writeNote(note); err != nil {
		return err
	}

	fmt.Println(note.ID)

	return nil
}

func (handler Handler) writeNote(note data.Note) error {
	if len(note.Content) <= 0 {
		return ErrEmptyContent
	}

	return handler.Repository.WriteNote(note)
}

func (handler Handler) SearchTodayNote(ctx *cli.Context) error {
	now := time.Now()

	searchFilters := getSearchFiltersFromContext(ctx)
	searchFilters.DateRange = &data.DateRange{
		From: time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()),
		To:   now,
	}

	if notes, err := handler.Repository.SearchNotes(searchFilters); err != nil {
		return err
	} else {
		handler.printNotes(notes, ctx.Bool("insecure"))
	}

	return nil
}

func (handler Handler) SearchYesterdayNote(ctx *cli.Context) error {
	yesterday := time.Now().AddDate(0, 0, -1)
	yesterdayStart := time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, yesterday.Location())
	yesterdayEnding := time.Date(yesterdayStart.Year(), yesterdayStart.Month(), yesterdayStart.Day(), 23, 59, 59, 0, yesterday.Location())

	searchFilters := getSearchFiltersFromContext(ctx)
	searchFilters.DateRange = &data.DateRange{
		From: yesterdayStart,
		To:   yesterdayEnding,
	}

	if notes, err := handler.Repository.SearchNotes(searchFilters); err != nil {
		return err
	} else {
		handler.printNotes(notes, ctx.Bool("insecure"))
	}

	return nil
}

func (handler Handler) SearchNotes(ctx *cli.Context) error {
	searchFilters := getSearchFiltersFromContext(ctx)

	// Check for age.
	age := strings.ToLower(ctx.String("age"))
	if len(age) > 0 {
		validAge := regexp.MustCompile(`^\d+d$`)
		if !validAge.MatchString(age) {
			return errors.New("invalid age: " + age)
		} else {
			ageDays, _ := strconv.Atoi(strings.TrimRight(age, "d"))
			end := time.Now()
			start := end.AddDate(0, 0, -ageDays)

			searchFilters.DateRange = &data.DateRange{
				From: start,
				To:   end,
			}
		}
	}

	if notes, err := handler.Repository.SearchNotes(searchFilters); err != nil {
		return err
	} else {
		handler.printNotes(notes, ctx.Bool("insecure"))
	}

	return nil
}

func (handler Handler) DeleteNote(ctx *cli.Context) error {
	if len(ctx.Args().Slice()) <= 0 {
		return errors.New("no note id provided")
	}

	noteId := strings.TrimSpace(ctx.Args().First())
	if len(noteId) <= 0 {
		return ErrMissingNoteId
	}

	// Validate it's a V4 UUID
	if _, err := uuid.FromString(noteId); err != nil {
		return ErrMalformedNoteId
	}

	return handler.Repository.DeleteNote(noteId)
}
