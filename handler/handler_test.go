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
		// Given
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

		filters := repository.Calls[0].Arguments.Get(0).(data.Filters)
		assert.EqualValues(t, test.expectedTags, filters.Tags)
		assert.Equal(t, test.expectedContent, filters.Content)
		assert.EqualValues(t, test.expectedDateRange, *filters.DateRange)
	}
}
