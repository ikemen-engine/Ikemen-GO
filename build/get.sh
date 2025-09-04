#!/bin/bash
# ---------------------------------------------------------------------------
# Dependency fetch script for Ikemen GO.
# Downloads Go module dependencies required for building the project.
# Should be run before build scripts to populate go.mod/go.sum.
# Expects Go installed and network access.
# ---------------------------------------------------------------------------

cd ..  # Move to repository root so module paths resolve.
export CGO_ENABLED=1  # Enable CGO for C-backed libraries. [PITFALL] Requires a C toolchain.

echo "Downloading dependencies..."
echo ""

if [ ! -f ./go.mod ]; then
        go mod init github.com/ikemen-engine/Ikemen-GO/src  # Initialize go module if missing.
        echo ""
fi

# Retrieve dependencies (-u updates to latest).
go get -v -u github.com/samhocevar/beep

go get -v -u github.com/flopp/go-findfont

go get -v -u github.com/go-gl/gl/v2.1/gl

go get -v -u github.com/go-gl/glfw/v3.3/glfw

go get -v -u github.com/ikemen-engine/glfont

go get -v -u github.com/sqweek/dialog

go get -v -u github.com/yuin/gopher-lua

go get -v -u github.com/golang/freetype  # Font rendering support.
