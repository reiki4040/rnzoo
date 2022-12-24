package main

import (
	"fmt"
	"os"

	"github.com/urfave/cli/v2"
)

func ErrExit(format string, args ...interface{}) error {
	return cli.Exit(fmt.Sprintf(format, args...), 1)
}

func OkExit(format string, args ...interface{}) error {
	return cli.Exit(fmt.Sprintf(format, args...), 0)
}

var (
	version  string
	revision string
)

func main() {
	cliFlags := []cli.Flag{
		&cli.BoolFlag{
			Name:  OPT_SILENT,
			Usage: "if you want do not output messages.",
		},
		&cli.BoolFlag{
			Name:  OPT_VERBOSE,
			Usage: "if you want show debug messages.",
		},
	}

	commands := []*cli.Command{
		&commandInit,
		&commandEc2run,
		&commandEc2list,
		&commandEc2start,
		&commandEc2stop,
		&commandEc2type,
		&commandEc2terminate,
		&commandEc2Tag,
		&commandAttachEIP,
		&commandMoveEIP,
		&commandDetachEIP,
		&commandGetBilling,
	}
	app := &cli.App{
		Name:     "rnzoo",
		Usage:    "useful commands for ec2.",
		Commands: commands,
		Version:  version + " (" + revision + ")",
		Authors: []*cli.Author{
			{
				Name: "reiki4040",
			},
		},
		Flags: cliFlags,
	}

	app.Run(os.Args)
}
