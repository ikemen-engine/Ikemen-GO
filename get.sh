#!/bin/sh
GOPATH=`pwd`/go
export GOPATH
go get -u github.com/Shopify/go-lua
go get -u github.com/go-gl/glfw/v3.2/glfw
go get -u github.com/xlab/vorbis-go/vorbis
go get -u github.com/gordonklaus/portaudio
