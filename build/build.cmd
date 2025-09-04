@echo off
REM ---------------------------------------------------------------------------
REM Build script for Ikemen GO (Windows).
REM Compiles release executable for Windows using Go. Requires Go and MinGW.
REM Expects dependencies already fetched via get.cmd.
REM ---------------------------------------------------------------------------

cd ..                                 REM Move to repository root.
set CGO_ENABLED=1                     REM Enable CGO for C bindings. [PITFALL] Needs MinGW toolchain.
set GOOS=windows                      REM Target OS for Go compiler.

if not exist go.mod (
        echo Missing dependencies, please run get.cmd
        echo.
        pause
        exit /b
)
if not exist bin (
        mkdir bin                      REM Create output directory if missing.
)

echo Building Ikemen GO...
echo.

go build -trimpath -v -ldflags -H=windowsgui -o ./bin/Ikemen_GO.exe ./src  REM Build GUI executable.

pause
