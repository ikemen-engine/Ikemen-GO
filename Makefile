.POSIX:
.SUFFIXES:
.PHONY: all clean cross linux macos pkg windows

CROSS=Ikemen_GO_Linux Ikemen_GO_MacOS Ikemen_GO_x64.exe Ikemen_GO_x86.exe
ZIP=${CROSS} data external font sound License.txt SoftOpenAL32.dll SoftOpenAL64.dll
SCREENPACK=elecbyte/chars elecbyte/data elecbyte/font elecbyte/stages

GOFILES=src/anim.go\
	src/bgdef.go\
	src/bytecode.go\
	src/camera.go\
	src/char.go\
	src/common.go\
	src/compiler.go\
	src/font.go\
	src/image.go\
	src/input.go\
	src/lifebar.go\
	src/main.go\
	src/render.go\
	src/script.go\
	src/sound.go\
	src/stage.go\
	src/stdout_windows.go\
	src/system.go

all:
	@echo targets: clean cross linux macos pkg windows

Ikemen_GO: ${GOFILES}
	go build ${TAGS} -o $@ ./src

Ikemen_GO.exe: ${GOFILES}
	go build -ldflags "-H windowsgui" -o $@ ./src

linux: Ikemen_GO

macos: Ikemen_GO

windows: Ikemen_GO.exe


# cross-compiling from Linux
Ikemen_GO_Linux: ${GOFILES}
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
	go build ${TAGS} -o $@ ./src

Ikemen_GO_MacOS: ${GOFILES}
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 \
	CC=o64-clang \
	CXX=o64-clang++ \
	go build -tags al_cmpt -o $@ ./src

Ikemen_GO_x64.exe: ${GOFILES} windres/Ikemen_Cylia_x64.syso
	cp 'windres/Ikemen_Cylia_x64.syso' 'src/Ikemen_Cylia_x64.syso'
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 \
	CC=x86_64-w64-mingw32-gcc \
	CXX=x86_64-w64-mingw32-g++ \
	go build -ldflags '-H windowsgui' -o $@ ./src
	rm 'src/Ikemen_Cylia_x64.syso'

Ikemen_GO_x86.exe: ${GOFILES} windres/Ikemen_Cylia_x86.syso
	cp 'windres/Ikemen_Cylia_x86.syso' 'src/Ikemen_Cylia_x86.syso'
	CGO_ENABLED=1 GOOS=windows GOARCH=386 \
	CC=i686-w64-mingw32-gcc \
	CXX=i686-w64-mingw32-g++ \
	go build -ldflags '-H windowsgui' -o $@ ./src
	rm 'src/Ikemen_Cylia_x86.syso'

cross: ${CROSS}


# zip packing
save save/replays sound:
	mkdir -p $@

elecbyte:
	git clone https://github.com/ikemen-engine/Ikemen_GO-Elecbyte-Screenpack.git elecbyte

SoftOpenAL32.dll:
	curl -sSLfO https://github.com/ikemen-engine/go-openal/raw/master/openal/lib/SoftOpenAL32.dll

SoftOpenAL64.dll:
	curl -sSLfO https://github.com/ikemen-engine/go-openal/raw/master/openal/lib/SoftOpenAL64.dll

Ikemen_GO_CoreOnly.zip: ${CROSS} ${ZIP} build/Ikemen_GO.command
	mkdir -p Ikemen_GO_CoreOnly
	cp -r ${ZIP} Ikemen_GO_CoreOnly
	cp build/Ikemen_GO.command Ikemen_GO_CoreOnly/Ikemen_GO.command
	rm -f $@
	7z a -tzip $@ Ikemen_GO_CoreOnly
	rm -r Ikemen_GO_CoreOnly

Ikemen_GO.zip: ${CROSS} ${ZIP} elecbyte build/Ikemen_GO.command
	mkdir -p Ikemen_GO
	cp -r ${ZIP} ${SCREENPACK} Ikemen_GO
	cp build/Ikemen_GO.command Ikemen_GO/Ikemen_GO.command
	cp elecbyte/LICENCE.txt Ikemen_GO/ScreenpackLicense.txt
	rm -f $@
	7z a -tzip $@ Ikemen_GO
	rm -r Ikemen_GO

pkg: Ikemen_GO_CoreOnly.zip Ikemen_GO.zip

clean:
	rm -rf Ikemen_GO* elecbyte SoftOpenAL32.dll SoftOpenAL64.dll save
