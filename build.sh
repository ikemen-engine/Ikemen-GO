#!/bin/sh
export GOPATH=$PWD/go
CGO_ENABLED=1 CGO_CFLAGS=-I/home/suehiro/mingw-root/usr/local/include CGO_LDFLAGS=-L/home/suehiro/mingw-root/usr/local/lib CC=x86_64-w64-mingw32-gcc CXX=x86_64-w64-mingw32-g++ GOOS=windows GOARCH=amd64 go build -o ikemen.exe ./src
