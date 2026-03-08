package main

import (
	"fmt"
	"math"
	"strings"

	"github.com/veandco/go-sdl2/sdl"
)

var ModAlt ModifierKey
var ModCtrlAlt ModifierKey
var ModCtrlAltShift ModifierKey

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
	if ModAlt == 0 {
		ModAlt = NewModifierKey(false, true, false)
		ModCtrlAlt = NewModifierKey(true, true, false)
		ModCtrlAltShift = NewModifierKey(true, true, true)
	}
	sk := &ShortcutKey{}
	sk.Key = key
	sk.Mod = NewModifierKey(ctrl, alt, shift)
	return sk
}

func (sk ShortcutKey) Test(k Key, m ModifierKey) bool {
	trgtMods := sk.Mod & ModCtrlAltShift
	var expandCurr sdl.Keymod
	if (m & sdl.KMOD_GUI) != 0 {
		expandCurr |= sdl.KMOD_GUI
	}
	if (m & sdl.KMOD_CTRL) != 0 {
		expandCurr |= sdl.KMOD_CTRL
	}
	if (m & sdl.KMOD_ALT) != 0 {
		expandCurr |= sdl.KMOD_ALT
	}
	if (m & sdl.KMOD_SHIFT) != 0 {
		expandCurr |= sdl.KMOD_SHIFT
	}

	return k == sk.Key && trgtMods == expandCurr
}

func OnKeyReleased(key Key, mk ModifierKey) {
	if key != KeyUnknown {
		sys.keyState[key] = false
		// sys.keyInput is used as an edge-triggered "new key press" signal for Lua (getKey()).
		if sys.keyInput == key {
			sys.keyInput = KeyUnknown
			sys.keyString = ""
		}
	}
}

func OnKeyPressed(key Key, mk ModifierKey) {
	if key != KeyUnknown {
		// Treat sys.keyInput as edge-triggered. Ignore repeats.
		if sys.keyState[key] {
			return
		}
		sys.keyState[key] = true
		sys.keyInput = key
		sys.esc = sys.esc ||
			key == KeyEscape && (mk&ModCtrlAlt) == 0
		for k, v := range sys.shortcutScripts {
			if sys.netConnection == nil && (sys.replayFile == nil || !v.DebugKey) &&
				//(!sys.paused || sys.frameStepFlag || v.Pause) &&
				(sys.cfg.Debug.AllowDebugKeys || !v.DebugKey) {
				v.Activate = v.Activate || k.Test(key, mk)
			}
		}
		if key == KeyF12 {
			captureScreen()
		}
		if key == KeyF5 && sys.credits != -1 {
			sys.credits += 1
			sys.motif.Snd.play(sys.motif.AttractMode.Credits.Snd, 100, 0, 0, 0, 0)
		}
		if key == KeyEnter && (mk&ModAlt) != 0 {
			sys.window.toggleFullscreen()
		}
		if !sys.gameRunning && sys.netConnection == nil {
			if key == KeyPause {
				sys.paused = !sys.paused
			}
			if key == KeyScrollLock {
				sys.frameStepFlag = true
			}
		}
	}
}

func OnTextEntered(s string) {
	sys.keyString = s
}

// Return the state of all keyboard keys for a player
func GetKeyboardState(kc KeyConfig) [14]bool {
	var out [14]bool

	// If this config is for a joystick, return no input
	if kc.Joy >= 0 {
		return out
	}

	return [14]bool{
		sys.keyState[Key(kc.dU)],
		sys.keyState[Key(kc.dD)],
		sys.keyState[Key(kc.dL)],
		sys.keyState[Key(kc.dR)],
		sys.keyState[Key(kc.bA)],
		sys.keyState[Key(kc.bB)],
		sys.keyState[Key(kc.bC)],
		sys.keyState[Key(kc.bX)],
		sys.keyState[Key(kc.bY)],
		sys.keyState[Key(kc.bZ)],
		sys.keyState[Key(kc.bS)],
		sys.keyState[Key(kc.bD)],
		sys.keyState[Key(kc.bW)],
		sys.keyState[Key(kc.bM)],
	}
}

// Return the state of all joystick keys for a player
// This is now called only once instead of per button and retrieves
// values from a shared state for both buttons and axes.
// All XInput axes and digital buttons are supported.
func GetJoystickState(kc KeyConfig) [14]bool {
	var out [14]bool
	joy := kc.Joy

	// If this config is for keyboard or out of range, return no input
	if joy < 0 || joy >= input.GetMaxJoystickCount() {
		return out
	}

	if !input.IsJoystickPresent(joy) {
		return out
	}

	axes := input.GetJoystickAxes(joy)
	btns := input.GetJoystickButtons(joy)

	// Convert button polling results to bools
	getBtn := func(idx int) bool {
		return idx >= 0 && idx < len(btns) && btns[idx] != 0
	}

	// axes as buttons
	handleAxisBtn := func(axisBtn int) bool {
		var axis int = 0
		if axisBtn == 16 || axisBtn == 17 { // LS_X
			axis = 0
		} else if axisBtn == 15 || axisBtn == 18 { // LS_Y
			axis = 1
		} else if axisBtn == 22 || axisBtn == 23 { // RS_X
			axis = 2
		} else if axisBtn == 21 || axisBtn == 24 { // RS_Y
			axis = 3
		} else if axisBtn == 19 { // LT
			axis = 4
		} else if axisBtn == 20 { // RT
			axis = 5
		} else { // Invalid
			return false
		}
		val := axes[axis]

		// Evaluate LR triggers on the Xbox 360 controller
		if axis == 4 || axis == 5 {
			return val > sys.cfg.Input.XinputTriggerSensitivity
		}

		if val < 0 && (axisBtn == 15 || axisBtn == 16 || axisBtn == 21 || axisBtn == 22) {
			return -val > sys.cfg.Input.ControllerStickSensitivity
		} else if axisBtn == 17 || axisBtn == 18 || axisBtn == 23 || axisBtn == 24 {
			return val > sys.cfg.Input.ControllerStickSensitivity
		} else {
			return false
		}
	}

	// Apply axis button logic
	axisIndices := []int{
		kc.dU, kc.dD, kc.dL, kc.dR,
		kc.bA, kc.bB, kc.bC, kc.bX, kc.bY, kc.bZ,
		kc.bS, kc.bD, kc.bW, kc.bM,
	}
	for i, idx := range axisIndices {
		if idx >= 15 && idx <= 24 {
			out[i] = handleAxisBtn(idx)
		} else {
			out[i] = getBtn(idx)
		}
	}

	return out
}

type CommandSpec struct {
	Cmd            string
	Time           int32
	BufTime        int32
	BufferHitpause bool
	BufferPauseend bool
	StepTime       int32
}

type KeyConfig struct {
	Joy                                                    int
	dU, dD, dL, dR, bA, bB, bC, bX, bY, bZ, bS, bD, bW, bM int
	isInitialized                                          bool
	rumbleOn                                               bool
	GUID                                                   string
}

func (kc *KeyConfig) set(v [14]int) {
	kc.dU = v[0]
	kc.dD = v[1]
	kc.dL = v[2]
	kc.dR = v[3]
	kc.bA = v[4]
	kc.bB = v[5]
	kc.bC = v[6]
	kc.bX = v[7]
	kc.bY = v[8]
	kc.bZ = v[9]
	kc.bS = v[10]
	kc.bD = v[11]
	kc.bW = v[12]
	kc.bM = v[13]
}

func (kc *KeyConfig) swap(kc2 *KeyConfig) {
	joy := kc.Joy
	// dD := kc.dD
	// dL := kc.dL
	// dR := kc.dR
	// dU := kc.dU
	// bA := kc.bA
	// bB := kc.bB
	// bC := kc.bC
	// bD := kc.bD
	// bW := kc.bW
	// bX := kc.bX
	// bY := kc.bY
	// bZ := kc.bZ
	// bM := kc.bM
	// bS := kc.bS

	kc.Joy = kc2.Joy
	// kc.dD = kc2.dD
	// kc.dL = kc2.dL
	// kc.dR = kc2.dR
	// kc.dU = kc2.dU
	// kc.bA = kc2.bA
	// kc.bB = kc2.bB
	// kc.bC = kc2.bC
	// kc.bD = kc2.bD
	// kc.bW = kc2.bW
	// kc.bX = kc2.bX
	// kc.bY = kc2.bY
	// kc.bZ = kc2.bZ
	// kc.bM = kc2.bM
	// kc.bS = kc2.bS

	kc2.Joy = joy
	// kc2.dD = dD
	// kc2.dL = dL
	// kc2.dR = dR
	// kc2.dU = dU
	// kc2.bA = bA
	// kc2.bB = bB
	// kc2.bC = bC
	// kc2.bD = bD
	// kc2.bW = bW
	// kc2.bX = bX
	// kc2.bY = bY
	// kc2.bZ = bZ
	// kc2.bM = bM
	// kc2.bS = bS

	kc.isInitialized = true
	kc2.isInitialized = true
}

type InputBits int16

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

func (ir *InputReader) LocalInput(in int) [14]bool {
	// Keyboard
	var keyIn [14]bool
	if in < len(sys.keyConfig) {
		keyIn = GetKeyboardState(sys.keyConfig[in])
	}

	// Joystick
	var joyIn [14]bool
	if in < len(sys.joystickConfig) {
		joyIn = GetJoystickState(sys.joystickConfig[in])
	}

	// Merge both
	U := keyIn[0] || joyIn[0]
	D := keyIn[1] || joyIn[1]
	L := keyIn[2] || joyIn[2]
	R := keyIn[3] || joyIn[3]
	a := keyIn[4] || joyIn[4]
	b := keyIn[5] || joyIn[5]
	c := keyIn[6] || joyIn[6]
	x := keyIn[7] || joyIn[7]
	y := keyIn[8] || joyIn[8]
	z := keyIn[9] || joyIn[9]
	s := keyIn[10] || joyIn[10]
	d := keyIn[11] || joyIn[11]
	w := keyIn[12] || joyIn[12]
	m := keyIn[13] || joyIn[13]

	// Apply button assist
	// Checked locally so that network inputs are processed before being sent
	if sys.cfg.Input.ButtonAssist {
		result := ir.ButtonAssistCheck([9]bool{a, b, c, x, y, z, s, d, w})
		a, b, c, x, y, z, s, d, w = result[0], result[1], result[2], result[3], result[4], result[5], result[6], result[7], result[8]
	}

	return [14]bool{U, D, L, R, a, b, c, x, y, z, s, d, w, m}
}

func (ir *InputReader) LocalAnalogInput(in int) [6]int8 {
	if in < 0 || in >= len(sys.joystickConfig) {
		return [6]int8{}
	}

	joy := sys.joystickConfig[in].Joy
	if joy < 0 || joy >= len(input.controllerstate) {
		return [6]int8{}
	}
	if input.controllerstate[joy] == nil {
		return [6]int8{}
	}

	return input.controllerstate[joy].Axes
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
func (ir *InputReader) ButtonAssistCheck(current [9]bool) [9]bool {
	var result [9]bool

	// Disable assist during pauses and screenpack inputs
	if sys.paused || !sys.middleOfMatch() {
		// Consume any buffer leftovers so we don't drop inputs when pausing
		// TODO: This will also mean pressing a button then opening the menu in the next frame can select an option. However that is a separate issue
		// Disabling because this may bring more trouble than it's worth at the moment
		//for i := range ir.ButtonAssistBuffer {
		//	result[i] = current[i] || ir.ButtonAssistBuffer[i]
		//	ir.ButtonAssistBuffer[i] = false
		//}
		//return result
		ir.ButtonAssistBuffer = [9]bool{}
		return current
	}

	// Check if any button was held in the previous frame
	prevAny := false
	for i := range ir.ButtonAssistBuffer {
		if ir.ButtonAssistBuffer[i] {
			prevAny = true
			break
		}
	}

	// If any, check button inputs in both the current and previous frames
	// Otherwise check only the previous frame
	for i := range ir.ButtonAssistBuffer {
		result[i] = ir.ButtonAssistBuffer[i] || (current[i] && prevAny)
	}

	// Save current frame's buttons to be checked in the next frame
	ir.ButtonAssistBuffer = current

	return result
}

// This used to hold button state variables (e.g. U), but that didn't have any info we can't derive from the *b (e.g. Ub) vars
type InputBuffer struct {
	InputReader                            *InputReader
	Bb, Db, Fb, Ub, Lb, Rb, Nb             int32 // Current state of buffer
	ab, bb, cb, xb, yb, zb, sb, db, wb, mb int32
	Bp, Dp, Fp, Up, Lp, Rp, Np             int32 // Previous state of buffer
	ap, bp, cp, xp, yp, zp, sp, dp, wp, mp int32
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

// Get the state of any symbol/key combination
// An attempt was made to cache these states in a map, but computing them every time is already faster than looking up a map
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

/*
// Time since last change of any key. Used for ">" type commands
func (__ *InputBuffer) LastChangeTime() int32 {
	dir := Min(Abs(__.Ub), Abs(__.Db), Abs(__.Bb), Abs(__.Fb), Abs(__.Lb), Abs(__.Rb))
	btn := Min(Abs(__.ab), Abs(__.bb), Abs(__.cb), Abs(__.xb), Abs(__.yb), Abs(__.zb), Abs(__.sb), Abs(__.db), Abs(__.wb), Abs(__.mb))

	return Min(dir, btn)
}
*/

// Check if any recently changed key invalidates a ">" step
// TODO: Make this work with the new $ replacement symbol
func (c *Command) GreaterCheckFail(i int, ibuf *InputBuffer) bool {
	// Determine which directional groups to check
	// Otherwise B/F presses can invalidate L/R and vice-versa
	var useLR bool
	for _, sk := range c.steps[i].keys {
		switch sk.key {
		case CK_L, CK_R, CK_UL, CK_UR, CK_DL, CK_DR:
			useLR = true
		}
	}

	// Check each recent key to see if they belong in the step
	checkKey := func(k CommandKey) bool {
		// Press
		if ibuf.State(CommandStepKey{key: k, tilde: false}) == 1 {
			allowed := false
			for _, sk := range c.steps[i].keys {
				if sk.key == k && !sk.tilde {
					allowed = true
					break
				}
			}
			if !allowed {
				return true
			}
			return false
		}
		// Release
		if ibuf.State(CommandStepKey{key: k, tilde: true}) == 1 {
			allowed := false
			for _, sk := range c.steps[i].keys {
				if sk.key == k && sk.tilde {
					allowed = true
					break
				}
			}
			if !allowed {
				return true
			}
		}
		return false
	}

	// Directions
	for _, k := range [2]CommandKey{CK_U, CK_D} {
		if checkKey(k) {
			return true
		}
	}
	if useLR {
		for _, k := range [6]CommandKey{CK_L, CK_R, CK_UL, CK_UR, CK_DL, CK_DR} {
			if checkKey(k) {
				return true
			}
		}
	} else {
		for _, k := range [6]CommandKey{CK_B, CK_F, CK_UF, CK_UB, CK_DF, CK_DB} {
			if checkKey(k) {
				return true
			}
		}
	}

	// Buttons
	for _, k := range [10]CommandKey{CK_a, CK_b, CK_c, CK_x, CK_y, CK_z, CK_s, CK_d, CK_w, CK_m} {
		if checkKey(k) {
			return true
		}
	}

	return false
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

	// Helper to jam a button for a given time
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

	// Pick a random single direction
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
			hasSlash := false
			for _, k := range c.steps[i].keys {
				if k.slash {
					hasSlash = true
				}
			}
			if !hasSlash {
				return
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

		// Check ">" steps
		if c.steps[i].greater && i > 0 && len(c.steps) >= 2 && c.completed[i-1] && !c.completed[i] {
			if c.GreaterCheckFail(i, ibuf) {
				inputMatched = false
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
	Buffer                *InputBuffer
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
		DefaultStepTime:       -1, // Undefined. Later defaults to same as time
		DefaultAutoGreater:    true,
		DefaultBufferTime:     1,
		DefaultBufferHitpause: true,
		DefaultBufferPauseEnd: true,
		DefaultBufferShared:   true,
	}
}

// Compiles a command string and adds it to this CommandList using the provided spec.
func (cl *CommandList) AddCommand(name string, spec CommandSpec) error {
	if cl == nil {
		return fmt.Errorf("AddCommand called on nil CommandList")
	}

	cmdstr := strings.TrimSpace(spec.Cmd)
	if cmdstr == "" {
		// Nothing to add, but not an error.
		return nil
	}

	cm := newCommand()
	cm.name = name

	if err := cm.ReadCommandSymbols(cmdstr, NewCommandKeyRemap()); err != nil {
		return err
	}

	cm.maxtime = spec.Time
	cm.maxbuftime = spec.BufTime
	cm.buffer_hitpause = spec.BufferHitpause
	cm.buffer_pauseend = spec.BufferPauseend
	cm.maxsteptime = spec.StepTime

	cl.Add(*cm)
	return nil
}

// Read inputs from the correct source (local, AI, net or replay) in order to update the input buffer
func (cl *CommandList) InputUpdate(char *Char, controller int) bool {
	if cl.Buffer == nil {
		return false
	}

	isAI := controller < 0

	// Needed for motif
	hadStepped := cl.Buffer.Ub != 0 || cl.Buffer.Db != 0 || cl.Buffer.Lb != 0 || cl.Buffer.Rb != 0

	var buttons [14]bool
	var axes [6]float32

	if isAI {
		if char != nil && !char.asf(ASF_noaibuttonjam) {
			// Since AI inputs use random numbers, we handle them locally to avoid desync
			idx := ^controller
			if idx >= 0 && idx < len(sys.aiInput) {
				aiLevel := sys.aiLevel[char.playerNo]
				sys.aiInput[idx].Update(aiLevel)
				buttons = sys.aiInput[idx].Buttons()
				char.analogAxes = [6]float32{0, 0, 0, 0, 0, 0}
			}
		}
	} else if sys.replayFile != nil {
		buttons = sys.replayFile.readReplayInput(controller)
		rawAxes := sys.replayFile.readReplayInputAnalog(controller)
		axes = NormalizeAxes(&rawAxes)
	} else if sys.netConnection != nil {
		buttons = sys.netConnection.readNetInput(controller)
		rawAxes := sys.netConnection.readNetInputAnalog(controller)
		axes = NormalizeAxes(&rawAxes)
	} else if sys.rollback.session != nil {
		buttons = sys.rollback.readRollbackInput(controller)
		rawAxes := sys.rollback.readRollbackInputAnalog(controller)
		axes = NormalizeAxes(&rawAxes)
	} else {
		// If not AI, replay, or network, then it's a local human player
		if controller >= 0 {
			if controller < len(sys.inputRemap) {
				in := sys.inputRemap[controller] // remapped input index/config
				buttons = cl.Buffer.InputReader.LocalInput(in)
				// Keep analog axes in sync with the same remap used for digital inputs
				if in >= 0 && in < len(sys.joystickConfig) &&
					sys.joystickConfig[in].Joy >= 0 &&
					sys.joystickConfig[in].Joy < input.GetMaxJoystickCount() &&
					input.IsJoystickPresent(sys.joystickConfig[in].Joy) {
					axes = input.GetJoystickAxes(sys.joystickConfig[in].Joy)
				} else {
					axes = [6]float32{0, 0, 0, 0, 0, 0}
				}
			}
		}
	}

	// Convert bool slice back to named inputs
	U, D, L, R := buttons[0], buttons[1], buttons[2], buttons[3]
	a, b, c := buttons[4], buttons[5], buttons[6]
	x, y, z := buttons[7], buttons[8], buttons[9]
	s, d, w, m := buttons[10], buttons[11], buttons[12], buttons[13]
	B, F := L, R

	// UI-only: let Left Stick drive U/D/L/R even if the player's config uses DPAD.
	if char == nil {
		thr := sys.cfg.Input.ControllerStickSensitivity
		// LS_X (axes[0]), LS_Y (axes[1])
		if axes[1] > thr {
			D = true
		} else if axes[1] < -thr {
			U = true
		}
		if axes[0] > thr {
			R = true
		} else if axes[0] < -thr {
			L = true
		}
		//B, F = L, R
		// Apply SOCD resolution for UI too (same helper used for characters).
		//U, D, B, F = cl.Buffer.InputReader.SocdResolution(U, D, B, F)
		//L, R = B, F
	}

	// Character-specific features
	if char != nil {
		// AssertInput flags
		// Skips button assist. Respects SOCD
		ibit := char.inputFlag
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
		if char.inputShift != nil {
			// Collect current input states and prepare remap states
			inputs := []bool{U, D, L, R, a, b, c, x, y, z, s, d, w, m}
			output := make([]bool, len(inputs))

			// Use a map for fast lookup
			swapMap := make(map[int]int)
			for _, pair := range char.inputShift {
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
		if char.fbFlip {
			B, F = R, L
		} else {
			B, F = L, R
		}

		// Resolve SOCD for U/D and B/F
		U, D, B, F = cl.Buffer.InputReader.SocdResolution(U, D, B, F)

		// Get L and R back from B and F
		if char.fbFlip {
			L, R = F, B
		} else {
			L, R = B, F
		}

		// Update analog axes
		for i := 0; i < len(axes); i++ {
			char.analogAxes[i] = axes[i]
		}
	}

	// Send final inputs to buffer
	cl.Buffer.updateInputTime(U, D, L, R, B, F, a, b, c, x, y, z, s, d, w, m)

	// This check is currently needed to prevent screenpack inputs from rapid firing
	// It forces the command list update to wait one frame after a buffer reset
	// Previously it was checked outside of screenpacks as well, but that caused 1 frame delay in several places of the code
	// Such as making players wait one frame after creation to input anything or a continuous NoInput flag only resetting the buffer every two frames
	// https://github.com/ikemen-engine/Ikemen-GO/issues/1201 and https://github.com/ikemen-engine/Ikemen-GO/issues/2203
	if char == nil {
		return hadStepped
	}

	return true
}

// Normalize from [-32768,32767] to [-1.0,1.0]
func NormalizeAxes(axes *[6]int8) [6]float32 {
	const MAX_VALUE float32 = 128.0
	normalizedAxes := [6]float32{0, 0, 0, 0, 0, 0}
	for i := 0; i < len(axes); i++ {
		if (*axes)[i] < 0 {
			normalizedAxes[i] = float32((*axes)[i]) / MAX_VALUE
		} else {
			normalizedAxes[i] = float32((*axes)[i]) / (MAX_VALUE - 1)
		}
	}
	return normalizedAxes
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

// Checks raw controller tokens (A/B/X/Y, DP_*, LS_*, RS_*, LT/RT).
// controllerIdx is the index in sys.commandLists (0-based).
func (cl *CommandList) IsControllerButtonPressed(token string, controllerIdx int) bool {
	token = strings.TrimSpace(token)
	if token == "" {
		return false
	}
	idx, ok := StringToButtonLUT[token]
	if !ok || idx == 25 { // "Not used"
		return false
	}
	// Resolve controllerIdx -> physical joystick index
	joyIdx := controllerIdx
	if controllerIdx >= 0 {
		in := controllerIdx
		if controllerIdx < len(sys.inputRemap) {
			m := sys.inputRemap[controllerIdx]
			if m >= 0 {
				in = m
			}
		}
		if in >= 0 && in < len(sys.joystickConfig) && sys.joystickConfig[in].Joy >= 0 {
			joyIdx = sys.joystickConfig[in].Joy
		} else {
			joyIdx = in
		}
	}

	if joyIdx < 0 || joyIdx >= input.GetMaxJoystickCount() || !input.IsJoystickPresent(joyIdx) {
		return false
	}

	// Axis tokens
	if idx >= 15 && idx <= 24 {
		axes := input.GetJoystickAxes(joyIdx)
		btns := input.GetJoystickButtons(joyIdx)
		// Determine active axis token (sticks first, then triggers)
		active := CheckAxisForDpad(&axes, len(btns))
		if active == "" {
			active = CheckAxisForTrigger(&axes)
		}
		// Axis tokens are treated as "held" state checks here.
		return active != "" && active == token
	}

	// Digital buttons / DPAD: direct state check.
	buttons := input.GetJoystickButtons(joyIdx)
	if len(buttons) == 0 {
		return false
	}
	for i, b := range buttonOrder {
		if int(b) == idx {
			return i < len(buttons) && buttons[i] != 0
		}
	}
	return false
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
	// Determine whether the command is eligible.
	// Charge commands (/) and commands that are too short are excluded.
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

	// Find the first directional key input.
	firstInputKey := originalCmd.cmd[0].key[0]

	var repeatPattern []cmdElem
	repeatPos := -1

	// Starting from the second element, look for an element that contains the same key as the first.
	for i := 1; i < len(originalCmd.cmd); i++ {
		found := false
		for _, k := range originalCmd.cmd[i].key {
			// Compare the raw key while ignoring ~ and $.
			if withoutTildeKey(k) == withoutTildeKey(firstInputKey) {
				found = true
				break
			}
		}
		if found {
			repeatPos = i
			// Treat the sequence from the first input up to just before it reappears as the pattern.
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

	// Replace the second and subsequent key inputs with $N.
	if len(modifiedPattern) > 1 {
		for i := 1; i < len(modifiedPattern); i++ {
			elem := &modifiedPattern[i]
			elem.key = []CommandKey{CK_Ns}
		}
	}

	// Build the auto-generated command.
	newCmdSlice := make([]cmdElem, 0, len(originalCmd.cmd)+len(modifiedPattern))
	newCmdSlice = append(newCmdSlice, modifiedPattern...)
	newCmdSlice = append(newCmdSlice, originalCmd.cmd...)

	// Create a new Command struct.
	generatedCmd := *originalCmd
	generatedCmd.cmd = newCmdSlice
	generatedCmd.held = make([]bool, len(generatedCmd.hold))

	// Add "grace frames" (time extension) based on the number of inputs in the repeating pattern.
	timeExtension := int32(len(modifiedPattern)) * 4
	generatedCmd.maxtime += timeExtension

	return &generatedCmd
}
*/
