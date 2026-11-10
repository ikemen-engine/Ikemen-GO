package main

import (
	"encoding/hex"
	"fmt"
	"hash/fnv"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"unsafe"

	"github.com/veandco/go-sdl2/sdl"
)

type ControllerState struct {
	Axes      [6]int8
	Buttons   map[sdl.GameControllerButton]byte
	HasRumble bool
}

type Input struct {
	controllers     [MaxPlayerNo]*sdl.GameController // pads in the controller database
	joysticks       [MaxPlayerNo]*sdl.Joystick       // every pad, including the raw ones SDL has no mapping for
	controllerstate [MaxPlayerNo]*ControllerState
}

type Key = sdl.Keycode
type ModifierKey = sdl.Keymod

const (
	KeyUnknown    = sdl.K_UNKNOWN
	KeyEscape     = sdl.K_ESCAPE
	KeyEnter      = sdl.K_RETURN
	KeyInsert     = sdl.K_INSERT
	KeyF5         = sdl.K_F5
	KeyF12        = sdl.K_F12
	KeyPause      = sdl.K_PAUSE
	KeyScrollLock = sdl.K_SCROLLLOCK
)

var KeyToStringLUT map[sdl.Keycode]string
var StringToKeyLUT map[string]sdl.Keycode

var input Input

// Controller type used to pick glyph artwork, e.g. "ps5". SDL knows common pads,
// the motif can match others by GUID or name.
func (in *Input) GlyphPack(device *InputDevice, glyphs *ControllerGlyphsProperties) string {
	if device == nil {
		return "generic"
	}
	// A pack chosen for this device wins. An unknown name falls through to
	// detection, which also covers a pack the current motif does not define.
	if chosen := sys.cfg.controllerGlyphs(device); chosen != "" {
		if set := glyphs.Controller[chosen]; set != nil && strings.EqualFold(set.Type, glyphDeviceType(device)) {
			return chosen
		}
	}
	if device.Joy < 0 {
		return "keyboard"
	}
	if !in.IsJoystickPresent(device.Joy) {
		return "generic"
	}
	if controller := glyphs.GUID[strings.ToLower(device.GUID)]; controller != "" {
		return strings.ToLower(controller)
	}
	// Longest matching name wins, so "switch 2" beats "switch". Map order is
	// random, so equal-length patterns are decided alphabetically.
	name := strings.ToLower(in.GetJoystickName(device.Joy))
	match, controller := "", ""
	for pattern, candidate := range glyphs.Name {
		if pattern == "" || !strings.Contains(name, strings.ToLower(pattern)) {
			continue
		}
		if len(pattern) > len(match) || (len(pattern) == len(match) && pattern < match) {
			match, controller = pattern, candidate
		}
	}
	if controller != "" {
		return strings.ToLower(controller)
	}
	if in.IsRawJoystick(device.Joy) {
		return "generic"
	}
	switch in.controllers[device.Joy].Type() {
	case sdl.CONTROLLER_TYPE_XBOX360:
		return "xbox360"
	case sdl.CONTROLLER_TYPE_XBOXONE:
		return "xbox"
	case sdl.CONTROLLER_TYPE_PS3:
		return "ps3"
	case sdl.CONTROLLER_TYPE_PS4:
		return "ps4"
	case sdl.CONTROLLER_TYPE_PS5:
		return "ps5"
	case sdl.CONTROLLER_TYPE_NINTENDO_SWITCH_PRO:
		return "switch"
	}
	return "generic"
}

// Reports whether SDL is set to report button positions instead of the labels printed on Nintendo pads.
func (in *Input) GlyphSwapNintendo() bool {
	return sdl.GetHint("SDL_GAMECONTROLLER_USE_BUTTON_LABELS") == "0"
}

func initLUTs() {
	KeyToStringLUT = map[sdl.Keycode]string{
		sdl.K_RETURN:       "RETURN",
		sdl.K_ESCAPE:       "ESCAPE",
		sdl.K_BACKSPACE:    "BACKSPACE",
		sdl.K_TAB:          "TAB",
		sdl.K_SPACE:        "SPACE",
		sdl.K_QUOTE:        "QUOTE",
		sdl.K_COMMA:        "COMMA",
		sdl.K_MINUS:        "MINUS",
		sdl.K_PERIOD:       "PERIOD",
		sdl.K_SLASH:        "SLASH",
		sdl.K_0:            "0",
		sdl.K_1:            "1",
		sdl.K_2:            "2",
		sdl.K_3:            "3",
		sdl.K_4:            "4",
		sdl.K_5:            "5",
		sdl.K_6:            "6",
		sdl.K_7:            "7",
		sdl.K_8:            "8",
		sdl.K_9:            "9",
		sdl.K_SEMICOLON:    "SEMICOLON",
		sdl.K_EQUALS:       "EQUALS",
		sdl.K_LEFTBRACKET:  "LBRACKET",
		sdl.K_BACKSLASH:    "BACKSLASH",
		sdl.K_RIGHTBRACKET: "RBRACKET",
		sdl.K_BACKQUOTE:    "BACKQUOTE",
		sdl.K_a:            "a",
		sdl.K_b:            "b",
		sdl.K_c:            "c",
		sdl.K_d:            "d",
		sdl.K_e:            "e",
		sdl.K_f:            "f",
		sdl.K_g:            "g",
		sdl.K_h:            "h",
		sdl.K_i:            "i",
		sdl.K_j:            "j",
		sdl.K_k:            "k",
		sdl.K_l:            "l",
		sdl.K_m:            "m",
		sdl.K_n:            "n",
		sdl.K_o:            "o",
		sdl.K_p:            "p",
		sdl.K_q:            "q",
		sdl.K_r:            "r",
		sdl.K_s:            "s",
		sdl.K_t:            "t",
		sdl.K_u:            "u",
		sdl.K_v:            "v",
		sdl.K_w:            "w",
		sdl.K_x:            "x",
		sdl.K_y:            "y",
		sdl.K_z:            "z",
		sdl.K_CAPSLOCK:     "CAPSLOCK",
		sdl.K_F1:           "F1",
		sdl.K_F2:           "F2",
		sdl.K_F3:           "F3",
		sdl.K_F4:           "F4",
		sdl.K_F5:           "F5",
		sdl.K_F6:           "F6",
		sdl.K_F7:           "F7",
		sdl.K_F8:           "F8",
		sdl.K_F9:           "F9",
		sdl.K_F10:          "F10",
		sdl.K_F11:          "F11",
		sdl.K_F12:          "F12",
		sdl.K_PRINTSCREEN:  "PRINTSCREEN",
		sdl.K_SCROLLLOCK:   "SCROLLLOCK",
		sdl.K_PAUSE:        "PAUSE",
		sdl.K_INSERT:       "INSERT",
		sdl.K_HOME:         "HOME",
		sdl.K_PAGEUP:       "PAGEUP",
		sdl.K_DELETE:       "DELETE",
		sdl.K_END:          "END",
		sdl.K_PAGEDOWN:     "PAGEDOWN",
		sdl.K_RIGHT:        "RIGHT",
		sdl.K_LEFT:         "LEFT",
		sdl.K_DOWN:         "DOWN",
		sdl.K_UP:           "UP",
		sdl.K_NUMLOCKCLEAR: "NUMLOCKCLEAR",
		sdl.K_KP_DIVIDE:    "KP_DIVIDE",
		sdl.K_KP_MULTIPLY:  "KP_MULTIPLY",
		sdl.K_KP_MINUS:     "KP_MINUS",
		sdl.K_KP_PLUS:      "KP_PLUS",
		sdl.K_KP_ENTER:     "KP_ENTER",
		sdl.K_KP_1:         "KP_1",
		sdl.K_KP_2:         "KP_2",
		sdl.K_KP_3:         "KP_3",
		sdl.K_KP_4:         "KP_4",
		sdl.K_KP_5:         "KP_5",
		sdl.K_KP_6:         "KP_6",
		sdl.K_KP_7:         "KP_7",
		sdl.K_KP_8:         "KP_8",
		sdl.K_KP_9:         "KP_9",
		sdl.K_KP_0:         "KP_0",
		sdl.K_KP_PERIOD:    "KP_PERIOD",
		sdl.K_KP_EQUALS:    "KP_EQUALS",
		sdl.K_F13:          "F13",
		sdl.K_F14:          "F14",
		sdl.K_F15:          "F15",
		sdl.K_F16:          "F16",
		sdl.K_F17:          "F17",
		sdl.K_F18:          "F18",
		sdl.K_F19:          "F19",
		sdl.K_F20:          "F20",
		sdl.K_F21:          "F21",
		sdl.K_F22:          "F22",
		sdl.K_F23:          "F23",
		sdl.K_F24:          "F24",
		sdl.K_MENU:         "MENU",
		sdl.K_LCTRL:        "LCTRL",
		sdl.K_LSHIFT:       "LSHIFT",
		sdl.K_LALT:         "LALT",
		sdl.K_LGUI:         "LGUI",
		sdl.K_RCTRL:        "RCTRL",
		sdl.K_RSHIFT:       "RSHIFT",
		sdl.K_RALT:         "RALT",
		sdl.K_RGUI:         "RGUI",
	}


	// Explicitly allocate the maps here
	StringToKeyLUT = make(map[string]sdl.Keycode)

	for k, v := range KeyToStringLUT {
		StringToKeyLUT[v] = k
	}

	input = Input{}
}

func StringToKey(s string) sdl.Keycode {
	if key, ok := StringToKeyLUT[s]; ok {
		return key
	}
	return sdl.K_UNKNOWN
}

func KeyToString(k sdl.Keycode) string {
	if s, ok := KeyToStringLUT[k]; ok {
		return s
	}
	return ""
}

func NewModifierKey(ctrl, alt, shift bool) (mod sdl.Keymod) {
	if ctrl {
		// Convert Ctrl to Command (⌘) key for macOS if user prefers it
		if runtime.GOOS == "darwin" && sys.cfg.Debug.MacOSUseCommandKey {
			mod |= sdl.KMOD_GUI
		} else {
			mod |= sdl.KMOD_CTRL
		}
	}
	if alt {
		mod |= sdl.KMOD_ALT
	}
	if shift {
		mod |= sdl.KMOD_SHIFT
	}
	return
}

func (input *Input) UpdateGamepadMappings(path string) {
	b, err := os.ReadFile(path)
	if err != nil {
		LogMessage("%v", err)
		return
	}
	lines := strings.Split(string(b), "\n")
	for _, line := range lines {
		sdl.GameControllerAddMapping(line)
	}
}

// SDL joystick behind a pad slot, or nil.
func (input *Input) joystick(joy int) *sdl.Joystick {
	if joy < 0 || joy >= len(input.joysticks) {
		return nil
	}
	return input.joysticks[joy]
}

func (input *Input) IsJoystickPresent(joy int) bool {
	j := input.joystick(joy)
	return j != nil && j.Attached()
}

// True for a pad SDL has no controller database mapping for. Its buttons, hats and axes are
// read directly, so only its Controller Mapping gives them a meaning.
func (input *Input) IsRawJoystick(joy int) bool {
	return input.joystick(joy) != nil && input.controllers[joy] == nil
}

func (input *Input) GetJoystickName(joy int) string {
	if j := input.joystick(joy); j != nil {
		return j.Name()
	}
	return ""
}

func (input *Input) GetJoystickGUID(joy int) string {
	j := input.joystick(joy)
	if j == nil {
		return ""
	}
	// go-sdl2's JoystickGetGUIDString cuts the GUID to 11 of its 16 bytes.
	guid := j.GUID()
	return hex.EncodeToString((*[16]byte)(unsafe.Pointer(&guid))[:])
}

// Keep serial and location separate so a port can distinguish pads with duplicate serials.
func (input *Input) GetJoystickKey(joy int) string {
	j := input.joystick(joy)
	if j == nil {
		return ""
	}
	key := input.GetJoystickGUID(joy)
	serial, location := joystickHardwareIdentity(j.Path())
	if s := strings.TrimSpace(j.Serial()); s != "" {
		serial = s
	}
	for _, part := range []struct {
		kind byte
		id   string
	}{{'s', serial}, {'p', location}} {
		if part.id != "" {
			h := fnv.New64a()
			h.Write([]byte(part.id))
			key += fmt.Sprintf("_%c%016x", part.kind, h.Sum64())
		}
	}
	return key
}

// Recover hardware identity from SDL's path without saving temporary device numbers.
func joystickHardwareIdentity(path string) (serial, location string) {
	switch runtime.GOOS {
	case "windows":
		// HID paths can follow a device or a port; XInput slot paths are temporary.
		if strings.HasPrefix(path, `\\?\`) {
			location = strings.ToLower(path)
		}
	case "linux":
		if resolved, err := filepath.EvalSymlinks(path); err == nil {
			path = resolved
		}
		name := filepath.Base(path)
		if strings.HasPrefix(name, "event") || strings.HasPrefix(name, "js") {
			dir := filepath.Join("/sys/class/input", name, "device")
			uniq, _ := os.ReadFile(filepath.Join(dir, "uniq"))
			phys, _ := os.ReadFile(filepath.Join(dir, "phys"))
			serial, location = strings.TrimSpace(string(uniq)), strings.TrimSpace(string(phys))
		} else if strings.HasPrefix(name, "hidraw") {
			data, _ := os.ReadFile(filepath.Join("/sys/class/hidraw", name, "device", "uevent"))
			for _, line := range strings.Split(string(data), "\n") {
				if v, ok := strings.CutPrefix(line, "HID_UNIQ="); ok {
					serial = strings.TrimSpace(v)
				} else if v, ok := strings.CutPrefix(line, "HID_PHYS="); ok {
					location = strings.TrimSpace(v)
				}
			}
		}
	default:
		// Other SDL backends expose no path or a temporary one. Use serials or numbering.
	}
	return
}

// Whether SDL can rumble the pad in this slot.
func (input *Input) HasRumble(joy int) bool {
	return input.IsJoystickPresent(joy) && input.controllerstate[joy] != nil && input.controllerstate[joy].HasRumble
}

func (input *Input) RumbleController(participant int, lo, hi uint16, ticks uint32) {
	if sys.replayFile != nil {
		return
	}
	nc := sys.netConnection
	if sys.rollback.session != nil {
		nc = sys.rollback.netConnection
	}
	if nc != nil {
		if participant != nc.locIn {
			return
		}
		participant = 0
	}
	for _, id := range sys.inputs.participantDevices(participant) {
		if device := sys.inputs.device(id); device != nil && device.Connected && device.Joy >= 0 && sys.cfg.rumbleEnabled(device) {
			input.rumbleJoystick(device.Joy, lo, hi, ticks)
		}
	}
}

func (input *Input) rumbleJoystick(joy int, lo, hi uint16, ticks uint32) {
	if !input.IsJoystickPresent(joy) || input.controllerstate[joy] == nil {
		return
	}

	// Only pads that can rumble
	if input.controllerstate[joy].HasRumble {
		gls := sys.gameLogicSpeed()

		if gls > 0 && sys.turbo > 0 {
			var framerate_ms uint32 = uint32(math.Ceil(1.0 / float64(gls) * float64(sys.turbo) * 1000.0))
			var buffertime_ms uint32 = framerate_ms >> 1 // makes rumble feel more consistent between frames
			var multiplier float32 = 1.0

			// This makes macOS rumble, which is less pronounced,
			// feel closer to Linux rumble, which is more forceful.
			// TODO: Compare *NIX-like systems with Windows.
			if runtime.GOOS == "darwin" {
				multiplier = 1.625
			}
			if ticks > 0 {
				input.joysticks[joy].Rumble(uint16(float32(lo)*multiplier), uint16(float32(hi)*multiplier), (ticks*framerate_ms)+buffertime_ms)
			} else {
				input.joysticks[joy].Rumble(0, 0, 0)
			}
		}
	}
}

// Raw joystick inputs, as Controller Mapping stores them: BUTTON3, HAT0_UP (_DOWN, _LEFT, _RIGHT),
// AXIS1- or AXIS1+. SDL reads a missing button, hat or axis as released.
var rawHatBits = map[string]sdl.JoystickHat{"UP": sdl.HAT_UP, "DOWN": sdl.HAT_DOWN, "LEFT": sdl.HAT_LEFT, "RIGHT": sdl.HAT_RIGHT}

// How far a raw joystick input is pushed, from 0 to 127. An axis resting at one end,
// like most DirectInput triggers, only counts when it moves away from its rest.
func rawControlLevel(joy int, token string) int32 {
	j := input.joystick(joy)
	if j == nil {
		return 0
	}
	if n, ok := strings.CutPrefix(token, "BUTTON"); ok {
		if i, err := strconv.Atoi(n); err == nil && j.Button(i) != 0 {
			return 127
		}
	} else if n, ok := strings.CutPrefix(token, "HAT"); ok {
		n, dir, _ := strings.Cut(n, "_")
		if i, err := strconv.Atoi(n); err == nil && sdl.JoystickHat(j.Hat(sdl.JoystickHat(i)))&rawHatBits[dir] != 0 {
			return 127
		}
	} else if n, ok := strings.CutPrefix(token, "AXIS"); ok && len(n) > 1 {
		i, err := strconv.Atoi(n[:len(n)-1])
		if err != nil {
			return 0
		}
		rest, _ := j.AxisInitialState(i)
		v, r := int32(j.Axis(i)), int32(rest)
		if n[len(n)-1] == '-' {
			v, r = -v, -r
		}
		if r < -16384 {
			// Rests at the opposite end: the whole range counts.
			return Clamp((v-r)*127/(32767-r), 0, 127)
		}
		if r > 16384 {
			return 0
		}
		return Clamp(v*127/32767, 0, 127)
	}
	return 0
}

// First raw joystick input pushed past half way, axes first, for Controller Mapping capture.
func rawHeld(joy int) string {
	j := input.joystick(joy)
	if j == nil {
		return ""
	}
	var tokens []string
	for i := 0; i < j.NumAxes(); i++ {
		tokens = append(tokens, fmt.Sprintf("AXIS%d-", i), fmt.Sprintf("AXIS%d+", i))
	}
	for i := 0; i < j.NumHats(); i++ {
		for _, dir := range []string{"UP", "DOWN", "LEFT", "RIGHT"} {
			tokens = append(tokens, fmt.Sprintf("HAT%d_%s", i, dir))
		}
	}
	for i := 0; i < j.NumButtons(); i++ {
		tokens = append(tokens, fmt.Sprintf("BUTTON%d", i))
	}
	for _, token := range tokens {
		if rawControlLevel(joy, token) > 64 {
			return token
		}
	}
	return ""
}

// Starting Controller Mapping of a raw joystick, in VirtualControl order: the common
// DirectInput layout. It is often wrong for a given pad, which is why the pad needs mapping.
var defaultRawControllerMapping = ControllerMapping{
	"HAT0_UP", "HAT0_DOWN", "HAT0_LEFT", "HAT0_RIGHT", // D-pad
	"BUTTON0", "BUTTON1", "BUTTON2", "BUTTON3", // A, B, X, Y
	"BUTTON9", "BUTTON8", // Start, Back
	"BUTTON4", "BUTTON5", "BUTTON6", "BUTTON7", // LB, RB, LT, RT
	"AXIS1-", "AXIS1+", "AXIS0-", "AXIS0+", // left stick up, down, left, right
	"AXIS3-", "AXIS3+", "AXIS2-", "AXIS2+", // right stick up, down, left, right
	"BUTTON10", "BUTTON11", // LS, RS
}
