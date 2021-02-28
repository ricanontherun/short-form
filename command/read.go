package command

import (
	"github.com/ricanontherun/short-form/utils"
	"strings"
)

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
