package repository

import "github.com/ricanontherun/short-form/models"

type Repository interface {
	// Write a new note to the database
	WriteNote(note models.Note) error
	// Update a note.
	UpdateNote(note models.Note) error
	// Apply tags to a note
	TagNote(note models.Note, tags []string) error
	// Search for notes by tag, date or content
	SearchNotes(ctx models.SearchFilters) ([]*models.Note, error)

	// Delete a note (hard delete)
	DeleteNote(noteId string) error
	DeleteNoteByTag(tag string) error

	// Fetch a single note from the database
	LookupNote(noteId string) (*models.Note, error)
	// Fetch a single note from the database along with it's tags.
	LookupNoteWithTags(noteId string) (*models.Note, error)
	// Fetch
	LookupNotesByShortId(shortId string) ([]*models.Note, error)
}
