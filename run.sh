#!/bin/sh
GOPATH=$PWD/go
export GOPATH
go fmt ./src/*.go
# godoc -src ./src .* > godoc.txt
go generate ./src/main.go && go run ./src/*.go
