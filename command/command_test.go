package command

import (
	"fmt"
	"github.com/ricanontherun/short-form/database"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func cleanup(t *testing.T) {
	assert.Nil(t, executeQuery(t, "DELETE FROM notes;"))
	assert.Nil(t, executeQuery(t, "DELETE FROM note_tags;"))
}

func executeQuery(t *testing.T, query string) error {
	dbConnection := database.GetInstance().GetConnection()

	stmt, prepareErr := dbConnection.Prepare(query)
	assert.Nil(t, prepareErr)
	defer stmt.Close()

	_, queryErr := stmt.Exec()
	assert.Nil(t, queryErr)
	return nil
}

func queryOne(query string, scans ...interface{}) error {
	dbConnection := database.GetInstance().GetConnection()

	stmt, prepareErr := dbConnection.Prepare(query)
	if prepareErr != nil {
		return prepareErr
	}
	defer stmt.Close()

	resultSet, queryErr := stmt.Query()
	if queryErr != nil {
		return queryErr
	}
	defer resultSet.Close()

	resultSet.Next()
	if scanErr := resultSet.Scan(scans...); scanErr != nil {
		//t.Error(scanErr)
	}

	return nil
}

func TestWriteCommand_Execute(t *testing.T) {
	writeCommand := NewWriteCommand(&Note{
		Content: "ok",
		Tags: []string{"tag1", "tag2"},
	})

	assert.Nil(t, writeCommand.Execute())

	var id string
	assert.Nil(t, queryOne("SELECT id FROM notes WHERE content = 'ok' LIMIT 1;", &id))
	assert.NotNil(t, id)
	_, uuidErr := uuid.FromString(id)
	assert.Nil(t, uuidErr)

	var tags string
	assert.Nil(t, queryOne(fmt.Sprintf("SELECT GROUP_CONCAT(tag) from note_tags where note_id='%s' LIMIT 1;", id), &tags))
	assert.EqualValues(t, "tag1,tag2", tags)

	cleanup(t)
}

// These tests interact with a local SQLite database.
func TestMain(m *testing.M) {
	// remove test database.

	// initialize test database and set singleton instance.
	database.InitializeDatabaseSingleton("./test.db")
	database.GetInstance().GetConnection()
	os.Exit(m.Run())
}