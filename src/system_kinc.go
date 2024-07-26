//go:build kinc

package main

import (
	"image"
	"unsafe"
)

/*
#include <kinc/system.h>
#include <kinc/window.h> // kinc_window_set_close_callback()
#include <kinc/graphics4/graphics.h> // kinc_g4_swap_buffers()

typedef bool (*close_callback_t)(void *);

#if _WIN32
extern __declspec(dllexport) uint8_t close_callback(void *);
#else
extern uint8_t close_callback(void *);
#endif
*/
import "C"

type Window struct {
	width      int
	height     int
	fullscreen bool
	closing    bool
	handle     C.int
}

func (s *System) newWindow(w, h int) (*Window, error) {
	ret := &Window{width: w, height: h}
	handle := C.kinc_init(C.CString(s.windowTitle), C.int(w), C.int(h), nil, nil)
	C.kinc_window_set_close_callback(handle, (C.close_callback_t)(C.close_callback),
		unsafe.Pointer(&ret.closing))
	// TODO: add keyboard input callbacks
	return ret, nil
}

// export close_callback
func close_callback(closing unsafe.Pointer) bool {
	*(*bool)(closing) = true
	return true
}

func (w *Window) SwapBuffers() {
	C.kinc_g4_swap_buffers()
}

func (w *Window) SetIcon(icon []image.Image) {
	// TODO
}

func (w *Window) SetSwapInterval(interval int) {
	// TODO
}

func (w *Window) GetSize() (int, int) {
	return w.width, w.height
}

func (w *Window) GetClipboardString() (string, error) {
	// TODO
	return "", nil
}

func (w *Window) toggleFullscreen() {
	// TODO
}

func (w *Window) pollEvents() {
	C.kinc_internal_frame()
}

func (w *Window) shouldClose() bool {
	return w.closing
}

func (w *Window) Close() {
}
