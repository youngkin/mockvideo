#!/bin/bash

pre() {
    go vet ./...
    go fmt ./...
    /Users/rich_youngkin/Software/repos/go/bin/golint
}

build() {
    cd src/cmd/accountd
    go build
    cd -
}

buildARM() {
    cd src/cmd/accountd
    /usr/bin/env GOOS=linux GOARCH=arm GOARM=7 go build
    cd -
}

test() {
    go test -race ./...
}

if [ $1 = "pre" ] 
then
    pre
elif [ $1 = "build" ] 
then
    build
elif [ $1 = "buildARM" ] 
then
    buildARM
elif [ $1 = "test" ]
then
    test
else
    echo "usage:"
    echo "  build.sh [pre | build | buildARM | test]"
    exit 1
fi