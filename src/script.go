package main

import (
	"fmt"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
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
func nilArg(l *lua.LState, argi int) bool {
	//if l.GetTop() < argi {
	//	return true
	//}
	lv := l.Get(argi)
	return lua.LVIsFalse(lv) && lv != lua.LFalse
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

// Converts a hitflag to a LString.
// Previously in hitdefvar, moved here
// for reusability in gethitvar
func flagLStr(flag int32) lua.LString {
	str := ""
	if flag&int32(HF_H) != 0 {
		str += "H"
	}
	if flag&int32(HF_L) != 0 {
		str += "L"
	}
	if flag&int32(HF_A) != 0 {
		str += "A"
	}
	if flag&int32(HF_F) != 0 {
		str += "F"
	}
	if flag&int32(HF_D) != 0 {
		str += "D"
	}
	if flag&int32(HF_P) != 0 {
		str += "P"
	}
	if flag&int32(HF_MNS) != 0 {
		str += "-"
	}
	if flag&int32(HF_PLS) != 0 {
		str += "+"
	}
	return lua.LString(str)
}

// Converts an attr (statetype, attacktype) to a LString.
// Used in gethitvar("attr.flag").
func attrLStr(attr int32) lua.LString {
	str := ""
	if attr == 0 {
		return lua.LString(str)
	} // no attr? return an empty string
	st := attr & int32(ST_MASK)  // state type
	at := attr & ^int32(ST_MASK) // attack type (everything that's not statetype)
	// flag1
	if st&int32(ST_S) != 0 {
		str += "S"
	}
	if st&int32(ST_C) != 0 {
		str += "C"
	}
	if st&int32(ST_A) != 0 {
		str += "A"
	}
	str += ", "
	// first char
	if at&int32(AT_AN) != 0 {
		str += "N"
	}
	if at&int32(AT_AS) != 0 {
		str += "S"
	}
	if at&int32(AT_AH) != 0 {
		str += "H"
	}
	// second char
	if at&int32(AT_AA) != 0 {
		str += "A"
	}
	if at&int32(AT_AT) != 0 {
		str += "T"
	}
	if at&int32(AT_AP) != 0 {
		str += "P"
	}

	return lua.LString(str)
}

func toLValue(l *lua.LState, v interface{}) lua.LValue {
	rv := reflect.ValueOf(v)

	// Handle pointer types
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return lua.LNil
		}
		rv = rv.Elem() // Dereference pointer
	}

	switch rv.Kind() {
	case reflect.Struct:
		table := l.NewTable()
		for i := 0; i < rv.NumField(); i++ {
			field := rv.Type().Field(i)

			// **Skip unexported fields**
			if field.PkgPath != "" {
				continue
			}

			fieldValue := rv.Field(i)

			// Use 'lua' tag as the key, fall back to 'ini' tag, then field name
			key := field.Tag.Get("lua")
			if key == "" {
				key = field.Tag.Get("ini")
			}
			if key == "" {
				key = field.Name
			}

			if fieldValue.Kind() == reflect.Ptr && fieldValue.IsNil() {
				continue // Skip nil pointers
			}

			// Recursively convert field value
			table.RawSetString(key, toLValue(l, fieldValue.Interface()))
		}
		return table

	case reflect.Map:
		table := l.NewTable()
		for _, key := range rv.MapKeys() {
			value := rv.MapIndex(key)
			luaKey := lua.LString(fmt.Sprintf("%v", key.Interface())) // Convert map key to string
			table.RawSet(luaKey, toLValue(l, value.Interface()))
		}
		return table

	case reflect.Array, reflect.Slice:
		table := l.NewTable()
		for i := 0; i < rv.Len(); i++ {
			table.Append(toLValue(l, rv.Index(i).Interface()))
		}
		return table

	case reflect.String:
		return lua.LString(rv.String())

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return lua.LNumber(rv.Int())

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return lua.LNumber(rv.Uint())

	case reflect.Float32, reflect.Float64:
		return lua.LNumber(RoundFloat(rv.Float(), 6))

	case reflect.Bool:
		if rv.Bool() {
			return lua.LTrue
		}
		return lua.LFalse

	default:
		// Fallback for unsupported types
		return lua.LString(fmt.Sprintf("%v", rv.Interface()))
	}
}

// -------------------------------------------------------------------------------------------------
// Register external functions to be called from Lua scripts
func systemScriptInit(l *lua.LState) {
	triggerFunctions(l)
	deprecatedFunctions(l)
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
	luaRegister(l, "animGetPreloadedCharData", func(l *lua.LState) int {
		if anim := sys.sel.GetChar(int(numArg(l, 1))).anims.get(int16(numArg(l, 2)), int16(numArg(l, 3))); anim != nil {
			pfx := newPalFX()
			pfx.clear()
			pfx.time = -1
			a := &Anim{anim: anim, window: sys.scrrect, xscl: 1, yscl: 1, palfx: pfx}
			if !nilArg(l, 4) && !boolArg(l, 4) && a.anim.totaltime == a.anim.looptime {
				a.anim.totaltime = -1
				a.anim.looptime = 0
			}
			l.Push(newUserData(l, a))
			return 1
		}
		return 0
	})
	luaRegister(l, "animGetPreloadedStageData", func(l *lua.LState) int {
		if anim := sys.sel.GetStage(int(numArg(l, 1))).anims.get(int16(numArg(l, 2)), int16(numArg(l, 3))); anim != nil {
			pfx := newPalFX()
			pfx.clear()
			pfx.time = -1
			a := &Anim{anim: anim, window: sys.scrrect, xscl: 1, yscl: 1, palfx: pfx}
			if !nilArg(l, 4) && !boolArg(l, 4) && a.anim.totaltime == a.anim.looptime {
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
		if !nilArg(l, 3) {
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
				switch strings.ToLower(string(k)) {
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
						a.palfx.cycletime[0] = -s[3]
					} else {
						a.palfx.sinadd[0] = s[0]
						a.palfx.sinadd[1] = s[1]
						a.palfx.sinadd[2] = s[2]
						a.palfx.cycletime[0] = s[3]
					}
				case "sinmul":
					var s [4]int32
					switch v := value.(type) {
					case *lua.LTable:
						v.ForEach(func(key2, value2 lua.LValue) {
							s[int(lua.LVAsNumber(key2))-1] = int32(lua.LVAsNumber(value2))
						})
					}
					if s[3] < 0 {
						a.palfx.sinmul[0] = -s[0]
						a.palfx.sinmul[1] = -s[1]
						a.palfx.sinmul[2] = -s[2]
						a.palfx.cycletime[1] = -s[3]
					} else {
						a.palfx.sinmul[0] = s[0]
						a.palfx.sinmul[1] = s[1]
						a.palfx.sinmul[2] = s[2]
						a.palfx.cycletime[1] = s[3]
					}
				case "sincolor":
					var s [2]int32
					switch v := value.(type) {
					case *lua.LTable:
						v.ForEach(func(key2, value2 lua.LValue) {
							s[int(lua.LVAsNumber(key2))-1] = int32(lua.LVAsNumber(value2))
						})
					}
					if s[1] < 0 {
						a.palfx.sincolor = -s[0]
						a.palfx.cycletime[2] = -s[1]
					} else {
						a.palfx.sincolor = s[0]
						a.palfx.cycletime[2] = s[1]
					}
				case "sinhue":
					var s [2]int32
					switch v := value.(type) {
					case *lua.LTable:
						v.ForEach(func(key2, value2 lua.LValue) {
							s[int(lua.LVAsNumber(key2))-1] = int32(lua.LVAsNumber(value2))
						})
					}
					if s[1] < 0 {
						a.palfx.sinhue = -s[0]
						a.palfx.cycletime[3] = -s[1]
					} else {
						a.palfx.sinhue = s[0]
						a.palfx.cycletime[3] = s[1]
					}
				case "invertall":
					a.palfx.invertall = lua.LVAsNumber(value) == 1
				case "invertblend":
					a.palfx.invertblend = int32(lua.LVAsNumber(value))
				case "color":
					a.palfx.color = float32(lua.LVAsNumber(value)) / 256
				case "hue":
					a.palfx.hue = float32(lua.LVAsNumber(value)) / 256
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
		if nilArg(l, 4) || boolArg(l, 4) {
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
		var tx, ty int32
		if boolArg(l, 2) {
			tx = 1
		}
		if boolArg(l, 3) {
			ty = 1
		}
		var sx, sy int32
		if !nilArg(l, 4) {
			sx = int32(numArg(l, 4))
			if !nilArg(l, 5) {
				sy = int32(numArg(l, 5))
			} else {
				sy = sx
			}
		}
		a.SetTile(tx, ty, sx, sy)
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
	luaRegister(l, "batchDraw", func(*lua.LState) int {
		tbl := l.ToTable(1)
		if tbl == nil {
			l.RaiseError("batchDraw requires a table as its first argument")
			return 0
		}

		tbl.ForEach(func(_, val lua.LValue) {
			item, ok := val.(*lua.LTable)
			if !ok {
				// l.RaiseError("batchDraw expects a table of tables")
				return
			}

			luaAnim := item.RawGetString("anim")

			ud, ok := luaAnim.(*lua.LUserData)
			if !ok {
				return
			}

			anim, ok := ud.Value.(*Anim)
			if !ok {
				return
			}

			x := float32(lua.LVAsNumber(item.RawGetString("x")))
			y := float32(lua.LVAsNumber(item.RawGetString("y")))

			facing := float32(lua.LVAsNumber(item.RawGetString("facing")))

			anim.SetPos(x/sys.luaSpriteScale+sys.luaSpriteOffsetX, y/sys.luaSpriteScale)
			anim.SetFacing(facing)
			anim.Draw()
		})
		return 0
	})
	luaRegister(l, "bgDraw", func(*lua.LState) int {
		bg, ok := toUserData(l, 1).(*BGDef)
		if !ok {
			userDataError(l, 1, bg)
		}
		layer := int32(0)
		var x, y, scl float32 = 0, 0, 1
		if !nilArg(l, 2) {
			if numArg(l, 2) == 1 {
				layer = 1
			} else {
				layer = 0
			}
		}
		if !nilArg(l, 3) {
			x = float32(numArg(l, 3))
		}
		if !nilArg(l, 4) {
			y = float32(numArg(l, 4))
		}
		if !nilArg(l, 5) {
			scl = float32(numArg(l, 5))
		}
		bg.draw(layer, x, y, scl)
		return 0
	})
	luaRegister(l, "bgNew", func(*lua.LState) int {
		s, ok := toUserData(l, 1).(*Sff)
		if !ok {
			userDataError(l, 1, s)
		}
		model, ok := toUserData(l, 4).(*Model)
		bg, err := loadBGDef(s, model, strArg(l, 2), strArg(l, 3))
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
	luaRegister(l, "changeAnim", func(l *lua.LState) int {
		//anim_no, anim_elem, ffx
		an := int32(numArg(l, 1))
		c := sys.chars[sys.debugWC.playerNo]
		if c[0].selfAnimExist(BytecodeInt(an)) == BytecodeBool(true) {
			ffx := false
			if !nilArg(l, 3) {
				ffx = boolArg(l, 3)
			}
			prefix := ""
			if ffx {
				prefix = "f"
			}
			c[0].changeAnim(an, c[0].playerNo, -1, prefix)
			if !nilArg(l, 2) {
				c[0].setAnimElem(int32(numArg(l, 2)), 0)
			}
			l.Push(lua.LBool(true))
			return 1
		}
		l.Push(lua.LBool(false))
		return 1
	})
	luaRegister(l, "changeColorPalette", func(*lua.LState) int {
		//Changes the actual color of the palette
		a, _ := toUserData(l, 1).(*Anim)
		a.anim.palettedata.paletteMap[0] = int(numArg(l, 2)) - 1
		l.Push(newUserData(l, a))
		return 1
	})
	luaRegister(l, "changeState", func(l *lua.LState) int {
		//state_no
		st := int32(numArg(l, 1))
		c := sys.chars[sys.debugWC.playerNo]
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
			c[0].changeState(st, -1, -1, "")
			l.Push(lua.LBool(true))
			return 1
		}
		l.Push(lua.LBool(false))
		return 1
	})
	luaRegister(l, "charSpriteDraw", func(l *lua.LState) int {
		// pn, spr_tbl (1 or more pairs), x, y, scaleX, scaleY, facing, window
		pn := int(numArg(l, 1))
		if pn < 1 || pn > len(sys.chars) || len(sys.chars[pn-1]) == 0 {
			l.RaiseError("\nPlayer not found: %v\n", pn)
		}
		window := &sys.scrrect
		if !nilArg(l, 11) {
			window = &[...]int32{int32(numArg(l, 8)), int32(numArg(l, 9)), int32(numArg(l, 10)), int32(numArg(l, 11))}
		}
		var ok bool
		var group int16
		tableArg(l, 2).ForEach(func(key, value lua.LValue) {
			if !ok {
				if int(lua.LVAsNumber(key))%2 == 1 {
					group = int16(lua.LVAsNumber(value))
				} else {
					sprite := sys.cgi[pn-1].sff.getOwnPalSprite(group, int16(lua.LVAsNumber(value)), &sys.cgi[pn-1].palettedata.palList)
					if fspr := sprite; fspr != nil {
						pfx := sys.chars[pn-1][0].getPalfx()
						sys.cgi[pn-1].palettedata.palList.SwapPalMap(&pfx.remap)
						fspr.Pal = nil
						fspr.Pal = fspr.GetPal(&sys.cgi[pn-1].palettedata.palList)
						sys.cgi[pn-1].palettedata.palList.SwapPalMap(&pfx.remap)
						x := (float32(numArg(l, 3)) + sys.lifebarOffsetX) * sys.lifebarScale
						y := (float32(numArg(l, 4)) + sys.lifebarOffsetY) * sys.lifebarScale
						scale := [...]float32{float32(numArg(l, 5)), float32(numArg(l, 6))}
						facing := int8(numArg(l, 7))
						fscale := sys.chars[pn-1][0].localscl
						if sprite.coldepth <= 8 && sprite.PalTex == nil {
							sprite.CachePalette(sprite.Pal)
						}
						sprite.Draw(x, y, scale[0]*float32(facing)*fscale, scale[1]*fscale, 0,
							Rotation{0, 0, 0}, pfx, window)
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
		alpha := int32(255)
		if !nilArg(l, 4) {
			alpha = int32(numArg(l, 4))
		}
		col := uint32(int32(numArg(l, 3))&0xff | int32(numArg(l, 2))&0xff<<8 | int32(numArg(l, 1))&0xff<<16)
		src := alpha
		dst := 255 - alpha
		FillRect(sys.scrrect, col, [2]int32{src, dst})
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
	luaRegister(l, "colorPortrait", func(l *lua.LState) int {
		//Creates a duplicate animation, to avoid palette sharing across players, then removes palettes from sprites
		a, ok := toUserData(l, 1).(*Anim)
		if !ok {
			userDataError(l, 1, a)
		}
		if sys.usePalette == true {
			copyAnim := CopyAnim(a)
			for _, c := range copyAnim.anim.frames {
				if copyAnim.anim.sff.sprites[[...]int16{c.Group, c.Number}].palidx == 0 && len(sys.sel.GetChar(int(numArg(l, 2))).pal) > 0 {
					copyAnim.anim.sff.sprites[[...]int16{c.Group, c.Number}].Pal = nil
				}
			}
			l.Push(newUserData(l, copyAnim))
		} else {
			l.Push(newUserData(l, a))
		}
		return 1
	})
	luaRegister(l, "commandAdd", func(l *lua.LState) int {
		cl, ok := toUserData(l, 1).(*CommandList)
		if !ok {
			userDataError(l, 1, cl)
		}

		name := strArg(l, 2)
		cmdstr := strArg(l, 3)

		cm := newCommand()
		cm.name = name
		if err := cm.ReadCommandSymbols(cmdstr, NewCommandKeyRemap()); err != nil {
			l.RaiseError(err.Error())
		}

		time := cl.DefaultTime
		buftime := cl.DefaultBufferTime
		buffer_hitpause := cl.DefaultBufferHitpause
		buffer_pauseend := cl.DefaultBufferPauseEnd
		steptime := cl.DefaultStepTime

		if !nilArg(l, 4) {
			time = int32(numArg(l, 4))
		}
		if !nilArg(l, 5) {
			buftime = Max(1, int32(numArg(l, 5)))
		}
		if !nilArg(l, 6) {
			buffer_hitpause = boolArg(l, 6)
		}
		if !nilArg(l, 7) {
			buffer_pauseend = boolArg(l, 7)
		}
		if !nilArg(l, 8) {
			steptime = int32(numArg(l, 8))
		}

		cm.maxtime = time
		cm.maxbuftime = buftime
		cm.buffer_hitpause = buffer_hitpause
		cm.buffer_pauseend = buffer_pauseend
		cm.maxsteptime = steptime

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
		if !ok || cl == nil {
			userDataError(l, 1, cl)
			l.Push(lua.LBool(false))
			return 1 // Attempt to fix a rare registry overflow error while the window is unfocused
		}
		l.Push(lua.LBool(cl.GetState(strArg(l, 2))))
		return 1
	})
	luaRegister(l, "commandInput", func(l *lua.LState) int {
		cl, ok := toUserData(l, 1).(*CommandList)
		if !ok || cl == nil {
			userDataError(l, 1, cl)
			return 0 // Attempt to fix a rare registry overflow error while the window is unfocused
		}
		controller := int(numArg(l, 2)) - 1
		if cl.InputUpdate(nil, controller, 0, true) {
			cl.Step(false, false, false, false, 0)
		}
		return 0
	})
	luaRegister(l, "commandNew", func(l *lua.LState) int {
		l.Push(newUserData(l, NewCommandList(NewInputBuffer())))
		return 1
	})
	luaRegister(l, "connected", func(*lua.LState) int {
		l.Push(lua.LBool(sys.netConnection.IsConnected())) // No need to check rollback here as this deals with the main menu connection
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
		if sys.netConnection != nil {
			l.RaiseError("\nConnection already established.\n")
		}
		sys.chars = [len(sys.chars)][]*Char{}
		sys.netConnection = NewNetConnection()

		//Rollback only
		if sys.cfg.Netplay.RollbackNetcode {
			rs := NewRollbackSession(sys.cfg.Netplay.Rollback)
			sys.rollback.session = &rs
		}

		if host := strArg(l, 1); host != "" {
			//Rollback only
			if sys.cfg.Netplay.RollbackNetcode {
				sys.rollback.session.host = host
			}

			sys.netConnection.Connect(host, sys.cfg.Netplay.ListenPort)
		} else {
			if err := sys.netConnection.Accept(sys.cfg.Netplay.ListenPort); err != nil {
				l.RaiseError(err.Error())
			}
		}

		return 0
	})
	luaRegister(l, "enterReplay", func(*lua.LState) int {
		if sys.cfg.Video.VSync >= 0 {
			sys.window.SetSwapInterval(1) // broken frame skipping when set to 0
		}
		sys.chars = [len(sys.chars)][]*Char{}
		sys.replayFile = OpenReplayFile(strArg(l, 1))
		return 0
	})
	luaRegister(l, "esc", func(l *lua.LState) int {
		if !nilArg(l, 1) {
			sys.esc = boolArg(l, 1)
		}
		l.Push(lua.LBool(sys.esc))
		return 1
	})
	luaRegister(l, "exitNetPlay", func(*lua.LState) int {
		if sys.cfg.Netplay.RollbackNetcode {
			if sys.rollback.session != nil {
				sys.rollback.session.Close()
				sys.rollback.session = nil
			}
		}
		if sys.netConnection != nil {
			sys.netConnection.Close()
			sys.netConnection = nil
		}
		return 0
	})
	luaRegister(l, "exitReplay", func(*lua.LState) int {
		if sys.cfg.Video.VSync >= 0 {
			sys.window.SetSwapInterval(sys.cfg.Video.VSync)
		}
		if sys.replayFile != nil {
			sys.replayFile.Close()
			sys.replayFile = nil
		}
		return 0
	})
	luaRegister(l, "fade", func(l *lua.LState) int {
		rect := [4]int32{int32(numArg(l, 1)), int32(numArg(l, 2)), int32(numArg(l, 3)), int32(numArg(l, 4))}
		alpha := int32(numArg(l, 5))
		src := alpha>>uint(Btoi(sys.clsnDisplay)) + Btoi(sys.clsnDisplay)*128
		dst := 255 - src
		FillRect(rect, 0, [2]int32{src, dst})
		return 0
	})
	luaRegister(l, "fadeColor", func(l *lua.LState) int {
		if int32(numArg(l, 2)) > sys.frameCounter {
			l.Push(lua.LBool(true)) // delayed fade
			return 1
		}
		frame := float64(sys.frameCounter - int32(numArg(l, 2)))
		length := numArg(l, 3)
		if frame > length || length <= 0 {
			l.Push(lua.LBool(false))
			return 1
		}
		r, g, b, alpha := int32(0), int32(0), int32(0), float64(0)
		if strArg(l, 1) == "fadeout" {
			alpha = math.Floor(float64(255) / length * frame)
		} else if strArg(l, 1) == "fadein" {
			alpha = math.Floor(255 - 255*(frame-1)/length)
		}
		alpha = float64(ClampF(float32(alpha), 0, 255))
		src := int32(alpha)
		dst := 255 - src
		if !nilArg(l, 6) {
			r = int32(numArg(l, 4))
			g = int32(numArg(l, 5))
			b = int32(numArg(l, 6))
		}
		col := uint32(int32(b)&0xff | int32(g)&0xff<<8 | int32(r)&0xff<<16)
		FillRect(sys.scrrect, col, [2]int32{src, dst})
		l.Push(lua.LBool(true))
		return 1
	})
	luaRegister(l, "fileExists", func(l *lua.LState) int {
		path := strArg(l, 1)
		l.Push(lua.LBool(FileExist(path) != ""))
		return 1
	})
	luaRegister(l, "fillRect", func(l *lua.LState) int {
		rect := [4]int32{
			int32((float32(numArg(l, 1))/sys.luaSpriteScale + float32(sys.gameWidth-320)/2 + sys.luaSpriteOffsetX) * sys.widthScale),
			int32((float32(numArg(l, 2))/sys.luaSpriteScale + float32(sys.gameHeight-240)) * sys.heightScale),
			int32((float32(numArg(l, 3)) / sys.luaSpriteScale) * sys.widthScale),
			int32((float32(numArg(l, 4)) / sys.luaSpriteScale) * sys.heightScale),
		}
		col := uint32(int32(numArg(l, 7))&0xff | int32(numArg(l, 6))&0xff<<8 | int32(numArg(l, 5))&0xff<<16)
		src := int32(numArg(l, 8))
		dst := int32(numArg(l, 9))
		FillRect(rect, col, [2]int32{src, dst})
		return 0
	})
	luaRegister(l, "findEntityByPlayerId", func(*lua.LState) int {
		if !sys.debugModeAllowed() {
			return 0
		}

		pid := int32(numArg(l, 1))

		pn := sys.debugRef[0]
		hn := sys.debugRef[1]
		pnStart := pn
		hnStart := hn + 1

		found := false

		// Don't let these loops fool you, this is still iterating
		// over n entities.
		for i := pnStart; i < len(sys.chars) && !found; i++ {
			if !sys.debugDisplay {
				for j := 0; j < len(sys.chars[i]) && !found; j++ {
					if sys.chars[i][j] != nil && (j == 0 || !sys.chars[i][j].csf(CSF_destroy)) {
						if sys.chars[i][j].id == pid {
							sys.debugRef[0] = i
							sys.debugRef[1] = j
							found = true
							sys.debugDisplay = true
							break
						}
					}
				}
			} else {
				for j := hnStart; j < len(sys.chars[i]) && !found; j++ {
					if sys.chars[i][j] != nil && (j == 0 || !sys.chars[i][j].csf(CSF_destroy)) {
						if sys.chars[i][j].id == pid {
							sys.debugRef[0] = i
							sys.debugRef[1] = j
							found = true
							break
						}
					}
				}
				// gotta reset after one full iteration
				hnStart = 0
			}
		}

		// Come back around if we haven't found it and we're not starting from 0
		if (pnStart > 0 || hnStart > 0) && !found {
			for i := 0; i < len(sys.chars) && !found; i++ {
				for j := 0; j < len(sys.chars[i]); j++ {
					if sys.chars[i][j].id == pid {
						sys.debugRef[0] = i
						sys.debugRef[1] = j
						found = true
						break
					}
				}
			}
		}

		if !found {
			l.RaiseError("Could not find an entity matching player ID %d", pid)
		}

		return 0
	})
	luaRegister(l, "findEntityByName", func(*lua.LState) int {
		if !sys.debugModeAllowed() {
			return 0
		}

		text := strings.ToLower(strArg(l, 1))
		nameLower := ""

		pn := sys.debugRef[0]
		hn := sys.debugRef[1]
		pnStart := pn
		hnStart := hn + 1

		found := false

		// Don't let these loops fool you, this is still iterating
		// over n entities.
		for i := pnStart; i < len(sys.chars) && !found; i++ {
			if !sys.debugDisplay {
				for j := 0; j < len(sys.chars[i]) && !found; j++ {
					if sys.chars[i][j] != nil && (j == 0 || !sys.chars[i][j].csf(CSF_destroy)) {
						nameLower = strings.ToLower(sys.chars[i][j].name)
						if strings.Contains(nameLower, text) {
							sys.debugRef[0] = i
							sys.debugRef[1] = j
							found = true
							sys.debugDisplay = true
							break
						}
					}
				}
			} else {
				for j := hnStart; j < len(sys.chars[i]) && !found; j++ {
					if sys.chars[i][j] != nil && !sys.chars[i][j].csf(CSF_destroy) {
						nameLower = strings.ToLower(sys.chars[i][j].name)
						if strings.Contains(nameLower, text) {
							sys.debugRef[0] = i
							sys.debugRef[1] = j
							found = true
							break
						}
					}
				}
				// gotta reset after one full iteration
				hnStart = 0
			}
		}

		// Come back around if we haven't found it and we're not starting from 0
		if (pnStart > 0 || hnStart > 0) && !found {
			for i := 0; i < len(sys.chars) && !found; i++ {
				for j := 0; j < len(sys.chars[i]); j++ {
					nameLower = strings.ToLower(sys.chars[i][j].name)
					if strings.Contains(nameLower, text) {
						sys.debugRef[0] = i
						sys.debugRef[1] = j
						found = true
						break
					}
				}
			}
		}

		if !found {
			l.RaiseError("Could not find an entity matching \"%s\"", nameLower)
		}

		return 0
	})
	luaRegister(l, "findHelperById", func(*lua.LState) int {
		if !sys.debugModeAllowed() {
			return 0
		}

		hid := int32(numArg(l, 1))

		pn := sys.debugRef[0]
		hn := sys.debugRef[1]
		pnStart := pn
		hnStart := hn + 1

		found := false

		// Don't let these loops fool you, this is still iterating over n entities.
		for i := pnStart; i < len(sys.chars) && !found; i++ {
			if !sys.debugDisplay {
				for j := 1; j < len(sys.chars[i]) && !found; j++ {
					if sys.chars[i][j] != nil && (j == 0 || !sys.chars[i][j].csf(CSF_destroy)) {
						if sys.chars[i][j].helperId == hid {
							sys.debugRef[0] = i
							sys.debugRef[1] = j
							found = true
							sys.debugDisplay = true
							break
						}
					}
				}
			} else {
				for j := hnStart; j < len(sys.chars[i]) && !found; j++ {
					if sys.chars[i][j] != nil && (j == 0 || !sys.chars[i][j].csf(CSF_destroy)) {
						if sys.chars[i][j].helperId == hid {
							sys.debugRef[0] = i
							sys.debugRef[1] = j
							found = true
							break
						}
					}
				}
				// gotta reset after one full iteration
				hnStart = 0
			}
		}

		// Come back around if we haven't found it and we're not starting from 0
		if (pnStart > 0 || hnStart > 0) && !found {
			for i := 0; i < len(sys.chars) && !found; i++ {
				for j := 1; j < len(sys.chars[i]); j++ {
					if sys.chars[i][j].helperId == hid {
						sys.debugRef[0] = i
						sys.debugRef[1] = j
						found = true
						break
					}
				}
			}
		}

		if !found {
			l.RaiseError("Could not find any helpers matching helper ID %d", hid)
		}

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
		if !nilArg(l, 3) {
			bank = int32(numArg(l, 3))
		}
		l.Push(lua.LNumber(fnt.TextWidth(strArg(l, 2), bank)))
		return 1
	})
	luaRegister(l, "fontNew", func(l *lua.LState) int {
		var height int32 = -1
		if !nilArg(l, 2) {
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
				sys.await(sys.cfg.Config.Framerate)
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
				nextId := sys.cfg.Config.HelperMax
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
				for i := MaxSimul * 2; i < MaxPlayerNo; i += 1 {
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

				// Match loop
				if sys.runMatch() {
					// Match is restarting
					for i, b := range sys.reloadCharSlot {
						if b {
							if !sys.cfg.Debug.KeepSpritesOnReload {
								if s := sys.cgi[i].sff; s != nil {
									removeSFFCache(s.filename)
								}
							}
							sys.chars[i] = []*Char{}
							b = false
						}
					}
					if sys.reloadStageFlg {
						sys.stage = nil
					}
					if sys.reloadLifebarFlg {
						if err := sys.lifebar.reloadLifebar(); err != nil {
							l.RaiseError(err.Error())
						}
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
			if sys.netConnection != nil {
				sys.netConnection.Stop()
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
				tbl.RawSetString("roundTime", lua.LNumber(sys.maxRoundTime))
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
					sys.bgm.Open("", 1, 100, 0, 0, 0, 1.0, 1)
					sys.playBgmFlg = false
				}
				sys.clearAllSound()
				sys.allPalFX = newPalFX()
				sys.bgPalFX = newPalFX()
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
				sys.paused = false
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
		if chk := FileExist(def); len(chk) != 0 {
			def = chk
		} else {
			if strings.ToLower(def[0:6]) != "chars/" && strings.ToLower(def[1:3]) != ":/" && (def[0] != '/' || idx > 0 && !strings.Contains(def[:idx], ":")) {
				def = "chars/" + def
			}
			if def = FileExist(def); len(def) == 0 {
				return 0
			}
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
		// palettes
		subt = l.NewTable()
		if len(c.pal) > 0 {
			for k, v := range c.pal {
				subt.RawSetInt(k+1, lua.LNumber(v))
			}
		} else {
			subt.RawSetInt(1, lua.LNumber(1))
		}
		tbl.RawSetString("pal", subt)
		// default palettes
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
		if n == 0 {
			subt.RawSetInt(1, lua.LNumber(1))
		}
		tbl.RawSetString("pal_defaults", subt)
		// palette keymap
		subt = l.NewTable()
		if len(c.pal_keymap) > 0 {
			for k, v := range c.pal_keymap {
				if int32(k+1) != v { // only actual remaps are relevant
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
		if !nilArg(l, 1) {
			pn = int(numArg(l, 1))
		}
		if pn != 0 && (pn < 1 || pn > MaxPlayerNo) {
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
		if !nilArg(l, 2) {
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
	luaRegister(l, "getClipboardString", func(*lua.LState) int {
		s := sys.window.Window.GetClipboardString()
		l.Push(lua.LString(s))
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
		tn := int(numArg(l, 1))
		if tn < 1 || tn > 2 {
			l.RaiseError("\nInvalid team side: %v\n", tn)
		}
		l.Push(lua.LNumber(sys.consecutiveWins[tn-1]))
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
	luaRegister(l, "getJoystickGUID", func(*lua.LState) int {
		l.Push(lua.LString(input.GetJoystickGUID(int(numArg(l, 1)))))
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
		if !nilArg(l, 1) {
			min = int(Clamp(int32(numArg(l, 1)), 0, int32(max-1)))
			max = min + 1
		}
		for joy = min; joy < max; joy++ {
			if input.IsJoystickPresent(joy) {
				axes := input.GetJoystickAxes(joy)
				btns := input.GetJoystickButtons(joy)

				s = CheckAxisForDpad(joy, &axes, len(btns))
				if s != "" {
					break
				}
				s = CheckAxisForTrigger(joy, &axes)
				if s != "" {
					break
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
		if nilArg(l, 1) {
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
				s = sys.window.GetClipboardString()
			} else {
				s = sys.keyString
			}
		}
		l.Push(lua.LString(s))
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
		l.Push(lua.LNumber(sys.maxRoundTime))
		return 1
	})
	luaRegister(l, "getStageInfo", func(*lua.LState) int {
		c := sys.sel.GetStage(int(numArg(l, 1)))
		tbl := l.NewTable()
		tbl.RawSetString("name", lua.LString(c.name))
		tbl.RawSetString("def", lua.LString(c.def))
		tbl.RawSetString("portrait_scale", lua.LNumber(c.portrait_scale))
		acTable := l.NewTable()
		for _, v := range c.attachedchardef {
			acTable.Append(lua.LString(v))
		}
		tbl.RawSetString("attachedchardef", acTable)
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
		// path, group, sound, loops before give up searching for group/sound pair (optional)
		var max uint32
		if !nilArg(l, 4) {
			max = uint32(numArg(l, 4))
		}
		w, err := loadFromSnd(strArg(l, 1), int32(numArg(l, 2)), int32(numArg(l, 3)), max)
		if err != nil {
			l.RaiseError(err.Error())
		}
		l.Push(newUserData(l, w))
		return 1
	})
	luaRegister(l, "loadState", func(*lua.LState) int {
		sys.loadStateFlag = true
		return 0
	})
	luaRegister(l, "loadDebugFont", func(l *lua.LState) int {
		ts := NewTextSprite()
		f, err := loadFnt(strArg(l, 1), -1)
		if err != nil {
			l.RaiseError("\nCan't load %v: %v\n", strArg(l, 1), err.Error())
		}
		ts.fnt = f
		if !nilArg(l, 2) {
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
	luaRegister(l, "loadGameOption", func(l *lua.LState) int {
		cfg := sys.cfg
		if !nilArg(l, 1) {
			cfg, err := loadConfig(strArg(l, 1))
			if err != nil {
				l.RaiseError("\nCan't load %v: %v\n", strArg(l, 1), err.Error())
			}
			sys.cfg = *cfg
		}
		lv := toLValue(l, cfg)
		l.Push(lv)
		return 1
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
	luaRegister(l, "loadPalettes", func(*lua.LState) int {
		//Actual process of loading palettes
		a, _ := toUserData(l, 1).(*Anim)
		if sys.usePalette == true {
			loadCharPalettes(a.anim.sff, a.anim.sff.filename, int(numArg(l, 2)))
		}
		l.Push(newUserData(l, a))
		return 1
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
	luaRegister(l, "loadText", func(l *lua.LState) int {
		path := strArg(l, 1)
		content, err := LoadText(path)
		if err != nil {
			l.Push(lua.LNil)
			return 1
		}
		l.Push(lua.LString(content))
		return 1
	})
	luaRegister(l, "modifyGameOption", func(l *lua.LState) int {
		query := strArg(l, 1)
		// Handle the second argument which can be nil, string, or a table
		val := l.Get(2)
		var value interface{}
		if val.Type() == lua.LTBool {
			// Convert Lua bools to native Go bools
			value = lua.LVAsBool(val)
		} else if val == lua.LNil {
			// nil value means remove a map entry or clear an array depending on context
			value = nil
		} else if tbl, ok := val.(*lua.LTable); ok {
			// If a table is provided, treat it as an array of strings
			var arr []string
			tbl.ForEach(func(k, v lua.LValue) {
				arr = append(arr, v.String())
			})
			value = arr
		} else {
			// Otherwise, treat it as a string
			value = val.String()
		}

		// Pass interface{} value
		err := sys.cfg.SetValueUpdate(query, value)
		if err == nil {
			return 0
		}
		l.RaiseError("\nmodifyGameOption: %v\n", err.Error())
		return 0
	})
	luaRegister(l, "mapSet", func(*lua.LState) int {
		//map_name, value, map_type
		var scType int32
		if !nilArg(l, 3) && strArg(l, 3) == "add" {
			scType = 1
		}
		sys.debugWC.mapSet(strArg(l, 1), float32(numArg(l, 2)), scType)
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
				switch strings.ToLower(string(k)) {
				case "life":
					sys.sel.ocd[tn-1][mn-1].life = int32(lua.LVAsNumber(value))
				case "lifemax":
					sys.sel.ocd[tn-1][mn-1].lifeMax = int32(lua.LVAsNumber(value))
				case "power":
					sys.sel.ocd[tn-1][mn-1].power = int32(lua.LVAsNumber(value))
				case "dizzypoints":
					sys.sel.ocd[tn-1][mn-1].dizzyPoints = int32(lua.LVAsNumber(value))
				case "guardpoints":
					sys.sel.ocd[tn-1][mn-1].guardPoints = int32(lua.LVAsNumber(value))
				case "ratiolevel":
					sys.sel.ocd[tn-1][mn-1].ratioLevel = int32(lua.LVAsNumber(value))
				case "liferatio":
					sys.sel.ocd[tn-1][mn-1].lifeRatio = float32(lua.LVAsNumber(value))
				case "attackratio":
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
		var loop, loopcount, volume, loopstart, loopend, startposition int = 1, -1, 100, 0, 0, 0
		var freqmul float32 = 1.0
		if !nilArg(l, 2) {
			loop = int(numArg(l, 2))
		}
		if !nilArg(l, 3) {
			volume = int(numArg(l, 3))
		}
		if !nilArg(l, 4) {
			loopstart = int(numArg(l, 4))
		}
		if !nilArg(l, 5) && numArg(l, 5) > 1 {
			loopend = int(numArg(l, 5))
		}
		if !nilArg(l, 6) && numArg(l, 6) > 1 {
			startposition = int(numArg(l, 6))
		}
		if !nilArg(l, 7) {
			freqmul = ClampF(float32(numArg(l, 7)), 0.01, 5.0)
		}
		if !nilArg(l, 8) {
			loopcount = int(numArg(l, 8))
		}
		sys.bgm.Open(strArg(l, 1), loop, volume, loopstart, loopend, startposition, freqmul, loopcount)
		return 0
	})
	luaRegister(l, "playSnd", func(l *lua.LState) int {
		f, lw, lp, stopgh, stopcs := false, false, false, false, false
		var g, n, ch, vo, priority, lc int32 = -1, 0, -1, 100, 0, 0
		var loopstart, loopend, startposition int = 0, 0, 0
		var p, fr float32 = 0, 1
		x := &sys.debugWC.pos[0]
		ls := sys.debugWC.localscl
		if !nilArg(l, 1) { // group_no
			g = int32(numArg(l, 1))
		}
		if !nilArg(l, 2) { // sound_no
			n = int32(numArg(l, 2))
		}
		if !nilArg(l, 3) { // volumescale
			vo = int32(numArg(l, 3))
		}
		if !nilArg(l, 4) { // commonSnd
			f = boolArg(l, 4)
		}
		if !nilArg(l, 5) { // channel
			ch = int32(numArg(l, 5))
		}
		if !nilArg(l, 6) { // lowpriority
			lw = boolArg(l, 6)
		}
		if !nilArg(l, 7) { // freqmul
			fr = float32(numArg(l, 7))
		}
		if !nilArg(l, 8) { // loop
			lp = boolArg(l, 8)
		}
		if !nilArg(l, 9) { // pan
			p = float32(numArg(l, 9))
		}
		if !nilArg(l, 10) { // priority
			priority = int32(numArg(l, 10))
		}
		if !nilArg(l, 11) { // loopstart
			loopstart = int(numArg(l, 11))
		}
		if !nilArg(l, 12) { // loopend
			loopend = int(numArg(l, 12))
		}
		if !nilArg(l, 13) { // startposition
			startposition = int(numArg(l, 13))
		}
		if !nilArg(l, 14) { // loopcount
			lc = int32(numArg(l, 14))
		}
		if !nilArg(l, 15) { // StopOnGetHit
			stopgh = boolArg(l, 15)
		}
		if !nilArg(l, 16) { // StopOnChangeState
			stopcs = boolArg(l, 16)
		}
		prefix := ""
		if f {
			prefix = "f"
		}
		// If the loopcount is 0, then read the loop parameter
		if lc == 0 {
			if lp {
				sys.debugWC.playSound(prefix, lw, -1, g, n, ch, vo, p, fr, ls, x, false, priority, loopstart, loopend, startposition, stopgh, stopcs)
			} else {
				sys.debugWC.playSound(prefix, lw, 0, g, n, ch, vo, p, fr, ls, x, false, priority, loopstart, loopend, startposition, stopgh, stopcs)
			}
			// Otherwise, read the loopcount parameter directly
		} else {
			sys.debugWC.playSound(prefix, lw, lc, g, n, ch, vo, p, fr, ls, x, false, priority, loopstart, loopend, startposition, stopgh, stopcs)
		}
		return 0
	})
	luaRegister(l, "playerBufReset", func(*lua.LState) int {
		if !nilArg(l, 1) {
			pn := int(numArg(l, 1))
			if pn < 1 || pn > len(sys.chars) || len(sys.chars[pn-1]) == 0 {
				return 0
			}
			for j := range sys.chars[pn-1][0].cmd {
				sys.chars[pn-1][0].cmd[j].BufReset()
				sys.chars[pn-1][0].setASF(ASF_nohardcodedkeys)
			}
		} else {
			for _, p := range sys.chars {
				if len(p) > 0 {
					for j := range p[0].cmd {
						p[0].cmd[j].BufReset()
						p[0].setASF(ASF_nohardcodedkeys)
					}
				}
			}
		}
		return 0
	})
	luaRegister(l, "preloadListChar", func(*lua.LState) int {
		if !nilArg(l, 2) {
			sys.sel.charSpritePreload[[...]int16{int16(numArg(l, 1)), int16(numArg(l, 2))}] = true
		} else {
			sys.sel.charAnimPreload = append(sys.sel.charAnimPreload, int32(numArg(l, 1)))
		}
		return 0
	})
	luaRegister(l, "preloadListStage", func(*lua.LState) int {
		if !nilArg(l, 2) {
			sys.sel.stageSpritePreload[[...]int16{int16(numArg(l, 1)), int16(numArg(l, 2))}] = true
		} else {
			sys.sel.stageAnimPreload = append(sys.sel.stageAnimPreload, int32(numArg(l, 1)))
		}
		return 0
	})
	luaRegister(l, "printConsole", func(l *lua.LState) int {
		if !nilArg(l, 2) && boolArg(l, 2) {
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
		if sys.netConnection != nil {
			sys.netConnection.recording, _ = os.Create(strArg(l, 1))
		}
		return 0
	})
	luaRegister(l, "replayStop", func(*lua.LState) int {
		if sys.cfg.Netplay.RollbackNetcode {
			if sys.rollback.session != nil && sys.rollback.session.recording != nil {
				sys.rollback.session.recording.Close()
				sys.rollback.session.recording = nil
			}
		} else {
			if sys.netConnection != nil && sys.netConnection.recording != nil {
				sys.netConnection.recording.Close()
				sys.netConnection.recording = nil
			}
		}
		return 0
	})
	luaRegister(l, "resetKey", func(*lua.LState) int {
		sys.keyInput = KeyUnknown
		sys.keyString = ""
		return 0
	})
	luaRegister(l, "resetAILevel", func(l *lua.LState) int {
		for i := range sys.aiLevel {
			sys.aiLevel[i] = 0
		}
		return 0
	})
	luaRegister(l, "resetMatchData", func(*lua.LState) int {
		sys.allPalFX = newPalFX()
		sys.bgPalFX = newPalFX()
		sys.resetGblEffect()
		for i, p := range sys.chars {
			if len(p) > 0 {
				sys.clearPlayerAssets(i, boolArg(l, 1))
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
	luaRegister(l, "saveState", func(*lua.LState) int {
		sys.saveStateFlag = true
		return 0
	})
	luaRegister(l, "saveGameOption", func(l *lua.LState) int {
		path := sys.cfg.Def
		if !nilArg(l, 1) {
			path = strArg(l, 1)
		}
		if err := sys.cfg.Save(path); err != nil {
			l.RaiseError("\nsaveGameOption: %v\n", err.Error())
		}
		return 0
	})
	luaRegister(l, "screenshot", func(*lua.LState) int {
		if !sys.isTakingScreenshot {
			sys.isTakingScreenshot = true
		}
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
		if !nilArg(l, 1) {
			sff, err := loadSff(strArg(l, 1), false)
			if err != nil {
				l.RaiseError("\nCan't load %v: %v\n", strArg(l, 1), err.Error())
			}
			sys.runMainThreadTask()
			l.Push(newUserData(l, sff))
		} else {
			l.Push(newUserData(l, newSff()))
		}
		return 1
	})
	luaRegister(l, "modelNew", func(l *lua.LState) int {
		if !nilArg(l, 1) {
			mdl, err := loadglTFModel(strArg(l, 1))
			if err != nil {
				l.RaiseError("\nCan't load %v: %v\n", strArg(l, 1), err.Error())
			}
			sys.mainThreadTask <- func() {
				gfx.SetModelVertexData(1, mdl.vertexBuffer)
				gfx.SetModelIndexData(1, mdl.elementBuffer...)
			}
			sys.runMainThreadTask()
			l.Push(newUserData(l, mdl))
		}
		return 1
	})
	luaRegister(l, "selfState", func(*lua.LState) int {
		sys.debugWC.selfState(int32(numArg(l, 1)), -1, -1, 1, "")
		return 0
	})
	luaRegister(l, "setAccel", func(*lua.LState) int {
		sys.debugAccel = float32(numArg(l, 1))
		return 0
	})
	luaRegister(l, "setAILevel", func(*lua.LState) int {
		level := float32(numArg(l, 1))
		sys.aiLevel[sys.debugWC.playerNo] = level
		for _, c := range sys.chars[sys.debugWC.playerNo] {
			if level == 0 {
				c.controller = sys.debugWC.playerNo
			} else {
				c.controller = ^sys.debugWC.playerNo
			}
		}
		return 0
	})
	luaRegister(l, "setAutoLevel", func(*lua.LState) int {
		sys.autolevel = boolArg(l, 1)
		return 0
	})
	luaRegister(l, "setCom", func(*lua.LState) int {
		pn := int(numArg(l, 1))
		ailv := float32(numArg(l, 2))
		if pn < 1 || pn > MaxPlayerNo {
			l.RaiseError("\nInvalid player number: %v\n", pn)
		}
		if ailv > 0 {
			sys.aiLevel[pn-1] = ailv
		} else {
			sys.aiLevel[pn-1] = 0
		}
		return 0
	})
	luaRegister(l, "setConsecutiveWins", func(l *lua.LState) int {
		tn := int(numArg(l, 1))
		if tn < 1 || tn > 2 {
			l.RaiseError("\nInvalid team side: %v\n", tn)
		}
		sys.consecutiveWins[tn-1] = int32(numArg(l, 2))
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
				switch strings.ToLower(string(k)) {
				case "active": // enabled by default
					sys.lifebar.active = lua.LVAsBool(value)
				case "bars": // enabled by default
					sys.lifebar.bars = lua.LVAsBool(value)
				case "guardbar": // enabled depending on config.ini
					sys.lifebar.guardbar = lua.LVAsBool(value)
				case "hidebars": // enabled depending on dialogue system.def settings
					sys.lifebar.hidebars = lua.LVAsBool(value)
				case "match":
					sys.lifebar.ma.active = lua.LVAsBool(value)
				case "mode": // enabled by default
					sys.lifebar.mode = lua.LVAsBool(value)
				case "p1ailevel":
					sys.lifebar.ai[0].active = lua.LVAsBool(value)
				case "p1score":
					sys.lifebar.sc[0].active = lua.LVAsBool(value)
				case "p1wincount":
					sys.lifebar.wc[0].active = lua.LVAsBool(value)
				case "p2ailevel":
					sys.lifebar.ai[1].active = lua.LVAsBool(value)
				case "p2score":
					sys.lifebar.sc[1].active = lua.LVAsBool(value)
				case "p2wincount":
					sys.lifebar.wc[1].active = lua.LVAsBool(value)
				case "redlifebar": // enabled depending on config.ini
					sys.lifebar.redlifebar = lua.LVAsBool(value)
				case "stunbar": // enabled depending on config.ini
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
	luaRegister(l, "setLifebarOffsetY", func(l *lua.LState) int {
		sys.lifebarOffsetY = float32(numArg(l, 1))
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
		if !nilArg(l, 2) {
			sys.scoreStart[1] = float32(numArg(l, 2))
		}
		return 0
	})
	luaRegister(l, "setLifebarTimer", func(*lua.LState) int {
		sys.timerStart = int32(numArg(l, 1))
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
	luaRegister(l, "setMotifDir", func(*lua.LState) int {
		sys.motifDir = strArg(l, 1)
		return 0
	})
	luaRegister(l, "setPlayers", func(l *lua.LState) int {
		total := int(numArg(l, 1))

		// Resize sys.keyConfig
		if len(sys.keyConfig) > total {
			sys.keyConfig = sys.keyConfig[:total]
		} else {
			for i := len(sys.keyConfig); i < total; i++ {
				sys.keyConfig = append(sys.keyConfig, KeyConfig{})
			}
		}

		// Cleanup sys.cfg.Keys
		for key := range sys.cfg.Keys {
			var num int
			if _, err := fmt.Sscanf(key, "keys_p%d", &num); err != nil || num > total {
				delete(sys.cfg.Keys, key)
				sys.cfg.IniFile.DeleteSection(fmt.Sprintf("Keys_P%d", num))
			}
		}

		// Resize sys.joystickConfig
		if len(sys.joystickConfig) > total {
			sys.joystickConfig = sys.joystickConfig[:total]
		} else {
			for i := len(sys.joystickConfig); i < total; i++ {
				sys.joystickConfig = append(sys.joystickConfig, KeyConfig{})
			}
		}

		// Cleanup sys.cfg.Joystick
		for key := range sys.cfg.Joystick {
			var num int
			if _, err := fmt.Sscanf(key, "joystick_p%d", &num); err != nil || num > total {
				delete(sys.cfg.Joystick, key)
				sys.cfg.IniFile.DeleteSection(fmt.Sprintf("Joystick_P%d", num))
			}
		}

		return 0
	})
	luaRegister(l, "setPower", func(*lua.LState) int {
		sys.debugWC.setPower(int32(numArg(l, 1)))
		return 0
	})
	luaRegister(l, "setRedLife", func(*lua.LState) int {
		sys.debugWC.redLifeSet(int32(numArg(l, 1)))
		return 0
	})
	luaRegister(l, "setGameSpeed", func(*lua.LState) int {
		sys.cfg.Options.GameSpeed = int(numArg(l, 1))
		return 0
	})
	luaRegister(l, "setRoundTime", func(l *lua.LState) int {
		sys.maxRoundTime = int32(numArg(l, 1))
		return 0
	})
	luaRegister(l, "setConsecutiveRounds", func(l *lua.LState) int {
		sys.consecutiveRounds = boolArg(l, 1)
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
		sys.curRoundTime = int32(numArg(l, 1))
		return 0
	})
	luaRegister(l, "setTimeFramesPerCount", func(l *lua.LState) int {
		sys.lifebar.ti.framespercount = int32(numArg(l, 1))
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
		if !nilArg(l, 4) {
			volumescale = int32(numArg(l, 4))
		}
		var pan float32
		if !nilArg(l, 5) {
			pan = float32(numArg(l, 5))
		}
		var loopstart, loopend, startposition int
		if !nilArg(l, 6) {
			loopstart = int(numArg(l, 6))
		}
		if !nilArg(l, 7) {
			loopend = int(numArg(l, 7))
		}
		if !nilArg(l, 8) {
			startposition = int(numArg(l, 8))
		}
		s.play([...]int32{int32(numArg(l, 2)), int32(numArg(l, 3))}, volumescale, pan, loopstart, loopend, startposition)
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
	luaRegister(l, "frameStep", func(*lua.LState) int {
		sys.frameStepFlag = true
		return 0
	})
	luaRegister(l, "stopAllSound", func(l *lua.LState) int {
		sys.stopAllCharSound()
		return 0
	})
	luaRegister(l, "stopSnd", func(l *lua.LState) int {
		sys.debugWC.soundChannels.SetSize(0)
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
	luaRegister(l, "textImgSetXShear", func(*lua.LState) int {
		ts, ok := toUserData(l, 1).(*TextSprite)
		if !ok {
			userDataError(l, 1, ts)
		}
		ts.xshear = float32(numArg(l, 2))
		return 0
	})
	luaRegister(l, "textImgSetAngle", func(*lua.LState) int {
		ts, ok := toUserData(l, 1).(*TextSprite)
		if !ok {
			userDataError(l, 1, ts)
		}
		ts.angle = float32(numArg(l, 2))
		return 0
	})
	luaRegister(l, "toggleClsnDisplay", func(*lua.LState) int {
		if !sys.debugModeAllowed() {
			return 0
		}
		if !nilArg(l, 1) {
			sys.clsnDisplay = boolArg(l, 1)
		} else {
			sys.clsnDisplay = !sys.clsnDisplay
		}
		return 0
	})
	luaRegister(l, "toggleContinueScreen", func(*lua.LState) int {
		if !nilArg(l, 1) {
			sys.continueScreenFlg = boolArg(l, 1)
		} else {
			sys.continueScreenFlg = !sys.continueScreenFlg
		}
		return 0
	})
	luaRegister(l, "toggleDebugDisplay", func(*lua.LState) int {
		if !sys.debugModeAllowed() {
			return 0
		}
		if !nilArg(l, 1) {
			sys.debugDisplay = !sys.debugDisplay
			return 0
		}
		if !sys.debugDisplay {
			sys.debugDisplay = true
		} else {
			idx := 0
			// Find index of current debug player
			for i := 0; i < len(sys.charList.runOrder); i++ {
				ro := sys.charList.runOrder[i]
				if ro.playerNo == sys.debugRef[0] && ro.helperIndex == int32(sys.debugRef[1]) {
					idx = i + 1 // Then check the next one
					break
				}
			}
			if idx == 0 || idx >= len(sys.charList.runOrder) {
				sys.debugRef[0] = 0
				sys.debugRef[1] = 0
				sys.debugDisplay = false
			} else {
				sys.debugRef[0] = sys.charList.runOrder[idx].playerNo
				sys.debugRef[1] = int(sys.charList.runOrder[idx].helperIndex)
			}
		}
		return 0
	})
	luaRegister(l, "toggleDialogueBars", func(*lua.LState) int {
		if !nilArg(l, 1) {
			sys.dialogueBarsFlg = boolArg(l, 1)
		} else {
			sys.dialogueBarsFlg = !sys.dialogueBarsFlg
		}
		return 0
	})
	luaRegister(l, "toggleFullscreen", func(*lua.LState) int {
		fs := !sys.window.fullscreen
		if !nilArg(l, 1) {
			fs = boolArg(l, 1)
		}
		if fs != sys.window.fullscreen {
			sys.window.toggleFullscreen()
		}
		return 0
	})
	luaRegister(l, "toggleLifebarDisplay", func(*lua.LState) int {
		if !nilArg(l, 1) {
			sys.lifebarHide = boolArg(l, 1)
		} else {
			sys.lifebarHide = !sys.lifebarHide
		}
		return 0
	})
	luaRegister(l, "toggleMaxPowerMode", func(*lua.LState) int {
		if !nilArg(l, 1) {
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
		if !nilArg(l, 1) {
			sys.noSoundFlg = boolArg(l, 1)
		} else {
			sys.noSoundFlg = !sys.noSoundFlg
		}
		return 0
	})
	luaRegister(l, "togglePause", func(*lua.LState) int {
		if !nilArg(l, 1) {
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
		if !nilArg(l, 1) {
			sys.postMatchFlg = boolArg(l, 1)
		} else {
			sys.postMatchFlg = !sys.postMatchFlg
		}
		return 0
	})
	luaRegister(l, "toggleVictoryScreen", func(*lua.LState) int {
		if !nilArg(l, 1) {
			sys.victoryScreenFlg = boolArg(l, 1)
		} else {
			sys.victoryScreenFlg = !sys.victoryScreenFlg
		}
		return 0
	})
	luaRegister(l, "toggleVSync", func(*lua.LState) int {
		if !nilArg(l, 1) {
			sys.cfg.Video.VSync = int(numArg(l, 1))
		} else if sys.cfg.Video.VSync == 0 {
			sys.cfg.Video.VSync = 1
		} else {
			sys.cfg.Video.VSync = 0
		}
		sys.window.SetSwapInterval(sys.cfg.Video.VSync)
		return 0
	})
	luaRegister(l, "toggleWinScreen", func(*lua.LState) int {
		if !nilArg(l, 1) {
			sys.winScreenFlg = boolArg(l, 1)
		} else {
			sys.winScreenFlg = !sys.winScreenFlg
		}
		return 0
	})
	luaRegister(l, "toggleWireframeDisplay", func(*lua.LState) int {
		if !sys.debugModeAllowed() {
			return 0
		}
		if !nilArg(l, 1) {
			sys.wireframeDisplay = boolArg(l, 1)
		} else {
			sys.wireframeDisplay = !sys.wireframeDisplay
		}
		return 0
	})
	luaRegister(l, "updateVolume", func(l *lua.LState) int {
		sys.bgm.UpdateVolume()
		return 0
	})
	luaRegister(l, "usePalette", func(l *lua.LState) int {
		//Sets palettes to be able to be used, this is used to avoid spending any time with the functions to make it work.
		sys.usePalette = boolArg(l, 1)
		return 1
	})
	luaRegister(l, "wavePlay", func(l *lua.LState) int {
		s, ok := toUserData(l, 1).(*Sound)
		if !ok {
			userDataError(l, 1, s)
		}
		var g, n int32
		if !nilArg(l, 2) {
			g = int32(numArg(l, 2))
		}
		if !nilArg(l, 3) {
			n = int32(numArg(l, 3))
		}
		sys.soundChannels.Play(s, g, n, 100, 0.0, 0, 0, 0)
		return 0
	})
}

// Trigger Functions
func triggerFunctions(l *lua.LState) {
	// Create a temporary dummy character to avoid possible nil checks
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
		if c := sys.debugWC.parent(true); c != nil {
			sys.debugWC, ret = c, true
		}
		l.Push(lua.LBool(ret))
		return 1
	})
	luaRegister(l, "root", func(*lua.LState) int {
		ret := false
		if c := sys.debugWC.root(true); c != nil {
			sys.debugWC, ret = c, true
		}
		l.Push(lua.LBool(ret))
		return 1
	})
	luaRegister(l, "helper", func(l *lua.LState) int {
		ret := false
		id, index := int32(-1), 0
		// Check if ID is provided
		if !nilArg(l, 1) {
			id = int32(numArg(l, 1))
		}
		// Check if index is provided
		if !nilArg(l, 2) {
			index = int(numArg(l, 2))
		}
		if c := sys.debugWC.helperTrigger(id, index); c != nil {
			sys.debugWC, ret = c, true
		}
		l.Push(lua.LBool(ret))
		return 1
	})
	luaRegister(l, "target", func(l *lua.LState) int {
		ret := false
		id, index := int32(-1), 0
		// Check if ID is provided
		if !nilArg(l, 1) {
			id = int32(numArg(l, 1))
		}
		// Check if index is provided
		if !nilArg(l, 2) {
			index = int(numArg(l, 2))
		}
		if c := sys.debugWC.targetTrigger(id, index); c != nil {
			sys.debugWC, ret = c, true
		}
		l.Push(lua.LBool(ret))
		return 1
	})
	luaRegister(l, "partner", func(*lua.LState) int {
		ret, n := false, int32(0)
		if !nilArg(l, 1) {
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
		if !nilArg(l, 1) {
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
		if !nilArg(l, 1) {
			n = int32(numArg(l, 1))
		}
		if c := sys.debugWC.enemyNearTrigger(n); c != nil {
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
	luaRegister(l, "playerindex", func(*lua.LState) int {
		ret := false
		if c := sys.playerIndexRedirect(int32(numArg(l, 1))); c != nil {
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
	luaRegister(l, "stateowner", func(*lua.LState) int {
		ret := false
		if c := sys.chars[sys.debugWC.ss.sb.playerNo][0]; c != nil {
			sys.debugWC, ret = c, true
		}
		l.Push(lua.LBool(ret))
		return 1
	})
	luaRegister(l, "helperindex", func(*lua.LState) int {
		ret, id := false, int32(0)
		if !nilArg(l, 1) {
			id = int32(numArg(l, 1))
		}
		if c := sys.debugWC.helperIndexTrigger(id, true); c != nil {
			sys.debugWC, ret = c, true
		}
		l.Push(lua.LBool(ret))
		return 1
	})
	// vanilla triggers
	luaRegister(l, "ailevel", func(*lua.LState) int {
		if !sys.debugWC.asf(ASF_noailevel) {
			l.Push(lua.LNumber(sys.debugWC.getAILevel()))
		} else {
			l.Push(lua.LNumber(0))
		}
		return 1
	})
	luaRegister(l, "airjumpcount", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.airJumpCount))
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
	// animelem (deprecated by animelemtime)
	luaRegister(l, "animelemno", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.animElemNo(int32(numArg(l, 1)) - 1).ToI())) // Offset by 1 because animations step before scripts run
		return 1
	})
	luaRegister(l, "animelemtime", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.animElemTime(int32(numArg(l, 1))).ToI()) - 1) // Offset by 1 because animations step before scripts run
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
	luaRegister(l, "attack", func(*lua.LState) int {
		base := float32(sys.debugWC.gi().attackBase) * sys.debugWC.ocd().attackRatio / 100
		l.Push(lua.LNumber(base * sys.debugWC.attackMul[0] * 100))
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
	luaRegister(l, "bgmvar", func(*lua.LState) int {
		arg := strings.ToLower(strArg(l, 1))
		// If the streamer is nil, return nil for strings
		if arg == "filename" {
			if sys.bgm.streamer == nil {
				l.Push(lua.LNil)
			} else {
				l.Push(lua.LString(sys.bgm.filename))
			}
			// Return a number
		} else {
			ln := lua.LNumber(0)
			if sys.bgm.streamer != nil {
				switch arg {
				case "freqmul":
					ln = lua.LNumber(sys.bgm.freqmul)
				case "length":
					ln = lua.LNumber(int32(sys.bgm.streamer.Len()))
				case "loop":
					ln = lua.LNumber(int32(sys.bgm.loop))
				case "loopcount":
					if sl, ok := sys.bgm.volctrl.Streamer.(*StreamLooper); ok {
						ln = lua.LNumber(sl.loopcount)
					}
				case "loopend":
					if sl, ok := sys.bgm.volctrl.Streamer.(*StreamLooper); ok {
						ln = lua.LNumber(sl.loopend)
					}
				case "loopstart":
					if sl, ok := sys.bgm.volctrl.Streamer.(*StreamLooper); ok {
						ln = lua.LNumber(sl.loopstart)
					}
				case "position":
					ln = lua.LNumber(int32(sys.bgm.streamer.Position()))
				case "startposition":
					ln = lua.LNumber(int32(sys.bgm.startPos))
				case "volume":
					ln = lua.LNumber(int32(sys.bgm.bgmVolume))
				}
			}
			l.Push(ln)
		}
		return 1
	})
	luaRegister(l, "bottomedge", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.bottomEdge()))
		return 1
	})
	luaRegister(l, "botboundbodydist", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.botBoundBodyDist()))
		return 1
	})
	luaRegister(l, "botbounddist", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.botBoundDist()))
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
	luaRegister(l, "clsnoverlap", func(l *lua.LState) int {
		id := int32(numArg(l, 2))
		var c1, c2 int32
		// Get box 1 type
		switch strings.ToLower(strArg(l, 1)) {
		case "clsn1":
			c1 = 1
		case "clsn2":
			c1 = 2
		case "size":
			c1 = 3
		default:
			l.RaiseError("Invalid collision box type")
		}
		// Get box 2 type
		switch strArg(l, 3) {
		case "clsn1":
			c2 = 1
		case "clsn2":
			c2 = 2
		case "size":
			c2 = 3
		default:
			l.RaiseError("Invalid collision box type")
		}
		l.Push(lua.LBool(sys.debugWC.clsnOverlapTrigger(c1, id, c2)))
		return 1
	})
	luaRegister(l, "clsnvar", func(l *lua.LState) int {
		c := strArg(l, 1)
		idx := int(numArg(l, 2))
		t := strArg(l, 3)
		v := lua.LNumber(math.NaN())

		getClsnCoord := func(offset int) {
			switch c {
			case "size":
				v = lua.LNumber(sys.debugWC.sizeBox[offset])
			case "clsn1":
				clsn := sys.debugWC.curFrame.Clsn1
				if clsn != nil && idx >= 0 && idx < len(clsn) {
					v = lua.LNumber(clsn[idx][offset])
				}
			case "clsn2":
				clsn := sys.debugWC.curFrame.Clsn2
				if clsn != nil && idx >= 0 && idx < len(clsn) {
					v = lua.LNumber(clsn[idx][offset])
				}
			}
		}

		switch t {
		case "back":
			getClsnCoord(0)
		case "top":
			getClsnCoord(1)
		case "front":
			getClsnCoord(2)
		case "bottom":
			getClsnCoord(3)
		}

		l.Push(v)
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
		case "data.dizzypoints":
			ln = lua.LNumber(c.gi().data.dizzypoints)
		case "data.guardpoints":
			ln = lua.LNumber(c.gi().data.guardpoints)
		case "data.attack":
			ln = lua.LNumber(c.gi().data.attack)
		case "data.defence":
			ln = lua.LNumber(c.gi().data.defence)
		case "data.fall.defence_up":
			ln = lua.LNumber(c.gi().data.fall.defence_up)
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
		case "data.hitsound.channel":
			ln = lua.LNumber(c.gi().data.hitsound_channel)
		case "data.guardsound.channel":
			ln = lua.LNumber(c.gi().data.guardsound_channel)
		case "data.ko.echo":
			ln = lua.LNumber(c.gi().data.ko.echo)
		case "data.volume":
			ln = lua.LNumber(c.gi().data.volume)
		case "data.intpersistindex":
			ln = lua.LNumber(c.gi().data.intpersistindex)
		case "data.floatpersistindex":
			ln = lua.LNumber(c.gi().data.floatpersistindex)
		case "size.xscale":
			ln = lua.LNumber(c.size.xscale)
		case "size.yscale":
			ln = lua.LNumber(c.size.yscale)
		case "size.ground.back":
			ln = lua.LNumber(-c.size.standbox[0])
		case "size.ground.front":
			ln = lua.LNumber(c.size.standbox[2])
		case "size.air.back":
			ln = lua.LNumber(-c.size.airbox[0])
		case "size.air.front":
			ln = lua.LNumber(c.size.airbox[2])
		case "size.height":
			ln = lua.LNumber(-c.size.standbox[1])
		case "size.attack.dist", "size.attack.dist.width.front": // Optional new syntax for consistency
			ln = lua.LNumber(c.size.attack.dist.width[0])
		case "size.attack.dist.width.back":
			ln = lua.LNumber(c.size.attack.dist.width[1])
		case "size.attack.dist.height.top":
			ln = lua.LNumber(c.size.attack.dist.height[0])
		case "size.attack.dist.height.bottom":
			ln = lua.LNumber(c.size.attack.dist.height[1])
		case "size.attack.dist.depth.top":
			ln = lua.LNumber(c.size.attack.dist.depth[0])
		case "size.attack.dist.depth.bottom":
			ln = lua.LNumber(c.size.attack.dist.depth[1])
		case "size.attack.depth.top":
			ln = lua.LNumber(c.size.attack.depth[0])
		case "size.attack.depth.bottom":
			ln = lua.LNumber(c.size.attack.depth[1])
		case "size.proj.attack.dist", "size.proj.attack.dist.width.front": // Optional new syntax for consistency
			ln = lua.LNumber(c.size.proj.attack.dist.width[0])
		case "size.proj.attack.dist.width.back":
			ln = lua.LNumber(c.size.proj.attack.dist.width[1])
		case "size.proj.attack.dist.height.top":
			ln = lua.LNumber(c.size.proj.attack.dist.height[0])
		case "size.proj.attack.dist.height.bottom":
			ln = lua.LNumber(c.size.proj.attack.dist.height[1])
		case "size.proj.attack.dist.depth.top":
			ln = lua.LNumber(c.size.proj.attack.dist.depth[0])
		case "size.proj.attack.dist.depth.bottom":
			ln = lua.LNumber(c.size.proj.attack.dist.depth[1])
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
		case "size.depth.top":
			ln = lua.LNumber(c.size.depth[0])
		case "size.depth.bottom":
			ln = lua.LNumber(c.size.depth[1])
		case "size.weight":
			ln = lua.LNumber(c.size.weight)
		case "size.pushfactor":
			ln = lua.LNumber(c.size.pushfactor)
		case "velocity.air.gethit.airrecover.add.x":
			ln = lua.LNumber(c.gi().velocity.air.gethit.airrecover.add[0])
		case "velocity.air.gethit.airrecover.add.y":
			ln = lua.LNumber(c.gi().velocity.air.gethit.airrecover.add[1])
		case "velocity.air.gethit.airrecover.back":
			ln = lua.LNumber(c.gi().velocity.air.gethit.airrecover.back)
		case "velocity.air.gethit.airrecover.down":
			ln = lua.LNumber(c.gi().velocity.air.gethit.airrecover.down)
		case "velocity.air.gethit.airrecover.fwd":
			ln = lua.LNumber(c.gi().velocity.air.gethit.airrecover.fwd)
		case "velocity.air.gethit.airrecover.mul.x":
			ln = lua.LNumber(c.gi().velocity.air.gethit.airrecover.mul[0])
		case "velocity.air.gethit.airrecover.mul.y":
			ln = lua.LNumber(c.gi().velocity.air.gethit.airrecover.mul[1])
		case "velocity.air.gethit.airrecover.up":
			ln = lua.LNumber(c.gi().velocity.air.gethit.airrecover.up)
		case "velocity.air.gethit.groundrecover.x":
			ln = lua.LNumber(c.gi().velocity.air.gethit.groundrecover[0])
		case "velocity.air.gethit.groundrecover.y":
			ln = lua.LNumber(c.gi().velocity.air.gethit.groundrecover[1])
		case "velocity.air.gethit.ko.add.x":
			ln = lua.LNumber(c.gi().velocity.air.gethit.ko.add[0])
		case "velocity.air.gethit.ko.add.y":
			ln = lua.LNumber(c.gi().velocity.air.gethit.ko.add[1])
		case "velocity.air.gethit.ko.ymin":
			ln = lua.LNumber(c.gi().velocity.air.gethit.ko.ymin)
		case "velocity.airjump.back.x":
			ln = lua.LNumber(c.gi().velocity.airjump.back)
		case "velocity.airjump.down.x":
			ln = lua.LNumber(c.gi().velocity.airjump.down[0])
		case "velocity.airjump.fwd.x":
			ln = lua.LNumber(c.gi().velocity.airjump.fwd)
		case "velocity.airjump.neu.x":
			ln = lua.LNumber(c.gi().velocity.airjump.neu[0])
		case "velocity.airjump.up.x":
			ln = lua.LNumber(c.gi().velocity.airjump.up[0])
		case "velocity.airjump.up.y":
			ln = lua.LNumber(c.gi().velocity.airjump.up[1])
		case "velocity.airjump.up.z":
			ln = lua.LNumber(c.gi().velocity.airjump.up[2])
		case "velocity.airjump.y":
			ln = lua.LNumber(c.gi().velocity.airjump.neu[1])
		case "velocity.ground.gethit.ko.add.x":
			ln = lua.LNumber(c.gi().velocity.ground.gethit.ko.add[0])
		case "velocity.ground.gethit.ko.add.y":
			ln = lua.LNumber(c.gi().velocity.ground.gethit.ko.add[1])
		case "velocity.ground.gethit.ko.xmul":
			ln = lua.LNumber(c.gi().velocity.ground.gethit.ko.xmul)
		case "velocity.ground.gethit.ko.ymin":
			ln = lua.LNumber(c.gi().velocity.ground.gethit.ko.ymin)
		case "velocity.jump.back.x":
			ln = lua.LNumber(c.gi().velocity.jump.back)
		case "velocity.jump.down.x":
			ln = lua.LNumber(c.gi().velocity.jump.down[0])
		case "velocity.jump.fwd.x":
			ln = lua.LNumber(c.gi().velocity.jump.fwd)
		case "velocity.jump.neu.x":
			ln = lua.LNumber(c.gi().velocity.jump.neu[0])
		case "velocity.jump.up.x":
			ln = lua.LNumber(c.gi().velocity.jump.up[0])
		case "velocity.jump.up.y":
			ln = lua.LNumber(c.gi().velocity.jump.up[1])
		case "velocity.jump.up.z":
			ln = lua.LNumber(c.gi().velocity.jump.up[2])
		case "velocity.jump.y":
			ln = lua.LNumber(c.gi().velocity.jump.neu[1])
		case "velocity.run.back.x":
			ln = lua.LNumber(c.gi().velocity.run.back[0])
		case "velocity.run.back.y":
			ln = lua.LNumber(c.gi().velocity.run.back[1])
		case "velocity.run.down.x":
			ln = lua.LNumber(c.gi().velocity.run.down[0])
		case "velocity.run.down.y":
			ln = lua.LNumber(c.gi().velocity.run.down[1])
		case "velocity.run.fwd.x":
			ln = lua.LNumber(c.gi().velocity.run.fwd[0])
		case "velocity.run.fwd.y":
			ln = lua.LNumber(c.gi().velocity.run.fwd[1])
		case "velocity.run.up.x":
			ln = lua.LNumber(c.gi().velocity.run.up[0])
		case "velocity.run.up.y":
			ln = lua.LNumber(c.gi().velocity.run.up[1])
		case "velocity.run.up.z":
			ln = lua.LNumber(c.gi().velocity.run.up[2])
		case "velocity.runjump.back.x":
			ln = lua.LNumber(c.gi().velocity.runjump.back[0])
		case "velocity.runjump.back.y":
			ln = lua.LNumber(c.gi().velocity.runjump.back[1])
		case "velocity.runjump.down.x":
			ln = lua.LNumber(c.gi().velocity.runjump.down[0])
		case "velocity.runjump.fwd.x":
			ln = lua.LNumber(c.gi().velocity.runjump.fwd[0])
		case "velocity.runjump.up.x":
			ln = lua.LNumber(c.gi().velocity.runjump.up[0])
		case "velocity.runjump.up.y":
			ln = lua.LNumber(c.gi().velocity.runjump.up[1])
		case "velocity.runjump.up.z":
			ln = lua.LNumber(c.gi().velocity.runjump.up[2])
		case "velocity.runjump.y":
			ln = lua.LNumber(c.gi().velocity.runjump.fwd[1])
		case "velocity.walk.back.x":
			ln = lua.LNumber(c.gi().velocity.walk.back)
		case "velocity.walk.down.x":
			ln = lua.LNumber(c.gi().velocity.walk.down[0])
		case "velocity.walk.fwd.x":
			ln = lua.LNumber(c.gi().velocity.walk.fwd)
		case "velocity.walk.up.x":
			ln = lua.LNumber(c.gi().velocity.walk.up[0])
		case "velocity.walk.up.y":
			ln = lua.LNumber(c.gi().velocity.walk.up[1])
		case "velocity.walk.up.z":
			ln = lua.LNumber(c.gi().velocity.walk.up[2])
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
		case "movement.down.gethit.offset.x":
			ln = lua.LNumber(c.gi().movement.down.gethit.offset[0])
		case "movement.down.gethit.offset.y":
			ln = lua.LNumber(c.gi().movement.down.gethit.offset[1])
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
	luaRegister(l, "const1080p", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.constp(1920, float32(numArg(l, 1))).ToF()))
		return 1
	})
	luaRegister(l, "ctrl", func(*lua.LState) int {
		l.Push(lua.LBool(sys.debugWC.ctrl()))
		return 1
	})
	luaRegister(l, "defence", func(*lua.LState) int {
		l.Push(lua.LNumber(float32(sys.debugWC.finalDefense * 100)))
		return 1
	})
	luaRegister(l, "drawgame", func(*lua.LState) int {
		l.Push(lua.LBool(sys.debugWC.drawgame()))
		return 1
	})
	luaRegister(l, "envshakevar", func(*lua.LState) int {
		var ln lua.LNumber
		switch strings.ToLower(strArg(l, 1)) {
		case "time":
			ln = lua.LNumber(sys.envShake.time)
		case "freq":
			ln = lua.LNumber(sys.envShake.freq / float32(math.Pi) * 180)
		case "ampl":
			ln = lua.LNumber(sys.envShake.ampl)
		case "dir":
			ln = lua.LNumber(sys.envShake.dir / float32(math.Pi) * 180)
		default:
			l.RaiseError("\nInvalid argument: %v\n", strArg(l, 1))
		}
		l.Push(ln)
		return 1
	})
	luaRegister(l, "explodvar", func(*lua.LState) int {
		ln := lua.LNumber(math.NaN())
		id := int32(numArg(l, 1))
		idx := int(numArg(l, 2))
		vname := strArg(l, 3)

		for i, e := range sys.debugWC.getExplods(id) {
			if i == idx {
				switch strings.ToLower(vname) {
				case "accel x":
					ln = lua.LNumber(e.accel[0])
				case "accel y":
					ln = lua.LNumber(e.accel[1])
				case "accel z":
					ln = lua.LNumber(e.accel[2])
				case "anim":
					ln = lua.LNumber(e.animNo)
				case "angle":
					ln = lua.LNumber(e.anglerot[0] + e.interpolate_angle[0])
				case "angle x":
					ln = lua.LNumber(e.anglerot[1] + e.interpolate_angle[1])
				case "angle y":
					ln = lua.LNumber(e.anglerot[2] + e.interpolate_angle[2])
				case "animelem":
					ln = lua.LNumber(e.anim.curelem + 1)
				case "animelemtime":
					ln = lua.LNumber(e.anim.curelemtime)
				case "animplayerno":
					ln = lua.LNumber(e.animPN + 1)
				case "spriteplayerno":
					ln = lua.LNumber(e.spritePN + 1)
				case "bindtime":
					ln = lua.LNumber(e.bindtime)
				case "drawpal group":
					ln = lua.LNumber(sys.debugWC.explodDrawPal(e)[0])
				case "drawpal index":
					ln = lua.LNumber(sys.debugWC.explodDrawPal(e)[1])
				case "facing":
					ln = lua.LNumber(e.facing)
				case "friction x":
					ln = lua.LNumber(e.friction[0])
				case "friction y":
					ln = lua.LNumber(e.friction[1])
				case "friction z":
					ln = lua.LNumber(e.friction[2])
				case "id":
					ln = lua.LNumber(e.id)
				case "layerno":
					ln = lua.LNumber(e.layerno)
				case "pausemovetime":
					ln = lua.LNumber(e.pausemovetime)
				case "pos x":
					ln = lua.LNumber(e.pos[0] + e.offset[0] + e.relativePos[0] + e.interpolate_pos[0])
				case "pos y":
					ln = lua.LNumber(e.pos[1] + e.offset[1] + e.relativePos[1] + e.interpolate_pos[1])
				case "pos z":
					ln = lua.LNumber(e.pos[2] + e.offset[2] + e.relativePos[2] + e.interpolate_pos[2])
				case "removetime":
					ln = lua.LNumber(e.removetime)
				case "scale x":
					ln = lua.LNumber(e.scale[0] * e.interpolate_scale[0])
				case "scale y":
					ln = lua.LNumber(e.scale[1] * e.interpolate_scale[1])
				case "sprpriority":
					ln = lua.LNumber(e.sprpriority)
				case "time":
					ln = lua.LNumber(e.time)
				case "vel x":
					ln = lua.LNumber(e.velocity[0])
				case "vel y":
					ln = lua.LNumber(e.velocity[1])
				case "vel z":
					ln = lua.LNumber(e.velocity[2])
				case "xshear":
					ln = lua.LNumber(e.xshear)
				default:
					l.RaiseError("\nInvalid argument: %v\n", vname)
				}
				break
			}
		}
		l.Push(ln)
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
	luaRegister(l, "gameOption", func(l *lua.LState) int {
		value, err := sys.cfg.GetValue(strArg(l, 1))
		if err == nil {
			lv := toLValue(l, value)
			l.Push(lv)
			return 1
		}
		l.RaiseError("\nInvalid argument: %v\n", strArg(l, 1))
		return 0
	})
	luaRegister(l, "gametime", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.gameTime + sys.preFightTime))
		return 1
	})
	luaRegister(l, "gamewidth", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.gameWidth()))
		return 1
	})
	luaRegister(l, "hitdefvar", func(*lua.LState) int {
		c := sys.debugWC
		switch strings.ToLower(strArg(l, 1)) {
		case "guardflag":
			l.Push(flagLStr(c.hitdef.guardflag))
		case "hitflag":
			l.Push(flagLStr(c.hitdef.hitflag))
		case "hitdamage":
			l.Push(lua.LNumber(c.hitdef.hitdamage))
		case "guarddamage":
			l.Push(lua.LNumber(c.hitdef.guarddamage))
		case "p1stateno":
			l.Push(lua.LNumber(c.hitdef.p1stateno))
		case "p2stateno":
			l.Push(lua.LNumber(c.hitdef.p2stateno))
		case "priority":
			l.Push(lua.LNumber(c.hitdef.priority))
		case "id":
			l.Push(lua.LNumber(c.hitdef.id))
		case "sparkno":
			l.Push(lua.LNumber(c.hitdef.sparkno))
		case "guard.sparkno":
			l.Push(lua.LNumber(c.hitdef.guard_sparkno))
		case "sparkx":
			l.Push(lua.LNumber(c.hitdef.sparkxy[0]))
		case "sparky":
			l.Push(lua.LNumber(c.hitdef.sparkxy[1]))
		case "pausetime":
			l.Push(lua.LNumber(c.hitdef.pausetime[0]))
		case "guard.pausetime":
			l.Push(lua.LNumber(c.hitdef.guard_pausetime[0]))
		case "shaketime":
			l.Push(lua.LNumber(c.hitdef.pausetime[1]))
		case "guard.shaketime":
			l.Push(lua.LNumber(c.hitdef.guard_pausetime[1]))
		case "hitsound.group":
			l.Push(lua.LNumber(c.hitdef.hitsound[0]))
		case "hitsound.number":
			l.Push(lua.LNumber(c.hitdef.hitsound[1]))
		case "guardsound.group":
			l.Push(lua.LNumber(c.hitdef.guardsound[0]))
		case "guardsound.number":
			l.Push(lua.LNumber(c.hitdef.guardsound[1]))
		default:
			l.RaiseError("\nInvalid argument: %v\n", strArg(l, 1))
		}
		return 1
	})
	luaRegister(l, "gethitvar", func(*lua.LState) int {
		c := sys.debugWC
		var ln lua.LNumber
		switch strings.ToLower(strArg(l, 1)) {
		case "fall.envshake.dir":
			ln = lua.LNumber(c.ghv.fall_envshake_dir)
		case "animtype":
			ln = lua.LNumber(c.gethitAnimtype())
		case "air.animtype":
			ln = lua.LNumber(c.ghv.airanimtype)
		case "ground.animtype":
			ln = lua.LNumber(c.ghv.groundanimtype)
		case "fall.animtype":
			ln = lua.LNumber(c.ghv.fall_animtype)
		case "type":
			ln = lua.LNumber(c.ghv._type)
		case "airtype":
			ln = lua.LNumber(c.ghv.airtype)
		case "groundtype":
			ln = lua.LNumber(c.ghv.groundtype)
		case "damage":
			ln = lua.LNumber(c.ghv.damage)
		case "guardcount":
			ln = lua.LNumber(c.ghv.guardcount)
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
		case "recovertime", "down.recovertime": // Added second term for consistency
			ln = lua.LNumber(c.ghv.down_recovertime)
		case "xoff":
			ln = lua.LNumber(c.ghv.xoff)
		case "yoff":
			ln = lua.LNumber(c.ghv.yoff)
		case "zoff":
			ln = lua.LNumber(c.ghv.zoff)
		case "xvel":
			ln = lua.LNumber(c.ghv.xvel)
		case "yvel":
			ln = lua.LNumber(c.ghv.yvel)
		case "zvel":
			ln = lua.LNumber(c.ghv.zvel)
		case "xaccel":
			ln = lua.LNumber(c.ghv.xaccel)
		case "yaccel":
			ln = lua.LNumber(c.ghv.yaccel)
		case "zaccel":
			ln = lua.LNumber(c.ghv.zaccel)
		case "xveladd":
			ln = lua.LNumber(c.ghv.xveladd)
		case "yveladd":
			ln = lua.LNumber(c.ghv.yveladd)
		case "hitid", "chainid":
			ln = lua.LNumber(c.ghv.chainId())
		case "guarded":
			ln = lua.LNumber(Btoi(c.ghv.guarded))
		case "isbound":
			ln = lua.LNumber(Btoi(c.isTargetBound()))
		case "fall":
			ln = lua.LNumber(Btoi(c.ghv.fallflag))
		case "fall.damage":
			ln = lua.LNumber(c.ghv.fall_damage)
		case "fall.xvel":
			if math.IsNaN(float64(c.ghv.fall_xvelocity)) {
				ln = lua.LNumber(-32760) // Winmugen behavior
			} else {
				ln = lua.LNumber(c.ghv.fall_xvelocity)
			}
		case "fall.yvel":
			ln = lua.LNumber(c.ghv.fall_yvelocity)
		case "fall.zvel":
			if math.IsNaN(float64(c.ghv.fall_zvelocity)) {
				ln = lua.LNumber(-32760) // Winmugen behavior
			} else {
				ln = lua.LNumber(c.ghv.fall_zvelocity)
			}
		case "fall.recover":
			ln = lua.LNumber(Btoi(c.ghv.fall_recover))
		case "fall.time":
			ln = lua.LNumber(c.fallTime)
		case "fall.recovertime":
			ln = lua.LNumber(c.ghv.fall_recovertime)
		case "fall.kill":
			ln = lua.LNumber(Btoi(c.ghv.fall_kill))
		case "fall.envshake.time":
			ln = lua.LNumber(c.ghv.fall_envshake_time)
		case "fall.envshake.freq":
			ln = lua.LNumber(c.ghv.fall_envshake_freq)
		case "fall.envshake.ampl":
			ln = lua.LNumber(c.ghv.fall_envshake_ampl)
		case "fall.envshake.phase":
			ln = lua.LNumber(c.ghv.fall_envshake_phase)
		case "fall.envshake.mul":
			ln = lua.LNumber(c.ghv.fall_envshake_mul)
		case "attr":
			// return here, because ln is a
			// LNumber (we have a LString)
			l.Push(attrLStr(c.ghv.attr))
			return 1
		case "dizzypoints":
			ln = lua.LNumber(c.ghv.dizzypoints)
		case "guardpoints":
			ln = lua.LNumber(c.ghv.guardpoints)
		case "id":
			ln = lua.LNumber(c.ghv.playerId)
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
		case "power":
			ln = lua.LNumber(c.ghv.power)
		case "hitpower":
			ln = lua.LNumber(c.ghv.hitpower)
		case "guardpower":
			ln = lua.LNumber(c.ghv.guardpower)
		case "kill":
			ln = lua.LNumber(Btoi(c.ghv.kill))
		case "priority":
			ln = lua.LNumber(c.ghv.priority)
		case "facing":
			ln = lua.LNumber(c.ghv.facing)
		case "ground.velocity.x":
			ln = lua.LNumber(c.ghv.ground_velocity[0])
		case "ground.velocity.y":
			ln = lua.LNumber(c.ghv.ground_velocity[1])
		case "ground.velocity.z":
			ln = lua.LNumber(c.ghv.ground_velocity[2])
		case "air.velocity.x":
			ln = lua.LNumber(c.ghv.air_velocity[0])
		case "air.velocity.y":
			ln = lua.LNumber(c.ghv.air_velocity[1])
		case "air.velocity.z":
			ln = lua.LNumber(c.ghv.air_velocity[2])
		case "down.velocity.x":
			ln = lua.LNumber(c.ghv.down_velocity[0])
		case "down.velocity.y":
			ln = lua.LNumber(c.ghv.down_velocity[1])
		case "down.velocity.z":
			ln = lua.LNumber(c.ghv.down_velocity[2])
		case "guard.velocity.x":
			ln = lua.LNumber(c.ghv.guard_velocity[0])
		case "guard.velocity.y":
			ln = lua.LNumber(c.ghv.guard_velocity[1])
		case "guard.velocity.z":
			ln = lua.LNumber(c.ghv.guard_velocity[2])
		case "airguard.velocity.x":
			ln = lua.LNumber(c.ghv.airguard_velocity[0])
		case "airguard.velocity.y":
			ln = lua.LNumber(c.ghv.airguard_velocity[1])
		case "airguard.velocity.z":
			ln = lua.LNumber(c.ghv.airguard_velocity[2])
		case "frame":
			ln = lua.LNumber(Btoi(c.ghv.frame))
		case "down.recover":
			ln = lua.LNumber(Btoi(c.ghv.down_recover))
		case "guardflag":
			// return here, because ln is a
			// LNumber (we have a LString)
			l.Push(flagLStr(c.ghv.guardflag))
			return 1
		default:
			l.RaiseError("\nInvalid argument: %v\n", strArg(l, 1))
		}
		l.Push(ln)
		return 1
	})
	luaRegister(l, "groundlevel", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.groundLevel))
		return 1
	})
	luaRegister(l, "guardcount", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.guardCount))
		return 1
	})
	luaRegister(l, "helperid", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.helperId))
		return 1
	})
	luaRegister(l, "hitbyattr", func(*lua.LState) int {
		flg := int32(0)
		input := strings.ToLower(strArg(l, 1))
		// Split input at the commas
		attr := strings.Split(input, ",")
		if len(attr) < 2 {
			l.RaiseError("Attribute must contain at least two flags separated by ','")
			return 0
		}
		// Get SCA attribute
		attrsca := strings.TrimSpace(attr[0])
		for _, letter := range attrsca {
			switch letter {
			case 's':
				flg |= int32(ST_S)
			case 'c':
				flg |= int32(ST_C)
			case 'a':
				flg |= int32(ST_A)
			default:
				l.RaiseError("Invalid attribute: %c", letter)
				return 0
			}
		}
		// Get attack type attributes
		for i := 1; i < len(attr); i++ {
			attrtype := strings.TrimSpace(attr[i])
			switch attrtype {
			case "na":
				flg |= int32(AT_NA)
			case "nt":
				flg |= int32(AT_NT)
			case "np":
				flg |= int32(AT_NP)
			case "sa":
				flg |= int32(AT_SA)
			case "st":
				flg |= int32(AT_ST)
			case "sp":
				flg |= int32(AT_SP)
			case "ha":
				flg |= int32(AT_HA)
			case "ht":
				flg |= int32(AT_HT)
			case "hp":
				flg |= int32(AT_HP)
			case "aa":
				flg |= int32(AT_AA)
			case "at":
				flg |= int32(AT_AT)
			case "ap":
				flg |= int32(AT_AP)
			case "n":
				flg |= int32(AT_NA | AT_NT | AT_NP)
			case "s":
				flg |= int32(AT_SA | AT_ST | AT_SP)
			case "h", "a":
				flg |= int32(AT_HA | AT_HT | AT_HP)
			default:
				l.RaiseError("Invalid attribute: %s", attrtype)
				return 0
			}
		}
		l.Push(lua.LBool(sys.debugWC.hitByAttrTrigger(flg)))
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
		l.Push(lua.LBool(sys.debugWC.ghv.fallflag))
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
		l.Push(lua.LNumber(sys.debugWC.ghv.xvel * sys.debugWC.facing))
		return 1
	})
	luaRegister(l, "hitvelY", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.ghv.yvel))
		return 1
	})
	luaRegister(l, "hitvelZ", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.ghv.zvel))
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
	luaRegister(l, "isclsnproxy", func(*lua.LState) int {
		l.Push(lua.LBool(sys.debugWC.isclsnproxy))
		return 1
	})
	luaRegister(l, "ishelper", func(l *lua.LState) int {
		id, index := int32(-1), -1
		// Check if ID is provided
		if !nilArg(l, 1) {
			id = int32(numArg(l, 1))
		}
		// Check if index is provided
		if !nilArg(l, 2) {
			index = int(numArg(l, 2))
		}
		l.Push(lua.LBool(sys.debugWC.isHelper(id, index)))
		return 1
	})
	luaRegister(l, "ishometeam", func(*lua.LState) int {
		l.Push(lua.LBool(sys.debugWC.teamside == sys.home))
		return 1
	})
	luaRegister(l, "index", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.indexTrigger()))
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
	luaRegister(l, "movehitvar", func(*lua.LState) int {
		c := sys.debugWC
		var ln lua.LNumber
		switch strings.ToLower(strArg(l, 1)) {
		case "cornerpush":
			ln = lua.LNumber(c.mhv.cornerpush)
		case "frame":
			ln = lua.LNumber(Btoi(c.mhv.frame))
		case "id":
			ln = lua.LNumber(c.mhv.playerId)
		case "overridden":
			ln = lua.LNumber(Btoi(c.mhv.overridden))
		case "playerno":
			ln = lua.LNumber(c.mhv.playerNo)
		case "sparkx":
			ln = lua.LNumber(c.mhv.sparkxy[0])
		case "sparky":
			ln = lua.LNumber(c.mhv.sparkxy[1])
		case "uniqhit":
			ln = lua.LNumber(len(c.hitdefTargets))
		default:
			l.RaiseError("\nInvalid argument: %v\n", strArg(l, 1))
		}
		l.Push(ln)
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
	// name also returns p1name-p8name variants and helpername
	luaRegister(l, "name", func(*lua.LState) int {
		n := int32(1)
		if !nilArg(l, 1) {
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
			if p := sys.charList.enemyNear(sys.debugWC, n/2-1, true, false); p != nil {
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
		if !nilArg(l, 1) {
			id = int32(numArg(l, 1))
		}
		l.Push(lua.LNumber(sys.debugWC.numExplod(BytecodeInt(id)).ToI()))
		return 1
	})
	luaRegister(l, "numhelper", func(*lua.LState) int {
		id := int32(0)
		if !nilArg(l, 1) {
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
	luaRegister(l, "numstagebg", func(*lua.LState) int {
		id := int32(-1)
		if !nilArg(l, 1) {
			id = int32(numArg(l, 1))
		}
		l.Push(lua.LNumber(sys.debugWC.numStageBG(BytecodeInt(id)).ToI()))
		return 1
	})
	luaRegister(l, "numtarget", func(*lua.LState) int {
		id := int32(-1)
		if !nilArg(l, 1) {
			id = int32(numArg(l, 1))
		}
		l.Push(lua.LNumber(sys.debugWC.numTarget(BytecodeInt(id)).ToI()))
		return 1
	})
	luaRegister(l, "numtext", func(*lua.LState) int {
		id := int32(-1)
		if !nilArg(l, 1) {
			id = int32(numArg(l, 1))
		}
		l.Push(lua.LNumber(sys.debugWC.numText(BytecodeInt(id)).ToI()))
		return 1
	})
	luaRegister(l, "palfxvar", func(*lua.LState) int {
		var ln lua.LNumber
		switch strings.ToLower(strArg(l, 1)) {
		case "time":
			ln = lua.LNumber(sys.debugWC.palfxvar(0))
		case "add.r":
			ln = lua.LNumber(sys.debugWC.palfxvar(1))
		case "add.g":
			ln = lua.LNumber(sys.debugWC.palfxvar(2))
		case "add.b":
			ln = lua.LNumber(sys.debugWC.palfxvar(3))
		case "mul.r":
			ln = lua.LNumber(sys.debugWC.palfxvar(4))
		case "mul.g":
			ln = lua.LNumber(sys.debugWC.palfxvar(5))
		case "mul.b":
			ln = lua.LNumber(sys.debugWC.palfxvar(6))
		case "color":
			ln = lua.LNumber(sys.debugWC.palfxvar2(1))
		case "hue":
			ln = lua.LNumber(sys.debugWC.palfxvar2(2))
		case "invertall":
			ln = lua.LNumber(sys.debugWC.palfxvar(-1))
		case "invertblend":
			ln = lua.LNumber(sys.debugWC.palfxvar(-2))
		case "bg.time":
			ln = lua.LNumber(sys.palfxvar(0, 1))
		case "bg.add.r":
			ln = lua.LNumber(sys.palfxvar(1, 1))
		case "bg.add.g":
			ln = lua.LNumber(sys.palfxvar(2, 1))
		case "bg.add.b":
			ln = lua.LNumber(sys.palfxvar(3, 1))
		case "bg.mul.r":
			ln = lua.LNumber(sys.palfxvar(4, 1))
		case "bg.mul.g":
			ln = lua.LNumber(sys.palfxvar(5, 1))
		case "bg.mul.b":
			ln = lua.LNumber(sys.palfxvar(6, 1))
		case "bg.color":
			ln = lua.LNumber(sys.palfxvar2(1, 1))
		case "bg.hue":
			ln = lua.LNumber(sys.palfxvar2(2, 1))
		case "bg.invertall":
			ln = lua.LNumber(sys.palfxvar(-1, 1))
		case "all.time":
			ln = lua.LNumber(sys.palfxvar(0, 2))
		case "all.add.r":
			ln = lua.LNumber(sys.palfxvar(1, 2))
		case "all.add.g":
			ln = lua.LNumber(sys.palfxvar(2, 2))
		case "all.add.b":
			ln = lua.LNumber(sys.palfxvar(3, 2))
		case "all.mul.r":
			ln = lua.LNumber(sys.palfxvar(4, 2))
		case "all.mul.g":
			ln = lua.LNumber(sys.palfxvar(5, 2))
		case "all.mul.b":
			ln = lua.LNumber(sys.palfxvar(6, 2))
		case "all.color":
			ln = lua.LNumber(sys.palfxvar2(1, 2))
		case "all.hue":
			ln = lua.LNumber(sys.palfxvar2(2, 2))
		case "all.invertall":
			ln = lua.LNumber(sys.palfxvar(-1, 2))
		case "all.invertblend":
			ln = lua.LNumber(sys.palfxvar(-2, 2))
		default:
			l.RaiseError("\nInvalid argument: %v\n", strArg(l, 1))
		}
		l.Push(ln)
		return 1
	})
	// p1name and other variants can be checked via name
	luaRegister(l, "p2bodydistX", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.p2BodyDistX(sys.debugWC).ToI()))
		return 1
	})
	luaRegister(l, "p2bodydistY", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.p2BodyDistY(sys.debugWC).ToI()))
		return 1
	})
	luaRegister(l, "p2bodydistZ", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.p2BodyDistZ(sys.debugWC).ToI()))
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
	luaRegister(l, "p2distZ", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.rdDistZ(sys.debugWC.p2(), sys.debugWC).ToI()))
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
	luaRegister(l, "drawpal", func(*lua.LState) int {
		var ln lua.LNumber
		switch strings.ToLower(strArg(l, 1)) {
		case "group":
			ln = lua.LNumber(sys.debugWC.drawPal()[0])
		case "index":
			ln = lua.LNumber(sys.debugWC.drawPal()[1])
		default:
			l.RaiseError("\nInvalid argument: %v\n", strArg(l, 1))
		}
		l.Push(ln)
		return 1
	})
	luaRegister(l, "parentdistX", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.rdDistX(sys.debugWC.parent(true), sys.debugWC).ToI()))
		return 1
	})
	luaRegister(l, "parentdistY", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.rdDistY(sys.debugWC.parent(true), sys.debugWC).ToI()))
		return 1
	})
	luaRegister(l, "parentdistZ", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.rdDistZ(sys.debugWC.parent(true), sys.debugWC).ToI()))
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
	luaRegister(l, "prevanim", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.prevAnimNo))
		return 1
	})
	luaRegister(l, "prevmovetype", func(*lua.LState) int {
		var s string
		switch sys.debugWC.ss.prevMoveType {
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
	luaRegister(l, "prevstateno", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.ss.prevno))
		return 1
	})
	luaRegister(l, "prevstatetype", func(*lua.LState) int {
		var s string
		switch sys.debugWC.ss.prevStateType {
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
	luaRegister(l, "projcanceltime", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.projCancelTime(
			BytecodeInt(int32(numArg(l, 1)))).ToI()))
		return 1
	})
	// projcontact (deprecated by projcontacttime)
	luaRegister(l, "projcontacttime", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.projContactTime(
			BytecodeInt(int32(numArg(l, 1)))).ToI()))
		return 1
	})
	luaRegister(l, "projclsnoverlap", func(l *lua.LState) int {
		idx := int32(numArg(l, 1))
		pid := int32(numArg(l, 2))
		cboxStr := strings.ToLower(strArg(l, 3))

		var cbox int32
		switch cboxStr {
		case "clsn1":
			cbox = 1
		case "clsn2":
			cbox = 2
		case "size":
			cbox = 3
		default:
			l.RaiseError("Invalid collision box type: " + cboxStr)
			l.Push(lua.LBool(false))
			return 1
		}
		l.Push(lua.LBool(sys.debugWC.projClsnOverlapTrigger(idx, pid, cbox)))
		return 1
	})
	// projguarded (deprecated by projguardedtime)
	luaRegister(l, "projguardedtime", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.projGuardedTime(
			BytecodeInt(int32(numArg(l, 1)))).ToI()))
		return 1
	})
	// projhit (deprecated by projhittime)
	luaRegister(l, "projhittime", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.projHitTime(
			BytecodeInt(int32(numArg(l, 1)))).ToI()))
		return 1
	})
	luaRegister(l, "projvar", func(*lua.LState) int {
		var lv lua.LValue
		id := int32(numArg(l, 1))
		idx := int(numArg(l, 2))
		vname := strArg(l, 3)

		for i, p := range sys.debugWC.getProjs(id) {
			if i == idx {
				switch strings.ToLower(vname) {
				case "accel x":
					lv = lua.LNumber(p.accel[0])
				case "accel y":
					lv = lua.LNumber(p.accel[1])
				case "accel z":
					lv = lua.LNumber(p.accel[2])
				case "anim":
					lv = lua.LNumber(p.anim)
				case "animelem":
					lv = lua.LNumber(p.ani.curelem + 1)
				case "angle":
					lv = lua.LNumber(p.anglerot[0])
				case "angle x":
					lv = lua.LNumber(p.anglerot[1])
				case "angle y":
					lv = lua.LNumber(p.anglerot[2])
				case "drawpal group":
					lv = lua.LNumber(sys.debugWC.projDrawPal(p)[0])
				case "drawpal index":
					lv = lua.LNumber(sys.debugWC.projDrawPal(p)[1])
				case "facing":
					lv = lua.LNumber(p.facing)
				case "highbound":
					lv = lua.LNumber(p.heightbound[1])
				case "lowbound":
					lv = lua.LNumber(p.heightbound[0])
				case "pausemovetime":
					lv = lua.LNumber(p.pausemovetime)
				case "pos x":
					lv = lua.LNumber(p.pos[0])
				case "pos y":
					lv = lua.LNumber(p.pos[1])
				case "pos z":
					lv = lua.LNumber(p.pos[2])
				case "projcancelanim":
					lv = lua.LNumber(p.cancelanim)
				case "projedgebound":
					lv = lua.LNumber(p.edgebound)
				case "projhitanim":
					lv = lua.LNumber(p.hitanim)
				case "projhits":
					lv = lua.LNumber(p.hits)
				case "projhitsmax":
					lv = lua.LNumber(p.totalhits)
				case "projid":
					lv = lua.LNumber(p.id)
				case "projlayerno":
					lv = lua.LNumber(p.layerno)
				case "projmisstime":
					lv = lua.LNumber(p.curmisstime)
				case "projpriority":
					lv = lua.LNumber(p.priority)
				case "projremove":
					lv = lua.LBool(p.remove)
				case "projremanim":
					lv = lua.LNumber(p.remanim)
				case "projremovetime":
					lv = lua.LNumber(p.removetime)
				case "projsprpriority":
					lv = lua.LNumber(p.sprpriority)
				case "projstagebound":
					lv = lua.LNumber(p.stagebound)
				case "remvelocity x":
					lv = lua.LNumber(p.remvelocity[0])
				case "remvelocity y":
					lv = lua.LNumber(p.remvelocity[1])
				case "remvelocity z":
					lv = lua.LNumber(p.remvelocity[2])
				case "scale x":
					lv = lua.LNumber(p.scale[0])
				case "scale y":
					lv = lua.LNumber(p.scale[1])
				case "shadow b":
					lv = lua.LNumber(p.shadow[0])
				case "shadow g":
					lv = lua.LNumber(p.shadow[0])
				case "shadow r":
					lv = lua.LNumber(p.shadow[0])
				case "supermovetime":
					lv = lua.LNumber(p.supermovetime)
				case "teamside":
					lv = lua.LNumber(p.hitdef.teamside)
				case "time":
					lv = lua.LNumber(p.time)
				case "vel x":
					lv = lua.LNumber(p.velocity[0])
				case "vel y":
					lv = lua.LNumber(p.velocity[1])
				case "vel z":
					lv = lua.LNumber(p.velocity[2])
				case "velmul x":
					lv = lua.LNumber(p.velmul[0])
				case "velmul y":
					lv = lua.LNumber(p.velmul[1])
				case "velmul z":
					lv = lua.LNumber(p.velmul[2])
				case "xshear":
					lv = lua.LNumber(p.xshear)
				//case "guardflag":
				//	lv = lua.LBool(p.hitdef.guardflag&fl != 0)
				//case "hitflag":
				//	lv = lua.LNumber(p.hitdef.hitflag&fl != 0)
				default:
					l.RaiseError("\nInvalid argument: %v\n", vname)
				}
				break
			}
		}
		l.Push(lv)
		return 1
	})
	// random (dedicated functionality already exists in Lua)
	//luaRegister(l, "random", func(*lua.LState) int {
	//	l.Push(lua.LNumber(Rand(0, 999)))
	//	return 1
	//})
	luaRegister(l, "runorder", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.runorder))
		return 1
	})
	luaRegister(l, "reversaldefattr", func(*lua.LState) int {
		attr, str := sys.debugWC.hitdef.reversal_attr, ""
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
	luaRegister(l, "rightedge", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.rightEdge()))
		return 1
	})
	luaRegister(l, "rootdistX", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.rdDistX(sys.debugWC.root(true), sys.debugWC).ToI()))
		return 1
	})
	luaRegister(l, "rootdistY", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.rdDistY(sys.debugWC.root(true), sys.debugWC).ToI()))
		return 1
	})
	luaRegister(l, "rootdistZ", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.rdDistZ(sys.debugWC.root(true), sys.debugWC).ToI()))
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
		l.Push(lua.LNumber(sys.roundState()))
		return 1
	})
	luaRegister(l, "roundswon", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.roundsWon()))
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
	luaRegister(l, "soundvar", func(*lua.LState) int {
		var lv lua.LValue
		id := int32(numArg(l, 1))
		vname := strArg(l, 2)
		var ch *SoundChannel

		if id < 0 {
			for _, ch := range sys.debugWC.soundChannels.channels {
				if ch.sfx != nil {
					if ch.IsPlaying() {
						break
					}
				}
			}
		} else {
			ch = sys.debugWC.soundChannels.Get(id)
		}

		if ch != nil && ch.sfx != nil {
			switch strings.ToLower(vname) {
			case "group":
				lv = lua.LNumber(ch.group)
			case "number":
				lv = lua.LNumber(ch.number)
			case "freqmul":
				lv = lua.LNumber(ch.sfx.freqmul)
			case "isplaying":
				lv = lua.LBool(ch.IsPlaying())
			case "length":
				if ch.streamer != nil {
					lv = lua.LNumber(ch.streamer.Len())
				} else {
					lv = lua.LNumber(0)
				}
			case "loopcount":
				if sl, ok := ch.sfx.streamer.(*StreamLooper); ok {
					lv = lua.LNumber(sl.loopcount)
				} else {
					lv = lua.LNumber(0)
				}
			case "loopend":
				if sl, ok := ch.sfx.streamer.(*StreamLooper); ok {
					lv = lua.LNumber(sl.loopend)
				} else {
					lv = lua.LNumber(0)
				}
			case "loopstart":
				if sl, ok := ch.sfx.streamer.(*StreamLooper); ok {
					lv = lua.LNumber(sl.loopstart)
				} else {
					lv = lua.LNumber(0)
				}
			case "pan":
				lv = lua.LNumber(ch.sfx.p)
			case "position":
				if sl, ok := ch.sfx.streamer.(*StreamLooper); ok {
					lv = lua.LNumber(sl.Position())
				} else {
					lv = lua.LNumber(0)
				}
			case "priority":
				lv = lua.LNumber(ch.sfx.priority)
			case "startposition":
				lv = lua.LNumber(ch.sfx.startPos)
			case "volumescale":
				lv = lua.LNumber(ch.sfx.volume / 256.0 * 100.0)
			default:
				l.RaiseError("\nInvalid argument: %v\n", vname)
				break
			}
		}
		l.Push(lv)
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
	luaRegister(l, "stagebgvar", func(l *lua.LState) int {
		id := int32(numArg(l, 1))
		idx := int(numArg(l, 2))
		vname := strArg(l, 3)
		var ln lua.LNumber
		// Get stage background element
		bg := sys.debugWC.getStageBg(id, idx, false)
		// Handle returns
		if bg != nil {
			switch strings.ToLower(vname) {
			case "anim":
				ln = lua.LNumber(bg.actionno)
			case "delta.x":
				ln = lua.LNumber(bg.delta[0])
			case "delta.y":
				ln = lua.LNumber(bg.delta[1])
			case "id":
				ln = lua.LNumber(bg.id)
			case "layerno":
				ln = lua.LNumber(bg.layerno)
			case "pos.x":
				ln = lua.LNumber(bg.bga.pos[0])
			case "pos.y":
				ln = lua.LNumber(bg.bga.pos[1])
			case "start.x":
				ln = lua.LNumber(bg.start[0])
			case "start.y":
				ln = lua.LNumber(bg.start[1])
			case "tile.x":
				ln = lua.LNumber(bg.anim.tile.xflag)
			case "tile.y":
				ln = lua.LNumber(bg.anim.tile.yflag)
			case "velocity.x":
				ln = lua.LNumber(bg.bga.vel[0])
			case "velocity.y":
				ln = lua.LNumber(bg.bga.vel[1])
			default:
				l.RaiseError("\nInvalid argument: %v\n", vname)
			}
		}
		l.Push(ln)
		return 1
	})
	luaRegister(l, "stagevar", func(*lua.LState) int {
		switch strings.ToLower(strArg(l, 1)) {
		case "info.author":
			l.Push(lua.LString(sys.stage.author))
		case "info.displayname":
			l.Push(lua.LString(sys.stage.displayname))
		case "info.ikemenversion":
			l.Push(lua.LNumber(sys.stage.ikemenverF))
		case "info.mugenversion":
			l.Push(lua.LNumber(sys.stage.mugenverF))
		case "info.name":
			l.Push(lua.LString(sys.stage.name))
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
		case "camera.tensionvel":
			l.Push(lua.LNumber(sys.stage.stageCamera.tensionvel))
		case "camera.cuthigh":
			l.Push(lua.LNumber(sys.stage.stageCamera.cuthigh))
		case "camera.cutlow":
			l.Push(lua.LNumber(sys.stage.stageCamera.cutlow))
		case "camera.startzoom":
			l.Push(lua.LNumber(sys.stage.stageCamera.startzoom))
		case "camera.zoomout":
			l.Push(lua.LNumber(sys.stage.stageCamera.zoomout))
		case "camera.zoomin":
			l.Push(lua.LNumber(sys.stage.stageCamera.zoomin))
		case "camera.zoomindelay":
			l.Push(lua.LNumber(sys.stage.stageCamera.zoomindelay))
		case "camera.zoominspeed":
			l.Push(lua.LNumber(sys.stage.stageCamera.zoominspeed))
		case "camera.zoomoutspeed":
			l.Push(lua.LNumber(sys.stage.stageCamera.zoomoutspeed))
		case "camera.yscrollspeed":
			l.Push(lua.LNumber(sys.stage.stageCamera.yscrollspeed))
		case "camera.ytension.enable":
			l.Push(lua.LBool(sys.stage.stageCamera.ytensionenable))
		case "camera.autocenter":
			l.Push(lua.LBool(sys.stage.stageCamera.autocenter))
		case "camera.lowestcap":
			l.Push(lua.LBool(sys.stage.stageCamera.lowestcap))
		case "playerinfo.leftbound":
			l.Push(lua.LNumber(sys.stage.leftbound))
		case "playerinfo.rightbound":
			l.Push(lua.LNumber(sys.stage.rightbound))
		case "playerinfo.topbound":
			l.Push(lua.LNumber(sys.stage.topbound))
		case "playerinfo.botbound":
			l.Push(lua.LNumber(sys.stage.botbound))
		case "playerinfo.p1startx":
			l.Push(lua.LNumber(sys.stage.p[0].startx))
		case "playerinfo.p1starty":
			l.Push(lua.LNumber(sys.stage.p[0].starty))
		case "playerinfo.p2startx":
			l.Push(lua.LNumber(sys.stage.p[1].startx))
		case "playerinfo.p2starty":
			l.Push(lua.LNumber(sys.stage.p[1].starty))
		case "playerinfo.p1startz":
			l.Push(lua.LNumber(sys.stage.p[0].startz))
		case "playerinfo.p2startz":
			l.Push(lua.LNumber(sys.stage.p[1].startz))
		case "playerinfo.p1facing":
			l.Push(lua.LNumber(sys.stage.p[0].facing))
		case "playerinfo.p2facing":
			l.Push(lua.LNumber(sys.stage.p[1].facing))
		case "scaling.topz":
			l.Push(lua.LNumber(sys.stage.stageCamera.topz))
		case "scaling.botz":
			l.Push(lua.LNumber(sys.stage.stageCamera.botz))
		case "scaling.topscale":
			l.Push(lua.LNumber(sys.stage.stageCamera.ztopscale))
		case "scaling.botscale":
			l.Push(lua.LNumber(sys.stage.stageCamera.ztopscale))
		case "bound.screenleft":
			l.Push(lua.LNumber(sys.stage.screenleft))
		case "bound.screenright":
			l.Push(lua.LNumber(sys.stage.screenright))
		case "stageinfo.localcoord.x":
			l.Push(lua.LNumber(sys.stage.stageCamera.localcoord[0]))
		case "stageinfo.localcoord.y":
			l.Push(lua.LNumber(sys.stage.stageCamera.localcoord[1]))
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
		case "shadow.offset.x":
			l.Push(lua.LNumber(sys.stage.sdw.offset[0]))
		case "shadow.offset.y":
			l.Push(lua.LNumber(sys.stage.sdw.offset[1]))
		case "reflection.intensity":
			l.Push(lua.LNumber(sys.stage.reflection.intensity))
		case "reflection.yscale":
			l.Push(lua.LNumber(sys.stage.reflection.yscale))
		case "reflection.offset.x":
			l.Push(lua.LNumber(sys.stage.reflection.offset[0]))
		case "reflection.offset.y":
			l.Push(lua.LNumber(sys.stage.reflection.offset[1]))
		case "reflection.xshear":
			l.Push(lua.LNumber(sys.stage.reflection.xshear))
		case "reflection.color.r":
			l.Push(lua.LNumber(int32((sys.stage.reflection.color & 0xFF0000) >> 16)))
		case "reflection.color.g":
			l.Push(lua.LNumber(int32((sys.stage.reflection.color & 0xFF00) >> 8)))
		case "reflection.color.b":
			l.Push(lua.LNumber(int32(sys.stage.reflection.color & 0xFF)))
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
		l.Push(lua.LNumber(sys.gameLogicSpeed()))
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
	luaRegister(l, "topboundbodydist", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.topBoundBodyDist()))
		return 1
	})
	luaRegister(l, "topbounddist", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.topBoundDist()))
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
		l.Push(lua.LBool(sys.debugWC.winType(WT_Special)))
		return 1
	})
	luaRegister(l, "winhyper", func(*lua.LState) int {
		l.Push(lua.LBool(sys.debugWC.winType(WT_Hyper)))
		return 1
	})

	// new triggers
	// atan2 (dedicated functionality already exists in Lua)
	luaRegister(l, "angle", func(*lua.LState) int {
		if sys.debugWC.csf(CSF_angledraw) {
			l.Push(lua.LNumber(sys.debugWC.anglerot[0]))
		} else {
			l.Push(lua.LNumber(0))
		}
		return 1
	})
	luaRegister(l, "xangle", func(*lua.LState) int {
		if sys.debugWC.csf(CSF_angledraw) {
			l.Push(lua.LNumber(sys.debugWC.anglerot[1]))
		} else {
			l.Push(lua.LNumber(0))
		}
		return 1
	})
	luaRegister(l, "yangle", func(*lua.LState) int {
		if sys.debugWC.csf(CSF_angledraw) {
			l.Push(lua.LNumber(sys.debugWC.anglerot[2]))
		} else {
			l.Push(lua.LNumber(0))
		}
		return 1
	})
	luaRegister(l, "alpha", func(*lua.LState) int {
		switch strings.ToLower(strArg(l, 1)) {
		case "source":
			l.Push(lua.LNumber(sys.debugWC.alpha[0]))
		case "dest":
			l.Push(lua.LNumber(sys.debugWC.alpha[1]))
		default:
			l.RaiseError("\nInvalid argument: %v\n", strArg(l, 1))
		}
		return 1
	})
	luaRegister(l, "xshear", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.xshear))
		return 1
	})
	luaRegister(l, "animelemvar", func(l *lua.LState) int {
		vname := strings.ToLower(strArg(l, 1))
		var ln lua.LNumber
		// Because the char's animation steps at the end of each frame, before the scripts run,
		// AnimElemVar Lua version uses curFrame instead of anim.CurrentFrame()
		f := sys.debugWC.curFrame
		if f != nil {
			switch vname {
			case "alphadest":
				ln = lua.LNumber(f.DstAlpha)
			case "alphasource":
				ln = lua.LNumber(f.SrcAlpha)
			case "angle":
				ln = lua.LNumber(f.Angle)
			case "group":
				ln = lua.LNumber(f.Group)
			case "hflip":
				ln = lua.LNumber(Btoi(f.Hscale < 0))
			case "image":
				ln = lua.LNumber(f.Number)
			case "numclsn1":
				ln = lua.LNumber(len(f.Clsn1))
			case "numclsn2":
				ln = lua.LNumber(len(f.Clsn2))
			case "time":
				ln = lua.LNumber(f.Time)
			case "vflip":
				ln = lua.LNumber(Btoi(f.Vscale < 0))
			case "xoffset":
				ln = lua.LNumber(f.Xoffset)
			case "xscale":
				ln = lua.LNumber(f.Xscale)
			case "yoffset":
				ln = lua.LNumber(f.Yoffset)
			case "yscale":
				ln = lua.LNumber(f.Yscale)
			default:
				l.RaiseError("\nInvalid argument: %v\n", vname)
			}
		}
		l.Push(ln)
		return 1
	})
	luaRegister(l, "animlength", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.anim.totaltime))
		return 1
	})
	luaRegister(l, "animplayerno", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.animPN + 1))
		return 1
	})
	luaRegister(l, "attack", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.attackMul[0] * 100))
		return 1
	})
	luaRegister(l, "clamp", func(*lua.LState) int {
		v1, v2, v3, retv := float32(numArg(l, 1)), float32(numArg(l, 2)), float32(numArg(l, 3)), float32(0)
		retv = MaxF(v2, MinF(v1, v3))
		l.Push(lua.LNumber(retv))
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
	luaRegister(l, "debugmode", func(*lua.LState) int {
		switch strings.ToLower(strArg(l, 1)) {
		case "accel":
			l.Push(lua.LNumber(sys.debugAccel))
		case "clsndisplay":
			l.Push(lua.LBool(sys.clsnDisplay))
		case "debugdisplay":
			l.Push(lua.LBool(sys.debugDisplay))
		case "lifebarhide":
			l.Push(lua.LBool(sys.lifebarHide))
		case "roundreset":
			l.Push(lua.LBool(sys.roundResetFlg))
		case "wireframedisplay":
			l.Push(lua.LBool(sys.wireframeDisplay))
		default:
			l.RaiseError("\nInvalid argument: %v\n", strArg(l, 1))
		}
		return 1
	})
	luaRegister(l, "decisiveround", func(*lua.LState) int {
		l.Push(lua.LBool(sys.decisiveRound[^sys.debugWC.playerNo&1]))
		return 1
	})
	// deg (dedicated functionality already exists in Lua)
	luaRegister(l, "defence", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.finalDefense * 100))
		return 1
	})
	luaRegister(l, "displayname", func(*lua.LState) int {
		l.Push(lua.LString(sys.debugWC.gi().displayname))
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
	luaRegister(l, "fightscreenstate", func(*lua.LState) int {
		switch strings.ToLower(strArg(l, 1)) {
		case "fightdisplay":
			l.Push(lua.LBool(sys.lifebar.ro.triggerFightDisplay))
		case "kodisplay":
			l.Push(lua.LBool(sys.lifebar.ro.triggerKODisplay))
		case "rounddisplay":
			l.Push(lua.LBool(sys.lifebar.ro.triggerFightDisplay))
		case "windisplay":
			l.Push(lua.LBool(sys.lifebar.ro.triggerWinDisplay))
		default:
			l.RaiseError("\nInvalid argument: %v\n", strArg(l, 1))
		}
		return 1
	})
	luaRegister(l, "fightscreenvar", func(*lua.LState) int {
		switch strings.ToLower(strArg(l, 1)) {
		case "info.name":
			l.Push(lua.LString(sys.lifebar.name))
		case "info.localcoord.x":
			l.Push(lua.LNumber(sys.lifebarLocalcoord[0]))
		case "info.localcoord.y":
			l.Push(lua.LNumber(sys.lifebarLocalcoord[1]))
		case "info.author":
			l.Push(lua.LString(sys.lifebar.author))
		case "round.ctrl.time":
			l.Push(lua.LNumber(sys.lifebar.ro.ctrl_time))
		case "round.over.hittime":
			l.Push(lua.LNumber(sys.lifebar.ro.over_hittime))
		case "round.over.time":
			l.Push(lua.LNumber(sys.lifebar.ro.over_time))
		case "round.over.waittime":
			l.Push(lua.LNumber(sys.lifebar.ro.over_waittime))
		case "round.over.wintime":
			l.Push(lua.LNumber(sys.lifebar.ro.over_wintime))
		case "round.slow.time":
			l.Push(lua.LNumber(sys.lifebar.ro.slow_time))
		case "round.start.waittime":
			l.Push(lua.LNumber(sys.lifebar.ro.start_waittime))
		case "round.callfight.time":
			l.Push(lua.LNumber(sys.lifebar.ro.callfight_time))
		case "time.framespercount":
			l.Push(lua.LNumber(sys.lifebar.ti.framespercount))
		default:
			l.RaiseError("\nInvalid argument: %v\n", strArg(l, 1))
		}
		return 1
	})
	luaRegister(l, "fighttime", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.gameTime))
		return 1
	})
	luaRegister(l, "firstattack", func(*lua.LState) int {
		l.Push(lua.LBool(sys.firstAttack[sys.debugWC.teamside] == sys.debugWC.playerNo))
		return 1
	})
	luaRegister(l, "gamemode", func(*lua.LState) int {
		if !nilArg(l, 1) {
			l.Push(lua.LBool(sys.gameMode == strArg(l, 1)))
		} else {
			l.Push(lua.LString(sys.gameMode))
		}
		return 1
	})
	luaRegister(l, "gamevar", func(*lua.LState) int {
		switch strings.ToLower(strArg(l, 1)) {
		case "introtime":
			if sys.intro > 0 {
				l.Push(lua.LNumber(sys.intro))
			} else {
				l.Push(lua.LNumber(0))
			}
		case "outrotime":
			if sys.intro < 0 {
				l.Push(lua.LNumber(-sys.intro))
			} else {
				l.Push(lua.LNumber(0))
			}
		case "pausetime":
			l.Push(lua.LNumber(sys.pausetime))
		case "slowtime":
			l.Push(lua.LNumber(sys.getSlowtime()))
		case "superpausetime":
			l.Push(lua.LNumber(sys.supertime))
		default:
			l.RaiseError("\nInvalid argument: %v\n", strArg(l, 1))
		}
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
	luaRegister(l, "helperindexexist", func(*lua.LState) int {
		l.Push(lua.LBool(sys.debugWC.helperByIndexExist(
			BytecodeInt(int32(numArg(l, 1)))).ToB()))
		return 1
	})
	luaRegister(l, "hitoverridden", func(*lua.LState) int {
		l.Push(lua.LBool(sys.debugWC.hoverIdx >= 0))
		return 1
	})
	luaRegister(l, "ikemenversion", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.gi().ikemenverF))
		return 1
	})
	luaRegister(l, "incustomanim", func(*lua.LState) int {
		l.Push(lua.LBool(sys.debugWC.animPN != sys.debugWC.playerNo))
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
	luaRegister(l, "inputtime", func(l *lua.LState) int {
		key := strArg(l, 1)
		var ln lua.LNumber
		// Check if keyctrl and cmd are valid
		if sys.debugWC.keyctrl[0] && sys.debugWC.cmd != nil {
			switch key {
			case "B":
				ln = lua.LNumber(sys.debugWC.cmd[0].Buffer.Bb)
			case "D":
				ln = lua.LNumber(sys.debugWC.cmd[0].Buffer.Db)
			case "F":
				ln = lua.LNumber(sys.debugWC.cmd[0].Buffer.Fb)
			case "U":
				ln = lua.LNumber(sys.debugWC.cmd[0].Buffer.Ub)
			case "L":
				ln = lua.LNumber(sys.debugWC.cmd[0].Buffer.Lb)
			case "R":
				ln = lua.LNumber(sys.debugWC.cmd[0].Buffer.Rb)
			case "a":
				ln = lua.LNumber(sys.debugWC.cmd[0].Buffer.ab)
			case "b":
				ln = lua.LNumber(sys.debugWC.cmd[0].Buffer.bb)
			case "c":
				ln = lua.LNumber(sys.debugWC.cmd[0].Buffer.cb)
			case "x":
				ln = lua.LNumber(sys.debugWC.cmd[0].Buffer.xb)
			case "y":
				ln = lua.LNumber(sys.debugWC.cmd[0].Buffer.yb)
			case "z":
				ln = lua.LNumber(sys.debugWC.cmd[0].Buffer.zb)
			case "s":
				ln = lua.LNumber(sys.debugWC.cmd[0].Buffer.sb)
			case "d":
				ln = lua.LNumber(sys.debugWC.cmd[0].Buffer.db)
			case "w":
				ln = lua.LNumber(sys.debugWC.cmd[0].Buffer.wb)
			case "m":
				ln = lua.LNumber(sys.debugWC.cmd[0].Buffer.mb)
			default:
				l.RaiseError("\nInvalid argument: %v\n", key)
				return 1
			}
		}
		l.Push(lua.LNumber(ln))
		return 1
	})
	luaRegister(l, "introstate", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.introState()))
		return 1
	})
	luaRegister(l, "isasserted", func(*lua.LState) int {
		switch strings.ToLower(strArg(l, 1)) {
		// AssertSpecialFlag (Mugen)
		case "invisible":
			l.Push(lua.LBool(sys.debugWC.asf(ASF_invisible)))
		case "noairguard":
			l.Push(lua.LBool(sys.debugWC.asf(ASF_noairguard)))
		case "noautoturn":
			l.Push(lua.LBool(sys.debugWC.asf(ASF_noautoturn)))
		case "nocrouchguard":
			l.Push(lua.LBool(sys.debugWC.asf(ASF_nocrouchguard)))
		case "nojugglecheck":
			l.Push(lua.LBool(sys.debugWC.asf(ASF_nojugglecheck)))
		case "noko":
			l.Push(lua.LBool(sys.debugWC.asf(ASF_noko)))
		case "noshadow":
			l.Push(lua.LBool(sys.debugWC.asf(ASF_noshadow)))
		case "nostandguard":
			l.Push(lua.LBool(sys.debugWC.asf(ASF_nostandguard)))
		case "nowalk":
			l.Push(lua.LBool(sys.debugWC.asf(ASF_nowalk)))
		case "unguardable":
			l.Push(lua.LBool(sys.debugWC.asf(ASF_unguardable)))
		// AssertSpecialFlag (Ikemen)
		case "animatehitpause":
			l.Push(lua.LBool(sys.debugWC.asf(ASF_animatehitpause)))
		case "animfreeze":
			l.Push(lua.LBool(sys.debugWC.asf(ASF_animfreeze)))
		case "autoguard":
			l.Push(lua.LBool(sys.debugWC.asf(ASF_autoguard)))
		case "drawunder":
			l.Push(lua.LBool(sys.debugWC.asf(ASF_drawunder)))
		case "noaibuttonjam":
			l.Push(lua.LBool(sys.debugWC.asf(ASF_noaibuttonjam)))
		case "noaicheat":
			l.Push(lua.LBool(sys.debugWC.asf(ASF_noaicheat)))
		case "noailevel":
			l.Push(lua.LBool(sys.debugWC.asf(ASF_noailevel)))
		case "noairjump":
			l.Push(lua.LBool(sys.debugWC.asf(ASF_noairjump)))
		case "nobrake":
			l.Push(lua.LBool(sys.debugWC.asf(ASF_nobrake)))
		case "nocombodisplay":
			l.Push(lua.LBool(sys.debugWC.asf(ASF_nocombodisplay)))
		case "nocrouch":
			l.Push(lua.LBool(sys.debugWC.asf(ASF_nocrouch)))
		case "nodizzypointsdamage":
			l.Push(lua.LBool(sys.debugWC.asf(ASF_nodizzypointsdamage)))
		case "nofacep2":
			l.Push(lua.LBool(sys.debugWC.asf(ASF_nofacep2)))
		case "nofallcount":
			l.Push(lua.LBool(sys.debugWC.asf(ASF_nofallcount)))
		case "nofalldefenceup":
			l.Push(lua.LBool(sys.debugWC.asf(ASF_nofalldefenceup)))
		case "nofallhitflag":
			l.Push(lua.LBool(sys.debugWC.asf(ASF_nofallhitflag)))
		case "nofastrecoverfromliedown":
			l.Push(lua.LBool(sys.debugWC.asf(ASF_nofastrecoverfromliedown)))
		case "nogetupfromliedown":
			l.Push(lua.LBool(sys.debugWC.asf(ASF_nogetupfromliedown)))
		case "noguarddamage":
			l.Push(lua.LBool(sys.debugWC.asf(ASF_noguarddamage)))
		case "noguardko":
			l.Push(lua.LBool(sys.debugWC.asf(ASF_noguardko)))
		case "noguardpointsdamage":
			l.Push(lua.LBool(sys.debugWC.asf(ASF_noguardpointsdamage)))
		case "nohardcodedkeys":
			l.Push(lua.LBool(sys.debugWC.asf(ASF_nohardcodedkeys)))
		case "nohitdamage":
			l.Push(lua.LBool(sys.debugWC.asf(ASF_nohitdamage)))
		case "noinput":
			l.Push(lua.LBool(sys.debugWC.asf(ASF_noinput)))
		case "nointroreset":
			l.Push(lua.LBool(sys.debugWC.asf(ASF_nointroreset)))
		case "nojump":
			l.Push(lua.LBool(sys.debugWC.asf(ASF_nojump)))
		case "nokofall":
			l.Push(lua.LBool(sys.debugWC.asf(ASF_nokofall)))
		case "nokovelocity":
			l.Push(lua.LBool(sys.debugWC.asf(ASF_nokovelocity)))
		case "nomakedust":
			l.Push(lua.LBool(sys.debugWC.asf(ASF_nomakedust)))
		case "nolifebaraction":
			l.Push(lua.LBool(sys.debugWC.asf(ASF_nolifebaraction)))
		case "nolifebardisplay":
			l.Push(lua.LBool(sys.debugWC.asf(ASF_nolifebardisplay)))
		case "nopowerbardisplay":
			l.Push(lua.LBool(sys.debugWC.asf(ASF_nopowerbardisplay)))
		case "noguardbardisplay":
			l.Push(lua.LBool(sys.debugWC.asf(ASF_noguardbardisplay)))
		case "nostunbardisplay":
			l.Push(lua.LBool(sys.debugWC.asf(ASF_nostunbardisplay)))
		case "nofacedisplay":
			l.Push(lua.LBool(sys.debugWC.asf(ASF_nofacedisplay)))
		case "nonamedisplay":
			l.Push(lua.LBool(sys.debugWC.asf(ASF_nonamedisplay)))
		case "nowinicondisplay":
			l.Push(lua.LBool(sys.debugWC.asf(ASF_nowinicondisplay)))
		case "noredlifedamage":
			l.Push(lua.LBool(sys.debugWC.asf(ASF_noredlifedamage)))
		case "nostand":
			l.Push(lua.LBool(sys.debugWC.asf(ASF_nostand)))
		case "noturntarget":
			l.Push(lua.LBool(sys.debugWC.asf(ASF_noturntarget)))
		case "postroundinput":
			l.Push(lua.LBool(sys.debugWC.asf(ASF_postroundinput)))
		case "projtypecollision":
			l.Push(lua.LBool(sys.debugWC.asf(ASF_projtypecollision)))
		case "runfirst":
			l.Push(lua.LBool(sys.debugWC.asf(ASF_runfirst)))
		case "runlast":
			l.Push(lua.LBool(sys.debugWC.asf(ASF_runlast)))
		case "sizepushonly":
			l.Push(lua.LBool(sys.debugWC.asf(ASF_sizepushonly)))
		// GlobalSpecialFlag (Mugen)
		case "globalnoko":
			l.Push(lua.LBool(sys.gsf(GSF_globalnoko)))
		case "globalnoshadow":
			l.Push(lua.LBool(sys.gsf(GSF_globalnoshadow)))
		case "intro":
			l.Push(lua.LBool(sys.gsf(GSF_intro)))
		case "nobardisplay":
			l.Push(lua.LBool(sys.gsf(GSF_nobardisplay)))
		case "nobg":
			l.Push(lua.LBool(sys.gsf(GSF_nobg)))
		case "nofg":
			l.Push(lua.LBool(sys.gsf(GSF_nofg)))
		case "nokoslow":
			l.Push(lua.LBool(sys.gsf(GSF_nokoslow)))
		case "nokosnd":
			l.Push(lua.LBool(sys.gsf(GSF_nokosnd)))
		case "nomusic":
			l.Push(lua.LBool(sys.gsf(GSF_nomusic)))
		case "roundnotover":
			l.Push(lua.LBool(sys.gsf(GSF_roundnotover)))
		case "timerfreeze":
			l.Push(lua.LBool(sys.gsf(GSF_timerfreeze)))
		// GlobalSpecialFlag (Ikemen)
		case "camerafreeze":
			l.Push(lua.LBool(sys.gsf(GSF_camerafreeze)))
		case "roundfreeze":
			l.Push(lua.LBool(sys.gsf(GSF_roundfreeze)))
		case "roundnotskip":
			l.Push(lua.LBool(sys.gsf(GSF_roundnotskip)))
		case "skipfightdisplay":
			l.Push(lua.LBool(sys.gsf(GSF_skipfightdisplay)))
		case "skipkodisplay":
			l.Push(lua.LBool(sys.gsf(GSF_skipkodisplay)))
		case "skiprounddisplay":
			l.Push(lua.LBool(sys.gsf(GSF_skiprounddisplay)))
		case "skipwindisplay":
			l.Push(lua.LBool(sys.gsf(GSF_skipwindisplay)))
		// SystemCharFlag
		case "disabled":
			l.Push(lua.LBool(sys.debugWC.scf(SCF_disabled)))
		case "over":
			l.Push(lua.LBool(sys.debugWC.scf(SCF_over_alive) || sys.debugWC.scf(SCF_over_ko)))
		default:
			l.RaiseError("\nInvalid argument: %v\n", strArg(l, 1))
		}
		return 1
	})
	luaRegister(l, "ishost", func(*lua.LState) int {
		l.Push(lua.LBool(sys.debugWC.isHost()))
		return 1
	})
	luaRegister(l, "jugglepoints", func(*lua.LState) int {
		id := int32(-1)
		if !nilArg(l, 1) {
			id = int32(numArg(l, 1))
		}
		l.Push(lua.LNumber(sys.debugWC.jugglePoints(id)))
		return 1
	})
	luaRegister(l, "lastplayerid", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.nextCharId - 1))
		return 1
	})
	luaRegister(l, "layerNo", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.layerNo))
		return 1
	})
	luaRegister(l, "lerp", func(*lua.LState) int {
		a, b, amount, retv := float32(numArg(l, 1)), float32(numArg(l, 2)), float32(numArg(l, 3)), float32(0)
		retv = float32(a + (b-a)*MaxF(0, MinF(amount, 1)))
		l.Push(lua.LNumber(retv))
		return 1
	})
	luaRegister(l, "localcoordX", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.cgi[sys.debugWC.playerNo].localcoord[0]))
		return 1
	})
	luaRegister(l, "localcoordY", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.cgi[sys.debugWC.playerNo].localcoord[1]))
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
	luaRegister(l, "motifstate", func(*lua.LState) int {
		switch strings.ToLower(strArg(l, 1)) {
		case "continuescreen":
			l.Push(lua.LBool(sys.continueScreenFlg))
		case "victoryscreen":
			l.Push(lua.LBool(sys.victoryScreenFlg))
		case "winscreen":
			l.Push(lua.LBool(sys.winScreenFlg))
		default:
			l.RaiseError("\nInvalid argument: %v\n", strArg(l, 1))
		}
		return 1
	})
	luaRegister(l, "movecountered", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.moveCountered()))
		return 1
	})
	luaRegister(l, "mugenversion", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.gi().mugenverF))
		return 1
	})
	luaRegister(l, "numplayer", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.numPlayer()))
		return 1
	})
	luaRegister(l, "offsetX", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.offset[0]))
		return 1
	})
	luaRegister(l, "offsetY", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.offset[1]))
		return 1
	})
	luaRegister(l, "outrostate", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.outroState()))
		return 1
	})
	luaRegister(l, "pausetime", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.pauseTimeTrigger()))
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
	luaRegister(l, "playerindexexist", func(*lua.LState) int {
		l.Push(lua.LBool(sys.playerIndexExist(
			BytecodeInt(int32(numArg(l, 1)))).ToB()))
		return 1
	})
	luaRegister(l, "playerno", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.playerNo + 1))
		return 1
	})
	luaRegister(l, "playernoexist", func(*lua.LState) int {
		l.Push(lua.LBool(sys.playerNoExist(
			BytecodeInt(int32(numArg(l, 1)))).ToB()))
		return 1
	})
	// rad (dedicated functionality already exists in Lua)
	// randomrange (dedicated functionality already exists in Lua)
	luaRegister(l, "ratiolevel", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.ocd().ratioLevel))
		return 1
	})
	luaRegister(l, "receivedhits", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.receivedHits))
		return 1
	})
	luaRegister(l, "receiveddamage", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.receivedDmg))
		return 1
	})
	luaRegister(l, "redlife", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.redLife))
		return 1
	})
	luaRegister(l, "roundtime", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.tickCount))
		return 1
	})
	luaRegister(l, "scaleX", func(*lua.LState) int {
		if sys.debugWC.csf(CSF_angledraw) {
			l.Push(lua.LNumber(sys.debugWC.angleDrawScale[0]))
		} else {
			l.Push(lua.LNumber(1))
		}
		return 1
	})
	luaRegister(l, "scaleY", func(*lua.LState) int {
		if sys.debugWC.csf(CSF_angledraw) {
			l.Push(lua.LNumber(sys.debugWC.angleDrawScale[1]))
		} else {
			l.Push(lua.LNumber(1))
		}
		return 1
	})
	luaRegister(l, "scaleZ", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.zScale))
		return 1
	})
	luaRegister(l, "sign", func(*lua.LState) int {
		v, retv := float32(numArg(l, 1)), int32(0)
		if v < 0 {
			v = -1
		} else if v > 0 {
			v = 1
		} else {
			v = 0
		}
		retv = int32(v)
		l.Push(lua.LNumber(retv))
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
	luaRegister(l, "stagebackedgedist", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.stageBackEdgeDist()))
		return 1
	})
	luaRegister(l, "stageconst", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.stage.constants[strings.ToLower(strArg(l, 1))]))
		return 1
	})
	luaRegister(l, "stagefrontedgedist", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.stageFrontEdgeDist()))
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
	luaRegister(l, "animtimesum", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.anim.curtime))
		return 1
	})
	luaRegister(l, "continue", func(*lua.LState) int {
		l.Push(lua.LBool(sys.continueFlg))
		return 1
	})
	luaRegister(l, "gameend", func(*lua.LState) int {
		l.Push(lua.LBool(sys.gameEnd))
		return 1
	})
	luaRegister(l, "gamefps", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.gameFPS))
		return 1
	})
	luaRegister(l, "gamespeed", func(*lua.LState) int {
		l.Push(lua.LNumber(100 * float32(sys.gameLogicSpeed()) / float32(sys.gameRenderSpeed())))
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
	luaRegister(l, "matchtime", func(*lua.LState) int {
		var ti int32
		for _, v := range sys.timerRounds {
			ti += v
		}
		l.Push(lua.LNumber(ti))
		return 1
	})
	luaRegister(l, "network", func(*lua.LState) int {
		l.Push(lua.LBool(sys.rollback.session != nil || sys.netConnection != nil || sys.replayFile != nil))
		return 1
	})
	luaRegister(l, "paused", func(*lua.LState) int {
		l.Push(lua.LBool(sys.paused && !sys.frameStepFlag))
		return 1
	})
	luaRegister(l, "postmatch", func(*lua.LState) int {
		l.Push(lua.LBool(sys.postMatchFlg))
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
	luaRegister(l, "stateownerid", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.chars[sys.debugWC.ss.sb.playerNo][0].id))
		return 1
	})
	luaRegister(l, "stateownername", func(*lua.LState) int {
		l.Push(lua.LString(sys.chars[sys.debugWC.ss.sb.playerNo][0].name))
		return 1
	})
	luaRegister(l, "stateownerplayerno", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.ss.sb.playerNo + 1))
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
			} else if sys.winTeam >= 0 || sys.roundState() >= 3 {
				winp = int32(sys.winTeam) + 1
			}
		}
		l.Push(lua.LNumber(winp))
		return 1
	})
}

// Legacy functions that may be removed in future, once script refactoring is finished
func deprecatedFunctions(l *lua.LState) {
	// deprecated by changeAnim
	luaRegister(l, "charChangeAnim", func(l *lua.LState) int {
		// pn, anim_no, anim_elem, ffx
		pn := int(numArg(l, 1))
		an := int32(numArg(l, 2))
		if pn >= 1 && pn <= len(sys.chars) && len(sys.chars[pn-1]) > 0 {
			c := sys.chars[pn-1]
			if c[0].selfAnimExist(BytecodeInt(an)) == BytecodeBool(true) {
				ffx := false
				if l.GetTop() >= 4 {
					ffx = boolArg(l, 4)
				}
				prefix := ""
				if ffx {
					prefix = "f"
				}
				c[0].changeAnim(an, c[0].playerNo, -1, prefix)
				if l.GetTop() >= 3 {
					c[0].setAnimElem(int32(numArg(l, 3)), 0)
				}
				l.Push(lua.LBool(true))
				return 1
			}
		}
		l.Push(lua.LBool(false))
		return 1
	})
	// deprecated by changeState
	luaRegister(l, "charChangeState", func(l *lua.LState) int {
		// pn, state_no
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
				c[0].changeState(st, -1, -1, "")
				l.Push(lua.LBool(true))
				return 1
			}
		}
		l.Push(lua.LBool(false))
		return 1
	})
	// deprecated by mapSet
	luaRegister(l, "charMapSet", func(*lua.LState) int {
		// pn, map_name, value, map_type
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
	// deprecated by playSnd
	luaRegister(l, "charSndPlay", func(l *lua.LState) int {
		// pn, group_no, sound_no, volumescale, commonSnd, channel, lowpriority, freqmul, loop, pan
		pn := int(numArg(l, 1))
		if pn < 1 || pn > len(sys.chars) || len(sys.chars[pn-1]) == 0 {
			l.RaiseError("\nPlayer not found: %v\n", pn)
		}
		f, lw, lp, stopgh, stopcs := false, false, false, false, false
		var g, n, ch, vo, priority, lc int32 = -1, 0, -1, 100, 0, 0
		var loopstart, loopend, startposition int = 0, 0, 0
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
		if l.GetTop() >= 12 {
			loopstart = int(numArg(l, 12))
		}
		if l.GetTop() >= 13 {
			loopend = int(numArg(l, 13))
		}
		if l.GetTop() >= 14 {
			startposition = int(numArg(l, 14))
		}
		if l.GetTop() >= 15 {
			lc = int32(numArg(l, 15))
		}
		if l.GetTop() >= 15 { // StopOnGetHit
			stopgh = boolArg(l, 16)
		}
		if l.GetTop() >= 16 { // StopOnChangeState
			stopcs = boolArg(l, 17)
		}
		prefix := ""
		if f {
			prefix = "f"
		}

		// If the loopcount is 0, then read the loop parameter
		if lc == 0 {
			if lp {
				sys.chars[pn-1][0].playSound(prefix, lw, -1, g, n, ch, vo, p, fr, ls, x, false, priority, loopstart, loopend, startposition, stopgh, stopcs)
			} else {
				sys.chars[pn-1][0].playSound(prefix, lw, 0, g, n, ch, vo, p, fr, ls, x, false, priority, loopstart, loopend, startposition, stopgh, stopcs)
			}

			// Otherwise, read the loopcount parameter directly
		} else {
			sys.chars[pn-1][0].playSound(prefix, lw, lc, g, n, ch, vo, p, fr, ls, x, false, priority, loopstart, loopend, startposition, stopgh, stopcs)
		}
		return 0
	})
	// deprecated by stopSnd, stopAllSound
	luaRegister(l, "charSndStop", func(l *lua.LState) int {
		if l.GetTop() == 0 {
			sys.stopAllCharSound()
			return 0
		}
		pn := int(numArg(l, 1))
		if pn < 1 || pn > len(sys.chars) || len(sys.chars[pn-1]) == 0 {
			l.RaiseError("\nPlayer not found: %v\n", pn)
		}
		sys.chars[pn-1][0].soundChannels.SetSize(0)
		return 0
	})
}
