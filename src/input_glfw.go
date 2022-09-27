//go:build !kinc

package main

import (
	glfw "github.com/fyne-io/glfw-js"
)

type Input struct {
	joystick []glfw.Joystick
}

var input = Input{
	joystick: []glfw.Joystick{glfw.Joystick1, glfw.Joystick2, glfw.Joystick3,
		glfw.Joystick4, glfw.Joystick5, glfw.Joystick6, glfw.Joystick7,
		glfw.Joystick8, glfw.Joystick9, glfw.Joystick10, glfw.Joystick11,
		glfw.Joystick12, glfw.Joystick13, glfw.Joystick14, glfw.Joystick15,
		glfw.Joystick16 },
}

func (input* Input) GetMaxJoystickCount() int {
	return len(input.joystick)
}

func (input* Input) IsJoystickPresent(joy int) bool {
	if joy < 0 || joy >= len(input.joystick) {
		return false
	}
	return input.joystick[joy].IsPresent()
}

func (input* Input) GetJoystickName(joy int) string {
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
