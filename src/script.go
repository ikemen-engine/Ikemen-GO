package main

import "github.com/Shopify/go-lua"

func luaRegister(l *lua.State, name string, f func(*lua.State) int) {
	l.PushGoFunction(f)
	l.SetGlobal(name)
}
func scriptCommonInit(l *lua.State) {
}
func systemScriptInit(l *lua.State) {
	scriptCommonInit(l)
	luaRegister(l, "refresh", func(*lua.State) int {
		await(60)
		if gameEnd {
			panic(lua.RuntimeError("<game end>"))
		}
		return 0
	})
}
