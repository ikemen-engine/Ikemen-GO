package main

import (
	"github.com/Shopify/go-lua"
	"github.com/go-gl/glfw/v3.2/glfw"
	"runtime"
)

func init() {
	runtime.LockOSThread()
}
func main() {
	if err := glfw.Init(); err != nil {
		panic(err)
	}
	defer glfw.Terminate()
	window, err := glfw.CreateWindow(640, 480, "Ikemen GO", nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()
	l := lua.NewState()
	lua.OpenLibraries(l)
	l.PushGoFunction(func(*lua.State) int {
		if window.ShouldClose() {
			panic(lua.RuntimeError("<game end>"))
		}
		window.SwapBuffers()
		glfw.PollEvents()
		return 0
	})
	l.SetGlobal("refresh")
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
