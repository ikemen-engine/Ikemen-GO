package main

import (
	"fmt"
	"io"
	"math"
	"os"
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
	SCF_inputwait
	SCF_ko
	SCF_ko_round_middle
	SCF_over
	SCF_standby
)

// These flags are reset manually
type CharSpecialFlag uint32

const (
	CSF_angledraw CharSpecialFlag = 1 << iota
	CSF_backedge
	CSF_backwidth
	CSF_bottomheight
	CSF_destroy
	CSF_frontedge
	CSF_frontwidth
	CSF_gethit
	CSF_movecamera_x
	CSF_movecamera_y
	CSF_playerpush
	CSF_posfreeze
	CSF_screenbound
	CSF_stagebound
	CSF_topheight
	CSF_trans
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
	ASF_nobrake
	ASF_nocrouch
	ASF_nostand
	ASF_nojump
	ASF_noairjump
	ASF_nohardcodedkeys
	ASF_nogetupfromliedown
	ASF_nofastrecoverfromliedown
	ASF_nofallcount
	ASF_nofalldefenceup
	ASF_noturntarget
	ASF_noinput
	ASF_nopowerbardisplay
	ASF_autoguard
	ASF_animfreeze
	ASF_postroundinput
	ASF_nohitdamage
	ASF_noguarddamage
	ASF_nodizzypointsdamage
	ASF_noguardpointsdamage
	ASF_noredlifedamage
	ASF_nomakedust
	ASF_noguardko
	ASF_nokovelocity
	ASF_noailevel
	ASF_nointroreset
	ASF_sizepushonly
	ASF_animatehitpause
	ASF_drawunder
	ASF_runfirst
	ASF_runlast
	ASF_projtypecollision // TODO: Make this a parameter for normal projectiles as well?
	ASF_nofallhitflag
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
	GSF_nofightdisplay
	GSF_nokodisplay
	GSF_norounddisplay
	GSF_nowindisplay
	GSF_roundnotskip
	GSF_roundfreeze
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

func (cr *ClsnRect) Add(clsn []float32, x, y, xs, ys, angle float32) {
	x = (x - sys.cam.Pos[0]) * sys.cam.Scale
	y = (y*sys.cam.Scale - sys.cam.Pos[1]) + sys.cam.GroundLevel()
	xs *= sys.cam.Scale
	ys *= sys.cam.Scale
	sw := float32(sys.gameWidth)
	sh := float32(0) //float32(sys.gameHeight)
	for i := 0; i+3 < len(clsn); i += 4 {
		offx := sw / 2
		offy := sh
		rect := [...]float32{
			AbsF(xs) * clsn[i], AbsF(ys) * clsn[i+1],
			xs * (clsn[i+2] - clsn[i]), ys * (clsn[i+3] - clsn[i+1]),
			(x + offx) * sys.widthScale, (y + offy) * sys.heightScale, angle}
		*cr = append(*cr, rect)
	}
}
func (cr ClsnRect) draw(trans int32) {
	paltex := PaletteToTexture(sys.clsnSpr.Pal)
	for _, c := range cr {
		params := RenderParams{
			sys.clsnSpr.Tex, paltex, sys.clsnSpr.Size,
			-c[0] * sys.widthScale, -c[1] * sys.heightScale, notiling,
			c[2] * sys.widthScale, c[2] * sys.widthScale, c[3] * sys.heightScale, 1, 0,
			1, 1, Rotation{c[6], 0, 0}, 0, trans, -1, nil, &sys.scrrect, c[4], c[5], 0, 0, 0, 0,
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
	cd.intpersistindex = NumVar
	cd.floatpersistindex = NumFvar
}

type CharSize struct {
	xscale float32
	yscale float32
	ground struct {
		back  float32
		front float32
	}
	air struct {
		back  float32
		front float32
	}
	height struct {
		stand  float32
		crouch float32
		air    [2]float32
		down   float32
	}
	attack struct {
		dist struct {
			front float32
			back  float32
		}
		depth struct {
			front float32
			back  float32
		}
	}
	proj struct {
		attack struct {
			dist struct {
				front float32
				back  float32
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
	depth      float32 // Former depth
	weight     int32
	pushfactor float32
}

func (cs *CharSize) init() {
	*cs = CharSize{}
	cs.xscale = 1
	cs.yscale = 1
	cs.ground.back = 15
	cs.ground.front = 16
	cs.air.back = 12
	cs.air.front = 12
	cs.height.stand = 60
	cs.height.crouch = 60
	cs.height.air = [...]float32{60, 0}
	cs.height.down = 60
	cs.attack.dist.front = 160
	cs.attack.dist.back = 0
	cs.proj.attack.dist.front = 90
	cs.proj.attack.dist.back = 0
	cs.proj.doscale = 0
	cs.head.pos = [...]float32{-5, -90}
	cs.mid.pos = [...]float32{-5, -60}
	cs.shadowoffset = 0
	cs.draw.offset = [...]float32{0, 0}
	cs.depth = 3
	cs.attack.depth.front = 4
	cs.attack.depth.back = 4
	cs.weight = 100
	cs.pushfactor = 1
}

type CharVelocity struct {
	walk struct {
		fwd  float32
		back float32
		up   float32
		down float32
	}
	run struct {
		fwd  [2]float32
		back [2]float32
		up   [2]float32
		down [2]float32
	}
	jump struct {
		neu  [2]float32
		back float32
		fwd  float32
		up   float32
		down float32
	}
	runjump struct {
		back [2]float32
		fwd  [2]float32
		up   float32
		down float32
	}
	airjump struct {
		neu  [2]float32
		back float32
		fwd  float32
		up   float32
		down float32
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
	cv.air.gethit.ko.add = [...]float32{-2, -2}
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

// Aiuchi = trading hits
type AiuchiType int32

const (
	AT_Hit AiuchiType = iota
	AT_Miss
	AT_Dodge
)

type Fall struct {
	animtype       Reaction
	xvelocity      float32
	yvelocity      float32
	zvelocity      float32
	recover        bool
	recovertime    int32
	damage         int32
	kill           bool
	envshake_time  int32
	envshake_freq  float32
	envshake_ampl  int32
	envshake_phase float32
	envshake_mul   float32
}

type HitDef struct {
	isprojectile               bool // Projectile state controller
	attr                       int32
	reversal_attr              int32
	hitflag                    int32
	guardflag                  int32
	affectteam                 int32 // -1F, 0B, 1E
	teamside                   int
	animtype                   Reaction
	air_animtype               Reaction
	priority                   int32
	bothhittype                AiuchiType
	hitdamage                  int32
	guarddamage                int32
	pausetime                  int32
	shaketime                  int32
	guard_pausetime            int32
	guard_shaketime            int32
	sparkno                    int32
	sparkno_ffx                string
	sparkangle                 float32
	guard_sparkno              int32
	guard_sparkno_ffx          string
	guard_sparkangle           float32
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
	guard_dist                 [2]int32
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
	forcestand                 int32
	forcecrouch                int32
	ground_fall                bool
	air_fall                   bool
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
	mindist                    [3]float32
	maxdist                    [3]float32
	snap                       [3]float32
	snaptime                   int32
	fall                       Fall
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
	attack                     struct {
		depth [2]float32
	}
}

func (hd *HitDef) clear(localscl float32) {
	// Convert local scale back to 4:3 in order to keep values consistent in widescreen
	originLs := localscl * (320 / float32(sys.gameWidth))

	*hd = HitDef{
		isprojectile:       false,
		hitflag:            int32(ST_S | ST_C | ST_A | ST_F),
		affectteam:         1,
		teamside:           -1,
		animtype:           RA_Light,
		air_animtype:       RA_Unknown,
		priority:           4,
		bothhittype:        AT_Hit,
		sparkno:            -1,
		sparkno_ffx:        "f",
		sparkangle:         0,
		guard_sparkno:      -1,
		guard_sparkno_ffx:  "f",
		guard_sparkangle:   0,
		hitsound:           [2]int32{-1, 0},
		hitsound_channel:   -1,
		hitsound_ffx:       "f",
		guardsound:         [2]int32{-1, 0},
		guardsound_channel: -1,
		guardsound_ffx:     "f",
		ground_type:        HT_High,
		air_type:           HT_Unknown,
		// Both default to 20, not documented in Mugen docs.
		air_hittime:  20,
		down_hittime: 20,

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

		p1sprpriority:    1,
		p1stateno:        -1,
		p2stateno:        -1,
		forcestand:       IErr,
		forcecrouch:      IErr,
		guard_dist:       [...]int32{-1, -1},
		id:               -1, // fixes an issue where targetState with ID can trigger on hitdefs with unset ID's
		chainid:          -1,
		nochainid:        [8]int32{-1, -1, -1, -1, -1, -1, -1, -1},
		numhits:          1,
		hitgetpower:      IErr,
		guardgetpower:    IErr,
		hitgivepower:     IErr,
		guardgivepower:   IErr,
		envshake_freq:    60,
		envshake_ampl:    -4,
		envshake_phase:   float32(math.NaN()),
		envshake_mul:     1.0,
		mindist:          [...]float32{float32(math.NaN()), float32(math.NaN()), float32(math.NaN())},
		maxdist:          [...]float32{float32(math.NaN()), float32(math.NaN()), float32(math.NaN())},
		snap:             [...]float32{float32(math.NaN()), float32(math.NaN()), float32(math.NaN())},
		hitonce:          -1,
		kill:             true,
		guard_kill:       true,
		playerNo:         -1,
		dizzypoints:      IErr,
		guardpoints:      IErr,
		hitredlife:       IErr,
		guardredlife:     IErr,
		score:            [...]float32{float32(math.NaN()), float32(math.NaN())},
		p2clsncheck:      -1,
		p2clsnrequire:    -1,
		down_recover:     true,
		down_recovertime: -1,
		air_juggle:       IErr,
		// Fall group
		fall: Fall{
			animtype:       RA_Unknown,
			xvelocity:      float32(math.NaN()),
			yvelocity:      -4.5 / originLs,
			zvelocity:      0, // Should this work like the X component instead?
			recover:        true,
			recovertime:    4,
			kill:           true,
			envshake_freq:  60,
			envshake_ampl:  IErr,
			envshake_phase: float32(math.NaN()),
			envshake_mul:   1.0,
		},
		// Attack depth
		attack: struct{ depth [2]float32 }{
			[2]float32{4 / originLs, 4 / originLs},
		},
	}
	// PalFX
	hd.palfx.mul = [...]int32{255, 255, 255}
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
	hitBy [][2]int32
	//hit1           [2]int32
	//hit2           [2]int32
	attr              int32
	_type             HitType
	animtype          Reaction
	airanimtype       Reaction
	groundanimtype    Reaction
	airtype           HitType
	groundtype        HitType
	damage            int32
	hitcount          int32
	guardcount        int32
	fallcount         int32
	hitshaketime      int32
	hittime           int32
	slidetime         int32
	ctrltime          int32
	xvel              float32
	yvel              float32
	zvel              float32
	xaccel            float32
	yaccel            float32
	zaccel            float32
	hitid             int32
	xoff              float32
	yoff              float32
	zoff              float32
	fall              Fall
	playerNo          int
	fallflag          bool
	guarded           bool
	p2getp1state      bool
	forcestand        bool
	forcecrouch       bool
	id                int32
	dizzypoints       int32
	guardpoints       int32
	redlife           int32
	score             float32
	hitdamage         int32
	guarddamage       int32
	power             int32
	hitpower          int32
	guardpower        int32
	hitredlife        int32
	guardredlife      int32
	fatal             bool
	kill              bool
	priority          int32
	facing            int32
	ground_velocity   [3]float32
	air_velocity      [3]float32
	down_velocity     [3]float32
	guard_velocity    [3]float32
	airguard_velocity [3]float32
	frame             bool
	cheeseKO          bool
	down_recover      bool
	down_recovertime  int32
}

func (ghv *GetHitVar) clear(c *Char) {
	// Convert local scale back to 4:3 in order to keep values consistent in widescreen
	originLs := c.localscl * (320 / float32(sys.gameWidth))

	*ghv = GetHitVar{
		hittime:  -1,
		yaccel:   0.35 / originLs,
		xoff:     ghv.xoff,
		yoff:     ghv.yoff,
		zoff:     ghv.zoff,
		hitid:    -1,
		playerNo: -1,
		// Fall group
		fall: Fall{
			animtype:  RA_Unknown,
			xvelocity: float32(math.NaN()),
			yvelocity: -4.5 / originLs,
			zvelocity: float32(math.NaN()),
		},
	}
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
	for _, v := range ghv.hitBy {
		if v[0] == id || v[0] == -id {
			return true
		}
	}
	return false
}
func (ghv GetHitVar) getJuggle(id, defaultJuggle int32) int32 {
	for _, v := range ghv.hitBy {
		if v[0] == id {
			return v[1]
		}
	}
	return defaultJuggle
}
func (ghv *GetHitVar) dropId(id int32) {
	for i, v := range ghv.hitBy {
		if v[0] == id {
			ghv.hitBy = append(ghv.hitBy[:i], ghv.hitBy[i+1:]...)
			break
		}
	}
}
func (ghv *GetHitVar) addId(id, juggle int32) {
	juggle = ghv.getJuggle(id, juggle)
	ghv.dropId(id)
	ghv.hitBy = append(ghv.hitBy, [...]int32{id, juggle})
}

type HitBy struct {
	flag     int32
	time     int32
	not      bool
	playerid int32
	playerno int
	stack    bool
}

type HitOverride struct {
	attr      int32
	stateno   int32
	time      int32
	forceair  bool
	keepState bool
	playerNo  int
}

func (ho *HitOverride) clear() {
	*ho = HitOverride{stateno: -1, keepState: false, playerNo: -1}
}

type MoveHitVar struct {
	cornerpush float32
	frame      bool
	id         int32
	overridden bool
	playerNo   int
	sparkxy    [2]float32
	uniqhit    int32
}

func (mhv *MoveHitVar) clear() {
	*mhv = MoveHitVar{}
}

type aimgImage struct {
	anim       Animation
	pos        [2]float32
	scl        [2]float32
	ascl       [2]float32
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
	alpha          [2]int32
	palfx          []PalFX
	imgs           [64]aimgImage
	imgidx         int32
	restgap        int32
	reccount       int32
	timecount      int32
	priority       int32
	ignorehitpause bool
}

func newAfterImage() *AfterImage {
	ai := &AfterImage{palfx: make([]PalFX, sys.afterImageMax)}
	for i := range ai.palfx {
		ai.palfx[i].enable, ai.palfx[i].negType = true, true
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
	ai.alpha = [...]int32{-1, 0}
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
		img.anim = *sd.anim
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
		img.pos = sd.pos
		img.scl = sd.scl
		img.rot = sd.rot
		img.projection = sd.projection
		img.fLength = sd.fLength
		img.ascl = sd.ascl
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
	end := Min(sys.afterImageMax,
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
			sprs.add(&SprData{&img.anim, &ai.palfx[step], img.pos,
				img.scl, ai.alpha, img.priority - step, // Afterimages decrease in sprpriority over time
				img.rot, img.ascl, false, sd.bright, sd.oldVer, sd.facing,
				sd.posLocalscl, img.projection, img.fLength, sd.window})
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
	animelem            int32
	animfreeze          bool
	//ontop                bool
	under          bool
	alpha          [2]int32
	ownpal         bool
	ignorehitpause bool
	rot            Rotation
	anglerot       [3]float32
	projection     Projection
	fLength        float32
	oldPos         [3]float32
	newPos         [3]float32
	playerId       int32
	palfx          *PalFX
	palfxdef       PalFXDef
	window         [4]float32
	//lockSpriteFacing     bool
	localscl             float32
	blendmode            int32
	start_animelem       int32
	start_scale          [2]float32
	start_rot            [3]float32
	start_alpha          [2]int32
	start_fLength        float32
	interpolate          bool
	interpolate_time     [2]int32
	interpolate_animelem [3]int32
	interpolate_scale    [4]float32
	interpolate_alpha    [5]int32
	interpolate_pos      [6]float32
	interpolate_angle    [6]float32
	interpolate_fLength  [2]float32
	animNo               int32
	interPos             [3]float32
}

func (e *Explod) clear() {
	*e = Explod{
		id:                IErr,
		bindtime:          1,
		scale:             [...]float32{1, 1},
		removetime:        -2,
		postype:           PT_P1,
		space:             Space_none,
		relativef:         1,
		facing:            1,
		vfacing:           1,
		localscl:          1,
		projection:        Projection_Orthographic,
		window:            [4]float32{0, 0, 0, 0},
		animelem:          1,
		blendmode:         0,
		alpha:             [...]int32{-1, 0},
		playerId:          -1,
		bindId:            -2,
		ignorehitpause:    true,
		interpolate_scale: [...]float32{1, 1, 0, 0},
		friction:          [3]float32{1, 1, 1},
	}
}
func (e *Explod) setX(x float32) {
	e.pos[0], e.oldPos[0], e.newPos[0] = x, x, x
}
func (e *Explod) setY(y float32) {
	e.pos[1], e.oldPos[1], e.newPos[1] = y, y, y
}
func (e *Explod) setZ(z float32) {
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
			e.offset[2] = ClampF(posZ, sys.stage.stageCamera.topz, sys.stage.stageCamera.botz)
		} else {
			e.setX(posX)
			e.setY(posY)
			e.setZ(posZ)
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
		if p2 := sys.charList.enemyNear(c, 0, true, true, false); p2 != nil {
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
func (e *Explod) setAnimElem() {
	if e.anim != nil && e.animelem >= 1 {
		e.anim.SetAnimElem(Clamp(e.animelem, 1, int32(len(e.anim.frames))))
	}
}
func (e *Explod) update(oldVer bool, playerNo int) {
	if e.anim == nil {
		e.id = IErr
	}
	if e.id == IErr {
		e.anim = nil
		return
	}
	if sys.chars[playerNo][0].scf(SCF_disabled) {
		return
	}
	var c *Char
	if !e.ignorehitpause || e.removeongethit || e.removeonchangestate {
		c = sys.playerID(e.playerId)
	}
	// Remove on get hit
	if sys.tickNextFrame() && e.removeongethit &&
		c != nil && c.csf(CSF_gethit) && !c.inGuardState() {
		e.id, e.anim = IErr, nil
		return
	}
	// Remove on ChangeState
	if sys.tickNextFrame() && e.removeonchangestate && e.statehaschanged {
		e.id, e.anim = IErr, nil
		return
	}
	p := false
	if sys.super > 0 {
		p = (e.supermovetime >= 0 && e.time >= e.supermovetime) || e.supermovetime < -2
	} else if sys.pause > 0 {
		p = (e.pausemovetime >= 0 && e.time >= e.pausemovetime) || e.pausemovetime < -2
	}
	act := !p
	if act && !e.ignorehitpause {
		act = c == nil || c.acttmp%2 >= 0
	}
	if sys.tickFrame() {
		if e.removetime >= 0 && e.time >= e.removetime ||
			act && e.removetime < -1 && e.anim.loopend {
			e.id, e.anim = IErr, nil
			return
		}
	}
	if e.bindtime != 0 && (e.space == Space_stage ||
		(e.space == Space_screen && e.postype <= PT_P2)) {
		if c := sys.playerID(e.bindId); c != nil {
			e.pos[0] = c.interPos[0]*c.localscl/e.localscl + c.offsetX()*c.localscl/e.localscl
			e.pos[1] = c.interPos[1]*c.localscl/e.localscl + c.offsetY()*c.localscl/e.localscl
			e.pos[2] = c.interPos[2] * c.localscl / e.localscl
		} else {
			// Doesn't seem necessary to do this, since MUGEN 1.1 seems to carry bindtime even if
			// you change bindId to something that doesn't point to any character
			// e.bindtime = 0
			// e.setX(e.pos[0])
			// e.setY(e.pos[1])
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
	if e.palfx != nil && (e.anim.sff != sys.ffx["f"].fsff || e.ownpal) {
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
	if e.interpolate {
		e.Interpolate(act, &scale, &alp, &anglerot, &fLength)
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
	e.interPos = [3]float32{
		(e.pos[0] + e.offset[0] + off[0] + e.interpolate_pos[0]) * e.localscl,
		(e.pos[1] + e.offset[1] + off[1] + e.interpolate_pos[1]) * e.localscl,
		(e.pos[2] + e.offset[2] + off[2] + e.interpolate_pos[2]) * e.localscl,
	}

	// Set drawing position
	drawpos := [2]float32{e.interPos[0], e.interPos[1]}

	// Set scale
	drawscale := [2]float32{facing * scale[0] * e.localscl, e.vfacing * scale[1] * e.localscl}

	// Apply Z axis perspective
	if e.space == Space_stage && sys.zmin != sys.zmax {
		zscale := sys.updateZScale(e.pos[2], e.localscl)
		drawpos[0] *= zscale
		drawpos[1] *= zscale
		drawpos[1] += e.interPos[2] * e.localscl
		drawscale[0] *= zscale
		drawscale[1] *= zscale
	}

	var ewin = [4]float32{e.window[0] * e.localscl * facing, e.window[1] * e.localscl * e.vfacing, e.window[2] * e.localscl * facing, e.window[3] * e.localscl * e.vfacing}

	// Add sprite to draw list
	sd := &SprData{e.anim, pfx, drawpos, drawscale,
		alp, e.sprpriority + int32(e.pos[2]*e.localscl), rot, [...]float32{1, 1},
		e.space == Space_screen, playerNo == sys.superplayer, oldVer, facing, 1, int32(e.projection), fLength, ewin}
	sprs.add(sd)

	// Add shadow if color is not 0
	sdwclr := e.shadow[0]<<16 | e.shadow[1]&0xff<<8 | e.shadow[2]&0xff
	if sdwclr != 0 {
		sdwalp := 255 - alp[1]
		if sdwalp < 0 {
			sdwalp = 256
		}
		sys.shadows.add(&ShadowSprite{sd, sdwclr, sdwalp, [2]float32{0, 0}, [2]float32{0, 0}, 0})
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

		if e.bindtime > 0 {
			e.bindtime--
		}
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
		} else {
			e.setX(e.pos[0])
			e.setY(e.pos[1])
			e.setZ(e.pos[2])
		}
	}
}

func (e *Explod) Interpolate(act bool, scale *[2]float32, alpha *[2]int32, anglerot *[3]float32, fLength *float32) {
	if sys.tickNextFrame() && act {
		t := float32(e.interpolate_time[1]) / float32(e.interpolate_time[0])
		e.interpolate_fLength[0] = Lerp(e.interpolate_fLength[1], e.start_fLength, t)
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
				if e.blendmode == 1 {
					e.interpolate_alpha[i] = Clamp(int32(Lerp(float32(e.interpolate_alpha[i+2]), float32(e.start_alpha[i]), t)), 0, 255)
				}
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
			if e.blendmode == 1 {
				if (*alpha)[0] == 1 && (*alpha)[1] == 255 {
					(*alpha)[0] = 0
				} else {
					(*alpha)[i] = int32(float32(e.interpolate_alpha[i]) * (float32(e.alpha[i]) / 255))
				}

			}
		}
		(*anglerot)[i] = e.interpolate_angle[i] + e.anglerot[i]
	}
	*fLength = e.interpolate_fLength[0] + e.fLength
}

func (e *Explod) setStartParams(pfd *PalFXDef) {
	e.start_animelem = e.animelem
	e.start_fLength = e.fLength
	for i := 0; i < 3; i++ {
		if i < 2 {
			e.start_scale[i] = e.scale[i]
			e.start_alpha[i] = e.alpha[i]
		}
		e.start_rot[i] = e.anglerot[i]
	}
	if e.interpolate {
		e.fLength = 0
		for i := 0; i < 3; i++ {
			if e.ownpal {
				pfd.mul[i] = 256
				pfd.add[i] = 0
			}
			if i < 2 {
				e.scale[i] = 1
				if e.blendmode == 1 {
					e.alpha[i] = 255
				}
			}
			e.anglerot[i] = 0
		}
		if e.ownpal {
			pfd.color = 1
			pfd.hue = 0
		}
	}
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
	}
}

type Projectile struct {
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
	angle           float32
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
	localscl        float32
	parentAttackmul [4]float32
	platform        bool
	platformWidth   [2]float32
	platformHeight  [2]float32
	platformAngle   float32
	platformFence   bool
	remflag         bool
	freezeflag      bool
	contactflag     bool
}

func newProjectile() *Projectile {
	p := &Projectile{}
	p.clear()
	return p
}

func (p *Projectile) clear() {
	*p = Projectile{
		id:             IErr,
		hitanim:        -1,
		remanim:        IErr,
		cancelanim:     IErr,
		scale:          [...]float32{1, 1},
		clsnScale:      [...]float32{1, 1},
		clsnAngle:      0,
		remove:         true,
		localscl:       1,
		removetime:     -1,
		velmul:         [...]float32{1, 1, 1},
		hits:           1,
		totalhits:      1,
		priority:       1,
		priorityPoints: 1,
		sprpriority:    3,
		edgebound:      40,
		stagebound:     40,
		heightbound:    [...]int32{-240, 1},
		depthbound:     math.MaxInt32,
		facing:         1,
		aimg:           *newAfterImage(),
		platformFence:  true,
	}
	p.hitdef.clear(p.localscl)
}

func (p *Projectile) setPos(pos [3]float32) {
	p.pos, p.oldPos, p.newPos = pos, pos, pos
}

func (p *Projectile) paused(playerNo int) bool {
	//if !sys.chars[playerNo][0].pause() {
	if sys.super > 0 {
		if p.supermovetime == 0 || p.supermovetime < -1 {
			return true
		}
	} else if sys.pause > 0 {
		if p.pausemovetime == 0 || p.pausemovetime < -1 {
			return true
		}
	}
	//}
	return false
}
func (p *Projectile) update(playerNo int) {
	// Check projectile removal conditions
	if sys.tickFrame() && !p.paused(playerNo) && p.hitpause == 0 {
		if p.anim >= 0 {
			if !p.remflag {
				remove := true
				if p.hits < 0 {
					// Remove behavior
					if p.hits == -1 && p.remove {
						if p.hitanim != p.anim || p.hitanim_ffx != p.anim_ffx {
							p.ani = sys.chars[playerNo][0].getAnim(p.hitanim, p.hitanim_ffx, true)
						}
					}
					// Cancel behavior
					if p.hits == -2 {
						if p.cancelanim != p.anim || p.cancelanim_ffx != p.anim_ffx {
							p.ani = sys.chars[playerNo][0].getAnim(p.cancelanim, p.cancelanim_ffx, true)
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
						p.ani = sys.chars[playerNo][0].getAnim(p.remanim, p.remanim_ffx, true)
					}
				} else {
					remove = false
				}
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
					// In Mugen, projectiles can hit even after their removetime expires - https://github.com/ikemen-engine/Ikemen-GO/issues/1362
					//if p.hits >= 0 {
					//	p.hits = -1
					//}
				}
			}
		}
		if p.remflag {
			if p.ani != nil && (p.ani.totaltime <= 0 || p.ani.AnimTime() == 0) {
				p.ani = nil
			}
			if p.ani == nil && p.id >= 0 {
				p.id = ^p.id
			}
		}
	}
	if p.paused(playerNo) || p.hitpause > 0 || p.freezeflag {
		p.setPos(p.pos)
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

// This subtracts projectile hits when two projectiles clash
func (p *Projectile) cancelHits(opp *Projectile) {
	// Check priority
	if p.priorityPoints > opp.priorityPoints {
		p.priorityPoints--
	} else {
		p.hits--
	}
	// Flag for removal
	if p.hits <= 0 {
		p.hits = -2 // -2 hits means the projectile was cancelled
	}
	// Set hitpause
	if p.hits > 0 {
		p.hitpause = Max(0, p.hitdef.pausetime) // -Btoi(c.gi().mugenver[0] == 0))
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
	if p.ani == nil || len(p.ani.frames) == 0 || p.ani.CurrentFrame().Clsn2() == nil {
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
			pr := &sys.projs[i][j]

			// Skip if other projectile can't trade
			if pr.remflag || pr.hits < 0 || pr.id < 0 {
				continue
			}

			// Skip if other projectile can't run collision check
			if pr.ani == nil || len(pr.ani.frames) == 0 || pr.ani.CurrentFrame().Clsn2() == nil {
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
			if !sys.zAxisOverlap(p.pos[2], p.hitdef.attack.depth[0], p.hitdef.attack.depth[1], p.localscl,
				pr.pos[2], pr.hitdef.attack.depth[0], pr.hitdef.attack.depth[1], pr.localscl) {
				continue
			}

			// Run Clsn check
			clsn1 := p.ani.CurrentFrame().Clsn2() // Projectiles trade with their Clsn2 only
			clsn2 := pr.ani.CurrentFrame().Clsn2()
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

func (p *Projectile) tick(playerNo int) {
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
	if !p.paused(playerNo) {
		if p.hitpause <= 0 {
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
func (p *Projectile) cueDraw(oldVer bool, playerNo int) {
	notpause := p.hitpause <= 0 && !p.paused(playerNo)
	if sys.tickFrame() && p.ani != nil && notpause {
		p.ani.UpdateSprite()
	}
	// Projectile Clsn display
	if sys.clsnDraw && p.ani != nil {
		if frm := p.ani.drawFrame(); frm != nil {
			if clsn := frm.Clsn1(); len(clsn) > 0 {
				sys.debugc1hit.Add(clsn, p.pos[0]*p.localscl, p.pos[1]*p.localscl,
					p.clsnScale[0]*p.localscl*p.facing*p.zScale,
					p.clsnScale[1]*p.localscl*p.zScale,
					p.clsnAngle*p.facing)
			}
			if clsn := frm.Clsn2(); len(clsn) > 0 {
				sys.debugc2hb.Add(clsn, p.pos[0]*p.localscl, p.pos[1]*p.localscl,
					p.clsnScale[0]*p.localscl*p.facing*p.zScale,
					p.clsnScale[1]*p.localscl*p.zScale,
					p.clsnAngle*p.facing)
			}
		}
	}
	if sys.tickNextFrame() && (notpause || !p.paused(playerNo)) {
		if p.ani != nil && notpause {
			p.ani.Action()
		}
	}

	pos := [2]float32{p.interPos[0] * p.localscl, p.interPos[1] * p.localscl}

	scl := [...]float32{p.facing * p.scale[0] * p.localscl * p.zScale,
		p.scale[1] * p.localscl * p.zScale}

	// Apply Z axis perspective
	if sys.zmin != sys.zmax {
		pos[0] *= p.zScale
		pos[1] *= p.zScale
		pos[1] += p.interPos[2] * p.localscl
	}

	sprs := &sys.spritesLayer0
	if p.layerno > 0 {
		sprs = &sys.spritesLayer1
	} else if p.layerno < 0 {
		sprs = &sys.spritesLayerN1
	}

	if p.ani != nil {
		// Add sprite to draw list
		sd := &SprData{p.ani, p.palfx, pos, scl, [2]int32{-1},
			p.sprpriority + int32(p.pos[2]*p.localscl), Rotation{p.facing * p.angle, 0, 0}, [...]float32{1, 1}, false, playerNo == sys.superplayer,
			sys.cgi[playerNo].mugenver[0] != 1, p.facing, 1, 0, 0, [4]float32{0, 0, 0, 0}}
		p.aimg.recAndCue(sd, sys.tickNextFrame() && notpause, false, p.layerno)
		sprs.add(sd)
		// Add a shadow if color is not 0
		sdwclr := p.shadow[0]<<16 | p.shadow[1]&255<<8 | p.shadow[2]&255
		if sdwclr != 0 {
			sys.shadows.add(&ShadowSprite{sd, sdwclr, 0, [2]float32{0, p.pos[2]}, [2]float32{0, p.pos[2]}, 0})
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
	def              string
	nameLow          string
	displayname      string
	displaynameLow   string
	author           string
	authorLow        string
	lifebarname      string
	palkeymap        [MaxPalNo]int32
	sff              *Sff
	palettedata      *Palette
	snd              *Snd
	anim             AnimationTable
	palno, drawpalno int32
	pal              [MaxPalNo]string
	palExist         [MaxPalNo]bool
	palSelectable    [MaxPalNo]bool
	ikemenver        [3]uint16
	ikemenverF       float32
	mugenver         [2]uint16
	data             CharData
	velocity         CharVelocity
	movement         CharMovement
	states           map[int32]StateBytecode
	wakewakaLength   int32
	pctype           ProjContact
	pctime, pcid     int32
	projidcount      int
	quotes           [MaxQuotes]string
	portraitscale    float32
	constants        map[string]float32
	remapPreset      map[string]RemapPreset
	remappedpal      [2]int32
	localcoord       [2]float32
	fnt              [10]*Fnt
}

func (cgi *CharGlobalInfo) clearPCTime() {
	cgi.pctype = PC_Hit
	cgi.pctime = -1
	cgi.pcid = 0
}

// StateState contains the state variables like stateNo, prevStateNo, time, stateType, moveType, and physics of the current state.
type StateState struct {
	stateType       StateType
	prevStateType   StateType
	moveType        MoveType
	prevMoveType    MoveType
	storeMoveType   bool
	physics         StateType
	ps              []int32
	wakegawakaranai [MaxSimul*2 + MaxAttachedChar][]bool
	no, prevno      int32
	time            int32
	sb              StateBytecode
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
	for i, v := range ss.wakegawakaranai {
		if len(v) < int(sys.cgi[i].wakewakaLength) {
			ss.wakegawakaranai[i] = make([]bool, sys.cgi[i].wakewakaLength)
		} else {
			for i := range v {
				v[i] = false
			}
		}
	}
	ss.clearWw()
	ss.no, ss.prevno = 0, 0
	ss.time = 0
	ss.sb = StateBytecode{}
}
func (ss *StateState) clearWw() {
	for _, v := range ss.wakegawakaranai {
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
	airJumpCount      int32
	assertFlag        AssertSpecialFlag
	hitCount          int32
	guardCount        int32
	uniqHitCount      int32
	pauseMovetime     int32
	superMovetime     int32
	prevPauseMovetime int32
	prevSuperMovetime int32
	unhittableTime    int32
	bindTime          int32
	bindToId          int32
	bindPos           [3]float32
	bindPosAdd        [3]float32
	bindFacing        float32
	hitPauseTime      int32
	angle             float32
	angleScale        [2]float32
	alpha             [2]int32
	systemFlag        SystemCharFlag
	specialFlag       CharSpecialFlag
	sprPriority       int32
	layerNo           int32
	receivedDmg       int32
	receivedHits      int32
	cornerVelOff      float32
	width             [2]float32
	edge              [2]float32
	height            [2]float32
	attackMul         [4]float32 // 0 Damage, 1 Red Life, 2 Dizzy Points, 3 Guard Points
	superDefenseMul   float32
	fallDefenseMul    float32
	customDefense     float32
	finalDefense      float64
	defenseMulDelay   bool
	counterHit        bool
	prevNoStandGuard  bool
}

type Char struct {
	name                string
	palfx               *PalFX
	anim                *Animation
	curFrame            *AnimFrame
	cmd                 []CommandList
	ss                  StateState
	key                 int
	id                  int32
	index               int32
	runorder            int32
	helperId            int32
	helperIndex         int32
	parentIndex         int32
	playerNo            int
	teamside            int
	keyctrl             [4]bool
	player              bool
	hprojectile         bool // Helper type projectile. Currently unused
	animPN              int
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
	localcoord          float32
	localscl            float32
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
	ho                  [8]HitOverride
	hoIdx               int
	hoKeepState         bool
	mctype              MoveContact
	mctime              int32
	children            []*Char
	targets             []int32
	hitdefTargets       []int32
	hitdefTargetsBuffer []int32
	enemynear           [2][]*Char
	p2enemy             []*Char
	pos                 [3]float32
	interPos            [3]float32 // Interpolated position. For the visuals when game and logic speed are different
	oldPos              [3]float32
	dustOldPos          float32
	vel                 [3]float32
	facing              float32
	ivar                [NumVar + NumSysVar]int32
	fvar                [NumFvar + NumSysFvar]float32
	CharSystemVar
	aimg            AfterImage
	soundChannels   SoundChannels
	p1facing        float32
	cpucmd          int32
	attackDist      [2]float32
	offset          [2]float32
	stchtmp         bool
	inguarddist     bool
	pushed          bool
	hitdefContact   bool
	atktmp          int8 // 1 hitdef can hit, 0 cannot hit, -1 other
	hittmp          int8 // 0 idle, 1 being hit, 2 falling, -1 reversaldef
	acttmp          int8 // 1 unpaused, 0 default, -1 hitpause, -2 pause
	minus           int8 // current negative state
	platformPosY    float32
	groundAngle     float32
	ownpal          bool
	winquote        int32
	memberNo        int
	selectNo        int
	inheritJuggle   int32
	inheritChannels int32
	mapArray        map[string]float32
	mapDefault      map[string]float32
	remapSpr        RemapPreset
	clipboardText   []string
	dialogue        []string
	immortal        bool
	kovelocity      bool
	preserve        int32
	inputFlag       InputBits
	pauseBool       bool
	downHitOffset   bool
	koEchoTime      int32
	groundLevel     float32
	sizeBox         []float32
	shadowOffset    [2]float32
	reflectOffset   [2]float32
	ownclsnscale    bool
	pushPriority    int32
}

func newChar(n int, idx int32) (c *Char) {
	c = &Char{aimg: *newAfterImage(), zScale: 1}
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
	c.clear1()
	c.playerNo, c.helperIndex = n, idx
	c.animPN = c.playerNo
	if c.helperIndex == 0 {
		c.player = true
		c.kovelocity = true
		c.keyctrl = [...]bool{true, true, true, true}
	} else {
		c.mapArray = make(map[string]float32)
		c.remapSpr = make(RemapPreset)
	}
	c.key = n
	if n >= 0 && n < len(sys.com) && sys.com[n] != 0 {
		c.key ^= -1
	}
}
func (c *Char) clearState() {
	c.ss.clear()
	c.hitdef.clear(c.localscl)
	c.ghv.clear(c)
	c.ghv.clearOff()
	c.hitby = [8]HitBy{}
	c.mhv.clear()
	for i := range c.ho {
		c.ho[i].clear()
	}
	c.mctype = MC_Hit
	c.mctime = 0
	c.counterHit = false
	c.fallTime = 0
	c.hitdefContact = false
}
func (c *Char) clear1() {
	c.anim = nil
	c.cmd = nil
	c.curFrame = nil
	c.clearState()
	c.hoIdx = -1
	c.mctype, c.mctime = MC_Hit, 0
	c.counterHit = false
	c.fallTime = 0
	c.varRangeSet(0, int32(NumVar)-1, 0)
	c.fvarRangeSet(0, int32(NumFvar)-1, 0)
	c.superDefenseMul = 1
	c.fallDefenseMul = 1
	c.customDefense = 1
	c.defenseMulDelay = false
	c.key = -1
	c.id = -1
	c.index = -1
	c.runorder = -1
	c.helperId = 0
	c.helperIndex = -1
	c.parentIndex = IErr
	c.playerNo = -1
	c.ownpal = true
	c.facing = 1
	c.keyctrl = [...]bool{false, false, false, true}
	c.player = false
	c.animPN = -1
	c.animNo = 0
	c.stchtmp = false
	c.inguarddist = false
	c.p1facing = 0
	c.pushed = false
	c.atktmp, c.hittmp, c.acttmp, c.minus = 0, 0, 0, 2
	c.winquote = -1
	c.inheritJuggle = 0
	c.immortal = false
	c.kovelocity = false
	c.preserve = 0
}

func (c *Char) clsnOverlapTrigger(box1, pid, box2 int32) bool {
	getter := sys.playerID(pid)
	// Invalid getter ID
	if getter == nil {
		return false
	}
	return c.clsnCheck(getter, box1, box2, false, true)
}

func (c *Char) copyParent(p *Char) {
	c.parentIndex = p.helperIndex
	c.name, c.key, c.size, c.teamside = p.name+"'s helper", p.key, p.size, p.teamside
	c.life, c.lifeMax, c.powerMax = p.lifeMax, p.lifeMax, p.powerMax
	if sys.maxPowerMode {
		c.power = c.powerMax
	} else {
		c.power = 0
	}
	c.dizzyPoints, c.dizzyPointsMax = p.dizzyPointsMax, p.dizzyPointsMax
	c.guardPoints, c.guardPointsMax = p.guardPointsMax, p.guardPointsMax
	c.redLife = c.lifeMax
	c.clear2()
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
func (c *Char) enemyNearClear() {
	c.enemynear[0] = c.enemynear[0][:0]
	c.enemynear[1] = c.enemynear[1][:0]
}
func (c *Char) clear2() {
	c.sysVarRangeSet(0, int32(NumSysVar)-1, 0)
	c.sysFvarRangeSet(0, int32(NumSysFvar)-1, 0)
	atk := float32(c.gi().data.attack) * c.ocd().attackRatio / 100
	c.CharSystemVar = CharSystemVar{
		bindToId:        -1,
		angleScale:      [...]float32{1, 1},
		alpha:           [...]int32{255, 0},
		width:           [...]float32{c.baseWidthFront(), c.baseWidthBack()},
		height:          [...]float32{c.baseHeightTop(), c.baseHeightBottom()},
		attackMul:       [4]float32{atk, atk, atk, atk},
		fallDefenseMul:  1,
		superDefenseMul: 1,
		customDefense:   1,
		finalDefense:    1.0,
	}
	c.widthToSizeBox()
	c.oldPos, c.interPos = c.pos, c.pos
	if c.helperIndex == 0 && c.teamside != -1 {
		if sys.roundsExisted[c.playerNo&1] > 0 {
			c.palfx.clear()
		} else {
			c.palfx = newPalFX()
		}
	} else {
		c.palfx = nil
		if c.teamside == -1 {
			c.setSCF(SCF_standby)
		}
	}
	c.aimg.timegap = -1
	c.enemyNearClear()
	c.p2enemy = c.p2enemy[:0]
	c.targets = c.targets[:0]
	c.cpucmd = -1
}
func (c *Char) clearCachedData() {
	c.anim = nil
	c.curFrame = nil
	c.hoIdx = -1
	c.mctype, c.mctime = MC_Hit, 0
	c.counterHit = false
	c.fallTime = 0
	c.superDefenseMul = 1
	c.fallDefenseMul = 1
	c.customDefense = 1
	c.defenseMulDelay = false
	c.ownpal = true
	c.animPN = -1
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
				cns, sprite = is["cns"], is["sprite"]
				anim, sound = is["anim"], is["sound"]
				for i := range gi.pal {
					gi.pal[i] = is[fmt.Sprintf("pal%v", i+1)]
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
		case fmt.Sprintf("%v.info", sys.language):
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
		case fmt.Sprintf("%v.files", sys.language):
			if lanFiles {
				files = false
				lanFiles = false
				cns, sprite = is["cns"], is["sprite"]
				anim, sound = is["anim"], is["sound"]
				for i := range gi.pal {
					gi.pal[i] = is[fmt.Sprintf("pal%v", i+1)]
				}
				for i := range fnt {
					fnt[i][0] = is[fmt.Sprintf("font%v", i)]
					fnt[i][1] = is[fmt.Sprintf("fnt_height%v", i)]
				}
			}
		case fmt.Sprintf("%v.palette ", sys.language):
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
		case fmt.Sprintf("%v.map", sys.language):
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
	gi.constants["default.ignoredefeatedenemies"] = 1
	gi.constants["input.pauseonhitpause"] = 1

	for _, s := range sys.commonConst {
		if err := LoadFile(&s, []string{def, sys.motifDir, sys.lifebar.def, "", "data/"}, func(filename string) error {
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

	// Init constants
	// Correct engine default values to character's own localcoord
	gi.data.init()
	c.size.init()

	coordRatio := float32(c.gi().localcoord[0]) / 320

	if coordRatio != 1 {
		c.size.ground.back *= coordRatio
		c.size.ground.front *= coordRatio
		c.size.air.back *= coordRatio
		c.size.air.front *= coordRatio
		c.size.height.stand *= coordRatio
		c.size.height.crouch *= coordRatio
		c.size.height.air[0] *= coordRatio
		c.size.height.air[1] *= coordRatio
		c.size.height.down *= coordRatio
		c.size.attack.dist.front *= coordRatio
		c.size.attack.dist.back *= coordRatio
		c.size.proj.attack.dist.front *= coordRatio
		c.size.proj.attack.dist.back *= coordRatio
		c.size.head.pos[0] *= coordRatio
		c.size.head.pos[1] *= coordRatio
		c.size.mid.pos[0] *= coordRatio
		c.size.mid.pos[1] *= coordRatio
		c.size.shadowoffset *= coordRatio
		c.size.draw.offset[0] *= coordRatio
		c.size.draw.offset[1] *= coordRatio
		c.size.depth *= coordRatio
		c.size.attack.depth.front *= coordRatio
		c.size.attack.depth.back *= coordRatio
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
		if err := LoadFile(&cns, []string{def, "", sys.motifDir, "data/"}, func(filename string) error {
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
						is.ReadI32("defence", &gi.data.defence)
						is.ReadI32("fall.defence_up", &gi.data.fall.defence_up)
						gi.data.fall.defence_mul = (float32(gi.data.fall.defence_up) + 100) / 100
						var i32 int32
						if is.ReadI32("liedown.time", &i32) {
							gi.data.liedown.time = Max(1, i32)
						}
						is.ReadI32("airjuggle", &gi.data.airjuggle)
						is.ReadI32("sparkno", &gi.data.sparkno)
						is.ReadI32("guard.sparkno", &gi.data.guard.sparkno)
						is.ReadI32("hitsound.channel", &gi.data.hitsound_channel)
						is.ReadI32("guardsound.channel", &gi.data.guardsound_channel)
						is.ReadI32("ko.echo", &gi.data.ko.echo)
						if is.ReadI32("volume", &i32) {
							gi.data.volume = i32/2 + 256
						}
						if is.ReadI32("volumescale", &i32) {
							gi.data.volume = i32 * 64 / 25
						}
						if _, ok := is["intpersistindex"]; ok {
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
						is.ReadF32("ground.back", &c.size.ground.back)
						is.ReadF32("ground.front", &c.size.ground.front)
						is.ReadF32("air.back", &c.size.air.back)
						is.ReadF32("air.front", &c.size.air.front)
						is.ReadF32("height", &c.size.height.stand)
						is.ReadF32("height.stand", &c.size.height.stand)
						// New height constants default to old height constant
						c.size.height.crouch = c.size.height.stand
						c.size.height.air[0] = c.size.height.stand
						c.size.height.down = c.size.height.stand
						is.ReadF32("height.crouch", &c.size.height.crouch)
						is.ReadF32("height.air", &c.size.height.air[0], &c.size.height.air[1])
						is.ReadF32("height.down", &c.size.height.down)
						is.ReadF32("attack.dist", &c.size.attack.dist.front)
						is.ReadF32("attack.dist.back", &c.size.attack.dist.back)
						is.ReadF32("proj.attack.dist", &c.size.proj.attack.dist.front)
						is.ReadF32("proj.attack.dist.back", &c.size.proj.attack.dist.back)
						is.ReadI32("proj.doscale", &c.size.proj.doscale)
						is.ReadF32("head.pos", &c.size.head.pos[0], &c.size.head.pos[1])
						is.ReadF32("mid.pos", &c.size.mid.pos[0], &c.size.mid.pos[1])
						is.ReadF32("shadowoffset", &c.size.shadowoffset)
						is.ReadF32("draw.offset",
							&c.size.draw.offset[0], &c.size.draw.offset[1])
						is.ReadF32("depth", &c.size.depth)
						is.ReadF32("attack.depth", &c.size.attack.depth.front, &c.size.attack.depth.back)
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

						// Mugen accepts these but they are not documented
						// Possible leftovers of Z axis implementation
						is.ReadF32("walk.up", &gi.velocity.walk.up) // Should be "z" but Elecbyte decided on "x"
						is.ReadF32("walk.down", &gi.velocity.walk.down)
						is.ReadF32("run.up",
							&gi.velocity.run.up[0], &gi.velocity.run.up[1])
						is.ReadF32("run.down",
							&gi.velocity.run.down[0], &gi.velocity.run.down[1]) // Z and Y?
						is.ReadF32("jump.up", &gi.velocity.jump.up) // Mugen accepts them with this syntax, but they need "x" when retrieved with const trigger
						is.ReadF32("jump.down", &gi.velocity.jump.down)
						is.ReadF32("runjump.up", &gi.velocity.runjump.up)
						is.ReadF32("runjump.down", &gi.velocity.runjump.down)
						is.ReadF32("airjump.up", &gi.velocity.airjump.up)
						is.ReadF32("airjump.down", &gi.velocity.airjump.down)
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
								gi.quotes[i], _, _ = is.getText(fmt.Sprintf("victory%v", i))
							}
						}
					}
				case fmt.Sprintf("%v.quotes", sys.language):
					if lanQuotes {
						quotes = false
						lanQuotes = false
						for i := range gi.quotes {
							if is[fmt.Sprintf("victory%v", i)] != "" {
								gi.quotes[i], _, _ = is.getText(fmt.Sprintf("victory%v", i))
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
		if LoadFile(&sprite, []string{def, "", sys.motifDir, "data/"}, func(filename string) error {
			var err error
			gi.sff, err = loadSff(filename, true)
			return err
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
		PalTex:     append([]*Texture{}, gi.sff.palList.PalTex...),
	}
	for key, value := range gi.sff.palList.PalTable {
		gi.palettedata.palList.PalTable[key] = value
	}
	for key, value := range gi.sff.palList.numcols {
		gi.palettedata.palList.numcols[key] = value
	}
	str = ""
	if len(anim) > 0 {
		if LoadFile(&anim, []string{def, "", sys.motifDir, "data/"}, func(filename string) error {
			var err error
			str, err = LoadText(filename)
			if err != nil {
				return err
			}
			return nil
		}); err != nil {
			return err
		}
	}
	for _, s := range sys.commonAir {
		if err := LoadFile(&s, []string{def, sys.motifDir, sys.lifebar.def, "", "data/"}, func(filename string) error {
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
	lines, i = SplitAndTrim(str, "\n"), 0
	gi.anim = ReadAnimationTable(gi.sff, &gi.palettedata.palList, lines, &i)
	if len(sound) > 0 {
		if LoadFile(&sound, []string{def, "", sys.motifDir, "data/"}, func(filename string) error {
			var err error
			gi.snd, err = LoadSnd(filename)
			return err
		}); err != nil {
			return err
		}
	} else {
		gi.snd = newSnd()
	}
	if c.teamside != -1 {
		// Get fonts from preloaded data
		gi.fnt = sys.sel.GetChar(c.selectNo).fnt
	} else {
		// Load fonts for AttachedChar
		for i, f := range fnt {
			if len(f[0]) > 0 {
				LoadFile(&f[0], []string{def, sys.motifDir, "", "data/", "font/"}, func(filename string) error {
					var err error
					var height int32 = -1
					if len(f[1]) > 0 {
						height = Atoi(f[1])
					}
					if gi.fnt[i], err = loadFnt(filename, height); err != nil {
						sys.errLog.Printf("failed to load %v (char font): %v", filename, err)
					}
					return nil
				})
			}
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
			var f *os.File
			var err error
			if LoadFile(&gi.pal[i], []string{gi.def, "", sys.motifDir, "data/"}, func(file string) error {
				f, err = os.Open(file)
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
	gi.drawpalno = gi.palno
	starti := gi.palno - 1
	if !gi.palExist[starti] {
		starti %= 6
	}
	i := starti
	for {
		if gi.palExist[i] {
			j := 0
			for ; j < len(sys.chars); j++ {
				if j != c.playerNo && len(sys.chars[j]) > 0 &&
					sys.cgi[j].def == gi.def && sys.cgi[j].drawpalno == i+1 {
					break
				}
			}
			if j >= len(sys.chars) {
				gi.drawpalno = i + 1
				if !gi.palExist[gi.palno-1] {
					gi.palno = gi.drawpalno
				}
				break
			}
		}
		i++
		if i >= MaxPalNo {
			i = 0
		}
		if i == starti {
			if !gi.palExist[gi.palno-1] {
				i := 0
				for ; i < len(gi.palExist); i++ {
					if gi.palExist[i] {
						gi.palno, gi.drawpalno = int32(i+1), int32(i+1)
						break
					}
				}
				if i >= len(gi.palExist) {
					gi.palno, gi.palExist[0] = 1, true
					gi.palSelectable[0] = true
				}
			}
			break
		}
	}
	gi.remappedpal = [...]int32{1, gi.palno}
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
	c.hitdef.clear(c.localscl)
}

func (c *Char) changeAnimEx(animNo int32, playerNo int, ffx string, alt bool) {
	if a := sys.chars[playerNo][0].getAnim(animNo, ffx, false); a != nil {
		c.anim = a
		c.anim.remap = c.remapSpr
		c.animPN = c.playerNo
		c.prevAnimNo = c.animNo
		c.animNo = animNo
		// If using ChangeAnim2, the animation is changed but the sff is kept
		if alt {
			c.animPN = playerNo
			a.sff = sys.cgi[c.playerNo].sff
			a.palettedata = &sys.cgi[c.playerNo].palettedata.palList
			// Fix palette if anim doesn't belong to char and sff header version is 1.x
		} else if c.playerNo != playerNo && c.anim.sff.header.Ver0 == 1 {
			di := c.anim.palettedata.PalTable[[...]int16{1, 1}]
			spr := c.anim.sff.GetSprite(0, 0)
			if spr != nil {
				c.anim.palettedata.Remap(spr.palidx, di)
			}
			spr = c.anim.sff.GetSprite(9000, 0)
			if spr != nil {
				c.anim.palettedata.Remap(spr.palidx, di)
			}
		}
		// Update animation local scale
		c.animlocalscl = 320 / sys.chars[c.animPN][0].localcoord
		// Clsn scale depends on the animation owner's scale, so it must be updated
		c.updateClsnBaseScale()
		if c.hitPause() {
			c.curFrame = a.CurrentFrame()
		}
	}
}
func (c *Char) changeAnim(animNo int32, playerNo int, ffx string) {
	if animNo < 0 && animNo != -2 {
		// MUGEN 1.1 exports a warning message when attempting to change anim to a negative value through ChangeAnim SCTRL,
		// then sets the character animation to "0". Ikemen GO uses "-2" as a no-sprite/invisible anim, so we make
		// an exception here
		sys.appendToConsole(c.warn() + fmt.Sprintf("attempted change to negative anim (different from -2)"))
		animNo = 0
	}
	c.changeAnimEx(animNo, playerNo, ffx, false)
}
func (c *Char) changeAnim2(animNo int32, playerNo int, ffx string) {
	if animNo < 0 && animNo != -2 {
		sys.appendToConsole(c.warn() + fmt.Sprintf("attempted change to negative anim (different from -2)"))
		animNo = 0
	}
	c.changeAnimEx(animNo, playerNo, ffx, true)
}
func (c *Char) setAnimElem(e int32) {
	if c.anim != nil {
		c.anim.SetAnimElem(e)
		c.curFrame = c.anim.CurrentFrame()
		if int(e) < 0 {
			sys.appendToConsole(c.warn() + fmt.Sprintf("changed to negative animelem"))
		} else if int(e) > len(c.anim.frames) {
			sys.appendToConsole(c.warn() + fmt.Sprintf("changed to invalid animelem %v within action %v", e, c.animNo))
		}
	}
}
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
}
func (c *Char) unsetSCF(scf SystemCharFlag) {
	c.systemFlag &^= scf
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
func (c *Char) parent() *Char {
	if c.parentIndex == IErr {
		sys.appendToConsole(c.warn() + "has no parent")
		return nil
	}
	if c.parentIndex < 0 {
		sys.appendToConsole(c.warn() + "parent has been already destroyed")
		if !sys.ignoreMostErrors {
			sys.errLog.Println(c.name + " parent has been already destroyed")
		}
	}
	return sys.chars[c.playerNo][Abs(c.parentIndex)]
}
func (c *Char) root() *Char {
	if c.helperIndex == 0 {
		sys.appendToConsole(c.warn() + "has no root")
		return nil
	}
	return sys.chars[c.playerNo][0]
}
func (c *Char) helper(id int32) *Char {
	for _, h := range sys.chars[c.playerNo][1:] {
		if !h.csf(CSF_destroy) && (id <= 0 || id == h.helperId) {
			return h
		}
	}
	sys.appendToConsole(c.warn() + fmt.Sprintf("has no helper: %v", id))
	return nil
}
func (c *Char) getPlayerHelperIndex(n int32, ex bool) *Char {
	if n <= 0 {
		return c
	}
	return sys.charList.getHelperIndex(c, n, ex)
}
func (c *Char) helperByIndexExist(id BytecodeValue) BytecodeValue {
	if id.IsSF() {
		return BytecodeSF()
	}
	return BytecodeBool(c.getPlayerHelperIndex(id.ToI(), true) != nil)
}
func (c *Char) target(id int32) *Char {
	for _, tid := range c.targets {
		if t := sys.playerID(tid); t != nil && (id < 0 || id == t.ghv.hitid) {
			return t
		}
	}
	if id != -1 {
		sys.appendToConsole(c.warn() + fmt.Sprintf("has no target: %v", id))
	}
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
func (c *Char) partnerV2(n int32) *Char {
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
	if n < 0 || n >= c.numEnemy() {
		sys.appendToConsole(c.warn() + fmt.Sprintf("has no enemy: %v", n))
		return nil
	}
	if c.teamside == -1 {
		return sys.chars[n][0]
	}
	for i := n*2 + int32(^c.playerNo&1); i < sys.numSimul[^c.playerNo&1]*2; i += 2 {
		if !sys.chars[i][0].scf(SCF_standby) && !sys.chars[i][0].scf(SCF_disabled) {
			return sys.chars[i][0]
		}
	}
	//return sys.chars[n*2+int32(^c.playerNo&1)][0]
	return nil
}
func (c *Char) enemyNear(n int32) *Char {
	return sys.charList.enemyNear(c, n, false, c.gi().constants["default.ignoredefeatedenemies"] > 0, false)
}
func (c *Char) p2() *Char {
	p2 := sys.charList.enemyNear(c, 0, true, true, false)
	if p2 != nil && p2.scf(SCF_ko) && p2.scf(SCF_over) {
		return nil
	}
	return p2
}
func (c *Char) aiLevel() float32 {
	if c.helperIndex != 0 && c.gi().mugenver[0] == 1 {
		return 0
	}
	return sys.com[c.playerNo]
}
func (c *Char) alive() bool {
	return !c.scf(SCF_ko)
}
func (c *Char) animElemNo(time int32) BytecodeValue {
	if c.anim != nil && time >= -c.anim.sumtime {
		return BytecodeInt(c.anim.AnimElemNo(time))
	}
	return BytecodeSF()
}
func (c *Char) animElemTime(e int32) BytecodeValue {
	if e >= 1 && c.anim != nil && int(e) <= len(c.anim.frames) {
		return BytecodeInt(c.anim.AnimElemTime(e))
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
	return c.backEdgeDist() - c.edge[1] - offset
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
func (c *Char) botBoundDist() float32 {
	return sys.zmax/c.localscl - c.pos[2]
}
func (c *Char) canRecover() bool {
	return c.ghv.fall.recover && c.fallTime >= c.ghv.fall.recovertime
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
	cl := c.cmd[pn].At(i)
	// Check if any command with that name is buffered
	for _, c := range cl {
		if c.curbuftime > 0 {
			return true
		}
	}
	// AI cheating for commands longer than 1 button
	if c.key < 0 && len(cl) > 0 {
		if c.helperIndex != 0 || len(cl[0].cmd) > 1 || len(cl[0].cmd[0].key) > 1 ||
			int(Btoi(cl[0].cmd[0].slash)) != len(cl[0].hold) {
			if i == int(c.cpucmd) {
				return true
			}
		}
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
	ok := false
	// Assert the command in every command list
	for i := range c.cmd {
		ok = c.cmd[i].Assert(name, time) || ok
	}
	if !ok {
		sys.appendToConsole(c.warn() + fmt.Sprintf("attempted to assert an invalid command"))
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
	return c.frontEdgeDist() - c.edge[0] - offset
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
func (c *Char) getPower() int32 {
	if sys.powerShare[c.playerNo&1] && c.teamside != -1 {
		return sys.chars[c.playerNo&1][0].power
	}
	return sys.chars[c.playerNo][0].power
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
func (c *Char) isHelper(hid BytecodeValue) BytecodeValue {
	if hid.IsSF() {
		return BytecodeSF()
	}
	id := hid.ToI()
	return BytecodeBool(c.helperIndex != 0 && (id == math.MinInt32 || c.helperId == id))
}
func (c *Char) isHost() bool {
	return sys.netInput != nil && sys.netInput.host
}
func (c *Char) jugglePoints(hid BytecodeValue) BytecodeValue {
	if hid.IsSF() {
		return BytecodeSF()
	}
	tid := hid.ToI()
	max := c.gi().data.airjuggle
	jp := max // If no target is found it returns the char's maximum juggle points
	for _, ct := range c.targets {
		if ct >= 0 {
			t := sys.playerID(ct)
			if t != nil && t.id == tid {
				jp = t.ghv.getJuggle(c.id, max)
			}
		}
	}
	return BytecodeInt(jp)
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

// Mugen version trigger
func (c *Char) mugenVersionF() float32 {
	// Here the version is always checked directly in the character instead of the working state
	// This is because in a custom state this trigger will be used to know the enemy's version rather than our own
	if c.gi().ikemenver[0] != 0 || c.gi().ikemenver[1] != 0 {
		return 1.1
	} else if c.gi().mugenver[0] == 1 && c.gi().mugenver[1] == 1 {
		return 1.1
	} else if c.gi().mugenver[0] == 1 && c.gi().mugenver[1] == 0 {
		return 1.0
	} else if c.gi().mugenver[0] != 1 {
		return 0.5 // Arbitrary value
	} else {
		return 0
	}
}

func (c *Char) numEnemy() int32 {
	var n int32
	if c.teamside == -1 {
		for i := 0; i < int(sys.numSimul[0]+sys.numSimul[1]); i++ {
			if len(sys.chars[i]) > 0 && !sys.chars[i][0].scf(SCF_standby) && !sys.chars[i][0].scf(SCF_disabled) {
				n += 1
			}
		}
		return n
	}
	for i := ^c.playerNo & 1; i < int(sys.numSimul[^c.playerNo&1]*2); i += 2 {
		if len(sys.chars[i]) > 0 && !sys.chars[i][0].scf(SCF_standby) && !sys.chars[i][0].scf(SCF_disabled) {
			n += 1
		}
	}
	return n
}
func (c *Char) numPlayer() int32 {
	n := int32(0)
	for i := 0; i < len(sys.chars)-1; i++ {
		//&& !sys.chars[i][0].scf(SCF_standby)
		if len(sys.chars[i]) > 0 && !sys.chars[i][0].scf(SCF_disabled) {
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
	for _, e := range sys.explods[c.playerNo] {
		if e.matchId(id, c.id) {
			n++
		}
	}
	return BytecodeInt(n)
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
			case OC_ex2_explodvar_animelem:
				v = BytecodeInt(e.anim.current + 1)
			case OC_ex2_explodvar_pos_x:
				v = BytecodeFloat(e.pos[0] + e.offset[0] + e.relativePos[0] + e.interpolate_pos[0])
			case OC_ex2_explodvar_pos_y:
				v = BytecodeFloat(e.pos[1] + e.offset[1] + e.relativePos[1] + e.interpolate_pos[1])
			case OC_ex2_explodvar_pos_z:
				v = BytecodeFloat(e.pos[2] + e.offset[2] + e.relativePos[2] + e.interpolate_pos[2])
			case OC_ex2_explodvar_scale_x:
				v = BytecodeFloat(e.scale[0] * e.interpolate_scale[0])
			case OC_ex2_explodvar_scale_y:
				v = BytecodeFloat(e.scale[1] * e.interpolate_scale[1])
			case OC_ex2_explodvar_angle:
				v = BytecodeFloat(e.anglerot[0] + e.interpolate_angle[0])
			case OC_ex2_explodvar_angle_x:
				v = BytecodeFloat(e.anglerot[1] + e.interpolate_angle[1])
			case OC_ex2_explodvar_angle_y:
				v = BytecodeFloat(e.anglerot[2] + e.interpolate_angle[2])
			case OC_ex2_explodvar_vel_x:
				v = BytecodeFloat(e.velocity[0])
			case OC_ex2_explodvar_vel_y:
				v = BytecodeFloat(e.velocity[1])
			case OC_ex2_explodvar_vel_z:
				v = BytecodeFloat(e.velocity[2])
			case OC_ex2_explodvar_removetime:
				v = BytecodeInt(e.removetime)
			case OC_ex2_explodvar_pausemovetime:
				v = BytecodeInt(e.pausemovetime)
			case OC_ex2_explodvar_sprpriority:
				v = BytecodeInt(e.sprpriority)
			case OC_ex2_explodvar_layerno:
				v = BytecodeInt(e.layerno)
			case OC_ex2_explodvar_id:
				v = BytecodeInt(e.id)
			case OC_ex2_explodvar_bindtime:
				v = BytecodeInt(e.bindtime)
			case OC_ex2_explodvar_facing:
				v = BytecodeInt(int32(e.facing))
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
			case OC_ex2_projvar_projremove:
				v = BytecodeBool(p.remove)
			case OC_ex2_projvar_projremovetime:
				v = BytecodeInt(p.removetime)
			case OC_ex2_projvar_projshadow_r:
				v = BytecodeInt(p.shadow[0])
			case OC_ex2_projvar_projshadow_g:
				v = BytecodeInt(p.shadow[1])
			case OC_ex2_projvar_projshadow_b:
				v = BytecodeInt(p.shadow[2])
			case OC_ex2_projvar_projmisstime:
				v = BytecodeInt(p.curmisstime)
			case OC_ex2_projvar_projhits:
				v = BytecodeInt(p.hits)
			case OC_ex2_projvar_projhitsmax:
				v = BytecodeInt(p.totalhits)
			case OC_ex2_projvar_projpriority:
				v = BytecodeInt(p.priority)
			case OC_ex2_projvar_projhitanim:
				v = BytecodeInt(p.hitanim)
			case OC_ex2_projvar_projremanim:
				v = BytecodeInt(p.remanim)
			case OC_ex2_projvar_projcancelanim:
				v = BytecodeInt(p.cancelanim)
			case OC_ex2_projvar_vel_x:
				v = BytecodeFloat(p.velocity[0] * p.localscl / oc.localscl)
			case OC_ex2_projvar_vel_y:
				v = BytecodeFloat(p.velocity[1] * p.localscl / oc.localscl)
			case OC_ex2_projvar_velmul_x:
				v = BytecodeFloat(p.velmul[0])
			case OC_ex2_projvar_velmul_y:
				v = BytecodeFloat(p.velmul[1])
			case OC_ex2_projvar_remvelocity_x:
				v = BytecodeFloat(p.remvelocity[0] * p.localscl / oc.localscl)
			case OC_ex2_projvar_remvelocity_y:
				v = BytecodeFloat(p.remvelocity[1] * p.localscl / oc.localscl)
			case OC_ex2_projvar_accel_x:
				v = BytecodeFloat(p.accel[0] * p.localscl)
			case OC_ex2_projvar_accel_y:
				v = BytecodeFloat(p.accel[1] * p.localscl)
			case OC_ex2_projvar_projscale_x:
				v = BytecodeFloat(p.scale[0])
			case OC_ex2_projvar_projscale_y:
				v = BytecodeFloat(p.scale[1])
			case OC_ex2_projvar_projangle:
				v = BytecodeFloat(p.angle)
			case OC_ex2_projvar_pos_x:
				v = BytecodeFloat((p.pos[0]*p.localscl - sys.cam.Pos[0]) / oc.localscl)
			case OC_ex2_projvar_pos_y:
				v = BytecodeFloat(p.pos[1] * p.localscl / oc.localscl)
			case OC_ex2_projvar_pos_z:
				v = BytecodeFloat(p.pos[2] * p.localscl / oc.localscl)
			case OC_ex2_projvar_projsprpriority:
				v = BytecodeInt(p.sprpriority)
			case OC_ex2_projvar_projlayerno:
				v = BytecodeInt(p.layerno)
			case OC_ex2_projvar_projstagebound:
				v = BytecodeInt(int32(float32(p.stagebound) * p.localscl / oc.localscl))
			case OC_ex2_projvar_projedgebound:
				v = BytecodeInt(int32(float32(p.edgebound) * p.localscl / oc.localscl))
			case OC_ex2_projvar_lowbound:
				v = BytecodeInt(int32(float32(p.heightbound[0]) * p.localscl / oc.localscl))
			case OC_ex2_projvar_highbound:
				v = BytecodeInt(int32(float32(p.heightbound[1]) * p.localscl / oc.localscl))
			case OC_ex2_projvar_projanim:
				v = BytecodeInt(p.anim)
			case OC_ex2_projvar_animelem:
				v = BytecodeInt(p.ani.current + 1)
			case OC_ex2_projvar_supermovetime:
				v = BytecodeInt(p.supermovetime)
			case OC_ex2_projvar_pausemovetime:
				v = BytecodeInt(p.pausemovetime)
			case OC_ex2_projvar_projid:
				v = BytecodeInt(int32(p.id))
			case OC_ex2_projvar_teamside:
				v = BytecodeInt(int32(p.hitdef.teamside))
			case OC_ex2_projvar_guardflag:
				v = BytecodeBool(p.hitdef.guardflag&fl != 0)
			case OC_ex2_projvar_hitflag:
				v = BytecodeBool(p.hitdef.hitflag&fl != 0)
			}
			break
		}
	}
	return v
}
func (c *Char) numHelper(hid BytecodeValue) BytecodeValue {
	if hid.IsSF() {
		return BytecodeSF()
	}
	var id, n int32 = hid.ToI(), 0
	for _, h := range sys.chars[c.playerNo][1:] {
		if !h.csf(CSF_destroy) && (id <= 0 || h.helperId == id) {
			n++
		}
	}
	return BytecodeInt(n)
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
	for _, p := range sys.projs[c.playerNo] {
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
	for _, p := range sys.projs[c.playerNo] {
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
func (c *Char) palno() int32 {
	if c.helperIndex != 0 && c.gi().mugenver[0] != 1 {
		return 1
	}
	return c.gi().palno
}
func (c *Char) pauseTime() int32 {
	var p int32
	if sys.super > 0 && c.prevSuperMovetime == 0 {
		p = sys.super
	}
	if sys.pause > 0 && c.prevPauseMovetime == 0 && p < sys.pause {
		p = sys.pause
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
	if (id > 0 && id != c.gi().pcid) || c.helperIndex > 0 {
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
func (c *Char) roundType() int32 {
	if sys.roundType[0] == RT_Final {
		return 3
	} else if sys.roundType[c.playerNo&1] == RT_Deciding {
		return 2
	} else if sys.roundType[^c.playerNo&1] == RT_Deciding {
		return 1
	}
	return 0
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
func (c *Char) topBoundDist() float32 {
	return sys.zmin/c.localscl - c.pos[2]
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
	var s *Sound
	if ffx == "" || ffx == "s" {
		if c.gi().snd != nil {
			s = c.gi().snd.Get([...]int32{g, n})
		}
	} else {
		if sys.ffx[ffx] != nil && sys.ffx[ffx].fsnd != nil {
			s = sys.ffx[ffx].fsnd.Get([...]int32{g, n})
		}
	}
	if s == nil {
		if log {
			if ffx != "" {
				sys.appendToConsole(c.warn() + fmt.Sprintf("sound %v %v,%v doesn't exist", strings.ToUpper(ffx), g, n))
			} else {
				sys.appendToConsole(c.warn() + fmt.Sprintf("sound %v,%v doesn't exist", g, n))
			}
		}
		if !sys.ignoreMostErrors {
			str := "Sound doesn't exist: "
			if ffx != "" {
				str += ffx + ":"
			} else {
				str += fmt.Sprintf("P%v:", c.playerNo+1)
			}
			sys.errLog.Printf("%v%v,%v\n", str, g, n)
			return
		}
	}
	crun := c
	if c.inheritChannels == 1 && c.parent() != nil {
		crun = c.parent()
	} else if c.inheritChannels == 2 && c.root() != nil {
		crun = c.root()
	}
	if ch := crun.soundChannels.New(chNo, lowpriority, priority); ch != nil {
		ch.Play(s, loopCount, freqmul, loopstart, loopend, startposition)
		vol = Clamp(vol, -25600, 25600)
		//if c.gi().mugenver[0] == 1 {
		if ffx != "" {
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

func (c *Char) turn() {
	if c.helperIndex == 0 {
		if e := sys.charList.enemyNear(c, 0, true, true, false); c.rdDistX(e, c).ToF() < 0 && !e.asf(ASF_noturntarget) {
			switch c.ss.stateType {
			case ST_S:
				if c.animNo != 5 {
					c.changeAnimEx(5, c.playerNo, "", false)
				}
			case ST_C:
				if c.animNo != 6 {
					c.changeAnimEx(6, c.playerNo, "", false)
				}
			}
			c.setFacing(-c.facing)
		}
	}
}
func (c *Char) stateChange1(no int32, pn int) bool {
	if sys.changeStateNest > 2500 {
		sys.appendToConsole(c.warn() + fmt.Sprintf("state machine stuck in loop (stopped after 2500 loops): %v -> %v -> %v", c.ss.prevno, c.ss.no, no))
		sys.errLog.Printf("2500 loops: %v, %v -> %v -> %v\n", c.name, c.ss.prevno, c.ss.no, no)
		return false
	}
	c.ss.no, c.ss.prevno, c.ss.time = Max(0, no), c.ss.no, 0
	//if c.ss.sb.playerNo != c.playerNo && pn != c.ss.sb.playerNo {
	//	c.enemyExplodsRemove(c.ss.sb.playerNo)
	//}
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
		c.ghv.fall.xvelocity *= lsRatio
		c.ghv.fall.yvelocity *= lsRatio
		c.ghv.fall.zvelocity *= lsRatio
		c.ghv.xaccel *= lsRatio
		c.ghv.yaccel *= lsRatio
		c.ghv.zaccel *= lsRatio

		c.width[0] *= lsRatio
		c.width[1] *= lsRatio
		c.edge[0] *= lsRatio
		c.edge[1] *= lsRatio
		c.height[0] *= lsRatio
		c.height[1] *= lsRatio
		c.widthToSizeBox()

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
	if c.minus <= 0 && c.scf(SCF_ctrl) && sys.roundState() <= 2 &&
		(c.ss.stateType == ST_S || c.ss.stateType == ST_C) && !c.asf(ASF_noautoturn) && sys.stage.autoturn {
		c.turn()
	}
	if anim != -1 {
		c.changeAnim(anim, c.playerNo, ffx)
	}
	if ctrl >= 0 {
		c.setCtrl(ctrl != 0)
	}
	if c.stateChange1(no, pn) && sys.changeStateNest == 0 && c.minus == 0 {
		for c.stchtmp && sys.changeStateNest < 2500 {
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
		c.exitTarget(true)
		c.receivedDmg = 0
		c.receivedHits = 0
		if c.player {
			sys.charList.p2enemyDelete(c)
		}
		for _, tid := range c.targets {
			if t := sys.playerID(tid); t != nil {
				if t.bindToId == c.id {
					if t.ss.moveType == MT_H {
						t.selfState(5050, -1, -1, -1, "")
					}
				}
				t.gethitBindClear()
				t.ghv.dropId(c.id)
			}
		}
		if c.parentIndex >= 0 {
			if p := c.parent(); p != nil {
				for i, ch := range p.children {
					if ch == c {
						p.children[i] = nil
					}
				}
			}
		}
		for _, ch := range c.children {
			if ch != nil {
				ch.parentIndex *= -1
			}
		}
		c.children = c.children[:0]
		sys.charList.delete(c)
		c.helperIndex = -1
		c.setCSF(CSF_destroy)
	}
}

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

func (c *Char) newHelper() (h *Char) {
	// If any existing helper entries are valid for overwriting, use that one
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
		if i >= sys.helperMax {
			return
		}
		h = newChar(c.playerNo, i)
		sys.chars[c.playerNo] = append(sys.chars[c.playerNo], h)
	}
	h.id = sys.newCharId()
	h.helperId = 0
	h.ownpal = false
	h.copyParent(c)
	c.addChild(h)
	sys.charList.add(h)
	return
}

func (c *Char) helperInit(h *Char, st int32, pt PosType, x, y, z float32,
	facing int32, rp [2]int32, extmap bool) {
	p := c.helperPos(pt, [...]float32{x, y, z}, facing, &h.facing, h.localscl, false)
	h.setX(p[0])
	h.setY(p[1])
	h.setZ(p[2])
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
		if p2 := sys.charList.enemyNear(c, 0, true, true, false); p2 != nil {
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
		p[2] = c.pos[2]
		*dstFacing *= c.facing
	case PT_Left:
		p[0] = c.leftEdge()*(c.localscl/localscl) + pos[0]
		p[1] = pos[1]
		if isProj {
			*dstFacing *= c.facing
		}
	case PT_Right:
		p[0] = c.rightEdge()*(c.localscl/localscl) + pos[0]
		p[1] = pos[1]
		if isProj {
			*dstFacing *= c.facing
		}
		p[2] = c.pos[2]
	case PT_None:
		p = [3]float32{pos[0], pos[1], c.pos[2]}
		if isProj {
			*dstFacing *= c.facing
		}
	}
	return
}

func (c *Char) newExplod() (*Explod, int) {
	explinit := func(expl *Explod) *Explod {
		expl.clear()
		// Explod defaults
		expl.id = -1
		expl.playerId = c.id
		expl.layerno = c.layerNo
		expl.palfx = c.getPalfx()
		expl.palfxdef = PalFXDef{color: 1, hue: 0, mul: [...]int32{256, 256, 256}}
		if c.stWgi().mugenver[0] == 1 && c.stWgi().mugenver[1] == 1 && c.stWgi().ikemenver[0] == 0 && c.stWgi().ikemenver[1] == 0 {
			expl.projection = Projection_Perspective
		} else {
			expl.projection = Projection_Orthographic
		}
		return expl
	}
	// Reuse free explod slots
	for i := range sys.explods[c.playerNo] {
		if sys.explods[c.playerNo][i].id == IErr {
			return explinit(&sys.explods[c.playerNo][i]), i
		}
	}
	// Otherwise append it
	i := len(sys.explods[c.playerNo])
	if i < sys.explodMax {
		sys.explods[c.playerNo] = append(sys.explods[c.playerNo], Explod{})
		return explinit(&sys.explods[c.playerNo][i]), i
	}
	return nil, -1
}
func (c *Char) getExplods(id int32) (expls []*Explod) {
	for i, e := range sys.explods[c.playerNo] {
		if e.matchId(id, c.id) {
			expls = append(expls, &sys.explods[c.playerNo][i])
		}
	}
	return
}
func (c *Char) insertExplodEx(i int, rp [2]int32) {
	e := &sys.explods[c.playerNo][i]
	if e.anim == nil {
		e.id = IErr
		return
	}
	e.anim.UpdateSprite()
	if e.ownpal {
		if e.anim.sff != sys.ffx["f"].fsff {
			remap := make([]int, len(e.palfx.remap))
			copy(remap, e.palfx.remap)
			e.palfx = newPalFX()
			e.palfx.remap = remap
			e.palfx.PalFXDef = e.palfxdef
			c.forceRemapPal(e.palfx, rp)
		} else {
			e.palfx = newPalFX()
			e.palfx.PalFXDef = e.palfxdef
			e.palfx.remap = nil
		}
	}
	if e.layerno > 0 {
		td := &sys.explodsLayer1[c.playerNo]
		for ii, te := range *td {
			if te < 0 {
				(*td)[ii] = i
				return
			}
		}
		*td = append(*td, i)
	} else if e.layerno < 0 {
		td := &sys.explodsLayerN1[c.playerNo]
		for ii, te := range *td {
			if te < 0 {
				(*td)[ii] = i
				return
			}
		}
		*td = append(*td, i)
	} else {
		ed := &sys.explodsLayer0[c.playerNo]
		for ii, ex := range *ed {
			pid := sys.explods[c.playerNo][ex].playerId
			if pid >= c.id && (pid > c.id || ex < i) {
				*ed = append(*ed, 0)
				copy((*ed)[ii+1:], (*ed)[ii:])
				(*ed)[ii] = i
				return
			}
		}
		*ed = append(*ed, i)
	}
}
func (c *Char) insertExplod(i int) {
	c.insertExplodEx(i, [...]int32{-1, 0})
}
func (c *Char) explodBindTime(id, time int32) {
	for i, e := range sys.explods[c.playerNo] {
		if e.matchId(id, c.id) {
			sys.explods[c.playerNo][i].bindtime = time
		}
	}
}
func (c *Char) removeExplod(id, idx int32) {

	remove := func(drawlist *[]int, drop bool) {
		n := int32(0)
		for i := len(*drawlist) - 1; i >= 0; i-- {
			ei := (*drawlist)[i]
			if ei >= 0 && sys.explods[c.playerNo][ei].matchId(id, c.id) {
				if idx == n || idx < 0 {
					sys.explods[c.playerNo][ei].id = IErr
					if drop {
						*drawlist = append((*drawlist)[:i], (*drawlist)[i+1:]...)
					} else {
						(*drawlist)[i] = -1
					}
					if idx == n {
						break
					}
				}
				n++
			}
		}
	}
	remove(&sys.explodsLayerN1[c.playerNo], true)
	remove(&sys.explodsLayer0[c.playerNo], true)
	remove(&sys.explodsLayer1[c.playerNo], false)
}
func (c *Char) enemyExplodsRemove(en int) {
	remove := func(drawlist *[]int, drop bool) {
		for i := len(*drawlist) - 1; i >= 0; i-- {
			ei := (*drawlist)[i]
			if ei >= 0 && sys.explods[en][ei].bindtime != 0 &&
				sys.explods[en][ei].bindId == c.id {
				sys.explods[en][ei].id = IErr
				if drop {
					*drawlist = append((*drawlist)[:i], (*drawlist)[i+1:]...)
				} else {
					(*drawlist)[i] = -1
				}
			}
		}
	}
	remove(&sys.explodsLayerN1[en], true)
	remove(&sys.explodsLayer0[en], true)
	remove(&sys.explodsLayer1[en], false)
}
func (c *Char) getAnim(n int32, ffx string, fx bool) (a *Animation) {
	if n == -2 {
		return &Animation{}
	}
	if n == -1 {
		return nil
	}
	if ffx != "" && ffx != "s" {
		if sys.ffx[ffx] != nil && sys.ffx[ffx].fat != nil {
			a = sys.ffx[ffx].fat.get(n)
		}
	} else {
		a = c.gi().anim.get(n)
	}
	if a == nil {
		if fx {
			if ffx != "" && ffx != "s" {
				sys.appendToConsole(c.warn() + fmt.Sprintf("called invalid action %v %v", strings.ToUpper(ffx), n))
			} else {
				sys.appendToConsole(c.warn() + fmt.Sprintf("called invalid action %v", n))
			}
		} else {
			if ffx != "" && ffx != "s" {
				sys.appendToConsole(c.warn() + fmt.Sprintf("changed to invalid action %v %v", strings.ToUpper(ffx), n))
			} else {
				sys.appendToConsole(c.warn() + fmt.Sprintf("changed to invalid action %v", n))
			}
		}
		if !sys.ignoreMostErrors {
			str := "Invalid action: "
			if ffx != "" && ffx != "s" {
				str += strings.ToUpper(ffx) + ":"
			} else {
				str += fmt.Sprintf("P%v:", c.playerNo+1)
			}
			sys.errLog.Printf("%v%v\n", str, n)
		}
	} else if ffx != "" && ffx != "s" {
		a.start_scale[0] /= c.localscl
		a.start_scale[1] /= c.localscl
	}
	return
}

// Position functions
func (c *Char) setPosX(x float32) {
	if c.pos[0] != x {
		c.pos[0] = x
		// We do this because Mugen is very sensitive to enemy position changes
		// Perhaps what it does is only calculate who "enemynear" is when the trigger is called?
		// "P2" enemy reference is less sensitive than this however
		c.enemyNearClear()
		if c.player {
			for i := ^c.playerNo & 1; i < len(sys.chars); i += 2 {
				for j := range sys.chars[i] {
					sys.chars[i][j].enemyNearClear()
				}
			}
		}
	}
}

func (c *Char) setPosY(y float32) { // These functions mostly exist right now so we don't forget to use setPosX for X
	c.pos[1] = y
}

func (c *Char) setPosZ(z float32) {
	c.pos[2] = z
}

func (c *Char) posReset() {
	if c.teamside == -1 {
		c.facing = 1
		c.setX(0)
		c.setY(0)
		c.setZ(0)
	} else {
		c.facing = float32(sys.stage.p[c.playerNo].facing)
		c.setX((float32(sys.stage.p[c.playerNo].startx) * sys.stage.localscl) / c.localscl)
		c.setY(float32(sys.stage.p[c.playerNo].starty) * sys.stage.localscl / c.localscl)
		c.setZ(float32(sys.stage.p[c.playerNo].startz) * sys.stage.localscl / c.localscl)
	}
	c.vel[0] = 0
	c.vel[1] = 0
	c.vel[2] = 0
}

func (c *Char) setX(x float32) {
	c.oldPos[0], c.interPos[0] = x, x
	c.setPosX(x)
}
func (c *Char) setY(y float32) {
	c.oldPos[1], c.interPos[1] = y, y
	c.setPosY(y)
}
func (c *Char) setZ(z float32) {
	c.oldPos[2], c.interPos[2] = z, z
	c.setPosZ(z)
}

func (c *Char) addX(x float32) {
	c.setX(c.pos[0] + c.facing*x)
}
func (c *Char) addY(y float32) {
	c.setY(c.pos[1] + y)
}
func (c *Char) addZ(z float32) {
	c.setZ(c.pos[2] + z)
}

func (c *Char) shadXOff(xv float32, isReflect bool) {
	if !isReflect {
		c.shadowOffset[0] = xv
	} else {
		c.reflectOffset[0] = xv
	}
}
func (c *Char) shadYOff(yv float32, isReflect bool) {
	if !isReflect {
		c.shadowOffset[1] = yv
	} else {
		c.reflectOffset[1] = yv
	}
}

// --------------------

func (c *Char) hitAdd(h int32) {
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
		// in mugen HitAdd increases combo count even without targets
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
func (c *Char) newProj() *Projectile {
	for i := c.gi().projidcount; i < len(sys.projs[c.playerNo]); i++ {
		if sys.projs[c.playerNo][i].id < 0 {
			sys.projs[c.playerNo][i].clear()
			// Projectile defaults
			sys.projs[c.playerNo][i].id = 0
			sys.projs[c.playerNo][i].layerno = c.layerNo
			sys.projs[c.playerNo][i].palfx = c.getPalfx()
			c.gi().projidcount = i
			return &sys.projs[c.playerNo][i]
		}
	}
	if i := len(sys.projs[c.playerNo]); i < sys.playerProjectileMax {
		sys.projs[c.playerNo] = append(sys.projs[c.playerNo], *newProjectile())
		p := &sys.projs[c.playerNo][i]
		p.id, p.palfx = 0, c.getPalfx()
		return p
	}
	return nil
}

func (c *Char) projInit(p *Projectile, pt PosType, x, y, z float32,
	op bool, rpg, rpn int32, clsnscale bool) {
	pos := c.helperPos(pt, [...]float32{x, y, z}, 1, &p.facing, p.localscl, true)
	p.setPos([...]float32{pos[0], pos[1], pos[2]})
	p.parentAttackmul = c.attackMul
	if p.anim < -1 {
		p.anim = 0
	}
	p.ani = c.getAnim(p.anim, p.anim_ffx, true)
	if p.ani == nil && c.anim != nil {
		p.ani = &Animation{}
		*p.ani = *c.anim
		p.ani.SetAnimElem(1)
		p.anim = c.animNo
	}
	if p.ani != nil {
		p.ani.UpdateSprite()
	}
	p.totalhits = p.hits // Save total hits for later use
	if c.size.proj.doscale != 0 {
		p.scale[0] *= c.size.xscale
		p.scale[1] *= c.size.yscale
	}
	// Default Clsn scale
	if !clsnscale {
		p.clsnScale = c.clsnBaseScale
	}
	if c.stWgi().ikemenver[0] == 0 && c.stWgi().ikemenver[1] == 0 {
		p.hitdef.chainid = -1
		p.hitdef.nochainid = [8]int32{-1, -1, -1, -1, -1, -1, -1, -1}
	}
	p.removefacing = c.facing
	if p.velocity[0] < 0 {
		p.facing *= -1
		p.velocity[0] *= -1
		p.accel[0] *= -1
	}
	if op {
		remap := make([]int, len(p.palfx.remap))
		copy(remap, p.palfx.remap)
		p.palfx = newPalFX()
		p.palfx.remap = remap
		c.forceRemapPal(p.palfx, [...]int32{rpg, rpn})
	}
}

func (c *Char) getProjs(id int32) (projs []*Projectile) {
	for i, p := range sys.projs[c.playerNo] {
		if p.id >= 0 && (id < 0 || p.id == id) { // Removed projectiles have negative ID
			projs = append(projs, &sys.projs[c.playerNo][i])
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
	// Assign a value to a NaN field
	ifnanset := func(dst *float32, src float32) {
		if math.IsNaN(float64(*dst)) {
			*dst = src
		}
	}
	// Assign a value to an IErr field
	ifierrset := func(dst *int32, src int32) bool {
		if *dst == IErr {
			*dst = src
			return true
		}
		return false
	}
	ifnanset(&hd.guard_velocity[0], hd.ground_velocity[0])
	ifnanset(&hd.guard_velocity[2], hd.ground_velocity[2])
	ifnanset(&hd.airguard_velocity[0], hd.air_velocity[0]*1.5)
	ifnanset(&hd.airguard_velocity[1], hd.air_velocity[1]*0.5)
	ifnanset(&hd.airguard_velocity[2], hd.air_velocity[2]*1.5)
	ifnanset(&hd.down_velocity[0], hd.air_velocity[0])
	ifnanset(&hd.down_velocity[1], hd.air_velocity[1])
	ifnanset(&hd.down_velocity[2], hd.air_velocity[2])
	ifierrset(&hd.fall.envshake_ampl, -4)

	if hd.air_animtype == RA_Unknown {
		hd.air_animtype = hd.animtype
	}
	if hd.fall.animtype == RA_Unknown {
		if hd.air_animtype >= RA_Up {
			hd.fall.animtype = hd.air_animtype
		} else {
			hd.fall.animtype = RA_Back
		}
	}
	if hd.air_type == HT_Unknown {
		hd.air_type = hd.ground_type
	}
	ifierrset(&hd.forcestand, Btoi(hd.ground_velocity[1] != 0)) // Having a Y velocity causes ForceStand
	ifierrset(&hd.forcecrouch, 0)
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
	if hd.p2clsncheck == -1 {
		if hd.reversal_attr != 0 {
			hd.p2clsncheck = 1
		} else {
			hd.p2clsncheck = 2
		}
	}
	// In Mugen, only projectiles can use air.juggle
	// Ikemen characters can use it to update their juggle points
	if hd.air_juggle == IErr {
		if hd.isprojectile {
			hd.air_juggle = 0
		}
	} else {
		if c.stWgi().ikemenver[0] != 0 || c.stWgi().ikemenver[1] != 0 {
			c.juggle = hd.air_juggle
		}
	}
}

func (c *Char) baseWidthFront() float32 {
	if c.ss.stateType == ST_A {
		return float32(c.size.air.front)
	}
	return float32(c.size.ground.front)
}

func (c *Char) baseWidthBack() float32 {
	if c.ss.stateType == ST_A {
		return float32(c.size.air.back)
	}
	return float32(c.size.ground.back)
}

func (c *Char) baseHeightTop() float32 {
	if c.ss.stateType == ST_L {
		return float32(c.size.height.down)
	} else if c.ss.stateType == ST_A {
		return float32(c.size.height.air[0])
	} else if c.ss.stateType == ST_C {
		return float32(c.size.height.crouch)
	} else {
		return float32(c.size.height.stand)
	}
}

func (c *Char) baseHeightBottom() float32 {
	if c.ss.stateType == ST_A {
		return float32(c.size.height.air[1])
	} else {
		return 0
	}
}

func (c *Char) setFEdge(fe float32) {
	c.edge[0] = fe
	c.setCSF(CSF_frontedge)
}

func (c *Char) setBEdge(be float32) {
	c.edge[1] = be
	c.setCSF(CSF_backedge)
}

func (c *Char) setFWidth(fw float32) {
	c.width[0] = c.baseWidthFront()*((320/c.localcoord)/c.localscl) + fw
	c.setCSF(CSF_frontwidth)
}

func (c *Char) setBWidth(bw float32) {
	c.width[1] = c.baseWidthBack()*((320/c.localcoord)/c.localscl) + bw
	c.setCSF(CSF_backwidth)
}

func (c *Char) setTHeight(th float32) {
	c.height[0] = c.baseHeightTop()*((320/c.localcoord)/c.localscl) + th
	c.setCSF(CSF_topheight)
}

func (c *Char) setBHeight(bh float32) {
	c.height[1] = c.baseHeightBottom()*((320/c.localcoord)/c.localscl) + bh
	c.setCSF(CSF_bottomheight)
}

func (c *Char) updateClsnBaseScale() {
	// Helper parameter
	if c.ownclsnscale && c.animPN == c.playerNo {
		c.clsnBaseScale = [...]float32{c.size.xscale, c.size.yscale}
		return
	}
	// Index range checks. Prevents crashing if chars don't have animations
	// https://github.com/ikemen-engine/Ikemen-GO/issues/1982
	if c.animPN >= 0 && c.animPN < len(sys.chars) && len(sys.chars[c.animPN]) > 0 {
		// The char's base Clsn scale
		// Based on the animation owner's scale constants
		c.clsnBaseScale = [...]float32{
			sys.chars[c.animPN][0].size.xscale,
			sys.chars[c.animPN][0].size.yscale,
		}
	} else {
		// Normally not used. Just a safeguard
		c.clsnBaseScale = [...]float32{1.0, 1.0}
	}
}

func (c *Char) widthToSizeBox() {
	if len(c.width) < 2 || len(c.height) < 2 {
		c.sizeBox = []float32{0, 0, 0, 0}
	} else {
		// Correct left/right and top/bottom
		// Same behavior as Clsn boxes
		// https://github.com/ikemen-engine/Ikemen-GO/issues/2008
		back := -c.width[1]
		front := c.width[0]
		top := -c.height[0]
		bottom := c.height[1]
		if back > front {
			back, front = front, back
		}
		if top > bottom { // Negative sign
			top, bottom = bottom, top
		}
		c.sizeBox = []float32{back, top, front, bottom}
	}
}

func (c *Char) gethitAnimtype() Reaction {
	if c.ghv.fallflag {
		return c.ghv.fall.animtype
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
func (c *Char) isBound() bool {
	return c.ghv.idMatch(c.bindToId)
}
func (c *Char) varGet(i int32) BytecodeValue {
	if i >= 0 && i < int32(NumVar) {
		return BytecodeInt(c.ivar[i])
	}
	sys.appendToConsole(c.warn() + fmt.Sprintf("var index %v out of range", i))
	return BytecodeSF()
}
func (c *Char) fvarGet(i int32) BytecodeValue {
	if i >= 0 && i < int32(NumFvar) {
		return BytecodeFloat(c.fvar[i])
	}
	sys.appendToConsole(c.warn() + fmt.Sprintf("fvar index %v out of range", i))
	return BytecodeSF()
}
func (c *Char) sysVarGet(i int32) BytecodeValue {
	if i >= 0 && i < int32(NumSysVar) {
		return BytecodeInt(c.ivar[i+int32(NumVar)])
	}
	sys.appendToConsole(c.warn() + fmt.Sprintf("sysvar index %v out of range", i))
	return BytecodeSF()
}
func (c *Char) sysFvarGet(i int32) BytecodeValue {
	if i >= 0 && i < int32(NumSysFvar) {
		return BytecodeFloat(c.fvar[i+int32(NumFvar)])
	}
	sys.appendToConsole(c.warn() + fmt.Sprintf("sysfvar index %v out of range", i))
	return BytecodeSF()
}
func (c *Char) varSet(i, v int32) BytecodeValue {
	if i >= 0 && i < int32(NumVar) {
		c.ivar[i] = v
		return BytecodeInt(v)
	}
	sys.appendToConsole(c.warn() + fmt.Sprintf("var index %v out of range", i))
	return BytecodeSF()
}
func (c *Char) fvarSet(i int32, v float32) BytecodeValue {
	if i >= 0 && i < int32(NumFvar) {
		c.fvar[i] = v
		return BytecodeFloat(v)
	}
	sys.appendToConsole(c.warn() + fmt.Sprintf("fvar index %v out of range", i))
	return BytecodeSF()
}
func (c *Char) sysVarSet(i, v int32) BytecodeValue {
	if i >= 0 && i < int32(NumSysVar) {
		c.ivar[i+int32(NumVar)] = v
		return BytecodeInt(v)
	}
	sys.appendToConsole(c.warn() + fmt.Sprintf("sysvar index %v out of range", i))
	return BytecodeSF()
}
func (c *Char) sysFvarSet(i int32, v float32) BytecodeValue {
	if i >= 0 && i < int32(NumSysFvar) {
		c.fvar[i+int32(NumFvar)] = v
		return BytecodeFloat(v)
	}
	sys.appendToConsole(c.warn() + fmt.Sprintf("sysfvar index %v out of range", i))
	return BytecodeSF()
}
func (c *Char) varAdd(i, v int32) BytecodeValue {
	if i >= 0 && i < int32(NumVar) {
		c.ivar[i] += v
		return BytecodeInt(c.ivar[i])
	}
	sys.appendToConsole(c.warn() + fmt.Sprintf("var index %v out of range", i))
	return BytecodeSF()
}
func (c *Char) fvarAdd(i int32, v float32) BytecodeValue {
	if i >= 0 && i < int32(NumFvar) {
		c.fvar[i] += v
		return BytecodeFloat(c.fvar[i])
	}
	sys.appendToConsole(c.warn() + fmt.Sprintf("fvar index %v out of range", i))
	return BytecodeSF()
}
func (c *Char) sysVarAdd(i, v int32) BytecodeValue {
	if i >= 0 && i < int32(NumSysVar) {
		c.ivar[i+int32(NumVar)] += v
		return BytecodeInt(c.ivar[i+int32(NumVar)])
	}
	sys.appendToConsole(c.warn() + fmt.Sprintf("sysvar index %v out of range", i))
	return BytecodeSF()
}
func (c *Char) sysFvarAdd(i int32, v float32) BytecodeValue {
	if i >= 0 && i < int32(NumSysFvar) {
		c.fvar[i+int32(NumFvar)] += v
		return BytecodeFloat(c.fvar[i+int32(NumFvar)])
	}
	sys.appendToConsole(c.warn() + fmt.Sprintf("sysfvar index %v out of range", i))
	return BytecodeSF()
}
func (c *Char) varRangeSet(s, e, v int32) {
	if s >= 0 {
		for i := s; i <= e && i < int32(NumVar); i++ {
			c.ivar[i] = v
		}
	}
}
func (c *Char) fvarRangeSet(s, e int32, v float32) {
	if s >= 0 {
		for i := s; i <= e && i < int32(NumFvar); i++ {
			c.fvar[i] = v
		}
	}
}
func (c *Char) sysVarRangeSet(s, e, v int32) {
	if s >= 0 {
		for i := s; i <= e && i < int32(NumSysVar); i++ {
			c.ivar[i+int32(NumVar)] = v
		}
	}
}
func (c *Char) sysFvarRangeSet(s, e int32, v float32) {
	if s >= 0 {
		for i := s; i <= e && i < int32(NumSysFvar); i++ {
			c.fvar[i+int32(NumFvar)] = v
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
func (c *Char) getTarget(id int32) []int32 {
	if id < 0 { // In Mugen the ID must be specifically -1
		return c.targets
	}
	var tg []int32
	for _, tid := range c.targets {
		if t := sys.playerID(tid); t != nil {
			if t.ghv.hitid == id {
				tg = append(tg, tid)
			}
		}
	}
	return tg
}
func (c *Char) targetFacing(tar []int32, f int32) {
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
			t.setBindToId(c)
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
				c.setX(t.pos[0]*(t.localscl/c.localscl) + t.facing*x)
			}
			if !math.IsNaN(float64(y)) {
				c.setY(t.pos[1]*(t.localscl/c.localscl) + y)
			}
			if !math.IsNaN(float64(z)) {
				c.setZ(t.pos[2]*(t.localscl/c.localscl) + z)
			}
			c.targetBind(tar[:1], time,
				c.facing*c.distX(t, c),
				(t.pos[1]*(t.localscl/c.localscl))-(c.pos[1]*(c.localscl/t.localscl)),
				(t.pos[2]*(t.localscl/c.localscl))-(c.pos[2]*(c.localscl/t.localscl)))
		}
	}
}
func (c *Char) targetLifeAdd(tar []int32, add int32, kill, absolute, dizzy, redlife bool) {
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
	for _, tid := range tar {
		if t := sys.playerID(tid); t != nil && t.player {
			t.powerAdd(power)
		}
	}
}
func (c *Char) targetDizzyPointsAdd(tar []int32, add int32, absolute bool) {
	for _, tid := range tar {
		if t := sys.playerID(tid); t != nil && !t.scf(SCF_dizzy) && !t.asf(ASF_nodizzypointsdamage) {
			t.dizzyPointsAdd(float64(t.computeDamage(float64(add), false, absolute, 1, c, false)), true)
		}
	}
}
func (c *Char) targetGuardPointsAdd(tar []int32, add int32, absolute bool) {
	for _, tid := range tar {
		if t := sys.playerID(tid); t != nil && !t.asf(ASF_noguardpointsdamage) {
			t.guardPointsAdd(float64(t.computeDamage(float64(add), false, absolute, 1, c, false)), true)
		}
	}
}
func (c *Char) targetRedLifeAdd(tar []int32, add int32, absolute bool) {
	for _, tid := range tar {
		if t := sys.playerID(tid); t != nil && !t.asf(ASF_noredlifedamage) {
			t.redLifeAdd(float64(t.computeDamage(float64(add), false, absolute, 1, c, true)), true)
		}
	}
}
func (c *Char) targetScoreAdd(tar []int32, s float32) {
	for _, tid := range tar {
		if t := sys.playerID(tid); t != nil && t.player {
			t.scoreAdd(s)
		}
	}
}
func (c *Char) targetState(tar []int32, state int32) {
	if state >= 0 {
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
				if t.isBound() {
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
func (c *Char) computeDamage(damage float64, kill, absolute bool,
	atkmul float32, attacker *Char, bounds bool) int32 {
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

// A lot of this logic seems the same as computeDamage. Maybe LifeAdd is supposed to use that function as well
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
	// Limit value if kill is false
	if !kill && add <= float64(-c.life) {
		if c.life > 0 || c.stWgi().ikemenver[0] == 0 && c.stWgi().ikemenver[1] == 0 { // See computeDamage
			add = float64(1 - c.life)
		}
	}
	if add < 0 {
		c.receivedDmg += Min(c.life, F64toI32(-add))
	}
	// Safely convert from float64 back to int32 after all calculations are done
	int := F64toI32(float64(c.life) + math.Round(add))
	c.lifeSet(int)
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
		if c.player && c.teamside != -1 {
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
	if !sys.lifebar.redlifebar {
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
	// Safely convert from float64 back to int32 after all calculations are done
	int := F64toI32(float64(c.getPower()) + math.Round(float64(add)))
	if sys.powerShare[c.playerNo&1] && c.teamside != -1 {
		sys.chars[c.playerNo&1][0].setPower(int)
	} else {
		sys.chars[c.playerNo][0].setPower(int)
	}
}

// This only for the PowerSet state controller
func (c *Char) powerSet(pow int32) {
	if sys.powerShare[c.playerNo&1] && c.teamside != -1 {
		sys.chars[c.playerNo&1][0].setPower(pow)
	} else {
		sys.chars[c.playerNo][0].setPower(pow)
	}
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
	if sys.lifebar.stunbar && !sys.roundNoDamage() {
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
	if sys.lifebar.guardbar && !sys.roundNoDamage() {
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
	} else if sys.lifebar.redlifebar && !sys.roundNoDamage() {
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
	if c.teamside == -1 {
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
func (c *Char) distX(opp *Char, oc *Char) float32 {
	cpos := c.pos[0] * c.localscl
	opos := opp.pos[0] * opp.localscl
	// Update distance while bound. Mugen chars only
	if c.stWgi().ikemenver[0] == 0 && c.stWgi().ikemenver[1] == 0 {
		if c.bindToId > 0 && !math.IsNaN(float64(c.bindPos[0])) {
			if bt := sys.playerID(c.bindToId); bt != nil {
				f := bt.facing
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

func (c *Char) bodyDistX(opp *Char, oc *Char) float32 {
	// In Mugen P2BodyDist X does not account for changes in Width like Ikemen does here
	dist := c.distX(opp, oc)
	var oppw float32
	if dist == 0 || (dist < 0) != (opp.facing < 0) {
		oppw = opp.facing * opp.sizeBox[2] * (opp.localscl / oc.localscl)
	} else {
		oppw = -opp.facing * opp.sizeBox[0] * (opp.localscl / oc.localscl)
	}
	return dist + oppw - c.facing*c.sizeBox[2]*(c.localscl/oc.localscl)
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
	ctop := (c.pos[2] - c.size.depth) * c.localscl
	cbot := (c.pos[2] + c.size.depth) * c.localscl
	otop := (opp.pos[2] - opp.size.depth) * opp.localscl
	obot := (opp.pos[2] + opp.size.depth) * opp.localscl
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
	if ^pausetime < sys.pausetime || c.playerNo != c.ss.sb.playerNo ||
		sys.pauseplayer == c.playerNo {
		sys.pausetime = ^pausetime
		sys.pauseplayer = c.playerNo
		if sys.pauseendcmdbuftime < 0 || sys.pauseendcmdbuftime > pausetime {
			sys.pauseendcmdbuftime = 0
		}
	}
	c.pauseMovetime = Max(0, movetime)
	if c.pauseMovetime > pausetime {
		c.pauseMovetime = 0
	} else if sys.pause > 0 && c.pauseMovetime > 0 {
		c.pauseMovetime--
	}
}
func (c *Char) setSuperPauseTime(pausetime, movetime int32, unhittable bool) {
	if ^pausetime < sys.supertime || c.playerNo != c.ss.sb.playerNo ||
		sys.superplayer == c.playerNo {
		sys.supertime = ^pausetime
		sys.superplayer = c.playerNo
		if sys.superendcmdbuftime < 0 || sys.superendcmdbuftime > pausetime {
			sys.superendcmdbuftime = 0
		}
	}
	c.superMovetime = Max(0, movetime)
	if c.superMovetime > pausetime {
		c.superMovetime = 0
	} else if sys.super > 0 && c.superMovetime > 0 {
		c.superMovetime--
	}
	if unhittable {
		c.unhittableTime = pausetime + Btoi(pausetime > 0)
	}
}
func (c *Char) getPalfx() *PalFX {
	if c.palfx != nil {
		return c.palfx
	}
	if c.parentIndex >= 0 {
		if p := c.parent(); p != nil {
			return p.getPalfx()
		}
	}
	c.palfx = newPalFX()
	// Mugen 1.1 behavior if invertblend param is omitted(Only if char mugenversion = 1.1)
	if c.stWgi().mugenver[0] == 1 && c.stWgi().mugenver[1] == 1 && c.stWgi().ikemenver[0] == 0 && c.stWgi().ikemenver[1] == 0 && c.palfx != nil {
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
	c.angle = a
}
func (c *Char) inputOver() bool {
	if c.asf(ASF_postroundinput) {
		return false
	} else {
		// KO'd characters are covered by the inputwait flag
		return sys.time == 0 || sys.intro <= -sys.lifebar.ro.over_time || c.scf(SCF_inputwait)
	}
}
func (c *Char) over() bool {
	return c.scf(SCF_over) || c.ss.no == 5150
}
func (c *Char) makeDust(x, y, z float32) {
	if c.asf(ASF_nomakedust) {
		return
	}
	if e, i := c.newExplod(); e != nil {
		e.anim = c.getAnim(120, "f", true)
		if e.anim != nil {
			e.anim.start_scale[0] *= c.localscl
			e.anim.start_scale[1] *= c.localscl
		}
		e.sprpriority = math.MaxInt32
		e.layerno = c.layerNo
		e.ownpal = true
		e.relativePos = [...]float32{x, y, z}
		e.setPos(c)
		c.insertExplod(i)
	}
}
func (c *Char) hitFallDamage() {
	if c.ss.moveType == MT_H {
		c.lifeAdd(-float64(c.ghv.fall.damage), c.ghv.fall.kill, false)
		c.ghv.fall.damage = 0
	}
}
func (c *Char) hitFallVel() {
	if c.ss.moveType == MT_H {
		if !math.IsNaN(float64(c.ghv.fall.xvelocity)) {
			c.vel[0] = c.ghv.fall.xvelocity
		}
		c.vel[1] = c.ghv.fall.yvelocity
		if !math.IsNaN(float64(c.ghv.fall.zvelocity)) {
			c.vel[2] = c.ghv.fall.zvelocity
		}
	}
}
func (c *Char) hitFallSet(f int32, xv, yv, zv float32) {
	if f >= 0 {
		c.ghv.fallflag = f != 0
	}
	if !math.IsNaN(float64(xv)) {
		c.ghv.fall.xvelocity = xv
	}
	if !math.IsNaN(float64(yv)) {
		c.ghv.fall.yvelocity = yv
	}
	if !math.IsNaN(float64(zv)) {
		c.ghv.fall.zvelocity = zv
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
	if src[0] == -1 {
		c.forceRemapPal(pfx, dst)
		return
	}
	if src[0] < 0 || src[1] < 0 || dst[0] < 0 || dst[1] < 0 {
		return
	}
	si, ok := c.gi().palettedata.palList.PalTable[[...]int16{int16(src[0]),
		int16(src[1])}]
	if !ok || si < 0 {
		sys.appendToConsole(c.warn() + fmt.Sprintf("has no source palette for RemapPal: %v,%v", src[0], src[1]))
		return
	}
	var di int
	di, ok = c.gi().palettedata.palList.PalTable[[...]int16{int16(dst[0]),
		int16(dst[1])}]
	if !ok || di < 0 {
		sys.appendToConsole(c.warn() + fmt.Sprintf("has no dest palette for RemapPal: %v,%v", dst[0], dst[1]))
		return
	}
	if pfx.remap == nil {
		pfx.remap = c.gi().palettedata.palList.GetPalMap()
	}
	if c.gi().palettedata.palList.SwapPalMap(&pfx.remap) {
		c.gi().palettedata.palList.Remap(si, di)
		if src[0] == 1 && src[1] == 1 && c.gi().sff.header.Ver0 == 1 {
			spr := c.gi().sff.GetSprite(0, 0)
			if spr != nil {
				c.gi().palettedata.palList.Remap(spr.palidx, di)
			}
			spr = c.gi().sff.GetSprite(9000, 0)
			if spr != nil {
				c.gi().palettedata.palList.Remap(spr.palidx, di)
			}
		}
		c.gi().palettedata.palList.SwapPalMap(&pfx.remap)
	}
	c.gi().remappedpal = [...]int32{dst[0], dst[1]}
}
func (c *Char) forceRemapPal(pfx *PalFX, dst [2]int32) {
	if dst[0] < 0 || dst[1] < 0 {
		return
	}
	di, ok := c.gi().palettedata.palList.PalTable[[...]int16{int16(dst[0]),
		int16(dst[1])}]
	if !ok || di < 0 {
		return
	}
	if pfx.remap == nil {
		pfx.remap = c.gi().palettedata.palList.GetPalMap()
	}
	for i := range pfx.remap {
		pfx.remap[i] = di
	}
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
	case 0:
		c.mapArray[key] = Value
	case 1:
		c.mapArray[key] += Value
	case 2:
		if c.parent() != nil {
			c.parent().mapArray[key] = Value
		} else {
			c.mapArray[key] = Value
		}
	case 3:
		if c.parent() != nil {
			c.parent().mapArray[key] += Value
		} else {
			c.mapArray[key] += Value
		}
	case 4:
		if c.root() != nil {
			c.root().mapArray[key] = Value
		} else {
			c.mapArray[key] = Value
		}
	case 5:
		if c.root() != nil {
			c.root().mapArray[key] += Value
		} else {
			c.mapArray[key] += Value
		}
	case 6:
		if c.teamside == -1 {
			for i := MaxSimul * 2; i < MaxSimul*2+MaxAttachedChar; i += 1 {
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
	case 7:
		if c.teamside == -1 {
			for i := MaxSimul * 2; i < MaxSimul*2+MaxAttachedChar; i += 1 {
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

func (c *Char) appendLifebarAction(text string, snd, spr [2]int32, anim, time int32, timemul float32, top bool) {
	if c.teamside == -1 {
		return
	}
	if _, ok := sys.lifebar.missing["[action]"]; ok {
		return
	}
	if snd[0] != -1 {
		sys.lifebar.snd.play(snd, 100, 0, 0, 0, 0)
	}
	index := 0
	if !top {
		for k, v := range sys.lifebar.ac[c.teamside].messages {
			if v.del {
				sys.lifebar.ac[c.teamside].messages = removeLbMsg(sys.lifebar.ac[c.teamside].messages, k)
				break
			}
			index++
		}
	}
	if time == -1 {
		time = sys.lifebar.ac[c.teamside].displaytime
	}
	msg := newLbMsg(text, int32(float32(time)*timemul), c.teamside)
	if anim != -1 || spr[0] != -1 {
		delete(sys.lifebar.ac[c.teamside].is, fmt.Sprintf("team%v.front.anim", c.teamside+1))
		delete(sys.lifebar.ac[c.teamside].is, fmt.Sprintf("team%v.front.spr", c.teamside+1))
		if anim != -1 {
			sys.lifebar.ac[c.teamside].is[fmt.Sprintf("team%v.front.anim", c.teamside+1)] = fmt.Sprintf("%v", anim)
		} else {
			sys.lifebar.ac[c.teamside].is[fmt.Sprintf("team%v.front.spr", c.teamside+1)] = fmt.Sprintf("%v,%v", spr[0], spr[1])
		}
		msg.bg = *ReadAnimLayout(fmt.Sprintf("team%v.bg.", c.teamside+1), sys.lifebar.ac[c.teamside].is, sys.lifebar.sff, sys.lifebar.at, 2)
		msg.front = *ReadAnimLayout(fmt.Sprintf("team%v.front.", c.teamside+1), sys.lifebar.ac[c.teamside].is, sys.lifebar.sff, sys.lifebar.at, 2)
	}
	sys.lifebar.ac[c.teamside].messages = insertLbMsg(sys.lifebar.ac[c.teamside].messages, msg, index)
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
		if len(c.clipboardText) > sys.clipboardRows {
			c.clipboardText = c.clipboardText[len(c.clipboardText)-sys.clipboardRows:]
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
	friction := float32(0.7)
	if c.cornerVelOff != 0 && sys.super == 0 {
		for _, p := range sys.chars {
			if len(p) > 0 && p[0].ss.moveType == MT_H && p[0].ghv.id == c.id {
				npos := (p[0].pos[0] + p[0].vel[0]*p[0].facing) * p[0].localscl
				if p[0].trackableByCamera() && p[0].csf(CSF_screenbound) && (npos <= sys.xmin || npos >= sys.xmax) {
					c.mhv.cornerpush = c.cornerVelOff
				}
				// In Ikemen the cornerpush friction is defined by the target instead
				if c.stWgi().ikemenver[0] == 0 && c.stWgi().ikemenver[1] == 0 {
					friction = 0.7
				} else {
					if p[0].ss.stateType == ST_C || p[0].ss.stateType == ST_L {
						friction = p[0].gi().movement.crouch.friction
					} else {
						friction = p[0].gi().movement.stand.friction
					}
				}
			}
		}
	}
	nobind := [...]bool{c.bindTime == 0 || math.IsNaN(float64(c.bindPos[0])),
		c.bindTime == 0 || math.IsNaN(float64(c.bindPos[1])),
		c.bindTime == 0 || math.IsNaN(float64(c.bindPos[2]))}
	for i := range nobind {
		if nobind[i] {
			c.oldPos[i], c.interPos[i] = c.pos[i], c.pos[i]
		}
	}
	if c.csf(CSF_posfreeze) {
		if nobind[0] {
			c.setPosX(c.oldPos[0] + c.mhv.cornerpush)
		}
	} else {
		// Controls speed
		if nobind[0] {
			c.setPosX(c.oldPos[0] + c.vel[0]*c.facing + c.mhv.cornerpush)
		}
		if nobind[1] {
			c.setPosY(c.oldPos[1] + c.vel[1])
		}
		if nobind[2] {
			c.setPosZ(c.oldPos[2] + c.vel[2])
		}

		switch c.ss.physics {
		case ST_S:
			c.vel[0] *= c.gi().movement.stand.friction
			if AbsF(c.vel[0]) < 1 {
				c.vel[0] = 0
			}
		case ST_C:
			c.vel[0] *= c.gi().movement.crouch.friction
		case ST_A:
			c.gravity()
		}
	}
	if sys.super == 0 {
		c.cornerVelOff *= friction
		if AbsF(c.cornerVelOff) < 1 {
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
func (c *Char) setBindTime(time int32) {
	c.bindTime = time
	if time == 0 {
		c.bindToId = -1
		c.bindFacing = 0
	}
}
func (c *Char) setBindToId(to *Char) {
	if c.bindToId != to.id {
		c.bindToId = to.id
	}
	if c.bindFacing == 0 {
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
			if AbsF(c.bindFacing) == 2 {
				f = c.bindFacing / 2
			}
			c.setX(bt.pos[0]*bt.localscl/c.localscl + f*(c.bindPos[0]+c.bindPosAdd[0]))
			c.interPos[0] += bt.interPos[0] - bt.pos[0]
			c.oldPos[0] += bt.oldPos[0] - bt.pos[0]
			c.pushed = c.pushed || bt.pushed
			c.ghv.xoff = 0
		}
		if !math.IsNaN(float64(c.bindPos[1])) {
			c.setY(bt.pos[1]*bt.localscl/c.localscl + (c.bindPos[1] + c.bindPosAdd[1]))
			c.interPos[1] += bt.interPos[1] - bt.pos[1]
			c.oldPos[1] += bt.oldPos[1] - bt.pos[1]
			c.ghv.yoff = 0
		}
		if !math.IsNaN(float64(c.bindPos[2])) {
			c.setZ(bt.pos[2]*bt.localscl/c.localscl + (c.bindPos[2] + c.bindPosAdd[2]))
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
		min, max := c.edge[0], -c.edge[1]
		if c.facing > 0 {
			min, max = -max, -min
		}
		x = ClampF(x, min+sys.xmin/c.localscl, max+sys.xmax/c.localscl)
	}
	if c.csf(CSF_stagebound) {
		x = ClampF(x, sys.stage.leftbound*sys.stage.localscl/c.localscl, sys.stage.rightbound*sys.stage.localscl/c.localscl)
	}
	c.setPosX(x)
}

func (c *Char) zDepthBound() {
	posz := c.pos[2]
	if c.csf(CSF_stagebound) {
		posz = ClampF(posz, sys.zmin/c.localscl, sys.zmax/c.localscl)
	}
	c.setPosZ(posz)
}

func (c *Char) xPlatformBound(pxmin, pxmax float32) {
	x := c.pos[0]
	if c.ss.stateType != ST_A {
		min, max := c.edge[0], -c.edge[1]
		if c.facing > 0 {
			min, max = -max, -min
		}
		x = ClampF(x, min+pxmin/c.localscl, max+pxmax/c.localscl)
	}
	c.setX(x)
	c.xScreenBound()
}
func (c *Char) gethitBindClear() {
	if c.isBound() {
		c.setBindTime(0)
	}
}
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
func (c *Char) exitTarget(explremove bool) {
	if c.hittmp >= 0 {
		for _, hb := range c.ghv.hitBy {
			if e := sys.playerID(hb[0]); e != nil {
				if e.hitdef.reversal_attr == 0 || e.hitdef.reversal_attr == -1<<31 {
					e.removeTarget(c.id)
					//if explremove {
					//	c.enemyExplodsRemove(e.playerNo)
					//}
				} else {
					c.ghv.hitid = c.ghv.hitid >> 31
				}
			}
		}
		c.gethitBindClear()
	}
	c.ghv.hitBy = c.ghv.hitBy[:0]
}
func (c *Char) offsetX() float32 {
	return float32(c.size.draw.offset[0])*c.facing + c.offset[0]/c.localscl
}
func (c *Char) offsetY() float32 {
	return float32(c.size.draw.offset[1]) + c.offset[1]/c.localscl
}

func (c *Char) projClsnCheck(p *Projectile, cbox, pbox int32) bool {
	if p.ani == nil || c.curFrame == nil || c.scf(SCF_standby) || c.scf(SCF_disabled) {
		return false
	}
	frm := p.ani.CurrentFrame()
	if frm == nil {
		return false
	}

	// Accepted box types
	if cbox != 1 && cbox != 2 && cbox != 3 {
		return false
	}

	// Required boxes not found
	if p.hitdef.p2clsnrequire == 1 && c.curFrame.Clsn1() == nil ||
		p.hitdef.p2clsnrequire == 2 && c.curFrame.Clsn2() == nil {
		return false
	}

	// Decide which box types should collide
	var clsn1, clsn2 []float32
	if c.asf(ASF_projtypecollision) { // Projectiles trade with their Clsn2 only
		clsn1 = frm.Clsn2()
		clsn2 = c.curFrame.Clsn2()
	} else {
		if pbox == 2 {
			clsn1 = frm.Clsn2()
		} else {
			clsn1 = frm.Clsn1()
		}
		if cbox == 1 {
			clsn2 = c.curFrame.Clsn1()
			if clsn2 == nil && p.hitdef.p2clsnrequire == 1 {
				return false
			}
		} else if cbox == 3 {
			clsn2 = c.sizeBox
			// Size box always exists
		} else {
			clsn2 = c.curFrame.Clsn2()
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

	return sys.clsnOverlap(clsn1,
		[...]float32{p.clsnScale[0] * p.localscl * p.zScale, p.clsnScale[1] * p.localscl * p.zScale},
		[...]float32{p.pos[0] * p.localscl, p.pos[1] * p.localscl},
		p.facing,
		p.clsnAngle,
		clsn2,
		charscale,
		[...]float32{c.pos[0]*c.localscl + c.offsetX()*c.localscl,
			c.pos[1]*c.localscl + c.offsetY()*c.localscl},
		c.facing,
		charangle)
}

func (c *Char) clsnCheck(getter *Char, charbox, getterbox int32, reqcheck, trigger bool) bool {

	// What this does is normally check the Clsn in the currently displayed frame
	// But in the ClsnOverlap trigger, we must check the frame that *will* be displayed instead
	charframe := c.curFrame
	getterframe := getter.curFrame
	if trigger {
		charframe = c.anim.CurrentFrame()
		getterframe = getter.anim.CurrentFrame()
	}

	// Nil anim & standby check.
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
		if c.hitdef.p2clsnrequire == 1 && getterframe.Clsn1() == nil ||
			c.hitdef.p2clsnrequire == 2 && getterframe.Clsn2() == nil {
			return false
		}
	}

	// Decide which box types should collide
	var clsn1, clsn2 []float32
	if c.asf(ASF_projtypecollision) && getter.asf(ASF_projtypecollision) { // Projectiles trade with their Clsn2 only
		clsn1 = charframe.Clsn2()
		clsn2 = getterframe.Clsn2()
	} else {
		if charbox == 1 {
			clsn1 = charframe.Clsn1()
		} else if charbox == 3 {
			clsn1 = c.sizeBox
		} else {
			clsn1 = charframe.Clsn2()
		}
		if getterbox == 1 {
			clsn2 = getterframe.Clsn1()
		} else if getterbox == 3 {
			clsn2 = getter.sizeBox
		} else {
			clsn2 = getterframe.Clsn2()
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

	return sys.clsnOverlap(clsn1,
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
		getterangle)
}

func (c *Char) hitByAttrCheck(attr int32, gstyp StateType) bool {
	// Get state type (SCA) from among the attributes
	styp := attr & int32(ST_MASK)
	// Note: In Mugen, invincibility is checked against both the Hitdef attribute and the enemy's actual statetype
	// Ikemen characters work as documented. Invincibility only cares about the Hitdef's attributes (including its statetype)
	if c.stWgi().ikemenver[0] == 0 && c.stWgi().ikemenver[1] == 0 {
		if gstyp == ST_N {
			styp = attr & int32(ST_MASK)
		} else {
			styp = int32(gstyp)
		}
	}

	hit := true
	for _, hb := range c.hitby {
		if hb.time != 0 {
			if hb.flag&styp == 0 || hb.flag&attr&^int32(ST_MASK) == 0 {
				hit = false
				if hb.stack { // Stack parameter makes the hit happen if any HitBy slot would allow it
					continue
				} else {
					break
				}
			}
			if hb.stack {
				hit = true
				break
			}
		}
	}
	return hit
}

func (c *Char) hitByPlayerNoCheck(getterno int) bool {
	hit := true
	for _, hb := range c.hitby {
		if hb.time != 0 {
			if hb.playerno >= 0 && hb.playerno != getterno {
				if hb.not {
					hit = true
					if hb.stack {
						continue
					} else {
						break
					}
				} else {
					hit = false
					if hb.stack {
						continue
					} else {
						break
					}
				}
			}
		}
	}
	return hit
}

func (c *Char) hitByPlayerIdCheck(getterid int32) bool {
	hit := true
	for _, hb := range c.hitby {
		if hb.time != 0 {
			if hb.playerid >= 0 && hb.playerid != getterid {
				if hb.not {
					hit = true
					if hb.stack {
						continue
					} else {
						break
					}
				} else {
					hit = false
					if hb.stack {
						continue
					} else {
						break
					}
				}
			}
		}
	}
	return hit
}

// Check if Hitdef attributes can hit a player
func (c *Char) attrCheck(ghd *HitDef, getter *Char, gstyp StateType) bool {
	if c.unhittableTime > 0 || ghd.chainid >= 0 && c.ghv.hitid != ghd.chainid && ghd.nochainid[0] == -1 {
		return false
	}
	if (len(c.ghv.hitBy) > 0 && c.ghv.hitBy[len(c.ghv.hitBy)-1][0] == getter.id) || c.ghv.hitshaketime > 0 { // https://github.com/ikemen-engine/Ikemen-GO/issues/320
		for _, nci := range ghd.nochainid {
			if nci >= 0 && c.ghv.hitid == nci && c.ghv.id == ghd.attackerID {
				return false
			}
		}
	}
	if ghd.reversal_attr > 0 {
		return c.atktmp != 0 && c.hitdef.attr > 0 &&
			(c.hitdef.attr&ghd.reversal_attr&int32(ST_MASK)) != 0 &&
			(c.hitdef.attr&ghd.reversal_attr&^int32(ST_MASK)) != 0
	}
	if ghd.attr <= 0 || ghd.hitflag&int32(c.ss.stateType) == 0 ||
		(ghd.hitflag&int32(ST_F) == 0 || getter.asf(ASF_nofallhitflag)) && c.hittmp >= 2 ||
		ghd.hitflag&int32(MT_MNS) != 0 && c.hittmp > 0 ||
		ghd.hitflag&int32(MT_PLS) != 0 && (c.hittmp <= 0 || c.inGuardState()) {
		return false
	}

	// https://github.com/ikemen-engine/Ikemen-GO/issues/308
	//if ghd.chainid < 0 {

	// HitBy and NotHitBy checks
	if !c.hitByAttrCheck(ghd.attr, gstyp) {
		return false
	}
	if !c.hitByPlayerNoCheck(getter.playerNo) {
		return false
	}
	if !c.hitByPlayerIdCheck(getter.id) {
		return false
	}
	return true
}

// Check if the enemy (c) Hitdef should lose to the current one, if applicable
func (c *Char) hittableByChar(ghd *HitDef, getter *Char, gst StateType, proj bool) bool {

	// Enemy can't be hit by Hitdef attributes at all
	// No more checks needed
	if !c.attrCheck(ghd, getter, gst) {
		return false
	}

	// Enemy's Hitdef already hit the original char
	// Can skip priority checking
	if c.hasTargetOfHitdef(getter.id) {
		return true
	}

	// Check if enemy can trade hits with original char
	// This should probably be a function that both players access instead of being handled like this
	countercheck := func(hd *HitDef) bool {
		if proj {
			return false
		} else {
			return (getter.atktmp >= 0 || !c.hasTarget(getter.id)) &&
				!getter.hasTargetOfHitdef(c.id) &&
				getter.attrCheck(hd, c, c.ss.stateType) &&
				c.clsnCheck(getter, 1, c.hitdef.p2clsncheck, true, false) &&
				sys.zAxisOverlap(c.pos[2], c.hitdef.attack.depth[0], c.hitdef.attack.depth[1], c.localscl,
					getter.pos[2], getter.size.depth, getter.size.depth, getter.localscl)
		}
	}

	// Hitdef priority check
	if c.atktmp != 0 && (c.hitdef.attr > 0 && c.ss.stateType != ST_L || c.hitdef.reversal_attr > 0) {
		switch {
		case c.hitdef.reversal_attr > 0:
			if ghd.reversal_attr > 0 { // Reversaldef vs Reversaldef
				if countercheck(&c.hitdef) {
					c.atktmp = -1
					return getter.atktmp < 0
				}
				return true
			}
		case ghd.reversal_attr > 0:
			return true
		case ghd.priority < c.hitdef.priority:
		case ghd.priority == c.hitdef.priority:
			switch {
			case c.hitdef.bothhittype == AT_Dodge:
			case ghd.bothhittype != AT_Hit:
			case c.hitdef.bothhittype == AT_Hit:
				if (c.hitdef.p1stateno >= 0 || c.hitdef.attr&int32(AT_AT) != 0 &&
					ghd.hitonce != 0) && countercheck(&c.hitdef) {
					c.atktmp = -1
					return getter.atktmp < 0 || Rand(0, 1) == 1
				}
				return true
			default:
				return true
			}
		default:
			return true
		}
		return !countercheck(&c.hitdef)
	}
	return true
}

func (c *Char) actionPrepare() {
	if c.minus != 2 || c.csf(CSF_destroy) || c.scf(SCF_disabled) {
		return
	}
	c.pauseBool = false
	if c.cmd != nil {
		if sys.super > 0 {
			c.pauseBool = c.superMovetime == 0
		} else if sys.pause > 0 && c.pauseMovetime == 0 {
			c.pauseBool = true
		}
	}
	c.acttmp = -int8(Btoi(c.pauseBool)) * 2
	// Due to the nature of how pauses are processed, these are needed to fix an "off by 1" error in the PauseTime trigger
	c.prevSuperMovetime = c.superMovetime
	c.prevPauseMovetime = c.pauseMovetime
	if !c.pauseBool {
		// Perform basic actions
		if c.keyctrl[0] && c.cmd != nil {
			// In Mugen, characters can perform basic actions even if they are KO
			if c.ctrl() && !c.inputOver() && (c.key >= 0 || c.helperIndex == 0) {
				if !c.asf(ASF_nohardcodedkeys) {
					if !c.asf(ASF_nojump) && c.ss.stateType == ST_S && c.cmd[0].Buffer.U > 0 &&
						(!(sys.intro < 0 && sys.intro > -sys.lifebar.ro.over_waittime) || c.asf(ASF_postroundinput)) {
						if c.ss.no != 40 {
							c.changeState(40, -1, -1, "")
						}
					} else if !c.asf(ASF_noairjump) && c.ss.stateType == ST_A && c.cmd[0].Buffer.Ub == 1 &&
						c.pos[1] <= -float32(c.gi().movement.airjump.height) &&
						c.airJumpCount < c.gi().movement.airjump.num {
						if c.ss.no != 45 || c.ss.time > 0 {
							c.airJumpCount++
							c.changeState(45, -1, -1, "")
						}
					} else {
						if !c.asf(ASF_nocrouch) && c.ss.stateType == ST_S && c.cmd[0].Buffer.D > 0 {
							if c.ss.no != 10 {
								if c.ss.no != 100 {
									c.vel[0] = 0
								}
								c.changeState(10, -1, -1, "")
							}
						} else if !c.asf(ASF_nostand) && c.ss.stateType == ST_C && c.cmd[0].Buffer.D < 0 {
							if c.ss.no != 12 {
								c.changeState(12, -1, -1, "")
							}
						} else if !c.asf(ASF_nowalk) && c.ss.stateType == ST_S &&
							(c.cmd[0].Buffer.F > 0 != ((!c.inguarddist || c.prevNoStandGuard) && c.cmd[0].Buffer.B > 0)) {
							if c.ss.no != 20 {
								c.changeState(20, -1, -1, "")
							}
						} else if !c.asf(ASF_nobrake) && c.ss.no == 20 &&
							(c.cmd[0].Buffer.B > 0) == (c.cmd[0].Buffer.F > 0) {
							c.changeState(0, -1, -1, "")
						}
						if c.inguarddist && c.scf(SCF_guard) && c.cmd[0].Buffer.B > 0 &&
							!c.inGuardState() {
							c.changeState(120, -1, -1, "")
						}
					}
				}
			} else {
				switch c.ss.no {
				case 11:
					if !c.asf(ASF_nostand) {
						c.changeState(12, -1, -1, "")
					}
				case 20:
					if !c.asf(ASF_nobrake) && c.cmd[0].Buffer.U < 0 && c.cmd[0].Buffer.D < 0 &&
						c.cmd[0].Buffer.B < 0 && c.cmd[0].Buffer.F < 0 {
						c.changeState(0, -1, -1, "")
					}
				}
			}
		}
		if c.ss.stateType != ST_A {
			c.airJumpCount = 0
		}
		if !c.hitPause() {
			if !sys.roundEnd() {
				if c.alive() && c.life > 0 {
					c.unsetSCF(SCF_over | SCF_ko_round_middle)
				}
				if c.ss.no == 5150 || c.scf(SCF_over) {
					c.setSCF(SCF_ko_round_middle)
				}
			}
			if c.ss.no == 5150 && c.life <= 0 {
				c.setSCF(SCF_over)
			}
			c.specialFlag = 0
			c.inputFlag = 0
			c.setCSF(CSF_stagebound)
			if c.player {
				if c.alive() || c.ss.no != 5150 || c.numPartner() == 0 {
					c.setCSF(CSF_screenbound | CSF_movecamera_x | CSF_movecamera_y)
				}
				if sys.roundState() > 0 && (c.alive() || c.numPartner() == 0) {
					c.setCSF(CSF_playerpush)
				}
			}
			c.pushPriority = 0 // Reset player pushing priority
			c.attackDist = [2]float32{c.size.attack.dist.front, c.size.attack.dist.back}
			// HitBy timers
			// In Mugen this seems to happen at the end of each frame instead
			for i, hb := range c.hitby {
				if hb.time > 0 {
					c.hitby[i].time--
				}
			}
			// HitOverride timers
			// In Mugen they decrease even during hitpause. However no issues have arised from not doing that yet
			for i, ho := range c.ho {
				if ho.time > 0 {
					c.ho[i].time--
				}
			}
			if sys.super > 0 {
				if c.superMovetime > 0 {
					c.superMovetime--
				}
			} else if sys.pause > 0 && c.pauseMovetime > 0 {
				c.pauseMovetime--
			}
		}
		// This AssertSpecial flag is special in that it must always reset regardless of hitpause
		c.unsetASF(ASF_animatehitpause)
		// The flags in this block are to be reset even during hitpause
		// Exception for WinMugen chars, where they persisted during hitpause
		if c.stWgi().ikemenver[0] != 0 || c.stWgi().ikemenver[1] != 0 || c.stWgi().mugenver[0] == 1 || !c.hitPause() {
			c.unsetCSF(CSF_angledraw | CSF_trans)
			c.angleScale = [...]float32{1, 1}
			c.offset = [2]float32{}
			// Reset all AssertSpecial flags except the following, which are reset elsewhere in the code
			c.assertFlag = (c.assertFlag&ASF_nostandguard | c.assertFlag&ASF_nocrouchguard | c.assertFlag&ASF_noairguard |
				c.assertFlag&ASF_runfirst | c.assertFlag&ASF_runlast)
		}
		// The flags below also reset during hitpause, but are new to Ikemen and don't need the exception above
		// Reset Clsn modifiers
		c.clsnScaleMul = [...]float32{1.0, 1.0}
		c.clsnAngle = 0
		// Reset shadow offsets
		c.shadowOffset = [2]float32{}
		c.reflectOffset = [2]float32{}
	}
	// Decrease unhittable timer
	// This used to be in tick(), but Mugen Clsn display suggests it happens sooner than that
	// This used to be CharGlobalInfo, but that made root and helpers share the same timer
	// In Mugen this timer won't decrease unless the char has a Clsn box (of any type)
	if c.unhittableTime > 0 {
		c.unhittableTime--
	}
	c.dropTargets()
	if c.downHitOffset {
		c.pos[0] += c.gi().movement.down.gethit.offset[0] * (320 / c.localcoord) / c.localscl * c.facing
		c.pos[1] += c.gi().movement.down.gethit.offset[1] * (320 / c.localcoord) / c.localscl
		c.downHitOffset = false
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
		if c.ss.sb.playerNo == c.playerNo && (c.player || c.keyctrl[2]) {
			if sb, ok := c.gi().states[-3]; ok {
				sb.run(c)
			}
		}
		// Run state -2
		c.minus = -2
		if c.player || c.keyctrl[1] {
			if sb, ok := c.gi().states[-2]; ok {
				sb.run(c)
			}
		}
		// Run state -1
		c.minus = -1
		if c.ss.sb.playerNo == c.playerNo && (c.player || c.keyctrl[0]) {
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
	if sys.autoguard[c.playerNo] {
		c.setASF(ASF_autoguard)
	}
	if !c.inputOver() &&
		((c.scf(SCF_ctrl) || c.ss.no == 52) &&
			c.ss.moveType == MT_I || c.inGuardState()) && c.cmd != nil &&
		(c.cmd[0].Buffer.B > 0 || c.asf(ASF_autoguard)) &&
		(c.ss.stateType == ST_S && !c.asf(ASF_nostandguard) ||
			c.ss.stateType == ST_C && !c.asf(ASF_nocrouchguard) ||
			c.ss.stateType == ST_A && !c.asf(ASF_noairguard)) {
		c.setSCF(SCF_guard)
	}
	if !c.pauseBool {
		if c.keyctrl[0] && c.cmd != nil {
			if c.ctrl() && !c.inputOver() && (c.key >= 0 || c.helperIndex == 0) {
				if !c.asf(ASF_nohardcodedkeys) {
					if c.inguarddist && c.scf(SCF_guard) && c.cmd[0].Buffer.B > 0 &&
						!c.inGuardState() {
						c.changeState(120, -1, -1, "")
						// In Mugen the characters *can* change to the guarding states during pauses
						// They can still block in Ikemen despite not changing state here
					}
				}
			}
		}
	}
	// This variable is necessary because NoStandGuard is reset before the walking instructions are checked
	// https://github.com/ikemen-engine/Ikemen-GO/issues/1966
	c.prevNoStandGuard = c.asf(ASF_nostandguard)
	c.unsetASF(ASF_nostandguard | ASF_nocrouchguard | ASF_noairguard)
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
		if !c.csf(CSF_frontwidth) {
			c.width[0] = c.baseWidthFront() * ((320 / c.localcoord) / c.localscl)
		}
		if !c.csf(CSF_backwidth) {
			c.width[1] = c.baseWidthBack() * ((320 / c.localcoord) / c.localscl)
		}
		if !c.csf(CSF_frontedge) {
			c.edge[0] = 0
		}
		if !c.csf(CSF_backedge) {
			c.edge[1] = 0
		}
		if !c.csf(CSF_topheight) {
			c.height[0] = c.baseHeightTop() * ((320 / c.localcoord) / c.localscl)
		}
		if !c.csf(CSF_bottomheight) {
			c.height[1] = c.baseHeightBottom() * ((320 / c.localcoord) / c.localscl)
		}
	}
	// Update size box according to player width and height
	// This box will replace width and height values in some other parts of the code
	// TODO: More refactoring so the box can replace width and height entirely
	c.widthToSizeBox()
	if !c.pauseBool {
		if !c.hitPause() {
			if c.ss.no == 5110 && c.ghv.down_recovertime <= 0 && c.alive() && !c.asf(ASF_nogetupfromliedown) {
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
			c.groundLevel = 0 // Only after position is updated
			c.setFacing(c.p1facing)
			c.p1facing = 0
			c.ss.time++
			if c.mctime > 0 {
				c.mctime++
			}
			if c.anim != nil {
				c.curFrame = c.anim.CurrentFrame()
			} else {
				c.curFrame = nil
			}
		}
		if c.ghv.damage != 0 {
			// HitOverride KeepState flag still allows damage to get through
			if c.ss.moveType == MT_H || c.hoKeepState {
				c.lifeAdd(-float64(c.ghv.damage), true, true)
			}
			c.ghv.damage = 0
		}
		if c.ghv.redlife != 0 {
			if c.ss.moveType == MT_H || c.hoKeepState {
				c.redLifeAdd(-float64(c.ghv.redlife), true)
			}
			c.ghv.redlife = 0
		}
		if c.ghv.dizzypoints != 0 {
			if c.ss.moveType == MT_H || c.hoKeepState {
				c.dizzyPointsAdd(-float64(c.ghv.dizzypoints), true)
			}
			c.ghv.dizzypoints = 0
		}
		if c.ghv.guardpoints != 0 {
			if c.ss.moveType == MT_H || c.hoKeepState {
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
					// HitOverride KeepState preserves some GetHitVars for 1 frame so they can be accessed by the char
					if !c.hoKeepState {
						c.ghv.hitshaketime = 0
						c.ghv.attr = 0
						c.ghv.id = 0
						c.ghv.playerNo = -1
					}
					c.superDefenseMul = 1
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
		c.gi().projidcount = 0
	}
	c.xScreenBound()
	c.zDepthBound()

	// Final scale calculations
	// Clsn and size box scale used to factor zScale here, but they shouldn't
	// Game logic should stay the same regardless of Z scale. Only drawing changes
	c.zScale = sys.updateZScale(c.pos[2], c.localscl)                                 // Must be placed after posUpdate()
	c.clsnScale = [2]float32{c.clsnBaseScale[0] * c.clsnScaleMul[0] * c.animlocalscl, // No facing here
		c.clsnBaseScale[1] * c.clsnScaleMul[1] * c.animlocalscl}

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
	if (c.minus < 1) || c.csf(CSF_destroy) || c.scf(SCF_disabled) {
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
	c.minus = 1
}
func (c *Char) track() {
	if c.trackableByCamera() {
		min, max := c.edge[0], -c.edge[1]
		if c.facing > 0 {
			min, max = -max, -min
		}
		if !sys.cam.roundstart && c.csf(CSF_screenbound) && !c.scf(SCF_standby) {
			c.interPos[0] = ClampF(c.interPos[0], min+sys.xmin/c.localscl, max+sys.xmax/c.localscl)
		}
		if c.csf(CSF_movecamera_x) && !c.scf(SCF_standby) {
			if c.interPos[0]*c.localscl-min*c.localscl < sys.cam.leftest {
				sys.cam.leftest = MinF(c.interPos[0]*c.localscl-min*c.localscl, sys.cam.leftest)
				if c.acttmp > 0 && !c.csf(CSF_posfreeze) &&
					(c.bindTime == 0 || math.IsNaN(float64(c.bindPos[0]))) {
					sys.cam.leftestvel = c.vel[0] * c.localscl * c.facing
				} else {
					sys.cam.leftestvel = 0
				}
			}
			if c.interPos[0]*c.localscl-max*c.localscl > sys.cam.rightest {
				sys.cam.rightest = MaxF(c.interPos[0]*c.localscl-max*c.localscl, sys.cam.rightest)
				if c.acttmp > 0 && !c.csf(CSF_posfreeze) &&
					(c.bindTime == 0 || math.IsNaN(float64(c.bindPos[0]))) {
					sys.cam.rightestvel = c.vel[0] * c.localscl * c.facing
				} else {
					sys.cam.rightestvel = 0
				}
			}
		}
		if c.csf(CSF_movecamera_y) && !c.scf(SCF_standby) {
			sys.cam.highest = MinF(c.interPos[1]*c.localscl, sys.cam.highest)
			sys.cam.lowest = MaxF(c.interPos[1]*c.localscl, sys.cam.lowest)
			sys.cam.Pos[1] = 0
		}
	}
}
func (c *Char) update() {
	if c.scf(SCF_disabled) {
		return
	}
	if sys.tickFrame() {
		if c.csf(CSF_destroy) {
			c.destroy()
			return
		}
		if !c.pause() && !c.isBound() {
			c.bind()
		}
		if c.acttmp > 0 {
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
		}
		if c.ss.moveType == MT_H {
			// Set opposing team's First Attack flag
			if sys.firstAttack[2] == 0 && (c.teamside == 0 || c.teamside == 1) {
				if sys.firstAttack[1-c.teamside] < 0 && c.ghv.playerNo >= 0 && c.ghv.guarded == false {
					sys.firstAttack[1-c.teamside] = c.ghv.playerNo
				}
			}
			if sys.super <= 0 && sys.pause <= 0 {
				c.superMovetime, c.pauseMovetime = 0, 0
			}
			c.hittmp = int8(Btoi(c.ghv.fallflag)) + 1
			if c.acttmp > 0 && (c.ss.no == 5100 || c.ss.no == 5070) && c.ss.time == 1 {
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
		// https://github.com/ikemen-engine/Ikemen-GO/issues/1592
		if c.acttmp > 0 && c.ss.moveType != MT_H || c.ss.no == 5150 {
			c.exitTarget(true)
		}
		c.platformPosY = 0
		c.groundAngle = 0
		// Hit detection should happen even during hitpause
		// https://github.com/ikemen-engine/Ikemen-GO/issues/1660
		c.atktmp = int8(Btoi(c.ss.moveType != MT_I || c.hitdef.reversal_attr > 0))
		c.hoIdx = -1
		c.hoKeepState = false
		if c.acttmp > 0 {
			if c.inGuardState() {
				c.setSCF(SCF_guard)
			}
			if ((c.ss.moveType == MT_H && (c.ss.stateType == ST_S || c.ss.stateType == ST_C)) || c.ss.no == 52) && c.pos[1] == 0 &&
				AbsF(c.pos[0]-c.dustOldPos) >= 1 && c.ss.time%3 == 0 {
				c.makeDust(0, 0, 0)
			}
		}
	}
	var customDefense float32 = 1
	if !c.defenseMulDelay || c.ss.moveType == MT_H {
		customDefense = c.customDefense
	}
	c.finalDefense = float64(((float32(c.gi().data.defence) * customDefense * c.superDefenseMul * c.fallDefenseMul) / 100))
	if sys.tickNextFrame() {
		c.pushed = false
	}
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
	if c.koEchoTime > 0 {
		if !c.scf(SCF_ko) || sys.gsf(GSF_nokosnd) {
			c.koEchoTime = 0
		} else {
			if c.koEchoTime == 60 || c.koEchoTime == 120 {
				vo := int32(100 * (240 - (c.koEchoTime + 60)) / 240)
				c.playSound("", false, 0, 11, 0, -1, vo, 0, 1, c.localscl, &c.pos[0], false, 0, 0, 0, 0, false, false)
			}
			c.koEchoTime++
		}
	}
}
func (c *Char) tick() {
	if c.scf(SCF_disabled) {
		return
	}
	if c.acttmp > 0 || (!c.pauseBool && c.hitPause() && c.asf(ASF_animatehitpause)) {
		if c.anim != nil && !c.asf(ASF_animfreeze) {
			c.anim.Action()
		}
	}
	if c.bindTime > 0 {
		if c.isBound() {
			if bt := sys.playerID(c.bindToId); bt != nil && !bt.pause() {
				c.bindTime -= 1
			}
		} else {
			if !c.pause() {
				c.bindTime -= 1
			}
		}
	}
	if c.cmd == nil {
		if c.keyctrl[0] {
			c.cmd = make([]CommandList, len(sys.chars))
			c.cmd[0].Buffer = NewCommandBuffer()
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
	if c.csf(CSF_gethit) && !c.hoKeepState {
		c.ss.changeMoveType(MT_H)
		// This flag prevents the previous move type from being changed twice
		c.ss.storeMoveType = true
		if c.hitPauseTime > 0 {
			c.ss.clearWw()
		}
		c.hitPauseTime = 0
		//c.targetDrop(-1, false) // GitHub #1148
		if c.hoIdx >= 0 && c.ho[c.hoIdx].forceair {
			c.ss.changeStateType(ST_A)
		}
		pn := c.playerNo
		if c.ghv.p2getp1state && !c.ghv.guarded {
			pn = c.ghv.playerNo
		}
		if c.stchtmp {
			// For Mugen compatibility, PrevStateNo returns these if the character is hit into a custom state (see GitHub #765)
			// This could be disabled if the state owner is an Ikemen character
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
				c.changeStateEx(5010, pn, -1, 0, "")
			default:
				c.changeStateEx(5020, pn, -1, 0, "")
			}
		}
		// Prepare down get hit offset
		if c.ss.stateType == ST_L && c.pos[1] == 0 && c.ghv.yvel != 0 {
			c.downHitOffset = true
		}
		// Change to HitOverride state
		if c.hoIdx >= 0 {
			c.stateChange1(c.ho[c.hoIdx].stateno, c.ho[c.hoIdx].playerNo)
		}
	}
	if !c.pause() {
		if c.hitPauseTime > 0 {
			c.hitPauseTime--
			if c.hitPauseTime == 0 {
				c.ss.clearWw()
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
		if !c.stchtmp {
			if c.helperIndex == 0 && (c.alive() || c.ss.no == 0) && c.life <= 0 &&
				c.ss.moveType != MT_H && !sys.gsf(GSF_globalnoko) && !c.asf(ASF_noko) &&
				(!c.ghv.guarded || !c.asf(ASF_noguardko)) {
				c.ghv.fallflag = true
				c.selfState(5030, -1, -1, 0, "") // Mugen sets control to 0 here
				c.ss.time = 1
			} else if c.ss.no == 5150 && c.ss.time >= 90 && c.alive() {
				c.selfState(5120, -1, -1, -1, "")
			}
		}
	}
	if !c.hitPause() && !c.pauseBool {
		// Set KO flag
		if c.life <= 0 && !sys.gsf(GSF_globalnoko) && !c.asf(ASF_noko) && (!c.ghv.guarded || !c.asf(ASF_noguardko)) {
			// KO sound
			if !sys.gsf(GSF_nokosnd) && c.alive() {
				vo := int32(100)
				c.playSound("", false, 0, 11, 0, -1, vo, 0, 1, c.localscl, &c.pos[0], false, 0, 0, 0, 0, false, false)
				if c.gi().data.ko.echo != 0 {
					c.koEchoTime = 1
				}
			}
			c.setSCF(SCF_ko)
			sys.charList.p2enemyDelete(c)
		}
	}
}
func (c *Char) cueDraw() {
	if c.helperIndex < 0 || c.scf(SCF_disabled) {
		return
	}
	x := c.pos[0] * c.localscl
	y := c.pos[1] * c.localscl
	xoff := x + c.offsetX()*c.localscl
	yoff := y + c.offsetY()*c.localscl
	xs := c.clsnScale[0] * c.facing
	ys := c.clsnScale[1]
	angle := c.clsnAngle * c.facing
	nhbtxt := ""
	// Debug Clsn display
	if sys.clsnDraw && c.curFrame != nil {
		// Add Clsn1
		if clsn := c.curFrame.Clsn1(); len(clsn) > 0 {
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
		if clsn := c.curFrame.Clsn2(); len(clsn) > 0 {
			hb, mtk := false, false
			if c.unhittableTime > 0 {
				mtk = true
			} else {
				for _, h := range c.hitby {
					if h.time != 0 {
						// If carrying invincibility from previous iterations
						if h.stack && flags != int32(ST_SCA)|int32(AT_ALL) {
							nhbtxt = "Stacked"
							hb = true
							mtk = false
							break
						}
						// If player-specific invincibility
						if h.playerno >= 0 || h.playerid >= 0 {
							nhbtxt = "Player-specific"
							hb = true
							mtk = false
							break
						}
						// Combine all NotHitBy flags
						if h.flag != 0 {
							flags &= h.flag
						}
					}
				}
				// If not stacked and not player-specific
				if nhbtxt == "" {
					if flags != int32(ST_SCA)|int32(AT_ALL) {
						hb = true
						mtk = flags&int32(ST_SCA) == 0 || flags&int32(AT_ALL) == 0
					}
				}
			}
			if c.scf(SCF_standby) {
				sys.debugc2stb.Add(clsn, xoff, yoff, xs, ys, angle)
			} else if mtk {
				// Add fully invincible Clsn2
				sys.debugc2mtk.Add(clsn, xoff, yoff, xs, ys, angle)
			} else if hb {
				// Add partially invincible Clsn2
				sys.debugc2hb.Add(clsn, xoff, yoff, xs, ys, angle)
			} else if c.inguarddist && c.scf(SCF_guard) {
				// Add guarding Clsn2
				sys.debugc2grd.Add(clsn, xoff, yoff, xs, ys, angle)
			} else {
				// Add regular Clsn2
				sys.debugc2.Add(clsn, xoff, yoff, xs, ys, angle)
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
			sys.debugcsize.Add(c.sizeBox, x, y, c.facing*c.localscl, c.localscl, 0)
		}
		// Add crosshair
		sys.debugch.Add([]float32{-1, -1, 1, 1}, x, y, 1, 1, 0)
	}
	// Prepare information for debug text
	if sys.debugDraw {
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
	// Add char sprite
	if c.anim != nil {
		pos := [2]float32{c.interPos[0]*c.localscl + c.offsetX()*c.localscl,
			c.interPos[1]*c.localscl + c.offsetY()*c.localscl}

		scl := [...]float32{c.facing * c.size.xscale * c.zScale * (320 / c.localcoord),
			c.size.yscale * c.zScale * (320 / c.localcoord)}

		// Apply Z axis perspective
		if sys.zmin != sys.zmax {
			pos[0] *= c.zScale
			pos[1] *= c.zScale
			pos[1] += c.interPos[2] * c.localscl
		}
		//if sys.zmin != sys.zmax {
		//	ratio := float32(1.618) // Possible stage parameter?
		//	pos[0] *= 1 + (ratio-1)*(c.zScale-1)
		//	pos[1] *= 1 + (ratio-1)*(c.zScale-1)
		//	pos[1] += c.interPos[2] * c.localscl
		//}

		agl := float32(0)
		if c.csf(CSF_angledraw) {
			agl = c.angle
			if agl == 0 {
				agl = 360
			} else if c.facing < 0 {
				agl *= -1
			}
		}
		rec := sys.tickNextFrame() && c.acttmp > 0

		//if rec {
		//	c.aimg.recAfterImg(sdf(), c.hitPause())
		//}

		//if c.gi().mugenver[0] != 1 && c.csf(CSF_angledraw) && !c.csf(CSF_trans) {
		//	c.setCSF(CSF_trans)
		//	c.alpha = [...]int32{255, 0}
		//}

		sd := &SprData{c.anim, c.getPalfx(), pos,
			scl, c.alpha, c.sprPriority + int32(c.pos[2]*c.localscl), Rotation{agl, 0, 0}, c.angleScale, false,
			c.playerNo == sys.superplayer, c.gi().mugenver[0] != 1, c.facing,
			c.localcoord / sys.chars[c.animPN][0].localcoord, // https://github.com/ikemen-engine/Ikemen-GO/issues/1459 and 1778
			0, 0, [4]float32{0, 0, 0, 0}}
		if !c.csf(CSF_trans) {
			sd.alpha[0] = -1
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
			sdwalp := int32(255)
			if c.csf(CSF_trans) {
				sdwalp = 255 - c.alpha[1]
			}
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
				charposz := c.interPos[2] * c.localscl
				sys.shadows.add(&ShadowSprite{sd, -1, sdwalp,
					[2]float32{c.shadowOffset[0] * c.localscl, (c.size.shadowoffset+c.shadowOffset[1])*c.localscl + sys.stage.sdw.yscale*charposz + charposz}, // Shadow offset
					[2]float32{c.reflectOffset[0] * c.localscl, c.reflectOffset[1]*c.localscl + sys.stage.reflection.yscale*charposz + charposz},              // Reflection offset
					c.offsetY()}) // Fade offset
			}
		}
	}
	if sys.tickNextFrame() {
		if sys.supertime < 0 && c.teamside != sys.superplayer&1 {
			c.superDefenseMul *= sys.superp2defmul
		}
		c.minus = 2
		c.oldPos = c.pos
		c.dustOldPos = c.pos[0]
	}
}

type CharList struct {
	runOrder, drawOrder []*Char
	idMap               map[int32]*Char
}

func (cl *CharList) clear() {
	*cl = CharList{idMap: make(map[int32]*Char)}
	sys.nextCharId = sys.helperMax
}
func (cl *CharList) add(c *Char) {
	// Append to run order
	cl.runOrder = append(cl.runOrder, c)
	c.index = int32(len(cl.runOrder))
	// If any entries in the draw order are empty, use that one
	i := 0
	for ; i < len(cl.drawOrder); i++ {
		if cl.drawOrder[i] == nil {
			cl.drawOrder[i] = c
			break
		}
	}
	// Otherwise appends to the end
	if i >= len(cl.drawOrder) {
		cl.drawOrder = append(cl.drawOrder, c)
	}
	cl.idMap[c.id] = c
}
func (cl *CharList) replace(dc *Char, pn int, idx int32) bool {
	var ok bool
	// Replace run order
	for i, c := range cl.runOrder {
		if c.playerNo == pn && c.helperIndex == idx {
			c.index = int32(i) + 1
			cl.runOrder[i] = dc
			ok = true
			break
		}
	}
	if ok {
		// Replace draw order
		for i, c := range cl.drawOrder {
			if c.playerNo == pn && c.helperIndex == idx {
				cl.drawOrder[i] = dc
				break
			}
		}
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
	for i, c := range cl.drawOrder {
		if c == dc {
			cl.drawOrder[i] = nil
			break
		}
	}
}

// Sort all characters into a list based on their processing order
func (cl *CharList) sortActionRunOrder() []int {

	sortedOrder := []int{}

	// Reset all run order values
	for i := 0; i < len(cl.runOrder); i++ {
		cl.runOrder[i].runorder = -1
	}

	// Sort characters with priority flag
	for i := 0; i < len(cl.runOrder); i++ {
		if cl.runOrder[i].runorder < 0 && cl.runOrder[i].asf(ASF_runfirst) {
			sortedOrder = append(sortedOrder, i)
			cl.runOrder[i].runorder = int32(len(sortedOrder))
		}
	}

	// Sort attacking players and helpers
	for i := 0; i < len(cl.runOrder); i++ {
		if cl.runOrder[i].runorder < 0 && !cl.runOrder[i].asf(ASF_runlast) &&
			cl.runOrder[i].ss.moveType == MT_A {
			sortedOrder = append(sortedOrder, i)
			cl.runOrder[i].runorder = int32(len(sortedOrder))
		}
	}

	// Sort idle players
	for i := 0; i < len(cl.runOrder); i++ {
		if cl.runOrder[i].runorder < 0 && !cl.runOrder[i].asf(ASF_runlast) &&
			cl.runOrder[i].helperIndex == 0 && cl.runOrder[i].ss.moveType == MT_I {
			sortedOrder = append(sortedOrder, i)
			cl.runOrder[i].runorder = int32(len(sortedOrder))
		}
	}

	// Sort remaining players
	for i := 0; i < len(cl.runOrder); i++ {
		if cl.runOrder[i].runorder < 0 && !cl.runOrder[i].asf(ASF_runlast) &&
			cl.runOrder[i].helperIndex == 0 {
			sortedOrder = append(sortedOrder, i)
			cl.runOrder[i].runorder = int32(len(sortedOrder))
		}
	}

	// Sort idle helpers
	for i := 0; i < len(cl.runOrder); i++ {
		if cl.runOrder[i].runorder < 0 && !cl.runOrder[i].asf(ASF_runlast) &&
			cl.runOrder[i].helperIndex != 0 && cl.runOrder[i].ss.moveType == MT_I {
			sortedOrder = append(sortedOrder, i)
			cl.runOrder[i].runorder = int32(len(sortedOrder))
		}
	}

	// Sort remaining helpers
	for i := 0; i < len(cl.runOrder); i++ {
		if cl.runOrder[i].runorder < 0 && !cl.runOrder[i].asf(ASF_runlast) &&
			cl.runOrder[i].helperIndex != 0 {
			sortedOrder = append(sortedOrder, i)
			cl.runOrder[i].runorder = int32(len(sortedOrder))
		}
	}

	// Sort anyone missed (RunLast flag)
	for i := 0; i < len(cl.runOrder); i++ {
		if cl.runOrder[i].runorder < 0 {
			sortedOrder = append(sortedOrder, i)
			cl.runOrder[i].runorder = int32(len(sortedOrder))
		}
	}

	// Reset priority flags as they are only needed during this function
	for i := 0; i < len(cl.runOrder); i++ {
		cl.runOrder[i].unsetASF(ASF_runfirst | ASF_runlast)
	}

	return sortedOrder
}

func (cl *CharList) action() {
	sys.commandUpdate()

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

func (cl *CharList) update() {
	ro := make([]*Char, len(cl.runOrder))
	copy(ro, cl.runOrder)
	for _, c := range ro {
		c.update()
		c.track()
	}
}

func (cl *CharList) hitDetection(getter *Char, proj bool) {
	if getter.scf(SCF_standby) || getter.scf(SCF_disabled) {
		return // Stop entire function if getter is disabled
	}

	// hitTypeGet() function definition start
	hitTypeGet := func(c *Char, hd *HitDef, pos [3]float32, projf float32, attackMul [4]float32) (hitType int32) {

		// Early exits
		if !proj && c.ss.stateType == ST_L && hd.reversal_attr <= 0 {
			c.hitdef.ltypehit = true
			return 0
		}

		if getter.stchtmp && getter.ss.sb.playerNo != hd.playerNo {
			if getter.csf(CSF_gethit) {
				if hd.p2stateno >= 0 {
					return 0
				}
			} else if getter.acttmp > 0 {
				return 0
			}
		}

		if hd.p1stateno >= 0 && (c.csf(CSF_gethit) || c.stchtmp && c.ss.sb.playerNo != hd.playerNo) {
			return 0
		}

		if getter.csf(CSF_gethit) && getter.ghv.attr&int32(AT_AT) != 0 {
			return 0
		}

		// Check if the enemy can guard this attack
		canguard := (proj || !c.asf(ASF_unguardable)) && getter.scf(SCF_guard) &&
			(!getter.csf(CSF_gethit) || getter.ghv.guarded)

		// Automatically choose high or low in case of auto guard
		if canguard && getter.asf(ASF_autoguard) && getter.acttmp > 0 && !getter.csf(CSF_gethit) {
			if int32(getter.ss.stateType)&hd.guardflag == 0 {
				if getter.ss.stateType == ST_S {
					// High to Low
					if int32(ST_C)&hd.guardflag != 0 && !getter.asf(ASF_nocrouchguard) {
						getter.ss.changeStateType(ST_C)
					}
				} else if getter.ss.stateType == ST_C {
					// Low to High
					if int32(ST_S)&hd.guardflag != 0 && !getter.asf(ASF_nostandguard) {
						getter.ss.changeStateType(ST_S)
					}
				}
			}
		}

		hitType = 1
		getter.ghv.kill = hd.kill
		// If enemy is guarding the correct way, "hitType" is set to "guard"
		if canguard && int32(getter.ss.stateType)&hd.guardflag != 0 {
			getter.ghv.kill = hd.guard_kill
			// We only switch to guard behavior if the enemy can survive guarding the attack
			if getter.life > getter.computeDamage(float64(hd.guarddamage), hd.guard_kill, false, attackMul[0], c, true) ||
				sys.gsf(GSF_globalnoko) || getter.asf(ASF_noko) || getter.asf(ASF_noguardko) {
				hitType = 2
			} else {
				getter.ghv.cheeseKO = true // TODO: find a better name then expose this variable
			}
		}

		// If any previous hit in the current frame will KO the enemy, the following ones will not prevent it
		if getter.ghv.damage >= getter.life {
			getter.ghv.kill = true
		}
		if hd.reversal_attr > 0 {
			hitType *= -1
		} else if getter.ss.stateType == ST_A {
			if hd.air_type == HT_None {
				hitType *= -1
			}
		} else if hd.ground_type == HT_None {
			hitType *= -1
		}

		p2s := false
		// Check HitOverride
		if !getter.stchtmp || !getter.csf(CSF_gethit) {
			c.mhv.overridden = false
			for i, ho := range getter.ho {
				if ho.time == 0 || ho.attr&hd.attr&^int32(ST_MASK) == 0 {
					continue
				}
				if proj {
					if ho.attr&hd.attr&int32(ST_MASK) == 0 {
						continue
					}
				} else {
					if ho.attr&int32(c.ss.stateType) == 0 {
						continue
					}
				}
				if !proj && Abs(hitType) == 1 &&
					(hd.p2stateno >= 0 || hd.p1stateno >= 0) {
					return 0
				}
				if ho.stateno >= 0 || ho.keepState {
					if ho.keepState {
						getter.hoKeepState = true
					}
					getter.hoIdx = i
					c.mhv.overridden = true
					break
				}
			}
			if !c.mhv.overridden {
				if Abs(hitType) == 1 && hd.p2stateno >= 0 {
					pn := getter.playerNo
					if hd.p2getp1state {
						pn = hd.playerNo
					}
					if getter.stateChange1(hd.p2stateno, pn) {
						getter.setCtrl(false)
						p2s = true
						getter.hoIdx = -1
					}
				}
			}
		}
		if !proj {
			c.hitdefTargetsBuffer = append(c.hitdefTargetsBuffer, getter.id)
			c.mhv.uniqhit = int32(len(c.hitdefTargets))
		}
		ghvset := !getter.csf(CSF_gethit) || !getter.stchtmp || p2s
		// Variables that are set even if Hitdef type is "None"
		if ghvset {
			if !proj {
				c.sprPriority = hd.p1sprpriority
			}
			getter.sprPriority = hd.p2sprpriority
			getter.ghv.hitid = hd.id
			getter.ghv.playerNo = hd.playerNo
			getter.ghv.id = hd.attackerID
			getter.ghv.groundtype = hd.ground_type
			getter.ghv.airtype = hd.air_type
			if getter.ss.stateType == ST_A {
				getter.ghv._type = getter.ghv.airtype
			} else {
				getter.ghv._type = getter.ghv.groundtype
			}
		}
		byf := c.facing
		if proj {
			byf = projf
		}
		if !proj && hitType == 1 {
			if hd.p1getp2facing != 0 {
				byf = getter.facing
				if hd.p1getp2facing < 0 {
					byf *= -1
				}
			} else if hd.p1facing < 0 {
				byf *= -1
			}
		}
		// HitDef connects
		if hitType > 0 {
			// Stop enemy's flagged sounds. In Mugen this only happens with channel 0
			if hitType == 1 {
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
				// Save existing variables that should persist or stack
				dmg, hdmg, gdmg := ghv.damage, ghv.hitdamage, ghv.guarddamage
				pwr, hpwr, gpwr := ghv.power, ghv.hitpower, ghv.guardpower
				dpnt, gpnt := ghv.dizzypoints, ghv.guardpoints
				fall, hc, gc, fc, by := ghv.fallflag, ghv.hitcount, ghv.guardcount, ghv.fallcount, ghv.hitBy
				drec := ghv.down_recovertime
				kill := ghv.kill
				cheese := ghv.cheeseKO
				// Clear variables
				// TODO: It's possible that this doesn't need to happen, like ReversalDef
				ghv.clear(getter)
				// Restore persistent variables
				ghv.hitBy = by
				ghv.damage = dmg
				ghv.hitdamage = hdmg
				ghv.guarddamage = gdmg
				ghv.power = pwr
				ghv.hitpower = hpwr
				ghv.guardpower = gpwr
				ghv.dizzypoints = dpnt
				ghv.guardpoints = gpnt
				ghv.down_recovertime = drec
				ghv.kill = kill
				ghv.cheeseKO = cheese
				// Update variables
				ghv.attr = hd.attr
				ghv.hitid = hd.id
				ghv.playerNo = hd.playerNo
				ghv.id = hd.attackerID
				ghv.xaccel = hd.xaccel * (c.localscl / getter.localscl) * -byf
				ghv.yaccel = hd.yaccel * (c.localscl / getter.localscl)
				ghv.zaccel = hd.zaccel * (c.localscl / getter.localscl)
				ghv.groundtype = hd.ground_type
				ghv.airtype = hd.air_type
				if hd.forcenofall {
					fall = false
				}
				if getter.ss.stateType == ST_A {
					ghv._type = ghv.airtype
				} else {
					ghv._type = ghv.groundtype
				}
				if !math.IsNaN(float64(hd.score[0])) {
					ghv.score = hd.score[0]
				}
				ghv.fatal = false
				// If attack is guarded
				if hitType == 2 {
					ghv.guarded = true
					ghv.hitshaketime = Max(0, hd.guard_shaketime)
					ghv.hittime = Max(0, hd.guard_hittime)
					ghv.slidetime = hd.guard_slidetime
					if getter.ss.stateType == ST_A {
						ghv.ctrltime = hd.airguard_ctrltime
						ghv.xvel = hd.airguard_velocity[0] * (c.localscl / getter.localscl) * -byf
						ghv.yvel = hd.airguard_velocity[1] * (c.localscl / getter.localscl)
						ghv.zvel = hd.airguard_velocity[2] * (c.localscl / getter.localscl)
					} else {
						ghv.ctrltime = hd.guard_ctrltime
						ghv.xvel = hd.guard_velocity[0] * (c.localscl / getter.localscl) * -byf
						// Mugen does not accept a Y component for ground guard velocity
						// But since we're adding Z to the other parameters, let's add Y here as well to keep things consistent
						ghv.yvel = hd.guard_velocity[1] * (c.localscl / getter.localscl)
						ghv.zvel = hd.guard_velocity[2] * (c.localscl / getter.localscl)
					}
					ghv.hitcount = hc
					ghv.guardcount = gc + 1
				} else {
					ghv.hitshaketime = Max(0, hd.shaketime)
					ghv.slidetime = hd.ground_slidetime
					ghv.p2getp1state = hd.p2getp1state
					ghv.forcestand = hd.forcestand != 0
					ghv.forcecrouch = hd.forcecrouch != 0

					ghv.fall = hd.fall // The group, not the flag
					getter.fallTime = 0
					ghv.fall.envshake_ampl = int32(float32(hd.fall.envshake_ampl) * (c.localscl / getter.localscl))
					ghv.fall.xvelocity = hd.fall.xvelocity * (c.localscl / getter.localscl)
					ghv.fall.yvelocity = hd.fall.yvelocity * (c.localscl / getter.localscl)
					ghv.fall.zvelocity = hd.fall.zvelocity * (c.localscl / getter.localscl)

					if getter.ss.stateType == ST_A {
						ghv.hittime = hd.air_hittime
						// Note: ctrl time is not affected on hit in Mugen
						// This is further proof that gethitvars don't need to be reset above
						ghv.ctrltime = hd.air_hittime
						ghv.xvel = hd.air_velocity[0] * (c.localscl / getter.localscl) * -byf
						ghv.yvel = hd.air_velocity[1] * (c.localscl / getter.localscl)
						ghv.zvel = hd.air_velocity[2] * (c.localscl / getter.localscl)
						ghv.fallflag = hd.air_fall
					} else if getter.ss.stateType == ST_L {
						ghv.hittime = hd.down_hittime
						ghv.ctrltime = hd.down_hittime
						ghv.fallflag = hd.ground_fall
						if getter.pos[1] == 0 {
							ghv.xvel = hd.down_velocity[0] * (c.localscl / getter.localscl) * -byf
							ghv.yvel = hd.down_velocity[1] * (c.localscl / getter.localscl)
							ghv.zvel = hd.down_velocity[2] * (c.localscl / getter.localscl)
							if !hd.down_bounce && ghv.yvel != 0 {
								ghv.fall.xvelocity = float32(math.NaN())
								ghv.fall.yvelocity = 0
								ghv.fall.zvelocity = float32(math.NaN())
							}
						} else {
							ghv.xvel = hd.air_velocity[0] * (c.localscl / getter.localscl) * -byf
							ghv.yvel = hd.air_velocity[1] * (c.localscl / getter.localscl)
							ghv.zvel = hd.air_velocity[1] * (c.localscl / getter.localscl)
						}
					} else {
						ghv.ctrltime = hd.ground_hittime
						ghv.xvel = hd.ground_velocity[0] * (c.localscl / getter.localscl) * -byf
						ghv.yvel = hd.ground_velocity[1] * (c.localscl / getter.localscl)
						ghv.zvel = hd.ground_velocity[2] * (c.localscl / getter.localscl)
						ghv.fallflag = hd.ground_fall
						if ghv.fallflag && ghv.yvel == 0 {
							// Mugen does this as some form of internal workaround
							ghv.yvel = -0.001 * (c.localscl / getter.localscl)
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
					if cmb {
						ghv.hitcount = hc + 1
					} else {
						ghv.hitcount = 1
					}
					ghv.guardcount = gc
					ghv.fallcount = fc
					ghv.fallflag = ghv.fallflag || fall // If falling now or before the hit
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
					// This compensates for characters being able to guard one frame sooner in Ikemen than in Mugen
					if c.stWgi().ikemenver[0] == 0 && c.stWgi().ikemenver[1] == 0 && ghv.hittime > 0 {
						ghv.hittime += 1
					}
				}
				// Save velocities regardless of statetype
				ghv.ground_velocity[0] = hd.ground_velocity[0] * (c.localscl / getter.localscl) * -byf
				ghv.ground_velocity[1] = hd.ground_velocity[1] * (c.localscl / getter.localscl)
				ghv.ground_velocity[2] = hd.ground_velocity[2] * (c.localscl / getter.localscl)
				ghv.air_velocity[0] = hd.air_velocity[0] * (c.localscl / getter.localscl) * -byf
				ghv.air_velocity[1] = hd.air_velocity[1] * (c.localscl / getter.localscl)
				ghv.air_velocity[2] = hd.air_velocity[2] * (c.localscl / getter.localscl)
				ghv.down_velocity[0] = hd.down_velocity[0] * (c.localscl / getter.localscl) * -byf
				ghv.down_velocity[1] = hd.down_velocity[1] * (c.localscl / getter.localscl)
				ghv.down_velocity[2] = hd.down_velocity[2] * (c.localscl / getter.localscl)
				ghv.guard_velocity[0] = hd.guard_velocity[0] * (c.localscl / getter.localscl) * -byf
				ghv.guard_velocity[1] = hd.guard_velocity[1] * (c.localscl / getter.localscl)
				ghv.guard_velocity[2] = hd.guard_velocity[2] * (c.localscl / getter.localscl)
				ghv.airguard_velocity[0] = hd.airguard_velocity[0] * (c.localscl / getter.localscl) * -byf
				ghv.airguard_velocity[1] = hd.airguard_velocity[1] * (c.localscl / getter.localscl)
				ghv.airguard_velocity[2] = hd.airguard_velocity[2] * (c.localscl / getter.localscl)
				ghv.airanimtype = hd.air_animtype
				ghv.groundanimtype = hd.animtype
				ghv.animtype = getter.gethitAnimtype() // This must be placed after ghv.yvel
				ghv.priority = hd.priority
				byPos := c.pos
				if proj {
					for i, p := range pos {
						byPos[i] += p
					}
				}
				snap := [...]float32{float32(math.NaN()), float32(math.NaN()), float32(math.NaN())}
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
				if hitType == 1 || getter.ss.stateType == ST_A {
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
				if !math.IsNaN(float64(snap[0])) {
					ghv.xoff = snap[0]*(c.localscl/getter.localscl) - getter.pos[0]
				}
				if !math.IsNaN(float64(snap[1])) {
					ghv.yoff = snap[1]*(c.localscl/getter.localscl) - getter.pos[1]
				}
				if !math.IsNaN(float64(snap[2])) {
					ghv.zoff = snap[2]*(c.localscl/getter.localscl) - getter.pos[2]
				}
				if hd.snaptime != 0 && getter.hoIdx < 0 {
					getter.setBindToId(c)
					getter.setBindTime(hd.snaptime + Btoi(hd.snaptime > 0 && !c.pause()))
					getter.bindFacing = 0
					if !math.IsNaN(float64(snap[0])) {
						getter.bindPos[0] = hd.mindist[0] * (c.localscl / getter.localscl)
					} else {
						getter.bindPos[0] = float32(math.NaN())
					}
					if !math.IsNaN(float64(snap[1])) &&
						(hitType == 1 || getter.ss.stateType == ST_A) {
						getter.bindPos[1] = hd.mindist[1] * (c.localscl / getter.localscl)
					} else {
						getter.bindPos[1] = float32(math.NaN())
					}
					if !math.IsNaN(float64(snap[2])) {
						getter.bindPos[2] = hd.mindist[2] * (c.localscl / getter.localscl)
					} else {
						getter.bindPos[2] = float32(math.NaN())
					}
				} else if getter.bindToId == c.id {
					getter.setBindTime(0)
				}
			}
			if sys.super > 0 {
				getter.superMovetime =
					Max(getter.superMovetime, getter.ghv.hitshaketime)
			} else if sys.pause > 0 {
				getter.pauseMovetime =
					Max(getter.pauseMovetime, getter.ghv.hitshaketime)
			}
			if !p2s && !getter.csf(CSF_gethit) {
				getter.stchtmp = false
			}
			getter.setCSF(CSF_gethit)
			getter.ghv.frame = true
			// In Mugen, having any HitOverride active allows GetHitVar Damage to exceed the remaining life
			bnd := true
			for _, ho := range getter.ho {
				if ho.time != 0 {
					bnd = false
					break
				}
			}
			// Damage on hit
			if hitType == 1 {
				// Life
				if !getter.asf(ASF_nohitdamage) {
					getter.ghv.damage += getter.computeDamage(
						float64(hd.hitdamage), getter.ghv.kill, false, attackMul[0], c, bnd)
				}
				// Red life
				if !getter.asf(ASF_noredlifedamage) {
					getter.ghv.redlife += getter.computeDamage(
						float64(hd.hitredlife), true, false, attackMul[1], c, bnd)
				}
				// Dizzy points
				if !getter.asf(ASF_nodizzypointsdamage) && !getter.scf(SCF_dizzy) {
					getter.ghv.dizzypoints += getter.computeDamage(
						float64(hd.dizzypoints), true, false, attackMul[2], c, false)
				}
			}
			// Damage on guard
			if hitType == 2 {
				// Life
				if !getter.asf(ASF_noguarddamage) {
					getter.ghv.damage += getter.computeDamage(
						float64(hd.guarddamage), getter.ghv.kill, false, attackMul[0], c, bnd)
				}
				// Red life
				if !getter.asf(ASF_noredlifedamage) {
					getter.ghv.redlife += getter.computeDamage(
						float64(hd.guardredlife), true, false, attackMul[1], c, bnd)
				}
				// Guard points
				if !getter.asf(ASF_noguardpointsdamage) {
					getter.ghv.guardpoints += getter.computeDamage(
						float64(hd.guardpoints), true, false, attackMul[3], c, false)
				}
			}
			// Absolute values
			// These do not affect the player and are only used in GetHitVar
			getter.ghv.hitpower += hd.hitgivepower
			getter.ghv.guardpower += hd.guardgivepower
			getter.ghv.hitdamage += getter.computeDamage(
				float64(hd.hitdamage), true, false, attackMul[0], c, false)
			getter.ghv.guarddamage += getter.computeDamage(
				float64(hd.guarddamage), true, false, attackMul[0], c, false)
			getter.ghv.hitredlife += getter.computeDamage(
				float64(hd.hitredlife), true, false, attackMul[1], c, bnd)
			getter.ghv.guardredlife += getter.computeDamage(
				float64(hd.guardredlife), true, false, attackMul[1], c, bnd)
			// Hit behavior on KO
			if ghvset && getter.ghv.damage >= getter.life {
				if getter.ghv.kill || !getter.alive() {
					getter.ghv.fatal = true
					getter.ghv.fallflag = true
					getter.ghv.animtype = getter.gethitAnimtype() // Update to fall anim type
					if getter.kovelocity && !getter.asf(ASF_nokovelocity) {
						if getter.ss.stateType == ST_A {
							if getter.ghv.xvel < 0 {
								getter.ghv.xvel += getter.gi().velocity.air.gethit.ko.add[0]
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
							if getter.ghv.xvel < 0 {
								getter.ghv.xvel += getter.gi().velocity.ground.gethit.ko.add[0]
							}
							if getter.ghv.yvel <= 0 {
								getter.ghv.yvel += getter.gi().velocity.ground.gethit.ko.add[1]
								if getter.ghv.yvel > getter.gi().velocity.ground.gethit.ko.ymin {
									getter.ghv.yvel = getter.gi().velocity.ground.gethit.ko.ymin
								}
							}
						}
					}
				} else {
					getter.ghv.damage = getter.life - 1
				}
			}
		}
		hitspark := func(p1, p2 *Char, animNo int32, ffx string, sparkangle float32) {
			// This is mostly for offset in projectiles
			off := [3]float32{pos[0], pos[1], pos[2]}

			// Get reference position
			if !proj {
				off[0] = p2.pos[0]*p2.localscl - p1.pos[0]*p1.localscl
				if (p1.facing < 0) != (p2.facing < 0) {
					off[0] += p2.facing * p2.width[0] * p2.localscl
				} else {
					off[0] -= p2.facing * p2.width[1] * p2.localscl
				}
				off[2] = p2.pos[2]*p2.localscl - p1.pos[2]*p1.localscl
			}
			off[0] *= p1.facing

			// Apply sparkxy
			if proj {
				off[0] *= c.localscl
				off[1] *= c.localscl
				off[2] *= c.localscl
				off[0] += hd.sparkxy[0] * projf * p1.facing * c.localscl
			} else {
				off[0] -= hd.sparkxy[0] * c.localscl
			}
			off[1] += hd.sparkxy[1] * c.localscl

			// Reversaldef spark (?)
			if c.id != p1.id {
				off[1] += p1.hitdef.sparkxy[1] * c.localscl
			}

			// Save hitspark position
			c.mhv.sparkxy[0] = off[0]
			c.mhv.sparkxy[1] = off[1]

			if e, i := c.newExplod(); e != nil {
				e.anim = c.getAnim(animNo, ffx, true)
				e.layerno = 1 // e.ontop = true
				e.sprpriority = math.MinInt32
				e.ownpal = true
				e.relativePos = [...]float32{off[0], off[1], off[2]}
				e.supermovetime = -1
				e.pausemovetime = -1
				e.localscl = 1
				if ffx == "" || ffx == "s" {
					e.scale = [...]float32{c.localscl, c.localscl}
				} else if e.anim != nil {
					e.anim.start_scale[0] *= c.localscl
					e.anim.start_scale[1] *= c.localscl
				}
				e.setPos(p1)
				e.anglerot[0] = sparkangle
				c.insertExplod(i)
			}
		}
		if Abs(hitType) == 1 {
			if hd.sparkno >= 0 {
				if hd.reversal_attr > 0 {
					hitspark(getter, c, hd.sparkno, hd.sparkno_ffx, hd.sparkangle)
				} else {
					hitspark(c, getter, hd.sparkno, hd.sparkno_ffx, hd.sparkangle)
				}
			}
			if hd.hitsound[0] >= 0 {
				vo := int32(100)
				c.playSound(hd.hitsound_ffx, false, 0, hd.hitsound[0], hd.hitsound[1],
					hd.hitsound_channel, vo, 0, 1, getter.localscl, &getter.pos[0], true, 0, 0, 0, 0, false, false)
			}
			if hitType > 0 {
				c.powerAdd(hd.hitgetpower)
				if getter.player {
					getter.powerAdd(hd.hitgivepower)
					getter.ghv.power += hd.hitgivepower
				}
				if !math.IsNaN(float64(hd.score[0])) {
					c.scoreAdd(hd.score[0])
				}
				if getter.player {
					if !math.IsNaN(float64(hd.score[1])) {
						getter.scoreAdd(hd.score[1])
					}
				}
				c.counterHit = getter.ss.moveType == MT_A
			}
			if (ghvset || getter.csf(CSF_gethit)) && getter.hoIdx < 0 &&
				!(c.hitdef.air_type == HT_None && getter.ss.stateType == ST_A || getter.ss.stateType != ST_A && c.hitdef.ground_type == HT_None) {
				getter.receivedHits += hd.numhits
				if c.teamside != -1 {
					sys.lifebar.co[c.teamside].combo += hd.numhits
				}
			}
		} else {
			if hd.guard_sparkno >= 0 {
				if hd.reversal_attr > 0 {
					hitspark(getter, c, hd.guard_sparkno, hd.guard_sparkno_ffx, hd.guard_sparkangle)
				} else {
					hitspark(c, getter, hd.guard_sparkno, hd.guard_sparkno_ffx, hd.guard_sparkangle)
				}
			}
			if hd.guardsound[0] >= 0 {
				vo := int32(100)
				c.playSound(hd.guardsound_ffx, false, 0, hd.guardsound[0], hd.guardsound[1],
					hd.guardsound_channel, vo, 0, 1, getter.localscl, &getter.pos[0], true, 0, 0, 0, 0, false, false)
			}
			if hitType > 0 {
				c.powerAdd(hd.guardgetpower)
				if getter.player {
					getter.powerAdd(hd.guardgivepower)
					getter.ghv.power += hd.guardgivepower
				}
			}
		}
		if !ghvset {
			return
		}
		getter.p1facing = 0
		invertXvel := func(byf float32) {
			if !proj {
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
		if getter.hoIdx >= 0 {
			invertXvel(byf)
			return
		}
		if !proj && hd.hitonce > 0 {
			// Drop all targets except the current one
			c.targetDrop(-1, getter.id, true)
		}
		// Juggle points inheriting
		if c.helperIndex != 0 && c.inheritJuggle != 0 {
			// Update parent's or root's target list and juggle points
			sendJuggle := func(origin *Char) {
				origin.addTarget(getter.id)
				jg := origin.gi().data.airjuggle
				for _, v := range getter.ghv.hitBy {
					if len(v) >= 2 && (v[0] == origin.id || v[0] == c.id) && v[1] < jg {
						jg = v[1]
					}
				}
				getter.ghv.dropId(origin.id)
				getter.ghv.hitBy = append(getter.ghv.hitBy, [...]int32{origin.id, jg - c.juggle})
			}
			if c.inheritJuggle == 1 && c.parent() != nil {
				sendJuggle(c.parent())
			} else if c.inheritJuggle == 2 && c.root() != nil {
				sendJuggle(c.root())
			}
		}
		c.addTarget(getter.id)
		getter.ghv.addId(c.id, c.gi().data.airjuggle)
		if Abs(hitType) == 1 {
			if !proj && (hd.p1getp2facing != 0 || hd.p1facing < 0) &&
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
			if getter.ghv.fallflag {
				if !c.asf(ASF_nojugglecheck) {
					jug := &getter.ghv.hitBy[len(getter.ghv.hitBy)-1][1]
					if proj {
						*jug -= hd.air_juggle
					} else {
						*jug -= c.juggle
					}
				}
				// Juggle cost is reset regardless of NoJuggleCheck
				// https://github.com/ikemen-engine/Ikemen-GO/issues/1905
				c.juggle = 0
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
				sys.envShake.setDefaultPhase()
			}
			// Set corner push
			// In Mugen it is only set if the enemy is already in the corner before the hit
			// In Ikemen it is set regardless, with corner distance being checked later
			if hitType > 0 && !proj {
				switch getter.ss.stateType {
				case ST_S, ST_C:
					c.cornerVelOff = hd.ground_cornerpush_veloff * c.facing
				case ST_A:
					c.cornerVelOff = hd.air_cornerpush_veloff * c.facing
				case ST_L:
					c.cornerVelOff = hd.down_cornerpush_veloff * c.facing
				}
			}
		} else {
			if hitType > 0 && !proj {
				switch getter.ss.stateType {
				case ST_S, ST_C:
					c.cornerVelOff = hd.guard_cornerpush_veloff * c.facing
				case ST_A:
					c.cornerVelOff = hd.airguard_cornerpush_veloff * c.facing
				}
			}
		}
		invertXvel(byf)
		return
	}

	// Ignore Standby and Disabled Chars.
	if getter.scf(SCF_standby) || getter.scf(SCF_disabled) {
		return
	}

	// Projectile hitting player check
	// TODO: Disable projectiles if player is disabled?
	if proj {
		for i, pr := range sys.projs {
			if len(sys.projs[i]) == 0 {
				continue
			}
			c := sys.chars[i][0]
			orgatktmp := c.atktmp
			c.atktmp = -1
			ap_projhit := false
			for j := range pr {
				p := &pr[j]

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

				dist := (getter.pos[0]*getter.localscl - (p.pos[0])*p.localscl) * p.facing

				// Projectile guard distance
				if !p.platform && p.hitdef.attr > 0 { // https://github.com/ikemen-engine/Ikemen-GO/issues/1445
					if p.hitdef.guard_dist[0] < 0 {
						if dist <= float32(c.size.proj.attack.dist.front)*c.localscl &&
							dist >= -float32(c.size.proj.attack.dist.back)*c.localscl {
							getter.inguarddist = true
						}
					} else {
						if dist <= float32(p.hitdef.guard_dist[0]) &&
							dist >= -float32(p.hitdef.guard_dist[1]) {
							getter.inguarddist = true
						}
					}
				}

				if p.platform {
					// Check if the character is above the platform's surface
					if getter.pos[1]*getter.localscl-getter.vel[1]*getter.localscl <= (p.pos[1]+p.platformHeight[1])*p.localscl &&
						getter.platformPosY*getter.localscl >= (p.pos[1]+p.platformHeight[0])*p.localscl {
						angleSinValue := float32(math.Sin(float64(p.platformAngle) / 180 * math.Pi))
						angleCosValue := float32(math.Cos(float64(p.platformAngle) / 180 * math.Pi))
						oldDist := (getter.oldPos[0]*getter.localscl - (p.pos[0])*p.localscl) * p.facing
						onPlatform := func(protrude bool) {
							getter.platformPosY = ((p.pos[1]+p.platformHeight[0]+p.velocity[1])*p.localscl - angleSinValue*(oldDist/angleCosValue)) / getter.localscl
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
						if dist >= (p.platformWidth[0]*angleCosValue)*p.localscl && dist <= (p.platformWidth[1]*angleCosValue)*p.localscl {
							onPlatform(false)
						} else if p.platformFence && oldDist >= (p.platformWidth[0]*angleCosValue)*p.localscl &&
							oldDist <= (p.platformWidth[1]*angleCosValue)*p.localscl {
							onPlatform(true)
						}
					}
				}

				// Cancel a projectile with hitflag P
				if getter.atktmp != 0 && (getter.hitdef.affectteam == 0 ||
					(p.hitdef.teamside-1 != getter.teamside) == (getter.hitdef.affectteam > 0)) &&
					getter.hitdef.hitflag&int32(ST_P) != 0 &&
					getter.projClsnCheck(p, 1, 2) &&
					sys.zAxisOverlap(getter.pos[2], getter.hitdef.attack.depth[0], getter.hitdef.attack.depth[1], getter.localscl,
						p.pos[2], p.hitdef.attack.depth[0], p.hitdef.attack.depth[1], p.localscl) {
					if getter.hitdef.p1stateno >= 0 && getter.stateChange1(getter.hitdef.p1stateno, getter.hitdef.playerNo) {
						getter.setCtrl(false)
					}
					p.hits = -2
					sys.cgi[i].pctype = PC_Cancel
					sys.cgi[i].pctime = 0
					sys.cgi[i].pcid = p.id
					getter.hitdefContact = true
					//getter.mhv.frame = true
					continue
				}
				if !(getter.stchtmp && (getter.csf(CSF_gethit) || getter.acttmp > 0)) &&
					// Projectiles always check juggle points even if the enemy is not already a target
					(c.asf(ASF_nojugglecheck) || getter.ghv.getJuggle(c.id, c.gi().data.airjuggle) >= p.hitdef.air_juggle) &&
					(!ap_projhit || p.hitdef.attr&int32(AT_AP) == 0) &&
					(p.hitpause <= 0 || p.contactflag) && p.curmisstime <= 0 && p.hitdef.hitonce >= 0 &&
					getter.hittableByChar(&p.hitdef, c, ST_N, true) {
					orghittmp := getter.hittmp
					if getter.csf(CSF_gethit) {
						getter.hittmp = int8(Btoi(getter.ghv.fallflag)) + 1
					}

					if getter.projClsnCheck(p, p.hitdef.p2clsncheck, 1) &&
						sys.zAxisOverlap(p.pos[2], p.hitdef.attack.depth[0], p.hitdef.attack.depth[1], p.localscl,
							getter.pos[2], getter.size.depth, getter.size.depth, getter.localscl) {

						if ht := hitTypeGet(c, &p.hitdef, [...]float32{p.pos[0] - c.pos[0]*(c.localscl/p.localscl),
							p.pos[1] - c.pos[1]*(c.localscl/p.localscl), p.pos[2] - c.pos[2]*(c.localscl/p.localscl)},
							p.facing, p.parentAttackmul); ht != 0 {

							p.contactflag = true
							if Abs(ht) == 1 {
								sys.cgi[i].pctype = PC_Hit
								p.hitpause = Max(0, p.hitdef.pausetime-Btoi(c.gi().mugenver[0] == 0)) // Winmugen projectiles are 1 frame short on hitpauses
							} else {
								sys.cgi[i].pctype = PC_Guarded
								p.hitpause = Max(0, p.hitdef.guard_pausetime-Btoi(c.gi().mugenver[0] == 0))
							}
							sys.cgi[i].pctime = 0
							sys.cgi[i].pcid = p.id
						}
						// In MUGEN, it seems that projectiles with the "P" attribute in their "attr" only hit once on frame 1.
						// This flag prevents two projectiles of the same player from hitting in the same frame
						// In Mugen, projectiles (sctrl) give 1F of projectile invincibility to the getter instead. Timer persists during (super)pause
						if p.hitdef.attr&int32(AT_AP) != 0 {
							ap_projhit = true
						}
					}
					getter.hittmp = orghittmp
				}
			}
			c.atktmp = orgatktmp
		}
	}

	// Player check
	if !proj {
		getter.inguarddist = false
		getter.unsetCSF(CSF_gethit)
		getter.enemyNearClear()
		for _, c := range cl.runOrder {

			// Stop current iteration if this char is disabled
			if c.scf(SCF_standby) || c.scf(SCF_disabled) {
				continue
			}

			if c.atktmp != 0 && c.id != getter.id && (c.hitdef.affectteam == 0 ||
				((getter.teamside != c.hitdef.teamside-1) == (c.hitdef.affectteam > 0) && c.hitdef.teamside >= 0) ||
				((getter.teamside != c.teamside) == (c.hitdef.affectteam > 0) && c.hitdef.teamside < 0)) {

				dist := -getter.distX(c, getter) * c.facing

				// Default guard distance
				if c.ss.moveType == MT_A && c.hitdef.guard_dist[0] < 0 &&
					dist <= c.attackDist[0]*(c.localscl/getter.localscl) &&
					dist >= -c.attackDist[1]*(c.localscl/getter.localscl) {
					getter.inguarddist = true
				}

				if c.helperIndex != 0 {
					//inherit parent's or root's juggle points
					if c.inheritJuggle == 1 && c.parent() != nil {
						for _, v := range getter.ghv.hitBy {
							if v[0] == c.parent().id {
								getter.ghv.addId(c.id, v[1])
								break
							}
						}
					} else if c.inheritJuggle == 2 && c.root() != nil {
						for _, v := range getter.ghv.hitBy {
							if v[0] == c.root().id {
								getter.ghv.addId(c.id, v[1])
								break
							}
						}
					}
				}

				if c.hitdef.hitonce >= 0 && !c.hasTargetOfHitdef(getter.id) &&
					(c.hitdef.reversal_attr <= 0 || !getter.hasTargetOfHitdef(c.id)) &&
					(getter.hittmp < 2 || c.asf(ASF_nojugglecheck) || !c.hasTarget(getter.id) || getter.ghv.getJuggle(c.id, c.gi().data.airjuggle) >= c.juggle) &&
					getter.hittableByChar(&c.hitdef, c, c.ss.stateType, false) {

					// Guard distance
					if c.ss.moveType == MT_A &&
						dist <= float32(c.hitdef.guard_dist[0]) &&
						dist >= -float32(c.hitdef.guard_dist[1]) {
						getter.inguarddist = true
					}

					// Z axis check
					// Reversaldef checks attack depth vs attack depth
					zok := true
					if c.hitdef.reversal_attr > 0 {
						zok = sys.zAxisOverlap(c.pos[2], c.hitdef.attack.depth[0], c.hitdef.attack.depth[1], c.localscl,
							getter.pos[2], getter.hitdef.attack.depth[0], getter.hitdef.attack.depth[1], getter.localscl)
					} else {
						zok = sys.zAxisOverlap(c.pos[2], c.hitdef.attack.depth[0], c.hitdef.attack.depth[1], c.localscl,
							getter.pos[2], getter.size.depth, getter.size.depth, getter.localscl)
					}

					if zok && c.clsnCheck(getter, 1, c.hitdef.p2clsncheck, true, false) {
						if ht := hitTypeGet(c, &c.hitdef, [3]float32{}, 0, c.attackMul); ht != 0 {
							mvh := ht > 0 || c.hitdef.reversal_attr > 0
							if Abs(ht) == 1 {
								if mvh {
									c.mctype = MC_Hit
									c.mctime = -1
								}
								// ReversalDef connects
								if c.hitdef.reversal_attr > 0 {
									// ReversalDef seems to set an arbitrary set of get hit variables in Mugen
									c.powerAdd(c.hitdef.hitgetpower)
									getter.hitdef.hitflag = 0
									getter.mctype = MC_Reversed
									getter.mctime = -1
									getter.hitdefContact = true
									getter.mhv.frame = true
									getter.mhv.id = c.id
									getter.mhv.playerNo = c.playerNo
									getter.hitdef.hitonce = -1 // Neutralize Hitdef
									getter.unhittableTime = 1  // Reversaldef makes the target invincible for 1 frame (but not the attacker)

									// In Mugen, ReversalDef does not clear the enemy's GetHitVars
									// https://github.com/ikemen-engine/Ikemen-GO/issues/1891
									// fall, by := getter.ghv.fallflag, getter.ghv.hitBy
									// getter.ghv.clear(getter)
									// getter.ghv.hitBy = by
									// getter.ghv.fall = c.hitdef.fall

									getter.ghv.attr = c.hitdef.attr
									getter.ghv.hitid = c.hitdef.id
									getter.ghv.playerNo = c.playerNo
									getter.ghv.id = c.id

									getter.ghv.fall = c.hitdef.fall // The group, not the flag
									getter.fallTime = 0
									// https://github.com/ikemen-engine/Ikemen-GO/issues/2012
									getter.ghv.fall.envshake_ampl = int32(float32(c.hitdef.fall.envshake_ampl) * (c.localscl / getter.localscl))
									getter.ghv.fall.xvelocity = c.hitdef.fall.xvelocity * (c.localscl / getter.localscl)
									getter.ghv.fall.yvelocity = c.hitdef.fall.yvelocity * (c.localscl / getter.localscl)
									getter.ghv.fall.zvelocity = c.hitdef.fall.zvelocity * (c.localscl / getter.localscl)

									if c.hitdef.forcenofall {
										getter.ghv.fallflag = false
									} else if !getter.ghv.fallflag {
										if getter.ss.stateType == ST_A {
											getter.ghv.fallflag = c.hitdef.air_fall
										} else {
											getter.ghv.fallflag = c.hitdef.ground_fall
										}
									}

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
										getter.hitPauseTime = Max(1, c.hitdef.shaketime+Btoi(c.gi().mugenver[0] == 1))
									}
								}
								if !c.csf(CSF_gethit) && (getter.ss.stateType == ST_A && c.hitdef.air_type != HT_None ||
									getter.ss.stateType != ST_A && c.hitdef.ground_type != HT_None) {
									c.hitPauseTime = Max(1, c.hitdef.pausetime+Btoi(c.gi().mugenver[0] == 1))
									// Attacker hitpauses were off by 1 frame in Winmugen. Mugen 1.0 fixed it by compensating
								}
								c.uniqHitCount++
							} else {
								if mvh {
									c.mctype = MC_Guarded
									c.mctime = -1
								}
								if !c.csf(CSF_gethit) {
									c.hitPauseTime = Max(1, c.hitdef.guard_pausetime+Btoi(c.gi().mugenver[0] == 1))
								}
							}
							if c.hitdef.hitonce > 0 {
								c.hitdef.hitonce = -1
							}
							c.hitdefContact = true
							c.mhv.frame = true
							c.mhv.id = getter.id
							c.mhv.playerNo = getter.playerNo
						}
					}
				}
			}
		}
	}
}

func (cl *CharList) pushDetection(getter *Char) {
	var gxmin, gxmax float32
	if !getter.csf(CSF_playerpush) || getter.scf(SCF_standby) || getter.scf(SCF_disabled) {
		return // Stop entire function if getter won't push
	}
	for _, c := range cl.runOrder {
		if !c.csf(CSF_playerpush) || c.teamside == getter.teamside || c.scf(SCF_standby) || c.scf(SCF_disabled) {
			continue // Stop current iteration if char won't push
		}

		// Pushbox vertical size and coordinates
		ctop := (c.pos[1] + c.sizeBox[1]) * c.localscl
		cbot := (c.pos[1] + c.sizeBox[3]) * c.localscl
		gtop := (getter.pos[1] + getter.sizeBox[1]) * getter.localscl
		gbot := (getter.pos[1] + getter.sizeBox[3]) * getter.localscl

		if cbot >= gtop && ctop <= gbot { // Pushbox vertical overlap

			// We skip the zAxisCheck function because we'll need to calculate the overlap again anyway

			// Normal collision check
			cxleft := c.sizeBox[0] * c.localscl
			cxright := c.sizeBox[2] * c.localscl
			if c.facing < 0 {
				cxleft, cxright = -cxright, -cxleft
			}

			cxleft += c.pos[0] * c.localscl
			cxright += c.pos[0] * c.localscl

			gxleft := getter.sizeBox[0] * getter.localscl
			gxright := getter.sizeBox[2] * getter.localscl
			if getter.facing < 0 {
				gxleft, gxright = -gxright, -gxleft
			}

			gxleft += getter.pos[0] * getter.localscl
			gxright += getter.pos[0] * getter.localscl

			// X axis fail
			if gxleft >= cxright || cxleft >= gxright {
				continue
			}

			czback := c.pos[2]*c.localscl - c.size.depth*c.localscl
			czfront := c.pos[2]*c.localscl + c.size.depth*c.localscl

			gzback := getter.pos[2]*getter.localscl - getter.size.depth*getter.localscl
			gzfront := getter.pos[2]*getter.localscl + getter.size.depth*getter.localscl

			// Z axis fail
			if gzback >= czfront || czback >= gzfront {
				continue
			}

			// Push characters away from each other
			if c.asf(ASF_sizepushonly) || getter.clsnCheck(c, 2, 2, false, false) {

				gxmin = getter.edge[0]
				gxmax = -getter.edge[1]
				if getter.facing > 0 {
					gxmin, gxmax = -gxmax, -gxmin
				}
				gxmin += sys.xmin / getter.localscl
				gxmax += sys.xmax / getter.localscl

				getter.pushed, c.pushed = true, true

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
				cfactor := float32(getter.size.weight) / float32(c.size.weight+getter.size.weight) * c.size.pushfactor * cpushed
				gfactor := float32(c.size.weight) / float32(c.size.weight+getter.size.weight) * getter.size.pushfactor * gpushed

				// Determine in which axes to push the players
				// This needs to check both if the players have velocity or if their positions changed
				pushx := sys.zmin == sys.zmax ||
					getter.vel[0] != 0 || c.vel[0] != 0 || getter.pos[0] != getter.oldPos[0] || c.pos[0] != c.oldPos[0]
				pushz := sys.zmin != sys.zmax &&
					(getter.vel[2] != 0 || c.vel[2] != 0 || getter.pos[2] != getter.oldPos[2] || c.pos[2] != c.oldPos[2])

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
				}

				// TODO: Z axis push might need some decision for who stays in the corner, like X axis
				if pushz {
					if getter.pos[2] >= c.pos[2] {
						if c.pushPriority >= getter.pushPriority {
							getter.pos[2] -= ((czfront - gzback) * gfactor) / getter.localscl
						}
						if c.pushPriority <= getter.pushPriority {
							c.pos[2] += ((czfront - gzback) * cfactor) / c.localscl
						}
					} else {
						if c.pushPriority >= getter.pushPriority {
							getter.pos[2] -= ((gzfront - czback) * gfactor) / getter.localscl
						}
						if c.pushPriority <= getter.pushPriority {
							c.pos[2] += ((gzfront - czback) * cfactor) / c.localscl
						}
					}
					// Clamp Z positions
					c.zDepthBound()
					getter.zDepthBound()

				}

				if getter.trackableByCamera() && getter.csf(CSF_screenbound) {
					getter.pos[0] = ClampF(getter.pos[0], gxmin, gxmax)
				}
				if c.trackableByCamera() && c.csf(CSF_screenbound) {
					l, r := c.edge[0], -c.edge[1]
					if c.facing > 0 {
						l, r = -r, -l
					}
					c.pos[0] = ClampF(c.pos[0], l+sys.xmin/c.localscl, r+sys.xmax/c.localscl)
				}
				getter.pos[0] = ClampF(getter.pos[0], sys.stage.leftbound*(sys.stage.localscl/getter.localscl), sys.stage.rightbound*(sys.stage.localscl/getter.localscl))
				c.pos[0] = ClampF(c.pos[0], sys.stage.leftbound*(sys.stage.localscl/c.localscl), sys.stage.rightbound*(sys.stage.localscl/c.localscl))
				getter.interPos[0], c.interPos[0] = getter.pos[0], c.pos[0]
			}
		}
	}
}

func (cl *CharList) collisionDetection() {

	sortedOrder := []int{}
	// Check ReversalDefs first
	for i, c := range cl.runOrder {
		if c.hitdef.reversal_attr > 0 {
			sortedOrder = append(sortedOrder, i)
		}
	}
	// Check Hitdefs second
	for i, c := range cl.runOrder {
		if c.hitdef.attr > 0 && c.hitdef.reversal_attr == 0 {
			sortedOrder = append(sortedOrder, i)
		}
	}
	soNum := []int{}
	soCount := 0
	for i := 0; i < len(cl.runOrder); i++ {
		if soCount < len(sortedOrder) {
			if sortedOrder[soCount] == i {
				soCount++
				continue
			}
		}
		soNum = append(soNum, i)
	}
	sortedOrder = append(sortedOrder, soNum...)

	// Push detection for players
	// This must happen before hit detection
	// https://github.com/ikemen-engine/Ikemen-GO/issues/1941
	// An attempt was made to skip redundant player pair checks, but that makes chars push each other too slowly in screen corners
	for i := 0; i < len(cl.runOrder); i++ {
		cl.pushDetection(cl.runOrder[sortedOrder[i]])
	}
	// Hit detection for players
	for i := 0; i < len(cl.runOrder); i++ {
		cl.hitDetection(cl.runOrder[sortedOrder[i]], false)
	}
	// Hit detection for projectiles
	for _, c := range cl.runOrder {
		cl.hitDetection(c, true)
	}
}

func (cl *CharList) tick() {
	sys.gameTime++
	for _, c := range cl.runOrder {
		c.tick()
	}
}
func (cl *CharList) cueDraw() {
	for _, c := range cl.drawOrder {
		if c != nil {
			c.cueDraw()
		}
	}
}
func (cl *CharList) get(id int32) *Char {
	if id < 0 {
		return nil
	}
	return cl.idMap[id]
}
func (cl *CharList) getIndex(id int32) *Char {
	for j, p := range cl.runOrder {
		if (id - 1) == int32(j) {
			return p
		}
	}
	return nil
}
func (cl *CharList) getHelperIndex(c *Char, id int32, ex bool) *Char {
	var t []int32
	parent := func(c *Char) *Char {
		if c.parentIndex == IErr {
			return nil
		}
		return sys.chars[c.playerNo][Abs(c.parentIndex)]
	}
	for j, h := range cl.runOrder {
		if c.id != h.id {
			if c.helperIndex == 0 {
				hr := sys.chars[h.playerNo][0]
				if h.helperIndex != 0 && hr != nil && c.id == hr.id {
					t = append(t, int32(j))
				}
			} else {
				hp := parent(h)
				for hp != nil {
					if hp.id == c.id {
						t = append(t, int32(j))
					}
					hp = parent(hp)
				}
			}
		}
	}
	for i := 0; i < len(t); i++ {
		ch := cl.runOrder[int32(t[i])]
		if (id-1) == int32(i) && ch != nil {
			return ch
		}
	}
	if !ex {
		sys.appendToConsole(c.warn() + fmt.Sprintf("has no helper with index: %v", id))
	}
	return nil
}
func (cl *CharList) p2enemyDelete(c *Char) {
	for _, e := range cl.runOrder {
		for i, p2cl := range e.p2enemy {
			if p2cl == c {
				e.p2enemy = e.p2enemy[:i]
				break
			}
		}
	}
}
func (cl *CharList) enemyNear(c *Char, n int32, p2, ignoreDefeatedEnemy, log bool) *Char {
	if n < 0 {
		if log {
			sys.appendToConsole(c.warn() + fmt.Sprintf("has no nearest enemy: %v", n))
		}
		return nil
	}
	cache := &c.enemynear[Btoi(p2)]
	if int(n) < len(*cache) {
		return (*cache)[n]
	}
	if p2 {
		cache = &c.p2enemy
	} else {
		*cache = (*cache)[:0]
	}
	var add func(*Char, int, float32)
	add = func(e *Char, idx int, adddist float32) {
		for i := idx; i <= int(n); i++ {
			if i >= len(*cache) {
				*cache = append(*cache, e)
				return
			}
			if AbsF(c.distX(e, c))+adddist < AbsF(c.distX((*cache)[i], c)) {
				add((*cache)[i], i+1, adddist)
				(*cache)[i] = e
				return
			}
		}
	}
	for _, e := range cl.runOrder {
		if e.player && e.teamside != c.teamside && !e.scf(SCF_standby) {
			if p2 && !e.scf(SCF_ko_round_middle) {
				add(e, 0, 30)
			}
			if !p2 && e.helperIndex == 0 && (!ignoreDefeatedEnemy || ignoreDefeatedEnemy && (!e.scf(SCF_ko_round_middle) || sys.roundEnd())) {
				add(e, 0, 0)
			}
		}
	}
	if int(n) >= len(*cache) {
		if log {
			sys.appendToConsole(c.warn() + fmt.Sprintf("has no nearest enemy: %v", n))
		}
		return nil
	}
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
