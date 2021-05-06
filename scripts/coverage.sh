#!/usr/bin/env bash

set -eu

cd "$(dirname "$0")/.."

if [ $# -eq 0 ]; then
  target="$(go list ./... | fzf)"
  if [ "$target" = "" ]; then
    exit 0
  fi
elif [ $# -eq 1 ]; then
  target=$1
else
  echo "too many arguments are given: $*" >&2
  exit 1
fi

set -x
dir=.coverage/${target#./}
mkdir -p "$dir"
go test "$target" -coverprofile="${dir}/coverage.txt" -covermode=atomic -race
go tool cover -html="${dir}/coverage.txt"
