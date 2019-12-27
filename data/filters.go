package data

import "github.com/ricanontherun/short-form/utils"

type Filters struct {
	DateRange *utils.DateRange

	Tags []string

	Content string
}
