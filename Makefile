# Set Bash as the shell.
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
	src/system.go \
	src/util_desktop.go \
	src/util_js.go

# Windows 64-bit target
Ikemen_GO.exe: ${srcFiles}
	cd ./build && bash ./build.sh Win64

# Windows 32-bit target
Ikemen_GO_86.exe: ${srcFiles}
	cd ./build && bash ./build.sh Win32

# Linux target
Ikemen_GO_Linux: ${srcFiles}
	cd ./build && ./build.sh Linux

# MacOS x64 target
Ikemen_GO_MacOS: ${srcFiles}
	cd ./build && bash ./build.sh MacOS

# MacOS app bundle
appbundle:
	mkdir -p I.K.E.M.E.N-Go.app
	mkdir -p I.K.E.M.E.N-Go.app/Contents
	mkdir -p I.K.E.M.E.N-Go.app/Contents/MacOS
	mkdir -p I.K.E.M.E.N-Go.app/Contents/Resources
	cp bin/Ikemen_GO_MacOS I.K.E.M.E.N-Go.app/Contents/MacOS/Ikemen_GO_MacOS
	cp ./build/Info.plist I.K.E.M.E.N-Go.app/Contents/Info.plist
	cp ./build/bundle_run.sh I.K.E.M.E.N-Go.app/Contents/MacOS/bundle_run.sh
	chmod +x I.K.E.M.E.N-Go.app/Contents/MacOS/bundle_run.sh
	chmod +x I.K.E.M.E.N-Go.app/Contents/MacOS/Ikemen_GO_MacOS
	cd ./build && mkdir -p ./icontmp/icon.iconset && \
	cp ../external/icons/IkemenCylia_256.png ./icontmp/icon.iconset/icon_256x256.png && \
	iconutil -c icns ./icontmp/icon.iconset && \
	cp icontmp/icon.icns ../I.K.E.M.E.N-Go.app/Contents/Resources/icon.icns && \
	rm -rf icontmp

clean_appbundle:
	rm -rf I.K.E.M.E.N-Go.app