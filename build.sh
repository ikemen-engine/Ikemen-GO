#!/bin/sh
export GOPATH=$PWD/go
CGO_ENABLED=1 go build -o Ikemen_GO ./src
