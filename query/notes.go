package query

import (
	"github.com/ricanontherun/short-form/database"
	"github.com/ricanontherun/short-form/repository"
)

type countQuery struct {
	tags []string
	repository repository.Repository
}

// Fetch the number of notes which are associated with ANY of the provided tags.
func NewTagsCountQuery(tags []string) Query {
	return &countQuery{
		tags,
		repository.NewSqlRepository(database.GetInstance()),
	}
}

func (query *countQuery) Run() (interface{}, error) {
	return query.repository.GetNoteCountByTags(query.tags)
}