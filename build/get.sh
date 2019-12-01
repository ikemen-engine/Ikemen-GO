#!/bin/sh
cd ..
export GOPATH=$PWD/go
export CGO_ENABLED=1
if [ -n "$OS" ];    then 
    case "$OS" in
    "windows")
        export GOOS=windows
        export CC=x86_64-w64-mingw32-gcc
        export CXX=x86_64-w64-mingw32-g++
        BINARY_NAME="Ikemen_GO_Win_x64.exe"; 

        ;;
    "mac") 
        export GOOS=darwin
        export CC=o64-clang 
        export CXX=o64-clang++
        BINARY_NAME="Ikemen_GO_mac"; 

        ;;
    "linux") 
        BINARY_NAME="Ikemen_GO_linux"; 
        ;;
	"windows32")
	    export GOOS=windows
		export GOARCH=386
        export CC=i686-w64-mingw32-gcc
        export CXX=i686-w64-mingw32-g++
        BINARY_NAME="Ikemen_GO_Win_x86.exe"; 
        ;;
    esac 
else 
    BINARY_NAME="Ikemen_GO";
fi;

go get -u github.com/yuin/gopher-lua
go get -u github.com/go-gl/glfw/v3.3/glfw
go get -u github.com/go-gl/gl/v2.1/gl
go get -u github.com/timshannon/go-openal/openal
go get -u github.com/Windblade-GR01/glfont
go get -u github.com/flopp/go-findfont
go get -u github.com/faiface/beep
go get -u github.com/hajimehoshi/oto
go get -u github.com/hajimehoshi/go-mp3
go get -u github.com/pkg/errors
go get -u github.com/jfreymuth/oggvorbis
go get -u github.com/mewkiz/flac
