//go:build kinc

package main

/*
#include <kinc/input/gamepad.h>

typedef void (*button_callback_t)(int gamepad, int button, float value);
typedef void (*axis_callback_t)(int gamepad, int axis, float value);

#if _WIN32
extern __declspec(dllexport) void axis_callback(int gamepad, int axis, float value);
extern __declspec(dllexport) void button_callback(int gamepad, int button, float value);
#else
extern void axis_callback(int gamepad, int axis, float value);
extern void button_callback(int gamepad, int button, float value);
#endif
*/
import "C"

const (
	MAX_JOYSTICK_COUNT = 8
	MAX_BUTTON_COUNT = 16
	MAX_AXIS_COUNT = 8
)

type Joystick struct {
	buttons [MAX_BUTTON_COUNT]int32
	axes [MAX_AXIS_COUNT]float32
}

type Input struct {
	joysticks [MAX_JOYSTICK_COUNT]Joystick
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

func newInput() *Input {
	C.kinc_gamepad_set_axis_callback((C.axis_callback_t)(C.axis_callback))
	C.kinc_gamepad_set_button_callback((C.button_callback_t)(C.button_callback))
	return &Input{}
}

func (input* Input) GetMaxJoystickCount() int {
	return MAX_JOYSTICK_COUNT
}

func (input* Input) IsJoystickPresent(joy int) bool {
	return bool(C.kinc_gamepad_connected(C.int(joy)))
}

func (input* Input) GetJoystickName(joy int) string {
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
