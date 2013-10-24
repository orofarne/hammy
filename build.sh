#!/bin/sh

set -ex

# Setup golang
hg clone https://code.google.com/p/go golang
cd golang/src
./all.bash
cd -

# Setup hammy
. ./setup_env.sh
