package command

import (
	"fmt"
	"github.com/ricanontherun/short-form/database"
	"github.com/ricanontherun/short-form/dto"
	"github.com/ricanontherun/short-form/repository"
)

type NoteCollisionError struct {
	noteId string
	notes  []*dto.Note
}

func newCollisionError(noteId string, notes []*dto.Note) error {
	return &NoteCollisionError{noteId, notes}
}

func (e *NoteCollisionError) Error() string {
	return fmt.Sprintf("that short ID (%s) is not unique, please try again with a full ID", e.noteId)
}

func (e *NoteCollisionError) GetNotes() []*dto.Note {
	return e.notes
}

type deleteByNoteIDCommand struct {
	data       *DeleteByNoteID
	repository repository.Repository
}

func (command *deleteByNoteIDCommand) Execute() error {
	noteId := command.data.NoteID

	// "Simple" case, we're given a full UUID note ID.
	if len(noteId) == dto.LongIDLength {
		if _, err := command.repository.LookupNoteWithTags(noteId); err != nil {
			return err
		} else { // DeleteByNoteID by PK
			return command.repository.DeleteNote(noteId)
		}
	}

	// More complicated case, we're given a short ID.
	// Albeit unlikely, it's possible that multiple notes have IDs which start with the provided short ID.
	if notesBeginningWithId, err := command.repository.LookupNotesByShortId(noteId); err != nil {
		return err
	} else {
		numNotes := len(notesBeginningWithId)

		switch numNotes {
		case 0:
			return errNoteNotFound
		case 1: // ideal, we've found a single note.
			return command.repository.DeleteNote(notesBeginningWithId[0].ID)
		default: // unlikely, >1 note begins with noteId
			return newCollisionError(noteId, notesBeginningWithId)
		}
	}
}

func (command *deleteByNoteIDCommand) deleteById() error {
	noteId := command.data.NoteID

	// "Simple" case, we're given a full UUID note ID.
	if len(noteId) == dto.LongIDLength {
		if _, err := command.repository.LookupNoteWithTags(noteId); err != nil {
			return err
		} else { // DeleteByNoteID by PK
			return command.repository.DeleteNote(noteId)
		}
	}

	// More complicated case, we're given a short ID.
	// Albeit unlikely, it's possible that multiple notes have IDs which start with the provided short ID.
	if notesBeginningWithId, err := command.repository.LookupNotesByShortId(noteId); err != nil {
		return err
	} else {
		numNotes := len(notesBeginningWithId)

		switch numNotes {
		case 0:
			return errNoteNotFound
		case 1: // ideal, we've found a single note.
			return command.repository.DeleteNote(notesBeginningWithId[0].ID)
		default: // unlikely, >1 note begins with noteId
			return newCollisionError(noteId, notesBeginningWithId)
		}
	}
}

func NewDeleteCommand(dto *DeleteByNoteID) Command {
	return &deleteByNoteIDCommand{dto, repository.NewSqlRepository(database.GetInstance())}
}

// -- deleteByTagsCommand
type deleteByTagsCommand struct {
	tags       []string
	repository repository.Repository
}

func NewDeleteByTagsCommand(tags []string) Command {
	return &deleteByTagsCommand{
		tags,
		repository.NewSqlRepository(database.GetInstance()),
	}
}

func (command *deleteByTagsCommand) Execute() error {
	return command.repository.DeleteNoteByTags(command.tags)
}
