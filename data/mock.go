package data

import "github.com/stretchr/testify/mock"

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

func (mockRepository) SearchNotes(ctx Filters) ([]Note, error) {
	return nil, nil
}

func (mockRepository) DeleteNote(noteId string) error {
	return nil
}

func (mockRepository) GetNote(noteId string) (*Note, error) {
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
