package main

//go:generate go run ./gen/gen.go
import (
	"github.com/Shopify/go-lua"
	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"runtime"
	"time"
)

var scrrect = [4]int32{0, 0, 320, 240}
var gameWidth, gameHeight int32 = 320, 240
var widthScale, heightScale float32 = 1, 1
var window *glfw.Window
var gameEnd, frameSkip = false, false
var redrawWait = struct{ nextTime, lastDraw time.Time }{}
var brightness = 256

func init() {
	runtime.LockOSThread()
}
func setWindowSize(w, h int32) {
	scrrect[2], scrrect[3] = w, h
	if scrrect[2]*3 > scrrect[3]*4 {
		gameWidth, gameHeight = scrrect[2]*3*320/(scrrect[3]*4), 240
	} else {
		gameWidth, gameHeight = 320, scrrect[3]*4*240/(scrrect[2]*3)
	}
	widthScale = float32(scrrect[2]) / float32(gameWidth)
	heightScale = float32(scrrect[3]) / float32(gameHeight)
}
func chk(err error) {
	if err != nil {
		panic(err)
	}
}
func await(fps int) {
	playSound()
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
		gl.Viewport(0, 0, int32(scrrect[2]), int32(scrrect[3]))
		gl.Clear(gl.COLOR_BUFFER_BIT)
	}
}
func main() {
	chk(glfw.Init())
	defer glfw.Terminate()
	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 2)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	setWindowSize(640, 480)
	var err error
	window, err = glfw.CreateWindow(int(scrrect[2]), int(scrrect[3]),
		"Ikemen GO", nil, nil)
	chk(err)
	window.MakeContextCurrent()
	glfw.SwapInterval(1)
	chk(gl.Init())
	RenderInit()
	audioOpen()
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
