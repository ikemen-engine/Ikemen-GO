#!/bin/sh
GOPATH=$PWD/go
export GOPATH
go get -u github.com/yuin/gopher-lua
go get -u github.com/go-gl/glfw/v3.2/glfw
go get -u github.com/go-gl/gl/v2.1/gl
go get -u github.com/jfreymuth/go-vorbis/ogg/vorbis
go get -u github.com/timshannon/go-openal/openal
