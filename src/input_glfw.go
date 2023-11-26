//go:build !kinc

package main

import (
	glfw "github.com/go-gl/glfw/v3.3/glfw"
)

type Input struct {
	joystick []glfw.Joystick
}

type Key = glfw.Key
type ModifierKey = glfw.ModifierKey

const (
	KeyUnknown = glfw.KeyUnknown
	KeyEscape  = glfw.KeyEscape
	KeyEnter   = glfw.KeyEnter
	KeyInsert  = glfw.KeyInsert
	KeyF12     = glfw.KeyF12
)

var KeyToStringLUT = map[glfw.Key]string{
	glfw.KeyEnter:        "RETURN",
	glfw.KeyEscape:       "ESCAPE",
	glfw.KeyBackspace:    "BACKSPACE",
	glfw.KeyTab:          "TAB",
	glfw.KeySpace:        "SPACE",
	glfw.KeyApostrophe:   "QUOTE",
	glfw.KeyComma:        "COMMA",
	glfw.KeyMinus:        "MINUS",
	glfw.KeyPeriod:       "PERIOD",
	glfw.KeySlash:        "SLASH",
	glfw.Key0:            "0",
	glfw.Key1:            "1",
	glfw.Key2:            "2",
	glfw.Key3:            "3",
	glfw.Key4:            "4",
	glfw.Key5:            "5",
	glfw.Key6:            "6",
	glfw.Key7:            "7",
	glfw.Key8:            "8",
	glfw.Key9:            "9",
	glfw.KeySemicolon:    "SEMICOLON",
	glfw.KeyEqual:        "EQUALS",
	glfw.KeyLeftBracket:  "LBRACKET",
	glfw.KeyBackslash:    "BACKSLASH",
	glfw.KeyRightBracket: "RBRACKET",
	glfw.KeyGraveAccent:  "BACKQUOTE",
	glfw.KeyA:            "a",
	glfw.KeyB:            "b",
	glfw.KeyC:            "c",
	glfw.KeyD:            "d",
	glfw.KeyE:            "e",
	glfw.KeyF:            "f",
	glfw.KeyG:            "g",
	glfw.KeyH:            "h",
	glfw.KeyI:            "i",
	glfw.KeyJ:            "j",
	glfw.KeyK:            "k",
	glfw.KeyL:            "l",
	glfw.KeyM:            "m",
	glfw.KeyN:            "n",
	glfw.KeyO:            "o",
	glfw.KeyP:            "p",
	glfw.KeyQ:            "q",
	glfw.KeyR:            "r",
	glfw.KeyS:            "s",
	glfw.KeyT:            "t",
	glfw.KeyU:            "u",
	glfw.KeyV:            "v",
	glfw.KeyW:            "w",
	glfw.KeyX:            "x",
	glfw.KeyY:            "y",
	glfw.KeyZ:            "z",
	glfw.KeyCapsLock:     "CAPSLOCK",
	glfw.KeyF1:           "F1",
	glfw.KeyF2:           "F2",
	glfw.KeyF3:           "F3",
	glfw.KeyF4:           "F4",
	glfw.KeyF5:           "F5",
	glfw.KeyF6:           "F6",
	glfw.KeyF7:           "F7",
	glfw.KeyF8:           "F8",
	glfw.KeyF9:           "F9",
	glfw.KeyF10:          "F10",
	glfw.KeyF11:          "F11",
	glfw.KeyF12:          "F12",
	glfw.KeyPrintScreen:  "PRINTSCREEN",
	glfw.KeyScrollLock:   "SCROLLLOCK",
	glfw.KeyPause:        "PAUSE",
	glfw.KeyInsert:       "INSERT",
	glfw.KeyHome:         "HOME",
	glfw.KeyPageUp:       "PAGEUP",
	glfw.KeyDelete:       "DELETE",
	glfw.KeyEnd:          "END",
	glfw.KeyPageDown:     "PAGEDOWN",
	glfw.KeyRight:        "RIGHT",
	glfw.KeyLeft:         "LEFT",
	glfw.KeyDown:         "DOWN",
	glfw.KeyUp:           "UP",
	glfw.KeyNumLock:      "NUMLOCKCLEAR",
	glfw.KeyKPDivide:     "KP_DIVIDE",
	glfw.KeyKPMultiply:   "KP_MULTIPLY",
	glfw.KeyKPSubtract:   "KP_MINUS",
	glfw.KeyKPAdd:        "KP_PLUS",
	glfw.KeyKPEnter:      "KP_ENTER",
	glfw.KeyKP1:          "KP_1",
	glfw.KeyKP2:          "KP_2",
	glfw.KeyKP3:          "KP_3",
	glfw.KeyKP4:          "KP_4",
	glfw.KeyKP5:          "KP_5",
	glfw.KeyKP6:          "KP_6",
	glfw.KeyKP7:          "KP_7",
	glfw.KeyKP8:          "KP_8",
	glfw.KeyKP9:          "KP_9",
	glfw.KeyKP0:          "KP_0",
	glfw.KeyKPDecimal:    "KP_PERIOD",
	glfw.KeyKPEqual:      "KP_EQUALS",
	glfw.KeyF13:          "F13",
	glfw.KeyF14:          "F14",
	glfw.KeyF15:          "F15",
	glfw.KeyF16:          "F16",
	glfw.KeyF17:          "F17",
	glfw.KeyF18:          "F18",
	glfw.KeyF19:          "F19",
	glfw.KeyF20:          "F20",
	glfw.KeyF21:          "F21",
	glfw.KeyF22:          "F22",
	glfw.KeyF23:          "F23",
	glfw.KeyF24:          "F24",
	glfw.KeyMenu:         "MENU",
	glfw.KeyLeftControl:  "LCTRL",
	glfw.KeyLeftShift:    "LSHIFT",
	glfw.KeyLeftAlt:      "LALT",
	glfw.KeyLeftSuper:    "LGUI",
	glfw.KeyRightControl: "RCTRL",
	glfw.KeyRightShift:   "RSHIFT",
	glfw.KeyRightAlt:     "RALT",
	glfw.KeyRightSuper:   "RGUI",
}

var StringToKeyLUT = map[string]glfw.Key{}

func init() {
	for k, v := range KeyToStringLUT {
		StringToKeyLUT[v] = k
	}
}

func StringToKey(s string) glfw.Key {
	if key, ok := StringToKeyLUT[s]; ok {
		return key
	}
	return glfw.KeyUnknown
}

func KeyToString(k glfw.Key) string {
	if s, ok := KeyToStringLUT[k]; ok {
		return s
	}
	return ""
}

func NewModifierKey(ctrl, alt, shift bool) (mod glfw.ModifierKey) {
	if ctrl {
		mod |= glfw.ModControl
	}
	if alt {
		mod |= glfw.ModAlt
	}
	if shift {
		mod |= glfw.ModShift
	}
	return
}

var input = Input{
	joystick: []glfw.Joystick{glfw.Joystick1, glfw.Joystick2, glfw.Joystick3,
		glfw.Joystick4, glfw.Joystick5, glfw.Joystick6, glfw.Joystick7,
		glfw.Joystick8, glfw.Joystick9, glfw.Joystick10, glfw.Joystick11,
		glfw.Joystick12, glfw.Joystick13, glfw.Joystick14, glfw.Joystick15,
		glfw.Joystick16},
}

func (input *Input) GetMaxJoystickCount() int {
	return len(input.joystick)
}

func (input *Input) IsJoystickPresent(joy int) bool {
	if joy < 0 || joy >= len(input.joystick) {
		return false
	}
	return input.joystick[joy].Present()
}

func (input *Input) GetJoystickName(joy int) string {
	if joy < 0 || joy >= len(input.joystick) {
		return ""
	}
	return input.joystick[joy].GetGamepadName()
}

func (input *Input) GetJoystickAxes(joy int) []float32 {
	if joy < 0 || joy >= len(input.joystick) {
		return []float32{}
	}
	return input.joystick[joy].GetAxes()
}

func (input *Input) GetJoystickButtons(joy int) []glfw.Action {
	if joy < 0 || joy >= len(input.joystick) {
		return []glfw.Action{}
	}
	return input.joystick[joy].GetButtons()
}
