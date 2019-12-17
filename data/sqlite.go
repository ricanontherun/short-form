package data

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/ricanontherun/short-form/conf"
	"github.com/ricanontherun/short-form/utils"
	"log"
	"strings"
)

var (
	ErrNoteNotFound = errors.New("not not found")
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

func buildSearchQueryFromContext(ctx Filters) string {
	var where []string

	if ctx.DateRange != nil {
		filter := fmt.Sprintf(
			" timestamp BETWEEN datetime('%s') and datetime('%s') ",
			ctx.DateRange.From.Format("2006-01-02 15:04:05"),
			ctx.DateRange.To.Format("2006-01-02 15:04:05"),
		)

		where = append(where, filter)
	}

	if len(ctx.Tags) > 0 {
		quotedTags := make([]string, 0, len(ctx.Tags))
		for _, tag := range ctx.Tags {
			quotedTags = append(quotedTags, "'"+tag+"'")
		}

		filter := fmt.Sprintf(" note_tags.tag in (%s)", strings.Join(quotedTags, ","))

		where = append(where, filter)
	}

	whereClauseString := ""
	if len(where) > 0 {
		whereClauseString = "WHERE " + strings.Join(where, "AND")
	}

	return fmt.Sprintf(SQLSearchForNotes, whereClauseString)
}

func makeInsertValuesForTags(noteId string, tags []string) string {
	inserts := make([]string, 0, len(tags))

	for _, tag := range tags {
		inserts = append(inserts, fmt.Sprintf("('%s', '%s')", noteId, tag))
	}

	return strings.Join(inserts, ",")
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
			if stmt, err := repository.conn.Prepare(SQLGetNoteTags); err != nil {
				return nil, err
			} else {
				defer stmt.Close()

				var tagString string

				if err := stmt.QueryRow(note.ID).Scan(&tagString); err != nil {
					return nil, err
				} else {
					note.Tags = strings.Split(tagString, ",")
				}
			}
		}

		// Filter by content.
		if len(ctx.Content) > 0 {
			content := note.Content

			if note.Secure {
				if insecureBytes, err := repository.encryptor.Decrypt([]byte(content)); err != nil {
					return nil, err
				} else {
					content = string(insecureBytes)
				}
			}

			if !strings.Contains(content, ctx.Content) {
				continue
			}
		}

		notes = append(notes, note)
	}

	return notes, nil
}

func (repository sqlRepository) DeleteNote(noteId string) error {
	return repository.executeWithinTransaction(func(tx *sql.Tx) error {
		stmt, err := repository.conn.Prepare(SQLDeleteNote)
		if err != nil {
			return err
		}
		defer stmt.Close()

		if rs, err := stmt.Exec(noteId); err != nil {
			return err
		} else {
			numDeleted, err := rs.RowsAffected()
			if err != nil {
				return err
			}

			if numDeleted <= 0 {
				return ErrNoteNotFound
			}
		}

		stmt, err = repository.conn.Prepare(SQLDeleteNoteTags)
		if err != nil {
			return err
		}
		defer stmt.Close()
		if _, err = stmt.Exec(noteId); err != nil {
			return err
		}

		return nil
	})
}

func (repository sqlRepository) UpdateNoteContent(noteId string, content string) error {
	if stmt, err := repository.conn.Prepare(SQLUpdateNote); err != nil {
		return err
	} else {
		defer stmt.Close()
		if _, err := stmt.Exec(noteId, content); err != nil {
			return err
		}
	}

	return nil
}

func (repository sqlRepository) GetNote(noteId string) (Note, error) {
	return Note{}, nil
}

func (repository sqlRepository) prepareNote(note Note) (Note, error) {
	if note.Secure { // Clone + encrypt
		clone := note.Clone()

		if secureContentBytes, err := repository.encryptor.Encrypt([]byte(note.Content)); err != nil {
			return Note{}, err
		} else {
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
