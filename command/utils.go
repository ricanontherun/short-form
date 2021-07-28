package command

import (
	"bufio"
	"fmt"
	"github.com/ricanontherun/short-form/logging"
	"github.com/ricanontherun/short-form/output"
	"github.com/ricanontherun/short-form/utils"
	"github.com/urfave/cli/v2"
	"io"
	"os"
	"strings"
)

type parsedInput struct {
	title   string
	content string
	tags    []string
}

func getPrintOptionsFromContext(ctx *cli.Context) output.Options {
	options := output.NewOptions()

	search := strings.TrimSpace(strings.Join(ctx.Args().Slice(), " "))
	if len(search) > 0 {
		options.SearchContent = search
		options.SearchTags = []string{search}
	} else {
		options.SearchContent = ctx.String(flagContent)
		options.SearchTags = getTagsFromContext(ctx)
	}

	return options
}

// Prompt the user for input.
// Returns the trimmed, lowercase input.
func (handler handler) promptUser(message string) string {
	fmt.Print(message)
	return strings.TrimSpace(strings.ToLower(handler.inputController.GetString()))
}

// Prompt the user with a confirmation message.
// Returns whether the user answered 'y' or 'yes'
func (handler handler) confirmAction(message string) bool {
	return utils.SliceContainsElement(handler.promptUser(message+" [y/n]: "), []string{
		"yes",
		"y",
	})
}

func (handler handler) getContentFromInput(ctx *cli.Context) (*parsedInput, error) {
	stdinStat, err := os.Stdin.Stat()
	if err != nil {
		return nil, err
	}

	var content string
	var title string

	// Command args form the title.
	args := ctx.Args().Slice()
	if len(args) != 0 {
		title = strings.Join(args, " ")
	} else { // Each note requires a title.
		return nil, errMissingTitle
	}

	logging.Debug(fmt.Sprintf("title=%s", title))

	if stdinStat.Size() != 0 {
		logging.Debug("reading content from stdin")
		stdinReader := bufio.NewReader(os.Stdin)
		var stdinBuilder strings.Builder

		for {
			if r, _, readErr := stdinReader.ReadRune(); readErr != nil && readErr == io.EOF {
				break
			} else {
				if readErr != nil { // unexpected (non EOF) error.
					panic(readErr)
				}
				stdinBuilder.WriteRune(r)
			}
		}

		content = strings.TrimSuffix(stdinBuilder.String(), "\n")
	} else { // Prompt the user for input.
		logging.Debug("prompting user for input")
		content = handler.inputController.GetContentFromUserInput()
	}

	tags := getTagsFromContext(ctx)
	logging.Debug(fmt.Sprintf("title=%s, content=%s, tags=%v+", title, content, tags))
	return &parsedInput{title, content, tags}, nil
}

// Return a cleaned array of tags provided as --tags=t1,t2,t3, as ['t1', 't2', 't3']
func getTagsFromContext(c *cli.Context) []string {
	return cleanTagsFromString(c.String(flagTags))
}

func cleanTagsFromString(tagString string) []string {
	tags := utils.NewSet()

	for _, tag := range strings.Split(tagString, ",") {
		trimmed := strings.TrimSpace(tag)

		if len(trimmed) > 0 {
			tags.Add(trimmed)
		}
	}

	return tags.Entries()
}
