package models

type SearchFilters struct {
	DateRange *DateRange

	Tags []string

	Content string
}
