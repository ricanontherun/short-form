package repository

import (
	"github.com/ricanontherun/short-form/models"
	"github.com/stretchr/testify/mock"
)

type mockRepository struct {
	mock.Mock
}

func NewMockRepository() mockRepository {
	return mockRepository{}
}

func (repository *mockRepository) WriteNote(note models.Note) error {
	return repository.Called(note).Error(0)
}

func (repository *mockRepository) SearchNotes(ctx *models.SearchFilters) ([]*models.Note, error) {
	args := repository.Called(ctx)

	notesArgs := args.Get(0)

	if notesArgs != nil {
		return notesArgs.([]*models.Note), args.Error(0)
	} else {
		return nil, args.Error(0)
	}
}

func (repository *mockRepository) DeleteNote(noteId string) error {
	return repository.Called(noteId).Error(0)
}

func (repository *mockRepository) LookupNote(noteId string) (*models.Note, error) {
	args := repository.Called(noteId)
	return args.Get(0).(*models.Note), args.Error(1)
}

func (repository *mockRepository) LookupNoteWithTags(noteId string) (*models.Note, error) {
	args := repository.Called(noteId)
	return args.Get(0).(*models.Note), args.Error(1)
}

func (repository *mockRepository) UpdateNote(note models.Note) error {
	return nil
}

func (repository *mockRepository) TagNote(note models.Note, tags []string) error {
	return nil
}

func (repository *mockRepository) LookupNotesByShortId(shortId string) ([]*models.Note, error) {
	args := repository.Called(shortId)
	notes := args.Get(0)

	if notes != nil {
		return notes.([]*models.Note), args.Error(1)
	} else {
		return nil, args.Error(1)
	}
}

func (repository *mockRepository) DeleteNoteByTag(tag string) error {
	args := repository.Called(tag)

	return args.Error(0)
}

func (repository *mockRepository) Close() {
}
