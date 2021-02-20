package command

import (
	"github.com/ricanontherun/short-form/database"
	"github.com/ricanontherun/short-form/models"
	"github.com/ricanontherun/short-form/repository"
)

type writeCommand struct {
	input *Note
}

func (command *writeCommand) Execute() error {
	repo := repository.NewSqlRepository(database.GetInstance())
	return repo.WriteNote(models.NewNote(command.input.Tags, command.input.Content))
}

func NewWriteCommand(data *Note) Command {
	return &writeCommand{
		data,
	}
}
