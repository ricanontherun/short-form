package command

import (
	"bufio"
	"fmt"
	"github.com/ricanontherun/short-form/models"
	"github.com/ricanontherun/short-form/output"
	"github.com/ricanontherun/short-form/utils"
	"github.com/urfave/cli/v2"
	"io"
	"os"
	"strings"
)

type parsedInput struct {
	content string
	tags    []string
}

func getPrintOptionsFromContext(ctx *cli.Context) output.Options {
	return output.Options{
		SearchContent: ctx.String(flagContent),
		Detailed:      ctx.Bool(flagDetailed),
		Pretty:        ctx.Bool(flagPretty),
		SearchTags:    getTagsFromContext(ctx),
	}
}

// Prompt the user for input.
// Returns the trimmed, lowercase input.
func (handler handler) promptUser(message string) string {
	fmt.Print(message)
	return strings.TrimSpace(strings.ToLower(handler.inputController.GetString()))
}

// Prompt the user with a confirmation message.
// Returns whether the user answered 'y' or 'yes'
func (handler handler) makeUserConfirmAction(message string) bool {
	return utils.SliceContainsElement(handler.promptUser(message+" [y/n]: "), []string{
		"yes",
		"y",
	})
}

func getContentFromInput(ctx *cli.Context) (*parsedInput, error) {
	stdinStat, err := os.Stdin.Stat()
	if err != nil {
		return nil, err
	}

	var content string

	args := ctx.Args().Slice()
	if len(args) != 0 {
		fmt.Println("args: " + strings.Join(args, " "))
		content = strings.Join(args, " ")
	} else if stdinStat.Mode()&os.ModeCharDevice != 0 || stdinStat.Size() != 0 {
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
	}

	return &parsedInput{
		content: content,
		tags:    getTagsFromContext(ctx),
	}, nil
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

func getSearchFiltersFromContext(c *cli.Context) models.SearchFilters {
	return models.SearchFilters{
		Tags:    getTagsFromContext(c),
		Content: strings.TrimSpace(c.String(flagContent)),
	}
}
