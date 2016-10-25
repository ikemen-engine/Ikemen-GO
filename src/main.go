package main

import (
	"github.com/Shopify/go-lua"
	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"runtime"
	"time"
)

var window *glfw.Window
var windowWidth, windowHeight = 640, 480
var gameEnd, frameSkip = false, false
var redrawWait = struct{ nextTime, lastDraw time.Time }{}

func init() {
	runtime.LockOSThread()
}
func chk(err error) {
	if err != nil {
		panic(err)
	}
}
func await(fps int) {
	if !frameSkip {
		window.SwapBuffers()
	}
	now := time.Now()
	diff := redrawWait.nextTime.Sub(now)
	wait := time.Second / time.Duration(fps)
	redrawWait.nextTime = redrawWait.nextTime.Add(wait)
	switch {
	case diff >= 0 && diff < wait+2*time.Millisecond:
		time.Sleep(diff)
		fallthrough
	case now.Sub(redrawWait.lastDraw) > 250*time.Millisecond:
		fallthrough
	case diff >= -17*time.Millisecond:
		redrawWait.lastDraw = now
		frameSkip = false
	default:
		if diff < -150*time.Millisecond {
			redrawWait.nextTime = now.Add(wait)
		}
		frameSkip = true
	}
	glfw.PollEvents()
	gameEnd = window.ShouldClose()
	if !frameSkip {
		windowWidth, windowHeight = window.GetFramebufferSize()
		gl.Viewport(0, 0, int32(windowWidth), int32(windowHeight))
		gl.Clear(gl.COLOR_BUFFER_BIT)
	}
}
func main() {
	chk(glfw.Init())
	defer glfw.Terminate()
	chk(gl.Init())
	var err error
	window, err =
		glfw.CreateWindow(windowWidth, windowHeight, "Ikemen GO", nil, nil)
	chk(err)
	window.MakeContextCurrent()
	glfw.SwapInterval(1)
	l := lua.NewState()
	audioOpen()
	lua.OpenLibraries(l)
	systemScriptInit(l)
	if err := lua.DoFile(l, "script/main.lua"); err != nil {
		switch err.(type) {
		case lua.RuntimeError:
			errstr := err.Error()
			if len(errstr) < 10 || errstr[len(errstr)-10:] != "<game end>" {
				panic(err)
			}
		default:
			panic(err)
		}
	}
}
