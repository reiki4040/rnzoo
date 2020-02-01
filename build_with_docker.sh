#!/bin/bash

#--------------------------------------------------------------#
# build rnzoo on golang docker image.
# 1. mount local rnzoo directory on container.
# 2. build rnzoo with build.sh on container.
# 3. stored binary to mounted directory.
# 4. got rnzoo binary that compiled docker image go version.
#--------------------------------------------------------------#

function usage() {
  cat <<_EOB
rnzoo build script with docker.

  - build rnzoo binary.
  - create release archive and show sha256 (for homebrew formula)

[Options]
  -a: create archive for release
  -g: run glide up

_EOB
}

function build() {
  local opt=""
  if [ $mode = "archive" ]; then
    opt="-a"
  fi

  # run rnzoo build with docker
  docker run --rm \
		  -v $GOPATH/src/github.com/reiki4040/rnzoo:/go/src/github.com/reiki4040/rnzoo \
		  -w /go/src/github.com/reiki4040/rnzoo \
		  golang:1.13 bash build.sh $opt
}

mode="build"
glideup=
while getopts ah OPT
do
  case $OPT in
    a) mode="archive"
       ;;
    h) usage
       exit 0
       ;;
    *) echo "unknown option."
       usage
       exit 1
       ;;
  esac
done
shift $((OPTIND - 1))

build
