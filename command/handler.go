package command

import (
	"errors"
	"fmt"
	"github.com/ricanontherun/short-form/conf"
	"github.com/ricanontherun/short-form/models"
	"github.com/ricanontherun/short-form/output"
	"github.com/ricanontherun/short-form/repository"
	uuid "github.com/satori/go.uuid"
	"github.com/urfave/cli/v2"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type nowSupplier func() time.Time

type handler struct {
	repository      repository.Repository
	nowSupplyingFn  nowSupplier
	inputController UserInputController
	printer         output.Printer
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
		handler.nowSupplyingFn = builder.nowSupplier
	} else {
		handler.nowSupplyingFn = DefaultNowSupplier
	}

	if builder.inputController != nil {
		handler.inputController = builder.inputController
	} else {
		handler.inputController = NewUserInputController()
	}

	handler.printer = output.NewPrinter()

	return handler
}

func DefaultNowSupplier() time.Time {
	return time.Now()
}

func (handler handler) WriteNote(ctx *cli.Context) error {
	input, err := getContentFromInput(ctx)

	if err != nil {
		return err
	}

	if len(input.content) == 0 {
		return errEmptyContent
	}

	if err := handler.repository.WriteNote(models.NewNote(input.tags, input.content)); err != nil {
		return err
	}

	fmt.Println("note saved")
	return nil
}

func (handler handler) writeNote(note models.Note) error {
	return handler.repository.WriteNote(note)
}

func (handler handler) SearchToday(ctx *cli.Context) error {
	now := handler.nowSupplyingFn()

	searchFilters := getSearchFiltersFromContext(ctx)
	dateRange := models.GetRangeToday(now)
	searchFilters.DateRange = &dateRange

	if notes, err := handler.repository.SearchNotes(searchFilters); err != nil {
		return err
	} else {
		handler.printer.PrintNotes(notes, getPrintOptionsFromContext(ctx))
	}

	return nil
}

func (handler handler) SearchYesterday(ctx *cli.Context) error {
	baseFilters := getSearchFiltersFromContext(ctx)

	dateRange := models.GetRangeYesterday(handler.nowSupplyingFn())
	baseFilters.DateRange = &dateRange

	if notes, err := handler.repository.SearchNotes(baseFilters); err != nil {
		return err
	} else {
		handler.printer.PrintNotes(notes, getPrintOptionsFromContext(ctx))
	}

	return nil
}

func (handler handler) SearchNotes(ctx *cli.Context) error {
	searchFilters := getSearchFiltersFromContext(ctx)

	age := strings.ToLower(ctx.String(flagAge))
	if len(age) > 0 {
		validAge := regexp.MustCompile(`^\d+d$`)
		if !validAge.MatchString(age) {
			return errInvalidAge
		} else {
			ageDays, _ := strconv.Atoi(strings.TrimRight(age, "d"))
			end := handler.nowSupplyingFn()
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
		handler.printer.PrintNotes(notes, getPrintOptionsFromContext(ctx))
	}

	return nil
}

func (handler handler) DeleteNote(ctx *cli.Context) error {
	noteId := strings.TrimSpace(ctx.Args().First())
	if len(noteId) <= 0 {
		return errMissingNoteId
	}

	if _, err := uuid.FromString(noteId); err != nil {
		return errInvalidNoteId
	}

	if _, err := handler.repository.LookupNote(noteId); err != nil {
		return err
	}

	if !ctx.Bool(flagNoConfirm) {
		if ok := handler.makeUserConfirmAction("This will delete 1 note, are you sure?"); !ok {
			fmt.Println("cancelled")
			return nil
		}
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
		return errMissingNoteId
	}

	if _, err := uuid.FromString(noteId); err != nil {
		return errInvalidNoteId
	}

	note, err := handler.repository.LookupNoteWithTags(noteId)
	if err != nil {
		if err == repository.ErrNoteNotFound {
			return errNoteNotFound
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
	handler.printer.PrintNote(note, getPrintOptionsFromContext(ctx))

	return nil
}

func (handler handler) ConfigureDatabase(cli *cli.Context, conf conf.Config) error {
	path := cli.String("path")
	if len(path) == 0 {
		return errors.New("invalid path, empty")
	}

	path, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	if err := conf.SetDatabasePath(path); err != nil {
		return err
	}

	return conf.Save()
}
