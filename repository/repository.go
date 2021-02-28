package repository

import "github.com/ricanontherun/short-form/dto"

type Repository interface {
	// Write a new note to the database
	WriteNote(note dto.Note) error
	// Update a note.
	UpdateNote(note dto.Note) error
	// Apply tags to a note
	TagNote(note dto.Note, tags []string) error
	// Search for notes by tag, date or content
	SearchNotes(ctx *dto.SearchFilters) ([]*dto.Note, error)

	// Delete a note (hard delete)
	DeleteNote(noteId string) error
	DeleteNoteByTags(tags []string) error

	// Fetch a single note from the database
	LookupNote(noteId string) (*dto.Note, error)
	// Fetch a single note from the database along with it's tags.
	LookupNoteWithTags(noteId string) (*dto.Note, error)

	// Fetch
	LookupNotesByShortId(shortId string) ([]*dto.Note, error)

	// Fetch the number of notes associated with a set of tags.
	GetNoteCountByTags(tags []string) (uint64, error)
}
