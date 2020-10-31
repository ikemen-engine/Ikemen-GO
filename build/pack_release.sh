#!/bin/sh
cd ..

curl -SLO https://kcat.strangesoft.net/openal-binaries/openal-soft-1.20.1-bin.zip
7z x ./openal-soft-1.20.1-bin.zip
mv openal-soft-1.20.1-bin AL_Temp_416840
mv ./AL_Temp_416840/bin/Win64/soft_oal.dll ./bin/soft_oal_x64.dll
mv ./AL_Temp_416840/bin/Win32/soft_oal.dll ./bin/soft_oal_x86.dll
rm -rf AL_Temp_416840
rm openal-soft-1.20.1-bin.zip

cd bin
mkdir release

7z a -tzip ./release/Ikemen_GO_Win_x86.zip ../script ../data 'Ikemen_GO_Win_x86.exe' 'soft_oal_x86.dll'
7z rn ./release/Ikemen_GO_Win_x86.zip 'Ikemen_GO_Win_x86.exe' 'Ikemen_GO.exe'
7z rn ./release/Ikemen_GO_Win_x86.zip 'soft_oal_x86.dll' 'OpenAL32.dll'

7z a -tzip ./release/Ikemen_GO_Win_x64.zip ../script ../data 'Ikemen_GO_Win_x64.exe' 'soft_oal_x64.dll'
7z rn ./release/Ikemen_GO_Win_x64.zip 'Ikemen_GO_Win_x64.exe' 'Ikemen_GO.exe'
7z rn ./release/Ikemen_GO_Win_x64.zip 'soft_oal_x64.dll' 'OpenAL32.dll'

7z a -tzip ./release/Ikemen_GO_Mac.zip ../script ../data Ikemen_GO.command Ikemen_GO_mac
7z a -tzip ./release/Ikemen_GO_Linux.zip ../script ../data Ikemen_GO.command Ikemen_GO_linux