#!/bin/sh
version=0.4.0

WORK_DIR="work"
DEST_DIR="archives"
current=$(pwd)
if [ -z "$current" ]; then
  exit 1
fi
oss="darwin"
archs="amd64"

files="LICENSE README.md"

mkdir -p $current/$DEST_DIR

for os in $oss
do
  for arch in $archs
  do
    echo "start $os/$arch build and create archive file."
    rnzoo_prefix="rnzoo-$version-$os-$arch"
    archive_dir="$current/$WORK_DIR/$rnzoo_prefix"
    mkdir -p "$archive_dir"

	# build
    cd $current
    HASH=$(git rev-parse --verify HEAD)
	GOVERSION=$(go version)
    GOOS="$os" GOARCH="$arch" go build -o "$archive_dir/rnzoo" -ldflags "-X main.version=$version -X main.hash=$HASH -X \"main.goversion=$GOVERSION\""

	# something
    for f in $files
    do
      cp -a $current/$f $archive_dir/
    done

    echo "creating zip archive..."
    cd $current/$WORK_DIR
    zip -r "$rnzoo_prefix".zip "./$rnzoo_prefix"
    mv "$rnzoo_prefix".zip $current/$DEST_DIR/
    shasum -a 256 "$current/$DEST_DIR/$rnzoo_prefix.zip"
    echo "finished $os/$arch build and create archive file."
    echo ""
  done
done
