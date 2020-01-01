package command

import (
	"bufio"
	"os"
	"strings"
)

type UserInputController interface {
	GetString() string
}

type userInput struct{}

func NewUserInputController() UserInputController {
	return userInput{}
}

// Read a string from stdin, stopping at the first newline.
func (input userInput) GetString() string {
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	return strings.TrimSpace(strings.ToLower(text))
}
