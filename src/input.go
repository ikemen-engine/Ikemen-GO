package main

import (
	"encoding/binary"
	"fmt"
	"math"
	"net"
	"os"
	"strings"
	"time"
)

var ModAlt = NewModifierKey(false, true, false)
var ModCtrlAlt = NewModifierKey(true, true, false)
var ModCtrlAltShift = NewModifierKey(true, true, true)

// CommandList > Command > CommandStep > CommandStepKey
type CommandStepKey struct {
	key        CommandKey
	slash      bool
	tilde      bool
	dollar     bool
	chargetime int32
}

type CommandKey byte

const (
	CK_U CommandKey = iota
	CK_D
	CK_B
	CK_F
	CK_L
	CK_R
	CK_UB
	CK_UF
	CK_DB
	CK_DF
	CK_UL
	CK_UR
	CK_DL
	CK_DR
	CK_N
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
)

func (ck CommandStepKey) IsDirectionPress() bool {
	return !ck.tilde && ck.key >= CK_U && ck.key <= CK_N
}

func (ck CommandStepKey) IsDirectionRelease() bool {
	return ck.tilde && ck.key >= CK_U && ck.key <= CK_N
}

func (ck CommandStepKey) IsButtonPress() bool {
	return !ck.tilde && ck.key >= CK_a && ck.key <= CK_m
}

func (ck CommandStepKey) IsButtonRelease() bool {
	return ck.tilde && ck.key >= CK_a && ck.key <= CK_m
}

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
			if sys.netConnection == nil && (sys.replayFile == nil || !v.DebugKey) &&
				//(!sys.paused || sys.step || v.Pause) &&
				(sys.cfg.Debug.AllowDebugKeys || !v.DebugKey) {
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

// There's nothing wrong with this function, but the way code was structured it needed to be called many times per frame to get all buttons
/*
func JoystickState(joy, button int) bool {
	if joy < 0 {
		return sys.keyState[Key(button)]
	}
	if joy >= input.GetMaxJoystickCount() {
		return false
	}
	axes := input.GetJoystickAxes(joy)
	if button >= 0 {
		// Query button state
		btns := input.GetJoystickButtons(joy)

		if button >= len(btns) {
			if len(btns) == 0 {
				return false
				// Prevent OOB errors #2141
			} else if len(axes) > 0 {
				if button == sys.joystickConfig[joy].dR {
					return axes[0] > sys.cfg.Input.ControllerStickSensitivity
				}
				if button == sys.joystickConfig[joy].dL {
					return -axes[0] > sys.cfg.Input.ControllerStickSensitivity
				}

				// Prevent OOB errors #2141
				if len(axes) > 1 {
					if button == sys.joystickConfig[joy].dU {
						return -axes[1] > sys.cfg.Input.ControllerStickSensitivity
					}
					if button == sys.joystickConfig[joy].dD {
						return axes[1] > sys.cfg.Input.ControllerStickSensitivity
					}
				}
				return false
			} else {
				return false
			}
		}

		// override with axes if they exist #2141
		if len(axes) > 0 {
			if button == sys.joystickConfig[joy].dR {
				if axes[0] > sys.cfg.Input.ControllerStickSensitivity {
					btns[button] = 1
				}
			}
			if button == sys.joystickConfig[joy].dL {
				if -axes[0] > sys.cfg.Input.ControllerStickSensitivity {
					btns[button] = 1
				}
			}

			// prevent OOB errors #2141
			if len(axes) > 1 {
				if button == sys.joystickConfig[joy].dU {
					if -axes[1] > sys.cfg.Input.ControllerStickSensitivity {
						btns[button] = 1
					}
				}
				if button == sys.joystickConfig[joy].dD {
					if axes[1] > sys.cfg.Input.ControllerStickSensitivity {
						btns[button] = 1
					}
				}
			}
		}

		return btns[button] != 0
	} else {
		// Query axis state
		axis := -button - 1
		if axis >= len(axes)*2 {
			return false
		}

		// Read value and invert sign for odd indices
		val := axes[axis/2] * float32((axis&1)*2-1)

		var joyName = input.GetJoystickName(joy)

		// Evaluate LR triggers on the Xbox 360 controller
		if (axis == 9 || axis == 11) && (strings.Contains(joyName, "XInput") || strings.Contains(joyName, "X360") ||
			strings.Contains(joyName, "Xbox Wireless") || strings.Contains(joyName, "Xbox Elite") || strings.Contains(joyName, "Xbox One") ||
			strings.Contains(joyName, "Xbox Series") || strings.Contains(joyName, "Xbox Adaptive")) {
			return val > sys.cfg.Input.XinputTriggerSensitivity
		}

		// Ignore trigger axis on PS4 (We already have buttons)
		if (axis >= 6 && axis <= 9) && joyName == "PS4 Controller" {
			return false
		}

		return val > sys.cfg.Input.ControllerStickSensitivity
	}
}
*/

// Checks keyboard and/or joystick input states
// This is now called only once instead of per button
// Note: Joystick axes cannot be assigned to buttons, only directions
// TODO: Maybe an even better solution would be to poll keyboard and joysticks in the same place once per frame then use that cache
func ControllerState(kc KeyConfig) [14]bool {
	var out [14]bool
	joy := kc.Joy

	// Keyboard early exit
	if joy < 0 {
		return [14]bool{
			sys.keyState[Key(kc.dU)],
			sys.keyState[Key(kc.dD)],
			sys.keyState[Key(kc.dL)],
			sys.keyState[Key(kc.dR)],
			sys.keyState[Key(kc.kA)],
			sys.keyState[Key(kc.kB)],
			sys.keyState[Key(kc.kC)],
			sys.keyState[Key(kc.kX)],
			sys.keyState[Key(kc.kY)],
			sys.keyState[Key(kc.kZ)],
			sys.keyState[Key(kc.kS)],
			sys.keyState[Key(kc.kD)],
			sys.keyState[Key(kc.kW)],
			sys.keyState[Key(kc.kM)],
		}
	}

	// Joystick out of bounds
	if joy >= input.GetMaxJoystickCount() {
		return out
	}

	axes := input.GetJoystickAxes(joy)
	btns := input.GetJoystickButtons(joy)
	joyName := input.GetJoystickName(joy)

	// Convert button polling results to bools
	getBtn := func(idx int) bool {
		return idx >= 0 && idx < len(btns) && btns[idx] != 0
	}

	// Convert axes polling results to bools
	getDir := func(axisIdx int, sign float32, btnIdx int) bool {
		// Check axes normally
		if axisIdx >= 0 && axisIdx < len(axes) {
			if sign*axes[axisIdx] > sys.cfg.Input.ControllerStickSensitivity {
				return true
			}
		}

		// Fallback: override even if button index is OOB
		if len(axes) > 0 {
			switch btnIdx {
			case kc.dL:
				if -axes[0] > sys.cfg.Input.ControllerStickSensitivity {
					return true
				}
			case kc.dR:
				if axes[0] > sys.cfg.Input.ControllerStickSensitivity {
					return true
				}
			case kc.dU:
				if len(axes) > 1 && -axes[1] > sys.cfg.Input.ControllerStickSensitivity {
					return true
				}
			case kc.dD:
				if len(axes) > 1 && axes[1] > sys.cfg.Input.ControllerStickSensitivity {
					return true
				}
			}
		}

		// Fallback to buttons
		return getBtn(btnIdx)
	}

	// Directions
	out[0] = getDir(1, -1, kc.dU)
	out[1] = getDir(1, +1, kc.dD)
	out[2] = getDir(0, -1, kc.dL)
	out[3] = getDir(0, +1, kc.dR)

	// Buttons
	out[4] = getBtn(kc.kA)
	out[5] = getBtn(kc.kB)
	out[6] = getBtn(kc.kC)
	out[7] = getBtn(kc.kX)
	out[8] = getBtn(kc.kY)
	out[9] = getBtn(kc.kZ)
	out[10] = getBtn(kc.kS)
	out[11] = getBtn(kc.kD)
	out[12] = getBtn(kc.kW)
	out[13] = getBtn(kc.kM)

	// Negative indices: axes as buttons (triggers)
	handleAxisBtn := func(axisBtn int) bool {
		if axisBtn >= 0 {
			return false
		}
		axis := -axisBtn - 1
		if axis >= len(axes)*2 {
			return false
		}

		// Read value and invert sign for odd indices
		val := axes[axis/2] * float32((axis&1)*2-1)

		// Evaluate LR triggers on the Xbox 360 controller
		if (axis == 9 || axis == 11) && (strings.Contains(joyName, "XInput") ||
			strings.Contains(joyName, "X360") ||
			strings.Contains(joyName, "Xbox Wireless") ||
			strings.Contains(joyName, "Xbox Elite") ||
			strings.Contains(joyName, "Xbox One") ||
			strings.Contains(joyName, "Xbox Series") ||
			strings.Contains(joyName, "Xbox Adaptive")) {
			return val > sys.cfg.Input.XinputTriggerSensitivity
		}

		// Ignore trigger axis on PS4 (We already have buttons)
		if (axis >= 6 && axis <= 9) && joyName == "PS4 Controller" {
			return false
		}

		return val > sys.cfg.Input.ControllerStickSensitivity
	}

	// Apply axis button logic
	axisIndices := []int{
		kc.dU, kc.dD, kc.dL, kc.dR,
		kc.kA, kc.kB, kc.kC, kc.kX, kc.kY, kc.kZ,
		kc.kS, kc.kD, kc.kW, kc.kM,
	}
	for i, idx := range axisIndices {
		if idx < 0 {
			out[i] = handleAxisBtn(idx)
		}
	}

	return out
}

type KeyConfig struct {
	Joy, dU, dD, dL, dR, kA, kB, kC, kX, kY, kZ, kS, kD, kW, kM int
	GUID                                                        string
	isInitialized                                               bool
}

func (kc *KeyConfig) swap(kc2 *KeyConfig) {
	// joy := kc.Joy
	dD := kc.dD
	dL := kc.dL
	dR := kc.dR
	dU := kc.dU
	kA := kc.kA
	kB := kc.kB
	kC := kc.kC
	kD := kc.kD
	kW := kc.kW
	kX := kc.kX
	kY := kc.kY
	kZ := kc.kZ
	kM := kc.kM
	kS := kc.kS

	// kc.Joy = kc2.Joy
	kc.dD = kc2.dD
	kc.dL = kc2.dL
	kc.dR = kc2.dR
	kc.dU = kc2.dU
	kc.kA = kc2.kA
	kc.kB = kc2.kB
	kc.kC = kc2.kC
	kc.kD = kc2.kD
	kc.kW = kc2.kW
	kc.kX = kc2.kX
	kc.kY = kc2.kY
	kc.kZ = kc2.kZ
	kc.kM = kc2.kM
	kc.kS = kc2.kS

	// kc2.Joy = joy
	kc2.dD = dD
	kc2.dL = dL
	kc2.dR = dR
	kc2.dU = dU
	kc2.kA = kA
	kc2.kB = kB
	kc2.kC = kC
	kc2.kD = kD
	kc2.kW = kW
	kc2.kX = kX
	kc2.kY = kY
	kc2.kZ = kZ
	kc2.kM = kM
	kc2.kS = kS

	kc.isInitialized = true
	kc2.isInitialized = true
}

/*
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
*/

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
func (ibit *InputBits) KeysToBits(buttons [14]bool) {
	*ibit = InputBits(Btoi(buttons[0]) |
		Btoi(buttons[1])<<1 |
		Btoi(buttons[2])<<2 |
		Btoi(buttons[3])<<3 |
		Btoi(buttons[4])<<4 |
		Btoi(buttons[5])<<5 |
		Btoi(buttons[6])<<6 |
		Btoi(buttons[7])<<7 |
		Btoi(buttons[8])<<8 |
		Btoi(buttons[9])<<9 |
		Btoi(buttons[10])<<10 |
		Btoi(buttons[11])<<11 |
		Btoi(buttons[12])<<12 |
		Btoi(buttons[13])<<13)
}

func (ibit InputBits) RollbackBitsToKeys(cb *InputBuffer, facing int32) {
	var U, D, L, R, B, F, a, b, c, x, y, z, s, d, w, m bool
	// Convert bits to logical symbols
	U = ibit&IB_PU != 0
	D = ibit&IB_PD != 0
	L = ibit&IB_PL != 0
	R = ibit&IB_PR != 0
	if facing < 0 {
		B, F = ibit&IB_PR != 0, ibit&IB_PL != 0
	} else {
		B, F = ibit&IB_PL != 0, ibit&IB_PR != 0
	}
	a = ibit&IB_A != 0
	b = ibit&IB_B != 0
	c = ibit&IB_C != 0
	x = ibit&IB_X != 0
	y = ibit&IB_Y != 0
	z = ibit&IB_Z != 0
	s = ibit&IB_S != 0
	d = ibit&IB_D != 0
	w = ibit&IB_W != 0
	m = ibit&IB_M != 0
	// Absolute priority SOCD resolution is enforced during netplay
	// TODO: Port the other options as well
	if U && D {
		D = false
	}
	if B && F {
		B = false
		if facing < 0 {
			R = false
		} else {
			L = false
		}
	}
	cb.updateInputTime(U, D, L, R, B, F, a, b, c, x, y, z, s, d, w, m)
}

// Convert received input bits back into keys
func (ibit InputBits) BitsToKeys() [14]bool {
	var U, D, L, R, a, b, c, x, y, z, s, d, w, m bool

	// Convert bits to logical symbols
	U = ibit&IB_PU != 0
	D = ibit&IB_PD != 0
	L = ibit&IB_PL != 0
	R = ibit&IB_PR != 0
	a = ibit&IB_A != 0
	b = ibit&IB_B != 0
	c = ibit&IB_C != 0
	x = ibit&IB_X != 0
	y = ibit&IB_Y != 0
	z = ibit&IB_Z != 0
	s = ibit&IB_S != 0
	d = ibit&IB_D != 0
	w = ibit&IB_W != 0
	m = ibit&IB_M != 0

	return [14]bool{U, D, L, R, a, b, c, x, y, z, s, d, w, m}
}

type CommandKeyRemap struct {
	a, b, c, x, y, z, s, d, w, m CommandKey
}

func NewCommandKeyRemap() *CommandKeyRemap {
	return &CommandKeyRemap{CK_a, CK_b, CK_c, CK_x, CK_y, CK_z, CK_s, CK_d, CK_w, CK_m}
}

type InputReader struct {
	SocdAllow          [4]bool // Up, down, back, forward
	SocdFirst          [4]bool
	ButtonAssistBuffer [9]bool
}

func NewInputReader() *InputReader {
	return &InputReader{}
}

func (ir *InputReader) Reset() {
	*ir = InputReader{}
}

func (ir *InputReader) LocalInput(in int, script bool) [14]bool {
	var U, D, L, R, a, b, c, x, y, z, s, d, w, m bool

	// Keyboard
	if in < len(sys.keyConfig) {
		joy := sys.keyConfig[in].Joy
		if joy < 0 {
			buttons := ControllerState(sys.keyConfig[in])
			U = buttons[0]
			D = buttons[1]
			L = buttons[2]
			R = buttons[3]
			a = buttons[4]
			b = buttons[5]
			c = buttons[6]
			x = buttons[7]
			y = buttons[8]
			z = buttons[9]
			s = buttons[10]
			d = buttons[11]
			w = buttons[12]
			m = buttons[13]
		}
		/*
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
		*/
	}

	// Joystick
	if in < len(sys.joystickConfig) {
		joy := sys.joystickConfig[in].Joy
		if joy >= 0 {
			buttons := ControllerState(sys.joystickConfig[in])
			U = U || buttons[0] // Does not override keyboard
			D = D || buttons[1]
			L = L || buttons[2]
			R = R || buttons[3]
			a = a || buttons[4]
			b = b || buttons[5]
			c = c || buttons[6]
			x = x || buttons[7]
			y = y || buttons[8]
			z = z || buttons[9]
			s = s || buttons[10]
			d = d || buttons[11]
			w = w || buttons[12]
			m = m || buttons[13]
		}
	}

	/*
		if in < len(sys.joystickConfig) {
			joyS := sys.joystickConfig[in].Joy
			if joyS >= 0 {
				U = U || sys.joystickConfig[in].U() // Does not override keyboard
				D = D || sys.joystickConfig[in].D()
				L = L || sys.joystickConfig[in].L()
				R = R || sys.joystickConfig[in].R()
				a = a || sys.joystickConfig[in].a()
				b = b || sys.joystickConfig[in].b()
				c = c || sys.joystickConfig[in].c()
				x = x || sys.joystickConfig[in].x()
				y = y || sys.joystickConfig[in].y()
				z = z || sys.joystickConfig[in].z()
				s = s || sys.joystickConfig[in].s()
				d = d || sys.joystickConfig[in].d()
				w = w || sys.joystickConfig[in].w()
				m = m || sys.joystickConfig[in].m()
			}
		}
	*/

	// Button assist is checked locally so that the sent inputs are already processed
	if sys.cfg.Input.ButtonAssist {
		if script {
			ir.ButtonAssistBuffer = [9]bool{}
		} else {
			result := ir.ButtonAssistCheck([9]bool{a, b, c, x, y, z, s, d, w})
			a, b, c, x, y, z, s, d, w = result[0], result[1], result[2], result[3], result[4], result[5], result[6], result[7], result[8]
		}
	}

	return [14]bool{U, D, L, R, a, b, c, x, y, z, s, d, w, m}
}

// Resolve Simultaneous Opposing Cardinal Directions (SOCD)
// Left and Right are solved in CommandList Input based on B and F outcome
func (ir *InputReader) SocdResolution(U, D, B, F bool) (bool, bool, bool, bool) {
	method := sys.cfg.Input.SOCDResolution

	// Neutral resolution is enforced during netplay
	// Note: Since configuration does not work online yet, it's best if the forced setting matches the default config
	// TODO: Figure out how to make local options work online
	if sys.netConnection != nil || sys.replayFile != nil {
		method = 4
	}

	// Resolve U and D conflicts based on SOCD resolution config
	resolveUD := func(U, D bool) (bool, bool) {
		// Check first direction held
		if method == 1 || method == 3 {
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
		}
		// Apply SOCD resolution according to config
		if D && U {
			switch method {
			case 0: // Allow both directions (no resolution)
				ir.SocdAllow[0] = true
				ir.SocdAllow[1] = true
			case 1: // Last direction priority
				if ir.SocdFirst[0] {
					ir.SocdAllow[0] = false
					ir.SocdAllow[1] = true
				} else {
					ir.SocdAllow[0] = true
					ir.SocdAllow[1] = false
				}
			case 2: // Absolute priority (offense over defense)
				ir.SocdAllow[0] = true
				ir.SocdAllow[1] = false
			case 3: // First direction priority
				if ir.SocdFirst[0] {
					ir.SocdAllow[0] = true
					ir.SocdAllow[1] = false
				} else {
					ir.SocdAllow[0] = false
					ir.SocdAllow[1] = true
				}
			default: // Deny either direction (neutral resolution)
				ir.SocdAllow[0] = false
				ir.SocdAllow[1] = false
			}
		} else {
			ir.SocdAllow[0] = true
			ir.SocdAllow[1] = true
		}

		return U, D
	}

	// Resolve B and F conflicts based on SOCD resolution config
	resolveBF := func(B, F bool) (bool, bool) {
		// Check first direction held
		if method == 1 || method == 3 {
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
		}
		// Apply SOCD resolution according to config
		if B && F {
			switch method {
			case 0: // Allow both directions (no resolution)
				ir.SocdAllow[2] = true
				ir.SocdAllow[3] = true
			case 1: // Last direction priority
				if ir.SocdFirst[3] {
					ir.SocdAllow[2] = true
					ir.SocdAllow[3] = false
				} else {
					ir.SocdAllow[2] = false
					ir.SocdAllow[3] = true
				}
			case 2: // Absolute priority (offense over defense)
				ir.SocdAllow[2] = false
				ir.SocdAllow[3] = true
			case 3: // First direction priority
				if ir.SocdFirst[3] {
					ir.SocdAllow[2] = false
					ir.SocdAllow[3] = true
				} else {
					ir.SocdAllow[2] = true
					ir.SocdAllow[3] = false
				}
			default: // Deny either direction (neutral resolution)
				ir.SocdAllow[2] = false
				ir.SocdAllow[3] = false
			}
		} else {
			ir.SocdAllow[2] = true
			ir.SocdAllow[3] = true
		}

		return B, F
	}

	// Resolve up and down
	U, D = resolveUD(U, D)
	// Resolve back and forward
	B, F = resolveBF(B, F)
	// Apply resulting resolution
	U = U && ir.SocdAllow[0]
	D = D && ir.SocdAllow[1]
	B = B && ir.SocdAllow[2]
	F = F && ir.SocdAllow[3]

	return U, D, B, F
}

// Add extra frame of leniency when checking button presses
func (ir *InputReader) ButtonAssistCheck(curr [9]bool) [9]bool {
	var result [9]bool

	// Check if any button was pressed in the previous frame
	prev := false
	for i := range ir.ButtonAssistBuffer {
		if ir.ButtonAssistBuffer[i] {
			prev = true
			break
		}
	}

	// Check both current and previous frame if any button was pressed in the previous frame
	// Otherwise just use the previous frame's buttons
	for i := range ir.ButtonAssistBuffer {
		result[i] = ir.ButtonAssistBuffer[i] || (curr[i] && prev)
	}

	// Save current frame's buttons to be checked in the next frame
	ir.ButtonAssistBuffer = curr

	return result
}

// This used to hold button state variables (e.g. U), but that didn't have any info we can't derive from the *b (e.g. Ub) vars
type InputBuffer struct {
	Bb, Db, Fb, Ub, Lb, Rb, Nb             int32 // Current state of buffer
	ab, bb, cb, xb, yb, zb, sb, db, wb, mb int32
	Bp, Dp, Fp, Up, Lp, Rp, Np             int32 // Previous state of buffer
	ap, bp, cp, xp, yp, zp, sp, dp, wp, mp int32
	InputReader                            *InputReader
}

func NewInputBuffer() *InputBuffer {
	return &InputBuffer{
		InputReader: NewInputReader(),
	}
}

func (ib *InputBuffer) Reset() {
	ir := ib.InputReader
	*ib = InputBuffer{
		InputReader: ir,
	}
	ib.InputReader.Reset()
}

// Updates how long ago a char pressed or released a button
func (ib *InputBuffer) updateInputTime(U, D, L, R, B, F, a, b, c, x, y, z, s, d, w, m bool) {
	// Save previous buffer state before updating
	ib.Up = ib.Ub
	ib.Dp = ib.Db
	ib.Lp = ib.Lb
	ib.Rp = ib.Rb
	ib.Bp = ib.Bb
	ib.Fp = ib.Fb
	ib.Np = ib.Nb
	ib.ap = ib.ab
	ib.bp = ib.bb
	ib.cp = ib.cb
	ib.xp = ib.xb
	ib.yp = ib.yb
	ib.zp = ib.zb
	ib.sp = ib.sb
	ib.dp = ib.db
	ib.wp = ib.wb
	ib.mp = ib.mb

	// Function to update current buffer state of each key
	update := func(held bool, buffer *int32) {
		// Detect change
		if held != (*buffer > 0) {
			if held {
				*buffer = 1
			} else {
				*buffer = -1
			}
			return
		}

		// Advance buffer timer
		if held {
			*buffer += 1
		} else {
			*buffer -= 1
		}
	}

	// Directions
	update(U, &ib.Ub)
	update(D, &ib.Db)
	update(L, &ib.Lb)
	update(R, &ib.Rb)
	update(B, &ib.Bb)
	update(F, &ib.Fb)

	// Neutral
	nodir := !(U || D || L || R || B || F)
	update(nodir, &ib.Nb)

	// Buttons
	update(a, &ib.ab)
	update(b, &ib.bb)
	update(c, &ib.cb)
	update(x, &ib.xb)
	update(y, &ib.yb)
	update(z, &ib.zb)
	update(s, &ib.sb)
	update(d, &ib.db)
	update(w, &ib.wb)
	update(m, &ib.mb)
}

// Check the buffer state of each key
func (__ *InputBuffer) State(ck CommandStepKey) int32 {

	// Hold simple directions
	if !ck.tilde && !ck.dollar {
		switch ck.key {

		case CK_U:
			conflict := -Max(__.Bb, Max(__.Db, __.Fb))
			intended := __.Ub
			return Min(conflict, intended)

		case CK_D:
			conflict := -Max(__.Bb, Max(__.Ub, __.Fb))
			intended := __.Db
			return Min(conflict, intended)

		case CK_B:
			conflict := -Max(__.Db, Max(__.Ub, __.Fb))
			intended := __.Bb
			return Min(conflict, intended)

		case CK_F:
			conflict := -Max(__.Db, Max(__.Ub, __.Bb))
			intended := __.Fb
			return Min(conflict, intended)

		case CK_L:
			conflict := -Max(__.Db, Max(__.Ub, __.Rb))
			intended := __.Lb
			return Min(conflict, intended)

		case CK_R:
			conflict := -Max(__.Db, Max(__.Ub, __.Lb))
			intended := __.Rb
			return Min(conflict, intended)

		case CK_UF:
			conflict := -Max(__.Db, __.Bb)
			intended := Min(__.Ub, __.Fb)
			return Min(conflict, intended)

		case CK_UB:
			conflict := -Max(__.Db, __.Fb)
			intended := Min(__.Ub, __.Bb)
			return Min(conflict, intended)

		case CK_DF:
			conflict := -Max(__.Ub, __.Bb)
			intended := Min(__.Db, __.Fb)
			return Min(conflict, intended)

		case CK_DB:
			conflict := -Max(__.Ub, __.Fb)
			intended := Min(__.Db, __.Bb)
			return Min(conflict, intended)

		case CK_UL:
			conflict := -Max(__.Db, __.Rb)
			intended := Min(__.Ub, __.Lb)
			return Min(conflict, intended)

		case CK_UR:
			conflict := -Max(__.Db, __.Lb)
			intended := Min(__.Ub, __.Rb)
			return Min(conflict, intended)

		case CK_DL:
			conflict := -Max(__.Ub, __.Rb)
			intended := Min(__.Db, __.Lb)
			return Min(conflict, intended)

		case CK_DR:
			conflict := -Max(__.Ub, __.Lb)
			intended := Min(__.Db, __.Rb)
			return Min(conflict, intended)

		case CK_N:
			return __.Nb

		}
	}

	// This would be the proper way to do it but it breaks some legacy characters
	// TODO: Add new symbol with this behavior
	/*
	// Hold dollar directions
	if !ck.tilde && ck.dollar {
		switch ck.key {

		case CK_U:
			return __.Ub

		case CK_D:
			return __.Db

		case CK_B:
			return __.Bb

		case CK_F:
			return __.Fb

		case CK_L:
			return __.Lb

		case CK_R:
			return __.Rb

		// What '$' seems to do in Mugen is ignore conflicting directions
		// So it also works on diagonals. For instance, $DB is true even if you also press U or F, but DB isn't
		case CK_UB:
			return Min(__.Ub, __.Bb)

		case CK_UF:
			return Min(__.Ub, __.Fb)

		case CK_DB:
			return Min(__.Db, __.Bb)

		case CK_DF:
			return Min(__.Db, __.Fb)

		case CK_UL:
			return Min(__.Ub, __.Lb)

		case CK_UR:
			return Min(__.Ub, __.Rb)

		case CK_DL:
			return Min(__.Db, __.Lb)

		case CK_DR:
			return Min(__.Db, __.Rb)

		}
	}
*/

	// Hold dollar directions
	// The backward compatible way
	if !ck.tilde && ck.dollar {
		switch ck.key {

		case CK_U:
			if __.Ub > 0 {
				return Min(Abs(__.Ub), Abs(__.Db), Abs(__.Bb), Abs(__.Fb))
			}

		case CK_D:
			if __.Db > 0 {
				return Min(Abs(__.Ub), Abs(__.Db), Abs(__.Bb), Abs(__.Fb))
			}

		case CK_B:
			if __.Bb > 0 {
				return Min(Abs(__.Ub), Abs(__.Db), Abs(__.Bb), Abs(__.Fb))
			}

		case CK_F:
			if __.Fb > 0 {
				return Min(Abs(__.Ub), Abs(__.Db), Abs(__.Bb), Abs(__.Fb))
			}

		case CK_L:
			if __.Lb > 0 {
				return Min(Abs(__.Ub), Abs(__.Db), Abs(__.Lb), Abs(__.Rb))
			}

		case CK_R:
			if __.Rb > 0 {
				return Min(Abs(__.Ub), Abs(__.Db), Abs(__.Lb), Abs(__.Rb))
			}

		// What '$' seems to do in Mugen is ignore conflicting directions
		// So it also works on diagonals. For instance, $DB is true even if you also press U or F, but DB isn't
		case CK_UB:
			if __.Ub > 0 && __.Bb > 0 {
				return Min(Abs(__.Ub), Abs(__.Db), Abs(__.Bb), Abs(__.Fb))
			}

		case CK_UF:
			if __.Ub > 0 && __.Fb > 0 {
				return Min(Abs(__.Ub), Abs(__.Db), Abs(__.Bb), Abs(__.Fb))
			}

		case CK_DB:
			if __.Db > 0 && __.Bb > 0 {
				return Min(Abs(__.Ub), Abs(__.Db), Abs(__.Bb), Abs(__.Fb))
			}

		case CK_DF:
			if __.Db > 0 && __.Fb > 0 {
				return Min(Abs(__.Ub), Abs(__.Db), Abs(__.Bb), Abs(__.Fb))
			}

		case CK_UL:
			if __.Ub > 0 && __.Lb > 0 {
				return Min(Abs(__.Ub), Abs(__.Db), Abs(__.Lb), Abs(__.Rb))
			}

		case CK_UR:
			if __.Ub > 0 && __.Rb > 0 {
				return Min(Abs(__.Ub), Abs(__.Db), Abs(__.Lb), Abs(__.Rb))
			}

		case CK_DL:
			if __.Db > 0 && __.Lb > 0 {
				return Min(Abs(__.Ub), Abs(__.Db), Abs(__.Lb), Abs(__.Rb))
			}

		case CK_DR:
			if __.Db > 0 && __.Rb > 0 {
				return Min(Abs(__.Ub), Abs(__.Db), Abs(__.Lb), Abs(__.Rb))
			}

		}
	}

	// Release simple directions
	if ck.tilde && !ck.dollar {
		switch ck.key {
		case CK_U:
			// If already not held or still held in the previous frame. Prevents for instance UF from trigger ~U
			// https://github.com/ikemen-engine/Ikemen-GO/issues/2626
			if __.Ub < 0 || __.Up > 0 {
				conflict := -Max(__.Bb, Max(__.Db, __.Fb))
				intended := __.Ub
				return -Min(conflict, intended)
			}

		case CK_D:
			if __.Db < 0 || __.Dp > 0 {
				conflict := -Max(__.Bb, Max(__.Ub, __.Fb))
				intended := __.Db
				return -Min(conflict, intended)
			}

		case CK_B:
			if __.Bb < 0 || __.Bp > 0 {
				conflict := -Max(__.Db, Max(__.Ub, __.Fb))
				intended := __.Bb
				return -Min(conflict, intended)
			}

		case CK_F:
			if __.Fb < 0 || __.Fp > 0 {
				conflict := -Max(__.Db, Max(__.Ub, __.Bb))
				intended := __.Fb
				return -Min(conflict, intended)
			}

		case CK_L:
			if __.Lb < 0 || __.Lp > 0 {
				conflict := -Max(__.Db, Max(__.Ub, __.Rb))
				intended := __.Lb
				return -Min(conflict, intended)
			}

		case CK_R:
			if __.Rb < 0 || __.Rp > 0 {
				conflict := -Max(__.Db, Max(__.Ub, __.Lb))
				intended := __.Rb
				return -Min(conflict, intended)
			}

		case CK_UF:
			if (__.Ub < 0 || __.Up > 0) && (__.Fb < 0 || __.Fp > 0) {
				conflict := -Max(__.Db, __.Bb)
				intended := Min(__.Ub, __.Fb)
				return -Min(conflict, intended)
			}

		case CK_UB:
			if (__.Ub < 0 || __.Up > 0) && (__.Bb < 0 || __.Bp > 0) {
				conflict := -Max(__.Db, __.Fb)
				intended := Min(__.Ub, __.Bb)
				return -Min(conflict, intended)
			}

		case CK_DF:
			if (__.Db < 0 || __.Dp > 0) && (__.Fb < 0 || __.Fp > 0) {
				conflict := -Max(__.Ub, __.Bb)
				intended := Min(__.Db, __.Fb)
				return -Min(conflict, intended)
			}

		case CK_DB:
			if (__.Db < 0 || __.Dp > 0) && (__.Bb < 0 || __.Bp > 0) {
				conflict := -Max(__.Ub, __.Fb)
				intended := Min(__.Db, __.Bb)
				return -Min(conflict, intended)
			}

		case CK_UL:
			if (__.Ub < 0 || __.Up > 0) && (__.Lb < 0 || __.Lp > 0) {
				conflict := -Max(__.Db, __.Rb)
				intended := Min(__.Ub, __.Lb)
				return -Min(conflict, intended)
			}

		case CK_UR:
			if (__.Ub < 0 || __.Up > 0) && (__.Rb < 0 || __.Rp > 0) {
				conflict := -Max(__.Db, __.Lb)
				intended := Min(__.Ub, __.Rb)
				return -Min(conflict, intended)
			}

		case CK_DL:
			if (__.Db < 0 || __.Dp > 0) && (__.Lb < 0 || __.Lp > 0) {
				conflict := -Max(__.Ub, __.Rb)
				intended := Min(__.Db, __.Lb)
				return -Min(conflict, intended)
			}

		case CK_DR:
			if (__.Db < 0 || __.Dp > 0) && (__.Rb < 0 || __.Rp > 0) {
				conflict := -Max(__.Ub, __.Lb)
				intended := Min(__.Db, __.Rb)
				return -Min(conflict, intended)
			}

		case CK_N:
			return -__.Nb

		}
	}

	// This would be the proper way to do it but it breaks some legacy characters
	// TODO: Add new symbol with this behavior
	/*
	// Release dollar directions
	if ck.tilde && ck.dollar {
		switch ck.key {
		case CK_U:
			if __.Ub < 0 || __.Up > 0 {
				return -__.Ub
			}

		case CK_D:
			if __.Db < 0 || __.Dp > 0 {
				return -__.Db
			}

		case CK_B:
			if __.Bb < 0 || __.Bp > 0 {
				return -__.Bb
			}

		case CK_F:
			if __.Fb < 0 || __.Fp > 0 {
				return -__.Fb
			}

		case CK_L:
			if __.Lb < 0 || __.Lp > 0 {
				return -__.Lb
			}

		case CK_R:
			if __.Rb < 0 || __.Rp > 0 {
				return -__.Rb
			}

		case CK_UB:
			if (__.Ub < 0 || __.Up > 0) && (__.Bb < 0 || __.Bp > 0) {
				return -Min(__.Ub, __.Bb)
			}

		case CK_UF:
			if (__.Ub < 0 || __.Up > 0) && (__.Fb < 0 || __.Fp > 0) {
				return -Min(__.Ub, __.Fb)
			}

		case CK_DB:
			if (__.Db < 0 || __.Dp > 0) && (__.Bb < 0 || __.Bp > 0) {
				return -Min(__.Db, __.Bb)
			}

		case CK_DF:
			if (__.Db < 0 || __.Dp > 0) && (__.Fb < 0 || __.Fp > 0) {
				return -Min(__.Db, __.Fb)
			}

		case CK_UL:
			if (__.Ub < 0 || __.Up > 0) && (__.Lb < 0 || __.Lp > 0) {
				return -Min(__.Ub, __.Lb)
			}

		case CK_UR:
			if (__.Ub < 0 || __.Up > 0) && (__.Rb < 0 || __.Rp > 0) {
				return -Min(__.Ub, __.Rb)
			}

		case CK_DL:
			if (__.Db < 0 || __.Dp > 0) && (__.Lb < 0 || __.Lp > 0) {
				return -Min(__.Db, __.Lb)
			}

		case CK_DR:
			if (__.Db < 0 || __.Dp > 0) && (__.Rb < 0 || __.Rp > 0) {
				return -Min(__.Db, __.Rb)
			}
		}
	}
	*/

	// Release dollar directions
	// The backward compatible way
	if ck.tilde && ck.dollar {
		switch ck.key {

		case CK_U:
			if __.Ub < 0 || __.Up > 0 {
				if __.Ub < 0 {
					return -__.Ub
				}
				return Min(Abs(__.Db), Abs(__.Bb), Abs(__.Fb))
			}

		case CK_D:
			if __.Db < 0 || __.Dp > 0 {
				if __.Db < 0 {
					return -__.Db
				}
				return Min(Abs(__.Ub), Abs(__.Bb), Abs(__.Fb))
			}

		case CK_B:
			if __.Bb < 0 || __.Bp > 0 {
				if __.Bb < 0 {
					return -__.Bb
				}
				return Min(Abs(__.Ub), Abs(__.Db), Abs(__.Fb))
			}

		case CK_F:
			if __.Fb < 0 || __.Fp > 0 {
				if __.Fb < 0 {
					return -__.Fb
				}
				return Min(Abs(__.Ub), Abs(__.Db), Abs(__.Bb))
			}

		case CK_L:
			if __.Lb < 0 || __.Lp > 0 {
				if __.Lb < 0 {
					return -__.Lb
				}
				return Min(Abs(__.Ub), Abs(__.Db), Abs(__.Rb))
			}

		case CK_R:
			if __.Rb < 0 || __.Rp > 0 {
				if __.Rb < 0 {
					return -__.Rb
				}
				return Min(Abs(__.Ub), Abs(__.Db), Abs(__.Lb))
			}

		case CK_UB:
			if (__.Ub < 0 || __.Up > 0) && (__.Bb < 0 || __.Bp > 0) {
				if __.Ub < 0 || __.Bb < 0 {
					return -Min(__.Ub, __.Bb)
				}
				return Min(Abs(__.Db), Abs(__.Fb))
			}

		case CK_UF:
			if (__.Ub < 0 || __.Up > 0) && (__.Fb < 0 || __.Fp > 0) {
				if __.Ub < 0 || __.Fb < 0 {
					return -Min(__.Ub, __.Fb)
				}
				return Min(Abs(__.Db), Abs(__.Bb))
			}

		case CK_DB:
			if (__.Db < 0 || __.Dp > 0) && (__.Bb < 0 || __.Bp > 0) {
				if __.Db < 0 || __.Bb < 0 {
					return -Min(__.Db, __.Bb)
				}
				return Min(Abs(__.Ub), Abs(__.Fb))
			}

		case CK_DF:
			if (__.Db < 0 || __.Dp > 0) && (__.Fb < 0 || __.Fp > 0) {
				if __.Db < 0 || __.Fb < 0 {
					return -Min(__.Db, __.Fb)
				}
				return Min(Abs(__.Ub), Abs(__.Bb))
			}

		case CK_UL:
			if (__.Ub < 0 || __.Up > 0) && (__.Lb < 0 || __.Lp > 0) {
				if __.Ub < 0 || __.Lb < 0 {
					return -Min(__.Ub, __.Lb)
				}
				return Min(Abs(__.Db), Abs(__.Rb))
			}

		case CK_UR:
			if (__.Ub < 0 || __.Up > 0) && (__.Rb < 0 || __.Rp > 0) {
				if __.Ub < 0 || __.Rb < 0 {
					return -Min(__.Ub, __.Rb)
				}
				return Min(Abs(__.Db), Abs(__.Rb))
			}

		case CK_DL:
			if (__.Db < 0 || __.Dp > 0) && (__.Lb < 0 || __.Lp > 0) {
				if __.Db < 0 || __.Lb < 0 {
					return -Min(__.Db, __.Lb)
				}
				return Min(Abs(__.Ub), Abs(__.Rb))
			}

		case CK_DR:
			if (__.Db < 0 || __.Dp > 0) && (__.Rb < 0 || __.Rp > 0) {
				if __.Db < 0 || __.Rb < 0 {
					return -Min(__.Db, __.Rb)
				}
				return Min(Abs(__.Ub), Abs(__.Rb))
			}
		}

	}

	// Hold buttons
	if !ck.tilde {
		switch ck.key {

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

		}
	}

	// Release buttons
	if ck.tilde {
		switch ck.key {
		case CK_a:
			if __.ab < 0 || __.ap > 0 {
				return -__.ab
			}

		case CK_b:
			if __.bb < 0 || __.bp > 0 {
				return -__.bb
			}

		case CK_c:
			if __.cb < 0 || __.cp > 0 {
				return -__.cb
			}

		case CK_x:
			if __.xb < 0 || __.xp > 0 {
				return -__.xb
			}

		case CK_y:
			if __.yb < 0 || __.yp > 0 {
				return -__.yb
			}

		case CK_z:
			if __.zb < 0 || __.zp > 0 {
				return -__.zb
			}

		case CK_s:
			if __.sb < 0 || __.sp > 0 {
				return -__.sb
			}

		case CK_d:
			if __.db < 0 || __.dp > 0 {
				return -__.db
			}

		case CK_w:
			if __.wb < 0 || __.wp > 0 {
				return -__.wb
			}

		case CK_m:
			if __.mb < 0 || __.mp > 0 {
				return -__.mb
			}
		}
	}

	// Special $N (and ~$N) case
	// This one somehow returns "any change" in Mugen. Since "any neutral" is useless anyway we'll just add support for that
	if ck.dollar && ck.key == CK_N {
		return Min(Abs(__.Ub), Abs(__.Db), Abs(__.Bb), Abs(__.Fb), Abs(__.ab), Abs(__.bb), Abs(__.cb), Abs(__.xb), Abs(__.yb), Abs(__.zb), Abs(__.sb), Abs(__.db), Abs(__.wb))
	}

	return 0
}

// Return charge time of a key
func (ib *InputBuffer) StateCharge(ck CommandStepKey) int32 {
	// Ignore a direction that was just pressed
	// Fixes an issue where charge for a strict direction release (e.g. ~B) will be overridden if you press a different direction in the next frame
	// This is a consequence of imagining charge as "release" like Elecbyte did. Of course, Mugen has that same issue
	ignoreRecent := func(buf int32) int32 {
		if buf == 1 {
			return math.MinInt32
		}
		return buf
	}

	// Hold dollar directions
	if !ck.tilde && ck.dollar {
		switch ck.key {

		case CK_U:
			return ib.Ub

		case CK_D:
			return ib.Db

		case CK_B:
			return ib.Bb

		case CK_F:
			return ib.Fb

		case CK_L:
			return ib.Lb

		case CK_R:
			return ib.Rb

		}
	}

	// Release dollar directions
	if ck.tilde && ck.dollar {
		switch ck.key {

		case CK_U:
			return ib.Up // Check previous buffer state instead

		case CK_D:
			return ib.Dp

		case CK_B:
			return ib.Bp

		case CK_F:
			return ib.Fp

		case CK_L:
			return ib.Lp

		case CK_R:
			return ib.Rp

		}
	}

	// Hold simple directions
	// Mugen doesn't use "hold charge" but we could in the future
	if !ck.tilde && !ck.dollar {
		switch ck.key {

		case CK_U:
			conflict := -Max(ib.Db, Max(ib.Bb, ib.Fb))
			strict := Min(conflict, ib.Ub)
			return Max(0, strict)

		case CK_D:
			conflict := -Max(ib.Ub, Max(ib.Bb, ib.Fb))
			strict := Min(conflict, ib.Db)
			return Max(0, strict)

		case CK_B:
			conflict := -Max(ib.Ub, Max(ib.Db, ib.Fb))
			strict := Min(conflict, ib.Bb)
			return Max(0, strict)

		case CK_F:
			conflict := -Max(ib.Ub, Max(ib.Db, ib.Bb))
			strict := Min(conflict, ib.Fb)
			return Max(0, strict)

		case CK_L:
			conflict := -Max(ib.Ub, Max(ib.Db, ib.Rb))
			strict := Min(conflict, ib.Lb)
			return Max(0, strict)

		case CK_R:
			conflict := -Max(ib.Ub, Max(ib.Db, ib.Lb))
			strict := Min(conflict, ib.Rb)
			return Max(0, strict)

		case CK_UF:
			conflict := -Max(ib.Db, ib.Bb) // Just in case of SOCD funny business
			strict := Min(conflict, Min(ib.Ub, ib.Fb))
			return Max(0, strict)

		case CK_UB:
			conflict := -Max(ib.Db, ib.Fb)
			strict := Min(conflict, Min(ib.Ub, ib.Bb))
			return Max(0, strict)

		case CK_DF:
			conflict := -Max(ib.Ub, ib.Bb)
			strict := Min(conflict, Min(ib.Db, ib.Fb))
			return Max(0, strict)

		case CK_DB:
			conflict := -Max(ib.Ub, ib.Fb)
			strict := Min(conflict, Min(ib.Db, ib.Bb))
			return Max(0, strict)

		case CK_UL:
			conflict := -Max(ib.Db, ib.Rb)
			strict := Min(conflict, Min(ib.Ub, ib.Lb))
			return Max(0, strict)

		case CK_UR:
			conflict := -Max(ib.Db, ib.Lb)
			strict := Min(conflict, Min(ib.Ub, ib.Rb))
			return Max(0, strict)

		case CK_DL:
			conflict := -Max(ib.Ub, ib.Rb)
			strict := Min(conflict, Min(ib.Db, ib.Lb))
			return Max(0, strict)

		case CK_DR:
			conflict := -Max(ib.Ub, ib.Lb)
			strict := Min(conflict, Min(ib.Db, ib.Rb))
			return Max(0, strict)

		case CK_N: // CK_Ns, CK_N: // TODO: Mugen's bugged $N
			return ib.Nb
		}
	}

	// Release simple directions
	if ck.tilde && !ck.dollar {
		switch ck.key {

		case CK_U:
			B := ignoreRecent(ib.Bb)
			D := ignoreRecent(ib.Db)
			F := ignoreRecent(ib.Fb)
			conflict := -Max(B, Max(D, F))
			strict := Min(conflict, ib.Up)
			return Max(0, strict)

		case CK_D:
			U := ignoreRecent(ib.Ub)
			B := ignoreRecent(ib.Bb)
			F := ignoreRecent(ib.Fb)
			conflict := -Max(U, Max(B, F))
			strict := Min(conflict, ib.Dp)
			return Max(0, strict)

		case CK_B:
			U := ignoreRecent(ib.Ub)
			D := ignoreRecent(ib.Db)
			F := ignoreRecent(ib.Fb)
			conflict := -Max(U, Max(D, F))
			strict := Min(conflict, ib.Bp)
			return Max(0, strict)

		case CK_F:
			U := ignoreRecent(ib.Ub)
			D := ignoreRecent(ib.Db)
			B := ignoreRecent(ib.Bb)
			conflict := -Max(U, Max(D, B))
			strict := Min(conflict, ib.Fp)
			return Max(0, strict)

		case CK_L:
			U := ignoreRecent(ib.Ub)
			D := ignoreRecent(ib.Db)
			R := ignoreRecent(ib.Rb)
			conflict := -Max(U, Max(D, R))
			strict := Min(conflict, ib.Lp)
			return Max(0, strict)

		case CK_R:
			U := ignoreRecent(ib.Ub)
			D := ignoreRecent(ib.Db)
			L := ignoreRecent(ib.Lb)
			conflict := -Max(U, Max(D, L))
			strict := Min(conflict, ib.Rp)
			return Max(0, strict)

		case CK_UF:
			D := ignoreRecent(ib.Db)
			B := ignoreRecent(ib.Bb)
			conflict := -Max(D, B)
			strict := Min(conflict, Min(ib.Up, ib.Fp))
			return Max(0, strict)

		case CK_UB:
			D := ignoreRecent(ib.Db)
			F := ignoreRecent(ib.Fb)
			conflict := -Max(D, F)
			strict := Min(conflict, Min(ib.Up, ib.Bp))
			return Max(0, strict)

		case CK_DB:
			U := ignoreRecent(ib.Ub)
			F := ignoreRecent(ib.Fb)
			conflict := -Max(U, F)
			strict := Min(conflict, Min(ib.Dp, ib.Bp))
			return Max(0, strict)

		case CK_DF:
			U := ignoreRecent(ib.Ub)
			B := ignoreRecent(ib.Bb)
			conflict := -Max(U, B)
			strict := Min(conflict, Min(ib.Dp, ib.Fp))
			return Max(0, strict)

		case CK_UL:
			D := ignoreRecent(ib.Db)
			R := ignoreRecent(ib.Rb)
			conflict := -Max(D, R)
			strict := Min(conflict, Min(ib.Up, ib.Lp))
			return Max(0, strict)

		case CK_UR:
			D := ignoreRecent(ib.Db)
			L := ignoreRecent(ib.Lb)
			conflict := -Max(D, L)
			strict := Min(conflict, Min(ib.Up, ib.Rp))
			return Max(0, strict)

		case CK_DL:
			U := ignoreRecent(ib.Ub)
			R := ignoreRecent(ib.Rb)
			conflict := -Max(U, R)
			strict := Min(conflict, Min(ib.Dp, ib.Lp))
			return Max(0, strict)

		case CK_DR:
			U := ignoreRecent(ib.Ub)
			L := ignoreRecent(ib.Lb)
			conflict := -Max(U, L)
			strict := Min(conflict, Min(ib.Dp, ib.Rp))
			return Max(0, strict)

		case CK_N:
			return ib.Np
		}
	}

	// Hold sign diagonals
	// These allow conflicts. Not very useful but is consistent with Mugen's "$" symbol
	if !ck.tilde && ck.dollar {
		switch ck.key {

		case CK_UF:
			return Min(ib.Ub, ib.Fb)

		case CK_UB:
			return Min(ib.Ub, ib.Bb)

		case CK_DF:
			return Min(ib.Db, ib.Fb)

		case CK_DB:
			return Min(ib.Db, ib.Bb)

		case CK_UL:
			return Min(ib.Ub, ib.Lb)

		case CK_UR:
			return Min(ib.Ub, ib.Rb)

		case CK_DL:
			return Min(ib.Db, ib.Lb)

		case CK_DR:
			return Min(ib.Db, ib.Rb)
		}
	}

	// Release sign diagonals
	if ck.tilde && ck.dollar {
		switch ck.key {

		case CK_UF:
			return Min(ib.Up, ib.Fp)

		case CK_UB:
			return Min(ib.Up, ib.Bp)

		case CK_DF:
			return Min(ib.Dp, ib.Fp)

		case CK_DB:
			return Min(ib.Dp, ib.Bp)

		case CK_UL:
			return Min(ib.Up, ib.Lp)

		case CK_UR:
			return Min(ib.Up, ib.Rp)

		case CK_DL:
			return Min(ib.Dp, ib.Lp)

		case CK_DR:
			return Min(ib.Dp, ib.Rp)
		}
	}

	// Hold buttons
	if !ck.tilde {
		switch ck.key {
		case CK_a:
			return ib.ab

		case CK_b:
			return ib.bb

		case CK_c:
			return ib.cb

		case CK_x:
			return ib.xb

		case CK_y:
			return ib.yb

		case CK_z:
			return ib.zb

		case CK_s:
			return ib.sb

		case CK_d:
			return ib.db

		case CK_w:
			return ib.wb

		case CK_m:
			return ib.mb
		}
	}

	// Release buttons
	if ck.tilde {
		switch ck.key {
		case CK_a:
			return ib.ap

		case CK_b:
			return ib.bp

		case CK_c:
			return ib.cp

		case CK_x:
			return ib.xp

		case CK_y:
			return ib.yp

		case CK_z:
			return ib.zp

		case CK_s:
			return ib.sp

		case CK_d:
			return ib.dp

		case CK_w:
			return ib.wp

		case CK_m:
			return ib.mp
		}
	}

	return 0
}

// Time since last change of any key. Used for ">" type commands
func (__ *InputBuffer) LastChangeTime() int32 {
	dir := Min(Abs(__.Ub), Abs(__.Db), Abs(__.Bb), Abs(__.Fb), Abs(__.Lb), Abs(__.Rb))
	btn := Min(Abs(__.ab), Abs(__.bb), Abs(__.cb), Abs(__.xb), Abs(__.yb), Abs(__.zb), Abs(__.sb), Abs(__.db), Abs(__.wb), Abs(__.mb))

	return Min(dir, btn)
}

// NetBuffer holds the inputs that are sent between players
type NetBuffer struct {
	buf              [32]InputBits
	curT, inpT, senT int32
	InputReader      *InputReader
}

func NewNetBuffer() NetBuffer {
	return NetBuffer{
		InputReader: NewInputReader(),
	}
}

func (nb *NetBuffer) reset(time int32) {
	nb.curT, nb.inpT, nb.senT = time, time, time
	nb.InputReader.Reset()
}

// Convert local player's key inputs into input bits for sending
func (nb *NetBuffer) writeNetBuffer(in int) {
	if nb.inpT-nb.curT < 32 {
		nb.buf[nb.inpT&31].KeysToBits(nb.InputReader.LocalInput(in, false))
		nb.inpT++
	}
}

// Read input bits from the net buffer
func (nb *NetBuffer) readNetBuffer() [14]bool {
	if nb.curT < nb.inpT {
		return nb.buf[nb.curT&31].BitsToKeys()
	}
	return [14]bool{}
}

// NetConnection manages the communication between players
type NetConnection struct {
	ln           *net.TCPListener
	conn         *net.TCPConn
	st           NetState
	sendEnd      chan bool
	recvEnd      chan bool
	buf          [MaxPlayerNo]NetBuffer
	locIn        int
	remIn        int
	time         int32
	stoppedcnt   int32
	delay        int32
	recording    *os.File
	host         bool
	preFightTime int32
}

func NewNetConnection() *NetConnection {
	nc := &NetConnection{st: NS_Stop,
		sendEnd: make(chan bool, 1), recvEnd: make(chan bool, 1)}
	nc.sendEnd <- true
	nc.recvEnd <- true

	for i := range nc.buf {
		nc.buf[i] = NewNetBuffer()
	}

	return nc
}

func (nc *NetConnection) Close() {
	if nc.ln != nil {
		nc.ln.Close()
		nc.ln = nil
	}
	if nc.conn != nil {
		nc.conn.Close()
	}
	if nc.sendEnd != nil {
		<-nc.sendEnd
		close(nc.sendEnd)
		nc.sendEnd = nil
	}
	if nc.recvEnd != nil {
		<-nc.recvEnd
		close(nc.recvEnd)
		nc.recvEnd = nil
	}
	nc.conn = nil
}

func (nc *NetConnection) GetHostGuestRemap() (host, guest int) {
	host, guest = -1, -1
	for i, c := range sys.aiLevel {
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
		guest = (host + 1) % len(nc.buf)
	}
	return
}

func (nc *NetConnection) Accept(port string) error {
	if ln, err := net.Listen("tcp", ":"+port); err != nil {
		return err
	} else {
		nc.ln = ln.(*net.TCPListener)
		nc.host = true
		nc.locIn, nc.remIn = nc.GetHostGuestRemap()
		go func() {
			ln := nc.ln
			if conn, err := ln.AcceptTCP(); err == nil {
				nc.conn = conn
				if sys.cfg.Netplay.RollbackNetcode {
					sys.rollback.session.remoteIp = conn.RemoteAddr().(*net.TCPAddr).IP.String()
				}
			}
			ln.Close()
		}()
	}
	return nil
}

func (nc *NetConnection) Connect(server, port string) {
	nc.host = false
	nc.remIn, nc.locIn = nc.GetHostGuestRemap()
	go func() {
		if conn, err := net.Dial("tcp", server+":"+port); err == nil {
			nc.conn = conn.(*net.TCPConn)
		}
	}()
}

func (nc *NetConnection) IsConnected() bool {
	return nc != nil && nc.conn != nil
}

func (nc *NetConnection) readNetInput(i int) [14]bool {
	if i >= 0 && i < len(nc.buf) {
		return nc.buf[sys.inputRemap[i]].readNetBuffer()
	}
	return [14]bool{}
}

func (nc *NetConnection) AnyButton() bool {
	for _, nb := range nc.buf {
		if nb.buf[nb.curT&31]&IB_anybutton != 0 {
			return true
		}
	}
	return false
}

func (nc *NetConnection) Stop() {
	if sys.esc {
		nc.end()
	} else {
		if nc.st != NS_End && nc.st != NS_Error {
			nc.st = NS_Stop
		}
		<-nc.sendEnd
		nc.sendEnd <- true
		<-nc.recvEnd
		nc.recvEnd <- true
	}
}

func (nc *NetConnection) end() {
	if nc.st != NS_Error {
		nc.st = NS_End
	}
	nc.Close()
}

func (nc *NetConnection) readI32() (int32, error) {
	b := [4]byte{}
	if _, err := nc.conn.Read(b[:]); err != nil {
		return 0, err
	}
	return int32(b[0]) | int32(b[1])<<8 | int32(b[2])<<16 | int32(b[3])<<24, nil
}

func (nc *NetConnection) writeI32(i32 int32) error {
	b := [...]byte{byte(i32), byte(i32 >> 8), byte(i32 >> 16), byte(i32 >> 24)}
	if _, err := nc.conn.Write(b[:]); err != nil {
		return err
	}
	return nil
}

func (nc *NetConnection) Synchronize() error {
	if !nc.IsConnected() || nc.st == NS_Error {
		return Error("Cannot connect to the other player")
	}
	nc.Stop()
	var seed int32
	if nc.host {
		seed = Random()
		if err := nc.writeI32(seed); err != nil {
			return err
		}
	} else {
		var err error
		if seed, err = nc.readI32(); err != nil {
			return err
		}
	}
	Srand(seed)
	var pfTime int32
	if nc.host {
		pfTime = sys.preFightTime
		if err := nc.writeI32(pfTime); err != nil {
			return err
		}
	} else {
		var err error
		if pfTime, err = nc.readI32(); err != nil {
			return err
		}
	}
	nc.preFightTime = pfTime
	if nc.recording != nil {
		binary.Write(nc.recording, binary.LittleEndian, &seed)
		binary.Write(nc.recording, binary.LittleEndian, &pfTime)
	}
	if err := nc.writeI32(nc.time); err != nil {
		return err
	}
	if tmp, err := nc.readI32(); err != nil {
		return err
	} else if tmp != nc.time {
		return Error("Synchronization error")
	}
	if sys.rollback.session != nil {
		sys.rollback.session.netTime = nc.time
	}
	nc.buf[nc.locIn].reset(nc.time)
	nc.buf[nc.remIn].reset(nc.time)
	nc.st = NS_Playing
	<-nc.sendEnd
	go func(nb *NetBuffer) {
		defer func() { nc.sendEnd <- true }()
		for nc.st == NS_Playing {
			if nb.senT < nb.inpT {
				if err := nc.writeI32(int32(nb.buf[nb.senT&31])); err != nil {
					nc.st = NS_Error
					return
				}
				nb.senT++
			}
			time.Sleep(time.Millisecond)
		}
		nc.writeI32(-1)
	}(&nc.buf[nc.locIn])
	<-nc.recvEnd
	go func(nb *NetBuffer) {
		defer func() { nc.recvEnd <- true }()
		for nc.st == NS_Playing {
			if nb.inpT-nb.curT < 32 {
				if tmp, err := nc.readI32(); err != nil {
					nc.st = NS_Error
					return
				} else {
					nb.buf[nb.inpT&31] = InputBits(tmp)
					if tmp < 0 {
						nc.st = NS_Stopped
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
			if tmp, err = nc.readI32(); err != nil {
				break
			}
		}
	}(&nc.buf[nc.remIn])
	nc.Update()
	return nil
}

func (nc *NetConnection) Update() bool {
	if nc.st != NS_Stopped {
		nc.stoppedcnt = 0
	}
	if !sys.gameEnd {
		switch nc.st {
		case NS_Stopped:
			nc.stoppedcnt++
			if nc.stoppedcnt > 60 {
				nc.st = NS_End
				break
			}
			fallthrough
		case NS_Playing:
			for {
				foo := Min(nc.buf[nc.locIn].senT, nc.buf[nc.remIn].senT)
				tmp := nc.buf[nc.remIn].inpT + nc.delay>>3 - nc.buf[nc.locIn].inpT
				if tmp >= 0 {
					nc.buf[nc.locIn].writeNetBuffer(0)
					if nc.delay > 0 {
						nc.delay--
					}
				} else if tmp < -1 {
					nc.delay += 4
				}
				if nc.time >= foo {
					if sys.esc || !sys.await(sys.cfg.Config.Framerate) || nc.st != NS_Playing {
						break
					}
					continue
				}
				nc.buf[nc.locIn].curT = nc.time
				nc.buf[nc.remIn].curT = nc.time
				if nc.recording != nil {
					for i := 0; i < MaxSimul*2; i++ {
						binary.Write(nc.recording, binary.LittleEndian, &nc.buf[i].buf[nc.time&31])
					}
				}
				nc.time++
				if nc.time >= foo {
					nc.buf[nc.locIn].writeNetBuffer(0)
				}
				break
			}
		case NS_End, NS_Error:
			sys.esc = true
		}
	}
	if sys.esc {
		nc.end()
	}
	return !sys.gameEnd
}

type ReplayFile struct {
	f      *os.File
	ibit   [MaxPlayerNo]InputBits
	pfTime int32
}

func OpenReplayFile(filename string) *ReplayFile {
	rf := &ReplayFile{}
	rf.f, _ = os.Open(filename)
	return rf
}

func (rf *ReplayFile) Close() {
	if rf.f != nil {
		rf.f.Close()
		rf.f = nil
	}
}

// Read input buttons from replay input
func (rf *ReplayFile) readReplayFile(i int) [14]bool {
	if i >= 0 && i < len(rf.ibit) {
		return rf.ibit[sys.inputRemap[i]].BitsToKeys()
	}
	return [14]bool{}
}

func (rf *ReplayFile) AnyButton() bool {
	for _, b := range rf.ibit {
		if b&IB_anybutton != 0 {
			return true
		}
	}
	return false
}

func (rf *ReplayFile) Synchronize() {
	if rf.f != nil {
		var seed int32
		if binary.Read(rf.f, binary.LittleEndian, &seed) == nil {
			Srand(seed)
		}
		var pfTime int32
		if binary.Read(rf.f, binary.LittleEndian, &pfTime) == nil {
			rf.pfTime = pfTime
			rf.Update()
		}
	}
}

func (rf *ReplayFile) Update() bool {
	if rf.f == nil {
		sys.esc = true
	} else {
		if sys.oldNextAddTime > 0 {
			for i := range rf.ibit {
				rf.ibit[i] = 0
			}
			err := binary.Read(rf.f, binary.LittleEndian, rf.ibit[:MaxSimul*2])
			if err != nil {
				sys.esc = true
			}
		}
		if sys.esc {
			rf.Close()
		}
	}
	return !sys.gameEnd
}

type AiInput struct {
	dir, dirt, at, bt, ct, xt, yt, zt, st, dt, wt, mt int32
}

func (ai *AiInput) Buttons() [14]bool {
	return [14]bool{
		ai.U(), ai.D(), ai.L(), ai.R(),
		ai.a(), ai.b(), ai.c(),
		ai.x(), ai.y(), ai.z(),
		ai.s(), ai.d(), ai.w(), ai.m(),
	}
}

// AI button jamming
func (ai *AiInput) Update(level float32) {
	// Not during intros and win poses
	if sys.intro != 0 {
		ai.dirt, ai.at, ai.bt, ai.ct = 0, 0, 0, 0
		ai.xt, ai.yt, ai.zt, ai.st = 0, 0, 0, 0
		ai.dt, ai.wt, ai.mt = 0, 0, 0
		return
	}
	var chance, time int32 = 15, 60
	jam := func(t *int32) bool {
		(*t)--
		if *t <= 0 {
			// TODO: Balance AI Scaling
			if Rand(1, chance) == 1 {
				*t = Rand(1, time)
				return true
			}
			*t = 0
		}
		return false
	}
	// Pick a random direction to press
	if jam(&ai.dirt) {
		ai.dir = Rand(0, 7)
	}
	chance, time = int32((-11.25*level+165)*7), 30
	jam(&ai.at)
	jam(&ai.bt)
	jam(&ai.ct)
	jam(&ai.xt)
	jam(&ai.yt)
	jam(&ai.zt)
	jam(&ai.dt)
	jam(&ai.wt)
	chance = 3600 // Start is jammed less often
	jam(&ai.st)
	//jam(&ai.mt) // We really don't need the AI to jam the menu button
}

// 0 = U, 1 = UR, 2 = R, 3 = DR, 4 = D, 5 = DL, 6 = L, 7 = UL
func (ai *AiInput) U() bool { return ai.dirt != 0 && (ai.dir == 7 || ai.dir == 0 || ai.dir == 1) }
func (ai *AiInput) D() bool { return ai.dirt != 0 && (ai.dir == 3 || ai.dir == 4 || ai.dir == 5) }
func (ai *AiInput) L() bool { return ai.dirt != 0 && (ai.dir == 5 || ai.dir == 6 || ai.dir == 7) }
func (ai *AiInput) R() bool { return ai.dirt != 0 && (ai.dir == 1 || ai.dir == 2 || ai.dir == 3) }
func (ai *AiInput) a() bool { return ai.at != 0 }
func (ai *AiInput) b() bool { return ai.bt != 0 }
func (ai *AiInput) c() bool { return ai.ct != 0 }
func (ai *AiInput) x() bool { return ai.xt != 0 }
func (ai *AiInput) y() bool { return ai.yt != 0 }
func (ai *AiInput) z() bool { return ai.zt != 0 }
func (ai *AiInput) s() bool { return ai.st != 0 }
func (ai *AiInput) d() bool { return ai.dt != 0 }
func (ai *AiInput) w() bool { return ai.wt != 0 }
func (ai *AiInput) m() bool { return ai.mt != 0 }

// CommandStep refers to each of the steps required to complete a command
// Each step can have multiple keys
type CommandStep struct {
	keys    []CommandStepKey
	greater bool
	orLogic bool
}

// Used to detect consecutive directions
func (cs *CommandStep) IsSingleDirection() bool {
	// Released directions are not taken into account here
	return len(cs.keys) == 1 && (cs.keys[0].IsDirectionPress() || cs.keys[0].IsDirectionRelease())
}

// Check if two command elements can be checked in the same frame
// This logic seems more complex in Mugen because of variable input delay
func (cs *CommandStep) IsDirToButton(next CommandStep) bool {
	// Not if second step must be held
	for _, k := range next.keys {
		if k.slash {
			return false
		}
	}
	// Not if first step includes button press or release
	for _, k := range cs.keys {
		if k.IsButtonPress() || k.IsButtonRelease() {
			return false
		}
	}
	// Not if both steps share keys
	for _, k := range cs.keys {
		for _, n := range next.keys {
			if k.key == n.key {
				return false
			}
		}
	}
	// Yes if second step includes a button press
	for _, n := range next.keys {
		if n.IsButtonPress() {
			return true
		}
	}
	// Yes if release direction then not release direction (includes buttons)
	for _, k := range cs.keys {
		if k.IsDirectionRelease() {
			for _, n := range next.keys {
				if !n.IsDirectionRelease() {
					return true
				}
			}
		}
	}
	return false
}

// Check if two steps are idential. For ">" expansion
func (cs CommandStep) EqualSteps(n CommandStep) bool {
	if cs.greater != n.greater {
		return false
	}
	if len(cs.keys) != len(n.keys) {
		return false
	}
	for i := range cs.keys {
		if cs.keys[i] != n.keys[i] {
			return false
		}
	}
	return true
}

// Command refers to each individual command from the CMD file
type Command struct {
	name                     string
	steps                    []CommandStep
	maxtime, curtime         int32
	maxbuftime, curbuftime   int32
	maxsteptime, cursteptime int32
	autogreater              bool
	buffer_hitpause          bool
	buffer_pauseend          bool
	buffer_shared            bool
	completeframe            bool
	completed                []bool
	stepTimers               []int32
	loopOrder                []int
}

func newCommand() *Command {
	return &Command{maxtime: 1, maxbuftime: 1}
}

// This is used to first compile the commands
func (c *Command) ReadCommandSymbols(cmdstr string, kr *CommandKeyRemap) (err error) {
	// Empty steps case
	if strings.TrimSpace(cmdstr) == "" {
		c.steps = nil
		return
	}

	steps := strings.Split(cmdstr, ",")

	for i, stepkeys := range steps {
		// Add new step
		c.steps = append(c.steps, CommandStep{})
		cs := &c.steps[len(c.steps)-1]
		stepkeys = strings.TrimSpace(stepkeys)

		// The first step is allowed to be blank so that blank commands can be defined
		if i > 0 && len(stepkeys) == 0 {
			err = Error(fmt.Sprintf("Empty command step found"))
			continue
		}

		// Check if using AND or OR logic then split accordingly
		var keyParts []string

		if strings.Contains(stepkeys, "+") && strings.Contains(stepkeys, "|") {
			err = Error("Cannot use both '+' and '|' in the same command step")
			continue
		} else if strings.Contains(stepkeys, "|") {
			cs.orLogic = true
			keyParts = strings.Split(stepkeys, "|")
		} else {
			cs.orLogic = false
			keyParts = strings.Split(stepkeys, "+")
		}

		// Process each of the parts
		for _, part := range keyParts {
			part = strings.TrimSpace(part)
			if len(part) == 0 {
				err = Error("Unexpected '+' with no key")
				continue
			}

			// Check if there's exactly 1 key in each "+" section
			keyCount := 0
			i := 0
			for i < len(part) {
				c0 := part[i]
				// Handle diagonals as a single key
				if (c0 == 'U' || c0 == 'D') && i+1 < len(part) {
					c1 := part[i+1]
					if c1 == 'B' || c1 == 'F' || c1 == 'L' || c1 == 'R' {
						keyCount++
						i += 2
						continue
					}
				}
				// Regular direction or button
				if c0 == 'U' || c0 == 'D' || c0 == 'B' || c0 == 'F' ||
					c0 == 'L' || c0 == 'R' || c0 == 'N' ||
					c0 == 'a' || c0 == 'b' || c0 == 'c' ||
					c0 == 'x' || c0 == 'y' || c0 == 'z' ||
					c0 == 's' || c0 == 'd' || c0 == 'w' || c0 == 'm' {
					keyCount++
				}
				i++
			}
			if keyCount < 1 {
				err = Error("No keys found in command step")
			} else if keyCount > 1 {
				err = Error("Multiple keys found without '+' separator")
			}

			// Parse prefix symbols
			var slash, tilde, dollar bool
			var chargetime int32

			getChar := func() rune {
				if len(part) > 0 {
					return rune(part[0])
				}
				return rune(-1)
			}

			nextChar := func() rune {
				if len(part) > 0 {
					part = strings.TrimSpace(part[1:])
				}
				return getChar()
			}

			parseChargeTime := func() int32 {
				n := int32(0)
				for {
					r := getChar()
					if r >= '0' && r <= '9' {
						n = n*10 + int32(r-'0')
						nextChar()
					} else {
						break
					}
				}
				return n
			}

			// Get prefix symbols first
			for len(part) > 0 {
				switch getChar() {
				case '>':
					if cs.greater {
						err = Error("Duplicate '>' symbol found")
					}
					cs.greater = true // Save to whole step
					nextChar()
				case '~':
					if tilde {
						err = Error("Duplicate '~' symbol found")
					}
					tilde = true
					nextChar()
					n := parseChargeTime()
					if n > 0 {
						chargetime = n
					}
				case '/':
					if slash {
						err = Error("Duplicate '/' symbol found")
					}
					slash = true
					nextChar()
					n := parseChargeTime()
					if n > 0 {
						chargetime = n
					}
				case '$':
					if dollar {
						err = Error("Duplicate '$' symbol found")
					}
					dollar = true
					nextChar()
				default:
					goto ParseKey // Break out of prefix loop
				}
			}

		ParseKey:
			// Get keys
			c0 := getChar()
			switch c0 {
			case 'B', 'D', 'F', 'L', 'R', 'U', 'N':
				var k CommandKey
				switch c0 {
				case 'U':
					k = CK_U
				case 'D':
					k = CK_D
				case 'B':
					k = CK_B
				case 'F':
					k = CK_F
				case 'L':
					k = CK_L
				case 'R':
					k = CK_R
				case 'N':
					k = CK_N
				}
				// Handle diagonals
				if len(part) > 1 {
					c1 := part[1]
					if (c0 == 'U' || c0 == 'D') && (c1 == 'B' || c1 == 'F' || c1 == 'L' || c1 == 'R') {
						switch c1 {
						case 'B':
							if c0 == 'U' {
								k = CK_UB
							} else if c0 == 'D' {
								k = CK_DB
							}
						case 'F':
							if c0 == 'U' {
								k = CK_UF
							} else if c0 == 'D' {
								k = CK_DF
							}
						case 'L':
							if c0 == 'U' {
								k = CK_UL
							} else if c0 == 'D' {
								k = CK_DL
							}
						case 'R':
							if c0 == 'U' {
								k = CK_UR
							} else if c0 == 'D' {
								k = CK_DR
							}
						}
						nextChar()
					}
				}
				cs.keys = append(cs.keys, CommandStepKey{key: k, slash: slash, tilde: tilde, dollar: dollar, chargetime: chargetime})
				nextChar()
			case 'a', 'b', 'c', 'x', 'y', 'z', 's', 'd', 'w', 'm':
				// Maybe too restrictive. Will make people blame poor module code on IkemenVersion characters
				//if dollar {
				//	err = Error("'$' symbol not supported for buttons")
				//}
				// Compile buttons according to remaps
				var k CommandKey
				switch c0 {
				case 'a':
					k = kr.a
				case 'b':
					k = kr.b
				case 'c':
					k = kr.c
				case 'x':
					k = kr.x
				case 'y':
					k = kr.y
				case 'z':
					k = kr.z
				case 's':
					k = kr.s
				case 'd':
					k = kr.d
				case 'w':
					k = kr.w
				case 'm':
					k = kr.m
				}
				cs.keys = append(cs.keys, CommandStepKey{key: k, slash: slash, tilde: tilde, dollar: false, chargetime: chargetime})
				nextChar()
			default:
				err = Error(fmt.Sprintf("Invalid symbol '%c' found", c0))
				nextChar()
			}
		}
	}

	// Expand duplicate directions if applicable
	c.AutoGreaterExpand()

	// Prepare step completion trackers
	c.completed = make([]bool, len(c.steps))
	c.stepTimers = make([]int32, len(c.steps))

	// Determine order in which command steps will be evaluated later
	// Using a reverse order prevents one input from completing two consecutive steps
	// The exception is "IsDirToButton" steps, which are checked forwards precisely so they can be checked in the same frame
	// This reversal of the loop order is the most robust method tried so far
	c.loopOrder = c.loopOrder[:0] // Clear just in case
	for i := len(c.steps) - 1; i >= 0; {
		if i > 0 && c.steps[i-1].IsDirToButton(c.steps[i]) {
			// Forward order for an entire "IsDirToButton" sequence
			start := i - 1
			end := i
			for start > 0 && c.steps[start-1].IsDirToButton(c.steps[start]) {
				start--
			}
			for j := start; j <= end; j++ {
				c.loopOrder = append(c.loopOrder, j)
			}
			i = start - 1
		} else {
			// Reverse order for the rest
			c.loopOrder = append(c.loopOrder, i)
			i--
		}
	}

	return err
}

// Expand consecutive identical directions into "X, >~X, >X"
// This was implemented to try fixing a bug with ">" inputs
// The fix didn't work but maybe this is still worth keeping since Mugen's documentation explicitly mentions doing this
func (c *Command) AutoGreaterExpand() {
	if !c.autogreater || len(c.steps) < 2 {
		return
	}

	// Check if expansion is needed before doing all the work
	needExpansion := false
	for i := 1; i < len(c.steps); i++ {
		if c.steps[i-1].IsSingleDirection() && c.steps[i].IsSingleDirection() && c.steps[i-1].EqualSteps(c.steps[i]) {
			needExpansion = true
			break
		}
	}
	if !needExpansion {
		return
	}

	// Replace command with new expanded command
	// Keep the first step, expand each additional identical step into ">~X, >X"
	newCmd := make([]CommandStep, 0, len(c.steps)*2) // Overestimate new capacity
	newCmd = append(newCmd, c.steps[0])

	for i := 1; i < len(c.steps); i++ {
		prev := c.steps[i-1]
		curr := c.steps[i]

		if prev.IsSingleDirection() && curr.IsSingleDirection() && prev.EqualSteps(curr) {
			// Expand repeat into ">~X, >X"
			newCmd = append(newCmd,
				CommandStep{
					greater: true,
					keys: []CommandStepKey{{
						key:    curr.keys[0].key,
						tilde:  !curr.keys[0].tilde,
						dollar: curr.keys[0].dollar,
					}},
				},
				CommandStep{
					greater: true,
					keys:    curr.keys,
				},
			)
		} else {
			// No expansion, just copy
			newCmd = append(newCmd, curr)
		}
	}

	c.steps = newCmd
}

func (c *Command) Clear(bufreset bool) {
	c.curtime = 0
	c.cursteptime = 0
	if bufreset {
		c.curbuftime = 0
	}
	for i := range c.completed {
		c.completed[i] = false
	}
	for i := range c.stepTimers {
		c.stepTimers[i] = 0
	}
}

// Update an individual command
func (c *Command) Step(ibuf *InputBuffer, ai, isHelper, hpbuf, pausebuf bool, extratime int32) {
	// Skip hitpause buffering
	if !c.buffer_hitpause {
		hpbuf = false
		extratime = 0
	}

	// Skip Pause/SuperPause buffering
	if !c.buffer_pauseend {
		pausebuf = false
		extratime = 0
	}

	// Decrease current buffer timer if not paused
	if c.curbuftime > 0 && !hpbuf && !pausebuf {
		c.curbuftime--
	}

	// Skip blank input commands
	if len(c.steps) == 0 {
		return
	}

	// Update timers and reset expired completed steps
	anydone := false
	for i := range c.steps {
		if c.completed[i] {
			c.stepTimers[i]++
			if c.maxsteptime > 0 && c.stepTimers[i] > c.maxsteptime {
				c.completed[i] = false
				c.stepTimers[i] = 0
				continue // Don't flag "anydone"
			}
			anydone = true
		}
	}

	// Advance overall command timer only if any step is complete
	// Otherwise reset the command
	if anydone {
		c.curtime++
	} else if c.curtime > 0 {
		c.Clear(false)
	}

	// Match inputs to command steps
	// Process steps in the predetermined iteration order
	for _, i := range c.loopOrder {
		// Skip if previous step is not complete
		if i > 0 && !c.completed[i-1] {
			continue
		}

		var inputMatched bool

		// MUGEN's internal AI can't use commands without the "/" symbol on helpers
		if ai && isHelper {
			for _, k := range c.steps[i].keys {
				if !k.slash {
					inputMatched = false
					break
				}
			}
		}

		// Match current inputs to each key of the current command step
		// Ikemen's parser makes /B+a mean "press a while holding B" which seems consistent
		// This does not work in Mugen. For instance "/B+a" and "/a+B" can both be completed by just holding B
		if c.steps[i].orLogic {
			// OR logic: any key matches
			inputMatched = false
			for _, k := range c.steps[i].keys {
				t := ibuf.State(k)
				keyOk := false

				if k.slash {
					keyOk = t > 0
				} else {
					keyOk = t == 1
				}

				// Check if charge is defined and enough charge is stored
				if keyOk && k.chargetime > 1 && ibuf.StateCharge(k) < k.chargetime {
					keyOk = false
				}

				if keyOk {
					inputMatched = true // OR logic already satisfied
					break
				}
			}
		} else {
			// AND logic: all keys match
			inputMatched = true
			for _, k := range c.steps[i].keys {
				t := ibuf.State(k)
				keyOk := false

				if k.slash {
					keyOk = t > 0
				} else {
					keyOk = t == 1
				}

				if keyOk && k.chargetime > 1 && ibuf.StateCharge(k) < k.chargetime {
					keyOk = false
				}

				if !keyOk {
					inputMatched = false // AND logic already failed
					break
				}
			}
		}

		// ">" check
		// This approach has a quirk in that foreign inputs are accepted if they're entered the same frame that the intended input matched
		// Should be harmless because at least that's very hard for a human to perform
		// Out of the methods tried, this one has the best results for the least work
		if !inputMatched && c.steps[i].greater && 
			i > 0 && len(c.steps) >= 2 && c.completed[i-1] && !c.completed[i] {

			// Check if the last change in inputs can be found in the previous step
			hasLast := false
			for _, key := range c.steps[i-1].keys {
				if ibuf.State(key) == ibuf.LastChangeTime() {
					hasLast = true
					break
				}
			}

			// Ikemen used to do a c.Clear(false) here
			// But Mugen seems to do something like this instead. Or it's more like a ">" failure just prevents "inputMatched"
			// This makes mashing some commands like "D, D, D" easier to do
			// Clear previous step only
			if !hasLast {
				c.completed[i-1] = false
				c.stepTimers[i-1] = 0
			}
		}

		if inputMatched {
			// Mark as completed and reset timer
			c.completed[i] = true
			c.stepTimers[i] = 0

			// Clear previous step to prevent refreshing the current one
			if i > 0 {
				c.completed[i-1] = false
				c.stepTimers[i-1] = 0
			}

			// Reset global timer when first step completes (start the window)
			if i == 0 {
				c.curtime = 0
			}
		}
	}

	// Command is complete if last step is completed
	c.completeframe = len(c.completed) > 0 && c.completed[len(c.completed)-1]

	if !c.completeframe {
		// AI ignores timers
		// TODO: Maybe this isn't necessary since the AI already cheats anyway
		if ai {
			return
		}
		// If still within allowed overall time, keep going
		if c.curtime < c.maxtime { // Using <= makes maxtime of 0 or 1 act the same
			return
		}
	}

	// Clear command if complete or timers expired
	c.Clear(false)

	if c.completeframe {
		c.curbuftime = Max(c.curbuftime, c.maxbuftime+extratime)
	}
}

// Command List refers to the entire set of a character's commands
// Each player has multiple lists: one with its own commands, and a copy of each other player's lists
type CommandList struct {
	Buffer                *InputBuffer // TODO: This should exist higher up in the character. Is probably here because of current menu implementation
	Names                 map[string]int
	Commands              [][]Command // [name][commands]
	DefaultTime           int32
	DefaultStepTime       int32
	DefaultAutoGreater    bool
	DefaultBufferTime     int32
	DefaultBufferHitpause bool
	DefaultBufferPauseEnd bool
	DefaultBufferShared   bool
}

func NewCommandList(cb *InputBuffer) *CommandList {
	return &CommandList{
		Buffer:                cb,
		Names:                 make(map[string]int),
		DefaultTime:           15,
		DefaultStepTime:       -1, // Undefined. Later defaults to same as time. Maybe this should be 15 as well
		DefaultAutoGreater:    true,
		DefaultBufferTime:     1,
		DefaultBufferHitpause: true,
		DefaultBufferPauseEnd: true,
		DefaultBufferShared:   true,
	}
}

// Read inputs from the correct source (local, AI, net or replay) in order to update the input buffer
func (cl *CommandList) InputUpdate(controller int, flipbf bool, aiLevel float32, ibit InputBits, shifting [][2]int, script bool) bool {
	if cl.Buffer == nil {
		return false
	}

	// This check is currently needed to prevent screenpack inputs from rapid firing
	// Previously it was checked outside of screenpacks as well, but that caused 1 frame delay in several places of the code
	// Such as making players wait one frame after creation to input anything or a continuous NoInput flag only resetting the buffer every two frames
	// https://github.com/ikemen-engine/Ikemen-GO/issues/1201 and https://github.com/ikemen-engine/Ikemen-GO/issues/2203
	step := true
	if script {
		step = cl.Buffer.Bb != 0
	}

	isAI := controller < 0

	var buttons [14]bool

	if isAI {
		// Since AI inputs use random numbers, we handle them locally to avoid desync
		idx := ^controller
		if idx >= 0 && idx < len(sys.aiInput) {
			sys.aiInput[idx].Update(aiLevel)
			buttons = sys.aiInput[idx].Buttons()
		}
	} else if sys.replayFile != nil {
		buttons = sys.replayFile.readReplayFile(controller)
	} else if sys.netConnection != nil {
		buttons = sys.netConnection.readNetInput(controller)
	} else {
		// If not AI, replay, or network, then it's a local human player
		if controller < len(sys.inputRemap) {
			buttons = cl.Buffer.InputReader.LocalInput(sys.inputRemap[controller], script)
		}
	}

	// Convert bool slice back to named inputs
	U, D, L, R := buttons[0], buttons[1], buttons[2], buttons[3]
	a, b, c := buttons[4], buttons[5], buttons[6]
	x, y, z := buttons[7], buttons[8], buttons[9]
	s, d, w, m := buttons[10], buttons[11], buttons[12], buttons[13]

	// AssertInput flags
	// Skips button assist. Respects SOCD
	if ibit > 0 {
		U = U || ibit&IB_PU != 0
		D = D || ibit&IB_PD != 0
		L = L || ibit&IB_PL != 0
		R = R || ibit&IB_PR != 0
		a = a || ibit&IB_A != 0
		b = b || ibit&IB_B != 0
		c = c || ibit&IB_C != 0
		x = x || ibit&IB_X != 0
		y = y || ibit&IB_Y != 0
		z = z || ibit&IB_Z != 0
		s = s || ibit&IB_S != 0
		d = d || ibit&IB_D != 0
		w = w || ibit&IB_W != 0
		m = m || ibit&IB_M != 0
	}

	// Apply ShiftInput
	if shifting != nil {
		// Collect current input states and prepare remap states
		inputs := []bool{U, D, L, R, a, b, c, x, y, z, s, d, w, m}
		output := make([]bool, len(inputs))

		// Use a map for fast lookup
		swapMap := make(map[int]int)
		for _, pair := range shifting {
			src, dst := pair[0], pair[1]
			swapMap[src] = dst
		}

		// Apply remapping logic to active keys
		for i, active := range inputs {
			if !active {
				continue
			}
			// If current key has a remap, use it
			if dst, ok := swapMap[i]; ok {
				if dst >= 0 && dst < len(output) {
					output[dst] = true // Apply remap to output
				}
				// Negative dest disables input, so do nothing
			} else {
				output[i] = true // No remap, retain original input
			}
		}

		// Assign back to input variables
		U, D, L, R = output[0], output[1], output[2], output[3]
		a, b, c, x, y, z = output[4], output[5], output[6], output[7], output[8], output[9]
		s, d, w, m = output[10], output[11], output[12], output[13]
	}

	// Get B and F from L and R for SOCD resolution
	var B, F bool
	if flipbf {
		B, F = R, L
	} else {
		B, F = L, R
	}

	// Resolve SOCD for U/D and B/F
	U, D, B, F = cl.Buffer.InputReader.SocdResolution(U, D, B, F)

	// Get L and R back from B and F
	if flipbf {
		L, R = F, B
	} else {
		L, R = B, F
	}

	// Send final inputs to buffer
	cl.Buffer.updateInputTime(U, D, L, R, B, F, a, b, c, x, y, z, s, d, w, m)

	// Decide whether commands should be updated
	// Normally they should, but script inputs need this check
	return step
}

// Assert commands with a given name for a given time
func (cl *CommandList) Assert(name string, time int32) bool {
	i, ok := cl.Names[name]
	if !ok {
		return false
	}

	found := false
	for j := range cl.Commands[i] {
		if cl.Commands[i][j].name == name { // Redundant, but safer
			cl.Commands[i][j].curbuftime = time
			found = true
		}
	}

	return found
}

// Reset command when another command with the same name is completed
// This prevents "piano inputs" from triggering the same special move with each button
func (cl *CommandList) ClearName(name string) {
	i, ok := cl.Names[name]
	if !ok {
		return
	}
	for j := range cl.Commands[i] {
		cmd := &cl.Commands[i][j]
		if !cmd.completeframe && cmd.buffer_shared && cmd.name == name { // Name check should be redundant but works as safeguard
			cmd.Clear(false) // Keep their buffer time. Mugen doesn't do this but it seems like the right thing to do
		}
	}
}

// Used when updating commands in each frame
func (cl *CommandList) Step(ai, isHelper, hpbuf, pausebuf bool, extratime int32) {
	if cl.Buffer == nil {
		return
	}

	completed := make(map[string]bool)

	for i := range cl.Commands {
		for j := range cl.Commands[i] {
			cl.Commands[i][j].Step(cl.Buffer, ai, isHelper, hpbuf, pausebuf, extratime)
			if cl.Commands[i][j].completeframe {
				cl.Commands[i][j].Clear(false)           // Clear this specific command
				completed[cl.Commands[i][j].name] = true // Track completed names
				cl.Commands[i][j].completeframe = false
			}
		}
	}

	// Clear duplicates of completed ones
	for name := range completed {
		cl.ClearName(name)
	}
}

func (cl *CommandList) BufReset() {
	if cl.Buffer != nil {
		cl.Buffer.Reset()
	}
	for i := range cl.Commands {
		for j := range cl.Commands[i] {
			cl.Commands[i][j].Clear(true)
		}
	}
}

// Used when compiling commands
func (cl *CommandList) Add(c Command) {
	i, ok := cl.Names[c.name]
	if !ok || i < 0 || i >= len(cl.Commands) {
		i = len(cl.Commands)
		cl.Commands = append(cl.Commands, nil)
	}
	cl.Commands[i] = append(cl.Commands[i], c)
	cl.Names[c.name] = i

	// We won't be needing this fix if the input refactor works out
	/*
		generatedCmd := autoGenerateExtendedCommand(&c)
		if generatedCmd != nil {
			cl.Commands[i] = append(cl.Commands[i], *generatedCmd)
		}
	*/

}

// Used for command trigger
func (cl *CommandList) At(i int) []Command {
	if i < 0 || i >= len(cl.Commands) {
		return nil
	}
	return cl.Commands[i]
}

// Used in Lua scripts
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

// Copy command lists from other players
// For cases where one player's inputs are compared to another's commands
func (cl *CommandList) CopyList(src CommandList) {
	cl.Names = src.Names
	cl.Commands = make([][]Command, len(src.Commands))
	for i, ca := range src.Commands {
		cl.Commands[i] = make([]Command, len(ca))
		copy(cl.Commands[i], ca)
		// These need individual copies or else the slices will point to the original player
		for j, c := range ca {
			cl.Commands[i][j].completed = make([]bool, len(c.completed))
			cl.Commands[i][j].stepTimers = make([]int32, len(c.stepTimers))
		}
	}
}

/*
func withoutTildeKey(k CommandKey) CommandKey {
	if k >= CK_rU && k <= CK_rN {
		return k - (CK_rU - CK_U)
	} else if k >= CK_Us && k <= CK_DRs {
		return k - (CK_Us - CK_U)
	} else if k >= CK_rUs && k <= CK_rNs {
		return k - (CK_rUs - CK_U)
	} else if k >= CK_ra && k <= CK_rm {
		return k - (CK_ra - CK_a)
	}
	return k
}
*/

/*
func autoGenerateExtendedCommand(originalCmd *Command) *Command {
	// 対象コマンドか判定
	// タメコマンド(/)や短すぎるコマンドは対象外
	if len(originalCmd.cmd) < 3 {
		return nil
	}
	for _, ce := range originalCmd.cmd {
		if ce.slash {
			return nil
		}
	}

	if len(originalCmd.cmd[0].key) == 0 {
		return nil
	}

	// 最初の方向キー入力を探す
	firstInputKey := originalCmd.cmd[0].key[0]

	var repeatPattern []cmdElem
	repeatPos := -1

	// 2番目の要素からループを開始し、最初のキーと同じキーを含む要素を探す
	for i := 1; i < len(originalCmd.cmd); i++ {
		found := false
		for _, k := range originalCmd.cmd[i].key {
			// `~` や `$` を無視して純粋なキーが同じか比較
			if withoutTildeKey(k) == withoutTildeKey(firstInputKey) {
				found = true
				break
			}
		}
		if found {
			repeatPos = i
			// 最初の入力から、それが再度現れる直前までをパターンとする
			repeatPattern = originalCmd.cmd[0:repeatPos]
			break
		}
	}

	if repeatPos == -1 {
		return nil
	}

	modifiedPattern := make([]cmdElem, len(repeatPattern))
	for i, ce := range repeatPattern {
		newKeys := make([]CommandKey, len(ce.key))
		copy(newKeys, ce.key)
		modifiedPattern[i] = ce
		modifiedPattern[i].key = newKeys
	}

	// 2番目以降のキー入力を$Nに置き換える
	if len(modifiedPattern) > 1 {
		for i := 1; i < len(modifiedPattern); i++ {
			elem := &modifiedPattern[i]
			elem.key = []CommandKey{CK_Ns}
		}
	}

	// 自動生成コマンドを作成
	newCmdSlice := make([]cmdElem, 0, len(originalCmd.cmd)+len(modifiedPattern))
	newCmdSlice = append(newCmdSlice, modifiedPattern...)
	newCmdSlice = append(newCmdSlice, originalCmd.cmd...)

	// 新規Command構造体を生成
	generatedCmd := *originalCmd
	generatedCmd.cmd = newCmdSlice
	generatedCmd.held = make([]bool, len(generatedCmd.hold))

	// 繰り返しパターンの入力数に応じて猶予フレーム数 を加算する
	timeExtension := int32(len(modifiedPattern)) * 4
	generatedCmd.maxtime += timeExtension

	return &generatedCmd
}
*/
