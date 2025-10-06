package main

import (
	"fmt"
	"io"
	"math"
	"path/filepath"
	"sort"
	"strings"
)

const MaxPalNo = 12
const MaxQuotes = 100

type SystemCharFlag uint32

const (
	SCF_ctrl SystemCharFlag = 1 << iota
	SCF_disabled
	SCF_dizzy
	SCF_guard
	SCF_guardbreak
	SCF_ko
	SCF_over_alive // Has reached win or lose poses
	SCF_over_ko    // Has reached state 5150
	SCF_standby
)

// These flags are reset manually
type CharSpecialFlag uint32

const (
	CSF_angledraw CharSpecialFlag = 1 << iota
	CSF_depth
	CSF_depthedge
	CSF_destroy
	CSF_gethit
	CSF_height
	CSF_movecamera_x
	CSF_movecamera_y
	CSF_playerpush
	CSF_posfreeze
	CSF_screenbound
	CSF_stagebound
	CSF_width
	CSF_widthedge
)

// Flags set by AssertSpecial. They are reset together every frame
type AssertSpecialFlag uint64

const (
	// Mugen flags
	ASF_invisible AssertSpecialFlag = 1 << iota
	ASF_noairguard
	ASF_noautoturn
	ASF_nocrouchguard
	ASF_nojugglecheck
	ASF_noko
	ASF_noshadow
	ASF_nostandguard
	ASF_nowalk
	ASF_unguardable
	// Ikemen flags
	ASF_animatehitpause
	ASF_animfreeze
	ASF_autoguard
	ASF_drawunder
	ASF_noaibuttonjam
	ASF_noaicheat
	ASF_noailevel
	ASF_noairjump
	ASF_nobrake
	ASF_nocombodisplay
	ASF_nocrouch
	ASF_nodizzypointsdamage
	ASF_nofacedisplay
	ASF_nofacep2
	ASF_nofallcount
	ASF_nofalldefenceup
	ASF_nofallhitflag
	ASF_nofastrecoverfromliedown
	ASF_nogetupfromliedown
	ASF_noguardbardisplay
	ASF_noguarddamage
	ASF_noguardko
	ASF_noguardpointsdamage
	ASF_nohardcodedkeys
	ASF_nohitdamage
	ASF_noinput
	ASF_nointroreset
	ASF_nojump
	ASF_nokofall // In Mugen this seems hardcoded into Training mode
	ASF_nokovelocity
	ASF_nolifebaraction
	ASF_nolifebardisplay
	ASF_nomakedust
	ASF_nonamedisplay
	ASF_nopowerbardisplay
	ASF_noredlifedamage
	ASF_nostand
	ASF_nostunbardisplay
	ASF_noturntarget
	ASF_nowinicondisplay
	ASF_postroundinput
	ASF_projtypecollision // TODO: Make this a parameter for normal projectiles as well?
	ASF_runfirst
	ASF_runlast
	ASF_sizepushonly
)

type GlobalSpecialFlag uint32

const (
	// Mugen flags
	GSF_globalnoko GlobalSpecialFlag = 1 << iota
	GSF_globalnoshadow
	GSF_intro
	GSF_nobardisplay
	GSF_nobg
	GSF_nofg
	GSF_nokoslow
	GSF_nokosnd
	GSF_nomusic
	GSF_roundnotover
	GSF_timerfreeze
	// Ikemen flags
	GSF_camerafreeze
	GSF_roundfreeze
	GSF_roundnotskip
	GSF_skipfightdisplay
	GSF_skipkodisplay
	GSF_skiprounddisplay
	GSF_skipwindisplay
)

type PosType int32

const (
	PT_P1 PosType = iota
	PT_P2
	PT_Front
	PT_Back
	PT_Left
	PT_Right
	PT_None
)

type Space int32

const (
	Space_none Space = iota
	Space_stage
	Space_screen
)

type Projection int32

const (
	Projection_Orthographic Projection = iota
	Projection_Perspective
	Projection_Perspective2
)

type SaveData int32

const (
	SaveData_map SaveData = iota
	SaveData_var
	SaveData_fvar
)

type ClsnText struct {
	x, y    float32
	text    string
	r, g, b int32
}

type ClsnRect [][7]float32

func (cr *ClsnRect) Add(clsn [][4]float32, x, y, xs, ys, angle float32) {
	x = (x - sys.cam.Pos[0]) * sys.cam.Scale
	y = (y*sys.cam.Scale - sys.cam.Pos[1]) + sys.cam.GroundLevel()
	xs *= sys.cam.Scale
	ys *= sys.cam.Scale
	sw := float32(sys.gameWidth)
	sh := float32(0) //float32(sys.gameHeight)
	for i := 0; i < len(clsn); i++ {
		offx := sw / 2
		offy := sh
		rect := [...]float32{
			AbsF(xs) * clsn[i][0], AbsF(ys) * clsn[i][1],
			xs * (clsn[i][2] - clsn[i][0]), ys * (clsn[i][3] - clsn[i][1]),
			(x + offx) * sys.widthScale, (y + offy) * sys.heightScale, angle}
		*cr = append(*cr, rect)
	}
}

func (cr ClsnRect) draw(blendAlpha [2]int32) {
	paltex := PaletteToTexture(sys.clsnSpr.Pal)
	for _, c := range cr {
		params := RenderParams{
			tex:            sys.clsnSpr.Tex,
			paltex:         paltex,
			size:           sys.clsnSpr.Size,
			x:              -c[0] * sys.widthScale,
			y:              -c[1] * sys.heightScale,
			tile:           notiling,
			xts:            c[2] * sys.widthScale,
			xbs:            c[2] * sys.widthScale,
			ys:             c[3] * sys.heightScale,
			vs:             1,
			rxadd:          0,
			xas:            1,
			yas:            1,
			rot:            Rotation{angle: c[6]},
			tint:           0,
			blendMode:      TT_add,
			blendAlpha:     blendAlpha,
			mask:           -1,
			pfx:            nil,
			window:         &sys.scrrect,
			rcx:            c[4],
			rcy:            c[5],
			projectionMode: 0,
			fLength:        0,
			xOffset:        0,
			yOffset:        0,
		}
		RenderSprite(params)
	}
}

type CharData struct {
	life        int32
	power       int32
	dizzypoints int32
	guardpoints int32
	attack      int32
	defence     int32
	fall        struct {
		defence_up  int32
		defence_mul float32
	}
	liedown struct {
		time int32
	}
	airjuggle int32
	sparkno   int32
	guard     struct {
		sparkno int32
	}
	hitsound_channel   int32
	guardsound_channel int32
	ko                 struct {
		echo int32
	}
	volume            int32
	intpersistindex   int32
	floatpersistindex int32
}

func (cd *CharData) init() {
	*cd = CharData{}
	cd.life = 1000
	cd.power = 3000
	cd.dizzypoints = 1000
	cd.guardpoints = 1000
	cd.attack = 100
	cd.defence = 100
	cd.fall.defence_up = 50
	cd.fall.defence_mul = 1.5
	cd.liedown.time = 60
	cd.airjuggle = 15
	cd.sparkno = 2
	cd.guard.sparkno = 40
	cd.hitsound_channel = -1
	cd.guardsound_channel = -1
	cd.ko.echo = 0
	cd.volume = 256
	cd.intpersistindex = int32(math.MaxInt32)
	cd.floatpersistindex = int32(math.MaxInt32)
}

type CharSize struct {
	xscale    float32
	yscale    float32
	standbox  [4]float32 // Replaces ground.front, ground.back and height
	crouchbox [4]float32
	airbox    [4]float32 // Replaces air.front and air.back
	downbox   [4]float32
	attack    struct {
		dist struct {
			width  [2]float32
			height [2]float32
			depth  [2]float32
		}
		depth [2]float32
	}
	proj struct {
		attack struct {
			dist struct {
				width  [2]float32
				height [2]float32
				depth  [2]float32
			}
		}
		doscale int32
	}
	head struct {
		pos [2]float32
	}
	mid struct {
		pos [2]float32
	}
	shadowoffset float32
	draw         struct {
		offset [2]float32
	}
	depth      [2]float32 // Called z.width in Mugen 2000.01.01
	weight     int32
	pushfactor float32
}

func (cs *CharSize) init() {
	*cs = CharSize{}
	cs.xscale = 1
	cs.yscale = 1
	cs.standbox = [4]float32{-16, -60, 16, 0}
	cs.crouchbox = [4]float32{-16, -60, 16, 0}
	cs.airbox = [4]float32{-12, -60, 12, 0}
	cs.downbox = [4]float32{-16, -60, 16, 0}
	cs.attack.dist.width = [...]float32{160, 0}
	cs.attack.dist.height = [...]float32{1000, 1000}
	cs.attack.dist.depth = [...]float32{4, 4}
	cs.proj.attack.dist.width = [...]float32{90, 0}
	cs.proj.attack.dist.height = [...]float32{1000, 1000}
	cs.proj.attack.dist.depth = [...]float32{10, 10}
	cs.proj.doscale = 0
	cs.head.pos = [...]float32{-5, -90}
	cs.mid.pos = [...]float32{-5, -60}
	cs.shadowoffset = 0
	cs.draw.offset = [...]float32{0, 0}
	cs.depth = [...]float32{3, 3}
	cs.attack.depth = [...]float32{4, 4}
	cs.weight = 100
	cs.pushfactor = 1
}

type CharVelocity struct {
	walk struct {
		fwd  float32
		back float32
		up   [3]float32
		down [3]float32
	}
	run struct {
		fwd  [2]float32
		back [2]float32
		up   [3]float32
		down [3]float32
	}
	jump struct {
		neu  [2]float32
		back float32
		fwd  float32
		up   [3]float32
		down [3]float32
	}
	runjump struct {
		back [2]float32
		fwd  [2]float32
		up   [3]float32
		down [3]float32
	}
	airjump struct {
		neu  [2]float32
		back float32
		fwd  float32
		up   [3]float32
		down [3]float32
	}
	air struct {
		gethit struct {
			groundrecover [2]float32
			airrecover    struct {
				mul  [2]float32
				add  [2]float32
				back float32
				fwd  float32
				up   float32
				down float32
			}
			ko struct {
				add  [2]float32
				ymin float32
			}
		}
	}
	ground struct {
		gethit struct {
			ko struct {
				xmul float32
				add  [2]float32
				ymin float32
			}
		}
	}
}

func (cv *CharVelocity) init() {
	*cv = CharVelocity{}
	cv.air.gethit.groundrecover = [...]float32{-0.15, -3.5}
	cv.air.gethit.airrecover.mul = [...]float32{0.5, 0.2}
	cv.air.gethit.airrecover.add = [...]float32{0.0, -4.5}
	cv.air.gethit.airrecover.back = -1.0
	cv.air.gethit.airrecover.fwd = 0.0
	cv.air.gethit.airrecover.up = -2.0
	cv.air.gethit.airrecover.down = 1.5
	cv.air.gethit.ko.add = [...]float32{-2.5, -2}
	cv.air.gethit.ko.ymin = -3
	cv.ground.gethit.ko.xmul = 0.66
	cv.ground.gethit.ko.add = [...]float32{-2.5, -2}
	cv.ground.gethit.ko.ymin = -6
}

type CharMovement struct {
	airjump struct {
		num    int32
		height float32
	}
	yaccel float32
	stand  struct {
		friction           float32
		friction_threshold float32
	}
	crouch struct {
		friction           float32
		friction_threshold float32
	}
	air struct {
		gethit struct {
			groundlevel   float32
			groundrecover struct {
				ground struct {
					threshold float32
				}
				groundlevel float32
			}
			airrecover struct {
				threshold float32
				yaccel    float32
			}
			trip struct {
				groundlevel float32
			}
		}
	}
	down struct {
		bounce struct {
			offset      [2]float32
			yaccel      float32
			groundlevel float32
		}
		gethit struct {
			offset [2]float32
		}
		friction_threshold float32
	}
}

func (cm *CharMovement) init() {
	*cm = CharMovement{}
	cm.yaccel = 0.44
	cm.stand.friction = 0.85
	cm.stand.friction_threshold = 2.0
	cm.crouch.friction = 0.82
	cm.crouch.friction_threshold = 0.0
	cm.air.gethit.groundlevel = 10.0
	cm.air.gethit.groundrecover.ground.threshold = -20.0
	cm.air.gethit.groundrecover.groundlevel = 10.0
	cm.air.gethit.airrecover.threshold = -1.0
	cm.air.gethit.airrecover.yaccel = 0.35
	cm.air.gethit.trip.groundlevel = 15.0
	cm.down.bounce.offset = [...]float32{0, 20}
	cm.down.bounce.yaccel = 0.4
	cm.down.bounce.groundlevel = 12.0
	cm.down.gethit.offset = [...]float32{0, 15}
	cm.down.friction_threshold = 0.05
}

type Reaction int32

const (
	RA_Light   Reaction = 0
	RA_Medium  Reaction = 1
	RA_Hard    Reaction = 2
	RA_Back    Reaction = 3
	RA_Up      Reaction = 4
	RA_Diagup  Reaction = 5
	RA_Unknown Reaction = -1
)

type HitType int32

const (
	HT_None    HitType = 0
	HT_High    HitType = 1
	HT_Low     HitType = 2
	HT_Trip    HitType = 3
	HT_Unknown HitType = -1
)

type TradeType int32

const (
	TT_Hit TradeType = iota
	TT_Miss
	TT_Dodge
)

type HitDef struct {
	isprojectile               bool // Projectile state controller
	attr                       int32
	reversal_attr              int32
	hitflag                    int32
	guardflag                  int32
	reversal_guardflag         int32
	reversal_guardflag_not     int32
	affectteam                 int32 // -1F, 0B, 1E
	teamside                   int
	animtype                   Reaction
	air_animtype               Reaction
	priority                   int32
	prioritytype               TradeType
	hitdamage                  int32
	guarddamage                int32
	pausetime                  [2]int32
	guard_pausetime            [2]int32
	sparkno                    int32
	sparkno_ffx                string
	sparkangle                 float32
	sparkscale                 [2]float32
	guard_sparkno              int32
	guard_sparkno_ffx          string
	guard_sparkangle           float32
	guard_sparkscale           [2]float32
	sparkxy                    [2]float32
	hitsound                   [2]int32
	hitsound_channel           int32
	hitsound_ffx               string
	guardsound                 [2]int32
	guardsound_channel         int32
	guardsound_ffx             string
	ground_type                HitType
	air_type                   HitType
	ground_slidetime           int32
	guard_slidetime            int32
	ground_hittime             int32
	guard_hittime              int32
	air_hittime                int32
	guard_ctrltime             int32
	airguard_ctrltime          int32
	guard_dist_x               [2]float32
	guard_dist_y               [2]float32
	guard_dist_z               [2]float32
	xaccel                     float32
	yaccel                     float32
	zaccel                     float32
	ground_velocity            [3]float32
	guard_velocity             [3]float32
	air_velocity               [3]float32
	airguard_velocity          [3]float32
	ground_cornerpush_veloff   float32
	air_cornerpush_veloff      float32
	down_cornerpush_veloff     float32
	guard_cornerpush_veloff    float32
	airguard_cornerpush_veloff float32
	air_juggle                 int32
	p1sprpriority              int32
	p2sprpriority              int32
	p1getp2facing              int32
	p1facing                   int32
	p2facing                   int32
	p1stateno                  int32
	p2stateno                  int32
	p2getp1state               bool
	missonoverride             int32
	forcestand                 int32
	forcecrouch                int32
	ground_fall                bool
	air_fall                   int32 // Technically a bool but it needs an undefined state for "ifierrset"
	down_velocity              [3]float32
	down_hittime               int32
	down_bounce                bool
	down_recover               bool
	down_recovertime           int32
	id                         int32
	chainid                    int32
	nochainid                  [8]int32
	hitonce                    int32
	numhits                    int32
	hitgetpower                int32
	guardgetpower              int32
	hitgivepower               int32
	guardgivepower             int32
	palfx                      PalFXDef
	envshake_time              int32
	envshake_freq              float32
	envshake_ampl              int32
	envshake_phase             float32
	envshake_mul               float32
	envshake_dir               float32
	mindist                    [3]float32
	maxdist                    [3]float32
	snap                       [3]float32
	snaptime                   int32
	fall_animtype              Reaction // Old fall struct
	fall_xvelocity             float32
	fall_yvelocity             float32
	fall_zvelocity             float32
	fall_recover               bool
	fall_recovertime           int32
	fall_damage                int32
	fall_kill                  bool
	fall_envshake_time         int32
	fall_envshake_freq         float32
	fall_envshake_ampl         int32
	fall_envshake_phase        float32
	fall_envshake_mul          float32
	fall_envshake_dir          float32
	playerNo                   int
	kill                       bool
	guard_kill                 bool
	forcenofall                bool
	ltypehit                   bool
	attackerID                 int32
	dizzypoints                int32
	guardpoints                int32
	hitredlife                 int32
	guardredlife               int32
	score                      [2]float32
	p2clsncheck                int32
	p2clsnrequire              int32
	attack_depth               [2]float32
	unhittabletime             [2]int32
}

func (hd *HitDef) clear(c *Char, localscl float32) {
	var originLs float32
	if c.gi().constants["default.legacyfallyvelyaccel"] == 1 {
		originLs = 1
	} else {
		// Convert local scale back to 4:3 in order to keep values consistent in widescreen
		originLs = c.localscl * (320 / float32(sys.gameWidth))
	}

	*hd = HitDef{
		isprojectile:       false,
		playerNo:           -1,
		hitflag:            int32(HF_H | HF_L | HF_A | HF_F),
		guardflag:          0,
		affectteam:         1,
		teamside:           -1,
		animtype:           RA_Light,
		air_animtype:       RA_Unknown,
		priority:           4,
		prioritytype:       TT_Hit,
		sparkno:            c.gi().data.sparkno,
		sparkno_ffx:        "f",
		sparkangle:         0,
		sparkscale:         [2]float32{1, 1},
		guard_sparkno:      c.gi().data.guard.sparkno,
		guard_sparkno_ffx:  "f",
		guard_sparkangle:   0,
		guard_sparkscale:   [2]float32{1, 1},
		hitsound:           [2]int32{-1, 0},
		hitsound_channel:   c.gi().data.hitsound_channel,
		hitsound_ffx:       "f",
		guardsound:         [2]int32{-1, 0},
		guardsound_channel: c.gi().data.guardsound_channel,
		guardsound_ffx:     "f",
		ground_type:        HT_High,
		air_type:           HT_Unknown,
		air_hittime:        20,
		down_hittime:       20, // Not documented in Mugen docs

		guard_pausetime:   [2]int32{IErr, IErr},
		guard_hittime:     IErr,
		guard_slidetime:   IErr,
		guard_ctrltime:    IErr,
		airguard_ctrltime: IErr,

		ground_velocity:            [3]float32{0, 0, 0},
		air_velocity:               [3]float32{0, 0, 0},
		down_velocity:              [3]float32{float32(math.NaN()), float32(math.NaN()), float32(math.NaN())},
		guard_velocity:             [3]float32{float32(math.NaN()), 0, float32(math.NaN())}, // We don't want chars to be launched in Y while guarding
		airguard_velocity:          [3]float32{float32(math.NaN()), float32(math.NaN()), float32(math.NaN())},
		ground_cornerpush_veloff:   float32(math.NaN()),
		air_cornerpush_veloff:      float32(math.NaN()),
		down_cornerpush_veloff:     float32(math.NaN()),
		guard_cornerpush_veloff:    float32(math.NaN()),
		airguard_cornerpush_veloff: float32(math.NaN()),

		xaccel: 0,
		yaccel: 0.35 / originLs,
		zaccel: 0,

		p1sprpriority:       1,
		p1stateno:           -1,
		p2stateno:           -1,
		missonoverride:      -1,
		forcestand:          IErr,
		forcecrouch:         IErr,
		air_fall:            IErr,
		guard_dist_x:        hd.guard_dist_x, // These default to no change
		guard_dist_y:        hd.guard_dist_y, // They are reset when hitdefpersist = 0
		guard_dist_z:        hd.guard_dist_z,
		chainid:             -1,
		nochainid:           [8]int32{-1, -1, -1, -1, -1, -1, -1, -1},
		numhits:             1,
		hitgetpower:         IErr,
		guardgetpower:       IErr,
		hitgivepower:        IErr,
		guardgivepower:      IErr,
		envshake_freq:       60,
		envshake_ampl:       -4,
		envshake_phase:      float32(math.NaN()),
		envshake_mul:        1.0,
		envshake_dir:        0.0,
		mindist:             [...]float32{float32(math.NaN()), float32(math.NaN()), float32(math.NaN())},
		maxdist:             [...]float32{float32(math.NaN()), float32(math.NaN()), float32(math.NaN())},
		snap:                [...]float32{float32(math.NaN()), float32(math.NaN()), float32(math.NaN())},
		hitonce:             -1,
		kill:                true,
		guard_kill:          true,
		dizzypoints:         IErr,
		guardpoints:         IErr,
		hitredlife:          IErr,
		guardredlife:        IErr,
		score:               [...]float32{float32(math.NaN()), float32(math.NaN())},
		p2clsncheck:         -1,
		p2clsnrequire:       -1,
		down_recover:        true,
		down_recovertime:    -1,
		air_juggle:          IErr,
		fall_animtype:       RA_Unknown,
		fall_xvelocity:      float32(math.NaN()),
		fall_yvelocity:      -4.5 / originLs,
		fall_zvelocity:      0, // Should this work like the X component instead?
		fall_recover:        true,
		fall_recovertime:    4,
		fall_kill:           true,
		fall_envshake_freq:  60,
		fall_envshake_ampl:  IErr,
		fall_envshake_phase: float32(math.NaN()),
		fall_envshake_mul:   1.0,
		fall_envshake_dir:   0.0,
		attack_depth:        [2]float32{c.size.attack.depth[0], c.size.attack.depth[1]},
		unhittabletime:      [2]int32{IErr, IErr},

		reversal_guardflag:     IErr,
		reversal_guardflag_not: IErr,
	}

	// PalFX
	hd.palfx.mul = [3]int32{255, 255, 255}
	hd.palfx.color = 1
	hd.palfx.hue = 0
}

// When a Hitdef connects, its statetype attribute will be updated to the character's current type
// Even if the Hitdef has multiple statetype attributes
// TODO: This is an oddly specific Mugen thing that might not be needed in future Ikemen characters
func (hd *HitDef) updateStateType(stateType StateType) {
	hd.attr = hd.attr&^int32(ST_MASK) | int32(stateType) | -1<<31
	hd.reversal_attr |= -1 << 31
	hd.ltypehit = false
}

func (hd *HitDef) testAttr(attr int32) bool {
	attr &= hd.attr
	return attr&int32(ST_MASK) != 0 && attr&^int32(ST_MASK)&^(-1<<31) != 0
}

func (hd *HitDef) testReversalAttr(attr int32) bool {
	attr &= hd.reversal_attr
	return attr&int32(ST_MASK) != 0 && attr&^int32(ST_MASK)&^(-1<<31) != 0
}

type GetHitVar struct {
	targetedBy          [][2]int32 // ID, current juggle
	attr                int32
	_type               HitType
	animtype            Reaction
	airanimtype         Reaction
	groundanimtype      Reaction
	airtype             HitType
	groundtype          HitType
	damage              int32
	hitcount            int32
	guardcount          int32
	fallcount           int32
	hitshaketime        int32
	hittime             int32
	slidetime           int32
	ctrltime            int32
	xvel                float32
	yvel                float32
	zvel                float32
	xaccel              float32
	yaccel              float32
	zaccel              float32
	xveladd             float32
	yveladd             float32
	hitid               int32
	xoff                float32
	yoff                float32
	zoff                float32
	fall_animtype       Reaction // Old fall struct
	fall_xvelocity      float32
	fall_yvelocity      float32
	fall_zvelocity      float32
	fall_recover        bool
	fall_recovertime    int32
	fall_damage         int32
	fall_kill           bool
	fall_envshake_time  int32
	fall_envshake_freq  float32
	fall_envshake_ampl  int32
	fall_envshake_phase float32
	fall_envshake_mul   float32
	fall_envshake_dir   float32
	playerId            int32
	playerNo            int
	fallflag            bool
	guarded             bool
	p2getp1state        bool
	forcestand          bool
	forcecrouch         bool
	dizzypoints         int32
	guardpoints         int32
	redlife             int32
	score               float32
	hitdamage           int32
	guarddamage         int32
	power               int32
	hitpower            int32
	guardpower          int32
	hitredlife          int32
	guardredlife        int32
	kill                bool
	priority            int32
	facing              int32
	ground_velocity     [3]float32
	air_velocity        [3]float32
	down_velocity       [3]float32
	guard_velocity      [3]float32
	airguard_velocity   [3]float32
	frame               bool
	cheeseKO            bool
	down_recover        bool
	down_recovertime    int32
	guardflag           int32
}

func (ghv *GetHitVar) clear(c *Char) {
	var originLs float32
	if c.gi().constants["default.legacyfallyvelyaccel"] == 1 {
		originLs = 1
	} else {
		// Convert local scale back to 4:3 in order to keep values consistent in widescreen
		originLs = c.localscl * (320 / float32(sys.gameWidth))
	}

	*ghv = GetHitVar{
		hittime:        -1,
		yaccel:         0.35 / originLs,
		xoff:           ghv.xoff,
		yoff:           ghv.yoff,
		zoff:           ghv.zoff,
		hitid:          -1,
		playerNo:       -1,
		fall_animtype:  RA_Unknown,
		fall_xvelocity: float32(math.NaN()),
		fall_yvelocity: -4.5 / originLs,
		fall_zvelocity: float32(math.NaN()),
	}
}

// In Mugen, Hitdef and Reversaldef do not clear GetHitVars at all between successive hits
// However, this approach helps ensure that the hit properties from one move do not bleed into other moves
// https://github.com/ikemen-engine/Ikemen-GO/issues/1891
func (ghv *GetHitVar) selectiveClear(c *Char) {
	// Save variables that should persist or stack
	cheeseKO := ghv.cheeseKO
	damage := ghv.damage
	dizzypoints := ghv.dizzypoints
	down_recovertime := ghv.down_recovertime
	fallcount := ghv.fallcount
	fallflag := ghv.fallflag
	frame := ghv.frame
	guardcount := ghv.guardcount
	guarddamage := ghv.guarddamage
	guardpoints := ghv.guardpoints
	guardpower := ghv.guardpower
	targetedBy := ghv.targetedBy
	hitcount := ghv.hitcount
	hitdamage := ghv.hitdamage
	hitpower := ghv.hitpower
	kill := ghv.kill
	power := ghv.power

	ghv.clear(c)

	// Restore variables
	ghv.cheeseKO = cheeseKO
	ghv.damage = damage
	ghv.dizzypoints = dizzypoints
	ghv.down_recovertime = down_recovertime
	ghv.fallcount = fallcount
	ghv.fallflag = fallflag
	ghv.frame = frame
	ghv.guardcount = guardcount
	ghv.guarddamage = guarddamage
	ghv.guardpoints = guardpoints
	ghv.guardpower = guardpower
	ghv.targetedBy = targetedBy
	ghv.hitcount = hitcount
	ghv.hitdamage = hitdamage
	ghv.hitpower = hitpower
	ghv.kill = kill
	ghv.power = power
}

func (ghv *GetHitVar) clearOff() {
	ghv.xoff, ghv.yoff, ghv.zoff = 0, 0, 0
}

func (ghv GetHitVar) chainId() int32 {
	if ghv.hitid > 0 {
		return ghv.hitid
	}
	return 0
}

func (ghv GetHitVar) idMatch(id int32) bool {
	for _, v := range ghv.targetedBy {
		if v[0] == id || v[0] == -id {
			return true
		}
	}
	return false
}

func (ghv GetHitVar) getJuggle(id, defaultJuggle int32) int32 {
	for _, v := range ghv.targetedBy {
		if v[0] == id {
			return v[1]
		}
	}
	return defaultJuggle
}

func (ghv *GetHitVar) dropId(id int32) {
	for i, v := range ghv.targetedBy {
		if v[0] == id {
			ghv.targetedBy = append(ghv.targetedBy[:i], ghv.targetedBy[i+1:]...)
			break
		}
	}
}

func (ghv *GetHitVar) addId(id, juggle int32) {
	juggle = ghv.getJuggle(id, juggle)
	ghv.dropId(id)
	ghv.targetedBy = append(ghv.targetedBy, [...]int32{id, juggle})
}

// Same as testAttr from HitDef
func (ghv *GetHitVar) testAttr(attr int32) bool {
	attr &= ghv.attr
	return (attr&int32(ST_MASK) != 0 && attr&^int32(ST_MASK)&^(-1<<31) != 0)
}

type HitBy struct {
	flag     int32
	time     int32
	not      bool
	playerid int32
	playerno int
	stack    bool
}

func (hb *HitBy) clear() {
	*hb = HitBy{}
}

type HitOverride struct {
	attr          int32
	stateno       int32
	time          int32
	forceair      bool
	forceguard    bool
	guardflag     int32
	guardflag_not int32
	keepState     bool
	playerNo      int
}

func (ho *HitOverride) clear() {
	*ho = HitOverride{stateno: -1, playerNo: -1}
}

type MoveHitVar struct {
	cornerpush float32
	frame      bool
	overridden bool
	playerId   int32
	playerNo   int
	sparkxy    [2]float32
}

func (mhv *MoveHitVar) clear() {
	*mhv = MoveHitVar{}
}

type aimgImage struct {
	anim       *Animation
	pos        [2]float32
	scl        [2]float32
	priority   int32
	rot        Rotation
	projection int32
	fLength    float32
	oldVer     bool
}

type AfterImage struct {
	time           int32
	length         int32
	postbright     [3]int32
	add            [3]int32
	mul            [3]float32
	timegap        int32
	framegap       int32
	trans          TransType
	alpha          [2]int32
	palfx          []*PalFX
	imgs           [64]aimgImage
	imgidx         int32
	restgap        int32
	reccount       int32
	timecount      int32
	priority       int32
	ignorehitpause bool
}

func newAfterImage() *AfterImage {
	ai := &AfterImage{
		palfx: make([]*PalFX, sys.cfg.Config.AfterImageMax),
	}
	for i := range ai.palfx {
		ai.palfx[i] = newPalFX()
		ai.palfx[i].enable = true
		ai.palfx[i].negType = true
	}
	ai.clear()
	return ai
}

func (ai *AfterImage) clear() {
	ai.time = 0
	ai.length = 20
	if len(ai.palfx) > 0 {
		ai.palfx[0].eColor = 1
		ai.palfx[0].eHue = 0
		ai.palfx[0].eInvertall = false
		ai.palfx[0].eInvertblend = 0
		ai.palfx[0].eAdd = [...]int32{30, 30, 30}
		ai.palfx[0].eMul = [...]int32{120, 120, 220}
	}
	ai.postbright = [3]int32{}
	ai.add = [...]int32{10, 10, 25}
	ai.mul = [...]float32{0.65, 0.65, 0.75}
	ai.timegap = 1
	ai.framegap = 4
	ai.trans = TT_default
	ai.alpha = [2]int32{-1, 0}
	ai.imgidx = 0
	ai.restgap = 0
	ai.reccount = 0
	ai.timecount = 0
	ai.ignorehitpause = true
}

func (ai *AfterImage) setPalColor(color int32) {
	if len(ai.palfx) > 0 {
		ai.palfx[0].eColor = float32(Clamp(color, 0, 256)) / 256
	}
}

func (ai *AfterImage) setPalHueShift(huesh int32) {
	if len(ai.palfx) > 0 {
		ai.palfx[0].eHue = (float32(Clamp(huesh, -256, 256)) / 256)
	}
}

func (ai *AfterImage) setPalInvertall(invertall bool) {
	if len(ai.palfx) > 0 {
		ai.palfx[0].eInvertall = invertall
	}
}

func (ai *AfterImage) setPalInvertblend(invertblend int32) {
	if len(ai.palfx) > 0 {
		ai.palfx[0].invertblend = invertblend
	}
}

func (ai *AfterImage) setPalBrightR(addr int32) {
	if len(ai.palfx) > 0 {
		ai.palfx[0].eAdd[0] = addr
	}
}

func (ai *AfterImage) setPalBrightG(addg int32) {
	if len(ai.palfx) > 0 {
		ai.palfx[0].eAdd[1] = addg
	}
}

func (ai *AfterImage) setPalBrightB(addb int32) {
	if len(ai.palfx) > 0 {
		ai.palfx[0].eAdd[2] = addb
	}
}

func (ai *AfterImage) setPalContrastR(mulr int32) {
	if len(ai.palfx) > 0 {
		ai.palfx[0].eMul[0] = mulr
	}
}

func (ai *AfterImage) setPalContrastG(mulg int32) {
	if len(ai.palfx) > 0 {
		ai.palfx[0].eMul[1] = mulg
	}
}

func (ai *AfterImage) setPalContrastB(mulb int32) {
	if len(ai.palfx) > 0 {
		ai.palfx[0].eMul[2] = mulb
	}
}

// Set up every frame's PalFX in advance
func (ai *AfterImage) setupPalFX() {
	pb := ai.postbright

	if ai.palfx[0].invertblend <= -2 && ai.palfx[0].eInvertall {
		ai.palfx[0].eInvertblend = 3
	} else {
		ai.palfx[0].eInvertblend = ai.palfx[0].invertblend
	}

	for i := 1; i < len(ai.palfx); i++ {
		ai.palfx[i].eColor = ai.palfx[i-1].eColor
		ai.palfx[i].eHue = ai.palfx[i-1].eHue
		ai.palfx[i].eInvertall = ai.palfx[i-1].eInvertall
		ai.palfx[i].eInvertblend = ai.palfx[i-1].eInvertblend
		for j := range pb {
			ai.palfx[i].eAdd[j] = ai.palfx[i-1].eAdd[j] + ai.add[j] + pb[j]
			ai.palfx[i].eMul[j] = int32(float32(ai.palfx[i-1].eMul[j]) * ai.mul[j])
		}
		pb = [3]int32{}
	}
}

func (ai *AfterImage) recAfterImg(sd *SprData, hitpause bool) {
	if ai.time == 0 {
		ai.reccount, ai.timegap = 0, 0
		return
	}
	if ai.restgap <= 0 {
		img := &ai.imgs[ai.imgidx]
		if sd.anim != nil {
			img.anim = &Animation{}
			*img.anim = *sd.anim
			if sd.anim.spr != nil {
				img.anim.spr = newSprite()
				*img.anim.spr = *sd.anim.spr
				if sd.anim.palettedata != nil {
					sd.anim.palettedata.SwapPalMap(&sd.fx.remap)
					img.anim.spr.Pal = sd.anim.spr.GetPal(sd.anim.palettedata)
					sd.anim.palettedata.SwapPalMap(&sd.fx.remap)
				} else {
					sd.anim.sff.palList.SwapPalMap(&sd.fx.remap)
					img.anim.spr.Pal = sd.anim.spr.GetPal(&sd.anim.sff.palList)
					sd.anim.sff.palList.SwapPalMap(&sd.fx.remap)
				}
			}
		} else {
			img.anim = nil
		}
		img.pos = sd.pos
		img.scl = sd.scl
		img.rot = sd.rot
		img.projection = sd.projection
		img.fLength = sd.fLength
		img.oldVer = sd.oldVer
		img.priority = sd.priority - 2 // Starting afterimage sprpriority offset
		ai.imgidx = (ai.imgidx + 1) & 63
		ai.reccount++
		ai.restgap = ai.timegap
	}
	ai.restgap--
	ai.timecount++
}

func (ai *AfterImage) recAndCue(sd *SprData, rec bool, hitpause bool, layer int32) {
	if ai.time == 0 || (ai.timecount >= ai.timegap*ai.length+ai.time-1 && ai.time > 0) ||
		ai.timegap < 1 || ai.timegap > 32767 ||
		ai.framegap < 1 || ai.framegap > 32767 {
		ai.time = 0
		ai.reccount, ai.timecount, ai.timegap = 0, 0, 0
		return
	}

	end := Min(sys.cfg.Config.AfterImageMax,
		(Min(Min(ai.reccount, int32(len(ai.imgs))), ai.length)/ai.framegap)*ai.framegap)

	// Decide layering
	sprs := &sys.spritesLayer0
	if layer > 0 {
		sprs = &sys.spritesLayer1
	} else if layer < 0 {
		sprs = &sys.spritesLayerN1
	}

	for i := ai.framegap; i <= end; i += ai.framegap {
		img := &ai.imgs[(ai.imgidx-i)&63]
		if img.priority >= sd.priority { // Maximum afterimage sprpriority offset
			img.priority = sd.priority - 2
		}
		if ai.time < 0 || (ai.timecount/ai.timegap-i) < (ai.time-2)/ai.timegap+1 {
			step := i/ai.framegap - 1
			ai.palfx[step].remap = sd.fx.remap
			sprs.add(&SprData{
				anim:         img.anim,
				fx:           ai.palfx[step],
				pos:          img.pos,
				scl:          img.scl,
				trans:        ai.trans,
				alpha:        ai.alpha,
				priority:     img.priority - step, // Afterimages decrease in sprpriority over time
				rot:          img.rot,
				screen:       false,
				undarken:     sd.undarken,
				oldVer:       sd.oldVer,
				facing:       sd.facing,
				airOffsetFix: sd.airOffsetFix,
				projection:   img.projection,
				fLength:      img.fLength,
				window:       sd.window,
				xshear:       sd.xshear,
			})
			// Afterimages don't cast shadows or reflections
		}
	}
	if rec || hitpause && ai.ignorehitpause {
		ai.recAfterImg(sd, hitpause)
	}
}

type Explod struct {
	id                  int32
	time                int32
	postype             PosType
	space               Space
	bindId              int32
	bindtime            int32
	pos                 [3]float32
	relativePos         [3]float32
	offset              [3]float32
	relativef           float32
	facing              float32
	vfacing             float32
	scale               [2]float32
	removeongethit      bool
	removeonchangestate bool
	statehaschanged     bool
	removetime          int32
	velocity            [3]float32
	friction            [3]float32
	accel               [3]float32
	sprpriority         int32
	layerno             int32
	shadow              [3]int32
	supermovetime       int32
	pausemovetime       int32
	anim                *Animation
	animNo              int32
	anim_ffx            string
	animPN              int
	spritePN            int
	animelem            int32
	animelemtime        int32
	animfreeze          bool
	ontop               bool // Legacy compatibility
	under               bool
	trans               TransType
	alpha               [2]int32
	ownpal              bool
	remappal            [2]int32
	ignorehitpause      bool
	rot                 Rotation
	anglerot            [3]float32
	xshear              float32
	projection          Projection
	fLength             float32
	oldPos              [3]float32
	newPos              [3]float32
	interPos            [3]float32
	playerId            int32
	palfx               *PalFX
	palfxdef            PalFXDef
	window              [4]float32
	//lockSpriteFacing     bool
	localscl   float32
	localcoord float32
	//blendmode            int32
	start_animelem       int32
	start_scale          [2]float32
	start_rot            [3]float32
	start_alpha          [2]int32
	start_fLength        float32
	start_xshear         float32
	interpolate          bool
	interpolate_time     [2]int32
	interpolate_animelem [3]int32
	interpolate_scale    [4]float32
	interpolate_alpha    [4]int32
	interpolate_pos      [6]float32
	interpolate_angle    [6]float32
	interpolate_fLength  [2]float32
	interpolate_xshear   [2]float32
}

func newExplod() *Explod {
	return &Explod{}
}

// Set default values according to char who creates the explod
func (e *Explod) initFromChar(c *Char) *Explod {
	*e = Explod{
		id:           -1,
		playerId:     c.id,
		animPN:       c.playerNo,
		spritePN:     c.playerNo,
		layerno:      c.layerNo,
		palfx:        c.getPalfx(),   // Safeguard. Overridden later
		palfxdef:     *newPalFXDef(), // Actual PalFX handled later
		bindtime:     1,              // Not documented but confirmed
		scale:        [2]float32{1, 1},
		removetime:   -2,
		postype:      PT_P1,
		space:        Space_none,
		relativef:    1,
		facing:       1,
		vfacing:      1,
		localscl:     c.localscl,
		localcoord:   c.localcoord,
		projection:   Projection_Orthographic,
		window:       [4]float32{0, 0, 0, 0},
		animelem:     1,
		animelemtime: 0,
		//blendmode:         0,
		trans:             TT_default,
		alpha:             [2]int32{-1, 0},
		bindId:            -2,
		ignorehitpause:    true,
		interpolate_scale: [4]float32{1, 1, 0, 0},
		friction:          [3]float32{1, 1, 1},
		remappal:          [2]int32{-1, 0},
	}

	// Backward compatibility
	if c.stWgi().ikemenver[0] == 0 && c.stWgi().ikemenver[1] == 0 &&
		c.stWgi().mugenver[0] == 1 && c.stWgi().mugenver[1] == 1 {
		e.projection = Projection_Perspective
	}

	return e
}

func (e *Explod) setAllPosX(x float32) {
	e.pos[0], e.oldPos[0], e.newPos[0] = x, x, x
}

func (e *Explod) setAllPosY(y float32) {
	e.pos[1], e.oldPos[1], e.newPos[1] = y, y, y
}

func (e *Explod) setAllPosZ(z float32) {
	e.pos[2], e.oldPos[2], e.newPos[2] = z, z, z
}

func (e *Explod) setBind(bId int32) {
	if e.space == Space_screen && (e.postype == PT_P1 || e.postype == PT_P2) {
		return
	}
	e.bindId = bId
}

// Set explod position based on postype and space
func (e *Explod) setPos(c *Char) {
	pPos := func(c *Char) {
		e.bindId, e.facing = c.id, c.facing

		e.relativePos[0] *= c.facing

		posX := (c.pos[0] + c.offsetX()) * c.localscl / e.localscl
		posY := (c.pos[1] + c.offsetY()) * c.localscl / e.localscl
		posZ := c.pos[2] * c.localscl / e.localscl

		if e.space == Space_screen {
			e.offset[0] = posX
			e.offset[1] = sys.cam.GroundLevel()*e.localscl + posY
			e.offset[2] = 0 // posZ? Technically screen has no depth
		} else {
			e.setAllPosX(posX)
			e.setAllPosY(posY)
			e.setAllPosZ(posZ)
		}
	}
	lPos := func() {
		if e.space == Space_screen {
			e.offset[0] = -(float32(sys.gameWidth) / e.localscl / 2)
		} else {
			e.offset[0] = sys.cam.ScreenPos[0] / e.localscl
		}
	}
	rPos := func() {
		if e.space == Space_screen {
			e.offset[0] = float32(sys.gameWidth) / e.localscl / 2
		} else {
			e.offset[0] = sys.cam.ScreenPos[0] / e.localscl
		}
	}
	// Set space based on postype in case it's missing
	if e.space == Space_none {
		switch e.postype {
		case PT_Front, PT_Back, PT_Left, PT_Right:
			e.space = Space_screen
		default:
			e.space = Space_stage
		}
	}
	switch e.postype {
	case PT_P1:
		pPos(c)
	case PT_P2:
		if p2 := c.p2(); p2 != nil {
			pPos(p2)
		}
	case PT_Front, PT_Back:
		if e.postype == PT_Back {
			e.facing = c.facing
		}
		// Convert back and front types to left and right
		if c.facing > 0 && e.postype == PT_Front || c.facing < 0 && e.postype == PT_Back {
			if e.postype == PT_Back {
				e.relativePos[0] *= -1
			}
			e.postype = PT_Right
			rPos()
		} else {
			// postype = front does not cause pos to invert based on the character's facing
			//if e.postype == PT_Front && c.gi().mugenver[0] != 1 {
			// In older versions, front does not reflect the character's facing direction
			// It seems that even in version 1.1, it is not reflected
			//	e.facing = e.relativef
			//}
			e.postype = PT_Left
			lPos()
		}
	case PT_Left:
		lPos()
	case PT_Right:
		rPos()
	case PT_None:
		if e.space == Space_screen {
			e.offset[0] = -(float32(sys.gameWidth) / e.localscl / 2)
		}
	}
}

func (e *Explod) matchId(eid, pid int32) bool {
	return e.id >= 0 && e.playerId == pid && (eid < 0 || e.id == eid)
}

func (e *Explod) setAnim() {
	c := sys.playerID(e.playerId)
	if c == nil {
		return
	}

	// Validate AnimPlayerNo
	if e.animPN < 0 {
		e.animPN = c.playerNo
	} else if e.animPN >= len(sys.chars) || len(sys.chars[e.animPN]) == 0 {
		sys.appendToConsole(c.warn() + fmt.Sprintf("Invalid Explod animPlayerNo: %v", e.animPN+1))
		return
	}

	// Validate SpritePlayerNo
	if e.spritePN < 0 {
		e.spritePN = c.playerNo
	} else if e.spritePN >= len(sys.chars) || len(sys.chars[e.spritePN]) == 0 {
		sys.appendToConsole(c.warn() + fmt.Sprintf("Invalid Explod spritePlayerNo: %v", e.spritePN+1))
		return
	}

	// Get animation with sprite owner context
	a := c.getAnimSprite(e.animNo, e.animPN, e.spritePN, e.anim_ffx, e.ownpal, true)
	if a == nil {
		return
	}
	e.anim = a

	// For common FX, mark owner as undefined for triggers
	if e.anim_ffx != "" && e.anim_ffx != "s" {
		e.animPN = -1
		e.spritePN = -1
	}
}

func (e *Explod) setAnimElem() {
	if e.anim != nil {
		// Validate animelem just in case
		if e.animelem < 1 || int(e.animelem) > len(e.anim.frames) {
			e.animelem = 1
		}
		// Validate animelemtime just in case
		if e.animelemtime != 0 {
			frametime := e.anim.frames[e.animelem-1].Time
			if e.animelemtime < 0 || (frametime != -1 && e.animelemtime >= frametime) {
				e.animelemtime = 0
			}
		}
		// Set them
		e.anim.SetAnimElem(e.animelem, e.animelemtime)
	}
}

func (e *Explod) update(playerNo int) {
	if e.anim == nil {
		e.id = IErr
	}

	if e.id == IErr {
		e.anim = nil
		return
	}

	parent := sys.playerID(e.playerId)
	root := sys.chars[playerNo][0]

	if root.scf(SCF_disabled) {
		return
	}

	// Remove on get hit
	if sys.tickNextFrame() && e.removeongethit &&
		parent != nil && parent.csf(CSF_gethit) && !parent.inGuardState() {
		e.id, e.anim = IErr, nil
		return
	}

	// Remove on ChangeState
	if sys.tickNextFrame() && e.removeonchangestate && e.statehaschanged {
		e.id, e.anim = IErr, nil
		return
	}

	paused := false
	if sys.supertime > 0 {
		paused = (e.supermovetime >= 0 && e.time >= e.supermovetime) || e.supermovetime < -2
	} else if sys.pausetime > 0 {
		paused = (e.pausemovetime >= 0 && e.time >= e.pausemovetime) || e.pausemovetime < -2
	}

	act := !paused

	if act && !e.ignorehitpause {
		act = parent == nil || parent.acttmp%2 >= 0
	}

	if sys.tickFrame() {
		if e.removetime >= 0 && e.time >= e.removetime ||
			act && e.removetime < -1 && e.anim.loopend {
			e.id, e.anim = IErr, nil
			return
		}
	}

	oldVer := root.gi().mugenverF < 1.1

	// Bind explod to parent
	// In Mugen this only happens if the explod is not paused, hence "act"
	if act && e.bindtime != 0 &&
		(e.space == Space_stage || (e.space == Space_screen && (e.postype <= PT_P2 || oldVer))) {
		if bindchar := sys.playerID(e.bindId); bindchar != nil {
			e.pos[0] = bindchar.interPos[0]*bindchar.localscl/e.localscl + bindchar.offsetX()*bindchar.localscl/e.localscl
			e.pos[1] = bindchar.interPos[1]*bindchar.localscl/e.localscl + bindchar.offsetY()*bindchar.localscl/e.localscl
			e.pos[2] = bindchar.interPos[2] * bindchar.localscl / e.localscl
		} else {
			// Doesn't seem necessary to do this, since MUGEN 1.1 seems to carry bindtime even if
			// you change bindId to something that doesn't point to any character
			// e.bindtime = 0
			// e.setAllPosX(e.pos[0])
			// e.setAllPosY(e.pos[1])
		}
	} else {
		// Explod position interpolation
		spd := sys.tickInterpolation()
		for i := range e.pos {
			e.pos[i] = e.newPos[i] - (e.newPos[i]-e.oldPos[i])*(1-spd)
		}
	}
	off := e.relativePos
	// Left and right pos types change relative position depending on stage camera zoom and game width
	if e.space == Space_stage {
		if e.postype == PT_Left {
			off[0] = off[0] / sys.cam.Scale
		} else if e.postype == PT_Right {
			off[0] = (off[0] + float32(sys.gameWidth)) / sys.cam.Scale
		}
	}

	var facing float32 = e.facing * e.relativef
	//if e.lockSpriteFacing {
	//	facing = -1
	//}

	if sys.tickFrame() && act {
		e.anim.UpdateSprite()
	}

	sprs := &sys.spritesLayer0
	if e.layerno > 0 {
		sprs = &sys.spritesLayer1
	} else if e.layerno < 0 {
		sprs = &sys.spritesLayerN1
	} else if e.under {
		sprs = &sys.spritesLayerU
	}

	var pfx *PalFX
	if e.palfx != nil && (!e.anim.isCommonFX() || e.ownpal) {
		pfx = e.palfx
	} else {
		pfx = &PalFX{}
		*pfx = *e.palfx
		pfx.remap = nil
	}

	alp := e.alpha
	anglerot := e.anglerot
	fLength := e.fLength
	scale := e.scale
	xshear := e.xshear
	if e.interpolate {
		e.Interpolate(act, &scale, &alp, &anglerot, &fLength, &xshear)
	}
	if alp[0] < 0 {
		alp[0] = -1
	}
	if (e.facing*e.relativef < 0) != (e.vfacing < 0) {
		anglerot[0] *= -1
		anglerot[2] *= -1
	}
	sdwalp := 255 - alp[1]
	if sdwalp < 0 {
		sdwalp = 256
	}
	if fLength <= 0 {
		fLength = 2048
	}
	fLength = fLength * e.localscl
	rot := e.rot
	rot.angle = anglerot[0]
	rot.xangle = anglerot[1]
	rot.yangle = anglerot[2]

	// Interpolated position
	// With z-axis it's important that we don't use localscl here yet
	e.interPos = [3]float32{
		e.pos[0] + e.offset[0] + off[0] + e.interpolate_pos[0],
		e.pos[1] + e.offset[1] + off[1] + e.interpolate_pos[1],
		e.pos[2] + e.offset[2] + off[2] + e.interpolate_pos[2],
	}

	// Set drawing position
	drawpos := [2]float32{e.interPos[0] * e.localscl, e.interPos[1] * e.localscl}

	// Set scale
	// Mugen uses "localscl" instead of "320 / e.localcoord" but that makes the scale jump in custom states of different localcoord
	drawscale := [2]float32{
		facing * scale[0] * (320 / e.localcoord),
		e.vfacing * scale[1] * (320 / e.localcoord),
	}

	// Apply Z axis perspective
	if e.space == Space_stage && sys.zEnabled() {
		zscale := sys.updateZScale(e.interPos[2], e.localscl)
		drawpos = sys.drawposXYfromZ(drawpos, e.localscl, e.interPos[2], zscale)
		drawscale[0] *= zscale
		drawscale[1] *= zscale
	}

	var ewin = [4]float32{
		e.window[0] * drawscale[0],
		e.window[1] * drawscale[1],
		e.window[2] * drawscale[0],
		e.window[3] * drawscale[1],
	}

	// Add sprite to draw list
	sd := &SprData{
		anim:         e.anim,
		fx:           pfx,
		pos:          drawpos,
		scl:          drawscale,
		trans:        e.trans,
		alpha:        alp,
		priority:     e.sprpriority + int32(e.interPos[2]*e.localscl),
		rot:          rot,
		screen:       e.space == Space_screen,
		undarken:     parent != nil && parent.ignoreDarkenTime > 0,
		oldVer:       oldVer,
		facing:       facing,
		airOffsetFix: [2]float32{1, 1},
		projection:   int32(e.projection),
		fLength:      fLength,
		window:       ewin,
		xshear:       xshear,
	}
	sprs.add(sd)

	// Add shadow if color is not 0
	sdwclr := e.shadow[0]<<16 | e.shadow[1]&0xff<<8 | e.shadow[2]&0xff
	if sdwclr != 0 {
		sdwalp := 255 - alp[1]
		if sdwalp < 0 {
			sdwalp = 256
		}
		drawZoff := sys.posZtoYoffset(e.interPos[2], e.localscl)
		// Add shadow sprite
		sys.shadows.add(&ShadowSprite{
			SprData:      sd,
			shadowColor:  sdwclr,
			shadowAlpha:  sdwalp,
			shadowOffset: [2]float32{0, sys.stage.sdw.yscale*drawZoff + drawZoff},
			fadeOffset:   drawZoff,
		})
		// Add reflection sprite
		sys.reflections.add(&ReflectionSprite{
			SprData:       sd,
			reflectOffset: [2]float32{0, sys.stage.reflection.yscale*drawZoff + drawZoff},
			fadeOffset:    drawZoff,
		})
	}
	if sys.tickNextFrame() {

		//if e.space == Space_screen && e.bindtime == 0 {
		//	if e.space <= Space_none {
		//		switch e.postype {
		//		case PT_Left:
		//			for i := range e.pos {
		//				e.pos[i] = sys.cam.ScreenPos[i] + e.offset[i]/sys.cam.Scale
		//			}
		//		case PT_Right:
		//			e.pos[0] = sys.cam.ScreenPos[0] +
		//				(float32(sys.gameWidth)+e.offset[0])/sys.cam.Scale
		//			e.pos[1] = sys.cam.ScreenPos[1] + e.offset[1]/sys.cam.Scale
		//		}
		//	} else if e.space == Space_screen {
		//		for i := range e.pos {
		//			e.pos[i] = sys.cam.ScreenPos[i] + e.offset[i]/sys.cam.Scale
		//		}
		//	}
		//}

		if act {
			if e.palfx != nil && e.ownpal {
				e.palfx.step()
			}
			e.oldPos = e.pos
			e.newPos[0] = e.pos[0] + e.velocity[0]*e.facing
			e.newPos[1] = e.pos[1] + e.velocity[1]
			e.newPos[2] = e.pos[2] + e.velocity[2]
			for i := range e.velocity {
				e.velocity[i] *= e.friction[i]
				e.velocity[i] += e.accel[i]
				if math.Abs(float64(e.velocity[i])) < 0.1 && math.Abs(float64(e.friction[i])) < 1 {
					e.velocity[i] = 0
				}
			}
			eleminterpolate := e.interpolate && e.interpolate_time[1] > 0 && e.interpolate_animelem[1] >= 0
			if e.animfreeze || eleminterpolate {
				e.setAnimElem()
			} else {
				e.anim.Action()
			}
			e.time++
			if e.bindtime > 0 {
				e.bindtime--
			}
		} else {
			e.setAllPosX(e.pos[0])
			e.setAllPosY(e.pos[1])
			e.setAllPosZ(e.pos[2])
		}
	}
}

func (e *Explod) Interpolate(act bool, scale *[2]float32, alpha *[2]int32, anglerot *[3]float32, fLength *float32, xshear *float32) {
	if sys.tickNextFrame() && act {
		t := float32(e.interpolate_time[1]) / float32(e.interpolate_time[0])
		e.interpolate_fLength[0] = Lerp(e.interpolate_fLength[1], e.start_fLength, t)
		e.interpolate_xshear[0] = Lerp(e.interpolate_xshear[1], e.start_xshear, t)
		if e.interpolate_animelem[1] >= 0 {
			elem := Ceil(Lerp(float32(e.interpolate_animelem[0]-1), float32(e.interpolate_animelem[1]), 1-t))

			if e.interpolate_animelem[0] > e.interpolate_animelem[1] {
				elem = Ceil(Lerp(float32(e.interpolate_animelem[1]-1), float32(e.interpolate_animelem[0]), t))
			}
			e.animelem = Clamp(elem, Min(e.interpolate_animelem[0], e.interpolate_animelem[1]), Max(e.interpolate_animelem[0], e.interpolate_animelem[1]))
		}
		for i := 0; i < 3; i++ {
			e.interpolate_pos[i] = Lerp(e.interpolate_pos[i+3], 0, t)
			if i < 2 {
				e.interpolate_scale[i] = Lerp(e.interpolate_scale[i+2], e.start_scale[i], t) //-e.start_scale[i]
				e.interpolate_alpha[i] = Clamp(int32(Lerp(float32(e.interpolate_alpha[i+2]), float32(e.start_alpha[i]), t)), 0, 255)
			}
			e.interpolate_angle[i] = Lerp(e.interpolate_angle[i+3], e.start_rot[i], t)
		}
		if e.interpolate_time[1] > 0 {
			e.interpolate_time[1]--
		}
	}
	for i := 0; i < 3; i++ {
		if i < 2 {
			(*scale)[i] = e.interpolate_scale[i] * e.scale[i]
			// Update alpha regardless of transparency type. Let the type handle the rendering
			(*alpha)[i] = int32(float32(e.interpolate_alpha[i]) * (float32(e.alpha[i]) / 255))
		}
		(*anglerot)[i] = e.interpolate_angle[i] + e.anglerot[i]
	}
	*fLength = e.interpolate_fLength[0] + e.fLength
	*xshear = e.interpolate_xshear[0]
}

func (e *Explod) resetInterpolation(pfd *PalFXDef) {
	for i := 0; i < 3; i++ {
		for j := 0; j < 2; j++ {
			v := (i + (j * 3))
			if e.ownpal {
				pfd.iadd[v] = pfd.add[i]
				pfd.imul[v] = pfd.mul[i]
			}
			e.interpolate_angle[v] = e.anglerot[i]
			if i < 2 {
				v = (i + (j * 2))
				e.interpolate_pos[v] = 0
				e.interpolate_scale[v] = e.scale[i]
				e.interpolate_alpha[v] = e.alpha[i]
				if j == 0 && e.ownpal {
					pfd.icolor[i] = pfd.color
					pfd.ihue[i] = pfd.hue
				}
			}
		}
	}
	for i := 0; i < 2; i++ {
		e.interpolate_animelem[i] = -1
		e.interpolate_fLength[i] = e.fLength
		e.interpolate_xshear[i] = e.xshear
	}
}

type Projectile struct {
	playerno        int
	hitdef          HitDef
	id              int32
	anim            int32
	anim_ffx        string
	hitanim         int32
	hitanim_ffx     string
	remanim         int32
	remanim_ffx     string
	cancelanim      int32
	cancelanim_ffx  string
	scale           [2]float32
	anglerot        [3]float32
	rot             Rotation
	projection      Projection
	fLength         float32
	clsnScale       [2]float32
	clsnAngle       float32
	zScale          float32
	remove          bool
	removetime      int32
	velocity        [3]float32
	remvelocity     [3]float32
	accel           [3]float32
	velmul          [3]float32
	hits            int32
	totalhits       int32
	misstime        int32
	priority        int32
	priorityPoints  int32
	sprpriority     int32
	layerno         int32
	edgebound       int32
	stagebound      int32
	heightbound     [2]int32
	depthbound      int32
	pos             [3]float32
	interPos        [3]float32
	facing          float32
	removefacing    float32
	shadow          [3]int32
	supermovetime   int32
	pausemovetime   int32
	ani             *Animation
	curmisstime     int32
	hitpause        int32
	oldPos          [3]float32
	newPos          [3]float32
	aimg            AfterImage
	palfx           *PalFX
	window          [4]float32
	xshear          float32
	localscl        float32
	localcoord      float32
	parentAttackMul [4]float32
	platform        bool
	platformWidth   [2]float32
	platformHeight  [2]float32
	platformAngle   float32
	platformFence   bool
	remflag         bool
	freezeflag      bool
	contactflag     bool
	time            int32
}

func newProjectile() *Projectile {
	return &Projectile{}
}

// Set defaults according to projectile owner
// TODO: Check how much should come from char who uses Projectile sctrl versus from the root
func (p *Projectile) initFromChar(c *Char) *Projectile {
	// Local scale exception
	localscl := c.localscl
	if c.minus == -2 || c.minus == -4 {
		localscl = 320 / c.localcoord
	}

	*p = Projectile{
		id:              0,
		playerno:        c.playerNo,
		hitanim:         -1,
		remanim:         IErr,
		cancelanim:      IErr,
		scale:           [2]float32{1, 1},
		clsnScale:       [2]float32{1, 1},
		clsnAngle:       0,
		remove:          true,
		localscl:        localscl,
		localcoord:      c.localcoord,
		layerno:         c.layerNo,
		palfx:           c.getPalfx(),
		parentAttackMul: c.attackMul, // Projectile attackmul is decided upon its creation only
		removetime:      -1,
		velmul:          [3]float32{1, 1, 1},
		hits:            1,
		totalhits:       1,
		priority:        1,
		priorityPoints:  1,
		sprpriority:     3,
		edgebound:       int32(40 / localscl), // TODO: These probably need "originLocalscl"
		stagebound:      int32(40 / localscl),
		heightbound:     [2]int32{int32(-240 / localscl), int32(1 / localscl)},
		depthbound:      math.MaxInt32,
		facing:          1,
		aimg:            *newAfterImage(),
		projection:      Projection_Orthographic,
		platformFence:   true,
	}

	// Backward compatibility
	if c.stWgi().ikemenver[0] == 0 && c.stWgi().ikemenver[1] == 0 &&
		c.stWgi().mugenver[0] == 1 && c.stWgi().mugenver[1] == 1 {
		p.projection = Projection_Perspective
	}

	// Initialize projectile Hitdef. Must be placed after its localscl is determined
	// https://github.com/ikemen-engine/Ikemen-GO/issues/2087
	p.hitdef.clear(c, p.localscl)
	p.hitdef.isprojectile = true
	p.hitdef.playerNo = sys.workingState.playerNo
	p.hitdef.guard_dist_x = [2]float32{c.size.proj.attack.dist.width[0], c.size.proj.attack.dist.width[1]}
	p.hitdef.guard_dist_y = [2]float32{c.size.proj.attack.dist.height[0], c.size.proj.attack.dist.height[1]}
	p.hitdef.guard_dist_z = [2]float32{c.size.proj.attack.dist.depth[0], c.size.proj.attack.dist.depth[1]}

	return p
}

func (p *Projectile) setAllPos(pos [3]float32) {
	p.pos = pos
	p.oldPos = pos
	p.newPos = pos
	p.interPos = pos
}

func (p *Projectile) paused(playerNo int) bool {
	//if !sys.chars[playerNo][0].pause() {
	if sys.supertime > 0 {
		if p.supermovetime == 0 || p.supermovetime < -1 {
			return true
		}
	} else if sys.pausetime > 0 {
		if p.pausemovetime == 0 || p.pausemovetime < -1 {
			return true
		}
	}
	//}
	return false
}

func (p *Projectile) update() {
	// Check projectile removal conditions
	if sys.tickFrame() && !p.paused(p.playerno) && p.hitpause == 0 {
		if p.anim >= 0 && !p.remflag {
			remove := true
			root := sys.chars[p.playerno][0]
			if p.hits < 0 {
				// Remove behavior
				if p.hits == -1 && p.remove {
					if p.hitanim != p.anim || p.hitanim_ffx != p.anim_ffx {
						if p.hitanim == -1 {
							p.ani = nil
						} else if ani := root.getSelfAnimSprite(p.hitanim, p.hitanim_ffx, true, true); ani != nil {
							p.ani = ani
						}
					}
				}
				// Cancel behavior
				if p.hits == -2 {
					if p.cancelanim != p.anim || p.cancelanim_ffx != p.anim_ffx {
						if p.cancelanim == -1 {
							p.ani = nil
						} else if ani := root.getSelfAnimSprite(p.cancelanim, p.cancelanim_ffx, true, true); ani != nil {
							p.ani = ani
						}
					}
				}
			} else if p.removetime == 0 ||
				p.removetime <= -2 && (p.ani == nil || p.ani.loopend) ||
				p.pos[0] < (sys.xmin-sys.screenleft)/p.localscl-float32(p.edgebound) ||
				p.pos[0] > (sys.xmax+sys.screenright)/p.localscl+float32(p.edgebound) ||
				p.velocity[0]*p.facing < 0 && p.pos[0] < sys.cam.XMin/p.localscl-float32(p.stagebound) ||
				p.velocity[0]*p.facing > 0 && p.pos[0] > sys.cam.XMax/p.localscl+float32(p.stagebound) ||
				p.velocity[1] > 0 && p.pos[1] > float32(p.heightbound[1]) ||
				p.velocity[1] < 0 && p.pos[1] < float32(p.heightbound[0]) ||
				p.pos[2] < (sys.zmin/p.localscl-float32(p.depthbound)) ||
				p.pos[2] > (sys.zmax/p.localscl+float32(p.depthbound)) {
				if p.remanim != p.anim || p.remanim_ffx != p.anim_ffx {
					if p.remanim != -2 {
						if p.remanim == -1 {
							p.ani = nil
						} else if ani := root.getSelfAnimSprite(p.remanim, p.remanim_ffx, true, true); ani != nil {
							p.ani = ani
							// In Mugen, if remanim is invalid the projectile will keep the current one
							// https://github.com/ikemen-engine/Ikemen-GO/issues/2584
						}
					}
				}
				remove = true
			} else {
				remove = false
			}
			// Active to removing transition
			if remove {
				p.remflag = true
				if p.ani != nil {
					p.ani.UpdateSprite()
				}
				p.velocity = p.remvelocity
				if p.facing == p.removefacing {
					p.facing = p.removefacing
				} else {
					p.velocity[0] *= -1
				}
				p.accel = [3]float32{0, 0, 0}
				p.velmul = [3]float32{1, 1, 1}
				p.anim = -1
				// In Mugen, projectiles can hit even after their removetime expires
				// https://github.com/ikemen-engine/Ikemen-GO/issues/1362
				//if p.hits >= 0 {
				//	p.hits = -1
				//}
			}
		}
		// Remove projectile
		if p.remflag {
			if p.ani != nil && (p.ani.totaltime <= 0 || p.ani.AnimTime() == 0) {
				p.ani = nil
			}
			if p.ani == nil && p.id >= 0 {
				p.id = ^p.id
			}
		}
	}
	if p.paused(p.playerno) || p.hitpause > 0 || p.freezeflag {
		p.setAllPos(p.pos)
		// There's a minor issue here where a projectile will lag behind one frame relative to Mugen if created during a pause
	} else {
		if sys.tickFrame() {
			p.pos = [...]float32{p.pos[0] + p.velocity[0]*p.facing, p.pos[1] + p.velocity[1], p.pos[2] + p.velocity[2]}
			p.interPos = [...]float32{p.pos[0], p.pos[1], p.pos[2]}
		}
		spd := sys.tickInterpolation()
		for i := 0; i < 3; i++ {
			p.interPos[i] = p.pos[i] - (p.pos[i]-p.oldPos[i])*(1-spd)
		}
		if sys.tickNextFrame() {
			p.oldPos = p.pos
			for i := range p.velocity {
				p.velocity[i] += p.accel[i]
				p.velocity[i] *= p.velmul[i]
			}
			if p.velocity[0] < 0 && p.anim != -1 {
				p.facing *= -1
				p.velocity[0] *= -1
				p.accel[0] *= -1
			}
		}
	}
	// Update Z scale
	p.zScale = sys.updateZScale(p.pos[2], p.localscl)
}

// Flag a projectile as cancelled
func (p *Projectile) flagProjCancel() {
	p.hits = -2
	if p.playerno >= 0 && p.playerno < len(sys.cgi) {
		r := &sys.cgi[p.playerno]
		if r != nil {
			r.pctype = PC_Cancel
			r.pctime = 0
			r.pcid = p.id
		}
	}
}

// This subtracts projectile hits when two projectiles clash
func (p *Projectile) cancelHits(opp *Projectile) {
	// Check priority
	if p.priorityPoints > opp.priorityPoints {
		p.priorityPoints--
	} else {
		p.hits--
	}
	// Flag as cancelled
	if p.hits <= 0 {
		p.flagProjCancel()
	}
	// Set hitpause
	if p.hits > 0 {
		p.hitpause = Max(0, p.hitdef.pausetime[0]) // -Btoi(c.gi().mugenver[0] == 0))
	} else {
		p.hitpause = 0
	}
}

// This function only checks if a projectile hits another projectile
func (p *Projectile) tradeDetection(playerNo, index int) {

	// Skip if this projectile can't trade at all
	// Projectiles can trade even if they are spawned with 0 hits
	if p.remflag || p.hits < 0 || p.id < 0 {
		return
	}

	// Skip if this projectile can't run a collision check at all
	if p.ani == nil || len(p.ani.frames) == 0 || p.ani.CurrentFrame().Clsn2 == nil {
		return
	}

	// Loop through all players starting from the current one
	// Previous players are skipped to prevent checking the same projectile pairs twice
	for i := playerNo; i < len(sys.chars) && p.hits >= 0; i++ {
		if len(sys.chars[i]) == 0 {
			continue
		}

		// If at parent's index, skip self and previously checked pairs
		// In Mugen, projectiles just never hit other projectiles from the same player
		startj := 0
		if i == playerNo {
			startj = index + 1
		}

		// Loop through their projectiles
		for j := startj; j < len(sys.projs[i]); j++ {
			pr := sys.projs[i][j]

			// Skip if other projectile can't trade
			if pr.remflag || pr.hits < 0 || pr.id < 0 {
				continue
			}

			// Skip if other projectile can't run collision check
			if pr.ani == nil || len(pr.ani.frames) == 0 || pr.ani.CurrentFrame().Clsn2 == nil {
				continue
			}

			// Teamside check for both projectiles
			if p.hitdef.affectteam != 0 && pr.hitdef.affectteam != 0 {
				friendly := p.hitdef.teamside == pr.hitdef.teamside
				if (p.hitdef.affectteam > 0 && pr.hitdef.affectteam > 0 && friendly) ||
					(p.hitdef.affectteam < 0 && pr.hitdef.affectteam < 0 && !friendly) {
					continue
				}
			}

			// Run Z axis check
			if !sys.zAxisOverlap(p.pos[2], p.hitdef.attack_depth[0], p.hitdef.attack_depth[1], p.localscl,
				pr.pos[2], pr.hitdef.attack_depth[0], pr.hitdef.attack_depth[1], pr.localscl) {
				continue
			}

			// Run Clsn check
			clsn1 := p.ani.CurrentFrame().Clsn2 // Projectiles trade with their Clsn2 only
			clsn2 := pr.ani.CurrentFrame().Clsn2
			if clsn1 != nil && clsn2 != nil {
				if sys.clsnOverlap(clsn1,
					[...]float32{p.clsnScale[0] * p.localscl, p.clsnScale[1] * p.localscl},
					[...]float32{p.pos[0] * p.localscl, p.pos[1] * p.localscl},
					p.facing,
					p.clsnAngle,
					clsn2,
					[...]float32{pr.clsnScale[0] * pr.localscl, pr.clsnScale[1] * pr.localscl},
					[...]float32{pr.pos[0] * pr.localscl, pr.pos[1] * pr.localscl},
					pr.facing,
					pr.clsnAngle) {
					// Subtract projectile hits from each other
					p.cancelHits(pr)
					pr.cancelHits(p)
					// Stop entire loop when out of projectile hits
					if p.hits < 0 {
						break
					}
				}
			}
		}
	}
}

func (p *Projectile) tick() {
	if p.contactflag {
		p.contactflag = false
		// Projectile hitpause should maybe be set in this place instead of using "(p.hitpause <= 0 || p.contactflag)" for hit checking
		p.curmisstime = Max(0, p.misstime)
		if p.hits >= 0 {
			p.hits--
			if p.hits <= 0 {
				p.hits = -1
				p.hitpause = 0
			}
		}
		p.hitdef.air_juggle = 0
	}
	if !p.paused(p.playerno) {
		if p.hitpause <= 0 {
			p.time++ // Only used in ProjVar currently
			if p.removetime > 0 {
				p.removetime--
			}
			if p.curmisstime > 0 {
				p.curmisstime--
			}
			if p.supermovetime > 0 {
				p.supermovetime--
			}
			if p.pausemovetime > 0 {
				p.pausemovetime--
			}
			p.freezeflag = false
		} else {
			p.hitpause--
			p.freezeflag = true // This flag makes projectiles halt in place between multiple hits
		}
	}
}

func (p *Projectile) cueDraw(oldVer bool) {
	notpause := p.hitpause <= 0 && !p.paused(p.playerno)
	if sys.tickFrame() && p.ani != nil && notpause {
		p.ani.UpdateSprite()
	}

	// Projectile Clsn display
	if sys.clsnDisplay && p.ani != nil {
		if frm := p.ani.drawFrame(); frm != nil {
			if clsn := frm.Clsn1; clsn != nil && len(clsn) > 0 {
				sys.debugc1hit.Add(clsn, p.pos[0]*p.localscl, p.pos[1]*p.localscl,
					p.clsnScale[0]*p.localscl*p.facing*p.zScale,
					p.clsnScale[1]*p.localscl*p.zScale,
					p.clsnAngle*p.facing)
			}
			if clsn := frm.Clsn2; clsn != nil && len(clsn) > 0 {
				sys.debugc2hb.Add(clsn, p.pos[0]*p.localscl, p.pos[1]*p.localscl,
					p.clsnScale[0]*p.localscl*p.facing*p.zScale,
					p.clsnScale[1]*p.localscl*p.zScale,
					p.clsnAngle*p.facing)
			}
		}
	}

	if sys.tickNextFrame() && (notpause || !p.paused(p.playerno)) {
		if p.ani != nil && notpause {
			p.ani.Action()
		}
	}

	// Set position
	pos := [2]float32{p.interPos[0] * p.localscl, p.interPos[1] * p.localscl}

	// Set scale
	// Mugen uses "localscl" instead of "320 / e.localcoord" but that makes the scale jump in custom states of different localcoord
	drawscale := [2]float32{
		p.facing * p.scale[0] * p.zScale * (320 / p.localcoord),
		p.scale[1] * p.zScale * (320 / p.localcoord),
	}

	// Apply Z axis perspective
	if sys.zEnabled() {
		pos = sys.drawposXYfromZ(pos, p.localscl, p.interPos[2], p.zScale)
	}

	anglerot := p.anglerot
	fLength := p.fLength

	if fLength <= 0 {
		fLength = 2048
	}

	if p.facing < 0 {
		anglerot[0] *= -1
		anglerot[2] *= -1
	}
	fLength = fLength * p.localscl
	rot := p.rot
	rot.angle = anglerot[0]
	rot.xangle = anglerot[1]
	rot.yangle = anglerot[2]

	sprs := &sys.spritesLayer0
	if p.layerno > 0 {
		sprs = &sys.spritesLayer1
	} else if p.layerno < 0 {
		sprs = &sys.spritesLayerN1
	}

	var pwin = [4]float32{
		p.window[0] * drawscale[0],
		p.window[1] * drawscale[1],
		p.window[2] * drawscale[0],
		p.window[3] * drawscale[1],
	}

	if p.ani != nil {
		// Add sprite to draw list
		sd := &SprData{
			anim:         p.ani,
			fx:           p.palfx,
			pos:          pos,
			scl:          drawscale,
			trans:        TT_default,
			alpha:        [2]int32{-1, 0},
			priority:     p.sprpriority + int32(p.pos[2]*p.localscl),
			rot:          rot,
			screen:       false,
			undarken:     sys.chars[p.playerno][0] != nil && sys.chars[p.playerno][0].ignoreDarkenTime > 0, //p.playerno == sys.superplayerno,
			oldVer:       sys.cgi[p.playerno].mugenver[0] != 1,
			facing:       p.facing,
			airOffsetFix: [2]float32{1, 1},
			projection:   int32(p.projection),
			fLength:      fLength,
			window:       pwin,
			xshear:       p.xshear,
		}
		p.aimg.recAndCue(sd, sys.tickNextFrame() && notpause, false, p.layerno)
		sprs.add(sd)
		// Add a shadow if color is not 0
		sdwclr := p.shadow[0]<<16 | p.shadow[1]&0xff<<8 | p.shadow[2]&0xff
		if sdwclr != 0 {
			drawZoff := sys.posZtoYoffset(p.interPos[2], p.localscl)
			// Add shadow
			sys.shadows.add(&ShadowSprite{
				SprData:      sd,
				shadowColor:  sdwclr,
				shadowAlpha:  255,
				shadowOffset: [2]float32{0, sys.stage.sdw.yscale*drawZoff + drawZoff},
				fadeOffset:   drawZoff,
			})
			// Add reflection
			sys.reflections.add(&ReflectionSprite{
				SprData:       sd,
				reflectOffset: [2]float32{0, sys.stage.reflection.yscale*drawZoff + drawZoff},
				fadeOffset:    drawZoff,
			})
		}
	}
}

type MoveContact int32

const (
	MC_Hit MoveContact = iota
	MC_Guarded
	MC_Reversed
)

type ProjContact int32

const (
	PC_Hit ProjContact = iota
	PC_Guarded
	PC_Cancel
)

type CharGlobalInfo struct {
	def                     string
	nameLow                 string
	displayname             string
	displaynameLow          string
	author                  string
	authorLow               string
	lifebarname             string
	palkeymap               [MaxPalNo]int32
	sff                     *Sff
	palettedata             *Palette
	snd                     *Snd
	anim                    AnimationTable
	palno                   int32
	pal                     [MaxPalNo]string
	palExist                [MaxPalNo]bool
	palSelectable           [MaxPalNo]bool
	ikemenver               [3]uint16
	ikemenverF              float32
	mugenver                [2]uint16
	mugenverF               float32
	data                    CharData
	velocity                CharVelocity
	movement                CharMovement
	states                  map[int32]StateBytecode
	hitPauseToggleFlagCount int32
	pctype                  ProjContact
	pctime, pcid            int32
	quotes                  [MaxQuotes]string
	portraitscale           float32
	constants               map[string]float32
	remapPreset             map[string]RemapPreset
	remappedpal             [2]int32
	localcoord              [2]float32
	fnt                     [10]*Fnt
	fightfxPrefix           string
	fxPath                  []string
	attackBase              int32
	defenceBase             int32
}

func (cgi *CharGlobalInfo) clearPCTime() {
	cgi.pctype = PC_Hit
	cgi.pctime = -1
	cgi.pcid = 0
}

// StateState contains the state variables like stateNo, prevStateNo, time, stateType, moveType, and physics of the current state.
type StateState struct {
	stateType                    StateType
	prevStateType                StateType
	moveType                     MoveType
	prevMoveType                 MoveType
	storeMoveType                bool
	physics                      StateType
	ps                           []int32
	hitPauseExecutionToggleFlags [MaxPlayerNo][]bool // Flags if an sctrl runs during a hit pause on the current tick.
	no, prevno                   int32
	time                         int32
	sb                           StateBytecode
}

func (ss *StateState) changeStateType(t StateType) {
	ss.prevStateType = ss.stateType
	ss.stateType = t
}

func (ss *StateState) changeMoveType(t MoveType) {
	ss.prevMoveType = ss.moveType
	ss.moveType = t
}

func (ss *StateState) clear() {
	ss.changeStateType(ST_S)
	ss.changeMoveType(MT_I)
	ss.physics = ST_N
	ss.ps = nil
	// Iterate over each player's hitPauseExecutionToggleFlags
	for i, v := range ss.hitPauseExecutionToggleFlags {
		// Ensure the slice has enough capacity based on hitPauseToggleFlagCount
		if len(v) < int(sys.cgi[i].hitPauseToggleFlagCount) {
			ss.hitPauseExecutionToggleFlags[i] = make([]bool, sys.cgi[i].hitPauseToggleFlagCount)
		} else {
			// Reset all flags to false
			for i := range v {
				v[i] = false
			}
		}
	}
	// Further clear the hitPauseExecutionToggleFlags
	ss.clearHitPauseExecutionToggleFlags()
	ss.no, ss.prevno = 0, 0
	ss.time = 0
	ss.sb = StateBytecode{}
}

// Resets all hitPauseExecutionToggleFlags to false.
// This ensures that all state controllers are set to execute on the next eligible tick.
func (ss *StateState) clearHitPauseExecutionToggleFlags() {
	for _, v := range ss.hitPauseExecutionToggleFlags {
		for i := range v {
			v[i] = false
		}
	}
}

type HMF int32

const (
	HMF_H HMF = iota
	HMF_M
	HMF_F
)

type CharSystemVar struct {
	airJumpCount          int32
	assertFlag            AssertSpecialFlag
	hitCount              int32
	guardCount            int32
	uniqHitCount          int32
	pauseMovetime         int32
	superMovetime         int32
	ignoreDarkenTime      int32
	unhittableTime        int32
	bindTime              int32
	bindToId              int32
	bindPos               [3]float32
	bindPosAdd            [3]float32
	bindFacing            float32
	hitPauseTime          int32
	rot                   Rotation
	anglerot              [3]float32
	xshear                float32
	projection            Projection
	fLength               float32
	angleDrawScale        [2]float32
	trans                 TransType
	alpha                 [2]int32
	window                [4]float32
	systemFlag            SystemCharFlag
	specialFlag           CharSpecialFlag
	sprPriority           int32
	layerNo               int32
	receivedDmg           int32
	receivedHits          int32
	cornerVelOff          float32
	sizeWidth             [2]float32
	edgeWidth             [2]float32
	sizeHeight            [2]float32
	sizeDepth             [2]float32
	edgeDepth             [2]float32
	sizeBox               [4]float32
	attackMul             [4]float32 // 0 Damage, 1 Red Life, 2 Dizzy Points, 3 Guard Points
	superDefenseMul       float32
	superDefenseMulBuffer float32
	fallDefenseMul        float32
	customDefense         float32
	finalDefense          float64
	defenseMulDelay       bool
	counterHit            bool
	prevNoStandGuard      bool
	prevPauseMovetime     int32
	prevSuperMovetime     int32
}

type Char struct {
	name                string
	palfx               *PalFX
	anim                *Animation
	animBackup          *Animation
	curFrame            *AnimFrame
	cmd                 []CommandList
	ss                  StateState
	controller          int
	id                  int32
	runorder            int32
	helperId            int32
	helperIndex         int32
	parentIndex         int32
	playerNo            int
	teamside            int
	keyctrl             [4]bool
	playerFlag          bool // Root and player type helpers
	hprojectile         bool // Helper type projectile. Currently unused
	animPN              int
	spritePN            int
	animNo              int32
	prevAnimNo          int32
	life                int32
	lifeMax             int32
	power               int32
	powerMax            int32
	dizzyPoints         int32
	dizzyPointsMax      int32
	guardPoints         int32
	guardPointsMax      int32
	redLife             int32
	juggle              int32
	fallTime            int32
	localcoord          float32 // Char localcoord[0] scaled to game resolution
	localscl            float32 // Ratio between 320 and the localcoord of the current state
	animlocalscl        float32
	size                CharSize
	clsnBaseScale       [2]float32
	clsnScaleMul        [2]float32 // From TransformClsn
	clsnScale           [2]float32 // The final one
	clsnAngle           float32
	zScale              float32
	hitdef              HitDef
	ghv                 GetHitVar
	mhv                 MoveHitVar
	hitby               [8]HitBy
	hover               [8]HitOverride
	hoverIdx            int
	hoverKeepState      bool
	mctype              MoveContact
	mctime              int32
	children            []*Char
	isclsnproxy         bool
	targets             []int32
	hitdefTargets       []int32
	hitdefTargetsBuffer []int32
	enemyNearList       []*Char // Enemies retrieved by EnemyNear
	p2EnemyList         []*Char // Enemies retrieved by P2, P4, P6 and P8
	p2EnemyBackup       *Char   // Backup of last valid P2 enemy
	pos                 [3]float32
	interPos            [3]float32 // Interpolated position. For the visuals when game and logic speed are different
	oldPos              [3]float32
	vel                 [3]float32
	facing              float32
	fbFlip              bool
	cnsvar              map[int32]int32
	cnsfvar             map[int32]float32
	cnssysvar           map[int32]int32
	cnssysfvar          map[int32]float32
	CharSystemVar
	aimg              AfterImage
	soundChannels     SoundChannels
	p1facing          float32
	cpucmd            int32
	offset            [2]float32
	stchtmp           bool
	inguarddist       bool
	pushed            bool
	hitdefContact     bool
	atktmp            int8 // 1 hitdef can hit, 0 cannot hit, -1 other
	hittmp            int8 // 0 idle, 1 being hit, 2 falling, -1 reversaldef
	acttmp            int8 // 1 unpaused, 0 default, -1 hitpause, -2 pause
	minus             int8 // Essentially the current negative state
	platformPosY      float32
	groundAngle       float32
	ownpal            bool
	winquote          int32
	memberNo          int
	selectNo          int
	inheritJuggle     int32
	inheritChannels   int32
	mapArray          map[string]float32
	mapDefault        map[string]float32
	remapSpr          RemapPreset
	clipboardText     []string
	dialogue          []string
	immortal          bool
	kovelocity        bool
	preserve          bool
	inputFlag         InputBits
	inputShift        [][2]int
	pauseBool         bool
	downHitOffset     bool
	koEchoTimer       int32
	groundLevel       float32
	sizeBox           [4]float32
	shadowColor       [3]int32
	shadowIntensity   int32
	shadowOffset      [2]float32
	shadowWindow      [4]float32
	shadowXshear      float32
	shadowYscale      float32
	shadowRot         Rotation
	shadowProjection  Projection
	shadowfLength     float32
	reflectColor      [3]int32
	reflectIntensity  int32
	reflectOffset     [2]float32
	reflectWindow     [4]float32
	reflectXshear     float32
	reflectYscale     float32
	reflectRot        Rotation
	reflectProjection Projection
	reflectfLength    float32
	ownclsnscale      bool
	pushPriority      int32
	prevfallflag      bool
	dustOldPos        [3]float32
	dustTime          int
}

// Add a new char to the game
func newChar(n int, idx int32) (c *Char) {
	c = &Char{}
	c.init(n, idx)
	return c
}

func (c *Char) warn() string {
	return fmt.Sprintf("%v: WARNING: %v (%v) in state %v: ", sys.tickCount, c.name, c.id, c.ss.no)
}

func (c *Char) panic() {
	if sys.workingState != &c.ss.sb {
		sys.errLog.Panicf("%v\n%v\n%v\n%+v\n", c.gi().def, c.name,
			sys.cgi[sys.workingState.playerNo].def, sys.workingState)
	}
	sys.errLog.Panicf("%v\n%v\n%v\n%+v\n", c.gi().def, c.name,
		sys.cgi[c.ss.sb.playerNo].def, c.ss)
}

func (c *Char) init(n int, idx int32) {
	// Reset struct with defaults
	*c = Char{
		playerNo:      n,
		helperIndex:   idx,
		controller:    n,
		animPN:        n,
		id:            -1,
		runorder:      -1,
		parentIndex:   IErr,
		hoverIdx:      -1,
		mctype:        MC_Hit,
		ownpal:        true,
		facing:        1,
		minus:         2,
		winquote:      -1,
		clsnBaseScale: [2]float32{1, 1},
		clsnScaleMul:  [2]float32{1, 1},
		clsnScale:     [2]float32{1, 1},
		zScale:        1,
		aimg:          *newAfterImage(),
		CharSystemVar: CharSystemVar{
			superDefenseMul: 1.0,
			fallDefenseMul:  1.0,
			customDefense:   1.0,
			finalDefense:    1.0,
			projection:      Projection_Orthographic,
		},
	}

	// Set player or helper defaults
	if idx == 0 {
		c.playerFlag = true
		c.kovelocity = true
		c.keyctrl = [4]bool{true, true, true, true}
	} else {
		c.playerFlag = false
		c.kovelocity = false
		c.keyctrl = [4]bool{false, false, false, true}
	}

	// Set controller to CPU if applicable
	if n >= 0 && n < len(sys.aiLevel) && sys.aiLevel[n] != 0 {
		c.controller ^= -1
	}

	c.clearState()
}

func (c *Char) clearState() {
	c.ss.clear()
	c.hitdef.clear(c, c.localscl)
	c.ghv.clear(c)
	c.ghv.clearOff()
	c.mhv.clear()
	for i := range c.hitby {
		c.hitby[i].clear()
	}
	for i := range c.hover {
		c.hover[i].clear()
	}
	c.mctype = MC_Hit
	c.mctime = 0
	c.counterHit = false
	c.hitdefContact = false
	c.fallTime = 0
	c.dustTime = 0
}

func (c *Char) clsnOverlapTrigger(box1, pid, box2 int32) bool {
	getter := sys.playerID(pid)
	// Invalid getter ID
	if getter == nil {
		return false
	}
	return c.clsnCheck(getter, box1, box2, false, true)
}

func (c *Char) addChild(ch *Char) {
	for i, chi := range c.children {
		if chi == nil {
			c.children[i] = ch
			return
		}
	}
	c.children = append(c.children, ch)
}

// Clear EnemyNear and P2 lists. For instance when player positions change
// A new list will be built the next time the redirect is called
// In Mugen, EnemyNear is updated instantly when the character uses PosAdd, but "P2" is not
func (c *Char) enemyNearP2Clear() {
	c.enemyNearList = c.enemyNearList[:0]
	c.p2EnemyList = c.p2EnemyList[:0]
}

// Clear character variables upon a new round or creation of a new helper
func (c *Char) prepareNextRound() {
	c.sysVarRangeSet(0, math.MaxInt32, 0)
	c.sysFvarRangeSet(0, math.MaxInt32, 0)
	atk := float32(c.gi().data.attack) * c.ocd().attackRatio / 100
	c.CharSystemVar = CharSystemVar{
		bindToId:              -1,
		angleDrawScale:        [2]float32{1, 1},
		trans:                 TT_default,
		alpha:                 [2]int32{255, 0},
		sizeWidth:             [2]float32{c.baseWidthFront(), c.baseWidthBack()},
		sizeHeight:            [2]float32{c.baseHeightTop(), c.baseHeightBottom()},
		sizeDepth:             [2]float32{c.baseDepthTop(), c.baseDepthBottom()},
		attackMul:             [4]float32{atk, atk, atk, atk},
		fallDefenseMul:        1,
		superDefenseMul:       1,
		superDefenseMulBuffer: 1,
		customDefense:         1,
		finalDefense:          float64(c.gi().data.defence) / 100,
	}
	c.updateSizeBox()
	c.oldPos, c.interPos = c.pos, c.pos
	if c.helperIndex == 0 {
		if sys.roundsExisted[c.playerNo&1] > 0 {
			c.palfx.clear()
		} else {
			c.palfx = newPalFX()
		}
		if c.teamside == -1 {
			c.setSCF(SCF_standby)
		}
	} else {
		c.palfx = nil
	}
	c.aimg.timegap = -1
	c.enemyNearP2Clear()
	c.targets = c.targets[:0]
	c.cpucmd = -1
}

// Clear data when loading a new instance of the same character
func (c *Char) clearCachedData() {
	c.anim = nil
	c.animBackup = nil
	c.curFrame = nil
	c.hoverIdx = -1
	c.mctype, c.mctime = MC_Hit, 0
	c.counterHit = false
	c.fallTime = 0
	c.superDefenseMul = 1
	c.superDefenseMulBuffer = 1
	c.fallDefenseMul = 1
	c.customDefense = 1
	c.defenseMulDelay = false
	c.ownpal = true
	c.preserve = true // Just in case
	c.animPN = -1
	c.spritePN = -1
	c.animNo = 0
	c.prevAnimNo = 0
	c.stchtmp = false
	c.inguarddist = false
	c.p1facing = 0
	c.pushed = false
	c.atktmp, c.hittmp, c.acttmp, c.minus = 0, 0, 0, 2
	c.winquote = -1
	c.mapArray = make(map[string]float32)
	c.remapSpr = make(RemapPreset)
	c.gi().attackBase = c.gi().data.attack
	c.gi().defenceBase = c.gi().data.defence
}

// Return Char Global Info normally
func (c *Char) gi() *CharGlobalInfo {
	return &sys.cgi[c.playerNo]
}

// Return Char Global Info from the state owner
func (c *Char) stOgi() *CharGlobalInfo {
	return &sys.cgi[c.ss.sb.playerNo]
}

// Return Char Global Info according to working state
// Essentially check it in the character itself during negative states and in the state owner otherwise
// There was a bug in the default values of DefenceMulSet and Explod when a character threw another character with a different engine version
// This showed that engine version should always be checked in the player that owns the code
// So this function was added to replace stOgi() in version checks
// Version checks should probably be refactored in the future, regardless
func (c *Char) stWgi() *CharGlobalInfo {
	if c.minus == 0 {
		return &sys.cgi[c.ss.sb.playerNo]
	} else {
		return &sys.cgi[c.playerNo]
	}
}

func (c *Char) ocd() *OverrideCharData {
	team := c.teamside
	if c.teamside == -1 {
		team = 2
	}
	// This check prevents a crash when modifying helpers to be teamside 0
	// This happens because OverrideCharData is indexed by teamside
	// TODO: Perhaps ModifyPlayer or OverrideCharData could be refactored to not need this and be safer overall
	if c.memberNo < len(sys.sel.ocd[team]) {
		return &sys.sel.ocd[team][c.memberNo]
	}
	// Return default values as safeguard
	return newOverrideCharData()
}

func (c *Char) load(def string) error {
	gi := &sys.cgi[c.playerNo]
	gi.def, gi.displayname, gi.lifebarname, gi.author = def, "", "", ""
	gi.sff, gi.palettedata, gi.snd, gi.quotes = nil, nil, nil, [MaxQuotes]string{}
	gi.anim = NewAnimationTable()
	gi.fnt = [10]*Fnt{}
	for i := range gi.palkeymap {
		gi.palkeymap[i] = int32(i)
	}
	c.mapDefault = make(map[string]float32)
	// Helper to resolve paths relative to the .def file's logical location
	resolvePathRelativeToDef := func(pathInDefFile string) string {
		isZipDef, zipArchiveOfDef, defSubPathInZip := IsZipPath(gi.def)
		pathInDefFile = filepath.ToSlash(pathInDefFile)

		if filepath.IsAbs(pathInDefFile) {
			return pathInDefFile
		}
		isEngineRootRelative := strings.HasPrefix(pathInDefFile, "data/") ||
			strings.HasPrefix(pathInDefFile, "font/") ||
			strings.HasPrefix(pathInDefFile, "stages/")

		if isZipDef {
			if isEngineRootRelative {
				return pathInDefFile
			}
			baseDirWithinZip := filepath.ToSlash(filepath.Dir(defSubPathInZip))
			if baseDirWithinZip == "." || baseDirWithinZip == "" {
				return filepath.ToSlash(filepath.Join(zipArchiveOfDef, pathInDefFile))
			}
			return filepath.ToSlash(filepath.Join(zipArchiveOfDef, baseDirWithinZip, pathInDefFile))
		}
		return pathInDefFile
	}
	if err := c.loadFx(def); err != nil {
		sys.errLog.Printf("Error loading FX for %s: %v", def, err)
	}
	str, err := LoadText(def)
	if err != nil {
		return err
	}
	lines, i := SplitAndTrim(str, "\n"), 0
	cns, sprite, anim, sound := "", "", "", ""
	info, files, keymap, mapArray, lanInfo, lanFiles, lanKeymap, lanMapArray := true, true, true, true, true, true, true, true
	gi.localcoord = [...]float32{320, 240}
	c.localcoord = 320 / (float32(sys.gameWidth) / 320)
	c.localscl = 320 / c.localcoord
	gi.portraitscale = 1
	var fnt [10][2]string
	for i < len(lines) {
		is, name, subname := ReadIniSection(lines, &i)
		switch name {
		case "info":
			if info {
				info = false
				c.name, _, _ = is.getText("name")
				var ok bool
				if gi.displayname, ok, _ = is.getText("displayname"); !ok {
					gi.displayname = c.name
				}
				if gi.lifebarname, ok, _ = is.getText("lifebarname"); !ok {
					gi.lifebarname = gi.displayname
				}
				gi.author, _, _ = is.getText("author")
				gi.nameLow = strings.ToLower(c.name)
				gi.displaynameLow = strings.ToLower(gi.displayname)
				gi.authorLow = strings.ToLower(gi.author)
				if is.ReadF32("localcoord", &gi.localcoord[0], &gi.localcoord[1]) {
					gi.portraitscale = 320 / gi.localcoord[0]
					c.localcoord = gi.localcoord[0] / (float32(sys.gameWidth) / 320)
					c.localscl = 320 / c.localcoord
				}
				is.ReadF32("portraitscale", &gi.portraitscale)
			}
		case "files":
			if files {
				files = false
				cns = decodeShiftJIS(is["cns"])
				sprite = decodeShiftJIS(is["sprite"])
				anim = decodeShiftJIS(is["anim"])
				sound = decodeShiftJIS(is["sound"])
				for i := range gi.pal {
					gi.pal[i] = decodeShiftJIS(is[fmt.Sprintf("pal%v", i+1)])
				}
				for i := range fnt {
					fnt[i][0] = is[fmt.Sprintf("font%v", i)]
					fnt[i][1] = is[fmt.Sprintf("fnt_height%v", i)]
				}
			}
		case "palette ":
			if keymap &&
				len(subname) >= 6 && strings.ToLower(subname[:6]) == "keymap" {
				keymap = false
				for i, v := range [12]string{"a", "b", "c", "x", "y", "z",
					"a2", "b2", "c2", "x2", "y2", "z2"} {
					var i32 int32
					if is.ReadI32(v, &i32) {
						if i32 < 1 || int(i32) > len(gi.palkeymap) {
							i32 = 1
						}
						gi.palkeymap[i] = i32 - 1
					}
				}
			}
		case "map":
			if mapArray {
				mapArray = false
				for key, value := range is {
					c.mapDefault[key] = float32(Atof(value))
				}
			}
		case fmt.Sprintf("%v.info", sys.cfg.Config.Language):
			if lanInfo {
				info = false
				lanInfo = false
				c.name, _, _ = is.getText("name")
				var ok bool
				if gi.displayname, ok, _ = is.getText("displayname"); !ok {
					gi.displayname = c.name
				}
				if gi.lifebarname, ok, _ = is.getText("lifebarname"); !ok {
					gi.lifebarname = gi.displayname
				}
				gi.author, _, _ = is.getText("author")
				gi.nameLow = strings.ToLower(c.name)
				gi.displaynameLow = strings.ToLower(gi.displayname)
				gi.authorLow = strings.ToLower(gi.author)
				if is.ReadF32("localcoord", &gi.localcoord[0], &gi.localcoord[1]) {
					gi.portraitscale = 320 / gi.localcoord[0]
					c.localcoord = gi.localcoord[0] / (float32(sys.gameWidth) / 320)
					c.localscl = 320 / c.localcoord
				}
				is.ReadF32("portraitscale", &gi.portraitscale)
			}
		case fmt.Sprintf("%v.files", sys.cfg.Config.Language):
			if lanFiles {
				files = false
				lanFiles = false
				cns = decodeShiftJIS(is["cns"])
				sprite = decodeShiftJIS(is["sprite"])
				anim = decodeShiftJIS(is["anim"])
				sound = decodeShiftJIS(is["sound"])
				for i := range gi.pal {
					gi.pal[i] = decodeShiftJIS(is[fmt.Sprintf("pal%v", i+1)])
				}
				for i := range fnt {
					fnt[i][0] = is[fmt.Sprintf("font%v", i)]
					fnt[i][1] = is[fmt.Sprintf("fnt_height%v", i)]
				}
			}
		case fmt.Sprintf("%v.palette ", sys.cfg.Config.Language):
			if lanKeymap &&
				len(subname) >= 6 && strings.ToLower(subname[:6]) == "keymap" {
				lanKeymap = false
				keymap = false
				for i, v := range [12]string{"a", "b", "c", "x", "y", "z",
					"a2", "b2", "c2", "x2", "y2", "z2"} {
					var i32 int32
					if is.ReadI32(v, &i32) {
						if i32 < 1 || int(i32) > len(gi.palkeymap) {
							i32 = 1
						}
						gi.palkeymap[i] = i32 - 1
					}
				}
			}
		case fmt.Sprintf("%v.map", sys.cfg.Config.Language):
			if lanMapArray {
				mapArray = false
				lanMapArray = false
				for key, value := range is {
					c.mapDefault[key] = float32(Atof(value))
				}
			}
		}
	}

	gi.constants = make(map[string]float32)

	// Init default values to ensure we have these maps
	gi.constants["default.attack.lifetopowermul"] = 0.7
	gi.constants["super.attack.lifetopowermul"] = 0
	gi.constants["default.gethit.lifetopowermul"] = 0.6
	gi.constants["super.gethit.lifetopowermul"] = 0.6
	gi.constants["super.targetdefencemul"] = 1.5
	gi.constants["default.lifetoguardpointsmul"] = 1.5
	gi.constants["super.lifetoguardpointsmul"] = -0.33
	gi.constants["default.lifetodizzypointsmul"] = 1.8
	gi.constants["super.lifetodizzypointsmul"] = 0
	gi.constants["default.lifetoredlifemul"] = 0.75
	gi.constants["super.lifetoredlifemul"] = 0.75
	gi.constants["default.legacygamedistancespec"] = 0
	gi.constants["default.legacyfallyvelyaccel"] = 0
	//gi.constants["default.ignoredefeatedenemies"] = 0
	gi.constants["input.pauseonhitpause"] = 1
	gi.constants["input.fbflipenemydistance"] = -1

	for _, key := range SortedKeys(sys.cfg.Common.Const) {
		for _, v := range sys.cfg.Common.Const[key] {
			if err := LoadFile(&v, []string{def, sys.motifDir, sys.lifebar.def, "", "data/"}, func(filename string) error {
				str, err = LoadText(filename)
				if err != nil {
					return err
				}
				lines, i = SplitAndTrim(str, "\n"), 0
				is, _, _ := ReadIniSection(lines, &i)
				for key, value := range is {
					gi.constants[key] = float32(Atof(value))
				}
				return nil
			}); err != nil {
				return err
			}
		}
	}

	// Init constants
	// Correct engine default values to character's own localcoord
	gi.data.init()
	c.size.init()
	gi.attackBase = 100
	gi.defenceBase = 100

	coordRatio := float32(c.gi().localcoord[0]) / 320

	if coordRatio != 1 {
		for i := 0; i < 4; i++ {
			c.size.standbox[i] *= coordRatio
			c.size.crouchbox[i] *= coordRatio
			c.size.airbox[i] *= coordRatio
			c.size.downbox[i] *= coordRatio
		}
		c.size.attack.dist.width[0] *= coordRatio
		c.size.attack.dist.width[1] *= coordRatio
		c.size.attack.dist.height[0] *= coordRatio
		c.size.attack.dist.height[1] *= coordRatio
		c.size.attack.dist.depth[0] *= coordRatio
		c.size.attack.dist.depth[1] *= coordRatio
		c.size.proj.attack.dist.width[0] *= coordRatio
		c.size.proj.attack.dist.width[1] *= coordRatio
		c.size.proj.attack.dist.height[0] *= coordRatio
		c.size.proj.attack.dist.height[1] *= coordRatio
		c.size.proj.attack.dist.depth[0] *= coordRatio
		c.size.proj.attack.dist.depth[1] *= coordRatio
		c.size.head.pos[0] *= coordRatio
		c.size.head.pos[1] *= coordRatio
		c.size.mid.pos[0] *= coordRatio
		c.size.mid.pos[1] *= coordRatio
		c.size.shadowoffset *= coordRatio
		c.size.draw.offset[0] *= coordRatio
		c.size.draw.offset[1] *= coordRatio
		c.size.depth[0] *= coordRatio
		c.size.depth[1] *= coordRatio
		c.size.attack.depth[0] *= coordRatio
		c.size.attack.depth[1] *= coordRatio
	}

	gi.velocity.init()

	if coordRatio != 1 {
		gi.velocity.air.gethit.groundrecover[0] *= coordRatio
		gi.velocity.air.gethit.groundrecover[1] *= coordRatio
		gi.velocity.air.gethit.airrecover.add[0] *= coordRatio
		gi.velocity.air.gethit.airrecover.add[1] *= coordRatio
		gi.velocity.air.gethit.airrecover.back *= coordRatio
		gi.velocity.air.gethit.airrecover.fwd *= coordRatio
		gi.velocity.air.gethit.airrecover.up *= coordRatio
		gi.velocity.air.gethit.airrecover.down *= coordRatio

		gi.velocity.airjump.neu[0] *= coordRatio
		gi.velocity.airjump.neu[1] *= coordRatio
		gi.velocity.airjump.back *= coordRatio
		gi.velocity.airjump.fwd *= coordRatio

		gi.velocity.air.gethit.ko.add[0] *= coordRatio
		gi.velocity.air.gethit.ko.add[1] *= coordRatio
		gi.velocity.air.gethit.ko.ymin *= coordRatio
		gi.velocity.ground.gethit.ko.add[0] *= coordRatio
		gi.velocity.ground.gethit.ko.add[1] *= coordRatio
		gi.velocity.ground.gethit.ko.ymin *= coordRatio
	}

	gi.movement.init()

	if coordRatio != 1 {
		gi.movement.airjump.height *= coordRatio
		gi.movement.yaccel *= coordRatio
		gi.movement.stand.friction_threshold *= coordRatio
		gi.movement.crouch.friction_threshold *= coordRatio
		gi.movement.air.gethit.groundlevel *= coordRatio
		gi.movement.air.gethit.groundrecover.ground.threshold *= coordRatio
		gi.movement.air.gethit.groundrecover.groundlevel *= coordRatio
		gi.movement.air.gethit.airrecover.threshold *= coordRatio
		gi.movement.air.gethit.airrecover.yaccel *= coordRatio
		gi.movement.air.gethit.trip.groundlevel *= coordRatio
		gi.movement.down.bounce.offset[0] *= coordRatio
		gi.movement.down.bounce.offset[1] *= coordRatio
		gi.movement.down.bounce.yaccel *= coordRatio
		gi.movement.down.bounce.groundlevel *= coordRatio
		gi.movement.down.gethit.offset[0] *= coordRatio
		gi.movement.down.gethit.offset[1] *= coordRatio
		gi.movement.down.friction_threshold *= coordRatio
	}

	gi.remapPreset = make(map[string]RemapPreset)

	data, size, velocity, movement, quotes, lanQuotes, constants := true, true, true, true, true, true, true

	if len(cns) > 0 {
		cns_resolved := resolvePathRelativeToDef(cns)
		if err := LoadFile(&cns_resolved, []string{def, "", sys.motifDir, "data/"}, func(filename string) error {
			str, err := LoadText(filename)
			if err != nil {
				return err
			}
			lines, i = SplitAndTrim(str, "\n"), 0
			for i < len(lines) {
				is, name, subname := ReadIniSection(lines, &i)
				switch name {
				case "data":
					if data {
						data = false
						is.ReadI32("life", &gi.data.life)
						c.lifeMax = gi.data.life
						is.ReadI32("power", &gi.data.power)
						c.powerMax = gi.data.power
						gi.data.dizzypoints = c.lifeMax
						is.ReadI32("dizzypoints", &gi.data.dizzypoints)
						c.dizzyPointsMax = gi.data.dizzypoints
						gi.data.guardpoints = c.lifeMax
						is.ReadI32("guardpoints", &gi.data.guardpoints)
						c.guardPointsMax = gi.data.guardpoints
						is.ReadI32("attack", &gi.data.attack)
						gi.attackBase = gi.data.attack
						is.ReadI32("defence", &gi.data.defence)
						gi.defenceBase = gi.data.defence
						is.ReadI32("fall.defence_up", &gi.data.fall.defence_up)
						gi.data.fall.defence_mul = (float32(gi.data.fall.defence_up) + 100) / 100
						is.ReadI32("liedown.time", &gi.data.liedown.time)
						//gi.data.liedown.time = Max(1, gi.data.liedown.time) // Mugen doesn't actually handle it like this
						is.ReadI32("airjuggle", &gi.data.airjuggle)
						is.ReadI32("sparkno", &gi.data.sparkno)
						is.ReadI32("guard.sparkno", &gi.data.guard.sparkno)
						is.ReadI32("hitsound.channel", &gi.data.hitsound_channel)
						is.ReadI32("guardsound.channel", &gi.data.guardsound_channel)
						is.ReadI32("ko.echo", &gi.data.ko.echo)
						var i32 int32
						if is.ReadI32("volume", &i32) {
							gi.data.volume = i32/2 + 256
						}
						if is.ReadI32("volumescale", &i32) {
							gi.data.volume = i32 * 64 / 25
						}
						if _, ok := is["intpersistindex"]; ok {
							// We set these to 0 first in case the value turns out to be empty
							// https://github.com/ikemen-engine/Ikemen-GO/issues/2422
							gi.data.intpersistindex = 0
							is.ReadI32("intpersistindex", &gi.data.intpersistindex)
						}
						if _, ok := is["floatpersistindex"]; ok {
							gi.data.floatpersistindex = 0
							is.ReadI32("floatpersistindex", &gi.data.floatpersistindex)
						}
					}
				case "size":
					if size {
						size = false
						is.ReadF32("xscale", &c.size.xscale)
						is.ReadF32("yscale", &c.size.yscale)
						// Read legacy size constants first
						if is.ReadF32("ground.back", &c.size.standbox[0], &c.size.standbox[2]) {
							c.size.standbox[0] *= -1
						}
						if is.ReadF32("air.back", &c.size.airbox[0], &c.size.airbox[2]) {
							c.size.airbox[0] *= -1
						}
						if is.ReadF32("height", &c.size.standbox[1]) {
							c.size.standbox[1] *= -1
						}
						// Default boxes to the old constants we just read
						c.size.standbox = [4]float32{c.size.standbox[0], c.size.standbox[1], c.size.standbox[2], c.size.standbox[3]}
						c.size.crouchbox = c.size.standbox
						c.size.airbox = [4]float32{c.size.airbox[0], c.size.standbox[1], c.size.airbox[2], c.size.standbox[3]}
						c.size.downbox = c.size.standbox
						// Read new size constants to override them
						is.ReadF32("stand.sizebox", &c.size.standbox[0], &c.size.standbox[1], &c.size.standbox[2], &c.size.standbox[3])
						is.ReadF32("crouch.sizebox", &c.size.crouchbox[0], &c.size.crouchbox[1], &c.size.crouchbox[2], &c.size.crouchbox[3])
						is.ReadF32("air.sizebox", &c.size.airbox[0], &c.size.airbox[1], &c.size.airbox[2], &c.size.airbox[3])
						is.ReadF32("down.sizebox", &c.size.downbox[0], &c.size.downbox[1], &c.size.downbox[2], &c.size.downbox[3])
						/*
							is.ReadF32("height.stand", &c.size.height)
							// New height constants default to old height constant
							c.size.height.crouch = c.size.height
							c.size.height.air[0] = c.size.height
							c.size.height.down = c.size.height
							is.ReadF32("height.crouch", &c.size.height.crouch)
							is.ReadF32("height.air", &c.size.height.air[0], &c.size.height.air[1])
							is.ReadF32("height.down", &c.size.height.down)
						*/
						is.ReadF32("attack.dist", &c.size.attack.dist.width[0])
						is.ReadF32("attack.dist.width", &c.size.attack.dist.width[0], &c.size.attack.dist.width[1])
						is.ReadF32("attack.dist.height", &c.size.attack.dist.height[0], &c.size.attack.dist.height[1])
						is.ReadF32("attack.dist.depth", &c.size.attack.dist.depth[0], &c.size.attack.dist.depth[1])
						is.ReadF32("proj.attack.dist", &c.size.proj.attack.dist.width[0])
						is.ReadF32("proj.attack.dist.width", &c.size.proj.attack.dist.width[0], &c.size.proj.attack.dist.width[1])
						is.ReadF32("proj.attack.dist.height", &c.size.proj.attack.dist.height[0], &c.size.proj.attack.dist.height[1])
						is.ReadF32("proj.attack.dist.depth", &c.size.proj.attack.dist.depth[0], &c.size.proj.attack.dist.depth[1])
						is.ReadI32("proj.doscale", &c.size.proj.doscale)
						is.ReadF32("head.pos", &c.size.head.pos[0], &c.size.head.pos[1])
						is.ReadF32("mid.pos", &c.size.mid.pos[0], &c.size.mid.pos[1])
						is.ReadF32("shadowoffset", &c.size.shadowoffset)
						is.ReadF32("draw.offset", &c.size.draw.offset[0], &c.size.draw.offset[1])
						is.ReadF32("depth", &c.size.depth[0], &c.size.depth[1])
						is.ReadF32("attack.depth", &c.size.attack.depth[0], &c.size.attack.depth[1])
						is.ReadI32("weight", &c.size.weight)
						is.ReadF32("pushfactor", &c.size.pushfactor)
					}
				case "velocity":
					if velocity {
						velocity = false
						is.ReadF32("walk.fwd", &gi.velocity.walk.fwd)
						is.ReadF32("walk.back", &gi.velocity.walk.back)
						is.ReadF32("run.fwd", &gi.velocity.run.fwd[0], &gi.velocity.run.fwd[1])
						is.ReadF32("run.back",
							&gi.velocity.run.back[0], &gi.velocity.run.back[1])
						is.ReadF32("jump.neu",
							&gi.velocity.jump.neu[0], &gi.velocity.jump.neu[1])
						is.ReadF32("jump.back", &gi.velocity.jump.back)
						is.ReadF32("jump.fwd", &gi.velocity.jump.fwd)
						// Running and air jumps default to regular jump velocities
						c.gi().velocity.runjump.back[0] = c.gi().velocity.jump.back
						c.gi().velocity.runjump.back[1] = c.gi().velocity.jump.neu[1]
						c.gi().velocity.runjump.fwd[0] = c.gi().velocity.jump.fwd
						c.gi().velocity.runjump.fwd[1] = c.gi().velocity.jump.neu[1]
						c.gi().velocity.airjump.neu = c.gi().velocity.jump.neu
						c.gi().velocity.airjump.back = c.gi().velocity.jump.back
						c.gi().velocity.airjump.fwd = c.gi().velocity.jump.fwd
						is.ReadF32("runjump.back",
							&gi.velocity.runjump.back[0], &gi.velocity.runjump.back[1])
						is.ReadF32("runjump.fwd",
							&gi.velocity.runjump.fwd[0], &gi.velocity.runjump.fwd[1])
						is.ReadF32("airjump.neu",
							&gi.velocity.airjump.neu[0], &gi.velocity.airjump.neu[1])
						is.ReadF32("airjump.back", &gi.velocity.airjump.back)
						is.ReadF32("airjump.fwd", &gi.velocity.airjump.fwd)
						is.ReadF32("air.gethit.groundrecover",
							&gi.velocity.air.gethit.groundrecover[0],
							&gi.velocity.air.gethit.groundrecover[1])
						is.ReadF32("air.gethit.airrecover.mul",
							&gi.velocity.air.gethit.airrecover.mul[0],
							&gi.velocity.air.gethit.airrecover.mul[1])
						is.ReadF32("air.gethit.airrecover.add",
							&gi.velocity.air.gethit.airrecover.add[0],
							&gi.velocity.air.gethit.airrecover.add[1])
						is.ReadF32("air.gethit.airrecover.back",
							&gi.velocity.air.gethit.airrecover.back)
						is.ReadF32("air.gethit.airrecover.fwd",
							&gi.velocity.air.gethit.airrecover.fwd)
						is.ReadF32("air.gethit.airrecover.up",
							&gi.velocity.air.gethit.airrecover.up)
						is.ReadF32("air.gethit.airrecover.down",
							&gi.velocity.air.gethit.airrecover.down)
						is.ReadF32("air.gethit.ko.add", &gi.velocity.air.gethit.ko.add[0],
							&gi.velocity.air.gethit.ko.add[1])
						is.ReadF32("air.gethit.ko.ymin", &gi.velocity.air.gethit.ko.ymin)
						is.ReadF32("ground.gethit.ko.xmul", &gi.velocity.ground.gethit.ko.xmul)
						is.ReadF32("ground.gethit.ko.add", &gi.velocity.ground.gethit.ko.add[0],
							&gi.velocity.ground.gethit.ko.add[1])
						is.ReadF32("ground.gethit.ko.ymin", &gi.velocity.ground.gethit.ko.ymin)

						// Mugen accepts these but they are not documented. Possible leftovers of Z axis implementation
						// In Ikemen we're making them accept 3 values each, for the 3 axes
						is.ReadF32("walk.up", &gi.velocity.walk.up[0], &gi.velocity.walk.up[1], &gi.velocity.walk.up[2])
						is.ReadF32("walk.down", &gi.velocity.walk.down[0], &gi.velocity.walk.down[1], &gi.velocity.walk.down[2])
						is.ReadF32("run.up", &gi.velocity.run.up[0], &gi.velocity.run.up[1], &gi.velocity.run.up[2])
						is.ReadF32("run.down", &gi.velocity.run.down[0], &gi.velocity.run.down[1], &gi.velocity.run.down[2])
						is.ReadF32("jump.up", &gi.velocity.jump.up[0], &gi.velocity.jump.up[1], &gi.velocity.jump.up[2])
						is.ReadF32("jump.down", &gi.velocity.jump.down[0], &gi.velocity.jump.down[1], &gi.velocity.jump.down[2])
						is.ReadF32("runjump.up", &gi.velocity.runjump.up[0], &gi.velocity.runjump.up[1], &gi.velocity.runjump.up[2])
						is.ReadF32("runjump.down", &gi.velocity.runjump.down[0], &gi.velocity.runjump.down[1], &gi.velocity.runjump.down[2])
						is.ReadF32("airjump.up", &gi.velocity.airjump.up[0], &gi.velocity.airjump.up[1], &gi.velocity.airjump.up[2])
						is.ReadF32("airjump.down", &gi.velocity.airjump.down[0], &gi.velocity.airjump.down[1], &gi.velocity.airjump.down[2])
					}
				case "movement":
					if movement {
						movement = false
						is.ReadI32("airjump.num", &gi.movement.airjump.num)
						is.ReadF32("airjump.height", &gi.movement.airjump.height)
						is.ReadF32("yaccel", &gi.movement.yaccel)
						is.ReadF32("stand.friction", &gi.movement.stand.friction)
						is.ReadF32("stand.friction.threshold",
							&gi.movement.stand.friction_threshold)
						is.ReadF32("crouch.friction", &gi.movement.crouch.friction)
						is.ReadF32("crouch.friction.threshold",
							&gi.movement.crouch.friction_threshold)
						is.ReadF32("air.gethit.groundlevel",
							&gi.movement.air.gethit.groundlevel)
						is.ReadF32("air.gethit.groundrecover.ground.threshold",
							&gi.movement.air.gethit.groundrecover.ground.threshold)
						is.ReadF32("air.gethit.groundrecover.groundlevel",
							&gi.movement.air.gethit.groundrecover.groundlevel)
						is.ReadF32("air.gethit.airrecover.threshold",
							&gi.movement.air.gethit.airrecover.threshold)
						is.ReadF32("air.gethit.airrecover.yaccel",
							&gi.movement.air.gethit.airrecover.yaccel)
						is.ReadF32("air.gethit.trip.groundlevel",
							&gi.movement.air.gethit.trip.groundlevel)
						is.ReadF32("down.bounce.offset",
							&gi.movement.down.bounce.offset[0],
							&gi.movement.down.bounce.offset[1])
						is.ReadF32("down.bounce.yaccel", &gi.movement.down.bounce.yaccel)
						is.ReadF32("down.bounce.groundlevel",
							&gi.movement.down.bounce.groundlevel)
						is.ReadF32("down.friction.threshold",
							&gi.movement.down.friction_threshold)
						is.ReadF32("down.gethit.offset",
							&gi.movement.down.gethit.offset[0],
							&gi.movement.down.gethit.offset[1])
					}
				case "quotes":
					if quotes {
						quotes = false
						for i := range gi.quotes {
							if is[fmt.Sprintf("victory%v", i)] != "" {
								victoryQuotes, _, _ := is.getText(fmt.Sprintf("victory%v", i))
								gi.quotes[i] = decodeShiftJIS(victoryQuotes)
							}
						}
					}
				case fmt.Sprintf("%v.quotes", sys.cfg.Config.Language):
					if lanQuotes {
						quotes = false
						lanQuotes = false
						for i := range gi.quotes {
							if is[fmt.Sprintf("victory%v", i)] != "" {
								victoryQuotes, _, _ := is.getText(fmt.Sprintf("victory%v", i))
								gi.quotes[i] = decodeShiftJIS(victoryQuotes)
							}
						}
					}
				case "constants":
					if constants {
						constants = false
						for key, value := range is {
							gi.constants[key] = float32(Atof(value))
						}
					}
				case "remappreset ":
					if len(subname) >= 1 {
						if _, ok := gi.remapPreset[subname]; !ok {
							gi.remapPreset[subname] = make(RemapPreset)
						}
						for key := range is {
							k := strings.Split(key, ",")
							if len(k) == 2 {
								var v [2]int32
								is.ReadI32(key, &v[0], &v[1])
								if _, ok := gi.remapPreset[subname][int16(Atoi(k[0]))]; !ok {
									gi.remapPreset[subname][int16(Atoi(k[0]))] = make(RemapTable)
								}
								gi.remapPreset[subname][int16(Atoi(k[0]))][int16(Atoi(k[1]))] = [...]int16{int16(v[0]), int16(v[1])}
							}
						}
					}
				}
			}
			return nil
		}); err != nil {
			return err
		}
	}

	if len(sprite) > 0 {
		sprite_resolved := resolvePathRelativeToDef(sprite)
		if err := LoadFile(&sprite_resolved, []string{gi.def, "", sys.motifDir, "data/"}, func(filename string) error {
			var err_sff error
			gi.sff, err_sff = loadSff(filename, true) // loadSff uses OpenFile
			return err_sff
		}); err != nil {
			return err
		}
	} else {
		gi.sff = newSff()
	}
	gi.palettedata = newPaldata()
	gi.palettedata.palList = PaletteList{
		palettes:   append([][]uint32{}, gi.sff.palList.palettes...),
		paletteMap: append([]int{}, gi.sff.palList.paletteMap...),
		PalTable:   make(map[[2]int16]int),
		numcols:    make(map[[2]int16]int),
		PalTex:     append([]Texture{}, gi.sff.palList.PalTex...),
	}
	for key, value := range gi.sff.palList.PalTable {
		gi.palettedata.palList.PalTable[key] = value
	}
	for key, value := range gi.sff.palList.numcols {
		gi.palettedata.palList.numcols[key] = value
	}
	str = ""
	if len(anim) > 0 {
		anim_resolved := resolvePathRelativeToDef(anim)
		if LoadFile(&anim_resolved, []string{def, "", sys.motifDir, "data/"}, func(filename string) error {
			var err_air error
			str, err_air = LoadText(filename)
			if err_air != nil {
				return err_air
			}
			return nil
		}); err != nil {
			return err
		}
	}
	for _, key := range SortedKeys(sys.cfg.Common.Air) {
		for _, v := range sys.cfg.Common.Air[key] {
			if err := LoadFile(&v, []string{def, sys.motifDir, sys.lifebar.def, "", "data/"}, func(filename string) error {
				txt, err := LoadText(filename)
				if err != nil {
					return err
				}
				str += "\n" + txt
				return nil
			}); err != nil {
				return err
			}
		}
	}
	lines, i = SplitAndTrim(str, "\n"), 0
	gi.anim = ReadAnimationTable(gi.sff, &gi.palettedata.palList, lines, &i)
	if len(sound) > 0 {
		sound_resolved := resolvePathRelativeToDef(sound)
		if LoadFile(&sound_resolved, []string{def, "", sys.motifDir, "data/"}, func(filename string) error {
			var err error
			gi.snd, err = LoadSnd(filename)
			return err
		}); err != nil {
			return err
		}
	} else {
		gi.snd = newSnd()
	}
	// Load fonts
	for i_fnt, f_fnt_pair := range fnt {
		if len(f_fnt_pair[0]) > 0 {
			resolvedFntPath := resolvePathRelativeToDef(f_fnt_pair[0])
			i := i_fnt
			f_pair := f_fnt_pair
			LoadFile(&resolvedFntPath, []string{def, sys.motifDir, "", "data/", "font/"}, func(filename string) error {
				// Defer the font loading to the main thread
				sys.mainThreadTask <- func() {
					var err error
					var height int32 = -1
					if len(f_pair[1]) > 0 {
						height = Atoi(f_pair[1])
					}
					if gi.fnt[i], err = loadFnt(filename, height); err != nil {
						sys.errLog.Printf("failed to load %v (char font): %v", filename, err)
						// Assign a new empty font on failure to prevent nil pointer panics
						gi.fnt[i] = newFnt()
					}
				}
				return nil
			})
		}
	}
	return nil
}

func (c *Char) loadPalette() {
	gi := c.gi()
	if gi.sff.header.Ver0 == 1 {
		gi.palettedata.palList.ResetRemap()
		tmp := 0
		for i := 0; i < MaxPalNo; i++ {
			pl := gi.palettedata.palList.Get(i)
			var f io.ReadSeekCloser
			var err error
			if LoadFile(&gi.pal[i], []string{gi.def, "", sys.motifDir, "data/"}, func(file string) error {
				f, err = OpenFile(file)
				return err
			}) == nil {
				for i := 255; i >= 0; i-- {
					var rgb [3]byte
					if _, err = io.ReadFull(f, rgb[:]); err != nil {
						break
					}
					var alpha byte = 255
					if i == 0 {
						alpha = 0
					}
					pl[i] = uint32(alpha)<<24 | uint32(rgb[2])<<16 | uint32(rgb[1])<<8 | uint32(rgb[0])
				}
				chk(f.Close())
				if err == nil {
					if tmp == 0 && i > 0 {
						copy(gi.palettedata.palList.Get(0), pl)
					}
					gi.palExist[i] = true
					// Palette Texture Generation
					gi.palettedata.palList.PalTex[i] = PaletteToTexture(pl)
					tmp = i + 1
				}
			} else if f != nil {
				chk(f.Close())
			}
			if err != nil {
				gi.palExist[i] = false
				if i > 0 {
					delete(gi.palettedata.palList.PalTable, [...]int16{1, int16(i + 1)})
				}
			}
		}
		if tmp == 0 {
			delete(gi.palettedata.palList.PalTable, [...]int16{1, 1})
		}
	} else {
		for i := 0; i < MaxPalNo; i++ {
			_, gi.palExist[i] =
				gi.palettedata.palList.PalTable[[...]int16{1, int16(i + 1)}]
		}
		if gi.sff.header.NumberOfPalettes > 0 {
			numPals := int(gi.sff.header.NumberOfPalettes)
			if len(gi.palettedata.palList.PalTex) < numPals {
				gi.palettedata.palList.PalTex = make([]Texture, numPals)
			}
			for i := 0; i < numPals; i++ {
				pal := gi.sff.palList.Get(i)
				if pal != nil {
					gi.palettedata.palList.PalTex[i] = PaletteToTexture(pal)
				}
			}
		}
	}
	for i := range gi.palSelectable {
		gi.palSelectable[i] = false
	}
	for i := 0; i < MaxPalNo; i++ {
		startj := gi.palkeymap[i]
		if !gi.palExist[startj] {
			startj %= 6
		}
		j := startj
		for {
			if gi.palExist[j] {
				gi.palSelectable[j] = true
				break
			}
			j++
			if j >= MaxPalNo {
				j = 0
			}
			if j == startj {
				break
			}
		}
	}
	palIdx := gi.palno - 1
	if palIdx < 0 || palIdx >= int32(len(gi.palExist)) || !gi.palExist[palIdx] {
		found := false
		for i := 0; i < len(gi.palExist); i++ {
			if gi.palExist[i] {
				gi.palno = int32(i + 1)
				found = true
				break
			}
		}
		if !found {
			gi.palno = 1
			gi.palExist[0] = true
			gi.palSelectable[0] = true
		}
	}

	gi.remappedpal = [2]int32{1, gi.palno}
}
func (c *Char) loadFx(def string) error {
	gi := c.gi()
	gi.fxPath = []string{} // Always initialize before loading.

	charDefContent, err := LoadText(def)
	if err != nil {
		return err
	}

	// Helper function to resolve paths referenced inside the .def file.
	resolvePathRelativeToDef := func(pathInDefFile string) string {
		isZipDef, zipArchiveOfDef, defSubPathInZip := IsZipPath(def)
		pathInDefFile = filepath.ToSlash(pathInDefFile)
		if filepath.IsAbs(pathInDefFile) {
			return pathInDefFile
		}
		if isZipRel, _, _ := IsZipPath(pathInDefFile); isZipRel {
			return pathInDefFile
		}
		isEngineRootRelative := strings.HasPrefix(pathInDefFile, "data/") || strings.HasPrefix(pathInDefFile, "font/") || strings.HasPrefix(pathInDefFile, "stages/")
		if isZipDef {
			if isEngineRootRelative {
				return pathInDefFile
			}
			baseDirWithinZip := filepath.ToSlash(filepath.Dir(defSubPathInZip))
			if baseDirWithinZip == "." || baseDirWithinZip == "" {
				return filepath.ToSlash(filepath.Join(zipArchiveOfDef, pathInDefFile))
			}
			return filepath.ToSlash(filepath.Join(zipArchiveOfDef, baseDirWithinZip, pathInDefFile))
		}
		return pathInDefFile
	}

	lines, i := SplitAndTrim(charDefContent, "\n"), 0
	info, files, lanInfo, lanFiles := true, true, true, true

	for i < len(lines) {
		isec, name, _ := ReadIniSection(lines, &i)
		switch name {
		case "info":
			if info {
				info = false
				fightfxPrefixName, _, _ := isec.getText("fightfx.prefix")
				gi.fightfxPrefix = strings.ToLower(fightfxPrefixName)
			}
		case fmt.Sprintf("%v.info", sys.cfg.Config.Language):
			if lanInfo {
				info = false
				lanInfo = false
				fightfxPrefixName, _, _ := isec.getText("fightfx.prefix")
				gi.fightfxPrefix = strings.ToLower(fightfxPrefixName)
			}
		case "files":
			if files {
				files = false
				if fx_paths_str, ok := isec["fx"]; ok {
					for _, fx_path := range strings.Split(fx_paths_str, ",") {
						fx_path = strings.TrimSpace(fx_path)
						if fx_path == "" {
							continue
						}
						resolved_path := resolvePathRelativeToDef(fx_path)

						if found_path := FileExist(resolved_path); found_path != "" {
							if err := loadFightFx(found_path, false); err != nil {
								sys.errLog.Printf("Could not load CommonFX %s for char %s: %v", found_path, def, err)
							} else {
								gi.fxPath = append(gi.fxPath, found_path)
							}
						} else {
							if found_path_fallback := SearchFile(fx_path, []string{def, "", sys.motifDir, "data/"}); found_path_fallback != "" {
								if err := loadFightFx(found_path_fallback, false); err != nil {
									sys.errLog.Printf("Could not load CommonFX %s for char %s: %v", found_path_fallback, def, err)
								} else {
									gi.fxPath = append(gi.fxPath, found_path_fallback)
								}
							} else {
								sys.errLog.Printf("CommonFX file not found for char %s: %s (resolved to %s)", def, fx_path, resolved_path)
							}
						}
					}
				}
			}
		case fmt.Sprintf("%v.files", sys.cfg.Config.Language):
			if lanFiles {
				files = false
				lanFiles = false
				if fx_paths_str, ok := isec["fx"]; ok {
					for _, fx_path := range strings.Split(fx_paths_str, ",") {
						fx_path = strings.TrimSpace(fx_path)
						if fx_path == "" {
							continue
						}
						resolved_fx_path := resolvePathRelativeToDef(fx_path)
						if resolved_fx_path != "" {
							if err := loadFightFx(resolved_fx_path, false); err != nil {
								sys.errLog.Printf("Could not load CommonFX %s for char %s: %v", resolved_fx_path, def, err)
							} else {
								gi.fxPath = append(gi.fxPath, resolved_fx_path)
							}
						}
					}
				}
			}
		}
	}
	return nil
}
func (c *Char) clearHitCount() {
	c.hitCount = 0
	c.uniqHitCount = 0
	c.guardCount = 0
}

func (c *Char) clearMoveHit() {
	c.mctime = 0
	c.counterHit = false
}

func (c *Char) clearHitDef() {
	c.hitdef.clear(c, c.localscl)
}

func (c *Char) changeAnimEx(animNo int32, animPlayerNo int, spritePlayerNo int, ffx string) {
	// Get the animation
	a := c.getAnimSprite(animNo, animPlayerNo, spritePlayerNo, ffx, c.ownpal, false)

	// If invalid
	if a == nil {
		return
	}

	// Assign animation to character
	c.anim = a
	c.anim.remap = c.remapSpr
	c.prevAnimNo = c.animNo
	c.animNo = animNo

	// Animation is valid, so we update these variables
	// Common FX set playerNo to undefined
	if ffx != "" && ffx != "s" {
		c.animPN = -1
		c.spritePN = -1
	} else {
		c.animPN = animPlayerNo
		c.spritePN = spritePlayerNo
	}

	// Update animation local scale
	animOwner := c.animPN
	if animOwner < 0 || animOwner >= len(sys.chars) || len(sys.chars[animOwner]) == 0 {
		animOwner = c.playerNo
	}
	c.animlocalscl = 320 / sys.chars[animOwner][0].localcoord

	// Clsn scale depends on the animation owner's scale, so it must be updated
	c.updateClsnScale()
	// Update reference frame
	c.updateCurFrame()
}

func (c *Char) changeAnim(animNo int32, animPlayerNo int, spritePlayerNo int, ffx string) {
	if animNo < 0 && animNo != -2 {
		// MUGEN 1.1 exports a warning message when attempting to change anim to a negative value through ChangeAnim SCTRL,
		// then sets the character animation to "0". Ikemen GO uses "-2" as a no-sprite/invisible anim, so we make
		// an exception here
		sys.appendToConsole(c.warn() + fmt.Sprintf("attempted change to negative anim (different from -2)"))
		animNo = 0
	}

	// Validate AnimPlayerNo
	if animPlayerNo < 0 {
		animPlayerNo = c.playerNo
	} else if animPlayerNo >= len(sys.chars) || len(sys.chars[animPlayerNo]) == 0 {
		sys.appendToConsole(c.warn() + fmt.Sprintf("Invalid animPlayerNo: %v", animPlayerNo+1))
		animPlayerNo = c.playerNo
	}

	// Validate SpritePlayerNo
	if spritePlayerNo < 0 {
		spritePlayerNo = c.playerNo
	} else if spritePlayerNo >= len(sys.chars) || len(sys.chars[spritePlayerNo]) == 0 {
		sys.appendToConsole(c.warn() + fmt.Sprintf("Invalid spritePlayerNo: %v", spritePlayerNo+1))
		spritePlayerNo = c.playerNo
	}

	c.changeAnimEx(animNo, animPlayerNo, spritePlayerNo, ffx)
}

func (c *Char) changeAnim2(animNo int32, animPlayerNo int, ffx string) {
	if animNo < 0 && animNo != -2 {
		sys.appendToConsole(c.warn() + fmt.Sprintf("attempted change to negative anim (different from -2)"))
		animNo = 0
	}

	c.changeAnimEx(animNo, animPlayerNo, c.playerNo, ffx)
}

func (c *Char) setAnimElem(elem, elemtime int32) {
	if c.anim == nil {
		return
	}

	// These parameters are already validated in anim.SetAnimElem,
	// but since we must check for error messages we might as well validate them here too

	// Validate elem
	if elem < 1 || int(elem) > len(c.anim.frames) {
		sys.appendToConsole(c.warn() + fmt.Sprintf("changed to invalid animelem %v within action %v", elem, c.animNo))
		elem = 1
		elemtime = 0
	} else if elemtime != 0 {
		// Validate elemtime only if it's unusual and elem is valid
		frametime := c.anim.frames[elem-1].Time
		if elemtime < 0 || (frametime != -1 && elemtime >= frametime) {
			sys.appendToConsole(c.warn() + fmt.Sprintf("changed to invalid elemtime %v in animelem %v", elemtime, elem))
			elemtime = 0
		}
	}

	// Set them
	c.anim.SetAnimElem(elem, elemtime)
	c.updateCurFrame()
}

/*
func (c *Char) validatePlayerNo(pn int, pname, scname string) bool {
	valid := pn >= 0 && pn < len(sys.chars) &&
		len(sys.chars[pn]) > 0 && sys.chars[pn][0] != nil
	if !valid {
		sys.appendToConsole(c.warn() + fmt.Sprintf("Invalid %s for %s: %v", pname, scname, pn+1))
		return false
	}
	return true
}
*/

func (c *Char) setCtrl(ctrl bool) {
	if ctrl {
		c.setSCF(SCF_ctrl)
	} else {
		c.unsetSCF(SCF_ctrl)
	}
}

func (c *Char) setDizzy(set bool) {
	if set {
		c.setSCF(SCF_dizzy)
	} else {
		c.unsetSCF(SCF_dizzy)
	}
}

func (c *Char) setGuardBreak(set bool) {
	if set {
		c.setSCF(SCF_guardbreak)
	} else {
		c.unsetSCF(SCF_guardbreak)
	}
}

func (c *Char) scf(scf SystemCharFlag) bool {
	return c.systemFlag&scf != 0
}

func (c *Char) setSCF(scf SystemCharFlag) {
	c.systemFlag |= scf
	// Clear enemy lists if changing flags that affect them
	if c.playerFlag && (scf == SCF_disabled || scf == SCF_over_ko || scf == SCF_standby) {
		sys.charList.enemyNearChanged = true
	}
}

func (c *Char) unsetSCF(scf SystemCharFlag) {
	c.systemFlag &^= scf
	// Clear enemy lists if changing flags that affect them
	if c.playerFlag && (scf == SCF_disabled || scf == SCF_over_ko || scf == SCF_standby) {
		sys.charList.enemyNearChanged = true
	}
}

func (c *Char) csf(csf CharSpecialFlag) bool {
	return c.specialFlag&csf != 0
}

func (c *Char) setCSF(csf CharSpecialFlag) {
	c.specialFlag |= csf
}

func (c *Char) unsetCSF(csf CharSpecialFlag) {
	c.specialFlag &^= csf
}

func (c *Char) asf(asf AssertSpecialFlag) bool {
	return c.assertFlag&asf != 0
}

func (c *Char) setASF(asf AssertSpecialFlag) {
	c.assertFlag |= asf
}

func (c *Char) unsetASF(asf AssertSpecialFlag) {
	c.assertFlag &^= asf
}

func (c *Char) parent(log bool) *Char {
	if c.parentIndex == IErr {
		if log {
			sys.appendToConsole(c.warn() + "has no parent")
		}
		return nil
	}

	// In Mugen, after the original parent has been destroyed, "parent" can still be valid if a new helper ends up occupying the same slot
	// That is undesirable behavior however, and is probably only used by exploit characters, which already don't work correctly anyway
	if c.parentIndex < 0 {
		if log {
			sys.appendToConsole(c.warn() + "parent has already been destroyed")
			if !sys.ignoreMostErrors {
				sys.errLog.Println(c.name + " parent has already been destroyed")
			}
		}
		return nil
	}

	return sys.chars[c.playerNo][c.parentIndex]
}

func (c *Char) root(log bool) *Char {
	if c.helperIndex == 0 {
		if log {
			sys.appendToConsole(c.warn() + "has no root")
		}
		return nil
	}

	return sys.chars[c.playerNo][0]
}

func (c *Char) helperTrigger(id int32, idx int) *Char {
	// Invalid index
	if idx < 0 {
		sys.appendToConsole(c.warn() + "helper redirection index cannot be negative")
		return nil
	}

	var count int
	for _, h := range sys.charList.runOrder {
		// Skip roots, helpers from other players and destroyed helpers
		// Mugen confirmed to skip helpers under DestroySelf in the same frame
		if h.helperIndex == 0 || h.playerNo != c.playerNo || h.csf(CSF_destroy) {
			continue
		}

		// Skip if helper ID doesn't match (except id <= 0, which matches any)
		if id > 0 && h.helperId != id {
			continue
		}

		// Found a valid helper
		if count == idx {
			return h
		}

		count++
	}

	// No valid helper found
	sys.appendToConsole(c.warn() + fmt.Sprintf("has no helper with ID %v and index %v", id, idx))
	return nil
}

func (c *Char) helperIndexTrigger(n int32, log bool) *Char {
	if n <= 0 {
		return c
	}
	return sys.charList.getHelperIndex(c, n, log)
}

func (c *Char) helperByIndexExist(id BytecodeValue) BytecodeValue {
	if id.IsSF() {
		return BytecodeSF()
	}
	return BytecodeBool(c.helperIndexTrigger(id.ToI(), false) != nil)
}

func (c *Char) indexTrigger() int32 {
	// Ignore destroyed helpers for the sake of consistency
	var searchIdx int32
	for _, p := range sys.charList.runOrder {
		if p != nil && !p.csf(CSF_destroy) {
			if c == p {
				return searchIdx
			}
			searchIdx++
		}
	}

	//for i, p := range sys.charList.runOrder {
	//	if c == p {
	//		return int32(i)
	//	}
	//}

	return -1
}

// Target redirection
func (c *Char) targetTrigger(id int32, idx int) *Char {
	// Invalid index
	if idx < 0 {
		sys.appendToConsole(c.warn() + "target redirection index cannot be negative")
		return nil
	}

	// Filter targets with the specified ID
	var filteredTargets []*Char
	for _, tid := range c.targets {
		if t := sys.playerID(tid); t != nil && (id < 0 || id == t.ghv.hitid) {
			filteredTargets = append(filteredTargets, t)
			// Target found at requested index
			if idx >= 0 && len(filteredTargets) == idx+1 {
				return filteredTargets[idx]
			}
		}
	}

	// No valid target found
	sys.appendToConsole(c.warn() + fmt.Sprintf("has no target with hit ID %v and index %v", id, idx))
	return nil
}

func (c *Char) partner(n int32, log bool) *Char {
	n = Max(0, n)
	if int(n) > len(sys.chars)/2-2 {
		if log {
			sys.appendToConsole(c.warn() + fmt.Sprintf("has no partner: %v", n))
		}
		return nil
	}
	// X>>1 = X/2
	// X<<1 = X*2
	// X&1 = X%2
	var p int
	if int(n) == c.playerNo>>1 {
		p = c.playerNo + 2
	} else {
		p = c.playerNo&1 + int(n)<<1
		if int(n) > c.playerNo>>1 {
			p += 2
		}
	}
	if len(sys.chars[p]) > 0 && sys.chars[p][0].teamside != -1 {
		return sys.chars[p][0]
	}
	if log {
		sys.appendToConsole(c.warn() + fmt.Sprintf("has no partner: %v", n))
	}
	return nil
}

func (c *Char) partnerTag(n int32) *Char {
	n = Max(0, n)
	if int(n) > len(sys.chars)/2-2 {
		return nil
	}
	var p int = (c.playerNo + int(n)<<1) + 2
	if p>>1 > int(c.numPartner()) {
		p -= int(c.numPartner()*2) + 2
	}
	if len(sys.chars[p]) > 0 && sys.chars[p][0].teamside != -1 {
		return sys.chars[p][0]
	}
	return nil
}

func (c *Char) enemy(n int32) *Char {
	if n < 0 {
		sys.appendToConsole(c.warn() + fmt.Sprintf("has no enemy with index %v", n))
		return nil
	}

	// Iterate until nth enemy
	var count int32
	for _, e := range sys.chars {
		if len(e) > 0 && e[0] != nil && c.isEnemyOf(e[0]) {
			if count == n {
				return e[0]
			}
			count++
		}
	}

	// No enemy found
	sys.appendToConsole(c.warn() + fmt.Sprintf("has no enemy with index %v", n))
	return nil
}

// This is only used to simplify the redirection call
func (c *Char) enemyNearTrigger(n int32) *Char {
	return sys.charList.enemyNear(c, n, false, true)
}

// Get the "P2" enemy reference
func (c *Char) p2() *Char {
	p := sys.charList.enemyNear(c, 0, true, false)
	// Cache last valid P2 enemy
	// Mugen seems to do this for the sake of auto turning before win poses
	if p != nil {
		c.p2EnemyBackup = p
	}
	return p
}

// Checks if a player should be considered an enemy at all for the "Enemy" and "P2" triggers, before filtering them further
func (c *Char) isEnemyOf(e *Char) bool {
	// Disabled players
	if e.scf(SCF_disabled) {
		return false
	}
	// Neutral players or partners
	if e.teamside < 0 || e.teamside == c.teamside {
		return false
	}
	// Standby enemies
	if e.scf(SCF_standby) {
		return false
	}
	// KO special ignore flag
	// TODO: This flag is obsolete in Tag. For Simul, we could either re-enable it or tag out KO players via ZSS
	//if sys.roundState() == 2 && e.scf(SCF_over_ko) && c.gi().constants["default.ignoredefeatedenemies"] != 0 {
	//	return false
	//}
	// Else a valid enemy
	return true
}

// Returns AI level as a float. Is truncated for AIlevel trigger, or not for AIlevelF
func (c *Char) getAILevel() float32 {
	if c.helperIndex != 0 && c.gi().mugenver[0] == 1 {
		return 0
	}
	if c.playerNo >= 0 && int(c.playerNo) < len(sys.aiLevel) {
		return sys.aiLevel[c.playerNo]
	}
	return 0
}

func (c *Char) setAILevel(level float32) {
	if c.playerNo < 0 || c.playerNo >= len(sys.aiLevel) {
		return
	}
	sys.aiLevel[c.playerNo] = level
	for _, c := range sys.chars[c.playerNo] {
		if level == 0 {
			c.controller = c.playerNo
		} else {
			c.controller = ^c.playerNo
		}
	}
}

func (c *Char) alive() bool {
	return !c.scf(SCF_ko)
}

func (c *Char) animElemNo(time int32) BytecodeValue {
	if c.anim != nil && time >= -c.anim.curtime {
		return BytecodeInt(c.anim.AnimElemNo(time))
	}
	return BytecodeSF()
}

func (c *Char) animElemTime(elem int32) BytecodeValue {
	if elem >= 1 && c.anim != nil && int(elem) <= len(c.anim.frames) {
		return BytecodeInt(c.anim.AnimElemTime(elem))
	}
	return BytecodeSF()
}

func (c *Char) animExist(wc *Char, anim BytecodeValue) BytecodeValue {
	if anim.IsSF() {
		return BytecodeSF()
	}
	if c != wc {
		return c.selfAnimExist(anim)
	}
	return sys.chars[c.ss.sb.playerNo][0].selfAnimExist(anim)
}

func (c *Char) animTime() int32 {
	if c.anim != nil {
		return c.anim.AnimTime()
	}
	return 0
}

// Update reference animation frame
func (c *Char) updateCurFrame() {
	if c.anim != nil {
		c.curFrame = c.anim.CurrentFrame()
	} else {
		c.curFrame = nil
	}
}

func (c *Char) backEdge() float32 {
	if c.facing < 0 {
		return c.rightEdge()
	}
	return c.leftEdge()
}

func (c *Char) backEdgeBodyDist() float32 {
	// In Mugen, edge body distance is changed when the character is in statetype A or L
	// This is undocumented and doesn't seem to offer any benefit
	offset := float32(0)
	if c.stWgi().ikemenver[0] == 0 && c.stWgi().ikemenver[1] == 0 {
		if c.ss.stateType == ST_A {
			offset = 0.5 / c.localscl
		} else if c.ss.stateType == ST_L {
			offset = 1.0 / c.localscl
		}
	}
	return c.backEdgeDist() - c.edgeWidth[1] - offset
}

func (c *Char) backEdgeDist() float32 {
	if c.facing < 0 {
		return sys.xmax/c.localscl - c.pos[0]
	}
	return c.pos[0] - sys.xmin/c.localscl
}

func (c *Char) bottomEdge() float32 {
	return sys.cam.ScreenPos[1]/c.localscl + c.gameHeight()
}

func (c *Char) botBoundBodyDist() float32 {
	return c.botBoundDist() - c.edgeDepth[1]
}

func (c *Char) botBoundDist() float32 {
	return sys.zmax/c.localscl - c.pos[2]
}

func (c *Char) canRecover() bool {
	return c.ghv.fall_recover && c.fallTime >= c.ghv.fall_recovertime
}

func (c *Char) comboCount() int32 {
	if c.teamside == -1 {
		return 0
	}
	return sys.lifebar.co[c.teamside].combo
}

func (c *Char) command(pn, i int) bool {
	if !c.keyctrl[0] || c.cmd == nil {
		return false
	}

	// Get all commands with the specified first index (name)
	cl := c.cmd[pn].At(i)

	// Check if any of them are buffered
	for _, c := range cl {
		if c.curbuftime > 0 {
			return true
		}
	}

	// AI cheating for commands longer than 1 button
	// Maybe it could just cheat all of them and skip these checks
	if !c.asf(ASF_noaicheat) && c.controller < 0 && len(cl) > 0 {
		steps := cl[0].steps
		multiStep := len(steps) > 1
		multiKey := len(steps) > 0 && len(steps[0].keys) > 1

		if c.helperIndex != 0 || multiStep || multiKey {
			if i == int(c.cpucmd) {
				return true
			}
		}
		// Our AI cheating is more efficient but it's not accurate to Mugen
		// In Mugen, essentially command trigger returns true on average once every 10 seconds for difficulty 8,
		// and once every 30 seconds for difficulty 1. Other difficulties are probably linearly interpolated
	}

	return false
}

func (c *Char) commandByName(name string) bool {
	if c.cmd == nil {
		return false
	}
	i, ok := c.cmd[c.playerNo].Names[name]
	return ok && c.command(c.playerNo, i)
}

func (c *Char) assertCommand(name string, time int32) {
	// If no command name is provided, select one randomly.
	if name == "" {
		cmdList := &c.cmd[c.playerNo]
		if len(cmdList.Commands) == 0 {
			return
		}
		randomIndex := Rand(0, int32(len(cmdList.Commands)-1))
		cmdInstances := cmdList.Commands[randomIndex]
		if len(cmdInstances) == 0 {
			return
		}
		name = cmdInstances[0].name
		if time <= 0 {
			time = cmdInstances[0].maxbuftime
			if time <= 0 {
				time = 1
			}
		}
	}

	// Assert the command in every command list
	found := false
	for i := range c.cmd {
		found = c.cmd[i].Assert(name, time) || found
	}

	if !found {
		sys.appendToConsole(c.warn() + fmt.Sprintf("attempted to assert an invalid command: %s", name))
	}
}

func (c *Char) constp(coordinate, value float32) BytecodeValue {
	return BytecodeFloat(c.stOgi().localcoord[0] / coordinate * value)
}

func (c *Char) ctrl() bool {
	return c.scf(SCF_ctrl) && !c.scf(SCF_standby) &&
		!c.scf(SCF_dizzy) && !c.scf(SCF_guardbreak)
}

func (c *Char) drawgame() bool {
	return sys.roundState() >= 3 && sys.winTeam < 0
}

func (c *Char) frontEdge() float32 {
	if c.facing > 0 {
		return c.rightEdge()
	}
	return c.leftEdge()
}

func (c *Char) frontEdgeBodyDist() float32 {
	// See BackEdgeBodyDist
	offset := float32(0)
	if c.stWgi().ikemenver[0] == 0 && c.stWgi().ikemenver[1] == 0 {
		if c.ss.stateType == ST_A {
			offset = 0.5 / c.localscl
		} else if c.ss.stateType == ST_L {
			offset = 1.0 / c.localscl
		}
	}
	return c.frontEdgeDist() - c.edgeWidth[0] - offset
}

func (c *Char) frontEdgeDist() float32 {
	if c.facing > 0 {
		return sys.xmax/c.localscl - c.pos[0]
	}
	return c.pos[0] - sys.xmin/c.localscl
}

func (c *Char) gameHeight() float32 {
	return c.screenHeight() / sys.cam.Scale
}

func (c *Char) gameWidth() float32 {
	return c.screenWidth() / sys.cam.Scale
}

func (c *Char) getPlayerID(pn int) int32 {
	if pn >= 1 && pn <= len(sys.chars) && len(sys.chars[pn-1]) > 0 {
		return sys.chars[pn-1][0].id
	}
	return 0
}

// Handle power sharing
// Neutral players (attached characters) have no apparent reasons to share power
func (c *Char) powerOwner() *Char {
	if sys.cfg.Options.Team.PowerShare && (c.teamside == 0 || c.teamside == 1) {
		return sys.chars[c.teamside][0]
		// TODO: If we ever expand on teamside switching, this could loop over sys.chars and return the first one on the player's side
		// But currently this method is just slightly more efficient
	}
	return sys.chars[c.playerNo][0]
}

func (c *Char) getPower() int32 {
	return c.powerOwner().power
}

func (c *Char) hitDefAttr(attr int32) bool {
	return c.ss.moveType == MT_A && c.hitdef.testAttr(attr)
}

func (c *Char) hitOver() bool {
	return c.ghv.hittime < 0
}

func (c *Char) hitShakeOver() bool {
	return c.ghv.hitshaketime <= 0
}

func (c *Char) isHelper(id int32, idx int) bool {
	// Not a helper at all
	if c.helperIndex == 0 {
		return false
	}

	// Backward compatibility
	if c.stWgi().ikemenver[0] == 0 && c.stWgi().ikemenver[1] == 0 {
		// Some Mugen characters used "isHelper(-1)" even though it was meaningless there
		// https://github.com/ikemen-engine/Ikemen-GO/issues/2415
		if id < 0 && id != math.MinInt32 && idx == math.MinInt32 {
			sys.appendToConsole(c.warn() + fmt.Sprintf("invalid IsHelper ID for char engine version: %v", id))
			return false
		}
	}

	// Any helper
	if id < 0 && idx < 0 {
		return true
	}

	// Check ID only
	if id >= 0 && idx < 0 {
		return c.helperId == id
	}

	// Check specific ID or index
	var count int
	for _, h := range sys.charList.runOrder {
		// Skip roots, helpers from other players and destroyed helpers
		// Mugen does not skip DestroySelf helpers here. What it does is clear helperId when DestroySelf is called
		// However, skipping them is more consistent with the other helper triggers
		if h.helperIndex == 0 || h.playerNo != c.playerNo || h.csf(CSF_destroy) {
			continue
		}

		// Check specific ID
		if id >= 0 && h.helperId != id {
			continue
		}

		// Check any index
		if idx < 0 {
			if h == c {
				return true
			}
			continue
		}

		// Check specific index
		if count == idx {
			return h == c
		}

		count++
	}

	return false
}

func (c *Char) isHost() bool {
	// Local play has no host
	if sys.netConnection == nil && sys.replayFile == nil && sys.rollback.session == nil {
		return false
	}

	// Find first human player like in GetHostGuestRemap()
	// This doesn't seem ideal somehow, but it's better than not having it
	// When you think about it, it's almost the same as just returning true for player 1
	var host int
	for i, v := range sys.aiLevel {
		if v == 0 {
			host = i
			break
		}
	}

	// "host" already defaults to 0 so player 1 is the fallback host
	return c.playerNo == host

	// TODO: For Tag mode, this should probably return true for all characters controlled by the player

	// For the host, this returned true for any player
	// For the guest, it returned false for any player
	// https://github.com/ikemen-engine/Ikemen-GO/issues/2523
	//return sys.netConnection != nil && sys.netConnection.host
}

func (c *Char) jugglePoints(id int32) int32 {
	max := c.gi().data.airjuggle

	// Check if ID is already a target
	for _, ct := range c.targets {
		t := sys.playerID(ct)
		if t != nil && t.id == id {
			return t.ghv.getJuggle(c.id, max)
		}
	}

	// If no target is found we just return the char's maximum juggle points
	return max
}

func (c *Char) leftEdge() float32 {
	return sys.cam.ScreenPos[0] / c.localscl
}

func (c *Char) lose() bool {
	if c.teamside == -1 {
		return false
	}
	return sys.winTeam == ^c.playerNo&1
}

func (c *Char) loseKO() bool {
	return c.lose() && sys.finishType == FT_KO
}

func (c *Char) loseTime() bool {
	return c.lose() && sys.finishType == FT_TO
}

func (c *Char) moveContact() int32 {
	if c.mctype != MC_Reversed {
		return Abs(c.mctime)
	}
	return 0
}

func (c *Char) moveCountered() int32 {
	if c.counterHit {
		return Abs(c.mctime)
	}
	return 0
}

func (c *Char) moveGuarded() int32 {
	if c.mctype == MC_Guarded {
		return Abs(c.mctime)
	}
	return 0
}

func (c *Char) moveHit() int32 {
	if c.mctype == MC_Hit {
		return Abs(c.mctime)
	}
	return 0
}

func (c *Char) moveReversed() int32 {
	if c.mctype == MC_Reversed {
		return Abs(c.mctime)
	}
	return 0
}

func (c *Char) numEnemy() int32 {
	var n int32

	// Same as enemy() loop
	for _, e := range sys.chars {
		if len(e) > 0 && e[0] != nil && c.isEnemyOf(e[0]) {
			n += 1
		}
	}
	return n
}

func (c *Char) numExplod(eid BytecodeValue) BytecodeValue {
	if eid.IsSF() {
		return BytecodeSF()
	}
	var id, n int32 = eid.ToI(), 0
	for i := range sys.explods[c.playerNo] {
		e := sys.explods[c.playerNo][i]
		if e.matchId(id, c.id) {
			n++
		}
	}
	return BytecodeInt(n)
}

func (c *Char) numPlayer() int32 {
	var count int32

	// Ignore destroyed helpers for the sake of consistency
	for _, ch := range sys.charList.runOrder {
		if !ch.csf(CSF_destroy) {
			count++
		}
	}

	return count

	//return int32(len(sys.charList.runOrder))
}

func (c *Char) numText(textid BytecodeValue) BytecodeValue {
	if textid.IsSF() {
		return BytecodeSF()
	}
	var id, n int32 = textid.ToI(), 0
	for _, ts := range sys.lifebar.textsprite {
		if ts.id == id && ts.ownerid == c.id {
			n++
		}
	}
	return BytecodeInt(n)
}

func (c *Char) explodVar(eid BytecodeValue, idx BytecodeValue, vtype OpCode) BytecodeValue {
	if eid.IsSF() {
		return BytecodeSF()
	}
	var id = eid.ToI()
	var i = idx.ToI()
	var v BytecodeValue
	for n, e := range c.getExplods(id) {
		if i == int32(n) {
			switch vtype {
			case OC_ex2_explodvar_anim:
				v = BytecodeInt(e.animNo)
			case OC_ex2_explodvar_angle:
				v = BytecodeFloat(e.anglerot[0] + e.interpolate_angle[0])
			case OC_ex2_explodvar_angle_x:
				v = BytecodeFloat(e.anglerot[1] + e.interpolate_angle[1])
			case OC_ex2_explodvar_angle_y:
				v = BytecodeFloat(e.anglerot[2] + e.interpolate_angle[2])
			case OC_ex2_explodvar_animelem:
				v = BytecodeInt(e.anim.curelem + 1)
			case OC_ex2_explodvar_animelemtime:
				v = BytecodeInt(e.anim.curelemtime)
			case OC_ex2_explodvar_animplayerno:
				v = BytecodeInt(int32(e.animPN) + 1)
			case OC_ex2_explodvar_spriteplayerno:
				v = BytecodeInt(int32(e.spritePN) + 1)
			case OC_ex2_explodvar_bindtime:
				v = BytecodeInt(e.bindtime)
			case OC_ex2_explodvar_drawpal_group:
				v = BytecodeInt(c.explodDrawPal(e)[0])
			case OC_ex2_explodvar_drawpal_index:
				v = BytecodeInt(c.explodDrawPal(e)[1])
			case OC_ex2_explodvar_facing:
				v = BytecodeInt(int32(e.facing))
			case OC_ex2_explodvar_id:
				v = BytecodeInt(e.id)
			case OC_ex2_explodvar_layerno:
				v = BytecodeInt(e.layerno)
			case OC_ex2_explodvar_pausemovetime:
				v = BytecodeInt(e.pausemovetime)
			case OC_ex2_explodvar_pos_x:
				v = BytecodeFloat(e.pos[0] + e.offset[0] + e.relativePos[0] + e.interpolate_pos[0])
			case OC_ex2_explodvar_pos_y:
				v = BytecodeFloat(e.pos[1] + e.offset[1] + e.relativePos[1] + e.interpolate_pos[1])
			case OC_ex2_explodvar_pos_z:
				v = BytecodeFloat(e.pos[2] + e.offset[2] + e.relativePos[2] + e.interpolate_pos[2])
			case OC_ex2_explodvar_removetime:
				v = BytecodeInt(e.removetime)
			case OC_ex2_explodvar_scale_x:
				v = BytecodeFloat(e.scale[0] * e.interpolate_scale[0])
			case OC_ex2_explodvar_scale_y:
				v = BytecodeFloat(e.scale[1] * e.interpolate_scale[1])
			case OC_ex2_explodvar_sprpriority:
				v = BytecodeInt(e.sprpriority)
			case OC_ex2_explodvar_time:
				v = BytecodeInt(e.time)
			case OC_ex2_explodvar_vel_x:
				v = BytecodeFloat(e.velocity[0])
			case OC_ex2_explodvar_vel_y:
				v = BytecodeFloat(e.velocity[1])
			case OC_ex2_explodvar_vel_z:
				v = BytecodeFloat(e.velocity[2])
			case OC_ex2_explodvar_xshear:
				v = BytecodeFloat(e.xshear)
			}
			break
		}
	}
	return v
}

func (c *Char) projVar(pid BytecodeValue, idx BytecodeValue, flag BytecodeValue, vtype OpCode, oc *Char) BytecodeValue {
	if pid.IsSF() {
		return BytecodeSF()
	}

	// See compiler.go:ProjVar
	var id int32 = pid.ToI()
	if id > 0 {
		id--
	}

	var i = idx.ToI()
	var fl int32 = flag.ToI()
	var v BytecodeValue
	projs := c.getProjs(id)
	if len(projs) == 0 {
		return BytecodeSF()
	}
	for n, p := range projs {
		if i == int32(n) {
			switch vtype {
			case OC_ex2_projvar_accel_x:
				v = BytecodeFloat(p.accel[0] * p.localscl)
			case OC_ex2_projvar_accel_y:
				v = BytecodeFloat(p.accel[1] * p.localscl)
			case OC_ex2_projvar_accel_z:
				v = BytecodeFloat(p.accel[2] * p.localscl)
			case OC_ex2_projvar_animelem:
				v = BytecodeInt(p.ani.curelem + 1)
			case OC_ex2_projvar_drawpal_group:
				v = BytecodeInt(c.projDrawPal(p)[0])
			case OC_ex2_projvar_drawpal_index:
				v = BytecodeInt(c.projDrawPal(p)[1])
			case OC_ex2_projvar_facing:
				v = BytecodeFloat(p.facing)
			case OC_ex2_projvar_guardflag:
				v = BytecodeBool(p.hitdef.guardflag&fl != 0)
			case OC_ex2_projvar_highbound:
				v = BytecodeInt(int32(float32(p.heightbound[1]) * p.localscl / oc.localscl))
			case OC_ex2_projvar_hitflag:
				v = BytecodeBool(p.hitdef.hitflag&fl != 0)
			case OC_ex2_projvar_lowbound:
				v = BytecodeInt(int32(float32(p.heightbound[0]) * p.localscl / oc.localscl))
			case OC_ex2_projvar_pausemovetime:
				v = BytecodeInt(p.pausemovetime)
			case OC_ex2_projvar_pos_x:
				v = BytecodeFloat((p.pos[0]*p.localscl - sys.cam.Pos[0]) / oc.localscl)
			case OC_ex2_projvar_pos_y:
				v = BytecodeFloat(p.pos[1] * p.localscl / oc.localscl)
			case OC_ex2_projvar_pos_z:
				v = BytecodeFloat(p.pos[2] * p.localscl / oc.localscl)
			case OC_ex2_projvar_projanim:
				v = BytecodeInt(p.anim)
			case OC_ex2_projvar_projangle:
				v = BytecodeFloat(p.anglerot[0])
			case OC_ex2_projvar_projyangle:
				v = BytecodeFloat(p.anglerot[2])
			case OC_ex2_projvar_projxangle:
				v = BytecodeFloat(p.anglerot[1])
			case OC_ex2_projvar_projcancelanim:
				v = BytecodeInt(p.cancelanim)
			case OC_ex2_projvar_projedgebound:
				v = BytecodeInt(int32(float32(p.edgebound) * p.localscl / oc.localscl))
			case OC_ex2_projvar_projhitanim:
				v = BytecodeInt(p.hitanim)
			case OC_ex2_projvar_projhits:
				v = BytecodeInt(p.hits)
			case OC_ex2_projvar_projhitsmax:
				v = BytecodeInt(p.totalhits)
			case OC_ex2_projvar_projid:
				v = BytecodeInt(int32(p.id))
			case OC_ex2_projvar_projlayerno:
				v = BytecodeInt(p.layerno)
			case OC_ex2_projvar_projmisstime:
				v = BytecodeInt(p.curmisstime)
			case OC_ex2_projvar_projpriority:
				v = BytecodeInt(p.priority)
			case OC_ex2_projvar_projremove:
				v = BytecodeBool(p.remove)
			case OC_ex2_projvar_projremanim:
				v = BytecodeInt(p.remanim)
			case OC_ex2_projvar_projremovetime:
				v = BytecodeInt(p.removetime)
			case OC_ex2_projvar_projscale_x:
				v = BytecodeFloat(p.scale[0])
			case OC_ex2_projvar_projscale_y:
				v = BytecodeFloat(p.scale[1])
			case OC_ex2_projvar_projshadow_b:
				v = BytecodeInt(p.shadow[2])
			case OC_ex2_projvar_projshadow_g:
				v = BytecodeInt(p.shadow[1])
			case OC_ex2_projvar_projshadow_r:
				v = BytecodeInt(p.shadow[0])
			case OC_ex2_projvar_projsprpriority:
				v = BytecodeInt(p.sprpriority)
			case OC_ex2_projvar_projstagebound:
				v = BytecodeInt(int32(float32(p.stagebound) * p.localscl / oc.localscl))
			case OC_ex2_projvar_projxshear:
				v = BytecodeFloat(p.xshear)
			case OC_ex2_projvar_remvelocity_x:
				v = BytecodeFloat(p.remvelocity[0] * p.localscl / oc.localscl)
			case OC_ex2_projvar_remvelocity_y:
				v = BytecodeFloat(p.remvelocity[1] * p.localscl / oc.localscl)
			case OC_ex2_projvar_remvelocity_z:
				v = BytecodeFloat(p.remvelocity[2] * p.localscl / oc.localscl)
			case OC_ex2_projvar_supermovetime:
				v = BytecodeInt(p.supermovetime)
			case OC_ex2_projvar_teamside:
				v = BytecodeInt(int32(p.hitdef.teamside))
			case OC_ex2_projvar_time:
				v = BytecodeInt(p.time)
			case OC_ex2_projvar_vel_x:
				v = BytecodeFloat(p.velocity[0] * p.localscl / oc.localscl)
			case OC_ex2_projvar_vel_y:
				v = BytecodeFloat(p.velocity[1] * p.localscl / oc.localscl)
			case OC_ex2_projvar_vel_z:
				v = BytecodeFloat(p.velocity[2] * p.localscl / oc.localscl)
			case OC_ex2_projvar_velmul_x:
				v = BytecodeFloat(p.velmul[0])
			case OC_ex2_projvar_velmul_y:
				v = BytecodeFloat(p.velmul[1])
			case OC_ex2_projvar_velmul_z:
				v = BytecodeFloat(p.velmul[2])
			}
			break
		}
	}
	return v
}

func (c *Char) soundVar(chid BytecodeValue, vtype OpCode) BytecodeValue {
	if chid.IsSF() {
		return BytecodeSF()
	}

	// See compiler.go:SoundVar
	var id = chid.ToI()
	if id > 0 {
		id--
	}
	var ch *SoundChannel

	// First, grab a channel.
	if id >= 0 {
		ch = c.soundChannels.Get(id)
	} else {
		if c != nil && c.soundChannels.channels != nil {
			for i := 0; i < int(c.soundChannels.count()); i++ {
				if c.soundChannels.channels[i].sfx != nil {
					if c.soundChannels.channels[i].IsPlaying() {
						ch = &c.soundChannels.channels[i]
						break
					}
				}
			}
		}
	}

	// Now get the data we want
	switch vtype {
	case OC_ex2_soundvar_group:
		if ch != nil && ch.sound != nil {
			return BytecodeInt(ch.group)
		}
		return BytecodeInt(-1)
	case OC_ex2_soundvar_number:
		if ch != nil && ch.sound != nil {
			return BytecodeInt(ch.number)
		}
		return BytecodeInt(-1)
	case OC_ex2_soundvar_freqmul:
		if ch != nil && ch.sfx != nil {
			return BytecodeFloat(ch.sfx.freqmul)
		}
		return BytecodeFloat(1)
	case OC_ex2_soundvar_isplaying:
		if ch != nil && ch.sfx != nil {
			return BytecodeBool(ch.IsPlaying())
		}
		return BytecodeBool(false)
	case OC_ex2_soundvar_length:
		if ch != nil && ch.streamer != nil {
			return BytecodeInt64(int64(ch.streamer.Len()))
		}
		return BytecodeInt64(int64(0))
	case OC_ex2_soundvar_loopcount:
		if ch != nil {
			if sl, ok := ch.sfx.streamer.(*StreamLooper); ok {
				return BytecodeInt(int32(sl.loopcount))
			}
		}
		return BytecodeInt(0)
	case OC_ex2_soundvar_loopend:
		if ch != nil {
			if sl, ok := ch.sfx.streamer.(*StreamLooper); ok {
				return BytecodeInt64(int64(sl.loopend))
			}
		}
		return BytecodeInt64(0)
	case OC_ex2_soundvar_loopstart:
		if ch != nil {
			if sl, ok := ch.sfx.streamer.(*StreamLooper); ok {
				return BytecodeInt64(int64(sl.loopstart))
			}
		}
		return BytecodeInt64(0)
	case OC_ex2_soundvar_pan:
		if ch != nil && ch.sfx != nil {
			return BytecodeFloat(ch.sfx.p)
		}
		return BytecodeFloat(0)
	case OC_ex2_soundvar_position:
		if ch != nil {
			if sl, ok := ch.sfx.streamer.(*StreamLooper); ok {
				return BytecodeInt64(int64(sl.Position()))
			}
		}
		return BytecodeInt64(0)
	case OC_ex2_soundvar_priority:
		if ch != nil && ch.sfx != nil {
			return BytecodeInt(ch.sfx.priority)
		}
		return BytecodeInt(0)
	case OC_ex2_soundvar_startposition:
		if ch != nil && ch.sfx != nil {
			return BytecodeInt64(int64(ch.sfx.startPos))
		}
		return BytecodeInt64(int64(0))
	case OC_ex2_soundvar_volumescale:
		if ch != nil && ch.sfx != nil {
			return BytecodeFloat(ch.sfx.volume / 256.0 * 100.0)
		}
		return BytecodeFloat(0)
	}

	return BytecodeSF()
}

func (c *Char) numHelper(hid BytecodeValue) BytecodeValue {
	if hid.IsSF() {
		return BytecodeSF()
	}
	var id, count int32 = hid.ToI(), 0

	// Mugen confirmed to skip helpers under DestroySelf in the same frame
	for _, h := range sys.chars[c.playerNo][1:] {
		if !h.csf(CSF_destroy) && (id <= 0 || h.helperId == id) {
			count++
		}
	}

	return BytecodeInt(count)
}

func (c *Char) numPartner() int32 {
	if (sys.tmode[c.playerNo&1] != TM_Simul && sys.tmode[c.playerNo&1] != TM_Tag) || c.teamside == -1 {
		return 0
	}
	return sys.numSimul[c.playerNo&1] - 1
}

func (c *Char) numProj() int32 {
	// Helpers cannot own projectiles
	if c.helperIndex != 0 {
		return 0
	}
	n := int32(0)
	for i := range sys.projs[c.playerNo] {
		p := sys.projs[c.playerNo][i]
		if p.id >= 0 && !((p.hits < 0 && p.remove) || p.remflag) {
			n++
		}
	}
	return n
}

func (c *Char) numProjID(pid BytecodeValue) BytecodeValue {
	if pid.IsSF() {
		return BytecodeSF()
	}
	// Helpers cannot own projectiles
	if c.helperIndex != 0 {
		return BytecodeInt(0)
	}
	var id, n int32 = Max(0, pid.ToI()), 0
	for i := range sys.projs[c.playerNo] {
		p := sys.projs[c.playerNo][i]
		if p.id == id && !((p.hits < 0 && p.remove) || p.remflag) {
			n++
		}
	}
	return BytecodeInt(n)
}

func (c *Char) numTarget(hid BytecodeValue) BytecodeValue {
	if hid.IsSF() {
		return BytecodeSF()
	}
	var id, n int32 = hid.ToI(), 0
	for _, tid := range c.targets {
		if tid >= 0 {
			if id < 0 {
				n++
			} else if t := sys.playerID(tid); t != nil && t.ghv.hitid == id {
				n++
			}
		}
	}
	return BytecodeInt(n)
}

func (c *Char) palfxvar(x int32) int32 {
	n := int32(0)
	if x >= 4 {
		n = 256
	}
	if c.palfx != nil && c.palfx.enable {
		switch x {
		case -2:
			n = c.palfx.eInvertblend
		case -1:
			n = Btoi(c.palfx.eInvertall)
		case 0:
			n = c.palfx.time
		case 1:
			n = c.palfx.eAdd[0]
		case 2:
			n = c.palfx.eAdd[1]
		case 3:
			n = c.palfx.eAdd[2]
		case 4:
			n = c.palfx.eMul[0]
		case 5:
			n = c.palfx.eMul[1]
		case 6:
			n = c.palfx.eMul[2]
		default:
			n = 0
		}
	}
	return n
}

func (c *Char) palfxvar2(x int32) float32 {
	n := float32(1)
	if x > 1 {
		n = 0
	}
	if c.palfx != nil && c.palfx.enable {
		switch x {
		case 1:
			n = c.palfx.eColor
		case 2:
			n = c.palfx.eHue
		default:
			n = 0
		}
	}
	return n * 256
}

func (c *Char) pauseTimeTrigger() int32 {
	var p int32
	if sys.supertime > 0 && c.prevSuperMovetime == 0 {
		p = sys.supertime
	}
	if sys.pausetime > 0 && c.prevPauseMovetime == 0 && p < sys.pausetime {
		p = sys.pausetime
	}
	return p
}

func (c *Char) projCancelTime(pid BytecodeValue) BytecodeValue {
	if pid.IsSF() {
		return BytecodeSF()
	}
	id := pid.ToI()
	if (id > 0 && id != c.gi().pcid) || c.gi().pctype != PC_Cancel || c.helperIndex > 0 {
		return BytecodeInt(-1)
	}
	return BytecodeInt(c.gi().pctime)
}

func (c *Char) projContactTime(pid BytecodeValue) BytecodeValue {
	if pid.IsSF() {
		return BytecodeSF()
	}
	id := pid.ToI()
	if (id > 0 && id != c.gi().pcid) || c.gi().pctype == PC_Cancel || c.helperIndex > 0 {
		return BytecodeInt(-1)
	}
	return BytecodeInt(c.gi().pctime)
}

func (c *Char) projGuardedTime(pid BytecodeValue) BytecodeValue {
	if pid.IsSF() {
		return BytecodeSF()
	}
	id := pid.ToI()
	if (id > 0 && id != c.gi().pcid) || c.gi().pctype != PC_Guarded || c.helperIndex > 0 {
		return BytecodeInt(-1)
	}
	return BytecodeInt(c.gi().pctime)
}

func (c *Char) projHitTime(pid BytecodeValue) BytecodeValue {
	if pid.IsSF() {
		return BytecodeSF()
	}
	id := pid.ToI()
	if (id > 0 && id != c.gi().pcid) || c.gi().pctype != PC_Hit || c.helperIndex > 0 {
		return BytecodeInt(-1)
	}
	return BytecodeInt(c.gi().pctime)
}

func (c *Char) reversalDefAttr(attr int32) bool {
	return c.hitdef.testReversalAttr(attr)
}

func (c *Char) rightEdge() float32 {
	return sys.cam.ScreenPos[0]/c.localscl + c.gameWidth()
}

func (c *Char) roundsExisted() int32 {
	if c.teamside == -1 {
		return sys.round - 1
	}
	return sys.roundsExisted[c.playerNo&1]
}

func (c *Char) roundsWon() int32 {
	if c.teamside == -1 {
		return 0
	}
	return sys.wins[c.playerNo&1]
}

// TODO: These are supposed to be affected by zoom camera shifting
// In Mugen 1.1 they don't work properly when zoom scale is actually used
// Perhaps in Ikemen they could return the final rendering position of the chars
func (c *Char) screenPosX() float32 {
	return (c.pos[0]*c.localscl - sys.cam.ScreenPos[0]) // * sys.cam.Scale
}

func (c *Char) screenPosY() float32 {
	return (c.pos[1]*c.localscl - sys.cam.ScreenPos[1]) // * sys.cam.Scale
}

func (c *Char) screenHeight() float32 {
	return sys.screenHeight() / (320.0 / float32(c.stOgi().localcoord[0])) /
		((3.0 / 4.0) / (float32(sys.scrrect[3]) / float32(sys.scrrect[2])))
}

func (c *Char) screenWidth() float32 {
	return c.stOgi().localcoord[0]
}

func (c *Char) selfAnimExist(anim BytecodeValue) BytecodeValue {
	if anim.IsSF() {
		return BytecodeSF()
	}
	return BytecodeBool(c.gi().anim.get(anim.ToI()) != nil)
}

func (c *Char) selfStatenoExist(stateno BytecodeValue) BytecodeValue {
	if stateno.IsSF() {
		return BytecodeSF()
	}
	_, ok := c.gi().states[stateno.ToI()]
	return BytecodeBool(ok)
}

// If the stage is coded incorrectly we must check distance to "leftbound" or "rightbound"
// https://github.com/ikemen-engine/Ikemen-GO/issues/1996
func (c *Char) stageFrontEdgeDist() float32 {
	corner := float32(0)
	if c.facing < 0 {
		corner = MaxF(sys.cam.XMin/c.localscl+sys.screenleft/c.localscl,
			sys.stage.leftbound*sys.stage.localscl/c.localscl)
		return c.pos[0] - corner
	} else {
		corner = MinF(sys.cam.XMax/c.localscl-sys.screenright/c.localscl,
			sys.stage.rightbound*sys.stage.localscl/c.localscl)
		return corner - c.pos[0]
	}
}

func (c *Char) stageBackEdgeDist() float32 {
	corner := float32(0)
	if c.facing < 0 {
		corner = MinF(sys.cam.XMax/c.localscl-sys.screenright/c.localscl,
			sys.stage.rightbound*sys.stage.localscl/c.localscl)
		return corner - c.pos[0]
	} else {
		corner = MaxF(sys.cam.XMin/c.localscl+sys.screenleft/c.localscl,
			sys.stage.leftbound*sys.stage.localscl/c.localscl)
		return c.pos[0] - corner
	}
}

func (c *Char) teamLeader() int {
	if c.teamside == -1 || sys.tmode[c.playerNo&1] == TM_Single || sys.tmode[c.playerNo&1] == TM_Turns {
		return c.playerNo + 1
	}
	return sys.teamLeader[c.playerNo&1] + 1
}

func (c *Char) teamSize() int32 {
	if c.teamside == -1 {
		var n int32
		for i := MaxSimul * 2; i < len(sys.chars); i++ {
			if len(sys.chars[i]) > 0 {
				n += 1
			}
		}
		return n
	}
	if sys.tmode[c.playerNo&1] == TM_Turns {
		return sys.numTurns[c.playerNo&1]
	}
	return sys.numSimul[c.playerNo&1]
}

func (c *Char) time() int32 {
	return c.ss.time
}

func (c *Char) topEdge() float32 {
	return sys.cam.ScreenPos[1] / c.localscl
}

func (c *Char) topBoundBodyDist() float32 {
	return c.topBoundDist() - c.edgeDepth[0]
}

func (c *Char) topBoundDist() float32 {
	return c.pos[2] - sys.zmin/c.localscl
}

func (c *Char) win() bool {
	if c.teamside == -1 {
		return false
	}
	return sys.winTeam == c.playerNo&1
}

func (c *Char) winKO() bool {
	return c.win() && sys.finishType == FT_KO
}

func (c *Char) winTime() bool {
	return c.win() && sys.finishType == FT_TO
}

func (c *Char) winPerfect() bool {
	return c.win() && sys.winType[c.playerNo&1] >= WT_PNormal
}

func (c *Char) winType(wt WinType) bool {
	return c.win() && sys.winTrigger[c.playerNo&1] == wt
}

func (c *Char) playSound(ffx string, lowpriority bool, loopCount int32, g, n, chNo, vol int32,
	p, freqmul, ls float32, x *float32, log bool, priority int32, loopstart, loopend, startposition int, stopgh, stopcs bool) {
	if g < 0 {
		return
	}
	current_ffx := ffx
	if current_ffx == "f" {
		if c.gi().fightfxPrefix != "" {
			current_ffx = c.gi().fightfxPrefix
		}
	}
	// Don't do anything if we have the nosound command line flag
	if _, ok := sys.cmdFlags["-nosound"]; ok {
		return
	}
	var s *Sound
	if current_ffx == "" || current_ffx == "s" {
		if c.gi().snd != nil {
			s = c.gi().snd.Get([...]int32{g, n})
		}
	} else {
		if sys.ffx[current_ffx] != nil && sys.ffx[current_ffx].fsnd != nil {
			s = sys.ffx[current_ffx].fsnd.Get([...]int32{g, n})
		}
	}
	if s == nil {
		if log {
			if current_ffx != "" {
				sys.appendToConsole(c.warn() + fmt.Sprintf("sound %v %v,%v doesn't exist", strings.ToUpper(current_ffx), g, n))
			} else {
				sys.appendToConsole(c.warn() + fmt.Sprintf("sound %v,%v doesn't exist", g, n))
			}
		}
		if !sys.ignoreMostErrors {
			str := "Sound doesn't exist: "
			if current_ffx != "" {
				str += current_ffx + ":"
			} else {
				str += fmt.Sprintf("P%v:", c.playerNo+1)
			}
			sys.errLog.Printf("%v%v,%v\n", str, g, n)
		}
		return
	}
	crun := c
	if c.inheritChannels == 1 && c.parent(false) != nil {
		crun = c.parent(false)
	} else if c.inheritChannels == 2 && c.root(false) != nil {
		crun = c.root(false)
	}
	if ch := crun.soundChannels.New(chNo, lowpriority, priority); ch != nil {
		ch.Play(s, g, n, loopCount, freqmul, loopstart, loopend, startposition)
		vol = Clamp(vol, -25600, 25600)
		//if c.gi().mugenver[0] == 1 {
		if current_ffx != "" {
			ch.SetVolume(float32(vol * 64 / 25))
		} else {
			ch.SetVolume(float32(c.gi().data.volume * vol / 100))
		}
		if chNo >= 0 {
			ch.SetChannel(chNo)
			if priority != 0 {
				ch.SetPriority(priority)
			}
		}
		//} else {
		//	if f {
		//		ch.SetVolume(float32(vol + 256))
		//	} else {
		//		ch.SetVolume(float32(c.gi().data.volume + vol))
		//	}
		//}
		ch.stopOnGetHit = stopgh
		ch.stopOnChangeState = stopcs
		ch.SetPan(p*c.facing, ls, x)
	}
}

func (c *Char) autoTurn() {
	if c.helperIndex == 0 && !c.asf(ASF_noautoturn) && sys.stage.autoturn && c.shouldFaceP2() {
		switch c.ss.stateType {
		case ST_S:
			if c.animNo != 5 {
				c.changeAnim(5, c.playerNo, -1, "")
			}
		case ST_C:
			if c.animNo != 6 {
				c.changeAnim(6, c.playerNo, -1, "")
			}
		}
		c.setFacing(-c.facing)
	}
}

// Flag if B and F directions should reverse, i.e. respectively use R and L
// In Mugen this is hardcoded to be based on facing
func (c *Char) updateFBFlip() {
	setting := c.gi().constants["input.fbflipenemydistance"]

	if setting >= 0 {
		// See shouldFaceP2()
		e := c.p2()
		if e == nil {
			e = c.p2EnemyBackup
		}
		if e != nil {
			distX := c.rdDistX(e, c).ToF() // Already in the char's localcoord

			if c.facing > 0 {
				c.fbFlip = distX < -setting
			} else {
				c.fbFlip = distX > -setting
			}
		}
	} else {
		c.fbFlip = (c.facing < 0)
	}
}

// Check if P2 enemy is behind the player and the player is allowed to face them
func (c *Char) shouldFaceP2() bool {
	// Face P2 normally
	e := c.p2()

	// If P2 was not found, fall back to the last valid one
	// Maybe this should only happen during win poses?
	if e == nil {
		e = c.p2EnemyBackup
	}

	if e != nil && !e.asf(ASF_noturntarget) {
		distX := c.rdDistX(e, c).ToF()
		if sys.zEnabled() {
			// Use a z position tie breaker when the x positions are the same
			if distX < 0 ||
				distX == 0 && (c.rdDistZ(e, c).ToF() < 0) == (c.pos[0]*c.facing > 0) {
				return true
			}
		} else {
			if distX < 0 {
				return true
			}
		}
	}
	return false
}

func (c *Char) stateChange1(no int32, pn int) bool {
	if sys.changeStateNest >= MaxLoop {
		sys.appendToConsole(c.warn() + fmt.Sprintf("state machine stuck in loop (stopped after %v loops): %v -> %v -> %v", sys.changeStateNest, c.ss.prevno, c.ss.no, no))
		sys.errLog.Printf("Maximum ChangeState loops: %v, %v, %v -> %v -> %v\n", sys.changeStateNest, c.name, c.ss.prevno, c.ss.no, no)
		return false
	}

	c.ss.prevno = c.ss.no
	c.ss.no = Max(0, no)
	c.ss.time = 0

	// Local scale updates
	// If the new state uses a different localcoord, some values need to be updated in the same frame
	if newLs := 320 / sys.chars[pn][0].localcoord; c.localscl != newLs {
		lsRatio := c.localscl / newLs
		c.pos[0] *= lsRatio
		c.pos[1] *= lsRatio
		c.pos[2] *= lsRatio
		c.oldPos = c.pos
		c.interPos = c.pos

		c.vel[0] *= lsRatio
		c.vel[1] *= lsRatio
		c.vel[2] *= lsRatio

		c.ghv.xvel *= lsRatio
		c.ghv.yvel *= lsRatio
		c.ghv.zvel *= lsRatio
		c.ghv.fall_xvelocity *= lsRatio
		c.ghv.fall_yvelocity *= lsRatio
		c.ghv.fall_zvelocity *= lsRatio
		c.ghv.xaccel *= lsRatio
		c.ghv.yaccel *= lsRatio
		c.ghv.zaccel *= lsRatio

		c.sizeWidth[0] *= lsRatio
		c.sizeWidth[1] *= lsRatio
		c.sizeHeight[0] *= lsRatio
		c.sizeHeight[1] *= lsRatio
		c.updateSizeBox()

		c.sizeDepth[0] *= lsRatio
		c.sizeDepth[1] *= lsRatio

		c.edgeWidth[0] *= lsRatio
		c.edgeWidth[1] *= lsRatio
		c.edgeDepth[0] *= lsRatio
		c.edgeDepth[1] *= lsRatio

		c.bindPos[0] *= lsRatio
		c.bindPos[1] *= lsRatio

		c.localscl = newLs
	}
	var ok bool
	// Check if player is trying to change to a negative state.
	if no < 0 {
		sys.appendToConsole(c.warn() + "attempted to change to negative state")
		if !sys.ignoreMostErrors {
			sys.errLog.Printf("Attempted to change to negative state: P%v:%v\n", pn+1, no)
		}
	}
	// Check if player is trying to change to a state number that exceeds the limit
	if no >= math.MaxInt32 {
		sys.appendToConsole(c.warn() + "changed to out of bounds state number")
		if !sys.ignoreMostErrors {
			sys.errLog.Printf("Changed to out of bounds state number: P%v:%v\n", pn+1, no)
		}
	}
	// Always attempt to change to the state we set to.
	if c.ss.sb, ok = sys.cgi[pn].states[c.ss.no]; !ok {
		sys.appendToConsole(c.warn() + fmt.Sprintf("changed to invalid state %v (from state %v)", no, c.ss.prevno))
		if !sys.ignoreMostErrors {
			sys.errLog.Printf("Invalid state: P%v:%v\n", pn+1, no)
		}
		c.ss.sb = *newStateBytecode(pn)
		c.ss.sb.stateType, c.ss.sb.moveType, c.ss.sb.physics = ST_U, MT_U, ST_U
	}
	// Reset persistent counters for this state (Ikemen chars)
	// This used to belong to (*StateBytecode).init(), but was moved outside there
	// due to a MUGEN 1.1 problem where persistent was not getting reset until the end
	// of a hitpause when attempting to change state during the hitpause.
	// Ikemenver chars aren't affected by this.
	if c.stWgi().ikemenver[0] != 0 || c.stWgi().ikemenver[1] != 0 {
		c.ss.sb.ctrlsps = make([]int32, len(c.ss.sb.ctrlsps))
	}
	c.stchtmp = true
	return true
}

func (c *Char) stateChange2() bool {
	if c.stchtmp && !c.hitPause() {
		c.ss.sb.init(c)
		// Reset persistent counters for this state (MUGEN chars)
		if c.stWgi().ikemenver[0] == 0 && c.stWgi().ikemenver[1] == 0 {
			c.ss.sb.ctrlsps = make([]int32, len(c.ss.sb.ctrlsps))
		}
		// Flag RemoveOnChangeState explods for removal
		for i := range sys.explods[c.playerNo] {
			if sys.explods[c.playerNo][i].playerId == c.id && sys.explods[c.playerNo][i].removeonchangestate {
				sys.explods[c.playerNo][i].statehaschanged = true
			}
		}
		// Stop flagged sound channels
		for i := range c.soundChannels.channels {
			if c.soundChannels.channels[i].stopOnChangeState {
				c.soundChannels.channels[i].Stop()
				c.soundChannels.channels[i].stopOnChangeState = false
			}
		}
		c.stchtmp = false
		return true
	}
	return false
}

func (c *Char) changeStateEx(no int32, pn int, anim, ctrl int32, ffx string) {
	// This is a very specific and undocumented Mugen behavior that was probably superseded by "facep2"
	// It serves very little purpose while negatively affecting some new Ikemen features like NoTurnTarget
	// It could be removed in the future
	// https://github.com/ikemen-engine/Ikemen-GO/issues/1755
	//if c.minus <= 0 && c.scf(SCF_ctrl) && sys.roundState() <= 2 {
	//	c.autoTurn()
	//}
	if anim != -1 {
		c.changeAnim(anim, c.playerNo, -1, ffx)
	}
	if ctrl >= 0 {
		c.setCtrl(ctrl != 0)
	}
	if c.stateChange1(no, pn) && sys.changeStateNest == 0 && c.minus == 0 {
		for c.stchtmp && sys.changeStateNest < MaxLoop {
			c.stateChange2()
			sys.changeStateNest++
			if !c.ss.sb.run(c) {
				break
			}
		}
		sys.changeStateNest = 0
	}
}

func (c *Char) changeState(no, anim, ctrl int32, ffx string) {
	c.changeStateEx(no, c.ss.sb.playerNo, anim, ctrl, ffx)
}

func (c *Char) selfState(no, anim, readplayerid, ctrl int32, ffx string) {
	var playerno int
	if readplayerid >= 0 {
		playerno = int(readplayerid)
	} else {
		playerno = c.playerNo
	}
	c.changeStateEx(no, playerno, anim, ctrl, ffx)
}

func (c *Char) destroy() {
	if c.helperIndex > 0 {
		c.exitTarget()
		c.receivedDmg = 0
		c.receivedHits = 0
		// Remove ID from target's GetHitVars
		for _, tid := range c.targets {
			if t := sys.playerID(tid); t != nil {
				t.ghv.dropId(c.id)
			}
		}
		// Remove ID from parent's children list
		if c.parentIndex >= 0 {
			if p := c.parent(false); p != nil {
				for i, ch := range p.children {
					if ch == c {
						p.children[i] = nil
					}
				}
			}
		}
		// Remove ID from children
		for _, ch := range c.children {
			if ch != nil && ch.parentIndex > 0 {
				ch.parentIndex *= -1
			}
		}
		c.children = c.children[:0]
		if c.playerFlag {
			// sys.charList.p2enemyDelete(c)
			sys.charList.enemyNearChanged = true
		}
		sys.charList.delete(c)
		c.helperIndex = -1
		c.setCSF(CSF_destroy)
	}
}

// Mugen clears the helper ID here, before fully removing the helper (c.helperID = 0)
// We don't so that all helper triggers behave the same
func (c *Char) destroySelf(recursive, removeexplods, removetexts bool) bool {
	if c.helperIndex <= 0 {
		return false
	}

	c.setCSF(CSF_destroy)

	if removeexplods {
		c.removeExplod(-1, -1)
	}

	if removetexts {
		sys.lifebar.RemoveText(-1, c.id)
	}

	if recursive {
		for _, ch := range c.children {
			if ch != nil {
				ch.destroySelf(recursive, removeexplods, removetexts)
			}
		}
	}

	return true
}

// Make a new helper before reading the bytecode parameters
func (c *Char) newHelper() (h *Char) {
	// If any existing helper entry is valid for overwriting, use that one
	i := int32(0)
	for ; int(i) < len(sys.chars[c.playerNo]); i++ {
		if sys.chars[c.playerNo][i].helperIndex < 0 {
			h = sys.chars[c.playerNo][i]
			h.init(c.playerNo, i)
			break
		}
	}
	// Otherwise append to the end
	if int(i) >= len(sys.chars[c.playerNo]) {
		if i >= sys.cfg.Config.HelperMax {
			return
		}
		h = newChar(c.playerNo, i)
		sys.chars[c.playerNo] = append(sys.chars[c.playerNo], h)
	}

	// Init default helper parameters
	h.name = c.name + "'s helper"
	h.id = sys.newCharId()
	h.helperId = 0
	h.ownpal = false
	h.preserve = false
	h.initCnsVar()
	h.mapArray = make(map[string]float32)
	h.remapSpr = make(RemapPreset)

	// Copy some parent parameters
	h.parentIndex = c.helperIndex
	h.controller = c.controller
	h.teamside = c.teamside
	h.size = c.size
	h.life, h.lifeMax = c.lifeMax, c.lifeMax
	h.powerMax = c.powerMax
	h.dizzyPoints, h.dizzyPointsMax = c.dizzyPointsMax, c.dizzyPointsMax
	h.guardPoints, h.guardPointsMax = c.guardPointsMax, c.guardPointsMax
	h.redLife = h.lifeMax
	h.prepareNextRound()

	// Add to player lists
	c.addChild(h)
	sys.charList.add(h)
	return
}

// Init helper after reading the bytecode parameters
func (c *Char) helperInit(h *Char, st int32, pt PosType, x, y, z float32,
	facing int32, rp [2]int32, extmap bool) {
	p := c.helperPos(pt, [...]float32{x, y, z}, facing, &h.facing, h.localscl, false)
	h.setAllPosX(p[0])
	h.setAllPosY(p[1])
	h.setAllPosZ(p[2])
	h.vel = [3]float32{}
	if h.ownpal {
		h.palfx = newPalFX()
		if c.getPalfx().remap == nil {
			c.palfx.remap = c.gi().palettedata.palList.GetPalMap()
		}
		tmp := c.getPalfx().remap
		h.palfx.remap = make([]int, len(tmp))
		copy(h.palfx.remap, tmp)
		c.forceRemapPal(h.palfx, rp)
	} else {
		h.palfx = c.getPalfx()
	}
	if extmap {
		for key, value := range c.mapArray {
			h.mapArray[key] = value
		}
	}
	// Mugen 1.1 behavior if invertblend param is omitted(Only if char mugenversion = 1.1)
	if h.stWgi().mugenver[0] == 1 && h.stWgi().mugenver[1] == 1 && h.stWgi().ikemenver[0] == 0 && h.stWgi().ikemenver[1] == 0 {
		h.palfx.invertblend = -2
	}
	h.changeStateEx(st, c.playerNo, 0, 1, "")
	// Helper ID must be positive
	if h.helperId < 0 {
		sys.appendToConsole(h.warn() + fmt.Sprintf("has negative Helper ID"))
		h.helperId = 0
	}
	// Prepare newly created helper so it can be successfully run later via actionRun() in charList.action()
	h.actionPrepare()
}

func (c *Char) helperPos(pt PosType, pos [3]float32, facing int32,
	dstFacing *float32, localscl float32, isProj bool) (p [3]float32) {
	if facing < 0 {
		*dstFacing *= -1
	}
	switch pt {
	case PT_P1:
		p[0] = c.pos[0]*(c.localscl/localscl) + pos[0]*c.facing
		p[1] = c.pos[1]*(c.localscl/localscl) + pos[1]
		p[2] = c.pos[2]*(c.localscl/localscl) + pos[2]
		*dstFacing *= c.facing
	case PT_P2:
		if p2 := c.p2(); p2 != nil {
			p[0] = p2.pos[0]*(p2.localscl/localscl) + pos[0]*p2.facing
			p[1] = p2.pos[1]*(p2.localscl/localscl) + pos[1]
			p[2] = p2.pos[2]*(p2.localscl/localscl) + pos[2]
			if isProj {
				*dstFacing *= c.facing
			} else {
				*dstFacing *= p2.facing
			}
		}
	case PT_Front, PT_Back:
		if c.facing > 0 && pt == PT_Front || c.facing < 0 && pt == PT_Back {
			p[0] = c.rightEdge() * (c.localscl / localscl)
		} else {
			p[0] = c.leftEdge() * (c.localscl / localscl)
		}
		if c.facing > 0 {
			p[0] += pos[0]
		} else {
			p[0] -= pos[0]
		}
		p[1] = pos[1]
		p[2] = pos[2]
		*dstFacing *= c.facing
	case PT_Left:
		p[0] = c.leftEdge()*(c.localscl/localscl) + pos[0]
		p[1] = pos[1]
		if isProj {
			*dstFacing *= c.facing
		}
		p[2] = pos[2]
	case PT_Right:
		p[0] = c.rightEdge()*(c.localscl/localscl) + pos[0]
		p[1] = pos[1]
		if isProj {
			*dstFacing *= c.facing
		}
		p[2] = pos[2]
	case PT_None:
		p = [3]float32{pos[0], pos[1], pos[2]}
		if isProj {
			*dstFacing *= c.facing
		}
	}
	return
}

// Always append to preserve insertion order
func (c *Char) spawnExplod() (*Explod, int) {
	playerExplods := &sys.explods[c.playerNo]

	// Do nothing if explod limit reached
	if len(*playerExplods) >= sys.cfg.Config.ExplodMax {
		return nil, -1
	}

	e := newExplod()
	*playerExplods = append(*playerExplods, e)
	idx := len(*playerExplods) - 1

	e.initFromChar(c)
	return e, idx
}

func (c *Char) getExplods(id int32) (expls []*Explod) {
	for i := range sys.explods[c.playerNo] {
		e := sys.explods[c.playerNo][i]
		if e.matchId(id, c.id) {
			expls = append(expls, e)
		}
	}
	return
}

func (c *Char) explodDrawPal(e *Explod) [2]int32 {
	if len(e.palfx.remap) == 0 {
		return [2]int32{0, 0}
	}
	return c.getDrawPal(e.palfx.remap[0])
}

// Run final setup before explod goes live
func (c *Char) commitExplod(i int) {
	e := sys.explods[c.playerNo][i]

	// Init animation
	e.setAnim()
	e.setAnimElem()

	// If invalid animation, whole explod becomes invalid
	// Note: If animation is not specified, it defaults to 0. If it is specified but invalid, explod is invalid
	if e.anim == nil {
		e.id = IErr
		return
	}

	// Set up interpolation
	e.start_animelem = e.animelem
	e.start_fLength = e.fLength
	e.start_xshear = e.xshear

	for j := 0; j < 3; j++ {
		if j < 2 {
			e.start_scale[j] = e.scale[j]
			e.start_alpha[j] = e.alpha[j]
		}
		e.start_rot[j] = e.anglerot[j]
	}

	if e.interpolate {
		e.fLength = 0
		for j := 0; j < 3; j++ {
			if e.ownpal {
				e.palfxdef.mul[j] = 256
				e.palfxdef.add[j] = 0
			}
			if j < 2 {
				e.scale[j] = 1
				//if e.blendmode == 1 {
				//if e.trans == TT_add { // Any add?
				e.alpha[j] = 255
				//}
			}
			e.anglerot[j] = 0
		}
		if e.ownpal {
			e.palfxdef.color = 1
			e.palfxdef.hue = 0
		}
	}

	// Init "ownpal" PalFX and RemapPal
	// Note: Must be placed after setting up interpolation
	if e.ownpal {
		if !e.anim.isCommonFX() {
			// Keep parent's remapped palette while resetting PalFX
			parentRemap := make([]int, len(c.getPalfx().remap))
			copy(parentRemap, c.getPalfx().remap)
			e.palfx = newPalFX()
			e.palfx.remap = parentRemap
			e.palfx.PalFXDef = e.palfxdef
			c.forceRemapPal(e.palfx, e.remappal)
		} else {
			e.palfx = newPalFX()
			e.palfx.PalFXDef = e.palfxdef
			e.palfx.remap = nil
		}
	}

	// Emulate legacy ontop behavior
	// Move from the end of the slice to the beginning to invert drawing order
	if e.ontop {
		playerExplods := &sys.explods[c.playerNo]
		copy((*playerExplods)[1:i+1], (*playerExplods)[0:i])
		(*playerExplods)[0] = e
	}

	// Explod ready
	e.anim.UpdateSprite()
}

func (c *Char) explodBindTime(id, time int32) {
	for i := range sys.explods[c.playerNo] {
		e := sys.explods[c.playerNo][i]
		if e.matchId(id, c.id) {
			e.bindtime = time
		}
	}
}

// Marks matching explods invalid and prunes the slice immediately
func (c *Char) removeExplod(id, idx int32) {
	playerExplods := &sys.explods[c.playerNo]
	n := int32(0)

	// Mark matching explods invalid
	for _, e := range *playerExplods {
		if e.matchId(id, c.id) {
			if idx < 0 || idx == n {
				e.id = IErr
				if idx == n {
					break
				}
			}
			n++
		}
	}

	// Compact the slice to remove invalid explods
	tempSlice := (*playerExplods)[:0] // Reuse backing array
	for _, e := range *playerExplods {
		if e.id != IErr {
			tempSlice = append(tempSlice, e)
		}
	}
	*playerExplods = tempSlice
}

// Get animation and apply sprite owner properties to it
func (c *Char) getAnimSprite(animNo int32, animPlayerNo, spritePlayerNo int, ffx string, ownpal bool, fx bool) *Animation {
	// Get raw animation
	a := sys.chars[animPlayerNo][0].getAnim(animNo, ffx, fx)
	if a == nil {
		return nil
	}

	// Apply sprite owner context
	c.animSpriteSetup(a, spritePlayerNo, ffx, ownpal)

	return a
}

// Calls getAnimSprite without the extra anim/sprite playerNo features
// For projectiles essentially
func (c *Char) getSelfAnimSprite(animNo int32, ffx string, ownpal bool, fx bool) *Animation {
	a := c.getAnimSprite(animNo, c.playerNo, c.playerNo, ffx, ownpal, false)

	return a
}

// Same old getAnim, but now without the FFX scale adjustment
func (c *Char) getAnim(n int32, ffx string, fx bool) (a *Animation) {
	if n == -2 {
		return &Animation{}
	}

	if n == -1 {
		return nil
	}

	current_ffx := ffx

	if current_ffx == "f" {
		if c.gi().fightfxPrefix != "" {
			current_ffx = c.gi().fightfxPrefix // Override with the character-specific prefix
		}
	}

	if current_ffx != "" && current_ffx != "s" {
		if sys.ffx[current_ffx] != nil && sys.ffx[current_ffx].fat != nil {
			a = sys.ffx[current_ffx].fat.get(n)
		}
	} else {
		a = c.gi().anim.get(n)
	}

	// Log invalid animations
	if a == nil {
		if fx {
			if current_ffx != "" && current_ffx != "s" {
				sys.appendToConsole(c.warn() + fmt.Sprintf("called invalid action %v %v", strings.ToUpper(ffx), n))
			} else {
				sys.appendToConsole(c.warn() + fmt.Sprintf("called invalid action %v", n))
			}
		} else {
			if current_ffx != "" && current_ffx != "s" {
				sys.appendToConsole(c.warn() + fmt.Sprintf("changed to invalid action %v %v", strings.ToUpper(ffx), n))
			} else {
				sys.appendToConsole(c.warn() + fmt.Sprintf("changed to invalid action %v", n))
			}
		}
		if !sys.ignoreMostErrors {
			str := "Invalid action: "
			if current_ffx != "" && current_ffx != "s" {
				str += strings.ToUpper(current_ffx) + ":"
			} else {
				str += fmt.Sprintf("P%v:", c.playerNo+1)
			}
			sys.errLog.Printf("%v%v\n", str, n)
		}
	}

	return
}

func (c *Char) animSpriteSetup(a *Animation, spritePN int, ffx string, ownpal bool) {
	// Validate parameters
	if a == nil || spritePN < 0 || spritePN >= len(sys.chars) {
		return
	}
	if len(sys.chars[spritePN]) == 0 || len(sys.chars[c.playerNo]) == 0 {
		return
	}

	owner := sys.chars[spritePN][0]
	self := sys.chars[c.playerNo][0]

	if !a.isCommonFX() {
		// Set SFF and palette
		a.sff = sys.cgi[spritePN].sff
		a.palettedata = &sys.cgi[spritePN].palettedata.palList

		// If changing sprites
		if spritePN != c.playerNo {
			// Remap palette to sprite owner's current palette if allowed
			if ownpal {
				ownerPal := owner.drawPal()
				key := [2]int16{int16(ownerPal[0]), int16(ownerPal[1])}

				if di, ok := a.palettedata.PalTable[key]; ok {
					for _, id := range [...]int32{0, 9000} {
						if spr := a.sff.GetSprite(int16(id), 0); spr != nil {
							a.palettedata.Remap(spr.palidx, di)
						}
					}
				}
			}

			// Update sprite scale according to SFF owner
			// We use localcoord to avoid fluctuations while characters are in custom states
			if self.localcoord != 0 {
				a.start_scale[0] *= self.localcoord / owner.localcoord
				a.start_scale[1] *= self.localcoord / owner.localcoord
			}
		}
	} else {
		// Otherwise just adapt scale
		if self.localcoord != 0 {
			a.start_scale[0] /= 320 / self.localcoord
			a.start_scale[1] /= 320 / self.localcoord
		}
	}
}

// Position functions
func (c *Char) setPosX(x float32) {
	if c.pos[0] == x {
		return
	}

	c.pos[0] = x
	// We do this because Mugen is very sensitive to enemy position changes
	// Perhaps what it does is only calculate who "enemynear" is when the trigger is called?
	// "P2" enemy reference is less sensitive than this however, and seems to update only once per frame
	if c.playerFlag {
		sys.charList.enemyNearChanged = true
	} else {
		c.enemyNearP2Clear()
	}
}

func (c *Char) setPosY(y float32) { // This function mostly exists right now so we don't forget to use the other two
	c.pos[1] = y
}

func (c *Char) setPosZ(z float32) {
	if c.pos[2] == z {
		return
	}

	c.pos[2] = z
	// Z distance is also factored into enemy near lists
	if sys.zEnabled() {
		if c.playerFlag {
			sys.charList.enemyNearChanged = true
		} else {
			c.enemyNearP2Clear()
		}
	}
}

func (c *Char) posReset() {
	if c.teamside == -1 || c.playerNo < 0 || c.playerNo >= len(sys.stage.p) {
		c.facing = 1
		c.setAllPosX(0)
		c.setAllPosY(0)
		c.setAllPosZ(0)
	} else {
		c.facing = float32(sys.stage.p[c.playerNo].facing)
		c.setAllPosX((float32(sys.stage.p[c.playerNo].startx) * sys.stage.localscl) / c.localscl)
		c.setAllPosY(float32(sys.stage.p[c.playerNo].starty) * sys.stage.localscl / c.localscl)
		c.setAllPosZ(float32(sys.stage.p[c.playerNo].startz) * sys.stage.localscl / c.localscl)
	}
	c.vel[0] = 0
	c.vel[1] = 0
	c.vel[2] = 0
}

func (c *Char) setAllPosX(x float32) {
	c.oldPos[0], c.interPos[0] = x, x
	c.setPosX(x)
}

func (c *Char) setAllPosY(y float32) {
	c.oldPos[1], c.interPos[1] = y, y
	c.setPosY(y)
}

func (c *Char) setAllPosZ(z float32) {
	c.oldPos[2], c.interPos[2] = z, z
	c.setPosZ(z)
}

func (c *Char) addX(x float32) {
	c.setAllPosX(c.pos[0] + c.facing*x)
}

func (c *Char) addY(y float32) {
	c.setAllPosY(c.pos[1] + y)
}

func (c *Char) addZ(z float32) {
	c.setAllPosZ(c.pos[2] + z)
}

func (c *Char) hitAdd(h int32) {
	if h == 0 {
		return
	}
	c.hitCount += h
	c.uniqHitCount += h
	if len(c.targets) > 0 {
		for _, tid := range c.targets {
			if t := sys.playerID(tid); t != nil {
				t.receivedHits += h
				if c.teamside != -1 {
					sys.lifebar.co[c.teamside].combo += h
				}
			}
		}
	} else if c.teamside != -1 {
		// In Mugen, HitAdd can increase combo count even without targets
		for i, p := range sys.chars {
			if len(p) > 0 && c.teamside == ^i&1 {
				if p[0].receivedHits != 0 || p[0].ss.moveType == MT_H {
					p[0].receivedHits += h
					sys.lifebar.co[c.teamside].combo += h
				}
			}
		}
	}
}

func (c *Char) spawnProjectile() *Projectile {
	var p *Projectile
	playerProjs := &sys.projs[c.playerNo]

	// Reuse inactive projectile slot if available
	for i := range *playerProjs {
		if (*playerProjs)[i].id < 0 {
			p = (*playerProjs)[i]
			break
		}
	}

	// If no inactive projectile was found, append a new one within the max limit
	if p == nil && len(*playerProjs) < sys.cfg.Config.PlayerProjectileMax {
		newP := newProjectile()
		*playerProjs = append(*playerProjs, newP)
		p = newP
	}

	// Set default values
	if p != nil {
		p.initFromChar(c)
	}

	return p
}

// Run final setup before projectile goes live
func (c *Char) commitProjectile(p *Projectile, pt PosType, offx, offy, offz float32,
	op bool, rpg, rpn int32, clsnscale bool) {
	// Set starting position
	pos := c.helperPos(pt, [...]float32{offx, offy, offz}, 1, &p.facing, p.localscl, true)
	p.setAllPos([...]float32{pos[0], pos[1], pos[2]})

	if p.anim < -1 {
		p.anim = 0
	}

	// Get animation with sprite context
	p.ani = c.getSelfAnimSprite(p.anim, p.anim_ffx, true, true)

	if p.ani == nil && c.anim != nil {
		// Fallback: copy character's current animation
		p.ani = &Animation{}
		*p.ani = *c.anim
		p.ani.SetAnimElem(1, 0)
		p.anim = c.animNo
	}

	// Save total hits for later use
	p.totalhits = p.hits

	// Use "doscale" if applicable
	if c.size.proj.doscale != 0 {
		p.scale[0] *= c.size.xscale
		p.scale[1] *= c.size.yscale
	}

	// Default Clsn scale
	if !clsnscale {
		p.clsnScale = c.clsnBaseScale
	}

	// Backward compatibility
	if c.stWgi().ikemenver[0] == 0 && c.stWgi().ikemenver[1] == 0 {
		p.hitdef.chainid = -1
		p.hitdef.nochainid = [8]int32{-1, -1, -1, -1, -1, -1, -1, -1}
	}

	// Facing handling
	p.removefacing = c.facing
	if p.velocity[0] < 0 {
		p.facing *= -1
		p.velocity[0] *= -1
		p.accel[0] *= -1
	}

	// Ownpal
	if op {
		remap := make([]int, len(p.palfx.remap))
		copy(remap, p.palfx.remap)
		p.palfx = newPalFX()
		p.palfx.remap = remap
		c.forceRemapPal(p.palfx, [...]int32{rpg, rpn})
	}
}

func (c *Char) projDrawPal(p *Projectile) [2]int32 {
	if len(p.palfx.remap) == 0 {
		return [2]int32{0, 0}
	}
	return c.getDrawPal(p.palfx.remap[0])
}

func (c *Char) getProjs(id int32) (projs []*Projectile) {
	for i := range sys.projs[c.playerNo] {
		p := sys.projs[c.playerNo][i]
		if p.id >= 0 && (id < 0 || p.id == id) { // Removed projectiles have negative ID
			projs = append(projs, p)
		}
	}
	return
}

func (c *Char) setHitdefDefault(hd *HitDef) {
	hd.playerNo = c.ss.sb.playerNo
	hd.attackerID = c.id

	if !hd.isprojectile {
		c.hitdefTargets = c.hitdefTargets[:0]
	}

	if hd.attr&^int32(ST_MASK) == 0 {
		hd.attr = 0
	}

	if hd.hitonce < 0 {
		if hd.attr&int32(AT_AT) != 0 {
			hd.hitonce = 1
		} else {
			hd.hitonce = 0
		}
	}

	// Set a parameter if it's Nan
	ifnanset := func(dst *float32, src float32) {
		if math.IsNaN(float64(*dst)) {
			*dst = src
		}
	}

	// Set a parameter if it's IErr
	ifierrset := func(dst *int32, src int32) bool {
		if *dst == IErr {
			*dst = src
			return true
		}
		return false
	}

	ifierrset(&hd.guard_pausetime[0], hd.pausetime[0])
	ifierrset(&hd.guard_pausetime[1], hd.pausetime[1])

	// In Mugen this one acts diferent from the documentation
	// Ikemen characters follow the documentation since it makes more sense
	if c.stWgi().ikemenver[0] == 0 && c.stWgi().ikemenver[1] == 0 {
		ifierrset(&hd.guard_hittime, hd.ground_slidetime)
	} else {
		ifierrset(&hd.guard_hittime, hd.ground_hittime)
	}

	ifierrset(&hd.guard_slidetime, hd.guard_hittime)
	ifierrset(&hd.guard_ctrltime, hd.guard_slidetime)
	ifierrset(&hd.airguard_ctrltime, hd.guard_ctrltime)

	ifnanset(&hd.guard_velocity[0], hd.ground_velocity[0])
	ifnanset(&hd.guard_velocity[2], hd.ground_velocity[2])
	ifnanset(&hd.airguard_velocity[0], hd.air_velocity[0]*1.5)
	ifnanset(&hd.airguard_velocity[1], hd.air_velocity[1]*0.5)
	ifnanset(&hd.airguard_velocity[2], hd.air_velocity[2]*1.5)
	ifnanset(&hd.down_velocity[0], hd.air_velocity[0])
	ifnanset(&hd.down_velocity[1], hd.air_velocity[1])
	ifnanset(&hd.down_velocity[2], hd.air_velocity[2])

	ifierrset(&hd.fall_envshake_ampl, -4)
	if hd.air_animtype == RA_Unknown {
		hd.air_animtype = hd.animtype
	}
	if hd.fall_animtype == RA_Unknown {
		if hd.air_animtype >= RA_Up {
			hd.fall_animtype = hd.air_animtype
		} else {
			hd.fall_animtype = RA_Back
		}
	}
	if hd.air_type == HT_Unknown {
		hd.air_type = hd.ground_type
	}

	ifierrset(&hd.forcestand, Btoi(hd.ground_velocity[1] != 0)) // Having a Y velocity causes ForceStand
	ifierrset(&hd.forcecrouch, 0)

	ifierrset(&hd.air_fall, Btoi(hd.ground_fall))

	// Cornerpush defaults to same as respective velocities if character has Ikemenversion, instead of Mugen magic numbers
	if hd.attr&int32(ST_A) != 0 {
		ifnanset(&hd.ground_cornerpush_veloff, 0)
	} else {
		if c.stWgi().ikemenver[0] == 0 && c.stWgi().ikemenver[1] == 0 {
			ifnanset(&hd.ground_cornerpush_veloff, hd.guard_velocity[0]*1.3)
		} else {
			ifnanset(&hd.ground_cornerpush_veloff, hd.ground_velocity[0])
		}
	}
	ifnanset(&hd.air_cornerpush_veloff, hd.ground_cornerpush_veloff)
	ifnanset(&hd.down_cornerpush_veloff, hd.ground_cornerpush_veloff)
	ifnanset(&hd.guard_cornerpush_veloff, hd.ground_cornerpush_veloff)
	ifnanset(&hd.airguard_cornerpush_veloff, hd.ground_cornerpush_veloff)

	// Super attack behaviour
	if hd.attr&int32(AT_AH) != 0 {
		ifierrset(&hd.hitgetpower,
			int32(c.gi().constants["super.attack.lifetopowermul"]*float32(hd.hitdamage)))
		ifierrset(&hd.hitgivepower,
			int32(c.gi().constants["super.gethit.lifetopowermul"]*float32(hd.hitdamage)))
		ifierrset(&hd.dizzypoints,
			int32(c.gi().constants["super.lifetodizzypointsmul"]*float32(hd.hitdamage)))
		ifierrset(&hd.guardpoints,
			int32(c.gi().constants["super.lifetoguardpointsmul"]*float32(hd.hitdamage)))
		ifierrset(&hd.hitredlife,
			int32(c.gi().constants["super.lifetoredlifemul"]*float32(hd.hitdamage)))
		ifierrset(&hd.guardredlife,
			int32(c.gi().constants["super.lifetoredlifemul"]*float32(hd.guarddamage)))
	} else {
		ifierrset(&hd.hitgetpower,
			int32(c.gi().constants["default.attack.lifetopowermul"]*float32(hd.hitdamage)))
		ifierrset(&hd.hitgivepower,
			int32(c.gi().constants["default.gethit.lifetopowermul"]*float32(hd.hitdamage)))
		ifierrset(&hd.dizzypoints,
			int32(c.gi().constants["default.lifetodizzypointsmul"]*float32(hd.hitdamage)))
		ifierrset(&hd.guardpoints,
			int32(c.gi().constants["default.lifetoguardpointsmul"]*float32(hd.hitdamage)))
		ifierrset(&hd.hitredlife,
			int32(c.gi().constants["default.lifetoredlifemul"]*float32(hd.hitdamage)))
		ifierrset(&hd.guardredlife,
			int32(c.gi().constants["default.lifetoredlifemul"]*float32(hd.guarddamage)))
	}

	ifierrset(&hd.guardgetpower, int32(float32(hd.hitgetpower)*0.5))
	ifierrset(&hd.guardgivepower, int32(float32(hd.hitgivepower)*0.5))

	if !math.IsNaN(float64(hd.snap[0])) {
		hd.maxdist[0], hd.mindist[0] = hd.snap[0], hd.snap[0]
	}
	if !math.IsNaN(float64(hd.snap[1])) {
		hd.maxdist[1], hd.mindist[1] = hd.snap[1], hd.snap[1]
	}
	if !math.IsNaN(float64(hd.snap[2])) {
		hd.maxdist[2], hd.mindist[2] = hd.snap[2], hd.snap[2]
	}

	if hd.teamside == -1 {
		hd.teamside = c.teamside + 1
	}

	if hd.p2clsncheck < 0 {
		if hd.reversal_attr != 0 {
			hd.p2clsncheck = 1
		} else {
			hd.p2clsncheck = 2
		}
	}

	if hd.unhittabletime[0] == IErr || hd.unhittabletime[1] == IErr {
		extra := hd.pausetime[0] + 1
		// In Mugen, Reversaldef makes the target invincible for 1 frame (but not the attacker)
		if hd.reversal_attr != 0 {
			hd.unhittabletime[1] = extra
		}
		// In Mugen, a throw attribute sets this to 1 for both p1 and p2
		if hd.attr&int32(AT_AT) != 0 {
			hd.unhittabletime[0] = extra
			hd.unhittabletime[1] = extra
		}
		// Defaults
		ifierrset(&hd.unhittabletime[0], -1)
		ifierrset(&hd.unhittabletime[1], -1)
	}

	// In Mugen, only projectiles can use air.juggle
	// Ikemen characters can use it to update their StateDef juggle points
	if hd.air_juggle == IErr {
		hd.air_juggle = 0
	} else if !hd.isprojectile && (c.stWgi().ikemenver[0] != 0 || c.stWgi().ikemenver[1] != 0) {
		c.juggle = hd.air_juggle
	}
}

func (c *Char) baseWidthFront() float32 {
	switch c.ss.stateType {
	case ST_C:
		return float32(c.size.crouchbox[2])
	case ST_A:
		return float32(c.size.airbox[2])
	case ST_L:
		return float32(c.size.downbox[2])
	default:
		return float32(c.size.standbox[2])
	}
}

// Because dimensions are positive we will invert the constants here
func (c *Char) baseWidthBack() float32 {
	switch c.ss.stateType {
	case ST_C:
		return -float32(c.size.crouchbox[0])
	case ST_A:
		return -float32(c.size.airbox[0])
	case ST_L:
		return -float32(c.size.downbox[0])
	default:
		return -float32(c.size.standbox[0])
	}
}

// Because dimensions are positive we will invert the constants here
func (c *Char) baseHeightTop() float32 {
	switch c.ss.stateType {
	case ST_C:
		return -float32(c.size.crouchbox[1])
	case ST_A:
		return -float32(c.size.airbox[1])
	case ST_L:
		return -float32(c.size.downbox[1])
	default:
		return -float32(c.size.standbox[1])
	}
}

func (c *Char) baseHeightBottom() float32 {
	switch c.ss.stateType {
	case ST_C:
		return float32(c.size.crouchbox[3])
	case ST_A:
		return float32(c.size.airbox[3])
	case ST_L:
		return float32(c.size.downbox[3])
	default:
		return float32(c.size.standbox[3])
	}
}

func (c *Char) baseDepthTop() float32 {
	return float32(c.size.depth[0])
}

func (c *Char) baseDepthBottom() float32 {
	return float32(c.size.depth[1])
}

func (c *Char) setWidth(fw, bw float32) {
	coordRatio := (320 / c.localcoord) / c.localscl

	c.sizeWidth[0] = c.baseWidthFront()*coordRatio + fw
	c.sizeWidth[1] = c.baseWidthBack()*coordRatio + bw

	c.updateSizeBox()
	c.setCSF(CSF_width)
}

func (c *Char) setHeight(th, bh float32) {
	coordRatio := (320 / c.localcoord) / c.localscl

	c.sizeHeight[0] = c.baseHeightTop()*coordRatio + th
	c.sizeHeight[1] = c.baseHeightBottom()*coordRatio + bh

	c.updateSizeBox()
	c.setCSF(CSF_height)
}

func (c *Char) setDepth(td, bd float32) {
	coordRatio := (320 / c.localcoord) / c.localscl

	c.sizeDepth[0] = c.baseDepthTop()*coordRatio + td
	c.sizeDepth[1] = c.baseDepthBottom()*coordRatio + bd

	c.setCSF(CSF_depth)
}

func (c *Char) setWidthEdge(fe, be float32) {
	// TODO: confirm if these don't need "coordRatio"
	c.edgeWidth = [2]float32{fe, be}
	c.setCSF(CSF_widthedge)
}

func (c *Char) setDepthEdge(tde, bde float32) {
	c.edgeDepth[0] = tde
	c.edgeDepth[1] = bde
	c.setCSF(CSF_depthedge)
}

func (c *Char) updateClsnScale() {
	// Update base scale
	if c.ownclsnscale && c.animPN == c.playerNo {
		// Helper parameter. Use own scale instead of animation owner's
		c.clsnBaseScale = [...]float32{c.size.xscale, c.size.yscale}
	} else if c.animPN >= 0 && c.animPN < len(sys.chars) && len(sys.chars[c.animPN]) > 0 {
		// Index range checks. Prevents crashing if chars don't have animations
		// https://github.com/ikemen-engine/Ikemen-GO/issues/1982
		// The char's base Clsn scale is based on the animation owner's scale constants
		c.clsnBaseScale = [...]float32{
			sys.chars[c.animPN][0].size.xscale,
			sys.chars[c.animPN][0].size.yscale,
		}
	} else {
		// Normally not used. Just a safeguard
		c.clsnBaseScale = [...]float32{1.0, 1.0}
	}
	// Calculate final scale
	// Clsn and size box scale used to factor zScale here, but they shouldn't
	// Game logic should stay the same regardless of Z scale. Only drawing scale should change
	c.clsnScale = [2]float32{c.clsnBaseScale[0] * c.clsnScaleMul[0] * c.animlocalscl, // Facing is not used here
		c.clsnBaseScale[1] * c.clsnScaleMul[1] * c.animlocalscl}
}

// Convert size variables to a Clsn-like box
// This box will replace width and height values in some other parts of the code
func (c *Char) updateSizeBox() {
	// Correct left/right and top/bottom
	// Same behavior as Clsn boxes
	// https://github.com/ikemen-engine/Ikemen-GO/issues/2008
	back := -c.sizeWidth[1]
	front := c.sizeWidth[0]
	if back > front {
		back, front = front, back
	}
	top := -c.sizeHeight[0]
	bottom := c.sizeHeight[1]
	if top > bottom { // Negative sign
		top, bottom = bottom, top
	}
	c.sizeBox = [4]float32{back, top, front, bottom}
}

// Returns the size box in the same format as Clsn boxes
func (c *Char) sizeBoxToClsn() [][4]float32 {
	return [][4]float32{c.sizeBox}
}

func (c *Char) gethitAnimtype() Reaction {
	if c.ghv.fallflag {
		return c.ghv.fall_animtype
	} else if c.ss.stateType == ST_A {
		return c.ghv.airanimtype
	} else {
		if c.ghv.groundanimtype >= RA_Back && c.ghv.yvel == 0 {
			return RA_Hard
		} else {
			return c.ghv.groundanimtype
		}
	}
}

func (c *Char) isTargetBound() bool {
	return c.ghv.idMatch(c.bindToId)
}

func (c *Char) initCnsVar() {
	c.cnsvar = make(map[int32]int32)
	c.cnsfvar = make(map[int32]float32)
	c.cnssysvar = make(map[int32]int32)
	c.cnssysfvar = make(map[int32]float32)
}

func (c *Char) varGet(i int32) BytecodeValue {
	if i < 0 {
		sys.appendToConsole(c.warn() + fmt.Sprintf("var index %v must be positive", i))
		return BytecodeSF()
	}
	// Check var (map)
	val, ok := c.cnsvar[i]
	// If key found
	if ok {
		return BytecodeInt(val)
	}
	// If var not set yet
	return BytecodeInt(0)
}

func (c *Char) fvarGet(i int32) BytecodeValue {
	if i < 0 {
		sys.appendToConsole(c.warn() + fmt.Sprintf("fvar index %v must be positive", i))
		return BytecodeSF()
	}

	val, ok := c.cnsfvar[i]
	if ok {
		return BytecodeFloat(val)
	}
	return BytecodeFloat(0)
}

func (c *Char) sysVarGet(i int32) BytecodeValue {
	if i < 0 {
		sys.appendToConsole(c.warn() + fmt.Sprintf("sysvar index %v must be positive", i))
		return BytecodeSF()
	}

	val, ok := c.cnssysvar[i]
	if ok {
		return BytecodeInt(val)
	}
	return BytecodeInt(0)
}

func (c *Char) sysFvarGet(i int32) BytecodeValue {
	if i < 0 {
		sys.appendToConsole(c.warn() + fmt.Sprintf("sysfvar index %v must be positive", i))
		return BytecodeSF()
	}

	val, ok := c.cnssysfvar[i]
	if ok {
		return BytecodeFloat(val)
	}
	return BytecodeFloat(0)
}

func (c *Char) varSet(i, v int32) BytecodeValue {
	if i < 0 {
		sys.appendToConsole(c.warn() + fmt.Sprintf("var index %v must be positive", i))
		return BytecodeSF()
	}

	c.cnsvar[i] = v // Create or update the key
	return BytecodeInt(v)
}

func (c *Char) fvarSet(i int32, v float32) BytecodeValue {
	if i < 0 {
		sys.appendToConsole(c.warn() + fmt.Sprintf("fvar index %v must be positive", i))
		return BytecodeSF()
	}

	c.cnsfvar[i] = v
	return BytecodeFloat(v)
}

func (c *Char) sysVarSet(i, v int32) BytecodeValue {
	if i < 0 {
		sys.appendToConsole(c.warn() + fmt.Sprintf("sysvar index %v must be positive", i))
		return BytecodeSF()
	}

	c.cnssysvar[i] = v
	return BytecodeInt(v)
}

func (c *Char) sysFvarSet(i int32, v float32) BytecodeValue {
	if i < 0 {
		sys.appendToConsole(c.warn() + fmt.Sprintf("sysfvar index %v must be positive", i))
		return BytecodeSF()
	}

	c.cnssysfvar[i] = v
	return BytecodeFloat(v)
}

func (c *Char) varAdd(i, v int32) BytecodeValue {
	if i < 0 {
		sys.appendToConsole(c.warn() + fmt.Sprintf("var index %v must be positive", i))
		return BytecodeSF()
	}

	if _, ok := c.cnsvar[i]; ok {
		c.cnsvar[i] += v
	} else {
		c.cnsvar[i] = v
	}
	return BytecodeInt(c.cnsvar[i])
}

func (c *Char) fvarAdd(i int32, v float32) BytecodeValue {
	if i < 0 {
		sys.appendToConsole(c.warn() + fmt.Sprintf("fvar index %v must be positive", i))
		return BytecodeSF()
	}

	if _, ok := c.cnsfvar[i]; ok {
		c.cnsfvar[i] += v
	} else {
		c.cnsfvar[i] = v
	}
	return BytecodeFloat(c.cnsfvar[i])
}

func (c *Char) sysVarAdd(i, v int32) BytecodeValue {
	if i < 0 {
		sys.appendToConsole(c.warn() + fmt.Sprintf("sysvar index %v must be positive", i))
		return BytecodeSF()
	}

	if _, ok := c.cnssysvar[i]; ok {
		c.cnssysvar[i] += v
	} else {
		c.cnssysvar[i] = v
	}
	return BytecodeInt(c.cnssysvar[i])
}

func (c *Char) sysFvarAdd(i int32, v float32) BytecodeValue {
	if i < 0 {
		sys.appendToConsole(c.warn() + fmt.Sprintf("sysfvar index %v must be positive", i))
		return BytecodeSF()
	}

	if _, ok := c.cnssysfvar[i]; ok {
		c.cnssysfvar[i] += v
	} else {
		c.cnssysfvar[i] = v
	}
	return BytecodeFloat(c.cnssysfvar[i])
}

func (c *Char) varRangeSet(first, last, val int32) {
	if first < 0 || first > last {
		return
	}

	loopCount := 0
	if val == 0 {
		// Delete existing maps within range. Don't make new ones
		for k := range c.cnsvar {
			if k >= first && k <= last {
				delete(c.cnsvar, k)
				loopCount++
				if loopCount >= MaxLoop {
					sys.appendToConsole(c.warn() + fmt.Sprintf("VarRangeSet limit reached after setting %v variables", loopCount))
					break
				}
			}
		}
	} else {
		// Set entire map range to value
		for i := first; i <= last; i++ {
			c.cnsvar[i] = val
			loopCount++
			if loopCount >= MaxLoop {
				sys.appendToConsole(c.warn() + fmt.Sprintf("VarRangeSet limit reached after setting %v variables", loopCount))
				break
			}
		}
	}
}

func (c *Char) fvarRangeSet(first, last int32, val float32) {
	if first < 0 || first > last {
		return
	}

	if val == 0 {
		for k := range c.cnsfvar {
			if k >= first && k <= last {
				delete(c.cnsfvar, k)
			}
		}
	} else {
		for i := first; i <= last; i++ {
			c.cnsfvar[i] = val
		}
	}
}

func (c *Char) sysVarRangeSet(first, last, val int32) {
	if first < 0 || first > last {
		return
	}

	if val == 0 {
		for k := range c.cnssysvar {
			if k >= first && k <= last {
				delete(c.cnssysvar, k)
			}
		}
	} else {
		for i := first; i <= last; i++ {
			c.cnssysvar[i] = val
		}
	}
}

func (c *Char) sysFvarRangeSet(first, last int32, val float32) {
	if first < 0 || first > last {
		return
	}

	if val == 0 {
		for k := range c.cnssysfvar {
			if k >= first && k <= last {
				delete(c.cnssysfvar, k)
			}
		}
	} else {
		for i := first; i <= last; i++ {
			c.cnssysfvar[i] = val
		}
	}
}

func (c *Char) setFacing(f float32) {
	if f != 0 {
		if (c.facing < 0) != (f < 0) {
			c.facing *= -1
			c.vel[0] *= -1
		}
	}
}

// Get stage BG elements for StageBGVar trigger
func (c *Char) getStageBg(id int32, idx int, log bool) *backGround {
	// Invalid index
	if idx < 0 {
		if log {
			sys.appendToConsole(c.warn() + "stage BG element index cannot be negative")
		}
		return nil
	}

	// Filter background elements with the specified ID
	var filteredBg []*backGround
	for _, bg := range sys.stage.bg {
		if id < 0 || id == bg.id {
			filteredBg = append(filteredBg, bg)
			// Background element found at requested index
			if idx >= 0 && len(filteredBg) == idx+1 {
				return filteredBg[idx]
			}
		}
	}

	// No valid background element found
	if log {
		sys.appendToConsole(c.warn() + fmt.Sprintf("found no stage BG element with ID %v and index %v", id, idx))
	}
	return nil
}

// Get multiple stage BG elements for ModifyStageBG sctrl
func (c *Char) getMultipleStageBg(id int32, idx int, log bool) []*backGround {
	// Filter background elements with the specified ID
	var filteredBg []*backGround
	for _, bg := range sys.stage.bg {
		if id < 0 || id == bg.id {
			filteredBg = append(filteredBg, bg)
			// If idx is valid and we've reached the requested index, return the single element
			if idx >= 0 && len(filteredBg) == idx+1 {
				return []*backGround{filteredBg[idx]}
			}
		}
	}

	// Return multiple instances if idx is negative
	if idx < 0 {
		return filteredBg
	}

	// No valid background element found
	if log {
		sys.appendToConsole(c.warn() + fmt.Sprintf("found no stage BG element with ID %v and index %v", id, idx))
	}
	return nil
}

// For NumStageBG trigger
func (c *Char) numStageBG(id BytecodeValue) BytecodeValue {
	if id.IsSF() {
		return BytecodeSF()
	}

	bid := id.ToI()
	n := 0

	// We do this instead of getMultipleStageBg because that one returns actual BG pointers
	for _, bg := range sys.stage.bg {
		if bid < 0 || bid == bg.id {
			n++
		}
	}

	return BytecodeInt(int32(n))
}

// Get list of targets for the Target state controllers
func (c *Char) getTarget(id int32, idx int) []int32 {
	// If ID and index are negative, just return all targets
	// In Mugen the ID must be specifically -1
	if id < 0 && idx < 0 {
		return c.targets
	}

	// Filter targets with the specified ID
	var filteredTargets []int32
	for _, tid := range c.targets {
		if t := sys.playerID(tid); t != nil && (id < 0 || t.ghv.hitid == id) {
			filteredTargets = append(filteredTargets, tid)
		}
		// Target found at requested index
		if idx >= 0 && len(filteredTargets) == idx+1 {
			return []int32{filteredTargets[idx]}
		}
	}

	// If index is negative, return all targets with specified ID
	if idx < 0 {
		return filteredTargets
	}

	// No valid target found
	return nil
}

func (c *Char) targetFacing(tar []int32, f int32) {
	if f == 0 {
		return
	}
	tf := c.facing
	if f < 0 {
		tf *= -1
	}
	for _, tid := range tar {
		if t := sys.playerID(tid); t != nil {
			t.setFacing(tf)
		}
	}
}

func (c *Char) targetBind(tar []int32, time int32, x, y, z float32) {
	for _, tid := range tar {
		if t := sys.playerID(tid); t != nil {
			t.setBindToId(c, true)
			t.setBindTime(time)
			t.bindFacing = 0
			x *= c.localscl / t.localscl
			y *= c.localscl / t.localscl
			z *= c.localscl / t.localscl
			t.bindPos = [...]float32{x, y, z}
		}
	}
}

func (c *Char) bindToTarget(tar []int32, time int32, x, y, z float32, hmf HMF) {
	if len(tar) > 0 {
		if t := sys.playerID(tar[0]); t != nil {
			switch hmf {
			case HMF_M:
				x += t.size.mid.pos[0] * ((320 / t.localcoord) / c.localscl)
				y += t.size.mid.pos[1] * ((320 / t.localcoord) / c.localscl)
			case HMF_H:
				x += t.size.head.pos[0] * ((320 / t.localcoord) / c.localscl)
				y += t.size.head.pos[1] * ((320 / t.localcoord) / c.localscl)
			}
			if !math.IsNaN(float64(x)) {
				c.setAllPosX(t.pos[0]*(t.localscl/c.localscl) + t.facing*x)
			}
			if !math.IsNaN(float64(y)) {
				c.setAllPosY(t.pos[1]*(t.localscl/c.localscl) + y)
			}
			if !math.IsNaN(float64(z)) {
				c.setAllPosZ(t.pos[2]*(t.localscl/c.localscl) + z)
			}
			c.targetBind(tar[:1], time,
				c.facing*c.distX(t, c),
				(t.pos[1]*(t.localscl/c.localscl))-(c.pos[1]*(c.localscl/t.localscl)),
				(t.pos[2]*(t.localscl/c.localscl))-(c.pos[2]*(c.localscl/t.localscl)))
		}
	}
}

func (c *Char) targetLifeAdd(tar []int32, add int32, kill, absolute, dizzy, redlife bool) {
	if add == 0 {
		return
	}
	for _, tid := range tar {
		if t := sys.playerID(tid); t != nil {
			// We flip the sign of "add" so that it operates under the same logic as Hitdef damage
			// Note: LifeAdd and similar state controllers always ignore the attack multiplier
			dmg := float64(t.computeDamage(-float64(add), kill, absolute, 1, c, true))
			// Subtract life
			t.lifeAdd(-dmg, true, true)
			// Subtract red life
			if redlife {
				if t.ghv.attr&int32(AT_AH) != 0 {
					t.redLifeAdd(-dmg*float64(c.gi().constants["super.lifetoredlifemul"]), true)
				} else {
					t.redLifeAdd(-dmg*float64(c.gi().constants["default.lifetoredlifemul"]), true)
				}
			}
			// Subtract dizzy points
			if dizzy && !t.scf(SCF_dizzy) && !t.asf(ASF_nodizzypointsdamage) {
				if t.ghv.attr&int32(AT_AH) != 0 {
					t.dizzyPointsAdd(-dmg*float64(c.gi().constants["super.lifetodizzypointsmul"]), true)
				} else {
					t.dizzyPointsAdd(-dmg*float64(c.gi().constants["default.lifetodizzypointsmul"]), true)
				}
			}
			t.ghv.kill = kill
		}
	}
}

func (c *Char) targetPowerAdd(tar []int32, power int32) {
	if power == 0 {
		return
	}
	for _, tid := range tar {
		if t := sys.playerID(tid); t != nil && t.playerFlag {
			t.powerAdd(power)
		}
	}
}

func (c *Char) targetDizzyPointsAdd(tar []int32, add int32, absolute bool) {
	if add == 0 {
		return
	}
	for _, tid := range tar {
		if t := sys.playerID(tid); t != nil && !t.scf(SCF_dizzy) && !t.asf(ASF_nodizzypointsdamage) {
			t.dizzyPointsAdd(float64(t.computeDamage(float64(add), false, absolute, 1, c, false)), true)
		}
	}
}

func (c *Char) targetGuardPointsAdd(tar []int32, add int32, absolute bool) {
	if add == 0 {
		return
	}
	for _, tid := range tar {
		if t := sys.playerID(tid); t != nil && !t.asf(ASF_noguardpointsdamage) {
			t.guardPointsAdd(float64(t.computeDamage(float64(add), false, absolute, 1, c, false)), true)
		}
	}
}

func (c *Char) targetRedLifeAdd(tar []int32, add int32, absolute bool) {
	if add == 0 {
		return
	}
	for _, tid := range tar {
		if t := sys.playerID(tid); t != nil && !t.asf(ASF_noredlifedamage) {
			t.redLifeAdd(float64(t.computeDamage(float64(add), false, absolute, 1, c, true)), true)
		}
	}
}

func (c *Char) targetScoreAdd(tar []int32, s float32) {
	if s == 0 {
		return
	}
	for _, tid := range tar {
		if t := sys.playerID(tid); t != nil && t.playerFlag {
			t.scoreAdd(s)
		}
	}
}

func (c *Char) targetState(tar []int32, state int32) {
	if len(tar) > 0 && state >= 0 {
		pn := c.ss.sb.playerNo
		if c.minus == -2 || c.minus == -4 {
			pn = c.playerNo
		}
		for _, tid := range tar {
			if t := sys.playerID(tid); t != nil {
				t.setCtrl(false)
				t.stateChange1(state, pn)
			}
		}
	}
}

func (c *Char) targetVelSetX(tar []int32, x float32) {
	for _, tid := range tar {
		if t := sys.playerID(tid); t != nil {
			x *= c.localscl / t.localscl
			t.vel[0] = x
		}
	}
}

func (c *Char) targetVelSetY(tar []int32, y float32) {
	for _, tid := range tar {
		if t := sys.playerID(tid); t != nil {
			y *= c.localscl / t.localscl
			t.vel[1] = y
		}
	}
}

func (c *Char) targetVelSetZ(tar []int32, z float32) {
	for _, tid := range tar {
		if t := sys.playerID(tid); t != nil {
			z *= c.localscl / t.localscl
			t.vel[2] = z
		}
	}
}

func (c *Char) targetVelAddX(tar []int32, x float32) {
	for _, tid := range tar {
		if t := sys.playerID(tid); t != nil {
			x *= c.localscl / t.localscl
			t.vel[0] += x
		}
	}
}

func (c *Char) targetVelAddY(tar []int32, y float32) {
	for _, tid := range tar {
		if t := sys.playerID(tid); t != nil {
			y *= c.localscl / t.localscl
			t.vel[1] += y
		}
	}
}

func (c *Char) targetVelAddZ(tar []int32, z float32) {
	for _, tid := range tar {
		if t := sys.playerID(tid); t != nil {
			z *= c.localscl / t.localscl
			t.vel[2] += z
		}
	}
}

func (c *Char) targetDrop(excludeid int32, excludechar int32, keepone bool) {
	var tg []int32
	// Keep the player with this "player ID". Used with "HitOnce" attacks such as throws
	if keepone && excludechar > 0 {
		for _, tid := range c.targets {
			if t := sys.playerID(tid); t != nil {
				if t.id == excludechar {
					tg = append(tg, tid)
				} else {
					t.gethitBindClear()
					t.ghv.dropId(c.id)
				}
			}
		}
		c.targets = tg
		return
	}
	// Keep the players with this "hit ID". Used with "TargetDrop" state controller
	if excludeid < 0 {
		tg = c.targets
	} else {
		for _, tid := range c.targets {
			if t := sys.playerID(tid); t != nil {
				if t.ghv.hitid == excludeid {
					tg = append(tg, tid)
				} else {
					t.gethitBindClear()
					t.ghv.dropId(c.id)
				}
			}
		}
	}
	// If more than one target still remains and "keepone" is true, pick one to keep at random
	if (keepone || excludeid < 0) && len(tg) > 0 {
		c.targets = nil
		r := -1
		if keepone && excludeid >= 0 {
			r = int(Rand(0, int32(len(tg))-1))
		}
		for i, tid := range tg {
			if i == r {
				c.targets = append(c.targets, tid)
			} else if t := sys.playerID(tid); t != nil {
				if t.isTargetBound() {
					if c.csf(CSF_gethit) {
						t.selfState(5050, -1, -1, -1, "")
					}
					t.setBindTime(0)
				}
				t.ghv.dropId(c.id)
			}
		}
	} else {
		c.targets = tg
	}
}

// Process raw damage into the value that will actually be used
// Calculations are done in float64 for the sake of precision
func (c *Char) computeDamage(damage float64, kill, absolute bool, atkmul float32, attacker *Char, bounds bool) int32 {
	// Skip further calculations
	if damage == 0 || !absolute && atkmul == 0 {
		return 0
	}
	// Apply attack and defense multipliers
	if !absolute {
		damage *= float64(atkmul) / c.finalDefense
	}
	// In Mugen, an extremely high defense or low attack still results in at least 1 damage. Not true when healing
	if damage > 0 && damage < 1 {
		damage = 1
	}
	// Normally damage cannot exceed the char's remaining life
	if bounds && damage > float64(c.life) {
		damage = float64(c.life)
	}
	// Limit damage if kill is false
	// In Mugen, if a character attacks a char with 0 life and kill = 0, the attack will actually heal 1 point
	// https://github.com/ikemen-engine/Ikemen-GO/issues/1200
	if !kill && damage >= float64(c.life) {
		// If a Mugen character attacks a char with 0 life and kill = 0, the attack will actually heal
		if c.life > 0 || c.stWgi().ikemenver[0] == 0 && c.stWgi().ikemenver[1] == 0 {
			damage = float64(c.life - 1)
		}
	}
	// Safely convert from float64 back to int32 after all calculations are done
	int := F64toI32(math.Round(damage))
	return int
}

func (c *Char) lifeAdd(add float64, kill, absolute bool) {
	if add == 0 {
		return
	}
	if !absolute {
		add /= c.finalDefense
	}
	// In Mugen, an extremely high defense or low attack still results in at least 1 damage. Not true when healing
	if add > -1 && add < 0 {
		add = -1
	}

	add_i64 := int64(math.Round(add))
	prev_life := c.life
	new_life_i64 := int64(prev_life) + add_i64
	new_life_i32 := int32(new_life_i64)

	// MUGEN Overflow/Underflow damage compatibility
	// For healing (positive add), if it overflows and becomes negative, it's a KO.
	// For damage (negative add), if it underflows and becomes positive, it's also a KO.
	if (add_i64 > 0 && new_life_i32 < prev_life) || (add_i64 < 0 && new_life_i32 > prev_life) {
		new_life_i32 = 0 // Overflow/Underflow results in KO.
	}

	// Limit value if kill is false
	if !kill && new_life_i32 <= 0 {
		if c.life > 0 || (c.stWgi().ikemenver[0] == 0 && c.stWgi().ikemenver[1] == 0) {
			new_life_i32 = 1
		} else {
			new_life_i32 = 0
		}
	}
	if add_i64 < 0 {
		c.receivedDmg += Min(c.life, int32(-add_i64))
	}
	c.lifeSet(new_life_i32)
	// Using LifeAdd currently does not touch the red life value
	// This could be expanded in the future, as with TargetLifeAdd
}

func (c *Char) lifeSet(life int32) {
	if c.alive() && sys.roundNoDamage() {
		return
	}
	c.life = Clamp(life, 0, c.lifeMax)
	if c.life == 0 {
		// Check win type
		if c.playerFlag && c.teamside != -1 {
			if c.alive() && c.helperIndex == 0 {
				if c.ss.moveType != MT_H {
					if c.playerNo == c.ss.sb.playerNo {
						sys.winType[^c.playerNo&1] = WT_Suicide
					} else if c.playerNo&1 == c.ss.sb.playerNo&1 {
						sys.winType[^c.playerNo&1] = WT_Teammate
					}
				} else if c.playerNo == c.ghv.playerNo {
					sys.winType[^c.playerNo&1] = WT_Suicide
				} else if c.ghv.playerNo >= 0 && c.playerNo&1 == c.ghv.playerNo&1 {
					sys.winType[^c.playerNo&1] = WT_Teammate
				} else if c.ghv.cheeseKO {
					sys.winType[^c.playerNo&1] = WT_Cheese
				} else if c.ghv.attr&int32(AT_AH) != 0 {
					sys.winType[^c.playerNo&1] = WT_Hyper
				} else if c.ghv.attr&int32(AT_AS) != 0 {
					sys.winType[^c.playerNo&1] = WT_Special
				} else if c.ghv.attr&int32(AT_AT) != 0 {
					sys.winType[^c.playerNo&1] = WT_Throw
				} else {
					sys.winType[^c.playerNo&1] = WT_Normal
				}
			}
		} else if c.immortal { // in mugen even non-player helpers can die
			c.life = 1
		}
		c.redLife = 0
	}
	if c.teamside != c.ghv.playerNo&1 && c.teamside != -1 && c.ghv.playerNo < MaxSimul*2 { // attacker and receiver from opposite teams
		sys.lastHitter[^c.playerNo&1] = c.ghv.playerNo
	}
	// Disable red life. Placing this here makes it never lag behind life
	if !c.redLifeEnabled() {
		c.redLife = c.life
	}
}

func (c *Char) setPower(pow int32) {
	// In Mugen, power cannot be changed at all after the round ends
	// TODO: This is probably too restrictive
	if sys.intro < 0 {
		return
	}
	if sys.maxPowerMode {
		c.power = c.powerMax
	} else {
		c.power = Clamp(pow, 0, c.powerMax)
	}
}

func (c *Char) powerAdd(add int32) {
	if add == 0 {
		return
	}
	// Safely convert from float64 back to int32 after all calculations are done
	int := F64toI32(float64(c.getPower()) + math.Round(float64(add)))
	c.powerOwner().setPower(int)
}

// This is only for the PowerSet state controller
func (c *Char) powerSet(pow int32) {
	c.powerOwner().setPower(pow)
}

func (c *Char) dizzyPointsAdd(add float64, absolute bool) {
	if add == 0 {
		return
	}
	if !absolute {
		add /= c.finalDefense
	}
	// Safely convert from float64 back to int32 after all calculations are done
	int := F64toI32(float64(c.dizzyPoints) + math.Round(add))
	c.dizzyPointsSet(int)
}

func (c *Char) dizzyPointsSet(set int32) {
	if c.dizzyEnabled() && !sys.roundNoDamage() {
		c.dizzyPoints = Clamp(set, 0, c.dizzyPointsMax)
	}
}

func (c *Char) guardPointsAdd(add float64, absolute bool) {
	if add == 0 {
		return
	}
	if !absolute {
		add /= c.finalDefense
	}
	// Safely convert from float64 back to int32 after all calculations are done
	int := F64toI32(float64(c.guardPoints) + math.Round(add))
	c.guardPointsSet(int)
}

func (c *Char) guardPointsSet(set int32) {
	if c.guardBreakEnabled() && !sys.roundNoDamage() {
		c.guardPoints = Clamp(set, 0, c.guardPointsMax)
	}
}

func (c *Char) redLifeAdd(add float64, absolute bool) {
	if add == 0 {
		return
	}
	if !absolute {
		add /= c.finalDefense
	}
	// Safely convert from float64 back to int32 after all calculations are done
	int := F64toI32(float64(c.redLife) + math.Round(add))
	c.redLifeSet(int)
}

func (c *Char) redLifeSet(set int32) {
	if !c.alive() {
		c.redLife = 0
	} else if c.redLifeEnabled() && !sys.roundNoDamage() {
		c.redLife = Clamp(set, c.life, c.lifeMax)
	}
}

func (c *Char) score() float32 {
	if c.teamside == -1 {
		return 0
	}
	return sys.lifebar.sc[c.teamside].scorePoints
}

func (c *Char) scoreAdd(val float32) {
	if val == 0 || c.teamside == -1 {
		return
	}
	sys.lifebar.sc[c.teamside].scorePoints += val
}

func (c *Char) scoreTotal() float32 {
	if c.teamside == -1 {
		return 0
	}
	s := sys.scoreStart[c.teamside]
	for _, v := range sys.scoreRounds {
		s += v[c.teamside]
	}
	if !sys.postMatchFlg {
		s += c.score()
	}
	return s
}

func (c *Char) consecutiveWins() int32 {
	if c.teamside == -1 {
		return 0
	}
	return sys.consecutiveWins[c.teamside]
}

func (c *Char) dizzyEnabled() bool {
	return sys.lifebar.stunbar
	/*
		switch sys.tmode[c.playerNo&1] {
		case TM_Single:
			return sys.cfg.Options.Single.Dizzy
		case TM_Simul:
			return sys.cfg.Options.Simul.Dizzy
		case TM_Tag:
			return sys.cfg.Options.Tag.Dizzy
		case TM_Turns:
			return sys.cfg.Options.Turns.Dizzy
		default:
			return false
		}
	*/
}

func (c *Char) guardBreakEnabled() bool {
	return sys.lifebar.guardbar
	/*
		switch sys.tmode[c.playerNo&1] {
		case TM_Single:
			return sys.cfg.Options.Single.GuardBreak
		case TM_Simul:
			return sys.cfg.Options.Simul.GuardBreak
		case TM_Tag:
			return sys.cfg.Options.Tag.GuardBreak
		case TM_Turns:
			return sys.cfg.Options.Turns.GuardBreak
		default:
			return false
		}
	*/
}

func (c *Char) redLifeEnabled() bool {
	return sys.lifebar.redlifebar
	/*
			switch sys.tmode[c.playerNo&1] {
			case TM_Single:
				return sys.cfg.Options.Single.RedLife
			case TM_Simul:
				return sys.cfg.Options.Simul.RedLife
			case TM_Tag:
				return sys.cfg.Options.Tag.RedLife
			case TM_Turns:
				return sys.cfg.Options.Turns.RedLife
			default:
				return false
		    }
	*/
}

func (c *Char) distX(opp *Char, oc *Char) float32 {
	cpos := c.pos[0] * c.localscl
	opos := opp.pos[0] * opp.localscl
	// Update distance while bound. Mugen chars only
	if c.stWgi().ikemenver[0] == 0 && c.stWgi().ikemenver[1] == 0 {
		if c.bindToId > 0 && !math.IsNaN(float64(c.bindPos[0])) {
			if bt := sys.playerID(c.bindToId); bt != nil {
				f := bt.facing
				// We only need to correct for target binds (and snaps)
				if AbsF(c.bindFacing) == 2 {
					f = c.bindFacing / 2
				}
				cpos = bt.pos[0]*bt.localscl + f*(c.bindPos[0]+c.bindPosAdd[0])*c.localscl
			}
		}
	}
	dist := (opos - cpos) / oc.localscl
	if AbsF(dist) < 0.0001 {
		dist = 0
	}
	return dist
}

func (c *Char) distY(opp *Char, oc *Char) float32 {
	cpos := c.pos[1] * c.localscl
	opos := opp.pos[1] * opp.localscl
	// Update distance while bound. Mugen chars only
	if c.stWgi().ikemenver[0] == 0 && c.stWgi().ikemenver[1] == 0 {
		if c.bindToId > 0 && !math.IsNaN(float64(c.bindPos[0])) {
			if bt := sys.playerID(c.bindToId); bt != nil {
				cpos = bt.pos[1]*bt.localscl + (c.bindPos[1]+c.bindPosAdd[1])*c.localscl
			}
		}
	}
	return (opos - cpos) / oc.localscl
}

func (c *Char) distZ(opp *Char, oc *Char) float32 {
	cpos := c.pos[2] * c.localscl
	opos := opp.pos[2] * opp.localscl
	return (opos - cpos) / oc.localscl
}

// In Mugen, P2BodyDist X does not account for changes in Width like Ikemen does here
func (c *Char) bodyDistX(opp *Char, oc *Char) float32 {
	var cw, oppw float32
	dist := c.distX(opp, oc)

	// Char reference
	// The player reference is always the front width but the enemy reference varies
	// https://github.com/ikemen-engine/Ikemen-GO/issues/2432
	cw = c.sizeBox[2] * c.facing * (c.localscl / oc.localscl)

	// Enemy reference
	if ((dist * c.facing) >= 0) == (c.facing != opp.facing) {
		// Use front width
		oppw = opp.sizeBox[2] * opp.facing * (opp.localscl / oc.localscl)
	} else {
		// Use back width
		oppw = opp.sizeBox[0] * opp.facing * (opp.localscl / oc.localscl)
	}

	return dist - cw + oppw
}

func (c *Char) bodyDistY(opp *Char, oc *Char) float32 {
	ctop := (c.pos[1] + c.sizeBox[1]) * c.localscl
	cbot := (c.pos[1] + c.sizeBox[3]) * c.localscl
	otop := (opp.pos[1] + opp.sizeBox[1]) * opp.localscl
	obot := (opp.pos[1] + opp.sizeBox[3]) * opp.localscl
	if cbot < otop {
		return (otop - cbot) / oc.localscl
	} else if ctop > obot {
		return (obot - ctop) / oc.localscl
	} else {
		return 0
	}
}

func (c *Char) bodyDistZ(opp *Char, oc *Char) float32 {
	ctop := (c.pos[2] - c.sizeDepth[0]) * c.localscl
	cbot := (c.pos[2] + c.sizeDepth[1]) * c.localscl
	otop := (opp.pos[2] - opp.sizeDepth[0]) * opp.localscl
	obot := (opp.pos[2] + opp.sizeDepth[1]) * opp.localscl
	if cbot < otop {
		return (otop - cbot) / oc.localscl
	} else if ctop > obot {
		return (obot - ctop) / oc.localscl
	} else {
		return 0
	}
}

func (c *Char) rdDistX(rd *Char, oc *Char) BytecodeValue {
	if rd == nil {
		return BytecodeSF()
	}
	dist := c.facing * c.distX(rd, oc)
	if c.stWgi().ikemenver[0] == 0 && c.stWgi().ikemenver[1] == 0 {
		if c.stWgi().mugenver[0] != 1 {
			// Before Mugen 1.0, rounding down to the nearest whole number was performed.
			dist = float32(int32(dist))
		}
	}
	return BytecodeFloat(dist)
}

func (c *Char) rdDistY(rd *Char, oc *Char) BytecodeValue {
	if rd == nil {
		return BytecodeSF()
	}
	dist := c.distY(rd, oc)
	if c.stWgi().ikemenver[0] == 0 && c.stWgi().ikemenver[1] == 0 {
		if c.stWgi().mugenver[0] != 1 {
			// Before Mugen 1.0, rounding down to the nearest whole number was performed.
			dist = float32(int32(dist))
		}
	}
	return BytecodeFloat(dist)
}

func (c *Char) rdDistZ(rd *Char, oc *Char) BytecodeValue {
	if rd == nil {
		return BytecodeSF()
	}
	dist := c.distZ(rd, oc)
	return BytecodeFloat(dist)
}

func (c *Char) p2BodyDistX(oc *Char) BytecodeValue {
	if p2 := c.p2(); p2 == nil {
		return BytecodeSF()
	} else {
		dist := c.facing * c.bodyDistX(p2, oc)
		if c.stWgi().mugenver[0] != 1 {
			dist = float32(int32(dist)) // In the old version, decimal truncation was used
		}
		return BytecodeFloat(dist)
	}
}

func (c *Char) p2BodyDistY(oc *Char) BytecodeValue {
	if p2 := c.p2(); p2 == nil {
		return BytecodeSF()
	} else if oc.stWgi().ikemenver[0] == 0 && oc.stWgi().ikemenver[1] == 0 {
		return c.rdDistY(c.p2(), oc) // In Mugen, P2BodyDist Y simply does the same as P2Dist Y
	} else {
		return BytecodeFloat(c.bodyDistY(p2, oc))
	}
}

func (c *Char) p2BodyDistZ(oc *Char) BytecodeValue {
	if p2 := c.p2(); p2 == nil {
		return BytecodeSF()
	} else {
		return BytecodeFloat(c.bodyDistZ(p2, oc))
	}
}

func (c *Char) setPauseTime(pausetime, movetime int32) {
	// Buffer a new Pause only if its timer is higher than the current one or the same player is overriding their own pause
	// This method is more complex but also fairer than Mugen, where only the last pause triggered matters
	if ^pausetime < sys.pausetimebuffer || sys.pauseplayerno == c.playerNo || c.playerNo != c.ss.sb.playerNo {
		sys.pausetimebuffer = ^pausetime
		sys.pauseplayerno = c.playerNo
		if sys.pauseendcmdbuftime < 0 || sys.pauseendcmdbuftime > pausetime {
			sys.pauseendcmdbuftime = 0
		}
	}
	c.pauseMovetime = Max(0, movetime)
	if c.pauseMovetime > pausetime {
		c.pauseMovetime = 0
	} else if sys.pausetime > 0 && c.pauseMovetime > 0 {
		c.pauseMovetime--
	}
}

func (c *Char) setSuperPauseTime(pausetime, movetime int32, unhittable bool, p2defmul float32) {
	// See setPauseTime
	if ^pausetime < sys.supertimebuffer || sys.superplayerno == c.playerNo || c.playerNo != c.ss.sb.playerNo {
		sys.supertimebuffer = ^pausetime
		sys.superplayerno = c.playerNo
		if sys.superendcmdbuftime < 0 || sys.superendcmdbuftime > pausetime {
			sys.superendcmdbuftime = 0
		}

	}

	c.superMovetime = Max(0, movetime)

	if c.superMovetime > pausetime {
		c.superMovetime = 0
	} else if sys.supertime > 0 && c.superMovetime > 0 {
		c.superMovetime--
	}

	if unhittable {
		c.unhittableTime = pausetime + Btoi(pausetime > 0)
	}

	c.ignoreDarkenTime = pausetime

	// Apply superp2defmul to other teams
	// Having this here makes it stack when partners initiate a double pause. Mugen does the same
	if p2defmul != 1 {
		for i := range sys.chars {
			for j := range sys.chars[i] {
				e := sys.chars[i][j]
				if e != nil && e.teamside != c.teamside {
					e.superDefenseMulBuffer *= p2defmul
				}
			}
		}
	}
}

func (c *Char) getPalfx() *PalFX {
	if c.palfx != nil {
		return c.palfx
	}
	if c.parentIndex >= 0 {
		if p := c.parent(false); p != nil {
			return p.getPalfx()
		}
	}
	c.palfx = newPalFX()
	// Mugen 1.1 behavior if invertblend param is omitted (only if char mugenversion = 1.1)
	if c.stWgi().ikemenver[0] == 0 && c.stWgi().ikemenver[1] == 0 && c.stWgi().mugenver[0] == 1 && c.stWgi().mugenver[1] == 1 && c.palfx != nil {
		c.palfx.PalFXDef.invertblend = -2
	}
	return c.palfx
}

func (c *Char) getPalMap() []int {
	return c.getPalfx().remap
}

func (c *Char) pause() bool {
	return c.acttmp <= -2
}

func (c *Char) hitPause() bool {
	return c.hitPauseTime > 0
}

func (c *Char) angleSet(a float32) {
	c.anglerot[0] = a
}

func (c *Char) XangleSet(xa float32) {
	c.anglerot[1] = xa
}

func (c *Char) YangleSet(ya float32) {
	c.anglerot[2] = ya
}

func (c *Char) inputWait() bool {
	if c.asf(ASF_postroundinput) {
		return false
	}
	// If time over
	if sys.curRoundTime == 0 {
		return true
	}
	// If after round "over.waittime" and the win poses have not started
	if sys.intro <= -sys.lifebar.ro.over_waittime && sys.wintime >= 0 {
		return true
	}
	return false
	// In Mugen, once the win poses start the winners can use inputs again but the losers (including draws) cannot
	// This is not currently reproduced and may not be necessary
}

func (c *Char) makeDust(x, y, z float32, spacing int) {
	if c.asf(ASF_nomakedust) {
		return
	}
	if spacing < 1 {
		sys.appendToConsole(c.warn() + "invalid MakeDust spacing")
		spacing = 1
	}
	if c.dustTime >= spacing {
		c.dustTime = 0
	} else {
		return
	}
	if e, i := c.spawnExplod(); e != nil {
		e.animNo = 120
		e.anim_ffx = "f"
		e.sprpriority = math.MaxInt32
		e.layerno = c.layerNo
		e.ownpal = true
		e.relativePos = [...]float32{x, y, z}
		e.setPos(c)
		c.commitExplod(i)
	}
}

func (c *Char) hitFallDamage() {
	if c.ss.moveType == MT_H {
		c.lifeAdd(-float64(c.ghv.fall_damage), c.ghv.fall_kill, false)
		c.ghv.fall_damage = 0
	}
}

func (c *Char) hitFallVel() {
	if c.ss.moveType == MT_H {
		if !math.IsNaN(float64(c.ghv.fall_xvelocity)) {
			c.vel[0] = c.ghv.fall_xvelocity
		}
		c.vel[1] = c.ghv.fall_yvelocity
		if !math.IsNaN(float64(c.ghv.fall_zvelocity)) {
			c.vel[2] = c.ghv.fall_zvelocity
		}
	}
}

func (c *Char) hitFallSet(f int32, xv, yv, zv float32) {
	if f >= 0 {
		c.ghv.fallflag = f != 0
	}
	if !math.IsNaN(float64(xv)) {
		c.ghv.fall_xvelocity = xv
	}
	if !math.IsNaN(float64(yv)) {
		c.ghv.fall_yvelocity = yv
	}
	if !math.IsNaN(float64(zv)) {
		c.ghv.fall_zvelocity = zv
	}
}

func (c *Char) remapPal(pfx *PalFX, src [2]int32, dst [2]int32) {
	// Clear all remaps
	if src[0] == -1 && dst[0] == -1 {
		pfx.remap = nil
		return
	}

	// Reset specified source
	if dst[0] == -1 {
		dst = src
	}

	// Force remap on all palettes
	if src[0] == -1 {
		c.forceRemapPal(pfx, dst)
		return
	}

	// Invalid inputs
	if src[0] < 0 || src[1] < 0 || dst[0] < 0 || dst[1] < 0 {
		return
	}

	plist := c.gi().palettedata.palList

	// Look up source and destination palettes
	si, ok := plist.PalTable[[...]int16{int16(src[0]), int16(src[1])}]
	if !ok || si < 0 {
		sys.appendToConsole(c.warn() + fmt.Sprintf("has no source palette for RemapPal: %v,%v", src[0], src[1]))
		return
	}
	di, ok := plist.PalTable[[...]int16{int16(dst[0]), int16(dst[1])}]
	if !ok || di < 0 {
		sys.appendToConsole(c.warn() + fmt.Sprintf("has no dest palette for RemapPal: %v,%v", dst[0], dst[1]))
		return
	}

	// Init palette remap if needed
	if pfx.remap == nil {
		pfx.remap = plist.GetPalMap()
	}

	// Perform palette remap
	if plist.SwapPalMap(&pfx.remap) {
		plist.Remap(si, di)

		// Remap palette 1, 1 in SFF v1
		if src[0] == 1 && src[1] == 1 && c.gi().sff.header.Ver0 == 1 {
			if spr := c.gi().sff.GetSprite(0, 0); spr != nil {
				plist.Remap(spr.palidx, di)
			}
			if spr := c.gi().sff.GetSprite(9000, 0); spr != nil {
				plist.Remap(spr.palidx, di)
			}
		}

		plist.SwapPalMap(&pfx.remap)
	}

	c.gi().remappedpal = [2]int32{dst[0], dst[1]}
}

func (c *Char) forceRemapPal(pfx *PalFX, dst [2]int32) {
	// Do not remap. Usually because RemapPal parameter was not used
	if dst[0] < 0 || dst[1] < 0 {
		return
	}

	// Get new palette
	di, ok := c.gi().palettedata.palList.PalTable[[...]int16{int16(dst[0]), int16(dst[1])}]
	if !ok || di < 0 {
		return
	}

	// Clear previous remaps
	pfx.remap = make([]int, len(c.gi().palettedata.palList.paletteMap))

	// Apply the new remap
	for i := range pfx.remap {
		pfx.remap[i] = di
	}
}

func (c *Char) getDrawPal(palIndex int) [2]int32 {
	for key, val := range c.gi().palettedata.palList.PalTable {
		if val == palIndex {
			return [2]int32{int32(key[0]), int32(key[1])}
		}
	}
	return [2]int32{0, 0}
}

func (c *Char) drawPal() [2]int32 {
	palMap := c.getPalMap()
	if len(palMap) == 0 {
		return [2]int32{0, 0}
	}
	return c.getDrawPal(palMap[0])
}

type RemapTable map[int16][2]int16
type RemapPreset map[int16]RemapTable

func (c *Char) remapSprite(src [2]int16, dst [2]int16) {
	if src[0] < 0 || src[1] < 0 || dst[0] < 0 || dst[1] < 0 {
		return
	}
	if _, ok := c.remapSpr[src[0]]; !ok {
		c.remapSpr[src[0]] = make(RemapTable)
	}
	c.remapSpr[src[0]][src[1]] = [...]int16{dst[0], dst[1]}
}

func (c *Char) remapSpritePreset(preset string) {
	if _, ok := c.gi().remapPreset[preset]; !ok {
		return
	}
	var src, dst [2]int16
	for src[0] = range c.gi().remapPreset[preset] {
		for src[1], dst = range c.gi().remapPreset[preset][src[0]] {
			c.remapSprite(src, dst)
		}
	}
}

// MapSet() sets a map to a specific value.
func (c *Char) mapSet(s string, Value float32, scType int32) BytecodeValue {
	if s == "" {
		return BytecodeSF()
	}
	key := strings.ToLower(s)
	switch scType {
	case 0: // MapSet
		c.mapArray[key] = Value
	case 1: // MapAdd
		c.mapArray[key] += Value
	case 2: // ParentMapSet
		if p := c.parent(true); p != nil {
			p.mapArray[key] = Value
		}
	case 3: // ParentMapAdd
		if p := c.parent(true); p != nil {
			p.mapArray[key] += Value
		}
	case 4: // RootMapSet
		if r := c.root(true); r != nil {
			r.mapArray[key] = Value
		}
	case 5: // RootMapAdd
		if r := c.root(true); r != nil {
			r.mapArray[key] += Value
		}
	case 6: // TeamMapSet
		if c.teamside == -1 {
			for i := MaxSimul * 2; i < MaxPlayerNo; i += 1 {
				if len(sys.chars[i]) > 0 {
					sys.chars[i][0].mapArray[key] = Value
				}
			}
		} else {
			for i := c.teamside; i < MaxSimul*2; i += 2 {
				if len(sys.chars[i]) > 0 {
					sys.chars[i][0].mapArray[key] = Value
				}
			}
		}
	case 7: // TeamMapAdd
		if c.teamside == -1 {
			for i := MaxSimul * 2; i < MaxPlayerNo; i += 1 {
				if len(sys.chars[i]) > 0 {
					sys.chars[i][0].mapArray[key] += Value
				}
			}
		} else {
			for i := c.teamside; i < MaxSimul*2; i += 2 {
				if len(sys.chars[i]) > 0 {
					sys.chars[i][0].mapArray[key] += Value
				}
			}
		}
	}
	return BytecodeFloat(Value)
}

func (c *Char) appendLifebarAction(text, s_ffx, a_ffx string, snd, spr [2]int32, anim, time int32, timemul float32, top bool) {
	if c.teamside == -1 {
		return
	}
	if _, ok := sys.lifebar.missing["[action]"]; ok { //"
		return
	}

	// Play sound
	if snd[0] != -1 && snd[1] != -1 {
		if s_ffx != "" && s_ffx != "s" && sys.ffx[s_ffx] != nil && sys.ffx[s_ffx].fsnd != nil {
			s := sys.ffx[s_ffx].fsnd.Get(snd) //Common FX
			if s != nil {
				sys.soundChannels.Play(s, snd[0], snd[1], 100, 0, 0, 0, 0)
			}
		} else {
			sys.lifebar.snd.play(snd, 100, 0, 0, 0, 0)
		}
	}

	// If sound only, stop here
	if anim == -1 && (spr[0] == -1 || spr[1] == -1) && text == "" {
		return
	}

	teammsg := sys.lifebar.ac[c.teamside]

	// If adding a new message while exceeding the maximum number allowed, make the oldest message go away faster
	var count int32
	for _, v := range teammsg.messages {
		if !v.del {
			count++
		}
	}
	if count >= teammsg.max {
		var oldest int
		var oldesttimer int32
		// Reset timer for last messages if "top"
		for i := 0; i < len(teammsg.messages); i++ {
			msg := teammsg.messages[i]
			if !msg.del && msg.resttime > 0 && msg.agetimer > oldesttimer {
				oldest = i
				oldesttimer = msg.agetimer
			}
		}
		if oldest < len(teammsg.messages) && teammsg.messages[oldest] != nil {
			teammsg.messages[oldest].resttime = 0
		}
	}

	// Use index 0 if "top", otherwise find the first free message slot
	index := 0
	if !top {
		for k, v := range teammsg.messages {
			if v.del {
				teammsg.messages = removeLbMsg(teammsg.messages, k)
				break
			}
			index++
		}
	}

	// Get default display time from the lifebar
	if time == -1 {
		time = teammsg.displaytime
	}

	// Prepare contents of new message
	msg := newLbMsg(text, int32(float32(time)*timemul), c.teamside)
	delete(teammsg.is, fmt.Sprintf("team%v.front.anim", c.teamside+1))
	delete(teammsg.is, fmt.Sprintf("team%v.front.spr", c.teamside+1))

	// Read animation
	if anim != -1 {
		teammsg.is[fmt.Sprintf("team%v.front.anim", c.teamside+1)] = fmt.Sprintf("%v", anim)
	}
	// Read sprite
	if spr[0] != -1 && spr[1] != -1 {
		teammsg.is[fmt.Sprintf("team%v.front.spr", c.teamside+1)] = fmt.Sprintf("%v,%v", spr[0], spr[1])
	}
	// Read background
	msg.bg = ReadAnimLayout(fmt.Sprintf("team%v.bg.", c.teamside+1), teammsg.is, sys.lifebar.sff, sys.lifebar.at, 2)
	// Read front
	if a_ffx != "" && a_ffx != "s" { //Common FX
		msg.front = ReadAnimLayout(fmt.Sprintf("team%v.front.", c.teamside+1), teammsg.is, sys.ffx[a_ffx].fsff, sys.ffx[a_ffx].fat, 2)
	} else {
		msg.front = ReadAnimLayout(fmt.Sprintf("team%v.front.", c.teamside+1), teammsg.is, sys.lifebar.sff, sys.lifebar.at, 2)
	}

	// Insert new message
	teammsg.messages = insertLbMsg(teammsg.messages, msg, index)
}

func (c *Char) appendDialogue(s string, reset bool) {
	if reset {
		c.dialogue = nil
	}
	c.dialogue = append(c.dialogue, s)
}

func (c *Char) appendToClipboard(pn, sn int, a ...interface{}) {
	spl := sys.stringPool[pn].List
	if sn >= 0 && sn < len(spl) {
		for i, str := range strings.Split(OldSprintf(spl[sn], a...), "\n") {
			if i == 0 && len(c.clipboardText) > 0 {
				c.clipboardText[len(c.clipboardText)-1] += str
			} else {
				c.clipboardText = append(c.clipboardText, str)
			}
		}
		if len(c.clipboardText) > sys.cfg.Debug.ClipboardRows {
			c.clipboardText = c.clipboardText[len(c.clipboardText)-sys.cfg.Debug.ClipboardRows:]
		}
	}
}

func (c *Char) inGuardState() bool {
	return c.ss.no == 120 || (c.ss.no >= 130 && c.ss.no <= 132) ||
		c.ss.no == 140 || (c.ss.no >= 150 && c.ss.no <= 155)
}

func (c *Char) gravity() {
	c.vel[1] += c.gi().movement.yaccel * ((320 / c.localcoord) / c.localscl)
}

// Updates pos based on multiple factors
func (c *Char) posUpdate() {
	// In WinMugen, the threshold for corner push to happen is 4 pixels from the corner
	// In Mugen 1.0 and 1.1 this threshold is bugged, varying with game resolution
	// In Ikemen, this threshold is obsolete
	c.mhv.cornerpush = 0
	pushmul := float32(0.7)
	if c.cornerVelOff != 0 && sys.supertime == 0 {
		for _, p := range sys.chars {
			if len(p) > 0 && p[0].ss.moveType == MT_H && p[0].ghv.playerId == c.id {
				npos := (p[0].pos[0] + p[0].vel[0]*p[0].facing) * p[0].localscl
				if p[0].trackableByCamera() && p[0].csf(CSF_screenbound) && (npos <= sys.xmin || npos >= sys.xmax) {
					c.mhv.cornerpush = c.cornerVelOff
				}
				// In Mugen cornerpush friction is hardcoded at 0.7
				// In Ikemen the cornerpush friction is defined by the target instead
				if c.stWgi().ikemenver[0] == 0 && c.stWgi().ikemenver[1] == 0 {
					pushmul = 0.7
				} else {
					if p[0].ss.stateType == ST_C || p[0].ss.stateType == ST_L {
						pushmul = p[0].gi().movement.crouch.friction
					} else {
						pushmul = p[0].gi().movement.stand.friction
					}
				}
			}
		}
	}

	// Check if character is bound
	nobind := [3]bool{c.bindTime == 0 || math.IsNaN(float64(c.bindPos[0])),
		c.bindTime == 0 || math.IsNaN(float64(c.bindPos[1])),
		c.bindTime == 0 || math.IsNaN(float64(c.bindPos[2]))}
	for i := range nobind {
		if nobind[i] {
			c.oldPos[i], c.interPos[i] = c.pos[i], c.pos[i]
		}
	}

	// Offset position when character is hit off the ground
	// This used to be in actionPrepare(), which would be ideal, but that caused https://github.com/ikemen-engine/Ikemen-GO/issues/2188
	if c.downHitOffset {
		if nobind[0] {
			c.addX(c.gi().movement.down.gethit.offset[0] * (320 / c.localcoord) / c.localscl * c.facing)
		}
		if nobind[1] {
			c.addY(c.gi().movement.down.gethit.offset[1] * (320 / c.localcoord) / c.localscl)
		}
		c.downHitOffset = false
	}

	// Apply velocity
	if c.csf(CSF_posfreeze) {
		if nobind[0] {
			c.setPosX(c.oldPos[0] + c.mhv.cornerpush) // PosFreeze does not disable cornerpush in Mugen
		}
	} else {
		if nobind[0] {
			c.setPosX(c.oldPos[0] + c.vel[0]*c.facing + c.mhv.cornerpush)
		}
		if nobind[1] {
			c.setPosY(c.oldPos[1] + c.vel[1])
		}
		if nobind[2] {
			c.setPosZ(c.oldPos[2] + c.vel[2])
		}
	}

	originLs := c.localscl * (320 / float32(sys.gameWidth))

	// Apply physics types
	switch c.ss.physics {
	case ST_S:
		c.vel[0] *= c.gi().movement.stand.friction
		if AbsF(c.vel[0]) < 1/originLs { // TODO: These probably shouldn't be hardcoded
			c.vel[0] = 0
		}
		c.vel[2] *= c.gi().movement.stand.friction
		if AbsF(c.vel[2]) < 1/originLs {
			c.vel[2] = 0
		}
	case ST_C:
		c.vel[0] *= c.gi().movement.crouch.friction
		c.vel[2] *= c.gi().movement.crouch.friction
	case ST_A:
		c.gravity()
	}

	// Apply friction to corner push
	if sys.supertime == 0 {
		c.cornerVelOff *= pushmul
		if AbsF(c.cornerVelOff) < 1/originLs {
			c.cornerVelOff = 0
		}
	}

	c.bindPosAdd = [...]float32{0, 0, 0}
}

func (c *Char) addTarget(id int32) {
	if !c.hasTarget(id) {
		c.targets = append(c.targets, id)
	}
}

func (c *Char) hasTarget(id int32) bool {
	for _, tid := range c.targets {
		if tid == id {
			return true
		}
	}
	return false
}

func (c *Char) hasTargetOfHitdef(id int32) bool {
	for _, tid := range c.hitdefTargets {
		if tid == id {
			return true
		}
	}
	return false
}

func (c *Char) targetAddSctrl(id int32) {
	// Check if ID exists
	t := sys.playerID(id)
	if t == nil {
		sys.appendToConsole(c.warn() + fmt.Sprintf("Invalid player ID for TargetAdd: %v", id))
		return
	}

	// Add target to char's "target" list
	// These two functions already prevent duplicating players
	c.addTarget(id)

	// Add original char to target's "targeted by" list
	t.ghv.addId(c.id, c.gi().data.airjuggle)
}

func (c *Char) setBindTime(time int32) {
	c.bindTime = time
	if time == 0 {
		c.bindToId = -1
		c.bindFacing = 0
	}
}

func (c *Char) setBindToId(to *Char, isTargetBind bool) {
	if c.bindToId != to.id {
		c.bindToId = to.id
	}
	// Target binds are all we need to correct with this logic.
	// By the time this gets to the bind() method, it's going to
	// default to setting the facing to the same as the "bindTo"
	// facing at that point. So as weird as it may be to default
	// to 0 here, this behavior does seem to be what MUGEN
	// actually does for helpers.
	if c.bindFacing == 0 && isTargetBind {
		c.bindFacing = to.facing * 2
	}
	if to.bindToId == c.id {
		to.setBindTime(0)
	}
}

func (c *Char) bind() {
	if c.bindTime == 0 {
		if bt := sys.playerID(c.bindToId); bt != nil {
			if bt.hasTarget(c.id) {
				if bt.csf(CSF_destroy) {
					sys.appendToConsole(c.warn() + fmt.Sprintf("SelfState 5050, helper destroyed: %v", bt.name))
					if c.ss.moveType == MT_H {
						c.selfState(5050, -1, -1, -1, "")
					}
					c.setBindTime(0)
					return
				}
			}
		}
		if c.bindToId > 0 {
			c.setBindTime(0)
		}
		return
	}
	if bt := sys.playerID(c.bindToId); bt != nil {
		if bt.hasTarget(c.id) {
			if !math.IsNaN(float64(c.bindPos[0])) {
				c.vel[0] = c.facing * bt.facing * bt.vel[0]
			}
			if !math.IsNaN(float64(c.bindPos[1])) {
				c.vel[1] = bt.vel[1]
			}
			if !math.IsNaN(float64(c.bindPos[2])) {
				c.vel[2] = bt.vel[2]
			}
		}
		if !math.IsNaN(float64(c.bindPos[0])) {
			f := bt.facing
			// We only need to correct for target binds (and snaps)
			if AbsF(c.bindFacing) == 2 {
				f = c.bindFacing / 2
			}
			c.setAllPosX(bt.pos[0]*bt.localscl/c.localscl + f*(c.bindPos[0]+c.bindPosAdd[0]))
			c.interPos[0] += bt.interPos[0] - bt.pos[0]
			c.oldPos[0] += bt.oldPos[0] - bt.pos[0]
			c.pushed = c.pushed || bt.pushed
			c.ghv.xoff = 0
		}
		if !math.IsNaN(float64(c.bindPos[1])) {
			c.setAllPosY(bt.pos[1]*bt.localscl/c.localscl + (c.bindPos[1] + c.bindPosAdd[1]))
			c.interPos[1] += bt.interPos[1] - bt.pos[1]
			c.oldPos[1] += bt.oldPos[1] - bt.pos[1]
			c.ghv.yoff = 0
		}
		if !math.IsNaN(float64(c.bindPos[2])) {
			c.setAllPosZ(bt.pos[2]*bt.localscl/c.localscl + (c.bindPos[2] + c.bindPosAdd[2]))
			c.interPos[2] += bt.interPos[2] - bt.pos[2]
			c.oldPos[2] += bt.oldPos[2] - bt.pos[2]
			c.ghv.zoff = 0
		}
		if AbsF(c.bindFacing) == 1 {
			if c.bindFacing > 0 {
				c.setFacing(bt.facing)
			} else {
				c.setFacing(-bt.facing)
			}
		}
	} else {
		c.setBindTime(0)
		return
	}
}

func (c *Char) trackableByCamera() bool {
	return sys.cam.View == Fighting_View || sys.cam.View == Follow_View && c == sys.cam.FollowChar
}

func (c *Char) xScreenBound() {
	x := c.pos[0]
	if !sys.cam.roundstart && c.trackableByCamera() && c.csf(CSF_screenbound) && !c.scf(SCF_standby) {
		min, max := c.edgeWidth[0], -c.edgeWidth[1]
		if c.facing > 0 {
			min, max = -max, -min
		}
		x = ClampF(x, min+sys.xmin/c.localscl, max+sys.xmax/c.localscl)
	}
	if c.csf(CSF_stagebound) {
		x = ClampF(x, sys.stage.leftbound*sys.stage.localscl/c.localscl, sys.stage.rightbound*sys.stage.localscl/c.localscl)
	}
	c.setAllPosX(x)
}

func (c *Char) zDepthBound() {
	posz := c.pos[2]
	if c.csf(CSF_stagebound) {
		min := c.edgeDepth[0]
		max := -c.edgeDepth[1]
		posz = ClampF(posz, min+sys.zmin/c.localscl, max+sys.zmax/c.localscl)
	}
	c.setAllPosZ(posz)
}

func (c *Char) xPlatformBound(pxmin, pxmax float32) {
	x := c.pos[0]
	if c.ss.stateType != ST_A {
		min, max := c.edgeWidth[0], -c.edgeWidth[1]
		if c.facing > 0 {
			min, max = -max, -min
		}
		x = ClampF(x, min+pxmin/c.localscl, max+pxmax/c.localscl)
	}
	c.setAllPosX(x)
	c.xScreenBound()
}

func (c *Char) gethitBindClear() {
	if c.isTargetBound() {
		c.setBindTime(0)
	}
}

// Drop targets that no longer fit the requirements
func (c *Char) dropTargets() {
	if c.hitdef.reversal_attr == 0 || c.hitdef.reversal_attr == -1<<31 {
		i := 0
		for i < len(c.targets) {
			if i >= len(c.targets) {
				break
			}
			if t := sys.playerID(c.targets[i]); t != nil {
				if t.ss.moveType != MT_H && !t.stchtmp {
					c.targets[i] = c.targets[len(c.targets)-1]
					c.targets = c.targets[:len(c.targets)-1]
					if t.ghv._type != 0 { // https://github.com/ikemen-engine/Ikemen-GO/issues/1268
						t.ghv.hitid = -1
					}
				} else {
					i++
				}
				continue
			}
			i++
		}
	}
}

func (c *Char) removeTarget(pid int32) {
	for i, t := range c.targets {
		if t == pid {
			c.targets = append(c.targets[:i], c.targets[i+1:]...)
			break
		}
	}
}

// Remove self from the target lists of other players
func (c *Char) exitTarget() {
	if c.hittmp >= 0 { // If not being hit by ReversalDef
		for _, hb := range c.ghv.targetedBy {
			if e := sys.playerID(hb[0]); e != nil {
				if e.hitdef.reversal_attr == 0 || e.hitdef.reversal_attr == -1<<31 {
					e.removeTarget(c.id)
				} else {
					c.ghv.hitid = c.ghv.hitid >> 31
				}
			}
		}
		c.gethitBindClear()
		// This line used to be outside the "c.hittmp >= 0" condition, but this happened
		// https://github.com/ikemen-engine/Ikemen-GO/issues/2581
		c.ghv.targetedBy = c.ghv.targetedBy[:0]
	}
}

func (c *Char) offsetX() float32 {
	return float32(c.size.draw.offset[0])*c.facing + c.offset[0]/c.localscl
}

func (c *Char) offsetY() float32 {
	return float32(c.size.draw.offset[1]) + c.offset[1]/c.localscl
}

// Gather the character as well as all its proxy children (and their proxy children) in a flat slice
func (c *Char) flattenClsnProxies() []*Char {
	var list []*Char

	// Start with the base character
	queue := []*Char{c}

	// Process the queue until all characters (base + proxies) have been handled
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]

		list = append(list, current)

		for _, child := range current.children {
			if child != nil && child.isclsnproxy {
				queue = append(queue, child)
			}
		}
	}

	return list
}

func (c *Char) projClsnCheck(p *Projectile, cbox, pbox int32) bool {
	// Safety checks
	if p.ani == nil || c.curFrame == nil || c.scf(SCF_standby) || c.scf(SCF_disabled) {
		return false
	}

	// Clsnproxies do not hit nor get hit themselves. They act as extensions of their parent's clsn boxes.
	if c.isclsnproxy {
		return false
	}

	// Get char and its proxies
	var charTotal []*Char
	if cbox == 3 {
		charTotal = []*Char{c} // Except if size box
	} else {
		charTotal = c.flattenClsnProxies()
	}

	// Loop through all characters and check collision
	for _, charSingle := range charTotal {
		if charSingle.projClsnCheckSingle(p, cbox, pbox) {
			return true
		}
	}

	return false
}

func (c *Char) projClsnCheckSingle(p *Projectile, cbox, pbox int32) bool {
	// Safety checks
	if p.ani == nil || c.curFrame == nil || c.scf(SCF_standby) || c.scf(SCF_disabled) {
		return false
	}

	// Get projectile animation frame
	frm := p.ani.CurrentFrame()

	// Check if animation frames are valid
	if frm == nil || c.curFrame == nil {
		return false
	}

	// Accepted box types
	if cbox != 1 && cbox != 2 && cbox != 3 {
		return false
	}

	// Required boxes not found
	if p.hitdef.p2clsnrequire == 1 && c.curFrame.Clsn1 == nil ||
		p.hitdef.p2clsnrequire == 2 && c.curFrame.Clsn2 == nil {
		return false
	}

	// Decide which box types should collide
	var clsn1, clsn2 [][4]float32
	if c.asf(ASF_projtypecollision) { // Projectiles trade with their Clsn2 only
		clsn1 = frm.Clsn2
		clsn2 = c.curFrame.Clsn2
	} else {
		if pbox == 2 {
			clsn1 = frm.Clsn2
		} else {
			clsn1 = frm.Clsn1
		}
		if cbox == 1 {
			clsn2 = c.curFrame.Clsn1
			if clsn2 == nil && p.hitdef.p2clsnrequire == 1 {
				return false
			}
		} else if cbox == 3 {
			clsn2 = c.sizeBoxToClsn()
			if clsn2 == nil && p.hitdef.p2clsnrequire == 3 {
				return false // Size box should alway exist, but...
			}
		} else {
			clsn2 = c.curFrame.Clsn2
			if clsn2 == nil && p.hitdef.p2clsnrequire == 2 {
				return false
			}
		}
	}

	if clsn1 == nil || clsn2 == nil {
		return false
	}

	// Exceptions for size boxes as they don't rescale or rotate
	charscale := c.clsnScale
	charangle := c.clsnAngle
	if cbox == 3 {
		charscale = [2]float32{c.localscl, c.localscl}
		charangle = 0
	}

	return sys.clsnOverlap(
		clsn1,
		[...]float32{p.clsnScale[0] * p.localscl * p.zScale, p.clsnScale[1] * p.localscl * p.zScale},
		[...]float32{p.pos[0] * p.localscl, p.pos[1] * p.localscl},
		p.facing,
		p.clsnAngle,
		clsn2,
		charscale,
		[...]float32{c.pos[0]*c.localscl + c.offsetX()*c.localscl,
			c.pos[1]*c.localscl + c.offsetY()*c.localscl},
		c.facing,
		charangle,
	)
}
func (c *Char) projClsnOverlapTrigger(index, targetID, boxType int32) bool {
	projs := c.getProjs(-1)

	if index < 0 || int(index) >= len(projs) {
		return false
	}
	proj := projs[index]

	target := sys.playerID(targetID)
	if target == nil {
		return false
	}

	return target.projClsnCheck(proj, boxType, 1) || target.projClsnCheck(proj, boxType, 2)
}
func (c *Char) clsnCheck(getter *Char, charbox, getterbox int32, reqcheck, trigger bool) bool {
	// Safety checks
	if c == nil || getter == nil || c.anim == nil || getter.anim == nil {
		return false
	}

	// Clsnproxies do not hit nor get hit themselves. They act as extensions of their parent's clsn boxes.
	if c.isclsnproxy || getter.isclsnproxy {
		return false
	}

	// Determine which characters to check
	var charTotal []*Char
	if charbox == 3 {
		// Only base character for size box
		charTotal = []*Char{c}
	} else {
		// Otherwise include all proxies
		charTotal = c.flattenClsnProxies()
	}

	var getterTotal []*Char
	if getterbox == 3 {
		getterTotal = []*Char{getter}
	} else {
		getterTotal = getter.flattenClsnProxies()
	}

	// Check collision for all combinations
	for _, charSingle := range charTotal {
		for _, getterSingle := range getterTotal {
			if charSingle.clsnCheckSingle(getterSingle, charbox, getterbox, reqcheck, trigger) {
				return true
			}
		}
	}

	return false
}

func (c *Char) clsnCheckSingle(getter *Char, charbox, getterbox int32, reqcheck, trigger bool) bool {
	// Safety checks
	if c == nil || getter == nil || c.anim == nil || getter.anim == nil {
		return false
	}

	// What this does is normally check the Clsn in the currently displayed frame
	// But in the ClsnOverlap trigger, we must check the frame that *will* be displayed instead
	charframe := c.curFrame
	getterframe := getter.curFrame
	if trigger {
		charframe = c.anim.CurrentFrame()
		getterframe = getter.anim.CurrentFrame()
	}

	// Nil anim & standby check
	if charframe == nil || getterframe == nil ||
		c.scf(SCF_standby) || getter.scf(SCF_standby) ||
		c.scf(SCF_disabled) || getter.scf(SCF_disabled) {
		return false
	}

	// Accepted box types
	if charbox != 1 && charbox != 2 && charbox != 3 {
		return false
	}
	if getterbox != 1 && getterbox != 2 && getterbox != 3 {
		return false
	}

	// Required boxes not found
	// Only Hitdef and Reversaldef do this check
	if reqcheck {
		if c.hitdef.p2clsnrequire == 1 && getterframe.Clsn1 == nil ||
			c.hitdef.p2clsnrequire == 2 && getterframe.Clsn2 == nil {
			return false
		}
	}

	// Decide which box types should collide
	var clsn1, clsn2 [][4]float32
	if c.asf(ASF_projtypecollision) && getter.asf(ASF_projtypecollision) { // Projectiles trade with their Clsn2 only
		clsn1 = charframe.Clsn2
		clsn2 = getterframe.Clsn2
	} else {
		if charbox == 1 {
			clsn1 = charframe.Clsn1
		} else if charbox == 3 {
			clsn1 = c.sizeBoxToClsn()
		} else {
			clsn1 = charframe.Clsn2
		}
		if getterbox == 1 {
			clsn2 = getterframe.Clsn1
		} else if getterbox == 3 {
			clsn2 = [][4]float32{getter.sizeBox}
		} else {
			clsn2 = getterframe.Clsn2
		}
	}

	if clsn1 == nil || clsn2 == nil {
		return false
	}

	// Exceptions for size boxes as they don't rescale or rotate
	charscale := c.clsnScale
	charangle := c.clsnAngle
	if charbox == 3 {
		charscale = [2]float32{c.localscl, c.localscl}
		charangle = 0
	}

	getterscale := getter.clsnScale
	getterangle := getter.clsnAngle
	if getterbox == 3 {
		getterscale = [2]float32{getter.localscl, getter.localscl}
		getterangle = 0
	}

	return sys.clsnOverlap(
		clsn1,
		charscale,
		[...]float32{c.pos[0]*c.localscl + c.offsetX()*c.localscl,
			c.pos[1]*c.localscl + c.offsetY()*c.localscl},
		c.facing,
		charangle,
		clsn2, // Getter
		getterscale,
		[...]float32{getter.pos[0]*getter.localscl + getter.offsetX()*getter.localscl,
			getter.pos[1]*getter.localscl + getter.offsetY()*getter.localscl},
		getter.facing,
		getterangle,
	)
}

func (c *Char) hitByAttrTrigger(attr int32) bool {
	// Unhittable timer invalidates all hits
	if c.unhittableTime > 0 {
		return false
	}

	// Get state type (SCA) from among the attributes
	attrsca := attr & int32(ST_MASK)

	// Compare given attributes to character's HitBy slots
	return c.checkHitByAllSlots(-1, -1, attr, attrsca)
}

// Check vulnerability in a single HitBy slot
func (c *Char) checkHitBySlot(hb HitBy, getterno int, getterid, ghdattr, attrsca int32) bool {
	// Attribute
	// Note: State type and attack attributes must be checked individually
	attrCheck := hb.flag >= 0
	scaMatch := hb.flag&attrsca != 0
	atkMatch := hb.flag&ghdattr&^int32(ST_MASK) != 0

	// Player number
	pnoCheck := hb.playerno >= 0
	pnoMatch := hb.playerno == getterno

	// Player ID
	pidCheck := hb.playerid >= 0
	pidMatch := hb.playerid == getterid

	// For NotHitBy the hit is allowed only if no defined parameter matches
	if hb.not {
		anyMatch := (attrCheck && (scaMatch || atkMatch)) || (pnoCheck && pnoMatch) || (pidCheck && pidMatch)
		return !anyMatch
	}

	// For HitBy the hit is allowed only if all defined parameters match
	allMatch := (!attrCheck || (scaMatch && atkMatch)) && (!pnoCheck || pnoMatch) && (!pidCheck || pidMatch)
	return allMatch
}

// checkHitByAllSlots evaluates all of the character's HitBy/NotHitBy slots
// to determine if the character is vulnerable to the current attack.
func (c *Char) checkHitByAllSlots(getterno int, getterid, ghdattr, attrsca int32) bool {
	stackHit := false
	hasStackSlot := false
	nonStackHit := true

	for _, hb := range c.hitby {
		// Skip inactive slots
		if hb.time == 0 {
			continue
		}

		if hb.stack {
			// OR logic: If vulnerable in any of the stack slots, the attack will hit
			hasStackSlot = true
			if c.checkHitBySlot(hb, getterno, getterid, ghdattr, attrsca) {
				stackHit = true
			}
		} else {
			// AND logic: If there is even one slot without vulnerability, the attack will miss
			if !c.checkHitBySlot(hb, getterno, getterid, ghdattr, attrsca) {
				nonStackHit = false
			}
		}
	}

	// Combine OR (stack) and AND (non-stack)
	if hasStackSlot {
		return stackHit && nonStackHit
	}

	// Was vulnerable in all non-stack slots
	return nonStackHit
}

// Check if HitDef attributes can hit a player
func (c *Char) attrCheck(getter *Char, ghd *HitDef, gstyp StateType) bool {

	// Invalid attributes
	if ghd.attr <= 0 && ghd.reversal_attr <= 0 {
		return false
	}

	// Unhittable and ChainID checks
	if c.unhittableTime > 0 || ghd.chainid >= 0 && c.ghv.hitid != ghd.chainid && ghd.nochainid[0] == -1 {
		return false
	}
	if (len(c.ghv.targetedBy) > 0 && c.ghv.targetedBy[len(c.ghv.targetedBy)-1][0] == getter.id) || c.ghv.hitshaketime > 0 { // https://github.com/ikemen-engine/Ikemen-GO/issues/320
		for _, nci := range ghd.nochainid {
			if nci >= 0 && c.ghv.hitid == nci && c.ghv.playerId == ghd.attackerID {
				return false
			}
		}
	}

	// https://github.com/ikemen-engine/Ikemen-GO/issues/308
	//if ghd.chainid < 0 {

	// ReversalDef vs HitDef attributes check
	if ghd.reversal_attr > 0 {
		// Check HitDef validity
		if c.hitdef.attr <= 0 || c.atktmp == 0 {
			return false
		}

		// Check attributes
		if (c.hitdef.attr&ghd.reversal_attr&int32(ST_MASK)) == 0 ||
			(c.hitdef.attr&ghd.reversal_attr&^int32(ST_MASK)) == 0 {
			return false
		}

		// Check guardflag
		if ghd.reversal_guardflag != IErr &&
			(ghd.reversal_guardflag&c.hitdef.guardflag == 0 || c.asf(ASF_unguardable)) {
			return false
		}

		// Check guardflag.not
		if ghd.reversal_guardflag_not != IErr &&
			(ghd.reversal_guardflag_not&c.hitdef.guardflag != 0 && !c.asf(ASF_unguardable)) {
			return false
		}

		return true
	}

	// Main hitflag checks
	if ghd.hitflag&int32(HF_H) == 0 && c.ss.stateType == ST_S ||
		ghd.hitflag&int32(HF_L) == 0 && c.ss.stateType == ST_C ||
		ghd.hitflag&int32(HF_A) == 0 && c.ss.stateType == ST_A ||
		ghd.hitflag&int32(HF_D) == 0 && c.ss.stateType == ST_L {
		return false
	}

	// "F" hitflag check
	if (ghd.hitflag&int32(HF_F) == 0 || getter.asf(ASF_nofallhitflag)) && c.hittmp >= 2 {
		return false
	}

	// "-" hitflag check
	if ghd.hitflag&int32(HF_MNS) != 0 && c.hittmp > 0 {
		return false
	}

	// "+" hitflag check
	if ghd.hitflag&int32(HF_PLS) != 0 && (c.hittmp <= 0 || c.inGuardState()) {
		return false
	}

	// Get state type (SCA) from among the Hitdef attributes
	attrsca := ghd.attr & int32(ST_MASK)
	// Note: In Mugen, invincibility is checked against the enemy's actual statetype instead of the Hitdef's SCA attribute
	// Exception for projectiles, where it respects the SCA attribute
	// Ikemen characters work as documented. Invincibility only cares about the HitDef's SCA attribute
	if getter.stWgi().ikemenver[0] == 0 && getter.stWgi().ikemenver[1] == 0 {
		if gstyp == ST_N { // Projectiles mostly
			attrsca = ghd.attr & int32(ST_MASK)
		} else {
			attrsca = int32(gstyp)
		}
	}

	// HitBy and NotHitBy checks
	if !c.checkHitByAllSlots(getter.playerNo, getter.id, ghd.attr, attrsca) {
		return false
	}

	return true
}

// Check if the enemy's (c) HitDef should lose to the player's (getter), if applicable
func (c *Char) hittableByChar(getter *Char, ghd *HitDef, gst StateType, proj bool) bool {

	// Enemy (c) always wins if they can't be hit by the player's (getter) HitDef attributes at all
	if !c.attrCheck(getter, ghd, gst) {
		return false
	}

	// Enemy (c) always loses if their HitDef already hit the player (getter)
	if c.hasTargetOfHitdef(getter.id) {
		return true
	}

	// Enemy (c) always loses if they have an invalid HitDef or ReversalDef
	if c.atktmp == 0 || (c.hitdef.attr <= 0 || c.ss.stateType == ST_L) && c.hitdef.reversal_attr <= 0 {
		return true
	}

	// Check if the enemy (c) can also hit the player (getter)
	// Used to check for instance if a lower priority exchanges with a higher priority but the higher priority Clsn1 misses
	// This could probably be a function that both players access instead of being handled like this
	countercheck := func(hd *HitDef) bool {
		if proj {
			return false
		} else {
			return (getter.atktmp >= 0 || !c.hasTarget(getter.id)) &&
				!getter.hasTargetOfHitdef(c.id) &&
				getter.attrCheck(c, hd, c.ss.stateType) &&
				c.clsnCheck(getter, 1, c.hitdef.p2clsncheck, true, false) &&
				sys.zAxisOverlap(c.pos[2], c.hitdef.attack_depth[0], c.hitdef.attack_depth[1], c.localscl,
					getter.pos[2], getter.sizeDepth[0], getter.sizeDepth[1], getter.localscl)
		}
	}

	// Enemy (c) ReversalDef check
	if c.hitdef.reversal_attr > 0 {
		if ghd.reversal_attr > 0 { // ReversalDef vs ReversalDef
			if countercheck(&c.hitdef) {
				c.atktmp = -1
				return getter.atktmp < 0
			}
			return true
		}
		return !countercheck(&c.hitdef)
	}

	// Enemy (c) loses if player (getter) has ReversalDef
	if ghd.reversal_attr > 0 {
		return true
	}

	// Enemy (c) loses if their HitDef has lower priority
	if c.hitdef.priority < ghd.priority {
		return true
	}

	// Enemy (c) HitDef has higher priority. Run counter check
	if c.hitdef.priority > ghd.priority {
		return !countercheck(&c.hitdef)
	}

	// Both HitDefs have same priority. Check trade types
	if ghd.priority == c.hitdef.priority {
		switch {
		case c.hitdef.prioritytype == TT_Dodge:
			return !countercheck(&c.hitdef)
		case ghd.prioritytype == TT_Dodge:
			return !countercheck(&c.hitdef)
		case ghd.prioritytype == TT_Miss:
			return !countercheck(&c.hitdef)
		case c.hitdef.prioritytype == TT_Hit:
			// if (c.hitdef.p1stateno >= 0 || c.hitdef.attr&int32(AT_AT) != 0 && ghd.hitonce != 0) && countercheck(&c.hitdef) {
			// Since the unhittabletime is what's behind needing to randomize throws, we will check it instead
			if (c.hitdef.unhittabletime[0] > 0 && ghd.hitonce != 0) && countercheck(&c.hitdef) {
				c.atktmp = -1
				return getter.atktmp < 0 || Rand(0, 1) == 1
			}
			return true
		default:
			return true
		}
	}

	// Other cases
	return true
}

func (c *Char) hitResultCheck(getter *Char, proj *Projectile) (hitResult int32) {

	// Player hit check
	hd := &c.hitdef
	attackMul := &c.attackMul
	posDiff := [3]float32{0, 0, 0}
	isProjectile := false

	// Projectile hit check
	if proj != nil {
		hd = &proj.hitdef
		attackMul = &proj.parentAttackMul
		isProjectile = true

		posDiff = [3]float32{
			proj.pos[0] - c.pos[0]*(c.localscl/proj.localscl),
			proj.pos[1] - c.pos[1]*(c.localscl/proj.localscl),
			proj.pos[2] - c.pos[2]*(c.localscl/proj.localscl),
		}
	}

	// If attacking while statetype L
	if isProjectile && c.ss.stateType == ST_L && hd.reversal_attr <= 0 {
		c.hitdef.ltypehit = true
		return 0
	}

	// If using p2stateno but the enemy is already changing states
	if getter.stchtmp && getter.ss.sb.playerNo != hd.playerNo {
		if getter.csf(CSF_gethit) {
			if hd.p2stateno >= 0 {
				return 0
			}
		} else if getter.acttmp > 0 {
			return 0
		}
	}

	// If using p1stateno but the char was already hit or is already changing states
	if hd.p1stateno >= 0 && (c.csf(CSF_gethit) || c.stchtmp && c.ss.sb.playerNo != hd.playerNo) {
		return 0
	}

	// If getter already hit by a throw
	if getter.csf(CSF_gethit) && getter.ghv.attr&int32(AT_AT) != 0 {
		return 0
	}

	// Check if the enemy can guard this attack
	// Unguardable flag also affects projectiles
	// https://github.com/ikemen-engine/Ikemen-GO/issues/2367
	canguard := !c.asf(ASF_unguardable) && getter.scf(SCF_guard) &&
		(!getter.csf(CSF_gethit) || getter.ghv.guarded)

	// Automatically choose high or low in case of auto guard
	if canguard && getter.asf(ASF_autoguard) && getter.acttmp > 0 && !getter.csf(CSF_gethit) {
		highflag := hd.guardflag&int32(HF_H) != 0
		lowflag := hd.guardflag&int32(HF_L) != 0
		if highflag != lowflag {
			if lowflag && getter.ss.stateType == ST_S { // High to low
				getter.ss.changeStateType(ST_C)
			} else if highflag && getter.ss.stateType == ST_C { // Low to high
				getter.ss.changeStateType(ST_S)
			}
		}
	}

	// Default hit type and kill flag to "hit" (1)
	hitResult = 1
	getter.ghv.kill = hd.kill
	// If enemy is guarding the correct way, "hitResult" is set to "guard" (2)
	if canguard {
		// Guardflag checks
		if hd.guardflag&int32(HF_H) != 0 && getter.ss.stateType == ST_S ||
			hd.guardflag&int32(HF_L) != 0 && getter.ss.stateType == ST_C ||
			hd.guardflag&int32(HF_A) != 0 && getter.ss.stateType == ST_A { // Statetype L is left out here
			// Switch kill flag to guard if attempting to guard correctly
			getter.ghv.kill = hd.guard_kill
			// We only switch to guard behavior if the enemy can survive guarding the attack
			if getter.life > getter.computeDamage(float64(hd.guarddamage), hd.guard_kill, false, attackMul[0]*(float32(c.gi().attackBase)/100), c, true) ||
				sys.gsf(GSF_globalnoko) || getter.asf(ASF_noko) || getter.asf(ASF_noguardko) {
				hitResult = 2
			} else {
				getter.ghv.cheeseKO = true // TODO: find a better name then expose this variable
			}
		}
	}

	// If using ReversalDef or hitting with type None, the hitResult is negative
	if hitResult > 0 {
		if hd.reversal_attr > 0 {
			hitResult *= -1
		} else if getter.ss.stateType == ST_A {
			if hd.air_type == HT_None {
				hitResult *= -1
			}
		} else if hd.ground_type == HT_None {
			hitResult *= -1
		}
	}

	// If any previous hit in the current frame will KO the enemy, the following ones will not prevent it
	if getter.ghv.damage >= getter.life {
		getter.ghv.kill = true
	}

	// Check if P2 should change state
	p2s := false
	if !getter.stchtmp || !getter.csf(CSF_gethit) {
		// Check HitOverride
		c.mhv.overridden = false
		for i := 0; i < len(getter.hover); i++ {
			ho := &getter.hover[i]
			// Check timer and attack attributes
			if ho.time == 0 || ho.attr&hd.attr&^int32(ST_MASK) == 0 {
				continue
			}
			// Check SCA attributes
			if isProjectile {
				if ho.attr&hd.attr&int32(ST_MASK) == 0 {
					continue
				}
			} else {
				if ho.attr&int32(c.ss.stateType) == 0 {
					continue
				}
			}
			// Check guardflag
			if ho.guardflag != IErr && (ho.guardflag&hd.guardflag == 0 || c.asf(ASF_unguardable)) {
				continue
			}
			// Check guardflag.not
			if ho.guardflag_not != IErr && (ho.guardflag_not&hd.guardflag != 0 && !c.asf(ASF_unguardable)) {
				continue
			}
			// Miss if using p1stateno or p2stateno and HitOverride together
			// In Mugen, it misses even if the enemy guards // && Abs(hitResult) == 1
			if hd.missonoverride == 1 ||
				(hd.missonoverride == -1 && !isProjectile && (hd.p1stateno >= 0 || hd.p2stateno >= 0)) {
				return 0
			}
			// Set flags
			c.mhv.overridden = true
			if ho.keepState {
				getter.hoverKeepState = true
			}
			// Select this HitOverride slot
			getter.hoverIdx = i
			break
		}

		// Apply HitOverride properties
		if c.mhv.overridden && getter.hoverIdx >= 0 {
			ho := &getter.hover[getter.hoverIdx]
			// Forceair behavior
			if ho.forceair {
				if hitResult > 0 && hd.air_type == HT_None || hitResult < 0 && hd.ground_type == HT_None && hd.air_type != HT_None {
					hitResult *= -1
				}
				if Abs(hitResult) == 1 {
					getter.ss.changeStateType(ST_A)
				}
			}
			// Force behavior to hit, unless ForceGuard is used
			if hitResult > 0 {
				if ho.forceguard {
					hitResult = 2
				} else {
					hitResult = 1
				}
			}
		}

		// Apply P2StateNo
		// In Mugen, an undefined HitOverride stateno still invalidates P2StateNo
		if !c.mhv.overridden {
			if Abs(hitResult) == 1 && hd.p2stateno >= 0 {
				pn := getter.playerNo
				if hd.p2getp1state {
					pn = hd.playerNo
				}
				if getter.stateChange1(hd.p2stateno, pn) {
					// In Mugen, using p2stateno forces movetype to H
					// https://github.com/ikemen-engine/Ikemen-GO/issues/2466
					getter.ss.changeMoveType(MT_H)
					getter.setCtrl(false)
					p2s = true
					getter.hoverIdx = -1
				}
			}
		}
	}

	if !isProjectile {
		c.hitdefTargetsBuffer = append(c.hitdefTargetsBuffer, getter.id)
	}

	// Determine if GetHitVars should be updated
	ghvset := !getter.csf(CSF_gethit) || !getter.stchtmp || p2s

	// Variables that are set by default even if Hitdef type is "None"
	if ghvset {
		getter.ghv.hitid = hd.id
		getter.ghv.playerNo = hd.playerNo
		getter.ghv.playerId = hd.attackerID
		getter.ghv.groundtype = hd.ground_type
		getter.ghv.airtype = hd.air_type
		if getter.ss.stateType == ST_A {
			getter.ghv._type = getter.ghv.airtype
		} else {
			getter.ghv._type = getter.ghv.groundtype
		}
		if !isProjectile {
			c.sprPriority = hd.p1sprpriority
		}
		getter.sprPriority = hd.p2sprpriority
	}

	// Attacker facing
	byf := c.facing
	if isProjectile {
		byf = proj.facing
	}
	if !isProjectile && hitResult == 1 {
		if hd.p1getp2facing != 0 {
			byf = getter.facing
			if hd.p1getp2facing < 0 {
				byf *= -1
			}
		} else if hd.p1facing < 0 {
			byf *= -1
		}
	}

	// Getter is hit or guards the Hitdef
	if hitResult > 0 {
		// Stop enemy's flagged sounds. In Mugen this only happens with channel 0
		if hitResult == 1 {
			for i := range getter.soundChannels.channels {
				if getter.soundChannels.channels[i].stopOnGetHit {
					getter.soundChannels.channels[i].Stop()
					getter.soundChannels.channels[i].stopOnGetHit = false
				}
			}
		}
		if getter.bindToId == c.id {
			getter.setBindTime(0)
		}
		if ghvset {
			ghv := &getter.ghv
			cmb := (getter.ss.moveType == MT_H || getter.csf(CSF_gethit)) && !ghv.guarded
			// Precompute localcoord conversion factor
			scaleratio := c.localscl / getter.localscl

			// Clear GetHitVars while stacking those that need it
			// Skipping this step makes the test case in #1891 work, but for different reasons than in Mugen
			ghv.selectiveClear(getter)

			ghv.attr = hd.attr
			ghv.guardflag = hd.guardflag
			ghv.hitid = hd.id
			ghv.playerNo = hd.playerNo
			ghv.playerId = hd.attackerID
			ghv.xaccel = hd.xaccel * scaleratio * -byf
			ghv.yaccel = hd.yaccel * scaleratio
			ghv.zaccel = hd.zaccel * scaleratio
			ghv.groundtype = hd.ground_type
			ghv.airtype = hd.air_type
			if hd.forcenofall {
				ghv.fallflag = false
			}
			if getter.ss.stateType == ST_A {
				ghv._type = ghv.airtype
			} else {
				ghv._type = ghv.groundtype
			}
			// If attack is guarded
			if hitResult == 2 {
				ghv.guarded = true
				ghv.hitshaketime = Max(0, hd.guard_pausetime[1])
				ghv.hittime = Max(0, hd.guard_hittime)
				ghv.slidetime = hd.guard_slidetime
				if getter.ss.stateType == ST_A {
					ghv.ctrltime = hd.airguard_ctrltime
					ghv.xvel = hd.airguard_velocity[0] * scaleratio * -byf
					ghv.yvel = hd.airguard_velocity[1] * scaleratio
					ghv.zvel = hd.airguard_velocity[2] * scaleratio
				} else {
					ghv.ctrltime = hd.guard_ctrltime
					ghv.xvel = hd.guard_velocity[0] * scaleratio * -byf
					// Mugen does not accept a Y component for ground guard velocity
					// But since we're adding Z to the other parameters, let's add Y here as well to keep things consistent
					ghv.yvel = hd.guard_velocity[1] * scaleratio
					ghv.zvel = hd.guard_velocity[2] * scaleratio
				}
				ghv.fallflag = false
				ghv.guardcount++
			} else {
				ghv.guarded = false
				ghv.hitshaketime = Max(0, hd.pausetime[1])
				ghv.slidetime = hd.ground_slidetime
				ghv.p2getp1state = hd.p2getp1state
				ghv.forcestand = hd.forcestand != 0
				ghv.forcecrouch = hd.forcecrouch != 0
				getter.fallTime = 0

				if hd.unhittabletime[1] >= 0 {
					getter.unhittableTime = hd.unhittabletime[1]
				}

				// Fall group
				ghv.fall_xvelocity = hd.fall_xvelocity * scaleratio
				ghv.fall_yvelocity = hd.fall_yvelocity * scaleratio
				ghv.fall_zvelocity = hd.fall_zvelocity * scaleratio
				ghv.fall_recover = hd.fall_recover
				ghv.fall_recovertime = hd.fall_recovertime
				ghv.fall_damage = hd.fall_damage
				ghv.fall_kill = hd.fall_kill
				ghv.fall_envshake_time = hd.fall_envshake_time
				ghv.fall_envshake_freq = hd.fall_envshake_freq
				ghv.fall_envshake_ampl = int32(float32(hd.fall_envshake_ampl) * scaleratio)
				ghv.fall_envshake_phase = hd.fall_envshake_phase
				ghv.fall_envshake_mul = hd.fall_envshake_mul
				ghv.fall_envshake_dir = hd.fall_envshake_dir
				if getter.ss.stateType == ST_A {
					ghv.hittime = hd.air_hittime
					// Note: ctrl time is not affected on hit in Mugen
					// This is further proof that gethitvars don't need to be reset above
					ghv.ctrltime = hd.air_hittime
					ghv.xvel = hd.air_velocity[0] * scaleratio * -byf
					ghv.yvel = hd.air_velocity[1] * scaleratio
					ghv.zvel = hd.air_velocity[2] * scaleratio
					ghv.fallflag = ghv.fallflag || hd.air_fall != 0
				} else if getter.ss.stateType == ST_L {
					ghv.hittime = hd.down_hittime
					ghv.ctrltime = hd.down_hittime
					ghv.fallflag = ghv.fallflag || hd.ground_fall
					if getter.pos[1] == 0 {
						ghv.xvel = hd.down_velocity[0] * scaleratio * -byf
						ghv.yvel = hd.down_velocity[1] * scaleratio
						ghv.zvel = hd.down_velocity[2] * scaleratio
						if !hd.down_bounce && ghv.yvel != 0 {
							ghv.fall_xvelocity = float32(math.NaN())
							ghv.fall_yvelocity = 0
							ghv.fall_zvelocity = float32(math.NaN())
						}
					} else {
						ghv.xvel = hd.air_velocity[0] * scaleratio * -byf
						ghv.yvel = hd.air_velocity[1] * scaleratio
						ghv.zvel = hd.air_velocity[1] * scaleratio
					}
				} else {
					ghv.ctrltime = hd.ground_hittime
					ghv.xvel = hd.ground_velocity[0] * scaleratio * -byf
					ghv.yvel = hd.ground_velocity[1] * scaleratio
					ghv.zvel = hd.ground_velocity[2] * scaleratio
					ghv.fallflag = ghv.fallflag || hd.ground_fall
					if ghv.fallflag && ghv.yvel == 0 {
						// Mugen does this as some form of internal workaround
						ghv.yvel = -0.001 * scaleratio
					}
					if ghv.yvel != 0 {
						ghv.hittime = hd.air_hittime
					} else {
						ghv.hittime = hd.ground_hittime
					}
				}
				if ghv.hittime < 0 {
					ghv.hittime = 0
				}
				// This compensates for characters being able to guard one frame sooner in Ikemen than in Mugen
				if c.stWgi().ikemenver[0] == 0 && c.stWgi().ikemenver[1] == 0 && ghv.hittime > 0 {
					ghv.hittime += 1
				}
				if cmb {
					ghv.hitcount++
				} else {
					ghv.hitcount = 1
				}
				ghv.down_recover = hd.down_recover
				// Down recovery time
				// When the char is already down this can't normally be increased
				// https://github.com/ikemen-engine/Ikemen-GO/issues/2026
				if hd.down_recovertime < 0 { // Default to char constant
					if ghv.down_recovertime > getter.gi().data.liedown.time || getter.ss.stateType != ST_L {
						ghv.down_recovertime = getter.gi().data.liedown.time
					}
				} else {
					if ghv.down_recovertime > hd.down_recovertime || getter.ss.stateType != ST_L {
						ghv.down_recovertime = hd.down_recovertime
					}
				}
			}
			// Anim reaction types
			ghv.groundanimtype = hd.animtype
			ghv.airanimtype = hd.air_animtype
			ghv.fall_animtype = hd.fall_animtype
			// Determine the actual animation type to use. Must be placed after ghv.yvel and the animtype ghv's
			ghv.animtype = getter.gethitAnimtype()
			// Min, Maxdist and Snap parameters
			// When Snap is defined, Min and Max distances are set to Snap
			byPos := c.pos
			if isProjectile {
				byPos[0] += posDiff[0]
				byPos[1] += posDiff[1]
				byPos[2] += posDiff[2]
			}
			snap := [...]float32{float32(math.NaN()), float32(math.NaN()), float32(math.NaN())}
			// MinDist X
			if !math.IsNaN(float64(hd.mindist[0])) {
				if byf < 0 {
					if getter.pos[0] > byPos[0]-hd.mindist[0] {
						snap[0] = byPos[0] - hd.mindist[0]
					}
				} else {
					if getter.pos[0] < byPos[0]+hd.mindist[0] {
						snap[0] = byPos[0] + hd.mindist[0]
					}
				}
			}
			// MaxDist X
			if !math.IsNaN(float64(hd.maxdist[0])) {
				if byf < 0 {
					if getter.pos[0]*(getter.localscl/c.localscl) < byPos[0]-hd.maxdist[0] {
						snap[0] = byPos[0] - hd.maxdist[0]
					}
				} else {
					if getter.pos[0]*(getter.localscl/c.localscl) > byPos[0]+hd.maxdist[0] {
						snap[0] = byPos[0] + hd.maxdist[0]
					}
				}
			}
			// Min and MaxDist Y
			if hitResult == 1 || getter.ss.stateType == ST_A {
				if !math.IsNaN(float64(hd.mindist[1])) {
					if getter.pos[1]*(getter.localscl/c.localscl) < byPos[1]+hd.mindist[1] {
						snap[1] = byPos[1] + hd.mindist[1]
					}
				}
				if !math.IsNaN(float64(hd.maxdist[1])) {
					if getter.pos[1]*(getter.localscl/c.localscl) > byPos[1]+hd.maxdist[1] {
						snap[1] = byPos[1] + hd.maxdist[1]
					}
				}
			}
			// Min and MaxDist Z
			if !math.IsNaN(float64(hd.mindist[2])) {
				if getter.pos[2]*(getter.localscl/c.localscl) < byPos[2]+hd.mindist[2] {
					snap[2] = byPos[2] + hd.mindist[2]
				}
			}
			if !math.IsNaN(float64(hd.maxdist[2])) {
				if getter.pos[2]*(getter.localscl/c.localscl) > byPos[2]+hd.maxdist[2] {
					snap[2] = byPos[2] + hd.maxdist[2]
				}
			}
			// Save snap offsets
			if !math.IsNaN(float64(snap[0])) {
				ghv.xoff = snap[0]*scaleratio - getter.pos[0]
			}
			if !math.IsNaN(float64(snap[1])) {
				ghv.yoff = snap[1]*scaleratio - getter.pos[1]
			}
			if !math.IsNaN(float64(snap[2])) {
				ghv.zoff = snap[2]*scaleratio - getter.pos[2]
			}
			// Snap time
			if hd.snaptime != 0 && getter.hoverIdx < 0 {
				getter.setBindToId(c, true)
				getter.setBindTime(hd.snaptime + Btoi(hd.snaptime > 0 && !c.pause()))
				getter.bindFacing = 0
				if !math.IsNaN(float64(snap[0])) {
					getter.bindPos[0] = hd.mindist[0] * scaleratio
				} else {
					getter.bindPos[0] = float32(math.NaN())
				}
				if !math.IsNaN(float64(snap[1])) &&
					(hitResult == 1 || getter.ss.stateType == ST_A) {
					getter.bindPos[1] = hd.mindist[1] * scaleratio
				} else {
					getter.bindPos[1] = float32(math.NaN())
				}
				if !math.IsNaN(float64(snap[2])) {
					getter.bindPos[2] = hd.mindist[2] * scaleratio
				} else {
					getter.bindPos[2] = float32(math.NaN())
				}
			} else if getter.bindToId == c.id {
				getter.setBindTime(0)
			}
			// Save other gethitvars that don't directly affect gameplay
			ghv.ground_velocity[0] = hd.ground_velocity[0] * scaleratio * -byf
			ghv.ground_velocity[1] = hd.ground_velocity[1] * scaleratio
			ghv.ground_velocity[2] = hd.ground_velocity[2] * scaleratio
			ghv.air_velocity[0] = hd.air_velocity[0] * scaleratio * -byf
			ghv.air_velocity[1] = hd.air_velocity[1] * scaleratio
			ghv.air_velocity[2] = hd.air_velocity[2] * scaleratio
			ghv.down_velocity[0] = hd.down_velocity[0] * scaleratio * -byf
			ghv.down_velocity[1] = hd.down_velocity[1] * scaleratio
			ghv.down_velocity[2] = hd.down_velocity[2] * scaleratio
			ghv.guard_velocity[0] = hd.guard_velocity[0] * scaleratio * -byf
			ghv.guard_velocity[1] = hd.guard_velocity[1] * scaleratio
			ghv.guard_velocity[2] = hd.guard_velocity[2] * scaleratio
			ghv.airguard_velocity[0] = hd.airguard_velocity[0] * scaleratio * -byf
			ghv.airguard_velocity[1] = hd.airguard_velocity[1] * scaleratio
			ghv.airguard_velocity[2] = hd.airguard_velocity[2] * scaleratio
			ghv.priority = hd.priority
		}
		// Hitting the enemy allows them to briefly move during a pause
		if sys.supertime > 0 {
			getter.superMovetime = Max(getter.superMovetime, getter.ghv.hitshaketime)
		} else if sys.pausetime > 0 {
			getter.pauseMovetime = Max(getter.pauseMovetime, getter.ghv.hitshaketime)
		}
		if !p2s && !getter.csf(CSF_gethit) {
			getter.stchtmp = false
		}
		// Flag enemy as getting hit
		getter.setCSF(CSF_gethit)
		getter.ghv.frame = true
		// In Mugen, having any HitOverride active allows GetHitVar Damage to exceed the remaining life
		bnd := true
		for _, ho := range getter.hover {
			if ho.time != 0 {
				bnd = false
				break
			}
		}
		// Damage on hit
		if hitResult == 1 {
			// Life
			if !getter.asf(ASF_nohitdamage) {
				getter.ghv.damage += getter.computeDamage(float64(hd.hitdamage), getter.ghv.kill, false, attackMul[0]*(float32(c.gi().attackBase)/100), c, bnd)
			}
			// Red life
			if !getter.asf(ASF_noredlifedamage) {
				getter.ghv.redlife += getter.computeDamage(float64(hd.hitredlife), true, false, attackMul[1]*(float32(c.gi().attackBase)/100), c, bnd)
			}
			// Dizzy points
			if !getter.asf(ASF_nodizzypointsdamage) && !getter.scf(SCF_dizzy) {
				getter.ghv.dizzypoints += getter.computeDamage(float64(hd.dizzypoints), true, false, attackMul[2]*(float32(c.gi().attackBase)/100), c, false)
			}
		}
		// Damage on guard
		if hitResult == 2 {
			// Life
			if !getter.asf(ASF_noguarddamage) {
				getter.ghv.damage += getter.computeDamage(float64(hd.guarddamage), getter.ghv.kill, false, attackMul[0]*(float32(c.gi().attackBase)/100), c, bnd)
			}
			// Red life
			if !getter.asf(ASF_noredlifedamage) {
				getter.ghv.redlife += getter.computeDamage(float64(hd.guardredlife), true, false, attackMul[1]*(float32(c.gi().attackBase)/100), c, bnd)
			}
			// Guard points
			if !getter.asf(ASF_noguardpointsdamage) {
				getter.ghv.guardpoints += getter.computeDamage(float64(hd.guardpoints), true, false, attackMul[3]*(float32(c.gi().attackBase)/100), c, false)
			}
		}
		// Save absolute values
		// These do not affect the player and are only used in GetHitVar
		getter.ghv.hitpower += hd.hitgivepower
		getter.ghv.guardpower += hd.guardgivepower
		getter.ghv.hitdamage += getter.computeDamage(float64(hd.hitdamage), true, false, attackMul[0]*(float32(c.gi().attackBase)/100), c, false)
		getter.ghv.guarddamage += getter.computeDamage(float64(hd.guarddamage), true, false, attackMul[0]*(float32(c.gi().attackBase)/100), c, false)
		getter.ghv.hitredlife += getter.computeDamage(float64(hd.hitredlife), true, false, attackMul[1]*(float32(c.gi().attackBase)/100), c, false)
		getter.ghv.guardredlife += getter.computeDamage(float64(hd.guardredlife), true, false, attackMul[1]*(float32(c.gi().attackBase)/100), c, false)

		// Hit behavior on KO
		if ghvset && getter.ghv.damage >= getter.life {
			if getter.ghv.kill || !getter.alive() {
				// Set fall behavior
				if !getter.asf(ASF_nokofall) {
					getter.ghv.fallflag = true
					getter.ghv.animtype = getter.gethitAnimtype() // Update to fall anim type
				}
				// Add extra velocity
				if getter.kovelocity && !getter.asf(ASF_nokovelocity) {
					startx := getter.ghv.xvel
					starty := getter.ghv.yvel
					if getter.ss.stateType == ST_A {
						if getter.ghv.xvel != 0 {
							getter.ghv.xvel += getter.gi().velocity.air.gethit.ko.add[0] * SignF(getter.ghv.xvel) * -1
						}
						if getter.ghv.yvel <= 0 {
							getter.ghv.yvel += getter.gi().velocity.air.gethit.ko.add[1]
							if getter.ghv.yvel > getter.gi().velocity.air.gethit.ko.ymin {
								getter.ghv.yvel = getter.gi().velocity.air.gethit.ko.ymin
							}
						}
					} else if getter.ss.stateType != ST_L {
						if getter.ghv.yvel == 0 {
							getter.ghv.xvel *= getter.gi().velocity.ground.gethit.ko.xmul
						}
						if getter.ghv.xvel != 0 {
							getter.ghv.xvel += getter.gi().velocity.ground.gethit.ko.add[0] * SignF(getter.ghv.xvel) * -1
						}
						if getter.ghv.yvel <= 0 {
							getter.ghv.yvel += getter.gi().velocity.ground.gethit.ko.add[1]
							if getter.ghv.yvel > getter.gi().velocity.ground.gethit.ko.ymin {
								getter.ghv.yvel = getter.gi().velocity.ground.gethit.ko.ymin
							}
						}
					}
					// Save difference to xveladd and yveladd
					// Not particularly useful, but it's a trigger that was documented in Mugen and did not work
					getter.ghv.xveladd = getter.ghv.xvel - startx
					getter.ghv.yveladd = getter.ghv.yvel - starty
				}
			} else {
				getter.ghv.damage = getter.life - 1
			}
		}
	}

	// Power management
	if hitResult > 0 {
		if Abs(hitResult) == 1 {
			c.powerAdd(hd.hitgetpower)
			if getter.playerFlag {
				getter.powerAdd(hd.hitgivepower)
				getter.ghv.power += hd.hitgivepower
			}
		} else {
			c.powerAdd(hd.guardgetpower)
			if getter.playerFlag {
				getter.powerAdd(hd.guardgivepower)
				getter.ghv.power += hd.guardgivepower
			}
		}
	}

	// Counter hit flag
	if hitResult == 1 {
		c.counterHit = getter.ss.moveType == MT_A
	}

	// Score and combo counters
	// ReversalDef can also add to them
	if Abs(hitResult) == 1 {
		if (ghvset || getter.csf(CSF_gethit)) && getter.hoverIdx < 0 &&
			!(c.hitdef.air_type == HT_None && getter.ss.stateType == ST_A || getter.ss.stateType != ST_A && c.hitdef.ground_type == HT_None) {
			getter.receivedHits += hd.numhits
			if c.teamside != -1 {
				sys.lifebar.co[c.teamside].combo += hd.numhits
			}
		}
		if !math.IsNaN(float64(hd.score[0])) {
			c.scoreAdd(hd.score[0])
			getter.ghv.score = hd.score[0] // TODO: The gethitvar refers to the enemy's score, which is counterintuitive
		}
		if getter.playerFlag {
			if !math.IsNaN(float64(hd.score[1])) {
				getter.scoreAdd(hd.score[1])
			}
		}
	}

	// Hitspark creation function
	// This used to be called only when a hitspark is actually created, but with the addition of the MoveHitVar trigger it became useful to save the offset at all times
	hitspark := func(p1, p2 *Char, animNo int32, ffx string, sparkangle float32, sparkscale [2]float32) {
		// This is mostly for offset in projectiles
		off := posDiff

		// Get reference position
		if !isProjectile {
			off[0] = p2.pos[0]*p2.localscl - p1.pos[0]*p1.localscl
			if (p1.facing < 0) != (p2.facing < 0) {
				off[0] += p2.facing * p2.sizeWidth[0] * p2.localscl
			} else {
				off[0] -= p2.facing * p2.sizeWidth[1] * p2.localscl
			}
			off[2] = p2.pos[2]*p2.localscl - p1.pos[2]*p1.localscl
		}
		off[0] *= p1.facing

		// Apply sparkxy
		if isProjectile {
			off[0] *= c.localscl
			off[1] *= c.localscl
			off[2] *= c.localscl
			off[0] += hd.sparkxy[0] * proj.facing * p1.facing * c.localscl
		} else {
			off[0] -= hd.sparkxy[0] * c.localscl
		}
		off[1] += hd.sparkxy[1] * c.localscl

		// Reversaldef spark (?)
		if c.id != p1.id {
			off[1] += p1.hitdef.sparkxy[1] * c.localscl
		}

		// Convert offset back to character's coordinate space
		for i := range off {
			off[i] /= c.localscl
		}

		// Save hitspark position to MoveHitVar
		if !isProjectile {
			c.mhv.sparkxy[0] = off[0]
			c.mhv.sparkxy[1] = off[1]
		}

		if animNo >= 0 {
			if e, i := c.spawnExplod(); e != nil {
				//e.anim = c.getAnim(animNo, ffx, true)
				e.animNo = animNo
				e.anim_ffx = ffx
				e.layerno = 1 // e.ontop = true
				e.sprpriority = math.MinInt32
				e.ownpal = true
				e.relativePos = [3]float32{off[0], off[1], off[2]}
				e.supermovetime = -1
				e.pausemovetime = -1
				e.scale = [2]float32{sparkscale[0], sparkscale[1]}
				//e.localscl = 1
				//if ffx == "" || ffx == "s" {
				//	e.scale = [2]float32{c.localscl * sparkscale[0], c.localscl * sparkscale[1]}
				//} else if e.anim != nil {
				//	e.anim.start_scale[0] *= c.localscl * sparkscale[0]
				//	e.anim.start_scale[1] *= c.localscl * sparkscale[1]
				//}
				e.setPos(p1)
				e.anglerot[0] = sparkangle
				c.commitExplod(i)
			}
		}
	}

	// Play hit sounds and sparks
	if Abs(hitResult) == 1 {
		if hd.reversal_attr > 0 {
			hitspark(getter, c, hd.sparkno, hd.sparkno_ffx, hd.sparkangle, hd.sparkscale)
		} else {
			hitspark(c, getter, hd.sparkno, hd.sparkno_ffx, hd.sparkangle, hd.sparkscale)
		}
		if hd.hitsound[0] >= 0 && hd.hitsound[1] >= 0 {
			vo := int32(100)
			c.playSound(hd.hitsound_ffx, false, 0, hd.hitsound[0], hd.hitsound[1],
				hd.hitsound_channel, vo, 0, 1, getter.localscl, &getter.pos[0], true, 0, 0, 0, 0, false, false)
		}
	} else {
		if hd.reversal_attr > 0 {
			hitspark(getter, c, hd.guard_sparkno, hd.guard_sparkno_ffx, hd.guard_sparkangle, hd.guard_sparkscale)
		} else {
			hitspark(c, getter, hd.guard_sparkno, hd.guard_sparkno_ffx, hd.guard_sparkangle, hd.guard_sparkscale)
		}
		if hd.guardsound[0] >= 0 && hd.guardsound[1] >= 0 {
			vo := int32(100)
			c.playSound(hd.guardsound_ffx, false, 0, hd.guardsound[0], hd.guardsound[1],
				hd.guardsound_channel, vo, 0, 1, getter.localscl, &getter.pos[0], true, 0, 0, 0, 0, false, false)
		}
	}

	// If not setting GetHitVars then the rest is skipped
	if !ghvset {
		return
	}

	getter.p1facing = 0
	invertXvel := func(byf float32) {
		if !isProjectile {
			if c.p1facing != 0 {
				byf = c.p1facing
			} else {
				byf = c.facing
			}
		}
		// Flip low and high hit animations when hitting enemy from behind
		if (getter.facing < 0) == (byf < 0) {
			if getter.ghv.groundtype == 1 || getter.ghv.groundtype == 2 {
				getter.ghv.groundtype += 3 - getter.ghv.groundtype*2
			}
			if getter.ghv.airtype == 1 || getter.ghv.airtype == 2 {
				getter.ghv.airtype += 3 - getter.ghv.airtype*2
			}
		}
	}
	if getter.hoverIdx >= 0 {
		invertXvel(byf)
		return
	}

	// HitOnce drops all targets except the current one
	if !isProjectile && hd.hitonce > 0 {
		c.targetDrop(-1, getter.id, true)
	}

	// Juggle points inheriting
	if c.helperIndex != 0 && c.inheritJuggle != 0 {
		// Update parent's or root's target list and juggle points
		sendJuggle := func(origin *Char) {
			origin.addTarget(getter.id)
			jg := origin.gi().data.airjuggle
			for _, v := range getter.ghv.targetedBy {
				if len(v) >= 2 && (v[0] == origin.id || v[0] == c.id) && v[1] < jg {
					jg = v[1]
				}
			}
			getter.ghv.dropId(origin.id)
			getter.ghv.targetedBy = append(getter.ghv.targetedBy, [...]int32{origin.id, jg - c.juggle})
		}
		if c.inheritJuggle == 1 && c.parent(false) != nil {
			sendJuggle(c.parent(false))
		} else if c.inheritJuggle == 2 && c.root(false) != nil {
			sendJuggle(c.root(false))
		}
	}

	// Add players to each other's lists
	c.addTarget(getter.id)
	getter.ghv.addId(c.id, c.gi().data.airjuggle)

	// On hit, reversal or type None
	if Abs(hitResult) == 1 {
		if !isProjectile && (hd.p1getp2facing != 0 || hd.p1facing < 0) &&
			c.facing != byf {
			c.p1facing = byf
		}
		if hd.p2facing < 0 {
			getter.p1facing = byf
		} else if hd.p2facing > 0 {
			getter.p1facing = -byf
		}
		if getter.p1facing == getter.facing {
			getter.p1facing = 0
		}
		getter.ghv.facing = hd.p2facing
		if hd.p1stateno >= 0 && c.stateChange1(hd.p1stateno, hd.playerNo) {
			c.setCtrl(false)
		}
		// Juggle points are subtracted if the target was falling either before or after the hit
		// https://github.com/ikemen-engine/Ikemen-GO/issues/2287
		if getter.prevfallflag || getter.ghv.fallflag {
			if !c.asf(ASF_nojugglecheck) {
				jug := &getter.ghv.targetedBy[len(getter.ghv.targetedBy)-1][1]
				if isProjectile {
					*jug -= hd.air_juggle
				} else {
					*jug -= c.juggle
				}
			}
			// Juggle cost is reset regardless of NoJuggleCheck
			// https://github.com/ikemen-engine/Ikemen-GO/issues/1905
			if !isProjectile {
				c.juggle = 0
			}
		}
		if hd.palfx.time > 0 && getter.palfx != nil {
			getter.palfx.clear2(true)
			getter.palfx.PalFXDef = hd.palfx
		}
		if hd.envshake_time > 0 {
			sys.envShake.time = hd.envshake_time
			sys.envShake.freq = hd.envshake_freq * float32(math.Pi) / 180
			sys.envShake.ampl = float32(int32(float32(hd.envshake_ampl) * c.localscl))
			sys.envShake.phase = hd.envshake_phase
			sys.envShake.mul = hd.envshake_mul
			sys.envShake.dir = hd.envshake_dir * float32(math.Pi) / 180
			sys.envShake.setDefaultPhase()
		}
		// Cornerpush on hit
		// In Mugen it is only set if the enemy is already in the corner before the hit
		// In Ikemen it is set regardless, with corner distance being checked later
		if hitResult > 0 && !isProjectile {
			switch getter.ss.stateType {
			case ST_S, ST_C:
				c.cornerVelOff = hd.ground_cornerpush_veloff * c.facing
			case ST_A:
				c.cornerVelOff = hd.air_cornerpush_veloff * c.facing
			case ST_L:
				c.cornerVelOff = hd.down_cornerpush_veloff * c.facing
			}
		}
	}
	// Cornerpush on block
	if hitResult == 2 && !isProjectile {
		switch getter.ss.stateType {
		case ST_S, ST_C:
			c.cornerVelOff = hd.guard_cornerpush_veloff * c.facing
		case ST_A:
			c.cornerVelOff = hd.airguard_cornerpush_veloff * c.facing
		}
	}
	invertXvel(byf)
	return
}

func (c *Char) actionPrepare() {
	if c.minus != 2 || c.csf(CSF_destroy) || c.scf(SCF_disabled) {
		return
	}
	c.pauseBool = false
	if c.cmd != nil {
		if sys.supertime > 0 {
			c.pauseBool = c.superMovetime == 0
		} else if sys.pausetime > 0 && c.pauseMovetime == 0 {
			c.pauseBool = true
		}
	}
	c.acttmp = -int8(Btoi(c.pauseBool)) * 2
	// Due to the nature of how pauses are processed, these are needed to fix an "off by 1" error in the PauseTime trigger
	c.prevSuperMovetime = c.superMovetime
	c.prevPauseMovetime = c.pauseMovetime
	if !c.pauseBool {
		// Perform basic actions
		if c.keyctrl[0] && c.cmd != nil && (c.helperIndex == 0 || c.controller >= 0) {
			// In Mugen, characters can perform basic actions even if they are KO
			if !c.asf(ASF_nohardcodedkeys) {
				if c.ctrl() {
					if c.scf(SCF_guard) && c.inguarddist && !c.inGuardState() && c.ss.stateType != ST_L && c.cmd[0].Buffer.Bb > 0 {
						c.changeState(120, -1, -1, "") // Start guarding
					} else if !c.asf(ASF_nojump) && c.ss.stateType == ST_S && c.cmd[0].Buffer.Ub > 0 &&
						(!(sys.intro < 0 && sys.intro > -sys.lifebar.ro.over_waittime) || c.asf(ASF_postroundinput)) {
						if c.ss.no != 40 {
							c.changeState(40, -1, -1, "") // Jump
						}
					} else if !c.asf(ASF_noairjump) && c.ss.stateType == ST_A && c.cmd[0].Buffer.Ub == 1 &&
						c.pos[1] <= -float32(c.gi().movement.airjump.height) &&
						c.airJumpCount < c.gi().movement.airjump.num {
						if c.ss.no != 45 || c.ss.time > 0 {
							c.airJumpCount++
							c.changeState(45, -1, -1, "") // Air jump
						}
					} else if !c.asf(ASF_nocrouch) && c.ss.stateType == ST_S && c.cmd[0].Buffer.Db > 0 {
						if c.ss.no != 10 {
							if c.ss.no != 100 {
								c.vel[0] = 0
							}
							c.changeState(10, -1, -1, "") // Stand to crouch
						}
					} else if !c.asf(ASF_nostand) && c.ss.stateType == ST_C && c.cmd[0].Buffer.Db <= 0 {
						if c.ss.no != 12 {
							c.changeState(12, -1, -1, "") // Crouch to stand
						}
					} else if !c.asf(ASF_nowalk) && c.ss.stateType == ST_S &&
						(c.cmd[0].Buffer.Fb > 0 != ((!c.inguarddist || c.prevNoStandGuard) && c.cmd[0].Buffer.Bb > 0)) {
						if c.ss.no != 20 {
							c.changeState(20, -1, -1, "") // Walk
						}
					}
				}
				// Braking is special in that it does not require ctrl
				if !c.asf(ASF_nobrake) && c.ss.no == 20 &&
					(c.cmd[0].Buffer.Bb > 0) == (c.cmd[0].Buffer.Fb > 0) {
					c.changeState(0, -1, -1, "")
				}
				// At least one character has been found where forcing them to stand up when crouching without ctrl will break them
			}
		}
		if c.ss.stateType != ST_A {
			c.airJumpCount = 0
		}
		if !c.hitPause() {
			c.specialFlag = 0
			c.setCSF(CSF_stagebound)
			if c.playerFlag {
				if c.alive() || c.ss.no != 5150 || c.numPartner() == 0 {
					c.setCSF(CSF_screenbound | CSF_movecamera_x | CSF_movecamera_y)
				}
				if sys.roundState() > 0 && (c.alive() || c.numPartner() == 0) {
					c.setCSF(CSF_playerpush)
				}
			}
			// Reset player pushing priority
			c.pushPriority = 0
			// HitBy timers
			// In Mugen this seems to happen at the end of each frame instead
			for i := range c.hitby {
				if c.hitby[i].time > 0 {
					c.hitby[i].time--
					if c.hitby[i].time == 0 {
						c.hitby[i].clear()
					}
				}
			}
			// HitOverride timers
			// In Mugen they decrease even during hitpause. However no issues have arised from not doing that yet
			for i := range c.hover {
				if c.hover[i].time > 0 {
					c.hover[i].time--
					if c.hover[i].time == 0 {
						c.hover[i].clear()
					}
				}
			}
			if sys.supertime > 0 {
				if c.superMovetime > 0 {
					c.superMovetime--
				}
			} else if sys.pausetime > 0 && c.pauseMovetime > 0 {
				c.pauseMovetime--
			}
			if c.ignoreDarkenTime > 0 {
				c.ignoreDarkenTime--
			}
		}
		// Reset input modifiers
		c.inputFlag = 0
		c.inputShift = c.inputShift[:0]
		// This AssertSpecial flag is special in that it must always reset regardless of hitpause
		c.unsetASF(ASF_animatehitpause)
		// The flags in this block are to be reset even during hitpause
		// Exception for WinMugen chars, where they persisted during hitpause
		if c.stWgi().ikemenver[0] != 0 || c.stWgi().ikemenver[1] != 0 || c.stWgi().mugenver[0] == 1 || !c.hitPause() {
			c.unsetCSF(CSF_angledraw)
			c.angleDrawScale = [2]float32{1, 1}
			c.trans = TT_default
			c.alpha = [2]int32{255, 0}
			c.offset = [2]float32{}
			// Reset all AssertSpecial flags except the following, which are reset elsewhere in the code
			c.assertFlag = (c.assertFlag&ASF_nostandguard | c.assertFlag&ASF_nocrouchguard | c.assertFlag&ASF_noairguard |
				c.assertFlag&ASF_runfirst | c.assertFlag&ASF_runlast)
		}
		// The flags below also reset during hitpause, but are new to Ikemen and don't need the exception above
		// Reset Clsn modifiers
		c.clsnScaleMul = [2]float32{1.0, 1.0}
		c.clsnAngle = 0
		// Reset modifyShadow
		c.shadowColor = [3]int32{-1, -1, -1}
		c.shadowIntensity = -1
		c.shadowOffset = [2]float32{}
		c.shadowWindow = [4]float32{}
		c.shadowXshear = 0
		c.shadowYscale = 0
		c.shadowRot = Rotation{0, 0, 0}
		c.shadowProjection = -1
		c.shadowfLength = 0
		// Reset modifyReflection
		c.reflectColor = [3]int32{-1, -1, -1}
		c.reflectIntensity = -1
		c.reflectOffset = [2]float32{}
		c.reflectWindow = [4]float32{}
		c.reflectXshear = 0
		c.reflectYscale = 0
		c.reflectRot = Rotation{0, 0, 0}
		c.reflectProjection = -1
		c.reflectfLength = 0
		// Reset TransformSprite
		c.window = [4]float32{}
		c.xshear = 0
		c.fLength = 2048
		c.projection = Projection_Orthographic
	}
	// Decrease unhittable timer
	// This used to be in tick(), but Mugen Clsn display suggests it happens sooner than that
	// This also used to be CharGlobalInfo, but that made root and helpers share the same timer
	// In Mugen this timer won't decrease unless the char has a Clsn box (of any type)
	if c.unhittableTime > 0 {
		c.unhittableTime--
	}
	c.dropTargets()
	// Enable autoguard. This placement gives it similar properties to other AssertSpecial flags
	if sys.cfg.Options.AutoGuard {
		c.setASF(ASF_autoguard)
	}
}

func (c *Char) actionRun() {
	if c.minus != 2 || c.csf(CSF_destroy) || c.scf(SCF_disabled) {
		return
	}
	// Run state -4
	c.minus = -4
	if sb, ok := c.gi().states[-4]; ok {
		sb.run(c)
	}
	if !c.pauseBool {
		// Run state -3
		c.minus = -3
		if c.ss.sb.playerNo == c.playerNo && (c.playerFlag || c.keyctrl[2]) {
			if sb, ok := c.gi().states[-3]; ok {
				sb.run(c)
			}
		}
		// Run state -2
		c.minus = -2
		if c.playerFlag || c.keyctrl[1] {
			if sb, ok := c.gi().states[-2]; ok {
				sb.run(c)
			}
		}
		// Run state -1
		c.minus = -1
		if c.ss.sb.playerNo == c.playerNo && c.keyctrl[0] {
			if sb, ok := c.gi().states[-1]; ok {
				sb.run(c)
			}
		}
		// Change into buffered state
		c.stateChange2()
		// Run current state
		c.minus = 0
		c.ss.sb.run(c)
	}
	// Guarding instructions
	c.unsetSCF(SCF_guard)
	if ((c.scf(SCF_ctrl) || c.ss.no == 52) &&
		c.ss.moveType == MT_I || c.inGuardState()) && c.cmd != nil &&
		(c.cmd[0].Buffer.Bb > 0 || c.asf(ASF_autoguard)) &&
		(c.ss.stateType == ST_S && !c.asf(ASF_nostandguard) ||
			c.ss.stateType == ST_C && !c.asf(ASF_nocrouchguard) ||
			c.ss.stateType == ST_A && !c.asf(ASF_noairguard)) {
		c.setSCF(SCF_guard)
	}
	if !c.pauseBool {
		if c.keyctrl[0] && c.cmd != nil {
			if c.ctrl() && (c.controller >= 0 || c.helperIndex == 0) {
				if !c.asf(ASF_nohardcodedkeys) {
					if c.inguarddist && c.scf(SCF_guard) && !c.inGuardState() && c.cmd[0].Buffer.Bb > 0 {
						c.changeState(120, -1, -1, "")
						// In Mugen the characters *can* change to the guarding states during pauses
						// They can still block in Ikemen despite not changing state here
					}
				}
			}
		}
	}
	// Run state +1
	// Uses minus -4 because its properties are similar
	c.minus = -4
	if sb, ok := c.gi().states[-10]; ok {
		sb.run(c)
	}
	// Set minus back to normal
	c.minus = 0
	// If State +1 changed the current state, run the next one as well
	if !c.pauseBool && c.stchtmp {
		c.stateChange2()
		c.ss.sb.run(c)
	}
	// Reset char width and height values
	// TODO: Some of this code could probably be integrated with the new size box
	if !c.hitPause() {
		coordRatio := ((320 / c.localcoord) / c.localscl)
		if !c.csf(CSF_width) {
			c.sizeWidth = [2]float32{c.baseWidthFront() * coordRatio, c.baseWidthBack() * coordRatio}
		}
		if !c.csf(CSF_widthedge) {
			c.edgeWidth = [2]float32{0, 0}
		}
		if !c.csf(CSF_height) {
			c.sizeHeight = [2]float32{c.baseHeightTop() * coordRatio, c.baseHeightBottom() * coordRatio}
		}
		if !c.csf(CSF_depth) {
			c.sizeDepth = [2]float32{c.baseDepthTop() * coordRatio, c.baseDepthBottom() * coordRatio}
		}
		if !c.csf(CSF_depthedge) {
			c.edgeDepth = [2]float32{0, 0}
		}
	}
	c.updateSizeBox()
	if !c.pauseBool {
		if !c.hitPause() {
			// In Mugen chars are forced to stay in state 5110 at least one frame before getting up
			if c.ss.no == 5110 && c.ss.time >= 1 && c.ghv.down_recovertime <= 0 && c.alive() && !c.asf(ASF_nogetupfromliedown) {
				c.changeState(5120, -1, -1, "")
			}
			for c.ss.no == 140 && (c.anim == nil || len(c.anim.frames) == 0 ||
				c.ss.time >= c.anim.totaltime) {
				c.changeState(Btoi(c.ss.stateType == ST_C)*11+
					Btoi(c.ss.stateType == ST_A)*51, -1, -1, "")
			}
			c.posUpdate()
			// Land from aerial physics
			// This was a loop before like Mugen, so setting state 52 to physics A caused a crash
			if c.ss.physics == ST_A {
				if c.vel[1] > 0 && (c.pos[1]-c.groundLevel-c.platformPosY) >= 0 && c.ss.no != 105 {
					c.changeState(52, -1, -1, "")
				}
			}
			c.groundLevel = 0 // Reset only after position has been updated
			c.setFacing(c.p1facing)
			c.p1facing = 0
			c.ss.time++
			if c.mctime > 0 {
				c.mctime++
			}
		}
		// Commit current animation frame to memory
		// This frame will be used for hit detection and as reference for Lua scripts (including debug info)
		if !c.hitPause() || c.asf(ASF_animatehitpause) {
			c.updateCurFrame()
		}
		if c.ghv.damage != 0 {
			if c.ss.moveType == MT_H {
				c.lifeAdd(-float64(c.ghv.damage), true, true)
			}
			c.ghv.damage = 0
		}
		if c.ghv.redlife != 0 {
			if c.ss.moveType == MT_H {
				c.redLifeAdd(-float64(c.ghv.redlife), true)
			}
			c.ghv.redlife = 0
		}
		if c.ghv.dizzypoints != 0 {
			if c.ss.moveType == MT_H {
				c.dizzyPointsAdd(-float64(c.ghv.dizzypoints), true)
			}
			c.ghv.dizzypoints = 0
		}
		if c.ghv.guardpoints != 0 {
			if c.ss.moveType == MT_H {
				c.guardPointsAdd(-float64(c.ghv.guardpoints), true)
			}
			c.ghv.guardpoints = 0
		}
		c.ghv.hitdamage = 0
		c.ghv.guarddamage = 0
		c.ghv.power = 0
		c.ghv.hitpower = 0
		c.ghv.guardpower = 0
		// The following block used to be in char.update()
		// That however caused a breaking difference with Mugen when checking these variables between different players
		// https://github.com/ikemen-engine/Ikemen-GO/issues/1540
		if !c.hitPause() {
			if c.ss.moveType == MT_H {
				if c.ghv.guarded {
					c.receivedDmg = 0
					c.receivedHits = 0
				}
				if c.ghv.hitshaketime > 0 {
					c.ghv.hitshaketime--
				}
				if c.ghv.fallflag {
					c.fallTime++
				}
			} else {
				if c.hittmp > 0 {
					c.hittmp = 0
				}
				if !c.scf(SCF_dizzy) {
					// HitOverride KeepState used to freeze some GetHitVars around here to keep them from resetting instantly,
					// but that no longer seems necessary with this being placed in actionRun()
					c.ghv.hitshaketime = 0
					c.ghv.attr = 0
					c.ghv.guardflag = 0
					c.ghv.playerId = 0
					c.ghv.playerNo = -1
					c.superDefenseMul = 1
					c.superDefenseMulBuffer = 1
					c.fallDefenseMul = 1
					c.ghv.fallflag = false
					c.ghv.fallcount = 0
					c.ghv.hitid = c.ghv.hitid >> 31
					// HitCount doesn't reset here, like Mugen, but there's no apparent reason to keep that behavior with GuardCount
					c.ghv.guardcount = 0
					c.receivedDmg = 0
					c.receivedHits = 0
					c.ghv.score = 0
					c.ghv.down_recovertime = c.gi().data.liedown.time
					// In Mugen, when returning to idle, characters cannot act until the next frame
					// To account for this, combos in Mugen linger one frame longer than they normally would in a fighting game
					// Ikemen's "fake combo" code used to replicate this behavior
					// After guarding was adjusted so that chars could guard when returning to idle, the fake combo code became obsolete
					// https://github.com/ikemen-engine/Ikemen-GO/issues/597
					//if c.comboExtraFrameWindow <= 0 {
					//	c.fakeReceivedHits = 0
					//	c.fakeComboDmg = 0
					//	c.fakeCombo = false
					//} else {
					//	c.fakeCombo = true
					//	c.comboExtraFrameWindow--
					//}
				}
			}
			if c.ghv.hitshaketime <= 0 && c.ghv.hittime >= 0 {
				c.ghv.hittime--
			}
			if c.ghv.down_recovertime > 0 && c.ss.no == 5110 {
				c.ghv.down_recovertime--
			}
			// Reset juggle points
			// Mugen does not do this by default, so it is often overlooked
			if c.ss.moveType != MT_A {
				if c.stWgi().ikemenver[0] != 0 || c.stWgi().ikemenver[1] != 0 {
					c.juggle = 0
				}
			}
		}
		if c.helperIndex == 0 && c.gi().pctime >= 0 {
			c.gi().pctime++
		}
		c.dustTime++
	}
	c.xScreenBound()
	c.zDepthBound()
	if !c.pauseBool {
		for _, tid := range c.targets {
			if t := sys.playerID(tid); t != nil && t.bindToId == c.id {
				t.bind()
			}
		}
	}
	c.minus = 1
	c.acttmp += int8(Btoi(!c.pause() && !c.hitPause())) - int8(Btoi(c.hitPause()))
}

func (c *Char) actionFinish() {
	if c.minus < 1 || c.csf(CSF_destroy) || c.scf(SCF_disabled) {
		return
	}
	if !c.pauseBool {
		if c.palfx != nil && c.ownpal {
			c.palfx.step()
		}
		// Placing these two in Finish instead of Run makes them less susceptible to processing order inconsistency
		c.ghv.frame = false
		c.mhv.frame = false
	}
	// Reset inguarddist flag before running hit detection (where it will be updated)
	// https://github.com/ikemen-engine/Ikemen-GO/issues/2328
	c.inguarddist = false
	// This variable is necessary because NoStandGuard is reset before the walking instructions are checked
	// https://github.com/ikemen-engine/Ikemen-GO/issues/1966
	c.prevNoStandGuard = c.asf(ASF_nostandguard)
	c.unsetASF(ASF_nostandguard | ASF_nocrouchguard | ASF_noairguard)
	// Save current HitFall value before hit detection
	c.prevfallflag = c.ghv.fallflag
	// Update Z scale
	// Must be placed after posUpdate()
	c.zScale = sys.updateZScale(c.pos[2], c.localscl)
	// KO behavior
	if !c.hitPause() && !c.pauseBool {
		if c.alive() && c.life <= 0 && !sys.gsf(GSF_globalnoko) && !c.asf(ASF_noko) && (!c.ghv.guarded || !c.asf(ASF_noguardko)) {
			// KO sound
			if !sys.gsf(GSF_nokosnd) {
				c.playSound("", false, 0, 11, 0, -1, 100, 0, 1, c.localscl, &c.pos[0], false, 0, 0, 0, 0, false, false)
				if c.gi().data.ko.echo != 0 {
					c.koEchoTimer = 1
				}
			}
			// Set KO flag and force KO states if necessary
			c.setSCF(SCF_ko)
			c.unsetSCF(SCF_ctrl) // This can be seen in Mugen when you F1 a character with ctrl && movetype = H
			if !c.stchtmp && c.helperIndex == 0 && c.ss.moveType != MT_H {
				c.ghv.fallflag = true
				c.selfState(5030, -1, -1, 0, "")
				c.ss.time = 1
			}
		}
	}
	// Over flags (char is finished for the round)
	if c.alive() && c.life > 0 && !sys.roundEnded() {
		c.unsetSCF(SCF_over_alive | SCF_over_ko)
	}
	if c.ss.no == 5150 && !c.scf(SCF_over_ko) { // Actual KO is not required in Mugen
		c.setSCF(SCF_over_ko)
	}
	c.minus = 1
}

func (c *Char) track() {
	if c.trackableByCamera() {

		// This doesn't seem necessary currently. Handled by xScreenBound()
		//if !sys.cam.roundstart && c.csf(CSF_screenbound) && !c.scf(SCF_standby) {
		//	c.interPos[0] = ClampF(c.interPos[0], min+sys.xmin/c.localscl, max+sys.xmax/c.localscl)
		//}

		// X axis
		if c.csf(CSF_movecamera_x) && !c.scf(SCF_standby) {
			edgeleft, edgeright := -c.edgeWidth[1], c.edgeWidth[0]
			if c.facing < 0 {
				edgeleft, edgeright = -edgeright, -edgeleft
			}

			charleft := c.interPos[0]*c.localscl + edgeleft*c.localscl
			charright := c.interPos[0]*c.localscl + edgeright*c.localscl
			canmove := c.acttmp > 0 && !c.csf(CSF_posfreeze) && (c.bindTime == 0 || math.IsNaN(float64(c.bindPos[0])))

			if charleft < sys.cam.leftest {
				sys.cam.leftest = charleft
				if canmove {
					sys.cam.leftestvel = c.vel[0] * c.localscl * c.facing
				} else {
					sys.cam.leftestvel = 0
				}
			}
			if charright > sys.cam.rightest {
				sys.cam.rightest = charright
				if canmove {
					sys.cam.rightestvel = c.vel[0] * c.localscl * c.facing
				} else {
					sys.cam.rightestvel = 0
				}
			}
		}

		// Y axis
		if c.csf(CSF_movecamera_y) && !c.scf(SCF_standby) && !math.IsInf(float64(c.pos[1]), 0) {
			sys.cam.highest = MinF(c.interPos[1]*c.localscl, sys.cam.highest)
			sys.cam.lowest = MaxF(c.interPos[1]*c.localscl, sys.cam.lowest)
			//sys.cam.Pos[1] = 0 // This doesn't seem necessary in the current state of the code
			// Mugen ignores characters that have infinite position
			// https://github.com/ikemen-engine/Ikemen-GO/issues/1917
		}
	}
}

// This function runs every tick unlike the others
func (c *Char) update() {
	if c.scf(SCF_disabled) {
		return
	}
	if sys.tickFrame() {
		if c.csf(CSF_destroy) {
			c.destroy()
			return
		}
		if !c.pause() && !c.isTargetBound() {
			c.bind()
		}
		if c.acttmp > 0 {
			if c.inGuardState() {
				c.setSCF(SCF_guard)
			}
			if c.anim != nil {
				c.anim.UpdateSprite()
			}
			if c.ss.moveType == MT_H {
				if c.ghv.xoff != 0 {
					c.setPosX(c.pos[0] + c.ghv.xoff)
					c.ghv.xoff = 0
				}
				if c.ghv.yoff != 0 {
					c.setPosY(c.pos[1] + c.ghv.yoff)
					c.ghv.yoff = 0
				}
				if c.ghv.zoff != 0 {
					c.setPosZ(c.pos[2] + c.ghv.zoff)
					c.ghv.zoff = 0
				}
			}
			// Engine dust effects
			if sys.supertime == 0 && sys.pausetime == 0 &&
				((c.ss.moveType == MT_H && (c.ss.stateType == ST_S || c.ss.stateType == ST_C)) || c.ss.no == 52) &&
				c.pos[1] == 0 && (AbsF(c.pos[0]-c.dustOldPos[0]) >= 1 || AbsF(c.pos[2]-c.dustOldPos[2]) >= 1) {
				c.makeDust(0, 0, 0, 3) // Default spacing of 3
			}
		}
		if c.ss.moveType == MT_H {
			// Set opposing team's First Attack flag
			if sys.firstAttack[2] == 0 && (c.teamside == 0 || c.teamside == 1) {
				if sys.firstAttack[1-c.teamside] < 0 && c.ghv.playerNo >= 0 && c.ghv.guarded == false {
					sys.firstAttack[1-c.teamside] = c.ghv.playerNo
				}
			}
			// Cancel pause move times
			if sys.supertime <= 0 && sys.pausetime <= 0 {
				c.superMovetime, c.pauseMovetime = 0, 0
			}
			// Fall mechanics
			c.hittmp = int8(Btoi(c.ghv.fallflag)) + 1
			if c.acttmp > 0 && (c.ss.no == 5070 || c.ss.no == 5100) && c.ss.time == 1 {
				if !c.asf(ASF_nofalldefenceup) {
					c.fallDefenseMul *= c.gi().data.fall.defence_mul
				}
				if !c.asf(ASF_nofallcount) {
					c.ghv.fallcount++
				}
				// Mugen does not actually require the "fallcount" condition here
				// But that makes characters always invulnerable if their lie down time constant is <= 10
				if c.ghv.fallcount > 1 && c.ss.no == 5100 {
					if c.ghv.down_recovertime > 0 {
						c.ghv.down_recovertime = int32(math.Floor(float64(c.ghv.down_recovertime) / 2))
					}
					//if c.ghv.fallcount > 3 || c.ghv.down_recovertime <= 0 {
					if c.ghv.down_recovertime <= 10 {
						c.hitby[0].flag = ^int32(ST_SCA)
						c.hitby[0].time = 180 // Mugen uses infinite time here
					}
				}
			}
		}
		// Remove self from target lists of other players
		// In Mugen, this seems to happen only after hit detection and even if the players are paused
		// This placement makes more sense but this difference seems partially responsible for the test case in issue #1891 not working currently
		// Also https://github.com/ikemen-engine/Ikemen-GO/issues/1592
		if c.ss.moveType != MT_H || c.ss.no == 5150 {
			c.exitTarget()
		}
		c.platformPosY = 0
		c.groundAngle = 0
		// Hit detection should happen even during hitpause
		// https://github.com/ikemen-engine/Ikemen-GO/issues/1660
		c.atktmp = int8(Btoi(c.ss.moveType != MT_I || c.hitdef.reversal_attr > 0))
		c.hoverIdx = -1
		c.hoverKeepState = false
		// Apply buffered SuperPause p2defmul
		if c.superDefenseMulBuffer != 1 {
			c.superDefenseMul *= c.superDefenseMulBuffer
			c.superDefenseMulBuffer = 1
		}
		// Update final defense
		var customDefense float32 = 1
		if !c.defenseMulDelay || c.ss.moveType == MT_H {
			customDefense = c.customDefense
		}
		c.finalDefense = float64(((float32(c.gi().defenceBase) * customDefense * c.superDefenseMul * c.fallDefenseMul) / 100))
	}
	// Update position interpolation
	if c.acttmp > 0 {
		spd := sys.tickInterpolation()
		if c.pushed {
			spd = 0
		}
		if !c.csf(CSF_posfreeze) {
			for i := 0; i < 3; i++ {
				c.interPos[i] = c.pos[i] - (c.pos[i]-c.oldPos[i])*(1-spd)
			}
		}
	}
	// KO sound echo
	if c.koEchoTimer > 0 {
		if !c.scf(SCF_ko) || sys.gsf(GSF_nokosnd) {
			c.koEchoTimer = 0
		} else {
			if c.koEchoTimer == 60 || c.koEchoTimer == 120 {
				vo := int32(100 * (240 - (c.koEchoTimer + 60)) / 240)
				c.playSound("", false, 0, 11, 0, -1, vo, 0, 1, c.localscl, &c.pos[0], false, 0, 0, 0, 0, false, false)
			}
			c.koEchoTimer++
		}
	}
}

// This function runs during tickNextFrame. After collision detection
func (c *Char) tick() {
	if c.scf(SCF_disabled) {
		return
	}
	// Step animation
	if c.acttmp > 0 || !c.pauseBool && (!c.hitPause() || c.asf(ASF_animatehitpause)) {
		// Update reference frame first
		c.updateCurFrame()
		// Animate
		if c.anim != nil && !c.asf(ASF_animfreeze) {
			c.anim.Action()
		}
		// Save last valid drawing frame
		// This step prevents the char from disappearing when changing animation during hitpause
		// Because the whole c.anim is used for rendering, saving just the sprite is not enough
		// https://github.com/ikemen-engine/Ikemen-GO/issues/1550
		c.animBackup = c.anim
	}
	if c.bindTime > 0 {
		if c.isTargetBound() {
			bt := sys.playerID(c.bindToId)
			if bt == nil || bt.csf(CSF_gethit) || bt.csf(CSF_destroy) {
				// SelfState if binder gets hit or destroys self
				// https://github.com/ikemen-engine/Ikemen-GO/issues/2347
				c.selfState(5050, -1, -1, -1, "")
				c.gethitBindClear()
			} else if !bt.pause() {
				//setBindTime is not used here because the CSF_destroy flag may be enabled in a frame with BindTime=0. If bindTime becomes 0, the setBindTime processing will be performed later
				c.bindTime -= 1
				//c.setBindTime(c.bindTime - 1)
			}
		} else {
			if !c.pause() {
				// c.bindTime -= 1
				c.setBindTime(c.bindTime - 1)
				// The fix below was necessary before because bindTime should not be decremented directly but rather via setBindTime
				// Fixes BindToRoot/BindToParent of 1 immediately after PosSets (MUGEN 1.0/1.1 behavior)
				// This must not run for target binds so that they end the same time as MUGEN's do.
				//if c.bindToId > 0 {
				//	c.setBindTime(c.bindTime)
				//}
			}
		}
	}
	if c.cmd == nil {
		if c.keyctrl[0] {
			c.cmd = make([]CommandList, len(sys.chars))
			c.cmd[0].Buffer = NewInputBuffer()
			for i := range c.cmd {
				c.cmd[i].Buffer = c.cmd[0].Buffer
				c.cmd[i].CopyList(sys.chars[c.playerNo][0].cmd[i])
				c.cmd[i].BufReset()
			}
		} else {
			c.cmd = sys.chars[c.playerNo][0].cmd
		}
	}
	if c.hitdefContact {
		if c.hitdef.hitonce != 0 || c.moveReversed() != 0 {
			c.hitdef.updateStateType(c.ss.stateType)
		}
		c.hitdefContact = false
	} else if c.hitdef.ltypehit {
		c.hitdef.attr = c.hitdef.attr&^int32(ST_MASK) | int32(c.ss.stateType)
		c.hitdef.ltypehit = false
	}
	// Get Hitdef targets from the buffer. Using a buffer mitigates processing order errors
	// https://github.com/ikemen-engine/Ikemen-GO/issues/1798
	if len(c.hitdefTargetsBuffer) > 0 {
		c.hitdefTargets = append(c.hitdefTargets, c.hitdefTargetsBuffer...)
		c.hitdefTargetsBuffer = c.hitdefTargetsBuffer[:0]
	}
	if c.mctime < 0 {
		c.mctime = 1
		if c.mctype == MC_Hit {
			c.hitCount += c.hitdef.numhits
		} else if c.mctype == MC_Guarded {
			c.guardCount += c.hitdef.numhits
		}
	}
	// Change to get hit states
	if c.csf(CSF_gethit) && !c.hoverKeepState {
		// This flag prevents prevMoveType from being changed twice
		c.ss.storeMoveType = true
		c.ss.changeMoveType(MT_H)
		if c.hitPauseTime > 0 {
			c.ss.clearHitPauseExecutionToggleFlags()
		}
		c.hitPauseTime = 0
		//c.targetDrop(-1, false) // GitHub #1148
		pn := c.playerNo
		if c.ghv.p2getp1state && !c.ghv.guarded {
			pn = c.ghv.playerNo
		}
		if c.stchtmp {
			// For Mugen compatibility, PrevStateNo returns these values if the character is hit into a custom state
			// https://github.com/ikemen-engine/Ikemen-GO/issues/765
			// This could maybe be disabled if the state owner is an Ikemen character
			// Maybe what actually happens in Mugen is P2StateNo is handled later like HitOverride
			if c.ss.stateType == ST_L && c.pos[1] == 0 {
				c.ss.prevno = 5080
			} else if c.ghv._type == HT_Trip {
				c.ss.prevno = 5070
			} else if c.ss.stateType == ST_S {
				c.ss.prevno = 5000
			} else if c.ss.stateType == ST_C {
				c.ss.prevno = 5010
			} else {
				c.ss.prevno = 5020
			}
		} else if c.ghv.guarded &&
			(c.ghv.damage < c.life || sys.gsf(GSF_globalnoko) || c.asf(ASF_noko) || c.asf(ASF_noguardko)) {
			switch c.ss.stateType {
			// All of these state changes remove ctrl from the char
			// Guarding is not affected by P2getP1state
			case ST_S:
				c.selfState(150, -1, -1, 0, "")
			case ST_C:
				c.selfState(152, -1, -1, 0, "")
			default:
				c.selfState(154, -1, -1, 0, "")
			}
		} else if c.ss.stateType == ST_L && c.pos[1] == 0 {
			c.changeStateEx(5080, pn, -1, 0, "")
		} else if c.ghv._type == HT_Trip {
			c.changeStateEx(5070, pn, -1, 0, "")
		} else {
			if c.ghv.forcestand && c.ss.stateType == ST_C {
				c.ss.changeStateType(ST_S)
			} else if c.ghv.forcecrouch && c.ss.stateType == ST_S {
				c.ss.changeStateType(ST_C)
			}
			switch c.ss.stateType {
			case ST_S:
				c.changeStateEx(5000, pn, -1, 0, "")
			case ST_C:
				// Go to standing on KO
				// Mugen does this, but does it really need to be hardcoded?
				if c.ghv.damage >= c.life && !sys.gsf(GSF_globalnoko) && !c.asf(ASF_noko) {
					c.changeStateEx(5000, pn, -1, 0, "")
				} else {
					c.changeStateEx(5010, pn, -1, 0, "")
				}
			default:
				c.changeStateEx(5020, pn, -1, 0, "")
			}
		}
		// Prepare down get hit offset
		if c.ss.stateType == ST_L && c.pos[1] == 0 && c.ghv.yvel != 0 {
			c.downHitOffset = true
		}
	}
	// Change to HitOverride state
	// This doesn't actually require getting hit
	// https://github.com/ikemen-engine/Ikemen-GO/issues/2262
	if c.hoverIdx >= 0 && c.hoverIdx < len(c.hover) && !c.hoverKeepState {
		if c.hover[c.hoverIdx].stateno >= 0 {
			c.stateChange1(c.hover[c.hoverIdx].stateno, c.hover[c.hoverIdx].playerNo)
		}
	}
	if !c.pause() {
		if c.hitPauseTime > 0 {
			c.hitPauseTime--
			if c.hitPauseTime == 0 {
				c.ss.clearHitPauseExecutionToggleFlags()
			}
		}
		// Fast recovery from lie down
		if c.ghv.down_recover && c.ghv.down_recovertime > 0 &&
			!c.asf(ASF_nofastrecoverfromliedown) &&
			(c.ghv.fallcount > 0 || c.ss.stateType == ST_L) &&
			(c.cmd[0].Buffer.Bb == 1 || c.cmd[0].Buffer.Db == 1 ||
				c.cmd[0].Buffer.Fb == 1 || c.cmd[0].Buffer.Ub == 1 ||
				c.cmd[0].Buffer.ab == 1 || c.cmd[0].Buffer.bb == 1 ||
				c.cmd[0].Buffer.cb == 1 || c.cmd[0].Buffer.xb == 1 ||
				c.cmd[0].Buffer.yb == 1 || c.cmd[0].Buffer.zb == 1 ||
				c.cmd[0].Buffer.sb == 1 || c.cmd[0].Buffer.db == 1 ||
				c.cmd[0].Buffer.wb == 1) { // Menu button not included
			c.ghv.down_recovertime -= RandI(1, (c.ghv.down_recovertime+1)/2)
		}
	}
	// Reset pushed flag
	// This flag is apparently used to prevent position interpolation when chars push each other
	c.pushed = false
}

// Prepare collision boxes and debug text for drawing
func (c *Char) cueDebugDraw() {
	x := c.pos[0] * c.localscl
	y := c.pos[1] * c.localscl
	xoff := x + c.offsetX()*c.localscl
	yoff := y + c.offsetY()*c.localscl
	xs := c.clsnScale[0] * c.facing
	ys := c.clsnScale[1]
	angle := c.clsnAngle * c.facing
	nhbtxt := ""
	// Debug Clsn display
	if sys.clsnDisplay {
		if c.curFrame != nil {
			// Add Clsn1
			if clsn := c.curFrame.Clsn1; len(clsn) > 0 {
				if c.scf(SCF_standby) {
					// Add nothing
				} else if c.atktmp != 0 && c.hitdef.reversal_attr > 0 {
					sys.debugc1rev.Add(clsn, xoff, yoff, xs, ys, angle)
				} else if c.atktmp != 0 && c.hitdef.attr > 0 {
					sys.debugc1hit.Add(clsn, xoff, yoff, xs, ys, angle)
				} else {
					sys.debugc1not.Add(clsn, xoff, yoff, xs, ys, angle)
				}
			}

			// Check invincibility to decide box colors
			flags := int32(ST_SCA) | int32(AT_ALL)
			if clsn := c.curFrame.Clsn2; len(clsn) > 0 {
				hb, mtk := false, false

				if c.unhittableTime > 0 {
					mtk = true
				} else {
					for _, h := range c.hitby {
						if h.time == 0 {
							continue
						}

						// If carrying invincibility from previous iterations
						if h.stack && flags != int32(ST_SCA)|int32(AT_ALL) {
							nhbtxt = "Stacked"
							hb = true
							mtk = false
							break
						}

						// Player-specific invincibility
						if h.playerno >= 0 || h.playerid >= 0 {
							nhbtxt = "Player-specific"
							hb = true
							mtk = false
							break
						}

						// Combine flags for HitBy and NotHitBy
						if h.flag >= 0 {
							if h.not {
								// NotHitBy removes flags
								flags &= ^h.flag
							} else {
								// HitBy keeps only allowed flags
								flags &= h.flag
							}
						}
					}

					// If not stacked and not player-specific
					if nhbtxt == "" && flags != int32(ST_SCA)|int32(AT_ALL) {
						hb = true
						mtk = flags&int32(ST_SCA) == 0 || flags&int32(AT_ALL) == 0
					}
				}

				// Decide which debug box to add
				switch {
				case c.scf(SCF_standby):
					sys.debugc2stb.Add(clsn, xoff, yoff, xs, ys, angle) // Standby
				case mtk:
					sys.debugc2mtk.Add(clsn, xoff, yoff, xs, ys, angle) // Fully invincible
				case hb:
					sys.debugc2hb.Add(clsn, xoff, yoff, xs, ys, angle) // Partially invincible
				case c.inguarddist && c.scf(SCF_guard):
					sys.debugc2grd.Add(clsn, xoff, yoff, xs, ys, angle) // Guarding
				default:
					sys.debugc2.Add(clsn, xoff, yoff, xs, ys, angle) // Normal
				}

				// Add invulnerability text
				if nhbtxt == "" {
					if mtk {
						nhbtxt = "Invincible"
					} else if hb {
						// Statetype
						if flags&int32(ST_S) == 0 || flags&int32(ST_C) == 0 || flags&int32(ST_A) == 0 {
							if flags&int32(ST_S) == 0 {
								nhbtxt += "S"
							}
							if flags&int32(ST_C) == 0 {
								nhbtxt += "C"
							}
							if flags&int32(ST_A) == 0 {
								nhbtxt += "A"
							}
							nhbtxt += " Any"
						}
						// Attack
						if flags&int32(AT_NA) == 0 || flags&int32(AT_SA) == 0 || flags&int32(AT_HA) == 0 {
							if nhbtxt != "" {
								nhbtxt += ", "
							}
							if flags&int32(AT_NA) == 0 {
								nhbtxt += "N"
							}
							if flags&int32(AT_SA) == 0 {
								nhbtxt += "S"
							}
							if flags&int32(AT_HA) == 0 {
								nhbtxt += "H"
							}
							nhbtxt += " Atk"
						}
						// Throw
						if flags&int32(AT_NT) == 0 || flags&int32(AT_ST) == 0 || flags&int32(AT_HT) == 0 {
							if nhbtxt != "" {
								nhbtxt += ", "
							}
							if flags&int32(AT_NT) == 0 {
								nhbtxt += "N"
							}
							if flags&int32(AT_ST) == 0 {
								nhbtxt += "S"
							}
							if flags&int32(AT_HT) == 0 {
								nhbtxt += "H"
							}
							nhbtxt += " Thr"
						}
						// Projectile
						if flags&int32(AT_NP) == 0 || flags&int32(AT_SP) == 0 || flags&int32(AT_HP) == 0 {
							if nhbtxt != "" {
								nhbtxt += ", "
							}
							if flags&int32(AT_NP) == 0 {
								nhbtxt += "N"
							}
							if flags&int32(AT_SP) == 0 {
								nhbtxt += "S"
							}
							if flags&int32(AT_HP) == 0 {
								nhbtxt += "H"
							}
							nhbtxt += " Prj"
						}
					}
				}
			}

			// Add size box (width * height)
			if c.csf(CSF_playerpush) {
				sys.debugcsize.Add(c.sizeBoxToClsn(), x, y, c.facing*c.localscl, c.localscl, 0)
			}
		}
		// Add crosshair
		sys.debugch.Add([][4]float32{{-1, -1, 1, 1}}, x, y, 1, 1, 0)
	}
	// Prepare information for debug text
	if sys.debugDisplay {
		// Add debug clsnText
		x = (x-sys.cam.Pos[0])*sys.cam.Scale + ((320-float32(sys.gameWidth))/2 + 1) + float32(sys.gameWidth)/2
		y = (y*sys.cam.Scale - sys.cam.Pos[1]) + sys.cam.GroundLevel() + 1 // "1" is just for spacing
		y += float32(sys.debugFont.fnt.Size[1]) * sys.debugFont.yscl / sys.heightScale
		// Name and ID
		sys.clsnText = append(sys.clsnText, ClsnText{x: x, y: y, text: fmt.Sprintf("%s, %d", c.name, c.id), r: 255, g: 255, b: 255})
		// NotHitBy
		if nhbtxt != "" {
			y += float32(sys.debugFont.fnt.Size[1]) * sys.debugFont.yscl / sys.heightScale
			sys.clsnText = append(sys.clsnText, ClsnText{x: x, y: y, text: fmt.Sprintf(nhbtxt), r: 191, g: 255, b: 255})
		}
		// Targets
		for _, tid := range c.targets {
			if t := sys.playerID(tid); t != nil {
				y += float32(sys.debugFont.fnt.Size[1]) * sys.debugFont.yscl / sys.heightScale
				jg := t.ghv.getJuggle(c.id, c.gi().data.airjuggle)
				sys.clsnText = append(sys.clsnText, ClsnText{x: x, y: y, text: fmt.Sprintf("Target %d: %d", tid, jg), r: 255, g: 191, b: 255})
			}
		}
	}
}

// Prepare character sprites for drawing
func (c *Char) cueDraw() {
	if c.helperIndex < 0 || c.scf(SCF_disabled) {
		return
	}
	// Add debug info
	c.cueDebugDraw()
	// Add char sprite
	if c.anim != nil {
		pos := [2]float32{c.interPos[0]*c.localscl + c.offsetX()*c.localscl,
			c.interPos[1]*c.localscl + c.offsetY()*c.localscl}

		drawscale := [2]float32{
			c.facing * c.size.xscale * c.angleDrawScale[0] * c.zScale * (320 / c.localcoord),
			c.size.yscale * c.angleDrawScale[1] * c.zScale * (320 / c.localcoord),
		}

		// Apply Z axis perspective
		if sys.zEnabled() {
			pos = sys.drawposXYfromZ(pos, c.localscl, c.interPos[2], c.zScale)
		}

		//if sys.zEnabled() {
		//	ratio := float32(1.618) // Possible stage parameter?
		//	pos[0] *= 1 + (ratio-1)*(c.zScale-1)
		//	pos[1] *= 1 + (ratio-1)*(c.zScale-1)
		//	pos[1] += c.interPos[2] * c.localscl
		//}

		anglerot := c.anglerot
		fLength := c.fLength

		if fLength <= 0 {
			fLength = 2048
		}

		if c.facing < 0 {
			anglerot[0] *= -1
			anglerot[2] *= -1
		}
		fLength = fLength * c.localscl
		rot := c.rot

		if c.csf(CSF_angledraw) {
			rot.angle = anglerot[0]
			rot.xangle = anglerot[1]
			rot.yangle = anglerot[2]
		}

		rec := sys.tickNextFrame() && c.acttmp > 0

		//if rec {
		//	c.aimg.recAfterImg(sdf(), c.hitPause())
		//}

		// Determine AIR offset multiplier
		// This must take into account both the coordinate spaces and the scale constants
		// This seems more complicated than it ought to be. Probably because our drawing functions are different from Mugen
		// https://github.com/ikemen-engine/Ikemen-GO/issues/1459, 1778 and 2089
		airOffsetFix := [2]float32{1, 1}
		if c.playerNo != c.animPN && c.animPN >= 0 && c.animPN < len(sys.chars) && len(sys.chars[c.animPN]) > 0 {
			self := sys.chars[c.playerNo][0]
			owner := sys.chars[c.animPN][0]
			airOffsetFix = [2]float32{
				(self.localcoord / owner.localcoord) / (self.size.xscale / owner.size.xscale),
				(self.localcoord / owner.localcoord) / (self.size.yscale / owner.size.yscale),
			}
		}

		var cwin = [4]float32{
			c.window[0] * drawscale[0],
			c.window[1] * drawscale[1],
			c.window[2] * drawscale[0],
			c.window[3] * drawscale[1],
		}

		// Use animation backup if char used ChangeAnim during hitpause
		anim := c.anim
		if c.animNo >= 0 && c.anim.spr == nil && c.animBackup != nil {
			anim = c.animBackup
		}

		// Define sprite data
		sd := &SprData{
			anim:         anim,
			fx:           c.getPalfx(),
			pos:          pos,
			scl:          drawscale,
			trans:        c.trans,
			alpha:        c.alpha,
			priority:     c.sprPriority + int32(c.pos[2]*c.localscl),
			rot:          rot,
			screen:       false,
			undarken:     c.ignoreDarkenTime > 0,
			oldVer:       c.gi().mugenver[0] != 1,
			facing:       c.facing,
			airOffsetFix: airOffsetFix,
			projection:   int32(c.projection),
			fLength:      fLength,
			xshear:       c.xshear,
			window:       cwin,
		}

		// Record afterimage
		c.aimg.recAndCue(sd, rec, sys.tickNextFrame() && c.hitPause(), c.layerNo)
		// Hitshake effect
		if c.ghv.hitshaketime > 0 && c.ss.time&1 != 0 {
			sd.pos[0] -= c.facing
		}
		// Draw char according to layer number
		sprs := &sys.spritesLayer0
		if c.layerNo > 0 {
			sprs = &sys.spritesLayer1
		} else if c.layerNo < 0 {
			sprs = &sys.spritesLayerN1
		} else if c.asf(ASF_drawunder) {
			sprs = &sys.spritesLayerU
		}

		if !c.asf(ASF_invisible) {
			sdwalp := 255 - c.alpha[1]
			sdwclr := c.shadowColor[0]<<16 | c.shadowColor[1]<<8 | c.shadowColor[2]
			reflectclr := c.reflectColor[0]<<16 | c.reflectColor[1]<<8 | c.reflectColor[2]

			// Add sprite to draw list
			sprs.add(sd)

			// Add shadow
			if !c.asf(ASF_noshadow) {
				// Previously Ikemen applied a multiplier of 1.5 to c.size.shadowoffset for Winmugen chars
				// That doesn't seem to actually happen in either Winmugen or Mugen 1.1
				//soy := c.size.shadowoffset
				//if sd.oldVer {
				//	soy *= 1.5
				//}
				// Mugen uses some odd math for the shadow offset here, factoring in the stage's shadow scale
				// Meaning the character's shadow offset constant is unable to offset it correctly in every stage
				// Ikemen works differently and as you'd expect it to
				drawZoff := sys.posZtoYoffset(c.interPos[2], c.localscl)
				// Gets the Yscale defined by ModifyShadow/Reflection or keeps the one from the stage
				getYscale := func(char, stage float32) float32 {
					if char != 0 {
						return char
					}
					return stage
				}
				sdwYscale := getYscale(c.shadowYscale, sys.stage.sdw.yscale)
				refYscale := getYscale(c.reflectYscale, sys.stage.reflection.yscale)

				// Add shadow to shadow list
				sys.shadows.add(&ShadowSprite{
					SprData:         sd,
					shadowColor:     sdwclr,
					shadowAlpha:     sdwalp,
					shadowIntensity: c.shadowIntensity,
					shadowOffset: [2]float32{
						c.shadowOffset[0] * c.localscl,
						(c.size.shadowoffset+c.shadowOffset[1])*c.localscl + sdwYscale*drawZoff + drawZoff,
					},
					shadowWindow:     c.shadowWindow,
					shadowXshear:     c.shadowXshear,
					shadowYscale:     c.shadowYscale,
					shadowRot:        c.shadowRot,
					shadowProjection: int32(c.shadowProjection),
					shadowfLength:    c.shadowfLength,
					fadeOffset:       c.offsetY() + drawZoff,
				})

				// Add reflection to reflection list
				sys.reflections.add(&ReflectionSprite{
					SprData:          sd,
					reflectColor:     reflectclr,
					reflectIntensity: c.reflectIntensity,
					reflectOffset: [2]float32{
						c.reflectOffset[0] * c.localscl,
						(c.size.shadowoffset+c.reflectOffset[1])*c.localscl + refYscale*drawZoff + drawZoff,
					},
					reflectWindow:     c.reflectWindow,
					reflectXshear:     c.reflectXshear,
					reflectYscale:     c.reflectYscale,
					reflectRot:        c.reflectRot,
					reflectProjection: int32(c.reflectProjection),
					reflectfLength:    c.reflectfLength,
					fadeOffset:        c.offsetY() + drawZoff,
				})
			}
		}
	}
	if sys.tickNextFrame() {
		c.minus = 2
		c.oldPos = c.pos
		c.dustOldPos = c.pos // We need this one separated because PosAdd and such change oldPos
	}
}

type CharList struct {
	runOrder         []*Char
	idMap            map[int32]*Char
	enemyNearChanged bool
}

func (cl *CharList) clear() {
	*cl = CharList{idMap: make(map[int32]*Char)}
	sys.nextCharId = sys.cfg.Config.HelperMax
}

func (cl *CharList) add(c *Char) {
	// Append to run order
	cl.runOrder = append(cl.runOrder, c)

	// Update char ID map for fast lookup
	cl.idMap[c.id] = c
}

func (cl *CharList) replace(dc *Char, pn int, idx int32) bool {
	var ok bool

	// Replace in runOrder
	for i, c := range cl.runOrder {
		if c.playerNo == pn && c.helperIndex == idx {
			cl.runOrder[i] = dc
			ok = true
			break
		}
	}

	if ok {
		// Update ID map
		cl.idMap[dc.id] = dc
	}

	return ok
}

func (cl *CharList) delete(dc *Char) {
	for i, c := range cl.runOrder {
		if c == dc {
			delete(cl.idMap, c.id)
			cl.runOrder = append(cl.runOrder[:i], cl.runOrder[i+1:]...)
			break
		}
	}
	// Mugen and older versions of Ikemen could reuse the drawing order of an old removed helper for a new helper
	// However not reusing it creates a more predictable drawing order
}

func (cl *CharList) commandUpdate() {
	// Iterate players
	for i, p := range sys.chars {
		if len(p) > 0 {
			root := p[0]
			// Select a random command for AI cheating
			// The way this only allows one command to be cheated at a time may be the cause of issue #2022
			cheat := int32(-1)
			if root.controller < 0 {
				if sys.roundState() == 2 && RandF32(0, sys.aiLevel[i]/2+32) > 32 { // TODO: Balance AI scaling
					cheat = Rand(0, int32(len(root.cmd[root.ss.sb.playerNo].Commands))-1)
				}
			}
			// Iterate root and helpers
			for _, c := range p {
				act := true
				if sys.supertime > 0 {
					act = c.superMovetime != 0
				} else if sys.pausetime > 0 && c.pauseMovetime == 0 {
					act = false
				}
				// Auto turning check for the root
				// Having this here makes B and F inputs reverse the same instant the character turns
				if act && c.helperIndex == 0 && (c.scf(SCF_ctrl) || sys.roundState() > 2) &&
					(c.ss.no == 0 || c.ss.no == 11 || c.ss.no == 20 || c.ss.no == 52) {
					c.autoTurn()
				}

				// Update Forward/Back flipping flag
				c.updateFBFlip()

				if (c.helperIndex == 0 || c.helperIndex > 0 && &c.cmd[0] != &root.cmd[0]) &&
					c.cmd[0].InputUpdate(c, c.controller, sys.aiLevel[i], false) {
					// Clear input buffers and skip the rest of the loop
					// This used to apply only to the root, but that caused some issues with helper-based custom input systems
					if c.inputWait() || c.asf(ASF_noinput) {
						for i := range c.cmd {
							c.cmd[i].BufReset()
						}
						continue
					}
					hpbuf := false
					pausebuf := false
					winbuf := false
					// Buffer during hitpause
					if c.hitPause() && c.gi().constants["input.pauseonhitpause"] != 0 { // TODO: Deprecated constant
						hpbuf = true
						// In Winmugen, commands were buffered for one extra frame after hitpause (but not after Pause/SuperPause)
						// This was fixed in Mugen 1.0
						if c.stWgi().ikemenver[0] == 0 && c.stWgi().ikemenver[1] == 0 && c.stWgi().mugenver[0] != 1 {
							winbuf = true
						}
					}
					// Buffer during Pause and SuperPause
					if sys.supertime > 0 {
						if !act && sys.supertime <= sys.superendcmdbuftime {
							pausebuf = true
						}
					} else if sys.pausetime > 0 {
						if !act && sys.pausetime <= sys.pauseendcmdbuftime {
							pausebuf = true
						}
					}
					// Update commands
					for i := range c.cmd {
						extratime := Btoi(hpbuf || pausebuf) + Btoi(winbuf)
						helperbug := c.helperIndex != 0 && c.stWgi().ikemenver[0] == 0 && c.stWgi().ikemenver[1] == 0
						c.cmd[i].Step(c.controller < 0, helperbug, hpbuf, pausebuf, extratime)
					}
					// Enable AI cheated command
					c.cpucmd = cheat
				}
			}
		}
	}
}

// Sort all characters into a list based on their processing order
func (cl *CharList) sortActionRunOrder() []int {
	// Temp sorting list
	sorting := make([][2]int, len(cl.runOrder)) // [2]int{index, priority}

	// Decide priority of each player
	for i, c := range cl.runOrder {
		var pr int                                      // Fallback priority of 0
		if c.asf(ASF_runfirst) && !c.asf(ASF_runlast) { // Any character with runfirst flag
			pr = 100
		} else if c.asf(ASF_runlast) && !c.asf(ASF_runfirst) { // Any character with runlast flag
			pr = -100
		} else if c.ss.moveType == MT_A { // Attacking players and helpers
			pr = 5
		} else if c.helperIndex == 0 {
			if c.ss.moveType == MT_I { // Idle players
				pr = 4
			} else { // Remaining players
				pr = 3
			}
		} else {
			if c.ss.moveType == MT_I { // Idle helpers
				pr = 2
			} else { // Remaining helpers
				pr = 1
			}
		}
		sorting[i] = [2]int{i, pr}
	}

	// Sort by priority
	sort.SliceStable(sorting, func(i, j int) bool {
		return sorting[i][1] > sorting[j][1]
	})

	// Create new sorted list and update each char's runOrder
	sortedOrder := make([]int, len(sorting))
	for i := 0; i < len(sorting); i++ {
		sortedOrder[i] = sorting[i][0]
		cl.runOrder[sorting[i][0]].runorder = int32(i + 1)
	}

	// Reset priority flags as they are only needed during this function
	for i := range cl.runOrder {
		cl.runOrder[i].unsetASF(ASF_runfirst | ASF_runlast)
	}

	return sortedOrder
}

func (cl *CharList) action() {
	// Update commands for all chars
	cl.commandUpdate()

	// Prepare characters before performing their actions
	for i := 0; i < len(cl.runOrder); i++ {
		cl.runOrder[i].actionPrepare()
	}

	// Run actions for each character in the sorted list
	// Sorting the characters first makes new helpers wait for their turn and allows RunOrder trigger accuracy
	sortedOrder := cl.sortActionRunOrder()
	for i := 0; i < len(sortedOrder); i++ {
		if sortedOrder[i] < len(cl.runOrder) {
			cl.runOrder[sortedOrder[i]].actionRun()
		}
	}

	// Run actions for anyone missed (new helpers)
	extra := len(sortedOrder) + 1
	for i := 0; i < len(cl.runOrder); i++ {
		if cl.runOrder[i].runorder < 0 {
			cl.runOrder[i].runorder = int32(extra)
			cl.runOrder[i].actionRun()
			extra++
		}
	}

	// Finish performing character actions
	for i := 0; i < len(cl.runOrder); i++ {
		cl.runOrder[i].actionFinish()
	}
}

func (cl *CharList) xScreenBound() {
	ro := make([]*Char, len(cl.runOrder))
	copy(ro, cl.runOrder)
	for _, c := range ro {
		c.xScreenBound()
	}
}

// This function runs every tick
func (cl *CharList) update() {
	ro := make([]*Char, len(cl.runOrder))
	copy(ro, cl.runOrder)
	for _, c := range ro {
		c.update()
		c.track()
	}
}

// Check player vs player hits
func (cl *CharList) hitDetectionPlayer(getter *Char) {

	// Stop outer loop if enemy is disabled
	if getter.scf(SCF_standby) || getter.scf(SCF_disabled) {
		return
	}

	getter.unsetCSF(CSF_gethit)

	// This forces an enemy list cache reset every frame
	// Has a perfomance impact and is probably not necessary in the current state of the code
	//getter.enemyNearP2Clear()

	for _, c := range cl.runOrder {

		// Stop current iteration if this char is disabled
		if c.scf(SCF_standby) || c.scf(SCF_disabled) {
			continue
		}

		if c.atktmp != 0 && c.id != getter.id && (c.hitdef.affectteam == 0 ||
			((getter.teamside != c.hitdef.teamside-1) == (c.hitdef.affectteam > 0) && c.hitdef.teamside >= 0) ||
			((getter.teamside != c.teamside) == (c.hitdef.affectteam > 0) && c.hitdef.teamside < 0)) {

			// Guard distance check
			// Mugen uses < checks so that 0 does not trigger proximity guard at 0 distance
			// Localcoord conversion is already built into the dist functions, so it will be skipped
			if c.ss.moveType == MT_A {
				var inguardx, inguardy, inguardz bool

				// Get distances
				distX := c.distX(getter, c) * c.facing
				distY := c.distY(getter, c)
				distZ := c.distZ(getter, c)

				// Check X distance
				inguardx = distX < c.hitdef.guard_dist_x[0] && distX > -c.hitdef.guard_dist_x[1]

				// Check Y distance
				if distY == 0 { // Compatibility safeguard
					inguardy = true
				} else {
					inguardy = distY > -c.hitdef.guard_dist_y[0] && distY < c.hitdef.guard_dist_y[1]
				}

				// Check Z distance
				if distZ == 0 { // Compatibility safeguard
					inguardz = true
				} else {
					inguardz = distZ > -c.hitdef.guard_dist_z[0] && distZ < c.hitdef.guard_dist_z[1]
				}

				// Set flag
				if inguardx && inguardy && inguardz {
					getter.inguarddist = true
				}
			}

			if c.helperIndex != 0 {
				// Inherit parent's or root's juggle points
				if c.inheritJuggle == 1 && c.parent(false) != nil {
					for _, v := range getter.ghv.targetedBy {
						if v[0] == c.parent(false).id {
							getter.ghv.addId(c.id, v[1])
							break
						}
					}
				} else if c.inheritJuggle == 2 && c.root(false) != nil {
					for _, v := range getter.ghv.targetedBy {
						if v[0] == c.root(false).id {
							getter.ghv.addId(c.id, v[1])
							break
						}
					}
				}
			}

			// In Mugen, you can no longer hit a standing target if you don't have enough points
			// In Mugen, you can juggle any enemy if they're not your target yet
			// If IkemenVersion, the rules are a little more consistent
			canjuggle := false
			if c.asf(ASF_nojugglecheck) ||
				c.juggle <= getter.ghv.getJuggle(c.id, c.gi().data.airjuggle) ||
				(c.gi().ikemenver[0] != 0 || c.gi().ikemenver[1] != 0) && getter.hittmp < 2 ||
				(c.gi().ikemenver[0] == 0 && c.gi().ikemenver[1] == 0 && !c.hasTarget(getter.id)) {
				canjuggle = true
			}

			// If getter can be hit by this Hitdef
			if canjuggle && c.hitdef.hitonce >= 0 && !c.hasTargetOfHitdef(getter.id) &&
				(c.hitdef.reversal_attr <= 0 || !getter.hasTargetOfHitdef(c.id)) &&
				getter.hittableByChar(c, &c.hitdef, c.ss.stateType, false) {

				// Z axis check
				// ReversalDef checks attack depth vs attack depth
				zok := true
				if c.hitdef.reversal_attr > 0 {
					zok = sys.zAxisOverlap(c.pos[2], c.hitdef.attack_depth[0], c.hitdef.attack_depth[1], c.localscl,
						getter.pos[2], getter.hitdef.attack_depth[0], getter.hitdef.attack_depth[1], getter.localscl)
				} else {
					zok = sys.zAxisOverlap(c.pos[2], c.hitdef.attack_depth[0], c.hitdef.attack_depth[1], c.localscl,
						getter.pos[2], getter.sizeDepth[0], getter.sizeDepth[1], getter.localscl)
				}

				// If collision OK then get the hit type and act accordingly
				if zok && c.clsnCheck(getter, 1, c.hitdef.p2clsncheck, true, false) {
					if hitResult := c.hitResultCheck(getter, nil); hitResult != 0 {
						// Check if MoveContact should be updated
						// Hit type None should also set MoveHit here
						mvc := hitResult >= -1 || c.hitdef.reversal_attr > 0

						// Attacker hitpauses were off by 1 frame in WinMugen. Mugen 1.0 fixed it
						// The way this should actually happen is that WinMugen chars have 1 subtracted from their hitpause in bytecode.go
						// But because of the order that events happen in in Ikemen, it must be fixed the other way around
						hpfix := c.gi().ikemenver[0] != 0 || c.gi().ikemenver[1] != 0 || c.gi().mugenver[0] == 1

						if Abs(hitResult) == 1 {
							if mvc {
								c.mctype = MC_Hit
								c.mctime = -1
							}
							// Successful ReversalDef
							if c.hitdef.reversal_attr > 0 {
								c.powerAdd(c.hitdef.hitgetpower)

								// Precompute localcoord conversion factor
								scaleratio := c.localscl / getter.localscl

								// ReversalDef seems to set an arbitrary collection of get hit variables in Mugen
								getter.hitdef.hitflag = 0
								getter.mctype = MC_Reversed
								getter.mctime = -1
								getter.hitdefContact = true
								getter.mhv.frame = true
								getter.mhv.playerId = c.id
								getter.mhv.playerNo = c.playerNo
								getter.hitdef.hitonce = -1 // Neutralize Hitdef

								if c.hitdef.unhittabletime[1] >= 0 {
									getter.unhittableTime = c.hitdef.unhittabletime[1] // 1
								}

								// Clear GetHitVars while stacking those that need it
								getter.ghv.selectiveClear(getter)

								getter.ghv.attr = c.hitdef.attr
								getter.ghv.hitid = c.hitdef.id
								getter.ghv.playerNo = c.playerNo
								getter.ghv.playerId = c.id
								getter.fallTime = 0

								// Fall flag
								if c.hitdef.forcenofall {
									getter.ghv.fallflag = false
								} else if !getter.ghv.fallflag {
									if getter.ss.stateType == ST_A {
										getter.ghv.fallflag = c.hitdef.air_fall != 0
									} else {
										getter.ghv.fallflag = c.hitdef.ground_fall
									}
								}

								// Fall group
								getter.ghv.fall_animtype = c.hitdef.fall_animtype
								getter.ghv.fall_xvelocity = c.hitdef.fall_xvelocity * scaleratio
								getter.ghv.fall_yvelocity = c.hitdef.fall_yvelocity * scaleratio
								getter.ghv.fall_zvelocity = c.hitdef.fall_zvelocity * scaleratio
								getter.ghv.fall_recover = c.hitdef.fall_recover
								getter.ghv.fall_recovertime = c.hitdef.fall_recovertime
								getter.ghv.fall_damage = c.hitdef.fall_damage
								getter.ghv.fall_kill = c.hitdef.fall_kill
								getter.ghv.fall_envshake_time = c.hitdef.fall_envshake_time
								getter.ghv.fall_envshake_freq = c.hitdef.fall_envshake_freq
								getter.ghv.fall_envshake_ampl = int32(float32(c.hitdef.fall_envshake_ampl) * scaleratio)
								getter.ghv.fall_envshake_phase = c.hitdef.fall_envshake_phase
								getter.ghv.fall_envshake_mul = c.hitdef.fall_envshake_mul
								getter.ghv.fall_envshake_dir = c.hitdef.fall_envshake_dir

								getter.ghv.down_recover = c.hitdef.down_recover
								if c.hitdef.down_recovertime < 0 {
									getter.ghv.down_recovertime = getter.gi().data.liedown.time
								} else {
									getter.ghv.down_recovertime = c.hitdef.down_recovertime
								}

								getter.hitdefTargetsBuffer = append(getter.hitdefTargetsBuffer, c.id)
								if getter.hittmp == 0 {
									getter.hittmp = -1
								}
								if !getter.csf(CSF_gethit) {
									getter.hitPauseTime = Max(1, c.hitdef.pausetime[1]+Btoi(hpfix))
								}
							}
							if !c.csf(CSF_gethit) && (getter.ss.stateType == ST_A && c.hitdef.air_type != HT_None ||
								getter.ss.stateType != ST_A && c.hitdef.ground_type != HT_None) {
								c.hitPauseTime = Max(1, c.hitdef.pausetime[0]+Btoi(hpfix))
								// In Mugen, the hitpause only actually takes effect in the next frame
								// In Mugen, despite hit type None being supposed to apply hitpause, that doesn't happen
								// Curiously, if a HitOverride is used the hitpause will be restored
							}
							c.uniqHitCount++
						} else {
							if mvc {
								c.mctype = MC_Guarded
								c.mctime = -1
							}
							if !c.csf(CSF_gethit) {
								c.hitPauseTime = Max(1, c.hitdef.guard_pausetime[0]+Btoi(hpfix))
							}
						}
						if c.hitdef.hitonce > 0 {
							c.hitdef.hitonce = -1
						}
						c.hitdefContact = true
						c.mhv.frame = true
						c.mhv.playerId = getter.id
						c.mhv.playerNo = getter.playerNo
						if c.hitdef.unhittabletime[0] >= 0 {
							c.unhittableTime = c.hitdef.unhittabletime[0]
						}
					}
				}
			}
		}
	}
}

// Check projectile vs player hits
func (cl *CharList) hitDetectionProjectile(getter *Char) {

	// Stop outer loop if enemy is disabled
	if getter.scf(SCF_standby) || getter.scf(SCF_disabled) {
		return
	}

	for i := range sys.projs {
		// Skip if this player number has no projectiles
		if len(sys.projs[i]) == 0 {
			continue
		}

		c := sys.chars[i][0]
		ap_projhit := false

		// Save root's atktmp var so we can temporarily modify it
		// Maybe this is no longer necessary
		//orgatktmp := c.atktmp
		//c.atktmp = -1

		for j := range sys.projs[i] {
			p := sys.projs[i][j]

			// Skip if projectile can't hit
			if p.id < 0 || p.hits <= 0 {
				continue
			}

			// In Mugen, projectiles couldn't hit their root even with the proper affectteam
			if i == getter.playerNo && getter.helperIndex == 0 &&
				(getter.teamside == p.hitdef.teamside-1) && !p.platform {
				continue
			}

			// Teamside check
			// Since the teamside parameter is new to Ikemen, we can make that one allow the projectile to hit the root
			if p.hitdef.affectteam != 0 &&
				((getter.teamside != p.hitdef.teamside-1) != (p.hitdef.affectteam > 0) ||
					(getter.teamside == p.hitdef.teamside-1) != (p.hitdef.affectteam < 0)) {
				continue
			}

			// Projectile guard distance check
			distX := (getter.pos[0]*getter.localscl - (p.pos[0])*p.localscl) * p.facing
			distY := (getter.pos[1]*getter.localscl - (p.pos[1])*p.localscl)
			distZ := (getter.pos[2]*getter.localscl - (p.pos[2])*p.localscl)

			if !p.platform && p.hitdef.attr > 0 { // https://github.com/ikemen-engine/Ikemen-GO/issues/1445
				var inguardx, inguardy, inguardz bool

				// Check X distance
				inguardx = distX < p.hitdef.guard_dist_x[0]*p.localscl &&
					distX > -p.hitdef.guard_dist_x[1]*p.localscl

				// Check Y distance
				if distY == 0 { // Compatibility safeguard
					inguardy = true
				} else {
					inguardy = distY > -p.hitdef.guard_dist_y[0]*p.localscl &&
						distY < p.hitdef.guard_dist_y[1]*p.localscl
				}

				// Check Z distance
				if distZ == 0 { // Compatibility safeguard
					inguardz = true
				} else {
					inguardz = distZ > -p.hitdef.guard_dist_z[0]*p.localscl &&
						distZ < p.hitdef.guard_dist_z[1]*p.localscl
				}

				// Set flag
				if inguardx && inguardy && inguardz {
					getter.inguarddist = true
				}
			}

			if p.platform {
				// Check if the character is above the platform's surface
				if getter.pos[1]*getter.localscl-getter.vel[1]*getter.localscl <= (p.pos[1]+p.platformHeight[1])*p.localscl &&
					getter.platformPosY*getter.localscl >= (p.pos[1]+p.platformHeight[0])*p.localscl {
					angleSinValue := float32(math.Sin(float64(p.platformAngle) / 180 * math.Pi))
					angleCosValue := float32(math.Cos(float64(p.platformAngle) / 180 * math.Pi))
					oldDistX := (getter.oldPos[0]*getter.localscl - (p.pos[0])*p.localscl) * p.facing
					onPlatform := func(protrude bool) {
						getter.platformPosY = ((p.pos[1]+p.platformHeight[0]+p.velocity[1])*p.localscl - angleSinValue*(oldDistX/angleCosValue)) / getter.localscl
						getter.groundAngle = p.platformAngle
						// Condition when the character is on the platform
						if getter.ss.stateType != ST_A {
							getter.pos[0] += p.velocity[0] * p.facing * (p.localscl / getter.localscl)
							getter.pos[1] += p.velocity[1] * (p.localscl / getter.localscl)
							if protrude {
								if p.facing > 0 {
									getter.xPlatformBound((p.pos[0]+p.velocity[0]*2*p.facing+p.platformWidth[0]*angleCosValue*p.facing)*p.localscl, (p.pos[0]-p.velocity[0]*2*p.facing+p.platformWidth[1]*angleCosValue*p.facing)*p.localscl)
								} else {
									getter.xPlatformBound((p.pos[0]-p.velocity[0]*2*p.facing+p.platformWidth[1]*angleCosValue*p.facing)*p.localscl, (p.pos[0]+p.velocity[0]*2*p.facing+p.platformWidth[0]*angleCosValue*p.facing)*p.localscl)
								}
							}
						}
					}
					if distX >= (p.platformWidth[0]*angleCosValue)*p.localscl && distX <= (p.platformWidth[1]*angleCosValue)*p.localscl {
						onPlatform(false)
					} else if p.platformFence && oldDistX >= (p.platformWidth[0]*angleCosValue)*p.localscl &&
						oldDistX <= (p.platformWidth[1]*angleCosValue)*p.localscl {
						onPlatform(true)
					}
				}
			}

			// Cancel a projectile with hitflag P
			if getter.atktmp != 0 && (getter.hitdef.affectteam == 0 ||
				(p.hitdef.teamside-1 != getter.teamside) == (getter.hitdef.affectteam > 0)) &&
				getter.hitdef.hitflag&int32(HF_P) != 0 &&
				getter.projClsnCheck(p, 1, 2) &&
				sys.zAxisOverlap(getter.pos[2], getter.hitdef.attack_depth[0], getter.hitdef.attack_depth[1], getter.localscl,
					p.pos[2], p.hitdef.attack_depth[0], p.hitdef.attack_depth[1], p.localscl) {
				if getter.hitdef.p1stateno >= 0 && getter.stateChange1(getter.hitdef.p1stateno, getter.hitdef.playerNo) {
					getter.setCtrl(false)
				}
				p.flagProjCancel()
				getter.hitdefContact = true
				//getter.mhv.frame = true // Doesn't make sense to flag it when cancelling a projectile
				continue
			}

			// Projectile juggling is a little different from player juggling
			// In Mugen, they check juggle points even if the enemy is not yet a target or even falling at all
			// IkemenVersion once again makes the logic more consistent
			canjuggle := false
			if c.asf(ASF_nojugglecheck) ||
				(c.gi().ikemenver[0] != 0 || c.gi().ikemenver[1] != 0) && getter.hittmp < 2 ||
				p.hitdef.air_juggle <= getter.ghv.getJuggle(c.id, c.gi().data.airjuggle) {
				canjuggle = true
			}

			if canjuggle && !(getter.stchtmp && (getter.csf(CSF_gethit) || getter.acttmp > 0)) &&
				(!ap_projhit || p.hitdef.attr&int32(AT_AP) == 0) &&
				(p.hitpause <= 0 || p.contactflag) && p.curmisstime <= 0 && p.hitdef.hitonce >= 0 &&
				getter.hittableByChar(c, &p.hitdef, ST_N, true) {

				// Save enemy's hittmp var so we can temporarily modify it
				// Maybe this is no longer necessary
				//orghittmp := getter.hittmp
				//if getter.csf(CSF_gethit) {
				//	getter.hittmp = int8(Btoi(getter.ghv.fallflag)) + 1
				//}

				if getter.projClsnCheck(p, p.hitdef.p2clsncheck, 1) &&
					sys.zAxisOverlap(p.pos[2], p.hitdef.attack_depth[0], p.hitdef.attack_depth[1], p.localscl,
						getter.pos[2], getter.sizeDepth[0], getter.sizeDepth[1], getter.localscl) {

					if hitResult := c.hitResultCheck(getter, p); hitResult != 0 {

						p.contactflag = true
						if Abs(hitResult) == 1 {
							sys.cgi[i].pctype = PC_Hit
							p.hitpause = Max(0, p.hitdef.pausetime[0]-Btoi(c.gi().mugenver[0] == 0)) // Winmugen projectiles are 1 frame short on hitpauses
						} else {
							sys.cgi[i].pctype = PC_Guarded
							p.hitpause = Max(0, p.hitdef.guard_pausetime[0]-Btoi(c.gi().mugenver[0] == 0))
						}
						sys.cgi[i].pctime = 0
						sys.cgi[i].pcid = p.id
					}
					// This flag prevents multiple projectiles from the same player from hitting in the same frame
					// In Mugen, projectiles (sctrl) give 1F of projectile invincibility to the getter instead. This timer persists during (super)pause
					if p.hitdef.attr&int32(AT_AP) != 0 {
						ap_projhit = true
					}
				}
				// Restore enemy's hittmp var
				//getter.hittmp = orghittmp
			}
		}

		// Restore root's atktmp var
		//c.atktmp = orgatktmp
	}
}

func (cl *CharList) pushDetection(getter *Char) {
	var gxmin, gxmax float32

	// Stop outer loop if getter won't push
	if !getter.csf(CSF_playerpush) || getter.scf(SCF_standby) || getter.scf(SCF_disabled) {
		return
	}

	for _, c := range cl.runOrder {
		// Stop current iteration if char won't push
		if !c.csf(CSF_playerpush) || c.teamside == getter.teamside || c.scf(SCF_standby) || c.scf(SCF_disabled) {
			continue
		}

		// Pushbox vertical size and coordinates
		cytop := (c.pos[1] + c.sizeBox[1]) * c.localscl
		cybot := (c.pos[1] + c.sizeBox[3]) * c.localscl
		gytop := (getter.pos[1] + getter.sizeBox[1]) * getter.localscl
		gybot := (getter.pos[1] + getter.sizeBox[3]) * getter.localscl

		if cybot >= gytop && cytop <= gybot { // Pushbox vertical overlap

			// We skip the zAxisCheck function because we'll need to calculate the overlap again anyway

			// Normal collision check
			cposx := c.pos[0] * c.localscl
			cxleft := c.sizeBox[0] * c.localscl
			cxright := c.sizeBox[2] * c.localscl
			if c.facing < 0 {
				cxleft, cxright = -cxright, -cxleft
			}

			cxleft += cposx
			cxright += cposx

			gposx := getter.pos[0] * getter.localscl
			gxleft := getter.sizeBox[0] * getter.localscl
			gxright := getter.sizeBox[2] * getter.localscl
			if getter.facing < 0 {
				gxleft, gxright = -gxright, -gxleft
			}

			gxleft += gposx
			gxright += gposx

			// X axis fail
			if gxleft >= cxright || cxleft >= gxright {
				continue
			}

			cposz := c.pos[2] * c.localscl
			cztop := cposz - c.sizeDepth[0]*c.localscl
			czbot := cposz + c.sizeDepth[1]*c.localscl

			gposz := getter.pos[2] * getter.localscl
			gztop := gposz - getter.sizeDepth[0]*getter.localscl
			gzbot := gposz + getter.sizeDepth[1]*getter.localscl

			// Z axis fail
			if gztop >= czbot || cztop >= gzbot {
				continue
			}

			// Push characters away from each other
			if c.asf(ASF_sizepushonly) || getter.clsnCheck(c, 2, 2, false, false) {

				getter.pushed, c.pushed = true, true

				gxmin = getter.edgeWidth[0]
				gxmax = -getter.edgeWidth[1]
				if getter.facing > 0 {
					gxmin, gxmax = -gxmax, -gxmin
				}
				gxmin += sys.xmin / getter.localscl
				gxmax += sys.xmax / getter.localscl

				// Decide who gets pushed
				cpushed := float32(0.5)
				gpushed := float32(0.5)
				if c.pushPriority > getter.pushPriority {
					cpushed = 0
					gpushed = 1
				} else if c.pushPriority < getter.pushPriority {
					cpushed = 1
					gpushed = 0
				}

				// Compare player weights and apply pushing factors
				// Weight determines which player is pushed more. Factor determines how fast the player overlap is resolved
				cfactor := float32(getter.size.weight) / float32(c.size.weight+getter.size.weight) * c.size.pushfactor * cpushed
				gfactor := float32(c.size.weight) / float32(c.size.weight+getter.size.weight) * getter.size.pushfactor * gpushed

				// Determine in which axes to push the players
				// This needs to check both if the players have velocity or if their positions have changed
				var pushx, pushz bool
				if sys.zEnabled() && gposz != cposz { // If tied on Z axis we fall back to X pushing
					// Get distances in both axes
					distx := AbsF(gposx - cposx)
					distz := AbsF(gposz - cposz)

					// Check how much each axis should weigh on the decision
					// Adjust z-distance to same scale as x-distance, since character depths are usually smaller than widths
					xtotal := AbsF(gxleft-gxright) + AbsF(cxleft-cxright)
					ztotal := AbsF(gztop-gzbot) + AbsF(cztop-czbot)
					distzadj := distz
					if ztotal != 0 {
						distzadj = (xtotal / ztotal) * distz
					}

					// Push farthest axis or both if distances are similar
					similar := float32(0.75) // Ratio at which distances are considered similar. Arbitrary number. Maybe there's a better way
					if distzadj != 0 && AbsF(distx/distzadj) > similar && AbsF(distx/distzadj) < (1/similar) {
						pushx = true
						pushz = true
					} else if distx >= distzadj {
						pushx = true
					} else {
						pushz = true
					}
				} else {
					pushx = true
				}

				if pushx {
					tmp := getter.distX(c, getter)
					if tmp == 0 {
						// Decide direction in which to push each player in case of a tie in position
						// This also decides who gets to stay in the corner
						// Some of these checks are similar to char run order, but this approach allows better tie break control
						// https://github.com/ikemen-engine/Ikemen-GO/issues/1426
						if c.pushPriority > getter.pushPriority {
							if c.pos[0] >= 0 {
								tmp = 1
							} else {
								tmp = -1
							}
						} else if c.pushPriority < getter.pushPriority {
							if getter.pos[0] >= 0 {
								tmp = -1
							} else {
								tmp = 1
							}
						} else if c.ss.moveType == MT_H && getter.ss.moveType != MT_H {
							tmp = -c.facing
						} else if c.ss.moveType != MT_H && getter.ss.moveType == MT_H {
							tmp = getter.facing
						} else if c.ss.moveType == MT_A && getter.ss.moveType != MT_A {
							tmp = getter.facing
						} else if c.ss.moveType != MT_A && getter.ss.moveType == MT_A {
							tmp = -c.facing
						} else if c.pos[1]*c.localscl < getter.pos[1]*getter.localscl {
							tmp = getter.facing
						} else {
							tmp = -c.facing
						}
					}
					if tmp > 0 {
						if c.pushPriority >= getter.pushPriority {
							getter.pos[0] -= ((gxright - cxleft) * gfactor) / getter.localscl
						}
						if c.pushPriority <= getter.pushPriority {
							c.pos[0] += ((gxright - cxleft) * cfactor) / c.localscl
						}
					} else {
						if c.pushPriority >= getter.pushPriority {
							getter.pos[0] += ((cxright - gxleft) * gfactor) / getter.localscl
						}
						if c.pushPriority <= getter.pushPriority {
							c.pos[0] -= ((cxright - gxleft) * cfactor) / c.localscl
						}
					}
					// Clamp X positions
					c.xScreenBound()
					getter.xScreenBound()
				}

				// TODO: Z axis push might need some decision for who stays in the corner, like X axis
				if pushz {
					if gposz < cposz {
						if c.pushPriority >= getter.pushPriority {
							getter.pos[2] -= ((gzbot - cztop) * gfactor) / getter.localscl
						}
						if c.pushPriority <= getter.pushPriority {
							c.pos[2] += ((gzbot - cztop) * cfactor) / c.localscl
						}
					} else if gposz > cposz {
						if c.pushPriority >= getter.pushPriority {
							getter.pos[2] += ((czbot - gztop) * gfactor) / getter.localscl
						}
						if c.pushPriority <= getter.pushPriority {
							c.pos[2] -= ((czbot - gztop) * cfactor) / c.localscl
						}
					}
					// Clamp Z positions
					c.zDepthBound()
					getter.zDepthBound()
				}
			}
		}
	}
}

func (cl *CharList) collisionDetection() {
	// Temp sorting list
	sorting := make([][2]int, len(cl.runOrder)) // [2]int{index, priority}

	// Decide priority of each player
	// TODO: Maybe this could also be affected by runfirst/runlast
	for i, c := range cl.runOrder {
		var pr int
		if c.hitdef.reversal_attr > 0 { // ReversalDef first
			pr = 2
		} else if c.hitdef.attr > 0 { // Then HitDef
			pr = 1
		} else { // Everyone else
			pr = 0
		}
		sorting[i] = [2]int{i, pr}
	}

	// Sort by priority
	sort.SliceStable(sorting, func(i, j int) bool {
		return sorting[i][1] > sorting[j][1]
	})

	// Create the new sorted list
	sortedOrder := make([]int, len(sorting))
	for i := 0; i < len(sorting); i++ {
		sortedOrder[i] = sorting[i][0]
	}

	// Push detection for players
	// This must happen before hit detection
	// https://github.com/ikemen-engine/Ikemen-GO/issues/1941
	// It doesn't need to run in "sortedOrder", but it should be harmless
	// An attempt was made to skip redundant player pair checks, but that makes chars push each other too slowly in screen corners
	for _, idx := range sortedOrder {
		cl.pushDetection(cl.runOrder[idx])
	}

	// Player hit detection
	for _, idx := range sortedOrder {
		cl.hitDetectionPlayer(cl.runOrder[idx])
	}

	// Projectile hit detection
	for _, c := range cl.runOrder {
		cl.hitDetectionProjectile(c)
	}
}

func (cl *CharList) tick() {
	sys.gameTime++
	for _, c := range cl.runOrder {
		c.tick()
	}
}

// Prepare characters for drawing
// We once again check the movetype to minimize the difference between player sides
func (cl *CharList) cueDraw() {
	for _, c := range cl.runOrder {
		if c != nil && c.ss.moveType == MT_A {
			c.cueDraw()
		}
	}
	for _, c := range cl.runOrder {
		if c != nil && c.ss.moveType == MT_I {
			c.cueDraw()
		}
	}
	for _, c := range cl.runOrder {
		if c != nil && c.ss.moveType == MT_H {
			c.cueDraw()
		}
	}
}

func (cl *CharList) getCharWithID(id int32) *Char {
	if id < 0 {
		return nil
	}

	// Invalid ID
	ch, ok := cl.idMap[id]
	if !ok {
		return nil
	}

	// Mugen skips DestroySelf helpers here
	if ch.csf(CSF_destroy) {
		return nil
	}

	return ch
}

func (cl *CharList) getHelperIndex(c *Char, idx int32, log bool) *Char {
	var t []int32

	// Find all helpers in parent-child chain
	for j, h := range cl.runOrder {
		// Check only the relevant player number
		if h.playerNo != c.playerNo {
			continue
		}
		if c.id != h.id {
			if c.helperIndex == 0 {
				// Helpers created by the root. Direct check
				hr := h.root(false)
				if h.helperIndex != 0 && hr != nil && c.id == hr.id {
					t = append(t, int32(j))
				}
			} else {
				// Helpers created by other helpers
				hp := h.parent(false)

				// Track checked helpers to prevent infinite loops when parentIndex repeats itself
				// https://github.com/ikemen-engine/Ikemen-GO/issues/2462
				// This should no longer be necessary now that destroyed helpers are no longer valid parents
				//checked := make(map[*Char]bool)

				// Iterate until reaching the root or some error
				for hp != nil {
					//if checked[hp] {
					//	if log {
					//		sys.appendToConsole(c.warn() + "stopped infinite loop while determining helper index")
					//	}
					//	break
					//}
					//checked[hp] = true

					// Original player found to be this helper's (grand)parent. Add helper to list
					if hp.id == c.id {
						t = append(t, int32(j))
						break
					}
					// Search further up the parent chain for a relation to the original player
					hp = hp.parent(false)
				}
			}
		}
	}

	// Return the Nth helper we found
	for i := 0; i < len(t); i++ {
		ch := cl.runOrder[int32(t[i])]
		if (idx-1) == int32(i) && ch != nil {
			return ch
		}
	}

	if log {
		sys.appendToConsole(c.warn() + fmt.Sprintf("has no helper with index: %v", idx))
	}

	return nil
}

// Remove player from P2 references if it becomes invalid (standby etc)
// This function was added to selectively update every player's "P2 enemy" list instead of just clearing them,
// But because the lists are already being cleared essentially every frame anyway, the performance gain of doing this is lost
//func (cl *CharList) p2enemyDelete(c *Char) {
//	for _, e := range cl.runOrder {
//		for i, p2cl := range e.p2EnemyList {
//			if p2cl == c {
//				e.p2EnemyList = append(e.p2EnemyList[:i], e.p2EnemyList[i+1:]...)
//				break
//			}
//		}
//	}
//}

// Update enemy near or "P2" lists and return specified index
// The current approach makes the distance calculation loops only be done when necessary, using cached enemies the rest of the time
// In Mugen the P2 enemy reference seems to only refresh at the start of each frame instead
func (cl *CharList) enemyNear(c *Char, n int32, p2list, log bool) *Char {
	// Invalid reference
	if n < 0 {
		if log {
			sys.appendToConsole(c.warn() + fmt.Sprintf("has no nearest enemy: %v", n))
		}
		return nil
	}

	// Clear every player's lists if something changed
	if cl.enemyNearChanged {
		for _, c := range cl.runOrder {
			c.enemyNearP2Clear()
		}
		cl.enemyNearChanged = false
	}

	// Select EnemyNear or P2 cache
	var cache *[]*Char
	if p2list { // List for P2 redirects as well as P4, P6 and P8 triggers
		cache = &c.p2EnemyList
	} else {
		cache = &c.enemyNearList
	}

	// If we already have the Nth enemy cached, then return it
	if int(n) < len(*cache) {
		return (*cache)[n]
	}

	// Else reset the cache and start over
	*cache = (*cache)[:0]

	// Gather all valid enemies
	var enemies []*Char
	for _, e := range cl.runOrder {
		if e.playerFlag && c.isEnemyOf(e) {
			// P2 checks for alive enemies even if they are player type helpers
			if p2list && !e.scf(SCF_standby) && !e.scf(SCF_over_ko) {
				enemies = append(enemies, e)
			}
			// EnemyNear checks for dead or alive root players
			if !p2list && e.helperIndex == 0 {
				enemies = append(enemies, e)
			}
		}
	}

	// Calculate distances between all valid enemies and the player
	type enemyDist struct {
		enemy *Char
		dist  float32
	}
	pairs := make([]enemyDist, 0, len(enemies))

	for _, e := range enemies {
		// Factor x distance first
		distX := c.distX(e, c) * c.facing
		dist := distX
		// If an enemy is behind the player, an extra distance buffer is added for the "P2" list
		// This makes the player turn less frequently when surrounded
		// Mugen uses a hardcoded value of 30 pixels. Maybe it could be a character constant instead in Ikemen
		if p2list && distX < 0 {
			dist -= 30.0
		}
		// Factor z distance if applicable
		if sys.zEnabled() {
			distZ := c.distZ(e, c) * 4.0
			if p2list {
				// We'll arbitrarily give more weight to the z axis, so that the player doesn't turn as easily to enemies on a different plane
				// 4.0 is a magic number, roughly based on default x and z size ratio
				// TODO: Calculate z weight like in distzadj in player pushing, or add a global var for x/z ratio
				distZ *= 4.0
			}
			// Calculate the hypotenuse between both
			dist = float32(math.Hypot(float64(distX), float64(distZ)))
		}
		// Append this enemy and their distance
		pairs = append(pairs, enemyDist{enemy: e, dist: dist})
	}

	// Sort enemies by shortest absolute distance
	sort.SliceStable(pairs, func(i, j int) bool {
		return AbsF(pairs[i].dist) < AbsF(pairs[j].dist)
	})

	// Rebuild cache
	*cache = make([]*Char, len(pairs))
	for i, pair := range pairs {
		(*cache)[i] = pair.enemy
	}

	// If reference exceeds number of valid enemies
	if int(n) >= len(*cache) {
		if log {
			sys.appendToConsole(c.warn() + fmt.Sprintf("has no nearest enemy: %v", n))
		}
		return nil
	}

	// Return Nth enemy
	return (*cache)[n]
}

type Platform struct {
	name string
	id   int32

	pos    [2]float32
	size   [2]int32
	offset [2]int32

	anim        int32
	activeTime  int32
	isSolid     bool
	borderFall  bool
	destroySelf bool

	localScale float32
	ownerID    int32
}
