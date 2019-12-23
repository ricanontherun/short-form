package data

import (
	uuid "github.com/satori/go.uuid"
	"time"
)

type Note struct {
	ID        string
	Tags      []string
	Content   string
	Timestamp time.Time
}

func NewNote(tags []string, content string) Note {
	return Note{
		ID:        uuid.NewV4().String(),
		Timestamp: time.Now(),
		Tags:      tags,
		Content:   content,
	}
}

func (note Note) Clone() Note {
	return Note{
		ID:        note.ID,
		Tags:      note.Tags,
		Content:   note.Content,
		Timestamp: note.Timestamp,
	}
}
