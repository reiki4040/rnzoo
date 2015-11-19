package main

import (
	"os"

	"github.com/codegangsta/cli"
)

var (
	version   string
	hash      string
	builddate string
)

var CliFlags = []cli.Flag{
	cli.BoolFlag{
		Name:  OPT_SILENT,
		Usage: "if you want do not output messages.",
	},
}

var Commands = []cli.Command{
	commandEc2start,
	commandEc2stop,
	commandEc2list,
}

func main() {
	app := cli.NewApp()
	app.Name = "rnzoo"
	app.Version = version + " (" + hash + ") built:" + builddate
	app.Usage = "useful commands for ec2."
	app.Author = "reiki4040"
	app.Email = ""
	app.Commands = Commands
	app.Flags = CliFlags
	app.Run(os.Args)
}
