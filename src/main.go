package main

import (
	"github.com/Shopify/go-lua"
	"github.com/go-gl/glfw/v3.2/glfw"
	"runtime"
	"time"
)

var window *glfw.Window
var gameEnd, frameSkip = false, false
var redrawWait = struct{ nextTime, lastDraw time.Time }{}

func init() {
	runtime.LockOSThread()
}
func await(fps int) {
	if window.ShouldClose() {
		gameEnd = true
		return
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
		if -diff > 150*time.Millisecond {
			redrawWait.nextTime = now.Add(wait)
		}
		frameSkip = true
	}
	window.SwapBuffers()
	glfw.PollEvents()
}
func main() {
	if err := glfw.Init(); err != nil {
		panic(err)
	}
	defer glfw.Terminate()
	w, err := glfw.CreateWindow(640, 480, "Ikemen GO", nil, nil)
	if err != nil {
		panic(err)
	}
	window = w
	window.MakeContextCurrent()
	l := lua.NewState()
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
