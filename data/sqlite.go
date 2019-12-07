package data

import (
	"database/sql"
	"errors"
	_ "github.com/mattn/go-sqlite3"
	"github.com/ricanontherun/short-form/conf"
	"github.com/ricanontherun/short-form/utils"
	"log"
	"strings"
)

type sqlRepository struct {
	conn      *sql.DB
	encryptor utils.Encryptor
}

func NewSqlRepository(encryptor utils.Encryptor) (Repository, error) {
	db, err := sql.Open("sqlite3", conf.ResolveDatabaseFilePath())
	if err != nil {
		return nil, err
	}

	repository := sqlRepository{db, encryptor}

	// TODO: Initialize only if necessary (file doesn't exist, version mismatch.
	if err := repository.initialize(); err != nil {
		db.Close()
		return nil, errors.New("failed to initialize database: " + err.Error())
	}

	return repository, nil
}

func (repository sqlRepository) WriteNote(note Note) error {
	return repository.executeWithinTransaction(func(tx *sql.Tx) error {
		preparedNote, err := repository.prepareNote(note)
		if err != nil {
			return err
		}

		if err = repository.writeNote(tx, preparedNote); err != nil {
			return err
		}

		if note.Tags != nil && len(note.Tags) > 0 {
			if err := repository.writeNoteTags(tx, note.ID, note.Tags); err != nil {
				return err
			}
		}

		return nil
	})
}

func (repository sqlRepository) writeNote(tx *sql.Tx, note Note) error {
	noteInsertStatement, err := tx.Prepare(SQLInsertNote)
	if err != nil {
		return err
	}
	defer noteInsertStatement.Close()

	if _, err = noteInsertStatement.Exec(note.ID, note.Timestamp, note.Content, note.Secure); err != nil {
		return err
	}

	return nil
}

func (repository sqlRepository) writeNoteTags(tx *sql.Tx, noteId string, tags []string) error {
	sqlString := SQLInsertTags + " " + makeInsertValuesForTags(noteId, tags)
	tagInsertPreparedStatement, err := tx.Prepare(sqlString)

	if err != nil {
		return err
	}

	defer tagInsertPreparedStatement.Close()
	if _, err := tagInsertPreparedStatement.Exec(); err != nil {
		return err
	}

	return nil
}

func (repository sqlRepository) executeWithinTransaction(callback func(*sql.Tx) error) error {
	if transaction, err := repository.conn.Begin(); err != nil {
		return err
	} else {
		if err := callback(transaction); err != nil {
			if rollbackErr := transaction.Rollback(); rollbackErr != nil {
				return rollbackErr
			}

			return err
		}

		return transaction.Commit()
	}
}

func (repository sqlRepository) SearchNotes(ctx Filters) ([]Note, error) {
	stmt, err := repository.conn.Prepare(buildSearchQueryFromContext(ctx))
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rs, err := stmt.Query()
	if err != nil {
		return nil, err
	}

	var notes []Note
	for rs.Next() {
		var note Note
		var tagString string

		if err := rs.Scan(&note.ID, &note.Content, &tagString, &note.Timestamp, &note.Secure); err != nil {
			return nil, err
		}

		if len(tagString) > 0 {
			note.Tags = strings.Split(tagString, ",")
		}

		notes = append(notes, note)
	}

	return notes, nil
}

func (repository sqlRepository) prepareNote(note Note) (Note, error) {
	if note.Secure { // Clone + encrypt
		clone := note.Clone()

		if secureContentBytes, err := repository.encryptor.Encrypt([]byte(note.Content)); err != nil {
			return Note{}, err
		} else {
			clone.secured = true
			clone.Content = string(secureContentBytes)
		}

		return clone, nil
	} else {
		return note, nil
	}
}

// Initialize the database structure.
func (repository sqlRepository) initialize() error {
	if _, err := repository.conn.Exec(SQLInitializeDatabase); err != nil {
		return err
	}

	return nil
}

func (repository sqlRepository) Close() {
	if repository.conn != nil {
		if err := repository.conn.Close(); err != nil {
			log.Fatal(err)
		}
	}
}
