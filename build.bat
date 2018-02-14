set GOPATH=%cd%/go
set CGO_ENABLED=1
go build -ldflags -H=windowsgui -o Ikemen_GO.exe ./src
pause
