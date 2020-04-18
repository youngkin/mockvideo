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

allLocal() {
    pre
    build
    test
}

allARM() {
    pre
    build
    test
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
elif [ $1 = "allLocal" ]
then
    test
elif [ $1 = "allARM" ]
then
    test
else
    echo "usage:"
    echo "  build.sh [pre | build | buildARM | test | allLocal | allARM]"
    exit 1
fi