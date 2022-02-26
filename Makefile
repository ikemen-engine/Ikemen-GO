# Set Bash as the shell to get $OSTYPE
SHELL=/bin/bash

# /src files
srcFiles=src/anim.go \
	src/bgdef.go \
	src/bytecode.go \
	src/camera.go \
	src/char.go \
	src/common.go \
	src/compiler.go \
	src/compiler_functions.go \
	src/font.go \
	src/image.go \
	src/input.go \
	src/lifebar.go \
	src/main.go \
	src/render.go \
	src/script.go \
	src/sound.go \
	src/stage.go \
	src/stdout_windows.go \
	src/system.go

# Windows x64 target
Ikemen_GO.exe: ${srcFiles}
	cd ./build && bash ./build.sh Win64
	cd ./build && curl -sSLfO https://github.com/ikemen-engine/go-openal/raw/master/openal/lib/SoftOpenAL64.dll

Ikemen_GO_86.exe: ${srcFiles}
	cd ./build && bash ./build.sh Win32
	cd ./build && curl -sSLfO https://github.com/ikemen-engine/go-openal/raw/master/openal/lib/SoftOpenAL32.dll
	
Ikemen_GO_Linux: ${srcFiles}
	cd ./build && /build.sh Linux ${tags}
	
Ikemen_GO_MacOS: ${srcFiles}
	cd ./build && bash ./build.sh MacOS