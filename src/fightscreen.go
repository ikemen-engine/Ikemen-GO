package main

import (
	"fmt"
	"math"
	"path/filepath"
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
	WT_Normal WinType = iota
	WT_Special
	WT_Hyper
	WT_Cheese
	WT_Time
	WT_Throw
	WT_Suicide
	WT_Teammate
	WT_Perfect
	WT_Clutch
	WT_NumTypes
	WT_PNormal
	WT_PSpecial
	WT_PHyper
	WT_PCheese
	WT_PTime
	WT_PThrow
	WT_PSuicide
	WT_PTeammate
	WT_CNormal
	WT_CSpecial
	WT_CHyper
	WT_CCheese
	WT_CTime
	WT_CThrow
	WT_CSuicide
	WT_CTeammate
)

func (wt *WinType) SetPerfect() {
	if *wt >= WT_Normal && *wt < WT_Perfect {
		*wt += WT_PNormal - WT_Normal
	}
}

func (wt *WinType) SetClutch() {
	if *wt >= WT_Normal && *wt < WT_Perfect {
		*wt += WT_CNormal - WT_Normal
	}
}

type FightFx struct {
	animTable  AnimationTable
	fileName   string
	sff        *Sff
	snd        *Snd
	fx_scale   float32
	localcoord [2]int32
	refCount   int
	isCharFX   bool
}

func newFightFx() *FightFx {
	return &FightFx{
		sff:        &Sff{},
		fx_scale:   1.0,
		localcoord: sys.fightScreen.localcoord,
	}
}

func loadCommonFightFx(def string, isMainThread bool) error {
	for _, key := range SortedKeys(sys.cfg.Common.Fx) {
		for _, v := range sys.cfg.Common.Fx[key] {
			if err := LoadFile(&v, []string{def, sys.motif.Def, "", "data/"}, "",
				func(filename string) error {
					for _, ffx := range sys.ffx {
						if ffx != nil && !ffx.isCharFX && ffx.fileName == filename {
							return nil
						}
					}
					return loadFightFx(filename, false, isMainThread)
				}); err != nil {
				return err
			}
		}
	}
	return nil
}

func loadFightFx(def string, isCharFX bool, isMainThread bool) error {
	str, err := LoadText(def)
	if err != nil {
		return err
	}
	ffx := newFightFx()
	prefix := ""
	lines, i := SplitAndTrim(str, "\n"), 0
	info, files := true, true
	for i < len(lines) {
		// Parse each ini section
		is, name, _ := ReadIniSection(lines, &i)
		switch name {
		case "info":
			// Read info for FightFx storing and scaling
			if info {
				info = false
				var ok bool
				prefix, ok, _ = is.getText("prefix")
				if !ok || prefix == "" {
					return Error("A prefix must be declared")
				}
				prefix = strings.ToLower(prefix)
				// Check if prefix overlaps a reserved one
				if prefix == "f" || prefix == "s" {
					return Error(fmt.Sprintf("The %s prefix is reserved for the system and cannot be used", strings.ToUpper(prefix)))
				}
				// Check if prefix conflicts with trigger names
				if _, ok := triggerMap[prefix]; ok {
					return Error(fmt.Sprintf("The %s prefix conflicts with an existing trigger name and cannot be used", strings.ToUpper(prefix)))
				}
				if ffx, ok := sys.ffx[prefix]; ok {
					// Char FX are the only refcounted FX. Everything else is cached once loaded and kept around.
					if isCharFX && ffx.isCharFX {
						if ffx.refCount < 8 {
							ffx.refCount += 1 + int((sys.numSimul[0]+sys.numSimul[1])/2)
						}
					}
					// If the same file later appears in Common.Fx, promote it to
					// cached non-char FX instead of letting refcount cleanup remove it.
					if !isCharFX && ffx.fileName == def {
						ffx.isCharFX = false
						ffx.refCount = 99999
					}
					if ffx.fileName != def {
						sys.appendToConsole(fmt.Sprintf(
							"Duplicate FightFX prefix found: %s already loaded from %s, ignoring %s",
							strings.ToUpper(prefix), ffx.fileName, def))
					}
					return nil
				}
				is.ReadF32("fx.scale", &ffx.fx_scale)
				is.ReadI32("localcoord", &ffx.localcoord[0], &ffx.localcoord[1])
			}
		case "files":
			// Read files section
			if files {
				files = false
				if is.LoadFile("sff", []string{def, sys.motif.Def, "", "data/"}, "",
					func(filename string) error {
						s, err := loadSff(filename, false, isMainThread, false)
						if err != nil {
							return err
						}
						*ffx.sff = *s
						return nil
					}); err != nil {
					return err
				}
				if is.LoadFile("air", []string{def, sys.motif.Def, "", "data/"}, "",
					func(filename string) error {
						str, err := LoadText(filename)
						if err != nil {
							return err
						}
						lines, i := SplitAndTrim(str, "\n"), 0
						ffx.animTable = ReadAnimationTable(filename, ffx.sff, &ffx.sff.palList, lines, &i, true)
						return nil
					}); err != nil {
					return err
				}
				if is.LoadFile("snd", []string{def, sys.motif.Def, "", "data/"}, "",
					func(filename string) error {
						ffx.snd, err = LoadSnd(filename)
						return err
					}); err != nil {
					return err
				}
			}
		}
	}
	// Set fx scale to anims
	// Now calculated in each animation call
	//for _, a := range ffx.animTable {
	//	a.start_scale = [...]float32{ffx.fx_scale, ffx.fx_scale}
	//}
	// Adding used prefixes to a list is no longer necessary
	// We will check if they're being used in the ffx map directly
	//if sys.ffx[prefix] == nil {
	//	sys.ffxRegexp += "|^(" + prefix + ")"
	//	sys.ffxPrefixes = append(sys.ffxPrefixes, prefix)
	//}
	ffx.fileName = def
	ffx.isCharFX = isCharFX
	if isCharFX {
		ffx.refCount = 3 + int(sys.numSimul[0]+sys.numSimul[1])
	} else {
		ffx.refCount = 99999
	}
	sys.ffx[prefix] = ffx
	return nil
}

type FSText struct {
	font       [8]int32 // to match Lua arg count regardless
	text       string
	lay        Layout
	palfx      *PalFX
	frgba      [4]float32 // ttf fonts
	forcecolor bool
	pfxinit    int32
}

func newFSText(align int32) *FSText {
	return &FSText{
		font:  [...]int32{-1, 0, align, 255, 255, 255, 255, -1},
		palfx: newPalFX(),
		frgba: [...]float32{1.0, 1.0, 1.0, 1.0},
	}
}

// helper to safely obtain a font from a map
func getFont(f map[int]*Fnt, idx int32) *Fnt {
	if idx < 0 {
		return nil
	}
	if ff, ok := f[int(idx)]; ok {
		return ff
	}
	return nil
}

func readFSText(pre string, is IniSection, str string, ln int16, f map[int]*Fnt, align int32) *FSText {
	txt := newFSText(align)

	txt.font[3], txt.font[4], txt.font[5], txt.font[6], txt.font[7] = -1, -1, -1, 255, -1
	is.ReadI32(pre+"font", &txt.font[0], &txt.font[1], &txt.font[2],
		&txt.font[3], &txt.font[4], &txt.font[5], &txt.font[6], &txt.font[7])
	if txt.font[0] >= 0 && getFont(f, txt.font[0]) == nil {
		LogMessage("Undefined font %v referenced by fight screen parameter: %v", txt.font[0], pre+"font")
		txt.font[0] = -1
	}
	if _, ok := is[pre+"text"]; ok {
		txt.text, _, _ = is.getText(pre + "text")
	} else {
		txt.text = str
	}
	txt.lay = *ReadLayout(pre, is, ln)
	if txt.font[3] >= 0 && txt.font[4] >= 0 && txt.font[5] >= 0 {
		txt.SetColor(txt.font[3], txt.font[4], txt.font[5], txt.font[6])
	}
	txt.pfxinit = ReadPalFX(pre+"palfx.", is, txt.palfx)
	return txt
}

func (txt *FSText) SetColor(r, g, b, a int32) {
	txt.forcecolor = true
	txt.palfx.setColor(r, g, b)
	txt.frgba = [...]float32{float32(r) / 255, float32(g) / 255,
		float32(b) / 255, float32(a) / 255}
}

func (txt *FSText) step() {
	if txt.palfx != nil && !txt.forcecolor {
		txt.palfx.step()
	}
}

func (txt *FSText) resetTxtPfx() {
	txt.palfx.time = txt.pfxinit
}

type FSBgTextSnd struct {
	pos         [2]int32
	text        FSText
	bg          AnimLayout
	time        int32
	displaytime int32
	snd         [2]int32
	sndtime     int32
	timer       int32
}

func newFSBgTextSnd() FSBgTextSnd {
	return FSBgTextSnd{snd: [2]int32{-1}}
}

func readFSBgTextSnd(pre string, is IniSection, sff *Sff, at AnimationTable, ln int16, f map[int]*Fnt) FSBgTextSnd {
	bts := newFSBgTextSnd()

	is.ReadI32(pre+"pos", &bts.pos[0], &bts.pos[1])
	bts.text = *readFSText(pre+"text.", is, "", ln, f, 0)
	bts.bg = ReadAnimLayout(pre+"bg.", is, sff, at, ln)
	is.ReadI32(pre+"time", &bts.time)
	is.ReadI32(pre+"displaytime", &bts.displaytime)
	is.ReadI32(pre+"snd", &bts.snd[0], &bts.snd[1])
	bts.sndtime = bts.time
	is.ReadI32(pre+"sndtime", &bts.sndtime)
	return bts
}

func (bts *FSBgTextSnd) step(snd *Snd) {
	if bts.timer == bts.sndtime {
		snd.play(bts.snd, 100, 0, 0, 0, 0)
	}
	if bts.timer >= bts.time {
		bts.bg.Action()
	}
	bts.timer++
	bts.text.step()
}

func (bts *FSBgTextSnd) reset() {
	bts.timer = 0
	bts.bg.Reset()
	bts.text.resetTxtPfx()
}

func (bts *FSBgTextSnd) bgDraw(layerno int16) {
	if bts.timer > bts.time && bts.timer <= bts.time+bts.displaytime {
		bts.bg.Draw(float32(bts.pos[0])+sys.fightScreen.offsetX, float32(bts.pos[1]), layerno, sys.fightScreen.scale)
	}
}

func (bts *FSBgTextSnd) draw(layerno int16, f map[int]*Fnt) {
	if bts.timer > bts.time && bts.timer <= bts.time+bts.displaytime &&
		bts.text.font[0] >= 0 && getFont(f, bts.text.font[0]) != nil {
		bts.text.lay.DrawText(float32(bts.pos[0])+sys.fightScreen.offsetX, float32(bts.pos[1]), sys.fightScreen.scale, layerno,
			bts.text.text, getFont(f, bts.text.font[0]), bts.text.font[1], bts.text.font[2], bts.text.palfx, bts.text.frgba)
	}
}

// Reads multiple bar values e.g. multiple front elements
func readMultipleValues(pre string, name string, is IniSection, sff *Sff, at AnimationTable) map[int32]*AnimLayout {
	result := make(map[int32]*AnimLayout)
	r, _ := regexp.Compile(pre + name + "[0-9]+\\.")
	for k := range is {
		if r.MatchString(k) {
			re := regexp.MustCompile("[0-9]+")
			submatchall := re.FindAllString(k, -1)
			if len(submatchall) == 2 {
				v := Atoi(submatchall[1])
				if _, ok := result[v]; !ok {
					tmp := ReadAnimLayout(pre+name+fmt.Sprintf("%v", v)+".", is, sff, at, 0)
					result[v] = &tmp
				}
			}
		}
	}
	return result
}

// Float version of readMultipleValues
func readMultipleValuesF(pre string, name string, is IniSection, sff *Sff, at AnimationTable) map[float32]*AnimLayout {
	result := make(map[float32]*AnimLayout)
	r, _ := regexp.Compile(pre + name + "[0-9]+\\.")
	for k := range is {
		if r.MatchString(k) {
			re := regexp.MustCompile("[0-9]+")
			submatchall := re.FindAllString(k, -1)
			if len(submatchall) == 2 {
				v := Atof(submatchall[1])
				key := float32(v)
				if _, ok := result[key]; !ok {
					tmp := ReadAnimLayout(pre+name+fmt.Sprintf("%v", v)+".", is, sff, at, 0)
					result[key] = &tmp
				}
			}
		}
	}
	return result
}

// Text version of readMultipleValues
func readMultipleFSText(pre string, name string, is IniSection, fmtstr string, ln int16, f map[int]*Fnt, align int32) map[int32]*FSText {
	result := make(map[int32]*FSText)
	r, _ := regexp.Compile(pre + name + "[0-9]+\\.")
	for k := range is {
		if r.MatchString(k) {
			re := regexp.MustCompile("[0-9]+")
			submatchall := re.FindAllString(k, -1)
			if len(submatchall) >= 1 {
				v := Atoi(submatchall[len(submatchall)-1])
				if _, ok := result[v]; !ok {
					result[v] = readFSText(pre+name+fmt.Sprintf("%v", v)+".", is, fmtstr, ln, f, align)
				}
			}
		}
	}
	return result
}

// Calculates the visible portion (rect) of a fill bar element
func calcBarFillRect(pos int32, range_ [2]int32, offset, scale, screenScale, midPos float32, fill float32) (start, size int32) {
	isDescending := range_[0] > range_[1]
	var r0, r1 int32

	if isDescending {
		r0, r1 = range_[1], range_[0]
	} else {
		r0, r1 = range_[0], range_[1]
	}

	fillLength := float32(r1 - r0 + 1)
	size = int32((fillLength * scale * fill * screenScale) + 0.5)

	base := float32(pos + r0)
	start = int32(((base+offset)*scale+midPos)*screenScale + 0.5)

	if isDescending {
		start = int32(((float32(pos+r1+1)+offset)*scale+midPos)*screenScale+0.5) - size
	}
	return
}

type LifeBar struct {
	pos         [2]int32
	range_x     [2]int32
	range_y     [2]int32
	bg0         AnimLayout
	bg1         AnimLayout
	bg2         AnimLayout
	top         AnimLayout
	mid         AnimLayout
	red         map[int32]*AnimLayout
	front       map[float32]*AnimLayout
	shift       AnimLayout
	warn        AnimLayout
	warn_range  [2]int32
	value       map[int32]*FSText
	red_value   map[int32]*FSText
	toplife     float32
	oldlife     float32
	midlife     float32
	midlifeMin  float32
	mlifetime   int32
	mid_shift   bool
	mid_freeze  bool
	mid_delay   int32
	mid_mult    float32
	mid_steps   float32
	gethit      bool
	scalefill   bool
	leaderontop bool
}

func newLifeBar() *LifeBar {
	return &LifeBar{
		oldlife:    1,
		midlife:    1,
		midlifeMin: 1,
		red:        make(map[int32]*AnimLayout),
		front:      make(map[float32]*AnimLayout),
		value:      make(map[int32]*FSText),
		red_value:  make(map[int32]*FSText),
		mid_freeze: true,
		mid_delay:  30,
		mid_mult:   1.0,
		mid_steps:  8.0,
	}
}

func readLifeBar(pre string, is IniSection, sff *Sff, at AnimationTable, f map[int]*Fnt) *LifeBar {
	lb := newLifeBar()

	is.ReadI32(pre+"pos", &lb.pos[0], &lb.pos[1])
	is.ReadI32(pre+"range.x", &lb.range_x[0], &lb.range_x[1])
	is.ReadI32(pre+"range.y", &lb.range_y[0], &lb.range_y[1])

	// Single value layouts
	lb.bg0 = ReadAnimLayout(pre+"bg0.", is, sff, at, 0)
	lb.bg1 = ReadAnimLayout(pre+"bg1.", is, sff, at, 0)
	lb.bg2 = ReadAnimLayout(pre+"bg2.", is, sff, at, 0)
	lb.top = ReadAnimLayout(pre+"top.", is, sff, at, 0)
	lb.mid = ReadAnimLayout(pre+"mid.", is, sff, at, 0)
	lb.shift = ReadAnimLayout(pre+"shift.", is, sff, at, 0)
	lb.warn = ReadAnimLayout(pre+"warn.", is, sff, at, 0)

	// Multiple value layouts
	lb.front = readMultipleValuesF(pre, "front", is, sff, at)
	if _, ok := lb.front[0]; !ok { // Set a value for 0 if it's missing
		tmp := ReadAnimLayout(pre+"front.", is, sff, at, 0)
		lb.front[0] = &tmp
	}

	lb.red = readMultipleValues(pre, "red", is, sff, at)
	if _, ok := lb.red[0]; !ok {
		tmp := ReadAnimLayout(pre+"red.", is, sff, at, 0)
		lb.red[0] = &tmp
	}

	// Text fields
	lb.value[0] = readFSText(pre+"value.", is, "%d", 0, f, 0)
	for k, v := range readMultipleFSText(pre, "value", is, "%d", 0, f, 0) {
		lb.value[k] = v
	}

	lb.red_value[0] = readFSText(pre+"red.value.", is, "%d", 0, f, 0)
	for k, v := range readMultipleFSText(pre, "red.value", is, "%d", 0, f, 0) {
		lb.red_value[k] = v
	}

	// TODO: These probably didn't have player prefixes for the sake of making parameters less redundant
	// With the addition of player operators we could deprecate these forms now and use prefixes as well
	is.ReadBool("mid.shift", &lb.mid_shift)
	is.ReadBool("mid.freeze", &lb.mid_freeze)
	is.ReadI32("mid.delay", &lb.mid_delay)
	is.ReadF32("mid.mult", &lb.mid_mult)
	is.ReadF32("mid.steps", &lb.mid_steps)
	lb.mid_steps = Max(1, lb.mid_steps)

	is.ReadI32(pre+"warn.range", &lb.warn_range[0], &lb.warn_range[1])
	is.ReadBool(pre+"scalefill", &lb.scalefill)
	is.ReadBool("leaderontop", &lb.leaderontop)

	return lb
}

func (lb *LifeBar) step(charpn int, lbr *LifeBar) {
	refChar := sys.chars[charpn][0]
	var life float32 = float32(refChar.life) / float32(refChar.lifeMax)
	var redVal int32 = refChar.redLife - refChar.life
	var getHit bool = (refChar.receivedHits != 0 || refChar.ss.moveType == MT_H) && !refChar.scf(SCF_over_ko)

	if lbr.toplife > life {
		lbr.toplife += (life - lbr.toplife) / 2
	} else {
		lbr.toplife = life
	}

	// Element shifting gradient
	lb.shift.anim.srcAlpha = int16(255 * (1 - life))
	lb.shift.anim.dstAlpha = int16(255 * life)

	if !lb.mid_freeze && getHit && !lb.gethit && len(lb.mid.anim.frames) > 0 {
		lbr.mlifetime = lb.mid_delay
		lbr.midlife = lbr.oldlife
		lbr.midlifeMin = lbr.oldlife
	}
	lb.gethit = getHit
	if lb.mid_freeze && getHit && len(lb.mid.anim.frames) > 0 {
		if lbr.mlifetime < lb.mid_delay {
			lbr.mlifetime = lb.mid_delay
			lbr.midlife = lbr.oldlife
			lbr.midlifeMin = lbr.oldlife
		}
	} else {
		if lbr.mlifetime > 0 {
			lbr.mlifetime--
		}
		if len(lb.mid.anim.frames) > 0 && lbr.mlifetime <= 0 && life < lbr.midlifeMin {
			lbr.midlifeMin += (life - lbr.midlifeMin) * (1 / (12 - (life-lbr.midlifeMin)*144)) * lb.mid_mult
			if lbr.midlifeMin < life {
				lbr.midlifeMin = life
			}
		} else {
			lbr.midlifeMin = life
		}
		if (len(lb.mid.anim.frames) == 0 || lbr.mlifetime <= 0) && lbr.midlife > lbr.midlifeMin {
			lbr.midlife += (lbr.midlifeMin - lbr.midlife) / lb.mid_steps
		}
		lbr.oldlife = life
	}

	mlmin := Max(lbr.midlifeMin, life)
	if lbr.midlife < mlmin {
		lbr.midlife += (mlmin - lbr.midlife) / 2
	}

	lb.bg0.Action()
	lb.bg1.Action()
	lb.bg2.Action()
	lb.top.Action()
	lb.mid.Action()
	// Multiple front elements - red life
	if refChar.redLifeEnabled() {
		var rv int32
		for k := range lb.red {
			if k > rv && redVal >= k {
				rv = k
			}
		}
		lb.red[rv].Action()

		// Multiple red_value fonts - life
		var rv2 int32
		for k := range lb.red_value {
			if k > rv2 && redVal >= k {
				rv2 = k
			}
		}
		lb.red_value[rv2].step()
	}

	// Multiple front elements - life
	var fv float32
	for k := range lb.front {
		if k > fv && life >= k/100 {
			fv = k
		}
	}
	lb.front[fv].Action()

	lb.shift.Action()
	lb.warn.Action()

	// Multiple value fonts - life
	var fv2 int32
	for k := range lb.value {
		if k > fv2 && life >= float32(k)/100 {
			fv2 = k
		}
	}
	lb.value[fv2].step()
}

func (lb *LifeBar) reset() {
	lb.bg0.Reset()
	lb.bg1.Reset()
	lb.bg2.Reset()
	lb.top.Reset()
	lb.mid.Reset()
	for i := range lb.front {
		lb.front[i].Reset()
	}

	lb.shift.Reset()
	lb.shift.anim.transType = TT_add // Since default is AIR transparency
	lb.shift.anim.srcAlpha = 0
	lb.shift.anim.dstAlpha = 255

	for i := range lb.red {
		lb.red[i].Reset()
	}

	lb.warn.Reset()
}

func (lb *LifeBar) bgDraw(layerno int16) {
	lb.bg0.Draw(float32(lb.pos[0])+sys.fightScreen.offsetX, float32(lb.pos[1])+sys.fightScreen.offsetY, layerno, sys.fightScreen.scale)
	lb.bg1.Draw(float32(lb.pos[0])+sys.fightScreen.offsetX, float32(lb.pos[1])+sys.fightScreen.offsetY, layerno, sys.fightScreen.scale)
	lb.bg2.Draw(float32(lb.pos[0])+sys.fightScreen.offsetX, float32(lb.pos[1])+sys.fightScreen.offsetY, layerno, sys.fightScreen.scale)
}

func (lb *LifeBar) draw(layerno int16, charpn int, lbr *LifeBar, f map[int]*Fnt) {
	// Get reference values
	refChar := sys.chars[charpn][0]
	life := float32(refChar.life) / float32(refChar.lifeMax)
	redlife := float32(refChar.redLife) / float32(refChar.lifeMax)
	redval := refChar.redLife - refChar.life

	var MidPosX = (float32(sys.gameWidth-320) / 2)
	var MidPosY = (float32(sys.gameHeight-240) / 2)
	// Calculates the clipping rectangle based on current bar settings
	getBarClipRect := func(life float32) [4]int32 {
		r := sys.scrrect

		if lb.scalefill {
			life = 1
		}

		if lb.range_x != [2]int32{0, 0} {
			r[0], r[2] = calcBarFillRect(lb.pos[0], lb.range_x, sys.fightScreen.offsetX, sys.fightScreen.scale, sys.widthScale, MidPosX, life)
		}

		if lb.range_y != [2]int32{0, 0} {
			r[1], r[3] = calcBarFillRect(lb.pos[1], lb.range_y, sys.fightScreen.offsetY, sys.fightScreen.scale, sys.heightScale, MidPosY, life)
		}
		return r
	}

	// This is already handled by step()
	// It makes the mid layer misbehave between rounds
	//if len(lb.mid.anim.frames) == 0 || life > lbr.midlife {
	//	life = lbr.midlife
	//}

	// Draw the three rectangles: top, mid, and red
	lr, mr, rr := getBarClipRect(lbr.toplife), getBarClipRect(lbr.midlife), getBarClipRect(redlife)

	var (
		lxs, mxs, rxs float32 = 1.0, 1.0, 1.0
		lys, mys, rys float32 = 1.0, 1.0, 1.0
	)
	if lb.scalefill { // Scale the sprite's size instead of adjusting the rectangle
		v := [3]float32{lbr.toplife, lbr.midlife, redlife}
		if lb.range_y != [2]int32{0, 0} {
			lys, mys, rys = v[0], v[1], v[2]
		} else {
			lxs, mxs, rxs = v[0], v[1], v[2]
		}
	} else {
		if lb.range_y != [2]int32{0, 0} {
			if lb.range_y[0] < lb.range_y[1] {
				mr[1] += lr[3]
				rr[1] += lr[3]
			}
			mr[3] -= Min(mr[3], lr[3])
			rr[3] -= Min(rr[3], lr[3])
		} else {
			if lb.range_x[0] < lb.range_x[1] {
				mr[0] += lr[2]
				rr[0] += lr[2]
			}
			mr[2] -= Min(mr[2], lr[2])
			rr[2] -= Min(rr[2], lr[2])
		}
	}

	// Draw red life
	if refChar.redLifeEnabled() {
		var rv int32
		for k := range lb.red {
			if k > rv && redval >= k {
				rv = k
			}
		}
		lb.red[rv].lay.DrawAnim(&rr, float32(lb.pos[0])+sys.fightScreen.offsetX, float32(lb.pos[1])+sys.fightScreen.offsetY, sys.fightScreen.scale, rxs, rys,
			layerno, lb.red[rv].anim, lb.red[rv].palfx)

		if lb.red_value[0].font[0] >= 0 && getFont(f, lb.red_value[0].font[0]) != nil {
			// Multiple red_value fonts according to redval
			var rv2 int32
			for k := range lb.red_value {
				if k > rv2 && redval >= k {
					rv2 = k
				}
			}
			text := strings.Replace(lb.red_value[rv2].text, "%d", fmt.Sprintf("%v", refChar.redLife), 1)
			text = strings.Replace(text, "%p", fmt.Sprintf("%v", math.Round(float64(redlife)*100)), 1)
			lb.red_value[rv2].lay.DrawText(
				float32(lb.pos[0])+sys.fightScreen.offsetX,
				float32(lb.pos[1])+sys.fightScreen.offsetY, sys.fightScreen.scale,
				layerno,
				text,
				getFont(f, lb.red_value[rv2].font[0]),
				lb.red_value[rv2].font[1],
				lb.red_value[rv2].font[2],
				lb.red_value[rv2].palfx,
				lb.red_value[rv2].frgba,
			)
		}
	}

	lb.mid.lay.DrawAnim(&mr, float32(lb.pos[0])+sys.fightScreen.offsetX, float32(lb.pos[1])+sys.fightScreen.offsetY, sys.fightScreen.scale, mxs, mys,
		layerno, lb.mid.anim, lb.mid.palfx)

	if lb.mid_shift {
		lb.shift.lay.DrawAnim(&mr, float32(lb.pos[0])+sys.fightScreen.offsetX, float32(lb.pos[1])+sys.fightScreen.offsetY, sys.fightScreen.scale, mxs, mys,
			layerno, lb.shift.anim, lb.shift.palfx)
	}

	// Multiple front elements
	var fv float32
	for k := range lb.front {
		if k > fv && life >= k/100 {
			fv = k
		}
	}
	lb.front[fv].lay.DrawAnim(&lr, float32(lb.pos[0])+sys.fightScreen.offsetX, float32(lb.pos[1])+sys.fightScreen.offsetY, sys.fightScreen.scale, lxs, lys,
		layerno, lb.front[fv].anim, lb.front[fv].palfx)

	lb.shift.lay.DrawAnim(&lr, float32(lb.pos[0])+sys.fightScreen.offsetX, float32(lb.pos[1])+sys.fightScreen.offsetY, sys.fightScreen.scale, lxs, lys,
		layerno, lb.shift.anim, lb.shift.palfx)

	if lb.value[0].font[0] >= 0 && getFont(f, lb.value[0].font[0]) != nil {
		// Multiple value fonts according to life value
		var fv2 int32
		for k := range lb.value {
			if k > fv2 && life >= float32(k)/100 {
				fv2 = k
			}
		}
		text := strings.Replace(lb.value[fv2].text, "%d", fmt.Sprintf("%v", refChar.life), 1)
		text = strings.Replace(text, "%p", fmt.Sprintf("%v", math.Round(float64(life)*100)), 1)
		lb.value[fv2].lay.DrawText(
			float32(lb.pos[0])+sys.fightScreen.offsetX,
			float32(lb.pos[1])+sys.fightScreen.offsetY, sys.fightScreen.scale,
			layerno,
			text,
			getFont(f, lb.value[fv2].font[0]),
			lb.value[fv2].font[1],
			lb.value[fv2].font[2],
			lb.value[fv2].palfx,
			lb.value[fv2].frgba,
		)
	}

	lb.top.Draw(float32(lb.pos[0])+sys.fightScreen.offsetX, float32(lb.pos[1])+sys.fightScreen.offsetY, layerno, sys.fightScreen.scale)

	if life <= float32(lb.warn_range[0])/100 && life >= float32(lb.warn_range[1])/100 {
		lb.warn.Draw(float32(lb.pos[0])+sys.fightScreen.offsetX, float32(lb.pos[1])+sys.fightScreen.offsetY, layerno, sys.fightScreen.scale)
	}
}

type PowerBar struct {
	pos              [2]int32
	range_x          [2]int32
	range_y          [2]int32
	bg0              map[int32]*AnimLayout
	bg1              AnimLayout
	bg2              AnimLayout
	top              AnimLayout
	mid              AnimLayout
	front            map[int32]*AnimLayout
	shift            AnimLayout
	counter          map[int32]*FSText
	counter_rounding int32
	value            map[int32]*FSText
	value_rounding   int32
	level_snd        [9][2]int32
	levelmax_snd     [2]int32
	midpower         float32
	midpowerMin      float32
	prevLevel        int32
	prevPower        int32
	levelbars        bool
	scalefill        bool
	leaderontop      bool
}

func newPowerBar() *PowerBar {
	newBar := &PowerBar{
		front:            make(map[int32]*AnimLayout),
		bg0:              make(map[int32]*AnimLayout),
		counter:          make(map[int32]*FSText),
		counter_rounding: 1000,
		value:            make(map[int32]*FSText),
		value_rounding:   1,
		levelmax_snd:     [2]int32{-1, -1},
	}
	// Default power level sounds to -1,-1
	for i := range newBar.level_snd {
		newBar.level_snd[i] = [2]int32{-1, -1}
	}
	return newBar
}

func readPowerBar(pre string, is IniSection, sff *Sff, at AnimationTable, f map[int]*Fnt) *PowerBar {
	pb := newPowerBar()

	is.ReadI32(pre+"pos", &pb.pos[0], &pb.pos[1])
	is.ReadI32(pre+"range.x", &pb.range_x[0], &pb.range_x[1])
	is.ReadI32(pre+"range.y", &pb.range_y[0], &pb.range_y[1])

	readPBKeys := func(name string) map[int32]string {
		result := make(map[int32]string)
		prefix := strings.ToLower(pre + name)
		for k := range is {
			kl := strings.ToLower(k)
			if !strings.HasPrefix(kl, prefix) {
				continue
			}
			rest := k[len(pre+name):]
			dot := strings.Index(rest, ".")
			if dot <= 0 {
				continue
			}
			keyStr := rest[:dot]
			if strings.EqualFold(keyStr, "max") {
				result[-1] = keyStr
			} else if IsNumeric(keyStr) {
				key := Atoi(keyStr)
				if _, ok := result[key]; !ok {
					result[key] = keyStr
				}
			}
		}
		return result
	}

	readMultipeAnimLayouts := func(name string) map[int32]*AnimLayout {
		result := make(map[int32]*AnimLayout)
		keys := readPBKeys(name)

		for k, keyStr := range keys {
			tmp := ReadAnimLayout(pre+name+keyStr+".", is, sff, at, 0)
			result[k] = &tmp
		}
		return result
	}

	readMultipleFSTexts := func(name, fmtstr string) map[int32]*FSText {
		result := make(map[int32]*FSText)
		keys := readPBKeys(name)

		for k, keyStr := range keys {
			result[k] = readFSText(pre+name+keyStr+".", is, fmtstr, 0, f, 0)
		}
		return result
	}

	pb.bg0 = readMultipeAnimLayouts("bg0")
	if _, ok := pb.bg0[0]; !ok {
		tmp := ReadAnimLayout(pre+"bg0.", is, sff, at, 0)
		pb.bg0[0] = &tmp
	}

	pb.bg1 = ReadAnimLayout(pre+"bg1.", is, sff, at, 0)
	pb.bg2 = ReadAnimLayout(pre+"bg2.", is, sff, at, 0)
	pb.mid = ReadAnimLayout(pre+"mid.", is, sff, at, 0)
	pb.top = ReadAnimLayout(pre+"top.", is, sff, at, 0)

	pb.front = readMultipeAnimLayouts("front")
	if _, ok := pb.front[0]; !ok {
		tmp := ReadAnimLayout(pre+"front.", is, sff, at, 0)
		pb.front[0] = &tmp
	}

	// Power counter
	pb.shift = ReadAnimLayout(pre+"shift.", is, sff, at, 0)
	pb.counter[0] = readFSText(pre+"counter.", is, "%i", 0, f, 0)
	for k, v := range readMultipleFSTexts("counter", "%i") {
		pb.counter[k] = v
	}
	pb.value[0] = readFSText(pre+"value.", is, "%d", 0, f, 0)
	for k, v := range readMultipleFSTexts("value", "%d") {
		pb.value[k] = v
	}

	// Format options.
	is.ReadI32(pre+"counter.format.rounding", &pb.counter_rounding)
	is.ReadI32(pre+"value.format.power.rounding", &pb.value_rounding)

	// Avoid division by 0.
	if pb.counter_rounding < 1 {
		pb.counter_rounding = 1000
	}
	if pb.value_rounding < 1 {
		pb.value_rounding = 1
	}

	// Level sounds.
	for i := range pb.level_snd {
		if !is.ReadI32(fmt.Sprintf("%vlevel%v.snd", pre, i+1), &pb.level_snd[i][0], &pb.level_snd[i][1]) {
			is.ReadI32(fmt.Sprintf("level%v.snd", i+1), &pb.level_snd[i][0], &pb.level_snd[i][1])
		}
	}

	// Level max sound.
	if !is.ReadI32(pre+"levelmax.snd", &pb.levelmax_snd[0], &pb.levelmax_snd[1]) {
		is.ReadI32("levelmax.snd", &pb.levelmax_snd[0], &pb.levelmax_snd[1])
	}

	is.ReadBool(pre+"levelbars", &pb.levelbars)
	is.ReadBool(pre+"scalefill", &pb.scalefill)
	is.ReadBool("leaderontop", &pb.leaderontop)

	return pb
}

func resolvePBKey[T any](m map[int32]T, pbval int32, max int32) int32 {
	var choice int32
	hasMax := false
	for k := range m {
		if k == -1 {
			hasMax = true
			continue
		}
		if k > choice && pbval >= k {
			choice = k
		}
	}
	if hasMax && pbval >= max {
		return -1
	}
	return choice
}

func (pb *PowerBar) step(charpn int, pbr *PowerBar, snd *Snd) {
	// Get reference values
	refChar := sys.chars[charpn][0]
	pbval := refChar.getPower()
	power := float32(pbval) / float32(refChar.powerMax)
	level := pbval / 1000

	if pb.levelbars {
		power = float32(pbval)/1000 - Min(float32(level), float32(refChar.powerMax)/1000-1)
	}

	// Element shifting gradient
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

	// Initialize "previous power" in the first frame of the round
	// This prevents level sounds and similar actions from happening in that frame
	if sys.matchTime == 0 {
		pbr.prevLevel = level
		pbr.prevPower = pbval
	}

	// Level sounds
	// Skipped if the bar is invisible
	if !sys.gsf(GSF_nobardisplay) && !refChar.powerOwner().asf(ASF_nopowerbardisplay) {
		if pbval >= refChar.powerMax && pbr.prevPower < refChar.powerMax && pb.levelmax_snd[0] != -1 {
			snd.play(pb.levelmax_snd, 100, 0, 0, 0, 0)
		} else if level > pbr.prevLevel {
			i := int(level - 1)
			if i >= 0 && i < len(pb.level_snd) {
				snd.play(pb.level_snd[i], 100, 0, 0, 0, 0)
			}
		}
	}

	// Reset PalFX if level number changes at all
	// TODO: Maybe this should happen if the "multiple element" changes instead
	if level != pbr.prevLevel {
		for i := range pb.counter {
			pb.counter[i].resetTxtPfx()
		}
		for i := range pb.value {
			pb.value[i].resetTxtPfx()
		}
	}

	// Save current level for reference in the next frame
	pbr.prevLevel = level
	pbr.prevPower = pbval

	// Multiple front elements
	fv1 := resolvePBKey(pb.bg0, pbval, refChar.powerMax)
	pb.bg0[fv1].Action()

	pb.bg1.Action()
	pb.bg2.Action()
	pb.top.Action()
	pb.mid.Action()

	// Multiple front elements
	fv2 := resolvePBKey(pb.front, pbval, refChar.powerMax)
	pb.front[fv2].Action()

	pb.shift.Action()

	// Multiple counter fonts
	cv := resolvePBKey(pb.counter, pbval, refChar.powerMax)
	pb.counter[cv].step()

	// Multiple value fonts
	cv2 := resolvePBKey(pb.value, pbval, refChar.powerMax)
	pb.value[cv2].step()
}

func (pb *PowerBar) reset() {
	for i := range pb.bg0 {
		pb.bg0[i].Reset()
	}
	pb.bg1.Reset()
	pb.bg2.Reset()
	pb.top.Reset()
	pb.mid.Reset()
	for i := range pb.front {
		pb.front[i].Reset()
	}

	pb.shift.Reset()
	pb.shift.anim.transType = TT_add
	pb.shift.anim.srcAlpha = 0
	pb.shift.anim.dstAlpha = 255
}

func (pb *PowerBar) bgDraw(layerno int16, charpn int) {
	pbval := sys.chars[charpn][0].getPower()
	refChar := sys.chars[charpn][0]
	fv := resolvePBKey(pb.bg0, pbval, refChar.powerMax)
	pb.bg0[fv].Draw(float32(pb.pos[0])+sys.fightScreen.offsetX, float32(pb.pos[1])+sys.fightScreen.offsetY, layerno, sys.fightScreen.scale)
	pb.bg1.Draw(float32(pb.pos[0])+sys.fightScreen.offsetX, float32(pb.pos[1])+sys.fightScreen.offsetY, layerno, sys.fightScreen.scale)
	pb.bg2.Draw(float32(pb.pos[0])+sys.fightScreen.offsetX, float32(pb.pos[1])+sys.fightScreen.offsetY, layerno, sys.fightScreen.scale)
}

func (pb *PowerBar) draw(layerno int16, charpn int, pbr *PowerBar, f map[int]*Fnt) {
	refChar := sys.chars[charpn][0]
	pbval := refChar.getPower()
	power := float32(pbval) / float32(refChar.powerMax)
	level := pbval / 1000

	if pb.levelbars {
		power = float32(pbval)/1000 - Min(float32(level), float32(refChar.powerMax)/1000-1)
	}

	var MidPosX = (float32(sys.gameWidth-320) / 2)
	var MidPosY = (float32(sys.gameHeight-240) / 2)
	getBarClipRect := func(power float32) [4]int32 {
		r := sys.scrrect

		if pb.scalefill {
			power = 1
		}

		if pb.range_x != [2]int32{0, 0} {
			r[0], r[2] = calcBarFillRect(pb.pos[0], pb.range_x, sys.fightScreen.offsetX, sys.fightScreen.scale, sys.widthScale, MidPosX, power)
		}

		if pb.range_y != [2]int32{0, 0} {
			r[1], r[3] = calcBarFillRect(pb.pos[1], pb.range_y, sys.fightScreen.offsetY, sys.fightScreen.scale, sys.heightScale, MidPosY, power)
		}
		return r
	}
	pr, mr := getBarClipRect(power), getBarClipRect(pbr.midpower)

	var (
		pxs, mxs float32 = 1.0, 1.0
		pys, mys float32 = 1.0, 1.0
	)
	if pb.scalefill {
		v := [3]float32{power, pbr.midpower}
		if pb.range_y != [2]int32{0, 0} {
			pys, mys = v[0], v[1]
		} else {
			pxs, mxs = v[0], v[1]
		}
	} else {
		if pb.range_y != [2]int32{0, 0} {
			if pb.range_y[0] < pb.range_y[1] {
				mr[1] += pr[3]
			}
			mr[3] -= Min(mr[3], pr[3])
		} else {
			if pb.range_x[0] < pb.range_x[1] {
				mr[0] += pr[2]
			}
			mr[2] -= Min(mr[2], pr[2])
		}
	}
	pb.mid.lay.DrawAnim(&mr, float32(pb.pos[0])+sys.fightScreen.offsetX, float32(pb.pos[1])+sys.fightScreen.offsetY, sys.fightScreen.scale, mxs, mys,
		layerno, pb.mid.anim, pb.mid.palfx)

	// Multiple front elements
	fv := resolvePBKey(pb.front, pbval, refChar.powerMax)
	pb.front[fv].lay.DrawAnim(&pr, float32(pb.pos[0])+sys.fightScreen.offsetX, float32(pb.pos[1])+sys.fightScreen.offsetY, sys.fightScreen.scale, pxs, pys,
		layerno, pb.front[fv].anim, pb.front[fv].palfx)

	pb.shift.lay.DrawAnim(&pr, float32(pb.pos[0])+sys.fightScreen.offsetX, float32(pb.pos[1])+sys.fightScreen.offsetY, sys.fightScreen.scale, pxs, pys,
		layerno, pb.shift.anim, pb.shift.palfx)

	// Powerbar text.
	if pb.counter[0].font[0] >= 0 && getFont(f, pb.counter[0].font[0]) != nil {
		// Multiple counter fonts according to powerbar level
		cv := resolvePBKey(pb.counter, pbval, refChar.powerMax)

		pb.counter[cv].lay.DrawText(
			float32(pb.pos[0])+sys.fightScreen.offsetX,
			float32(pb.pos[1])+sys.fightScreen.offsetY,
			sys.fightScreen.scale,
			layerno,
			strings.Replace(pb.counter[cv].text, "%i", fmt.Sprintf("%v", pbval/pb.counter_rounding), 1),
			getFont(f, pb.counter[cv].font[0]),
			pb.counter[cv].font[1],
			pb.counter[cv].font[2],
			pb.counter[cv].palfx,
			pb.counter[cv].frgba,
		)
	}

	// Per-level powerbar text.
	if pb.value[0].font[0] >= 0 && getFont(f, pb.value[0].font[0]) != nil {
		// Multiple value fonts according to powerbar level
		cv2 := resolvePBKey(pb.counter, pbval, refChar.powerMax)
		text := strings.Replace(pb.value[cv2].text, "%d", fmt.Sprintf("%v", pbval/pb.value_rounding), 1)
		text = strings.Replace(text, "%p", fmt.Sprintf("%v", math.Round(float64(power)*100)), 1)

		pb.value[cv2].lay.DrawText(
			float32(pb.pos[0])+sys.fightScreen.offsetX,
			float32(pb.pos[1])+sys.fightScreen.offsetY,
			sys.fightScreen.scale,
			layerno,
			text,
			getFont(f, pb.value[cv2].font[0]),
			pb.value[cv2].font[1],
			pb.value[cv2].font[2],
			pb.value[cv2].palfx,
			pb.value[cv2].frgba,
		)
	}
	pb.top.Draw(float32(pb.pos[0])+sys.fightScreen.offsetX, float32(pb.pos[1])+sys.fightScreen.offsetY, layerno, sys.fightScreen.scale)
}

type GuardBar struct {
	pos         [2]int32
	range_x     [2]int32
	range_y     [2]int32
	bg0         AnimLayout
	bg1         AnimLayout
	bg2         AnimLayout
	top         AnimLayout
	mid         AnimLayout
	warn        AnimLayout
	warn_range  [2]int32
	value       map[int32]*FSText
	front       map[float32]*AnimLayout
	shift       AnimLayout
	midpower    float32
	midpowerMin float32
	invertfill  bool
	scalefill   bool
	leaderontop bool
}

func newGuardBar() (gb *GuardBar) {
	gb = &GuardBar{
		front: make(map[float32]*AnimLayout),
		value: make(map[int32]*FSText),
	}
	return
}

func readGuardBar(pre string, is IniSection,
	sff *Sff, at AnimationTable, f map[int]*Fnt) *GuardBar {
	gb := newGuardBar()

	is.ReadI32(pre+"pos", &gb.pos[0], &gb.pos[1])
	is.ReadI32(pre+"range.x", &gb.range_x[0], &gb.range_x[1])
	is.ReadI32(pre+"range.y", &gb.range_y[0], &gb.range_y[1])

	gb.bg0 = ReadAnimLayout(pre+"bg0.", is, sff, at, 0)
	gb.bg1 = ReadAnimLayout(pre+"bg1.", is, sff, at, 0)
	gb.bg2 = ReadAnimLayout(pre+"bg2.", is, sff, at, 0)
	gb.mid = ReadAnimLayout(pre+"mid.", is, sff, at, 0)
	gb.top = ReadAnimLayout(pre+"top.", is, sff, at, 0)

	gb.front = readMultipleValuesF(pre, "front", is, sff, at)
	if _, ok := gb.front[0]; !ok {
		tmp := ReadAnimLayout(pre+"front.", is, sff, at, 0)
		gb.front[0] = &tmp
	}

	gb.shift = ReadAnimLayout(pre+"shift.", is, sff, at, 0)
	gb.value[0] = readFSText(pre+"value.", is, "%d", 0, f, 0)
	for k, v := range readMultipleFSText(pre, "value", is, "%d", 0, f, 0) {
		gb.value[k] = v
	}
	is.ReadI32(pre+"warn.range", &gb.warn_range[0], &gb.warn_range[1])
	gb.warn = ReadAnimLayout(pre+"warn.", is, sff, at, 0)
	is.ReadBool(pre+"invertfill", &gb.invertfill)
	is.ReadBool(pre+"scalefill", &gb.scalefill)
	is.ReadBool("leaderontop", &gb.leaderontop)
	return gb
}

func (gb *GuardBar) step(charpn int, gbr *GuardBar, snd *Snd) {
	refChar := sys.chars[charpn][0]
	if !refChar.guardBreakEnabled() {
		return
	}

	points := float32(refChar.guardPoints) / float32(refChar.guardPointsMax)
	if gb.invertfill {
		points = 1 - points
	}

	// Element shifting gradient
	gb.shift.anim.srcAlpha = int16(255 * (1 - points))
	gb.shift.anim.dstAlpha = int16(255 * points)

	gbr.midpower -= 1.0 / 144
	if points < gbr.midpowerMin {
		gbr.midpowerMin += (points - gbr.midpowerMin) * (1 / (12 - (points-gbr.midpowerMin)*144))
	} else {
		gbr.midpowerMin = points
	}
	if gbr.midpower < gbr.midpowerMin {
		gbr.midpower = gbr.midpowerMin
	}
	gb.bg0.Action()
	gb.bg1.Action()
	gb.bg2.Action()
	gb.top.Action()
	gb.mid.Action()

	// Multiple front elements
	var mv float32
	for k := range gb.front {
		if k > mv && points >= k/100 {
			mv = k
		}
	}
	gb.front[mv].Action()

	// Multiple value fonts
	var mv2 int32
	for k := range gb.value {
		if k > mv2 && points >= float32(k)/100 {
			mv2 = k
		}
	}
	gb.value[mv2].step()

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
	gb.shift.anim.transType = TT_add
	gb.shift.anim.srcAlpha = 0
	gb.shift.anim.dstAlpha = 255

	gb.warn.Reset()
}

func (gb *GuardBar) bgDraw(layerno int16) {
	// Handled in outer loop
	//if !sys.fightScreen.guardbar {
	//	return
	//}

	gb.bg0.Draw(float32(gb.pos[0])+sys.fightScreen.offsetX, float32(gb.pos[1])+sys.fightScreen.offsetY, layerno, sys.fightScreen.scale)
	gb.bg1.Draw(float32(gb.pos[0])+sys.fightScreen.offsetX, float32(gb.pos[1])+sys.fightScreen.offsetY, layerno, sys.fightScreen.scale)
	gb.bg2.Draw(float32(gb.pos[0])+sys.fightScreen.offsetX, float32(gb.pos[1])+sys.fightScreen.offsetY, layerno, sys.fightScreen.scale)
}

func (gb *GuardBar) draw(layerno int16, charpn int, gbr *GuardBar, f map[int]*Fnt) {
	// Handled in outer loop
	//if !sys.fightScreen.guardbar {
	//	return
	//}

	refChar := sys.chars[charpn][0]
	points := float32(refChar.guardPoints) / float32(refChar.guardPointsMax)
	if gb.invertfill {
		points = 1 - points
	}

	var MidPosX = (float32(sys.gameWidth-320) / 2)
	var MidPosY = (float32(sys.gameHeight-240) / 2)
	getBarClipRect := func(points float32) [4]int32 {
		r := sys.scrrect

		if gb.scalefill {
			points = 1
		}

		if gb.range_x != [2]int32{0, 0} {
			r[0], r[2] = calcBarFillRect(gb.pos[0], gb.range_x, sys.fightScreen.offsetX, sys.fightScreen.scale, sys.widthScale, MidPosX, points)
		}

		if gb.range_y != [2]int32{0, 0} {
			r[1], r[3] = calcBarFillRect(gb.pos[1], gb.range_y, sys.fightScreen.offsetY, sys.fightScreen.scale, sys.heightScale, MidPosY, points)
		}
		return r
	}

	pr, mr := getBarClipRect(points), getBarClipRect(gbr.midpower)

	var (
		pxs, mxs float32 = 1.0, 1.0
		pys, mys float32 = 1.0, 1.0
	)
	if gb.scalefill {
		v := [3]float32{points, gbr.midpower}
		if gb.range_y != [2]int32{0, 0} {
			pys, mys = v[0], v[1]
		} else {
			pxs, mxs = v[0], v[1]
		}
	} else {
		if gb.range_y != [2]int32{0, 0} {
			if gb.range_y[0] < gb.range_y[1] {
				mr[1] += pr[3]
			}
			mr[3] -= Min(mr[3], pr[3])
		} else {
			if gb.range_x[0] < gb.range_x[1] {
				mr[0] += pr[2]
			}
			mr[2] -= Min(mr[2], pr[2])
		}
	}
	gb.mid.lay.DrawAnim(&mr, float32(gb.pos[0])+sys.fightScreen.offsetX, float32(gb.pos[1])+sys.fightScreen.offsetY, sys.fightScreen.scale, mxs, mys,
		layerno, gb.mid.anim, gb.mid.palfx)

	// Multiple front elements
	var mv float32
	for k := range gb.front {
		if k > mv && points >= k/100 {
			mv = k
		}
	}
	gb.front[mv].lay.DrawAnim(&pr, float32(gb.pos[0])+sys.fightScreen.offsetX, float32(gb.pos[1])+sys.fightScreen.offsetY, sys.fightScreen.scale, pxs, pys,
		layerno, gb.front[mv].anim, gb.front[mv].palfx)

	gb.shift.lay.DrawAnim(&pr, float32(gb.pos[0])+sys.fightScreen.offsetX, float32(gb.pos[1])+sys.fightScreen.offsetY, sys.fightScreen.scale, pxs, pys,
		layerno, gb.shift.anim, gb.shift.palfx)

	if gb.value[0].font[0] >= 0 && getFont(f, gb.value[0].font[0]) != nil {
		// Multiple value fonts according to guardbar points
		var mv2 int32
		for k := range gb.value {
			if k > mv2 && points >= float32(k)/100 {
				mv2 = k
			}
		}
		text := strings.Replace(gb.value[mv2].text, "%d", fmt.Sprintf("%v", refChar.guardPoints), 1)
		text = strings.Replace(text, "%p", fmt.Sprintf("%v", math.Round(float64(points)*100)), 1)
		gb.value[mv2].lay.DrawText(
			float32(gb.pos[0])+sys.fightScreen.offsetX,
			float32(gb.pos[1])+sys.fightScreen.offsetY,
			sys.fightScreen.scale,
			layerno,
			text,
			getFont(f, gb.value[mv2].font[0]),
			gb.value[mv2].font[1],
			gb.value[mv2].font[2],
			gb.value[mv2].palfx,
			gb.value[mv2].frgba,
		)
	}

	if points <= float32(gb.warn_range[0])/100 && points >= float32(gb.warn_range[1])/100 {
		gb.warn.Draw(float32(gb.pos[0])+sys.fightScreen.offsetX, float32(gb.pos[1])+sys.fightScreen.offsetY, layerno, sys.fightScreen.scale)
	}

	gb.top.Draw(float32(gb.pos[0])+sys.fightScreen.offsetX, float32(gb.pos[1])+sys.fightScreen.offsetY, layerno, sys.fightScreen.scale)
}

type StunBar struct {
	pos         [2]int32
	range_x     [2]int32
	range_y     [2]int32
	bg0         AnimLayout
	bg1         AnimLayout
	bg2         AnimLayout
	top         AnimLayout
	mid         AnimLayout
	warn_range  [2]int32
	warn        AnimLayout
	value       map[int32]*FSText
	front       map[float32]*AnimLayout
	shift       AnimLayout
	midpower    float32
	midpowerMin float32
	invertfill  bool
	scalefill   bool
	leaderontop bool
}

func newStunBar() (sb *StunBar) {
	sb = &StunBar{
		front: make(map[float32]*AnimLayout),
		value: make(map[int32]*FSText),
	}
	return
}

func readStunBar(pre string, is IniSection,
	sff *Sff, at AnimationTable, f map[int]*Fnt) *StunBar {
	sb := newStunBar()

	is.ReadI32(pre+"pos", &sb.pos[0], &sb.pos[1])
	is.ReadI32(pre+"range.x", &sb.range_x[0], &sb.range_x[1])
	is.ReadI32(pre+"range.y", &sb.range_y[0], &sb.range_y[1])

	sb.bg0 = ReadAnimLayout(pre+"bg0.", is, sff, at, 0)
	sb.bg1 = ReadAnimLayout(pre+"bg1.", is, sff, at, 0)
	sb.bg2 = ReadAnimLayout(pre+"bg2.", is, sff, at, 0)
	sb.mid = ReadAnimLayout(pre+"mid.", is, sff, at, 0)
	sb.top = ReadAnimLayout(pre+"top.", is, sff, at, 0)

	sb.front = readMultipleValuesF(pre, "front", is, sff, at)
	if _, ok := sb.front[0]; !ok {
		tmp := ReadAnimLayout(pre+"front.", is, sff, at, 0)
		sb.front[0] = &tmp
	}

	sb.shift = ReadAnimLayout(pre+"shift.", is, sff, at, 0)
	sb.value[0] = readFSText(pre+"value.", is, "%d", 0, f, 0)
	for k, v := range readMultipleFSText(pre, "value", is, "%d", 0, f, 0) {
		sb.value[k] = v
	}
	is.ReadI32(pre+"warn.range", &sb.warn_range[0], &sb.warn_range[1])
	sb.warn = ReadAnimLayout(pre+"warn.", is, sff, at, 0)
	is.ReadBool(pre+"invertfill", &sb.invertfill)
	is.ReadBool(pre+"scalefill", &sb.scalefill)
	is.ReadBool("leaderontop", &sb.leaderontop)
	return sb
}

func (sb *StunBar) step(charpn int, sbr *StunBar, snd *Snd) {
	refChar := sys.chars[charpn][0]
	if !refChar.dizzyEnabled() {
		return
	}

	points := float32(refChar.dizzyPoints) / float32(refChar.dizzyPointsMax)
	if sb.invertfill {
		points = 1 - points
	}

	// Element shifting gradient
	sb.shift.anim.srcAlpha = int16(255 * (1 - points))
	sb.shift.anim.dstAlpha = int16(255 * points)

	sbr.midpower -= 1.0 / 144
	if points < sbr.midpowerMin {
		sbr.midpowerMin += (points - sbr.midpowerMin) * (1 / (12 - (points-sbr.midpowerMin)*144))
	} else {
		sbr.midpowerMin = points
	}
	if sbr.midpower < sbr.midpowerMin {
		sbr.midpower = sbr.midpowerMin
	}
	sb.bg0.Action()
	sb.bg1.Action()
	sb.bg2.Action()
	sb.top.Action()
	sb.mid.Action()
	// Multiple front elements
	var mv float32
	for k := range sb.front {
		if k > mv && points >= k/100 {
			mv = k
		}
	}
	sb.front[mv].Action()
	// Multiple value fonts
	var mv2 int32
	for k := range sb.value {
		if k > mv2 && points >= float32(k)/100 {
			mv2 = k
		}
	}
	sb.value[mv2].step()

	sb.shift.Action()
	sb.warn.Action()
}

func (sb *StunBar) reset() {
	sb.bg0.Reset()
	sb.bg1.Reset()
	sb.bg2.Reset()
	sb.top.Reset()
	sb.mid.Reset()
	for i := range sb.front {
		sb.front[i].Reset()
	}

	sb.shift.Reset()
	sb.shift.anim.transType = TT_add
	sb.shift.anim.srcAlpha = 255
	sb.shift.anim.dstAlpha = 0

	sb.warn.Reset()
}

func (sb *StunBar) bgDraw(layerno int16) {
	// Handled in outer loop
	//if !sys.fightScreen.stunbar {
	//	return
	//}

	sb.bg0.Draw(float32(sb.pos[0])+sys.fightScreen.offsetX, float32(sb.pos[1])+sys.fightScreen.offsetY, layerno, sys.fightScreen.scale)
	sb.bg1.Draw(float32(sb.pos[0])+sys.fightScreen.offsetX, float32(sb.pos[1])+sys.fightScreen.offsetY, layerno, sys.fightScreen.scale)
	sb.bg2.Draw(float32(sb.pos[0])+sys.fightScreen.offsetX, float32(sb.pos[1])+sys.fightScreen.offsetY, layerno, sys.fightScreen.scale)
}

func (sb *StunBar) draw(layerno int16, charpn int, sbr *StunBar, f map[int]*Fnt) {
	// Handled in outer loop
	//if !sys.fightScreen.stunbar {
	//	return
	//}

	refChar := sys.chars[charpn][0]
	points := float32(refChar.dizzyPoints) / float32(refChar.dizzyPointsMax)
	if sb.invertfill {
		points = 1 - points
	}

	var MidPosX = (float32(sys.gameWidth-320) / 2)
	var MidPosY = (float32(sys.gameHeight-240) / 2)
	getBarClipRect := func(points float32) [4]int32 {
		r := sys.scrrect

		if sb.scalefill {
			points = 1
		}

		if sb.range_x != [2]int32{0, 0} {
			r[0], r[2] = calcBarFillRect(sb.pos[0], sb.range_x, sys.fightScreen.offsetX, sys.fightScreen.scale, sys.widthScale, MidPosX, points)
		}

		if sb.range_y != [2]int32{0, 0} {
			r[1], r[3] = calcBarFillRect(sb.pos[1], sb.range_y, sys.fightScreen.offsetY, sys.fightScreen.scale, sys.heightScale, MidPosY, points)
		}
		return r
	}

	pr, mr := getBarClipRect(points), getBarClipRect(sbr.midpower)

	var (
		pxs, mxs float32 = 1.0, 1.0
		pys, mys float32 = 1.0, 1.0
	)
	if sb.scalefill {
		v := [3]float32{points, sbr.midpower}
		if sb.range_y != [2]int32{0, 0} {
			pys, mys = v[0], v[1]
		} else {
			pxs, mxs = v[0], v[1]
		}
	} else {
		if sb.range_y != [2]int32{0, 0} {
			if sb.range_y[0] < sb.range_y[1] {
				mr[1] += pr[3]
			}
			mr[3] -= Min(mr[3], pr[3])
		} else {
			if sb.range_x[0] < sb.range_x[1] {
				mr[0] += pr[2]
			}
			mr[2] -= Min(mr[2], pr[2])
		}
	}
	sb.mid.lay.DrawAnim(&mr, float32(sb.pos[0])+sys.fightScreen.offsetX, float32(sb.pos[1])+sys.fightScreen.offsetY, sys.fightScreen.scale, mxs, mys,
		layerno, sb.mid.anim, sb.mid.palfx)

	// Multiple front elements
	var mv float32
	for k := range sb.front {
		if k > mv && points >= k/100 {
			mv = k
		}
	}
	sb.front[mv].lay.DrawAnim(&pr, float32(sb.pos[0])+sys.fightScreen.offsetX, float32(sb.pos[1])+sys.fightScreen.offsetY, sys.fightScreen.scale, pxs, pys,
		layerno, sb.front[mv].anim, sb.front[mv].palfx)

	sb.shift.lay.DrawAnim(&pr, float32(sb.pos[0])+sys.fightScreen.offsetX, float32(sb.pos[1])+sys.fightScreen.offsetY, sys.fightScreen.scale, pxs, pys,
		layerno, sb.shift.anim, sb.shift.palfx)

	if sb.value[0].font[0] >= 0 && getFont(f, sb.value[0].font[0]) != nil {
		// Multiple value fonts according to stunbar points
		var mv2 int32
		for k := range sb.value {
			if k > mv2 && points >= float32(k)/100 {
				mv2 = k
			}
		}
		text := strings.Replace(sb.value[mv2].text, "%d", fmt.Sprintf("%v", refChar.dizzyPoints), 1)
		text = strings.Replace(text, "%p", fmt.Sprintf("%v", math.Round(float64(points)*100)), 1)
		sb.value[mv2].lay.DrawText(
			float32(sb.pos[0])+sys.fightScreen.offsetX,
			float32(sb.pos[1])+sys.fightScreen.offsetY,
			sys.fightScreen.scale,
			layerno,
			text,
			getFont(f, sb.value[mv2].font[0]),
			sb.value[mv2].font[1],
			sb.value[mv2].font[2],
			sb.value[mv2].palfx,
			sb.value[mv2].frgba,
		)
	}

	if points >= float32(sb.warn_range[0])/100 && points <= float32(sb.warn_range[1])/100 {
		sb.warn.Draw(float32(sb.pos[0])+sys.fightScreen.offsetX, float32(sb.pos[1])+sys.fightScreen.offsetY, layerno, sys.fightScreen.scale)
	}

	sb.top.Draw(float32(sb.pos[0])+sys.fightScreen.offsetX, float32(sb.pos[1])+sys.fightScreen.offsetY, layerno, sys.fightScreen.scale)
}

type FightScreenFace struct {
	pos                    [2]int32
	bg                     AnimLayout
	bg0                    AnimLayout
	bg1                    AnimLayout
	bg2                    AnimLayout
	top                    AnimLayout
	ko                     AnimLayout
	face_spr               [2]int32
	face                   *Sprite
	face_lay               Layout
	face_palshare          bool
	face_palfxshare        bool
	face_darkenshare       bool
	teammate_pos           [2]int32
	teammate_spacing       [2]int32
	teammate_bg            AnimLayout
	teammate_bg0           AnimLayout
	teammate_bg1           AnimLayout
	teammate_bg2           AnimLayout
	teammate_top           AnimLayout
	teammate_ko            AnimLayout
	teammate_face_spr      [2]int32
	teammate_face          []*Sprite
	teammate_face_lay      Layout
	teammate_scale         []float32
	teammate_ko_hide       bool
	teammate_face_palshare bool
	numko                  int32
	old_spr                [2]int32
	old_pal                [2]int32
	face_pfx               *PalFX
	teammate_face_pfx      []*PalFX
	leaderontop            bool
}

func newFightScreenFace() *FightScreenFace {
	return &FightScreenFace{
		face_spr:               [2]int32{-1},
		teammate_face_spr:      [2]int32{-1},
		face_palshare:          true,
		face_pfx:               newPalFX(),
		teammate_face_palshare: true,
		teammate_face_pfx:      nil, // Allocated later
	}
	// Before Mugen 1.1, portraits used to share the character's palette in every way
	// The equivalent of enabling face_palshare, face_palfxshare and face_darkenshare
	// Mugen 1.1 has some bugs that broke all off that
	// However, we will default face_palfxshare to disabled because it's no longer a common fighting game feature
	// face_darkenshare has the same default as face_palfxshare because the two are related
}

func readFightScreenFace(pre string, is IniSection, sff *Sff, at AnimationTable) *FightScreenFace {
	fa := newFightScreenFace()

	is.ReadI32(pre+"pos", &fa.pos[0], &fa.pos[1])
	fa.bg = ReadAnimLayout(pre+"bg.", is, sff, at, 0)
	fa.bg0 = ReadAnimLayout(pre+"bg0.", is, sff, at, 0)
	fa.bg1 = ReadAnimLayout(pre+"bg1.", is, sff, at, 0)
	fa.bg2 = ReadAnimLayout(pre+"bg2.", is, sff, at, 0)
	fa.top = ReadAnimLayout(pre+"top.", is, sff, at, 0)
	fa.ko = ReadAnimLayout(pre+"ko.", is, sff, at, 0)

	is.ReadI32(pre+"face.spr", &fa.face_spr[0], &fa.face_spr[1])
	fa.face_lay = *ReadLayout(pre+"face.", is, 0)
	is.ReadBool(pre+"face.palshare", &fa.face_palshare)
	is.ReadBool(pre+"face.palfxshare", &fa.face_palfxshare)
	is.ReadBool(pre+"face.darkenshare", &fa.face_darkenshare)
	is.ReadBool("leaderontop", &fa.leaderontop)

	// Teammates
	is.ReadI32(pre+"teammate.pos", &fa.teammate_pos[0], &fa.teammate_pos[1])
	is.ReadI32(pre+"teammate.spacing", &fa.teammate_spacing[0],
		&fa.teammate_spacing[1])
	fa.teammate_bg = ReadAnimLayout(pre+"teammate.bg.", is, sff, at, 0)
	fa.teammate_bg0 = ReadAnimLayout(pre+"teammate.bg0.", is, sff, at, 0)
	fa.teammate_bg1 = ReadAnimLayout(pre+"teammate.bg1.", is, sff, at, 0)
	fa.teammate_bg2 = ReadAnimLayout(pre+"teammate.bg2.", is, sff, at, 0)
	fa.teammate_top = ReadAnimLayout(pre+"teammate.top.", is, sff, at, 0)
	fa.teammate_ko = ReadAnimLayout(pre+"teammate.ko.", is, sff, at, 0)
	is.ReadI32(pre+"teammate.face.spr", &fa.teammate_face_spr[0],
		&fa.teammate_face_spr[1])
	if fa.teammate_face_spr[0] != -1 {
		sys.sel.charSpritePreload[[...]uint16{uint16(fa.teammate_face_spr[0]), uint16(fa.teammate_face_spr[1])}] = true
	}
	fa.teammate_face_lay = *ReadLayout(pre+"teammate.face.", is, 0)
	is.ReadBool(pre+"teammate.ko.hide", &fa.teammate_ko_hide)
	is.ReadBool(pre+"teammate.face.palshare", &fa.teammate_face_palshare)

	return fa
}

func (fa *FightScreenFace) step(charpn int, refFace *FightScreenFace) {
	refChar := sys.chars[charpn][0]
	group, number := fa.face_spr[0], fa.face_spr[1]
	if refChar != nil && refChar.anim != nil {
		if mg, ok := refChar.anim.remap[group]; ok {
			if mn, ok := mg[number]; ok {
				group, number = mn[0], mn[1]
			}
		}
	}

	// Update sprite only when necessary
	refgi := sys.cgi[charpn]
	if refFace.old_spr[0] != group || refFace.old_spr[1] != number ||
		refFace.old_pal[0] != refgi.remappedpal[0] || refFace.old_pal[1] != refgi.remappedpal[1] {
		refFace.face = refgi.sff.cloneSpriteWithPal(uint16(group), uint16(number), &refgi.palettedata.palList)
		refFace.old_spr = [2]int32{group, number}
		refFace.old_pal = [2]int32{refgi.remappedpal[0], refgi.remappedpal[1]}
	}

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

func (fa *FightScreenFace) reset() {
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

func (fa *FightScreenFace) bgDraw(layerno int16) {
	fa.bg.Draw(float32(fa.pos[0])+sys.fightScreen.offsetX, float32(fa.pos[1]), layerno, sys.fightScreen.scale)
	fa.bg0.Draw(float32(fa.pos[0])+sys.fightScreen.offsetX, float32(fa.pos[1]), layerno, sys.fightScreen.scale)
	fa.bg1.Draw(float32(fa.pos[0])+sys.fightScreen.offsetX, float32(fa.pos[1]), layerno, sys.fightScreen.scale)
	fa.bg2.Draw(float32(fa.pos[0])+sys.fightScreen.offsetX, float32(fa.pos[1]), layerno, sys.fightScreen.scale)
}

func (fa *FightScreenFace) draw(layerno int16, charpn int, refFace *FightScreenFace) {
	refChar := sys.chars[charpn][0]

	if refFace.face != nil {
		// Get player current PalFX if applicable
		// These flags should check "fa" instead of "refFace"
		// https://github.com/ikemen-engine/Ikemen-GO/issues/2269
		if fa.face_palfxshare {
			*fa.face_pfx = *refChar.getPalfx()
		}

		// Update portrait palette
		if refFace.face.coldepth <= 8 {
			// Check the player's current palette
			palIdx := refFace.face.palidx
			if fa.face_palshare {
				remap := refChar.getPalfx().remap
				if int(palIdx) < len(remap) {
					palIdx = remap[palIdx]
				}
			}
			// Retrieve that palette
			palList := &sys.cgi[charpn].palettedata.palList
			charPal := palList.Get(palIdx)
			// Update portrait palette and cache
			refFace.face.Pal = charPal
			refFace.face.PalTex = refFace.face.CachePalTex(charPal)
		}

		// Save current brightness
		oldBright := sys.brightness

		// Reset system brightness if player initiated SuperPause (cancel "darken" parameter)
		if fa.face_darkenshare && refChar.ignoreDarkenTime > 0 {
			sys.brightness = 1.0
		}

		// Draw the actual face sprite
		fa.face_lay.DrawFaceSprite((float32(fa.pos[0])+sys.fightScreen.offsetX)*sys.fightScreen.scale, float32(fa.pos[1])*sys.fightScreen.scale, layerno,
			refFace.face, fa.face_pfx, sys.cgi[charpn].portraitscale*sys.fightScreen.portraitScale, &fa.face_lay.window)

		// Draw KO layer
		if !refChar.alive() {
			fa.ko.Draw(float32(fa.pos[0])+sys.fightScreen.offsetX, float32(fa.pos[1]), layerno, sys.fightScreen.scale)
		}

		// Restore original system brightness
		sys.brightness = oldBright
	}

	// Draw top layer
	fa.top.Draw(float32(fa.pos[0])+sys.fightScreen.offsetX, float32(fa.pos[1]), layerno, sys.fightScreen.scale)
}

func (fa *FightScreenFace) drawTeammates(layerno int16) {
	if len(fa.teammate_face) == 0 {
		return
	}

	// We could check this, but not checking it forces our code to be cleaner
	//if sys.tmode[sys.chars[charpn][0].teamside] != TM_Turns {
	//	return
	//}

	// Save current brightness
	oldBright := sys.brightness

	// Reset system brightness if player initiated SuperPause (cancel "darken" parameter)
	// Update: Mugen doesn't do this and it doesn't make much sense to begin with
	//refChar := sys.chars[charpn][0]
	//if refChar.ignoreDarkenTime > 0 {
	//	sys.brightness = 1.0
	//}

	teamSize := int32(len(fa.teammate_face))

	// Determine how many teammates to draw
	visibleCount := teamSize - 1
	if fa.teammate_ko_hide {
		visibleCount -= fa.numko
	}

	if visibleCount <= 0 {
		sys.brightness = oldBright
		return
	}

	// Calculate starting offset for the furthest teammate slot
	x := float32(fa.teammate_pos[0] + fa.teammate_spacing[0]*(visibleCount-1))
	y := float32(fa.teammate_pos[1] + fa.teammate_spacing[1]*(visibleCount-1))

	// Loop backwards through the visible slots
	for j := visibleCount - 1; j >= 0; j-- {
		// Round-robin rotation
		// Start after active player, wrap around to dead players so they appear last
		i := (j + fa.numko + 1) % teamSize

		// Skip the active player (numko) if the modulo lands on them
		if i == fa.numko {
			i = (i + 1) % teamSize
		}

		// Draw backgrounds
		fa.teammate_bg.Draw((x + sys.fightScreen.offsetX), y, layerno, sys.fightScreen.scale)
		fa.teammate_bg0.Draw((x + sys.fightScreen.offsetX), y, layerno, sys.fightScreen.scale)
		fa.teammate_bg1.Draw((x + sys.fightScreen.offsetX), y, layerno, sys.fightScreen.scale)
		fa.teammate_bg2.Draw((x + sys.fightScreen.offsetX), y, layerno, sys.fightScreen.scale)

		// Fetch the PalFX for this member
		pfx := fa.teammate_face_pfx[i]

		// TODO: Defeated PalFX
		//if pfx != nil && i < fa.numko {
		//}

		// Draw face
		fa.teammate_face_lay.DrawFaceSprite((x+sys.fightScreen.offsetX)*sys.fightScreen.scale, y*sys.fightScreen.scale, layerno,
			fa.teammate_face[i], pfx, fa.teammate_scale[i]*sys.fightScreen.portraitScale, &fa.teammate_face_lay.window)

		// Draw KO layer
		if i < fa.numko {
			fa.teammate_ko.Draw((x + sys.fightScreen.offsetX), y, layerno, sys.fightScreen.scale)
		}

		// Shift offset back toward the anchor
		x -= float32(fa.teammate_spacing[0])
		y -= float32(fa.teammate_spacing[1])
	}

	// Restore original system brightness
	sys.brightness = oldBright
}

type FightScreenName struct {
	pos                   [2]int32
	name                  FSText
	bg                    AnimLayout
	top                   AnimLayout
	teammate_pos          [2]int32
	teammate_spacing      [2]int32
	teammate_name         FSText
	teammate_name_strings []string
	teammate_bg           AnimLayout
	numko                 int32
	teammate_ko_hide      bool
	leaderontop           bool
}

func newFightScreenName() *FightScreenName {
	return &FightScreenName{}
}

func readFightScreenName(pre string, is IniSection, sff *Sff, at AnimationTable, f map[int]*Fnt) *FightScreenName {
	nm := newFightScreenName()

	is.ReadI32(pre+"pos", &nm.pos[0], &nm.pos[1])
	is.ReadBool("leaderontop", &nm.leaderontop)
	nm.name = *readFSText(pre+"name.", is, "", 0, f, 0)
	nm.bg = ReadAnimLayout(pre+"bg.", is, sff, at, 0)
	nm.top = ReadAnimLayout(pre+"top.", is, sff, at, 0)

	// Teammates
	is.ReadI32(pre+"teammate.pos", &nm.teammate_pos[0], &nm.teammate_pos[1])
	is.ReadI32(pre+"teammate.spacing", &nm.teammate_spacing[0], &nm.teammate_spacing[1])
	nm.teammate_name = *readFSText(pre+"teammate.name.", is, "", 0, f, 0)
	nm.teammate_bg = ReadAnimLayout(pre+"teammate.bg.", is, sff, at, 0)
	is.ReadBool(pre+"teammate.ko.hide", &nm.teammate_ko_hide)
	return nm
}

func (nm *FightScreenName) step() {
	nm.bg.Action()
	nm.teammate_bg.Action()
	nm.top.Action()
	nm.name.step()
	nm.teammate_name.step()
}

func (nm *FightScreenName) reset() {
	nm.bg.Reset()
	nm.teammate_bg.Reset()
	nm.top.Reset()
}

func (nm *FightScreenName) bgDraw(layerno int16) {
	nm.bg.Draw(float32(nm.pos[0])+sys.fightScreen.offsetX, float32(nm.pos[1]), layerno, sys.fightScreen.scale)
}

func (nm *FightScreenName) draw(layerno int16, charpn int, f map[int]*Fnt, side int) {
	if nm.name.font[0] >= 0 && getFont(f, nm.name.font[0]) != nil {
		nm.name.lay.DrawText((float32(nm.pos[0]) + sys.fightScreen.offsetX), float32(nm.pos[1]), sys.fightScreen.scale, layerno,
			sys.cgi[charpn].lifebarname, getFont(f, nm.name.font[0]), nm.name.font[1], nm.name.font[2], nm.name.palfx, nm.name.frgba)
	}

	nm.top.Draw(float32(nm.pos[0])+sys.fightScreen.offsetX, float32(nm.pos[1]), layerno, sys.fightScreen.scale)
}

func (nm *FightScreenName) drawTeammates(layerno int16, f map[int]*Fnt, side int) {
	if len(nm.teammate_name_strings) == 0 {
		return
	}

	teamSize := int32(len(nm.teammate_name_strings))

	// Determine how many names to draw
	visibleCount := teamSize - 1
	if nm.teammate_ko_hide {
		visibleCount -= nm.numko
	}

	if visibleCount <= 0 {
		return
	}

	// Calculate starting offset for the furthest teammate slot
	x := float32(nm.teammate_pos[0] + nm.teammate_spacing[0]*(visibleCount-1))
	y := float32(nm.teammate_pos[1] + nm.teammate_spacing[1]*(visibleCount-1))

	// Loop backwards through visible slots
	for j := visibleCount - 1; j >= 0; j-- {
		// Round-robin rotation
		// Start after active player, wrap around to dead players so they appear last
		i := (j + nm.numko + 1) % teamSize

		if i == nm.numko {
			i = (i + 1) % teamSize
		}

		// Draw background
		nm.teammate_bg.Draw((x + sys.fightScreen.offsetX), y, layerno, sys.fightScreen.scale)

		// Draw pre-compiled name text
		if nm.teammate_name.font[0] >= 0 && getFont(f, nm.teammate_name.font[0]) != nil {
			nm.teammate_name.lay.DrawText((x + sys.fightScreen.offsetX), y, sys.fightScreen.scale, layerno,
				nm.teammate_name_strings[i], getFont(f, nm.teammate_name.font[0]),
				nm.teammate_name.font[1], nm.teammate_name.font[2], nm.teammate_name.palfx, nm.teammate_name.frgba)
		}

		// Shift offset back toward the anchor
		x -= float32(nm.teammate_spacing[0])
		y -= float32(nm.teammate_spacing[1])
	}

	// TODO: Confirm it teammates are really supposed to not have "top" layer
}

type FightScreenWinIcon struct {
	pos         [2]int32
	iconoffset  [2]int32
	useiconupto int32
	counter     FSText
	bg0         AnimLayout
	top         AnimLayout
	icon        [WT_NumTypes]AnimLayout
	wins        []WinType
	numWins     int
	added       *Animation
	addedP      *Animation
	addedC      *Animation
}

func newFightScreenWinIcon() *FightScreenWinIcon {
	return &FightScreenWinIcon{useiconupto: 4}
}

func readFightScreenWinIcon(pre string, is IniSection, sff *Sff, at AnimationTable, f map[int]*Fnt) *FightScreenWinIcon {
	wi := newFightScreenWinIcon()

	is.ReadI32(pre+"pos", &wi.pos[0], &wi.pos[1])
	is.ReadI32(pre+"iconoffset", &wi.iconoffset[0], &wi.iconoffset[1])
	is.ReadI32("useiconupto", &wi.useiconupto)
	wi.counter = *readFSText(pre+"counter.", is, "%i", 0, f, 0)
	wi.bg0 = ReadAnimLayout(pre+"bg0.", is, sff, at, 0)
	wi.top = ReadAnimLayout(pre+"top.", is, sff, at, 0)
	wi.icon[WT_Normal] = ReadAnimLayout(pre+"n.", is, sff, at, 0)
	wi.icon[WT_Special] = ReadAnimLayout(pre+"s.", is, sff, at, 0)
	wi.icon[WT_Hyper] = ReadAnimLayout(pre+"h.", is, sff, at, 0)
	wi.icon[WT_Cheese] = ReadAnimLayout(pre+"c.", is, sff, at, 0)
	wi.icon[WT_Time] = ReadAnimLayout(pre+"t.", is, sff, at, 0)
	wi.icon[WT_Throw] = ReadAnimLayout(pre+"throw.", is, sff, at, 0)
	wi.icon[WT_Suicide] = ReadAnimLayout(pre+"suicide.", is, sff, at, 0)
	wi.icon[WT_Teammate] = ReadAnimLayout(pre+"teammate.", is, sff, at, 0)
	wi.icon[WT_Perfect] = ReadAnimLayout(pre+"perfect.", is, sff, at, 0)
	wi.icon[WT_Clutch] = ReadAnimLayout(pre+"clutch.", is, sff, at, 0)
	return wi
}

func (wi *FightScreenWinIcon) add(wt WinType) {
	wi.wins = append(wi.wins, wt)
	if wt >= WT_PNormal && wt < WT_CNormal {
		wi.addedP = &Animation{}
		wi.addedP = wi.icon[WT_Perfect].anim
		wi.addedP.Reset()
		wt -= WT_PNormal
	} else if wt >= WT_CNormal {
		wi.addedC = &Animation{}
		wi.addedC = wi.icon[WT_Clutch].anim
		wi.addedC.Reset()
		wt -= WT_CNormal
	}
	wi.added = &Animation{}
	wi.added = wi.icon[wt].anim
	wi.added.Reset()
}

func (wi *FightScreenWinIcon) step(numwin int32) {
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
	if wi.addedC != nil {
		wi.addedC.Action()
	}
}

func (wi *FightScreenWinIcon) reset() {
	wi.bg0.Reset()
	wi.top.Reset()
	for i := range wi.icon {
		wi.icon[i].Reset()
	}
	wi.numWins = len(wi.wins)
	wi.added, wi.addedP, wi.addedC = nil, nil, nil
}

func (wi *FightScreenWinIcon) clear() {
	wi.wins = nil
}

func (wi *FightScreenWinIcon) draw(layerno int16, f map[int]*Fnt, side int) {
	bg0num := sys.matchWins[side&1]

	// Already handled when sys.matchWins is defined
	//if sys.tmode[^side&1] == TM_Turns {
	//	bg0num = sys.numTurns[^side&1]
	//}

	iconLimit := Min(wi.useiconupto, bg0num)

	for i := 0; i < int(iconLimit); i++ {
		wi.bg0.Draw(float32(wi.pos[0]+wi.iconoffset[0]*int32(i))+sys.fightScreen.offsetX,
			float32(wi.pos[1]+wi.iconoffset[1]*int32(i)), layerno, sys.fightScreen.scale)
	}

	if len(wi.wins) > int(wi.useiconupto) {
		// Use the generic win count icon if icon limit exceeded
		if wi.counter.font[0] >= 0 && getFont(f, wi.counter.font[0]) != nil {
			wi.counter.lay.DrawText(float32(wi.pos[0])+sys.fightScreen.offsetX, float32(wi.pos[1]), sys.fightScreen.scale,
				layerno, strings.Replace(wi.counter.text, "%i", fmt.Sprintf("%v", len(wi.wins)), 1),
				getFont(f, wi.counter.font[0]), wi.counter.font[1], wi.counter.font[2], wi.counter.palfx, wi.counter.frgba)
		}
	} else {
		// Use the specific win type icons
		i := 0
		for ; i < wi.numWins; i++ {
			wt, pf, cl := wi.wins[i], false, false
			if wt >= WT_PNormal && wt < WT_CNormal {
				wt -= WT_PNormal
				pf = true
			} else if wt >= WT_CNormal {
				wt -= WT_CNormal
				cl = true
			}
			wi.icon[wt].Draw(float32(wi.pos[0]+wi.iconoffset[0]*int32(i))+sys.fightScreen.offsetX,
				float32(wi.pos[1]+wi.iconoffset[1]*int32(i)), layerno, sys.fightScreen.scale)
			if pf {
				wi.icon[WT_Perfect].Draw(float32(wi.pos[0]+wi.iconoffset[0]*int32(i))+sys.fightScreen.offsetX,
					float32(wi.pos[1]+wi.iconoffset[1]*int32(i)), layerno, sys.fightScreen.scale)
			}
			if cl {
				wi.icon[WT_Clutch].Draw(float32(wi.pos[0]+wi.iconoffset[0]*int32(i))+sys.fightScreen.offsetX,
					float32(wi.pos[1]+wi.iconoffset[1]*int32(i)), layerno, sys.fightScreen.scale)
			}
		}
		if wi.added != nil {
			wt, pf, cl := wi.wins[i], false, false
			if wi.addedP != nil {
				wt -= WT_PNormal
				pf = true
			} else if wi.addedC != nil {
				wt -= WT_CNormal
				cl = true
			}
			wi.icon[wt].lay.DrawAnim(&wi.icon[wt].lay.window,
				float32(wi.pos[0]+wi.iconoffset[0]*int32(i))+sys.fightScreen.offsetX,
				float32(wi.pos[1]+wi.iconoffset[1]*int32(i)), sys.fightScreen.scale, 1, 1, layerno, wi.added, nil)
			if pf {
				wi.icon[WT_Perfect].lay.DrawAnim(&wi.icon[WT_Perfect].lay.window,
					float32(wi.pos[0]+wi.iconoffset[0]*int32(i))+sys.fightScreen.offsetX,
					float32(wi.pos[1]+wi.iconoffset[1]*int32(i)), sys.fightScreen.scale, 1, 1, layerno, wi.addedP, nil)
			}
			if cl {
				wi.icon[WT_Clutch].lay.DrawAnim(&wi.icon[WT_Clutch].lay.window,
					float32(wi.pos[0]+wi.iconoffset[0]*int32(i))+sys.fightScreen.offsetX,
					float32(wi.pos[1]+wi.iconoffset[1]*int32(i)), sys.fightScreen.scale, 1, 1, layerno, wi.addedC, nil)
			}
		}
	}

	for i := 0; i < int(iconLimit); i++ {
		wi.top.Draw(float32(wi.pos[0]+wi.iconoffset[0]*int32(i))+sys.fightScreen.offsetX,
			float32(wi.pos[1]+wi.iconoffset[1]*int32(i)), layerno, sys.fightScreen.scale)
	}
}

type FightScreenTime struct {
	pos            [2]int32
	counter        map[int32]*FSText
	bg             AnimLayout
	top            AnimLayout
	framespercount int32 // The original value as read in fight.def
	activeIdx      int32
}

func newFightScreenTime() *FightScreenTime {
	return &FightScreenTime{
		counter:        make(map[int32]*FSText),
		framespercount: 60,
	}
}

func readFightScreenTime(is IniSection, sff *Sff, at AnimationTable, f map[int]*Fnt) *FightScreenTime {
	ti := newFightScreenTime()

	is.ReadI32("pos", &ti.pos[0], &ti.pos[1])
	ti.counter[0] = readFSText("counter.", is, "", 0, f, 0)
	for k, v := range readMultipleFSText("", "counter", is, "", 0, f, 0) {
		ti.counter[k] = v
	}
	ti.bg = ReadAnimLayout("bg.", is, sff, at, 0)
	ti.top = ReadAnimLayout("top.", is, sff, at, 0)
	is.ReadI32("framespercount", &ti.framespercount)
	return ti
}

func (ti *FightScreenTime) step() {
	ti.bg.Action()
	ti.top.Action()
	ti.counter[ti.activeIdx].step()
}

func (ti *FightScreenTime) reset() {
	ti.bg.Reset()
	ti.top.Reset()
}

func (ti *FightScreenTime) bgDraw(layerno int16) {
	ti.bg.Draw(float32(ti.pos[0])+sys.fightScreen.offsetX, float32(ti.pos[1]), layerno, sys.fightScreen.scale)
}

func (ti *FightScreenTime) draw(layerno int16, f map[int]*Fnt) {
	if sys.curFramesPerCount > 0 &&
		ti.counter[0].font[0] >= 0 && getFont(f, ti.counter[0].font[0]) != nil {
		var timeval int32 = -1
		time := "o"
		if sys.curRoundTime >= 0 {
			if sys.cfg.Config.LegacyTime {
				timeval = sys.curRoundTime / sys.curFramesPerCount
			} else {
				timeval = int32(math.Ceil(float64(sys.curRoundTime) / float64(sys.curFramesPerCount)))
			}
			time = fmt.Sprintf("%v", timeval)
		}
		// Multiple fonts according to time remaining
		var tv int32
		if timeval < 0 { // Infinite time. Select the highest group
			for k := range ti.counter {
				if k > tv {
					tv = k
				}
			}
		} else {
			for k := range ti.counter {
				if k > tv && timeval >= k {
					tv = k
				}
			}
		}
		ti.activeIdx = tv
		ti.counter[tv].lay.DrawText(float32(ti.pos[0])+sys.fightScreen.offsetX, float32(ti.pos[1]), sys.fightScreen.scale, layerno,
			time, getFont(f, ti.counter[tv].font[0]), ti.counter[tv].font[1], ti.counter[tv].font[2], ti.counter[tv].palfx,
			ti.counter[tv].frgba)
	}
	ti.top.Draw(float32(ti.pos[0])+sys.fightScreen.offsetX, float32(ti.pos[1]), layerno, sys.fightScreen.scale)
}

type FightScreenCombo struct {
	pos          [2]int32
	start_x      float32
	counter      map[int32]*FSText
	text         map[int32]*FSText
	bg           AnimLayout
	top          AnimLayout
	displaytime  int32
	showspeed    float32
	hidespeed    float32
	separator    string
	places       int32
	shownHits    int32
	shownDmg     int32
	shownPct     float32
	resttime     int32
	counterX     float32
	autoalign    bool
	newCombo     bool
	counterShake ComboShake
	textShake    ComboShake
}

func newFightScreenCombo() *FightScreenCombo {
	return &FightScreenCombo{
		displaytime: 90,
		showspeed:   8,
		hidespeed:   4,
		counter:     make(map[int32]*FSText),
		text:        make(map[int32]*FSText),
		autoalign:   true,
		counterShake: ComboShake{
			freq:  60,  // Same as EnvShake
			decay: 1.0, // Linear
			scale: 1.0, // No change
		},
		textShake: ComboShake{
			freq:  60,
			decay: 1.0,
			scale: 1.0,
		},
	}
}

func readFightScreenCombo(pre string, is IniSection,
	sff *Sff, at AnimationTable, f map[int]*Fnt, side int) *FightScreenCombo {
	co := newFightScreenCombo()
	is.ReadI32(pre+"pos", &co.pos[0], &co.pos[1])
	is.ReadF32(pre+"start.x", &co.start_x)
	if side == 1 {
		// Mugen 1.0 implementation reuses Winmugen code where both sides shared the same values
		if pre == "team2." {
			co.start_x = float32(sys.fightScreen.localcoord[0]) - co.start_x
		} else {
			co.pos[0] = sys.fightScreen.localcoord[0] - co.pos[0]
		}
	}
	var align int32
	if pre == "" {
		if side == 0 {
			align = 1
		} else {
			align = -1
		}
	}

	// Read main counter string then optional multiple values
	co.counter[0] = readFSText(pre+"counter.", is, "%i", 2, f, align)
	for k, v := range readMultipleFSText(pre, "counter", is, "%i", 2, f, align) {
		co.counter[k] = v
	}

	// Old counter shake syntax
	// TODO: Shouldn't shaking parameters also support "multiple values"?
	var old bool
	if is.ReadBool(pre+"counter.shake", &old) {
		if old {
			co.counterShake.time = 8     // Mugen value
			co.counterShake.phase = 90   // Like old Ikemen default
			co.counterShake.scale = 1.35 // Derived from old Ikemen formula
		}
	}
	is.ReadI32(pre+"counter.time", &co.counterShake.time)
	var mult float32
	if is.ReadF32(pre+"counter.mult", &mult) {
		co.counterShake.scale = 1 + float32(co.counterShake.time)*mult
	}

	// New counter shake syntax
	is.ReadI32(pre+"counter.shake.time", &co.counterShake.time)
	if is.ReadF32(pre+"counter.shake.freq", &co.counterShake.freq) {
		co.counterShake.setDefaultPhase()
	}
	is.ReadF32(pre+"counter.shake.phase", &co.counterShake.phase)
	is.ReadF32(pre+"counter.shake.ampl", &co.counterShake.ampl)
	is.ReadF32(pre+"counter.shake.scale", &co.counterShake.scale)
	is.ReadF32(pre+"counter.shake.dir", &co.counterShake.dir)
	is.ReadF32(pre+"counter.shake.diradd", &co.counterShake.diradd)
	is.ReadF32(pre+"counter.shake.decay", &co.counterShake.decay)

	// Read main text string then optional multiple values
	co.text[0] = readFSText(pre+"text.", is, "", 2, f, align)
	for k, v := range readMultipleFSText(pre, "text", is, "", 2, f, align) {
		co.text[k] = v
	}

	// Text shake
	is.ReadI32(pre+"text.shake.time", &co.textShake.time)
	if is.ReadF32(pre+"text.shake.freq", &co.textShake.freq) {
		co.textShake.setDefaultPhase()
	}
	is.ReadF32(pre+"text.shake.phase", &co.textShake.phase)
	is.ReadF32(pre+"text.shake.ampl", &co.textShake.ampl)
	is.ReadF32(pre+"text.shake.scale", &co.textShake.scale)
	is.ReadF32(pre+"text.shake.dir", &co.textShake.dir)
	is.ReadF32(pre+"text.shake.diradd", &co.textShake.diradd)
	is.ReadF32(pre+"text.shake.decay", &co.textShake.decay)

	co.bg = ReadAnimLayout(pre+"bg0.", is, sff, at, 2)
	co.top = ReadAnimLayout(pre+"top.", is, sff, at, 2)
	is.ReadI32(pre+"displaytime", &co.displaytime)
	is.ReadF32(pre+"showspeed", &co.showspeed)
	co.showspeed = Max(1, co.showspeed)
	is.ReadF32(pre+"hidespeed", &co.hidespeed)
	co.separator, _, _ = is.getText("format.decimal.separator")
	is.ReadI32("format.decimal.places", &co.places)
	is.ReadBool(pre+"autoalign", &co.autoalign)

	return co
}

func (co *FightScreenCombo) step(side int, hits, damage int32, percentage float32) {
	co.bg.Action()
	co.top.Action()

	// Allows a new identical combo to still retrigger later
	if hits < 2 && co.shownHits >= 2 {
		co.newCombo = true
	}

	// Reset team combo if no player was found getting hit
	if hits == 0 {
		sys.comboCount[side] = 0
	}
	trueHits := sys.comboCount[side]

	// Handle show/hide speed
	if co.resttime > 0 {
		// Slide in
		co.counterX -= co.counterX / co.showspeed
		// Snap to visible position
		if Abs(co.counterX) < 1 {
			co.counterX = 0
		}
	} else if trueHits < 2 {
		// Slide out when combo ends
		co.counterX -= sys.fightScreen.fnt_scale * co.hidespeed * float32(sys.fightScreen.localcoord[0]) / 320
		// Snap to starting position
		if co.counterX < co.start_x*2 {
			co.counterX = co.start_x * 2
		}
	}

	// The displayed time only decrements when the counter is in the visible position
	// Currently, the way the timer decrements while the combo is still ongoing can make it stay visible varying amounts of time past the end of the combo
	// This makes it not always sync correctly with the "nice combo" actions
	if co.counterX == 0 {
		co.resttime--
	}

	// Update if number of hits or total damage change
	if trueHits >= 2 && (co.newCombo || co.shownHits != trueHits || co.shownDmg != damage) {
		// Reset visuals when hits changed
		if co.newCombo || co.shownHits != trueHits {
			co.counterShake.restart()
			co.textShake.restart()
			for i := range co.counter {
				co.counter[i].resetTxtPfx()
			}
			for i := range co.text {
				co.text[i].resetTxtPfx()
			}
		}
		// Time resets if either hits or damage changed
		co.resttime = co.displaytime
		// Update state
		co.shownHits = trueHits
		co.shownDmg = damage
		co.shownPct = percentage
		co.newCombo = false
	}

	// Update shaking effects
	co.counterShake.update()
	co.textShake.update()

	// Multiple counter fonts
	var cv int32
	for k := range co.counter {
		if k > cv && co.shownHits >= k {
			cv = k
		}
	}

	// Multiple text fonts
	var tv int32
	for k := range co.text {
		if k > tv && co.shownHits >= k {
			tv = k
		}
	}

	if co.counterX != co.start_x*2 {
		co.counter[cv].step()
		co.text[tv].step()
	}
}

func (co *FightScreenCombo) reset() {
	co.bg.Reset()
	co.top.Reset()
	co.shownHits = 0
	co.shownDmg = 0
	co.shownPct = 0
	co.resttime = 0
	co.counterX = co.start_x * 2

	// Reset combo shake
	co.counterShake.curTime = 0
	co.textShake.curTime = 0
}

func (co *FightScreenCombo) draw(layerno int16, f map[int]*Fnt, side int) {
	if co.resttime <= 0 && co.counterX == co.start_x*2 {
		return
	}

	// Multiple counter fonts according to combo hits
	var cv int32
	for k := range co.counter {
		if k > cv && co.shownHits >= k {
			cv = k
		}
	}

	// Multiple text fonts according to combo hits
	var tv int32
	for k := range co.text {
		if k > tv && co.shownHits >= k {
			tv = k
		}
	}

	// Replace operator with current combo value
	counter := strings.Replace(co.counter[cv].text, "%i", fmt.Sprintf("%v", co.shownHits), 1)

	// Compute base x position
	x := float32(co.pos[0])
	if side == 0 {
		if co.start_x <= 0 {
			x += co.counterX
		}
		// Apply autoalign
		if co.counter[cv].font[0] >= 0 && co.autoalign {
			if ff := getFont(f, co.counter[cv].font[0]); ff != nil {
				x += float32(ff.TextWidth(counter, co.counter[cv].font[1], 0)) *
					co.counter[cv].lay.scale[0] * sys.fightScreen.fnt_scale
			}
		}
	} else {
		if co.start_x <= 0 {
			x -= co.counterX
		}
	}

	// BG
	co.bg.Draw(x+sys.fightScreen.offsetX, float32(co.pos[1]), layerno, sys.fightScreen.scale)

	// Text block pre-processing
	var ffText *Fnt
	var textBlockWidth float32
	var lineHeight float32
	var lines []string
	if co.text[tv].font[0] >= 0 && getFont(f, co.text[tv].font[0]) != nil {
		// Replace operator with current combo hits
		text := strings.Replace(co.text[tv].text, "%i", fmt.Sprintf("%v", co.shownHits), 1)
		// Replace operator with current combo damage
		text = strings.Replace(text, "%d", fmt.Sprintf("%v", co.shownDmg), 1)

		// Only process combo damage percentage if necessary
		if strings.Contains(co.text[tv].text, "%p") {
			// Truncate the percentage to avoid rounding to 100% unless the enemy is defeated
			truncatedPct := math.Floor(float64(co.shownPct)*math.Pow10(int(co.places))) / math.Pow10(int(co.places))
			// Split float value
			s := strings.Split(fmt.Sprintf("%.[2]*[1]f", truncatedPct, co.places), ".")
			// Decimal separator
			if co.places > 0 && len(s) > 1 {
				s[0] = s[0] + co.separator + s[1]
			}
			// Replace %p with formatted string
			text = strings.Replace(text, "%p", s[0], 1)
		}

		// Split on new line
		lines = strings.Split(text, "\\n")

		// Compute text block dimensions
		ffText = getFont(f, co.text[tv].font[0])
		if ffText != nil {
			lineHeight = float32(ffText.Size[1])*co.text[tv].lay.scale[1]*sys.fightScreen.fnt_scale +
				float32(ffText.Spacing[1])*co.text[tv].lay.scale[1]*sys.fightScreen.fnt_scale
			for _, line := range lines {
				w := float32(ffText.TextWidth(line, co.text[tv].font[1], 0)) *
					co.text[tv].lay.scale[0] * sys.fightScreen.fnt_scale
				if w > textBlockWidth {
					textBlockWidth = w
				}
			}
		}
	}

	// Draw text
	if len(lines) > 0 && ffText != nil {
		blockX := x + sys.fightScreen.offsetX
		blockY := float32(co.pos[1])

		// Apply shake effect
		textShakeScale := co.textShake.getScale()
		textShakeOffset := co.textShake.getOffset()

		drawBlockX := (blockX + textShakeOffset[0]) / textShakeScale
		drawBlockY := (blockY + textShakeOffset[1]) / textShakeScale
		finalScale := textShakeScale * sys.fightScreen.scale

		// Draw each line
		for i, line := range lines {
			y := drawBlockY + float32(i)*lineHeight
			co.text[tv].lay.DrawText(drawBlockX, y, finalScale, layerno, line,
				ffText, co.text[tv].font[1], co.text[tv].font[2],
				co.text[tv].palfx, co.text[tv].frgba)
		}
	}

	// Counter
	if co.counter[cv].font[0] >= 0 && getFont(f, co.counter[cv].font[0]) != nil {
		// Apply autoalign
		var counterShift float32
		if side == 0 && co.autoalign {
			if ff := getFont(f, co.counter[cv].font[0]); ff != nil {
				counterShift = float32(ff.TextWidth(counter, co.counter[cv].font[1], 0)) *
					co.counter[cv].lay.scale[0] * sys.fightScreen.fnt_scale
			}
		}
		if side == 1 && co.autoalign {
			counterShift = counterShift + textBlockWidth
		}

		// Apply shake effect
		shakeScale := co.counterShake.getScale()
		shakeOffset := co.counterShake.getOffset()

		drawX := (x - counterShift + sys.fightScreen.offsetX + shakeOffset[0]) / shakeScale
		drawY := (float32(co.pos[1]) + shakeOffset[1]) / shakeScale
		finalScale := shakeScale * sys.fightScreen.scale

		co.counter[cv].lay.DrawText(drawX, drawY, finalScale, layerno,
			counter, getFont(f, co.counter[cv].font[0]), co.counter[cv].font[1], co.counter[cv].font[2],
			co.counter[cv].palfx, co.counter[cv].frgba)
	}

	// Top
	co.top.Draw(x+sys.fightScreen.offsetX, float32(co.pos[1]), layerno, sys.fightScreen.scale)
}

// This is like a miniature EnvShake applied to a single object
type ComboShake struct {
	time      int32
	ampl      float32
	scale     float32
	freq      float32
	phase     float32
	dir       float32 // It's an angle but this name is consistent with EnvShake
	diradd    float32
	decay     float32
	curTime   int32
	curOffset [2]float32
	curScale  float32
}

// Uses a reset instead of a clear because we don't want to lose the original parameters
func (cs *ComboShake) reset() {
	cs.curTime = 0
	cs.curOffset = [2]float32{0, 0}
	cs.curScale = 1
}

func (cs *ComboShake) restart() {
	if cs.time <= 0 {
		cs.reset() // Just in case
		return
	}
	cs.curTime = cs.time
}

func (cs *ComboShake) setDefaultPhase() {
	// No need for NaN
	if cs.freq >= 90.0 {
		cs.phase = 90.0
	} else {
		cs.phase = 0
	}
}

func (cs *ComboShake) update() {
	if cs.curTime <= 0 {
		if cs.time > 0 {
			cs.reset()
		}
		return
	}

	// Handle decay
	var decay float32 = 1
	if cs.decay != 0 && cs.time > 0 {
		t := float32(cs.curTime) / float32(cs.time)
		decay = float32(math.Pow(float64(t), float64(cs.decay)))
	}

	elapsed := float32(cs.time - cs.curTime)

	// Calculate offset and cache it
	// There's no benefit to caching it yet, but it may be useful for future features and keeps the code similar to EnvShake
	if cs.ampl == 0 {
		cs.curOffset = [2]float32{0, 0}
	} else {
		curAmp := cs.ampl * decay
		phaseRad := Rad(cs.phase) + Rad(cs.freq)*elapsed
		val := curAmp * float32(math.Sin(float64(phaseRad)))

		currentAngleDeg := cs.dir + cs.diradd*elapsed
		radAng := Rad(currentAngleDeg)

		x := val * float32(math.Cos(float64(radAng)))
		y := -val * float32(math.Sin(float64(radAng))) // Negated for counter‑clockwise

		cs.curOffset = [2]float32{x, y}
	}

	// Calculate scale and cache it
	if cs.scale == 1 {
		cs.curScale = 1
	} else {
		phaseRad := Rad(cs.phase) + Rad(cs.freq)*elapsed
		sinVal := float32(math.Sin(float64(phaseRad)))
		exponent := float32(math.Log(float64(cs.scale))) * decay * sinVal
		cs.curScale = float32(math.Exp(float64(exponent)))
	}

	// Step timer only after computing offset and scale
	cs.curTime--
}

func (cs *ComboShake) getOffset() [2]float32 {
	return cs.curOffset
}

func (cs *ComboShake) getScale() float32 {
	return cs.curScale
}

type FSMsg struct {
	resttime     int32
	agetimer     int32
	counterX     float32
	text         string
	fontNo       int32
	fontBank     int32
	fontAlign    int32
	fontColor    [4]int32
	fontColorSet bool
	palfx        *PalFX
	bg           AnimLayout
	front        AnimLayout
	del          bool
	snd          [2]int32
	snd_ffx      string
	spr          [2]int32
	animNo       int32
	anim_ffx     string
	top          bool
}

func newFSMsg(side int) *FSMsg {
	return &FSMsg{
		resttime:  -1,
		counterX:  sys.fightScreen.actions[side].start_x * 2,
		fontNo:    -1,
		fontBank:  -1,
		fontAlign: IErr, // Default to the font's
		fontColor: [4]int32{255, 255, 255, 255},
		snd:       [2]int32{-1, -1},
		spr:       [2]int32{-1, -1},
		animNo:    -1,
		top:       true,
	}
}

func insertFSMsg(array []*FSMsg, value *FSMsg, index int) []*FSMsg {
	// Remove one empty index before appending the new one
	// This reduces the chance of an empty space pushing older messages
	if index == 0 {
		// Check from the top
		for i := 0; i < len(array); i++ {
			if array[i].del {
				// Remove the element at index i by shifting elements down
				copy(array[i:], array[i+1:])
				array = array[:len(array)-1]
				break
			}
		}
	} else {
		// Check from the bottom
		for i := len(array) - 1; i >= index; i-- {
			if array[i].del {
				// Remove the element at index i by shifting elements down
				copy(array[i:], array[i+1:])
				array = array[:len(array)-1]
				break
			}
		}
	}
	// Insert the new message at the specified index
	if index == len(array) {
		array = append(array, value)
	} else {
		array = append(array[:index], append([]*FSMsg{value}, array[index:]...)...)
	}
	return array
}

func removeFSMsg(array []*FSMsg, index int) []*FSMsg {
	return SliceDelete(array, index)
}

func (m *FSMsg) isEqual(other *FSMsg) bool {
	if m == other {
		return true
	}
	if m == nil || other == nil {
		return false
	}
	// Compare font settings
	if m.fontNo != other.fontNo ||
		m.fontBank != other.fontBank ||
		m.fontAlign != other.fontAlign ||
		m.fontColorSet != other.fontColorSet ||
		m.fontColor != other.fontColor {
		return false
	}
	// Compare animation
	if m.animNo != other.animNo || m.anim_ffx != other.anim_ffx {
		return false
	}
	// Compare sprite
	if m.spr != other.spr {
		return false
	}
	// Compare text (slowest last)
	if m.text != other.text {
		return false
	}
	return true
}

type FightScreenAction struct {
	oldleader   int
	pos         [2]int32
	spacing     [2]int32
	start_x     float32
	text        FSText
	displaytime int32
	showspeed   float32
	hidespeed   float32
	messages    []*FSMsg
	is          IniSection
	max         int32
}

func newFightScreenAction() *FightScreenAction {
	return &FightScreenAction{
		displaytime: 90,
		showspeed:   8,
		hidespeed:   4,
		max:         8,
		is:          make(map[string]string),
	}
}

func readFightScreenAction(pre string, is IniSection, f map[int]*Fnt) *FightScreenAction {
	ac := newFightScreenAction()
	ac.is = is
	is.ReadI32(pre+"pos", &ac.pos[0], &ac.pos[1])
	is.ReadI32(pre+"spacing", &ac.spacing[0], &ac.spacing[1])
	is.ReadF32(pre+"start.x", &ac.start_x)
	if pre == "team2." {
		ac.start_x = float32(sys.fightScreen.localcoord[0]) - ac.start_x
	}
	ac.text = *readFSText(pre+"text.", is, "", 2, f, 0)
	is.ReadI32(pre+"displaytime", &ac.displaytime)
	is.ReadF32(pre+"showspeed", &ac.showspeed)
	ac.showspeed = Max(1, ac.showspeed)
	is.ReadF32(pre+"hidespeed", &ac.hidespeed)
	is.ReadI32(pre+"max", &ac.max)
	return ac
}

func (ac *FightScreenAction) step(leader int) {
	if ac.oldleader != leader {
		ac.oldleader = leader
	}
	for _, v := range ac.messages {
		v.bg.Action()
		v.front.Action()
		if v.palfx != nil {
			v.palfx.step()
		}
		if v.resttime > 0 {
			v.counterX -= v.counterX / ac.showspeed
		} else {
			v.counterX -= sys.fightScreen.fnt_scale * ac.hidespeed * float32(sys.fightScreen.localcoord[0]) / 320
			if v.counterX < ac.start_x*2 {
				v.del = true
			}
		}
		if Abs(v.counterX) < 1 {
			v.resttime--
		}
		v.agetimer++
	}
	if len(ac.messages) > 0 && ac.messages[len(ac.messages)-1].del {
		ac.messages = removeFSMsg(ac.messages, len(ac.messages)-1)
	}
	ac.text.step()
}

func (ac *FightScreenAction) reset(leader int) {
	ac.oldleader = leader
	ac.messages = []*FSMsg{}
}

func (ac *FightScreenAction) draw(layerno int16, f map[int]*Fnt, side int) {
	for k, v := range ac.messages {
		if v.resttime <= 0 && v.counterX == ac.start_x*2 {
			continue
		}
		x := float32(ac.pos[0])
		if side == 0 {
			if ac.start_x <= 0 {
				x += v.counterX
			}
		} else {
			if ac.start_x <= 0 {
				x -= v.counterX
			}
		}
		// Previously the spacing accounted for the size of the current element, like sprite tilespacing
		// That created a lot of trouble when aligning things, so currently the spacing is fixed, like typical font and animation spacing
		// https://github.com/ikemen-engine/Ikemen-GO/issues/2166
		// Draw background
		if v.bg.anim.spr != nil {
			v.bg.Draw(x+sys.fightScreen.offsetX+float32(k)*float32(ac.spacing[0]),
				float32(ac.pos[1])+float32(k)*float32(ac.spacing[1]),
				layerno, sys.fightScreen.scale)
		}
		// Draw animation/sprite
		if v.front.anim.spr != nil {
			v.front.Draw(x+sys.fightScreen.offsetX+float32(k)*float32(ac.spacing[0]),
				float32(ac.pos[1])+float32(k)*float32(ac.spacing[1]),
				layerno, sys.fightScreen.scale)
		}
		// Draw text
		if v.text != "" && ac.text.font[0] >= 0 {
			fontno := ac.text.font[0]
			if v.fontNo >= 0 {
				fontno = v.fontNo
			}

			ff := getFont(f, fontno)
			if ff != nil {
				bank := ac.text.font[1]
				if v.fontBank >= 0 {
					bank = v.fontBank
				}
				align := ac.text.font[2]
				if v.fontAlign != IErr {
					align = v.fontAlign
				}
				palfx := ac.text.palfx

				if v.palfx != nil {
					palfx = v.palfx
				}

				frgba := ac.text.frgba
				if v.fontColorSet {
					r := Clamp(v.fontColor[0], 0, 255)
					g := Clamp(v.fontColor[1], 0, 255)
					b := Clamp(v.fontColor[2], 0, 255)
					a := Clamp(v.fontColor[3], 0, 255)

					frgba[0] = float32(r) / 255
					frgba[1] = float32(g) / 255
					frgba[2] = float32(b) / 255
					frgba[3] = float32(a) / 255

					pf := newPalFX()
					if palfx != nil {
						*pf = *palfx
					}
					pf.setColor(r, g, b)
					palfx = pf
				}

				ac.text.lay.DrawText(x+sys.fightScreen.offsetX+float32(k)*float32(ac.spacing[0]),
					float32(ac.pos[1])+float32(k)*float32(ac.spacing[1]),
					sys.fightScreen.scale, layerno, v.text, ff, bank, align,
					palfx, frgba)
			}
		}
	}
}

type ResultAnnouncement struct {
	text [2]AnimTextSnd
	top  [2]AnimLayout
	bg   [2][32]AnimLayout
}

type FightScreenRound struct {
	snd               *Snd
	pos               [2]int32
	start_waittime    int32
	round_time        int32
	round_sndtime     int32
	roundDisplayTimer int32
	roundDisplayPhase int32
	round             [9]AnimTextSnd
	round_default     AnimTextSnd
	round_default_top AnimLayout
	round_default_bg  [32]AnimLayout
	round_single      AnimTextSnd
	round_single_top  AnimLayout
	round_single_bg   [32]AnimLayout
	round_final       AnimTextSnd
	round_final_top   AnimLayout
	round_final_bg    [32]AnimLayout
	fight_time        int32
	fight_sndtime     int32
	fightDisplayTimer int32
	fightDisplayPhase int32
	fight             AnimTextSnd
	fight_top         AnimLayout
	fight_bg          [32]AnimLayout
	ctrl_time         int32
	ko_time           int32
	ko_sndtime        int32
	koDisplayTimer    int32
	koDisplayPhase    int32
	dko_time          int32
	dko_sndtime       int32
	dko_showdraw      bool
	to_time           int32
	to_sndtime        int32
	ko, dko, to       AnimTextSnd
	ko_top            AnimLayout
	ko_bg             [32]AnimLayout
	dko_top           AnimLayout
	dko_bg            [32]AnimLayout
	to_top            AnimLayout
	to_bg             [32]AnimLayout
	slow_time         int32
	slow_fadetime     int32
	slow_speed        float32
	over_waittime     int32
	over_hittime      int32
	over_wintime      int32
	over_forcewintime int32
	over_time         int32
	win_time          int32
	win_sndtime       int32
	winDisplayTimer   int32
	winDisplayPhase   int32
	win               [4]ResultAnnouncement
	aiLose            [4]ResultAnnouncement
	aiWin             [4]ResultAnnouncement
	aiWinExists       [2][4]bool
	aiLoseExists      [2][4]bool
	drawgame          AnimTextSnd
	drawgame_top      AnimLayout
	drawgame_bg       [32]AnimLayout
	timerActive       bool
	winType           [WT_NumTypes * 2]FSBgTextSnd
	fadeIn            *Fade
	fadeOut           *Fade
	shutterTimer      int32
	shutter_time      int32
	shutter_col       uint32
	callfight_time    int32
	clutch_threshold  int32
	//match_wins          [2]int32 // Handled by system
	//match_maxdrawgames  [2]int32 // Handled by system
}

func newFightScreenRound(snd *Snd) *FightScreenRound {
	return &FightScreenRound{
		snd:               snd,
		start_waittime:    30,
		ctrl_time:         30,
		slow_time:         60,
		slow_fadetime:     45,
		slow_speed:        0.25,
		over_waittime:     45,
		over_hittime:      10,
		over_wintime:      45,
		over_forcewintime: 900, // Hardcoded in Mugen
		over_time:         210,
		shutter_time:      15,
		callfight_time:    60, // Hardcoded in Mugen
		clutch_threshold:  10,
		//match_wins:         [...]int32{2, 2},
		//match_maxdrawgames: [...]int32{1, 1},
	}
}

func readLbFade(pre string, is IniSection, sff *Sff, at AnimationTable) *Fade {
	fp := newFade()
	is.ReadI32(pre+"time", &fp.time)
	is.ReadI32(pre+"col", &fp.col[0], &fp.col[1], &fp.col[2])
	fp.colEncoded = uint32(fp.col[0]&0xff<<16 | fp.col[1]&0xff<<8 | fp.col[2]&0xff)
	is.ReadI32(pre+"snd", &fp.snd[0], &fp.snd[1])
	anim := int32(-1)
	is.ReadI32(pre+"anim", &anim)
	fp.animData = NewAnim(sff, "-1,0, 0,0, -1")
	//fp.animData.SetLocalcoord(float32(sys.scrrect[2]), float32(sys.scrrect[3]))
	fp.animData.SetLocalcoord(float32(sys.fightScreen.localcoord[0]), float32(sys.fightScreen.localcoord[1]))
	fp.animData.SetPos(0, 0)   // TODO: use fp.animData.Reset() instead
	fp.animData.SetScale(1, 1) // TODO: use fp.animData.Reset() instead
	//fp.animData.Reset() // TODO: comment out pos and scale adjustments
	if animData, exists := at.anims[anim]; exists {
		fp.animData.anim = animData
	} else {
		fp.animData.anim = nil
	}
	return fp
}

func readFightScreenRound(is IniSection,
	sff *Sff, at AnimationTable, snd *Snd, f map[int]*Fnt) *FightScreenRound {
	ro := newFightScreenRound(snd)

	is.ReadI32("pos", &ro.pos[0], &ro.pos[1])
	//is.ReadI32("match.wins", &ro.match_wins[0], &ro.match_wins[1]) // Replaced by ingame options
	//is.ReadI32("match.maxdrawgames", &ro.match_maxdrawgames[0], &ro.match_maxdrawgames[1]) // Replaced by ingame options
	is.ReadI32("start.waittime", &ro.start_waittime)
	if ro.start_waittime < 1 {
		ro.start_waittime = 1
	}
	is.ReadI32("round.time", &ro.round_time)
	ro.round_sndtime = ro.round_time
	is.ReadI32("round.sndtime", &ro.round_sndtime)
	for i := range ro.round {
		ro.round[i] = *ReadAnimTextSnd(fmt.Sprintf("round%v.", i+1), is, sff, at, 2, f)
	}
	ro.round_default = *ReadAnimTextSnd("round.default.", is, sff, at, 2, f)
	ro.round_default_top = ReadAnimLayout("round.default.top.", is, sff, at, 2)
	for i := range ro.round_default_bg {
		ro.round_default_bg[i] = ReadAnimLayout(fmt.Sprintf("round.default.bg%v.", i), is, sff, at, 2)
	}
	// Single round animations and sounds
	ro.round_single = *ReadAnimTextSnd("round.single.", is, sff, at, 2, f)
	ro.round_single_top = ReadAnimLayout("round.single.top.", is, sff, at, 2)
	for i := range ro.round_single_bg {
		ro.round_single_bg[i] = ReadAnimLayout(fmt.Sprintf("round.single.bg%v.", i), is, sff, at, 2)
	}
	// Final round animations and sounds
	ro.round_final = *ReadAnimTextSnd("round.final.", is, sff, at, 2, f)
	ro.round_final_top = ReadAnimLayout("round.final.top.", is, sff, at, 2)
	for i := range ro.round_final_bg {
		ro.round_final_bg[i] = ReadAnimLayout(fmt.Sprintf("round.final.bg%v.", i), is, sff, at, 2)
	}
	is.ReadI32("fight.time", &ro.fight_time)
	ro.fight_sndtime = ro.fight_time
	is.ReadI32("fight.sndtime", &ro.fight_sndtime)
	ro.fight = *ReadAnimTextSnd("fight.", is, sff, at, 2, f)
	ro.fight_top = ReadAnimLayout("fight.top.", is, sff, at, 2)
	for i := range ro.fight_bg {
		ro.fight_bg[i] = ReadAnimLayout(fmt.Sprintf("fight.bg%v.", i), is, sff, at, 2)
	}
	is.ReadI32("ctrl.time", &ro.ctrl_time)
	if ro.ctrl_time < 1 {
		ro.ctrl_time = 1
	}

	// KO
	is.ReadI32("ko.time", &ro.ko_time)
	ro.ko_sndtime = ro.ko_time
	is.ReadI32("ko.sndtime", &ro.ko_sndtime)
	ro.ko = *ReadAnimTextSnd("ko.", is, sff, at, 1, f)
	ro.ko_top = ReadAnimLayout("ko.top.", is, sff, at, 1)
	for i := range ro.ko_bg {
		ro.ko_bg[i] = ReadAnimLayout(fmt.Sprintf("ko.bg%v.", i), is, sff, at, 2)
	}

	// Default new timers to KO timers
	ro.dko_time = ro.ko_time
	ro.dko_sndtime = ro.ko_sndtime
	ro.to_time = ro.ko_time
	ro.to_sndtime = ro.ko_sndtime

	// Double KO
	is.ReadI32("dko.time", &ro.dko_time)
	is.ReadI32("dko.sndtime", &ro.dko_sndtime)
	is.ReadBool("dko.showdraw", &ro.dko_showdraw)
	ro.dko = *ReadAnimTextSnd("dko.", is, sff, at, 1, f)
	ro.dko_top = ReadAnimLayout("dko.top.", is, sff, at, 1)
	for i := range ro.dko_bg {
		ro.dko_bg[i] = ReadAnimLayout(fmt.Sprintf("dko.bg%v.", i), is, sff, at, 2)
	}

	// Time Over
	is.ReadI32("to.time", &ro.to_time)
	is.ReadI32("to.sndtime", &ro.to_sndtime)
	ro.to = *ReadAnimTextSnd("to.", is, sff, at, 1, f)
	ro.to_top = ReadAnimLayout("to.top.", is, sff, at, 1)
	for i := range ro.to_bg {
		ro.to_bg[i] = ReadAnimLayout(fmt.Sprintf("to.bg%v.", i), is, sff, at, 2)
	}

	is.ReadI32("slow.time", &ro.slow_time)
	ro.slow_fadetime = int32(float32(ro.slow_time) * 0.75) // Default fadetime
	is.ReadI32("slow.fadetime", &ro.slow_fadetime)
	if ro.slow_fadetime > ro.slow_time {
		ro.slow_fadetime = ro.slow_time
	}
	is.ReadF32("slow.speed", &ro.slow_speed)
	// A speed over 1 defeats the purpose, 0 freezes the game, and negative is not safe
	ro.slow_speed = Clamp(ro.slow_speed, 0.01, 1)
	is.ReadI32("over.hittime", &ro.over_hittime)
	if ro.over_hittime < 1 {
		ro.over_hittime = 1
	}
	is.ReadI32("over.waittime", &ro.over_waittime)
	if ro.over_waittime < 1 {
		ro.over_waittime = 1
	}
	is.ReadI32("over.wintime", &ro.over_wintime)
	if ro.over_wintime < 1 {
		ro.over_wintime = 1
	}
	is.ReadI32("over.forcewintime", &ro.over_forcewintime)
	if ro.over_forcewintime < 1 {
		ro.over_forcewintime = 1
	}
	is.ReadI32("over.time", &ro.over_time)
	if ro.over_time < 1 {
		ro.over_time = 1
	}
	is.ReadI32("win.time", &ro.win_time)
	ro.win_sndtime = ro.win_time
	is.ReadI32("win.sndtime", &ro.win_sndtime)

	sectionExists := func(pre string) bool {
		_, okT := is[pre+"text"]
		_, okS := is[pre+"spr"]
		_, okA := is[pre+"anim"]
		return okT || okS || okA
	}
	readATS := func(pPre, pre string) AnimTextSnd {
		if sectionExists(pPre) {
			return *ReadAnimTextSnd(pPre, is, sff, at, 1, f)
		}
		return *ReadAnimTextSnd(pre, is, sff, at, 1, f)
	}
	readLayout := func(pPre, pre string) AnimLayout {
		if sectionExists(pPre) {
			return ReadAnimLayout(pPre, is, sff, at, 1)
		}
		return ReadAnimLayout(pre, is, sff, at, 1)
	}
	readBGs := func(target *[32]AnimLayout, pPre, pre string) {
		for j := range target {
			pBg, bg := fmt.Sprintf("%vbg%v.", pPre, j), fmt.Sprintf("%vbg%v.", pre, j)
			if sectionExists(pBg) || sectionExists(bg) {
				target[j] = readLayout(pBg, bg)
			}
		}
	}

	for i := 0; i < 2; i++ {
		p := fmt.Sprintf("p%v.", i+1)
		suffixes := []string{"", "2", "3", "4"}

		keys := []struct {
			name   string
			data   *[4]ResultAnnouncement
			exists *[4]bool
		}{
			{"win", &ro.win, nil},
			{"ai.win", &ro.aiWin, &ro.aiWinExists[i]},
			{"ai.lose", &ro.aiLose, &ro.aiLoseExists[i]},
		}

		for _, key := range keys {
			for idx, s := range suffixes {
				pre := fmt.Sprintf("%v%v.", key.name, s)
				pPre := fmt.Sprintf("%v%v%v.", p, key.name, s)
				exists := sectionExists(pPre) || sectionExists(pre)
				current := &key.data[idx]

				if idx >= 2 && !exists {
					*current = key.data[1]
					if key.exists != nil {
						key.exists[idx] = key.exists[1]
					}
				} else {
					if exists && key.exists != nil {
						key.exists[idx] = true
					}
					current.text[i] = readATS(pPre, pre)
					current.top[i] = readLayout(pPre+"top.", pre+"top.")
					readBGs(&current.bg[i], pPre, pre)
				}
			}
		}
	}
	ro.drawgame = *ReadAnimTextSnd("draw.", is, sff, at, 1, f)
	ro.drawgame_top = ReadAnimLayout("draw.top.", is, sff, at, 1)
	for i := range ro.drawgame_bg {
		ro.drawgame_bg[i] = ReadAnimLayout(fmt.Sprintf("draw.bg%v.", i), is, sff, at, 1)
	}
	ro.winType[WT_Normal] = readFSBgTextSnd("p1.n.", is, sff, at, 0, f)
	ro.winType[WT_Special] = readFSBgTextSnd("p1.s.", is, sff, at, 0, f)
	ro.winType[WT_Hyper] = readFSBgTextSnd("p1.h.", is, sff, at, 0, f)
	ro.winType[WT_Cheese] = readFSBgTextSnd("p1.c.", is, sff, at, 0, f)
	ro.winType[WT_Time] = readFSBgTextSnd("p1.t.", is, sff, at, 0, f)
	ro.winType[WT_Throw] = readFSBgTextSnd("p1.throw.", is, sff, at, 0, f)
	ro.winType[WT_Suicide] = readFSBgTextSnd("p1.suicide.", is, sff, at, 0, f)
	ro.winType[WT_Teammate] = readFSBgTextSnd("p1.teammate.", is, sff, at, 0, f)
	ro.winType[WT_Perfect] = readFSBgTextSnd("p1.perfect.", is, sff, at, 0, f)
	ro.winType[WT_Clutch] = readFSBgTextSnd("p1.clutch.", is, sff, at, 0, f)
	ro.winType[WT_Normal+WT_NumTypes] = readFSBgTextSnd("p2.n.", is, sff, at, 0, f)
	ro.winType[WT_Special+WT_NumTypes] = readFSBgTextSnd("p2.s.", is, sff, at, 0, f)
	ro.winType[WT_Hyper+WT_NumTypes] = readFSBgTextSnd("p2.h.", is, sff, at, 0, f)
	ro.winType[WT_Cheese+WT_NumTypes] = readFSBgTextSnd("p2.c.", is, sff, at, 0, f)
	ro.winType[WT_Time+WT_NumTypes] = readFSBgTextSnd("p2.t.", is, sff, at, 0, f)
	ro.winType[WT_Throw+WT_NumTypes] = readFSBgTextSnd("p2.throw.", is, sff, at, 0, f)
	ro.winType[WT_Suicide+WT_NumTypes] = readFSBgTextSnd("p2.suicide.", is, sff, at, 0, f)
	ro.winType[WT_Teammate+WT_NumTypes] = readFSBgTextSnd("p2.teammate.", is, sff, at, 0, f)
	ro.winType[WT_Perfect+WT_NumTypes] = readFSBgTextSnd("p2.perfect.", is, sff, at, 0, f)
	ro.winType[WT_Clutch+WT_NumTypes] = readFSBgTextSnd("p2.clutch.", is, sff, at, 0, f)
	ro.fadeIn = readLbFade("fadein.", is, sff, at)
	ro.fadeOut = readLbFade("fadeout.", is, sff, at)
	is.ReadI32("shutter.time", &ro.shutter_time)
	is.ReadI32("clutch.threshold", &ro.clutch_threshold)
	var col [3]int32
	if is.ReadI32("shutter.col", &col[0], &col[1], &col[2]) {
		ro.shutter_col = uint32(col[0]&0xff<<16 | col[1]&0xff<<8 | col[2]&0xff)
	}
	is.ReadI32("callfight.time", &ro.callfight_time)
	return ro
}

func (ro *FightScreenRound) overTime() int32 {
	return Max(ro.over_time, ro.fadeOut.duration())
}

// Check is sys.intro timer should step
func (ro *FightScreenRound) act() bool {
	// Early exits
	if (sys.paused && !sys.frameStepFlag) || sys.gsf(GSF_roundfreeze) {
		return false
	}
	// Step fades
	if ro.fadeOut.isActive() {
		ro.fadeOut.step()
	} else if ro.fadeIn.isActive() {
		ro.fadeIn.step()
	}
	if ro.shutterTimer > 0 {
		// Signal system to skip intros when shutter is about to be fully closed
		// This ensures the intros will skip even if/when the shutter updates at a different rate than characters
		// https://github.com/ikemen-engine/Ikemen-GO/issues/2720
		if ro.shutterTimer == (ro.shutter_time + 1) {
			sys.introSkipCall = true
			ro.fadeIn.timeRemaining = 0
		}
		ro.shutterTimer--
	}

	// Pre-intro
	if sys.intro > ro.ctrl_time {
		// TODO: These resets are probably obsolete now
		ro.roundDisplayPhase = 0
		ro.fightDisplayPhase = 0
		ro.roundDisplayTimer = 0
		ro.fightDisplayTimer = 0
	} else if (sys.intro >= 0 && !sys.tickNextFrame()) || sys.motif.di.active || ro.shutterTimer > 0 {
		// Skip announcements during the middle of the round, "shuttertime" or dialogues
		// Mugen ignores the "shuttertime" here, but that makes the round/fight announcement too abrupt
		return false
	} else {
		// Intro
		if ro.roundDisplayPhase < 2 || ro.fightDisplayPhase < 2 {
			ro.handleRoundIntro()
		}
		// Outro
		if ro.fightDisplayPhase == 2 && sys.intro < 0 && (sys.finishType != FT_NotYet || sys.curRoundTime == 0) {
			ro.handleRoundOutro()
		} else {
			return ro.fightDisplayPhase > 0
		}
	}
	// Because round state should step before characters, this should be tickFrame()
	//return sys.tickNextFrame()
	return sys.tickFrame()
}

/*
// Check if current round animation can be skipped
// This prevents cutting an animation after it's already running
func (ro *FightScreenRound) canSkipPhase(phase int) bool {
	if phase < len(ro.waitTimer) && phase < len(ro.waitSoundTimer) && phase < len(ro.drawTimer) {
		return ro.waitTimer[phase] >= 0 && ro.waitSoundTimer[phase] >= 0 && ro.drawTimer[phase] <= 0
	}
	return false
}
*/

// Consists of round and fight calls
func (ro *FightScreenRound) handleRoundIntro() {
	// Previously skipping the char intros took us to the fight call, like Mugen
	// Most games go to the round call instead so this was changed
	//if sys.introSkipped && !sys.dialogueFlg {
	//	ro.roundDisplayPhase = 2
	//	ro.callFight()
	//	sys.introSkipped = false
	//}

	// Skip round call
	if sys.gsf(GSF_skiprounddisplay) {
		ro.roundDisplayPhase = 2
	}

	// Round call
	if ro.roundDisplayPhase < 2 {
		roundNum := sys.roundNo
		if sys.sel.gameParams.PersistRounds {
			roundNum = sys.persistRoundCount
		}
		// Sounds
		if ro.roundDisplayTimer == ro.round_sndtime {
			if sys.roundIsSingle() && ro.round_single.snd[0] != -1 {
				ro.snd.play(ro.round_single.snd, 100, 0, 0, 0, 0)
			} else if sys.roundIsFinal() && ro.round_final.snd[0] != -1 {
				ro.snd.play(ro.round_final.snd, 100, 0, 0, 0, 0)
			} else if int(roundNum) <= len(ro.round) && ro.round[roundNum-1].snd[0] != -1 {
				ro.snd.play(ro.round[roundNum-1].snd, 100, 0, 0, 0, 0)
			} else {
				ro.snd.play(ro.round_default.snd, 100, 0, 0, 0, 0)
			}
		}
		// Animations
		if ro.roundDisplayTimer >= ro.round_time {
			ro.roundDisplayPhase = 1
			animTime := ro.roundDisplayTimer - ro.round_time + 1 // Because animations start on frame 1 rather than 0
			if sys.roundIsSingle() && ro.round_single.snd[0] != -1 {
				if ro.round_single.End(animTime, true) && ro.round_default.End(animTime, true) {
					ro.roundDisplayPhase = 2
				} else {
					if len(ro.round_single_top.anim.frames) > 0 {
						ro.round_single_top.Action()
					} else {
						ro.round_default_top.Action()
					}
					ro.round_single.Action()
					ro.round_default.Action()
					if len(ro.round_single_bg[0].anim.frames) > 0 {
						for i := len(ro.round_single_bg) - 1; i >= 0; i-- {
							ro.round_single_bg[i].Action()
						}
					} else {
						for i := len(ro.round_default_bg) - 1; i >= 0; i-- {
							ro.round_default_bg[i].Action()
						}
					}
				}
			} else if sys.roundIsFinal() && ro.round_final.snd[0] != -1 {
				if ro.round_final.End(animTime, true) && ro.round_default.End(animTime, true) {
					ro.roundDisplayPhase = 2
				} else {
					if len(ro.round_final_top.anim.frames) > 0 {
						ro.round_final_top.Action()
					} else {
						ro.round_default_top.Action()
					}
					ro.round_final.Action()
					ro.round_default.Action()
					if len(ro.round_final_bg[0].anim.frames) > 0 {
						for i := len(ro.round_final_bg) - 1; i >= 0; i-- {
							ro.round_final_bg[i].Action()
						}
					} else {
						for i := len(ro.round_default_bg) - 1; i >= 0; i-- {
							ro.round_default_bg[i].Action()
						}
					}
				}
			} else if int(roundNum) <= len(ro.round) {
				if ro.round[roundNum-1].End(animTime, true) && ro.round_default.End(animTime, true) {
					ro.roundDisplayPhase = 2
				} else {
					ro.round_default_top.Action()
					ro.round[roundNum-1].Action()
					ro.round_default.Action()
					for i := len(ro.round_default_bg) - 1; i >= 0; i-- {
						ro.round_default_bg[i].Action()
					}
				}
			} else {
				if ro.round_default.End(animTime, true) {
					ro.roundDisplayPhase = 2
				} else {
					ro.round_default_top.Action()
					ro.round_default.Action()
					for i := len(ro.round_default_bg) - 1; i >= 0; i-- {
						ro.round_default_bg[i].Action()
					}
				}
			}
		}
		ro.roundDisplayTimer++
	}

	// Skip fight call
	// Cannot be skipped unless round call is finished or also skipped
	if ro.roundDisplayPhase == 2 && sys.gsf(GSF_skipfightdisplay) {
		ro.fightDisplayPhase = 2
		if sys.intro > 1 {
			sys.intro = 1 // Skip ctrl waiting time
		}
	}

	// Fight call
	if ro.fightDisplayPhase < 2 { // A bit redundant but matches the other screens
		if ro.fightDisplayPhase == 0 {
			// Handle delay between Round and Fight
			// Fight will start "callfight.time" frames after Round has started
			if ro.roundDisplayPhase > 0 {
				if ro.fightDisplayTimer >= ro.callfight_time {
					ro.fight.Reset()
					ro.fight_top.Reset()
					ro.fightDisplayPhase = 1
					ro.fightDisplayTimer = 0 // Reset timer so we can reuse it for the actual Fight display
					sys.timerCount = append(sys.timerCount, sys.matchTime)
					ro.timerActive = true
				} else {
					ro.fightDisplayTimer++
				}
			}
		} else if ro.fightDisplayPhase == 1 {
			// Active phase
			if ro.fightDisplayTimer == ro.fight_sndtime {
				ro.snd.play(ro.fight.snd, 100, 0, 0, 0, 0)
			}
			if ro.fightDisplayTimer >= ro.fight_time {
				ro.fight_top.Action()
				ro.fight.Action()
				for i := len(ro.fight_bg) - 1; i >= 0; i-- {
					ro.fight_bg[i].Action()
				}
				animTime := ro.fightDisplayTimer - ro.fight_time + 1
				if ro.fight.End(animTime, true) && ro.fightDisplayTimer >= ro.fight_sndtime {
					ro.fightDisplayPhase = 2
				}
			}
			ro.fightDisplayTimer++
		}
	}
}

// Consists of KO screen and winner messages
func (ro *FightScreenRound) handleRoundOutro() {
	if ro.timerActive {
		if sys.matchTime-sys.timerCount[sys.roundNo-1] > 0 {
			sys.timerCount[sys.roundNo-1] = sys.matchTime - sys.timerCount[sys.roundNo-1]
			sys.timerRounds = append(sys.timerRounds, sys.timeElapsed())
		} else {
			sys.timerCount[sys.roundNo-1] = 0
		}
		ro.timerActive = false
	}

	stepCallTimers := func(ats *AnimTextSnd, elapsed, time, sndtime, delay int32, name string) {
		// Play sound
		if elapsed == sndtime+delay {
			ro.snd.play(ats.snd, 100, 0, 0, 0, 0)
		}
		// Play animations
		if elapsed >= time+delay {
			animTime := elapsed - (time + delay) + 1

			if ats.End(animTime, false) {
				switch name { // TODO: We only need a name string here because KO and Win announcements have different data types
				case "ko":
					ro.koDisplayPhase = 2
				case "win":
					ro.winDisplayPhase = 2
				}
			} else {
				ats.Action()
				switch name {
				case "ko":
					ro.koDisplayPhase = 1
				case "win":
					ro.winDisplayPhase = 1
				}
			}
		}
	}

	// Skip KO screen
	if sys.gsf(GSF_skipkodisplay) {
		ro.koDisplayPhase = 2
	}

	// KO screen
	if ro.koDisplayPhase < 2 {
		var ats *AnimTextSnd
		var top *AnimLayout
		var bg *[32]AnimLayout
		var time, sndtime int32

		// Mugen deliberately delays these screens for some reason
		// It's not even related to the over.hittime parameter
		// This is probably always unwanted, but we'll limit it to "mugenversion" for now
		var delay int32
		if sys.fightScreen.ikemenver[0] == 0 && sys.fightScreen.ikemenver[1] == 0 {
			delay = 10
		}

		// Determine which finish screen and timers to use
		switch sys.finishType {
		case FT_DKO:
			ats, top, bg = &ro.dko, &ro.dko_top, &ro.dko_bg
			time, sndtime = ro.dko_time, ro.dko_sndtime
		case FT_TO, FT_TODraw:
			ats, top, bg = &ro.to, &ro.to_top, &ro.to_bg
			time, sndtime = ro.to_time, ro.to_sndtime
		default:
			ats, top, bg = &ro.ko, &ro.ko_top, &ro.ko_bg
			time, sndtime = ro.ko_time, ro.ko_sndtime
		}

		top.Action()
		stepCallTimers(ats, ro.koDisplayTimer, time, sndtime, delay, "ko")
		for i := len(bg) - 1; i >= 0; i-- {
			bg[i].Action()
		}
		ro.koDisplayTimer++
	}

	// Skip winner announcement
	if sys.gsf(GSF_skipwindisplay) {
		ro.winDisplayPhase = 2
	}

	// Winner announcement
	if ro.winDisplayPhase < 2 && sys.intro < -(ro.over_waittime) {
		wt := sys.winTeam
		if wt < 0 {
			wt = 0
		}
		lt := wt ^ 1
		if sys.finishType == FT_TODraw || (sys.finishType == FT_DKO && ro.dko_showdraw) {
			ro.drawgame_top.Action()
			stepCallTimers(&ro.drawgame, ro.winDisplayTimer, ro.win_time, ro.win_sndtime, 0, "win")
			for i := len(ro.drawgame_bg) - 1; i >= 0; i-- {
				ro.drawgame_bg[i].Action()
			}
		} else if sys.winTeam >= 0 {
			isPlayerWin, isAiWin := false, false
			for i := wt; i < MaxSimul*2; i += 2 {
				if len(sys.chars[i]) > 0 && sys.aiLevel[i] == 0 {
					isPlayerWin = true
					break
				}
			}
			for i := lt; i < MaxSimul*2; i += 2 {
				if len(sys.chars[i]) > 0 && sys.aiLevel[i] == 0 {
					isAiWin = true
					break
				}
			}
			var res *ResultAnnouncement
			activeTeam := wt

			idxL := sys.numSimul[lt] - 1
			if idxL < 0 {
				idxL = 0
			}

			idxW := sys.numSimul[wt] - 1
			if idxW < 0 {
				idxW = 0
			}
			if isAiWin && !isPlayerWin && ro.aiWinExists[lt][idxL] {
				// player lose vs ai opponent
				activeTeam = lt
				res = &ro.aiWin[idxL]
			} else if !isAiWin && isPlayerWin && ro.aiLoseExists[wt][idxW] {
				// player win vs ai opponent
				res = &ro.aiLose[idxW]
			} else {
				// default win
				res = &ro.win[idxW]
			}
			res.top[activeTeam].Action()
			for i := len(res.bg[activeTeam]) - 1; i >= 0; i-- {
				res.bg[activeTeam][i].Action()
			}
			stepCallTimers(&res.text[activeTeam], ro.winDisplayTimer, ro.win_time, ro.win_sndtime, 0, "win")
		}
		// Perfect and other special win types
		if sys.winTeam >= 0 {
			index := sys.winType[sys.winTeam]
			p2offset := WinType(0)
			if sys.winTeam == 1 {
				p2offset = WT_NumTypes
			}
			if index >= WT_CNormal {
				ro.winType[WT_Clutch+p2offset].step(ro.snd)
				index -= (WT_CNormal - WT_Normal)
			} else if index >= WT_PNormal {
				ro.winType[WT_Perfect+p2offset].step(ro.snd)
				index -= (WT_PNormal - WT_Normal)
			}
			ro.winType[index+p2offset].step(ro.snd)
		}
		ro.winDisplayTimer++
	}
}

func (ro *FightScreenRound) reset() {
	ro.shutterTimer = 0
	ro.fadeOut.reset()
	ro.fadeIn.init(ro.fadeIn, true)

	resetElement := func(main *AnimTextSnd, top *AnimLayout, bgs []AnimLayout) {
		main.Reset()
		top.Reset()
		for i := range bgs {
			bgs[i].Reset()
		}
	}
	resetResults := func(res *[4]ResultAnnouncement) {
		for i := range res {
			for j := 0; j < 2; j++ {
				res[i].text[j].Reset()
				res[i].top[j].Reset()
				for k := range res[i].bg[j] {
					res[i].bg[j][k].Reset()
				}
			}
		}
	}
	// Round animations
	resetElement(&ro.round_default, &ro.round_default_top, ro.round_default_bg[:])
	for i := range ro.round {
		ro.round[i].Reset()
	}
	// Single round animations
	resetElement(&ro.round_single, &ro.round_single_top, ro.round_single_bg[:])
	// Final round animations
	resetElement(&ro.round_final, &ro.round_final_top, ro.round_final_bg[:])
	// Fight call animations
	resetElement(&ro.fight, &ro.fight_top, ro.fight_bg[:])
	// KO animations
	resetElement(&ro.ko, &ro.ko_top, ro.ko_bg[:])
	resetElement(&ro.dko, &ro.dko_top, ro.dko_bg[:])
	// Time Over animations
	resetElement(&ro.to, &ro.to_top, ro.to_bg[:])
	// Draw game
	resetElement(&ro.drawgame, &ro.drawgame_top, ro.drawgame_bg[:])

	// Winner Annoucement
	resetResults(&ro.win)
	// Player Win vs AI Annoucement
	resetResults(&ro.aiLose)
	// Player Lose vs AI Annoucement
	resetResults(&ro.aiWin)
	// Win types
	for i := range ro.winType {
		ro.winType[i].reset()
	}

	// Reset action timers
	ro.roundDisplayTimer = 0
	ro.fightDisplayTimer = 0
	ro.koDisplayTimer = 0
	ro.winDisplayTimer = 0
	ro.roundDisplayPhase = 0
	ro.fightDisplayPhase = 0
	ro.koDisplayPhase = 0
	ro.winDisplayPhase = 0
}

func (ro *FightScreenRound) draw(layerno int16, f map[int]*Fnt) {
	// Temporarily override system brightness
	oldBright := sys.brightness
	sys.brightness = 1.0

	// Round call animations
	if ro.roundDisplayPhase == 1 {
		// Draw default round background
		for i := range ro.round_default_bg {
			ro.round_default_bg[i].Draw(
				float32(ro.pos[0])+sys.fightScreen.offsetX,
				float32(ro.pos[1]),
				layerno,
				sys.fightScreen.scale,
			)
		}

		// Check round number
		var round_ref AnimTextSnd
		roundNum := sys.roundNo
		if sys.sel.gameParams.PersistRounds {
			roundNum = sys.persistRoundCount
		}

		// Draw background and select round reference
		switch {
		case sys.roundIsSingle() &&
			(ro.round_single.HasDrawable() || len(ro.round_single_bg) > 0 && ro.round_single_bg[0].HasAnim()):
			// Single round
			for i := range ro.round_single_bg {
				ro.round_single_bg[i].Draw(
					float32(ro.pos[0])+sys.fightScreen.offsetX,
					float32(ro.pos[1]),
					layerno,
					sys.fightScreen.scale,
				)
			}
			round_ref = ro.round_single
		case sys.roundIsFinal() &&
			(ro.round_final.HasDrawable() || len(ro.round_final_bg) > 0 && ro.round_final_bg[0].HasAnim()):
			// Final round
			for i := range ro.round_final_bg {
				ro.round_final_bg[i].Draw(
					float32(ro.pos[0])+sys.fightScreen.offsetX,
					float32(ro.pos[1]),
					layerno,
					sys.fightScreen.scale,
				)
			}
			round_ref = ro.round_final
		case int(roundNum) <= len(ro.round) && ro.round[roundNum-1].HasDrawable():
			// Normal round
			round_ref = ro.round[roundNum-1]
		}

		// Backup default text
		tmp := ro.round_default.text.text

		// If round_ref text is empty, format the default round text
		if round_ref.text.text == "" {
			ro.round_default.text.text = OldSprintf(tmp, roundNum)
		} else {
			ro.round_default.text.text = ""
		}

		// Draw default round
		ro.round_default.Draw(float32(ro.pos[0])+sys.fightScreen.offsetX, float32(ro.pos[1]), layerno, f, sys.fightScreen.scale)

		// Restore default text
		ro.round_default.text.text = tmp

		// Backup round_ref text
		tmp = round_ref.text.text

		// Format the round_ref text with the round number
		round_ref.text.text = OldSprintf(tmp, roundNum)

		// Draw round-specific elements
		round_ref.Draw(float32(ro.pos[0])+sys.fightScreen.offsetX, float32(ro.pos[1]), layerno, f, sys.fightScreen.scale)

		// Restore round_ref text
		round_ref.text.text = tmp

		// Draw top layer
		switch {
		case sys.roundIsSingle() && ro.round_single_top.anim != nil && len(ro.round_single_top.anim.frames) > 0:
			// Single round
			ro.round_single_top.Draw(
				float32(ro.pos[0])+sys.fightScreen.offsetX,
				float32(ro.pos[1]),
				layerno,
				sys.fightScreen.scale,
			)
		case sys.roundIsFinal() && ro.round_final_top.anim != nil && len(ro.round_final_top.anim.frames) > 0:
			// Final round
			ro.round_final_top.Draw(
				float32(ro.pos[0])+sys.fightScreen.offsetX,
				float32(ro.pos[1]),
				layerno,
				sys.fightScreen.scale,
			)
		default:
			// Normal round
			ro.round_default_top.Draw(
				float32(ro.pos[0])+sys.fightScreen.offsetX,
				float32(ro.pos[1]),
				layerno,
				sys.fightScreen.scale,
			)
		}
	}

	// "Fight!" animations
	if ro.fightDisplayPhase == 1 {
		for i := range ro.fight_bg {
			ro.fight_bg[i].Draw(float32(ro.pos[0])+sys.fightScreen.offsetX, float32(ro.pos[1]), layerno, sys.fightScreen.scale)
		}
		ro.fight.Draw(float32(ro.pos[0])+sys.fightScreen.offsetX, float32(ro.pos[1]), layerno, f, sys.fightScreen.scale)
		ro.fight_top.Draw(float32(ro.pos[0])+sys.fightScreen.offsetX, float32(ro.pos[1]), layerno, sys.fightScreen.scale)
	}

	// KO animations
	if ro.koDisplayPhase == 1 {
		var ats *AnimTextSnd
		var top *AnimLayout
		var bg *[32]AnimLayout
		var time int32

		var delay int32
		if sys.fightScreen.ikemenver[0] == 0 && sys.fightScreen.ikemenver[1] == 0 {
			delay = 10
		}

		switch sys.finishType {
		case FT_DKO:
			ats, top, bg = &ro.dko, &ro.dko_top, &ro.dko_bg
			time = ro.dko_time
		case FT_TO, FT_TODraw:
			ats, top, bg = &ro.to, &ro.to_top, &ro.to_bg
			time = ro.to_time
		default:
			ats, top, bg = &ro.ko, &ro.ko_top, &ro.ko_bg
			time = ro.ko_time
		}

		if ro.koDisplayTimer >= time+delay {
			for i := range bg {
				bg[i].Draw(float32(ro.pos[0])+sys.fightScreen.offsetX, float32(ro.pos[1]), layerno, sys.fightScreen.scale)
			}
			ats.Draw(float32(ro.pos[0])+sys.fightScreen.offsetX, float32(ro.pos[1]), layerno, f, sys.fightScreen.scale)
			top.Draw(float32(ro.pos[0])+sys.fightScreen.offsetX, float32(ro.pos[1]), layerno, sys.fightScreen.scale)
		}
	}

	// Winner announcement
	if ro.winDisplayPhase == 1 {
		wt := sys.winTeam
		if wt < 0 {
			wt = 0
		}
		lt := wt ^ 1
		if sys.finishType == FT_TODraw || (sys.finishType == FT_DKO && ro.dko_showdraw) {
			for i := range ro.drawgame_bg {
				ro.drawgame_bg[i].Draw(float32(ro.pos[0])+sys.fightScreen.offsetX, float32(ro.pos[1]), layerno, sys.fightScreen.scale)
			}
			ro.drawgame.Draw(float32(ro.pos[0])+sys.fightScreen.offsetX, float32(ro.pos[1]), layerno, f, sys.fightScreen.scale)
			ro.drawgame_top.Draw(float32(ro.pos[0])+sys.fightScreen.offsetX, float32(ro.pos[1]), layerno, sys.fightScreen.scale)
		} else if sys.winTeam >= 0 {
			isPlayerWin, isAiWin := false, false
			for i := wt; i < MaxSimul*2; i += 2 {
				if len(sys.chars[i]) > 0 && sys.aiLevel[i] == 0 {
					isPlayerWin = true
					break
				}
			}
			for i := lt; i < MaxSimul*2; i += 2 {
				if len(sys.chars[i]) > 0 && sys.aiLevel[i] == 0 {
					isAiWin = true
					break
				}
			}
			var res *ResultAnnouncement
			activeTeam := wt

			idxL := sys.numSimul[lt] - 1
			if idxL < 0 {
				idxL = 0
			}

			idxW := sys.numSimul[wt] - 1
			if idxW < 0 {
				idxW = 0
			}

			if isAiWin && !isPlayerWin && ro.aiWinExists[lt][idxL] {
				// player lose vs ai opponent
				activeTeam = lt
				res = &ro.aiWin[idxL]
			} else if !isAiWin && isPlayerWin && ro.aiLoseExists[wt][idxW] {
				// player win vs ai opponent
				res = &ro.aiLose[idxW]
			} else {
				// default win
				res = &ro.win[idxW]
			}
			// BGs
			for i := range res.bg[activeTeam] {
				res.bg[activeTeam][i].Draw(float32(ro.pos[0])+sys.fightScreen.offsetX, float32(ro.pos[1]), layerno, sys.fightScreen.scale)
			}
			// Text
			oldTxt := res.text[activeTeam].text.text
			var args []interface{}
			for i := activeTeam; i < len(sys.chars); i += 2 {
				if len(sys.chars[i]) > 0 {
					if strings.Contains(oldTxt, "%d") {
						args = append(args, i+1) // playerNum
					}
					if strings.Contains(oldTxt, "%s") {
						args = append(args, sys.cgi[i].displayname)
					}
				}
			}
			if len(args) > 0 {
				res.text[activeTeam].text.text = OldSprintf(oldTxt, args...)
			}
			res.text[activeTeam].Draw(float32(ro.pos[0])+sys.fightScreen.offsetX, float32(ro.pos[1]), layerno, f, sys.fightScreen.scale)
			res.text[activeTeam].text.text = oldTxt
			// Top
			res.top[activeTeam].Draw(float32(ro.pos[0])+sys.fightScreen.offsetX, float32(ro.pos[1]), layerno, sys.fightScreen.scale)
		}
		// Perfect and other special win types
		if sys.winTeam >= 0 {
			index := sys.winType[sys.winTeam]
			clutch, perfect := false, false

			if index >= WT_CNormal {
				clutch = true
				index -= (WT_CNormal - WT_Normal)
			} else if index >= WT_PNormal {
				perfect = true
				index -= (WT_PNormal - WT_Normal)
			}
			p2offset := WinType(0)
			if sys.winTeam == 1 {
				p2offset = WT_NumTypes
			}
			if clutch {
				ro.winType[WT_Clutch+p2offset].bgDraw(layerno)
				ro.winType[WT_Clutch+p2offset].draw(layerno, f)
			} else if perfect {
				ro.winType[WT_Perfect+p2offset].bgDraw(layerno)
				ro.winType[WT_Perfect+p2offset].draw(layerno, f)
			}
			ro.winType[index+p2offset].bgDraw(layerno)
			ro.winType[index+p2offset].draw(layerno, f)
		}
	}

	// Restore system brightness
	sys.brightness = oldBright
}

func (ro *FightScreenRound) drawFade() {
	if ro.fadeOut.isActive() {
		ro.fadeOut.draw()
	} else if ro.fadeIn.isActive() {
		ro.fadeIn.draw()
	} else if sys.clsnDisplay && sys.cfg.Debug.ClsnDarken {
		ro.fadeIn.drawRect(sys.scrrect, 0, 0)
	}

	if ro.shutterTimer > 0 {
		// shutterTimer is a countdown from (shutter_time*2) to 0:
		// 2T -> open, T -> fully closed, 0 -> open
		rect := sys.scrrect
		half := (sys.scrrect[3] + 1) >> 1
		var cover int32
		if ro.shutter_time > 0 {
			if ro.shutterTimer > ro.shutter_time {
				// Closing phase
				t := ro.shutter_time*2 - ro.shutterTimer
				if t < 0 {
					t = 0
				}
				if t > ro.shutter_time {
					t = ro.shutter_time
				}
				cover = t * half / ro.shutter_time
			} else {
				// Opening phase
				t := ro.shutterTimer
				if t < 0 {
					t = 0
				}
				if t > ro.shutter_time {
					t = ro.shutter_time
				}
				cover = t * half / ro.shutter_time
			}
		}
		rect[3] = cover
		ro.fadeIn.drawRect(rect, ro.shutter_col, 255)
		rect[1] = sys.scrrect[3] - rect[3]
		ro.fadeIn.drawRect(rect, ro.shutter_col, 255)
	}
}

type FightScreenTimer struct {
	pos     [2]int32
	text    FSText
	bg      AnimLayout
	top     AnimLayout
	enabled map[string]bool
	active  bool
}

func newFightScreenTimer() *FightScreenTimer {
	return &FightScreenTimer{enabled: make(map[string]bool)}
}

func readFightScreenTimer(is IniSection, sff *Sff, at AnimationTable, f map[int]*Fnt) *FightScreenTimer {
	tr := newFightScreenTimer()

	is.ReadI32("pos", &tr.pos[0], &tr.pos[1])
	tr.text = *readFSText("text.", is, "", 0, f, 0)
	tr.bg = ReadAnimLayout("bg.", is, sff, at, 0)
	tr.top = ReadAnimLayout("top.", is, sff, at, 0)
	for k := range is {
		sp := strings.Split(k, ".")
		if len(sp) == 2 && sp[0] == "enabled" {
			var b bool
			if is.ReadBool(k, &b) {
				tr.enabled[sp[1]] = b
			}
		}
	}
	return tr
}

func (tr *FightScreenTimer) step() {
	tr.bg.Action()
	tr.top.Action()
	tr.text.step()
}

func (tr *FightScreenTimer) reset() {
	tr.bg.Reset()
	tr.top.Reset()
}

func (tr *FightScreenTimer) bgDraw(layerno int16) {
	if tr.active {
		tr.bg.Draw(float32(tr.pos[0])+sys.fightScreen.offsetX, float32(tr.pos[1]), layerno, sys.fightScreen.scale)
	}
}

func (tr *FightScreenTimer) draw(layerno int16, f map[int]*Fnt) {
	if tr.active && sys.curFramesPerCount > 0 &&
		tr.text.font[0] >= 0 && getFont(f, tr.text.font[0]) != nil {
		text := tr.text.text
		totalSec := float64(sys.timeTotal()) / 60
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
		tr.text.lay.DrawText(float32(tr.pos[0])+sys.fightScreen.offsetX, float32(tr.pos[1]), sys.fightScreen.scale, layerno,
			text, getFont(f, tr.text.font[0]), tr.text.font[1], tr.text.font[2], tr.text.palfx, tr.text.frgba)
		tr.top.Draw(float32(tr.pos[0])+sys.fightScreen.offsetX, float32(tr.pos[1]), layerno, sys.fightScreen.scale)
	}
}

type FightScreenScore struct {
	pos       [2]int32
	text      FSText
	bg        AnimLayout
	top       AnimLayout
	separator [2]string
	pad       int32
	places    int32
	min       float32
	max       float32
	enabled   map[string]bool
	active    bool
}

func newFightScreenScore() *FightScreenScore {
	return &FightScreenScore{separator: [2]string{"", "."}, enabled: make(map[string]bool)}
}

func readFightScreenScore(pre string, is IniSection, sff *Sff, at AnimationTable, f map[int]*Fnt) *FightScreenScore {
	sc := newFightScreenScore()

	is.ReadI32(pre+"pos", &sc.pos[0], &sc.pos[1])
	sc.text = *readFSText(pre+"text.", is, "", 0, f, 0)
	sc.separator[0], _, _ = is.getText("format.integer.separator")
	sc.separator[1], _, _ = is.getText("format.decimal.separator")
	is.ReadI32("format.integer.pad", &sc.pad)
	is.ReadI32("format.decimal.places", &sc.places)
	is.ReadF32("score.min", &sc.min)
	is.ReadF32("score.max", &sc.max)
	sc.bg = ReadAnimLayout(pre+"bg.", is, sff, at, 0)
	sc.top = ReadAnimLayout(pre+"top.", is, sff, at, 0)
	for k := range is {
		sp := strings.Split(k, ".")
		if len(sp) == 3 && pre == fmt.Sprintf("%v.", sp[0]) && sp[1] == "enabled" {
			var b bool
			if is.ReadBool(k, &b) {
				sc.enabled[sp[2]] = b
			}
		}
	}
	return sc
}

func (sc *FightScreenScore) step() {
	sc.bg.Action()
	sc.top.Action()
	sc.text.step()
}

func (sc *FightScreenScore) reset() {
	sc.bg.Reset()
	sc.top.Reset()
}

func (sc *FightScreenScore) bgDraw(layerno int16) {
	if sc.active {
		sc.bg.Draw(float32(sc.pos[0])+sys.fightScreen.offsetX, float32(sc.pos[1]), layerno, sys.fightScreen.scale)
	}
}

func (sc *FightScreenScore) draw(layerno int16, f map[int]*Fnt, side int) {
	if sc.active && sc.text.font[0] >= 0 && getFont(f, sc.text.font[0]) != nil {
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
		// split float value
		s := strings.Split(fmt.Sprintf("%f", total), ".")
		// integer left padding (add leading zeros)
		for i := int(sc.pad) - len(s[0]); i > 0; i-- {
			s[0] = "0" + s[0]
		}
		// integer thousands separator
		for i := len(s[0]) - 3; i > 0; i -= 3 {
			s[0] = s[0][:i] + sc.separator[0] + s[0][i:]
		}
		// decimal places (trim trailing numbers)
		if int(sc.places) < len(s[1]) {
			s[1] = s[1][:sc.places]
		}
		// decimal separator
		ds := ""
		if sc.places > 0 {
			ds = sc.separator[1]
		}
		// replace %s with formatted string
		text = strings.Replace(text, "%s", s[0]+ds+s[1], 1)
		sc.text.lay.DrawText(float32(sc.pos[0])+sys.fightScreen.offsetX, float32(sc.pos[1]), sys.fightScreen.scale, layerno,
			text, getFont(f, sc.text.font[0]), sc.text.font[1], sc.text.font[2], sc.text.palfx, sc.text.frgba)
		sc.top.Draw(float32(sc.pos[0])+sys.fightScreen.offsetX, float32(sc.pos[1]), layerno, sys.fightScreen.scale)
	}
}

type FightScreenMatch struct {
	pos     [2]int32
	text    FSText
	bg      AnimLayout
	top     AnimLayout
	enabled map[string]bool
	active  bool
}

func newFightScreenMatch() *FightScreenMatch {
	return &FightScreenMatch{enabled: make(map[string]bool)}
}

func readFightScreenMatch(is IniSection, sff *Sff, at AnimationTable, f map[int]*Fnt) *FightScreenMatch {
	ma := newFightScreenMatch()

	is.ReadI32("pos", &ma.pos[0], &ma.pos[1])
	ma.text = *readFSText("text.", is, "", 0, f, 0)
	ma.bg = ReadAnimLayout("bg.", is, sff, at, 0)
	ma.top = ReadAnimLayout("top.", is, sff, at, 0)
	for k := range is {
		sp := strings.Split(k, ".")
		if len(sp) == 2 && sp[0] == "enabled" {
			var b bool
			if is.ReadBool(k, &b) {
				ma.enabled[sp[1]] = b
			}
		}
	}
	return ma
}

func (ma *FightScreenMatch) step() {
	ma.bg.Action()
	ma.top.Action()
	ma.text.step()
}

func (ma *FightScreenMatch) reset() {
	ma.bg.Reset()
	ma.top.Reset()
}

func (ma *FightScreenMatch) bgDraw(layerno int16) {
	if ma.active {
		ma.bg.Draw(float32(ma.pos[0])+sys.fightScreen.offsetX, float32(ma.pos[1]), layerno, sys.fightScreen.scale)
	}
}

func (ma *FightScreenMatch) draw(layerno int16, f map[int]*Fnt) {
	if ma.active && ma.text.font[0] >= 0 && getFont(f, ma.text.font[0]) != nil {
		text := ma.text.text
		text = strings.Replace(text, "%s", fmt.Sprintf("%v", sys.matchNo), 1)
		ma.text.lay.DrawText(float32(ma.pos[0])+sys.fightScreen.offsetX, float32(ma.pos[1]), sys.fightScreen.scale, layerno,
			text, getFont(f, ma.text.font[0]), ma.text.font[1], ma.text.font[2], ma.text.palfx, ma.text.frgba)
		ma.top.Draw(float32(ma.pos[0])+sys.fightScreen.offsetX, float32(ma.pos[1]), layerno, sys.fightScreen.scale)
	}
}

type FightScreenAiLevel struct {
	pos       [2]int32
	text      FSText
	bg        AnimLayout
	top       AnimLayout
	separator string
	places    int32
	enabled   map[string]bool
	active    bool
}

func newFightScreenAiLevel() *FightScreenAiLevel {
	return &FightScreenAiLevel{separator: ".", enabled: make(map[string]bool)}
}

func readFightScreenAiLevel(pre string, is IniSection, sff *Sff, at AnimationTable, f map[int]*Fnt) *FightScreenAiLevel {
	ai := newFightScreenAiLevel()

	is.ReadI32(pre+"pos", &ai.pos[0], &ai.pos[1])
	ai.text = *readFSText(pre+"text.", is, "", 0, f, 0)
	ai.separator, _, _ = is.getText("format.decimal.separator")
	is.ReadI32("format.decimal.places", &ai.places)
	ai.bg = ReadAnimLayout(pre+"bg.", is, sff, at, 0)
	ai.top = ReadAnimLayout(pre+"top.", is, sff, at, 0)
	for k := range is {
		sp := strings.Split(k, ".")
		if len(sp) == 3 && pre == fmt.Sprintf("%v.", sp[0]) && sp[1] == "enabled" {
			var b bool
			if is.ReadBool(k, &b) {
				ai.enabled[sp[2]] = b
			}
		}
	}
	return ai
}

func (ai *FightScreenAiLevel) step() {
	ai.bg.Action()
	ai.top.Action()
	ai.text.step()
}

func (ai *FightScreenAiLevel) reset() {
	ai.bg.Reset()
	ai.top.Reset()
}

func (ai *FightScreenAiLevel) bgDraw(layerno int16) {
	if ai.active {
		ai.bg.Draw(float32(ai.pos[0])+sys.fightScreen.offsetX, float32(ai.pos[1]), layerno, sys.fightScreen.scale)
	}
}

func (ai *FightScreenAiLevel) draw(layerno int16, f map[int]*Fnt, ailv float32) {
	if ai.active && ailv > 0 && ai.text.font[0] >= 0 && getFont(f, ai.text.font[0]) != nil {
		text := ai.text.text
		// split float value
		s := strings.Split(fmt.Sprintf("%f", ailv), ".")
		// decimal places (trim trailing numbers)
		if int(ai.places) < len(s[1]) {
			s[1] = s[1][:ai.places]
		}
		// decimal separator
		ds := ""
		if ai.places > 0 {
			ds = ai.separator
		}
		// replace %s with formatted string
		text = strings.Replace(text, "%s", s[0]+ds+s[1], 1)
		// percentage value
		p := ailv / 8 * 100
		text = strings.Replace(text, "%p", fmt.Sprintf("%.0f", p), 1)
		ai.text.lay.DrawText(float32(ai.pos[0])+sys.fightScreen.offsetX, float32(ai.pos[1]), sys.fightScreen.scale, layerno,
			text, getFont(f, ai.text.font[0]), ai.text.font[1], ai.text.font[2], ai.text.palfx, ai.text.frgba)
		ai.top.Draw(float32(ai.pos[0])+sys.fightScreen.offsetX, float32(ai.pos[1]), layerno, sys.fightScreen.scale)
	}
}

type FightScreenWinCount struct {
	pos     [2]int32
	text    FSText
	bg      AnimLayout
	top     AnimLayout
	wins    int32
	enabled map[string]bool
	active  bool
}

func newFightScreenWinCount() *FightScreenWinCount {
	return &FightScreenWinCount{enabled: make(map[string]bool)}
}

func readFightScreenWinCount(pre string, is IniSection, sff *Sff, at AnimationTable, f map[int]*Fnt) *FightScreenWinCount {
	wc := newFightScreenWinCount()

	is.ReadI32(pre+"pos", &wc.pos[0], &wc.pos[1])
	wc.text = *readFSText(pre+"text.", is, "", 0, f, 0)
	wc.bg = ReadAnimLayout(pre+"bg.", is, sff, at, 0)
	wc.top = ReadAnimLayout(pre+"top.", is, sff, at, 0)
	for k := range is {
		sp := strings.Split(k, ".")
		if len(sp) == 3 && pre == fmt.Sprintf("%v.", sp[0]) && sp[1] == "enabled" {
			var b bool
			if is.ReadBool(k, &b) {
				wc.enabled[sp[2]] = b
			}
		}
	}
	return wc
}

func (wc *FightScreenWinCount) step() {
	wc.bg.Action()
	wc.top.Action()
	wc.text.step()
}

func (wc *FightScreenWinCount) reset() {
	wc.bg.Reset()
	wc.top.Reset()
}

func (wc *FightScreenWinCount) bgDraw(layerno int16) {
	if wc.active {
		wc.bg.Draw(float32(wc.pos[0])+sys.fightScreen.offsetX, float32(wc.pos[1]), layerno, sys.fightScreen.scale)
	}
}

func (wc *FightScreenWinCount) draw(layerno int16, f map[int]*Fnt, side int) {
	if wc.active && wc.text.font[0] >= 0 && getFont(f, wc.text.font[0]) != nil {
		text := wc.text.text
		text = strings.Replace(text, "%s", fmt.Sprintf("%v", wc.wins), 1)
		wc.text.lay.DrawText(float32(wc.pos[0])+sys.fightScreen.offsetX, float32(wc.pos[1]), sys.fightScreen.scale, layerno,
			text, getFont(f, wc.text.font[0]), wc.text.font[1], wc.text.font[2], wc.text.palfx, wc.text.frgba)
		wc.top.Draw(float32(wc.pos[0])+sys.fightScreen.offsetX, float32(wc.pos[1]), layerno, sys.fightScreen.scale)
	}
}

type FightScreenMode struct {
	pos  [2]int32
	text FSText
	bg   AnimLayout
	top  AnimLayout
}

func newFightScreenMode() *FightScreenMode {
	return &FightScreenMode{}
}

func readFightScreenMode(is IniSection,
	sff *Sff, at AnimationTable, f map[int]*Fnt) map[string]*FightScreenMode {
	mo := make(map[string]*FightScreenMode)
	for k := range is {
		sp := strings.Split(k, ".")
		if _, ok := mo[sp[0]]; !ok {
			mo[sp[0]] = newFightScreenMode()
			is.ReadI32(sp[0]+".pos", &mo[sp[0]].pos[0], &mo[sp[0]].pos[1])
			mo[sp[0]].text = *readFSText(sp[0]+".text.", is, "", 0, f, 0)
			mo[sp[0]].bg = ReadAnimLayout(sp[0]+".bg.", is, sff, at, 0)
			mo[sp[0]].top = ReadAnimLayout(sp[0]+".top.", is, sff, at, 0)
		}
	}
	return mo
}

func (mo *FightScreenMode) step() {
	mo.bg.Action()
	mo.top.Action()
	mo.text.step()
}

func (mo *FightScreenMode) reset() {
	mo.bg.Reset()
	mo.top.Reset()
}

func (mo *FightScreenMode) bgDraw(layerno int16) {
	if sys.fightScreen.mode {
		mo.bg.Draw(float32(mo.pos[0])+sys.fightScreen.offsetX, float32(mo.pos[1]), layerno, sys.fightScreen.scale)
	}
}

func (mo *FightScreenMode) draw(layerno int16, f map[int]*Fnt) {
	if sys.fightScreen.mode && mo.text.font[0] >= 0 && getFont(f, mo.text.font[0]) != nil {
		mo.text.lay.DrawText(float32(mo.pos[0])+sys.fightScreen.offsetX, float32(mo.pos[1]), sys.fightScreen.scale, layerno,
			mo.text.text, getFont(f, mo.text.font[0]), mo.text.font[1], mo.text.font[2], mo.text.palfx, mo.text.frgba)
		mo.top.Draw(float32(mo.pos[0])+sys.fightScreen.offsetX, float32(mo.pos[1]), layerno, sys.fightScreen.scale)
	}
}

type FightScreen struct {
	def           string
	name          string
	nameLow       string
	author        string
	authorLow     string
	localcoord    [2]int32
	ikemenver     [3]uint16
	ikemenverF    float32
	mugenver      [2]uint16
	mugenverF     float32
	offsetX       float32
	offsetY       float32
	scale         float32
	portraitScale float32
	animTable     AnimationTable
	sff           *Sff
	snd           *Snd
	fnt           map[int]*Fnt
	curLayout     [2]int   // Which layout each team is currently using. Single, Simul 3P, etc
	teamOrder     [2][]int // Player order on each team. Because it can vary in Tag
	lifeBars      [8][]*LifeBar
	powerBars     [8][]*PowerBar
	guardBars     [8][]*GuardBar
	stunBars      [8][]*StunBar
	faces         [8][]*FightScreenFace
	names         [8][]*FightScreenName
	winIcons      [2]*FightScreenWinIcon
	time          *FightScreenTime // The normal round time
	combos        [2]*FightScreenCombo
	actions       [2]*FightScreenAction
	round         *FightScreenRound
	timer         *FightScreenTimer // For Time Attack and such
	scores        [2]*FightScreenScore
	match         *FightScreenMatch
	aiLevels      [2]*FightScreenAiLevel
	winCounts     [2]*FightScreenWinCount
	modes         map[string]*FightScreenMode
	missing       map[string]int
	active        bool
	bars          bool
	mode          bool
	redlifebar    bool
	guardbar      bool
	stunbar       bool
	hidebars      bool
	fnt_scale     float32
	fx_limit      int
}

func loadFightScreen(def string) (*FightScreen, error) {
	if def == "" {
		sys.fightScreen.resolvePath()
		def = sys.fightScreen.def
	}
	str, err := LoadText(def)
	if err != nil {
		return nil, err
	}
	fs := &FightScreen{
		localcoord:    [2]int32{320, 240},
		scale:         1,
		portraitScale: 1,
		sff:           &Sff{},
		snd:           &Snd{},
		lifeBars: [...][]*LifeBar{make([]*LifeBar, 2), make([]*LifeBar, 8),
			make([]*LifeBar, 2), make([]*LifeBar, 8), make([]*LifeBar, 6),
			make([]*LifeBar, 8), make([]*LifeBar, 6), make([]*LifeBar, 8)},
		powerBars: [...][]*PowerBar{make([]*PowerBar, 2), make([]*PowerBar, 8),
			make([]*PowerBar, 2), make([]*PowerBar, 8), make([]*PowerBar, 6),
			make([]*PowerBar, 8), make([]*PowerBar, 6), make([]*PowerBar, 8)},
		guardBars: [...][]*GuardBar{make([]*GuardBar, 2), make([]*GuardBar, 8),
			make([]*GuardBar, 2), make([]*GuardBar, 8), make([]*GuardBar, 6),
			make([]*GuardBar, 8), make([]*GuardBar, 6), make([]*GuardBar, 8)},
		stunBars: [...][]*StunBar{make([]*StunBar, 2), make([]*StunBar, 8),
			make([]*StunBar, 2), make([]*StunBar, 8), make([]*StunBar, 6),
			make([]*StunBar, 8), make([]*StunBar, 6), make([]*StunBar, 8)},
		faces: [...][]*FightScreenFace{make([]*FightScreenFace, 2), make([]*FightScreenFace, 8),
			make([]*FightScreenFace, 2), make([]*FightScreenFace, 8), make([]*FightScreenFace, 6),
			make([]*FightScreenFace, 8), make([]*FightScreenFace, 6), make([]*FightScreenFace, 8)},
		names: [...][]*FightScreenName{make([]*FightScreenName, 2), make([]*FightScreenName, 8),
			make([]*FightScreenName, 2), make([]*FightScreenName, 8), make([]*FightScreenName, 6),
			make([]*FightScreenName, 8), make([]*FightScreenName, 6), make([]*FightScreenName, 8)},
		active:    true,
		bars:      true,
		mode:      true,
		fnt_scale: 1,
		fx_limit:  3,
	}
	fs.fnt = make(map[int]*Fnt)
	fs.missing = map[string]int{
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
		"[tag_3p name]": 6, "[tag_4p name]": 7, "[action]": -1, "[timer]": -1,
		"[score]": -1, "[match]": -1, "[ailevel]": -1, "[wincount]": -1,
		"[mode]": -1,
	}
	strc := strings.ToLower(strings.TrimSpace(str))
	for k := range fs.missing {
		strc = strings.Replace(strc, ";"+k, "", -1)
		if strings.Contains(strc, k) {
			delete(fs.missing, k)
		} else {
			str += "\n" + k
		}
	}

	lines, lnidx := SplitAndTrim(str, "\n"), 0
	fs.animTable = ReadAnimationTable(def, fs.sff, &fs.sff.palList, lines, &lnidx, true)
	lnidx = 0
	filesflg := true

	// Pre-scan [info] to initialize fight screen localcoord/scale before any FightFX load
	pre := 0
	for pre < len(lines) {
		// If fight screen doesn't have localcoord assigned we're using screenpack localcoord
		fs.localcoord[0], fs.localcoord[1] = sys.motif.Info.Localcoord[0], sys.motif.Info.Localcoord[1]
		is, name, _ := ReadIniSection(lines, &pre)
		if name == "info" {
			is.ReadI32("localcoord", &fs.localcoord[0], &fs.localcoord[1])
			break
		}
	}
	fs.setScale()
	// Seed sys.fightScreen with values used during loading (before the final assignment)
	sys.fightScreen.localcoord = fs.localcoord
	sys.fightScreen.offsetX = fs.offsetX
	sys.fightScreen.scale = fs.scale
	sys.fightScreen.portraitScale = fs.portraitScale

	// Prepare fight screen FightFX
	ffx := newFightFx()

	for lnidx < len(lines) {
		is, name, subname := ReadIniSection(lines, &lnidx)

		// Helper to return only the parameters for a given player number
		resolvedSection := func(prefixBase string, n int) IniSection {
			// Legacy versions cannot use prefix operators ('*' and '|')
			// Update: The odds of an operator being found in a legacy fight screen are extremely slim, so we won't limit it for now
			//if fs.ikemenver[0] == 0 && fs.ikemenver[1] == 0 { // Not "sys.fightScreen" yet
			//	return is
			//}
			targetPrefix := fmt.Sprintf("%s%v.", prefixBase, n)
			return resolveLifebarPrefixOperators(is, targetPrefix, prefixBase)
		}

		switch name {
		case "info":
			fs.name, _, _ = is.getText("name")
			fs.nameLow = strings.ToLower(fs.name)
			fs.author, _, _ = is.getText("author")
			fs.authorLow = strings.ToLower(fs.author)
			// Read MugenVersion
			if str, ok := is["mugenversion"]; ok {
				fs.mugenver, fs.mugenverF = ParseMugenVersion(str)
			}
			// Read IkemenVersion
			if str, ok := is["ikemenversion"]; ok {
				fs.ikemenver, fs.ikemenverF = ParseIkemenVersion(str)
			}
			var b bool
			is.ReadBool("doubleres", &b)
			if b {
				fs.fnt_scale = 0.5
			}
			// Localcoord/scale already pre-initialized above to unblock early FightFX
		case "files":
			if filesflg {
				filesflg = false
				if is.LoadFile("sff", []string{def, sys.motif.Def, "", "data/"}, "",
					func(filename string) error {
						s, err := loadSff(filename, false, true, false)
						if err != nil {
							return err
						}
						*fs.sff = *s
						return nil
					}); err != nil {
					return nil, err
				}
				if is.LoadFile("snd", []string{def, sys.motif.Def, "", "data/"}, "",
					func(filename string) error {
						s, err := LoadSnd(filename)
						if err != nil {
							return err
						}
						*fs.snd = *s
						return nil
					}); err != nil {
					return nil, err
				}
				if is.LoadFile("fightfx.sff", []string{def, sys.motif.Def, "", "data/"}, "",
					func(filename string) error {
						s, err := loadSff(filename, false, true, false)
						if err != nil {
							return err
						}
						*ffx.sff = *s
						return nil
					}); err != nil {
					return nil, err
				}
				if is.LoadFile("fightfx.air", []string{def, sys.motif.Def, "", "data/"}, "",
					func(filename string) error {
						str, err := LoadText(filename)
						if err != nil {
							return err
						}
						lines, i := SplitAndTrim(str, "\n"), 0
						ffx.animTable = ReadAnimationTable(filename, ffx.sff, &ffx.sff.palList, lines, &i, true)
						return nil
					}); err != nil {
					return nil, err
				}
				if is.LoadFile("common.snd", []string{def, sys.motif.Def, "", "data/"}, "",
					func(filename string) error {
						ffx.snd, err = LoadSnd(filename)
						return err
					}); err != nil {
					return nil, err
				}
				for i := 1; i <= fs.fx_limit; i++ {
					if err := is.LoadFile(fmt.Sprintf("fx%v", i), []string{def, sys.motif.Def, "", "data/"}, "",
						func(filename string) error {
							if err := loadFightFx(filename, false, true); err != nil {
								return err
							}
							return nil
						}); err != nil {
						return nil, err
					}
				}

				type fontSpec struct {
					path   string
					height int32
				}
				fntSpecs := map[int]fontSpec{}
				parseFonts := func(is IniSection) {
					for k, v := range is {
						if strings.HasPrefix(k, "font") {
							rest := k[4:]
							if rest == "" {
								continue
							}
							j := 0
							for j < len(rest) && rest[j] >= '0' && rest[j] <= '9' {
								j++
							}
							if j == 0 {
								continue
							}
							idx := int(Atoi(rest[:j]))
							tail := rest[j:]
							fs := fntSpecs[idx]
							if tail == "" {
								fs.path = v
							} else if tail == ".height" {
								fs.height = int32(Atoi(v))
							}
							fntSpecs[idx] = fs
						}
					}
				}
				parseFonts(is)

				// Load each declared font using the parsed specs
				for i, spec := range fntSpecs {
					if len(spec.path) == 0 {
						continue
					}
					_ = LoadFile(&spec.path, []string{def, sys.motif.Def, "", "data/"}, "font/",
						func(filename string) error {
							h := int32(-1)
							if spec.height != 0 {
								h = spec.height
							}
							if fnt, err := loadFnt(filename, h); err != nil {
								LogMessage("Failed to load %v (lifebar font%v): %v", filename, i, err)
								fs.fnt[i] = newFnt()
							} else {
								fs.fnt[i] = fnt
							}
							return nil
						},
					)
				}
			}
		case "fightfx":
			is.ReadF32("scale", &ffx.fx_scale)
		case "lifebar":
			if fs.lifeBars[0][0] == nil {
				fs.lifeBars[0][0] = readLifeBar("p1.", resolvedSection("p", 1), fs.sff, fs.animTable, fs.fnt)
			}
			if fs.lifeBars[0][1] == nil {
				fs.lifeBars[0][1] = readLifeBar("p2.", resolvedSection("p", 2), fs.sff, fs.animTable, fs.fnt)
			}
		case "powerbar":
			if fs.powerBars[0][0] == nil {
				fs.powerBars[0][0] = readPowerBar("p1.", resolvedSection("p", 1), fs.sff, fs.animTable, fs.fnt)
			}
			if fs.powerBars[0][1] == nil {
				fs.powerBars[0][1] = readPowerBar("p2.", resolvedSection("p", 2), fs.sff, fs.animTable, fs.fnt)
			}
		case "guardbar":
			if fs.guardBars[0][0] == nil {
				fs.guardBars[0][0] = readGuardBar("p1.", resolvedSection("p", 1), fs.sff, fs.animTable, fs.fnt)
			}
			if fs.guardBars[0][1] == nil {
				fs.guardBars[0][1] = readGuardBar("p2.", resolvedSection("p", 2), fs.sff, fs.animTable, fs.fnt)
			}
		case "stunbar":
			if fs.stunBars[0][0] == nil {
				fs.stunBars[0][0] = readStunBar("p1.", resolvedSection("p", 1), fs.sff, fs.animTable, fs.fnt)
			}
			if fs.stunBars[0][1] == nil {
				fs.stunBars[0][1] = readStunBar("p2.", resolvedSection("p", 2), fs.sff, fs.animTable, fs.fnt)
			}
		case "face":
			if fs.faces[0][0] == nil {
				fs.faces[0][0] = readFightScreenFace("p1.", resolvedSection("p", 1), fs.sff, fs.animTable)
			}
			if fs.faces[0][1] == nil {
				fs.faces[0][1] = readFightScreenFace("p2.", resolvedSection("p", 2), fs.sff, fs.animTable)
			}
		case "name":
			if fs.names[0][0] == nil {
				fs.names[0][0] = readFightScreenName("p1.", resolvedSection("p", 1), fs.sff, fs.animTable, fs.fnt)
			}
			if fs.names[0][1] == nil {
				fs.names[0][1] = readFightScreenName("p2.", resolvedSection("p", 2), fs.sff, fs.animTable, fs.fnt)
			}
		case "turns ":
			subname = strings.ToLower(subname)
			switch {
			case len(subname) >= 7 && subname[:7] == "lifebar":
				if fs.lifeBars[2][0] == nil {
					fs.lifeBars[2][0] = readLifeBar("p1.", resolvedSection("p", 1), fs.sff, fs.animTable, fs.fnt)
				}
				if fs.lifeBars[2][1] == nil {
					fs.lifeBars[2][1] = readLifeBar("p2.", resolvedSection("p", 2), fs.sff, fs.animTable, fs.fnt)
				}
			case len(subname) >= 8 && subname[:8] == "powerbar":
				if fs.powerBars[2][0] == nil {
					fs.powerBars[2][0] = readPowerBar("p1.", resolvedSection("p", 1), fs.sff, fs.animTable, fs.fnt)
				}
				if fs.powerBars[2][1] == nil {
					fs.powerBars[2][1] = readPowerBar("p2.", resolvedSection("p", 2), fs.sff, fs.animTable, fs.fnt)
				}
			case len(subname) >= 8 && subname[:8] == "guardbar":
				if fs.guardBars[2][0] == nil {
					fs.guardBars[2][0] = readGuardBar("p1.", resolvedSection("p", 1), fs.sff, fs.animTable, fs.fnt)
				}
				if fs.guardBars[2][1] == nil {
					fs.guardBars[2][1] = readGuardBar("p2.", resolvedSection("p", 2), fs.sff, fs.animTable, fs.fnt)
				}
			case len(subname) >= 7 && subname[:7] == "stunbar":
				if fs.stunBars[2][0] == nil {
					fs.stunBars[2][0] = readStunBar("p1.", resolvedSection("p", 1), fs.sff, fs.animTable, fs.fnt)
				}
				if fs.stunBars[2][1] == nil {
					fs.stunBars[2][1] = readStunBar("p2.", resolvedSection("p", 2), fs.sff, fs.animTable, fs.fnt)
				}
			case len(subname) >= 4 && subname[:4] == "face":
				if fs.faces[2][0] == nil {
					fs.faces[2][0] = readFightScreenFace("p1.", resolvedSection("p", 1), fs.sff, fs.animTable)
				}
				if fs.faces[2][1] == nil {
					fs.faces[2][1] = readFightScreenFace("p2.", resolvedSection("p", 2), fs.sff, fs.animTable)
				}
			case len(subname) >= 4 && subname[:4] == "name":
				if fs.names[2][0] == nil {
					fs.names[2][0] = readFightScreenName("p1.", resolvedSection("p", 1), fs.sff, fs.animTable, fs.fnt)
				}
				if fs.names[2][1] == nil {
					fs.names[2][1] = readFightScreenName("p2.", resolvedSection("p", 2), fs.sff, fs.animTable, fs.fnt)
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
				if fs.lifeBars[i][0] == nil {
					fs.lifeBars[i][0] = readLifeBar("p1.", resolvedSection("p", 1), fs.sff, fs.animTable, fs.fnt)
				}
				if fs.lifeBars[i][1] == nil {
					fs.lifeBars[i][1] = readLifeBar("p2.", resolvedSection("p", 2), fs.sff, fs.animTable, fs.fnt)
				}
				if fs.lifeBars[i][2] == nil {
					fs.lifeBars[i][2] = readLifeBar("p3.", resolvedSection("p", 3), fs.sff, fs.animTable, fs.fnt)
				}
				if fs.lifeBars[i][3] == nil {
					fs.lifeBars[i][3] = readLifeBar("p4.", resolvedSection("p", 4), fs.sff, fs.animTable, fs.fnt)
				}
				if fs.lifeBars[i][4] == nil {
					fs.lifeBars[i][4] = readLifeBar("p5.", resolvedSection("p", 5), fs.sff, fs.animTable, fs.fnt)
				}
				if fs.lifeBars[i][5] == nil {
					fs.lifeBars[i][5] = readLifeBar("p6.", resolvedSection("p", 6), fs.sff, fs.animTable, fs.fnt)
				}
				if i != 4 && i != 6 {
					if fs.lifeBars[i][6] == nil {
						fs.lifeBars[i][6] = readLifeBar("p7.", resolvedSection("p", 7), fs.sff, fs.animTable, fs.fnt)
					}
					if fs.lifeBars[i][7] == nil {
						fs.lifeBars[i][7] = readLifeBar("p8.", resolvedSection("p", 8), fs.sff, fs.animTable, fs.fnt)
					}
				}
			case len(subname) >= 8 && subname[:8] == "powerbar":
				if fs.powerBars[i][0] == nil {
					fs.powerBars[i][0] = readPowerBar("p1.", resolvedSection("p", 1), fs.sff, fs.animTable, fs.fnt)
				}
				if fs.powerBars[i][1] == nil {
					fs.powerBars[i][1] = readPowerBar("p2.", resolvedSection("p", 2), fs.sff, fs.animTable, fs.fnt)
				}
				if fs.powerBars[i][2] == nil {
					fs.powerBars[i][2] = readPowerBar("p3.", resolvedSection("p", 3), fs.sff, fs.animTable, fs.fnt)
				}
				if fs.powerBars[i][3] == nil {
					fs.powerBars[i][3] = readPowerBar("p4.", resolvedSection("p", 4), fs.sff, fs.animTable, fs.fnt)
				}
				if fs.powerBars[i][4] == nil {
					fs.powerBars[i][4] = readPowerBar("p5.", resolvedSection("p", 5), fs.sff, fs.animTable, fs.fnt)
				}
				if fs.powerBars[i][5] == nil {
					fs.powerBars[i][5] = readPowerBar("p6.", resolvedSection("p", 6), fs.sff, fs.animTable, fs.fnt)
				}
				if i != 4 && i != 6 {
					if fs.powerBars[i][6] == nil {
						fs.powerBars[i][6] = readPowerBar("p7.", resolvedSection("p", 7), fs.sff, fs.animTable, fs.fnt)
					}
					if fs.powerBars[i][7] == nil {
						fs.powerBars[i][7] = readPowerBar("p8.", resolvedSection("p", 8), fs.sff, fs.animTable, fs.fnt)
					}
				}
			case len(subname) >= 8 && subname[:8] == "guardbar":
				if fs.guardBars[i][0] == nil {
					fs.guardBars[i][0] = readGuardBar("p1.", resolvedSection("p", 1), fs.sff, fs.animTable, fs.fnt)
				}
				if fs.guardBars[i][1] == nil {
					fs.guardBars[i][1] = readGuardBar("p2.", resolvedSection("p", 2), fs.sff, fs.animTable, fs.fnt)
				}
				if fs.guardBars[i][2] == nil {
					fs.guardBars[i][2] = readGuardBar("p3.", resolvedSection("p", 3), fs.sff, fs.animTable, fs.fnt)
				}
				if fs.guardBars[i][3] == nil {
					fs.guardBars[i][3] = readGuardBar("p4.", resolvedSection("p", 4), fs.sff, fs.animTable, fs.fnt)
				}
				if fs.guardBars[i][4] == nil {
					fs.guardBars[i][4] = readGuardBar("p5.", resolvedSection("p", 5), fs.sff, fs.animTable, fs.fnt)
				}
				if fs.guardBars[i][5] == nil {
					fs.guardBars[i][5] = readGuardBar("p6.", resolvedSection("p", 6), fs.sff, fs.animTable, fs.fnt)
				}
				if i != 4 && i != 6 {
					if fs.guardBars[i][6] == nil {
						fs.guardBars[i][6] = readGuardBar("p7.", resolvedSection("p", 7), fs.sff, fs.animTable, fs.fnt)
					}
					if fs.guardBars[i][7] == nil {
						fs.guardBars[i][7] = readGuardBar("p8.", resolvedSection("p", 8), fs.sff, fs.animTable, fs.fnt)
					}
				}
			case len(subname) >= 7 && subname[:7] == "stunbar":
				if fs.stunBars[i][0] == nil {
					fs.stunBars[i][0] = readStunBar("p1.", resolvedSection("p", 1), fs.sff, fs.animTable, fs.fnt)
				}
				if fs.stunBars[i][1] == nil {
					fs.stunBars[i][1] = readStunBar("p2.", resolvedSection("p", 2), fs.sff, fs.animTable, fs.fnt)
				}
				if fs.stunBars[i][2] == nil {
					fs.stunBars[i][2] = readStunBar("p3.", resolvedSection("p", 3), fs.sff, fs.animTable, fs.fnt)
				}
				if fs.stunBars[i][3] == nil {
					fs.stunBars[i][3] = readStunBar("p4.", resolvedSection("p", 4), fs.sff, fs.animTable, fs.fnt)
				}
				if fs.stunBars[i][4] == nil {
					fs.stunBars[i][4] = readStunBar("p5.", resolvedSection("p", 5), fs.sff, fs.animTable, fs.fnt)
				}
				if fs.stunBars[i][5] == nil {
					fs.stunBars[i][5] = readStunBar("p6.", resolvedSection("p", 6), fs.sff, fs.animTable, fs.fnt)
				}
				if i != 4 && i != 6 {
					if fs.stunBars[i][6] == nil {
						fs.stunBars[i][6] = readStunBar("p7.", resolvedSection("p", 7), fs.sff, fs.animTable, fs.fnt)
					}
					if fs.stunBars[i][7] == nil {
						fs.stunBars[i][7] = readStunBar("p8.", resolvedSection("p", 8), fs.sff, fs.animTable, fs.fnt)
					}
				}
			case len(subname) >= 4 && subname[:4] == "face":
				if fs.faces[i][0] == nil {
					fs.faces[i][0] = readFightScreenFace("p1.", resolvedSection("p", 1), fs.sff, fs.animTable)
				}
				if fs.faces[i][1] == nil {
					fs.faces[i][1] = readFightScreenFace("p2.", resolvedSection("p", 2), fs.sff, fs.animTable)
				}
				if fs.faces[i][2] == nil {
					fs.faces[i][2] = readFightScreenFace("p3.", resolvedSection("p", 3), fs.sff, fs.animTable)
				}
				if fs.faces[i][3] == nil {
					fs.faces[i][3] = readFightScreenFace("p4.", resolvedSection("p", 4), fs.sff, fs.animTable)
				}
				if fs.faces[i][4] == nil {
					fs.faces[i][4] = readFightScreenFace("p5.", resolvedSection("p", 5), fs.sff, fs.animTable)
				}
				if fs.faces[i][5] == nil {
					fs.faces[i][5] = readFightScreenFace("p6.", resolvedSection("p", 6), fs.sff, fs.animTable)
				}
				if i != 4 && i != 6 {
					if fs.faces[i][6] == nil {
						fs.faces[i][6] = readFightScreenFace("p7.", resolvedSection("p", 7), fs.sff, fs.animTable)
					}
					if fs.faces[i][7] == nil {
						fs.faces[i][7] = readFightScreenFace("p8.", resolvedSection("p", 8), fs.sff, fs.animTable)
					}
				}
			case len(subname) >= 4 && subname[:4] == "name":
				if fs.names[i][0] == nil {
					fs.names[i][0] = readFightScreenName("p1.", resolvedSection("p", 1), fs.sff, fs.animTable, fs.fnt)
				}
				if fs.names[i][1] == nil {
					fs.names[i][1] = readFightScreenName("p2.", resolvedSection("p", 2), fs.sff, fs.animTable, fs.fnt)
				}
				if fs.names[i][2] == nil {
					fs.names[i][2] = readFightScreenName("p3.", resolvedSection("p", 3), fs.sff, fs.animTable, fs.fnt)
				}
				if fs.names[i][3] == nil {
					fs.names[i][3] = readFightScreenName("p4.", resolvedSection("p", 4), fs.sff, fs.animTable, fs.fnt)
				}
				if fs.names[i][4] == nil {
					fs.names[i][4] = readFightScreenName("p5.", resolvedSection("p", 5), fs.sff, fs.animTable, fs.fnt)
				}
				if fs.names[i][5] == nil {
					fs.names[i][5] = readFightScreenName("p6.", resolvedSection("p", 6), fs.sff, fs.animTable, fs.fnt)
				}
				if i != 4 && i != 6 {
					if fs.names[i][6] == nil {
						fs.names[i][6] = readFightScreenName("p7.", resolvedSection("p", 7), fs.sff, fs.animTable, fs.fnt)
					}
					if fs.names[i][7] == nil {
						fs.names[i][7] = readFightScreenName("p8.", resolvedSection("p", 8), fs.sff, fs.animTable, fs.fnt)
					}
				}
			}
		case "winicon":
			if fs.winIcons[0] == nil {
				fs.winIcons[0] = readFightScreenWinIcon("p1.", resolvedSection("p", 1), fs.sff, fs.animTable, fs.fnt)
			}
			if fs.winIcons[1] == nil {
				fs.winIcons[1] = readFightScreenWinIcon("p2.", resolvedSection("p", 2), fs.sff, fs.animTable, fs.fnt)
			}
		case "time":
			if fs.time == nil {
				fs.time = readFightScreenTime(is, fs.sff, fs.animTable, fs.fnt)
			}
		case "combo":
			for team := 0; team < 2; team++ {
				if fs.combos[team] != nil {
					continue
				}

				teamNum := team + 1
				teamPrefix := fmt.Sprintf("team%v.", teamNum)

				if fs.ikemenver[0] == 0 && fs.ikemenver[1] == 0 { // Not sys.fightScreen yet
					// Check if "teamX.pos" exists to determine whether or not to use "teamX" prefixes
					// Mugen works slightly different and apparently checks whether "pos" or "team1.pos" is found first in order to determine what syntax to use
					// Such prefixes were only added in Mugen 1.0, hence it running this check
					if _, ok := is[teamPrefix+"pos"]; ok {
						fs.combos[team] = readFightScreenCombo(teamPrefix, is, fs.sff, fs.animTable, fs.fnt, team)
					} else {
						fs.combos[team] = readFightScreenCombo("", is, fs.sff, fs.animTable, fs.fnt, team)
					}
				} else {
					// In Ikemen version we just read them like other prefixed sections
					fs.combos[team] = readFightScreenCombo(teamPrefix, resolvedSection("team", teamNum), fs.sff, fs.animTable, fs.fnt, team)
				}
			}
		case "action":
			if fs.actions[0] == nil {
				fs.actions[0] = readFightScreenAction("team1.", resolvedSection("team", 1), fs.fnt)
			}
			if fs.actions[1] == nil {
				fs.actions[1] = readFightScreenAction("team2.", resolvedSection("team", 2), fs.fnt)
			}
		case "round":
			if fs.round == nil {
				fs.round = readFightScreenRound(is, fs.sff, fs.animTable, fs.snd, fs.fnt)
			}
		case "timer":
			if fs.timer == nil {
				fs.timer = readFightScreenTimer(is, fs.sff, fs.animTable, fs.fnt)
			}
		case "score":
			if fs.scores[0] == nil {
				fs.scores[0] = readFightScreenScore("p1.", resolvedSection("p", 1), fs.sff, fs.animTable, fs.fnt)
			}
			if fs.scores[1] == nil {
				fs.scores[1] = readFightScreenScore("p2.", resolvedSection("p", 2), fs.sff, fs.animTable, fs.fnt)
			}
		case "match":
			if fs.match == nil {
				fs.match = readFightScreenMatch(is, fs.sff, fs.animTable, fs.fnt)
			}
		case "ailevel":
			if fs.aiLevels[0] == nil {
				fs.aiLevels[0] = readFightScreenAiLevel("p1.", resolvedSection("p", 1), fs.sff, fs.animTable, fs.fnt)
			}
			if fs.aiLevels[1] == nil {
				fs.aiLevels[1] = readFightScreenAiLevel("p2.", resolvedSection("p", 2), fs.sff, fs.animTable, fs.fnt)
			}
		case "wincount":
			if fs.winCounts[0] == nil {
				fs.winCounts[0] = readFightScreenWinCount("p1.", resolvedSection("p", 1), fs.sff, fs.animTable, fs.fnt)
			}
			if fs.winCounts[1] == nil {
				fs.winCounts[1] = readFightScreenWinCount("p2.", resolvedSection("p", 2), fs.sff, fs.animTable, fs.fnt)
			}
		case "mode":
			if fs.modes == nil {
				fs.modes = readFightScreenMode(is, fs.sff, fs.animTable, fs.fnt)
			}
		}
	}
	sys.ffx["f"] = ffx
	// fightfx scale
	//if math.IsNaN(float64(sys.ffx["f"].fx_scale)) {
	//	sys.ffx["f"].fx_scale = float32(fs.localcoord[0]) / 320
	//}
	// Now calculated in each animation call
	//for _, a := range sys.ffx["f"].animTable {
	//	a.start_scale = [...]float32{sys.ffx["f"].fx_scale, sys.ffx["f"].fx_scale}
	//}
	// Iterate over map in a stable iteration order
	keys := make([]string, 0, len(fs.missing))
	for k := range fs.missing {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		if strings.Contains(k, " lifebar") {
			for i := 3; i < len(fs.lifeBars); i++ {
				if i == fs.missing[k] {
					for j := 0; j < len(fs.lifeBars[i]); j++ {
						if i == 6 || i == 7 {
							fs.lifeBars[i][j] = fs.lifeBars[3][j]
						} else {
							fs.lifeBars[i][j] = fs.lifeBars[1][j]
						}
					}
				}
			}
		} else if strings.Contains(k, " powerbar") {
			for i := 1; i < len(fs.powerBars); i++ {
				if i == fs.missing[k] {
					for j := 0; j < 2; j++ {
						switch i {
						case 4, 5:
							fs.powerBars[i][j] = fs.powerBars[1][j]
						case 6, 7:
							fs.powerBars[i][j] = fs.powerBars[3][j]
						default:
							fs.powerBars[i][j] = fs.powerBars[0][j]
						}
					}
				}
			}
		} else if strings.Contains(k, " guardbar") {
			for i := 1; i < len(fs.guardBars); i++ {
				if i == fs.missing[k] {
					for j := 0; j < 2; j++ {
						switch i {
						case 4, 5:
							fs.guardBars[i][j] = fs.guardBars[1][j]
						case 6, 7:
							fs.guardBars[i][j] = fs.guardBars[3][j]
						default:
							fs.guardBars[i][j] = fs.guardBars[0][j]
						}
					}
				}
			}
		} else if strings.Contains(k, " stunbar") {
			for i := 1; i < len(fs.stunBars); i++ {
				if i == fs.missing[k] {
					for j := 0; j < 2; j++ {
						switch i {
						case 4, 5:
							fs.stunBars[i][j] = fs.stunBars[1][j]
						case 6, 7:
							fs.stunBars[i][j] = fs.stunBars[3][j]
						default:
							fs.stunBars[i][j] = fs.stunBars[0][j]
						}
					}
				}
			}
		} else if strings.Contains(k, " face") {
			for i := 3; i < len(fs.faces); i++ {
				if i == fs.missing[k] {
					for j := 0; j < len(fs.faces[i]); j++ {
						if i == 6 || i == 7 {
							fs.faces[i][j] = fs.faces[3][j]
						} else {
							fs.faces[i][j] = fs.faces[1][j]
						}
					}
				}
			}
		} else if strings.Contains(k, " name") {
			for i := 3; i < len(fs.names); i++ {
				if i == fs.missing[k] {
					for j := 0; j < len(fs.names[i]); j++ {
						if i == 6 || i == 7 {
							fs.names[i][j] = fs.names[3][j]
						} else {
							fs.names[i][j] = fs.names[1][j]
						}
					}
				}
			}
		}
	}
	fs.def = def

	/*
		// TODO: This is a hack
		// During team life sharing, replace Tag/Simul lifebars with single
		// For some reason just doing this in the drawing function doesn't work
		if sys.cfg.Options.Team.LifeShare {
			teamModeIndices := []int{1, 3, 4, 5, 6, 7} // Simul, Tag, and their 3P/4P variants

			for _, i := range teamModeIndices {
				if i < len(fs.lb) {
					// Replace with Single lifebars
					fs.lifeBars[i][0] = fs.lifeBars[0][0]
					fs.lifeBars[i][1] = fs.lifeBars[0][1]
				}
			}
		}
	*/

	return fs, nil
}

func (fs *FightScreen) reload() error {
	new, err := loadFightScreen(fs.def)
	if err != nil {
		return err
	}
	new.time.framespercount = fs.time.framespercount
	//new.round.match_wins = fs.round.match_wins
	//new.round.match_maxdrawgames = fs.round.match_maxdrawgames
	new.timer.active = fs.timer.active
	new.scores[0].active = fs.scores[0].active
	new.scores[1].active = fs.scores[1].active
	new.match.active = fs.match.active
	new.aiLevels[0].active = fs.aiLevels[0].active
	new.aiLevels[1].active = fs.aiLevels[1].active
	new.winCounts[0].active = fs.winCounts[0].active
	new.winCounts[1].active = fs.winCounts[1].active
	new.winCounts[0].wins = fs.winCounts[0].wins
	new.winCounts[1].wins = fs.winCounts[1].wins
	new.active = fs.active
	new.bars = fs.bars
	new.mode = fs.mode
	new.redlifebar = fs.redlifebar
	new.guardbar = fs.guardbar
	new.stunbar = fs.stunbar
	//new.fx_scale = fs.fx_scale
	sys.fightScreen = *new
	return nil
}

func (fs *FightScreen) step() {
	if sys.paused && !sys.frameStepFlag {
		return
	}
	// Bars
	for ti := range sys.tmode {
		layout := fs.curLayout[ti]
		for i, charpn := range fs.teamOrder[ti] {
			index := i*2 + ti
			// LifeBar
			fs.lifeBars[layout][index].step(charpn, fs.lifeBars[layout][charpn])
			// PowerBar
			fs.powerBars[layout][index].step(charpn, fs.powerBars[layout][charpn], fs.snd)
			// GuardBar
			fs.guardBars[layout][index].step(charpn, fs.guardBars[layout][charpn], fs.snd)
			// StunBar
			fs.stunBars[layout][index].step(charpn, fs.stunBars[layout][charpn], fs.snd)
			// Face
			fs.faces[layout][index].step(charpn, fs.faces[layout][charpn])
			// Name
			fs.names[layout][index].step()
		}
	}
	// WinIcon
	for i := range fs.winIcons {
		fs.winIcons[i].step(sys.wins[i])
	}
	// Time
	fs.time.step()
	cb, cd, cp := [2]int32{}, [2]int32{}, [2]float32{}
	targets := [2]int32{}
	// Combo
	for _, ch := range sys.chars {
		// Iterate through all players to see the combo status of each team
		for _, c := range ch {
			if c.receivedHits > 0 && (c.teamside == 0 || c.teamside == 1) && (c.alive() || !c.scf(SCF_over_ko)) { // If alive or not alive but not in state 5150 yet
				side := 1 - c.teamside
				cb[side] += c.receivedHits
				cd[side] += c.receivedDmg
				// Perhaps helper percentages shouldn't be tracked, but ignoring them creates scenarios where the lifebars show 0% damage which looks wrong
				cp[side] += float32(c.receivedDmg) / float32(c.lifeMax) * 100
				targets[side]++
			}
		}
	}
	for side := 0; side < 2; side++ {
		if targets[side] > 0 {
			cp[side] /= float32(targets[side]) // Divide damage percentage by number of valid enemies
		}
	}
	for i := range fs.combos {
		fs.combos[i].step(i, cb[i], cd[i], cp[i]) // Combo hits, combo damage, combo damage percentage
	}
	// Action
	for i := range fs.actions {
		fs.actions[i].step(fs.teamOrder[i][0])
	}
	// Timer
	fs.timer.step()
	// Score
	for i := range fs.scores {
		fs.scores[i].step()
	}
	// Match
	fs.match.step()
	// AiLevel
	for i := range fs.aiLevels {
		fs.aiLevels[i].step()
	}
	// WinCount
	for i := range fs.winCounts {
		fs.winCounts[i].step()
	}
	// Mode
	if _, ok := fs.modes[sys.gameMode]; ok {
		fs.modes[sys.gameMode].step()
	}
}

// Resets fight screen as well as prepares team mode configuration for each player
func (fs *FightScreen) reset() {
	//var num [2]int

	// Update team mode layout for each player
	for ti, tm := range sys.tmode {
		fs.curLayout[ti] = int(tm)
		if tm == TM_Simul {
			if sys.numSimul[ti] == 3 {
				fs.curLayout[ti] = 4 // Simul_3P (6)
			} else if sys.numSimul[ti] >= 4 {
				fs.curLayout[ti] = 5 // Simul_4P (8)
			} else {
				fs.curLayout[ti] = 1 // Simul (8)
			}
		} else if tm == TM_Tag {
			if sys.numSimul[ti] == 3 {
				fs.curLayout[ti] = 6 // Tag_3P (6)
			} else if sys.numSimul[ti] >= 4 {
				fs.curLayout[ti] = 7 // Tag_4P (8)
			} else {
				fs.curLayout[ti] = 3 // Tag (8)
			}
		} else if tm == TM_Turns {
			fs.curLayout[ti] = 2 // Turns (2)
		} else {
			fs.curLayout[ti] = 0 // Single (2)
		}

		// Determine number of players in each team
		//if tm == TM_Simul || tm == TM_Tag {
		//	num[ti] = int(math.Min(8, float64(sys.numSimul[ti])*2))
		//} else {
		//	num[ti] = len(fs.lifeBars[fs.curLayout[ti]]) // TODO: Why did it check this for single/turns but not simul/tag?
		//}
		// Build team order by ascending player number
		//fs.teamOrder[ti] = []int{}
		//for i := ti; i < num[ti]; i += 2 {
		//	fs.teamOrder[ti] = append(fs.teamOrder[ti], i)
		//}
	}

	// Rebuild order of teams
	for side := range fs.teamOrder {
		fs.syncTeamOrder(side)
	}

	// Reset fight screen elements
	for i := range fs.lifeBars {
		for j := range fs.lifeBars[i] {
			fs.lifeBars[i][j].reset()
		}
	}
	for i := range fs.powerBars {
		for j := range fs.powerBars[i] {
			fs.powerBars[i][j].reset()
		}
	}
	for i := range fs.guardBars {
		for j := range fs.guardBars[i] {
			fs.guardBars[i][j].reset()
		}
	}
	for i := range fs.stunBars {
		for j := range fs.stunBars[i] {
			fs.stunBars[i][j].reset()
		}
	}
	for i := range fs.faces {
		for j := range fs.faces[i] {
			fs.faces[i][j].reset()
		}
	}
	for i := range fs.names {
		for j := range fs.names[i] {
			fs.names[i][j].reset()
		}
	}
	for i := range fs.winIcons {
		fs.winIcons[i].reset()
	}
	fs.time.reset()
	for i := range fs.combos {
		fs.combos[i].reset()
	}
	for i := range fs.actions {
		fs.actions[i].reset(fs.teamOrder[i][0])
	}
	fs.round.reset()
	fs.timer.reset()
	for i := range fs.scores {
		fs.scores[i].reset()
	}
	fs.match.reset()
	for i := range fs.aiLevels {
		fs.aiLevels[i].reset()
	}
	for i := range fs.winCounts {
		fs.winCounts[i].reset()
	}
	if _, ok := fs.modes[sys.gameMode]; ok {
		fs.modes[sys.gameMode].reset()
	}
}

func (fs *FightScreen) visible() bool {
	return fs.active &&
		!sys.postMatchFlg &&
		!sys.lifebarHide &&
		!(sys.dialogueHideBars || sys.motif.di.active) &&
		!(sys.motif.me.active && sys.motif.PauseMenu["pause_menu"].HideBars &&
			(!sys.motif.me.closeRequested || sys.paused))
}

func (fs *FightScreen) draw(layerno int16) {
	if fs.visible() {
		if !sys.gsf(GSF_nobardisplay) && fs.bars {
			// Helper to determine whether to iterate elements forward or backward (drawing order)
			iterationOrder := func(leaderontop bool) (int, int, int) {
				if leaderontop {
					return MaxSimul - 1, -1, -1
				}
				return 0, MaxSimul, 1
			}
			// LifeBar
			for side := 0; side < len(sys.tmode); side++ {
				layout := fs.curLayout[side]
				slotStart, slotEnd, slotStep := iterationOrder(fs.lifeBars[layout][side].leaderontop)

				for slot := slotStart; slot != slotEnd; slot += slotStep {
					if slot >= len(fs.teamOrder[side]) {
						continue
					}
					barpn := slot*2 + side
					charpn := fs.teamOrder[side][slot]
					c := sys.chars[charpn][0]
					if c.asf(ASF_nolifebardisplay) {
						continue
					}
					fs.lifeBars[layout][barpn].bgDraw(layerno)
					fs.lifeBars[layout][barpn].draw(layerno, charpn, fs.lifeBars[layout][charpn], fs.fnt)
				}
			}

			// PowerBar
			for side := 0; side < len(sys.tmode); side++ {
				layout := fs.curLayout[side]
				slotStart, slotEnd, slotStep := iterationOrder(fs.powerBars[layout][side].leaderontop)

				for slot := slotStart; slot != slotEnd; slot += slotStep {
					if slot >= len(fs.teamOrder[side]) {
						continue
					}
					barpn := slot*2 + side
					charpn := fs.teamOrder[side][slot]
					c := sys.chars[charpn][0]
					if c.asf(ASF_nopowerbardisplay) {
						continue
					}
					tm := sys.tmode[side]
					if slot != 0 && (tm == TM_Simul || tm == TM_Tag) && sys.cfg.Options.Team.PowerShare {
						continue
					}
					fs.powerBars[layout][barpn].bgDraw(layerno, barpn)
					fs.powerBars[layout][barpn].draw(layerno, charpn, fs.powerBars[layout][charpn], fs.fnt)
				}
			}

			// GuardBar
			for side := 0; side < len(sys.tmode); side++ {
				layout := fs.curLayout[side]
				slotStart, slotEnd, slotStep := iterationOrder(fs.guardBars[layout][side].leaderontop)

				for slot := slotStart; slot != slotEnd; slot += slotStep {
					if slot >= len(fs.teamOrder[side]) {
						continue
					}
					barpn := slot*2 + side
					charpn := fs.teamOrder[side][slot]
					c := sys.chars[charpn][0]
					if !c.guardBreakEnabled() || c.asf(ASF_noguardbardisplay) {
						continue
					}
					fs.guardBars[layout][barpn].bgDraw(layerno)
					fs.guardBars[layout][barpn].draw(layerno, charpn, fs.guardBars[layout][charpn], fs.fnt)
				}
			}

			// StunBar
			for side := 0; side < len(sys.tmode); side++ {
				layout := fs.curLayout[side]
				slotStart, slotEnd, slotStep := iterationOrder(fs.stunBars[layout][side].leaderontop)

				for slot := slotStart; slot != slotEnd; slot += slotStep {
					if slot >= len(fs.teamOrder[side]) {
						continue
					}
					barpn := slot*2 + side
					charpn := fs.teamOrder[side][slot]
					c := sys.chars[charpn][0]
					if !c.dizzyEnabled() || c.asf(ASF_nostunbardisplay) {
						continue
					}
					fs.stunBars[layout][barpn].bgDraw(layerno)
					fs.stunBars[layout][barpn].draw(layerno, charpn, fs.stunBars[layout][charpn], fs.fnt)
				}
			}

			// Face
			for side := 0; side < len(sys.tmode); side++ {
				layout := fs.curLayout[side]
				slotStart, slotEnd, slotStep := iterationOrder(fs.faces[layout][side].leaderontop)

				for slot := slotStart; slot != slotEnd; slot += slotStep {
					if slot >= len(fs.teamOrder[side]) {
						continue
					}
					barpn := slot*2 + side
					charpn := fs.teamOrder[side][slot]
					c := sys.chars[charpn][0]
					if c.asf(ASF_nofacedisplay) {
						continue
					}
					// Draw Turns teammates from the first bar only
					if slot == 0 {
						fs.faces[layout][side].drawTeammates(layerno)
					}
					fs.faces[layout][barpn].bgDraw(layerno)
					fs.faces[layout][barpn].draw(layerno, charpn, fs.faces[layout][charpn])
				}
			}

			// Name
			for side := 0; side < len(sys.tmode); side++ {
				layout := fs.curLayout[side]
				slotStart, slotEnd, slotStep := iterationOrder(fs.names[layout][side].leaderontop)

				for slot := slotStart; slot != slotEnd; slot += slotStep {
					if slot >= len(fs.teamOrder[side]) {
						continue
					}
					barpn := slot*2 + side
					charpn := fs.teamOrder[side][slot]
					c := sys.chars[charpn][0]
					if c.asf(ASF_nonamedisplay) {
						continue
					}
					// Draw Turns teammates from the first bar only
					if slot == 0 {
						fs.names[layout][side].drawTeammates(layerno, fs.fnt, side)
					}
					fs.names[layout][barpn].bgDraw(layerno)
					fs.names[layout][barpn].draw(layerno, charpn, fs.fnt, side)
				}
			}

			// Time
			if !sys.gsf(GSF_notimedisplay) {
				fs.time.bgDraw(layerno)
				fs.time.draw(layerno, fs.fnt)
			}

			// WinIcon
			for i := 0; i < len(fs.winIcons); i++ {
				if !sys.chars[i][0].asf(ASF_nowinicondisplay) {
					fs.winIcons[i].draw(layerno, fs.fnt, i)
				}
			}

			// Timer
			fs.timer.bgDraw(layerno)
			fs.timer.draw(layerno, fs.fnt)

			// Score
			for i := 0; i < len(fs.scores); i++ {
				fs.scores[i].bgDraw(layerno)
				fs.scores[i].draw(layerno, fs.fnt, i)
			}

			// Match
			fs.match.bgDraw(layerno)
			fs.match.draw(layerno, fs.fnt)

			// AiLevel
			for i := 0; i < len(fs.aiLevels); i++ {
				fs.aiLevels[i].bgDraw(layerno)
				fs.aiLevels[i].draw(layerno, fs.fnt, sys.aiLevel[sys.chars[i][0].playerNo])
			}

			// WinCount
			for i := 0; i < len(fs.winCounts); i++ {
				fs.winCounts[i].bgDraw(layerno)
				fs.winCounts[i].draw(layerno, fs.fnt, i)
			}
		}

		// Combo
		for i := 0; i < len(fs.combos); i++ {
			if !sys.chars[i][0].asf(ASF_nocombodisplay) {
				fs.combos[i].draw(layerno, fs.fnt, i)
			}
		}

		// Action
		for i := 0; i < len(fs.actions); i++ {
			if !sys.chars[i][0].asf(ASF_nolifebaraction) {
				fs.actions[i].draw(layerno, fs.fnt, i)
			}
		}

		// Mode
		if _, ok := fs.modes[sys.gameMode]; ok {
			fs.modes[sys.gameMode].bgDraw(layerno)
			fs.modes[sys.gameMode].draw(layerno, fs.fnt)
		}
	}

	if fs.active && !sys.postMatchFlg {
		// Round
		fs.round.draw(layerno, fs.fnt)
	}
}

func (fs *FightScreen) drawFade() {
	if fs.active && !sys.postMatchFlg {
		fs.round.drawFade()
	}
}

func (fs *FightScreen) setScale() {
	// Calculate viewport for a 4:3 aspect ratio
	viewport43 := getViewport(float64(fs.localcoord[0]), float64(fs.localcoord[1]), 4, 3)

	// Calculate viewport based on the game's configured width and height
	viewport := getViewport(float64(fs.localcoord[0]), float64(fs.localcoord[1]), float64(sys.gameWidth), float64(sys.gameHeight))

	localW := float32(fs.localcoord[0])
	calcScale := float32(viewport[2]) / localW
	calcOffsetX := (float32(viewport43[2]) - localW*calcScale) / 2

	fs.offsetX = float32(calcOffsetX * calcScale)
	fs.scale = 320.0 / float32(viewport43[2]) * calcScale
	fs.portraitScale = localW / float32(viewport43[2]) * calcScale
}

func readMotifFightFromDef(def string) string {
	if def == "" {
		return ""
	}
	tmp := def
	var fight string
	// Use existing path resolution logic.
	_ = LoadFile(&tmp, []string{tmp, "", "data/"}, "", func(filename string) error {
		if filename == "" {
			return nil
		}
		str, err := LoadText(filename)
		if err != nil {
			return err
		}
		// Minimal parse: scan sections until [Files], then grab "fight".
		lines := SplitAndTrim(NormalizeNewlines(str), "\n")
		i := 0
		for i < len(lines) {
			is, name, _ := ReadIniSection(lines, &i)
			if len(name) == 0 {
				break
			}
			if strings.EqualFold(name, "Files") {
				if v, ok := is["fight"]; ok {
					fight = strings.TrimSpace(v)
				}
				break
			}
		}
		return nil
	})
	return fight
}

func (fs *FightScreen) resolvePath() {
	var v string
	if x, ok := sys.cmdFlags["-fight"]; ok {
		v = x
	} else {
		// Prefer already-parsed value.
		v = sys.motif.Files.Fight
		// If motif.Files.Fight is not set yet, resolve motif path and read only the [Files] fight entry from sys.motif.Def.
		if v == "" {
			sys.motif.resolvePath()
			v = readMotifFightFromDef(sys.motif.Def)
		}
	}
	v = filepath.ToSlash(v)
	if v == "" {
		v = "fight.def"
	}
	resolved := SearchFile(v, []string{sys.motif.Def, "", "data/"})
	if FileExist(resolved) != "" {
		fs.def = filepath.ToSlash(resolved)
		return
	}
	fs.def = v
}

func (fs *FightScreen) appendAction(c *Char, msg *FSMsg, refresh int32) {
	if c.teamside < 0 {
		return
	}
	if _, ok := fs.missing["[action]"]; ok {
		return
	}

	// Play sound
	if msg.snd[0] != -1 && msg.snd[1] != -1 {
		if msg.snd_ffx != "" && msg.snd_ffx != "s" && sys.ffx[msg.snd_ffx] != nil && sys.ffx[msg.snd_ffx].snd != nil {
			s := sys.ffx[msg.snd_ffx].snd.Get(msg.snd)
			if s != nil {
				sys.soundChannels.Play(s, msg.snd[0], msg.snd[1], 100, 0, 0, 0, 0)
			}
		} else {
			fs.snd.play(msg.snd, 100, 0, 0, 0, 0)
		}
	}

	// If sound only, we can stop here
	if msg.animNo == -1 && msg.spr[0] == -1 && msg.text == "" {
		return
	}

	// Select side of screen
	teammsg := fs.actions[c.teamside]

	// Duplicate detection
	if refresh == 1 || refresh == 2 {
		for _, existing := range teammsg.messages {
			if existing.del {
				continue
			}
			if existing.isEqual(msg) {
				// Found an identical message: refresh its timer to keep it alive
				existing.resttime = msg.resttime
				existing.agetimer = 0
				existing.del = false
				// Also reappear from outside screen
				if refresh == 2 {
					existing.counterX = teammsg.start_x * 2
				}
				// Don't add another
				return
			}
		}
	}

	// If adding a new message while exceeding the maximum number allowed, make the oldest message go away faster
	var count int32
	for _, v := range teammsg.messages {
		if !v.del {
			count++
		}
	}
	if count >= teammsg.max {
		var oldest int
		var oldestTimer int32
		for i, m := range teammsg.messages {
			if !m.del && m.resttime > 0 && m.agetimer > oldestTimer {
				oldest = i
				oldestTimer = m.agetimer
			}
		}
		if oldest < len(teammsg.messages) {
			teammsg.messages[oldest].resttime = 0
		}
	}

	// Use index 0 if "top". Otherwise find the first free message slot
	index := 0
	if !msg.top {
		for k, v := range teammsg.messages {
			if v.del {
				teammsg.messages = removeFSMsg(teammsg.messages, k)
				break
			}
			index++
		}
	}

	// Layout logic
	prefix := fmt.Sprintf("team%v.front.", c.teamside+1)
	delete(teammsg.is, prefix+"anim")
	delete(teammsg.is, prefix+"spr")

	// Read animation
	if msg.animNo != -1 {
		teammsg.is[prefix+"anim"] = fmt.Sprintf("%v", msg.animNo)
	}

	// Read sprite
	if msg.spr[0] != -1 {
		teammsg.is[prefix+"spr"] = fmt.Sprintf("%v,%v", msg.spr[0], msg.spr[1])
	}

	// Read background
	msg.bg = ReadAnimLayout(fmt.Sprintf("team%v.bg.", c.teamside+1), teammsg.is, fs.sff, fs.animTable, 2)

	// Default to fight screen assets
	sff := fs.sff
	at := fs.animTable
	alscale := float32(1.0)

	// Use animation prefixes to change asset source
	if msg.anim_ffx != "" {
		if msg.anim_ffx == "s" {
			// Use the character
			sff = sys.cgi[c.playerNo].sff
			at = sys.cgi[c.playerNo].animTable
			alscale = float32(fs.localcoord[0]) / float32(sys.chars[c.playerNo][0].localcoord)
		} else if sys.ffx[msg.anim_ffx] != nil {
			// Use common FX
			fx := sys.ffx[msg.anim_ffx]
			sff = fx.sff
			at = fx.animTable
			alscale = fx.fx_scale * float32(fs.localcoord[0]) / float32(fx.localcoord[0])
		}
	}

	msg.front = ReadAnimLayout(prefix, teammsg.is, sff, at, 2)

	// Normalize localcoord scaling across different asset sources
	if msg.front.anim != nil && alscale != 1.0 {
		msg.front.anim.start_scale[0] *= alscale
		msg.front.anim.start_scale[1] *= alscale
	}

	// Insert the new message
	teammsg.messages = insertFSMsg(teammsg.messages, msg, index)
}

// We track the combo separately from each player's received hits
// So that if partners take turns doing hits to different enemies the combo still adds up
// Only some games do this so we don't strictly need to do it. But it is also harmless. Maybe make it an option
// TODO: Combo damage should probably also stack like this, either way
func (fs *FightScreen) addComboHits(side int, n int32) {
	if side < 0 || side >= len(fs.combos) {
		return
	}
	sys.comboCount[side] += n
}

// Update team order based on the actual character member numbers
func (fs *FightScreen) syncTeamOrder(side int) {
	order := make([]int, 0, MaxSimul)
	// Collect all the members of this team
	for pn := side; pn < MaxSimul*2; pn += 2 {
		if len(sys.chars[pn]) == 0 || sys.chars[pn][0] == nil {
			continue
		}
		if sys.chars[pn][0].teamside != side {
			continue
		}
		order = append(order, pn)
	}
	// Sort them by memberNo
	sort.Slice(order, func(i, j int) bool {
		return sys.chars[order[i]][0].memberNo < sys.chars[order[j]][0].memberNo
	})
	// Save result
	fs.teamOrder[side] = order
}
