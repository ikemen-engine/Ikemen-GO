#!/bin/bash
binName="IkemenGO-generic"
cmpt=0

# Main function.
function main() {
	# Go to the main folder.
	cd ..
	# Enable CGO.
	export CGO_ENABLED=1
	
	# CMPT flag.
	if [[ "${2,,}" == "cmpt"  ]]; then
		cmpt=1
	fi
	
	# Build
	case "${1,,}" in
		"win64")
			varWin64
			buildWin
		;;
		"win32")
			varWin32
			buildWin
		;;
		"macos")
			varMacOS
			buildAlt
		;;
		"linux")
			varLinux
			if [[ cmpt -eq 1 ]]; then
				buildAlt
			else
				build
			fi
		;;
	esac

	# Mark file as executable.
	chmod +x ./bin/$binName
}

# Export Variables
function varWin32() {
	export GOOS=windows
	export GOARCH=386
	export CC=i686-w64-mingw32-gcc
	export CXX=i686-w64-mingw32-g++
	binName="IkemenGO_x86.exe"
}

function varWin64() {
	export GOOS=windows
	export CC=x86_64-w64-mingw32-gcc
	export CXX=x86_64-w64-mingw32-g++
	binName="IkemenGO.exe"
}

function varMacOS() {
	export GOOS=darwin
	export CC=o64-clang 
	export CXX=o64-clang++
	binName="IkemenGO_mac"
}
function varLinux() {
	export GOOS=linux
	#export CC=gcc
	#export CXX=g++
	binName="IkemenGO_linux"
}

# Build functions.
function build() {
	#echo "buildNormal"
	#echo "$binName"
	go build -o ./bin/$binName ./src
}

function buildAlt() {
	#echo "buildAlt"
	#echo "$binName"
	go build -tags al_cmpt -o ./bin/$binName ./src
}

function buildWin() {
	#echo "buildWin"
	#echo "$binName"
	go build -ldflags "-H windowsgui" -o ./bin/$binName ./src
}

# Exec Main
main $1 $2
