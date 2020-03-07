package main

import (
	"fmt"
	"math"
	"strings"
)

type FinishType int32

const (
	FT_NotYet FinishType = iota
	FT_KO
	FT_DKO
	FT_TO
	FT_TODraw
)

type WinType int32

const (
	WT_N WinType = iota
	WT_S
	WT_H
	WT_C
	WT_T
	WT_Throw
	WT_Suicide
	WT_Teammate
	WT_Perfect
	WT_NumTypes
	WT_PN
	WT_PS
	WT_PH
	WT_PC
	WT_PT
	WT_PThrow
	WT_PSuicide
	WT_PTeammate
)

type RoundType int32

const (
	RT_Normal RoundType = iota
	RT_Deciding
	RT_Final
)

func (wt *WinType) SetPerfect() {
	if *wt >= WT_N && *wt <= WT_Teammate {
		*wt += WT_PN - WT_N
	}
}

type BgTextSnd struct {
	pos         [2]int32
	text_font   [3]int32
	text_text   string
	text_lay    Layout
	bg          AnimLayout
	time        int32
	displaytime int32
	snd         [2]int32
	sndtime     int32
	cnt         int32
}
func newBgTextSnd() BgTextSnd {
	return BgTextSnd{text_font: [3]int32{-1}, snd: [2]int32{-1}}
}
func readBgTextSnd(pre string, is IniSection,
	sff *Sff, at AnimationTable) BgTextSnd {
	bts := newBgTextSnd()
	is.ReadI32(pre+"pos", &bts.pos[0], &bts.pos[1])
	is.ReadI32(pre+"text.font", &bts.text_font[0], &bts.text_font[1],
		&bts.text_font[2])
	bts.text_text, _ = is.getString(pre+"text.text")
	bts.text_lay = *ReadLayout(pre+"text.", is, 0)
	bts.bg = *ReadAnimLayout(pre+"bg.", is, sff, at, 0)
	is.ReadI32(pre+"time", &bts.time)
	is.ReadI32(pre+"displaytime", &bts.displaytime)
	is.ReadI32(pre+"snd", &bts.snd[0], &bts.snd[1])
	is.ReadI32(pre+"sndtime", &bts.sndtime)
	return bts
}
func (bts *BgTextSnd) step(snd *Snd)  {
	if bts.cnt == bts.sndtime {
		snd.play(bts.snd)
	}
	if bts.cnt >= bts.time {
		bts.bg.Action()
	}
	bts.cnt++
}
func (bts *BgTextSnd) reset() {
	bts.cnt = 0
	bts.bg.Reset()
}
func (bts *BgTextSnd) bgDraw(layerno int16) {
	if bts.cnt > bts.time && bts.cnt <= bts.time + bts.displaytime {
		bts.bg.DrawScaled(float32(bts.pos[0])+sys.lifebarOffsetX, float32(bts.pos[1]), layerno, sys.lifebarScale)
	}
}
func (bts *BgTextSnd) draw(layerno int16, f []*Fnt) {
	if bts.cnt > bts.time && bts.cnt <= bts.time + bts.displaytime &&
		bts.text_font[0] >= 0 && int(bts.text_font[0]) < len(f) {
		bts.text_lay.DrawText(float32(bts.pos[0])+sys.lifebarOffsetX, float32(bts.pos[1]), sys.lifebarScale, layerno,
			bts.text_text, f[bts.text_font[0]], bts.text_font[1], bts.text_font[2])
	}
}

type HealthBar struct {
	pos        [2]int32
	range_x    [2]int32
	bg0        AnimLayout
	bg1        AnimLayout
	bg2        AnimLayout
	mid        AnimLayout
	front      AnimLayout
	oldlife    float32
	midlife    float32
	midlifeMin float32
	mlifetime  int32
}

func readHealthBar(pre string, is IniSection,
	sff *Sff, at AnimationTable) *HealthBar {
	hb := &HealthBar{oldlife: 1, midlife: 1, midlifeMin: 1}
	is.ReadI32(pre+"pos", &hb.pos[0], &hb.pos[1])
	is.ReadI32(pre+"range.x", &hb.range_x[0], &hb.range_x[1])
	hb.bg0 = *ReadAnimLayout(pre+"bg0.", is, sff, at, 0)
	hb.bg1 = *ReadAnimLayout(pre+"bg1.", is, sff, at, 0)
	hb.bg2 = *ReadAnimLayout(pre+"bg2.", is, sff, at, 0)
	hb.mid = *ReadAnimLayout(pre+"mid.", is, sff, at, 0)
	hb.front = *ReadAnimLayout(pre+"front.", is, sff, at, 0)
	return hb
}
func (hb *HealthBar) step(life float32, gethit bool) {
	if len(hb.mid.anim.frames) > 0 && gethit {
		if hb.mlifetime < 30 {
			hb.mlifetime = 30
			hb.midlife, hb.midlifeMin = hb.oldlife, hb.oldlife
		}
	} else {
		if hb.mlifetime > 0 {
			hb.mlifetime--
		}
		if len(hb.mid.anim.frames) > 0 && hb.mlifetime <= 0 &&
			life < hb.midlifeMin {
			hb.midlifeMin += (life - hb.midlifeMin) *
				(1 / (12 - (life-hb.midlifeMin)*144))
		} else {
			hb.midlifeMin = life
		}
		if (len(hb.mid.anim.frames) == 0 || hb.mlifetime <= 0) &&
			hb.midlife > hb.midlifeMin {
			hb.midlife += (hb.midlifeMin - hb.midlife) / 8
		}
		hb.oldlife = life
	}
	mlmin := MaxF(hb.midlifeMin, life)
	if hb.midlife < mlmin {
		hb.midlife += (mlmin - hb.midlife) / 2
	}
	hb.bg0.Action()
	hb.bg1.Action()
	hb.bg2.Action()
	hb.mid.Action()
	hb.front.Action()
}
func (hb *HealthBar) reset() {
	hb.bg0.Reset()
	hb.bg1.Reset()
	hb.bg2.Reset()
	hb.mid.Reset()
	hb.front.Reset()
}
func (hb *HealthBar) bgDraw(layerno int16) {
	hb.bg0.DrawScaled(float32(hb.pos[0])+sys.lifebarOffsetX, float32(hb.pos[1]), layerno, sys.lifebarScale)
	hb.bg1.DrawScaled(float32(hb.pos[0])+sys.lifebarOffsetX, float32(hb.pos[1]), layerno, sys.lifebarScale)
	hb.bg2.DrawScaled(float32(hb.pos[0])+sys.lifebarOffsetX, float32(hb.pos[1]), layerno, sys.lifebarScale)
}
func (hb *HealthBar) draw(layerno int16, life float32) {
	var MidPos = (float32(sys.gameWidth-320) / 2)
	width := func(life float32) (r [4]int32) {
		r = sys.scrrect
		if hb.range_x[0] < hb.range_x[1] {
			r[0] = int32((((float32(hb.pos[0]+hb.range_x[0])+sys.lifebarOffsetX)*sys.lifebarScale)+
				MidPos)*sys.widthScale + 0.5)
			r[2] = int32((float32(hb.range_x[1]-hb.range_x[0]+1)*sys.lifebarScale)*life*
				sys.widthScale + 0.5)
		} else {
			r[2] = int32(((float32(hb.range_x[0]-hb.range_x[1]+1)*sys.lifebarScale)*life-(sys.lifebarOffsetX*sys.lifebarScale))*
				sys.widthScale + 0.5)
			r[0] = int32(((float32(hb.pos[0]+hb.range_x[0]+1)*sys.lifebarScale)+
				MidPos)*sys.widthScale+0.5) - r[2]
		}
		return
	}
	if len(hb.mid.anim.frames) == 0 || life > hb.midlife {
		life = hb.midlife
	}
	lr, mr := width(life), width(hb.midlife)
	if hb.range_x[0] < hb.range_x[1] {
		mr[0] += lr[2]
	}
	mr[2] -= Min(mr[2], lr[2])
	hb.mid.lay.DrawAnim(&mr, float32(hb.pos[0])+sys.lifebarOffsetX, float32(hb.pos[1]), sys.lifebarScale,
		layerno, &hb.mid.anim)
	hb.front.lay.DrawAnim(&lr, float32(hb.pos[0])+sys.lifebarOffsetX, float32(hb.pos[1]), sys.lifebarScale,
		layerno, &hb.front.anim)
}

type PowerBar struct {
	snd          *Snd
	pos          [2]int32
	range_x      [2]int32
	bg0          AnimLayout
	bg1          AnimLayout
	bg2          AnimLayout
	mid          AnimLayout
	front        AnimLayout
	counter_font [3]int32
	counter_lay  Layout
	level_snd    [3][2]int32
	midpower     float32
	midpowerMin  float32
	prevLevel    int32
}

func newPowerBar(snd *Snd) (pb *PowerBar) {
	pb = &PowerBar{snd: snd, counter_font: [3]int32{-1},
		level_snd: [...][2]int32{{-1}, {-1}, {-1}}}
	return
}
func readPowerBar(pre string, is IniSection,
	sff *Sff, at AnimationTable, snd *Snd) *PowerBar {
	pb := newPowerBar(snd)
	is.ReadI32(pre+"pos", &pb.pos[0], &pb.pos[1])
	is.ReadI32(pre+"range.x", &pb.range_x[0], &pb.range_x[1])
	pb.bg0 = *ReadAnimLayout(pre+"bg0.", is, sff, at, 0)
	pb.bg1 = *ReadAnimLayout(pre+"bg1.", is, sff, at, 0)
	pb.bg2 = *ReadAnimLayout(pre+"bg2.", is, sff, at, 0)
	pb.mid = *ReadAnimLayout(pre+"mid.", is, sff, at, 0)
	pb.front = *ReadAnimLayout(pre+"front.", is, sff, at, 0)
	is.ReadI32(pre+"counter.font", &pb.counter_font[0], &pb.counter_font[1],
		&pb.counter_font[2])
	pb.counter_lay = *ReadLayout(pre+"counter.", is, 0)
	for i := range pb.level_snd {
		if !is.ReadI32(fmt.Sprintf("%vlevel%v.snd", pre, i+1), &pb.level_snd[i][0],
			&pb.level_snd[i][1]) {
			is.ReadI32(fmt.Sprintf("level%v.snd", i+1), &pb.level_snd[i][0],
				&pb.level_snd[i][1])
		}
	}
	return pb
}
func (pb *PowerBar) step(power float32, level int32) {
	pb.midpower -= 1.0 / 144
	if power < pb.midpowerMin {
		pb.midpowerMin += (power - pb.midpowerMin) *
			(1 / (12 - (power-pb.midpowerMin)*144))
	} else {
		pb.midpowerMin = power
	}
	if pb.midpower < pb.midpowerMin {
		pb.midpower = pb.midpowerMin
	}
	if level > pb.prevLevel {
		i := Min(2, level-1)
		pb.snd.play(pb.level_snd[i])
	}
	pb.prevLevel = level
	pb.bg0.Action()
	pb.bg1.Action()
	pb.bg2.Action()
	pb.mid.Action()
	pb.front.Action()
}
func (pb *PowerBar) reset() {
	pb.bg0.Reset()
	pb.bg1.Reset()
	pb.bg2.Reset()
	pb.mid.Reset()
	pb.front.Reset()
}
func (pb *PowerBar) bgDraw(layerno int16) {
	pb.bg0.DrawScaled(float32(pb.pos[0])+sys.lifebarOffsetX, float32(pb.pos[1]), layerno, sys.lifebarScale)
	pb.bg1.DrawScaled(float32(pb.pos[0])+sys.lifebarOffsetX, float32(pb.pos[1]), layerno, sys.lifebarScale)
	pb.bg2.DrawScaled(float32(pb.pos[0])+sys.lifebarOffsetX, float32(pb.pos[1]), layerno, sys.lifebarScale)
}
func (pb *PowerBar) draw(layerno int16, power float32,
	level int32, f []*Fnt) {

	var MidPos = (float32(sys.gameWidth-320) / 2)
	width := func(power float32) (r [4]int32) {
		r = sys.scrrect
		if pb.range_x[0] < pb.range_x[1] {
			r[0] = int32((((float32(pb.pos[0]+pb.range_x[0])+sys.lifebarOffsetX)*sys.lifebarScale)+
				MidPos)*sys.widthScale + 0.5)
			r[2] = int32((float32(pb.range_x[1]-pb.range_x[0]+1)*sys.lifebarScale)*power*
				sys.widthScale + 0.5)
		} else {
			r[2] = int32(((float32(pb.range_x[0]-pb.range_x[1]+1)*sys.lifebarScale)*power-(sys.lifebarOffsetX*sys.lifebarScale))*
				sys.widthScale + 0.5)
			r[0] = int32(((float32(pb.pos[0]+pb.range_x[0]+1)*sys.lifebarScale)+
				MidPos)*sys.widthScale+0.5) - r[2]
		}
		return
	}
	pr, mr := width(power), width(pb.midpower)
	if pb.range_x[0] < pb.range_x[1] {
		mr[0] += pr[2]
	}
	mr[2] -= Min(mr[2], pr[2])
	pb.mid.lay.DrawAnim(&mr, float32(pb.pos[0])+sys.lifebarOffsetX, float32(pb.pos[1]), sys.lifebarScale,
		layerno, &pb.mid.anim)
	pb.front.lay.DrawAnim(&pr, float32(pb.pos[0])+sys.lifebarOffsetX, float32(pb.pos[1]), sys.lifebarScale,
		layerno, &pb.front.anim)

	if pb.counter_font[0] >= 0 && int(pb.counter_font[0]) < len(f) {
		pb.counter_lay.DrawText(float32(pb.pos[0])+sys.lifebarOffsetX, float32(pb.pos[1]), sys.lifebarScale,
			layerno, fmt.Sprintf("%v", level),
			f[pb.counter_font[0]], pb.counter_font[1], pb.counter_font[2])
	}
}

type LifeBarFace struct {
	pos               [2]int32
	bg                AnimLayout
	bg0               AnimLayout
	bg1               AnimLayout
	bg2               AnimLayout
	mid               AnimLayout
	front             AnimLayout
	face_spr          [2]int32
	face              *Sprite
	face_lay          Layout
	teammate_pos      [2]int32
	teammate_spacing  [2]int32
	teammate_bg       AnimLayout
	teammate_ko       AnimLayout
	teammate_face_spr [2]int32
	teammate_face     []*Sprite
	teammate_face_lay Layout
	numko             int32
	scale             float32
	teammate_scale    []float32
}

func newLifeBarFace() *LifeBarFace {
	return &LifeBarFace{face_spr: [2]int32{-1}, teammate_face_spr: [2]int32{-1}, scale: 1}
}
func readLifeBarFace(pre string, is IniSection,
	sff *Sff, at AnimationTable) *LifeBarFace {
	f := newLifeBarFace()
	is.ReadI32(pre+"pos", &f.pos[0], &f.pos[1])

	f.bg = *ReadAnimLayout(pre+"bg.", is, sff, at, 0)
	f.bg0 = *ReadAnimLayout(pre+"bg0.", is, sff, at, 0)
	f.bg1 = *ReadAnimLayout(pre+"bg1.", is, sff, at, 0)
	f.bg2 = *ReadAnimLayout(pre+"bg2.", is, sff, at, 0)
	f.bg2 = *ReadAnimLayout(pre+"mid.", is, sff, at, 0)
	f.bg2 = *ReadAnimLayout(pre+"front.", is, sff, at, 0)

	is.ReadI32(pre+"face.spr", &f.face_spr[0], &f.face_spr[1])
	f.face_lay = *ReadLayout(pre+"face.", is, 0)
	is.ReadI32(pre+"teammate.pos", &f.teammate_pos[0], &f.teammate_pos[1])
	is.ReadI32(pre+"teammate.spacing", &f.teammate_spacing[0],
		&f.teammate_spacing[1])
	f.teammate_bg = *ReadAnimLayout(pre+"teammate.bg.", is, sff, at, 0)
	f.teammate_ko = *ReadAnimLayout(pre+"teammate.ko.", is, sff, at, 0)
	is.ReadI32(pre+"teammate.face.spr", &f.teammate_face_spr[0],
		&f.teammate_face_spr[1])
	f.teammate_face_lay = *ReadLayout(pre+"teammate.face.", is, 0)
	return f
}
func (f *LifeBarFace) step() {
	f.bg.Action()
	f.bg0.Action()
	f.bg1.Action()
	f.bg2.Action()
	f.mid.Action()
	f.front.Action()
	f.teammate_bg.Action()
	f.teammate_ko.Action()
}
func (f *LifeBarFace) reset() {
	f.bg.Reset()
	f.bg0.Reset()
	f.bg1.Reset()
	f.bg2.Reset()
	f.mid.Reset()
	f.front.Reset()
	f.teammate_bg.Reset()
	f.teammate_ko.Reset()
}
func (f *LifeBarFace) bgDraw(layerno int16) {
	f.bg.DrawScaled(float32(f.pos[0])+sys.lifebarOffsetX, float32(f.pos[1]), layerno, sys.lifebarScale)
	f.bg0.DrawScaled(float32(f.pos[0])+sys.lifebarOffsetX, float32(f.pos[1]), layerno, sys.lifebarScale)
	f.bg1.DrawScaled(float32(f.pos[0])+sys.lifebarOffsetX, float32(f.pos[1]), layerno, sys.lifebarScale)
	f.bg2.DrawScaled(float32(f.pos[0])+sys.lifebarOffsetX, float32(f.pos[1]), layerno, sys.lifebarScale)
}
func (f *LifeBarFace) draw(layerno int16, fx *PalFX, superplayer bool) {
	ob := sys.brightness
	if superplayer {
		sys.brightness = 256
	}
	f.face_lay.DrawSprite((float32(f.pos[0])+sys.lifebarOffsetX)*sys.lifebarScale, float32(f.pos[1])*sys.lifebarScale, layerno,
		f.face, fx, f.scale*sys.lifebarPortraitScale)
	sys.brightness = ob
	i := int32(len(f.teammate_face)) - 1
	x := float32(f.teammate_pos[0] + f.teammate_spacing[0]*(i-1))
	y := float32(f.teammate_pos[1] + f.teammate_spacing[1]*(i-1))
	for ; i >= 0; i-- {
		if i != f.numko {
			f.teammate_bg.DrawScaled((x + sys.lifebarOffsetX), y, layerno, sys.lifebarScale)
			f.teammate_face_lay.DrawSprite((x+sys.lifebarOffsetX)*sys.lifebarScale, y*sys.lifebarScale, layerno, f.teammate_face[i], nil, f.teammate_scale[i]*sys.lifebarPortraitScale)
			if i < f.numko {
				f.teammate_ko.DrawScaled((x + sys.lifebarOffsetX), y, layerno, sys.lifebarScale)
			}
			x -= float32(f.teammate_spacing[0])
			y -= float32(f.teammate_spacing[1])
		}
	}

	f.mid.lay.DrawAnim(&sys.scrrect, float32(f.pos[0])+sys.lifebarOffsetX, float32(f.pos[1]), sys.lifebarScale, layerno, &f.mid.anim)
	f.front.lay.DrawAnim(&sys.scrrect, float32(f.pos[0])+sys.lifebarOffsetX, float32(f.pos[1]), sys.lifebarScale, layerno, &f.front.anim)
}

type LifeBarName struct {
	pos       [2]int32
	name_font [3]int32
	name_lay  Layout
	bg        AnimLayout
}

func newLifeBarName() *LifeBarName {
	return &LifeBarName{name_font: [3]int32{-1}}
}
func readLifeBarName(pre string, is IniSection,
	sff *Sff, at AnimationTable) *LifeBarName {
	n := newLifeBarName()
	is.ReadI32(pre+"pos", &n.pos[0], &n.pos[1])
	is.ReadI32(pre+"name.font", &n.name_font[0], &n.name_font[1],
		&n.name_font[2])
	n.name_lay = *ReadLayout(pre+"name.", is, 0)
	n.bg = *ReadAnimLayout(pre+"bg.", is, sff, at, 0)
	return n
}
func (n *LifeBarName) step()  { n.bg.Action() }
func (n *LifeBarName) reset() { n.bg.Reset() }
func (n *LifeBarName) bgDraw(layerno int16) {
	n.bg.DrawScaled(float32(n.pos[0])+sys.lifebarOffsetX, float32(n.pos[1]), layerno, sys.lifebarScale)
}
func (n *LifeBarName) draw(layerno int16, f []*Fnt, name string) {
	if n.name_font[0] >= 0 && int(n.name_font[0]) < len(f) {
		n.name_lay.DrawText((float32(n.pos[0]) + sys.lifebarOffsetX), float32(n.pos[1]), sys.lifebarScale, layerno, name,
			f[n.name_font[0]], n.name_font[1], n.name_font[2])
	}
}

type LifeBarWinIcon struct {
	pos           [2]int32
	iconoffset    [2]int32
	useiconupto   int32
	counter_font  [3]int32
	counter_lay   Layout
	bg0           AnimLayout
	icon          [WT_NumTypes]AnimLayout
	wins          []WinType
	numWins       int
	added, addedP *Animation
}

func newLifeBarWinIcon() *LifeBarWinIcon {
	return &LifeBarWinIcon{useiconupto: 4, counter_font: [3]int32{-1}}
}
func readLifeBarWinIcon(pre string, is IniSection,
	sff *Sff, at AnimationTable) *LifeBarWinIcon {
	wi := newLifeBarWinIcon()
	is.ReadI32(pre+"pos", &wi.pos[0], &wi.pos[1])
	is.ReadI32(pre+"iconoffset", &wi.iconoffset[0], &wi.iconoffset[1])
	is.ReadI32("useiconupto", &wi.useiconupto)
	is.ReadI32(pre+"counter.font", &wi.counter_font[0], &wi.counter_font[1],
		&wi.counter_font[2])
	wi.counter_lay = *ReadLayout(pre+"counter.", is, 0)
	wi.bg0 = *ReadAnimLayout(pre+"bg0.", is, sff, at, 0)
	wi.icon[WT_N] = *ReadAnimLayout(pre+"n.", is, sff, at, 0)
	wi.icon[WT_S] = *ReadAnimLayout(pre+"s.", is, sff, at, 0)
	wi.icon[WT_H] = *ReadAnimLayout(pre+"h.", is, sff, at, 0)
	wi.icon[WT_C] = *ReadAnimLayout(pre+"c.", is, sff, at, 0)
	wi.icon[WT_T] = *ReadAnimLayout(pre+"t.", is, sff, at, 0)
	wi.icon[WT_Throw] = *ReadAnimLayout(pre+"throw.", is, sff, at, 0)
	wi.icon[WT_Suicide] = *ReadAnimLayout(pre+"suicide.", is, sff, at, 0)
	wi.icon[WT_Teammate] = *ReadAnimLayout(pre+"teammate.", is, sff, at, 0)
	wi.icon[WT_Perfect] = *ReadAnimLayout(pre+"perfect.", is, sff, at, 0)
	return wi
}
func (wi *LifeBarWinIcon) add(wt WinType) {
	wi.wins = append(wi.wins, wt)
	if wt >= WT_PN {
		wi.addedP = &Animation{}
		*wi.addedP = wi.icon[WT_Perfect].anim
		wi.addedP.Reset()
		wt -= WT_PN
	}
	wi.added = &Animation{}
	*wi.added = wi.icon[wt].anim
	wi.added.Reset()
}
func (wi *LifeBarWinIcon) step(numwin int32) {
	wi.bg0.Action()
	if int(numwin) < len(wi.wins) {
		wi.wins = wi.wins[:numwin]
		wi.reset()
	}
	for i := range wi.icon {
		wi.icon[i].Action()
	}
	if wi.added != nil {
		wi.added.Action()
	}
	if wi.addedP != nil {
		wi.addedP.Action()
	}
}
func (wi *LifeBarWinIcon) reset() {
	wi.bg0.Reset()
	for i := range wi.icon {
		wi.icon[i].Reset()
	}
	wi.numWins = len(wi.wins)
	wi.added, wi.addedP = nil, nil
}
func (wi *LifeBarWinIcon) clear() { wi.wins = nil }
func (wi *LifeBarWinIcon) draw(layerno int16, f []*Fnt) {
	for i := 0; i < int(math.Min(float64(wi.useiconupto), float64(sys.lifebar.ro.match_wins))); i++ {
		wi.bg0.DrawScaled(float32(wi.pos[0]+wi.iconoffset[0]*int32(i))+sys.lifebarOffsetX,
				float32(wi.pos[1]+wi.iconoffset[1]*int32(i)), layerno, sys.lifebarScale)
	}
	if len(wi.wins) > int(wi.useiconupto) {
		if wi.counter_font[0] >= 0 && int(wi.counter_font[0]) < len(f) {
			wi.counter_lay.DrawText(float32(wi.pos[0])+sys.lifebarOffsetX, float32(wi.pos[1]), sys.lifebarScale,
				layerno, fmt.Sprintf("%v", len(wi.wins)),
				f[wi.counter_font[0]], wi.counter_font[1], wi.counter_font[2])
		}
	} else {
		i := 0
		for ; i < wi.numWins; i++ {
			wt, p := wi.wins[i], false
			if wt >= WT_PN {
				wt -= WT_PN
				p = true
			}
			wi.icon[wt].DrawScaled(float32(wi.pos[0]+wi.iconoffset[0]*int32(i))+sys.lifebarOffsetX,
				float32(wi.pos[1]+wi.iconoffset[1]*int32(i)), layerno, sys.lifebarScale)
			if p {
				wi.icon[WT_Perfect].DrawScaled(float32(wi.pos[0]+wi.iconoffset[0]*int32(i))+sys.lifebarOffsetX,
					float32(wi.pos[1]+wi.iconoffset[1]*int32(i)), layerno, sys.lifebarScale)
			}
		}
		if wi.added != nil {
			wt, p := wi.wins[i], false
			if wi.addedP != nil {
				wt -= WT_PN
				p = true
			}
			wi.icon[wt].lay.DrawAnim(&sys.scrrect,
				float32(wi.pos[0]+wi.iconoffset[0]*int32(i))+sys.lifebarOffsetX,
				float32(wi.pos[1]+wi.iconoffset[1]*int32(i)), sys.lifebarScale, layerno, wi.added)
			if p {
				wi.icon[WT_Perfect].lay.DrawAnim(&sys.scrrect,
					float32(wi.pos[0]+wi.iconoffset[0]*int32(i))+sys.lifebarOffsetX,
					float32(wi.pos[1]+wi.iconoffset[1]*int32(i)), sys.lifebarScale, layerno, wi.addedP)
			}
		}
	}
}

type LifeBarTime struct {
	pos            [2]int32
	counter_font   [3]int32
	counter_lay    Layout
	bg             AnimLayout
	framespercount int32
}

func newLifeBarTime() *LifeBarTime {
	return &LifeBarTime{counter_font: [3]int32{-1}, framespercount: 60}
}
func readLifeBarTime(is IniSection,
	sff *Sff, at AnimationTable) *LifeBarTime {
	t := newLifeBarTime()
	is.ReadI32("pos", &t.pos[0], &t.pos[1])
	is.ReadI32("counter.font", &t.counter_font[0], &t.counter_font[1],
		&t.counter_font[2])
	t.counter_lay = *ReadLayout("counter.", is, 0)
	t.bg = *ReadAnimLayout("bg.", is, sff, at, 0)
	is.ReadI32("framespercount", &t.framespercount)
	return t
}
func (t *LifeBarTime) step()  { t.bg.Action() }
func (t *LifeBarTime) reset() { t.bg.Reset() }
func (t *LifeBarTime) bgDraw(layerno int16) {
	t.bg.DrawScaled(float32(t.pos[0])+sys.lifebarOffsetX, float32(t.pos[1]), layerno, sys.lifebarScale)
}
func (t *LifeBarTime) draw(layerno int16, f []*Fnt) {
	if t.framespercount > 0 &&
		t.counter_font[0] >= 0 && int(t.counter_font[0]) < len(f) {
		time := "o"
		if sys.time >= 0 {
			time = fmt.Sprintf("%v", sys.time/t.framespercount)
		}
		t.counter_lay.DrawText(float32(t.pos[0])+sys.lifebarOffsetX, float32(t.pos[1]), sys.lifebarScale, layerno,
			time, f[t.counter_font[0]], t.counter_font[1], t.counter_font[2])
	}
}

type LifeBarCombo struct {
	pos           [2][2]int32
	start_x       [2]float32
	counter_font  [2][3]int32
	counter_shake [2]bool
	counter_lay   [2]Layout
	text_font     [2][3]int32
	text_text     [2]string
	text_lay      [2]Layout
	displaytime   [2]int32
	cur, old      [2]int32
	curd, oldd    [2]int32
	resttime      [2]int32
	counterX      [2]float32
	shaketime     [2]int32
	teamMode      bool
	firstAttack   int
	counterHits   [2]int
	separator     string
	places        int32
}

func newLifeBarCombo() *LifeBarCombo {
	return &LifeBarCombo{counter_font: [2][3]int32{{-1}, {-1}}, text_font: [2][3]int32{{-1}, {-1}},
		displaytime: [2]int32{90}}
}
func readLifeBarCombo(is IniSection) *LifeBarCombo {
	c := newLifeBarCombo()
	c.teamMode = false
	for i := 0; i < 2; i++ {
		is.ReadI32("pos", &c.pos[i][0], &c.pos[i][1])
		is.ReadF32("start.x", &c.start_x[i])
		is.ReadI32("counter.font", &c.counter_font[i][0], &c.counter_font[i][1],
			&c.counter_font[i][2])
		is.ReadBool("counter.shake", &c.counter_shake[i])
		c.counter_lay[i] = *ReadLayout("counter.", is, 2)
		c.counter_lay[i].offset = [2]float32{}
		is.ReadI32("text.font", &c.text_font[i][0], &c.text_font[i][1], &c.text_font[i][2])
		c.text_text[i], _ = is.getString("text.text")
		c.text_lay[i] = *ReadLayout("text.", is, 2)
		is.ReadI32("displaytime", &c.displaytime[i])
	}

	//Load team 1
	is.ReadI32("team1.pos", &c.pos[0][0], &c.pos[0][1])
	is.ReadF32("team1.start.x", &c.start_x[0])
	is.ReadI32("team1.counter.font", &c.counter_font[0][0], &c.counter_font[0][1],
		&c.counter_font[0][2])
	is.ReadBool("team1.counter.shake", &c.counter_shake[0])
	c.counter_lay[0] = *ReadLayout("team1.counter.", is, 2)
	c.counter_lay[0].offset = [2]float32{}
	is.ReadI32("team1.text.font", &c.text_font[0][0], &c.text_font[0][1], &c.text_font[0][2])
	if len(is["team1.text.text"]) > 0 {
		c.text_text[0], _ = is.getString("team1.text.text")
	}
	c.text_lay[0] = *ReadLayout("team1.text.", is, 2)
	is.ReadI32("team1.displaytime", &c.displaytime[0])

	//Load team 2
	if is.ReadI32("team2.pos", &c.pos[1][0], &c.pos[1][1]) == true {
		c.teamMode = true
	}
	is.ReadF32("team2.start.x", &c.start_x[1])
	is.ReadI32("team2.counter.font", &c.counter_font[1][0], &c.counter_font[1][1],
		&c.counter_font[1][2])
	is.ReadBool("team2.counter.shake", &c.counter_shake[1])
	c.counter_lay[1] = *ReadLayout("team2.counter.", is, 2)
	c.counter_lay[1].offset = [2]float32{}
	is.ReadI32("team2.text.font", &c.text_font[1][0], &c.text_font[1][1], &c.text_font[1][2])
	if len(is["team2.text.text"]) > 0 {
		c.text_text[1], _ = is.getString("team2.text.text")
	}
	c.text_lay[1] = *ReadLayout("team2.text.", is, 2)
	is.ReadI32("team2.displaytime", &c.displaytime[1])

	c.separator, _ = is.getString("format.decimal.separator")
	is.ReadI32("format.decimal.places", &c.places)
	return c
}

func (c *LifeBarCombo) step(combo [2]int32, damage [2]int32) {
	for i := range c.cur {
		if c.resttime[i] > 0 {
			c.counterX[i] -= c.counterX[i] / 8
		} else {
			c.counterX[i] -= sys.lifebarFontScale * 4
			if c.counterX[i] < c.start_x[i]*2 {
				c.counterX[i] = c.start_x[i] * 2
			}
		}
		if c.shaketime[i] > 0 {
			c.shaketime[i]--
		}
		if AbsF(c.counterX[i]) < 1 {
			c.resttime[i]--
		}
		if combo[i] >= 2 {
			if c.old[i] != combo[i] {
				c.cur[i] = combo[i]
				c.resttime[i] = c.displaytime[i]
				if c.counter_shake[i] {
					c.shaketime[i] = 15
				}
			}
			if c.oldd[i] != damage[i] {
				c.curd[i] = damage[i]
			}
		}
		c.old[i] = combo[i]
		c.oldd[i] = damage[i]
	}
}

func (c *LifeBarCombo) reset() {
	c.cur, c.old, c.curd, c.oldd, c.resttime = [2]int32{}, [2]int32{}, [2]int32{}, [2]int32{}, [2]int32{}
	c.counterX = [...]float32{c.start_x[0] * 2, c.start_x[1] * 2}
	c.shaketime = [2]int32{}
	c.firstAttack = -1
	c.counterHits = [2]int{}
}

func (c *LifeBarCombo) draw(layerno int16, f []*Fnt) {
	for i := range c.cur {
		haba := func(n int32) float32 {
			if f[c.counter_font[i][0]] == nil || c.counter_font[i][0] < 0 || int(c.counter_font[i][0]) >= len(f) {
				return 0
			}
			return float32(f[c.counter_font[i][0]].TextWidth(fmt.Sprintf("%v", n))) *
				c.text_lay[i].scale[0]
		}
		if c.counter_font[i][0] < 0 || c.resttime[i] <= 0 && c.counterX[i] == c.start_x[i]*2 {
			continue
		}
		var x float32
		if i&1 == 0 {
			if c.start_x[i] <= 0 {
				x = c.counterX[i]
			}
			x += float32(c.pos[i][0]) + haba(c.cur[i])
		} else {
			if c.start_x[i] <= 0 {
				x = -c.counterX[i]
			}
			if c.teamMode == false {
				x += 320/sys.lifebarScale - sys.lifebarOffsetX*2 - float32(c.pos[i][0])
			} else {
				x += float32(c.pos[i][0])
			}
		}
		if c.text_font[i][0] >= 0 && int(c.text_font[i][0]) < len(f) {
			//text := OldSprintf(c.text_text[i], c.cur[i])
			text := strings.Replace(c.text_text[i], "%i", fmt.Sprintf("%d", c.cur[i]), 1)
			text = strings.Replace(text, "%d", fmt.Sprintf("%d", c.curd[i]), 1)
			//split float value, round to decimal place
			s := strings.Split(fmt.Sprintf("%s", fmt.Sprintf("%.[2]*[1]f", float32(c.curd[i])/float32(sys.chars[^i&1][0].lifeMax)*100, c.places)), ".")
			//decimal separator
			if c.places > 0 {
				if len(s) > 1 {
					s[0] = s[0] + c.separator + s[1]
				}
			}
			//replace %p with formatted string
			text = strings.Replace(text, "%p", s[0], 1)
			
			if i&1 == 0 {
				if c.pos[i][0] != 0 {
					x += c.text_lay[i].offset[0] *
						((1 - sys.lifebarFontScale) * sys.lifebarFontScale)
				}
			} else {
				tmp := c.text_lay[i].offset[0]
				if c.pos[i][0] == 0 {
					tmp *= sys.lifebarFontScale
				}
				x -= tmp + float32(f[c.text_font[i][0]].TextWidth(text))*
					c.text_lay[i].scale[0]*sys.lifebarFontScale
			}
			c.text_lay[i].DrawText(x+sys.lifebarOffsetX, float32(c.pos[i][1]), sys.lifebarScale, layerno,
				text, f[c.text_font[i][0]], c.text_font[i][1], 1)
		}
		if c.counter_font[i][0] >= 0 && f[c.counter_font[i][0]] != nil && int(c.counter_font[i][0]) < len(f) {
			z := 1 + float32(c.shaketime[i])*(1.0/20)*
				float32(math.Sin(float64(c.shaketime[i])*(math.Pi/2.5)))
			c.counter_lay[i].DrawText((x+sys.lifebarOffsetX)/z, float32(c.pos[i][1])/z, z*sys.lifebarScale, layerno,
				fmt.Sprintf("%v", c.cur[i]), f[c.counter_font[i][0]],
				c.counter_font[i][1], -1)
		}
	}
}

type LifeBarRound struct {
	snd                *Snd
	pos                [2]int32
	match_wins         int32
	match_maxdrawgames int32
	start_waittime     int32
	round_time         int32
	round_sndtime      int32
	round_default      AnimTextSnd
	round              [9]AnimTextSnd
	round_final        AnimTextSnd
	fight_time         int32
	fight_sndtime      int32
	fight              AnimTextSnd
	ctrl_time          int32
	ko_time            int32
	ko_sndtime         int32
	ko, dko, to        AnimTextSnd
	slow_time          int32
	over_waittime      int32
	over_hittime       int32
	over_wintime       int32
	over_time          int32
	win_time           int32
	win_sndtime        int32
	win, win2, drawn   AnimTextSnd
	cur                int32
	wt, swt, dt        [2]int32
	fnt                []*Fnt
	timerActive        bool
	wint               [WT_NumTypes*2]BgTextSnd
}

func newLifeBarRound(snd *Snd, fnt []*Fnt) *LifeBarRound {
	return &LifeBarRound{snd: snd, match_wins: 2, match_maxdrawgames: 1,
		start_waittime: 30, ctrl_time: 30, slow_time: 60, over_waittime: 45,
		over_hittime: 10, over_wintime: 45, over_time: 210, win_sndtime: 60,
		fnt: fnt}
}
func readLifeBarRound(is IniSection,
	sff *Sff, at AnimationTable, snd *Snd, fnt []*Fnt) *LifeBarRound {
	r := newLifeBarRound(snd, fnt)
	var tmp int32
	is.ReadI32("pos", &r.pos[0], &r.pos[1])
	is.ReadI32("match.wins", &r.match_wins)
	is.ReadI32("match.maxdrawgames", &r.match_maxdrawgames)
	if is.ReadI32("start.waittime", &tmp) {
		r.start_waittime = Max(1, tmp)
	}
	is.ReadI32("round.time", &r.round_time)
	is.ReadI32("round.sndtime", &r.round_sndtime)
	r.round_default = *ReadAnimTextSnd("round.default.", is, sff, at, 2)
	for i := range r.round {
		r.round[i] = *ReadAnimTextSnd(fmt.Sprintf("round%v.", i+1), is, sff, at, 2)
	}
	r.round_final = *ReadAnimTextSnd("round.final.", is, sff, at, 2)
	is.ReadI32("fight.time", &r.fight_time)
	is.ReadI32("fight.sndtime", &r.fight_sndtime)
	r.fight = *ReadAnimTextSnd("fight.", is, sff, at, 2)
	if is.ReadI32("ctrl.time", &tmp) {
		r.ctrl_time = Max(1, tmp)
	}
	is.ReadI32("ko.time", &r.ko_time)
	is.ReadI32("ko.sndtime", &r.ko_sndtime)
	r.ko = *ReadAnimTextSnd("ko.", is, sff, at, 1)
	r.dko = *ReadAnimTextSnd("dko.", is, sff, at, 1)
	r.to = *ReadAnimTextSnd("to.", is, sff, at, 1)
	is.ReadI32("slow.time", &r.slow_time)
	if is.ReadI32("over.hittime", &tmp) {
		r.over_hittime = Max(0, tmp)
	}
	if is.ReadI32("over.waittime", &tmp) {
		r.over_waittime = Max(1, tmp)
	}
	if is.ReadI32("over.wintime", &tmp) {
		r.over_wintime = Max(1, tmp)
	}
	if is.ReadI32("over.time", &tmp) {
		r.over_time = Max(r.over_wintime+1, tmp)
	}
	is.ReadI32("win.time", &r.win_time)
	is.ReadI32("win.sndtime", &r.win_sndtime)
	r.win = *ReadAnimTextSnd("win.", is, sff, at, 1)
	r.win2 = *ReadAnimTextSnd("win2.", is, sff, at, 1)
	r.drawn = *ReadAnimTextSnd("draw.", is, sff, at, 1)
	r.wint[WT_N] = readBgTextSnd("p1.n.", is, sff, at)
	r.wint[WT_S] = readBgTextSnd("p1.s.", is, sff, at)
	r.wint[WT_H] = readBgTextSnd("p1.h.", is, sff, at)
	r.wint[WT_C] = readBgTextSnd("p1.c.", is, sff, at)
	r.wint[WT_T] = readBgTextSnd("p1.t.", is, sff, at)
	r.wint[WT_Throw] = readBgTextSnd("p1.throw.", is, sff, at)
	r.wint[WT_Suicide] = readBgTextSnd("p1.suicide.", is, sff, at)
	r.wint[WT_Teammate] = readBgTextSnd("p1.teammate.", is, sff, at)
	r.wint[WT_Perfect] = readBgTextSnd("p1.perfect.", is, sff, at)
	r.wint[WT_N+WT_NumTypes] = readBgTextSnd("p2.n.", is, sff, at)
	r.wint[WT_S+WT_NumTypes] = readBgTextSnd("p2.s.", is, sff, at)
	r.wint[WT_H+WT_NumTypes] = readBgTextSnd("p2.h.", is, sff, at)
	r.wint[WT_C+WT_NumTypes] = readBgTextSnd("p2.c.", is, sff, at)
	r.wint[WT_T+WT_NumTypes] = readBgTextSnd("p2.t.", is, sff, at)
	r.wint[WT_Throw+WT_NumTypes] = readBgTextSnd("p2.throw.", is, sff, at)
	r.wint[WT_Suicide+WT_NumTypes] = readBgTextSnd("p2.suicide.", is, sff, at)
	r.wint[WT_Teammate+WT_NumTypes] = readBgTextSnd("p2.teammate.", is, sff, at)
	r.wint[WT_Perfect+WT_NumTypes] = readBgTextSnd("p2.perfect.", is, sff, at)
	return r
}
func (r *LifeBarRound) callFight() {
	r.fight.Reset()
	r.cur, r.wt[0], r.swt[0], r.dt[0] = 1, r.fight_time, r.fight_sndtime, 0
	sys.timerCount = append(sys.timerCount, sys.gameTime)
	r.timerActive = true
}
func (r *LifeBarRound) act() bool {
	if sys.intro > r.ctrl_time {
		r.cur, r.wt[0], r.swt[0], r.dt[0] = 0, r.round_time, r.round_sndtime, 0
	} else if sys.intro >= 0 || r.cur < 2 {
		if !sys.tickNextFrame() {
			return false
		}
		switch r.cur {
		case 0:
			if r.swt[0] == 0 {
				if sys.roundType[0] == RT_Final && r.round_final.snd[0] != -1 {
					r.snd.play(r.round_final.snd)
				} else if int(sys.round) <= len(r.round) && r.round[sys.round-1].snd[0] != -1 {
					r.snd.play(r.round[sys.round-1].snd)
				} else {
					r.snd.play(r.round_default.snd)
				}
			}
			r.swt[0]--
			if r.wt[0] <= 0 {
				r.dt[0]++
				end := false
				if sys.roundType[0] == RT_Final && r.round_final.snd[0] != -1 {
					r.round_final.Action()
					r.round_default.Action()
					end = r.round_final.End(r.dt[0]) && r.round_default.End(r.dt[0])
				} else if int(sys.round) <= len(r.round) {
					r.round[sys.round-1].Action()
					r.round_default.Action()
					end = r.round[sys.round-1].End(r.dt[0]) && r.round_default.End(r.dt[0])
				} else {
					r.round_default.Action()
					end = r.round_default.End(r.dt[0])
				}
				if end {
					r.callFight()
					return true
				}
			}
			r.wt[0]--
			return false
		case 1:
			if r.swt[0] == 0 {
				r.snd.play(r.fight.snd)
			}
			r.swt[0]--
			if r.wt[0] <= 0 {
				r.dt[0]++
				r.fight.Action()
				if r.fight.End(r.dt[0]) {
					r.cur, r.wt[0], r.swt[0], r.dt[0] = 2, r.ko_time, r.ko_sndtime, 0
					r.wt[1], r.swt[1], r.dt[1] = r.win_time, r.win_sndtime, 0
					break
				}
			}
			r.wt[0]--
		}
	} else if r.cur == 2 && (sys.intro < 0) && (sys.finish != FT_NotYet || sys.time == 0) {
		if r.timerActive {
			if sys.gameTime-sys.timerCount[sys.round-1] > 0 {
				sys.timerCount[sys.round-1] = sys.gameTime - sys.timerCount[sys.round-1]
				sys.timerRounds = append(sys.timerRounds, sys.roundTime-sys.time)
			} else {
				sys.timerCount[sys.round-1] = 0
			}
			r.timerActive = false
		}
		f := func(ats *AnimTextSnd, t int) {
			if -r.swt[t]-10 == 0 {
				r.snd.play(ats.snd)
				r.swt[t]--
			}
			if sys.tickNextFrame() {
				r.swt[t]--
			}
			if ats.End(r.dt[t]) {
				r.wt[t] = 2
			}
			if sys.intro < -r.ko_time-10 {
				r.dt[t]++
				ats.Action()
			}
			r.wt[t]--
		}
		switch sys.finish {
		case FT_KO:
			f(&r.ko, 0)
		case FT_DKO:
			f(&r.dko, 0)
		default:
			f(&r.to, 0)
		}
		if sys.intro < -(r.over_hittime + r.over_waittime + r.over_wintime) {
			if /*sys.finish == FT_DKO ||*/ sys.finish == FT_TODraw {
				f(&r.drawn, 1)
			} else if sys.winTeam >= 0 && (sys.tmode[sys.winTeam] == TM_Simul || sys.tmode[sys.winTeam] == TM_Tag) {
				f(&r.win2, 1)
			} else {
				f(&r.win, 1)
			}
		}
	}
	if sys.winTeam >= 0 {
		index := sys.winType[sys.winTeam]
		if index > WT_NumTypes {
			if sys.winTeam == 0 {
				r.wint[WT_Perfect].step(r.snd)
				index = index - WT_NumTypes - 1
			} else {
				r.wint[WT_Perfect+WT_NumTypes].step(r.snd)
				index = index - 1
			}
		}
		r.wint[index].step(r.snd)
	}
	return sys.tickNextFrame()
}
func (r *LifeBarRound) reset() {
	r.round_default.Reset()
	for i := range r.round {
		r.round[i].Reset()
	}
	r.round_final.Reset()
	r.fight.Reset()
	r.ko.Reset()
	r.dko.Reset()
	r.to.Reset()
	r.win.Reset()
	r.win2.Reset()
	r.drawn.Reset()
	for i := range r.wint {
		r.wint[i].reset()
	}
}
func (r *LifeBarRound) draw(layerno int16) {
	ob := sys.brightness
	sys.brightness = 255
	switch r.cur {
	case 0:
		if r.wt[0] < 0 && sys.intro <= r.ctrl_time {
			tmp := r.round_default.text
			r.round_default.text = OldSprintf(tmp, sys.round)
			r.round_default.DrawScaled(float32(r.pos[0])+sys.lifebarOffsetX, float32(r.pos[1]),
				layerno, r.fnt, sys.lifebarScale)
			r.round_default.text = tmp
			if sys.roundType[0] == RT_Final && (r.round_final.font[0] != -1 ||
				len(r.round_final.anim.anim.frames) > 0) {
				tmp = r.round_final.text
				r.round_final.text = OldSprintf(tmp, sys.round)
				r.round_final.DrawScaled(float32(r.pos[0])+sys.lifebarOffsetX, float32(r.pos[1]),
					layerno, r.fnt, sys.lifebarScale)
				r.round_final.text = tmp
			} else if int(sys.round) <= len(r.round) {
				tmp = r.round[sys.round-1].text
				r.round[sys.round-1].text = OldSprintf(tmp, sys.round)
				r.round[sys.round-1].DrawScaled(float32(r.pos[0])+sys.lifebarOffsetX, float32(r.pos[1]),
					layerno, r.fnt, sys.lifebarScale)
				r.round[sys.round-1].text = tmp
			}
		}
	case 1:
		if r.wt[0] < 0 {
			r.fight.DrawScaled(float32(r.pos[0])+sys.lifebarOffsetX, float32(r.pos[1]), layerno, r.fnt, sys.lifebarScale)
		}
	case 2:
		if r.wt[0] < 0 && sys.intro < -r.ko_time-10 {
			switch sys.finish {
			case FT_KO:
				r.ko.DrawScaled(float32(r.pos[0])+sys.lifebarOffsetX, float32(r.pos[1]), layerno, r.fnt, sys.lifebarScale)
			case FT_DKO:
				r.dko.DrawScaled(float32(r.pos[0])+sys.lifebarOffsetX, float32(r.pos[1]), layerno, r.fnt, sys.lifebarScale)
			default:
				r.to.DrawScaled(float32(r.pos[0])+sys.lifebarOffsetX, float32(r.pos[1]), layerno, r.fnt, sys.lifebarScale)
			}
		}
		if r.wt[1] < 0 {
			if /*sys.finish == FT_DKO ||*/ sys.finish == FT_TODraw {
				r.drawn.DrawScaled(float32(r.pos[0])+sys.lifebarOffsetX, float32(r.pos[1]), layerno, r.fnt, sys.lifebarScale)
			} else if sys.winTeam >= 0 && (sys.tmode[sys.winTeam] == TM_Simul || sys.tmode[sys.winTeam] == TM_Tag) {
				tmp := r.win2.text
				var inter []interface{}
				for i := sys.winTeam; i < len(sys.chars); i += 2 {
					if len(sys.chars[i]) > 0 {
						inter = append(inter, sys.cgi[i].displayname)
					}
				}
				r.win2.text = OldSprintf(tmp, inter...)
				r.win2.DrawScaled(float32(r.pos[0])+sys.lifebarOffsetX, float32(r.pos[1]), layerno, r.fnt, sys.lifebarScale)
				r.win2.text = tmp
			} else if sys.winTeam >= 0 {
				tmp := r.win.text
				r.win.text = OldSprintf(tmp, sys.cgi[sys.winTeam].displayname)
				r.win.DrawScaled(float32(r.pos[0])+sys.lifebarOffsetX, float32(r.pos[1]), layerno, r.fnt, sys.lifebarScale)
				r.win.text = tmp
			}
		}
	}
	if sys.winTeam >= 0 {
		index := sys.winType[sys.winTeam]
		perfect := false
		if index > WT_NumTypes {
			if sys.winTeam == 0 {
				index = index - WT_NumTypes - 1
			} else {
				index = index - 1
			}
			perfect = true
		}
		if perfect {
			if sys.winTeam == 0 {
				r.wint[WT_Perfect].bgDraw(layerno)
				r.wint[WT_Perfect].draw(layerno, r.fnt)
			} else {
				r.wint[WT_Perfect+WT_NumTypes].bgDraw(layerno)
				r.wint[WT_Perfect+WT_NumTypes].draw(layerno, r.fnt)
			}
		}
		r.wint[index].bgDraw(layerno)
		r.wint[index].draw(layerno, r.fnt)
	}
	sys.brightness = ob
}

type LifeBarChallenger struct {
	challenger  BgTextSnd
	over_pause  int32
	over_time   int32
}

func newLifeBarChallenger() *LifeBarChallenger {
	return &LifeBarChallenger{}
}
func readLifeBarChallenger(is IniSection,
	sff *Sff, at AnimationTable) *LifeBarChallenger {
	ch := newLifeBarChallenger()
	ch.challenger = readBgTextSnd("", is, sff, at)
	var tmp int32
	if is.ReadI32("over.pause", &tmp) {
		ch.over_pause = Max(1, tmp)
	}
	if is.ReadI32("over.time", &tmp) {
		ch.over_time = Max(ch.over_pause+1, tmp)
	}
	return ch
}
func (ch *LifeBarChallenger) step(snd *Snd)  {
	if sys.challenger > 0 {
		ch.challenger.step(snd)
		if ch.challenger.cnt == ch.over_pause {
			sys.paused = true
		}
	}
}
func (ch *LifeBarChallenger) reset() {
	ch.challenger.reset()
}
func (ch *LifeBarChallenger) bgDraw(layerno int16) {
	if sys.challenger > 0 {
		ch.challenger.bgDraw(layerno)
	}
}
func (ch *LifeBarChallenger) draw(layerno int16, f []*Fnt) {
	if sys.challenger > 0 {
		ch.challenger.draw(layerno, f)
	}
}

type LifeBarRatio struct {
	pos    [2]int32
	icon   [4]AnimLayout
}

func newLifeBarRatio() *LifeBarRatio {
	return &LifeBarRatio{}
}
func readLifeBarRatio(pre string, is IniSection,
	sff *Sff, at AnimationTable) *LifeBarRatio {
	ra := newLifeBarRatio()
	is.ReadI32(pre+"pos", &ra.pos[0], &ra.pos[1])
	ra.icon[0] = *ReadAnimLayout(pre+"level1.", is, sff, at, 0)
	ra.icon[1] = *ReadAnimLayout(pre+"level2.", is, sff, at, 0)
	ra.icon[2] = *ReadAnimLayout(pre+"level3.", is, sff, at, 0)
	ra.icon[3] = *ReadAnimLayout(pre+"level4.", is, sff, at, 0)
	return ra
}
func (ra *LifeBarRatio) step(num int32) {
	ra.icon[num].Action()
}
func (ra *LifeBarRatio) reset() {
	for i := range ra.icon {
		ra.icon[i].Reset()
	}
}
func (ra *LifeBarRatio) draw(layerno int16, num int32) {
	ra.icon[num].DrawScaled(float32(ra.pos[0])+sys.lifebarOffsetX,
		float32(ra.pos[1]), layerno, sys.lifebarScale)
}

type LifeBarTimer struct {
	pos       [2]int32
	text_font [3]int32
	text_text string
	text_lay  Layout
	bg        AnimLayout
	active    bool
}

func newLifeBarTimer() *LifeBarTimer {
	return &LifeBarTimer{text_font: [3]int32{-1}}
}
func readLifeBarTimer(is IniSection,
	sff *Sff, at AnimationTable) *LifeBarTimer {
	tr := newLifeBarTimer()
	is.ReadI32("pos", &tr.pos[0], &tr.pos[1])
	is.ReadI32("text.font", &tr.text_font[0], &tr.text_font[1],
		&tr.text_font[2])
	tr.text_text, _ = is.getString("text.text")
	tr.text_lay = *ReadLayout("text.", is, 0)
	tr.bg = *ReadAnimLayout("bg.", is, sff, at, 0)
	return tr
}
func (tr *LifeBarTimer) step()  { tr.bg.Action() }
func (tr *LifeBarTimer) reset() { tr.bg.Reset() }
func (tr *LifeBarTimer) bgDraw(layerno int16) {
	if tr.active {
		tr.bg.DrawScaled(float32(tr.pos[0])+sys.lifebarOffsetX, float32(tr.pos[1]), layerno, sys.lifebarScale)
	}
}
func (tr *LifeBarTimer) draw(layerno int16, f []*Fnt) {
	if tr.active && sys.lifebar.ti.framespercount > 0 &&
		tr.text_font[0] >= 0 && int(tr.text_font[0]) < len(f) && sys.time >= 0 {
		text := tr.text_text
		totalSec := float64(timeTotal()) / 60
		h := math.Floor(totalSec / 3600)
		m := math.Floor((totalSec / 3600 - h) * 60)
		s := math.Floor(((totalSec / 3600 - h) * 60 - m) * 60)
		x := math.Floor((((totalSec / 3600 - h) * 60 - m) * 60 - s) * 100)
		ms, ss, xs := fmt.Sprintf("%.0f", m), fmt.Sprintf("%.0f", s), fmt.Sprintf("%.0f", x)
		if len(ms) < 2 {
			ms = "0" + ms
		}
		if len(ss) < 2 {
			ss = "0" + ss
		}
		if len(xs) < 2 {
			xs = "0" + xs
		}
		text = strings.Replace(text, "%m", ms, 1)
		text = strings.Replace(text, "%s", ss, 1)
		text = strings.Replace(text, "%x", xs, 1)
		tr.text_lay.DrawText(float32(tr.pos[0])+sys.lifebarOffsetX, float32(tr.pos[1]), sys.lifebarScale, layerno,
			text, f[tr.text_font[0]], tr.text_font[1], tr.text_font[2])
	}
}

type LifeBarScore struct {
	pos       [2]int32
	text_font [3]int32
	text_text string
	text_lay  Layout
	bg        AnimLayout
	separator [2]string
	pad       int32
	places    int32
	min       float32
	max       float32
	active    bool
}

func newLifeBarScore() *LifeBarScore {
	return &LifeBarScore{text_font: [3]int32{-1}, separator: [2]string{"", "."}}
}
func readLifeBarScore(pre string, is IniSection,
	sff *Sff, at AnimationTable) *LifeBarScore {
	sc := newLifeBarScore()
	is.ReadI32(pre+"pos", &sc.pos[0], &sc.pos[1])
	is.ReadI32(pre+"text.font", &sc.text_font[0], &sc.text_font[1],
		&sc.text_font[2])
	sc.text_text, _ = is.getString(pre+"text.text")
	sc.text_lay = *ReadLayout(pre+"text.", is, 0)
	sc.separator[0], _ = is.getString("format.integer.separator")
	sc.separator[1], _ = is.getString("format.decimal.separator")
	is.ReadI32("format.integer.pad", &sc.pad)
	is.ReadI32("format.decimal.places", &sc.places)
	is.ReadF32("score.min", &sc.min)
	is.ReadF32("score.max", &sc.max)
	sc.bg = *ReadAnimLayout(pre+"bg.", is, sff, at, 0)
	return sc
}
func (sc *LifeBarScore) step() { sc.bg.Action() }
func (sc *LifeBarScore) reset() { sc.bg.Reset() }
func (sc *LifeBarScore) bgDraw(layerno int16) {
	if sc.active {
		sc.bg.DrawScaled(float32(sc.pos[0])+sys.lifebarOffsetX, float32(sc.pos[1]), layerno, sys.lifebarScale)
	}
}
func (sc *LifeBarScore) draw(layerno int16, f []*Fnt, side int) {
	if sc.active && sc.text_font[0] >= 0 && int(sc.text_font[0]) < len(f) {
		text := sc.text_text
		total := scoreTotal(side)
		if total == 0 && sc.pad == 0 {
			return
		}
		if total > sc.max {
			total = sc.max
		} else if total < sc.min {
			total = sc.min
		}
		//split float value
		s := strings.Split(fmt.Sprintf("%f", total), ".")
		//integer left padding (add leading zeros)
		for i := int(sc.pad)-len(s[0]); i > 0; i-- {
			s[0] = "0" + s[0]
		}
		//integer thousands separator
		for i := len(s[0]) - 3; i > 0; i -= 3 {
			s[0] = s[0][:i] + sc.separator[0] + s[0][i:]
		}
		//decimal places (trim trailing numbers)
		if int(sc.places) < len(s[1]) {
			s[1] = s[1][:sc.places]
		}
		//decimal separator
		ds := ""
		if sc.places > 0 {
			ds = sc.separator[1]
		}
		//replace %s with formatted string
		text = strings.Replace(text, "%s", s[0]+ds+s[1], 1)
		sc.text_lay.DrawText(float32(sc.pos[0])+sys.lifebarOffsetX, float32(sc.pos[1]), sys.lifebarScale, layerno,
			text, f[sc.text_font[0]], sc.text_font[1], sc.text_font[2])
	}
}

type LifeBarMatch struct {
	pos       [2]int32
	text_font [3]int32
	text_text string
	text_lay  Layout
	bg        AnimLayout
	active    bool
}

func newLifeBarMatch() *LifeBarMatch {
	return &LifeBarMatch{text_font: [3]int32{-1}}
}
func readLifeBarMatch(is IniSection,
	sff *Sff, at AnimationTable) *LifeBarMatch {
	ma := newLifeBarMatch()
	is.ReadI32("pos", &ma.pos[0], &ma.pos[1])
	is.ReadI32("text.font", &ma.text_font[0], &ma.text_font[1],
		&ma.text_font[2])
	ma.text_text, _ = is.getString("text.text")
	ma.text_lay = *ReadLayout("text.", is, 0)
	ma.bg = *ReadAnimLayout("bg.", is, sff, at, 0)
	return ma
}
func (ma *LifeBarMatch) step()  { ma.bg.Action() }
func (ma *LifeBarMatch) reset() { ma.bg.Reset() }
func (ma *LifeBarMatch) bgDraw(layerno int16) {
	if ma.active {
		ma.bg.DrawScaled(float32(ma.pos[0])+sys.lifebarOffsetX, float32(ma.pos[1]), layerno, sys.lifebarScale)
	}
}
func (ma *LifeBarMatch) draw(layerno int16, f []*Fnt) {
	if ma.active && ma.text_font[0] >= 0 && int(ma.text_font[0]) < len(f) {
		text := ma.text_text
		text = strings.Replace(text, "%s", fmt.Sprintf("%v", sys.match), 1)
		ma.text_lay.DrawText(float32(ma.pos[0])+sys.lifebarOffsetX, float32(ma.pos[1]), sys.lifebarScale, layerno,
			text, f[ma.text_font[0]], ma.text_font[1], ma.text_font[2])
	}
}

type LifeBarAiLevel struct {
	pos       [2]int32
	text_font [3]int32
	text_text string
	text_lay  Layout
	bg        AnimLayout
	active    bool
}

func newLifeBarAiLevel() *LifeBarAiLevel {
	return &LifeBarAiLevel{text_font: [3]int32{-1}}
}
func readLifeBarAiLevel(pre string, is IniSection,
	sff *Sff, at AnimationTable) *LifeBarAiLevel {
	ai := newLifeBarAiLevel()
	is.ReadI32(pre+"pos", &ai.pos[0], &ai.pos[1])
	is.ReadI32(pre+"text.font", &ai.text_font[0], &ai.text_font[1],
		&ai.text_font[2])
	ai.text_text, _ = is.getString(pre+"text.text")
	ai.text_lay = *ReadLayout(pre+"text.", is, 0)
	ai.bg = *ReadAnimLayout(pre+"bg.", is, sff, at, 0)
	return ai
}
func (ai *LifeBarAiLevel) step() { ai.bg.Action() }
func (ai *LifeBarAiLevel) reset() { ai.bg.Reset() }
func (ai *LifeBarAiLevel) bgDraw(layerno int16) {
	if ai.active {
		ai.bg.DrawScaled(float32(ai.pos[0])+sys.lifebarOffsetX, float32(ai.pos[1]), layerno, sys.lifebarScale)
	}
}
func (ai *LifeBarAiLevel) draw(layerno int16, f []*Fnt, ailv float32) {
	if ai.active && ailv > 0 && ai.text_font[0] >= 0 && int(ai.text_font[0]) < len(f) {
		text := ai.text_text
		p := ailv / 8 * 100
		text = strings.Replace(text, "%s", fmt.Sprintf("%.0f", ailv), 1)
		text = strings.Replace(text, "%p", fmt.Sprintf("%.0f", p), 1)
		ai.text_lay.DrawText(float32(ai.pos[0])+sys.lifebarOffsetX, float32(ai.pos[1]), sys.lifebarScale, layerno,
			text, f[ai.text_font[0]], ai.text_font[1], ai.text_font[2])
	}
}

type LifeBarMode struct {
	pos       [2]int32
	text_font [3]int32
	text_text string
	text_lay  Layout
	bg        AnimLayout
}

func newLifeBarMode() *LifeBarMode {
	return &LifeBarMode{text_font: [3]int32{-1}}
}
func readLifeBarMode(is IniSection,
	sff *Sff, at AnimationTable) map[string]*LifeBarMode {
	mo := make(map[string]*LifeBarMode)
	for k, _ := range is {
		sp := strings.Split(k, ".")
		if _, ok := mo[sp[0]]; !ok {
			mo[sp[0]] = newLifeBarMode()
			is.ReadI32(sp[0]+".pos", &mo[sp[0]].pos[0], &mo[sp[0]].pos[1])
			is.ReadI32(sp[0]+".text.font", &mo[sp[0]].text_font[0], &mo[sp[0]].text_font[1],
					&mo[sp[0]].text_font[2])
			mo[sp[0]].text_text, _ = is.getString(sp[0]+".text.text")
			mo[sp[0]].text_lay = *ReadLayout(sp[0]+".text.", is, 0)
			mo[sp[0]].bg = *ReadAnimLayout(sp[0]+".bg.", is, sff, at, 0)
		}
	}
	return mo
}
func (mo *LifeBarMode) step()  { mo.bg.Action() }
func (mo *LifeBarMode) reset() { mo.bg.Reset() }
func (mo *LifeBarMode) bgDraw(layerno int16) {
	if sys.lifebar.activeMode {
		mo.bg.DrawScaled(float32(mo.pos[0])+sys.lifebarOffsetX, float32(mo.pos[1]), layerno, sys.lifebarScale)
	}
}
func (mo *LifeBarMode) draw(layerno int16, f []*Fnt) {
	if sys.lifebar.activeMode && mo.text_font[0] >= 0 && int(mo.text_font[0]) < len(f) {
		mo.text_lay.DrawText(float32(mo.pos[0])+sys.lifebarOffsetX, float32(mo.pos[1]), sys.lifebarScale, layerno,
			mo.text_text, f[mo.text_font[0]], mo.text_font[1], mo.text_font[2])
	}
}

type Lifebar struct {
	fat        AnimationTable
	fsff       *Sff
	snd, fsnd  *Snd
	fnt        [10]*Fnt
	ref        [4][2]int
	num        [4][2]int
	hb         [8][]*HealthBar
	pb         [6][]*PowerBar
	fa         [4][]*LifeBarFace
	nm         [4][]*LifeBarName
	wi         [2]*LifeBarWinIcon
	ti         *LifeBarTime
	co         *LifeBarCombo
	ro         *LifeBarRound
	ch         *LifeBarChallenger
	ra         [2]*LifeBarRatio
	tr         *LifeBarTimer
	sc         [2]*LifeBarScore
	ma         *LifeBarMatch
	ai         [2]*LifeBarAiLevel
	mo         map[string]*LifeBarMode
	active     bool
	activeBars bool
	activeMode bool
}

func loadLifebar(deffile string) (*Lifebar, error) {
	str, err := LoadText(deffile)
	if err != nil {
		return nil, err
	}
	l := &Lifebar{fsff: &Sff{}, snd: &Snd{},
		hb: [...][]*HealthBar{make([]*HealthBar, 2), make([]*HealthBar, 4),
			make([]*HealthBar, 2), make([]*HealthBar, 4), make([]*HealthBar, 6),
			make([]*HealthBar, 8), make([]*HealthBar, 6), make([]*HealthBar, 8)},
		pb: [...][]*PowerBar{make([]*PowerBar, 2), make([]*PowerBar, 4),
			make([]*PowerBar, 2), make([]*PowerBar, 2), make([]*PowerBar, 6),
			make([]*PowerBar, 8)},
		fa: [...][]*LifeBarFace{make([]*LifeBarFace, 2), make([]*LifeBarFace, 8),
			make([]*LifeBarFace, 2), make([]*LifeBarFace, 8)},
		nm: [...][]*LifeBarName{make([]*LifeBarName, 2), make([]*LifeBarName, 8),
			make([]*LifeBarName, 2), make([]*LifeBarName, 8)},
		active: true, activeBars: true, activeMode: true}
	missing := map[string]int{"[simul_3p lifebar]": 3, "[simul_4p lifebar]": 4,
		"[tag lifebar]": 5, "[tag_3p lifebar]": 6, "[tag_4p lifebar]": 7,
		"[simul powerbar]": 1, "[turns powerbar]": 2, "[simul_3p powerbar]": 3,
		"[simul_4p powerbar]": 4, "[tag powerbar]": 5, "[tag face]": -1,
		"[tag name]": -1, "[challenger]": -1, "[ratio]": -1, "[timer]": -1,
		"[score]": -1, "[match]": -1, "[ailevel]": -1, "[mode]": -1}
	strc := strings.ToLower(strings.TrimSpace(str))
	for k, _ := range missing {
		strc = strings.Replace(strc, ";"+k, "", -1)
		if strings.Contains(strc, k) {
			delete(missing, k)
		} else {
			str += "\n" + k
		}
	}
	sff, lines, i := &Sff{}, SplitAndTrim(str, "\n"), 0
	at := ReadAnimationTable(sff, lines, &i)
	i = 0
	filesflg := true
	for i < len(lines) {
		is, name, subname := ReadIniSection(lines, &i)
		switch name {
		case "files":
			if filesflg {
				filesflg = false
				if is.LoadFile("sff", deffile, func(filename string) error {
					s, err := loadSff(filename, false)
					if err != nil {
						return err
					}
					*sff = *s
					return nil
				}); err != nil {
					return nil, err
				}
				if is.LoadFile("snd", deffile, func(filename string) error {
					s, err := LoadSnd(filename)
					if err != nil {
						return err
					}
					*l.snd = *s
					return nil
				}); err != nil {
					return nil, err
				}
				if is.LoadFile("fightfx.sff", deffile, func(filename string) error {
					s, err := loadSff(filename, false)
					if err != nil {
						return err
					}
					*l.fsff = *s
					return nil
				}); err != nil {
					return nil, err
				}
				if is.LoadFile("fightfx.air", deffile, func(filename string) error {
					str, err := LoadText(filename)
					if err != nil {
						return err
					}
					lines, i := SplitAndTrim(str, "\n"), 0
					l.fat = ReadAnimationTable(l.fsff, lines, &i)
					return nil
				}); err != nil {
					return nil, err
				}
				if is.LoadFile("common.snd", deffile, func(filename string) error {
					l.fsnd, err = LoadSnd(filename)
					return err
				}); err != nil {
					return nil, err
				}
				for i := range l.fnt {
					if is.LoadFile(fmt.Sprintf("font%v", i), deffile,
						func(filename string) error {
							//h := int32(0)
							//if len(is[fmt.Sprintf("font%v.height", i)]) > 0 {
							//	h = Atoi(is[fmt.Sprintf("font%v.height", i)])
							//}
							//l.fnt[i], err = loadFnt(filename, h)
							l.fnt[i], err = loadFnt(filename)
							return err
						}); err != nil {
						return nil, err
					}
				}
			}
		case "fonts":
			is.ReadF32("scale", &sys.lifebarFontScale)
		case "lifebar":
			if l.hb[0][0] == nil {
				l.hb[0][0] = readHealthBar("p1.", is, sff, at)
			}
			if l.hb[0][1] == nil {
				l.hb[0][1] = readHealthBar("p2.", is, sff, at)
			}
		case "powerbar":
			if l.pb[0][0] == nil {
				l.pb[0][0] = readPowerBar("p1.", is, sff, at, l.snd)
			}
			if l.pb[0][1] == nil {
				l.pb[0][1] = readPowerBar("p2.", is, sff, at, l.snd)
			}
		case "face":
			if l.fa[0][0] == nil {
				l.fa[0][0] = readLifeBarFace("p1.", is, sff, at)
			}
			if l.fa[0][1] == nil {
				l.fa[0][1] = readLifeBarFace("p2.", is, sff, at)
			}
		case "name":
			if l.nm[0][0] == nil {
				l.nm[0][0] = readLifeBarName("p1.", is, sff, at)
			}
			if l.nm[0][1] == nil {
				l.nm[0][1] = readLifeBarName("p2.", is, sff, at)
			}
		case "simul ":
			subname = strings.ToLower(subname)
			switch {
			case len(subname) >= 7 && subname[:7] == "lifebar":
				if l.hb[1][0] == nil {
					l.hb[1][0] = readHealthBar("p1.", is, sff, at)
				}
				if l.hb[1][1] == nil {
					l.hb[1][1] = readHealthBar("p2.", is, sff, at)
				}
				if l.hb[1][2] == nil {
					l.hb[1][2] = readHealthBar("p3.", is, sff, at)
				}
				if l.hb[1][3] == nil {
					l.hb[1][3] = readHealthBar("p4.", is, sff, at)
				}
			case len(subname) >= 8 && subname[:8] == "powerbar":
				if l.pb[1][0] == nil {
					l.pb[1][0] = readPowerBar("p1.", is, sff, at, l.snd)
				}
				if l.pb[1][1] == nil {
					l.pb[1][1] = readPowerBar("p2.", is, sff, at, l.snd)
				}
				if l.pb[1][2] == nil {
					l.pb[1][2] = readPowerBar("p3.", is, sff, at, l.snd)
				}
				if l.pb[1][3] == nil {
					l.pb[1][3] = readPowerBar("p4.", is, sff, at, l.snd)
				}
			case len(subname) >= 4 && subname[:4] == "face":
				if l.fa[1][0] == nil {
					l.fa[1][0] = readLifeBarFace("p1.", is, sff, at)
				}
				if l.fa[1][1] == nil {
					l.fa[1][1] = readLifeBarFace("p2.", is, sff, at)
				}
				if l.fa[1][2] == nil {
					l.fa[1][2] = readLifeBarFace("p3.", is, sff, at)
				}
				if l.fa[1][3] == nil {
					l.fa[1][3] = readLifeBarFace("p4.", is, sff, at)
				}
				if l.fa[1][4] == nil {
					l.fa[1][4] = readLifeBarFace("p5.", is, sff, at)
				}
				if l.fa[1][5] == nil {
					l.fa[1][5] = readLifeBarFace("p6.", is, sff, at)
				}
				if l.fa[1][6] == nil {
					l.fa[1][6] = readLifeBarFace("p7.", is, sff, at)
				}
				if l.fa[1][7] == nil {
					l.fa[1][7] = readLifeBarFace("p8.", is, sff, at)
				}
			case len(subname) >= 4 && subname[:4] == "name":
				if l.nm[1][0] == nil {
					l.nm[1][0] = readLifeBarName("p1.", is, sff, at)
				}
				if l.nm[1][1] == nil {
					l.nm[1][1] = readLifeBarName("p2.", is, sff, at)
				}
				if l.nm[1][2] == nil {
					l.nm[1][2] = readLifeBarName("p3.", is, sff, at)
				}
				if l.nm[1][3] == nil {
					l.nm[1][3] = readLifeBarName("p4.", is, sff, at)
				}
				if l.nm[1][4] == nil {
					l.nm[1][4] = readLifeBarName("p5.", is, sff, at)
				}
				if l.nm[1][5] == nil {
					l.nm[1][5] = readLifeBarName("p6.", is, sff, at)
				}
				if l.nm[1][6] == nil {
					l.nm[1][6] = readLifeBarName("p7.", is, sff, at)
				}
				if l.nm[1][7] == nil {
					l.nm[1][7] = readLifeBarName("p8.", is, sff, at)
				}
			}
		case "turns ":
			subname = strings.ToLower(subname)
			switch {
			case len(subname) >= 7 && subname[:7] == "lifebar":
				if l.hb[2][0] == nil {
					l.hb[2][0] = readHealthBar("p1.", is, sff, at)
				}
				if l.hb[2][1] == nil {
					l.hb[2][1] = readHealthBar("p2.", is, sff, at)
				}
			case len(subname) >= 8 && subname[:8] == "powerbar":
				if l.pb[2][0] == nil {
					l.pb[2][0] = readPowerBar("p1.", is, sff, at, l.snd)
				}
				if l.pb[2][1] == nil {
					l.pb[2][1] = readPowerBar("p2.", is, sff, at, l.snd)
				}
			case len(subname) >= 4 && subname[:4] == "face":
				if l.fa[2][0] == nil {
					l.fa[2][0] = readLifeBarFace("p1.", is, sff, at)
				}
				if l.fa[2][1] == nil {
					l.fa[2][1] = readLifeBarFace("p2.", is, sff, at)
				}
			case len(subname) >= 4 && subname[:4] == "name":
				if l.nm[2][0] == nil {
					l.nm[2][0] = readLifeBarName("p1.", is, sff, at)
				}
				if l.nm[2][1] == nil {
					l.nm[2][1] = readLifeBarName("p2.", is, sff, at)
				}
			}
		case "tag ":
			subname = strings.ToLower(subname)
			switch {
			case len(subname) >= 7 && subname[:7] == "lifebar":
				if l.hb[3][0] == nil {
					l.hb[3][0] = readHealthBar("p1.", is, sff, at)
				}
				if l.hb[3][1] == nil {
					l.hb[3][1] = readHealthBar("p2.", is, sff, at)
				}
				if l.hb[3][2] == nil {
					l.hb[3][2] = readHealthBar("p3.", is, sff, at)
				}
				if l.hb[3][3] == nil {
					l.hb[3][3] = readHealthBar("p4.", is, sff, at)
				}
			case len(subname) >= 8 && subname[:8] == "powerbar":
				if l.pb[3][0] == nil {
					l.pb[3][0] = readPowerBar("p1.", is, sff, at, l.snd)
				}
				if l.pb[3][1] == nil {
					l.pb[3][1] = readPowerBar("p2.", is, sff, at, l.snd)
				}
			case len(subname) >= 4 && subname[:4] == "face":
				if l.fa[3][0] == nil {
					l.fa[3][0] = readLifeBarFace("p1.", is, sff, at)
				}
				if l.fa[3][1] == nil {
					l.fa[3][1] = readLifeBarFace("p2.", is, sff, at)
				}
				if l.fa[3][2] == nil {
					l.fa[3][2] = readLifeBarFace("p3.", is, sff, at)
				}
				if l.fa[3][3] == nil {
					l.fa[3][3] = readLifeBarFace("p4.", is, sff, at)
				}
				if l.fa[3][4] == nil {
					l.fa[3][4] = readLifeBarFace("p5.", is, sff, at)
				}
				if l.fa[3][5] == nil {
					l.fa[3][5] = readLifeBarFace("p6.", is, sff, at)
				}
				if l.fa[3][6] == nil {
					l.fa[3][6] = readLifeBarFace("p7.", is, sff, at)
				}
				if l.fa[3][7] == nil {
					l.fa[3][7] = readLifeBarFace("p8.", is, sff, at)
				}
			case len(subname) >= 4 && subname[:4] == "name":
				if l.nm[3][0] == nil {
					l.nm[3][0] = readLifeBarName("p1.", is, sff, at)
				}
				if l.nm[3][1] == nil {
					l.nm[3][1] = readLifeBarName("p2.", is, sff, at)
				}
				if l.nm[3][2] == nil {
					l.nm[3][2] = readLifeBarName("p3.", is, sff, at)
				}
				if l.nm[3][3] == nil {
					l.nm[3][3] = readLifeBarName("p4.", is, sff, at)
				}
				if l.nm[3][4] == nil {
					l.nm[3][4] = readLifeBarName("p5.", is, sff, at)
				}
				if l.nm[3][5] == nil {
					l.nm[3][5] = readLifeBarName("p6.", is, sff, at)
				}
				if l.nm[3][6] == nil {
					l.nm[3][6] = readLifeBarName("p7.", is, sff, at)
				}
				if l.nm[3][7] == nil {
					l.nm[3][7] = readLifeBarName("p8.", is, sff, at)
				}
			}
		case "simul_3p ", "simul_4p ", "tag_3p ", "tag_4p ":
			i := 4
			switch name {
			case "simul_4p ":
				i = 5
			case "tag_3p ":
				i = 6
			case "tag_4p ":
				i = 7
			}
			subname = strings.ToLower(subname)
			switch {
			case len(subname) >= 7 && subname[:7] == "lifebar":
				if l.hb[i][0] == nil {
					l.hb[i][0] = readHealthBar("p1.", is, sff, at)
				}
				if l.hb[i][1] == nil {
					l.hb[i][1] = readHealthBar("p2.", is, sff, at)
				}
				if l.hb[i][2] == nil {
					l.hb[i][2] = readHealthBar("p3.", is, sff, at)
				}
				if l.hb[i][3] == nil {
					l.hb[i][3] = readHealthBar("p4.", is, sff, at)
				}
				if l.hb[i][4] == nil {
					l.hb[i][4] = readHealthBar("p5.", is, sff, at)
				}
				if l.hb[i][5] == nil {
					l.hb[i][5] = readHealthBar("p6.", is, sff, at)
				}
				if i == 5 || i == 7 {
					if l.hb[i][6] == nil {
						l.hb[i][6] = readHealthBar("p7.", is, sff, at)
					}
					if l.hb[i][7] == nil {
						l.hb[i][7] = readHealthBar("p8.", is, sff, at)
					}
				}
			case len(subname) >= 8 && subname[:8] == "powerbar":
				if l.pb[i][0] == nil {
					l.pb[i][0] = readPowerBar("p1.", is, sff, at, l.snd)
				}
				if l.pb[i][1] == nil {
					l.pb[i][1] = readPowerBar("p2.", is, sff, at, l.snd)
				}
				if l.pb[i][2] == nil {
					l.pb[i][2] = readPowerBar("p3.", is, sff, at, l.snd)
				}
				if l.pb[i][3] == nil {
					l.pb[i][3] = readPowerBar("p4.", is, sff, at, l.snd)
				}
				if l.pb[i][4] == nil {
					l.pb[i][4] = readPowerBar("p5.", is, sff, at, l.snd)
				}
				if l.pb[i][5] == nil {
					l.pb[i][5] = readPowerBar("p6.", is, sff, at, l.snd)
				}
				if i == 5 || i == 7 {
					if l.pb[i][6] == nil {
						l.pb[i][6] = readPowerBar("p7.", is, sff, at, l.snd)
					}
					if l.pb[i][7] == nil {
						l.pb[i][7] = readPowerBar("p8.", is, sff, at, l.snd)
					}
				}
			}
		case "winicon":
			if l.wi[0] == nil {
				l.wi[0] = readLifeBarWinIcon("p1.", is, sff, at)
			}
			if l.wi[1] == nil {
				l.wi[1] = readLifeBarWinIcon("p2.", is, sff, at)
			}
		case "time":
			if l.ti == nil {
				l.ti = readLifeBarTime(is, sff, at)
			}
		case "combo":
			if l.co == nil {
				l.co = readLifeBarCombo(is)
			}
		case "round":
			if l.ro == nil {
				l.ro = readLifeBarRound(is, sff, at, l.snd, l.fnt[:])
			}
		case "challenger":
			if l.ch == nil {
				l.ch = readLifeBarChallenger(is, sff, at)
			}
		case "ratio":
			if l.ra[0] == nil {
				l.ra[0] = readLifeBarRatio("p1.", is, sff, at)
			}
			if l.ra[1] == nil {
				l.ra[1] = readLifeBarRatio("p2.", is, sff, at)
			}
		case "timer":
			if l.tr == nil {
				l.tr = readLifeBarTimer(is, sff, at)
			}
		case "score":
			if l.sc[0] == nil {
				l.sc[0] = readLifeBarScore("p1.", is, sff, at)
			}
			if l.sc[1] == nil {
				l.sc[1] = readLifeBarScore("p2.", is, sff, at)
			}
		case "match":
			if l.ma == nil {
				l.ma = readLifeBarMatch(is, sff, at)
			}
		case "ailevel":
			if l.ai[0] == nil {
				l.ai[0] = readLifeBarAiLevel("p1.", is, sff, at)
			}
			if l.ai[1] == nil {
				l.ai[1] = readLifeBarAiLevel("p2.", is, sff, at)
			}
		case "mode":
			if l.mo == nil {
				l.mo = readLifeBarMode(is, sff, at)
			}
		}
	}
	for k, v := range missing {
		if strings.Contains(k, "lifebar") {
			for i := 3; i < len(l.hb); i++ {
				if i == v {
					for j, d := range l.hb[1] {
						l.hb[i][j] = d
					}
				}
			}
		} else if strings.Contains(k, "powerbar") {
			for i := 1; i < len(l.pb); i++ {
				if i == v {
					for j, d := range l.pb[0] {
						l.pb[i][j] = d
					}
				}
			}
		} else if strings.Contains(k, "tag face") {
			for j, d := range l.fa[1] {
				l.fa[3][j] = d
			}
		} else if strings.Contains(k, "tag name") {
			for j, d := range l.nm[1] {
				l.nm[3][j] = d
			}
		}
	}
	return l, nil
}
func (l *Lifebar) step() {
	for ti, _ := range sys.tmode {
		for i := ti; i < l.num[0][ti]; i += 2 {
			l.hb[l.ref[0][ti]][i].step(float32(sys.chars[i][0].life)/
				float32(sys.chars[i][0].lifeMax), (sys.chars[i][0].getcombo != 0 ||
				sys.chars[i][0].ss.moveType == MT_H) &&
				!sys.chars[i][0].scf(SCF_over))
		}
	}
	for ti, _ := range sys.tmode {
		for i := ti; i < l.num[1][ti]; i += 2 {
			l.pb[l.ref[1][ti]][i].step(float32(sys.chars[i][0].power)/
				float32(sys.chars[i][0].powerMax), sys.chars[i][0].power/1000)
		}
	}
	for ti, _ := range sys.tmode {
		for i := ti; i < l.num[2][ti]; i += 2 {
			l.fa[l.ref[2][ti]][i].step()
		}
	}
	for ti, _ := range sys.tmode {
		for i := ti; i < l.num[3][ti]; i += 2 {
			l.nm[l.ref[3][ti]][i].step()
		}
	}
	for i := range l.wi {
		l.wi[i].step(sys.wins[i])
	}
	l.ti.step()
	cb, cd := [2]int32{}, [2]int32{}
	for i, ch := range sys.chars {
		for _, c := range ch {
			cb[^i&1] = Min(999, Max(c.getcombo, cb[^i&1]))
			cd[^i&1] = Max(c.getcombodmg, cd[^i&1])
		}
	}
	l.co.step(cb, cd)
	l.ch.step(l.snd)
	for ti, tm := range sys.tmode {
		if tm == TM_Turns {
			rl := sys.chars[ti][0].ratioLevel()
			if rl > 0 {
				l.ra[ti].step(rl-1)
			}
		}
	}
	l.tr.step()
	for i := range l.sc {
		l.sc[i].step()
	}
	l.ma.step()
	for i := range l.ai {
		l.ai[i].step()
	}
	if _, ok := l.mo[sys.gameMode]; ok {
		l.mo[sys.gameMode].step()
	}
}
func (l *Lifebar) reset() {
	for ti, tm := range sys.tmode {
		l.ref[0][ti] = int(tm)
		l.ref[1][ti] = int(tm)
		l.ref[2][ti] = int(tm)
		l.ref[3][ti] = int(tm)
		if tm == TM_Tag {
			if sys.numSimul[ti] == 2 { //Tag 2P
				l.ref[0][ti] = 3
				l.ref[1][ti] = 3
				l.ref[2][ti] = 3
				l.ref[3][ti] = 3
			} else { //Tag 3P/4P
				l.ref[0][ti] = int(sys.numSimul[ti]) + 3
				l.ref[1][ti] = 3
				l.ref[2][ti] = 3
				l.ref[3][ti] = 3
			}
		} else if tm == TM_Simul && sys.numSimul[ti] > 2 { //Simul 3P/4P
			l.ref[0][ti] = int(sys.numSimul[ti]) + 1
			l.ref[1][ti] = int(sys.numSimul[ti]) + 1
		}
		l.num[0][ti] = len(l.hb[l.ref[0][ti]])
		l.num[1][ti] = len(l.pb[l.ref[1][ti]])
		l.num[2][ti] = len(l.fa[l.ref[2][ti]])
		l.num[3][ti] = len(l.nm[l.ref[3][ti]])
		if tm == TM_Simul || tm == TM_Tag {
			l.num[0][ti] = int(sys.numSimul[ti]) * 2
			if sys.powerShare[ti] {
				l.num[1][ti] = 2
			} else if tm == TM_Simul {
				l.num[1][ti] = int(sys.numSimul[ti]) * 2
			}
			l.num[2][ti] = int(sys.numSimul[ti]) * 2
			l.num[3][ti] = int(sys.numSimul[ti]) * 2
		}
	}
	for _, hb := range l.hb {
		for i := range hb {
			hb[i].reset()
		}
	}
	for _, pb := range l.pb {
		for i := range pb {
			pb[i].reset()
		}
	}
	for _, fa := range l.fa {
		for i := range fa {
			fa[i].reset()
		}
	}
	for _, nm := range l.nm {
		for i := range nm {
			nm[i].reset()
		}
	}
	for i := range l.wi {
		l.wi[i].reset()
	}
	l.ti.reset()
	l.co.reset()
	l.ro.reset()
	l.ch.reset()
	for i := range l.ra {
		l.ra[i].reset()
	}
	l.tr.reset()
	for i := range l.sc {
		l.sc[i].reset()
	}
	l.ma.reset()
	for i := range l.ai {
		l.ai[i].reset()
	}
	if _, ok := l.mo[sys.gameMode]; ok {
		l.mo[sys.gameMode].reset()
	}
}
func (l *Lifebar) draw(layerno int16) {
	if !sys.statusDraw || !l.active {
		return
	}
	if !sys.sf(GSF_nobardisplay) && l.activeBars {
		for ti, _ := range sys.tmode {
			for i := ti; i < l.num[0][ti]; i += 2 {
				l.hb[l.ref[0][ti]][i].bgDraw(layerno)
			}
		}
		for ti, _ := range sys.tmode {
			for i := ti; i < l.num[0][ti]; i += 2 {
				l.hb[l.ref[0][ti]][i].draw(layerno, float32(sys.chars[i][0].life)/
					float32(sys.chars[i][0].lifeMax))
			}
		}
		for ti, _ := range sys.tmode {
			for i := ti; i < l.num[1][ti]; i += 2 {
				l.pb[l.ref[1][ti]][i].bgDraw(layerno)
			}
		}
		for ti, _ := range sys.tmode {
			for i := ti; i < l.num[1][ti]; i += 2 {
				l.pb[l.ref[1][ti]][i].draw(layerno, float32(sys.chars[i][0].power)/
					float32(sys.chars[i][0].powerMax), sys.chars[i][0].power/1000,
					l.fnt[:])
			}
		}
		for ti, _ := range sys.tmode {
			for i := ti; i < l.num[2][ti]; i += 2 {
				l.fa[l.ref[2][ti]][i].bgDraw(layerno)
			}
		}
		for ti, _ := range sys.tmode {
			for i := ti; i < l.num[2][ti]; i += 2 {
				if fspr := l.fa[l.ref[2][ti]][i].face; fspr != nil {
					pfx := sys.chars[i][0].getPalfx()
					sys.cgi[i].sff.palList.SwapPalMap(&pfx.remap)
					fspr.Pal = nil
					fspr.Pal = fspr.GetPal(&sys.cgi[i].sff.palList)
					sys.cgi[i].sff.palList.SwapPalMap(&pfx.remap)
					l.fa[l.ref[2][ti]][i].draw(layerno, pfx, i == sys.superplayer)
				}
			}
		}
		for ti, _ := range sys.tmode {
			for i := ti; i < l.num[3][ti]; i += 2 {
				l.nm[l.ref[3][ti]][i].bgDraw(layerno)
			}
		}
		for ti, _ := range sys.tmode {
			for i := ti; i < l.num[3][ti]; i += 2 {
				l.nm[l.ref[3][ti]][i].draw(layerno, l.fnt[:], sys.cgi[i].lifebarname)
			}
		}
		l.ti.bgDraw(layerno)
		l.ti.draw(layerno, l.fnt[:])
		for i := range l.wi {
			l.wi[i].draw(layerno, l.fnt[:])
		}
		for ti, tm := range sys.tmode {
			if tm == TM_Turns {
				rl := sys.chars[ti][0].ratioLevel()
				if rl > 0 {
					l.ra[ti].draw(layerno, rl-1)
				}
			}
		}
		l.tr.bgDraw(layerno)
		l.tr.draw(layerno, l.fnt[:])
		for i := range l.sc {
			l.sc[i].bgDraw(layerno)
			l.sc[i].draw(layerno, l.fnt[:], i)
		}
		l.ma.bgDraw(layerno)
		l.ma.draw(layerno, l.fnt[:])
		for i := range l.ai {
			l.ai[i].bgDraw(layerno)
			l.ai[i].draw(layerno, l.fnt[:], sys.com[sys.chars[i][0].playerNo])
		}

	}
	l.co.draw(layerno, l.fnt[:])
	l.ch.bgDraw(layerno)
	l.ch.draw(layerno, l.fnt[:])
	if _, ok := l.mo[sys.gameMode]; ok {
		l.mo[sys.gameMode].bgDraw(layerno)
		l.mo[sys.gameMode].draw(layerno, l.fnt[:])
	}
}
