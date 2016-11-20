package main

import "github.com/go-gl/glfw/v3.2/glfw"

type CommandKey int32

const (
	CK_B CommandKey = iota
	CK_D
	CK_F
	CK_U
	CK_DB
	CK_UB
	CK_DF
	CK_UF
	CK_nB
	CK_nD
	CK_nF
	CK_nU
	CK_nDB
	CK_nUB
	CK_nDF
	CK_nUF
	CK_Bs
	CK_Ds
	CK_Fs
	CK_Us
	CK_DBs
	CK_UBs
	CK_DFs
	CK_UFs
	CK_nBs
	CK_nDs
	CK_nFs
	CK_nUs
	CK_nDBs
	CK_nUBs
	CK_nDFs
	CK_nUFs
	CK_a
	CK_b
	CK_c
	CK_x
	CK_y
	CK_z
	CK_s
	CK_na
	CK_nb
	CK_nc
	CK_nx
	CK_ny
	CK_nz
	CK_ns
)

var keySatate = make(map[glfw.Key]bool)

func keyCallback(_ *glfw.Window, key glfw.Key, _ int,
	action glfw.Action, _ glfw.ModifierKey) {
	switch action {
	case glfw.Release:
		keySatate[key] = false
	case glfw.Press:
		keySatate[key] = true
	}
}

var joystick = [...]glfw.Joystick{glfw.Joystick1, glfw.Joystick2,
	glfw.Joystick3, glfw.Joystick4, glfw.Joystick5, glfw.Joystick6,
	glfw.Joystick7, glfw.Joystick8, glfw.Joystick9, glfw.Joystick10,
	glfw.Joystick11, glfw.Joystick12, glfw.Joystick13, glfw.Joystick14,
	glfw.Joystick15, glfw.Joystick16}

func JoystickState(joy int32, button int32) bool {
	if joy < 0 {
		return keySatate[glfw.Key(button)]
	}
	if int(joy) >= len(joystick) {
		return false
	}
	if button < 0 {
		button = -button - 1
		axes := glfw.GetJoystickAxes(joystick[joy])
		if len(axes)*2 <= int(button) {
			return false
		}
		switch button & 1 {
		case 0:
			return axes[button/2] < -0.1
		case 1:
			return axes[button/2] > 0.1
		}
	}
	btns := glfw.GetJoystickButtons(joystick[joy])
	if len(btns) <= int(button) {
		return false
	}
	return btns[button] != 0
}

type commandBuffer struct {
	Bb, Db, Fb, Ub             int32
	ab, bb, cb, xb, yb, zb, sb int32
	B, D, F, U                 int8
	a, b, c, x, y, z, s        int8
}

func newCommandBuffer() *commandBuffer {
	return &commandBuffer{B: -1, D: -1, F: -1, U: -1,
		a: -1, b: -1, c: -1, x: -1, y: -1, z: -1, s: -1}
}
func (__ *commandBuffer) Input(B, D, F, U, a, b, c, x, y, z, s bool) {
	if (B && !F) != (__.B > 0) {
		__.Bb = 0
		__.B *= -1
	}
	__.Bb += int32(__.B)
	if (D && !U) != (__.D > 0) {
		__.Db = 0
		__.D *= -1
	}
	__.Db += int32(__.D)
	if (F && !B) != (__.F > 0) {
		__.Fb = 0
		__.F *= -1
	}
	__.Fb += int32(__.F)
	if (U && !D) != (__.U > 0) {
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
}
func (__ *commandBuffer) State(ck CommandKey) int32 {
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
		return Min(-Max(__.Ub, __.Fb), Max(__.Db, __.Bb))
	case CK_UBs:
		return Min(-Max(__.Db, __.Fb), Max(__.Ub, __.Bb))
	case CK_DFs:
		return Min(-Max(__.Ub, __.Bb), Max(__.Db, __.Fb))
	case CK_UFs:
		return Min(-Max(__.Db, __.Bb), Max(__.Ub, __.Fb))
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
	case CK_nB:
		return -Min(-Max(__.Db, __.Ub), __.Bb)
	case CK_nD:
		return -Min(-Max(__.Bb, __.Fb), __.Db)
	case CK_nF:
		return -Min(-Max(__.Db, __.Ub), __.Fb)
	case CK_nU:
		return -Min(-Max(__.Bb, __.Fb), __.Ub)
	case CK_nDB:
		return -Min(__.Db, __.Bb)
	case CK_nUB:
		return -Min(__.Ub, __.Bb)
	case CK_nDF:
		return -Min(__.Db, __.Fb)
	case CK_nUF:
		return -Min(__.Ub, __.Fb)
	case CK_nBs:
		return -__.Bb
	case CK_nDs:
		return -__.Db
	case CK_nFs:
		return -__.Fb
	case CK_nUs:
		return -__.Ub
	case CK_nDBs:
		return -Min(-Max(__.Ub, __.Fb), Max(__.Db, __.Bb))
	case CK_nUBs:
		return -Min(-Max(__.Db, __.Fb), Max(__.Ub, __.Bb))
	case CK_nDFs:
		return -Min(-Max(__.Ub, __.Bb), Max(__.Db, __.Fb))
	case CK_nUFs:
		return -Min(-Max(__.Db, __.Bb), Max(__.Ub, __.Fb))
	case CK_na:
		return -__.ab
	case CK_nb:
		return -__.bb
	case CK_nc:
		return -__.cb
	case CK_nx:
		return -__.xb
	case CK_ny:
		return -__.yb
	case CK_nz:
		return -__.zb
	case CK_ns:
		return -__.sb
	}
	return 0
}
func (__ *commandBuffer) State2(ck CommandKey) int32 {
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
	case CK_DBs:
		if s := __.State(CK_DBs); s < 0 {
			return s
		}
		return Min(Abs(__.Db), Abs(__.Bb))
	case CK_UBs:
		if s := __.State(CK_UBs); s < 0 {
			return s
		}
		return Min(Abs(__.Ub), Abs(__.Bb))
	case CK_DFs:
		if s := __.State(CK_DFs); s < 0 {
			return s
		}
		return Min(Abs(__.Db), Abs(__.Fb))
	case CK_UFs:
		if s := __.State(CK_UFs); s < 0 {
			return s
		}
		return Min(Abs(__.Ub), Abs(__.Fb))
	case CK_nBs:
		return f(__.State(CK_B), __.State(CK_UB), __.State(CK_DB))
	case CK_nDs:
		return f(__.State(CK_D), __.State(CK_DB), __.State(CK_DF))
	case CK_nFs:
		return f(__.State(CK_F), __.State(CK_DF), __.State(CK_UF))
	case CK_nUs:
		return f(__.State(CK_U), __.State(CK_UB), __.State(CK_UF))
	case CK_nDBs:
		return f(__.State(CK_DB), __.State(CK_D), __.State(CK_B))
	case CK_nUBs:
		return f(__.State(CK_UB), __.State(CK_U), __.State(CK_B))
	case CK_nDFs:
		return f(__.State(CK_DF), __.State(CK_D), __.State(CK_F))
	case CK_nUFs:
		return f(__.State(CK_UF), __.State(CK_U), __.State(CK_F))
	}
	return __.State(ck)
}
func (__ *commandBuffer) LastDirectionTime() int32 {
	return Min(Abs(__.Bb), Abs(__.Db), Abs(__.Fb), Abs(__.Ub))
}
func (__ *commandBuffer) LastChangeTime() int32 {
	return Min(__.LastDirectionTime(), Abs(__.ab), Abs(__.bb), Abs(__.cb),
		Abs(__.xb), Abs(__.yb), Abs(__.zb), Abs(__.sb))
}
