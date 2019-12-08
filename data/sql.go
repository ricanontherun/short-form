package data

import (
	"fmt"
	"strings"
)

const SQLInitializeDatabase = `
CREATE TABLE IF NOT EXISTS notes
(
	id CHAR(16) not null
		constraint notes_pk
			primary key,
	timestamp TIMESTAMP not null,
	content TEXT not null,
	secure int not null
);

CREATE INDEX IF NOT EXISTS notes_content_index ON notes (content);

CREATE UNIQUE INDEX IF NOT EXISTS notes_id_uindex ON notes (id);

CREATE INDEX IF NOT EXISTS notes_timestamp_index ON notes (TIMESTAMP);

CREATE TABLE IF NOT EXISTS note_tags (note_id CHAR(16) NOT NULL,
                                      tag VARCHAR(50) NOT NULL);

CREATE INDEX IF NOT EXISTS note_tags_note_id_index ON note_tags (note_id);

CREATE INDEX IF NOT EXISTS note_tags_tag_index ON note_tags (tag);
`

const SQLInsertNote = `
INSERT INTO notes (id, timestamp, content, secure)
VALUES (?, ?, ?, ?)
`

const SQLInsertTags = `INSERT INTO note_tags (note_id, tag) VALUES`

const SQLSearchForNotes = `
SELECT notes.id, notes.content, COALESCE(GROUP_CONCAT(distinct note_tags.tag), "") as tags, notes.timestamp, notes.secure
FROM notes
LEFT JOIN note_tags
    ON note_tags.note_id = notes.id
-- WHERE clause
%s

-- TODO: The fact that this having doesn't work as expected seems like a bug.
-- HAVING tag_notes.tag IN (?)
GROUP BY notes.id
ORDER BY notes.timestamp
`

const SQLDeleteNote = "DELETE FROM notes WHERE notes.id = ?"
const SQLDeleteNoteTags = "DELETE FROM note_tags WHERE note_tags.note_id = ?"
const SQLGetNote = `
SELECT notes.id, notes.content, COALESCE(GROUP_CONCAT(distinct note_tags.tag), "") as tags, notes.timestamp, notes.secure
FROM notes
LEFT JOIN note_tags
    ON note_tags.note_id = notes.id
WHERE notes.id = ?
`

const SQLGetNoteTags = `SELECT GROUP_CONCAT(tag) as tags FROM note_tags WHERE note_id = ?`

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

	whereClauseString := "WHERE " + strings.Join(where, "AND")
	return fmt.Sprintf(SQLSearchForNotes, whereClauseString)
}

func makeInsertValuesForTags(noteId string, tags []string) string {
	inserts := make([]string, 0, len(tags))

	for _, tag := range tags {
		inserts = append(inserts, fmt.Sprintf("('%s', '%s')", noteId, tag))
	}

	return strings.Join(inserts, ",")
}
