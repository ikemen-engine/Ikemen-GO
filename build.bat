set GOPATH=%cd%/go
set CGO_ENABLED=1
set GOOS=windows
go build -o Ikemen_GO.exe ./src
pause
