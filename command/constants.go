package command

import "errors"

var (
	ErrEmptyContent  = errors.New("empty content")
	ErrMissingNoteId = errors.New("missing note id")
	ErrInvalidNoteId = errors.New("invalid note id")
	ErrNoteNotFound  = errors.New("note not found")
	ErrInvalidAge    = errors.New("invalid age")
)

const (
	FlagTags     = "tags"
	FlagAge      = "age"
	FlagContent  = "content"
	FlagDetailed = "detailed"
)
