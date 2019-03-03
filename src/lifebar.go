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

func (wt *WinType) SetPerfect() {
	if *wt >= WT_N && *wt <= WT_Teammate {
		*wt += WT_PN - WT_N
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
	hb.bg0.DrawScaled(float32(hb.pos[0]), float32(hb.pos[1]), layerno, sys.LifebarScale)
	hb.bg1.DrawScaled(float32(hb.pos[0]), float32(hb.pos[1]), layerno, sys.LifebarScale)
	hb.bg2.DrawScaled(float32(hb.pos[0]), float32(hb.pos[1]), layerno, sys.LifebarScale)
}
func (hb *HealthBar) draw(layerno int16, life float32) {
	width := func(life float32) (r [4]int32) {
		r = sys.scrrect
		if hb.range_x[0] < hb.range_x[1] {
			r[0] = int32((float32(hb.pos[0]+hb.range_x[0])+
				float32(sys.gameWidth-320)/2)*sys.widthScale + 0.5)
			r[2] = int32(float32(hb.range_x[1]-hb.range_x[0]+1)*life*
				sys.widthScale + 0.5)
		} else {
			r[2] = int32(float32(hb.range_x[0]-hb.range_x[1]+1)*life*
				sys.widthScale + 0.5)
			r[0] = int32((float32(hb.pos[0]+hb.range_x[0]+1)+
				float32(sys.gameWidth-320)/2)*sys.widthScale+0.5) - r[2]
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
	hb.mid.lay.DrawAnim(&mr, float32(hb.pos[0]), float32(hb.pos[1]), 1,
		layerno, &hb.mid.anim)
	hb.front.lay.DrawAnim(&lr, float32(hb.pos[0]), float32(hb.pos[1]), 1,
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
		is.ReadI32(fmt.Sprintf("%vlevel%v.snd", pre, i+1), &pb.level_snd[i][0],
			&pb.level_snd[i][1])
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
	pb.bg0.DrawScaled(float32(pb.pos[0]), float32(pb.pos[1]), layerno, sys.LifebarScale)
	pb.bg1.DrawScaled(float32(pb.pos[0]), float32(pb.pos[1]), layerno, sys.LifebarScale)
	pb.bg2.DrawScaled(float32(pb.pos[0]), float32(pb.pos[1]), layerno, sys.LifebarScale)
}
func (pb *PowerBar) draw(layerno int16, power float32,
	level int32, f []*Fnt) {
	width := func(power float32) (r [4]int32) {
		r = sys.scrrect
		if pb.range_x[0] < pb.range_x[1] {
			r[0] = int32((float32(pb.pos[0]+pb.range_x[0])+
				float32(sys.gameWidth-320)/2)*sys.widthScale + 0.5)
			r[2] = int32(float32(pb.range_x[1]-pb.range_x[0]+1)*power*
				sys.widthScale + 0.5)
		} else {
			r[2] = int32(float32(pb.range_x[0]-pb.range_x[1]+1)*power*
				sys.widthScale + 0.5)
			r[0] = int32((float32(pb.pos[0]+pb.range_x[0]+1)+
				float32(sys.gameWidth-320)/2)*sys.widthScale+0.5) - r[2]
		}
		return
	}
	pr, mr := width(power), width(pb.midpower)
	if pb.range_x[0] < pb.range_x[1] {
		mr[0] += pr[2]
	}
	mr[2] -= Min(mr[2], pr[2])
	pb.mid.lay.DrawAnim(&mr, float32(pb.pos[0]), float32(pb.pos[1]), 1,
		layerno, &pb.mid.anim)
	pb.front.lay.DrawAnim(&pr, float32(pb.pos[0]), float32(pb.pos[1]), 1,
		layerno, &pb.front.anim)
	if pb.counter_font[0] >= 0 && int(pb.counter_font[0]) < len(f) {
		pb.counter_lay.DrawText(float32(pb.pos[0]), float32(pb.pos[1]), 1,
			layerno, fmt.Sprintf("%v", level),
			f[pb.counter_font[0]], pb.counter_font[1], pb.counter_font[2])
	}
}

type LifeBarFace struct {
	pos               [2]int32
	bg                AnimLayout
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
	f.teammate_bg.Action()
	f.teammate_ko.Action()
}
func (f *LifeBarFace) reset() {
	f.bg.Reset()
	f.teammate_bg.Reset()
	f.teammate_ko.Reset()
}
func (f *LifeBarFace) bgDraw(layerno int16) {
	f.bg.DrawScaled(float32(f.pos[0]), float32(f.pos[1]), layerno, sys.LifebarScale)
}
func (f *LifeBarFace) draw(layerno int16, fx *PalFX, superplayer bool) {
	ob := sys.brightness
	if superplayer {
		sys.brightness = 256
	}
	f.face_lay.DrawSprite(float32(f.pos[0]), float32(f.pos[1]), layerno,
		f.face, fx, f.scale)
	sys.brightness = ob
	i := int32(len(f.teammate_face)) - 1
	x := float32(f.teammate_pos[0] + f.teammate_spacing[0]*(i-1))
	y := float32(f.teammate_pos[1] + f.teammate_spacing[1]*(i-1))
	for ; i >= 0; i-- {
		if i != f.numko {
			f.teammate_bg.Draw(x, y, layerno)
			f.teammate_face_lay.DrawSprite(x, y, layerno, f.teammate_face[i], nil, f.teammate_scale[i])
			if i < f.numko {
				f.teammate_ko.Draw(x, y, layerno)
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
	n.bg.DrawScaled(float32(n.pos[0]), float32(n.pos[1]), layerno, sys.LifebarScale)
}
func (n *LifeBarName) draw(layerno int16, f []*Fnt, name string) {
	if n.name_font[0] >= 0 && int(n.name_font[0]) < len(f) {
		n.name_lay.DrawText(float32(n.pos[0]), float32(n.pos[1]), 1, layerno, name,
			f[n.name_font[0]], n.name_font[1], n.name_font[2])
	}
}

type LifeBarWinIcon struct {
	pos           [2]int32
	iconoffset    [2]int32
	useiconupto   int32
	counter_font  [3]int32
	counter_lay   Layout
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
	for i := range wi.icon {
		wi.icon[i].Reset()
	}
	wi.numWins = len(wi.wins)
	wi.added, wi.addedP = nil, nil
}
func (wi *LifeBarWinIcon) clear() { wi.wins = nil }
func (wi *LifeBarWinIcon) draw(layerno int16, f []*Fnt) {
	if len(wi.wins) > int(wi.useiconupto) {
		if wi.counter_font[0] >= 0 && int(wi.counter_font[0]) < len(f) {
			wi.counter_lay.DrawText(float32(wi.pos[0]), float32(wi.pos[1]), 1,
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
			wi.icon[wt].Draw(float32(wi.pos[0]+wi.iconoffset[0]*int32(i)),
				float32(wi.pos[1]+wi.iconoffset[1]*int32(i)), layerno)
			if p {
				wi.icon[WT_Perfect].Draw(float32(wi.pos[0]+wi.iconoffset[0]*int32(i)),
					float32(wi.pos[1]+wi.iconoffset[1]*int32(i)), layerno)
			}
		}
		if wi.added != nil {
			wt, p := wi.wins[i], false
			if wi.addedP != nil {
				wt -= WT_PN
				p = true
			}
			wi.icon[wt].lay.DrawAnim(&sys.scrrect,
				float32(wi.pos[0]+wi.iconoffset[0]*int32(i)),
				float32(wi.pos[1]+wi.iconoffset[1]*int32(i)), 1, layerno, wi.added)
			if p {
				wi.icon[WT_Perfect].lay.DrawAnim(&sys.scrrect,
					float32(wi.pos[0]+wi.iconoffset[0]*int32(i)),
					float32(wi.pos[1]+wi.iconoffset[1]*int32(i)), 1, layerno, wi.addedP)
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
	t.bg.DrawScaled(float32(t.pos[0]), float32(t.pos[1]), layerno, sys.LifebarScale)
}
func (t *LifeBarTime) draw(layerno int16, f []*Fnt) {
	if t.framespercount > 0 &&
		t.counter_font[0] >= 0 && int(t.counter_font[0]) < len(f) {
		time := "o"
		if sys.time >= 0 {
			time = fmt.Sprintf("%v", sys.time/t.framespercount)
		}
		t.counter_lay.DrawText(float32(t.pos[0]), float32(t.pos[1]), 1, layerno,
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
	cur, old      [2]int32
	resttime      [2]int32
	counterX      [2]float32
	shaketime     [2]int32
}

func newLifeBarCombo() *LifeBarCombo {
	return &LifeBarCombo{counter_font: [3]int32{-1}, text_font: [3]int32{-1},
		displaytime: 90}
}
func readLifeBarCombo(is IniSection) *LifeBarCombo {
	c := newLifeBarCombo()
	is.ReadI32("pos", &c.pos[0], &c.pos[1])
	is.ReadF32("start.x", &c.start_x)
	is.ReadI32("counter.font", &c.counter_font[0], &c.counter_font[1],
		&c.counter_font[2])
	is.ReadBool("counter.shake", &c.counter_shake)
	c.counter_lay = *ReadLayout("counter.", is, 2)
	c.counter_lay.offset = [2]float32{}
	is.ReadI32("text.font", &c.text_font[0], &c.text_font[1], &c.text_font[2])
	c.text_text = is["text.text"]
	c.text_lay = *ReadLayout("text.", is, 2)
	is.ReadI32("displaytime", &c.displaytime)
	return c
}
func (c *LifeBarCombo) step(combo [2]int32) {
	for i := range c.cur {
		if c.resttime[i] > 0 {
			c.counterX[i] -= c.counterX[i] / 8
		} else {
			c.counterX[i] -= sys.lifebarFontScale * 4
			if c.counterX[i] < c.start_x*2 {
				c.counterX[i] = c.start_x * 2
			}
		}
		if c.shaketime[i] > 0 {
			c.shaketime[i]--
		}
		if AbsF(c.counterX[i]) < 1 {
			c.resttime[i]--
		}
		if combo[i] >= 2 && c.old[i] != combo[i] {
			c.cur[i] = combo[i]
			c.resttime[i] = c.displaytime
			if c.counter_shake {
				c.shaketime[i] = 15
			}
		}
		c.old[i] = combo[i]
	}
}
func (c *LifeBarCombo) reset() {
	c.cur, c.old, c.resttime = [2]int32{}, [2]int32{}, [2]int32{}
	c.counterX = [...]float32{c.start_x * 2, c.start_x * 2}
	c.shaketime = [2]int32{}
}
func (c *LifeBarCombo) draw(layerno int16, f []*Fnt) {
	haba := func(n int32) float32 {
		if c.counter_font[0] < 0 || int(c.counter_font[0]) >= len(f) {
			return 0
		}
		return float32(f[c.counter_font[0]].TextWidth(fmt.Sprintf("%v", n)))
	}
	for i := range c.cur {
		if c.resttime[i] <= 0 && c.counterX[i] == c.start_x*2 {
			continue
		}
		var x float32
		if i&1 == 0 {
			if c.start_x <= 0 {
				x = c.counterX[i]
			}
			x += float32(c.pos[0]) + haba(c.cur[i])
		} else {
			if c.start_x <= 0 {
				x = -c.counterX[i]
			}
			x += 320 - float32(c.pos[0])
		}
		if c.text_font[0] >= 0 && int(c.text_font[0]) < len(f) {
			text := OldSprintf(c.text_text, c.cur[i])
			if i&1 == 0 {
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
			c.text_lay.DrawText(x, float32(c.pos[1]), 1, layerno,
				text, f[c.text_font[0]], c.text_font[1], 1)
		}
		if c.counter_font[0] >= 0 && int(c.counter_font[0]) < len(f) {
			z := 1 + float32(c.shaketime[i])*(1.0/20)*
				float32(math.Sin(float64(c.shaketime[i])*(math.Pi/2.5)))
			c.counter_lay.DrawText(x/z, float32(c.pos[1])/z, z, layerno,
				fmt.Sprintf("%v", c.cur[i]), f[c.counter_font[0]],
				c.counter_font[1], -1)
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
	tmp = Atoi(sys.cmdFlags["-rounds"])
	if tmp > 0 {
		r.match_wins = tmp
	} else {
		is.ReadI32("match.wins", &r.match_wins)
	}
	is.ReadI32("match.maxdrawgames", &r.match_maxdrawgames)
	if is.ReadI32("start.waittime", &tmp) {
		r.start_waittime = Max(1, tmp)
	}
	is.ReadI32("round.time", &r.round_time)
	is.ReadI32("round.sndtime", &r.round_sndtime)
	r.round_default = *ReadAnimTextSnd("round.default.", is, sff, at, 2)
	for i := range r.round {
		r.round[i] = r.round_default
		r.round[i].Read(fmt.Sprintf("round%v.", i+1), is, at, 2)
	}
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
				if int(sys.round) <= len(r.round) {
					r.snd.play(r.round[sys.round-1].snd)
				} else {
					r.snd.play(r.round_default.snd)
				}
			}
			r.swt[0]--
			if r.wt[0] <= 0 {
				r.dt[0]++
				end := false
				if int(sys.round) <= len(r.round) {
					r.round[sys.round-1].Action()
					end = r.round[sys.round-1].End(r.dt[0])
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
	} else if r.cur == 2 && (sys.finish != FT_NotYet || sys.time == 0) {
		if r.timerActive {
			if sys.gameTime-sys.timerCount[sys.round-1] > 0 {
				sys.timerCount[sys.round-1] = sys.gameTime - sys.timerCount[sys.round-1]
			} else {
				sys.timerCount[sys.round-1] = 0
			}
			r.timerActive = false
		}
		f := func(ats *AnimTextSnd, t int) {
			if r.swt[t] == 0 {
				r.snd.play(ats.snd)
			}
			r.swt[t]--
			if ats.End(r.dt[t]) {
				r.wt[t] = 2
			}
			if r.wt[t] <= 0 {
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
			if sys.finish == FT_DKO || sys.finish == FT_TODraw {
				f(&r.drawn, 1)
			} else if sys.winTeam >= 0 && sys.tmode[sys.winTeam] == TM_Simul {
				f(&r.win2, 1)
			} else {
				f(&r.win, 1)
			}
		}
	}
	return sys.tickNextFrame()
}
func (r *LifeBarRound) reset() {
	r.round_default.Reset()
	for i := range r.round {
		r.round[i].Reset()
	}
	r.fight.Reset()
	r.ko.Reset()
	r.dko.Reset()
	r.to.Reset()
	r.win.Reset()
	r.win2.Reset()
	r.drawn.Reset()
}
func (r *LifeBarRound) draw(layerno int16) {
	ob := sys.brightness
	sys.brightness = 255
	switch r.cur {
	case 0:
		if r.wt[0] < 0 && sys.intro <= r.ctrl_time {
			if int(sys.round) <= len(r.round) {
				tmp := r.round[sys.round-1].text
				r.round[sys.round-1].text = OldSprintf(tmp, sys.round)
				r.round[sys.round-1].Draw(float32(r.pos[0]), float32(r.pos[1]),
					layerno, r.fnt)
				r.round[sys.round-1].text = tmp
			} else {
				tmp := r.round_default.text
				r.round_default.text = OldSprintf(tmp, sys.round)
				r.round_default.Draw(float32(r.pos[0]), float32(r.pos[1]),
					layerno, r.fnt)
				r.round_default.text = tmp
			}
		}
	case 1:
		if r.wt[0] < 0 {
			r.fight.Draw(float32(r.pos[0]), float32(r.pos[1]), layerno, r.fnt)
		}
	case 2:
		if r.wt[0] < 0 {
			switch sys.finish {
			case FT_KO:
				r.ko.Draw(float32(r.pos[0]), float32(r.pos[1]), layerno, r.fnt)
			case FT_DKO:
				r.dko.Draw(float32(r.pos[0]), float32(r.pos[1]), layerno, r.fnt)
			default:
				r.to.Draw(float32(r.pos[0]), float32(r.pos[1]), layerno, r.fnt)
			}
		}
		if r.wt[1] < 0 {
			if sys.finish == FT_DKO || sys.finish == FT_TODraw {
				r.drawn.Draw(float32(r.pos[0]), float32(r.pos[1]), layerno, r.fnt)
			} else if sys.tmode[sys.winTeam] == TM_Simul {
				tmp := r.win2.text
				var inter []interface{}
				for i := sys.winTeam; i < len(sys.chars); i += 2 {
					if len(sys.chars[i]) > 0 {
						inter = append(inter, sys.cgi[i].displayname)
					}
				}
				r.win2.text = OldSprintf(tmp, inter...)
				r.win2.Draw(float32(r.pos[0]), float32(r.pos[1]), layerno, r.fnt)
				r.win2.text = tmp
			} else {
				tmp := r.win.text
				r.win.text = OldSprintf(tmp, sys.cgi[sys.winTeam].displayname)
				r.win.Draw(float32(r.pos[0]), float32(r.pos[1]), layerno, r.fnt)
				r.win.text = tmp
			}
		}
	}
	sys.brightness = ob
}

type Lifebar struct {
	fat       AnimationTable
	fsff      *Sff
	snd, fsnd *Snd
	fnt       [10]*Fnt
	hb        [3][]*HealthBar
	pb        [2]*PowerBar
	fa        [3][]*LifeBarFace
	nm        [3][]*LifeBarName
	wi        [2]*LifeBarWinIcon
	ti        *LifeBarTime
	co        *LifeBarCombo
	ro        *LifeBarRound
}

func loadLifebar(deffile string) (*Lifebar, error) {
	str, err := LoadText(deffile)
	if err != nil {
		return nil, err
	}
	l := &Lifebar{fsff: &Sff{}, snd: &Snd{},
		hb: [...][]*HealthBar{make([]*HealthBar, 2), make([]*HealthBar, 4),
			make([]*HealthBar, 2)},
		fa: [...][]*LifeBarFace{make([]*LifeBarFace, 2), make([]*LifeBarFace, 4),
			make([]*LifeBarFace, 2)},
		nm: [...][]*LifeBarName{make([]*LifeBarName, 2), make([]*LifeBarName, 4),
			make([]*LifeBarName, 2)}}
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
			if l.pb[0] == nil {
				l.pb[0] = readPowerBar("p1.", is, sff, at, l.snd)
			}
			if l.pb[1] == nil {
				l.pb[1] = readPowerBar("p2.", is, sff, at, l.snd)
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
		}
	}
	return l, nil
}
func (l *Lifebar) step() {
	for ti, tm := range sys.tmode {
		for i := ti; i < len(l.hb[tm]); i += 2 {
			l.hb[tm][i].step(float32(sys.chars[i][0].life)/
				float32(sys.chars[i][0].lifeMax), (sys.chars[i][0].getcombo != 0 ||
				sys.chars[i][0].ss.moveType == MT_H) &&
				!sys.chars[i][0].scf(SCF_over))
		}
	}
	for i := range l.pb {
		lvi := i
		if sys.tmode[i] == TM_Simul {
			lvi += 2
		}
		l.pb[i].step(float32(sys.chars[i][0].power)/
			float32(sys.chars[i][0].powerMax), sys.chars[lvi][0].power/1000)
	}
	for ti, tm := range sys.tmode {
		for i := ti; i < len(l.fa[tm]); i += 2 {
			l.fa[tm][i].step()
		}
	}
	for ti, tm := range sys.tmode {
		for i := ti; i < len(l.nm[tm]); i += 2 {
			l.nm[tm][i].step()
		}
	}
	for i := range l.wi {
		l.wi[i].step(sys.wins[i])
	}
	l.ti.step()
	cb := [2]int32{}
	for i, ch := range sys.chars {
		for _, c := range ch {
			cb[^i&1] = Min(999, Max(c.getcombo, cb[^i&1]))
		}
	}
	l.co.step(cb)
}
func (l *Lifebar) reset() {
	for _, hb := range l.hb {
		for i := range hb {
			hb[i].reset()
		}
	}
	for i := range l.pb {
		l.pb[i].reset()
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
}
func (l *Lifebar) draw(layerno int16) {
	if !sys.statusDraw {
		return
	}
	if !sys.sf(GSF_nobardisplay) {
		for ti, tm := range sys.tmode {
			for i := ti; i < len(l.hb[tm]); i += 2 {
				l.hb[tm][i].bgDraw(layerno)
			}
		}
		for ti, tm := range sys.tmode {
			for i := ti; i < len(l.hb[tm]); i += 2 {
				l.hb[tm][i].draw(layerno, float32(sys.chars[i][0].life)/
					float32(sys.chars[i][0].lifeMax))
			}
		}
		for i := range l.pb {
			l.pb[i].bgDraw(layerno)
		}
		for i := range l.pb {
			l.pb[i].draw(layerno, float32(sys.chars[i][0].power)/
				float32(sys.chars[i][0].powerMax), sys.chars[i][0].power/1000,
				l.fnt[:])
		}
		for ti, tm := range sys.tmode {
			for i := ti; i < len(l.fa[tm]); i += 2 {
				l.fa[tm][i].bgDraw(layerno)
			}
		}
		for ti, tm := range sys.tmode {
			for i := ti; i < len(l.fa[tm]); i += 2 {
				if fspr := l.fa[tm][i].face; fspr != nil {
					pfx := sys.chars[i][0].getPalfx()
					sys.cgi[i].sff.palList.SwapPalMap(&pfx.remap)
					fspr.Pal = nil
					fspr.Pal = fspr.GetPal(&sys.cgi[i].sff.palList)
					sys.cgi[i].sff.palList.SwapPalMap(&pfx.remap)
					l.fa[tm][i].draw(layerno, pfx, i == sys.superplayer)
				}
			}
		}
		for ti, tm := range sys.tmode {
			for i := ti; i < len(l.nm[tm]); i += 2 {
				l.nm[tm][i].bgDraw(layerno)
			}
		}
		for ti, tm := range sys.tmode {
			for i := ti; i < len(l.nm[tm]); i += 2 {
				l.nm[tm][i].draw(layerno, l.fnt[:], sys.cgi[i].displayname)
			}
		}
		l.ti.bgDraw(layerno)
		l.ti.draw(layerno, l.fnt[:])
		for i := range l.wi {
			l.wi[i].draw(layerno, l.fnt[:])
		}
	}
	l.co.draw(layerno, l.fnt[:])
}
