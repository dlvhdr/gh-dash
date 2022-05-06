#!/bin/bash
set -e

# TODO ARM support. figure out mapping to uname -m output.

tag="${1}"

if [ "${tag}" == "" ]; then
  echo "tag argument required"
  exit 1
fi

rm -rf dist
GOOS=darwin GOARCH=amd64 go build -o "dist/darwin-amd64"
GOOS=darwin GOARCH=arm64 go build -o "dist/darwin-arm64"
GOOS=freebsd GOARCH=386 go build -o "dist/freebsd-386"
GOOS=freebsd GOARCH=amd64 go build -o "dist/freebsd-amd64"
GOOS=freebsd GOARCH=arm64 go build -o "dist/freebsd-arm64"
GOOS=linux GOARCH=386 go build -o "dist/linux-i386"
GOOS=linux GOARCH=amd64 go build -o "dist/linux-amd64"
GOOS=linux GOARCH=arm GOARM=6 go build -o "dist/linux-arm-6"
GOOS=linux GOARCH=arm64 go build -o "dist/linux-arm64"
GOOS=windows GOARCH=386 go build -o "dist/windows-i386.exe"
GOOS=windows GOARCH=amd64 go build -o "dist/windows-amd64.exe"

gh release create $tag ./dist/* --title="${tag}" --notes "${tag}"
