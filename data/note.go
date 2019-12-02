package data

import (
	"github.com/ricanontherun/short-form/utils"
)

type Note struct {
	ID        string
	Tags      []string
	Content   string
	Timestamp string
	Secure    bool
}

type NoteMeta struct {
	Tags   []string `json:"tags"`
	Secure bool     `json:"secure"`
}

// Return a secure copy of this note.
func (note Note) EncryptNote(secret string) (*Note, error) {
	encryptor := utils.MakeEncryptor(secret)

	noteCopy := &Note{
		ID:        note.ID,
		Timestamp: note.Timestamp,
		Tags:      note.Tags,
	}

	if contentBytes, err := encryptor.Encrypt([]byte(note.Content)); err != nil {
		return nil, err
	} else {
		noteCopy.Content = string(contentBytes)
	}

	return noteCopy, nil
}
