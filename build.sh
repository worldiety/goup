#!/usr/bin/env bash

set -e

rm -rf builds
mkdir -p builds


mkdir -p builds/linux-amd64
env GOOS=linux GOARCH=amd64 go build
mv goup builds/linux-amd64

mkdir -p builds/darwin-amd64
env GOOS=darwin GOARCH=amd64 go build
mv goup builds/darwin-amd64

mkdir -p builds/windows-amd64
env GOOS=windows GOARCH=amd64 go build
mv goup.exe builds/windows-amd64