package command

import (
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

type Note struct {
	Content string
	Tags    tagsType
}

type Delete struct {
	// Delete using the note's short or long ID
	NoteID string

	// Delete notes which
	Tags []string
}

func NewDeleteFromContext(ctx *cli.Context) (*Delete, error) {
	d := &Delete{}

	// Read and validate note ID
	noteId := strings.TrimSpace(ctx.Args().First())
	if len(noteId) > 0 && !isValidNoteId(noteId) {
		return nil, errInvalidNoteId
	}
	d.NoteID = noteId

	// Read any "delete by" tags
	set := utils.NewSet()
	for _, tag := range strings.Split(strings.ToLower(strings.TrimSpace(ctx.String("tags"))), ",") {
		set.Add(tag)
	}
	d.Tags = set.Entries()

	return d, nil
}

func NewNoteFromContext(ctx *cli.Context) (*Note, error) {
	note := &Note{}

	if inputContent, err := readContentFromContext(ctx); err != nil {
		return nil, err
	} else if len(inputContent) == 0 {
		return nil, errEmptyContent
	} else {
		note.Content = inputContent
	}

	note.Tags = ReadTagsFromContext(ctx)

	return note, nil
}
