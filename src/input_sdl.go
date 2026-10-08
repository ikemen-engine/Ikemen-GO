package main

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/veandco/go-sdl2/sdl"
)

type ControllerState struct {
	Axes      [6]int8
	Buttons   map[sdl.GameControllerButton]byte
	HasRumble bool
}

// A controller mapping being edited. It stays a draft until it is saved, and it
// only describes the SDL layer: gameplay profiles and player ownership are
// separate.
type GamepadMapping struct {
	Instance   sdl.JoystickID
	Name       string
	Recognized bool
	Original   string
	Bindings   map[string]string
	neutral    []int32
	control    string
	pending    string
	armed      bool
}

const (
	// Raw axis range is -32768..32767. An axis has to move about a quarter of
	// the way to count as moved, and half way to be taken as the binding, so a
	// noisy stick cannot bind itself.
	gamepadMappingAxisMoved = 8000
	gamepadMappingAxisBind  = 16000
)

// The SDL controls the setup screen walks through, in order. The four sticks
// are captured one direction at a time, so split axes and digital sticks work.
var gamepadMappingControls = []string{
	"a", "b", "x", "y", "dpup", "dpdown", "dpleft", "dpright",
	"leftshoulder", "rightshoulder", "lefttrigger", "righttrigger",
	"back", "start", "guide", "leftstick", "rightstick",
	"leftxnegative", "leftxpositive", "leftynegative", "leftypositive",
	"rightxnegative", "rightxpositive", "rightynegative", "rightypositive", "misc1",
	"paddle1", "paddle2", "paddle3", "paddle4", "touchpad",
}

type Input struct {
	rawJoysticks    map[sdl.JoystickID]*sdl.Joystick
	mapping         *GamepadMapping
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

// Returns the controller type used to pick glyph artwork, for example "ps5".
// SDL knows the type of the common controllers after applying gamecontrollerdb
// mappings. The motif can match on GUID or controller name for the rest.
func (in *Input) GlyphController(src *InputSource, glyphs *ControllerGlyphsProperties) string {
	if src == nil || src.Joy < 0 || !in.IsJoystickPresent(src.Joy) {
		return "generic"
	}
	if controller := glyphs.GUID[strings.ToLower(src.GUID)]; controller != "" {
		return strings.ToLower(controller)
	}
	// Longest matching name wins, so "switch 2" beats "switch". Map order is
	// random, so equal-length patterns are decided alphabetically.
	name := strings.ToLower(in.GetJoystickName(src.Joy))
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
	switch in.controllers[src.Joy].Type() {
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
		25:                                       "-", // unbound, shown as menu.valuename.nokey
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
		LogMessage("%v", err)
		return
	}
	lines := strings.Split(string(b), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "#") && gamepadMappingPlatform(line) {
			sdl.GameControllerAddMapping(line)
		}
	}
}

// Reports whether a database line applies to this operating system. The
// database holds entries for every platform, and SDL adds one line at a time.
func gamepadMappingPlatform(line string) bool {
	fields := strings.Split(line, ",")
	for _, field := range fields[Min(2, len(fields)):] {
		if strings.HasPrefix(field, "platform:") {
			return strings.TrimPrefix(field, "platform:") == sdl.GetPlatform()
		}
	}
	return true
}

// Replaces the mapping database in one step, keeping its permissions.
func writeGamepadMappings(path string, contents []byte, mode os.FileMode) error {
	file, err := os.CreateTemp(filepath.Dir(path), ".gamecontrollerdb-*")
	if err != nil {
		return err
	}
	temporary := file.Name()
	defer os.Remove(temporary)
	if _, err = file.Write(contents); err == nil {
		err = file.Chmod(mode)
	}
	if err == nil {
		err = file.Sync()
	}
	closeErr := file.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	return os.Rename(temporary, path)
}

// Returns the raw joystick with a button held down, so a controller SDL does not
// recognise can still be detected. Instance IDs last until it is unplugged.
func (input *Input) GetJoystickRawId() int {
	ids := make([]int, 0, len(input.rawJoysticks))
	for id := range input.rawJoysticks {
		ids = append(ids, int(id))
	}
	sort.Ints(ids)
	for _, id := range ids {
		joy := input.rawJoysticks[sdl.JoystickID(id)]
		if !joy.Attached() {
			continue
		}
		for button := 0; button < joy.NumButtons(); button++ {
			if joy.Button(button) != 0 {
				return id
			}
		}
	}
	return -1
}

// Describes a raw joystick for the setup screen: its name, the gamepad slot it
// occupies once SDL recognises it, and -1 for an unrecognised one.
func (input *Input) GetJoystickRawInfo(instance int) (name string, slot int, ok bool) {
	id := sdl.JoystickID(instance)
	joy := input.rawJoysticks[id]
	if joy == nil || !joy.Attached() {
		return "", -1, false
	}
	slot = findControllerIndex(id)
	if slot >= 0 {
		return input.controllers[slot].Name(), slot, true
	}
	return joy.Name(), slot, true
}

// Starts a mapping draft for a raw joystick, replacing any draft in progress.
func (input *Input) BeginGamepadMapping(instance int) bool {
	m := &GamepadMapping{Instance: sdl.JoystickID(instance), Bindings: make(map[string]string)}
	joy := m.joystick()
	if joy == nil {
		return false
	}
	m.Name = joy.Name()
	if slot := findControllerIndex(m.Instance); slot >= 0 {
		m.Recognized = true
		m.Name = input.controllers[slot].Name()
		m.Original = input.controllers[slot].Mapping()
	}
	input.mapping = m
	m.calibrate()
	return true
}

// Current normalized controls for the setup screen's controller test. Reading
// them together allows several buttons, directions and triggers to light up.
func (m *GamepadMapping) state() map[string]bool {
	state := make(map[string]bool)
	if m == nil || m.joystick() == nil {
		return state
	}
	slot := findControllerIndex(m.Instance)
	if slot < 0 {
		return state
	}
	controller := input.controllers[slot]
	for _, name := range gamepadMappingControls {
		button := sdl.GameControllerGetButtonFromString(name)
		if button != sdl.CONTROLLER_BUTTON_INVALID {
			state[name] = controller.Button(button) != 0
			continue
		}
		axisName := strings.TrimSuffix(strings.TrimSuffix(name, "negative"), "positive")
		axis := sdl.GameControllerGetAxisFromString(axisName)
		if axis == sdl.CONTROLLER_AXIS_INVALID {
			continue
		}
		value := int32(controller.Axis(axis))
		if strings.HasSuffix(name, "negative") {
			state[name] = value < -gamepadMappingAxisMoved
		} else {
			state[name] = value > gamepadMappingAxisMoved
		}
	}
	return state
}

// Opens a joystick without waiting for SDL to recognise it, so a pad with no
// database entry can still be set up.
func (in *Input) openRawJoystick(index int) {
	if _, disabled := sys.cmdFlags["-nojoy"]; disabled {
		return
	}
	id := sdl.JoystickGetDeviceInstanceID(index)
	if id < 0 || in.rawJoysticks[id] != nil {
		return
	}
	if joy := sdl.JoystickOpen(index); joy != nil {
		if in.rawJoysticks == nil {
			in.rawJoysticks = make(map[sdl.JoystickID]*sdl.Joystick)
		}
		in.rawJoysticks[id] = joy
	}
}

// The joystick being mapped, or nil once it is unplugged.
func (m *GamepadMapping) joystick() *sdl.Joystick {
	if m != nil {
		if joy := input.rawJoysticks[m.Instance]; joy != nil && joy.Attached() {
			return joy
		}
	}
	return nil
}

// Records where the axes rest, so a resting trigger is not read as pressed.
func (m *GamepadMapping) calibrate() {
	if joy := m.joystick(); joy != nil {
		m.neutral = make([]int32, joy.NumAxes())
		for i := range m.neutral {
			m.neutral[i] = int32(joy.Axis(i))
		}
	}
	m.control, m.pending, m.armed = "", "", false
}

// Watches the raw buttons, hats and axes for the control being configured and
// returns its binding once the player releases everything again. Waiting for
// the release stops one press from filling in two controls.
func (m *GamepadMapping) capture(control string) string {
	joy := m.joystick()
	if joy == nil {
		return ""
	}
	valid := false
	for _, name := range gamepadMappingControls {
		valid = valid || control == name
	}
	if !valid {
		return ""
	}
	if control != m.control {
		m.control, m.pending, m.armed = control, "", false
	}
	stick := strings.HasSuffix(control, "negative") || strings.HasSuffix(control, "positive")
	trigger := control == "lefttrigger" || control == "righttrigger"
	neutral, candidate := true, ""
	for i := 0; i < joy.NumButtons(); i++ {
		if joy.Button(i) != 0 {
			neutral = false
			if candidate == "" {
				candidate = fmt.Sprintf("b%d", i)
			}
		}
	}
	for i := 0; i < joy.NumHats(); i++ {
		hat := joy.Hat(sdl.JoystickHat(i))
		if hat != 0 {
			neutral = false
			// Diagonals are ambiguous. Wait for a single cardinal direction.
			if candidate == "" && hat&(hat-1) == 0 {
				candidate = fmt.Sprintf("h%d.%d", i, hat)
			}
		}
	}
	var strongest int32
	for i, rest := range m.neutral {
		delta := int32(joy.Axis(i)) - rest
		magnitude := Abs(delta)
		if magnitude > gamepadMappingAxisMoved {
			neutral = false
		}
		if magnitude <= gamepadMappingAxisBind || magnitude <= strongest {
			continue
		}
		strongest = magnitude
		if candidate != "" && !stick && !trigger {
			continue
		}
		// An axis resting at an endpoint is a trigger, so bind its full range.
		if Abs(rest) > gamepadMappingAxisBind {
			candidate = fmt.Sprintf("a%d", i)
			if delta < 0 {
				candidate += "~"
			}
		} else if delta > 0 {
			candidate = fmt.Sprintf("+a%d", i)
		} else {
			candidate = fmt.Sprintf("-a%d", i)
		}
	}
	if neutral {
		m.armed = true
		if m.pending != "" {
			binding := m.pending
			m.Bindings[control] = binding
			m.pending = ""
			return binding
		}
	} else if m.armed && m.pending == "" && candidate != "" {
		m.pending = candidate
	}
	return ""
}

// Appends the finished mapping to the database, applies it right away, and
// restores the file if SDL rejects it.
func (m *GamepadMapping) save(name string) error {
	joy := m.joystick()
	if joy == nil {
		return fmt.Errorf("controller disconnected")
	}
	if m.Recognized {
		name = m.Name
	}
	name = strings.TrimSpace(name)
	if name == "" || strings.ContainsAny(name, ",\r\n") {
		return fmt.Errorf("enter a controller name without commas or line breaks")
	}
	if len(m.Bindings) == 0 {
		return fmt.Errorf("assign at least one control before saving")
	}
	// SDL handles the GUID checksum itself. Keep the physical GUID intact,
	// as entries produced by external mapping tools do.
	fields := []string{sdl.JoystickGetGUIDString(joy.GUID()), name}
	for _, control := range gamepadMappingControls {
		binding := m.Bindings[control]
		if binding == "" {
			continue
		}
		output := control
		if strings.HasSuffix(control, "negative") {
			output = "-" + strings.TrimSuffix(control, "negative")
		} else if strings.HasSuffix(control, "positive") {
			output = "+" + strings.TrimSuffix(control, "positive")
		}
		fields = append(fields, output+":"+binding)
	}
	// Retain SDL's controller type on an override.
	for _, field := range strings.Split(m.Original, ",") {
		if strings.HasPrefix(field, "type:") {
			fields = append(fields, field)
		}
	}
	fields = append(fields, "platform:"+sdl.GetPlatform(), "")
	line := strings.Join(fields, ",")
	path := sys.cfg.Config.GamepadMappings
	if path == "" {
		return fmt.Errorf("Config.GamepadMappings has no file path")
	}
	old, err := os.ReadFile(path)
	existed := err == nil
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	mode := os.FileMode(0644)
	if info, err := os.Stat(path); err == nil {
		mode = info.Mode().Perm()
	}
	contents := append([]byte(nil), old...)
	if len(contents) > 0 && contents[len(contents)-1] != '\n' {
		contents = append(contents, '\n')
	}
	contents = append(contents, []byte(line+"\n")...)
	if err := writeGamepadMappings(path, contents, mode); err != nil {
		return err
	}
	var mappingErr error
	if sdl.GameControllerAddMapping(line) < 0 {
		mappingErr = sdl.GetError()
	}
	if mappingErr != nil {
		var restoreErr error
		if existed {
			restoreErr = writeGamepadMappings(path, old, mode)
		} else {
			restoreErr = os.Remove(path)
		}
		if restoreErr != nil {
			return fmt.Errorf("SDL: %v; unable to restore mapping file: %v", mappingErr, restoreErr)
		}
		return fmt.Errorf("SDL: %v", mappingErr)
	}
	// A previously unknown pad is now usable. Existing sources keep their IDs,
	// profiles and participant ownership; also update identical connected pads.
	for i := 0; i < sdl.NumJoysticks(); i++ {
		id := sdl.JoystickGetDeviceInstanceID(i)
		if slot := findControllerIndex(id); slot >= 0 {
			controller := input.controllers[slot]
			if controller.Joystick().GUID() != joy.GUID() {
				continue
			}
			// A new hardware-specific entry may replace an automatic SDL mapping.
			// Reopen it in the same slot so the current instance uses the new entry.
			controller.Close()
			input.controllers[slot] = sdl.GameControllerOpen(i)
			if input.controllers[slot] == nil {
				sys.detachGamepad(slot)
				return fmt.Errorf("mapping saved; reconnect controller: %v", sdl.GetError())
			}
		} else {
			attachController(i)
		}
	}
	for i, controller := range input.controllers {
		if controller == nil {
			continue
		}
		resetControllerState(i)
		state := input.controllerstate[i]
		state.HasRumble = controller.HasRumble()
		for axis := range state.Axes {
			state.Axes[axis] = convertI16toI8(controller.Axis(sdl.GameControllerAxis(axis)))
		}
		for _, button := range buttonOrder {
			state.Buttons[button] = controller.Button(button)
		}
		for _, source := range sys.inputs.Sources {
			if source.Connected && source.Joy == i {
				source.Name = controller.Name()
			}
		}
	}
	sys.clearCommandBuffers()
	return nil
}

func (input *Input) GetMaxJoystickCount() int {
	return len(input.controllers)
}

func (input *Input) IsJoystickPresent(joy int) bool {
	if joy < 0 || joy >= len(input.controllers) || input.controllers[joy] == nil {
		return false
	}
	if input.controllers[joy] == nil {
		return false
	}
	return input.controllers[joy].Attached()
}

func (input *Input) GetJoystickName(joy int) string {
	if joy < 0 || joy >= len(input.controllers) || input.controllers[joy] == nil {
		return ""
	}
	return input.controllers[joy].Name()
}

func (input *Input) GetJoystickAxes(joy int) [6]float32 {
	if joy < 0 || joy >= len(input.controllerstate) || input.controllerstate[joy] == nil {
		return [6]float32{0, 0, 0, 0, 0, 0}
	}
	axes := NormalizeAxes(&input.controllerstate[joy].Axes)
	return axes
}

func (input *Input) GetJoystickButtons(joy int) []byte {
	if joy < 0 || joy >= len(input.controllerstate) || input.controllerstate[joy] == nil {
		return []byte{}
	}
	buttons := make([]byte, len(buttonOrder))
	for i, button := range buttonOrder {
		buttons[i] = input.controllerstate[joy].Buttons[button]
	}
	return buttons
}

// Reports which menu controls a gamepad is pressing right now, as button
// indexes. Menus accept the d-pad, the left stick, Start and Back whatever the
// profile binds them to, so these are read straight from the controller.
func (input *Input) GetJoystickMenuState(joy int) [14]bool {
	var state [14]bool
	if joy < 0 || joy >= len(input.controllerstate) || input.controllerstate[joy] == nil {
		return state
	}
	axes := NormalizeAxes(&input.controllerstate[joy].Axes)
	buttons := input.controllerstate[joy].Buttons
	deadzone := sys.cfg.Input.ControllerStickSensitivity
	stickX, stickY := axes[sdl.CONTROLLER_AXIS_LEFTX], axes[sdl.CONTROLLER_AXIS_LEFTY]
	state[BI_PU] = buttons[sdl.CONTROLLER_BUTTON_DPAD_UP] != 0 || stickY < -deadzone
	state[BI_PD] = buttons[sdl.CONTROLLER_BUTTON_DPAD_DOWN] != 0 || stickY > deadzone
	state[BI_PL] = buttons[sdl.CONTROLLER_BUTTON_DPAD_LEFT] != 0 || stickX < -deadzone
	state[BI_PR] = buttons[sdl.CONTROLLER_BUTTON_DPAD_RIGHT] != 0 || stickX > deadzone
	state[BI_S] = buttons[sdl.CONTROLLER_BUTTON_START] != 0
	state[BI_M] = buttons[sdl.CONTROLLER_BUTTON_BACK] != 0
	return state
}

func (input *Input) GetJoystickPath(joy int) string {
	if joy < 0 || joy >= len(input.controllers) || input.controllers[joy] == nil {
		return ""
	}
	return input.controllers[joy].Path()
}

func (input *Input) GetJoystickGUID(joy int) string {
	if joy < 0 || joy >= len(input.controllers) || input.controllers[joy] == nil {
		return ""
	}
	return sdl.JoystickGetGUIDString(input.controllers[joy].Joystick().GUID())
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
	for _, id := range sys.inputs.participantSources(participant) {
		if kc, found := sys.inputSourceConfig(sys.inputs.inputSource(id)); found {
			input.rumbleConfig(kc, lo, hi, ticks)
		}
	}
}

func (input *Input) rumbleConfig(kc KeyConfig, lo, hi uint16, ticks uint32) {
	joy := kc.Joy
	if !input.IsJoystickPresent(joy) || input.controllerstate[joy] == nil {
		return
	}

	// Only if Rumble Enabled for this config
	if input.controllerstate[joy].HasRumble && kc.rumbleOn {
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
