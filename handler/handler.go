package handler

import (
	"errors"
	"fmt"
	"github.com/ricanontherun/short-form/data"
	"github.com/ricanontherun/short-form/utils"
	uuid "github.com/satori/go.uuid"
	"github.com/urfave/cli/v2"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	ErrEmptyContent    = errors.New("empty content")
	ErrMissingNoteId   = errors.New("missing note id")
	ErrMalformedNoteId = errors.New("invalid note id")
	ErrNoteNotFound    = errors.New("note not found")
)

type handler struct {
	repository data.Repository
}

// Create a new handler.
// A handler serves as the entry point to the application, fulfilling user commands.
func NewHandler(repository data.Repository) handler {
	return handler{repository}
}

func (handler handler) WriteNote(ctx *cli.Context) error {
	input := getInputFromContext(ctx)

	if len(input.content) == 0 {
		return ErrEmptyContent
	}

	note := data.NewNote(input.tags, input.content)
	if err := handler.repository.WriteNote(note); err != nil {
		return err
	}

	fmt.Println(note.ID)

	return nil
}

func (handler handler) writeNote(note data.Note) error {
	return handler.repository.WriteNote(note)
}

func (handler handler) SearchToday(ctx *cli.Context) error {
	now := time.Now()

	searchFilters := getSearchFiltersFromContext(ctx)
	dateRange := utils.GetRangeToday(now)
	searchFilters.DateRange = &dateRange

	if notes, err := handler.repository.SearchNotes(searchFilters); err != nil {
		return err
	} else {
		handler.printNotes(notes, getPrintOptionsFromContext(ctx))
	}

	return nil
}

func (handler handler) SearchYesterday(ctx *cli.Context) error {
	baseFilters := getSearchFiltersFromContext(ctx)

	dateRange := utils.GetRangeYesterday(time.Now())
	baseFilters.DateRange = &dateRange

	if notes, err := handler.repository.SearchNotes(baseFilters); err != nil {
		return err
	} else {
		handler.printNotes(notes, getPrintOptionsFromContext(ctx))
	}

	return nil
}

func (handler handler) SearchNotes(ctx *cli.Context) error {
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

			searchFilters.DateRange = &utils.DateRange{
				From: start,
				To:   end,
			}
		}
	}

	if notes, err := handler.repository.SearchNotes(searchFilters); err != nil {
		return err
	} else {
		handler.printNotes(notes, getPrintOptionsFromContext(ctx))
	}

	return nil
}

func (handler handler) DeleteNote(ctx *cli.Context) error {
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

	// Make sure the note exists.
	if _, err := handler.repository.GetNote(noteId); err != nil {
		return err
	}

	// Prompt the user for confirmation.
	if ok := makeUserConfirmAction("This will delete 1 note, are you sure?"); !ok {
		fmt.Println("cancelled")
		return nil
	}

	if err := handler.repository.DeleteNote(noteId); err != nil {
		return err
	} else {
		fmt.Println("ok")
	}

	return nil
}

func (handler handler) EditNote(ctx *cli.Context) error {
	// Get the noteId from context.
	noteId := ctx.Args().First()

	if len(noteId) == 0 {
		return ErrMissingNoteId
	}

	if _, err := uuid.FromString(noteId); err != nil {
		return ErrMalformedNoteId
	}

	note, err := handler.repository.GetNote(noteId)
	if err != nil {
		if err == data.ErrNoteNotFound {
			return ErrNoteNotFound
		}

		return err
	}

	changed := false
	tagsChanged := false

	newContent := promptUser("New Content: ")
	if len(newContent) != 0 {
		changed = true
		note.Content = newContent
	}

	newTagsString := promptUser("New Tags: ")
	if len(newTagsString) != 0 {
		tagsChanged = true
		note.Tags = cleanTagsFromString(newTagsString)
	}

	if changed {
		if err := handler.repository.UpdateNote(*note); err != nil {
			return err
		}
	}

	if tagsChanged {
		if err := handler.repository.TagNote(*note, note.Tags); err != nil {
			return err
		}
	}

	return nil
}
