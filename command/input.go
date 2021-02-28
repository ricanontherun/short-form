package command

import (
	"bufio"
	"fmt"
	"github.com/ricanontherun/short-form/utils"
	"os"
	"strings"
)

type UserInputController interface {
	GetString() string
	ConfirmAction(message string) bool
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

// Prompt the user for dto.
// Returns the trimmed, lowercase dto.
func (input userInput) promptUser(message string) string {
	fmt.Print(message)
	return strings.TrimSpace(strings.ToLower(input.GetString()))
}

func (input userInput) ConfirmAction(message string) bool {
	return utils.SliceContainsElement(input.promptUser(message+" [y/n]: "), []string{
		"yes",
		"y",
	})
}