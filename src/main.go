package main

import (
	"fmt"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/yuin/gopher-lua"
	"runtime"
	"strings"
)

func init() {
	runtime.LockOSThread()
}
func chk(err error) {
	if err != nil {
		panic(err)
	}
}
func unimplemented() {
	_, fn, line, _ := runtime.Caller(1)
	panic(Error(fmt.Sprintf("%v:%v: unimplemented", fn, line)))
}
func main() {
	chk(glfw.Init())
	defer glfw.Terminate()
	l := sys.init(640, 480)
	if err := l.DoFile("script/main.lua"); err != nil {
		switch err.(type) {
		case *lua.ApiError:
			errstr := strings.Split(err.Error(), "\n")[0]
			if len(errstr) < 10 || errstr[len(errstr)-10:] != "<game end>" {
				panic(err)
			}
		default:
			panic(err)
		}
	}
}
