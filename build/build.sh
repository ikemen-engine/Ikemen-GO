#!/bin/bash

# Exit in case of failure; print function-trap friendly errors
set -o errtrace
set -euo pipefail

# Bash 3.2 (macOS/GitHub runners) does NOT support ${var,,}.
# Use a portable helper instead.
tolower() {
  # no trailing newline
  printf '%s' "$1" | tr '[:upper:]' '[:lower:]'
}

# Portable in-place sed (GNU vs BSD sed)
sed_inplace() {
  local expr="$1" file="$2"
  if sed --version >/dev/null 2>&1; then
    # GNU sed
    sed -i "$expr" "$file"
  else
    # BSD/macOS sed
    sed -i '' "$expr" "$file"
  fi
}

# For Android builds:
# - If vendor/ exists, force -mod=vendor so we use the vendored (and patchable) copy.
# - Otherwise, enable -modcacherw so the module cache can be patched (if needed).
ensure_go_flags_android() {
  [[ "${GOOS:-}" != "android" ]] && return 0

  if [[ -f "$REPO_ROOT/vendor/modules.txt" ]]; then
    if [[ " ${GOFLAGS:-} " != *" -mod=vendor "* ]]; then
      export GOFLAGS="${GOFLAGS:-} -mod=vendor"
    fi
  else
    if [[ " ${GOFLAGS:-} " != *" -modcacherw "* ]]; then
      export GOFLAGS="${GOFLAGS:-} -modcacherw"
    fi
  fi
}

## Resolve repo root _first_ so all path defaults are stable no matter where
## the script is invoked from (top dir, build dir, or elsewhere).
REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd -P)"
cd "$REPO_ROOT"

# Fail fast on unsafe repo paths (MSYS2/Cygwin): only allow [A-Za-z0-9._/-]
require_safe_path() {
  case "$OSTYPE" in
	msys|cygwin)
	  if [[ "$REPO_ROOT" =~ [^A-Za-z0-9._/-] ]]; then
		echo "ERROR: Repository path contains characters unsafe for MSYS2/autotools:" >&2
		echo "  $REPO_ROOT" >&2
		echo "Use only letters, digits, '_', '-', '.'" >&2
		echo "Tip: move to something like: C:\dev\Ikemen-GO" >&2
		exit 1
	  fi
	  if ((${#REPO_ROOT} > 220)); then
		echo "WARNING: Very long path (${#REPO_ROOT} chars) may hit Windows MAX_PATH issues." >&2
	  fi
	  ;;
  esac
}

# Int vars
DEBUG_BUILD="${DEBUG_BUILD:-0}"
binName="Default"
targetOS="${1:-}"
currentOS="Unknown"
OUTDIR="bin"         # may be overridden below
LIBDIR="lib"         # runtime libs live here at repo root
BUILDDIR="build"
DELAYLIB_DIR="$BUILDDIR/delaylib"
FFMPEG_SRCDIR="$BUILDDIR/ffmpeg-src"

# Elecbyte screenpack assets
SCREENPACK_REPO="${SCREENPACK_REPO:-https://github.com/ikemen-engine/Ikemen_GO-Elecbyte-Screenpack.git}"
SCREENPACK_REF="${SCREENPACK_REF:-master}"
SCREENPACK_DIR="$REPO_ROOT/$BUILDDIR/elecbyte-screenpack"

# Android APK packaging (ikemen-droid)
BUILD_ANDROID_APK="${BUILD_ANDROID_APK:-1}"  # 1=yes, 0=no
# ANDROID_APK_REPO can be a git URL (default) OR a local directory path to an existing ikemen-droid checkout.
ANDROID_APK_REPO="${ANDROID_APK_REPO:-https://github.com/Jesuszilla/ikemen-droid.git}"
ANDROID_APK_REF="${ANDROID_APK_REF:-main}"
ANDROID_APK_DIR="$REPO_ROOT/$BUILDDIR/android-apk/ikemen-droid"
ANDROID_APK_OUT="${ANDROID_APK_OUT:-$REPO_ROOT/bin/ikemen-go.apk}"

# FFmpeg config
FFMPEG_REV="${FFMPEG_REV:-release/7.1}"
# Always default to build/ffmpeg under the repo root (absolute path).
FFMPEG_PREFIX="${FFMPEG_PREFIX:-$REPO_ROOT/$BUILDDIR/ffmpeg}"
BUILD_FFMPEG="${BUILD_FFMPEG:-auto}"   # auto|yes|no

# ---- App metadata (overridden by CI)
APP_VERSION="${APP_VERSION:-nightly}"
APP_BUILDTIME="${APP_BUILDTIME:-$(date '+%Y.%m.%d')}"

# --- Always pause on Windows on exit (covers success and failure) ----------
pause_always_windows() {
	local status="${1:-0}"
	case "$OSTYPE" in
		msys|cygwin)
		# Don't pause in CI environments (prevents hanging GitHub Actions).
		if [[ -n "${CI:-}" || -n "${GITHUB_ACTIONS:-}" ]]; then
			return 0
		fi
		# Choose message by exit status
		local msg
		if [[ "$status" -eq 0 ]]; then
			msg="Build finished successfully. Press any key to close..."
		else
			msg="Build failed (exit $status). Press any key to close..."
		fi
		# If stdin/stdout are terminals, just read a key.
		if [[ -t 0 && -t 1 ]]; then
			printf "%s" "$msg" ; IFS= read -r -n1 _ ; echo
			return 0
		fi
	;;
	esac
}
trap 'st=$?; pause_always_windows "$st"' EXIT

# Ensure Go works (especially under MSYS2 where GOROOT may be required)
ensure_go_env() {
	# Prefer MinGW bin early on PATH in MSYS shells
	case "$OSTYPE" in
		msys|cygwin)
			if [[ ":${PATH}:" != *":/mingw64/bin:"* ]]; then
				export PATH="/mingw64/bin:${PATH}"
			fi
			# Set a sensible default GOROOT for MSYS2 trimmed Go if not set
			if [[ -z "${GOROOT:-}" ]] && [[ -d "/mingw64/lib/go" ]]; then
				export GOROOT="/mingw64/lib/go"
			fi
		;;
	esac

	# Verify we can run go and it knows its GOROOT
	if ! command -v go >/dev/null 2>&1; then
		echo "ERROR: 'go' not found on PATH. On MSYS2 install 'mingw-w64-x86_64-go' and use the MINGW64 shell." >&2
		exit 1
	fi
	if ! go version >/dev/null 2>&1; then
		echo "ERROR: 'go version' failed. On MSYS2 you may need:  export GOROOT=/mingw64/lib/go" >&2
		echo "Current GOROOT='${GOROOT:-<unset>}' PATH='${PATH}'" >&2
		exit 1
	fi
}

# Dependency preflight (prints actionable commands)
check_deps() {
	local missing=()
	need() { command -v "$1" >/dev/null 2>&1 || missing+=("$1"); }

	case "$OSTYPE" in
		msys|cygwin)
			# MSYS2/MINGW64
			need git
			need make
			need pkg-config
			need gcc
			need g++
			need nasm
			need go
			need gendef
			need dlltool
			if ((${#missing[@]})); then
				echo "Missing package: ${missing[@]}" >&2
				echo "Install from the MINGW64 shell:" >&2
				echo "  pacman -Syu --noconfirm" >&2
				echo "  pacman -S --noconfirm \\" >&2
				echo "    git make diffutils mingw-w64-x86_64-pkg-config \\" >&2
				echo "    mingw-w64-x86_64-go mingw-w64-x86_64-toolchain \\" >&2
				echo "    mingw-w64-x86_64-nasm mingw-w64-x86_64-yasm \\" >&2
				echo "    mingw-w64-x86_64-tools-git mingw-w64-x86_64-libxmp \\" >&2
				echo "    mingw-w64-x86_64-SDL2" >&2
				exit 1
			fi
		;;
		darwin*)
			need git
			need pkg-config
			need clang
			need clang++
			need nasm
			need go
			# yasm not strictly required when using NASM, but list it if missing for parity
			command -v yasm >/dev/null 2>&1 || echo "Note: yasm not found (nasm present). If build fails, try: brew install yasm" >&2
			if ((${#missing[@]})); then
				echo "ERROR: Missing tools: ${missing[*]}" >&2
				echo "Install with Homebrew:" >&2
				echo "  brew update && brew install git go pkg-config nasm libxmp sdl2 molten-vk" >&2
				exit 1
			fi
		;;
		linux*)
			need git
			need pkg-config
			need gcc
			need g++
			need make
			need nasm
			need yasm
			need go
			if ((${#missing[@]})); then
				echo "ERROR: Missing tools: ${missing[*]}" >&2
				echo "Install (Debian/Ubuntu):" >&2
				echo "  sudo apt update && sudo apt install -y golang-go git pkg-config make nasm yasm build-essential libxmp-dev libsdl2-dev" >&2
				exit 1
			fi
		;;
	esac
}

# Ensure SDL2 (via pkg-config) is available
require_sdl2() {
	local pc="${PKG_CONFIG:-pkg-config}"
	if ! $pc --exists sdl2 2>/dev/null; then
	case "$OSTYPE" in
		msys|cygwin)
			echo "ERROR: SDL2 dev package not found. Install: pacman -S mingw-w64-x86_64-SDL2" >&2
			;;
		darwin*)
			echo "ERROR: SDL2 not found. Install with Homebrew: brew install sdl2" >&2
			;;
		linux*)
			echo "ERROR: SDL2 dev package not found. Install (Debian/Ubuntu): sudo apt install -y libsdl2-dev" >&2
			;;
		*)
			echo "ERROR: SDL2 dev package not found (pkg-config module 'sdl2')." >&2
			;;
		esac
		exit 1
	fi
}

# Ensure libxmp (via pkg-config) is available
require_libxmp() {
	local pc="${PKG_CONFIG:-pkg-config}"
	if ! $pc --exists libxmp 2>/dev/null; then
	case "$OSTYPE" in
		msys|cygwin)
			echo "ERROR: libxmp dev package not found. Install: pacman -S mingw-w64-x86_64-libxmp" >&2
			;;
		darwin*)
			echo "ERROR: libxmp not found. Install with Homebrew: brew install libxmp" >&2
			;;
		linux*)
			echo "ERROR: libxmp dev package not found. Install (Debian/Ubuntu): sudo apt install -y libxmp-dev" >&2
			;;
		*)
			echo "ERROR: libxmp dev package not found (pkg-config module 'libxmp')." >&2
			;;
		esac
		exit 1
	fi
}

# MoltenVK (macOS Vulkan) detection
MVK_DYLIB=""
MVK_LIB_DIR=""
find_moltenvk() {
	local candidates=()
	# Prefer Homebrew paths if available
	if command -v brew >/dev/null 2>&1; then
		local bp; bp="$(brew --prefix 2>/dev/null || true)"
		[[ -n "$bp" ]] && candidates+=("$bp/opt/molten-vk/lib/libMoltenVK.dylib")
		local mp; mp="$(brew --prefix molten-vk 2>/dev/null || true)"
		[[ -n "$mp" ]] && candidates+=("$mp/lib/libMoltenVK.dylib")
	fi
	# Common fallbacks (Apple Silicon / Intel)
	candidates+=("/opt/homebrew/lib/libMoltenVK.dylib" "/usr/local/lib/libMoltenVK.dylib")
	for p in "${candidates[@]}"; do
		if [[ -f "$p" ]]; then
			MVK_DYLIB="$p"
			MVK_LIB_DIR="$(dirname "$p")"
			return 0
		fi
	done
	return 1
}

require_moltenvk() {
	[[ "$GOOS" != "darwin" ]] && return 0
	if find_moltenvk; then
		# Ensure clang can find MoltenVK when CGO links
		if [[ "${CGO_LDFLAGS:-}" != *"-L$MVK_LIB_DIR"* ]]; then
			export CGO_LDFLAGS="${CGO_LDFLAGS:-} -L$MVK_LIB_DIR"
		fi
	else
		echo "ERROR: MoltenVK not found (required for the Vulkan renderer on macOS)." >&2
		echo "Install with Homebrew:" >&2
		echo "  brew install molten-vk" >&2
		exit 1
	fi
}

# Main function.
function main() {
	# Enable CGO.
	export CGO_ENABLED=1

	# Enable arenas (required for rollback)
	export GOEXPERIMENT=arenas

	# Decide output location:
	# - Non-macOS: binary in top-level (.) and runtime libs in ./lib
	# - macOS: app bundle uses bin/ later from Makefile
	# - Android: ALWAYS put outputs in bin/ (so we get bin/libmain.so + bin/libmain.h)
	case "$(tolower "${targetOS}")" in
		android) OUTDIR="bin" ;;
		*)
			case "$OSTYPE" in
				darwin*) OUTDIR="bin" ;;
				*)       OUTDIR="."  ;;
			esac
		;;
	esac
	mkdir -p "$OUTDIR"
	mkdir -p "$LIBDIR"

	# Make sure Go toolchain is usable
	ensure_go_env

	# Make sure required build tools exist
	check_deps

	# Enforce safe path
	require_safe_path

	# Check OS
	checkOS "$targetOS"
	# If a build target has not been specified use the current OS.
	if [[ -z "$targetOS" ]]; then
		targetOS="$currentOS"
	fi
	
	# Build
	case "${targetOS}" in
		[wW][iI][nN]64)
			: "${CC:=x86_64-w64-mingw32-gcc}"
			: "${CXX:=x86_64-w64-mingw32-g++}"
			varWin64
			buildWin
		;;
		[wW][iI][nN]32)
			: "${CC:=i686-w64-mingw32-gcc}"
			: "${CXX:=i686-w64-mingw32-g++}"
			varWin32
			buildWin
		;;
		[mM][aA][cC][oO][sS][aA][rR][mM])
			varMacOSARM
			build
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
		[aA][nN][dD][rR][oO][iI][dD])
			varAndroid
			build
		;;
		*)
			echo "Unknown target: ${targetOS}"
			echo "Valid targets: Win64 Win32 MacOS MacOSARM Linux LinuxARM"
			exit 1
		;;
	esac
}

# Export Variables
function varWin32() {
	export GOOS=windows
	export GOARCH=386
	if [[ "$(tolower "${currentOS}")" != "win32" ]]; then
		export CC=i686-w64-mingw32-gcc
		export CXX=i686-w64-mingw32-g++
	fi
	binName="Ikemen_GO_x86.exe"
}
 
function varWin64() {
	export GOOS=windows
	export GOARCH=amd64
	if [[ "$(tolower "${currentOS}")" != "win64" ]]; then
		export CC=x86_64-w64-mingw32-gcc
		export CXX=x86_64-w64-mingw32-g++
	fi
	binName="Ikemen_GO.exe"
}

function varMacOSARM() {
	export GOOS=darwin
	export GOARCH=arm64
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
	binName="Ikemen_GO_MacOSARM"
}
function varMacOS() {
	export GOOS=darwin
	export GOARCH=amd64
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
function varAndroid() {
	local host_os="linux" # default to Linux as that's what the runner will be using
	[[ "$OSTYPE" == "darwin"* ]] && host_os="darwin"
	export GOOS=android
	export GOARCH=arm64
	export CGO_ENABLED=1
	# ANDROID_NDK_HOME is an environment variable as its path can be different per platform.
	export TOOLCHAIN="$ANDROID_NDK_HOME/toolchains/llvm/prebuilt/${host_os}-x86_64"
	export TARGET="aarch64-linux-android34" # for Android 14
	export CC="$TOOLCHAIN/bin/${TARGET}-clang"
	export CXX="$TOOLCHAIN/bin/${TARGET}-clang++"
	export ANDROID_DEPS_PATH="$REPO_ROOT/build/android-deps"
	# Force pkg-config to ONLY look at the Android libraries
	export PKG_CONFIG_LIBDIR="$ANDROID_DEPS_PATH/lib/pkgconfig"
	export PKG_CONFIG_SYSROOT_DIR="$ANDROID_DEPS_PATH"
	# Ensure we don't pick up host libraries by clearing this
	export PKG_CONFIG_PATH=""
	binName="libmain.so"
}

# --- FFmpeg detection / build ---
function have_ffmpeg_pc() {
	local pc="${PKG_CONFIG:-pkg-config}"
	$pc --exists libavformat libavcodec libavutil libswresample libswscale libavfilter 2>/dev/null
}

function ensure_pkg_config_path() {
	if [[ -d "$FFMPEG_PREFIX/lib/pkgconfig" ]]; then
		if [[ "${RUNNER_OS:-}" == "Windows" || "$OSTYPE" == msys || "$OSTYPE" == cygwin ]]; then
			export PKG_CONFIG_PATH="$FFMPEG_PREFIX/lib/pkgconfig:/mingw64/lib/pkgconfig:${PKG_CONFIG_PATH:-}"
		else
			export PKG_CONFIG_PATH="$FFMPEG_PREFIX/lib/pkgconfig:${PKG_CONFIG_PATH:-}"
		fi
	fi
}

function build_ffmpeg() {
	echo "==> Building minimal FFmpeg to $FFMPEG_PREFIX (sources in $FFMPEG_SRCDIR)"
	mkdir -p "$BUILDDIR"
	rm -rf "$FFMPEG_SRCDIR"
	git clone --depth=1 -b "$FFMPEG_REV" https://github.com/FFmpeg/FFmpeg.git "$FFMPEG_SRCDIR"
	pushd "$FFMPEG_SRCDIR" >/dev/null
	git checkout "$FFMPEG_REV"

	echo "==> Starting configure..."
	if [[ "$GOOS" == "android" ]]; then
		local arch="aarch64"
		local api_level="34"
		local prefix="$ANDROID_DEPS_PATH"
		
		# Path to the specific Android Clang
		local cc_compiler="$TOOLCHAIN/bin/${arch}-linux-android${api_level}-clang"
		local ar_tool="$TOOLCHAIN/bin/llvm-ar"
		local nm_tool="$TOOLCHAIN/bin/llvm-nm"
		local strip_tool="$TOOLCHAIN/bin/llvm-strip"
		
		local configure_args=(
			"--prefix=$prefix"
			"--enable-cross-compile"
			"--target-os=android" "--arch=$arch"
			"--cc=$cc_compiler" "--ar=$ar_tool" "--nm=$nm_tool" "--strip=$strip_tool"
			"--extra-cflags=-fPIC"
			"--extra-ldflags=-Wl,-z,max-page-size=16384"
			"--enable-shared"
			"--install-name-dir=@rpath"
			"--enable-shared" "--disable-static"
			"--disable-gpl" "--disable-nonfree"
			"--disable-debug" "--disable-doc" "--disable-programs" "--disable-everything"
			"--disable-autodetect"
			"--enable-avformat" "--enable-avcodec" "--enable-avutil" "--enable-swresample" "--enable-swscale"
			"--enable-avfilter" "--enable-filter=buffer,buffersink,format,scale,pad,crop"
			"--enable-protocol=file"
			"--enable-demuxer=matroska,webm"
			"--enable-decoder=vp8,vp9,opus,vorbis"
			"--enable-parser=vp8,vp9,opus,vorbis"
			"--enable-jni" "--enable-mediacodec"
			"--pkg-config=$(which pkg-config)"
		)
	else
		local configure_args=(
			"--prefix=$FFMPEG_PREFIX"
			"--install-name-dir=@rpath"
			"--enable-shared" "--disable-static"
			"--disable-gpl" "--disable-nonfree"
			"--disable-debug" "--disable-doc" "--disable-programs" "--disable-everything"
			"--disable-autodetect"
			"--enable-avformat" "--enable-avcodec" "--enable-avutil" "--enable-swresample" "--enable-swscale"
			"--enable-avfilter" "--enable-filter=buffer,buffersink,format,scale,pad,crop"
			"--enable-protocol=file"
			"--enable-demuxer=matroska,webm"
			"--enable-decoder=vp8,vp9,opus,vorbis"
			"--enable-parser=vp8,vp9,opus,vorbis"
	)
	fi

	# Creates $FFMPEG_SRCDIR/BUILDINFO.txt with revision, configure args, and compiler info.
	local _cc="${CC:-cc}"
	local _cc_path; _cc_path="$(command -v "$_cc" 2>/dev/null || echo "$_cc")"
	local _cc_ver;  _cc_ver="$("$_cc" --version 2>/dev/null | head -n1 || true)"
	local _rev_sha; _rev_sha="$(git rev-parse HEAD 2>/dev/null || true)"
	{
		echo "Revision: ${FFMPEG_REV}${_rev_sha:+ (commit ${_rev_sha})}"
		echo -n "Configured with:"
		printf " %s" "${configure_args[@]}"
		echo
		echo "Compiler: ${_cc_path}${_cc_ver:+ — ${_cc_ver}}"
	} > BUILDINFO.txt

	./configure "${configure_args[@]}"
	echo "==> Configure complete. Starting make..."
	make -j"$(getconf _NPROCESSORS_ONLN || echo 2)"
	echo "==> Build complete. Installing..."
	make install
	# sanity: show the produced pkg-config files so we can see them in logs
	local pcdir="$FFMPEG_PREFIX/lib/pkgconfig"
	if [[ "$GOOS" == "android" ]]; then
		pcdir="$ANDROID_DEPS_PATH/lib/pkgconfig"
	fi
	echo "==> FFmpeg pkg-config files installed to: $pcdir"
	ls -l "$pcdir" || true
	popd >/dev/null
}

function build_libxmp_android() {
	echo "==> Building LibXMP for Android..."
	local src="$BUILDDIR/libxmp-src"
	[[ -d "$src" ]] || git clone https://github.com/libxmp/libxmp.git "$src"
	
	mkdir -p "$src/build-android"
	pushd "$src/build-android" >/dev/null
	
	cmake .. \
		-DCMAKE_TOOLCHAIN_FILE="$ANDROID_NDK_HOME/build/cmake/android.toolchain.cmake" \
		-DANDROID_ABI="arm64-v8a" \
		-DANDROID_PLATFORM=android-21 \
		-DCMAKE_INSTALL_PREFIX="$ANDROID_DEPS_PATH" \
		-DBUILD_STATIC=OFF \
		-DBUILD_SHARED=ON \
		-DCMAKE_SHARED_LINKER_FLAGS="-Wl,-z,max-page-size=16384"

	echo "==> Configure complete. Starting make..."
	make -j"$(getconf _NPROCESSORS_ONLN || echo 2)"
	echo "==> Build complete. Installing..."
	make install
	# sanity: show the produced pkg-config files so we can see them in logs
	echo "==> LibXMP files installed to: $ANDROID_DEPS_PATH"
	ls -l "$ANDROID_DEPS_PATH/lib/libxmp.so" || true
	popd >/dev/null
}

function build_sdl2_android() {
	echo "==> Building SDL2 for Android..."
	local src="$BUILDDIR/sdl2-src"
	[[ -d "$src" ]] || git clone https://github.com/libsdl-org/SDL.git "$src"

	pushd "$src" >/dev/null
	git fetch --tags
	git checkout tags/release-2.32.10
	popd >/dev/null
	
	mkdir -p "$src/build-android"
	pushd "$src/build-android" >/dev/null
	
	cmake .. \
		-DCMAKE_TOOLCHAIN_FILE="$ANDROID_NDK_HOME/build/cmake/android.toolchain.cmake" \
		-DANDROID_ABI="arm64-v8a" \
		-DANDROID_PLATFORM=android-21 \
		-DCMAKE_INSTALL_PREFIX="$ANDROID_DEPS_PATH" \
		-DSDL_ANDROID_PACKAGE_NAME=org.ikemen_engine.ikemen_go \
		-DSDL_STATIC=OFF \
		-DSDL_SHARED=ON \
		-DCMAKE_SHARED_LINKER_FLAGS="-Wl,-z,max-page-size=16384"

	echo "==> Configure complete. Starting make..."
	make -j"$(getconf _NPROCESSORS_ONLN || echo 2)"
	echo "==> Build complete. Installing..."
	make install
	# sanity: show the produced pkg-config files so we can see them in logs
	echo "==> SDL2 files installed to: $ANDROID_DEPS_PATH"
	ls -l "$ANDROID_DEPS_PATH/lib/libSDL2.so" || true
	popd >/dev/null
}

function create_dummy_gl_pc() {
	local pc_dir="$ANDROID_DEPS_PATH/lib/pkgconfig"
	mkdir -p "$pc_dir"
	
	local gl_pc="$pc_dir/gl.pc"
	if [[ ! -f "$gl_pc" ]]; then
		echo "==> Creating dummy gl.pc for Android..."
		cat > "$gl_pc" <<EOF
Name: gl
Description: Android GLES fake
Version: 1.0
Libs:
Cflags:
EOF
	fi
}

download_file() {
	local url="$1" out="$2"
	mkdir -p "$(dirname "$out")"
	if command -v curl >/dev/null 2>&1; then
		curl -fsSL "$url" -o "$out"
	elif command -v wget >/dev/null 2>&1; then
		wget -qO "$out" "$url"
	else
		echo "ERROR: Need curl or wget to download: $url" >&2
		return 1
	fi
}

ensure_screenpack() {
	mkdir -p "$(dirname "$SCREENPACK_DIR")"
	if [[ ! -d "$SCREENPACK_DIR/.git" ]]; then
		rm -rf "$SCREENPACK_DIR"
		echo "==> Cloning Elecbyte screenpack: $SCREENPACK_REPO ($SCREENPACK_REF) -> $SCREENPACK_DIR"
		git clone --depth=1 -b "$SCREENPACK_REF" "$SCREENPACK_REPO" "$SCREENPACK_DIR"
	else
		echo "==> Updating Elecbyte screenpack in $SCREENPACK_DIR ($SCREENPACK_REF)"
		( cd "$SCREENPACK_DIR" \
			&& git fetch --depth=1 origin "$SCREENPACK_REF" \
			&& git checkout -f FETCH_HEAD )
	fi
	# Avoid "dubious ownership" on volume mounts
	git config --system --add safe.directory "$SCREENPACK_DIR" >/dev/null 2>&1 || true
}

ensure_android_runtime_assets() {
	[[ "$GOOS" != "android" ]] && return 0

	# The Android APK must ship with the same runtime assets as desktop releases.
	# If the engine repo doesn't contain them (CI), pull from the Elecbyte screenpack.
	ensure_screenpack

	# external/gamecontrollerdb.txt (referenced by Android manifest)
	local db="$REPO_ROOT/external/gamecontrollerdb.txt"
	if [[ ! -f "$db" ]]; then
		echo "==> Downloading SDL GameController DB..."
		download_file \
			"https://raw.githubusercontent.com/mdqinc/SDL_GameControllerDB/refs/heads/master/gamecontrollerdb.txt" \
			"$db"
	fi

	# data/system.base.def (referenced by Android manifest)
	local sys="$REPO_ROOT/data/system.base.def"
	if [[ ! -f "$sys" ]]; then
		echo "==> Generating data/system.base.def from src/resources/defaultMotif.ini..."
		mkdir -p "$REPO_ROOT/data"
		cp -a "$REPO_ROOT/src/resources/defaultMotif.ini" "$sys"
	fi
}

function patch_go_sdl2_android() {
	[[ "$GOOS" != "android" ]] && return 0

	# go-sdl2 has (in some versions) a build break on Android:
	# it tries to convert SDL_bool (C enum) to Go bool directly.
	# Fix: compare against SDL_FALSE instead.
	local f=""

	# 1) Prefer vendored copy (writable, deterministic)
	local vendorf="$REPO_ROOT/vendor/github.com/veandco/go-sdl2/sdl/system_android.go"
	if [[ -f "$vendorf" ]]; then
		f="$vendorf"
	else
		# 2) Fallback to module cache (may be read-only unless -modcacherw is used)
		local modver modcache moddir
		modver="$(go list -m -f '{{.Version}}' github.com/veandco/go-sdl2 2>/dev/null || true)"
		[[ -z "$modver" ]] && return 0

		modcache="$(go env GOMODCACHE 2>/dev/null || true)"
		[[ -z "$modcache" ]] && return 0

		moddir="$modcache/github.com/veandco/go-sdl2@${modver}"
		f="$moddir/sdl/system_android.go"
	fi

	[[ -f "$f" ]] || return 0

	if grep -q "return bool(C.SDL_AndroidRequestPermission" "$f"; then
		echo "==> Patching go-sdl2 AndroidRequestPermission() SDL_bool->bool conversion..."

		# sed -i needs write permission to the DIRECTORY (temp file) and the file itself.
		local d; d="$(dirname "$f")"
		if [[ ! -w "$d" ]]; then
			# Try to make it writable (works when cache is just chmod'd read-only)
			chmod u+w "$d" 2>/dev/null || true
		fi
		if [[ ! -w "$d" ]]; then
			echo "ERROR: Cannot patch go-sdl2; directory is not writable:" >&2
			echo "  $d" >&2
			echo "If you want to use vendor/, run 'go mod vendor' and re-run the build." >&2
			echo "Otherwise, fix module cache perms (or build via Docker ./build/build_android.sh)." >&2
			exit 1
		fi

		chmod u+w "$f" 2>/dev/null || true
		sed_inplace \
			's/return bool(C.SDL_AndroidRequestPermission(_permission))/return C.SDL_AndroidRequestPermission(_permission) != C.SDL_FALSE/' \
			"$f"
	fi
}

function prepare_android_deps() {
	[[ "$GOOS" != "android" ]] && return 0
	build_sdl2_android
	build_libxmp_android
	build_ffmpeg
	create_dummy_gl_pc
}

function ensure_android_sdk_for_gradle() {
	[[ "$GOOS" != "android" ]] && return 0

	# Prefer ANDROID_SDK_ROOT, fall back to ANDROID_HOME
	local sdk="${ANDROID_SDK_ROOT:-${ANDROID_HOME:-}}"
	if [[ -z "$sdk" ]]; then
		echo "ERROR: ANDROID_SDK_ROOT/ANDROID_HOME not set; required to build APK via Gradle." >&2
		echo "       In Docker this should be provided by the image. If running locally, install Android SDK cmdline-tools and set ANDROID_SDK_ROOT." >&2
		exit 1
	fi
	export ANDROID_SDK_ROOT="$sdk"
	export ANDROID_HOME="$sdk"

	# Make sure Gradle can find sdkmanager/aapt if needed
	if [[ -x "$sdk/cmdline-tools/latest/bin/sdkmanager" ]]; then
		export PATH="$sdk/cmdline-tools/latest/bin:$sdk/platform-tools:${PATH}"
	fi
}

function sync_android_apk_repo() {
	mkdir -p "$(dirname "$ANDROID_APK_DIR")"

	# Local repo support
	if [[ -d "$ANDROID_APK_REPO" ]]; then
		local src
		src="$(cd "$ANDROID_APK_REPO" && pwd -P)"
		if [[ "$src" != "$ANDROID_APK_DIR" ]]; then
			rm -rf "$ANDROID_APK_DIR"
			echo "==> Using local ikemen-droid checkout: $src -> $ANDROID_APK_DIR"
			cp -a "$src" "$ANDROID_APK_DIR"
		else
			echo "==> Using local ikemen-droid checkout in-place: $ANDROID_APK_DIR"
		fi
		# Avoid "dubious ownership" on volume mounts
		git config --system --add safe.directory "$ANDROID_APK_DIR" >/dev/null 2>&1 || true
		return 0
	fi

	if [[ ! -d "$ANDROID_APK_DIR/.git" ]]; then
		rm -rf "$ANDROID_APK_DIR"
		echo "==> Cloning ikemen-droid: $ANDROID_APK_REPO ($ANDROID_APK_REF) -> $ANDROID_APK_DIR"
		git clone --depth=1 -b "$ANDROID_APK_REF" "$ANDROID_APK_REPO" "$ANDROID_APK_DIR"
	else
		echo "==> Updating ikemen-droid in $ANDROID_APK_DIR ($ANDROID_APK_REF)"
		( cd "$ANDROID_APK_DIR" \
			&& git fetch --depth=1 origin "$ANDROID_APK_REF" \
			&& git checkout -f FETCH_HEAD )
	fi
	# Avoid "dubious ownership" on volume mounts
	git config --system --add safe.directory "$ANDROID_APK_DIR" >/dev/null 2>&1 || true
}

function stage_android_apk_libs() {
	local app_dir="$ANDROID_APK_DIR/app"
	# Clean any pre-existing native libs that Gradle might pick up
	find "$app_dir" -type f \( -name "*.so" -o -name "*.so.*" \) \
		\( -path "*/src/main/jniLibs/*" -o -path "*/jniLibs/*" -o -path "*/libs/*" \) \
		-print -delete 2>/dev/null || true
	local abi_dir="$app_dir/src/main/jniLibs/arm64-v8a"

	mkdir -p "$abi_dir"
	# Clean previous staged libs (but keep the dir)
	rm -f "$abi_dir"/*.so* 2>/dev/null || true

	echo "==> Staging native libs into: $abi_dir"
	# Engine
	cp -av "$REPO_ROOT/bin/libmain.so" "$abi_dir/"
	# Deps (only .so, ignore Windows DLLs)
	cp -av "$REPO_ROOT/lib/"*.so* "$abi_dir/" 2>/dev/null || true
}

function stage_android_apk_assets() {
	local app_dir="$ANDROID_APK_DIR/app"
	local assets_dir="$app_dir/src/main/assets"
	local manifest="$assets_dir/manifest.txt"

	if [[ ! -f "$manifest" ]]; then
		echo "ERROR: ikemen-droid manifest not found at: $manifest" >&2
		exit 1
	fi

	echo "==> Staging assets into: $assets_dir (from $manifest)"
	local screenpack="$SCREENPACK_DIR"
	# Remove everything except manifest.txt so the APK content matches your engine tree
	find "$assets_dir" -mindepth 1 -maxdepth 1 ! -name "manifest.txt" -exec rm -rf {} + 2>/dev/null || true

	# manifest.txt is whitespace-separated (often a single long line)
	local p src dst
	while read -r p; do
		[[ -z "$p" ]] && continue
		src="$REPO_ROOT/$p"
		dst="$assets_dir/$p"

		# If not present in the engine repo, fall back to the Elecbyte screenpack clone.
		if [[ ! -e "$src" && -n "$screenpack" && -e "$screenpack/$p" ]]; then
			src="$screenpack/$p"
		fi

		if [[ -d "$src" ]]; then
			mkdir -p "$dst"
			cp -a "$src/." "$dst/" 2>/dev/null || true
		elif [[ -f "$src" ]]; then
			mkdir -p "$(dirname "$dst")"
			cp -a "$src" "$dst"
		else
			# If your tree doesn't include some optional paths from manifest, just warn.
			echo "WARNING: Asset path missing in Ikemen-GO tree: $p" >&2
		fi
	done < <(tr -s '[:space:]' '\n' < "$manifest")
}

function build_android_apk() {
	[[ "$GOOS" != "android" ]] && return 0
	[[ "$BUILD_ANDROID_APK" == "0" ]] && { echo "==> BUILD_ANDROID_APK=0: skipping APK build"; return 0; }

	ensure_android_sdk_for_gradle
	sync_android_apk_repo
	ensure_android_runtime_assets
	stage_android_apk_libs
	stage_android_apk_assets

	echo "==> Checking for stray native libs in apk repo..."
	find "$ANDROID_APK_DIR/app" -type f -name "libmain.so" -print || true

	echo "==> Building APK with Gradle (debug)..."
	pushd "$ANDROID_APK_DIR" >/dev/null

	# Ensure Gradle sees the SDK location
	printf "sdk.dir=%s\n" "${ANDROID_SDK_ROOT}" > local.properties
	chmod +x ./gradlew || true

	./gradlew --no-daemon clean assembleDebug

	local apk_src="$ANDROID_APK_DIR/app/build/outputs/apk/debug/app-debug.apk"
	if [[ ! -f "$apk_src" ]]; then
		echo "ERROR: Gradle finished but APK not found at: $apk_src" >&2
		exit 1
	fi
	mkdir -p "$(dirname "$ANDROID_APK_OUT")"
	cp -av "$apk_src" "$ANDROID_APK_OUT"

	echo "==> APK ready: $ANDROID_APK_OUT"
	popd >/dev/null
}

function maybe_build_ffmpeg() {
	case "$BUILD_FFMPEG" in
		yes)
			echo "==> BUILD_FFMPEG=yes: forcing local FFmpeg build"
			build_ffmpeg
			ensure_pkg_config_path   # prepend our $FFMPEG_PREFIX so pkg-config prefers it
			return 0
		;;
	no)
		ensure_pkg_config_path
		if have_ffmpeg_pc; then
			echo "==> Using system FFmpeg (BUILD_FFMPEG=no)."
			return 0
		fi
		echo "ERROR: FFmpeg dev libraries not found (pkg-config)."
		echo "       Install distro dev packages (e.g. libavformat-dev etc.),"
		echo "       or re-run with BUILD_FFMPEG=yes to build a minimal copy locally."
		exit 1
		;;
	auto|*)
		ensure_pkg_config_path
		if have_ffmpeg_pc; then
			echo "==> Found FFmpeg via pkg-config; using it."
			return 0
		fi
		echo "==> FFmpeg not found via pkg-config; building a minimal local copy."
		build_ffmpeg
		ensure_pkg_config_path
		if ! have_ffmpeg_pc; then
			echo "ERROR: FFmpeg pkg-config files still not visible after build."
			exit 1
		fi
		;;
	esac
}

# Generate delay-load import libraries for MinGW (Windows)
function create_delay_import_libs_windows() {
	[[ "$GOOS" != "windows" ]] && return 0
	mkdir -p "$DELAYLIB_DIR"
	shopt -s nullglob
	local d base name libname
	for d in "$FFMPEG_PREFIX"/bin/*.dll /mingw64/bin/libxmp*.dll /mingw64/bin/libwinpthread-1.dll /mingw64/bin/SDL2*.dll; do
		[[ -f "$d" ]] || continue
		base="$(basename "$d")"          # e.g. libxmp-4.dll
		name="${base%.dll}"              # libxmp-4
		libname="${name%%-*}"            # libxmp
		libname="${libname#lib}"         # xmp
		( cd "$DELAYLIB_DIR" && gendef "$d" >/dev/null )
		dlltool --dllname "$base" --def "$DELAYLIB_DIR/${name}.def" --output-delaylib "$DELAYLIB_DIR/lib${libname}.dll.a"
		rm -f "$DELAYLIB_DIR/${name}.def"
	done
	shopt -u nullglob
}

# --- Build functions ---
function build() {
	if [[ "$GOOS" == "android" ]]; then
		ensure_go_flags_android
		prepare_android_deps
		patch_go_sdl2_android
		# MANUALLY define flags for Android to avoid pkg-config errors
		export CGO_CFLAGS="-I$ANDROID_DEPS_PATH/include -I$ANDROID_DEPS_PATH/include/SDL2 ${CGO_CFLAGS:-}"
		local deps_libs="-L$ANDROID_DEPS_PATH/lib -lSDL2 -lxmp -lavformat -lavcodec -lavutil -lswscale -lswresample -lavfilter"
		# Link against Android system libraries (GLES, OpenSLES, log)
		export CGO_LDFLAGS="${deps_libs} ${CGO_LDFLAGS:-} -lGLESv2 -lOpenSLES -llog -Wl,-z,max-page-size=16384"
	else
		maybe_build_ffmpeg
		export PKG_CONFIG="${PKG_CONFIG:-pkg-config}"
		# Ensure libxmp is present
		require_libxmp
		# Ensure SDL2 is present
		require_sdl2
		# Pull dependency flags from pkg-config (FFmpeg + libxmp + SDL2)
		export CGO_CFLAGS="$($PKG_CONFIG --cflags libavformat libavcodec libavutil libswscale libswresample libavfilter libxmp sdl2) ${CGO_CFLAGS:-}"
		local deps_libs
		deps_libs="$($PKG_CONFIG --libs libavformat libavcodec libavutil libswscale libswresample libavfilter libxmp sdl2)"
		# RPATH for local libs on *nix; macOS adds rpath to bundle/exec path
		if [[ "$GOOS" == "linux" ]]; then
			export CGO_LDFLAGS="${deps_libs} -lpthread -lm -ldl -lz -Wl,-rpath,\$ORIGIN -Wl,-rpath,\$ORIGIN/lib ${CGO_LDFLAGS:-}"
		elif [[ "$GOOS" == "darwin" ]]; then
			export CGO_LDFLAGS="${deps_libs} ${CGO_LDFLAGS:-} -Wl,-rpath,@executable_path -Wl,-rpath,@executable_path/../Frameworks"
		fi		
	fi

	# On macOS we require MoltenVK for the Vulkan renderer
	require_moltenvk

	echo "==> Building Go binary (this may take a while)..."

	# Android has slightly different steps.
	if [[ "$GOOS" == "android" ]]; then
		go build -buildmode=c-shared -trimpath -v -tags=android,gles2 \
		-ldflags="-s -w -X 'main.Version=${APP_VERSION}' -X 'main.BuildTime=${APP_BUILDTIME}' -X 'runtime.godebugDefault=asyncpreemptoff=1,sigaltstack=0'" \
		-o "$OUTDIR/$binName" ./src
	else
		go build -trimpath -v \
		-ldflags "-s -w -X 'main.Version=${APP_VERSION}' -X 'main.BuildTime=${APP_BUILDTIME}'" \
		-o "$OUTDIR/$binName" ./src
	fi

	# bundle libs
	bundle_shared_libs

	# For Android, optionally build the APK via ikemen-droid
	build_android_apk

	echo "==> Build successful"
	echo "    Binary: $OUTDIR/$binName"
	[[ -d "$LIBDIR" ]] && echo "    Runtime libs (if any): $LIBDIR/"
}

function buildWin() {
	stage_windows_resources
	maybe_build_ffmpeg
	export PKG_CONFIG="${PKG_CONFIG:-pkg-config}"
	# Ensure libxmp is present
	require_libxmp
	# Ensure SDL2 is present
	require_sdl2
	# Pull dependency flags from pkg-config (FFmpeg + libxmp + SDL2)
	export CGO_CFLAGS="$($PKG_CONFIG --cflags libavformat libavcodec libavutil libswscale libswresample libavfilter libxmp sdl2) ${CGO_CFLAGS:-}"
	local deps_libs
	deps_libs="$($PKG_CONFIG --libs libavformat libavcodec libavutil libswscale libswresample libavfilter libxmp sdl2)"
	create_delay_import_libs_windows
	# Prefer our delay-libs first so -lavcodec resolves delay-load flavor
	export CGO_LDFLAGS="-L$PWD/$DELAYLIB_DIR ${deps_libs} ${CGO_LDFLAGS:-}"

	echo "==> Building Go binary (this may take a while)..."
	if [[ "${DEBUG_BUILD:-}" -eq 1 ]]; then
		# Console subsystem: keep a terminal for logs/panics while debugging
		go build -trimpath -v \
		  -ldflags "-s -w -X 'main.Version=${APP_VERSION}' -X 'main.BuildTime=${APP_BUILDTIME}'" \
		  -o "$OUTDIR/$binName" ./src
	else
		# GUI subsystem: hides console window
		go build -trimpath -v \
		  -ldflags "-H windowsgui -s -w -X 'main.Version=${APP_VERSION}' -X 'main.BuildTime=${APP_BUILDTIME}'" \
		  -o "$OUTDIR/$binName" ./src
	fi

	# bundle libs
	bundle_shared_libs

	# Clean embedded resource object
	rm -f src/rsrc_windows.syso 2>/dev/null || true

	echo "==> Build successful (Windows)"
	echo "    Binary: $OUTDIR/$binName"
	[[ -d "$LIBDIR" ]] && echo "    Runtime DLLs: $LIBDIR/"
}

# Convert an arbitrary tag (e.g. "v1.2.3", "1.2", "nightly") to a valid
# SxS version "A.B.C.D" (all numeric, 0-65535). Unknowns -> 0.0.0.0
sanitize_sxs_version() {
	local in="$1" s out parts i p
	# strip leading "v" or "V"
	s="${in#v}"; s="${s#V}"
	# keep digits and dots only
	if [[ ! "$s" =~ ^[0-9.]+$ ]]; then
		echo "0.0.0.0"; return
	fi
	IFS='.' read -r -a parts <<<"$s"
	# pad / clamp to 4 parts
	for ((i=${#parts[@]}; i<4; i++)); do parts+=("0"); done
	# coerce empty/non-numeric to 0, clamp to 0..65535
	out=()
	for p in "${parts[@]:0:4}"; do
		[[ "$p" =~ ^[0-9]+$ ]] || p=0
		(( p<0 )) && p=0
		(( p>65535 )) && p=65535
		out+=("$p")
	done
	echo "${out[0]}.${out[1]}.${out[2]}.${out[3]}"
}

# Compile Windows resources and generate a fresh RC + manifest with version/date.
function stage_windows_resources() {
	[[ "$GOOS" != "windows" ]] && return 0

	mkdir -p src build/winres

	# Prepare numeric SxS version & components for VERSIONINFO
	local SXS_VERSION
	SXS_VERSION="$(sanitize_sxs_version "$APP_VERSION")"
	IFS='.' read -r VMAJ VMIN VPAT VREV <<<"$SXS_VERSION"
	# assembly name and arch
	local ASM_NAME="Ikemen_GO"
	local ASM_ARCH
	if [[ "$GOARCH" == "amd64" ]]; then ASM_ARCH="amd64"; else ASM_ARCH="x86"; fi

	# Optionally generate a minimal application manifest (or none)
	cat > build/winres/Ikemen_GO.exe.manifest <<EOF
<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<assembly xmlns="urn:schemas-microsoft-com:asm.v1" manifestVersion="1.0">
  <assemblyIdentity type="win32" name="${ASM_NAME}" version="${SXS_VERSION}" processorArchitecture="${ASM_ARCH}"/>
  <dependency>
	<dependentAssembly>
	  <assemblyIdentity type="win32" name="Microsoft.Windows.Common-Controls"
		version="6.0.0.0" processorArchitecture="*" publicKeyToken="6595b64144ccf1df" language="*"/>
	</dependentAssembly>
  </dependency>
</assembly>
EOF

COPY_START_YEAR="${COPY_START_YEAR:-2016}"
BUILD_YEAR="${APP_BUILDTIME%%-*}"
APP_COPYRIGHT="${APP_COPYRIGHT:-(c) ${COPY_START_YEAR}-${BUILD_YEAR} Ikemen GO team (MIT)}"

	# Generate a RC
	cat > build/winres/Ikemen_GO.rc <<EOF
#include <windows.h>
#include <winver.h>
1 ICON "Ikemen_Cylia_V2.ico"
1 RT_MANIFEST "Ikemen_GO.exe.manifest"

VS_VERSION_INFO VERSIONINFO
 FILEVERSION ${VMAJ},${VMIN},${VPAT},${VREV}
 PRODUCTVERSION ${VMAJ},${VMIN},${VPAT},${VREV}
 FILEFLAGSMASK 0x3fL
 FILEFLAGS 0x0L
 FILEOS 0x4L
 FILETYPE 0x1L
 FILESUBTYPE 0x0L
BEGIN
	BLOCK "StringFileInfo"
	BEGIN
		BLOCK "040904B0"
		BEGIN
			VALUE "CompanyName", "Ikemen GO\\0"
			VALUE "FileDescription", "Ikemen GO\\0"
			VALUE "FileVersion", "${SXS_VERSION}\\0"
			VALUE "ProductName", "Ikemen GO\\0"
			VALUE "ProductVersion", "${SXS_VERSION}\\0"
			VALUE "OriginalFilename", "Ikemen_GO.exe\\0"
			VALUE "InternalName", "Ikemen_GO\\0"
			VALUE "BuildDate", "${APP_BUILDTIME}\\0"
			VALUE "LegalCopyright", "${APP_COPYRIGHT}\\0"
		END
	END
	BLOCK "VarFileInfo"
	BEGIN
		VALUE "Translation", 0x0409, 1200
	END
END
EOF

	# Compile RC -> COFF object that Go will auto-link (.syso in package dir)

	local wr targetflag
	if [[ "$GOARCH" == "amd64" ]]; then
		wr="${WINDRES:-x86_64-w64-mingw32-windres}"
		targetflag="--target=pe-x86-64"
	else
		wr="${WINDRES:-i686-w64-mingw32-windres}"
		targetflag="--target=pe-i386"
	fi
	command -v "$wr" >/dev/null 2>&1 || wr="windres"
	echo "==> Embedding Windows resources (icon + manifest) with $wr..."
	"$wr" $targetflag \
	  -I build/winres -I external/icons \
	  -i build/winres/Ikemen_GO.rc \
	  -O coff -o src/rsrc_windows.syso
}

# Copy FFmpeg shared libs next to produced binary for easy runtime
function bundle_shared_libs() {
	local dest_lib="$LIBDIR"
	mkdir -p "$dest_lib"
	if [[ -d "$FFMPEG_PREFIX/bin" ]]; then
		# Windows
		cp -av "$FFMPEG_PREFIX"/bin/*.dll "$dest_lib/" 2>/dev/null || true
		# MSYS2 runtime dep
		#cp -av /mingw64/bin/libwinpthread-1.dll "$dest_lib/" 2>/dev/null || true
		for d in \
			/mingw64/bin/libwinpthread-1.dll \
			/mingw64/bin/libgcc_s_seh-1.dll \
			/mingw64/bin/libstdc++-6.dll \
			/mingw64/bin/libxmp*.dll \
			/mingw64/bin/SDL2*.dll ; do
			cp -av "$d" "$dest_lib/" 2>/dev/null || true
		done
	elif [[ -d "$FFMPEG_PREFIX/lib" && "$GOOS" != "android" ]]; then
		# Linux & macOS
		cp -av "$FFMPEG_PREFIX"/lib/lib*.so* "$dest_lib/" 2>/dev/null || true
		cp -av "$FFMPEG_PREFIX"/lib/lib*.dylib "$dest_lib/" 2>/dev/null || true
	fi
	# Bundle MoltenVK on macOS
	if [[ "$GOOS" == "darwin" ]]; then
		local mvk="${MVK_DYLIB:-}"
		if [[ -z "$mvk" ]]; then
			for d in \
				"/opt/homebrew/lib/libMoltenVK.dylib" \
				"/usr/local/lib/libMoltenVK.dylib"; do
				[[ -f "$d" ]] && { mvk="$d"; break; }
			done
		fi
		[[ -n "$mvk" ]] && cp -av "$mvk" "$dest_lib/" 2>/dev/null || true
	fi
	# Android specific bundling
	if [[ "$GOOS" == "android" ]]; then
		echo "==> Bundling Android dependencies from $ANDROID_DEPS_PATH..."
		cp -av "$ANDROID_DEPS_PATH"/lib/*.so* "$dest_lib/" 2>/dev/null || true
	fi
	# Always try to bundle libxmp for portable runtime on Linux/macOS.
	# (Windows and Android were handled above.)
	if [[ "$GOOS" != "windows" && "$GOOS" != "android" ]]; then
		# Prefer pkg-config to locate the correct lib directory.
		local pc libdir libdir_sdl2
		pc="${PKG_CONFIG:-pkg-config}"
		libdir="$($pc --variable=libdir libxmp 2>/dev/null || true)"
		if [[ -n "$libdir" && -d "$libdir" ]]; then
			cp -av "${libdir}"/libxmp*.so*   "$dest_lib/" 2>/dev/null || true
			cp -av "${libdir}"/libxmp*.dylib "$dest_lib/" 2>/dev/null || true
		fi
		libdir_sdl2="$($pc --variable=libdir sdl2 2>/dev/null || true)"
		if [[ -n "$libdir_sdl2" && -d "$libdir_sdl2" ]]; then
			cp -av "${libdir_sdl2}"/libSDL2*.so*   "$dest_lib/" 2>/dev/null || true
			cp -av "${libdir_sdl2}"/libSDL2*.dylib "$dest_lib/" 2>/dev/null || true
		fi
	fi
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
			if [[ -z "${1:-}" ]]; then
				echo "Unknown system \"${OSTYPE}\".";
				exit 1
			fi
		;;
	esac
}

if [ ! -f ./go.mod ]; then
	echo "Missing go.mod. Run:  go mod init <module> && go mod download"
	exit 1
fi

main "$@"
