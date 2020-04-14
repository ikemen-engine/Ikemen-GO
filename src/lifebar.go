package main

import (
	"fmt"
	"math"
	"strings"
	"regexp"
	"sort"
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
	front      map[float32]*AnimLayout
	oldlife    float32
	midlife    float32
	midlifeMin float32
	mlifetime  int32
}

func readHealthBar(pre string, is IniSection,
	sff *Sff, at AnimationTable) *HealthBar {
	hb := &HealthBar{oldlife: 1, midlife: 1, midlifeMin: 1,
	front: make(map[float32]*AnimLayout)}
	is.ReadI32(pre+"pos", &hb.pos[0], &hb.pos[1])
	is.ReadI32(pre+"range.x", &hb.range_x[0], &hb.range_x[1])
	hb.bg0 = *ReadAnimLayout(pre+"bg0.", is, sff, at, 0)
	hb.bg1 = *ReadAnimLayout(pre+"bg1.", is, sff, at, 0)
	hb.bg2 = *ReadAnimLayout(pre+"bg2.", is, sff, at, 0)
	hb.mid = *ReadAnimLayout(pre+"mid.", is, sff, at, 0)
	hb.front[0] = ReadAnimLayout(pre+"front.", is, sff, at, 0)
	for k, _ := range is {
		match, _ := regexp.MatchString(pre+"front[0-9]+\\.", k)
		if match {
			re := regexp.MustCompile("[0-9]+")
			submatchall := re.FindAllString(k, -1)
			if len(submatchall) == 2 {
				v := Atof(submatchall[1])
				hb.front[float32(v)] = ReadAnimLayout(pre+"front"+fmt.Sprintf("%v", v)+".", is, sff, at, 0)
			}
		}
	}
	return hb
}
func (hb *HealthBar) step(ref int, hbr *HealthBar) {
	life := float32(sys.chars[ref][0].life) / float32(sys.chars[ref][0].lifeMax)
	gethit := (sys.chars[ref][0].getcombo != 0 || sys.chars[ref][0].ss.moveType == MT_H) && !sys.chars[ref][0].scf(SCF_over)
	if len(hb.mid.anim.frames) > 0 && gethit {
		if hbr.mlifetime < 30 {
			hbr.mlifetime = 30
			hbr.midlife = hbr.oldlife
			hbr.midlifeMin = hbr.oldlife
		}
	} else {
		if hbr.mlifetime > 0 {
			hbr.mlifetime--
		}
		if len(hb.mid.anim.frames) > 0 && hbr.mlifetime <= 0 &&
			life < hbr.midlifeMin {
			hbr.midlifeMin += (life - hbr.midlifeMin) *
				(1 / (12 - (life-hbr.midlifeMin)*144))
		} else {
			hbr.midlifeMin = life
		}
		if (len(hb.mid.anim.frames) == 0 || hbr.mlifetime <= 0) &&
			hbr.midlife > hbr.midlifeMin {
			hbr.midlife += (hbr.midlifeMin - hbr.midlife) / 8
		}
		hbr.oldlife = life
	}
	mlmin := MaxF(hbr.midlifeMin, life)
	if hbr.midlife < mlmin {
		hbr.midlife += (mlmin - hbr.midlife) / 2
	}
	hb.bg0.Action()
	hb.bg1.Action()
	hb.bg2.Action()
	hb.mid.Action()
	var mv float32
	for k, _ := range hb.front {
		if k > mv && life >= k/100 {
			mv = k
		}
	}
	hb.front[mv].Action()
}
func (hb *HealthBar) reset() {
	hb.bg0.Reset()
	hb.bg1.Reset()
	hb.bg2.Reset()
	hb.mid.Reset()
	for _, v := range hb.front {
		v.Reset()
	}
}
func (hb *HealthBar) bgDraw(layerno int16) {
	hb.bg0.DrawScaled(float32(hb.pos[0])+sys.lifebarOffsetX, float32(hb.pos[1]), layerno, sys.lifebarScale)
	hb.bg1.DrawScaled(float32(hb.pos[0])+sys.lifebarOffsetX, float32(hb.pos[1]), layerno, sys.lifebarScale)
	hb.bg2.DrawScaled(float32(hb.pos[0])+sys.lifebarOffsetX, float32(hb.pos[1]), layerno, sys.lifebarScale)
}
func (hb *HealthBar) draw(layerno int16, ref int, hbr *HealthBar) {
	life := float32(sys.chars[ref][0].life) / float32(sys.chars[ref][0].lifeMax)
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
	if len(hb.mid.anim.frames) == 0 || life > hbr.midlife {
		life = hbr.midlife
	}
	lr, mr := width(life), width(hbr.midlife)
	if hb.range_x[0] < hb.range_x[1] {
		mr[0] += lr[2]
	}
	mr[2] -= Min(mr[2], lr[2])
	hb.mid.lay.DrawAnim(&mr, float32(hb.pos[0])+sys.lifebarOffsetX, float32(hb.pos[1]), sys.lifebarScale,
		layerno, &hb.mid.anim)
	var mv float32
	for k, _ := range hb.front {
		if k > mv && life >= k/100 {
			mv = k
		}
	}
	hb.front[mv].lay.DrawAnim(&lr, float32(hb.pos[0])+sys.lifebarOffsetX, float32(hb.pos[1]), sys.lifebarScale,
		layerno, &hb.front[mv].anim)
}

type PowerBar struct {
	snd          *Snd
	pos          [2]int32
	range_x      [2]int32
	bg0          AnimLayout
	bg1          AnimLayout
	bg2          AnimLayout
	mid          AnimLayout
	front        map[int32]*AnimLayout
	counter_font [3]int32
	counter_lay  Layout
	level_snd    [3][2]int32
	midpower     float32
	midpowerMin  float32
	prevLevel    int32
}

func newPowerBar(snd *Snd) (pb *PowerBar) {
	pb = &PowerBar{snd: snd, counter_font: [3]int32{-1},
		level_snd: [...][2]int32{{-1}, {-1}, {-1}},
		front: make(map[int32]*AnimLayout)}
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
	pb.front[0] = ReadAnimLayout(pre+"front.", is, sff, at, 0)
	for k, _ := range is {
		match, _ := regexp.MatchString(pre+"front[0-9]+\\.", k)
		if match {
			re := regexp.MustCompile("[0-9]+")
			submatchall := re.FindAllString(k, -1)
			if len(submatchall) == 2 {
				v := Atoi(submatchall[1])
				pb.front[v] = ReadAnimLayout(pre+"front"+fmt.Sprintf("%v", v)+".", is, sff, at, 0)
			}
		}
	}
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
func (pb *PowerBar) step(ref int, pbr *PowerBar) {
	power := float32(sys.chars[ref][0].power) / float32(sys.chars[ref][0].powerMax)
	level := sys.chars[ref][0].power / 1000
	value := sys.chars[ref][0].power
	pbr.midpower -= 1.0 / 144
	if power < pbr.midpowerMin {
		pbr.midpowerMin += (power - pbr.midpowerMin) *
			(1 / (12 - (power-pbr.midpowerMin)*144))
	} else {
		pbr.midpowerMin = power
	}
	if pbr.midpower < pbr.midpowerMin {
		pbr.midpower = pbr.midpowerMin
	}
	if level > pbr.prevLevel {
		i := Min(2, level-1)
		pb.snd.play(pb.level_snd[i])
	}
	pbr.prevLevel = level
	pb.bg0.Action()
	pb.bg1.Action()
	pb.bg2.Action()
	pb.mid.Action()
	var mv int32
	for k, _ := range pb.front {
		if k > mv && value >= k {
			mv = k
		}
	}
	pb.front[mv].Action()
}
func (pb *PowerBar) reset() {
	pb.bg0.Reset()
	pb.bg1.Reset()
	pb.bg2.Reset()
	pb.mid.Reset()
	for _, v := range pb.front {
		v.Reset()
	}
}
func (pb *PowerBar) bgDraw(layerno int16) {
	pb.bg0.DrawScaled(float32(pb.pos[0])+sys.lifebarOffsetX, float32(pb.pos[1]), layerno, sys.lifebarScale)
	pb.bg1.DrawScaled(float32(pb.pos[0])+sys.lifebarOffsetX, float32(pb.pos[1]), layerno, sys.lifebarScale)
	pb.bg2.DrawScaled(float32(pb.pos[0])+sys.lifebarOffsetX, float32(pb.pos[1]), layerno, sys.lifebarScale)
}
func (pb *PowerBar) draw(layerno int16, ref int, pbr *PowerBar, f []*Fnt) {
	power := float32(sys.chars[ref][0].power) / float32(sys.chars[ref][0].powerMax)
	level := sys.chars[ref][0].power / 1000
	value := sys.chars[ref][0].power
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
	pr, mr := width(power), width(pbr.midpower)
	if pb.range_x[0] < pb.range_x[1] {
		mr[0] += pr[2]
	}
	mr[2] -= Min(mr[2], pr[2])
	pb.mid.lay.DrawAnim(&mr, float32(pb.pos[0])+sys.lifebarOffsetX, float32(pb.pos[1]), sys.lifebarScale,
		layerno, &pb.mid.anim)
	var mv int32
	for k, _ := range pb.front {
		if k > mv && value >= k {
			mv = k
		}
	}
	pb.front[mv].lay.DrawAnim(&pr, float32(pb.pos[0])+sys.lifebarOffsetX, float32(pb.pos[1]), sys.lifebarScale,
		layerno, &pb.front[mv].anim)

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
	ko                AnimLayout
	face_spr          [2]int32
	face              *Sprite
	face_lay          Layout
	teammate_pos      [2]int32
	teammate_spacing  [2]int32
	teammate_bg       AnimLayout
	teammate_bg0      AnimLayout
	teammate_bg1      AnimLayout
	teammate_bg2      AnimLayout
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
	f.ko = *ReadAnimLayout(pre+"ko.", is, sff, at, 0)
	is.ReadI32(pre+"face.spr", &f.face_spr[0], &f.face_spr[1])
	f.face_lay = *ReadLayout(pre+"face.", is, 0)
	is.ReadI32(pre+"teammate.pos", &f.teammate_pos[0], &f.teammate_pos[1])
	is.ReadI32(pre+"teammate.spacing", &f.teammate_spacing[0],
		&f.teammate_spacing[1])
	f.teammate_bg = *ReadAnimLayout(pre+"teammate.bg.", is, sff, at, 0)
	f.teammate_bg0 = *ReadAnimLayout(pre+"teammate.bg0.", is, sff, at, 0)
	f.teammate_bg1 = *ReadAnimLayout(pre+"teammate.bg1.", is, sff, at, 0)
	f.teammate_bg2 = *ReadAnimLayout(pre+"teammate.bg2.", is, sff, at, 0)
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
	f.ko.Action()
	f.teammate_bg.Action()
	f.teammate_bg0.Action()
	f.teammate_bg1.Action()
	f.teammate_bg2.Action()
	f.teammate_ko.Action()
}
func (f *LifeBarFace) reset() {
	f.bg.Reset()
	f.bg0.Reset()
	f.bg1.Reset()
	f.bg2.Reset()
	f.ko.Reset()
	f.teammate_bg.Reset()
	f.teammate_bg0.Reset()
	f.teammate_bg1.Reset()
	f.teammate_bg2.Reset()
	f.teammate_ko.Reset()
}
func (f *LifeBarFace) bgDraw(layerno int16) {
	f.bg.DrawScaled(float32(f.pos[0])+sys.lifebarOffsetX, float32(f.pos[1]), layerno, sys.lifebarScale)
	f.bg0.DrawScaled(float32(f.pos[0])+sys.lifebarOffsetX, float32(f.pos[1]), layerno, sys.lifebarScale)
	f.bg1.DrawScaled(float32(f.pos[0])+sys.lifebarOffsetX, float32(f.pos[1]), layerno, sys.lifebarScale)
	f.bg2.DrawScaled(float32(f.pos[0])+sys.lifebarOffsetX, float32(f.pos[1]), layerno, sys.lifebarScale)
}
func (f *LifeBarFace) draw(layerno int16, ref int, far *LifeBarFace) {
	fspr := far.face
	if fspr == nil {
		return
	}
	pfx := sys.chars[ref][0].getPalfx()
	sys.cgi[ref].sff.palList.SwapPalMap(&pfx.remap)
	fspr.Pal = nil
	fspr.Pal = fspr.GetPal(&sys.cgi[ref].sff.palList)
	sys.cgi[ref].sff.palList.SwapPalMap(&pfx.remap)

	ob := sys.brightness
	if ref == sys.superplayer {
		sys.brightness = 256
	}
	f.face_lay.DrawSprite((float32(f.pos[0])+sys.lifebarOffsetX)*sys.lifebarScale, float32(f.pos[1])*sys.lifebarScale, layerno,
		far.face, pfx, f.scale*sys.lifebarPortraitScale, &sys.scrrect)
	if !sys.chars[ref][0].alive() {
		f.ko.DrawScaled(float32(f.pos[0])+sys.lifebarOffsetX, float32(f.pos[1]), layerno, sys.lifebarScale)
	}
	sys.brightness = ob
	i := int32(len(far.teammate_face)) - 1
	x := float32(f.teammate_pos[0] + f.teammate_spacing[0]*(i-1))
	y := float32(f.teammate_pos[1] + f.teammate_spacing[1]*(i-1))
	for ; i >= 0; i-- {
		if i != f.numko {
			f.teammate_bg.DrawScaled((x + sys.lifebarOffsetX), y, layerno, sys.lifebarScale)
			f.teammate_bg0.DrawScaled((x + sys.lifebarOffsetX), y, layerno, sys.lifebarScale)
			f.teammate_bg1.DrawScaled((x + sys.lifebarOffsetX), y, layerno, sys.lifebarScale)
			f.teammate_bg2.DrawScaled((x + sys.lifebarOffsetX), y, layerno, sys.lifebarScale)
			f.teammate_face_lay.DrawSprite((x+sys.lifebarOffsetX)*sys.lifebarScale, y*sys.lifebarScale, layerno, far.teammate_face[i], nil, f.teammate_scale[i]*sys.lifebarPortraitScale, &sys.scrrect)
			if i < f.numko {
				f.teammate_ko.DrawScaled((x + sys.lifebarOffsetX), y, layerno, sys.lifebarScale)
			}
			x -= float32(f.teammate_spacing[0])
			y -= float32(f.teammate_spacing[1])
		}
	}
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
func (n *LifeBarName) draw(layerno int16, ref int, f []*Fnt) {
	if n.name_font[0] >= 0 && int(n.name_font[0]) < len(f) {
		n.name_lay.DrawText((float32(n.pos[0]) + sys.lifebarOffsetX), float32(n.pos[1]), sys.lifebarScale, layerno,
			sys.cgi[ref].lifebarname, f[n.name_font[0]], n.name_font[1], n.name_font[2])
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
func (wi *LifeBarWinIcon) draw(layerno int16, f []*Fnt, side int) {
	bg0num := float64(sys.lifebar.ro.match_wins)
	if sys.tmode[^side&1] == TM_Turns {
		bg0num = float64(sys.numTurns[^side&1])
	}
	for i := 0; i < int(math.Min(float64(wi.useiconupto), bg0num)); i++ {
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
	pos           [2]int32
	start_x       float32
	counter_font  [3]int32
	counter_shake bool
	counter_lay   Layout
	text_font     [3]int32
	text_text     string
	text_lay      Layout
	displaytime   int32
	cur, old      int32
	curd, oldd    int32
	curp, oldp    float32
	resttime      int32
	counterX      float32
	shaketime     int32
	showspeed     float32
	hidespeed     float32
	separator     string
	places        int32
	firstAttack   bool
	counterHits   int
}

func newLifeBarCombo() *LifeBarCombo {
	return &LifeBarCombo{counter_font: [3]int32{-1}, text_font: [3]int32{-1},
		displaytime: 90, showspeed: 8, hidespeed: 4}
}
func readLifeBarCombo(pre string, is IniSection) *LifeBarCombo {
	c := newLifeBarCombo()
	is.ReadI32(pre+"pos", &c.pos[0], &c.pos[1])
	is.ReadF32(pre+"start.x", &c.start_x)
	if pre == "team2." { //mugen 1.0 implementation reuses winmugen code where both sides shared the same values
		c.pos[0] = sys.lifebarLocalcoord[0] - c.pos[0]
		c.start_x = float32(sys.lifebarLocalcoord[0]) - c.start_x
	}
	is.ReadI32(pre+"counter.font", &c.counter_font[0], &c.counter_font[1],
		&c.counter_font[2])
	is.ReadBool(pre+"counter.shake", &c.counter_shake)
	c.counter_lay = *ReadLayout(pre+"counter.", is, 2)
	c.counter_lay.offset = [2]float32{}
	is.ReadI32(pre+"text.font", &c.text_font[0], &c.text_font[1], &c.text_font[2])
	if _, ok := is[pre+"text.text"]; ok {
		c.text_text, _ = is.getString(pre+"text.text")
	}
	c.text_lay = *ReadLayout(pre+"text.", is, 2)
	is.ReadI32(pre+"displaytime", &c.displaytime)
	is.ReadF32(pre+"showspeed", &c.showspeed)
	c.showspeed = MaxF(1, c.showspeed)
	is.ReadF32(pre+"hidespeed", &c.hidespeed)
	c.separator, _ = is.getString("format.decimal.separator")
	is.ReadI32("format.decimal.places", &c.places)
	return c
}
func (c *LifeBarCombo) step(combo, damage int32, percentage float32) {
	if c.resttime > 0 {
		c.counterX -= c.counterX / c.showspeed
	} else {
		c.counterX -= sys.lifebarFontScale * c.hidespeed * float32(sys.lifebarLocalcoord[0])/320
		if c.counterX < c.start_x*2 {
			c.counterX = c.start_x * 2
		}
	}
	if c.shaketime > 0 {
		c.shaketime--
	}
	if AbsF(c.counterX) < 1 {
		c.resttime--
	}
	if combo >= 2 {
		if c.old != combo {
			c.cur = combo
			c.resttime = c.displaytime
			if c.counter_shake {
				c.shaketime = 15
			}
		}
		if c.oldd != damage {
			c.curd = damage
		}
		if c.oldp != percentage {
			c.curp = percentage
		}
	}
	c.old = combo
	c.oldd = damage
	c.oldp = percentage
}
func (c *LifeBarCombo) reset() {
	c.cur, c.old, c.curd, c.oldd, c.curp, c.oldp, c.resttime = 0, 0, 0, 0, 0, 0, 0
	c.counterX = c.start_x * 2
	c.shaketime = 0
	c.firstAttack = false
	c.counterHits = 0
}
func (c *LifeBarCombo) draw(layerno int16, f []*Fnt, side int) {
	haba := func(n int32) float32 {
		if c.counter_font[0] < 0 || int(c.counter_font[0]) >= len(f) {
			return 0
		}
		return float32(f[c.counter_font[0]].TextWidth(fmt.Sprintf("%v", n)))
	}
	if c.resttime <= 0 && c.counterX == c.start_x*2 {
		return
	}
	var x float32
	if side == 0 {
		if c.start_x <= 0 {
			x = c.counterX
		}
		x += float32(c.pos[0]) + haba(c.cur)
	} else {
		if c.start_x <= 0 {
			x = -c.counterX
		}
		x += 320/sys.lifebarScale - sys.lifebarOffsetX*2 - float32(c.pos[0])
	}
	if c.text_font[0] >= 0 && int(c.text_font[0]) < len(f) {
		//text := OldSprintf(c.text_text, c.cur)
		text := strings.Replace(c.text_text, "%i", fmt.Sprintf("%d", c.cur), 1)
		text = strings.Replace(text, "%d", fmt.Sprintf("%d", c.curd), 1)
		//split float value, round to decimal place
		s := strings.Split(fmt.Sprintf("%s", fmt.Sprintf("%.[2]*[1]f", c.curp, c.places)), ".")
		//decimal separator
		if c.places > 0 {
			if len(s) > 1 {
				s[0] = s[0] + c.separator + s[1]
			}
		}
		//replace %p with formatted string
		text = strings.Replace(text, "%p", s[0], 1)

		if side == 0 {
			if c.pos[0] != 0 {
				x += c.text_lay.offset[0] *
					((1 - sys.lifebarFontScale) * sys.lifebarFontScale)
			}
		} else {
			tmp := c.text_lay.offset[0]
			if c.pos[0] == 0 {
				tmp *= sys.lifebarFontScale
			}
			x -= tmp + float32(f[c.text_font[0]].TextWidth(text))*
				c.text_lay.scale[0]*sys.lifebarFontScale
		}
		c.text_lay.DrawText(x+sys.lifebarOffsetX, float32(c.pos[1]), sys.lifebarScale, layerno,
			text, f[c.text_font[0]], c.text_font[1], 1)
	}
	if c.counter_font[0] >= 0 && int(c.counter_font[0]) < len(f) {
		z := 1 + float32(c.shaketime)*(1.0/20)*
			float32(math.Sin(float64(c.shaketime)*(math.Pi/2.5)))
		c.counter_lay.DrawText((x+sys.lifebarOffsetX)/z, float32(c.pos[1])/z, z*sys.lifebarScale, layerno,
			fmt.Sprintf("%v", c.cur), f[c.counter_font[0]],
			c.counter_font[1], -1)
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
	win3, win4         AnimTextSnd
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
	if _, ok := is["win3.text"]; ok {
		r.win3 = *ReadAnimTextSnd("win3.", is, sff, at, 1)
	} else {
		r.win3 = r.win2
	}
	if _, ok := is["win4.text"]; ok {
		r.win4 = *ReadAnimTextSnd("win4.", is, sff, at, 1)
	} else {
		r.win4 = r.win2
	}
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
				if sys.numSimul[sys.winTeam] == 2 {
					f(&r.win2, 1)
				} else if sys.numSimul[sys.winTeam] == 3 {
					f(&r.win3, 1)
				} else {
					f(&r.win4, 1)
				}
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
	r.win3.Reset()
	r.win4.Reset()
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
				var inter []interface{}
				for i := sys.winTeam; i < len(sys.chars); i += 2 {
					if len(sys.chars[i]) > 0 {
						inter = append(inter, sys.cgi[i].displayname)
					}
				}
				if sys.numSimul[sys.winTeam] == 2 {
					tmp := r.win2.text
					r.win2.text = OldSprintf(tmp, inter...)
					r.win2.DrawScaled(float32(r.pos[0])+sys.lifebarOffsetX, float32(r.pos[1]), layerno, r.fnt, sys.lifebarScale)
					r.win2.text = tmp
				} else if sys.numSimul[sys.winTeam] == 3 {
					tmp := r.win3.text
					r.win3.text = OldSprintf(tmp, inter...)
					r.win3.DrawScaled(float32(r.pos[0])+sys.lifebarOffsetX, float32(r.pos[1]), layerno, r.fnt, sys.lifebarScale)
					r.win3.text = tmp
				} else {
					tmp := r.win4.text
					r.win4.text = OldSprintf(tmp, inter...)
					r.win4.DrawScaled(float32(r.pos[0])+sys.lifebarOffsetX, float32(r.pos[1]), layerno, r.fnt, sys.lifebarScale)
					r.win4.text = tmp
				}
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
	ref        [2]int
	order      [2][]int
	hb         [8][]*HealthBar
	pb         [8][]*PowerBar
	fa         [8][]*LifeBarFace
	nm         [8][]*LifeBarName
	wi         [2]*LifeBarWinIcon
	ti         *LifeBarTime
	co         [2]*LifeBarCombo
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
		hb: [...][]*HealthBar{make([]*HealthBar, 2), make([]*HealthBar, 8),
			make([]*HealthBar, 2), make([]*HealthBar, 8), make([]*HealthBar, 6),
			make([]*HealthBar, 8), make([]*HealthBar, 6), make([]*HealthBar, 8)},
		pb: [...][]*PowerBar{make([]*PowerBar, 2), make([]*PowerBar, 8),
			make([]*PowerBar, 2), make([]*PowerBar, 8), make([]*PowerBar, 6),
			make([]*PowerBar, 8), make([]*PowerBar, 6), make([]*PowerBar, 8)},
		fa: [...][]*LifeBarFace{make([]*LifeBarFace, 2), make([]*LifeBarFace, 8),
			make([]*LifeBarFace, 2), make([]*LifeBarFace, 8), make([]*LifeBarFace, 6),
			make([]*LifeBarFace, 8), make([]*LifeBarFace, 6), make([]*LifeBarFace, 8)},
		nm: [...][]*LifeBarName{make([]*LifeBarName, 2), make([]*LifeBarName, 8),
			make([]*LifeBarName, 2), make([]*LifeBarName, 8), make([]*LifeBarName, 6),
			make([]*LifeBarName, 8), make([]*LifeBarName, 6), make([]*LifeBarName, 8)},
		active: true, activeBars: true, activeMode: true}
	missing := map[string]int{
		"[tag lifebar]": 3, "[simul_3p lifebar]": 4, "[simul_4p lifebar]": 5,
		"[tag_3p lifebar]": 6, "[tag_4p lifebar]": 7, "[simul powerbar]": 1,
		"[turns powerbar]": 2, "[tag powerbar]": 3, "[simul_3p powerbar]": 4,
		"[simul_4p powerbar]": 5, "[tag_3p powerbar]": 6, "[tag_4p powerbar]": 7,
		"[tag face]": 3, "[simul_3p face]": 4, "[simul_4p face]": 5, "[tag_3p face]": 6,
		"[tag_4p face]": 7, "[tag name]": 3, "[simul_3p name]": 4, "[simul_4p name]": 5,
		"[tag_3p name]": 6, "[tag_4p name]": 7, "[challenger]": -1, "[ratio]": -1,
		"[timer]": -1, "[score]": -1, "[match]": -1, "[ailevel]": -1, "[mode]": -1,
	}
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
		case "simul ", "simul_3p ", "simul_4p ", "tag ", "tag_3p ", "tag_4p ":
			i := 1 //"simul "
			switch name {
			case "tag ":
				i = 3
			case "simul_3p ":
				i = 4
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
				if i != 4 && i != 6 {
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
				if i != 4 && i != 6 {
					if l.pb[i][6] == nil {
						l.pb[i][6] = readPowerBar("p7.", is, sff, at, l.snd)
					}
					if l.pb[i][7] == nil {
						l.pb[i][7] = readPowerBar("p8.", is, sff, at, l.snd)
					}
				}
			case len(subname) >= 4 && subname[:4] == "face":
				if l.fa[i][0] == nil {
					l.fa[i][0] = readLifeBarFace("p1.", is, sff, at)
				}
				if l.fa[i][1] == nil {
					l.fa[i][1] = readLifeBarFace("p2.", is, sff, at)
				}
				if l.fa[i][2] == nil {
					l.fa[i][2] = readLifeBarFace("p3.", is, sff, at)
				}
				if l.fa[i][3] == nil {
					l.fa[i][3] = readLifeBarFace("p4.", is, sff, at)
				}
				if l.fa[i][4] == nil {
					l.fa[i][4] = readLifeBarFace("p5.", is, sff, at)
				}
				if l.fa[i][5] == nil {
					l.fa[i][5] = readLifeBarFace("p6.", is, sff, at)
				}
				if i != 4 && i != 6 {
					if l.fa[i][6] == nil {
						l.fa[i][6] = readLifeBarFace("p7.", is, sff, at)
					}
					if l.fa[i][7] == nil {
						l.fa[i][7] = readLifeBarFace("p8.", is, sff, at)
					}
				}
			case len(subname) >= 4 && subname[:4] == "name":
				if l.nm[i][0] == nil {
					l.nm[i][0] = readLifeBarName("p1.", is, sff, at)
				}
				if l.nm[i][1] == nil {
					l.nm[i][1] = readLifeBarName("p2.", is, sff, at)
				}
				if l.nm[i][2] == nil {
					l.nm[i][2] = readLifeBarName("p3.", is, sff, at)
				}
				if l.nm[i][3] == nil {
					l.nm[i][3] = readLifeBarName("p4.", is, sff, at)
				}
				if l.nm[i][4] == nil {
					l.nm[i][4] = readLifeBarName("p5.", is, sff, at)
				}
				if l.nm[i][5] == nil {
					l.nm[i][5] = readLifeBarName("p6.", is, sff, at)
				}
				if i != 4 && i != 6 {
					if l.nm[i][6] == nil {
						l.nm[i][6] = readLifeBarName("p7.", is, sff, at)
					}
					if l.nm[i][7] == nil {
						l.nm[i][7] = readLifeBarName("p8.", is, sff, at)
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
			if l.co[0] == nil {
				if _, ok := is["team1.pos"]; ok {
					l.co[0] = readLifeBarCombo("team1.", is)
				} else {
					l.co[0] = readLifeBarCombo("", is)
				}
			}
			if l.co[1] == nil {
				if _, ok := is["team2.pos"]; ok {
					l.co[1] = readLifeBarCombo("team2.", is)
				} else {
					l.co[1] = readLifeBarCombo("", is)
				}
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
	//Iterate over map in a stable iteration order
	keys := make([]string, 0, len(missing))
	for k := range missing {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if strings.Contains(k, "lifebar") {
			for i := 3; i < len(l.hb); i++ {
				if i == missing[k] {
					for j := 0; j < len(l.hb[i]); j++ {
						if i == 6 || i == 7 {
							l.hb[i][j] = l.hb[3][j]
						} else {
							l.hb[i][j] = l.hb[1][j]
						}
					}
				}
			}
		} else if strings.Contains(k, "powerbar") {
			for i := 1; i < len(l.pb); i++ {
				if i == missing[k] {
					for j := 0; j < 2; j++ {
						l.pb[i][j] = l.pb[0][j]
					}
				}
			}
		} else if strings.Contains(k, "face") {
			for i := 3; i < len(l.fa); i++ {
				if i == missing[k] {
					for j := 0; j < len(l.fa[i]); j++ {
						if i == 6 || i == 7 {
							l.fa[i][j] = l.fa[3][j]
						} else {
							l.fa[i][j] = l.fa[1][j]
						}
					}
				}
			}
		} else if strings.Contains(k, "name") {
			for i := 3; i < len(l.nm); i++ {
				if i == missing[k] {
					for j := 0; j < len(l.nm[i]); j++ {
						if i == 6 || i == 7 {
							l.nm[i][j] = l.nm[3][j]
						} else {
							l.nm[i][j] = l.nm[1][j]
						}
					}
				}
			}
		}
	}
	return l, nil
}
func (l *Lifebar) step() {
	for ti, tm := range sys.tmode {
		if tm == TM_Tag && l.ro.timerActive {
			for i, v := range l.order[ti] {
				if !sys.chars[v][0].scf(SCF_standby) && sys.chars[v][0].alive() {
					if i != 0 {
						if i == len(l.order[ti])-1 {
							l.order[ti] = sliceMoveInt(l.order[ti], i, 0)
						} else {
							last := len(l.order[ti]) - 1
							for n := last; n > 0; n-- {
								if !sys.chars[l.order[ti][n]][0].alive() {
									last -= 1
								}
							}
							l.order[ti] = sliceMoveInt(l.order[ti], 0, last)
						}
					}
					break
				}
			}
		}
	}
	for ti, _ := range sys.tmode {
		for i, v := range l.order[ti] {
			//HealthBar
			l.hb[l.ref[ti]][i*2+ti].step(v, l.hb[l.ref[ti]][v])
			//PowerBar
			l.pb[l.ref[ti]][i*2+ti].step(v, l.pb[l.ref[ti]][v])
			//LifeBarFace
			l.fa[l.ref[ti]][i*2+ti].step()
			//LifeBarName
			l.nm[l.ref[ti]][i*2+ti].step()
		}
	}
	//LifeBarWinIcon
	for i := range l.wi {
		l.wi[i].step(sys.wins[i])
	}
	//LifeBarTime
	l.ti.step()
	//LifeBarCombo
	cb, cd, cp := [2]int32{}, [2]int32{}, [2]float32{}
	for i, ch := range sys.chars {
		for _, c := range ch {
			if c.getcombo > cb[^i&1] {
				cb[^i&1] = Min(999, Max(c.getcombo, cb[^i&1]))
				cd[^i&1] = Max(c.getcombodmg, cd[^i&1])
				cp[^i&1] = float32(cd[^i&1]) / float32(c.lifeMax) * 100
			}
		}
	}
	for i := range l.co {
		l.co[i].step(cb[i], cd[i], cp[i])
	}
	//LifeBarChallenger
	l.ch.step(l.snd)
	//LifeBarRatio
	for ti, tm := range sys.tmode {
		if tm == TM_Turns {
			rl := sys.chars[ti][0].ratioLevel()
			if rl > 0 {
				l.ra[ti].step(rl-1)
			}
		}
	}
	//LifeBarTimer
	l.tr.step()
	//LifeBarScore
	for i := range l.sc {
		l.sc[i].step()
	}
	//LifeBarMatch
	l.ma.step()
	//LifeBarAiLevel
	for i := range l.ai {
		l.ai[i].step()
	}
	//LifeBarMode
	if _, ok := l.mo[sys.gameMode]; ok {
		l.mo[sys.gameMode].step()
	}
}
func (l *Lifebar) reset() {
	var num [2]int
	for ti, tm := range sys.tmode {
		l.ref[ti] = int(tm)
		if tm == TM_Simul {
			if sys.numSimul[ti] == 3 {
				l.ref[ti] = 4 //Simul_3P (6)
			} else if sys.numSimul[ti] >= 4 {
				l.ref[ti] = 5 //Simul_4P (8)
			} else {
				l.ref[ti] = 1 //Simul (8)
			}
		} else if tm == TM_Tag {
			if sys.numSimul[ti] == 3 {
				l.ref[ti] = 6 //Tag_3P (6)
			} else if sys.numSimul[ti] >= 4 {
				l.ref[ti] = 7 //Tag_4P (8)
			} else {
				l.ref[ti] = 3 //Tag (8)
			}
		} else if tm == TM_Turns {
			l.ref[ti] = 2 //Turns (2)
		} else {
			l.ref[ti] = 0 //Single (2)
		}
		if tm == TM_Simul || tm == TM_Tag {
			num[ti] = int(math.Min(8, float64(sys.numSimul[ti])*2))
		} else {
			num[ti] = len(l.hb[l.ref[ti]])
		}
		l.order[ti] = []int{}
		for i := ti; i < num[ti]; i += 2 {
			l.order[ti] = append(l.order[ti], i)
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
	for i := range l.co {
		l.co[i].reset()
	}
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
		for ti, tm := range sys.tmode {
			for i, v := range l.order[ti] {
				//HealthBar
				l.hb[l.ref[ti]][i*2+ti].bgDraw(layerno)
				l.hb[l.ref[ti]][i*2+ti].draw(layerno, v, l.hb[l.ref[ti]][v])
				//PowerBar
				if sys.powerShare[ti] && (tm == TM_Simul || tm == TM_Tag) {
					if i == 0 {
						l.pb[l.ref[ti]][i*2+ti].bgDraw(layerno)
						l.pb[l.ref[ti]][i*2+ti].draw(layerno, i*2+ti, l.pb[l.ref[ti]][i*2+ti], l.fnt[:])
					}
				} else {
					l.pb[l.ref[ti]][i*2+ti].bgDraw(layerno)
					l.pb[l.ref[ti]][i*2+ti].draw(layerno, v, l.pb[l.ref[ti]][v], l.fnt[:])
				}
				//LifeBarFace
				l.fa[l.ref[ti]][i*2+ti].bgDraw(layerno)
				l.fa[l.ref[ti]][i*2+ti].draw(layerno, v, l.fa[l.ref[ti]][v])
				//LifeBarName
				l.nm[l.ref[ti]][i*2+ti].bgDraw(layerno)
				l.nm[l.ref[ti]][i*2+ti].draw(layerno, v, l.fnt[:])
			}
		}
		//LifeBarTime
		l.ti.bgDraw(layerno)
		l.ti.draw(layerno, l.fnt[:])
		//LifeBarWinIcon
		for i := range l.wi {
			l.wi[i].draw(layerno, l.fnt[:], i)
		}
		//LifeBarRatio
		for ti, tm := range sys.tmode {
			if tm == TM_Turns {
				rl := sys.chars[ti][0].ratioLevel()
				if rl > 0 {
					l.ra[ti].draw(layerno, rl-1)
				}
			}
		}
		//LifeBarTimer
		l.tr.bgDraw(layerno)
		l.tr.draw(layerno, l.fnt[:])
		//LifeBarScore
		for i := range l.sc {
			l.sc[i].bgDraw(layerno)
			l.sc[i].draw(layerno, l.fnt[:], i)
		}
		//LifeBarMatch
		l.ma.bgDraw(layerno)
		l.ma.draw(layerno, l.fnt[:])
		//LifeBarAiLevel
		for i := range l.ai {
			l.ai[i].bgDraw(layerno)
			l.ai[i].draw(layerno, l.fnt[:], sys.com[sys.chars[i][0].playerNo])
		}

	}
	//LifeBarCombo
	for i := range l.co {
		l.co[i].draw(layerno, l.fnt[:], i)
	}
	//LifeBarChallenger
	l.ch.bgDraw(layerno)
	l.ch.draw(layerno, l.fnt[:])
	//LifeBarMode
	if _, ok := l.mo[sys.gameMode]; ok {
		l.mo[sys.gameMode].bgDraw(layerno)
		l.mo[sys.gameMode].draw(layerno, l.fnt[:])
	}
}
