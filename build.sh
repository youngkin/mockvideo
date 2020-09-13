#!/bin/bash

pre() {
    go vet ./...
    go fmt ./...
    /Users/rich_youngkin/Software/repos/go/bin/golint
}

build() {
    pre
    genProtobuf
    cd cmd/accountd
    go build
    cd -
}

buildARM() {
    pre
    cd cmd/accountd
    genProtobuf
    /usr/bin/env GOOS=linux GOARCH=arm GOARM=7 go build
    cd -
}

dockerBuild() {
    buildARM
    genProtobuf
    cd cmd/accountd
    docker build -t local/accountd .
    cd -
}

genProtobuf() {
    protoc --go_out=plugins=grpc:. ./pkg/protobuf/accountd/user_service.proto
}

test() {
    pre
    genProtobuf
   go test -race ./... -count=1
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
elif [ $1 = "protobuf" ] 
then
    genProtobuf
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