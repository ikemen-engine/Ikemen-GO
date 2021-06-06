#!/bin/bash

# Int vars
binName="Default"
cmpt=0
targetOS=$1

# Go to the main folder.
cd ..

# Main function.
function main() {
	# Enable CGO.
	export CGO_ENABLED=1

	# Create "bin" folder.
	if [ ! -d ./bin ]; then
		mkdir bin
	fi
	
	# CMPT flag.
	if [[ "${1,,}" == "cmpt"  ]] || [[ "${2,,}" == "cmpt"  ]]; then
		cmpt=1
	fi

	# Check OS if a build target has not been specified.
	if [[ "$1" == "" ]] || [[ "${1,,}" == "cmpt" ]]; then
		checkOS
	fi
	
	# Build
	case "${targetOS,,}" in
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

	if [[ "${binName}" != "Default" ]]; then
		# Mark file as executable.
		chmod +x ./bin/$binName
	else
		echo "Invalid target architecture \"${targetOS}\".";
		exit 1
	fi
}

# Export Variables
function varWin32() {
	export GOOS=windows
	export GOARCH=386
	export CC=i686-w64-mingw32-gcc
	export CXX=i686-w64-mingw32-g++
	binName="Ikemen_GO_x86.exe"
}

function varWin64() {
	export GOOS=windows
	export CC=x86_64-w64-mingw32-gcc
	export CXX=x86_64-w64-mingw32-g++
	binName="Ikemen_GO.exe"
}

function varMacOS() {
	export GOOS=darwin
	export CC=o64-clang 
	export CXX=o64-clang++
	binName="Ikemen_GO_MacOS"
}
function varLinux() {
	export GOOS=linux
	#export CC=gcc
	#export CXX=g++
	binName="Ikemen_GO_Linux"
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

# Determine the target OS.
function checkOS() {
	osArch=`uname -m`
	case "$OSTYPE" in
		darwin*)
			targetOS="MacOS"
		;;
		linux*) 
			targetOS="Linux"
		;;
		msys)
			if [[ "$osArch" == "x86_64" ]];then
				targetOS="Win64"
			else
				targetOS="Win32"
			fi
		;;
		*)
			echo "Unknow system \"${OSTYPE}\".";
			exit 1
		;;
	esac
}

# Check if "go.mod" exists.
if [ ! -f ./go.mod ]; then
	echo "Missing dependencies, please run \"get.sh\"."
	exit 1
else
	# Exec Main
	main $1 $2
fi
