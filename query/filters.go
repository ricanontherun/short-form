package query

import (
	"errors"
	"fmt"
	"github.com/ricanontherun/short-form/logging"
	"github.com/ricanontherun/short-form/user_input"
	"github.com/urfave/cli/v2"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	errInvalidAge = errors.New("invalid age")
)

func GetSearchFiltersFromContext(ctx *cli.Context) (*SearchFilters, error) {
	searchFilters := &SearchFilters{
		Tags:    user_input.GetTagsFromContext(ctx),
		Content: user_input.GetContentFlagFromContext(ctx),
		String:  user_input.GetGreedyArgumentFromContext(ctx),
	}

	age := strings.ToLower(ctx.String("age"))
	if len(age) > 0 {
		validAge := regexp.MustCompile(`^\d+d$`)
		if !validAge.MatchString(age) {
			return nil, errInvalidAge
		} else {
			ageDays, _ := strconv.Atoi(strings.TrimRight(age, "d"))
			end := time.Now()
			start := end.AddDate(0, 0, -ageDays)

			searchFilters.DateRange = &DateRange{
				From: start,
				To:   end,
			}
		}
	}

	logging.Debug(fmt.Sprintf("search filters = %+v", searchFilters))
	return searchFilters, nil
}
