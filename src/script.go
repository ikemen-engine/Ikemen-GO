package main

import (
	"fmt"
	"runtime"
	"strings"

	"math/rand"

	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/yuin/gopher-lua"
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

type InputDialog interface {
	Popup(title string) (ok bool)
	IsDone() bool
	GetStr() string
}

func newInputDialog() InputDialog {
	return newCommandLineInput()
}

type commandLineInput struct {
	str  string
	done bool
}

func newCommandLineInput() *commandLineInput {
	return &commandLineInput{done: true}
}
func (cli *commandLineInput) Popup(title string) bool {
	if !cli.done {
		return false
	}
	cli.done = false
	print(title + ": ")
	return true
}
func (cli *commandLineInput) IsDone() bool {
	if !cli.done {
		select {
		case cli.str = <-sys.commandLine:
			cli.done = true
		default:
		}
	}
	return cli.done
}
func (cli *commandLineInput) GetStr() string {
	if !cli.IsDone() {
		return ""
	}
	return cli.str
}

// Script Common

func scriptCommonInit(l *lua.LState) {
	luaRegister(l, "sffNew", func(l *lua.LState) int {
		sff, err := loadSff(strArg(l, 1), false)
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
		fnt, err := loadFnt(strArg(l, 1))
		if err != nil {
			l.RaiseError(err.Error())
		}
		l.Push(newUserData(l, fnt))
		return 1
	})
	luaRegister(l, "commandNew", func(l *lua.LState) int {
		l.Push(newUserData(l, NewCommandList(NewCommandBuffer())))
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
	luaRegister(l, "inputDialogPopup", func(l *lua.LState) int {
		id, ok := toUserData(l, 1).(InputDialog)
		if !ok {
			userDataError(l, 1, id)
		}
		id.Popup(strArg(l, 2))
		return 0
	})
	luaRegister(l, "inputDialogIsDone", func(l *lua.LState) int {
		id, ok := toUserData(l, 1).(InputDialog)
		if !ok {
			userDataError(l, 1, id)
		}
		l.Push(lua.LBool(id.IsDone()))
		return 1
	})
	luaRegister(l, "inputDialogGetStr", func(l *lua.LState) int {
		id, ok := toUserData(l, 1).(InputDialog)
		if !ok {
			userDataError(l, 1, id)
		}
		l.Push(lua.LString(id.GetStr()))
		return 1
	})
	luaRegister(l, "sndPlay", func(l *lua.LState) int {
		s, ok := toUserData(l, 1).(*Snd)
		if !ok {
			userDataError(l, 1, s)
		}
		s.play([...]int32{int32(numArg(l, 2)), int32(numArg(l, 3))})
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
	luaRegister(l, "sszRandom", func(l *lua.LState) int {
		l.Push(lua.LNumber(Random()))
		return 1
	})
	luaRegister(l, "setAutoguard", func(l *lua.LState) int {
		pn := int(numArg(l, 1))
		if pn < 1 || pn > MaxSimul*2 {
			l.RaiseError("プレイヤー番号(%v)が不正です。", pn)
		}
		sys.autoguard[pn-1] = boolArg(l, 2)
		return 0
	})
	luaRegister(l, "setPowerShare", func(l *lua.LState) int {
		tn := int(numArg(l, 1))
		if tn < 1 || tn > 2 {
			l.RaiseError("チーム番号(%v)が不正です。", tn)
		}
		sys.powerShare[tn-1] = boolArg(l, 2)
		return 0
	})
	luaRegister(l, "setRoundTime", func(l *lua.LState) int {
		sys.roundTime = int32(numArg(l, 1))
		return 0
	})
	luaRegister(l, "getRoundTime", func(l *lua.LState) int {
		l.Push(lua.LNumber(sys.roundTime))
		return 1
	})
	luaRegister(l, "setHomeTeam", func(l *lua.LState) int {
		tn := int(numArg(l, 1))
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
	luaRegister(l, "setLifeShare", func(l *lua.LState) int {
		sys.teamLifeShare = boolArg(l, 1)
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
			l.RaiseError("\n%v\n\nデータの読み込みに失敗しました。", act)
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
		var sx, sy int32 = 0, 0
		if l.GetTop() >= 4 {
			sx = int32(numArg(l, 4))
			if l.GetTop() >= 5 {
				sy = int32(numArg(l, 5))
			} else {
				sy = sx
			}
		}
		a.SetTile(int32(numArg(l, 2)), int32(numArg(l, 3)), sx, sy)
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
	luaRegister(l, "animReset", func(*lua.LState) int {
		a, ok := toUserData(l, 1).(*Anim)
		if !ok {
			userDataError(l, 1, a)
		}
		a.ResetFrames()
		return 0
	})
	luaRegister(l, "enterNetPlay", func(*lua.LState) int {
		if sys.netInput != nil {
			l.RaiseError("すでに通信中です。")
		}
		sys.chars = [len(sys.chars)][]*Char{}
		sys.netInput = NewNetInput("replay/netplay.replay")
		if host := strArg(l, 1); host != "" {
			sys.netInput.Connect(host, sys.listenPort)
		} else {
			if err := sys.netInput.Accept(sys.listenPort); err != nil {
				l.RaiseError(err.Error())
			}
		}
		return 0
	})
	luaRegister(l, "exitNetPlay", func(*lua.LState) int {
		if sys.netInput != nil {
			sys.netInput.Close()
			sys.netInput = nil
		}
		return 0
	})
	luaRegister(l, "enterReplay", func(*lua.LState) int {
		sys.chars = [len(sys.chars)][]*Char{}
		sys.fileInput = OpenFileInput(strArg(l, 1))
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
	luaRegister(l, "connected", func(*lua.LState) int {
		l.Push(lua.LBool(sys.netInput.IsConnected()))
		return 1
	})
	luaRegister(l, "setListenPort", func(*lua.LState) int {
		sys.listenPort = strArg(l, 1)
		return 0
	})
	luaRegister(l, "synchronize", func(*lua.LState) int {
		if err := sys.synchronize(); err != nil {
			l.RaiseError(err.Error())
		}
		return 0
	})
	luaRegister(l, "addChar", func(l *lua.LState) int {
		for _, c := range strings.Split(strings.TrimSpace(strArg(l, 1)), "\n") {
			c = strings.Trim(c, "\r")
			if len(c) > 0 {
				sys.sel.addCahr(c)
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
		sys.sel.randomspr = sff.getOwnPalSprite(int16(numArg(l, 2)),
			int16(numArg(l, 3)))
		sys.sel.randomscl = [...]float32{float32(numArg(l, 4)),
			float32(numArg(l, 5))}
		return 0
	})
	luaRegister(l, "setSelColRow", func(*lua.LState) int {
		sys.sel.columns = int(numArg(l, 1))
		sys.sel.rows = int(numArg(l, 2))
		return 0
	})
	luaRegister(l, "setSelCellSize", func(*lua.LState) int {
		sys.sel.cellsize = [...]float32{float32(numArg(l, 1)),
			float32(numArg(l, 2))}
		return 0
	})
	luaRegister(l, "setSelCellScale", func(*lua.LState) int {
		sys.sel.cellscale = [...]float32{float32(numArg(l, 1)),
			float32(numArg(l, 2))}
		return 0
	})
	luaRegister(l, "numSelCells", func(*lua.LState) int {
		l.Push(lua.LNumber(len(sys.sel.charlist)))
		return 1
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
		nt := int32(numArg(l, 3))
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
	luaRegister(l, "getCharFileName", func(*lua.LState) int {
		c := sys.sel.GetChar(int(numArg(l, 1)))
		l.Push(lua.LString(c.def))
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
				if len(sys.sel.selected[tn-1]) >= int(sys.numSimul[tn-1]) {
					ret = 2
				} else {
					ret = 1
				}
			case TM_Turns:
				if len(sys.sel.selected[tn-1]) >= int(sys.numTurns[tn-1]) {
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
		sys.playSound()
		if !sys.update() {
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
				if c.portrait_scale != 1 {
					xscl *= c.portrait_scale
					yscl *= c.portrait_scale
				}
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
								y+float32(j)*sys.sel.cellsize[1], sys.sel.cellscale[0]*c.portrait_scale,
								sys.sel.cellscale[1]*c.portrait_scale, c.sportrait.Pal)
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
		lb, err := loadLifebar(strArg(l, 1))
		if err != nil {
			l.RaiseError(err.Error())
		}
		sys.lifebar = *lb
		return 0
	})
	luaRegister(l, "loadDebugFont", func(l *lua.LState) int {
		f, err := loadFnt(strArg(l, 1))
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
		sys.cam.ZoomEnable = boolArg(l, 1)
		return 0
	})
	luaRegister(l, "setZoomMin", func(l *lua.LState) int {
		sys.cam.ZoomMin = float32(numArg(l, 1))
		return 0
	})
	luaRegister(l, "setZoomMax", func(l *lua.LState) int {
		sys.cam.ZoomMax = float32(numArg(l, 1))
		return 0
	})
	luaRegister(l, "setZoomSpeed", func(l *lua.LState) int {
		sys.cam.ZoomSpeed = 12 - float32(numArg(l, 1))
		return 0
	})
	luaRegister(l, "resetRemapInput", func(l *lua.LState) int {
		sys.resetRemapInput()
		return 0
	})
	luaRegister(l, "remapInput", func(l *lua.LState) int {
		src, dst := int(numArg(l, 1)), int(numArg(l, 2))
		if src < 1 || src > len(sys.inputRemap) ||
			dst < 1 || dst > len(sys.inputRemap) {
			l.RaiseError("プレイヤー番号(%v, %v)が不正です。", src, dst)
		}
		sys.inputRemap[src-1] = dst - 1
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
		tbl := l.NewTable()
		tbl_chars := l.NewTable()
		load := func() error {
			sys.loader.runTread()
			for sys.loader.state != LS_Complete {
				if sys.loader.state == LS_Error {
					return sys.loader.err
				} else if sys.loader.state == LS_Cancel {
					return nil
				}
				sys.await(FPS)
			}
			for i := range sys.cgi {
				if i < len(sys.lifebar.fa[sys.tmode[i&1]]) {
					fa := sys.lifebar.fa[sys.tmode[i&1]][i]
					fa.face = sys.cgi[i].sff.getOwnPalSprite(
						int16(fa.face_spr[0]), int16(fa.face_spr[1]))

					fa.scale = sys.cgi[i].portraitscale
				}

			}
			runtime.GC()
			return nil
		}
		for {
			if sys.gameEnd {
				l.Push(lua.LNumber(-1))
				return 1
			}
			winp := int32(0)
			p := make([]*Char, len(sys.chars))
			sys.roundsExisted = [2]int32{}
			sys.matchWins = [2]int32{}
			for i := range sys.lifebar.wi {
				sys.lifebar.wi[i].clear()
			}
			sys.draws = 0
			fight := func() (int32, error) {
				if err := load(); err != nil {
					return -1, err
				}
				if sys.loader.state == LS_Cancel {
					return -1, nil
				}
				sys.charList.clear()
				for i := 0; i < len(sys.chars); i += 2 {
					if len(sys.chars[i]) > 0 {
						sys.chars[i][0].id = sys.newCharId()
					}
				}
				for i := 1; i < len(sys.chars); i += 2 {
					if len(sys.chars[i]) > 0 {
						sys.chars[i][0].id = sys.newCharId()
					}
				}
				for i, c := range sys.chars {
					if len(c) > 0 {
						p[i] = c[0]
						sys.charList.add(c[0])
						if sys.roundsExisted[i&1] == 0 {
							c[0].loadPallet()
						}
						for j, cj := range sys.chars {
							if i != j && len(cj) > 0 {
								if len(cj[0].cmd) == 0 {
									cj[0].cmd = make([]CommandList, len(sys.chars))
								}
								cj[0].cmd[i].CopyList(c[0].cmd[i])
							}
						}
					}
				}
				if sys.round == 1 {
					if sys.tmode[1] == TM_Turns {
						sys.matchWins[0] = sys.numTurns[1]
					} else {
						sys.matchWins[0] = sys.lifebar.ro.match_wins
					}
					if sys.tmode[0] == TM_Turns {
						sys.matchWins[1] = sys.numTurns[0]
					} else {
						sys.matchWins[1] = sys.lifebar.ro.match_wins
					}
					sys.stage.reset()
				}
				winp := int32(0)
				if sys.fight() {
					sys.chars = [len(sys.chars)][]*Char{}
					sys.loaderReset()
					winp = -2
				} else if sys.esc {
					winp = -1
				} else {
					w1 := sys.wins[0] >= sys.matchWins[0]
					w2 := sys.wins[1] >= sys.matchWins[1]
					if w1 != w2 {
						winp = Btoi(w1) + Btoi(w2)*2
					}
					tbl_roundNo := l.NewTable()
					for _, p := range sys.chars {
						if len(p) > 0 {
							tmp := l.NewTable()
							tmp.RawSetString("name", lua.LString(p[0].name))
							tmp.RawSetString("memberNo", lua.LNumber(p[0].memberNo))
							tmp.RawSetString("selectNo", lua.LNumber(p[0].selectNo))
							tmp.RawSetString("life", lua.LNumber(p[0].life))
							tmp.RawSetString("lifeMax", lua.LNumber(p[0].lifeMax))
							tmp.RawSetString("winquote", lua.LNumber(p[0].winquote))
							tmp.RawSetString("aiLevel", lua.LNumber(p[0].aiLevel()))
							tmp.RawSetString("palno", lua.LNumber(p[0].palno()))
							tmp.RawSetString("win", lua.LBool(p[0].win()))
							tmp.RawSetString("winKO", lua.LBool(p[0].winKO()))
							tmp.RawSetString("winTime", lua.LBool(p[0].winTime()))
							tmp.RawSetString("winPerfect", lua.LBool(p[0].winPerfect()))
							tmp.RawSetString("drawgame", lua.LBool(p[0].drawgame()))
							tmp.RawSetString("ko", lua.LBool(p[0].scf(SCF_ko)))
							tmp.RawSetString("ko_round_middle", lua.LBool(p[0].scf(SCF_ko_round_middle)))
							tbl_roundNo.RawSetInt(p[0].playerNo+1, tmp)
						}
					}
					tbl_chars.RawSetInt(int(sys.round-1), tbl_roundNo)
				}
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
					sys.wins[0] >= sys.matchWins[0] || sys.wins[1] >= sys.matchWins[1] ||
					sys.gameEnd {
					break
				}
				for i := 0; i < 2; i++ {
					if p[i].life <= 0 && sys.tmode[i] == TM_Turns {
						sys.lifebar.fa[TM_Turns][i].numko++
						sys.roundsExisted[i] = 0
					}
				}
				sys.loader.reset()
			}
			if winp != -2 {
				time := int32(0)
				tbl_time := l.NewTable()
				for k, v := range sys.timerCount {
					tbl_time.RawSetInt(k+1, lua.LNumber(v))
					time = time + v
				}
				tbl.RawSetString("chars", tbl_chars)
				tbl.RawSetString("time_rounds", tbl_time)
				tbl.RawSetString("time", lua.LNumber(time))
				tbl.RawSetString("roundTime", lua.LNumber(sys.roundTime))
				tbl.RawSetString("winTeam", lua.LNumber(sys.winTeam))
				tbl.RawSetString("lastRound", lua.LNumber(sys.round-1))
				tbl.RawSetString("draws", lua.LNumber(sys.draws))
				tbl.RawSetString("P1wins", lua.LNumber(sys.wins[0]))
				tbl.RawSetString("P2wins", lua.LNumber(sys.wins[1]))
				tbl.RawSetString("P1tmode", lua.LNumber(sys.tmode[0]))
				tbl.RawSetString("P2tmode", lua.LNumber(sys.tmode[1]))
				sys.timerCount = []int32{}
				l.Push(lua.LNumber(winp))
				l.Push(tbl)
				return 2
			}
		}
	})
	luaRegister(l, "getCharVar", func(*lua.LState) int {
		for _, p := range sys.chars {
			if len(p) > 0 && p[0].playerNo+1 == int(numArg(l, 1)) {
				if strArg(l, 2) == "varGet" {
					l.Push(lua.LNumber(p[0].varGet(int32(numArg(l, 3))).ToI()))
				} else if strArg(l, 2) == "fvarGet" {
					l.Push(lua.LNumber(p[0].fvarGet(int32(numArg(l, 3))).ToI()))
				} else if strArg(l, 2) == "sysVarGet" {
					l.Push(lua.LNumber(p[0].sysVarGet(int32(numArg(l, 3))).ToI()))
				} else if strArg(l, 2) == "sysFvarGet" {
					l.Push(lua.LNumber(p[0].sysFvarGet(int32(numArg(l, 3))).ToI()))
				}
				break
			}
		}
		return 1
	})
	luaRegister(l, "getCharVictoryQuote", func(*lua.LState) int {
		v := int(-1)
		for _, p := range sys.chars {
			if len(p) > 0 && p[0].playerNo+1 == int(numArg(l, 1)) {
				if l.GetTop() >= 2 {
					v = int(numArg(l, 2))
				} else {
					v = int(p[0].winquote)
				}
				if v < 0 || v >= MaxQuotes {
					t := []string{}
					for i, q := range sys.cgi[p[0].playerNo].quotes {
						if sys.cgi[p[0].playerNo].quotes[i] != "" {
							t = append(t, q)
						}
					}
					if len(t) > 0 {
						v = rand.Int() % len(t)
					} else {
						v = -1
					}
				}
				if len(sys.cgi[p[0].playerNo].quotes) == MaxQuotes && v != -1 {
					l.Push(lua.LString(sys.cgi[p[0].playerNo].quotes[v]))
				} else {
					l.Push(lua.LString(""))
				}
				break
			}
		}
		return 1
	})
	luaRegister(l, "getCommandLineFlags", func(*lua.LState) int {
		tbl := l.NewTable()
		for k, v := range sys.cmdFlags {
			tbl.RawSetString(k, lua.LString(v))
		}
		l.Push(tbl)
		return 1
	})
	luaRegister(l, "getCommandLineValue", func(*lua.LState) int {
		l.Push(lua.LString(sys.cmdFlags[strArg(l, 1)]))
		return 1
	})
	luaRegister(l, "setPortrait", func(*lua.LState) int {
		p := int(numArg(l, 3))
		if p == 1 {
			sys.sel.lportrait = [...]int16{int16(numArg(l, 1)), int16(numArg(l, 2))}
		} else if p == 2 {
			sys.sel.sportrait = [...]int16{int16(numArg(l, 1)), int16(numArg(l, 2))}
		} else if p == 3 {
			sys.sel.vsportrait = [...]int16{int16(numArg(l, 1)), int16(numArg(l, 2))}
		} else if p == 4 {
			sys.sel.vportrait = [...]int16{int16(numArg(l, 1)), int16(numArg(l, 2))}
		}
		return 0
	})
	luaRegister(l, "drawSmallPortrait", func(l *lua.LState) int {
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
			if c != nil && c.sportrait != nil {
				if c.portrait_scale != 1 {
					xscl *= c.portrait_scale
					yscl *= c.portrait_scale
				}
				c.sportrait.Draw(x, y, xscl, yscl, c.sportrait.Pal)
			}
		}
		return 0
	})
	luaRegister(l, "drawVersusPortrait", func(l *lua.LState) int {
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
			if c != nil && c.vsportrait != nil {
				if c.portrait_scale != 1 {
					xscl *= c.portrait_scale
					yscl *= c.portrait_scale
				}
				c.vsportrait.Draw(x, y, xscl, yscl, c.vsportrait.Pal)
			}
		}
		return 0
	})
	luaRegister(l, "drawVictoryPortrait", func(l *lua.LState) int {
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
			if c != nil && c.vportrait != nil {
				if c.portrait_scale != 1 {
					xscl *= c.portrait_scale
					yscl *= c.portrait_scale
				}
				c.vportrait.Draw(x, y, xscl, yscl, c.vportrait.Pal)
			}
		}
		return 0
	})
	luaRegister(l, "getCharIntro", func(*lua.LState) int {
		c := sys.sel.GetChar(int(numArg(l, 1)))
		l.Push(lua.LString(c.intro_storyboard))
		return 1
	})
	luaRegister(l, "getCharEnding", func(*lua.LState) int {
		c := sys.sel.GetChar(int(numArg(l, 1)))
		l.Push(lua.LString(c.ending_storyboard))
		return 1
	})
	luaRegister(l, "getCharPalettes", func(*lua.LState) int {
		c := sys.sel.GetChar(int(numArg(l, 1)))
		tbl := l.NewTable()
		var pal []int32
		if sys.aiRandomColor {
			pal = c.pal
		} else {
			pal = c.pal_defaults
		}
		if len(pal) > 0 {
			for k, v := range pal {
				tbl.RawSetInt(k+1, lua.LNumber(v))
			}
		} else {
			tbl.RawSetInt(1, lua.LNumber(1))
		}
		l.Push(tbl)
		return 1
	})
	luaRegister(l, "getCharRandomPalette", func(*lua.LState) int {
		c := sys.sel.GetChar(int(numArg(l, 1)))
		if len(c.pal) > 0 {
			n := rand.Int() % len(c.pal)
			l.Push(lua.LNumber(c.pal[n]))
		} else {
			l.Push(lua.LNumber(1))
		}
		return 1
	})
	luaRegister(l, "getStageInfo", func(*lua.LState) int {
		a, b, c, d := sys.sel.GetStageInfo(int(numArg(l, 1)))
		l.Push(lua.LString(a))
		l.Push(lua.LString(b))
		l.Push(lua.LString(c))
		l.Push(lua.LString(d))
		return 4
	})
	luaRegister(l, "getKey", func(*lua.LState) int {
		s := ""
		if sys.keyInput != glfw.KeyUnknown {
			s = KeyToString(sys.keyInput)
		}
		l.Push(lua.LString(s))
		return 1
	})
	luaRegister(l, "getKeyText", func(*lua.LState) int {
		s := ""
		if sys.keyInput != glfw.KeyUnknown {
			if sys.keyInput == glfw.KeyInsert {
				s, _ = sys.window.GetClipboardString()
			} else {
				s = sys.keyString
			}
		}
		l.Push(lua.LString(s))
		return 1
	})
	luaRegister(l, "resetKey", func(*lua.LState) int {
		sys.keyInput = glfw.KeyUnknown
		sys.keyString = ""
		return 0
	})
	luaRegister(l, "getSpriteInfo", func(*lua.LState) int {
		var s *Sprite
		var err error
		def := strArg(l, 1)
		err = LoadFile(&def, "", func(file string) error {
			s, err = loadFromSff(file, int16(numArg(l, 2)), int16(numArg(l, 3)))
			return err
		})
		if err != nil {
			l.Push(lua.LNumber(0))
			l.Push(lua.LNumber(0))
			l.Push(lua.LNumber(0))
			l.Push(lua.LNumber(0))
			return 4
		}
		l.Push(lua.LNumber(s.Size[0]))
		l.Push(lua.LNumber(s.Size[1]))
		l.Push(lua.LNumber(s.Offset[0]))
		l.Push(lua.LNumber(s.Offset[1]))
		return 4
	})
}

// Trigger Script

func triggerScriptInit(l *lua.LState) {
	sys.debugWC = sys.chars[0][0]
	luaRegister(l, "player", func(*lua.LState) int {
		pn := int(numArg(l, 1))
		ret := false
		if pn >= 1 && pn <= len(sys.chars) && len(sys.chars[pn-1]) > 0 {
			sys.debugWC, ret = sys.chars[pn-1][0], true
		}
		l.Push(lua.LBool(ret))
		return 1
	})
	luaRegister(l, "parent", func(*lua.LState) int {
		ret := false
		if c := sys.debugWC.parent(); c != nil {
			sys.debugWC, ret = c, true
		}
		l.Push(lua.LBool(ret))
		return 1
	})
	luaRegister(l, "root", func(*lua.LState) int {
		ret := false
		if c := sys.debugWC.root(); c != nil {
			sys.debugWC, ret = c, true
		}
		l.Push(lua.LBool(ret))
		return 1
	})
	luaRegister(l, "helper", func(*lua.LState) int {
		ret, id := false, int32(0)
		if l.GetTop() >= 1 {
			id = int32(numArg(l, 1))
		}
		if c := sys.debugWC.helper(id); c != nil {
			sys.debugWC, ret = c, true
		}
		l.Push(lua.LBool(ret))
		return 1
	})
	luaRegister(l, "target", func(*lua.LState) int {
		ret, id := false, int32(-1)
		if l.GetTop() >= 1 {
			id = int32(numArg(l, 1))
		}
		if c := sys.debugWC.target(id); c != nil {
			sys.debugWC, ret = c, true
		}
		l.Push(lua.LBool(ret))
		return 1
	})
	luaRegister(l, "partner", func(*lua.LState) int {
		ret := false
		if c := sys.debugWC.partner(0); c != nil {
			sys.debugWC, ret = c, true
		}
		l.Push(lua.LBool(ret))
		return 1
	})
	luaRegister(l, "enemy", func(*lua.LState) int {
		ret, n := false, int32(0)
		if l.GetTop() >= 1 {
			n = int32(numArg(l, 1))
		}
		if c := sys.debugWC.enemy(n); c != nil {
			sys.debugWC, ret = c, true
		}
		l.Push(lua.LBool(ret))
		return 1
	})
	luaRegister(l, "enemynear", func(*lua.LState) int {
		ret, n := false, int32(0)
		if l.GetTop() >= 1 {
			n = int32(numArg(l, 1))
		}
		if c := sys.debugWC.enemyNear(n); c != nil {
			sys.debugWC, ret = c, true
		}
		l.Push(lua.LBool(ret))
		return 1
	})
	luaRegister(l, "playerid", func(*lua.LState) int {
		ret := false
		if c := sys.playerID(int32(numArg(l, 1))); c != nil {
			sys.debugWC, ret = c, true
		}
		l.Push(lua.LBool(ret))
		return 1
	})
	luaRegister(l, "ailevel", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.aiLevel()))
		return 1
	})
	luaRegister(l, "alive", func(*lua.LState) int {
		l.Push(lua.LBool(sys.debugWC.alive()))
		return 1
	})
	luaRegister(l, "anim", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.animNo))
		return 1
	})
	luaRegister(l, "animOwner", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.animPN) + 1)
		return 1
	})
	luaRegister(l, "animelemno", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.animElemNo(int32(numArg(l, 1))).ToI()))
		return 1
	})
	luaRegister(l, "animelemtime", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.animElemTime(int32(numArg(l, 1))).ToI()))
		return 1
	})
	luaRegister(l, "animexist", func(*lua.LState) int {
		l.Push(lua.LBool(sys.debugWC.animExist(sys.debugWC,
			BytecodeInt(int32(numArg(l, 1)))).ToB()))
		return 1
	})
	luaRegister(l, "animtime", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.animTime()))
		return 1
	})
	luaRegister(l, "authorname", func(*lua.LState) int {
		l.Push(lua.LString(sys.debugWC.gi().author))
		return 1
	})
	luaRegister(l, "backedge", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.backEdge()))
		return 1
	})
	luaRegister(l, "backedgebodydist", func(*lua.LState) int {
		l.Push(lua.LNumber(int32(sys.debugWC.backEdgeBodyDist())))
		return 1
	})
	luaRegister(l, "backedgedist", func(*lua.LState) int {
		l.Push(lua.LNumber(int32(sys.debugWC.backEdgeDist())))
		return 1
	})
	luaRegister(l, "bottomedge", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.bottomEdge()))
		return 1
	})
	luaRegister(l, "cameraposX", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.cam.Pos[0]))
		return 1
	})
	luaRegister(l, "cameraposY", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.cam.Pos[1]))
		return 1
	})
	luaRegister(l, "camerazoom", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.cam.Scale))
		return 1
	})
	luaRegister(l, "canrecover", func(*lua.LState) int {
		l.Push(lua.LBool(sys.debugWC.canRecover()))
		return 1
	})
	luaRegister(l, "command", func(*lua.LState) int {
		l.Push(lua.LBool(sys.debugWC.commandByName(strArg(l, 1))))
		return 1
	})
	luaRegister(l, "const", func(*lua.LState) int {
		c := sys.debugWC
		var ln lua.LNumber
		switch strArg(l, 1) {
		case "data.life":
			ln = lua.LNumber(c.gi().data.life)
		case "data.power":
			ln = lua.LNumber(c.gi().data.power)
		case "data.attack":
			ln = lua.LNumber(c.gi().data.attack)
		case "data.defence":
			ln = lua.LNumber(c.gi().data.defence)
		case "data.fall.defence_mul":
			ln = lua.LNumber(c.gi().data.fall.defence_mul)
		case "data.liedown.time":
			ln = lua.LNumber(c.gi().data.liedown.time)
		case "data.airjuggle":
			ln = lua.LNumber(c.gi().data.airjuggle)
		case "data.sparkno":
			ln = lua.LNumber(c.gi().data.sparkno)
		case "data.guard.sparkno":
			ln = lua.LNumber(c.gi().data.guard.sparkno)
		case "data.ko.echo":
			ln = lua.LNumber(c.gi().data.ko.echo)
		case "data.intpersistindex":
			ln = lua.LNumber(c.gi().data.intpersistindex)
		case "data.floatpersistindex":
			ln = lua.LNumber(c.gi().data.floatpersistindex)
		case "size.xscale":
			ln = lua.LNumber(c.size.xscale)
		case "size.yscale":
			ln = lua.LNumber(c.size.yscale)
		case "size.ground.back":
			ln = lua.LNumber(c.size.ground.back)
		case "size.ground.front":
			ln = lua.LNumber(c.size.ground.front)
		case "size.air.back":
			ln = lua.LNumber(c.size.air.back)
		case "size.air.front":
			ln = lua.LNumber(c.size.air.front)
		case "size.z.width":
			ln = lua.LNumber(c.size.z.width)
		case "size.height":
			ln = lua.LNumber(c.size.height)
		case "size.attack.dist":
			ln = lua.LNumber(c.size.attack.dist)
		case "size.attack.z.width.back":
			ln = lua.LNumber(c.size.attack.z.width[1])
		case "size.attack.z.width.front":
			ln = lua.LNumber(c.size.attack.z.width[0])
		case "size.proj.attack.dist":
			ln = lua.LNumber(c.size.proj.attack.dist)
		case "size.proj.doscale":
			ln = lua.LNumber(c.size.proj.doscale)
		case "size.head.pos.x":
			ln = lua.LNumber(c.size.head.pos[0])
		case "size.head.pos.y":
			ln = lua.LNumber(c.size.head.pos[1])
		case "size.mid.pos.x":
			ln = lua.LNumber(c.size.mid.pos[0])
		case "size.mid.pos.y":
			ln = lua.LNumber(c.size.mid.pos[1])
		case "size.shadowoffset":
			ln = lua.LNumber(c.size.shadowoffset)
		case "size.draw.offset.x":
			ln = lua.LNumber(c.size.draw.offset[0])
		case "size.draw.offset.y":
			ln = lua.LNumber(c.size.draw.offset[1])
		case "velocity.walk.fwd.x":
			ln = lua.LNumber(c.gi().velocity.walk.fwd)
		case "velocity.walk.back.x":
			ln = lua.LNumber(c.gi().velocity.walk.back)
		case "velocity.walk.up.x":
			ln = lua.LNumber(c.gi().velocity.walk.up.x)
		case "velocity.walk.down.x":
			ln = lua.LNumber(c.gi().velocity.walk.down.x)
		case "velocity.run.fwd.x":
			ln = lua.LNumber(c.gi().velocity.run.fwd[0])
		case "velocity.run.fwd.y":
			ln = lua.LNumber(c.gi().velocity.run.fwd[1])
		case "velocity.run.back.x":
			ln = lua.LNumber(c.gi().velocity.run.back[0])
		case "velocity.run.back.y":
			ln = lua.LNumber(c.gi().velocity.run.back[1])
		case "velocity.run.up.x":
			ln = lua.LNumber(c.gi().velocity.run.up.x)
		case "velocity.run.up.y":
			ln = lua.LNumber(c.gi().velocity.run.up.y)
		case "velocity.run.down.x":
			ln = lua.LNumber(c.gi().velocity.run.down.x)
		case "velocity.run.down.y":
			ln = lua.LNumber(c.gi().velocity.run.down.y)
		case "velocity.jump.y":
			ln = lua.LNumber(c.gi().velocity.jump.neu[1])
		case "velocity.jump.neu.x":
			ln = lua.LNumber(c.gi().velocity.jump.neu[0])
		case "velocity.jump.back.x":
			ln = lua.LNumber(c.gi().velocity.jump.back)
		case "velocity.jump.fwd.x":
			ln = lua.LNumber(c.gi().velocity.jump.fwd)
		case "velocity.jump.up.x":
			ln = lua.LNumber(c.gi().velocity.jump.up.x)
		case "velocity.jump.down.x":
			ln = lua.LNumber(c.gi().velocity.jump.down.x)
		case "velocity.runjump.back.x":
			ln = lua.LNumber(c.gi().velocity.runjump.back[0])
		case "velocity.runjump.back.y":
			ln = lua.LNumber(c.gi().velocity.runjump.back[1])
		case "velocity.runjump.y":
			ln = lua.LNumber(c.gi().velocity.runjump.fwd[1])
		case "velocity.runjump.fwd.x":
			ln = lua.LNumber(c.gi().velocity.runjump.fwd[0])
		case "velocity.runjump.up.x":
			ln = lua.LNumber(c.gi().velocity.runjump.up.x)
		case "velocity.runjump.down.x":
			ln = lua.LNumber(c.gi().velocity.runjump.down.x)
		case "velocity.airjump.y":
			ln = lua.LNumber(c.gi().velocity.airjump.neu[1])
		case "velocity.airjump.neu.x":
			ln = lua.LNumber(c.gi().velocity.airjump.neu[0])
		case "velocity.airjump.back.x":
			ln = lua.LNumber(c.gi().velocity.airjump.back)
		case "velocity.airjump.fwd.x":
			ln = lua.LNumber(c.gi().velocity.airjump.fwd)
		case "velocity.airjump.up.x":
			ln = lua.LNumber(c.gi().velocity.airjump.up.x)
		case "velocity.airjump.down.x":
			ln = lua.LNumber(c.gi().velocity.airjump.down.x)
		case "velocity.air.gethit.groundrecover.x":
			ln = lua.LNumber(c.gi().velocity.air.gethit.groundrecover[0])
		case "velocity.air.gethit.groundrecover.y":
			ln = lua.LNumber(c.gi().velocity.air.gethit.groundrecover[1])
		case "velocity.air.gethit.airrecover.mul.x":
			ln = lua.LNumber(c.gi().velocity.air.gethit.airrecover.mul[0])
		case "velocity.air.gethit.airrecover.mul.y":
			ln = lua.LNumber(c.gi().velocity.air.gethit.airrecover.mul[1])
		case "velocity.air.gethit.airrecover.add.x":
			ln = lua.LNumber(c.gi().velocity.air.gethit.airrecover.add[0])
		case "velocity.air.gethit.airrecover.add.y":
			ln = lua.LNumber(c.gi().velocity.air.gethit.airrecover.add[1])
		case "velocity.air.gethit.airrecover.back":
			ln = lua.LNumber(c.gi().velocity.air.gethit.airrecover.back)
		case "velocity.air.gethit.airrecover.fwd":
			ln = lua.LNumber(c.gi().velocity.air.gethit.airrecover.fwd)
		case "velocity.air.gethit.airrecover.up":
			ln = lua.LNumber(c.gi().velocity.air.gethit.airrecover.up)
		case "velocity.air.gethit.airrecover.down":
			ln = lua.LNumber(c.gi().velocity.air.gethit.airrecover.down)
		case "movement.airjump.num":
			ln = lua.LNumber(c.gi().movement.airjump.num)
		case "movement.airjump.height":
			ln = lua.LNumber(c.gi().movement.airjump.height)
		case "movement.yaccel":
			ln = lua.LNumber(c.gi().movement.yaccel)
		case "movement.stand.friction":
			ln = lua.LNumber(c.gi().movement.stand.friction)
		case "movement.crouch.friction":
			ln = lua.LNumber(c.gi().movement.crouch.friction)
		case "movement.stand.friction.threshold":
			ln = lua.LNumber(c.gi().movement.stand.friction_threshold)
		case "movement.crouch.friction.threshold":
			ln = lua.LNumber(c.gi().movement.crouch.friction_threshold)
		case "movement.air.gethit.groundlevel":
			ln = lua.LNumber(c.gi().movement.air.gethit.groundlevel)
		case "movement.air.gethit.groundrecover.ground.threshold":
			ln = lua.LNumber(
				c.gi().movement.air.gethit.groundrecover.ground.threshold)
		case "movement.air.gethit.groundrecover.groundlevel":
			ln = lua.LNumber(c.gi().movement.air.gethit.groundrecover.groundlevel)
		case "movement.air.gethit.airrecover.threshold":
			ln = lua.LNumber(c.gi().movement.air.gethit.airrecover.threshold)
		case "movement.air.gethit.airrecover.yaccel":
			ln = lua.LNumber(c.gi().movement.air.gethit.airrecover.yaccel)
		case "movement.air.gethit.trip.groundlevel":
			ln = lua.LNumber(c.gi().movement.air.gethit.trip.groundlevel)
		case "movement.down.bounce.offset.x":
			ln = lua.LNumber(c.gi().movement.down.bounce.offset[0])
		case "movement.down.bounce.offset.y":
			ln = lua.LNumber(c.gi().movement.down.bounce.offset[1])
		case "movement.down.bounce.yaccel":
			ln = lua.LNumber(c.gi().movement.down.bounce.yaccel)
		case "movement.down.bounce.groundlevel":
			ln = lua.LNumber(c.gi().movement.down.bounce.groundlevel)
		case "movement.down.friction.threshold":
			ln = lua.LNumber(c.gi().movement.down.friction_threshold)
		}
		l.Push(ln)
		return 1
	})
	luaRegister(l, "ctrl", func(*lua.LState) int {
		l.Push(lua.LBool(sys.debugWC.ctrl()))
		return 1
	})
	luaRegister(l, "drawgame", func(*lua.LState) int {
		l.Push(lua.LBool(sys.debugWC.drawgame()))
		return 1
	})
	luaRegister(l, "facing", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.facing))
		return 1
	})
	luaRegister(l, "frontedge", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.frontEdge()))
		return 1
	})
	luaRegister(l, "frontedgebodydist", func(*lua.LState) int {
		l.Push(lua.LNumber(int32(sys.debugWC.frontEdgeBodyDist())))
		return 1
	})
	luaRegister(l, "frontedgedist", func(*lua.LState) int {
		l.Push(lua.LNumber(int32(sys.debugWC.frontEdgeDist())))
		return 1
	})
	luaRegister(l, "fvar", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.fvarGet(int32(numArg(l, 1))).ToF()))
		return 1
	})
	luaRegister(l, "gameheight", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.gameHeight()))
		return 1
	})
	luaRegister(l, "gametime", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.gameTime))
		return 1
	})
	luaRegister(l, "gamewidth", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.gameWidth()))
		return 1
	})
	luaRegister(l, "gethitvar", func(*lua.LState) int {
		c := sys.debugWC
		var ln lua.LNumber
		switch strArg(l, 1) {
		case "xveladd":
			ln = lua.LNumber(0)
		case "yveladd":
			ln = lua.LNumber(0)
		case "type":
			ln = lua.LNumber(0)
		case "zoff":
			ln = lua.LNumber(0)
		case "fall.envshake.dir":
			ln = lua.LNumber(0)
		case "animtype":
			ln = lua.LNumber(c.gethitAnimtype())
		case "airtype":
			ln = lua.LNumber(c.ghv.airtype)
		case "groundtype":
			ln = lua.LNumber(c.ghv.groundtype)
		case "damage":
			ln = lua.LNumber(c.ghv.damage)
		case "hitcount":
			ln = lua.LNumber(c.ghv.hitcount)
		case "fallcount":
			ln = lua.LNumber(c.ghv.fallcount)
		case "hitshaketime":
			ln = lua.LNumber(c.ghv.hitshaketime)
		case "hittime":
			ln = lua.LNumber(c.ghv.hittime)
		case "slidetime":
			ln = lua.LNumber(c.ghv.slidetime)
		case "ctrltime":
			ln = lua.LNumber(c.ghv.ctrltime)
		case "recovertime":
			ln = lua.LNumber(c.recoverTime)
		case "xoff":
			ln = lua.LNumber(c.ghv.xoff)
		case "yoff":
			ln = lua.LNumber(c.ghv.yoff)
		case "xvel":
			ln = lua.LNumber(c.ghv.xvel * c.facing)
		case "yvel":
			ln = lua.LNumber(c.ghv.yvel)
		case "yaccel":
			ln = lua.LNumber(c.ghv.getYaccel(c))
		case "hitid", "chainid":
			ln = lua.LNumber(c.ghv.chainId())
		case "guarded":
			ln = lua.LNumber(Btoi(c.ghv.guarded))
		case "isbound":
			ln = lua.LNumber(Btoi(c.isBound()))
		case "fall":
			ln = lua.LNumber(Btoi(c.ghv.fallf))
		case "fall.damage":
			ln = lua.LNumber(c.ghv.fall.damage)
		case "fall.xvel":
			ln = lua.LNumber(c.ghv.fall.xvel())
		case "fall.yvel":
			ln = lua.LNumber(c.ghv.fall.yvelocity)
		case "fall.recover":
			ln = lua.LNumber(Btoi(c.ghv.fall.recover))
		case "fall.time":
			ln = lua.LNumber(c.fallTime)
		case "fall.recovertime":
			ln = lua.LNumber(c.ghv.fall.recovertime)
		case "fall.kill":
			ln = lua.LNumber(Btoi(c.ghv.fall.kill))
		case "fall.envshake.time":
			ln = lua.LNumber(c.ghv.fall.envshake_time)
		case "fall.envshake.freq":
			ln = lua.LNumber(c.ghv.fall.envshake_freq)
		case "fall.envshake.ampl":
			ln = lua.LNumber(c.ghv.fall.envshake_ampl)
		case "fall.envshake.phase":
			ln = lua.LNumber(c.ghv.fall.envshake_phase)
		}
		l.Push(ln)
		return 1
	})
	luaRegister(l, "hitcount", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.hitCount))
		return 1
	})
	luaRegister(l, "hitdefattr", func(*lua.LState) int {
		attr, str := sys.debugWC.hitdef.attr, ""
		if sys.debugWC.ss.moveType == MT_A {
			if attr&int32(ST_S) != 0 {
				str += "S"
			}
			if attr&int32(ST_C) != 0 {
				str += "C"
			}
			if attr&int32(ST_A) != 0 {
				str += "A"
			}
			if attr&int32(AT_NA) != 0 {
				str += ", NA"
			}
			if attr&int32(AT_NT) != 0 {
				str += ", NT"
			}
			if attr&int32(AT_NP) != 0 {
				str += ", NP"
			}
			if attr&int32(AT_SA) != 0 {
				str += ", SA"
			}
			if attr&int32(AT_ST) != 0 {
				str += ", ST"
			}
			if attr&int32(AT_SP) != 0 {
				str += ", SP"
			}
			if attr&int32(AT_HA) != 0 {
				str += ", HA"
			}
			if attr&int32(AT_HT) != 0 {
				str += ", HT"
			}
			if attr&int32(AT_HP) != 0 {
				str += ", HP"
			}
		}
		l.Push(lua.LString(str))
		return 1
	})
	luaRegister(l, "hitfall", func(*lua.LState) int {
		l.Push(lua.LBool(sys.debugWC.ghv.fallf))
		return 1
	})
	luaRegister(l, "hitover", func(*lua.LState) int {
		l.Push(lua.LBool(sys.debugWC.hitOver()))
		return 1
	})
	luaRegister(l, "hitpausetime", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.hitPauseTime))
		return 1
	})
	luaRegister(l, "hitshakeover", func(*lua.LState) int {
		l.Push(lua.LBool(sys.debugWC.hitShakeOver()))
		return 1
	})
	luaRegister(l, "hitvelX", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.hitVelX()))
		return 1
	})
	luaRegister(l, "hitvelY", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.hitVelY()))
		return 1
	})
	luaRegister(l, "id", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.id))
		return 1
	})
	luaRegister(l, "inguarddist", func(*lua.LState) int {
		l.Push(lua.LBool(sys.debugWC.inguarddist))
		return 1
	})
	luaRegister(l, "ishelper", func(*lua.LState) int {
		id := int32(0)
		if l.GetTop() >= 1 {
			id = int32(numArg(l, 1))
		}
		l.Push(lua.LBool(sys.debugWC.isHelper(BytecodeInt(id)).ToB()))
		return 1
	})
	luaRegister(l, "ishometeam", func(*lua.LState) int {
		l.Push(lua.LBool(sys.debugWC.playerNo&1 == sys.home))
		return 1
	})
	luaRegister(l, "leftedge", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.leftEdge()))
		return 1
	})
	luaRegister(l, "life", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.life))
		return 1
	})
	luaRegister(l, "lifemax", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.lifeMax))
		return 1
	})
	luaRegister(l, "lose", func(*lua.LState) int {
		l.Push(lua.LBool(sys.debugWC.lose()))
		return 1
	})
	luaRegister(l, "loseko", func(*lua.LState) int {
		l.Push(lua.LBool(sys.debugWC.loseKO()))
		return 1
	})
	luaRegister(l, "losetime", func(*lua.LState) int {
		l.Push(lua.LBool(sys.debugWC.loseTime()))
		return 1
	})
	luaRegister(l, "matchno", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.match))
		return 1
	})
	luaRegister(l, "matchover", func(*lua.LState) int {
		l.Push(lua.LBool(sys.matchOver()))
		return 1
	})
	luaRegister(l, "movecontact", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.moveContact()))
		return 1
	})
	luaRegister(l, "moveguarded", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.moveGuarded()))
		return 1
	})
	luaRegister(l, "movehit", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.moveHit()))
		return 1
	})
	luaRegister(l, "movereversed", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.moveReversed()))
		return 1
	})
	luaRegister(l, "movetype", func(*lua.LState) int {
		var s string
		switch sys.debugWC.ss.moveType {
		case MT_I:
			s = "I"
		case MT_A:
			s = "A"
		case MT_H:
			s = "H"
		}
		l.Push(lua.LString(s))
		return 1
	})
	luaRegister(l, "name", func(*lua.LState) int {
		l.Push(lua.LString(sys.debugWC.name))
		return 1
	})
	luaRegister(l, "numenemy", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.numEnemy()))
		return 1
	})
	luaRegister(l, "numexplod", func(*lua.LState) int {
		id := int32(-1)
		if l.GetTop() >= 1 {
			id = int32(numArg(l, 1))
		}
		l.Push(lua.LNumber(sys.debugWC.numExplod(BytecodeInt(id)).ToI()))
		return 1
	})
	luaRegister(l, "numhelper", func(*lua.LState) int {
		id := int32(0)
		if l.GetTop() >= 1 {
			id = int32(numArg(l, 1))
		}
		l.Push(lua.LNumber(sys.debugWC.numHelper(BytecodeInt(id)).ToI()))
		return 1
	})
	luaRegister(l, "numpartner", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.numPartner()))
		return 1
	})
	luaRegister(l, "numproj", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.numProj()))
		return 1
	})
	luaRegister(l, "numprojid", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.numProjID(
			BytecodeInt(int32(numArg(l, 1)))).ToI()))
		return 1
	})
	luaRegister(l, "numtarget", func(*lua.LState) int {
		id := int32(-1)
		if l.GetTop() >= 1 {
			id = int32(numArg(l, 1))
		}
		l.Push(lua.LNumber(sys.debugWC.numTarget(BytecodeInt(id)).ToI()))
		return 1
	})
	luaRegister(l, "palno", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.gi().palno))
		return 1
	})
	luaRegister(l, "physics", func(*lua.LState) int {
		var s string
		switch sys.debugWC.ss.physics {
		case ST_S:
			s = "S"
		case ST_C:
			s = "C"
		case ST_A:
			s = "A"
		case ST_N:
			s = "N"
		}
		l.Push(lua.LString(s))
		return 1
	})
	luaRegister(l, "posX", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.pos[0] - sys.cam.Pos[0]))
		return 1
	})
	luaRegister(l, "posY", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.pos[1]))
		return 1
	})
	luaRegister(l, "power", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.getPower()))
		return 1
	})
	luaRegister(l, "powermax", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.powerMax))
		return 1
	})
	luaRegister(l, "playeridexist", func(*lua.LState) int {
		l.Push(lua.LBool(sys.playerIDExist(
			BytecodeInt(int32(numArg(l, 1)))).ToB()))
		return 1
	})
	luaRegister(l, "prevstateno", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.ss.prevno))
		return 1
	})
	luaRegister(l, "projcanceltime", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.projCancelTime(
			BytecodeInt(int32(numArg(l, 1)))).ToI()))
		return 1
	})
	luaRegister(l, "projcontacttime", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.projContactTime(
			BytecodeInt(int32(numArg(l, 1)))).ToI()))
		return 1
	})
	luaRegister(l, "projguardedtime", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.projGuardedTime(
			BytecodeInt(int32(numArg(l, 1)))).ToI()))
		return 1
	})
	luaRegister(l, "projhittime", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.projHitTime(
			BytecodeInt(int32(numArg(l, 1)))).ToI()))
		return 1
	})
	luaRegister(l, "rightedge", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.rightEdge()))
		return 1
	})
	luaRegister(l, "roundno", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.round))
		return 1
	})
	luaRegister(l, "roundsexisted", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.roundsExisted()))
		return 1
	})
	luaRegister(l, "roundstate", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.roundState()))
		return 1
	})
	luaRegister(l, "screenheight", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.screenHeight()))
		return 1
	})
	luaRegister(l, "screenposX", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.screenPosX()))
		return 1
	})
	luaRegister(l, "screenposY", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.screenPosY()))
		return 1
	})
	luaRegister(l, "screenwidth", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.screenWidth()))
		return 1
	})
	luaRegister(l, "selfanimexist", func(*lua.LState) int {
		l.Push(lua.LBool(sys.debugWC.selfAnimExist(
			BytecodeInt(int32(numArg(l, 1)))).ToB()))
		return 1
	})
	luaRegister(l, "stateno", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.ss.no))
		return 1
	})
	luaRegister(l, "stateOwner", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.ss.sb.playerNo + 1))
		return 1
	})
	luaRegister(l, "statetype", func(*lua.LState) int {
		var s string
		switch sys.debugWC.ss.stateType {
		case ST_S:
			s = "S"
		case ST_C:
			s = "C"
		case ST_A:
			s = "A"
		case ST_L:
			s = "L"
		}
		l.Push(lua.LString(s))
		return 1
	})
	luaRegister(l, "stagevar", func(*lua.LState) int {
		var s string
		switch strArg(l, 1) {
		case "info.name":
			s = sys.stage.name
		case "info.displayname":
			s = sys.stage.displayname
		case "info.author":
			s = sys.stage.author
		}
		l.Push(lua.LString(s))
		return 1
	})
	luaRegister(l, "sysfvar", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.sysFvarGet(int32(numArg(l, 1))).ToF()))
		return 1
	})
	luaRegister(l, "sysvar", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.sysVarGet(int32(numArg(l, 1))).ToI()))
		return 1
	})
	luaRegister(l, "teammode", func(*lua.LState) int {
		var s string
		switch sys.tmode[sys.debugWC.playerNo&1] {
		case TM_Single:
			s = "single"
		case TM_Simul:
			s = "simul"
		case TM_Turns:
			s = "turns"
		}
		l.Push(lua.LString(s))
		return 1
	})
	luaRegister(l, "teamside", func(*lua.LState) int {
		l.Push(lua.LNumber(int32(sys.debugWC.playerNo)&1 + 1))
		return 1
	})
	luaRegister(l, "tickspersecond", func(*lua.LState) int {
		l.Push(lua.LNumber(FPS))
		return 1
	})
	luaRegister(l, "time", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.ss.time))
		return 1
	})
	luaRegister(l, "topedge", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.topEdge()))
		return 1
	})
	luaRegister(l, "uniqhitcount", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.uniqHitCount))
		return 1
	})
	luaRegister(l, "var", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.varGet(int32(numArg(l, 1))).ToI()))
		return 1
	})
	luaRegister(l, "velX", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.vel[0]))
		return 1
	})
	luaRegister(l, "velY", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.vel[1]))
		return 1
	})
	luaRegister(l, "win", func(*lua.LState) int {
		l.Push(lua.LBool(sys.debugWC.win()))
		return 1
	})
	luaRegister(l, "winko", func(*lua.LState) int {
		l.Push(lua.LBool(sys.debugWC.winKO()))
		return 1
	})
	luaRegister(l, "wintime", func(*lua.LState) int {
		l.Push(lua.LBool(sys.debugWC.winTime()))
		return 1
	})
	luaRegister(l, "winperfect", func(*lua.LState) int {
		l.Push(lua.LBool(sys.debugWC.winPerfect()))
		return 1
	})
}

// Debug Script

func debugScriptInit(l *lua.LState, file string) error {
	scriptCommonInit(l)
	triggerScriptInit(l)
	luaRegister(l, "puts", func(*lua.LState) int {
		fmt.Println(strArg(l, 1))
		return 0
	})
	luaRegister(l, "setLife", func(*lua.LState) int {
		if sys.netInput == nil && sys.fileInput == nil {
			sys.debugWC.lifeSet(int32(numArg(l, 1)))
		}
		return 0
	})
	luaRegister(l, "setPower", func(*lua.LState) int {
		if sys.netInput == nil && sys.fileInput == nil {
			sys.debugWC.setPower(int32(numArg(l, 1)))
		}
		return 0
	})
	luaRegister(l, "selfState", func(*lua.LState) int {
		if sys.netInput == nil && sys.fileInput == nil {
			sys.debugWC.selfState(int32(numArg(l, 1)), -1, 1)
		}
		return 0
	})
	luaRegister(l, "addHotkey", func(*lua.LState) int {
		l.Push(lua.LBool(func() bool {
			k := StringToKey(strArg(l, 1))
			if k == glfw.KeyUnknown {
				return false
			}
			sk := *NewShortcutKey(k, boolArg(l, 2), boolArg(l, 3), boolArg(l, 4))
			scr := strArg(l, 5)
			sys.shortcutScripts[sk] = &ShortcutScript{Script: scr}
			return true
		}()))
		return 1
	})
	luaRegister(l, "toggleClsnDraw", func(*lua.LState) int {
		sys.clsnDraw = !sys.clsnDraw
		return 0
	})
	luaRegister(l, "toggleDebugDraw", func(*lua.LState) int {
		sys.debugDraw = !sys.debugDraw
		return 0
	})
	luaRegister(l, "togglePause", func(*lua.LState) int {
		if sys.netInput == nil {
			sys.paused = !sys.paused
		}
		return 0
	})
	luaRegister(l, "step", func(*lua.LState) int {
		sys.step = true
		return 0
	})
	luaRegister(l, "toggleStatusDraw", func(*lua.LState) int {
		sys.statusDraw = !sys.statusDraw
		return 0
	})
	luaRegister(l, "roundReset", func(*lua.LState) int {
		if sys.netInput == nil && sys.fileInput == nil {
			sys.roundResetFlg = true
		}
		return 0
	})
	luaRegister(l, "reload", func(*lua.LState) int {
		if sys.netInput == nil && sys.fileInput == nil {
			sys.reloadFlg = true
		}
		return 0
	})
	luaRegister(l, "setAccel", func(*lua.LState) int {
		if sys.netInput == nil && sys.fileInput == nil {
			sys.accel = float32(numArg(l, 1))
		}
		return 0
	})
	luaRegister(l, "setAILevel", func(*lua.LState) int {
		if sys.netInput == nil && sys.fileInput == nil {
			level := int32(numArg(l, 1))
			sys.com[sys.debugWC.playerNo] = level
			for _, c := range sys.chars[sys.debugWC.playerNo] {
				if level == 0 {
					c.key = sys.debugWC.playerNo
				} else {
					c.key = ^sys.debugWC.playerNo
				}
			}
		}
		return 0
	})
	luaRegister(l, "setTime", func(*lua.LState) int {
		if sys.netInput == nil && sys.fileInput == nil {
			sys.time = int32(numArg(l, 1))
		}
		return 0
	})
	luaRegister(l, "clear", func(*lua.LState) int {
		for i := range sys.clipboardText {
			sys.clipboardText[i] = nil
		}
		return 0
	})
	luaRegister(l, "getAllowDebugKeys", func(*lua.LState) int {
		l.Push(lua.LBool(sys.allowDebugKeys))
		return 1
	})
	return l.DoFile(file)
}
