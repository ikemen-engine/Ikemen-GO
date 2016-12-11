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
	VT_Variant ValueType = iota
	VT_Float
	VT_Int
	VT_Bool
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
	OC_dup
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
	OC_matchover
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
	OC_matchno
	OC_ishometeam
	OC_tickspersecond
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

type BytecodeExp []OpCode

func (be *BytecodeExp) appendFloat(f float32) {
	*be = append(*be, (*(*[4]OpCode)(unsafe.Pointer(&f)))[:]...)
}
func (be *BytecodeExp) appendInt(i int32) {
	*be = append(*be, (*(*[4]OpCode)(unsafe.Pointer(&i)))[:]...)
}
func (be BytecodeExp) toF() float32 {
	return *(*float32)(unsafe.Pointer(&be[0]))
}
func (be BytecodeExp) toI() int32 {
	return *(*int32)(unsafe.Pointer(&be[0]))
}
func (be *BytecodeExp) AppendValue(t ValueType, v float64) (ok bool) {
	if math.IsNaN(v) {
		return false
	}
	switch t {
	case VT_Float:
		*be = append(*be, OC_float)
		be.appendFloat(float32(v))
	case VT_Int:
		if v >= -128 || v <= 127 {
			*be = append(*be, OC_int8, OpCode(v))
		} else {
			*be = append(*be, OC_int)
			be.appendInt(int32(v))
		}
	case VT_Bool:
		if v != 0 {
			*be = append(*be, OC_int8, 1)
		} else {
			*be = append(*be, OC_int8, 0)
		}
	default:
		return false
	}
	return true
}
func (be BytecodeExp) run(c *Char) (t ValueType, v float64) {
	unimplemented()
	return VT_Int, 0
}
func (be BytecodeExp) eval(c *Char) float64 {
	_, v := be.run(c)
	return v
}

type StateController interface {
	Run(c *Char) (changeState bool)
}

const (
	SCID_trigger byte = 0
	SCID_const   byte = 128
)

type StateControllerBase []byte

func (scb StateControllerBase) beToExp(be ...BytecodeExp) []BytecodeExp {
	return be
}
func (scb StateControllerBase) fToExp(f ...float32) (exp []BytecodeExp) {
	for _, v := range f {
		var be BytecodeExp
		be.appendFloat(v)
		exp = append(exp, be)
	}
	return
}
func (scb StateControllerBase) iToExp(i ...int32) (exp []BytecodeExp) {
	for _, v := range i {
		var be BytecodeExp
		be.appendInt(v)
		exp = append(exp, be)
	}
	return
}
func (scb *StateControllerBase) add(id byte, exp []BytecodeExp) {
	*scb = append(*scb, id, byte(len(exp)))
	for _, e := range exp {
		l := int32(len(e))
		*scb = append(*scb, (*(*[4]byte)(unsafe.Pointer(&l)))[:]...)
		*scb = append(*scb, (*(*[]byte)(unsafe.Pointer(&e)))...)
	}
}
func (scb StateControllerBase) run(f func(byte, []BytecodeExp) bool) bool {
	for i := 0; i < len(scb); {
		id := scb[i]
		i++
		n := scb[i]
		i++
		exp := make([]BytecodeExp, n)
		for m := byte(0); m < n; m++ {
			l := *(*int32)(unsafe.Pointer(&scb[i]))
			i += 4
			exp[m] = (*(*BytecodeExp)(unsafe.Pointer(&scb)))[i : i+int(l)]
			i += int(l)
		}
		if !f(id, exp) {
			return false
		}
	}
	return true
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
