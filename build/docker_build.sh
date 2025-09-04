#!/bin/bash
# ---------------------------------------------------------------------------
# Docker-based release build script.
# Uses the windblade/ikemen-go-dev image to cross-compile Ikemen GO for
# Linux, macOS and Windows targets from a Linux host.
# Requires Docker and network access.
# ---------------------------------------------------------------------------

###############################################################################
# Existing note: Simple script that uses a docker image to build the binaries of Ikemen_GO for
# Windows/Linux/OSX platforms. The only dependencies are the source code of Ikemen_GO_plus and Docker.
# @author Daniel Porto
# https://github.com/danielporto
###############################################################################

# Parameters explained:
#  run  : download and execute the docker container with the building tools
#  --rm : discard the container after using it. It saves disk space
#  -e   : set environment variables used by the scripts called inside the container
#         these variables select the cross-compiling parameters invoked.
#         Look inside the get.sh and build.sh for details.
#  -v   : maps a volume (folder) inside the  container (makes the current source code accessible inside the container)
#         $(pwd):/ikemen is source:destination and $(pwd) maps to current directory where the script is called.
#  -i   : interactive.
#  -t   : allocate a pseudo terminal.
#  windblade/ikemen-go-dev:latest                       : docker image configured with the tooling required to build the binaries.
#  bash -c 'bash -c 'cd /ikemen/build && bash build.sh' : command called when the container launches. It changes to the code directory
#  then execute both get and build scripts

cd ..  # Switch to repository root.

# Download Docker image
# [PITFALL] Requires Docker daemon running and internet access.
docker pull windblade/ikemen-go-dev:latest

# Create directories
if [ ! -d ./bin ]; then
        mkdir bin  # Ensure output directory exists.
fi

echo "------------------------------------------------------------"
echo "Starting Build of Ikemen GO"

echo "------------------------------------------------------------"
echo "Building Linux binary..."
docker run --rm -v $(pwd):/ikemen -i windblade/ikemen-go-dev:latest bash -c 'cd /ikemen/build && bash build.sh Linux'  # GOOS=linux inside container.

echo "------------------------------------------------------------"
echo "Building MacOS binary..."
docker run --rm -v $(pwd):/ikemen -i windblade/ikemen-go-dev:latest bash -c 'cd /ikemen/build && bash build.sh MacOS'  # GOOS=darwin.

echo "------------------------------------------------------------"
echo "Building Windows x64 binary..."
cp 'windres/Ikemen_Cylia_x64.syso' 'src/Ikemen_V2_x64.syso'  # Provide Windows resources.

docker run --rm -v $(pwd):/ikemen -i windblade/ikemen-go-dev:latest bash -c 'cd /ikemen/build && bash build.sh Win64'

rm 'src/Ikemen_V2_x64.syso'

echo "------------------------------------------------------------"
echo "Building Windows x86 binary..."
cp 'windres/Ikemen_Cylia_x86.syso' 'src/Ikemen_V2_x86.syso'

docker run --rm -v $(pwd):/ikemen -i windblade/ikemen-go-dev:latest bash -c 'cd /ikemen/build && bash build.sh Win32'

rm 'src/Ikemen_V2_x86.syso'

echo "------------------------------------------------------------"
echo "Finished"
echo "------------------------------------------------------------"
