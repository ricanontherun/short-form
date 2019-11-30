package data

import "time"

type DateRange struct {
	From time.Time
	To   time.Time
}

type Filters struct {
	DateRange *DateRange

	Tags []string
}
