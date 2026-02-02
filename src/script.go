package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	lua "github.com/yuin/gopher-lua"
	"gopkg.in/ini.v1"
)

// ExecFunc executes a Lua function by name and returns its boolean result.
func ExecFunc(l *lua.LState, funcName string) (bool, error) {
	// Retrieve the Lua function by name
	luaFunc := l.GetGlobal(funcName)
	if luaFunc == lua.LNil {
		return false, fmt.Errorf("function %s not found in Lua state", funcName)
	}

	// Call the Lua function
	if err := l.CallByParam(lua.P{
		Fn:      luaFunc,
		NRet:    1, // Expecting 1 return value
		Protect: true,
	}); err != nil {
		return false, fmt.Errorf("error calling Lua function %s: %v", funcName, err)
	}

	// Retrieve the return value from the Lua stack
	ret := l.Get(-1) // Get the top value
	l.Pop(1)         // Remove it from the stack

	// Convert Lua value to boolean
	if luaBool, ok := ret.(lua.LBool); ok {
		return bool(luaBool), nil
	}

	return false, fmt.Errorf("unexpected return type: %T", ret)
}

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
func tableHasKey(t *lua.LTable, key string) bool {
	return t.RawGetString(key) != lua.LNil
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

// Helper: flatten anonymous embedded structs into a parent table while
// preserving Go struct field order.
// Only applies to anonymous embedded fields without explicit `lua`/`ini` tags.
func flattenEmbeddedStructToLuaTable(l *lua.LState, table *lua.LTable, v reflect.Value) {
	if !v.IsValid() {
		return
	}
	// Unwrap pointers.
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return
		}
		v = v.Elem()
	}
	if !v.IsValid() || v.Kind() != reflect.Struct {
		return
	}
	t := v.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		// Skip unexported.
		if field.PkgPath != "" {
			continue
		}
		fv := v.Field(i)
		// Recursively flatten nested anonymous embedded structs with no tags.
		if field.Anonymous &&
			field.Tag.Get("lua") == "" &&
			field.Tag.Get("ini") == "" {
			ft := field.Type
			if ft.Kind() == reflect.Ptr {
				ft = ft.Elem()
			}
			if ft.Kind() == reflect.Struct {
				flattenEmbeddedStructToLuaTable(l, table, fv)
				continue
			}
		}
		if fv.Kind() == reflect.Ptr && fv.IsNil() {
			continue
		}
		key := field.Tag.Get("lua")
		if key == "" {
			key = field.Tag.Get("ini")
		}
		if key == "" {
			key = field.Name
		}
		// Don't clobber keys already set on the parent table.
		if table.RawGetString(key) != lua.LNil {
			continue
		}
		table.RawSetString(key, toLValue(l, fv.Interface()))
	}
}

func toLValue(l *lua.LState, v interface{}) lua.LValue {
	if v == nil {
		return lua.LNil
	}

	// If this came from encoding/json with Decoder.UseNumber(),
	// treat json.Number precisely: prefer int when possible, else float.
	if n, ok := v.(json.Number); ok {
		if i, err := n.Int64(); err == nil {
			return lua.LNumber(i)
		}
		if f, err := n.Float64(); err == nil {
			return lua.LNumber(f)
		}
		return lua.LNil
	}

	rv := reflect.ValueOf(v)
	//rt := reflect.TypeOf(v)

	switch val := v.(type) {
	case *Anim, *BGDef, *Fnt, *Sff, *Snd, *TextSprite, *Animation, *PalFX, *Rect, *Fade, *Model:
		// If it's one of our recognized pointer types, store it as userdata.
		ud := l.NewUserData()
		ud.Value = val
		return ud
	case *bgMusic:
		// Expose bgMusic as a plain Lua table
		t := l.NewTable()
		t.RawSetString("bgm", lua.LString(val.bgmusic))
		t.RawSetString("loop", lua.LNumber(val.bgmloop))
		t.RawSetString("volume", lua.LNumber(val.bgmvolume))
		t.RawSetString("loopstart", lua.LNumber(val.bgmloopstart))
		t.RawSetString("loopend", lua.LNumber(val.bgmloopend))
		t.RawSetString("startposition", lua.LNumber(val.bgmstartposition))
		t.RawSetString("freqmul", lua.LNumber(val.bgmfreqmul))
		t.RawSetString("loopcount", lua.LNumber(val.bgmloopcount))
		return t
	}

	// If pointer but not recognized:
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return lua.LNil
		}
		rv = rv.Elem() // Dereference
	}

	// Handle by reflected kind:
	switch rv.Kind() {
	case reflect.Struct:
		table := l.NewTable()
		t := rv.Type()
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)

			// Skip unexported.
			if field.PkgPath != "" {
				continue
			}

			fieldValue := rv.Field(i)

			// If this is an anonymous embedded struct (or *struct) with no explicit
			// lua/ini tag, flatten its fields into the parent table in declaration order.
			if field.Anonymous &&
				field.Tag.Get("lua") == "" &&
				field.Tag.Get("ini") == "" {

				ft := field.Type
				if ft.Kind() == reflect.Ptr {
					ft = ft.Elem()
				}
				if ft.Kind() == reflect.Struct {
					flattenEmbeddedStructToLuaTable(l, table, fieldValue)
					continue
				}
			}

			// Skip nil pointers.
			if fieldValue.Kind() == reflect.Ptr && fieldValue.IsNil() {
				continue
			}

			// Special case: pattern-mapped maps (ini:"map:...") with an EMPTY lua tag should flatten
			// their entriesdirectly into the parent Lua table instead of creating a nested subtable.
			if fieldValue.Kind() == reflect.Map {
				iniTag := field.Tag.Get("ini")
				luaTag := field.Tag.Get("lua")
				if luaTag == "" && strings.HasPrefix(strings.ToLower(iniTag), "map:") &&
					fieldValue.Type().Key().Kind() == reflect.String {

					keys := fieldValue.MapKeys()
					strKeys := make([]string, 0, len(keys))
					for _, k := range keys {
						strKeys = append(strKeys, k.String())
					}
					sort.Strings(strKeys)
					for _, sk := range strKeys {
						v := fieldValue.MapIndex(reflect.ValueOf(sk))
						table.RawSetString(sk, toLValue(l, v.Interface()))
					}
					continue
				}
			}

			// Use 'lua' tag as the key, fall back to 'ini' tag, then field name.
			key := field.Tag.Get("lua")
			if key == "" {
				key = field.Tag.Get("ini")
			}
			if key == "" {
				key = field.Name
			}

			// Recursively convert field value.
			table.RawSetString(key, toLValue(l, fieldValue.Interface()))
		}
		return table

	case reflect.Map:
		table := l.NewTable()
		// Preserve numeric keys as numeric Lua keys
		keyKind := rv.Type().Key().Kind()
		keys := rv.MapKeys()

		switch keyKind {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			ikeys := make([]int64, 0, len(keys))
			valuesByKey := make(map[int64]reflect.Value, len(keys))
			for _, k := range keys {
				iv := k.Int()
				ikeys = append(ikeys, iv)
				valuesByKey[iv] = rv.MapIndex(k)
			}
			sort.Slice(ikeys, func(i, j int) bool { return ikeys[i] < ikeys[j] })
			for _, iv := range ikeys {
				table.RawSet(lua.LNumber(iv), toLValue(l, valuesByKey[iv].Interface()))
			}
			return table

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
			ukeys := make([]uint64, 0, len(keys))
			valuesByKey := make(map[uint64]reflect.Value, len(keys))
			for _, k := range keys {
				uv := k.Uint()
				ukeys = append(ukeys, uv)
				valuesByKey[uv] = rv.MapIndex(k)
			}
			sort.Slice(ukeys, func(i, j int) bool { return ukeys[i] < ukeys[j] })
			for _, uv := range ukeys {
				table.RawSet(lua.LNumber(uv), toLValue(l, valuesByKey[uv].Interface()))
			}
			return table
		}

		// Fallback: deterministic order via stringified keys
		strKeys := make([]string, 0, len(keys))
		valuesByKey := make(map[string]reflect.Value, len(keys))
		for _, k := range keys {
			sk := fmt.Sprintf("%v", k.Interface())
			strKeys = append(strKeys, sk)
			valuesByKey[sk] = rv.MapIndex(k)
		}
		sort.Strings(strKeys)
		for _, sk := range strKeys {
			table.RawSetString(sk, toLValue(l, valuesByKey[sk].Interface()))
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
			//return lua.LNumber(int(1))
		}
		return lua.LFalse
		//return lua.LNumber(int(0))

	default:
		// Fallback for unsupported types
		return lua.LString(fmt.Sprintf("%v", rv.Interface()))
	}
}

func setNestedLuaKey(l *lua.LState, tbl *lua.LTable, key string, val lua.LValue) {
	parts := strings.Split(key, ".")
	cur := tbl
	for i, part := range parts {
		if i == len(parts)-1 {
			// last segment: set the value
			cur.RawSetString(part, val)
			return
		}
		// intermediate segment: ensure there's a table here
		existing := cur.RawGetString(part)
		if subTbl, ok := existing.(*lua.LTable); ok {
			cur = subTbl
		} else {
			// overwrite non-table or nil with a new table
			newTbl := l.NewTable()
			cur.RawSetString(part, newTbl)
			cur = newTbl
		}
	}
}

// Lowercases section names and replaces spaces with underscores.
func normalizeSectionName(name string) string {
	name = strings.TrimSpace(strings.ToLower(name))
	if name == "" {
		return name
	}
	// Collapse any whitespace runs (spaces/tabs/etc.) into single underscores.
	return strings.Join(strings.Fields(name), "_")
}

// Splits a comma-separated INI value into tokens, but only on commas that are not inside single or double quotes.
func splitIniListOutsideQuotes(s string) []string {
	var out []string
	var b strings.Builder
	inSingle := false
	inDouble := false
	escaped := false
	flush := func() {
		tok := strings.TrimSpace(b.String())
		b.Reset()
		if tok != "" {
			out = append(out, tok)
		}
	}
	for _, r := range s {
		if escaped {
			// Keep escaped char as-is; escape handling is mainly to avoid incorrectly toggling quote state on \" or \'.
			b.WriteRune(r)
			escaped = false
			continue
		}
		if r == '\\' && (inSingle || inDouble) {
			escaped = true
			b.WriteRune(r)
			continue
		}
		switch r {
		case '"':
			if !inSingle {
				inDouble = !inDouble
			}
			b.WriteRune(r)
			continue
		case '\'':
			if !inDouble {
				inSingle = !inSingle
			}
			b.WriteRune(r)
			continue
		case ',':
			if !inSingle && !inDouble {
				flush()
				continue
			}
			b.WriteRune(r)
			continue
		default:
			b.WriteRune(r)
		}
	}
	flush()
	// If we didn't actually split (i.e. no top-level commas), return the original value as a single token so callers can treat it as scalar.
	trimmed := strings.TrimSpace(s)
	if len(out) == 0 {
		if trimmed == "" {
			return nil
		}
		return []string{trimmed}
	}
	if len(out) == 1 && out[0] == trimmed {
		return []string{trimmed}
	}
	return out
}

// Converts an INI value string into a typed Lua value
func parseIniLuaValue(l *lua.LState, raw string) lua.LValue {
	s := strings.TrimSpace(raw)
	if s == "" {
		// Empty stays as empty string (matches typical INI semantics)
		return lua.LString("")
	}
	// 1) Comma-separated list (split only on commas OUTSIDE quotes) -> Lua array table
	parts := splitIniListOutsideQuotes(s)
	if len(parts) > 1 {
		tbl := l.NewTable()
		for _, p := range parts {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			tbl.Append(parseIniLuaValue(l, p))
		}
		return tbl
	}
	// 2) Quoted string wins over bool/number when it's a single scalar token
	if len(s) >= 2 && ((s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'')) {
		if unq, err := strconv.Unquote(s); err == nil {
			return lua.LString(unq)
		}
		// Fallback: strip outer quotes if Unquote fails
		return lua.LString(s[1 : len(s)-1])
	}
	// 3) Bool
	switch strings.ToLower(s) {
	case "true":
		return lua.LTrue
	case "false":
		return lua.LFalse
	}
	// 4) Number (prefer int, else float)
	if i, err := strconv.ParseInt(s, 0, 64); err == nil {
		return lua.LNumber(i)
	}
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return lua.LNumber(RoundFloat(f, 6))
	}
	// Defensive fallback (should be rare)
	return lua.LString(s)
}

func iniToLuaTable(l *lua.LState, f *ini.File) *lua.LTable {
	t := l.NewTable()
	if f == nil {
		return t
	}
	for _, sec := range f.Sections() {
		secTable := l.NewTable()
		for _, k := range sec.Keys() {
			name := k.Name()
			val := parseIniLuaValue(l, k.Value())
			if strings.Contains(name, ".") {
				// use nested tables for dotted keys
				setNestedLuaKey(l, secTable, name, val)
			} else {
				// plain key, just set directly
				secTable.RawSetString(name, val)
			}
		}
		t.RawSetString(normalizeSectionName(sec.Name()), secTable)
	}
	return t
}

func jsonToLuaValue(L *lua.LState, r io.Reader) (lua.LValue, error) {
	var data any
	dec := json.NewDecoder(r)
	dec.UseNumber() // key to get json.Number instead of float64
	if err := dec.Decode(&data); err != nil {
		return lua.LNil, err
	}
	return toLValue(L, data), nil
}

func luaToJsonValue(v lua.LValue, seen map[*lua.LTable]struct{}) (any, error) {
	if v == lua.LNil {
		return nil, nil
	}
	switch x := v.(type) {
	case lua.LBool:
		return bool(x), nil
	case lua.LNumber:
		return luaNumberToJson(x)
	case lua.LString:
		return string(x), nil
	case *lua.LTable:
		// circular detection
		if _, ok := seen[x]; ok {
			return nil, fmt.Errorf("jsonEncode: circular reference detected")
		}
		if seen == nil {
			seen = make(map[*lua.LTable]struct{}, 8)
		}
		seen[x] = struct{}{}
		defer delete(seen, x)

		// Decide array vs object
		// Array if keys are exactly 1..n with no holes and NO extra non-integer keys.
		n := x.Len() // gopher-lua's raw length (#) of array part
		isArray := true

		// Quick check: if any non 1..n key exists, it's an object
		x.ForEach(func(k, _ lua.LValue) {
			if !isArray {
				return
			}
			switch kk := k.(type) {
			case lua.LNumber:
				f := float64(kk)
				if f < 1 || f != math.Trunc(f) || int(f) > n {
					isArray = false
				}
			case lua.LString:
				isArray = false
			default:
				// keys like booleans, tables, functions => treat as object
				isArray = false
			}
		})

		if isArray {
			// Encode as []any of length n
			arr := make([]any, n)
			for i := 1; i <= n; i++ {
				item, err := luaToJsonValue(x.RawGetInt(i), seen)
				if err != nil {
					return nil, err
				}
				arr[i-1] = item
			}
			return arr, nil
		}

		// Encode as map[string]any
		obj := make(map[string]any)
		x.ForEach(func(k, v lua.LValue) {
			key := luaKeyToString(k)
			val, err := luaToJsonValue(v, seen)
			if err != nil {
				// capture error by storing a sentinel, then overwrite below
				obj["__ERROR__"] = err.Error()
				return
			}
			obj[key] = val
		})
		if errStr, bad := obj["__ERROR__"]; bad {
			return nil, fmt.Errorf("%v", errStr)
		}
		return obj, nil

	case *lua.LUserData:
		// If you need to support userdata, customize here.
		// Default: refuse (safer than stringifying).
		return nil, fmt.Errorf("jsonEncode: cannot encode userdata (%T)", x.Value)

	case *lua.LFunction, *lua.LChannel:
		return nil, fmt.Errorf("jsonEncode: unsupported Lua type %T", x)

	default:
		// Shouldn't happen, but be defensive.
		return nil, fmt.Errorf("jsonEncode: unknown Lua type %T", v)
	}
}

func luaNumberToJson(n lua.LNumber) (any, error) {
	f := float64(n)
	if math.IsNaN(f) || math.IsInf(f, 0) {
		return nil, fmt.Errorf("jsonEncode: NaN/Inf not permitted in JSON numbers")
	}
	// Emit as integer when possible to avoid "1.0" in output
	if f == math.Trunc(f) {
		// Prefer int64 if in range
		if f >= float64(math.MinInt64) && f <= float64(math.MaxInt64) {
			return int64(f), nil
		}
		// Large positive integers can be encoded as uint64 (JSON number)
		if f >= 0 && f <= float64(^uint64(0)) {
			return uint64(f), nil
		}
		// Fall through to float64 if out of range
	}
	return f, nil
}

func luaKeyToString(k lua.LValue) string {
	switch kk := k.(type) {
	case lua.LString:
		return string(kk)
	case lua.LNumber:
		f := float64(kk)
		if f == math.Trunc(f) {
			// integer-like; avoid "1.0" keys
			return strconv.FormatInt(int64(f), 10)
		}
		// non-integer numeric key
		return strconv.FormatFloat(f, 'f', -1, 64)
	default:
		return k.String() // fallback
	}
}

// -------------------------------------------------------------------------------------------------
// Register external functions to be called from Lua scripts
func systemScriptInit(l *lua.LState) {
	triggerFunctions(l)
	luaRegister(l, "addChar", func(l *lua.LState) int {
		if sc := sys.sel.AddChar(strArg(l, 1)); sc != nil {
			if !nilArg(l, 2) {
				entries := SplitAndTrim(strArg(l, 2), ",")
				if sc.scp == nil {
					sc.scp = newSelectCharParams()
				}
				sc.scp.AppendParams(entries)
				// Feed normalized music params to Music
				sc.music.AppendParams(sc.scp.MusicEntries())
			}
			l.Push(lua.LBool(true))
		} else {
			l.Push(lua.LBool(false))
		}
		return 1
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
		if ss, err := sys.sel.AddStage(strArg(l, 1)); err == nil {
			if !nilArg(l, 2) {
				entries := SplitAndTrim(strArg(l, 2), ",")
				if ss.ssp == nil {
					ss.ssp = newSelectStageParams()
				}
				ss.ssp.AppendParams(entries)
				// Feed normalized music params to Music.
				ss.music.AppendParams(ss.ssp.MusicEntries())
			}
			l.Push(lua.LBool(true))
		} else {
			l.Push(lua.LBool(false))
		}
		return 1
	})
	luaRegister(l, "animAddPos", func(*lua.LState) int {
		a, ok := toUserData(l, 1).(*Anim)
		if !ok {
			userDataError(l, 1, a)
		}
		a.AddPos(float32(numArg(l, 2)), float32(numArg(l, 3)))
		return 0
	})
	luaRegister(l, "animApplyVel", func(*lua.LState) int {
		a, ok := toUserData(l, 1).(*Anim)
		if !ok {
			userDataError(l, 1, a)
		}
		src, ok := toUserData(l, 2).(*Anim)
		if !ok {
			userDataError(l, 2, src)
		}
		a.velocityInit = src.velocityInit
		a.xvel, a.yvel = src.xvel, src.yvel
		a.vel = src.vel
		return 0
	})
	luaRegister(l, "animDebug", func(*lua.LState) int {
		a, ok := toUserData(l, 1).(*Anim)
		if !ok {
			userDataError(l, 1, a)
		}
		str := ""
		if !nilArg(l, 2) {
			str = strArg(l, 2)
		}
		fmt.Printf("%s *Anim=%p %+v\n", str, a, *a)
		if a.anim == nil {
			fmt.Printf("%s *Animation=nil\n", str)
		} else {
			fmt.Printf("%s *Animation=%p %+v\n", str, a.anim, *a.anim)
		}
		return 0
	})
	luaRegister(l, "animDraw", func(*lua.LState) int {
		a, ok := toUserData(l, 1).(*Anim)
		if !ok {
			userDataError(l, 1, a)
		}
		layer := a.layerno
		if !nilArg(l, 2) {
			layer = int16(numArg(l, 2))
		}
		aSnap := *a
		layerLocal := layer
		sys.luaQueueLayerDraw(int(layerLocal), func() {
			(&aSnap).Draw(layerLocal)
		})
		return 0
	})
	luaRegister(l, "animGetLength", func(*lua.LState) int {
		a, ok := toUserData(l, 1).(*Anim)
		if !ok {
			userDataError(l, 1, a)
		}
		var length, totaltime int32
		if a.anim != nil {
			length = a.GetLength()
			totaltime = a.anim.totaltime
		}
		l.Push(lua.LNumber(length))
		l.Push(lua.LNumber(totaltime))
		return 2
	})
	luaRegister(l, "animGetPreloadedCharData", func(l *lua.LState) int {
		if anim := sys.sel.GetChar(int(numArg(l, 1))).anims.get(int32(numArg(l, 2)), int32(numArg(l, 3))); anim != nil {
			a := NewAnim(nil, "")
			a.anim = anim
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
		if anim := sys.sel.GetStage(int(numArg(l, 1))).anims.get(int32(numArg(l, 2)), int32(numArg(l, 3))); anim != nil {
			a := NewAnim(nil, "")
			a.anim = anim
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
			spr = a.anim.sff.GetSprite(uint16(numArg(l, 2)), uint16(numArg(l, 3)))
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
	luaRegister(l, "animLoadPalettes", func(*lua.LState) int {
		a, ok := toUserData(l, 1).(*Anim)
		if !ok {
			userDataError(l, 1, a)
			return 0
		}
		if sys.usePalette == true {
			loadCharPalettes(a.anim.sff, a.anim.sff.filename, int(numArg(l, 2)))
		}
		return 0
	})
	luaRegister(l, "animNew", func(*lua.LState) int {
		s, ok := toUserData(l, 1).(*Sff)
		if !ok {
			s = newSff()
			//userDataError(l, 1, s)
		}
		var anim *Anim
		switch l.Get(2).Type() {
		case lua.LTString:
			// Parse inline AIR text
			act := strArg(l, 2)
			anim = NewAnim(s, act)
			if anim == nil {
				l.RaiseError("\nFailed to read the data: %v\n", act)
			}
		case lua.LTUserData:
			// Accept *Animation
			a2, ok := toUserData(l, 2).(*Animation)
			if !ok {
				userDataError(l, 2, a2)
			}
			anim = NewAnim(nil, "")
			anim.anim = a2.ShallowCopy()
			if anim.anim == nil {
				l.RaiseError("\nanimNew: *Animation is nil\n")
			}
			// Ensure required pointers are present
			if anim.anim.sff == nil {
				anim.anim.sff = s
			}
			if anim.anim.palettedata == nil && anim.anim.sff != nil {
				anim.anim.palettedata = &anim.anim.sff.palList
			}
			// Start from the beginning for a "new" Anim instance
			anim.anim.Reset()
		default:
			l.RaiseError("\nanimNew: expected string or *Animation, got %v\n", l.Get(2).Type())
		}
		l.Push(newUserData(l, anim))
		return 1
	})
	luaRegister(l, "animPrepare", func(l *lua.LState) int {
		// Prepares an animation copy so each player can apply their own palette
		a, ok := toUserData(l, 1).(*Anim)
		if !ok {
			userDataError(l, 1, a)
			return 0
		}
		if sys.usePalette {
			copyAnim := a.Copy()
			char := sys.sel.GetChar(int(numArg(l, 2)))
			for _, c := range copyAnim.anim.frames {
				if c.Group < 0 || c.Number < 0 {
					continue
				}
				spr, ok := copyAnim.anim.sff.sprites[[2]uint16{uint16(c.Group), uint16(c.Number)}]
				if !ok || spr == nil {
					continue
				}
				// Remove base palette
				if len(char.pal) > 0 && spr.palidx == 0 {
					spr.Pal = nil
				}
			}
			l.Push(newUserData(l, copyAnim))
		} else {
			l.Push(newUserData(l, a))
		}
		return 1
	})
	luaRegister(l, "animReset", func(*lua.LState) int {
		a, ok := toUserData(l, 1).(*Anim)
		if !ok {
			userDataError(l, 1, a)
		}
		// If no table provided reset everything.
		if nilArg(l, 2) {
			a.Reset()
			return 0
		}
		// Apply actions as we encounter valid keys.
		tableArg(l, 2).ForEach(func(key, value lua.LValue) {
			switch v := value.(type) {
			case lua.LString:
				k := strings.ToLower(string(v))
				switch k {
				case "anim":
					if a.anim != nil {
						a.anim.Reset()
					}
				case "pos":
					a.SetPos(a.offsetInit[0], a.offsetInit[1])
				case "scale":
					a.SetScale(a.scaleInit[0], a.scaleInit[1])
				case "window":
					a.SetWindow(a.windowInit)
				case "velocity":
					a.SetVelocity(a.velocityInit[0], a.velocityInit[1])
				case "palfx":
					if a.palfx != nil {
						a.palfx.clear()
					}
				default:
					l.RaiseError("\nInvalid table key: %v\n", k)
				}
			default:
				l.RaiseError("\nInvalid table key type: %v\n", fmt.Sprintf("%T\n", value))
			}
		})
		return 0
	})
	luaRegister(l, "animSetAccel", func(*lua.LState) int {
		a, ok := toUserData(l, 1).(*Anim)
		if !ok {
			userDataError(l, 1, a)
		}
		a.SetAccel(float32(numArg(l, 2)), float32(numArg(l, 3)))
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
	luaRegister(l, "animSetAnimation", func(*lua.LState) int {
		a, ok := toUserData(l, 1).(*Anim)
		if !ok {
			userDataError(l, 1, a)
		}

		switch l.Get(2).Type() {
		case lua.LTString:
			act := strArg(l, 2)
			// Need an SFF to parse AIR text.
			var sff *Sff
			if a.anim != nil {
				sff = a.anim.sff
			}
			if sff == nil {
				l.RaiseError("\nanimSetAnimation: cannot set from string without a valid SFF (create with animNew(sff, ...) or pass *Animation)\n")
			}
			tmp := NewAnim(sff, act)
			if tmp == nil {
				l.RaiseError("\nFailed to read the data: %v\n", act)
			}
			a.anim = tmp.anim
			if a.anim != nil {
				a.anim.Reset()
			}
		case lua.LTUserData:
			a2, ok := toUserData(l, 2).(*Animation)
			if !ok {
				userDataError(l, 2, a2)
			}
			na := a2.ShallowCopy()
			if na == nil {
				l.RaiseError("\nanimSetAnimation: *Animation is nil\n")
			}
			// Preserve/repair required pointers if missing
			if na.sff == nil && a.anim != nil && a.anim.sff != nil {
				na.sff = a.anim.sff
			}
			if na.palettedata == nil && na.sff != nil {
				na.palettedata = &na.sff.palList
			}
			na.Reset()
			a.anim = na
		default:
			l.RaiseError("\nanimSetAnimation: expected string or *Animation, got %v\n", l.Get(2).Type())
		}
		// Allow update immediately after swapping animations
		a.lastUpdateFrame = -1
		// If swapping to a shared Animation that may have advanced this frame elsewhere, clear the stamp
		if a.anim != nil {
			a.anim.lastActionFrame = -1
		}
		return 0
	})
	luaRegister(l, "animSetFriction", func(*lua.LState) int {
		a, ok := toUserData(l, 1).(*Anim)
		if !ok {
			userDataError(l, 1, a)
		}
		a.friction[0] = float32(numArg(l, 2))
		a.friction[1] = float32(numArg(l, 3))
		return 0
	})
	luaRegister(l, "animSetLayerno", func(*lua.LState) int {
		a, ok := toUserData(l, 1).(*Anim)
		if !ok {
			userDataError(l, 1, a)
		}
		a.layerno = int16(numArg(l, 2))
		return 0
	})
	luaRegister(l, "animSetLocalcoord", func(*lua.LState) int {
		a, ok := toUserData(l, 1).(*Anim)
		if !ok {
			userDataError(l, 1, a)
		}
		a.SetLocalcoord(float32(numArg(l, 2)), float32(numArg(l, 3)))
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
		a.facing = float32(numArg(l, 2))
		return 0
	})
	luaRegister(l, "animSetMaxDist", func(*lua.LState) int {
		a, ok := toUserData(l, 1).(*Anim)
		if !ok {
			userDataError(l, 1, a)
		}
		a.SetMaxDist(float32(numArg(l, 2)), float32(numArg(l, 3)))
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
		x, y := a.offsetInit[0], a.offsetInit[1]
		if !nilArg(l, 2) {
			x = float32(numArg(l, 2))
		}
		if !nilArg(l, 3) {
			y = float32(numArg(l, 3))
		}
		a.SetPos(x, y)
		return 0
	})
	luaRegister(l, "animPaletteGet", func(*lua.LState) int {
		// Userdata
		a, ok := toUserData(l, 1).(*Anim)
		if !ok {
			userDataError(l, 1, a)
		}
		pal := a.anim.palettedata.palettes[int(numArg(l, 2))-1]
		tbl := l.NewTable()
		for k, v := range pal {
			col := l.NewTable()
			// R, G, B, A
			col.RawSetInt(1, lua.LNumber(v&0x000000FF))
			col.RawSetInt(2, lua.LNumber(v&0x0000FF00>>8))
			col.RawSetInt(3, lua.LNumber(v&0x00FF0000>>16))
			col.RawSetInt(4, lua.LNumber(v&0xFF000000>>24))
			tbl.RawSetInt(k+1, col)
		}
		l.Push(tbl)
		return 1
	})
	luaRegister(l, "animPaletteSet", func(*lua.LState) int {
		// Userdata
		a, ok := toUserData(l, 1).(*Anim)
		if !ok {
			userDataError(l, 1, a)
		}
		pal := int(numArg(l, 2)) - 1
		palData := a.anim.palettedata.palettes[pal]
		tableArg(l, 3).ForEach(func(key, value lua.LValue) {
			var color uint32
			switch v := value.(type) {
			case *lua.LTable:
				// animPaletteGet to uint32
				color += uint32((int(lua.LVAsNumber(v.RawGetInt(1))) & 0xff))
				color += uint32((int(lua.LVAsNumber(v.RawGetInt(2))) & 0xff) << 8)
				color += uint32((int(lua.LVAsNumber(v.RawGetInt(3))) & 0xff) << 16)
				color += uint32((int(lua.LVAsNumber(v.RawGetInt(4))) & 0xff) << 24)
				palData[int(lua.LVAsNumber(key))-1] = color
			}
		})
		a.anim.palettedata.SetSource(pal, palData)
		a.anim.palettedata.PalTex[pal] = NewTextureFromPalette(palData)
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
	luaRegister(l, "animSetVelocity", func(*lua.LState) int {
		a, ok := toUserData(l, 1).(*Anim)
		if !ok {
			userDataError(l, 1, a)
		}
		a.SetVelocity(float32(numArg(l, 2)), float32(numArg(l, 3)))
		return 0
	})
	luaRegister(l, "animSetWindow", func(*lua.LState) int {
		a, ok := toUserData(l, 1).(*Anim)
		if !ok {
			userDataError(l, 1, a)
		}
		a.SetWindow([4]float32{float32(numArg(l, 2)), float32(numArg(l, 3)), float32(numArg(l, 4)), float32(numArg(l, 5))})
		return 0
	})
	luaRegister(l, "animSetXShear", func(*lua.LState) int {
		a, ok := toUserData(l, 1).(*Anim)
		if !ok {
			userDataError(l, 1, a)
		}
		a.xshear = float32(numArg(l, 2))
		return 0
	})
	luaRegister(l, "animSetAngle", func(*lua.LState) int {
		a, ok := toUserData(l, 1).(*Anim)
		if !ok {
			userDataError(l, 1, a)
		}
		a.rot.angle = float32(numArg(l, 2))
		return 0
	})
	luaRegister(l, "animSetXAngle", func(*lua.LState) int {
		a, ok := toUserData(l, 1).(*Anim)
		if !ok {
			userDataError(l, 1, a)
		}
		a.rot.xangle = float32(numArg(l, 2))
		return 0
	})
	luaRegister(l, "animSetYAngle", func(*lua.LState) int {
		a, ok := toUserData(l, 1).(*Anim)
		if !ok {
			userDataError(l, 1, a)
		}
		a.rot.yangle = float32(numArg(l, 2))
		return 0
	})
	luaRegister(l, "animSetProjection", func(*lua.LState) int {
		a, ok := toUserData(l, 1).(*Anim)
		if !ok {
			userDataError(l, 1, a)
		}
		switch l.Get(2).Type() {

		case lua.LTNumber:
			a.projection = int32(numArg(l, 2))

		case lua.LTString:
			switch strings.ToLower(strings.TrimSpace(l.Get(2).String())) {
			case "orthographic":
				a.projection = int32(Projection_Orthographic)
			case "perspective":
				a.projection = int32(Projection_Perspective)
			case "perspective2":
				a.projection = int32(Projection_Perspective2)
			}
		}
		return 0
	})
	luaRegister(l, "animSetFocalLength", func(*lua.LState) int {
		a, ok := toUserData(l, 1).(*Anim)
		if !ok {
			userDataError(l, 1, a)
		}
		a.fLength = float32(numArg(l, 2))
		return 0
	})
	luaRegister(l, "animUpdate", func(*lua.LState) int {
		a, ok := toUserData(l, 1).(*Anim)
		if !ok {
			userDataError(l, 1, a)
		}
		force := false
		if !nilArg(l, 2) {
			force = boolArg(l, 2)
		}
		a.Update(force)
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

			anim.SetPos(anim.offsetInit[0], anim.offsetInit[1])
			anim.AddPos(x, y)
			anim.facing = facing

			aSnap := *anim

			if v := item.RawGetString("scale"); v.Type() == lua.LTTable {
				sTbl := v.(*lua.LTable)
				sclX := float32(lua.LVAsNumber(sTbl.RawGetInt(1)))
				sclY := float32(lua.LVAsNumber(sTbl.RawGetInt(2)))
				if sclX != 0 || sclY != 0 {
					(&aSnap).SetScale(sclX, sclY)
				}
			}

			if v := item.RawGetString("xshear"); v != lua.LNil {
				aSnap.xshear = float32(lua.LVAsNumber(v))
			}

			if v := item.RawGetString("angle"); v != lua.LNil {
				aSnap.rot.angle = float32(lua.LVAsNumber(v))
			}
			if v := item.RawGetString("xangle"); v != lua.LNil {
				aSnap.rot.xangle = float32(lua.LVAsNumber(v))
			}
			if v := item.RawGetString("yangle"); v != lua.LNil {
				aSnap.rot.yangle = float32(lua.LVAsNumber(v))
			}

			if v := item.RawGetString("projection"); v != lua.LNil {
				switch v.Type() {
				case lua.LTNumber:
					aSnap.projection = int32(lua.LVAsNumber(v))
				case lua.LTString:
					switch strings.ToLower(strings.TrimSpace(v.String())) {
					case "orthographic":
						aSnap.projection = int32(Projection_Orthographic)
					case "perspective":
						aSnap.projection = int32(Projection_Perspective)
					case "perspective2":
						aSnap.projection = int32(Projection_Perspective2)
					}
				}
			}

			if v := item.RawGetString("focallength"); v != lua.LNil {
				aSnap.fLength = float32(lua.LVAsNumber(v))
			}

			layerVal := item.RawGetString("layerno")
			layer := aSnap.layerno
			if layerVal != lua.LNil {
				layer = int16(lua.LVAsNumber(layerVal))
			}

			layerLocal := layer
			sys.luaQueueLayerDraw(int(layerLocal), func() {
				(&aSnap).Draw(layerLocal)
			})
			aSnap.Update(true)
		})
		return 0
	})
	luaRegister(l, "bgDebug", func(*lua.LState) int {
		bg, ok := toUserData(l, 1).(*BGDef)
		if !ok {
			userDataError(l, 1, bg)
		}
		str := ""
		if !nilArg(l, 2) {
			str = strArg(l, 2)
		}
		fmt.Printf("%s *BGDef=%p %+v\n", str, bg, *bg)
		for i, v := range bg.bg {
			if v == nil {
				fmt.Printf("%s bg.bg[%d]=nil\n", str, i)
				continue
			}
			fmt.Printf("%s bg.bg[%d]=%p %+v\n", str, i, v, *v)
		}
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
		// BGDef.Draw mutates BGDef state (time/BGCtrl/anim). Copying BGDef makes those updates happen on the copy only. Keep a pointer to the live BGDef.
		bgLocal := bg
		layerLocal := layer
		xLocal, yLocal, sclLocal := x, y, scl
		sys.luaQueueLayerDraw(int(layerLocal), func() {
			bgLocal.Draw(layerLocal, xLocal, yLocal, sclLocal)
		})
		return 0
	})
	luaRegister(l, "bgNew", func(*lua.LState) int {
		s, ok := toUserData(l, 1).(*Sff)
		if !ok {
			userDataError(l, 1, s)
		}
		model, ok := toUserData(l, 4).(*Model)
		var defaultlayer int32
		if !nilArg(l, 5) {
			defaultlayer = int32(numArg(l, 5))
		}
		bg, err := loadBGDef(s, model, strArg(l, 2), strArg(l, 3), defaultlayer)
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
		bg.Reset()
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
	luaRegister(l, "validatePal", func(l *lua.LState) int {
		palReq := int(numArg(l, 1))
		charRef := int(numArg(l, 2))
		valid := sys.sel.ValidatePalette(charRef, palReq)
		l.Push(lua.LNumber(valid))
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
		colLocal := col
		srcLocal, dstLocal := src, dst
		sys.luaQueuePreDraw(func() {
			FillRect(sys.scrrect, colLocal, [2]int32{srcLocal, dstLocal})
		})
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

		name := strArg(l, 2)
		cmdstr := strArg(l, 3)

		time := cl.DefaultTime
		buftime := cl.DefaultBufferTime
		bufferHitpause := cl.DefaultBufferHitpause
		bufferPauseend := cl.DefaultBufferPauseEnd
		steptime := cl.DefaultStepTime

		if !nilArg(l, 4) {
			time = int32(numArg(l, 4))
		}
		if !nilArg(l, 5) {
			buftime = Max(1, int32(numArg(l, 5)))
		}
		if !nilArg(l, 6) {
			bufferHitpause = boolArg(l, 6)
		}
		if !nilArg(l, 7) {
			bufferPauseend = boolArg(l, 7)
		}
		if !nilArg(l, 8) {
			steptime = int32(numArg(l, 8))
		}

		if err := cl.AddCommand(
			name,
			cmdstr,
			time,
			buftime,
			bufferHitpause,
			bufferPauseend,
			steptime,
		); err != nil {
			l.RaiseError(err.Error())
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
	luaRegister(l, "commandDebug", func(*lua.LState) int {
		cl, ok := toUserData(l, 1).(*CommandList)
		if !ok || cl == nil {
			userDataError(l, 1, cl)
			return 0
		}
		str := ""
		if !nilArg(l, 2) {
			str = strArg(l, 2)
		}
		var buf *InputBuffer
		if cl.Buffer != nil {
			buf = cl.Buffer
		}
		fmt.Printf("%s *CommandList=%p Names=%d Groups=%d Buffer=%p\n",
			str, cl, len(cl.Names), len(cl.Commands), buf)
		for name, idx := range cl.Names {
			if idx < 0 || idx >= len(cl.Commands) {
				fmt.Printf("%s  %q idx=%d (out of range)\n", str, name, idx)
				continue
			}
			fmt.Printf("%s  %q idx=%d variants=%d\n", str, name, idx, len(cl.Commands[idx]))
			for j := range cl.Commands[idx] {
				c := &cl.Commands[idx][j]
				fmt.Printf("%s    %d: time=%d buftime=%d steptime=%d hitpause=%v pauseend=%v shared=%v autogreater=%v\n",
					str, j,
					c.maxtime, c.maxbuftime, c.maxsteptime,
					c.buffer_hitpause, c.buffer_pauseend, c.buffer_shared,
					c.autogreater,
				)
			}
		}
		fmt.Printf("%s *Buffer=%p %+v\n", str, buf, *buf)
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
		if cl.InputUpdate(nil, controller) {
			cl.Step(false, false, false, false, 0)
		}
		return 0
	})
	luaRegister(l, "commandNew", func(l *lua.LState) int {
		var controllerNo int
		if !nilArg(l, 1) {
			controllerNo = int(numArg(l, 1))
		}
		cl := NewCommandList(NewInputBuffer())
		if controllerNo > 0 {
			idx := controllerNo - 1 // 0-based index
			// Grow sys.commandLists if needed
			if idx >= len(sys.commandLists) {
				tmp := make([]*CommandList, idx+1)
				copy(tmp, sys.commandLists)
				sys.commandLists = tmp
			}
			sys.commandLists[idx] = cl
		}
		l.Push(newUserData(l, cl))
		return 1
	})
	luaRegister(l, "computeRanking", func(l *lua.LState) int {
		mode := strArg(l, 1)
		cleared, place := computeAndSaveRanking(mode)
		l.Push(lua.LBool(cleared))
		l.Push(lua.LNumber(place))
		return 2
	})
	luaRegister(l, "connected", func(*lua.LState) int {
		l.Push(lua.LBool(sys.netConnection.IsConnected())) // No need to check rollback here as this deals with the main menu connection
		return 1
	})
	luaRegister(l, "endMatch", func(*lua.LState) int {
		sys.motif.PauseMenu["pause_menu"].FadeOut.FadeData.init(sys.motif.fadeOut, false)
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
			sys.netConnection.end()
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
	//luaRegister(l, "fadeInActive", func(*lua.LState) int {
	//	l.Push(lua.LBool(sys.motif.fadeIn.isActive()))
	//	return 1
	//})
	//luaRegister(l, "fadeInInit", func(*lua.LState) int {
	//	f, ok := toUserData(l, 1).(*Fade)
	//	if !ok {
	//		userDataError(l, 1, f)
	//	}
	//	f.init(sys.motif.fadeIn, true)
	//	return 0
	//})
	//luaRegister(l, "fadeOutActive", func(*lua.LState) int {
	//	l.Push(lua.LBool(sys.motif.fadeOut.isActive()))
	//	return 1
	//})
	//luaRegister(l, "fadeOutInit", func(*lua.LState) int {
	//	f, ok := toUserData(l, 1).(*Fade)
	//	if !ok {
	//		userDataError(l, 1, f)
	//	}
	//	f.init(sys.motif.fadeOut, false)
	//	return 0
	//})
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
		sys.luaQueueLayerDraw(2, func() {
			FillRect(sys.scrrect, col, [2]int32{src, dst})
		})
		l.Push(lua.LBool(true))
		return 1
	})
	luaRegister(l, "fileExists", func(l *lua.LState) int {
		path := strArg(l, 1)
		l.Push(lua.LBool(FileExist(path) != ""))
		return 1
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
	luaRegister(l, "fontNew", func(l *lua.LState) int {
		var height int32 = -1
		if !nilArg(l, 2) {
			height = int32(numArg(l, 2))
		}
		filename := SearchFile(strArg(l, 1), []string{"font/", sys.motif.Def, "", "data/"})
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
		sys.luaDiscardDrawQueue()
		sys.gameRunning = true
		sys.endMatch = false
		// Anonymous function to load characters and stages, and/or wait for them to finish loading
		load := func() error {
			sys.loader.runTread()
			for sys.loader.state != LS_Complete {
				if sys.loader.state == LS_Error {
					return sys.loader.err
				} else if sys.loader.state == LS_Cancel {
					return nil
				}
				sys.await(sys.cfg.Video.Framerate)
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
			sys.debugRef = [2]int{}
			sys.roundsExisted = [2]int32{}
			sys.matchWins = [2]int32{}
			sys.scoreRounds = [][2]float32{}

			// Reset lifebars
			for i := range sys.lifebar.wi {
				sys.lifebar.wi[i].clear()
			}

			sys.draws = 0
			sys.statsLog.startMatch()

			// Anonymous function to perform gameplay
			fight := func() (int32, error) {
				// Reset character list
				if sys.round == 1 {
					sys.charList.clear()
				}

				// Load characters and stage
				if err := load(); err != nil {
					return -1, err
				}
				if sys.loader.state == LS_Cancel {
					return -1, nil
				}

				// Assign round start player ID's
				sys.initPlayerID()

				for i, c := range sys.chars {
					if len(c) > 0 {
						// Add or replace in charList
						if sys.round == 1 {
							sys.charList.add(c[0])
						} else if c[0].roundsExisted() == 0 {
							if !sys.charList.replace(c[0], i, 0) {
								panic(fmt.Errorf("failed to replace player: %v", i))
							}
						}

						// Load palette if character is just joining the match
						if c[0].roundsExisted() == 0 {
							c[0].loadPalette()
						}

						// Copy each other's command lists
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
					// Match was quit in netplay or quick vs command line
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
				// Hard reset: drop the incomplete or challenger match stats and start a fresh one
				if winp == -2 || sys.gameMode == "challenger" {
					sys.statsLog.abortMatch()
					sys.statsLog.startMatch()
				}
				// If a team won, and not going to the next character in turns mode, break
				if winp < 0 ||
					(sys.tmode[0] != TM_Turns && sys.tmode[1] != TM_Turns) ||
					sys.wins[0] >= sys.matchWins[0] ||
					sys.wins[1] >= sys.matchWins[1] ||
					sys.gameEnd ||
					sys.endMatch {
					break
				}

				// Reset roundsExisted to 0 if the losing side is on turns mode
				for i := 0; i < 2; i++ {
					if sys.tmode[i] == TM_Turns && sys.effectiveLoss[i] {
						sys.roundsExisted[i] = 0
						// Increment number of lifebar KO portraits to render
						sys.lifebar.fa[TM_Turns][i].numko++
						sys.lifebar.nm[TM_Turns][i].numko++
					}
				}

				sys.loader.reset()
			}

			// If not restarting match
			if winp != -2 {
				//sys.endMatch = false
				sys.esc = false
				sys.keyInput = KeyUnknown
				sys.statsLog.finalizeMatch()
				// Cleanup
				sys.timerStart = 0
				sys.timerRounds = []int32{}
				sys.scoreStart = [2]float32{}
				//sys.scoreRounds = [][2]float32{}
				sys.timerCount = []int32{}
				sys.sel.cdefOverwrite = make(map[int]string)
				sys.sel.sdefOverwrite = ""
				if sys.playBgmFlg {
					sys.bgm.Stop()
					sys.playBgmFlg = false
				}
				sys.clearMatchSound()
				sys.allPalFX = newPalFX()
				sys.bgPalFX = newPalFX()
				sys.resetGblEffect()
				sys.dialogueForce = 0
				sys.dialogueBarsFlg = false
				sys.noSoundFlg = false
				sys.postMatchFlg = false
				sys.preMatchTime += sys.matchTime
				sys.matchTime = 0
				sys.consoleText = []string{}
				sys.stageLoopNo = 0
				sys.paused = false
				sys.gameRunning = false
				sys.clearSpriteData()
				sys.luaDiscardDrawQueue()
				//if !sys.skipMotifScaling() {
				sys.setGameSize(sys.scrrect[2], sys.scrrect[3])
				//}
				l.Push(lua.LNumber(winp))
				l.Push(lua.LNumber(sys.motif.ch.controllerNo + 1))
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
		tbl.RawSetString("localcoord", lua.LNumber(c.localcoord[0]))
		tbl.RawSetString("portraitscale", lua.LNumber(c.portraitscale))
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
			idx := int(RandI(0, int32(len(c.pal)-1)))
			l.Push(lua.LNumber(c.pal[idx]))
		} else {
			l.Push(lua.LNumber(1))
		}
		return 1
	})
	luaRegister(l, "getCharSelectParams", func(*lua.LState) int {
		c := sys.sel.GetChar(int(numArg(l, 1)))
		lv := toLValue(l, c.scp)
		lTable, ok := lv.(*lua.LTable)
		if !ok {
			l.RaiseError("Error: 'lv' is not a *lua.LTable")
			return 0
		}
		l.Push(lTable)
		return 1
	})
	luaRegister(l, "getClipboardString", func(*lua.LState) int {
		s := sys.window.GetClipboardString()
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
		if _, ok := sys.cmdFlags[strArg(l, 1)]; !ok {
			return 0
		}
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
	luaRegister(l, "getCommandInputSource", func(l *lua.LState) int {
		pn := int(numArg(l, 1))
		if pn < 1 || pn > len(sys.commandInputSource) {
			l.RaiseError("\nInvalid player number: %v\n", pn)
		}
		l.Push(lua.LNumber(sys.commandInputSource[pn-1] + 1))
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
	luaRegister(l, "getGameParams", func(*lua.LState) int {
		lv := toLValue(l, sys.sel.gameParams)
		lTable, ok := lv.(*lua.LTable)
		if !ok {
			l.RaiseError("Error: 'lv' is not a *lua.LTable")
			return 0
		}
		l.Push(lTable)
		return 1
	})
	luaRegister(l, "getInput", func(l *lua.LState) int {
		var players []int
		// Collect player numbers (1-based) from arg #1
		switch v := l.Get(1).(type) {
		case *lua.LTable:
			v.ForEach(func(_ lua.LValue, val lua.LValue) {
				n, ok := val.(lua.LNumber)
				if !ok {
					return
				}
				pn := int(n)
				if pn < 1 || pn > len(sys.commandLists) {
					return
				}
				players = append(players, pn)
			})
		case lua.LNumber:
			pn := int(v)
			if pn == -1 {
				players = make([]int, 0, len(sys.commandLists))
				for i := 1; i <= len(sys.commandLists); i++ {
					players = append(players, i)
				}
			} else {
				players = append(players, pn)
			}
		default:
			l.Push(lua.LBool(false))
			return 1
		}
		top := l.GetTop()
		for _, pn := range players {
			if pn <= 0 {
				continue
			}
			controllerIdx := pn - 1 // 0-based for sys.commandLists
			// Loop over all key tables passed
			for ai := 2; ai <= top; ai++ {
				arg := l.Get(ai)
				var btns []string
				switch t := arg.(type) {
				case *lua.LTable:
					t.ForEach(func(_ lua.LValue, v lua.LValue) {
						if s, ok := v.(lua.LString); ok {
							btns = append(btns, string(s))
						}
					})
				case lua.LString:
					btns = []string{string(t)}
				default:
					continue
				}
				if len(btns) > 0 && sys.button(btns, controllerIdx) {
					l.Push(lua.LBool(true))
					return 1
				}
			}
		}
		l.Push(lua.LBool(false))
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
		controllerIdx := -1
		if !nilArg(l, 1) {
			max := input.GetMaxJoystickCount()
			controllerIdx = int(Clamp(int32(numArg(l, 1)), 0, int32(max-1)))
		}

		s, joy := getJoystickKey(controllerIdx)

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
			//sys.keyInput = KeyUnknown
			return 1
		} else if strArg(l, 1) == "" {
			//sys.keyInput = KeyUnknown
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
	luaRegister(l, "getLastInputController", func(l *lua.LState) int {
		if sys.lastInputController >= 0 {
			l.Push(lua.LNumber(sys.lastInputController + 1))
		} else {
			l.Push(lua.LNumber(-1))
		}
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
	luaRegister(l, "getRemapInput", func(l *lua.LState) int {
		pn := int(numArg(l, 1))
		if pn < 1 || pn > len(sys.inputRemap) {
			l.RaiseError("\nInvalid player number: %v\n", pn)
		}
		l.Push(lua.LNumber(sys.inputRemap[pn-1] + 1))
		return 1
	})
	luaRegister(l, "getRoundTime", func(l *lua.LState) int {
		l.Push(lua.LNumber(sys.maxRoundTime))
		return 1
	})
	luaRegister(l, "getRuntimeOS", func(l *lua.LState) int {
		l.Push(lua.LString(runtime.GOOS))
		return 1
	})
	luaRegister(l, "getStageInfo", func(*lua.LState) int {
		c := sys.sel.GetStage(int(numArg(l, 1)))
		tbl := l.NewTable()
		tbl.RawSetString("name", lua.LString(c.name))
		tbl.RawSetString("def", lua.LString(c.def))
		tbl.RawSetString("localcoord", lua.LNumber(c.localcoord[0]))
		tbl.RawSetString("portraitscale", lua.LNumber(c.portraitscale))
		acTable := l.NewTable()
		for _, v := range c.attachedchardef {
			acTable.Append(lua.LString(v))
		}
		tbl.RawSetString("attachedchardef", acTable)
		l.Push(tbl)
		return 1
	})
	luaRegister(l, "getStageNo", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.sel.selectedStageNo))
		return 1
	})
	luaRegister(l, "getStageSelectParams", func(*lua.LState) int {
		c := sys.sel.GetStage(int(numArg(l, 1)))
		lv := toLValue(l, c.ssp)
		lTable, ok := lv.(*lua.LTable)
		if !ok {
			l.RaiseError("Error: 'lv' is not a *lua.LTable")
			return 0
		}
		l.Push(lTable)
		return 1
	})
	luaRegister(l, "getTimestamp", func(*lua.LState) int {
		format := "2006-01-02 15:04:05.000"
		if !nilArg(l, 1) {
			format = strArg(l, 1)
		}
		l.Push(lua.LString(time.Now().Format(format)))
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
	luaRegister(l, "jsonDecode", func(*lua.LState) int {
		path := strArg(l, 1)

		f, err := os.Open(path)
		if err != nil {
			l.RaiseError("jsonDecode: open %s: %v", path, err)
			return 0
		}
		defer f.Close()

		lv, err := jsonToLuaValue(l, f) // uses json.Decoder with UseNumber() inside
		if err != nil {
			l.RaiseError("jsonDecode: parse %s: %v", path, err)
			return 0
		}

		l.Push(lv)
		return 1
	})
	luaRegister(l, "jsonEncode", func(*lua.LState) int {
		lv := l.Get(1)
		path := strArg(l, 2)
		goVal, err := luaToJsonValue(lv, nil)
		if err != nil {
			l.RaiseError(err.Error())
			return 0
		}

		// Stable key order: encoding/json sorts map keys deterministically.
		// Use pretty output; switch to json.Marshal for compact if preferred.
		data, err := json.MarshalIndent(goVal, "", "  ")
		if err != nil {
			l.RaiseError("jsonEncode: %v", err)
			return 0
		}

		// Ensure parent directory exists (if any)
		dir := filepath.Dir(path)
		if dir != "." && dir != "" {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				l.RaiseError("jsonEncode: mkdir %s: %v", dir, err)
				return 0
			}
		}

		if err := os.WriteFile(path, data, 0o644); err != nil {
			l.RaiseError("jsonEncode: write %s: %v", path, err)
			return 0
		}
		return 0
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
		if !nilArg(l, 1) {
			cfg, err := loadConfig(strArg(l, 1))
			if err != nil {
				l.RaiseError("\nCan't load %v: %v\n", strArg(l, 1), err.Error())
			}
			sys.cfg = *cfg
		}
		lv := toLValue(l, sys.cfg)
		l.Push(lv)
		return 1
	})
	luaRegister(l, "loading", func(l *lua.LState) int {
		l.Push(lua.LBool(sys.loader.state == LS_Loading))
		return 1
	})
	luaRegister(l, "loadAnimTable", func(l *lua.LState) int {
		def := ""
		if !nilArg(l, 1) {
			def = strArg(l, 1)
		}
		if def == "" {
			l.RaiseError("loadAnimTable: expected animation file path")
		}
		// Arg2: SFF userdata (optional). If omitted/nil, we create an empty SFF.
		var sff *Sff
		if !nilArg(l, 2) {
			if v, ok := toUserData(l, 2).(*Sff); ok && v != nil {
				sff = v
			} else {
				// Be strict if provided but wrong type.
				userDataError(l, 2, sff)
				return 0
			}
		}
		if sff == nil {
			sff = newSff()
		}
		raw, err := LoadText(def)
		if err != nil {
			l.RaiseError("\nCan't load anim table %v: %v\n", def, err.Error())
			return 0
		}
		lines, i := SplitAndTrim(NormalizeNewlines(raw), "\n"), 0
		at := ReadAnimationTable(sff, &sff.palList, lines, &i)
		// Build Lua table with NUMERIC keys (so it prints like [110] => userdata ...)
		tbl := l.NewTable()
		// Deterministic iteration order
		keys := make([]int, 0, len(at))
		for k := range at {
			keys = append(keys, int(k))
		}
		sort.Ints(keys)
		for _, ki := range keys {
			anim := at[int32(ki)]
			if anim == nil {
				continue
			}
			// toLValue will wrap *Animation into userdata
			tbl.RawSetInt(ki, toLValue(l, anim))
		}
		l.Push(tbl)
		return 1
	})
	luaRegister(l, "loadIni", func(l *lua.LState) int {
		def := ""
		if !nilArg(l, 1) {
			def = strArg(l, 1)
		}
		if def == "" {
			l.RaiseError("loadIniTable: expected ini filename")
		}
		raw, err := LoadText(def)
		if err != nil {
			l.RaiseError("\nCan't load ini %v: %v\n", def, err.Error())
		}
		opts := ini.LoadOptions{
			SkipUnrecognizableLines:   true,
			PreserveSurroundedQuote:   true,
			UnescapeValueDoubleQuotes: false,
		}
		iniFile, err := ini.LoadSources(opts, []byte(NormalizeNewlines(raw)))
		if err != nil {
			l.RaiseError("\nCan't parse ini %v: %v\n", def, err.Error())
		}
		l.Push(iniToLuaTable(l, iniFile))
		return 1
	})
	luaRegister(l, "loadLifebar", func(l *lua.LState) int {
		def := ""
		if !nilArg(l, 1) {
			def = strArg(l, 1)
		}
		lb, err := loadLifebar(def)
		if err != nil {
			l.RaiseError("\nCan't load lifebar %v: %v\n", def, err.Error())
		}
		sys.lifebar = *lb
		return 0
	})
	luaRegister(l, "loadMotif", func(l *lua.LState) int {
		def := ""
		if !nilArg(l, 1) {
			def = strArg(l, 1)
		}
		m, err := loadMotif(def)
		if err != nil {
			l.RaiseError("\nCan't load motif %v: %v\n", def, err.Error())
		}
		sys.motif = *m

		// Initialize the LUT for nokey (helps fix #3091 for all cases)
		StringToButtonLUT[sys.motif.OptionInfo.Menu.Valuename["nokey"]] = 25

		// defaults-only INI (for values baseline)
		defIni := sys.motif.DefaultOnlyIni
		if defIni == nil {
			defIni, _ = ini.Load([]byte(preprocessINIContent(NormalizeNewlines(string(defaultMotif)))))
		}

		getTbl := func(t *lua.LTable, path []string) *lua.LTable {
			cur := t
			for _, k := range path {
				n, _ := cur.RawGetString(k).(*lua.LTable)
				if n == nil {
					return nil
				}
				cur = n
			}
			return cur
		}
		ensureTbl := func(p *lua.LTable, key string) *lua.LTable {
			if t, ok := p.RawGetString(key).(*lua.LTable); ok && t != nil {
				return t
			}
			t := l.NewTable()
			p.RawSetString(key, t)
			return t
		}
		isEmpty := func(v string) bool {
			v = strings.TrimSpace(v)
			v = strings.Trim(v, "\"")
			return v == ""
		}
		isSpacerKey := func(name string) bool {
			n := strings.ToLower(name)
			if !strings.HasPrefix(n, "spacer") {
				return false
			}
			if len(n) == len("spacer") {
				return true
			}
			for _, r := range n[len("spacer"):] {
				if r < '0' || r > '9' {
					return false
				}
			}
			return true
		}
		eachIniKeys := func(file *ini.File, sec, pref string, fn func(rest, val string)) {
			if file == nil || sec == "" {
				return
			}
			// Choose the language-appropriate section name transparently.
			lp := strings.ToLower(pref)
			// Overlay semantics: [Section] then [<lang>.Section]
			for _, secName := range ResolveLangSectionNames(file, sec, SelectedLanguage()) {
				s, err := file.GetSection(secName)
				if err != nil || s == nil {
					continue
				}
				for _, k := range s.Keys() {
					n := k.Name()
					if strings.HasPrefix(strings.ToLower(n), lp) {
						fn(n[len(pref):], k.Value())
					}
				}
			}
		}
		hasIniPrefix := func(file *ini.File, sec, pref string) bool {
			if file == nil || sec == "" {
				return false
			}
			lp := strings.ToLower(pref)
			for _, secName := range ResolveLangSectionNames(file, sec, SelectedLanguage()) {
				s, err := file.GetSection(secName)
				if err != nil || s == nil {
					continue
				}
				for _, k := range s.Keys() {
					if strings.HasPrefix(strings.ToLower(k.Name()), lp) {
						return true
					}
				}
			}
			return false
		}
		customPauseMenuSections := func(file *ini.File) []string {
			if file == nil {
				return nil
			}
			seen := map[string]bool{}
			var out []string
			for _, s := range file.Sections() {
				raw := s.Name()
				if raw == ini.DEFAULT_SECTION {
					continue
				}
				_, base, _ := splitLangPrefix(raw)
				base = strings.TrimSpace(base)
				if base == "" {
					continue
				}
				lower := strings.ToLower(base)
				if !strings.HasSuffix(lower, " pause menu") {
					continue
				}
				// root; all other [* Pause Menu] are derived
				if lower == "pause menu" || seen[lower] {
					continue
				}
				seen[lower] = true
				out = append(out, base)
			}
			return out
		}

		// values: defaults baseline + non-empty user overlays
		populateItemName := func(sec string, path []string, pref, mode string, root *lua.LTable) {
			dst := getTbl(root, path)
			if dst == nil {
				return
			}
			items, _ := dst.RawGetString("itemname").(*lua.LTable)
			if items == nil {
				return
			}

			apply := func(rest, v string, onlyIfMissing bool) {
				val := strings.TrimSpace(v)
				if mode == "flat" {
					k := strings.ReplaceAll(rest, ".", "_")
					// Empty-valued keys disable/remove the entry (override defaults).
					if isEmpty(val) {
						items.RawSetString(k, lua.LNil)
						return
					}
					if onlyIfMissing && items.RawGetString(k) != lua.LNil {
						return
					}
					items.RawSetString(k, lua.LString(val))
					return
				}
				parts := strings.Split(rest, ".")
				elem, field := "default", parts[0]
				if len(parts) >= 2 {
					elem, field = parts[0], parts[1]
				}
				// Empty disables/removes the field from the existing elem table (override defaults).
				if isEmpty(val) {
					if t, ok := items.RawGetString(elem).(*lua.LTable); ok && t != nil {
						t.RawSetString(field, lua.LNil)
					}
					return
				}
				t := ensureTbl(items, elem)
				if onlyIfMissing && t.RawGetString(field) != lua.LNil {
					return
				}
				t.RawSetString(field, lua.LString(val))
			}

			// baseline from defaults
			eachIniKeys(defIni, sec, pref, func(r, v string) { apply(r, v, false) })
			// user overlays (empty disables/removes)
			eachIniKeys(sys.motif.UserIniFile, sec, pref, func(r, v string) { apply(r, v, false) })
		}

		// order for flat itemname.*
		buildFlatOrder := func(sec string, path []string, pref string, root *lua.LTable) {
			dst := getTbl(root, path)
			if dst == nil {
				return
			}
			type entry struct {
				flat, base string
				disabled   bool
			}
			collect := func(file *ini.File, sec string) (out []entry) {
				if file == nil || sec == "" {
					return
				}
				lp := strings.ToLower(pref)
				// Build effective entries per overlay semantics:
				// [Section] then [<lang>.Section] overwriting values/disabled flags.
				byFlat := map[string]entry{}
				var order []string

				addFrom := func(s *ini.Section) {
					if s == nil {
						return
					}
					for _, k := range s.Keys() {
						n := k.Name()
						if !strings.HasPrefix(strings.ToLower(n), lp) {
							continue
						}
						p := n[len(pref):]
						base := p
						if i := strings.LastIndex(base, "."); i >= 0 {
							base = base[i+1:]
						}
						e := entry{
							flat:     strings.ReplaceAll(p, ".", "_"),
							base:     strings.ToLower(base),
							disabled: isEmpty(k.Value()),
						}
						if _, ok := byFlat[e.flat]; !ok {
							order = append(order, e.flat)
						}
						byFlat[e.flat] = e
					}
				}

				for _, secName := range ResolveLangSectionNames(file, sec, SelectedLanguage()) {
					s, err := file.GetSection(secName)
					if err != nil || s == nil {
						continue
					}
					addFrom(s)
				}

				for _, flat := range order {
					out = append(out, byFlat[flat])
				}
				return
			}

			user := collect(sys.motif.UserIniFile, sec)
			defs := collect(sys.motif.IniFile, sec)

			seenFlat := map[string]bool{}     // exact items we've added (e.g. "menuversus_back")
			seenBase := map[string]bool{}     // leafs seen among added items (e.g. "arcade", "survival")
			disabledFlat := map[string]bool{} // exact items disabled
			disabledBase := map[string]bool{} // leaf/suffix disabled via any user empty assignment

			// Pre-seed disabledFlat with user-disabled keys so descendant suppression works
			// even if children appear before the disabled parent key in INI ordering.
			enabledBase := map[string]bool{}
			userDisabledBase := map[string]bool{}
			for _, e := range user {
				if e.disabled {
					disabledFlat[e.flat] = true
					userDisabledBase[e.base] = true
				} else {
					enabledBase[e.base] = true
				}
			}
			for b := range userDisabledBase {
				// Only treat the leaf as globally disabled if the non-empty value for the same leaf is not defined somewhere else.
				if !enabledBase[b] {
					disabledBase[b] = true
				}
			}

			// If a parent/submenu key is disabled, its descendants must be suppressed too.
			isUnderDisabled := func(flat string) bool {
				for p := range disabledFlat {
					if strings.HasPrefix(flat, p+"_") {
						return true
					}
				}
				return false
			}

			var final []string
			process := func(arr []entry, isDefault bool) {
				for _, e := range arr {
					if e.disabled {
						disabledFlat[e.flat] = true
						continue
					}
					// If the user disabled this leaf via any suffix match, suppress default entries for it.
					if isDefault && disabledBase[e.base] {
						continue
					}
					// If this key is disabled (directly) or is a descendant of a disabled submenu, skip it.
					if disabledFlat[e.flat] || isUnderDisabled(e.flat) {
						continue
					}
					// Never include the same *exact* key twice (including spacers).
					if seenFlat[e.flat] {
						continue
					}
					// For defaults, suppress items whose leaf was already added from user entries (or earlier), but do not suppress back or spacer*
					if isDefault && e.base != "back" && !isSpacerKey(e.base) && seenBase[e.base] {
						continue
					}
					seenFlat[e.flat] = true
					seenBase[e.base] = true
					final = append(final, e.flat)
				}
			}

			process(user, false)
			process(defs, true)

			t := l.NewTable()
			for _, v := range final {
				t.Append(lua.LString(v))
			}
			dst.RawSetString("itemname_order", t)
		}

		// order for teammenu.itemname.*
		buildTeamOrder := func(primary string, path []string, root *lua.LTable) {
			dst := getTbl(root, path)
			if dst == nil {
				return
			}
			// We will build a map: element ("default", "teamarcade", etc.) -> ordered list of fields.
			items, _ := dst.RawGetString("itemname").(*lua.LTable)
			if items == nil {
				return
			}

			type entry struct {
				elem     string // "default" or gamemode (e.g. "teamarcade")
				field    string // single/simul/turns/tag/ratio
				disabled bool
			}
			collect := func(file *ini.File, sec string) (out []entry) {
				if file == nil || sec == "" {
					return
				}
				const pref = "teammenu.itemname."
				lp := strings.ToLower(pref)
				// Overlay semantics: [Section] then [<lang>.Section]
				byKey := map[string]entry{} // key = elem|field
				var order []string

				addFrom := func(s *ini.Section) {
					if s == nil {
						return
					}
					for _, k := range s.Keys() {
						n := k.Name()
						ln := strings.ToLower(n)
						if !strings.HasPrefix(ln, lp) {
							continue
						}
						rest := n[len(pref):]
						ps := strings.Split(rest, ".")
						elem := "default"
						field := ""
						if len(ps) == 1 {
							field = strings.ToLower(ps[0])
						} else {
							elem = strings.ToLower(ps[0])
							field = strings.ToLower(ps[len(ps)-1])
						}
						e := entry{elem: elem, field: field, disabled: isEmpty(k.Value())}
						key := elem + "|" + field
						if _, ok := byKey[key]; !ok {
							order = append(order, key)
						}
						byKey[key] = e
					}
				}

				for _, secName := range ResolveLangSectionNames(file, sec, SelectedLanguage()) {
					s, err := file.GetSection(secName)
					if err != nil || s == nil {
						continue
					}
					addFrom(s)
				}
				for _, k := range order {
					out = append(out, byKey[k])
				}
				return
			}

			user := collect(sys.motif.UserIniFile, primary)
			defs := collect(sys.motif.IniFile, primary)

			groupByElem := func(arr []entry) map[string][]entry {
				m := map[string][]entry{}
				for _, e := range arr {
					m[e.elem] = append(m[e.elem], e)
				}
				return m
			}
			uBy := groupByElem(user)
			dBy := groupByElem(defs)

			computeOrder := func(elem string) []string {
				seen := map[string]bool{}
				disabled := map[string]bool{}
				var final []string
				process := func(a []entry) {
					for _, e := range a {
						if e.disabled {
							disabled[e.field], seen[e.field] = true, true
							continue
						}
						if disabled[e.field] || seen[e.field] {
							continue
						}
						seen[e.field] = true
						final = append(final, e.field)
					}
				}
				process(uBy[strings.ToLower(elem)])
				process(dBy[strings.ToLower(elem)])
				return final
			}

			// Default order (used as fallback for elements without explicit order)
			defaultOrder := computeOrder("default")

			orders := l.NewTable()
			// Build an order table for every element present under itemname
			items.ForEach(func(k lua.LValue, _ lua.LValue) {
				elemName := k.String() // keep original key as-is
				elemKeyLower := strings.ToLower(elemName)
				order := computeOrder(elemKeyLower) // try specific gamemode
				if len(order) == 0 {
					order = defaultOrder // fallback to default order if none defined
				}
				list := l.NewTable()
				for _, f := range order {
					list.Append(lua.LString(f))
				}
				orders.RawSetString(elemName, list)
			})

			dst.RawSetString("itemname_order", orders)
		}

		lv := toLValue(l, sys.motif)
		lTable, ok := lv.(*lua.LTable)
		if !ok {
			l.RaiseError("Error: 'lv' is not a *lua.LTable")
			return 0
		}

		mi := []struct {
			sec  string
			path []string
		}{
			{"Title Info", []string{"title_info", "menu"}},
			{"Attract Mode", []string{"attract_mode", "menu"}},
			{"Option Info", []string{"option_info", "menu"}},
			{"Replay Info", []string{"replay_info", "menu"}},
			{"Pause Menu", []string{"pause_menu", "pause_menu", "menu"}},
		}
		for _, s := range mi {
			populateItemName(s.sec, s.path, "menu.itemname.", "flat", lTable)
			buildFlatOrder(s.sec, s.path, "menu.itemname.", lTable)
		}
		// Handle every derived [* Pause Menu]
		for _, sec := range customPauseMenuSections(sys.motif.IniFile) {
			key := normalizeSectionName(sec)
			path := []string{"pause_menu", key, "menu"}
			// Normally defaults seeding ensures derived sections have a baseline.
			// Keep a conservative fallback to [Pause Menu] if itemname is truly absent.
			src := sec
			if !hasIniPrefix(defIni, sec, "menu.itemname.") && !hasIniPrefix(sys.motif.UserIniFile, sec, "menu.itemname.") {
				src = "Pause Menu"
			}
			populateItemName(src, path, "menu.itemname.", "flat", lTable)
			buildFlatOrder(src, path, "menu.itemname.", lTable)
		}
		populateItemName("Option Info", []string{"option_info", "keymenu"}, "keymenu.itemname.", "flat", lTable)
		buildFlatOrder("Option Info", []string{"option_info", "keymenu"}, "keymenu.itemname.", lTable)
		populateItemName("Select Info", []string{"select_info", "teammenu"}, "teammenu.itemname.", "team", lTable)
		buildTeamOrder("Select Info", []string{"select_info", "teammenu"}, lTable)

		l.Push(lTable)
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
		// Always reset per-launch params; they must not leak across matches/modes.
		if sys.sel.gameParams == nil {
			sys.sel.gameParams = newGameParams()
		} else {
			sys.sel.gameParams.Reset()
		}
		sys.sel.music = make(Music)
		if !nilArg(l, 1) {
			entries := SplitAndTrim(strArg(l, 1), ",")
			sys.sel.gameParams.AppendParams(entries)
			// Feed normalized music params to Music.
			sys.sel.music.AppendParams(sys.sel.gameParams.MusicEntries())
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
	luaRegister(l, "loadStoryboard", func(l *lua.LState) int {
		if strArg(l, 1) == "" {
			return 0
		}
		s, err := loadStoryboard(strArg(l, 1))
		if err != nil {
			fmt.Printf("Warning: %v\n", err.Error())
			return 0
		}
		sys.storyboard = *s
		lv := toLValue(l, s)
		l.Push(lv)
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
	luaRegister(l, "modifyMotif", func(l *lua.LState) int {
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
		err := sys.motif.SetValueUpdate(query, value)
		if err == nil {
			return 0
		}
		l.RaiseError("\nmodifyMotif: %v\n", err.Error())
		return 0
	})
	luaRegister(l, "modifyStoryboard", func(l *lua.LState) int {
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
		err := sys.storyboard.SetValueUpdate(query, value)
		if err == nil {
			return 0
		}
		l.RaiseError("\nmodifyStoryboard: %v\n", err.Error())
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
	luaRegister(l, "panicError", func(*lua.LState) int {
		l.RaiseError(strArg(l, 1))
		return 0
	})
	luaRegister(l, "playBgm", func(l *lua.LState) int {
		t := tableArg(l, 1)
		// Defaults
		var (
			bgm               string
			loop              int     = 1
			loopcount         int     = -1
			volume            int     = 100
			loopstart         int     = 0
			loopend           int     = 0
			startposition     int     = 0
			freqmul           float32 = 1.0
			hasNewBGM         bool
			volExplicit       bool
			interrupt         bool
			interruptExplicit bool
		)
		// If source exists, seed values from it first (explicit fields may override below).
		if tableHasKey(t, "source") {
			if v, ok := t.RawGetString("source").(lua.LString); ok {
				srcStr := string(v)
				parts := strings.SplitN(srcStr, ".", 2)
				key := ""
				if len(parts) > 1 {
					key = parts[1]
				}
				switch strings.ToLower(parts[0]) {
				case "stagedef":
					if sys.stage != nil && sys.stage.music != nil {
						bgm, loop, volume, loopstart, loopend, startposition, freqmul, loopcount = sys.stage.music.Read(key, sys.stage.def)
						hasNewBGM = bgm != ""
					}
				case "stageparams":
					if sys.stage != nil && sys.stage.music != nil {
						bgm, loop, volume, loopstart, loopend, startposition, freqmul, loopcount = sys.stage.si().music.Read(key, sys.stage.def)
						hasNewBGM = bgm != ""
					}
				case "launchparams":
					if sys.sel.music != nil && sys.stage != nil {
						bgm, loop, volume, loopstart, loopend, startposition, freqmul, loopcount = sys.sel.music.Read(key, sys.stage.def)
						hasNewBGM = bgm != ""
					}
				case "motif":
					if sys.motif.Music != nil {
						bgm, loop, volume, loopstart, loopend, startposition, freqmul, loopcount = sys.motif.Music.Read(key, sys.motif.Def)
						hasNewBGM = bgm != ""
					}
				case "match":
					if sys.debugWC != nil && sys.debugWC.gi().music != nil {
						bgm, loop, volume, loopstart, loopend, startposition, freqmul, loopcount = sys.debugWC.gi().music.Read(key, sys.stage.def)
						hasNewBGM = bgm != ""
					}
				case "charparams":
					if sys.debugWC != nil && sys.debugWC.si().music != nil {
						bgm, loop, volume, loopstart, loopend, startposition, freqmul, loopcount = sys.debugWC.si().music.Read(key, sys.debugWC.gi().def)
						hasNewBGM = bgm != ""
					}
				default:
					l.RaiseError("\nInvalid source origin: %v\n", parts[0])
				}
			} else {
				l.RaiseError("\nInvalid source value type: %v\n", fmt.Sprintf("%T", t.RawGetString("source")))
			}
		}
		// Explicit fields override source/defaults. Unknown keys error.
		t.ForEach(func(key, value lua.LValue) {
			k, ok := key.(lua.LString)
			if !ok {
				l.RaiseError("\nInvalid table key type: %v\n", fmt.Sprintf("%T", key))
			}
			switch strings.ToLower(string(k)) {
			case "source":
				// already handled
			case "bgm":
				if s, ok := value.(lua.LString); ok {
					bgm = string(s)
					if bgm != "" {
						bgm = SearchFile(bgm, []string{sys.motif.Def, "", "sound/"})
						hasNewBGM = true
					}
				} else {
					l.RaiseError("\nInvalid value for bgm: %v\n", fmt.Sprintf("%T", value))
				}
			case "loop":
				if n, ok := value.(lua.LNumber); ok {
					loop = int(n)
				} else {
					l.RaiseError("\nInvalid value for loop: %v\n", fmt.Sprintf("%T", value))
				}
			case "volume":
				if n, ok := value.(lua.LNumber); ok {
					volume, volExplicit = int(n), true
				} else {
					l.RaiseError("\nInvalid value for volume: %v\n", fmt.Sprintf("%T", value))
				}
			case "loopstart":
				if n, ok := value.(lua.LNumber); ok {
					loopstart = int(n)
				} else {
					l.RaiseError("\nInvalid value for loopstart: %v\n", fmt.Sprintf("%T", value))
				}
			case "loopend":
				if n, ok := value.(lua.LNumber); ok {
					loopend = int(n)
				} else {
					l.RaiseError("\nInvalid value for loopend: %v\n", fmt.Sprintf("%T", value))
				}
			case "startposition":
				if n, ok := value.(lua.LNumber); ok {
					startposition = int(n)
				} else {
					l.RaiseError("\nInvalid value for startposition: %v\n", fmt.Sprintf("%T", value))
				}
			case "freqmul":
				if n, ok := value.(lua.LNumber); ok {
					freqmul = float32(n)
				} else {
					l.RaiseError("\nInvalid value for freqmul: %v\n", fmt.Sprintf("%T", value))
				}
			case "loopcount":
				if n, ok := value.(lua.LNumber); ok {
					loopcount = int(n)
				} else {
					l.RaiseError("\nInvalid value for loopcount: %v\n", fmt.Sprintf("%T", value))
				}
			case "interrupt":
				interrupt = lua.LVAsBool(value)
				interruptExplicit = true
			default:
				l.RaiseError("\nInvalid table key: %v\n", k)
			}
		})
		// Compare against currently playing filename.
		same := false
		if bgm != "" && sys.bgm.filename != "" {
			same = (filepath.Clean(strings.ToLower(bgm)) == filepath.Clean(strings.ToLower(sys.bgm.filename)))
		}
		// If an interrupt was explicitly requested, it must stop the currently playing music even when the requested BGM is undefined/empty.
		// The only time music should persist is when the "new" BGM resolves to the same file as the currently playing one.
		if interruptExplicit && interrupt && !same {
			if hasNewBGM && bgm != "" {
				sys.bgm.Open(bgm, loop, volume, loopstart, loopend, startposition, freqmul, loopcount)
			} else {
				sys.bgm.Stop()
			}
			return 0
		}
		// If the requested BGM is the same as the current one, never restart it.
		if same {
			if volExplicit {
				sys.bgm.bgmVolume = int(Min(int32(volume), int32(sys.cfg.Sound.MaxBGMVolume)))
				sys.bgm.UpdateVolume()
			}
			return 0
		}
		// Otherwise, apply if a new BGM is set. If interrupt wasn't explicitly provided, default to interrupting when switching to a different track.
		if hasNewBGM && bgm != "" {
			if !interruptExplicit {
				interrupt = true
			}
			if interrupt {
				sys.bgm.Open(bgm, loop, volume, loopstart, loopend, startposition, freqmul, loopcount)
			} else if volExplicit {
				sys.bgm.bgmVolume = int(Min(int32(volume), int32(sys.cfg.Sound.MaxBGMVolume)))
				sys.bgm.UpdateVolume()
			}
		} else if volExplicit {
			// No new BGM, volume-only update only
			sys.bgm.bgmVolume = int(Min(int32(volume), int32(sys.cfg.Sound.MaxBGMVolume)))
			sys.bgm.UpdateVolume()
		}
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
			sys.sel.charSpritePreload[[...]uint16{uint16(numArg(l, 1)), uint16(numArg(l, 2))}] = true
		} else {
			sys.sel.charAnimPreload[int32(numArg(l, 1))] = true
		}
		return 0
	})
	luaRegister(l, "preloadListStage", func(*lua.LState) int {
		if !nilArg(l, 2) {
			sys.sel.stageSpritePreload[[...]uint16{uint16(numArg(l, 1)), uint16(numArg(l, 2))}] = true
		} else {
			sys.sel.stageAnimPreload[int32(numArg(l, 1))] = true
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
	luaRegister(l, "readGameStats", func(*lua.LState) int {
		l.Push(toLValue(l, sys.statsLog))
		return 1
	})
	luaRegister(l, "refresh", func(*lua.LState) int {
		sys.tickSound()
		if !sys.frameSkip {
			sys.luaFlushDrawQueue()
			if sys.motif.fadeIn.isActive() {
				sys.motif.fadeIn.step()
				sys.motif.fadeIn.draw()
			} else if sys.motif.fadeOut.isActive() {
				sys.motif.fadeOut.step()
				sys.motif.fadeOut.draw()
			}
		} else {
			// On skipped frames, discard queued draws to avoid buildup.
			sys.luaDiscardDrawQueue()
		}
		sys.stepCommandLists()
		if !sys.update() {
			l.RaiseError("<game end>")
		}
		return 0
	})
	luaRegister(l, "rectDebug", func(*lua.LState) int {
		r, ok := toUserData(l, 1).(*Rect)
		if !ok {
			userDataError(l, 1, r)
		}
		str := ""
		if !nilArg(l, 2) {
			str = strArg(l, 2)
		}
		fmt.Printf("%s *Rect=%p %+v\n", str, r, *r)
		return 0
	})
	luaRegister(l, "rectDraw", func(*lua.LState) int {
		r, ok := toUserData(l, 1).(*Rect)
		if !ok {
			userDataError(l, 1, r)
		}
		layer := r.layerno
		if !nilArg(l, 2) {
			layer = int16(numArg(l, 2))
		}
		//r.Draw(layer)
		rSnap := *r
		layerLocal := layer
		sys.luaQueueLayerDraw(int(layerLocal), func() {
			(&rSnap).Draw(layerLocal)
		})
		return 0
	})
	luaRegister(l, "rectNew", func(*lua.LState) int {
		rect := NewRect()
		l.Push(newUserData(l, rect))
		return 1
	})
	luaRegister(l, "rectReset", func(*lua.LState) int {
		r, ok := toUserData(l, 1).(*Rect)
		if !ok {
			userDataError(l, 1, r)
		}
		r.Reset()
		return 0
	})
	luaRegister(l, "rectSetColor", func(*lua.LState) int {
		r, ok := toUserData(l, 1).(*Rect)
		if !ok {
			userDataError(l, 1, r)
		}
		r.SetColor([3]int32{int32(numArg(l, 2)), int32(numArg(l, 3)), int32(numArg(l, 4))})
		return 0
	})
	luaRegister(l, "rectSetAlpha", func(*lua.LState) int {
		r, ok := toUserData(l, 1).(*Rect)
		if !ok {
			userDataError(l, 1, r)
		}
		r.SetAlpha([2]int32{int32(numArg(l, 2)), int32(numArg(l, 3))})
		return 0
	})
	luaRegister(l, "rectSetAlphaPulse", func(*lua.LState) int {
		r, ok := toUserData(l, 1).(*Rect)
		if !ok {
			userDataError(l, 1, r)
		}
		r.SetAlphaPulse(int32(numArg(l, 2)), int32(numArg(l, 3)), int32(numArg(l, 4)))
		return 0
	})
	luaRegister(l, "rectSetLayerno", func(*lua.LState) int {
		r, ok := toUserData(l, 1).(*Rect)
		if !ok {
			userDataError(l, 1, r)
		}
		r.layerno = int16(numArg(l, 2))
		return 0
	})
	luaRegister(l, "rectSetLocalcoord", func(*lua.LState) int {
		r, ok := toUserData(l, 1).(*Rect)
		if !ok {
			userDataError(l, 1, r)
		}
		r.SetLocalcoord(float32(numArg(l, 2)), float32(numArg(l, 3)))
		return 0
	})
	luaRegister(l, "rectSetWindow", func(*lua.LState) int {
		r, ok := toUserData(l, 1).(*Rect)
		if !ok {
			userDataError(l, 1, r)
		}
		r.SetWindow([4]float32{float32(numArg(l, 2)), float32(numArg(l, 3)), float32(numArg(l, 4)), float32(numArg(l, 5))})
		return 0
	})
	luaRegister(l, "rectUpdate", func(*lua.LState) int {
		r, ok := toUserData(l, 1).(*Rect)
		if !ok {
			userDataError(l, 1, r)
		}
		r.Update()
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
	luaRegister(l, "resetCommandInputSource", func(l *lua.LState) int {
		sys.resetCommandInputSource()
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
		sys.resetMatchData(boolArg(l, 1))
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
	luaRegister(l, "resetGameStats", func(*lua.LState) int {
		sys.statsLog.reset()
		sys.continueFlg = false
		return 0
	})
	luaRegister(l, "roundReset", func(*lua.LState) int {
		sys.roundResetFlg = true
		sys.roundResetMatchStart = true
		return 0
	})
	luaRegister(l, "runStoryboard", func(*lua.LState) int {
		if !sys.paused || sys.frameStepFlag {
			if sys.storyboard.IniFile != nil && !sys.storyboard.initialized {
				sys.storyboard.init()
			}
			if sys.storyboard.active {
				sys.storyboard.step()
			}
		}
		if sys.storyboard.active && !sys.frameSkip {
			sys.storyboard.draw(0)
			sys.storyboard.draw(1)
			sys.storyboard.draw(2)
		}
		l.Push(lua.LBool(sys.storyboard.active))
		return 1
	})
	luaRegister(l, "getStoryboardScene", func(l *lua.LState) int {
		if sys.storyboard.active {
			l.Push(lua.LNumber(sys.storyboard.currentSceneIndex))
		} else {
			l.Push(lua.LNil)
		}
		return 1
	})
	luaRegister(l, "runHiscore", func(*lua.LState) int {
		if !sys.paused || sys.frameStepFlag {
			if !sys.motif.hi.initialized {
				var mode string
				var place, endtime int32
				var nofade, nobgs, nooverlay bool
				if !nilArg(l, 1) {
					mode = strArg(l, 1)
				}
				if !nilArg(l, 2) {
					place = int32(numArg(l, 2))
				}
				if !nilArg(l, 3) {
					endtime = int32(numArg(l, 3))
				}
				if !nilArg(l, 4) {
					nofade = boolArg(l, 4)
				}
				if !nilArg(l, 5) {
					nobgs = boolArg(l, 5)
				}
				if !nilArg(l, 6) {
					nooverlay = boolArg(l, 6)
				}
				sys.motif.hi.init(&sys.motif, mode, place, endtime, nofade, nobgs, nooverlay)
			}
			if sys.motif.hi.active {
				sys.motif.hi.step(&sys.motif)
			}
		}
		if sys.motif.hi.active && !sys.frameSkip {
			sys.motif.hi.draw(&sys.motif, 0)
			sys.motif.hi.draw(&sys.motif, 1)
			sys.motif.hi.draw(&sys.motif, 2)
		}
		l.Push(lua.LBool(sys.motif.hi.active))
		return 1
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
		if pl < 1 || pl > sys.cfg.Config.PaletteMax {
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
			isActPal := false
			if l.GetTop() >= 2 {
				isActPal = boolArg(l, 2)
			}
			sff, err := loadSff(strArg(l, 1), false, true, isActPal)
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
	luaRegister(l, "sffCacheDelete", func(l *lua.LState) int {
		removeSFFCache(strArg(l, 1))
		return 0
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
	luaRegister(l, "setCommandInputSource", func(l *lua.LState) int {
		pn, src := int(numArg(l, 1)), int(numArg(l, 2))
		if pn < 1 || pn > len(sys.commandInputSource) ||
			src < 1 || src > len(sys.commandInputSource) {
			l.RaiseError("\nInvalid player number: %v, %v\n", pn, src)
		}
		sys.commandInputSource[pn-1] = src - 1
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
	luaRegister(l, "setCredits", func(*lua.LState) int {
		sys.credits = int32(numArg(l, 1))
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
					sys.keyConfig[pn-1].bA = btn
				case 6:
					sys.keyConfig[pn-1].bB = btn
				case 7:
					sys.keyConfig[pn-1].bC = btn
				case 8:
					sys.keyConfig[pn-1].bX = btn
				case 9:
					sys.keyConfig[pn-1].bY = btn
				case 10:
					sys.keyConfig[pn-1].bZ = btn
				case 11:
					sys.keyConfig[pn-1].bS = btn
				case 12:
					sys.keyConfig[pn-1].bD = btn
				case 13:
					sys.keyConfig[pn-1].bW = btn
				case 14:
					sys.keyConfig[pn-1].bM = btn
				}
			} else {
				btn := StringToButtonLUT[lua.LVAsString(value)]
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
					sys.joystickConfig[pn-1].bA = btn
				case 6:
					sys.joystickConfig[pn-1].bB = btn
				case 7:
					sys.joystickConfig[pn-1].bC = btn
				case 8:
					sys.joystickConfig[pn-1].bX = btn
				case 9:
					sys.joystickConfig[pn-1].bY = btn
				case 10:
					sys.joystickConfig[pn-1].bZ = btn
				case 11:
					sys.joystickConfig[pn-1].bS = btn
				case 12:
					sys.joystickConfig[pn-1].bD = btn
				case 13:
					sys.joystickConfig[pn-1].bW = btn
				case 14:
					sys.joystickConfig[pn-1].bM = btn
				}
			}
		})
		return 0
	})
	luaRegister(l, "setLastInputController", func(l *lua.LState) int {
		// Lua-facing controller indices are 1-based
		n := int(numArg(l, 1))
		if n >= 1 {
			sys.lastInputController = n - 1
		} else {
			sys.lastInputController = -1
		}
		return 0
	})
	luaRegister(l, "setLife", func(*lua.LState) int {
		if sys.debugWC.alive() {
			sys.debugWC.lifeSet(int32(numArg(l, 1)))
		}
		return 0
	})
	luaRegister(l, "setLifebarElements", func(*lua.LState) int {
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
				case "hidebars": // enabled depending on [Dialogue Info] motif settings
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
	luaRegister(l, "setMotifElements", func(*lua.LState) int {
		tableArg(l, 1).ForEach(func(key, value lua.LValue) {
			switch k := key.(type) {
			case lua.LString:
				switch strings.ToLower(string(k)) {
				case "challenger":
					sys.motif.ch.enabled = lua.LVAsBool(value)
				case "continuescreen":
					sys.motif.co.enabled = lua.LVAsBool(value)
				case "demo":
					sys.motif.de.enabled = lua.LVAsBool(value)
				case "dialogue":
					sys.motif.di.enabled = lua.LVAsBool(value)
				case "hiscore":
					sys.motif.hi.enabled = lua.LVAsBool(value)
				case "losescreen":
					sys.motif.wi.loseEnabled = lua.LVAsBool(value)
				case "versusscreen":
				case "versusmatchno":
				case "victoryscreen":
					sys.motif.vi.enabled = lua.LVAsBool(value)
				case "winscreen":
					sys.motif.wi.winEnabled = lua.LVAsBool(value)
				case "menu":
					sys.motif.me.enabled = lua.LVAsBool(value)
				default:
					l.RaiseError("\nInvalid table key: %v\n", k)
				}
			default:
				l.RaiseError("\nInvalid table key type: %v\n", fmt.Sprintf("%T\n", key))
			}
		})
		return 0
	})
	luaRegister(l, "setPlayers", func(l *lua.LState) int {
		total := int(numArg(l, 1))

		if total < len(sys.commandLists) {
			sys.commandLists = sys.commandLists[:total]
		}

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

		sys.resetCommandInputSource()
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
		sys.sel.selected[tn-1] = nil
		//sys.sel.ocd[tn-1] = nil
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
	luaRegister(l, "stopBgm", func(l *lua.LState) int {
		sys.bgm.Stop()
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
	luaRegister(l, "textImgAddPos", func(*lua.LState) int {
		ts, ok := toUserData(l, 1).(*TextSprite)
		if !ok {
			userDataError(l, 1, ts)
		}
		ts.AddPos(float32(numArg(l, 2)), float32(numArg(l, 3)))
		return 0
	})
	luaRegister(l, "textImgAddText", func(*lua.LState) int {
		ts, ok := toUserData(l, 1).(*TextSprite)
		if !ok {
			userDataError(l, 1, ts)
		}
		ts.text = fmt.Sprintf(ts.text+"%v", strArg(l, 2))
		return 0
	})
	luaRegister(l, "textImgApplyVel", func(*lua.LState) int {
		ts, ok := toUserData(l, 1).(*TextSprite)
		if !ok {
			userDataError(l, 1, ts)
		}
		src, ok := toUserData(l, 2).(*TextSprite)
		if !ok {
			userDataError(l, 2, src)
		}
		ts.velocityInit = src.velocityInit
		ts.xvel, ts.yvel = src.xvel, src.yvel
		ts.vel = src.vel
		return 0
	})
	luaRegister(l, "textImgDebug", func(*lua.LState) int {
		ts, ok := toUserData(l, 1).(*TextSprite)
		if !ok {
			userDataError(l, 1, ts)
		}
		str := ""
		if !nilArg(l, 2) {
			str = strArg(l, 2)
		}
		fmt.Printf("%s *TextSprite=%p %+v\n", str, ts, *ts)
		return 0
	})
	luaRegister(l, "textImgDraw", func(*lua.LState) int {
		ts, ok := toUserData(l, 1).(*TextSprite)
		if !ok {
			userDataError(l, 1, ts)
		}
		layer := ts.layerno
		if !nilArg(l, 2) {
			layer = int16(numArg(l, 2))
		}
		//tsSnap := *ts
		tsSnap := ts.Copy()
		layerLocal := layer
		sys.luaQueueLayerDraw(int(layerLocal), func() {
			//(&tsSnap).Draw(layerLocal)
			tsSnap.Draw(layerLocal)
		})
		return 0
	})
	luaRegister(l, "textImgGetTextWidth", func(*lua.LState) int {
		ts, ok := toUserData(l, 1).(*TextSprite)
		if !ok {
			userDataError(l, 1, ts)
		}
		l.Push(lua.LNumber(ts.TextWidth(strArg(l, 2))))
		return 1
	})
	luaRegister(l, "textImgNew", func(*lua.LState) int {
		l.Push(newUserData(l, NewTextSprite()))
		return 1
	})
	luaRegister(l, "textImgReset", func(*lua.LState) int {
		ts, ok := toUserData(l, 1).(*TextSprite)
		if !ok {
			userDataError(l, 1, ts)
		}
		// If no table provided reset everything.
		if nilArg(l, 2) {
			ts.Reset()
			return 0
		}
		// Apply actions as we encounter valid keys.
		tableArg(l, 2).ForEach(func(key, value lua.LValue) {
			switch v := value.(type) {
			case lua.LString:
				k := strings.ToLower(string(v))
				switch k {
				case "pos":
					ts.SetPos(ts.offsetInit[0], ts.offsetInit[1])
				case "scale":
					ts.SetScale(ts.scaleInit[0], ts.scaleInit[1])
				case "window":
					ts.SetWindow(ts.windowInit)
				case "velocity":
					ts.SetVelocity(ts.velocityInit[0], ts.velocityInit[1])
				case "text":
					ts.text = ts.textInit
				case "palfx":
					if ts.palfx != nil {
						ts.palfx.clear()
					}
				case "delay":
					ts.elapsedTicks = 0
				default:
					l.RaiseError("\nInvalid table key: %v\n", k)
				}
			default:
				l.RaiseError("\nInvalid table key type: %v\n", fmt.Sprintf("%T\n", value))
			}
		})
		return 0
	})
	luaRegister(l, "textImgSetAccel", func(*lua.LState) int {
		ts, ok := toUserData(l, 1).(*TextSprite)
		if !ok {
			userDataError(l, 1, ts)
		}
		ts.SetAccel(float32(numArg(l, 2)), float32(numArg(l, 3)))
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
		// Default alpha to 255 for compatibility
		a := int32(255)
		if !nilArg(l, 5) {
			a = int32(MinI(255, int(numArg(l, 5))))
		}
		ts.SetColor(int32(numArg(l, 2)), int32(numArg(l, 3)), int32(numArg(l, 4)), a)
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
		x, y := ts.offsetInit[0], ts.offsetInit[1]
		if !nilArg(l, 2) {
			x = float32(numArg(l, 2))
		}
		if !nilArg(l, 3) {
			y = float32(numArg(l, 3))
		}
		ts.SetPos(x, y)
		return 0
	})
	luaRegister(l, "textImgSetFriction", func(*lua.LState) int {
		ts, ok := toUserData(l, 1).(*TextSprite)
		if !ok {
			userDataError(l, 1, ts)
		}
		ts.friction[0] = float32(numArg(l, 2))
		ts.friction[1] = float32(numArg(l, 3))
		return 0
	})
	luaRegister(l, "textImgSetLayerno", func(*lua.LState) int {
		ts, ok := toUserData(l, 1).(*TextSprite)
		if !ok {
			userDataError(l, 1, ts)
		}
		ts.layerno = int16(numArg(l, 2))
		return 0
	})
	luaRegister(l, "textImgSetLocalcoord", func(*lua.LState) int {
		ts, ok := toUserData(l, 1).(*TextSprite)
		if !ok {
			userDataError(l, 1, ts)
		}
		ts.SetLocalcoord(float32(numArg(l, 2)), float32(numArg(l, 3)))
		return 0
	})
	luaRegister(l, "textImgSetMaxDist", func(*lua.LState) int {
		ts, ok := toUserData(l, 1).(*TextSprite)
		if !ok {
			userDataError(l, 1, ts)
		}
		ts.SetMaxDist(float32(numArg(l, 2)), float32(numArg(l, 3)))
		return 0
	})
	luaRegister(l, "textImgSetScale", func(*lua.LState) int {
		ts, ok := toUserData(l, 1).(*TextSprite)
		if !ok {
			userDataError(l, 1, ts)
		}
		ts.SetScale(float32(numArg(l, 2)), float32(numArg(l, 3)))
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
	luaRegister(l, "textImgSetTextDelay", func(*lua.LState) int {
		ts, ok := toUserData(l, 1).(*TextSprite)
		if !ok {
			userDataError(l, 1, ts)
		}
		ts.textDelay = float32(numArg(l, 2))
		return 0
	})
	luaRegister(l, "textImgSetTextSpacing", func(*lua.LState) int {
		ts, ok := toUserData(l, 1).(*TextSprite)
		if !ok {
			userDataError(l, 1, ts)
		}
		ts.SetTextSpacing(float32(numArg(l, 2)), float32(numArg(l, 3)))
		return 0
	})
	luaRegister(l, "textImgSetTextWrap", func(*lua.LState) int {
		ts, ok := toUserData(l, 1).(*TextSprite)
		if !ok {
			userDataError(l, 1, ts)
		}
		ts.textWrap = boolArg(l, 2)
		return 0
	})
	luaRegister(l, "textImgSetVelocity", func(*lua.LState) int {
		ts, ok := toUserData(l, 1).(*TextSprite)
		if !ok {
			userDataError(l, 1, ts)
		}
		ts.SetVelocity(float32(numArg(l, 2)), float32(numArg(l, 3)))
		return 0
	})
	luaRegister(l, "textImgSetWindow", func(*lua.LState) int {
		ts, ok := toUserData(l, 1).(*TextSprite)
		if !ok {
			userDataError(l, 1, ts)
		}
		ts.SetWindow([4]float32{float32(numArg(l, 2)), float32(numArg(l, 3)), float32(numArg(l, 4)), float32(numArg(l, 5))})
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
		ts.rot.angle = float32(numArg(l, 2))
		return 0
	})
	luaRegister(l, "textImgSetXAngle", func(*lua.LState) int {
		ts, ok := toUserData(l, 1).(*TextSprite)
		if !ok {
			userDataError(l, 1, ts)
		}
		ts.rot.xangle = float32(numArg(l, 2))
		return 0
	})
	luaRegister(l, "textImgSetYAngle", func(*lua.LState) int {
		ts, ok := toUserData(l, 1).(*TextSprite)
		if !ok {
			userDataError(l, 1, ts)
		}
		ts.rot.yangle = float32(numArg(l, 2))
		return 0
	})
	luaRegister(l, "textImgSetProjection", func(*lua.LState) int {
		ts, ok := toUserData(l, 1).(*TextSprite)
		if !ok {
			userDataError(l, 1, ts)
		}
		switch l.Get(2).Type() {

		case lua.LTNumber:
			ts.projection = int32(numArg(l, 2))

		case lua.LTString:
			switch strings.ToLower(strings.TrimSpace(l.Get(2).String())) {
			case "orthographic":
				ts.projection = int32(Projection_Orthographic)
			case "perspective":
				ts.projection = int32(Projection_Perspective)
			case "perspective2":
				ts.projection = int32(Projection_Perspective2)
			}
		}
		return 0
	})
	luaRegister(l, "textImgSetFocalLength", func(*lua.LState) int {
		ts, ok := toUserData(l, 1).(*TextSprite)
		if !ok {
			userDataError(l, 1, ts)
		}
		ts.fLength = float32(numArg(l, 2))
		return 0
	})
	luaRegister(l, "textImgUpdate", func(*lua.LState) int {
		ts, ok := toUserData(l, 1).(*TextSprite)
		if !ok {
			userDataError(l, 1, ts)
		}
		ts.Update()
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
	luaRegister(l, "toggleDebugDisplay", func(*lua.LState) int {
		if !sys.debugModeAllowed() {
			return 0
		}
		// Shift+D behavior: just toggle without cycling players
		if !nilArg(l, 1) {
			sys.debugDisplay = !sys.debugDisplay
			return 0
		}
		// Make a copy of runOrder
		sorted := make([]*Char, len(sys.charList.runOrder))
		copy(sorted, sys.charList.runOrder)
		// Sort the copy by player number and ID
		sort.Slice(sorted, func(i, j int) bool {
			if sorted[i].playerNo != sorted[j].playerNo {
				return sorted[i].playerNo < sorted[j].playerNo
			}
			return sorted[i].id < sorted[j].id
		})
		// Find reference char
		var nextChar *Char
		if sys.debugDisplay {
			// Search for the first character that comes after the current one
			for _, c := range sorted {
				isLaterPlayer := c.playerNo > sys.debugRef[0]
				isSamePlayerNewerID := c.playerNo == sys.debugRef[0] && c.id > sys.debugLastID
				if isLaterPlayer || isSamePlayerNewerID {
					nextChar = c
					break
				}
			}
		} else if len(sorted) > 0 {
			// If display was off, start at the beginning of the sorted list
			nextChar = sorted[0]
		}
		// Update debug reference or disable debug
		if nextChar != nil {
			sys.debugRef[0] = nextChar.playerNo
			sys.debugRef[1] = int(nextChar.helperIndex)
			sys.debugLastID = nextChar.id
			sys.debugDisplay = true
		} else {
			// If no "next" character exists in the remainder of the list, reset and close
			sys.debugRef[0] = 0
			sys.debugRef[1] = 0
			sys.debugLastID = -1
			sys.debugDisplay = false
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
	luaRegister(l, "version", func(l *lua.LState) int {
		ver := fmt.Sprintf("%s - %s", Version, BuildTime)
		l.Push(lua.LString(ver))
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
		// Script version doesn't log errors because debug mode uses it
		// TODO: Script redirects should either all log errors or none of them should
		if c := sys.debugWC.playerIDTrigger(int32(numArg(l, 1)), false); c != nil {
			sys.debugWC, ret = c, true
		}
		l.Push(lua.LBool(ret))
		return 1
	})
	luaRegister(l, "playerindex", func(*lua.LState) int {
		ret := false
		if c := sys.debugWC.playerIndexTrigger(int32(numArg(l, 1))); c != nil {
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
		ret, idx := false, int32(0)
		if !nilArg(l, 1) {
			idx = int32(numArg(l, 1))
		}
		if c := sys.debugWC.helperIndexTrigger(idx); c != nil {
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
	luaRegister(l, "attackmul", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.attackMul[0]))
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
		arg1 := strArg(l, 1)
		id := int32(numArg(l, 2))
		arg3 := strArg(l, 3)
		var c1, c2 int32
		// Get box 1 type
		switch strings.ToLower(arg1) {
		case "clsn1":
			c1 = 1
		case "clsn2":
			c1 = 2
		case "size":
			c1 = 3
		default:
			l.RaiseError("Invalid collision box type: %v", arg1)
		}
		// Get box 2 type
		switch strings.ToLower(arg3) {
		case "clsn1":
			c2 = 1
		case "clsn2":
			c2 = 2
		case "size":
			c2 = 3
		default:
			l.RaiseError("Invalid collision box type: %v", arg3)
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
			case "clsn1":
				clsn := sys.debugWC.getClsn(1)
				if clsn != nil && idx >= 0 && idx < len(clsn) {
					v = lua.LNumber(clsn[idx][offset])
				}
			case "clsn2":
				clsn := sys.debugWC.getClsn(2)
				if clsn != nil && idx >= 0 && idx < len(clsn) {
					v = lua.LNumber(clsn[idx][offset])
				}
			case "size":
				clsn := sys.debugWC.getClsn(3)
				if clsn != nil && len(clsn) > 0 {
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
		l.Push(lua.LNumber(sys.debugWC.finalDefense * 100))
		return 1
	})
	luaRegister(l, "defencemul", func(*lua.LState) int {
		l.Push(lua.LNumber(float32(sys.debugWC.finalDefense / float64(sys.debugWC.gi().defenceBase) * 100)))
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
		var lv lua.LValue
		id := int32(numArg(l, 1))
		idx := int(numArg(l, 2))
		vname := strArg(l, 3)
		// Get explod
		e := sys.debugWC.getSingleExplod(id, idx, true)
		// Handle returns
		if e != nil {
			switch strings.ToLower(vname) {
			case "accel x":
				lv = lua.LNumber(e.accel[0])
			case "accel y":
				lv = lua.LNumber(e.accel[1])
			case "accel z":
				lv = lua.LNumber(e.accel[2])
			case "anim":
				lv = lua.LNumber(e.animNo)
			case "angle":
				lv = lua.LNumber(e.anglerot[0] + e.interpolate_angle[0])
			case "angle x":
				lv = lua.LNumber(e.anglerot[1] + e.interpolate_angle[1])
			case "angle y":
				lv = lua.LNumber(e.anglerot[2] + e.interpolate_angle[2])
			case "animelem":
				lv = lua.LNumber(e.anim.curelem + 1)
			case "animelemtime":
				lv = lua.LNumber(e.anim.curelemtime)
			case "animplayerno":
				lv = lua.LNumber(e.animPN + 1)
			case "animtime":
				lv = lua.LNumber(e.anim.AnimTime())
			case "spriteplayerno":
				lv = lua.LNumber(e.spritePN + 1)
			case "bindtime":
				lv = lua.LNumber(e.bindtime)
			case "drawpal group":
				lv = lua.LNumber(sys.debugWC.explodDrawPal(e)[0])
			case "drawpal index":
				lv = lua.LNumber(sys.debugWC.explodDrawPal(e)[1])
			case "facing":
				lv = lua.LNumber(e.trueFacing())
			case "friction x":
				lv = lua.LNumber(e.friction[0])
			case "friction y":
				lv = lua.LNumber(e.friction[1])
			case "friction z":
				lv = lua.LNumber(e.friction[2])
			case "id":
				lv = lua.LNumber(e.id)
			case "layerno":
				lv = lua.LNumber(e.layerno)
			case "pausemovetime":
				lv = lua.LNumber(e.pausemovetime)
			case "pos x":
				lv = lua.LNumber(e.pos[0] + e.offset[0] + e.relativePos[0] + e.interpolate_pos[0])
			case "pos y":
				lv = lua.LNumber(e.pos[1] + e.offset[1] + e.relativePos[1] + e.interpolate_pos[1])
			case "pos z":
				lv = lua.LNumber(e.pos[2] + e.offset[2] + e.relativePos[2] + e.interpolate_pos[2])
			case "removetime":
				lv = lua.LNumber(e.removetime)
			case "scale x":
				lv = lua.LNumber(e.scale[0] * e.interpolate_scale[0])
			case "scale y":
				lv = lua.LNumber(e.scale[1] * e.interpolate_scale[1])
			case "sprpriority":
				lv = lua.LNumber(e.sprpriority)
			case "time":
				lv = lua.LNumber(e.time)
			case "vel x":
				lv = lua.LNumber(e.velocity[0])
			case "vel y":
				lv = lua.LNumber(e.velocity[1])
			case "vel z":
				lv = lua.LNumber(e.velocity[2])
			case "xshear":
				lv = lua.LNumber(e.xshear)
			default:
				l.RaiseError("\nInvalid argument: %v\n", vname)
			}
		}
		l.Push(lv)
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
		l.Push(lua.LNumber(sys.gameTime()))
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
		var lv lua.LValue
		switch strings.ToLower(strArg(l, 1)) {
		case "fall.envshake.dir":
			lv = lua.LNumber(c.ghv.fall_envshake_dir)
		case "animtype":
			lv = lua.LNumber(c.gethitAnimtype())
		case "air.animtype":
			lv = lua.LNumber(c.ghv.airanimtype)
		case "ground.animtype":
			lv = lua.LNumber(c.ghv.groundanimtype)
		case "fall.animtype":
			lv = lua.LNumber(c.ghv.fall_animtype)
		case "type":
			lv = lua.LNumber(c.ghv._type)
		case "airtype":
			lv = lua.LNumber(c.ghv.airtype)
		case "groundtype":
			lv = lua.LNumber(c.ghv.groundtype)
		case "damage":
			lv = lua.LNumber(c.ghv.damage)
		case "guardcount":
			lv = lua.LNumber(c.ghv.guardcount)
		case "hitcount":
			lv = lua.LNumber(c.ghv.hitcount)
		case "fallcount":
			lv = lua.LNumber(c.ghv.fallcount)
		case "hitshaketime":
			lv = lua.LNumber(c.ghv.hitshaketime)
		case "hittime":
			lv = lua.LNumber(c.ghv.hittime)
		case "stand.friction":
			sf := c.ghv.standfriction
			if math.IsNaN(float64(sf)) {
				sf = c.gi().movement.stand.friction
			}
			lv = lua.LNumber(sf)
		case "crouch.friction":
			cf := c.ghv.crouchfriction
			if math.IsNaN(float64(cf)) {
				cf = c.gi().movement.crouch.friction
			}
			lv = lua.LNumber(cf)
		case "slidetime":
			lv = lua.LNumber(c.ghv.slidetime)
		case "ctrltime":
			lv = lua.LNumber(c.ghv.ctrltime)
		case "recovertime", "down.recovertime": // Added second term for consistency
			lv = lua.LNumber(c.ghv.down_recovertime)
		case "xoff":
			lv = lua.LNumber(c.ghv.xoff)
		case "yoff":
			lv = lua.LNumber(c.ghv.yoff)
		case "zoff":
			lv = lua.LNumber(c.ghv.zoff)
		case "xvel":
			lv = lua.LNumber(c.ghv.xvel)
		case "yvel":
			lv = lua.LNumber(c.ghv.yvel)
		case "zvel":
			lv = lua.LNumber(c.ghv.zvel)
		case "xaccel":
			lv = lua.LNumber(c.ghv.xaccel)
		case "yaccel":
			lv = lua.LNumber(c.ghv.yaccel)
		case "zaccel":
			lv = lua.LNumber(c.ghv.zaccel)
		case "xveladd":
			lv = lua.LNumber(c.ghv.xveladd)
		case "yveladd":
			lv = lua.LNumber(c.ghv.yveladd)
		case "hitid", "chainid":
			lv = lua.LNumber(c.ghv.chainId())
		case "guarded":
			lv = lua.LBool(c.ghv.guarded)
		case "isbound":
			lv = lua.LBool(c.isTargetBound())
		case "fall":
			lv = lua.LBool(c.ghv.fallflag)
		case "fall.damage":
			lv = lua.LNumber(c.ghv.fall_damage)
		case "fall.xvel":
			if math.IsNaN(float64(c.ghv.fall_xvelocity)) {
				lv = lua.LNumber(-32760) // Winmugen behavior
			} else {
				lv = lua.LNumber(c.ghv.fall_xvelocity)
			}
		case "fall.yvel":
			lv = lua.LNumber(c.ghv.fall_yvelocity)
		case "fall.zvel":
			if math.IsNaN(float64(c.ghv.fall_zvelocity)) {
				lv = lua.LNumber(-32760) // Winmugen behavior
			} else {
				lv = lua.LNumber(c.ghv.fall_zvelocity)
			}
		case "fall.recover":
			lv = lua.LBool(c.ghv.fall_recover)
		case "fall.time":
			lv = lua.LNumber(c.fallTime)
		case "fall.recovertime":
			lv = lua.LNumber(c.ghv.fall_recovertime)
		case "fall.kill":
			lv = lua.LBool(c.ghv.fall_kill)
		case "fall.envshake.time":
			lv = lua.LNumber(c.ghv.fall_envshake_time)
		case "fall.envshake.freq":
			lv = lua.LNumber(c.ghv.fall_envshake_freq)
		case "fall.envshake.ampl":
			lv = lua.LNumber(c.ghv.fall_envshake_ampl)
		case "fall.envshake.phase":
			lv = lua.LNumber(c.ghv.fall_envshake_phase)
		case "fall.envshake.mul":
			lv = lua.LNumber(c.ghv.fall_envshake_mul)
		case "attr":
			lv = attrLStr(c.ghv.attr)
		case "dizzypoints":
			lv = lua.LNumber(c.ghv.dizzypoints)
		case "guardpoints":
			lv = lua.LNumber(c.ghv.guardpoints)
		case "playerid":
			lv = lua.LNumber(c.ghv.playerid)
		case "playerno":
			lv = lua.LNumber(c.ghv.playerno + 1)
		case "redlife":
			lv = lua.LNumber(c.ghv.redlife)
		case "score":
			lv = lua.LNumber(c.ghv.score)
		case "hitdamage":
			lv = lua.LNumber(c.ghv.hitdamage)
		case "guarddamage":
			lv = lua.LNumber(c.ghv.guarddamage)
		case "power":
			lv = lua.LNumber(c.ghv.power)
		case "hitpower":
			lv = lua.LNumber(c.ghv.hitpower)
		case "guardpower":
			lv = lua.LNumber(c.ghv.guardpower)
		case "kill":
			lv = lua.LBool(c.ghv.kill)
		case "priority":
			lv = lua.LNumber(c.ghv.priority)
		case "facing":
			lv = lua.LNumber(c.ghv.facing)
		case "ground.velocity.x":
			lv = lua.LNumber(c.ghv.ground_velocity[0])
		case "ground.velocity.y":
			lv = lua.LNumber(c.ghv.ground_velocity[1])
		case "ground.velocity.z":
			lv = lua.LNumber(c.ghv.ground_velocity[2])
		case "air.velocity.x":
			lv = lua.LNumber(c.ghv.air_velocity[0])
		case "air.velocity.y":
			lv = lua.LNumber(c.ghv.air_velocity[1])
		case "air.velocity.z":
			lv = lua.LNumber(c.ghv.air_velocity[2])
		case "down.velocity.x":
			lv = lua.LNumber(c.ghv.down_velocity[0])
		case "down.velocity.y":
			lv = lua.LNumber(c.ghv.down_velocity[1])
		case "down.velocity.z":
			lv = lua.LNumber(c.ghv.down_velocity[2])
		case "guard.velocity.x":
			lv = lua.LNumber(c.ghv.guard_velocity[0])
		case "guard.velocity.y":
			lv = lua.LNumber(c.ghv.guard_velocity[1])
		case "guard.velocity.z":
			lv = lua.LNumber(c.ghv.guard_velocity[2])
		case "airguard.velocity.x":
			lv = lua.LNumber(c.ghv.airguard_velocity[0])
		case "airguard.velocity.y":
			lv = lua.LNumber(c.ghv.airguard_velocity[1])
		case "airguard.velocity.z":
			lv = lua.LNumber(c.ghv.airguard_velocity[2])
		case "frame":
			lv = lua.LBool(c.ghv.frame)
		case "down.recover":
			lv = lua.LBool(c.ghv.down_recover)
		case "guardflag":
			lv = flagLStr(c.ghv.guardflag)
		case "keepstate":
			lv = lua.LBool(c.ghv.keepstate)
		case "projid":
			lv = lua.LNumber(c.ghv.projid)
		case "guardko":
			lv = lua.LBool(c.ghv.guardko)
		default:
			l.RaiseError("\nInvalid argument: %v\n", strArg(l, 1))
		}
		l.Push(lv)
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
		if sys.debugWC.ss.moveType == MT_A {
			l.Push(attrLStr(sys.debugWC.hitdef.attr))
		} else {
			l.Push(lua.LString(""))
		}
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
		var lv lua.LValue
		switch strings.ToLower(strArg(l, 1)) {
		case "cornerpush.veloff":
			lv = lua.LNumber(c.mhv.cornerpush_veloff)
		case "frame":
			lv = lua.LBool(c.mhv.frame)
		case "overridden":
			lv = lua.LBool(c.mhv.overridden)
		case "playerid":
			lv = lua.LNumber(c.mhv.playerid)
		case "playerno":
			lv = lua.LNumber(c.mhv.playerno + 1)
		case "sparkx":
			lv = lua.LNumber(c.mhv.sparkxy[0])
		case "sparky":
			lv = lua.LNumber(c.mhv.sparkxy[1])
		case "uniqhit":
			lv = lua.LNumber(len(c.hitdefTargets))
		default:
			l.RaiseError("\nInvalid argument: %v\n", strArg(l, 1))
		}
		l.Push(lv)
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
		idx := int(numArg(l, 1))
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
		// Get projectile
		p := sys.debugWC.getSingleProj(id, idx, true)
		// Handle returns
		if p != nil {
			switch vname {
			case "accel x":
				lv = lua.LNumber(p.accel[0])
			case "accel y":
				lv = lua.LNumber(p.accel[1])
			case "accel z":
				lv = lua.LNumber(p.accel[2])
			case "anim":
				lv = lua.LNumber(p.animNo)
			case "animelem":
				lv = lua.LNumber(p.anim.curelem + 1)
			case "angle":
				lv = lua.LNumber(p.anglerot[0])
			case "angle x":
				lv = lua.LNumber(p.anglerot[1])
			case "angle y":
				lv = lua.LNumber(p.anglerot[2])
			case "attr":
				lv = attrLStr(p.hitdef.attr) // Return string like HitDefAttr
			case "drawpal group":
				lv = lua.LNumber(sys.debugWC.projDrawPal(p)[0])
			case "drawpal index":
				lv = lua.LNumber(sys.debugWC.projDrawPal(p)[1])
			case "facing":
				lv = lua.LNumber(p.facing)
			case "guardflag":
				lv = flagLStr(p.hitdef.guardflag) // Return string like HitDefVar
			case "highbound":
				lv = lua.LNumber(p.heightbound[1])
			case "hitflag":
				lv = flagLStr(p.hitdef.hitflag) // Return string like HitDefVar
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
			case "shadow r":
				lv = lua.LNumber(p.shadow[0])
			case "shadow g":
				lv = lua.LNumber(p.shadow[1])
			case "shadow b":
				lv = lua.LNumber(p.shadow[2])
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
			default:
				l.RaiseError("\nInvalid argument: %v\n", vname)
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
		l.Push(lua.LNumber(sys.debugWC.runOrderTrigger()))
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
		bg := sys.debugWC.getSingleStageBg(id, idx, true)
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
		case "stageinfo.autoturn":
			l.Push(lua.LBool(sys.stage.autoturn))
		case "stageinfo.resetbg":
			l.Push(lua.LBool(sys.stage.resetbg))
		case "stageinfo.xscale":
			l.Push(lua.LNumber(sys.stage.scale[0]))
		case "stageinfo.yscale":
			l.Push(lua.LNumber(sys.stage.scale[1]))
		case "stageinfo.zoffset":
			l.Push(lua.LNumber(sys.stage.stageCamera.zoffset))
		case "stageinfo.zoffsetlink":
			l.Push(lua.LNumber(sys.stage.zoffsetlink))
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
		case "shadow.ydelta":
			l.Push(lua.LNumber(sys.stage.sdw.ydelta))
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
		case "reflection.ydelta":
			l.Push(lua.LNumber(sys.stage.reflection.ydelta))
		case "reflection.fade.range.begin":
			l.Push(lua.LNumber(sys.stage.reflection.fadebgn))
		case "reflection.fade.range.end":
			l.Push(lua.LNumber(sys.stage.reflection.fadeend))
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
	luaRegister(l, "winclutch", func(*lua.LState) int {
		l.Push(lua.LBool(sys.debugWC.winClutch()))
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
		var lv lua.LValue
		// Because the char's animation steps at the end of each frame, before the scripts run,
		// AnimElemVar Lua version uses curFrame instead of anim.CurrentFrame()
		f := sys.debugWC.curFrame
		if f != nil {
			switch vname {
			case "alphadest":
				lv = lua.LNumber(f.DstAlpha)
			case "alphasource":
				lv = lua.LNumber(f.SrcAlpha)
			case "angle":
				lv = lua.LNumber(f.Angle)
			case "group":
				lv = lua.LNumber(f.Group)
			case "hflip":
				lv = lua.LBool(f.Hscale < 0)
			case "image":
				lv = lua.LNumber(f.Number)
			case "numclsn1":
				lv = lua.LNumber(len(f.Clsn1))
			case "numclsn2":
				lv = lua.LNumber(len(f.Clsn2))
			case "time":
				lv = lua.LNumber(f.Time)
			case "vflip":
				lv = lua.LBool(f.Vscale < 0)
			case "xoffset":
				lv = lua.LNumber(f.Xoffset)
			case "xscale":
				lv = lua.LNumber(f.Xscale)
			case "yoffset":
				lv = lua.LNumber(f.Yoffset)
			case "yscale":
				lv = lua.LNumber(f.Yscale)
			default:
				l.RaiseError("\nInvalid argument: %v\n", vname)
			}
		}
		l.Push(lv)
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
		l.Push(lua.LBool(sys.decisiveRound[sys.debugWC.playerNo&1]))
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
			l.Push(lua.LNumber(sys.lifebar.localcoord[0]))
		case "info.localcoord.y":
			l.Push(lua.LNumber(sys.lifebar.localcoord[1]))
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
		l.Push(lua.LNumber(sys.matchTime))
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
		case "persistlife":
			l.Push(lua.LBool(sys.sel.gameParams.PersistLife))
		case "persistmusic":
			l.Push(lua.LBool(sys.sel.gameParams.PersistMusic))
		case "persistrounds":
			l.Push(lua.LBool(sys.sel.gameParams.PersistRounds))
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
		l.Push(lua.LBool(sys.debugWC.helperIndexExist(
			BytecodeInt(int32(numArg(l, 1)))).ToB()))
		return 1
	})
	luaRegister(l, "helpervar", func(l *lua.LState) int {
		vname := strArg(l, 1)
		var lv lua.LValue
		c := sys.debugWC
		if c.helperIndex > 0 {
			switch strings.ToLower(vname) {
			case "clsnproxy":
				lv = lua.LBool(c.isclsnproxy)
			case "helpertype":
				lv = lua.LNumber(c.helperType)
			case "id":
				lv = lua.LNumber(c.helperId)
			case "keyctrl":
				lv = lua.LBool(c.keyctrl[0])
			case "ownclsnscale":
				lv = lua.LBool(c.ownclsnscale)
			case "ownpal":
				lv = lua.LBool(c.ownpal)
			case "preserve":
				lv = lua.LBool(c.preserve)
			default:
				l.RaiseError("\nInvalid argument: %v\n", vname)
			}
		}
		l.Push(lv)
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
		case "nocornerpush":
			l.Push(lua.LBool(sys.debugWC.asf(ASF_nocornerpush)))
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
		case "noscore":
			l.Push(lua.LBool(sys.debugWC.asf(ASF_noscore)))
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
		l.Push(lua.LNumber(sys.lastCharId))
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
		case "challenger":
			l.Push(lua.LBool(sys.motif.ch.active))
		case "continuescreen":
			l.Push(lua.LBool(sys.motif.co.active))
		case "continueyes":
			l.Push(lua.LBool(sys.motif.co.active && sys.motif.co.selected && sys.continueFlg))
		case "continueno":
			l.Push(lua.LBool(sys.motif.co.active && sys.motif.co.selected && !sys.continueFlg))
		case "demo":
			l.Push(lua.LBool(sys.motif.de.active))
		case "dialogue":
			l.Push(lua.LBool(sys.motif.di.active))
		case "menu":
			l.Push(lua.LBool(sys.motif.me.active))
		case "victoryscreen":
			l.Push(lua.LBool(sys.motif.vi.active))
		case "winscreen":
			l.Push(lua.LBool(sys.motif.wi.active))
		case "hiscore":
			l.Push(lua.LBool(sys.motif.hi.active))
		default:
			l.RaiseError("\nInvalid argument: %v\n", strArg(l, 1))
		}
		return 1
	})
	luaRegister(l, "motifvar", func(l *lua.LState) int {
		value, err := sys.motif.GetValue(strArg(l, 1))
		if err == nil {
			lv := toLValue(l, value)
			l.Push(lv)
			return 1
		}
		l.RaiseError("\nInvalid argument: %v\n", strArg(l, 1))
		return 0
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
	luaRegister(l, "spritevar", func(l *lua.LState) int {
		vname := strings.ToLower(strArg(l, 1))
		var lv lua.LValue
		// Check for valid sprite
		var spr *Sprite
		if sys.debugWC.anim != nil {
			spr = sys.debugWC.anim.spr
		}
		// Handle output
		if spr != nil {
			switch vname {
			case "group":
				lv = lua.LNumber(spr.Group)
			case "height":
				lv = lua.LNumber(spr.Size[1])
			case "image":
				lv = lua.LNumber(spr.Number)
			case "width":
				lv = lua.LNumber(spr.Size[0])
			case "xoffset":
				lv = lua.LNumber(spr.Offset[0])
			case "yoffset":
				lv = lua.LNumber(spr.Offset[1])
			default:
				l.RaiseError("\nInvalid argument: %v\n", vname)
			}
		}
		l.Push(lv)
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
		l.Push(lua.LNumber(sys.timeElapsed()))
		return 1
	})
	luaRegister(l, "timeremaining", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.timeRemaining()))
		return 1
	})
	luaRegister(l, "timetotal", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.timeTotal()))
		return 1
	})
	luaRegister(l, "zoomvar", func(*lua.LState) int {
		var ln lua.LNumber
		switch strings.ToLower(strArg(l, 1)) {
		case "scale":
			ln = lua.LNumber(sys.drawScale)
		case "pos.x":
			ln = lua.LNumber(sys.zoomPosXLag)
		case "pos.y":
			ln = lua.LNumber(sys.zoomPosYLag)
		case "lag":
			ln = lua.LNumber(sys.zoomlag)
		case "time":
			ln = lua.LNumber(sys.enableZoomtime)
		default:
			l.RaiseError("\nInvalid argument: %v\n", strArg(l, 1))
		}
		l.Push(ln)
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
	luaRegister(l, "credits", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.credits))
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
		l.Push(lua.LNumber(100 * sys.gameLogicSpeed() / 60))
		return 1
	})
	luaRegister(l, "gameRunning", func(l *lua.LState) int {
		l.Push(lua.LBool(sys.gameRunning))
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
		l.Push(lua.LBool(sys.netplay()))
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
		l.Push(lua.LNumber(sys.winnerTeam()))
		return 1
	})
}
