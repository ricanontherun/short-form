package command

import (
	"github.com/ricanontherun/short-form/user_input"
	"github.com/ricanontherun/short-form/utils"
	uuid "github.com/satori/go.uuid"
	"github.com/urfave/cli/v2"
	"strings"
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

type tagsType []string

type NoteDTO struct {
	Content string
	Tags    tagsType
}

type DeleteByNoteID struct {
	NoteID string
}

type deleteByTags struct {
	Tags []string
}

func NewDeleteFromContext(ctx *cli.Context) (*DeleteByNoteID, error) {
	d := &DeleteByNoteID{}

	noteId := strings.TrimSpace(ctx.Args().First())
	if len(noteId) > 0 && !isValidNoteId(noteId) {
		return nil, errInvalidNoteId
	}
	d.NoteID = noteId

	return d, nil
}

func NewDeleteByTagsFromContext(ctx *cli.Context) *deleteByTags {
	d := &deleteByTags{}

	set := utils.NewSet()
	for _, tag := range strings.Split(strings.ToLower(strings.TrimSpace(ctx.String("tags"))), ",") {
		set.Add(tag)
	}
	d.Tags = set.Entries()

	return d
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
