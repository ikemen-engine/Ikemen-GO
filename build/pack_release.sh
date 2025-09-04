#!/bin/bash
# ---------------------------------------------------------------------------
# Package release script.
# Bundles compiled binaries and required assets into a zip archive for
# distribution. Run after building the binaries.
# Requires the 7z (p7zip) tool to be installed.
# ---------------------------------------------------------------------------

cd ..  # Move to repository root.

if [ ! -d ./bin ]; then
        exit  # No binaries to package.
fi

cd bin

if [ ! -d ./release ]; then
        mkdir release  # Create output directory if absent.
fi

# Archive binaries and resources into release package.
7z a -tzip ./release/Ikemen_GO.zip ./external ../data ../font License.txt 'IkemenGO_x86.exe' 'IkemenGO.exe' Ikemen_GO.command Ikemen_GO_Mac Ikemen_GO_Linux  # [PITFALL] Requires 7z in PATH.
