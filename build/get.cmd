@echo off
cd ..
set GOPATH=%cd%/go
set CGO_ENABLED=1
set GOOS=windows

echo Downloading dependencies...
echo. 

if not exist go.mod (
	go mod init github.com/Windblade-GR01/Ikemen_GO/src
	echo. 
)

go get -u github.com/yuin/gopher-lua
go get -u github.com/go-gl/glfw/v3.3/glfw
go get -u github.com/go-gl/gl/v2.1/gl
go get -u github.com/Windblade-GR01/go-openal/openal
go get -u github.com/Windblade-GR01/glfont
go get -u github.com/flopp/go-findfont
go get -u github.com/faiface/beep
go get -u github.com/hajimehoshi/go-mp3@v0.2.1
go get -u github.com/hajimehoshi/oto@v0.5.4
go get -u github.com/pkg/errors
go get -u github.com/jfreymuth/oggvorbis
go get -u github.com/sqweek/dialog

echo. 
pause