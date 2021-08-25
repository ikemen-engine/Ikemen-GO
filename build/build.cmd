@echo off
cd ..
set CGO_ENABLED = 1
set GOOS = windows

if "%~1" == "zig" (
	echo Using Zig for compilation.
	set CC = zig cc
	set CCX = zig c++
)

if not exist go.mod (
	echo Missing dependencies, please run get.cmd
	echo.
	pause
	exit
)
if not exist bin (
	MKDIR bin
) 

echo Building Ikemen GO...
echo. 

go build -v -ldflags -H=windowsgui -o ./bin/Ikemen_GO.exe ./src

pause