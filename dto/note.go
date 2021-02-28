package dto

import (
	uuid "github.com/satori/go.uuid"
	"time"
)

// Main note model.
type Note struct {
	ID        string
	Tags      []string
	Content   string
	Timestamp time.Time
}

// NewNote creates a note with a given content and tags.
func NewNote(tags []string, content string) Note {
	return Note{
		ID:        uuid.NewV4().String(),
		Timestamp: time.Now(),
		Tags:      tags,
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
