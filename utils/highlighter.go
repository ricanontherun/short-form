package utils

import (
	"github.com/fatih/color"
	"strings"
)

type highlight struct {
	Left      string
	Highlight string
	Right     string
}

// HighlightString highlights (using terminal codes) the occurrences of highlight in original.
func HighlightString(original string, highlight string) string {
	highlights := parseHighlights(original, highlight)

	if len(highlights) > 0 {
		highlighted := ""
		colorPrinter := color.New(color.Bold)

		for _, hl := range highlights {
			highlighted += hl.Left + colorPrinter.Sprint(hl.Highlight) + hl.Right
		}

		return highlighted
	}

	return original
}

// TODO: This could be much more efficient.
func parseHighlights(highlightString string, original string) []highlight {
	var highlights []highlight

	cursor := highlightString
	for cursor != "" {
		startIndex := strings.Index(cursor, original)
		if startIndex == -1 {
			break
		}

		cursorBytes := []byte(cursor)
		highlight := highlight{
			Left:      string(cursorBytes[0:startIndex]),
			Highlight: string(cursorBytes[startIndex : startIndex+len(original)]),
			Right:     "",
		}

		cursor = strings.Replace(cursor, highlight.Left+highlight.Highlight, "", 1)

		// If we're on the the last occurrence, keep the tail.
		nextIndex := strings.Index(cursor, original)
		if nextIndex == -1 {
			highlight.Right = string(cursorBytes[startIndex+len(original):])
		}

		highlights = append(highlights, highlight)
	}

	return highlights
}
