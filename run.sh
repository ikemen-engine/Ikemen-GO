#!/bin/sh
GOPATH=`pwd`/go
export GOPATH
go fmt ./src/*.go
go run ./src/main.go
