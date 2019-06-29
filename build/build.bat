cd ..
set GOPATH=%cd%/go
set CGO_ENABLED=1
set GOOS=windows
MKDIR bin
go build -ldflags -H=windowsgui -o ./bin/Ikemen_GO.exe ./src 
pause
