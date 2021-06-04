#!/bin/bash
cd ..
export CGO_ENABLED=1

echo "Downloading dependencies..."
echo ""

if [ ! -f ./go.mod ]; then
	go mod init github.com/Windblade-GR01/Ikemen_GO/src
	echo ""
fi

go get -u github.com/golang/freetype

go get -u github.com/faiface/beep v1.0.2
go get -u github.com/flopp/go-findfont
go get -u github.com/go-gl/gl
go get -u github.com/go-gl/glfw/v3.3/glfw
go get -u github.com/ikemen-engine/glfont
go get -u github.com/ikemen-engine/go-openal
go get -u github.com/sqweek/dialog
go get -u github.com/yuin/gopher-lua