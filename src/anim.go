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
	Srcalpha      byte
	Dstalpha      byte
	H, V          int8
	Ex            [][]float32
}

func newAnimFrame() *AnimFrame {
	return &AnimFrame{Time: -1, Group: -1, Srcalpha: 255, H: 1, V: 1}
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
		af.Srcalpha, af.Dstalpha = 255, 128
	case len(a) > 0 && a[0] == 's':
		af.Srcalpha, af.Dstalpha = 1, 255
	case len(a) >= 2 && a[:2] == "as":
		i := strings.IndexAny(a, "d")
		if i >= 0 {
			sa := Atoi(a[2:i])
			if sa <= 0 {
				af.Srcalpha = 0
			} else if sa >= 255 {
				af.Srcalpha = 255
			} else {
				af.Srcalpha = byte(sa)
			}
			da := Atoi(a[i+1:])
			if da <= 0 {
				af.Dstalpha = 0
			} else if da >= 255 {
				af.Dstalpha = 255
			} else {
				af.Dstalpha = byte(da)
			}
			if af.Srcalpha == 1 && af.Dstalpha == 255 {
				af.Srcalpha = 0
			}
		}
	case len(a) > 0 && a[0] == 'a':
		af.Srcalpha, af.Dstalpha = 255, 255
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
	srcalpha  int16
	dstalpha  int16
	newframe  bool
	loopend   bool
}

func newAnimation(sff *Sff) *Animation {
	return &Animation{sff: sff, mask: -1, srcalpha: -1, newframe: true}
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
			AppendAF(&a.frames, *af)
			def1, def2 = true, true
		case len(line) >= 9 && line[:9] == "loopstart":
			a.loopstart = int32(len(a.frames))
		case len(line) >= 4 && line[:4] == "clsn":
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
			n := int32(0)
			for ; *i < len(lines); (*i)++ {
				if (n+1)*4 > size {
					break
				}
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
				n++
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
func (a *Animation) AnimTime() int32 {
	return a.sumtime - a.totaltime
}
func (a *Animation) CurFrame() *AnimFrame {
	return &a.frames[a.current]
}
func (a *Animation) animSeek(elem int32) {
	if elem < 0 {
		elem = 0
	}
	foo := true
	for {
		a.current = elem
		for a.CurFrame().Time <= 0 && int(a.current) < len(a.frames) {
			if int(a.current) == len(a.frames)-1 && a.CurFrame().Time == -1 {
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
		a.spr = a.sff.GetSprite(a.CurFrame().Group, a.CurFrame().Number)
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
					int(a.current) >= len(a.frames) || a.CurFrame().Time > 0 {
					break
				}
			}
		}
	}
	curFrameTime := a.CurFrame().Time
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
	if a.srcalpha >= 0 {
		sa = byte(a.srcalpha)
		da = byte(a.dstalpha)
	} else {
		sa = a.frames[a.drawidx].Srcalpha
		da = a.frames[a.drawidx].Dstalpha
	}
	if sa == 1 && da == 255 {
		return -2
	}
	sa = byte(int32(sa) * brightness >> 8)
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
	if pfx != nil && len(pfx.Remap) > 0 {
		a.sff.palList.SwapPalMap(&pfx.Remap)
	}
	p = a.spr.GetPal(&a.sff.palList)
	if pfx != nil && len(pfx.Remap) > 0 {
		a.sff.palList.SwapPalMap(&pfx.Remap)
	}
	if len(p) == 0 {
		return
	}
	return
}
func (a *Animation) Draw(window *[4]int32, x, y, xcs, ycs, xs, xbs, ys,
	rxadd, angle, rcx float32, pfx *PalFX, old bool) {
	if a.spr == nil || a.spr.Tex == nil {
		return
	}
	h, v := float32(a.frames[a.drawidx].H), float32(a.frames[a.drawidx].V)
	if len(a.frames[a.drawidx].Ex) > 2 {
		if len(a.frames[a.drawidx].Ex[2]) > 0 {
			h *= a.frames[a.drawidx].Ex[2][0]
			if len(a.frames[a.drawidx].Ex[2]) > 1 {
				v *= a.frames[a.drawidx].Ex[2][1]
				if len(a.frames[a.drawidx].Ex[2]) > 2 {
					angle += a.frames[a.drawidx].Ex[2][2]
				}
			}
		}
	}
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
		rcx, rcy = rcx*widthScale, 0
		x, y = -x+xs*float32(a.spr.Offset[0]), -y+ys*float32(a.spr.Offset[1])
	} else {
		rcx, rcy = (x+rcx)*widthScale, y*heightScale
		x, y = AbsF(xs)*float32(a.spr.Offset[0]), AbsF(ys)*float32(a.spr.Offset[1])
	}
	a.spr.glDraw(a.pal(pfx), int32(a.mask), x*widthScale, y*heightScale,
		&a.tile, xs*widthScale, xcs*xbs*h*widthScale, ys*heightScale,
		xcs*rxadd*widthScale/heightScale, angle, a.alpha(), window,
		rcx, rcy, pfx)
}

type Anim struct {
	anim             *Animation
	window           [4]int32
	x, y, xscl, yscl float32
}

func NewAnim(sff *Sff, action string) *Anim {
	lines, i := SplitAndTrim(action, "\n"), 0
	a := &Anim{anim: ReadAnimation(sff, lines, &i),
		window: scrrect, xscl: 1, yscl: 1}
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
	a.anim.srcalpha, a.anim.dstalpha = src, dst
}
func (a *Anim) SetScale(x, y float32) {
	a.xscl, a.yscl = x, y
}
func (a *Anim) SetWindow(x, y, w, h float32) {
	a.window[0] = int32((x + float32(gameWidth-320)/2) * widthScale)
	a.window[1] = int32((y + float32(gameHeight-240)) * heightScale)
	a.window[2] = int32(w*widthScale + 0.5)
	a.window[3] = int32(h*heightScale + 0.5)
}
func (a *Anim) Update() {
	a.anim.Action()
}
func (a *Anim) Draw() {
	if !frameSkip {
		a.anim.Draw(&a.window, a.x+float32(gameWidth-320)/2,
			a.y+float32(gameHeight-240), 1, 1, a.xscl, a.xscl, a.yscl,
			0, 0, 0, nil, false)
	}
}
