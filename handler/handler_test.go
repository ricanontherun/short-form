// These tests assert that the handler can correctly process user CLI input.
package handler_test

import (
	"flag"
	"github.com/ricanontherun/short-form/data"
	"github.com/ricanontherun/short-form/handler"
	"github.com/ricanontherun/short-form/utils"
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
		expectedNote data.Note
		expectedErr  error
	}{
		// Missing content
		{
			inputArgs:   []string{""},
			inputTags:   "one,two",
			expectedErr: handler.ErrEmptyContent,
		},

		// Tags aren't required.
		{
			inputArgs: []string{"tags", "aren't", "required"},
			expectedNote: data.Note{
				Tags:    []string{},
				Content: "tags aren't required",
			},
		},

		{
			inputTags: "  git,   CLI  	",
			inputArgs: []string{"This", "is", "THE", "CONTENT"},
			expectedNote: data.Note{
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

		repository := data.NewMockRepository()
		repository.On("WriteNote", mock.Anything).Return(nil)

		h := handler.NewHandler(&repository)

		err := h.WriteNote(context)
		if test.expectedErr != nil {
			repository.AssertNotCalled(t, "WriteNote", mock.Anything)

			assert.EqualValues(t, test.expectedErr, err)
		} else {
			repository.AssertNumberOfCalls(t, "WriteNote", 1)
			noteArgument := repository.Calls[0].Arguments.Get(0).(data.Note)

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
		expectedDateRange utils.DateRange
	}{
		// No input
		{
			inputContent:      "",
			inputTags:         "",
			expectedTags:      []string{},
			expectedContent:   "",
			expectedDateRange: utils.GetRangeToday(now),
		},

		// Intentionally space padded input
		{
			inputContent: " rebase  		",
			inputTags:         "    git,    cli   ",
			expectedTags:      []string{"git", "cli"},
			expectedContent:   "rebase",
			expectedDateRange: utils.GetRangeToday(now),
		},
	}

	for _, test := range tests {
		var flags = map[string]string{
			"tags":    test.inputTags,
			"content": test.inputContent,
		}

		context := createAppContext(flags, []string{})
		repository := data.NewMockRepository()
		repository.On("SearchNotes", mock.Anything).Return(nil)
		h := handler.NewHandler(&repository)

		// When
		if err := h.SearchToday(context); err != nil {
			t.Fatalf("Should not have returned an error, received: %s", err.Error())
		}

		// Then
		repository.AssertNumberOfCalls(t, "SearchNotes", 1)

		sort.Strings(test.expectedTags)
		filters := repository.Calls[0].Arguments.Get(0).(data.Filters)
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
		expectedDateRange utils.DateRange
	}{
		// No input
		{
			inputContent:      "",
			inputTags:         "",
			expectedTags:      []string{},
			expectedContent:   "",
			expectedDateRange: utils.GetRangeYesterday(now),
		},

		// Intentionally space padded input
		{
			inputContent: " rebase  		",
			inputTags:         "    git,    cli   ",
			expectedTags:      []string{"git", "cli"},
			expectedContent:   "rebase",
			expectedDateRange: utils.GetRangeYesterday(now),
		},
	}

	for _, test := range tests {
		var flags = map[string]string{
			"tags":    test.inputTags,
			"content": test.inputContent,
		}

		context := createAppContext(flags, []string{})
		repository := data.NewMockRepository()
		repository.On("SearchNotes", mock.Anything).Return(nil)
		h := handler.NewHandler(&repository)

		// When
		if err := h.SearchYesterday(context); err != nil {
			t.Fatalf("Should not have returned an error, received: %s", err.Error())
		}

		// Then
		repository.AssertNumberOfCalls(t, "SearchNotes", 1)

		filters := repository.Calls[0].Arguments.Get(0).(data.Filters)
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
		expectedDateRange *utils.DateRange
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
			expectedErr: handler.ErrInvalidAge,
		},
		{
			inputAge:    "1m",
			expectedErr: handler.ErrInvalidAge,
		},
		{
			inputAge:    "10",
			expectedErr: handler.ErrInvalidAge,
		},
		{
			inputAge:     "2d",
			expectedTags: []string{},
			expectedDateRange: &utils.DateRange{
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
			expectedDateRange: &utils.DateRange{
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

		repository := data.NewMockRepository()
		repository.On("SearchNotes", mock.Anything).Return(nil)
		h := handler.NewHandlerFromArguments(&repository, func() time.Time {
			return now
		}, handler.NewUserInput())

		err := h.SearchNotes(context)
		if test.expectedErr != nil {
			assert.EqualValues(t, test.expectedErr, err)
		} else {
			repository.AssertNumberOfCalls(t, "SearchNotes", 1)
			filters := repository.Calls[0].Arguments.Get(0).(data.Filters)

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

	for _, test := range tests {
		var flags = map[string]string{
			"no-confirm": "true", // >.>
		}

		context := createAppContext(flags, test.inputArgs)
		repository := data.NewMockRepository()

		repository.On("GetNote", noteId).Return(&data.Note{
			ID: noteId,
		}, nil)

		repository.On("DeleteNote", noteId).Return(nil)

		h := handler.NewHandler(&repository)
		err := h.DeleteNote(context)

		assert.Nil(t, err)
		repository.AssertNumberOfCalls(t, "GetNote", 1)
		repository.AssertCalled(t, "DeleteNote", noteId)
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
			expectedErr: handler.ErrMissingNoteId,
		},

		// Bad note id.
		{
			inputArgs:   []string{invalidNoteId},
			expectedErr: handler.ErrInvalidNoteId,
		},
	}

	for _, test := range tests {
		var flags = map[string]string{
			"no-confirm": "true", // >.>
		}

		context := createAppContext(flags, test.inputArgs)
		repository := data.NewMockRepository()

		h := handler.NewHandler(&repository)
		err := h.DeleteNote(context)

		assert.NotNil(t, err)
		assert.EqualValues(t, test.expectedErr, err)
	}
}
