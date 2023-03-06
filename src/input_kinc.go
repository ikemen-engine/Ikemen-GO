//go:build kinc

package main

/*
#include <kinc/input/gamepad.h>
#include <kinc/input/keyboard.h>

typedef void (*button_callback_t)(int gamepad, int button, float value);
typedef void (*axis_callback_t)(int gamepad, int axis, float value);
typedef void (*key_callback_t)(int key);
typedef void (*char_callback_t)(unsigned ch);

#if _WIN32
extern __declspec(dllexport) void axis_callback(int gamepad, int axis, float value);
extern __declspec(dllexport) void button_callback(int gamepad, int button, float value);
extern __declspec(dllexport) void key_down_callback(int key);
extern __declspec(dllexport) void key_up_callback(int key);
extern __declspec(dllexport) void char_callback(unsigned ch);
#else
extern void axis_callback(int gamepad, int axis, float value);
extern void button_callback(int gamepad, int button, float value);
extern void key_down_callback(int key);
extern void key_up_callback(int key);
extern void char_callback(unsigned ch);
#endif
*/
import "C"

const (
	MAX_JOYSTICK_COUNT = 8
	MAX_BUTTON_COUNT   = 16
	MAX_AXIS_COUNT     = 8
)

type Joystick struct {
	buttons [MAX_BUTTON_COUNT]int32
	axes    [MAX_AXIS_COUNT]float32
}

type Input struct {
	joysticks [MAX_JOYSTICK_COUNT]Joystick
}

type Key C.int
type ModifierKey C.int

const (
	KeyUnknown = C.KINC_KEY_UNKNOWN
	KeyEscape  = C.KINC_KEY_ESCAPE
	KeyEnter   = C.KINC_KEY_RETURN
	KeyInsert  = C.KINC_KEY_INSERT
	KeyF12     = C.KINC_KEY_F12
)

var KeyToStringLUT = map[Key]string{
	C.KINC_KEY_RETURN:        "RETURN",
	C.KINC_KEY_ESCAPE:        "ESCAPE",
	C.KINC_KEY_BACKSPACE:     "BACKSPACE",
	C.KINC_KEY_TAB:           "TAB",
	C.KINC_KEY_SPACE:         "SPACE",
	C.KINC_KEY_QUOTE:         "QUOTE",
	C.KINC_KEY_COMMA:         "COMMA",
	C.KINC_KEY_HYPHEN_MINUS:  "MINUS",
	C.KINC_KEY_PERIOD:        "PERIOD",
	C.KINC_KEY_SLASH:         "SLASH",
	C.KINC_KEY_0:             "0",
	C.KINC_KEY_1:             "1",
	C.KINC_KEY_2:             "2",
	C.KINC_KEY_3:             "3",
	C.KINC_KEY_4:             "4",
	C.KINC_KEY_5:             "5",
	C.KINC_KEY_6:             "6",
	C.KINC_KEY_7:             "7",
	C.KINC_KEY_8:             "8",
	C.KINC_KEY_9:             "9",
	C.KINC_KEY_SEMICOLON:     "SEMICOLON",
	C.KINC_KEY_EQUALS:        "EQUALS",
	C.KINC_KEY_OPEN_BRACKET:  "LBRACKET",
	C.KINC_KEY_BACK_SLASH:    "BACKSLASH",
	C.KINC_KEY_CLOSE_BRACKET: "RBRACKET",
	C.KINC_KEY_BACK_QUOTE:    "BACKQUOTE",
	C.KINC_KEY_A:             "a",
	C.KINC_KEY_B:             "b",
	C.KINC_KEY_C:             "c",
	C.KINC_KEY_D:             "d",
	C.KINC_KEY_E:             "e",
	C.KINC_KEY_F:             "f",
	C.KINC_KEY_G:             "g",
	C.KINC_KEY_H:             "h",
	C.KINC_KEY_I:             "i",
	C.KINC_KEY_J:             "j",
	C.KINC_KEY_K:             "k",
	C.KINC_KEY_L:             "l",
	C.KINC_KEY_M:             "m",
	C.KINC_KEY_N:             "n",
	C.KINC_KEY_O:             "o",
	C.KINC_KEY_P:             "p",
	C.KINC_KEY_Q:             "q",
	C.KINC_KEY_R:             "r",
	C.KINC_KEY_S:             "s",
	C.KINC_KEY_T:             "t",
	C.KINC_KEY_U:             "u",
	C.KINC_KEY_V:             "v",
	C.KINC_KEY_W:             "w",
	C.KINC_KEY_X:             "x",
	C.KINC_KEY_Y:             "y",
	C.KINC_KEY_Z:             "z",
	C.KINC_KEY_CAPS_LOCK:     "CAPSLOCK",
	C.KINC_KEY_F1:            "F1",
	C.KINC_KEY_F2:            "F2",
	C.KINC_KEY_F3:            "F3",
	C.KINC_KEY_F4:            "F4",
	C.KINC_KEY_F5:            "F5",
	C.KINC_KEY_F6:            "F6",
	C.KINC_KEY_F7:            "F7",
	C.KINC_KEY_F8:            "F8",
	C.KINC_KEY_F9:            "F9",
	C.KINC_KEY_F10:           "F10",
	C.KINC_KEY_F11:           "F11",
	C.KINC_KEY_F12:           "F12",
	C.KINC_KEY_PRINT_SCREEN:  "PRINTSCREEN",
	C.KINC_KEY_SCROLL_LOCK:   "SCROLLLOCK",
	C.KINC_KEY_PAUSE:         "PAUSE",
	C.KINC_KEY_INSERT:        "INSERT",
	C.KINC_KEY_HOME:          "HOME",
	C.KINC_KEY_PAGE_UP:       "PAGEUP",
	C.KINC_KEY_DELETE:        "DELETE",
	C.KINC_KEY_END:           "END",
	C.KINC_KEY_PAGE_DOWN:     "PAGEDOWN",
	C.KINC_KEY_RIGHT:         "RIGHT",
	C.KINC_KEY_LEFT:          "LEFT",
	C.KINC_KEY_DOWN:          "DOWN",
	C.KINC_KEY_UP:            "UP",
	C.KINC_KEY_NUM_LOCK:      "NUMLOCKCLEAR",
	C.KINC_KEY_DIVIDE:        "KP_DIVIDE",
	C.KINC_KEY_MULTIPLY:      "KP_MULTIPLY",
	C.KINC_KEY_SUBTRACT:      "KP_MINUS",
	C.KINC_KEY_ADD:           "KP_PLUS",
	//C.KINC_KEY_NUMPAD_ENTER: "KP_ENTER",
	C.KINC_KEY_NUMPAD_1: "KP_1",
	C.KINC_KEY_NUMPAD_2: "KP_2",
	C.KINC_KEY_NUMPAD_3: "KP_3",
	C.KINC_KEY_NUMPAD_4: "KP_4",
	C.KINC_KEY_NUMPAD_5: "KP_5",
	C.KINC_KEY_NUMPAD_6: "KP_6",
	C.KINC_KEY_NUMPAD_7: "KP_7",
	C.KINC_KEY_NUMPAD_8: "KP_8",
	C.KINC_KEY_NUMPAD_9: "KP_9",
	C.KINC_KEY_NUMPAD_0: "KP_0",
	C.KINC_KEY_DECIMAL:  "KP_PERIOD",
	//C.KINC_KEY_NUMPAD_EQUAL: "KP_EQUALS",
	C.KINC_KEY_F13:          "F13",
	C.KINC_KEY_F14:          "F14",
	C.KINC_KEY_F15:          "F15",
	C.KINC_KEY_F16:          "F16",
	C.KINC_KEY_F17:          "F17",
	C.KINC_KEY_F18:          "F18",
	C.KINC_KEY_F19:          "F19",
	C.KINC_KEY_F20:          "F20",
	C.KINC_KEY_F21:          "F21",
	C.KINC_KEY_F22:          "F22",
	C.KINC_KEY_F23:          "F23",
	C.KINC_KEY_F24:          "F24",
	C.KINC_KEY_CONTEXT_MENU: "MENU",
	//C.KINC_KEY_LEFT_CONTROL: "LCTRL",
	//C.KINC_KEY_LEFT_SHIFT: "LSHIFT",
	//C.KINC_KEY_LEFT_ALT: "LALT",
	//C.KINC_KEY_LEFT_SUPER: "LGUI",
	//C.KINC_KEY_RIGHT_CONTROL: "RCTRL",
	//C.KINC_KEY_RIGHT_SHIFT: "RSHIFT",
	//C.KINC_KEY_RIGHT_ALT: "RALT",
	//C.KINC_KEY_RIGHT_SUPER: "RGUI",
}

var StringToKeyLUT = map[string]Key{}

func init() {
	for k, v := range KeyToStringLUT {
		StringToKeyLUT[v] = k
	}
}

func StringToKey(s string) Key {
	if key, ok := StringToKeyLUT[s]; ok {
		return key
	}
	return C.KINC_KEY_UNKNOWN
}

func KeyToString(k Key) string {
	if s, ok := KeyToStringLUT[k]; ok {
		return s
	}
	return ""
}

func NewModifierKey(ctrl, alt, shift bool) (mod ModifierKey) {
	// TODO: implement modifiers
	if ctrl || alt || shift {
		mod = 1
	}
	return
}

var input *Input = newInput()

//export button_callback
func button_callback(gamepad int32, button int32, value float32) {
	if gamepad >= 0 && gamepad < MAX_JOYSTICK_COUNT {
		if button >= 0 && button < MAX_BUTTON_COUNT {
			input.joysticks[gamepad].buttons[button] = int32(value)
		}
	}
}

//export axis_callback
func axis_callback(gamepad int32, axis int32, value float32) {
	if gamepad >= 0 && gamepad < MAX_JOYSTICK_COUNT {
		if axis >= 0 && axis < MAX_AXIS_COUNT {
			input.joysticks[gamepad].axes[axis] = value
		}
	}
}

//export key_down_callback
func key_down_callback(key int32) {
	OnKeyPressed(Key(key), 0)
}

//export key_up_callback
func key_up_callback(key int32) {
	OnKeyReleased(Key(key), 0)
}

//export char_callback
func char_callback(ch uint32) {
	OnTextEntered(string(rune(ch)))
}

func newInput() *Input {
	C.kinc_gamepad_set_axis_callback((C.axis_callback_t)(C.axis_callback))
	C.kinc_gamepad_set_button_callback((C.button_callback_t)(C.button_callback))
	C.kinc_keyboard_set_key_down_callback((C.key_callback_t)(C.key_down_callback))
	C.kinc_keyboard_set_key_up_callback((C.key_callback_t)(C.key_up_callback))
	C.kinc_keyboard_set_key_press_callback((C.char_callback_t)(C.char_callback))
	return &Input{}
}

func (input *Input) GetMaxJoystickCount() int {
	return MAX_JOYSTICK_COUNT
}

func (input *Input) IsJoystickPresent(joy int) bool {
	return bool(C.kinc_gamepad_connected(C.int(joy)))
}

func (input *Input) GetJoystickName(joy int) string {
	return C.GoString(C.kinc_gamepad_product_name(C.int(joy)))
}

func (input *Input) GetJoystickAxes(joy int) []float32 {
	if joy >= 0 && joy < MAX_JOYSTICK_COUNT {
		return input.joysticks[joy].axes[:]
	}
	return []float32{}
}

func (input *Input) GetJoystickButtons(joy int) []int32 {
	if joy >= 0 && joy < MAX_JOYSTICK_COUNT {
		return input.joysticks[joy].buttons[:]
	}
	return []int32{}
}
