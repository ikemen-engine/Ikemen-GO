#!/bin/sh
export GOPATH=$PWD/go
GOOS=windows GOARCH=amd64 CGO_ENABLED=1 go build -o ikemen ./src
