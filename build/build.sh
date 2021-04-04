#!/bin/sh
cd ..
export CGO_ENABLED=1

case "$OSTYPE" in
	darwin*)
		BINARY_NAME="Ikemen_GO_mac"
	;;
	linux*) 
		BINARY_NAME="Ikemen_GO_linux" 
	;;
	*)
		echo "System not recognized";
		exit 1
	;;
esac

if [ ! -f ./go.mod ]; then
	echo "Missing dependencies, please run get.sh"
	exit
fi
if [ ! -d ./bin ]; then
	mkdir bin
fi

go build -i -o ./bin/$BINARY_NAME ./src
chmod +x ./bin/$BINARY_NAME

cp ./build/Ikemen_GO.command ./bin/Ikemen_GO.command
cp -r ./external/ ./bin/external/
cp -r ./data/ ./bin/data/
cp -r ./font/ ./bin/font/
