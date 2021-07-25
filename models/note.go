package models

import (
	uuid "github.com/satori/go.uuid"
	"time"
)

// Main note model.
type Note struct {
	ID        string
	Tags      []string
	Title     string
	Content   string
	Timestamp time.Time
}

// NewNote creates a note with a given content and tags.
func NewNote(title string, content string, tags []string) Note {
	return Note{
		ID:        uuid.NewV4().String(),
		Timestamp: time.Now(),
		Tags:      tags,
		Title:     title,
		Content:   content,
	}
}

// Clone creates a copy of a note.
func (note Note) Clone() Note {
	return Note{
		ID:        note.ID,
		Tags:      note.Tags,
		Content:   note.Content,
		Timestamp: note.Timestamp,
	}
}
