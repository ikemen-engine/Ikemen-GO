# Building Ikemen GO

Ikemen GO links against **FFmpeg** for background video (VP9/Opus/Vorbis in WebM/Matroska).
`build/build.sh` **auto-detects your OS** and, by default, **auto-builds a minimal FFmpeg**
(same config as CI). You don't need system FFmpeg dev packages unless you prefer them.

---

## Windows (MSYS2 / MINGW64)

### Dependencies
Install MSYS2 from https://www.msys2.org and open **MSYS2 MINGW64**, then:
```bash
pacman -Syu --noconfirm
pacman -S --noconfirm \
  git make diffutils mingw-w64-x86_64-pkg-config \
  mingw-w64-x86_64-go mingw-w64-x86_64-toolchain \
  mingw-w64-x86_64-nasm mingw-w64-x86_64-yasm \
  mingw-w64-x86_64-tools-git
```
> On MSYS2 we auto-fix "trimmed" Go by setting `GOROOT=/mingw64/lib/go` if needed.

### Build 64-bit (Ikemen_GO.exe)
```bash
git clone https://github.com/ikemen-engine/Ikemen-GO.git
cd Ikemen-GO
# build.sh (matches CI default)
./build/build.sh Win64
# or make
make Ikemen_GO.exe
```

### Build 32-bit (Ikemen_GO_x86.exe)
> Requires 32-bit MinGW cross tools in addition to the above:
> `pacman -S --noconfirm mingw-w64-i686-toolchain mingw-w64-i686-pkg-config mingw-w64-i686-nasm mingw-w64-i686-yasm`
```bash
# build.sh
./build/build.sh Win32
# or make
make Ikemen_GO_x86.exe
```

### Run (Windows)
```bash
./Ikemen_GO.exe          # 64-bit
./Ikemen_GO_x86.exe      # 32-bit
```

### Use system FFmpeg instead (optional)
Install `mingw-w64-x86_64-ffmpeg` (and/or i686 variant for 32-bit), then:
```bash
BUILD_FFMPEG=no ./build/build.sh Win64   # or Win32
```

---

## Linux

### Dependencies (Debian/Ubuntu)
```bash
sudo apt update && sudo apt install -y \
  golang-go git pkg-config make nasm yasm build-essential libxmp-dev
```

### Build x86-64 (Ikemen_GO_Linux)
```bash
git clone https://github.com/ikemen-engine/Ikemen-GO.git
cd Ikemen-GO
# build.sh (matches CI default)
./build/build.sh Linux
# or make
make Ikemen_GO_Linux
```

### Build ARM64 on an ARM host (Ikemen_GO_LinuxARM)
> On an **ARM64 (aarch64) machine**, the same dependencies apply.
```bash
# build.sh
./build/build.sh LinuxARM
# or make
make Ikemen_GO_LinuxARM
```
> Cross-compiling x86→ARM with CGO/FFmpeg requires an ARM cross toolchain and is not covered here.

### Run (Linux)
```bash
./Ikemen_GO_Linux        # x86-64
./Ikemen_GO_LinuxARM     # ARM64
# If you need a GL fallback on some drivers:
MESA_GL_VERSION_OVERRIDE=2.1 ./Ikemen_GO_Linux
```
You can also double-click **`build/Ikemen_GO.command`** on Linux.

### Use system FFmpeg instead (optional)
```bash
sudo apt install -y ffmpeg libavcodec-dev libavformat-dev libavutil-dev libswscale-dev libswresample-dev libavfilter-dev
BUILD_FFMPEG=no ./build/build.sh Linux      # or LinuxARM
```

---

## macOS (Apple Silicon by default; Intel supported)

### Dependencies (Homebrew)
```bash
brew update && brew install git go pkg-config nasm
# Optional: brew install yasm
```

### Build (Apple Silicon default)
```bash
git clone https://github.com/ikemen-engine/Ikemen-GO.git
cd Ikemen-GO
# build.sh (matches CI default)
./build/build.sh MacOSARM
# or make
make Ikemen_GO_MacOSARM
```

### Build (Intel)
```bash
./build/build.sh MacOS
# or
make Ikemen_GO_MacOS
```

### App bundle (optional)
```bash
make appbundle BINNAME=bin/Ikemen_GO_MacOSARM   # or BINNAME=bin/Ikemen_GO_MacOS
open I.K.E.M.E.N-Go.app
```
You can also double-click **`build/Ikemen_GO.command`**; it starts the bundle or the binary.

### Run (raw binary)
```bash
./bin/Ikemen_GO_MacOSARM   # Apple Silicon
./bin/Ikemen_GO_MacOS      # Intel
```

### Use system FFmpeg instead (optional)
```bash
brew install ffmpeg
BUILD_FFMPEG=no ./build/build.sh MacOSARM   # or MacOS
```

---

## Assets required to run
Place these folders **next to the executable or app bundle**:
`data`, `external`, `font`, and a screenpack (see our Elecbyte screenpack repo).
The release CI bundles these automatically.

---

## Notes & licensing
- The minimal FFmpeg we build matches CI: shared libs only; `file` protocol; Matroska/WebM demuxers;
  VP9/Opus/Vorbis decoders and parsers; no FFmpeg CLI tools.
- FFmpeg is used under **LGPL v2.1**; releases attach the corresponding source snapshot.
- Ikemen GO sources are MIT; bundled screenpack assets have their own licenses.

---

## Troubleshooting
- **Missing tools**: re-run the dependency commands for your OS/arch.
- **FFmpeg link errors**: use the default `build.sh` (auto-builds FFmpeg), or install system FFmpeg
  dev packages and run with `BUILD_FFMPEG=no`.
- **Windows DLLs**: verify `.\lib\*.dll` exists (local build places FFmpeg DLLs there).
- **Linux GL compatibility**: try `MESA_GL_VERSION_OVERRIDE=2.1` for a fallback.
