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
	ErrEmptyContent       = errors.New("empty content")
	ErrMissingNoteId      = errors.New("missing note id")
	ErrMalformedNoteId    = errors.New("malformed note id")
	ErrMalformedHighlight = errors.New("malformed highlight")
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

type printOptions struct {
	insecure  bool
	highlight string
}

type highlight struct {
	left      string
	highlight string
	right     string
}

func getPrintOptionsFromContext(ctx *cli.Context) printOptions {
	return printOptions{
		insecure:  ctx.Bool("insecure"),
		highlight: ctx.String("content"),
	}
}

func getHighlight(original string, hl string) []highlight {
	var highlights []highlight

	cursor := original
	for cursor != "" {
		startIndex := strings.Index(cursor, hl)
		if startIndex == -1 {
			break
		}

		cursorBytes := []byte(cursor)
		highlight := highlight{
			left:      string(cursorBytes[0:startIndex]),
			highlight: string(cursorBytes[startIndex : startIndex+len(hl)]),
			right:     "",
		}

		cursor = strings.Replace(cursor, highlight.left+highlight.highlight, "", 1)

		// If we're on the the last occurrence, keep the tail.
		nextIndex := strings.Index(cursor, hl)
		if nextIndex == -1 {
			highlight.right = string(cursorBytes[startIndex+len(hl):])
		}

		highlights = append(highlights, highlight)
	}

	return highlights
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

func (handler Handler) printNotes(notes []data.Note, options printOptions) {
	noteCount := len(notes)
	noteCountString := ""

	if noteCount == 1 {
		noteCountString = fmt.Sprintf("1 note found")
	} else {
		noteCountString = fmt.Sprintf("%d notes found", noteCount)
	}

	fmt.Println(noteCountString)

	if len(notes) <= 0 {
		return
	}

	fmt.Println()

	for _, note := range notes {
		handler.printNote(note, options)
	}
}

func (handler Handler) printNote(note data.Note, options printOptions) {
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

	contentString := ""

	if note.Secure {
		if options.insecure {
			if insecureContent, err := handler.encryptor.Decrypt([]byte(note.Content)); err != nil {
				log.Fatalln(err)
			} else {
				contentString = string(insecureContent)
			}
		} else {
			contentString = "*****************"
		}
	} else {
		contentString = note.Content
	}

	if options.highlight != "" && !options.insecure {
		highlights := getHighlight(note.Content, options.highlight)

		if len(highlights) > 0 {
			contentString = ""
			colorPrinter := color.New(color.Bold)
			for _, hl := range getHighlight(note.Content, options.highlight) {
				contentString += hl.left + colorPrinter.Sprint(hl.highlight) + hl.right
			}
		}
	}

	fmt.Println(contentString)
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
		handler.printNotes(notes, getPrintOptionsFromContext(ctx))
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
		handler.printNotes(notes, getPrintOptionsFromContext(ctx))
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
		handler.printNotes(notes, getPrintOptionsFromContext(ctx))
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

	result := handler.Repository.DeleteNote(noteId)
	if result == nil {
		fmt.Println("ok")
	} else if result == data.ErrNoteNotFound {
		fmt.Println("not found")
	} else {
		return result
	}

	return nil
}
