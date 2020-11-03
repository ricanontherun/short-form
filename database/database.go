package database

import (
	"database/sql"
	"fmt"
	"github.com/ricanontherun/short-form/logging"
	"github.com/ricanontherun/short-form/utils"
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

func NewDatabase(path string) Database {
	return &database{path, nil, nil}
}

func (database *database) GetConnection() *sql.DB {
	once.Do(func() {
		if exists, err := utils.EnsureFilePath(database.path); err != nil {
			panic(err)
		} else if !exists {
			logging.Debug(fmt.Sprintf("initialized new database %s", database.path))
		}

		db, err := NewDatabaseConnection(database.path)
		if err != nil {
			panic(err)
		}

		if database.postInit != nil {
			if err := database.postInit(db); err != nil {
				panic(err)
			}
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
