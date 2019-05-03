#!/usr/bin/env bash

# this script bootstraps the gomobilebuilder setup

set -e

if ! [ -x "$(command -v go)" ]; then
  echo 'Error: go is not installed.' >&2
  exit 1
fi


if ! [ -x "$(command -v .gmb/gomobilebuilder)" ]; then
  mkdir ./.gmb
  cd ./.gmb
  go get -u github.com/worldiety/gomobilebuilder
  go build
  cd ..
fi

