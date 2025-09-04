#!/bin/bash
# ---------------------------------------------------------------------------
# Helper run script for the macOS app bundle.
# Locates the Ikemen_GO_MacOS binary inside the bundle and launches it
# so that relative paths to game assets resolve correctly.
# ---------------------------------------------------------------------------

# Directory of this script within the .app bundle's MacOS/ folder.
SCRIPT_DIR="$(dirname "$0")"

# Directory containing the entire .app bundle.
APP_DIR="$(cd "$SCRIPT_DIR/../../" && pwd)"

# Path to the compiled executable relative to this script.
APP_EXEC="$SCRIPT_DIR/Ikemen_GO_MacOS"

# Output for debugging
echo "SCRIPT_DIR: $SCRIPT_DIR"
echo "APP_DIR: $APP_DIR"
echo "APP_EXEC: $APP_EXEC"

# Ensure the binary exists and is executable.
if [ ! -x "$APP_EXEC" ]; then
    echo "Executable $APP_EXEC not found or not executable"  # [PITFALL] Binary may be quarantined on macOS.
    exit 1
fi

# Change directory to the parent directory of the .app bundle
cd "$APP_DIR/../" || {
    echo "Failed to change directory to $APP_DIR/../"
    exit 1
}

# Output the current working directory for debugging
echo "Current working directory: $(pwd)"

# Launch the macOS app executable, forwarding any CLI args.
"$APP_EXEC" "$@" -AppleMagnifiedMode YES
