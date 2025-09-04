//go:build !kinc

// system_glfw.go provides a desktop backend based on the GLFW library.
// It handles window creation, event polling and OpenGL context setup for
// the main engine. Compared to the Kinc backend, this implementation relies
// on GLFW's cross-platform facilities and is tightly coupled with OpenGL.
package main

import (
	"fmt"
	"image"

	glfw "github.com/go-gl/glfw/v3.3/glfw"
)

type Window struct {
	*glfw.Window
	title      string
	fullscreen bool
	x, y, w, h int
}

func (s *System) newWindow(w, h int) (*Window, error) {
	var err error
	var window *glfw.Window
	var monitor *glfw.Monitor

	// Step 1: initialize the GLFW library and global state. Failure here
	// propagates through chk(), aborting startup if the OS backend cannot
	// be prepared.
	chk(glfw.Init())

	// Step 2: query the primary monitor to determine fullscreen metrics.
	if monitor = glfw.GetPrimaryMonitor(); monitor == nil {
		return nil, fmt.Errorf("failed to obtain primary monitor")
	}

	// Step 3: resolve desired fullscreen/windowed mode. Ownership of the
	// resulting window remains with the System for later shutdown.
	_, forceWindowed := sys.cmdFlags["-windowed"]
	fullscreen := s.cfg.Video.Fullscreen && !forceWindowed

	// Step 4: compute window size and offsets prior to creation.
	var mode = monitor.GetVideoMode()
	var w2, h2 = w, h
	if sys.cfg.Video.WindowWidth > 0 || sys.cfg.Video.WindowHeight > 0 {
		w2, h2 = sys.cfg.Video.WindowWidth, sys.cfg.Video.WindowHeight
	}
	var x, y = (mode.Width - w2) / 2, (mode.Height - h2) / 2

	glfw.WindowHint(glfw.Resizable, glfw.True)

	// Step 5: configure the desired OpenGL context.
	if sys.cfg.Video.RenderMode == "OpenGL 3.2" {
		glfw.WindowHint(glfw.ContextVersionMajor, 3)
		glfw.WindowHint(glfw.ContextVersionMinor, 2)
		glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
		glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	} else {
		glfw.WindowHint(glfw.ContextVersionMajor, 2)
		glfw.WindowHint(glfw.ContextVersionMinor, 1)
	}

	// Step 6: create the native window. Errors are wrapped and returned so
	// that initialization can abort gracefully.
	// NOTE: Borderless fullscreen is in reality just a window without borders.
	if fullscreen && !s.cfg.Video.Borderless {
		window, err = glfw.CreateWindow(w2, h2, s.cfg.Config.WindowTitle, monitor, nil)
	} else {
		window, err = glfw.CreateWindow(w2, h2, s.cfg.Config.WindowTitle, nil, nil)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create window: %w", err)
	}

	// Step 7: apply window attributes and input callbacks. The System owns
	// the resulting window and is responsible for closing it on shutdown.
	if fullscreen {
		window.SetPos(0, 0)
		if s.cfg.Video.Borderless {
			window.SetAttrib(glfw.Decorated, 0)
			window.SetSize(mode.Width, mode.Height)
		}
		window.SetInputMode(glfw.CursorMode, glfw.CursorHidden)
	} else {
		window.SetSize(w2, h2)
		window.SetInputMode(glfw.CursorMode, glfw.CursorNormal)
		if s.cfg.Video.WindowCentered {
			window.SetPos(x, y)
		}
	}

	window.MakeContextCurrent()
	window.SetKeyCallback(keyCallback)
	window.SetCharModsCallback(charCallback)
	window.SetRefreshCallback(refreshCallback)

	// V-Sync
	if s.cfg.Video.VSync >= 0 {
		glfw.SwapInterval(s.cfg.Video.VSync)
	}

	// Step 8: return the initialized window to the caller.
	ret := &Window{window, s.cfg.Config.WindowTitle, fullscreen, x, y, w, h}
	return ret, err
}

func (w *Window) SwapBuffers() {
	w.Window.SwapBuffers()
	// Retrieve GL timestamp now
	glNow := glfw.GetTime()
	if glNow-sys.prevTimestamp >= 1 {
		sys.gameFPS = sys.absTickCountF / float32(glNow-sys.prevTimestamp)
		sys.absTickCountF = 0
		sys.prevTimestamp = glNow
	}
}

func (w *Window) SetIcon(icon []image.Image) {
	w.Window.SetIcon(icon)
}

func (w *Window) SetSwapInterval(interval int) {
	glfw.SwapInterval(interval)
}

func (w *Window) GetSize() (int, int) {
	return w.Window.GetSize()
}

func (w *Window) GetScaledViewportSize() (int32, int32, int32, int32) {
	// calculates a position and size for the viewport to fill the window while centered (see render_gl.go)
	// returns x, y, width, height respectively
	winWidth, winHeight := w.GetSize()
	ratioWidth := float32(winWidth) / float32(sys.gameWidth)
	ratioHeight := float32(winHeight) / float32(sys.gameHeight)
	var ratio float32
	var x, y, resizedWidth, resizedHeight int32 = 0, 0, int32(winWidth), int32(winHeight)

	if int32(winWidth) == sys.scrrect[2] && int32(winHeight) == sys.scrrect[3] {
		return 0, 0, int32(winWidth), int32(winHeight)
	}

	if ratioWidth < ratioHeight {
		ratio = ratioWidth
	} else {
		ratio = ratioHeight
	}

	if sys.cfg.Video.KeepAspect {
		resizedWidth = int32(float32(sys.gameWidth) * ratio)
		resizedHeight = int32(float32(sys.gameHeight) * ratio)

		// calculate offsets for the resized width to center it to the window
		if resizedWidth < int32(winWidth) {
			x = (int32(winWidth) - resizedWidth) / 2
		}
		if resizedHeight < int32(winHeight) {
			y = (int32(winHeight) - resizedHeight) / 2
		}
	}

	return x, y, resizedWidth, resizedHeight
}

func (w *Window) GetClipboardString() string {
	return w.Window.GetClipboardString()
}

func (w *Window) toggleFullscreen() {
	var mode = glfw.GetPrimaryMonitor().GetVideoMode()

	if w.fullscreen {
		w.SetAttrib(glfw.Decorated, 1)
		w.SetMonitor(&glfw.Monitor{}, w.x, w.y, w.w, w.h, mode.RefreshRate)
		w.SetInputMode(glfw.CursorMode, glfw.CursorNormal)
	} else {
		w.SetAttrib(glfw.Decorated, 0)
		if sys.cfg.Video.Borderless {
			w.SetSize(mode.Width, mode.Height)
			w.SetMonitor(&glfw.Monitor{}, 0, 0, mode.Width, mode.Height, mode.RefreshRate)
		} else {
			w.x, w.y = w.GetPos()
			w.SetMonitor(glfw.GetPrimaryMonitor(), w.x, w.y, w.w, w.h, mode.RefreshRate)
		}
		w.SetInputMode(glfw.CursorMode, glfw.CursorHidden)
	}
	if sys.cfg.Video.VSync != -1 {
		glfw.SwapInterval(sys.cfg.Video.VSync)
	}
	w.fullscreen = !w.fullscreen
}

// pollEvents pumps the GLFW event loop. It must run on the same goroutine
// that created the window as GLFW is not goroutine-safe.
func (w *Window) pollEvents() {
	glfw.PollEvents()
}

func (w *Window) shouldClose() bool {
	return w.Window.ShouldClose()
}

// Close releases GLFW's global resources. There is no error return from
// glfw.Terminate(), so shutdown proceeds without additional propagation.
func (w *Window) Close() {
	glfw.Terminate()
}

func refreshCallback(w *glfw.Window) {
	gfx.EndFrame()
	w.SwapBuffers()
}

func keyCallback(_ *glfw.Window, key Key, _ int, action glfw.Action, mk ModifierKey) {
	switch action {
	case glfw.Release:
		OnKeyReleased(key, mk)
	case glfw.Press:
		OnKeyPressed(key, mk)
	}
}

func charCallback(_ *glfw.Window, char rune, mk ModifierKey) {
	OnTextEntered(string(char))
}
