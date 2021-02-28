package command

import (
	"github.com/ricanontherun/short-form/database"
	"github.com/ricanontherun/short-form/dto"
	"github.com/ricanontherun/short-form/repository"
)

type writeCommand struct {
	dto *NoteDTO
}

func (command *writeCommand) Execute() error {
	repo := repository.NewSqlRepository(database.GetInstance())
	return repo.WriteNote(dto.NewNote(command.dto.Tags, command.dto.Content))
}

func NewWriteCommand(dto *NoteDTO) Command {
	return &writeCommand{dto,}
}
