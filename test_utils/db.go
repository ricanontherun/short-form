package test_utils

import (
	"github.com/ricanontherun/short-form/database"
	"github.com/stretchr/testify/assert"
	"testing"
)

func VerifyDatabaseIsEmpty(t *testing.T) {
	var noteCount int
	assert.Nil(t, QueryOne("SELECT COUNT(*) as c FROM notes;", &noteCount))
	assert.Equal(t, 0, noteCount)

	var tagCount int
	assert.Nil(t, QueryOne("SELECT COUNT(*) as c FROM note_tags;", &tagCount))
	assert.Equal(t, 0, tagCount)
}

func CleanupDatabase(t *testing.T) {
	assert.Nil(t, ExecuteQuery(t, "DELETE FROM notes;"))
	assert.Nil(t, ExecuteQuery(t, "DELETE FROM note_tags;"))
}

func ExecuteQuery(t *testing.T, query string) error {
	dbConnection := database.GetInstance().GetConnection()

	stmt, prepareErr := dbConnection.Prepare(query)
	assert.Nil(t, prepareErr)
	defer stmt.Close()

	_, queryErr := stmt.Exec()
	assert.Nil(t, queryErr)
	return nil
}

func QueryOne(query string, scans ...interface{}) error {
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
