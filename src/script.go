package main

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	lua "github.com/yuin/gopher-lua"
)

// Data handlers
func luaRegister(l *lua.LState, name string, f func(*lua.LState) int) {
	l.Register(name, f)
}
func strArg(l *lua.LState, argi int) string {
	if !lua.LVCanConvToString(l.Get(argi)) {
		l.RaiseError("\nArgument %v is not a string: %v\n", argi, l.Get(argi))
	}
	return l.ToString(argi)
}
func numArg(l *lua.LState, argi int) float64 {
	num, ok := l.Get(argi).(lua.LNumber)
	if !ok {
		l.RaiseError("\nArgument %v is not a number: %v\n", argi, l.Get(argi))
	}
	return float64(num)
}
func boolArg(l *lua.LState, argi int) bool {
	return l.ToBool(argi)
}
func tableArg(l *lua.LState, argi int) *lua.LTable {
	return l.ToTable(argi)
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
	l.RaiseError("\nArgument %v is not a userdata of type: %T\n", argi, udtype)
}

// -------------------------------------------------------------------------------------------------
// Register external functions to be called from Lua scripts
func systemScriptInit(l *lua.LState) {
	triggerFunctions(l)
	luaRegister(l, "addChar", func(l *lua.LState) int {
		for _, c := range strings.Split(strings.TrimSpace(strArg(l, 1)), "\n") {
			c = strings.Trim(c, "\r")
			if len(c) > 0 {
				sys.sel.addChar(c)
			}
		}
		return 0
	})
	luaRegister(l, "addHotkey", func(*lua.LState) int {
		l.Push(lua.LBool(func() bool {
			k := StringToKey(strArg(l, 1))
			if k == KeyUnknown {
				return false
			}
			sk := *NewShortcutKey(k, boolArg(l, 2), boolArg(l, 3), boolArg(l, 4))
			sys.shortcutScripts[sk] = &ShortcutScript{Pause: boolArg(l, 5), DebugKey: boolArg(l, 6), Script: strArg(l, 7)}
			return true
		}()))
		return 1
	})
	luaRegister(l, "addStage", func(l *lua.LState) int {
		var n int
		for _, c := range SplitAndTrim(strings.TrimSpace(strArg(l, 1)), "\n") {
			if err := sys.sel.AddStage(c); err == nil {
				n++
			}
		}
		l.Push(lua.LNumber(n))
		return 1
	})
	luaRegister(l, "animAddPos", func(*lua.LState) int {
		a, ok := toUserData(l, 1).(*Anim)
		if !ok {
			userDataError(l, 1, a)
		}
		a.AddPos(float32(numArg(l, 2))/sys.luaSpriteScale, float32(numArg(l, 3))/sys.luaSpriteScale)
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
	luaRegister(l, "animGetLength", func(*lua.LState) int {
		a, ok := toUserData(l, 1).(*Anim)
		if !ok {
			userDataError(l, 1, a)
		}
		var sum int32
		for _, f := range a.anim.frames {
			if f.Time == -1 {
				sum += 1
			} else {
				sum += f.Time
			}
		}
		l.Push(lua.LNumber(sum))
		l.Push(lua.LNumber(a.anim.totaltime))
		return 2
	})
	luaRegister(l, "animGetPreloadedData", func(l *lua.LState) int {
		var anim *Animation
		if strArg(l, 1) == "char" {
			anim = sys.sel.GetChar(int(numArg(l, 2))).anims.get(int16(numArg(l, 3)), int16(numArg(l, 4)))
		} else if strArg(l, 1) == "stage" {
			anim = sys.sel.GetStage(int(numArg(l, 2))).anims.get(int16(numArg(l, 3)), int16(numArg(l, 4)))
		}
		if anim != nil {
			pfx := newPalFX()
			pfx.clear()
			pfx.time = -1
			//TODO: palette changing depending on palette currently loaded on character
			a := &Anim{anim: anim, window: sys.scrrect, xscl: 1, yscl: 1, palfx: pfx}
			if l.GetTop() >= 5 && !boolArg(l, 5) && a.anim.totaltime == a.anim.looptime {
				a.anim.totaltime = -1
				a.anim.looptime = 0
			}
			l.Push(newUserData(l, a))
			return 1
		}
		return 0
	})
	luaRegister(l, "animGetSpriteInfo", func(*lua.LState) int {
		a, ok := toUserData(l, 1).(*Anim)
		if !ok {
			userDataError(l, 1, a)
		}
		if len(a.anim.frames) == 0 {
			return 0
		}
		var spr *Sprite
		if l.GetTop() >= 3 {
			spr = a.anim.sff.GetSprite(int16(numArg(l, 2)), int16(numArg(l, 3)))
		} else {
			spr = a.anim.spr
		}
		if spr == nil {
			return 0
		}
		tbl := l.NewTable()
		tbl.RawSetString("Group", lua.LNumber(spr.Group))
		tbl.RawSetString("Number", lua.LNumber(spr.Number))
		subt := l.NewTable()
		for k, v := range spr.Size {
			subt.RawSetInt(k+1, lua.LNumber(v))
		}
		tbl.RawSetString("Size", subt)
		subt = l.NewTable()
		for k, v := range spr.Offset {
			subt.RawSetInt(k+1, lua.LNumber(v))
		}
		tbl.RawSetString("Offset", subt)
		tbl.RawSetString("palidx", lua.LNumber(spr.palidx))
		l.Push(tbl)
		return 1
	})
	luaRegister(l, "animNew", func(*lua.LState) int {
		s, ok := toUserData(l, 1).(*Sff)
		if !ok {
			userDataError(l, 1, s)
		}
		act := strArg(l, 2)
		anim := NewAnim(s, act)
		if anim == nil {
			l.RaiseError("\nFailed to read the data: %v\n", act)
		}
		l.Push(newUserData(l, anim))
		return 1
	})
	luaRegister(l, "animReset", func(*lua.LState) int {
		a, ok := toUserData(l, 1).(*Anim)
		if !ok {
			userDataError(l, 1, a)
		}
		a.ResetFrames()
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
	luaRegister(l, "animSetColorKey", func(*lua.LState) int {
		a, ok := toUserData(l, 1).(*Anim)
		if !ok {
			userDataError(l, 1, a)
		}
		a.SetColorKey(int16(numArg(l, 2)))
		return 0
	})
	luaRegister(l, "animSetFacing", func(*lua.LState) int {
		a, ok := toUserData(l, 1).(*Anim)
		if !ok {
			userDataError(l, 1, a)
		}
		a.SetFacing(float32(numArg(l, 2)))
		return 0
	})
	luaRegister(l, "animSetPalFX", func(*lua.LState) int {
		a, ok := toUserData(l, 1).(*Anim)
		if !ok {
			userDataError(l, 1, a)
		}
		tableArg(l, 2).ForEach(func(key, value lua.LValue) {
			switch k := key.(type) {
			case lua.LString:
				switch string(k) {
				case "time":
					a.palfx.time = int32(lua.LVAsNumber(value))
				case "add":
					switch v := value.(type) {
					case *lua.LTable:
						v.ForEach(func(key2, value2 lua.LValue) {
							a.palfx.add[int(lua.LVAsNumber(key2))-1] = int32(lua.LVAsNumber(value2))
						})
					}
				case "mul":
					switch v := value.(type) {
					case *lua.LTable:
						v.ForEach(func(key2, value2 lua.LValue) {
							a.palfx.mul[int(lua.LVAsNumber(key2))-1] = int32(lua.LVAsNumber(value2))
						})
					}
				case "sinadd":
					var s [4]int32
					switch v := value.(type) {
					case *lua.LTable:
						v.ForEach(func(key2, value2 lua.LValue) {
							s[int(lua.LVAsNumber(key2))-1] = int32(lua.LVAsNumber(value2))
						})
					}
					if s[3] < 0 {
						a.palfx.sinadd[0] = -s[0]
						a.palfx.sinadd[1] = -s[1]
						a.palfx.sinadd[2] = -s[2]
						a.palfx.cycletime = -s[3]
					} else {
						a.palfx.sinadd[0] = s[0]
						a.palfx.sinadd[1] = s[1]
						a.palfx.sinadd[2] = s[2]
						a.palfx.cycletime = s[3]
					}
				case "invertall":
					a.palfx.invertall = lua.LVAsNumber(value) == 1
				case "color":
					a.palfx.color = float32(lua.LVAsNumber(value)) / 256
				default:
					l.RaiseError("\nInvalid table key: %v\n", k)
				}
			default:
				l.RaiseError("\nInvalid table key type: %v\n", fmt.Sprintf("%T\n", key))
			}
		})
		return 0
	})
	luaRegister(l, "animSetPos", func(*lua.LState) int {
		a, ok := toUserData(l, 1).(*Anim)
		if !ok {
			userDataError(l, 1, a)
		}
		a.SetPos(float32(numArg(l, 2))/sys.luaSpriteScale+sys.luaSpriteOffsetX, float32(numArg(l, 3))/sys.luaSpriteScale)
		return 0
	})
	luaRegister(l, "animSetScale", func(*lua.LState) int {
		a, ok := toUserData(l, 1).(*Anim)
		if !ok {
			userDataError(l, 1, a)
		}
		if l.GetTop() < 4 || boolArg(l, 4) {
			a.SetScale(float32(numArg(l, 2))/sys.luaSpriteScale, float32(numArg(l, 3))/sys.luaSpriteScale)
		} else {
			a.SetScale(float32(numArg(l, 2)), float32(numArg(l, 3)))
		}
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
	luaRegister(l, "animSetWindow", func(*lua.LState) int {
		a, ok := toUserData(l, 1).(*Anim)
		if !ok {
			userDataError(l, 1, a)
		}
		a.SetWindow(float32(numArg(l, 2))/sys.luaSpriteScale+sys.luaSpriteOffsetX, float32(numArg(l, 3))/sys.luaSpriteScale,
			float32(numArg(l, 4))/sys.luaSpriteScale, float32(numArg(l, 5))/sys.luaSpriteScale)
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
	luaRegister(l, "bgDraw", func(*lua.LState) int {
		bg, ok := toUserData(l, 1).(*BGDef)
		if !ok {
			userDataError(l, 1, bg)
		}
		top := false
		var x, y, scl float32 = 0, 0, 1
		if l.GetTop() >= 2 {
			top = boolArg(l, 2)
		}
		if l.GetTop() >= 3 {
			x = float32(numArg(l, 3))
		}
		if l.GetTop() >= 4 {
			y = float32(numArg(l, 4))
		}
		if l.GetTop() >= 5 {
			scl = float32(numArg(l, 5))
		}
		bg.draw(top, x, y, scl)
		return 0
	})
	luaRegister(l, "bgNew", func(*lua.LState) int {
		s, ok := toUserData(l, 1).(*Sff)
		if !ok {
			userDataError(l, 1, s)
		}
		bg, err := loadBGDef(s, strArg(l, 2), strArg(l, 3))
		if err != nil {
			l.RaiseError("\nCan't load %v (%v): %v\n", strArg(l, 3), strArg(l, 2), err.Error())
		}
		l.Push(newUserData(l, bg))
		return 1
	})
	luaRegister(l, "bgReset", func(*lua.LState) int {
		bg, ok := toUserData(l, 1).(*BGDef)
		if !ok {
			userDataError(l, 1, bg)
		}
		bg.reset()
		return 0
	})
	luaRegister(l, "charChangeAnim", func(l *lua.LState) int {
		//pn, anim_no, anim_elem, ffx
		pn := int(numArg(l, 1))
		an := int32(numArg(l, 2))
		if pn >= 1 && pn <= len(sys.chars) && len(sys.chars[pn-1]) > 0 {
			c := sys.chars[pn-1]
			if c[0].selfAnimExist(BytecodeInt(an)) == BytecodeBool(true) {
				ffx := false
				if l.GetTop() >= 4 {
					ffx = boolArg(l, 4)
				}
				c[0].changeAnim(an, ffx)
				if l.GetTop() >= 3 {
					c[0].setAnimElem(int32(numArg(l, 3)))
				}
				l.Push(lua.LBool(true))
				return 1
			}
		}
		l.Push(lua.LBool(false))
		return 1
	})
	luaRegister(l, "charChangeState", func(l *lua.LState) int {
		//pn, state_no
		pn := int(numArg(l, 1))
		st := int32(numArg(l, 2))
		if pn >= 1 && pn <= len(sys.chars) && len(sys.chars[pn-1]) > 0 {
			c := sys.chars[pn-1]
			if st == -1 {
				for _, ch := range c {
					ch.setSCF(SCF_disabled)
				}
			} else if c[0].selfStatenoExist(BytecodeInt(st)) == BytecodeBool(true) {
				for _, ch := range c {
					if ch.scf(SCF_disabled) {
						ch.unsetSCF(SCF_disabled)
					}
				}
				c[0].changeState(st, -1, -1, false)
				l.Push(lua.LBool(true))
				return 1
			}
		}
		l.Push(lua.LBool(false))
		return 1
	})
	luaRegister(l, "charMapSet", func(*lua.LState) int {
		//pn, map_name, value, map_type
		pn := int(numArg(l, 1))
		var scType int32
		if l.GetTop() >= 4 && strArg(l, 4) == "add" {
			scType = 1
		}
		if pn >= 1 && pn <= len(sys.chars) && len(sys.chars[pn-1]) > 0 {
			sys.chars[pn-1][0].mapSet(strArg(l, 2), float32(numArg(l, 3)), scType)
		}
		return 0
	})
	luaRegister(l, "charSndPlay", func(l *lua.LState) int {
		//pn, group_no, sound_no, volumescale, commonSnd, channel, lowpriority, freqmul, loop, pan
		pn := int(numArg(l, 1))
		if pn < 1 || pn > len(sys.chars) || len(sys.chars[pn-1]) == 0 {
			l.RaiseError("\nPlayer not found: %v\n", pn)
		}
		f, lw, lp := false, false, false
		var g, n, ch, vo, priority int32 = -1, 0, -1, 100, 0
		var p, fr float32 = 0, 1
		x := &sys.chars[pn-1][0].pos[0]
		ls := sys.chars[pn-1][0].localscl
		if l.GetTop() >= 2 {
			g = int32(numArg(l, 2))
		}
		if l.GetTop() >= 3 {
			n = int32(numArg(l, 3))
		}
		if l.GetTop() >= 4 {
			vo = int32(numArg(l, 4))
		}
		if l.GetTop() >= 5 {
			f = boolArg(l, 5)
		}
		if l.GetTop() >= 6 {
			ch = int32(numArg(l, 6))
		}
		if l.GetTop() >= 7 {
			lw = boolArg(l, 7)
		}
		if l.GetTop() >= 8 {
			fr = float32(numArg(l, 8))
		}
		if l.GetTop() >= 9 {
			lp = boolArg(l, 9)
		}
		if l.GetTop() >= 10 {
			p = float32(numArg(l, 10))
		}
		if l.GetTop() >= 11 {
			priority = int32(numArg(l, 11))
		}
		sys.chars[pn-1][0].playSound(f, lw, lp, g, n, ch, vo, p, fr, ls, x, false, priority)
		return 0
	})
	luaRegister(l, "charSndStop", func(l *lua.LState) int {
		if l.GetTop() == 0 {
			sys.stopAllSound()
			return 0
		}
		pn := int(numArg(l, 1))
		if pn < 1 || pn > len(sys.chars) || len(sys.chars[pn-1]) == 0 {
			l.RaiseError("\nPlayer not found: %v\n", pn)
		}
		sys.chars[pn-1][0].soundChannels.SetSize(0)
		return 0
	})
	luaRegister(l, "charSpriteDraw", func(l *lua.LState) int {
		//pn, spr_tbl (1 or more pairs), x, y, scaleX, scaleY, facing, window
		pn := int(numArg(l, 1))
		if pn < 1 || pn > len(sys.chars) || len(sys.chars[pn-1]) == 0 {
			l.RaiseError("\nPlayer not found: %v\n", pn)
		}
		window := &sys.scrrect
		if l.GetTop() >= 11 {
			window = &[...]int32{int32(numArg(l, 8)), int32(numArg(l, 9)), int32(numArg(l, 10)), int32(numArg(l, 11))}
		}
		var ok bool
		var group int16
		tableArg(l, 2).ForEach(func(key, value lua.LValue) {
			if !ok {
				if int(lua.LVAsNumber(key))%2 == 1 {
					group = int16(lua.LVAsNumber(value))
				} else {
					sprite := sys.cgi[pn-1].sff.getOwnPalSprite(group, int16(lua.LVAsNumber(value)))
					if fspr := sprite; fspr != nil {
						pfx := sys.chars[pn-1][0].getPalfx()
						sys.cgi[pn-1].sff.palList.SwapPalMap(&pfx.remap)
						fspr.Pal = nil
						fspr.Pal = fspr.GetPal(&sys.cgi[pn-1].sff.palList)
						sys.cgi[pn-1].sff.palList.SwapPalMap(&pfx.remap)
						x := (float32(numArg(l, 3)) + sys.lifebarOffsetX) * sys.lifebarScale
						y := float32(numArg(l, 4)) * sys.lifebarScale
						scale := [...]float32{float32(numArg(l, 5)), float32(numArg(l, 6))}
						facing := int8(numArg(l, 7))
						fscale := sys.chars[pn-1][0].localscl
						if sprite.coldepth <= 8 && sprite.PalTex == nil {
							sprite.CachePalette(sprite.Pal)
						}
						sprite.Draw(x, y, scale[0]*float32(facing)*fscale, scale[1]*fscale,
							0, pfx, window)
						ok = true
					}
				}
			}
		})
		l.Push(lua.LBool(ok))
		return 1
	})
	luaRegister(l, "clear", func(*lua.LState) int {
		for _, p := range sys.chars {
			for _, c := range p {
				//for i := range c.clipboardText {
				//	c.clipboardText[i] = nil
				//}
				c.clipboardText = nil
			}
		}
		return 0
	})
	luaRegister(l, "clearAllSound", func(l *lua.LState) int {
		sys.clearAllSound()
		return 0
	})
	luaRegister(l, "clearColor", func(l *lua.LState) int {
		a := int32(255)
		if l.GetTop() >= 4 {
			a = int32(numArg(l, 4))
		}
		col := uint32(int32(numArg(l, 3))&0xff | int32(numArg(l, 2))&0xff<<8 |
			int32(numArg(l, 1))&0xff<<16)
		FillRect(sys.scrrect, col, a)
		return 0
	})
	luaRegister(l, "clearConsole", func(*lua.LState) int {
		sys.consoleText = nil
		return 0
	})
	luaRegister(l, "clearSelected", func(l *lua.LState) int {
		sys.sel.ClearSelected()
		return 0
	})
	luaRegister(l, "commandAdd", func(l *lua.LState) int {
		cl, ok := toUserData(l, 1).(*CommandList)
		if !ok {
			userDataError(l, 1, cl)
		}
		cm, err := ReadCommand(strArg(l, 2), strArg(l, 3), NewCommandKeyRemap())
		if err != nil {
			l.RaiseError(err.Error())
		}
		time, buftime := cl.DefaultTime, cl.DefaultBufferTime
		if l.GetTop() >= 4 {
			time = int32(numArg(l, 4))
		}
		if l.GetTop() >= 5 {
			buftime = Max(1, int32(numArg(l, 5)))
		}
		cm.time = time
		cm.buftime = buftime
		cl.Add(*cm)
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
		if cl.Input(int(numArg(l, 2))-1, 1, 0, 0) {
			cl.Step(1, false, false, 0)
		}
		return 0
	})
	luaRegister(l, "commandNew", func(l *lua.LState) int {
		l.Push(newUserData(l, NewCommandList(NewCommandBuffer())))
		return 1
	})
	luaRegister(l, "commonLuaInsert", func(l *lua.LState) int {
		sys.commonLua = append(sys.commonLua, strArg(l, 1))
		return 0
	})
	luaRegister(l, "commonLuaDelete", func(l *lua.LState) int {
		for k, v := range sys.commonLua {
			if v == strArg(l, 1) {
				// shift left one index
				copy(sys.commonLua[k:], sys.commonLua[k+1:])
				// erase last element (write zero value)
				sys.commonLua[len(sys.commonLua)-1] = ""
				// truncate slice
				sys.commonLua = sys.commonLua[:len(sys.commonLua)-1]
				break
			}
		}
		return 0
	})
	luaRegister(l, "commonStatesInsert", func(l *lua.LState) int {
		sys.commonStates = append(sys.commonStates, strArg(l, 1))
		return 0
	})
	luaRegister(l, "commonStatesDelete", func(l *lua.LState) int {
		for k, v := range sys.commonStates {
			if v == strArg(l, 1) {
				copy(sys.commonStates[k:], sys.commonStates[k+1:])
				sys.commonStates[len(sys.commonStates)-1] = ""
				sys.commonStates = sys.commonStates[:len(sys.commonStates)-1]
				break
			}
		}
		return 0
	})
	luaRegister(l, "connected", func(*lua.LState) int {
		l.Push(lua.LBool(sys.netInput.IsConnected()))
		return 1
	})
	luaRegister(l, "dialogueReset", func(*lua.LState) int {
		for _, p := range sys.chars {
			if len(p) > 0 {
				p[0].dialogue = nil
			}
		}
		sys.dialogueFlg = false
		sys.dialogueForce = 0
		sys.dialogueBarsFlg = false
		return 0
	})
	luaRegister(l, "endMatch", func(*lua.LState) int {
		sys.endMatch = true
		return 0
	})
	luaRegister(l, "enterNetPlay", func(*lua.LState) int {
		if sys.netInput != nil {
			l.RaiseError("\nConnection already established.\n")
		}
		sys.chars = [len(sys.chars)][]*Char{}
		sys.netInput = NewNetInput()
		if host := strArg(l, 1); host != "" {
			sys.netInput.Connect(host, sys.listenPort)
		} else {
			if err := sys.netInput.Accept(sys.listenPort); err != nil {
				l.RaiseError(err.Error())
			}
		}
		return 0
	})
	luaRegister(l, "enterReplay", func(*lua.LState) int {
		if sys.vRetrace >= 0 {
			sys.window.SetSwapInterval(1) //broken frame skipping when set to 0
		}
		sys.chars = [len(sys.chars)][]*Char{}
		sys.fileInput = OpenFileInput(strArg(l, 1))
		return 0
	})
	luaRegister(l, "esc", func(l *lua.LState) int {
		if l.GetTop() >= 1 {
			sys.esc = boolArg(l, 1)
		}
		l.Push(lua.LBool(sys.esc))
		return 1
	})
	luaRegister(l, "exitNetPlay", func(*lua.LState) int {
		if sys.netInput != nil {
			sys.netInput.Close()
			sys.netInput = nil
		}
		return 0
	})
	luaRegister(l, "exitReplay", func(*lua.LState) int {
		if sys.vRetrace >= 0 {
			sys.window.SetSwapInterval(sys.vRetrace)
		}
		if sys.fileInput != nil {
			sys.fileInput.Close()
			sys.fileInput = nil
		}
		return 0
	})
	luaRegister(l, "fade", func(l *lua.LState) int {
		rect := [4]int32{int32(numArg(l, 1)), int32(numArg(l, 2)), int32(numArg(l, 3)), int32(numArg(l, 4))}
		alpha := int32(numArg(l, 5))
		FillRect(rect, 0, alpha>>uint(Btoi(sys.clsnDraw))+Btoi(sys.clsnDraw)*128)
		return 0
	})
	luaRegister(l, "fadeColor", func(l *lua.LState) int {
		if int32(numArg(l, 2)) > sys.frameCounter {
			l.Push(lua.LBool(true)) //delayed fade
			return 1
		}
		frame := float64(sys.frameCounter - int32(numArg(l, 2)))
		length := numArg(l, 3)
		if frame > length || length <= 0 {
			l.Push(lua.LBool(false))
			return 1
		}
		r, g, b, a := int32(0), int32(0), int32(0), float64(0)
		if strArg(l, 1) == "fadeout" {
			a = math.Floor(float64(255) / length * frame)
		} else if strArg(l, 1) == "fadein" {
			a = math.Floor(255 - 255*(frame-1)/length)
		}
		a = float64(ClampF(float32(a), 0, 255))
		if l.GetTop() >= 6 {
			r = int32(numArg(l, 4))
			g = int32(numArg(l, 5))
			b = int32(numArg(l, 6))
		}
		col := uint32(int32(b)&0xff | int32(g)&0xff<<8 | int32(r)&0xff<<16)
		FillRect(sys.scrrect, col, int32(a))
		l.Push(lua.LBool(true))
		return 1
	})
	luaRegister(l, "fillRect", func(l *lua.LState) int {
		rect := [4]int32{int32((float32(numArg(l, 1))/sys.luaSpriteScale + float32(sys.gameWidth-320)/2 + sys.luaSpriteOffsetX) * sys.widthScale),
			int32((float32(numArg(l, 2))/sys.luaSpriteScale + float32(sys.gameHeight-240)) * sys.heightScale),
			int32((float32(numArg(l, 3)) / sys.luaSpriteScale) * sys.widthScale),
			int32((float32(numArg(l, 4)) / sys.luaSpriteScale) * sys.heightScale)}
		col := uint32(int32(numArg(l, 7))&0xff | int32(numArg(l, 6))&0xff<<8 | int32(numArg(l, 5))&0xff<<16)
		a := int32(int32(numArg(l, 8))&0xff | int32(numArg(l, 9))&0xff<<10)
		FillRect(rect, col, a)
		return 0
	})
	luaRegister(l, "fontGetDef", func(l *lua.LState) int {
		fnt, ok := toUserData(l, 1).(*Fnt)
		if !ok {
			userDataError(l, 1, fnt)
		}
		tbl := l.NewTable()
		tbl.RawSetString("Type", lua.LString(fnt.Type))
		subt := l.NewTable()
		subt.Append(lua.LNumber(fnt.Size[0]))
		subt.Append(lua.LNumber(fnt.Size[1]))
		tbl.RawSetString("Size", subt)
		subt = l.NewTable()
		subt.Append(lua.LNumber(fnt.Spacing[0]))
		subt.Append(lua.LNumber(fnt.Spacing[1]))
		tbl.RawSetString("Spacing", subt)
		subt = l.NewTable()
		subt.Append(lua.LNumber(fnt.offset[0]))
		subt.Append(lua.LNumber(fnt.offset[1]))
		tbl.RawSetString("offset", subt)
		l.Push(tbl)
		return 1
	})
	luaRegister(l, "fontGetTextWidth", func(*lua.LState) int {
		fnt, ok := toUserData(l, 1).(*Fnt)
		if !ok {
			userDataError(l, 1, fnt)
		}
		var bank int32
		if l.GetTop() >= 3 {
			bank = int32(numArg(l, 3))
		}
		l.Push(lua.LNumber(fnt.TextWidth(strArg(l, 2), bank)))
		return 1
	})
	luaRegister(l, "fontNew", func(l *lua.LState) int {
		var height int32 = -1
		if l.GetTop() >= 2 {
			height = int32(numArg(l, 2))
		}
		filename := SearchFile(strArg(l, 1), []string{"font/", sys.motifDir, "", "data/"})
		fnt, err := loadFnt(filename, height)
		if err != nil {
			sys.errLog.Printf("failed to load %v (screenpack font): %v", filename, err)
			fnt = newFnt()
		}
		l.Push(newUserData(l, fnt))
		return 1
	})
	// Execute a match of gameplay
	luaRegister(l, "game", func(l *lua.LState) int {
		// Anonymous function to load characters and stages, and/or wait for them to finish loading
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
			sys.debugRef = [2]int{}
			sys.roundsExisted = [2]int32{}
			sys.matchWins = [2]int32{}

			// Reset lifebars
			for i := range sys.lifebar.wi {
				sys.lifebar.wi[i].clear()
			}

			sys.draws = 0
			tbl := l.NewTable()
			sys.matchData = l.NewTable()

			// Anonymous function to perform gameplay
			fight := func() (int32, error) {
				// Load characters and stage
				if err := load(); err != nil {
					return -1, err
				}
				if sys.loader.state == LS_Cancel {
					return -1, nil
				}

				// Reset and setup characters
				if sys.round == 1 {
					sys.charList.clear()
				}
				nextId := sys.helperMax
				for i := 0; i < MaxSimul*2; i += 2 {
					if len(sys.chars[i]) > 0 {
						if sys.round == 1 {
							sys.chars[i][0].id = sys.newCharId()
						} else if sys.chars[i][0].roundsExisted() == 0 {
							sys.chars[i][0].id = nextId
						}
						nextId++
					}
				}
				for i := 1; i < MaxSimul*2; i += 2 {
					if len(sys.chars[i]) > 0 {
						if sys.round == 1 {
							sys.chars[i][0].id = sys.newCharId()
						} else if sys.chars[i][0].roundsExisted() == 0 {
							sys.chars[i][0].id = nextId
						}
						nextId++
					}
				}
				for i := MaxSimul * 2; i < MaxSimul*2+MaxAttachedChar; i += 1 {
					if len(sys.chars[i]) > 0 {
						if sys.round == 1 {
							sys.chars[i][0].id = sys.newCharId()
						} else if sys.chars[i][0].roundsExisted() == 0 {
							sys.chars[i][0].id = nextId
						}
						nextId++
					}
				}
				for i, c := range sys.chars {
					if len(c) > 0 {
						p[i] = c[0]
						if sys.round == 1 {
							sys.charList.add(c[0])
						} else if c[0].roundsExisted() == 0 {
							if !sys.charList.replace(c[0], i, 0) {
								panic(fmt.Errorf("failed to replace player: %v", i))
							}
						}
						if c[0].roundsExisted() == 0 {
							c[0].loadPalette()
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

				// If first round
				if sys.round == 1 {
					// Update wins, reset stage
					sys.endMatch = false
					if sys.tmode[1] == TM_Turns {
						sys.matchWins[0] = sys.numTurns[1]
					} else {
						sys.matchWins[0] = sys.lifebar.ro.match_wins[1]
					}
					if sys.tmode[0] == TM_Turns {
						sys.matchWins[1] = sys.numTurns[0]
					} else {
						sys.matchWins[1] = sys.lifebar.ro.match_wins[0]
					}
					sys.teamLeader = [...]int{0, 1}
					sys.stage.reset()
				}

				// Winning player index
				// -1 on quit, -2 on restarting match
				winp := int32(0)

				//fight loop
				if sys.fight() {
					// Match is restarting
					for i, b := range sys.reloadCharSlot {
						if b {
							sys.chars[i] = []*Char{}
							b = false
						}
					}
					if sys.reloadStageFlg {
						sys.stage = nil
					}
					if sys.reloadLifebarFlg {
						sys.lifebar.reloadLifebar()
					}
					sys.loaderReset()
					winp = -2
				} else if sys.esc {
					// Match was quit
					winp = -1
				} else {
					// Determine winner
					w1 := sys.wins[0] >= sys.matchWins[0]
					w2 := sys.wins[1] >= sys.matchWins[1]
					if w1 != w2 {
						winp = Btoi(w1) + Btoi(w2)*2
					}
				}
				return winp, nil
			}

			// Reset net inputs
			if sys.netInput != nil {
				sys.netInput.Stop()
			}

			// Defer synchronizing with external inputs on return
			defer sys.synchronize()

			// Loop calling gameplay until match ends
			// Will repeat on turns mode character change and hard reset
			for {
				var err error
				// Call gameplay anonymous function
				if winp, err = fight(); err != nil {
					l.RaiseError(err.Error())
				}
				// If a team won, and not going to the next character in turns mode, break
				if winp < 0 || sys.tmode[0] != TM_Turns && sys.tmode[1] != TM_Turns ||
					sys.wins[0] >= sys.matchWins[0] || sys.wins[1] >= sys.matchWins[1] ||
					sys.gameEnd {
					break
				}
				// Reset roundsExisted to 0 if the losing side is on turns mode
				for i := 0; i < 2; i++ {
					if !p[i].win() && sys.tmode[i] == TM_Turns {
						sys.lifebar.fa[TM_Turns][i].numko++
						sys.lifebar.nm[TM_Turns][i].numko++
						sys.roundsExisted[i] = 0
					}
				}
				sys.loader.reset()
			}

			// If not restarting match
			if winp != -2 {
				// Cleanup
				var ti int32
				tbl_time := l.NewTable()
				for k, v := range sys.timerRounds {
					tbl_time.RawSetInt(k+1, lua.LNumber(v))
					ti += v
				}
				sc := sys.scoreStart
				tbl_score := l.NewTable()
				for k, v := range sys.scoreRounds {
					tbl_tmp := l.NewTable()
					tbl_tmp.RawSetInt(1, lua.LNumber(v[0]))
					tbl_tmp.RawSetInt(2, lua.LNumber(v[1]))
					tbl_score.RawSetInt(k+1, tbl_tmp)
					sc[0] += v[0]
					sc[1] += v[1]
				}
				tbl.RawSetString("match", sys.matchData)
				tbl.RawSetString("scoreRounds", tbl_score)
				tbl.RawSetString("timerRounds", tbl_time)
				tbl.RawSetString("matchTime", lua.LNumber(ti))
				tbl.RawSetString("roundTime", lua.LNumber(sys.roundTime))
				tbl.RawSetString("winTeam", lua.LNumber(sys.winTeam))
				tbl.RawSetString("lastRound", lua.LNumber(sys.round-1))
				tbl.RawSetString("draws", lua.LNumber(sys.draws))
				tbl.RawSetString("p1wins", lua.LNumber(sys.wins[0]))
				tbl.RawSetString("p2wins", lua.LNumber(sys.wins[1]))
				tbl.RawSetString("p1tmode", lua.LNumber(sys.tmode[0]))
				tbl.RawSetString("p2tmode", lua.LNumber(sys.tmode[1]))
				tbl.RawSetString("p1score", lua.LNumber(sc[0]))
				tbl.RawSetString("p2score", lua.LNumber(sc[1]))
				sys.timerStart = 0
				sys.timerRounds = []int32{}
				sys.scoreStart = [2]float32{}
				sys.scoreRounds = [][2]float32{}
				sys.timerCount = []int32{}
				sys.sel.cdefOverwrite = make(map[int]string)
				sys.sel.sdefOverwrite = ""
				l.Push(lua.LNumber(winp))
				l.Push(tbl)
				if sys.playBgmFlg {
					sys.bgm.Open("", 1, 100, 0, 0, 0)
					sys.playBgmFlg = false
				}
				sys.clearAllSound()
				sys.allPalFX = *newPalFX()
				sys.bgPalFX = *newPalFX()
				sys.superpmap = *newPalFX()
				sys.resetGblEffect()
				sys.dialogueFlg = false
				sys.dialogueForce = 0
				sys.dialogueBarsFlg = false
				sys.noSoundFlg = false
				sys.postMatchFlg = false
				sys.preFightTime += sys.gameTime
				sys.gameTime = 0
				sys.consoleText = []string{}
				sys.stageLoopNo = 0
				return 2
			}
		}
	})
	luaRegister(l, "getCharAttachedInfo", func(*lua.LState) int {
		def := strArg(l, 1)
		idx := strings.Index(def, "/")
		if len(def) >= 4 && strings.ToLower(def[len(def)-4:]) == ".def" {
			if idx < 0 {
				return 0
			}
		} else if idx < 0 {
			def += "/" + def + ".def"
		} else {
			def += ".def"
		}
		if strings.ToLower(def[0:6]) != "chars/" && strings.ToLower(def[1:3]) != ":/" && (def[0] != '/' || idx > 0 && !strings.Contains(def[:idx], ":")) {
			def = "chars/" + def
		}
		if def = FileExist(def); len(def) == 0 {
			return 0
		}
		str, err := LoadText(def)
		if err != nil {
			return 0
		}
		lines, i, info, files, name, sound := SplitAndTrim(str, "\n"), 0, true, true, "", ""
		for i < len(lines) {
			var is IniSection
			is, name, _ = ReadIniSection(lines, &i)
			switch name {
			case "info":
				if info {
					info = false
					var ok bool
					if name, ok, _ = is.getText("displayname"); !ok {
						name, _, _ = is.getText("name")
					}
				}
			case "files":
				if files {
					files = false
					sound = is["sound"]
				}
			}
		}
		tbl := l.NewTable()
		tbl.RawSetString("name", lua.LString(name))
		tbl.RawSetString("def", lua.LString(def))
		tbl.RawSetString("sound", lua.LString(sound))
		l.Push(tbl)
		return 1
	})
	luaRegister(l, "getCharFileName", func(*lua.LState) int {
		c := sys.sel.GetChar(int(numArg(l, 1)))
		l.Push(lua.LString(c.def))
		return 1
	})
	luaRegister(l, "getCharInfo", func(*lua.LState) int {
		c := sys.sel.GetChar(int(numArg(l, 1)))
		tbl := l.NewTable()
		tbl.RawSetString("name", lua.LString(c.name))
		tbl.RawSetString("author", lua.LString(c.author))
		tbl.RawSetString("def", lua.LString(c.def))
		tbl.RawSetString("sound", lua.LString(c.sound))
		tbl.RawSetString("intro", lua.LString(c.intro))
		tbl.RawSetString("ending", lua.LString(c.ending))
		tbl.RawSetString("arcadepath", lua.LString(c.arcadepath))
		tbl.RawSetString("ratiopath", lua.LString(c.ratiopath))
		tbl.RawSetString("portrait_scale", lua.LNumber(c.portrait_scale))
		subt := l.NewTable()
		for k, v := range c.cns_scale {
			subt.RawSetInt(k+1, lua.LNumber(v))
			subt.RawSetInt(k+1, lua.LNumber(v))
		}
		tbl.RawSetString("cns_scale", subt)
		//palettes
		subt = l.NewTable()
		if len(c.pal) > 0 {
			for k, v := range c.pal {
				subt.RawSetInt(k+1, lua.LNumber(v))
			}
		} else {
			subt.RawSetInt(1, lua.LNumber(1))
		}
		tbl.RawSetString("pal", subt)
		//default palettes
		subt = l.NewTable()
		pals := make(map[int32]bool)
		var n int
		if len(c.pal_defaults) > 0 {
			for _, v := range c.pal_defaults {
				if v > 0 && int(v) <= len(c.pal) {
					n++
					subt.RawSetInt(n, lua.LNumber(v))
					pals[v] = true
				}
			}
		}
		if len(c.pal) > 0 {
			for _, v := range c.pal {
				if !pals[v] {
					n++
					subt.RawSetInt(n, lua.LNumber(v))
				}
			}
		}
		if n == 0 {
			subt.RawSetInt(1, lua.LNumber(1))
		}
		tbl.RawSetString("pal_defaults", subt)
		//palette keymap
		subt = l.NewTable()
		if len(c.pal_keymap) > 0 {
			for k, v := range c.pal_keymap {
				if int32(k+1) != v { //only actual remaps are relevant
					subt.RawSetInt(k+1, lua.LNumber(v))
				}
			}
		}
		tbl.RawSetString("pal_keymap", subt)
		l.Push(tbl)
		return 1
	})
	luaRegister(l, "getCharDialogue", func(*lua.LState) int {
		pn := sys.dialogueForce
		if l.GetTop() >= 1 {
			pn = int(numArg(l, 1))
		}
		if pn != 0 && (pn < 1 || pn > MaxSimul*2+MaxAttachedChar) {
			l.RaiseError("\nInvalid player number: %v\n", pn)
		}
		tbl := l.NewTable()
		if pn == 0 {
			r := make([]int, 0)
			for i, p := range sys.chars {
				if len(p) > 0 && len(p[0].dialogue) > 0 {
					r = append(r, i)
				}
			}
			if len(r) > 0 {
				pn = r[rand.Int()%len(r)] + 1
			}
		}
		if pn >= 1 && pn <= len(sys.chars) && len(sys.chars[pn-1]) > 0 {
			for k, v := range sys.chars[pn-1][0].dialogue {
				tbl.RawSetInt(k+1, lua.LString(v))
			}
		}
		l.Push(tbl)
		l.Push(lua.LNumber(pn))
		return 2
	})
	luaRegister(l, "getCharMovelist", func(*lua.LState) int {
		c := sys.sel.GetChar(int(numArg(l, 1)))
		l.Push(lua.LString(c.movelist))
		return 1
	})
	luaRegister(l, "getCharName", func(*lua.LState) int {
		c := sys.sel.GetChar(int(numArg(l, 1)))
		l.Push(lua.LString(c.name))
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
	luaRegister(l, "getCharVictoryQuote", func(*lua.LState) int {
		pn := int(numArg(l, 1))
		if pn < 1 || pn > len(sys.chars) || len(sys.chars[pn-1]) == 0 {
			l.RaiseError("\nPlayer not found: %v\n", pn)
		}
		v := -1
		if l.GetTop() >= 2 {
			v = int(numArg(l, 2))
		} else {
			v = int(sys.chars[pn-1][0].winquote)
		}
		if v < 0 || v >= MaxQuotes {
			t := []int{}
			for i, q := range sys.cgi[sys.chars[pn-1][0].playerNo].quotes {
				if q != "" {
					t = append(t, i)
				}
			}
			if len(t) > 0 {
				v = rand.Int() % len(t)
				v = t[v]
			} else {
				v = -1
			}
		}
		if len(sys.cgi[sys.chars[pn-1][0].playerNo].quotes) == MaxQuotes && v != -1 {
			l.Push(lua.LString(sys.cgi[sys.chars[pn-1][0].playerNo].quotes[v]))
		} else {
			l.Push(lua.LString(""))
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
	luaRegister(l, "getConsecutiveWins", func(l *lua.LState) int {
		l.Push(lua.LNumber(sys.consecutiveWins[int(numArg(l, 1))-1]))
		return 1
	})
	luaRegister(l, "getDirectoryFiles", func(*lua.LState) int {
		dir := l.NewTable()
		filepath.Walk(strArg(l, 1), func(path string, info os.FileInfo, err error) error {
			dir.Append(lua.LString(path))
			return nil
		})
		l.Push(dir)
		return 1
	})
	luaRegister(l, "getFrameCount", func(l *lua.LState) int {
		l.Push(lua.LNumber(sys.frameCounter))
		return 1
	})
	luaRegister(l, "getJoystickName", func(*lua.LState) int {
		l.Push(lua.LString(input.GetJoystickName(int(numArg(l, 1)))))
		return 1
	})
	luaRegister(l, "getJoystickPresent", func(*lua.LState) int {
		l.Push(lua.LBool(input.IsJoystickPresent(int(numArg(l, 1)))))
		return 1
	})
	luaRegister(l, "getJoystickKey", func(*lua.LState) int {
		var s string
		var joy, min, max int = 0, 0, input.GetMaxJoystickCount()
		if l.GetTop() >= 1 {
			min = int(Clamp(int32(numArg(l, 1)), 0, int32(max-1)))
			max = min + 1
		}
		for joy = min; joy < max; joy++ {
			if input.IsJoystickPresent(joy) {
				axes := input.GetJoystickAxes(joy)
				btns := input.GetJoystickButtons(joy)
				name := input.GetJoystickName(joy)
				for i := range axes {
					if strings.Contains(name, "XInput") || strings.Contains(name, "X360") {
						if axes[i] > 0.5 {
							s = strconv.Itoa(-i*2 - 2)
						} else if axes[i] < -0.5 && i < 4 {
							s = strconv.Itoa(-i*2 - 1)
						}
					} else if name == "PS3 Controller" {
						if (len(axes) == 8 && i != 3 && i != 4 && i != 6 && i != 7) ||
							(len(axes) == 6 && i != 2 && i != 5) {
							// 8 axes in Windows (need to skip 3, 4, 6, 7) and
							// 6 axes in Linux (need to skip 2 and 5)
							if axes[i] < -0.2 {
								s = strconv.Itoa(-i*2 - 1)
							} else if axes[i] > 0.2 {
								s = strconv.Itoa(-i*2 - 2)
							}
						}
					} else if name != "PS4 Controller" || !(i == 3 || i == 4) {
						if axes[i] < -0.2 {
							s = strconv.Itoa(-i*2 - 1)
						} else if axes[i] > 0.2 {
							s = strconv.Itoa(-i*2 - 2)
						}
					}
				}
				for i := range btns {
					if btns[i] > 0 {
						s = strconv.Itoa(i)
					}
				}
				if s != "" {
					break
				}
			}
		}
		l.Push(lua.LString(s))
		if s != "" {
			l.Push(lua.LNumber(joy + 1))
		} else {
			l.Push(lua.LNumber(-1))
		}
		return 2
	})
	luaRegister(l, "getKey", func(*lua.LState) int {
		var s string
		if sys.keyInput != KeyUnknown {
			s = KeyToString(sys.keyInput)
		}
		if l.GetTop() == 0 {
			l.Push(lua.LString(s))
			return 1
		} else if strArg(l, 1) == "" {
			l.Push(lua.LBool(false))
			return 1
		}
		l.Push(lua.LBool(s == strArg(l, 1)))
		return 1
	})
	luaRegister(l, "getKeyText", func(*lua.LState) int {
		s := ""
		if sys.keyInput != KeyUnknown {
			if sys.keyInput == KeyInsert {
				s, _ = sys.window.GetClipboardString()
			} else {
				s = sys.keyString
			}
		}
		l.Push(lua.LString(s))
		return 1
	})
	luaRegister(l, "getListenPort", func(*lua.LState) int {
		l.Push(lua.LString(sys.listenPort))
		return 1
	})
	luaRegister(l, "getMatchMaxDrawGames", func(l *lua.LState) int {
		tn := int(numArg(l, 1))
		if tn < 1 || tn > 2 {
			l.RaiseError("\nInvalid team side: %v\n", tn)
		}
		l.Push(lua.LNumber(sys.lifebar.ro.match_maxdrawgames[tn-1]))
		return 1
	})
	luaRegister(l, "getMatchWins", func(l *lua.LState) int {
		tn := int(numArg(l, 1))
		if tn < 1 || tn > 2 {
			l.RaiseError("\nInvalid team side: %v\n", tn)
		}
		l.Push(lua.LNumber(sys.lifebar.ro.match_wins[tn-1]))
		return 1
	})
	luaRegister(l, "getRoundTime", func(l *lua.LState) int {
		l.Push(lua.LNumber(sys.roundTime))
		return 1
	})
	luaRegister(l, "getStageInfo", func(*lua.LState) int {
		c := sys.sel.GetStage(int(numArg(l, 1)))
		tbl := l.NewTable()
		tbl.RawSetString("name", lua.LString(c.name))
		tbl.RawSetString("def", lua.LString(c.def))
		tbl.RawSetString("portrait_scale", lua.LNumber(c.portrait_scale))
		tbl.RawSetString("attachedchardef", lua.LString(c.attachedchardef))
		subt := l.NewTable()
		for k, v := range c.stagebgm {
			subt.RawSetString(k, lua.LString(v))
		}
		tbl.RawSetString("stagebgm", subt)
		l.Push(tbl)
		return 1
	})
	luaRegister(l, "getStageNo", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.sel.selectedStageNo))
		return 1
	})
	luaRegister(l, "getWaveData", func(*lua.LState) int {
		//path, group, sound, loops before give up searching for group/sound pair (optional)
		var max uint32
		if l.GetTop() >= 4 {
			max = uint32(numArg(l, 4))
		}
		w, err := loadFromSnd(strArg(l, 1), int32(numArg(l, 2)), int32(numArg(l, 3)), max)
		if err != nil {
			l.RaiseError(err.Error())
		}
		l.Push(newUserData(l, w))
		return 1
	})
	luaRegister(l, "loadDebugFont", func(l *lua.LState) int {
		ts := NewTextSprite()
		f, err := loadFnt(strArg(l, 1), -1)
		if err != nil {
			l.RaiseError("\nCan't load %v: %v\n", strArg(l, 1), err.Error())
		}
		ts.fnt = f
		if l.GetTop() >= 2 {
			ts.xscl, ts.yscl = float32(numArg(l, 2)), float32(numArg(l, 2))
		}
		sys.debugFont = ts
		return 0
	})
	luaRegister(l, "loadDebugInfo", func(l *lua.LState) int {
		tableArg(l, 1).ForEach(func(_, value lua.LValue) {
			sys.listLFunc = append(sys.listLFunc, sys.luaLState.GetGlobal(lua.LVAsString(value)).(*lua.LFunction))
		})
		return 0
	})
	luaRegister(l, "loadDebugStatus", func(l *lua.LState) int {
		sys.statusLFunc, _ = sys.luaLState.GetGlobal(strArg(l, 1)).(*lua.LFunction)
		return 0
	})
	luaRegister(l, "loading", func(l *lua.LState) int {
		l.Push(lua.LBool(sys.loader.state == LS_Loading))
		return 1
	})
	luaRegister(l, "loadLifebar", func(l *lua.LState) int {
		lb, err := loadLifebar(strArg(l, 1))
		if err != nil {
			l.RaiseError("\nCan't load %v: %v\n", strArg(l, 1), err.Error())
		}
		sys.lifebar = *lb
		return 0
	})
	luaRegister(l, "loadStart", func(l *lua.LState) int {
		if sys.gameMode != "randomtest" {
			for k, v := range sys.sel.selected {
				if len(v) < int(sys.numSimul[k]) {
					l.RaiseError("\nNot enough P%v side chars to load: expected %v, got %v\n", k+1, sys.numSimul[k], len(v))
				}
			}
		}
		if sys.sel.selectedStageNo == -1 {
			l.RaiseError("\nStage not selected for load\n")
		}
		sys.loadStart()
		return 0
	})
	luaRegister(l, "numberToRune", func(l *lua.LState) int {
		l.Push(lua.LString(fmt.Sprint('A' - 1 + int(numArg(l, 1)))))
		return 1
	})
	luaRegister(l, "overrideCharData", func(l *lua.LState) int {
		tn := int(numArg(l, 1))
		if tn < 1 || tn > 2 {
			l.RaiseError("\nInvalid team side: %v\n", tn)
		}
		mn := int(numArg(l, 2))
		if len(sys.sel.ocd[tn-1]) == 0 {
			l.RaiseError("\noverrideCharData function used before loading player %v, member %v\n", tn, mn)
		}
		tableArg(l, 3).ForEach(func(key, value lua.LValue) {
			switch k := key.(type) {
			case lua.LString:
				switch string(k) {
				case "life":
					sys.sel.ocd[tn-1][mn-1].life = int32(lua.LVAsNumber(value))
				case "lifeMax":
					sys.sel.ocd[tn-1][mn-1].lifeMax = int32(lua.LVAsNumber(value))
				case "power":
					sys.sel.ocd[tn-1][mn-1].power = int32(lua.LVAsNumber(value))
				case "dizzyPoints":
					sys.sel.ocd[tn-1][mn-1].dizzyPoints = int32(lua.LVAsNumber(value))
				case "guardPoints":
					sys.sel.ocd[tn-1][mn-1].guardPoints = int32(lua.LVAsNumber(value))
				case "ratioLevel":
					sys.sel.ocd[tn-1][mn-1].ratioLevel = int32(lua.LVAsNumber(value))
				case "lifeRatio":
					sys.sel.ocd[tn-1][mn-1].lifeRatio = float32(lua.LVAsNumber(value))
				case "attackRatio":
					sys.sel.ocd[tn-1][mn-1].attackRatio = float32(lua.LVAsNumber(value))
				case "existed":
					sys.sel.ocd[tn-1][mn-1].existed = lua.LVAsBool(value)
				default:
					l.RaiseError("\nInvalid table key: %v\n", k)
				}
			default:
				l.RaiseError("\nInvalid table key type: %v\n", fmt.Sprintf("%T\n", key))
			}
		})
		return 0
	})
	luaRegister(l, "panicError", func(*lua.LState) int {
		l.RaiseError(strArg(l, 1))
		return 0
	})
	luaRegister(l, "playBGM", func(l *lua.LState) int {
		var loop, volume, loopstart, loopend, startposition int = 1, 100, 0, 0, 0
		if l.GetTop() >= 2 {
			loop = int(numArg(l, 2))
		}
		if l.GetTop() >= 3 {
			volume = int(numArg(l, 3))
		}
		if l.GetTop() >= 4 {
			loopstart = int(numArg(l, 4))
		}
		if l.GetTop() >= 5 && numArg(l, 5) > 1 {
			loopend = int(numArg(l, 5))
		}
		if l.GetTop() >= 6 && numArg(l, 6) > 1 {
			startposition = int(numArg(l, 6))
		}
		sys.bgm.Open(strArg(l, 1), loop, volume, loopstart, loopend, startposition)
		return 0
	})
	luaRegister(l, "playerBufReset", func(*lua.LState) int {
		if l.GetTop() >= 1 {
			pn := int(numArg(l, 1))
			if pn < 1 || pn > len(sys.chars) || len(sys.chars[pn-1]) == 0 {
				return 0
			}
			for j := range sys.chars[pn-1][0].cmd {
				sys.chars[pn-1][0].cmd[j].BufReset()
				sys.chars[pn-1][0].setSF(CSF_nohardcodedkeys)
			}
		} else {
			for _, p := range sys.chars {
				if len(p) > 0 {
					for j := range p[0].cmd {
						p[0].cmd[j].BufReset()
						p[0].setSF(CSF_nohardcodedkeys)
					}
				}
			}
		}
		return 0
	})
	luaRegister(l, "preloadListChar", func(*lua.LState) int {
		if l.GetTop() >= 2 {
			sys.sel.charSpritePreload[[...]int16{int16(numArg(l, 1)), int16(numArg(l, 2))}] = true
		} else {
			sys.sel.charAnimPreload = append(sys.sel.charAnimPreload, int32(numArg(l, 1)))
		}
		return 0
	})
	luaRegister(l, "preloadListStage", func(*lua.LState) int {
		if l.GetTop() >= 2 {
			sys.sel.stageSpritePreload[[...]int16{int16(numArg(l, 1)), int16(numArg(l, 2))}] = true
		} else {
			sys.sel.stageAnimPreload = append(sys.sel.stageAnimPreload, int32(numArg(l, 1)))
		}
		return 0
	})
	luaRegister(l, "printConsole", func(l *lua.LState) int {
		if l.GetTop() >= 2 && boolArg(l, 2) {
			sys.consoleText[len(sys.consoleText)-1] += strArg(l, 1)
		} else {
			sys.appendToConsole(strArg(l, 1))
		}
		fmt.Println(strArg(l, 1))
		return 0
	})
	luaRegister(l, "puts", func(*lua.LState) int {
		fmt.Println(strArg(l, 1))
		return 0
	})
	luaRegister(l, "refresh", func(*lua.LState) int {
		sys.tickSound()
		if !sys.update() {
			l.RaiseError("<game end>")
		}
		return 0
	})
	luaRegister(l, "reload", func(*lua.LState) int {
		sys.reloadFlg = true
		for i := range sys.reloadCharSlot {
			sys.reloadCharSlot[i] = true
		}
		sys.reloadStageFlg = true
		sys.reloadLifebarFlg = true
		return 0
	})
	luaRegister(l, "remapInput", func(l *lua.LState) int {
		src, dst := int(numArg(l, 1)), int(numArg(l, 2))
		if src < 1 || src > len(sys.inputRemap) ||
			dst < 1 || dst > len(sys.inputRemap) {
			l.RaiseError("\nInvalid player number: %v, %v\n", src, dst)
		}
		sys.inputRemap[src-1] = dst - 1
		return 0
	})
	luaRegister(l, "removeDizzy", func(*lua.LState) int {
		sys.debugWC.unsetSCF(SCF_dizzy)
		return 0
	})
	luaRegister(l, "replayRecord", func(*lua.LState) int {
		if sys.netInput != nil {
			sys.netInput.rep, _ = os.Create(strArg(l, 1))
		}
		return 0
	})
	luaRegister(l, "replayStop", func(*lua.LState) int {
		if sys.netInput != nil && sys.netInput.rep != nil {
			sys.netInput.rep.Close()
			sys.netInput.rep = nil
		}
		return 0
	})
	luaRegister(l, "resetKey", func(*lua.LState) int {
		sys.keyInput = KeyUnknown
		sys.keyString = ""
		return 0
	})
	luaRegister(l, "resetAILevel", func(l *lua.LState) int {
		for i := range sys.com {
			sys.com[i] = 0
		}
		return 0
	})
	luaRegister(l, "resetMatchData", func(*lua.LState) int {
		sys.allPalFX = *newPalFX()
		sys.bgPalFX = *newPalFX()
		sys.superpmap = *newPalFX()
		sys.resetGblEffect()
		for i, p := range sys.chars {
			if len(p) > 0 {
				sys.playerClear(i, boolArg(l, 1))
			}
		}
		return 0
	})
	luaRegister(l, "resetRemapInput", func(l *lua.LState) int {
		sys.resetRemapInput()
		return 0
	})
	luaRegister(l, "resetScore", func(*lua.LState) int {
		tn := int(numArg(l, 1))
		if tn < 1 || tn > 2 {
			l.RaiseError("\nInvalid team side: %v\n", tn)
		}
		sys.lifebar.sc[tn-1].scorePoints = 0
		return 0
	})
	luaRegister(l, "roundReset", func(*lua.LState) int {
		sys.roundResetFlg = true
		return 0
	})
	luaRegister(l, "screenshot", func(*lua.LState) int {
		captureScreen()
		return 0
	})
	luaRegister(l, "searchFile", func(l *lua.LState) int {
		var dirs []string
		tableArg(l, 2).ForEach(func(key, value lua.LValue) {
			dirs = append(dirs, lua.LVAsString(value))
		})
		l.Push(lua.LString(SearchFile(strArg(l, 1), dirs)))
		return 1
	})
	luaRegister(l, "selectChar", func(*lua.LState) int {
		cn := int(numArg(l, 2))
		if cn < 0 || cn >= len(sys.sel.charlist) {
			l.RaiseError("\nInvalid char ref: %v\n", cn)
		}
		tn := int(numArg(l, 1))
		if tn < 1 || tn > 2 {
			l.RaiseError("%v\nInvalid team side: %v\n", sys.sel.GetChar(cn).def, tn)
		}
		pl := int(numArg(l, 3))
		if pl < 1 || pl > 12 {
			l.RaiseError("%v\nInvalid palette: %v\n", sys.sel.GetChar(cn).def, pl)
		}
		var ret int
		if sys.sel.AddSelectedChar(tn-1, cn, pl) {
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
			case TM_Tag:
				if len(sys.sel.selected[tn-1]) >= int(sys.numSimul[tn-1]) {
					ret = 2
				} else {
					ret = 1
				}
			}
		}
		l.Push(lua.LNumber(ret))
		return 1
	})
	luaRegister(l, "selectStage", func(*lua.LState) int {
		sn := int(numArg(l, 1))
		if sn < 0 || sn > len(sys.sel.stagelist) {
			l.RaiseError("\nInvalid stage ref: %v\n", sn)
		}
		sys.sel.SelectStage(sn)
		return 0
	})
	luaRegister(l, "selectStart", func(l *lua.LState) int {
		sys.sel.ClearSelected()
		sys.loadStart()
		return 0
	})
	luaRegister(l, "sffNew", func(l *lua.LState) int {
		if l.GetTop() == 0 {
			l.Push(newUserData(l, newSff()))
			return 1
		}
		sff, err := loadSff(strArg(l, 1), false)
		if err != nil {
			l.RaiseError("\nCan't load %v: %v\n", strArg(l, 1), err.Error())
		}
		sys.runMainThreadTask()
		l.Push(newUserData(l, sff))
		return 1
	})
	luaRegister(l, "selfState", func(*lua.LState) int {
		sys.debugWC.selfState(int32(numArg(l, 1)), -1, -1, 1, false)
		return 0
	})
	luaRegister(l, "setAccel", func(*lua.LState) int {
		sys.accel = float32(numArg(l, 1))
		return 0
	})
	luaRegister(l, "setAILevel", func(*lua.LState) int {
		level := float32(numArg(l, 1))
		sys.com[sys.debugWC.playerNo] = level
		for _, c := range sys.chars[sys.debugWC.playerNo] {
			if level == 0 {
				c.key = sys.debugWC.playerNo
			} else {
				c.key = ^sys.debugWC.playerNo
			}
		}
		return 0
	})
	luaRegister(l, "setAllowDebugKeys", func(l *lua.LState) int {
		sys.allowDebugKeys = boolArg(l, 1)
		return 0
	})
	luaRegister(l, "setAllowDebugMode", func(l *lua.LState) int {
		d := boolArg(l, 1)
		if !d {
			if sys.clsnDraw {
				sys.clsnDraw = false
			}
			if sys.debugDraw {
				sys.debugDraw = false
			}
		}
		sys.allowDebugMode = d
		return 0
	})
	luaRegister(l, "setAudioDucking", func(l *lua.LState) int {
		sys.audioDucking = boolArg(l, 1)
		return 0
	})
	luaRegister(l, "setAutoguard", func(l *lua.LState) int {
		pn := int(numArg(l, 1))
		if pn < 1 || pn > MaxSimul*2+MaxAttachedChar {
			l.RaiseError("\nInvalid player number: %v\n", pn)
		}
		sys.autoguard[pn-1] = boolArg(l, 2)
		return 0
	})
	luaRegister(l, "setAutoLevel", func(*lua.LState) int {
		sys.autolevel = boolArg(l, 1)
		return 0
	})
	luaRegister(l, "setCom", func(*lua.LState) int {
		pn := int(numArg(l, 1))
		ailv := float32(numArg(l, 2))
		if pn < 1 || pn > MaxSimul*2+MaxAttachedChar {
			l.RaiseError("\nInvalid player number: %v\n", pn)
		}
		if ailv > 0 {
			sys.com[pn-1] = ailv
		} else {
			sys.com[pn-1] = 0
		}
		return 0
	})
	luaRegister(l, "setConsecutiveWins", func(l *lua.LState) int {
		sys.consecutiveWins[int(numArg(l, 1))-1] = int32(numArg(l, 2))
		return 0
	})
	luaRegister(l, "setContinue", func(l *lua.LState) int {
		sys.continueFlg = boolArg(l, 1)
		return 0
	})
	luaRegister(l, "setDizzyPoints", func(*lua.LState) int {
		sys.debugWC.dizzyPointsSet(int32(numArg(l, 1)))
		return 0
	})
	luaRegister(l, "setGameMode", func(*lua.LState) int {
		sys.gameMode = strArg(l, 1)
		return 0
	})
	luaRegister(l, "setGameSpeed", func(*lua.LState) int {
		sys.gameSpeed = float32(numArg(l, 1)) / float32(FPS)
		return 0
	})
	luaRegister(l, "setGuardPoints", func(*lua.LState) int {
		sys.debugWC.guardPointsSet(int32(numArg(l, 1)))
		return 0
	})
	luaRegister(l, "setHomeTeam", func(l *lua.LState) int {
		tn := int(numArg(l, 1))
		if tn < 1 || tn > 2 {
			l.RaiseError("\nInvalid team side: %v\n", tn)
		}
		sys.home = tn - 1
		return 0
	})
	luaRegister(l, "setKeyConfig", func(l *lua.LState) int {
		pn := int(numArg(l, 1))
		joy := int(numArg(l, 2))
		if pn < 1 || (joy == -1 && pn > len(sys.keyConfig)) || (joy >= 0 && pn > len(sys.joystickConfig)) {
			l.RaiseError("\nInvalid player number: %v\n", pn)
		}
		if joy < -1 || joy > len(sys.joystickConfig) {
			l.RaiseError("\nInvalid controller number: %v\n", joy)
		}
		tableArg(l, 3).ForEach(func(key, value lua.LValue) {
			if joy == -1 {
				btn := int(StringToKey(lua.LVAsString(value)))
				switch int(lua.LVAsNumber(key)) {
				case 1:
					sys.keyConfig[pn-1].dU = btn
				case 2:
					sys.keyConfig[pn-1].dD = btn
				case 3:
					sys.keyConfig[pn-1].dL = btn
				case 4:
					sys.keyConfig[pn-1].dR = btn
				case 5:
					sys.keyConfig[pn-1].kA = btn
				case 6:
					sys.keyConfig[pn-1].kB = btn
				case 7:
					sys.keyConfig[pn-1].kC = btn
				case 8:
					sys.keyConfig[pn-1].kX = btn
				case 9:
					sys.keyConfig[pn-1].kY = btn
				case 10:
					sys.keyConfig[pn-1].kZ = btn
				case 11:
					sys.keyConfig[pn-1].kS = btn
				case 12:
					sys.keyConfig[pn-1].kD = btn
				case 13:
					sys.keyConfig[pn-1].kW = btn
				case 14:
					sys.keyConfig[pn-1].kM = btn
				}
			} else {
				btn, err := strconv.Atoi(lua.LVAsString(value))
				if err != nil {
					btn = 999
				}
				switch int(lua.LVAsNumber(key)) {
				case 1:
					sys.joystickConfig[pn-1].dU = btn
				case 2:
					sys.joystickConfig[pn-1].dD = btn
				case 3:
					sys.joystickConfig[pn-1].dL = btn
				case 4:
					sys.joystickConfig[pn-1].dR = btn
				case 5:
					sys.joystickConfig[pn-1].kA = btn
				case 6:
					sys.joystickConfig[pn-1].kB = btn
				case 7:
					sys.joystickConfig[pn-1].kC = btn
				case 8:
					sys.joystickConfig[pn-1].kX = btn
				case 9:
					sys.joystickConfig[pn-1].kY = btn
				case 10:
					sys.joystickConfig[pn-1].kZ = btn
				case 11:
					sys.joystickConfig[pn-1].kS = btn
				case 12:
					sys.joystickConfig[pn-1].kD = btn
				case 13:
					sys.joystickConfig[pn-1].kW = btn
				case 14:
					sys.joystickConfig[pn-1].kM = btn
				}
			}
		})
		return 0
	})
	luaRegister(l, "setLife", func(*lua.LState) int {
		if sys.debugWC.alive() {
			sys.debugWC.lifeSet(int32(numArg(l, 1)))
		}
		return 0
	})
	luaRegister(l, "setLifeShare", func(l *lua.LState) int {
		tn := int(numArg(l, 1))
		if tn < 1 || tn > 2 {
			l.RaiseError("\nInvalid team side: %v\n", tn)
		}
		sys.lifeShare[tn-1] = boolArg(l, 2)
		return 0
	})
	luaRegister(l, "setLifebarElements", func(*lua.LState) int {
		// elements enabled via fight.def, depending on game mode
		if _, ok := sys.lifebar.ma.enabled[sys.gameMode]; ok {
			sys.lifebar.ma.active = sys.lifebar.ma.enabled[sys.gameMode]
		}
		for _, v := range sys.lifebar.ai {
			if _, ok := v.enabled[sys.gameMode]; ok {
				v.active = v.enabled[sys.gameMode]
			}
		}
		for _, v := range sys.lifebar.sc {
			if _, ok := v.enabled[sys.gameMode]; ok {
				v.active = v.enabled[sys.gameMode]
			}
		}
		for _, v := range sys.lifebar.wc {
			if _, ok := v.enabled[sys.gameMode]; ok {
				v.active = v.enabled[sys.gameMode]
			}
		}
		if _, ok := sys.lifebar.tr.enabled[sys.gameMode]; ok {
			sys.lifebar.tr.active = sys.lifebar.tr.enabled[sys.gameMode]
		}
		// elements forced by lua scripts
		tableArg(l, 1).ForEach(func(key, value lua.LValue) {
			switch k := key.(type) {
			case lua.LString:
				switch string(k) {
				case "active": //enabled by default
					sys.lifebar.active = lua.LVAsBool(value)
				case "bars": //enabled by default
					sys.lifebar.bars = lua.LVAsBool(value)
				case "guardbar": //enabled depending on config.json
					sys.lifebar.guardbar = lua.LVAsBool(value)
				case "hidebars": //enabled depending on dialogue system.def settings
					sys.lifebar.hidebars = lua.LVAsBool(value)
				case "match":
					sys.lifebar.ma.active = lua.LVAsBool(value)
				case "mode": //enabled by default
					sys.lifebar.mode = lua.LVAsBool(value)
				case "p1aiLevel":
					sys.lifebar.ai[0].active = lua.LVAsBool(value)
				case "p1score":
					sys.lifebar.sc[0].active = lua.LVAsBool(value)
				case "p1winCount":
					sys.lifebar.wc[0].active = lua.LVAsBool(value)
				case "p2aiLevel":
					sys.lifebar.ai[1].active = lua.LVAsBool(value)
				case "p2score":
					sys.lifebar.sc[1].active = lua.LVAsBool(value)
				case "p2winCount":
					sys.lifebar.wc[1].active = lua.LVAsBool(value)
				case "redlifebar": //enabled depending on config.json
					sys.lifebar.redlifebar = lua.LVAsBool(value)
				case "stunbar": //enabled depending on config.json
					sys.lifebar.stunbar = lua.LVAsBool(value)
				case "timer":
					sys.lifebar.tr.active = lua.LVAsBool(value)
				default:
					l.RaiseError("\nInvalid table key: %v\n", k)
				}
			default:
				l.RaiseError("\nInvalid table key type: %v\n", fmt.Sprintf("%T\n", key))
			}
		})
		return 0
	})
	luaRegister(l, "setLifebarLocalcoord", func(l *lua.LState) int {
		sys.lifebarLocalcoord[0] = int32(numArg(l, 1))
		sys.lifebarLocalcoord[1] = int32(numArg(l, 2))
		return 0
	})
	luaRegister(l, "setLifebarOffsetX", func(l *lua.LState) int {
		sys.lifebarOffsetX = float32(numArg(l, 1))
		return 0
	})
	luaRegister(l, "setLifebarScale", func(l *lua.LState) int {
		sys.lifebarScale = float32(numArg(l, 1))
		return 0
	})
	luaRegister(l, "setLifebarPortraitScale", func(l *lua.LState) int {
		sys.lifebarPortraitScale = float32(numArg(l, 1))
		return 0
	})
	luaRegister(l, "setLifebarScore", func(*lua.LState) int {
		sys.scoreStart[0] = float32(numArg(l, 1))
		if l.GetTop() >= 2 {
			sys.scoreStart[1] = float32(numArg(l, 2))
		}
		return 0
	})
	luaRegister(l, "setLifebarTimer", func(*lua.LState) int {
		sys.timerStart = int32(numArg(l, 1))
		return 0
	})
	luaRegister(l, "setLifeMul", func(l *lua.LState) int {
		sys.lifeMul = float32(numArg(l, 1))
		return 0
	})
	luaRegister(l, "setListenPort", func(*lua.LState) int {
		sys.listenPort = strArg(l, 1)
		return 0
	})
	luaRegister(l, "setLoseSimul", func(l *lua.LState) int {
		sys.loseSimul = boolArg(l, 1)
		return 0
	})
	luaRegister(l, "setLoseTag", func(l *lua.LState) int {
		sys.loseTag = boolArg(l, 1)
		return 0
	})
	luaRegister(l, "setLuaLocalcoord", func(l *lua.LState) int {
		sys.luaLocalcoord[0] = int32(numArg(l, 1))
		sys.luaLocalcoord[1] = int32(numArg(l, 2))
		return 0
	})
	luaRegister(l, "setLuaPortraitScale", func(l *lua.LState) int {
		sys.luaPortraitScale = float32(numArg(l, 1))
		return 0
	})
	luaRegister(l, "setLuaSpriteOffsetX", func(l *lua.LState) int {
		sys.luaSpriteOffsetX = float32(numArg(l, 1))
		return 0
	})
	luaRegister(l, "setLuaSpriteScale", func(l *lua.LState) int {
		sys.luaSpriteScale = float32(numArg(l, 1))
		return 0
	})
	luaRegister(l, "setMatchMaxDrawGames", func(l *lua.LState) int {
		tn := int(numArg(l, 1))
		if tn < 1 || tn > 2 {
			l.RaiseError("\nInvalid team side: %v\n", tn)
		}
		sys.lifebar.ro.match_maxdrawgames[tn-1] = int32(numArg(l, 2))
		return 0
	})
	luaRegister(l, "setMatchNo", func(l *lua.LState) int {
		sys.match = int32(numArg(l, 1))
		return 0
	})
	luaRegister(l, "setMatchWins", func(l *lua.LState) int {
		tn := int(numArg(l, 1))
		if tn < 1 || tn > 2 {
			l.RaiseError("\nInvalid team side: %v\n", tn)
		}
		sys.lifebar.ro.match_wins[tn-1] = int32(numArg(l, 2))
		return 0
	})
	luaRegister(l, "setMaxAfterImage", func(l *lua.LState) int {
		sys.afterImageMax = int32(numArg(l, 1))
		return 0
	})
	luaRegister(l, "setMaxExplod", func(l *lua.LState) int {
		sys.explodMax = int(numArg(l, 1))
		return 0
	})
	luaRegister(l, "setMaxHelper", func(l *lua.LState) int {
		sys.helperMax = int32(numArg(l, 1))
		return 0
	})
	luaRegister(l, "setMaxPlayerProjectile", func(l *lua.LState) int {
		sys.playerProjectileMax = int(numArg(l, 1))
		return 0
	})
	luaRegister(l, "setMotifDir", func(*lua.LState) int {
		sys.motifDir = strArg(l, 1)
		return 0
	})
	luaRegister(l, "setPanningRange", func(l *lua.LState) int {
		sys.panningRange = float32(numArg(l, 1))
		return 0
	})
	luaRegister(l, "setPlayers", func(l *lua.LState) int {
		total := int(numArg(l, 1))
		if len(sys.keyConfig) > total {
			sys.keyConfig = sys.keyConfig[:total]
		} else {
			for i := len(sys.keyConfig); i < total; i++ {
				sys.keyConfig = append(sys.keyConfig, KeyConfig{})
			}
		}
		if len(sys.joystickConfig) > total {
			sys.joystickConfig = sys.joystickConfig[:total]
		} else {
			for i := len(sys.joystickConfig); i < total; i++ {
				sys.joystickConfig = append(sys.joystickConfig, KeyConfig{})
			}
		}
		return 0
	})
	luaRegister(l, "setPower", func(*lua.LState) int {
		sys.debugWC.setPower(int32(numArg(l, 1)))
		return 0
	})
	luaRegister(l, "setPowerShare", func(l *lua.LState) int {
		tn := int(numArg(l, 1))
		if tn < 1 || tn > 2 {
			l.RaiseError("\nInvalid team side: %v\n", tn)
		}
		sys.powerShare[tn-1] = boolArg(l, 2)
		return 0
	})
	luaRegister(l, "setRedLife", func(*lua.LState) int {
		sys.debugWC.redLifeSet(int32(numArg(l, 1)))
		return 0
	})
	luaRegister(l, "setRoundTime", func(l *lua.LState) int {
		sys.roundTime = int32(numArg(l, 1))
		return 0
	})
	luaRegister(l, "setStereoEffects", func(l *lua.LState) int {
		sys.stereoEffects = boolArg(l, 1)
		return 0
	})
	luaRegister(l, "setTeam1VS2Life", func(l *lua.LState) int {
		sys.team1VS2Life = float32(numArg(l, 1))
		return 0
	})
	luaRegister(l, "setTeamMode", func(*lua.LState) int {
		tn := int(numArg(l, 1))
		if tn < 1 || tn > 2 {
			l.RaiseError("\nInvalid team side: %v\n", tn)
		}
		tm := TeamMode(numArg(l, 2))
		if tm < 0 || tm > TM_LAST {
			l.RaiseError("\nInvalid team mode: %v\n", tm)
		}
		nt := int32(numArg(l, 3))
		if nt < 1 || (tm != TM_Turns && nt > MaxSimul) {
			l.RaiseError("\nInvalid team size: %v\n", nt)
		}
		sys.sel.selected[tn-1], sys.sel.ocd[tn-1] = nil, nil
		sys.tmode[tn-1] = tm
		if tm == TM_Turns {
			sys.numSimul[tn-1] = 1
		} else {
			sys.numSimul[tn-1] = nt
		}
		sys.numTurns[tn-1] = nt
		if (tm == TM_Simul || tm == TM_Tag) && nt == 1 {
			sys.tmode[tn-1] = TM_Single
		}
		return 0
	})
	luaRegister(l, "setTime", func(*lua.LState) int {
		sys.time = int32(numArg(l, 1))
		return 0
	})
	luaRegister(l, "setTimeFramesPerCount", func(l *lua.LState) int {
		sys.lifebar.ti.framespercount = int32(numArg(l, 1))
		return 0
	})
	luaRegister(l, "setVolumeMaster", func(l *lua.LState) int {
		sys.masterVolume = int(numArg(l, 1))
		sys.bgm.UpdateVolume()
		return 0
	})
	luaRegister(l, "setVolumeBgm", func(l *lua.LState) int {
		sys.bgmVolume = int(numArg(l, 1))
		sys.bgm.UpdateVolume()
		return 0
	})
	luaRegister(l, "setVolumeSfx", func(l *lua.LState) int {
		sys.wavVolume = int(numArg(l, 1))
		return 0
	})
	luaRegister(l, "setWinCount", func(*lua.LState) int {
		tn := int(numArg(l, 1))
		if tn < 1 || tn > 2 {
			l.RaiseError("\nInvalid team side: %v\n", tn)
		}
		sys.lifebar.wc[tn-1].wins = int32(numArg(l, 2))
		return 0
	})
	luaRegister(l, "setZoom", func(l *lua.LState) int {
		sys.cam.ZoomActive = boolArg(l, 1)
		return 0
	})
	luaRegister(l, "setZoomMax", func(l *lua.LState) int {
		sys.cam.ZoomMax = float32(numArg(l, 1))
		return 0
	})
	luaRegister(l, "setZoomMin", func(l *lua.LState) int {
		sys.cam.ZoomMin = float32(numArg(l, 1))
		return 0
	})
	luaRegister(l, "setZoomSpeed", func(l *lua.LState) int {
		sys.cam.ZoomSpeed = 12 - float32(numArg(l, 1))
		return 0
	})
	luaRegister(l, "sleep", func(l *lua.LState) int {
		time.Sleep(time.Duration((numArg(l, 1))) * time.Second)
		return 0
	})
	luaRegister(l, "sndNew", func(l *lua.LState) int {
		snd, err := LoadSnd(strArg(l, 1))
		if err != nil {
			l.RaiseError("\nCan't load %v: %v\n", strArg(l, 1), err.Error())
		}
		l.Push(newUserData(l, snd))
		return 1
	})
	luaRegister(l, "sndPlay", func(l *lua.LState) int {
		s, ok := toUserData(l, 1).(*Snd)
		if !ok {
			userDataError(l, 1, s)
		}
		volumescale := int32(100)
		if l.GetTop() >= 4 {
			volumescale = int32(numArg(l, 4))
		}
		var pan float32
		if l.GetTop() >= 5 {
			pan = float32(numArg(l, 5))
		}
		s.play([...]int32{int32(numArg(l, 2)), int32(numArg(l, 3))}, volumescale, pan)
		return 0
	})
	luaRegister(l, "sndPlaying", func(*lua.LState) int {
		s, ok := toUserData(l, 1).(*Snd)
		if !ok {
			userDataError(l, 1, s)
		}
		var f bool
		if w := s.Get([...]int32{int32(numArg(l, 2)), int32(numArg(l, 3))}); w != nil {
			f = sys.soundChannels.IsPlaying(w)
		}
		l.Push(lua.LBool(f))
		return 1
	})
	luaRegister(l, "sndStop", func(l *lua.LState) int {
		s, ok := toUserData(l, 1).(*Snd)
		if !ok {
			userDataError(l, 1, s)
		}
		s.stop([...]int32{int32(numArg(l, 2)), int32(numArg(l, 3))})
		return 0
	})
	luaRegister(l, "sszRandom", func(l *lua.LState) int {
		l.Push(lua.LNumber(Random()))
		return 1
	})
	luaRegister(l, "step", func(*lua.LState) int {
		sys.step = true
		return 0
	})
	luaRegister(l, "synchronize", func(*lua.LState) int {
		if err := sys.synchronize(); err != nil {
			l.RaiseError(err.Error())
		}
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
	luaRegister(l, "textImgNew", func(*lua.LState) int {
		l.Push(newUserData(l, NewTextSprite()))
		return 1
	})
	luaRegister(l, "textImgSetAlign", func(*lua.LState) int {
		ts, ok := toUserData(l, 1).(*TextSprite)
		if !ok {
			userDataError(l, 1, ts)
		}
		ts.align = int32(numArg(l, 2))
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
	luaRegister(l, "textImgSetColor", func(*lua.LState) int {
		ts, ok := toUserData(l, 1).(*TextSprite)
		if !ok {
			userDataError(l, 1, ts)
		}
		ts.SetColor(int32(numArg(l, 2)), int32(numArg(l, 3)), int32(numArg(l, 4)))
		return 0
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
	luaRegister(l, "textImgSetPos", func(*lua.LState) int {
		ts, ok := toUserData(l, 1).(*TextSprite)
		if !ok {
			userDataError(l, 1, ts)
		}
		if boolArg(l, 3) {
			ts.x, ts.y = float32(numArg(l, 2))/sys.luaSpriteScale+sys.luaSpriteOffsetX, float32(numArg(l, 3))/sys.luaSpriteScale
		}
		return 0
	})
	luaRegister(l, "textImgSetScale", func(*lua.LState) int {
		ts, ok := toUserData(l, 1).(*TextSprite)
		if !ok {
			userDataError(l, 1, ts)
		}
		ts.xscl, ts.yscl = float32(numArg(l, 2))/sys.luaSpriteScale, float32(numArg(l, 3))/sys.luaSpriteScale
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
	luaRegister(l, "textImgSetWindow", func(*lua.LState) int {
		ts, ok := toUserData(l, 1).(*TextSprite)
		if !ok {
			userDataError(l, 1, ts)
		}
		ts.SetWindow(float32(numArg(l, 2))/sys.luaSpriteScale+sys.luaSpriteOffsetX, float32(numArg(l, 3))/sys.luaSpriteScale,
			float32(numArg(l, 4))/sys.luaSpriteScale, float32(numArg(l, 5))/sys.luaSpriteScale)
		return 0
	})
	luaRegister(l, "toggleClsnDraw", func(*lua.LState) int {
		if !sys.allowDebugMode {
			return 0
		}
		if l.GetTop() >= 1 {
			sys.clsnDraw = boolArg(l, 1)
		} else {
			sys.clsnDraw = !sys.clsnDraw
		}
		return 0
	})
	luaRegister(l, "toggleDebugDraw", func(*lua.LState) int {
		if !sys.allowDebugMode {
			return 0
		}
		if l.GetTop() >= 1 {
			sys.debugDraw = !sys.debugDraw
			return 0
		}
		if !sys.debugDraw {
			sys.debugDraw = true
		} else {
			pn := sys.debugRef[0]
			hn := sys.debugRef[1]
			for i := hn + 1; i <= len(sys.chars[pn]); i++ {
				hn = i
				if hn >= len(sys.chars[pn]) {
					pn += 1
					hn = 0
					break
				}
				if sys.chars[pn][hn] != nil && !sys.chars[pn][hn].sf(CSF_destroy) {
					break
				}
			}
			ok := false
			for pn < len(sys.chars) {
				if len(sys.chars[pn]) > 0 {
					ok = true
					break
				}
				pn += 1

			}
			if !ok {
				pn = 0
				hn = 0
				sys.debugDraw = false
			}
			sys.debugRef[0] = pn
			sys.debugRef[1] = hn
		}
		return 0
	})
	luaRegister(l, "toggleDialogueBars", func(*lua.LState) int {
		if l.GetTop() >= 1 {
			sys.dialogueBarsFlg = boolArg(l, 1)
		} else {
			sys.dialogueBarsFlg = !sys.dialogueBarsFlg
		}
		return 0
	})
	luaRegister(l, "toggleFullscreen", func(*lua.LState) int {
		fs := !sys.window.fullscreen
		if l.GetTop() >= 1 {
			fs = boolArg(l, 1)
		}
		if fs != sys.window.fullscreen {
			sys.window.toggleFullscreen()
		}
		return 0
	})
	luaRegister(l, "toggleMaxPowerMode", func(*lua.LState) int {
		if l.GetTop() >= 1 {
			sys.maxPowerMode = boolArg(l, 1)
		} else {
			sys.maxPowerMode = !sys.maxPowerMode
		}
		if sys.maxPowerMode {
			for _, c := range sys.chars {
				if len(c) > 0 {
					c[0].power = c[0].powerMax
				}
			}
		}
		return 0
	})
	luaRegister(l, "toggleNoSound", func(*lua.LState) int {
		if l.GetTop() >= 1 {
			sys.noSoundFlg = boolArg(l, 1)
		} else {
			sys.noSoundFlg = !sys.noSoundFlg
		}
		return 0
	})
	luaRegister(l, "togglePause", func(*lua.LState) int {
		if l.GetTop() >= 1 {
			sys.paused = boolArg(l, 1)
		} else {
			sys.paused = !sys.paused
		}
		return 0
	})
	luaRegister(l, "togglePlayer", func(*lua.LState) int {
		pn := int(numArg(l, 1))
		if pn < 1 || pn > len(sys.chars) || len(sys.chars[pn-1]) == 0 {
			return 0
		}
		for _, ch := range sys.chars[pn-1] {
			if ch.scf(SCF_disabled) {
				ch.unsetSCF(SCF_disabled)
			} else {
				ch.setSCF(SCF_disabled)
			}
		}
		return 0
	})
	luaRegister(l, "togglePostMatch", func(*lua.LState) int {
		if l.GetTop() >= 1 {
			sys.postMatchFlg = boolArg(l, 1)
		} else {
			sys.postMatchFlg = !sys.postMatchFlg
		}
		return 0
	})
	luaRegister(l, "toggleStatusDraw", func(*lua.LState) int {
		if l.GetTop() >= 1 {
			sys.statusDraw = boolArg(l, 1)
		} else {
			sys.statusDraw = !sys.statusDraw
		}
		return 0
	})
	luaRegister(l, "toggleVsync", func(*lua.LState) int {
		if l.GetTop() >= 1 {
			sys.vRetrace = int(numArg(l, 1))
		} else if sys.vRetrace == 0 {
			sys.vRetrace = 1
		} else {
			sys.vRetrace = 0
		}
		sys.window.SetSwapInterval(sys.vRetrace)
		return 0
	})
	luaRegister(l, "updateVolume", func(l *lua.LState) int {
		if l.GetTop() >= 1 {
			sys.bgm.bgmVolume = int(Min(int32(numArg(l, 1)), int32(sys.maxBgmVolume)))
		}
		sys.bgm.UpdateVolume()
		return 0
	})
	luaRegister(l, "wavePlay", func(l *lua.LState) int {
		s, ok := toUserData(l, 1).(*Sound)
		if !ok {
			userDataError(l, 1, s)
		}
		sys.soundChannels.Play(s, 100, 0.0)
		return 0
	})
}

// Trigger Functions
func triggerFunctions(l *lua.LState) {
	sys.debugWC = newChar(0, 0)
	// redirection
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
		ret, n := false, int32(0)
		if l.GetTop() >= 1 {
			n = int32(numArg(l, 1))
		}
		if c := sys.debugWC.partner(n, true); c != nil {
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
	luaRegister(l, "p2", func(*lua.LState) int {
		ret := false
		if c := sys.debugWC.p2(); c != nil {
			sys.debugWC, ret = c, true
		}
		l.Push(lua.LBool(ret))
		return 1
	})
	// vanilla triggers
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
	//animelem (deprecated by animelemtime)
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
		switch strings.ToLower(strArg(l, 1)) {
		case "data.life":
			ln = lua.LNumber(c.gi().data.life)
		case "data.power":
			ln = lua.LNumber(c.gi().data.power)
		case "data.guardpoints":
			ln = lua.LNumber(c.gi().data.guardpoints)
		case "data.dizzypoints":
			ln = lua.LNumber(c.gi().data.dizzypoints)
		case "data.attack":
			ln = lua.LNumber(c.gi().data.attack)
		case "data.defence":
			ln = lua.LNumber(c.gi().data.defence)
		case "data.fall.defence_mul":
			ln = lua.LNumber(1.0 / c.gi().data.fall.defence_mul)
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
		case "size.z.width":
			ln = lua.LNumber(c.size.z.width)
		case "size.z.enable":
			ln = lua.LNumber(Btoi(c.size.z.enable))
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
		default:
			ln = lua.LNumber(c.gi().constants[strings.ToLower(strArg(l, 1))])
		}
		l.Push(ln)
		return 1
	})
	luaRegister(l, "const240p", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.constp(320, float32(numArg(l, 1))).ToF()))
		return 1
	})
	luaRegister(l, "const480p", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.constp(640, float32(numArg(l, 1))).ToF()))
		return 1
	})
	luaRegister(l, "const720p", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.constp(1280, float32(numArg(l, 1))).ToF()))
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
		l.Push(lua.LNumber(sys.gameTime + sys.preFightTime))
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
		case "air.animtype":
			ln = lua.LNumber(c.ghv.airanimtype)
		case "ground.animtype":
			ln = lua.LNumber(c.ghv.groundanimtype)
		case "fall.animtype":
			ln = lua.LNumber(c.ghv.fall.animtype)
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
		case "attr":
			ln = lua.LNumber(c.ghv.attr)
		case "dizzypoints":
			ln = lua.LNumber(c.ghv.dizzypoints)
		case "guardpoints":
			ln = lua.LNumber(c.ghv.guardpoints)
		case "id":
			ln = lua.LNumber(c.ghv.id)
		case "playerno":
			ln = lua.LNumber(c.ghv.playerNo)
		case "redlife":
			ln = lua.LNumber(c.ghv.redlife)
		case "score":
			ln = lua.LNumber(c.ghv.score)
		case "hitdamage":
			ln = lua.LNumber(c.ghv.hitdamage)
		case "guarddamage":
			ln = lua.LNumber(c.ghv.guarddamage)
		case "hitpower":
			ln = lua.LNumber(c.ghv.hitpower)
		case "guardpower":
			ln = lua.LNumber(c.ghv.guardpower)
		default:
			l.RaiseError("\nInvalid argument: %v\n", strArg(l, 1))
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
		l.Push(lua.LBool(sys.debugWC.teamside == sys.home))
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
	luaRegister(l, "movereversed", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.moveReversed()))
		return 1
	})
	//name also returns p1name-p8name variants and helpername
	luaRegister(l, "name", func(*lua.LState) int {
		n := int32(1)
		if l.GetTop() >= 1 {
			n = int32(numArg(l, 1))
		}
		if n <= 2 {
			l.Push(lua.LString(sys.debugWC.name))
		} else if ^n&1+1 == 1 {
			if p := sys.debugWC.partner(n/2-1, false); p != nil {
				l.Push(lua.LString(p.name))
			} else {
				l.Push(lua.LString(""))
			}
		} else {
			if p := sys.charList.enemyNear(sys.debugWC, n/2-1, true, true, false); p != nil &&
				!(p.scf(SCF_ko) && p.scf(SCF_over)) {
				l.Push(lua.LString(p.name))
			} else {
				l.Push(lua.LString(""))
			}
		}
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
	//p1name and other variants can be checked via name
	luaRegister(l, "p2bodydistX", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.p2BodyDistX(sys.debugWC).ToI()))
		return 1
	})
	luaRegister(l, "p2bodydistY", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.rdDistY(sys.debugWC.p2(), sys.debugWC).ToI()))
		return 1
	})
	luaRegister(l, "p2distX", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.rdDistX(sys.debugWC.p2(), sys.debugWC).ToI()))
		return 1
	})
	luaRegister(l, "p2distY", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.rdDistY(sys.debugWC.p2(), sys.debugWC).ToI()))
		return 1
	})
	luaRegister(l, "p2life", func(*lua.LState) int {
		if p2 := sys.debugWC.p2(); p2 != nil {
			l.Push(lua.LNumber(p2.life))
		} else {
			l.Push(lua.LNumber(-1))
		}
		return 1
	})
	luaRegister(l, "p2movetype", func(*lua.LState) int {
		if p2 := sys.debugWC.p2(); p2 != nil {
			var s string
			switch p2.ss.moveType {
			case MT_I:
				s = "I"
			case MT_A:
				s = "A"
			case MT_H:
				s = "H"
			}
			l.Push(lua.LString(s))
		} else {
			l.Push(lua.LString(""))
		}
		return 1
	})
	luaRegister(l, "p2stateno", func(*lua.LState) int {
		if p2 := sys.debugWC.p2(); p2 != nil {
			l.Push(lua.LNumber(p2.ss.no))
		}
		return 1
	})
	luaRegister(l, "p2statetype", func(*lua.LState) int {
		if p2 := sys.debugWC.p2(); p2 != nil {
			var s string
			switch p2.ss.stateType {
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
		} else {
			l.Push(lua.LString(""))
		}
		return 1
	})
	luaRegister(l, "palno", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.gi().palno))
		return 1
	})
	luaRegister(l, "parentdistX", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.rdDistX(sys.debugWC.parent(), sys.debugWC).ToI()))
		return 1
	})
	luaRegister(l, "parentdistY", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.rdDistY(sys.debugWC.parent(), sys.debugWC).ToI()))
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
	luaRegister(l, "posZ", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.pos[2]))
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
	//projcontact (deprecated by projcontacttime)
	luaRegister(l, "projcontacttime", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.projContactTime(
			BytecodeInt(int32(numArg(l, 1)))).ToI()))
		return 1
	})
	//projguarded (deprecated by projguardedtime)
	luaRegister(l, "projguardedtime", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.projGuardedTime(
			BytecodeInt(int32(numArg(l, 1)))).ToI()))
		return 1
	})
	//projhit (deprecated by projhittime)
	luaRegister(l, "projhittime", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.projHitTime(
			BytecodeInt(int32(numArg(l, 1)))).ToI()))
		return 1
	})
	//luaRegister(l, "random", func(*lua.LState) int {
	//	l.Push(lua.LNumber(Rand(0, 999)))
	//	return 1
	//})
	luaRegister(l, "rightedge", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.rightEdge()))
		return 1
	})
	luaRegister(l, "rootdistX", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.rdDistX(sys.debugWC.root(), sys.debugWC).ToI()))
		return 1
	})
	luaRegister(l, "rootdistY", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.rdDistY(sys.debugWC.root(), sys.debugWC).ToI()))
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
		switch strArg(l, 1) {
		case "info.name":
			l.Push(lua.LString(sys.stage.name))
		case "info.displayname":
			l.Push(lua.LString(sys.stage.displayname))
		case "info.author":
			l.Push(lua.LString(sys.stage.author))
		case "camera.boundleft":
			l.Push(lua.LNumber(sys.stage.stageCamera.boundleft))
		case "camera.boundright":
			l.Push(lua.LNumber(sys.stage.stageCamera.boundright))
		case "camera.boundhigh":
			l.Push(lua.LNumber(sys.stage.stageCamera.boundhigh))
		case "camera.boundlow":
			l.Push(lua.LNumber(sys.stage.stageCamera.boundlow))
		case "camera.verticalfollow":
			l.Push(lua.LNumber(sys.stage.stageCamera.verticalfollow))
		case "camera.floortension":
			l.Push(lua.LNumber(sys.stage.stageCamera.floortension))
		case "camera.tensionhigh":
			l.Push(lua.LNumber(sys.stage.stageCamera.tensionhigh))
		case "camera.tensionlow":
			l.Push(lua.LNumber(sys.stage.stageCamera.tensionlow))
		case "camera.tension":
			l.Push(lua.LNumber(sys.stage.stageCamera.tension))
		case "camera.startzoom":
			l.Push(lua.LNumber(sys.stage.stageCamera.startzoom))
		case "camera.zoomout":
			l.Push(lua.LNumber(sys.stage.stageCamera.zoomout))
		case "camera.zoomin":
			l.Push(lua.LNumber(sys.stage.stageCamera.zoomin))
		case "camera.ytension.enable":
			l.Push(lua.LBool(sys.stage.stageCamera.ytensionenable))
		case "playerinfo.leftbound":
			l.Push(lua.LNumber(sys.stage.leftbound))
		case "playerinfo.rightbound":
			l.Push(lua.LNumber(sys.stage.rightbound))
		case "scaling.topscale":
			l.Push(lua.LNumber(sys.stage.stageCamera.ztopscale))
		case "bound.screenleft":
			l.Push(lua.LNumber(sys.stage.screenleft))
		case "bound.screenright":
			l.Push(lua.LNumber(sys.stage.screenright))
		case "stageinfo.zoffset":
			l.Push(lua.LNumber(sys.stage.stageCamera.zoffset))
		case "stageinfo.zoffsetlink":
			l.Push(lua.LNumber(sys.stage.zoffsetlink))
		case "stageinfo.xscale":
			l.Push(lua.LNumber(sys.stage.scale[0]))
		case "stageinfo.yscale":
			l.Push(lua.LNumber(sys.stage.scale[1]))
		case "shadow.intensity":
			l.Push(lua.LNumber(sys.stage.sdw.intensity))
		case "shadow.color.r":
			l.Push(lua.LNumber(int32((sys.stage.sdw.color & 0xFF0000) >> 16)))
		case "shadow.color.g":
			l.Push(lua.LNumber(int32((sys.stage.sdw.color & 0xFF00) >> 8)))
		case "shadow.color.b":
			l.Push(lua.LNumber(int32(sys.stage.sdw.color & 0xFF)))
		case "shadow.yscale":
			l.Push(lua.LNumber(sys.stage.sdw.yscale))
		case "shadow.fade.range.begin":
			l.Push(lua.LNumber(sys.stage.sdw.fadebgn))
		case "shadow.fade.range.end":
			l.Push(lua.LNumber(sys.stage.sdw.fadeend))
		case "shadow.xshear":
			l.Push(lua.LNumber(sys.stage.sdw.xshear))
		case "reflection.intensity":
			l.Push(lua.LNumber(sys.stage.reflection))
		default:
			l.Push(lua.LString(""))
		}
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
		case TM_Tag:
			s = "tag"
		}
		l.Push(lua.LString(s))
		return 1
	})
	luaRegister(l, "teamside", func(*lua.LState) int {
		l.Push(lua.LNumber(int32(sys.debugWC.teamside) + 1))
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
	luaRegister(l, "timemod", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.ss.time % int32(numArg(l, 1))))
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
	luaRegister(l, "velZ", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.vel[2]))
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
	luaRegister(l, "winspecial", func(*lua.LState) int {
		l.Push(lua.LBool(sys.debugWC.winType(WT_S)))
		return 1
	})
	luaRegister(l, "winhyper", func(*lua.LState) int {
		l.Push(lua.LBool(sys.debugWC.winType(WT_H)))
		return 1
	})

	// new triggers
	luaRegister(l, "animelemlength", func(*lua.LState) int {
		if f := sys.debugWC.anim.CurrentFrame(); f != nil {
			l.Push(lua.LNumber(f.Time))
		} else {
			l.Push(lua.LNumber(0))
		}
		return 1
	})
	luaRegister(l, "animlength", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.anim.totaltime))
		return 1
	})
	luaRegister(l, "combocount", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.comboCount()))
		return 1
	})
	luaRegister(l, "consecutivewins", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.consecutiveWins[sys.debugWC.teamside]))
		return 1
	})
	luaRegister(l, "dizzy", func(*lua.LState) int {
		l.Push(lua.LBool(sys.debugWC.scf(SCF_dizzy)))
		return 1
	})
	luaRegister(l, "dizzypoints", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.dizzyPoints))
		return 1
	})
	luaRegister(l, "dizzypointsmax", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.dizzyPointsMax))
		return 1
	})
	luaRegister(l, "fighttime", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.gameTime))
		return 1
	})
	luaRegister(l, "firstattack", func(*lua.LState) int {
		l.Push(lua.LBool(sys.debugWC.firstAttack))
		return 1
	})
	luaRegister(l, "framespercount", func(l *lua.LState) int {
		l.Push(lua.LNumber(sys.lifebar.ti.framespercount))
		return 1
	})
	luaRegister(l, "gamemode", func(*lua.LState) int {
		if l.GetTop() == 0 {
			l.Push(lua.LString(sys.gameMode))
			return 1
		}
		l.Push(lua.LBool(sys.gameMode == strArg(l, 1)))
		return 1
	})
	luaRegister(l, "getplayerid", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.getPlayerID(int(numArg(l, 1)))))
		return 1
	})
	luaRegister(l, "guardbreak", func(*lua.LState) int {
		l.Push(lua.LBool(sys.debugWC.scf(SCF_guardbreak)))
		return 1
	})
	luaRegister(l, "guardpoints", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.guardPoints))
		return 1
	})
	luaRegister(l, "guardpointsmax", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.guardPointsMax))
		return 1
	})
	luaRegister(l, "hitoverridden", func(*lua.LState) int {
		l.Push(lua.LBool(sys.debugWC.hoIdx >= 0))
		return 1
	})
	luaRegister(l, "incustomstate", func(*lua.LState) int {
		l.Push(lua.LBool(sys.debugWC.ss.sb.playerNo != sys.debugWC.playerNo))
		return 1
	})
	luaRegister(l, "indialogue", func(*lua.LState) int {
		l.Push(lua.LBool(sys.dialogueFlg))
		return 1
	})
	luaRegister(l, "isasserted", func(*lua.LState) int {
		switch strArg(l, 1) {
		// CharSpecialFlag
		case "nostandguard":
			l.Push(lua.LBool(sys.debugWC.sf(CSF_nostandguard)))
		case "nocrouchguard":
			l.Push(lua.LBool(sys.debugWC.sf(CSF_nocrouchguard)))
		case "noairguard":
			l.Push(lua.LBool(sys.debugWC.sf(CSF_noairguard)))
		case "noshadow":
			l.Push(lua.LBool(sys.debugWC.sf(CSF_noshadow)))
		case "invisible":
			l.Push(lua.LBool(sys.debugWC.sf(CSF_invisible)))
		case "unguardable":
			l.Push(lua.LBool(sys.debugWC.sf(CSF_unguardable)))
		case "nojugglecheck":
			l.Push(lua.LBool(sys.debugWC.sf(CSF_nojugglecheck)))
		case "noautoturn":
			l.Push(lua.LBool(sys.debugWC.sf(CSF_noautoturn)))
		case "nowalk":
			l.Push(lua.LBool(sys.debugWC.sf(CSF_nowalk)))
		case "nobrake":
			l.Push(lua.LBool(sys.debugWC.sf(CSF_nobrake)))
		case "nocrouch":
			l.Push(lua.LBool(sys.debugWC.sf(CSF_nocrouch)))
		case "nostand":
			l.Push(lua.LBool(sys.debugWC.sf(CSF_nostand)))
		case "nojump":
			l.Push(lua.LBool(sys.debugWC.sf(CSF_nojump)))
		case "noairjump":
			l.Push(lua.LBool(sys.debugWC.sf(CSF_noairjump)))
		case "nohardcodedkeys":
			l.Push(lua.LBool(sys.debugWC.sf(CSF_nohardcodedkeys)))
		case "nogetupfromliedown":
			l.Push(lua.LBool(sys.debugWC.sf(CSF_nogetupfromliedown)))
		case "nofastrecoverfromliedown":
			l.Push(lua.LBool(sys.debugWC.sf(CSF_nofastrecoverfromliedown)))
		case "nofallcount":
			l.Push(lua.LBool(sys.debugWC.sf(CSF_nofallcount)))
		case "nofalldefenceup":
			l.Push(lua.LBool(sys.debugWC.sf(CSF_nofalldefenceup)))
		case "noturntarget":
			l.Push(lua.LBool(sys.debugWC.sf(CSF_noturntarget)))
		case "noinput":
			l.Push(lua.LBool(sys.debugWC.sf(CSF_noinput)))
		case "nopowerbardisplay":
			l.Push(lua.LBool(sys.debugWC.sf(CSF_nopowerbardisplay)))
		case "autoguard":
			l.Push(lua.LBool(sys.debugWC.sf(CSF_autoguard)))
		case "animfreeze":
			l.Push(lua.LBool(sys.debugWC.sf(CSF_animfreeze)))
		case "postroundinput":
			l.Push(lua.LBool(sys.debugWC.sf(CSF_postroundinput)))
		case "nodizzypointsdamage":
			l.Push(lua.LBool(sys.debugWC.sf(CSF_nodizzypointsdamage)))
		case "noguardpointsdamage":
			l.Push(lua.LBool(sys.debugWC.sf(CSF_noguardpointsdamage)))
		case "noredlifedamage":
			l.Push(lua.LBool(sys.debugWC.sf(CSF_noredlifedamage)))
		case "nomakedust":
			l.Push(lua.LBool(sys.debugWC.sf(CSF_nomakedust)))
		// GlobalSpecialFlag
		case "intro":
			l.Push(lua.LBool(sys.sf(GSF_intro)))
		case "roundnotover":
			l.Push(lua.LBool(sys.sf(GSF_roundnotover)))
		case "nomusic":
			l.Push(lua.LBool(sys.sf(GSF_nomusic)))
		case "nobardisplay":
			l.Push(lua.LBool(sys.sf(GSF_nobardisplay)))
		case "nobg":
			l.Push(lua.LBool(sys.sf(GSF_nobg)))
		case "nofg":
			l.Push(lua.LBool(sys.sf(GSF_nofg)))
		case "globalnoshadow":
			l.Push(lua.LBool(sys.sf(GSF_globalnoshadow)))
		case "timerfreeze":
			l.Push(lua.LBool(sys.sf(GSF_timerfreeze)))
		case "nokosnd":
			l.Push(lua.LBool(sys.sf(GSF_nokosnd)))
		case "nokoslow":
			l.Push(lua.LBool(sys.sf(GSF_nokoslow)))
		case "noko":
			l.Push(lua.LBool(sys.sf(GSF_noko)))
		case "nokovelocity":
			l.Push(lua.LBool(sys.sf(GSF_nokovelocity)))
		case "roundnotskip":
			l.Push(lua.LBool(sys.sf(GSF_roundnotskip)))
		case "roundfreeze":
			l.Push(lua.LBool(sys.sf(GSF_roundfreeze)))
		// SystemCharFlag
		case "over":
			l.Push(lua.LBool(sys.debugWC.scf(SCF_over)))
		case "koroundmiddle":
			l.Push(lua.LBool(sys.debugWC.scf(SCF_ko_round_middle)))
		case "disabled":
			l.Push(lua.LBool(sys.debugWC.scf(SCF_disabled)))
		default:
			l.RaiseError("\nInvalid argument: %v\n", strArg(l, 1))
		}
		return 1
	})
	luaRegister(l, "ishost", func(*lua.LState) int {
		l.Push(lua.LBool(sys.debugWC.isHost()))
		return 1
	})
	luaRegister(l, "localscale", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.localscl))
		return 1
	})
	luaRegister(l, "majorversion", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.gi().ver[0]))
		return 1
	})
	luaRegister(l, "map", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.mapArray[strings.ToLower(strArg(l, 1))]))
		return 1
	})
	luaRegister(l, "memberno", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.memberNo + 1))
		return 1
	})
	luaRegister(l, "movecountered", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.moveCountered()))
		return 1
	})
	luaRegister(l, "pausetime", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.pauseTime()))
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
	luaRegister(l, "playerno", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.playerNo + 1))
		return 1
	})
	luaRegister(l, "ratiolevel", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.ocd().ratioLevel))
		return 1
	})
	luaRegister(l, "receivedhits", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.receivedHits))
		return 1
	})
	luaRegister(l, "receiveddamage", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.comboDmg))
		return 1
	})
	luaRegister(l, "redlife", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.redLife))
		return 1
	})
	luaRegister(l, "roundtype", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.roundType()))
		return 1
	})
	luaRegister(l, "score", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.score()))
		return 1
	})
	luaRegister(l, "scoretotal", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.scoreTotal()))
		return 1
	})
	luaRegister(l, "selfstatenoexist", func(*lua.LState) int {
		l.Push(lua.LBool(sys.debugWC.selfStatenoExist(
			BytecodeInt(int32(numArg(l, 1)))).ToB()))
		return 1
	})
	luaRegister(l, "sprpriority", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.sprPriority))
		return 1
	})
	luaRegister(l, "stagebackedge", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.stageBackEdge()))
		return 1
	})
	luaRegister(l, "stageconst", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.stage.constants[strArg(l, 1)]))
		return 1
	})
	luaRegister(l, "stagefrontedge", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.stageFrontEdge()))
		return 1
	})
	luaRegister(l, "stagetime", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.stage.stageTime))
		return 1
	})
	luaRegister(l, "standby", func(*lua.LState) int {
		l.Push(lua.LBool(sys.debugWC.scf(SCF_standby)))
		return 1
	})
	luaRegister(l, "teamleader", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.teamLeader()))
		return 1
	})
	luaRegister(l, "teamsize", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.teamSize()))
		return 1
	})
	luaRegister(l, "timeelapsed", func(*lua.LState) int {
		l.Push(lua.LNumber(timeElapsed()))
		return 1
	})
	luaRegister(l, "timeremaining", func(*lua.LState) int {
		l.Push(lua.LNumber(timeRemaining()))
		return 1
	})
	luaRegister(l, "timetotal", func(*lua.LState) int {
		l.Push(lua.LNumber(timeTotal()))
		return 1
	})

	// lua/debug only triggers
	luaRegister(l, "animelemcount", func(*lua.LState) int {
		l.Push(lua.LNumber(len(sys.debugWC.anim.frames)))
		return 1
	})
	luaRegister(l, "animelemtimesum", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.anim.time))
		return 1
	})
	luaRegister(l, "animtimesum", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.anim.sumtime))
		return 1
	})
	luaRegister(l, "animowner", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.animPN) + 1)
		return 1
	})
	luaRegister(l, "attack", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.attackMul * 100))
		return 1
	})
	luaRegister(l, "continue", func(*lua.LState) int {
		l.Push(lua.LBool(sys.continueFlg))
		return 1
	})
	luaRegister(l, "defence", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.finalDefense * 100))
		return 1
	})
	luaRegister(l, "displayname", func(*lua.LState) int {
		l.Push(lua.LString(sys.debugWC.gi().displayname))
		return 1
	})
	luaRegister(l, "gameend", func(*lua.LState) int {
		l.Push(lua.LBool(sys.gameEnd))
		return 1
	})
	luaRegister(l, "gamespeed", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.gameSpeed * sys.accel * 100))
		return 1
	})
	luaRegister(l, "gameLogicSpeed", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.gameSpeed * sys.accel * float32(FPS)))
		return 1
	})
	luaRegister(l, "lasthitter", func(*lua.LState) int {
		tn := int(numArg(l, 1))
		if tn < 1 || tn > 2 {
			l.RaiseError("\nInvalid team side: %v\n", tn)
		}
		l.Push(lua.LNumber(sys.lastHitter[tn-1] + 1))
		return 1
	})
	luaRegister(l, "localcoord", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.localcoord))
		return 1
	})
	luaRegister(l, "matchtime", func(*lua.LState) int {
		var ti int32
		for _, v := range sys.timerRounds {
			ti += v
		}
		l.Push(lua.LNumber(ti))
		return 1
	})
	luaRegister(l, "network", func(*lua.LState) int {
		l.Push(lua.LBool(sys.netInput != nil || sys.fileInput != nil))
		return 1
	})
	luaRegister(l, "paused", func(*lua.LState) int {
		l.Push(lua.LBool(sys.paused && !sys.step))
		return 1
	})
	luaRegister(l, "roundover", func(*lua.LState) int {
		l.Push(lua.LBool(sys.roundOver()))
		return 1
	})
	luaRegister(l, "roundstart", func(*lua.LState) int {
		l.Push(lua.LBool(sys.tickCount == 1))
		return 1
	})
	luaRegister(l, "selectno", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.selectNo))
		return 1
	})
	luaRegister(l, "spritegroup", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.curFrame.Group))
		return 1
	})
	luaRegister(l, "spritenumber", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.curFrame.Number))
		return 1
	})
	luaRegister(l, "stateowner", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.ss.sb.playerNo + 1))
		return 1
	})
	luaRegister(l, "stateownerid", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.chars[sys.debugWC.ss.sb.playerNo][0].id))
		return 1
	})
	luaRegister(l, "stateownername", func(*lua.LState) int {
		l.Push(lua.LString(sys.chars[sys.debugWC.ss.sb.playerNo][0].name))
		return 1
	})
	luaRegister(l, "tickcount", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.tickCount))
		return 1
	})
	luaRegister(l, "vsync", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.vRetrace))
		return 1
	})
	luaRegister(l, "winnerteam", func(*lua.LState) int {
		var winp int32 = -1
		if !sys.endMatch {
			if sys.matchOver() && sys.roundOver() {
				w1 := sys.wins[0] >= sys.matchWins[0]
				w2 := sys.wins[1] >= sys.matchWins[1]
				if w1 != w2 {
					winp = Btoi(w1) + Btoi(w2)*2
				} else {
					winp = 0
				}
			} else if sys.winTeam >= 0 || sys.debugWC.roundState() >= 3 {
				winp = int32(sys.winTeam) + 1
			}
		}
		l.Push(lua.LNumber(winp))
		return 1
	})
}
