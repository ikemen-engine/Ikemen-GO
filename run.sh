#!/bin/sh
export GOPATH=$PWD/go
go fmt ./src/*.go
go run ./src/*.go
