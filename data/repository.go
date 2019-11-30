package data

import (
	"errors"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
	"log"
	"short-form/conf"
	"short-form/search"
	"short-form/utils"
	"strings"
)

type Repository interface {
	WriteNote(note Note) error
	SearchNotes(ctx search.Filters) ([]Note, error)
	Close()
}

type repository struct {
	metaDb *leveldb.DB
	logDb  *leveldb.DB
}

func discardTransactions(err error, ts ...*leveldb.Transaction) error {
	for _, transaction := range ts {
		transaction.Discard()
	}

	log.Printf("Failed to complete transaction: %s\n", err.Error())

	return err
}

const (
	prefixLogKey     = "l:"
	prefixContentKey = "c:"
	prefixTagKey     = "t:"
	formatLogKey     = prefixLogKey + "%s"
	formatContentKey = prefixContentKey + "%s"
	formatTagKey     = prefixTagKey + "%s:%s"
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

func cleanLogKey(key string) string {
	return strings.TrimPrefix(key, prefixLogKey)
}

func (repository repository) WriteNote(note Note) error {
	// TODO: Encryption

	metaTransaction, err := repository.metaDb.OpenTransaction()
	if err != nil {
		return err
	}

	logTransaction, err := repository.logDb.OpenTransaction()
	if err != nil {
		return err
	}

	logKey := makeLogKey(note.Timestamp)
	if err := logTransaction.Put([]byte(logKey), []byte(note.ID), nil); err != nil {
		return discardTransactions(err, metaTransaction, logTransaction)
	}

	contentKey := makeContentKey(note.ID)
	if err := metaTransaction.Put([]byte(contentKey), []byte(note.Content), nil); err != nil {
		return discardTransactions(err, metaTransaction, logTransaction)
	}

	if len(note.Tags) > 0 {
		batch := new(leveldb.Batch)

		for _, tag := range note.Tags {
			key := makeTagKey(tag, note.ID)
			batch.Put([]byte(key), []byte("1"))
		}

		if err := metaTransaction.Write(batch, nil); err != nil {
			return discardTransactions(err, metaTransaction, logTransaction)
		}
	}

	// Multiple independent transactions won't work.
	// Can we get all the appropriate key formats working together
	// in a single database?
	logTransaction.Commit()
	metaTransaction.Commit()
	return nil
}

func (repository repository) SearchNotes(ctx search.Filters) ([]Note, error) {
	// Search by tag
	// First, search by tag prefix
	// Second, search by date range
	// Lastly, search by note content
	var notes []Note

	if ctx.DateRange != nil {
		dateRange := util.Range{
			Start: []byte(makeLogKey(utils.ToUnixTimestampString(ctx.DateRange.From))),
		}

		dateRangeIter := repository.logDb.NewIterator(&dateRange, nil)
		for dateRangeIter.Next() {
			timestamp := cleanLogKey(string(dateRangeIter.Key()))
			notes = append(notes, Note{
				ID:        string(dateRangeIter.Value()),
				Timestamp: timestamp,
			})
		}

		dateRangeIter.Release()
	}

	// Search by tag.

	return notes, nil
}

func (repository repository) Close() {
	if repository.logDb != nil {
		if err := repository.logDb.Close(); err != nil {
			log.Printf("An error occured closing the database: %s\n", err.Error())
		}
	}

	if repository.metaDb != nil {
		if err := repository.metaDb.Close(); err != nil {
			log.Printf("An error occured closing the database: %s\n", err.Error())
		}
	}
}

func NewRepository() (Repository, error) {
	repository := repository{}

	logDb, err := leveldb.OpenFile(conf.ResolveLogDirectory(), nil)
	if err != nil {
		return nil, err
	}
	repository.logDb = logDb

	db, err := leveldb.OpenFile(conf.ResolveDataDirectory(), nil)
	if err != nil {
		return nil, err
	}
	repository.metaDb = db

	return repository, nil
}
