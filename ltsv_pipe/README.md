ltsv_pipe
===========

ltsv_pipe can filter ltsv data. ltsv_pipe is written by golang.

## Usage

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

## Install

To install, use `go get`:

```bash
$ go get -d github.com/reiki4040/ltsv_pipe
```

## TODO

- goltsv godoc

## Copyright and LICENSE

Copyright (c) 2014- [reiki4040](https://github.com/reiki4040)

MIT LICENSE
