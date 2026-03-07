package main

/*
#cgo LDFLAGS: -lxmp
#include <xmp.h>
#include <stdlib.h>
#include <string.h>
#include <stdio.h>
*/
import "C"
import (
	"fmt"
	"io"
	"os"
	"runtime"
	"unsafe"

	"github.com/gopxl/beep/v2"
)

// ------------------------------------------------------------------
// xmStreamer wraps libxmp context for streaming
type xmStreamer struct {
	ctx        C.xmp_context
	channels   int
	sampleRate int
	buffer     []int16
	closed     bool
	err        error

	// runtime tracking
	posFrames   int // frames already produced (a frame == one sample per channel)
	totalFrames int // estimated total frames (from total_time)
}

// Stream fills the provided buffer with audio frames (Optimized version).
func (x *xmStreamer) Stream(samples [][2]float64) (int, bool) {
	if x.closed || x.err != nil {
		return 0, false
	}

	frameCount := len(samples)
	if frameCount*2 > len(x.buffer) {
		frameCount = len(x.buffer) / 2
	}

	res := C.xmp_play_buffer(x.ctx, unsafe.Pointer(&x.buffer[0]), C.int(frameCount*2*2), 0)
	if res < 0 {
		x.err = Error("xmp playback ended or failed")
		return 0, false
	}

	buf := x.buffer
	const scale = 1.0 / 32768.0
	for i := 0; i < frameCount; i++ {
		j := i * 2
		samples[i][0] = float64(buf[j]) * scale
		samples[i][1] = float64(buf[j+1]) * scale
	}
	return frameCount, true
}

// Err returns the last error that occurred.
func (x *xmStreamer) Err() error { return x.err }

// Close releases all libxmp resources.
func (x *xmStreamer) Close() error {
	if x.closed {
		return nil
	}
	x.closed = true
	C.xmp_end_player(x.ctx)
	C.xmp_release_module(x.ctx)
	C.xmp_free_context(x.ctx)
	return nil
}

func (x *xmStreamer) Position() int {
	return 0
}

// Seek attempts to position to absolute frame p.
// Beep's Seek uses sample-frame positions (frames == sample pairs).
func (x *xmStreamer) Seek(p int) error {
	return nil
}

func (x *xmStreamer) Len() int {
	return x.totalFrames
}

// newXMStreamer initializes a libxmp context for a given XM file.
func newXMStreamer(f *os.File) (*xmStreamer, error) {
	ctx := C.xmp_create_context()
	if ctx == nil {
		return nil, Error("failed to create xmp context")
	}

	// Use libxmp’s native loader instead of in-memory parsing.
	cpath := C.CString(f.Name())
	defer C.free(unsafe.Pointer(cpath))

	if C.xmp_load_module(ctx, cpath) != 0 {
		C.xmp_free_context(ctx)
		return nil, Error("failed to load XM module")
	}

	// Convert Go file to C FILE*
	// mode := C.CString("rb")
	// defer C.free(unsafe.Pointer(mode))
	// cFileStream := C.fdopen(C.int(f.Fd()), mode)
	// if cFileStream == nil {
	//     C.xmp_free_context(ctx)
	//     return nil, Error("fdopen failed")
	// }

	// if C.xmp_load_module_from_file(ctx, unsafe.Pointer(cFileStream), 0) != 0 {
	// 	C.xmp_free_context(ctx)
	// 	return nil, Error("failed to load XM module")
	// }

	var info C.struct_xmp_frame_info
	C.xmp_get_frame_info(ctx, &info)

	if C.xmp_start_player(ctx, audioFrequency, 0) != 0 {
		C.xmp_release_module(ctx)
		C.xmp_free_context(ctx)
		return nil, Error("failed to start XM player")
	}

	s := &xmStreamer{
		ctx:         ctx,
		channels:    2,
		sampleRate:  audioFrequency,
		totalFrames: int(float64(info.total_time) * float64(audioFrequency) / 1000.0),
		buffer:      make([]int16, audioOutLen*2), // 2048 stereo frames → lower memory
	}
	runtime.SetFinalizer(s, func(s *xmStreamer) { s.Close() })
	return s, nil
}

func xmpDecode(f io.ReadSeekCloser) (beep.StreamSeekCloser, beep.Format, error) {
	file, ok := f.(*os.File)
	if !ok {
		return nil, beep.Format{}, fmt.Errorf("xmpDecode: expected *os.File, got %T", f)
	}
	streamer, err := newXMStreamer(file)
	if err != nil {
		return nil, beep.Format{}, err
	}
	format := beep.Format{
		SampleRate:  audioFrequency,
		NumChannels: 2,
		Precision:   2,
	}
	return streamer, format, nil
}
