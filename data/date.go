package data

import (
	"short-form/utils"
)

type searchByDate struct {
	repository Repository
}

func (search searchByDate) Execute(filters Filters) (map[string]*Note, error) {
	notes, err := search.repository.SearchNotesByDate(&DateRange{
		From: filters.DateRange.From,
		To:   filters.DateRange.To,
	})

	if err != nil {
		return nil, err
	}

	filterOnTags := len(filters.Tags) > 0
	// Filter by tag (behaves as an AND)
	for id := range notes {
		if noteTags, err := search.repository.GetNoteTags(id); err != nil {
			return nil, err
		} else {
			if filterOnTags && len(noteTags) == 0 {
				delete(notes, id)
				continue
			}

			match := true
			if filterOnTags {
				for _, tag := range filters.Tags {
					if !utils.InArray(tag, noteTags) {
						match = false
						break
					}
				}
			}

			if !match {
				delete(notes, id)
			} else {
				for _, noteTag := range noteTags {
					notes[id].Tags = append(notes[id].Tags, noteTag)
				}
			}
		}
	}

	// Get the content for each note.

	return notes, nil
}

func MakeByDateStrategy(repository Repository) Strategy {
	return searchByDate{repository}
}
