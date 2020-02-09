package command

import (
	"fmt"
	"github.com/ricanontherun/short-form/models"
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

func getInputFromContext(ctx *cli.Context) parsedInput {
	return parsedInput{
		content: strings.Join(ctx.Args().Slice(), " "),
		tags:    getTagsFromContext(ctx),
	}
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
