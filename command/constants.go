package command

import "errors"

var (
	errEmptyContent     = errors.New("empty content")
	errMissingNoteId    = errors.New("missing note id")
	errInvalidNoteId    = errors.New("invalid note id")
	errNoteNotFound     = errors.New("note not found")
	errShortIdCollision = errors.New("duplicate short id")
	errInvalidAge       = errors.New("invalid age")
	errMissingTitle 	= errors.New("missing title")
)

const (
	flagTags      = "tags"
	flagAge       = "age"
	flagContent   = "content"
	flagNoConfirm = "no-confirm"
)
