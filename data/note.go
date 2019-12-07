package data

import (
	uuid "github.com/satori/go.uuid"
	"time"
)

type Note struct {
	ID              string
	Tags            []string
	Content         string
	Timestamp       time.Time
	Secure          bool
	secured         bool
}

func NewSecureNote(tags []string, content string) Note {
	return newNote(tags, content, true)
}

func NewInsecureNote(tags []string, content string) Note {
	return newNote(tags, content, false)
}

func newNote(tags []string, content string, secure bool) Note {
	return Note{
		ID:        uuid.NewV4().String(),
		Timestamp: time.Now(),
		Tags:      tags,
		Content:   content,
		Secure:    secure,
		secured:   false,
	}
}

func (note Note) Clone() Note {
	return Note{
		ID:        note.ID,
		Tags:      note.Tags,
		Content:   note.Content,
		Timestamp: note.Timestamp,
		Secure:    note.Secure,
		secured:   note.secured,
	}
}