# Set Bash as the shell.
SHELL=/bin/bash

# NOTE: Only used for make's change detection; Go still builds ./src.
# /src files
srcFiles=src/resources/defaultConfig.ini \
	src/anim.go \
	src/audio_sdl.go \
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
	src/font_gl33.go \
	src/font_gles32.go \
	src/font_vk.go \
	src/hiscore_rank.go \
	src/image.go \
	src/iniutils.go \
	src/input.go \
	src/input_sdl.go \
	src/lifebar.go \
	src/main.go \
	src/motif.go \
	src/music.go \
	src/net.go \
	src/rect.go \
	src/render.go \
	src/render_gl33.go \
	src/render_gles32.go \
	src/render_vk.go \
	src/rollback.go \
	src/script.go \
	src/select_params.go \
	src/sound.go \
	src/sound_xm.go \
	src/stage.go \
	src/state.go \
	src/state_clone.go \
	src/stats.go \
	src/stdout_windows.go \
	src/storyboard.go \
	src/system.go \
	src/system_sdl.go \
	src/util_android.go \
	src/util_darwin.go \
	src/util_desktop.go \
	src/util_linux.go \
	src/util_raw.go \
	src/util_windows.go \
	src/video_ffmpeg.go
	

# Windows 64-bit target
Ikemen_GO.exe: ${srcFiles}
	bash ./build/build.sh Win64

# Windows 32-bit target
Ikemen_GO_x86.exe: ${srcFiles}
	bash ./build/build.sh Win32

# Linux target
Ikemen_GO_Linux: ${srcFiles}
	./build/build.sh Linux

# Linux ARM target
Ikemen_GO_LinuxARM: ${srcFiles}
	./build/build.sh LinuxARM

# MacOS x64 target
Ikemen_GO_MacOS: ${srcFiles}
	bash ./build/build.sh MacOS
	$(MAKE) clean_appbundle
	$(MAKE) appbundle BINNAME=bin/Ikemen_GO_MacOS

# MacOS Apple Silicon target
Ikemen_GO_MacOSARM: ${srcFiles}
	bash ./build/build.sh MacOSARM
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

.PHONY: android-apk
android-apk:
	bash ./build/build_android.sh

clean_appbundle:
	rm -rf I.K.E.M.E.N-Go.app
