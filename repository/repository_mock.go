package repository

import (
	"github.com/ricanontherun/short-form/dto"
	"github.com/stretchr/testify/mock"
)

type mockRepository struct {
	mock.Mock
}

func (repository *mockRepository) DeleteNoteByTags(tags []string) error {
	panic("implement me")
}

func (repository *mockRepository) GetNoteCountByTags(tags []string) (uint64, error) {
	panic("implement me")
}

func NewMockRepository() mockRepository {
	return mockRepository{}
}

func (repository *mockRepository) WriteNote(note dto.Note) error {
	return repository.Called(note).Error(0)
}

func (repository *mockRepository) SearchNotes(ctx *dto.SearchFilters) ([]*dto.Note, error) {
	args := repository.Called(ctx)

	notesArgs := args.Get(0)

	if notesArgs != nil {
		return notesArgs.([]*dto.Note), args.Error(0)
	} else {
		return nil, args.Error(0)
	}
}

func (repository *mockRepository) DeleteNote(noteId string) error {
	return repository.Called(noteId).Error(0)
}

func (repository *mockRepository) LookupNote(noteId string) (*dto.Note, error) {
	args := repository.Called(noteId)
	return args.Get(0).(*dto.Note), args.Error(1)
}

func (repository *mockRepository) LookupNoteWithTags(noteId string) (*dto.Note, error) {
	args := repository.Called(noteId)
	return args.Get(0).(*dto.Note), args.Error(1)
}

func (repository *mockRepository) UpdateNote(note dto.Note) error {
	return nil
}

func (repository *mockRepository) TagNote(note dto.Note, tags []string) error {
	return nil
}

func (repository *mockRepository) LookupNotesByShortId(shortId string) ([]*dto.Note, error) {
	args := repository.Called(shortId)
	notes := args.Get(0)

	if notes != nil {
		return notes.([]*dto.Note), args.Error(1)
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
