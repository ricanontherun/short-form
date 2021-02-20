package command

import (
	"fmt"
	"github.com/ricanontherun/short-form/output"
	"github.com/ricanontherun/short-form/utils"
	"github.com/urfave/cli/v2"
	"strings"
)

type parsedInput struct {
	content string
	tags    []string
}

func getPrintOptionsFromContext(ctx *cli.Context) output.Options {
	print_options := output.NewOptions()
	searchString := strings.TrimSpace(strings.Join(ctx.Args().Slice(), " "))

	if len(searchString) > 0 {
		print_options.SearchContent = searchString
		print_options.SearchTags = []string{searchString}
	} else {
		print_options.SearchContent = ctx.String(flagContent)
		print_options.SearchTags = ReadTagsFromContext(ctx)
	}

	return print_options
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
