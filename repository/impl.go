package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/ricanontherun/short-form/database"
	"github.com/ricanontherun/short-form/models"
	"strings"
)

type SearchTypeType string

const (
	SearchTypeAnd SearchTypeType = "AND"
	SearchTypeOr  SearchTypeType = "OR"
)

type sqlRepository struct {
	db database.Database
}

// Delete all the notes under a tag
func (repository sqlRepository) DeleteNoteByTag(tag string) error {
	return repository.transaction(func(tx *sql.Tx) error {
		if stmt, err := tx.Prepare(sqlDeleteNotesByTag); err != nil {
			return err
		} else {
			if _, err = stmt.Exec(tag); err != nil {
				return err
			}
		}

		if stmt, err := tx.Prepare(sqlDeleteTags); err != nil {
			return err
		} else {
			if _, err := stmt.Exec(tag); err != nil {
				return err
			}
		}

		return nil
	})
}

func NewSqlRepository(db database.Database) (Repository, error) {
	repository := sqlRepository{db}

	repository.db.SetPostInit(func(db *sql.DB) error {
		if err := repository.initialize(db); err != nil {
			return errors.New("failed to initialize database: " + err.Error())
		}

		return nil
	})

	return repository, nil
}

func buildSearchQueryFromFilters(searchFilters *models.SearchFilters) string {
	var where []string

	var searchType SearchTypeType
	if len(searchFilters.String) > 0 {
		searchType = SearchTypeOr
	} else {
		searchType = SearchTypeAnd
	}

	if searchFilters.DateRange != nil {
		filter := fmt.Sprintf(
			" timestamp BETWEEN datetime('%s') and datetime('%s') ",
			searchFilters.DateRange.From.Format("2006-01-02 15:04:05"),
			searchFilters.DateRange.To.Format("2006-01-02 15:04:05"),
		)

		where = append(where, filter)
	}

	if len(searchFilters.String) > 0 {
		where = append(where, prepareTagsFilter([]string{searchFilters.String}))
	} else if len(searchFilters.Tags) > 0 {
		where = append(where, prepareTagsFilter(searchFilters.Tags))
	}

	if len(searchFilters.String) > 0 {
		where = append(where, " notes.content LIKE '%"+searchFilters.String+"%'")
	} else if len(searchFilters.Content) > 0 {
		where = append(where, " notes.content LIKE '%"+searchFilters.Content+"%'")
	}

	whereClauseString := ""
	if len(where) > 0 {
		whereClauseString = "WHERE " + strings.Join(where, string(searchType))
	}

	return fmt.Sprintf(sqlSearchNotes, whereClauseString)
}

func prepareTagsFilter(tags []string) string {
	quotedTags := make([]string, 0, len(tags))
	for _, tag := range tags {
		quotedTags = append(quotedTags, "'"+tag+"'")
	}

	return fmt.Sprintf(" note_tags.tag in (%s)", strings.Join(quotedTags, ","))
}

func makeInsertValuesForTags(noteId string, tags []string) string {
	inserts := make([]string, 0, len(tags))

	for _, tag := range tags {
		inserts = append(inserts, fmt.Sprintf("('%s', '%s')", noteId, tag))
	}

	return strings.Join(inserts, ",")
}

func (repository sqlRepository) WriteNote(note models.Note) error {
	return repository.transaction(func(tx *sql.Tx) error {
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
	noteInsertStatement, err := tx.Prepare(sqlInsertNote)
	if err != nil {
		return err
	}
	defer noteInsertStatement.Close()

	if _, err = noteInsertStatement.Exec(note.ID, note.Timestamp, note.Title, note.Content); err != nil {
		return err
	}

	return nil
}

func (repository sqlRepository) TagNote(note models.Note, tags []string) error {
	return repository.transaction(func(tx *sql.Tx) error {
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
	sqlString := sqlInsertNoteTags + " " + makeInsertValuesForTags(noteId, tags)
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

func (repository sqlRepository) transaction(callback func(*sql.Tx) error) error {
	if transaction, err := repository.db.GetConnection().Begin(); err != nil {
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

func (repository sqlRepository) SearchNotes(ctx *models.SearchFilters) ([]*models.Note, error) {
	stmt, err := repository.db.GetConnection().Prepare(buildSearchQueryFromFilters(ctx))
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rs, err := stmt.Query()
	if err != nil {
		return nil, err
	}

	var notes []*models.Note
	for rs.Next() {
		var note models.Note
		var tagString string

		if err := rs.Scan(&note.ID, &note.Title, &note.Content, &tagString, &note.Timestamp); err != nil {
			return nil, err
		}

		if len(tagString) > 0 {
			if tags, err := repository.getNoteTags(note.ID); err != nil {
				return nil, err
			} else {
				note.Tags = tags
			}
		}

		notes = append(notes, &note)
	}

	return notes, nil
}

func (repository sqlRepository) getNoteTags(noteId string) ([]string, error) {
	if stmt, err := repository.db.GetConnection().Prepare(sqlGetNoteTags); err != nil {
		return nil, err
	} else {
		defer stmt.Close()

		var tagString string
		if err := stmt.QueryRow(noteId).Scan(&tagString); err != nil {
			return nil, err
		} else {
			return strings.Split(tagString, ","), nil
		}
	}
}

func (repository sqlRepository) DeleteNote(noteId string) error {
	return repository.transaction(func(tx *sql.Tx) error {
		stmt, err := tx.Prepare(sqlDeleteNote)
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
	if stmt, err := tx.Prepare(sqlDeleteNoteTags); err != nil {
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
	if stmt, err := repository.db.GetConnection().Prepare(sqlUpdateNote); err != nil {
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
func (repository sqlRepository) LookupNote(noteId string) (*models.Note, error) {
	return repository.getNote(noteId, false)
}

func (repository sqlRepository) LookupNoteWithTags(noteId string) (*models.Note, error) {
	return repository.getNote(noteId, true)
}

func (repository sqlRepository) getNote(noteId string, withTags bool) (*models.Note, error) {
	if stmt, err := repository.db.GetConnection().Prepare(sqlGetNote); err != nil {
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

		if withTags {
			if tags, err := repository.getNoteTags(noteId); err != nil {
				return nil, err
			} else {
				note.Tags = tags
			}
		}

		return &note, nil
	}
}

func (repository sqlRepository) UpdateNote(note models.Note) error {
	if stmt, err := repository.db.GetConnection().Prepare(sqlUpdateNoteContent); err != nil {
		return err
	} else {
		defer stmt.Close()
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

func (repository sqlRepository) LookupNotesByShortId(shortId string) ([]*models.Note, error) {
	if stmt, err := repository.db.GetConnection().Prepare(sqlSearchByShortId); err != nil {
		return nil, err
	} else {
		defer stmt.Close()

		rs, err := stmt.Query(shortId)
		if err != nil {
			return nil, err
		}

		var notes []*models.Note
		for rs.Next() {
			var note models.Note

			if scanErr := rs.Scan(&note.ID, &note.Timestamp, &note.Content); scanErr != nil {
				return nil, scanErr
			}

			if tags, err := repository.getNoteTags(note.ID); err != nil {
				return nil, err
			} else {
				note.Tags = tags
			}

			notes = append(notes, &note)
		}
		return notes, nil
	}
}

// Initialize the database structure.
func (repository sqlRepository) initialize(db *sql.DB) error {
	if _, err := db.Exec(sqlInitializeDatabase); err != nil {
		return err
	}

	// Check for db version table, make any necessary migrations.

	return nil
}
