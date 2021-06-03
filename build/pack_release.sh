#!/bin/bash
cd ..

if [ ! -d ./bin ]; then
	mkdir bin
fi

cd bin

if [ ! -d ./release ]; then
	mkdir release
fi

7z a -tzip ./release/Ikemen_GO.zip ./external ../data ../font License.txt 'IkemenGO_x86.exe' 'IkemenGO_x64.exe' Ikemen_GO.command Ikemen_GO_mac Ikemen_GO_linux