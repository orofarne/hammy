#!/bin/sh

PREFIX="/opt/hammy"

set -ex

cd `dirname "$(pwd)/$0"`

test -d "$PREFIX/bin" || mkdir -p "$PREFIX/bin"

cp bin/hammyd "$PREFIX/bin"
cp bin/hammycid "$PREFIX/bin"
cp bin/hammydatad "$PREFIX/bin"
cp -R worker "$PREFIX"
