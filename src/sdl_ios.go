//go:build ios

package main

/*
#include "SDLWindowFix.h"
*/
import "C"

import "github.com/veandco/go-sdl2/sdl"

// Gets the actual device pixel size (Retina displays are special)
func (w *Window) GetDrawablePixelSize() (int32, int32) {
	winWidth, winHeight := sdl.Metal_GetDrawableSize(w.Window)
	return int32(winWidth), int32(winHeight)
}

func attachIOSMetalLayer(windowID uint32) {
	C.IkemenAttachMetalLayer(C.uint32_t(windowID))
}

func setOrientationHints() {
	sdl.SetHint(sdl.HINT_ORIENTATIONS, "LandscapeLeft LandscapeRight Portrait PortraitUpsideDown")
}
