package command

import (
	"github.com/ricanontherun/short-form/user_input"
	uuid "github.com/satori/go.uuid"
	"github.com/urfave/cli/v2"
)

var (
	IdLength      = len(uuid.NamespaceDNS.String())
	ShortIdLength = 8
)

func isValidNoteId(id string) bool {
	var idLen = len(id)
	var validLength = idLen == IdLength || idLen == ShortIdLength
	var validForm = true

	if idLen == IdLength {
		if _, err := uuid.FromString(id); err != nil {
			validForm = false
		}
	}

	return validLength && validForm
}

type NoteDTO struct {
	Content string
	Tags    []string
}

type DeleteByNoteID struct {
	NoteID string
}

func NewNoteDTOFromContext(ctx *cli.Context) (*NoteDTO, error) {
	note := &NoteDTO{}

	if inputContent, err := user_input.GetContentFromContext(ctx); err != nil {
		return nil, err
	} else if len(inputContent) == 0 {
		return nil, errEmptyContent
	} else {
		note.Content = inputContent
	}

	note.Tags = user_input.GetTagsFromContext(ctx)

	return note, nil
}
