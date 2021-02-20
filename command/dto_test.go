package command

import (
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