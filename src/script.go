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
	luaRegister(l, "sffNew", func(l *lua.State) int {
		sff, err := LoadSff(strArg(l, 1), false)
		if err != nil {
			panic(lua.RuntimeError(err.Error()))
		}
		l.PushUserData(sff)
		return 1
	})
	luaRegister(l, "sndNew", func(l *lua.State) int {
		snd, err := LoadSnd(strArg(l, 1))
		if err != nil {
			panic(lua.RuntimeError(err.Error()))
		}
		l.PushUserData(snd)
		return 1
	})
	luaRegister(l, "fontNew", func(l *lua.State) int {
		fnt, err := LoadFnt(strArg(l, 1))
		if err != nil {
			panic(lua.RuntimeError(err.Error()))
		}
		l.PushUserData(fnt)
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
	luaRegister(l, "textImgNew", func(*lua.State) int {
		l.PushUserData(NewTextSprite())
		return 1
	})
	luaRegister(l, "textImgSetFont", func(*lua.State) int {
		ud := l.ToUserData(1)
		switch ud.(type) {
		case *TextSprite:
		default:
			panic(lua.RuntimeError("1番目の引数がTextSpriteではありません。"))
		}
		fnt := l.ToUserData(2)
		switch fnt.(type) {
		case *Fnt:
		default:
			panic(lua.RuntimeError("2番目の引数がFntではありません。"))
		}
		ud.(*TextSprite).fnt = fnt.(*Fnt)
		return 0
	})
	luaRegister(l, "textImgSetBank", func(*lua.State) int {
		ud := l.ToUserData(1)
		switch ud.(type) {
		case *TextSprite:
		default:
			panic(lua.RuntimeError("1番目の引数がTextSpriteではありません。"))
		}
		ud.(*TextSprite).bank = int32(numArg(l, 2))
		return 0
	})
	luaRegister(l, "textImgSetAlign", func(*lua.State) int {
		ud := l.ToUserData(1)
		switch ud.(type) {
		case *TextSprite:
		default:
			panic(lua.RuntimeError("1番目の引数がTextSpriteではありません。"))
		}
		ud.(*TextSprite).align = int32(numArg(l, 2))
		return 0
	})
	luaRegister(l, "textImgSetText", func(*lua.State) int {
		ud := l.ToUserData(1)
		switch ud.(type) {
		case *TextSprite:
		default:
			panic(lua.RuntimeError("1番目の引数がTextSpriteではありません。"))
		}
		ud.(*TextSprite).text = strArg(l, 2)
		return 0
	})
	luaRegister(l, "textImgSetPos", func(*lua.State) int {
		ud := l.ToUserData(1)
		switch ud.(type) {
		case *TextSprite:
		default:
			panic(lua.RuntimeError("1番目の引数がTextSpriteではありません。"))
		}
		ud.(*TextSprite).x = float32(numArg(l, 2))
		ud.(*TextSprite).y = float32(numArg(l, 3))
		return 0
	})
	luaRegister(l, "textImgSetScale", func(*lua.State) int {
		ud := l.ToUserData(1)
		switch ud.(type) {
		case *TextSprite:
		default:
			panic(lua.RuntimeError("1番目の引数がTextSpriteではありません。"))
		}
		ud.(*TextSprite).xscl = float32(numArg(l, 2))
		ud.(*TextSprite).yscl = float32(numArg(l, 3))
		return 0
	})
	luaRegister(l, "textImgDraw", func(*lua.State) int {
		ud := l.ToUserData(1)
		switch ud.(type) {
		case *TextSprite:
		default:
			panic(lua.RuntimeError("1番目の引数がTextSpriteではありません。"))
		}
		ud.(*TextSprite).Draw()
		return 0
	})
	luaRegister(l, "animNew", func(*lua.State) int {
		ud := l.ToUserData(1)
		switch ud.(type) {
		case *Sff:
		default:
			panic(lua.RuntimeError("1番目の引数がSffではありません。"))
		}
		act := strArg(l, 2)
		anim := NewAnim(ud.(*Sff), act)
		if anim == nil {
			panic(lua.RuntimeError(fmt.Sprintf(
				"\n%s\n\nデータの読み込みに失敗しました。", act)))
		}
		l.PushUserData(anim)
		return 1
	})
	luaRegister(l, "animSetPos", func(*lua.State) int {
		ud := l.ToUserData(1)
		switch ud.(type) {
		case *Anim:
		default:
			panic(lua.RuntimeError("1番目の引数がAnimではありません。"))
		}
		ud.(*Anim).SetPos(float32(numArg(l, 2)), float32(numArg(l, 3)))
		return 0
	})
	luaRegister(l, "animAddPos", func(*lua.State) int {
		ud := l.ToUserData(1)
		switch ud.(type) {
		case *Anim:
		default:
			panic(lua.RuntimeError("1番目の引数がAnimではありません。"))
		}
		ud.(*Anim).AddPos(float32(numArg(l, 2)), float32(numArg(l, 3)))
		return 0
	})
	luaRegister(l, "animSetTile", func(*lua.State) int {
		ud := l.ToUserData(1)
		switch ud.(type) {
		case *Anim:
		default:
			panic(lua.RuntimeError("1番目の引数がAnimではありません。"))
		}
		ud.(*Anim).SetTile(int32(numArg(l, 2)), int32(numArg(l, 3)))
		return 0
	})
	luaRegister(l, "animSetColorKey", func(*lua.State) int {
		ud := l.ToUserData(1)
		switch ud.(type) {
		case *Anim:
		default:
			panic(lua.RuntimeError("1番目の引数がAnimではありません。"))
		}
		ud.(*Anim).SetColorKey(int16(numArg(l, 2)))
		return 0
	})
	luaRegister(l, "animSetAlpha", func(*lua.State) int {
		ud := l.ToUserData(1)
		switch ud.(type) {
		case *Anim:
		default:
			panic(lua.RuntimeError("1番目の引数がAnimではありません。"))
		}
		ud.(*Anim).SetAlpha(int16(numArg(l, 2)), int16(numArg(l, 3)))
		return 0
	})
	luaRegister(l, "animSetScale", func(*lua.State) int {
		ud := l.ToUserData(1)
		switch ud.(type) {
		case *Anim:
		default:
			panic(lua.RuntimeError("1番目の引数がAnimではありません。"))
		}
		ud.(*Anim).SetScale(float32(numArg(l, 2)), float32(numArg(l, 3)))
		return 0
	})
	luaRegister(l, "animSetWindow", func(*lua.State) int {
		ud := l.ToUserData(1)
		switch ud.(type) {
		case *Anim:
		default:
			panic(lua.RuntimeError("1番目の引数がAnimではありません。"))
		}
		ud.(*Anim).SetWindow(float32(numArg(l, 2)), float32(numArg(l, 3)),
			float32(numArg(l, 4)), float32(numArg(l, 5)))
		return 0
	})
	luaRegister(l, "animUpdate", func(*lua.State) int {
		ud := l.ToUserData(1)
		switch ud.(type) {
		case *Anim:
		default:
			panic(lua.RuntimeError("1番目の引数がAnimではありません。"))
		}
		ud.(*Anim).Update()
		return 0
	})
	luaRegister(l, "animDraw", func(*lua.State) int {
		ud := l.ToUserData(1)
		switch ud.(type) {
		case *Anim:
		default:
			panic(lua.RuntimeError("1番目の引数がAnimではありません。"))
		}
		ud.(*Anim).Draw()
		return 0
	})
	luaRegister(l, "refresh", func(*lua.State) int {
		await(60)
		if gameEnd {
			panic(lua.RuntimeError("<game end>"))
		}
		return 0
	})
}
