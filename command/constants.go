package command

import "errors"

var (
	errEmptyContent  = errors.New("empty content")
	errMissingNoteId = errors.New("missing note id")
	errInvalidNoteId = errors.New("invalid note id")
	errNoteNotFound  = errors.New("note not found")
	errInvalidAge    = errors.New("invalid age")
)

const (
	flagTags     = "tags"
	flagAge      = "age"
	flagContent  = "content"
	flagDetailed = "detailed"
	flagPretty   = "pretty"
)
