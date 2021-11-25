#!/bin/bash
set -e

# TODO ARM support. figure out mapping to uname -m output.

tag="${1}"

if [ "${tag}" == "" ]; then
  echo "tag argument required"
  exit 1
fi

rm -rf dist
GOOS=darwin GOARCH=amd64 go build -o "dist/darwin-x86_64"
GOOS=darwin GOARCH=arm64 go build -o "dist/darwin-arm64"
GOOS=linux GOARCH=386 go build -o "dist/linux-i386"
GOOS=linux GOARCH=amd64 go build -o "dist/linux-x86_64"
GOOS=windows GOARCH=386 go build -o "dist/windows-i386"
GOOS=windows GOARCH=amd64 go build -o "dist/windows-x86_64"

gh release create $tag ./dist/* --title="${tag}" --notes "${tag}"
