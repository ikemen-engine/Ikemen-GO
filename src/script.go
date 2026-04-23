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
	triggerRedirection(l)
	triggerFunctions(l)
	luaRegister(l, "addChar", func(l *lua.LState) int {
		/*Add a character definition to the select screen.
		@function addChar
		@tparam string defpath Path to the character `.def` file (relative to `chars/` or absolute).
		@tparam[opt] string params Optional comma-separated parameter string (from select.def)
		@treturn boolean success `true` if the character was added successfully, `false` otherwise.
		function addChar(defpath, params) end*/
		if sc := sys.sel.AddChar(strArg(l, 1)); sc != nil {
			if !nilArg(l, 2) {
				entries := SplitAndTrim(StripComment(strArg(l, 2)), ",")
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
		/*Register a global keyboard shortcut that runs Lua code.
		@function addHotkey
		@tparam string key Key name as used by the engine (for example `"F1"`, `"C"`, etc.).
		@tparam[opt=false] boolean ctrl If `true`, the Ctrl key must be held.
		@tparam[opt=false] boolean alt If `true`, the Alt key must be held.
		@tparam[opt=false] boolean shift If `true`, the Shift key must be held.
		@tparam[opt=false] boolean allowDuringPause If `true`, the hotkey also works while the game is paused.
		@tparam[opt=false] boolean debugOnly If `true`, the hotkey is treated as a debug key
		  and only works when debug input is allowed.
		@tparam string script Lua code to execute when the shortcut is pressed.
		@treturn boolean success `true` if the shortcut was registered, `false` if the key name is invalid.
		function addHotkey(key, ctrl, alt, shift, allowDuringPause, debugOnly, script) end*/
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
		/*Add a stage definition to the select screen.
		@function addStage
		@tparam string defpath Path to the stage `.def` file (relative to `stages/` or absolute).
		@tparam[opt] string params Optional comma-separated parameter string (from select.def)
		@treturn boolean success `true` if the stage was added successfully, `false` otherwise.
		function addStage(defpath, params) end*/
		if ss, err := sys.sel.AddStage(strArg(l, 1)); err == nil {
			if !nilArg(l, 2) {
				entries := SplitAndTrim(StripComment(strArg(l, 2)), ",")
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
		/*Add an offset to an animation's current position.
		@function animAddPos
		@tparam Anim anim Animation userdata.
		@tparam float32 dx Offset to add on the X axis.
		@tparam float32 dy Offset to add on the Y axis.
		function animAddPos(anim, dx, dy) end*/
		a, ok := toUserData(l, 1).(*Anim)
		if !ok {
			userDataError(l, 1, a)
		}
		a.AddPos(float32(numArg(l, 2)), float32(numArg(l, 3)))
		return 0
	})
	luaRegister(l, "animApplyVel", func(*lua.LState) int {
		/*Copy velocity parameters from one animation to another.
		@function animApplyVel
		@tparam Anim target Target animation userdata to modify.
		@tparam Anim source Source animation userdata whose velocity/accel settings are copied.
		function animApplyVel(target, source) end*/
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
	luaRegister(l, "animCopy", func(l *lua.LState) int {
		/*Create a copy of an animation.
		@function animCopy
		@tparam Anim anim Animation userdata.
		@treturn Anim|nil copy New `Anim` userdata containing a copy of `anim`,
		  or `nil` if `anim` is `nil`.
		function animCopy(anim) end*/
		a, ok := toUserData(l, 1).(*Anim)
		if !ok {
			userDataError(l, 1, a)
		}
		if a == nil {
			l.Push(lua.LNil)
			return 1
		}
		l.Push(newUserData(l, a.Copy()))
		return 1
	})
	luaRegister(l, "animDebug", func(*lua.LState) int {
		/*Print debug information about an animation to the console.
		@function animDebug
		@tparam Anim anim Animation userdata.
		@tparam[opt] string prefix Optional text prefix printed before the debug info.
		function animDebug(anim, prefix) end*/
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
		/*Queue drawing of an animation on a render layer.
		@function animDraw
		@tparam Anim anim Animation userdata.
		@tparam[opt] int16 layer Render layer index; if omitted, the animation's own `layerno` is used.
		function animDraw(anim, layer) end*/
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
		/*Get timing information for an animation.
		@function animGetLength
		@tparam Anim anim Animation userdata.
		@treturn int32 length Effective animation length in ticks (as returned by `Anim.GetLength()`).
		@treturn int32 totaltime Raw `totaltime` field from the underlying `Animation`.
		function animGetLength(anim) end*/
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
		/*Get a preloaded character animation by sprite group/number.
		@function animGetPreloadedCharData
		@tparam int charRef 0-based character index in the select list.
		@tparam int32 group Sprite group number.
		@tparam int32 number Sprite number.
		@tparam[opt=true] boolean keepLoop If `true`, keep the original loop timing.
		  If `false` and the animation's `totaltime` equals `looptime`, convert it to an
		  infinite loop (`totaltime = -1`, `looptime = 0`).
		@treturn Anim|nil anim A new `Anim` userdata wrapping the preloaded animation,
		  or `nil` if no matching animation exists.
		function animGetPreloadedCharData(charRef, group, number, keepLoop) end*/
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
		/*Get a preloaded stage animation by sprite group/number.
		@function animGetPreloadedStageData
		@tparam int stageRef Stage reference used by the select system.
		  Positive values are 1-based stage slots. `0` is the random-stage sentinel.
		@tparam int32 group Sprite group number.
		@tparam int32 number Sprite number.
		@tparam[opt=true] boolean keepLoop If `true`, keep the original loop timing.
		  If `false` and the animation's `totaltime` equals `looptime`, convert it to an
		  infinite loop (`totaltime = -1`, `looptime = 0`).
		@treturn Anim|nil anim A new `Anim` userdata wrapping the preloaded animation,
		  or `nil` if no matching animation exists.
		function animGetPreloadedStageData(stageRef, group, number, keepLoop) end*/
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
		/*Get information about a sprite used by an animation.
		@function animGetSpriteInfo
		@tparam Anim anim Animation userdata.
		@tparam[opt] uint16 group Explicit sprite group number.
		@tparam[opt] uint16 number Explicit sprite number. These are only used when both
		  `group` and `number` are provided; otherwise the animation's current sprite is used.
		@treturn table|nil info Table with:
		  - `Group` (uint16) sprite group number
		  - `Number` (uint16) sprite number
		  - `Size` (uint16[2]) `{width, height}`
		  - `Offset` (int16[2]) `{x, y}`
		  - `palidx` (int) palette index used for this sprite,
		  or `nil` if no sprite is available.
		function animGetSpriteInfo(anim, group, number) end*/
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
		/*Load palettes for an animation's underlying sprite file, if palette usage is enabled.
		@function animLoadPalettes
		@tparam Anim anim Animation userdata.
		@tparam int param Palette parameter passed to `loadCharPalettes` (engine-specific semantics).
		function animLoadPalettes(anim, param) end*/
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
		/*Create a new animation from a sprite file and action definition.
		@function animNew
		@tparam[opt] Sff sff Sprite file userdata to use. If omitted or invalid,
		  a new empty SFF is created internally.
		@tparam string|Animation actOrAnim Either:
		  - a string definition of the animation to load, or
		  - an `Animation` userdata (for example from `motif.AnimTable[...]`).
		@treturn Anim anim Newly created animation userdata.
		function animNew(sff, actOrAnim) end*/
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
	luaRegister(l, "animPaletteGet", func(*lua.LState) int {
		/*Get a palette from an animation.
		@function animPaletteGet
		@tparam Anim anim Animation userdata.
		@tparam int paletteId 1-based palette index.
		@treturn table palette Array-like table where each entry is `{r, g, b, a}`.
		function animPaletteGet(anim, paletteId) end*/
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
		/*Set colors in an animation palette.
		@function animPaletteSet
		@tparam Anim anim Animation userdata.
		@tparam int paletteId 1-based palette index.
		@tparam table palette Array-like table where each entry is `{r, g, b, a}`.
		function animPaletteSet(anim, paletteId, palette) end*/
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
	luaRegister(l, "animPrepare", func(l *lua.LState) int {
		/*Prepare an animation so that each character can apply its own palette.
		@function animPrepare
		@tparam Anim anim Source animation userdata.
		@tparam int32 charRef 0-based character index in the select list.
		@treturn Anim preparedAnim Either a copy with adjusted palette data (when palette
		  usage is enabled) or the original `anim` when palette handling is disabled.
		function animPrepare(anim, charRef) end*/
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
		/*Reset an animation to its initial state, fully or partially.
		@function animReset
		@tparam Anim anim Animation userdata.
		@tparam[opt] table parts If omitted or `nil`, resets everything.
		  If provided, must be an array-like table of strings, each one of:
		  - `"anim"`: reset the underlying `Animation`
		  - `"pos"`: reset position to the initial offset
		  - `"scale"`: reset scale to the initial values
		  - `"window"`: reset the clipping window to the initial value
		  - `"velocity"`: reset velocity to the initial value
		  - `"palfx"`: clear PalFX
		function animReset(anim, parts) end*/
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
		/*Set gravity/acceleration applied to an animation.
		@function animSetAccel
		@tparam Anim anim Animation userdata.
		@tparam float32 ax Horizontal acceleration.
		@tparam float32 ay Vertical acceleration.
		function animSetAccel(anim, ax, ay) end*/
		a, ok := toUserData(l, 1).(*Anim)
		if !ok {
			userDataError(l, 1, a)
		}
		a.SetAccel(float32(numArg(l, 2)), float32(numArg(l, 3)))
		return 0
	})
	luaRegister(l, "animSetAlpha", func(*lua.LState) int {
		/*Set alpha blending for an animation.
		@function animSetAlpha
		@tparam Anim anim Animation userdata.
		@tparam int16 src Source alpha factor (0–256, engine-specific).
		@tparam int16 dst Destination alpha factor (0–256, engine-specific).
		function animSetAlpha(anim, src, dst) end*/
		a, ok := toUserData(l, 1).(*Anim)
		if !ok {
			userDataError(l, 1, a)
		}
		a.SetAlpha(int16(numArg(l, 2)), int16(numArg(l, 3)))
		return 0
	})
	luaRegister(l, "animSetAngle", func(*lua.LState) int {
		/*Set rotation angle for an animation.
		@function animSetAngle
		@tparam Anim anim Animation userdata.
		@tparam float32 angle Rotation angle.
		function animSetAngle(anim, angle) end*/
		a, ok := toUserData(l, 1).(*Anim)
		if !ok {
			userDataError(l, 1, a)
		}
		a.rot.angle = float32(numArg(l, 2))
		return 0
	})
	luaRegister(l, "animSetAnimation", func(*lua.LState) int {
		/*Replace an animation's underlying `Animation` data.
		@function animSetAnimation
		@tparam Anim anim Animation userdata to modify.
		@tparam string|Animation actOrAnim Either:
		  - a string definition of the animation to load, or
		  - an `Animation` userdata (for example from `motif.AnimTable[...]`).
		function animSetAnimation(anim, actOrAnim) end*/
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
	luaRegister(l, "animSetColorKey", func(*lua.LState) int {
		/*Set the color key (transparent index) used by an animation.
		@function animSetColorKey
		@tparam Anim anim Animation userdata.
		@tparam int16 index Palette index used as the transparency key.
		function animSetColorKey(anim, index) end*/
		a, ok := toUserData(l, 1).(*Anim)
		if !ok {
			userDataError(l, 1, a)
		}
		a.SetColorKey(int16(numArg(l, 2)))
		return 0
	})
	luaRegister(l, "animSetColorPalette", func(*lua.LState) int {
		/*Change the active palette mapping for an animation.
		@function animSetColorPalette
		@tparam Anim anim Animation userdata.
		@tparam int paletteId 1-based palette identifier to map to.
		@treturn Anim anim The same animation userdata (for chaining).
		function animSetColorPalette(anim, paletteId) end*/
		a, _ := toUserData(l, 1).(*Anim)
		if len(a.anim.palettedata.paletteMap) > 0 {
			a.anim.palettedata.paletteMap[0] = int(numArg(l, 2)) - 1
		}
		l.Push(newUserData(l, a))
		return 1
	})
	luaRegister(l, "animSetFacing", func(*lua.LState) int {
		/*Set the facing (horizontal flip) of an animation.
		@function animSetFacing
		@tparam Anim anim Animation userdata.
		@tparam float32 facing Facing multiplier, usually `1` or `-1`.
		function animSetFacing(anim, facing) end*/
		a, ok := toUserData(l, 1).(*Anim)
		if !ok {
			userDataError(l, 1, a)
		}
		a.facing = float32(numArg(l, 2))
		return 0
	})
	luaRegister(l, "animSetFocalLength", func(*lua.LState) int {
		/*Set focal length used for perspective projection on an animation.
		@function animSetFocalLength
		@tparam Anim anim Animation userdata.
		@tparam float32 fLength Focal length value.
		function animSetFocalLength(anim, fLength) end*/
		a, ok := toUserData(l, 1).(*Anim)
		if !ok {
			userDataError(l, 1, a)
		}
		a.fLength = float32(numArg(l, 2))
		return 0
	})
	luaRegister(l, "animSetFriction", func(*lua.LState) int {
		/*Set friction applied to an animation's velocity.
		@function animSetFriction
		@tparam Anim anim Animation userdata.
		@tparam float32 fx Horizontal friction.
		@tparam float32 fy Vertical friction.
		function animSetFriction(anim, fx, fy) end*/
		a, ok := toUserData(l, 1).(*Anim)
		if !ok {
			userDataError(l, 1, a)
		}
		a.friction[0] = float32(numArg(l, 2))
		a.friction[1] = float32(numArg(l, 3))
		return 0
	})
	luaRegister(l, "animSetLayerno", func(*lua.LState) int {
		/*Set the render layer used by an animation.
		@function animSetLayerno
		@tparam Anim anim Animation userdata.
		@tparam int16 layer Layer index to draw this animation on.
		function animSetLayerno(anim, layer) end*/
		a, ok := toUserData(l, 1).(*Anim)
		if !ok {
			userDataError(l, 1, a)
		}
		a.layerno = int16(numArg(l, 2))
		return 0
	})
	luaRegister(l, "animSetLocalcoord", func(*lua.LState) int {
		/*Set the local coordinate system for an animation.
		@function animSetLocalcoord
		@tparam Anim anim Animation userdata.
		@tparam float32 width Local coordinate width.
		@tparam float32 height Local coordinate height.
		function animSetLocalcoord(anim, width, height) end*/
		a, ok := toUserData(l, 1).(*Anim)
		if !ok {
			userDataError(l, 1, a)
		}
		a.SetLocalcoord(float32(numArg(l, 2)), float32(numArg(l, 3)))
		return 0
	})
	luaRegister(l, "animSetMaxDist", func(*lua.LState) int {
		/*Set maximum drawing distance (clipping bounds) for an animation.
		@function animSetMaxDist
		@tparam Anim anim Animation userdata.
		@tparam float32 maxX Maximum horizontal distance.
		@tparam float32 maxY Maximum vertical distance.
		function animSetMaxDist(anim, maxX, maxY) end*/
		a, ok := toUserData(l, 1).(*Anim)
		if !ok {
			userDataError(l, 1, a)
		}
		a.SetMaxDist(float32(numArg(l, 2)), float32(numArg(l, 3)))
		return 0
	})
	luaRegister(l, "animSetPalFX", func(*lua.LState) int {
		/*Configure palette effects for an animation.
		@function animSetPalFX
		@tparam Anim anim Animation userdata.
		@tparam table palfx Table of palette effect fields:
		  - `time` (int32): duration in ticks
		  - `add` (table int32[3]): additive RGB components
		  - `mul` (table int32[3]): multiplicative RGB components
		  - `sinadd` (table int32[4]): sinusoidal add `{r, g, b, period}`; negative period flips sign
		  - `sinmul` (table int32[4]): sinusoidal mul `{r, g, b, period}`; negative period flips sign
		  - `sincolor` (table int32[2]): sinusoidal color shift `{amount, period}`
		  - `sinhue` (table int32[2]): sinusoidal hue shift `{amount, period}`
		  - `invertall` (int32): set to `1` to invert all colors
		  - `invertblend` (int32): invert blend mode index
		  - `color` (float32): color saturation factor (`0–256` scaled to `0.0–1.0`)
		  - `hue` (float32): hue adjustment factor (`0–256` scaled to `0.0–0.5`)
		function animSetPalFX(anim, palfx) end*/
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
					a.palfx.hue = float32(lua.LVAsNumber(value)) / 512
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
		/*Set animation position, optionally overriding only one axis.
		@function animSetPos
		@tparam Anim anim Animation userdata.
		@tparam[opt] float32 x New X position; if omitted, initial X offset is used.
		@tparam[opt] float32 y New Y position; if omitted, initial Y offset is used.
		function animSetPos(anim, x, y) end*/
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
	luaRegister(l, "animSetProjection", func(*lua.LState) int {
		/*Set projection mode for an animation.
		@function animSetProjection
		@tparam Anim anim Animation userdata.
		@tparam int32|string projection Projection mode. Can be a numeric engine constant, or one of:
		  - `"orthographic"`
		  - `"perspective"`
		  - `"perspective2"`
		function animSetProjection(anim, projection) end*/
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
	luaRegister(l, "animSetScale", func(*lua.LState) int {
		/*Set the scale of an animation.
		@function animSetScale
		@tparam Anim anim Animation userdata.
		@tparam float32 sx Horizontal scale factor.
		@tparam float32 sy Vertical scale factor.
		function animSetScale(anim, sx, sy) end*/
		a, ok := toUserData(l, 1).(*Anim)
		if !ok {
			userDataError(l, 1, a)
		}
		a.SetScale(float32(numArg(l, 2)), float32(numArg(l, 3)))
		return 0
	})
	luaRegister(l, "animSetTile", func(*lua.LState) int {
		/*Configure tiling for an animation.
		@function animSetTile
		@tparam Anim anim Animation userdata.
		@tparam boolean tileX If `true`, tile horizontally.
		@tparam boolean tileY If `true`, tile vertically.
		@tparam[opt] int32 spacingX Horizontal tile spacing in pixels.
		@tparam[opt] int32 spacingY Vertical tile spacing in pixels (defaults to `spacingX` if omitted).
		function animSetTile(anim, tileX, tileY, spacingX, spacingY) end*/
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
		/*Set the base velocity of an animation.
		@function animSetVelocity
		@tparam Anim anim Animation userdata.
		@tparam float32 vx Horizontal velocity.
		@tparam float32 vy Vertical velocity.
		function animSetVelocity(anim, vx, vy) end*/
		a, ok := toUserData(l, 1).(*Anim)
		if !ok {
			userDataError(l, 1, a)
		}
		a.SetVelocity(float32(numArg(l, 2)), float32(numArg(l, 3)))
		return 0
	})
	luaRegister(l, "animSetWindow", func(*lua.LState) int {
		/*Set the clipping window for an animation.
		@function animSetWindow
		@tparam Anim anim Animation userdata.
		@tparam float32 x1 Left coordinate of the clipping window.
		@tparam float32 y1 Top coordinate of the clipping window.
		@tparam float32 x2 Right coordinate of the clipping window.
		@tparam float32 y2 Bottom coordinate of the clipping window.
		function animSetWindow(anim, x1, y1, x2, y2) end*/
		a, ok := toUserData(l, 1).(*Anim)
		if !ok {
			userDataError(l, 1, a)
		}
		a.SetWindow([4]float32{float32(numArg(l, 2)), float32(numArg(l, 3)), float32(numArg(l, 4)), float32(numArg(l, 5))})
		return 0
	})
	luaRegister(l, "animSetXAngle", func(*lua.LState) int {
		/*Set rotation angle around the X axis for an animation.
		@function animSetXAngle
		@tparam Anim anim Animation userdata.
		@tparam float32 xangle X-axis rotation angle (engine-specific units, usually degrees).
		function animSetXAngle(anim, xangle) end*/
		a, ok := toUserData(l, 1).(*Anim)
		if !ok {
			userDataError(l, 1, a)
		}
		a.rot.xangle = float32(numArg(l, 2))
		return 0
	})
	luaRegister(l, "animSetXShear", func(*lua.LState) int {
		/*Set the X shear factor applied when drawing an animation.
		@function animSetXShear
		@tparam Anim anim Animation userdata.
		@tparam float32 shear X shear factor.
		function animSetXShear(anim, shear) end*/
		a, ok := toUserData(l, 1).(*Anim)
		if !ok {
			userDataError(l, 1, a)
		}
		a.xshear = float32(numArg(l, 2))
		return 0
	})
	luaRegister(l, "animSetYAngle", func(*lua.LState) int {
		/*Set rotation angle around the Y axis for an animation.
		@function animSetYAngle
		@tparam Anim anim Animation userdata.
		@tparam float32 yangle Y-axis rotation angle.
		function animSetYAngle(anim, yangle) end*/
		a, ok := toUserData(l, 1).(*Anim)
		if !ok {
			userDataError(l, 1, a)
		}
		a.rot.yangle = float32(numArg(l, 2))
		return 0
	})
	luaRegister(l, "animUpdate", func(*lua.LState) int {
		/*Advance an animation by one tick. By default, only the first call per frame advances the animation.
		@function animUpdate
		@tparam Anim anim Animation userdata.
		@tparam[opt=false] boolean force If `true`, advance the animation even if it was
		  already updated this frame.
		function animUpdate(anim, force) end*/
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
		/*Queue drawing of many animations in one call.
		@function batchDraw
		@tparam table batch Array-like table of draw items. Each item is a table:
		  - `anim` (Anim) animation userdata
		  - `x` (float32) X position
		  - `y` (float32) Y position
		  - `facing` (float32) facing multiplier (usually `1` or `-1`)
		  - `scale` (table float32[2], optional) `{sx, sy}` scale override
		  - `xshear` (float32, optional) X shear override
		  - `angle` (float32, optional) rotation angle override
		  - `xangle` (float32, optional) X-axis rotation override
		  - `yangle` (float32, optional) Y-axis rotation override
		  - `projection` (int32|string, optional) projection override; accepts a numeric
		    engine constant or `"orthographic"`, `"perspective"`, `"perspective2"`
		  - `focallength` (float32, optional) focal length override
		  - `layerno` (int16, optional) layer override; defaults to `anim.layerno`
		function batchDraw(batch) end*/
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
		/*Print debug information about a background definition.
		@function bgDebug
		@tparam BGDef bg Background definition userdata.
		@tparam[opt] string prefix Optional text prefix printed before the debug info.
		function bgDebug(bg, prefix) end*/
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
		/*Queue drawing of a background definition.
		@function bgDraw
		@tparam BGDef bg Background definition userdata.
		@tparam[opt=0] int32 layer `0` for back layer, `1` for front layer.
		@tparam[opt=0] float32 x Global X offset applied when drawing.
		@tparam[opt=0] float32 y Global Y offset applied when drawing.
		@tparam[opt=1.0] float32 scale Uniform global scale multiplier.
		function bgDraw(bg, layer, x, y, scale) end*/
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
		/*Load a background definition from a sprite file and configuration.
		@function bgNew
		@tparam Sff sff Sprite file userdata used by the background.
		@tparam string defPath Path used for resolving background resources.
		@tparam string section Name or identifier of the background definition to load.
		@tparam[opt] Model model Stage/model userdata associated with this background.
		@tparam[opt=0] int32 defaultLayer Default layer index assigned to the background elements.
		@treturn BGDef bg Loaded background definition userdata.
		function bgNew(sff, defPath, section, model, defaultLayer) end*/
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
		/*Reset a background definition to its initial state.
		@function bgReset
		@tparam BGDef bg Background definition userdata.
		function bgReset(bg) end*/
		bg, ok := toUserData(l, 1).(*BGDef)
		if !ok {
			userDataError(l, 1, bg)
		}
		bg.Reset()
		return 0
	})
	luaRegister(l, "changeAnim", func(l *lua.LState) int {
		/*[redirectable] Change the character's current animation.
		@function changeAnim
		@tparam int32 animNo Animation number to switch to.
		@tparam[opt] int32 elem Optional animation element index to start from.
		@tparam[opt=false] boolean ffx If `true`, use the `"f"` animation prefix (FFX animation).
		@treturn boolean success `true` if the animation exists and was changed, `false` otherwise.
		function changeAnim(animNo, elem, ffx) end*/
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
	luaRegister(l, "changeState", func(l *lua.LState) int {
		/*[redirectable] Change the character's current state or disable it.
		@function changeState
		@tparam int32 stateNo State number to switch to, or `-1` to disable the character.
		@treturn boolean success `true` if an existing state was entered, `false` otherwise.
		  Passing `-1` disables the character and returns `false`.
		function changeState(stateNo) end*/
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
		/*Clear all characters' clipboard text buffers.
		@function clear
		function clear() end*/
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
		/*Stop all currently playing sounds.
		@function clearAllSound
		function clearAllSound() end*/
		sys.clearAllSound()
		return 0
	})
	luaRegister(l, "clearColor", func(l *lua.LState) int {
		/*Fill the screen with a solid color (with optional alpha).
		@function clearColor
		@tparam int32 r Red component (0–255).
		@tparam int32 g Green component (0–255).
		@tparam int32 b Blue component (0–255).
		@tparam[opt=255] int32 alpha Alpha value (0–255); `255` is fully opaque.
		function clearColor(r, g, b, alpha) end*/
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
			FillRect(sys.scrrect, colLocal, [2]int32{srcLocal, dstLocal}, nil)
		})
		return 0
	})
	luaRegister(l, "clearConsole", func(*lua.LState) int {
		/*Clear text printed to the in-engine console.
		@function clearConsole
		function clearConsole() end*/
		sys.consoleText = nil
		return 0
	})
	luaRegister(l, "clearSelected", func(l *lua.LState) int {
		/*Clear all current select-screen choices (characters, stages, music, game params).
		@function clearSelected
		function clearSelected() end*/
		sys.sel.ClearSelected()
		return 0
	})
	luaRegister(l, "commandAdd", func(l *lua.LState) int {
		/*Register a UI command definition.
		@function commandAdd
		@tparam string name Command name (used in triggers).
		@tparam string command Command string in engine input notation.
		@tparam[opt] int32 time Command input time window in ticks.
		@tparam[opt] int32 bufferTime Buffer time in ticks.
		@tparam[opt] boolean bufferHitpause Whether inputs are buffered during hitpause.
		@tparam[opt] boolean bufferPauseend Whether inputs are buffered during pause end.
		@tparam[opt] int32 stepTime Step granularity in ticks.
		function commandAdd(name, command, time, bufferTime, bufferHitpause, bufferPauseend, stepTime) end*/
		name := strArg(l, 1)
		cmdstr := strArg(l, 2)
		dcl := (*CommandList)(nil)
		for _, cl := range sys.commandLists {
			if cl != nil {
				dcl = cl
				break
			}
		}
		if dcl == nil {
			dcl = NewCommandList(nil)
		}
		time := dcl.DefaultTime
		buftime := dcl.DefaultBufferTime
		bufferHitpause := dcl.DefaultBufferHitpause
		bufferPauseend := dcl.DefaultBufferPauseEnd
		steptime := dcl.DefaultStepTime
		if !nilArg(l, 3) {
			time = int32(numArg(l, 3))
		}
		if !nilArg(l, 4) {
			buftime = Max(1, int32(numArg(l, 4)))
		}
		if !nilArg(l, 5) {
			bufferHitpause = boolArg(l, 5)
		}
		if !nilArg(l, 6) {
			bufferPauseend = boolArg(l, 6)
		}
		if !nilArg(l, 7) {
			steptime = int32(numArg(l, 7))
		}
		spec := CommandSpec{
			Cmd:            cmdstr,
			Time:           time,
			BufTime:        buftime,
			BufferHitpause: bufferHitpause,
			BufferPauseend: bufferPauseend,
			StepTime:       steptime,
		}
		if err := sys.uiRegisterCommand(name, spec); err != nil {
			l.RaiseError(err.Error())
		}
		return 0
	})
	luaRegister(l, "commandBufReset", func(l *lua.LState) int {
		/*Reset command input buffers.
		@function commandBufReset
		@tparam[opt] int playerNo 1-based player/controller index. If omitted, all command buffers are reset.
		function commandBufReset(playerNo) end*/
		if nilArg(l, 1) {
			for _, cl := range sys.commandLists {
				if cl == nil {
					continue
				}
				cl.BufReset()
			}
			return 0
		}
		pn := int(numArg(l, 1))
		if pn >= 1 && pn <= len(sys.commandLists) && sys.commandLists[pn-1] != nil {
			sys.commandLists[pn-1].BufReset()
		}
		return 0
	})
	luaRegister(l, "commandDebug", func(*lua.LState) int {
		/*Print debug information about a command list.
		@function commandDebug
		@tparam int playerNo 1-based player/controller index.
		@tparam[opt] string prefix Optional text prefix printed before the debug info.
		function commandDebug(playerNo, prefix) end*/
		pn := int(numArg(l, 1))
		if pn < 1 || pn > len(sys.commandLists) || sys.commandLists[pn-1] == nil {
			return 0
		}
		cl := sys.commandLists[pn-1]
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
		if buf != nil {
			fmt.Printf("%s *Buffer=%p %+v\n", str, buf, *buf)
		} else {
			fmt.Printf("%s *Buffer=nil\n", str)
		}
		return 0
	})
	luaRegister(l, "commandGetState", func(l *lua.LState) int {
		/*Query the current state of a named command.
		@function commandGetState
		@tparam int playerNo 1-based player/controller index.
		@tparam string name Command name to query.
		@treturn boolean active `true` if the command is currently active, `false` otherwise.
		function commandGetState(playerNo, name) end*/
		pn := int(numArg(l, 1))
		if pn < 1 || pn > len(sys.commandLists) || sys.commandLists[pn-1] == nil {
			l.Push(lua.LBool(false))
			return 1 // Attempt to fix a rare registry overflow error while the window is unfocused
		}
		l.Push(lua.LBool(sys.commandLists[pn-1].GetState(strArg(l, 2))))
		return 1
	})
	luaRegister(l, "computeRanking", func(l *lua.LState) int {
		/*Compute and store ranking data for a mode, returning whether it was cleared.
		@function computeRanking
		@tparam string mode Ranking mode identifier.
		@treturn boolean cleared `true` if the run cleared the mode's requirements.
		@treturn int32 place Ranking position (1-based), or `0` if unranked / skipped / not visible.
		function computeRanking(mode) end*/
		mode := strArg(l, 1)
		cleared, place := computeAndSaveRanking(mode)
		l.Push(lua.LBool(cleared))
		l.Push(lua.LNumber(place))
		return 2
	})
	luaRegister(l, "connected", func(*lua.LState) int {
		/*Check if the main menu network connection is established.
		@function connected
		@treturn boolean connected `true` if connected to a netplay peer, `false` otherwise.
		function connected() end*/
		l.Push(lua.LBool(sys.netConnection != nil && sys.netConnection.IsConnected())) // No need to check rollback here as this deals with the main menu connection
		return 1
	})
	luaRegister(l, "continued", func(*lua.LState) int {
		/*Check whether the current run used a continue.
		@function continued
		@treturn boolean continued `true` if the continue flag is set.
		function continued() end*/
		l.Push(lua.LBool(sys.continueFlg))
		return 1
	})
	luaRegister(l, "endMatch", func(*lua.LState) int {
		/*Signal that the current match should end (using menu fade-out settings).
		@function endMatch
		function endMatch() end*/
		sys.motif.PauseMenu["pause_menu"].FadeOut.FadeData.init(sys.motif.fadeOut, false)
		sys.uiResetTokenGuard()
		sys.endMatch = true
		return 0
	})
	luaRegister(l, "enterNetPlay", func(*lua.LState) int {
		/*Enter netplay as client or host.
		@function enterNetPlay
		@tparam string host Host address (IP or hostname). If an empty string, listen
		  for an incoming connection; otherwise connect to the given host.
		function enterNetPlay(host) end*/
		if sys.netConnection != nil {
			l.RaiseError("\nConnection already established.\n")
		}
		sys.sessionWarning = ""
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
		/*Enter replay playback mode from a replay file.
		@function enterReplay
		@tparam string path Path to the replay file.
		@treturn boolean success `true` if the replay file was opened and playback started,
		  `false` otherwise.
		function enterReplay(path) end*/
		sys.sessionWarning = ""
		if sys.cfg.Video.VSync >= 0 {
			sys.window.SetSwapInterval(1) // broken frame skipping when set to 0
		}
		sys.chars = [len(sys.chars)][]*Char{}
		sys.replayFile = OpenReplayFile(strArg(l, 1))
		if sys.replayFile != nil {
			if err := sys.beginReplaySession(sys.replayFile); err != nil {
				if sys.sessionWarning != "" {
					sys.replayFile.Close()
					sys.replayFile = nil
					l.Push(lua.LBool(false))
					return 1
				}
				sys.replayFile.Close()
				sys.replayFile = nil
				l.RaiseError(err.Error())
			}
		}
		l.Push(lua.LBool(sys.replayFile != nil))
		return 1
	})
	luaRegister(l, "esc", func(l *lua.LState) int {
		/*Get or set the global escape flag.
		@function esc
		@tparam[opt] boolean value If provided, sets the escape flag.
		@treturn boolean esc Current value of the escape flag.
		function esc(value) end*/
		if !nilArg(l, 1) {
			sys.esc = boolArg(l, 1)
		}
		l.Push(lua.LBool(sys.esc))
		return 1
	})
	luaRegister(l, "exitNetPlay", func(*lua.LState) int {
		/*Exit netplay mode and close any active netplay connection.
		@function exitNetPlay
		function exitNetPlay() end*/
		if err := sys.endSyncSessionOverride(); err != nil {
			l.RaiseError(err.Error())
		}
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
		/*Exit replay mode and restore normal video settings.
		@function exitReplay
		function exitReplay() end*/
		if err := sys.endSyncSessionOverride(); err != nil {
			l.RaiseError(err.Error())
		}
		if sys.cfg.Video.VSync >= 0 {
			sys.window.SetSwapInterval(sys.cfg.Video.VSync)
		}
		if sys.replayFile != nil {
			sys.replayFile.Close()
			sys.replayFile = nil
		}
		sys.uiResetTokenGuard()
		return 0
	})
	luaRegister(l, "fadeActive", func(*lua.LState) int {
		/*Check whether any global motif fade is active.
		@function fadeActive
		@treturn boolean active `true` if fade-in or fade-out is currently running.
		function fadeActive() end*/
		l.Push(lua.LBool(sys.motif.fadeOut.isFading() || sys.motif.fadeIn.isFading()))
		return 1
	})
	luaRegister(l, "fadeInInit", func(*lua.LState) int {
		/*Initialize a `Fade` object using motif fade-in settings.
		@function fadeInInit
		@tparam Fade fade Fade userdata to initialize.
		function fadeInInit(fade) end*/
		f, ok := toUserData(l, 1).(*Fade)
		if !ok {
			userDataError(l, 1, f)
		}
		sys.motif.fadeOut.reset()
		f.init(sys.motif.fadeIn, true)
		return 0
	})
	luaRegister(l, "fadeOutInit", func(*lua.LState) int {
		/*Initialize a `Fade` object using motif fade-out settings.
		@function fadeOutInit
		@tparam Fade fade Fade userdata to initialize.
		function fadeOutInit(fade) end*/
		f, ok := toUserData(l, 1).(*Fade)
		if !ok {
			userDataError(l, 1, f)
		}
		sys.motif.fadeIn.reset()
		f.init(sys.motif.fadeOut, false)
		return 0
	})
	luaRegister(l, "fadeSkip", func(*lua.LState) int {
		/*Immediately stop any running global motif fade.
		@function fadeSkip
		function fadeSkip() end*/
		sys.motif.fadeIn.reset()
		sys.motif.fadeOut.reset()
		return 0
	})
	luaRegister(l, "fileExists", func(l *lua.LState) int {
		/*Test whether a file exists, after engine path resolution.
		@function fileExists
		@tparam string path File path to test (relative or absolute).
		@treturn boolean exists `true` if the file exists, `false` otherwise.
		function fileExists(path) end*/
		path := strArg(l, 1)
		l.Push(lua.LBool(FileExist(path) != ""))
		return 1
	})
	luaRegister(l, "findEntityByName", func(*lua.LState) int {
		/*Find the next entity whose name contains the given text.
		@function findEntityByName
		@tparam string text Case-insensitive substring to search in entity names.
		function findEntityByName(text) end*/
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
			l.RaiseError("Could not find an entity matching \"%s\"", text)
		}

		return 0
	})
	luaRegister(l, "findEntityByPlayerId", func(*lua.LState) int {
		/*Find the next entity with the given player ID.
		@function findEntityByPlayerId
		@tparam int32 playerId Target entity `id` to search for.
		function findEntityByPlayerId(playerId) end*/
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
	luaRegister(l, "findHelperById", func(*lua.LState) int {
		/*Find the next helper with the given helper ID.
		@function findHelperById
		@tparam int32 helperId Target helper `helperId` to search for.
		function findHelperById(helperId) end*/
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
		/*Get basic font definition information.
		@function fontGetDef
		@tparam Fnt font Font userdata.
		@treturn table def A table:
		  - `Type` (string) font type identifier
		  - `Size` (uint16[2]) `{width, height}` in pixels
		  - `Spacing` (int32[2]) `{x, y}` spacing in pixels
		  - `offset` (int32[2]) `{x, y}` base drawing offset
		function fontGetDef(font) end*/
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
		/*Load a font from file.
		@function fontNew
		@tparam string filename Font filename (searched in `font/`, motif folder, current dir, `data/`).
		@tparam[opt=-1] int32 height Override font height; `-1` uses the height defined in the font file.
		@treturn Fnt font Loaded font userdata. If loading fails, a fallback font is returned.
		function fontNew(filename, height) end*/
		var height int32 = -1
		if !nilArg(l, 2) {
			height = int32(numArg(l, 2))
		}
		filename := SearchFile(strArg(l, 1), []string{"font/", sys.motif.Def, "", "data/"})
		fnt, err := loadFnt(filename, height)
		if err != nil {
			LogMessage("Failed to load %v (screenpack font): %v", filename, err)
			fnt = newFnt()
		}
		l.Push(newUserData(l, fnt))
		return 1
	})
	luaRegister(l, "frameStep", func(*lua.LState) int {
		/*Enable single-frame stepping mode.
		@function frameStep
		function frameStep() end*/
		sys.frameStepFlag = true
		return 0
	})
	luaRegister(l, "game", func(l *lua.LState) int {
		/*Execute a full match using the current configuration.
		@function game
		@treturn int32 winSide Winning side index (`1` or `2`), `0` for draw, `-1` if the game was ended externally.
		@treturn[opt] int controllerNo 1-based controller index of the challenger player interrupting `arcade` mode.
		function game() end*/
		sys.luaDiscardDrawQueue()
		sys.gameRunning = true
		sys.motif.fadeIn.reset()
		sys.motif.fadeOut.reset()
		sys.endMatch = false

		// Synchronize timing to prevent speed fluctuations when changing FPS (entering matches)
		sys.resetFrameTime()

		load := func() error {
			sys.loader.runTread()
			for sys.loader.state != LS_Complete {
				if sys.loader.state == LS_Error {
					return sys.loader.err
				} else if sys.loader.state == LS_Cancel {
					return nil
				}
				sys.await(sys.gameRenderSpeed())
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
			//sys.matchWins = [2]int32{} // Now set directly by Lua so we don't need to reset it
			sys.scoreRounds = [][2]float32{}

			// Reset lifebars
			for i := range sys.fightScreen.winIcons {
				sys.fightScreen.winIcons[i].clear()
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

						// Load palettes if character is just joining the match
						if c[0].roundsExisted() == 0 {
							c[0].loadPalettes()
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
					sys.teamLeader = [2]int{0, 1}
					sys.stage.reset()

					// Adjust matchWins for Turns mode
					// Let's trust that Lua already set this up correctly
					//for i := 0; i < 2; i++ {
					//	if sys.tmode[i] == TM_Turns {
					//		sys.matchWins[i^1] = sys.numTurns[i]
					//	}
					//}
				}

				// Winning player index
				// -1 on quit, -2 on restarting match
				winp := int32(0)

				// Match loop
				if sys.runMatch() {
					// Match is restarting
					for i, reload := range sys.reloadCharSlot {
						if !reload {
							continue
						}
						if sys.cgi[i].sff != nil && !sys.cfg.Debug.KeepSpritesOnReload {
							// removeSFFCache(sys.cgi[i].sff.filename)
							sys.cgi[i].sff = nil
						}
						if sys.reloadPreserveVars[i] {
							sys.saveCharVars(i)
						}
						sys.chars[i] = []*Char{}
						sys.reloadCharSlot[i] = false
					}
					if sys.reloadStageFlg {
						sys.stage = nil
					}
					if sys.reloadFightScreenFlg {
						if err := sys.fightScreen.reload(); err != nil {
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
				// Hard reset: drop the incomplete match stats and start a fresh one
				if winp == -2 {
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
						sys.fightScreen.faces[TM_Turns][i].numko++
						sys.fightScreen.names[TM_Turns][i].numko++
					}
				}

				sys.loader.reset()
			}

			// If not restarting match
			if winp != -2 {
				sys.esc = false
				sys.keyInput = KeyUnknown
				if sys.gameMode == "challenger" {
					sys.statsLog.discardCurrentMatch()
				} else {
					sys.statsLog.finalizeMatch()
				}
				// Cleanup
				sys.timerStart = 0
				sys.timerRounds = []int32{}
				sys.scoreStart = [2]float32{}
				//sys.scoreRounds = [][2]float32{}
				sys.timerCount = []int32{}
				sys.sel.cdefOverwrite = make(map[int]string)
				sys.sel.palOverwrite = make(map[int]int)
				sys.sel.sdefOverwrite = ""
				for i := range sys.reloadPreserveVars {
					sys.reloadPreserveVars[i] = false
				}
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
				sys.motif.fadeIn.reset()
				sys.motif.fadeOut.reset()
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
	luaRegister(l, "gameRunning", func(l *lua.LState) int {
		/*Check whether a match is currently running.
		@function gameRunning
		@treturn boolean running `true` if gameplay is currently active.
		function gameRunning() end*/
		l.Push(lua.LBool(sys.gameRunning))
		return 1
	})
	luaRegister(l, "getAnimElemCount", func(*lua.LState) int {
		/*[redirectable] Get the character's number of elements in the animation.
		@function getAnimElemCount
		@treturn int count Number of animation elements.
		function getAnimElemCount() end*/
		l.Push(lua.LNumber(len(sys.debugWC.anim.frames)))
		return 1
	})
	luaRegister(l, "getAnimTimeSum", func(*lua.LState) int {
		/*[redirectable] Get the character's current accumulated time of the animation.
		@function getAnimTimeSum
		@treturn int32 timeSum Current animation time value.
		function getAnimTimeSum() end*/
		l.Push(lua.LNumber(sys.debugWC.anim.curtime))
		return 1
	})
	luaRegister(l, "getCharAttachedInfo", func(*lua.LState) int {
		/*Resolve and read basic information from a character definition file.
		@function getCharAttachedInfo
		@tparam string def Character identifier or `.def` path. If no extension is given,
		  `"chars/<def>/<def>.def"` is assumed.
		@treturn table|nil info A table:
		  - `name` (string) character display name (or internal name as fallback)
		  - `def` (string) resolved `.def` path
		  - `sound` (string) sound file path from the `[Files]` section,
		  or `nil` if the `.def` file cannot be resolved.
		function getCharAttachedInfo(def) end*/
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
		lines, i := SplitAndTrim(str, "\n"), 0
		info, files, name, sound := true, true, "", ""
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
		/*Get the definition file path for a character slot.
		@function getCharFileName
		@tparam int charRef 0-based character index in the select list.
		@treturn string defPath Resolved `.def` path for this slot.
		function getCharFileName(charRef) end*/
		c := sys.sel.GetChar(int(numArg(l, 1)))
		l.Push(lua.LString(c.def))
		return 1
	})
	luaRegister(l, "getCharInfo", func(*lua.LState) int {
		/*Get detailed information about a character slot.
		@function getCharInfo
		@tparam int charRef 0-based character index in the select list.
		@treturn table info A table:
		  - `name` (string) character display name
		  - `author` (string) author string
		  - `def` (string) definition file path
		  - `sound` (string) sound file path
		  - `intro` (string) intro def path
		  - `ending` (string) ending def path
		  - `arcadepath` (string) arcade path override
		  - `ratiopath` (string) ratio path override
		  - `localcoord` (float32) base localcoord width
		  - `portraitscale` (float32) scale applied to portraits
		  - `cns_scale` (float32[]) scale values from the CNS configuration
		  - `pal` (int32[]) available palette numbers (at least `{1}`)
		  - `pal_defaults` (int32[]) default palette numbers (at least `{1}`)
		  - `pal_keymap` (table) palette key remaps, indexed by original palette slot
		function getCharInfo(charRef) end*/
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
		/*Get the movelist file path for a character slot.
		@function getCharMovelist
		@tparam int charRef 0-based character index in the select list.
		@treturn string movelist Path or identifier of the movelist file.
		function getCharMovelist(charRef) end*/
		c := sys.sel.GetChar(int(numArg(l, 1)))
		l.Push(lua.LString(c.movelist))
		return 1
	})
	luaRegister(l, "getCharName", func(*lua.LState) int {
		/*Get the display name of a character slot.
		@function getCharName
		@tparam int charRef 0-based character index in the select list.
		@treturn string name Character display name.
		function getCharName(charRef) end*/
		c := sys.sel.GetChar(int(numArg(l, 1)))
		l.Push(lua.LString(c.name))
		return 1
	})
	luaRegister(l, "getCharRandomPalette", func(*lua.LState) int {
		/*Get a random valid palette number for a character slot.
		@function getCharRandomPalette
		@tparam int charRef 0-based character index in the select list.
		@treturn int32 palNo Palette number; defaults to `1` if the character has no palette list.
		function getCharRandomPalette(charRef) end*/
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
		/*Get parsed select parameters for a character entry.
		@function getCharSelectParams
		@tparam int charRef 0-based character index in the select list.
		@treturn table params Lua table created from the comma-separated `params` string passed to `addChar()`.
		function getCharSelectParams(charRef) end*/
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
		/*Get the current system clipboard string.
		@function getClipboardString
		@treturn string text Clipboard contents, or an empty string if unavailable.
		function getClipboardString() end*/
		s := sys.window.GetClipboardString()
		l.Push(lua.LString(s))
		return 1
	})
	luaRegister(l, "getCommandLineFlags", func(*lua.LState) int {
		/*Get all command-line flags passed to the engine.
		@function getCommandLineFlags
		@treturn table flags A table mapping raw flag keys to their values (string).
		function getCommandLineFlags() end*/
		tbl := l.NewTable()
		for k, v := range sys.cmdFlags {
			tbl.RawSetString(k, lua.LString(v))
		}
		l.Push(tbl)
		return 1
	})
	luaRegister(l, "getCommandLineValue", func(*lua.LState) int {
		/*Get the value of a specific command-line flag.
		@function getCommandLineValue
		@tparam string flagName Exact flag key as stored in `sys.cmdFlags`
		@treturn string|nil value Value associated with the flag,
		  or `nil` if the flag is not present.
		function getCommandLineValue(flagName) end*/
		if _, ok := sys.cmdFlags[strArg(l, 1)]; !ok {
			return 0
		}
		l.Push(lua.LString(sys.cmdFlags[strArg(l, 1)]))
		return 1
	})
	luaRegister(l, "getConsecutiveWins", func(l *lua.LState) int {
		/*Get the number of consecutive wins for a team.
		@function getConsecutiveWins
		@tparam int teamSide Team side (`1` or `2`).
		@treturn int32 wins Number of consecutive wins for the given side.
		function getConsecutiveWins(teamSide) end*/
		tn := int(numArg(l, 1))
		if tn < 1 || tn > 2 {
			l.RaiseError("\nInvalid team side: %v\n", tn)
		}
		l.Push(lua.LNumber(sys.consecutiveWins[tn-1]))
		return 1
	})
	luaRegister(l, "getCredits", func(*lua.LState) int {
		/*Get the current credit count.
		@function getCredits
		@treturn int32 credits Current number of credits.
		function getCredits() end*/
		l.Push(lua.LNumber(sys.credits))
		return 1
	})
	luaRegister(l, "getDirectoryFiles", func(*lua.LState) int {
		/*Recursively list all paths under a directory.
		@function getDirectoryFiles
		@tparam string rootPath Starting directory path.
		@treturn table paths Array-like table of visited paths (files and directories).
		function getDirectoryFiles(rootPath) end*/
		dir := l.NewTable()
		filepath.Walk(strArg(l, 1), func(path string, info os.FileInfo, err error) error {
			dir.Append(lua.LString(path))
			return nil
		})
		l.Push(dir)
		return 1
	})
	luaRegister(l, "getFrameCount", func(l *lua.LState) int {
		/*Get the global frame counter value.
		@function getFrameCount
		@treturn int32 frameCount Number of frames elapsed since engine start.
		function getFrameCount() end*/
		l.Push(lua.LNumber(sys.frameCounter))
		return 1
	})
	luaRegister(l, "getGameFPS", func(*lua.LState) int {
		/*Get the current measured gameplay FPS.
		@function getGameFPS
		@treturn float32 fps Current gameplay frames per second.
		function getGameFPS() end*/
		l.Push(lua.LNumber(sys.gameFPS))
		return 1
	})
	luaRegister(l, "getGameParams", func(*lua.LState) int {
		/*Get the current game parameter table.
		@function getGameParams
		@treturn table params Current game parameters as a Lua table.
		function getGameParams() end*/
		lv := toLValue(l, sys.sel.gameParams)
		lTable, ok := lv.(*lua.LTable)
		if !ok {
			l.RaiseError("Error: 'lv' is not a *lua.LTable")
			return 0
		}
		l.Push(lTable)
		return 1
	})
	luaRegister(l, "getGameSpeed", func(*lua.LState) int {
		/*Get the current game logic speed as a percentage.
		@function getGameSpeed
		@treturn int32 speedPercent Integer game logic speed relative to `60 FPS` (`100` = normal speed).
		function getGameSpeed() end*/
		l.Push(lua.LNumber(100 * sys.gameLogicSpeed() / 60))
		return 1
	})
	luaRegister(l, "getGameStats", func(*lua.LState) int {
		/*Read accumulated game statistics.
		@function getGameStats
		@treturn table stats Statistics log object as a Lua table.
		function getGameStats() end*/
		l.Push(toLValue(l, sys.statsLog))
		return 1
	})
	luaRegister(l, "getGameStatsJson", func(l *lua.LState) int {
		/*Get a JSON snapshot of accumulated game statistics.
		@function getGameStatsJson
		@treturn string json JSON-encoded snapshot containing stats and related flags.
		function getGameStatsJson() end*/
		s := GameStatsSnapshot{
			StatsLog:          sys.statsLog,
			ContinueFlg:       sys.continueFlg,
			PersistRoundCount: sys.persistRoundCount,
		}
		data, err := json.Marshal(s)
		if err != nil {
			l.RaiseError("getGameStatsJson: %v", err)
			return 0
		}
		l.Push(lua.LString(data))
		return 1
	})
	luaRegister(l, "getInput", func(l *lua.LState) int {
		/*Check raw UI input for one or more players.
		@function getInput
		@tparam number|table players 1-based player/controller index, or an array-like table of indexes.
		  Passing `-1` checks all players.
		@tparam string|table ... One or more key/button tokens, or array-like tables of tokens.
		@treturn boolean pressed `true` if any provided token set is active for any selected player.
		function getInput(players, ...) end*/
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
				if len(btns) > 0 && sys.uiRawInput(btns, controllerIdx) {
					l.Push(lua.LBool(true))
					return 1
				}
			}
		}
		l.Push(lua.LBool(false))
		return 1
	})
	luaRegister(l, "getInputTime", func(l *lua.LState) int {
		/*Get the hold time of a raw input token for one or more players.
		@function getInputTime
		@tparam number|table players 1-based player/controller index, or an array-like table of indexes.
		  Passing `-1` checks all players.
		@tparam string|table ... One or more key/button tokens, or array-like tables of tokens.
		@treturn int32 time Hold time in ticks for the first active token found, or `0` if none are active.
		function getInputTime(players, ...) end*/
		var players []int
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
			l.Push(lua.LNumber(0))
			return 1
		}
		// Helper to compute the hold time for a key token on a given controller index.
		keyTime := func(controllerIdx int, key string) int32 {
			if controllerIdx == -2 && sys.lastInputController >= 0 {
				want := sys.lastInputController
				controllerIdx = -1
				for i := 0; i < len(sys.commandLists); i++ {
					if sys.uiControllerKey(i) == want {
						controllerIdx = i
						break
					}
				}
				if controllerIdx < 0 {
					return 0
				}
			}
			if controllerIdx < 0 || controllerIdx >= len(sys.commandLists) ||
				sys.commandLists[controllerIdx] == nil || sys.commandLists[controllerIdx].Buffer == nil {
				return 0
			}
			cl := sys.commandLists[controllerIdx]
			ib := sys.commandLists[controllerIdx].Buffer
			var v int32
			// Treat LS direction tokens as aliases for D/U/B/F.
			switch key {
			case "LS_Y+":
				key = "D"
			case "LS_Y-":
				key = "U"
			case "LS_X-":
				key = "B"
			case "LS_X+":
				key = "F"
			}
			switch key {
			case "B":
				v = ib.Bb
			case "D":
				v = ib.Db
			case "F":
				v = ib.Fb
			case "U":
				v = ib.Ub
			case "L":
				v = ib.Lb
			case "R":
				v = ib.Rb
			case "a":
				v = ib.ab
			case "b":
				v = ib.bb
			case "c":
				v = ib.cb
			case "x":
				v = ib.xb
			case "y":
				v = ib.yb
			case "z":
				v = ib.zb
			case "s":
				v = ib.sb
			case "d":
				v = ib.db
			case "w":
				v = ib.wb
			case "m":
				v = ib.mb
			default:
				// Other raw controller tokens (RS_*, LT/RT, etc): just report held (1) or not (0).
				if idx, ok := StringToButtonLUT[key]; ok && idx != 25 {
					if cl.IsControllerButtonPressed(key, controllerIdx) {
						return 1
					}
					return 0
				}
				l.RaiseError("\nInvalid argument: %v\n", key)
				return 0
			}
			// 0 if < 0, otherwise 1+
			if v < 0 {
				return 0
			} else if v == 0 {
				return 1
			}
			return v
		}
		top := l.GetTop()
		for _, pn := range players {
			if pn == 0 {
				continue
			}
			controllerIdx := pn - 1
			if pn == -1 {
				controllerIdx = -2
			}
			// Loop over all key tables/strings passed
			for ai := 2; ai <= top; ai++ {
				arg := l.Get(ai)
				var keys []string
				switch t := arg.(type) {
				case *lua.LTable:
					t.ForEach(func(_ lua.LValue, v lua.LValue) {
						if s, ok := v.(lua.LString); ok {
							keys = append(keys, string(s))
						}
					})
				case lua.LString:
					keys = []string{string(t)}
				default:
					continue
				}
				for _, key := range keys {
					if out := keyTime(controllerIdx, key); out > 0 {
						l.Push(lua.LNumber(out))
						return 1
					}
				}
			}
		}
		l.Push(lua.LNumber(0))
		return 1
	})
	luaRegister(l, "getJoystickGUID", func(*lua.LState) int {
		/*Get a joystick's GUID string.
		@function getJoystickGUID
		@tparam int index Joystick index (0-based).
		@treturn string guid GUID string for the joystick, or an empty string if invalid.
		function getJoystickGUID(index) end*/
		l.Push(lua.LString(input.GetJoystickGUID(int(numArg(l, 1)))))
		return 1
	})
	luaRegister(l, "getJoystickKey", func(*lua.LState) int {
		/*Poll joystick input and return the corresponding key string.
		@function getJoystickKey
		@tparam[opt] int controllerIdx Joystick index (0-based). If omitted, listens on any joystick.
		@treturn string keyName Engine key string for the pressed control (empty string if none).
		@treturn int joystickIndex 1-based joystick index that generated the input; `-1` if no input.
		function getJoystickKey(controllerIdx) end*/
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
	luaRegister(l, "getJoystickName", func(*lua.LState) int {
		/*Get a joystick's display name.
		@function getJoystickName
		@tparam int index Joystick index (0-based).
		@treturn string name Human-readable joystick name, or an empty string if invalid.
		function getJoystickName(index) end*/
		l.Push(lua.LString(input.GetJoystickName(int(numArg(l, 1)))))
		return 1
	})
	luaRegister(l, "getJoystickPresent", func(*lua.LState) int {
		/*Check whether a joystick is present.
		@function getJoystickPresent
		@tparam int index Joystick index (0-based).
		@treturn boolean present `true` if the joystick is connected, `false` otherwise.
		function getJoystickPresent(index) end*/
		l.Push(lua.LBool(input.IsJoystickPresent(int(numArg(l, 1)))))
		return 1
	})
	luaRegister(l, "getKey", func(*lua.LState) int {
		/*Query or compare the last pressed key.
		@function getKey
		@tparam[opt] string key If omitted, the last key name is returned. If a non-empty string
		  is given, returns whether it matches the last key. If an empty string is given,
		  always returns `false`.
		@treturn string|boolean result Last key name when called without arguments, or a boolean match result when `key` is provided.
		function getKey(key) end*/
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
		/*Get the last input text associated with the current key event.
		@function getKeyText
		@treturn string text If the last key was Insert, returns the clipboard contents,
		  otherwise returns the textual representation of the last key press. Empty string if none.
		function getKeyText() end*/
		s := ""
		if sys.keyInput != KeyUnknown {
			s = sys.keyString
		}
		l.Push(lua.LString(s))
		return 1
	})
	luaRegister(l, "getLastInputController", func(l *lua.LState) int {
		/*Get the last controller that produced UI input.
		@function getLastInputController
		@treturn int playerNo 1-based player/controller index, or `-1` if unavailable.
		function getLastInputController() end*/
		if sys.lastInputController >= 0 {
			l.Push(lua.LNumber(sys.lastInputController + 1))
		} else {
			l.Push(lua.LNumber(-1))
		}
		return 1
	})
	luaRegister(l, "getMatchTime", func(*lua.LState) int {
		/*Get the accumulated match time from completed rounds.
		@function getMatchTime
		@treturn int32 time Total round time accumulated in ticks.
		function getMatchTime() end*/
		var ti int32
		for _, v := range sys.timerRounds {
			ti += v
		}
		l.Push(lua.LNumber(ti))
		return 1
	})
	luaRegister(l, "getRandom", func(l *lua.LState) int {
		/*Return a 32-bit random number, updating the global seed.
		@function getRandom
		@treturn int32 value Random value (1 to 2147483646 inclusive).
		function getRandom() end*/
		l.Push(lua.LNumber(Random()))
		return 1
	})
	luaRegister(l, "getRemapInput", func(l *lua.LState) int {
		/*Get the input remap target for a player.
		@function getRemapInput
		@tparam int playerNo 1-based player/controller index.
		@treturn int mappedPlayerNo 1-based remapped player/controller index.
		function getRemapInput(playerNo) end*/
		pn := int(numArg(l, 1))
		if pn < 1 || pn > len(sys.inputRemap) {
			l.RaiseError("\nInvalid player number: %v\n", pn)
		}
		l.Push(lua.LNumber(sys.inputRemap[pn-1] + 1))
		return 1
	})
	luaRegister(l, "getRoundTime", func(l *lua.LState) int {
		/*Get the configured round time limit.
		@function getRoundTime
		@treturn int32 time Round time limit in ticks (or special values as configured).
		function getRoundTime() end*/
		l.Push(lua.LNumber(sys.maxRoundTime))
		return 1
	})
	luaRegister(l, "getRuntimeOS", func(l *lua.LState) int {
		/*Get the current runtime operating system identifier.
		@function getRuntimeOS
		@treturn string os Runtime OS name as reported by Go (for example `"windows"`, `"linux"`).
		function getRuntimeOS() end*/
		l.Push(lua.LString(runtime.GOOS))
		return 1
	})
	luaRegister(l, "getSelectNo", func(*lua.LState) int {
		/*[redirectable] Get the character's select slot index.
		@function getSelectNo
		@treturn int selectNo Current select slot index.
		function getSelectNo() end*/
		l.Push(lua.LNumber(sys.debugWC.selectNo))
		return 1
	})
	luaRegister(l, "getSessionWarning", func(*lua.LState) int {
		/*Pop the current session warning message.
		@function getSessionWarning
		@treturn string warning Current session warning message, or an empty string if none.
		function getSessionWarning() end*/
		l.Push(lua.LString(sys.popSessionWarning()))
		return 1
	})
	luaRegister(l, "getStageInfo", func(*lua.LState) int {
		/*Get information about a stage slot.
		@function getStageInfo
		@tparam int stageRef Stage reference used by the select system.
		  Positive values are 1-based stage slots. `0` is the random-stage sentinel.
		@treturn table|nil info A table:
		  - `name` (string) stage display name
		  - `def` (string) definition file path
		  - `localcoord` (float32) base localcoord width
		  - `portraitscale` (float32) scale applied to stage portraits
		  - `attachedchardef` (string[]) list of attached character `.def` paths
		function getStageInfo(stageRef) end*/
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
		/*Get the currently selected stage slot index.
		@function getStageNo
		@treturn int stageRef Currently selected stage reference as stored by the select system:
		  `0` means random stage, `-1` means no stage selected, positive values are 1-based stage slots.
		function getStageNo() end*/
		l.Push(lua.LNumber(sys.sel.selectedStageNo))
		return 1
	})
	luaRegister(l, "getStageSelectParams", func(*lua.LState) int {
		/*Get parsed select parameters for a stage entry.
		@function getStageSelectParams
		@tparam int stageRef Stage index as used by the select system.
		@treturn table params Lua table created from the comma-separated `params` string passed to `addStage()`.
		function getStageSelectParams(stageRef) end*/
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
	luaRegister(l, "getStateOwnerId", func(*lua.LState) int {
		/*[redirectable] Get the character's player ID.
		@function getStateOwnerId
		@treturn int32 playerId Player ID of the current state owner.
		function getStateOwnerId() end*/
		l.Push(lua.LNumber(sys.debugWC.stateOwner().id))
		return 1
	})
	luaRegister(l, "getStateOwnerName", func(*lua.LState) int {
		/*[redirectable] Get the character's name of the current state owner.
		@function getStateOwnerName
		@treturn string name Name of the current state owner.
		function getStateOwnerName() end*/
		l.Push(lua.LString(sys.debugWC.stateOwner().name))
		return 1
	})
	luaRegister(l, "getStateOwnerPlayerNo", func(*lua.LState) int {
		/*[redirectable] Get the character's player number of the current state owner.
		@function getStateOwnerPlayerNo
		@treturn int playerNo 1-based player number of the current state owner.
		function getStateOwnerPlayerNo() end*/
		l.Push(lua.LNumber(sys.debugWC.stateOwner().playerNo + 1))
		return 1
	})
	luaRegister(l, "getStoryboardScene", func(l *lua.LState) int {
		/*Get the current storyboard scene index.
		@function getStoryboardScene
		@treturn int|nil sceneIndex Current storyboard scene index, or `nil` if no storyboard is active.
		function getStoryboardScene() end*/
		if sys.storyboard.active {
			l.Push(lua.LNumber(sys.storyboard.currentSceneIndex))
		} else {
			l.Push(lua.LNil)
		}
		return 1
	})
	luaRegister(l, "getTimestamp", func(*lua.LState) int {
		/*Get a formatted timestamp string.
		@function getTimestamp
		@tparam[opt="2006-01-02 15:04:05.000"] string format Go-style time format layout.
		@treturn string timestamp Current time formatted according to `format`.
		function getTimestamp(format) end*/
		format := "2006-01-02 15:04:05.000"
		if !nilArg(l, 1) {
			format = strArg(l, 1)
		}
		l.Push(lua.LString(time.Now().Format(format)))
		return 1
	})
	luaRegister(l, "getWinnerTeam", func(*lua.LState) int {
		/*Get the winning team side of the current or last match.
		@function getWinnerTeam
		@treturn int32 teamSide Winning team side (`1` or `2`), `0` for draw/undecided, or `-1` when unavailable.
		function getWinnerTeam() end*/
		l.Push(lua.LNumber(sys.winnerTeam()))
		return 1
	})
	luaRegister(l, "isUIKeyAction", func(l *lua.LState) int {
		/*Check whether a UI action name is currently active.
		@function isUIKeyAction
		@tparam string action UI action name.
		@treturn boolean active `true` if the action is currently active.
		function isUIKeyAction(action) end*/
		l.Push(lua.LBool(sys.uiIsKeyAction(strArg(l, 1))))
		return 1
	})
	luaRegister(l, "jsonDecode", func(*lua.LState) int {
		/*Decode a JSON file into Lua values.
		@function jsonDecode
		@tparam string path Path to the JSON file.
		@treturn any value Decoded JSON root value (Lua `string`, `number`, `boolean`, `table` or `nil`).
		function jsonDecode(path) end*/
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
		/*Encode a Lua value to JSON and save it to a file.
		@function jsonEncode
		@tparam any value Lua value to encode (tables, numbers, strings, booleans, or `nil`).
		@tparam string path Output JSON file path (parent directories are created as needed).
		function jsonEncode(value, path) end*/
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
	luaRegister(l, "loadAnimTable", func(l *lua.LState) int {
		/*Load an AIR/animation definition file and return an animation table.
		@function loadAnimTable
		@tparam string path Animation file path (typically a .air file).
		@tparam[opt] Sff sff Sprite file userdata used to resolve sprites while parsing.
		  If omitted, an empty SFF is created internally.
		@treturn table animTable Table mapping action numbers (int32) to `Animation` userdata.
		  Each value is a parsed `*Animation` (usable with `animNew` and `animSetAnimation`).
		function loadAnimTable(path, sff) end*/
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
		at := ReadAnimationTable(def, sff, &sff.palList, lines, &i, true)
		// Build Lua table with NUMERIC keys (so it prints like [110] => userdata ...)
		tbl := l.NewTable()
		// Deterministic iteration order
		keys := make([]int, 0, len(at.anims))
		for k := range at.anims {
			keys = append(keys, int(k))
		}
		sort.Ints(keys)
		for _, ki := range keys {
			anim := at.anims[int32(ki)]
			if anim == nil {
				continue
			}
			// toLValue will wrap *Animation into userdata
			tbl.RawSetInt(ki, toLValue(l, anim))
		}
		l.Push(tbl)
		return 1
	})
	luaRegister(l, "loadDebugFont", func(l *lua.LState) int {
		/*Load and set the font used by the debug overlay.
		@function loadDebugFont
		@tparam string filename Font filename.
		@tparam[opt=1.0] float32 scale Uniform scale applied to debug text (both X and Y).
		function loadDebugFont(filename, scale) end*/
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
		/*Register Lua functions to be called for debug info display.
		@function loadDebugInfo
		@tparam table funcs Array-like table of global function names (string).
		function loadDebugInfo(funcs) end*/
		tableArg(l, 1).ForEach(func(_, value lua.LValue) {
			sys.listLFunc = append(sys.listLFunc, sys.luaLState.GetGlobal(lua.LVAsString(value)).(*lua.LFunction))
		})
		return 0
	})
	luaRegister(l, "loadDebugStatus", func(l *lua.LState) int {
		/*Register the Lua function used to draw debug status.
		@function loadDebugStatus
		@tparam string funcName Global Lua function name used for debug status.
		function loadDebugStatus(funcName) end*/
		sys.statusLFunc, _ = sys.luaLState.GetGlobal(strArg(l, 1)).(*lua.LFunction)
		return 0
	})
	luaRegister(l, "loadGameOption", func(l *lua.LState) int {
		/*Load game options from a config file and return the current config as a table.
		@function loadGameOption
		@tparam[opt] string filename Config file path. If omitted, current config is kept.
		@treturn table cfg Table representation of the current game configuration.
		function loadGameOption(filename) end*/
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
		/*Check whether resources are currently being loaded.
		@function loading
		@treturn boolean loading `true` if the loader is in `LS_Loading` state.
		function loading() end*/
		l.Push(lua.LBool(sys.loader.state == LS_Loading))
		return 1
	})
	luaRegister(l, "loadIni", func(l *lua.LState) int {
		/*Load an INI file and convert it to a nested Lua table.
		@function loadIni
		@tparam string filename INI file path.
		@treturn table ini Table of sections; each section is a table of keys to strings.
		  Dotted keys are converted to nested subtables.
		function loadIni(filename) end*/
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
	luaRegister(l, "loadFightScreen", func(l *lua.LState) int {
		/*Load the fight screen definition.
		@function loadFightScreen
		@tparam[opt] string defPath FightScreen def file path. If empty or omitted, uses default.
		function loadFightScreen(defPath) end*/
		def := ""
		if !nilArg(l, 1) {
			def = strArg(l, 1)
		}
		fs, err := loadFightScreen(def)
		if err != nil {
			l.RaiseError("\nCan't load fight screen %v: %v\n", def, err.Error())
		}
		sys.fightScreen = *fs
		return 0
	})
	luaRegister(l, "loadMotif", func(l *lua.LState) int {
		/*Load a motif and return its configuration as a table.
		@function loadMotif
		@tparam[opt] string defPath Motif def file path. If empty or omitted, uses default.
		@treturn table motif Motif configuration table (includes menus, fonts, sounds, etc.).
		function loadMotif(defPath) end*/
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
		// Requires PreserveSurroundedQuote=true when loading INI files.
		normalizeMenuValue := func(v string) (string, bool) {
			raw := v
			if len(raw) >= 2 {
				if (raw[0] == '"' && raw[len(raw)-1] == '"') ||
					(raw[0] == '\'' && raw[len(raw)-1] == '\'') {
					inner := raw[1 : len(raw)-1]
					return inner, inner == ""
				}
			}
			trimmed := strings.TrimSpace(raw)
			return trimmed, trimmed == ""
		}

		isEmpty := func(v string) bool {
			_, empty := normalizeMenuValue(v)
			return empty
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
				val, empty := normalizeMenuValue(v)
				if mode == "flat" {
					k := strings.ReplaceAll(rest, ".", "_")
					// Empty-valued keys disable/remove the entry (override defaults).
					if empty {
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
				if empty {
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

			sort.SliceStable(final, func(i, j int) bool {
				return strings.Count(final[i], "_") < strings.Count(final[j], "_")
			})

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
		/*Validate selection and start asynchronous loading of characters and stage.
		@function loadStart
		@tparam[opt] string params Optional comma-separated parameter string (from launchFight and quickvs options)
		function loadStart(params) end*/
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
			sys.sel.gameParams = newGameParamsFromMotif(&sys.motif)
		} else {
			sys.sel.gameParams.ResetFromMotif(&sys.motif)
		}
		sys.sel.music = make(Music)
		if !nilArg(l, 1) {
			entries := SplitAndTrim(StripComment(strArg(l, 1)), ",")
			sys.sel.gameParams.AppendParams(entries)
			// Feed normalized music params to Music.
			sys.sel.music.AppendParams(sys.sel.gameParams.MusicEntries())
		}
		sys.loadStart()
		return 0
	})
	luaRegister(l, "loadState", func(*lua.LState) int {
		/*Request loading of a previously saved state on the next frame.
		@function loadState
		function loadState() end*/
		sys.loadStateFlag = true
		return 0
	})
	luaRegister(l, "loadStoryboard", func(l *lua.LState) int {
		/*Load a storyboard and set it as the current storyboard.
		@function loadStoryboard
		@tparam string defPath Storyboard def file path.
		@treturn table|nil storyboard Storyboard configuration table on success,
		  or `nil` if no path is given or loading fails (a warning is printed).
		function loadStoryboard(defPath) end*/
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
	luaRegister(l, "loadText", func(l *lua.LState) int {
		/*Load a text file and return its contents.
		@function loadText
		@tparam string path Text file path.
		@treturn string|nil content File contents on success, or `nil` if the file
		  cannot be read.
		function loadText(path) end*/
		path := strArg(l, 1)
		content, err := LoadText(path)
		if err != nil {
			l.Push(lua.LNil)
			return 1
		}
		l.Push(lua.LString(content))
		return 1
	})
	luaRegister(l, "mapSet", func(*lua.LState) int {
		/*[redirectable] Set the character's map value.
		@function mapSet
		@tparam string name Map name to modify.
		@tparam float32 value Map value to set.
		@tparam[opt] string mapType Map operation type. `"add"` adds to the existing value,
		  anything else replaces it.
		function mapSet(name, value, mapType) end*/
		var scType int32
		if !nilArg(l, 3) && strArg(l, 3) == "add" {
			scType = 1
		}
		sys.debugWC.mapSet(strArg(l, 1), float32(numArg(l, 2)), scType)
		return 0
	})
	luaRegister(l, "modelNew", func(l *lua.LState) int {
		/*Load a 3D model (glTF) as a Model object.
		@function modelNew
		@tparam string filename glTF model file path.
		@treturn Model model Model userdata.
		function modelNew(filename) end*/
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
	luaRegister(l, "modifyGameOption", func(l *lua.LState) int {
		/*Modify a game option using a query string path.
		@function modifyGameOption
		@tparam string query Option path in config (for example `"GameSpeed.Value"`).
		@tparam any value New value:
		  - `boolean`: stored as boolean
		  - `nil`: remove/clear value depending on context
		  - `table`: treated as array of strings
		  - other: converted to string
		function modifyGameOption(query, value) end*/
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
		/*Modify a motif using a query string path.
		@function modifyMotif
		@tparam string query Parameter path in motif (for example `"attract_mode.credits.snd"`).
		@tparam any value New value:
		  - `boolean`: stored as boolean
		  - `nil`: remove/clear value depending on context
		  - `table`: treated as array of strings
		  - other: converted to string
		function modifyMotif(query, value) end*/
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
		/*Modify a currently loaded storyboard using a query string path.
		@function modifyStoryboard
		@tparam string query Parameter path in storyboard (for example `"scenedef.stopmusic"`).
		@tparam any value New value:
		  - `boolean`: stored as boolean
		  - `nil`: remove/clear value depending on context
		  - `table`: treated as array of strings
		  - other: converted to string
		function modifyStoryboard(query, value) end*/
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
	luaRegister(l, "netPlay", func(*lua.LState) int {
		/*Check whether the current session is running in netplay mode.
		@function netPlay
		@treturn boolean active `true` if netplay is currently active.
		function netPlay() end*/
		l.Push(lua.LBool(sys.netplay()))
		return 1
	})
	luaRegister(l, "panicError", func(*lua.LState) int {
		/*Raise an immediate Lua error with a custom message.
		@function panicError
		@tparam string message Error message.
		function panicError(message) end*/
		l.RaiseError(strArg(l, 1))
		return 0
	})
	luaRegister(l, "paused", func(*lua.LState) int {
		/*Check whether gameplay is currently paused.
		@function paused
		@treturn boolean paused `true` if the game is paused and not currently frame-stepping.
		function paused() end*/
		l.Push(lua.LBool(sys.paused && !sys.frameStepFlag))
		return 1
	})
	luaRegister(l, "playBgm", func(l *lua.LState) int {
		/*[redirectable] Control background music playback.
		@function playBgm
		@tparam table params Parameter table (keys are case-insensitive):
		  - `source` (string, opt) preset to use, formatted as `"<origin>.<key>"`
		     - `origin` = `"stagedef"`, `"stageparams"`, `"launchparams"`, `"motif"`, `"match"`, `"charparams"`
		     - `key` is origin-specific (for example `"bgm"`, `"win"` etc.)
		  - `bgm` (string, opt) BGM filename; searched relative to the current motif, current directory, and `sound/`
		  - `loop` (int, opt) Loop flag/mode (see engine BGM semantics)
		  - `volume` (int, opt) BGM volume (0–100, clamped to config MaxBGMVolume)
		  - `loopstart` (int, opt) Loop start position (samples/frames, engine-specific)
		  - `loopend` (int, opt) Loop end position
		  - `startposition` (int, opt) Initial playback position
		  - `freqmul` (float32, opt) Frequency multiplier (pitch)
		  - `loopcount` (int, opt) Loop count (`-1` for infinite)
		  - `interrupt` (boolean, opt) If `true`, always restart playback; if `false`, only update volume;
		    if omitted, interruption is decided automatically based on whether the file changed.
		function playBgm(params) end*/
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
	luaRegister(l, "playerBufReset", func(*lua.LState) int {
		/*Reset player input buffers and disable hardcoded keys.
		@function playerBufReset
		@tparam[opt] int playerNo Player index (1-based). If omitted, resets all players.
		function playerBufReset(playerNo) end*/
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
	luaRegister(l, "playSnd", func(l *lua.LState) int {
		/*[redirectable] Play the character's sound.
		@function playSnd
		@tparam[opt=-1] int32 group Sound group number. Negative values mean no sound is played.
		@tparam[opt=0] int32 sound Sound number within the group.
		@tparam[opt=100] int32 volumescale Volume scale (percent).
		@tparam[opt=false] boolean commonSnd If `true`, use the `"f"` fight/common FX sound prefix.
		@tparam[opt=-1] int32 channel Sound channel (`-1` = auto).
		@tparam[opt=false] boolean lowpriority If `true`, sound can be overridden by higher-priority sounds.
		@tparam[opt=1.0] float32 freqmul Frequency multiplier (pitch).
		@tparam[opt=false] boolean loop If `true`, sound loops (ignored if `loopcount` is non-zero).
		@tparam[opt=0.0] float32 pan Stereo panning (engine-specific range, usually -1.0 to 1.0).
		@tparam[opt=0] int32 priority Priority level (higher plays over lower).
		@tparam[opt=0] int loopstart Loop start position.
		@tparam[opt=0] int loopend Loop end position.
		@tparam[opt=0] int startposition Initial playback position.
		@tparam[opt=0] int32 loopcount Loop count: `0` uses `loop` flag, positive = exact loops, negative = infinite.
		@tparam[opt=false] boolean stopOnGetHit If `true`, stop this sound when the character is hit.
		@tparam[opt=false] boolean stopOnChangeState If `true`, stop this sound when the character changes state.
		function playSnd(group, sound, volumescale, commonSnd, channel, lowpriority, freqmul,
		  loop, pan, priority, loopstart, loopend, startposition, loopcount,
		  stopOnGetHit, stopOnChangeState) end*/
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
	luaRegister(l, "postMatch", func(*lua.LState) int {
		/*Check whether post-match processing is active.
		@function postMatch
		@treturn boolean active `true` if the engine is in post-match state.
		function postMatch() end*/
		l.Push(lua.LBool(sys.postMatchFlg))
		return 1
	})
	luaRegister(l, "preloadListChar", func(*lua.LState) int {
		/*Mark a character sprite or animation for preloading.
		@function preloadListChar
		@tparam int32|uint16 id Action number, or sprite group number when `number` is provided.
		@tparam[opt] uint16 number If provided, `id` and `number` are used as sprite group/number keys;
		  otherwise `id` is treated as an animation/action number (`int32`).
		function preloadListChar(id, number) end*/
		if !nilArg(l, 2) {
			sys.sel.charSpritePreload[[...]uint16{uint16(numArg(l, 1)), uint16(numArg(l, 2))}] = true
		} else {
			sys.sel.charAnimPreload[int32(numArg(l, 1))] = true
		}
		return 0
	})
	luaRegister(l, "preloadListStage", func(*lua.LState) int {
		/*Mark a stage sprite or animation for preloading.
		@function preloadListStage
		@tparam int32|uint16 id Action number, or sprite group number when `number` is provided.
		@tparam[opt] uint16 number If provided, `id` and `number` are used as sprite group/number keys;
		  otherwise `id` is treated as an animation/action number (`int32`).
		function preloadListStage(id, number) end*/
		if !nilArg(l, 2) {
			sys.sel.stageSpritePreload[[...]uint16{uint16(numArg(l, 1)), uint16(numArg(l, 2))}] = true
		} else {
			sys.sel.stageAnimPreload[int32(numArg(l, 1))] = true
		}
		return 0
	})
	luaRegister(l, "printConsole", func(l *lua.LState) int {
		/*Print text to the in-game console and standard output.
		@function printConsole
		@tparam string text Text to print.
		@tparam[opt=false] boolean appendLast If `true`, appends to the last console line; otherwise starts a new line.
		function printConsole(text, appendLast) end*/
		if !nilArg(l, 2) && boolArg(l, 2) && len(sys.consoleText) > 0 {
			sys.consoleText[len(sys.consoleText)-1] += strArg(l, 1)
		} else {
			sys.appendToConsole(strArg(l, 1))
		}
		fmt.Println(strArg(l, 1))
		return 0
	})
	luaRegister(l, "puts", func(*lua.LState) int {
		/*Print text to standard output (stdout) only.
		@function puts
		@tparam string text Text to print.
		function puts(text) end*/
		fmt.Println(strArg(l, 1))
		return 0
	})
	luaRegister(l, "rectDebug", func(*lua.LState) int {
		/*Print a rectangle's debug information.
		@function rectDebug
		@tparam Rect rect Rectangle userdata.
		@tparam[opt] string prefix Optional text printed before the rectangle.
		function rectDebug(rect, prefix) end*/
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
		/*Queue drawing of a rectangle.
		@function rectDraw
		@tparam Rect rect Rectangle userdata.
		@tparam[opt] int16 layer Layer number to draw on (defaults to `rect.layerno`).
		function rectDraw(rect, layer) end*/
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
		/*Create a new rectangle object.
		@function rectNew
		@treturn Rect rect Newly created rectangle userdata.
		function rectNew() end*/
		rect := NewRect()
		l.Push(newUserData(l, rect))
		return 1
	})
	luaRegister(l, "rectReset", func(*lua.LState) int {
		/*Reset rectangle parameters to defaults.
		@function rectReset
		@tparam Rect rect Rectangle userdata.
		function rectReset(rect) end*/
		r, ok := toUserData(l, 1).(*Rect)
		if !ok {
			userDataError(l, 1, r)
		}
		r.Reset()
		return 0
	})
	luaRegister(l, "rectSetAlpha", func(*lua.LState) int {
		/*Set rectangle alpha blending values.
		@function rectSetAlpha
		@tparam Rect rect Rectangle userdata.
		@tparam int32 src Source alpha.
		@tparam int32 dst Destination alpha.
		function rectSetAlpha(rect, src, dst) end*/
		r, ok := toUserData(l, 1).(*Rect)
		if !ok {
			userDataError(l, 1, r)
		}
		r.SetAlpha([2]int32{int32(numArg(l, 2)), int32(numArg(l, 3))})
		return 0
	})
	luaRegister(l, "rectSetAlphaPulse", func(*lua.LState) int {
		/*Enable pulsing alpha effect for a rectangle.
		@function rectSetAlphaPulse
		@tparam Rect rect Rectangle userdata.
		@tparam int32 min Minimum alpha.
		@tparam int32 max Maximum alpha.
		@tparam int32 time Pulse period (frames).
		function rectSetAlphaPulse(rect, min, max, time) end*/
		r, ok := toUserData(l, 1).(*Rect)
		if !ok {
			userDataError(l, 1, r)
		}
		r.SetAlphaPulse(int32(numArg(l, 2)), int32(numArg(l, 3)), int32(numArg(l, 4)))
		return 0
	})
	luaRegister(l, "rectSetColor", func(*lua.LState) int {
		/*Set rectangle RGB color.
		@function rectSetColor
		@tparam Rect rect Rectangle userdata.
		@tparam int32 r Red component (0–255).
		@tparam int32 g Green component (0–255).
		@tparam int32 b Blue component (0–255).
		function rectSetColor(rect, r, g, b) end*/
		r, ok := toUserData(l, 1).(*Rect)
		if !ok {
			userDataError(l, 1, r)
		}
		r.SetColor([3]int32{int32(numArg(l, 2)), int32(numArg(l, 3)), int32(numArg(l, 4))})
		return 0
	})
	luaRegister(l, "rectSetLayerno", func(*lua.LState) int {
		/*Set the rectangle's drawing layer.
		@function rectSetLayerno
		@tparam Rect rect Rectangle userdata.
		@tparam int16 layer Layer number.
		function rectSetLayerno(rect, layer) end*/
		r, ok := toUserData(l, 1).(*Rect)
		if !ok {
			userDataError(l, 1, r)
		}
		r.layerno = int16(numArg(l, 2))
		return 0
	})
	luaRegister(l, "rectSetLocalcoord", func(*lua.LState) int {
		/*Set the rectangle's local coordinate system.
		@function rectSetLocalcoord
		@tparam Rect rect Rectangle userdata.
		@tparam float32 x Local coordinate width.
		@tparam float32 y Local coordinate height.
		function rectSetLocalcoord(rect, x, y) end*/
		r, ok := toUserData(l, 1).(*Rect)
		if !ok {
			userDataError(l, 1, r)
		}
		r.SetLocalcoord(float32(numArg(l, 2)), float32(numArg(l, 3)))
		return 0
	})
	luaRegister(l, "rectSetWindow", func(*lua.LState) int {
		/*Set the rectangle's clipping window.
		@function rectSetWindow
		@tparam Rect rect Rectangle userdata.
		@tparam float32 x1 Left coordinate.
		@tparam float32 y1 Top coordinate.
		@tparam float32 x2 Right coordinate.
		@tparam float32 y2 Bottom coordinate.
		function rectSetWindow(rect, x1, y1, x2, y2) end*/
		r, ok := toUserData(l, 1).(*Rect)
		if !ok {
			userDataError(l, 1, r)
		}
		r.SetWindow([4]float32{float32(numArg(l, 2)), float32(numArg(l, 3)), float32(numArg(l, 4)), float32(numArg(l, 5))})
		return 0
	})
	luaRegister(l, "rectUpdate", func(*lua.LState) int {
		/*Update rectangle animation (alpha pulse, etc.).
		@function rectUpdate
		@tparam Rect rect Rectangle userdata.
		function rectUpdate(rect) end*/
		r, ok := toUserData(l, 1).(*Rect)
		if !ok {
			userDataError(l, 1, r)
		}
		r.Update()
		return 0
	})
	luaRegister(l, "refresh", func(*lua.LState) int {
		/*Advance one frame: process logic, drawing and fades.
		@function refresh
		function refresh() end*/
		sys.tickSound()
		if !sys.frameSkip {
			sys.luaFlushDrawQueue()
			if sys.motif.fadeOut.isActive() {
				sys.motif.fadeOut.draw()
			} else if sys.motif.fadeIn.isActive() {
				sys.motif.fadeIn.draw()
			}
		} else {
			// On skipped frames, discard queued draws to avoid buildup.
			sys.luaDiscardDrawQueue()
		}
		// Advance motif fades every logical UI frame, even on skipped frames.
		if sys.motif.fadeOut.isActive() {
			sys.motif.fadeOut.step()
		} else if sys.motif.fadeIn.isActive() {
			sys.motif.fadeIn.step()
		}
		sys.stepCommandLists()
		if !sys.update() {
			l.RaiseError("<game end>")
		}
		return 0
	})
	luaRegister(l, "reload", func(*lua.LState) int {
		/*Schedule reloading of characters, stage and fight screen.
		@function reload
		function reload() end*/
		sys.reloadFlg = true
		for i := range sys.reloadCharSlot {
			sys.reloadCharSlot[i] = true
		}
		sys.reloadStageFlg = true
		sys.reloadFightScreenFlg = true
		return 0
	})
	luaRegister(l, "remapInput", func(l *lua.LState) int {
		/*Remap logical player input to another player slot.
		@function remapInput
		@tparam int32 srcPlayer Source player number (1-based).
		@tparam int32 dstPlayer Destination player number (1-based).
		function remapInput(srcPlayer, dstPlayer) end*/
		src, dst := int(numArg(l, 1)), int(numArg(l, 2))
		if src < 1 || src > len(sys.inputRemap) ||
			dst < 1 || dst > len(sys.inputRemap) {
			l.RaiseError("\nInvalid player number: %v, %v\n", src, dst)
		}
		sys.inputRemap[src-1] = dst - 1
		return 0
	})
	luaRegister(l, "removeDizzy", func(*lua.LState) int {
		/*[redirectable] Clear the character's dizzy state.
		@function removeDizzy
		function removeDizzy() end*/
		sys.debugWC.unsetSCF(SCF_dizzy)
		return 0
	})
	luaRegister(l, "replayRecord", func(*lua.LState) int {
		/*Start recording rollback/netplay input to a file.
		@function replayRecord
		@tparam string path Output file path.
		function replayRecord(path) end*/
		if sys.netConnection != nil {
			sys.netConnection.recording, _ = os.Create(strArg(l, 1))
			sys.netConnection.headerWritten = false
		}
		return 0
	})
	luaRegister(l, "replayStop", func(*lua.LState) int {
		/*Stop input replay recording.
		@function replayStop
		function replayStop() end*/
		if sys.cfg.Netplay.RollbackNetcode {
			if sys.rollback.session != nil && sys.rollback.session.recording != nil {
				sys.rollback.session.recording.Close()
				sys.rollback.session.recording = nil
			}
		} else {
			if sys.netConnection != nil && sys.netConnection.recording != nil {
				sys.netConnection.recording.Close()
				sys.netConnection.recording = nil
				sys.netConnection.headerWritten = false
			}
		}
		return 0
	})
	luaRegister(l, "resetAILevel", func(l *lua.LState) int {
		/*Reset AI level for all players to 0 (human control).
		@function resetAILevel
		function resetAILevel() end*/
		for i := range sys.aiLevel {
			sys.aiLevel[i] = 0
		}
		return 0
	})
	luaRegister(l, "resetGameStats", func(*lua.LState) int {
		/*Clear all accumulated game statistics.
		@function resetGameStats
		function resetGameStats() end*/
		sys.statsLog.reset()
		sys.continueFlg = false
		sys.persistRoundCount = 0
		return 0
	})
	luaRegister(l, "resetKey", func(*lua.LState) int {
		/*Clear the last captured key and text input.
		@function resetKey
		function resetKey() end*/
		sys.keyInput = KeyUnknown
		sys.keyString = ""
		return 0
	})
	luaRegister(l, "resetMatchData", func(*lua.LState) int {
		/*Reset match-related runtime data.
		@function resetMatchData
		@tparam boolean fullReset If `true`, perform a full match data reset.
		function resetMatchData(fullReset) end*/
		sys.resetMatchData(boolArg(l, 1))
		return 0
	})
	luaRegister(l, "resetRemapInput", func(l *lua.LState) int {
		/*Reset all input remapping to defaults.
		@function resetRemapInput
		function resetRemapInput() end*/
		sys.resetRemapInput()
		return 0
	})
	luaRegister(l, "resetRound", func(*lua.LState) int {
		/*Request a round reset.
		@function resetRound
		function resetRound() end*/
		sys.roundResetFlg = true
		sys.roundResetMatchStart = true
		return 0
	})
	luaRegister(l, "resetScore", func(*lua.LState) int {
		/*Reset a team's score to zero.
		@function resetScore
		@tparam int teamSide Team side (`1` or `2`).
		function resetScore(teamSide) end*/
		tn := int(numArg(l, 1))
		if tn < 1 || tn > 2 {
			l.RaiseError("\nInvalid team side: %v\n", tn)
		}
		sys.fightScreen.scores[tn-1].scorePoints = 0
		return 0
	})
	luaRegister(l, "resetTokenGuard", func(*lua.LState) int {
		/*Reset the UI input token guard.
		@function resetTokenGuard
		function resetTokenGuard() end*/
		sys.uiResetTokenGuard()
		return 0
	})
	luaRegister(l, "roundOver", func(*lua.LState) int {
		/*Check whether the current round is over.
		@function roundOver
		@treturn boolean over `true` if the current round is over.
		function roundOver() end*/
		l.Push(lua.LBool(sys.roundOver()))
		return 1
	})
	luaRegister(l, "roundStart", func(*lua.LState) int {
		/*Check whether the current frame is the start of the round.
		@function roundStart
		@treturn boolean start `true` on the first tick of the round.
		function roundStart() end*/
		l.Push(lua.LBool(sys.tickCount == 1))
		return 1
	})
	luaRegister(l, "runHiscore", func(*lua.LState) int {
		/*Run the high-score screen for one frame.
		@function runHiscore
		@tparam[opt] string mode Optional hiscore mode identifier.
		@tparam[opt] int32 place Optional ranking position to highlight.
		@tparam[opt] int32 endtime Optional override for the hiscore screen duration.
		@tparam[opt=false] boolean nofade If `true`, disable fade-in and fade-out effects.
		@tparam[opt=false] boolean nobgs If `true`, disable background rendering.
		@tparam[opt=false] boolean nooverlay If `true`, disable overlay rendering.
		@treturn boolean active `true` while the hiscore screen is active.
		function runHiscore(mode, place, endtime, nofade, nobgs, nooverlay) end*/
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
	luaRegister(l, "runStoryboard", func(*lua.LState) int {
		/*Run the currently loaded storyboard for one frame.
		@function runStoryboard
		@treturn boolean active `true` while the storyboard is active.
		function runStoryboard() end*/
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
	luaRegister(l, "saveGameOption", func(l *lua.LState) int {
		/*Save current game options to file.
		@function saveGameOption
		@tparam[opt] string path Config file path. Defaults to the current config's `Def` path.
		function saveGameOption(path) end*/
		path := sys.cfg.Def
		if !nilArg(l, 1) {
			path = strArg(l, 1)
		}
		if err := sys.cfg.Save(path); err != nil {
			l.RaiseError("\nsaveGameOption: %v\n", err.Error())
		}
		return 0
	})
	luaRegister(l, "saveState", func(*lua.LState) int {
		/*Request saving of the current state on the next frame.
		@function saveState
		function saveState() end*/
		sys.saveStateFlag = true
		return 0
	})
	luaRegister(l, "screenshot", func(*lua.LState) int {
		/*Take a screenshot on the next frame.
		@function screenshot
		function screenshot() end*/
		if !sys.isTakingScreenshot {
			sys.isTakingScreenshot = true
		}
		return 0
	})
	luaRegister(l, "searchFile", func(l *lua.LState) int {
		/*Search for a file in a list of directories.
		@function searchFile
		@tparam string filename Filename to search for.
		@tparam table dirs Array-like table of directory paths (string).
		@treturn string path Resolved file path, or empty string if not found.
		function searchFile(filename, dirs) end*/
		var dirs []string
		tableArg(l, 2).ForEach(func(key, value lua.LValue) {
			dirs = append(dirs, lua.LVAsString(value))
		})
		l.Push(lua.LString(SearchFile(strArg(l, 1), dirs)))
		return 1
	})
	luaRegister(l, "selectChar", func(*lua.LState) int {
		/*Add a character to a team's selection.
		@function selectChar
		@tparam int teamSide Team side (`1` or `2`).
		@tparam int charRef 0-based character index in the select list.
		@tparam int palette Palette number.
		@treturn int status Selection status:
		  - `0` – character not added
		  - `1` – added, team is not yet full
		  - `2` – added, team is now full
		function selectChar(teamSide, charRef, palette) end*/
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
		/*Select a stage by index.
		@function selectStage
		@tparam int stageRef Stage reference used by the select system.
		  `0` selects the random stage sentinel; positive values are 1-based stage slots.
		function selectStage(stageRef) end*/
		sn := int(numArg(l, 1))
		if sn < 0 || sn > len(sys.sel.stagelist) {
			l.RaiseError("\nInvalid stage ref: %v\n", sn)
		}
		sys.sel.SelectStage(sn)
		return 0
	})
	luaRegister(l, "selectStart", func(l *lua.LState) int {
		/*Clear current selection and start loading the match.
		@function selectStart
		function selectStart() end*/
		sys.sel.ClearSelected()
		sys.loadStart()
		return 0
	})
	luaRegister(l, "selfState", func(*lua.LState) int {
		/*[redirectable] Force the character into a specified state.
		@function selfState
		@tparam int32 stateNo Target state number.
		function selfState(stateNo) end*/
		sys.debugWC.selfState(int32(numArg(l, 1)), -1, -1, 1, "")
		return 0
	})
	luaRegister(l, "setAccel", func(*lua.LState) int {
		/*Set debug time acceleration.
		@function setAccel
		@tparam float32 accel Time acceleration multiplier.
		function setAccel(accel) end*/
		sys.debugAccel = float32(numArg(l, 1))
		return 0
	})
	luaRegister(l, "setAILevel", func(*lua.LState) int {
		/*[redirectable] Set the character's AI level.
		@function setAILevel
		@tparam float32 level AI level (`0` = human control, >0 = AI).
		function setAILevel(level) end*/
		sys.debugWC.setAILevel(Clamp(float32(numArg(l, 1)), 0, 8))
		return 0
	})
	luaRegister(l, "setCom", func(*lua.LState) int {
		/*Set AI level for a specific player.
		@function setCom
		@tparam int playerNo Player number (1-based).
		@tparam float32 level AI level (`0` = off, >0 = AI).
		function setCom(playerNo, level) end*/
		pn := int(numArg(l, 1))
		if pn < 1 || pn > MaxPlayerNo {
			l.RaiseError("\nInvalid player number: %v\n", pn)
		}
		sys.aiLevel[pn-1] = Clamp(float32(numArg(l, 2)), 0, 8)
		return 0
	})
	luaRegister(l, "setConsecutiveWins", func(l *lua.LState) int {
		/*Set the number of consecutive wins for a team.
		@function setConsecutiveWins
		@tparam int teamSide Team side (`1` or `2`).
		@tparam int32 wins Number of consecutive wins.
		function setConsecutiveWins(teamSide, wins) end*/
		tn := int(numArg(l, 1))
		if tn < 1 || tn > 2 {
			l.RaiseError("\nInvalid team side: %v\n", tn)
		}
		sys.consecutiveWins[tn-1] = int32(numArg(l, 2))
		return 0
	})
	luaRegister(l, "setCredits", func(*lua.LState) int {
		/*Set the number of credits.
		@function setCredits
		@tparam int32 credits Credit count.
		function setCredits(credits) end*/
		sys.credits = int32(numArg(l, 1))
		return 0
	})
	luaRegister(l, "setDefaultConfig", func(l *lua.LState) int {
		/*Apply default key or joystick bindings for a player.
		@function setDefaultConfig
		@tparam string configType Configuration type: `"Keys"` or `"Joystick"`.
		@tparam int playerNo Player number (1-based).
		@tparam[opt] table enabled Optional table limiting which bindings are set.
		  Can be either:
		  - an array-like table of binding names, or
		  - a map-like table `{bindingName = true, ...}`.
		function setDefaultConfig(configType, playerNo, enabled) end*/
		cfgType := strArg(l, 1)
		pn := int(numArg(l, 2))
		var enabled map[string]bool
		if !nilArg(l, 3) {
			enabled = make(map[string]bool, 16)
			tableArg(l, 3).ForEach(func(k, v lua.LValue) {
				// map-style: enabled["up"]=true
				if ks, ok := k.(lua.LString); ok {
					if lua.LVAsBool(v) {
						enabled[string(ks)] = true
					}
					return
				}
				// array-style: {"up","down",...}
				if vs, ok := v.(lua.LString); ok {
					enabled[string(vs)] = true
				}
			})
		}
		switch cfgType {
		case "Keys":
			sys.uiSetConfigDefaults(pn, false, enabled)
		case "Joystick":
			sys.uiSetConfigDefaults(pn, true, enabled)
		default:
			l.RaiseError("\nInvalid config type: %v\n", cfgType)
		}
		return 0
	})
	luaRegister(l, "setDizzyPoints", func(*lua.LState) int {
		/*[redirectable] Set the character's dizzy points.
		@function setDizzyPoints
		@tparam int32 value Dizzy points value.
		function setDizzyPoints(value) end*/
		sys.debugWC.dizzyPointsSet(int32(numArg(l, 1)))
		return 0
	})
	luaRegister(l, "setGameMode", func(*lua.LState) int {
		/*Set current game mode identifier.
		@function setGameMode
		@tparam string mode Game mode name (for example `"arcade"`, `"versus"`, `"training"`).
		function setGameMode(mode) end*/
		sys.gameMode = strArg(l, 1)
		return 0
	})
	luaRegister(l, "setGameSpeed", func(*lua.LState) int {
		/*Set global game speed option.
		@function setGameSpeed
		@tparam int speed Game speed value (engine-specific range).
		function setGameSpeed(speed) end*/
		sys.cfg.Options.GameSpeed = int(numArg(l, 1))
		return 0
	})
	luaRegister(l, "setGameStatsJson", func(l *lua.LState) int {
		/*Restore accumulated game statistics from a JSON snapshot.
		@function setGameStatsJson
		@tparam string json JSON string previously produced by `getGameStatsJson()`.
		function setGameStatsJson(json) end*/
		var s GameStatsSnapshot
		if err := json.Unmarshal([]byte(strArg(l, 1)), &s); err != nil {
			l.RaiseError("setGameStatsJson: %v", err)
			return 0
		}
		sys.statsLog = s.StatsLog
		sys.continueFlg = s.ContinueFlg
		sys.persistRoundCount = s.PersistRoundCount
		return 0
	})
	luaRegister(l, "setGuardPoints", func(*lua.LState) int {
		/*[redirectable] Set the character's guard points.
		@function setGuardPoints
		@tparam int32 value Guard points value.
		function setGuardPoints(value) end*/
		sys.debugWC.guardPointsSet(int32(numArg(l, 1)))
		return 0
	})
	luaRegister(l, "setHomeTeam", func(l *lua.LState) int {
		/*Set which team is the home team.
		@function setHomeTeam
		@tparam int teamSide Team side (`1` or `2`).
		function setHomeTeam(teamSide) end*/
		tn := int(numArg(l, 1))
		if tn < 1 || tn > 2 {
			l.RaiseError("\nInvalid team side: %v\n", tn)
		}
		sys.home = tn - 1
		return 0
	})
	luaRegister(l, "setKeyConfig", func(l *lua.LState) int {
		/*Configure keyboard or joystick bindings for a player.
		@function setKeyConfig
		@tparam int playerNo Player number (1-based).
		@tparam int controllerId Input config target selector: `-1` updates keyboard bindings,
		  any value `>= 0` updates joystick bindings.
		@tparam table mapping Table mapping button indices (`1`–`14`) to key/button names (string).
		function setKeyConfig(playerNo, controllerId, mapping) end*/
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
		sys.uiResetTokenGuard()
		return 0
	})
	luaRegister(l, "setLastInputController", func(l *lua.LState) int {
		/*Set the last UI input controller.
		@function setLastInputController
		@tparam int playerNo 1-based player/controller index. Values less than `1` clear it.
		function setLastInputController(playerNo) end*/
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
		/*[redirectable] Set the character's life.
		@function setLife
		@tparam int32 life New life value (only applied if the character is alive).
		function setLife(life) end*/
		if sys.debugWC.alive() {
			sys.debugWC.lifeSet(int32(numArg(l, 1)))
		}
		return 0
	})
	luaRegister(l, "setFightScreenElements", func(*lua.LState) int {
		/*Force enable/disable of fight screen elements.
		@function setFightScreenElements
		@tparam table elements Table of boolean flags (keys are case-insensitive):
		  - `active` (boolean) enable lifebar drawing
		  - `bars` (boolean) main life bars
		  - `guardbar` (boolean) guard bar
		  - `hidebars` (boolean) hide bars during dialogue
		  - `match` (boolean) match info
		  - `mode` (boolean) mode display
		  - `p1ailevel`, `p2ailevel` (boolean) AI level displays
		  - `p1score`, `p2score` (boolean) score displays
		  - `p1wincount`, `p2wincount` (boolean) win count displays
		  - `redlifebar` (boolean) red life bar
		  - `stunbar` (boolean) stun bar
		  - `timer` (boolean) round timer
		function setFightScreenElements(elements) end*/
		tableArg(l, 1).ForEach(func(key, value lua.LValue) {
			switch k := key.(type) {
			case lua.LString:
				switch strings.ToLower(string(k)) {
				case "active": // enabled by default
					sys.fightScreen.active = lua.LVAsBool(value)
				case "bars": // enabled by default
					sys.fightScreen.bars = lua.LVAsBool(value)
				case "guardbar": // enabled depending on config.ini
					sys.fightScreen.guardbar = lua.LVAsBool(value)
				case "hidebars": // enabled depending on [Dialogue Info] motif settings
					sys.fightScreen.hidebars = lua.LVAsBool(value)
				case "match":
					sys.fightScreen.match.active = lua.LVAsBool(value)
				case "mode": // enabled by default
					sys.fightScreen.mode = lua.LVAsBool(value)
				case "p1ailevel":
					sys.fightScreen.aiLevels[0].active = lua.LVAsBool(value)
				case "p1score":
					sys.fightScreen.scores[0].active = lua.LVAsBool(value)
				case "p1wincount":
					sys.fightScreen.winCounts[0].active = lua.LVAsBool(value)
				case "p2ailevel":
					sys.fightScreen.aiLevels[1].active = lua.LVAsBool(value)
				case "p2score":
					sys.fightScreen.scores[1].active = lua.LVAsBool(value)
				case "p2wincount":
					sys.fightScreen.winCounts[1].active = lua.LVAsBool(value)
				case "redlifebar": // enabled depending on config.ini
					sys.fightScreen.redlifebar = lua.LVAsBool(value)
				case "stunbar": // enabled depending on config.ini
					sys.fightScreen.stunbar = lua.LVAsBool(value)
				case "timer":
					sys.fightScreen.timer.active = lua.LVAsBool(value)
				default:
					l.RaiseError("\nInvalid table key: %v\n", k)
				}
			default:
				l.RaiseError("\nInvalid table key type: %v\n", fmt.Sprintf("%T\n", key))
			}
		})
		// elements enabled via fight.def, depending on game mode
		if _, ok := sys.fightScreen.match.enabled[sys.gameMode]; ok {
			sys.fightScreen.match.active = sys.fightScreen.match.enabled[sys.gameMode]
		}
		for _, v := range sys.fightScreen.aiLevels {
			if _, ok := v.enabled[sys.gameMode]; ok {
				v.active = v.enabled[sys.gameMode]
			}
		}
		for _, v := range sys.fightScreen.scores {
			if _, ok := v.enabled[sys.gameMode]; ok {
				v.active = v.enabled[sys.gameMode]
			}
		}
		for _, v := range sys.fightScreen.winCounts {
			if _, ok := v.enabled[sys.gameMode]; ok {
				v.active = v.enabled[sys.gameMode]
			}
		}
		if _, ok := sys.fightScreen.timer.enabled[sys.gameMode]; ok {
			sys.fightScreen.timer.active = sys.fightScreen.timer.enabled[sys.gameMode]
		}
		return 0
	})
	luaRegister(l, "setFightScreenScore", func(*lua.LState) int {
		/*Set initial fight screen scores for both teams.
		@function setFightScreenScore
		@tparam float32 p1Score Starting score for team 1.
		@tparam[opt] float32 p2Score Starting score for team 2 (defaults to 0 if omitted).
		function setFightScreenScore(p1Score, p2Score) end*/
		sys.scoreStart[0] = float32(numArg(l, 1))
		if !nilArg(l, 2) {
			sys.scoreStart[1] = float32(numArg(l, 2))
		}
		return 0
	})
	luaRegister(l, "setFightScreenTimer", func(*lua.LState) int {
		/*Set initial round timer value displayed on the fight screen.
		@function setFightScreenTimer
		@tparam int32 time Initial timer value.
		function setFightScreenTimer(time) end*/
		sys.timerStart = int32(numArg(l, 1))
		return 0
	})
	luaRegister(l, "setMatchMaxDrawGames", func(l *lua.LState) int {
		/*Set maximum number of draw games allowed for a team.
		@function setMatchMaxDrawGames
		@tparam int teamSide Team side (`1` or `2`).
		@tparam int32 count Maximum draw games.
		function setMatchMaxDrawGames(teamSide, count) end*/
		tn := int(numArg(l, 1))
		if tn < 1 || tn > 2 {
			l.RaiseError("\nInvalid team side: %v\n", tn)
		}
		sys.maxDraws[tn-1] = int32(numArg(l, 2))
		return 0
	})
	luaRegister(l, "setMatchNo", func(l *lua.LState) int {
		/*Set the current match number.
		@function setMatchNo
		@tparam int32 matchNo Match index/number.
		function setMatchNo(matchNo) end*/
		sys.match = int32(numArg(l, 1))
		return 0
	})
	luaRegister(l, "setMatchWins", func(l *lua.LState) int {
		/*Set number of round wins required to win the match for a team.
		@function setMatchWins
		@tparam int teamSide Team side (`1` or `2`).
		@tparam int32 wins Required wins.
		function setMatchWins(teamSide, wins) end*/
		tn := int(numArg(l, 1))
		if tn < 1 || tn > 2 {
			l.RaiseError("\nInvalid team side: %v\n", tn)
		}
		sys.matchWins[tn-1] = int32(numArg(l, 2))
		return 0
	})
	luaRegister(l, "setMotifElements", func(*lua.LState) int {
		/*Enable/disable major motif elements.
		@function setMotifElements
		@tparam table elements Table of boolean flags (keys are case-insensitive):
		  - `challenger` (boolean) challenger screen
		  - `continuescreen` (boolean) continue screen
		  - `demo` (boolean) demo/attract mode
		  - `dialogue` (boolean) dialogue system
		  - `hiscore` (boolean) hiscore screen
		  - `losescreen` (boolean) lose screen
		  - `vsscreen` (boolean) versus screen
		  - `vsmatchno` (boolean) versus screen match number
		  - `victoryscreen` (boolean) victory screen
		  - `winscreen` (boolean) win screen
		  - `menu` (boolean) main menu
		function setMotifElements(elements) end*/
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
				case "vsscreen":
				case "vsmatchno":
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
		/*Resize player input configuration data to match `config.Players`.
		@function setPlayers
		function setPlayers() end*/
		total := sys.cfg.Config.Players
		if err := sys.uiEnsureCommandLists(total); err != nil {
			l.RaiseError("\nuiEnsureCommandLists: %v\n", err.Error())
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
		// Defaults for newly added players
		_, noJoy := sys.cmdFlags["-nojoy"]
		for pn := 1; pn <= total; pn++ {
			// Keyboard: apply defaults only if Keys_Pn section is missing.
			if kp, ok := sys.cfg.Keys[fmt.Sprintf("keys_p%d", pn)]; !ok || kp == nil {
				sys.uiSetConfigDefaults(pn, false, nil)
			}
			if !noJoy {
				if jp, ok := sys.cfg.Joystick[fmt.Sprintf("joystick_p%d", pn)]; !ok || jp == nil {
					sys.uiSetConfigDefaults(pn, true, nil)
				}
			}
		}
		sys.uiResetTokenGuard()
		return 0
	})
	luaRegister(l, "setPower", func(*lua.LState) int {
		/*[redirectable] Set the character's power.
		@function setPower
		@tparam int32 power Power value.
		function setPower(power) end*/
		sys.debugWC.setPower(int32(numArg(l, 1)))
		return 0
	})
	luaRegister(l, "setRedLife", func(*lua.LState) int {
		/*[redirectable] Set the character's red life.
		@function setRedLife
		@tparam int32 value Red life value.
		function setRedLife(value) end*/
		sys.debugWC.redLifeSet(int32(numArg(l, 1)))
		return 0
	})
	luaRegister(l, "setRoundTime", func(l *lua.LState) int {
		/*Set maximum round time (in ticks/counts).
		@function setRoundTime
		@tparam int32 time Maximum round time.
		function setRoundTime(time) end*/
		t := int32(numArg(l, 1))
		// Since legacy mode rounds down the timer, we must add an offset just under one count to compensate
		// This is also how Mugen handles it
		if t > 0 && sys.cfg.Config.LegacyTime {
			t += sys.curFramesPerCount - 1
		}
		sys.maxRoundTime = t
		return 0
	})
	luaRegister(l, "setTeamMode", func(*lua.LState) int {
		/*Configure a team's mode and team size.
		@function setTeamMode
		@tparam int teamSide Team side (`1` or `2`).
		@tparam int32 mode Team mode (for example `TM_Single`, `TM_Simul`, `TM_Turns`, `TM_Tag`).
		@tparam int32 teamSize Number of members (for non-turns: `1..MaxSimul`, for turns: `>=1`).
		function setTeamMode(teamSide, mode, teamSize) end*/
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
		/*Set the current round time value.
		@function setTime
		@tparam int32 time Current timer value.
		function setTime(time) end*/
		sys.curRoundTime = int32(numArg(l, 1))
		return 0
	})
	luaRegister(l, "setTimeFramesPerCount", func(l *lua.LState) int {
		/*Set how many frames correspond to one timer count.
		@function setTimeFramesPerCount
		@tparam int32 frames Frames per timer count.
		function setTimeFramesPerCount(frames) end*/
		sys.curFramesPerCount = int32(numArg(l, 1))
		return 0
	})
	luaRegister(l, "setWinCount", func(*lua.LState) int {
		/*Set win count for a team.
		@function setWinCount
		@tparam int teamSide Team side (`1` or `2`).
		@tparam int32 wins Win count.
		function setWinCount(teamSide, wins) end*/
		tn := int(numArg(l, 1))
		if tn < 1 || tn > 2 {
			l.RaiseError("\nInvalid team side: %v\n", tn)
		}
		sys.fightScreen.winCounts[tn-1].wins = int32(numArg(l, 2))
		return 0
	})
	//luaRegister(l, "sffCacheDelete", func(l *lua.LState) int {
	//	removeSFFCache(strArg(l, 1))
	//	return 0
	//})
	luaRegister(l, "sffNew", func(l *lua.LState) int {
		/*Load an SFF file or create an empty SFF.
		@function sffNew
		@tparam[opt] string filename SFF file path. If omitted, an empty SFF is created.
		@tparam[opt=false] boolean isActPal If `true`, prepare SFFv1 to receive ACT palettes.
		@treturn Sff sff SFF userdata.
		function sffNew(filename, isActPal) end*/
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
	luaRegister(l, "shutdown", func(*lua.LState) int {
		/*Check whether shutdown has been requested.
		@function shutdown
		@treturn boolean shutdown `true` if the global shutdown flag is set.
		function shutdown() end*/
		l.Push(lua.LBool(sys.gameEnd))
		return 1
	})
	luaRegister(l, "sleep", func(l *lua.LState) int {
		/*Block the current script for a number of seconds.
		@function sleep
		@tparam number seconds Time to sleep, in seconds.
		function sleep(seconds) end*/
		time.Sleep(time.Duration((numArg(l, 1))) * time.Second)
		return 0
	})
	luaRegister(l, "sndNew", func(l *lua.LState) int {
		/*Load a SND file.
		@function sndNew
		@tparam string filename SND file path.
		@treturn Snd snd SND userdata.
		function sndNew(filename) end*/
		snd, err := LoadSnd(strArg(l, 1))
		if err != nil {
			l.RaiseError("\nCan't load %v: %v\n", strArg(l, 1), err.Error())
		}
		l.Push(newUserData(l, snd))
		return 1
	})
	luaRegister(l, "sndPlay", func(l *lua.LState) int {
		/*Play a sound from a SND object.
		@function sndPlay
		@tparam Snd snd SND userdata.
		@tparam int32 group Sound group number.
		@tparam int32 number Sound number within the group.
		@tparam[opt=100] int32 volumescale Volume scale (percent).
		@tparam[opt=0.0] float32 pan Stereo panning (engine-specific range).
		@tparam[opt=0] int loopstart Loop start position.
		@tparam[opt=0] int loopend Loop end position.
		@tparam[opt=0] int startposition Start position.
		function sndPlay(snd, group, number, volumescale, pan, loopstart, loopend, startposition) end*/
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
		/*Check if a given sound is currently playing.
		@function sndPlaying
		@tparam Snd snd SND userdata.
		@tparam int32 group Sound group number.
		@tparam int32 number Sound number within the group.
		@treturn boolean playing `true` if the sound is playing.
		function sndPlaying(snd, group, number) end*/
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
		/*Stop a sound from a SND object.
		@function sndStop
		@tparam Snd snd SND userdata.
		@tparam int32 group Sound group number.
		@tparam int32 number Sound number within the group.
		function sndStop(snd, group, number) end*/
		s, ok := toUserData(l, 1).(*Snd)
		if !ok {
			userDataError(l, 1, s)
		}
		s.stop([...]int32{int32(numArg(l, 2)), int32(numArg(l, 3))})
		return 0
	})
	luaRegister(l, "stopAllCharSounds", func(l *lua.LState) int {
		/*Stop all character sounds.
		@function stopAllCharSounds
		function stopAllCharSounds() end*/
		sys.stopAllCharSounds()
		return 0
	})
	luaRegister(l, "stopBgm", func(l *lua.LState) int {
		/*Stop background music playback.
		@function stopBgm
		function stopBgm() end*/
		sys.bgm.Stop()
		return 0
	})
	luaRegister(l, "stopSnd", func(l *lua.LState) int {
		/*[redirectable] Stop all character's sounds.
		@function stopSnd
		function stopSnd() end*/
		sys.charSoundChannels[sys.debugWC.playerNo].SetSize(0) // TODO: Why does this use the hard reset?
		return 0
	})
	luaRegister(l, "synchronize", func(*lua.LState) int {
		/*Synchronize with external systems (e.g. netplay).
		@function synchronize
		@treturn boolean success `true` if synchronization succeeded, `false` if a
		  non-fatal session warning occurred.
		function synchronize() end*/
		if err := sys.synchronize(); err != nil {
			if sys.sessionWarning != "" {
				l.Push(lua.LBool(false))
				return 1
			}
			l.RaiseError(err.Error())
		}
		l.Push(lua.LBool(true))
		return 1
	})
	luaRegister(l, "textImgAddPos", func(*lua.LState) int {
		/*Offset a text sprite's position by the given amounts.
		@function textImgAddPos
		@tparam TextSprite ts Text sprite userdata.
		@tparam float32 dx X offset to add.
		@tparam float32 dy Y offset to add.
		function textImgAddPos(ts, dx, dy) end*/
		ts, ok := toUserData(l, 1).(*TextSprite)
		if !ok {
			userDataError(l, 1, ts)
		}
		ts.AddPos(float32(numArg(l, 2)), float32(numArg(l, 3)))
		return 0
	})
	luaRegister(l, "textImgAddText", func(*lua.LState) int {
		/*Append text to an existing text sprite.
		@function textImgAddText
		@tparam TextSprite ts Text sprite userdata.
		@tparam string text Text to append (no automatic newline).
		function textImgAddText(ts, text) end*/
		ts, ok := toUserData(l, 1).(*TextSprite)
		if !ok {
			userDataError(l, 1, ts)
		}
		ts.text = fmt.Sprintf(ts.text+"%v", strArg(l, 2))
		return 0
	})
	luaRegister(l, "textImgApplyVel", func(*lua.LState) int {
		/*Copy velocity settings from another text sprite.
		@function textImgApplyVel
		@tparam TextSprite ts Text sprite userdata to modify.
		@tparam TextSprite source Source text sprite whose velocity is copied.
		function textImgApplyVel(ts, source) end*/
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
		/*Print debug information about a text sprite.
		@function textImgDebug
		@tparam TextSprite ts Text sprite userdata.
		@tparam[opt] string prefix Optional text printed before the debug info.
		function textImgDebug(ts, prefix) end*/
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
		/*Queue drawing of a text sprite.
		@function textImgDraw
		@tparam TextSprite ts Text sprite userdata.
		@tparam[opt] int16 layer Layer to draw on (defaults to `ts.layerno`).
		function textImgDraw(ts, layer) end*/
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
		/*Measure the width of a text string for a font.
		@function textImgGetTextWidth
		@tparam TextSprite ts Text sprite userdata.
		@tparam string text Text to measure.
		@treturn int32 width Width of the rendered text in pixels.
		function textImgGetTextWidth(ts, text) end*/
		ts, ok := toUserData(l, 1).(*TextSprite)
		if !ok {
			userDataError(l, 1, ts)
		}
		l.Push(lua.LNumber(ts.TextWidth(strArg(l, 2))))
		return 1
	})
	luaRegister(l, "textImgNew", func(*lua.LState) int {
		/*Create a new empty text sprite.
		@function textImgNew
		@treturn TextSprite ts Newly created text sprite userdata.
		function textImgNew() end*/
		l.Push(newUserData(l, NewTextSprite()))
		return 1
	})
	luaRegister(l, "textImgReset", func(*lua.LState) int {
		/*Reset a text sprite to its initial values.
		@function textImgReset
		@tparam TextSprite ts Text sprite userdata.
		@tparam[opt] table parts If omitted or `nil`, resets everything.
		  If provided, must be an array-like table of strings, each one of:
		  - `"pos"` – reset position to initial
		  - `"scale"` – reset scale to initial
		  - `"window"` – reset window to initial
		  - `"velocity"` – reset velocity to initial
		  - `"text"` – reset text to initial
		  - `"palfx"` – clear PalFX
		  - `"delay"` – reset text delay timer
		function textImgReset(ts, parts) end*/
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
		/*Set per-frame acceleration for a text sprite.
		@function textImgSetAccel
		@tparam TextSprite ts Text sprite userdata.
		@tparam float32 ax X acceleration.
		@tparam float32 ay Y acceleration.
		function textImgSetAccel(ts, ax, ay) end*/
		ts, ok := toUserData(l, 1).(*TextSprite)
		if !ok {
			userDataError(l, 1, ts)
		}
		ts.SetAccel(float32(numArg(l, 2)), float32(numArg(l, 3)))
		return 0
	})
	luaRegister(l, "textImgSetAlign", func(*lua.LState) int {
		/*Set text alignment for a text sprite.
		@function textImgSetAlign
		@tparam TextSprite ts Text sprite userdata.
		@tparam int32 align Alignment value (engine-specific constants, e.g. left/center/right).
		function textImgSetAlign(ts, align) end*/
		ts, ok := toUserData(l, 1).(*TextSprite)
		if !ok {
			userDataError(l, 1, ts)
		}
		ts.align = int32(numArg(l, 2))
		return 0
	})
	luaRegister(l, "textImgSetAngle", func(*lua.LState) int {
		/*Set rotation angle for a text sprite.
		@function textImgSetAngle
		@tparam TextSprite ts Text sprite userdata.
		@tparam float32 angle Rotation angle in degrees.
		function textImgSetAngle(ts, angle) end*/
		ts, ok := toUserData(l, 1).(*TextSprite)
		if !ok {
			userDataError(l, 1, ts)
		}
		ts.rot.angle = float32(numArg(l, 2))
		return 0
	})
	luaRegister(l, "textImgSetBank", func(*lua.LState) int {
		/*Set the font bank index for a text sprite.
		@function textImgSetBank
		@tparam TextSprite ts Text sprite userdata.
		@tparam int32 bank Font bank index.
		function textImgSetBank(ts, bank) end*/
		ts, ok := toUserData(l, 1).(*TextSprite)
		if !ok {
			userDataError(l, 1, ts)
		}
		ts.bank = int32(numArg(l, 2))
		return 0
	})
	luaRegister(l, "textImgSetColor", func(*lua.LState) int {
		/*Set the RGBA color for a text sprite.
		@function textImgSetColor
		@tparam TextSprite ts Text sprite userdata.
		@tparam int32 r Red component (0–255).
		@tparam int32 g Green component (0–255).
		@tparam int32 b Blue component (0–255).
		@tparam[opt=255] int32 a Alpha component (0–255).
		function textImgSetColor(ts, r, g, b, a) end*/
		ts, ok := toUserData(l, 1).(*TextSprite)
		if !ok {
			userDataError(l, 1, ts)
		}
		// Default alpha to 255 for compatibility
		a := int32(255)
		if !nilArg(l, 5) {
			a = int32(Min(255, int(numArg(l, 5))))
		}
		ts.SetColor(int32(numArg(l, 2)), int32(numArg(l, 3)), int32(numArg(l, 4)), a)
		return 0
	})
	luaRegister(l, "textImgSetFocalLength", func(*lua.LState) int {
		/*Set focal length used for perspective projection on a text sprite.
		@function textImgSetFocalLength
		@tparam TextSprite ts Text sprite userdata.
		@tparam float32 fLength Focal length value.
		function textImgSetFocalLength(ts, fLength) end*/
		ts, ok := toUserData(l, 1).(*TextSprite)
		if !ok {
			userDataError(l, 1, ts)
		}
		ts.fLength = float32(numArg(l, 2))
		return 0
	})
	luaRegister(l, "textImgSetFont", func(*lua.LState) int {
		/*Assign a font object to a text sprite.
		@function textImgSetFont
		@tparam TextSprite ts Text sprite userdata.
		@tparam Fnt fnt Font userdata to use.
		function textImgSetFont(ts, fnt) end*/
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
	luaRegister(l, "textImgSetFriction", func(*lua.LState) int {
		/*Set friction applied to a text sprite's velocity each update.
		@function textImgSetFriction
		@tparam TextSprite ts Text sprite userdata.
		@tparam float32 fx X friction factor.
		@tparam float32 fy Y friction factor.
		function textImgSetFriction(ts, fx, fy) end*/
		ts, ok := toUserData(l, 1).(*TextSprite)
		if !ok {
			userDataError(l, 1, ts)
		}
		ts.friction[0] = float32(numArg(l, 2))
		ts.friction[1] = float32(numArg(l, 3))
		return 0
	})
	luaRegister(l, "textImgSetLayerno", func(*lua.LState) int {
		/*Set the drawing layer for a text sprite.
		@function textImgSetLayerno
		@tparam TextSprite ts Text sprite userdata.
		@tparam int16 layer Layer number.
		function textImgSetLayerno(ts, layer) end*/
		ts, ok := toUserData(l, 1).(*TextSprite)
		if !ok {
			userDataError(l, 1, ts)
		}
		ts.layerno = int16(numArg(l, 2))
		return 0
	})
	luaRegister(l, "textImgSetLocalcoord", func(*lua.LState) int {
		/*Set the local coordinate space for a text sprite.
		@function textImgSetLocalcoord
		@tparam TextSprite ts Text sprite userdata.
		@tparam float32 width Local coordinate width.
		@tparam float32 height Local coordinate height.
		function textImgSetLocalcoord(ts, width, height) end*/
		ts, ok := toUserData(l, 1).(*TextSprite)
		if !ok {
			userDataError(l, 1, ts)
		}
		ts.SetLocalcoord(int32(numArg(l, 2)), int32(numArg(l, 3)))
		return 0
	})
	luaRegister(l, "textImgSetMaxDist", func(*lua.LState) int {
		/*Set the maximum visible distance for a text sprite.
		@function textImgSetMaxDist
		@tparam TextSprite ts Text sprite userdata.
		@tparam float32 xDist Maximum X distance.
		@tparam float32 yDist Maximum Y distance.
		function textImgSetMaxDist(ts, xDist, yDist) end*/
		ts, ok := toUserData(l, 1).(*TextSprite)
		if !ok {
			userDataError(l, 1, ts)
		}
		ts.SetMaxDist(float32(numArg(l, 2)), float32(numArg(l, 3)))
		return 0
	})
	luaRegister(l, "textImgSetPos", func(*lua.LState) int {
		/*Set the position of a text sprite.
		@function textImgSetPos
		@tparam TextSprite ts Text sprite userdata.
		@tparam[opt] float32 x X position; if omitted, uses the initial X offset.
		@tparam[opt] float32 y Y position; if omitted, uses the initial Y offset.
		function textImgSetPos(ts, x, y) end*/
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
	luaRegister(l, "textImgSetProjection", func(*lua.LState) int {
		/*Set projection mode for a text sprite.
		@function textImgSetProjection
		@tparam TextSprite ts Text sprite userdata.
		@tparam int32|string projection Projection mode. Can be a numeric engine constant, or one of:
		  - `"orthographic"`
		  - `"perspective"`
		  - `"perspective2"`
		function textImgSetProjection(ts, projection) end*/
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
	luaRegister(l, "textImgSetScale", func(*lua.LState) int {
		/*Set the scale of a text sprite.
		@function textImgSetScale
		@tparam TextSprite ts Text sprite userdata.
		@tparam float32 sx X scale.
		@tparam float32 sy Y scale.
		function textImgSetScale(ts, sx, sy) end*/
		ts, ok := toUserData(l, 1).(*TextSprite)
		if !ok {
			userDataError(l, 1, ts)
		}
		ts.SetScale(float32(numArg(l, 2)), float32(numArg(l, 3)))
		return 0
	})
	luaRegister(l, "textImgSetText", func(*lua.LState) int {
		/*Set the text content of a text sprite.
		@function textImgSetText
		@tparam TextSprite ts Text sprite userdata.
		@tparam string text Text to display.
		function textImgSetText(ts, text) end*/
		ts, ok := toUserData(l, 1).(*TextSprite)
		if !ok {
			userDataError(l, 1, ts)
		}
		ts.text = strArg(l, 2)
		return 0
	})
	luaRegister(l, "textImgSetTextDelay", func(*lua.LState) int {
		/*Set per-character text delay for a text sprite.
		@function textImgSetTextDelay
		@tparam TextSprite ts Text sprite userdata.
		@tparam float32 delay Delay between characters (frames, engine-specific).
		function textImgSetTextDelay(ts, delay) end*/
		ts, ok := toUserData(l, 1).(*TextSprite)
		if !ok {
			userDataError(l, 1, ts)
		}
		ts.textDelay = float32(numArg(l, 2))
		return 0
	})
	luaRegister(l, "textImgSetTextSpacing", func(*lua.LState) int {
		/*Set text spacing for a text sprite.
		@function textImgSetTextSpacing
		@tparam TextSprite ts Text sprite userdata.
		@tparam float32 xSpacing Horizontal text spacing.
		@tparam float32 ySpacing Vertical text spacing.
		function textImgSetTextSpacing(ts, xSpacing, ySpacing) end*/
		ts, ok := toUserData(l, 1).(*TextSprite)
		if !ok {
			userDataError(l, 1, ts)
		}
		ts.SetTextSpacing(float32(numArg(l, 2)), float32(numArg(l, 3)))
		return 0
	})
	luaRegister(l, "textImgSetTextWrap", func(*lua.LState) int {
		/*Enable or disable word wrapping for a text sprite.
		@function textImgSetTextWrap
		@tparam TextSprite ts Text sprite userdata.
		@tparam boolean wrap If `true`, enables text wrapping.
		function textImgSetTextWrap(ts, wrap) end*/
		ts, ok := toUserData(l, 1).(*TextSprite)
		if !ok {
			userDataError(l, 1, ts)
		}
		ts.textWrap = boolArg(l, 2)
		return 0
	})
	luaRegister(l, "textImgSetVelocity", func(*lua.LState) int {
		/*Set velocity for a text sprite.
		@function textImgSetVelocity
		@tparam TextSprite ts Text sprite userdata.
		@tparam float32 vx X velocity.
		@tparam float32 vy Y velocity.
		function textImgSetVelocity(ts, vx, vy) end*/
		ts, ok := toUserData(l, 1).(*TextSprite)
		if !ok {
			userDataError(l, 1, ts)
		}
		ts.SetVelocity(float32(numArg(l, 2)), float32(numArg(l, 3)))
		return 0
	})
	luaRegister(l, "textImgSetWindow", func(*lua.LState) int {
		/*Set the clipping window for a text sprite.
		@function textImgSetWindow
		@tparam TextSprite ts Text sprite userdata.
		@tparam float32 x1 Left coordinate.
		@tparam float32 y1 Top coordinate.
		@tparam float32 x2 Right coordinate.
		@tparam float32 y2 Bottom coordinate.
		function textImgSetWindow(ts, x1, y1, x2, y2) end*/
		ts, ok := toUserData(l, 1).(*TextSprite)
		if !ok {
			userDataError(l, 1, ts)
		}
		ts.SetWindow([4]float32{float32(numArg(l, 2)), float32(numArg(l, 3)), float32(numArg(l, 4)), float32(numArg(l, 5))})
		return 0
	})
	luaRegister(l, "textImgSetXAngle", func(*lua.LState) int {
		/*Set rotation angle around the X axis for a text sprite.
		@function textImgSetXAngle
		@tparam TextSprite ts Text sprite userdata.
		@tparam float32 xangle X-axis rotation angle.
		function textImgSetXAngle(ts, xangle) end*/
		ts, ok := toUserData(l, 1).(*TextSprite)
		if !ok {
			userDataError(l, 1, ts)
		}
		ts.rot.xangle = float32(numArg(l, 2))
		return 0
	})
	luaRegister(l, "textImgSetXShear", func(*lua.LState) int {
		/*Set X shear (italic-style slant) for a text sprite.
		@function textImgSetXShear
		@tparam TextSprite ts Text sprite userdata.
		@tparam float32 xshear Shear value along X.
		function textImgSetXShear(ts, xshear) end*/
		ts, ok := toUserData(l, 1).(*TextSprite)
		if !ok {
			userDataError(l, 1, ts)
		}
		ts.xshear = float32(numArg(l, 2))
		return 0
	})
	luaRegister(l, "textImgSetYAngle", func(*lua.LState) int {
		/*Set rotation angle around the Y axis for a text sprite.
		@function textImgSetYAngle
		@tparam TextSprite ts Text sprite userdata.
		@tparam float32 yangle Y-axis rotation angle.
		function textImgSetYAngle(ts, yangle) end*/
		ts, ok := toUserData(l, 1).(*TextSprite)
		if !ok {
			userDataError(l, 1, ts)
		}
		ts.rot.yangle = float32(numArg(l, 2))
		return 0
	})

	luaRegister(l, "textImgUpdate", func(*lua.LState) int {
		/*Update a text sprite's internal state (position, delays, etc.).
		@function textImgUpdate
		@tparam TextSprite ts Text sprite userdata.
		function textImgUpdate(ts) end*/
		ts, ok := toUserData(l, 1).(*TextSprite)
		if !ok {
			userDataError(l, 1, ts)
		}
		ts.Update()
		return 0
	})
	luaRegister(l, "toggleClsnDisplay", func(*lua.LState) int {
		/*Toggle display of collision boxes.
		@function toggleClsnDisplay
		@tparam[opt] boolean state If provided, sets collision box display on/off; otherwise toggles it.
		function toggleClsnDisplay(state) end*/
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
		/*Toggle or cycle debug display.
		@function toggleDebugDisplay
		@tparam[opt] any mode If provided, simply toggles the debug display.
		  If omitted, cycles the debug display through characters and eventually disables it.
		function toggleDebugDisplay(dummy) end*/
		if !sys.debugModeAllowed() {
			return 0
		}
		// Shift+D behavior: just toggle without cycling players
		if !nilArg(l, 1) {
			sys.debugDisplay = !sys.debugDisplay
			return 0
		}
		// Ctrl+Shift+D behavior: cycle players in reverse order
		var reverse bool
		if !nilArg(l, 2) {
			reverse = boolArg(l, 2)
		}
		// Make a copy of creationOrder
		sorted := make([]*Char, len(sys.charList.creationOrder))
		copy(sorted, sys.charList.creationOrder)
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
			if reverse {
				// Search backwards
				for i := len(sorted) - 1; i >= 0; i-- {
					c := sorted[i]
					isEarlierPlayer := c.playerNo < sys.debugRef[0]
					isSamePlayerLowerID := c.playerNo == sys.debugRef[0] && c.id < sys.debugLastID
					if isEarlierPlayer || isSamePlayerLowerID {
						nextChar = c
						break
					}
				}
			} else {
				// Search for the first character that comes after the current one
				for _, c := range sorted {
					isLaterPlayer := c.playerNo > sys.debugRef[0]
					isSamePlayerHigherID := c.playerNo == sys.debugRef[0] && c.id > sys.debugLastID
					if isLaterPlayer || isSamePlayerHigherID {
						nextChar = c
						break
					}
				}
			}
		} else if len(sorted) > 0 {
			// If display was off, start at the beginning of the sorted list
			if reverse {
				nextChar = sorted[len(sorted)-1]
			} else {
				nextChar = sorted[0]
			}
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
		/*Toggle fullscreen mode.
		@function toggleFullscreen
		@tparam[opt] boolean state If provided, sets fullscreen on/off; otherwise toggles it.
		function toggleFullscreen(state) end*/
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
		/*Toggle lifebar visibility.
		@function toggleLifebarDisplay
		@tparam[opt] boolean hide If provided, hides (`true`) or shows (`false`) the lifebar; otherwise toggles.
		function toggleLifebarDisplay(hide) end*/
		if !nilArg(l, 1) {
			sys.lifebarHide = boolArg(l, 1)
		} else {
			sys.lifebarHide = !sys.lifebarHide
		}
		return 0
	})
	luaRegister(l, "toggleMaxPowerMode", func(*lua.LState) int {
		/*Toggle "max power" cheat mode.
		@function toggleMaxPowerMode
		@tparam[opt] boolean state If provided, sets max power mode on/off; otherwise toggles it.
		  When enabled, all root players' power is set to their maximum.
		function toggleMaxPowerMode(state) end*/
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
		/*Toggle global sound output.
		@function toggleNoSound
		@tparam[opt] boolean state If provided, sets mute on/off; otherwise toggles it.
		function toggleNoSound(state) end*/
		if !nilArg(l, 1) {
			sys.noSoundFlg = boolArg(l, 1)
		} else {
			sys.noSoundFlg = !sys.noSoundFlg
		}
		return 0
	})
	luaRegister(l, "togglePause", func(*lua.LState) int {
		/*Toggle game pause.
		@function togglePause
		@tparam[opt] boolean state If provided, sets pause on/off; otherwise toggles it.
		function togglePause(state) end*/
		if !nilArg(l, 1) {
			sys.paused = boolArg(l, 1)
		} else {
			sys.paused = !sys.paused
		}
		return 0
	})
	luaRegister(l, "togglePlayer", func(*lua.LState) int {
		/*Enable or disable all instances of a given player.
		@function togglePlayer
		@tparam int32 playerNo Player number (1-based).
		function togglePlayer(playerNo) end*/
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
		/*Toggle vertical sync (VSync).
		@function toggleVSync
		@tparam[opt] int mode If provided, sets the swap interval directly; otherwise toggles between `0` and `1`.
		function toggleVSync(mode) end*/
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
		/*Toggle wireframe rendering mode (debug only).
		@function toggleWireframeDisplay
		@tparam[opt] boolean state If provided, sets wireframe display on/off; otherwise toggles it.
		function toggleWireframeDisplay(state) end*/
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
		/*Update background music volume to match current settings.
		@function updateVolume
		function updateVolume() end*/
		sys.bgm.UpdateVolume()
		return 0
	})
	luaRegister(l, "validatePal", func(l *lua.LState) int {
		/*Validate a requested palette index for a character.
		@function validatePal
		@tparam int palReq Requested palette number (1-based).
		@tparam int charRef 0-based character index in the select list.
		@treturn int validPal Engine-validated palette number (may differ from `palReq`
		  depending on character configuration).
		function validatePal(palReq, charRef) end*/
		palReq := int(numArg(l, 1))
		charRef := int(numArg(l, 2))
		valid := sys.sel.ValidatePalette(charRef, palReq)
		l.Push(lua.LNumber(valid))
		return 1
	})
	luaRegister(l, "version", func(l *lua.LState) int {
		/*Get the engine version string.
		@function version
		@treturn string ver Engine version and build time.
		function version() end*/
		ver := fmt.Sprintf("%s - %s", Version, BuildTime)
		l.Push(lua.LString(ver))
		return 1
	})
	luaRegister(l, "waveNew", func(*lua.LState) int {
		/*Load a sound from an SND file using a group/sound pair.
		@function waveNew
		@tparam string path Path to the SND container.
		@tparam int32 group Group number in the SND.
		@tparam int32 sound Sound number in the SND.
		@tparam[opt=0] uint32 max Maximum scan limit passed to SND loading. If non-zero, loading stops after the first matching entry and also gives up after scanning that many entries without a match.
		@treturn Sound sound Sound userdata containing the loaded sound data.
		function waveNew(path, group, sound, maxLoops) end*/
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
	luaRegister(l, "wavePlay", func(l *lua.LState) int {
		/*Play a sound from a `Sound` object on the shared sound channel pool.
		@function wavePlay
		@tparam Sound s Sound userdata.
		@tparam[opt=0] int32 group Optional group number.
		@tparam[opt=0] int32 number Optional sound number within the group.
		function wavePlay(s, group, number) end*/
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

// Trigger redirection (equivalents of CNS/ZSS trigger redirections)
func triggerRedirection(l *lua.LState) {
	// Create a temporary dummy character to avoid possible nil checks
	sys.debugWC = newChar(0, 0)
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
	luaRegister(l, "enemyNear", func(*lua.LState) int {
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
	luaRegister(l, "helperIndex", func(*lua.LState) int {
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
	luaRegister(l, "p2", func(*lua.LState) int {
		ret := false
		if c := sys.debugWC.p2(); c != nil {
			sys.debugWC, ret = c, true
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
	luaRegister(l, "player", func(*lua.LState) int {
		pn := int(numArg(l, 1))
		ret := false
		if pn >= 1 && pn <= len(sys.chars) && len(sys.chars[pn-1]) > 0 {
			sys.debugWC, ret = sys.chars[pn-1][0], true
		}
		l.Push(lua.LBool(ret))
		return 1
	})
	luaRegister(l, "playerId", func(*lua.LState) int {
		ret := false
		// Script version doesn't log errors because debug mode uses it
		// TODO: Script redirects should either all log errors or none of them should
		if c := sys.debugWC.playerIDTrigger(int32(numArg(l, 1)), false); c != nil {
			sys.debugWC, ret = c, true
		}
		l.Push(lua.LBool(ret))
		return 1
	})
	luaRegister(l, "playerIndex", func(*lua.LState) int {
		ret := false
		if c := sys.debugWC.playerIndexTrigger(int32(numArg(l, 1))); c != nil {
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
	luaRegister(l, "stateOwner", func(*lua.LState) int {
		ret := false
		if c := sys.debugWC.stateOwner(); c != nil {
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
}

// Trigger functions (equivalents of CNS/ZSS triggers)
func triggerFunctions(l *lua.LState) {
	luaRegister(l, "aiLevel", func(*lua.LState) int {
		if !sys.debugWC.asf(ASF_noailevel) {
			l.Push(lua.LNumber(sys.debugWC.getAILevel()))
		} else {
			l.Push(lua.LNumber(0))
		}
		return 1
	})
	luaRegister(l, "airJumpCount", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.airJumpCount))
		return 1
	})
	luaRegister(l, "alive", func(*lua.LState) int {
		l.Push(lua.LBool(sys.debugWC.alive()))
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
	luaRegister(l, "angle", func(*lua.LState) int {
		if sys.debugWC.csf(CSF_angledraw) {
			l.Push(lua.LNumber(sys.debugWC.anglerot[0]))
		} else {
			l.Push(lua.LNumber(0))
		}
		return 1
	})
	luaRegister(l, "anim", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.animNo))
		return 1
	})
	// animElem (deprecated by animElemTime)
	luaRegister(l, "animElemNo", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.animElemNo(int32(numArg(l, 1)) - 1).ToI())) // Offset by 1 because animations step before scripts run
		return 1
	})
	luaRegister(l, "animElemTime", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.animElemTime(int32(numArg(l, 1))).ToI()) - 1) // Offset by 1 because animations step before scripts run
		return 1
	})
	luaRegister(l, "animElemVar", func(l *lua.LState) int {
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
	luaRegister(l, "animExist", func(*lua.LState) int {
		l.Push(lua.LBool(sys.debugWC.animExist(BytecodeInt(int32(numArg(l, 1)))).ToB()))
		return 1
	})
	luaRegister(l, "animLength", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.anim.totaltime))
		return 1
	})
	luaRegister(l, "animPlayerNo", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.animPN + 1))
		return 1
	})
	luaRegister(l, "animTime", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.animTime()))
		return 1
	})
	// atan2 (dedicated functionality already exists in Lua)
	luaRegister(l, "attack", func(*lua.LState) int {
		base := float32(sys.debugWC.gi().attackBase) * sys.debugWC.ocd().attackRatio / 100
		l.Push(lua.LNumber(base * sys.debugWC.attackMul[0] * 100))
		return 1
	})
	luaRegister(l, "attackMul", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.attackMul[0]))
		return 1
	})
	luaRegister(l, "authorName", func(*lua.LState) int {
		l.Push(lua.LString(sys.debugWC.gi().author))
		return 1
	})
	luaRegister(l, "backEdge", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.backEdge()))
		return 1
	})
	luaRegister(l, "backEdgeBodyDist", func(*lua.LState) int {
		l.Push(lua.LNumber(int32(sys.debugWC.backEdgeBodyDist())))
		return 1
	})
	luaRegister(l, "backEdgeDist", func(*lua.LState) int {
		l.Push(lua.LNumber(int32(sys.debugWC.backEdgeDist())))
		return 1
	})
	luaRegister(l, "bgmVar", func(*lua.LState) int {
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
	luaRegister(l, "botBoundBodyDist", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.botBoundBodyDist()))
		return 1
	})
	luaRegister(l, "botBoundDist", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.botBoundDist()))
		return 1
	})
	luaRegister(l, "bottomEdge", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.bottomEdge()))
		return 1
	})
	luaRegister(l, "cameraPosX", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.cam.Pos[0]))
		return 1
	})
	luaRegister(l, "cameraPosY", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.cam.Pos[1]))
		return 1
	})
	luaRegister(l, "cameraZoom", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.cam.Scale))
		return 1
	})
	luaRegister(l, "canRecover", func(*lua.LState) int {
		l.Push(lua.LBool(sys.debugWC.canRecover()))
		return 1
	})
	luaRegister(l, "clamp", func(*lua.LState) int {
		v1 := float32(numArg(l, 1))
		v2 := float32(numArg(l, 2))
		v3 := float32(numArg(l, 3))
		retv := Clamp(v1, v2, v3)
		l.Push(lua.LNumber(retv))
		return 1
	})
	luaRegister(l, "clsnOverlap", func(l *lua.LState) int {
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
	luaRegister(l, "clsnVar", func(l *lua.LState) int {
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
	luaRegister(l, "comboCount", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.comboCount()))
		return 1
	})
	luaRegister(l, "command", func(*lua.LState) int {
		l.Push(lua.LBool(sys.debugWC.commandByName(strArg(l, 1))))
		return 1
	})
	luaRegister(l, "consecutiveWins", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.consecutiveWins[sys.debugWC.teamside]))
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
	luaRegister(l, "const1080p", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.constp(1920, float32(numArg(l, 1))).ToF()))
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
	luaRegister(l, "debugMode", func(*lua.LState) int {
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
	luaRegister(l, "decisiveRound", func(*lua.LState) int {
		l.Push(lua.LBool(sys.decisiveRound[sys.debugWC.playerNo&1]))
		return 1
	})
	luaRegister(l, "defence", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.finalDefense * 100))
		return 1
	})
	luaRegister(l, "defenceMul", func(*lua.LState) int {
		l.Push(lua.LNumber(float32(sys.debugWC.finalDefense / float64(sys.debugWC.gi().defenceBase) * 100)))
		return 1
	})
	luaRegister(l, "displayName", func(*lua.LState) int {
		l.Push(lua.LString(sys.debugWC.gi().displayname))
		return 1
	})
	luaRegister(l, "dizzy", func(*lua.LState) int {
		l.Push(lua.LBool(sys.debugWC.scf(SCF_dizzy)))
		return 1
	})
	luaRegister(l, "dizzyPoints", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.dizzyPoints))
		return 1
	})
	luaRegister(l, "dizzyPointsMax", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.dizzyPointsMax))
		return 1
	})
	luaRegister(l, "drawGame", func(*lua.LState) int {
		l.Push(lua.LBool(sys.debugWC.drawgame()))
		return 1
	})
	luaRegister(l, "drawPal", func(*lua.LState) int {
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
	luaRegister(l, "envShakeVar", func(*lua.LState) int {
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
	luaRegister(l, "explodVar", func(*lua.LState) int {
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
			case "bindid":
				lv = lua.LNumber(e.bindId)
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
	luaRegister(l, "fightScreenState", func(*lua.LState) int {
		switch strings.ToLower(strArg(l, 1)) {
		case "fightdisplay":
			l.Push(lua.LBool(sys.fightScreen.round.fightDisplayPhase == 1))
		case "kodisplay":
			l.Push(lua.LBool(sys.fightScreen.round.koDisplayPhase == 1))
		case "rounddisplay":
			l.Push(lua.LBool(sys.fightScreen.round.roundDisplayPhase == 1))
		case "windisplay":
			l.Push(lua.LBool(sys.fightScreen.round.winDisplayPhase == 1))
		default:
			l.RaiseError("\nInvalid argument: %v\n", strArg(l, 1))
		}
		return 1
	})
	luaRegister(l, "fightScreenVar", func(*lua.LState) int {
		switch strings.ToLower(strArg(l, 1)) {
		case "info.name":
			l.Push(lua.LString(sys.fightScreen.name))
		case "info.localcoord.x":
			l.Push(lua.LNumber(sys.fightScreen.localcoord[0]))
		case "info.localcoord.y":
			l.Push(lua.LNumber(sys.fightScreen.localcoord[1]))
		case "info.author":
			l.Push(lua.LString(sys.fightScreen.author))
		case "round.ctrl.time":
			l.Push(lua.LNumber(sys.fightScreen.round.ctrl_time))
		case "round.over.hittime":
			l.Push(lua.LNumber(sys.fightScreen.round.over_hittime))
		case "round.over.time":
			l.Push(lua.LNumber(sys.fightScreen.round.over_time))
		case "round.over.waittime":
			l.Push(lua.LNumber(sys.fightScreen.round.over_waittime))
		case "round.over.wintime":
			l.Push(lua.LNumber(sys.fightScreen.round.over_wintime))
		case "round.slow.time":
			l.Push(lua.LNumber(sys.fightScreen.round.slow_time))
		case "round.start.waittime":
			l.Push(lua.LNumber(sys.fightScreen.round.start_waittime))
		case "round.callfight.time":
			l.Push(lua.LNumber(sys.fightScreen.round.callfight_time))
		case "time.framespercount":
			l.Push(lua.LNumber(sys.fightScreen.time.framespercount))
		default:
			l.RaiseError("\nInvalid argument: %v\n", strArg(l, 1))
		}
		return 1
	})
	luaRegister(l, "fightTime", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.matchTime))
		return 1
	})
	luaRegister(l, "firstAttack", func(*lua.LState) int {
		l.Push(lua.LBool(sys.firstAttack[sys.debugWC.teamside] == sys.debugWC.playerNo))
		return 1
	})
	luaRegister(l, "frontEdge", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.frontEdge()))
		return 1
	})
	luaRegister(l, "frontEdgeBodyDist", func(*lua.LState) int {
		l.Push(lua.LNumber(int32(sys.debugWC.frontEdgeBodyDist())))
		return 1
	})
	luaRegister(l, "frontEdgeDist", func(*lua.LState) int {
		l.Push(lua.LNumber(int32(sys.debugWC.frontEdgeDist())))
		return 1
	})
	luaRegister(l, "fvar", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.fvarGet(int32(numArg(l, 1))).ToF()))
		return 1
	})
	luaRegister(l, "gameHeight", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.gameHeight()))
		return 1
	})
	luaRegister(l, "gameMode", func(*lua.LState) int {
		if !nilArg(l, 1) {
			l.Push(lua.LBool(sys.gameMode == strArg(l, 1)))
		} else {
			l.Push(lua.LString(sys.gameMode))
		}
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
	luaRegister(l, "gameTime", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.gameTime()))
		return 1
	})
	luaRegister(l, "gameVar", func(*lua.LState) int {
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
	luaRegister(l, "gameWidth", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.gameWidth()))
		return 1
	})
	luaRegister(l, "getHitVar", func(*lua.LState) int {
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
		case "teamside":
			lv = lua.LNumber(c.ghv.teamside + 1)
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
	luaRegister(l, "groundLevel", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.groundLevel))
		return 1
	})
	luaRegister(l, "guardBreak", func(*lua.LState) int {
		l.Push(lua.LBool(sys.debugWC.scf(SCF_guardbreak)))
		return 1
	})
	luaRegister(l, "guardCount", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.guardCount))
		return 1
	})
	luaRegister(l, "guardPoints", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.guardPoints))
		return 1
	})
	luaRegister(l, "guardPointsMax", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.guardPointsMax))
		return 1
	})
	luaRegister(l, "helperIndexExist", func(*lua.LState) int {
		l.Push(lua.LBool(sys.debugWC.helperIndexExist(
			BytecodeInt(int32(numArg(l, 1)))).ToB()))
		return 1
	})
	luaRegister(l, "helperVar", func(l *lua.LState) int {
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
	luaRegister(l, "hitByAttr", func(*lua.LState) int {
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
	luaRegister(l, "hitCount", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.hitCount))
		return 1
	})
	luaRegister(l, "hitDefAttr", func(*lua.LState) int {
		if sys.debugWC.ss.moveType == MT_A {
			l.Push(attrLStr(sys.debugWC.hitdef.attr))
		} else {
			l.Push(lua.LString(""))
		}
		return 1
	})
	luaRegister(l, "hitDefVar", func(*lua.LState) int {
		c := sys.debugWC
		switch strings.ToLower(strArg(l, 1)) {
		case "guard.dist.width.back":
			l.Push(lua.LNumber(c.hitdef.guard_dist_x[1]))
		case "guard.dist.width.front":
			l.Push(lua.LNumber(c.hitdef.guard_dist_x[0]))
		case "guard.dist.height.bottom":
			l.Push(lua.LNumber(c.hitdef.guard_dist_y[1]))
		case "guard.dist.height.top":
			l.Push(lua.LNumber(c.hitdef.guard_dist_y[0]))
		case "guard.dist.depth.bottom":
			l.Push(lua.LNumber(c.hitdef.guard_dist_z[1]))
		case "guard.dist.depth.top":
			l.Push(lua.LNumber(c.hitdef.guard_dist_z[0]))
		case "guard.pausetime":
			l.Push(lua.LNumber(c.hitdef.guard_pausetime[0]))
		case "guard.shaketime":
			l.Push(lua.LNumber(c.hitdef.guard_pausetime[1]))
		case "guard.sparkno":
			l.Push(lua.LNumber(c.hitdef.guard_sparkno))
		case "guarddamage":
			l.Push(lua.LNumber(c.hitdef.guarddamage))
		case "guardflag":
			l.Push(flagLStr(c.hitdef.guardflag))
		case "guardsound.group":
			l.Push(lua.LNumber(c.hitdef.guardsound[0]))
		case "guardsound.number":
			l.Push(lua.LNumber(c.hitdef.guardsound[1]))
		case "hitdamage":
			l.Push(lua.LNumber(c.hitdef.hitdamage))
		case "hitflag":
			l.Push(flagLStr(c.hitdef.hitflag))
		case "hitsound.group":
			l.Push(lua.LNumber(c.hitdef.hitsound[0]))
		case "hitsound.number":
			l.Push(lua.LNumber(c.hitdef.hitsound[1]))
		case "id":
			l.Push(lua.LNumber(c.hitdef.id))
		case "p1stateno":
			l.Push(lua.LNumber(c.hitdef.p1stateno))
		case "p2stateno":
			l.Push(lua.LNumber(c.hitdef.p2stateno))
		case "pausetime":
			l.Push(lua.LNumber(c.hitdef.pausetime[0]))
		case "priority":
			l.Push(lua.LNumber(c.hitdef.priority))
		case "shaketime":
			l.Push(lua.LNumber(c.hitdef.pausetime[1]))
		case "sparkno":
			l.Push(lua.LNumber(c.hitdef.sparkno))
		case "sparkx":
			l.Push(lua.LNumber(c.hitdef.sparkxy[0]))
		case "sparky":
			l.Push(lua.LNumber(c.hitdef.sparkxy[1]))
		case "xaccel":
			l.Push(lua.LNumber(c.hitdef.xaccel))
		case "yaccel":
			l.Push(lua.LNumber(c.hitdef.yaccel))
		case "zaccel":
			l.Push(lua.LNumber(c.hitdef.zaccel))
		case "ground.velocity.x":
			l.Push(lua.LNumber(c.hitdef.ground_velocity[0]))
		case "ground.velocity.y":
			l.Push(lua.LNumber(c.hitdef.ground_velocity[1]))
		case "ground.velocity.z":
			l.Push(lua.LNumber(c.hitdef.ground_velocity[2]))
		case "air.velocity.x":
			l.Push(lua.LNumber(c.hitdef.air_velocity[0]))
		case "air.velocity.y":
			l.Push(lua.LNumber(c.hitdef.air_velocity[1]))
		case "air.velocity.z":
			l.Push(lua.LNumber(c.hitdef.air_velocity[2]))
		case "down.velocity.x":
			l.Push(lua.LNumber(c.hitdef.down_velocity[0]))
		case "down.velocity.y":
			l.Push(lua.LNumber(c.hitdef.down_velocity[1]))
		case "down.velocity.z":
			l.Push(lua.LNumber(c.hitdef.down_velocity[2]))
		case "guard.velocity.x":
			l.Push(lua.LNumber(c.hitdef.guard_velocity[0]))
		case "guard.velocity.y":
			l.Push(lua.LNumber(c.hitdef.guard_velocity[1]))
		case "guard.velocity.z":
			l.Push(lua.LNumber(c.hitdef.guard_velocity[2]))
		case "airguard.velocity.x":
			l.Push(lua.LNumber(c.hitdef.airguard_velocity[0]))
		case "airguard.velocity.y":
			l.Push(lua.LNumber(c.hitdef.airguard_velocity[1]))
		case "airguard.velocity.z":
			l.Push(lua.LNumber(c.hitdef.airguard_velocity[2]))
		case "ground.cornerpush.veloff":
			l.Push(lua.LNumber(c.hitdef.ground_cornerpush_veloff))
		case "air.cornerpush.veloff":
			l.Push(lua.LNumber(c.hitdef.air_cornerpush_veloff))
		case "down.cornerpush.veloff":
			l.Push(lua.LNumber(c.hitdef.down_cornerpush_veloff))
		case "guard.cornerpush.veloff":
			l.Push(lua.LNumber(c.hitdef.guard_cornerpush_veloff))
		case "airguard.cornerpush.veloff":
			l.Push(lua.LNumber(c.hitdef.airguard_cornerpush_veloff))
		case "fall.xvelocity":
			l.Push(lua.LNumber(c.hitdef.fall_xvelocity))
		case "fall.yvelocity":
			l.Push(lua.LNumber(c.hitdef.fall_yvelocity))
		case "fall.zvelocity":
			l.Push(lua.LNumber(c.hitdef.fall_zvelocity))
		default:
			l.RaiseError("\nInvalid argument: %v\n", strArg(l, 1))
		}
		return 1
	})
	luaRegister(l, "hitFall", func(*lua.LState) int {
		l.Push(lua.LBool(sys.debugWC.ghv.fallflag))
		return 1
	})
	luaRegister(l, "hitOver", func(*lua.LState) int {
		l.Push(lua.LBool(sys.debugWC.hitOver()))
		return 1
	})
	luaRegister(l, "hitOverridden", func(*lua.LState) int {
		l.Push(lua.LBool(sys.debugWC.hoverIdx >= 0))
		return 1
	})
	luaRegister(l, "hitPauseTime", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.hitPauseTime))
		return 1
	})
	luaRegister(l, "hitShakeOver", func(*lua.LState) int {
		l.Push(lua.LBool(sys.debugWC.hitShakeOver()))
		return 1
	})
	luaRegister(l, "hitVelX", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.ghv.xvel * sys.debugWC.facing))
		return 1
	})
	luaRegister(l, "hitVelY", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.ghv.yvel))
		return 1
	})
	luaRegister(l, "hitVelZ", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.ghv.zvel))
		return 1
	})
	luaRegister(l, "id", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.id))
		return 1
	})
	luaRegister(l, "ikemenVersion", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.gi().ikemenverF))
		return 1
	})
	luaRegister(l, "inCustomAnim", func(*lua.LState) int {
		l.Push(lua.LBool(sys.debugWC.animPN != sys.debugWC.playerNo))
		return 1
	})
	luaRegister(l, "inCustomState", func(*lua.LState) int {
		l.Push(lua.LBool(sys.debugWC.ss.sb.playerNo != sys.debugWC.playerNo))
		return 1
	})
	luaRegister(l, "index", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.indexTrigger()))
		return 1
	})
	luaRegister(l, "inGuardDist", func(*lua.LState) int {
		l.Push(lua.LBool(sys.debugWC.inguarddist))
		return 1
	})
	luaRegister(l, "inputTime", func(l *lua.LState) int {
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
	luaRegister(l, "introState", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.introState()))
		return 1
	})
	luaRegister(l, "isAsserted", func(*lua.LState) int {
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
		case "nodestroyself":
			l.Push(lua.LBool(sys.debugWC.asf(ASF_nodestroyself)))
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
	luaRegister(l, "isHelper", func(l *lua.LState) int {
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
	luaRegister(l, "isHomeTeam", func(*lua.LState) int {
		l.Push(lua.LBool(sys.debugWC.teamside == sys.home))
		return 1
	})
	luaRegister(l, "isHost", func(*lua.LState) int {
		l.Push(lua.LBool(sys.debugWC.isHost()))
		return 1
	})
	luaRegister(l, "jugglePoints", func(*lua.LState) int {
		id := int32(-1)
		if !nilArg(l, 1) {
			id = int32(numArg(l, 1))
		}
		l.Push(lua.LNumber(sys.debugWC.jugglePoints(id)))
		return 1
	})
	luaRegister(l, "lastPlayerId", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.lastCharId))
		return 1
	})
	luaRegister(l, "layerNo", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.layerNo))
		return 1
	})
	luaRegister(l, "leftEdge", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.leftEdge()))
		return 1
	})
	luaRegister(l, "lerp", func(*lua.LState) int {
		a := float32(numArg(l, 1))
		b := float32(numArg(l, 2))
		amount := float32(numArg(l, 3))
		retv := a + (b-a)*Clamp(amount, 0, 1)
		l.Push(lua.LNumber(retv))
		return 1
	})
	luaRegister(l, "life", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.life))
		return 1
	})
	luaRegister(l, "lifeMax", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.lifeMax))
		return 1
	})
	luaRegister(l, "localCoordX", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.cgi[sys.debugWC.playerNo].localcoord[0]))
		return 1
	})
	luaRegister(l, "localCoordY", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.cgi[sys.debugWC.playerNo].localcoord[1]))
		return 1
	})
	luaRegister(l, "lose", func(*lua.LState) int {
		l.Push(lua.LBool(sys.debugWC.lose()))
		return 1
	})
	luaRegister(l, "loseKO", func(*lua.LState) int {
		l.Push(lua.LBool(sys.debugWC.loseKO()))
		return 1
	})
	luaRegister(l, "loseTime", func(*lua.LState) int {
		l.Push(lua.LBool(sys.debugWC.loseTime()))
		return 1
	})
	luaRegister(l, "map", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.mapArray[strings.ToLower(strArg(l, 1))]))
		return 1
	})
	luaRegister(l, "matchNo", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.match))
		return 1
	})
	luaRegister(l, "matchOver", func(*lua.LState) int {
		l.Push(lua.LBool(sys.matchOver()))
		return 1
	})
	luaRegister(l, "memberNo", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.memberNo + 1))
		return 1
	})
	luaRegister(l, "motifState", func(*lua.LState) int {
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
	luaRegister(l, "motifVar", func(l *lua.LState) int {
		value, err := sys.motif.GetValue(strArg(l, 1))
		if err == nil {
			lv := toLValue(l, value)
			l.Push(lv)
			return 1
		}
		l.RaiseError("\nInvalid argument: %v\n", strArg(l, 1))
		return 0
	})
	luaRegister(l, "moveContact", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.moveContact()))
		return 1
	})
	luaRegister(l, "moveCountered", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.moveCountered()))
		return 1
	})
	luaRegister(l, "moveGuarded", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.moveGuarded()))
		return 1
	})
	luaRegister(l, "moveHit", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.moveHit()))
		return 1
	})
	luaRegister(l, "moveHitVar", func(*lua.LState) int {
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
		case "power":
			lv = lua.LNumber(c.mhv.power)
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
	luaRegister(l, "moveReversed", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.moveReversed()))
		return 1
	})
	luaRegister(l, "moveType", func(*lua.LState) int {
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
	luaRegister(l, "mugenVersion", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.gi().mugenverF))
		return 1
	})
	// name also returns p1Name-p8Name variants and helperName
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
			if p := sys.charList.enemyNear(sys.debugWC, n/2-1, true); p != nil {
				l.Push(lua.LString(p.name))
			} else {
				l.Push(lua.LString(""))
			}
		}
		return 1
	})
	luaRegister(l, "numEnemy", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.numEnemy()))
		return 1
	})
	luaRegister(l, "numExplod", func(*lua.LState) int {
		id := int32(-1)
		if !nilArg(l, 1) {
			id = int32(numArg(l, 1))
		}
		l.Push(lua.LNumber(sys.debugWC.numExplod(BytecodeInt(id)).ToI()))
		return 1
	})
	luaRegister(l, "numHelper", func(*lua.LState) int {
		id := int32(0)
		if !nilArg(l, 1) {
			id = int32(numArg(l, 1))
		}
		l.Push(lua.LNumber(sys.debugWC.numHelper(BytecodeInt(id)).ToI()))
		return 1
	})
	luaRegister(l, "numPartner", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.numPartner()))
		return 1
	})
	luaRegister(l, "numPlayer", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.numPlayer()))
		return 1
	})
	luaRegister(l, "numProj", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.numProj()))
		return 1
	})
	luaRegister(l, "numProjId", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.numProjID(
			BytecodeInt(int32(numArg(l, 1)))).ToI()))
		return 1
	})
	luaRegister(l, "numStageBg", func(*lua.LState) int {
		id := int32(-1)
		if !nilArg(l, 1) {
			id = int32(numArg(l, 1))
		}
		l.Push(lua.LNumber(sys.debugWC.numStageBG(BytecodeInt(id)).ToI()))
		return 1
	})
	luaRegister(l, "numTarget", func(*lua.LState) int {
		id := int32(-1)
		if !nilArg(l, 1) {
			id = int32(numArg(l, 1))
		}
		l.Push(lua.LNumber(sys.debugWC.numTarget(BytecodeInt(id)).ToI()))
		return 1
	})
	luaRegister(l, "numText", func(*lua.LState) int {
		id := int32(-1)
		if !nilArg(l, 1) {
			id = int32(numArg(l, 1))
		}
		l.Push(lua.LNumber(sys.debugWC.numText(BytecodeInt(id)).ToI()))
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
	luaRegister(l, "outroState", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.outroState()))
		return 1
	})
	// p1Name and other variants can be checked via name
	luaRegister(l, "p2BodyDistX", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.p2BodyDistX(sys.debugWC).ToI()))
		return 1
	})
	luaRegister(l, "p2BodyDistY", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.p2BodyDistY(sys.debugWC).ToI()))
		return 1
	})
	luaRegister(l, "p2BodyDistZ", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.p2BodyDistZ(sys.debugWC).ToI()))
		return 1
	})
	luaRegister(l, "p2DistX", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.rdDistX(sys.debugWC.p2(), sys.debugWC).ToI()))
		return 1
	})
	luaRegister(l, "p2DistY", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.rdDistY(sys.debugWC.p2(), sys.debugWC).ToI()))
		return 1
	})
	luaRegister(l, "p2DistZ", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.rdDistZ(sys.debugWC.p2(), sys.debugWC).ToI()))
		return 1
	})
	luaRegister(l, "p2Life", func(*lua.LState) int {
		if p2 := sys.debugWC.p2(); p2 != nil {
			l.Push(lua.LNumber(p2.life))
		} else {
			l.Push(lua.LNumber(-1))
		}
		return 1
	})
	luaRegister(l, "p2MoveType", func(*lua.LState) int {
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
	luaRegister(l, "p2StateNo", func(*lua.LState) int {
		if p2 := sys.debugWC.p2(); p2 != nil {
			l.Push(lua.LNumber(p2.ss.no))
		}
		return 1
	})
	luaRegister(l, "p2StateType", func(*lua.LState) int {
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
	luaRegister(l, "palFxVar", func(*lua.LState) int {
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
	luaRegister(l, "palNo", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.gi().palno))
		return 1
	})
	luaRegister(l, "parentDistX", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.rdDistX(sys.debugWC.parent(true), sys.debugWC).ToI()))
		return 1
	})
	luaRegister(l, "parentDistY", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.rdDistY(sys.debugWC.parent(true), sys.debugWC).ToI()))
		return 1
	})
	luaRegister(l, "parentDistZ", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.rdDistZ(sys.debugWC.parent(true), sys.debugWC).ToI()))
		return 1
	})
	luaRegister(l, "parentExist", func(*lua.LState) int {
		l.Push(lua.LBool(sys.debugWC.parentExist()))
		return 1
	})
	luaRegister(l, "pauseTime", func(*lua.LState) int {
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
	luaRegister(l, "playerIdExist", func(*lua.LState) int {
		l.Push(lua.LBool(sys.playerIDExist(
			BytecodeInt(int32(numArg(l, 1)))).ToB()))
		return 1
	})
	luaRegister(l, "playerIndexExist", func(*lua.LState) int {
		l.Push(lua.LBool(sys.playerIndexExist(
			BytecodeInt(int32(numArg(l, 1)))).ToB()))
		return 1
	})
	luaRegister(l, "playerNo", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.playerNo + 1))
		return 1
	})
	luaRegister(l, "playerNoExist", func(*lua.LState) int {
		l.Push(lua.LBool(sys.playerNoExist(
			BytecodeInt(int32(numArg(l, 1)))).ToB()))
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
	luaRegister(l, "powerMax", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.powerMax))
		return 1
	})
	luaRegister(l, "prevAnim", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.prevAnimNo))
		return 1
	})
	luaRegister(l, "prevMoveType", func(*lua.LState) int {
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
	luaRegister(l, "prevStateNo", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.ss.prevno))
		return 1
	})
	luaRegister(l, "prevStateType", func(*lua.LState) int {
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
	luaRegister(l, "projCancelTime", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.projCancelTime(
			BytecodeInt(int32(numArg(l, 1)))).ToI()))
		return 1
	})
	luaRegister(l, "projClsnOverlap", func(l *lua.LState) int {
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
	// projContact (deprecated by projContactTime)
	luaRegister(l, "projContactTime", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.projContactTime(
			BytecodeInt(int32(numArg(l, 1)))).ToI()))
		return 1
	})
	// projGuarded (deprecated by projGuardedTime)
	luaRegister(l, "projGuardedTime", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.projGuardedTime(
			BytecodeInt(int32(numArg(l, 1)))).ToI()))
		return 1
	})
	// projhit (deprecated by projhittime)
	luaRegister(l, "projHitTime", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.projHitTime(
			BytecodeInt(int32(numArg(l, 1)))).ToI()))
		return 1
	})
	luaRegister(l, "projVar", func(*lua.LState) int {
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
				lv = lua.LNumber(p.hitdef.teamside + 1)
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
	// rad (dedicated functionality already exists in Lua)
	// random (dedicated functionality already exists in Lua)
	// randomRange (dedicated functionality already exists in Lua)
	luaRegister(l, "ratioLevel", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.ocd().ratioLevel))
		return 1
	})
	luaRegister(l, "receivedDamage", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.receivedDmg))
		return 1
	})
	luaRegister(l, "receivedHits", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.receivedHits))
		return 1
	})
	luaRegister(l, "redLife", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.redLife))
		return 1
	})
	luaRegister(l, "reversalDefAttr", func(*lua.LState) int {
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
	luaRegister(l, "rightEdge", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.rightEdge()))
		return 1
	})
	luaRegister(l, "rootDistX", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.rdDistX(sys.debugWC.root(true), sys.debugWC).ToI()))
		return 1
	})
	luaRegister(l, "rootDistY", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.rdDistY(sys.debugWC.root(true), sys.debugWC).ToI()))
		return 1
	})
	luaRegister(l, "rootDistZ", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.rdDistZ(sys.debugWC.root(true), sys.debugWC).ToI()))
		return 1
	})
	luaRegister(l, "roundNo", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.round))
		return 1
	})
	luaRegister(l, "roundsExisted", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.roundsExisted()))
		return 1
	})
	luaRegister(l, "roundState", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.roundState()))
		return 1
	})
	luaRegister(l, "roundsWon", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.roundsWon()))
		return 1
	})
	luaRegister(l, "roundTime", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.tickCount))
		return 1
	})
	luaRegister(l, "runOrder", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.runOrderTrigger()))
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
	luaRegister(l, "score", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.score()))
		return 1
	})
	luaRegister(l, "scoreTotal", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.scoreTotal()))
		return 1
	})
	luaRegister(l, "screenHeight", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.screenHeight()))
		return 1
	})
	luaRegister(l, "screenPosX", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.screenPosX()))
		return 1
	})
	luaRegister(l, "screenPosY", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.screenPosY()))
		return 1
	})
	luaRegister(l, "screenWidth", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.screenWidth()))
		return 1
	})
	luaRegister(l, "selfAnimExist", func(*lua.LState) int {
		l.Push(lua.LBool(sys.debugWC.selfAnimExist(
			BytecodeInt(int32(numArg(l, 1)))).ToB()))
		return 1
	})
	luaRegister(l, "selfStateNoExist", func(*lua.LState) int {
		l.Push(lua.LBool(sys.debugWC.selfStatenoExist(
			BytecodeInt(int32(numArg(l, 1)))).ToB()))
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
	luaRegister(l, "soundVar", func(*lua.LState) int {
		var lv lua.LValue
		id := int32(numArg(l, 1))
		vname := strArg(l, 2)
		var ch *SoundChannel
		c := sys.debugWC

		if id >= 0 {
			ch = sys.charSoundChannels[c.playerNo].Get(c.id, id)
		} else {
			for i := range sys.charSoundChannels[c.playerNo] {
				v := &sys.charSoundChannels[c.playerNo][i]
				if v.sfx != nil && v.IsPlaying() {
					ch = v
					break
				}
			}
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
				lv = lua.LNumber(ch.sfx.pan)
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
	luaRegister(l, "spriteVar", func(l *lua.LState) int {
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
	luaRegister(l, "sprPriority", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.sprPriority))
		return 1
	})
	luaRegister(l, "stageBackEdgeDist", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.stageBackEdgeDist()))
		return 1
	})
	luaRegister(l, "stageBgVar", func(l *lua.LState) int {
		id := int32(numArg(l, 1))
		idx := int(numArg(l, 2))
		vname := strArg(l, 3)
		var ln lua.LNumber
		// Get stage background element
		bg := sys.debugWC.getSingleStageBg(id, idx, true)
		// Handle returns
		if bg != nil {
			switch strings.ToLower(vname) {
			case "actionno":
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
	luaRegister(l, "stageConst", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.stage.constants[strings.ToLower(strArg(l, 1))]))
		return 1
	})
	luaRegister(l, "stageFrontEdgeDist", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.stageFrontEdgeDist()))
		return 1
	})
	luaRegister(l, "stageTime", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.stage.stageTime))
		return 1
	})
	luaRegister(l, "stageVar", func(*lua.LState) int {
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
	luaRegister(l, "standby", func(*lua.LState) int {
		l.Push(lua.LBool(sys.debugWC.scf(SCF_standby)))
		return 1
	})
	luaRegister(l, "stateNo", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.ss.no))
		return 1
	})
	luaRegister(l, "stateType", func(*lua.LState) int {
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
	luaRegister(l, "sysFvar", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.sysFvarGet(int32(numArg(l, 1))).ToF()))
		return 1
	})
	luaRegister(l, "sysVar", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.sysVarGet(int32(numArg(l, 1))).ToI()))
		return 1
	})
	luaRegister(l, "teamLeader", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.teamLeader()))
		return 1
	})
	luaRegister(l, "teamMode", func(*lua.LState) int {
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
	luaRegister(l, "teamSide", func(*lua.LState) int {
		l.Push(lua.LNumber(int32(sys.debugWC.teamside) + 1))
		return 1
	})
	luaRegister(l, "teamSize", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.teamSize()))
		return 1
	})
	luaRegister(l, "ticksPerSecond", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.gameLogicSpeed()))
		return 1
	})
	luaRegister(l, "time", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.ss.time))
		return 1
	})
	luaRegister(l, "timeElapsed", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.timeElapsed()))
		return 1
	})
	luaRegister(l, "timeMod", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.ss.time % int32(numArg(l, 1))))
		return 1
	})
	luaRegister(l, "timeRemaining", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.timeRemaining()))
		return 1
	})
	luaRegister(l, "timeTotal", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.timeTotal()))
		return 1
	})
	luaRegister(l, "topBoundBodyDist", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.topBoundBodyDist()))
		return 1
	})
	luaRegister(l, "topBoundDist", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.topBoundDist()))
		return 1
	})
	luaRegister(l, "topEdge", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.topEdge()))
		return 1
	})
	luaRegister(l, "uniqHitCount", func(*lua.LState) int {
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
	luaRegister(l, "winClutch", func(*lua.LState) int {
		l.Push(lua.LBool(sys.debugWC.winClutch()))
		return 1
	})
	luaRegister(l, "winHyper", func(*lua.LState) int {
		l.Push(lua.LBool(sys.debugWC.winType(WT_Hyper)))
		return 1
	})
	luaRegister(l, "winKO", func(*lua.LState) int {
		l.Push(lua.LBool(sys.debugWC.winKO()))
		return 1
	})
	luaRegister(l, "winPerfect", func(*lua.LState) int {
		l.Push(lua.LBool(sys.debugWC.winPerfect()))
		return 1
	})
	luaRegister(l, "winSpecial", func(*lua.LState) int {
		l.Push(lua.LBool(sys.debugWC.winType(WT_Special)))
		return 1
	})
	luaRegister(l, "winTime", func(*lua.LState) int {
		l.Push(lua.LBool(sys.debugWC.winTime()))
		return 1
	})
	luaRegister(l, "xAngle", func(*lua.LState) int {
		if sys.debugWC.csf(CSF_angledraw) {
			l.Push(lua.LNumber(sys.debugWC.anglerot[1]))
		} else {
			l.Push(lua.LNumber(0))
		}
		return 1
	})
	luaRegister(l, "xShear", func(*lua.LState) int {
		l.Push(lua.LNumber(sys.debugWC.xshear))
		return 1
	})
	luaRegister(l, "yAngle", func(*lua.LState) int {
		if sys.debugWC.csf(CSF_angledraw) {
			l.Push(lua.LNumber(sys.debugWC.anglerot[2]))
		} else {
			l.Push(lua.LNumber(0))
		}
		return 1
	})
	luaRegister(l, "zoomVar", func(*lua.LState) int {
		var ln lua.LNumber
		switch strings.ToLower(strArg(l, 1)) {
		case "scale":
			ln = lua.LNumber(sys.zoom.curScale)
		case "pos.x":
			ln = lua.LNumber(sys.zoom.curPos[0])
		case "pos.y":
			ln = lua.LNumber(sys.zoom.curPos[1])
		case "lag":
			ln = lua.LNumber(sys.zoom.lag)
		case "endlag":
			ln = lua.LNumber(sys.zoom.endLag)
		case "time":
			ln = lua.LNumber(sys.zoom.time)
		default:
			l.RaiseError("\nInvalid argument: %v\n", strArg(l, 1))
		}
		l.Push(ln)
		return 1
	})
}
