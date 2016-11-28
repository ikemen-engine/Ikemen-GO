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
		lua.Errorf(l, "%d番目の引数が文字列ではありません。", argi)
	}
	return str
}
func numArg(l *lua.State, argi int) float64 {
	num, ok := l.ToNumber(argi)
	if !ok {
		lua.Errorf(l, "%d番目の引数が数ではありません。", argi)
	}
	return num
}
func boolArg(l *lua.State, argi int) bool {
	if !l.IsBoolean(argi) {
		lua.Errorf(l, "%d番目の引数が論理値ではありません。", argi)
	}
	return l.ToBoolean(argi)
}
func userDataError(l *lua.State, argi int, udtype interface{}) {
	lua.Errorf(l, fmt.Sprintf("%d番目の引数が%Tではありません。", argi, udtype))
}
func scriptCommonInit(l *lua.State) {
	luaRegister(l, "sffNew", func(l *lua.State) int {
		sff, err := LoadSff(strArg(l, 1), false)
		if err != nil {
			lua.Errorf(l, err.Error())
		}
		l.PushUserData(sff)
		return 1
	})
	luaRegister(l, "sndNew", func(l *lua.State) int {
		snd, err := LoadSnd(strArg(l, 1))
		if err != nil {
			lua.Errorf(l, err.Error())
		}
		l.PushUserData(snd)
		return 1
	})
	luaRegister(l, "fontNew", func(l *lua.State) int {
		fnt, err := LoadFnt(strArg(l, 1))
		if err != nil {
			lua.Errorf(l, err.Error())
		}
		l.PushUserData(fnt)
		return 1
	})
	luaRegister(l, "sndPlay", func(l *lua.State) int {
		s, ok := l.ToUserData(1).(*Snd)
		if !ok {
			userDataError(l, 1, s)
		}
		s.Play(int32(numArg(l, 2)), int32(numArg(l, 3)))
		return 0
	})
	luaRegister(l, "playBGM", func(l *lua.State) int {
		bgm.Open(strArg(l, 1))
		return 0
	})
	luaRegister(l, "setRoundTime", func(l *lua.State) int {
		roundTime = int32(numArg(l, 1))
		return 0
	})
}

// System Script

func systemScriptInit(l *lua.State) {
	scriptCommonInit(l)
	luaRegister(l, "textImgNew", func(*lua.State) int {
		l.PushUserData(NewTextSprite())
		return 1
	})
	luaRegister(l, "textImgSetFont", func(*lua.State) int {
		ts, ok := l.ToUserData(1).(*TextSprite)
		if !ok {
			userDataError(l, 1, ts)
		}
		fnt, ok2 := l.ToUserData(2).(*Fnt)
		if !ok2 {
			userDataError(l, 2, fnt)
		}
		ts.fnt = fnt
		return 0
	})
	luaRegister(l, "textImgSetBank", func(*lua.State) int {
		ts, ok := l.ToUserData(1).(*TextSprite)
		if !ok {
			userDataError(l, 1, ts)
		}
		ts.bank = int32(numArg(l, 2))
		return 0
	})
	luaRegister(l, "textImgSetAlign", func(*lua.State) int {
		ts, ok := l.ToUserData(1).(*TextSprite)
		if !ok {
			userDataError(l, 1, ts)
		}
		ts.align = int32(numArg(l, 2))
		return 0
	})
	luaRegister(l, "textImgSetText", func(*lua.State) int {
		ts, ok := l.ToUserData(1).(*TextSprite)
		if !ok {
			userDataError(l, 1, ts)
		}
		ts.text = strArg(l, 2)
		return 0
	})
	luaRegister(l, "textImgSetPos", func(*lua.State) int {
		ts, ok := l.ToUserData(1).(*TextSprite)
		if !ok {
			userDataError(l, 1, ts)
		}
		ts.x, ts.y = float32(numArg(l, 2)), float32(numArg(l, 3))
		return 0
	})
	luaRegister(l, "textImgSetScale", func(*lua.State) int {
		ts, ok := l.ToUserData(1).(*TextSprite)
		if !ok {
			userDataError(l, 1, ts)
		}
		ts.xscl, ts.yscl = float32(numArg(l, 2)), float32(numArg(l, 3))
		return 0
	})
	luaRegister(l, "textImgDraw", func(*lua.State) int {
		ts, ok := l.ToUserData(1).(*TextSprite)
		if !ok {
			userDataError(l, 1, ts)
		}
		ts.Draw()
		return 0
	})
	luaRegister(l, "animNew", func(*lua.State) int {
		s, ok := l.ToUserData(1).(*Sff)
		if !ok {
			userDataError(l, 1, s)
		}
		act := strArg(l, 2)
		anim := NewAnim(s, act)
		if anim == nil {
			lua.Errorf(l, "\n%s\n\nデータの読み込みに失敗しました。", act)
		}
		l.PushUserData(anim)
		return 1
	})
	luaRegister(l, "animSetPos", func(*lua.State) int {
		a, ok := l.ToUserData(1).(*Anim)
		if !ok {
			userDataError(l, 1, a)
		}
		a.SetPos(float32(numArg(l, 2)), float32(numArg(l, 3)))
		return 0
	})
	luaRegister(l, "animAddPos", func(*lua.State) int {
		a, ok := l.ToUserData(1).(*Anim)
		if !ok {
			userDataError(l, 1, a)
		}
		a.AddPos(float32(numArg(l, 2)), float32(numArg(l, 3)))
		return 0
	})
	luaRegister(l, "animSetTile", func(*lua.State) int {
		a, ok := l.ToUserData(1).(*Anim)
		if !ok {
			userDataError(l, 1, a)
		}
		a.SetTile(int32(numArg(l, 2)), int32(numArg(l, 3)))
		return 0
	})
	luaRegister(l, "animSetColorKey", func(*lua.State) int {
		a, ok := l.ToUserData(1).(*Anim)
		if !ok {
			userDataError(l, 1, a)
		}
		a.SetColorKey(int16(numArg(l, 2)))
		return 0
	})
	luaRegister(l, "animSetAlpha", func(*lua.State) int {
		a, ok := l.ToUserData(1).(*Anim)
		if !ok {
			userDataError(l, 1, a)
		}
		a.SetAlpha(int16(numArg(l, 2)), int16(numArg(l, 3)))
		return 0
	})
	luaRegister(l, "animSetScale", func(*lua.State) int {
		a, ok := l.ToUserData(1).(*Anim)
		if !ok {
			userDataError(l, 1, a)
		}
		a.SetScale(float32(numArg(l, 2)), float32(numArg(l, 3)))
		return 0
	})
	luaRegister(l, "animSetWindow", func(*lua.State) int {
		a, ok := l.ToUserData(1).(*Anim)
		if !ok {
			userDataError(l, 1, a)
		}
		a.SetWindow(float32(numArg(l, 2)), float32(numArg(l, 3)),
			float32(numArg(l, 4)), float32(numArg(l, 5)))
		return 0
	})
	luaRegister(l, "animUpdate", func(*lua.State) int {
		a, ok := l.ToUserData(1).(*Anim)
		if !ok {
			userDataError(l, 1, a)
		}
		a.Update()
		return 0
	})
	luaRegister(l, "animDraw", func(*lua.State) int {
		a, ok := l.ToUserData(1).(*Anim)
		if !ok {
			userDataError(l, 1, a)
		}
		a.Draw()
		return 0
	})
	luaRegister(l, "refresh", func(*lua.State) int {
		await(60)
		if gameEnd {
			lua.Errorf(l, "<game end>")
		}
		return 0
	})
	luaRegister(l, "loadLifebar", func(l *lua.State) int {
		lb, err := LoadLifebar(strArg(l, 1))
		if err != nil {
			lua.Errorf(l, err.Error())
		}
		lifebar = *lb
		return 0
	})
	luaRegister(l, "loadDebugFont", func(l *lua.State) int {
		f, err := LoadFnt(strArg(l, 1))
		if err != nil {
			lua.Errorf(l, err.Error())
		}
		debugFont = f
		return 0
	})
	luaRegister(l, "setDebugScript", func(l *lua.State) int {
		debugScript = strArg(l, 1)
		return 0
	})
	luaRegister(l, "setLifeMul", func(l *lua.State) int {
		lifeMul = float32(numArg(l, 1))
		return 0
	})
	luaRegister(l, "setTeam1VS2Life", func(l *lua.State) int {
		team1VS2Life = float32(numArg(l, 1))
		return 0
	})
	luaRegister(l, "setTurnsRecoveryRate", func(l *lua.State) int {
		turnsRecoveryRate = float32(numArg(l, 1))
		return 0
	})
	luaRegister(l, "setZoom", func(l *lua.State) int {
		zoomEnable = boolArg(l, 1)
		return 0
	})
	luaRegister(l, "setZoomMin", func(l *lua.State) int {
		zoomMin = float32(numArg(l, 1))
		return 0
	})
	luaRegister(l, "setZoomMax", func(l *lua.State) int {
		zoomMax = float32(numArg(l, 1))
		return 0
	})
	luaRegister(l, "setZoomSpeed", func(l *lua.State) int {
		zoomSpeed = float32(numArg(l, 1))
		return 0
	})
}
