package main

import (
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"strings"
)

const MaxPalNo = 12

type SystemCharFlag uint32

const (
	SCF_ko SystemCharFlag = 1 << iota
	SCF_ctrl
	SCF_standby
	SCF_guard
	SCF_airjump
	SCF_over
	SCF_ko_round_middle
)

type CharSpecialFlag uint32

const (
	CSF_nostandguard CharSpecialFlag = 1 << iota
	CSF_nocrouchguard
	CSF_noairguard
	CSF_noshadow
	CSF_invisible
	CSF_unguardable
	CSF_nojugglecheck
	CSF_noautoturn
	CSF_nowalk
	CSF_screenbound
	CSF_movecamera_x
	CSF_movecamera_y
	CSF_posfreeze
	CSF_playerpush
	CSF_angledraw
	CSF_destroy
	CSF_frontedge
	CSF_backedge
	CSF_frontwidth
	CSF_backwidth
	CSF_trans
	CSF_gethit
	CSF_assertspecial CharSpecialFlag = CSF_nostandguard | CSF_nocrouchguard |
		CSF_noairguard | CSF_noshadow | CSF_invisible | CSF_unguardable |
		CSF_nojugglecheck | CSF_noautoturn | CSF_nowalk
)

type GlobalSpecialFlag uint32

const (
	GSF_intro GlobalSpecialFlag = 1 << iota
	GSF_roundnotover
	GSF_nomusic
	GSF_nobardisplay
	GSF_nobg
	GSF_nofg
	GSF_globalnoshadow
	GSF_timerfreeze
	GSF_nokosnd
	GSF_nokoslow
	GSF_noko
)

type PosType int32

const (
	PT_P1 PosType = iota
	PT_P2
	PT_F
	PT_B
	PT_L
	PT_R
	PT_N
)

type ClsnRect [][4]float32

func (cr *ClsnRect) Add(clsn []float32, x, y, xs, ys float32) {
	for i := 0; i+3 < len(clsn); i += 4 {
		*cr = append(*cr, [...]float32{x + xs*clsn[i] + float32(sys.gameWidth)/2,
			y + ys*clsn[i+1] + float32(sys.gameHeight-240),
			xs * (clsn[i+2] - clsn[i]), ys * (clsn[i+3] - clsn[i+1])})
	}
}

type CharData struct {
	life    int32
	power   int32
	attack  int32
	defence int32
	fall    struct {
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
	ko struct {
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
	cd.attack = 100
	cd.defence = 100
	cd.fall.defence_mul = 1.5
	cd.liedown.time = 60
	cd.airjuggle = 15
	cd.sparkno = 2
	cd.guard.sparkno = 40
	cd.ko.echo = 0
	cd.volume = 256
	cd.intpersistindex = 0
	cd.floatpersistindex = 0
}

type CharSize struct {
	xscale float32
	yscale float32
	ground struct {
		back  int32
		front int32
	}
	air struct {
		back  int32
		front int32
	}
	height int32
	attack struct {
		dist int32
		z    struct {
			width [2]int32
		}
	}
	proj struct {
		attack struct {
			dist int32
		}
		doscale int32
		xscale  float32
		yscale  float32
	}
	head struct {
		pos [2]int32
	}
	mid struct {
		pos [2]int32
	}
	shadowoffset int32
	draw         struct {
		offset [2]int32
	}
	z struct {
		width int32
	}
}

func (cs *CharSize) init() {
	*cs = CharSize{}
	cs.xscale = 1
	cs.yscale = 1
	cs.ground.back = 15
	cs.ground.front = 16
	cs.air.back = 12
	cs.air.front = 12
	cs.height = 60
	cs.attack.dist = 160
	cs.proj.attack.dist = 90
	cs.proj.doscale = 0
	cs.proj.xscale = 1
	cs.proj.yscale = 1
	cs.head.pos = [...]int32{-5, -90}
	cs.mid.pos = [...]int32{-5, -60}
	cs.shadowoffset = 0
	cs.draw.offset = [...]int32{0, 0}
	cs.z.width = 3
	cs.attack.z.width = [...]int32{4, 4}
}

type CharVelocity struct {
	walk struct {
		fwd  float32
		back float32
		up   struct {
			x float32
		}
		down struct {
			x float32
		}
	}
	run struct {
		fwd  [2]float32
		back [2]float32
		up   struct {
			x float32
			y float32
		}
		down struct {
			x float32
			y float32
		}
	}
	jump struct {
		neu  [2]float32
		back float32
		fwd  float32
		up   struct {
			x float32
		}
		down struct {
			x float32
		}
	}
	runjump struct {
		back [2]float32
		fwd  [2]float32
		up   struct {
			x float32
		}
		down struct {
			x float32
		}
	}
	airjump struct {
		neu  [2]float32
		back float32
		fwd  float32
		up   struct {
			x float32
		}
		down struct {
			x float32
		}
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
}

type CharMovement struct {
	airjump struct {
		num    int32
		height int32
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
		friction_threshold float32
	}
}

func (cm *CharMovement) init() {
	*cm = CharMovement{}
	cm.airjump.num = 0
	cm.airjump.height = 35
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
	cm.down.bounce.offset = [...]float32{0.0, 20.0}
	cm.down.bounce.yaccel = 0.4
	cm.down.bounce.groundlevel = 12.0
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
	recover        bool
	recovertime    int32
	damage         int32
	kill           bool
	envshake_time  int32
	envshake_freq  float32
	envshake_ampl  int32
	envshake_phase float32
}

func (f *Fall) clear() {
	*f = Fall{animtype: RA_Unknown, xvelocity: float32(math.NaN()),
		yvelocity: -4.5}
}
func (f *Fall) setDefault() {
	*f = Fall{animtype: RA_Unknown, xvelocity: float32(math.NaN()),
		yvelocity: -4.5, recover: true, recovertime: 4, kill: true,
		envshake_freq: 60, envshake_ampl: -4, envshake_phase: float32(math.NaN())}
}
func (f *Fall) xvel() float32 {
	if math.IsNaN(float64(f.xvelocity)) {
		return -32760
	}
	return f.xvelocity
}

type HitDef struct {
	attr                       int32
	reversal_attr              int32
	hitflag                    int32
	guardflag                  int32
	affectteam                 int32
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
	guard_sparkno              int32
	sparkxy                    [2]float32
	hitsound                   [2]int32
	guardsound                 [2]int32
	ground_type                HitType
	air_type                   HitType
	ground_slidetime           int32
	guard_slidetime            int32
	ground_hittime             int32
	guard_hittime              int32
	air_hittime                int32
	guard_ctrltime             int32
	airguard_ctrltime          int32
	guard_dist                 int32
	yaccel                     float32
	ground_velocity            [2]float32
	guard_velocity             float32
	air_velocity               [2]float32
	airguard_velocity          [2]float32
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
	ground_fall                bool
	air_fall                   bool
	down_velocity              [2]float32
	down_hittime               int32
	down_bounce                bool
	id                         int32
	chainid                    int32
	nochainid                  [2]int32
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
	mindist                    [2]float32
	maxdist                    [2]float32
	snap                       [2]float32
	snapt                      int32
	fall                       Fall
	playerNo                   int
	kill                       bool
	guard_kill                 bool
	forcenofall                bool
	lhit                       bool
}

func (hd *HitDef) clear() {
	*hd = HitDef{hitflag: int32(ST_S | ST_C | ST_A | ST_F), affectteam: 1,
		animtype: RA_Light, air_animtype: RA_Unknown, priority: 4,
		bothhittype: AT_Hit, sparkno: IErr, guard_sparkno: IErr,
		hitsound: [...]int32{IErr, -1}, guardsound: [...]int32{IErr, -1},
		ground_type: HT_High, air_type: HT_Unknown, air_hittime: 20,
		yaccel: float32(math.NaN()), guard_velocity: float32(math.NaN()),
		airguard_velocity: [...]float32{float32(math.NaN()),
			float32(math.NaN())},
		ground_cornerpush_veloff:   float32(math.NaN()),
		air_cornerpush_veloff:      float32(math.NaN()),
		down_cornerpush_veloff:     float32(math.NaN()),
		guard_cornerpush_veloff:    float32(math.NaN()),
		airguard_cornerpush_veloff: float32(math.NaN()), p1sprpriority: 1,
		p1stateno: -1, p2stateno: -1, forcestand: IErr,
		down_velocity: [...]float32{float32(math.NaN()), float32(math.NaN())},
		chainid:       -1, nochainid: [...]int32{-1, -1}, numhits: 1,
		hitgetpower: IErr, guardgetpower: IErr, hitgivepower: IErr,
		guardgivepower: IErr, envshake_freq: 60, envshake_ampl: -4,
		envshake_phase: float32(math.NaN()),
		mindist:        [...]float32{float32(math.NaN()), float32(math.NaN())},
		maxdist:        [...]float32{float32(math.NaN()), float32(math.NaN())},
		snap:           [...]float32{float32(math.NaN()), float32(math.NaN())},
		kill:           true, guard_kill: true, playerNo: -1}
	hd.palfx.mul, hd.palfx.color = [...]int32{255, 255, 255}, 1
	hd.fall.setDefault()
}
func (hd *HitDef) invalidate(stateType StateType) {
	hd.attr = hd.attr&^int32(ST_MASK) | int32(stateType) | -1<<31
	hd.reversal_attr |= -1 << 31
	hd.lhit = false
}
func (hd *HitDef) testAttr(attr int32) bool {
	attr &= hd.attr
	return attr&int32(ST_MASK) != 0 && attr&^int32(ST_MASK)&^(-1<<31) != 0
}

type GetHitVar struct {
	hitBy          [][2]int32
	hit1           [2]int32
	hit2           [2]int32
	attr           int32
	_type          HitType
	airanimtype    Reaction
	groundanimtype Reaction
	airtype        HitType
	groundtype     HitType
	damage         int32
	hitcount       int32
	fallcount      int32
	hitshaketime   int32
	hittime        int32
	slidetime      int32
	ctrltime       int32
	xvel           float32
	yvel           float32
	yaccel         float32
	hitid          int32
	xoff           float32
	yoff           float32
	fall           Fall
	playerNo       int
	fallf          bool
	guarded        bool
	p2getp1state   bool
	forcestand     bool
}

func (ghv *GetHitVar) clear() {
	*ghv = GetHitVar{_type: -1, hittime: -1, yaccel: float32(math.NaN()),
		xoff: ghv.xoff, yoff: ghv.yoff, hitid: -1, playerNo: -1}
	ghv.fall.clear()
}
func (ghv *GetHitVar) clearOff() {
	ghv.xoff, ghv.yoff = 0, 0
}
func (ghv GetHitVar) getYaccel() float32 {
	if math.IsNaN(float64(ghv.yaccel)) {
		return 0.35
	}
	return ghv.yaccel
}
func (ghv GetHitVar) chainId() int32 {
	if ghv.hitid > 0 {
		return ghv.hitid
	}
	return 0
}
func (ghv GetHitVar) idMatch(id int32) bool {
	for _, v := range ghv.hitBy {
		if v[0] == id {
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
	ghv.dropId(id)
	ghv.hitBy = append(ghv.hitBy, [...]int32{id, juggle})
}

type HitBy struct {
	falg, time int32
}
type HitOverride struct {
	attr     int32
	stateno  int32
	time     int32
	forceair bool
	playerNo int
}

func (ho *HitOverride) clear() {
	*ho = HitOverride{stateno: -1, playerNo: -1}
}

type aimgImage struct {
	anim           Animation
	pos, scl, ascl [2]float32
	angle          float32
	aset, oldVer   bool
}

type AfterImage struct {
	time       int32
	length     int32
	postbright [3]int32
	add        [3]int32
	mul        [3]float32
	timegap    int32
	framegap   int32
	alpha      [2]int32
	palfx      []PalFX
	imgs       [64]aimgImage
	imgidx     int32
	restgap    int32
	reccount   int32
}

func newAfterImage() *AfterImage {
	ai := &AfterImage{palfx: make([]PalFX, sys.afterImageMax)}
	for i := range ai.palfx {
		ai.palfx[i].enable, ai.palfx[i].negType = true, true
	}
	ai.clear()
	ai.timegap = 0
	return ai
}
func (ai *AfterImage) clear() {
	ai.time = 0
	ai.length = 20
	if len(ai.palfx) > 0 {
		ai.palfx[0].eColor = 1
		ai.palfx[0].eInvertall = false
		ai.palfx[0].eAdd = [...]int32{30, 30, 30}
		ai.palfx[0].eMul = [...]int32{120, 120, 220}
	}
	ai.postbright = [3]int32{}
	ai.add = [...]int32{10, 10, 25}
	ai.mul = [...]float32{0.65, 0.65, 0.75}
	ai.timegap = 1
	ai.framegap = 6
	ai.alpha = [...]int32{-1, 0}
	ai.imgidx = 0
	ai.restgap = 0
	ai.reccount = 0
}
func (ai *AfterImage) setPalColor(color int32) {
	if len(ai.palfx) > 0 {
		ai.palfx[0].eColor = float32(Max(0, Min(256, color))) / 256
	}
}
func (ai *AfterImage) setPalInvertall(invertall bool) {
	if len(ai.palfx) > 0 {
		ai.palfx[0].eInvertall = invertall
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
	for i := 1; i < len(ai.palfx); i++ {
		ai.palfx[i].eColor = ai.palfx[i-1].eColor
		ai.palfx[i].eInvertall = ai.palfx[i-1].eInvertall
		ai.palfx[i].eAdd[0] = ai.palfx[i-1].eAdd[0] + pb[0]
		ai.palfx[i].eAdd[1] = ai.palfx[i-1].eAdd[1] + pb[1]
		ai.palfx[i].eAdd[2] = ai.palfx[i-1].eAdd[2] + pb[2]
		pb = [3]int32{}
		ai.palfx[i].eMul[0] = int32(float32(ai.palfx[i-1].eMul[0]) * ai.mul[0])
		ai.palfx[i].eMul[1] = int32(float32(ai.palfx[i-1].eMul[1]) * ai.mul[1])
		ai.palfx[i].eMul[2] = int32(float32(ai.palfx[i-1].eMul[2]) * ai.mul[2])
	}
}
func (ai *AfterImage) recAfterImg(sd *SprData) {
	if ai.time == 0 {
		ai.reccount, ai.timegap = 0, 0
		return
	}
	if ai.time > 0 {
		ai.time--
	}
	if ai.restgap <= 0 {
		img := &ai.imgs[ai.imgidx]
		img.anim = *sd.anim
		if sd.anim.spr != nil {
			img.anim.spr = newSprite()
			*img.anim.spr = *sd.anim.spr
			sd.anim.sff.palList.SwapPalMap(&sd.fx.remap)
			img.anim.spr.Pal = sd.anim.spr.GetPal(&sd.anim.sff.palList)
			sd.anim.sff.palList.SwapPalMap(&sd.fx.remap)
		}
		img.pos = sd.pos
		img.scl = sd.scl
		img.angle = sd.angle
		img.ascl = sd.ascl
		img.aset = sd.aset
		img.oldVer = sd.oldVer
		ai.imgidx = (ai.imgidx + 1) & 63
		if int(ai.reccount) < len(ai.imgs) {
			ai.reccount++
		}
		ai.restgap = ai.timegap
	}
	ai.restgap--
}
func (ai *AfterImage) recAndCue(sd *SprData, rec bool) {
	if ai.time == 0 || ai.timegap < 1 || ai.timegap > 32767 ||
		ai.framegap < 1 || ai.framegap > 32767 {
		ai.time = 0
		ai.reccount, ai.timegap = 0, 0
		return
	}
	end := Min(sys.afterImageMax,
		(Min(ai.reccount, ai.length)/ai.framegap)*ai.framegap)
	for i := ai.framegap; i <= end; i += ai.framegap {
		img := &ai.imgs[(ai.imgidx-i)&63]
		sys.sprites.add(&SprData{&img.anim, &ai.palfx[i/ai.framegap-1], img.pos,
			img.scl, ai.alpha, sd.priority - 2, img.angle, img.ascl, img.aset,
			false, sd.bright, sd.oldVer}, 0, 0, 0)
	}
	if rec {
		ai.recAfterImg(sd)
	}
}

type Explod struct {
	id             int32
	bindtime       int32
	scale          [2]float32
	time           int32
	removeongethit bool
	removetime     int32
	velocity       [2]float32
	accel          [2]float32
	sprpriority    int32
	postype        PosType
	offset         [2]float32
	relativef      int32
	pos            [2]float32
	facing         float32
	vfacing        int32
	shadow         [3]int32
	supermovetime  int32
	pausemovetime  int32
	anim           *Animation
	ontop          bool
	alpha          [2]int32
	ownpal         bool
	playerId       int32
	bindId         int32
	ignorehitpause bool
	angle          float32
	oldPos         [2]float32
	newPos         [2]float32
	palfx          *PalFX
}

func (e *Explod) clear() {
	*e = Explod{id: IErr, scale: [...]float32{1, 1}, removetime: -2,
		postype: PT_P1, relativef: 1, facing: 1, vfacing: 1,
		alpha: [...]int32{-1, 0}, playerId: -1, bindId: -1, ignorehitpause: true}
}
func (e *Explod) setPos(c *Char) {
	unimplemented()
}
func (e *Explod) matchId(eid, pid int32) bool {
	return e.id >= 0 && e.playerId == pid && (eid < 0 || e.id == eid)
}
func (e *Explod) update(oldVer bool, playerNo int) {
	if e.anim == nil {
		e.id = IErr
	}
	if e.id == IErr {
		e.anim = nil
		return
	}
	var c *Char
	if !e.ignorehitpause || e.removeongethit {
		c = sys.charList.get(e.playerId)
	}
	p := false
	if sys.super > 0 {
		p = e.supermovetime >= 0 && e.time >= e.supermovetime
	} else if sys.pause > 0 {
		p = e.pausemovetime >= 0 && e.time >= e.pausemovetime
	}
	act := !p
	if act && !e.ignorehitpause {
		act = c == nil || c.acttmp%2 >= 0
	}
	if sys.tickFrame() {
		if c != nil && e.removeongethit && c.ss.moveType == MT_H ||
			e.removetime >= 0 && e.time >= e.removetime ||
			act && e.removetime < -1 && e.anim.loopend {
			e.id, e.anim = IErr, nil
			return
		}
	}
	unimplemented()
}

type Projectile struct {
	hitdef        HitDef
	id            int32
	anim          int32
	hitanim       int32
	remanim       int32
	cancelanim    int32
	scale         [2]float32
	clsnscale     [2]float32
	remove        bool
	removetime    int32
	velocity      [2]float32
	remvelocity   [2]float32
	accel         [2]float32
	velmul        [2]float32
	hits          int32
	misstime      int32
	priority      int32
	prioritypoint int32
	sprpriority   int32
	edgebound     int32
	stagebound    int32
	heightbound   [2]int32
	pos           [2]float32
	facing        float32
	shadow        [3]int32
	supermovetime int32
	pausemovetime int32
	ani           *Animation
	timemiss      int32
	hitpause      int32
	oldPos        [2]float32
	newPos        [2]float32
	aimg          AfterImage
	palfx         *PalFX
}

func (p *Projectile) clear() {
	*p = Projectile{id: IErr, hitanim: -1, remanim: IErr, cancelanim: IErr,
		scale: [...]float32{1, 1}, clsnscale: [...]float32{1, 1}, remove: true,
		removetime: -1, velmul: [...]float32{1, 1}, hits: 1, priority: 1,
		prioritypoint: 1, sprpriority: 3, edgebound: 40, stagebound: 40,
		heightbound: [...]int32{-240, 1}, facing: 1}
	p.hitdef.clear()
}
func (p *Projectile) update(playerNo int) {
	unimplemented()
}
func (p *Projectile) clsn(playerNo int) {
	unimplemented()
}
func (p *Projectile) tick(playerNo int) {
	unimplemented()
}
func (p *Projectile) anime(oldVer bool, playerNo int) {
	unimplemented()
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
	displayname      string
	author           string
	palkeymap        [MaxPalNo]int32
	sff              *Sff
	snd              *Snd
	anim             AnimationTable
	palno, drawpalno int32
	pal              [MaxPalNo]string
	palExist         [MaxPalNo]bool
	palSelectable    [MaxPalNo]bool
	ver              [2]uint16
	data             CharData
	velocity         CharVelocity
	movement         CharMovement
	states           map[int32]StateBytecode
	wakewakaLength   int32
	pctype           ProjContact
	pctime, pcid     int32
}

func (cgi *CharGlobalInfo) clearPCTime() {
	cgi.pctype = PC_Hit
	cgi.pctime = -1
	cgi.pcid = 0
}

type StateState struct {
	stateType       StateType
	moveType        MoveType
	physics         StateType
	ps              []int32
	wakegawakaranai [][]bool
	no, prevno      int32
	time            int32
	sb              StateBytecode
}

func (ss *StateState) clear() {
	ss.stateType, ss.moveType, ss.physics = ST_S, MT_I, ST_N
	ss.ps = nil
	ss.wakegawakaranai = make([][]bool, len(sys.cgi))
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
	airJumpCount  int32
	hitCount      int32
	uniqHitCount  int32
	pauseMovetime int32
	superMovetime int32
	bindTime      int32
	bindToId      int32
	bindPos       [2]float32
	bindFacing    float32
	hitPauseTime  int32
	angle         float32
	angleScalse   [2]float32
	alpha         [2]int32
	recoverTime   int32
	systemFlag    SystemCharFlag
	sprPriority   int32
	getcombo      int32
	veloff        float32
	width, edge   [2]float32
	attackMul     float32
	defenceMul    float32
}
type Char struct {
	name            string
	palfx           *PalFX
	anim            *Animation
	curFrame        *AnimFrame
	cmd             []CommandList
	ss              StateState
	key             int
	id              int32
	helperId        int32
	helperIndex     int32
	parentChar      *Char
	playerNo        int
	keyctrl         bool
	player          bool
	animPN          int
	animNo          int32
	life            int32
	lifeMax         int32
	power           int32
	powerMax        int32
	juggle          int32
	fallTime        int32
	size            CharSize
	clsnScale       [2]float32
	hitdef          HitDef
	ghv             GetHitVar
	hitby           [2]HitBy
	ho              [8]HitOverride
	hoIdx           int
	mctype          MoveContact
	mctime          int32
	children        []*Char
	targets         []int32
	targetsOfHitdef []int32
	enemyNear       [2][]*Char
	specialFlag     CharSpecialFlag
	pos             [2]float32
	drawPos         [2]float32
	oldPos          [2]float32
	vel             [2]float32
	facing          float32
	ivar            [NumVar + NumSysVar]int32
	fvar            [NumFvar + NumSysFvar]float32
	CharSystemVar
	aimg          AfterImage
	sounds        Sounds
	p1facing      float32
	cpucmd        int32
	attackDist    float32
	offset        [2]float32
	angleset      bool
	stchtmp       bool
	inguarddist   bool
	pushed        bool
	hitdefContact bool
	atktmp        int8
	hittmp        int8
	acttmp        int8
	minus         int8
}

func newChar(n int, idx int32) (c *Char) {
	c = &Char{}
	c.init(n, idx)
	return c
}
func (c *Char) init(n int, idx int32) {
	c.clear1()
	c.playerNo, c.helperIndex = n, idx
	if c.helperIndex == 0 {
		c.keyctrl, c.player = true, true
	}
	c.key = n
	if n >= 0 && n < len(sys.com) && sys.com[n] != 0 {
		c.key ^= -1
	}
}
func (c *Char) clearState() {
	c.ss.clear()
	c.hitdef.clear()
	c.ghv.clear()
	c.ghv.clearOff()
	c.hitby = [2]HitBy{}
	for i := range c.ho {
		c.ho[i].clear()
	}
	c.mctype = MC_Hit
	c.mctime = 0
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
	c.fallTime = 0
	c.varRangeSet(0, int32(NumVar)-1, 0)
	c.fvarRangeSet(0, int32(NumFvar)-1, 0)
	c.key = -1
	c.id = IErr
	c.helperId = 0
	c.helperIndex = -1
	c.parentChar = nil
	c.playerNo = -1
	c.facing = 1
	c.keyctrl = false
	c.player = false
	c.animPN = -1
	c.animNo = 0
	c.angleset = false
	c.stchtmp = false
	c.inguarddist = false
	c.p1facing = 0
	c.pushed = false
	c.atktmp, c.hittmp, c.acttmp, c.minus = 0, 0, 0, 2
}
func (c *Char) copyParent(p *Char) {
	c.parentChar = p
	c.name, c.key, c.size = p.name+"'s helper", p.key, p.size
	c.life, c.lifeMax, c.power, c.powerMax = p.lifeMax, p.lifeMax, 0, p.powerMax
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
	c.enemyNear[0] = c.enemyNear[0][:0]
	c.enemyNear[1] = c.enemyNear[1][:0]
}
func (c *Char) clear2() {
	c.sysVarRangeSet(0, int32(NumSysVar)-1, 0)
	c.sysFvarRangeSet(0, int32(NumSysFvar)-1, 0)
	c.CharSystemVar = CharSystemVar{bindToId: -1,
		angleScalse: [...]float32{1, 1}, alpha: [...]int32{255, 0},
		width:      [...]float32{c.defFW(), c.defBW()},
		attackMul:  float32(c.gi().data.attack) / 100,
		defenceMul: float32(c.gi().data.defence) / 100}
	c.oldPos, c.drawPos = c.pos, c.pos
	if c.helperIndex == 0 {
		if sys.roundsExisted[c.playerNo&1] > 0 {
			c.palfx.clear()
		} else {
			c.palfx = newPalFX()
		}
	} else {
		c.palfx = nil
	}
	c.aimg.timegap = -1
	c.enemyNearClear()
	c.targets = c.targets[:0]
	c.cpucmd = -1
}
func (c *Char) gi() *CharGlobalInfo {
	return &sys.cgi[c.playerNo]
}
func (c *Char) stCgi() *CharGlobalInfo {
	return &sys.cgi[c.ss.sb.playerNo]
}
func (c *Char) load(def string) error {
	gi := &sys.cgi[c.playerNo]
	gi.def, gi.displayname, gi.author, gi.sff, gi.snd = def, "", "", nil, nil
	gi.anim = NewAnimationTable()
	for i := range gi.palkeymap {
		gi.palkeymap[i] = int32(i)
	}
	str, err := LoadText(def)
	if err != nil {
		return err
	}
	lines, i := SplitAndTrim(str, "\n"), 0
	cns, sprite, anim, sound := "", "", "", ""
	info, files, keymap := true, true, true
	for i < len(lines) {
		is, name, subname := ReadIniSection(lines, &i)
		switch name {
		case "info":
			if info {
				info = false
				c.name, _, _ = is.getText("name")
				var ok bool
				gi.displayname, ok, _ = is.getText("displayname")
				if !ok {
					gi.displayname = c.name
				}
				gi.author, _, _ = is.getText("author")
			}
		case "files":
			if files {
				files = false
				cns, sprite = is["cns"], is["sprite"]
				anim, sound = is["anim"], is["sound"]
				for i := range gi.pal {
					gi.pal[i] = is[fmt.Sprintf("pal%d", i+1)]
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
		}
	}
	if err := LoadFile(&cns, def, func(filename string) error {
		str, err := LoadText(filename)
		if err != nil {
			return err
		}
		lines, i = SplitAndTrim(str, "\n"), 0
		return nil
	}); err != nil {
		return err
	}
	gi.data.init()
	c.size.init()
	gi.velocity.init()
	data, size, velocity, movement := true, true, true, true
	for i < len(lines) {
		is, name, _ := ReadIniSection(lines, &i)
		switch name {
		case "data":
			if data {
				data = false
				is.ReadI32("life", &gi.data.life)
				c.lifeMax = gi.data.life
				is.ReadI32("power", &gi.data.power)
				c.powerMax = gi.data.power
				is.ReadI32("attack", &gi.data.attack)
				is.ReadI32("defence", &gi.data.defence)
				var i32 int32
				if is.ReadI32("fall.defence_up", &i32) {
					gi.data.fall.defence_mul = (float32(i32) + 100) / 100
				}
				if is.ReadI32("liedown.time", &i32) {
					gi.data.liedown.time = Max(1, i32)
				}
				is.ReadI32("airjuggle", &gi.data.airjuggle)
				is.ReadI32("sparkno", &gi.data.sparkno)
				is.ReadI32("guard.sparkno", &gi.data.guard.sparkno)
				is.ReadI32("ko.echo", &gi.data.ko.echo)
				if gi.ver[0] == 1 {
					if is.ReadI32("volumescale", &i32) {
						gi.data.volume = i32 * 64 / 25
					}
				} else if is.ReadI32("volume", &i32) {
					gi.data.volume = i32 + 256
				}
				is.ReadI32("intpersistindex", &gi.data.intpersistindex)
				is.ReadI32("floatpersistindex", &gi.data.floatpersistindex)
			}
		case "size":
			if size {
				size = false
				is.ReadF32("xscale", &c.size.xscale)
				is.ReadF32("yscale", &c.size.yscale)
				is.ReadI32("ground.back", &c.size.ground.back)
				is.ReadI32("ground.front", &c.size.ground.front)
				is.ReadI32("air.back", &c.size.air.back)
				is.ReadI32("air.front", &c.size.air.front)
				is.ReadI32("height", &c.size.height)
				is.ReadI32("attack.dist", &c.size.attack.dist)
				is.ReadI32("proj.attack.dist", &c.size.proj.attack.dist)
				is.ReadI32("proj.doscale", &c.size.proj.doscale)
				if c.size.proj.doscale != 0 {
					c.size.proj.xscale, c.size.proj.yscale = c.size.xscale, c.size.yscale
				}
				is.ReadI32("head.pos", &c.size.head.pos[0], &c.size.head.pos[1])
				is.ReadI32("mid.pos", &c.size.mid.pos[0], &c.size.mid.pos[1])
				is.ReadI32("shadowoffset", &c.size.shadowoffset)
				is.ReadI32("draw.offset",
					&c.size.draw.offset[0], &c.size.draw.offset[1])
				is.ReadI32("z.width", &c.size.z.width)
				is.ReadI32("attack.z.width",
					&c.size.attack.z.width[0], &c.size.attack.z.width[1])
			}
		case "velocity":
			if velocity {
				velocity = false
				is.ReadF32("walk.fwd", &gi.velocity.walk.fwd)
				is.ReadF32("walk.back", &gi.velocity.walk.back)
				is.ReadF32("walk.up.x", &gi.velocity.walk.up.x)
				is.ReadF32("walk.down.x", &gi.velocity.walk.down.x)
				is.ReadF32("run.fwd", &gi.velocity.run.fwd[0], &gi.velocity.run.fwd[1])
				is.ReadF32("run.back",
					&gi.velocity.run.back[0], &gi.velocity.run.back[1])
				is.ReadF32("run.up.x", &gi.velocity.run.up.x)
				is.ReadF32("run.up.y", &gi.velocity.run.up.y)
				is.ReadF32("run.down.x", &gi.velocity.run.down.x)
				is.ReadF32("run.down.y", &gi.velocity.run.down.y)
				is.ReadF32("jump.neu",
					&gi.velocity.jump.neu[0], &gi.velocity.jump.neu[1])
				is.ReadF32("jump.back", &gi.velocity.jump.back)
				is.ReadF32("jump.fwd", &gi.velocity.jump.fwd)
				is.ReadF32("jump.up.x", &gi.velocity.jump.up.x)
				is.ReadF32("jump.down.x", &gi.velocity.jump.down.x)
				is.ReadF32("runjump.back",
					&gi.velocity.runjump.back[0], &gi.velocity.runjump.back[1])
				is.ReadF32("runjump.fwd",
					&gi.velocity.runjump.fwd[0], &gi.velocity.runjump.fwd[1])
				is.ReadF32("runjump.up.x", &gi.velocity.runjump.up.x)
				is.ReadF32("runjump.down.x", &gi.velocity.runjump.down.x)
				is.ReadF32("airjump.neu",
					&gi.velocity.airjump.neu[0], &gi.velocity.airjump.neu[1])
				is.ReadF32("airjump.back", &gi.velocity.airjump.back)
				is.ReadF32("airjump.fwd", &gi.velocity.airjump.fwd)
				is.ReadF32("airjump.up.x", &gi.velocity.airjump.up.x)
				is.ReadF32("airjump.down.x", &gi.velocity.airjump.down.x)
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
			}
		case "movement":
			if movement {
				movement = false
				is.ReadI32("airjump.num", &gi.movement.airjump.num)
				is.ReadI32("airjump.height", &gi.movement.airjump.height)
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
			}
		}
	}
	if LoadFile(&sprite, def, func(filename string) error {
		var err error
		gi.sff, err = LoadSff(filename, false)
		return err
	}); err != nil {
		return err
	}
	if LoadFile(&anim, def, func(filename string) error {
		str, err := LoadText(filename)
		if err != nil {
			return err
		}
		lines, i := SplitAndTrim(str, "\n"), 0
		gi.anim = ReadAnimationTable(gi.sff, lines, &i)
		return nil
	}); err != nil {
		return err
	}
	if len(sound) > 0 {
		if LoadFile(&sound, def, func(filename string) error {
			var err error
			gi.snd, err = LoadSnd(filename)
			return err
		}); err != nil {
			return err
		}
	} else {
		gi.snd = newSnd()
	}
	return nil
}
func (c *Char) loadPallet() {
	if c.gi().sff.header.Ver0 == 1 {
		c.gi().sff.palList.ResetRemap()
		tmp := 0
		for i := 0; i < MaxPalNo; i++ {
			pl := c.gi().sff.palList.Get(i)
			var f *os.File
			if LoadFile(&c.gi().pal[i], c.gi().def, func(file string) (err error) {
				f, err = os.Open(file)
				return
			}) == nil {
				for i := 255; i >= 0; i-- {
					var rgb [3]byte
					if binary.Read(f, binary.LittleEndian, rgb[:]) != nil {
						break
					}
					pl[i] = uint32(rgb[0])<<16 | uint32(rgb[1])<<8 | uint32(rgb[2])
				}
				if tmp == 0 && i > 0 {
					copy(c.gi().sff.palList.Get(0), pl)
				}
				tmp = i + 1
				c.gi().palExist[i] = true
			} else {
				c.gi().palExist[i] = false
				if i > 0 {
					delete(c.gi().sff.palList.PalTable, [...]int16{1, int16(i + 1)})
				}
			}
		}
		if tmp == 0 {
			if c.gi().ver[0] == 1 {
				delete(c.gi().sff.palList.PalTable, [...]int16{1, 1})
			} else {
				spr := c.gi().sff.GetSprite(9000, 0)
				if spr == nil {
					spr = c.gi().sff.GetSprite(0, 0)
				}
				if spr != nil {
					copy(c.gi().sff.palList.Get(0), c.gi().sff.palList.Get(spr.palidx))
				}
			}
		}
	} else {
		for i := 0; i < MaxPalNo; i++ {
			_, c.gi().palExist[i] =
				c.gi().sff.palList.PalTable[[...]int16{1, int16(i + 1)}]
		}
	}
	for i := range c.gi().palSelectable {
		c.gi().palSelectable[i] = false
	}
	for i := 0; i < MaxPalNo; i++ {
		startj := c.gi().palkeymap[i]
		if !c.gi().palExist[startj] {
			startj %= 6
		}
		j := startj
		for {
			if c.gi().palExist[j] {
				c.gi().palSelectable[j] = true
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
	c.gi().drawpalno = c.gi().palno
	starti := c.gi().palno - 1
	if !c.gi().palExist[starti] {
		starti %= 6
	}
	i := starti
	for {
		if c.gi().palExist[i] {
			j := 0
			for ; j < len(sys.chars); j++ {
				if j != c.playerNo && len(sys.chars[j]) > 0 &&
					sys.cgi[j].def == c.gi().def && sys.cgi[j].drawpalno == i+1 {
					break
				}
			}
			if j >= len(sys.chars) {
				c.gi().drawpalno = i + 1
				if !c.gi().palExist[c.gi().palno-1] {
					c.gi().palno = c.gi().drawpalno
				}
				break
			}
		}
		i++
		if i >= MaxPalNo {
			i = 0
		}
		if i == starti {
			if !c.gi().palExist[c.gi().palno-1] {
				i := 0
				for ; i < len(c.gi().palExist); i++ {
					if c.gi().palExist[i] {
						c.gi().palno, c.gi().drawpalno = int32(i+1), int32(i+1)
						break
					}
				}
				if i >= len(c.gi().palExist) {
					c.gi().palno, c.gi().palExist[0] = 1, true
					c.gi().palSelectable[0] = true
				}
			}
			break
		}
	}
}
func (c *Char) clearHitCount() {
	c.hitCount, c.uniqHitCount = 0, 0
}
func (c *Char) clearMoveHit() {
	c.mctime = 0
	if c.helperIndex == 0 {
		for i, pr := range sys.projs[c.playerNo] {
			if pr.id < 0 {
				sys.projs[c.playerNo][i].id = IErr
			}
		}
	}
}
func (c *Char) clearHitDef() {
	c.hitdef.clear()
}
func (c *Char) setSprPriority(sprpriority int32) {
	c.sprPriority = sprpriority
}
func (c *Char) setJuggle(juggle int32) {
	c.juggle = juggle
}
func (c *Char) setXV(xv float32) {
	c.vel[0] = xv
}
func (c *Char) setYV(yv float32) {
	c.vel[1] = yv
}
func (c *Char) changeAnim(animNo int32) {
	if a := c.getAnim(animNo, false); a != nil {
		c.anim = a
		c.animPN = c.playerNo
		c.animNo = animNo
		c.clsnScale = [...]float32{sys.chars[c.animPN][0].size.xscale,
			sys.chars[c.animPN][0].size.yscale}
		if c.hitPause() {
			c.curFrame = a.CurrentFrame()
		}
	}
}
func (c *Char) changeAnim2(animNo int32) {
	if a := sys.chars[c.ss.sb.playerNo][0].getAnim(animNo, false); a != nil {
		c.anim = a
		c.animPN = c.ss.sb.playerNo
		c.animNo = animNo
		c.clsnScale = [...]float32{sys.chars[c.animPN][0].size.xscale,
			sys.chars[c.animPN][0].size.yscale}
		a.sff = sys.cgi[c.animPN].sff
		if c.hitPause() {
			c.curFrame = a.CurrentFrame()
		}
	}
}
func (c *Char) setAnimElem(e int32) {
	unimplemented()
}
func (c *Char) setCtrl(ctrl bool) {
	if ctrl {
		c.setSCF(SCF_ctrl)
	} else {
		c.unsetSCF(SCF_ctrl)
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
func (c *Char) sf(csf CharSpecialFlag) bool {
	return c.specialFlag&csf != 0
}
func (c *Char) setSF(csf CharSpecialFlag) {
	c.specialFlag |= csf
}
func (c *Char) unsetSF(csf CharSpecialFlag) {
	c.specialFlag &^= csf
}
func (c *Char) time() int32 {
	return c.ss.time
}
func (c *Char) alive() bool {
	return !c.scf(SCF_ko)
}
func (c *Char) playSound(f, lw, lp bool, g, n, ch, vo int32,
	p, fr float32, x *float32) {
	unimplemented()
}
func (c *Char) furimuki() {
	if c.scf(SCF_ctrl) && c.helperIndex == 0 {
		unimplemented()
	}
}
func (c *Char) stateChange1(no int32, pn int) bool {
	if sys.changeStateNest >= 2500 {
		fmt.Printf("2500 loops: %v, %v -> %v -> %v\n",
			c.name, c.ss.prevno, c.ss.no, no)
		return false
	}
	c.ss.no, c.ss.prevno, c.ss.time = Max(0, no), c.ss.no, 0
	if c.ss.sb.playerNo != c.playerNo && pn != c.ss.sb.playerNo {
		c.enemyExplodsRemove(c.ss.sb.playerNo)
	}
	var ok bool
	if c.ss.sb, ok = sys.cgi[pn].states[no]; !ok {
		fmt.Printf("存在しないステート: P%v:%v\n", pn+1, no)
		c.ss.sb = *newStateBytecode(pn)
		c.ss.sb.stateType, c.ss.sb.moveType, c.ss.sb.physics = ST_U, MT_U, ST_U
	}
	c.stchtmp = true
	return true
}
func (c *Char) stateChange2() {
	if c.stchtmp && !c.hitPause() {
		c.ss.sb.init(c)
		c.stchtmp = false
	}
}
func (c *Char) changeStateEx(no int32, pn int, anim, ctrl int32) {
	if c.minus <= 0 && (c.ss.stateType == ST_S || c.ss.stateType == ST_C) {
		c.furimuki()
	}
	if anim >= 0 {
		c.changeAnim(anim)
	}
	if ctrl >= 0 {
		c.setCtrl(ctrl != 0)
	}
	if c.stateChange1(no, pn) && sys.changeStateNest == 0 && c.minus == 0 &&
		c.id >= 0 {
		for c.stchtmp && sys.changeStateNest < 2500 {
			c.stateChange2()
			sys.changeStateNest++
			c.ss.sb.run(c)
		}
		sys.changeStateNest = 0
	}
}
func (c *Char) changeState(no, anim, ctrl int32) {
	c.changeStateEx(no, c.ss.sb.playerNo, anim, ctrl)
}
func (c *Char) selfState(no, anim, ctrl int32) {
	c.changeStateEx(no, c.playerNo, anim, ctrl)
}
func (c *Char) partner(n int32) *Char {
	n = Max(0, n)
	if int(n) > len(sys.chars)/2-2 {
		return nil
	}
	var p int
	if int(n) == c.playerNo>>1 {
		p = c.playerNo + 2
	} else {
		p = c.playerNo&1 + int(n)<<1
		if int(n) > c.playerNo>>1 {
			p += 2
		}
	}
	if len(sys.chars[p]) > 0 {
		return sys.chars[p][0]
	}
	return nil
}
func (c *Char) destroy() {
	if c.helperIndex > 0 {
		return
	}
	unimplemented()
}
func (c *Char) destroySelf(recursive, removeexplods bool) bool {
	if c.helperIndex <= 0 {
		return false
	}
	c.setSF(CSF_destroy)
	if removeexplods {
		c.removeExplod(-1)
	}
	if recursive {
		for _, ch := range c.children {
			if ch != nil {
				ch.destroySelf(recursive, removeexplods)
			}
		}
	}
	unimplemented()
	return true
}
func (c *Char) newHelper() (h *Char) {
	i := int32(0)
	for ; int(i) < len(sys.chars[c.playerNo]); i++ {
		if sys.chars[c.playerNo][i].helperIndex < 0 {
			h = sys.chars[c.playerNo][i]
			h.init(c.playerNo, i)
			break
		}
	}
	if int(i) >= len(sys.chars[c.playerNo]) {
		if i >= sys.helperMax {
			return
		}
		h = newChar(c.playerNo, i)
		sys.chars[c.playerNo] = append(sys.chars[c.playerNo], h)
	}
	h.id, h.helperId = ^sys.newCharId(), 0
	h.copyParent(c)
	c.addChild(h)
	sys.charList.add(h)
	return
}
func (c *Char) helperPos(pt PosType, pos [2]float32, facing int32,
	dstFacing *float32) (p [2]float32) {
	if facing < 0 {
		*dstFacing *= -1
	}
	switch pt {
	case PT_P1:
		p[0] = c.pos[0] + pos[0]*c.facing
		p[1] = c.pos[1] + pos[1]
		*dstFacing *= c.facing
	case PT_P2:
		if p2 := c.p2(); p2 != nil {
			p[0] = p2.pos[0] + pos[0]*p2.facing
			p[1] = p2.pos[1] + pos[1]
			*dstFacing *= p2.facing
		}
	case PT_F, PT_B:
		p[0] = sys.cam.ScreenPos[0]
		if c.facing > 0 && pt == PT_F || c.facing < 0 && pt == PT_B {
			p[0] += float32(sys.gameWidth) / sys.cam.Scale
		}
		if c.facing > 0 {
			p[0] += pos[0]
		} else {
			p[0] -= pos[0]
		}
		p[1] = pos[1]
		*dstFacing *= c.facing
	case PT_L:
		p[0] = sys.cam.ScreenPos[0] + pos[0]
		p[1] = pos[1]
	case PT_R:
		p[0] = sys.cam.ScreenPos[0] + float32(sys.gameWidth)/sys.cam.Scale + pos[0]
		p[1] = pos[1]
	case PT_N:
		p = pos
	}
	return
}
func (c *Char) helperInit(h *Char, st int32, pt PosType, x, y float32,
	facing int32, ownpal bool) {
	p := c.helperPos(pt, [...]float32{x, y}, facing, &h.facing)
	h.setX(p[0])
	h.setY(p[1])
	h.vel = [2]float32{}
	if ownpal {
		h.palfx = newPalFX()
		tmp := c.getPalfx().remap
		h.palfx.remap = make([]int, len(tmp))
		copy(h.palfx.remap, tmp)
	}
	h.changeStateEx(st, c.playerNo, 0, 1)
}
func (c *Char) roundState() int32 {
	switch {
	case sys.intro > sys.lifebar.ro.ctrl_time+1:
		return 0
	case sys.lifebar.ro.cur == 0:
		return 1
	case !sys.roundEnd():
		return 2
	case sys.intro < -(sys.lifebar.ro.over_hittime+
		sys.lifebar.ro.over_waittime) && (sys.chars[c.playerNo][0].scf(SCF_over) ||
		sys.chars[c.playerNo][0].scf(SCF_ko)):
		return 4
	default:
		return 3
	}
}
func (c *Char) animTime() int32 {
	unimplemented()
	return 0
}
func (c *Char) animElemTime(e int32) int32 {
	unimplemented()
	return 0
}
func (c *Char) newExplod() (*Explod, int) {
	unimplemented()
	return nil, 0
}
func (c *Char) getExplods(id int32) []*Explod {
	unimplemented()
	return nil
}
func (c *Char) remapPalSub(pfx *PalFX, sg, sn, dg, dn int32) {
	unimplemented()
}
func (c *Char) insertExplodEx(i int, rp [2]int32) {
	unimplemented()
}
func (c *Char) insertExplod(i int) {
	c.insertExplodEx(i, [...]int32{-1, 0})
}
func (c *Char) removeExplod(id int32) {
	remove := func(drawlist *[]int, drop bool) {
		for i := len(*drawlist) - 1; i >= 0; i-- {
			ei := (*drawlist)[i]
			if ei >= 0 && sys.explods[c.playerNo][ei].matchId(id, c.id) {
				sys.explods[c.playerNo][ei].id = IErr
				if drop {
					*drawlist = append((*drawlist)[:i], (*drawlist)[i+1:]...)
				} else {
					(*drawlist)[i] = -1
				}
			}
		}
	}
	remove(&sys.explDrawlist[c.playerNo], true)
	remove(&sys.topexplDrawlist[c.playerNo], false)
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
	remove(&sys.explDrawlist[en], true)
	remove(&sys.topexplDrawlist[en], false)
}
func (c *Char) getAnim(n int32, ffx bool) (a *Animation) {
	if ffx {
		a = sys.lifebar.fat.get(n)
	} else {
		a = c.gi().anim.get(n)
	}
	if a == nil {
		fmt.Printf("存在しないアニメ: P%v:%v\n", (c.playerNo+1)*int(1-2*Btoi(ffx)), n)
	}
	return
}
func (c *Char) setPosX(x float32) {
	if c.pos[0] != x {
		c.pos[0] = x
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
func (c *Char) setPosY(y float32) {
	c.pos[1] = y
}
func (c *Char) posReset() {
	c.facing = 1 - 2*float32(c.playerNo&1)
	c.setX(float32(sys.stage.p[c.playerNo&1].startx-sys.cam.startx)*
		sys.stage.localscl - c.facing*float32(c.playerNo>>1)*P1P3Dist)
	c.setY(0)
	c.setXV(0)
	c.setYV(0)
}
func (c *Char) setX(x float32) {
	c.oldPos[0], c.drawPos[0] = x, x
	c.setPosX(x)
}
func (c *Char) setY(y float32) {
	c.oldPos[1], c.drawPos[1] = y, y
	c.setPosY(y)
}
func (c *Char) addX(x float32) {
	c.setX(c.pos[0] + x)
}
func (c *Char) addY(y float32) {
	c.setY(c.pos[1] + y)
}
func (c *Char) addXV(xv float32) {
	c.vel[0] += xv
}
func (c *Char) addYV(yv float32) {
	c.vel[1] += yv
}
func (c *Char) mulXV(xv float32) {
	c.vel[0] *= xv
}
func (c *Char) mulYV(yv float32) {
	c.vel[1] *= yv
}
func (c *Char) parent() *Char {
	return c.parentChar
}
func (c *Char) root() *Char {
	if c.helperIndex == 0 {
		return nil
	}
	return sys.chars[c.playerNo][0]
}
func (c *Char) helper(id int32) *Char {
	unimplemented()
	return nil
}
func (c *Char) target(id int32) *Char {
	unimplemented()
	return nil
}
func (c *Char) enemy(n int32) *Char {
	unimplemented()
	return nil
}
func (c *Char) enemynear(n int32) *Char {
	return sys.charList.enemyNear(c, n, false)
}
func (c *Char) playerid(id int32) *Char {
	unimplemented()
	return nil
}
func (c *Char) p2() *Char {
	p2 := sys.charList.enemyNear(c, 0, true)
	if p2 != nil && p2.scf(SCF_ko) && p2.scf(SCF_over) {
		return nil
	}
	return p2
}
func (c *Char) newProj() *Projectile {
	unimplemented()
	return nil
}
func (c *Char) projInit(p *Projectile, pt PosType, x, y float32,
	op bool, rpg, rpn int32) {
	unimplemented()
}
func (c *Char) setHitdefDefault(hd *HitDef, proj bool) {
	if !proj {
		c.targetsOfHitdef = c.targetsOfHitdef[:0]
	}
	if hd.attr&^int32(ST_MASK) == 0 {
		hd.attr = 0
	}
	if hd.hitonce < 0 || hd.attr&int32(AT_AT) != 0 {
		hd.hitonce = 1
	}
	ifnanset := func(dst *float32, src float32) {
		if math.IsNaN(float64(*dst)) {
			*dst = src
		}
	}
	ifierrset := func(dst *int32, src int32) {
		if *dst == IErr {
			*dst = src
		}
	}
	ifnanset(&hd.ground_velocity[0], 0)
	ifnanset(&hd.ground_velocity[1], 0)
	ifnanset(&hd.air_velocity[0], 0)
	ifnanset(&hd.air_velocity[1], 0)
	ifnanset(&hd.guard_velocity, hd.ground_velocity[0])
	ifnanset(&hd.airguard_velocity[0], hd.air_velocity[0]*1.5)
	ifnanset(&hd.airguard_velocity[1], hd.air_velocity[1]*0.5)
	ifnanset(&hd.down_velocity[0], hd.air_velocity[0])
	ifnanset(&hd.down_velocity[1], hd.air_velocity[1])
	if hd.fall.animtype == RA_Unknown {
		if hd.air_animtype != RA_Unknown {
			hd.fall.animtype = hd.air_animtype
		} else if hd.animtype < RA_Back {
			hd.fall.animtype = RA_Back
		} else {
			hd.fall.animtype = hd.animtype
		}
	}
	if hd.air_animtype == RA_Unknown {
		hd.air_animtype = hd.animtype
	}
	if hd.air_type == HT_Unknown {
		if hd.ground_type == HT_Trip {
			hd.air_type = HT_High
		} else {
			hd.air_type = hd.ground_type
		}
	}
	ifierrset(&hd.forcestand, Btoi(hd.ground_velocity[1] != 0))
	if hd.attr&int32(ST_A) != 0 {
		ifnanset(&hd.ground_cornerpush_veloff, 0)
	} else {
		ifnanset(&hd.ground_cornerpush_veloff, hd.guard_velocity*1.3)
	}
	ifnanset(&hd.air_cornerpush_veloff, hd.ground_cornerpush_veloff)
	ifnanset(&hd.down_cornerpush_veloff, hd.ground_cornerpush_veloff)
	ifnanset(&hd.guard_cornerpush_veloff, hd.ground_cornerpush_veloff)
	ifnanset(&hd.airguard_cornerpush_veloff, hd.ground_cornerpush_veloff)
	ifierrset(&hd.hitgetpower,
		int32(sys.attack_LifeToPowerMul*float32(hd.hitdamage)))
	ifierrset(&hd.guardgetpower,
		int32(sys.attack_LifeToPowerMul*float32(hd.hitdamage)*0.5))
	ifierrset(&hd.hitgivepower,
		int32(sys.getHit_LifeToPowerMul*float32(hd.hitdamage)))
	ifierrset(&hd.guardgivepower,
		int32(sys.getHit_LifeToPowerMul*float32(hd.hitdamage)*0.5))
	if !math.IsNaN(float64(hd.snap[0])) {
		hd.maxdist[0], hd.mindist[0] = hd.snap[0], hd.snap[0]
	}
	if !math.IsNaN(float64(hd.snap[1])) {
		hd.maxdist[1], hd.mindist[1] = hd.snap[1], hd.snap[1]
	}
}
func (c *Char) setFEdge(fe float32) {
	c.edge[0] = fe
	c.setSF(CSF_frontedge)
}
func (c *Char) setBEdge(be float32) {
	c.edge[1] = be
	c.setSF(CSF_backedge)
}
func (c *Char) setFWidth(fw float32) {
	c.width[0] = fw
	c.setSF(CSF_frontwidth)
}
func (c *Char) setBWidth(bw float32) {
	c.width[1] = bw
	c.setSF(CSF_backwidth)
}
func (c *Char) moveContact() int32 {
	if c.mctype != MC_Reversed {
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
func (c *Char) moveGuarded() int32 {
	if c.mctype == MC_Guarded {
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
func (c *Char) gethitAnimtype() Reaction {
	if c.ghv.fallf {
		return c.ghv.fall.animtype
	} else if c.ss.stateType == ST_A {
		return c.ghv.airanimtype
	}
	return c.ghv.groundanimtype
}
func (c *Char) isBound() bool {
	return c.ghv.idMatch(c.bindToId)
}
func (c *Char) canRecover() bool {
	return c.ghv.fall.recover && c.fallTime >= c.ghv.fall.recovertime
}
func (c *Char) command(pn, i int) bool {
	unimplemented()
	return false
}
func (c *Char) varGet(i int32) BytecodeValue {
	if i >= 0 && i < int32(NumVar) {
		return BytecodeInt(c.ivar[i])
	}
	return BytecodeSF()
}
func (c *Char) fvarGet(i int32) BytecodeValue {
	if i >= 0 && i < int32(NumFvar) {
		return BytecodeFloat(c.fvar[i])
	}
	return BytecodeSF()
}
func (c *Char) sysVarGet(i int32) BytecodeValue {
	if i >= 0 && i < int32(NumSysVar) {
		return BytecodeInt(c.ivar[i+int32(NumVar)])
	}
	return BytecodeSF()
}
func (c *Char) sysFvarGet(i int32) BytecodeValue {
	if i >= 0 && i < int32(NumSysFvar) {
		return BytecodeFloat(c.fvar[i+int32(NumFvar)])
	}
	return BytecodeSF()
}
func (c *Char) varSet(i, v int32) BytecodeValue {
	if i >= 0 && i < int32(NumVar) {
		c.ivar[i] = v
		return BytecodeInt(v)
	}
	return BytecodeSF()
}
func (c *Char) fvarSet(i int32, v float32) BytecodeValue {
	if i >= 0 && i < int32(NumFvar) {
		c.fvar[i] = v
		return BytecodeFloat(v)
	}
	return BytecodeSF()
}
func (c *Char) sysVarSet(i, v int32) BytecodeValue {
	if i >= 0 && i < int32(NumSysVar) {
		c.ivar[i+int32(NumVar)] = v
		return BytecodeInt(v)
	}
	return BytecodeSF()
}
func (c *Char) sysFvarSet(i int32, v float32) BytecodeValue {
	if i >= 0 && i < int32(NumSysFvar) {
		c.fvar[i+int32(NumFvar)] = v
		return BytecodeFloat(v)
	}
	return BytecodeSF()
}
func (c *Char) varAdd(i, v int32) BytecodeValue {
	if i >= 0 && i < int32(NumVar) {
		c.ivar[i] += v
		return BytecodeInt(c.ivar[i])
	}
	return BytecodeSF()
}
func (c *Char) fvarAdd(i int32, v float32) BytecodeValue {
	if i >= 0 && i < int32(NumFvar) {
		c.fvar[i] += v
		return BytecodeFloat(c.fvar[i])
	}
	return BytecodeSF()
}
func (c *Char) sysVarAdd(i, v int32) BytecodeValue {
	if i >= 0 && i < int32(NumSysVar) {
		c.ivar[i+int32(NumVar)] += v
		return BytecodeInt(c.ivar[i+int32(NumVar)])
	}
	return BytecodeSF()
}
func (c *Char) sysFvarAdd(i int32, v float32) BytecodeValue {
	if i >= 0 && i < int32(NumSysFvar) {
		c.fvar[i+int32(NumFvar)] += v
		return BytecodeFloat(c.fvar[i+int32(NumFvar)])
	}
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
			c.vel[1] *= -1
			c.ghv.xvel *= -1
		}
	}
}
func (c *Char) getTarget(id int32) []int32 {
	unimplemented()
	return nil
}
func (c *Char) targetFacing(tar []int32, f int32) {
	unimplemented()
}
func (c *Char) targetBind(tar []int32, t int32, x, y float32) {
	unimplemented()
}
func (c *Char) bindToTarget(tar []int32, t int32, x, y float32, hnf HMF) {
	unimplemented()
}
func (c *Char) targetLifeAdd(tar []int32, add int32, kill, absolute bool) {
	unimplemented()
}
func (c *Char) targetState(tar []int32, state int32) {
	unimplemented()
}
func (c *Char) targetVelSetX(tar []int32, x float32) {
	unimplemented()
}
func (c *Char) targetVelSetY(tar []int32, y float32) {
	unimplemented()
}
func (c *Char) targetVelAddX(tar []int32, x float32) {
	unimplemented()
}
func (c *Char) targetVelAddY(tar []int32, y float32) {
	unimplemented()
}
func (c *Char) targetPowerAdd(tar []int32, power int32) {
	unimplemented()
}
func (c *Char) targetDrop(excludeid int32, keepone bool) {
	unimplemented()
}
func (c *Char) lifeAdd(add float64, kill, absolute bool) {
	if add != 0 && c.roundState() != 3 {
		if !absolute {
			add /= float64(c.defenceMul)
		}
		add = math.Floor(add)
		max := float64(c.gi().data.life - c.life)
		if add > max {
			add = max
		}
		min := float64(-c.life)
		if !kill {
			min += 1
		}
		if add < min {
			add = min
		}
		c.lifeSet(c.life + int32(add))
	}
}
func (c *Char) lifeSet(life int32) {
	if c.life = Max(0, Min(c.gi().data.life, life)); c.life == 0 {
		if c.player {
			if c.alive() {
				unimplemented()
			}
		} else {
			c.life = 1
		}
	}
}
func (c *Char) setPower(pow int32) {
	if !sys.roundEnd() {
		c.power = Max(0, Min(c.powerMax, pow))
	}
}
func (c *Char) powerAdd(add int32) {
	if sys.powerShare[c.playerNo&1] {
		sys.chars[c.playerNo&1][0].setPower(c.getPower() + add)
	} else {
		sys.chars[c.playerNo][0].setPower(c.getPower() + add)
	}
}
func (c *Char) powerSet(pow int32) {
	if sys.powerShare[c.playerNo&1] {
		sys.chars[c.playerNo&1][0].setPower(pow)
	} else {
		sys.chars[c.playerNo][0].setPower(pow)
	}
}
func (c *Char) distX(opp *Char) float32 {
	return opp.pos[0] - c.pos[0]
}
func (c *Char) bodyDistX(opp *Char) float32 {
	dist := c.distX(opp)
	var oppw float32
	if dist == 0 || (dist < 0) != (opp.facing < 0) {
		oppw = opp.facing * opp.width[0]
	} else {
		oppw = -opp.facing * opp.width[1]
	}
	return dist + oppw - c.facing*c.width[0]
}
func (c *Char) rdDistX(rd *Char) BytecodeValue {
	if rd == nil {
		return BytecodeSF()
	}
	dist := c.facing * c.distX(rd)
	if c.stCgi().ver[0] != 1 {
		dist = float32(int32(dist))
	}
	return BytecodeFloat(dist)
}
func (c *Char) rdDistY(rd *Char) BytecodeValue {
	if rd == nil {
		return BytecodeSF()
	}
	return BytecodeFloat(rd.pos[1] - c.pos[1])
}
func (c *Char) p2BodyDistX() BytecodeValue {
	if p2 := c.p2(); p2 == nil {
		return BytecodeSF()
	} else {
		dist := c.facing * c.bodyDistX(p2)
		if c.stCgi().ver[0] != 1 {
			dist = float32(int32(dist))
		}
		return BytecodeFloat(dist)
	}
}
func (c *Char) hitShakeOver() bool {
	return c.ghv.hitshaketime <= 0
}
func (c *Char) hitVelSetX() {
	if c.ss.moveType == MT_H {
		c.setXV(c.ghv.xvel)
	}
}
func (c *Char) hitVelSetY() {
	if c.ss.moveType == MT_H {
		c.setYV(c.ghv.yvel)
	}
}
func (c *Char) getEdge(base float32, actually bool) float32 {
	if !actually || c.gi().ver[0] != 1 {
		switch c.ss.stateType {
		case ST_A:
			return base + 1
		case ST_L:
			return base + 2
		}
	}
	return base
}
func (c *Char) defFW() float32 {
	if c.ss.stateType == ST_A {
		return float32(c.size.air.front)
	}
	return float32(c.size.ground.front)
}
func (c *Char) defBW() float32 {
	if c.ss.stateType == ST_A {
		return float32(c.size.air.back)
	}
	return float32(c.size.ground.back)
}
func (c *Char) frontEdgeDist() float32 {
	if c.facing > 0 {
		return sys.xmax - c.pos[0]
	}
	return c.pos[0] - sys.xmin
}
func (c *Char) frontEdgeBodyDist() float32 {
	return c.frontEdgeDist() - c.getEdge(c.edge[0], false)
}
func (c *Char) frontEdge() float32 {
	if c.facing > 0 {
		return c.rightEdge()
	}
	return c.leftEdge()
}
func (c *Char) backEdgeDist() float32 {
	if c.facing < 0 {
		return sys.xmax - c.pos[0]
	}
	return c.pos[0] - sys.xmin
}
func (c *Char) backEdgeBodyDist() float32 {
	return c.backEdgeDist() - c.getEdge(c.edge[1], false)
}
func (c *Char) backEdge() float32 {
	if c.facing < 0 {
		return c.rightEdge()
	}
	return c.leftEdge()
}
func (c *Char) leftEdge() float32 {
	unimplemented()
	return 0
}
func (c *Char) rightEdge() float32 {
	unimplemented()
	return 0
}
func (c *Char) topEdge() float32 {
	unimplemented()
	return 0
}
func (c *Char) bottomEdge() float32 {
	unimplemented()
	return 0
}
func (c *Char) screenPosX() float32 {
	unimplemented()
	return 0
}
func (c *Char) screenPosY() float32 {
	unimplemented()
	return 0
}
func (c *Char) height() float32 {
	return float32(c.size.height)
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
func (c *Char) selfAnimExist(anim BytecodeValue) BytecodeValue {
	if anim.IsSF() {
		return BytecodeSF()
	}
	unimplemented()
	return BytecodeBool(false)
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
func (c *Char) setSuperPauseTime(pausetime, movetime int32) {
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
}
func (c *Char) getPalfx() *PalFX {
	if c.palfx != nil {
		return c.palfx
	}
	if c.parentChar == nil {
		c.palfx = newPalFX()
		return c.palfx
	}
	return c.parentChar.getPalfx()
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
func (c *Char) getPower() int32 {
	if sys.powerShare[c.playerNo&1] {
		return sys.chars[c.playerNo&1][0].power
	}
	return sys.chars[c.playerNo][0].power
}
func (c *Char) isHelper(hid BytecodeValue) BytecodeValue {
	if hid.IsSF() {
		return BytecodeSF()
	}
	id := hid.ToI()
	return BytecodeBool(c.helperIndex != 0 && (id <= 0 || c.helperId == id))
}
func (c *Char) numHelper(hid BytecodeValue) BytecodeValue {
	if hid.IsSF() {
		return BytecodeSF()
	}
	id := hid.ToI()
	n := int32(0)
	for _, h := range sys.chars[c.playerNo][1:] {
		if !h.sf(CSF_destroy) && (id <= 0 || h.helperId == id) {
			n++
		}
	}
	return BytecodeInt(n)
}
func (c *Char) angleSet(a float32) {
	c.angle = a
	if a != 0 {
		c.angleset = true
	}
}
func (c *Char) roundsExisted() int32 {
	return sys.roundsExisted[c.playerNo&1]
}
func (c *Char) ctrlOver() bool {
	return sys.time == 0 ||
		sys.intro < -(sys.lifebar.ro.over_hittime+sys.lifebar.ro.over_waittime)
}
func (c *Char) canCtrl() bool {
	return c.scf(SCF_ctrl) && !c.scf(SCF_ko) && !c.ctrlOver()
}
func (c *Char) win() bool {
	return sys.winTeam == c.playerNo&1
}
func (c *Char) lose() bool {
	return sys.winTeam == ^c.playerNo&1
}
func (c *Char) hitDefAttr(attr int32) bool {
	return c.ss.moveType == MT_A && c.hitdef.testAttr(attr)
}
func (c *Char) makeDust(x, y float32) {
	if e, i := c.newExplod(); e != nil {
		e.anim = c.getAnim(120, true)
		e.sprpriority = math.MaxInt32
		e.ownpal = true
		e.offset = [...]float32{x, y}
		e.setPos(c)
		c.insertExplod(i)
	}
}
func (c *Char) hitOver() bool {
	return c.ghv.hittime < 0
}
func (c *Char) hitFallDamage() {
	if c.ss.moveType == MT_H {
		c.lifeAdd(-float64(c.ghv.fall.damage), c.ghv.fall.kill, false)
	}
}
func (c *Char) hitFallVel() {
	if c.ss.moveType == MT_H {
		if !math.IsNaN(float64(c.ghv.fall.xvelocity)) {
			c.setXV(c.ghv.fall.xvelocity)
		}
		c.setYV(c.ghv.fall.yvelocity)
	}
}
func (c *Char) hitFallSet(f int32, xv, yv float32) {
	if c.ss.moveType == MT_H {
		if f >= 0 {
			c.ghv.fallf = f != 0
		}
		if !math.IsNaN(float64(xv)) {
			c.ghv.fall.xvelocity = xv
		}
		if !math.IsNaN(float64(yv)) {
			c.ghv.fall.yvelocity = yv
		}
	}
}
func (c *Char) remapPal(pfx *PalFX, src [2]int32, dst [2]int32) {
	if src[0] < 0 || src[1] < 0 || dst[0] < 0 || dst[1] < 0 {
		return
	}
	si, ok := c.gi().sff.palList.PalTable[[...]int16{int16(src[0]),
		int16(src[1])}]
	if !ok {
		return
	}
	var di int
	di, ok = c.gi().sff.palList.PalTable[[...]int16{int16(dst[0]),
		int16(dst[1])}]
	if !ok {
		di = si
	}
	if pfx.remap == nil {
		pfx.remap = c.gi().sff.palList.GetPalMap()
	}
	if c.gi().sff.palList.SwapPalMap(&pfx.remap) {
		c.gi().sff.palList.Remap(si, di)
		if src[0] == 1 && src[1] == 1 && c.gi().sff.header.Ver0 == 1 {
			spr := c.gi().sff.GetSprite(0, 0)
			if spr != nil {
				c.gi().sff.palList.Remap(spr.palidx, di)
			}
			spr = c.gi().sff.GetSprite(9000, 0)
			if spr != nil {
				c.gi().sff.palList.Remap(spr.palidx, di)
			}
		}
		c.gi().sff.palList.SwapPalMap(&pfx.remap)
	}
}
func (c *Char) inGuardState() bool {
	return c.ss.no == 120 || (c.ss.no >= 130 && c.ss.no <= 132) ||
		c.ss.no == 140 || (c.ss.no >= 150 && c.ss.no <= 155)
}
func (c *Char) gravity() {
	c.vel[1] += c.gi().movement.yaccel
}
func (c *Char) posUpdate() {
	nobind := [...]bool{c.bindTime == 0 || math.IsNaN(float64(c.bindPos[0])),
		c.bindTime == 0 || math.IsNaN(float64(c.bindPos[1]))}
	for i := range nobind {
		if nobind[i] {
			c.oldPos[i], c.drawPos[i] = c.pos[i], c.pos[i]
		}
	}
	if c.sf(CSF_posfreeze) {
		if nobind[0] {
			c.setPosX(c.oldPos[0] + c.veloff)
		}
	} else {
		if nobind[0] {
			c.setPosX(c.oldPos[0] + c.vel[0]*c.facing + c.veloff)
		}
		if nobind[1] {
			c.setPosY(c.oldPos[1] + c.vel[1])
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
	c.veloff *= 0.7
	if AbsF(c.veloff) < 1 {
		c.veloff = 0
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
func (c *Char) setBindTime(time int32) {
	c.bindTime = time
	if time == 0 {
		c.bindToId = -1
	}
}
func (c *Char) bind() {
	if c.bindTime == 0 {
		return
	}
	if bt := sys.charList.get(c.bindToId); bt != nil {
		if bt.hasTarget(c.id) {
			if bt.sf(CSF_destroy) {
				c.selfState(5050, -1, -1)
				c.setBindTime(0)
				return
			}
			if !math.IsNaN(float64(c.bindPos[0])) {
				c.setXV(c.facing * bt.facing * bt.vel[0])
			}
			if !math.IsNaN(float64(c.bindPos[1])) {
				c.setYV(bt.vel[1])
			}
		}
		if !math.IsNaN(float64(c.bindPos[0])) {
			f := bt.facing
			if AbsF(c.bindFacing) == 2 {
				f = c.bindFacing / 2
			}
			c.setX(bt.pos[0] + f*c.bindPos[0])
			c.drawPos[0] += bt.drawPos[0] - bt.pos[0]
			c.oldPos[0] += bt.oldPos[0] - bt.pos[0]
			c.pushed = c.pushed || bt.pushed
			c.ghv.xoff = 0
		}
		if !math.IsNaN(float64(c.bindPos[1])) {
			c.setY(bt.pos[1] + c.bindPos[1])
			c.drawPos[1] += bt.drawPos[1] - bt.pos[1]
			c.oldPos[1] += bt.oldPos[1] - bt.pos[1]
			c.ghv.yoff = 0
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
func (c *Char) xScreenBound() {
	x := c.pos[0]
	if c.sf(CSF_screenbound) {
		min, max := c.getEdge(c.edge[0], true), -c.getEdge(c.edge[1], true)
		if c.facing > 0 {
			min, max = -max, -min
		}
		x = MaxF(min+sys.xmin, MinF(max+sys.xmax, x))
	}
	x = MaxF(sys.stage.leftbound, MinF(sys.stage.rightbound, x))
	c.setPosX(x)
}
func (c *Char) gethitBindClear() {
	if c.isBound() {
		c.setBindTime(0)
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
			if e := sys.charList.get(hb[0]); e != nil {
				e.removeTarget(c.id)
				if explremove {
					c.enemyExplodsRemove(e.playerNo)
				}
			}
		}
		c.gethitBindClear()
	}
	c.ghv.hitBy = c.ghv.hitBy[:0]
}
func (c *Char) offsetX() float32 {
	return float32(c.size.draw.offset[0])*c.facing + c.offset[0]
}
func (c *Char) offsetY() float32 {
	return float32(c.size.draw.offset[1]) + c.offset[1]
}
func (c *Char) clsnCheck(atk *Char, c1atk, c1slf bool) bool {
	if atk.curFrame == nil || c.curFrame == nil {
		return false
	}
	var clsn1, clsn2 []float32
	if c1atk {
		clsn1 = atk.curFrame.Clsn1()
	} else {
		clsn1 = atk.curFrame.Clsn2()
	}
	if c1slf {
		clsn2 = c.curFrame.Clsn1()
	} else {
		clsn2 = c.curFrame.Clsn2()
	}
	return sys.clsnHantei(clsn1, atk.clsnScale,
		[...]float32{atk.pos[0] + atk.offsetX(), atk.pos[1] + atk.offsetY()},
		atk.facing, clsn2, c.clsnScale, [...]float32{c.pos[0] + c.offsetX(),
			c.pos[1] + c.offsetY()}, c.facing)
}
func (c *Char) action() {
	if c.minus != 2 || c.sf(CSF_destroy) {
		return
	}
	p := false
	if c.cmd != nil {
		if sys.super > 0 {
			p = c.superMovetime == 0
		} else if sys.pause > 0 && c.pauseMovetime == 0 {
			p = true
		}
	}
	c.acttmp = -int8(Btoi(p)) * 2
	c.unsetSCF(SCF_guard)
	if !(c.scf(SCF_ko) || c.ctrlOver()) && (c.scf(SCF_ctrl) || c.ss.no == 52) &&
		c.ss.moveType == MT_I && c.cmd != nil &&
		(sys.autoguard[c.playerNo] || c.cmd[0].Buffer.B > 0) &&
		(c.ss.stateType == ST_S && !c.sf(CSF_nostandguard) ||
			c.ss.stateType == ST_C && !c.sf(CSF_nocrouchguard) ||
			c.ss.stateType == ST_A && !c.sf(CSF_noairguard)) {
		c.setSCF(SCF_guard)
	}
	if !p {
		if c.palfx != nil {
			c.palfx.step()
		}
		if c.keyctrl && c.cmd != nil {
			if c.ss.stateType == ST_A {
				if c.cmd[0].Buffer.U < 0 {
					c.setSCF(SCF_airjump)
				}
			} else {
				c.airJumpCount = 0
				c.unsetSCF(SCF_airjump)
			}
			if c.canCtrl() && (c.key >= 0 || c.helperIndex == 0) {
				if !sys.roundEnd() && c.ss.stateType == ST_S && c.cmd[0].Buffer.U > 0 {
					if c.ss.no != 40 {
						c.changeState(40, -1, -1)
					}
				} else if c.ss.stateType == ST_A && c.scf(SCF_airjump) &&
					c.pos[1] <= float32(c.gi().movement.airjump.height) &&
					c.airJumpCount < c.gi().movement.airjump.num &&
					c.cmd[0].Buffer.U > 0 {
					if c.ss.no != 45 {
						c.airJumpCount++
						c.unsetSCF(SCF_airjump)
						c.changeState(45, -1, -1)
					}
				} else {
					if c.ss.stateType == ST_S && c.cmd[0].Buffer.D > 0 {
						if c.ss.no != 10 {
							c.changeState(10, -1, -1)
						}
					} else if c.ss.stateType == ST_C && c.cmd[0].Buffer.D < 0 {
						if c.ss.no != 12 {
							c.changeState(12, -1, -1)
						}
					} else if !c.sf(CSF_nowalk) && c.ss.stateType == ST_S &&
						(c.cmd[0].Buffer.F > 0 || !(c.inguarddist && c.scf(SCF_guard)) &&
							c.cmd[0].Buffer.B > 0) {
						if c.ss.no != 20 {
							c.changeState(20, -1, -1)
						}
					} else if c.ss.no != 20 &&
						c.cmd[0].Buffer.B < 0 && c.cmd[0].Buffer.F < 0 {
						c.changeState(0, -1, -1)
					}
					if c.inguarddist && c.scf(SCF_guard) && c.cmd[0].Buffer.B > 0 &&
						!c.inGuardState() {
						c.changeState(120, -1, -1)
					}
				}
			} else if c.scf(SCF_ctrl) {
				switch c.ss.no {
				case 11:
					c.changeState(12, -1, -1)
				case 20:
					c.changeState(0, -1, -1)
				}
			}
		}
		if !c.hitPause() {
			if !c.sf(CSF_noautoturn) && c.ss.no == 52 {
				c.furimuki()
			}
			if !sys.roundEnd() {
				if c.alive() && c.life > 0 {
					c.unsetSCF(SCF_over | SCF_ko_round_middle)
				}
				if c.ss.no == 5150 || c.scf(SCF_over) {
					c.setSCF(SCF_ko_round_middle)
				}
			}
			if c.ss.no == 5150 {
				c.setSCF(SCF_over)
			}
			c.specialFlag = 0
			if c.player {
				if c.alive() || !c.scf(SCF_over) || !c.scf(SCF_ko_round_middle) {
					c.setSF(CSF_screenbound | CSF_movecamera_x | CSF_movecamera_y)
					if (c.alive() || !c.scf(SCF_over)) && c.roundState() > 0 {
						c.setSF(CSF_playerpush)
					}
				}
			}
			c.angleScalse = [...]float32{1, 1}
			c.attackDist = float32(c.size.attack.dist)
			c.offset = [2]float32{}
			for i, hb := range c.hitby {
				if hb.time > 0 {
					c.hitby[i].time--
				}
			}
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
		c.unsetSF(CSF_noautoturn)
		if c.gi().ver[0] == 1 {
			c.unsetSF(CSF_assertspecial | CSF_angledraw)
			c.angleScalse = [...]float32{1, 1}
			c.offset = [2]float32{}
		}
		c.minus = -3
		if c.ss.sb.playerNo == c.playerNo && c.player {
			if sb, ok := c.gi().states[-3]; ok {
				sb.run(c)
			}
		}
		c.minus = -2
		if c.player {
			if sb, ok := c.gi().states[-2]; ok {
				sb.run(c)
			}
		}
		c.minus = -1
		if c.keyctrl && c.ss.sb.playerNo == c.playerNo {
			if sb, ok := c.gi().states[-1]; ok {
				sb.run(c)
			}
		}
		c.minus = 0
		c.stateChange2()
		c.ss.sb.run(c)
		if !c.hitPause() {
			if c.ss.no == 5110 && c.recoverTime <= 0 && c.alive() {
				c.changeState(5120, -1, -1)
			}
			for c.ss.no == 140 || c.anim == nil || len(c.anim.frames) == 0 ||
				c.ss.time >= c.anim.totaltime {
				c.changeState(Btoi(c.ss.stateType == ST_C)*11+
					Btoi(c.ss.stateType == ST_A)*51, -1, -1)
			}
			for {
				c.posUpdate()
				if c.ss.physics != ST_A || c.vel[1] <= 0 || c.pos[1] < 0 ||
					c.ss.no == 105 {
					break
				}
				c.changeState(52, -1, -1)
			}
			c.ss.time++
			if c.mctime > 0 {
				c.mctime++
			}
			c.setFacing(c.p1facing)
			c.p1facing = 0
			if c.anim != nil {
				c.curFrame = c.anim.CurrentFrame()
			} else {
				c.curFrame = nil
			}
		}
		if c.ghv.damage != 0 {
			if c.ss.moveType == MT_H {
				c.lifeAdd(-float64(c.ghv.damage), true, true)
			}
			c.ghv.damage = 0
		}
		if c.helperIndex == 0 && c.gi().pctime >= 0 {
			c.gi().pctime++
		}
	}
	c.xScreenBound()
	if !p {
		for _, tid := range c.targets {
			if t := sys.charList.get(tid); t != nil && t.bindToId == c.id {
				t.bind()
			}
		}
	}
	c.minus = 1
	c.acttmp += int8(Btoi(!c.pause() && !c.hitPause())) -
		int8(Btoi(c.hitPause()))
	if !c.hitPause() {
		if !c.sf(CSF_frontwidth) {
			c.width[0] = c.defFW()
		}
		if !c.sf(CSF_backwidth) {
			c.width[1] = c.defBW()
		}
		if !c.sf(CSF_frontedge) {
			c.edge[0] = 0
		}
		if !c.sf(CSF_backedge) {
			c.edge[1] = 0
		}
	}
}
func (c *Char) update(cvmin, cvmax,
	highest, lowest, leftest, rightest *float32) {
	if sys.tickFrame() {
		if c.sf(CSF_destroy) {
			c.destroy()
			return
		}
		if c.acttmp > 0 {
			if c.anim != nil {
				c.anim.UpdateSprite()
			}
			if !c.isBound() {
				c.bind()
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
			}
		}
		if c.ss.moveType == MT_H {
			if sys.super <= 0 && sys.pause <= 0 {
				c.superMovetime, c.pauseMovetime = 0, 0
			}
			c.hittmp = int8(Btoi(c.ghv.fallf)) + 1
			if c.acttmp > 0 && (c.ss.no == 5100 || c.ss.no == 5070) &&
				c.ss.time == 1 {
				c.defenceMul *= c.gi().data.fall.defence_mul
				c.ghv.fallcount++
			}
		}
		if c.acttmp > 0 && c.ss.moveType != MT_H || c.roundState() == 2 &&
			c.scf(SCF_ko) && c.scf(SCF_over) {
			c.exitTarget(true)
		}
		c.atktmp = int8(Btoi((c.ss.moveType != MT_I ||
			c.hitdef.reversal_attr > 0) && !c.hitPause()))
		c.hoIdx = -1
		if c.acttmp > 0 {
			if c.inGuardState() {
				c.setSCF(SCF_guard)
			}
			if c.ss.moveType == MT_H {
				if c.ghv.guarded {
					c.getcombo = 0
				}
				if c.ghv.hitshaketime > 0 {
					c.ghv.hitshaketime--
				}
				if c.ghv.hitshaketime <= 0 && c.ghv.hittime >= 0 {
					c.ghv.hittime--
				}
				if c.ghv.fallf {
					c.fallTime++
				}
			} else {
				if c.hittmp > 0 {
					c.hittmp = 0
				}
				c.defenceMul = float32(c.gi().data.defence) / 100
				c.ghv.hittime = -1
				c.ghv.hitshaketime = 0
				c.ghv.fallf = false
				c.ghv.fallcount = 0
				c.ghv.hitid = -1
				c.getcombo = 0
			}
			if (c.ss.moveType == MT_H || c.ss.no == 52) && c.pos[1] == 0 &&
				AbsF(c.pos[0]-c.oldPos[0]) >= 1 && c.ss.time%3 == 0 {
				c.makeDust(0, 0)
			}
		}
	}
	if sys.tickNextFrame() {
		c.pushed = false
	}
	if c.acttmp > 0 {
		spd := sys.tickInterpola()
		if c.pushed {
			spd = 0
		}
		if !c.sf(CSF_posfreeze) {
			for i := 0; i < 2; i++ {
				c.drawPos[i] = c.pos[i] - (c.pos[i]-c.oldPos[i])*(1-spd)
			}
		}
	}
	min, max := c.getEdge(c.edge[0], true), -c.getEdge(c.edge[1], true)
	if c.facing > 0 {
		min, max = -max, -min
	}
	if c.sf(CSF_screenbound) {
		c.drawPos[0] = MaxF(min+sys.xmin, MinF(max+sys.xmax, c.drawPos[0]))
	}
	if c.sf(CSF_movecamera_x) {
		*leftest = MaxF(sys.xmin, MinF(c.drawPos[0]-min, *leftest))
		*rightest = MinF(sys.xmax, MaxF(c.drawPos[0]-max, *rightest))
		if c.acttmp > 0 && !c.sf(CSF_posfreeze) &&
			(c.bindTime == 0 || math.IsNaN(float64(c.bindPos[0]))) {
			*cvmin = MinF(*cvmin, c.vel[0]*c.facing)
			*cvmax = MaxF(*cvmax, c.vel[0]*c.facing)
		}
	}
	if c.sf(CSF_movecamera_y) {
		*highest = MinF(c.drawPos[1], *highest)
		*lowest = MinF(0, MaxF(c.drawPos[1], *lowest))
	}
}
func (c *Char) tick() {
	if c.acttmp > 0 && c.anim != nil {
		c.anim.Action()
	}
	if c.bindTime > 0 {
		if c.isBound() {
			if bt := sys.charList.get(c.bindToId); bt != nil && !bt.pause() {
				c.setBindTime(c.bindTime - 1)
			}
		} else {
			if !c.pause() {
				c.setBindTime(c.bindTime - 1)
			}
		}
	}
	if c.cmd == nil {
		c.cmd = sys.chars[c.playerNo][0].cmd
	}
	if c.hitdefContact {
		if c.hitdef.hitonce != 0 || c.moveReversed() != 0 {
			c.hitdef.invalidate(c.ss.stateType)
		}
		c.hitdefContact = false
	} else if c.hitdef.lhit {
		c.hitdef.attr = c.hitdef.attr&^int32(ST_MASK) | int32(c.ss.stateType)
		c.hitdef.lhit = false
	}
	if c.mctime < 0 {
		c.mctime = 1
		if c.mctype == MC_Hit {
			c.juggle = 0
			c.hitCount += c.hitdef.numhits
		}
	}
	if c.sf(CSF_gethit) {
		c.ss.moveType = MT_H
		if c.hitPauseTime > 0 {
			c.ss.clearWw()
		}
		c.hitPauseTime = 0
		if c.hoIdx >= 0 && c.ho[c.hoIdx].forceair {
			c.ss.stateType = ST_A
		}
		pn := c.playerNo
		if c.ghv.p2getp1state {
			pn = c.ghv.playerNo
		}
		if c.stchtmp {
			c.ss.prevno = 0
		} else if c.ss.stateType == ST_L {
			c.changeStateEx(5080, pn, -1, 0)
		} else if c.ghv.guarded && (c.ghv.damage < c.life || sys.sf(GSF_noko)) {
			switch c.ss.stateType {
			case ST_S:
				c.selfState(150, -1, 0)
			case ST_C:
				c.selfState(152, -1, 0)
			case ST_A:
				c.selfState(154, -1, 0)
			}
		} else if c.ghv._type == HT_Trip {
			c.changeStateEx(5070, pn, -1, 0)
		} else {
			if c.ghv.forcestand && c.ss.stateType == ST_C {
				c.ss.stateType = ST_S
			}
			switch c.ss.stateType {
			case ST_S:
				c.changeStateEx(5000, pn, -1, 0)
			case ST_C:
				c.changeStateEx(5010, pn, -1, 0)
			case ST_A:
				c.changeStateEx(5020, pn, -1, 0)
			}
		}
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
		if c.hitPauseTime <= 0 && c.ss.stateType == ST_L && c.recoverTime > 0 &&
			c.ss.sb.playerNo == c.playerNo && (c.cmd[0].Buffer.Bb == 1 ||
			c.cmd[0].Buffer.Db == 1 || c.cmd[0].Buffer.Fb == 1 ||
			c.cmd[0].Buffer.Ub == 1 || c.cmd[0].Buffer.ab == 1 ||
			c.cmd[0].Buffer.bb == 1 || c.cmd[0].Buffer.cb == 1 ||
			c.cmd[0].Buffer.xb == 1 || c.cmd[0].Buffer.yb == 1 ||
			c.cmd[0].Buffer.zb == 1 || c.cmd[0].Buffer.sb == 1) {
			c.recoverTime -= Rand(1, (c.recoverTime+1)/2)
		}
		if !c.stchtmp {
			if c.helperIndex == 0 && (c.alive() || c.ss.no == 0) && c.life <= 0 &&
				c.ss.moveType != MT_H && !sys.sf(GSF_noko) {
				c.ghv.fallf = true
				c.selfState(5030, -1, -1)
				c.ss.time = 1
			} else if c.ss.no == 5150 && c.ss.time >= 90 && c.alive() {
				c.selfState(5120, -1, -1)
			}
		}
	}
	if !c.hitPause() {
		if c.life <= 0 && !sys.sf(GSF_noko) {
			if !sys.sf(GSF_nokosnd) && c.alive() {
				vo := int32(0)
				if c.gi().ver[0] == 1 {
					vo = 100
				}
				c.playSound(false, false, false, 11, 0, -1, vo, 0, 1, &c.pos[0])
			}
			c.setSCF(SCF_ko)
		}
		if c.ss.moveType != MT_H {
			c.recoverTime = c.gi().data.liedown.time
		}
		if c.ss.no == 5110 && c.recoverTime > 0 && !c.pause() {
			c.recoverTime--
		}
	}
}
func (c *Char) cueDraw() {
	if c.helperIndex < 0 || c.id < 0 {
		return
	}
	if sys.clsnDraw && c.curFrame != nil {
		x, y := c.pos[0]+c.offsetX(), c.pos[1]+c.offsetY()
		xs, ys := c.facing*c.clsnScale[0], c.clsnScale[1]
		if clsn := c.curFrame.Clsn1(); len(clsn) > 0 && c.atktmp != 0 {
			sys.drawc1.Add(clsn, x, y, xs, ys)
		}
		if clsn := c.curFrame.Clsn2(); len(clsn) > 0 {
			hb, mtk := false, false
			for _, h := range c.hitby {
				if h.time != 0 {
					hb = true
					mtk = mtk || h.falg&int32(ST_SCA) == 0 || h.falg&int32(AT_ALL) == 0
				}
			}
			if mtk {
				sys.drawc2mtk.Add(clsn, x, y, xs, ys)
			} else if hb {
				sys.drawc2sp.Add(clsn, x, y, xs, ys)
			} else {
				sys.drawc2.Add(clsn, x, y, xs, ys)
			}
		}
		if c.sf(CSF_playerpush) {
			sys.drawwh.Add([]float32{-c.width[1], -c.height(), c.width[0], 0},
				c.pos[0], c.pos[1], c.facing, 1)
		}
	}
	if c.anim != nil {
		pos := [...]float32{c.drawPos[0] + c.offsetX(), c.drawPos[1] + c.offsetY()}
		scl := [...]float32{c.facing * c.size.xscale, c.size.yscale}
		var agl float32
		if c.sf(CSF_angledraw) {
			c.angleset = c.angleset || c.angle != 0
			if c.angleset {
				agl = c.angle
			} else {
				agl = 360
			}
		} else {
			c.angleset = false
		}
		rec := sys.tickNextFrame() && c.acttmp > 0
		sdf := func() *SprData {
			sd := &SprData{c.anim, c.getPalfx(), pos,
				scl, c.alpha, c.sprPriority, agl, c.angleScalse, c.angleset,
				false, c.playerNo == sys.superplayer, c.gi().ver[0] != 1}
			if !c.sf(CSF_trans) {
				sd.alpha[0] = -1
			}
			return sd
		}
		if c.sf(CSF_invisible) {
			if rec {
				c.aimg.recAfterImg(sdf())
			}
		} else {
			if c.gi().ver[0] != 1 && c.sf(CSF_angledraw) && !c.sf(CSF_trans) {
				c.setSF(CSF_trans)
				c.alpha = [...]int32{255, 0}
			}
			sd := sdf()
			c.aimg.recAndCue(sd, rec)
			if c.ghv.hitshaketime > 0 && c.ss.time&1 != 0 {
				sd.pos[0] -= c.facing
			}
			var sc, sa int32 = -1, 255
			if c.sf(CSF_noshadow) {
				sc = 0
			}
			if c.sf(CSF_trans) {
				sa = c.alpha[0]
			}
			sys.sprites.add(sd, sc, sa, 0)
		}
	}
	if sys.tickNextFrame() {
		if c.roundState() == 4 {
			c.exitTarget(false)
		}
		if sys.supertime < 0 {
			if c.playerNo&1 != sys.superplayer&1 {
				c.defenceMul *= sys.superp2defmul
			}
			c.minus = 2
			c.oldPos = c.pos
		}
	}
}

type CharList struct {
	runOrder, drawOrder []*Char
}

func (cl *CharList) clear() {
	*cl = CharList{}
	sys.nextCharId = sys.helperMax
}
func (cl *CharList) add(c *Char) {
	sys.clearPlayerIdCache()
	cl.runOrder = append(cl.runOrder, c)
	i := 0
	for ; i < len(cl.drawOrder); i++ {
		if cl.drawOrder[i] == nil {
			cl.drawOrder[i] = c
			break
		}
	}
	if i >= len(cl.drawOrder) {
		cl.drawOrder = append(cl.drawOrder, c)
	}
}
func (cl *CharList) action(x float32, cvmin, cvmax,
	highest, lowest, leftest, rightest *float32) {
	sys.commandUpdate()
	for _, c := range cl.runOrder {
		if c.id < 0 {
			c.id = ^c.id
		}
		if c.ss.moveType == MT_A {
			c.action()
		}
	}
	for _, c := range cl.runOrder {
		if c.id < 0 {
			c.id = ^c.id
		}
		c.action()
	}
	sys.charUpdate(cvmin, cvmax, highest, lowest, leftest, rightest)
}
func (cl *CharList) update(cvmin, cvmax,
	highest, lowest, leftest, rightest *float32) {
	for _, c := range cl.runOrder {
		if c.id >= 0 {
			c.update(cvmin, cvmax, highest, lowest, leftest, rightest)
		}
	}
}
func (cl *CharList) clsn(getter *Char, proj bool) {
	var gxmin, gxmax float32
	if proj {
		for i, pr := range sys.projs {
			if i == getter.playerNo || len(sys.projs[0]) == 0 {
				continue
			}
			orgatktmp := sys.chars[i][0].atktmp
			sys.chars[i][0].atktmp = -1
			for _, p := range pr {
				if p.id < 0 || p.hits < 0 || p.hitdef.affectteam != 0 &&
					(getter.playerNo&1 != i&1) != (p.hitdef.affectteam > 0) {
					continue
				}
				unimplemented()
			}
			sys.chars[i][0].atktmp = orgatktmp
		}
	} else {
		gxmin = getter.getEdge(getter.edge[0], true)
		gxmax = -getter.getEdge(getter.edge[1], true)
		if getter.facing > 0 {
			gxmin, gxmax = -gxmax, -gxmin
		}
		gxmin += sys.xmin
		gxmax += sys.xmax
		getter.inguarddist = false
		getter.unsetSF(CSF_gethit)
		gl, gr := -getter.width[0], getter.width[1]
		if getter.facing > 0 {
			gl, gr = -gr, -gl
		}
		gl += getter.pos[0]
		gr += getter.pos[0]
		getter.enemyNearClear()
		for _, c := range cl.runOrder {
			if c.id < 0 {
				continue
			}
			contact := 0
			if c.atktmp != 0 && c.id != getter.id && (c.hitdef.affectteam == 0 ||
				(getter.playerNo&1 != c.playerNo&1) == (c.hitdef.affectteam > 0)) {
				unimplemented()
			}
			if getter.playerNo&1 != c.playerNo&1 && getter.sf(CSF_playerpush) &&
				c.sf(CSF_playerpush) && (getter.ss.stateType == ST_A ||
				getter.pos[1]-c.pos[1] < getter.height()) &&
				(c.ss.stateType == ST_A || c.pos[1]-getter.pos[1] < c.height()) {
				cl, cr := -c.width[0], c.width[1]
				if c.facing > 0 {
					cl, cr = -cr, -cl
				}
				cl += c.pos[0]
				cr += c.pos[0]
				if gl < cr && cl < gr && (contact > 0 ||
					getter.clsnCheck(c, false, false)) {
					getter.pushed, c.pushed = true, true
					tmp := getter.distX(c)
					if tmp == 0 {
						if getter.pos[1] > c.pos[1] {
							tmp = getter.facing
						} else {
							tmp = -c.facing
						}
					}
					if tmp > 0 {
						getter.pos[0] -= (gr - cl) * 0.5
						c.pos[0] += (gr - cl) * 0.5
					} else {
						getter.pos[0] += (cr - gl) * 0.5
						c.pos[0] -= (cr - gl) * 0.5
					}
					if getter.sf(CSF_screenbound) {
						getter.pos[0] = MaxF(gxmin, MinF(gxmax, getter.pos[0]))
					}
					if c.sf(CSF_screenbound) {
						l, r := c.getEdge(c.edge[0], true), -c.getEdge(c.edge[1], true)
						if c.facing > 0 {
							l, r = -r, -l
						}
						c.pos[0] = MaxF(l+sys.xmin, MinF(r+sys.xmax, c.pos[0]))
					}
					getter.pos[0] = MaxF(sys.stage.leftbound, MinF(sys.stage.rightbound,
						getter.pos[0]))
					c.pos[0] = MaxF(sys.stage.leftbound, MinF(sys.stage.rightbound,
						c.pos[0]))
					getter.drawPos[0], c.drawPos[0] = getter.pos[0], c.pos[0]
				}
			}
		}
	}
}
func (cl *CharList) getHit() {
	for _, c := range cl.runOrder {
		if c.id >= 0 {
			cl.clsn(c, false)
		}
	}
	for _, c := range cl.runOrder {
		if c.id >= 0 {
			cl.clsn(c, true)
		}
	}
}
func (cl *CharList) tick() {
	sys.gameTime++
	for _, c := range cl.runOrder {
		if c.id >= 0 {
			c.tick()
		}
	}
}
func (cl *CharList) cueDraw() {
	sys.gameTime++
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
	if c, ok := sys.playerIdCache[id]; ok {
		return c
	}
	for _, c := range cl.runOrder {
		if c.id == id {
			sys.playerIdCache[id] = c
			return c
		}
	}
	sys.playerIdCache[id] = nil
	return nil
}
func (cl *CharList) enemyNear(c *Char, n int32, p2 bool) *Char {
	if n < 0 {
		return nil
	}
	cache := &c.enemyNear[Btoi(p2)]
	if int(n) < len(*cache) {
		return (*cache)[n]
	}
	*cache = (*cache)[:0]
	var add func(*Char, int)
	add = func(e *Char, idx int) {
		for i := idx; i <= int(n); i++ {
			if i >= len(*cache) {
				*cache = append(*cache, e)
				return
			}
			if p2 && !e.scf(SCF_ko_round_middle) &&
				(*cache)[i].scf(SCF_ko_round_middle) || (!p2 ||
				e.scf(SCF_ko_round_middle) == (*cache)[i].scf(SCF_ko_round_middle)) &&
				AbsF(c.distX(e)) < AbsF(c.distX((*cache)[i])) {
				add((*cache)[i], i+1)
				(*cache)[i] = e
			}
		}
	}
	for _, e := range cl.runOrder {
		if e.playerNo&1 != c.playerNo&1 {
			add(e, 0)
		}
	}
	if int(n) >= len(*cache) {
		return nil
	}
	return (*cache)[n]
}
