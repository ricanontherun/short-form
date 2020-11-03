// These tests assert that the handler can correctly process user CLI input.
package command

import (
	"flag"
	"github.com/ricanontherun/short-form/models"
	"github.com/ricanontherun/short-form/repository"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/urfave/cli/v2"
	"log"
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
			expectedErr: errEmptyContent,
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

		h := NewHandlerBuilder(&r).Build()

		err := h.WriteNote(context)

		if test.expectedErr != nil {
			r.AssertNotCalled(t, "WriteNote", mock.Anything)
			assert.EqualValues(t, test.expectedErr, err)
		} else {
			if written := r.AssertNumberOfCalls(t, "WriteNote", 1); written {
				noteArgument := r.Calls[0].Arguments.Get(0).(models.Note)

				sort.Strings(noteArgument.Tags)
				sort.Strings(test.expectedNote.Tags)

				assert.EqualValues(t, test.expectedNote.Content, noteArgument.Content)
				assert.EqualValues(t, test.expectedNote.Tags, noteArgument.Tags)

				_, err := uuid.FromString(noteArgument.ID)
				assert.Nil(t, err)
			} else {
				log.Fatalf("test=+%v, context=+%v", test, context)
			}
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
		h := NewHandlerBuilder(&r).Build()

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

		h := NewHandlerBuilder(&r).Build()

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
			expectedErr: errInvalidAge,
		},
		{
			inputAge:    "1m",
			expectedErr: errInvalidAge,
		},
		{
			inputAge:    "10",
			expectedErr: errInvalidAge,
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
		h := NewHandlerBuilder(&r).WithNowSupplier(func() time.Time {
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

	input := NewMockInput()
	input.On("GetString", mock.Anything).Return("y")

	for _, test := range tests {
		var flags = map[string]string{
			"no-confirm": "true",
		}

		context := createAppContext(flags, test.inputArgs)
		r := repository.NewMockRepository()

		r.On("LookupNoteWithTags", noteId).Return(&models.Note{
			ID: noteId,
		}, nil)

		r.On("DeleteNote", noteId).Return(nil)

		h := NewHandlerBuilder(&r).WithUserInputController(input).Build()
		err := h.DeleteNote(context)

		assert.Nil(t, err)
		r.AssertNumberOfCalls(t, "LookupNoteWithTags", 1)
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
			expectedErr: errMissingNoteId,
		},

		// Bad note id.
		{
			inputArgs:   []string{invalidNoteId},
			expectedErr: errInvalidNoteId,
		},
	}

	input := NewMockInput()
	input.On("GetString", mock.Anything).Return("y")

	for _, test := range tests {
		var flags = map[string]string{}

		context := createAppContext(flags, test.inputArgs)
		r := repository.NewMockRepository()

		h := NewHandlerBuilder(&r).WithUserInputController(input).Build()
		err := h.DeleteNote(context)

		assert.NotNil(t, err)
		assert.EqualValues(t, test.expectedErr, err)
	}
}

func TestHandler_DeleteNote_ShortId(t *testing.T) {
	// Given a short ID which belongs to a single note
	var fullId = uuid.NewV4().String()
	var shortId = fullId[0:8]

	var flags = map[string]string{
		"no-confirm": "true",
	}
	var context = createAppContext(flags, []string{shortId})

	var mockRepository = repository.NewMockRepository()
	mockRepository.On("LookupNotesByShortId", shortId).Return([]*models.Note{
		{ID: fullId},
	}, nil)
	mockRepository.On("DeleteNote", fullId).Return(nil)

	// when we delete a note by that short id
	var handler = NewHandlerBuilder(&mockRepository).Build()
	err := handler.DeleteNote(context)

	// it should successfully call the repository delete method using the full ID
	// returned from the note lookup
	assert.Nil(t, err)
	mockRepository.AssertCalled(t, "DeleteNote", fullId)
}

func TestHandler_EditNote_MissingNoteId(t *testing.T) {
	var flags = map[string]string{}
	context := createAppContext(flags, []string{})

	r := repository.NewMockRepository()
	h := NewHandlerBuilder(&r).Build()

	err := h.EditNote(context)

	assert.NotNil(t, err)
	assert.EqualValues(t, errMissingNoteId, err)

	r.AssertNotCalled(t, "LookupNoteWithTags", mock.Anything)
}

// Invalid note id.
func TestHandler_EditNote_InvalidNoteid(t *testing.T) {
	var flags = map[string]string{}
	context := createAppContext(flags, []string{"not a UUID"})

	r := repository.NewMockRepository()
	h := NewHandlerBuilder(&r).Build()

	err := h.EditNote(context)

	assert.NotNil(t, err)
	assert.EqualValues(t, errInvalidNoteId, err)

	r.AssertNotCalled(t, "LookupNoteWithTags", mock.Anything)
}

// Note not found.
func TestHandler_EditNote_InvalidNoteNotFound(t *testing.T) {
	var flags = map[string]string{}
	noteId := uuid.NewV4().String()
	context := createAppContext(flags, []string{noteId})

	r := repository.NewMockRepository()
	r.On("LookupNoteWithTags", noteId).Return(&models.Note{}, errNoteNotFound)

	h := NewHandlerBuilder(&r).Build()

	err := h.EditNote(context)

	assert.NotNil(t, err)
	assert.EqualValues(t, errNoteNotFound, err)
	r.AssertNumberOfCalls(t, "LookupNoteWithTags", 1)
}

func TestHandler_EditNote_AcceptInput(t *testing.T) {
}

// Success case.
func TestHandler_EditNote(t *testing.T) {
	var flags = map[string]string{}
	noteId := uuid.NewV4().String()
	context := createAppContext(flags, []string{noteId})

	r := repository.NewMockRepository()
	note := &models.Note{
		ID:      noteId,
		Content: "note content",
		Tags:    []string{"music", "general"},
	}

	r.On("LookupNoteWithTags", noteId).Return(note, nil)

	h := NewHandlerBuilder(&r).Build()

	err := h.EditNote(context)

	assert.Nil(t, err)
	r.AssertNumberOfCalls(t, "LookupNoteWithTags", 1)
}
