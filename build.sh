#!/bin/sh

set -ex

cd `dirname "$(pwd)/$0"`

# Setup golang
hg clone -u release https://code.google.com/p/go golang
cd golang/src
./all.bash
cd -

# Setup hammy
. ./env_ex.sh
go run bootstrap.go
go test hammy && go install hammy hammyd hammycid hammydatad

# Setup worker
cd worker
mkdir build
cd build
cmake ..
make
make test
cd ../..
