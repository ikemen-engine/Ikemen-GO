# Set Bash as the shell.
SHELL=/bin/bash

# NOTE: Only used for make's change detection; Go still builds ./src.
# /src files
srcFiles=src/resources/defaultConfig.ini \
	src/anim.go \
	src/bgdef.go \
	src/bytecode.go \
	src/camera.go \
	src/char.go \
	src/common.go \
	src/compiler.go \
	src/compiler_functions.go \
	src/config.go \
	src/dllsearch_windows.go \
	src/font.go \
	src/image.go \
	src/iniutils.go \
	src/input.go \
	src/input_glfw.go \
	src/lifebar.go \
	src/main.go \
	src/net.go \
	src/render.go \
	src/render_gl.go \
	src/render_gl_gl32.go \
	src/rollback.go \
	src/script.go \
	src/sound.go \
	src/stage.go \
	src/state.go \
	src/state_clone.go \
	src/stdout_windows.go \
	src/system.go \
	src/system_glfw.go \
	src/util_desktop.go \
	src/util_js.go \
	src/util_raw.go \
	src/video_ffmpeg.go
	

# Windows 64-bit target
Ikemen_GO.exe: ${srcFiles}
	bash ./build.sh Win64

# Windows 32-bit target
Ikemen_GO_x86.exe: ${srcFiles}
	bash ./build.sh Win32

# Linux target
Ikemen_GO_Linux: ${srcFiles}
	./build.sh Linux

# Linux ARM target
Ikemen_GO_LinuxARM: ${srcFiles}
	./build.sh LinuxARM

# MacOS x64 target
Ikemen_GO_MacOS: ${srcFiles}
	bash ./build.sh MacOS
	$(MAKE) clean_appbundle
	$(MAKE) appbundle BINNAME=bin/Ikemen_GO_MacOS

# MacOS Apple Silicon target
Ikemen_GO_MacOSARM: ${srcFiles}
	bash ./build.sh MacOSARM
	$(MAKE) clean_appbundle
	$(MAKE) appbundle BINNAME=bin/Ikemen_GO_MacOSARM

# MacOS app bundle
appbundle:
	mkdir -p I.K.E.M.E.N-Go.app
	mkdir -p I.K.E.M.E.N-Go.app/Contents
	mkdir -p I.K.E.M.E.N-Go.app/Contents/MacOS
	mkdir -p I.K.E.M.E.N-Go.app/Contents/Resources
	# BINNAME can be a full path (e.g. bin/Ikemen_GO_MacOS) or just the filename.
	cp $(BINNAME) I.K.E.M.E.N-Go.app/Contents/MacOS/$(notdir $(BINNAME))
	cp ./build/Info.plist I.K.E.M.E.N-Go.app/Contents/Info.plist
	cp ./build/bundle_run.sh I.K.E.M.E.N-Go.app/Contents/MacOS/bundle_run.sh
	chmod +x I.K.E.M.E.N-Go.app/Contents/MacOS/bundle_run.sh
	chmod +x I.K.E.M.E.N-Go.app/Contents/MacOS/$(notdir $(BINNAME))
	mkdir -p build/icontmp/icon.iconset
	cp external/icons/IkemenCylia_256.png build/icontmp/icon.iconset/icon_256x256.png
	iconutil -c icns build/icontmp/icon.iconset -o build/icontmp/icon.icns
	cp build/icontmp/icon.icns I.K.E.M.E.N-Go.app/Contents/Resources/icon.icns
	rm -rf build/icontmp

clean_appbundle:
	rm -rf I.K.E.M.E.N-Go.app
