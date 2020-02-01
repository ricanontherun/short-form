// These tests assert that the handler can correctly process user CLI input.
package command_test

import (
	"flag"
	"github.com/ricanontherun/short-form/command"
	"github.com/ricanontherun/short-form/models"
	"github.com/ricanontherun/short-form/repository"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/urfave/cli/v2"
	"sort"
	"testing"
	"time"
)

func createAppContext(flags map[string]string, args []string) *cli.Context {
	flagSet := flag.NewFlagSet("tests", flag.ContinueOnError)

	for key, value := range flags {
		flagSet.String(key, value, "")
	}

	if err := flagSet.Parse(args); err != nil {
		panic(err)
	}

	return cli.NewContext(cli.NewApp(), flagSet, nil)
}

func TestHandler_WriteNote(t *testing.T) {
	tests := []struct {
		inputTags    string
		inputArgs    []string
		expectedNote models.Note
		expectedErr  error
	}{
		// Missing content
		{
			inputArgs:   []string{""},
			inputTags:   "one,two",
			expectedErr: command.ErrEmptyContent,
		},

		// Tags aren't required.
		{
			inputArgs: []string{"tags", "aren't", "required"},
			expectedNote: models.Note{
				Tags:    []string{},
				Content: "tags aren't required",
			},
		},

		{
			inputTags: "  git,   CLI  	",
			inputArgs: []string{"This", "is", "THE", "CONTENT"},
			expectedNote: models.Note{
				Tags:    []string{"git", "CLI"},
				Content: "This is THE CONTENT",
			},
		},
	}

	for _, test := range tests {
		var flags = map[string]string{
			"tags": test.inputTags,
		}

		context := createAppContext(flags, test.inputArgs)

		r := repository.NewMockRepository()
		r.On("WriteNote", mock.Anything).Return(nil)

		h := command.NewHandlerBuilder(&r).Build()

		err := h.WriteNote(context)
		if test.expectedErr != nil {
			r.AssertNotCalled(t, "WriteNote", mock.Anything)

			assert.EqualValues(t, test.expectedErr, err)
		} else {
			r.AssertNumberOfCalls(t, "WriteNote", 1)
			noteArgument := r.Calls[0].Arguments.Get(0).(models.Note)

			sort.Strings(noteArgument.Tags)
			sort.Strings(test.expectedNote.Tags)

			assert.EqualValues(t, test.expectedNote.Content, noteArgument.Content)
			assert.EqualValues(t, test.expectedNote.Tags, noteArgument.Tags)

			_, err := uuid.FromString(noteArgument.ID)
			assert.Nil(t, err)
		}
	}
}

func TestHandler_SearchToday(t *testing.T) {
	now := time.Now()
	tests := []struct {
		inputTags         string
		inputContent      string
		expectedTags      []string
		expectedContent   string
		expectedDateRange models.DateRange
	}{
		// No input
		{
			inputContent:      "",
			inputTags:         "",
			expectedTags:      []string{},
			expectedContent:   "",
			expectedDateRange: models.GetRangeToday(now),
		},

		// Intentionally space padded input
		{
			inputContent: " rebase  		",
			inputTags:         "    git,    cli   ",
			expectedTags:      []string{"git", "cli"},
			expectedContent:   "rebase",
			expectedDateRange: models.GetRangeToday(now),
		},
	}

	for _, test := range tests {
		var flags = map[string]string{
			"tags":    test.inputTags,
			"content": test.inputContent,
		}

		context := createAppContext(flags, []string{})
		r := repository.NewMockRepository()
		r.On("SearchNotes", mock.Anything).Return(nil)
		h := command.NewHandlerBuilder(&r).Build()

		// When
		if err := h.SearchToday(context); err != nil {
			t.Fatalf("Should not have returned an error, received: %s", err.Error())
		}

		// Then
		r.AssertNumberOfCalls(t, "SearchNotes", 1)

		sort.Strings(test.expectedTags)
		filters := r.Calls[0].Arguments.Get(0).(models.SearchFilters)
		sort.Strings(filters.Tags)
		assert.EqualValues(t, test.expectedTags, filters.Tags)
		assert.EqualValues(t, test.expectedContent, filters.Content)
		assert.EqualValues(t, test.expectedDateRange, *filters.DateRange)
	}
}

func TestHandler_SearchYesterday(t *testing.T) {
	now := time.Now()
	tests := []struct {
		inputTags         string
		inputContent      string
		expectedTags      []string
		expectedContent   string
		expectedDateRange models.DateRange
	}{
		// No input
		{
			inputContent:      "",
			inputTags:         "",
			expectedTags:      []string{},
			expectedContent:   "",
			expectedDateRange: models.GetRangeYesterday(now),
		},

		// Intentionally space padded input
		{
			inputContent: " rebase  		",
			inputTags:         "    git,    cli   ",
			expectedTags:      []string{"git", "cli"},
			expectedContent:   "rebase",
			expectedDateRange: models.GetRangeYesterday(now),
		},
	}

	for _, test := range tests {
		var flags = map[string]string{
			"tags":    test.inputTags,
			"content": test.inputContent,
		}

		context := createAppContext(flags, []string{})
		r := repository.NewMockRepository()
		r.On("SearchNotes", mock.Anything).Return(nil)

		h := command.NewHandlerBuilder(&r).Build()

		// When
		if err := h.SearchYesterday(context); err != nil {
			t.Fatalf("Should not have returned an error, received: %s", err.Error())
		}

		// Then
		r.AssertNumberOfCalls(t, "SearchNotes", 1)

		filters := r.Calls[0].Arguments.Get(0).(models.SearchFilters)
		sort.Strings(test.expectedTags)
		sort.Strings(filters.Tags)
		assert.EqualValues(t, test.expectedTags, filters.Tags)
		assert.EqualValues(t, test.expectedContent, filters.Content)
		assert.EqualValues(t, test.expectedDateRange, *filters.DateRange)
	}
}

func TestHandler_SearchNotes(t *testing.T) {
	now := time.Now()

	tests := []struct {
		inputTags    string
		inputContent string
		inputAge     string

		expectedTags      []string
		expectedContent   string
		expectedDateRange *models.DateRange
		expectedErr       error
	}{
		// No input
		{
			expectedTags:      []string{},
			expectedDateRange: nil,
		},

		// Invalid age
		{
			inputAge:    "what",
			expectedErr: command.ErrInvalidAge,
		},
		{
			inputAge:    "1m",
			expectedErr: command.ErrInvalidAge,
		},
		{
			inputAge:    "10",
			expectedErr: command.ErrInvalidAge,
		},
		{
			inputAge:     "2d",
			expectedTags: []string{},
			expectedDateRange: &models.DateRange{
				From: now.AddDate(0, 0, -2),
				To:   now,
			},
		},
		{
			inputTags:    "git, CLI, version-CONTROL",
			inputContent: "REBASE",
			inputAge:     "100d",

			expectedTags:    []string{"git", "CLI", "version-CONTROL"},
			expectedContent: "REBASE",
			expectedDateRange: &models.DateRange{
				From: now.AddDate(0, 0, -100),
				To:   now,
			},
		},
	}

	for _, test := range tests {
		var flags = map[string]string{
			"tags":    test.inputTags,
			"content": test.inputContent,
			"age":     test.inputAge,
		}

		context := createAppContext(flags, []string{})

		r := repository.NewMockRepository()
		r.On("SearchNotes", mock.Anything).Return(nil)
		h := command.NewHandlerBuilder(&r).WithNowSupplier(func() time.Time {
			return now
		}).Build()

		err := h.SearchNotes(context)
		if test.expectedErr != nil {
			assert.EqualValues(t, test.expectedErr, err)
		} else {
			r.AssertNumberOfCalls(t, "SearchNotes", 1)
			filters := r.Calls[0].Arguments.Get(0).(models.SearchFilters)

			sort.Strings(filters.Tags)
			sort.Strings(test.expectedTags)

			assert.EqualValues(t, test.expectedTags, filters.Tags)
			assert.EqualValues(t, test.expectedContent, filters.Content)
			assert.EqualValues(t, test.expectedDateRange, filters.DateRange)
		}
	}
}

func TestHandler_DeleteNote(t *testing.T) {
	noteId := uuid.NewV4().String()

	tests := []struct {
		inputArgs []string
	}{
		{
			inputArgs: []string{noteId},
		},
	}

	input := command.NewMockInput()
	input.On("GetString", mock.Anything).Return("y")

	for _, test := range tests {
		var flags = map[string]string{}

		context := createAppContext(flags, test.inputArgs)
		r := repository.NewMockRepository()

		r.On("LookupNote", noteId).Return(&models.Note{
			ID: noteId,
		}, nil)

		r.On("DeleteNote", noteId).Return(nil)

		h := command.NewHandlerBuilder(&r).WithUserInputController(input).Build()
		err := h.DeleteNote(context)

		assert.Nil(t, err)
		r.AssertNumberOfCalls(t, "LookupNote", 1)
		r.AssertCalled(t, "DeleteNote", noteId)
	}
}

func TestHandler_DeleteNote_Validation(t *testing.T) {
	invalidNoteId := "not a uuid"

	tests := []struct {
		inputArgs   []string
		expectedErr error
	}{
		// Missing ID
		{
			inputArgs:   []string{},
			expectedErr: command.ErrMissingNoteId,
		},

		// Bad note id.
		{
			inputArgs:   []string{invalidNoteId},
			expectedErr: command.ErrInvalidNoteId,
		},
	}

	input := command.NewMockInput()
	input.On("GetString", mock.Anything).Return("y")

	for _, test := range tests {
		var flags = map[string]string{}

		context := createAppContext(flags, test.inputArgs)
		r := repository.NewMockRepository()

		h := command.NewHandlerBuilder(&r).WithUserInputController(input).Build()
		err := h.DeleteNote(context)

		assert.NotNil(t, err)
		assert.EqualValues(t, test.expectedErr, err)
	}
}
