package main

import (
	"fmt"
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
	hb.bg0 = *ReadAnimLayout(pre+"bg0.", is, sff, at)
	hb.bg1 = *ReadAnimLayout(pre+"bg1.", is, sff, at)
	hb.bg2 = *ReadAnimLayout(pre+"bg2.", is, sff, at)
	hb.mid = *ReadAnimLayout(pre+"mid.", is, sff, at)
	hb.front = *ReadAnimLayout(pre+"front.", is, sff, at)
	return hb
}
func (hb *HealthBar) step(life float32, gethit bool) {
	if len(hb.mid.anim.frames) > 0 && gethit {
		if hb.mlifetime > 0 && gethit {
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
	pb.bg0 = *ReadAnimLayout(pre+"bg0.", is, sff, at)
	pb.bg1 = *ReadAnimLayout(pre+"bg1.", is, sff, at)
	pb.bg2 = *ReadAnimLayout(pre+"bg2.", is, sff, at)
	pb.mid = *ReadAnimLayout(pre+"mid.", is, sff, at)
	pb.front = *ReadAnimLayout(pre+"front.", is, sff, at)
	is.ReadI32(pre+"counter.font", &pb.counter_font[0], &pb.counter_font[1],
		&pb.counter_font[2])
	pb.counter_lay = *ReadLayout(pre+"counter.", is)
	for i := range pb.level_snd {
		is.ReadI32(fmt.Sprintf("%slevel%d.snd", pre, i+1), &pb.level_snd[i][0],
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
		pb.snd.Play(pb.level_snd[i][0], pb.level_snd[i][1])
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
}

func newLifeBarFace() *LifeBarFace {
	return &LifeBarFace{face_spr: [2]int32{-1}, teammate_face_spr: [2]int32{-1}}
}
func readLifeBarFace(pre string, is IniSection,
	sff *Sff, at AnimationTable) *LifeBarFace {
	f := newLifeBarFace()
	is.ReadI32(pre+"pos", &f.pos[0], &f.pos[1])
	f.bg = *ReadAnimLayout(pre+"bg.", is, sff, at)
	is.ReadI32(pre+"face.spr", &f.face_spr[0], &f.face_spr[1])
	f.face_lay = *ReadLayout(pre+"face.", is)
	is.ReadI32(pre+"teammate.pos", &f.teammate_pos[0], &f.teammate_pos[1])
	is.ReadI32(pre+"teammate.spacing", &f.teammate_spacing[0],
		&f.teammate_spacing[1])
	f.teammate_bg = *ReadAnimLayout(pre+"teammate.bg.", is, sff, at)
	f.teammate_ko = *ReadAnimLayout(pre+"teammate.ko.", is, sff, at)
	is.ReadI32(pre+"teammate.face.spr", &f.teammate_face_spr[0],
		&f.teammate_face_spr[1])
	f.teammate_face_lay = *ReadLayout(pre+"teammate.face.", is)
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
	n.name_lay = *ReadLayout(pre+"name.", is)
	n.bg = *ReadAnimLayout(pre+"bg.", is, sff, at)
	return n
}
func (n *LifeBarName) step()  { n.bg.Action() }
func (n *LifeBarName) reset() { n.bg.Reset() }

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
	is.ReadI32(pre+"useiconupto", &wi.useiconupto)
	is.ReadI32(pre+"counter.font", &wi.counter_font[0], &wi.counter_font[1],
		&wi.counter_font[2])
	wi.counter_lay = *ReadLayout(pre+"counter.", is)
	wi.icon[WT_N] = *ReadAnimLayout(pre+"n.", is, sff, at)
	wi.icon[WT_S] = *ReadAnimLayout(pre+"s.", is, sff, at)
	wi.icon[WT_H] = *ReadAnimLayout(pre+"h.", is, sff, at)
	wi.icon[WT_C] = *ReadAnimLayout(pre+"c.", is, sff, at)
	wi.icon[WT_T] = *ReadAnimLayout(pre+"t.", is, sff, at)
	wi.icon[WT_Throw] = *ReadAnimLayout(pre+"throw.", is, sff, at)
	wi.icon[WT_Suicide] = *ReadAnimLayout(pre+"suicide.", is, sff, at)
	wi.icon[WT_Teammate] = *ReadAnimLayout(pre+"teammate.", is, sff, at)
	wi.icon[WT_Perfect] = *ReadAnimLayout(pre+"perfect.", is, sff, at)
	return wi
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
	t.counter_lay = *ReadLayout("counter.", is)
	t.bg = *ReadAnimLayout("bg.", is, sff, at)
	is.ReadI32("framespercount", &t.framespercount)
	return t
}
func (t *LifeBarTime) step()  { t.bg.Action() }
func (t *LifeBarTime) reset() { t.bg.Reset() }

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
	is.ReadBool("counter.shake", &c.counter_shake)
	c.counter_lay = *ReadLayout("counter.", is)
	c.counter_lay.offset = [2]float32{}
	is.ReadI32("text.font", &c.text_font[0], &c.text_font[1], &c.text_font[2])
	c.text_text = is["text.text"]
	c.text_lay = *ReadLayout("text.", is)
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
}

func newLifeBarRound(snd *Snd) *LifeBarRound {
	return &LifeBarRound{snd: snd, match_wins: 2, match_maxdrawgames: 1,
		start_waittime: 30, ctrl_time: 30, slow_time: 60, over_waittime: 45,
		over_hittime: 10, over_wintime: 45, over_time: 210, win_sndtime: 60}
}
func readLifeBarRound(is IniSection,
	sff *Sff, at AnimationTable, snd *Snd) *LifeBarRound {
	r := newLifeBarRound(snd)
	var tmp int32
	is.ReadI32("pos", &r.pos[0], &r.pos[1])
	is.ReadI32("match.wins", &r.match_wins)
	is.ReadI32("match.maxdrawgames", &r.match_maxdrawgames)
	if is.ReadI32("start.waittime", &tmp) {
		r.start_waittime = Max(1, tmp)
	}
	is.ReadI32("round.time", &r.round_time)
	is.ReadI32("round.sndtime", &r.round_sndtime)
	r.round_default = *ReadAnimTextSnd("round.default.", is, sff, at)
	for i := range r.round {
		r.round[i] = *ReadAnimTextSnd(fmt.Sprintf("round%d.", i+1), is, sff, at)
	}
	is.ReadI32("fight.time", &r.fight_time)
	is.ReadI32("fight.sndtime", &r.fight_sndtime)
	r.fight = *ReadAnimTextSnd("fight.", is, sff, at)
	if is.ReadI32("ctrl.time", &tmp) {
		r.ctrl_time = Max(1, tmp)
	}
	is.ReadI32("ko.time", &r.ko_time)
	is.ReadI32("ko.sndtime", &r.ko_sndtime)
	r.ko = *ReadAnimTextSnd("ko.", is, sff, at)
	r.dko = *ReadAnimTextSnd("dko.", is, sff, at)
	r.to = *ReadAnimTextSnd("to.", is, sff, at)
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
	r.win = *ReadAnimTextSnd("win.", is, sff, at)
	r.win2 = *ReadAnimTextSnd("win2.", is, sff, at)
	r.drawn = *ReadAnimTextSnd("draw.", is, sff, at)
	return r
}
func (r *LifeBarRound) callFight() {
	r.fight.Reset()
	r.cur, r.wt[0], r.swt[0], r.dt[0] = 1, r.fight_time, r.fight_sndtime, 0
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

type Lifebar struct {
	fat       AnimationTable
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
	l := &Lifebar{snd: &Snd{}, hb: [...][]*HealthBar{make([]*HealthBar, 2),
		make([]*HealthBar, 4), make([]*HealthBar, 2)},
		fa: [...][]*LifeBarFace{make([]*LifeBarFace, 2), make([]*LifeBarFace, 4),
			make([]*LifeBarFace, 2)},
		nm: [...][]*LifeBarName{make([]*LifeBarName, 2), make([]*LifeBarName, 4),
			make([]*LifeBarName, 2)}}
	sff, fsff, lines, i := &Sff{}, &Sff{}, SplitAndTrim(str, "\n"), 0
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
				l.ro = readLifeBarRound(is, sff, at, l.snd)
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
			cb[^i&1] = Min(999, Max(c.getcombo, cb[i&1]))
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
