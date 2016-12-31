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
	ST_MASK = 1<<iota - 1
	ST_D    = ST_L
	ST_F    = ST_N
	ST_P    = ST_U
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
	AT_AA = AT_NA | AT_SA | AT_HA
	AT_AT = AT_NT | AT_ST | AT_HT
	AT_AP = AT_NP | AT_SP | AT_HP
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
	VT_None ValueType = iota
	VT_Float
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
	OC_nordrun
	OC_jsf8
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
	OC_neg
	OC_blnot
	OC_bland
	OC_blxor
	OC_blor
	OC_not
	OC_and
	OC_xor
	OC_or
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
	OC_roundsexisted
	OC_matchover
	OC_parent
	OC_root
	OC_helper
	OC_target
	OC_partner
	OC_enemy
	OC_enemynear
	OC_playerid
	OC_p2
	OC_rdreset
	OC_const_
	OC_st_
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
	OC_const_name
	OC_const_authorname
	OC_const_p2name
	OC_const_p3name
	OC_const_p4name
	OC_const_stagevar_info_author
	OC_const_stagevar_info_displayname
	OC_const_stagevar_info_name
)
const (
	OC_st_var OpCode = iota + OC_var*2
	OC_st_sysvar
	OC_st_fvar
	OC_st_sysfvar
	OC_st_varadd
	OC_st_sysvaradd
	OC_st_fvaradd
	OC_st_sysfvaradd
	OC_st_var0        OpCode = OC_var0
	OC_st_sysvar0     OpCode = OC_sysvar0
	OC_st_fvar0       OpCode = OC_fvar0
	OC_st_sysfvar0    OpCode = OC_sysfvar0
	OC_st_var0add     OpCode = OC_var + OC_var0
	OC_st_sysvar0add  OpCode = OC_var + OC_sysvar0
	OC_st_fvar0add    OpCode = OC_var + OC_fvar0
	OC_st_sysfvar0add OpCode = OC_var + OC_sysfvar0
)
const (
	OC_ex_p2dist_x OpCode = iota
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
	OC_ex_matchno
	OC_ex_roundno
	OC_ex_ishometeam
	OC_ex_tickspersecond
	OC_ex_timemod
	OC_ex_gethitvar_animtype
	OC_ex_gethitvar_airtype
	OC_ex_gethitvar_groundtype
	OC_ex_gethitvar_damage
	OC_ex_gethitvar_hitcount
	OC_ex_gethitvar_fallcount
	OC_ex_gethitvar_hitshaketime
	OC_ex_gethitvar_hittime
	OC_ex_gethitvar_slidetime
	OC_ex_gethitvar_ctrltime
	OC_ex_gethitvar_recovertime
	OC_ex_gethitvar_xoff
	OC_ex_gethitvar_yoff
	OC_ex_gethitvar_xvel
	OC_ex_gethitvar_yvel
	OC_ex_gethitvar_yaccel
	OC_ex_gethitvar_chainid
	OC_ex_gethitvar_guarded
	OC_ex_gethitvar_isbound
	OC_ex_gethitvar_fall
	OC_ex_gethitvar_fall_damage
	OC_ex_gethitvar_fall_xvel
	OC_ex_gethitvar_fall_yvel
	OC_ex_gethitvar_fall_recover
	OC_ex_gethitvar_fall_time
	OC_ex_gethitvar_fall_recovertime
	OC_ex_gethitvar_fall_kill
	OC_ex_gethitvar_fall_envshake_time
	OC_ex_gethitvar_fall_envshake_freq
	OC_ex_gethitvar_fall_envshake_ampl
	OC_ex_gethitvar_fall_envshake_phase
)
const (
	NumVar     = OC_sysvar0 - OC_var0
	NumSysVar  = OC_fvar0 - OC_sysvar0
	NumFvar    = OC_sysfvar0 - OC_fvar0
	NumSysFvar = OC_var - OC_sysfvar0
)

type StringPool struct {
	List []string
	Map  map[string]int
}

func NewStringPool() *StringPool {
	return &StringPool{Map: make(map[string]int)}
}
func (sp *StringPool) Clear() {
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

func (bv BytecodeValue) IsNone() bool { return bv.t == VT_None }
func (bv BytecodeValue) IsSF() bool   { return bv.t == VT_SFalse }
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
	if math.IsNaN(float64(f)) {
		*bv = BytecodeSF()
	} else {
		*bv = BytecodeValue{VT_Float, float64(f)}
	}
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

func bvNone() BytecodeValue {
	return BytecodeValue{VT_None, 0}
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
	case VT_SFalse:
		be.append(OC_int8, 0)
	default:
		return false
	}
	return true
}
func (be *BytecodeExp) appendI32Op(op OpCode, addr int32) {
	be.append(OC_int)
	be.append((*(*[4]OpCode)(unsafe.Pointer(&addr)))[:]...)
}
func (_ BytecodeExp) neg(v *BytecodeValue) {
	if v.t == VT_Bool {
		v.SetI(-v.ToI())
	} else {
		v.v *= -1
	}
}
func (_ BytecodeExp) not(v *BytecodeValue) {
	v.SetI(^v.ToI())
}
func (_ BytecodeExp) blnot(v *BytecodeValue) {
	v.SetB(!v.ToB())
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
func (_ BytecodeExp) abs(v1 *BytecodeValue) {
	v1.v = math.Abs(v1.v)
}
func (_ BytecodeExp) exp(v1 *BytecodeValue) {
	v1.SetF(float32(math.Exp(v1.v)))
}
func (_ BytecodeExp) ln(v1 *BytecodeValue) {
	if v1.v <= 0 {
		*v1 = BytecodeSF()
	} else {
		v1.SetF(float32(math.Log(v1.v)))
	}
}
func (_ BytecodeExp) log(v1 *BytecodeValue, v2 BytecodeValue) {
	if v1.v <= 0 || v2.v <= 0 {
		*v1 = BytecodeSF()
	} else {
		v1.SetF(float32(math.Log(v1.v) / math.Log(v2.v)))
	}
}
func (_ BytecodeExp) cos(v1 *BytecodeValue) {
	v1.SetF(float32(math.Cos(v1.v)))
}
func (_ BytecodeExp) sin(v1 *BytecodeValue) {
	v1.SetF(float32(math.Sin(v1.v)))
}
func (_ BytecodeExp) tan(v1 *BytecodeValue) {
	v1.SetF(float32(math.Tan(v1.v)))
}
func (_ BytecodeExp) acos(v1 *BytecodeValue) {
	v1.SetF(float32(math.Acos(v1.v)))
}
func (_ BytecodeExp) asin(v1 *BytecodeValue) {
	v1.SetF(float32(math.Asin(v1.v)))
}
func (_ BytecodeExp) atan(v1 *BytecodeValue) {
	v1.SetF(float32(math.Atan(v1.v)))
}
func (_ BytecodeExp) floor(v1 *BytecodeValue) {
	if v1.t == VT_Float {
		f := math.Floor(v1.v)
		if math.IsNaN(f) {
			*v1 = BytecodeSF()
		} else {
			v1.SetI(int32(f))
		}
	}
}
func (_ BytecodeExp) ceil(v1 *BytecodeValue) {
	if v1.t == VT_Float {
		f := math.Ceil(v1.v)
		if math.IsNaN(f) {
			*v1 = BytecodeSF()
		} else {
			v1.SetI(int32(f))
		}
	}
}
func (be BytecodeExp) run(c *Char) BytecodeValue {
	oc := c
	for i := 1; i <= len(be); i++ {
		switch be[i-1] {
		case OC_jsf8:
			if sys.bcStack.Top().IsSF() {
				if be[i] == 0 {
					i = len(be)
				} else {
					i += int(uint8(be[i])) + 1
				}
			} else {
				i++
			}
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
		case OC_parent:
			if c = c.parent(); c != nil {
				i += 4
				continue
			}
			sys.bcStack.Push(BytecodeSF())
			i += int(*(*int32)(unsafe.Pointer(&be[i]))) + 4
		case OC_root:
			if c = c.root(); c != nil {
				i += 4
				continue
			}
			sys.bcStack.Push(BytecodeSF())
			i += int(*(*int32)(unsafe.Pointer(&be[i]))) + 4
		case OC_helper:
			if c = c.helper(sys.bcStack.Pop().ToI()); c != nil {
				i += 4
				continue
			}
			sys.bcStack.Push(BytecodeSF())
			i += int(*(*int32)(unsafe.Pointer(&be[i]))) + 4
		case OC_target:
			if c = c.target(sys.bcStack.Pop().ToI()); c != nil {
				i += 4
				continue
			}
			sys.bcStack.Push(BytecodeSF())
			i += int(*(*int32)(unsafe.Pointer(&be[i]))) + 4
		case OC_partner:
			if c = c.partner(sys.bcStack.Pop().ToI()); c != nil {
				i += 4
				continue
			}
			sys.bcStack.Push(BytecodeSF())
			i += int(*(*int32)(unsafe.Pointer(&be[i]))) + 4
		case OC_enemy:
			if c = c.enemy(sys.bcStack.Pop().ToI()); c != nil {
				i += 4
				continue
			}
			sys.bcStack.Push(BytecodeSF())
			i += int(*(*int32)(unsafe.Pointer(&be[i]))) + 4
		case OC_enemynear:
			if c = c.enemynear(sys.bcStack.Pop().ToI()); c != nil {
				i += 4
				continue
			}
			sys.bcStack.Push(BytecodeSF())
			i += int(*(*int32)(unsafe.Pointer(&be[i]))) + 4
		case OC_playerid:
			if c = c.playerid(sys.bcStack.Pop().ToI()); c != nil {
				i += 4
				continue
			}
			sys.bcStack.Push(BytecodeSF())
			i += int(*(*int32)(unsafe.Pointer(&be[i]))) + 4
		case OC_p2:
			if c = c.p2(); c != nil {
				i += 4
				continue
			}
			sys.bcStack.Push(BytecodeSF())
			i += int(*(*int32)(unsafe.Pointer(&be[i]))) + 4
		case OC_rdreset:
			// NOP
		case OC_run:
			l := int(*(*int32)(unsafe.Pointer(&be[i])))
			sys.bcStack.Push(be[i+4 : i+4+l].run(c))
			i += 4 + l
		case OC_nordrun:
			l := int(*(*int32)(unsafe.Pointer(&be[i])))
			sys.bcStack.Push(be[i+4 : i+4+l].run(oc))
			i += 4 + l
			continue
		case OC_int8:
			sys.bcStack.Push(BytecodeInt(int32(int8(be[i]))))
			i++
		case OC_int:
			sys.bcStack.Push(BytecodeInt(*(*int32)(unsafe.Pointer(&be[i]))))
			i += 4
		case OC_float:
			sys.bcStack.Push(BytecodeFloat(*(*float32)(unsafe.Pointer(&be[i]))))
			i += 4
		case OC_command:
			sys.bcStack.Push(BytecodeBool(c.command(sys.workingChar.ss.sb.playerNo,
				int(*(*int32)(unsafe.Pointer(&be[i]))))))
			i += 4
		case OC_neg:
			be.neg(sys.bcStack.Top())
		case OC_not:
			be.not(sys.bcStack.Top())
		case OC_blnot:
			be.blnot(sys.bcStack.Top())
		case OC_pow:
			v2 := sys.bcStack.Pop()
			be.pow(sys.bcStack.Top(), v2, sys.workingChar.ss.sb.playerNo)
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
		case OC_abs:
			be.abs(sys.bcStack.Top())
		case OC_exp:
			be.exp(sys.bcStack.Top())
		case OC_ln:
			be.ln(sys.bcStack.Top())
		case OC_log:
			v2 := sys.bcStack.Pop()
			be.log(sys.bcStack.Top(), v2)
		case OC_cos:
			be.cos(sys.bcStack.Top())
		case OC_sin:
			be.sin(sys.bcStack.Top())
		case OC_tan:
			be.tan(sys.bcStack.Top())
		case OC_acos:
			be.acos(sys.bcStack.Top())
		case OC_asin:
			be.asin(sys.bcStack.Top())
		case OC_atan:
			be.atan(sys.bcStack.Top())
		case OC_floor:
			be.floor(sys.bcStack.Top())
		case OC_ceil:
			be.ceil(sys.bcStack.Top())
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
		case OC_time:
			sys.bcStack.Push(BytecodeInt(c.time()))
		case OC_alive:
			sys.bcStack.Push(BytecodeBool(c.alive()))
		case OC_random:
			sys.bcStack.Push(BytecodeInt(Rand(0, 999)))
		case OC_roundstate:
			sys.bcStack.Push(BytecodeInt(c.roundState()))
		case OC_anim:
			sys.bcStack.Push(BytecodeInt(c.animNo))
		case OC_animtime:
			sys.bcStack.Push(BytecodeInt(c.animTime()))
		case OC_animelemtime:
			*sys.bcStack.Top() = BytecodeInt(c.animElemTime(sys.bcStack.Top().ToI()))
		case OC_animexist:
			*sys.bcStack.Top() = c.animExist(sys.workingChar, *sys.bcStack.Top())
		case OC_selfanimexist:
			*sys.bcStack.Top() = c.selfAnimExist(*sys.bcStack.Top())
		case OC_stateno:
			sys.bcStack.Push(BytecodeInt(c.ss.no))
		case OC_prevstateno:
			sys.bcStack.Push(BytecodeInt(c.ss.prevno))
		case OC_movecontact:
			sys.bcStack.Push(BytecodeInt(c.moveContact()))
		case OC_movehit:
			sys.bcStack.Push(BytecodeInt(c.moveHit()))
		case OC_moveguarded:
			sys.bcStack.Push(BytecodeInt(c.moveGuarded()))
		case OC_movereversed:
			sys.bcStack.Push(BytecodeInt(c.moveReversed()))
		case OC_vel_x:
			sys.bcStack.Push(BytecodeFloat(c.vel[0]))
		case OC_vel_y:
			sys.bcStack.Push(BytecodeFloat(c.vel[1]))
		case OC_pos_x:
			sys.bcStack.Push(BytecodeFloat(c.pos[0] - sys.cameraPos[0]))
		case OC_pos_y:
			sys.bcStack.Push(BytecodeFloat(c.pos[1]))
		case OC_screenpos_x:
			sys.bcStack.Push(BytecodeFloat(c.screenPosX()))
		case OC_screenpos_y:
			sys.bcStack.Push(BytecodeFloat(c.screenPosY()))
		case OC_canrecover:
			sys.bcStack.Push(BytecodeBool(c.canRecover()))
		case OC_hitshakeover:
			sys.bcStack.Push(BytecodeBool(c.hitShakeOver()))
		case OC_matchover:
			sys.bcStack.Push(BytecodeBool(sys.matchOver()))
		case OC_frontedgedist:
			sys.bcStack.Push(BytecodeInt(c.frontEdgeDist()))
		case OC_frontedgebodydist:
			sys.bcStack.Push(BytecodeInt(c.frontEdgeBodyDist()))
		case OC_frontedge:
			sys.bcStack.Push(BytecodeFloat(c.frontEdge()))
		case OC_backedgedist:
			sys.bcStack.Push(BytecodeInt(c.backEdgeDist()))
		case OC_backedgebodydist:
			sys.bcStack.Push(BytecodeInt(c.backEdgeBodyDist()))
		case OC_backedge:
			sys.bcStack.Push(BytecodeFloat(c.backEdge()))
		case OC_leftedge:
			sys.bcStack.Push(BytecodeFloat(c.leftEdge()))
		case OC_rightedge:
			sys.bcStack.Push(BytecodeFloat(c.rightEdge()))
		case OC_topedge:
			sys.bcStack.Push(BytecodeFloat(c.topEdge()))
		case OC_bottomedge:
			sys.bcStack.Push(BytecodeFloat(c.bottomEdge()))
		case OC_st_:
			be.run_st(c, &i)
		case OC_ex_:
			be.run_ex(c, &i)
		case OC_var:
			*sys.bcStack.Top() = c.varGet(sys.bcStack.Top().ToI())
		case OC_sysvar:
			*sys.bcStack.Top() = c.sysVarGet(sys.bcStack.Top().ToI())
		case OC_fvar:
			*sys.bcStack.Top() = c.fvarGet(sys.bcStack.Top().ToI())
		case OC_sysfvar:
			*sys.bcStack.Top() = c.sysFvarGet(sys.bcStack.Top().ToI())
		default:
			vi := be[i-1]
			if vi < OC_sysvar0+NumSysVar {
				sys.bcStack.Push(BytecodeInt(c.ivar[vi-OC_var0]))
			} else if vi < OC_sysfvar0+NumSysFvar {
				sys.bcStack.Push(BytecodeFloat(c.fvar[vi-OC_fvar0]))
			} else {
				println(be[i-1])
				unimplemented()
			}
		}
		c = oc
	}
	return sys.bcStack.Pop()
}
func (be BytecodeExp) run_st(c *Char, i *int) {
	(*i)++
	switch be[*i-1] {
	case OC_st_var:
		v := sys.bcStack.Pop().ToI()
		*sys.bcStack.Top() = c.varSet(sys.bcStack.Top().ToI(), v)
	case OC_st_sysvar:
		v := sys.bcStack.Pop().ToI()
		*sys.bcStack.Top() = c.sysVarSet(sys.bcStack.Top().ToI(), v)
	case OC_st_fvar:
		v := sys.bcStack.Pop().ToF()
		*sys.bcStack.Top() = c.fvarSet(sys.bcStack.Top().ToI(), v)
	case OC_st_sysfvar:
		v := sys.bcStack.Pop().ToF()
		*sys.bcStack.Top() = c.sysFvarSet(sys.bcStack.Top().ToI(), v)
	case OC_st_varadd:
		v := sys.bcStack.Pop().ToI()
		*sys.bcStack.Top() = c.varAdd(sys.bcStack.Top().ToI(), v)
	case OC_st_sysvaradd:
		v := sys.bcStack.Pop().ToI()
		*sys.bcStack.Top() = c.sysVarAdd(sys.bcStack.Top().ToI(), v)
	case OC_st_fvaradd:
		v := sys.bcStack.Pop().ToF()
		*sys.bcStack.Top() = c.fvarAdd(sys.bcStack.Top().ToI(), v)
	case OC_st_sysfvaradd:
		v := sys.bcStack.Pop().ToF()
		*sys.bcStack.Top() = c.sysFvarAdd(sys.bcStack.Top().ToI(), v)
	default:
		vi := be[*i-1]
		if vi < OC_st_sysvar0+NumSysVar {
			c.ivar[vi-OC_st_var0] = sys.bcStack.Top().ToI()
			sys.bcStack.Top().SetI(c.ivar[vi-OC_st_var0])
		} else if vi < OC_st_sysfvar0+NumSysFvar {
			c.fvar[vi-OC_st_fvar0] = sys.bcStack.Top().ToF()
			sys.bcStack.Top().SetF(c.fvar[vi-OC_st_fvar0])
		} else if vi < OC_st_sysvar0add+NumSysVar {
			c.ivar[vi-OC_st_var0add] += sys.bcStack.Top().ToI()
			sys.bcStack.Top().SetI(c.ivar[vi-OC_st_var0add])
		} else if vi < OC_st_sysfvar0add+NumSysFvar {
			c.fvar[vi-OC_st_fvar0add] += sys.bcStack.Top().ToF()
			sys.bcStack.Top().SetF(c.fvar[vi-OC_st_fvar0add])
		} else {
			println(be[*i-1])
			unimplemented()
		}
	}
}
func (be BytecodeExp) run_ex(c *Char, i *int) {
	(*i)++
	switch be[*i-1] {
	case OC_ex_p2dist_x:
		sys.bcStack.Push(c.p2DistX())
	case OC_ex_p2dist_y:
		sys.bcStack.Push(c.p2DistY())
	case OC_ex_p2bodydist_x:
		sys.bcStack.Push(c.p2BodyDistX())
	case OC_ex_p2bodydist_y:
		sys.bcStack.Push(c.p2BodyDistY())
	case OC_ex_rootdist_x:
		sys.bcStack.Push(c.rootDistX())
	case OC_ex_rootdist_y:
		sys.bcStack.Push(c.rootDistY())
	case OC_ex_parentdist_x:
		sys.bcStack.Push(c.parentDistX())
	case OC_ex_parentdist_y:
		sys.bcStack.Push(c.parentDistY())
	case OC_ex_gethitvar_animtype:
		sys.bcStack.Push(BytecodeInt(int32(c.gethitAnimtype())))
	case OC_ex_gethitvar_airtype:
		sys.bcStack.Push(BytecodeInt(int32(c.ghv.airtype)))
	case OC_ex_gethitvar_groundtype:
		sys.bcStack.Push(BytecodeInt(int32(c.ghv.groundtype)))
	case OC_ex_gethitvar_damage:
		sys.bcStack.Push(BytecodeInt(c.ghv.damage))
	case OC_ex_gethitvar_hitcount:
		sys.bcStack.Push(BytecodeInt(c.ghv.hitcount))
	case OC_ex_gethitvar_fallcount:
		sys.bcStack.Push(BytecodeInt(c.ghv.fallcount))
	case OC_ex_gethitvar_hitshaketime:
		sys.bcStack.Push(BytecodeInt(c.ghv.hitshaketime))
	case OC_ex_gethitvar_hittime:
		sys.bcStack.Push(BytecodeInt(c.ghv.hittime))
	case OC_ex_gethitvar_slidetime:
		sys.bcStack.Push(BytecodeInt(c.ghv.slidetime))
	case OC_ex_gethitvar_ctrltime:
		sys.bcStack.Push(BytecodeInt(c.ghv.ctrltime))
	case OC_ex_gethitvar_recovertime:
		sys.bcStack.Push(BytecodeInt(c.recovertime))
	case OC_ex_gethitvar_xoff:
		sys.bcStack.Push(BytecodeFloat(c.ghv.xoff))
	case OC_ex_gethitvar_yoff:
		sys.bcStack.Push(BytecodeFloat(c.ghv.yoff))
	case OC_ex_gethitvar_xvel:
		sys.bcStack.Push(BytecodeFloat(c.ghv.xvel))
	case OC_ex_gethitvar_yvel:
		sys.bcStack.Push(BytecodeFloat(c.ghv.yvel))
	case OC_ex_gethitvar_yaccel:
		sys.bcStack.Push(BytecodeFloat(c.ghv.getYaccel()))
	case OC_ex_gethitvar_chainid:
		sys.bcStack.Push(BytecodeInt(c.ghv.chainId()))
	case OC_ex_gethitvar_guarded:
		sys.bcStack.Push(BytecodeBool(c.ghv.guarded))
	case OC_ex_gethitvar_isbound:
		sys.bcStack.Push(BytecodeBool(c.isBound()))
	case OC_ex_gethitvar_fall:
		sys.bcStack.Push(BytecodeBool(c.ghv.fallf))
	case OC_ex_gethitvar_fall_damage:
		sys.bcStack.Push(BytecodeInt(c.ghv.fall.damage))
	case OC_ex_gethitvar_fall_xvel:
		sys.bcStack.Push(BytecodeFloat(c.ghv.fall.xvel()))
	case OC_ex_gethitvar_fall_yvel:
		sys.bcStack.Push(BytecodeFloat(c.ghv.fall.yvelocity))
	case OC_ex_gethitvar_fall_recover:
		sys.bcStack.Push(BytecodeBool(c.ghv.fall.recover))
	case OC_ex_gethitvar_fall_time:
		sys.bcStack.Push(BytecodeInt(c.fallTime))
	case OC_ex_gethitvar_fall_recovertime:
		sys.bcStack.Push(BytecodeInt(c.ghv.fall.recovertime))
	case OC_ex_gethitvar_fall_kill:
		sys.bcStack.Push(BytecodeBool(c.ghv.fall.kill))
	case OC_ex_gethitvar_fall_envshake_time:
		sys.bcStack.Push(BytecodeInt(c.ghv.fall.envshake_time))
	case OC_ex_gethitvar_fall_envshake_freq:
		sys.bcStack.Push(BytecodeFloat(c.ghv.fall.envshake_freq))
	case OC_ex_gethitvar_fall_envshake_ampl:
		sys.bcStack.Push(BytecodeInt(c.ghv.fall.envshake_ampl))
	case OC_ex_gethitvar_fall_envshake_phase:
		sys.bcStack.Push(BytecodeFloat(c.ghv.fall.envshake_phase))
	default:
		println(be[*i-1])
		unimplemented()
	}
}
func (be BytecodeExp) evalF(c *Char) float32 {
	return be.run(c).ToF()
}
func (be BytecodeExp) evalI(c *Char) int32 {
	return be.run(c).ToI()
}
func (be BytecodeExp) evalB(c *Char) bool {
	return be.run(c).ToB()
}

type StateController interface {
	Run(c *Char, ps []int32) (changeState bool)
}

const SCID_trigger byte = 255

type StateBlock struct {
	persistent      int32
	persistentIndex int32
	ignorehitpause  bool
	trigger         BytecodeExp
	ctrls           []StateController
}

func newStateBlock() *StateBlock {
	return &StateBlock{persistent: 1, persistentIndex: -1}
}
func (b StateBlock) Run(c *Char, ps []int32) (changeState bool) {
	if !b.ignorehitpause && c.hitPause() {
		return false
	}
	if b.persistentIndex >= 0 {
		ps[b.persistentIndex]--
		if ps[b.persistentIndex] > 0 {
			return false
		}
	}
	sys.workingChar = c
	if len(b.trigger) > 0 && !b.trigger.evalB(c) {
		return false
	}
	for _, sc := range b.ctrls {
		if sc.Run(c, ps) {
			return true
		}
	}
	if b.persistentIndex >= 0 {
		ps[b.persistentIndex] = b.persistent
	}
	return false
}

type StateControllerBase []byte

func newStateControllerBase() *StateControllerBase {
	return (*StateControllerBase)(&[]byte{})
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
	*scb = append(*scb, id, byte(len(exp)))
	for _, e := range exp {
		l := int32(len(e))
		*scb = append(*scb, (*(*[4]byte)(unsafe.Pointer(&l)))[:]...)
		*scb = append(*scb, (*(*[]byte)(unsafe.Pointer(&e)))...)
	}
}
func (scb StateControllerBase) run(c *Char,
	f func(byte, []BytecodeExp) bool) {
	for i := 0; i < len(scb); {
		id := scb[i]
		i++
		n := scb[i]
		i++
		exp := make([]BytecodeExp, n)
		for m := 0; m < int(n); m++ {
			l := *(*int32)(unsafe.Pointer(&scb[i]))
			i += 4
			exp[m] = (*(*BytecodeExp)(unsafe.Pointer(&scb)))[i : i+int(l)]
			i += int(l)
		}
		if !f(id, exp) {
			break
		}
	}
	if len(sys.bcStack) != 0 {
		unimplemented()
	}
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

func (sc stateDef) Run(c *Char, _ []int32) bool {
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case stateDef_hitcountpersist:
			if !exp[0].evalB(c) {
				c.clearHitCount()
			}
		case stateDef_movehitpersist:
			if !exp[0].evalB(c) {
				c.clearMoveHit()
			}
		case stateDef_hitdefpersist:
			if !exp[0].evalB(c) {
				c.clearHitDef()
			}
		case stateDef_sprpriority:
			c.setSprPriority(exp[0].evalI(c))
		case stateDef_facep2:
			if exp[0].evalB(c) {
				c.faceP2()
			}
		case stateDef_juggle:
			c.setJuggle(exp[0].evalI(c))
		case stateDef_velset:
			c.setXV(exp[0].evalF(c))
			if len(exp) > 1 {
				c.setYV(exp[1].evalF(c))
				if len(exp) > 2 {
					exp[2].run(c)
				}
			}
		case stateDef_anim:
			c.changeAnim(exp[0].evalI(c))
		case stateDef_ctrl:
			c.setCtrl(exp[0].evalB(c))
		case stateDef_poweradd:
			c.powerAdd(exp[0].evalI(c))
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

func (sc hitBy) Run(c *Char, _ []int32) bool {
	time := int32(1)
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case hitBy_time:
			time = exp[0].evalI(c)
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

func (sc notHitBy) Run(c *Char, _ []int32) bool {
	time := int32(1)
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case hitBy_time:
			time = exp[0].evalI(c)
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

func (sc assertSpecial) Run(c *Char, _ []int32) bool {
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case assertSpecial_flag:
			c.setSF(CharSpecialFlag(exp[0].evalI(c)))
		case assertSpecial_flag_g:
			sys.setSF(GlobalSpecialFlag(exp[0].evalI(c)))
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

func (sc playSnd) Run(c *Char, _ []int32) bool {
	f, lw, lp := false, false, false
	var g, n, ch, vo int32 = -1, 0, -1, 0
	if sys.cgi[sys.workingChar.ss.sb.playerNo].ver[0] == 1 {
		vo = 100
	}
	var p, fr float32 = 0, 1
	x := &c.pos[0]
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case playSnd_value:
			f = exp[0].evalB(c)
			g = exp[1].evalI(c)
			if len(exp) > 2 {
				n = exp[2].evalI(c)
			}
		case playSnd_channel:
			ch = exp[0].evalI(c)
		case playSnd_lowpriority:
			lw = exp[0].evalB(c)
		case playSnd_pan:
			p = exp[0].evalF(c)
		case playSnd_abspan:
			x = nil
			p = exp[0].evalF(c)
		case playSnd_volume:
			vo = exp[0].evalI(c)
		case playSnd_freqmul:
			fr = exp[0].evalF(c)
		case playSnd_loop:
			lp = exp[0].evalB(c)
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

func (sc changeState) Run(c *Char, _ []int32) bool {
	var v, a, ctrl int32 = -1, -1, -1
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case changeState_value:
			v = exp[0].evalI(c)
		case changeState_ctrl:
			ctrl = exp[0].evalI(c)
		case changeState_anim:
			a = exp[0].evalI(c)
		}
		return true
	})
	c.changeState(v, a, ctrl)
	return true
}

type selfState changeState

func (sc selfState) Run(c *Char, _ []int32) bool {
	var v, a, ctrl int32 = -1, -1, -1
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case changeState_value:
			v = exp[0].evalI(c)
		case changeState_ctrl:
			ctrl = exp[0].evalI(c)
		case changeState_anim:
			a = exp[0].evalI(c)
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

func (sc tagIn) Run(c *Char, _ []int32) bool {
	var p *Char
	sn := int32(-1)
	ret := false
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		if p == nil {
			p = c.partner(0)
			if p == nil {
				return false
			}
		}
		switch id {
		case tagIn_stateno:
			sn = exp[0].evalI(c)
		case tagIn_partnerstateno:
			if psn := exp[0].evalI(c); psn >= 0 {
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

func (sc tagOut) Run(c *Char, _ []int32) bool {
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
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

func (sc destroySelf) Run(c *Char, _ []int32) bool {
	rec, rem := false, false
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case destroySelf_recursive:
			rec = exp[0].evalB(c)
		case destroySelf_removeexplods:
			rem = exp[0].evalB(c)
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

func (sc changeAnim) Run(c *Char, _ []int32) bool {
	var elem int32
	setelem := false
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case changeAnim_elem:
			elem = exp[0].evalI(c)
			setelem = true
		case changeAnim_value:
			c.changeAnim(exp[0].evalI(c))
			if setelem {
				c.setAnimElem(elem)
			}
		}
		return true
	})
	return false
}

type changeAnim2 changeAnim

func (sc changeAnim2) Run(c *Char, _ []int32) bool {
	var elem int32
	setelem := false
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case changeAnim_elem:
			elem = exp[0].evalI(c)
			setelem = true
		case changeAnim_value:
			c.changeAnim2(exp[0].evalI(c))
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
	helper_postype
	helper_ownpal
	helper_size_xscale
	helper_size_yscale
	helper_size_ground_back
	helper_size_ground_front
	helper_size_air_back
	helper_size_air_front
	helper_size_height
	helper_size_proj_doscale
	helper_size_head_pos
	helper_size_mid_pos
	helper_size_shadowoffset
	helper_stateno
	helper_keyctrl
	helper_id
	helper_pos
	helper_facing
	helper_pausemovetime
	helper_supermovetime
)

func (sc helper) Run(c *Char, _ []int32) bool {
	var h *Char
	pt := PT_P1
	var f, st int32 = 0, 1
	op := false
	var x, y float32 = 0, 0
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		if h == nil {
			h = c.newHelper()
			if h == nil {
				return false
			}
		}
		switch id {
		case helper_helpertype:
			h.player = exp[0].evalB(c)
		case helper_name:
			h.name = string(*(*[]byte)(unsafe.Pointer(&exp[0])))
		case helper_postype:
			pt = PosType(exp[0].evalI(c))
		case helper_ownpal:
			op = exp[0].evalB(c)
		case helper_size_xscale:
			h.size.xscale = exp[0].evalF(c)
		case helper_size_yscale:
			h.size.yscale = exp[0].evalF(c)
		case helper_size_ground_back:
			h.size.ground.back = exp[0].evalI(c)
		case helper_size_ground_front:
			h.size.ground.front = exp[0].evalI(c)
		case helper_size_air_back:
			h.size.air.back = exp[0].evalI(c)
		case helper_size_air_front:
			h.size.air.front = exp[0].evalI(c)
		case helper_size_height:
			h.size.height = exp[0].evalI(c)
		case helper_size_proj_doscale:
			h.size.proj.doscale = exp[0].evalI(c)
		case helper_size_head_pos:
			h.size.head.pos[0] = exp[0].evalI(c)
			if len(exp) > 1 {
				h.size.head.pos[1] = exp[1].evalI(c)
			}
		case helper_size_mid_pos:
			h.size.mid.pos[0] = exp[0].evalI(c)
			if len(exp) > 1 {
				h.size.mid.pos[1] = exp[1].evalI(c)
			}
		case helper_size_shadowoffset:
			h.size.shadowoffset = exp[0].evalI(c)
		case helper_stateno:
			st = exp[0].evalI(c)
		case helper_keyctrl:
			h.keyctrl = exp[0].evalB(c)
		case helper_id:
			h.helperId = exp[0].evalI(c)
		case helper_pos:
			x = exp[0].evalF(c)
			if len(exp) > 1 {
				y = exp[1].evalF(c)
			}
		case helper_facing:
			f = exp[0].evalI(c)
		case helper_pausemovetime:
			h.pauseMovetime = exp[0].evalI(c)
		case helper_supermovetime:
			h.superMovetime = exp[0].evalI(c)
		}
		return true
	})
	if h != nil {
		c.helperInit(h, st, pt, x, y, f, op)
	}
	return false
}

type ctrlSet StateControllerBase

const (
	ctrlSet_value byte = iota
)

func (sc ctrlSet) Run(c *Char, _ []int32) bool {
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case ctrlSet_value:
			c.setCtrl(exp[0].evalB(c))
		}
		return true
	})
	return false
}

type explod StateControllerBase

const (
	explod_ownpal byte = iota
	explod_remappal
	explod_id
	explod_facing
	explod_vfacing
	explod_pos
	explod_random
	explod_postype
	explod_velocity
	explod_accel
	explod_scale
	explod_bindtime
	explod_removetime
	explod_supermove
	explod_supermovetime
	explod_pausemovetime
	explod_sprpriority
	explod_ontop
	explod_shadow
	explod_removeongethit
	explod_trans
	explod_anim
	explod_angle
	explod_yangle
	explod_xangle
	explod_ignorehitpause
)

func (sc explod) Run(c *Char, _ []int32) bool {
	var e *Explod
	var i int
	rp := [2]int32{-1, 0}
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		if e == nil {
			e, i = c.newExplod()
			if e == nil {
				return false
			}
		}
		switch id {
		case explod_ownpal:
			e.ownpal = exp[0].evalB(c)
		case explod_remappal:
			rp[0] = exp[0].evalI(c)
			if len(exp) > 1 {
				rp[1] = exp[1].evalI(c)
			}
		case explod_id:
			e.id = Max(0, exp[0].evalI(c))
		case explod_facing:
			if exp[0].evalI(c) < 0 {
				e.relativef = -1
			} else {
				e.relativef = 1
			}
		case explod_vfacing:
			if exp[0].evalI(c) < 0 {
				e.vfacing = -1
			} else {
				e.vfacing = 1
			}
		case explod_pos:
			e.offset[0] = exp[0].evalF(c)
			if len(exp) > 1 {
				e.offset[1] = exp[1].evalF(c)
			}
		case explod_random:
			rndx := exp[0].evalF(c)
			e.offset[0] += RandF(-rndx, rndx)
			if len(exp) > 1 {
				rndy := exp[1].evalF(c)
				e.offset[1] += RandF(-rndy, rndy)
			}
		case explod_postype:
			e.postype = PosType(exp[0].evalI(c))
		case explod_velocity:
			e.velocity[0] = exp[0].evalF(c)
			if len(exp) > 1 {
				e.velocity[1] = exp[1].evalF(c)
			}
		case explod_accel:
			e.accel[0] = exp[0].evalF(c)
			if len(exp) > 1 {
				e.accel[1] = exp[1].evalF(c)
			}
		case explod_scale:
			e.scale[0] = exp[0].evalF(c)
			if len(exp) > 1 {
				e.scale[1] = exp[1].evalF(c)
			}
		case explod_bindtime:
			e.bindtime = exp[0].evalI(c)
		case explod_removetime:
			e.removetime = exp[0].evalI(c)
		case explod_supermove:
			if exp[0].evalB(c) {
				e.supermovetime = -1
			} else {
				e.supermovetime = 0
			}
		case explod_supermovetime:
			e.supermovetime = exp[0].evalI(c)
		case explod_pausemovetime:
			e.pausemovetime = exp[0].evalI(c)
		case explod_sprpriority:
			e.sprpriority = exp[0].evalI(c)
		case explod_ontop:
			e.ontop = exp[0].evalB(c)
			if e.ontop {
				e.sprpriority = 0
			}
		case explod_shadow:
			e.shadow[0] = exp[0].evalI(c)
			if len(exp) > 1 {
				e.shadow[1] = exp[1].evalI(c)
				if len(exp) > 2 {
					e.shadow[2] = exp[2].evalI(c)
				}
			}
		case explod_removeongethit:
			e.removeongethit = exp[0].evalB(c)
		case explod_trans:
			e.alpha = [2]int32{exp[0].evalI(c), exp[1].evalI(c)}
			if len(exp) >= 3 {
				e.alpha[0] = Max(0, Min(255, e.alpha[0]))
				e.alpha[1] = Max(0, Min(255, e.alpha[1]))
				if len(exp) >= 4 {
					e.alpha[1] = ^e.alpha[1]
				}
			}
		case explod_anim:
			e.anim = c.getAnim(exp[1].evalI(c), exp[0].evalB(c))
		case explod_angle:
			e.angle = exp[0].evalF(c)
		case explod_yangle:
			exp[0].run(c)
		case explod_xangle:
			exp[0].run(c)
		case explod_ignorehitpause:
			e.ignorehitpause = exp[0].evalB(c)
		}
		return true
	})
	if e != nil {
		e.setPos(c)
		c.insertExplodEx(i, rp[0], rp[1])
	}
	return false
}

type modifyExplod explod

func (sc modifyExplod) Run(c *Char, _ []int32) bool {
	eid := int32(-1)
	var expls []*Explod
	rp := [2]int32{-1, 0}
	eachExpl := func(f func(e *Explod)) {
		for _, e := range expls {
			f(e)
		}
	}
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case explod_remappal:
			rp[0] = exp[0].evalI(c)
			if len(exp) > 1 {
				rp[1] = exp[1].evalI(c)
			}
		case explod_id:
			eid = exp[0].evalI(c)
		default:
			if len(expls) == 0 {
				expls = c.getExplods(eid)
				if len(expls) == 0 {
					return false
				}
				eachExpl(func(e *Explod) {
					if e.ownpal {
						c.remapPalSub(e.palfx, 1, 1, rp[0], rp[1])
					}
				})
			}
			switch id {
			case explod_facing:
				if exp[0].evalI(c) < 0 {
					eachExpl(func(e *Explod) { e.relativef = -1 })
				} else {
					eachExpl(func(e *Explod) { e.relativef = 1 })
				}
			case explod_vfacing:
				if exp[0].evalI(c) < 0 {
					eachExpl(func(e *Explod) { e.vfacing = -1 })
				} else {
					eachExpl(func(e *Explod) { e.vfacing = 1 })
				}
			case explod_pos:
				x := exp[0].evalF(c)
				eachExpl(func(e *Explod) { e.offset[0] = x })
				if len(exp) > 1 {
					y := exp[1].evalF(c)
					eachExpl(func(e *Explod) { e.offset[1] = y })
				}
			case explod_random:
				rndx := exp[0].evalF(c)
				rndx = RandF(-rndx, rndx)
				eachExpl(func(e *Explod) { e.offset[0] += rndx })
				if len(exp) > 1 {
					rndy := exp[1].evalF(c)
					rndy = RandF(-rndy, rndy)
					eachExpl(func(e *Explod) { e.offset[1] += rndy })
				}
			case explod_postype:
				pt := PosType(exp[0].evalI(c))
				eachExpl(func(e *Explod) {
					e.postype = pt
					e.setPos(c)
				})
			case explod_velocity:
				x := exp[0].evalF(c)
				eachExpl(func(e *Explod) { e.velocity[0] = x })
				if len(exp) > 1 {
					y := exp[1].evalF(c)
					eachExpl(func(e *Explod) { e.velocity[1] = y })
				}
			case explod_accel:
				x := exp[0].evalF(c)
				eachExpl(func(e *Explod) { e.accel[0] = x })
				if len(exp) > 1 {
					y := exp[1].evalF(c)
					eachExpl(func(e *Explod) { e.accel[1] = y })
				}
			case explod_scale:
				x := exp[0].evalF(c)
				eachExpl(func(e *Explod) { e.scale[0] = x })
				if len(exp) > 1 {
					y := exp[1].evalF(c)
					eachExpl(func(e *Explod) { e.scale[1] = y })
				}
			case explod_bindtime:
				t := exp[0].evalI(c)
				eachExpl(func(e *Explod) { e.bindtime = t })
			case explod_removetime:
				t := exp[0].evalI(c)
				eachExpl(func(e *Explod) { e.removetime = t })
			case explod_supermove:
				if exp[0].evalB(c) {
					eachExpl(func(e *Explod) { e.supermovetime = -1 })
				} else {
					eachExpl(func(e *Explod) { e.supermovetime = 0 })
				}
			case explod_supermovetime:
				t := exp[0].evalI(c)
				eachExpl(func(e *Explod) { e.supermovetime = t })
			case explod_pausemovetime:
				t := exp[0].evalI(c)
				eachExpl(func(e *Explod) { e.pausemovetime = t })
			case explod_sprpriority:
				t := exp[0].evalI(c)
				eachExpl(func(e *Explod) { e.sprpriority = t })
			case explod_ontop:
				t := exp[0].evalB(c)
				eachExpl(func(e *Explod) {
					e.ontop = t
					if e.ontop {
						e.sprpriority = 0
					}
				})
			case explod_shadow:
				r := exp[0].evalI(c)
				eachExpl(func(e *Explod) { e.shadow[0] = r })
				if len(exp) > 1 {
					g := exp[1].evalI(c)
					eachExpl(func(e *Explod) { e.shadow[1] = g })
					if len(exp) > 2 {
						b := exp[2].evalI(c)
						eachExpl(func(e *Explod) { e.shadow[2] = b })
					}
				}
			case explod_removeongethit:
				t := exp[0].evalB(c)
				eachExpl(func(e *Explod) { e.removeongethit = t })
			case explod_trans:
				s, d := exp[0].evalI(c), exp[1].evalI(c)
				if len(exp) > 2 {
					s = Max(0, Min(255, s))
					d = Max(0, Min(255, d))
				}
				eachExpl(func(e *Explod) { e.alpha = [2]int32{s, d} })
			case explod_angle:
				a := exp[0].evalF(c)
				eachExpl(func(e *Explod) { e.angle = a })
			case explod_yangle:
				exp[0].run(c)
			case explod_xangle:
				exp[0].run(c)
			}
		}
		return true
	})
	return false
}

type gameMakeAnim StateControllerBase

const (
	gameMakeAnim_pos byte = iota
	gameMakeAnim_random
	gameMakeAnim_under
	gameMakeAnim_anim
)

func (sc gameMakeAnim) Run(c *Char, _ []int32) bool {
	var e *Explod
	var i int
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		if e == nil {
			e, i = c.newExplod()
			if e == nil {
				return false
			}
			e.ontop, e.sprpriority, e.ownpal = true, math.MinInt32, true
		}
		switch id {
		case gameMakeAnim_pos:
			e.offset[0] = exp[0].evalF(c)
			if len(exp) > 1 {
				e.offset[1] = exp[1].evalF(c)
			}
		case gameMakeAnim_random:
			rndx := exp[0].evalF(c)
			e.offset[0] += RandF(-rndx, rndx)
			if len(exp) > 1 {
				rndy := exp[1].evalF(c)
				e.offset[1] += RandF(-rndy, rndy)
			}
		case gameMakeAnim_under:
			e.ontop = !exp[0].evalB(c)
		case gameMakeAnim_anim:
			e.anim = c.getAnim(exp[1].evalI(c), exp[0].evalB(c))
		}
		return true
	})
	if e != nil {
		e.offset[0] -= float32(c.size.draw.offset[0])
		e.offset[1] -= float32(c.size.draw.offset[1])
		e.setPos(c)
		c.insertExplod(i)
	}
	return false
}

type posSet StateControllerBase

const (
	posSet_x byte = iota
	posSet_y
	posSet_z
)

func (sc posSet) Run(c *Char, _ []int32) bool {
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case posSet_x:
			c.setX(exp[0].evalF(c))
		case posSet_y:
			c.setY(exp[0].evalF(c))
		case posSet_z:
			exp[0].run(c)
		}
		return true
	})
	return false
}

type posAdd posSet

func (sc posAdd) Run(c *Char, _ []int32) bool {
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case posSet_x:
			c.addX(exp[0].evalF(c))
		case posSet_y:
			c.addY(exp[0].evalF(c))
		case posSet_z:
			exp[0].run(c)
		}
		return true
	})
	return false
}

type velSet posSet

func (sc velSet) Run(c *Char, _ []int32) bool {
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case posSet_x:
			c.setXV(exp[0].evalF(c))
		case posSet_y:
			c.setYV(exp[0].evalF(c))
		case posSet_z:
			exp[0].run(c)
		}
		return true
	})
	return false
}

type velAdd posSet

func (sc velAdd) Run(c *Char, _ []int32) bool {
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case posSet_x:
			c.addXV(exp[0].evalF(c))
		case posSet_y:
			c.addYV(exp[0].evalF(c))
		case posSet_z:
			exp[0].run(c)
		}
		return true
	})
	return false
}

type velMul posSet

func (sc velMul) Run(c *Char, _ []int32) bool {
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case posSet_x:
			c.mulXV(exp[0].evalF(c))
		case posSet_y:
			c.mulYV(exp[0].evalF(c))
		case posSet_z:
			exp[0].run(c)
		}
		return true
	})
	return false
}

type palFX StateControllerBase

const (
	palFX_time byte = iota
	palFX_color
	palFX_add
	palFX_mul
	palFX_sinadd
	palFX_invertall
	palFX_last byte = iota - 1
)

func (sc palFX) runSub(c *Char, pfd *PalFXDef,
	id byte, exp []BytecodeExp) bool {
	switch id {
	case palFX_time:
		pfd.time = exp[0].evalI(c)
	case palFX_color:
		pfd.color = MaxF(0, MinF(1, exp[0].evalF(c)/256))
	case palFX_add:
		pfd.add = [3]int32{exp[0].evalI(c), exp[1].evalI(c), exp[2].evalI(c)}
	case palFX_mul:
		pfd.mul = [3]int32{exp[0].evalI(c), exp[1].evalI(c), exp[2].evalI(c)}
	case palFX_sinadd:
		pfd.sinadd = [3]int32{exp[0].evalI(c), exp[1].evalI(c), exp[2].evalI(c)}
		if len(exp) > 3 {
			pfd.cycletime = exp[3].evalI(c)
		}
	case palFX_invertall:
		pfd.invertall = exp[0].evalB(c)
	default:
		return false
	}
	return true
}
func (sc palFX) Run(c *Char, _ []int32) bool {
	pf := c.palfx
	if pf == nil {
		pf = NewPalFX()
	}
	pf.clear2(true)
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		sc.runSub(c, &pf.def, id, exp)
		return true
	})
	return false
}

type allPalFX palFX

func (sc allPalFX) Run(c *Char, _ []int32) bool {
	sys.allPalFX.clear()
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		palFX(sc).runSub(c, &sys.allPalFX.def, id, exp)
		return true
	})
	return false
}

type bgPalFX palFX

func (sc bgPalFX) Run(c *Char, _ []int32) bool {
	sys.bgPalFX.clear()
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		palFX(sc).runSub(c, &sys.bgPalFX.def, id, exp)
		return true
	})
	return false
}

type afterImage palFX

const (
	afterImage_trans byte = iota + palFX_last + 1
	afterImage_time
	afterImage_length
	afterImage_timegap
	afterImage_framegap
	afterImage_palcolor
	afterImage_palinvertall
	afterImage_palbright
	afterImage_palcontrast
	afterImage_palpostbright
	afterImage_paladd
	afterImage_palmul
	afterImage_last byte = iota - 1
)

func (sc afterImage) runSub(c *Char, ai *AfterImage,
	id byte, exp []BytecodeExp) {
	switch id {
	case afterImage_trans:
		ai.alpha = [2]int32{exp[0].evalI(c), exp[1].evalI(c)}
	case afterImage_time:
		ai.time = exp[0].evalI(c)
	case afterImage_length:
		ai.length = exp[0].evalI(c)
	case afterImage_timegap:
		ai.timegap = Max(1, exp[0].evalI(c))
	case afterImage_framegap:
		ai.framegap = exp[0].evalI(c)
	case afterImage_palcolor:
		ai.setPalColor(exp[0].evalI(c))
	case afterImage_palinvertall:
		ai.setPalInvertall(exp[0].evalB(c))
	case afterImage_palbright:
		ai.setPalBrightR(exp[0].evalI(c))
		if len(exp) > 1 {
			ai.setPalBrightG(exp[1].evalI(c))
			if len(exp) > 2 {
				ai.setPalBrightB(exp[2].evalI(c))
			}
		}
	case afterImage_palcontrast:
		ai.setPalContrastR(exp[0].evalI(c))
		if len(exp) > 1 {
			ai.setPalContrastG(exp[1].evalI(c))
			if len(exp) > 2 {
				ai.setPalContrastB(exp[2].evalI(c))
			}
		}
	case afterImage_palpostbright:
		ai.postbright[0] = exp[0].evalI(c)
		if len(exp) > 1 {
			ai.postbright[1] = exp[1].evalI(c)
			if len(exp) > 2 {
				ai.postbright[2] = exp[2].evalI(c)
			}
		}
	case afterImage_paladd:
		ai.add[0] = exp[0].evalI(c)
		if len(exp) > 1 {
			ai.add[1] = exp[1].evalI(c)
			if len(exp) > 2 {
				ai.add[2] = exp[2].evalI(c)
			}
		}
	case afterImage_palmul:
		ai.mul[0] = exp[0].evalF(c)
		if len(exp) > 1 {
			ai.mul[1] = exp[1].evalF(c)
			if len(exp) > 2 {
				ai.mul[2] = exp[2].evalF(c)
			}
		}
	}
}
func (sc afterImage) Run(c *Char, _ []int32) bool {
	c.aimg.clear()
	c.aimg.time = 1
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		sc.runSub(c, &c.aimg, id, exp)
		return true
	})
	c.aimg.setupPalFX()
	return false
}

type afterImageTime StateControllerBase

const (
	afterImageTime_time byte = iota
)

func (sc afterImageTime) Run(c *Char, _ []int32) bool {
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		if c.aimg.timegap <= 0 {
			return false
		}
		switch id {
		case afterImageTime_time:
			c.aimg.time = exp[0].evalI(c)
		}
		return true
	})
	return false
}

type hitDef afterImage

const (
	hitDef_attr byte = iota + afterImage_last + 1
	hitDef_guardflag
	hitDef_hitflag
	hitDef_ground_type
	hitDef_air_type
	hitDef_animtype
	hitDef_air_animtype
	hitDef_fall_animtype
	hitDef_affectteam
	hitDef_id
	hitDef_chainid
	hitDef_nochainid
	hitDef_kill
	hitDef_guard_kill
	hitDef_fall_kill
	hitDef_hitonce
	hitDef_air_juggle
	hitDef_getpower
	hitDef_damage
	hitDef_givepower
	hitDef_numhits
	hitDef_hitsound
	hitDef_guardsound
	hitDef_priority
	hitDef_p1stateno
	hitDef_p2stateno
	hitDef_p2getp1state
	hitDef_p1sprpriority
	hitDef_p2sprpriority
	hitDef_forcestand
	hitDef_forcenofall
	hitDef_fall_damage
	hitDef_fall_xvelocity
	hitDef_fall_yvelocity
	hitDef_fall_recover
	hitDef_fall_recovertime
	hitDef_sparkno
	hitDef_guard_sparkno
	hitDef_sparkxy
	hitDef_down_hittime
	hitDef_p1facing
	hitDef_p1getp2facing
	hitDef_mindist
	hitDef_maxdist
	hitDef_snap
	hitDef_p2facing
	hitDef_air_hittime
	hitDef_fall
	hitDef_air_fall
	hitDef_air_cornerpush_veloff
	hitDef_down_bounce
	hitDef_down_velocity
	hitDef_down_cornerpush_veloff
	hitDef_ground_hittime
	hitDef_guard_hittime
	hitDef_guard_dist
	hitDef_pausetime
	hitDef_guard_pausetime
	hitDef_air_velocity
	hitDef_airguard_velocity
	hitDef_ground_slidetime
	hitDef_guard_slidetime
	hitDef_guard_ctrltime
	hitDef_airguard_ctrltime
	hitDef_ground_velocity_x
	hitDef_ground_velocity_y
	hitDef_ground_velocity
	hitDef_guard_velocity
	hitDef_ground_cornerpush_veloff
	hitDef_guard_cornerpush_veloff
	hitDef_airguard_cornerpush_veloff
	hitDef_yaccel
	hitDef_envshake_time
	hitDef_envshake_ampl
	hitDef_envshake_phase
	hitDef_envshake_freq
	hitDef_fall_envshake_time
	hitDef_fall_envshake_ampl
	hitDef_fall_envshake_phase
	hitDef_fall_envshake_freq
	hitDef_last byte = iota - 1
)

func (sc hitDef) runSub(c *Char, hd *HitDef, id byte, exp []BytecodeExp) bool {
	switch id {
	case hitDef_attr:
		hd.attr = exp[0].evalI(c)
	case hitDef_guardflag:
		hd.guardflag = exp[0].evalI(c)
	case hitDef_hitflag:
		hd.hitflag = exp[0].evalI(c)
	case hitDef_ground_type:
		hd.ground_type = HitType(exp[0].evalI(c))
	case hitDef_air_type:
		hd.air_type = HitType(exp[0].evalI(c))
	case hitDef_animtype:
		hd.animtype = Reaction(exp[0].evalI(c))
	case hitDef_air_animtype:
		hd.air_animtype = Reaction(exp[0].evalI(c))
	case hitDef_fall_animtype:
		hd.fall.animtype = Reaction(exp[0].evalI(c))
	case hitDef_affectteam:
		hd.affectteam = exp[0].evalI(c)
	case hitDef_id:
		hd.id = Max(0, exp[0].evalI(c))
	case hitDef_chainid:
		hd.chainid = exp[0].evalI(c)
	case hitDef_nochainid:
		hd.nochainid[0] = exp[0].evalI(c)
		if len(exp) > 1 {
			hd.nochainid[1] = exp[1].evalI(c)
		}
	case hitDef_kill:
		hd.kill = exp[0].evalB(c)
	case hitDef_guard_kill:
		hd.guard_kill = exp[0].evalB(c)
	case hitDef_fall_kill:
		hd.fall.kill = exp[0].evalB(c)
	case hitDef_hitonce:
		hd.hitonce = Btoi(exp[0].evalB(c))
	case hitDef_air_juggle:
		hd.air_juggle = exp[0].evalI(c)
	case hitDef_getpower:
		hd.hitgetpower = Max(IErr+1, exp[0].evalI(c))
		if len(exp) > 1 {
			hd.guardgetpower = Max(IErr+1, exp[1].evalI(c))
		}
	case hitDef_damage:
		hd.hitdamage = exp[0].evalI(c)
		if len(exp) > 1 {
			hd.guarddamage = exp[1].evalI(c)
		}
	case hitDef_givepower:
		hd.hitgivepower = Max(IErr+1, exp[0].evalI(c))
		if len(exp) > 1 {
			hd.guardgivepower = Max(IErr+1, exp[1].evalI(c))
		}
	case hitDef_numhits:
		hd.numhits = exp[0].evalI(c)
	case hitDef_hitsound:
		n := exp[1].evalI(c)
		if n < 0 {
			hd.hitsound[0] = IErr
		} else if exp[0].evalB(c) {
			hd.hitsound[0] = ^n
		} else {
			hd.hitsound[0] = n
		}
		if len(exp) > 2 {
			hd.hitsound[1] = exp[2].evalI(c)
		}
	case hitDef_guardsound:
		n := exp[1].evalI(c)
		if n < 0 {
			hd.guardsound[0] = IErr
		} else if exp[0].evalB(c) {
			hd.guardsound[0] = ^n
		} else {
			hd.guardsound[0] = n
		}
		if len(exp) > 2 {
			hd.guardsound[1] = exp[2].evalI(c)
		}
	case hitDef_priority:
		hd.priority = exp[0].evalI(c)
		hd.bothhittype = AiuchiType(exp[1].evalI(c))
	case hitDef_p1stateno:
		hd.p1stateno = exp[0].evalI(c)
	case hitDef_p2stateno:
		hd.p2stateno = exp[0].evalI(c)
		hd.p2getp1state = true
	case hitDef_p2getp1state:
		hd.p2getp1state = exp[0].evalB(c)
	case hitDef_p1sprpriority:
		hd.p1sprpriority = exp[0].evalI(c)
	case hitDef_p2sprpriority:
		hd.p2sprpriority = exp[0].evalI(c)
	case hitDef_forcestand:
		hd.forcestand = Btoi(exp[0].evalB(c))
	case hitDef_forcenofall:
		hd.forcenofall = exp[0].evalB(c)
	case hitDef_fall_damage:
		hd.fall.damage = exp[0].evalI(c)
	case hitDef_fall_xvelocity:
		hd.fall.xvelocity = exp[0].evalF(c)
	case hitDef_fall_yvelocity:
		hd.fall.yvelocity = exp[0].evalF(c)
	case hitDef_fall_recover:
		hd.fall.recover = exp[0].evalB(c)
	case hitDef_fall_recovertime:
		hd.fall.recovertime = exp[0].evalI(c)
	case hitDef_sparkno:
		n := exp[1].evalI(c)
		if n < 0 {
			hd.sparkno = IErr
		} else if exp[0].evalB(c) {
			hd.sparkno = ^n
		} else {
			hd.sparkno = n
		}
	case hitDef_guard_sparkno:
		n := exp[1].evalI(c)
		if n < 0 {
			hd.guard_sparkno = IErr
		} else if exp[0].evalB(c) {
			hd.guard_sparkno = ^n
		} else {
			hd.guard_sparkno = n
		}
	case hitDef_sparkxy:
		hd.sparkxy[0] = exp[0].evalF(c)
		if len(exp) > 1 {
			hd.sparkxy[1] = exp[1].evalF(c)
		}
	case hitDef_down_hittime:
		hd.down_hittime = exp[0].evalI(c)
	case hitDef_p1facing:
		hd.p1facing = exp[0].evalI(c)
	case hitDef_p1getp2facing:
		hd.p1getp2facing = exp[0].evalI(c)
	case hitDef_mindist:
		hd.mindist[0] = exp[0].evalF(c)
		if len(exp) > 1 {
			hd.mindist[1] = exp[1].evalF(c)
		}
	case hitDef_maxdist:
		hd.maxdist[0] = exp[0].evalF(c)
		if len(exp) > 1 {
			hd.maxdist[1] = exp[1].evalF(c)
		}
	case hitDef_snap:
		hd.snap[0] = exp[0].evalF(c)
		if len(exp) > 1 {
			hd.snap[1] = exp[1].evalF(c)
			if len(exp) > 2 {
				exp[2].run(c)
				if len(exp) > 3 {
					hd.snapt = exp[3].evalI(c)
				}
			}
		}
	case hitDef_p2facing:
		hd.p2facing = exp[0].evalI(c)
	case hitDef_air_hittime:
		hd.air_hittime = exp[0].evalI(c)
	case hitDef_fall:
		hd.ground_fall = exp[0].evalB(c)
		hd.air_fall = hd.ground_fall
	case hitDef_air_fall:
		hd.air_fall = exp[0].evalB(c)
	case hitDef_air_cornerpush_veloff:
		hd.air_cornerpush_veloff = exp[0].evalF(c)
	case hitDef_down_bounce:
		hd.down_bounce = exp[0].evalB(c)
	case hitDef_down_velocity:
		hd.down_velocity[0] = exp[0].evalF(c)
		if len(exp) > 1 {
			hd.down_velocity[1] = exp[1].evalF(c)
		}
	case hitDef_down_cornerpush_veloff:
		hd.down_cornerpush_veloff = exp[0].evalF(c)
	case hitDef_ground_hittime:
		hd.ground_hittime = exp[0].evalI(c)
		hd.guard_hittime = hd.ground_hittime
	case hitDef_guard_hittime:
		hd.guard_hittime = exp[0].evalI(c)
	case hitDef_guard_dist:
		hd.guard_dist = exp[0].evalI(c)
	case hitDef_pausetime:
		hd.pausetime = exp[0].evalI(c)
		hd.guard_pausetime = hd.pausetime
		if len(exp) > 1 {
			hd.shaketime = exp[1].evalI(c)
			hd.guard_shaketime = hd.shaketime
		}
	case hitDef_guard_pausetime:
		hd.guard_pausetime = exp[0].evalI(c)
		if len(exp) > 1 {
			hd.guard_shaketime = exp[1].evalI(c)
		}
	case hitDef_air_velocity:
		hd.air_velocity[0] = exp[0].evalF(c)
		if len(exp) > 1 {
			hd.air_velocity[1] = exp[1].evalF(c)
		}
	case hitDef_airguard_velocity:
		hd.airguard_velocity[0] = exp[0].evalF(c)
		if len(exp) > 1 {
			hd.airguard_velocity[1] = exp[1].evalF(c)
		}
	case hitDef_ground_slidetime:
		hd.ground_slidetime = exp[0].evalI(c)
		hd.guard_slidetime = hd.ground_slidetime
		hd.guard_ctrltime = hd.ground_slidetime
		hd.airguard_ctrltime = hd.ground_slidetime
	case hitDef_guard_slidetime:
		hd.guard_slidetime = exp[0].evalI(c)
		hd.guard_ctrltime = hd.guard_slidetime
		hd.airguard_ctrltime = hd.guard_slidetime
	case hitDef_guard_ctrltime:
		hd.guard_ctrltime = exp[0].evalI(c)
		hd.airguard_ctrltime = hd.guard_ctrltime
	case hitDef_airguard_ctrltime:
		hd.airguard_ctrltime = exp[0].evalI(c)
	case hitDef_ground_velocity_x:
		hd.ground_velocity[0] = exp[0].evalF(c)
	case hitDef_ground_velocity_y:
		hd.ground_velocity[1] = exp[0].evalF(c)
	case hitDef_guard_velocity:
		hd.guard_velocity = exp[0].evalF(c)
	case hitDef_ground_cornerpush_veloff:
		hd.ground_cornerpush_veloff = exp[0].evalF(c)
	case hitDef_guard_cornerpush_veloff:
		hd.guard_cornerpush_veloff = exp[0].evalF(c)
	case hitDef_airguard_cornerpush_veloff:
		hd.airguard_cornerpush_veloff = exp[0].evalF(c)
	case hitDef_yaccel:
		hd.yaccel = exp[0].evalF(c)
	case hitDef_envshake_time:
		hd.envshake_time = exp[0].evalI(c)
	case hitDef_envshake_ampl:
		hd.envshake_ampl = exp[0].evalI(c)
	case hitDef_envshake_phase:
		hd.envshake_phase = exp[0].evalF(c)
	case hitDef_envshake_freq:
		hd.envshake_freq = MaxF(0, exp[0].evalF(c))
	case hitDef_fall_envshake_time:
		hd.fall.envshake_time = exp[0].evalI(c)
	case hitDef_fall_envshake_ampl:
		hd.fall.envshake_ampl = exp[0].evalI(c)
	case hitDef_fall_envshake_phase:
		hd.fall.envshake_phase = exp[0].evalF(c)
	case hitDef_fall_envshake_freq:
		hd.fall.envshake_freq = MaxF(0, exp[0].evalF(c))
	default:
		if !palFX(sc).runSub(c, &hd.palfx, id, exp) {
			return false
		}
	}
	return true
}
func (sc hitDef) Run(c *Char, _ []int32) bool {
	c.hitdef.clear()
	c.hitdef.sparkno = ^sys.cgi[c.playerNo].data.sparkno
	c.hitdef.guard_sparkno = ^sys.cgi[c.playerNo].data.guard.sparkno
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		sc.runSub(c, &c.hitdef, id, exp)
		return true
	})
	c.setHitdefDefault(&c.hitdef, false)
	return false
}

type reversalDef hitDef

const (
	reversalDef_reversal_attr byte = iota + hitDef_last + 1
)

func (sc reversalDef) Run(c *Char, _ []int32) bool {
	c.hitdef.clear()
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case reversalDef_reversal_attr:
			c.hitdef.reversal_attr = exp[0].evalI(c)
		default:
			hitDef(sc).runSub(c, &c.hitdef, id, exp)
		}
		return true
	})
	c.setHitdefDefault(&c.hitdef, false)
	return false
}

type projectile hitDef

const (
	projectile_postype byte = iota + hitDef_last + 1
	projectile_projid
	projectile_projremove
	projectile_projremovetime
	projectile_projshadow
	projectile_projmisstime
	projectile_projhits
	projectile_projpriority
	projectile_projhitanim
	projectile_projremanim
	projectile_projcancelanim
	projectile_velocity
	projectile_velmul
	projectile_remvelocity
	projectile_accel
	projectile_projscale
	projectile_offset
	projectile_projsprpriority
	projectile_projstagebound
	projectile_projedgebound
	projectile_projheightbound
	projectile_projanim
	projectile_supermovetime
	projectile_pausemovetime
	projectile_ownpal
	projectile_remappal
)

func (sc projectile) Run(c *Char, _ []int32) bool {
	var p *Projectile
	pt := PT_P1
	var x, y float32 = 0, 0
	op := false
	rp := [2]int32{-1, 0}
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		if p == nil {
			p = c.newProj()
			if p == nil {
				return false
			}
			p.aimg.clear()
		}
		switch id {
		case projectile_postype:
			pt = PosType(exp[0].evalI(c))
		case projectile_projid:
			p.id = exp[0].evalI(c)
		case projectile_projremove:
			p.remove = exp[0].evalB(c)
		case projectile_projremovetime:
			p.removetime = exp[0].evalI(c)
		case projectile_projshadow:
			p.shadow[0] = exp[0].evalI(c)
			if len(exp) > 1 {
				p.shadow[1] = exp[1].evalI(c)
				if len(exp) > 2 {
					p.shadow[2] = exp[2].evalI(c)
				}
			}
		case projectile_projmisstime:
			p.misstime = exp[0].evalI(c)
		case projectile_projhits:
			p.hits = exp[0].evalI(c)
		case projectile_projpriority:
			p.priority = exp[0].evalI(c)
		case projectile_projhitanim:
			p.hitanim = exp[0].evalI(c)
		case projectile_projremanim:
			p.remanim = Max(-1, exp[0].evalI(c))
		case projectile_projcancelanim:
			p.cancelanim = Max(-1, exp[0].evalI(c))
		case projectile_velocity:
			p.velocity[0] = exp[0].evalF(c)
			if len(exp) > 1 {
				p.velocity[1] = exp[1].evalF(c)
			}
		case projectile_velmul:
			p.velmul[0] = exp[0].evalF(c)
			if len(exp) > 1 {
				p.velmul[1] = exp[1].evalF(c)
			}
		case projectile_remvelocity:
			p.remvelocity[0] = exp[0].evalF(c)
			if len(exp) > 1 {
				p.remvelocity[1] = exp[1].evalF(c)
			}
		case projectile_accel:
			p.accel[0] = exp[0].evalF(c)
			if len(exp) > 1 {
				p.accel[1] = exp[1].evalF(c)
			}
		case projectile_projscale:
			p.scale[0] = exp[0].evalF(c)
			if len(exp) > 1 {
				p.scale[1] = exp[1].evalF(c)
			}
		case projectile_offset:
			x = exp[0].evalF(c)
			if len(exp) > 1 {
				y = exp[1].evalF(c)
			}
		case projectile_projsprpriority:
			p.sprpriority = exp[0].evalI(c)
		case projectile_projstagebound:
			p.stagebound = exp[0].evalI(c)
		case projectile_projedgebound:
			p.edgebound = exp[0].evalI(c)
		case projectile_projheightbound:
			p.heightbound[0] = exp[0].evalI(c)
			if len(exp) > 1 {
				p.heightbound[1] = exp[1].evalI(c)
			}
		case projectile_projanim:
			p.anim = exp[0].evalI(c)
		case projectile_supermovetime:
			p.supermovetime = exp[0].evalI(c)
		case projectile_pausemovetime:
			p.pausemovetime = exp[0].evalI(c)
		case projectile_ownpal:
			op = exp[0].evalB(c)
		case projectile_remappal:
			rp[0] = exp[0].evalI(c)
			if len(exp) > 1 {
				rp[1] = exp[1].evalI(c)
			}
		default:
			if !hitDef(sc).runSub(c, &p.hitdef, id, exp) {
				afterImage(sc).runSub(c, &p.aimg, id, exp)
			}
		}
		return true
	})
	if p != nil {
		c.setHitdefDefault(&c.hitdef, true)
		if p.remanim == IErr {
			p.remanim = p.hitanim
		}
		if p.cancelanim == IErr {
			p.cancelanim = p.remanim
		}
		if p.aimg.time != 0 {
			p.aimg.setupPalFX()
		}
		c.projInit(p, pt, x, y, op, rp[0], rp[1])
	}
	return false
}

type width StateControllerBase

const (
	width_edge byte = iota
	width_player
	width_value
)

func (sc width) Run(c *Char, _ []int32) bool {
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case width_edge:
			c.setFEdge(exp[0].evalF(c))
			if len(exp) > 1 {
				c.setBEdge(exp[1].evalF(c))
			}
		case width_player:
			c.setFWidth(exp[0].evalF(c))
			if len(exp) > 1 {
				c.setBWidth(exp[1].evalF(c))
			}
		case width_value:
			v1 := exp[0].evalF(c)
			c.setFEdge(v1)
			c.setFWidth(v1)
			if len(exp) > 1 {
				v2 := exp[1].evalF(c)
				c.setBEdge(v2)
				c.setBWidth(v2)
			}
		}
		return true
	})
	return false
}

type sprPriority StateControllerBase

const (
	sprPriority_value byte = iota
)

func (sc sprPriority) Run(c *Char, _ []int32) bool {
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case sprPriority_value:
			c.sprpriority = exp[0].evalI(c)
		}
		return true
	})
	return false
}

type varSet StateControllerBase

const (
	varSet_ byte = iota
)

func (sc varSet) Run(c *Char, _ []int32) bool {
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case varSet_:
			exp[0].run(c)
		}
		return true
	})
	return false
}

type turn StateControllerBase

const (
	turn_ byte = iota
)

func (sc turn) Run(c *Char, _ []int32) bool {
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case turn_:
			c.setFacing(-c.facing)
		}
		return true
	})
	return false
}

type targetFacing StateControllerBase

const (
	targetFacing_id byte = iota
	targetFacing_value
)

func (sc targetFacing) Run(c *Char, _ []int32) bool {
	var tar []int32
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		if len(tar) == 0 {
			tar = c.getTarget(-1)
			if len(tar) == 0 {
				return false
			}
		}
		switch id {
		case targetFacing_id:
			tar = c.getTarget(exp[0].evalI(c))
			if len(tar) == 0 {
				return false
			}
		case targetFacing_value:
			c.targetFacing(tar, exp[0].evalI(c))
		}
		return true
	})
	return false
}

type targetBind StateControllerBase

const (
	targetBind_id byte = iota
	targetBind_time
	targetBind_pos
)

func (sc targetBind) Run(c *Char, _ []int32) bool {
	var tar []int32
	t := int32(1)
	var x, y float32 = 0, 0
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		if len(tar) == 0 {
			tar = c.getTarget(-1)
			if len(tar) == 0 {
				return false
			}
		}
		switch id {
		case targetBind_id:
			tar = c.getTarget(exp[0].evalI(c))
			if len(tar) == 0 {
				return false
			}
		case targetBind_time:
			t = exp[0].evalI(c)
		case targetBind_pos:
			x = exp[0].evalF(c)
			if len(exp) > 1 {
				y = exp[1].evalF(c)
			}
		}
		return true
	})
	c.targetBind(tar, t, x, y)
	return false
}

type bindToTarget StateControllerBase

const (
	bindToTarget_id byte = iota
	bindToTarget_time
	bindToTarget_pos
)

func (sc bindToTarget) Run(c *Char, _ []int32) bool {
	var tar []int32
	t := int32(1)
	x, y := float32(0), float32(math.NaN())
	hmf := HMF_F
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		if len(tar) == 0 {
			tar = c.getTarget(-1)
			if len(tar) == 0 {
				return false
			}
		}
		switch id {
		case bindToTarget_id:
			tar = c.getTarget(exp[0].evalI(c))
			if len(tar) == 0 {
				return false
			}
		case bindToTarget_time:
			t = exp[0].evalI(c)
		case bindToTarget_pos:
			x = exp[0].evalF(c)
			if len(exp) > 1 {
				y = exp[1].evalF(c)
				if len(exp) > 2 {
					hmf = HMF(exp[2].evalI(c))
				}
			}
		}
		return true
	})
	c.bindToTarget(tar, t, x, y, hmf)
	return false
}

type targetLifeAdd StateControllerBase

const (
	targetLifeAdd_id byte = iota
	targetLifeAdd_absolute
	targetLifeAdd_kill
	targetLifeAdd_value
)

func (sc targetLifeAdd) Run(c *Char, _ []int32) bool {
	var tar []int32
	a, k := false, true
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		if len(tar) == 0 {
			tar = c.getTarget(-1)
			if len(tar) == 0 {
				return false
			}
		}
		switch id {
		case targetLifeAdd_id:
			tar = c.getTarget(exp[0].evalI(c))
			if len(tar) == 0 {
				return false
			}
		case targetLifeAdd_absolute:
			a = exp[0].evalB(c)
		case targetLifeAdd_kill:
			k = exp[0].evalB(c)
		case targetLifeAdd_value:
			c.targetLifeAdd(tar, exp[0].evalI(c), k, a)
		}
		return true
	})
	return false
}

type targetState StateControllerBase

const (
	targetState_id byte = iota
	targetState_value
)

func (sc targetState) Run(c *Char, _ []int32) bool {
	var tar []int32
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		if len(tar) == 0 {
			tar = c.getTarget(-1)
			if len(tar) == 0 {
				return false
			}
		}
		switch id {
		case targetState_id:
			tar = c.getTarget(exp[0].evalI(c))
			if len(tar) == 0 {
				return false
			}
		case targetState_value:
			c.targetState(tar, exp[0].evalI(c))
		}
		return true
	})
	return false
}

type targetVelSet StateControllerBase

const (
	targetVelSet_id byte = iota
	targetVelSet_x
	targetVelSet_y
)

func (sc targetVelSet) Run(c *Char, _ []int32) bool {
	var tar []int32
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		if len(tar) == 0 {
			tar = c.getTarget(-1)
			if len(tar) == 0 {
				return false
			}
		}
		switch id {
		case targetVelSet_id:
			tar = c.getTarget(exp[0].evalI(c))
			if len(tar) == 0 {
				return false
			}
		case targetVelSet_x:
			c.targetVelSetX(tar, exp[0].evalF(c))
		case targetVelSet_y:
			c.targetVelSetY(tar, exp[0].evalF(c))
		}
		return true
	})
	return false
}

type targetVelAdd StateControllerBase

const (
	targetVelAdd_id byte = iota
	targetVelAdd_x
	targetVelAdd_y
)

func (sc targetVelAdd) Run(c *Char, _ []int32) bool {
	var tar []int32
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		if len(tar) == 0 {
			tar = c.getTarget(-1)
			if len(tar) == 0 {
				return false
			}
		}
		switch id {
		case targetVelAdd_id:
			tar = c.getTarget(exp[0].evalI(c))
			if len(tar) == 0 {
				return false
			}
		case targetVelAdd_x:
			c.targetVelAddX(tar, exp[0].evalF(c))
		case targetVelAdd_y:
			c.targetVelAddY(tar, exp[0].evalF(c))
		}
		return true
	})
	return false
}

type targetPowerAdd StateControllerBase

const (
	targetPowerAdd_id byte = iota
	targetPowerAdd_value
)

func (sc targetPowerAdd) Run(c *Char, _ []int32) bool {
	var tar []int32
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		if len(tar) == 0 {
			tar = c.getTarget(-1)
			if len(tar) == 0 {
				return false
			}
		}
		switch id {
		case targetPowerAdd_id:
			tar = c.getTarget(exp[0].evalI(c))
			if len(tar) == 0 {
				return false
			}
		case targetPowerAdd_value:
			c.targetPowerAdd(tar, exp[0].evalI(c))
		}
		return true
	})
	return false
}

type targetDrop StateControllerBase

const (
	targetDrop_excludeid byte = iota
	targetDrop_keepone
)

func (sc targetDrop) Run(c *Char, _ []int32) bool {
	var tar []int32
	eid := int32(-1)
	ko := true
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		if len(tar) == 0 {
			tar = c.getTarget(-1)
			if len(tar) == 0 {
				return false
			}
		}
		switch id {
		case targetDrop_excludeid:
			eid = exp[0].evalI(c)
		case targetDrop_keepone:
			ko = exp[0].evalB(c)
		}
		return true
	})
	c.targetDrop(eid, ko)
	return false
}

type lifeAdd StateControllerBase

const (
	lifeAdd_absolute byte = iota
	lifeAdd_kill
	lifeAdd_value
)

func (sc lifeAdd) Run(c *Char, _ []int32) bool {
	a, k := false, true
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case lifeAdd_absolute:
			a = exp[0].evalB(c)
		case lifeAdd_kill:
			k = exp[0].evalB(c)
		case lifeAdd_value:
			c.lifeAdd(exp[0].evalI(c), k, a)
		}
		return true
	})
	return false
}

type lifeSet StateControllerBase

const (
	lifeSet_value byte = iota
)

func (sc lifeSet) Run(c *Char, _ []int32) bool {
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case lifeSet_value:
			c.lifeSet(exp[0].evalI(c))
		}
		return true
	})
	return false
}

type powerAdd StateControllerBase

const (
	powerAdd_value byte = iota
)

func (sc powerAdd) Run(c *Char, _ []int32) bool {
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case powerAdd_value:
			c.powerAdd(exp[0].evalI(c))
		}
		return true
	})
	return false
}

type powerSet StateControllerBase

const (
	powerSet_value byte = iota
)

func (sc powerSet) Run(c *Char, _ []int32) bool {
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case powerSet_value:
			c.powerSet(exp[0].evalI(c))
		}
		return true
	})
	return false
}

type hitVelSet StateControllerBase

const (
	hitVelSet_x byte = iota
	hitVelSet_y
)

func (sc hitVelSet) Run(c *Char, _ []int32) bool {
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case hitVelSet_x:
			if exp[0].evalB(c) {
				c.hitVelSetX()
			}
		case hitVelSet_y:
			if exp[0].evalB(c) {
				c.hitVelSetY()
			}
		}
		return true
	})
	return false
}

type screenBound StateControllerBase

const (
	screenBound_value byte = iota
	screenBound_movecamera
)

func (sc screenBound) Run(c *Char, _ []int32) bool {
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case screenBound_value:
			if exp[0].evalB(c) {
				c.setSF(CSF_screenbound)
			} else {
				c.unsetSF(CSF_screenbound)
			}
		case screenBound_movecamera:
			if exp[0].evalB(c) {
				c.setSF(CSF_movecamera_x)
			} else {
				c.unsetSF(CSF_movecamera_x)
			}
			if len(exp) > 1 {
				if exp[1].evalB(c) {
					c.setSF(CSF_movecamera_y)
				} else {
					c.unsetSF(CSF_movecamera_y)
				}
			}
		}
		return true
	})
	return false
}

type posFreeze StateControllerBase

const (
	posFreeze_value byte = iota
)

func (sc posFreeze) Run(c *Char, _ []int32) bool {
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case posFreeze_value:
			if exp[0].evalB(c) {
				c.setSF(CSF_posfreeze)
			}
		}
		return true
	})
	return false
}

type envShake StateControllerBase

const (
	envShake_time byte = iota
	envShake_ampl
	envShake_phase
	envShake_freq
)

func (sc envShake) Run(c *Char, _ []int32) bool {
	sys.envShake.clear()
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case envShake_time:
			sys.envShake.time = exp[0].evalI(c)
		case envShake_ampl:
			sys.envShake.ampl = exp[0].evalI(c)
		case envShake_phase:
			sys.envShake.phase = MaxF(0, exp[0].evalF(c)*float32(math.Pi)/180)
		case envShake_freq:
			sys.envShake.freq = exp[0].evalF(c)
		}
		return true
	})
	sys.envShake.setDefPhase()
	return false
}

type hitOverride StateControllerBase

const (
	hitOverride_attr byte = iota
	hitOverride_slot
	hitOverride_stateno
	hitOverride_time
	hitOverride_forceair
)

func (sc hitOverride) Run(c *Char, _ []int32) bool {
	var a, s, st, t int32 = 0, 0, -1, 1
	f := false
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case hitOverride_attr:
			a = exp[0].evalI(c)
		case hitOverride_slot:
			s = Max(0, exp[0].evalI(c))
			if s > 7 {
				s = 0
			}
		case hitOverride_stateno:
			st = exp[0].evalI(c)
		case hitOverride_time:
			t = exp[0].evalI(c)
		case hitOverride_forceair:
			f = exp[0].evalB(c)
		}
		return true
	})
	if st < 0 {
		t = 0
	}
	c.ho[s] = HitOverride{attr: a, stateno: st, time: t, forceair: f,
		playerNo: sys.workingChar.ss.sb.playerNo}
	return false
}

type pause StateControllerBase

const (
	pause_time byte = iota
	pause_movetime
	pause_pausebg
	pause_endcmdbuftime
)

func (sc pause) Run(c *Char, _ []int32) bool {
	var t, mt int32 = 0, 0
	sys.pausebg, sys.pauseendcmdbuftime = true, 0
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case pause_time:
			t = exp[0].evalI(c)
		case pause_movetime:
			mt = exp[0].evalI(c)
		case pause_pausebg:
			sys.pausebg = exp[0].evalB(c)
		case pause_endcmdbuftime:
			sys.pauseendcmdbuftime = exp[0].evalI(c)
		}
		return true
	})
	c.setPauseTime(t, mt)
	return false
}

type superPause StateControllerBase

const (
	superPause_time byte = iota
	superPause_movetime
	superPause_pausebg
	superPause_endcmdbuftime
	superPause_darken
	superPause_anim
	superPause_pos
	superPause_p2defmul
	superPause_poweradd
	superPause_unhittable
	superPause_sound
)

func (sc superPause) Run(c *Char, _ []int32) bool {
	var t, mt int32 = 0, 0
	sys.superanim, sys.superpmap.remap = c.getAnim(30, true), nil
	sys.superpos, sys.superfacing = c.pos, c.facing
	sys.superpausebg, sys.superendcmdbuftime, sys.superdarken = true, 0, true
	sys.superp2defmul, sys.superunhittable = sys.super_TargetDefenceMul, true
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case superPause_time:
			t = exp[0].evalI(c)
		case superPause_movetime:
			mt = exp[0].evalI(c)
		case superPause_pausebg:
			sys.superpausebg = exp[0].evalB(c)
		case superPause_endcmdbuftime:
			sys.superendcmdbuftime = exp[0].evalI(c)
		case superPause_darken:
			sys.superdarken = exp[0].evalB(c)
		case superPause_anim:
			f := exp[0].evalB(c)
			if sys.superanim = c.getAnim(exp[1].evalI(c), f); sys.superanim != nil {
				if f {
					sys.superpmap.remap = nil
				} else {
					sys.superpmap.remap = c.getPalMap()
				}
			}
		case superPause_pos:
			sys.superpos[0] += c.facing * exp[0].evalF(c)
			if len(exp) > 1 {
				sys.superpos[1] += exp[1].evalF(c)
			}
		case superPause_p2defmul:
			if f := c.facing * exp[0].evalF(c); f != 0 {
				sys.superp2defmul = f
			}
		case superPause_poweradd:
			c.powerAdd(exp[0].evalI(c))
		case superPause_unhittable:
			sys.superunhittable = exp[0].evalB(c)
		case superPause_sound:
			n := int32(0)
			if len(exp) > 2 {
				n = exp[2].evalI(c)
			}
			vo := int32(0)
			if sys.cgi[sys.workingChar.ss.sb.playerNo].ver[0] == 1 {
				vo = 100
			}
			c.playSound(exp[0].evalB(c), false, false, exp[1].evalI(c), n, -1,
				vo, 0, 1, &c.pos[0])
		}
		return true
	})
	c.setSuperPauseTime(t, mt)
	return false
}

type StateBytecode struct {
	stateType StateType
	moveType  MoveType
	physics   StateType
	playerNo  int
	stateDef  StateController
	block     StateBlock
	ctrlsps   []int32
}

func newStateBytecode(pn int) *StateBytecode {
	sb := &StateBytecode{stateType: ST_S, moveType: MT_I, physics: ST_N,
		playerNo: pn}
	sb.block.ignorehitpause = true
	return sb
}

type Bytecode struct{ states map[int32]StateBytecode }

func newBytecode() *Bytecode {
	return &Bytecode{states: make(map[int32]StateBytecode)}
}
