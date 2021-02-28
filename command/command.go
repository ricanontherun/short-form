package command

import "github.com/ricanontherun/short-form/repository"

type Command interface {
	Execute() error
}

type baseCommand struct {
	repository repository.Repository
}