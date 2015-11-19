#!/bin/sh
version=0.3.0
os="darwin"
arch="amd64"

HASH=$(git rev-parse --verify HEAD)
BUILDDATE=$(date '+%Y/%m/%d %H:%M:%S %Z')
GOOS="$os" GOARCH="$arch" gom build -o "rnzoo" -ldflags "-X main.version=$version -X main.hash=$HASH -X \"main.builddate=$BUILDDATE\""
