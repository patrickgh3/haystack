#!/bin/bash
# Runs "go install" then copies binary to "test" directory.
# (not using $GOPATH/bin/ because program reads config file from its directory)

go install
if [ $? != 0 ]; then
    exit 1
fi
mkdir -p test
cp $GOPATH/bin/following test/

