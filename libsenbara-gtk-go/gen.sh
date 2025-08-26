#!/usr/bin/env bash

set -e

echo "generating go files..."
go generate

tmp=$(mktemp -d)

echo "making sure imports are correct..."
GOPATH="$tmp" goimports --srcdir . -w v4 || { rm -r "$tmp"; exit 1; }
rm -r "$tmp"

echo "formatting files..."
go fmt "github.com/pojntfx/senbara/libsenbara-gtk-go/v4/..."

echo "running a second pass for goimports..."
goimports -w v4

echo "running go vet..."
go vet ./v4/...
