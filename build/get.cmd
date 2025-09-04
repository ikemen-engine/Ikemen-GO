@echo off
REM ---------------------------------------------------------------------------
REM Dependency fetch script for Ikemen GO (Windows).
REM Downloads Go modules needed for building the project.
REM Requires Go with network access. Run prior to build.cmd.
REM ---------------------------------------------------------------------------

cd ..                                  REM Move to repository root.
set CGO_ENABLED=1                      REM Enable CGO for packages requiring C. [PITFALL] Needs MinGW toolchain.
set GOOS=windows                       REM Ensure Go pulls Windows-specific packages if any.

echo Downloading dependencies...

if not exist go.mod (
        go mod init github.com/ikemen-engine/Ikemen-GO/src
)

REM Fetch dependencies (-u updates). Network required.
go get -v -u github.com/ikemen-engine/beep

go get -v -u github.com/flopp/go-findfont

go get -v -u github.com/go-gl/gl/v2.1/gl

go get -v -u github.com/go-gl/glfw/v3.3/glfw

go get -v -u github.com/ikemen-engine/glfont

go get -v -u github.com/sqweek/dialog

go get -v -u github.com/yuin/gopher-lua

go get -v -u github.com/golang/freetype

go get -v -u github.com/lukegb/dds

go get -v -u github.com/qmuntal/gltf

echo.
pause
