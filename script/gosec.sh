#!/bin/bash -e
# Basic Inspects source code for security problems by scanning the Go AST.
# https://github.com/securego/gosec
dir="$(dirname $(realpath ${BASH_SOURCE[0]}))" && cd $dir
dir=$dir/.. && cd $dir
export PATH=$dir/bin:$PATH

if ! [ -x "$(command -v gosec)" ]; then
    go install github.com/securego/gosec/v2/cmd/gosec@latest
fi

if [ -x "$(command -v gosec)" ]; then
    gosec ./...
fi