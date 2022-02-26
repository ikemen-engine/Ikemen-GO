@echo off
cd ..

if not exist bin (
	MKDIR bin
)
cd bin

@echo on
curl -SLfO https://github.com/ikemen-engine/go-openal/raw/master/openal/lib/SoftOpenAL64.dll
curl -SLfO https://github.com/ikemen-engine/go-openal/raw/master/openal/lib/SoftOpenAL32.dll