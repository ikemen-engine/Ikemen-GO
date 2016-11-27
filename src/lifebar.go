package main

import (
	"fmt"
	"strings"
)

var lifebar Lifebar

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
func readSpr(name string, is IniSection, a *Animation) {
	a.frames = make([]AnimFrame, 1)
	var g, n int32 = -1, 0
	is.ReadI32(name, &g, &n)
	a.frames[0].Group, a.frames[0].Number = I32ToI16(g), I32ToI16(n)
	a.mask = 0
}
func readAnm(name string, is IniSection, a *Animation, at *AnimationTable) {
	n := int32(-1)
	is.ReadI32(name, &n)
	ani := at.get(n)
	if ani != nil {
		*a = *ani
	}
}

type HealthBar struct {
	pos       [2]int32
	range_x   [2]int32
	bg0       Animation
	bg0_lay   Layout
	bg1       Animation
	bg1_lay   Layout
	bg2       Animation
	bg2_lay   Layout
	mid       Animation
	mid_lay   Layout
	front     Animation
	front_lay Layout
}

func newHealthBar(sff *Sff) (hb *HealthBar) {
	hb = &HealthBar{bg0: *newAnimation(sff), bg1: *newAnimation(sff),
		bg2: *newAnimation(sff), mid: *newAnimation(sff),
		front: *newAnimation(sff)}
	return
}
func readHealthBar(pre string, is IniSection,
	sff *Sff, at *AnimationTable) *HealthBar {
	hb := newHealthBar(sff)
	is.ReadI32(pre+"pos", &hb.pos[0], &hb.pos[1])
	is.ReadI32(pre+"range.x", &hb.range_x[0], &hb.range_x[1])
	readSpr(pre+"bg0.spr", is, &hb.bg0)
	readAnm(pre+"bg0.anim", is, &hb.bg0, at)
	hb.bg0_lay = *readLayout(pre+"bg0.", is)
	readSpr(pre+"bg1.spr", is, &hb.bg1)
	readAnm(pre+"bg1.anim", is, &hb.bg1, at)
	hb.bg1_lay = *readLayout(pre+"bg1.", is)
	readSpr(pre+"bg2.spr", is, &hb.bg2)
	readAnm(pre+"bg2.anim", is, &hb.bg2, at)
	hb.bg2_lay = *readLayout(pre+"bg2.", is)
	readSpr(pre+"mid.spr", is, &hb.mid)
	readAnm(pre+"mid.anim", is, &hb.mid, at)
	hb.mid_lay = *readLayout(pre+"mid.", is)
	readSpr(pre+"front.spr", is, &hb.front)
	readAnm(pre+"front.anim", is, &hb.front, at)
	hb.front_lay = *readLayout(pre+"front.", is)
	return hb
}
func (hb *HealthBar) reset() {
	hb.bg0.reset()
	hb.bg1.reset()
	hb.bg2.reset()
	hb.mid.reset()
	hb.front.reset()
}

type PowerBar struct {
	snd          *Snd
	pos          [2]int32
	range_x      [2]int32
	bg0          Animation
	bg0_lay      Layout
	bg1          Animation
	bg1_lay      Layout
	bg2          Animation
	bg2_lay      Layout
	mid          Animation
	mid_lay      Layout
	front        Animation
	front_lay    Layout
	counter_font [3]int32
	counter_lay  Layout
	level_snd    [3][2]int32
}

func newPowerBar(sff *Sff, snd *Snd) (pb *PowerBar) {
	pb = &PowerBar{snd: snd, bg0: *newAnimation(sff),
		bg1: *newAnimation(sff), bg2: *newAnimation(sff), mid: *newAnimation(sff),
		front: *newAnimation(sff), counter_font: [3]int32{-1},
		level_snd: [3][2]int32{{-1}, {-1}, {-1}}}
	return
}
func readPowerBar(pre string, is IniSection,
	sff *Sff, at *AnimationTable, snd *Snd) *PowerBar {
	pb := newPowerBar(sff, snd)
	is.ReadI32(pre+"pos", &pb.pos[0], &pb.pos[1])
	is.ReadI32(pre+"range.x", &pb.range_x[0], &pb.range_x[1])
	readSpr(pre+"bg0.spr", is, &pb.bg0)
	readAnm(pre+"bg0.anim", is, &pb.bg0, at)
	pb.bg0_lay = *readLayout(pre+"bg0.", is)
	readSpr(pre+"bg1.spr", is, &pb.bg1)
	readAnm(pre+"bg1.anim", is, &pb.bg1, at)
	pb.bg1_lay = *readLayout(pre+"bg1.", is)
	readSpr(pre+"bg2.spr", is, &pb.bg2)
	readAnm(pre+"bg2.anim", is, &pb.bg2, at)
	pb.bg2_lay = *readLayout(pre+"bg2.", is)
	readSpr(pre+"mid.spr", is, &pb.mid)
	readAnm(pre+"mid.anim", is, &pb.mid, at)
	pb.mid_lay = *readLayout(pre+"mid.", is)
	readSpr(pre+"front.spr", is, &pb.front)
	readAnm(pre+"front.anim", is, &pb.front, at)
	pb.front_lay = *readLayout(pre+"front.", is)
	is.ReadI32(pre+"counter.font", &pb.counter_font[0], &pb.counter_font[1],
		&pb.counter_font[2])
	pb.counter_lay = *readLayout(pre+"counter.", is)
	for i := range pb.level_snd {
		is.ReadI32(fmt.Sprintf("%slevel%d.snd", pre, i+1), &pb.level_snd[i][0],
			&pb.level_snd[i][1])
	}
	return pb
}
func (pb *PowerBar) reset() {
	pb.bg0.reset()
	pb.bg1.reset()
	pb.bg2.reset()
	pb.mid.reset()
	pb.front.reset()
}

type LifeBarFace struct {
	pos               [2]int32
	bg                Animation
	bg_lay            Layout
	face_spr          [2]int32
	face              *Sprite
	face_lay          Layout
	teammate_pos      [2]int32
	teammate_spacing  [2]int32
	teammate_bg       Animation
	teammate_bg_lay   Layout
	teammate_ko       Animation
	teammate_ko_lay   Layout
	teammate_face_spr [2]int32
	teammate_face     []*Sprite
	teammate_face_lay Layout
}

func newLifeBarFace(sff *Sff) *LifeBarFace {
	return &LifeBarFace{bg: *newAnimation(sff), face_spr: [2]int32{-1},
		teammate_bg: *newAnimation(sff), teammate_ko: *newAnimation(sff),
		teammate_face_spr: [2]int32{-1}}
}
func readLifeBarFace(pre string, is IniSection,
	sff *Sff, at *AnimationTable) *LifeBarFace {
	f := newLifeBarFace(sff)
	is.ReadI32(pre+"pos", &f.pos[0], &f.pos[1])
	readSpr(pre+"bg.spr", is, &f.bg)
	readAnm(pre+"bg.anim", is, &f.bg, at)
	f.bg_lay = *readLayout(pre+"bg.", is)
	is.ReadI32(pre+"face.spr", &f.face_spr[0], &f.face_spr[1])
	f.face_lay = *readLayout(pre+"face.", is)
	is.ReadI32(pre+"teammate.pos", &f.teammate_pos[0], &f.teammate_pos[1])
	is.ReadI32(pre+"teammate.spacing", &f.teammate_spacing[0],
		&f.teammate_spacing[1])
	readSpr(pre+"teammate.bg.spr", is, &f.teammate_bg)
	readAnm(pre+"teammate.bg.anim", is, &f.teammate_bg, at)
	f.teammate_bg_lay = *readLayout(pre+"teammate.bg.", is)
	readSpr(pre+"teammate.ko.spr", is, &f.teammate_ko)
	readAnm(pre+"teammate.ko.anim", is, &f.teammate_ko, at)
	f.teammate_ko_lay = *readLayout(pre+"teammate.ko.", is)
	is.ReadI32(pre+"teammate.face.spr", &f.teammate_face_spr[0],
		&f.teammate_face_spr[1])
	f.teammate_face_lay = *readLayout(pre+"teammate.face.", is)
	return f
}
func (f *LifeBarFace) reset() {
	f.bg.reset()
	f.teammate_bg.reset()
	f.teammate_ko.reset()
}

type LifeBarName struct {
	pos       [2]int32
	name_font [3]int32
	name_lay  Layout
	bg        Animation
	bg_lay    Layout
}

func newLifeBarName(sff *Sff) *LifeBarName {
	return &LifeBarName{name_font: [3]int32{-1}, bg: *newAnimation(sff)}
}
func readLifeBarName(pre string, is IniSection,
	sff *Sff, at *AnimationTable) *LifeBarName {
	n := newLifeBarName(sff)
	is.ReadI32(pre+"pos", &n.pos[0], &n.pos[1])
	is.ReadI32(pre+"name.font", &n.name_font[0], &n.name_font[1],
		&n.name_font[2])
	n.name_lay = *readLayout(pre+"name.", is)
	readSpr(pre+"bg.spr", is, &n.bg)
	readAnm(pre+"bg.anim", is, &n.bg, at)
	n.bg_lay = *readLayout(pre+"bg.", is)
	return n
}
func (n *LifeBarName) reset() { n.bg.reset() }

type LifeBarWinIcon struct {
	pos           [2]int32
	iconoffset    [2]int32
	useiconupto   int32
	counter_font  [3]int32
	counter_lay   Layout
	icon          [WT_NumTypes]Animation
	icon_lay      [WT_NumTypes]Layout
	wins          []WinType
	numWins       int
	added, addedp *Animation
}

func newLifeBarWinIcon(sff *Sff) (wi *LifeBarWinIcon) {
	wi = &LifeBarWinIcon{useiconupto: 4, counter_font: [3]int32{-1}}
	for i := range wi.icon {
		wi.icon[i] = *newAnimation(sff)
	}
	return
}
func readLifeBarWinIcon(pre string, is IniSection,
	sff *Sff, at *AnimationTable) *LifeBarWinIcon {
	wi := newLifeBarWinIcon(sff)
	is.ReadI32(pre+"pos", &wi.pos[0], &wi.pos[1])
	is.ReadI32(pre+"iconoffset", &wi.iconoffset[0], &wi.iconoffset[1])
	is.ReadI32(pre+"useiconupto", &wi.useiconupto)
	is.ReadI32(pre+"counter.font", &wi.counter_font[0], &wi.counter_font[1],
		&wi.counter_font[2])
	wi.counter_lay = *readLayout(pre+"counter.", is)
	readSpr(pre+"n.spr", is, &wi.icon[WT_N])
	readAnm(pre+"n.anim", is, &wi.icon[WT_N], at)
	wi.icon_lay[WT_N] = *readLayout(pre+"n.", is)
	readSpr(pre+"s.spr", is, &wi.icon[WT_S])
	readAnm(pre+"s.anim", is, &wi.icon[WT_S], at)
	wi.icon_lay[WT_S] = *readLayout(pre+"s.", is)
	readSpr(pre+"h.spr", is, &wi.icon[WT_H])
	readAnm(pre+"h.anim", is, &wi.icon[WT_H], at)
	wi.icon_lay[WT_H] = *readLayout(pre+"h.", is)
	readSpr(pre+"c.spr", is, &wi.icon[WT_C])
	readAnm(pre+"c.anim", is, &wi.icon[WT_C], at)
	wi.icon_lay[WT_C] = *readLayout(pre+"c.", is)
	readSpr(pre+"t.spr", is, &wi.icon[WT_T])
	readAnm(pre+"t.anim", is, &wi.icon[WT_T], at)
	wi.icon_lay[WT_T] = *readLayout(pre+"t.", is)
	readSpr(pre+"throw.spr", is, &wi.icon[WT_Throw])
	readAnm(pre+"throw.anim", is, &wi.icon[WT_Throw], at)
	wi.icon_lay[WT_Throw] = *readLayout(pre+"throw.", is)
	readSpr(pre+"suicide.spr", is, &wi.icon[WT_Suicide])
	readAnm(pre+"suicide.anim", is, &wi.icon[WT_Suicide], at)
	wi.icon_lay[WT_Suicide] = *readLayout(pre+"suicide.", is)
	readSpr(pre+"teammate.spr", is, &wi.icon[WT_Teammate])
	readAnm(pre+"teammate.anim", is, &wi.icon[WT_Teammate], at)
	wi.icon_lay[WT_Teammate] = *readLayout(pre+"teammate.", is)
	readSpr(pre+"perfect.spr", is, &wi.icon[WT_Perfect])
	readAnm(pre+"perfect.anim", is, &wi.icon[WT_Perfect], at)
	wi.icon_lay[WT_Perfect] = *readLayout(pre+"perfect.", is)
	return wi
}
func (wi *LifeBarWinIcon) reset() {
	for i := range wi.icon {
		wi.icon[i].reset()
	}
	wi.numWins = len(wi.wins)
	wi.added, wi.addedp = nil, nil
}
func (wi *LifeBarWinIcon) clear() { wi.wins = nil }

type LifeBarTime struct {
	pos            [2]int32
	counter_font   [3]int32
	counter_lay    Layout
	bg             Animation
	bg_lay         Layout
	framespercount int32
}

func newLifeBarTime(sff *Sff) *LifeBarTime {
	return &LifeBarTime{counter_font: [3]int32{-1}, bg: *newAnimation(sff),
		framespercount: 60}
}
func readLifeBarTime(is IniSection,
	sff *Sff, at *AnimationTable) *LifeBarTime {
	t := newLifeBarTime(sff)
	is.ReadI32("pos", &t.pos[0], &t.pos[1])
	is.ReadI32("counter.font", &t.counter_font[0], &t.counter_font[1],
		&t.counter_font[2])
	t.counter_lay = *readLayout("counter.", is)
	readSpr("bg.spr", is, &t.bg)
	readAnm("bg.anim", is, &t.bg, at)
	t.bg_lay = *readLayout("bg.", is)
	is.ReadI32("framespercount", &t.framespercount)
	return t
}
func (t *LifeBarTime) reset() { t.bg.reset() }

type LifeBarCombo struct {
	pos           [2]int32
	start_x       float32
	counter_font  [3]int32
	counter_shake int32
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
	return &LifeBarCombo{counter_font: [3]int32{-1},
		counter_lay: Layout{layerno: 2},
		text_font:   [3]int32{-1}, text_lay: Layout{layerno: 2}, displaytime: 90}
}
func readLifeBarCombo(is IniSection) *LifeBarCombo {
	c := newLifeBarCombo()
	is.ReadI32("pos", &c.pos[0], &c.pos[1])
	is.ReadF32("start.x", &c.start_x)
	is.ReadI32("counter.font", &c.counter_font[0], &c.counter_font[1],
		&c.counter_font[2])
	is.ReadI32("counter.shake", &c.counter_shake)
	c.counter_lay = *readLayout("counter.", is)
	c.counter_lay.offset = [2]float32{0, 0}
	is.ReadI32("text.font", &c.text_font[0], &c.text_font[1], &c.text_font[2])
	c.text_text = is["text.text"]
	c.text_lay = *readLayout("text.", is)
	is.ReadI32("displaytime", &c.displaytime)
	return c
}
func (c *LifeBarCombo) reset() {
	c.cur, c.old, c.resttime = [2]int32{}, [2]int32{}, [2]int32{}
	c.counterX = [2]float32{c.start_x * 2, c.start_x * 2}
	c.shaketime = [2]int32{}
}

type LifeBarRound struct{}

func newLifeBarRound(sff *Sff) *LifeBarRound {
	return &LifeBarRound{}
}

type Lifebar struct {
	at, fat   *AnimationTable
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

func LoadLifebar(deffile string) (*Lifebar, error) {
	str, err := LoadText(deffile)
	if err != nil {
		return nil, err
	}
	sff, fsff, lines, i := &Sff{}, &Sff{}, SplitAndTrim(str, "\n"), 0
	l := &Lifebar{at: ReadAnimationTable(sff, lines, &i), snd: &Snd{},
		hb: [3][]*HealthBar{make([]*HealthBar, 2), make([]*HealthBar, 4),
			make([]*HealthBar, 2)},
		fa: [3][]*LifeBarFace{make([]*LifeBarFace, 2), make([]*LifeBarFace, 4),
			make([]*LifeBarFace, 2)},
		nm: [3][]*LifeBarName{make([]*LifeBarName, 2), make([]*LifeBarName, 4),
			make([]*LifeBarName, 2)}}
	i = 0
	filesflg := true
	for i < len(lines) {
		is, name, subname := ReadIniSection(lines, &i)
		switch name {
		case "files":
			if filesflg {
				filesflg = false
				if is.LoadFile("sff", deffile, func(filename string) error {
					s, err := LoadSff(filename, false)
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
					s, err := LoadSff(filename, false)
					if err != nil {
						return err
					}
					*fsff = *s
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
					l.fat = ReadAnimationTable(fsff, lines, &i)
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
					if is.LoadFile(fmt.Sprintf("font%d", i), deffile,
						func(filename string) error {
							l.fnt[i], err = LoadFnt(filename)
							return err
						}); err != nil {
						return nil, err
					}
				}
			}
		case "lifebar":
			if l.hb[0][0] == nil {
				l.hb[0][0] = readHealthBar("p1.", is, sff, l.at)
			}
			if l.hb[0][1] == nil {
				l.hb[0][1] = readHealthBar("p2.", is, sff, l.at)
			}
		case "powerbar":
			if l.pb[0] == nil {
				l.pb[0] = readPowerBar("p1.", is, sff, l.at, l.snd)
			}
			if l.pb[1] == nil {
				l.pb[1] = readPowerBar("p2.", is, sff, l.at, l.snd)
			}
		case "face":
			if l.fa[0][0] == nil {
				l.fa[0][0] = readLifeBarFace("p1.", is, sff, l.at)
			}
			if l.fa[0][1] == nil {
				l.fa[0][1] = readLifeBarFace("p2.", is, sff, l.at)
			}
		case "name":
			if l.nm[0][0] == nil {
				l.nm[0][0] = readLifeBarName("p1.", is, sff, l.at)
			}
			if l.nm[0][1] == nil {
				l.nm[0][1] = readLifeBarName("p2.", is, sff, l.at)
			}
		case "simul ":
			subname = strings.ToLower(subname)
			switch {
			case len(subname) >= 7 && subname[:7] == "lifebar":
				if l.hb[1][0] == nil {
					l.hb[1][0] = readHealthBar("p1.", is, sff, l.at)
				}
				if l.hb[1][1] == nil {
					l.hb[1][1] = readHealthBar("p2.", is, sff, l.at)
				}
				if l.hb[1][2] == nil {
					l.hb[1][2] = readHealthBar("p3.", is, sff, l.at)
				}
				if l.hb[1][3] == nil {
					l.hb[1][3] = readHealthBar("p4.", is, sff, l.at)
				}
			case len(subname) >= 4 && subname[:4] == "face":
				if l.fa[1][0] == nil {
					l.fa[1][0] = readLifeBarFace("p1.", is, sff, l.at)
				}
				if l.fa[1][1] == nil {
					l.fa[1][1] = readLifeBarFace("p2.", is, sff, l.at)
				}
				if l.fa[1][2] == nil {
					l.fa[1][2] = readLifeBarFace("p3.", is, sff, l.at)
				}
				if l.fa[1][3] == nil {
					l.fa[1][3] = readLifeBarFace("p4.", is, sff, l.at)
				}
			case len(subname) >= 4 && subname[:4] == "name":
				if l.nm[1][0] == nil {
					l.nm[1][0] = readLifeBarName("p1.", is, sff, l.at)
				}
				if l.nm[1][1] == nil {
					l.nm[1][1] = readLifeBarName("p2.", is, sff, l.at)
				}
				if l.nm[1][2] == nil {
					l.nm[1][2] = readLifeBarName("p3.", is, sff, l.at)
				}
				if l.nm[1][3] == nil {
					l.nm[1][3] = readLifeBarName("p4.", is, sff, l.at)
				}
			}
		case "turns ":
			subname = strings.ToLower(subname)
			switch {
			case len(subname) >= 7 && subname[:7] == "lifebar":
				if l.hb[2][0] == nil {
					l.hb[2][0] = readHealthBar("p1.", is, sff, l.at)
				}
				if l.hb[2][1] == nil {
					l.hb[2][1] = readHealthBar("p2.", is, sff, l.at)
				}
			case len(subname) >= 4 && subname[:4] == "face":
				if l.fa[2][0] == nil {
					l.fa[2][0] = readLifeBarFace("p1.", is, sff, l.at)
				}
				if l.fa[2][1] == nil {
					l.fa[2][1] = readLifeBarFace("p2.", is, sff, l.at)
				}
			case len(subname) >= 4 && subname[:4] == "name":
				if l.nm[2][0] == nil {
					l.nm[2][0] = readLifeBarName("p1.", is, sff, l.at)
				}
				if l.nm[2][1] == nil {
					l.nm[2][1] = readLifeBarName("p2.", is, sff, l.at)
				}
			}
		case "winicon":
			if l.wi[0] == nil {
				l.wi[0] = readLifeBarWinIcon("p1.", is, sff, l.at)
			}
			if l.wi[1] == nil {
				l.wi[1] = readLifeBarWinIcon("p2.", is, sff, l.at)
			}
		case "time":
			if l.ti == nil {
				l.ti = readLifeBarTime(is, sff, l.at)
			}
		case "combo":
			if l.co == nil {
				l.co = readLifeBarCombo(is)
			}
		case "round":
		}
	}
	return l, nil
}
