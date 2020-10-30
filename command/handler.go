package command

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/ricanontherun/short-form/conf"
	"github.com/ricanontherun/short-form/logging"
	"github.com/ricanontherun/short-form/models"
	"github.com/ricanontherun/short-form/output"
	"github.com/ricanontherun/short-form/repository"
	uuid "github.com/satori/go.uuid"
	"github.com/urfave/cli/v2"
	"log"
	"os"
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
	noteIdLen := len(noteId)

	if noteIdLen == 0 {
		return errMissingNoteId
	}

	if noteIdLen != 8 && noteIdLen != len(uuid.NamespaceDNS.String()) {
		return errInvalidNoteId
	}

	if noteIdLen == 8 { // short id
		notes, err := handler.repository.LookupNotesByShortId(noteId)
		if err != nil {
			return fmt.Errorf("failed to find notes beginning with '%s': %s", noteId, err.Error())
		}
		numNotes := len(notes)

		switch numNotes {
		case 0:
			return errNoteNotFound
		case 1: // Normal case
			return handler.deleteNote(notes[0], !ctx.Bool(flagNoConfirm))
		default: // short ID collision
			fmt.Printf("%d notes were found starting with '%s', please try again using the full ID\n", numNotes, noteId)
			printOptions := output.NewOptions()
			printOptions.FullID = true
			printOptions.Search = struct {
				PrintSummary bool
			}{PrintSummary: false}
			handler.printer.PrintNotes(notes, printOptions)
		}
	} else if noteIdLen == len(uuid.NamespaceDNS.String()) { // full id
		note, err := handler.repository.LookupNote(noteId)
		if err != nil {
			return err
		}
		return handler.deleteNote(note, !ctx.Bool(flagNoConfirm))
	}

	return nil
}

func (handler handler) deleteNote(note *models.Note, confirm bool) error {
	logging.Debug("deleting note using full id " + note.ID)

	if confirm {
		fmt.Println("The following note will be deleted:")
		fmt.Println()

		handler.printer.PrintNote(note, output.Options{})

		if ok := handler.confirmAction("Proceed?"); !ok {
			fmt.Println("canceled")
			return nil
		}
	}

	err := handler.repository.DeleteNote(note.ID)
	if err == nil {
		fmt.Println("note deleted")
	}

	return err
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

	if handler.confirmAction("update content?") {
		newContent := handler.promptUser("new content: ")

		if len(newContent) != 0 {
			contentChanged = true
			note.Content = newContent
		}
	}

	if handler.confirmAction("update tags?") {
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

func (handler handler) StreamNotes(cli *cli.Context) error {
	tags := getTagsFromContext(cli)
	input := ""
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("streaming notes. Separated by newlines, terminated by EOL")

	for {
		fmt.Print("-> ")
		input, _ = reader.ReadString('\n')
		trimmedInput := strings.Replace(strings.TrimSpace(input), "\n", "", -1)

		if len(trimmedInput) == 0 {
			break
		}

		note := models.NewNote(tags, trimmedInput)
		if err := handler.repository.WriteNote(note); err != nil {
			log.Println("failed to save note: " + err.Error())
		}
	}

	return nil
}
