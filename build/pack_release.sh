#!/bin/bash
cd ..

if [ ! -d ./bin ]; then
	exit
fi

cd bin

if [ ! -d ./release ]; then
	mkdir release
fi

7z a -tzip ./release/Ikemen_GO.zip ./external ../data ../font License.txt 'IkemenGO_x86.exe' 'IkemenGO.exe' Ikemen_GO.command Ikemen_GO_Mac Ikemen_GO_Linux