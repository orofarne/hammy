#!/bin/sh

set -ex

cd `dirname "$(pwd)/$0"`

rm -rf bin
rm -rf pkg
rm -rf golang
rm -rf contrib
rm -rf worker/build
