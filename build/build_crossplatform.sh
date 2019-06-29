#!/bin/sh
cd ..
export GOPATH=$PWD/go
export CGO_ENABLED=1
IS_WINDOWS="0"

if [ -n "$OS" ];    then 
    case "$OS" in
    "windows")
        export GOOS=windows
        export CC=x86_64-w64-mingw32-gcc
        export CXX=x86_64-w64-mingw32-g++
        BINARY_NAME="Ikemen_GO_Win_x64.exe";
		IS_WINDOWS="1"

        ;;
    "mac") 
        export GOOS=darwin
        export CC=o64-clang 
        export CXX=o64-clang++
        BINARY_NAME="Ikemen_GO_mac"; 
		IS_WINDOWS="1"
		
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

mkdir bin

if [ "$IS_WINDOWS" = "1" ]; then
	go build -ldflags "-H windowsgui" -o ./bin/$BINARY_NAME ./src
else
	go build -o ./bin/$BINARY_NAME ./src
fi

chmod +x ./bin/$BINARY_NAME
