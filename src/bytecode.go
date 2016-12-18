package main

import (
	"math"
	"unsafe"
)

type StateType int32

const (
	ST_S StateType = 1 << iota
	ST_C
	ST_A
	ST_L
	ST_N
	ST_U
	ST_D = ST_L
	ST_F = ST_N
	ST_P = ST_U
)

type AttackType int32

const (
	AT_NA AttackType = 1 << (iota + 6)
	AT_NT
	AT_NP
	AT_SA
	AT_ST
	AT_SP
	AT_HA
	AT_HT
	AT_HP
)

type MoveType int32

const (
	MT_I MoveType = 1 << (iota + 15)
	MT_H
	MT_A   = MT_I + 1
	MT_U   = MT_H + 1
	MT_MNS = MT_I
	MT_PLS = MT_H
)

type ValueType int

const (
	VT_Float ValueType = iota
	VT_Int
	VT_Bool
	VT_SFalse
)

type OpCode byte

const (
	OC_var OpCode = iota + 110
	OC_sysvar
	OC_fvar
	OC_sysfvar
	OC_int8
	OC_int
	OC_float
	OC_pop
	OC_dup
	OC_swap
	OC_run
	OC_jmp8
	OC_jz8
	OC_jnz8
	OC_jmp
	OC_jz
	OC_jnz
	OC_eq
	OC_ne
	OC_gt
	OC_le
	OC_lt
	OC_ge
	OC_blnot
	OC_bland
	OC_blxor
	OC_blor
	OC_not
	OC_and
	OC_xor
	OC_or
	OC_shl
	OC_shr
	OC_add
	OC_sub
	OC_mul
	OC_div
	OC_mod
	OC_pow
	OC_abs
	OC_exp
	OC_ln
	OC_log
	OC_cos
	OC_sin
	OC_tan
	OC_acos
	OC_asin
	OC_atan
	OC_floor
	OC_ceil
	OC_ifelse
	OC_time
	OC_animtime
	OC_animelemtime
	OC_animelemno
	OC_statetype
	OC_movetype
	OC_ctrl
	OC_command
	OC_random
	OC_pos_x
	OC_pos_y
	OC_vel_x
	OC_vel_y
	OC_screenpos_x
	OC_screenpos_y
	OC_facing
	OC_anim
	OC_animexist
	OC_selfanimexist
	OC_alive
	OC_life
	OC_lifemax
	OC_power
	OC_powermax
	OC_canrecover
	OC_roundstate
	OC_ishelper
	OC_numhelper
	OC_numexplod
	OC_numprojid
	OC_numproj
	OC_teammode
	OC_teamside
	OC_hitdefattr
	OC_inguarddist
	OC_movecontact
	OC_movehit
	OC_moveguarded
	OC_movereversed
	OC_projcontacttime
	OC_projhittime
	OC_projguardedtime
	OC_projcanceltime
	OC_backedge
	OC_backedgedist
	OC_backedgebodydist
	OC_frontedge
	OC_frontedgedist
	OC_frontedgebodydist
	OC_leftedge
	OC_rightedge
	OC_topedge
	OC_bottomedge
	OC_camerapos_x
	OC_camerapos_y
	OC_camerazoom
	OC_gamewidth
	OC_gameheight
	OC_screenwidth
	OC_screenheight
	OC_stateno
	OC_prevstateno
	OC_id
	OC_playeridexist
	OC_gametime
	OC_numtarget
	OC_numenemy
	OC_numpartner
	OC_ailevel
	OC_palno
	OC_hitcount
	OC_uniqhitcount
	OC_hitpausetime
	OC_hitover
	OC_hitshakeover
	OC_hitfall
	OC_hitvel_x
	OC_hitvel_y
	OC_roundno
	OC_roundsexisted
	OC_ishometeam
	OC_parent
	OC_root
	OC_helper
	OC_target
	OC_partner
	OC_enemy
	OC_enemynear
	OC_playerid
	OC_p2
	OC_const_
	OC_gethitvar_
	OC_stagevar_
	OC_ex_
	OC_var0     OpCode = 0
	OC_sysvar0  OpCode = 60
	OC_fvar0    OpCode = 65
	OC_sysfvar0 OpCode = 105
)
const (
	OC_const_data_life OpCode = iota
	OC_const_data_power
	OC_const_data_attack
	OC_const_data_defence
	OC_const_data_fall_defence_mul
	OC_const_data_liedown_time
	OC_const_data_airjuggle
	OC_const_data_sparkno
	OC_const_data_guard_sparkno
	OC_const_data_ko_echo
	OC_const_data_intpersistindex
	OC_const_data_floatpersistindex
	OC_const_size_xscale
	OC_const_size_yscale
	OC_const_size_ground_back
	OC_const_size_ground_front
	OC_const_size_air_back
	OC_const_size_air_front
	OC_const_size_z_width
	OC_const_size_height
	OC_const_size_attack_dist
	OC_const_size_attack_z_width_back
	OC_const_size_attack_z_width_front
	OC_const_size_proj_attack_dist
	OC_const_size_proj_doscale
	OC_const_size_head_pos_x
	OC_const_size_head_pos_y
	OC_const_size_mid_pos_x
	OC_const_size_mid_pos_y
	OC_const_size_shadowoffset
	OC_const_size_draw_offset_x
	OC_const_size_draw_offset_y
	OC_const_velocity_walk_fwd_x
	OC_const_velocity_walk_back_x
	OC_const_velocity_walk_up_x
	OC_const_velocity_walk_down_x
	OC_const_velocity_run_fwd_x
	OC_const_velocity_run_fwd_y
	OC_const_velocity_run_back_x
	OC_const_velocity_run_back_y
	OC_const_velocity_run_up_x
	OC_const_velocity_run_up_y
	OC_const_velocity_run_down_x
	OC_const_velocity_run_down_y
	OC_const_velocity_jump_y
	OC_const_velocity_jump_neu_x
	OC_const_velocity_jump_back_x
	OC_const_velocity_jump_fwd_x
	OC_const_velocity_jump_up_x
	OC_const_velocity_jump_down_x
	OC_const_velocity_runjump_back_x
	OC_const_velocity_runjump_back_y
	OC_const_velocity_runjump_fwd_x
	OC_const_velocity_runjump_fwd_y
	OC_const_velocity_runjump_up_x
	OC_const_velocity_runjump_down_x
	OC_const_velocity_airjump_y
	OC_const_velocity_airjump_neu_x
	OC_const_velocity_airjump_back_x
	OC_const_velocity_airjump_fwd_x
	OC_const_velocity_airjump_up_x
	OC_const_velocity_airjump_down_x
	OC_const_velocity_air_gethit_groundrecover_x
	OC_const_velocity_air_gethit_groundrecover_y
	OC_const_velocity_air_gethit_airrecover_mul_x
	OC_const_velocity_air_gethit_airrecover_mul_y
	OC_const_velocity_air_gethit_airrecover_add_x
	OC_const_velocity_air_gethit_airrecover_add_y
	OC_const_velocity_air_gethit_airrecover_back
	OC_const_velocity_air_gethit_airrecover_fwd
	OC_const_velocity_air_gethit_airrecover_up
	OC_const_velocity_air_gethit_airrecover_down
	OC_const_movement_airjump_num
	OC_const_movement_airjump_height
	OC_const_movement_yaccel
	OC_const_movement_stand_friction
	OC_const_movement_crouch_friction
	OC_const_movement_stand_friction_threshold
	OC_const_movement_crouch_friction_threshold
	OC_const_movement_jump_changeanim_threshold
	OC_const_movement_air_gethit_groundlevel
	OC_const_movement_air_gethit_groundrecover_ground_threshold
	OC_const_movement_air_gethit_groundrecover_groundlevel
	OC_const_movement_air_gethit_airrecover_threshold
	OC_const_movement_air_gethit_airrecover_yaccel
	OC_const_movement_air_gethit_trip_groundlevel
	OC_const_movement_down_bounce_offset_x
	OC_const_movement_down_bounce_offset_y
	OC_const_movement_down_bounce_yaccel
	OC_const_movement_down_bounce_groundlevel
	OC_const_movement_down_friction_threshold
)
const (
	OC_gethitvar_animtype OpCode = iota
	OC_gethitvar_airtype
	OC_gethitvar_groundtype
	OC_gethitvar_damage
	OC_gethitvar_hitcount
	OC_gethitvar_fallcount
	OC_gethitvar_hitshaketime
	OC_gethitvar_hittime
	OC_gethitvar_slidetime
	OC_gethitvar_ctrltime
	OC_gethitvar_recovertime
	OC_gethitvar_xoff
	OC_gethitvar_yoff
	OC_gethitvar_xvel
	OC_gethitvar_yvel
	OC_gethitvar_yaccel
	OC_gethitvar_chainid
	OC_gethitvar_guarded
	OC_gethitvar_isbound
	OC_gethitvar_fall
	OC_gethitvar_fall_damage
	OC_gethitvar_fall_xvel
	OC_gethitvar_fall_yvel
	OC_gethitvar_fall_recover
	OC_gethitvar_fall_recovertime
	OC_gethitvar_fall_kill
	OC_gethitvar_fall_envshake_time
	OC_gethitvar_fall_envshake_freq
	OC_gethitvar_fall_envshake_ampl
	OC_gethitvar_fall_envshake_phase
)
const (
	OC_stagevar_info_author OpCode = iota
	OC_stagevar_info_displayname
	OC_stagevar_info_name
)
const (
	OC_ex_name OpCode = iota
	OC_ex_authorname
	OC_ex_p2name
	OC_ex_p3name
	OC_ex_p4name
	OC_ex_p2dist_x
	OC_ex_p2dist_y
	OC_ex_p2bodydist_x
	OC_ex_p2bodydist_y
	OC_ex_parentdist_x
	OC_ex_parentdist_y
	OC_ex_rootdist_x
	OC_ex_rootdist_y
	OC_ex_win
	OC_ex_winko
	OC_ex_wintime
	OC_ex_winperfect
	OC_ex_lose
	OC_ex_loseko
	OC_ex_losetime
	OC_ex_drawgame
	OC_ex_matchover
	OC_ex_matchno
	OC_ex_tickspersecond
)

type StringPool struct {
	List []string
	Map  map[string]int
}

func NewStringPool() *StringPool {
	return &StringPool{Map: make(map[string]int)}
}
func (sp *StringPool) Clear(s string) {
	sp.List, sp.Map = nil, make(map[string]int)
}
func (sp *StringPool) Add(s string) int {
	i, ok := sp.Map[s]
	if !ok {
		i = len(sp.List)
		sp.List = append(sp.List, s)
		sp.Map[s] = i
	}
	return i
}

type BytecodeValue struct {
	t ValueType
	v float64
}

func (bv BytecodeValue) IsSF() bool { return bv.t == VT_SFalse }
func (bv BytecodeValue) ToF() float32 {
	if bv.IsSF() {
		return 0
	}
	return float32(bv.v)
}
func (bv BytecodeValue) ToI() int32 {
	if bv.IsSF() {
		return 0
	}
	return int32(bv.v)
}
func (bv BytecodeValue) ToB() bool {
	if bv.IsSF() || bv.v == 0 {
		return false
	}
	return true
}
func (bv *BytecodeValue) SetF(f float32) {
	*bv = BytecodeValue{VT_Float, float64(f)}
}
func (bv *BytecodeValue) SetI(i int32) {
	*bv = BytecodeValue{VT_Int, float64(i)}
}
func (bv *BytecodeValue) SetB(b bool) {
	bv.t = VT_Bool
	if b {
		bv.v = 1
	} else {
		bv.v = 0
	}
}

func BytecodeSF() BytecodeValue {
	return BytecodeValue{VT_SFalse, math.NaN()}
}
func BytecodeFloat(f float32) BytecodeValue {
	return BytecodeValue{VT_Float, float64(f)}
}
func BytecodeInt(i int32) BytecodeValue {
	return BytecodeValue{VT_Int, float64(i)}
}
func BytecodeBool(b bool) BytecodeValue {
	return BytecodeValue{VT_Bool, float64(Btoi(b))}
}

type BytecodeStack []BytecodeValue

func (bs *BytecodeStack) Clear()                { *bs = (*bs)[:0] }
func (bs *BytecodeStack) Push(bv BytecodeValue) { *bs = append(*bs, bv) }
func (bs BytecodeStack) Top() *BytecodeValue {
	return &bs[len(bs)-1]
}
func (bs *BytecodeStack) Pop() (bv BytecodeValue) {
	bv, *bs = *bs.Top(), (*bs)[:len(*bs)-1]
	return
}
func (bs *BytecodeStack) Dup() {
	bs.Push(*bs.Top())
}
func (bs *BytecodeStack) Swap() {
	*bs.Top(), (*bs)[len(*bs)-2] = (*bs)[len(*bs)-2], *bs.Top()
}

type BytecodeExp []OpCode

func (be *BytecodeExp) append(op ...OpCode) {
	*be = append(*be, op...)
}
func (be *BytecodeExp) appendValue(bv BytecodeValue) (ok bool) {
	switch bv.t {
	case VT_Float:
		be.append(OC_float)
		f := float32(bv.v)
		be.append((*(*[4]OpCode)(unsafe.Pointer(&f)))[:]...)
	case VT_Int:
		if bv.v >= -128 || bv.v <= 127 {
			be.append(OC_int8, OpCode(bv.v))
		} else {
			be.append(OC_int)
			i := int32(bv.v)
			be.append((*(*[4]OpCode)(unsafe.Pointer(&i)))[:]...)
		}
	case VT_Bool:
		if bv.v != 0 {
			be.append(OC_int8, 1)
		} else {
			be.append(OC_int8, 0)
		}
	default:
		return false
	}
	return true
}
func (be *BytecodeExp) appendJmp(op OpCode, addr int32) {
	be.append(OC_int)
	be.append((*(*[4]OpCode)(unsafe.Pointer(&addr)))[:]...)
}
func (_ BytecodeExp) blnot(v *BytecodeValue) {
	if v.ToB() {
		v.v = 0
	} else {
		v.v = 1
	}
	v.t = VT_Int
}
func (_ BytecodeExp) pow(v1 *BytecodeValue, v2 BytecodeValue, pn int) {
	if ValueType(Min(int32(v1.t), int32(v2.t))) == VT_Float {
		v1.SetF(float32(math.Pow(float64(v1.ToF()), float64(v2.ToF()))))
	} else if v2.ToF() < 0 {
		if sys.cgi[pn].ver[0] == 1 {
			v1.SetF(float32(math.Pow(float64(v1.ToI()), float64(v2.ToI()))))
		} else {
			f := float32(math.Pow(float64(v1.ToI()), float64(v2.ToI())))
			v1.SetI(*(*int32)(unsafe.Pointer(&f)) << 29)
		}
	} else {
		i1, i2, hb := v1.ToI(), v2.ToI(), int32(-1)
		for uint32(i2)>>uint(hb+1) != 0 {
			hb++
		}
		var i, bit, tmp int32 = 1, 0, i1
		for ; bit <= hb; bit++ {
			var shift uint
			if bit == hb || sys.cgi[pn].ver[0] == 1 {
				shift = uint(bit)
			} else {
				shift = uint((hb - 1) - bit)
			}
			if i2&(1<<shift) != 0 {
				i *= tmp
			}
			tmp *= tmp
		}
		v1.SetI(i)
	}
}
func (_ BytecodeExp) mul(v1 *BytecodeValue, v2 BytecodeValue) {
	if ValueType(Min(int32(v1.t), int32(v2.t))) == VT_Float {
		v1.SetF(v1.ToF() * v2.ToF())
	} else {
		v1.SetI(v1.ToI() * v2.ToI())
	}
}
func (_ BytecodeExp) div(v1 *BytecodeValue, v2 BytecodeValue) {
	if ValueType(Min(int32(v1.t), int32(v2.t))) == VT_Float {
		v1.SetF(v1.ToF() / v2.ToF())
	} else if v2.ToI() == 0 {
		*v1 = BytecodeSF()
	} else {
		v1.SetI(v1.ToI() / v2.ToI())
	}
}
func (_ BytecodeExp) mod(v1 *BytecodeValue, v2 BytecodeValue) {
	if v2.ToI() == 0 {
		*v1 = BytecodeSF()
	} else {
		v1.SetI(v1.ToI() % v2.ToI())
	}
}
func (_ BytecodeExp) add(v1 *BytecodeValue, v2 BytecodeValue) {
	if ValueType(Min(int32(v1.t), int32(v2.t))) == VT_Float {
		v1.SetF(v1.ToF() + v2.ToF())
	} else {
		v1.SetI(v1.ToI() + v2.ToI())
	}
}
func (_ BytecodeExp) sub(v1 *BytecodeValue, v2 BytecodeValue) {
	if ValueType(Min(int32(v1.t), int32(v2.t))) == VT_Float {
		v1.SetF(v1.ToF() - v2.ToF())
	} else {
		v1.SetI(v1.ToI() - v2.ToI())
	}
}
func (_ BytecodeExp) gt(v1 *BytecodeValue, v2 BytecodeValue) {
	if ValueType(Min(int32(v1.t), int32(v2.t))) == VT_Float {
		v1.SetB(v1.ToF() > v2.ToF())
	} else {
		v1.SetB(v1.ToI() > v2.ToI())
	}
}
func (_ BytecodeExp) ge(v1 *BytecodeValue, v2 BytecodeValue) {
	if ValueType(Min(int32(v1.t), int32(v2.t))) == VT_Float {
		v1.SetB(v1.ToF() >= v2.ToF())
	} else {
		v1.SetB(v1.ToI() >= v2.ToI())
	}
}
func (_ BytecodeExp) lt(v1 *BytecodeValue, v2 BytecodeValue) {
	if ValueType(Min(int32(v1.t), int32(v2.t))) == VT_Float {
		v1.SetB(v1.ToF() < v2.ToF())
	} else {
		v1.SetB(v1.ToI() < v2.ToI())
	}
}
func (_ BytecodeExp) le(v1 *BytecodeValue, v2 BytecodeValue) {
	if ValueType(Min(int32(v1.t), int32(v2.t))) == VT_Float {
		v1.SetB(v1.ToF() <= v2.ToF())
	} else {
		v1.SetB(v1.ToI() <= v2.ToI())
	}
}
func (_ BytecodeExp) eq(v1 *BytecodeValue, v2 BytecodeValue) {
	if ValueType(Min(int32(v1.t), int32(v2.t))) == VT_Float {
		v1.SetB(v1.ToF() == v2.ToF())
	} else {
		v1.SetB(v1.ToI() == v2.ToI())
	}
}
func (_ BytecodeExp) ne(v1 *BytecodeValue, v2 BytecodeValue) {
	if ValueType(Min(int32(v1.t), int32(v2.t))) == VT_Float {
		v1.SetB(v1.ToF() != v2.ToF())
	} else {
		v1.SetB(v1.ToI() != v2.ToI())
	}
}
func (_ BytecodeExp) and(v1 *BytecodeValue, v2 BytecodeValue) {
	v1.SetI(v1.ToI() & v2.ToI())
}
func (_ BytecodeExp) xor(v1 *BytecodeValue, v2 BytecodeValue) {
	v1.SetI(v1.ToI() ^ v2.ToI())
}
func (_ BytecodeExp) or(v1 *BytecodeValue, v2 BytecodeValue) {
	v1.SetI(v1.ToI() | v2.ToI())
}
func (_ BytecodeExp) bland(v1 *BytecodeValue, v2 BytecodeValue) {
	v1.SetB(v1.ToB() && v2.ToB())
}
func (_ BytecodeExp) blxor(v1 *BytecodeValue, v2 BytecodeValue) {
	v1.SetB(v1.ToB() != v2.ToB())
}
func (_ BytecodeExp) blor(v1 *BytecodeValue, v2 BytecodeValue) {
	v1.SetB(v1.ToB() || v2.ToB())
}
func (be BytecodeExp) run(c *Char, scpn int) BytecodeValue {
	sys.bcStack.Clear()
	orgc := c
	for i := 1; i <= len(be); i++ {
		switch be[i-1] {
		case OC_jz8, OC_jnz8:
			if sys.bcStack.Top().ToB() == (be[i-1] == OC_jz8) {
				i++
				break
			}
			fallthrough
		case OC_jmp8:
			if be[i] == 0 {
				i = len(be)
			} else {
				i += int(uint8(be[i])) + 1
			}
		case OC_jz, OC_jnz:
			if sys.bcStack.Top().ToB() == (be[i-1] == OC_jz) {
				i += 4
				break
			}
			fallthrough
		case OC_jmp:
			i += int(*(*int32)(unsafe.Pointer(&be[i]))) + 4
		case OC_int8:
			sys.bcStack.Push(BytecodeInt(int32(int8(be[i]))))
			i++
		case OC_int:
			sys.bcStack.Push(BytecodeInt(*(*int32)(unsafe.Pointer(&be[i]))))
			i += 4
		case OC_float:
			sys.bcStack.Push(BytecodeFloat(*(*float32)(unsafe.Pointer(&be[i]))))
			i += 4
		case OC_blnot:
			be.blnot(sys.bcStack.Top())
		case OC_pow:
			v2 := sys.bcStack.Pop()
			be.pow(sys.bcStack.Top(), v2, scpn)
		case OC_mul:
			v2 := sys.bcStack.Pop()
			be.mul(sys.bcStack.Top(), v2)
		case OC_div:
			v2 := sys.bcStack.Pop()
			be.div(sys.bcStack.Top(), v2)
		case OC_mod:
			v2 := sys.bcStack.Pop()
			be.mod(sys.bcStack.Top(), v2)
		case OC_add:
			v2 := sys.bcStack.Pop()
			be.add(sys.bcStack.Top(), v2)
		case OC_sub:
			v2 := sys.bcStack.Pop()
			be.sub(sys.bcStack.Top(), v2)
		case OC_gt:
			v2 := sys.bcStack.Pop()
			be.gt(sys.bcStack.Top(), v2)
		case OC_ge:
			v2 := sys.bcStack.Pop()
			be.ge(sys.bcStack.Top(), v2)
		case OC_lt:
			v2 := sys.bcStack.Pop()
			be.lt(sys.bcStack.Top(), v2)
		case OC_le:
			v2 := sys.bcStack.Pop()
			be.le(sys.bcStack.Top(), v2)
		case OC_eq:
			v2 := sys.bcStack.Pop()
			be.eq(sys.bcStack.Top(), v2)
		case OC_ne:
			v2 := sys.bcStack.Pop()
			be.ne(sys.bcStack.Top(), v2)
		case OC_and:
			v2 := sys.bcStack.Pop()
			be.and(sys.bcStack.Top(), v2)
		case OC_xor:
			v2 := sys.bcStack.Pop()
			be.xor(sys.bcStack.Top(), v2)
		case OC_or:
			v2 := sys.bcStack.Pop()
			be.or(sys.bcStack.Top(), v2)
		case OC_bland:
			v2 := sys.bcStack.Pop()
			be.bland(sys.bcStack.Top(), v2)
		case OC_blxor:
			v2 := sys.bcStack.Pop()
			be.blxor(sys.bcStack.Top(), v2)
		case OC_blor:
			v2 := sys.bcStack.Pop()
			be.blor(sys.bcStack.Top(), v2)
		case OC_ifelse:
			v3 := sys.bcStack.Pop()
			v2 := sys.bcStack.Pop()
			if sys.bcStack.Top().ToB() {
				*sys.bcStack.Top() = v2
			} else {
				*sys.bcStack.Top() = v3
			}
		case OC_pop:
			sys.bcStack.Pop()
		case OC_dup:
			sys.bcStack.Dup()
		case OC_swap:
			sys.bcStack.Swap()
		case OC_run:
			l := int(*(*int32)(unsafe.Pointer(&be[i])))
			sys.bcStack.Push(be[i+4:i+4+l].run(c, scpn))
			i += 4 + l
		case OC_time:
			sys.bcStack.Push(BytecodeInt(c.time()))
		case OC_alive:
			sys.bcStack.Push(BytecodeBool(c.alive()))
		case OC_random:
			sys.bcStack.Push(BytecodeInt(Rand(0, 999)))
		default:
			unimplemented()
		}
		c = orgc
	}
	return sys.bcStack.Pop()
}
func (be BytecodeExp) evalF(c *Char, scpn int) float32 {
	return be.run(c, scpn).ToF()
}
func (be BytecodeExp) evalI(c *Char, scpn int) int32 {
	return be.run(c, scpn).ToI()
}
func (be BytecodeExp) evalB(c *Char, scpn int) bool {
	return be.run(c, scpn).ToB()
}

type StateController interface {
	Run(c *Char, ps *int32) (changeState bool)
}

const SCID_trigger byte = 255

type StateControllerBase struct {
	playerNo       int
	persistent     int32
	ignorehitpause bool
	code           []byte
}

func newStateControllerBase(pn int) *StateControllerBase {
	return &StateControllerBase{playerNo: pn, persistent: 1}
}
func (_ StateControllerBase) beToExp(be ...BytecodeExp) []BytecodeExp {
	return be
}
func (_ StateControllerBase) fToExp(f ...float32) (exp []BytecodeExp) {
	for _, v := range f {
		var be BytecodeExp
		be.appendValue(BytecodeFloat(v))
		exp = append(exp, be)
	}
	return
}
func (_ StateControllerBase) iToExp(i ...int32) (exp []BytecodeExp) {
	for _, v := range i {
		var be BytecodeExp
		be.appendValue(BytecodeInt(v))
		exp = append(exp, be)
	}
	return
}
func (scb *StateControllerBase) add(id byte, exp []BytecodeExp) {
	scb.code = append(scb.code, id, byte(len(exp)))
	for _, e := range exp {
		l := int32(len(e))
		scb.code = append(scb.code, (*(*[4]byte)(unsafe.Pointer(&l)))[:]...)
		scb.code = append(scb.code, (*(*[]byte)(unsafe.Pointer(&e)))...)
	}
}
func (scb StateControllerBase) run(c *Char, ps *int32,
	f func(byte, []BytecodeExp) bool) bool {
	(*ps)--
	if *ps > 0 {
		return false
	}
	for i := 0; i < len(scb.code); {
		id := scb.code[i]
		i++
		n := scb.code[i]
		i++
		exp := make([]BytecodeExp, n)
		for m := 0; m < int(n); m++ {
			l := *(*int32)(unsafe.Pointer(&scb.code[i]))
			i += 4
			exp[m] = (*(*BytecodeExp)(unsafe.Pointer(&scb.code)))[i : i+int(l)]
			i += int(l)
		}
		if id == SCID_trigger {
			if !exp[0].evalB(c, scb.playerNo) {
				return false
			}
		} else if !f(id, exp) {
			break
		}
	}
	*ps = scb.persistent
	return true
}

type stateDef StateControllerBase

const (
	stateDef_hitcountpersist byte = iota
	stateDef_movehitpersist
	stateDef_hitdefpersist
	stateDef_sprpriority
	stateDef_facep2
	stateDef_juggle
	stateDef_velset
	stateDef_anim
	stateDef_ctrl
	stateDef_poweradd
)

func (sc stateDef) Run(c *Char, ps *int32) bool {
	StateControllerBase(sc).run(c, ps, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case stateDef_hitcountpersist:
			if !exp[0].evalB(c, sc.playerNo) {
				c.clearHitCount()
			}
		case stateDef_movehitpersist:
			if !exp[0].evalB(c, sc.playerNo) {
				c.clearMoveHit()
			}
		case stateDef_hitdefpersist:
			if !exp[0].evalB(c, sc.playerNo) {
				c.clearHitDef()
			}
		case stateDef_sprpriority:
			c.setSprPriority(exp[0].evalI(c, sc.playerNo))
		case stateDef_facep2:
			if exp[0].evalB(c, sc.playerNo) {
				c.faceP2()
			}
		case stateDef_juggle:
			c.setJuggle(exp[0].evalI(c, sc.playerNo))
		case stateDef_velset:
			c.setXV(exp[0].evalF(c, sc.playerNo))
			if len(exp) > 1 {
				c.setYV(exp[1].evalF(c, sc.playerNo))
				if len(exp) > 2 {
					exp[2].run(c, sc.playerNo)
				}
			}
		case stateDef_anim:
			c.changeAnim(exp[0].evalI(c, sc.playerNo))
		case stateDef_ctrl:
			c.setCtrl(exp[0].evalB(c, sc.playerNo))
		case stateDef_poweradd:
			c.addPower(exp[0].evalI(c, sc.playerNo))
		}
		return true
	})
	return false
}

type hitBy StateControllerBase

const (
	hitBy_value byte = iota
	hitBy_value2
	hitBy_time
)

func (sc hitBy) Run(c *Char, ps *int32) bool {
	time := int32(1)
	StateControllerBase(sc).run(c, ps, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case hitBy_time:
			time = exp[0].evalI(c, sc.playerNo)
		case hitBy_value:
			unimplemented()
		case hitBy_value2:
			unimplemented()
		}
		return true
	})
	return false
}

type notHitBy hitBy

func (sc notHitBy) Run(c *Char, ps *int32) bool {
	time := int32(1)
	StateControllerBase(sc).run(c, ps, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case hitBy_time:
			time = exp[0].evalI(c, sc.playerNo)
		case hitBy_value:
			unimplemented()
		case hitBy_value2:
			unimplemented()
		}
		return true
	})
	return false
}

type assertSpecial StateControllerBase

const (
	assertSpecial_flag byte = iota
	assertSpecial_flag_g
)

func (sc assertSpecial) Run(c *Char, ps *int32) bool {
	StateControllerBase(sc).run(c, ps, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case assertSpecial_flag:
			unimplemented()
		case assertSpecial_flag_g:
			sys.specialFlag |= GlobalSpecialFlag(exp[0].evalI(c, sc.playerNo))
		}
		return true
	})
	return false
}

type playSnd StateControllerBase

const (
	playSnd_value = iota
	playSnd_channel
	playSnd_lowpriority
	playSnd_pan
	playSnd_abspan
	playSnd_volume
	playSnd_freqmul
	playSnd_loop
)

func (sc playSnd) Run(c *Char, ps *int32) bool {
	f, lw, lp := false, false, false
	var g, n, ch, vo int32 = -1, 0, -1, 0
	if sys.cgi[sc.playerNo].ver[0] == 1 {
		vo = 100
	}
	var p, fr float32 = 0, 1
	x := &c.pos[0]
	StateControllerBase(sc).run(c, ps, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case playSnd_value:
			f = exp[0].evalB(c, sc.playerNo)
			g = exp[1].evalI(c, sc.playerNo)
			if len(exp) > 2 {
				n = exp[2].evalI(c, sc.playerNo)
			}
		case playSnd_channel:
			ch = exp[0].evalI(c, sc.playerNo)
		case playSnd_lowpriority:
			lw = exp[0].evalB(c, sc.playerNo)
		case playSnd_pan:
			p = exp[0].evalF(c, sc.playerNo)
		case playSnd_abspan:
			x = nil
			p = exp[0].evalF(c, sc.playerNo)
		case playSnd_volume:
			vo = exp[0].evalI(c, sc.playerNo)
		case playSnd_freqmul:
			fr = exp[0].evalF(c, sc.playerNo)
		case playSnd_loop:
			lp = exp[0].evalB(c, sc.playerNo)
		}
		return true
	})
	c.playSound(f, lw, lp, g, n, ch, vo, p, fr, x)
	return false
}

type changeState StateControllerBase

const (
	changeState_value byte = iota
	changeState_ctrl
	changeState_anim
)

func (sc changeState) Run(c *Char, ps *int32) bool {
	var v, a, ctrl int32 = -1, -1, -1
	StateControllerBase(sc).run(c, ps, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case changeState_value:
			v = exp[0].evalI(c, sc.playerNo)
		case changeState_ctrl:
			ctrl = exp[0].evalI(c, sc.playerNo)
		case changeState_anim:
			a = exp[0].evalI(c, sc.playerNo)
		}
		return true
	})
	c.changeState(v, a, ctrl)
	return true
}

type selfState changeState

func (sc selfState) Run(c *Char, ps *int32) bool {
	var v, a, ctrl int32 = -1, -1, -1
	StateControllerBase(sc).run(c, ps, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case changeState_value:
			v = exp[0].evalI(c, sc.playerNo)
		case changeState_ctrl:
			ctrl = exp[0].evalI(c, sc.playerNo)
		case changeState_anim:
			a = exp[0].evalI(c, sc.playerNo)
		}
		return true
	})
	c.selfState(v, a, ctrl)
	return true
}

type tagIn StateControllerBase

const (
	tagIn_stateno = iota
	tagIn_partnerstateno
)

func (sc tagIn) Run(c *Char, ps *int32) bool {
	var p *Char
	sn := int32(-1)
	ret := false
	StateControllerBase(sc).run(c, ps, func(id byte, exp []BytecodeExp) bool {
		if p == nil {
			p = c.partner(0)
			if p == nil {
				return false
			}
		}
		switch id {
		case tagIn_stateno:
			sn = exp[0].evalI(c, sc.playerNo)
		case tagIn_partnerstateno:
			if psn := exp[0].evalI(c, sc.playerNo); psn >= 0 {
				if sn >= 0 {
					c.changeState(sn, -1, -1)
				}
				p.standby = false
				p.changeState(psn, -1, -1)
				ret = true
			} else {
				return false
			}
		}
		return true
	})
	return ret
}

type tagOut StateControllerBase

const (
	tagOut_ = iota
)

func (sc tagOut) Run(c *Char, ps *int32) bool {
	StateControllerBase(sc).run(c, ps, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case tagOut_:
			c.standby = true
		}
		return true
	})
	return true
}

type destroySelf StateControllerBase

const (
	destroySelf_recursive = iota
	destroySelf_removeexplods
)

func (sc destroySelf) Run(c *Char, ps *int32) bool {
	rec, rem := false, false
	StateControllerBase(sc).run(c, ps, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case destroySelf_recursive:
			rec = exp[0].evalB(c, sc.playerNo)
		case destroySelf_removeexplods:
			rem = exp[0].evalB(c, sc.playerNo)
		}
		return true
	})
	return c.destroySelf(rec, rem)
}

type changeAnim StateControllerBase

const (
	changeAnim_elem byte = iota
	changeAnim_value
)

func (sc changeAnim) Run(c *Char, ps *int32) bool {
	var elem int32
	setelem := false
	StateControllerBase(sc).run(c, ps, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case changeAnim_elem:
			elem = exp[0].evalI(c, sc.playerNo)
			setelem = true
		case changeAnim_value:
			c.changeAnim(exp[0].evalI(c, sc.playerNo))
			if setelem {
				c.setAnimElem(elem)
			}
		}
		return true
	})
	return false
}

type changeAnim2 changeAnim

func (sc changeAnim2) Run(c *Char, ps *int32) bool {
	var elem int32
	setelem := false
	StateControllerBase(sc).run(c, ps, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case changeAnim_elem:
			elem = exp[0].evalI(c, sc.playerNo)
			setelem = true
		case changeAnim_value:
			c.changeAnim2(exp[0].evalI(c, sc.playerNo))
			if setelem {
				c.setAnimElem(elem)
			}
		}
		return true
	})
	return false
}

type helper StateControllerBase

const (
	helper_helpertype byte = iota
	helper_name
)

func (sc helper) Run(c *Char, ps *int32) bool {
	var h *Char
	StateControllerBase(sc).run(c, ps, func(id byte, exp []BytecodeExp) bool {
		if h == nil {
			h = c.newHelper()
			if h == nil {
				return false
			}
		}
		switch id {
		case helper_helpertype:
			h.player = exp[0].evalB(c, sc.playerNo)
		case helper_name:
			h.name = string(*(*[]byte)(unsafe.Pointer(&exp[0])))
		}
		unimplemented()
		return true
	})
	return false
}

type StateBytecode struct {
	stateType StateType
	moveType  MoveType
	physics   StateType
	stateDef  StateController
	ctrls     []StateController
}

func newStateBytecode() *StateBytecode {
	return &StateBytecode{stateType: ST_S, moveType: MT_I, physics: ST_N}
}

type Bytecode struct{ states map[int32]StateBytecode }

func newBytecode() *Bytecode {
	return &Bytecode{states: make(map[int32]StateBytecode)}
}
