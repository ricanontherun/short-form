package query

type Query interface {
	Run() (interface{}, error)
}
