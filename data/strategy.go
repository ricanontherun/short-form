package data

type Strategy interface {
	Execute(filters Filters) (map[string]*Note, error)
}
