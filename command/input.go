package command

import (
	"bufio"
	"os"
	"strings"
)

type UserInput interface {
	GetString() string
}

type userInput struct{}

func NewUserInput() UserInput {
	return userInput{}
}

// Read a string from stdin, stopping at the first newline.
func (input userInput) GetString() string {
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	return strings.TrimSpace(strings.ToLower(text))
}
