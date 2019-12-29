package handler

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

func (input userInput) GetString() string {
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	return strings.TrimSpace(strings.ToLower(text))
}
