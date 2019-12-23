package database

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"github.com/ricanontherun/short-form/conf"
)

func NewDatabaseConnection() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", conf.ResolveDatabaseFilePath())

	if err != nil {
		return nil, err
	}

	return db, nil
}
