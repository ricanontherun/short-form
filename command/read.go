package command

import (
	"bufio"
	"github.com/ricanontherun/short-form/utils"
	"github.com/urfave/cli/v2"
	"io"
	"os"
	"strings"
)

// TODO: private
// Return a cleaned array of tagsType provided as --tagsType=t1,t2,t3, as ['t1', 't2', 't3']
func ReadTagsFromContext(c *cli.Context) []string {
	return CleanTagsFromString(c.String(flagTags))
}

func readContentFromContext(ctx *cli.Context) (string, error) {
	stdinStat, err := os.Stdin.Stat()
	if err != nil {
		return "", err
	}

	var content string

	args := ctx.Args().Slice()
	if len(args) != 0 {
		content = strings.Join(args, " ")
	} else if stdinStat.Mode()&os.ModeCharDevice != 0 || stdinStat.Size() != 0 {
		stdinReader := bufio.NewReader(os.Stdin)
		var stdinBuilder strings.Builder

		for {
			if r, _, readErr := stdinReader.ReadRune(); readErr != nil && readErr == io.EOF {
				break
			} else {
				if readErr != nil { // unexpected (non EOF) error.
					return "", readErr
				}

				stdinBuilder.WriteRune(r)
			}
		}

		content = strings.TrimSuffix(stdinBuilder.String(), "\n")
	}

	return content, nil
}

func CleanTagsFromString(tagString string) []string {
	tags := utils.NewSet()

	for _, tag := range strings.Split(tagString, ",") {
		trimmed := strings.TrimSpace(tag)

		if len(trimmed) > 0 {
			tags.Add(trimmed)
		}
	}

	return tags.Entries()
}
