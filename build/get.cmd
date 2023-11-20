@echo off

cd ..
set CGO_ENABLED=1
set GOOS=windows

echo Downloading dependencies...

if not exist go.mod (
	go mod init github.com/ikemen-engine/Ikemen-GO/src
)

go get -v -u github.com/ikemen-engine/beep
go get -v -u github.com/flopp/go-findfont
go get -v -u github.com/go-gl/gl/v2.1/gl
go get -v -u github.com/go-gl/glfw/v3.3/glfw
go get -v -u github.com/ikemen-engine/glfont
go get -v -u github.com/sqweek/dialog
go get -v -u github.com/yuin/gopher-lua

go get -v -u github.com/golang/freetype

echo. 
pause