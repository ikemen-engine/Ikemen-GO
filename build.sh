#!/bin/sh
GOPATH=$PWD/go
export GOPATH
GOOS=windows GOARCH=amd64 CGO_ENABLED=1 go build -o ikemen ./src
go clean ./src
