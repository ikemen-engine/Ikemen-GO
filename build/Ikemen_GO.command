#!/bin/bash
# ---------------------------------------------------------------------------
# Cross-platform launcher for Ikemen GO.
# Determines the host OS and runs the matching binary, applying any required
# environment tweaks.
# ---------------------------------------------------------------------------

cd $(dirname $0)  # Move to directory containing binaries.

case "$OSTYPE" in
        darwin*)
                # macOS: ensure binary is executable then launch with HiDPI flag.
                chmod +x Ikemen_GO_MacOS  # [PITFALL] Unsigned binaries may be blocked by Gatekeeper.
                ./Ikemen_GO_MacOS -AppleMagnifiedMode YES
        ;;
        linux*)
                # Linux: optional Mesa overrides for compatibility with older drivers.
                #export MESA_GL_VERSION_OVERRIDE=2.1
                #export MESA_GLES_VERSION_OVERRIDE=1.5
                chmod +x Ikemen_GO_Linux
                ./Ikemen_GO_Linux
        ;;
        *)
                echo "System not recognized"  # [PITFALL] Unsupported platform.
                exit 1
        ;;
esac
