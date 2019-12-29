package data

import "errors"

var (
	ErrNoteNotFound       = errors.New("note not found")
	ErrFailedToUpdateNote = errors.New("failed to update note")
)

const SQLInitializeDatabase = `
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

const SQLInsertNote = `
INSERT INTO notes (id, timestamp, content)
VALUES (?, ?, ?)
`

const SQLInsertTags = `INSERT INTO note_tags (note_id, tag) VALUES`

const SQLSearchForNotes = `
SELECT notes.id, notes.content, COALESCE(GROUP_CONCAT(DISTINCT note_tags.tag), "") as tags, notes.timestamp
FROM notes
LEFT JOIN note_tags
    ON note_tags.note_id = notes.id

-- WHERE clause
%s
COLLATE NOCASE
GROUP BY notes.id
ORDER BY notes.timestamp
`

const SQLUpdateNote = `UPDATE notes SET content=? WHERE id=?`

const SQLDeleteNote = "DELETE FROM notes WHERE notes.id = ?"

const SQLDeleteNoteTags = "DELETE FROM note_tags WHERE note_tags.note_id = ?"

const SQLGetNoteTags = `SELECT GROUP_CONCAT(DISTINCT tag) as tags FROM note_tags WHERE note_id = ? LIMIT 1`

const SQLGetNote = `
SELECT notes.id, timestamp, content
FROM notes
WHERE notes.id = ?
`

const SqlUpdateNote = `
UPDATE notes
SET
	content = ?
WHERE id = ?
`
