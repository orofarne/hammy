#!/bin/sh

PREFIX="/opt/hammy"

set -ex

SRCROOT=`dirname "$(pwd)/$0"`

mkdir -p "$PREFIX"
cp -R "$SRCROOT/src" "$SRCROOT/ruby" "$PREFIX"

export GOPATH="$PREFIX"
cd "$PREFIX"
go run "$SRCROOT/bootstrap.go"
go test hammy && go install hammy hammyd hammycid
cd -
