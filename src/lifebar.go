package main

import (
	"fmt"
	"strings"
)

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
	at        *AnimationTable
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

func newHealthBar(sff *Sff, at *AnimationTable) (hb *HealthBar) {
	hb = &HealthBar{at: at, bg0: *newAnimation(sff), bg1: *newAnimation(sff),
		bg2: *newAnimation(sff), mid: *newAnimation(sff),
		front: *newAnimation(sff)}
	return
}
func readHealthBar(pre string, is IniSection,
	sff *Sff, at *AnimationTable) *HealthBar {
	hb := newHealthBar(sff, at)
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

type PowerBar struct {
	at        *AnimationTable
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

func newPowerBar(sff *Sff, at *AnimationTable) (pb *PowerBar) {
	pb = &PowerBar{at: at, bg0: *newAnimation(sff), bg1: *newAnimation(sff),
		bg2: *newAnimation(sff), mid: *newAnimation(sff),
		front: *newAnimation(sff)}
	return
}
func readPowerBar(pre string, is IniSection,
	sff *Sff, at *AnimationTable) *PowerBar {
	pb := newPowerBar(sff, at)
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
	return pb
}

type LifeBarFace struct{ at *AnimationTable }

func newLifeBarFace(sff *Sff, at *AnimationTable) *LifeBarFace {
	return &LifeBarFace{at: at}
}

type LifeBarName struct{ at *AnimationTable }

func newLifeBarName(sff *Sff, at *AnimationTable) *LifeBarName {
	return &LifeBarName{at: at}
}

type LifeBarWinIcon struct{ at *AnimationTable }

func newLifeBarWinIcon(sff *Sff, at *AnimationTable) *LifeBarWinIcon {
	return &LifeBarWinIcon{at: at}
}

type LifeBarTime struct{ at *AnimationTable }

func newLifeBarTime(sff *Sff, at *AnimationTable) *LifeBarTime {
	return &LifeBarTime{at: at}
}

type LifeBarCombo struct{}

func newLifeBarCombo() *LifeBarCombo {
	return &LifeBarCombo{}
}

type LifeBarRound struct{ at *AnimationTable }

func newLifeBarRound(sff *Sff, at *AnimationTable) *LifeBarRound {
	return &LifeBarRound{at: at}
}

type Lifebar struct {
	snd, fsnd *Snd
	at, fat   *AnimationTable
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
	l := &Lifebar{at: ReadAnimationTable(sff, lines, &i),
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
					l.snd, err = LoadSnd(filename)
					return err
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
		case "face":
		case "name":
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
			case len(subname) >= 4 && subname[:4] == "name":
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
			case len(subname) >= 4 && subname[:4] == "name":
			}
		case "winicon":
		case "time":
		case "combo":
		case "round":
		}
	}
	return l, nil
}
