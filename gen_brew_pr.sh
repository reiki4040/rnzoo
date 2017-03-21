#!/bin/bash

function usage() {
  cat <<_EOB
Create Pull Request that homebrew-rnzoo for new version.
Usage:
  $(basename $0) <version> <sha256hash>

_EOB
}

function err_exit() {
  echo $* >&2
  exit 1
}

function main() {
  # TODO check version format
  ver=$1
  if [ -z "$ver" ]; then
    err_exit "version is empty. it is required"
  fi

  # TODO check hex format
  s256=$2
  if [ -z "$s256" ]; then
    err_exit "sha256 is empty. it is required"
  fi

  tmpdir=$(mktemp -d /tmp/genpr.XXXXXX)
  echo "made" $tmpdir

  echo "clone homebrew-rnzoo repo..."
  tmprepo="$tmpdir/homebrew-rnzoo"
  git clone git@github.com:reiki4040/homebrew-rnzoo.git $tmprepo

  echo "creating version branch and editing version and sha256 hash..."
  cd $tmprepo
  bname="version/${ver}"
  git checkout -b ${bname}

  echo $ver | grep -q "-"
  if [ $? == 1 ]; then
    # replace version and sha256 (version is homebrew variable. so do not use '=')
	# version format is not allow v prefix (NG: v0.1.0, OK: 0.1.0)
    cat rnzoo.rb | sed -e "s/version \".*\"/version \"${ver}\"/" | sed -e "s/normal_sha256 = \".*\"/normal_sha256 = \"${s256}\"/" > rnzoo.rb.new
  else
    # replace devel_version and sha256
    # version format is not allow v prefix (NG: v0.1.0-dev1, OK: 0.1.0-dev1)
    cat rnzoo.rb | sed -e "s/devel_version = \".*\"/devel_version = \"${ver}\"/" | sed -e "s/devel_sha256 = \".*\"/devel_sha256 = \"${s256}\"/" > rnzoo.rb.new
  fi

  mv rnzoo.rb.new rnzoo.rb

  echo "brew audit..."
  brew audit --strict --online rnzoo.rb
  if [ $? != 0 ]; then
    err_exit "failed brew audit, so stop."
  fi
  echo "brew audit is OK"

  echo "git commit..."
  git commit -a -m "rnzoo version ${ver}"

  if [ -z "$push_origin" ]; then
    echo "commit done. if you want create pull request, then specify -p option."
	exit 0
  fi

  echo "pushing to github..."
  git push origin ${bname}

  echo "creating Pull Request"
  hub pull-request -m "update rnzoo ${ver}" -b "reiki4040:master" -h "reiki4040:${bname}"

  echo "done."
  echo "if you want remove temp files, please remove $tmprepo manually"
}

push_origin=""
while getopts hp OPT
do
  case $OPT in
    h) usage
       exit 0
       ;;
    p)
       push_origin="1"
       ;;
    *) echo "unknown option."
       usage
       exit 1
       ;;
  esac
done
shift $((OPTIND - 1))

main $1 $2
