package data

type Repository interface {
	// Write a new note to the database
	WriteNote(note Note) error

	// Search for notes by tag, date or content
	SearchNotes(ctx Filters) ([]Note, error)

	// Delete a note (hard delete)
	DeleteNote(noteId string) error

	// Fetch a single note from the database
	GetNote(noteId string) (Note, error)

	UpdateNoteContent(nodeId string, content string) error

	// Close the repository
	Close()
}
