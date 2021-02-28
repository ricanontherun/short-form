package dto

type SearchFilters struct {
	DateRange *DateRange
	Tags      []string
	Content   string
	String    string
}