#!/bin/bash

pre() {
    go vet ./...
    go fmt ./...
    /Users/rich_youngkin/Software/repos/go/bin/golint
}

build() {
    cd cmd/accountd
    go build
    cd -
}

buildARM() {
    cd cmd/accountd
    /usr/bin/env GOOS=linux GOARCH=arm GOARM=7 go build
    cd -
}

dockerBuild() {
    buildARM
    cd cmd/accountd
    docker build -t local/accountd .
    cd -
}

test() {
    pre
    go test -race ./...
}

allLocal() {
    pre
    build
    test
}

allARM() {
    pre
    dockerBuild
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
elif [ $1 = "dockerBuild" ] 
then
    dockerBuild
elif [ $1 = "test" ]
then
    test
elif [ $1 = "allLocal" ]
then
    allLocal
elif [ $1 = "allARM" ]
then
    allARM
else
    echo "usage:"
    echo "  build.sh [pre | build | buildARM | dockerBuild | test | allLocal | allARM]"
    exit 1
fi