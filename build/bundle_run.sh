#!/bin/bash

# Get the directory of this script (MacOS directory)
SCRIPT_DIR="$(dirname "$0")"

# Determine the directory containing the .app bundle
APP_DIR="$(cd "$SCRIPT_DIR/../../" && pwd)"

# Define the path to the app executable relative to the MacOS directory
APP_EXEC="$SCRIPT_DIR/Ikemen_GO_MacOS"

# Output for debugging
echo "SCRIPT_DIR: $SCRIPT_DIR"
echo "APP_DIR: $APP_DIR"
echo "APP_EXEC: $APP_EXEC"

# Check if the executable exists
if [ ! -x "$APP_EXEC" ]; then
    echo "Executable $APP_EXEC not found or not executable"
    exit 1
fi

# Change directory to the parent directory of the .app bundle
cd "$APP_DIR/../" || {
    echo "Failed to change directory to $APP_DIR/../"
    exit 1
}

# Output the current working directory for debugging
echo "Current working directory: $(pwd)"

# Launch the macOS app executable
"$APP_EXEC" "$@" -AppleMagnifiedMode YES