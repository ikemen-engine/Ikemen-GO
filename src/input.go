package main

import (
	"encoding/binary"
	"net"
	"os"
	"strings"
	"time"

	"github.com/go-gl/glfw/v3.3/glfw"
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
	CK_d
	CK_w
	CK_m
	CK_na
	CK_nb
	CK_nc
	CK_nx
	CK_ny
	CK_nz
	CK_ns
	CK_nd
	CK_nw
	CK_nm
	CK_Last = CK_nm
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
	Pause    bool
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
		sys.keyState[key] = false
		sys.keyInput = glfw.KeyUnknown
		sys.keyString = ""
	case glfw.Press:
		sys.keyState[key] = true
		sys.keyInput = key
		if key == glfw.KeyEscape && mk&(glfw.ModControl|glfw.ModAlt) == 0 {
			sys.esc = true
			if sys.netInput != nil || len(sys.commonLua) == 0 || sys.gameMode == "" {
				sys.endMatch = true
			}
		}
		for k, v := range sys.shortcutScripts {
			if sys.netInput == nil && (!sys.paused || sys.step || v.Pause) {
				v.Activate = v.Activate || k.Test(key, mk)
			}
		}
		if key == glfw.KeyF12 {
			captureScreen()
		}
		if key == glfw.KeyEnter && mk&(glfw.ModAlt) != 0 {
			sys.window.toggleFullscreen()
		}
	}
}

func charCallback(_ *glfw.Window, char rune, mk glfw.ModifierKey) {
	sys.keyString = string(char)
}

/* TODO: Why this did exist?
func joystickCallback(joy, event glfw.PeripheralEvent) {
	if event == glfw.Connected {
		// The joystick was connected
	} else if event == glfw.Disconnected {
		// The joystick was disconnected
	}
}*/

var joystick = [...]glfw.Joystick{glfw.Joystick1, glfw.Joystick2,
	glfw.Joystick3, glfw.Joystick4, glfw.Joystick5, glfw.Joystick6,
	glfw.Joystick7, glfw.Joystick8, glfw.Joystick9, glfw.Joystick10,
	glfw.Joystick11, glfw.Joystick12, glfw.Joystick13, glfw.Joystick14,
	glfw.Joystick15, glfw.Joystick16}

func JoystickState(joy, button int) bool {
	if joy < 0 {
		return sys.keyState[glfw.Key(button)]
	}
	if joy >= len(joystick) {
		return false
	}
	btns := joystick[joy].GetButtons()
	if button < 0 {
		button = -button - 1
		axes := joystick[joy].GetAxes()

		if len(axes)*2 <= button {
			return false
		}

		var joyName = joystick[joy].GetGamepadName()

		//Xbox360コントローラーのLRトリガー判定
		if (button == 9 || button == 11) && (strings.Contains(joyName, "XInput") || strings.Contains(joyName, "X360")) {
			return axes[button/2] > sys.xinputTriggerSensitivity
		}

		// Ignore trigger axis on PS4 (We already have buttons)
		if (button >= 6 && button <= 9) && joyName == "PS4 Controller" {
			return false
		}

		switch button & 1 {
		case 0:
			return axes[button/2] < -sys.controllerStickSensitivity
		case 1:
			return axes[button/2] > sys.controllerStickSensitivity
		}
	}
	if len(btns) <= button {
		return false
	}
	return btns[button] != 0
}

type KeyConfig struct{ Joy, dU, dD, dL, dR, kA, kB, kC, kX, kY, kZ, kS, kD, kW, kM int }

func (kc KeyConfig) U() bool { return JoystickState(kc.Joy, kc.dU) }
func (kc KeyConfig) D() bool { return JoystickState(kc.Joy, kc.dD) }
func (kc KeyConfig) L() bool { return JoystickState(kc.Joy, kc.dL) }
func (kc KeyConfig) R() bool { return JoystickState(kc.Joy, kc.dR) }
func (kc KeyConfig) a() bool { return JoystickState(kc.Joy, kc.kA) }
func (kc KeyConfig) b() bool { return JoystickState(kc.Joy, kc.kB) }
func (kc KeyConfig) c() bool { return JoystickState(kc.Joy, kc.kC) }
func (kc KeyConfig) x() bool { return JoystickState(kc.Joy, kc.kX) }
func (kc KeyConfig) y() bool { return JoystickState(kc.Joy, kc.kY) }
func (kc KeyConfig) z() bool { return JoystickState(kc.Joy, kc.kZ) }
func (kc KeyConfig) s() bool { return JoystickState(kc.Joy, kc.kS) }
func (kc KeyConfig) d() bool { return JoystickState(kc.Joy, kc.kD) }
func (kc KeyConfig) w() bool { return JoystickState(kc.Joy, kc.kW) }
func (kc KeyConfig) m() bool { return JoystickState(kc.Joy, kc.kM) }

type InputBits int32

const (
	IB_PU InputBits = 1 << iota
	IB_PD
	IB_PL
	IB_PR
	IB_A
	IB_B
	IB_C
	IB_X
	IB_Y
	IB_Z
	IB_S
	IB_D
	IB_W
	IB_M
	IB_anybutton = IB_A | IB_B | IB_C | IB_X | IB_Y | IB_Z | IB_S | IB_D | IB_W | IB_M
)

func (ib *InputBits) SetInput(in int) {
	if 0 <= in && in < len(sys.keyConfig) {
		*ib = InputBits(Btoi(sys.keyConfig[in].U() || sys.joystickConfig[in].U()) |
			Btoi(sys.keyConfig[in].D() || sys.joystickConfig[in].D())<<1 |
			Btoi(sys.keyConfig[in].L() || sys.joystickConfig[in].L())<<2 |
			Btoi(sys.keyConfig[in].R() || sys.joystickConfig[in].R())<<3 |
			Btoi(sys.keyConfig[in].a() || sys.joystickConfig[in].a())<<4 |
			Btoi(sys.keyConfig[in].b() || sys.joystickConfig[in].b())<<5 |
			Btoi(sys.keyConfig[in].c() || sys.joystickConfig[in].c())<<6 |
			Btoi(sys.keyConfig[in].x() || sys.joystickConfig[in].x())<<7 |
			Btoi(sys.keyConfig[in].y() || sys.joystickConfig[in].y())<<8 |
			Btoi(sys.keyConfig[in].z() || sys.joystickConfig[in].z())<<9 |
			Btoi(sys.keyConfig[in].s() || sys.joystickConfig[in].s())<<10 |
			Btoi(sys.keyConfig[in].d() || sys.joystickConfig[in].d())<<11 |
			Btoi(sys.keyConfig[in].w() || sys.joystickConfig[in].w())<<12 |
			Btoi(sys.keyConfig[in].m() || sys.joystickConfig[in].m())<<13)
	}
}
func (ib InputBits) GetInput(cb *CommandBuffer, facing int32) {
	var B, F bool
	if facing < 0 {
		B, F = ib&IB_PR != 0, ib&IB_PL != 0
	} else {
		B, F = ib&IB_PL != 0, ib&IB_PR != 0
	}
	cb.Input(B, ib&IB_PD != 0, F, ib&IB_PU != 0, ib&IB_A != 0, ib&IB_B != 0,
		ib&IB_C != 0, ib&IB_X != 0, ib&IB_Y != 0, ib&IB_Z != 0, ib&IB_S != 0,
		ib&IB_D != 0, ib&IB_W != 0, ib&IB_M != 0)
}

type CommandKeyRemap struct {
	a, b, c, x, y, z, s, d, w, m, na, nb, nc, nx, ny, nz, ns, nd, nw, nm CommandKey
}

func NewCommandKeyRemap() *CommandKeyRemap {
	return &CommandKeyRemap{CK_a, CK_b, CK_c, CK_x, CK_y, CK_z, CK_s, CK_d, CK_w, CK_m,
		CK_na, CK_nb, CK_nc, CK_nx, CK_ny, CK_nz, CK_ns, CK_nd, CK_nw, CK_nm}
}

type CommandBuffer struct {
	Bb, Db, Fb, Ub                         int32
	ab, bb, cb, xb, yb, zb, sb, db, wb, mb int32
	B, D, F, U                             int8
	a, b, c, x, y, z, s, d, w, m           int8
}

func NewCommandBuffer() (c *CommandBuffer) {
	c = &CommandBuffer{}
	c.Reset()
	return
}
func (__ *CommandBuffer) Reset() {
	*__ = CommandBuffer{B: -1, D: -1, F: -1, U: -1,
		a: -1, b: -1, c: -1, x: -1, y: -1, z: -1, s: -1, d: -1, w: -1, m: -1}
}
func (__ *CommandBuffer) Input(B, D, F, U, a, b, c, x, y, z, s, d, w, m bool) {
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
	if d != (__.d > 0) {
		__.db = 0
		__.d *= -1
	}
	__.db += int32(__.d)
	if w != (__.w > 0) {
		__.wb = 0
		__.w *= -1
	}
	__.wb += int32(__.w)
	if m != (__.m > 0) {
		__.mb = 0
		__.m *= -1
	}
	__.mb += int32(__.m)
}
func (__ *CommandBuffer) InputBits(ib InputBits, f int32) {
	var B, F bool
	if f < 0 {
		B, F = ib&IB_PR != 0, ib&IB_PL != 0
	} else {
		B, F = ib&IB_PL != 0, ib&IB_PR != 0
	}
	__.Input(B, ib&IB_PD != 0, F, ib&IB_PU != 0, ib&IB_A != 0, ib&IB_B != 0,
		ib&IB_C != 0, ib&IB_X != 0, ib&IB_Y != 0, ib&IB_Z != 0, ib&IB_S != 0,
		ib&IB_D != 0, ib&IB_W != 0, ib&IB_M != 0)
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
	case CK_d:
		return __.db
	case CK_w:
		return __.wb
	case CK_m:
		return __.mb
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
	case CK_nd:
		return -__.db
	case CK_nw:
		return -__.wb
	case CK_nm:
		return -__.mb
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
		Abs(__.xb), Abs(__.yb), Abs(__.zb), Abs(__.sb), Abs(__.db), Abs(__.wb),
		Abs(__.mb))
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

func NewNetInput() *NetInput {
	ni := &NetInput{st: NS_Stop,
		sendEnd: make(chan bool, 1), recvEnd: make(chan bool, 1)}
	ni.sendEnd <- true
	ni.recvEnd <- true
	return ni
}
func (ni *NetInput) Close() {
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
		return Error("接続がありません。" + "\n" + "Error: Can not connect to the other player.")
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
	if ni.rep != nil {
		binary.Write(ni.rep, binary.LittleEndian, &seed)
	}
	if err := ni.writeI32(ni.time); err != nil {
		return err
	}
	if tmp, err := ni.readI32(); err != nil {
		return err
	} else if tmp != ni.time {
		return Error("同期エラーです。" + "\n" + "Synchronization error.")
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
				if ni.rep != nil {
					for _, nb := range ni.buf {
						binary.Write(ni.rep, binary.LittleEndian, &nb.buf[ni.time&31])
					}
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
	ib [MaxSimul*2 + MaxAttachedChar]InputBits
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
	dir, dirt, at, bt, ct, xt, yt, zt, st, dt, wt, mt int32
}

func (ai *AiInput) Update(level float32) {
	if sys.intro != 0 {
		ai.dirt, ai.at, ai.bt, ai.ct = 0, 0, 0, 0
		ai.xt, ai.yt, ai.zt, ai.st = 0, 0, 0, 0
		ai.dt, ai.wt, ai.mt = 0, 0, 0
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
	if dec(&ai.dirt) {
		ai.dir = Rand(0, 7)
	}
	osu, hanasu = int32((-11.25*level+165)*7), 30
	dec(&ai.at)
	dec(&ai.bt)
	dec(&ai.ct)
	dec(&ai.xt)
	dec(&ai.yt)
	dec(&ai.zt)
	dec(&ai.dt)
	dec(&ai.wt)
	osu = 3600
	dec(&ai.st)
	//dec(&ai.mt)
}
func (ai *AiInput) L() bool {
	return ai.dirt != 0 && (ai.dir == 5 || ai.dir == 6 || ai.dir == 7)
}
func (ai *AiInput) R() bool {
	return ai.dirt != 0 && (ai.dir == 1 || ai.dir == 2 || ai.dir == 3)
}
func (ai *AiInput) U() bool {
	return ai.dirt != 0 && (ai.dir == 7 || ai.dir == 0 || ai.dir == 1)
}
func (ai *AiInput) D() bool {
	return ai.dirt != 0 && (ai.dir == 3 || ai.dir == 4 || ai.dir == 5)
}
func (ai *AiInput) a() bool {
	return ai.at != 0
}
func (ai *AiInput) b() bool {
	return ai.bt != 0
}
func (ai *AiInput) c() bool {
	return ai.ct != 0
}
func (ai *AiInput) x() bool {
	return ai.xt != 0
}
func (ai *AiInput) y() bool {
	return ai.yt != 0
}
func (ai *AiInput) z() bool {
	return ai.zt != 0
}
func (ai *AiInput) s() bool {
	return ai.st != 0
}
func (ai *AiInput) d() bool {
	return ai.dt != 0
}
func (ai *AiInput) w() bool {
	return ai.wt != 0
}
func (ai *AiInput) m() bool {
	return ai.mt != 0
}

type cmdElem struct {
	key                       []CommandKey
	tametime                  int32
	slash, greater, direction bool
}

func (ce *cmdElem) IsDirection() bool {
	//ここで~は方向コマンドとして返さない
	return !ce.slash && len(ce.key) == 1 && ce.key[0] < CK_nBs && (ce.key[0] < CK_nB || ce.key[0] > CK_nUF)
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
	cmdi, tamei         int
	time, cur           int32
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
			case 'd':
				if tilde {
					ce.key = append(ce.key, kr.nd)
				} else {
					ce.key = append(ce.key, kr.d)
				}
				tilde = false
			case 'w':
				if tilde {
					ce.key = append(ce.key, kr.nw)
				} else {
					ce.key = append(ce.key, kr.w)
				}
				tilde = false
			case 'm':
				if tilde {
					ce.key = append(ce.key, kr.nm)
				} else {
					ce.key = append(ce.key, kr.m)
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
	c.cmdi, c.tamei, c.cur, c.curbuftime = 0, -1, 0, 0
	for i := range c.held {
		c.held[i] = false
	}
}
func (c *Command) bufTest(cbuf *CommandBuffer, ai bool,
	holdTemp *[CK_Last + 1]bool) bool {
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
					if ks == 1 && (c.cmdi > 0 || len(c.hold) > 1) && !c.held[i] &&
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
		if c.cmdi == len(c.cmd)-1 && (!allHold || notHeld > 1) {
			return anyHeld || c.cmdi > 0
		}
	}
	if !ai && c.cmd[c.cmdi].slash {
		if c.cmdi > 0 {
			if notHeld == 1 {
				if len(c.cmd[c.cmdi-1].key) != 1 {
					return false
				}
				if CK_a <= c.cmd[c.cmdi-1].key[0] && c.cmd[c.cmdi-1].key[0] <= CK_s {
					ks := cbuf.State(c.cmd[c.cmdi-1].key[0])
					if 0 < ks && ks <= cbuf.LastDirectionTime() {
						return true
					}
				}
			} else if len(c.cmd[c.cmdi-1].key) > 1 {
				for _, k := range c.cmd[c.cmdi-1].key {
					if CK_a <= k && k <= CK_s && cbuf.State(k) > 0 {
						return false
					}
				}
			}
		}
		c.cmdi++
		return true
	}
	fail := func() bool {
		if c.cmdi == 0 {
			return anyHeld
		}
		if !ai && (c.cmd[c.cmdi].greater || c.cmd[c.cmdi].direction) {
			var t int32
			if c.cmd[c.cmdi].greater {
				t = cbuf.LastChangeTime()
			} else {
				t = cbuf.LastDirectionTime()
			}
			for _, k := range c.cmd[c.cmdi-1].key {
				if Abs(cbuf.State2(k)) == t {
					return true
				}
			}
			c.Clear()
			return c.bufTest(cbuf, ai, holdTemp)
		}
		return true
	}
	if c.tamei != c.cmdi {
		if c.cmd[c.cmdi].tametime > 1 {
			for _, k := range c.cmd[c.cmdi].key {
				ks := cbuf.State(k)
				if ks > 0 {
					return ai
				}
				if func() bool {
					if ai {
						return Rand(0, c.cmd[c.cmdi].tametime) != 0
					}
					return -ks < c.cmd[c.cmdi].tametime
				}() {
					return anyHeld || c.cmdi > 0
				}
			}
			c.tamei = c.cmdi
		} else if c.cmdi > 0 && len(c.cmd[c.cmdi-1].key) == 1 &&
			len(c.cmd[c.cmdi].key) == 1 && c.cmd[c.cmdi-1].key[0] < CK_Bs &&
			c.cmd[c.cmdi].key[0] < CK_nB && (c.cmd[c.cmdi-1].key[0]-
			c.cmd[c.cmdi].key[0])&7 == 0 {
			if cbuf.B < 0 && cbuf.D < 0 && cbuf.F < 0 && cbuf.U < 0 {
				c.tamei = c.cmdi
			} else {
				return fail()
			}
		}
	}
	foo := false
	for _, k := range c.cmd[c.cmdi].key {
		n := cbuf.State2(k)
		if c.cmd[c.cmdi].slash {
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
	c.cmdi++
	if c.cmdi < len(c.cmd) && c.cmd[c.cmdi-1].IsDToB(c.cmd[c.cmdi]) {
		return c.bufTest(cbuf, ai, holdTemp)
	}
	return true
}
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
	var holdTemp *[CK_Last + 1]bool
	if cbuf == nil || !c.bufTest(cbuf, ai, holdTemp) {
		foo := c.tamei == 0 && c.cmdi == 0
		c.Clear()
		if foo {
			c.tamei = 0
		}
		return
	}
	if c.cmdi == 1 && c.cmd[0].slash {
		c.cur = 0
	} else {
		c.cur++
	}
	complete := c.cmdi == len(c.cmd)
	if !complete && (ai || c.cur <= c.time) {
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
		var L, R, U, D, a, b, c, x, y, z, s, d, w, m bool
		if i < 0 {
			i = ^i
			if i < len(sys.aiInput) {
				L = sys.aiInput[i].L()
				R = sys.aiInput[i].R()
				U = sys.aiInput[i].U()
				D = sys.aiInput[i].D()
				a = sys.aiInput[i].a()
				b = sys.aiInput[i].b()
				c = sys.aiInput[i].c()
				x = sys.aiInput[i].x()
				y = sys.aiInput[i].y()
				z = sys.aiInput[i].z()
				s = sys.aiInput[i].s()
				d = sys.aiInput[i].d()
				w = sys.aiInput[i].w()
				m = sys.aiInput[i].m()
			}
		} else if i < len(sys.inputRemap) {
			in := sys.inputRemap[i]
			if in < len(sys.keyConfig) {
				joy := sys.keyConfig[in].Joy
				if joy == -1 {
					L = sys.keyConfig[in].L()
					R = sys.keyConfig[in].R()
					U = sys.keyConfig[in].U()
					D = sys.keyConfig[in].D()
					a = sys.keyConfig[in].a()
					b = sys.keyConfig[in].b()
					c = sys.keyConfig[in].c()
					x = sys.keyConfig[in].x()
					y = sys.keyConfig[in].y()
					z = sys.keyConfig[in].z()
					s = sys.keyConfig[in].s()
					d = sys.keyConfig[in].d()
					w = sys.keyConfig[in].w()
					m = sys.keyConfig[in].m()
				}
			}
			if in < len(sys.joystickConfig) {
				joyS := sys.joystickConfig[in].Joy
				if joyS >= 0 {
					if !L {
						L = sys.joystickConfig[in].L()
					}
					if !R {
						R = sys.joystickConfig[in].R()
					}
					if !U {
						U = sys.joystickConfig[in].U()
					}
					if !D {
						D = sys.joystickConfig[in].D()
					}
					if !a {
						a = sys.joystickConfig[in].a()
					}
					if !b {
						b = sys.joystickConfig[in].b()
					}
					if !c {
						c = sys.joystickConfig[in].c()
					}
					if !x {
						x = sys.joystickConfig[in].x()
					}
					if !y {
						y = sys.joystickConfig[in].y()
					}
					if !z {
						z = sys.joystickConfig[in].z()
					}
					if !s {
						s = sys.joystickConfig[in].s()
					}
					if !d {
						d = sys.joystickConfig[in].d()
					}
					if !w {
						w = sys.joystickConfig[in].w()
					}
					if !m {
						m = sys.joystickConfig[in].m()
					}
				}
			}
		}
		var B, F bool
		if facing < 0 {
			B, F = R, L
		} else {
			B, F = L, R
		}
		cl.Buffer.Input(B, D, F, U, a, b, c, x, y, z, s, d, w, m)
	}
	return step
}
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
