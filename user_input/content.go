package user_input

import (
	"bufio"
	"github.com/urfave/cli/v2"
	"io"
	"os"
	"strings"
)

func GetContentFlagFromContext(ctx *cli.Context) string {
	return strings.TrimSpace(ctx.String(flagContent))
}

func GetGreedyArgumentFromContext(ctx *cli.Context) string {
	return strings.TrimSpace(strings.Join(ctx.Args().Slice(), " "))
}

func GetContentFromContext(ctx *cli.Context) (string, error) {
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