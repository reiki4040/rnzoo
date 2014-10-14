package main

const (
	Version string = "0.1.0"
	Usage          = `Usage:
  ltsv_pipe can filter by label name items that from stdin.
  and the result output to stdout.

    echo -e "l1:v1\\tl2:v2\\tl3:v3\\tl4:v4" | ltsv_pipe l1 l2
    l1:v1<TAB>l2:v2

  you can specify order with arguments.

    echo -e "l1:v1\\tl2:v2\\tl3:v3\\tl4:v4" | ltsv_pipe l2 l1
    l2:v2<TAB>l1:v1

  and use tsv mode option (-t, --tsv), output with TSV format.

    echo -e "l1:v1\\tl2:v2\\tl3:v3\\tl4:v4" | ltsv_pipe -t l2 l1
    v2<TAB>v1

Options:

  -v, --version: show version.
  -h, --help:    show this usage.
  -t, --tsv:     output with TSV format.
`
)
