package data

import "time"

type SearchContext struct {
	// Allow arbitrary date range searching.
	From time.Time
	To time.Time

	Tags []string
}