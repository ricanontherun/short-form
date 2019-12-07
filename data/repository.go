package data

type Repository interface {
	WriteNote(note Note) error
	SearchNotes(ctx Filters) ([]Note, error)
	Close()
}
