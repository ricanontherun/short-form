package user_input

import (
	"github.com/urfave/cli/v2"
	"strings"
)

func GetNoteIdFromContext(ctx *cli.Context) string {
	return strings.TrimSpace(ctx.Args().First())
}
