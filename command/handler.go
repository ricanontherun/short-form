package command

import (
	"fmt"
	"github.com/ricanontherun/short-form/models"
	"github.com/ricanontherun/short-form/repository"
	uuid "github.com/satori/go.uuid"
	"github.com/urfave/cli/v2"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type nowSupplier func() time.Time

type handler struct {
	repository      repository.Repository
	nowSupplier     nowSupplier
	inputController UserInputController
}

type HandlerBuilder struct {
	repository      repository.Repository
	nowSupplier     nowSupplier
	inputController UserInputController
}

func NewHandlerBuilder(repository repository.Repository) *HandlerBuilder {
	return &HandlerBuilder{repository: repository}
}

func (builder *HandlerBuilder) WithNowSupplier(supplier nowSupplier) *HandlerBuilder {
	builder.nowSupplier = supplier
	return builder
}

func (builder *HandlerBuilder) WithUserInputController(inputController UserInputController) *HandlerBuilder {
	builder.inputController = inputController
	return builder
}

func (builder *HandlerBuilder) Build() handler {
	handler := handler{}

	handler.repository = builder.repository

	if builder.nowSupplier != nil {
		handler.nowSupplier = builder.nowSupplier
	} else {
		handler.nowSupplier = DefaultNowSupplier
	}

	if builder.inputController != nil {
		handler.inputController = builder.inputController
	} else {
		handler.inputController = NewUserInputController()
	}

	return handler
}

func DefaultNowSupplier() time.Time {
	return time.Now()
}

func (handler handler) WriteNote(ctx *cli.Context) error {
	input := getInputFromContext(ctx)

	if len(input.content) == 0 {
		return ErrEmptyContent
	}

	note := models.NewNote(input.tags, input.content)
	if err := handler.repository.WriteNote(note); err != nil {
		return err
	}

	return nil
}

func (handler handler) writeNote(note models.Note) error {
	return handler.repository.WriteNote(note)
}

func (handler handler) SearchToday(ctx *cli.Context) error {
	now := handler.nowSupplier()

	searchFilters := getSearchFiltersFromContext(ctx)
	dateRange := models.GetRangeToday(now)
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

	dateRange := models.GetRangeYesterday(handler.nowSupplier())
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

	age := strings.ToLower(ctx.String(FlagAge))
	if len(age) > 0 {
		validAge := regexp.MustCompile(`^\d+d$`)
		if !validAge.MatchString(age) {
			return ErrInvalidAge
		} else {
			ageDays, _ := strconv.Atoi(strings.TrimRight(age, "d"))
			end := handler.nowSupplier()
			start := end.AddDate(0, 0, -ageDays)

			searchFilters.DateRange = &models.DateRange{
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
	noteId := strings.TrimSpace(ctx.Args().First())
	if len(noteId) <= 0 {
		return ErrMissingNoteId
	}

	if _, err := uuid.FromString(noteId); err != nil {
		return ErrInvalidNoteId
	}

	if _, err := handler.repository.LookupNote(noteId); err != nil {
		return err
	}

	if ok := handler.makeUserConfirmAction("This will delete 1 note, are you sure?"); !ok {
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
	noteId := ctx.Args().First()

	if len(noteId) == 0 {
		return ErrMissingNoteId
	}

	if _, err := uuid.FromString(noteId); err != nil {
		return ErrInvalidNoteId
	}

	note, err := handler.repository.LookupNoteWithTags(noteId)
	if err != nil {
		if err == repository.ErrNoteNotFound {
			return ErrNoteNotFound
		}

		return err
	}

	contentChanged := false
	tagsChanged := false

	if handler.makeUserConfirmAction("update content?") {
		newContent := handler.promptUser("new content: ")

		if len(newContent) != 0 {
			contentChanged = true
			note.Content = newContent
		}
	}

	fmt.Println()

	if handler.makeUserConfirmAction("update tags?") {
		newTagsString := handler.promptUser("new tags: ")

		if len(newTagsString) != 0 {
			tagsChanged = true
			note.Tags = cleanTagsFromString(newTagsString)
		}
	}

	if contentChanged {
		if err := handler.repository.UpdateNote(*note); err != nil {
			return err
		}
	}

	if tagsChanged {
		if err := handler.repository.TagNote(*note, note.Tags); err != nil {
			return err
		}
	}

	fmt.Println()
	handler.printNote(note, printOptions{})

	return nil
}
