package main

import (
	"github.com/yuin/gopher-lua"
	"strings"
)

func luaRegister(l *lua.LState, name string, f func(*lua.LState) int) {
	l.Register(name, f)
}
func strArg(l *lua.LState, argi int) string {
	if !lua.LVCanConvToString(l.Get(argi)) {
		l.RaiseError("%v番目の引数が文字列ではありません。", argi)
	}
	return l.ToString(argi)
}
func numArg(l *lua.LState, argi int) float64 {
	num, ok := l.Get(argi).(lua.LNumber)
	if !ok {
		l.RaiseError("%v番目の引数が数ではありません。", argi)
	}
	return float64(num)
}
func boolArg(l *lua.LState, argi int) bool {
	return l.ToBool(argi)
}
func newUserData(l *lua.LState, value interface{}) *lua.LUserData {
	ud := l.NewUserData()
	ud.Value = value
	return ud
}
func toUserData(l *lua.LState, argi int) interface{} {
	if ud := l.ToUserData(argi); ud != nil {
		return ud.Value
	}
	return nil
}
func userDataError(l *lua.LState, argi int, udtype interface{}) {
	l.RaiseError("%v番目の引数が%Tではありません。", argi, udtype)
}

type InputDialog struct{}

func newInputDialog() *InputDialog {
	return &InputDialog{}
}

// Script Common

func scriptCommonInit(l *lua.LState) {
	luaRegister(l, "sffNew", func(l *lua.LState) int {
		sff, err := LoadSff(strArg(l, 1), false)
		if err != nil {
			l.RaiseError(err.Error())
		}
		l.Push(newUserData(l, sff))
		return 1
	})
	luaRegister(l, "sndNew", func(l *lua.LState) int {
		snd, err := LoadSnd(strArg(l, 1))
		if err != nil {
			l.RaiseError(err.Error())
		}
		l.Push(newUserData(l, snd))
		return 1
	})
	luaRegister(l, "fontNew", func(l *lua.LState) int {
		fnt, err := LoadFnt(strArg(l, 1))
		if err != nil {
			l.RaiseError(err.Error())
		}
		l.Push(newUserData(l, fnt))
		return 1
	})
	luaRegister(l, "commandNew", func(l *lua.LState) int {
		l.Push(newUserData(l, NewCommandList(&CommandBuffer{})))
		return 1
	})
	luaRegister(l, "commandAdd", func(l *lua.LState) int {
		cl, ok := toUserData(l, 1).(*CommandList)
		if !ok {
			userDataError(l, 1, cl)
		}
		c, err := ReadCommand(strArg(l, 2), strArg(l, 3), NewCommandKeyRemap())
		if err != nil {
			l.RaiseError(err.Error())
		}
		cl.Add(*c)
		return 0
	})
	luaRegister(l, "commandGetState", func(l *lua.LState) int {
		cl, ok := toUserData(l, 1).(*CommandList)
		if !ok {
			userDataError(l, 1, cl)
		}
		l.Push(lua.LBool(cl.GetState(strArg(l, 2))))
		return 1
	})
	luaRegister(l, "commandInput", func(l *lua.LState) int {
		cl, ok := toUserData(l, 1).(*CommandList)
		if !ok {
			userDataError(l, 1, cl)
		}
		if cl.Input(int(numArg(l, 2))-1, 1) {
			cl.Step(1, false, false, 0)
		}
		return 0
	})
	luaRegister(l, "commandBufReset", func(l *lua.LState) int {
		cl, ok := toUserData(l, 1).(*CommandList)
		if !ok {
			userDataError(l, 1, cl)
		}
		cl.BufReset()
		return 0
	})
	luaRegister(l, "inputDialogNew", func(l *lua.LState) int {
		l.Push(newUserData(l, newInputDialog()))
		return 1
	})
	luaRegister(l, "sndPlay", func(l *lua.LState) int {
		s, ok := toUserData(l, 1).(*Snd)
		if !ok {
			userDataError(l, 1, s)
		}
		s.Play(int32(numArg(l, 2)), int32(numArg(l, 3)))
		return 0
	})
	luaRegister(l, "playBGM", func(l *lua.LState) int {
		sys.bgm.Open(strArg(l, 1))
		return 0
	})
	luaRegister(l, "esc", func(l *lua.LState) int {
		l.Push(lua.LBool(sys.esc))
		return 1
	})
	luaRegister(l, "setRoundTime", func(l *lua.LState) int {
		sys.roundTime = int32(numArg(l, 1))
		return 0
	})
	luaRegister(l, "setHomeTeam", func(l *lua.LState) int {
		tn := int32(numArg(l, 1))
		if tn < 1 || tn > 2 {
			l.RaiseError("チーム番号(%v)が不正です。", tn)
		}
		sys.home = tn - 1
		return 0
	})
	luaRegister(l, "setMatchNo", func(l *lua.LState) int {
		sys.match = int32(numArg(l, 1))
		return 0
	})
}

// System Script

func systemScriptInit(l *lua.LState) {
	scriptCommonInit(l)
	luaRegister(l, "textImgNew", func(*lua.LState) int {
		l.Push(newUserData(l, NewTextSprite()))
		return 1
	})
	luaRegister(l, "textImgSetFont", func(*lua.LState) int {
		ts, ok := toUserData(l, 1).(*TextSprite)
		if !ok {
			userDataError(l, 1, ts)
		}
		fnt, ok2 := toUserData(l, 2).(*Fnt)
		if !ok2 {
			userDataError(l, 2, fnt)
		}
		ts.fnt = fnt
		return 0
	})
	luaRegister(l, "textImgSetBank", func(*lua.LState) int {
		ts, ok := toUserData(l, 1).(*TextSprite)
		if !ok {
			userDataError(l, 1, ts)
		}
		ts.bank = int32(numArg(l, 2))
		return 0
	})
	luaRegister(l, "textImgSetAlign", func(*lua.LState) int {
		ts, ok := toUserData(l, 1).(*TextSprite)
		if !ok {
			userDataError(l, 1, ts)
		}
		ts.align = int32(numArg(l, 2))
		return 0
	})
	luaRegister(l, "textImgSetText", func(*lua.LState) int {
		ts, ok := toUserData(l, 1).(*TextSprite)
		if !ok {
			userDataError(l, 1, ts)
		}
		ts.text = strArg(l, 2)
		return 0
	})
	luaRegister(l, "textImgSetPos", func(*lua.LState) int {
		ts, ok := toUserData(l, 1).(*TextSprite)
		if !ok {
			userDataError(l, 1, ts)
		}
		ts.x, ts.y = float32(numArg(l, 2)), float32(numArg(l, 3))
		return 0
	})
	luaRegister(l, "textImgSetScale", func(*lua.LState) int {
		ts, ok := toUserData(l, 1).(*TextSprite)
		if !ok {
			userDataError(l, 1, ts)
		}
		ts.xscl, ts.yscl = float32(numArg(l, 2)), float32(numArg(l, 3))
		return 0
	})
	luaRegister(l, "textImgDraw", func(*lua.LState) int {
		ts, ok := toUserData(l, 1).(*TextSprite)
		if !ok {
			userDataError(l, 1, ts)
		}
		ts.Draw()
		return 0
	})
	luaRegister(l, "animNew", func(*lua.LState) int {
		s, ok := toUserData(l, 1).(*Sff)
		if !ok {
			userDataError(l, 1, s)
		}
		act := strArg(l, 2)
		anim := NewAnim(s, act)
		if anim == nil {
			l.RaiseError("\n%s\n\nデータの読み込みに失敗しました。", act)
		}
		l.Push(newUserData(l, anim))
		return 1
	})
	luaRegister(l, "animSetPos", func(*lua.LState) int {
		a, ok := toUserData(l, 1).(*Anim)
		if !ok {
			userDataError(l, 1, a)
		}
		a.SetPos(float32(numArg(l, 2)), float32(numArg(l, 3)))
		return 0
	})
	luaRegister(l, "animAddPos", func(*lua.LState) int {
		a, ok := toUserData(l, 1).(*Anim)
		if !ok {
			userDataError(l, 1, a)
		}
		a.AddPos(float32(numArg(l, 2)), float32(numArg(l, 3)))
		return 0
	})
	luaRegister(l, "animSetTile", func(*lua.LState) int {
		a, ok := toUserData(l, 1).(*Anim)
		if !ok {
			userDataError(l, 1, a)
		}
		a.SetTile(int32(numArg(l, 2)), int32(numArg(l, 3)))
		return 0
	})
	luaRegister(l, "animSetColorKey", func(*lua.LState) int {
		a, ok := toUserData(l, 1).(*Anim)
		if !ok {
			userDataError(l, 1, a)
		}
		a.SetColorKey(int16(numArg(l, 2)))
		return 0
	})
	luaRegister(l, "animSetAlpha", func(*lua.LState) int {
		a, ok := toUserData(l, 1).(*Anim)
		if !ok {
			userDataError(l, 1, a)
		}
		a.SetAlpha(int16(numArg(l, 2)), int16(numArg(l, 3)))
		return 0
	})
	luaRegister(l, "animSetScale", func(*lua.LState) int {
		a, ok := toUserData(l, 1).(*Anim)
		if !ok {
			userDataError(l, 1, a)
		}
		a.SetScale(float32(numArg(l, 2)), float32(numArg(l, 3)))
		return 0
	})
	luaRegister(l, "animSetWindow", func(*lua.LState) int {
		a, ok := toUserData(l, 1).(*Anim)
		if !ok {
			userDataError(l, 1, a)
		}
		a.SetWindow(float32(numArg(l, 2)), float32(numArg(l, 3)),
			float32(numArg(l, 4)), float32(numArg(l, 5)))
		return 0
	})
	luaRegister(l, "animUpdate", func(*lua.LState) int {
		a, ok := toUserData(l, 1).(*Anim)
		if !ok {
			userDataError(l, 1, a)
		}
		a.Update()
		return 0
	})
	luaRegister(l, "animDraw", func(*lua.LState) int {
		a, ok := toUserData(l, 1).(*Anim)
		if !ok {
			userDataError(l, 1, a)
		}
		a.Draw()
		return 0
	})
	luaRegister(l, "exitNetPlay", func(*lua.LState) int {
		if sys.netInput != nil {
			sys.netInput.Close()
			sys.netInput = nil
		}
		return 0
	})
	luaRegister(l, "exitReplay", func(*lua.LState) int {
		if sys.fileInput != nil {
			sys.fileInput.Close()
			sys.fileInput = nil
		}
		return 0
	})
	luaRegister(l, "setCom", func(*lua.LState) int {
		pn := int(numArg(l, 1))
		if pn < 1 || pn > MaxSimul*2 {
			l.RaiseError("プレイヤー番号(%v)が不正です。", pn)
		}
		sys.com[pn-1] = Max(0, int32(numArg(l, 2)))
		return 0
	})
	luaRegister(l, "setAutoLevel", func(*lua.LState) int {
		sys.autolevel = boolArg(l, 1)
		return 0
	})
	luaRegister(l, "getListenPort", func(*lua.LState) int {
		l.Push(lua.LString(sys.listenPort))
		return 1
	})
	luaRegister(l, "addChar", func(l *lua.LState) int {
		for _, c := range strings.Split(strings.TrimSpace(strArg(l, 1)), "\n") {
			if len(c) > 0 {
				sys.sel.AddCahr(c)
			}
		}
		return 0
	})
	luaRegister(l, "addStage", func(l *lua.LState) int {
		for _, c := range SplitAndTrim(strings.TrimSpace(strArg(l, 1)), "\n") {
			if err := sys.sel.AddStage(c); err != nil {
				l.RaiseError(err.Error())
			}
		}
		return 0
	})
	luaRegister(l, "setRandomSpr", func(*lua.LState) int {
		sff, ok := toUserData(l, 1).(*Sff)
		if !ok {
			userDataError(l, 1, sff)
		}
		sys.sel.randomspr = sff.GetOwnPalSprite(int16(numArg(l, 2)),
			int16(numArg(l, 3)))
		sys.sel.randomscl = [2]float32{float32(numArg(l, 4)),
			float32(numArg(l, 5))}
		return 0
	})
	luaRegister(l, "setSelColRow", func(*lua.LState) int {
		sys.sel.columns = int(numArg(l, 1))
		sys.sel.rows = int(numArg(l, 2))
		return 0
	})
	luaRegister(l, "setSelCellSize", func(*lua.LState) int {
		sys.sel.cellsize = [2]float32{float32(numArg(l, 1)), float32(numArg(l, 2))}
		return 0
	})
	luaRegister(l, "setSelCellScale", func(*lua.LState) int {
		sys.sel.cellscale = [2]float32{float32(numArg(l, 1)),
			float32(numArg(l, 2))}
		return 0
	})
	luaRegister(l, "setStage", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.sel.SetStageNo(int(numArg(l, 1)))))
		return 1
	})
	luaRegister(l, "selectStage", func(*lua.LState) int {
		sys.sel.SelectStage(int(numArg(l, 1)))
		return 0
	})
	luaRegister(l, "setTeamMode", func(*lua.LState) int {
		tn := int(numArg(l, 1))
		if tn < 1 || tn > 2 {
			l.RaiseError("チーム番号(%v)が不正です。", tn)
		}
		tm := TeamMode(numArg(l, 2))
		if tm < 0 || tm > TM_LAST {
			l.RaiseError("モード番号(%v)が不正です。", tm)
		}
		nt := int(numArg(l, 3))
		if nt < 1 || nt > MaxSimul {
			l.RaiseError("チーム人数(%v)が不正です。", nt)
		}
		sys.sel.selected[tn-1], sys.tmode[tn-1] = nil, tm
		sys.numTurns[tn-1], sys.numSimul[tn-1] = nt, nt
		if tm == TM_Simul && nt == 1 {
			sys.tmode[tn-1] = TM_Single
		}
		return 0
	})
	luaRegister(l, "getCharName", func(*lua.LState) int {
		c := sys.sel.GetChar(int(numArg(l, 1)))
		l.Push(lua.LString(c.name))
		return 1
	})
	luaRegister(l, "selectChar", func(*lua.LState) int {
		tn := int(numArg(l, 1))
		if tn < 1 || tn > 2 {
			l.RaiseError("チーム番号(%v)が不正です。", tn)
		}
		cn, pl, ret := int(numArg(l, 2)), int(numArg(l, 3)), 0
		if pl >= 1 && pl <= 12 && sys.sel.AddSelectedChar(tn-1, cn, pl) {
			switch sys.tmode[tn-1] {
			case TM_Single:
				ret = 2
			case TM_Simul:
				if len(sys.sel.selected[tn-1]) >= sys.numSimul[tn-1] {
					ret = 2
				} else {
					ret = 1
				}
			case TM_Turns:
				if len(sys.sel.selected[tn-1]) >= sys.numTurns[tn-1] {
					ret = 2
				} else {
					ret = 1
				}
			}
		}
		l.Push(lua.LNumber(ret))
		return 1
	})
	luaRegister(l, "getStageName", func(*lua.LState) int {
		l.Push(lua.LString(sys.sel.GetStageName(int(numArg(l, 1)))))
		return 1
	})
	luaRegister(l, "refresh", func(*lua.LState) int {
		sys.await(60)
		if sys.gameEnd {
			l.RaiseError("<game end>")
		}
		return 0
	})
	luaRegister(l, "drawPortrait", func(l *lua.LState) int {
		n, x, y := int(numArg(l, 1)), float32(numArg(l, 2)), float32(numArg(l, 3))
		var xscl, yscl float32 = 1, 1
		if l.GetTop() >= 4 {
			xscl = float32(numArg(l, 4))
			if l.GetTop() >= 5 {
				yscl = float32(numArg(l, 5))
			}
		}
		if !sys.frameSkip {
			c := sys.sel.GetChar(n)
			if c != nil && c.lportrait != nil {
				c.lportrait.Draw(x, y, xscl, yscl, c.lportrait.Pal)
			}
		}
		return 0
	})
	luaRegister(l, "drawFace", func(l *lua.LState) int {
		x, y := float32(numArg(l, 1)), float32(numArg(l, 2))
		offset := 0
		if l.GetTop() >= 3 {
			offset = int(numArg(l, 3))
		}
		if !sys.frameSkip {
			for j := 0; j < sys.sel.rows; j++ {
				for i := 0; i < sys.sel.columns; i++ {
					c := sys.sel.GetChar(offset)
					offset++
					if c != nil {
						if c.sportrait != nil {
							c.sportrait.Draw(x+float32(i)*sys.sel.cellsize[0],
								y+float32(j)*sys.sel.cellsize[1], sys.sel.cellscale[0],
								sys.sel.cellscale[1], c.sportrait.Pal)
						} else if c.def == "randomselect" && sys.sel.randomspr != nil {
							sys.sel.randomspr.Draw(x+float32(i)*sys.sel.cellsize[0],
								y+float32(j)*sys.sel.cellsize[1], sys.sel.randomscl[0],
								sys.sel.randomscl[1], sys.sel.randomspr.Pal)
						}
					}
				}
			}
		}
		return 0
	})
	luaRegister(l, "loadLifebar", func(l *lua.LState) int {
		lb, err := LoadLifebar(strArg(l, 1))
		if err != nil {
			l.RaiseError(err.Error())
		}
		sys.lifebar = *lb
		return 0
	})
	luaRegister(l, "loadDebugFont", func(l *lua.LState) int {
		f, err := LoadFnt(strArg(l, 1))
		if err != nil {
			l.RaiseError(err.Error())
		}
		sys.debugFont = f
		return 0
	})
	luaRegister(l, "setDebugScript", func(l *lua.LState) int {
		sys.debugScript = strArg(l, 1)
		return 0
	})
	luaRegister(l, "setLifeMul", func(l *lua.LState) int {
		sys.lifeMul = float32(numArg(l, 1))
		return 0
	})
	luaRegister(l, "setTeam1VS2Life", func(l *lua.LState) int {
		sys.team1VS2Life = float32(numArg(l, 1))
		return 0
	})
	luaRegister(l, "setTurnsRecoveryRate", func(l *lua.LState) int {
		sys.turnsRecoveryRate = float32(numArg(l, 1))
		return 0
	})
	luaRegister(l, "setZoom", func(l *lua.LState) int {
		sys.zoomEnable = boolArg(l, 1)
		return 0
	})
	luaRegister(l, "setZoomMin", func(l *lua.LState) int {
		sys.zoomMin = float32(numArg(l, 1))
		return 0
	})
	luaRegister(l, "setZoomMax", func(l *lua.LState) int {
		sys.zoomMax = float32(numArg(l, 1))
		return 0
	})
	luaRegister(l, "setZoomSpeed", func(l *lua.LState) int {
		sys.zoomSpeed = float32(numArg(l, 1))
		return 0
	})
	luaRegister(l, "resetRemapInput", func(l *lua.LState) int {
		sys.resetRemapInput()
		return 0
	})
	luaRegister(l, "loadStart", func(l *lua.LState) int {
		sys.loadStart()
		return 0
	})
	luaRegister(l, "selectStart", func(l *lua.LState) int {
		sys.sel.ClearSelected()
		sys.loadStart()
		return 0
	})
	luaRegister(l, "game", func(l *lua.LState) int {
		if sys.gameEnd {
			l.Push(lua.LNumber(-1))
			return 1
		}
		winp := int32(0)
		sys.roundsExisted = [2]int32{0, 0}
		mw := [2]int32{0, 0}
		for i := range sys.lifebar.wi {
			sys.lifebar.wi[i].clear()
		}
		sys.draws = 0
		load := func() error {
			sys.loader.runTread()
			for sys.loader.state != LS_Complete {
				if sys.loader.state == LS_Error {
					return sys.loader.err
				}
				sys.await(60)
			}
			return nil
		}
		fight := func() (int32, error) {
			if err := load(); err != nil {
				return -1, err
			}
			unimplemented()
			if sys.round == 1 {
				if sys.tmode[1] == TM_Turns {
					mw[0] = int32(sys.numTurns[1])
				} else {
					mw[0] = sys.lifebar.ro.match_wins
				}
				if sys.tmode[0] == TM_Turns {
					mw[1] = int32(sys.numTurns[0])
				} else {
					mw[1] = sys.lifebar.ro.match_wins
				}
			}
			unimplemented()
			return winp, nil
		}
		if sys.netInput != nil {
			sys.netInput.Stop()
		}
		defer sys.synchronize()
		for {
			var err error
			if winp, err = fight(); err != nil {
				l.RaiseError(err.Error())
			}
			if winp < 0 || sys.tmode[0] != TM_Turns && sys.tmode[1] != TM_Turns ||
				sys.wins[0] >= mw[0] || sys.wins[1] >= mw[1] || sys.gameEnd {
				break
			}
			unimplemented()
		}
		l.Push(lua.LNumber(winp))
		return 1
	})
}
