package main

import (
	"fmt"
	"github.com/Shopify/go-lua"
)

func luaRegister(l *lua.State, name string, f func(*lua.State) int) {
	l.PushGoFunction(f)
	l.SetGlobal(name)
}
func strArg(l *lua.State, argi int) string {
	str, ok := l.ToString(argi)
	if !ok {
		panic(lua.RuntimeError(
			fmt.Sprintf("%d番目の引数が文字列ではありません。", argi)))
	}
	return str
}
func scriptCommonInit(l *lua.State) {
	luaRegister(l, "playBGM", func(l *lua.State) int {
		bgm.open(strArg(l, 1))
		return 0
	})
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
