#!/bin/sh
export GOPATH=$PWD/go
go get -u github.com/yuin/gopher-lua
go get -u github.com/go-gl/glfw/v3.2/glfw
go get -u github.com/go-gl/gl/v2.1/gl
go get -u github.com/jfreymuth/go-vorbis/ogg/vorbis
go get -u github.com/timshannon/go-openal/openal
go get -u github.com/faiface/beep
go get -u github.com/hajimehoshi/oto
go get -u github.com/hajimehoshi/go-mp3
