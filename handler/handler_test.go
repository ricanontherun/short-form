package handler_test

import (
	"flag"
	"github.com/ricanontherun/short-form/data"
	"github.com/ricanontherun/short-form/handler"
	"github.com/stretchr/testify/mock"
	"github.com/urfave/cli/v2"
	"reflect"
	"testing"
)

func createAppContext(flags map[string]string, args []string) *cli.Context {
	flagSet := flag.NewFlagSet("f1", flag.ContinueOnError)

	for key, value := range flags {
		flagSet.String(key, value, "")
	}

	if err := flagSet.Parse(args); err != nil {
		panic(err)
	}

	return cli.NewContext(cli.NewApp(), flagSet, nil)
}

func assertNoteEquality(t *testing.T, expected data.Note, actual data.Note) {
	if actual.Content != expected.Content {
		t.Error("should have been called with expected content")
	} else if !reflect.DeepEqual(actual.Tags, expected.Tags) {
		t.Error("should have been called with expected tags")
	} else { // This is fine.
		return
	}

	t.FailNow()
}

func TestHandler_WriteNote_FailsWhenNotGivenContent(t *testing.T) {
	var flags = map[string]string{}

	context := createAppContext(flags, []string{})

	repository := data.NewMockRepository()
	repository.On("WriteNote", mock.Anything).Return(nil)

	h := handler.NewHandler(&repository)

	repository.AssertNotCalled(t, "WriteNote", mock.Anything)

	if err := h.WriteNote(context); err == nil {
		t.Error("should have returned an error")
	} else if err != handler.ErrEmptyContent {
		t.Error("error should have been ErrEmptyContent")
	}
}

func TestHandler_WriteNote_HappyPath(t *testing.T) {
	var flags = map[string]string{
		"tags": "one,two,three",
	}

	context := createAppContext(flags, []string{
		"This",
		"is",
		"the",
		"content",
	})

	repository := data.NewMockRepository()
	repository.On("WriteNote", mock.Anything).Return(nil)

	h := handler.NewHandler(&repository)

	if err := h.WriteNote(context); err != nil {
		t.Error("should not have returned an error")
		t.FailNow()
	}

	repository.AssertNumberOfCalls(t, "WriteNote", 1)
	noteArgument := repository.Calls[0].Arguments.Get(0).(data.Note)

	assertNoteEquality(t, data.Note{
		Tags:    []string{"one", "two", "three"},
		Content: "This is the content",
	}, noteArgument)
}
