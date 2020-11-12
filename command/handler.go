package command

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/ricanontherun/short-form/config"
	"github.com/ricanontherun/short-form/logging"
	"github.com/ricanontherun/short-form/models"
	"github.com/ricanontherun/short-form/output"
	"github.com/ricanontherun/short-form/repository"
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
	userConfig      config.Config
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
	// Check for delete by tagString scenario.
	tagString := strings.ToLower(strings.TrimSpace(ctx.String("tags")))

	if len(tagString) != 0 { // Delete by tagString
		tags := strings.Split(tagString, ",")

		if notes, err := handler.repository.SearchNotes(models.SearchFilters{
			Tags: tags,
		}); err != nil {
			return err
		} else if len(notes) == 0 {
			fmt.Printf("0 notes found with tags '%s'", tagString)
			return nil
		} else {
			fmt.Println("The following notes will be deleted:")
			handler.printer.PrintNotes(notes, output.Options{})

			confirm := ""
			notesLen := len(notes)
			if len(notes) == 1 {
				confirm = "delete 1 note?"
			} else {
				confirm = fmt.Sprintf("delete %d notes?", notesLen)
			}

			if handler.confirmAction(confirm) {
				if _, exists := os.LookupEnv("SHORT_FORM_DRYRUN"); !exists {
					for _, tag := range tags {
						if deletedErr := handler.repository.DeleteNoteByTag(tag); deletedErr != nil {
							return deletedErr
						}
					}

					fmt.Println("ok")
				} else {
					fmt.Println("dry run, 0 notes deleted")
				}

				return nil
			}
		}

		return nil
	} else { // delete by id
		noteId := strings.TrimSpace(ctx.Args().First())
		noteIdLen := len(noteId)

		if noteIdLen == 0 {
			return errMissingNoteId
		}

		if !models.IsValidId(noteId) {
			return errInvalidNoteId
		}

		if note, err := handler.findNoteById(noteId); err != nil {
			return err
		} else {
			return handler.deleteNote(note, !ctx.Bool(flagNoConfirm))
		}
	}
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
	noteIdLen := len(noteId)

	if noteIdLen == 0 {
		return errMissingNoteId
	}

	if !models.IsValidId(noteId) {
		return errInvalidNoteId
	}

	var note *models.Note
	var err error
	if note, err = handler.findNoteById(noteId); err != nil {
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

func (handler handler) ConfigureDatabase(cli *cli.Context, conf config.Config) error {
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

// Attempt to find a note by it's ID, automatically resolving short vs long IDs.
func (handler *handler) findNoteById(noteId string) (*models.Note, error) {
	var note *models.Note
	var noteIdLen = len(noteId)

	if noteIdLen == models.ShortIdLength {
		if notes, err := handler.repository.LookupNotesByShortId(noteId); err != nil {
			return nil, err
		} else {
			var notesLen = len(notes)

			switch notesLen {
			case 0:
				return nil, errNoteNotFound
			case 1:
				note = notes[0]
			default:
				fmt.Printf("The following notes were found having IDs starting with '%s', please try again using the full ID\n", noteId)
				printOptions := output.NewOptions()
				printOptions.FullID = true
				printOptions.Search = struct {
					PrintSummary bool
				}{PrintSummary: false}
				handler.printer.PrintNotes(notes, printOptions)
				return nil, errShortIdCollision
			}
		}
	} else {
		var err error
		note, err = handler.repository.LookupNoteWithTags(noteId)
		if err != nil {
			return nil, err
		}
	}

	return note, nil
}
