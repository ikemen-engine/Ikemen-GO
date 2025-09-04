#!/bin/bash
# -----------------------------------------------------------------------------
# Ikemen GO release build script.
# Builds platform-specific binaries (Windows, macOS, Linux) using Go's cross-
# compilation. Optionally takes a target OS as the first argument.
# Expects Go toolchain, C cross-compilers and CGO dependencies to be installed.
# -----------------------------------------------------------------------------

# Exit in case of failure to avoid partially built artifacts.
set -e

# Internal vars
binName="Default"
targetOS=$1
currentOS="Unknown"

# Go to the repository root so relative paths resolve correctly.
cd "$(dirname "$0")/.."

# Main function.
function main() {
        # Enable CGO to allow compilation of C dependent packages.
        export CGO_ENABLED=1  # [PITFALL] Requires a working C compiler.

        # Ensure output folder exists for produced binaries.
        mkdir -p bin

        # Determine the host OS if none passed in.
        checkOS $targetOS
        if [[ "$1" == "" ]]; then
                targetOS=$currentOS
        fi

        # Invoke build function for chosen target.
        case "${targetOS}" in
                [wW][iI][nN]64)
                        varWin64      # Set env vars for Windows x64 build.
                        buildWin
                ;;
                [wW][iI][nN]32)
                        varWin32      # Set env vars for Windows x86 build.
                        buildWin
                ;;
                [mM][aA][cC][oO][sS])
                        varMacOS      # Set env vars for macOS build.
                        build
                ;;
                [lL][iI][nN][uU][xX][aA][rR][mM])
                        varLinuxARM   # Set env vars for Linux ARM64 build.
                        build
                ;;
                [lL][iI][nN][uU][xX])
                        varLinux      # Set env vars for Linux build.
                        build
                ;;
        esac

        if [[ "${binName}" == "Default" ]]; then
                echo "Invalid target architecture \"${targetOS}\"."
                exit 1
        fi
}

# Export Variables
function varWin32() {
        export GOOS=windows         # Target OS for Go compiler.
        export GOARCH=386           # 32-bit architecture.
        if [[ "${currentOS,,}" != "win32" ]]; then
                export CC=i686-w64-mingw32-gcc   # [PITFALL] Needs MinGW cross-compiler.
                export CXX=i686-w64-mingw32-g++  # "
        fi
        binName="Ikemen_GO_x86.exe"
}

function varWin64() {
        export GOOS=windows
        export GOARCH=amd64
        if [[ "${currentOS,,}" != "win64" ]]; then
                export CC=x86_64-w64-mingw32-gcc   # [PITFALL] Needs MinGW cross-compiler.
                export CXX=x86_64-w64-mingw32-g++
        fi
        binName="Ikemen_GO.exe"
}

function varMacOS() {
        export GOOS=darwin
        case "${currentOS}" in
                [mM][aA][cC][oO][sS])
                        export CC=clang           # Use native toolchain when on macOS.
                        export CXX=clang++
                ;;
                *)
                        export CC=o64-clang       # [PITFALL] Requires osxcross toolchain for cross builds.
                        export CXX=o64-clang++
                ;;
        esac
        binName="Ikemen_GO_MacOS"
}
function varLinux() {
        export GOOS=linux
        #export CC=gcc   # Uncomment to force specific compiler.
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
        # Compile the Go sources using pre-set GOOS/GOARCH.
        go build -trimpath -v -trimpath -o ./bin/$binName ./src  # Uses GOOS/GOARCH/CC/CXX. [PITFALL] CGO toolchain required.
}

function buildWin() {
        # Windows builds require GUI subsystem flag.
        go build -trimpath -v -trimpath -ldflags "-H windowsgui" -o ./bin/$binName ./src  # Uses GOOS/GOARCH/CC/CXX.
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
                msys|cygwin)
                        if [[ "$osArch" == "x86_64" ]]; then
                                currentOS="Win64"
                        else
                                currentOS="Win32"
                        fi
                ;;
                *)
                        if [[ "$1" == "" ]]; then
                                echo "Unknown system \"${OSTYPE}\"."
                                exit 1
                        fi
                ;;
        esac
}

# Check if "go.mod" exists to verify dependencies.
if [ ! -f ./go.mod ]; then
        echo "Missing dependencies, please run \"get.sh\"."
        exit 1
else
        # Exec Main
        main $1 $2
fi
