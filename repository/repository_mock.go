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
	repository.Called(note)
	return nil
}

func (repository *mockRepository) SearchNotes(ctx models.SearchFilters) ([]*models.Note, error) {
	repository.Called(ctx)
	return nil, nil
}

func (repository *mockRepository) DeleteNote(noteId string) error {
	return repository.Called(noteId).Error(0)
}

func (repository *mockRepository) LookupNote(noteId string) (*models.Note, error) {
	repository.Called(noteId)
	return nil, nil
}

func (repository *mockRepository) LookupNoteWithTags(noteId string) (*models.Note, error) {
	repository.Called(noteId)
	return nil, nil
}

func (repository *mockRepository) UpdateNote(note models.Note) error {
	return nil
}

func (repository *mockRepository) TagNote(note models.Note, tags []string) error {
	return nil
}

func (repository *mockRepository) Close() {
}
