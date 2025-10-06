#!/bin/bash
cd "$(dirname "$0")"

case "$OSTYPE" in
	darwin*)
		# Prefer the .app if present right here
		APPDIR="./I.K.E.M.E.N-Go.app"
		if [ -d "$APPDIR" ]; then
			xattr -d com.apple.quarantine "$APPDIR" 2>/dev/null || true
			chmod +x "$APPDIR/Contents/MacOS/bundle_run.sh" 2>/dev/null || true
			exec "$APPDIR/Contents/MacOS/bundle_run.sh"
		fi
		# Fallbacks: binaries in ./bin or current dir
		for BIN in "./bin/Ikemen_GO_MacOSARM" "./bin/Ikemen_GO_MacOS" \
		           "./Ikemen_GO_MacOSARM" "./Ikemen_GO_MacOS"; do
			if [ -x "$BIN" ]; then
				chmod +x "$BIN" 2>/dev/null || true
				exec "$BIN"
			fi
		done
		echo "Unable to locate I.K.E.M.E.N-Go.app or a macOS binary in the top directory." >&2
		exit 1
	;;
	linux*)
		# Prefer binary in current dir; accept ./bin as secondary
		for BIN in "./Ikemen_GO_Linux" "./bin/Ikemen_GO_Linux"; do
			if [ -x "$BIN" ]; then
				chmod +x "$BIN" 2>/dev/null || true
				exec "$BIN"
			fi
		done
		echo "Ikemen_GO_Linux not found in the top directory. Build it with: ./build/build.sh Linux" >&2
		exit 1
	;;
	*)
		echo "System not recognized"
		exit 1
	;;
esac
