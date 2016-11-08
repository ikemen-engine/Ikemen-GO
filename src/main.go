package main

//go:generate go run ./gen/gen.go
import (
	"github.com/Shopify/go-lua"
	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"runtime"
	"time"
)

var windowWidth, windowHeight int32 = 640, 480
var GameWidth, GameHeight int32 = 320, 240
var WidthScale = float32(windowWidth) / float32(GameWidth)
var HeightScale = float32(windowHeight) / float32(GameHeight)
var window *glfw.Window
var gameEnd, frameSkip = false, false
var redrawWait = struct{ nextTime, lastDraw time.Time }{}
var testSprite *Sprite

func init() {
	runtime.LockOSThread()
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
		gl.Viewport(0, 0, int32(windowWidth), int32(windowHeight))
		gl.Clear(gl.COLOR_BUFFER_BIT)
		testSprite.Draw(160, 120, 1, 1, testSprite.GetPal(nil))
	}
}
func main() {
	chk(glfw.Init())
	defer glfw.Terminate()
	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 2)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	var err error
	window, err = glfw.CreateWindow(int(windowWidth), int(windowHeight),
		"Ikemen GO", nil, nil)
	chk(err)
	window.MakeContextCurrent()
	glfw.SwapInterval(1)
	chk(gl.Init())
	RenderInit()
	if testSprite, err = LoadFromSff("data/testv2.sff", 0, 0); err != nil {
		panic(err)
	}
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
