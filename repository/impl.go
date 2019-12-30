package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/ricanontherun/short-form/models"
	"strings"
)

type sqlRepository struct {
	conn *sql.DB
}

func NewSqlRepository(db *sql.DB) (Repository, error) {
	repository := sqlRepository{db}

	if err := repository.initialize(); err != nil {
		db.Close()
		return nil, errors.New("failed to initialize database: " + err.Error())
	}

	return repository, nil
}

func buildSearchQueryFromContext(ctx models.SearchFilters) string {
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

	if len(ctx.Content) > 0 {
		where = append(where, " notes.content LIKE '%"+ctx.Content+"%'")
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

func (repository sqlRepository) WriteNote(note models.Note) error {
	return repository.executeWithinTransaction(func(tx *sql.Tx) error {
		if err := repository.writeNote(tx, note); err != nil {
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

func (repository sqlRepository) writeNote(tx *sql.Tx, note models.Note) error {
	noteInsertStatement, err := tx.Prepare(SQLInsertNote)
	if err != nil {
		return err
	}
	defer noteInsertStatement.Close()

	if _, err = noteInsertStatement.Exec(note.ID, note.Timestamp, note.Content); err != nil {
		return err
	}

	return nil
}

func (repository sqlRepository) TagNote(note models.Note, tags []string) error {
	return repository.executeWithinTransaction(func(tx *sql.Tx) error {
		if err := repository.deleteNoteTags(tx, note.ID); err != nil {
			return err
		}

		if err := repository.writeNoteTags(tx, note.ID, tags); err != nil {
			return err
		}

		return nil
	})
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

func (repository sqlRepository) SearchNotes(ctx models.SearchFilters) ([]models.Note, error) {
	stmt, err := repository.conn.Prepare(buildSearchQueryFromContext(ctx))
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rs, err := stmt.Query()
	if err != nil {
		return nil, err
	}

	var notes []models.Note
	for rs.Next() {
		var note models.Note
		var tagString string

		if err := rs.Scan(&note.ID, &note.Content, &tagString, &note.Timestamp); err != nil {
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

		notes = append(notes, note)
	}

	return notes, nil
}

func (repository sqlRepository) DeleteNote(noteId string) error {
	return repository.executeWithinTransaction(func(tx *sql.Tx) error {
		stmt, err := tx.Prepare(SQLDeleteNote)
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

		return repository.deleteNoteTags(tx, noteId)
	})
}

func (repository sqlRepository) deleteNoteTags(tx *sql.Tx, noteId string) error {
	if stmt, err := tx.Prepare(SQLDeleteNoteTags); err != nil {
		return err
	} else {
		defer stmt.Close()
		if _, err = stmt.Exec(noteId); err != nil {
			return err
		}
	}

	return nil
}

// Update a note's content
func (repository sqlRepository) UpdateNoteContent(noteId string, content string) error {
	if stmt, err := repository.conn.Prepare(SQLUpdateNote); err != nil {
		return err
	} else {
		if updateResult, err := stmt.Exec(content, noteId); err != nil {
			return err
		} else if count, err := updateResult.RowsAffected(); err != nil {
			return err
		} else if count <= 0 {
			return ErrNoteNotFound
		}
	}

	return nil
}

// Get a single note from the database.
func (repository sqlRepository) GetNote(noteId string) (*models.Note, error) {
	if stmt, err := repository.conn.Prepare(SQLGetNote); err != nil {
		return nil, err
	} else {
		var note models.Note

		record := stmt.QueryRow(noteId)
		err := record.Scan(&note.ID, &note.Timestamp, &note.Content)
		if err != nil {
			if err == sql.ErrNoRows { // This is fine.
				return nil, ErrNoteNotFound
			}

			return nil, err
		}

		return &note, nil
	}
}

func (repository sqlRepository) UpdateNote(note models.Note) error {
	if stmt, err := repository.conn.Prepare(SqlUpdateNote); err != nil {
		return err
	} else {
		if results, err := stmt.Exec(note.Content, note.ID); err != nil {
			return err
		} else if rows, err := results.RowsAffected(); err != nil {
			return err
		} else if rows == 0 {
			return ErrFailedToUpdateNote
		}
	}

	return nil
}

// Initialize the database structure.
func (repository sqlRepository) initialize() error {
	if _, err := repository.conn.Exec(SQLInitializeDatabase); err != nil {
		return err
	}

	return nil
}