package data

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ricanontherun/short-form/conf"
	"github.com/ricanontherun/short-form/utils"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
	"log"
	"strings"
)

type Repository interface {
	WriteNote(note Note, secure bool) error
	SearchNotes(ctx Filters) (map[string]*Note, error)
	Close()
}

type repository struct {
	db     *leveldb.DB
	config conf.ShortFormConfig
}

const (
	prefixLogKey       = "l:"
	prefixContentKey   = "c:"
	prefixTagKey       = "t:"
	prefixMetaKey      = "m:"
	formatContentKey   = prefixContentKey + "%s"
	formatTagKey       = prefixTagKey + "%s:%s"
	formatTagPrefixKey = prefixTagKey + "%s:"
	formatMetaKey      = prefixMetaKey + "%s"
)

var (
	ErrInvalidDateRange = errors.New("invalid date range")
)

func makeLogKey(timestamp string) string {
	return timestamp
}

func makeContentKey(id string) string {
	return fmt.Sprintf(formatContentKey, id)
}

func makeTagKey(tag string, id string) string {
	return fmt.Sprintf(formatTagKey, tag, id)
}

func makeMetaKey(id string) string {
	return fmt.Sprintf(formatMetaKey, id)
}

func cleanLogKey(key string) string {
	return strings.TrimPrefix(key, prefixLogKey)
}

// Execute a function in the context of a leveldb transaction
func (repository repository) withTransaction(callback func(transaction *leveldb.Transaction) error) error {
	transaction, err := repository.db.OpenTransaction()
	if err != nil {
		return err
	}

	if err := callback(transaction); err != nil {
		transaction.Discard()
		return err
	}

	return transaction.Commit()
}

// Create a new note, possibly secured.
func (repository repository) WriteNote(note Note, secure bool) error {
	var preparedNote Note = note

	if secure {
		if encryptedNote, err := note.EncryptNote(repository.config.Secret); err != nil {
			return err
		} else {
			preparedNote = *encryptedNote
		}
	}

	return repository.withTransaction(func(transaction *leveldb.Transaction) error {
		if err := repository.setKeyValue(transaction, preparedNote.Timestamp, preparedNote.ID); err != nil {
			return err
		}

		if err := repository.setKeyValue(transaction, makeContentKey(preparedNote.ID), preparedNote.Content); err != nil {
			return err
		}

		// Write Note metadata.
		noteMetadata := NoteMeta{Tags: preparedNote.Tags, Secure: secure, Timestamp: preparedNote.Timestamp}
		if jsonBytes, err := json.Marshal(noteMetadata); err != nil {
			return err
		} else {
			if err := repository.setKeyValue(transaction, makeMetaKey(preparedNote.ID), string(jsonBytes)); err != nil {
				return err
			}
		}

		if len(preparedNote.Tags) > 0 {
			if err := repository.writeNoteTags(transaction, preparedNote.ID, preparedNote.Tags); err != nil {
				return err
			}
		}

		return nil
	})
}

// Batch insert tags for a note.
func (repository repository) writeNoteTags(transaction *leveldb.Transaction, id string, tags []string) error {
	batch := new(leveldb.Batch)

	for _, tag := range tags {
		batch.Put([]byte(makeTagKey(strings.ToLower(tag), id)), []byte("1"))
	}

	if err := transaction.Write(batch, nil); err != nil {
		return err
	}

	return nil
}

func (repository repository) SearchNotes(filters Filters) (map[string]*Note, error) {
	var notes map[string]*Note

	if filters.DateRange != nil {
		if n, err := repository.searchNotesByDate(filters.DateRange); err != nil {
			return nil, err
		} else {
			notes = n
		}
	} else { // Assume searching by tag.
		if n, err := repository.searchNotesByTag(filters.Tags); err != nil {
			return nil, err
		} else {
			notes = n
		}
	}

	filterOnTags := len(filters.Tags) > 0

	for id := range notes {
		if noteMetadataString, err := repository.getKeyValue(makeMetaKey(id)); err != nil {
			return nil, err
		} else {
			var noteMetadata NoteMeta
			if err := json.Unmarshal([]byte(noteMetadataString), &noteMetadata); err != nil {
				return nil, err
			}
			notes[id].Secure = noteMetadata.Secure
			notes[id].Timestamp = noteMetadata.Timestamp

			noteTags := noteMetadata.Tags

			if filterOnTags {
				if len(noteTags) == 0 {
					delete(notes, id)
					continue
				}

				match := false
				for _, tag := range filters.Tags {
					if utils.InArray(tag, noteTags) {
						match = true
						break
					}
				}

				if !match {
					delete(notes, id)
					continue
				}
			}

			for _, noteTag := range noteTags {
				notes[id].Tags = append(notes[id].Tags, noteTag)
			}
		}
	}

	encryptor := utils.MakeEncryptor(repository.config.Secret)
	for id := range notes {
		if content, err := repository.getNoteContent(id); err != nil {
			return nil, err
		} else {
			if notes[id].Secure {
				if bytes, err := encryptor.Decrypt([]byte(content)); err != nil {
					return nil, err
				} else {
					notes[id].Content = string(bytes)
				}
			} else {
				notes[id].Content = content
			}
		}
	}

	return notes, nil
}

func (repository repository) searchNotesByDate(dRange *DateRange) (map[string]*Note, error) {
	var notes = make(map[string]*Note)

	dateRange := util.Range{
		Start: []byte(makeLogKey(utils.ToUnixTimestampString(dRange.From))),
		Limit: []byte(makeLogKey(utils.ToUnixTimestampString(dRange.To))),
	}

	dateRangeIter := repository.db.NewIterator(&dateRange, nil)
	for dateRangeIter.Next() {
		id := string(dateRangeIter.Value())
		timestamp := cleanLogKey(string(dateRangeIter.Key()))

		notes[id] = &Note{
			ID:        id,
			Timestamp: timestamp,
			Tags:      []string{},
		}
	}
	dateRangeIter.Release()

	return notes, nil
}

func (repository repository) searchNotesByTag(tags []string) (map[string]*Note, error) {
	var notes = make(map[string]*Note)

	for _, tag := range tags {
		tagPrefixKey := makeTagKey(tag, "")
		tagPrefixIter := repository.db.NewIterator(util.BytesPrefix([]byte(tagPrefixKey)), nil)

		for tagPrefixIter.Next() {
			key := string(tagPrefixIter.Key())

			noteId := strings.Replace(key, tagPrefixKey, "", 1)
			notes[noteId] = &Note{
				ID: noteId,
			}
		}
	}

	return notes, nil
}

func (repository repository) getNoteContent(id string) (string, error) {
	if contentBytes, err := repository.getKeyValue(makeContentKey(id)); err != nil {
		return "", err
	} else {
		return contentBytes, nil
	}
}

func (repository repository) getKeyValue(key string) (string, error) {
	if bytes, err := repository.db.Get([]byte(key), nil); err != nil {
		return "", err
	} else {
		return string(bytes), nil
	}
}

func (repository repository) setKeyValue(transaction *leveldb.Transaction, key string, value string) error {
	return transaction.Put([]byte(key), []byte(value), nil)
}

func (repository repository) Close() {
	if repository.db != nil {
		if err := repository.db.Close(); err != nil {
			log.Printf("An error occured closing the database: %s\n", err.Error())
		}
	}
}

func NewRepository(configFile conf.ShortFormConfig) (Repository, error) {
	repository := repository{config: configFile}

	db, err := leveldb.OpenFile(conf.ResolveDataDirectory(), nil)
	if err != nil {
		return nil, err
	}
	repository.db = db

	return repository, nil
}
