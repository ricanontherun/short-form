package utils

import (
	"strconv"
	"time"
)

func CurrentUnixTimestamp() string {
	return ToUnixTimestampString(time.Now())
}

func ToUnixTimestampString(t time.Time) string {
	return strconv.FormatInt(t.Unix(), 10)
}
