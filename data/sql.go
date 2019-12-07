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
                                      tag VARCHAR(50) NOT NULL,
                                      active tinyint DEFAULT 1 NOT NULL);

CREATE INDEX IF NOT EXISTS note_tags_active_index ON note_tags (active);

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

GROUP BY notes.id

-- Optional HAVING clause
%s

ORDER BY notes.timestamp
`

func buildSearchQueryFromContext(ctx Filters) string {
	var where string
	var having string

	if ctx.DateRange != nil {
		where = fmt.Sprintf(
			"WHERE timestamp BETWEEN datetime('%s') and datetime('%s')",
			ctx.DateRange.From.Format("2006-01-02 15:04:05"),
			ctx.DateRange.To.Format("2006-01-02 15:04:05"),
		)
	}

	if len(ctx.Tags) > 0 {
		quotedTags := make([]string, 0, len(ctx.Tags))
		for _, tag := range ctx.Tags {
			quotedTags = append(quotedTags, "'"+tag+"'")
		}

		having = fmt.Sprintf("HAVING note_tags.tag in (%s)", strings.Join(quotedTags, ","))
	}

	return fmt.Sprintf(SQLSearchForNotes, where, having)
}

func makeInsertValuesForTags(noteId string, tags []string) string {
	inserts := make([]string, 0, len(tags))

	for _, tag := range tags {
		inserts = append(inserts, fmt.Sprintf("('%s', '%s')", noteId, tag))
	}

	return strings.Join(inserts, ",")
}
