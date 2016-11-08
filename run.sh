#!/bin/sh
GOPATH=$PWD/go
export GOPATH
go fmt ./src/*.go
# godoc -src ./src .* > godoc.txt
go generate ./src/main.go && GODEBUG=cgocheck=0 go run ./src/*.go
