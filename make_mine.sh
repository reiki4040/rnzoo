#!/bin/sh

DEST_DIR="bin"
current=$(pwd)
if [ -z "$current" ]; then
  exit 1
fi

cmds="ec2list ltsv_pipe"
for cmd in $cmds
do
  cd "$current/$cmd/"
  go build -o "$current/$DEST_DIR/$cmd"
done

