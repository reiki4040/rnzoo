#!/bin/bash

function usage() {
  cat <<_EOB
rnzoo release script.

  - build rnzoo binary and create archive.
  - create github release and upload archive.
  - create homebrew pull request that new version

[Options]
  -d: develop version release

_EOB
}

function releaseflow() {
  bash build_with_docker.sh -a

  # TODO get version and archive path from build_with_docker...
  version=$(git describe --tags)
  archive="archives/rnzoo-${version}-darwin-amd64.tar.gz"

  # create github release
  # TODO release message creation
  hub release create -p -a $archive -m "$version" "$version"

  # create homebrew Pull Request
  sha256=$(shasum -a 256 $archive | cut -d' ' -f1)
  bash gen_brew_pr.sh -p "$version" "$sha256"
}

while getopts h OPT
do
  case $OPT in
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

releaseflow
