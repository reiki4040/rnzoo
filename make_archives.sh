#!/bin/sh
version=$(bin/rnssh -v)

WORK_DIR="work"
DEST_DIR="archives"
current=$(pwd)
if [ -z "$current" ]; then
  exit 1
fi
oss="darwin"
archs="386 amd64"

cmds="ec2list ltsv_pipe"
files="LICENSE README.md"

mkdir -p $current/$DEST_DIR

for os in $oss
do
  for arch in $archs
  do
    rnzoo_prefix="rnzoo-$version-$os-$arch"
    work_dir="$current/$WORK_DIR/$rnzoo_prefix"
    mkdir -p "$work_dir/bin"

	# golang command
    for cmd in $cmds
	do
      cd "$current/$cmd/"
      GOOS="$os" GOARCH="$arch" go build -o "$work_dir/bin/$cmd"
    done

	# script command
	cp -a $current/bin/rnssh $work_dir/bin/

	# something
	for f in $files
	do
      cp -a $current/$f $work_dir/
	done

	cd $current/$WORK_DIR
	# tar.gz
	tar -zcf "$rnzoo_prefix".tar.gz $rnzoo_prefix/*
	mv "$rnzoo_prefix".tar.gz $current/$DEST_DIR/

	# zip
	zip -r "$rnzoo_prefix".zip "./$rnzoo_prefix"
	mv "$rnzoo_prefix".zip $current/$DEST_DIR/
  done
done
