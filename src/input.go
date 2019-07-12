package main

import (
	"encoding/binary"
	"net"
	"os"
	"strings"
	"time"

	"github.com/go-gl/glfw/v3.2/glfw"
)

type CommandKey byte

const (
	CK_B CommandKey = iota
	CK_D
	CK_F
	CK_U
	CK_DB
	CK_UB
	CK_DF
	CK_UF
	CK_nB
	CK_nD
	CK_nF
	CK_nU
	CK_nDB
	CK_nUB
	CK_nDF
	CK_nUF
	CK_Bs
	CK_Ds
	CK_Fs
	CK_Us
	CK_DBs
	CK_UBs
	CK_DFs
	CK_UFs
	CK_nBs
	CK_nDs
	CK_nFs
	CK_nUs
	CK_nDBs
	CK_nUBs
	CK_nDFs
	CK_nUFs
	CK_a
	CK_b
	CK_c
	CK_x
	CK_y
	CK_z
	CK_s
	CK_v
	CK_w
	CK_na
	CK_nb
	CK_nc
	CK_nx
	CK_ny
	CK_nz
	CK_ns
	CK_nv
	CK_nw
	CK_Last = CK_nw
)

type NetState int

const (
	NS_Stop NetState = iota
	NS_Playing
	NS_End
	NS_Stopped
	NS_Error
)

func StringToKey(s string) glfw.Key {
	switch s {
	case "RETURN":
		return glfw.KeyEnter
	case "ESCAPE":
		return glfw.KeyEscape
	case "BACKSPACE":
		return glfw.KeyBackspace
	case "TAB":
		return glfw.KeyTab
	case "SPACE":
		return glfw.KeySpace
	case "QUOTE":
		return glfw.KeyApostrophe
	case "COMMA":
		return glfw.KeyComma
	case "MINUS":
		return glfw.KeyMinus
	case "PERIOD":
		return glfw.KeyPeriod
	case "SLASH":
		return glfw.KeySlash
	case "0":
		return glfw.Key0
	case "1":
		return glfw.Key1
	case "2":
		return glfw.Key2
	case "3":
		return glfw.Key3
	case "4":
		return glfw.Key4
	case "5":
		return glfw.Key5
	case "6":
		return glfw.Key6
	case "7":
		return glfw.Key7
	case "8":
		return glfw.Key8
	case "9":
		return glfw.Key9
	case "SEMICOLON":
		return glfw.KeySemicolon
	case "EQUALS":
		return glfw.KeyEqual
	case "LEFTBRACKET":
		return glfw.KeyLeftBracket
	case "BACKSLASH":
		return glfw.KeyBackslash
	case "RIGHTBRACKET":
		return glfw.KeyRightBracket
	case "BACKQUOTE":
		return glfw.KeyGraveAccent
	case "a":
		return glfw.KeyA
	case "b":
		return glfw.KeyB
	case "c":
		return glfw.KeyC
	case "d":
		return glfw.KeyD
	case "e":
		return glfw.KeyE
	case "f":
		return glfw.KeyF
	case "g":
		return glfw.KeyG
	case "h":
		return glfw.KeyH
	case "i":
		return glfw.KeyI
	case "j":
		return glfw.KeyJ
	case "k":
		return glfw.KeyK
	case "l":
		return glfw.KeyL
	case "m":
		return glfw.KeyM
	case "n":
		return glfw.KeyN
	case "o":
		return glfw.KeyO
	case "p":
		return glfw.KeyP
	case "q":
		return glfw.KeyQ
	case "r":
		return glfw.KeyR
	case "s":
		return glfw.KeyS
	case "t":
		return glfw.KeyT
	case "u":
		return glfw.KeyU
	case "v":
		return glfw.KeyV
	case "w":
		return glfw.KeyW
	case "x":
		return glfw.KeyX
	case "y":
		return glfw.KeyY
	case "z":
		return glfw.KeyZ
	case "CAPSLOCK":
		return glfw.KeyCapsLock
	case "F1":
		return glfw.KeyF1
	case "F2":
		return glfw.KeyF2
	case "F3":
		return glfw.KeyF3
	case "F4":
		return glfw.KeyF4
	case "F5":
		return glfw.KeyF5
	case "F6":
		return glfw.KeyF6
	case "F7":
		return glfw.KeyF7
	case "F8":
		return glfw.KeyF8
	case "F9":
		return glfw.KeyF9
	case "F10":
		return glfw.KeyF10
	case "F11":
		return glfw.KeyF11
	case "F12":
		return glfw.KeyF12
	case "PRINTSCREEN":
		return glfw.KeyPrintScreen
	case "SCROLLLOCK":
		return glfw.KeyScrollLock
	case "PAUSE":
		return glfw.KeyPause
	case "INSERT":
		return glfw.KeyInsert
	case "HOME":
		return glfw.KeyHome
	case "PAGEUP":
		return glfw.KeyPageUp
	case "DELETE":
		return glfw.KeyDelete
	case "END":
		return glfw.KeyEnd
	case "PAGEDOWN":
		return glfw.KeyPageDown
	case "RIGHT":
		return glfw.KeyRight
	case "LEFT":
		return glfw.KeyLeft
	case "DOWN":
		return glfw.KeyDown
	case "UP":
		return glfw.KeyUp
	case "NUMLOCKCLEAR":
		return glfw.KeyNumLock
	case "KP_DIVIDE":
		return glfw.KeyKPDivide
	case "KP_MULTIPLY":
		return glfw.KeyKPMultiply
	case "KP_MINUS":
		return glfw.KeyKPSubtract
	case "KP_PLUS":
		return glfw.KeyKPAdd
	case "KP_ENTER":
		return glfw.KeyKPEnter
	case "KP_1":
		return glfw.KeyKP1
	case "KP_2":
		return glfw.KeyKP2
	case "KP_3":
		return glfw.KeyKP3
	case "KP_4":
		return glfw.KeyKP4
	case "KP_5":
		return glfw.KeyKP5
	case "KP_6":
		return glfw.KeyKP6
	case "KP_7":
		return glfw.KeyKP7
	case "KP_8":
		return glfw.KeyKP8
	case "KP_9":
		return glfw.KeyKP9
	case "KP_0":
		return glfw.KeyKP0
	case "KP_PERIOD":
		return glfw.KeyKPDecimal
	case "KP_EQUALS":
		return glfw.KeyKPEqual
	case "F13":
		return glfw.KeyF13
	case "F14":
		return glfw.KeyF14
	case "F15":
		return glfw.KeyF15
	case "F16":
		return glfw.KeyF16
	case "F17":
		return glfw.KeyF17
	case "F18":
		return glfw.KeyF18
	case "F19":
		return glfw.KeyF19
	case "F20":
		return glfw.KeyF20
	case "F21":
		return glfw.KeyF21
	case "F22":
		return glfw.KeyF22
	case "F23":
		return glfw.KeyF23
	case "F24":
		return glfw.KeyF24
	case "MENU":
		return glfw.KeyMenu
	case "LCTRL":
		return glfw.KeyLeftControl
	case "LSHIFT":
		return glfw.KeyLeftShift
	case "LALT":
		return glfw.KeyLeftAlt
	case "LGUI":
		return glfw.KeyLeftSuper
	case "RCTRL":
		return glfw.KeyRightControl
	case "RSHIFT":
		return glfw.KeyRightShift
	case "RALT":
		return glfw.KeyRightAlt
	case "RGUI":
		return glfw.KeyRightSuper
	}
	return glfw.KeyUnknown
}

func KeyToString(k glfw.Key) string {
	switch k {
	case glfw.KeyEnter:
		return "RETURN"
	case glfw.KeyEscape:
		return "ESCAPE"
	case glfw.KeyBackspace:
		return "BACKSPACE"
	case glfw.KeyTab:
		return "TAB"
	case glfw.KeySpace:
		return "SPACE"
	case glfw.KeyApostrophe:
		return "QUOTE"
	case glfw.KeyComma:
		return "COMMA"
	case glfw.KeyMinus:
		return "MINUS"
	case glfw.KeyPeriod:
		return "PERIOD"
	case glfw.KeySlash:
		return "SLASH"
	case glfw.Key0:
		return "0"
	case glfw.Key1:
		return "1"
	case glfw.Key2:
		return "2"
	case glfw.Key3:
		return "3"
	case glfw.Key4:
		return "4"
	case glfw.Key5:
		return "5"
	case glfw.Key6:
		return "6"
	case glfw.Key7:
		return "7"
	case glfw.Key8:
		return "8"
	case glfw.Key9:
		return "9"
	case glfw.KeySemicolon:
		return "SEMICOLON"
	case glfw.KeyEqual:
		return "EQUALS"
	case glfw.KeyLeftBracket:
		return "LEFTBRACKET"
	case glfw.KeyBackslash:
		return "BACKSLASH"
	case glfw.KeyRightBracket:
		return "RIGHTBRACKET"
	case glfw.KeyGraveAccent:
		return "BACKQUOTE"
	case glfw.KeyA:
		return "a"
	case glfw.KeyB:
		return "b"
	case glfw.KeyC:
		return "c"
	case glfw.KeyD:
		return "d"
	case glfw.KeyE:
		return "e"
	case glfw.KeyF:
		return "f"
	case glfw.KeyG:
		return "g"
	case glfw.KeyH:
		return "h"
	case glfw.KeyI:
		return "i"
	case glfw.KeyJ:
		return "j"
	case glfw.KeyK:
		return "k"
	case glfw.KeyL:
		return "l"
	case glfw.KeyM:
		return "m"
	case glfw.KeyN:
		return "n"
	case glfw.KeyO:
		return "o"
	case glfw.KeyP:
		return "p"
	case glfw.KeyQ:
		return "q"
	case glfw.KeyR:
		return "r"
	case glfw.KeyS:
		return "s"
	case glfw.KeyT:
		return "t"
	case glfw.KeyU:
		return "u"
	case glfw.KeyV:
		return "v"
	case glfw.KeyW:
		return "w"
	case glfw.KeyX:
		return "x"
	case glfw.KeyY:
		return "y"
	case glfw.KeyZ:
		return "z"
	case glfw.KeyCapsLock:
		return "CAPSLOCK"
	case glfw.KeyF1:
		return "F1"
	case glfw.KeyF2:
		return "F2"
	case glfw.KeyF3:
		return "F3"
	case glfw.KeyF4:
		return "F4"
	case glfw.KeyF5:
		return "F5"
	case glfw.KeyF6:
		return "F6"
	case glfw.KeyF7:
		return "F7"
	case glfw.KeyF8:
		return "F8"
	case glfw.KeyF9:
		return "F9"
	case glfw.KeyF10:
		return "F10"
	case glfw.KeyF11:
		return "F11"
	case glfw.KeyF12:
		return "F12"
	case glfw.KeyPrintScreen:
		return "PRINTSCREEN"
	case glfw.KeyScrollLock:
		return "SCROLLLOCK"
	case glfw.KeyPause:
		return "PAUSE"
	case glfw.KeyInsert:
		return "INSERT"
	case glfw.KeyHome:
		return "HOME"
	case glfw.KeyPageUp:
		return "PAGEUP"
	case glfw.KeyDelete:
		return "DELETE"
	case glfw.KeyEnd:
		return "END"
	case glfw.KeyPageDown:
		return "PAGEDOWN"
	case glfw.KeyRight:
		return "RIGHT"
	case glfw.KeyLeft:
		return "LEFT"
	case glfw.KeyDown:
		return "DOWN"
	case glfw.KeyUp:
		return "UP"
	case glfw.KeyNumLock:
		return "NUMLOCKCLEAR"
	case glfw.KeyKPDivide:
		return "KP_DIVIDE"
	case glfw.KeyKPMultiply:
		return "KP_MULTIPLY"
	case glfw.KeyKPSubtract:
		return "KP_MINUS"
	case glfw.KeyKPAdd:
		return "KP_PLUS"
	case glfw.KeyKPEnter:
		return "KP_ENTER"
	case glfw.KeyKP1:
		return "KP_1"
	case glfw.KeyKP2:
		return "KP_2"
	case glfw.KeyKP3:
		return "KP_3"
	case glfw.KeyKP4:
		return "KP_4"
	case glfw.KeyKP5:
		return "KP_5"
	case glfw.KeyKP6:
		return "KP_6"
	case glfw.KeyKP7:
		return "KP_7"
	case glfw.KeyKP8:
		return "KP_8"
	case glfw.KeyKP9:
		return "KP_9"
	case glfw.KeyKP0:
		return "KP_0"
	case glfw.KeyKPDecimal:
		return "KP_PERIOD"
	case glfw.KeyKPEqual:
		return "KP_EQUALS"
	case glfw.KeyF13:
		return "F13"
	case glfw.KeyF14:
		return "F14"
	case glfw.KeyF15:
		return "F15"
	case glfw.KeyF16:
		return "F16"
	case glfw.KeyF17:
		return "F17"
	case glfw.KeyF18:
		return "F18"
	case glfw.KeyF19:
		return "F19"
	case glfw.KeyF20:
		return "F20"
	case glfw.KeyF21:
		return "F21"
	case glfw.KeyF22:
		return "F22"
	case glfw.KeyF23:
		return "F23"
	case glfw.KeyF24:
		return "F24"
	case glfw.KeyMenu:
		return "MENU"
	case glfw.KeyLeftControl:
		return "LCTRL"
	case glfw.KeyLeftShift:
		return "LSHIFT"
	case glfw.KeyLeftAlt:
		return "LALT"
	case glfw.KeyLeftSuper:
		return "LGUI"
	case glfw.KeyRightControl:
		return "RCTRL"
	case glfw.KeyRightShift:
		return "RSHIFT"
	case glfw.KeyRightAlt:
		return "RALT"
	case glfw.KeyRightSuper:
		return "RGUI"
	}
	return ""
}

type ShortcutScript struct {
	Activate bool
	Script   string
}
type ShortcutKey struct {
	Key glfw.Key
	Mod glfw.ModifierKey
}

func NewShortcutKey(key glfw.Key, ctrl, alt, shift bool) *ShortcutKey {
	sk := &ShortcutKey{}
	sk.Key = key
	sk.Mod = 0
	if ctrl {
		sk.Mod |= glfw.ModControl
	}
	if alt {
		sk.Mod |= glfw.ModAlt
	}
	if shift {
		sk.Mod |= glfw.ModShift
	}
	return sk
}
func (sk ShortcutKey) Test(k glfw.Key, m glfw.ModifierKey) bool {
	return k == sk.Key &&
		m&(glfw.ModShift|glfw.ModControl|glfw.ModAlt) == sk.Mod
}
func keyCallback(_ *glfw.Window, key glfw.Key, _ int,
	action glfw.Action, mk glfw.ModifierKey) {
	switch action {
	case glfw.Release:
		sys.keySatate[key] = false
		sys.keyInput = glfw.KeyUnknown
		sys.keyString = ""
	case glfw.Press:
		sys.keySatate[key] = true
		sys.keyInput = key
		sys.esc = sys.esc ||
			key == glfw.KeyEscape && mk&(glfw.ModControl|glfw.ModAlt) == 0
		for k, v := range sys.shortcutScripts {
			v.Activate = v.Activate || k.Test(key, mk)
		}
	}
}
func charCallback(_ *glfw.Window, char rune, mk glfw.ModifierKey) {
	sys.keyString = string(char)
}

var joystick = [...]glfw.Joystick{glfw.Joystick1, glfw.Joystick2,
	glfw.Joystick3, glfw.Joystick4, glfw.Joystick5, glfw.Joystick6,
	glfw.Joystick7, glfw.Joystick8, glfw.Joystick9, glfw.Joystick10,
	glfw.Joystick11, glfw.Joystick12, glfw.Joystick13, glfw.Joystick14,
	glfw.Joystick15, glfw.Joystick16}

func JoystickState(joy, button int) bool {
	if joy < 0 {
		return sys.keySatate[glfw.Key(button)]
	}
	if joy >= len(joystick) {
		return false
	}
	btns := glfw.GetJoystickButtons(joystick[joy])
	if button < 0 {
		button = -button - 1
		axes := glfw.GetJoystickAxes(joystick[joy])
		if len(axes)*2 <= button {
			return false
		}
		if (button == 8 || button == 10) && glfw.GetJoystickName(joystick[joy]) == "Xbox 360 Controller" { //Xbox360コントローラーのLRトリガー判定
			return axes[button/2] > 0
		}
		switch button & 1 {
		case 0:
			return axes[button/2] < -0.2
		case 1:
			return axes[button/2] > 0.2
		}
	}
	if len(btns) <= button {
		return false
	}
	return btns[button] != 0
}

type KeyConfig struct{ Joy, u, d, l, r, a, b, c, x, y, z, s, v, w int }

func (kc KeyConfig) U() bool { return JoystickState(kc.Joy, kc.u) }
func (kc KeyConfig) D() bool { return JoystickState(kc.Joy, kc.d) }
func (kc KeyConfig) L() bool { return JoystickState(kc.Joy, kc.l) }
func (kc KeyConfig) R() bool { return JoystickState(kc.Joy, kc.r) }
func (kc KeyConfig) A() bool { return JoystickState(kc.Joy, kc.a) }
func (kc KeyConfig) B() bool { return JoystickState(kc.Joy, kc.b) }
func (kc KeyConfig) C() bool { return JoystickState(kc.Joy, kc.c) }
func (kc KeyConfig) X() bool { return JoystickState(kc.Joy, kc.x) }
func (kc KeyConfig) Y() bool { return JoystickState(kc.Joy, kc.y) }
func (kc KeyConfig) Z() bool { return JoystickState(kc.Joy, kc.z) }
func (kc KeyConfig) S() bool { return JoystickState(kc.Joy, kc.s) }
func (kc KeyConfig) V() bool { return JoystickState(kc.Joy, kc.v) }
func (kc KeyConfig) W() bool { return JoystickState(kc.Joy, kc.w) }

type InputBits int32

const (
	IB_U InputBits = 1 << iota
	IB_D
	IB_L
	IB_R
	IB_A
	IB_B
	IB_C
	IB_X
	IB_Y
	IB_Z
	IB_S
	IB_V
	IB_W
	IB_anybutton = IB_A | IB_B | IB_C | IB_X | IB_Y | IB_Z | IB_V | IB_W
)

func (ib *InputBits) SetInput(in int) {
	if 0 <= in && in < len(sys.keyConfig) {
		*ib = InputBits(Btoi(sys.keyConfig[in].U() || sys.JoystickConfig[in].U()) |
			Btoi(sys.keyConfig[in].D() || sys.JoystickConfig[in].D())<<1 |
			Btoi(sys.keyConfig[in].L() || sys.JoystickConfig[in].L())<<2 |
			Btoi(sys.keyConfig[in].R() || sys.JoystickConfig[in].R())<<3 |
			Btoi(sys.keyConfig[in].A() || sys.JoystickConfig[in].A())<<4 |
			Btoi(sys.keyConfig[in].B() || sys.JoystickConfig[in].B())<<5 |
			Btoi(sys.keyConfig[in].C() || sys.JoystickConfig[in].C())<<6 |
			Btoi(sys.keyConfig[in].X() || sys.JoystickConfig[in].X())<<7 |
			Btoi(sys.keyConfig[in].Y() || sys.JoystickConfig[in].Y())<<8 |
			Btoi(sys.keyConfig[in].Z() || sys.JoystickConfig[in].Z())<<9 |
			Btoi(sys.keyConfig[in].S() || sys.JoystickConfig[in].S())<<10 |
			Btoi(sys.keyConfig[in].V() || sys.JoystickConfig[in].V())<<11 |
			Btoi(sys.keyConfig[in].W() || sys.JoystickConfig[in].W())<<12)
	}
}
func (ib InputBits) GetInput(cb *CommandBuffer, facing int32) {
	var b, f bool
	if facing < 0 {
		b, f = ib&IB_R != 0, ib&IB_L != 0
	} else {
		b, f = ib&IB_L != 0, ib&IB_R != 0
	}
	cb.Input(b, ib&IB_D != 0, f, ib&IB_U != 0, ib&IB_A != 0, ib&IB_B != 0,
		ib&IB_C != 0, ib&IB_X != 0, ib&IB_Y != 0, ib&IB_Z != 0, ib&IB_S != 0, ib&IB_V != 0, ib&IB_W != 0)
}

type CommandKeyRemap struct {
	a, b, c, x, y, z, s, v, w, na, nb, nc, nx, ny, nz, ns, nv, nw CommandKey
}

func NewCommandKeyRemap() *CommandKeyRemap {
	return &CommandKeyRemap{CK_a, CK_b, CK_c, CK_x, CK_y, CK_z, CK_s, CK_v, CK_w,
		CK_na, CK_nb, CK_nc, CK_nx, CK_ny, CK_nz, CK_ns, CK_nv, CK_nw}
}

type CommandBuffer struct {
	Bb, Db, Fb, Ub                     int32
	ab, bb, cb, xb, yb, zb, sb, vb, wb int32
	B, D, F, U                         int8
	a, b, c, x, y, z, s, v, w          int8
}

func NewCommandBuffer() (c *CommandBuffer) {
	c = &CommandBuffer{}
	c.Reset()
	return
}
func (__ *CommandBuffer) Reset() {
	*__ = CommandBuffer{B: -1, D: -1, F: -1, U: -1,
		a: -1, b: -1, c: -1, x: -1, y: -1, z: -1, s: -1, v: -1, w: -1}
}
func (__ *CommandBuffer) Input(B, D, F, U, a, b, c, x, y, z, s, v, w bool) {
	if (B && !F) != (__.B > 0) {
		__.Bb = 0
		__.B *= -1
	}
	__.Bb += int32(__.B)
	if (D && !U) != (__.D > 0) {
		__.Db = 0
		__.D *= -1
	}
	__.Db += int32(__.D)
	if (F && !B) != (__.F > 0) {
		__.Fb = 0
		__.F *= -1
	}
	__.Fb += int32(__.F)
	if (U && !D) != (__.U > 0) {
		__.Ub = 0
		__.U *= -1
	}
	__.Ub += int32(__.U)
	if a != (__.a > 0) {
		__.ab = 0
		__.a *= -1
	}
	__.ab += int32(__.a)
	if b != (__.b > 0) {
		__.bb = 0
		__.b *= -1
	}
	__.bb += int32(__.b)
	if c != (__.c > 0) {
		__.cb = 0
		__.c *= -1
	}
	__.cb += int32(__.c)
	if x != (__.x > 0) {
		__.xb = 0
		__.x *= -1
	}
	__.xb += int32(__.x)
	if y != (__.y > 0) {
		__.yb = 0
		__.y *= -1
	}
	__.yb += int32(__.y)
	if z != (__.z > 0) {
		__.zb = 0
		__.z *= -1
	}
	__.zb += int32(__.z)
	if s != (__.s > 0) {
		__.sb = 0
		__.s *= -1
	}
	__.sb += int32(__.s)
	if v != (__.v > 0) {
		__.vb = 0
		__.v *= -1
	}
	__.vb += int32(__.v)
	if w != (__.w > 0) {
		__.wb = 0
		__.w *= -1
	}
	__.wb += int32(__.w)
}
func (__ *CommandBuffer) InputBits(ib InputBits, f int32) {
	var B, F bool
	if f < 0 {
		B, F = ib&IB_R != 0, ib&IB_L != 0
	} else {
		B, F = ib&IB_L != 0, ib&IB_R != 0
	}
	__.Input(B, ib&IB_D != 0, F, ib&IB_U != 0, ib&IB_A != 0, ib&IB_B != 0,
		ib&IB_C != 0, ib&IB_X != 0, ib&IB_Y != 0, ib&IB_Z != 0, ib&IB_S != 0, ib&IB_V != 0, ib&IB_W != 0)
}
func (__ *CommandBuffer) State(ck CommandKey) int32 {
	switch ck {
	case CK_B:
		return Min(-Max(__.Db, __.Ub), __.Bb)
	case CK_D:
		return Min(-Max(__.Bb, __.Fb), __.Db)
	case CK_F:
		return Min(-Max(__.Db, __.Ub), __.Fb)
	case CK_U:
		return Min(-Max(__.Bb, __.Fb), __.Ub)
	case CK_DB:
		return Min(__.Db, __.Bb)
	case CK_UB:
		return Min(__.Ub, __.Bb)
	case CK_DF:
		return Min(__.Db, __.Fb)
	case CK_UF:
		return Min(__.Ub, __.Fb)
	case CK_Bs:
		return __.Bb
	case CK_Ds:
		return __.Db
	case CK_Fs:
		return __.Fb
	case CK_Us:
		return __.Ub
	case CK_DBs:
		return Min(__.Db, __.Bb)
	case CK_UBs:
		return Min(__.Ub, __.Bb)
	case CK_DFs:
		return Min(__.Db, __.Fb)
	case CK_UFs:
		return Min(__.Ub, __.Fb)
	case CK_a:
		return __.ab
	case CK_b:
		return __.bb
	case CK_c:
		return __.cb
	case CK_x:
		return __.xb
	case CK_y:
		return __.yb
	case CK_z:
		return __.zb
	case CK_s:
		return __.sb
	case CK_v:
		return __.vb
	case CK_w:
		return __.wb
	case CK_nB:
		return -Min(-Max(__.Db, __.Ub), __.Bb)
	case CK_nD:
		return -Min(-Max(__.Bb, __.Fb), __.Db)
	case CK_nF:
		return -Min(-Max(__.Db, __.Ub), __.Fb)
	case CK_nU:
		return -Min(-Max(__.Bb, __.Fb), __.Ub)
	case CK_nDB:
		return -Min(__.Db, __.Bb)
	case CK_nUB:
		return -Min(__.Ub, __.Bb)
	case CK_nDF:
		return -Min(__.Db, __.Fb)
	case CK_nUF:
		return -Min(__.Ub, __.Fb)
	case CK_nBs:
		return -__.Bb
	case CK_nDs:
		return -__.Db
	case CK_nFs:
		return -__.Fb
	case CK_nUs:
		return -__.Ub
	case CK_nDBs:
		return -Min(__.Db, __.Bb)
	case CK_nUBs:
		return -Min(__.Ub, __.Bb)
	case CK_nDFs:
		return -Min(__.Db, __.Fb)
	case CK_nUFs:
		return -Min(__.Ub, __.Fb)
	case CK_na:
		return -__.ab
	case CK_nb:
		return -__.bb
	case CK_nc:
		return -__.cb
	case CK_nx:
		return -__.xb
	case CK_ny:
		return -__.yb
	case CK_nz:
		return -__.zb
	case CK_ns:
		return -__.sb
	case CK_nv:
		return -__.vb
	case CK_nw:
		return -__.wb
	}
	return 0
}
func (__ *CommandBuffer) State2(ck CommandKey) int32 {
	f := func(a, b, c int32) int32 {
		switch {
		case a > 0:
			return -Max(b, c)
		case b > 0:
			return -Max(a, c)
		case c > 0:
			return -Max(a, b)
		}
		return -Max(a, b, c)
	}
	switch ck {
	case CK_Bs:
		if __.Bb < 0 {
			return __.Bb
		}
		return Min(Abs(__.Bb), Abs(__.Db), Abs(__.Ub))
	case CK_Ds:
		if __.Db < 0 {
			return __.Db
		}
		return Min(Abs(__.Db), Abs(__.Bb), Abs(__.Fb))
	case CK_Fs:
		if __.Fb < 0 {
			return __.Fb
		}
		return Min(Abs(__.Fb), Abs(__.Db), Abs(__.Ub))
	case CK_Us:
		if __.Ub < 0 {
			return __.Ub
		}
		return Min(Abs(__.Ub), Abs(__.Bb), Abs(__.Fb))
	//MUGENだと斜め入力に$を入れても意味がない
	//case CK_DBs:
	//	if s := __.State(CK_DBs); s < 0 {
	//		return s
	//	}
	//	return Min(Abs(__.Db), Abs(__.Bb))
	//case CK_UBs:
	//	if s := __.State(CK_UBs); s < 0 {
	//		return s
	//	}
	//	return Min(Abs(__.Ub), Abs(__.Bb))
	//case CK_DFs:
	//	if s := __.State(CK_DFs); s < 0 {
	//		return s
	//	}
	//	return Min(Abs(__.Db), Abs(__.Fb))
	//case CK_UFs:
	//	if s := __.State(CK_UFs); s < 0 {
	//		return s
	//	}
	//	return Min(Abs(__.Ub), Abs(__.Fb))
	case CK_nBs:
		return f(__.State(CK_B), __.State(CK_UB), __.State(CK_DB))
	case CK_nDs:
		return f(__.State(CK_D), __.State(CK_DB), __.State(CK_DF))
	case CK_nFs:
		return f(__.State(CK_F), __.State(CK_DF), __.State(CK_UF))
	case CK_nUs:
		return f(__.State(CK_U), __.State(CK_UB), __.State(CK_UF))
		//case CK_nDBs:
		//	return f(__.State(CK_DB), __.State(CK_D), __.State(CK_B))
		//case CK_nUBs:
		//	return f(__.State(CK_UB), __.State(CK_U), __.State(CK_B))
		//case CK_nDFs:
		//	return f(__.State(CK_DF), __.State(CK_D), __.State(CK_F))
		//case CK_nUFs:
		//	return f(__.State(CK_UF), __.State(CK_U), __.State(CK_F))
	}
	return __.State(ck)
}
func (__ *CommandBuffer) LastDirectionTime() int32 {
	return Min(Abs(__.Bb), Abs(__.Db), Abs(__.Fb), Abs(__.Ub))
}
func (__ *CommandBuffer) LastChangeTime() int32 {
	return Min(__.LastDirectionTime(), Abs(__.ab), Abs(__.bb), Abs(__.cb),
		Abs(__.xb), Abs(__.yb), Abs(__.zb), Abs(__.sb), Abs(__.vb), Abs(__.wb))
}

type NetBuffer struct {
	buf              [32]InputBits
	curT, inpT, senT int32
}

func (nb *NetBuffer) reset(time int32) {
	nb.curT, nb.inpT, nb.senT = time, time, time
}
func (nb *NetBuffer) localUpdate(in int) {
	if nb.inpT-nb.curT < 32 {
		nb.buf[nb.inpT&31].SetInput(in)
		nb.inpT++
	}
}
func (nb *NetBuffer) input(cb *CommandBuffer, f int32) {
	if nb.curT < nb.inpT {
		nb.buf[nb.curT&31].GetInput(cb, f)
	}
}

type NetInput struct {
	ln         *net.TCPListener
	conn       *net.TCPConn
	st         NetState
	sendEnd    chan bool
	recvEnd    chan bool
	buf        [MaxSimul*2 + MaxAttachedChar]NetBuffer
	locIn      int
	remIn      int
	time       int32
	stoppedcnt int32
	delay      int32
	rep        *os.File
	host       bool
}

func NewNetInput(replayfile string) *NetInput {
	ni := &NetInput{st: NS_Stop,
		sendEnd: make(chan bool, 1), recvEnd: make(chan bool, 1)}
	ni.sendEnd <- true
	ni.recvEnd <- true
	ni.rep, _ = os.Create(replayfile)
	return ni
}
func (ni *NetInput) Close() {
	if ni.rep != nil {
		ni.rep.Close()
		ni.rep = nil
	}
	if ni.ln != nil {
		ni.ln.Close()
		ni.ln = nil
	}
	if ni.conn != nil {
		ni.conn.Close()
	}
	if ni.sendEnd != nil {
		<-ni.sendEnd
		close(ni.sendEnd)
		ni.sendEnd = nil
	}
	if ni.recvEnd != nil {
		<-ni.recvEnd
		close(ni.recvEnd)
		ni.recvEnd = nil
	}
	ni.conn = nil
}
func (ni *NetInput) GetHostGuestRemap() (host, guest int) {
	host, guest = -1, -1
	for i, c := range sys.com {
		if c == 0 {
			if host < 0 {
				host = i
			} else if guest < 0 {
				guest = i
			}
		}
	}
	if host < 0 {
		host = 0
	}
	if guest < 0 {
		guest = (host + 1) % len(ni.buf)
	}
	return
}
func (ni *NetInput) Accept(port string) error {
	if ln, err := net.Listen("tcp", ":"+port); err != nil {
		return err
	} else {
		ni.ln = ln.(*net.TCPListener)
		ni.host = true
		ni.locIn, ni.remIn = ni.GetHostGuestRemap()
		go func() {
			ln := ni.ln
			if conn, err := ln.AcceptTCP(); err == nil {
				ni.conn = conn
			}
			ln.Close()
		}()
	}
	return nil
}
func (ni *NetInput) Connect(server, port string) {
	ni.host = false
	ni.remIn, ni.locIn = ni.GetHostGuestRemap()
	go func() {
		if conn, err := net.Dial("tcp", server+":"+port); err == nil {
			ni.conn = conn.(*net.TCPConn)
		}
	}()
}
func (ni *NetInput) IsConnected() bool {
	return ni != nil && ni.conn != nil
}
func (ni *NetInput) Input(cb *CommandBuffer, i int, facing int32) {
	if i >= 0 && i < len(ni.buf) {
		ni.buf[sys.inputRemap[i]].input(cb, facing)
	}
}
func (ni *NetInput) AnyButton() bool {
	for _, nb := range ni.buf {
		if nb.buf[nb.curT&31]&IB_anybutton != 0 {
			return true
		}
	}
	return false
}
func (ni *NetInput) Stop() {
	if sys.esc {
		ni.end()
	} else {
		if ni.st != NS_End && ni.st != NS_Error {
			ni.st = NS_Stop
		}
		<-ni.sendEnd
		ni.sendEnd <- true
		<-ni.recvEnd
		ni.recvEnd <- true
	}
}
func (ni *NetInput) end() {
	if ni.st != NS_Error {
		ni.st = NS_End
	}
	ni.Close()
}
func (ni *NetInput) readI32() (int32, error) {
	b := [4]byte{}
	if _, err := ni.conn.Read(b[:]); err != nil {
		return 0, err
	}
	return int32(b[0]) | int32(b[1])<<8 | int32(b[2])<<16 | int32(b[3])<<24, nil
}
func (ni *NetInput) writeI32(i32 int32) error {
	b := [...]byte{byte(i32), byte(i32 >> 8), byte(i32 >> 16), byte(i32 >> 24)}
	if _, err := ni.conn.Write(b[:]); err != nil {
		return err
	}
	return nil
}
func (ni *NetInput) Synchronize() error {
	if !ni.IsConnected() || ni.st == NS_Error {
		return Error("接続がありません。")
	}
	ni.Stop()
	var seed int32
	if ni.host {
		seed = Random()
		if err := ni.writeI32(seed); err != nil {
			return err
		}
	} else {
		var err error
		if seed, err = ni.readI32(); err != nil {
			return err
		}
	}
	Srand(seed)
	binary.Write(ni.rep, binary.LittleEndian, &seed)
	if err := ni.writeI32(ni.time); err != nil {
		return err
	}
	if tmp, err := ni.readI32(); err != nil {
		return err
	} else if tmp != ni.time {
		return Error("同期エラーです。")
	}
	ni.buf[ni.locIn].reset(ni.time)
	ni.buf[ni.remIn].reset(ni.time)
	ni.st = NS_Playing
	<-ni.sendEnd
	go func(nb *NetBuffer) {
		defer func() { ni.sendEnd <- true }()
		for ni.st == NS_Playing {
			if nb.senT < nb.inpT {
				if err := ni.writeI32(int32(nb.buf[nb.senT&31])); err != nil {
					ni.st = NS_Error
					return
				}
				nb.senT++
			}
			time.Sleep(time.Millisecond)
		}
		ni.writeI32(-1)
	}(&ni.buf[ni.locIn])
	<-ni.recvEnd
	go func(nb *NetBuffer) {
		defer func() { ni.recvEnd <- true }()
		for ni.st == NS_Playing {
			if nb.inpT-nb.curT < 32 {
				if tmp, err := ni.readI32(); err != nil {
					ni.st = NS_Error
					return
				} else {
					nb.buf[nb.inpT&31] = InputBits(tmp)
					if tmp < 0 {
						ni.st = NS_Stopped
						return
					} else {
						nb.inpT++
						nb.senT = nb.inpT
					}
				}
			}
			time.Sleep(time.Millisecond)
		}
		for tmp := int32(0); tmp != -1; {
			var err error
			if tmp, err = ni.readI32(); err != nil {
				break
			}
		}
	}(&ni.buf[ni.remIn])
	ni.Update()
	return nil
}
func (ni *NetInput) Update() bool {
	if ni.st != NS_Stopped {
		ni.stoppedcnt = 0
	}
	if !sys.gameEnd {
		switch ni.st {
		case NS_Stopped:
			ni.stoppedcnt++
			if ni.stoppedcnt > 60 {
				ni.st = NS_End
				break
			}
			fallthrough
		case NS_Playing:
			for {
				foo := Min(ni.buf[ni.locIn].senT, ni.buf[ni.remIn].senT)
				tmp := ni.buf[ni.remIn].inpT + ni.delay>>3 - ni.buf[ni.locIn].inpT
				if tmp >= 0 {
					ni.buf[ni.locIn].localUpdate(0)
					if ni.delay > 0 {
						ni.delay--
					}
				} else if tmp < -1 {
					ni.delay += 4
				}
				if ni.time >= foo {
					if sys.esc || !sys.await(FPS) || ni.st != NS_Playing {
						break
					}
					continue
				}
				ni.buf[ni.locIn].curT = ni.time
				ni.buf[ni.remIn].curT = ni.time
				for _, nb := range ni.buf {
					binary.Write(ni.rep, binary.LittleEndian, &nb.buf[ni.time&31])
				}
				ni.time++
				if ni.time >= foo {
					ni.buf[ni.locIn].localUpdate(0)
				}
				break
			}
		case NS_End, NS_Error:
			sys.esc = true
		}
	}
	if sys.esc {
		ni.end()
	}
	return !sys.gameEnd
}

type FileInput struct {
	f  *os.File
	ib [MaxSimul * 2]InputBits
}

func OpenFileInput(filename string) *FileInput {
	fi := &FileInput{}
	fi.f, _ = os.Open(filename)
	return fi
}
func (fi *FileInput) Close() {
	if fi.f != nil {
		fi.f.Close()
		fi.f = nil
	}
}
func (fi *FileInput) Input(cb *CommandBuffer, i int, facing int32) {
	if i >= 0 && i < len(fi.ib) {
		fi.ib[sys.inputRemap[i]].GetInput(cb, facing)
	}
}
func (fi *FileInput) AnyButton() bool {
	for _, b := range fi.ib {
		if b&IB_anybutton != 0 {
			return true
		}
	}
	return false
}
func (fi *FileInput) Synchronize() {
	if fi.f != nil {
		var seed int32
		if binary.Read(fi.f, binary.LittleEndian, &seed) == nil {
			Srand(seed)
			fi.Update()
		}
	}
}
func (fi *FileInput) Update() bool {
	if fi.f == nil {
		sys.esc = true
	} else {
		if sys.oldNextAddTime > 0 &&
			binary.Read(fi.f, binary.LittleEndian, fi.ib[:]) != nil {
			sys.esc = true
		}
		if sys.esc {
			fi.Close()
		}
	}
	return !sys.gameEnd
}

type AiInput struct {
	dir, dt, at, bt, ct, xt, yt, zt, st, vt, wt int32
}

func (__ *AiInput) Update(level float32) {
	if sys.intro != 0 {
		__.dt, __.at, __.bt, __.ct = 0, 0, 0, 0
		__.xt, __.yt, __.zt, __.st = 0, 0, 0, 0
		__.vt, __.wt = 0, 0
		return
	}
	var osu, hanasu int32 = 15, 60
	dec := func(t *int32) bool {
		(*t)--
		if *t <= 0 {
			// TODO: Balance AI Scaling
			if Rand(1, osu) == 1 {
				*t = Rand(1, hanasu)
				return true
			}
			*t = 0
		}
		return false
	}
	if dec(&__.dt) {
		__.dir = Rand(0, 7)
	}
	osu, hanasu = int32(-11.25*level+165), 30
	dec(&__.at)
	dec(&__.bt)
	dec(&__.ct)
	dec(&__.xt)
	dec(&__.yt)
	dec(&__.zt)
	dec(&__.vt)
	dec(&__.wt)
	osu = 3600
	dec(&__.st)
}
func (__ *AiInput) L() bool {
	return __.dt != 0 && (__.dir == 5 || __.dir == 6 || __.dir == 7)
}
func (__ *AiInput) R() bool {
	return __.dt != 0 && (__.dir == 1 || __.dir == 2 || __.dir == 3)
}
func (__ *AiInput) U() bool {
	return __.dt != 0 && (__.dir == 7 || __.dir == 0 || __.dir == 1)
}
func (__ *AiInput) D() bool {
	return __.dt != 0 && (__.dir == 3 || __.dir == 4 || __.dir == 5)
}
func (__ *AiInput) A() bool {
	return __.at != 0
}
func (__ *AiInput) B() bool {
	return __.bt != 0
}
func (__ *AiInput) C() bool {
	return __.ct != 0
}
func (__ *AiInput) X() bool {
	return __.xt != 0
}
func (__ *AiInput) Y() bool {
	return __.yt != 0
}
func (__ *AiInput) Z() bool {
	return __.zt != 0
}
func (__ *AiInput) S() bool {
	return __.st != 0
}
func (__ *AiInput) V() bool {
	return __.vt != 0
}
func (__ *AiInput) W() bool {
	return __.wt != 0
}

type cmdElem struct {
	key                       []CommandKey
	tametime                  int32
	slash, greater, direction bool
}

func (ce *cmdElem) IsDirection() bool {
	return !ce.slash && len(ce.key) == 1 && ce.key[0] < CK_a
}
func (ce *cmdElem) IsDToB(next cmdElem) bool {
	if next.slash {
		return false
	}
	btn := true
	for _, k := range ce.key {
		if k < CK_a {
			btn = false
			break
		}
	}
	if btn {
		return false
	}
	if len(ce.key) != len(next.key) {
		return true
	}
	for i, k := range ce.key {
		if k != next.key[i] && ((k < CK_nB || k > CK_nUF) &&
			(k < CK_nBs || k > CK_nUFs) ||
			(next.key[i] < CK_nB || next.key[i] > CK_nUF) &&
				(next.key[i] < CK_nBs || next.key[i] > CK_nUFs)) {
			return true
		}
	}
	return false
}

type Command struct {
	name                string
	hold                [][]CommandKey
	held                []bool
	cmd                 []cmdElem
	cmdi                [3]int
	cur                 [3]int32
	tamei               int
	time                int32
	buftime, curbuftime int32
}

func newCommand() *Command { return &Command{tamei: -1, time: 1, buftime: 1} }
func ReadCommand(name, cmdstr string, kr *CommandKeyRemap) (*Command, error) {
	c := newCommand()
	c.name = name
	cmd := strings.Split(cmdstr, ",")
	for _, cestr := range cmd {
		if len(c.cmd) > 0 && c.cmd[len(c.cmd)-1].slash {
			c.hold = append(c.hold, c.cmd[len(c.cmd)-1].key)
			c.cmd[len(c.cmd)-1] = cmdElem{tametime: 1}
		} else {
			c.cmd = append(c.cmd, cmdElem{tametime: 1})
		}
		ce := &c.cmd[len(c.cmd)-1]
		cestr = strings.TrimSpace(cestr)
		getChar := func() rune {
			if len(cestr) > 0 {
				return rune(cestr[0])
			}
			return rune(-1)
		}
		nextChar := func() rune {
			if len(cestr) > 0 {
				cestr = strings.TrimSpace(cestr[1:])
			}
			return getChar()
		}
		tilde := false
		switch getChar() {
		case '>':
			ce.greater = true
			r := nextChar()
			if r == '/' {
				ce.slash = true
				nextChar()
				break
			} else if r == '~' {
			} else {
				break
			}
			fallthrough
		case '~':
			tilde = true
			n := int32(0)
			for r := nextChar(); '0' <= r && r <= '9'; r = nextChar() {
				n = n*10 + int32(r-'0')
			}
			if n > 0 {
				ce.tametime = n
			}
		case '/':
			ce.slash = true
			nextChar()
		}
		for len(cestr) > 0 {
			switch getChar() {
			case 'B':
				if tilde {
					ce.key = append(ce.key, CK_nB)
				} else {
					ce.key = append(ce.key, CK_B)
				}
				tilde = false
			case 'D':
				if len(cestr) > 1 && cestr[1] == 'B' {
					nextChar()
					if tilde {
						ce.key = append(ce.key, CK_nDB)
					} else {
						ce.key = append(ce.key, CK_DB)
					}
				} else if len(cestr) > 1 && cestr[1] == 'F' {
					nextChar()
					if tilde {
						ce.key = append(ce.key, CK_nDF)
					} else {
						ce.key = append(ce.key, CK_DF)
					}
				} else {
					if tilde {
						ce.key = append(ce.key, CK_nD)
					} else {
						ce.key = append(ce.key, CK_D)
					}
				}
				tilde = false
			case 'F':
				if tilde {
					ce.key = append(ce.key, CK_nF)
				} else {
					ce.key = append(ce.key, CK_F)
				}
				tilde = false
			case 'U':
				if len(cestr) > 1 && cestr[1] == 'B' {
					nextChar()
					if tilde {
						ce.key = append(ce.key, CK_nUB)
					} else {
						ce.key = append(ce.key, CK_UB)
					}
				} else if len(cestr) > 1 && cestr[1] == 'F' {
					nextChar()
					if tilde {
						ce.key = append(ce.key, CK_nUF)
					} else {
						ce.key = append(ce.key, CK_UF)
					}
				} else {
					if tilde {
						ce.key = append(ce.key, CK_nU)
					} else {
						ce.key = append(ce.key, CK_U)
					}
				}
				tilde = false
			case 'a':
				if tilde {
					ce.key = append(ce.key, kr.na)
				} else {
					ce.key = append(ce.key, kr.a)
				}
				tilde = false
			case 'b':
				if tilde {
					ce.key = append(ce.key, kr.nb)
				} else {
					ce.key = append(ce.key, kr.b)
				}
				tilde = false
			case 'c':
				if tilde {
					ce.key = append(ce.key, kr.nc)
				} else {
					ce.key = append(ce.key, kr.c)
				}
				tilde = false
			case 'x':
				if tilde {
					ce.key = append(ce.key, kr.nx)
				} else {
					ce.key = append(ce.key, kr.x)
				}
				tilde = false
			case 'y':
				if tilde {
					ce.key = append(ce.key, kr.ny)
				} else {
					ce.key = append(ce.key, kr.y)
				}
				tilde = false
			case 'z':
				if tilde {
					ce.key = append(ce.key, kr.nz)
				} else {
					ce.key = append(ce.key, kr.z)
				}
				tilde = false
			case 's':
				if tilde {
					ce.key = append(ce.key, kr.ns)
				} else {
					ce.key = append(ce.key, kr.s)
				}
				tilde = false
			case 'v':
				if tilde {
					ce.key = append(ce.key, kr.nv)
				} else {
					ce.key = append(ce.key, kr.v)
				}
				tilde = false
			case 'd':
				if tilde {
					ce.key = append(ce.key, kr.nv)
				} else {
					ce.key = append(ce.key, kr.v)
				}
				tilde = false
			case 'w':
				if tilde {
					ce.key = append(ce.key, kr.nw)
				} else {
					ce.key = append(ce.key, kr.w)
				}
				tilde = false
			case '$':
				switch nextChar() {
				case 'B':
					if tilde {
						ce.key = append(ce.key, CK_nBs)
					} else {
						ce.key = append(ce.key, CK_Bs)
					}
					tilde = false
				case 'D':
					if len(cestr) > 1 && cestr[1] == 'B' {
						nextChar()
						if tilde {
							ce.key = append(ce.key, CK_nDBs)
						} else {
							ce.key = append(ce.key, CK_DBs)
						}
					} else if len(cestr) > 1 && cestr[1] == 'F' {
						nextChar()
						if tilde {
							ce.key = append(ce.key, CK_nDFs)
						} else {
							ce.key = append(ce.key, CK_DFs)
						}
					} else {
						if tilde {
							ce.key = append(ce.key, CK_nDs)
						} else {
							ce.key = append(ce.key, CK_Ds)
						}
					}
					tilde = false
				case 'F':
					if tilde {
						ce.key = append(ce.key, CK_nFs)
					} else {
						ce.key = append(ce.key, CK_Fs)
					}
					tilde = false
				case 'U':
					if len(cestr) > 1 && cestr[1] == 'B' {
						nextChar()
						if tilde {
							ce.key = append(ce.key, CK_nUBs)
						} else {
							ce.key = append(ce.key, CK_UBs)
						}
					} else if len(cestr) > 1 && cestr[1] == 'F' {
						nextChar()
						if tilde {
							ce.key = append(ce.key, CK_nUFs)
						} else {
							ce.key = append(ce.key, CK_UFs)
						}
					} else {
						if tilde {
							ce.key = append(ce.key, CK_nUs)
						} else {
							ce.key = append(ce.key, CK_Us)
						}
					}
					tilde = false
				default:
					// error
					continue
				}
			case '~':
				tilde = true
			case '+':
			default:
				// error
			}
			nextChar()
		}
		if len(c.cmd) >= 2 && ce.IsDirection() &&
			c.cmd[len(c.cmd)-2].IsDirection() {
			ce.direction = true
		}
	}
	if c.cmd[len(c.cmd)-1].slash {
		c.hold = append(c.hold, c.cmd[len(c.cmd)-1].key)
	}
	c.held = make([]bool, len(c.hold))
	return c, nil
}
func (c *Command) Clear() {
	c.tamei, c.curbuftime = -1, 0
	c.cmdi[0], c.cur[0] = 0, 0
	c.cmdi[1], c.cur[1] = 0, 0
	c.cmdi[2], c.cur[2] = 0, 0
	for i := range c.held {
		c.held[i] = false
	}
}
func (c *Command) ClearEach(bufline int) {
	c.cmdi[bufline], c.cur[bufline] = 0, 0
}

// AI level stuff here
func (c *Command) bufTest(cbuf *CommandBuffer, ai bool,
	holdTemp *[CK_Last + 1]bool, bufline int) bool {
	anyHeld, notHeld := false, 0
	if len(c.hold) > 0 && !ai {
		if holdTemp == nil {
			holdTemp = &[CK_Last + 1]bool{}
			for i := range *holdTemp {
				(*holdTemp)[i] = true
			}
		}
		allHold := true
		for i, h := range c.hold {
			func() {
				for _, k := range h {
					ks := cbuf.State(k)
					if ks == 1 && (c.cmdi[bufline] > 0 || len(c.hold) > 1) && !c.held[i] &&
						(*holdTemp)[int(k)] {
						c.held[i], (*holdTemp)[int(k)] = true, false
					}
					if ks > 0 {
						return
					}
				}
				allHold = false
			}()
			if c.held[i] {
				anyHeld = true
			} else {
				notHeld += 1
			}
		}
		if c.cmdi[bufline] == len(c.cmd)-1 && (!allHold || notHeld > 1) {
			return anyHeld || c.cmdi[bufline] > 0
		}
	}
	if !ai && c.cmd[c.cmdi[bufline]].slash {
		if c.cmdi[bufline] > 0 {
			if notHeld == 1 {
				if len(c.cmd[c.cmdi[bufline]-1].key) != 1 {
					return false
				}
				if CK_a <= c.cmd[c.cmdi[bufline]-1].key[0] && c.cmd[c.cmdi[bufline]-1].key[0] <= CK_s {
					ks := cbuf.State(c.cmd[c.cmdi[bufline]-1].key[0])
					if 0 < ks && ks <= cbuf.LastDirectionTime() {
						return true
					}
				}
			} else if len(c.cmd[c.cmdi[bufline]-1].key) > 1 {
				for _, k := range c.cmd[c.cmdi[bufline]-1].key {
					if CK_a <= k && k <= CK_s && cbuf.State(k) > 0 {
						return false
					}
				}
			}
		}
		c.cmdi[bufline]++
		return true
	}
	fail := func() bool {
		if c.cmdi[bufline] == 0 {
			return anyHeld
		}
		if !ai && (c.cmd[c.cmdi[bufline]].greater || c.cmd[c.cmdi[bufline]].direction) {
			var t int32
			if c.cmd[c.cmdi[bufline]].greater {
				t = cbuf.LastChangeTime()
			} else {
				t = cbuf.LastDirectionTime()
			}
			for _, k := range c.cmd[c.cmdi[bufline]-1].key {
				if Abs(cbuf.State2(k)) == t {
					return true
				}
			}
			c.ClearEach(bufline)
			return c.bufTest(cbuf, ai, holdTemp, bufline)
		}
		return true
	}
	if c.tamei != c.cmdi[bufline] {
		if c.cmd[c.cmdi[bufline]].tametime > 1 {
			for _, k := range c.cmd[c.cmdi[bufline]].key {
				ks := cbuf.State(k)
				if ks > 0 {
					return ai
				}
				if func() bool {
					if ai {
						return Rand(0, c.cmd[c.cmdi[bufline]].tametime) != 0
					}
					return -ks < c.cmd[c.cmdi[bufline]].tametime
				}() {
					return anyHeld || c.cmdi[bufline] > 0
				}
			}
			c.tamei = c.cmdi[bufline]
		} else if c.cmdi[bufline] > 0 && len(c.cmd[c.cmdi[bufline]-1].key) == 1 &&
			len(c.cmd[c.cmdi[bufline]].key) == 1 && c.cmd[c.cmdi[bufline]-1].key[0] < CK_Bs &&
			c.cmd[c.cmdi[bufline]].key[0] < CK_nB && (c.cmd[c.cmdi[bufline]-1].key[0]-
			c.cmd[c.cmdi[bufline]].key[0])&7 == 0 {
			if cbuf.B < 0 && cbuf.D < 0 && cbuf.F < 0 && cbuf.U < 0 {
				c.tamei = c.cmdi[bufline]
			} else {
				return fail()
			}
		}
	}
	foo := false
	for _, k := range c.cmd[c.cmdi[bufline]].key {
		n := cbuf.State2(k)
		if c.cmd[c.cmdi[bufline]].slash {
			foo = foo || n > 0
		} else if n < 1 || 7 < n {
			return fail()
		} else {
			foo = foo || n == 1
		}
	}
	if !foo {
		return fail()
	}
	c.cmdi[bufline]++
	if c.cmdi[bufline] < len(c.cmd) && c.cmd[c.cmdi[bufline]-1].IsDToB(c.cmd[c.cmdi[bufline]]) {
		return c.bufTest(cbuf, ai, holdTemp, bufline)
	}
	return true
}

// AI level stuff here
func (c *Command) Step(cbuf *CommandBuffer, ai, hitpause bool, buftime int32) {
	if !hitpause && c.curbuftime > 0 {
		c.curbuftime--
	}
	if len(c.cmd) == 0 {
		return
	}
	ocbt := c.curbuftime
	defer func() {
		if c.curbuftime < ocbt {
			c.curbuftime = ocbt
		}
	}()
	complete := false
	bufContinue := false
	firstCmdInput := 0
	var holdTemp *[CK_Last + 1]bool
	for bufline, cmditemp := range c.cmdi {
		if cbuf == nil || !c.bufTest(cbuf, ai, holdTemp, bufline) {
			foo := c.tamei == 0 && c.cmdi[bufline] == 0
			c.ClearEach(bufline)
			if foo {
				c.tamei = 0
			}
			if len(c.cmdi) == bufline+1 {
				return
			}
			continue
		}
		if c.cmdi[bufline] == 1 && c.cmd[0].slash {
			c.cur[bufline] = 0
		} else {
			c.cur[bufline]++
		}
		if c.cur[bufline] <= c.time {
			bufContinue = true
		} else {
			c.ClearEach(bufline)
		}
		if firstCmdInput > 0 && firstCmdInput == c.cmdi[bufline] {
			c.cmdi[bufline] = cmditemp
		}
		if c.cmdi[bufline] == len(c.cmd) {
			complete = true
			break
		}
		if cmditemp == 0 && c.cmdi[bufline] > 0 {
			firstCmdInput = c.cmdi[bufline]
		}
	}
	if !complete && (ai || bufContinue) {
		return
	}
	c.Clear()
	if complete {
		c.curbuftime = c.buftime + buftime
	}
}

type CommandList struct {
	Buffer            *CommandBuffer
	Names             map[string]int
	Commands          [][]Command
	DefaultTime       int32
	DefaultBufferTime int32
}

func NewCommandList(cb *CommandBuffer) *CommandList {
	return &CommandList{Buffer: cb, Names: make(map[string]int),
		DefaultTime: 15, DefaultBufferTime: 1}
}
func (cl *CommandList) Input(i int, facing int32, aiLevel float32) bool {
	if cl.Buffer == nil {
		return false
	}
	step := cl.Buffer.Bb != 0
	if i < 0 && ^i < len(sys.aiInput) {
		sys.aiInput[^i].Update(aiLevel) // 乱数を使うので同期がずれないようここで / Here we use random numbers so we can not get out of sync
	}
	_else := i < 0
	if _else {
	} else if sys.fileInput != nil {
		sys.fileInput.Input(cl.Buffer, i, facing)
	} else if sys.netInput != nil {
		sys.netInput.Input(cl.Buffer, i, facing)
	} else {
		_else = true
	}
	if _else {
		var l, r, u, d, a, b, c, x, y, z, s, v, w bool
		if i < 0 {
			i = ^i
			if i < len(sys.aiInput) {
				l = sys.aiInput[i].L()
				r = sys.aiInput[i].R()
				u = sys.aiInput[i].U()
				d = sys.aiInput[i].D()
				a = sys.aiInput[i].A()
				b = sys.aiInput[i].B()
				c = sys.aiInput[i].C()
				x = sys.aiInput[i].X()
				y = sys.aiInput[i].Y()
				z = sys.aiInput[i].Z()
				s = sys.aiInput[i].S()
				v = sys.aiInput[i].V()
				w = sys.aiInput[i].W()
			}
		} else if i < len(sys.inputRemap) {
			in := sys.inputRemap[i]
			if in < len(sys.keyConfig) {
				joy := sys.keyConfig[in].Joy
				if joy == -1 {
					l = sys.keyConfig[in].L()
					r = sys.keyConfig[in].R()
					u = sys.keyConfig[in].U()
					d = sys.keyConfig[in].D()
					a = sys.keyConfig[in].A()
					b = sys.keyConfig[in].B()
					c = sys.keyConfig[in].C()
					x = sys.keyConfig[in].X()
					y = sys.keyConfig[in].Y()
					z = sys.keyConfig[in].Z()
					s = sys.keyConfig[in].S()
					v = sys.keyConfig[in].V()
					w = sys.keyConfig[in].W()
				}
			}
			if in < len(sys.JoystickConfig) {
				joyS := sys.JoystickConfig[in].Joy
				if joyS >= 0 {
					if l == false {
						l = sys.JoystickConfig[in].L()
					}
					if r == false {
						r = sys.JoystickConfig[in].R()
					}
					if u == false {
						u = sys.JoystickConfig[in].U()
					}
					if d == false {
						d = sys.JoystickConfig[in].D()
					}
					if a == false {
						a = sys.JoystickConfig[in].A()
					}
					if b == false {
						b = sys.JoystickConfig[in].B()
					}
					if c == false {
						c = sys.JoystickConfig[in].C()
					}
					if x == false {
						x = sys.JoystickConfig[in].X()
					}
					if y == false {
						y = sys.JoystickConfig[in].Y()
					}
					if z == false {
						z = sys.JoystickConfig[in].Z()
					}
					if s == false {
						s = sys.JoystickConfig[in].S()
					}
					if v == false {
						v = sys.JoystickConfig[in].V()
					}
					if w == false {
						w = sys.JoystickConfig[in].W()
					}
				}
			}
		}
		var B, F bool
		if facing < 0 {
			B, F = r, l
		} else {
			B, F = l, r
		}
		cl.Buffer.Input(B, d, F, u, a, b, c, x, y, z, s, v, w)
	}
	return step
}

// AI level stuff here
func (cl *CommandList) Step(facing int32, ai, hitpause bool,
	buftime int32) {
	if cl.Buffer != nil {
		for i := range cl.Commands {
			for j := range cl.Commands[i] {
				cl.Commands[i][j].Step(cl.Buffer, ai, hitpause, buftime)
			}
		}
	}
}
func (cl *CommandList) BufReset() {
	if cl.Buffer != nil {
		cl.Buffer.Reset()
		for i := range cl.Commands {
			for j := range cl.Commands[i] {
				cl.Commands[i][j].Clear()
			}
		}
	}
}
func (cl *CommandList) Add(c Command) {
	i, ok := cl.Names[c.name]
	if !ok || i < 0 || i >= len(cl.Commands) {
		i = len(cl.Commands)
		cl.Commands = append(cl.Commands, nil)
	}
	cl.Commands[i] = append(cl.Commands[i], c)
	cl.Names[c.name] = i
}
func (cl *CommandList) At(i int) []Command {
	if i < 0 || i >= len(cl.Commands) {
		return nil
	}
	return cl.Commands[i]
}
func (cl *CommandList) Get(name string) []Command {
	i, ok := cl.Names[name]
	if !ok {
		return nil
	}
	return cl.At(i)
}
func (cl *CommandList) GetState(name string) bool {
	for _, c := range cl.Get(name) {
		if c.curbuftime > 0 {
			return true
		}
	}
	return false
}
func (cl *CommandList) CopyList(src CommandList) {
	cl.Names = src.Names
	cl.Commands = make([][]Command, len(src.Commands))
	for i, ca := range src.Commands {
		cl.Commands[i] = make([]Command, len(ca))
		copy(cl.Commands[i], ca)
		for j, c := range ca {
			cl.Commands[i][j].held = make([]bool, len(c.held))
		}
	}
}
