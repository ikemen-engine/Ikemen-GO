#!/bin/sh
cd ..

echo "------------------------------------------------------------"
echo "Packaging release..."
echo "------------------------------------------------------------"

curl -SLO https://openal-soft.org/openal-binaries/openal-soft-1.20.1-bin.zip
7z x ./openal-soft-1.20.1-bin.zip
mv openal-soft-1.20.1-bin AL_Temp_416840
mv ./AL_Temp_416840/bin/Win64/soft_oal.dll ./bin/soft_oal_x64.dll
mv ./AL_Temp_416840/bin/Win32/soft_oal.dll ./bin/soft_oal_x86.dll
rm -rf AL_Temp_416840
rm openal-soft-1.20.1-bin.zip

cp ./build/Ikemen_GO.command ./bin/Ikemen_GO.command 
cp ./License.txt ./bin/License.txt

git clone $1
mv ./Ikemen_GO-Elecbyte-Screenpack/chars ./bin/chars
mv ./Ikemen_GO-Elecbyte-Screenpack/data ./bin/data
mv ./Ikemen_GO-Elecbyte-Screenpack/font ./bin/font
mv ./Ikemen_GO-Elecbyte-Screenpack/stages ./bin/stages
mv ./Ikemen_GO-Elecbyte-Screenpack/LICENCE.txt ./bin/ScreenpackLicence.txt
rm -rf Ikemen_GO-Elecbyte-Screenpack

rsync -a ./external ./bin/
rsync -a ./data ./bin/
rsync -a ./font ./bin/

cd bin

mkdir save
mkdir sound
mkdir save/replays

mv ./soft_oal_x86.dll ./OpenAL32.dll
mv ./Ikemen_GO_Win_x86.exe ./Ikemen_GO.exe

echo "------------------------------------------------------------"

7z a -tzip ./release/Ikemen_GO_Win_x86.zip ./chars ./data ./font ./save ./external sound ./stages License.txt 'Ikemen_GO.exe' 'OpenAL32.dll'
7z a -tzip ./release/Ikemen_GO_Win_x86_Binaries_only.zip ./external License.txt 'Ikemen_GO.exe' 'OpenAL32.dll'

mv ./Ikemen_GO.exe ./Ikemen_GO_Win_x86.exe
mv ./OpenAL32.dll ./soft_oal_x86.dll

mv ./Ikemen_GO_Win_x64.exe ./Ikemen_GO.exe
mv ./soft_oal_x64.dll ./OpenAL32.dll

7z a -tzip ./release/Ikemen_GO_Win_x64.zip ./chars ./data ./font ./save ./external sound ./stages License.txt 'Ikemen_GO.exe' 'OpenAL32.dll'
7z a -tzip ./release/Ikemen_GO_Win_x64_Binaries_only.zip ./external ../data ../font License.txt 'Ikemen_GO.exe' 'OpenAL32.dll'

mv ./Ikemen_GO.exe ./Ikemen_GO_Win_x64.exe
mv ./OpenAL32.dll ./soft_oal_x64.dll

7z a -tzip ./release/Ikemen_GO_Mac.zip ./chars ./data ./font ./save ./external sound ./stages License.txt Ikemen_GO.command Ikemen_GO_mac
7z a -tzip ./release/Ikemen_GO_Mac_Binaries_only.zip ./external ../data ../font License.txt Ikemen_GO.command Ikemen_GO_mac

7z a -tzip ./release/Ikemen_GO_Linux.zip ./chars ./data ./font ./save ./external sound ./stages License.txt Ikemen_GO.command Ikemen_GO_linux
7z a -tzip ./release/Ikemen_GO_Linux_Binaries_only.zip ./external ../data ../font License.txt Ikemen_GO.command Ikemen_GO_linux

echo "------------------------------------------------------------"
echo "Packaging finished."
echo "------------------------------------------------------------"
