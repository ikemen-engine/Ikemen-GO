package main

import (
	"fmt"
	"math"
	"regexp"
	"sort"
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
	if *wt >= WT_N && *wt < WT_Perfect {
		*wt += WT_PN - WT_N
	}
}

type LbText struct {
	font  [6]int32
	text  string
	lay   Layout
	palfx *PalFX
	frgba [4]float32 //ttf fonts
}

func newLbText(align int32) *LbText {
	return &LbText{font: [...]int32{-1, 0, align, 255, 255, 255},
		palfx: newPalFX(), frgba: [...]float32{1.0, 1.0, 1.0, 1.0}}
}
func readLbText(pre string, is IniSection, str string, ln int16, f []*Fnt, align int32) *LbText {
	txt := newLbText(align)
	is.ReadI32(pre+"font", &txt.font[0], &txt.font[1], &txt.font[2],
		&txt.font[3], &txt.font[4], &txt.font[5])
	if txt.font[0] >= 0 && int(txt.font[0]) < len(f) && f[txt.font[0]] == nil {
		sys.appendToConsole(fmt.Sprintf("WARNING: Missing font assigned to %v lifebar parameter: %v", pre+"font", txt.font[0]))
		sys.errLog.Printf("Missing font assigned to %v lifebar parameter: %v\n", pre+"font", txt.font[0])
	}
	if _, ok := is[pre+"text"]; ok {
		txt.text, _ = is.getString(pre + "text")
	} else {
		txt.text = str
	}
	txt.lay = *ReadLayout(pre, is, ln)
	txt.palfx.setColor(txt.font[3], txt.font[4], txt.font[5])
	return txt
}

type LbBgTextSnd struct {
	pos         [2]int32
	text        LbText
	bg          AnimLayout
	time        int32
	displaytime int32
	snd         [2]int32
	sndtime     int32
	cnt         int32
}

func newLbBgTextSnd() LbBgTextSnd {
	return LbBgTextSnd{snd: [2]int32{-1}}
}
func readLbBgTextSnd(pre string, is IniSection,
	sff *Sff, at AnimationTable, ln int16, f []*Fnt) LbBgTextSnd {
	bts := newLbBgTextSnd()
	is.ReadI32(pre+"pos", &bts.pos[0], &bts.pos[1])
	bts.text = *readLbText(pre+"text.", is, "", ln, f, 0)
	bts.bg = *ReadAnimLayout(pre+"bg.", is, sff, at, ln)
	is.ReadI32(pre+"time", &bts.time)
	is.ReadI32(pre+"displaytime", &bts.displaytime)
	is.ReadI32(pre+"snd", &bts.snd[0], &bts.snd[1])
	is.ReadI32(pre+"sndtime", &bts.sndtime)
	return bts
}
func (bts *LbBgTextSnd) step(snd *Snd) {
	if bts.cnt == bts.sndtime {
		snd.play(bts.snd, 100)
	}
	if bts.cnt >= bts.time {
		bts.bg.Action()
	}
	bts.cnt++
}
func (bts *LbBgTextSnd) reset() {
	bts.cnt = 0
	bts.bg.Reset()
}
func (bts *LbBgTextSnd) bgDraw(layerno int16) {
	if bts.cnt > bts.time && bts.cnt <= bts.time+bts.displaytime {
		bts.bg.DrawScaled(float32(bts.pos[0])+sys.lifebarOffsetX, float32(bts.pos[1]), layerno, sys.lifebarScale)
	}
}
func (bts *LbBgTextSnd) draw(layerno int16, f []*Fnt) {
	if bts.cnt > bts.time && bts.cnt <= bts.time+bts.displaytime &&
		bts.text.font[0] >= 0 && int(bts.text.font[0]) < len(f) && f[bts.text.font[0]] != nil {
		bts.text.lay.DrawText(float32(bts.pos[0])+sys.lifebarOffsetX, float32(bts.pos[1]), sys.lifebarScale, layerno,
			bts.text.text, f[bts.text.font[0]], bts.text.font[1], bts.text.font[2], bts.text.palfx, bts.text.frgba)
	}
}

type HealthBar struct {
	pos        [2]int32
	range_x    [2]int32
	bg0        AnimLayout
	bg1        AnimLayout
	bg2        AnimLayout
	top        AnimLayout
	mid        AnimLayout
	red        map[int32]*AnimLayout
	front      map[float32]*AnimLayout
	shift      AnimLayout
	warn       AnimLayout
	warn_range [2]int32
	value      LbText
	oldlife    float32
	midlife    float32
	midlifeMin float32
	mlifetime  int32
	mid_shift  bool
	mid_freeze bool
	mid_delay  int32
	mid_mult   float32
	mid_steps  float32
	gethit     bool
}

func newHealthBar() *HealthBar {
	return &HealthBar{oldlife: 1, midlife: 1, midlifeMin: 1,
		red: make(map[int32]*AnimLayout), front: make(map[float32]*AnimLayout),
		mid_freeze: true, mid_delay: 30, mid_mult: 1.0, mid_steps: 8.0}
}
func readHealthBar(pre string, is IniSection,
	sff *Sff, at AnimationTable, f []*Fnt) *HealthBar {
	hb := newHealthBar()
	is.ReadI32(pre+"pos", &hb.pos[0], &hb.pos[1])
	is.ReadI32(pre+"range.x", &hb.range_x[0], &hb.range_x[1])
	hb.bg0 = *ReadAnimLayout(pre+"bg0.", is, sff, at, 0)
	hb.bg1 = *ReadAnimLayout(pre+"bg1.", is, sff, at, 0)
	hb.bg2 = *ReadAnimLayout(pre+"bg2.", is, sff, at, 0)
	hb.top = *ReadAnimLayout(pre+"top.", is, sff, at, 0)
	hb.mid = *ReadAnimLayout(pre+"mid.", is, sff, at, 0)
	hb.front[0] = ReadAnimLayout(pre+"front.", is, sff, at, 0)
	r, _ := regexp.Compile(pre + "front[0-9]+\\.")
	for k := range is {
		if r.MatchString(k) {
			re := regexp.MustCompile("[0-9]+")
			submatchall := re.FindAllString(k, -1)
			if len(submatchall) == 2 {
				v := Atof(submatchall[1])
				if _, ok := hb.front[float32(v)]; !ok {
					hb.front[float32(v)] = ReadAnimLayout(pre+"front"+fmt.Sprintf("%v", v)+".", is, sff, at, 0)
				}
			}
		}
	}
	hb.shift = *ReadAnimLayout(pre+"shift.", is, sff, at, 0)
	hb.red[0] = ReadAnimLayout(pre+"red.", is, sff, at, 0)
	r, _ = regexp.Compile(pre + "red[0-9]+\\.")
	for k := range is {
		if r.MatchString(k) {
			re := regexp.MustCompile("[0-9]+")
			submatchall := re.FindAllString(k, -1)
			if len(submatchall) == 2 {
				v := Atoi(submatchall[1])
				if _, ok := hb.red[v]; !ok {
					hb.red[v] = ReadAnimLayout(pre+"red"+fmt.Sprintf("%v", v)+".", is, sff, at, 0)
				}
			}
		}
	}
	hb.value = *readLbText(pre+"value.", is, "%d", 0, f, 0)
	is.ReadBool("mid.shift", &hb.mid_shift)
	is.ReadBool("mid.freeze", &hb.mid_freeze)
	is.ReadI32("mid.delay", &hb.mid_delay)
	is.ReadF32("mid.mult", &hb.mid_mult)
	is.ReadF32("mid.steps", &hb.mid_steps)
	hb.mid_steps = MaxF(1, hb.mid_steps)
	is.ReadI32(pre+"warn.range", &hb.warn_range[0], &hb.warn_range[1])
	hb.warn = *ReadAnimLayout(pre+"warn.", is, sff, at, 0)
	return hb
}
func (hb *HealthBar) step(ref int, hbr *HealthBar) {
	var life float32 = float32(sys.chars[ref][0].life) / float32(sys.chars[ref][0].lifeMax)
	//redlife := (float32(sys.chars[ref][0].life) + float32(sys.chars[ref][0].redLife)) / float32(sys.chars[ref][0].lifeMax)
	var redVal int32 = sys.chars[ref][0].redLife
	var getHit bool = (sys.chars[ref][0].fakeReceivedHits != 0 || sys.chars[ref][0].ss.moveType == MT_H) && !sys.chars[ref][0].scf(SCF_over)

	hb.shift.anim.srcAlpha = int16(255 * (1 - life))
	hb.shift.anim.dstAlpha = int16(255 * life)
	if !hb.mid_freeze && getHit && !hb.gethit && len(hb.mid.anim.frames) > 0 {
		hbr.mlifetime = hb.mid_delay
		hbr.midlife = hbr.oldlife
		hbr.midlifeMin = hbr.oldlife
	}
	hb.gethit = getHit
	if hb.mid_freeze && getHit && len(hb.mid.anim.frames) > 0 {
		if hbr.mlifetime < hb.mid_delay {
			hbr.mlifetime = hb.mid_delay
			hbr.midlife = hbr.oldlife
			hbr.midlifeMin = hbr.oldlife
		}
	} else {
		if hbr.mlifetime > 0 {
			hbr.mlifetime--
		}
		if len(hb.mid.anim.frames) > 0 && hbr.mlifetime <= 0 && life < hbr.midlifeMin {
			hbr.midlifeMin += (life - hbr.midlifeMin) * (1 / (12 - (life-hbr.midlifeMin)*144)) * hb.mid_mult
			if hbr.midlifeMin < life {
				hbr.midlifeMin = life
			}
		} else {
			hbr.midlifeMin = life
		}
		if (len(hb.mid.anim.frames) == 0 || hbr.mlifetime <= 0) && hbr.midlife > hbr.midlifeMin {
			hbr.midlife += (hbr.midlifeMin - hbr.midlife) / hb.mid_steps
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
	hb.top.Action()
	hb.mid.Action()
	var rv int32
	for k := range hb.red {
		if k > rv && redVal >= k {
			rv = k
		}
	}
	hb.red[rv].Action()
	var fv float32
	for k := range hb.front {
		if k > fv && life >= k/100 {
			fv = k
		}
	}
	hb.front[fv].Action()
	hb.shift.Action()
	hb.warn.Action()
}
func (hb *HealthBar) reset() {
	hb.bg0.Reset()
	hb.bg1.Reset()
	hb.bg2.Reset()
	hb.top.Reset()
	hb.mid.Reset()
	for _, v := range hb.front {
		v.Reset()
	}
	hb.shift.Reset()
	hb.shift.anim.srcAlpha = 0
	hb.shift.anim.dstAlpha = 255
	for _, v := range hb.red {
		v.Reset()
	}
	hb.warn.Reset()
}
func (hb *HealthBar) bgDraw(layerno int16) {
	hb.bg0.DrawScaled(float32(hb.pos[0])+sys.lifebarOffsetX, float32(hb.pos[1]), layerno, sys.lifebarScale)
	hb.bg1.DrawScaled(float32(hb.pos[0])+sys.lifebarOffsetX, float32(hb.pos[1]), layerno, sys.lifebarScale)
	hb.bg2.DrawScaled(float32(hb.pos[0])+sys.lifebarOffsetX, float32(hb.pos[1]), layerno, sys.lifebarScale)
}
func (hb *HealthBar) draw(layerno int16, ref int, hbr *HealthBar, f []*Fnt) {
	life := float32(sys.chars[ref][0].life) / float32(sys.chars[ref][0].lifeMax)
	redlife := (float32(sys.chars[ref][0].life) + float32(sys.chars[ref][0].redLife)) / float32(sys.chars[ref][0].lifeMax)
	redval := sys.chars[ref][0].redLife
	var MidPos = (float32(sys.gameWidth-320) / 2)
	width := func(life float32) (r [4]int32) {
		r = sys.scrrect
		if hb.range_x[0] < hb.range_x[1] {
			r[0] = int32((((float32(hb.pos[0]+hb.range_x[0])+sys.lifebarOffsetX)*sys.lifebarScale)+MidPos)*sys.widthScale + 0.5)
			r[2] = int32((float32(hb.range_x[1]-hb.range_x[0]+1)*sys.lifebarScale)*life*sys.widthScale + 0.5)
		} else {
			r[2] = int32(((float32(hb.range_x[0]-hb.range_x[1]+1)*sys.lifebarScale)*life-(sys.lifebarOffsetX*sys.lifebarScale))*sys.widthScale + 0.5)
			r[0] = int32(((float32(hb.pos[0]+hb.range_x[0]+1)*sys.lifebarScale)+MidPos)*sys.widthScale+0.5) - r[2]
		}
		return
	}
	if len(hb.mid.anim.frames) == 0 || life > hbr.midlife {
		life = hbr.midlife
	}
	lr, mr, rr := width(life), width(hbr.midlife), width(redlife)
	if hb.range_x[0] < hb.range_x[1] {
		mr[0] += lr[2]
		//rr[0] += lr[2]
	}
	mr[2] -= Min(mr[2], lr[2])
	//rr[2] -= Min(rr[2], lr[2])
	var rv int32
	if sys.lifebar.activeRl {
		for k := range hb.red {
			if k > rv && redval >= k {
				rv = k
			}
		}
		hb.red[rv].lay.DrawAnim(&rr, float32(hb.pos[0])+sys.lifebarOffsetX, float32(hb.pos[1]), sys.lifebarScale,
			layerno, &hb.red[rv].anim, hb.red[rv].palfx)
	}
	hb.mid.lay.DrawAnim(&mr, float32(hb.pos[0])+sys.lifebarOffsetX, float32(hb.pos[1]), sys.lifebarScale,
		layerno, &hb.mid.anim, hb.mid.palfx)
	if hb.mid_shift {
		hb.shift.lay.DrawAnim(&mr, float32(hb.pos[0])+sys.lifebarOffsetX, float32(hb.pos[1]), sys.lifebarScale,
			layerno, &hb.shift.anim, hb.shift.palfx)
	}
	var fv float32
	for k := range hb.front {
		if k > fv && life >= k/100 {
			fv = k
		}
	}
	hb.front[fv].lay.DrawAnim(&lr, float32(hb.pos[0])+sys.lifebarOffsetX, float32(hb.pos[1]), sys.lifebarScale,
		layerno, &hb.front[fv].anim, hb.front[fv].palfx)
	hb.shift.lay.DrawAnim(&lr, float32(hb.pos[0])+sys.lifebarOffsetX, float32(hb.pos[1]), sys.lifebarScale,
		layerno, &hb.shift.anim, hb.shift.palfx)
	if hb.value.font[0] >= 0 && int(hb.value.font[0]) < len(f) && f[hb.value.font[0]] != nil {
		text := strings.Replace(hb.value.text, "%d", fmt.Sprintf("%v", sys.chars[ref][0].life), 1)
		text = strings.Replace(text, "%p", fmt.Sprintf("%v", math.Round(float64(life)*100)), 1)
		hb.value.lay.DrawText(float32(hb.pos[0])+sys.lifebarOffsetX, float32(hb.pos[1]), sys.lifebarScale,
			layerno, text, f[hb.value.font[0]], hb.value.font[1], hb.value.font[2], hb.value.palfx, hb.value.frgba)
	}
	hb.top.DrawScaled(float32(hb.pos[0])+sys.lifebarOffsetX, float32(hb.pos[1]), layerno, sys.lifebarScale)
	if life < float32(hb.warn_range[0])/100 && life > float32(hb.warn_range[1])/100 {
		hb.warn.DrawScaled(float32(hb.pos[0])+sys.lifebarOffsetX, float32(hb.pos[1]), layerno, sys.lifebarScale)
	}
}

type PowerBar struct {
	pos         [2]int32
	range_x     [2]int32
	bg0         AnimLayout
	bg1         AnimLayout
	bg2         AnimLayout
	top         AnimLayout
	mid         AnimLayout
	front       map[int32]*AnimLayout
	shift       AnimLayout
	counter     LbText
	value       LbText
	level_snd   [9][2]int32
	midpower    float32
	midpowerMin float32
	prevLevel   int32
	levelbars   bool
}

func newPowerBar() *PowerBar {
	return &PowerBar{level_snd: [9][2]int32{{-1}, {-1}, {-1}},
		front: make(map[int32]*AnimLayout)}
}
func readPowerBar(pre string, is IniSection,
	sff *Sff, at AnimationTable, f []*Fnt) *PowerBar {
	pb := newPowerBar()
	is.ReadI32(pre+"pos", &pb.pos[0], &pb.pos[1])
	is.ReadI32(pre+"range.x", &pb.range_x[0], &pb.range_x[1])
	pb.bg0 = *ReadAnimLayout(pre+"bg0.", is, sff, at, 0)
	pb.bg1 = *ReadAnimLayout(pre+"bg1.", is, sff, at, 0)
	pb.bg2 = *ReadAnimLayout(pre+"bg2.", is, sff, at, 0)
	pb.top = *ReadAnimLayout(pre+"top.", is, sff, at, 0)
	pb.mid = *ReadAnimLayout(pre+"mid.", is, sff, at, 0)
	pb.front[0] = ReadAnimLayout(pre+"front.", is, sff, at, 0)
	r, _ := regexp.Compile(pre + "front[0-9]+\\.")
	for k := range is {
		if r.MatchString(k) {
			re := regexp.MustCompile("[0-9]+")
			submatchall := re.FindAllString(k, -1)
			if len(submatchall) == 2 {
				v := Atoi(submatchall[1])
				if _, ok := pb.front[v]; !ok {
					pb.front[v] = ReadAnimLayout(pre+"front"+fmt.Sprintf("%v", v)+".", is, sff, at, 0)
				}
			}
		}
	}
	pb.shift = *ReadAnimLayout(pre+"shift.", is, sff, at, 0)
	pb.counter = *readLbText(pre+"counter.", is, "%i", 0, f, 0)
	pb.value = *readLbText(pre+"value.", is, "", 0, f, 0)
	for i := range pb.level_snd {
		if !is.ReadI32(fmt.Sprintf("%vlevel%v.snd", pre, i+1), &pb.level_snd[i][0],
			&pb.level_snd[i][1]) {
			is.ReadI32(fmt.Sprintf("level%v.snd", i+1), &pb.level_snd[i][0],
				&pb.level_snd[i][1])
		}
	}
	is.ReadBool(pre+"levelbars", &pb.levelbars)
	return pb
}
func (pb *PowerBar) step(ref int, pbr *PowerBar, snd *Snd) {
	power := float32(sys.chars[ref][0].power) / float32(sys.chars[ref][0].powerMax)
	level := sys.chars[ref][0].power / 1000
	pbval := sys.chars[ref][0].power
	if pb.levelbars {
		power = float32(sys.chars[ref][0].power)/1000 - MinF(float32(level), float32(sys.chars[ref][0].powerMax)/1000-1)
	}
	pb.shift.anim.srcAlpha = int16(255 * (1 - power))
	pb.shift.anim.dstAlpha = int16(255 * power)
	pbr.midpower -= 1.0 / 144
	if power < pbr.midpowerMin {
		pbr.midpowerMin += (power - pbr.midpowerMin) * (1 / (12 - (power-pbr.midpowerMin)*144))
	} else {
		pbr.midpowerMin = power
	}
	if pbr.midpower < pbr.midpowerMin {
		pbr.midpower = pbr.midpowerMin
	}
	if level > pbr.prevLevel {
		i := Min(8, level-1)
		snd.play(pb.level_snd[i], 100)
	}
	pbr.prevLevel = level
	pb.bg0.Action()
	pb.bg1.Action()
	pb.bg2.Action()
	pb.top.Action()
	pb.mid.Action()
	var fv int32
	for k := range pb.front {
		if k > fv && pbval >= k {
			fv = k
		}
	}
	pb.front[fv].Action()
	pb.shift.Action()
}
func (pb *PowerBar) reset() {
	pb.bg0.Reset()
	pb.bg1.Reset()
	pb.bg2.Reset()
	pb.top.Reset()
	pb.mid.Reset()
	for _, v := range pb.front {
		v.Reset()
	}
	pb.shift.Reset()
	pb.shift.anim.srcAlpha = 0
	pb.shift.anim.dstAlpha = 255
}
func (pb *PowerBar) bgDraw(layerno int16) {
	pb.bg0.DrawScaled(float32(pb.pos[0])+sys.lifebarOffsetX, float32(pb.pos[1]), layerno, sys.lifebarScale)
	pb.bg1.DrawScaled(float32(pb.pos[0])+sys.lifebarOffsetX, float32(pb.pos[1]), layerno, sys.lifebarScale)
	pb.bg2.DrawScaled(float32(pb.pos[0])+sys.lifebarOffsetX, float32(pb.pos[1]), layerno, sys.lifebarScale)
}
func (pb *PowerBar) draw(layerno int16, ref int, pbr *PowerBar, f []*Fnt) {
	power := float32(sys.chars[ref][0].power) / float32(sys.chars[ref][0].powerMax)
	level := sys.chars[ref][0].power / 1000
	pbval := sys.chars[ref][0].power
	if pb.levelbars {
		power = float32(sys.chars[ref][0].power)/1000 - MinF(float32(level), float32(sys.chars[ref][0].powerMax)/1000-1)
	}
	var MidPos = (float32(sys.gameWidth-320) / 2)
	width := func(power float32) (r [4]int32) {
		r = sys.scrrect
		if pb.range_x[0] < pb.range_x[1] {
			r[0] = int32((((float32(pb.pos[0]+pb.range_x[0])+sys.lifebarOffsetX)*sys.lifebarScale)+MidPos)*sys.widthScale + 0.5)
			r[2] = int32((float32(pb.range_x[1]-pb.range_x[0]+1)*sys.lifebarScale)*power*sys.widthScale + 0.5)
		} else {
			r[2] = int32(((float32(pb.range_x[0]-pb.range_x[1]+1)*sys.lifebarScale)*power-(sys.lifebarOffsetX*sys.lifebarScale))*sys.widthScale + 0.5)
			r[0] = int32(((float32(pb.pos[0]+pb.range_x[0]+1)*sys.lifebarScale)+MidPos)*sys.widthScale+0.5) - r[2]
		}
		return
	}
	pr, mr := width(power), width(pbr.midpower)
	if pb.range_x[0] < pb.range_x[1] {
		mr[0] += pr[2]
	}
	mr[2] -= Min(mr[2], pr[2])
	pb.mid.lay.DrawAnim(&mr, float32(pb.pos[0])+sys.lifebarOffsetX, float32(pb.pos[1]), sys.lifebarScale,
		layerno, &pb.mid.anim, pb.mid.palfx)
	var fv int32
	for k := range pb.front {
		if k > fv && pbval >= k {
			fv = k
		}
	}
	pb.front[fv].lay.DrawAnim(&pr, float32(pb.pos[0])+sys.lifebarOffsetX, float32(pb.pos[1]), sys.lifebarScale,
		layerno, &pb.front[fv].anim, pb.front[fv].palfx)
	pb.shift.lay.DrawAnim(&pr, float32(pb.pos[0])+sys.lifebarOffsetX, float32(pb.pos[1]), sys.lifebarScale,
		layerno, &pb.shift.anim, pb.shift.palfx)
	if pb.counter.font[0] >= 0 && int(pb.counter.font[0]) < len(f) && f[pb.counter.font[0]] != nil {
		pb.counter.lay.DrawText(float32(pb.pos[0])+sys.lifebarOffsetX, float32(pb.pos[1]), sys.lifebarScale,
			layerno, strings.Replace(pb.counter.text, "%i", fmt.Sprintf("%v", level), 1), f[pb.counter.font[0]],
			pb.counter.font[1], pb.counter.font[2], pb.counter.palfx, pb.counter.frgba)
	}
	if pb.value.font[0] >= 0 && int(pb.value.font[0]) < len(f) && f[pb.value.font[0]] != nil {
		text := strings.Replace(pb.value.text, "%d", fmt.Sprintf("%v", pbval), 1)
		text = strings.Replace(text, "%p", fmt.Sprintf("%v", math.Round(float64(power)*100)), 1)
		pb.value.lay.DrawText(float32(pb.pos[0])+sys.lifebarOffsetX, float32(pb.pos[1]), sys.lifebarScale,
			layerno, text, f[pb.value.font[0]], pb.value.font[1], pb.value.font[2], pb.value.palfx, pb.value.frgba)
	}
	pb.top.DrawScaled(float32(pb.pos[0])+sys.lifebarOffsetX, float32(pb.pos[1]), layerno, sys.lifebarScale)
}

type GuardBar struct {
	pos         [2]int32
	range_x     [2]int32
	bg0         AnimLayout
	bg1         AnimLayout
	bg2         AnimLayout
	top         AnimLayout
	mid         AnimLayout
	warn        AnimLayout
	warn_range  [2]int32
	value       LbText
	front       map[float32]*AnimLayout
	shift       AnimLayout
	midpower    float32
	midpowerMin float32
	//prevLevel   int32
}

func newGuardBar() (gb *GuardBar) {
	gb = &GuardBar{front: make(map[float32]*AnimLayout)}
	return
}
func readGuardBar(pre string, is IniSection,
	sff *Sff, at AnimationTable, f []*Fnt) *GuardBar {
	gb := newGuardBar()
	is.ReadI32(pre+"pos", &gb.pos[0], &gb.pos[1])
	is.ReadI32(pre+"range.x", &gb.range_x[0], &gb.range_x[1])
	gb.bg0 = *ReadAnimLayout(pre+"bg0.", is, sff, at, 0)
	gb.bg1 = *ReadAnimLayout(pre+"bg1.", is, sff, at, 0)
	gb.bg2 = *ReadAnimLayout(pre+"bg2.", is, sff, at, 0)
	gb.top = *ReadAnimLayout(pre+"top.", is, sff, at, 0)
	gb.mid = *ReadAnimLayout(pre+"mid.", is, sff, at, 0)
	gb.front[0] = ReadAnimLayout(pre+"front.", is, sff, at, 0)
	r, _ := regexp.Compile(pre + "front[0-9]+\\.")
	for k := range is {
		if r.MatchString(k) {
			re := regexp.MustCompile("[0-9]+")
			submatchall := re.FindAllString(k, -1)
			if len(submatchall) == 2 {
				v := Atof(submatchall[1])
				if _, ok := gb.front[float32(v)]; !ok {
					gb.front[float32(v)] = ReadAnimLayout(pre+"front"+fmt.Sprintf("%v", v)+".", is, sff, at, 0)
				}
			}
		}
	}
	gb.shift = *ReadAnimLayout(pre+"shift.", is, sff, at, 0)
	gb.value = *readLbText(pre+"value.", is, "%d", 0, f, 0)
	is.ReadI32(pre+"warn.range", &gb.warn_range[0], &gb.warn_range[1])
	gb.warn = *ReadAnimLayout(pre+"warn.", is, sff, at, 0)
	return gb
}
func (gb *GuardBar) step(ref int, gbr *GuardBar, snd *Snd) {
	if !sys.lifebar.activeGb {
		return
	}
	power := float32(sys.chars[ref][0].guardPoints) / float32(sys.chars[ref][0].guardPointsMax)
	gb.shift.anim.srcAlpha = int16(255 * (1 - power))
	gb.shift.anim.dstAlpha = int16(255 * power)
	gbr.midpower -= 1.0 / 144
	if power < gbr.midpowerMin {
		gbr.midpowerMin += (power - gbr.midpowerMin) * (1 / (12 - (power-gbr.midpowerMin)*144))
	} else {
		gbr.midpowerMin = power
	}
	if gbr.midpower < gbr.midpowerMin {
		gbr.midpower = gbr.midpowerMin
	}
	gb.bg0.Action()
	gb.bg1.Action()
	gb.bg2.Action()
	gb.top.Action()
	gb.mid.Action()
	var mv float32
	for k := range gb.front {
		if k > mv && power >= k/100 {
			mv = k
		}
	}
	gb.front[mv].Action()
	gb.shift.Action()
	gb.warn.Action()
}
func (gb *GuardBar) reset() {
	gb.bg0.Reset()
	gb.bg1.Reset()
	gb.bg2.Reset()
	gb.top.Reset()
	gb.mid.Reset()
	for _, v := range gb.front {
		v.Reset()
	}
	gb.shift.Reset()
	gb.shift.anim.srcAlpha = 0
	gb.shift.anim.dstAlpha = 255
	gb.warn.Reset()
}
func (gb *GuardBar) bgDraw(layerno int16) {
	if !sys.lifebar.activeGb {
		return
	}
	gb.bg0.DrawScaled(float32(gb.pos[0])+sys.lifebarOffsetX, float32(gb.pos[1]), layerno, sys.lifebarScale)
	gb.bg1.DrawScaled(float32(gb.pos[0])+sys.lifebarOffsetX, float32(gb.pos[1]), layerno, sys.lifebarScale)
	gb.bg2.DrawScaled(float32(gb.pos[0])+sys.lifebarOffsetX, float32(gb.pos[1]), layerno, sys.lifebarScale)
}
func (gb *GuardBar) draw(layerno int16, ref int, gbr *GuardBar, f []*Fnt) {
	if !sys.lifebar.activeGb {
		return
	}
	power := float32(sys.chars[ref][0].guardPoints) / float32(sys.chars[ref][0].guardPointsMax)
	var MidPos = (float32(sys.gameWidth-320) / 2)
	width := func(power float32) (r [4]int32) {
		r = sys.scrrect
		if gb.range_x[0] < gb.range_x[1] {
			r[0] = int32((((float32(gb.pos[0]+gb.range_x[0])+sys.lifebarOffsetX)*sys.lifebarScale)+MidPos)*sys.widthScale + 0.5)
			r[2] = int32((float32(gb.range_x[1]-gb.range_x[0]+1)*sys.lifebarScale)*power*sys.widthScale + 0.5)
		} else {
			r[2] = int32(((float32(gb.range_x[0]-gb.range_x[1]+1)*sys.lifebarScale)*power-(sys.lifebarOffsetX*sys.lifebarScale))*sys.widthScale + 0.5)
			r[0] = int32(((float32(gb.pos[0]+gb.range_x[0]+1)*sys.lifebarScale)+MidPos)*sys.widthScale+0.5) - r[2]
		}
		return
	}
	pr, mr := width(power), width(gbr.midpower)
	if gb.range_x[0] < gb.range_x[1] {
		mr[0] += pr[2]
	}
	mr[2] -= Min(mr[2], pr[2])
	gb.mid.lay.DrawAnim(&mr, float32(gb.pos[0])+sys.lifebarOffsetX, float32(gb.pos[1]), sys.lifebarScale,
		layerno, &gb.mid.anim, gb.mid.palfx)
	var mv float32
	for k := range gb.front {
		if k > mv && power >= k/100 {
			mv = k
		}
	}
	gb.front[mv].lay.DrawAnim(&pr, float32(gb.pos[0])+sys.lifebarOffsetX, float32(gb.pos[1]), sys.lifebarScale,
		layerno, &gb.front[mv].anim, gb.front[mv].palfx)
	gb.shift.lay.DrawAnim(&pr, float32(gb.pos[0])+sys.lifebarOffsetX, float32(gb.pos[1]), sys.lifebarScale,
		layerno, &gb.shift.anim, gb.shift.palfx)
	if gb.value.font[0] >= 0 && int(gb.value.font[0]) < len(f) && f[gb.value.font[0]] != nil {
		text := strings.Replace(gb.value.text, "%d", fmt.Sprintf("%v", sys.chars[ref][0].guardPoints), 1)
		text = strings.Replace(text, "%p", fmt.Sprintf("%v", math.Round(float64(power)*100)), 1)
		gb.value.lay.DrawText(float32(gb.pos[0])+sys.lifebarOffsetX, float32(gb.pos[1]), sys.lifebarScale,
			layerno, text, f[gb.value.font[0]], gb.value.font[1], gb.value.font[2], gb.value.palfx, gb.value.frgba)
	}
	if power < float32(gb.warn_range[0])/100 && power > float32(gb.warn_range[1])/100 {
		gb.warn.DrawScaled(float32(gb.pos[0])+sys.lifebarOffsetX, float32(gb.pos[1]), layerno, sys.lifebarScale)
	}
	gb.top.DrawScaled(float32(gb.pos[0])+sys.lifebarOffsetX, float32(gb.pos[1]), layerno, sys.lifebarScale)
}

type StunBar struct {
	pos         [2]int32
	range_x     [2]int32
	bg0         AnimLayout
	bg1         AnimLayout
	bg2         AnimLayout
	top         AnimLayout
	mid         AnimLayout
	warn_range  [2]int32
	warn        AnimLayout
	value       LbText
	front       map[float32]*AnimLayout
	shift       AnimLayout
	midpower    float32
	midpowerMin float32
}

func newStunBar() (sb *StunBar) {
	sb = &StunBar{front: make(map[float32]*AnimLayout)}
	return
}
func readStunBar(pre string, is IniSection,
	sff *Sff, at AnimationTable, f []*Fnt) *StunBar {
	sb := newStunBar()
	is.ReadI32(pre+"pos", &sb.pos[0], &sb.pos[1])
	is.ReadI32(pre+"range.x", &sb.range_x[0], &sb.range_x[1])
	sb.bg0 = *ReadAnimLayout(pre+"bg0.", is, sff, at, 0)
	sb.bg1 = *ReadAnimLayout(pre+"bg1.", is, sff, at, 0)
	sb.bg2 = *ReadAnimLayout(pre+"bg2.", is, sff, at, 0)
	sb.top = *ReadAnimLayout(pre+"top.", is, sff, at, 0)
	sb.mid = *ReadAnimLayout(pre+"mid.", is, sff, at, 0)
	sb.front[0] = ReadAnimLayout(pre+"front.", is, sff, at, 0)
	r, _ := regexp.Compile(pre + "front[0-9]+\\.")
	for k := range is {
		if r.MatchString(k) {
			re := regexp.MustCompile("[0-9]+")
			submatchall := re.FindAllString(k, -1)
			if len(submatchall) == 2 {
				v := Atof(submatchall[1])
				if _, ok := sb.front[float32(v)]; !ok {
					sb.front[float32(v)] = ReadAnimLayout(pre+"front"+fmt.Sprintf("%v", v)+".", is, sff, at, 0)
				}
			}
		}
	}
	sb.shift = *ReadAnimLayout(pre+"shift.", is, sff, at, 0)
	sb.value = *readLbText(pre+"value.", is, "%d", 0, f, 0)
	is.ReadI32(pre+"warn.range", &sb.warn_range[0], &sb.warn_range[1])
	sb.warn = *ReadAnimLayout(pre+"warn.", is, sff, at, 0)
	return sb
}
func (sb *StunBar) step(ref int, sbr *StunBar, snd *Snd) {
	if !sys.lifebar.activeSb {
		return
	}
	power := 1 - float32(sys.chars[ref][0].dizzyPoints)/float32(sys.chars[ref][0].dizzyPointsMax)
	sb.shift.anim.srcAlpha = int16(255 * power)
	sb.shift.anim.dstAlpha = int16(255 * (1 - power))
	sbr.midpower -= 1.0 / 144
	if power < sbr.midpowerMin {
		sbr.midpowerMin += (power - sbr.midpowerMin) * (1 / (12 - (power-sbr.midpowerMin)*144))
	} else {
		sbr.midpowerMin = power
	}
	if sbr.midpower < sbr.midpowerMin {
		sbr.midpower = sbr.midpowerMin
	}
	sb.bg0.Action()
	sb.bg1.Action()
	sb.bg2.Action()
	sb.top.Action()
	sb.mid.Action()
	var mv float32
	for k := range sb.front {
		if k > mv && (1-power) >= k/100 {
			mv = k
		}
	}
	sb.front[mv].Action()
	sb.shift.Action()
	sb.warn.Action()
}
func (sb *StunBar) reset() {
	sb.bg0.Reset()
	sb.bg1.Reset()
	sb.bg2.Reset()
	sb.top.Reset()
	sb.mid.Reset()
	for _, v := range sb.front {
		v.Reset()
	}
	sb.shift.Reset()
	sb.shift.anim.srcAlpha = 255
	sb.shift.anim.dstAlpha = 0
	sb.warn.Reset()
}
func (sb *StunBar) bgDraw(layerno int16) {
	if !sys.lifebar.activeSb {
		return
	}
	sb.bg0.DrawScaled(float32(sb.pos[0])+sys.lifebarOffsetX, float32(sb.pos[1]), layerno, sys.lifebarScale)
	sb.bg1.DrawScaled(float32(sb.pos[0])+sys.lifebarOffsetX, float32(sb.pos[1]), layerno, sys.lifebarScale)
	sb.bg2.DrawScaled(float32(sb.pos[0])+sys.lifebarOffsetX, float32(sb.pos[1]), layerno, sys.lifebarScale)
}
func (sb *StunBar) draw(layerno int16, ref int, sbr *StunBar, f []*Fnt) {
	if !sys.lifebar.activeSb {
		return
	}
	power := 1 - float32(sys.chars[ref][0].dizzyPoints)/float32(sys.chars[ref][0].dizzyPointsMax)
	var MidPos = (float32(sys.gameWidth-320) / 2)
	width := func(power float32) (r [4]int32) {
		r = sys.scrrect
		if sb.range_x[0] < sb.range_x[1] {
			r[0] = int32((((float32(sb.pos[0]+sb.range_x[0])+sys.lifebarOffsetX)*sys.lifebarScale)+MidPos)*sys.widthScale + 0.5)
			r[2] = int32((float32(sb.range_x[1]-sb.range_x[0]+1)*sys.lifebarScale)*power*sys.widthScale + 0.5)
		} else {
			r[2] = int32(((float32(sb.range_x[0]-sb.range_x[1]+1)*sys.lifebarScale)*power-(sys.lifebarOffsetX*sys.lifebarScale))*sys.widthScale + 0.5)
			r[0] = int32(((float32(sb.pos[0]+sb.range_x[0]+1)*sys.lifebarScale)+MidPos)*sys.widthScale+0.5) - r[2]
		}
		return
	}
	pr, mr := width(power), width(sbr.midpower)
	if sb.range_x[0] < sb.range_x[1] {
		mr[0] += pr[2]
	}
	mr[2] -= Min(mr[2], pr[2])
	sb.mid.lay.DrawAnim(&mr, float32(sb.pos[0])+sys.lifebarOffsetX, float32(sb.pos[1]), sys.lifebarScale,
		layerno, &sb.mid.anim, sb.mid.palfx)
	var mv float32
	for k := range sb.front {
		if k > mv && power >= k/100 {
			mv = k
		}
	}
	sb.front[mv].lay.DrawAnim(&pr, float32(sb.pos[0])+sys.lifebarOffsetX, float32(sb.pos[1]), sys.lifebarScale,
		layerno, &sb.front[mv].anim, sb.front[mv].palfx)
	sb.shift.lay.DrawAnim(&pr, float32(sb.pos[0])+sys.lifebarOffsetX, float32(sb.pos[1]), sys.lifebarScale,
		layerno, &sb.shift.anim, sb.shift.palfx)
	if sb.value.font[0] >= 0 && int(sb.value.font[0]) < len(f) && f[sb.value.font[0]] != nil {
		text := strings.Replace(sb.value.text, "%d", fmt.Sprintf("%v", sys.chars[ref][0].dizzyPoints), 1)
		text = strings.Replace(text, "%p", fmt.Sprintf("%v", math.Round(float64(power)*100)), 1)
		sb.value.lay.DrawText(float32(sb.pos[0])+sys.lifebarOffsetX, float32(sb.pos[1]), sys.lifebarScale,
			layerno, text, f[sb.value.font[0]], sb.value.font[1], sb.value.font[2], sb.value.palfx, sb.value.frgba)
	}
	if power > float32(sb.warn_range[0])/100 && power < float32(sb.warn_range[1])/100 {
		sb.warn.DrawScaled(float32(sb.pos[0])+sys.lifebarOffsetX, float32(sb.pos[1]), layerno, sys.lifebarScale)
	}
	sb.top.DrawScaled(float32(sb.pos[0])+sys.lifebarOffsetX, float32(sb.pos[1]), layerno, sys.lifebarScale)
}

type LifeBarFace struct {
	pos               [2]int32
	bg                AnimLayout
	bg0               AnimLayout
	bg1               AnimLayout
	bg2               AnimLayout
	top               AnimLayout
	ko                AnimLayout
	face_spr          [2]int32
	face              *Sprite
	face_lay          Layout
	palshare          bool
	palfxshare        bool
	teammate_pos      [2]int32
	teammate_spacing  [2]int32
	teammate_bg       AnimLayout
	teammate_bg0      AnimLayout
	teammate_bg1      AnimLayout
	teammate_bg2      AnimLayout
	teammate_top      AnimLayout
	teammate_ko       AnimLayout
	teammate_face_spr [2]int32
	teammate_face     []*Sprite
	teammate_face_lay Layout
	teammate_scale    []float32
	numko             int32
	old_spr, old_pal  [2]int32
}

func newLifeBarFace() *LifeBarFace {
	return &LifeBarFace{face_spr: [2]int32{-1}, teammate_face_spr: [2]int32{-1}, palshare: true}
}
func readLifeBarFace(pre string, is IniSection,
	sff *Sff, at AnimationTable) *LifeBarFace {
	fa := newLifeBarFace()
	is.ReadI32(pre+"pos", &fa.pos[0], &fa.pos[1])

	fa.bg = *ReadAnimLayout(pre+"bg.", is, sff, at, 0)
	fa.bg0 = *ReadAnimLayout(pre+"bg0.", is, sff, at, 0)
	fa.bg1 = *ReadAnimLayout(pre+"bg1.", is, sff, at, 0)
	fa.bg2 = *ReadAnimLayout(pre+"bg2.", is, sff, at, 0)
	fa.top = *ReadAnimLayout(pre+"top.", is, sff, at, 0)
	fa.ko = *ReadAnimLayout(pre+"ko.", is, sff, at, 0)
	is.ReadI32(pre+"face.spr", &fa.face_spr[0], &fa.face_spr[1])
	fa.face_lay = *ReadLayout(pre+"face.", is, 0)
	is.ReadBool(pre+"face.palshare", &fa.palshare)
	is.ReadBool(pre+"face.palfxshare", &fa.palfxshare)
	is.ReadI32(pre+"teammate.pos", &fa.teammate_pos[0], &fa.teammate_pos[1])
	is.ReadI32(pre+"teammate.spacing", &fa.teammate_spacing[0],
		&fa.teammate_spacing[1])
	fa.teammate_bg = *ReadAnimLayout(pre+"teammate.bg.", is, sff, at, 0)
	fa.teammate_bg0 = *ReadAnimLayout(pre+"teammate.bg0.", is, sff, at, 0)
	fa.teammate_bg1 = *ReadAnimLayout(pre+"teammate.bg1.", is, sff, at, 0)
	fa.teammate_bg2 = *ReadAnimLayout(pre+"teammate.bg2.", is, sff, at, 0)
	fa.teammate_top = *ReadAnimLayout(pre+"teammate.top.", is, sff, at, 0)
	fa.teammate_ko = *ReadAnimLayout(pre+"teammate.ko.", is, sff, at, 0)
	is.ReadI32(pre+"teammate.face.spr", &fa.teammate_face_spr[0],
		&fa.teammate_face_spr[1])
	if fa.teammate_face_spr[0] != -1 {
		sys.sel.charSpritePreload[[...]int16{int16(fa.teammate_face_spr[0]), int16(fa.teammate_face_spr[1])}] = true
	}
	fa.teammate_face_lay = *ReadLayout(pre+"teammate.face.", is, 0)
	return fa
}
func (fa *LifeBarFace) step() {
	fa.bg.Action()
	fa.bg0.Action()
	fa.bg1.Action()
	fa.bg2.Action()
	fa.top.Action()
	fa.ko.Action()
	fa.teammate_bg.Action()
	fa.teammate_bg0.Action()
	fa.teammate_bg1.Action()
	fa.teammate_bg2.Action()
	fa.teammate_top.Action()
	fa.teammate_ko.Action()
}
func (fa *LifeBarFace) reset() {
	fa.bg.Reset()
	fa.bg0.Reset()
	fa.bg1.Reset()
	fa.bg2.Reset()
	fa.top.Reset()
	fa.ko.Reset()
	fa.teammate_bg.Reset()
	fa.teammate_bg0.Reset()
	fa.teammate_bg1.Reset()
	fa.teammate_bg2.Reset()
	fa.teammate_top.Reset()
	fa.teammate_ko.Reset()
	if !sys.roundResetFlg {
		fa.old_spr = [2]int32{}
		fa.old_pal = [2]int32{}
	}
}
func (fa *LifeBarFace) bgDraw(layerno int16) {
	fa.bg.DrawScaled(float32(fa.pos[0])+sys.lifebarOffsetX, float32(fa.pos[1]), layerno, sys.lifebarScale)
	fa.bg0.DrawScaled(float32(fa.pos[0])+sys.lifebarOffsetX, float32(fa.pos[1]), layerno, sys.lifebarScale)
	fa.bg1.DrawScaled(float32(fa.pos[0])+sys.lifebarOffsetX, float32(fa.pos[1]), layerno, sys.lifebarScale)
	fa.bg2.DrawScaled(float32(fa.pos[0])+sys.lifebarOffsetX, float32(fa.pos[1]), layerno, sys.lifebarScale)
}
func (fa *LifeBarFace) draw(layerno int16, ref int, far *LifeBarFace) {
	group, number := int16(fa.face_spr[0]), int16(fa.face_spr[1])
	if mg, ok := sys.chars[ref][0].anim.remap[group]; ok {
		if mn, ok := mg[number]; ok {
			group, number = mn[0], mn[1]
		}
	}
	if far.old_spr[0] != int32(group) || far.old_spr[1] != int32(number) ||
		far.old_pal[0] != sys.cgi[ref].remappedpal[0] || far.old_pal[1] != sys.cgi[ref].remappedpal[1] {
		far.face = sys.cgi[ref].sff.getOwnPalSprite(group, number)
		far.old_spr = [...]int32{int32(group), int32(number)}
		far.old_pal = [...]int32{sys.cgi[ref].remappedpal[0], sys.cgi[ref].remappedpal[1]}
	}
	if far.face != nil {
		far.face.Pal = nil
		far.face.Pal = far.face.GetPal(&sys.cgi[ref].sff.palList)
		pfx := newPalFX()
		if far.palfxshare {
			pfx = sys.chars[ref][0].getPalfx()
		}
		if far.palshare {
			sys.cgi[ref].sff.palList.SwapPalMap(&sys.chars[ref][0].getPalfx().remap)
		}

		ob := sys.brightness
		if ref == sys.superplayer {
			sys.brightness = 256
		}
		fa.face_lay.DrawSprite((float32(fa.pos[0])+sys.lifebarOffsetX)*sys.lifebarScale, float32(fa.pos[1])*sys.lifebarScale, layerno,
			far.face, pfx, sys.cgi[ref].portraitscale*sys.lifebarPortraitScale, &fa.face_lay.window)
		if !sys.chars[ref][0].alive() {
			fa.ko.DrawScaled(float32(fa.pos[0])+sys.lifebarOffsetX, float32(fa.pos[1]), layerno, sys.lifebarScale)
		}
		sys.brightness = ob
		i := int32(len(far.teammate_face)) - 1
		x := float32(fa.teammate_pos[0] + fa.teammate_spacing[0]*(i-1))
		y := float32(fa.teammate_pos[1] + fa.teammate_spacing[1]*(i-1))
		for ; i >= 0; i-- {
			if i != fa.numko {
				fa.teammate_bg.DrawScaled((x + sys.lifebarOffsetX), y, layerno, sys.lifebarScale)
				fa.teammate_bg0.DrawScaled((x + sys.lifebarOffsetX), y, layerno, sys.lifebarScale)
				fa.teammate_bg1.DrawScaled((x + sys.lifebarOffsetX), y, layerno, sys.lifebarScale)
				fa.teammate_bg2.DrawScaled((x + sys.lifebarOffsetX), y, layerno, sys.lifebarScale)
				fa.teammate_face_lay.DrawSprite((x+sys.lifebarOffsetX)*sys.lifebarScale, y*sys.lifebarScale, layerno, far.teammate_face[i], nil, far.teammate_scale[i]*sys.lifebarPortraitScale, &fa.teammate_face_lay.window)
				if i < fa.numko {
					fa.teammate_ko.DrawScaled((x + sys.lifebarOffsetX), y, layerno, sys.lifebarScale)
				}
				x -= float32(fa.teammate_spacing[0])
				y -= float32(fa.teammate_spacing[1])
			}
		}
	}
	fa.top.DrawScaled(float32(fa.pos[0])+sys.lifebarOffsetX, float32(fa.pos[1]), layerno, sys.lifebarScale)
}

type LifeBarName struct {
	pos              [2]int32
	name             LbText
	bg               AnimLayout
	top              AnimLayout
	teammate_pos     [2]int32
	teammate_spacing [2]int32
	teammate_name    LbText
	teammate_bg      AnimLayout
	numko            int32
}

func newLifeBarName() *LifeBarName {
	return &LifeBarName{}
}
func readLifeBarName(pre string, is IniSection,
	sff *Sff, at AnimationTable, f []*Fnt) *LifeBarName {
	nm := newLifeBarName()
	is.ReadI32(pre+"pos", &nm.pos[0], &nm.pos[1])
	nm.name = *readLbText(pre+"name.", is, "", 0, f, 0)
	nm.bg = *ReadAnimLayout(pre+"bg.", is, sff, at, 0)
	is.ReadI32(pre+"teammate.pos", &nm.teammate_pos[0], &nm.teammate_pos[1])
	is.ReadI32(pre+"teammate.spacing", &nm.teammate_spacing[0],
		&nm.teammate_spacing[1])
	nm.teammate_name = *readLbText(pre+"teammate.name.", is, "", 0, f, 0)
	nm.teammate_bg = *ReadAnimLayout(pre+"teammate.bg.", is, sff, at, 0)
	nm.top = *ReadAnimLayout(pre+"top.", is, sff, at, 0)
	return nm
}
func (nm *LifeBarName) step() {
	nm.bg.Action()
	nm.teammate_bg.Action()
	nm.top.Action()
}
func (nm *LifeBarName) reset() {
	nm.bg.Reset()
	nm.teammate_bg.Reset()
	nm.top.Reset()
}
func (nm *LifeBarName) bgDraw(layerno int16) {
	nm.bg.DrawScaled(float32(nm.pos[0])+sys.lifebarOffsetX, float32(nm.pos[1]), layerno, sys.lifebarScale)
}
func (nm *LifeBarName) draw(layerno int16, ref int, f []*Fnt, side int) {
	if nm.name.font[0] >= 0 && int(nm.name.font[0]) < len(f) && f[nm.name.font[0]] != nil {
		nm.name.lay.DrawText((float32(nm.pos[0]) + sys.lifebarOffsetX), float32(nm.pos[1]), sys.lifebarScale, layerno,
			sys.cgi[ref].lifebarname, f[nm.name.font[0]], nm.name.font[1], nm.name.font[2], nm.name.palfx, nm.name.frgba)
	}
	if sys.tmode[side] == TM_Turns {
		i := int32(len(sys.sel.selected[side])) - 1
		x := float32(nm.teammate_pos[0] + nm.teammate_spacing[0]*(i-nm.numko-1))
		y := float32(nm.teammate_pos[1] + nm.teammate_spacing[1]*(i-nm.numko-1))
		for ; i >= nm.numko+1; i-- {
			nm.teammate_bg.DrawScaled((x + sys.lifebarOffsetX), y, layerno, sys.lifebarScale)
			if nm.teammate_name.font[0] >= 0 && int(nm.teammate_name.font[0]) < len(f) && f[nm.teammate_name.font[0]] != nil {
				nm.teammate_name.lay.DrawText((float32(x) + sys.lifebarOffsetX), float32(y), sys.lifebarScale, layerno,
					sys.sel.GetChar(sys.sel.selected[side][i][0]).lifebarname, f[nm.teammate_name.font[0]], nm.teammate_name.font[1],
					nm.teammate_name.font[2], nm.teammate_name.palfx, nm.teammate_name.frgba)
			}
			x -= float32(nm.teammate_spacing[0])
			y -= float32(nm.teammate_spacing[1])
		}
	}
	nm.top.DrawScaled(float32(nm.pos[0])+sys.lifebarOffsetX, float32(nm.pos[1]), layerno, sys.lifebarScale)
}

type LifeBarWinIcon struct {
	pos           [2]int32
	iconoffset    [2]int32
	useiconupto   int32
	counter       LbText
	bg0           AnimLayout
	top           AnimLayout
	icon          [WT_NumTypes]AnimLayout
	wins          []WinType
	numWins       int
	added, addedP *Animation
}

func newLifeBarWinIcon() *LifeBarWinIcon {
	return &LifeBarWinIcon{useiconupto: 4}
}
func readLifeBarWinIcon(pre string, is IniSection,
	sff *Sff, at AnimationTable, f []*Fnt) *LifeBarWinIcon {
	wi := newLifeBarWinIcon()
	is.ReadI32(pre+"pos", &wi.pos[0], &wi.pos[1])
	is.ReadI32(pre+"iconoffset", &wi.iconoffset[0], &wi.iconoffset[1])
	is.ReadI32("useiconupto", &wi.useiconupto)
	wi.counter = *readLbText(pre+"counter.", is, "%i", 0, f, 0)
	wi.bg0 = *ReadAnimLayout(pre+"bg0.", is, sff, at, 0)
	wi.top = *ReadAnimLayout(pre+"top.", is, sff, at, 0)
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
	wi.top.Action()
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
	wi.top.Reset()
	for i := range wi.icon {
		wi.icon[i].Reset()
	}
	wi.numWins = len(wi.wins)
	wi.added, wi.addedP = nil, nil
}
func (wi *LifeBarWinIcon) clear() { wi.wins = nil }
func (wi *LifeBarWinIcon) draw(layerno int16, f []*Fnt, side int) {
	bg0num := float64(sys.lifebar.ro.match_wins[^side&1])
	if sys.tmode[^side&1] == TM_Turns {
		bg0num = float64(sys.numTurns[^side&1])
	}
	for i := 0; i < int(math.Min(float64(wi.useiconupto), bg0num)); i++ {
		wi.bg0.DrawScaled(float32(wi.pos[0]+wi.iconoffset[0]*int32(i))+sys.lifebarOffsetX,
			float32(wi.pos[1]+wi.iconoffset[1]*int32(i)), layerno, sys.lifebarScale)
	}
	if len(wi.wins) > int(wi.useiconupto) {
		if wi.counter.font[0] >= 0 && int(wi.counter.font[0]) < len(f) && f[wi.counter.font[0]] != nil {
			wi.counter.lay.DrawText(float32(wi.pos[0])+sys.lifebarOffsetX, float32(wi.pos[1]), sys.lifebarScale,
				layerno, strings.Replace(wi.counter.text, "%i", fmt.Sprintf("%v", len(wi.wins)), 1),
				f[wi.counter.font[0]], wi.counter.font[1], wi.counter.font[2], wi.counter.palfx, wi.counter.frgba)
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
			wi.icon[wt].lay.DrawAnim(&wi.icon[wt].lay.window,
				float32(wi.pos[0]+wi.iconoffset[0]*int32(i))+sys.lifebarOffsetX,
				float32(wi.pos[1]+wi.iconoffset[1]*int32(i)), sys.lifebarScale, layerno, wi.added, nil)
			if p {
				wi.icon[WT_Perfect].lay.DrawAnim(&wi.icon[WT_Perfect].lay.window,
					float32(wi.pos[0]+wi.iconoffset[0]*int32(i))+sys.lifebarOffsetX,
					float32(wi.pos[1]+wi.iconoffset[1]*int32(i)), sys.lifebarScale, layerno, wi.addedP, nil)
			}
		}
	}
	for i := 0; i < int(math.Min(float64(wi.useiconupto), bg0num)); i++ {
		wi.top.DrawScaled(float32(wi.pos[0]+wi.iconoffset[0]*int32(i))+sys.lifebarOffsetX,
			float32(wi.pos[1]+wi.iconoffset[1]*int32(i)), layerno, sys.lifebarScale)
	}
}

type LifeBarTime struct {
	pos            [2]int32
	counter        map[int32]*LbText
	bg             AnimLayout
	top            AnimLayout
	framespercount int32
}

func newLifeBarTime() *LifeBarTime {
	return &LifeBarTime{counter: make(map[int32]*LbText), framespercount: 60}
}
func readLifeBarTime(is IniSection,
	sff *Sff, at AnimationTable, f []*Fnt) *LifeBarTime {
	ti := newLifeBarTime()
	is.ReadI32("pos", &ti.pos[0], &ti.pos[1])
	ti.counter[0] = readLbText("counter.", is, "", 0, f, 0)
	r, _ := regexp.Compile(`counter[0-9]+\.`)
	for k := range is {
		if r.MatchString(k) {
			re := regexp.MustCompile("[0-9]+")
			submatchall := re.FindAllString(k, -1)
			if len(submatchall) == 1 {
				v := Atoi(submatchall[0])
				if _, ok := ti.counter[v]; !ok {
					ti.counter[v] = readLbText("counter"+fmt.Sprintf("%v", v)+".", is, "", 0, f, 0)
				}
			}
		}
	}
	ti.bg = *ReadAnimLayout("bg.", is, sff, at, 0)
	ti.top = *ReadAnimLayout("top.", is, sff, at, 0)
	is.ReadI32("framespercount", &ti.framespercount)
	return ti
}
func (ti *LifeBarTime) step() {
	ti.bg.Action()
	ti.top.Action()
}
func (ti *LifeBarTime) reset() {
	ti.bg.Reset()
	ti.top.Reset()
}
func (ti *LifeBarTime) bgDraw(layerno int16) {
	ti.bg.DrawScaled(float32(ti.pos[0])+sys.lifebarOffsetX, float32(ti.pos[1]), layerno, sys.lifebarScale)
}
func (ti *LifeBarTime) draw(layerno int16, f []*Fnt) {
	if ti.framespercount > 0 &&
		ti.counter[0].font[0] >= 0 && int(ti.counter[0].font[0]) < len(f) && f[ti.counter[0].font[0]] != nil {
		var timeval int32 = -1
		time := "o"
		if sys.time >= 0 {
			timeval = int32(math.Ceil(float64(sys.time) / float64(ti.framespercount)))
			time = fmt.Sprintf("%v", timeval)
		}
		var tv int32
		for k := range ti.counter {
			if k > tv && timeval >= k {
				tv = k
			}
		}
		ti.counter[tv].lay.DrawText(float32(ti.pos[0])+sys.lifebarOffsetX, float32(ti.pos[1]), sys.lifebarScale, layerno,
			time, f[ti.counter[tv].font[0]], ti.counter[tv].font[1], ti.counter[tv].font[2], ti.counter[tv].palfx,
			ti.counter[tv].frgba)
	}
	ti.top.DrawScaled(float32(ti.pos[0])+sys.lifebarOffsetX, float32(ti.pos[1]), layerno, sys.lifebarScale)
}

type LifeBarCombo struct {
	pos           [2]int32
	start_x       float32
	counter       LbText
	counter_shake bool
	counter_time  int32
	counter_mult  float32
	text          LbText
	displaytime   int32
	showspeed     float32
	hidespeed     float32
	separator     string
	places        int32
	cur, old      int32
	curd, oldd    int32
	curp, oldp    float32
	resttime      int32
	counterX      float32
	shaketime     int32
	combo         int32
	fakeCombo     int32
}

func newLifeBarCombo() *LifeBarCombo {
	return &LifeBarCombo{displaytime: 90, showspeed: 8, hidespeed: 4,
		counter_time: 7, counter_mult: 1.0 / 20}
}
func readLifeBarCombo(pre string, is IniSection, f []*Fnt) *LifeBarCombo {
	co := newLifeBarCombo()
	is.ReadI32(pre+"pos", &co.pos[0], &co.pos[1])
	is.ReadF32(pre+"start.x", &co.start_x)
	if pre == "team2." { //mugen 1.0 implementation reuses winmugen code where both sides shared the same values
		co.pos[0] = sys.lifebarLocalcoord[0] - co.pos[0]
		co.start_x = float32(sys.lifebarLocalcoord[0]) - co.start_x
	}
	co.counter = *readLbText(pre+"counter.", is, "%i", 2, f, 0)
	is.ReadBool(pre+"counter.shake", &co.counter_shake)
	is.ReadI32(pre+"counter.time", &co.counter_time)
	is.ReadF32(pre+"counter.mult", &co.counter_mult)
	if pre == "team2." {
		co.text = *readLbText(pre+"text.", is, "", 2, f, -1)
	} else {
		co.text = *readLbText(pre+"text.", is, "", 2, f, 1)
	}
	is.ReadI32(pre+"displaytime", &co.displaytime)
	is.ReadF32(pre+"showspeed", &co.showspeed)
	co.showspeed = MaxF(1, co.showspeed)
	is.ReadF32(pre+"hidespeed", &co.hidespeed)
	co.separator, _ = is.getString("format.decimal.separator")
	is.ReadI32("format.decimal.places", &co.places)
	return co
}
func (co *LifeBarCombo) step(combo, fakeCombo, damage int32, percentage float32, dizzy bool) {
	if fakeCombo > 0 {
		fakeCombo = co.fakeCombo
	} else {
		co.fakeCombo = 0
	}
	if combo <= 0 {
		co.combo = 0
	}

	if co.resttime > 0 {
		co.counterX -= co.counterX / co.showspeed
	} else if fakeCombo < 2 {
		co.counterX -= sys.lifebarFontScale * co.hidespeed * float32(sys.lifebarLocalcoord[0]) / 320
		if co.counterX < co.start_x*2 {
			co.counterX = co.start_x * 2
		}
	}
	if co.shaketime > 0 {
		co.shaketime--
	}
	if AbsF(co.counterX) < 1 && !dizzy {
		co.resttime--
	}
	if fakeCombo >= 2 {
		if co.old != fakeCombo {
			co.cur = fakeCombo
			co.resttime = co.displaytime
			if co.counter_shake {
				co.shaketime = co.counter_time
			}
		}
		if co.oldd != damage {
			co.curd = damage
		}
		if co.oldp != percentage {
			co.curp = percentage
		}
	}
	co.old = fakeCombo
	co.oldd = damage
	co.oldp = percentage
}
func (co *LifeBarCombo) reset() {
	co.cur, co.old, co.curd, co.oldd, co.curp, co.oldp, co.resttime = 0, 0, 0, 0, 0, 0, 0
	co.combo = 0
	co.fakeCombo = 0
	co.counterX = co.start_x * 2
	co.shaketime = 0
}
func (co *LifeBarCombo) draw(layerno int16, f []*Fnt, side int) {
	haba := func(str string) float32 {
		if co.counter.font[0] < 0 || int(co.counter.font[0]) >= len(f) || f[co.counter.font[0]] == nil {
			return 0
		}
		return float32(f[co.counter.font[0]].TextWidth(str)) *
			co.counter.lay.scale[0] * sys.lifebarFontScale
	}
	if co.resttime <= 0 && co.counterX == co.start_x*2 {
		return
	}
	counter := strings.Replace(co.counter.text, "%i", fmt.Sprintf("%v", co.cur), 1)
	var x float32
	if side == 0 {
		if co.start_x <= 0 {
			x = co.counterX
		}
		x += float32(co.pos[0]) + haba(counter)
	} else {
		if co.start_x <= 0 {
			x = -co.counterX
		}
		x += 320/sys.lifebarScale - sys.lifebarOffsetX*2 - float32(co.pos[0])
	}
	if co.text.font[0] >= 0 && int(co.text.font[0]) < len(f) && f[co.text.font[0]] != nil {
		text := strings.Replace(co.text.text, "%i", fmt.Sprintf("%v", co.cur), 1)
		text = strings.Replace(text, "%d", fmt.Sprintf("%v", co.curd), 1)
		//split float value, round to decimal place
		s := strings.Split(fmt.Sprintf("%.[2]*[1]f", co.curp, co.places), ".")
		//decimal separator
		if co.places > 0 {
			if len(s) > 1 {
				s[0] = s[0] + co.separator + s[1]
			}
		}
		//replace %p with formatted string
		text = strings.Replace(text, "%p", s[0], 1)
		//split on new line
		lines := strings.Split(text, "\\n")

		if side == 0 {
			if co.pos[0] != 0 {
				x += co.text.lay.offset[0] *
					((1 - sys.lifebarFontScale) * sys.lifebarFontScale)
			}
		} else {
			var length float32
			for _, v := range lines {
				if lt := float32(f[co.text.font[0]].TextWidth(v)) * co.text.lay.scale[0] * sys.lifebarFontScale; lt > length {
					length = lt
				}
			}
			x -= length
			/*tmp := c.text_lay.offset[0]
			if c.pos[0] == 0 {
				tmp *= sys.lifebarFontScale
			}
			x -= tmp*/
		}
		for k, v := range lines {
			co.text.lay.DrawText(x+sys.lifebarOffsetX, float32(co.pos[1])+
				float32(k)*(float32(f[co.text.font[0]].Size[1])*co.text.lay.scale[1]*sys.lifebarFontScale+
					float32(f[co.text.font[0]].Spacing[1])*co.text.lay.scale[1]*sys.lifebarFontScale),
				sys.lifebarScale, layerno, v, f[co.text.font[0]], co.text.font[1], 1, co.text.palfx,
				co.text.frgba)
		}
	}
	if co.counter.font[0] >= 0 && int(co.counter.font[0]) < len(f) && f[co.counter.font[0]] != nil {
		z := 1 + float32(co.shaketime)*co.counter_mult*
			float32(math.Sin(float64(co.shaketime)*(math.Pi/2.5)))
		co.counter.lay.DrawText((x+sys.lifebarOffsetX)/z, float32(co.pos[1])/z, z*sys.lifebarScale, layerno,
			counter, f[co.counter.font[0]], co.counter.font[1], -1, co.counter.palfx, co.counter.frgba)
	}
}

type LbMsg struct {
	resttime int32
	counterX float32
	text     string
	bg       AnimLayout
	front    AnimLayout
	del      bool
}

func newLbMsg(text string, time int32, side int) *LbMsg {
	return &LbMsg{resttime: time, counterX: sys.lifebar.ac[side].start_x * 2, text: text}
}
func insertLbMsg(array []*LbMsg, value *LbMsg, index int) []*LbMsg {
	return append(array[:index], append([]*LbMsg{value}, array[index:]...)...)
}
func removeLbMsg(array []*LbMsg, index int) []*LbMsg {
	return append(array[:index], array[index+1:]...)
}

type LifeBarAction struct {
	oldleader   int
	pos         [2]int32
	spacing     [2]int32
	start_x     float32
	text        LbText
	displaytime int32
	showspeed   float32
	hidespeed   float32
	messages    []*LbMsg
	is          IniSection
}

func newLifeBarAction() *LifeBarAction {
	return &LifeBarAction{displaytime: 90, showspeed: 8, hidespeed: 4, is: make(map[string]string)}
}
func readLifeBarAction(pre string, is IniSection, f []*Fnt) *LifeBarAction {
	ac := newLifeBarAction()
	ac.is = is
	is.ReadI32(pre+"pos", &ac.pos[0], &ac.pos[1])
	is.ReadI32(pre+"spacing", &ac.spacing[0], &ac.spacing[1])
	is.ReadF32(pre+"start.x", &ac.start_x)
	if pre == "team2." {
		ac.pos[0] = sys.lifebarLocalcoord[0] - ac.pos[0]
		ac.start_x = float32(sys.lifebarLocalcoord[0]) - ac.start_x
	}
	if pre == "team2." {
		ac.text = *readLbText(pre+"text.", is, "", 2, f, -1)
	} else {
		ac.text = *readLbText(pre+"text.", is, "", 2, f, 1)
	}
	is.ReadI32(pre+"displaytime", &ac.displaytime)
	is.ReadF32(pre+"showspeed", &ac.showspeed)
	ac.showspeed = MaxF(1, ac.showspeed)
	is.ReadF32(pre+"hidespeed", &ac.hidespeed)
	return ac
}
func (ac *LifeBarAction) step(leader int) {
	if ac.oldleader != leader {
		ac.oldleader = leader
	}
	for _, v := range ac.messages {
		v.bg.Action()
		v.front.Action()
		if v.resttime > 0 {
			v.counterX -= v.counterX / ac.showspeed
		} else {
			v.counterX -= sys.lifebarFontScale * ac.hidespeed * float32(sys.lifebarLocalcoord[0]) / 320
			if v.counterX < ac.start_x*2 {
				v.del = true
			}
		}
		if AbsF(v.counterX) < 1 {
			v.resttime--
		}
	}
	if len(ac.messages) > 0 && ac.messages[len(ac.messages)-1].del {
		ac.messages = removeLbMsg(ac.messages, len(ac.messages)-1)
	}
}
func (ac *LifeBarAction) reset(leader int) {
	ac.oldleader = leader
	ac.messages = []*LbMsg{}
}
func (ac *LifeBarAction) draw(layerno int16, f []*Fnt, side int) {
	for k, v := range ac.messages {
		if v.resttime <= 0 && v.counterX == ac.start_x*2 {
			continue
		}
		var x float32
		if side == 0 {
			if ac.start_x <= 0 {
				x = v.counterX
			}
			x += float32(ac.pos[0])
		} else {
			if ac.start_x <= 0 {
				x = -v.counterX
			}
			x += 320/sys.lifebarScale - sys.lifebarOffsetX*2 - float32(ac.pos[0])
		}
		if v.bg.anim.spr != nil {
			v.bg.DrawScaled(x+sys.lifebarOffsetX+float32(k)*float32(ac.spacing[0]),
				float32(ac.pos[1])+float32(k)*float32(ac.spacing[1])+
					float32(k)*(float32(v.bg.anim.spr.Size[1])*v.bg.lay.scale[1]),
				layerno, sys.lifebarScale)
		}
		if v.front.anim.spr != nil {
			v.front.DrawScaled(x+sys.lifebarOffsetX+float32(k)*float32(ac.spacing[0]),
				float32(ac.pos[1])+float32(k)*float32(ac.spacing[1])+
					float32(k)*(float32(v.front.anim.spr.Size[1])*v.front.lay.scale[1]),
				layerno, sys.lifebarScale)
		}
		if ac.text.font[0] >= 0 && int(ac.text.font[0]) < len(f) && f[ac.text.font[0]] != nil {
			if side == 0 {
				if ac.pos[0] != 0 {
					x += ac.text.lay.offset[0] *
						((1 - sys.lifebarFontScale) * sys.lifebarFontScale)
				}
			} else {
				x -= float32(f[ac.text.font[0]].TextWidth(v.text)) *
					ac.text.lay.scale[0] * sys.lifebarFontScale
				/*tmp := ac.text.lay.offset[0]
				if ac.pos[0] == 0 {
					tmp *= sys.lifebarFontScale
				}
				x -= tmp*/
			}
			ac.text.lay.DrawText(x+sys.lifebarOffsetX+float32(k)*float32(ac.spacing[0])*sys.lifebarFontScale,
				float32(ac.pos[1])+float32(k)*float32(ac.spacing[1])*sys.lifebarFontScale+
					float32(k)*(float32(f[ac.text.font[0]].Size[1])*ac.text.lay.scale[1]*sys.lifebarFontScale+
						float32(f[ac.text.font[0]].Spacing[1])*ac.text.lay.scale[1]*sys.lifebarFontScale),
				sys.lifebarScale, layerno, v.text, f[ac.text.font[0]], ac.text.font[1], 1, ac.text.palfx,
				ac.text.frgba)
		}
	}
}

type LifeBarRound struct {
	snd                *Snd
	pos                [2]int32
	match_wins         [2]int32
	match_maxdrawgames [2]int32
	start_waittime     int32
	round_time         int32
	round_sndtime      int32
	round_default      AnimTextSnd
	round              [9]AnimTextSnd
	round_default_top  AnimLayout
	round_default_bg   [32]AnimLayout
	round_final        AnimTextSnd
	round_final_top    AnimLayout
	round_final_bg     [32]AnimLayout
	fight_time         int32
	fight_sndtime      int32
	fight              AnimTextSnd
	fight_top          AnimLayout
	fight_bg           [32]AnimLayout
	ctrl_time          int32
	ko_time            int32
	ko_sndtime         int32
	ko, dko, to        AnimTextSnd
	ko_top             AnimLayout
	ko_bg              [32]AnimLayout
	dko_top            AnimLayout
	dko_bg             [32]AnimLayout
	to_top             AnimLayout
	to_bg              [32]AnimLayout
	slow_time          int32
	slow_fadetime      int32
	slow_speed         float32
	over_waittime      int32
	over_hittime       int32
	over_wintime       int32
	over_time          int32
	win_time           int32
	win_sndtime        int32
	win, win2, drawn   AnimTextSnd
	win_top            AnimLayout
	win_bg             [32]AnimLayout
	win2_top           AnimLayout
	win2_bg            [32]AnimLayout
	drawn_top          AnimLayout
	drawn_bg           [32]AnimLayout
	win3, win4         AnimTextSnd
	win3_top           AnimLayout
	win3_bg            [32]AnimLayout
	win4_top           AnimLayout
	win4_bg            [32]AnimLayout
	cur                int32
	wt, swt, dt        [4]int32
	timerActive        bool
	wint               [WT_NumTypes * 2]LbBgTextSnd
	fadein_time        int32
	fadein_col         uint32
	fadeout_time       int32
	fadeout_col        uint32
	shutter_time       int32
	shutter_col        uint32
	callfight_time     int32
	introState         [2]bool
	firstAttack        [2]bool
}

func newLifeBarRound(snd *Snd) *LifeBarRound {
	return &LifeBarRound{snd: snd, match_wins: [...]int32{2, 2},
		match_maxdrawgames: [...]int32{1, 1}, start_waittime: 30, fight_sndtime: 100,
		ctrl_time: 30, slow_time: 60, slow_fadetime: 45, slow_speed: 0.25,
		over_waittime: 45, over_hittime: 10, over_wintime: 45, over_time: 210,
		win_sndtime: 60, fadein_time: 30, fadeout_time: 30, shutter_time: 15,
		callfight_time: 60}
}
func readLifeBarRound(is IniSection,
	sff *Sff, at AnimationTable, snd *Snd, f []*Fnt) *LifeBarRound {
	ro := newLifeBarRound(snd)
	var tmp int32
	var ftmp float32
	is.ReadI32("pos", &ro.pos[0], &ro.pos[1])
	is.ReadI32("match.wins", &ro.match_wins[0], &ro.match_wins[1])
	is.ReadI32("match.maxdrawgames", &ro.match_maxdrawgames[0], &ro.match_maxdrawgames[1])
	if is.ReadI32("start.waittime", &tmp) {
		ro.start_waittime = Max(1, tmp)
	}
	is.ReadI32("round.time", &ro.round_time)
	is.ReadI32("round.sndtime", &ro.round_sndtime)
	ro.round_default_top = *ReadAnimLayout("round.default.top.", is, sff, at, 2)
	for i := range ro.round_default_bg {
		ro.round_default_bg[i] = *ReadAnimLayout(fmt.Sprintf("round.default.bg%v.", i), is, sff, at, 2)
	}
	ro.round_default = *ReadAnimTextSnd("round.default.", is, sff, at, 2, f)
	for i := range ro.round {
		ro.round[i] = *ReadAnimTextSnd(fmt.Sprintf("round%v.", i+1), is, sff, at, 2, f)
	}
	for i := range ro.round_final_bg {
		ro.round_final_bg[i] = *ReadAnimLayout(fmt.Sprintf("round.final.bg%v.", i), is, sff, at, 2)
	}
	ro.round_final_top = *ReadAnimLayout("round.final.top.", is, sff, at, 2)
	ro.round_final = *ReadAnimTextSnd("round.final.", is, sff, at, 2, f)
	is.ReadI32("fight.time", &ro.fight_time)
	is.ReadI32("fight.sndtime", &ro.fight_sndtime)
	ro.fight = *ReadAnimTextSnd("fight.", is, sff, at, 2, f)
	ro.fight_top = *ReadAnimLayout("fight.top.", is, sff, at, 2)
	for i := range ro.fight_bg {
		ro.fight_bg[i] = *ReadAnimLayout(fmt.Sprintf("fight.bg%v.", i), is, sff, at, 2)
	}
	if is.ReadI32("ctrl.time", &tmp) {
		ro.ctrl_time = Max(1, tmp)
	}
	is.ReadI32("ko.time", &ro.ko_time)
	is.ReadI32("ko.sndtime", &ro.ko_sndtime)
	ro.ko = *ReadAnimTextSnd("ko.", is, sff, at, 1, f)
	ro.ko_top = *ReadAnimLayout("ko.top.", is, sff, at, 1)
	for i := range ro.ko_bg {
		ro.ko_bg[i] = *ReadAnimLayout(fmt.Sprintf("ko.bg%v.", i), is, sff, at, 2)
	}
	ro.dko = *ReadAnimTextSnd("dko.", is, sff, at, 1, f)
	ro.dko_top = *ReadAnimLayout("dko.top.", is, sff, at, 1)
	for i := range ro.dko_bg {
		ro.dko_bg[i] = *ReadAnimLayout(fmt.Sprintf("dko.bg%v.", i), is, sff, at, 2)
	}
	ro.to = *ReadAnimTextSnd("to.", is, sff, at, 1, f)
	ro.to_top = *ReadAnimLayout("to.top.", is, sff, at, 1)
	for i := range ro.to_bg {
		ro.to_bg[i] = *ReadAnimLayout(fmt.Sprintf("to.bg%v.", i), is, sff, at, 2)
	}
	is.ReadI32("slow.time", &ro.slow_time)
	if is.ReadI32("slow.fadetime", &tmp) {
		ro.slow_fadetime = Min(ro.slow_time, tmp)
	} else {
		ro.slow_fadetime = int32(float32(ro.slow_time) * 0.75)
	}
	if is.ReadF32("slow.speed", &ftmp) {
		ro.slow_speed = MinF(1, ftmp)
	}
	if is.ReadI32("over.hittime", &tmp) {
		ro.over_hittime = Max(1, tmp)
	}
	if is.ReadI32("over.waittime", &tmp) {
		ro.over_waittime = Max(1, tmp)
	}
	if is.ReadI32("over.wintime", &tmp) {
		ro.over_wintime = Max(1, tmp)
	}
	if is.ReadI32("over.time", &tmp) {
		ro.over_time = Max(ro.over_wintime+1, tmp)
	}
	is.ReadI32("win.time", &ro.win_time)
	is.ReadI32("win.sndtime", &ro.win_sndtime)
	ro.win = *ReadAnimTextSnd("win.", is, sff, at, 1, f)
	ro.win_top = *ReadAnimLayout("win.top.", is, sff, at, 1)
	for i := range ro.win_bg {
		ro.win_bg[i] = *ReadAnimLayout(fmt.Sprintf("win.bg%v.", i), is, sff, at, 1)
	}
	ro.win2 = *ReadAnimTextSnd("win2.", is, sff, at, 1, f)
	ro.win2_top = *ReadAnimLayout("win2.top.", is, sff, at, 1)
	for i := range ro.win2_bg {
		ro.win2_bg[i] = *ReadAnimLayout(fmt.Sprintf("win2.bg%v.", i), is, sff, at, 1)
	}
	if _, ok := is["win3.text"]; ok {
		ro.win3 = *ReadAnimTextSnd("win3.", is, sff, at, 1, f)
		ro.win3_top = *ReadAnimLayout("win3.top.", is, sff, at, 1)
		for i := range ro.win3_bg {
			ro.win3_bg[i] = *ReadAnimLayout(fmt.Sprintf("win3.bg%v.", i), is, sff, at, 1)
		}
	} else {
		ro.win3 = ro.win2
		ro.win3_top = *ReadAnimLayout("win2.top.", is, sff, at, 1)
		for i := range ro.win2_bg {
			ro.win3_bg[i] = *ReadAnimLayout(fmt.Sprintf("win2.bg%v.", i), is, sff, at, 1)
		}
	}
	if _, ok := is["win4.text"]; ok {
		ro.win4 = *ReadAnimTextSnd("win4.", is, sff, at, 1, f)
		ro.win4_top = *ReadAnimLayout("win4.top.", is, sff, at, 1)
		for i := range ro.win4_bg {
			ro.win4_bg[i] = *ReadAnimLayout(fmt.Sprintf("win4.bg%v.", i), is, sff, at, 1)
		}
	} else {
		ro.win4 = ro.win2
		ro.win4_top = *ReadAnimLayout("win2.top.", is, sff, at, 1)
		for i := range ro.win2_bg {
			ro.win4_bg[i] = *ReadAnimLayout(fmt.Sprintf("win2.bg%v.", i), is, sff, at, 1)
		}
	}
	ro.drawn = *ReadAnimTextSnd("draw.", is, sff, at, 1, f)
	ro.drawn_top = *ReadAnimLayout("draw.top.", is, sff, at, 1)
	for i := range ro.drawn_bg {
		ro.drawn_bg[i] = *ReadAnimLayout(fmt.Sprintf("draw.bg%v.", i), is, sff, at, 1)
	}
	ro.wint[WT_N] = readLbBgTextSnd("p1.n.", is, sff, at, 0, f)
	ro.wint[WT_S] = readLbBgTextSnd("p1.s.", is, sff, at, 0, f)
	ro.wint[WT_H] = readLbBgTextSnd("p1.h.", is, sff, at, 0, f)
	ro.wint[WT_C] = readLbBgTextSnd("p1.c.", is, sff, at, 0, f)
	ro.wint[WT_T] = readLbBgTextSnd("p1.t.", is, sff, at, 0, f)
	ro.wint[WT_Throw] = readLbBgTextSnd("p1.throw.", is, sff, at, 0, f)
	ro.wint[WT_Suicide] = readLbBgTextSnd("p1.suicide.", is, sff, at, 0, f)
	ro.wint[WT_Teammate] = readLbBgTextSnd("p1.teammate.", is, sff, at, 0, f)
	ro.wint[WT_Perfect] = readLbBgTextSnd("p1.perfect.", is, sff, at, 0, f)
	ro.wint[WT_N+WT_NumTypes] = readLbBgTextSnd("p2.n.", is, sff, at, 0, f)
	ro.wint[WT_S+WT_NumTypes] = readLbBgTextSnd("p2.s.", is, sff, at, 0, f)
	ro.wint[WT_H+WT_NumTypes] = readLbBgTextSnd("p2.h.", is, sff, at, 0, f)
	ro.wint[WT_C+WT_NumTypes] = readLbBgTextSnd("p2.c.", is, sff, at, 0, f)
	ro.wint[WT_T+WT_NumTypes] = readLbBgTextSnd("p2.t.", is, sff, at, 0, f)
	ro.wint[WT_Throw+WT_NumTypes] = readLbBgTextSnd("p2.throw.", is, sff, at, 0, f)
	ro.wint[WT_Suicide+WT_NumTypes] = readLbBgTextSnd("p2.suicide.", is, sff, at, 0, f)
	ro.wint[WT_Teammate+WT_NumTypes] = readLbBgTextSnd("p2.teammate.", is, sff, at, 0, f)
	ro.wint[WT_Perfect+WT_NumTypes] = readLbBgTextSnd("p2.perfect.", is, sff, at, 0, f)
	is.ReadI32("fadein.time", &ro.fadein_time)
	var col [3]int32
	if is.ReadI32("fadein.col", &col[0], &col[1], &col[2]) {
		ro.fadein_col = uint32(col[0]&0xff | col[1]&0xff<<8 | col[2]&0xff<<16)
	}
	is.ReadI32("fadeout.time", &ro.fadeout_time)
	ro.over_time = Max(ro.fadeout_time, ro.over_time)
	col = [...]int32{0, 0, 0}
	if is.ReadI32("fadeout.col", &col[0], &col[1], &col[2]) {
		ro.fadeout_col = uint32(col[0]&0xff | col[1]&0xff<<8 | col[2]&0xff<<16)
	}
	is.ReadI32("shutter.time", &ro.shutter_time)
	col = [...]int32{0, 0, 0}
	if is.ReadI32("shutter.col", &col[0], &col[1], &col[2]) {
		ro.shutter_col = uint32(col[0]&0xff | col[1]&0xff<<8 | col[2]&0xff<<16)
	}
	is.ReadI32("callfight.time", &ro.callfight_time)
	return ro
}
func (ro *LifeBarRound) callFight() {
	ro.fight.Reset()
	ro.fight_top.Reset()
	ro.cur, ro.wt[1], ro.swt[1], ro.dt[1] = 1, ro.fight_time, ro.fight_sndtime, 0
	sys.timerCount = append(sys.timerCount, sys.gameTime)
	ro.timerActive = true
}
func (ro *LifeBarRound) act() bool {
	if sys.intro > ro.ctrl_time {
		ro.cur, ro.wt[0], ro.swt[0], ro.dt[0] = 0, ro.round_time, ro.round_sndtime, 0
		ro.wt[1] = ro.callfight_time
	} else if (sys.intro >= 0 && !sys.tickNextFrame()) || sys.dialogueFlg {
		return false
	} else {
		if !ro.introState[0] || !ro.introState[1] {
			if sys.round == 1 && sys.intro == ro.ctrl_time && len(sys.commonLua) > 0 {
				for _, p := range sys.chars {
					if len(p) > 0 && len(p[0].dialogue) > 0 {
						sys.posReset()
						sys.dialogueFlg = true
						return false
					}
				}
			}
			if sys.introSkipped && !sys.dialogueFlg {
				ro.introState[0] = true
				ro.callFight()
				sys.introSkipped = false
			}
			if !ro.introState[0] {
				if ro.swt[0] == 0 {
					if sys.roundType[0] == RT_Final && ro.round_final.snd[0] != -1 {
						ro.snd.play(ro.round_final.snd, 100)
					} else if int(sys.round) <= len(ro.round) && ro.round[sys.round-1].snd[0] != -1 {
						ro.snd.play(ro.round[sys.round-1].snd, 100)
					} else {
						ro.snd.play(ro.round_default.snd, 100)
					}
				}
				ro.swt[0]--
				if ro.wt[0] <= 0 {
					ro.dt[0]++
					if sys.roundType[0] == RT_Final && ro.round_final.snd[0] != -1 {
						if len(ro.round_final_top.anim.frames) > 0 {
							ro.round_final_top.Action()
						} else {
							ro.round_default_top.Action()
						}
						//if len(ro.round_final.anim.anim.frames) > 0 {
						ro.round_final.Action()
						//} else {
						ro.round_default.Action()
						//}
						if len(ro.round_final_bg[0].anim.frames) > 0 {
							for i := len(ro.round_final_bg) - 1; i >= 0; i-- {
								ro.round_final_bg[i].Action()
							}
						} else {
							for i := len(ro.round_default_bg) - 1; i >= 0; i-- {
								ro.round_default_bg[i].Action()
							}
						}
						ro.introState[0] = ro.round_final.End(ro.dt[0], true) && ro.round_default.End(ro.dt[0], true)
					} else if int(sys.round) <= len(ro.round) {
						ro.round_default_top.Action()
						ro.round[sys.round-1].Action()
						ro.round_default.Action()
						for i := len(ro.round_default_bg) - 1; i >= 0; i-- {
							ro.round_default_bg[i].Action()
						}
						ro.introState[0] = ro.round[sys.round-1].End(ro.dt[0], true) && ro.round_default.End(ro.dt[0], true)
					} else {
						ro.round_default_top.Action()
						ro.round_default.Action()
						for i := len(ro.round_default_bg) - 1; i >= 0; i-- {
							ro.round_default_bg[i].Action()
						}
						ro.introState[0] = ro.round_default.End(ro.dt[0], true)
					}
				}
				ro.wt[0]--
			}
			if ro.cur == 0 {
				if ro.wt[1] == 0 {
					ro.callFight()
				}
				ro.wt[1]--
			} else if !ro.introState[1] {
				if ro.swt[1] == 0 {
					ro.snd.play(ro.fight.snd, 100)
				}
				ro.swt[1]--
				if ro.wt[1] <= 0 {
					ro.dt[1]++
					ro.fight_top.Action()
					ro.fight.Action()
					for i := len(ro.fight_bg) - 1; i >= 0; i-- {
						ro.fight_bg[i].Action()
					}
					if ro.fight.End(ro.dt[1], true) && ro.swt[1] < 0 {
						ro.cur, ro.wt[2], ro.swt[2], ro.dt[2] = 2, ro.ko_time, ro.ko_sndtime, 0
						ro.wt[3], ro.swt[3], ro.dt[3] = ro.win_time, ro.win_sndtime, 0
						ro.introState[1] = true
					}
				}
				ro.wt[1]--
			}
		}
		if ro.cur == 2 && sys.intro < 0 && (sys.finish != FT_NotYet || sys.time == 0) {
			if ro.timerActive {
				if sys.gameTime-sys.timerCount[sys.round-1] > 0 {
					sys.timerCount[sys.round-1] = sys.gameTime - sys.timerCount[sys.round-1]
					sys.timerRounds = append(sys.timerRounds, sys.roundTime-sys.time)
				} else {
					sys.timerCount[sys.round-1] = 0
				}
				ro.timerActive = false
			}
			f := func(ats *AnimTextSnd, t int) {
				if -ro.swt[t]-10 == 0 {
					ro.snd.play(ats.snd, 100)
					ro.swt[t]--
				}
				if sys.tickNextFrame() {
					ro.swt[t]--
				}
				if ats.End(ro.dt[t], false) {
					ro.wt[t] = 2
				}
				if /*sys.intro < -ro.ko_time-10*/ ro.wt[t] < -ro.ko_time-10 {
					ro.dt[t]++
					ats.Action()
				}
				ro.wt[t]--
			}
			switch sys.finish {
			case FT_KO:
				ro.ko_top.Action()
				f(&ro.ko, 2)
				for i := len(ro.ko_bg) - 1; i >= 0; i-- {
					ro.ko_bg[i].Action()
				}
			case FT_DKO:
				ro.dko_top.Action()
				f(&ro.dko, 2)
				for i := len(ro.dko_bg) - 1; i >= 0; i-- {
					ro.dko_bg[i].Action()
				}
			default:
				ro.to_top.Action()
				f(&ro.to, 2)
				for i := len(ro.to_bg) - 1; i >= 0; i-- {
					ro.to_bg[i].Action()
				}
			}
			if sys.intro < -(ro.over_hittime + ro.over_waittime + ro.over_wintime) {
				if /*sys.finish == FT_DKO ||*/ sys.finish == FT_TODraw {
					ro.drawn_top.Action()
					f(&ro.drawn, 3)
					for i := len(ro.drawn_bg) - 1; i >= 0; i-- {
						ro.drawn_bg[i].Action()
					}
				} else if sys.winTeam >= 0 && (sys.tmode[sys.winTeam] == TM_Simul || sys.tmode[sys.winTeam] == TM_Tag) {
					if sys.numSimul[sys.winTeam] == 2 {
						ro.win2_top.Action()
						f(&ro.win2, 3)
						for i := len(ro.win2_bg) - 1; i >= 0; i-- {
							ro.win2_bg[i].Action()
						}
					} else if sys.numSimul[sys.winTeam] == 3 {
						ro.win3_top.Action()
						f(&ro.win3, 3)
						for i := len(ro.win3_bg) - 1; i >= 0; i-- {
							ro.win3_bg[i].Action()
						}
					} else {
						ro.win4_top.Action()
						f(&ro.win4, 3)
						for i := len(ro.win4_bg) - 1; i >= 0; i-- {
							ro.win4_bg[i].Action()
						}
					}
				} else {
					ro.win_top.Action()
					f(&ro.win, 3)
					for i := len(ro.win_bg) - 1; i >= 0; i-- {
						ro.win_bg[i].Action()
					}
				}
			}
		} else {
			return ro.cur > 0
		}
	}
	if sys.winTeam >= 0 {
		index := sys.winType[sys.winTeam]
		if index > WT_NumTypes {
			if sys.winTeam == 0 {
				ro.wint[WT_Perfect].step(ro.snd)
				index = index - WT_NumTypes - 1
			} else {
				ro.wint[WT_Perfect+WT_NumTypes].step(ro.snd)
				index = index - 1
			}
		}
		ro.wint[index].step(ro.snd)
	}
	if sys.winTeam >= 0 {
		index := sys.winType[sys.winTeam]
		if index > WT_NumTypes {
			if sys.winTeam == 0 {
				ro.wint[WT_Perfect].step(ro.snd)
				index = index - WT_NumTypes - 1
			} else {
				ro.wint[WT_Perfect+WT_NumTypes].step(ro.snd)
				index = index - 1
			}
		}
		ro.wint[index].step(ro.snd)
	}
	return sys.tickNextFrame()
}
func (ro *LifeBarRound) reset() {
	ro.round_default.Reset()
	ro.round_default_top.Reset()
	for i := range ro.round_default_bg {
		ro.round_default_bg[i].Reset()
	}
	for i := range ro.round {
		ro.round[i].Reset()
	}
	ro.round_final_top.Reset()
	for i := range ro.round_final_bg {
		ro.round_final_bg[i].Reset()
	}
	ro.round_final.Reset()
	ro.fight.Reset()
	ro.fight_top.Reset()
	for i := range ro.fight_bg {
		ro.fight_bg[i].Reset()
	}
	ro.ko.Reset()
	ro.ko_top.Reset()
	for i := range ro.ko_bg {
		ro.ko_bg[i].Reset()
	}
	ro.dko.Reset()
	ro.dko_top.Reset()
	for i := range ro.dko_bg {
		ro.dko_bg[i].Reset()
	}
	ro.to.Reset()
	ro.to_top.Reset()
	for i := range ro.to_bg {
		ro.to_bg[i].Reset()
	}
	ro.win.Reset()
	ro.win_top.Reset()
	for i := range ro.win_bg {
		ro.win_bg[i].Reset()
	}
	ro.win2.Reset()
	ro.win2_top.Reset()
	for i := range ro.win2_bg {
		ro.win2_bg[i].Reset()
	}
	ro.win3.Reset()
	ro.win3_top.Reset()
	for i := range ro.win3_bg {
		ro.win3_bg[i].Reset()
	}
	ro.win4.Reset()
	ro.win4_top.Reset()
	for i := range ro.win4_bg {
		ro.win4_bg[i].Reset()
	}
	ro.drawn.Reset()
	ro.drawn_top.Reset()
	for i := range ro.drawn_bg {
		ro.drawn_bg[i].Reset()
	}
	for i := range ro.wint {
		ro.wint[i].reset()
	}
	ro.introState = [2]bool{}
	ro.firstAttack = [2]bool{}
}
func (ro *LifeBarRound) draw(layerno int16, f []*Fnt) {
	ob := sys.brightness
	sys.brightness = 256
	if !ro.introState[0] && ro.wt[0] < 0 && sys.intro <= ro.ctrl_time {
		tmp := ro.round_default.text.text
		ro.round_default.text.text = OldSprintf(tmp, sys.round)
		for i := range ro.round_default_bg {
			ro.round_default_bg[i].DrawScaled(float32(ro.pos[0])+sys.lifebarOffsetX, float32(ro.pos[1]), layerno, sys.lifebarScale)
		}
		ro.round_default.DrawScaled(float32(ro.pos[0])+sys.lifebarOffsetX, float32(ro.pos[1]),
			layerno, f, sys.lifebarScale)
		ro.round_default.text.text = tmp
		if sys.roundType[0] == RT_Final && (ro.round_final.text.font[0] != -1 ||
			len(ro.round_final.anim.anim.frames) > 0 || len(ro.round_final_bg[0].anim.frames) > 0) {
			tmp = ro.round_final.text.text
			ro.round_final.text.text = OldSprintf(tmp, sys.round)
			for i := range ro.round_final_bg {
				ro.round_final_bg[i].DrawScaled(float32(ro.pos[0])+sys.lifebarOffsetX, float32(ro.pos[1]), layerno, sys.lifebarScale)
			}
			ro.round_final.DrawScaled(float32(ro.pos[0])+sys.lifebarOffsetX, float32(ro.pos[1]),
				layerno, f, sys.lifebarScale)
			ro.round_final.text.text = tmp
		} else if int(sys.round) <= len(ro.round) {
			tmp = ro.round[sys.round-1].text.text
			ro.round[sys.round-1].text.text = OldSprintf(tmp, sys.round)
			ro.round[sys.round-1].DrawScaled(float32(ro.pos[0])+sys.lifebarOffsetX, float32(ro.pos[1]),
				layerno, f, sys.lifebarScale)
			ro.round[sys.round-1].text.text = tmp
		}
		if sys.roundType[0] == RT_Final && len(ro.round_final_top.anim.frames) > 0 {
			ro.round_final_top.DrawScaled(float32(ro.pos[0])+sys.lifebarOffsetX, float32(ro.pos[1]), layerno, sys.lifebarScale)
		} else {
			ro.round_default_top.DrawScaled(float32(ro.pos[0])+sys.lifebarOffsetX, float32(ro.pos[1]), layerno, sys.lifebarScale)
		}
	}
	if !ro.introState[1] && ro.wt[1] < 0 {
		for i := range ro.fight_bg { //240z
			ro.fight_bg[i].DrawScaled(float32(ro.pos[0])+sys.lifebarOffsetX, float32(ro.pos[1]), layerno, sys.lifebarScale) //240z
		}
		ro.fight.DrawScaled(float32(ro.pos[0])+sys.lifebarOffsetX, float32(ro.pos[1]), layerno, f, sys.lifebarScale)
		ro.fight_top.DrawScaled(float32(ro.pos[0])+sys.lifebarOffsetX, float32(ro.pos[1]), layerno, sys.lifebarScale)
	}
	if ro.cur == 2 {
		if /*ro.wt[2] < 0 && sys.intro < -ro.ko_time-10*/ ro.wt[2] < -ro.ko_time-10 {
			switch sys.finish {
			case FT_KO:
				for i := range ro.ko_bg {
					ro.ko_bg[i].DrawScaled(float32(ro.pos[0])+sys.lifebarOffsetX, float32(ro.pos[1]), layerno, sys.lifebarScale)
				}
				ro.ko.DrawScaled(float32(ro.pos[0])+sys.lifebarOffsetX, float32(ro.pos[1]), layerno, f, sys.lifebarScale)
				ro.ko_top.DrawScaled(float32(ro.pos[0])+sys.lifebarOffsetX, float32(ro.pos[1]), layerno, sys.lifebarScale)
			case FT_DKO:
				for i := range ro.dko_bg {
					ro.dko_bg[i].DrawScaled(float32(ro.pos[0])+sys.lifebarOffsetX, float32(ro.pos[1]), layerno, sys.lifebarScale)
				}
				ro.dko.DrawScaled(float32(ro.pos[0])+sys.lifebarOffsetX, float32(ro.pos[1]), layerno, f, sys.lifebarScale)
				ro.dko_top.DrawScaled(float32(ro.pos[0])+sys.lifebarOffsetX, float32(ro.pos[1]), layerno, sys.lifebarScale)
			default:
				for i := range ro.to_bg {
					ro.to_bg[i].DrawScaled(float32(ro.pos[0])+sys.lifebarOffsetX, float32(ro.pos[1]), layerno, sys.lifebarScale)
				}
				ro.to.DrawScaled(float32(ro.pos[0])+sys.lifebarOffsetX, float32(ro.pos[1]), layerno, f, sys.lifebarScale)
				ro.to_top.DrawScaled(float32(ro.pos[0])+sys.lifebarOffsetX, float32(ro.pos[1]), layerno, sys.lifebarScale)
			}
		}
		if ro.wt[3] < 0 {
			if /*sys.finish == FT_DKO ||*/ sys.finish == FT_TODraw {
				for i := range ro.drawn_bg {
					ro.drawn_bg[i].DrawScaled(float32(ro.pos[0])+sys.lifebarOffsetX, float32(ro.pos[1]), layerno, sys.lifebarScale)
				}
				ro.drawn.DrawScaled(float32(ro.pos[0])+sys.lifebarOffsetX, float32(ro.pos[1]), layerno, f, sys.lifebarScale)
				ro.drawn_top.DrawScaled(float32(ro.pos[0])+sys.lifebarOffsetX, float32(ro.pos[1]), layerno, sys.lifebarScale)
			} else if sys.winTeam >= 0 && (sys.tmode[sys.winTeam] == TM_Simul || sys.tmode[sys.winTeam] == TM_Tag) {
				var inter []interface{}
				for i := sys.winTeam; i < len(sys.chars); i += 2 {
					if len(sys.chars[i]) > 0 {
						inter = append(inter, sys.cgi[i].displayname)
					}
				}
				if sys.numSimul[sys.winTeam] == 2 {
					tmp := ro.win2.text.text
					for i := range ro.win2_bg {
						ro.win2_bg[i].DrawScaled(float32(ro.pos[0])+sys.lifebarOffsetX, float32(ro.pos[1]), layerno, sys.lifebarScale)
					}
					ro.win2.text.text = OldSprintf(tmp, inter...)
					ro.win2.DrawScaled(float32(ro.pos[0])+sys.lifebarOffsetX, float32(ro.pos[1]), layerno, f, sys.lifebarScale)
					ro.win2.text.text = tmp
					ro.win2_top.DrawScaled(float32(ro.pos[0])+sys.lifebarOffsetX, float32(ro.pos[1]), layerno, sys.lifebarScale)
				} else if sys.numSimul[sys.winTeam] == 3 {
					tmp := ro.win3.text.text
					for i := range ro.win3_bg {
						ro.win3_bg[i].DrawScaled(float32(ro.pos[0])+sys.lifebarOffsetX, float32(ro.pos[1]), layerno, sys.lifebarScale)
					}
					ro.win3.text.text = OldSprintf(tmp, inter...)
					ro.win3.DrawScaled(float32(ro.pos[0])+sys.lifebarOffsetX, float32(ro.pos[1]), layerno, f, sys.lifebarScale)
					ro.win3.text.text = tmp
					ro.win3_top.DrawScaled(float32(ro.pos[0])+sys.lifebarOffsetX, float32(ro.pos[1]), layerno, sys.lifebarScale)
				} else {
					tmp := ro.win4.text.text
					for i := range ro.win4_bg {
						ro.win4_bg[i].DrawScaled(float32(ro.pos[0])+sys.lifebarOffsetX, float32(ro.pos[1]), layerno, sys.lifebarScale)
					}
					ro.win4.text.text = OldSprintf(tmp, inter...)
					ro.win4.DrawScaled(float32(ro.pos[0])+sys.lifebarOffsetX, float32(ro.pos[1]), layerno, f, sys.lifebarScale)
					ro.win4.text.text = tmp
					ro.win4_top.DrawScaled(float32(ro.pos[0])+sys.lifebarOffsetX, float32(ro.pos[1]), layerno, sys.lifebarScale)
				}
			} else if sys.winTeam >= 0 {
				tmp := ro.win.text.text
				for i := range ro.win_bg {
					ro.win_bg[i].DrawScaled(float32(ro.pos[0])+sys.lifebarOffsetX, float32(ro.pos[1]), layerno, sys.lifebarScale)
				}
				ro.win.text.text = OldSprintf(tmp, sys.cgi[sys.winTeam].displayname)
				ro.win.DrawScaled(float32(ro.pos[0])+sys.lifebarOffsetX, float32(ro.pos[1]), layerno, f, sys.lifebarScale)
				ro.win.text.text = tmp
				ro.win_top.DrawScaled(float32(ro.pos[0])+sys.lifebarOffsetX, float32(ro.pos[1]), layerno, sys.lifebarScale)
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
				ro.wint[WT_Perfect].bgDraw(layerno)
				ro.wint[WT_Perfect].draw(layerno, f)
			} else {
				ro.wint[WT_Perfect+WT_NumTypes].bgDraw(layerno)
				ro.wint[WT_Perfect+WT_NumTypes].draw(layerno, f)
			}
		}
		ro.wint[index].bgDraw(layerno)
		ro.wint[index].draw(layerno, f)
	}
	sys.brightness = ob
}

type LifeBarRatio struct {
	pos  [2]int32
	icon [4]AnimLayout
	bg   AnimLayout
	top  AnimLayout
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
	ra.bg = *ReadAnimLayout(pre+"bg.", is, sff, at, 0)
	return ra
}
func (ra *LifeBarRatio) step(num int32) {
	ra.icon[num].Action()
	ra.bg.Action()
}
func (ra *LifeBarRatio) reset() {
	for i := range ra.icon {
		ra.icon[i].Reset()
	}
	ra.bg.Reset()
}
func (ra *LifeBarRatio) bgDraw(layerno int16) {
	ra.bg.DrawScaled(float32(ra.pos[0])+sys.lifebarOffsetX, float32(ra.pos[1]), layerno, sys.lifebarScale)
}
func (ra *LifeBarRatio) draw(layerno int16, num int32) {
	ra.icon[num].DrawScaled(float32(ra.pos[0])+sys.lifebarOffsetX,
		float32(ra.pos[1]), layerno, sys.lifebarScale)
	ra.top.DrawScaled(float32(ra.pos[0])+sys.lifebarOffsetX, float32(ra.pos[1]), layerno, sys.lifebarScale)
}

type LifeBarTimer struct {
	pos    [2]int32
	text   LbText
	bg     AnimLayout
	top    AnimLayout
	active bool
}

func newLifeBarTimer() *LifeBarTimer {
	return &LifeBarTimer{}
}
func readLifeBarTimer(is IniSection,
	sff *Sff, at AnimationTable, f []*Fnt) *LifeBarTimer {
	tr := newLifeBarTimer()
	is.ReadI32("pos", &tr.pos[0], &tr.pos[1])
	tr.text = *readLbText("text.", is, "", 0, f, 0)
	tr.bg = *ReadAnimLayout("bg.", is, sff, at, 0)
	tr.top = *ReadAnimLayout("top.", is, sff, at, 0)
	return tr
}
func (tr *LifeBarTimer) step() {
	tr.bg.Action()
	tr.top.Action()
}
func (tr *LifeBarTimer) reset() {
	tr.bg.Reset()
	tr.top.Reset()
}
func (tr *LifeBarTimer) bgDraw(layerno int16) {
	if tr.active {
		tr.bg.DrawScaled(float32(tr.pos[0])+sys.lifebarOffsetX, float32(tr.pos[1]), layerno, sys.lifebarScale)
	}
}
func (tr *LifeBarTimer) draw(layerno int16, f []*Fnt) {
	if tr.active && sys.lifebar.ti.framespercount > 0 &&
		tr.text.font[0] >= 0 && int(tr.text.font[0]) < len(f) && f[tr.text.font[0]] != nil && sys.time >= 0 {
		text := tr.text.text
		totalSec := float64(timeTotal()) / 60
		h := math.Floor(totalSec / 3600)
		m := math.Floor((totalSec/3600 - h) * 60)
		s := math.Floor(((totalSec/3600-h)*60 - m) * 60)
		x := math.Floor((((totalSec/3600-h)*60-m)*60 - s) * 100)
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
		tr.text.lay.DrawText(float32(tr.pos[0])+sys.lifebarOffsetX, float32(tr.pos[1]), sys.lifebarScale, layerno,
			text, f[tr.text.font[0]], tr.text.font[1], tr.text.font[2], tr.text.palfx, tr.text.frgba)
		tr.top.DrawScaled(float32(tr.pos[0])+sys.lifebarOffsetX, float32(tr.pos[1]), layerno, sys.lifebarScale)
	}
}
func timeElapsed() int32 {
	return sys.roundTime - sys.time
}
func timeRemaining() int32 {
	if sys.time >= 0 {
		return sys.time
	}
	return -1
}
func timeTotal() int32 {
	t := sys.timerStart
	for _, v := range sys.timerRounds {
		t += v
	}
	if sys.lifebar.ro.timerActive {
		t += timeElapsed()
	}
	return t
}

type LifeBarScore struct {
	pos         [2]int32
	text        LbText
	bg          AnimLayout
	top         AnimLayout
	separator   [2]string
	pad         int32
	places      int32
	min         float32
	max         float32
	active      bool
	scorePoints float32
	rankPoints  map[string]float32
	rankIcons   []string
}

func newLifeBarScore() *LifeBarScore {
	return &LifeBarScore{separator: [2]string{"", "."}}
}
func readLifeBarScore(pre string, is IniSection,
	sff *Sff, at AnimationTable, f []*Fnt) *LifeBarScore {
	sc := newLifeBarScore()
	is.ReadI32(pre+"pos", &sc.pos[0], &sc.pos[1])
	sc.text = *readLbText(pre+"text.", is, "", 0, f, 0)
	sc.separator[0], _ = is.getString("format.integer.separator")
	sc.separator[1], _ = is.getString("format.decimal.separator")
	is.ReadI32("format.integer.pad", &sc.pad)
	is.ReadI32("format.decimal.places", &sc.places)
	is.ReadF32("score.min", &sc.min)
	is.ReadF32("score.max", &sc.max)
	sc.bg = *ReadAnimLayout(pre+"bg.", is, sff, at, 0)
	sc.top = *ReadAnimLayout(pre+"top.", is, sff, at, 0)
	return sc
}
func (sc *LifeBarScore) step() {
	sc.bg.Action()
	sc.top.Action()
}
func (sc *LifeBarScore) reset() {
	sc.bg.Reset()
	sc.top.Reset()
	sc.scorePoints = 0
	sc.rankPoints = make(map[string]float32)
	sc.rankIcons = []string{}
}
func (sc *LifeBarScore) bgDraw(layerno int16) {
	if sc.active {
		sc.bg.DrawScaled(float32(sc.pos[0])+sys.lifebarOffsetX, float32(sc.pos[1]), layerno, sys.lifebarScale)
	}
}
func (sc *LifeBarScore) draw(layerno int16, f []*Fnt, side int) {
	if sc.active && sc.text.font[0] >= 0 && int(sc.text.font[0]) < len(f) && f[sc.text.font[0]] != nil {
		text := sc.text.text
		total := sys.chars[side][0].scoreTotal()
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
		for i := int(sc.pad) - len(s[0]); i > 0; i-- {
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
		sc.text.lay.DrawText(float32(sc.pos[0])+sys.lifebarOffsetX, float32(sc.pos[1]), sys.lifebarScale, layerno,
			text, f[sc.text.font[0]], sc.text.font[1], sc.text.font[2], sc.text.palfx, sc.text.frgba)
		sc.top.DrawScaled(float32(sc.pos[0])+sys.lifebarOffsetX, float32(sc.pos[1]), layerno, sys.lifebarScale)
	}
}

type LifeBarMatch struct {
	pos    [2]int32
	text   LbText
	bg     AnimLayout
	top    AnimLayout
	active bool
}

func newLifeBarMatch() *LifeBarMatch {
	return &LifeBarMatch{}
}
func readLifeBarMatch(is IniSection,
	sff *Sff, at AnimationTable, f []*Fnt) *LifeBarMatch {
	ma := newLifeBarMatch()
	is.ReadI32("pos", &ma.pos[0], &ma.pos[1])
	ma.text = *readLbText("text.", is, "", 0, f, 0)
	ma.bg = *ReadAnimLayout("bg.", is, sff, at, 0)
	ma.top = *ReadAnimLayout("top.", is, sff, at, 0)
	return ma
}
func (ma *LifeBarMatch) step() {
	ma.bg.Action()
	ma.top.Action()
}
func (ma *LifeBarMatch) reset() {
	ma.bg.Reset()
	ma.top.Reset()
}
func (ma *LifeBarMatch) bgDraw(layerno int16) {
	if ma.active {
		ma.bg.DrawScaled(float32(ma.pos[0])+sys.lifebarOffsetX, float32(ma.pos[1]), layerno, sys.lifebarScale)
	}
}
func (ma *LifeBarMatch) draw(layerno int16, f []*Fnt) {
	if ma.active && ma.text.font[0] >= 0 && int(ma.text.font[0]) < len(f) && f[ma.text.font[0]] != nil {
		text := ma.text.text
		text = strings.Replace(text, "%s", fmt.Sprintf("%v", sys.match), 1)
		ma.text.lay.DrawText(float32(ma.pos[0])+sys.lifebarOffsetX, float32(ma.pos[1]), sys.lifebarScale, layerno,
			text, f[ma.text.font[0]], ma.text.font[1], ma.text.font[2], ma.text.palfx, ma.text.frgba)
		ma.top.DrawScaled(float32(ma.pos[0])+sys.lifebarOffsetX, float32(ma.pos[1]), layerno, sys.lifebarScale)
	}
}

type LifeBarAiLevel struct {
	pos       [2]int32
	text      LbText
	bg        AnimLayout
	top       AnimLayout
	separator string
	places    int32
	active    bool
}

func newLifeBarAiLevel() *LifeBarAiLevel {
	return &LifeBarAiLevel{separator: "."}
}
func readLifeBarAiLevel(pre string, is IniSection,
	sff *Sff, at AnimationTable, f []*Fnt) *LifeBarAiLevel {
	ai := newLifeBarAiLevel()
	is.ReadI32(pre+"pos", &ai.pos[0], &ai.pos[1])
	ai.text = *readLbText(pre+"text.", is, "", 0, f, 0)
	ai.separator, _ = is.getString("format.decimal.separator")
	is.ReadI32("format.decimal.places", &ai.places)
	ai.bg = *ReadAnimLayout(pre+"bg.", is, sff, at, 0)
	ai.top = *ReadAnimLayout(pre+"top.", is, sff, at, 0)
	return ai
}
func (ai *LifeBarAiLevel) step() {
	ai.bg.Action()
	ai.top.Action()
}
func (ai *LifeBarAiLevel) reset() {
	ai.bg.Reset()
	ai.top.Reset()
}
func (ai *LifeBarAiLevel) bgDraw(layerno int16) {
	if ai.active {
		ai.bg.DrawScaled(float32(ai.pos[0])+sys.lifebarOffsetX, float32(ai.pos[1]), layerno, sys.lifebarScale)
	}
}
func (ai *LifeBarAiLevel) draw(layerno int16, f []*Fnt, ailv float32) {
	if ai.active && ailv > 0 && ai.text.font[0] >= 0 && int(ai.text.font[0]) < len(f) && f[ai.text.font[0]] != nil {
		text := ai.text.text
		//split float value
		s := strings.Split(fmt.Sprintf("%f", ailv), ".")
		//decimal places (trim trailing numbers)
		if int(ai.places) < len(s[1]) {
			s[1] = s[1][:ai.places]
		}
		//decimal separator
		ds := ""
		if ai.places > 0 {
			ds = ai.separator
		}
		//replace %s with formatted string
		text = strings.Replace(text, "%s", s[0]+ds+s[1], 1)
		//percentage value
		p := ailv / 8 * 100
		text = strings.Replace(text, "%p", fmt.Sprintf("%.0f", p), 1)
		ai.text.lay.DrawText(float32(ai.pos[0])+sys.lifebarOffsetX, float32(ai.pos[1]), sys.lifebarScale, layerno,
			text, f[ai.text.font[0]], ai.text.font[1], ai.text.font[2], ai.text.palfx, ai.text.frgba)
		ai.top.DrawScaled(float32(ai.pos[0])+sys.lifebarOffsetX, float32(ai.pos[1]), layerno, sys.lifebarScale)
	}
}

type LifeBarMode struct {
	pos  [2]int32
	text LbText
	bg   AnimLayout
	top  AnimLayout
}

func newLifeBarMode() *LifeBarMode {
	return &LifeBarMode{}
}
func readLifeBarMode(is IniSection,
	sff *Sff, at AnimationTable, f []*Fnt) map[string]*LifeBarMode {
	mo := make(map[string]*LifeBarMode)
	for k := range is {
		sp := strings.Split(k, ".")
		if _, ok := mo[sp[0]]; !ok {
			mo[sp[0]] = newLifeBarMode()
			is.ReadI32(sp[0]+".pos", &mo[sp[0]].pos[0], &mo[sp[0]].pos[1])
			mo[sp[0]].text = *readLbText(sp[0]+".text.", is, "", 0, f, 0)
			mo[sp[0]].bg = *ReadAnimLayout(sp[0]+".bg.", is, sff, at, 0)
			mo[sp[0]].top = *ReadAnimLayout(sp[0]+".top.", is, sff, at, 0)
		}
	}
	return mo
}
func (mo *LifeBarMode) step() {
	mo.bg.Action()
	mo.top.Action()
}
func (mo *LifeBarMode) reset() {
	mo.bg.Reset()
	mo.top.Reset()
}
func (mo *LifeBarMode) bgDraw(layerno int16) {
	if sys.lifebar.activeMode {
		mo.bg.DrawScaled(float32(mo.pos[0])+sys.lifebarOffsetX, float32(mo.pos[1]), layerno, sys.lifebarScale)
	}
}
func (mo *LifeBarMode) draw(layerno int16, f []*Fnt) {
	if sys.lifebar.activeMode && mo.text.font[0] >= 0 && int(mo.text.font[0]) < len(f) && f[mo.text.font[0]] != nil {
		mo.text.lay.DrawText(float32(mo.pos[0])+sys.lifebarOffsetX, float32(mo.pos[1]), sys.lifebarScale, layerno,
			mo.text.text, f[mo.text.font[0]], mo.text.font[1], mo.text.font[2], mo.text.palfx, mo.text.frgba)
		mo.top.DrawScaled(float32(mo.pos[0])+sys.lifebarOffsetX, float32(mo.pos[1]), layerno, sys.lifebarScale)
	}
}

type Lifebar struct {
	at, fat    AnimationTable
	sff, fsff  *Sff
	snd, fsnd  *Snd
	fnt        [10]*Fnt
	ref        [2]int
	order      [2][]int
	hb         [8][]*HealthBar
	pb         [8][]*PowerBar
	gb         [8][]*GuardBar
	sb         [8][]*StunBar
	fa         [8][]*LifeBarFace
	nm         [8][]*LifeBarName
	wi         [2]*LifeBarWinIcon
	ti         *LifeBarTime
	co         [2]*LifeBarCombo
	ac         [2]*LifeBarAction
	ro         *LifeBarRound
	ra         [2]*LifeBarRatio
	tr         *LifeBarTimer
	sc         [2]*LifeBarScore
	ma         *LifeBarMatch
	ai         [2]*LifeBarAiLevel
	mo         map[string]*LifeBarMode
	missing    map[string]int
	active     bool
	activeBars bool
	activeMode bool
	activeRl   bool
	activeGb   bool
	activeSb   bool
	fx_scale   float32
	deffile    string
}

func loadLifebar(deffile string) (*Lifebar, error) {
	str, err := LoadText(deffile)
	if err != nil {
		return nil, err
	}
	l := &Lifebar{sff: &Sff{}, fsff: &Sff{}, snd: &Snd{},
		hb: [...][]*HealthBar{make([]*HealthBar, 2), make([]*HealthBar, 8),
			make([]*HealthBar, 2), make([]*HealthBar, 8), make([]*HealthBar, 6),
			make([]*HealthBar, 8), make([]*HealthBar, 6), make([]*HealthBar, 8)},
		pb: [...][]*PowerBar{make([]*PowerBar, 2), make([]*PowerBar, 8),
			make([]*PowerBar, 2), make([]*PowerBar, 8), make([]*PowerBar, 6),
			make([]*PowerBar, 8), make([]*PowerBar, 6), make([]*PowerBar, 8)},
		gb: [...][]*GuardBar{make([]*GuardBar, 2), make([]*GuardBar, 8),
			make([]*GuardBar, 2), make([]*GuardBar, 8), make([]*GuardBar, 6),
			make([]*GuardBar, 8), make([]*GuardBar, 6), make([]*GuardBar, 8)},
		sb: [...][]*StunBar{make([]*StunBar, 2), make([]*StunBar, 8),
			make([]*StunBar, 2), make([]*StunBar, 8), make([]*StunBar, 6),
			make([]*StunBar, 8), make([]*StunBar, 6), make([]*StunBar, 8)},
		fa: [...][]*LifeBarFace{make([]*LifeBarFace, 2), make([]*LifeBarFace, 8),
			make([]*LifeBarFace, 2), make([]*LifeBarFace, 8), make([]*LifeBarFace, 6),
			make([]*LifeBarFace, 8), make([]*LifeBarFace, 6), make([]*LifeBarFace, 8)},
		nm: [...][]*LifeBarName{make([]*LifeBarName, 2), make([]*LifeBarName, 8),
			make([]*LifeBarName, 2), make([]*LifeBarName, 8), make([]*LifeBarName, 6),
			make([]*LifeBarName, 8), make([]*LifeBarName, 6), make([]*LifeBarName, 8)},
		active: true, activeBars: true, activeMode: true, fx_scale: 1}
	l.missing = map[string]int{
		"[tag lifebar]": 3, "[simul_3p lifebar]": 4, "[simul_4p lifebar]": 5,
		"[tag_3p lifebar]": 6, "[tag_4p lifebar]": 7, "[simul powerbar]": 1,
		"[turns powerbar]": 2, "[tag powerbar]": 3, "[simul_3p powerbar]": 4,
		"[simul_4p powerbar]": 5, "[tag_3p powerbar]": 6, "[tag_4p powerbar]": 7,
		"[guardbar]": 0, "[simul guardbar]": 1, "[turns guardbar]": 2,
		"[tag guardbar]": 3, "[simul_3p guardbar]": 4, "[simul_4p guardbar]": 5,
		"[tag_3p guardbar]": 6, "[tag_4p guardbar]": 7, "[stunbar]": 0,
		"[simul stunbar]": 1, "[turns stunbar]": 2, "[tag stunbar]": 3,
		"[simul_3p stunbar]": 4, "[simul_4p stunbar]": 5, "[tag_3p stunbar]": 6,
		"[tag_4p stunbar]": 7, "[tag face]": 3, "[simul_3p face]": 4,
		"[simul_4p face]": 5, "[tag_3p face]": 6, "[tag_4p face]": 7,
		"[tag name]": 3, "[simul_3p name]": 4, "[simul_4p name]": 5,
		"[tag_3p name]": 6, "[tag_4p name]": 7, "[action]": -1, "[ratio]": -1,
		"[timer]": -1, "[score]": -1, "[match]": -1, "[ailevel]": -1, "[mode]": -1,
	}
	strc := strings.ToLower(strings.TrimSpace(str))
	for k := range l.missing {
		strc = strings.Replace(strc, ";"+k, "", -1)
		if strings.Contains(strc, k) {
			delete(l.missing, k)
		} else {
			str += "\n" + k
		}
	}
	lines, i := SplitAndTrim(str, "\n"), 0
	l.at = ReadAnimationTable(l.sff, lines, &i)
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
					*l.sff = *s
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
							var height int32 = -1
							if len(is[fmt.Sprintf("font%v.height", i)]) > 0 {
								height = Atoi(is[fmt.Sprintf("font%v.height", i)])
							}
							l.fnt[i], err = loadFnt(filename, height)
							if err != nil {
								err = fmt.Errorf("Failed to load %v font: %v", filename, err)
							}
							return err
						}); err != nil {
						return nil, err
					}
				}
			}
		case "fonts":
			is.ReadF32("scale", &sys.lifebarFontScale)
		case "fightfx":
			is.ReadF32("scale", &l.fx_scale)
		case "lifebar":
			if l.hb[0][0] == nil {
				l.hb[0][0] = readHealthBar("p1.", is, l.sff, l.at, l.fnt[:])
			}
			if l.hb[0][1] == nil {
				l.hb[0][1] = readHealthBar("p2.", is, l.sff, l.at, l.fnt[:])
			}
		case "powerbar":
			if l.pb[0][0] == nil {
				l.pb[0][0] = readPowerBar("p1.", is, l.sff, l.at, l.fnt[:])
			}
			if l.pb[0][1] == nil {
				l.pb[0][1] = readPowerBar("p2.", is, l.sff, l.at, l.fnt[:])
			}
		case "guardbar":
			if l.gb[0][0] == nil {
				l.gb[0][0] = readGuardBar("p1.", is, l.sff, l.at, l.fnt[:])
			}
			if l.gb[0][1] == nil {
				l.gb[0][1] = readGuardBar("p2.", is, l.sff, l.at, l.fnt[:])
			}
		case "stunbar":
			if l.sb[0][0] == nil {
				l.sb[0][0] = readStunBar("p1.", is, l.sff, l.at, l.fnt[:])
			}
			if l.sb[0][1] == nil {
				l.sb[0][1] = readStunBar("p2.", is, l.sff, l.at, l.fnt[:])
			}
		case "face":
			if l.fa[0][0] == nil {
				l.fa[0][0] = readLifeBarFace("p1.", is, l.sff, l.at)
			}
			if l.fa[0][1] == nil {
				l.fa[0][1] = readLifeBarFace("p2.", is, l.sff, l.at)
			}
		case "name":
			if l.nm[0][0] == nil {
				l.nm[0][0] = readLifeBarName("p1.", is, l.sff, l.at, l.fnt[:])
			}
			if l.nm[0][1] == nil {
				l.nm[0][1] = readLifeBarName("p2.", is, l.sff, l.at, l.fnt[:])
			}
		case "turns ":
			subname = strings.ToLower(subname)
			switch {
			case len(subname) >= 7 && subname[:7] == "lifebar":
				if l.hb[2][0] == nil {
					l.hb[2][0] = readHealthBar("p1.", is, l.sff, l.at, l.fnt[:])
				}
				if l.hb[2][1] == nil {
					l.hb[2][1] = readHealthBar("p2.", is, l.sff, l.at, l.fnt[:])
				}
			case len(subname) >= 8 && subname[:8] == "powerbar":
				if l.pb[2][0] == nil {
					l.pb[2][0] = readPowerBar("p1.", is, l.sff, l.at, l.fnt[:])
				}
				if l.pb[2][1] == nil {
					l.pb[2][1] = readPowerBar("p2.", is, l.sff, l.at, l.fnt[:])
				}
			case len(subname) >= 8 && subname[:8] == "guardbar":
				if l.gb[2][0] == nil {
					l.gb[2][0] = readGuardBar("p1.", is, l.sff, l.at, l.fnt[:])
				}
				if l.gb[2][1] == nil {
					l.gb[2][1] = readGuardBar("p2.", is, l.sff, l.at, l.fnt[:])
				}
			case len(subname) >= 7 && subname[:7] == "stunbar":
				if l.sb[2][0] == nil {
					l.sb[2][0] = readStunBar("p1.", is, l.sff, l.at, l.fnt[:])
				}
				if l.sb[2][1] == nil {
					l.sb[2][1] = readStunBar("p2.", is, l.sff, l.at, l.fnt[:])
				}
			case len(subname) >= 4 && subname[:4] == "face":
				if l.fa[2][0] == nil {
					l.fa[2][0] = readLifeBarFace("p1.", is, l.sff, l.at)
				}
				if l.fa[2][1] == nil {
					l.fa[2][1] = readLifeBarFace("p2.", is, l.sff, l.at)
				}
			case len(subname) >= 4 && subname[:4] == "name":
				if l.nm[2][0] == nil {
					l.nm[2][0] = readLifeBarName("p1.", is, l.sff, l.at, l.fnt[:])
				}
				if l.nm[2][1] == nil {
					l.nm[2][1] = readLifeBarName("p2.", is, l.sff, l.at, l.fnt[:])
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
					l.hb[i][0] = readHealthBar("p1.", is, l.sff, l.at, l.fnt[:])
				}
				if l.hb[i][1] == nil {
					l.hb[i][1] = readHealthBar("p2.", is, l.sff, l.at, l.fnt[:])
				}
				if l.hb[i][2] == nil {
					l.hb[i][2] = readHealthBar("p3.", is, l.sff, l.at, l.fnt[:])
				}
				if l.hb[i][3] == nil {
					l.hb[i][3] = readHealthBar("p4.", is, l.sff, l.at, l.fnt[:])
				}
				if l.hb[i][4] == nil {
					l.hb[i][4] = readHealthBar("p5.", is, l.sff, l.at, l.fnt[:])
				}
				if l.hb[i][5] == nil {
					l.hb[i][5] = readHealthBar("p6.", is, l.sff, l.at, l.fnt[:])
				}
				if i != 4 && i != 6 {
					if l.hb[i][6] == nil {
						l.hb[i][6] = readHealthBar("p7.", is, l.sff, l.at, l.fnt[:])
					}
					if l.hb[i][7] == nil {
						l.hb[i][7] = readHealthBar("p8.", is, l.sff, l.at, l.fnt[:])
					}
				}
			case len(subname) >= 8 && subname[:8] == "powerbar":
				if l.pb[i][0] == nil {
					l.pb[i][0] = readPowerBar("p1.", is, l.sff, l.at, l.fnt[:])
				}
				if l.pb[i][1] == nil {
					l.pb[i][1] = readPowerBar("p2.", is, l.sff, l.at, l.fnt[:])
				}
				if l.pb[i][2] == nil {
					l.pb[i][2] = readPowerBar("p3.", is, l.sff, l.at, l.fnt[:])
				}
				if l.pb[i][3] == nil {
					l.pb[i][3] = readPowerBar("p4.", is, l.sff, l.at, l.fnt[:])
				}
				if l.pb[i][4] == nil {
					l.pb[i][4] = readPowerBar("p5.", is, l.sff, l.at, l.fnt[:])
				}
				if l.pb[i][5] == nil {
					l.pb[i][5] = readPowerBar("p6.", is, l.sff, l.at, l.fnt[:])
				}
				if i != 4 && i != 6 {
					if l.pb[i][6] == nil {
						l.pb[i][6] = readPowerBar("p7.", is, l.sff, l.at, l.fnt[:])
					}
					if l.pb[i][7] == nil {
						l.pb[i][7] = readPowerBar("p8.", is, l.sff, l.at, l.fnt[:])
					}
				}
			case len(subname) >= 8 && subname[:8] == "guardbar":
				if l.gb[i][0] == nil {
					l.gb[i][0] = readGuardBar("p1.", is, l.sff, l.at, l.fnt[:])
				}
				if l.gb[i][1] == nil {
					l.gb[i][1] = readGuardBar("p2.", is, l.sff, l.at, l.fnt[:])
				}
				if l.gb[i][2] == nil {
					l.gb[i][2] = readGuardBar("p3.", is, l.sff, l.at, l.fnt[:])
				}
				if l.gb[i][3] == nil {
					l.gb[i][3] = readGuardBar("p4.", is, l.sff, l.at, l.fnt[:])
				}
				if l.gb[i][4] == nil {
					l.gb[i][4] = readGuardBar("p5.", is, l.sff, l.at, l.fnt[:])
				}
				if l.gb[i][5] == nil {
					l.gb[i][5] = readGuardBar("p6.", is, l.sff, l.at, l.fnt[:])
				}
				if i != 4 && i != 6 {
					if l.gb[i][6] == nil {
						l.gb[i][6] = readGuardBar("p7.", is, l.sff, l.at, l.fnt[:])
					}
					if l.gb[i][7] == nil {
						l.gb[i][7] = readGuardBar("p8.", is, l.sff, l.at, l.fnt[:])
					}
				}
			case len(subname) >= 7 && subname[:7] == "stunbar":
				if l.sb[i][0] == nil {
					l.sb[i][0] = readStunBar("p1.", is, l.sff, l.at, l.fnt[:])
				}
				if l.sb[i][1] == nil {
					l.sb[i][1] = readStunBar("p2.", is, l.sff, l.at, l.fnt[:])
				}
				if l.sb[i][2] == nil {
					l.sb[i][2] = readStunBar("p3.", is, l.sff, l.at, l.fnt[:])
				}
				if l.sb[i][3] == nil {
					l.sb[i][3] = readStunBar("p4.", is, l.sff, l.at, l.fnt[:])
				}
				if l.sb[i][4] == nil {
					l.sb[i][4] = readStunBar("p5.", is, l.sff, l.at, l.fnt[:])
				}
				if l.sb[i][5] == nil {
					l.sb[i][5] = readStunBar("p6.", is, l.sff, l.at, l.fnt[:])
				}
				if i != 4 && i != 6 {
					if l.sb[i][6] == nil {
						l.sb[i][6] = readStunBar("p7.", is, l.sff, l.at, l.fnt[:])
					}
					if l.sb[i][7] == nil {
						l.sb[i][7] = readStunBar("p8.", is, l.sff, l.at, l.fnt[:])
					}
				}
			case len(subname) >= 4 && subname[:4] == "face":
				if l.fa[i][0] == nil {
					l.fa[i][0] = readLifeBarFace("p1.", is, l.sff, l.at)
				}
				if l.fa[i][1] == nil {
					l.fa[i][1] = readLifeBarFace("p2.", is, l.sff, l.at)
				}
				if l.fa[i][2] == nil {
					l.fa[i][2] = readLifeBarFace("p3.", is, l.sff, l.at)
				}
				if l.fa[i][3] == nil {
					l.fa[i][3] = readLifeBarFace("p4.", is, l.sff, l.at)
				}
				if l.fa[i][4] == nil {
					l.fa[i][4] = readLifeBarFace("p5.", is, l.sff, l.at)
				}
				if l.fa[i][5] == nil {
					l.fa[i][5] = readLifeBarFace("p6.", is, l.sff, l.at)
				}
				if i != 4 && i != 6 {
					if l.fa[i][6] == nil {
						l.fa[i][6] = readLifeBarFace("p7.", is, l.sff, l.at)
					}
					if l.fa[i][7] == nil {
						l.fa[i][7] = readLifeBarFace("p8.", is, l.sff, l.at)
					}
				}
			case len(subname) >= 4 && subname[:4] == "name":
				if l.nm[i][0] == nil {
					l.nm[i][0] = readLifeBarName("p1.", is, l.sff, l.at, l.fnt[:])
				}
				if l.nm[i][1] == nil {
					l.nm[i][1] = readLifeBarName("p2.", is, l.sff, l.at, l.fnt[:])
				}
				if l.nm[i][2] == nil {
					l.nm[i][2] = readLifeBarName("p3.", is, l.sff, l.at, l.fnt[:])
				}
				if l.nm[i][3] == nil {
					l.nm[i][3] = readLifeBarName("p4.", is, l.sff, l.at, l.fnt[:])
				}
				if l.nm[i][4] == nil {
					l.nm[i][4] = readLifeBarName("p5.", is, l.sff, l.at, l.fnt[:])
				}
				if l.nm[i][5] == nil {
					l.nm[i][5] = readLifeBarName("p6.", is, l.sff, l.at, l.fnt[:])
				}
				if i != 4 && i != 6 {
					if l.nm[i][6] == nil {
						l.nm[i][6] = readLifeBarName("p7.", is, l.sff, l.at, l.fnt[:])
					}
					if l.nm[i][7] == nil {
						l.nm[i][7] = readLifeBarName("p8.", is, l.sff, l.at, l.fnt[:])
					}
				}
			}
		case "winicon":
			if l.wi[0] == nil {
				l.wi[0] = readLifeBarWinIcon("p1.", is, l.sff, l.at, l.fnt[:])
			}
			if l.wi[1] == nil {
				l.wi[1] = readLifeBarWinIcon("p2.", is, l.sff, l.at, l.fnt[:])
			}
		case "time":
			if l.ti == nil {
				l.ti = readLifeBarTime(is, l.sff, l.at, l.fnt[:])
			}
		case "combo":
			if l.co[0] == nil {
				if _, ok := is["team1.pos"]; ok {
					l.co[0] = readLifeBarCombo("team1.", is, l.fnt[:])
				} else {
					l.co[0] = readLifeBarCombo("", is, l.fnt[:])
				}
			}
			if l.co[1] == nil {
				if _, ok := is["team2.pos"]; ok {
					l.co[1] = readLifeBarCombo("team2.", is, l.fnt[:])
				} else {
					l.co[1] = readLifeBarCombo("", is, l.fnt[:])
				}
			}
		case "action":
			if l.ac[0] == nil {
				l.ac[0] = readLifeBarAction("team1.", is, l.fnt[:])
			}
			if l.ac[1] == nil {
				l.ac[1] = readLifeBarAction("team2.", is, l.fnt[:])
			}
		case "round":
			if l.ro == nil {
				l.ro = readLifeBarRound(is, l.sff, l.at, l.snd, l.fnt[:])
			}
		case "ratio":
			if l.ra[0] == nil {
				l.ra[0] = readLifeBarRatio("p1.", is, l.sff, l.at)
			}
			if l.ra[1] == nil {
				l.ra[1] = readLifeBarRatio("p2.", is, l.sff, l.at)
			}
		case "timer":
			if l.tr == nil {
				l.tr = readLifeBarTimer(is, l.sff, l.at, l.fnt[:])
			}
		case "score":
			if l.sc[0] == nil {
				l.sc[0] = readLifeBarScore("p1.", is, l.sff, l.at, l.fnt[:])
			}
			if l.sc[1] == nil {
				l.sc[1] = readLifeBarScore("p2.", is, l.sff, l.at, l.fnt[:])
			}
		case "match":
			if l.ma == nil {
				l.ma = readLifeBarMatch(is, l.sff, l.at, l.fnt[:])
			}
		case "ailevel":
			if l.ai[0] == nil {
				l.ai[0] = readLifeBarAiLevel("p1.", is, l.sff, l.at, l.fnt[:])
			}
			if l.ai[1] == nil {
				l.ai[1] = readLifeBarAiLevel("p2.", is, l.sff, l.at, l.fnt[:])
			}
		case "mode":
			if l.mo == nil {
				l.mo = readLifeBarMode(is, l.sff, l.at, l.fnt[:])
			}
		}
	}
	//fightfx scale
	//TODO: exact formula unknown, test if this code is accurate
	sc := sys.lifebarScale * l.fx_scale
	if sys.lifebarLocalcoord[1] < 240 {
		sc *= float32(sys.lifebarLocalcoord[1]) / 240
	}
	for _, a := range l.fat {
		a.start_scale = [...]float32{sc, sc}
	}
	//Iterate over map in a stable iteration order
	keys := make([]string, 0, len(l.missing))
	for k := range l.missing {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if strings.Contains(k, " lifebar") {
			for i := 3; i < len(l.hb); i++ {
				if i == l.missing[k] {
					for j := 0; j < len(l.hb[i]); j++ {
						if i == 6 || i == 7 {
							l.hb[i][j] = l.hb[3][j]
						} else {
							l.hb[i][j] = l.hb[1][j]
						}
					}
				}
			}
		} else if strings.Contains(k, " powerbar") {
			for i := 1; i < len(l.pb); i++ {
				if i == l.missing[k] {
					for j := 0; j < 2; j++ {
						switch i {
						case 4, 5:
							l.pb[i][j] = l.pb[1][j]
						case 6, 7:
							l.pb[i][j] = l.pb[3][j]
						default:
							l.pb[i][j] = l.pb[0][j]
						}
					}
				}
			}
		} else if strings.Contains(k, " guardbar") {
			for i := 1; i < len(l.gb); i++ {
				if i == l.missing[k] {
					for j := 0; j < 2; j++ {
						switch i {
						case 4, 5:
							l.gb[i][j] = l.gb[1][j]
						case 6, 7:
							l.gb[i][j] = l.gb[3][j]
						default:
							l.gb[i][j] = l.gb[0][j]
						}
					}
				}
			}
		} else if strings.Contains(k, " stunbar") {
			for i := 1; i < len(l.sb); i++ {
				if i == l.missing[k] {
					for j := 0; j < 2; j++ {
						switch i {
						case 4, 5:
							l.sb[i][j] = l.sb[1][j]
						case 6, 7:
							l.sb[i][j] = l.sb[3][j]
						default:
							l.sb[i][j] = l.sb[0][j]
						}
					}
				}
			}
		} else if strings.Contains(k, " face") {
			for i := 3; i < len(l.fa); i++ {
				if i == l.missing[k] {
					for j := 0; j < len(l.fa[i]); j++ {
						if i == 6 || i == 7 {
							l.fa[i][j] = l.fa[3][j]
						} else {
							l.fa[i][j] = l.fa[1][j]
						}
					}
				}
			}
		} else if strings.Contains(k, " name") {
			for i := 3; i < len(l.nm); i++ {
				if i == l.missing[k] {
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
	l.deffile = deffile
	return l, nil
}
func (l *Lifebar) reloadLifebar() {
	lb, _ := loadLifebar(l.deffile)
	lb.ti.framespercount = l.ti.framespercount
	lb.ro.match_maxdrawgames = l.ro.match_maxdrawgames
	lb.ro.match_wins = l.ro.match_wins
	lb.tr.active = l.tr.active
	lb.sc[0].active = l.sc[0].active
	lb.sc[1].active = l.sc[1].active
	lb.ma.active = l.ma.active
	lb.ai[0].active = l.ai[0].active
	lb.ai[1].active = l.ai[1].active
	lb.active = l.active
	lb.activeBars = l.activeBars
	lb.activeMode = l.activeMode
	lb.activeRl = l.activeRl
	lb.activeGb = l.activeGb
	lb.activeSb = l.activeSb
	lb.fx_scale = l.fx_scale
	sys.lifebar = *lb
}
func (l *Lifebar) step() {
	if sys.paused && !sys.step {
		return
	}
	for ti, tm := range sys.tmode {
		if tm == TM_Tag {
			for i, v := range l.order[ti] {
				if sys.teamLeader[sys.chars[v][0].teamside] == sys.chars[v][0].playerNo && sys.chars[v][0].alive() {
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
	for ti := range sys.tmode {
		for i, v := range l.order[ti] {
			//HealthBar
			l.hb[l.ref[ti]][i*2+ti].step(v, l.hb[l.ref[ti]][v])
			//PowerBar
			l.pb[l.ref[ti]][i*2+ti].step(v, l.pb[l.ref[ti]][v], l.snd)
			//GuardBar
			l.gb[l.ref[ti]][i*2+ti].step(v, l.gb[l.ref[ti]][v], l.snd)
			//StunBar
			l.sb[l.ref[ti]][i*2+ti].step(v, l.sb[l.ref[ti]][v], l.snd)
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
	cb, fcb, cd, cp, st := [2]int32{}, [2]int32{}, [2]int32{}, [2]float32{}, [2]bool{}
	for i, ch := range sys.chars {
		for _, c := range ch {
			if c.alive() || !c.scf(SCF_over) {
				// Fake Combo
				if c.receivedHits > cb[^i&1] {
					cb[^i&1] = Min(999, Max(c.fakeReceivedHits, cb[^i&1]))
				}
				if c.fakeReceivedHits > fcb[^i&1] {
					fcb[^i&1] = Min(999, Max(c.fakeReceivedHits, fcb[^i&1]))
					cd[^i&1] = Max(c.fakeComboDmg, cd[^i&1])
					cp[^i&1] = float32(cd[^i&1]) / float32(c.lifeMax) * 100
				}
				if c.fakeReceivedHits > 0 && !st[^i&1] && c.scf(SCF_dizzy) {
					st[^i&1] = true
				}
			}
		}
	}
	for i := range l.co {
		l.co[i].step(cb[i], fcb[i], cd[i], cp[i], st[i])
	}
	//LifeBarAction
	for i := range l.ac {
		l.ac[i].step(l.order[i][0])
	}
	//LifeBarRatio
	for ti, tm := range sys.tmode {
		if tm == TM_Turns {
			rl := sys.chars[ti][0].ratioLevel()
			if rl > 0 {
				l.ra[ti].step(rl - 1)
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
	for _, gb := range l.gb {
		for i := range gb {
			gb[i].reset()
		}
	}
	for _, sb := range l.sb {
		for i := range sb {
			sb[i].reset()
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
	for i := range l.ac {
		l.ac[i].reset(l.order[i][0])
	}
	l.ro.reset()
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
	if sys.postMatchFlg || sys.dialogueBarsFlg {
		return
	}
	if sys.statusDraw && l.active {
		if !sys.sf(GSF_nobardisplay) && l.activeBars {
			//HealthBar
			for ti := range sys.tmode {
				for i := range l.order[ti] {
					l.hb[l.ref[ti]][i*2+ti].bgDraw(layerno)
				}
			}
			for ti := range sys.tmode {
				for i, v := range l.order[ti] {
					l.hb[l.ref[ti]][i*2+ti].draw(layerno, v, l.hb[l.ref[ti]][v], l.fnt[:])
				}
			}
			//PowerBar
			for ti, tm := range sys.tmode {
				for i := range l.order[ti] {
					if !sys.chars[i*2+ti][0].sf(CSF_nopowerbardisplay) {
						if sys.powerShare[ti] && (tm == TM_Simul || tm == TM_Tag) {
							if i == 0 {
								l.pb[l.ref[ti]][i*2+ti].bgDraw(layerno)
							}
						} else {
							l.pb[l.ref[ti]][i*2+ti].bgDraw(layerno)
						}
					}
				}
			}
			for ti, tm := range sys.tmode {
				for i, v := range l.order[ti] {
					if !sys.chars[i*2+ti][0].sf(CSF_nopowerbardisplay) {
						if sys.powerShare[ti] && (tm == TM_Simul || tm == TM_Tag) {
							if i == 0 {
								l.pb[l.ref[ti]][i*2+ti].draw(layerno, i*2+ti, l.pb[l.ref[ti]][i*2+ti], l.fnt[:])
							}
						} else {
							l.pb[l.ref[ti]][i*2+ti].draw(layerno, v, l.pb[l.ref[ti]][v], l.fnt[:])
						}
					}
				}
			}
			//GuardBar
			for ti := range sys.tmode {
				for i := range l.order[ti] {
					l.gb[l.ref[ti]][i*2+ti].bgDraw(layerno)
				}
			}
			for ti := range sys.tmode {
				for i, v := range l.order[ti] {
					l.gb[l.ref[ti]][i*2+ti].draw(layerno, v, l.gb[l.ref[ti]][v], l.fnt[:])
				}
			}
			//StunBar
			for ti := range sys.tmode {
				for i := range l.order[ti] {
					l.sb[l.ref[ti]][i*2+ti].bgDraw(layerno)
				}
			}
			for ti := range sys.tmode {
				for i, v := range l.order[ti] {
					l.sb[l.ref[ti]][i*2+ti].draw(layerno, v, l.sb[l.ref[ti]][v], l.fnt[:])
				}
			}
			//LifeBarFace
			for ti := range sys.tmode {
				for i := range l.order[ti] {
					l.fa[l.ref[ti]][i*2+ti].bgDraw(layerno)
				}
			}
			for ti := range sys.tmode {
				for i, v := range l.order[ti] {
					l.fa[l.ref[ti]][i*2+ti].draw(layerno, v, l.fa[l.ref[ti]][v])
				}
			}
			//LifeBarName
			for ti := range sys.tmode {
				for i := range l.order[ti] {
					l.nm[l.ref[ti]][i*2+ti].bgDraw(layerno)
				}
			}
			for ti := range sys.tmode {
				for i, v := range l.order[ti] {
					l.nm[l.ref[ti]][i*2+ti].draw(layerno, v, l.fnt[:], ti)
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
					if rl := sys.chars[ti][0].ratioLevel(); rl > 0 {
						l.ra[ti].bgDraw(layerno)
					}
				}
			}
			for ti, tm := range sys.tmode {
				if tm == TM_Turns {
					if rl := sys.chars[ti][0].ratioLevel(); rl > 0 {
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
			}
			for i := range l.sc {
				l.sc[i].draw(layerno, l.fnt[:], i)
			}
			//LifeBarMatch
			l.ma.bgDraw(layerno)
			l.ma.draw(layerno, l.fnt[:])
			//LifeBarAiLevel
			for i := range l.ai {
				l.ai[i].bgDraw(layerno)
			}
			for i := range l.ai {
				l.ai[i].draw(layerno, l.fnt[:], sys.com[sys.chars[i][0].playerNo])
			}

		}
		//LifeBarCombo
		for i := range l.co {
			l.co[i].draw(layerno, l.fnt[:], i)
		}
		//LifeBarAction
		for i := range l.ac {
			l.ac[i].draw(layerno, l.fnt[:], i)
		}
		//LifeBarMode
		if _, ok := l.mo[sys.gameMode]; ok {
			l.mo[sys.gameMode].bgDraw(layerno)
			l.mo[sys.gameMode].draw(layerno, l.fnt[:])
		}
	}
	if l.active {
		//LifeBarRound
		l.ro.draw(layerno, l.fnt[:])
	}
}
