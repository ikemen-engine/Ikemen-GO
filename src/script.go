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
func numArg(l *lua.State, argi int) float64 {
	num, ok := l.ToNumber(argi)
	if !ok {
		panic(lua.RuntimeError(
			fmt.Sprintf("%d番目の引数が数ではありません。", argi)))
	}
	return num
}
func scriptCommonInit(l *lua.State) {
	luaRegister(l, "sndNew", func(l *lua.State) int {
		snd, err := LoadSndFile(strArg(l, 1))
		if err != nil {
			panic(lua.RuntimeError(err.Error()))
		}
		l.PushUserData(snd)
		return 1
	})
	luaRegister(l, "sndPlay", func(l *lua.State) int {
		ud := l.ToUserData(1)
		switch ud.(type) {
		case *Snd:
		default:
			panic(lua.RuntimeError("1番目の引数がSndではありません。"))
		}
		ud.(*Snd).Play(int32(numArg(l, 2)), int32(numArg(l, 3)))
		return 0
	})
	luaRegister(l, "playBGM", func(l *lua.State) int {
		bgm.Open(strArg(l, 1))
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
