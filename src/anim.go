package main

import (
	"fmt"
	"sort"
	"strings"
)

// AnimFrame holds frame data, used in animation tables.
type AnimFrame struct {
	Time          int32
	Group, Number int16
	Xoffset       int16
	Yoffset       int16
	TransType     TransType
	SrcAlpha      byte
	DstAlpha      byte
	Hscale        int8
	Vscale        int8
	Xscale        float32
	Yscale        float32
	Angle         float32
	Clsn1         [][4]float32
	Clsn2         [][4]float32
}

func newAnimFrame() *AnimFrame {
	return &AnimFrame{
		Time:      -1,
		Group:     -1,
		TransType: TT_none,
		SrcAlpha:  255,
		DstAlpha:  0,
		Hscale:    1, // These two are technically flags but are coded like scale for simplicity
		Vscale:    1,
		Xscale:    1,
		Yscale:    1,
		Angle:     0,
	}
}

func ReadAnimFrame(line string) *AnimFrame {
	if len(line) == 0 || (line[0] < '0' || '9' < line[0]) && line[0] != '-' {
		return nil
	}
	ary := strings.SplitN(line, ",", 10)
	if len(ary) < 5 {
		return nil
	}

	// Read required parameters
	af := newAnimFrame()
	af.Group = int16(Atoi(ary[0]))
	af.Number = int16(Atoi(ary[1]))
	af.Xoffset = int16(Atoi(ary[2]))
	af.Yoffset = int16(Atoi(ary[3]))
	af.Time = Atoi(ary[4])

	// Read H and V flags
	if len(ary) >= 6 {
		for i := range ary[5] {
			switch ary[5][i] {
			case 'H', 'h':
				af.Hscale = -1
				af.Xoffset *= -1
			case 'V', 'v':
				af.Vscale = -1
				af.Yoffset *= -1
			}
		}
	}

	// Read alpha
	if len(ary) >= 7 {
		// Helper to read source and destination
		parseAlphaNumbers := func(a string, start int, defSrc, defDst byte) (src, dst byte) {
			src, dst = defSrc, defDst
			i, alp := start, 0
			for ; i < len(a) && a[i] >= '0' && a[i] <= '9'; i++ {
				alp = alp*10 + int(a[i]-'0')
			}
			alp &= 0x3fff
			if alp < 255 {
				src = byte(alp)
			}
			if i < len(a) && a[i] == 'd' {
				i++
				if i < len(a) && a[i] >= '0' && a[i] <= '9' {
					alp = 0
					for ; i < len(a) && a[i] >= '0' && a[i] <= '9'; i++ {
						alp = alp*10 + int(a[i]-'0')
					}
					alp &= 0x3fff
					if alp < 255 {
						dst = byte(alp)
					}
				}
			}
			return
		}

		ia := strings.IndexAny(ary[6], "ASas")
		if ia >= 0 {
			ary[6] = ary[6][ia:]
		}
		a := strings.ToLower(SplitAndTrim(ary[6], ",")[0])
		switch {
		case a == "a1":
			af.TransType = TT_add
			af.SrcAlpha = 255
			af.DstAlpha = 128

		case len(a) >= 2 && a[:2] == "as": // Add with alpha
			af.TransType = TT_add
			af.SrcAlpha, af.DstAlpha = parseAlphaNumbers(a, 2, 255, 255)

		case len(a) > 0 && a[0] == 'a': // Plain Add
			af.TransType = TT_add
			af.SrcAlpha = 255
			af.DstAlpha = 255

		case len(a) >= 2 && a[:2] == "ss": // Sub with alpha
			af.TransType = TT_sub
			af.SrcAlpha, af.DstAlpha = parseAlphaNumbers(a, 2, 255, 255)

		case len(a) == 1 && a[0] == 's': // Plain sub
			af.TransType = TT_sub
			af.SrcAlpha = 255
			af.DstAlpha = 255
		}
	}

	// Read X scale
	// In Mugen 1.1 a blank parameter means 0
	// In Ikemen it means no change, like the other optional parameters
	if len(ary) >= 8 {
		if IsNumeric(ary[7]) {
			af.Xscale = float32(Atof(ary[7]))
		}
	}

	// Read Y scale
	if len(ary) >= 9 {
		if IsNumeric(ary[8]) {
			af.Yscale = float32(Atof(ary[8]))
		}
	}

	// Read angle
	if len(ary) >= 10 {
		if IsNumeric(ary[9]) {
			af.Angle = float32(Atof(ary[9]))
		}
	}

	return af
}

type Animation struct {
	sff                        *Sff
	palettedata                *PaletteList
	spr                        *Sprite
	frames                     []AnimFrame
	tile                       Tiling
	loopstart                  int32
	interpolate_offset         []int32
	interpolate_scale          []int32
	interpolate_angle          []int32
	interpolate_blend          []int32
	curtime                    int32
	curelem                    int32
	curelemtime                int32
	drawidx                    int32
	totaltime                  int32
	looptime                   int32
	prelooptime                int32
	mask                       int16
	transType                  TransType
	srcAlpha                   int16
	dstAlpha                   int16
	newframe                   bool
	loopend                    bool
	interpolate_offset_x       float32
	interpolate_offset_y       float32
	scale_x                    float32
	scale_y                    float32
	angle                      float32
	curtrans                   TransType
	interpolate_blend_srcalpha float32
	interpolate_blend_dstalpha float32
	remap                      RemapPreset
	start_scale                [2]float32
	isVideo                    bool // Because videos are rendered through Animation.Draw()
}

func newAnimation(sff *Sff, pal *PaletteList) *Animation {
	return &Animation{
		sff:                        sff,
		palettedata:                pal,
		mask:                       -1,
		transType:                  TT_default,
		srcAlpha:                   255,
		dstAlpha:                   0,
		curtrans:                   TT_none,
		interpolate_blend_srcalpha: 255,
		interpolate_blend_dstalpha: 0,
		newframe:                   true,
		remap:                      make(RemapPreset),
		start_scale:                [...]float32{1, 1},
	}
}

func ReadAnimation(sff *Sff, pal *PaletteList, lines []string, i *int) *Animation {
	a := newAnimation(sff, pal)

	a.mask = 0
	ols := int32(0)
	var clsn1, clsn1d, clsn2, clsn2d [][4]float32
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
			af.Clsn1 = clsn1
			af.Clsn2 = clsn2

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
			var clsn [][4]float32
			if line[4] == '1' {
				clsn1 = make([][4]float32, size)
				clsn = clsn1
				if len(line) >= 12 && line[5:12] == "default" {
					clsn1d = clsn1
				}
				def1 = false
			} else if line[4] == '2' {
				clsn2 = make([][4]float32, size)
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
				clsn[n][0], clsn[n][1], clsn[n][2], clsn[n][3] =
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
				a.prelooptime = 0
			}
			a.totaltime += f.Time
			if i < int(a.loopstart) {
				a.prelooptime += f.Time
				tmp += f.Time
			} else {
				a.looptime += f.Time
			}
		}
		if a.totaltime == -1 {
			a.prelooptime = 0
		}
	}
	return a
}

func ReadAction(sff *Sff, pal *PaletteList, lines []string, i *int) (no int32, a *Animation) {
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
	return Atoi(subname[spi+1:]), ReadAnimation(sff, pal, lines, i)
}

func (a *Animation) Reset() {
	a.curelem, a.drawidx = 0, 0
	a.curelemtime, a.curtime = 0, 0
	a.newframe, a.loopend = true, false
	a.spr = nil
}

func (a *Animation) isBlank() bool {
	return a.scale_x == 0 || a.scale_y == 0 || a.spr == nil || a.spr.isBlank()
}

func (a *Animation) isCommonFX() bool {
	for _, fx := range sys.ffx {
		if fx.fsff == a.sff {
			return true
		}
	}
	return false
}

func (a *Animation) AnimTime() int32 {
	return a.curtime - a.totaltime
}

func (a *Animation) AnimElemTime(elem int32) int32 {
	if int(elem) > len(a.frames) {
		t := a.AnimTime()
		if t > 0 {
			t = 0
		}
		return t
	}
	e, t := Max(0, elem)-1, a.curtime
	for i := int32(0); i < e; i++ {
		t -= Max(0, a.frames[i].Time)
	}
	return t
}

func (a *Animation) AnimElemNo(time int32) int32 {
	if len(a.frames) > 0 {
		i, oldt := a.curelem, int32(0)
		if time <= 0 {
			time += a.curelemtime
			loop := false
			for {
				if time >= 0 {
					return i + 1
				}
				i--
				if i < 0 || a.curelem >= a.loopstart && i < a.loopstart {
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
			time += a.curelemtime
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
	return &a.frames[a.curelem]
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

func (a *Animation) SetAnimElem(elem, elemtime int32) {
	a.curelem = Max(0, elem-1)
	// If trying to set an element higher than the last one in the animation
	if int(a.curelem) >= len(a.frames) {
		//if a.totaltime == -1 {
		//	a.current = int32(len(a.frames)) - 1
		//} else if int32(len(a.frames))-a.loopstart > 0 { // Prevent division by zero crash
		//	a.current = a.loopstart +
		//		(a.current-a.loopstart)%(int32(len(a.frames))-a.loopstart)
		//}
		// Mugen merely sets the element to 1
		a.curelem = 0
	}
	a.drawidx = a.curelem

	// Shortcut the most common elemtime
	// Out of range elemtime is also set to 0, as with elem
	if elemtime != 0 {
		frametime := a.frames[a.curelem].Time
		if elemtime < 0 || (frametime != -1 && elemtime >= frametime) {
			elemtime = 0
		}
	}
	a.curelemtime = elemtime

	a.newframe = true
	a.loopend = false
	a.UpdateSprite()

	a.curtime = 0 // Used within AnimElemTime, so must be set to 0 first
	a.curtime = -a.AnimElemTime(a.curelem+1) + a.curelemtime
}

func (a *Animation) animSeek(elem int32) {
	if elem < 0 {
		elem = 0
	}
	foo := true
	for {
		a.curelem = elem
		for int(a.curelem) < len(a.frames) && a.curFrame().Time <= 0 {
			if int(a.curelem) == len(a.frames)-1 && a.curFrame().Time == -1 {
				break
			}
			a.curelem++
		}
		if int(a.curelem) < len(a.frames) {
			break
		}
		foo = !foo
		if foo {
			a.curelem = int32(len(a.frames) - 1)
			break
		}
	}
	if a.curelem < 0 {
		a.curelem = 0
	} else if int(a.curelem) >= len(a.frames) {
		a.curelem = int32(len(a.frames) - 1)
	}
}

func (a *Animation) UpdateSprite() {
	if len(a.frames) == 0 {
		return
	}
	if a.totaltime > 0 {
		if a.curtime >= a.totaltime {
			a.curelemtime, a.newframe, a.curelem = 0, true, a.loopstart
		}
		a.animSeek(a.curelem)
		if a.prelooptime < 0 && a.curtime >= a.totaltime+a.prelooptime &&
			a.curtime >= a.totaltime-a.looptime &&
			(a.curtime == a.totaltime+a.prelooptime ||
				a.curtime == a.totaltime-a.looptime) {
			a.curelemtime = 0
			a.newframe = true
			a.curelem = 0
		}
	}
	if a.newframe && a.sff != nil && a.frames[a.curelem].Time != 0 {
		group, number := a.curFrame().Group, a.curFrame().Number
		if mg, ok := a.remap[group]; ok {
			if mn, ok := mg[number]; ok {
				group, number = mn[0], mn[1]
			}
		}
		a.spr = a.sff.GetSprite(group, number)
	}
	a.newframe, a.drawidx = false, a.curelem

	a.curtrans = a.frames[a.drawidx].TransType
	a.scale_x = a.frames[a.drawidx].Xscale
	a.scale_y = a.frames[a.drawidx].Yscale
	a.angle = a.frames[a.drawidx].Angle

	a.interpolate_offset_x = 0
	a.interpolate_offset_y = 0
	a.interpolate_blend_srcalpha = float32(a.frames[a.drawidx].SrcAlpha)
	a.interpolate_blend_dstalpha = float32(a.frames[a.drawidx].DstAlpha)

	nextDrawidx := a.drawidx + 1
	if int(a.drawidx) >= len(a.frames)-1 {
		nextDrawidx = a.loopstart
	}

	for _, i := range a.interpolate_offset {
		if nextDrawidx == i && (a.frames[a.drawidx].Time >= 0) {
			a.interpolate_offset_x = float32(a.frames[nextDrawidx].Xoffset-a.frames[a.drawidx].Xoffset) / float32(a.curFrame().Time) * float32(a.curelemtime)
			a.interpolate_offset_y = float32(a.frames[nextDrawidx].Yoffset-a.frames[a.drawidx].Yoffset) / float32(a.curFrame().Time) * float32(a.curelemtime)
			break
		}
	}

	for _, i := range a.interpolate_scale {
		if nextDrawidx == i && (a.frames[a.drawidx].Time >= 0) {
			var drawframe_scale_x, nextframe_scale_x, drawframe_scale_y, nextframe_scale_y float32 = 1, 1, 1, 1

			drawframe_scale_x = a.frames[a.drawidx].Xscale
			drawframe_scale_y = a.frames[a.drawidx].Yscale

			nextframe_scale_x = a.frames[nextDrawidx].Xscale
			nextframe_scale_y = a.frames[nextDrawidx].Yscale

			a.scale_x += (nextframe_scale_x - drawframe_scale_x) / float32(a.curFrame().Time) * float32(a.curelemtime)
			a.scale_y += (nextframe_scale_y - drawframe_scale_y) / float32(a.curFrame().Time) * float32(a.curelemtime)
			break
		}
	}
	a.scale_x *= a.start_scale[0]
	a.scale_y *= a.start_scale[1]

	for _, i := range a.interpolate_angle {
		if nextDrawidx == i && (a.frames[a.drawidx].Time >= 0) {
			var drawframe_angle, nextframe_angle float32 = 0, 0

			drawframe_angle = a.frames[a.drawidx].Angle
			nextframe_angle = a.frames[nextDrawidx].Angle

			a.angle += (nextframe_angle - drawframe_angle) / float32(a.curFrame().Time) * float32(a.curelemtime)
			break
		}
	}

	for _, i := range a.interpolate_blend {
		if nextDrawidx == i && (a.frames[a.drawidx].Time >= 0) {
			a.interpolate_blend_srcalpha += (float32(a.frames[nextDrawidx].SrcAlpha) - a.interpolate_blend_srcalpha) / float32(a.curFrame().Time) * float32(a.curelemtime)
			a.interpolate_blend_dstalpha += (float32(a.frames[nextDrawidx].DstAlpha) - a.interpolate_blend_dstalpha) / float32(a.curFrame().Time) * float32(a.curelemtime)
			break
		}
	}
}

func (a *Animation) Action() {
	// Ignore invalid animation instead of crashing engine
	if a == nil || a.frames == nil {
		return
	}
	if len(a.frames) == 0 {
		a.loopend = true
		return
	}
	a.UpdateSprite()
	next := func() {
		if a.totaltime != -1 || int(a.curelem) < len(a.frames)-1 {
			a.curelemtime = 0
			a.newframe = true
			for {
				a.curelem++
				if a.totaltime == -1 && int(a.curelem) == len(a.frames)-1 ||
					int(a.curelem) >= len(a.frames) || a.curFrame().Time > 0 {
					break
				}
			}
		}
	}
	if a.curFrame().Time <= 0 {
		next()
	}
	if int(a.curelem) < len(a.frames) {
		a.curelemtime++
		if a.curelemtime >= a.curFrame().Time {
			next()
			if int(a.curelem) >= len(a.frames) {
				a.curelem = a.loopstart
			}
		}
	} else {
		a.curelem = a.loopstart
	}
	if a.totaltime != -1 && a.curtime >= a.totaltime {
		a.curtime = a.totaltime - a.looptime
	}
	a.curtime++
	if a.totaltime != -1 && a.curtime >= a.totaltime {
		a.loopend = true
	}
}

// Convert animation transparency to RenderParams transparency
func (a *Animation) alphaToBlend() (blendMode TransType, blendAlpha [2]int32) {
	var sa, da byte

	blendMode = a.transType

	if blendMode == TT_default {
		blendMode = a.curtrans
		sa = byte(a.interpolate_blend_srcalpha)
		da = byte(a.interpolate_blend_dstalpha)
	} else {
		sa = byte(a.srcAlpha)
		da = byte(a.dstAlpha)
	}

	// Fallback: Make sure blend mode is not default
	if blendMode == TT_default {
		blendMode = TT_none
		sa = byte(255)
		da = byte(0)
	}

	// Apply system brightness
	sa = byte(int32(sa) * sys.brightness >> 8)

	blendAlpha = [2]int32{int32(sa), int32(da)}

	return
}

func (a *Animation) pal(pfx *PalFX) (p []uint32, plt Texture) {
	if a.palettedata != nil {
		// Apply temporary palette remap if provided
		if pfx != nil && len(pfx.remap) > 0 {
			a.palettedata.SwapPalMap(&pfx.remap)
		}

		// Get palette colors and texture
		p = a.spr.GetPal(a.palettedata)
		plt = a.spr.GetPalTex(a.palettedata)

		// Restore original palette mapping
		if pfx != nil && len(pfx.remap) > 0 {
			a.palettedata.SwapPalMap(&pfx.remap)
		}
	} else {
		if pfx != nil && len(pfx.remap) > 0 {
			a.sff.palList.SwapPalMap(&pfx.remap)
		}

		p = a.spr.GetPal(&a.sff.palList)
		plt = a.spr.GetPalTex(&a.sff.palList)

		if pfx != nil && len(pfx.remap) > 0 {
			a.sff.palList.SwapPalMap(&pfx.remap)
		}
	}

	return
}

func (a *Animation) drawSub1(angle, facing float32) (h, v, agl float32) {
	h, v = 1, 1
	if len(a.frames) > 0 {
		h, v = float32(a.frames[a.drawidx].Hscale), float32(a.frames[a.drawidx].Vscale)
	}
	agl = angle
	h *= a.scale_x
	v *= a.scale_y
	agl += a.angle * facing
	return
}

func (a *Animation) Draw(window *[4]int32, x, y, xcs, ycs, xs, xbs, ys,
	rxadd float32, rot Rotation, rcx float32, pfx *PalFX, old bool, facing float32,
	airOffsetFix [2]float32, projectionMode int32, fLength float32, color uint32, isReflection bool) {

	// Skip blank animations
	if a == nil || a.isBlank() {
		return
	}

	// Determine animation angle. Invert for reflection
	h, v, angle := a.drawSub1(rot.angle, facing)
	if isReflection && sys.stage.reflection.yscale > 0 {
		angle = -angle
	}
	rot.angle = angle
	xs *= xcs * h
	ys *= ycs * v

	// Compute X and Y AIR animation offsets
	var xoff, yoff float32
	if !a.isVideo {
		xoff = xs * airOffsetFix[0] * (float32(a.frames[a.drawidx].Xoffset) + a.interpolate_offset_x) * a.start_scale[0] * (1 / a.scale_x)
		yoff = ys * airOffsetFix[1] * (float32(a.frames[a.drawidx].Yoffset) + a.interpolate_offset_y) * a.start_scale[1] * (1 / a.scale_y)
	}

	x = xcs*x + xoff
	y = ycs*y + yoff

	var rcy float32
	if rot.IsZero() {
		if xs < 0 {
			x *= -1
			// This was deliberately replicating a Mugen bug, but we don't need that
			// TODO: Maybe we don't need all these "old" arguments in the functions anymore
			//if old {
			//	x += xs
			//}
		}
		if ys < 0 {
			y *= -1
			// This was deliberately replicating a Mugen bug, but we don't need that
			//if old {
			//	y += ys
			//}
		}
		if a.tile.xflag == 1 {
			space := xs * float32(a.tile.xspacing)
			if a.tile.xspacing <= 0 {
				space += xs * float32(a.spr.Size[0])
			}
			if space != 0 {
				x -= float32(int(x/space)) * space
			}
		}
		if a.tile.yflag == 1 {
			space := ys * float32(a.tile.yspacing)
			if a.tile.yspacing <= 0 {
				space += ys * float32(a.spr.Size[1])
			}
			if space != 0 {
				y -= float32(int(y/space)) * space
			}
		}
		rcx, rcy = rcx*sys.widthScale, 0
		x = -x + AbsF(xs)*float32(a.spr.Offset[0])
		y = -y + AbsF(ys)*float32(a.spr.Offset[1])
	} else {
		rcx, rcy = (x+rcx)*sys.widthScale, y*sys.heightScale
		x, y = AbsF(xs)*float32(a.spr.Offset[0]), AbsF(ys)*float32(a.spr.Offset[1])
		fLength *= ycs
	}

	blendMode, blendAlpha := a.alphaToBlend()

	var paltex Texture
	if !a.isVideo {
		var pal []uint32
		pal, paltex = a.pal(pfx)
		if a.spr.coldepth <= 8 && paltex == nil {
			paltex = a.spr.CachePalette(pal)
		}
	}

	rp := RenderParams{
		tex:            a.spr.Tex,
		paltex:         paltex,
		size:           a.spr.Size,
		x:              x * sys.widthScale,
		y:              y * sys.heightScale,
		tile:           a.tile,
		xts:            xs * sys.widthScale,
		xbs:            xcs * xbs * h * sys.widthScale,
		ys:             ys * sys.heightScale,
		vs:             1,
		rxadd:          xcs * rxadd * sys.widthScale / sys.heightScale,
		xas:            h,
		yas:            v,
		rot:            rot,
		tint:           color,
		blendMode:      blendMode,
		blendAlpha:     blendAlpha,
		mask:           int32(a.mask),
		pfx:            pfx,
		window:         window,
		rcx:            rcx,
		rcy:            rcy,
		projectionMode: projectionMode,
		fLength:        fLength * sys.heightScale,
		xOffset:        xoff * sys.widthScale,
		yOffset:        yoff * sys.heightScale,
	}

	RenderSprite(rp)
}

func (a *Animation) ShadowDraw(window *[4]int32, x, y, xscl, yscl, vscl, rxadd float32, rot Rotation,
	pfx *PalFX, old bool, color uint32, intensity int32, facing float32, airOffsetFix [2]float32, projectionMode int32, fLength float32) {

	// Skip blank shadows
	if a == nil || a.isBlank() {
		return
	}

	// Determine animation angle. Invert for shadows
	h, v, angle := a.drawSub1(rot.angle, facing)
	rot.angle = -angle
	if yscl < 0 && rot.angle != 0 {
		rxadd = -rxadd
	}

	// Compute X and Y AIR animation offsets
	xoff := xscl * airOffsetFix[0] * h * (float32(a.frames[a.drawidx].Xoffset) + a.interpolate_offset_x) * (1 / a.scale_x)
	yoff := yscl * airOffsetFix[1] * vscl * v * (float32(a.frames[a.drawidx].Yoffset) + a.interpolate_offset_y) * (1 / a.scale_y)

	x += xoff
	y += yoff

	rp := RenderParams{
		tex:            a.spr.Tex,
		paltex:         nil,
		size:           a.spr.Size,
		x:              AbsF(xscl*h) * float32(a.spr.Offset[0]) * sys.widthScale,
		y:              AbsF(yscl*v) * float32(a.spr.Offset[1]) * sys.heightScale,
		tile:           a.tile,
		xts:            xscl * h * sys.widthScale,
		xbs:            xscl * h * sys.widthScale,
		ys:             yscl * v * sys.heightScale,
		vs:             vscl,
		rxadd:          rxadd,
		xas:            h,
		yas:            v,
		rot:            rot,
		tint:           color | 0xff000000,
		blendMode:      TT_sub,
		blendAlpha:     [2]int32{0, 0},
		mask:           int32(a.mask),
		pfx:            nil,
		window:         window,
		rcx:            (x + float32(sys.gameWidth)/2) * sys.widthScale,
		rcy:            y * sys.heightScale,
		projectionMode: projectionMode,
		fLength:        fLength,
		xOffset:        xoff,
		yOffset:        yoff,
	}

	// TODO: This is redundant now that rp.tint is used to colorise the shadow
	//if a.spr.coldepth <= 8 {
	//	var pal [256]uint32
	//	if color != 0 || alpha > 0 {
	//		paltemp := a.spr.paltemp
	//		if len(paltemp) == 0 {
	//			if a.palettedata != nil {
	//				paltemp = a.spr.GetPal(a.palettedata)
	//			} else {
	//				paltemp = a.spr.GetPal(&a.sff.palList)
	//			}
	//		}
	//		for i := range pal {
	//			// Skip transparent colors
	//			if len(paltemp) > i && paltemp[i] != 0 {
	//				pal[i] = color | 0xff000000
	//			}
	//		}
	//	}
	//	rp.paltex = PaletteToTexture(pal[:])
	//}

	if a.spr.coldepth <= 8 && (color != 0 || intensity > 0) {
		if a.sff.header.Ver0 == 2 && a.sff.header.Ver2 == 1 {
			pal, _ := a.pal(pfx)
			if a.spr.PalTex == nil {
				a.spr.PalTex = a.spr.CachePalette(pal)
			}
			rp.paltex = a.spr.PalTex
		} else {
			rp.paltex = sys.whitePalTex
		}
	}

	// Draw shadow with one pass for intensity and another for color
	// Drawing twice is a bit wasteful, but results are the most accurate to Mugen
	if intensity > 0 {
		rp.blendMode = TT_add
		rp.blendAlpha = [2]int32{intensity, 255 - intensity}
		RenderSprite(rp)
	}
	if color != 0 {
		rp.blendMode = TT_sub
		rp.blendAlpha = [2]int32{255, 255}
		RenderSprite(rp)
	}

	// This method would draw color and intensity in a single pass, but it's less accurate
	// Another disadvantage of this method is that the double pass allows a wider range of color/intensity combinations
	//if intensity > 0 || color != 0 {
	//	rp.blendMode = TT_sub
	//	rp.blendAlpha = [2]int32{intensity, 255 - intensity}
	//	RenderSprite(rp)
	//}
}

type AnimationTable map[int32]*Animation

func NewAnimationTable() AnimationTable {
	return AnimationTable(make(map[int32]*Animation))
}

func (at AnimationTable) readAction(sff *Sff, pal *PaletteList,
	lines []string, i *int) *Animation {
	for *i < len(lines) {
		no, a := ReadAction(sff, pal, lines, i)
		if a != nil {
			if tmp := at[no]; tmp != nil {
				return tmp
			}
			at[no] = a
			for len(a.frames) == 0 && *i < len(lines) {
				if a2 := at.readAction(sff, pal, lines, i); a2 != nil {
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
func ReadAnimationTable(sff *Sff, pal *PaletteList, lines []string, i *int) AnimationTable {
	at := NewAnimationTable()
	for at.readAction(sff, pal, lines, i) != nil {
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
	anim         *Animation
	fx           *PalFX
	pos          [2]float32
	scl          [2]float32
	trans        TransType
	alpha        [2]int32
	priority     int32
	rot          Rotation
	screen       bool
	undarken     bool // Ignore SuperPause "darken"
	oldVer       bool
	facing       float32
	airOffsetFix [2]float32 // posLocalscl replacement
	projection   int32
	fLength      float32
	window       [4]float32
	xshear       float32
}

func (sd *SprData) isBlank() bool {
	return sd.scl[0] == 0 || sd.scl[1] == 0 || sd.anim == nil || sd.anim.isBlank()
}

type DrawList []*SprData

func (dl *DrawList) add(sd *SprData) {
	// Ignore if skipping the frame or adding a blank sprite
	if sys.frameSkip || sd == nil || sd.isBlank() {
		return
	}

	// Before: sort every time we add a sprite
	// After: add all sprites first then sort before drawing
	/*
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
	*/

	// Just append. We will sort everything later in one go
	*dl = append(*dl, sd)
}

func (dl DrawList) draw(cameraX, cameraY, cameraScl float32) {
	if len(dl) == 0 {
		return
	}

	// Sort by descending sprpriority
	sort.SliceStable(dl, func(i, j int) bool {
		if dl[i].priority != dl[j].priority {
			return dl[i].priority > dl[j].priority
		}
		return false
	})

	// Common variables
	shake := sys.envShake.getOffset()

	// Draw the entire list in reverse
	for i := len(dl) - 1; i >= 0; i-- {
		s := dl[i]

		// Skip blank SprData
		// https://github.com/ikemen-engine/Ikemen-GO/issues/2433
		if s.isBlank() {
			continue
		}

		// Save system brightness
		oldBright := sys.brightness
		if s.undarken {
			sys.brightness = 256
		}

		// Backup animation transparency to temporarily change it
		oldTransType := s.anim.transType
		oldSrcAlpha := s.anim.srcAlpha
		oldDstAlpha := s.anim.dstAlpha

		// Determine transparency
		if s.trans == TT_default {
			s.anim.transType = s.anim.curtrans
			s.anim.srcAlpha = int16(s.anim.interpolate_blend_srcalpha)
			s.anim.dstAlpha = int16(s.anim.interpolate_blend_dstalpha)
		} else {
			s.anim.transType = s.trans
			s.anim.srcAlpha = int16(s.alpha[0])
			s.anim.dstAlpha = int16(s.alpha[1])
		}

		var pos [2]float32
		cs := cameraScl
		if s.screen {
			pos = [2]float32{s.pos[0], s.pos[1] + float32(sys.gameHeight-240)}
			cs = 1
		} else {
			pos = [2]float32{(sys.cam.Offset[0]-shake[0])/cs - (cameraX - s.pos[0]),
				(sys.cam.GroundLevel()+(sys.cam.Offset[1]-shake[1]))/cs -
					(cameraY/cs - s.pos[1])}
		}

		// Xshear offset correction
		xshear := -s.xshear
		xsoffset := xshear * (float32(s.anim.spr.Offset[1]) * s.scl[1] * cs)

		drawwindow := &sys.scrrect
		// Sprite window, which can be from the Char, Explod, or Projectile
		if s.window != [4]float32{0, 0, 0, 0} {
			w := s.window
			var window [4]int32

			if w[0] > w[2] {
				w[0], w[2] = w[2], w[0]
			}
			if w[1] > w[3] {
				w[1], w[3] = w[3], w[1]
			}

			window[0] = int32((cs*(pos[0]+float32(w[0])) + float32(sys.gameWidth)/2) * sys.widthScale)
			window[1] = int32(cs * (pos[1] + float32(w[1])) * sys.heightScale)
			window[2] = int32(cs * (w[2] - w[0]) * sys.widthScale)
			window[3] = int32(cs * (w[3] - w[1]) * sys.heightScale)

			drawwindow = &window
		}

		s.anim.Draw(drawwindow, pos[0]-xsoffset, pos[1], cs, cs, s.scl[0], s.scl[0],
			s.scl[1], xshear, s.rot, float32(sys.gameWidth)/2, s.fx, s.oldVer, s.facing,
			s.airOffsetFix, s.projection, s.fLength, 0, false)

		// Restore original animation transparency just in case
		s.anim.transType = oldTransType
		s.anim.srcAlpha = oldSrcAlpha
		s.anim.dstAlpha = oldDstAlpha

		// Restore system brightness
		sys.brightness = oldBright
	}
	BlendReset()
}

type ShadowSprite struct {
	*SprData
	shadowColor      int32
	shadowAlpha      int32
	shadowIntensity  int32
	shadowOffset     [2]float32
	shadowWindow     [4]float32
	shadowXshear     float32
	shadowYscale     float32
	shadowRot        Rotation
	shadowProjection int32
	shadowfLength    float32
	fadeOffset       float32
}

type ShadowList []*ShadowSprite

func (sl *ShadowList) add(ss *ShadowSprite) {
	// Ignore if skipping the frame or adding a blank sprite
	if sys.frameSkip || ss.SprData == nil || ss.SprData.isBlank() {
		return
	}

	/*
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
	*/

	// Just append. We will sort everything later in one go
	*sl = append(*sl, ss)
}

func (sl ShadowList) draw(x, y, scl float32) {
	if len(sl) == 0 {
		return
	}

	// Sort by descending sprpriority
	sort.SliceStable(sl, func(i, j int) bool {
		if sl[i].priority != sl[j].priority {
			return sl[i].priority > sl[j].priority
		}
		return false
	})

	// Common variables
	shake := sys.envShake.getOffset()

	// Draw the entire list in reverse
	for i := len(sl) - 1; i >= 0; i-- {
		s := sl[i]

		// Skip blank shadows
		if s == nil || s.anim == nil || s.anim.isBlank() {
			continue
		}

		// Use stage or custom intensity
		var intensity int32
		if s.shadowIntensity != -1 {
			intensity = s.shadowIntensity
		} else {
			intensity = sys.stage.sdw.intensity
		}

		color, alpha := s.shadowColor, s.shadowAlpha

		if s.trans == TT_default {
			alpha = int32(255 - s.anim.interpolate_blend_dstalpha)
		}

		// Fading range
		fend := float32(sys.stage.sdw.fadeend) * sys.stage.localscl
		fbgn := float32(sys.stage.sdw.fadebgn) * sys.stage.localscl
		if fbgn <= fend {
			// Ignore incorrect fade effect
		} else if s.pos[1]-s.fadeOffset <= fend {
			continue // Do not render shadow
		} else if s.pos[1]-s.fadeOffset < fbgn {
			alpha = int32(float32(alpha) * (fend - (s.pos[1] - s.fadeOffset)) / (fend - fbgn))
		}

		if color < 0 {
			color = int32(sys.stage.sdw.color)
			if alpha < 255 {
				intensity = intensity * alpha >> 8
			}
		}

		color = color&0xff*alpha<<8&0xff0000 | color&0xff00*alpha>>8&0xff00 | color&0xff0000*alpha>>24&0xff

		var xshear float32
		if s.xshear != 0 {
			xshear = -s.xshear
		} else {
			xshear = sys.stage.sdw.xshear + s.shadowXshear
		}

		var yscale float32
		if s.shadowYscale != 0 {
			yscale = sys.stage.sdw.yscale * s.shadowYscale
		} else {
			yscale = sys.stage.sdw.yscale
		}

		if yscale > 0 {
			xshear = -xshear // Invert if sprite is flipped
		}

		offsetX := s.shadowOffset[0] + sys.stage.sdw.offset[0]
		offsetY := s.shadowOffset[1] + sys.stage.sdw.offset[1]

		// Rotation offset. Only shadow scale sign
		xrotoff := xshear * SignF(yscale) * (float32(s.anim.spr.Offset[1]) * s.scl[1])

		rotVal := func(vals ...float32) float32 {
			for _, v := range vals {
				if v != 0 {
					return v
				}
			}
			return 0
		}

		rot := Rotation{
			angle:  rotVal(s.shadowRot.angle, sys.stage.sdw.rot.angle, s.rot.angle),
			xangle: rotVal(s.shadowRot.xangle, sys.stage.sdw.rot.xangle, s.rot.xangle),
			yangle: rotVal(s.shadowRot.yangle, sys.stage.sdw.rot.yangle, s.rot.yangle),
		}

		if rot.angle != 0 {
			offsetX -= xrotoff
		} else {
			offsetX += xrotoff
		}

		// With a shearing effect, the Y position should also affect the X position when not grounded
		if xshear != 0 && s.pos[1] != 0 {
			offsetX += (-s.pos[1] + s.fadeOffset) * xshear * SignF(yscale)
		}

		var projection int32
		if s.shadowProjection != -1 {
			projection = int32(s.shadowProjection)
		} else if sys.stage.sdw.projection != 0 {
			projection = int32(sys.stage.sdw.projection)
		} else {
			projection = int32(s.projection)
		}

		var fLength float32
		if s.shadowfLength != 0 {
			fLength = s.shadowfLength
		} else if sys.stage.sdw.fLength != 0 {
			fLength = sys.stage.sdw.fLength
		} else {
			fLength = s.fLength
		}

		drawwindow := &sys.scrrect

		// TODO: If the char has an active window sctrl, shadows should also be affected, in addition to the stage window
		if sys.stage.sdw.window != [4]float32{0, 0, 0, 0} || s.shadowWindow != [4]float32{0, 0, 0, 0} {
			var w [4]float32
			var window [4]int32

			if s.shadowWindow != [4]float32{0, 0, 0, 0} {
				w = s.shadowWindow
			} else {
				w = sys.stage.sdw.window
			}

			w[1], w[3] = -w[1], -w[3]
			if w[0] > w[2] {
				w[0], w[2] = w[2], w[0]
			}
			if (w[1] > w[3] && yscale > 0) || (w[1] < w[3] && yscale < 0) {
				w[1], w[3] = w[3], w[1]
			}

			for i := range w {
				w[i] *= sys.stage.localscl
			}

			window[0] = int32(((sys.cam.Offset[0] - shake[0]) - (x * scl) + w[0]*scl + float32(sys.gameWidth)/2) * sys.widthScale)
			window[1] = int32((sys.cam.GroundLevel() + (sys.cam.Offset[1] - shake[1]) - y + w[1]*SignF(yscale)*scl) * sys.heightScale)
			window[2] = int32(scl * (w[2] - w[0]) * sys.widthScale)
			window[3] = int32(scl * (w[3] - w[1]) * sys.heightScale * SignF(yscale))

			drawwindow = &window
		}

		s.anim.ShadowDraw(drawwindow,
			(sys.cam.Offset[0]-shake[0])-((x-s.pos[0]-offsetX)*scl),
			sys.cam.GroundLevel()+(sys.cam.Offset[1]-shake[1])-y-(s.pos[1]*yscale-offsetY)*scl,
			scl*s.scl[0], scl*-s.scl[1],
			yscale, xshear, rot,
			s.fx, s.oldVer, uint32(color), intensity, s.facing, s.airOffsetFix, projection, fLength)
	}
}

type ReflectionSprite struct {
	*SprData
	reflectColor      int32
	reflectIntensity  int32
	reflectOffset     [2]float32
	reflectWindow     [4]float32
	reflectXshear     float32
	reflectYscale     float32
	reflectRot        Rotation
	reflectProjection int32
	reflectfLength    float32
	fadeOffset        float32
}

type ReflectionList []*ReflectionSprite

func (rl *ReflectionList) add(rs *ReflectionSprite) {
	if sys.frameSkip || rs.SprData == nil || rs.SprData.isBlank() {
		return
	}

	// Stage without reflections
	// TODO: Maybe ModifyReflection should be able to bypass this
	if sys.stage.reflection.intensity == 0 {
		return
	}

	*rl = append(*rl, rs)
}

func (rl ReflectionList) draw(x, y, scl float32) {
	if len(rl) == 0 {
		return
	}

	// Sort by descending sprpriority
	sort.SliceStable(rl, func(i, j int) bool {
		if rl[i].priority != rl[j].priority {
			return rl[i].priority > rl[j].priority
		}
		return false
	})

	// Common variables
	shake := sys.envShake.getOffset()

	// Draw the entire list in reverse
	for i := len(rl) - 1; i >= 0; i-- {
		s := rl[i]

		// Skip blank reflections
		if s == nil || s.anim == nil || s.anim.isBlank() {
			continue
		}

		// Backup animation transparency to temporarily change it
		oldTransType := s.anim.transType
		oldSrcAlpha := s.anim.srcAlpha
		oldDstAlpha := s.anim.dstAlpha

		// Get base alpha
		if s.trans == TT_default {
			s.anim.transType = s.anim.curtrans
			s.anim.srcAlpha = int16(s.anim.interpolate_blend_srcalpha)
			s.anim.dstAlpha = int16(s.anim.interpolate_blend_dstalpha)
		} else {
			s.anim.transType = s.trans
			s.anim.srcAlpha = int16(s.alpha[0])
			s.anim.dstAlpha = int16(s.alpha[1])
		}

		// Force reflections into Add transparency, unless original sprite has Sub
		if s.anim.transType != TT_sub {
			s.anim.transType = TT_add
		}

		// Apply reflection intensity
		var ref int32
		if s.reflectIntensity != -1 {
			ref = s.reflectIntensity
		} else {
			ref = sys.stage.reflection.intensity
		}

		// Scale intensity by linear interpolation
		s.anim.srcAlpha = int16(int32(s.alpha[0]) * ref / 255)
		s.anim.dstAlpha = int16(255 - (int32(255-s.alpha[1]) * ref / 255))

		// Set the tint if it's there
		var color uint32
		if s.reflectColor < 0 {
			color = sys.stage.reflection.color
		} else {
			color = uint32(s.reflectColor)
		}

		// Add alpha if color is specified
		if color != 0 {
			color |= uint32(ref << 24)
		}

		var xshear float32
		if s.xshear != 0 {
			xshear = -s.xshear
		} else {
			xshear = sys.stage.reflection.xshear + s.reflectXshear
		}

		var yscale float32
		if s.reflectYscale != 0 {
			yscale = sys.stage.reflection.yscale * s.reflectYscale
		} else {
			yscale = sys.stage.reflection.yscale
		}

		if yscale > 0 {
			xshear = -xshear // Invert if sprite is flipped
		}

		offsetX := s.reflectOffset[0] + sys.stage.reflection.offset[0]
		offsetY := s.reflectOffset[1] + sys.stage.reflection.offset[1]

		// Rotation offset
		xrotoff := xshear * yscale * (float32(s.anim.spr.Offset[1]) * s.scl[1] * scl)

		rotVal := func(vals ...float32) float32 {
			for _, v := range vals {
				if v != 0 {
					return v
				}
			}
			return 0
		}

		rot := Rotation{
			angle:  rotVal(s.reflectRot.angle, sys.stage.reflection.rot.angle, s.rot.angle),
			xangle: rotVal(s.reflectRot.xangle, sys.stage.reflection.rot.xangle, s.rot.xangle),
			yangle: rotVal(s.reflectRot.yangle, sys.stage.reflection.rot.yangle, s.rot.yangle),
		}

		if rot.angle != 0 {
			xshear = -xshear
			offsetX -= xrotoff
		} else {
			offsetX += xrotoff
		}

		// With a shearing effect, the Y position should also affect the X position when not grounded
		if xshear != 0 && s.pos[1] != 0 {
			offsetX += (-s.pos[1] + s.fadeOffset) * xshear * SignF(yscale)
		}

		var projection int32
		if s.reflectProjection != -1 {
			projection = int32(s.reflectProjection)
		} else if sys.stage.reflection.projection != 0 {
			projection = int32(sys.stage.reflection.projection)
		} else {
			projection = int32(s.projection)
		}

		var fLength float32
		if s.reflectfLength != 0 {
			fLength = s.reflectfLength
		} else if sys.stage.reflection.fLength != 0 {
			fLength = sys.stage.reflection.fLength
		} else {
			fLength = s.fLength
		}

		drawwindow := &sys.scrrect

		// TODO: If the char has an active window sctrl, reflections should also be affected, in addition to the stage window
		if sys.stage.reflection.window != [4]float32{0, 0, 0, 0} || s.reflectWindow != [4]float32{0, 0, 0, 0} {
			var w [4]float32
			var window [4]int32

			if s.reflectWindow != [4]float32{0, 0, 0, 0} {
				w = s.reflectWindow
			} else {
				w = sys.stage.reflection.window
			}

			w[1], w[3] = -w[1], -w[3]
			if w[0] > w[2] {
				w[0], w[2] = w[2], w[0]
			}
			if (w[1] > w[3] && yscale > 0) || (w[1] < w[3] && yscale < 0) {
				w[1], w[3] = w[3], w[1]
			}

			for i := range w {
				w[i] *= sys.stage.localscl
			}

			window[0] = int32(((sys.cam.Offset[0] - shake[0]) - (x * scl) + w[0]*scl + float32(sys.gameWidth)/2) * sys.widthScale)
			window[1] = int32((sys.cam.GroundLevel() + (sys.cam.Offset[1] - shake[1]) - y + w[1]*SignF(yscale)*scl) * sys.heightScale)
			window[2] = int32(scl * (w[2] - w[0]) * sys.widthScale)
			window[3] = int32(scl * (w[3] - w[1]) * sys.heightScale * SignF(yscale))

			drawwindow = &window
		}

		s.anim.Draw(drawwindow,
			(sys.cam.Offset[0]-shake[0])/scl-(x-s.pos[0]-offsetX),
			(sys.cam.GroundLevel()+sys.cam.Offset[1]-shake[1])/scl-y/scl-(s.pos[1]*yscale-offsetY),
			scl, scl, s.scl[0], s.scl[0],
			-s.scl[1]*yscale, xshear, rot, float32(sys.gameWidth)/2,
			s.fx, s.oldVer, s.facing, s.airOffsetFix, projection, fLength, color, true)

		// Restore original animation transparency just in case
		s.anim.transType = oldTransType
		s.anim.srcAlpha = oldSrcAlpha
		s.anim.dstAlpha = oldDstAlpha
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
	a := &Anim{anim: ReadAnimation(sff, &sff.palList, lines, &i),
		window: sys.scrrect, x: sys.luaSpriteOffsetX,
		xscl: 1, yscl: 1, palfx: newPalFX()}
	a.palfx.clear()
	a.palfx.time = -1
	if len(a.anim.frames) == 0 {
		return nil
	}
	return a
}

func CopyAnim(a *Anim) *Anim {
	//Copy information, this is used for showing palettes to avoid the same palette when a character is selected twice.
	copySff := newSff()
	copySff.header = a.anim.sff.header
	copySff.palList.palettes = a.anim.sff.palList.palettes
	x := 0
	copySff.palList.paletteMap = nil
	for x < len(a.anim.sff.palList.paletteMap) { //Copy each value from the palette map individually, without doing this, different sides/members of the same character will share palettes.
		copySff.palList.paletteMap = append(copySff.palList.paletteMap, x)
		x = x + 1
	}
	copySff.palList.PalTable = a.anim.sff.palList.PalTable
	copySff.palList.numcols = a.anim.sff.palList.numcols
	copySff.palList.PalTex = a.anim.sff.palList.PalTex
	frameAnims := ""
	x = 0
	for x < len(a.anim.frames) {
		frameAnims = frameAnims + fmt.Sprint(a.anim.frames[x].Group) + "," + fmt.Sprint(a.anim.frames[x].Number) + "," + fmt.Sprint(a.anim.frames[x].Xoffset) + "," + fmt.Sprint(a.anim.frames[x].Yoffset) + "," + fmt.Sprint(a.anim.frames[x].Time) + "\n"
		x = x + 1
	}

	//Create animation and copy animation data
	newAnim := NewAnim(copySff, frameAnims)
	newAnim.window = a.window
	newAnim.x = a.x
	newAnim.y = a.y
	newAnim.xscl = a.xscl
	newAnim.yscl = a.yscl
	newAnim.palfx = a.palfx

	//Information to match the current frame in the animation
	newAnim.anim.looptime = a.anim.looptime
	newAnim.anim.loopstart = a.anim.loopstart
	newAnim.anim.curtime = a.anim.curtime
	newAnim.anim.curelem = a.anim.curelem
	newAnim.anim.totaltime = a.anim.totaltime
	newAnim.anim.frames = a.anim.frames
	newAnim.anim.interpolate_blend_srcalpha = a.anim.interpolate_blend_srcalpha
	newAnim.anim.interpolate_scale = a.anim.interpolate_scale
	for _, c := range a.anim.frames {
		newAnim.anim.sff.sprites[[...]int16{c.Group, c.Number}] = newSprite()
		newAnim.anim.sff.sprites[[...]int16{c.Group, c.Number}].Pal = a.anim.sff.sprites[[...]int16{c.Group, c.Number}].Pal
		newAnim.anim.sff.sprites[[...]int16{c.Group, c.Number}].Tex = a.anim.sff.sprites[[...]int16{c.Group, c.Number}].Tex
		newAnim.anim.sff.sprites[[...]int16{c.Group, c.Number}].palidx = a.anim.sff.sprites[[...]int16{c.Group, c.Number}].palidx
		newAnim.anim.sff.sprites[[...]int16{c.Group, c.Number}].coldepth = a.anim.sff.sprites[[...]int16{c.Group, c.Number}].coldepth
		newAnim.anim.sff.sprites[[...]int16{c.Group, c.Number}].Offset[1] = a.anim.sff.sprites[[...]int16{c.Group, c.Number}].Offset[1]
		newAnim.anim.sff.sprites[[...]int16{c.Group, c.Number}].Size[1] = a.anim.sff.sprites[[...]int16{c.Group, c.Number}].Size[1]
		newAnim.anim.sff.sprites[[...]int16{c.Group, c.Number}].Offset[0] = a.anim.sff.sprites[[...]int16{c.Group, c.Number}].Offset[0]
		newAnim.anim.sff.sprites[[...]int16{c.Group, c.Number}].Size[0] = a.anim.sff.sprites[[...]int16{c.Group, c.Number}].Size[0]
	}
	return newAnim
}

func (a *Anim) SetPos(x, y float32) {
	a.x, a.y = x, y
}

func (a *Anim) AddPos(x, y float32) {
	a.x += x
	a.y += y
}

func (a *Anim) SetTile(x, y, sx, sy int32) {
	a.anim.tile.xflag, a.anim.tile.yflag, a.anim.tile.xspacing, a.anim.tile.yspacing = x, y, sx, sy
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
			0, Rotation{}, 0, a.palfx, false, 1, [2]float32{1, 1}, 0, 0, 0, false)
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
	anim := newAnimation(sff, &sff.palList)
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
