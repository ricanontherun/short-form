package handler

import (
	"bufio"
	"fmt"
	"github.com/ricanontherun/short-form/data"
	"github.com/ricanontherun/short-form/utils"
	"github.com/urfave/cli/v2"
	"os"
	"strings"
)

type parsedInput struct {
	content string
	tags    []string
}

type printOptions struct {
	highlight string
	detailed  bool
}

func getPrintOptionsFromContext(ctx *cli.Context) printOptions {
	return printOptions{
		highlight: ctx.String("content"),
		detailed:  ctx.Bool("detailed"),
	}
}

func promptUser(message string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print(message)
	text, _ := reader.ReadString('\n')
	return strings.TrimSpace(strings.ToLower(text))
}

func makeUserConfirmAction(message string) bool {
	return utils.InArray(promptUser(message+" [y/n]: "), []string{
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
	return cleanTagsFromString(c.String("tags"))
}

func cleanTagsFromString(tagString string) []string {
	tags := utils.NewSet()

	for _, tag := range strings.Split(tagString, ",") {
		trimmed := strings.TrimSpace(tag)

		if len(trimmed) > 0 {
			tags.Add(strings.ToLower(trimmed))
		}
	}

	return tags.Entries()
}

func getSearchFiltersFromContext(c *cli.Context) data.Filters {
	return data.Filters{
		Tags:    getTagsFromContext(c),
		Content: strings.ToLower(c.String("content")),
	}
}

func (handler handler) printNotes(notes []data.Note, options printOptions) {
	noteCount := len(notes)

	if noteCount <= 0 {
		return
	}

	noteCountString := ""
	if noteCount == 1 {
		noteCountString = fmt.Sprintf("1 note found")
	} else {
		noteCountString = fmt.Sprintf("%d notes found", noteCount)
	}

	fmt.Println(noteCountString)
	fmt.Println()

	for _, note := range notes {
		handler.printNote(note, options)
	}
}

func (handler handler) printNote(note data.Note, options printOptions) {
	bits := make([]string, 0, 4)

	bits = append(bits, note.Timestamp.Format("January 02, 2006 03:04 PM"))

	if options.detailed {
		bits = append(bits, note.ID)
	}

	if len(note.Tags) > 0 {
		bits = append(bits, strings.Join(note.Tags, ", "))
	}

	fmt.Println(strings.Join(bits, " | "))

	contentString := note.Content

	if options.highlight != "" {
		contentString = utils.HighlightString(note.Content, options.highlight)
	}

	fmt.Println(contentString)
	fmt.Println()
}
