package main

import (
	"encoding/binary"
	"encoding/hex"
	"math"
	"os"
	"runtime"
	"strings"

	"github.com/veandco/go-sdl2/sdl"
)

type ControllerState struct {
	Axes      [6]int8
	Buttons   map[sdl.GameControllerButton]byte
	HasRumble bool
}

type Input struct {
	controllers     [MaxPlayerNo]*sdl.GameController
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
var ButtonToStringLUT map[int]string
var buttonOrder []sdl.GameControllerButton
var StringToButtonLUT map[string]int

var input Input

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

	buttonOrder = []sdl.GameControllerButton{
		sdl.CONTROLLER_BUTTON_A,
		sdl.CONTROLLER_BUTTON_B,
		sdl.CONTROLLER_BUTTON_X,
		sdl.CONTROLLER_BUTTON_Y,
		sdl.CONTROLLER_BUTTON_BACK,
		sdl.CONTROLLER_BUTTON_GUIDE,
		sdl.CONTROLLER_BUTTON_START,
		sdl.CONTROLLER_BUTTON_LEFTSTICK,
		sdl.CONTROLLER_BUTTON_RIGHTSTICK,
		sdl.CONTROLLER_BUTTON_LEFTSHOULDER,
		sdl.CONTROLLER_BUTTON_RIGHTSHOULDER,
		sdl.CONTROLLER_BUTTON_DPAD_UP,
		sdl.CONTROLLER_BUTTON_DPAD_DOWN,
		sdl.CONTROLLER_BUTTON_DPAD_LEFT,
		sdl.CONTROLLER_BUTTON_DPAD_RIGHT,
	}

	ButtonToStringLUT = map[int]string{
		int(sdl.CONTROLLER_BUTTON_A):             "A",
		int(sdl.CONTROLLER_BUTTON_B):             "B",
		int(sdl.CONTROLLER_BUTTON_X):             "X",
		int(sdl.CONTROLLER_BUTTON_Y):             "Y",
		int(sdl.CONTROLLER_BUTTON_BACK):          "BACK",
		int(sdl.CONTROLLER_BUTTON_GUIDE):         "HOME",
		int(sdl.CONTROLLER_BUTTON_START):         "START",
		int(sdl.CONTROLLER_BUTTON_LEFTSTICK):     "LS",
		int(sdl.CONTROLLER_BUTTON_RIGHTSTICK):    "RS",
		int(sdl.CONTROLLER_BUTTON_LEFTSHOULDER):  "LB",
		int(sdl.CONTROLLER_BUTTON_RIGHTSHOULDER): "RB",
		int(sdl.CONTROLLER_BUTTON_DPAD_UP):       "DP_U",
		int(sdl.CONTROLLER_BUTTON_DPAD_DOWN):     "DP_D",
		int(sdl.CONTROLLER_BUTTON_DPAD_LEFT):     "DP_L",
		int(sdl.CONTROLLER_BUTTON_DPAD_RIGHT):    "DP_R",
		15:                                       "LS_Y-",
		16:                                       "LS_X-",
		17:                                       "LS_X+",
		18:                                       "LS_Y+",
		19:                                       "LT",
		20:                                       "RT",
		21:                                       "RS_Y-",
		22:                                       "RS_X-",
		23:                                       "RS_X+",
		24:                                       "RS_Y+",
		25:                                       "Not used",
	}

	// Explicitly allocate the maps here
	StringToKeyLUT = make(map[string]sdl.Keycode)
	StringToButtonLUT = make(map[string]int)

	for k, v := range KeyToStringLUT {
		StringToKeyLUT[v] = k
	}
	for k, v := range ButtonToStringLUT {
		StringToButtonLUT[v] = k
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
		sys.errLog.Printf("%v", err)
		return
	}
	lines := strings.Split(string(b), "\n")
	for _, line := range lines {
		sdl.GameControllerAddMapping(line)
	}
}

func (input *Input) GetMaxJoystickCount() int {
	return len(input.controllers)
}

func (input *Input) IsJoystickPresent(joy int) bool {
	if joy < 0 || joy >= len(input.controllers) {
		return false
	}
	if input.controllers[joy] == nil {
		return false
	}
	return input.controllers[joy].Attached()
}

func (input *Input) GetJoystickName(joy int) string {
	if joy < 0 || joy >= len(input.controllers) {
		return ""
	}
	return input.controllers[joy].Name()
}

func (input *Input) GetJoystickAxes(joy int) [6]float32 {
	if joy < 0 || joy >= len(input.controllerstate) {
		return [6]float32{0, 0, 0, 0, 0, 0}
	}
	axes := NormalizeAxes(&input.controllerstate[joy].Axes)
	return axes
}

func (input *Input) GetJoystickButtons(joy int) []byte {
	if joy < 0 || joy >= len(input.controllerstate) {
		return []byte{}
	}
	buttons := make([]byte, len(buttonOrder))
	for i, button := range buttonOrder {
		buttons[i] = input.controllerstate[joy].Buttons[button]
	}
	return buttons
}

func (input *Input) GetJoystickPath(joy int) string {
	if joy < 0 || joy >= len(input.controllers) {
		return ""
	}
	return input.controllers[joy].Path()
}

func (input *Input) GetJoystickGUID(joy int) string {
	if joy < 0 || joy >= len(input.controllers) {
		return ""
	}
	pid := uint16(input.controllers[joy].Product())
	pv := uint16(input.controllers[joy].ProductVersion())
	vid := uint16(input.controllers[joy].Vendor())
	guid := make([]byte, 16)

	guid[0] = 0x03
	binary.LittleEndian.PutUint16(guid[4:6], vid)
	binary.LittleEndian.PutUint16(guid[8:10], pid)
	binary.LittleEndian.PutUint16(guid[12:14], pv)
	s := hex.EncodeToString(guid[:])
	return s
}

func (input *Input) RumbleController(joy int, lo, hi uint16, ticks uint32) {
	if joy < 0 || joy >= len(input.controllers) || joy >= len(sys.joystickConfig) {
		return
	}

	// Only if Rumble Enabled for this config
	if input.controllerstate[joy].HasRumble && sys.joystickConfig[joy].rumbleOn {
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
				input.controllers[joy].Rumble(uint16(float32(lo)*multiplier), uint16(float32(hi)*multiplier), (ticks*framerate_ms)+buffertime_ms)
			} else {
				input.controllers[joy].Rumble(0, 0, 0)
			}
		}
	}
}

func CheckAxisForDpad(axes *[6]float32, base int) string {
	var s string = ""

	// Left stick
	if (*axes)[0] > sys.cfg.Input.ControllerStickSensitivity { // right (LS)
		s = ButtonToStringLUT[2+base]
	} else if -(*axes)[0] > sys.cfg.Input.ControllerStickSensitivity { // left (LS)
		s = ButtonToStringLUT[1+base]
	}
	// there are no OOB errors in SDL2 GameController configuration as
	// everything is in relation to XInput controls
	if (*axes)[1] > sys.cfg.Input.ControllerStickSensitivity { // down (LS)
		s = ButtonToStringLUT[3+base]
	} else if -(*axes)[1] > sys.cfg.Input.ControllerStickSensitivity { // up (LS)
		s = ButtonToStringLUT[base]
	}

	// Right stick
	if (*axes)[2] > sys.cfg.Input.ControllerStickSensitivity { // right (RS)
		s = ButtonToStringLUT[8+base]
	} else if -(*axes)[2] > sys.cfg.Input.ControllerStickSensitivity { // left (RS)
		s = ButtonToStringLUT[7+base]
	}
	if (*axes)[3] > sys.cfg.Input.ControllerStickSensitivity { // down (RS)
		s = ButtonToStringLUT[9+base]
	} else if -(*axes)[3] > sys.cfg.Input.ControllerStickSensitivity { // up (RS)
		s = ButtonToStringLUT[6+base]
	}

	return s
}

func CheckAxisForTrigger(axes *[6]float32) string {
	var s string = ""
	axesList := [2]int{int(sdl.CONTROLLER_AXIS_TRIGGERLEFT), int(sdl.CONTROLLER_AXIS_TRIGGERRIGHT)}
	for _, i := range axesList {
		// No need for "stuck axis" behavior anymore
		if (*axes)[i] > 0 {
			s = ButtonToStringLUT[15+i]
			break
		}
	}
	return s
}

// returns the first active button/axis token string and the joystick index.
func getJoystickKey(controllerIdx int) (string, int) {
	var s string
	min, max := 0, input.GetMaxJoystickCount()

	if controllerIdx >= 0 && controllerIdx < max {
		min, max = controllerIdx, controllerIdx+1
	}

	for joy := min; joy < max; joy++ {
		if !input.IsJoystickPresent(joy) {
			continue
		}

		axes := input.GetJoystickAxes(joy)
		btns := input.GetJoystickButtons(joy)

		// Prefer stick directions (LS_*/RS_* / DP_*).
		s = CheckAxisForDpad(&axes, len(btns))
		if s != "" {
			return s, joy
		}

		// Then triggers (LT / RT).
		s = CheckAxisForTrigger(&axes)
		if s != "" {
			return s, joy
		}

		// Finally, digital buttons (A/B/X/Y/etc.).
		for i := range btns {
			if btns[i] > 0 {
				s = ButtonToStringLUT[i]
			}
		}
		if s != "" && strings.ToLower(s) != "not used" {
			return s, joy
		}
	}

	return "", -1
}
