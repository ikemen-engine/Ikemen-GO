package main

import (
	"github.com/Shopify/go-lua"
	"runtime"
)

func init() {
	runtime.LockOSThread()
}
func chk(err error) {
	if err != nil {
		panic(err)
	}
}
func main() {
	l := sys.init(640, 480)
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
