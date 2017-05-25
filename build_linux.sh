#!/bin/sh
VERSION=$(git describe --tags)
HASH=$(git rev-parse --verify HEAD)
GOVERSION=$(go version)

ARCHIVE_INCLUDES_FILES="LICENSE README.md"

function usage() {
  cat <<_EOB
rnzoo build script for alpha platforms (linux)

  - build rnzoo binary.

[Options]
  -g: run glide up when build
  -s: show current build version for check
  -q: quiet mode

_EOB
}

function show_build_version() {
  echo $VERSION
}

quiet=""
function msg() {
  test -z "$quiet" && echo $*
}

function err_exit() {
  echo $* >&2
  exit 1
}

function build() {
  local dest_dir=$1

  if [ -n "$glideup" ]; then
    msg "run glide up..."
    if [ -n "$quiet" ]; then
      glide -q up
	else
      glide up
	fi
  fi

  #for platform in "linux" "windows"; do
  for platform in "linux"; do
    msg "start build rnzoo for $platform..."
    GOOS="$platform" GOARCH="amd64" go build -o "$dest_dir/rnzoo_${platform}_amd64" -ldflags "-X main.version=$VERSION -X main.hash=$HASH -X \"main.goversion=$GOVERSION\""
    msg "finished build rnzoo for $platform."
  done
}

mode="build"
glideup=""
while getopts ashqu OPT
do
  case $OPT in
    s) show_build_version
       exit 0
       ;;
    u) glideup="1"
       ;;
    h) usage
       exit 0
       ;;
    q) quiet=1
       ;;
    *) echo "unknown option."
       usage
       exit 1
       ;;
  esac
done
shift $((OPTIND - 1))

# run build or archive
case $mode in
  "build")
    build $(pwd)
    ;;
  *)
    echo "unknown mode"
    usage
    exit 1
    ;;
esac
