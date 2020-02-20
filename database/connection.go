package database

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

// NewDatabaseConnection creates a new database connection.
func NewDatabaseConnection(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path)

	if err != nil {
		return nil, err
	}

	return db, nil
}
