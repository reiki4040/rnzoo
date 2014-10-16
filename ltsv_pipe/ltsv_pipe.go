package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/reiki4040/ltsv_pipe/goltsv"

	flag "github.com/dotcloud/docker/pkg/mflag"
)

var (
	tsv_mode     bool
	show_version bool
	show_usage   bool
)

func parseflg() {
	flag.BoolVar(&show_version, []string{"v", "-version"}, false, "show version.")
	flag.BoolVar(&show_usage, []string{"h", "-help"}, false, "show this usage.")
	flag.BoolVar(&tsv_mode, []string{"t", "-tsv"}, false, "Output with TSV format")
	flag.Parse()
}

func version() {
	fmt.Printf("%s\n", Version)
}

func usage() {
	fmt.Printf("%s\n", Usage)
}

func main() {
	parseflg()

	if show_usage {
		usage()
		os.Exit(0)
	}

	if show_version {
		version()
		os.Exit(0)
	}

	ltsv_pipe()
}

func ltsv_pipe() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := strings.TrimRight(scanner.Text(), "\n")
		if line == "" {
			continue
		}

		items := goltsv.ParseLtsv(line)

		if tsv_mode {
			fmt.Printf("%s\n", goltsv.Map2OrderedTsv(items, flag.Args()...))
		} else {
			fmt.Printf("%s\n", goltsv.Map2OrderedLtsv(items, flag.Args()...))
		}
	}
}
