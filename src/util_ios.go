//go:build ios

package main

/*
#include <stdlib.h>
#include <unistd.h>
#include "SDL.h"
#include "SDLWindowFix.h"

// THIS IS THE KEY: A C constructor.
// This runs when the .framework is loaded, BEFORE Go starts itself up.
__attribute__((constructor))
static void prepare_go_runtime(void) {
    // cgocheck=0: Stop Go from scanning memory pointers (prevents possible crashes)
    // scavenge=off: Stop the background memory reclaimer thread
    setenv("GODEBUG", "asyncpreemptoff=1,cgocheck=0", 1);
}
*/
import "C"
import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"unsafe"

	findfont "github.com/flopp/go-findfont"
)

func init() {
}

// Log writer implementation
// On iOS, writing to Stderr redirects to the console output we want
func NewLogWriter() io.Writer {
	return os.Stderr
}

// TTF font loading
func LoadFntTtf(f *Fnt, fontfile string, filename string, height int32) {
	// 1. Path resolution
	// Attempt to find the file using the engine's SearchFile utility
	fileDir := SearchFile(filename, []string{fontfile, sys.motif.Def, "", "data/"}, "font/")

	// 2. iOS Path Correction
	// If the path isn't absolute and doesn't exist, anchor it to the SDL storage path
	if !filepath.IsAbs(fileDir) && FileExist(fileDir) == "" {
		fullPath := filepath.Join(sys.baseDir, fileDir)

		if FileExist(fullPath) != "" {
			fileDir = fullPath
		}
	}

	// Fallback to findfont if still not found
	if FileExist(fileDir) == "" {
		if found, err := findfont.Find(fileDir); err == nil {
			fileDir = found
		} else {
			Logcat(fmt.Sprintf("Font search failed for %s, trying direct path...", filename))
		}
	}

	// 2. Set dimensions
	if height == -1 {
		height = int32(f.Size[1])
	} else {
		f.Size[1] = uint16(height)
	}

	// 3. Load the TTF
	ttf, err := gfxFont.LoadFont(fileDir, height, int(sys.gameWidth), int(sys.gameHeight))
	if err != nil {
		// Instead of a pure panic, log exactly where it looked
		Logcat(fmt.Sprintf("ERROR: Failed to load TTF from %s", fileDir))
		panic(fmt.Errorf("failed to load ttf font %v: %w", fileDir, err))
	}

	f.ttf = ttf.(Font)

	// 4. Create dummy palettes
	f.palettes = make([][256]uint32, 1)
	for i := 0; i < 256; i++ {
		f.palettes[0][i] = 0
	}
}

// Message box implementation
// iOS doesn't have a simple "dialog" package like desktop.
func ShowInfoDialog(message, title string) {
	Logcat(fmt.Sprintf("INFO [%s]: %s", title, message))
}

func ShowErrorDialog(message string) {
	Logcat(fmt.Sprintf("CRITICAL ERROR: %s", message))
}

//export SDL_main
func SDL_main(argc C.int, argv **C.char) C.int {
	runtime.LockOSThread()

	// Get our args (we should have these from the Obj-C bridge)
	count := int(argc)
	var args []string
	if count > 0 && argv != nil {
		cargs := unsafe.Slice((**C.char)(unsafe.Pointer(argv)), count)

		args = make([]string, count)
		for i, arg := range cargs {
			if arg != nil {
				args[i] = C.GoString(arg)
			}
		}
	}

	// Set the baseDir before starting
	for i := 0; i+1 < len(args); i++ {
		if args[i] == "--assets" {
			sys.baseDir = args[i+1]
			break
		}
	}

	Logcat("SDL_main: Path ready, jumping to realMain")

	// Call realMain NOW that we're on the main thread
	realMain()

	return 0
}

func selectRenderer(cfgVal string) (Renderer, FontRenderer) {
	return &Renderer_VK{}, &FontRenderer_VK{}
}

func Logcat(s string) {
	fmt.Fprintln(os.Stderr, "[ikemen]", s)
}

//export GetMoltenVKWindowID
func GetMoltenVKWindowID() C.uint32_t {
	if sys.window == nil || sys.window.Window == nil {
		Logcat("GetMoltenVKWindowID: window not yet created")
		return 0
	}
	winID, e := sys.window.Window.GetID()
	if e != nil {
		Logcat(fmt.Sprintf("GetMoltenVKWindowID: %v", e))
		return 0
	}
	return C.uint32_t(winID)
}
