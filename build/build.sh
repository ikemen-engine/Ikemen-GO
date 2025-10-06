#!/bin/bash

# Exit in case of failure; print function-trap friendly errors
set -o errtrace
set -euo pipefail

## Resolve repo root _first_ so all path defaults are stable no matter where
## the script is invoked from (top dir, build dir, or elsewhere).
REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd -P)"
cd "$REPO_ROOT"

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

# FFmpeg config
FFMPEG_REV="${FFMPEG_REV:-release/7.1}"
# Always default to build/ffmpeg under the repo root (absolute path).
FFMPEG_PREFIX="${FFMPEG_PREFIX:-$REPO_ROOT/$BUILDDIR/ffmpeg}"
BUILD_FFMPEG="${BUILD_FFMPEG:-auto}"   # auto|yes|no

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
      need yasm
      need go
      need gendef
      need dlltool
      if ((${#missing[@]})); then
        echo "Install from the MINGW64 shell:" >&2
        echo "  pacman -Syu --noconfirm" >&2
        echo "  pacman -S --noconfirm \\" >&2
        echo "    git make diffutils mingw-w64-x86_64-pkg-config \\" >&2
        echo "    mingw-w64-x86_64-go mingw-w64-x86_64-toolchain \\" >&2
        echo "    mingw-w64-x86_64-nasm mingw-w64-x86_64-yasm \\" >&2
        echo "    mingw-w64-x86_64-tools-git" >&2
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
        echo "  brew update && brew install git go pkg-config nasm" >&2
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
        echo "  sudo apt update && sudo apt install -y golang-go git pkg-config make nasm yasm build-essential" >&2
        exit 1
      fi
      ;;
    *)
      :
      ;;
  esac
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
	case "$OSTYPE" in
		darwin*) OUTDIR="bin" ;;
		*)       OUTDIR="."  ;;
	esac
	mkdir -p "$OUTDIR"
	mkdir -p "$LIBDIR"

	# Make sure Go toolchain is usable
	ensure_go_env

	# Make sure required build tools exist
	check_deps

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
	git clone https://github.com/FFmpeg/FFmpeg.git "$FFMPEG_SRCDIR"
	pushd "$FFMPEG_SRCDIR" >/dev/null
	git checkout "$FFMPEG_REV"

	echo "==> Starting configure..."
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
	echo "==> FFmpeg pkg-config files installed to: $FFMPEG_PREFIX/lib/pkgconfig"
	ls -l "$FFMPEG_PREFIX/lib/pkgconfig" || true
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
	local d
	for d in "$FFMPEG_PREFIX"/bin/*.dll; do
		[[ -f "$d" ]] || continue
		local base name libname
		base="$(basename "$d")"          # e.g. avcodec-59.dll
		name="${base%.dll}"              # avcodec-59
		libname="${name%%-*}"            # avcodec
		libname="${libname#lib}"         # strip lib- prefix if any
		( cd "$DELAYLIB_DIR" && gendef "$d" >/dev/null )
		dlltool --dllname "$base" --def "$DELAYLIB_DIR/${name}.def" --output-delaylib "$DELAYLIB_DIR/lib${libname}.dll.a"
		rm -f "$DELAYLIB_DIR/${name}.def"
	done
	# Also delay-load libwinpthread (dependency of avutil on MSYS2)
	if [[ -f /mingw64/bin/libwinpthread-1.dll ]]; then
		( cd "$DELAYLIB_DIR" && gendef /mingw64/bin/libwinpthread-1.dll >/dev/null )
		dlltool --dllname libwinpthread-1.dll --def "$DELAYLIB_DIR/libwinpthread-1.def" --output-delaylib "$DELAYLIB_DIR/libwinpthread.dll.a"
		rm -f "$DELAYLIB_DIR/libwinpthread-1.def"
	fi
	shopt -u nullglob
}

# ---- App metadata (overridden by CI)
APP_VERSION="${APP_VERSION:-nightly}"
APP_BUILDTIME="${APP_BUILDTIME:-$(date +%F)}"

# --- Build functions ---
function build() {
	maybe_build_ffmpeg
	export PKG_CONFIG="${PKG_CONFIG:-pkg-config}"
	# Pull FFmpeg flags from pkg-config
	export CGO_CFLAGS="$($PKG_CONFIG --cflags libavformat libavcodec libavutil libswscale libswresample libavfilter) ${CGO_CFLAGS:-}"
	local ffmpeg_libs
	ffmpeg_libs="$($PKG_CONFIG --libs libavformat libavcodec libavutil libswscale libswresample libavfilter)"

	# RPATH for local libs on *nix; macOS adds rpath to bundle/exec path
	if [[ "$GOOS" == "linux" ]]; then
		export CGO_LDFLAGS="${ffmpeg_libs} -lpthread -lm -ldl -lz -Wl,-rpath,\$ORIGIN -Wl,-rpath,\$ORIGIN/lib ${CGO_LDFLAGS:-}"
	elif [[ "$GOOS" == "darwin" ]]; then
		export CGO_LDFLAGS="${ffmpeg_libs} ${CGO_LDFLAGS:-} -Wl,-rpath,@executable_path -Wl,-rpath,@executable_path/../Frameworks"
	fi

	echo "==> Building Go binary (this may take a while)..."
	go build -trimpath -v \
	  -ldflags "-X 'main.Version=${APP_VERSION}' -X 'main.BuildTime=${APP_BUILDTIME}'" \
	  -o "$OUTDIR/$binName" ./src

	# bundle libs
	bundle_shared_libs

	echo "==> Build successful"
	echo "    Binary: $OUTDIR/$binName"
	[[ -d "$LIBDIR" ]] && echo "    Runtime libs (if any): $LIBDIR/"
}

function buildWin() {
	stage_windows_resources
	maybe_build_ffmpeg
	export PKG_CONFIG="${PKG_CONFIG:-pkg-config}"
	export CGO_CFLAGS="$($PKG_CONFIG --cflags libavformat libavcodec libavutil libswscale libswresample libavfilter) ${CGO_CFLAGS:-}"
	local ffmpeg_libs
	ffmpeg_libs="$($PKG_CONFIG --libs libavformat libavcodec libavutil libswscale libswresample libavfilter)"
	create_delay_import_libs_windows
	# Prefer our delay-libs first so -lavcodec resolves delay-load flavor
	export CGO_LDFLAGS="-L$PWD/$DELAYLIB_DIR ${ffmpeg_libs} ${CGO_LDFLAGS:-}"

	echo "==> Building Go binary (this may take a while)..."
	if [[ "${DEBUG_BUILD:-}" -eq 1 ]]; then
		# Console subsystem: keep a terminal for logs/panics while debugging
		go build -trimpath -v \
		  -ldflags "-X 'main.Version=${APP_VERSION}' -X 'main.BuildTime=${APP_BUILDTIME}'" \
		  -o "$OUTDIR/$binName" ./src
	else
		# GUI subsystem: hides console window
		go build -trimpath -v \
		  -ldflags "-H windowsgui -X 'main.Version=${APP_VERSION}' -X 'main.BuildTime=${APP_BUILDTIME}'" \
		  -o "$OUTDIR/$binName" ./src
	fi

	# bundle libs
	bundle_shared_libs

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
			/mingw64/bin/libstdc++-6.dll ; do
			cp -av "$d" "$dest_lib/" 2>/dev/null || true
		done
	elif [[ -d "$FFMPEG_PREFIX/lib" ]]; then
		# Linux & macOS
		cp -av "$FFMPEG_PREFIX"/lib/lib*.so* "$dest_lib/" 2>/dev/null || true
		cp -av "$FFMPEG_PREFIX"/lib/lib*.dylib "$dest_lib/" 2>/dev/null || true
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
