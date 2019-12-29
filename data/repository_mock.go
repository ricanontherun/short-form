package data

import (
	"github.com/stretchr/testify/mock"
)

type mockRepository struct {
	mock.Mock
}

func NewMockRepository() mockRepository {
	return mockRepository{}
}

func (repository *mockRepository) WriteNote(note Note) error {
	args := repository.Called(note)
	return args.Error(0)
}

func (repository *mockRepository) SearchNotes(ctx Filters) ([]Note, error) {
	args := repository.Called(ctx)
	return []Note{}, args.Error(0)
}

func (repository *mockRepository) DeleteNote(noteId string) error {
	return repository.Called(noteId).Error(0)
}

func (repository *mockRepository) GetNote(noteId string) (*Note, error) {
	repository.Called(noteId)
	return nil, nil
}

func (mockRepository) UpdateNote(note Note) error {
	return nil
}

func (mockRepository) TagNote(note Note, tags []string) error {
	return nil
}

func (mockRepository) Close() {
}
