package command

import (
	"bufio"
	"os"
	"strings"
)

type UserInputController interface {
	GetString() string
	GetContentFromUserInput() string
}

type userInput struct{}

func NewUserInputController() UserInputController {
	return userInput{}
}

// GetContentFromUserInput Read a string from stdin, stopping at the first newline.
func (input userInput) GetContentFromUserInput() string {
	reader := bufio.NewReader(os.Stdin)

	contentString := ""
	consecutiveNewlines := 0

	// while we receive non-consecutive newlines, append them to contentString
	// as soon as we receive two consecutive newlines, terminate the loop.
	for {
		inputLine, _ := reader.ReadString('\n')

		if inputLine == "\n" {
			consecutiveNewlines++

			// terminate input loop after second newline
			if consecutiveNewlines == 2 {
				break
			}
		} else {
			consecutiveNewlines = 0
			contentString += inputLine
		}
	}

	return strings.TrimSuffix(strings.ToLower(contentString), "\n")
}

func (input userInput) GetString() string {
	reader := bufio.NewReader(os.Stdin)
	text, _ := reader.ReadString('\n')
	return strings.TrimSpace(strings.ToLower(text))
}
