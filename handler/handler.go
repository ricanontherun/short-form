package handler

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/ricanontherun/short-form/data"
	"github.com/ricanontherun/short-form/utils"
	uuid "github.com/satori/go.uuid"
	"github.com/urfave/cli/v2"
	"log"
	"os"
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
	repository data.Repository
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
	detailed  bool
}

func getPrintOptionsFromContext(ctx *cli.Context) printOptions {
	return printOptions{
		insecure:  ctx.Bool("insecure"),
		highlight: ctx.String("content"),
		detailed:  ctx.Bool("detailed"),
	}
}

func promptUser(message string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(message)
	text, _ := reader.ReadString('\n')
	return strings.TrimSpace(strings.ToLower(text))
}

func makeUserConfirmAction(message string) bool {
	return utils.InArray(promptUser(message+" [y/n]: "), []string{
		"yes",
		"y",
	})
}

func getArgStringFromContext(ctx *cli.Context) string {
	rawArgs := ctx.Args().Slice()
	clean := make([]string, 0, len(rawArgs))

	for _, arg := range rawArgs {
		clean = append(clean, strings.TrimSpace(arg))
	}

	return strings.Join(clean, " ")
}

func getInputFromContext(ctx *cli.Context) parsedInput {
	return parsedInput{
		content: strings.Join(ctx.Args().Slice(), " "),
		tags:    getTagsFromContext(ctx),
	}
}

// Return a cleaned array of tags provided as --tags=t1,t2,t3, as ['t1', 't2', 't3']
func getTagsFromContext(c *cli.Context) []string {
	return cleanTagsFromString(c.String("tags"))
}

func cleanTagsFromString(tagString string) []string {
	tags := utils.NewSet()

	for _, tag := range strings.Split(tagString, ",") {
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

	if noteCount <= 0 {
		return
	}

	noteCountString := ""
	if noteCount == 1 {
		noteCountString = fmt.Sprintf("1 note found")
	} else {
		noteCountString = fmt.Sprintf("%d notes found", noteCount)
	}

	fmt.Println(noteCountString)
	fmt.Println()

	for _, note := range notes {
		handler.printNote(note, options)
	}
}

func (handler Handler) printNote(note data.Note, options printOptions) {
	bits := make([]string, 0, 4)

	bits = append(bits, note.Timestamp.Format("January 02, 2006 03:04 PM"))

	if options.detailed {
		bits = append(bits, note.ID)

		if note.Secure {
			bits = append(bits, "secure")
		} else {
			bits = append(bits, "insecure")
		}
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
		contentString = utils.HighlightString(note.Content, options.highlight)
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

	return handler.repository.WriteNote(note)
}

func (handler Handler) SearchTodayNote(ctx *cli.Context) error {
	now := time.Now()

	searchFilters := getSearchFiltersFromContext(ctx)
	searchFilters.DateRange = &data.DateRange{
		From: time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()),
		To:   now,
	}

	if notes, err := handler.repository.SearchNotes(searchFilters); err != nil {
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

	if notes, err := handler.repository.SearchNotes(searchFilters); err != nil {
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

	if notes, err := handler.repository.SearchNotes(searchFilters); err != nil {
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
		fmt.Println("that's not a valid note id")
		return nil
	}

	// Make sure the note exists.
	if _, err := handler.repository.GetNote(noteId); err != nil {
		if err != data.ErrNoteNotFound {
			return err
		}

		fmt.Println("note not found")
		return nil
	}

	// Prompt the user for confirmation.
	if ok := makeUserConfirmAction("This will delete 1 note, are you sure?"); !ok {
		fmt.Println("cancelled")
		return nil
	}

	err := handler.repository.DeleteNote(noteId)
	if err == nil {
		fmt.Println("ok")
	} else if err == data.ErrNoteNotFound {
		fmt.Println("note not found")
	} else {
		return err
	}

	return nil
}

func (handler Handler) EditNote(ctx *cli.Context) error {
	// Get the noteId from context.
	noteId := ctx.Args().First()

	if len(noteId) == 0 {
		fmt.Println("missing note id")
		return nil
	}

	if _, err := uuid.FromString(noteId); err != nil {
		fmt.Println("invalid note id")
		return nil
	}

	note, err := handler.repository.GetNote(noteId)
	if err != nil {
		if err == data.ErrNoteNotFound {
			fmt.Println("not not found")
			return nil
		}

		return err
	}

	changed := false
	tagsChanged := false

	newContent := promptUser("New Content: ")
	if len(newContent) != 0 {
		if note.Secure {
			if newContentBytes, err := handler.encryptor.Encrypt([]byte(newContent)); err != nil {
				return err
			} else {
				newContent = string(newContentBytes)
			}
		}

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
