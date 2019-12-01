package data

import (
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
	WriteNote(note Note) error
	SearchNotes(ctx Filters) (map[string]*Note, error)
	SearchNotesByDate(dateRange *DateRange) (map[string]*Note, error)
	GetNoteTags(id string) ([]string, error)
	GetNoteContent(id string) (string, error)
	Close()
}

type repository struct {
	db        *leveldb.DB
	config    conf.ShortFormConfig
	encryptor utils.Encryptor
}

func discardTransaction(err error, transaction *leveldb.Transaction) error {
	transaction.Discard()

	log.Printf("Failed to complete transaction: %s\n", err.Error())

	return err
}

const (
	prefixLogKey     = "l:"
	prefixContentKey = "c:"
	prefixTagKey     = "t:"
	prefixTagSetKey  = "ts:"
	formatLogKey     = prefixLogKey + "%s"
	formatContentKey = prefixContentKey + "%s"
	formatTagKey     = prefixTagKey + "%s:%s"
	formatTagSetKey  = prefixTagSetKey + "%s"
)

var (
	ErrInvalidDateRange = errors.New("invalid date range")
)

func makeLogKey(timestamp string) string {
	return timestamp
	//return fmt.Sprintf(formatLogKey, timestamp)
}

func makeContentKey(id string) string {
	return fmt.Sprintf(formatContentKey, id)
}

func makeTagKey(tag string, id string) string {
	return fmt.Sprintf(formatTagKey, tag, id)
}

func makeTagsSetKey(id string) string {
	return fmt.Sprintf(formatTagSetKey, id)
}

func cleanLogKey(key string) string {
	return strings.TrimPrefix(key, prefixLogKey)
}

func (repository repository) WriteNote(note Note) error {
	// TODO: Encryption
	transaction, err := repository.db.OpenTransaction()
	if err != nil {
		return err
	}

	logKey := makeLogKey(note.Timestamp)
	if err := transaction.Put([]byte(logKey), []byte(note.ID), nil); err != nil {
		return discardTransaction(err, transaction)
	}

	contentKey := []byte(makeContentKey(note.ID))
	if contentValue, err := repository.encrypt([]byte(note.Content)); err != nil {
		return discardTransaction(err, transaction)
	} else {
		if err := transaction.Put(contentKey, []byte(contentValue), nil); err != nil {
			return discardTransaction(err, transaction)
		}
	}

	// Index by tag.
	if len(note.Tags) > 0 {
		batch := new(leveldb.Batch)

		for _, tag := range note.Tags {
			key := makeTagKey(strings.ToLower(tag), note.ID)
			batch.Put([]byte(key), []byte("1"))
		}

		if err := transaction.Write(batch, nil); err != nil {
			return discardTransaction(err, transaction)
		}

		tagSetKey := makeTagsSetKey(note.ID)
		tagsString := strings.Join(note.Tags, ",")
		if err := transaction.Put([]byte(tagSetKey), []byte(tagsString), nil); err != nil {
			return discardTransaction(err, transaction)
		}
	}

	return transaction.Commit()
}

func (repository repository) SearchNotes(filters Filters) (map[string]*Note, error) {
	var notes map[string]*Note

	if filters.DateRange != nil {
		if n, err := repository.SearchNotesByDate(filters.DateRange); err != nil {
			return nil, err
		} else {
			notes = n
		}
	}

	// Add tags, filter by them.
	filterOnTags := len(filters.Tags) > 0
	for id := range notes {
		if noteTags, err := repository.GetNoteTags(id); err != nil {
			return nil, err
		} else {
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

	// Add/filter content
	// Add content to each note.
	// TODO: Refactor to batch GET operations.
	for id := range notes {
		if content, err := repository.GetNoteContent(id); err != nil {
			return nil, err
		} else {
			if original, err := repository.Decrypt([]byte(content)); err != nil {
				return nil, err
			} else {
				notes[id].Content = string(original)
			}

		}
	}

	return notes, nil
}

func (repository repository) SearchNotesByDate(dRange *DateRange) (map[string]*Note, error) {
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

func (repository repository) GetNoteTags(id string) ([]string, error) {
	tagString, err := repository.db.Get([]byte(makeTagsSetKey(id)), nil)

	if err != nil {
		if err == leveldb.ErrNotFound { // ok
			return nil, nil
		}

		return nil, err
	}

	var tags []string
	for _, tag := range strings.Split(string(tagString), ",") {
		tags = append(tags, strings.ToLower(tag))
	}

	return tags, nil
}

func (repository repository) GetNoteContent(id string) (string, error) {
	if contentBytes, err := repository.db.Get([]byte(makeContentKey(id)), nil); err != nil {
		return "", err
	} else {
		return string(contentBytes), nil
	}
}

func (repository repository) Close() {
	if repository.db != nil {
		if err := repository.db.Close(); err != nil {
			log.Printf("An error occured closing the database: %s\n", err.Error())
		}
	}
}

func (repository repository) encrypt(content []byte) ([]byte, error) {
	if !repository.config.Secure {
		return content, nil
	}

	return repository.encryptor.Encrypt(content)
}

func (repository repository) Decrypt(content []byte) ([]byte, error) {
	if !repository.config.Secure {
		return content, nil
	}

	if original, err := repository.encryptor.Decrypt(content); err != nil {
		return content, err
	} else {
		return original, nil
	}
}

func NewRepository(configFile conf.ShortFormConfig) (Repository, error) {
	repository := repository{config: configFile}

	db, err := leveldb.OpenFile(conf.ResolveDataDirectory(), nil)
	if err != nil {
		return nil, err
	}
	repository.db = db

	if configFile.Secure {
		repository.encryptor = utils.MakeEncryptor(configFile.Secret)
	}

	return repository, nil
}
