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
	err := glfw.Init()
	if err != nil {
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
	if err := lua.DoFile(l, "script/main.lua"); err != nil {
		panic(err)
	}
	for !window.ShouldClose() {
		window.SwapBuffers()
		glfw.PollEvents()
	}
}
