@echo off

cd ..
set CGO_ENABLED=1
set GOOS=windows

echo Downloading dependencies...

if not exist go.mod (
	go mod init github.com/Windblade-GR01/Ikemen_GO/src
)

go get -u github.com/faiface/beep
go get -u github.com/flopp/go-findfont
go get -u github.com/go-gl/gl/v2.1/gl
go get -u github.com/go-gl/glfw/v3.3/glfw
go get -u github.com/ikemen-engine/glfont
go get -u github.com/ikemen-engine/go-openal
go get -u github.com/sqweek/dialog
go get -u github.com/yuin/gopher-lua

go get -u github.com/golang/freetype

echo. 
pause