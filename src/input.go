package main

import (
	"encoding/binary"
	"net"
	"os"
	"strings"
	"time"
)

var ModAlt = NewModifierKey(false, true, false)
var ModCtrlAlt = NewModifierKey(true, true, false)
var ModCtrlAltShift = NewModifierKey(true, true, true)

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
	CK_rB // r stands for release (~)
	CK_rD
	CK_rF
	CK_rU
	CK_rDB
	CK_rUB
	CK_rDF
	CK_rUF
	CK_Bs // s stands for sign ($)
	CK_Ds
	CK_Fs
	CK_Us
	CK_DBs
	CK_UBs
	CK_DFs
	CK_UFs
	CK_rBs
	CK_rDs
	CK_rFs
	CK_rUs
	CK_rDBs
	CK_rUBs
	CK_rDFs
	CK_rUFs
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
	CK_ra
	CK_rb
	CK_rc
	CK_rx
	CK_ry
	CK_rz
	CK_rs
	CK_rd
	CK_rw
	CK_rm
	CK_Last = CK_rm
)

type NetState int

const (
	NS_Stop NetState = iota
	NS_Playing
	NS_End
	NS_Stopped
	NS_Error
)

type ShortcutScript struct {
	Activate bool
	Script   string
	Pause    bool
	DebugKey bool
}

type ShortcutKey struct {
	Key Key
	Mod ModifierKey
}

func NewShortcutKey(key Key, ctrl, alt, shift bool) *ShortcutKey {
	sk := &ShortcutKey{}
	sk.Key = key
	sk.Mod = NewModifierKey(ctrl, alt, shift)
	return sk
}

func (sk ShortcutKey) Test(k Key, m ModifierKey) bool {
	return k == sk.Key && (m&ModCtrlAltShift) == sk.Mod
}

func OnKeyReleased(key Key, mk ModifierKey) {
	if key != KeyUnknown {
		sys.keyState[key] = false
		sys.keyInput = KeyUnknown
		sys.keyString = ""
	}
}

func OnKeyPressed(key Key, mk ModifierKey) {
	if key != KeyUnknown {
		sys.keyState[key] = true
		sys.keyInput = key
		sys.esc = sys.esc ||
			key == KeyEscape && (mk&ModCtrlAlt) == 0
		for k, v := range sys.shortcutScripts {
			if sys.netInput == nil && (sys.fileInput == nil || !v.DebugKey) &&
				//(!sys.paused || sys.step || v.Pause) &&
				(sys.allowDebugKeys || !v.DebugKey) {
				v.Activate = v.Activate || k.Test(key, mk)
			}
		}
		if key == KeyF12 {
			captureScreen()
		}
		if key == KeyEnter && (mk&ModAlt) != 0 {
			sys.window.toggleFullscreen()
		}
	}
}

func OnTextEntered(s string) {
	sys.keyString = s
}

func JoystickState(joy, button int) bool {
	if joy < 0 {
		return sys.keyState[Key(button)]
	}
	if joy >= input.GetMaxJoystickCount() {
		return false
	}
	if button >= 0 {
		// Query button state
		btns := input.GetJoystickButtons(joy)
		if button >= len(btns) {
			return false
		}
		return btns[button] != 0
	} else {
		// Query axis state
		axis := -button - 1
		axes := input.GetJoystickAxes(joy)
		if axis >= len(axes)*2 {
			return false
		}

		// Read value and invert sign for odd indices
		val := axes[axis/2] * float32((axis&1)*2-1)

		var joyName = input.GetJoystickName(joy)

		// Xbox360コントローラーのLRトリガー判定
		// "Evaluate LR triggers on the Xbox 360 controller"
		if (axis == 9 || axis == 11) && (strings.Contains(joyName, "XInput") || strings.Contains(joyName, "X360")) {
			return val > sys.xinputTriggerSensitivity
		}

		// Ignore trigger axis on PS4 (We already have buttons)
		if (axis >= 6 && axis <= 9) && joyName == "PS4 Controller" {
			return false
		}

		return val > sys.controllerStickSensitivity
	}
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

// Save local inputs as input bits to send or record
func (ib *InputBits) KeysToBits(U, D, L, R, a, b, c, x, y, z, s, d, w, m bool) {
	*ib = InputBits(Btoi(U) |
		Btoi(D)<<1 |
		Btoi(L)<<2 |
		Btoi(R)<<3 |
		Btoi(a)<<4 |
		Btoi(b)<<5 |
		Btoi(c)<<6 |
		Btoi(x)<<7 |
		Btoi(y)<<8 |
		Btoi(z)<<9 |
		Btoi(s)<<10 |
		Btoi(d)<<11 |
		Btoi(w)<<12 |
		Btoi(m)<<13)
}

// Convert received input bits back into keys
func (ib InputBits) BitsToKeys(cb *CommandBuffer, facing int32) {
	var U, D, B, F, a, b, c, x, y, z, s, d, w, m bool
	// Convert bits to logical symbols
	U = ib&IB_PU != 0
	D = ib&IB_PD != 0
	if facing < 0 {
		B, F = ib&IB_PR != 0, ib&IB_PL != 0
	} else {
		B, F = ib&IB_PL != 0, ib&IB_PR != 0
	}
	a = ib&IB_A != 0
	b = ib&IB_B != 0
	c = ib&IB_C != 0
	x = ib&IB_X != 0
	y = ib&IB_Y != 0
	z = ib&IB_Z != 0
	s = ib&IB_S != 0
	d = ib&IB_D != 0
	w = ib&IB_W != 0
	m = ib&IB_M != 0
	// Absolute priority SOCD resolution is enforced during netplay
	// TODO: Port the other options as well
	if U && D {
		D = false
	}
	if B && F {
		B = false
	}
	cb.Input(B, D, F, U, a, b, c, x, y, z, s, d, w, m)
}

type CommandKeyRemap struct {
	a, b, c, x, y, z, s, d, w, m, na, nb, nc, nx, ny, nz, ns, nd, nw, nm CommandKey
}

func NewCommandKeyRemap() *CommandKeyRemap {
	return &CommandKeyRemap{CK_a, CK_b, CK_c, CK_x, CK_y, CK_z, CK_s, CK_d, CK_w, CK_m,
		CK_ra, CK_rb, CK_rc, CK_rx, CK_ry, CK_rz, CK_rs, CK_rd, CK_rw, CK_rm}
}

type InputReader struct {
	SocdAllow          [4]bool
	SocdFirst          [4]bool
	ButtonAssist       bool
	ButtonAssistBuffer [9]bool
}

func NewInputReader() *InputReader {
	return &InputReader{
		SocdAllow:          [4]bool{},
		SocdFirst:          [4]bool{},
		ButtonAssist:       false,
		ButtonAssistBuffer: [9]bool{},
	}
}

// Reads controllers and converts inputs to letters for later processing
func (ir *InputReader) LocalInput(in int) (bool, bool, bool, bool, bool, bool, bool, bool, bool, bool, bool, bool, bool, bool) {
	var U, D, L, R, a, b, c, x, y, z, s, d, w, m bool
	// Keyboard
	if in < len(sys.keyConfig) {
		joy := sys.keyConfig[in].Joy
		if joy == -1 {
			U = sys.keyConfig[in].U()
			D = sys.keyConfig[in].D()
			L = sys.keyConfig[in].L()
			R = sys.keyConfig[in].R()
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
	// Joystick
	if in < len(sys.joystickConfig) {
		joyS := sys.joystickConfig[in].Joy
		if joyS >= 0 {
			U = sys.joystickConfig[in].U() || U // Does not override keyboard
			D = sys.joystickConfig[in].D() || D
			L = sys.joystickConfig[in].L() || L
			R = sys.joystickConfig[in].R() || R
			a = sys.joystickConfig[in].a() || a
			b = sys.joystickConfig[in].b() || b
			c = sys.joystickConfig[in].c() || c
			x = sys.joystickConfig[in].x() || x
			y = sys.joystickConfig[in].y() || y
			z = sys.joystickConfig[in].z() || z
			s = sys.joystickConfig[in].s() || s
			d = sys.joystickConfig[in].d() || d
			w = sys.joystickConfig[in].w() || w
			m = sys.joystickConfig[in].m() || m
		}
	}
	// Button assist is checked locally so the sent inputs are already processed
	if sys.inputButtonAssist {
		a, b, c, x, y, z, s, d, w = ir.ButtonAssistCheck(a, b, c, x, y, z, s, d, w)
	}
	return U, D, L, R, a, b, c, x, y, z, s, d, w, m
}

// Resolve Simultaneous Opposing Cardinal Directions
func (ir *InputReader) SocdResolution(U, D, B, F bool) (bool, bool, bool, bool) {
	// Absolute priority SOCD resolution is enforced during netplay
	if sys.netInput != nil || sys.fileInput != nil {
		if U && D {
			D = false
		}
		if B && F {
			B = false
		}
	} else {
		// Check first direction held between U and D
		if U || D {
			if !U {
				ir.SocdFirst[0] = false
			}
			if !D {
				ir.SocdFirst[1] = false
			}
			if !ir.SocdFirst[0] && !ir.SocdFirst[1] {
				if D {
					ir.SocdFirst[1] = true
				} else {
					ir.SocdFirst[0] = true
				}
			}
		} else {
			ir.SocdFirst[0] = false
			ir.SocdFirst[1] = false
		}
		// Check first direction held between U and D
		if B || F {
			if !B {
				ir.SocdFirst[2] = false
			}
			if !F {
				ir.SocdFirst[3] = false
			}
			if !ir.SocdFirst[2] && !ir.SocdFirst[3] {
				if B {
					ir.SocdFirst[2] = true
				} else {
					ir.SocdFirst[3] = true
				}
			}
		} else {
			ir.SocdFirst[2] = false
			ir.SocdFirst[3] = false
		}
		// SOCD for back and forward
		if B && F {
			switch sys.inputSOCDresolution {
			// Type 0 - Allow both directions (no resolution)
			case 0:
				ir.SocdAllow[2] = true
				ir.SocdAllow[3] = true
			// Type 1 - Last direction priority
			case 1:
				// if F was held before B, disable F
				if ir.SocdFirst[3] {
					ir.SocdAllow[2] = true
					ir.SocdAllow[3] = false
				} else {
					// else disable B
					ir.SocdAllow[2] = false
					ir.SocdAllow[3] = true
				}
			// Type 2 - Absolute priority (offense over defense)
			case 2:
				ir.SocdAllow[2] = false
				ir.SocdAllow[3] = true
			// Type 3 - First direction priority
			case 3:
				// if F was held before B, disable B
				if ir.SocdFirst[3] {
					ir.SocdAllow[2] = false
					ir.SocdAllow[3] = true
				} else {
					// else disable F
					ir.SocdAllow[2] = true
					ir.SocdAllow[3] = false
				}
			// Type 4 - Deny either direction (neutral)
			default:
				ir.SocdAllow[2] = false
				ir.SocdAllow[3] = false
			}
		} else {
			ir.SocdAllow[2] = true
			ir.SocdAllow[3] = true
		}
		// SOCD for down and up
		if D && U {
			switch sys.inputSOCDresolution {
			// Type 0 - Allow both directions (no resolution)
			case 0:
				ir.SocdAllow[0] = true
				ir.SocdAllow[1] = true
			// Type 1 - Last direction priority
			case 1:
				// if U was held before D, disable U
				if ir.SocdFirst[0] {
					ir.SocdAllow[0] = false
					ir.SocdAllow[1] = true
				} else {
					// else disable D
					ir.SocdAllow[0] = true
					ir.SocdAllow[1] = false
				}
			// Type 2 - Absolute priority (offense over defense)
			case 2:
				ir.SocdAllow[0] = true
				ir.SocdAllow[1] = false
			// Type 3 - First direction priority
			case 3:
				// if U was held before D, disable D
				if ir.SocdFirst[0] {
					ir.SocdAllow[0] = true
					ir.SocdAllow[1] = false
				} else {
					// else disable U
					ir.SocdAllow[0] = false
					ir.SocdAllow[1] = true
				}
			// Type 4 - Deny either direction (neutral)
			default:
				ir.SocdAllow[0] = false
				ir.SocdAllow[1] = false
			}
		} else {
			ir.SocdAllow[1] = true
			ir.SocdAllow[0] = true
		}
		// Apply rules
		U = U && ir.SocdAllow[0]
		D = D && ir.SocdAllow[1]
		B = B && ir.SocdAllow[2]
		F = F && ir.SocdAllow[3]
	}
	return U, D, B, F
}

// Add extra frame of leniency when checking button presses
func (ir *InputReader) ButtonAssistCheck(a, b, c, x, y, z, s, d, w bool) (bool, bool, bool, bool, bool, bool, bool, bool, bool) {
	// Set buttons to buffered state
	a = ir.ButtonAssistBuffer[0] || a
	b = ir.ButtonAssistBuffer[1] || b
	c = ir.ButtonAssistBuffer[2] || c
	x = ir.ButtonAssistBuffer[3] || x
	y = ir.ButtonAssistBuffer[4] || y
	z = ir.ButtonAssistBuffer[5] || z
	s = ir.ButtonAssistBuffer[6] || s
	d = ir.ButtonAssistBuffer[7] || d
	w = ir.ButtonAssistBuffer[8] || w
	ir.ButtonAssistBuffer = [9]bool{}
	// Reenable assist when no buttons are being held
	if !a && !b && !c && !x && !y && !z && !s && !d && !w {
		ir.ButtonAssist = true
	}
	// Disable then buffer buttons if assist is enabled
	if ir.ButtonAssist == true {
		if a || b || c || x || y || z || s || d || w {
			ir.ButtonAssist = false
			ir.ButtonAssistBuffer = [9]bool{a, b, c, x, y, z, s, d, w}
			a, b, c, x, y, z, s, d, w = false, false, false, false, false, false, false, false, false
		}
	}
	return a, b, c, x, y, z, s, d, w
}

type CommandBuffer struct {
	Bb, Db, Fb, Ub                         int32
	ab, bb, cb, xb, yb, zb, sb, db, wb, mb int32
	B, D, F, U                             int8
	a, b, c, x, y, z, s, d, w, m           int8
	InputReader                            *InputReader
}

func NewCommandBuffer() (c *CommandBuffer) {
	ir := NewInputReader()
	c = &CommandBuffer{InputReader: ir}
	c.Reset()
	return c
}

func (c *CommandBuffer) Reset() {
	*c = CommandBuffer{
		B: -1, D: -1, F: -1, U: -1,
		a: -1, b: -1, c: -1, x: -1, y: -1, z: -1, s: -1, d: -1, w: -1, m: -1,
		InputReader: NewInputReader(),
	}
}

// Update command buffer according to received inputs
func (__ *CommandBuffer) Input(B, D, F, U, a, b, c, x, y, z, s, d, w, m bool) {
	// SOCD resolution is now handled beforehand, so that it may be easier to port to netplay later
	if B != (__.B > 0) {
		__.Bb = 0
		__.B *= -1
	}
	__.Bb += int32(__.B)
	if D != (__.D > 0) {
		__.Db = 0
		__.D *= -1
	}
	__.Db += int32(__.D)
	if F != (__.F > 0) {
		__.Fb = 0
		__.F *= -1
	}
	__.Fb += int32(__.F)
	if U != (__.U > 0) {
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

// Check buffer state of each key
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
	case CK_rB:
		return -Min(-Max(__.Db, __.Ub), __.Bb)
	case CK_rD:
		return -Min(-Max(__.Bb, __.Fb), __.Db)
	case CK_rF:
		return -Min(-Max(__.Db, __.Ub), __.Fb)
	case CK_rU:
		return -Min(-Max(__.Bb, __.Fb), __.Ub)
	case CK_rDB:
		return -Min(__.Db, __.Bb)
	case CK_rUB:
		return -Min(__.Ub, __.Bb)
	case CK_rDF:
		return -Min(__.Db, __.Fb)
	case CK_rUF:
		return -Min(__.Ub, __.Fb)
	case CK_rBs:
		return -__.Bb
	case CK_rDs:
		return -__.Db
	case CK_rFs:
		return -__.Fb
	case CK_rUs:
		return -__.Ub
	case CK_rDBs:
		return -Min(__.Db, __.Bb)
	case CK_rUBs:
		return -Min(__.Ub, __.Bb)
	case CK_rDFs:
		return -Min(__.Db, __.Fb)
	case CK_rUFs:
		return -Min(__.Ub, __.Fb)
	case CK_ra:
		return -__.ab
	case CK_rb:
		return -__.bb
	case CK_rc:
		return -__.cb
	case CK_rx:
		return -__.xb
	case CK_ry:
		return -__.yb
	case CK_rz:
		return -__.zb
	case CK_rs:
		return -__.sb
	case CK_rd:
		return -__.db
	case CK_rw:
		return -__.wb
	case CK_rm:
		return -__.mb
	}
	return 0
}

// Check buffer state of each key
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
	// "In MUGEN, adding '$' to diagonal inputs doesn't have any meaning."
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
	case CK_rBs:
		return f(__.State(CK_B), __.State(CK_UB), __.State(CK_DB))
	case CK_rDs:
		return f(__.State(CK_D), __.State(CK_DB), __.State(CK_DF))
	case CK_rFs:
		return f(__.State(CK_F), __.State(CK_DF), __.State(CK_UF))
	case CK_rUs:
		return f(__.State(CK_U), __.State(CK_UB), __.State(CK_UF))
		//case CK_rDBs:
		//	return f(__.State(CK_DB), __.State(CK_D), __.State(CK_B))
		//case CK_rUBs:
		//	return f(__.State(CK_UB), __.State(CK_U), __.State(CK_B))
		//case CK_rDFs:
		//	return f(__.State(CK_DF), __.State(CK_D), __.State(CK_F))
		//case CK_rUFs:
		//	return f(__.State(CK_UF), __.State(CK_U), __.State(CK_F))
	}
	return __.State(ck)
}

// Time since last directional input was received
func (__ *CommandBuffer) LastDirectionTime() int32 {
	return Min(Abs(__.Bb), Abs(__.Db), Abs(__.Fb), Abs(__.Ub))
}

// Time since last input was received. Used for ">" type commands
func (__ *CommandBuffer) LastChangeTime() int32 {
	return Min(__.LastDirectionTime(), Abs(__.ab), Abs(__.bb), Abs(__.cb),
		Abs(__.xb), Abs(__.yb), Abs(__.zb), Abs(__.sb), Abs(__.db), Abs(__.wb),
		Abs(__.mb))
}

type NetBuffer struct {
	buf              [32]InputBits
	curT, inpT, senT int32
	InputReader      *InputReader
}

func (nb *NetBuffer) reset(time int32) {
	nb.curT, nb.inpT, nb.senT = time, time, time
	nb.InputReader = NewInputReader()
}

// Check local inputs
func (nb *NetBuffer) localUpdate(in int) {
	if nb.inpT-nb.curT < 32 {
		nb.buf[nb.inpT&31].KeysToBits(nb.InputReader.LocalInput(in))
		nb.inpT++
	}
}

// Convert bits to keys
func (nb *NetBuffer) input(cb *CommandBuffer, facing int32) {
	if nb.curT < nb.inpT {
		nb.buf[nb.curT&31].BitsToKeys(cb, facing)
	}
}

type NetInput struct {
	ln           *net.TCPListener
	conn         *net.TCPConn
	st           NetState
	sendEnd      chan bool
	recvEnd      chan bool
	buf          [MaxSimul*2 + MaxAttachedChar]NetBuffer
	locIn        int
	remIn        int
	time         int32
	stoppedcnt   int32
	delay        int32
	rep          *os.File
	host         bool
	preFightTime int32
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
		return Error("Can not connect to the other player")
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
	var pfTime int32
	if ni.host {
		pfTime = sys.preFightTime
		if err := ni.writeI32(pfTime); err != nil {
			return err
		}
	} else {
		var err error
		if pfTime, err = ni.readI32(); err != nil {
			return err
		}
	}
	ni.preFightTime = pfTime
	if ni.rep != nil {
		binary.Write(ni.rep, binary.LittleEndian, &seed)
		binary.Write(ni.rep, binary.LittleEndian, &pfTime)
	}
	if err := ni.writeI32(ni.time); err != nil {
		return err
	}
	if tmp, err := ni.readI32(); err != nil {
		return err
	} else if tmp != ni.time {
		return Error("Synchronization error")
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
	f      *os.File
	ib     [MaxSimul*2 + MaxAttachedChar]InputBits
	pfTime int32
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

// Convert bits to keys
func (fi *FileInput) Input(cb *CommandBuffer, i int, facing int32) {
	if i >= 0 && i < len(fi.ib) {
		fi.ib[sys.inputRemap[i]].BitsToKeys(cb, facing)
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
		}
		var pfTime int32
		if binary.Read(fi.f, binary.LittleEndian, &pfTime) == nil {
			fi.pfTime = pfTime
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
	// Disable AI button jamming
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
	// Pick a random direction to press
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
	//dec(&ai.mt) // We don't need the AI to jam the menu button
}

// 0 = U, 1 = UR, 2 = R, 3 = DR, 4 = D, 5 = DL, 6 = L, 7 = UL
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

// cmdElem refers to each of the inputs required to complete a command
type cmdElem struct {
	key        []CommandKey
	chargetime int32
	slash      bool
	greater    bool
}

func (ce *cmdElem) IsDirection() bool {
	//ここで~は方向コマンドとして返さない
	// "At this point, '~' is not returned as a directional command." (?)
	return !ce.slash && len(ce.key) == 1 && ce.key[0] < CK_rBs && (ce.key[0] < CK_rB || ce.key[0] > CK_rUF)
}

// Check if two command elements can be checked in the same frame
func (ce *cmdElem) IsDirToButton(next cmdElem) bool {
	if next.slash {
		return false
	}
	// This logic seems more complex in Mugen because of variable input delay
	// Not if first element includes button press or release
	for _, k := range ce.key {
		if k >= CK_a {
			return false
		}
	}
	// Not if both elements share keys
	for _, k := range ce.key {
		for _, n := range next.key {
			if k == n {
				return false
			}
		}
	}
	// Yes if second element includes a button press
	for range ce.key {
		for _, n := range next.key {
			if n >= CK_a && n < CK_ra {
				return true
			}
		}
	}
	// Yes if release direction then not release direction (includes buttons)
	for _, k := range ce.key {
		if k >= CK_rB && k <= CK_rUF || k >= CK_rBs && k <= CK_rUFs {
			for _, n := range next.key {
				if (n < CK_rB || n > CK_rUF) && (n < CK_rBs || n > CK_rUFs) {
					return true
				}
			}
		}
	}
	return false
}

// Command refers to each individual command from the CMD file
type Command struct {
	name                string
	hold                [][]CommandKey
	held                []bool
	cmd                 []cmdElem
	cmdi, chargei       int
	time, curtime       int32
	buftime, curbuftime int32
	completeflag        bool
}

func newCommand() *Command {
	return &Command{chargei: -1, time: 1, buftime: 1}
}

// This is used to first compile the commands
func ReadCommand(name, cmdstr string, kr *CommandKeyRemap) (*Command, error) {
	c := newCommand()
	c.name = name
	cmd := strings.Split(cmdstr, ",")
	for _, cestr := range cmd {
		if len(c.cmd) > 0 && c.cmd[len(c.cmd)-1].slash {
			c.hold = append(c.hold, c.cmd[len(c.cmd)-1].key)
			c.cmd[len(c.cmd)-1] = cmdElem{chargetime: 1}
		} else {
			c.cmd = append(c.cmd, cmdElem{chargetime: 1})
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
				ce.chargetime = n
			}
		case '/':
			ce.slash = true
			nextChar()
		}
		for len(cestr) > 0 {
			switch getChar() {
			case 'B':
				if tilde {
					ce.key = append(ce.key, CK_rB)
				} else {
					ce.key = append(ce.key, CK_B)
				}
				tilde = false
			case 'D':
				if len(cestr) > 1 && cestr[1] == 'B' {
					nextChar()
					if tilde {
						ce.key = append(ce.key, CK_rDB)
					} else {
						ce.key = append(ce.key, CK_DB)
					}
				} else if len(cestr) > 1 && cestr[1] == 'F' {
					nextChar()
					if tilde {
						ce.key = append(ce.key, CK_rDF)
					} else {
						ce.key = append(ce.key, CK_DF)
					}
				} else {
					if tilde {
						ce.key = append(ce.key, CK_rD)
					} else {
						ce.key = append(ce.key, CK_D)
					}
				}
				tilde = false
			case 'F':
				if tilde {
					ce.key = append(ce.key, CK_rF)
				} else {
					ce.key = append(ce.key, CK_F)
				}
				tilde = false
			case 'U':
				if len(cestr) > 1 && cestr[1] == 'B' {
					nextChar()
					if tilde {
						ce.key = append(ce.key, CK_rUB)
					} else {
						ce.key = append(ce.key, CK_UB)
					}
				} else if len(cestr) > 1 && cestr[1] == 'F' {
					nextChar()
					if tilde {
						ce.key = append(ce.key, CK_rUF)
					} else {
						ce.key = append(ce.key, CK_UF)
					}
				} else {
					if tilde {
						ce.key = append(ce.key, CK_rU)
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
						ce.key = append(ce.key, CK_rBs)
					} else {
						ce.key = append(ce.key, CK_Bs)
					}
					tilde = false
				case 'D':
					if len(cestr) > 1 && cestr[1] == 'B' {
						nextChar()
						if tilde {
							ce.key = append(ce.key, CK_rDBs)
						} else {
							ce.key = append(ce.key, CK_DBs)
						}
					} else if len(cestr) > 1 && cestr[1] == 'F' {
						nextChar()
						if tilde {
							ce.key = append(ce.key, CK_rDFs)
						} else {
							ce.key = append(ce.key, CK_DFs)
						}
					} else {
						if tilde {
							ce.key = append(ce.key, CK_rDs)
						} else {
							ce.key = append(ce.key, CK_Ds)
						}
					}
					tilde = false
				case 'F':
					if tilde {
						ce.key = append(ce.key, CK_rFs)
					} else {
						ce.key = append(ce.key, CK_Fs)
					}
					tilde = false
				case 'U':
					if len(cestr) > 1 && cestr[1] == 'B' {
						nextChar()
						if tilde {
							ce.key = append(ce.key, CK_rUBs)
						} else {
							ce.key = append(ce.key, CK_UBs)
						}
					} else if len(cestr) > 1 && cestr[1] == 'F' {
						nextChar()
						if tilde {
							ce.key = append(ce.key, CK_rUFs)
						} else {
							ce.key = append(ce.key, CK_UFs)
						}
					} else {
						if tilde {
							ce.key = append(ce.key, CK_rUs)
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
				// do nothing
			default:
				// error
			}
			nextChar()
		}
		// Two consecutive identical directions are considered ">"
		if len(c.cmd) >= 2 && ce.IsDirection() && c.cmd[len(c.cmd)-2].IsDirection() {
			if ce.key[0] == c.cmd[len(c.cmd)-2].key[0] {
				ce.greater = true
			}
		}
	}
	if c.cmd[len(c.cmd)-1].slash {
		c.hold = append(c.hold, c.cmd[len(c.cmd)-1].key)
	}
	c.held = make([]bool, len(c.hold))
	return c, nil
}

func (c *Command) Clear(buf bool) {
	c.cmdi = 0
	c.chargei = -1
	c.curtime = 0
	if !buf { // Otherwise keep buffer time. Mugen doesn't do this but it seems like the right thing to do
		c.curbuftime = 0
	}
	for i := range c.held {
		c.held[i] = false
	}
}

// Check if inputs match the command elements
func (c *Command) bufTest(cbuf *CommandBuffer, ai bool, holdTemp *[CK_Last + 1]bool) bool {
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
		// There's a bug here where for instance pressing DF does not invalidate F, F
		// Mugen does the same thing, however
		if !ai && c.cmd[c.cmdi].greater {
			for _, k := range c.cmd[c.cmdi-1].key {
				if Abs(cbuf.State2(k)) == cbuf.LastChangeTime() {
					return true
				}
			}
			c.Clear(false)
			return c.bufTest(cbuf, ai, holdTemp)
		}
		return true
	}
	if c.chargei != c.cmdi {
		if c.cmd[c.cmdi].chargetime > 1 {
			for _, k := range c.cmd[c.cmdi].key {
				ks := cbuf.State(k)
				if ks > 0 {
					return ai
				}
				if func() bool {
					if ai {
						return Rand(0, c.cmd[c.cmdi].chargetime) != 0
					}
					return -ks < c.cmd[c.cmdi].chargetime
				}() {
					return anyHeld || c.cmdi > 0
				}
			}
			c.chargei = c.cmdi
		} else if c.cmdi > 0 && len(c.cmd[c.cmdi-1].key) == 1 &&
			len(c.cmd[c.cmdi].key) == 1 && c.cmd[c.cmdi-1].key[0] < CK_Bs &&
			c.cmd[c.cmdi].key[0] < CK_rB && (c.cmd[c.cmdi-1].key[0]-
			c.cmd[c.cmdi].key[0])&7 == 0 {
			if cbuf.B < 0 && cbuf.D < 0 && cbuf.F < 0 && cbuf.U < 0 {
				c.chargei = c.cmdi
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
		} else if n < 1 || n > 7 {
			return fail()
		} else {
			foo = foo || n == 1
		}
	}
	if !foo {
		return fail()
	}
	c.cmdi++
	// Both inputs in a direction to button transition are checked in same the frame
	if c.cmdi < len(c.cmd) && c.cmd[c.cmdi-1].IsDirToButton(c.cmd[c.cmdi]) {
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
		foo := c.chargei == 0 && c.cmdi == 0
		c.Clear(false)
		if foo {
			c.chargei = 0
		}
		return
	}
	if c.cmdi == 1 && c.cmd[0].slash {
		c.curtime = 0
	} else {
		c.curtime++
	}
	c.completeflag = (c.cmdi == len(c.cmd))
	if !c.completeflag && (ai || c.curtime <= c.time) {
		return
	}
	c.Clear(false)
	if c.completeflag {
		// Update buffer only if it's lower. Mugen doesn't do this but it seems like the right thing to do
		c.curbuftime = Max(c.curbuftime, c.buftime+buftime)
	}
}

// Command List refers to the entire set of a character's commands
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

// Read inputs locally
func (cl *CommandList) Input(i int, facing int32, aiLevel float32, ib InputBits) bool {
	if cl.Buffer == nil {
		return false
	}
	step := cl.Buffer.Bb != 0
	if i < 0 && ^i < len(sys.aiInput) {
		sys.aiInput[^i].Update(aiLevel) // 乱数を使うので同期がずれないようここで / Here we use random numbers so we can not get out of sync
	}
	_else := i < 0
	if _else {
		// Do nothing
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
				U = sys.aiInput[i].U() || ib&IB_PU != 0
				D = sys.aiInput[i].D() || ib&IB_PD != 0
				L = sys.aiInput[i].L() || ib&IB_PL != 0
				R = sys.aiInput[i].R() || ib&IB_PR != 0
				a = sys.aiInput[i].a() || ib&IB_A != 0
				b = sys.aiInput[i].b() || ib&IB_B != 0
				c = sys.aiInput[i].c() || ib&IB_C != 0
				x = sys.aiInput[i].x() || ib&IB_X != 0
				y = sys.aiInput[i].y() || ib&IB_Y != 0
				z = sys.aiInput[i].z() || ib&IB_Z != 0
				s = sys.aiInput[i].s() || ib&IB_S != 0
				d = sys.aiInput[i].d() || ib&IB_D != 0
				w = sys.aiInput[i].w() || ib&IB_W != 0
				m = sys.aiInput[i].m() || ib&IB_M != 0
			}
		} else if i < len(sys.inputRemap) {
			U, D, L, R, a, b, c, x, y, z, s, d, w, m = cl.Buffer.InputReader.LocalInput(sys.inputRemap[i])
		}
		var B, F bool
		if facing < 0 {
			B, F = R, L
		} else {
			B, F = L, R
		}
		// Resolve SOCD conflicts
		U, D, B, F = cl.Buffer.InputReader.SocdResolution(U, D, B, F)
		// AssertInput Flags (no assists, can override SOCD)
		// Does not currently work over netplay because flags are stored at the character level rather than system level
		if ib > 0 {
			U = ib&IB_PU != 0 || U
			D = ib&IB_PD != 0 || D
			if facing > 0 {
				B = ib&IB_PL != 0 || B
				F = ib&IB_PR != 0 || F
			} else {
				B = ib&IB_PR != 0 || B
				F = ib&IB_PL != 0 || F
			}
			a = ib&IB_A != 0 || a
			b = ib&IB_B != 0 || b
			c = ib&IB_C != 0 || c
			x = ib&IB_X != 0 || x
			y = ib&IB_Y != 0 || y
			z = ib&IB_Z != 0 || z
			s = ib&IB_S != 0 || s
			d = ib&IB_D != 0 || d
			w = ib&IB_W != 0 || w
			m = ib&IB_M != 0 || m
		}
		// Send inputs to buffer
		cl.Buffer.Input(B, D, F, U, a, b, c, x, y, z, s, d, w, m)
		// TODO: Reorder all instances of B, F like input bits (U, D, L, R)
	}
	return step
}

// Assert commands with a given name for a given time
func (cl *CommandList) Assert(name string, time int32) bool {
	has := false
	for i := range cl.Commands {
		for j := range cl.Commands[i] {
			if cl.Commands[i][j].name == name {
				cl.Commands[i][j].curbuftime = time
				has = true
			}
		}
	}
	return has
}

// Reset commands with a given name
func (cl *CommandList) ClearName(name string) {
	for i := range cl.Commands {
		for j := range cl.Commands[i] {
			if !cl.Commands[i][j].completeflag && cl.Commands[i][j].name == name {
				cl.Commands[i][j].Clear(true)
			}
		}
	}
}

func (cl *CommandList) Step(facing int32, ai, hitpause bool, buftime int32) {
	if cl.Buffer != nil {
		for i := range cl.Commands {
			for j := range cl.Commands[i] {
				cl.Commands[i][j].Step(cl.Buffer, ai, hitpause, buftime)
			}
		}
		// Find completed commands and reset all duplicate instances
		// This loop must be run separately from the previous one
		for i := range cl.Commands {
			for j := range cl.Commands[i] {
				if cl.Commands[i][j].completeflag {
					cl.ClearName(cl.Commands[i][j].name)
					cl.Commands[i][j].completeflag = false
				}
			}
		}
	}
}

func (cl *CommandList) BufReset() {
	if cl.Buffer != nil {
		cl.Buffer.Reset()
		for i := range cl.Commands {
			for j := range cl.Commands[i] {
				cl.Commands[i][j].Clear(false)
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

// Used in Lua scripts
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
