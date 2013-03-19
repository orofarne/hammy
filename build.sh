#!/bin/sh

set -ex

cd `dirname "$(pwd)/$0"`

# Setup golang
hg clone -u release https://code.google.com/p/go golang
cd golang/src
./all.bash
cd -
export PATH="`pwd`/golang/bin:$PATH"

# Setup hammy
. ./env.sh
go run bootstrap.go
go test hammy && go install hammy hammyd hammycid hammydatad
