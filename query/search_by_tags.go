package query

import (
	"github.com/ricanontherun/short-form/database"
	"github.com/ricanontherun/short-form/dto"
	"github.com/ricanontherun/short-form/repository"
)

type searchByTags struct {
	tags       []string
	repository repository.Repository
}

func NewSearchByTagsQuery(tags []string) Query {
	return &searchByTags{tags, repository.NewSqlRepository(database.GetInstance())}
}

func (query *searchByTags) Run() (interface{}, error) {
	return query.repository.SearchNotes(&dto.SearchFilters{
		Tags: query.tags,
	})
}
