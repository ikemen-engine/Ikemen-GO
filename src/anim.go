package main

import (
	"math"
	"strconv"
	"strings"
)

type AnimFrame struct {
	Time          int32
	Group, Number int16
	X, Y          int16
	SrcAlpha      byte
	DstAlpha      byte
	H, V          int8
	Ex            [][]float32
}

func newAnimFrame() *AnimFrame {
	return &AnimFrame{Time: -1, Group: -1, SrcAlpha: 255, H: 1, V: 1}
}
func ReadAnimFrame(line string) *AnimFrame {
	if len(line) == 0 || (line[0] < '0' || '9' < line[0]) && line[0] != '-' {
		return nil
	}
	ary := strings.SplitN(line, ",", 7)
	if len(ary) < 5 {
		return nil
	}
	af := newAnimFrame()
	af.Group, af.Number = int16(Atoi(ary[0])), int16(Atoi(ary[1]))
	af.X, af.Y = int16(Atoi(ary[2])), int16(Atoi(ary[3]))
	af.Time = Atoi(ary[4])
	if len(ary) < 6 {
		return af
	}
	for i := range ary[5] {
		switch ary[5][i] {
		case 'H', 'h':
			af.H *= -1
		case 'V', 'v':
			af.V *= -1
		}
	}
	if af.H < 0 {
		af.X *= -1
	}
	if af.V < 0 {
		af.Y *= -1
	}
	if len(ary) < 7 {
		return af
	}
	ia := strings.IndexAny(ary[6], "ASas")
	if ia >= 0 {
		ary[6] = ary[6][ia:]
	}
	ary = SplitAndTrim(ary[6], ",")
	a := strings.ToLower(ary[0])
	switch {
	case a == "a1":
		af.SrcAlpha, af.DstAlpha = 255, 128
	case len(a) > 0 && a[0] == 's':
		af.SrcAlpha, af.DstAlpha = 1, 255
	case len(a) >= 2 && a[:2] == "as":
		i := strings.IndexAny(a, "d")
		if i >= 0 {
			sa := Atoi(a[2:i])
			if sa <= 0 {
				af.SrcAlpha = 0
			} else if sa >= 255 {
				af.SrcAlpha = 255
			} else {
				af.SrcAlpha = byte(sa)
			}
			da := Atoi(a[i+1:])
			if da <= 0 {
				af.DstAlpha = 0
			} else if da >= 255 {
				af.DstAlpha = 255
			} else {
				af.DstAlpha = byte(da)
			}
			if af.SrcAlpha == 1 && af.DstAlpha == 255 {
				af.SrcAlpha = 0
			} else if af.SrcAlpha == 255 && af.DstAlpha == 1 {
				af.DstAlpha = 0
			}
		}
	case len(a) > 0 && a[0] == 'a':
		af.SrcAlpha, af.DstAlpha = 255, 1
	}
	if len(ary) > 1 {
		af.Ex = make([][]float32, 3)
		f, err := strconv.ParseFloat(ary[1], 32)
		if err != nil {
			f = 1
		}
		af.Ex[2] = append(af.Ex[2], float32(f)) // X-Scale
		if len(ary) > 2 {
			f, err := strconv.ParseFloat(ary[2], 32)
			if err != nil {
				f = 1
			}
			af.Ex[2] = append(af.Ex[2], float32(f)) // Y-Scale
			if len(ary) > 3 {
				f, err := strconv.ParseFloat(ary[3], 32)
				if err != nil {
					f = 0
				}
				af.Ex[2] = append(af.Ex[2], float32(f*math.Pi/180)) // Angle
			}
		}
	}
	return af
}
func (af *AnimFrame) Clsn1() []float32 {
	if len(af.Ex) > 0 {
		return af.Ex[0]
	}
	return nil
}
func (af *AnimFrame) Clsn2() []float32 {
	if len(af.Ex) > 1 {
		return af.Ex[1]
	}
	return nil
}

type Animation struct {
	sff       *Sff
	spr       *Sprite
	frames    []AnimFrame
	tile      [4]int32
	loopstart int32
	current   int32
	drawidx   int32
	time      int32
	sumtime   int32
	totaltime int32
	looptime  int32
	nazotime  int32
	mask      int16
	srcAlpha  int16
	dstAlpha  int16
	newframe  bool
	loopend   bool
}

func newAnimation(sff *Sff) *Animation {
	return &Animation{sff: sff, mask: -1, srcAlpha: -1, newframe: true}
}
func ReadAnimation(sff *Sff, lines []string, i *int) *Animation {
	a := newAnimation(sff)
	a.mask = 0
	ols := int32(0)
	var clsn1, clsn1d, clsn2, clsn2d []float32
	def1, def2 := true, true
	for ; *i < len(lines); (*i)++ {
		if len(lines[*i]) > 0 && lines[*i][0] == '[' {
			(*i)--
			break
		}
		line := strings.ToLower(strings.TrimSpace(
			strings.SplitN(lines[*i], ";", 2)[0]))
		af := ReadAnimFrame(line)
		switch {
		case af != nil:
			ols = a.loopstart
			if def1 {
				clsn1 = clsn1d
			}
			if def2 {
				clsn2 = clsn2d
			}
			if len(clsn1) > 0 || len(clsn2) > 0 {
				if len(af.Ex) < 2 {
					af.Ex = make([][]float32, 2)
				}
				af.Ex[0] = clsn1
				af.Ex[1] = clsn2
			}
			a.frames = append(a.frames, *af)
			def1, def2 = true, true
		case len(line) >= 9 && line[:9] == "loopstart":
			a.loopstart = int32(len(a.frames))
		case len(line) >= 5 && line[:4] == "clsn":
			ii := strings.Index(line, ":")
			if ii < 0 {
				break
			}
			size := Atoi(line[ii+1:])
			if size < 0 {
				break
			}
			var clsn []float32
			if line[4] == '1' {
				clsn1 = make([]float32, size*4)
				clsn = clsn1
				if len(line) >= 12 && line[5:12] == "default" {
					clsn1d = clsn1
				}
				def1 = false
			} else if line[4] == '2' {
				clsn2 = make([]float32, size*4)
				clsn = clsn2
				if len(line) >= 12 && line[5:12] == "default" {
					clsn2d = clsn2
				}
				def2 = false
			} else {
				break
			}
			if size == 0 {
				break
			}
			(*i)++
			for n := int32(0); n < size && *i < len(lines); n++ {
				line := strings.ToLower(strings.TrimSpace(
					strings.SplitN(lines[*i], ";", 2)[0]))
				if len(line) == 0 {
					continue
				}
				if len(line) < 4 || line[:4] != "clsn" {
					break
				}
				ii := strings.Index(line, "=")
				if ii < 0 {
					break
				}
				ary := strings.Split(line[ii+1:], ",")
				if len(ary) < 4 {
					break
				}
				l, t, r, b := Atoi(ary[0]), Atoi(ary[1]), Atoi(ary[2]), Atoi(ary[3])
				if l > r {
					l, r = r, l
				}
				if t > b {
					t, b = b, t
				}
				clsn[n*4], clsn[n*4+1], clsn[n*4+2], clsn[n*4+3] =
					float32(l), float32(t), float32(r), float32(b)
				(*i)++
			}
			(*i)--
		}
	}
	if int(a.loopstart) >= len(a.frames) {
		a.loopstart = ols
	}
	if len(a.frames) == 0 {
	} else if a.frames[len(a.frames)-1].Time == -1 {
	} else {
		tmp := int32(0)
		for i := range a.frames {
			if a.frames[i].Time == -1 {
				a.totaltime = 0
				a.looptime = -tmp
				a.nazotime = 0
			}
			a.totaltime += a.frames[i].Time
			if i < int(a.loopstart) {
				a.nazotime += a.frames[i].Time
				tmp += a.frames[i].Time
			} else {
				a.looptime += a.frames[i].Time
			}
		}
		if a.totaltime == -1 {
			a.nazotime = 0
		}
	}
	return a
}
func ReadAction(sff *Sff, lines []string, i *int) (no int32, a *Animation) {
	var name, subname string
	for ; *i < len(lines); (*i)++ {
		name, subname = SectionName(lines[*i])
		if len(name) > 0 {
			break
		}
	}
	if name != "begin " {
		return
	}
	spi := strings.Index(subname, " ")
	if spi < 0 {
		return
	}
	if strings.ToLower(subname[:spi+1]) != "action " {
		return
	}
	(*i)++
	return Atoi(subname[spi+1:]), ReadAnimation(sff, lines, i)
}
func (a *Animation) Reset() {
	a.current, a.drawidx = 0, 0
	a.time, a.sumtime = 0, 0
	a.newframe, a.loopend = false, false
	a.spr = nil
}
func (a *Animation) AnimTime() int32 {
	return a.sumtime - a.totaltime
}
func (a *Animation) AnimElemTime(elem int32) int32 {
	if int(elem) > len(a.frames) {
		t := a.AnimTime()
		if t > 0 {
			t = 0
		}
		return t
	}
	e, t := Max(0, elem)-1, a.sumtime
	for i := int32(0); i < e; i++ {
		t -= Max(0, a.frames[i].Time)
	}
	return t
}
func (a *Animation) curFrame() *AnimFrame {
	return &a.frames[a.current]
}
func (a *Animation) CurrentFrame() *AnimFrame {
	if len(a.frames) == 0 {
		return nil
	}
	return a.curFrame()
}
func (a *Animation) animSeek(elem int32) {
	if elem < 0 {
		elem = 0
	}
	foo := true
	for {
		a.current = elem
		for a.curFrame().Time <= 0 && int(a.current) < len(a.frames) {
			if int(a.current) == len(a.frames)-1 && a.curFrame().Time == -1 {
				break
			}
			a.current++
		}
		if int(a.current) < len(a.frames) {
			break
		}
		foo = !foo
		if foo {
			a.current = int32(len(a.frames) - 1)
			break
		}
	}
	if a.current < 0 {
		a.current = 0
	} else if int(a.current) >= len(a.frames) {
		a.current = int32(len(a.frames) - 1)
	}
}
func (a *Animation) UpdateSprite() {
	if len(a.frames) == 0 {
		return
	}
	if a.totaltime > 0 {
		if a.sumtime >= a.totaltime {
			a.time, a.newframe, a.current = 0, true, a.loopstart
		}
		a.animSeek(a.current)
		if a.nazotime < 0 && a.sumtime >= a.totaltime+a.nazotime &&
			a.sumtime >= a.totaltime-a.looptime &&
			(a.sumtime == a.totaltime+a.nazotime ||
				a.sumtime == a.totaltime-a.looptime) {
			a.time, a.newframe, a.current = 0, true, 0
		}
	}
	if a.newframe && a.sff != nil {
		a.spr = a.sff.GetSprite(a.curFrame().Group, a.curFrame().Number)
	}
	a.newframe, a.drawidx = false, a.current
}
func (a *Animation) Action() {
	if len(a.frames) == 0 {
		a.loopend = true
		return
	}
	a.UpdateSprite()
	next := func() {
		if a.totaltime != -1 || int(a.current) < len(a.frames)-1 {
			a.time = 0
			a.newframe = true
			for {
				a.current++
				if a.totaltime == -1 && int(a.current) == len(a.frames)-1 ||
					int(a.current) >= len(a.frames) || a.curFrame().Time > 0 {
					break
				}
			}
		}
	}
	curFrameTime := a.curFrame().Time
	if curFrameTime <= 0 {
		next()
	}
	if int(a.current) < len(a.frames) {
		a.time++
		if a.time >= curFrameTime {
			next()
			if int(a.current) >= len(a.frames) {
				a.current = a.loopstart
			}
		}
	} else {
		a.current = a.loopstart
	}
	if a.totaltime != -1 && a.sumtime >= a.totaltime {
		a.sumtime = a.totaltime - a.looptime
	}
	a.sumtime++
	if a.totaltime != -1 && a.sumtime >= a.totaltime {
		a.loopend = true
	}
}
func (a *Animation) alpha() int32 {
	var sa, da byte
	if a.srcAlpha >= 0 {
		sa = byte(a.srcAlpha)
		if a.dstAlpha < 0 {
			da = byte((^a.dstAlpha + int16(a.frames[a.drawidx].DstAlpha)) >> 1)
		} else {
			da = byte(a.dstAlpha)
		}
	} else {
		sa = a.frames[a.drawidx].SrcAlpha
		da = a.frames[a.drawidx].DstAlpha
		if sa == 255 && da == 1 {
			da = 255
		}
	}
	if sa == 1 && da == 255 {
		return -2
	}
	sa = byte(int32(sa) * sys.brightness >> 8)
	if sa < 5 && da == 255 {
		return 0
	}
	if sa == 255 && da == 255 {
		return -1
	}
	trans := int32(sa)
	if int(sa)+int(da) < 254 || 256 < int(sa)+int(da) {
		trans |= int32(da)<<10 | 1<<9
	}
	return trans
}
func (a *Animation) pal(pfx *PalFX) (p []uint32) {
	if pfx != nil && len(pfx.remap) > 0 {
		a.sff.palList.SwapPalMap(&pfx.remap)
	}
	p = a.spr.GetPal(&a.sff.palList)
	if pfx != nil && len(pfx.remap) > 0 {
		a.sff.palList.SwapPalMap(&pfx.remap)
	}
	if len(p) == 0 {
		return
	}
	return
}
func (a *Animation) drawSub1(angle float32) (h, v, agl float32) {
	h, v = float32(a.frames[a.drawidx].H), float32(a.frames[a.drawidx].V)
	agl = float32(float64(angle) * math.Pi / 180)
	if len(a.frames[a.drawidx].Ex) > 2 {
		if len(a.frames[a.drawidx].Ex[2]) > 0 {
			h *= a.frames[a.drawidx].Ex[2][0]
			if len(a.frames[a.drawidx].Ex[2]) > 1 {
				v *= a.frames[a.drawidx].Ex[2][1]
				if len(a.frames[a.drawidx].Ex[2]) > 2 {
					agl += a.frames[a.drawidx].Ex[2][2]
				}
			}
		}
	}
	return
}
func (a *Animation) Draw(window *[4]int32, x, y, xcs, ycs, xs, xbs, ys,
	rxadd, angle, rcx float32, pfx *PalFX, old bool) {
	if a.spr == nil || a.spr.Tex == nil {
		return
	}
	h, v, angle := a.drawSub1(angle)
	xs *= xcs
	ys *= ycs
	if (xs < 0) != (ys < 0) {
		angle *= -1
	}
	xs *= h
	ys *= v
	x = xcs*x + xs*float32(a.frames[a.drawidx].X)
	y = ycs*y + ys*float32(a.frames[a.drawidx].Y)
	var rcy float32
	if angle == 0 {
		if xs < 0 {
			x *= -1
			if old {
				x += xs
			}
		}
		if ys < 0 {
			y *= -1
			if old {
				y += ys
			}
		}
		if a.tile[2] == 1 {
			tmp := xs * float32(a.tile[0])
			if a.tile[0] <= 0 {
				tmp += xs * float32(a.spr.Size[0])
			}
			if tmp != 0 {
				x -= float32(int(x/tmp)) * tmp
			}
		}
		if a.tile[3] == 1 {
			tmp := ys * float32(a.tile[1])
			if a.tile[1] <= 0 {
				tmp += ys * float32(a.spr.Size[1])
			}
			if tmp != 0 {
				y -= float32(int(y/tmp)) * tmp
			}
		}
		rcx, rcy = rcx*sys.widthScale, 0
		x = -x + AbsF(xs)*float32(a.spr.Offset[0])
		y = -y + AbsF(ys)*float32(a.spr.Offset[1])
	} else {
		rcx, rcy = (x+rcx)*sys.widthScale, y*sys.heightScale
		x, y = AbsF(xs)*float32(a.spr.Offset[0]), AbsF(ys)*float32(a.spr.Offset[1])
	}
	a.spr.glDraw(a.pal(pfx), int32(a.mask), x*sys.widthScale, y*sys.heightScale,
		&a.tile, xs*sys.widthScale, xcs*xbs*h*sys.widthScale, ys*sys.heightScale,
		xcs*rxadd*sys.widthScale/sys.heightScale, angle, a.alpha(), window,
		rcx, rcy, pfx)
}
func (a *Animation) ShadowDraw(x, y, xscl, yscl, vscl, angle float32,
	pfx *PalFX, old bool, color uint32, alpha int32) {
	if a.spr == nil || a.spr.Tex == nil {
		return
	}
	h, v, angle := a.drawSub1(angle)
	x += xscl * h * float32(a.frames[a.drawidx].X)
	y += yscl * vscl * v * float32(a.frames[a.drawidx].Y)
	if (xscl < 0) != (yscl < 0) {
		angle *= -1
	}
	var draw func(int32)
	if a.spr.rle == -12 {
		draw = func(trans int32) {
			RenderMugenFcS(*a.spr.Tex, a.spr.Size,
				AbsF(xscl*h)*float32(a.spr.Offset[0])*sys.widthScale,
				AbsF(yscl*v)*float32(a.spr.Offset[1])*sys.heightScale, &a.tile,
				xscl*h*sys.widthScale, xscl*h*sys.widthScale,
				yscl*v*sys.heightScale, vscl, 0, angle, trans, &sys.scrrect,
				(x+float32(sys.gameWidth)/2)*sys.widthScale, y*sys.heightScale, color)
		}
	} else {
		draw = func(trans int32) {
			var pal [256]uint32
			RenderMugen(*a.spr.Tex, pal[:], int32(a.mask), a.spr.Size,
				AbsF(xscl*h)*float32(a.spr.Offset[0])*sys.widthScale,
				AbsF(yscl*v)*float32(a.spr.Offset[1])*sys.heightScale, &a.tile,
				xscl*h*sys.widthScale, xscl*h*sys.widthScale,
				yscl*v*sys.heightScale, vscl, 0, angle, trans, &sys.scrrect,
				(x+float32(sys.gameWidth)/2)*sys.widthScale, y*sys.heightScale)
		}
	}
	if int32(color) > 0 {
		draw(-2)
	}
	if alpha > 0 {
		draw((256-alpha)<<10 | 1<<9)
	}
}

type AnimationTable map[int32]*Animation

func NewAnimationTable() AnimationTable {
	return AnimationTable(make(map[int32]*Animation))
}
func (at AnimationTable) readAction(sff *Sff,
	lines []string, i *int) *Animation {
	for *i < len(lines) {
		no, a := ReadAction(sff, lines, i)
		if a != nil {
			if tmp := at[no]; tmp != nil {
				return tmp
			}
			at[no] = a
			for len(a.frames) == 0 {
				a2 := at.readAction(sff, lines, i)
				if a2 != nil {
					*a = *a2
				}
			}
			return a
		} else {
			(*i)++
		}
	}
	return nil
}
func ReadAnimationTable(sff *Sff, lines []string, i *int) AnimationTable {
	at := NewAnimationTable()
	for at.readAction(sff, lines, i) != nil {
	}
	return at
}
func (at AnimationTable) get(no int32) *Animation {
	a := at[no]
	if a == nil {
		return a
	}
	ret := &Animation{}
	*ret = *a
	return ret
}

type SprData struct {
	anim     *Animation
	fx       *PalFX
	pos      [2]float32
	scl      [2]float32
	alpha    [2]int32
	priority int32
	angle    float32
	ascl     [2]float32
	screen   bool
	bright   bool
	oldVer   bool
}
type DrawList []*SprData

func (dl *DrawList) add(sd *SprData, sc, salp int32, so float32) {
	if sys.frameSkip || sd.anim == nil || sd.anim.spr == nil {
		return
	}
	if sd.angle != 0 {
		for i, as := range sd.ascl {
			sd.scl[i] *= as
		}
	}
	i, start := 0, 0
	for l := len(*dl); l > 0; {
		i := start + l>>1
		if sd.priority <= (*dl)[i].priority {
			l = i - start
		} else if i == start {
			i++
			l = 0
		} else {
			l -= i - start
			start = i
		}
	}
	*dl = append(*dl, nil)
	copy((*dl)[i+1:], (*dl)[i:])
	(*dl)[i] = sd
	if sc != 0 {
		sys.shadows.add(&ShadowSprite{sd, sc, salp, so})
	}
}
func (dl DrawList) draw(x, y, scl float32) {
	for _, s := range dl {
		s.anim.srcAlpha, s.anim.dstAlpha = int16(s.alpha[0]), int16(s.alpha[1])
		ob := sys.brightness
		if s.bright {
			sys.brightness = 256
		}
		var p [2]float32
		cs := scl
		if s.screen {
			p = [...]float32{s.pos[0], s.pos[1] + float32(sys.gameHeight-240)}
			cs = 1
		} else {
			p = [...]float32{sys.cam.Offset[0]/cs - (x - s.pos[0]),
				(sys.cam.GroundLevel()+sys.cam.Offset[1]-sys.envShake.getOffset())/cs -
					(y - s.pos[1])}
		}
		s.anim.Draw(&sys.scrrect, p[0], p[1], cs, cs, s.scl[0], s.scl[0],
			s.scl[1], 0, s.angle, float32(sys.gameWidth)/2, s.fx, s.oldVer)
		sys.brightness = ob
	}
}

type ShadowSprite struct {
	*SprData
	shadowColor int32
	shadowAlpha int32
	offsetY     float32
}
type ShadowList []*ShadowSprite

func (sl *ShadowList) add(ss *ShadowSprite) {
	i, start := 0, 0
	for l := len(*sl); l > 0; {
		i := start + l>>1
		if ss.priority <= (*sl)[i].priority {
			l = i - start
		} else if i == start {
			i++
			l = 0
		} else {
			l -= i - start
			start = i
		}
	}
	*sl = append(*sl, nil)
	copy((*sl)[i+1:], (*sl)[i:])
	(*sl)[i] = ss
}
func (sl ShadowList) draw(x, y, scl float32) {
	for _, s := range sl {
		intensity := sys.stage.sdw.intensity
		color, alpha := s.shadowColor, s.shadowAlpha
		fend := float32(sys.stage.sdw.fadeend) * sys.stage.localscl
		if s.pos[1] < fend {
			continue
		}
		fbgn := float32(sys.stage.sdw.fadebgn) * sys.stage.localscl
		if s.pos[1] < fbgn {
			alpha = int32(float32(alpha) * (fend - s.pos[1]) / (fend - fbgn))
		}
		comm := true
		if color < 0 {
			color = int32(sys.stage.sdw.color)
			if alpha < 255 {
				intensity = intensity * alpha >> 8
			} else {
				comm = false
			}
		} else {
			intensity = 0
		}
		if comm {
			color = color&0xff*alpha>>8&0xff | color&0xff00*alpha>>8&0xff00 |
				color&0xff0000*alpha>>8&0xff0000
		}
		s.anim.ShadowDraw(sys.cam.Offset[0]-(x-s.pos[0])*scl,
			sys.cam.GroundLevel()+sys.cam.Offset[1]-sys.envShake.getOffset()-
				(y+s.pos[1]*sys.stage.sdw.yscale-s.offsetY)*scl,
			scl*s.scl[0], scl*-s.scl[1], sys.stage.sdw.yscale, s.angle, &sys.bgPalFX,
			s.oldVer, uint32(color), intensity)
	}
}
func (sl ShadowList) drawReflection(x, y, scl float32) {
	for _, s := range sl {
		if s.alpha[0] < 0 {
			s.alpha[0] = int32(s.anim.frames[s.anim.drawidx].SrcAlpha)
			s.alpha[1] = int32(s.anim.frames[s.anim.drawidx].DstAlpha)
			if s.alpha[0] == 255 && s.alpha[1] == 1 {
				s.alpha[1] = 255
			}
		}
		ref := sys.stage.reflection * s.shadowAlpha >> 8
		s.alpha[0] = int32(float32(s.alpha[0]*ref) / 255)
		if s.alpha[1] < 0 {
			s.alpha[1] = 128
		}
		s.alpha[1] = Min(255, s.alpha[1]+255-ref)
		if s.alpha[0] == 1 && s.alpha[1] == 255 {
			s.alpha[0] = 0
		}
		s.anim.Draw(&sys.scrrect, sys.cam.Offset[0]/scl-(x-s.pos[0]),
			(sys.cam.GroundLevel()+sys.cam.Offset[1]-sys.envShake.getOffset())/scl-
				(y+s.pos[1]-s.offsetY), scl, scl, s.scl[0], s.scl[0], -s.scl[1], 0,
			s.angle, float32(sys.gameWidth)/2, s.fx, s.oldVer)
	}
}

type Anim struct {
	anim             *Animation
	window           [4]int32
	x, y, xscl, yscl float32
}

func NewAnim(sff *Sff, action string) *Anim {
	lines, i := SplitAndTrim(action, "\n"), 0
	a := &Anim{anim: ReadAnimation(sff, lines, &i),
		window: sys.scrrect, xscl: 1, yscl: 1}
	if len(a.anim.frames) == 0 {
		return nil
	}
	return a
}
func (a *Anim) SetPos(x, y float32) {
	a.x, a.y = x, y
}
func (a *Anim) AddPos(x, y float32) {
	a.x += x
	a.y += y
}
func (a *Anim) SetTile(x, y int32) {
	a.anim.tile[2], a.anim.tile[3] = x, y
}
func (a *Anim) SetColorKey(mask int16) {
	a.anim.mask = mask
}
func (a *Anim) SetAlpha(src, dst int16) {
	a.anim.srcAlpha, a.anim.dstAlpha = src, dst
}
func (a *Anim) SetScale(x, y float32) {
	a.xscl, a.yscl = x, y
}
func (a *Anim) SetWindow(x, y, w, h float32) {
	a.window[0] = int32((x + float32(sys.gameWidth-320)/2) * sys.widthScale)
	a.window[1] = int32((y + float32(sys.gameHeight-240)) * sys.heightScale)
	a.window[2] = int32(w*sys.widthScale + 0.5)
	a.window[3] = int32(h*sys.heightScale + 0.5)
}
func (a *Anim) Update() {
	a.anim.Action()
}
func (a *Anim) Draw() {
	if !sys.frameSkip {
		a.anim.Draw(&a.window, a.x+float32(sys.gameWidth-320)/2,
			a.y+float32(sys.gameHeight-240), 1, 1, a.xscl, a.xscl, a.yscl,
			0, 0, 0, nil, false)
	}
}
