package command

import (
	testing2 "github.com/ricanontherun/short-form/test_utils"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_IsValidNoteId(t *testing.T) {
	testUUID := uuid.NewV4().String()
	assert.True(t, isValidNoteId(testUUID))
	assert.True(t, isValidNoteId(testUUID[:8])) // short ids are the first 8 bytes of a UUID.

	assert.False(t, isValidNoteId(""))
	assert.False(t, isValidNoteId(testUUID[:9]))
}

func Test_NewNoteDTOFromContext(t *testing.T) {
	ctx := testing2.CreateAppContext(map[string]string{
		"tags": " one,two ",
	}, []string{"this", "is the", "content"})

	note, err := NewNoteDTOFromContext(ctx)

	assert.Nil(t, err)
	assert.NotNil(t, note)

	assert.Equal(t,[]string{"one", "two"}, note.Tags)
	assert.Equal(t, "this is the content", note.Content)
}