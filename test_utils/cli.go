package test_utils

import (
	"flag"
	"github.com/urfave/cli/v2"
)

func CreateAppContext(flags map[string]string, args []string) *cli.Context {
	flagSet := flag.NewFlagSet("tests", flag.ContinueOnError)

	for key, value := range flags {
		flagSet.String(key, value, "")
	}

	if err := flagSet.Parse(args); err != nil {
		panic(err)
	}

	return cli.NewContext(cli.NewApp(), flagSet, nil)
}

