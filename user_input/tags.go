package user_input

import (
	"github.com/ricanontherun/short-form/utils"
	"github.com/urfave/cli/v2"
	"strings"
)

func GetTagsFromContext(ctx *cli.Context) []string {
	set := utils.NewSet()
	for _, tag := range strings.Split(strings.ToLower(strings.TrimSpace(ctx.String("tags"))), ",") {
		trimmedTag := strings.TrimSpace(tag)
		if len(trimmedTag) > 0 {
			set.Add(tag)
		}
	}
	return set.Entries()
}
