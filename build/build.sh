#!/bin/bash

# Exit in case of failure
set -e

# Int vars
binName="Default"
targetOS=$1
currentOS="Unknown"

# Go to the main folder.
cd "$(dirname "$0")/.."

# Main function.
function main() {
	# Enable CGO.
	export CGO_ENABLED=1

	# Create "bin" folder.
	mkdir -p bin

	# Check OS
	checkOS
	# If a build target has not been specified use the current OS.
	if [[ "$1" == "" ]]; then
		targetOS=$currentOS
	fi
	
	# Build
	case "${targetOS}" in
		[wW][iI][nN]64)
			varWin64
			buildWin
		;;
		[wW][iI][nN]32)
			varWin32
			buildWin
		;;
		[mM][aA][cC][oO][sS])
			varMacOS
			build
		;;
		[lL][iI][nN][uU][xX][aA][rR][mM])
			varLinuxARM
			build
		;;
		[lL][iI][nN][uU][xX])
			varLinux
			build
		;;
	esac

	if [[ "${binName}" == "Default" ]]; then
		echo "Invalid target architecture \"${targetOS}\".";
		exit 1
	fi
}

# Export Variables
function varWin32() {
	export GOOS=windows
	export GOARCH=386
	if [[ "${currentOS,,}" != "win32" ]]; then
		export CC=i686-w64-mingw32-gcc
		export CXX=i686-w64-mingw32-g++
	fi
	binName="Ikemen_GO_x86.exe"
}

function varWin64() {
	export GOOS=windows
	export GOARCH=amd64
	if [[ "${currentOS,,}" != "win64" ]]; then
		export CC=x86_64-w64-mingw32-gcc
		export CXX=x86_64-w64-mingw32-g++
	fi
	binName="Ikemen_GO.exe"
}

function varMacOS() {
	export GOOS=darwin
	case "${currentOS}" in
		[mM][aA][cC][oO][sS])
			export CC=clang
			export CXX=clang++
		;;
		*)
			export CC=o64-clang
			export CXX=o64-clang++
		;;
	esac
	binName="Ikemen_GO_MacOS"
}
function varLinux() {
	export GOOS=linux
	#export CC=gcc
	#export CXX=g++
	binName="Ikemen_GO_Linux"
}
function varLinuxARM() {
	export GOOS=linux
	export GOARCH=arm64
	binName="Ikemen_GO_LinuxARM"
}

# Build functions.
function build() {
	#echo "buildNormal"
	#echo "$binName"
	go build -trimpath -v -trimpath -o ./bin/$binName ./src
}

function buildWin() {
	#echo "buildWin"
	#echo "$binName"
	go build -trimpath -v -trimpath -ldflags "-H windowsgui" -o ./bin/$binName ./src
}

# Determine the target OS.
function checkOS() {
	osArch=`uname -m`
	case "$OSTYPE" in
		darwin*)
			currentOS="MacOS"
		;;
		linux*)
			currentOS="Linux"
		;;
		msys)
			if [[ "$osArch" == "x86_64" ]]; then
				currentOS="Win64"
			else
				currentOS="Win32"
			fi
		;;
		*)
			if [[ "$1" == "" ]]; then
				echo "Unknown system \"${OSTYPE}\".";
				exit 1
			fi
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
