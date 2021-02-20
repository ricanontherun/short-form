package command

import (
	"errors"
	"github.com/ricanontherun/short-form/database"
	"github.com/ricanontherun/short-form/models"
	"github.com/ricanontherun/short-form/repository"
)

type deleteCommand struct {
	data *Delete
	repository repository.Repository
}

type deleteNoteByIDStrategy struct {
	noteId string
}

func (command *deleteCommand) Execute() error {
	// We prefer deleting by ids.
	if len(command.data.NoteID) > 0 {
		return command.deleteById()
	}

	return command.deleteByTags()
}

func (command *deleteCommand) deleteById() error {
	noteId := command.data.NoteID

	// "Simple" case, we're given a full UUID note ID.
	if len(noteId) == models.LongIDLength {
		if _, err := command.repository.LookupNoteWithTags(noteId); err != nil {
			return err
		} else { // Delete by PK
			return command.repository.DeleteNote(noteId)
		}
	}

	// More complicated case, we're given a short ID.
	// Albeit unlikely, it's possible that >1 note(s) have IDs which start with the provided short ID.
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
			return errors.New("you suck")
		}
	}
}

func (command *deleteCommand) deleteByTags() error {
	return nil
}

func NewDeleteCommand(dto *Delete) Command {
	return &deleteCommand{dto, repository.NewSqlRepository(database.GetInstance())}
}
