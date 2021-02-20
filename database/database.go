package database

import (
	"database/sql"
	"github.com/ricanontherun/short-form/utils"
	"log"
	"sync"
)

var once sync.Once

type postInitFunc func(db *sql.DB) error

type Database interface {
	SetPostInit(initFunc postInitFunc)
	GetConnection() *sql.DB
	Close()
}

type database struct {
	path     string
	database *sql.DB
	postInit func(*sql.DB) error
}

const sqlInitializeDatabase = `
CREATE TABLE IF NOT EXISTS notes
(
	id CHAR(16) not null
		constraint notes_pk
			primary key,
	timestamp TIMESTAMP not null,
	content TEXT not null
);

CREATE INDEX IF NOT EXISTS notes_content_index ON notes (content);

CREATE UNIQUE INDEX IF NOT EXISTS notes_id_uindex ON notes (id);

CREATE INDEX IF NOT EXISTS notes_timestamp_index ON notes (TIMESTAMP);

CREATE TABLE IF NOT EXISTS note_tags (note_id CHAR(16) NOT NULL,
                                      tag VARCHAR(50) NOT NULL);

CREATE INDEX IF NOT EXISTS note_tags_note_id_index ON note_tags (note_id);

CREATE INDEX IF NOT EXISTS note_tags_tag_index ON note_tags (tag);
`

// not thread safe, because it doesn't need to be.
// this is a single threaded program.
var singleton *database

func InitializeDatabaseSingleton(path string) {
	singleton = &database{path, nil, nil}
}

func GetInstance() *database {
	return singleton
}

func (database *database) GetConnection() *sql.DB {
	once.Do(func() {
		if exists, err := utils.EnsureFilePath(database.path); err != nil {
			panic(err)
		} else if !exists {
			log.Println("created new database file " + database.path)
		}

		db, err := NewDatabaseConnection(database.path)
		if err != nil {
			panic(err)
		}

		if _, err := db.Exec(sqlInitializeDatabase); err != nil {
			panic(err)
		}

		database.database = db
	})

	return database.database
}

func (database *database) SetPostInit(call postInitFunc) {
	database.postInit = call
}

func (database *database) Close() {
	if database.database != nil {
		database.database.Close()
	}
}
