#!/bin/sh

PREFIX="/opt/hammy"

set -ex

cd `dirname "$(pwd)/$0"`

mkdir -p "$PREFIX"

cp bin/hammy "$PREFIX"
cp bin/hammyd "$PREFIX"
cp bin/hammycid "$PREFIX"
cp bin/hammydatad "$PREFIX"
cp -R worker "$PREFIX"
