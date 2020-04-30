#!/bin/sh
export GOPATH=$PWD/go
go get -u github.com/yuin/gopher-lua
go get -u github.com/go-gl/glfw/v3.3/glfw
go get -u github.com/go-gl/gl/v2.1/gl
go get -u github.com/timshannon/go-openal/openal
go get -u github.com/Windblade-GR01/glfont
go get -u github.com/flopp/go-findfont
go get -u github.com/faiface/beep
go get -u github.com/hajimehoshi/go-mp3@v0.2.1
go get -u github.com/hajimehoshi/oto@v0.5.4
go get -u github.com/pkg/errors
go get -u github.com/jfreymuth/oggvorbis
go get -u github.com/sqweek/dialog
go get -u github.com/kbinani/screenshot