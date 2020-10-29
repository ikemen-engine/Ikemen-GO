package main

import (
	"strings"
)

// AnimFrame holds frame data, used in animation tables.
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
	ary := strings.SplitN(line, ",", 10)
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
	a := strings.ToLower(SplitAndTrim(ary[6], ",")[0])
	switch {
	case a == "a1":
		af.SrcAlpha, af.DstAlpha = 255, 128
	case len(a) > 0 && a[0] == 's':
		af.SrcAlpha, af.DstAlpha = 1, 255
	case len(a) >= 2 && a[:2] == "as":
		if len(a) > 2 && a[2] >= '0' && a[2] <= '9' {
			i, alp := 2, 0
			for ; i < len(a) && a[i] >= '0' && a[i] <= '9'; i++ {
				alp = alp*10 + int(a[i]-'0')
			}
			alp &= 0x3fff
			if alp >= 255 {
				af.SrcAlpha = 255
			} else {
				af.SrcAlpha = byte(alp)
			}
			if i < len(a) && a[i] == 'd' {
				i++
				if i < len(a) && a[i] >= '0' && a[i] <= '9' {
					alp = 0
					for ; i < len(a) && a[i] >= '0' && a[i] <= '9'; i++ {
						alp = alp*10 + int(a[i]-'0')
					}
					alp &= 0x3fff
					if alp >= 255 {
						af.DstAlpha = 255
					} else {
						af.DstAlpha = byte(alp)
					}
					if af.SrcAlpha == 1 && af.DstAlpha == 255 {
						af.SrcAlpha = 0
					}
				}
			}
		}
	case len(a) > 0 && a[0] == 'a':
		af.SrcAlpha, af.DstAlpha = 255, 255
	}
	if len(ary) < 8 {
		return af
	}
	af.Ex = make([][]float32, 3)
	var f float32
	if f = float32(Atof(ary[7])); f == 0 {
		f = 1
	}
	af.Ex[2] = append(af.Ex[2], f) // X-Scale
	if len(ary) < 9 {
		return af
	}
	if f = float32(Atof(ary[8])); f == 0 {
		f = 1
	}
	af.Ex[2] = append(af.Ex[2], f) // Y-Scale
	if len(ary) < 10 {
		return af
	}
	f = float32(Atof(ary[9]))
	af.Ex[2] = append(af.Ex[2], f) // Angle
	if len(ary) < 11 {
		return af
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
	sff                *Sff
	spr                *Sprite
	frames             []AnimFrame
	tile               [4]int32
	loopstart          int32
	interpolate_offset []int32
	interpolate_scale  []int32
	interpolate_angle  []int32
	interpolate_blend  []int32
	// Current frame
	current                    int32
	drawidx                    int32
	time                       int32
	sumtime                    int32
	totaltime                  int32
	looptime                   int32
	nazotime                   int32
	mask                       int16
	srcAlpha                   int16
	dstAlpha                   int16
	newframe                   bool
	loopend                    bool
	interpolate_offset_x       float32
	interpolate_offset_y       float32
	scale_x                    float32
	scale_y                    float32
	angle                      float32
	interpolate_blend_srcalpha float32
	interpolate_blend_dstalpha float32
	remap                      RemapPreset
	start_scale                [2]float32
}

func newAnimation(sff *Sff) *Animation {
	return &Animation{sff: sff, mask: -1, srcAlpha: -1, newframe: true,
		remap: make(RemapPreset), start_scale: [...]float32{1, 1}}
}
func ReadAnimation(sff *Sff, lines []string, i *int) *Animation {
	a := newAnimation(sff)
	a.mask = 0
	ols := int32(0)
	var clsn1, clsn1d, clsn2, clsn2d []float32
	def1, def2 := true, true
	for ; *i < len(lines); (*i)++ {
		if len(lines[*i]) > 0 && lines[*i][0] == '[' {
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
		case len(line) >= 18 && line[:18] == "interpolate offset":
			a.interpolate_offset = append(a.interpolate_offset, int32(len(a.frames)))
		case len(line) >= 17 && line[:17] == "interpolate scale":
			a.interpolate_scale = append(a.interpolate_scale, int32(len(a.frames)))
		case len(line) >= 17 && line[:17] == "interpolate angle":
			a.interpolate_angle = append(a.interpolate_angle, int32(len(a.frames)))
		case len(line) >= 17 && line[:17] == "interpolate blend":
			a.interpolate_blend = append(a.interpolate_blend, int32(len(a.frames)))
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
		a.totaltime = -1
	} else {
		tmp := int32(0)
		for i, f := range a.frames {
			if f.Time == -1 {
				a.totaltime = 0
				a.looptime = -tmp
				a.nazotime = 0
			}
			a.totaltime += f.Time
			if i < int(a.loopstart) {
				a.nazotime += f.Time
				tmp += f.Time
			} else {
				a.looptime += f.Time
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
	a.newframe, a.loopend = true, false
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
func (a *Animation) AnimElemNo(time int32) int32 {
	if len(a.frames) > 0 {
		i, oldt := a.current, int32(0)
		if time <= 0 {
			time += a.time
			loop := false
			for {
				if time >= 0 {
					return i + 1
				}
				i--
				if i < 0 || a.current >= a.loopstart && i < a.loopstart {
					if time == oldt {
						break
					}
					oldt = time
					loop = true
					i = int32(len(a.frames)) - 1
				}
				time += Max(0, a.frames[i].Time)
				if loop && i == int32(len(a.frames))-1 && a.frames[i].Time == -1 {
					return i + 1
				}
			}
		} else {
			time += a.time
			for {
				time -= Max(0, a.frames[i].Time)
				if time < 0 || i == int32(len(a.frames))-1 && a.frames[i].Time == -1 {
					return i + 1
				}
				i++
				if i >= int32(len(a.frames)) {
					if time == oldt {
						break
					}
					oldt = time
					i = a.loopstart
				}
			}
		}
	}
	return int32(len(a.frames))
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
func (a *Animation) drawFrame() *AnimFrame {
	if len(a.frames) == 0 {
		return nil
	}
	return &a.frames[a.drawidx]
}
func (a *Animation) SetAnimElem(elem int32) {
	a.current = Max(0, elem-1)
	if int(a.current) >= len(a.frames) {
		if a.totaltime == -1 {
			a.current = int32(len(a.frames)) - 1
		} else {
			a.current = a.loopstart +
				(a.current-a.loopstart)%(int32(len(a.frames))-a.loopstart)
		}
	}
	a.drawidx, a.time, a.newframe = a.current, 0, true
	a.UpdateSprite()
	a.loopend = false
	a.sumtime = 0 // AnimElemTime 内で使用
	a.sumtime = -a.AnimElemTime(a.current + 1)
}
func (a *Animation) animSeek(elem int32) {
	if elem < 0 {
		elem = 0
	}
	foo := true
	for {
		a.current = elem
		for int(a.current) < len(a.frames) && a.curFrame().Time <= 0 {
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
	if a.newframe && a.sff != nil && a.frames[a.current].Time != 0 {
		group, number := a.curFrame().Group, a.curFrame().Number
		if mg, ok := a.remap[group]; ok {
			if mn, ok := mg[number]; ok {
				group, number = mn[0], mn[1]
			}
		}
		a.spr = a.sff.GetSprite(group, number)
	}
	a.newframe, a.drawidx = false, a.current

	a.scale_x = a.start_scale[0]
	a.scale_y = a.start_scale[1]
	a.angle = 0
	a.interpolate_offset_x = 0
	a.interpolate_offset_y = 0
	a.interpolate_blend_srcalpha = float32(a.frames[a.drawidx].SrcAlpha)
	a.interpolate_blend_dstalpha = float32(a.frames[a.drawidx].DstAlpha)

	if len(a.frames[a.drawidx].Ex) > 2 {
		if len(a.frames[a.drawidx].Ex[2]) > 0 {
			a.scale_x *= a.frames[a.drawidx].Ex[2][0]
			if len(a.frames[a.drawidx].Ex[2]) > 1 {
				a.scale_y *= a.frames[a.drawidx].Ex[2][1]
				if len(a.frames[a.drawidx].Ex[2]) > 2 {
					a.angle = a.frames[a.drawidx].Ex[2][2]
				}
			}
		}
	}
	nextDrawidx := a.drawidx + 1
	if int(a.drawidx) >= len(a.frames)-1 {
		nextDrawidx = a.loopstart
	}
	for _, i := range a.interpolate_offset {
		if nextDrawidx == i {
			a.interpolate_offset_x = float32(a.frames[nextDrawidx].X-a.frames[a.drawidx].X) / float32(a.curFrame().Time) * float32(a.time)
			a.interpolate_offset_y = float32(a.frames[nextDrawidx].Y-a.frames[a.drawidx].Y) / float32(a.curFrame().Time) * float32(a.time)
			break
		}
	}
	for _, i := range a.interpolate_scale {
		if nextDrawidx == i {
			var drawframe_scale_x, nextframe_scale_x, drawframe_scale_y, nextframe_scale_y float32 = 1, 1, 1, 1
			if len(a.frames[a.drawidx].Ex) > 2 {
				if len(a.frames[a.drawidx].Ex[2]) > 0 {
					drawframe_scale_x = a.frames[a.drawidx].Ex[2][0]
				}
				if len(a.frames[a.drawidx].Ex[2]) > 1 {
					drawframe_scale_y = a.frames[a.drawidx].Ex[2][1]
				}
			}
			if len(a.frames[nextDrawidx].Ex) > 2 {
				if len(a.frames[nextDrawidx].Ex[2]) > 0 {
					nextframe_scale_x = a.frames[nextDrawidx].Ex[2][0]
				}
				if len(a.frames[nextDrawidx].Ex[2]) > 1 {
					nextframe_scale_y = a.frames[nextDrawidx].Ex[2][1]
				}
			}
			a.scale_x += (nextframe_scale_x - drawframe_scale_x) / float32(a.curFrame().Time) * float32(a.time)
			a.scale_y += (nextframe_scale_y - drawframe_scale_y) / float32(a.curFrame().Time) * float32(a.time)
			break
		}
	}
	for _, i := range a.interpolate_angle {
		if nextDrawidx == i {
			var drawframe_angle, nextframe_angle float32 = 0, 0
			if len(a.frames[a.drawidx].Ex) > 2 {
				if len(a.frames[a.drawidx].Ex[2]) > 2 {
					drawframe_angle = a.frames[a.drawidx].Ex[2][2]
				}
			}
			if len(a.frames[nextDrawidx].Ex) > 2 {
				if len(a.frames[nextDrawidx].Ex[2]) > 2 {
					nextframe_angle = a.frames[nextDrawidx].Ex[2][2]
				}
			}
			a.angle += (nextframe_angle - drawframe_angle) / float32(a.curFrame().Time) * float32(a.time)
			break
		}
	}
	if byte(a.interpolate_blend_srcalpha) != 1 ||
		byte(a.interpolate_blend_dstalpha) != 255 {
		for _, i := range a.interpolate_blend {
			if nextDrawidx == i {
				a.interpolate_blend_srcalpha += (float32(a.frames[nextDrawidx].SrcAlpha) - a.interpolate_blend_srcalpha) / float32(a.curFrame().Time) * float32(a.time)
				a.interpolate_blend_dstalpha += (float32(a.frames[nextDrawidx].DstAlpha) - a.interpolate_blend_dstalpha) / float32(a.curFrame().Time) * float32(a.time)
				if byte(a.interpolate_blend_srcalpha) == 1 && byte(a.interpolate_blend_dstalpha) == 255 {
					a.interpolate_blend_srcalpha = 0
				}
				break
			}
		}
	}
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
	if a.curFrame().Time <= 0 {
		next()
	}
	if int(a.current) < len(a.frames) {
		a.time++
		if a.time >= a.curFrame().Time {
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
			da = byte(^a.dstAlpha >> 1)
			if sa == 1 && da == 255 {
				sa = 0
			}
		} else {
			da = byte(a.dstAlpha)
		}
	} else {
		sa = byte(a.interpolate_blend_srcalpha)
		da = byte(a.interpolate_blend_dstalpha)
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
func (a *Animation) pal(pfx *PalFX, neg bool) (p []uint32, plt *Texture) {
	if pfx != nil && len(pfx.remap) > 0 {
		a.sff.palList.SwapPalMap(&pfx.remap)
	}
	p = a.spr.GetPal(&a.sff.palList)
	plt = a.spr.GetPalTex(&a.sff.palList)
	if pfx != nil && len(pfx.remap) > 0 {
		a.sff.palList.SwapPalMap(&pfx.remap)
	}
	return
}
func (a *Animation) drawSub1(angle, facing float32) (h, v, agl float32) {
	h, v = float32(a.frames[a.drawidx].H), float32(a.frames[a.drawidx].V)
	agl = angle
	h *= a.scale_x
	v *= a.scale_y
	agl += a.angle * facing
	return
}
func (a *Animation) Draw(window *[4]int32, x, y, xcs, ycs, xs, xbs, ys,
	rxadd, angle, yangle, xangle, rcx float32, pfx *PalFX, old bool, facing float32, isReflection bool, posLocalscl float32) {
	if a.spr == nil || a.spr.Tex == nil {
		return
	}
	h, v, angle := a.drawSub1(angle, facing)
	if isReflection {
		angle = -angle
	}
	xs *= xcs * h
	ys *= ycs * v
	x = xcs*x + xs*posLocalscl*(float32(a.frames[a.drawidx].X)+a.interpolate_offset_x)*(1/a.scale_x)
	y = ycs*y + ys*posLocalscl*(float32(a.frames[a.drawidx].Y)+a.interpolate_offset_y)*(1/a.scale_y)
	var rcy float32
	if angle == 0 && yangle == 0 && xangle == 0 {
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
	trans := a.alpha()
	pal, paltex := a.pal(pfx, trans == -2)
	a.spr.glDraw(pal, int32(a.mask), x*sys.widthScale,
		y*sys.heightScale, &a.tile, xs*sys.widthScale, xcs*xbs*h*sys.widthScale,
		ys*sys.heightScale, xcs*rxadd*sys.widthScale/sys.heightScale, angle, yangle, xangle,
		trans, window, rcx, rcy, pfx, paltex)
}
func (a *Animation) ShadowDraw(x, y, xscl, yscl, vscl, angle, yangle, xangle float32,
	pfx *PalFX, old bool, color uint32, alpha int32, facing float32, posLocalscl float32) {
	if a.spr == nil || a.spr.Tex == nil {
		return
	}
	h, v, angle := a.drawSub1(angle, facing)
	angle = -angle
	x += xscl * posLocalscl * h * (float32(a.frames[a.drawidx].X) + a.interpolate_offset_x) * (1 / a.scale_x)
	y += yscl * posLocalscl * vscl * v * (float32(a.frames[a.drawidx].Y) + a.interpolate_offset_y) * (1 / a.scale_x)
	var draw func(int32)
	if a.spr.rle <= -11 {
		draw = func(trans int32) {
			RenderMugenFcS(*a.spr.Tex, a.spr.Size,
				AbsF(xscl*h)*float32(a.spr.Offset[0])*sys.widthScale,
				AbsF(yscl*v)*float32(a.spr.Offset[1])*sys.heightScale, &a.tile,
				xscl*h*sys.widthScale, xscl*h*sys.widthScale,
				yscl*v*sys.heightScale, vscl, 0, angle, yangle, xangle, trans, &sys.scrrect,
				(x+float32(sys.gameWidth)/2)*sys.widthScale, y*sys.heightScale, color)
		}
	} else {
		var pal [256]uint32
		if color != 0 || alpha > 0 {
			for i := range pal {
				pal[i] = color | 0xff000000
			}
		}
		draw = func(trans int32) {
			RenderMugen(*a.spr.Tex, pal[:], int32(a.mask), a.spr.Size,
				AbsF(xscl*h)*float32(a.spr.Offset[0])*sys.widthScale,
				AbsF(yscl*v)*float32(a.spr.Offset[1])*sys.heightScale, &a.tile,
				xscl*h*sys.widthScale, xscl*h*sys.widthScale,
				yscl*v*sys.heightScale, vscl, 0, angle, yangle, xangle, trans, &sys.scrrect,
				(x+float32(sys.gameWidth)/2)*sys.widthScale, y*sys.heightScale)
		}
	}
	if color != 0 {
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
			for len(a.frames) == 0 && *i < len(lines) {
				if a2 := at.readAction(sff, lines, i); a2 != nil {
					*a = *a2
					break
				}
				(*i)++
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
	anim        *Animation
	fx          *PalFX
	pos         [2]float32
	scl         [2]float32
	alpha       [2]int32
	priority    int32
	angle       float32
	yangle      float32
	xangle      float32
	ascl        [2]float32
	screen      bool
	bright      bool
	oldVer      bool
	facing      float32
	posLocalscl float32
}
type DrawList []*SprData

func (dl *DrawList) add(sd *SprData, sc, salp int32, so, fo float32) {
	if sys.frameSkip || sd.anim == nil || sd.anim.spr == nil {
		return
	}
	if sd.angle != 0 {
		for i, as := range sd.ascl {
			sd.scl[i] *= as
		}
		sd.ascl = [...]float32{1, 1}
	}
	i, start := 0, 0
	for l := len(*dl); l > 0; {
		i = start + l>>1
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
		if sd.oldVer {
			so *= 1.5
		}
		sys.shadows.add(&ShadowSprite{sd, sc, salp, so, fo})
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
			s.scl[1], 0, s.angle, s.yangle, s.xangle, float32(sys.gameWidth)/2, s.fx, s.oldVer, s.facing, false, s.posLocalscl)
		sys.brightness = ob
	}
}

type ShadowSprite struct {
	*SprData
	shadowColor int32
	shadowAlpha int32
	offsetY     float32
	fadeOffset  float32
}
type ShadowList []*ShadowSprite

func (sl *ShadowList) add(ss *ShadowSprite) {
	i, start := 0, 0
	for l := len(*sl); l > 0; {
		i = start + l>>1
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
		if alpha >= 255 {
			alpha = int32(255 - s.anim.interpolate_blend_dstalpha)
		}
		fend := float32(sys.stage.sdw.fadeend) * sys.stage.localscl
		fbgn := float32(sys.stage.sdw.fadebgn) * sys.stage.localscl
		if fbgn <= fend {
		} else if s.pos[1]-s.fadeOffset <= fend {
			continue
		} else if s.pos[1]-s.fadeOffset < fbgn {
			alpha = int32(float32(alpha) *
				(fend - (s.pos[1] - s.fadeOffset)) / (fend - fbgn))
		}
		if color < 0 {
			color = int32(sys.stage.sdw.color)
			if alpha < 255 {
				intensity = intensity * alpha >> 8
			}
		} else {
			intensity = 0
		}
		color = color&0xff*alpha<<8&0xff0000 |
			color&0xff00*alpha>>8&0xff00 | color&0xff0000*alpha>>24&0xff
		s.anim.ShadowDraw(sys.cam.Offset[0]-(x-s.pos[0])*scl,
			sys.cam.GroundLevel()+sys.cam.Offset[1]-sys.envShake.getOffset()-
				(y+s.pos[1]*sys.stage.sdw.yscale-s.offsetY)*scl,
			scl*s.scl[0], scl*-s.scl[1], sys.stage.sdw.yscale, s.angle, s.yangle, s.xangle,
			&sys.bgPalFX, s.oldVer, uint32(color), intensity, s.facing, s.posLocalscl)
	}
}
func (sl ShadowList) drawReflection(x, y, scl float32) {
	for _, s := range sl {
		if s.alpha[0] < 0 {
			s.anim.srcAlpha = int16(s.anim.interpolate_blend_srcalpha)
			s.anim.dstAlpha = int16(s.anim.interpolate_blend_dstalpha)
		} else {
			s.anim.srcAlpha, s.anim.dstAlpha = int16(s.alpha[0]), int16(s.alpha[1])
		}
		ref := sys.stage.reflection * s.shadowAlpha >> 8
		s.anim.srcAlpha = int16(float32(int32(s.anim.srcAlpha)*ref) / 255)
		if s.anim.dstAlpha < 0 {
			s.anim.dstAlpha = 128
		}
		s.anim.dstAlpha = int16(Min(255, int32(s.anim.dstAlpha)+255-ref))
		if s.anim.srcAlpha == 1 && s.anim.dstAlpha == 255 {
			s.anim.srcAlpha = 0
		}
		s.anim.Draw(&sys.scrrect, sys.cam.Offset[0]/scl-(x-s.pos[0]),
			(sys.cam.GroundLevel()+sys.cam.Offset[1]-sys.envShake.getOffset())/scl-
				(y+s.pos[1]-s.offsetY), scl, scl, s.scl[0], s.scl[0], -s.scl[1], 0,
			s.angle, s.yangle, s.xangle, float32(sys.gameWidth)/2, s.fx, s.oldVer, s.facing, true, s.posLocalscl)
	}
}

type Anim struct {
	anim             *Animation
	window           [4]int32
	x, y, xscl, yscl float32
	palfx            *PalFX
}

func NewAnim(sff *Sff, action string) *Anim {
	lines, i := SplitAndTrim(action, "\n"), 0
	a := &Anim{anim: ReadAnimation(sff, lines, &i),
		window: sys.scrrect, xscl: 1, yscl: 1, palfx: newPalFX()}
	a.palfx.clear()
	a.palfx.time = -1
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
func (a *Anim) SetTile(x, y, sx, sy int32) {
	a.anim.tile[2], a.anim.tile[3], a.anim.tile[0], a.anim.tile[1] = x, y, sx, sy
}
func (a *Anim) SetColorKey(mask int16) {
	a.anim.mask = mask
}
func (a *Anim) SetAlpha(src, dst int16) {
	a.anim.srcAlpha, a.anim.dstAlpha = src, dst
}
func (a *Anim) SetFacing(fc float32) {
	if (fc == 1 && a.xscl < 0) || (fc == -1 && a.xscl > 0) {
		a.xscl *= -1
	}
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
	a.palfx.step()
	a.anim.Action()
}
func (a *Anim) Draw() {
	if !sys.frameSkip {
		a.anim.Draw(&a.window, a.x+float32(sys.gameWidth-320)/2,
			a.y+float32(sys.gameHeight-240), 1, 1, a.xscl, a.xscl, a.yscl,
			0, 0, 0, 0, 0, a.palfx, false, 1, false, 1)
	}
}
func (a *Anim) ResetFrames() {
	a.anim.Reset()
}

type PreloadedAnims map[[2]int16]*Animation

func NewPreloadedAnims() PreloadedAnims {
	return PreloadedAnims(make(map[[2]int16]*Animation))
}
func (pa PreloadedAnims) get(grp, idx int16) *Animation {
	a := pa[[...]int16{grp, idx}]
	if a == nil {
		return a
	}
	ret := &Animation{}
	*ret = *a
	return ret
}
func (pa PreloadedAnims) addAnim(anim *Animation, no int32) {
	pa[[...]int16{int16(no), -1}] = anim
}
func (pa PreloadedAnims) addSprite(sff *Sff, grp, idx int16) {
	if sff.GetSprite(grp, idx) == nil {
		return
	}
	anim := newAnimation(sff)
	anim.mask = 0
	af := newAnimFrame()
	af.Group, af.Number = grp, idx
	anim.frames = append(anim.frames, *af)
	pa[[...]int16{grp, idx}] = anim
}
func (pa PreloadedAnims) updateSff(sff *Sff) {
	for _, v := range pa {
		v.sff = sff
	}
}
