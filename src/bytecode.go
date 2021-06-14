package main

import (
	"encoding/gob"
	"math"
	"os"
	"path/filepath"
	"strings"
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
	ST_SCA  = ST_S | ST_C | ST_A
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
	AT_AA  = AT_NA | AT_SA | AT_HA
	AT_AT  = AT_NT | AT_ST | AT_HT
	AT_AP  = AT_NP | AT_SP | AT_HP
	AT_ALL = AT_AA | AT_AT | AT_AP
	AT_AN  = AT_NA | AT_NT | AT_NP
	AT_AS  = AT_SA | AT_ST | AT_SP
	AT_AH  = AT_HA | AT_HT | AT_HP
)

type MoveType int32

const (
	MT_I MoveType = 1 << (iota + 15)
	MT_H
	MT_A
	MT_U
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
	OC_localvar
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
	OC_var0     = 0
	OC_sysvar0  = 60
	OC_fvar0    = 65
	OC_sysfvar0 = 105
)
const (
	OC_const_data_life OpCode = iota
	OC_const_data_power
	OC_const_data_guardpoints
	OC_const_data_dizzypoints
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
	OC_const_velocity_runjump_y
	OC_const_velocity_runjump_fwd_x
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
	OC_const_p2name
	OC_const_p3name
	OC_const_p4name
	OC_const_p5name
	OC_const_p6name
	OC_const_p7name
	OC_const_p8name
	OC_const_authorname
	OC_const_stagevar_info_author
	OC_const_stagevar_info_displayname
	OC_const_stagevar_info_name
	OC_const_constants
	OC_const_stage_constants
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
	OC_st_var0        = OC_var0
	OC_st_sysvar0     = OC_sysvar0
	OC_st_fvar0       = OC_fvar0
	OC_st_sysfvar0    = OC_sysfvar0
	OC_st_var0add     = OC_var + OC_var0
	OC_st_sysvar0add  = OC_var + OC_sysvar0
	OC_st_fvar0add    = OC_var + OC_fvar0
	OC_st_sysfvar0add = OC_var + OC_sysfvar0
)
const (
	OC_ex_p2dist_x OpCode = iota
	OC_ex_p2dist_y
	OC_ex_p2bodydist_x
	OC_ex_parentdist_x
	OC_ex_parentdist_y
	OC_ex_rootdist_x
	OC_ex_rootdist_y
	OC_ex_win
	OC_ex_winko
	OC_ex_wintime
	OC_ex_winperfect
	OC_ex_winspecial
	OC_ex_winhyper
	OC_ex_lose
	OC_ex_loseko
	OC_ex_losetime
	OC_ex_drawgame
	OC_ex_matchover
	OC_ex_matchno
	OC_ex_roundno
	OC_ex_ishometeam
	OC_ex_tickspersecond
	OC_ex_majorversion
	OC_ex_drawpalno
	OC_ex_const240p
	OC_ex_const480p
	OC_ex_const720p
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
	OC_ex_gethitvar_attr
	OC_ex_gethitvar_dizzypoints
	OC_ex_gethitvar_guardpoints
	OC_ex_gethitvar_id
	OC_ex_gethitvar_playerno
	OC_ex_gethitvar_redlife
	OC_ex_gethitvar_score
	OC_ex_gethitvar_hitdamage
	OC_ex_gethitvar_guarddamage
	OC_ex_gethitvar_hitpower
	OC_ex_gethitvar_guardpower
	OC_ex_ailevelf
	OC_ex_animelemlength
	OC_ex_animlength
	OC_ex_combocount
	OC_ex_consecutivewins
	OC_ex_dizzy
	OC_ex_dizzypoints
	OC_ex_dizzypointsmax
	OC_ex_firstattack
	OC_ex_float
	OC_ex_gamemode
	OC_ex_getplayerid
	OC_ex_groundangle
	OC_ex_guardbreak
	OC_ex_guardpoints
	OC_ex_guardpointsmax
	OC_ex_helpername
	OC_ex_hitoverridden
	OC_ex_incustomstate
	OC_ex_indialogue
	OC_ex_isassertedchar
	OC_ex_isassertedglobal
	OC_ex_ishost
	OC_ex_localscale
	OC_ex_maparray
	OC_ex_max
	OC_ex_min
	OC_ex_memberno
	OC_ex_movecountered
	OC_ex_pausetime
	OC_ex_physics
	OC_ex_playerno
	OC_ex_rand
	OC_ex_rank
	OC_ex_ratiolevel
	OC_ex_receiveddamage
	OC_ex_receivedhits
	OC_ex_redlife
	OC_ex_round
	OC_ex_roundtype
	OC_ex_score
	OC_ex_scoretotal
	OC_ex_selfstatenoexist
	OC_ex_sprpriority
	OC_ex_stagebackedge
	OC_ex_stagefrontedge
	OC_ex_stagetime
	OC_ex_standby
	OC_ex_teamleader
	OC_ex_teamsize
	OC_ex_timeelapsed
	OC_ex_timeremaining
	OC_ex_timetotal
	OC_ex_pos_z
	OC_ex_vel_z
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
func (bs *BytecodeStack) PushI(i int32)         { bs.Push(BytecodeInt(i)) }
func (bs *BytecodeStack) PushF(f float32)       { bs.Push(BytecodeFloat(f)) }
func (bs *BytecodeStack) PushB(b bool)          { bs.Push(BytecodeBool(b)) }
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
func (bs *BytecodeStack) Alloc(size int) []BytecodeValue {
	if len(*bs)+size > cap(*bs) {
		tmp := *bs
		*bs = make(BytecodeStack, len(*bs)+size)
		copy(*bs, tmp)
	} else {
		*bs = (*bs)[:len(*bs)+size]
		for i := len(*bs) - size; i < len(*bs); i++ {
			(*bs)[i] = bvNone()
		}
	}
	return (*bs)[len(*bs)-size:]
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
		if bv.v >= -128 && bv.v <= 127 {
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
	be.append(op)
	be.append((*(*[4]OpCode)(unsafe.Pointer(&addr)))[:]...)
}
func (BytecodeExp) neg(v *BytecodeValue) {
	if v.t == VT_Float {
		v.v *= -1
	} else {
		v.SetI(-v.ToI())
	}
}
func (BytecodeExp) not(v *BytecodeValue) {
	v.SetI(^v.ToI())
}
func (BytecodeExp) blnot(v *BytecodeValue) {
	v.SetB(!v.ToB())
}
func (BytecodeExp) pow(v1 *BytecodeValue, v2 BytecodeValue, pn int) {
	if ValueType(Min(int32(v1.t), int32(v2.t))) == VT_Float {
		v1.SetF(Pow(v1.ToF(), v2.ToF()))
	} else if v2.ToF() < 0 {
		v1.SetF(Pow(v1.ToF(), v2.ToF()))
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
func (BytecodeExp) mul(v1 *BytecodeValue, v2 BytecodeValue) {
	if ValueType(Min(int32(v1.t), int32(v2.t))) == VT_Float {
		v1.SetF(v1.ToF() * v2.ToF())
	} else {
		v1.SetI(v1.ToI() * v2.ToI())
	}
}
func (BytecodeExp) div(v1 *BytecodeValue, v2 BytecodeValue) {
	if ValueType(Min(int32(v1.t), int32(v2.t))) == VT_Float {
		v1.SetF(v1.ToF() / v2.ToF())
	} else if v2.ToI() == 0 {
		*v1 = BytecodeSF()
	} else {
		v1.SetI(v1.ToI() / v2.ToI())
	}
}
func (BytecodeExp) mod(v1 *BytecodeValue, v2 BytecodeValue) {
	if v2.ToI() == 0 {
		*v1 = BytecodeSF()
	} else {
		v1.SetI(v1.ToI() % v2.ToI())
	}
}
func (BytecodeExp) add(v1 *BytecodeValue, v2 BytecodeValue) {
	if ValueType(Min(int32(v1.t), int32(v2.t))) == VT_Float {
		v1.SetF(v1.ToF() + v2.ToF())
	} else {
		v1.SetI(v1.ToI() + v2.ToI())
	}
}
func (BytecodeExp) sub(v1 *BytecodeValue, v2 BytecodeValue) {
	if ValueType(Min(int32(v1.t), int32(v2.t))) == VT_Float {
		v1.SetF(v1.ToF() - v2.ToF())
	} else {
		v1.SetI(v1.ToI() - v2.ToI())
	}
}
func (BytecodeExp) gt(v1 *BytecodeValue, v2 BytecodeValue) {
	if ValueType(Min(int32(v1.t), int32(v2.t))) == VT_Float {
		v1.SetB(v1.ToF() > v2.ToF())
	} else {
		v1.SetB(v1.ToI() > v2.ToI())
	}
}
func (BytecodeExp) ge(v1 *BytecodeValue, v2 BytecodeValue) {
	if ValueType(Min(int32(v1.t), int32(v2.t))) == VT_Float {
		v1.SetB(v1.ToF() >= v2.ToF())
	} else {
		v1.SetB(v1.ToI() >= v2.ToI())
	}
}
func (BytecodeExp) lt(v1 *BytecodeValue, v2 BytecodeValue) {
	if ValueType(Min(int32(v1.t), int32(v2.t))) == VT_Float {
		v1.SetB(v1.ToF() < v2.ToF())
	} else {
		v1.SetB(v1.ToI() < v2.ToI())
	}
}
func (BytecodeExp) le(v1 *BytecodeValue, v2 BytecodeValue) {
	if ValueType(Min(int32(v1.t), int32(v2.t))) == VT_Float {
		v1.SetB(v1.ToF() <= v2.ToF())
	} else {
		v1.SetB(v1.ToI() <= v2.ToI())
	}
}
func (BytecodeExp) eq(v1 *BytecodeValue, v2 BytecodeValue) {
	if ValueType(Min(int32(v1.t), int32(v2.t))) == VT_Float {
		v1.SetB(v1.ToF() == v2.ToF())
	} else {
		v1.SetB(v1.ToI() == v2.ToI())
	}
}
func (BytecodeExp) ne(v1 *BytecodeValue, v2 BytecodeValue) {
	if ValueType(Min(int32(v1.t), int32(v2.t))) == VT_Float {
		v1.SetB(v1.ToF() != v2.ToF())
	} else {
		v1.SetB(v1.ToI() != v2.ToI())
	}
}
func (BytecodeExp) and(v1 *BytecodeValue, v2 BytecodeValue) {
	v1.SetI(v1.ToI() & v2.ToI())
}
func (BytecodeExp) xor(v1 *BytecodeValue, v2 BytecodeValue) {
	v1.SetI(v1.ToI() ^ v2.ToI())
}
func (BytecodeExp) or(v1 *BytecodeValue, v2 BytecodeValue) {
	v1.SetI(v1.ToI() | v2.ToI())
}
func (BytecodeExp) bland(v1 *BytecodeValue, v2 BytecodeValue) {
	v1.SetB(v1.ToB() && v2.ToB())
}
func (BytecodeExp) blxor(v1 *BytecodeValue, v2 BytecodeValue) {
	v1.SetB(v1.ToB() != v2.ToB())
}
func (BytecodeExp) blor(v1 *BytecodeValue, v2 BytecodeValue) {
	v1.SetB(v1.ToB() || v2.ToB())
}
func (BytecodeExp) abs(v1 *BytecodeValue) {
	if v1.t == VT_Float {
		v1.v = math.Abs(v1.v)
	} else {
		v1.SetI(Abs(v1.ToI()))
	}
}
func (BytecodeExp) exp(v1 *BytecodeValue) {
	v1.SetF(float32(math.Exp(v1.v)))
}
func (BytecodeExp) ln(v1 *BytecodeValue) {
	if v1.v <= 0 {
		*v1 = BytecodeSF()
	} else {
		v1.SetF(float32(math.Log(v1.v)))
	}
}
func (BytecodeExp) log(v1 *BytecodeValue, v2 BytecodeValue) {
	if v1.v <= 0 || v2.v <= 0 {
		*v1 = BytecodeSF()
	} else {
		v1.SetF(float32(math.Log(v2.v) / math.Log(v1.v)))
	}
}
func (BytecodeExp) cos(v1 *BytecodeValue) {
	v1.SetF(float32(math.Cos(v1.v)))
}
func (BytecodeExp) sin(v1 *BytecodeValue) {
	v1.SetF(float32(math.Sin(v1.v)))
}
func (BytecodeExp) tan(v1 *BytecodeValue) {
	v1.SetF(float32(math.Tan(v1.v)))
}
func (BytecodeExp) acos(v1 *BytecodeValue) {
	v1.SetF(float32(math.Acos(v1.v)))
}
func (BytecodeExp) asin(v1 *BytecodeValue) {
	v1.SetF(float32(math.Asin(v1.v)))
}
func (BytecodeExp) atan(v1 *BytecodeValue) {
	v1.SetF(float32(math.Atan(v1.v)))
}
func (BytecodeExp) floor(v1 *BytecodeValue) {
	if v1.t == VT_Float {
		f := math.Floor(v1.v)
		if math.IsNaN(f) {
			*v1 = BytecodeSF()
		} else {
			v1.SetI(int32(f))
		}
	}
}
func (BytecodeExp) ceil(v1 *BytecodeValue) {
	if v1.t == VT_Float {
		f := math.Ceil(v1.v)
		if math.IsNaN(f) {
			*v1 = BytecodeSF()
		} else {
			v1.SetI(int32(f))
		}
	}
}
func (BytecodeExp) max(v1 *BytecodeValue, v2 BytecodeValue) {
	if v1.v >= v2.v {
		v1.SetF(float32(v1.v))
	} else {
		v1.SetF(float32(v2.v))
	}
}
func (BytecodeExp) min(v1 *BytecodeValue, v2 BytecodeValue) {
	if v1.v <= v2.v {
		v1.SetF(float32(v1.v))
	} else {
		v1.SetF(float32(v2.v))
	}
}
func (BytecodeExp) random(v1 *BytecodeValue, v2 BytecodeValue) {
	v1.SetI(RandI(int32(v1.v), int32(v2.v)))
}
func (BytecodeExp) round(v1 *BytecodeValue, v2 BytecodeValue) {
	shift := math.Pow(10, v2.v)
	v1.SetF(float32(math.Floor((v1.v*shift)+0.5) / shift))
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
			if c = c.enemyNear(sys.bcStack.Pop().ToI()); c != nil {
				i += 4
				continue
			}
			sys.bcStack.Push(BytecodeSF())
			i += int(*(*int32)(unsafe.Pointer(&be[i]))) + 4
		case OC_playerid:
			if c = sys.playerID(sys.bcStack.Pop().ToI()); c != nil {
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
			sys.bcStack.PushI(int32(int8(be[i])))
			i++
		case OC_int:
			sys.bcStack.PushI(*(*int32)(unsafe.Pointer(&be[i])))
			i += 4
		case OC_float:
			sys.bcStack.PushF(*(*float32)(unsafe.Pointer(&be[i])))
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
		case OC_ailevel:
			sys.bcStack.PushI(int32(c.aiLevel()))
		case OC_alive:
			sys.bcStack.PushB(c.alive())
		case OC_anim:
			sys.bcStack.PushI(c.animNo)
		case OC_animelemno:
			*sys.bcStack.Top() = c.animElemNo(sys.bcStack.Top().ToI())
		case OC_animelemtime:
			*sys.bcStack.Top() = c.animElemTime(sys.bcStack.Top().ToI())
		case OC_animexist:
			*sys.bcStack.Top() = c.animExist(sys.workingChar, *sys.bcStack.Top())
		case OC_animtime:
			sys.bcStack.PushI(c.animTime())
		case OC_backedge:
			sys.bcStack.PushF(c.backEdge())
		case OC_backedgebodydist:
			sys.bcStack.PushI(int32(c.backEdgeBodyDist()))
		case OC_backedgedist:
			sys.bcStack.PushI(int32(c.backEdgeDist()))
		case OC_bottomedge:
			sys.bcStack.PushF(c.bottomEdge())
		case OC_camerapos_x:
			sys.bcStack.PushF(sys.cam.Pos[0] / oc.localscl)
		case OC_camerapos_y:
			sys.bcStack.PushF(sys.cam.Pos[1] / oc.localscl)
		case OC_camerazoom:
			sys.bcStack.PushF(sys.cam.Scale)
		case OC_canrecover:
			sys.bcStack.PushB(c.canRecover())
		case OC_command:
			sys.bcStack.PushB(c.command(sys.workingState.playerNo,
				int(*(*int32)(unsafe.Pointer(&be[i])))))
			i += 4
		case OC_ctrl:
			sys.bcStack.PushB(c.ctrl())
		case OC_facing:
			sys.bcStack.PushI(int32(c.facing))
		case OC_frontedge:
			sys.bcStack.PushF(c.frontEdge())
		case OC_frontedgebodydist:
			sys.bcStack.PushI(int32(c.frontEdgeBodyDist()))
		case OC_frontedgedist:
			sys.bcStack.PushI(int32(c.frontEdgeDist()))
		case OC_gameheight:
			if c.gi().ver[0] == 1 && c.gi().ver[1] == 0 {
				sys.bcStack.PushF(sys.screenHeight() / oc.localscl)
			} else {
				sys.bcStack.PushF(c.gameHeight())
			}
		case OC_gametime:
			sys.bcStack.PushI(sys.gameTime)
		case OC_gamewidth:
			if c.gi().ver[0] == 1 && c.gi().ver[1] == 0 {
				sys.bcStack.PushF(sys.screenWidth() / oc.localscl)
			} else {
				sys.bcStack.PushF(c.gameWidth())
			}
		case OC_hitcount:
			sys.bcStack.PushI(c.hitCount)
		case OC_hitdefattr:
			sys.bcStack.PushB(c.hitDefAttr(*(*int32)(unsafe.Pointer(&be[i]))))
			i += 4
		case OC_hitfall:
			sys.bcStack.PushB(c.ghv.fallf)
		case OC_hitover:
			sys.bcStack.PushB(c.hitOver())
		case OC_hitpausetime:
			sys.bcStack.PushI(c.hitPauseTime)
		case OC_hitshakeover:
			sys.bcStack.PushB(c.hitShakeOver())
		case OC_hitvel_x:
			sys.bcStack.PushF(c.hitVelX() * c.localscl / oc.localscl)
		case OC_hitvel_y:
			sys.bcStack.PushF(c.hitVelY() * c.localscl / oc.localscl)
		case OC_id:
			sys.bcStack.PushI(c.id)
		case OC_inguarddist:
			sys.bcStack.PushB(c.inguarddist)
		case OC_ishelper:
			*sys.bcStack.Top() = c.isHelper(*sys.bcStack.Top())
		case OC_leftedge:
			sys.bcStack.PushF(c.leftEdge())
		case OC_life:
			sys.bcStack.PushI(c.life)
		case OC_lifemax:
			sys.bcStack.PushI(c.lifeMax)
		case OC_movecontact:
			sys.bcStack.PushI(c.moveContact())
		case OC_moveguarded:
			sys.bcStack.PushI(c.moveGuarded())
		case OC_movehit:
			sys.bcStack.PushI(c.moveHit())
		case OC_movereversed:
			sys.bcStack.PushI(c.moveReversed())
		case OC_movetype:
			sys.bcStack.PushB(c.ss.moveType == MoveType(be[i])<<15)
			i++
		case OC_numenemy:
			sys.bcStack.PushI(c.numEnemy())
		case OC_numexplod:
			*sys.bcStack.Top() = c.numExplod(*sys.bcStack.Top())
		case OC_numhelper:
			*sys.bcStack.Top() = c.numHelper(*sys.bcStack.Top())
		case OC_numpartner:
			sys.bcStack.PushI(c.numPartner())
		case OC_numproj:
			sys.bcStack.PushI(c.numProj())
		case OC_numprojid:
			*sys.bcStack.Top() = c.numProjID(*sys.bcStack.Top())
		case OC_numtarget:
			*sys.bcStack.Top() = c.numTarget(*sys.bcStack.Top())
		case OC_palno:
			sys.bcStack.PushI(c.palno())
		case OC_pos_x:
			sys.bcStack.PushF((c.pos[0]*c.localscl/oc.localscl - sys.cam.Pos[0]/oc.localscl))
		case OC_pos_y:
			sys.bcStack.PushF((c.pos[1] - c.platformPosY) * c.localscl / oc.localscl)
		case OC_power:
			sys.bcStack.PushI(c.getPower())
		case OC_powermax:
			sys.bcStack.PushI(c.powerMax)
		case OC_playeridexist:
			*sys.bcStack.Top() = sys.playerIDExist(*sys.bcStack.Top())
		case OC_prevstateno:
			sys.bcStack.PushI(c.ss.prevno)
		case OC_projcanceltime:
			*sys.bcStack.Top() = c.projCancelTime(*sys.bcStack.Top())
		case OC_projcontacttime:
			*sys.bcStack.Top() = c.projContactTime(*sys.bcStack.Top())
		case OC_projguardedtime:
			*sys.bcStack.Top() = c.projGuardedTime(*sys.bcStack.Top())
		case OC_projhittime:
			*sys.bcStack.Top() = c.projHitTime(*sys.bcStack.Top())
		case OC_random:
			sys.bcStack.PushI(Rand(0, 999))
		case OC_rightedge:
			sys.bcStack.PushF(c.rightEdge())
		case OC_roundsexisted:
			sys.bcStack.PushI(c.roundsExisted())
		case OC_roundstate:
			sys.bcStack.PushI(c.roundState())
		case OC_screenheight:
			sys.bcStack.PushF(sys.screenHeight() / oc.localscl)
		case OC_screenpos_x:
			sys.bcStack.PushF((c.screenPosX()) / oc.localscl)
		case OC_screenpos_y:
			sys.bcStack.PushF((c.screenPosY()) / oc.localscl)
		case OC_screenwidth:
			sys.bcStack.PushF(sys.screenWidth() / oc.localscl)
		case OC_selfanimexist:
			*sys.bcStack.Top() = c.selfAnimExist(*sys.bcStack.Top())
		case OC_stateno:
			sys.bcStack.PushI(c.ss.no)
		case OC_statetype:
			sys.bcStack.PushB(c.ss.stateType == StateType(be[i]))
			i++
		case OC_teammode:
			if c.teamside == -1 {
				sys.bcStack.PushB(TM_Single == TeamMode(be[i]))
			} else {
				sys.bcStack.PushB(sys.tmode[c.playerNo&1] == TeamMode(be[i]))
			}
			i++
		case OC_teamside:
			sys.bcStack.PushI(int32(c.teamside) + 1)
		case OC_time:
			sys.bcStack.PushI(c.time())
		case OC_topedge:
			sys.bcStack.PushF(c.topEdge())
		case OC_uniqhitcount:
			sys.bcStack.PushI(c.uniqHitCount)
		case OC_vel_x:
			sys.bcStack.PushF(c.vel[0] * c.localscl / oc.localscl)
		case OC_vel_y:
			sys.bcStack.PushF(c.vel[1] * c.localscl / oc.localscl)
		case OC_st_:
			be.run_st(c, &i)
		case OC_const_:
			be.run_const(c, &i, oc)
		case OC_ex_:
			be.run_ex(c, &i, oc)
		case OC_var:
			*sys.bcStack.Top() = c.varGet(sys.bcStack.Top().ToI())
		case OC_sysvar:
			*sys.bcStack.Top() = c.sysVarGet(sys.bcStack.Top().ToI())
		case OC_fvar:
			*sys.bcStack.Top() = c.fvarGet(sys.bcStack.Top().ToI())
		case OC_sysfvar:
			*sys.bcStack.Top() = c.sysFvarGet(sys.bcStack.Top().ToI())
		case OC_localvar:
			sys.bcStack.Push(sys.bcVar[uint8(be[i])])
			i++
		default:
			vi := be[i-1]
			if vi < OC_sysvar0+NumSysVar {
				sys.bcStack.PushI(c.ivar[vi-OC_var0])
			} else {
				sys.bcStack.PushF(c.fvar[vi-OC_fvar0])
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
			sys.errLog.Printf("%v\n", be[*i-1])
			c.panic()
		}
	}
}
func (be BytecodeExp) run_const(c *Char, i *int, oc *Char) {
	(*i)++
	switch be[*i-1] {
	case OC_const_data_life:
		sys.bcStack.PushI(c.gi().data.life)
	case OC_const_data_power:
		sys.bcStack.PushI(c.gi().data.power)
	case OC_const_data_dizzypoints:
		sys.bcStack.PushI(c.gi().data.dizzypoints)
	case OC_const_data_guardpoints:
		sys.bcStack.PushI(c.gi().data.guardpoints)
	case OC_const_data_attack:
		sys.bcStack.PushI(c.gi().data.attack)
	case OC_const_data_defence:
		sys.bcStack.PushI(c.gi().data.defence)
	case OC_const_data_fall_defence_mul:
		sys.bcStack.PushF(c.gi().data.fall.defence_mul)
	case OC_const_data_liedown_time:
		sys.bcStack.PushI(c.gi().data.liedown.time)
	case OC_const_data_airjuggle:
		sys.bcStack.PushI(c.gi().data.airjuggle)
	case OC_const_data_sparkno:
		sys.bcStack.PushI(c.gi().data.sparkno)
	case OC_const_data_guard_sparkno:
		sys.bcStack.PushI(c.gi().data.guard.sparkno)
	case OC_const_data_ko_echo:
		sys.bcStack.PushI(c.gi().data.ko.echo)
	case OC_const_data_intpersistindex:
		sys.bcStack.PushI(c.gi().data.intpersistindex)
	case OC_const_data_floatpersistindex:
		sys.bcStack.PushI(c.gi().data.floatpersistindex)
	case OC_const_size_xscale:
		sys.bcStack.PushF(c.size.xscale)
	case OC_const_size_yscale:
		sys.bcStack.PushF(c.size.yscale)
	case OC_const_size_ground_back:
		sys.bcStack.PushF(c.size.ground.back * (320 / float32(c.localcoord)) / oc.localscl)
	case OC_const_size_ground_front:
		sys.bcStack.PushF(c.size.ground.front * (320 / float32(c.localcoord)) / oc.localscl)
	case OC_const_size_air_back:
		sys.bcStack.PushF(c.size.air.back * (320 / float32(c.localcoord)) / oc.localscl)
	case OC_const_size_air_front:
		sys.bcStack.PushF(c.size.air.front * (320 / float32(c.localcoord)) / oc.localscl)
	case OC_const_size_z_width:
		sys.bcStack.PushF(c.size.z.width * (320 / float32(c.localcoord)) / oc.localscl)
	case OC_const_size_height:
		sys.bcStack.PushF(c.size.height * (320 / float32(c.localcoord)) / oc.localscl)
	case OC_const_size_attack_dist:
		sys.bcStack.PushF(c.size.attack.dist * (320 / float32(c.localcoord)) / oc.localscl)
	case OC_const_size_attack_z_width_back:
		sys.bcStack.PushF(c.size.attack.z.width[1] * (320 / float32(c.localcoord)) / oc.localscl)
	case OC_const_size_attack_z_width_front:
		sys.bcStack.PushF(c.size.attack.z.width[0] * (320 / float32(c.localcoord)) / oc.localscl)
	case OC_const_size_proj_attack_dist:
		sys.bcStack.PushF(c.size.proj.attack.dist * (320 / float32(c.localcoord)) / oc.localscl)
	case OC_const_size_proj_doscale:
		sys.bcStack.PushI(c.size.proj.doscale)
	case OC_const_size_head_pos_x:
		sys.bcStack.PushF(c.size.head.pos[0] * (320 / float32(c.localcoord)) / oc.localscl)
	case OC_const_size_head_pos_y:
		sys.bcStack.PushF(c.size.head.pos[1] * (320 / float32(c.localcoord)) / oc.localscl)
	case OC_const_size_mid_pos_x:
		sys.bcStack.PushF(c.size.mid.pos[0] * (320 / float32(c.localcoord)) / oc.localscl)
	case OC_const_size_mid_pos_y:
		sys.bcStack.PushF(c.size.mid.pos[1] * (320 / float32(c.localcoord)) / oc.localscl)
	case OC_const_size_shadowoffset:
		sys.bcStack.PushF(c.size.shadowoffset * (320 / float32(c.localcoord)) / oc.localscl)
	case OC_const_size_draw_offset_x:
		sys.bcStack.PushF(c.size.draw.offset[0] * (320 / float32(c.localcoord)) / oc.localscl)
	case OC_const_size_draw_offset_y:
		sys.bcStack.PushF(c.size.draw.offset[1] * (320 / float32(c.localcoord)) / oc.localscl)
	case OC_const_velocity_walk_fwd_x:
		sys.bcStack.PushF(c.gi().velocity.walk.fwd * (320 / float32(c.localcoord)) / oc.localscl)
	case OC_const_velocity_walk_back_x:
		sys.bcStack.PushF(c.gi().velocity.walk.back * (320 / float32(c.localcoord)) / oc.localscl)
	case OC_const_velocity_walk_up_x:
		sys.bcStack.PushF(c.gi().velocity.walk.up.x * (320 / float32(c.localcoord)) / oc.localscl)
	case OC_const_velocity_walk_down_x:
		sys.bcStack.PushF(c.gi().velocity.walk.down.x * (320 / float32(c.localcoord)) / oc.localscl)
	case OC_const_velocity_run_fwd_x:
		sys.bcStack.PushF(c.gi().velocity.run.fwd[0] * (320 / float32(c.localcoord)) / oc.localscl)
	case OC_const_velocity_run_fwd_y:
		sys.bcStack.PushF(c.gi().velocity.run.fwd[1] * (320 / float32(c.localcoord)) / oc.localscl)
	case OC_const_velocity_run_back_x:
		sys.bcStack.PushF(c.gi().velocity.run.back[0] * (320 / float32(c.localcoord)) / oc.localscl)
	case OC_const_velocity_run_back_y:
		sys.bcStack.PushF(c.gi().velocity.run.back[1] * (320 / float32(c.localcoord)) / oc.localscl)
	case OC_const_velocity_run_up_x:
		sys.bcStack.PushF(c.gi().velocity.run.up.x * (320 / float32(c.localcoord)) / oc.localscl)
	case OC_const_velocity_run_up_y:
		sys.bcStack.PushF(c.gi().velocity.run.up.y * (320 / float32(c.localcoord)) / oc.localscl)
	case OC_const_velocity_run_down_x:
		sys.bcStack.PushF(c.gi().velocity.run.down.x * (320 / float32(c.localcoord)) / oc.localscl)
	case OC_const_velocity_run_down_y:
		sys.bcStack.PushF(c.gi().velocity.run.down.y * (320 / float32(c.localcoord)) / oc.localscl)
	case OC_const_velocity_jump_y:
		sys.bcStack.PushF(c.gi().velocity.jump.neu[1] * (320 / float32(c.localcoord)) / oc.localscl)
	case OC_const_velocity_jump_neu_x:
		sys.bcStack.PushF(c.gi().velocity.jump.neu[0] * (320 / float32(c.localcoord)) / oc.localscl)
	case OC_const_velocity_jump_back_x:
		sys.bcStack.PushF(c.gi().velocity.jump.back * (320 / float32(c.localcoord)) / oc.localscl)
	case OC_const_velocity_jump_fwd_x:
		sys.bcStack.PushF(c.gi().velocity.jump.fwd * (320 / float32(c.localcoord)) / oc.localscl)
	case OC_const_velocity_jump_up_x:
		sys.bcStack.PushF(c.gi().velocity.jump.up.x * (320 / float32(c.localcoord)) / oc.localscl)
	case OC_const_velocity_jump_down_x:
		sys.bcStack.PushF(c.gi().velocity.jump.down.x * (320 / float32(c.localcoord)) / oc.localscl)
	case OC_const_velocity_runjump_back_x:
		sys.bcStack.PushF(c.gi().velocity.runjump.back[0] * (320 / float32(c.localcoord)) / oc.localscl)
	case OC_const_velocity_runjump_back_y:
		sys.bcStack.PushF(c.gi().velocity.runjump.back[1] * (320 / float32(c.localcoord)) / oc.localscl)
	case OC_const_velocity_runjump_y:
		sys.bcStack.PushF(c.gi().velocity.runjump.fwd[1] * (320 / float32(c.localcoord)) / oc.localscl)
	case OC_const_velocity_runjump_fwd_x:
		sys.bcStack.PushF(c.gi().velocity.runjump.fwd[0] * (320 / float32(c.localcoord)) / oc.localscl)
	case OC_const_velocity_runjump_up_x:
		sys.bcStack.PushF(c.gi().velocity.runjump.up.x * (320 / float32(c.localcoord)) / oc.localscl)
	case OC_const_velocity_runjump_down_x:
		sys.bcStack.PushF(c.gi().velocity.runjump.down.x * (320 / float32(c.localcoord)) / oc.localscl)
	case OC_const_velocity_airjump_y:
		sys.bcStack.PushF(c.gi().velocity.airjump.neu[1] * (320 / float32(c.localcoord)) / oc.localscl)
	case OC_const_velocity_airjump_neu_x:
		sys.bcStack.PushF(c.gi().velocity.airjump.neu[0] * (320 / float32(c.localcoord)) / oc.localscl)
	case OC_const_velocity_airjump_back_x:
		sys.bcStack.PushF(c.gi().velocity.airjump.back * (320 / float32(c.localcoord)) / oc.localscl)
	case OC_const_velocity_airjump_fwd_x:
		sys.bcStack.PushF(c.gi().velocity.airjump.fwd * (320 / float32(c.localcoord)) / oc.localscl)
	case OC_const_velocity_airjump_up_x:
		sys.bcStack.PushF(c.gi().velocity.airjump.up.x * (320 / float32(c.localcoord)) / oc.localscl)
	case OC_const_velocity_airjump_down_x:
		sys.bcStack.PushF(c.gi().velocity.airjump.down.x * (320 / float32(c.localcoord)) / oc.localscl)
	case OC_const_velocity_air_gethit_groundrecover_x:
		sys.bcStack.PushF(c.gi().velocity.air.gethit.groundrecover[0] * (320 / float32(c.localcoord)) / oc.localscl)
	case OC_const_velocity_air_gethit_groundrecover_y:
		sys.bcStack.PushF(c.gi().velocity.air.gethit.groundrecover[1] * (320 / float32(c.localcoord)) / oc.localscl)
	case OC_const_velocity_air_gethit_airrecover_mul_x:
		sys.bcStack.PushF(c.gi().velocity.air.gethit.airrecover.mul[0])
	case OC_const_velocity_air_gethit_airrecover_mul_y:
		sys.bcStack.PushF(c.gi().velocity.air.gethit.airrecover.mul[1])
	case OC_const_velocity_air_gethit_airrecover_add_x:
		sys.bcStack.PushF(c.gi().velocity.air.gethit.airrecover.add[0] * (320 / float32(c.localcoord)) / oc.localscl)
	case OC_const_velocity_air_gethit_airrecover_add_y:
		sys.bcStack.PushF(c.gi().velocity.air.gethit.airrecover.add[1] * (320 / float32(c.localcoord)) / oc.localscl)
	case OC_const_velocity_air_gethit_airrecover_back:
		sys.bcStack.PushF(c.gi().velocity.air.gethit.airrecover.back * (320 / float32(c.localcoord)) / oc.localscl)
	case OC_const_velocity_air_gethit_airrecover_fwd:
		sys.bcStack.PushF(c.gi().velocity.air.gethit.airrecover.fwd * (320 / float32(c.localcoord)) / oc.localscl)
	case OC_const_velocity_air_gethit_airrecover_up:
		sys.bcStack.PushF(c.gi().velocity.air.gethit.airrecover.up * (320 / float32(c.localcoord)) / oc.localscl)
	case OC_const_velocity_air_gethit_airrecover_down:
		sys.bcStack.PushF(c.gi().velocity.air.gethit.airrecover.down * (320 / float32(c.localcoord)) / oc.localscl)
	case OC_const_movement_airjump_num:
		sys.bcStack.PushI(c.gi().movement.airjump.num)
	case OC_const_movement_airjump_height:
		sys.bcStack.PushI(int32(float32(c.gi().movement.airjump.height) * (320 / float32(c.localcoord)) / oc.localscl))
	case OC_const_movement_yaccel:
		sys.bcStack.PushF(c.gi().movement.yaccel * (320 / float32(c.localcoord)) / oc.localscl)
	case OC_const_movement_stand_friction:
		sys.bcStack.PushF(c.gi().movement.stand.friction)
	case OC_const_movement_crouch_friction:
		sys.bcStack.PushF(c.gi().movement.crouch.friction)
	case OC_const_movement_stand_friction_threshold:
		sys.bcStack.PushF(c.gi().movement.stand.friction_threshold * (320 / float32(c.localcoord)) / oc.localscl)
	case OC_const_movement_crouch_friction_threshold:
		sys.bcStack.PushF(c.gi().movement.crouch.friction_threshold * (320 / float32(c.localcoord)) / oc.localscl)
	case OC_const_movement_air_gethit_groundlevel:
		sys.bcStack.PushF(c.gi().movement.air.gethit.groundlevel * (320 / float32(c.localcoord)) / oc.localscl)
	case OC_const_movement_air_gethit_groundrecover_ground_threshold:
		sys.bcStack.PushF(
			c.gi().movement.air.gethit.groundrecover.ground.threshold * (320 / float32(c.localcoord)) / oc.localscl)
	case OC_const_movement_air_gethit_groundrecover_groundlevel:
		sys.bcStack.PushF(c.gi().movement.air.gethit.groundrecover.groundlevel * (320 / float32(c.localcoord)) / oc.localscl)
	case OC_const_movement_air_gethit_airrecover_threshold:
		sys.bcStack.PushF(c.gi().movement.air.gethit.airrecover.threshold * (320 / float32(c.localcoord)) / oc.localscl)
	case OC_const_movement_air_gethit_airrecover_yaccel:
		sys.bcStack.PushF(c.gi().movement.air.gethit.airrecover.yaccel * (320 / float32(c.localcoord)) / oc.localscl)
	case OC_const_movement_air_gethit_trip_groundlevel:
		sys.bcStack.PushF(c.gi().movement.air.gethit.trip.groundlevel * (320 / float32(c.localcoord)) / oc.localscl)
	case OC_const_movement_down_bounce_offset_x:
		sys.bcStack.PushF(c.gi().movement.down.bounce.offset[0] * (320 / float32(c.localcoord)) / oc.localscl)
	case OC_const_movement_down_bounce_offset_y:
		sys.bcStack.PushF(c.gi().movement.down.bounce.offset[1] * (320 / float32(c.localcoord)) / oc.localscl)
	case OC_const_movement_down_bounce_yaccel:
		sys.bcStack.PushF(c.gi().movement.down.bounce.yaccel * (320 / float32(c.localcoord)) / oc.localscl)
	case OC_const_movement_down_bounce_groundlevel:
		sys.bcStack.PushF(c.gi().movement.down.bounce.groundlevel * (320 / float32(c.localcoord)) / oc.localscl)
	case OC_const_movement_down_friction_threshold:
		sys.bcStack.PushF(c.gi().movement.down.friction_threshold * (320 / float32(c.localcoord)) / oc.localscl)
	case OC_const_authorname:
		sys.bcStack.PushB(c.gi().authorLow ==
			sys.stringPool[sys.workingState.playerNo].List[*(*int32)(
				unsafe.Pointer(&be[*i]))])
		*i += 4
	case OC_const_name:
		sys.bcStack.PushB(c.gi().nameLow ==
			sys.stringPool[sys.workingState.playerNo].List[*(*int32)(
				unsafe.Pointer(&be[*i]))])
		*i += 4
	case OC_const_p2name:
		p2 := c.p2()
		sys.bcStack.PushB(p2 != nil && p2.gi().nameLow ==
			sys.stringPool[sys.workingState.playerNo].List[*(*int32)(
				unsafe.Pointer(&be[*i]))])
		*i += 4
	case OC_const_p3name:
		p3 := c.partner(0)
		sys.bcStack.PushB(p3 != nil && p3.gi().nameLow ==
			sys.stringPool[sys.workingState.playerNo].List[*(*int32)(
				unsafe.Pointer(&be[*i]))])
		*i += 4
	case OC_const_p4name:
		p4 := sys.charList.enemyNear(c, 1, true, false)
		sys.bcStack.PushB(p4 != nil && !(p4.scf(SCF_ko) && p4.scf(SCF_over)) &&
			p4.gi().nameLow ==
				sys.stringPool[sys.workingState.playerNo].List[*(*int32)(
					unsafe.Pointer(&be[*i]))])
		*i += 4
	case OC_const_p5name:
		p5 := c.partner(1)
		sys.bcStack.PushB(p5 != nil && p5.gi().nameLow ==
			sys.stringPool[sys.workingState.playerNo].List[*(*int32)(
				unsafe.Pointer(&be[*i]))])
		*i += 4
	case OC_const_p6name:
		p6 := sys.charList.enemyNear(c, 2, true, false)
		sys.bcStack.PushB(p6 != nil && !(p6.scf(SCF_ko) && p6.scf(SCF_over)) &&
			p6.gi().nameLow ==
				sys.stringPool[sys.workingState.playerNo].List[*(*int32)(
					unsafe.Pointer(&be[*i]))])
		*i += 4
	case OC_const_p7name:
		p7 := c.partner(2)
		sys.bcStack.PushB(p7 != nil && p7.gi().nameLow ==
			sys.stringPool[sys.workingState.playerNo].List[*(*int32)(
				unsafe.Pointer(&be[*i]))])
		*i += 4
	case OC_const_p8name:
		p8 := sys.charList.enemyNear(c, 3, true, false)
		sys.bcStack.PushB(p8 != nil && !(p8.scf(SCF_ko) && p8.scf(SCF_over)) &&
			p8.gi().nameLow ==
				sys.stringPool[sys.workingState.playerNo].List[*(*int32)(
					unsafe.Pointer(&be[*i]))])
		*i += 4
	case OC_const_stagevar_info_name:
		sys.bcStack.PushB(sys.stage.nameLow ==
			sys.stringPool[sys.workingState.playerNo].List[*(*int32)(
				unsafe.Pointer(&be[*i]))])
		*i += 4
	case OC_const_stagevar_info_displayname:
		sys.bcStack.PushB(sys.stage.displaynameLow ==
			sys.stringPool[sys.workingState.playerNo].List[*(*int32)(
				unsafe.Pointer(&be[*i]))])
		*i += 4
	case OC_const_stagevar_info_author:
		sys.bcStack.PushB(sys.stage.authorLow ==
			sys.stringPool[sys.workingState.playerNo].List[*(*int32)(
				unsafe.Pointer(&be[*i]))])
		*i += 4
	case OC_const_constants:
		sys.bcStack.PushF(c.gi().constants[sys.stringPool[sys.workingState.playerNo].List[*(*int32)(
			unsafe.Pointer(&be[*i]))]])
		*i += 4
	case OC_const_stage_constants:
		sys.bcStack.PushF(sys.stage.constants[sys.stringPool[sys.workingState.playerNo].List[*(*int32)(
			unsafe.Pointer(&be[*i]))]])
		*i += 4
	default:
		sys.errLog.Printf("%v\n", be[*i-1])
		c.panic()
	}
}
func (be BytecodeExp) run_ex(c *Char, i *int, oc *Char) {
	(*i)++
	switch be[*i-1] {
	case OC_ex_p2dist_x:
		sys.bcStack.Push(c.rdDistX(c.p2(), oc))
	case OC_ex_p2dist_y:
		sys.bcStack.Push(c.rdDistY(c.p2(), oc))
	case OC_ex_p2bodydist_x:
		sys.bcStack.Push(c.p2BodyDistX(oc))
	case OC_ex_parentdist_x:
		sys.bcStack.Push(c.rdDistX(c.parent(), oc))
	case OC_ex_parentdist_y:
		sys.bcStack.Push(c.rdDistY(c.parent(), oc))
	case OC_ex_rootdist_x:
		sys.bcStack.Push(c.rdDistX(c.root(), oc))
	case OC_ex_rootdist_y:
		sys.bcStack.Push(c.rdDistY(c.root(), oc))
	case OC_ex_win:
		sys.bcStack.PushB(c.win())
	case OC_ex_winko:
		sys.bcStack.PushB(c.winKO())
	case OC_ex_wintime:
		sys.bcStack.PushB(c.winTime())
	case OC_ex_winperfect:
		sys.bcStack.PushB(c.winPerfect())
	case OC_ex_winspecial:
		sys.bcStack.PushB(c.winType(WT_S))
	case OC_ex_winhyper:
		sys.bcStack.PushB(c.winType(WT_H))
	case OC_ex_lose:
		sys.bcStack.PushB(c.lose())
	case OC_ex_loseko:
		sys.bcStack.PushB(c.loseKO())
	case OC_ex_losetime:
		sys.bcStack.PushB(c.loseTime())
	case OC_ex_drawgame:
		sys.bcStack.PushB(c.drawgame())
	case OC_ex_matchover:
		sys.bcStack.PushB(sys.matchOver())
	case OC_ex_matchno:
		sys.bcStack.PushI(sys.match)
	case OC_ex_roundno:
		sys.bcStack.PushI(sys.round)
	case OC_ex_ishometeam:
		sys.bcStack.PushB(c.teamside == sys.home)
	case OC_ex_tickspersecond:
		sys.bcStack.PushI(int32(FPS))
	case OC_ex_const240p:
		*sys.bcStack.Top() = c.constp(320, sys.bcStack.Top().ToF())
	case OC_ex_const480p:
		*sys.bcStack.Top() = c.constp(640, sys.bcStack.Top().ToF())
	case OC_ex_const720p:
		*sys.bcStack.Top() = c.constp(960, sys.bcStack.Top().ToF())
	case OC_ex_gethitvar_animtype:
		sys.bcStack.PushI(int32(c.gethitAnimtype()))
	case OC_ex_gethitvar_airtype:
		sys.bcStack.PushI(int32(c.ghv.airtype))
	case OC_ex_gethitvar_groundtype:
		sys.bcStack.PushI(int32(c.ghv.groundtype))
	case OC_ex_gethitvar_damage:
		sys.bcStack.PushI(c.ghv.damage)
	case OC_ex_gethitvar_hitcount:
		sys.bcStack.PushI(c.ghv.hitcount)
	case OC_ex_gethitvar_fallcount:
		sys.bcStack.PushI(c.ghv.fallcount)
	case OC_ex_gethitvar_hitshaketime:
		sys.bcStack.PushI(c.ghv.hitshaketime)
	case OC_ex_gethitvar_hittime:
		sys.bcStack.PushI(c.ghv.hittime)
	case OC_ex_gethitvar_slidetime:
		sys.bcStack.PushI(c.ghv.slidetime)
	case OC_ex_gethitvar_ctrltime:
		sys.bcStack.PushI(c.ghv.ctrltime)
	case OC_ex_gethitvar_recovertime:
		sys.bcStack.PushI(c.recoverTime)
	case OC_ex_gethitvar_xoff:
		sys.bcStack.PushF(c.ghv.xoff * c.localscl / oc.localscl)
	case OC_ex_gethitvar_yoff:
		sys.bcStack.PushF(c.ghv.yoff * c.localscl / oc.localscl)
	case OC_ex_gethitvar_xvel:
		sys.bcStack.PushF(c.ghv.xvel * c.facing * c.localscl / oc.localscl)
	case OC_ex_gethitvar_yvel:
		sys.bcStack.PushF(c.ghv.yvel * c.localscl / oc.localscl)
	case OC_ex_gethitvar_yaccel:
		sys.bcStack.PushF(c.ghv.getYaccel(oc) * c.localscl / oc.localscl)
	case OC_ex_gethitvar_chainid:
		sys.bcStack.PushI(c.ghv.chainId())
	case OC_ex_gethitvar_guarded:
		sys.bcStack.PushB(c.ghv.guarded)
	case OC_ex_gethitvar_isbound:
		sys.bcStack.PushB(c.isBound())
	case OC_ex_gethitvar_fall:
		sys.bcStack.PushB(c.ghv.fallf)
	case OC_ex_gethitvar_fall_damage:
		sys.bcStack.PushI(c.ghv.fall.damage)
	case OC_ex_gethitvar_fall_xvel:
		sys.bcStack.PushF(c.ghv.fall.xvel() * c.localscl / oc.localscl)
	case OC_ex_gethitvar_fall_yvel:
		sys.bcStack.PushF(c.ghv.fall.yvelocity * c.localscl / oc.localscl)
	case OC_ex_gethitvar_fall_recover:
		sys.bcStack.PushB(c.ghv.fall.recover)
	case OC_ex_gethitvar_fall_time:
		sys.bcStack.PushI(c.fallTime)
	case OC_ex_gethitvar_fall_recovertime:
		sys.bcStack.PushI(c.ghv.fall.recovertime)
	case OC_ex_gethitvar_fall_kill:
		sys.bcStack.PushB(c.ghv.fall.kill)
	case OC_ex_gethitvar_fall_envshake_time:
		sys.bcStack.PushI(c.ghv.fall.envshake_time)
	case OC_ex_gethitvar_fall_envshake_freq:
		sys.bcStack.PushF(c.ghv.fall.envshake_freq)
	case OC_ex_gethitvar_fall_envshake_ampl:
		sys.bcStack.PushI(int32(float32(c.ghv.fall.envshake_ampl) * c.localscl / oc.localscl))
	case OC_ex_gethitvar_fall_envshake_phase:
		sys.bcStack.PushF(c.ghv.fall.envshake_phase * c.localscl / oc.localscl)
	case OC_ex_gethitvar_attr:
		sys.bcStack.PushI(c.ghv.attr)
	case OC_ex_gethitvar_dizzypoints:
		sys.bcStack.PushI(c.ghv.dizzypoints)
	case OC_ex_gethitvar_guardpoints:
		sys.bcStack.PushI(c.ghv.guardpoints)
	case OC_ex_gethitvar_id:
		sys.bcStack.PushI(c.ghv.id)
	case OC_ex_gethitvar_playerno:
		sys.bcStack.PushI(int32(c.ghv.playerNo) + 1)
	case OC_ex_gethitvar_redlife:
		sys.bcStack.PushI(c.ghv.redlife)
	case OC_ex_gethitvar_score:
		sys.bcStack.PushF(c.ghv.score)
	case OC_ex_gethitvar_hitdamage:
		sys.bcStack.PushI(c.ghv.hitdamage)
	case OC_ex_gethitvar_guarddamage:
		sys.bcStack.PushI(c.ghv.guarddamage)
	case OC_ex_gethitvar_hitpower:
		sys.bcStack.PushI(c.ghv.hitpower)
	case OC_ex_gethitvar_guardpower:
		sys.bcStack.PushI(c.ghv.guardpower)
	case OC_ex_ailevelf:
		sys.bcStack.PushF(c.aiLevel())
	case OC_ex_animelemlength:
		if f := c.anim.CurrentFrame(); f != nil {
			sys.bcStack.PushI(f.Time)
		} else {
			sys.bcStack.PushI(0)
		}
	case OC_ex_animlength:
		sys.bcStack.PushI(c.anim.totaltime)
	case OC_ex_combocount:
		sys.bcStack.PushI(c.comboCount())
	case OC_ex_consecutivewins:
		sys.bcStack.PushI(c.consecutiveWins())
	case OC_ex_dizzy:
		sys.bcStack.PushB(c.scf(SCF_dizzy))
	case OC_ex_dizzypoints:
		sys.bcStack.PushI(c.dizzyPoints)
	case OC_ex_dizzypointsmax:
		sys.bcStack.PushI(c.dizzyPointsMax)
	case OC_ex_drawpalno:
		sys.bcStack.PushI(c.gi().drawpalno)
	case OC_ex_firstattack:
		sys.bcStack.PushB(c.firstAttack)
	case OC_ex_float:
		*sys.bcStack.Top() = BytecodeFloat(sys.bcStack.Top().ToF())
	case OC_ex_gamemode:
		sys.bcStack.PushB(strings.ToLower(sys.gameMode) ==
			sys.stringPool[sys.workingState.playerNo].List[*(*int32)(
				unsafe.Pointer(&be[*i]))])
		*i += 4
	case OC_ex_getplayerid:
		sys.bcStack.Top().SetI(c.getPlayerID(int(sys.bcStack.Top().ToI())))
	case OC_ex_groundangle:
		sys.bcStack.PushF(c.groundAngle)
	case OC_ex_guardbreak:
		sys.bcStack.PushB(c.scf(SCF_guardbreak))
	case OC_ex_guardpoints:
		sys.bcStack.PushI(c.guardPoints)
	case OC_ex_guardpointsmax:
		sys.bcStack.PushI(c.guardPointsMax)
	case OC_ex_helpername:
		sys.bcStack.PushB(c.helperIndex != 0 && strings.ToLower(c.name) ==
			sys.stringPool[sys.workingState.playerNo].List[*(*int32)(
				unsafe.Pointer(&be[*i]))])
		*i += 4
	case OC_ex_hitoverridden:
		sys.bcStack.PushB(c.hoIdx >= 0)
	case OC_ex_incustomstate:
		sys.bcStack.PushB(c.ss.sb.playerNo != c.playerNo)
	case OC_ex_indialogue:
		sys.bcStack.PushB(sys.dialogueFlg)
	case OC_ex_isassertedchar:
		sys.bcStack.PushB(c.sf(CharSpecialFlag((*(*int32)(unsafe.Pointer(&be[*i]))))))
		*i += 4
	case OC_ex_isassertedglobal:
		sys.bcStack.PushB(sys.sf(GlobalSpecialFlag((*(*int32)(unsafe.Pointer(&be[*i]))))))
		*i += 4
	case OC_ex_ishost:
		sys.bcStack.PushB(c.isHost())
	case OC_ex_localscale:
		sys.bcStack.PushF(c.localscl)
	case OC_ex_majorversion:
		sys.bcStack.PushI(int32(c.gi().ver[0]))
	case OC_ex_maparray:
		sys.bcStack.PushF(c.mapArray[sys.stringPool[sys.workingState.playerNo].List[*(*int32)(unsafe.Pointer(&be[*i]))]])
		*i += 4
	case OC_ex_max:
		v2 := sys.bcStack.Pop()
		be.max(sys.bcStack.Top(), v2)
	case OC_ex_min:
		v2 := sys.bcStack.Pop()
		be.min(sys.bcStack.Top(), v2)
	case OC_ex_memberno:
		sys.bcStack.PushI(int32(c.memberNo) + 1)
	case OC_ex_movecountered:
		sys.bcStack.PushI(c.moveCountered())
	case OC_ex_pausetime:
		sys.bcStack.PushI(c.pauseTime())
	case OC_ex_physics:
		sys.bcStack.PushB(c.ss.physics == StateType(be[*i]))
		*i++
	case OC_ex_playerno:
		sys.bcStack.PushI(int32(c.playerNo) + 1)
	case OC_ex_rand:
		v2 := sys.bcStack.Pop()
		be.random(sys.bcStack.Top(), v2)
	case OC_ex_rank:
		sys.bcStack.PushF(c.rank())
	case OC_ex_ratiolevel:
		sys.bcStack.PushI(c.ratioLevel())
	case OC_ex_receiveddamage:
		sys.bcStack.PushI(c.comboDmg)
	case OC_ex_receivedhits:
		sys.bcStack.PushI(c.receivedHits)
	case OC_ex_redlife:
		sys.bcStack.PushI(c.redLife)
	case OC_ex_round:
		v2 := sys.bcStack.Pop()
		be.round(sys.bcStack.Top(), v2)
	case OC_ex_roundtype:
		sys.bcStack.PushI(c.roundType())
	case OC_ex_score:
		sys.bcStack.PushF(c.score())
	case OC_ex_scoretotal:
		sys.bcStack.PushF(c.scoreTotal())
	case OC_ex_selfstatenoexist:
		*sys.bcStack.Top() = c.selfStatenoExist(*sys.bcStack.Top())
	case OC_ex_sprpriority:
		sys.bcStack.PushI(c.sprPriority)
	case OC_ex_stagebackedge:
		sys.bcStack.PushF(c.stageBackEdge())
	case OC_ex_stagefrontedge:
		sys.bcStack.PushF(c.stageFrontEdge())
	case OC_ex_stagetime:
		sys.bcStack.PushI(sys.stage.stageTime)
	case OC_ex_standby:
		sys.bcStack.PushB(c.scf(SCF_standby))
	case OC_ex_teamleader:
		sys.bcStack.PushI(int32(c.teamLeader()))
	case OC_ex_teamsize:
		sys.bcStack.PushI(c.teamSize())
	case OC_ex_timeelapsed:
		sys.bcStack.PushI(timeElapsed())
	case OC_ex_timeremaining:
		sys.bcStack.PushI(timeRemaining())
	case OC_ex_timetotal:
		sys.bcStack.PushI(timeTotal())
	case OC_ex_pos_z:
		sys.bcStack.PushF(c.pos[2] * c.localscl / oc.localscl)
	case OC_ex_vel_z:
		sys.bcStack.PushF(c.vel[2] * c.localscl / oc.localscl)
	default:
		sys.errLog.Printf("%v\n", be[*i-1])
		c.panic()
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
type NullStateController struct{}

func (NullStateController) Run(_ *Char, _ []int32) bool { return false }

var nullStateController NullStateController

type bytecodeFunction struct {
	numVars int32
	numRets int32
	numArgs int32
	ctrls   []StateController
}

func (bf bytecodeFunction) run(c *Char, ret []uint8) (changeState bool) {
	oldv, oldvslen := sys.bcVar, len(sys.bcVarStack)
	sys.bcVar = sys.bcVarStack.Alloc(int(bf.numVars))
	if len(sys.bcStack) != int(bf.numArgs) {
		c.panic()
	}
	copy(sys.bcVar, sys.bcStack)
	sys.bcStack.Clear()
	for _, sc := range bf.ctrls {
		switch sc.(type) {
		case StateBlock:
		default:
			if c.hitPause() {
				continue
			}
		}
		if sc.Run(c, nil) {
			changeState = true
			break
		}
	}
	if !changeState {
		if len(ret) > 0 {
			if len(ret) != int(bf.numRets) {
				c.panic()
			}
			for i, r := range ret {
				oldv[r] = sys.bcVar[int(bf.numArgs)+i]
			}
		}
	}
	sys.bcVar, sys.bcVarStack = oldv, sys.bcVarStack[:oldvslen]
	return
}

type callFunction struct {
	bytecodeFunction
	arg BytecodeExp
	ret []uint8
}

func (cf callFunction) Run(c *Char, _ []int32) (changeState bool) {
	if len(cf.arg) > 0 {
		sys.bcStack.Push(cf.arg.run(c))
	}
	return cf.run(c, cf.ret)
}

type StateBlock struct {
	persistent          int32
	persistentIndex     int32
	ignorehitpause      int32
	ctrlsIgnorehitpause bool
	trigger             BytecodeExp
	elseBlock           *StateBlock
	ctrls               []StateController
}

func newStateBlock() *StateBlock {
	return &StateBlock{persistent: 1, persistentIndex: -1, ignorehitpause: -2}
}
func (b StateBlock) Run(c *Char, ps []int32) (changeState bool) {
	if c.hitPause() {
		if b.ignorehitpause < -1 {
			return false
		}
		if b.ignorehitpause >= 0 {
			ww := &c.ss.wakegawakaranai[sys.workingState.playerNo][b.ignorehitpause]
			*ww = !*ww
			if !*ww {
				return false
			}
		}
	}
	if b.persistentIndex >= 0 {
		ps[b.persistentIndex]--
		if ps[b.persistentIndex] > 0 {
			return false
		}
	}
	sys.workingChar = c
	if len(b.trigger) > 0 && !b.trigger.evalB(c) {
		if b.elseBlock != nil {
			return b.elseBlock.Run(c, ps)
		}
		return false
	}
	for _, sc := range b.ctrls {
		switch sc.(type) {
		case StateBlock:
		default:
			if !b.ctrlsIgnorehitpause && c.hitPause() {
				continue
			}
		}
		if sc.Run(c, ps) {
			return true
		}
	}
	if b.persistentIndex >= 0 {
		ps[b.persistentIndex] = b.persistent
	}
	return false
}

type StateExpr BytecodeExp

func (se StateExpr) Run(c *Char, _ []int32) (changeState bool) {
	BytecodeExp(se).run(c)
	return false
}

type varAssign struct {
	vari uint8
	be   BytecodeExp
}

func (va varAssign) Run(c *Char, _ []int32) (changeState bool) {
	sys.bcVar[va.vari] = va.be.run(c)
	return false
}

type StateControllerBase []byte

func newStateControllerBase() *StateControllerBase {
	return (*StateControllerBase)(&[]byte{})
}
func (StateControllerBase) beToExp(be ...BytecodeExp) []BytecodeExp {
	return be
}

/*func (StateControllerBase) fToExp(f ...float32) (exp []BytecodeExp) {
	for _, v := range f {
		var be BytecodeExp
		be.appendValue(BytecodeFloat(v))
		exp = append(exp, be)
	}
	return
}*/
func (StateControllerBase) iToExp(i ...int32) (exp []BytecodeExp) {
	for _, v := range i {
		var be BytecodeExp
		be.appendValue(BytecodeInt(v))
		exp = append(exp, be)
	}
	return
}

// Converts a bool to a []BytecodeExp
func (StateControllerBase) bToExp(i bool) (exp []BytecodeExp) {
	var be BytecodeExp
	be.appendValue(BytecodeBool(i))
	exp = append(exp, be)
	return
}
func (scb *StateControllerBase) add(id byte, exp []BytecodeExp) {
	*scb = append(*scb, id, byte(len(exp)))
	for _, e := range exp {
		l := int32(len(e))
		*scb = append(*scb, (*(*[4]byte)(unsafe.Pointer(&l)))[:]...)
		*scb = append(*scb, *(*[]byte)(unsafe.Pointer(&e))...)
	}
}
func (scb StateControllerBase) run(c *Char,
	f func(byte, []BytecodeExp) bool) {
	for i := 0; i < len(scb); {
		id := scb[i]
		i++
		n := scb[i]
		i++
		if cap(sys.workBe) < int(n) {
			sys.workBe = make([]BytecodeExp, n)
		} else {
			sys.workBe = sys.workBe[:n]
		}
		for m := 0; m < int(n); m++ {
			l := *(*int32)(unsafe.Pointer(&scb[i]))
			i += 4
			sys.workBe[m] = (*(*BytecodeExp)(unsafe.Pointer(&scb)))[i : i+int(l)]
			i += int(l)
		}
		if !f(id, sys.workBe) {
			break
		}
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

func (sc stateDef) Run(c *Char) {
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
			if exp[0].evalB(c) && c.rdDistX(c.p2(), c).ToF() < 0 {
				c.setFacing(-c.facing)
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
			c.changeAnim(exp[1].evalI(c), exp[0].evalB(c))
		case stateDef_ctrl:
			//in mugen fatal blow ignores statedef ctrl
			if !c.ghv.fatal {
				c.setCtrl(exp[0].evalB(c))
			} else {
				c.ghv.fatal = false
			}
		case stateDef_poweradd:
			c.powerAdd(exp[0].evalI(c))
		}
		return true
	})
}

type hitBy StateControllerBase

const (
	hitBy_value byte = iota
	hitBy_value2
	hitBy_time
	hitBy_redirectid
)

func (sc hitBy) Run(c *Char, _ []int32) bool {
	time := int32(1)
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case hitBy_time:
			time = exp[0].evalI(c)
		case hitBy_value:
			crun.hitby[0].time = time
			crun.hitby[0].flag = exp[0].evalI(c)
		case hitBy_value2:
			crun.hitby[1].time = time
			crun.hitby[1].flag = exp[0].evalI(c)
		case hitBy_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type notHitBy hitBy

func (sc notHitBy) Run(c *Char, _ []int32) bool {
	time := int32(1)
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case hitBy_time:
			time = exp[0].evalI(c)
		case hitBy_value:
			crun.hitby[0].time = time
			crun.hitby[0].flag = ^exp[0].evalI(c)
		case hitBy_value2:
			crun.hitby[1].time = time
			crun.hitby[1].flag = ^exp[0].evalI(c)

		case hitBy_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type assertSpecial StateControllerBase

const (
	assertSpecial_flag byte = iota
	assertSpecial_flag_g
	assertSpecial_redirectid
)

func (sc assertSpecial) Run(c *Char, _ []int32) bool {
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case assertSpecial_flag:
			crun.setSF(CharSpecialFlag(exp[0].evalI(c)))
		case assertSpecial_flag_g:
			sys.setSF(GlobalSpecialFlag(exp[0].evalI(c)))
		case assertSpecial_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
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
	playSnd_volumescale
	playSnd_freqmul
	playSnd_loop
	playSnd_redirectid
)

func (sc playSnd) Run(c *Char, _ []int32) bool {
	crun := c
	f, lw, lp := false, false, false
	var g, n, ch, vo int32 = -1, 0, -1, 100
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
			vo = vo + int32(float64(exp[0].evalI(c))*(25.0/64.0))
		case playSnd_volumescale:
			vo = exp[0].evalI(c)
		case playSnd_freqmul:
			fr = exp[0].evalF(c)
		case playSnd_loop:
			lp = exp[0].evalB(c)
		case playSnd_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	crun.playSound(f, lw, lp, g, n, ch, vo, p, fr, x, true)
	return false
}

type changeState StateControllerBase

const (
	changeState_value byte = iota
	changeState_ctrl
	changeState_anim
	changeState_readplayerid
	changeState_redirectid
)

func (sc changeState) Run(c *Char, _ []int32) bool {
	crun := c
	var v, a, ctrl int32 = -1, -1, -1
	fflg := false
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case changeState_value:
			v = exp[0].evalI(c)
		case changeState_ctrl:
			ctrl = exp[0].evalI(c)
		case changeState_anim:
			a = exp[1].evalI(c)
			fflg = exp[0].evalB(c)
		case changeState_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	crun.changeState(v, a, ctrl, fflg)
	return true
}

type selfState changeState

func (sc selfState) Run(c *Char, _ []int32) bool {
	crun := c
	var v, a, r, ctrl int32 = -1, -1, -1, -1
	fflg := false
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case changeState_value:
			v = exp[0].evalI(c)
		case changeState_ctrl:
			ctrl = exp[0].evalI(c)
		case changeState_anim:
			a = exp[1].evalI(c)
			fflg = exp[0].evalB(c)
		case changeState_readplayerid:
			if rpid := sys.playerID(exp[0].evalI(c)); rpid != nil {
				r = int32(rpid.playerNo)
			} else {
				return false
			}
		case changeState_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	crun.selfState(v, a, r, ctrl, fflg)
	return true
}

type tagIn StateControllerBase

const (
	tagIn_stateno = iota
	tagIn_partnerstateno
	tagIn_self
	tagIn_partner
	tagIn_ctrl
	tagIn_partnerctrl
	tagIn_leader
	tagIn_redirectid
)

func (sc tagIn) Run(c *Char, _ []int32) bool {
	crun := c
	var tagSCF int = -1
	var partnerNo int32 = -1
	var partnerStateNo int32 = -1
	var partnerCtrlSetting int = -1
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case tagIn_stateno:
			sn := exp[0].evalI(c)
			if sn >= 0 {
				crun.changeState(sn, -1, -1, false)
				if tagSCF == -1 {
					tagSCF = 1
				}
			} else {
				return false
			}
		case tagIn_partnerstateno:
			if psn := exp[0].evalI(c); psn >= 0 {
				partnerStateNo = psn
			} else {
				return false
			}
		case tagIn_self:
			sti := exp[0].evalB(c)
			if sti {
				tagSCF = 1
			} else {
				tagSCF = 0
			}
		case tagIn_partner:
			pti := exp[0].evalI(c)
			if pti >= 0 {
				partnerNo = pti
			} else {
				return false
			}
		case tagIn_ctrl:
			ctrls := exp[0].evalB(c)
			crun.setCtrl(ctrls)
			if tagSCF == -1 {
				tagSCF = 1
			}
		case tagIn_partnerctrl:
			pctrls := exp[0].evalB(c)
			if pctrls {
				partnerCtrlSetting = 1
			} else {
				partnerCtrlSetting = 0
			}
		case tagIn_leader:
			if crun.teamside != -1 {
				ld := int(exp[0].evalI(c)) - 1
				if ld&1 == crun.playerNo&1 && ld >= crun.teamside && ld <= int(sys.numSimul[crun.teamside])*2-^crun.teamside&1-1 {
					sys.teamLeader[crun.playerNo&1] = ld
				}
			}
		case tagIn_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	// Data adjustments
	if tagSCF == -1 && partnerNo == -1 {
		tagSCF = 1
	}
	if tagSCF == 1 {
		crun.unsetSCF(SCF_standby)
	}
	// Partner
	if partnerNo != -1 && crun.partnerV2(partnerNo) != nil {
		partner := crun.partnerV2(partnerNo)
		partner.unsetSCF(SCF_standby)
		if partnerStateNo >= 0 {
			partner.changeState(partnerStateNo, -1, -1, false)
		}
		if partnerCtrlSetting != -1 {
			if partnerCtrlSetting == 1 {
				partner.setCtrl(true)
			} else {
				partner.setCtrl(false)
			}
		}
	}
	return false
}

type tagOut StateControllerBase

const (
	tagOut_self = iota
	tagOut_partner
	tagOut_stateno
	tagOut_partnerstateno
	tagOut_redirectid
)

func (sc tagOut) Run(c *Char, _ []int32) bool {
	crun := c
	var tagSCF int = -1
	var partnerNo int32 = -1
	var partnerStateNo int32 = -1
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case tagOut_self:
			if exp[0].evalB(c) {
				tagSCF = 1
			} else {
				tagSCF = 0
			}
		case tagOut_stateno:
			sn := exp[0].evalI(c)
			if sn >= 0 {
				crun.changeState(sn, -1, -1, false)
				if tagSCF == -1 {
					tagSCF = 1
				}
			} else {
				return false
			}
		case tagOut_partner:
			pti := exp[0].evalI(c)
			if pti >= 0 {
				partnerNo = pti
			} else {
				return false
			}
		case tagOut_partnerstateno:
			if psn := exp[0].evalI(c); psn >= 0 {
				partnerStateNo = psn
			} else {
				return false
			}
		case tagOut_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	if tagSCF == -1 && partnerNo == -1 && partnerStateNo == -1 {
		tagSCF = 1
	}
	if tagSCF == 1 {
		crun.setSCF(SCF_standby)
	}
	if partnerNo != -1 && crun.partnerV2(partnerNo) != nil {
		partner := crun.partnerV2(partnerNo)
		partner.setSCF(SCF_standby)
		if partnerStateNo >= 0 {
			partner.changeState(partnerStateNo, -1, -1, false)
		}
	}
	return false
}

type destroySelf StateControllerBase

const (
	destroySelf_recursive = iota
	destroySelf_removeexplods
	destroySelf_redirectid
)

func (sc destroySelf) Run(c *Char, _ []int32) bool {
	crun := c
	rec, rem := false, false
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case destroySelf_recursive:
			rec = exp[0].evalB(c)
		case destroySelf_removeexplods:
			rem = exp[0].evalB(c)
		case destroySelf_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return crun.destroySelf(rec, rem)
}

type changeAnim StateControllerBase

const (
	changeAnim_elem byte = iota
	changeAnim_value
	changeAnim_redirectid
)

func (sc changeAnim) Run(c *Char, _ []int32) bool {
	crun := c
	var elem int32
	setelem := false
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case changeAnim_elem:
			elem = exp[0].evalI(c)
			setelem = true
		case changeAnim_value:
			crun.changeAnim(exp[1].evalI(c), exp[0].evalB(c))
			if setelem {
				crun.setAnimElem(elem)
			}
		case changeAnim_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type changeAnim2 changeAnim

func (sc changeAnim2) Run(c *Char, _ []int32) bool {
	crun := c
	var elem int32
	setelem := false
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case changeAnim_elem:
			elem = exp[0].evalI(c)
			setelem = true
		case changeAnim_value:
			crun.changeAnim2(exp[1].evalI(c), exp[0].evalB(c))
			if setelem {
				crun.setAnimElem(elem)
			}
		case changeAnim_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
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
	helper_redirectid
	helper_remappal
	helper_extendsmap
	helper_inheritjuggle
	helper_immortal
	helper_kovelocity
	helper_preserve
)

func (sc helper) Run(c *Char, _ []int32) bool {
	crun := c
	var lclscround float32 = 1.0
	var h *Char
	pt := PT_P1
	var f, st int32 = 1, 0
	var extmap bool
	var x, y float32 = 0, 0
	rp := [...]int32{-1, 0}
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		if h == nil {
			if id == helper_redirectid {
				if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
					crun = rid
					lclscround = c.localscl / crun.localscl
					h = crun.newHelper()
				} else {
					return false
				}
			} else {
				h = c.newHelper()
			}
		}
		if h == nil {
			return false
		}
		switch id {
		case helper_helpertype:
			h.player = exp[0].evalB(c)
		case helper_name:
			h.name = string(*(*[]byte)(unsafe.Pointer(&exp[0])))
		case helper_postype:
			pt = PosType(exp[0].evalI(c))
		case helper_ownpal:
			h.ownpal = exp[0].evalB(c)
		case helper_size_xscale:
			h.size.xscale = exp[0].evalF(c)
		case helper_size_yscale:
			h.size.yscale = exp[0].evalF(c)
		case helper_size_ground_back:
			h.size.ground.back = exp[0].evalF(c)
		case helper_size_ground_front:
			h.size.ground.front = exp[0].evalF(c)
		case helper_size_air_back:
			h.size.air.back = exp[0].evalF(c)
		case helper_size_air_front:
			h.size.air.front = exp[0].evalF(c)
		case helper_size_height:
			h.size.height = exp[0].evalF(c)
		case helper_size_proj_doscale:
			h.size.proj.doscale = exp[0].evalI(c)
		case helper_size_head_pos:
			h.size.head.pos[0] = exp[0].evalF(c)
			if len(exp) > 1 {
				h.size.head.pos[1] = exp[1].evalF(c)
			}
		case helper_size_mid_pos:
			h.size.mid.pos[0] = exp[0].evalF(c)
			if len(exp) > 1 {
				h.size.mid.pos[1] = exp[1].evalF(c)
			}
		case helper_size_shadowoffset:
			h.size.shadowoffset = exp[0].evalF(c)
		case helper_stateno:
			st = exp[0].evalI(c)
		case helper_keyctrl:
			for _, e := range exp {
				m := e.run(c).ToI()
				if m > 0 && m <= int32(len(h.keyctrl)) {
					h.keyctrl[m-1] = true
				}
			}
		case helper_id:
			h.helperId = exp[0].evalI(c)
		case helper_pos:
			x = exp[0].evalF(c) * lclscround
			if len(exp) > 1 {
				y = exp[1].evalF(c) * lclscround
			}
		case helper_facing:
			f = exp[0].evalI(c)
		case helper_pausemovetime:
			h.pauseMovetime = exp[0].evalI(c)
		case helper_supermovetime:
			h.superMovetime = exp[0].evalI(c)
		case helper_remappal:
			rp[0] = exp[0].evalI(c)
			if len(exp) > 1 {
				rp[1] = exp[1].evalI(c)
			}
		case helper_extendsmap:
			extmap = exp[0].evalB(c)
		case helper_inheritjuggle:
			h.inheritJuggle = exp[0].evalI(c)
		case helper_immortal:
			h.immortal = exp[0].evalB(c)
		case helper_kovelocity:
			h.kovelocity = exp[0].evalB(c)
		case helper_preserve:
			h.preserve = exp[0].evalB(c)
		}
		return true
	})
	if h == nil {
		return false
	}
	if crun.minus == -2 || crun.minus == -4 {
		h.localscl = (320 / float32(crun.localcoord))
		h.localcoord = crun.localcoord
	} else {
		h.localscl = crun.localscl
		h.localcoord = crun.localcoord
	}
	crun.helperInit(h, st, pt, x, y, f, rp, extmap)
	return false
}

type ctrlSet StateControllerBase

const (
	ctrlSet_value byte = iota
	ctrlSet_redirectid
)

func (sc ctrlSet) Run(c *Char, _ []int32) bool {
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case ctrlSet_value:
			crun.setCtrl(exp[0].evalB(c))
		case ctrlSet_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
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
	explod_strictontop
	explod_under
	explod_shadow
	explod_removeongethit
	explod_trans
	explod_anim
	explod_angle
	explod_yangle
	explod_xangle
	explod_ignorehitpause
	explod_bindid
	explod_space
	explod_redirectid
)

func (sc explod) Run(c *Char, _ []int32) bool {
	crun := c
	var lclscround float32 = 1.0
	var e *Explod
	var i int
	//e, i := crun.newExplod()
	rp := [...]int32{-1, 0}
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		if e == nil {
			if id == explod_redirectid {
				if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
					crun = rid
					lclscround = c.localscl / crun.localscl
					e, i = crun.newExplod()
					if e == nil {
						return false
					}
					e.id = 0
					if crun.stCgi().ver[0] == 1 && crun.stCgi().ver[1] == 1 {
						e.postype = PT_N
					}
				} else {
					return false
				}
			} else {
				e, i = crun.newExplod()
				if e == nil {
					return false
				}
				e.id = 0
				if crun.stCgi().ver[0] == 1 && crun.stCgi().ver[1] == 1 {
					e.postype = PT_N
				}
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
			e.offset[0] = exp[0].evalF(c) * lclscround
			if len(exp) > 1 {
				e.offset[1] = exp[1].evalF(c) * lclscround
			}
		case explod_random:
			rndx := exp[0].evalF(c) * lclscround
			e.offset[0] += RandF(-rndx, rndx)
			if len(exp) > 1 {
				rndy := exp[1].evalF(c) * lclscround
				e.offset[1] += RandF(-rndy, rndy)
			}
		case explod_postype:
			e.postype = PosType(exp[0].evalI(c))
		case explod_space:
			e.space = Space(exp[0].evalI(c))
		case explod_velocity:
			e.velocity[0] = exp[0].evalF(c) * lclscround
			if len(exp) > 1 {
				e.velocity[1] = exp[1].evalF(c) * lclscround
			}
		case explod_accel:
			e.accel[0] = exp[0].evalF(c) * lclscround
			if len(exp) > 1 {
				e.accel[1] = exp[1].evalF(c) * lclscround
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
		case explod_strictontop:
			if e.ontop {
				e.sprpriority = 0
			}
		case explod_under:
			if !e.ontop {
				e.under = exp[0].evalB(c)
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
			e.alpha[0] = exp[0].evalI(c)
			e.alpha[1] = exp[1].evalI(c)
			if len(exp) >= 3 {
				e.alpha[0] = Max(0, Min(255, e.alpha[0]))
				e.alpha[1] = Max(0, Min(255, e.alpha[1]))
				if len(exp) >= 4 {
					e.alpha[1] = ^e.alpha[1]
				} else if e.alpha[0] == 1 && e.alpha[1] == 255 {
					e.alpha[0] = 0
				}
			}
		case explod_anim:
			e.anim = crun.getAnim(exp[1].evalI(c), exp[0].evalB(c), false)
		case explod_angle:
			e.angle = exp[0].evalF(c)
		case explod_yangle:
			e.yangle = exp[0].evalF(c)
		case explod_xangle:
			e.xangle = exp[0].evalF(c)
		case explod_ignorehitpause:
			e.ignorehitpause = exp[0].evalB(c)
		case explod_bindid:
			bId := exp[0].evalI(c)
			if bId == -1 {
				bId = crun.id
			}
			e.bindId = bId
		}
		return true
	})
	if e == nil {
		return false
	}
	if c.minus == -2 || c.minus == -4 { //TODO: isn't this supposed to check crun instead of c?
		e.localscl = (320 / float32(crun.localcoord))
	} else {
		e.localscl = crun.localscl
	}
	e.setPos(crun)
	crun.insertExplodEx(i, rp)
	return false
}

type modifyExplod explod

func (sc modifyExplod) Run(c *Char, _ []int32) bool {
	crun := c
	var lclscround float32 = 1.0
	eid := int32(-1)
	var expls []*Explod
	rp := [...]int32{-1, 0}
	eachExpl := func(f func(e *Explod)) {
		for _, e := range expls {
			f(e)
		}
	}
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case explod_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
				lclscround = c.localscl / crun.localscl
			} else {
				return false
			}
		case explod_remappal:
			rp[0] = exp[0].evalI(c)
			if len(exp) > 1 {
				rp[1] = exp[1].evalI(c)
			}
		case explod_id:
			eid = exp[0].evalI(c)
		default:
			if len(expls) == 0 {
				expls = crun.getExplods(eid)
				if len(expls) == 0 {
					return false
				}
				eachExpl(func(e *Explod) {
					if e.ownpal {
						crun.remapPal(e.palfx, [...]int32{1, 1}, rp)
					}
				})
			}
			switch id {
			case explod_ownpal:
				op := exp[0].evalB(c)
				eachExpl(func(e *Explod) { e.ownpal = op })
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
				x := exp[0].evalF(c) * lclscround
				eachExpl(func(e *Explod) { e.offset[0] = x })
				if len(exp) > 1 {
					y := exp[1].evalF(c) * lclscround
					eachExpl(func(e *Explod) { e.offset[1] = y })
				}
			case explod_random:
				rndx := exp[0].evalF(c) * lclscround
				rndx = RandF(-rndx, rndx)
				eachExpl(func(e *Explod) { e.offset[0] += rndx })
				if len(exp) > 1 {
					rndy := exp[1].evalF(c) * lclscround
					rndy = RandF(-rndy, rndy)
					eachExpl(func(e *Explod) { e.offset[1] += rndy })
				}
			case explod_postype:
				pt := PosType(exp[0].evalI(c))
				eachExpl(func(e *Explod) {
					e.postype = pt
					e.setPos(c)
				})
			case explod_space:
				sp := Space(exp[0].evalI(c))
				eachExpl(func(e *Explod) { e.space = sp })
			case explod_velocity:
				x := exp[0].evalF(c) * lclscround
				eachExpl(func(e *Explod) { e.velocity[0] = x })
				if len(exp) > 1 {
					y := exp[1].evalF(c) * lclscround
					eachExpl(func(e *Explod) { e.velocity[1] = y })
				}
			case explod_accel:
				x := exp[0].evalF(c) * lclscround
				eachExpl(func(e *Explod) { e.accel[0] = x })
				if len(exp) > 1 {
					y := exp[1].evalF(c) * lclscround
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
					if e.ontop && e.under {
						e.under = false
					}
				})
			case explod_strictontop:
				eachExpl(func(e *Explod) {
					if e.ontop {
						e.sprpriority = 0
					}
				})
			case explod_under:
				t := exp[0].evalB(c)
				eachExpl(func(e *Explod) {
					e.under = t
					if e.under && e.ontop {
						e.ontop = false
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
				if len(exp) >= 3 {
					s, d = Max(0, Min(255, s)), Max(0, Min(255, d))
					if len(exp) >= 4 {
						d = ^d
					} else if s == 1 && d == 255 {
						s = 0
					}
				}
				eachExpl(func(e *Explod) { e.alpha = [...]int32{s, d} })
			case explod_anim:
				anim := crun.getAnim(exp[1].evalI(c), exp[0].evalB(c), false)
				eachExpl(func(e *Explod) { e.anim = anim })
			case explod_angle:
				a := exp[0].evalF(c)
				eachExpl(func(e *Explod) { e.angle = a })
			case explod_yangle:
				ya := exp[0].evalF(c)
				eachExpl(func(e *Explod) { e.yangle = ya })
			case explod_xangle:
				xa := exp[0].evalF(c)
				eachExpl(func(e *Explod) { e.xangle = xa })
			case explod_ignorehitpause:
				ihp := exp[0].evalB(c)
				eachExpl(func(e *Explod) { e.ignorehitpause = ihp })
			case explod_bindid:
				bId := exp[0].evalI(c)
				if bId == -1 {
					bId = crun.id
				}
				eachExpl(func(e *Explod) { e.bindId = bId })
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
	gameMakeAnim_redirectid
)

func (sc gameMakeAnim) Run(c *Char, _ []int32) bool {
	crun := c
	var e *Explod
	var i int
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		if e == nil {
			if id == gameMakeAnim_redirectid {
				if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
					crun = rid
					e, i = crun.newExplod()
					if e == nil {
						return false
					}
					e.id = 0
				} else {
					return false
				}
			} else {
				e, i = crun.newExplod()
				if e == nil {
					return false
				}
				e.id = 0
			}
		}
		switch id {
		case gameMakeAnim_pos:
			e.offset[0] = exp[0].evalF(c) * c.localscl / crun.localscl
			if len(exp) > 1 {
				e.offset[1] = exp[1].evalF(c) * c.localscl / crun.localscl
			}
		case gameMakeAnim_random:
			rndx := exp[0].evalF(c)
			e.offset[0] += RandF(-rndx, rndx) * c.localscl / crun.localscl
			if len(exp) > 1 {
				rndy := exp[1].evalF(c)
				e.offset[1] += RandF(-rndy, rndy) * c.localscl / crun.localscl
			}
		case gameMakeAnim_under:
			e.ontop = !exp[0].evalB(c)
		case gameMakeAnim_anim:
			e.anim = crun.getAnim(exp[1].evalI(c), exp[0].evalB(c), false)
		}
		return true
	})
	if e == nil {
		return false
	}
	e.ontop, e.sprpriority, e.ownpal = true, math.MinInt32, true
	e.offset[0] -= float32(crun.size.draw.offset[0])
	e.offset[1] -= float32(crun.size.draw.offset[1])
	e.setPos(crun)
	crun.insertExplod(i)
	return false
}

type posSet StateControllerBase

const (
	posSet_x byte = iota
	posSet_y
	posSet_z
	posSet_redirectid
)

func (sc posSet) Run(c *Char, _ []int32) bool {
	crun := c
	var lclscround float32 = 1.0
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case posSet_x:
			crun.setX(sys.cam.Pos[0]/crun.localscl + exp[0].evalF(c)*lclscround)
		case posSet_y:
			crun.setY(exp[0].evalF(c)*lclscround + crun.platformPosY)
		case posSet_z:
			if crun.size.z.enable {
				crun.setZ(exp[0].evalF(c) * lclscround)
			} else {
				exp[0].run(c)
			}
		case posSet_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
				lclscround = c.localscl / crun.localscl
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type posAdd posSet

func (sc posAdd) Run(c *Char, _ []int32) bool {
	crun := c
	var lclscround float32 = 1.0
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case posSet_x:
			crun.addX(exp[0].evalF(c) * lclscround)
		case posSet_y:
			crun.addY(exp[0].evalF(c) * lclscround)
		case posSet_z:
			if crun.size.z.enable {
				crun.addZ(exp[0].evalF(c) * lclscround)
			} else {
				exp[0].run(c)
			}
		case posSet_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
				lclscround = c.localscl / crun.localscl
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type velSet posSet

func (sc velSet) Run(c *Char, _ []int32) bool {
	crun := c
	var lclscround float32 = 1.0
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case posSet_x:
			crun.setXV(exp[0].evalF(c) * lclscround)
		case posSet_y:
			crun.setYV(exp[0].evalF(c) * lclscround)
		case posSet_z:
			if crun.size.z.enable {
				crun.setZV(exp[0].evalF(c) * lclscround)
			} else {
				exp[0].run(c)
			}
		case posSet_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
				lclscround = c.localscl / crun.localscl
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type velAdd posSet

func (sc velAdd) Run(c *Char, _ []int32) bool {
	crun := c
	var lclscround float32 = 1.0
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case posSet_x:
			crun.addXV(exp[0].evalF(c) * lclscround)
		case posSet_y:
			crun.addYV(exp[0].evalF(c) * lclscround)
		case posSet_z:
			if crun.size.z.enable {
				crun.addZV(exp[0].evalF(c) * lclscround)
			} else {
				exp[0].run(c)
			}
		case posSet_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
				lclscround = c.localscl / crun.localscl
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type velMul posSet

func (sc velMul) Run(c *Char, _ []int32) bool {
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case posSet_x:
			crun.mulXV(exp[0].evalF(c))
		case posSet_y:
			crun.mulYV(exp[0].evalF(c))
		case posSet_z:
			if crun.size.z.enable {
				crun.mulZV(exp[0].evalF(c))
			} else {
				exp[0].run(c)
			}
		case posSet_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
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
	palFX_last = iota - 1
	palFX_redirectid
)

func (sc palFX) runSub(c *Char, pfd *PalFXDef,
	id byte, exp []BytecodeExp) bool {
	switch id {
	case palFX_time:
		pfd.time = exp[0].evalI(c)
	case palFX_color:
		pfd.color = MaxF(0, MinF(1, exp[0].evalF(c)/256))
	case palFX_add:
		pfd.add[0] = exp[0].evalI(c)
		pfd.add[1] = exp[1].evalI(c)
		pfd.add[2] = exp[2].evalI(c)
	case palFX_mul:
		pfd.mul[0] = exp[0].evalI(c)
		pfd.mul[1] = exp[1].evalI(c)
		pfd.mul[2] = exp[2].evalI(c)
	case palFX_sinadd:
		pfd.sinadd[0] = exp[0].evalI(c)
		pfd.sinadd[1] = exp[1].evalI(c)
		pfd.sinadd[2] = exp[2].evalI(c)
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
	crun := c
	if !crun.ownpal {
		return false
	}
	pf := crun.palfx
	if pf == nil {
		pf = newPalFX()
	}
	pf.clear2(true)
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		if id == palFX_redirectid {
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
				pf = crun.palfx
				if pf == nil {
					pf = newPalFX()
				}
				pf.clear2(true)
			} else {
				return false
			}
		}
		sc.runSub(c, &pf.PalFXDef, id, exp)
		return true
	})
	return false
}

type allPalFX palFX

func (sc allPalFX) Run(c *Char, _ []int32) bool {
	sys.allPalFX.clear()
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		palFX(sc).runSub(c, &sys.allPalFX.PalFXDef, id, exp)
		return true
	})
	return false
}

type bgPalFX palFX

func (sc bgPalFX) Run(c *Char, _ []int32) bool {
	sys.bgPalFX.clear()
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		palFX(sc).runSub(c, &sys.bgPalFX.PalFXDef, id, exp)
		return true
	})
	return false
}

type afterImage palFX

const (
	afterImage_trans = iota + palFX_last + 1
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
	afterImage_ignorehitpause
	afterImage_last = iota + palFX_last + 1 - 1
	afterImage_redirectid
)

func (sc afterImage) runSub(c *Char, ai *AfterImage,
	id byte, exp []BytecodeExp) {
	switch id {
	case afterImage_trans:
		ai.alpha[0] = exp[0].evalI(c)
		ai.alpha[1] = exp[1].evalI(c)
		if len(exp) >= 3 {
			ai.alpha[0] = Max(0, Min(255, ai.alpha[0]))
			ai.alpha[1] = Max(0, Min(255, ai.alpha[1]))
			if len(exp) >= 4 {
				ai.alpha[1] = ^ai.alpha[1]
			} else if ai.alpha[0] == 1 && ai.alpha[1] == 255 {
				ai.alpha[0] = 0
			}
		}
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
	case afterImage_ignorehitpause:
		ai.ignorehitpause = exp[0].evalB(c)
	}
}
func (sc afterImage) Run(c *Char, _ []int32) bool {
	crun := c
	crun.aimg.clear()
	crun.aimg.time = 1
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		if id == afterImage_redirectid {
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
				crun.aimg.clear()
				crun.aimg.time = 1
			} else {
				return false
			}
		}
		sc.runSub(c, &crun.aimg, id, exp)
		return true
	})
	crun.aimg.setupPalFX()
	return false
}

type afterImageTime StateControllerBase

const (
	afterImageTime_time byte = iota
	afterImageTime_redirectid
)

func (sc afterImageTime) Run(c *Char, _ []int32) bool {
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		if id == afterImageTime_redirectid {
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		if crun.aimg.timegap <= 0 {
			return false
		}
		switch id {
		case afterImageTime_time:
			crun.aimg.time = exp[0].evalI(c)
			crun.aimg.timecount = 0
		}
		return true
	})
	return false
}

type hitDef afterImage

const (
	hitDef_attr = iota + afterImage_last + 1
	hitDef_guardflag
	hitDef_hitflag
	hitDef_ground_type
	hitDef_air_type
	hitDef_animtype
	hitDef_air_animtype
	hitDef_fall_animtype
	hitDef_affectteam
	hitDef_teamside
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
	hitDef_dizzypoints
	hitDef_guardpoints
	hitDef_redlife
	hitDef_score
	hitDef_last = iota + afterImage_last + 1 - 1
	hitDef_redirectid
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
	case hitDef_teamside:
		n := exp[0].evalI(c)
		if n > 2 {
			hd.teamside = 2
		} else if n < 0 {
			hd.teamside = 0
		} else {
			hd.teamside = int(n)
		}
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
			if len(exp) > 2 {
				exp[2].run(c)
			}
		}
	case hitDef_maxdist:
		hd.maxdist[0] = exp[0].evalF(c)
		if len(exp) > 1 {
			hd.maxdist[1] = exp[1].evalF(c)
			if len(exp) > 2 {
				exp[2].run(c)
			}
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
	case hitDef_dizzypoints:
		hd.dizzypoints = Max(IErr+1, exp[0].evalI(c))
	case hitDef_guardpoints:
		hd.guardpoints = Max(IErr+1, exp[0].evalI(c))
	case hitDef_redlife:
		hd.redlife = Max(IErr+1, exp[0].evalI(c))
	case hitDef_score:
		hd.score[0] = exp[0].evalF(c)
		if len(exp) > 1 {
			hd.score[1] = exp[1].evalF(c)
		}
	default:
		if !palFX(sc).runSub(c, &hd.palfx, id, exp) {
			return false
		}
	}
	return true
}
func (sc hitDef) Run(c *Char, _ []int32) bool {
	crun := c
	crun.hitdef.clear()
	crun.hitdef.playerNo = sys.workingState.playerNo
	crun.hitdef.sparkno = ^c.gi().data.sparkno
	crun.hitdef.guard_sparkno = ^c.gi().data.guard.sparkno
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		if id == hitDef_redirectid {
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
				crun.hitdef.clear()
				crun.hitdef.playerNo = sys.workingState.playerNo
				crun.hitdef.sparkno = ^c.gi().data.sparkno
				crun.hitdef.guard_sparkno = ^c.gi().data.guard.sparkno
			} else {
				return false
			}
		}
		sc.runSub(c, &crun.hitdef, id, exp)
		return true
	})
	//winmugenでHitdefのattrが投げ属性で自分側pausetimeが1以上の時、毎フレーム実行されなくなる
	if crun.hitdef.attr&int32(AT_AT) != 0 && crun.moveContact() == 1 &&
		c.gi().ver[0] != 1 && crun.hitdef.pausetime > 0 {
		crun.hitdef.attr = 0
		return false
	}
	crun.setHitdefDefault(&crun.hitdef, false)
	return false
}

type reversalDef hitDef

const (
	reversalDef_reversal_attr = iota + hitDef_last + 1
	reversalDef_redirectid
)

func (sc reversalDef) Run(c *Char, _ []int32) bool {
	crun := c
	crun.hitdef.clear()
	crun.hitdef.playerNo = sys.workingState.playerNo
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case reversalDef_reversal_attr:
			crun.hitdef.reversal_attr = exp[0].evalI(c)
		case reversalDef_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
				crun.hitdef.clear()
				crun.hitdef.playerNo = sys.workingState.playerNo
			} else {
				return false
			}
		default:
			hitDef(sc).runSub(c, &crun.hitdef, id, exp)
		}
		return true
	})
	crun.setHitdefDefault(&crun.hitdef, false)
	return false
}

type projectile hitDef

const (
	projectile_postype = iota + hitDef_last + 1
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
	projectile_projangle
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
	projectile_platform
	projectile_platformwidth
	projectile_platformheight
	projectile_platformfence
	projectile_platformangle
	projectile_redirectid
)

func (sc projectile) Run(c *Char, _ []int32) bool {
	crun := c
	var lclscround float32 = 1.0
	var p *Projectile
	pt := PT_P1
	var x, y float32 = 0, 0
	op := false
	rp := [...]int32{-1, 0}
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		if p == nil {
			if id == projectile_redirectid {
				if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
					crun = rid
					lclscround = c.localscl / crun.localscl
					p = crun.newProj()
					if p == nil {
						return false
					}
					p.hitdef.playerNo = sys.workingState.playerNo

				} else {
					return false
				}
			} else {
				p = crun.newProj()
				if p == nil {
					return false
				}
				p.hitdef.playerNo = sys.workingState.playerNo
			}
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
			p.priorityPoints = p.priority
		case projectile_projhitanim:
			p.hitanim = exp[1].evalI(c)
			p.hitanim_fflg = exp[0].evalB(c)
		case projectile_projremanim:
			p.remanim = Max(-1, exp[1].evalI(c))
			p.remanim_fflg = exp[0].evalB(c)
		case projectile_projcancelanim:
			p.cancelanim = Max(-1, exp[1].evalI(c))
			p.cancelanim_fflg = exp[0].evalB(c)
		case projectile_velocity:
			p.velocity[0] = exp[0].evalF(c) * lclscround
			if len(exp) > 1 {
				p.velocity[1] = exp[1].evalF(c) * lclscround
			}
		case projectile_velmul:
			p.velmul[0] = exp[0].evalF(c)
			if len(exp) > 1 {
				p.velmul[1] = exp[1].evalF(c)
			}
		case projectile_remvelocity:
			p.remvelocity[0] = exp[0].evalF(c) * lclscround
			if len(exp) > 1 {
				p.remvelocity[1] = exp[1].evalF(c) * lclscround
			}
		case projectile_accel:
			p.accel[0] = exp[0].evalF(c) * lclscround
			if len(exp) > 1 {
				p.accel[1] = exp[1].evalF(c) * lclscround
			}
		case projectile_projscale:
			p.scale[0] = exp[0].evalF(c)
			if len(exp) > 1 {
				p.scale[1] = exp[1].evalF(c)
			}
		case projectile_projangle:
			p.angle = exp[0].evalF(c)
		case projectile_offset:
			x = exp[0].evalF(c) * lclscround
			if len(exp) > 1 {
				y = exp[1].evalF(c) * lclscround
			}
		case projectile_projsprpriority:
			p.sprpriority = exp[0].evalI(c)
		case projectile_projstagebound:
			p.stagebound = int32(float32(exp[0].evalI(c)) * lclscround)
		case projectile_projedgebound:
			p.edgebound = int32(float32(exp[0].evalI(c)) * lclscround)
		case projectile_projheightbound:
			p.heightbound[0] = int32(float32(exp[0].evalI(c)) * lclscround)
			if len(exp) > 1 {
				p.heightbound[1] = int32(float32(exp[1].evalI(c)) * lclscround)
			}
		case projectile_projanim:
			p.anim = exp[1].evalI(c)
			p.anim_fflg = exp[0].evalB(c)
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
		case projectile_platform:
			p.platform = exp[0].evalB(c)
		case projectile_platformwidth:
			p.platformWidth[0] = exp[0].evalF(c) * lclscround
			if len(exp) > 1 {
				p.platformWidth[1] = exp[1].evalF(c) * lclscround
			}
		case projectile_platformheight:
			p.platformHeight[0] = exp[0].evalF(c) * lclscround
			if len(exp) > 1 {
				p.platformHeight[1] = exp[1].evalF(c) * lclscround
			}
		case projectile_platformangle:
			p.platformAngle = exp[0].evalF(c)
		case projectile_platformfence:
			p.platformFence = exp[0].evalB(c)
		default:
			if !hitDef(sc).runSub(c, &p.hitdef, id, exp) {
				afterImage(sc).runSub(c, &p.aimg, id, exp)
			}
		}
		return true
	})
	if p == nil {
		return false
	}
	crun.setHitdefDefault(&p.hitdef, true)
	if p.hitanim == -1 {
		p.hitanim_fflg = p.anim_fflg
	}
	if p.remanim == IErr {
		p.remanim = p.hitanim
		p.remanim_fflg = p.hitanim_fflg
	}
	if p.cancelanim == IErr {
		p.cancelanim = p.remanim
		p.cancelanim_fflg = p.remanim_fflg
	}
	if p.aimg.time != 0 {
		p.aimg.setupPalFX()
	}
	if crun.minus == -2 || crun.minus == -4 {
		p.localscl = (320 / float32(crun.localcoord))
	} else {
		p.localscl = crun.localscl
	}
	crun.projInit(p, pt, x, y, op, rp[0], rp[1])
	return false
}

type width StateControllerBase

const (
	width_edge byte = iota
	width_player
	width_value
	width_redirectid
)

func (sc width) Run(c *Char, _ []int32) bool {
	crun := c
	var lclscround float32 = 1.0
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case width_edge:
			crun.setFEdge(exp[0].evalF(c) * lclscround)
			if len(exp) > 1 {
				crun.setBEdge(exp[1].evalF(c) * lclscround)
			}
		case width_player:
			crun.setFWidth(exp[0].evalF(c) * lclscround)
			if len(exp) > 1 {
				crun.setBWidth(exp[1].evalF(c) * lclscround)
			}
		case width_value:
			v1 := exp[0].evalF(c) * lclscround
			crun.setFEdge(v1)
			crun.setFWidth(v1)
			if len(exp) > 1 {
				v2 := exp[1].evalF(c) * lclscround
				crun.setBEdge(v2)
				crun.setBWidth(v2)
			}
		case width_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
				lclscround = (320 / float32(c.localcoord)) / (320 / float32(crun.localcoord))
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type sprPriority StateControllerBase

const (
	sprPriority_value byte = iota
	sprPriority_redirectid
)

func (sc sprPriority) Run(c *Char, _ []int32) bool {
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case sprPriority_value:
			crun.setSprPriority(exp[0].evalI(c))
		case sprPriority_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type varSet StateControllerBase

const (
	varSet_ byte = iota
	varSet_redirectid
)

func (sc varSet) Run(c *Char, _ []int32) bool {
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case varSet_:
			exp[0].run(crun)
		case varSet_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type turn StateControllerBase

const (
	turn_ byte = iota
	turn_redirectid
)

func (sc turn) Run(c *Char, _ []int32) bool {
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case turn_:
			crun.setFacing(-crun.facing)
		case turn_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type targetFacing StateControllerBase

const (
	targetFacing_id byte = iota
	targetFacing_value
	targetFacing_redirectid
)

func (sc targetFacing) Run(c *Char, _ []int32) bool {
	crun := c
	tar := crun.getTarget(-1)
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case targetFacing_id:
			if len(tar) == 0 {
				return false
			}
			tar = crun.getTarget(exp[0].evalI(c))
		case targetFacing_value:
			if len(tar) == 0 {
				return false
			}
			crun.targetFacing(tar, exp[0].evalI(c))
		case targetFacing_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
				tar = crun.getTarget(-1)
				if len(tar) == 0 {
					return false
				}
			} else {
				return false
			}

		}
		return true
	})
	if len(tar) == 0 {
		return false
	}
	return false
}

type targetBind StateControllerBase

const (
	targetBind_id byte = iota
	targetBind_time
	targetBind_pos
	targetBind_redirectid
)

func (sc targetBind) Run(c *Char, _ []int32) bool {
	crun := c
	var lclscround float32 = 1.0
	tar := crun.getTarget(-1)
	t := int32(1)
	var x, y float32 = 0, 0
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case targetBind_id:
			if len(tar) == 0 {
				return false
			}
			tar = crun.getTarget(exp[0].evalI(c))
		case targetBind_time:
			t = exp[0].evalI(c)
		case targetBind_pos:
			x = exp[0].evalF(c) * lclscround
			if len(exp) > 1 {
				y = exp[1].evalF(c) * lclscround
			}
		case targetBind_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
				lclscround = c.localscl / crun.localscl
				tar = crun.getTarget(-1)
				if len(tar) == 0 {
					return false
				}
			} else {
				return false
			}

		}
		return true
	})
	if len(tar) == 0 {
		return false
	}
	crun.targetBind(tar, t, x, y)
	return false
}

type bindToTarget StateControllerBase

const (
	bindToTarget_id byte = iota
	bindToTarget_time
	bindToTarget_pos
	bindToTarget_redirectid
)

func (sc bindToTarget) Run(c *Char, _ []int32) bool {
	crun := c
	var lclscround float32 = 1.0
	tar := crun.getTarget(-1)
	t, x, y, hmf := int32(1), float32(0), float32(math.NaN()), HMF_F
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case bindToTarget_id:
			if len(tar) == 0 {
				return false
			}
			tar = crun.getTarget(exp[0].evalI(c))
		case bindToTarget_time:
			t = exp[0].evalI(c)
		case bindToTarget_pos:
			x = exp[0].evalF(c) * lclscround
			if len(exp) > 1 {
				y = exp[1].evalF(c) * lclscround
				if len(exp) > 2 {
					hmf = HMF(exp[2].evalI(c))
				}
			}
		case bindToTarget_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
				lclscround = c.localscl / crun.localscl
				tar = crun.getTarget(-1)
				if len(tar) == 0 {
					return false
				}
			} else {
				return false
			}
		}
		return true
	})
	if len(tar) == 0 {
		return false
	}
	crun.bindToTarget(tar, t, x, y, hmf)
	return false
}

type targetLifeAdd StateControllerBase

const (
	targetLifeAdd_id byte = iota
	targetLifeAdd_absolute
	targetLifeAdd_kill
	targetLifeAdd_value
	targetLifeAdd_redirectid
)

func (sc targetLifeAdd) Run(c *Char, _ []int32) bool {
	crun := c
	tar, a, k := crun.getTarget(-1), false, true
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case targetLifeAdd_id:
			if len(tar) == 0 {
				return false
			}
			tar = crun.getTarget(exp[0].evalI(c))
		case targetLifeAdd_absolute:
			a = exp[0].evalB(c)
		case targetLifeAdd_kill:
			k = exp[0].evalB(c)
		case targetLifeAdd_value:
			if len(tar) == 0 {
				return false
			}
			crun.targetLifeAdd(tar, exp[0].evalI(c), k, a)
		case targetLifeAdd_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
				tar = crun.getTarget(-1)
				if len(tar) == 0 {
					return false
				}
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type targetState StateControllerBase

const (
	targetState_id byte = iota
	targetState_value
	targetState_redirectid
)

func (sc targetState) Run(c *Char, _ []int32) bool {
	crun := c
	tar := crun.getTarget(-1)
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case targetState_id:
			if len(tar) == 0 {
				return false
			}
			tar = crun.getTarget(exp[0].evalI(c))
		case targetState_value:
			if len(tar) == 0 {
				return false
			}
			crun.targetState(tar, exp[0].evalI(c))
		case targetState_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
				tar = crun.getTarget(-1)
				if len(tar) == 0 {
					return false
				}
			} else {
				return false
			}
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
	targetVelSet_redirectid
)

func (sc targetVelSet) Run(c *Char, _ []int32) bool {
	crun := c
	var lclscround float32 = 1.0
	tar := crun.getTarget(-1)
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case targetVelSet_id:
			if len(tar) == 0 {
				return false
			}
			tar = crun.getTarget(exp[0].evalI(c))
		case targetVelSet_x:
			if len(tar) == 0 {
				return false
			}
			crun.targetVelSetX(tar, exp[0].evalF(c)*lclscround)
		case targetVelSet_y:
			if len(tar) == 0 {
				return false
			}
			crun.targetVelSetY(tar, exp[0].evalF(c)*lclscround)
		case targetVelSet_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
				lclscround = c.localscl / crun.localscl
				tar = crun.getTarget(-1)
				if len(tar) == 0 {
					return false
				}
			} else {
				return false
			}
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
	targetVelAdd_redirectid
)

func (sc targetVelAdd) Run(c *Char, _ []int32) bool {
	crun := c
	var lclscround float32 = 1.0
	tar := crun.getTarget(-1)
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case targetVelAdd_id:
			if len(tar) == 0 {
				return false
			}
			tar = crun.getTarget(exp[0].evalI(c))
		case targetVelAdd_x:
			if len(tar) == 0 {
				return false
			}
			crun.targetVelAddX(tar, exp[0].evalF(c)*lclscround)
		case targetVelAdd_y:
			if len(tar) == 0 {
				return false
			}
			crun.targetVelAddY(tar, exp[0].evalF(c)*lclscround)
		case targetVelAdd_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
				lclscround = c.localscl / crun.localscl
				tar = crun.getTarget(-1)
				if len(tar) == 0 {
					return false
				}
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type targetPowerAdd StateControllerBase

const (
	targetPowerAdd_id byte = iota
	targetPowerAdd_value
	targetPowerAdd_redirectid
)

func (sc targetPowerAdd) Run(c *Char, _ []int32) bool {
	crun := c
	tar := crun.getTarget(-1)
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case targetPowerAdd_id:
			if len(tar) == 0 {
				return false
			}
			tar = crun.getTarget(exp[0].evalI(c))
		case targetPowerAdd_value:
			if len(tar) == 0 {
				return false
			}
			crun.targetPowerAdd(tar, exp[0].evalI(c))
		case targetPowerAdd_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
				tar = crun.getTarget(-1)
				if len(tar) == 0 {
					return false
				}
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type targetDrop StateControllerBase

const (
	targetDrop_excludeid byte = iota
	targetDrop_keepone
	targetDrop_redirectid
)

func (sc targetDrop) Run(c *Char, _ []int32) bool {
	crun := c
	tar, eid, ko := crun.getTarget(-1), int32(-1), true
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case targetDrop_excludeid:
			eid = exp[0].evalI(c)
		case targetDrop_keepone:
			ko = exp[0].evalB(c)
		case targetDrop_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
				tar = crun.getTarget(-1)
				if len(tar) == 0 {
					return false
				}
			} else {
				return false
			}
		}
		return true
	})
	if len(tar) == 0 {
		return false
	}
	crun.targetDrop(eid, ko)
	return false
}

type lifeAdd StateControllerBase

const (
	lifeAdd_absolute byte = iota
	lifeAdd_kill
	lifeAdd_value
	lifeAdd_redirectid
)

func (sc lifeAdd) Run(c *Char, _ []int32) bool {
	a, k := false, true
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case lifeAdd_absolute:
			a = exp[0].evalB(c)
		case lifeAdd_kill:
			k = exp[0].evalB(c)
		case lifeAdd_value:
			crun.lifeAdd(float64(exp[0].evalI(c)), k, a)
		case lifeAdd_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type lifeSet StateControllerBase

const (
	lifeSet_value byte = iota
	lifeSet_redirectid
)

func (sc lifeSet) Run(c *Char, _ []int32) bool {
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case lifeSet_value:
			crun.lifeSet(exp[0].evalI(c))
		case lifeSet_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type powerAdd StateControllerBase

const (
	powerAdd_value byte = iota
	powerAdd_redirectid
)

func (sc powerAdd) Run(c *Char, _ []int32) bool {
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case powerAdd_value:
			crun.powerAdd(exp[0].evalI(c))
		case powerAdd_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type powerSet StateControllerBase

const (
	powerSet_value byte = iota
	powerSet_redirectid
)

func (sc powerSet) Run(c *Char, _ []int32) bool {
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case powerSet_value:
			crun.powerSet(exp[0].evalI(c))
		case powerSet_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type hitVelSet StateControllerBase

const (
	hitVelSet_x byte = iota
	hitVelSet_y
	hitVelSet_redirectid
)

func (sc hitVelSet) Run(c *Char, _ []int32) bool {
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case hitVelSet_x:
			if exp[0].evalB(c) {
				crun.hitVelSetX()
			}
		case hitVelSet_y:
			if exp[0].evalB(c) {
				crun.hitVelSetY()
			}
		case hitVelSet_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
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
	screenBound_redirectid
)

func (sc screenBound) Run(c *Char, _ []int32) bool {
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case screenBound_value:
			if exp[0].evalB(c) {
				crun.setSF(CSF_screenbound)
			} else {
				crun.unsetSF(CSF_screenbound)
			}
		case screenBound_movecamera:
			if exp[0].evalB(c) {
				crun.setSF(CSF_movecamera_x)
			} else {
				crun.unsetSF(CSF_movecamera_x)
			}
			if len(exp) > 1 {
				if exp[1].evalB(c) {
					crun.setSF(CSF_movecamera_y)
				} else {
					crun.unsetSF(CSF_movecamera_y)
				}
			} else {
				crun.unsetSF(CSF_movecamera_y)
			}
		case screenBound_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type posFreeze StateControllerBase

const (
	posFreeze_value byte = iota
	posFreeze_redirectid
)

func (sc posFreeze) Run(c *Char, _ []int32) bool {
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case posFreeze_value:
			if exp[0].evalB(c) {
				crun.setSF(CSF_posfreeze)
			}
		case posFreeze_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
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
			sys.envShake.ampl = int32(float32(exp[0].evalI(c)) * c.localscl)
		case envShake_phase:
			sys.envShake.phase = MaxF(0, exp[0].evalF(c)*float32(math.Pi)/180) * c.localscl
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
	hitOverride_redirectid
)

func (sc hitOverride) Run(c *Char, _ []int32) bool {
	crun := c
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
			if t < -1 || t == 0 {
				t = 1
			}
		case hitOverride_forceair:
			f = exp[0].evalB(c)
		case hitOverride_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	if st < 0 {
		t = 0
	}
	pn := crun.playerNo
	crun.ho[s] = HitOverride{attr: a, stateno: st, time: t, forceair: f,
		playerNo: pn}
	return false
}

type pause StateControllerBase

const (
	pause_time byte = iota
	pause_movetime
	pause_pausebg
	pause_endcmdbuftime
	pause_redirectid
)

func (sc pause) Run(c *Char, _ []int32) bool {
	crun := c
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
		case pause_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	crun.setPauseTime(t, mt)
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
	superPause_redirectid
)

func (sc superPause) Run(c *Char, _ []int32) bool {
	crun := c
	var t, mt int32 = 30, 0
	uh := true
	sys.superanim, sys.superpmap.remap = crun.getAnim(100, true, false), nil
	sys.superpos, sys.superfacing = [...]float32{crun.pos[0] * crun.localscl, crun.pos[1] * crun.localscl}, crun.facing
	sys.superpausebg, sys.superendcmdbuftime, sys.superdarken = true, 0, true
	sys.superp2defmul = crun.gi().constants["super.targetdefencemul"]
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
			if sys.superanim = crun.getAnim(exp[1].evalI(c), f, false); sys.superanim != nil {
				if f {
					sys.superpmap.remap = nil
				} else {
					sys.superpmap.remap = crun.getPalMap()
				}
			}
		case superPause_pos:
			sys.superpos[0] += crun.facing * exp[0].evalF(c) * c.localscl
			if len(exp) > 1 {
				sys.superpos[1] += exp[1].evalF(c) * c.localscl
			}
		case superPause_p2defmul:
			sys.superp2defmul = exp[0].evalF(c)
			if sys.superp2defmul == 0 {
				sys.superp2defmul = crun.gi().constants["super.targetdefencemul"]
			}
		case superPause_poweradd:
			crun.powerAdd(exp[0].evalI(c))
		case superPause_unhittable:
			uh = exp[0].evalB(c)
		case superPause_sound:
			n := int32(0)
			if len(exp) > 2 {
				n = exp[2].evalI(c)
			}
			vo := int32(100)
			crun.playSound(exp[0].evalB(c), false, false, exp[1].evalI(c), n, -1,
				vo, 0, 1, &crun.pos[0], false)
		case superPause_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
				sys.superanim, sys.superpmap.remap = crun.getAnim(30, true, false), nil
				sys.superpos, sys.superfacing = [...]float32{crun.pos[0] * crun.localscl, crun.pos[1] * crun.localscl}, crun.facing
			} else {
				return false
			}
		}
		return true
	})
	crun.setSuperPauseTime(t, mt, uh)
	return false
}

type trans StateControllerBase

const (
	trans_trans byte = iota
	trans_redirectid
)

func (sc trans) Run(c *Char, _ []int32) bool {
	crun := c
	crun.alpha[1] = 255
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case trans_trans:
			crun.alpha[0] = exp[0].evalI(c)
			crun.alpha[1] = exp[1].evalI(c)
			if len(exp) >= 3 {
				crun.alpha[0] = Max(0, Min(255, crun.alpha[0]))
				crun.alpha[1] = Max(0, Min(255, crun.alpha[1]))
				if len(exp) >= 4 {
					crun.alpha[1] = ^crun.alpha[1]
				} else if crun.alpha[0] == 1 && crun.alpha[1] == 255 {
					crun.alpha[0] = 0
				}
			}
		case trans_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	crun.setSF(CSF_trans)
	return false
}

type playerPush StateControllerBase

const (
	playerPush_value byte = iota
	playerPush_redirectid
)

func (sc playerPush) Run(c *Char, _ []int32) bool {
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case playerPush_value:
			if exp[0].evalB(c) {
				crun.setSF(CSF_playerpush)
			} else {
				crun.unsetSF(CSF_playerpush)
			}
		case playerPush_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type stateTypeSet StateControllerBase

const (
	stateTypeSet_statetype byte = iota
	stateTypeSet_movetype
	stateTypeSet_physics
	stateTypeSet_redirectid
)

func (sc stateTypeSet) Run(c *Char, _ []int32) bool {
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case stateTypeSet_statetype:
			crun.ss.stateType = StateType(exp[0].evalI(c))
		case stateTypeSet_movetype:
			crun.ss.moveType = MoveType(exp[0].evalI(c))
		case stateTypeSet_physics:
			crun.ss.physics = StateType(exp[0].evalI(c))
		case stateTypeSet_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type angleDraw StateControllerBase

const (
	angleDraw_value byte = iota
	angleDraw_scale
	angleDraw_redirectid
)

func (sc angleDraw) Run(c *Char, _ []int32) bool {
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case angleDraw_value:
			crun.angleSet(exp[0].evalF(c))
		case angleDraw_scale:
			crun.angleScalse[0] *= exp[0].evalF(c)
			if len(exp) > 1 {
				crun.angleScalse[1] *= exp[1].evalF(c)
			}
		case angleDraw_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	crun.setSF(CSF_angledraw)
	return false
}

type angleSet StateControllerBase

const (
	angleSet_value byte = iota
	angleSet_redirectid
)

func (sc angleSet) Run(c *Char, _ []int32) bool {
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case angleSet_value:
			crun.angleSet(exp[0].evalF(c))
		case angleSet_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type angleAdd StateControllerBase

const (
	angleAdd_value byte = iota
	angleAdd_redirectid
)

func (sc angleAdd) Run(c *Char, _ []int32) bool {
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case angleAdd_value:
			crun.angleSet(crun.angle + exp[0].evalF(c))
		case angleAdd_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type angleMul StateControllerBase

const (
	angleMul_value byte = iota
	angleMul_redirectid
)

func (sc angleMul) Run(c *Char, _ []int32) bool {
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case angleMul_value:
			crun.angleSet(crun.angle * exp[0].evalF(c))
		case angleMul_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type envColor StateControllerBase

const (
	envColor_value byte = iota
	envColor_time
	envColor_under
)

func (sc envColor) Run(c *Char, _ []int32) bool {
	sys.envcol = [...]int32{255, 255, 255}
	sys.envcol_time = 1
	sys.envcol_under = false
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case envColor_value:
			sys.envcol[0] = exp[0].evalI(c)
			sys.envcol[1] = exp[1].evalI(c)
			sys.envcol[2] = exp[2].evalI(c)
		case envColor_time:
			sys.envcol_time = exp[0].evalI(c)
		case envColor_under:
			sys.envcol_under = exp[0].evalB(c)
		}
		return true
	})
	return false
}

type displayToClipboard StateControllerBase

const (
	displayToClipboard_params byte = iota
	displayToClipboard_text
	displayToClipboard_redirectid
)

func (sc displayToClipboard) Run(c *Char, _ []int32) bool {
	crun := c
	params := []interface{}{}
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case displayToClipboard_params:
			for _, e := range exp {
				if bv := e.run(c); bv.t == VT_Float {
					params = append(params, bv.ToF())
				} else {
					params = append(params, bv.ToI())
				}
			}
		case displayToClipboard_text:
			crun.clipboardText = nil
			crun.appendToClipboard(sys.workingState.playerNo,
				int(exp[0].evalI(c)), params...)
		case displayToClipboard_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type appendToClipboard displayToClipboard

func (sc appendToClipboard) Run(c *Char, _ []int32) bool {
	crun := c
	params := []interface{}{}
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case displayToClipboard_params:
			for _, e := range exp {
				if bv := e.run(c); bv.t == VT_Float {
					params = append(params, bv.ToF())
				} else {
					params = append(params, bv.ToI())
				}
			}
		case displayToClipboard_text:
			crun.appendToClipboard(sys.workingState.playerNo,
				int(exp[0].evalI(c)), params...)
		case displayToClipboard_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type clearClipboard StateControllerBase

const (
	clearClipboard_ byte = iota
	clearClipboard_redirectid
)

func (sc clearClipboard) Run(c *Char, _ []int32) bool {
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case clearClipboard_:
			crun.clipboardText = nil
		case clearClipboard_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type makeDust StateControllerBase

const (
	makeDust_spacing byte = iota
	makeDust_pos
	makeDust_pos2
	makeDust_redirectid
)

func (sc makeDust) Run(c *Char, _ []int32) bool {
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case makeDust_spacing:
			s := Max(1, exp[0].evalI(c))
			if crun.time()%s != s-1 {
				return false
			}
		case makeDust_pos:
			x, y := exp[0].evalF(c), float32(0)
			if len(exp) > 1 {
				y = exp[1].evalF(c)
			}
			crun.makeDust(x-float32(crun.size.draw.offset[0]),
				y-float32(crun.size.draw.offset[1]))
		case makeDust_pos2:
			x, y := exp[0].evalF(c), float32(0)
			if len(exp) > 1 {
				y = exp[1].evalF(c)
			}
			crun.makeDust(x-float32(crun.size.draw.offset[0]),
				y-float32(crun.size.draw.offset[1]))
		case makeDust_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type attackDist StateControllerBase

const (
	attackDist_value byte = iota
	attackDist_redirectid
)

func (sc attackDist) Run(c *Char, _ []int32) bool {
	crun := c
	var lclscround float32 = 1.0
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case attackDist_value:
			crun.attackDist = exp[0].evalF(c) * lclscround
		case attackDist_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
				lclscround = c.localscl / crun.localscl
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type attackMulSet StateControllerBase

const (
	attackMulSet_value byte = iota
	attackMulSet_redirectid
)

func (sc attackMulSet) Run(c *Char, _ []int32) bool {
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case attackMulSet_value:
			crun.attackMul = float32(crun.gi().data.attack) * crun.ocd().attackRatio / 100 * exp[0].evalF(c)
		case attackMulSet_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type defenceMulSet StateControllerBase

const (
	defenceMulSet_value byte = iota
	defenceMulSet_redirectid
)

func (sc defenceMulSet) Run(c *Char, _ []int32) bool {
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case defenceMulSet_value:
			crun.customDefense = exp[0].evalF(c)
		case defenceMulSet_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type fallEnvShake StateControllerBase

const (
	fallEnvShake_ byte = iota
	fallEnvShake_redirectid
)

func (sc fallEnvShake) Run(c *Char, _ []int32) bool {
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case fallEnvShake_:
			sys.envShake = EnvShake{time: crun.ghv.fall.envshake_time,
				freq: crun.ghv.fall.envshake_freq * math.Pi / 180,
				ampl: crun.ghv.fall.envshake_ampl, phase: crun.ghv.fall.envshake_phase}
			sys.envShake.setDefPhase()
		case fallEnvShake_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type hitFallDamage StateControllerBase

const (
	hitFallDamage_ byte = iota
	hitFallDamage_redirectid
)

func (sc hitFallDamage) Run(c *Char, _ []int32) bool {
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case hitFallDamage_:
			crun.hitFallDamage()
		case hitFallDamage_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type hitFallVel StateControllerBase

const (
	hitFallVel_ byte = iota
	hitFallVel_redirectid
)

func (sc hitFallVel) Run(c *Char, _ []int32) bool {
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case hitFallVel_:
			crun.hitFallVel()
		case hitFallVel_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type hitFallSet StateControllerBase

const (
	hitFallSet_value byte = iota
	hitFallSet_xvel
	hitFallSet_yvel
	hitFallSet_redirectid
)

func (sc hitFallSet) Run(c *Char, _ []int32) bool {
	crun := c
	f, xv, yv := int32(-1), float32(math.NaN()), float32(math.NaN())
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case hitFallSet_value:
			f = exp[0].evalI(c)
			if len(crun.ghv.hitBy) == 0 {
				return false
			}
		case hitFallSet_xvel:
			xv = exp[0].evalF(c)
		case hitFallSet_yvel:
			yv = exp[0].evalF(c)
		case hitFallSet_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	crun.hitFallSet(f, xv, yv)
	return false
}

type varRangeSet StateControllerBase

const (
	varRangeSet_first byte = iota
	varRangeSet_last
	varRangeSet_value
	varRangeSet_fvalue
	varRangeSet_redirectid
)

func (sc varRangeSet) Run(c *Char, _ []int32) bool {
	crun := c
	var first, last int32 = 0, 0
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case varRangeSet_first:
			first = exp[0].evalI(c)
		case varRangeSet_last:
			last = exp[0].evalI(c)
		case varRangeSet_value:
			v := exp[0].evalI(c)
			if first >= 0 && last < int32(NumVar) {
				for i := first; i <= last; i++ {
					crun.ivar[i] = v
				}
			}
		case varRangeSet_fvalue:
			fv := exp[0].evalF(c)
			if first >= 0 && last < int32(NumFvar) {
				for i := first; i <= last; i++ {
					crun.fvar[i] = fv
				}
			}
		case varRangeSet_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type remapPal StateControllerBase

const (
	remapPal_source byte = iota
	remapPal_dest
	remapPal_redirectid
)

func (sc remapPal) Run(c *Char, _ []int32) bool {
	crun := c
	src := [...]int32{-1, -1}
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case remapPal_source:
			src[0] = exp[0].evalI(c)
			if len(exp) > 1 {
				src[1] = exp[1].evalI(c)
			}
			if src[0] == -1 {
				src[0] = 1
				src[1] = 1
			}
		case remapPal_dest:
			dst := [...]int32{exp[0].evalI(c), -1}
			if len(exp) > 1 {
				dst[1] = exp[1].evalI(c)
			}
			crun.remapPal(crun.getPalfx(), src, dst)
		case remapPal_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type stopSnd StateControllerBase

const (
	stopSnd_channel byte = iota
	stopSnd_redirectid
)

func (sc stopSnd) Run(c *Char, _ []int32) bool {
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case stopSnd_channel:
			if ch := Min(255, exp[0].evalI(c)); ch < 0 {
				sys.stopAllSound()
			} else if int(ch) < len(crun.sounds) {
				crun.sounds[ch].sound = nil
			}
		case stopSnd_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type sndPan StateControllerBase

const (
	sndPan_channel byte = iota
	sndPan_pan
	sndPan_abspan
	sndPan_redirectid
)

func (sc sndPan) Run(c *Char, _ []int32) bool {
	crun := c
	ch, pan, x := int32(-1), float32(0), &crun.pos[0]
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case sndPan_channel:
			ch = exp[0].evalI(c)
		case sndPan_pan:
			pan = exp[0].evalF(c)
		case sndPan_abspan:
			pan = exp[0].evalF(c)
			x = nil
		case sndPan_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
				x = &crun.pos[0]
			} else {
				return false
			}
		}
		return true
	})
	if ch <= 0 && int(ch) < len(crun.sounds) {
		crun.sounds[ch].SetPan(pan, x)
	}
	return false
}

type varRandom StateControllerBase

const (
	varRandom_v byte = iota
	varRandom_range
	varRandom_redirectid
)

func (sc varRandom) Run(c *Char, _ []int32) bool {
	crun := c
	var v int32
	var min, max int32 = 0, 1000
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case varRandom_v:
			v = exp[0].evalI(c)
		case varRandom_range:
			min, max = 0, exp[0].evalI(c)
			if len(exp) > 1 {
				min, max = max, exp[1].evalI(c)
			}
		case varRandom_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	crun.varSet(v, RandI(min, max))
	return false
}

type gravity StateControllerBase

const (
	gravity_ byte = iota
	gravity_redirectid
)

func (sc gravity) Run(c *Char, _ []int32) bool {
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case gravity_:
			crun.gravity()
		case gravity_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type bindToParent StateControllerBase

const (
	bindToParent_time byte = iota
	bindToParent_facing
	bindToParent_pos
	bindToParent_redirectid
)

func (sc bindToParent) Run(c *Char, _ []int32) bool {
	crun := c
	var lclscround float32 = 1.0
	p := crun.parent()
	var x, y float32 = 0, 0
	var time int32 = 1
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case bindToParent_time:
			time = exp[0].evalI(c)
		case bindToParent_facing:
			if f := exp[0].evalI(c); f < 0 {
				crun.bindFacing = -1
			} else if f > 0 {
				crun.bindFacing = 1
			}
		case bindToParent_pos:
			x = exp[0].evalF(c) * lclscround
			if len(exp) > 1 {
				y = exp[1].evalF(c) * lclscround
			}
		case bindToParent_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
				lclscround = c.localscl / crun.localscl
				p = crun.parent()
			} else {
				return false
			}
		}
		return true
	})
	if p == nil {
		return false
	}
	crun.bindPos[0] = x
	crun.bindPos[1] = y
	crun.setBindTime(time)
	crun.setBindToId(p)
	return false
}

type bindToRoot bindToParent

func (sc bindToRoot) Run(c *Char, _ []int32) bool {
	crun := c
	var lclscround float32 = 1.0
	r := crun.root()
	var x, y float32 = 0, 0
	var time int32 = 1
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case bindToParent_time:
			time = exp[0].evalI(c)
		case bindToParent_facing:
			if f := exp[0].evalI(c); f < 0 {
				crun.bindFacing = -1
			} else if f > 0 {
				crun.bindFacing = 1
			}
		case bindToParent_pos:
			x = exp[0].evalF(c) * lclscround
			if len(exp) > 1 {
				y = exp[1].evalF(c) * lclscround
			}
		case bindToParent_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
				lclscround = c.localscl / crun.localscl
				r = crun.root()
			} else {
				return false
			}
		}
		return true
	})
	if r == nil {
		return false
	}
	crun.bindPos[0] = x
	crun.bindPos[1] = y
	crun.setBindTime(time)
	crun.setBindToId(r)
	return false
}

type removeExplod StateControllerBase

const (
	removeExplod_id byte = iota
	removeExplod_redirectid
)

func (sc removeExplod) Run(c *Char, _ []int32) bool {
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case removeExplod_id:
			crun.removeExplod(exp[0].evalI(c))
		case removeExplod_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type explodBindTime StateControllerBase

const (
	explodBindTime_id byte = iota
	explodBindTime_time
	explodBindTime_redirectid
)

func (sc explodBindTime) Run(c *Char, _ []int32) bool {
	crun := c
	var eid, time int32 = -1, 0
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case explodBindTime_id:
			eid = exp[0].evalI(c)
		case explodBindTime_time:
			time = exp[0].evalI(c)
		case explodBindTime_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	crun.explodBindTime(eid, time)
	return false
}

type moveHitReset StateControllerBase

const (
	moveHitReset_ byte = iota
	moveHitReset_redirectid
)

func (sc moveHitReset) Run(c *Char, _ []int32) bool {
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case moveHitReset_:
			crun.clearMoveHit()
		case moveHitReset_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type hitAdd StateControllerBase

const (
	hitAdd_value byte = iota
	hitAdd_redirectid
)

func (sc hitAdd) Run(c *Char, _ []int32) bool {
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case hitAdd_value:
			crun.hitAdd(exp[0].evalI(c))
		case hitAdd_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type offset StateControllerBase

const (
	offset_x byte = iota
	offset_y
	offset_redirectid
)

func (sc offset) Run(c *Char, _ []int32) bool {
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case offset_x:
			crun.offset[0] = exp[0].evalF(c)
		case offset_y:
			crun.offset[1] = exp[0].evalF(c)
		case offset_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type victoryQuote StateControllerBase

const (
	victoryQuote_value byte = iota
	victoryQuote_redirectid
)

func (sc victoryQuote) Run(c *Char, _ []int32) bool {
	crun := c
	var v int32 = -1
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case victoryQuote_value:
			v = exp[0].evalI(c)
		case victoryQuote_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	crun.winquote = v
	return false
}

type zoom StateControllerBase

const (
	zoom_pos byte = iota
	zoom_scale
	zoom_lag
	zoom_redirectid
)

func (sc zoom) Run(c *Char, _ []int32) bool {
	crun := c
	zoompos := [2]float32{0, 0}
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case zoom_pos:
			zoompos[0] = exp[0].evalF(c) * crun.localscl
			if len(exp) > 1 {
				zoompos[1] = exp[1].evalF(c) * crun.localscl
			}
		case zoom_scale:
			sys.zoomScale = exp[0].evalF(c)
			sys.enableZoomstate = true
		case zoom_lag:
			sys.zoomlag = exp[0].evalF(c)
		case zoom_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	sys.zoomPos[0] = sys.zoomScale * zoompos[0]
	sys.zoomPos[1] = zoompos[1]
	return false
}

type dialogue StateControllerBase

const (
	dialogue_hidebars byte = iota
	dialogue_force
	dialogue_text
	dialogue_redirectid
)

func (sc dialogue) Run(c *Char, _ []int32) bool {
	crun := c
	reset := true
	force := false
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case dialogue_hidebars:
			sys.dialogueBarsFlg = exp[0].evalB(c)
		case dialogue_force:
			force = exp[0].evalB(c)
		case dialogue_text:
			sys.chars[crun.playerNo][0].appendDialogue(string(*(*[]byte)(unsafe.Pointer(&exp[0]))), reset)
			reset = false
		case dialogue_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	if force {
		sys.dialogueFlg = true
		sys.dialogueForce = crun.playerNo + 1
	}
	return false
}

type dizzyPointsAdd StateControllerBase

const (
	dizzyPointsAdd_value byte = iota
	dizzyPointsAdd_redirectid
)

func (sc dizzyPointsAdd) Run(c *Char, _ []int32) bool {
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case dizzyPointsAdd_value:
			crun.dizzyPointsAdd(exp[0].evalI(c))
		case dizzyPointsAdd_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type dizzyPointsSet StateControllerBase

const (
	dizzyPointsSet_value byte = iota
	dizzyPointsSet_redirectid
)

func (sc dizzyPointsSet) Run(c *Char, _ []int32) bool {
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case dizzyPointsSet_value:
			crun.dizzyPointsSet(exp[0].evalI(c))
		case dizzyPointsSet_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type dizzySet StateControllerBase

const (
	dizzySet_value byte = iota
	dizzySet_redirectid
)

func (sc dizzySet) Run(c *Char, _ []int32) bool {
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case dizzySet_value:
			crun.setDizzy(exp[0].evalB(c))
		case dizzySet_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type guardBreakSet StateControllerBase

const (
	guardBreakSet_value byte = iota
	guardBreakSet_redirectid
)

func (sc guardBreakSet) Run(c *Char, _ []int32) bool {
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case guardBreakSet_value:
			crun.setGuardBreak(exp[0].evalB(c))
		case guardBreakSet_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type guardPointsAdd StateControllerBase

const (
	guardPointsAdd_value byte = iota
	guardPointsAdd_redirectid
)

func (sc guardPointsAdd) Run(c *Char, _ []int32) bool {
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case guardPointsAdd_value:
			crun.guardPointsAdd(exp[0].evalI(c))
		case guardPointsAdd_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type guardPointsSet StateControllerBase

const (
	guardPointsSet_value byte = iota
	guardPointsSet_redirectid
)

func (sc guardPointsSet) Run(c *Char, _ []int32) bool {
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case guardPointsSet_value:
			crun.guardPointsSet(exp[0].evalI(c))
		case guardPointsSet_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type hitScaleSet StateControllerBase

const (
	hitScaleSet_id byte = iota
	hitScaleSet_affects_damage
	hitScaleSet_affects_hitTime
	hitScaleSet_affects_pauseTime
	hitScaleSet_mul
	hitScaleSet_add
	hitScaleSet_addType
	hitScaleSet_min
	hitScaleSet_max
	hitScaleSet_time
	hitScaleSet_reset
	hitScaleSet_force
	hitScaleSet_redirectid
)

// Takes the values given by Compiler.hitScaleSet and executes it.
func (sc hitScaleSet) Run(c *Char, _ []int32) bool {
	var crun = c
	// Default values
	var affects = []bool{false, false, false}
	// Target of the hitScale, -1 is default.
	var target int32 = -1
	var targetArray [3]*HitScale
	// Do we reset everithng back to default?
	var resetAll = false
	var reset = false
	// If false we wait to hit to apply hitScale.
	// If true we apply on call.
	var force = false
	// Holder variables
	var tempHitScale = newHitScale()

	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		// What is hitScale ging to affect.
		case hitScaleSet_affects_damage:
			affects[0] = true
		case hitScaleSet_affects_hitTime:
			affects[1] = true
		case hitScaleSet_affects_pauseTime:
			affects[2] = true
		// ID of the char to apply to.
		case hitScaleSet_id:
			target = exp[0].evalI(c)
		case hitScaleSet_mul:
			tempHitScale.mul = exp[0].evalF(c)
		case hitScaleSet_add:
			tempHitScale.add = exp[0].evalI(c)
		case hitScaleSet_addType:
			tempHitScale.addType = exp[0].evalI(c)
		case hitScaleSet_min:
			tempHitScale.min = exp[0].evalF(c)
		case hitScaleSet_max:
			tempHitScale.max = exp[0].evalF(c)
		case hitScaleSet_time:
			tempHitScale.time = exp[0].evalI(c)
		case hitScaleSet_reset:
			if exp[0].evalI(c) == 1 {
				reset = true
			} else if exp[0].evalI(c) == 2 {
				resetAll = true
			}
		case hitScaleSet_force:
			force = exp[0].evalB(c)
		// Genric redirectId.
		case hitScaleSet_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})

	// ----------------------------------------------------------------------

	if resetAll {
		for _, hs := range crun.defaultHitScale {
			hs.reset()
		}
		crun.nextHitScale = make(map[int32][3]*HitScale)
		crun.activeHitScale = make(map[int32][3]*HitScale)
	}

	targetArray = getHitScaleTarget(crun, target, force, reset)

	// Apply the new values and activate it.
	for i, hs := range targetArray {
		if affects[i] {
			if reset {
				if ahs, ok := crun.activeHitScale[target]; ok {
					ahs[int32(i)].reset()
				}
			}
			hs.copy(tempHitScale)
			hs.active = true
		}
	}

	return false
}

func getHitScaleTarget(char *Char, target int32, force bool, reset bool) [3]*HitScale {
	// Get our targets.
	if target <= -1 {
		return char.defaultHitScale
	} else { //Check if target exists.
		if force {
			if _, ok := char.activeHitScale[target]; !ok || reset {
				char.activeHitScale[target] = newHitScaleArray()
			}
			return char.activeHitScale[target]
		} else {
			if _, ok := char.nextHitScale[target]; !ok || reset {
				char.nextHitScale[target] = newHitScaleArray()
			}
			return char.nextHitScale[target]
		}
	}
}

type lifebarAction StateControllerBase

const (
	lifebarAction_top byte = iota
	lifebarAction_time
	lifebarAction_timemul
	lifebarAction_anim
	lifebarAction_spr
	lifebarAction_snd
	lifebarAction_text
	lifebarAction_redirectid
)

func (sc lifebarAction) Run(c *Char, _ []int32) bool {
	crun := c
	var top bool
	var text string
	var timemul float32 = 1
	var time, anim int32 = -1, -1
	spr := [2]int32{-1, -1}
	snd := [2]int32{-1, -1}
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case lifebarAction_top:
			top = exp[0].evalB(c)
		case lifebarAction_timemul:
			timemul = float32(exp[0].evalF(c))
		case lifebarAction_time:
			time = int32(exp[0].evalI(c))
		case lifebarAction_anim:
			anim = int32(exp[0].evalI(c))
		case lifebarAction_spr:
			spr = [2]int32{int32(exp[0].evalI(c)), int32(exp[1].evalI(c))}
		case lifebarAction_snd:
			snd = [2]int32{int32(exp[0].evalI(c)), int32(exp[1].evalI(c))}
		case lifebarAction_text:
			text = string(*(*[]byte)(unsafe.Pointer(&exp[0])))
		case lifebarAction_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	crun.appendLifebarAction(text, snd, spr, anim, time, timemul, top)
	return false
}

type loadFile StateControllerBase

const (
	loadFile_path byte = iota
	loadFile_saveData
	loadFile_redirectid
)

func (sc loadFile) Run(c *Char, _ []int32) bool {
	crun := c
	var path string
	var data SaveData
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case loadFile_path:
			path = string(*(*[]byte)(unsafe.Pointer(&exp[0])))
		case loadFile_saveData:
			data = SaveData(exp[0].evalI(c))
		case loadFile_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	if path != "" {
		decodeFile, err := os.Open(filepath.Dir(c.gi().def) + "/" + path)
		if err != nil {
			defer decodeFile.Close()
			return false
		}
		defer decodeFile.Close()
		decoder := gob.NewDecoder(decodeFile)
		switch data {
		case SaveData_map:
			if err := decoder.Decode(&crun.mapArray); err != nil {
				panic(err)
			}
		case SaveData_var:
			if err := decoder.Decode(&crun.ivar); err != nil {
				panic(err)
			}
		case SaveData_fvar:
			if err := decoder.Decode(&crun.fvar); err != nil {
				panic(err)
			}
		}
	}
	return false
}

type mapSet StateControllerBase

const (
	mapSet_mapArray byte = iota
	mapSet_value
	mapSet_redirectid
	mapSet_type
)

func (sc mapSet) Run(c *Char, _ []int32) bool {
	crun := c
	var s string
	var value float32
	var scType int32
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case mapSet_mapArray:
			s = string(*(*[]byte)(unsafe.Pointer(&exp[0])))
		case mapSet_value:
			value = exp[0].evalF(c)
		case mapSet_type:
			scType = exp[0].evalI(c)
		case mapSet_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	crun.mapSet(s, value, scType)
	return false
}

type matchRestart StateControllerBase

const (
	matchRestart_reload byte = iota
	matchRestart_stagedef
	matchRestart_p1def
	matchRestart_p2def
	matchRestart_p3def
	matchRestart_p4def
	matchRestart_p5def
	matchRestart_p6def
	matchRestart_p7def
	matchRestart_p8def
)

func (sc matchRestart) Run(c *Char, _ []int32) bool {
	var s string
	reloadFlag := false
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case matchRestart_reload:
			for i, p := range exp {
				sys.reloadCharSlot[i] = p.evalB(c)
				if sys.reloadCharSlot[i] {
					reloadFlag = true
				}
			}
		case matchRestart_stagedef:
			s = string(*(*[]byte)(unsafe.Pointer(&exp[0])))
			if filepath.IsAbs(s) {
				sys.sel.sdefOverwrite = s
			} else {
				sys.sel.sdefOverwrite = filepath.Dir(c.gi().def) + "/" + s
			}
			//sys.reloadStageFlg = true
			reloadFlag = true
		case matchRestart_p1def:
			s = string(*(*[]byte)(unsafe.Pointer(&exp[0])))
			if filepath.IsAbs(s) {
				sys.sel.cdefOverwrite[0] = s
			} else {
				sys.sel.cdefOverwrite[0] = filepath.Dir(c.gi().def) + "/" + s
			}
		case matchRestart_p2def:
			s = string(*(*[]byte)(unsafe.Pointer(&exp[0])))
			if filepath.IsAbs(s) {
				sys.sel.cdefOverwrite[1] = s
			} else {
				sys.sel.cdefOverwrite[1] = filepath.Dir(c.gi().def) + "/" + s
			}
		case matchRestart_p3def:
			s = string(*(*[]byte)(unsafe.Pointer(&exp[0])))
			if filepath.IsAbs(s) {
				sys.sel.cdefOverwrite[2] = s
			} else {
				sys.sel.cdefOverwrite[2] = filepath.Dir(c.gi().def) + "/" + s
			}
		case matchRestart_p4def:
			s = string(*(*[]byte)(unsafe.Pointer(&exp[0])))
			if filepath.IsAbs(s) {
				sys.sel.cdefOverwrite[3] = s
			} else {
				sys.sel.cdefOverwrite[3] = filepath.Dir(c.gi().def) + "/" + s
			}
		case matchRestart_p5def:
			s = string(*(*[]byte)(unsafe.Pointer(&exp[0])))
			if filepath.IsAbs(s) {
				sys.sel.cdefOverwrite[4] = s
			} else {
				sys.sel.cdefOverwrite[4] = filepath.Dir(c.gi().def) + "/" + s
			}
		case matchRestart_p6def:
			s = string(*(*[]byte)(unsafe.Pointer(&exp[0])))
			if filepath.IsAbs(s) {
				sys.sel.cdefOverwrite[5] = s
			} else {
				sys.sel.cdefOverwrite[5] = filepath.Dir(c.gi().def) + "/" + s
			}
		case matchRestart_p7def:
			s = string(*(*[]byte)(unsafe.Pointer(&exp[0])))
			if filepath.IsAbs(s) {
				sys.sel.cdefOverwrite[6] = s
			} else {
				sys.sel.cdefOverwrite[6] = filepath.Dir(c.gi().def) + "/" + s
			}
		case matchRestart_p8def:
			s = string(*(*[]byte)(unsafe.Pointer(&exp[0])))
			if filepath.IsAbs(s) {
				sys.sel.cdefOverwrite[7] = s
			} else {
				sys.sel.cdefOverwrite[7] = filepath.Dir(c.gi().def) + "/" + s
			}
		}
		return true
	})
	if sys.netInput == nil && sys.fileInput == nil {
		if reloadFlag {
			sys.reloadFlg = true
		} else {
			sys.roundResetFlg = true
		}
	}
	return false
}

type printToConsole StateControllerBase

const (
	printToConsole_params byte = iota
	printToConsole_text
)

func (sc printToConsole) Run(c *Char, _ []int32) bool {
	params := []interface{}{}
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case printToConsole_params:
			for _, e := range exp {
				if bv := e.run(c); bv.t == VT_Float {
					params = append(params, bv.ToF())
				} else {
					params = append(params, bv.ToI())
				}
			}
		case printToConsole_text:
			sys.printToConsole(sys.workingState.playerNo,
				int(exp[0].evalI(c)), params...)
		}
		return true
	})
	return false
}

type rankAdd StateControllerBase

const (
	rankAdd_value byte = iota
	rankAdd_max
	rankAdd_type
	rankAdd_icon
	rankAdd_redirectid
)

func (sc rankAdd) Run(c *Char, _ []int32) bool {
	crun := c
	var val, max float32
	var typ, ico string
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case rankAdd_icon:
			ico = string(*(*[]byte)(unsafe.Pointer(&exp[0])))
		case rankAdd_type:
			typ = string(*(*[]byte)(unsafe.Pointer(&exp[0])))
		case rankAdd_max:
			max = exp[0].evalF(c)
		case rankAdd_value:
			val = exp[0].evalF(c)
		case rankAdd_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	crun.rankAdd(val, max, typ, ico)
	return false
}

type redLifeAdd StateControllerBase

const (
	redLifeAdd_absolute byte = iota
	redLifeAdd_value
	redLifeAdd_redirectid
)

func (sc redLifeAdd) Run(c *Char, _ []int32) bool {
	a := false
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case redLifeAdd_absolute:
			a = exp[0].evalB(c)
		case redLifeAdd_value:
			crun.redLifeAdd(float64(exp[0].evalI(c)), a)
		case redLifeAdd_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type redLifeSet StateControllerBase

const (
	redLifeSet_value byte = iota
	redLifeSet_redirectid
)

func (sc redLifeSet) Run(c *Char, _ []int32) bool {
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case redLifeSet_value:
			crun.redLifeSet(exp[0].evalI(c))
		case redLifeSet_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type remapSprite StateControllerBase

const (
	remapSprite_reset byte = iota
	remapSprite_preset
	remapSprite_source
	remapSprite_dest
	remapSprite_redirectid
)

func (sc remapSprite) Run(c *Char, _ []int32) bool {
	crun := c
	src := [...]int16{-1, -1}
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case remapSprite_reset:
			if exp[0].evalB(c) {
				crun.remapSpr = make(RemapPreset)
			}
		case remapSprite_preset:
			crun.remapSpritePreset(string(*(*[]byte)(unsafe.Pointer(&exp[0]))))
		case remapSprite_source:
			src[0] = int16(exp[0].evalI(c))
			if len(exp) > 1 {
				src[1] = int16(exp[1].evalI(c))
			}
		case remapSprite_dest:
			dst := [...]int16{int16(exp[0].evalI(c)), -1}
			if len(exp) > 1 {
				dst[1] = int16(exp[1].evalI(c))
			}
			crun.remapSprite(src, dst)
		case remapSprite_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	crun.anim.remap = crun.remapSpr
	return false
}

type roundTimeAdd StateControllerBase

const (
	roundTimeAdd_value byte = iota
	roundTimeAdd_redirectid
)

func (sc roundTimeAdd) Run(c *Char, _ []int32) bool {
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case roundTimeAdd_value:
			sys.time = Min(sys.roundTime, sys.time+exp[0].evalI(c))
		}
		return true
	})
	return false
}

type roundTimeSet StateControllerBase

const (
	roundTimeSet_value byte = iota
	roundTimeSet_redirectid
)

func (sc roundTimeSet) Run(c *Char, _ []int32) bool {
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case roundTimeSet_value:
			sys.time = Min(sys.roundTime, exp[0].evalI(c))
		}
		return true
	})
	return false
}

type saveFile StateControllerBase

const (
	saveFile_path byte = iota
	saveFile_saveData
	saveFile_redirectid
)

func (sc saveFile) Run(c *Char, _ []int32) bool {
	crun := c
	var path string
	var data SaveData
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case saveFile_path:
			path = string(*(*[]byte)(unsafe.Pointer(&exp[0])))
		case saveFile_saveData:
			data = SaveData(exp[0].evalI(c))
		case saveFile_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	if path != "" {
		encodeFile, err := os.Create(filepath.Dir(c.gi().def) + "/" + path)
		if err != nil {
			panic(err)
		}
		defer encodeFile.Close()
		encoder := gob.NewEncoder(encodeFile)
		switch data {
		case SaveData_map:
			if err := encoder.Encode(crun.mapArray); err != nil {
				panic(err)
			}
		case SaveData_var:
			if err := encoder.Encode(crun.ivar); err != nil {
				panic(err)
			}
		case SaveData_fvar:
			if err := encoder.Encode(crun.fvar); err != nil {
				panic(err)
			}
		}
	}
	return false
}

type scoreAdd StateControllerBase

const (
	scoreAdd_value byte = iota
	scoreAdd_redirectid
)

func (sc scoreAdd) Run(c *Char, _ []int32) bool {
	crun := c
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case scoreAdd_value:
			crun.scoreAdd(exp[0].evalF(c))
		case scoreAdd_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type targetDizzyPointsAdd StateControllerBase

const (
	targetDizzyPointsAdd_id byte = iota
	targetDizzyPointsAdd_value
	targetDizzyPointsAdd_redirectid
)

func (sc targetDizzyPointsAdd) Run(c *Char, _ []int32) bool {
	crun := c
	tar := crun.getTarget(-1)
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case targetDizzyPointsAdd_id:
			if len(tar) == 0 {
				return false
			}
			tar = crun.getTarget(exp[0].evalI(c))
		case targetDizzyPointsAdd_value:
			if len(tar) == 0 {
				return false
			}
			crun.targetDizzyPointsAdd(tar, exp[0].evalI(c))
		case targetDizzyPointsAdd_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
				tar = crun.getTarget(-1)
				if len(tar) == 0 {
					return false
				}
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type targetGuardPointsAdd StateControllerBase

const (
	targetGuardPointsAdd_id byte = iota
	targetGuardPointsAdd_value
	targetGuardPointsAdd_redirectid
)

func (sc targetGuardPointsAdd) Run(c *Char, _ []int32) bool {
	crun := c
	tar := crun.getTarget(-1)
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case targetGuardPointsAdd_id:
			if len(tar) == 0 {
				return false
			}
			tar = crun.getTarget(exp[0].evalI(c))
		case targetGuardPointsAdd_value:
			if len(tar) == 0 {
				return false
			}
			crun.targetGuardPointsAdd(tar, exp[0].evalI(c))
		case targetGuardPointsAdd_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
				tar = crun.getTarget(-1)
				if len(tar) == 0 {
					return false
				}
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type targetRedLifeAdd StateControllerBase

const (
	targetRedLifeAdd_id byte = iota
	targetRedLifeAdd_absolute
	targetRedLifeAdd_value
	targetRedLifeAdd_redirectid
)

func (sc targetRedLifeAdd) Run(c *Char, _ []int32) bool {
	crun := c
	tar, a := crun.getTarget(-1), false
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case targetRedLifeAdd_id:
			if len(tar) == 0 {
				return false
			}
			tar = crun.getTarget(exp[0].evalI(c))
		case targetRedLifeAdd_absolute:
			a = exp[0].evalB(c)
		case targetRedLifeAdd_value:
			if len(tar) == 0 {
				return false
			}
			crun.targetRedLifeAdd(tar, float64(exp[0].evalI(c)), a)
		case targetRedLifeAdd_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
				tar = crun.getTarget(-1)
				if len(tar) == 0 {
					return false
				}
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type targetScoreAdd StateControllerBase

const (
	targetScoreAdd_id byte = iota
	targetScoreAdd_value
	targetScoreAdd_redirectid
)

func (sc targetScoreAdd) Run(c *Char, _ []int32) bool {
	crun := c
	tar := crun.getTarget(-1)
	StateControllerBase(sc).run(c, func(id byte, exp []BytecodeExp) bool {
		switch id {
		case targetScoreAdd_id:
			if len(tar) == 0 {
				return false
			}
			tar = crun.getTarget(exp[0].evalI(c))
		case targetScoreAdd_value:
			if len(tar) == 0 {
				return false
			}
			crun.targetScoreAdd(tar, exp[0].evalF(c))
		case targetScoreAdd_redirectid:
			if rid := sys.playerID(exp[0].evalI(c)); rid != nil {
				crun = rid
				tar = crun.getTarget(-1)
				if len(tar) == 0 {
					return false
				}
			} else {
				return false
			}
		}
		return true
	})
	return false
}

type StateBytecode struct {
	stateType StateType
	moveType  MoveType
	physics   StateType
	playerNo  int
	stateDef  stateDef
	block     StateBlock
	ctrlsps   []int32
	numVars   int32
}

func newStateBytecode(pn int) *StateBytecode {
	sb := &StateBytecode{stateType: ST_S, moveType: MT_I, physics: ST_N,
		playerNo: pn, block: *newStateBlock()}
	return sb
}
func (sb *StateBytecode) init(c *Char) {
	if sb.stateType != ST_U {
		c.ss.stateType = sb.stateType
	}
	if sb.moveType != MT_U {
		c.ss.moveType = sb.moveType
	}
	if sb.physics != ST_U {
		c.ss.physics = sb.physics
	}
	sb.ctrlsps = make([]int32, len(sb.ctrlsps))
	sys.workingState = sb
	sb.stateDef.Run(c)
}
func (sb *StateBytecode) run(c *Char) (changeState bool) {
	sys.bcVar = sys.bcVarStack.Alloc(int(sb.numVars))
	sys.workingState = sb
	changeState = sb.block.Run(c, sb.ctrlsps)
	if len(sys.bcStack) != 0 {
		sys.errLog.Println(sys.cgi[sb.playerNo].def)
		for _, v := range sys.bcStack {
			sys.errLog.Printf("%+v\n", v)
		}
		c.panic()
	}
	sys.bcVarStack.Clear()
	return
}
