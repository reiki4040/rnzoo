#!/bin/sh
version=0.4.0
os="darwin"
arch="amd64"

HASH=$(git rev-parse --verify HEAD)
GOVERSION=$(go version)
GOOS="$os" GOARCH="$arch" go build -o "rnzoo" -ldflags "-X main.version=$version -X main.hash=$HASH -X \"main.goversion=$GOVERSION\""
