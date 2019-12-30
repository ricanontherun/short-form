package repository

import "github.com/ricanontherun/short-form/models"

type Repository interface {
	// Write a new note to the database
	WriteNote(note models.Note) error

	// Search for notes by tag, date or content
	SearchNotes(ctx models.SearchFilters) ([]models.Note, error)

	// Delete a note (hard delete)
	DeleteNote(noteId string) error

	// Fetch a single note from the database
	GetNote(noteId string) (*models.Note, error)

	// Update a note.
	UpdateNote(note models.Note) error

	TagNote(note models.Note, tags []string) error
}
