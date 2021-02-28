package command

import "github.com/stretchr/testify/mock"

// Create a mock user dto controller to be dependency injected into
// handler for both delete and edit note flows.
type mockInput struct {
	mock.Mock
}

func NewMockInput() *mockInput {
	return &mockInput{}
}

func (input *mockInput) GetString() string {
	args := input.Called()
	return args.Get(0).(string)
}
