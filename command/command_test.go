package command

import (
	"fmt"
	"github.com/ricanontherun/short-form/database"
	testing2 "github.com/ricanontherun/short-form/test_utils"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestNewWriteCommand(t *testing.T) {
	// Each test:
		// Delete the test.db file from disk (to force a clean one being created on each command/test)
		// Issue a command: go run main.go ...
		// Run assertions against database and possibly command output.

}

func TestWriteCommand_Execute(t *testing.T) {
	testing2.VerifyDatabaseIsEmpty(t)

	writeCommand := NewWriteCommand(&NoteDTO{
		Content: "short-form is a command line journaling app",
		Tags: []string{"tag1", "tag2"},
	})

	assert.Nil(t, writeCommand.Execute())

	var id string
	assert.Nil(t, testing2.QueryOne("SELECT id FROM notes WHERE content = 'short-form is a command line journaling app' LIMIT 1;", &id))
	assert.NotNil(t, id)
	_, uuidErr := uuid.FromString(id)
	assert.Nil(t, uuidErr)

	var tags string
	assert.Nil(t, testing2.QueryOne(fmt.Sprintf("SELECT GROUP_CONCAT(tag) from note_tags where note_id='%s' LIMIT 1;", id), &tags))
	assert.EqualValues(t, "tag1,tag2", tags)

	testing2.CleanupDatabase(t)
}

// These tests interact with a local SQLite database.
func TestMain(m *testing.M) {
	// initialize test database and set singleton instance.
	database.InitializeDatabaseSingleton("./test.db")
	os.Exit(m.Run())
}