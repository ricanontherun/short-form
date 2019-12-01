package commands

import "github.com/urfave/cli"

type Command interface {
	Execute(ctx *cli.Context) error
}
