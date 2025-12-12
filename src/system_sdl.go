package main

import (
	"fmt"
	"image"
	"image/draw"

	"github.com/veandco/go-sdl2/sdl"
)

type Window struct {
	*sdl.Window
	title      string
	x, y, w, h int
	fullscreen bool
	closeflag  bool
}

func (s *System) newWindow(w, h int) (*Window, error) {
	var err error
	var window *sdl.Window

	// Initialize SDL
	chk(sdl.Init(sdl.INIT_VIDEO | sdl.INIT_JOYSTICK | sdl.INIT_EVENTS | sdl.INIT_GAMECONTROLLER | sdl.INIT_HAPTIC | sdl.INIT_TIMER))

	mode, err := sdl.GetDesktopDisplayMode(0)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain primary monitor")
	}

	// "-windowed" overrides the configuration setting but does not change it
	_, forceWindowed := sys.cmdFlags["-windowed"]
	fullscreen := s.cfg.Video.Fullscreen && !forceWindowed

	// Calculate window size & offset it
	var w2, h2 = int32(w), int32(h)
	if sys.cfg.Video.WindowWidth > 0 || sys.cfg.Video.WindowHeight > 0 {
		w2, h2 = int32(sys.cfg.Video.WindowWidth), int32(sys.cfg.Video.WindowHeight)
	}
	var x, y = (mode.W - w2) / 2, (mode.H - h2) / 2

	window.SetResizable(true)
	var windowFlags sdl.WindowFlags = sdl.WINDOW_INPUT_FOCUS

	if sys.cfg.Video.RenderMode == "OpenGL 3.2" {
		err = sdl.GLSetAttribute(sdl.GL_CONTEXT_PROFILE_MASK, sdl.GL_CONTEXT_PROFILE_CORE) // only GL 3.2 needs this
		err = sdl.GLSetAttribute(sdl.GL_CONTEXT_MAJOR_VERSION, 3)
		err = sdl.GLSetAttribute(sdl.GL_CONTEXT_MINOR_VERSION, 2)
		err = sdl.GLSetAttribute(sdl.GL_CONTEXT_FORWARD_COMPATIBLE_FLAG, 1)
		windowFlags |= sdl.WINDOW_OPENGL
	} else if sys.cfg.Video.RenderMode == "OpenGL 2.1" {
		err = sdl.GLSetAttribute(sdl.GL_CONTEXT_MAJOR_VERSION, 2)
		err = sdl.GLSetAttribute(sdl.GL_CONTEXT_MINOR_VERSION, 1)
		windowFlags |= sdl.WINDOW_OPENGL
	} else {
		windowFlags |= sdl.WINDOW_VULKAN
		// Ensure core profile is NOT set for Vulkan
		if err := sdl.GLSetAttribute(sdl.GL_CONTEXT_PROFILE_MASK, 0); err != nil {
			return nil, err
		}
		if err := sdl.GLSetAttribute(sdl.GL_CONTEXT_FLAGS, 0); err != nil {
			return nil, err
		}
	}

	// Create main window.
	// NOTE: borderless fullscreen is in reality just a window without borders.
	//       We want fake fullscreen so as not to mess with the users' other windows!
	//       On Windows, true exclusive fullscreen can resize other windows and blank
	//       the display if the game resolution is different from the desktop resolution.
	//       On macOS, this can cause flickering behavior. "Fake" fullscreen prevents all
	//       erratic behavior on all platforms.
	if fullscreen {
		windowFlags |= sdl.WINDOW_FULLSCREEN_DESKTOP
	} else {
		windowFlags |= sdl.WINDOW_SHOWN
	}

	// Because we ought to set these flags the same regardless of fullscreen or not.
	// It makes no sense to have the resizable flag on a window without borders.
	if !s.cfg.Video.Borderless {
		windowFlags |= sdl.WINDOW_RESIZABLE
	} else {
		windowFlags |= sdl.WINDOW_BORDERLESS
	}

	window, err = sdl.CreateWindow(s.cfg.Config.WindowTitle, sdl.WINDOWPOS_CENTERED, sdl.WINDOWPOS_CENTERED, w2, h2, windowFlags)
	if err != nil {
		return nil, fmt.Errorf("failed to create window: %w", err)
	}

	// Set window attributes
	if fullscreen {
		window.SetPosition(0, 0)
		if s.cfg.Video.Borderless {
			window.SetBordered(false)
			window.SetSize(mode.W, mode.H)
		}
		sdl.ShowCursor(sdl.DISABLE)
	} else {
		if !s.cfg.Video.Borderless {
			window.SetBordered(true)
		} else {
			window.SetBordered(false)
		}
		window.SetSize(w2, h2)
		sdl.ShowCursor(sdl.ENABLE)
		if s.cfg.Video.WindowCentered {
			window.SetPosition(x, y)
		}
	}

	if sys.cfg.Video.RenderMode == "OpenGL 3.2" || sys.cfg.Video.RenderMode == "OpenGL 2.1" {
		// V-Sync
		if s.cfg.Video.VSync >= 0 {
			sdl.GLSetSwapInterval(s.cfg.Video.VSync)
		}
	}

	for i := range input.controllers {
		input.controllerstate[i] = &ControllerState{Buttons: make(map[sdl.GameControllerButton]byte)}
	}

	ret := &Window{window, s.cfg.Config.WindowTitle, int(x), int(y), w, h, fullscreen, false}
	return ret, err
}

func (w *Window) SwapBuffers() {
	w.Window.GLSwap()
	// Retrieve GL timestamp now
	now := sdl.GetPerformanceCounter()
	diff := float32(now - sys.prevTimestamp)
	if diff*float32(sdl.GetPerformanceFrequency()) >= 1 {
		sys.gameFPS = float32(sdl.GetPerformanceFrequency()) / diff
		sys.absTickCountF = 0
		sys.prevTimestamp = now
	}
}

func (w *Window) imageToSurface(img image.Image) (*sdl.Surface, error) {
	bounds := img.Bounds()

	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)

	var format uint32 = uint32(sdl.PIXELFORMAT_ABGR8888)

	surface, err := sdl.CreateRGBSurfaceWithFormat(
		0,
		int32(bounds.Dx()),
		int32(bounds.Dy()),
		32,
		format,
	)
	if err != nil {
		return nil, err
	}

	// 4. Lock the surface, copy data, and unlock
	if err := surface.Lock(); err != nil {
		surface.Free()
		return nil, err
	}

	// Copy the pixel data from the Go image (rgba.Pix) to the SDL surface (surface.Pixels())
	copy(surface.Pixels(), rgba.Pix)

	surface.Unlock()
	return surface, nil
}

func (w *Window) SetIcon(icon []image.Image) {
	if surface, err := w.imageToSurface(icon[0]); err == nil {
		w.Window.SetIcon(surface)
	}
}

func (w *Window) SetSwapInterval(interval int) {
	if sys.cfg.Video.RenderMode == "OpenGL 3.2" || sys.cfg.Video.RenderMode == "OpenGL 2.1" {
		sdl.GLSetSwapInterval(interval)
	} else {
		gfx.SetVSync()
	}
}

func (w *Window) GetSize() (int, int) {
	w2, h2 := w.Window.GetSize()
	return int(w2), int(h2)
}

// Calculates a position and size for the viewport to fill the window while centered (see render_gl.go)
// Returns x, y, width, height respectively
func (w *Window) GetScaledViewportSize() (int32, int32, int32, int32) {
	winWidth, winHeight := w.GetSize()

	// If aspect ratio should not be kept, just return full window
	if !sys.cfg.Video.KeepAspect {
		return 0, 0, int32(winWidth), int32(winHeight)
	}

	var x, y, resizedWidth, resizedHeight int32 = 0, 0, int32(winWidth), int32(winHeight)

	// Select stage or default aspect ratio
	aspectGame := sys.getCurrentAspect()
	aspectWindow := float32(winWidth) / float32(winHeight)

	// Keep aspect ratio
	if aspectWindow > aspectGame {
		// Window is wider: black bars on sides
		resizedHeight = int32(winHeight)
		resizedWidth = int32(float32(resizedHeight) * aspectGame)
		x = (int32(winWidth) - resizedWidth) / 2
		y = 0
	} else {
		// Window is taller: black bars on top and bottom
		resizedWidth = int32(winWidth)
		resizedHeight = int32(float32(resizedWidth) / aspectGame)
		x = 0
		y = (int32(winHeight) - resizedHeight) / 2
	}

	return x, y, resizedWidth, resizedHeight
}

func (w *Window) GetClipboardString() string {
	s, err := sdl.GetClipboardText()
	if err != nil {
		return ""
	}
	return s
}

func (w *Window) toggleFullscreen() {
	// var mode, _ = sdl.GetDisplayMode(0, 0)

	if w.fullscreen {
		w.Window.SetBordered(false)
		w.Window.SetFullscreen(0)
		sdl.ShowCursor(sdl.ENABLE)
		w.Window.SetSize(int32(w.w), int32(w.h))
		w.Window.SetPosition(int32(w.x), int32(w.y))
	} else {
		x2, y2 := w.Window.GetPosition()
		w2, h2 := w.Window.GetSize()

		w.x, w.y = int(x2), int(y2)
		w.w, w.h = int(w2), int(h2)

		w.Window.SetBordered(!sys.cfg.Video.Borderless)
		w.Window.SetSize(int32(sys.cfg.Video.WindowWidth), int32(sys.cfg.Video.WindowHeight))
		w.Window.SetFullscreen(uint32(sdl.WINDOW_FULLSCREEN_DESKTOP))
		sdl.ShowCursor(sdl.DISABLE)
	}
	if sys.cfg.Video.VSync != -1 && (sys.cfg.Video.RenderMode == "OpenGL 3.2" || sys.cfg.Video.RenderMode == "OpenGL 2.1") {
		sdl.GLSetSwapInterval(sys.cfg.Video.VSync)
	}
	w.fullscreen = !w.fullscreen
}

func convertI16toI8(val int16) (converted int8) {
	const INT16_MAX float32 = 32768.0
	const INT8_MAX float32 = 128.0
	if val < 0 {
		return int8((float32(val) / INT16_MAX) * INT8_MAX)
	} else {
		return int8((float32(val) / (INT16_MAX - 1)) * (INT8_MAX - 1))
	}
}

// Helper to find the array index based on the SDL Instance ID
func findControllerIndex(instanceID sdl.JoystickID) int {
    for i, ctrl := range input.controllers {
        if ctrl != nil {
            // We need to check the Joystick associated with the GameController
            if joy := ctrl.Joystick(); joy != nil {
                if joy.InstanceID() == instanceID {
                    return i
                }
            }
        }
    }
    return -1
}

func (w *Window) pollEvents() {
    const MAX_VALUE float32 = 32768.0
    for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
        switch t := event.(type) {
        case sdl.ControllerAxisEvent:
            // FIX: Map Instance ID (t.Which) to Array Index
            if idx := findControllerIndex(t.Which); idx != -1 {
                input.controllerstate[idx].Axes[t.Axis] = convertI16toI8(t.Value)
            }
        case sdl.ControllerButtonEvent:
            // FIX: Map Instance ID (t.Which) to Array Index
            if idx := findControllerIndex(t.Which); idx != -1 {
                input.controllerstate[idx].Buttons[t.Button] = byte(t.State)
            }
        case sdl.QuitEvent:
            w.closeflag = true
        case sdl.KeyboardEvent:
            if t.State == sdl.PRESSED {
                OnKeyPressed(t.Keysym.Sym, t.Keysym.Mod)
            } else if t.State == sdl.RELEASED {
                OnKeyReleased(t.Keysym.Sym, t.Keysym.Mod)
            }
        case sdl.WindowEvent:
            if t.Event == sdl.WINDOWEVENT_EXPOSED {
                if sys.cfg.Video.RenderMode == "OpenGL 3.2" || sys.cfg.Video.RenderMode == "OpenGL 2.1" {
                    gfx.EndFrame()
                    w.SwapBuffers()
                }
            } else if t.Event == sdl.WINDOWEVENT_CLOSE {
                w.closeflag = true
            }
        case sdl.TextInputEvent:
            if len(t.Text) > 0 {
                OnTextEntered(t.Text)
            }
        case sdl.JoyDeviceAddedEvent:
            // t.Which here IS the device index (0, 1, 2...), so this is safe
            joyS := int(t.Which)
            if joyS < len(input.controllers) {
                input.controllers[joyS] = sdl.GameControllerOpen(joyS)
                if input.controllers[joyS] != nil {
                    input.controllerstate[joyS].HasRumble = input.controllers[joyS].HasRumble()
                }
            }
        case sdl.JoyDeviceRemovedEvent:
            // FIX: Map Instance ID (t.Which) to Array Index
            if idx := findControllerIndex(t.Which); idx != -1 {
                if controller := input.controllers[idx]; controller != nil {
                    controller.Close()
                    input.controllers[idx] = nil // Clear the reference
                }
            }
        }
    }
}

func (w *Window) shouldClose() bool {
	return w.closeflag
}

func (w *Window) Close() {
	if w.Window != nil {
		w.Window.Destroy()
		w.Window = nil
	}
	sdl.Quit()
}
