#!/bin/sh
GOPATH=`pwd`/go
export GOPATH
go fmt ./src/*.go
godoc -src ./src .* > godoc.txt
go run ./src/*.go
