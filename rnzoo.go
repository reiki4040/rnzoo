package main

import (
	"os"

	"github.com/codegangsta/cli"
)

var (
	version   string
	hash      string
	builddate string
	goversion string
)

var CliFlags = []cli.Flag{
	cli.BoolFlag{
		Name:  OPT_SILENT,
		Usage: "if you want do not output messages.",
	},
	cli.BoolFlag{
		Name:  OPT_VERBOSE,
		Usage: "if you want show debug messages.",
	},
}

var Commands = []cli.Command{
	commandInit,
	commandEc2run,
	commandEc2list,
	commandEc2start,
	commandEc2stop,
	commandEc2type,
	commandEc2terminate,
	commandEc2Tag,
	commandAttachEIP,
	commandMoveEIP,
	commandDetachEIP,
	commandGetBilling,
}

func main() {
	app := cli.NewApp()
	app.Name = "rnzoo"
	app.Version = version + " (" + hash + ") " + goversion
	app.Usage = "useful commands for ec2."
	app.Author = "reiki4040"
	app.Email = ""
	app.Commands = Commands
	app.Flags = CliFlags
	app.Run(os.Args)
}
