package models

import (
	"time"
)

type DateRange struct {
	From time.Time
	To   time.Time
}

func GetRangeToday(start time.Time) DateRange {
	return getRange(start)
}

func GetRangeYesterday(start time.Time) DateRange {
	return getRange(start.AddDate(0, 0, -1))
}

func getRange(t time.Time) DateRange {
	rangeStart := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	rangeEnd := time.Date(rangeStart.Year(), rangeStart.Month(), rangeStart.Day(), 23, 59, 59, 0, t.Location())

	return DateRange{
		From: rangeStart,
		To:   rangeEnd,
	}
}
