@echo off
cd ..
set CGO_ENABLED=1
set GOOS=windows

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

go build -v -ldflags -H=windowsgui -o ./bin/Ikemen_GO.exe ./src

echo.
pause