package command

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/ricanontherun/short-form/config"
	"github.com/ricanontherun/short-form/dto"
	"github.com/ricanontherun/short-form/logging"
	"github.com/ricanontherun/short-form/output"
	"github.com/ricanontherun/short-form/repository"
	"github.com/ricanontherun/short-form/user_input"
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

	return handler
}

func DefaultNowSupplier() time.Time {
	return time.Now()
}

func (handler handler) WriteNote(input *NoteDTO) error {
	if err := handler.repository.WriteNote(dto.NewNote(input.Tags, input.Content)); err != nil {
		return err
	}
	return nil
}

func (handler handler) writeNote(note dto.Note) error {
	return handler.repository.WriteNote(note)
}

func (handler handler) SearchToday(ctx *cli.Context) error {
	now := handler.nowSupplyingFn()

	if searchFilters, err := handler.getSearchFiltersFromContext(ctx); err != nil {
		return err
	} else {
		dateRange := dto.GetRangeToday(now)
		searchFilters.DateRange = &dateRange

		if notes, err := handler.repository.SearchNotes(searchFilters); err != nil {
			return err
		} else {
			output.PrintNotes(notes, getPrintOptionsFromContext(ctx))
		}

		return nil
	}
}

func (handler handler) SearchYesterday(ctx *cli.Context) error {
	if baseFilters, err := handler.getSearchFiltersFromContext(ctx); err != nil {
		return err
	} else {
		dateRange := dto.GetRangeYesterday(handler.nowSupplyingFn())
		baseFilters.DateRange = &dateRange

		if notes, err := handler.repository.SearchNotes(baseFilters); err != nil {
			return err
		} else {
			output.PrintNotes(notes, getPrintOptionsFromContext(ctx))
		}

		return nil
	}
}

func (handler handler) SearchNotes(ctx *cli.Context) error {
	if searchFilters, err := handler.getSearchFiltersFromContext(ctx); err != nil {
		return err
	} else {
		if notes, err := handler.repository.SearchNotes(searchFilters); err != nil {
			return err
		} else {
			output.PrintNotes(notes, getPrintOptionsFromContext(ctx))
		}

		return nil
	}
}

func (handler handler) deleteNote(note *dto.Note, confirm bool) error {
	logging.Debug("deleting note using full id " + note.ID)

	if confirm {
		fmt.Println("The following note will be deleted:")
		fmt.Println()

		output.PrintNote(note, output.Options{})

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

	if !dto.IsValidId(noteId) {
		return errInvalidNoteId
	}

	var note *dto.Note
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
			note.Tags = CleanTagsFromString(newTagsString)
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
	output.PrintNote(note, getPrintOptionsFromContext(ctx))

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
	tags := user_input.GetTagsFromContext(cli)
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

		note := dto.NewNote(tags, trimmedInput)
		if err := handler.repository.WriteNote(note); err != nil {
			log.Println("failed to save note: " + err.Error())
		}
	}

	return nil
}

// Attempt to find a note by it's ID, automatically resolving short vs long IDs.
func (handler *handler) findNoteById(noteId string) (*dto.Note, error) {
	var note *dto.Note
	var noteIdLen = len(noteId)

	if noteIdLen == dto.ShortIDLength {
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
				output.PrintNotes(notes, printOptions)
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

func (handler handler) getSearchFiltersFromContext(ctx *cli.Context) (*dto.SearchFilters, error) {
	searchFilters := &dto.SearchFilters{
		Tags:    user_input.GetTagsFromContext(ctx),
		Content: strings.TrimSpace(ctx.String("")),
		String:  strings.TrimSpace(strings.Join(ctx.Args().Slice(), " ")),
	}

	age := strings.ToLower(ctx.String(""))
	if len(age) > 0 {
		validAge := regexp.MustCompile(`^\d+d$`)
		if !validAge.MatchString(age) {
			return nil, errInvalidAge
		} else {
			ageDays, _ := strconv.Atoi(strings.TrimRight(age, "d"))
			end := handler.nowSupplyingFn()
			start := end.AddDate(0, 0, -ageDays)

			searchFilters.DateRange = &dto.DateRange{
				From: start,
				To:   end,
			}
		}
	}

	logging.Debug(fmt.Sprintf("search filters = %+v", searchFilters))

	return searchFilters, nil
}
