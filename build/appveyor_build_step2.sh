#!/bin/sh
cd ..
export CGO_ENABLED=1

rm 'src/Ikemen_Cylia_x64.syso'
rm 'src/Ikemen_Cylia_x86.syso'

echo "------------------------------------------------------------"
echo "Building linux binary..."
echo "------------------------------------------------------------"
go build -i -tags al_cmpt -o ./bin/Ikemen_GO_linux ./src

echo "------------------------------------------------------------"
echo "Building mac binary..."
echo "------------------------------------------------------------"
export GOOS=darwin
export CC=o64-clang 
export CXX=o64-clang++
go build -i -o ./bin/Ikemen_GO_mac ./src

echo "------------------------------------------------------------"
echo "Building windows x64 binary..."
echo "------------------------------------------------------------"
cp 'windres/Ikemen_Cylia_x64.syso' 'src/Ikemen_Cylia_x64.syso'
export GOOS=windows
export CC=x86_64-w64-mingw32-gcc
export CXX=x86_64-w64-mingw32-g++
go build -i -ldflags "-H windowsgui" -o ./bin/Ikemen_GO_Win_x64.exe ./src
rm 'src/Ikemen_Cylia_x64.syso'

echo "------------------------------------------------------------"
echo "Building windows x86 binary..."
echo "------------------------------------------------------------"
cp 'windres/Ikemen_Cylia_x86.syso' 'src/Ikemen_Cylia_x86.syso'
export GOOS=windows
export GOARCH=386
export CC=i686-w64-mingw32-gcc
export CXX=i686-w64-mingw32-g++
go build -i -ldflags "-H windowsgui" -o ./bin/Ikemen_GO_Win_x86.exe ./src
rm 'src/Ikemen_Cylia_x86.syso'
