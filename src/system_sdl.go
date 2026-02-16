package main

import (
	"fmt"
	"image"
	"image/draw"
	"runtime"

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
	var x, y int32
	var w2, h2 int32 = int32(w), int32(h)
	var fullscreen bool

	if runtime.GOOS == "android" {
		// On Android, we MUST use 0,0 or SDL ignores it anyway,
		// but flags are the critical part.
		window, err = sdl.CreateWindow(
			s.cfg.Config.WindowTitle,
			sdl.WINDOWPOS_UNDEFINED,
			sdl.WINDOWPOS_UNDEFINED,
			0, 0, // Android ignores these and uses screen size
			sdl.WINDOW_SHOWN|sdl.WINDOW_OPENGL|sdl.WINDOW_FULLSCREEN,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to create window: %w", err)
		}
		fullscreen = true
	} else {

		var windowFlags sdl.WindowFlags = sdl.WINDOW_INPUT_FOCUS

		// 1. INITIALIZATION & HINTS
		if runtime.GOOS == "android" {
			// Android ignores these anyway, but WINDOW_FULLSCREEN is the safest anchor
			windowFlags |= sdl.WINDOW_FULLSCREEN | sdl.WINDOW_SHOWN
		} else {
			chk(sdl.Init(sdl.INIT_AUDIO | sdl.INIT_VIDEO | sdl.INIT_JOYSTICK | sdl.INIT_EVENTS | sdl.INIT_GAMECONTROLLER | sdl.INIT_HAPTIC | sdl.INIT_TIMER))
		}

		// 2. DISPLAY MODE (The Crash Point)
		var desktopW, desktopH int32
		if runtime.GOOS != "android" {
			mode, err := sdl.GetDesktopDisplayMode(0)
			if err != nil {
				return nil, fmt.Errorf("failed to obtain primary monitor")
			}
			desktopW, desktopH = mode.W, mode.H
		} else {
			// On Android, we don't query the desktop. We let SDL fill the window to the surface.
			desktopW, desktopH = w2, h2
		}

		// 3. CALCULATION
		_, forceWindowed := sys.cmdFlags["-windowed"]
		fullscreen := s.cfg.Video.Fullscreen && !forceWindowed

		// Override default sizes if config specifies
		if sys.cfg.Video.WindowWidth > 0 || sys.cfg.Video.WindowHeight > 0 {
			w2, h2 = int32(sys.cfg.Video.WindowWidth), int32(sys.cfg.Video.WindowHeight)
		}

		if runtime.GOOS != "android" {
			x, y = (desktopW-w2)/2, (desktopH-h2)/2
		}

		// 4. RENDERER PROFILE SETUP
		renderName := gfx.GetName()
		if renderName == "OpenGL ES 3.2" {
			sdl.GLSetAttribute(sdl.GL_CONTEXT_PROFILE_MASK, sdl.GL_CONTEXT_PROFILE_ES)
			sdl.GLSetAttribute(sdl.GL_CONTEXT_MAJOR_VERSION, 3)
			sdl.GLSetAttribute(sdl.GL_CONTEXT_MINOR_VERSION, 2)
			sdl.GLSetAttribute(sdl.GL_ALPHA_SIZE, 0)
			sdl.GLSetAttribute(sdl.GL_DEPTH_SIZE, 24)
			windowFlags |= sdl.WINDOW_OPENGL
		} else if renderName == "OpenGL 3.2" {
			sdl.GLSetAttribute(sdl.GL_CONTEXT_PROFILE_MASK, sdl.GL_CONTEXT_PROFILE_CORE)
			sdl.GLSetAttribute(sdl.GL_CONTEXT_MAJOR_VERSION, 3)
			sdl.GLSetAttribute(sdl.GL_CONTEXT_MINOR_VERSION, 2)
			sdl.GLSetAttribute(sdl.GL_CONTEXT_FORWARD_COMPATIBLE_FLAG, 1)
			// Only load debug context if we will use it, so we avoid useless overhead
			if sys.cfg.Video.RendererDebugMode {
				sdl.GLSetAttribute(sdl.GL_CONTEXT_FLAGS, sdl.GL_CONTEXT_DEBUG_FLAG)
			}
			windowFlags |= sdl.WINDOW_OPENGL
		} else if renderName == "OpenGL 2.1" {
			sdl.GLSetAttribute(sdl.GL_CONTEXT_MAJOR_VERSION, 2)
			sdl.GLSetAttribute(sdl.GL_CONTEXT_MINOR_VERSION, 1)
			windowFlags |= sdl.WINDOW_OPENGL
		} else {
			windowFlags |= sdl.WINDOW_VULKAN
		}

		// 5. WINDOW CREATION
		if runtime.GOOS != "android" {
			if fullscreen {
				windowFlags |= sdl.WINDOW_FULLSCREEN_DESKTOP
			}
			windowFlags |= sdl.WINDOW_SHOWN
			if !s.cfg.Video.Borderless {
				windowFlags |= sdl.WINDOW_RESIZABLE
			} else {
				windowFlags |= sdl.WINDOW_BORDERLESS
			}
		}

		// On Android, use 0,0 for pos and allow SDL to resize w2/h2 to match the device screen
		posX, posY := x, y
		if runtime.GOOS == "android" {
			posX, posY = 0, 0
		}

		window, err = sdl.CreateWindow(s.cfg.Config.WindowTitle, posX, posY, w2, h2, windowFlags)
		if err != nil {
			return nil, fmt.Errorf("failed to create window: %w", err)
		}

		// 6. POST-CREATION ATTRIBUTES
		window.SetResizable(true)
		if fullscreen {
			window.SetPosition(0, 0)
			if s.cfg.Video.Borderless {
				window.SetBordered(false)
				window.SetSize(desktopW, desktopH)
			}
			sdl.ShowCursor(sdl.DISABLE)
		} else {
			window.SetBordered(!s.cfg.Video.Borderless)
			window.SetSize(w2, h2)
			sdl.ShowCursor(sdl.ENABLE)
			if s.cfg.Video.WindowCentered {
				window.SetPosition(x, y)
			}
		}
	}

	for i := range input.controllers {
		input.controllerstate[i] = &ControllerState{Buttons: make(map[sdl.GameControllerButton]byte)}
	}

	return &Window{window, s.cfg.Config.WindowTitle, int(x), int(y), w, h, fullscreen, false}, nil
}

func (w *Window) SwapBuffers() {
	w.Window.GLSwap()
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
	gfx.SetVSync(interval)
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
	if w.fullscreen {
		w.Window.SetFullscreen(0)
		w.Window.SetBordered(!sys.cfg.Video.Borderless)
		w.Window.SetSize(int32(w.w), int32(w.h))
		if runtime.GOOS != "android" && sys.cfg.Video.WindowCentered {
			displayIndex, err := w.Window.GetDisplayIndex()
			if err != nil {
				displayIndex = 0 // default to primary monitor if we have no success
			}
			w.Window.SetPosition(sdl.WINDOWPOS_CENTERED_MASK|int32(displayIndex), sdl.WINDOWPOS_CENTERED_MASK|int32(displayIndex))
		}
		sdl.ShowCursor(sdl.ENABLE)
	} else {
		x2, y2 := w.Window.GetPosition()
		w2, h2 := w.Window.GetSize()

		w.x, w.y = int(x2), int(y2)
		w.w, w.h = int(w2), int(h2)

		w.Window.SetBordered(!sys.cfg.Video.Borderless)
		w.Window.SetFullscreen(uint32(sdl.WINDOW_FULLSCREEN_DESKTOP))
		sdl.ShowCursor(sdl.DISABLE)
	}
	if sys.cfg.Video.VSync != -1 {
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

// Zero out controller state and ensure maps stay allocated so stale values do not leak between devices.
func resetControllerState(idx int) {
	if idx < 0 || idx >= len(input.controllerstate) {
		return
	}
	if input.controllerstate[idx] == nil {
		input.controllerstate[idx] = &ControllerState{Buttons: make(map[sdl.GameControllerButton]byte)}
	}
	input.controllerstate[idx].Axes = [6]int8{}
	for k := range input.controllerstate[idx].Buttons {
		delete(input.controllerstate[idx].Buttons, k)
	}
	input.controllerstate[idx].HasRumble = false
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

// Finds the first free slot in the controller array.
func findFreeControllerSlot() int {
	for i, ctrl := range input.controllers {
		if ctrl == nil {
			return i
		}
	}
	return -1
}

// Open the SDL device index and attach it to the first available slot.
func attachController(deviceIndex int) {
	controller := sdl.GameControllerOpen(deviceIndex)
	if controller == nil {
		return
	}

	// If SDL reuses an existing instance ID, reuse its slot; otherwise, pick the first free one.
	slot := findControllerIndex(controller.Joystick().InstanceID())
	if slot == -1 {
		slot = findFreeControllerSlot()
	}

	if slot == -1 {
		controller.Close()
		return
	}

	input.controllers[slot] = controller
	resetControllerState(slot)
	input.controllerstate[slot].HasRumble = controller.HasRumble()
}

func (w *Window) UpdateDebugFPS() {
	now := sdl.GetPerformanceCounter()
	freq := float32(sdl.GetPerformanceFrequency())
	diff := float32(now - sys.gameFPSprevcount)

	if diff > 0 {
		instantFPS := freq / diff
		// Use an EMA to apply smoothing
		sys.gameFPS = (sys.gameFPS * 0.95) + (float32(instantFPS) * 0.05)
	}

	sys.gameFPSprevcount = now
}

func (w *Window) pollEvents() {
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
				renderName := gfx.GetName()
				if renderName == "OpenGL 3.2" || renderName == "OpenGL 2.1" {
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
			attachController(int(t.Which))
		case sdl.JoyDeviceRemovedEvent:
			// FIX: Map Instance ID (t.Which) to Array Index
			if idx := findControllerIndex(t.Which); idx != -1 {
				if controller := input.controllers[idx]; controller != nil {
					controller.Close()
					input.controllers[idx] = nil // Clear the reference
					resetControllerState(idx)
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
