package main

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"image/draw"
	_ "image/jpeg"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"

	_ "github.com/lukegb/dds"

	mgl "github.com/go-gl/mathgl/mgl32"
	"github.com/qmuntal/gltf"
	"github.com/qmuntal/gltf/modeler"
	"golang.org/x/mobile/exp/f32"
)

type StageProps struct {
	roundpos bool
}

func newStageProps() StageProps {
	sp := StageProps{
		roundpos: false,
	}

	return sp
}

type EnvShake struct {
	time  int32
	freq  float32
	ampl  float32
	phase float32
	mul   float32
}

func (es *EnvShake) clear() {
	*es = EnvShake{freq: float32(math.Pi / 3), ampl: -4.0,
		phase: float32(math.NaN()), mul: 1.0}
}
func (es *EnvShake) setDefPhase() {
	if math.IsNaN(float64(es.phase)) {
		if es.freq >= math.Pi/2 {
			es.phase = math.Pi / 2
		} else {
			es.phase = 0
		}
	}
}
func (es *EnvShake) next() {
	if es.time > 0 {
		es.time--
		es.phase += es.freq
		if es.phase > math.Pi*2 {
			es.ampl *= es.mul
			es.phase -= math.Pi * 2
		}
	} else {
		es.ampl = 0
	}
}
func (es *EnvShake) getOffset() float32 {
	if es.time > 0 {
		return es.ampl * float32(math.Sin(float64(es.phase)))
	}
	return 0
}

type BgcType int32

const (
	BT_Null BgcType = iota
	BT_Anim
	BT_Visible
	BT_Enable
	BT_PalFX
	BT_PosSet
	BT_PosAdd
	BT_RemapPal
	BT_SinX
	BT_SinY
	BT_VelSet
	BT_VelAdd
)

type bgAction struct {
	offset      [2]float32
	sinoffset   [2]float32
	pos, vel    [2]float32
	radius      [2]float32
	sintime     [2]int32
	sinlooptime [2]int32
}

func (bga *bgAction) clear() {
	*bga = bgAction{}
}
func (bga *bgAction) action() {
	for i := 0; i < 2; i++ {
		bga.pos[i] += bga.vel[i]
		if bga.sinlooptime[i] > 0 {
			bga.sinoffset[i] = bga.radius[i] * float32(math.Sin(
				2*math.Pi*float64(bga.sintime[i])/float64(bga.sinlooptime[i])))
			bga.sintime[i]++
			if bga.sintime[i] >= bga.sinlooptime[i] {
				bga.sintime[i] = 0
			}
		} else {
			bga.sinoffset[i] = 0
		}
		bga.offset[i] = bga.pos[i] + bga.sinoffset[i]
	}
}

type backGround struct {
	typ                int
	palfx              *PalFX
	anim               Animation
	bga                bgAction
	id                 int32
	start              [2]float32
	xofs               float32
	delta              [2]float32
	width              [2]int32
	xscale             [2]float32
	rasterx            [2]float32
	yscalestart        float32
	yscaledelta        float32
	actionno           int32
	startv             [2]float32
	startrad           [2]float32
	startsint          [2]int32
	startsinlt         [2]int32
	visible            bool
	active             bool
	positionlink       bool
	toplayer           bool
	autoresizeparallax bool
	notmaskwindow      int32
	startrect          [4]int32
	windowdelta        [2]float32
	scalestart         [2]float32
	scaledelta         [2]float32
	zoomdelta          [2]float32
	zoomscaledelta     [2]float32
	xbottomzoomdelta   float32
	roundpos           bool
}

func newBackGround(sff *Sff) *backGround {
	return &backGround{palfx: newPalFX(), anim: *newAnimation(sff, &sff.palList), delta: [...]float32{1, 1}, zoomdelta: [...]float32{1, math.MaxFloat32},
		xscale: [...]float32{1, 1}, rasterx: [...]float32{1, 1}, yscalestart: 100, scalestart: [...]float32{1, 1}, xbottomzoomdelta: math.MaxFloat32,
		zoomscaledelta: [...]float32{math.MaxFloat32, math.MaxFloat32}, actionno: -1, visible: true, active: true, autoresizeparallax: false,
		startrect: [...]int32{-32768, -32768, 65535, 65535}}
}
func readBackGround(is IniSection, link *backGround,
	sff *Sff, at AnimationTable, sProps StageProps) *backGround {
	bg := newBackGround(sff)
	typ := is["type"]
	if len(typ) == 0 {
		return bg
	}
	switch typ[0] {
	case 'N', 'n':
		bg.typ = 0 // normal
	case 'A', 'a':
		bg.typ = 1 // anim
	case 'P', 'p':
		bg.typ = 2 // parallax
	case 'D', 'd':
		bg.typ = 3 // dummy
	default:
		return bg
	}
	var tmp int32
	if is.ReadI32("layerno", &tmp) {
		bg.toplayer = tmp == 1
		if tmp < 0 || tmp > 1 {
			bg.typ = 3
		}
	}
	if bg.typ != 3 {
		var hasAnim bool
		if (bg.typ != 0 || len(is["spriteno"]) == 0) &&
			is.ReadI32("actionno", &bg.actionno) {
			if a := at.get(bg.actionno); a != nil {
				bg.anim = *a
				hasAnim = true
			}
		}
		if hasAnim {
			if bg.typ == 0 {
				bg.typ = 1
			}
		} else {
			var g, n int32
			if is.readI32ForStage("spriteno", &g, &n) {
				bg.anim.frames = []AnimFrame{*newAnimFrame()}
				bg.anim.frames[0].Group, bg.anim.frames[0].Number =
					I32ToI16(g), I32ToI16(n)
			}
			if is.ReadI32("mask", &tmp) {
				if tmp != 0 {
					bg.anim.mask = 0
				} else {
					bg.anim.mask = -1
				}
			}
		}
	}
	is.ReadBool("positionlink", &bg.positionlink)
	if bg.positionlink && link != nil {
		bg.startv = link.startv
		bg.delta = link.delta
	}
	is.ReadBool("autoresizeparallax", &bg.autoresizeparallax)
	is.readF32ForStage("start", &bg.start[0], &bg.start[1])
	if !bg.positionlink {
		is.readF32ForStage("delta", &bg.delta[0], &bg.delta[1])
	}
	is.readF32ForStage("scalestart", &bg.scalestart[0], &bg.scalestart[1])
	is.readF32ForStage("scaledelta", &bg.scaledelta[0], &bg.scaledelta[1])
	is.readF32ForStage("xbottomzoomdelta", &bg.xbottomzoomdelta)
	is.readF32ForStage("zoomscaledelta", &bg.zoomscaledelta[0], &bg.zoomscaledelta[1])
	is.readF32ForStage("zoomdelta", &bg.zoomdelta[0], &bg.zoomdelta[1])
	if bg.zoomdelta[0] != math.MaxFloat32 && bg.zoomdelta[1] == math.MaxFloat32 {
		bg.zoomdelta[1] = bg.zoomdelta[0]
	}
	switch strings.ToLower(is["trans"]) {
	case "add":
		bg.anim.mask = 0
		bg.anim.srcAlpha = 255
		bg.anim.dstAlpha = 255
		s, d := int32(bg.anim.srcAlpha), int32(bg.anim.dstAlpha)
		if is.readI32ForStage("alpha", &s, &d) {
			bg.anim.srcAlpha = int16(Clamp(s, 0, 255))
			bg.anim.dstAlpha = int16(Clamp(d, 0, 255))
			if bg.anim.srcAlpha == 1 && bg.anim.dstAlpha == 255 {
				bg.anim.srcAlpha = 0
			}
		}
	case "add1":
		bg.anim.mask = 0
		bg.anim.srcAlpha = 255
		bg.anim.dstAlpha = ^255
		var s, d int32 = 255, 255
		if is.readI32ForStage("alpha", &s, &d) {
			bg.anim.srcAlpha = int16(Min(255, s))
			//bg.anim.dstAlpha = ^int16(Clamp(d, 0, 255))
			bg.anim.dstAlpha = int16(Clamp(d, 0, 255))
		}
	case "addalpha":
		bg.anim.mask = 0
		s, d := int32(bg.anim.srcAlpha), int32(bg.anim.dstAlpha)
		if is.readI32ForStage("alpha", &s, &d) {
			bg.anim.srcAlpha = int16(Clamp(s, 0, 255))
			bg.anim.dstAlpha = int16(Clamp(d, 0, 255))
			if bg.anim.srcAlpha == 1 && bg.anim.dstAlpha == 255 {
				bg.anim.srcAlpha = 0
			}
		}
	case "sub":
		bg.anim.mask = 0
		bg.anim.srcAlpha = 1
		bg.anim.dstAlpha = 255
	case "none":
		bg.anim.srcAlpha = -1
		bg.anim.dstAlpha = 0
	}
	if is.readI32ForStage("tile", &bg.anim.tile.x, &bg.anim.tile.y) {
		if bg.typ == 2 {
			bg.anim.tile.y = 0
		}
		if bg.anim.tile.x < 0 {
			bg.anim.tile.x = math.MaxInt32
		}
	}
	if bg.typ == 2 {
		if !is.readI32ForStage("width", &bg.width[0], &bg.width[1]) {
			is.readF32ForStage("xscale", &bg.rasterx[0], &bg.rasterx[1])
		}
		is.ReadF32("yscalestart", &bg.yscalestart)
		is.ReadF32("yscaledelta", &bg.yscaledelta)
	} else {
		is.ReadI32("tilespacing", &bg.anim.tile.sx, &bg.anim.tile.sy)
		//bg.anim.tile.sy = bg.anim.tile.sx
		if bg.actionno < 0 && len(bg.anim.frames) > 0 {
			if spr := sff.GetSprite(
				bg.anim.frames[0].Group, bg.anim.frames[0].Number); spr != nil {
				bg.anim.tile.sx += int32(spr.Size[0])
				bg.anim.tile.sy += int32(spr.Size[1])
			}
		} else {
			if bg.anim.tile.sx == 0 {
				bg.anim.tile.x = 0
			}
			if bg.anim.tile.sy == 0 {
				bg.anim.tile.y = 0
			}
		}
	}
	if is.readI32ForStage("window", &bg.startrect[0], &bg.startrect[1],
		&bg.startrect[2], &bg.startrect[3]) {
		bg.startrect[2] = Max(0, bg.startrect[2]+1-bg.startrect[0])
		bg.startrect[3] = Max(0, bg.startrect[3]+1-bg.startrect[1])
		bg.notmaskwindow = 1
	}
	if is.readI32ForStage("maskwindow", &bg.startrect[0], &bg.startrect[1],
		&bg.startrect[2], &bg.startrect[3]) {
		bg.startrect[2] = Max(0, bg.startrect[2]-bg.startrect[0])
		bg.startrect[3] = Max(0, bg.startrect[3]-bg.startrect[1])
		bg.notmaskwindow = 0
	}
	is.readF32ForStage("windowdelta", &bg.windowdelta[0], &bg.windowdelta[1])
	is.ReadI32("id", &bg.id)
	is.readF32ForStage("velocity", &bg.startv[0], &bg.startv[1])
	for i := 0; i < 2; i++ {
		var name string
		if i == 0 {
			name = "sin.x"
		} else {
			name = "sin.y"
		}
		r, slt, st := float32(math.NaN()), float32(math.NaN()), float32(math.NaN())
		if is.readF32ForStage(name, &r, &slt, &st) {
			if !math.IsNaN(float64(r)) {
				bg.startrad[i], bg.bga.radius[i] = r, r
			}
			if !math.IsNaN(float64(slt)) {
				var slti int32
				is.readI32ForStage(name, &tmp, &slti)
				bg.startsinlt[i], bg.bga.sinlooptime[i] = slti, slti
			}
			if bg.bga.sinlooptime[i] > 0 && !math.IsNaN(float64(st)) {
				bg.bga.sintime[i] = int32(st*float32(bg.bga.sinlooptime[i])/360) %
					bg.bga.sinlooptime[i]
				if bg.bga.sintime[i] < 0 {
					bg.bga.sintime[i] += bg.bga.sinlooptime[i]
				}
				bg.startsint[i] = bg.bga.sintime[i]
			}
		}
	}
	if !is.ReadBool("roundpos", &bg.roundpos) {
		bg.roundpos = sProps.roundpos
	}
	return bg
}
func (bg *backGround) reset() {
	bg.palfx.clear()
	bg.anim.Reset()
	bg.bga.clear()
	bg.bga.vel = bg.startv
	bg.bga.radius = bg.startrad
	bg.bga.sintime = bg.startsint
	bg.bga.sinlooptime = bg.startsinlt
	bg.palfx.time = -1
	bg.palfx.invertblend = -3
}
func (bg backGround) draw(pos [2]float32, scl, bgscl, lclscl float32,
	stgscl [2]float32, shakeY float32, isStage bool) {
	if bg.typ == 2 && (bg.width[0] != 0 || bg.width[1] != 0) && bg.anim.spr != nil {
		bg.xscale[0] = float32(bg.width[0]) / float32(bg.anim.spr.Size[0])
		bg.xscale[1] = float32(bg.width[1]) / float32(bg.anim.spr.Size[0])
		bg.xofs = -float32(bg.width[0])/2 + float32(bg.anim.spr.Offset[0])*bg.xscale[0]
	}
	xras := (bg.rasterx[1] - bg.rasterx[0]) / bg.rasterx[0]
	xbs, dx := bg.xscale[1], MaxF(0, bg.delta[0]*bgscl)
	var sclx_recip, sclx, scly float32 = 1, 1, 1
	lscl := [...]float32{lclscl * stgscl[0], lclscl * stgscl[1]}
	if bg.zoomdelta[0] != math.MaxFloat32 {
		sclx = scl + (1-scl)*(1-bg.zoomdelta[0])
		scly = scl + (1-scl)*(1-bg.zoomdelta[1])
		if !bg.autoresizeparallax {
			sclx_recip = (1 + bg.zoomdelta[0]*((1/(sclx*lscl[0])*lscl[0])-1))
		}
	} else {
		sclx = MaxF(0, scl+(1-scl)*(1-dx))
		scly = MaxF(0, scl+(1-scl)*(1-MaxF(0, bg.delta[1]*bgscl)))
	}
	if sclx != 0 && bg.autoresizeparallax {
		tmp := 1 / sclx
		if bg.xbottomzoomdelta != math.MaxFloat32 {
			xbs *= MaxF(0, scl+(1-scl)*(1-bg.xbottomzoomdelta*(xbs/bg.xscale[0]))) * tmp
		} else {
			xbs *= MaxF(0, scl+(1-scl)*(1-dx*(xbs/bg.xscale[0]))) * tmp
		}
		tmp *= MaxF(0, scl+(1-scl)*(1-dx*(xras+1)))
		xras -= tmp - 1
		xbs *= tmp
	}
	var xs3, ys3 float32 = 1, 1
	if bg.zoomscaledelta[0] != math.MaxFloat32 {
		xs3 = ((scl + (1-scl)*(1-bg.zoomscaledelta[0])) / sclx)
	}
	if bg.zoomscaledelta[1] != math.MaxFloat32 {
		ys3 = ((scl + (1-scl)*(1-bg.zoomscaledelta[1])) / scly)
	}
	// This handles the flooring of the camera position in MUGEN versions earlier than 1.0.
	if bg.roundpos {
		for i := 0; i < 2; i++ {
			pos[i] = float32(math.Floor(float64(pos[i])))
		}
	}
	x := bg.start[0] + bg.xofs - (pos[0]/stgscl[0])*bg.delta[0] +
		bg.bga.offset[0]
	// Hires breaks ydelta scrolling vel, so bgscl was commented from here.
	yScrollPos := (pos[1] / scl / stgscl[1]) * bg.delta[1] // * bgscl
	y := bg.start[1] - yScrollPos + bg.bga.offset[1]
	ys2 := bg.scaledelta[1] * pos[1] * bg.delta[1] * bgscl
	ys := ((100-(pos[1])*bg.yscaledelta)*bgscl/bg.yscalestart)*bg.scalestart[1] + ys2
	xs := bg.scaledelta[0] * pos[0] * bg.delta[0] * bgscl
	x *= bgscl
	if isStage {
		zoff := float32(sys.cam.zoffset) * lclscl
		y = y*bgscl + ((zoff-shakeY)/scly-zoff)/lclscl/stgscl[1]
		y -= sys.cam.aspectcorrection / (scly * lclscl * stgscl[1])
		y -= sys.cam.zoomanchorcorrection / (scly * lclscl * stgscl[1])
	} else {
		y = y*bgscl + ((float32(sys.gameHeight)-shakeY)/lclscl/scly-240)/stgscl[1]
	}

	sclx *= lscl[0]
	scly *= lclscl * stgscl[1]
	rect := bg.startrect
	var wscl [2]float32
	for i := range wscl {
		if bg.zoomdelta[i] != math.MaxFloat32 {
			wscl[i] = MaxF(0, scl+(1-scl)*(1-MaxF(0, bg.zoomdelta[i]))) *
				bgscl * lscl[i]
		} else {
			wscl[i] = MaxF(0, scl+(1-scl)*(1-MaxF(0, bg.windowdelta[i]*bgscl))) *
				bgscl * lscl[i]
		}
	}
	startrect0 := (float32(rect[0]) - (pos[0])*bg.windowdelta[0] + (float32(sys.gameWidth)/2/sclx - float32(bg.notmaskwindow)*(float32(sys.gameWidth)/2)*(1/lscl[0]))) * sys.widthScale * wscl[0]
	if !isStage && wscl[0] == 1 {
		startrect0 += float32(sys.gameWidth-320) / 2 * sys.widthScale
	}
	startrect1 := ((float32(rect[1])-pos[1]*bg.windowdelta[1]+(float32(sys.gameHeight)/scly-240))*wscl[1] - shakeY) * sys.heightScale
	rect[0] = int32(math.Floor(float64(startrect0)))
	rect[1] = int32(math.Floor(float64(startrect1)))
	rect[2] = int32(math.Floor(float64(startrect0 + (float32(rect[2]) * sys.widthScale * wscl[0]) - float32(rect[0]))))
	rect[3] = int32(math.Floor(float64(startrect1 + (float32(rect[3]) * sys.heightScale * wscl[1]) - float32(rect[1]))))
	if rect[0] < sys.scrrect[2] && rect[1] < sys.scrrect[3] && rect[0]+rect[2] > 0 && rect[1]+rect[3] > 0 {
		bg.anim.Draw(&rect, x, y, sclx, scly, bg.xscale[0]*bgscl*(bg.scalestart[0]+xs)*xs3, xbs*bgscl*(bg.scalestart[0]+xs)*xs3, ys*ys3,
			xras*x/(AbsF(ys*ys3)*lscl[1]*float32(bg.anim.spr.Size[1])*bg.scalestart[1])*sclx_recip*bg.scalestart[1],
			Rotation{}, float32(sys.gameWidth)/2, bg.palfx, true, 1, false, 1, 0, 0)
	}
}

type bgCtrl struct {
	bg           []*backGround
	currenttime  int32
	starttime    int32
	endtime      int32
	looptime     int32
	_type        BgcType
	x, y         float32
	v            [3]int32
	src          [2]int32
	dst          [2]int32
	add          [3]int32
	mul          [3]int32
	sinadd       [4]int32
	sinmul       [4]int32
	sincolor     [2]int32
	sinhue       [2]int32
	invall       bool
	invblend     int32
	color        float32
	hue          float32
	positionlink bool
	idx          int
	sctrlid      int32
}

func newBgCtrl() *bgCtrl {
	return &bgCtrl{looptime: -1, x: float32(math.NaN()), y: float32(math.NaN())}
}
func (bgc *bgCtrl) read(is IniSection, idx int) {
	bgc.idx = idx
	xy := false
	srcdst := false
	palfx := false
	switch strings.ToLower(is["type"]) {
	case "anim":
		bgc._type = BT_Anim
	case "visible":
		bgc._type = BT_Visible
	case "enable":
		bgc._type = BT_Enable
	case "null":
		bgc._type = BT_Null
	case "palfx":
		bgc._type = BT_PalFX
		palfx = true
		// Default values for PalFX
		bgc.add = [3]int32{0, 0, 0}
		bgc.mul = [3]int32{256, 256, 256}
		bgc.sinadd = [4]int32{0, 0, 0, 0}
		bgc.sinmul = [4]int32{0, 0, 0, 0}
		bgc.sincolor = [2]int32{0, 0}
		bgc.sinhue = [2]int32{0, 0}
		bgc.invall = false
		bgc.invblend = 0
		bgc.color = 1
		bgc.hue = 0
	case "posset":
		bgc._type = BT_PosSet
		xy = true
	case "posadd":
		bgc._type = BT_PosAdd
		xy = true
	case "remappal":
		bgc._type = BT_RemapPal
		srcdst = true
		// Default values for RemapPal
		bgc.src = [2]int32{-1, 0}
		bgc.dst = [2]int32{-1, 0}
	case "sinx":
		bgc._type = BT_SinX
	case "siny":
		bgc._type = BT_SinY
	case "velset":
		bgc._type = BT_VelSet
		xy = true
	case "veladd":
		bgc._type = BT_VelAdd
		xy = true
	}
	is.ReadI32("time", &bgc.starttime)
	bgc.endtime = bgc.starttime
	is.readI32ForStage("time", &bgc.starttime, &bgc.endtime, &bgc.looptime)
	is.ReadBool("positionlink", &bgc.positionlink)
	if xy {
		is.readF32ForStage("x", &bgc.x)
		is.readF32ForStage("y", &bgc.y)
	} else if srcdst {
		is.readI32ForStage("source", &bgc.src[0], &bgc.src[1])
		is.readI32ForStage("dest", &bgc.dst[0], &bgc.dst[1])
	} else if palfx {
		is.readI32ForStage("add", &bgc.add[0], &bgc.add[1], &bgc.add[2])
		is.readI32ForStage("mul", &bgc.mul[0], &bgc.mul[1], &bgc.mul[2])
		if is.readI32ForStage("sinadd", &bgc.sinadd[0], &bgc.sinadd[1], &bgc.sinadd[2], &bgc.sinadd[3]) {
			if bgc.sinadd[3] < 0 {
				for i := 0; i < 4; i++ {
					bgc.sinadd[i] = -bgc.sinadd[i]
				}
			}
		}
		if is.readI32ForStage("sinmul", &bgc.sinmul[0], &bgc.sinmul[1], &bgc.sinmul[2], &bgc.sinmul[3]) {
			if bgc.sinmul[3] < 0 {
				for i := 0; i < 4; i++ {
					bgc.sinmul[i] = -bgc.sinmul[i]
				}
			}
		}
		if is.readI32ForStage("sincolor", &bgc.sincolor[0], &bgc.sincolor[1]) {
			if bgc.sincolor[1] < 0 {
				bgc.sincolor[0] = -bgc.sincolor[0]
			}
		}
		if is.readI32ForStage("sinhue", &bgc.sinhue[0], &bgc.sinhue[1]) {
			if bgc.sinhue[1] < 0 {
				bgc.sinhue[0] = -bgc.sinhue[0]
			}
		}
		var tmp int32
		if is.ReadI32("invertall", &tmp) {
			bgc.invall = tmp != 0
		}
		if is.ReadI32("invertblend", &bgc.invblend) {
			bgc.invblend = bgc.invblend
		}
		if is.ReadF32("color", &bgc.color) {
			bgc.color = bgc.color / 256
		}
		if is.ReadF32("hue", &bgc.hue) {
			bgc.hue = bgc.hue / 256
		}
	} else if is.ReadF32("value", &bgc.x) {
		is.readI32ForStage("value", &bgc.v[0], &bgc.v[1], &bgc.v[2])
	}
	is.ReadI32("sctrlid", &bgc.sctrlid)
}
func (bgc *bgCtrl) xEnable() bool {
	return !math.IsNaN(float64(bgc.x))
}
func (bgc *bgCtrl) yEnable() bool {
	return !math.IsNaN(float64(bgc.y))
}

type bgctNode struct {
	bgc      []*bgCtrl
	waitTime int32
}
type bgcTimeLine struct {
	line []bgctNode
	al   []*bgCtrl
}

func (bgct *bgcTimeLine) clear() {
	*bgct = bgcTimeLine{}
}
func (bgct *bgcTimeLine) add(bgc *bgCtrl) {
	if bgc.looptime >= 0 && bgc.endtime > bgc.looptime {
		bgc.endtime = bgc.looptime
	}
	if bgc.starttime < 0 || bgc.starttime > bgc.endtime ||
		bgc.looptime >= 0 && bgc.starttime >= bgc.looptime {
		return
	}
	wtime := int32(0)
	if bgc.currenttime != 0 {
		if bgc.looptime < 0 {
			return
		}
		wtime += bgc.looptime - bgc.currenttime
	}
	wtime += bgc.starttime
	bgc.currenttime = bgc.starttime
	if wtime < 0 {
		bgc.currenttime -= wtime
		wtime = 0
	}
	i := 0
	for ; ; i++ {
		if i == len(bgct.line) {
			bgct.line = append(bgct.line,
				bgctNode{bgc: []*bgCtrl{bgc}, waitTime: wtime})
			return
		}
		if wtime <= bgct.line[i].waitTime {
			break
		}
		wtime -= bgct.line[i].waitTime
	}
	if wtime == bgct.line[i].waitTime {
		bgct.line[i].bgc = append(bgct.line[i].bgc, bgc)
	} else {
		bgct.line[i].waitTime -= wtime
		bgct.line = append(bgct.line, bgctNode{})
		copy(bgct.line[i+1:], bgct.line[i:])
		bgct.line[i] = bgctNode{bgc: []*bgCtrl{bgc}, waitTime: wtime}
	}
}
func (bgct *bgcTimeLine) step(s *Stage) {
	if len(bgct.line) > 0 && bgct.line[0].waitTime <= 0 {
		for _, b := range bgct.line[0].bgc {
			for i, a := range bgct.al {
				if b.idx < a.idx {
					bgct.al = append(bgct.al, nil)
					copy(bgct.al[i+1:], bgct.al[i:])
					bgct.al[i] = b
					b = nil
					break
				}
			}
			if b != nil {
				bgct.al = append(bgct.al, b)
			}
		}
		bgct.line = bgct.line[1:]
	}
	if len(bgct.line) > 0 {
		bgct.line[0].waitTime--
	}
	var el []*bgCtrl
	for i := 0; i < len(bgct.al); {
		s.runBgCtrl(bgct.al[i])
		if bgct.al[i].currenttime > bgct.al[i].endtime {
			el = append(el, bgct.al[i])
			bgct.al = append(bgct.al[:i], bgct.al[i+1:]...)
			continue
		}
		i++
	}
	for _, b := range el {
		bgct.add(b)
	}
}

type stageShadow struct {
	intensity int32
	color     uint32
	yscale    float32
	fadeend   int32
	fadebgn   int32
	xshear    float32
}
type stagePlayer struct {
	startx, starty, startz int32
}
type Stage struct {
	def              string
	bgmusic          string
	name             string
	displayname      string
	author           string
	nameLow          string
	displaynameLow   string
	authorLow        string
	attachedchardef  []string
	sff              *Sff
	at               AnimationTable
	bg               []*backGround
	bgc              []bgCtrl
	bgct             bgcTimeLine
	bga              bgAction
	sdw              stageShadow
	p                [2]stagePlayer
	leftbound        float32
	rightbound       float32
	screenleft       int32
	screenright      int32
	zoffsetlink      int32
	reflection       int32
	hires            bool
	autoturn         bool
	resetbg          bool
	debugbg          bool
	bgclearcolor     [3]int32
	localscl         float32
	scale            [2]float32
	bgmvolume        int32
	bgmloopstart     int32
	bgmloopend       int32
	bgmstartposition int32
	bgmfreqmul       float32
	bgmratiolife     int32
	bgmtriggerlife   int32
	bgmtriggeralt    int32
	mainstage        bool
	stageCamera      stageCamera
	stageTime        int32
	constants        map[string]float32
	p1p3dist         float32
	mugenver         [2]uint16
	reload           bool
	stageprops       StageProps
	model            *Model
	ikemenver        [3]uint16
}

func newStage(def string) *Stage {
	s := &Stage{def: def, leftbound: -1000,
		rightbound: 1000, screenleft: 15, screenright: 15,
		zoffsetlink: -1, autoturn: true, resetbg: true, localscl: 1,
		scale:        [...]float32{float32(math.NaN()), float32(math.NaN())},
		bgmratiolife: 30, stageCamera: *newStageCamera(),
		constants: make(map[string]float32), p1p3dist: 25, bgmvolume: 100}
	s.sdw.intensity = 128
	s.sdw.color = 0x808080
	s.sdw.yscale = 0.4
	s.p[0].startx, s.p[1].startx = -70, 70
	s.stageprops = newStageProps()
	return s
}
func loadStage(def string, main bool) (*Stage, error) {
	s := newStage(def)
	str, err := LoadText(def)
	if err != nil {
		return nil, err
	}
	s.sff = &Sff{}
	lines, i := SplitAndTrim(str, "\n"), 0
	s.at = ReadAnimationTable(s.sff, &s.sff.palList, lines, &i)
	i = 0
	defmap := make(map[string][]IniSection)
	for i < len(lines) {
		is, name, _ := ReadIniSection(lines, &i)
		if i := strings.IndexAny(name, " \t"); i >= 0 {
			if name[:i] == "bg" {
				defmap["bg"] = append(defmap["bg"], is)
			}
		} else {
			defmap[name] = append(defmap[name], is)
		}
	}
	if sec := defmap["info"]; len(sec) > 0 {
		var ok bool
		s.name, ok, _ = sec[0].getText("name")
		if !ok {
			s.name = def
		}
		s.displayname, ok, _ = sec[0].getText("displayname")
		if !ok {
			s.displayname = s.name
		}
		s.author, _, _ = sec[0].getText("author")
		s.nameLow = strings.ToLower(s.name)
		s.displaynameLow = strings.ToLower(s.displayname)
		s.authorLow = strings.ToLower(s.author)
		s.mugenver = [2]uint16{}
		if str, ok := sec[0]["mugenversion"]; ok {
			for k, v := range SplitAndTrim(str, ".") {
				if k >= len(s.mugenver) {
					break
				}
				if v, err := strconv.ParseUint(v, 10, 16); err == nil {
					s.mugenver[k] = uint16(v)
				} else {
					break
				}
			}
		}
		s.ikemenver = [3]uint16{}
		if str, ok := sec[0]["ikemenversion"]; ok {
			for k, v := range SplitAndTrim(str, ".") {
				if k >= len(s.ikemenver) {
					break
				}
				if v, err := strconv.ParseUint(v, 10, 16); err == nil {
					s.ikemenver[k] = uint16(v)
				} else {
					break
				}
			}
		}
		// If the MUGEN version is lower than 1.0, default to camera pixel rounding (floor)
		if s.ikemenver[0] == 0 && s.ikemenver[1] == 0 && s.mugenver[0] != 1 {
			s.stageprops.roundpos = true
		}
		if sec[0].LoadFile("attachedchar", []string{def, "", sys.motifDir, "data/"}, func(filename string) error {
			s.attachedchardef = append(s.attachedchardef, filename)
			return nil
		}); err != nil {
			return nil, err
		}
		if main {
			r, _ := regexp.Compile("^round[0-9]+def$")
			for k, v := range sec[0] {
				if r.MatchString(k) {
					re := regexp.MustCompile("[0-9]+")
					submatchall := re.FindAllString(k, -1)
					if len(submatchall) == 1 {
						if err := LoadFile(&v, []string{def, "", sys.motifDir, "data/"}, func(filename string) error {
							if sys.stageList[Atoi(submatchall[0])], err = loadStage(filename, false); err != nil {
								return fmt.Errorf("failed to load %v:\n%v", filename, err)
							}
							return nil
						}); err != nil {
							return nil, err
						}
					}
				}
			}
			sec[0].ReadBool("roundloop", &sys.stageLoop)
		}
	}
	if sec := defmap["constants"]; len(sec) > 0 {
		for key, value := range sec[0] {
			s.constants[key] = float32(Atof(value))
		}
	}
	if sec := defmap["playerinfo"]; len(sec) > 0 {
		sec[0].ReadI32("p1startx", &s.p[0].startx)
		sec[0].ReadI32("p1starty", &s.p[0].starty)
		sec[0].ReadI32("p1startz", &s.p[0].startz)
		sec[0].ReadI32("p2startx", &s.p[1].startx)
		sec[0].ReadI32("p2starty", &s.p[1].starty)
		sec[0].ReadI32("p2startz", &s.p[1].startz)
		sec[0].ReadF32("leftbound", &s.leftbound)
		sec[0].ReadF32("rightbound", &s.rightbound)
		sec[0].ReadF32("p1p3dist", &s.p1p3dist)
	}
	if sec := defmap["scaling"]; len(sec) > 0 {
		if s.mugenver[0] != 1 { //mugen 1.0+ removed support for topscale
			sec[0].ReadF32("topscale", &s.stageCamera.ztopscale)
		}
	}
	if sec := defmap["bound"]; len(sec) > 0 {
		sec[0].ReadI32("screenleft", &s.screenleft)
		sec[0].ReadI32("screenright", &s.screenright)
	}
	if sec := defmap["stageinfo"]; len(sec) > 0 {
		sec[0].ReadI32("zoffset", &s.stageCamera.zoffset)
		sec[0].ReadI32("zoffsetlink", &s.zoffsetlink)
		sec[0].ReadBool("hires", &s.hires)
		sec[0].ReadBool("autoturn", &s.autoturn)
		sec[0].ReadBool("resetbg", &s.resetbg)
		sec[0].readI32ForStage("localcoord", &s.stageCamera.localcoord[0],
			&s.stageCamera.localcoord[1])
		sec[0].ReadF32("xscale", &s.scale[0])
		sec[0].ReadF32("yscale", &s.scale[1])
	}
	if math.IsNaN(float64(s.scale[0])) {
		s.scale[0] = 1
	} else if s.hires {
		s.scale[0] *= 2
	}
	if math.IsNaN(float64(s.scale[1])) {
		s.scale[1] = 1
	} else if s.hires {
		s.scale[1] *= 2
	}
	s.localscl = float32(sys.gameWidth) / float32(s.stageCamera.localcoord[0])
	s.stageCamera.localscl = s.localscl
	if sec := defmap["camera"]; len(sec) > 0 {
		sec[0].ReadI32("startx", &s.stageCamera.startx)
		sec[0].ReadI32("starty", &s.stageCamera.starty) //does nothing in mugen
		sec[0].ReadI32("boundleft", &s.stageCamera.boundleft)
		sec[0].ReadI32("boundright", &s.stageCamera.boundright)
		sec[0].ReadI32("boundhigh", &s.stageCamera.boundhigh)
		sec[0].ReadI32("boundlow", &s.stageCamera.boundlow)
		sec[0].ReadF32("verticalfollow", &s.stageCamera.verticalfollow)
		sec[0].ReadI32("floortension", &s.stageCamera.floortension)
		sec[0].ReadI32("tension", &s.stageCamera.tension)
		sec[0].ReadF32("tensionvel", &s.stageCamera.tensionvel)
		sec[0].ReadI32("overdrawhigh", &s.stageCamera.overdrawhigh) //TODO: not implemented
		sec[0].ReadI32("overdrawlow", &s.stageCamera.overdrawlow)
		sec[0].ReadI32("cuthigh", &s.stageCamera.cuthigh)
		sec[0].ReadI32("cutlow", &s.stageCamera.cutlow)
		sec[0].ReadF32("startzoom", &s.stageCamera.startzoom)
		sec[0].ReadF32("fov", &s.stageCamera.fov)
		sec[0].ReadF32("yshift", &s.stageCamera.yshift)
		sec[0].ReadF32("near", &s.stageCamera.near)
		sec[0].ReadF32("far", &s.stageCamera.far)
		sec[0].ReadBool("autocenter", &s.stageCamera.autocenter)
		sec[0].ReadF32("zoomindelay", &s.stageCamera.zoomindelay)
		sec[0].ReadF32("boundhighdelta", &s.stageCamera.boundhighdelta)
		sec[0].ReadBool("lowestcap", &s.stageCamera.lowestcap)
		if sys.cam.ZoomMax == 0 {
			sec[0].ReadF32("zoomin", &s.stageCamera.zoomin)
		} else {
			s.stageCamera.zoomin = sys.cam.ZoomMax
		}
		if sys.cam.ZoomMin == 0 {
			sec[0].ReadF32("zoomout", &s.stageCamera.zoomout)
		} else {
			s.stageCamera.zoomout = sys.cam.ZoomMin
		}
		anchor, _, _ := sec[0].getText("zoomanchor")
		if strings.ToLower(anchor) == "bottom" {
			s.stageCamera.zoomanchor = true
		}
		if sec[0].ReadI32("tensionlow", &s.stageCamera.tensionlow) {
			s.stageCamera.ytensionenable = true
			sec[0].ReadI32("tensionhigh", &s.stageCamera.tensionhigh)
		}
	}
	if sec := defmap["music"]; len(sec) > 0 {
		s.bgmusic = sec[0]["bgmusic"]
		sec[0].ReadI32("bgmvolume", &s.bgmvolume)
		sec[0].ReadI32("bgmloopstart", &s.bgmloopstart)
		sec[0].ReadI32("bgmloopend", &s.bgmloopend)
		sec[0].ReadI32("bgmstartposition", &s.bgmstartposition)
		sec[0].ReadF32("bgmfreqmul", &s.bgmfreqmul)
		sec[0].ReadI32("bgmratio.life", &s.bgmratiolife)
		sec[0].ReadI32("bgmtrigger.life", &s.bgmtriggerlife)
		sec[0].ReadI32("bgmtrigger.alt", &s.bgmtriggeralt)
	}
	if sec := defmap["bgdef"]; len(sec) > 0 {
		if sec[0].LoadFile("spr", []string{def, "", sys.motifDir, "data/"}, func(filename string) error {
			sff, err := loadSff(filename, false)
			if err != nil {
				return err
			}
			*s.sff = *sff
			// SFF v2.01 was not available before Mugen 1.1, therefore we assume that's the minimum correct version for the stage
			if s.sff.header.Ver0 == 2 && s.sff.header.Ver2 == 1 {
				s.mugenver[0] = 1
				s.mugenver[1] = 1
			}
			return nil
		}); err != nil {
			return nil, err
		}
		if err = sec[0].LoadFile("model", []string{def, "", sys.motifDir, "data/"}, func(filename string) error {
			model, err := loadglTFStage(filename)
			if err != nil {
				return err
			}
			s.model = &Model{}
			*s.model = *model
			s.model.pfx = newPalFX()
			s.model.pfx.clear()
			s.model.pfx.time = -1
			// 3D models were not available before Ikemen 1.0, therefore we assume that's the minimum correct version for the stage
			if s.ikemenver[0] == 0 && s.ikemenver[1] == 0 {
				s.ikemenver[0] = 1
				s.ikemenver[1] = 0
			}
			return nil
		}); err != nil {
			return nil, err
		}
		sec[0].ReadBool("debugbg", &s.debugbg)
		sec[0].readI32ForStage("bgclearcolor", &s.bgclearcolor[0], &s.bgclearcolor[1], &s.bgclearcolor[2])
		sec[0].ReadBool("roundpos", &s.stageprops.roundpos)
	}
	if sec := defmap["model"]; len(sec) > 0 && s.model != nil {
		if str, ok := sec[0]["offset"]; ok {
			for k, v := range SplitAndTrim(str, ",") {
				if k >= len(s.model.offset) {
					break
				}
				if v, err := strconv.ParseFloat(v, 32); err == nil {
					s.model.offset[k] = float32(v)
				} else {
					break
				}
			}
		}
		posMul := float32(math.Tan(float64(s.stageCamera.fov*math.Pi/180)/2)) * -s.model.offset[2] / (float32(s.stageCamera.localcoord[1]) / 2)
		s.stageCamera.zoffset = int32(float32(s.stageCamera.localcoord[1])/2 - s.model.offset[1]/posMul - s.stageCamera.yshift*float32(sys.scrrect[3]/2)/float32(sys.gameHeight)*float32(s.stageCamera.localcoord[1])/sys.heightScale)
		if str, ok := sec[0]["scale"]; ok {
			for k, v := range SplitAndTrim(str, ",") {
				if k >= len(s.model.scale) {
					break
				}
				if v, err := strconv.ParseFloat(v, 32); err == nil {
					s.model.scale[k] = float32(v)
				} else {
					break
				}
			}
		}
	}
	reflect := true
	if sec := defmap["shadow"]; len(sec) > 0 {
		var tmp int32
		if sec[0].ReadI32("intensity", &tmp) {
			s.sdw.intensity = Clamp(tmp, 0, 255)
		}
		var r, g, b int32
		sec[0].readI32ForStage("color", &r, &g, &b)
		r, g, b = Clamp(r, 0, 255), Clamp(g, 0, 255), Clamp(b, 0, 255)
		// Disable color parameter specifically in Mugen 1.1 stages
		if s.ikemenver[0] == 0 && s.ikemenver[1] == 0 && s.mugenver[0] == 1 && s.mugenver[1] == 1 {
			r, g, b = 0, 0, 0
		}
		s.sdw.color = uint32(r<<16 | g<<8 | b)
		sec[0].ReadF32("yscale", &s.sdw.yscale)
		sec[0].ReadBool("reflect", &reflect)
		sec[0].readI32ForStage("fade.range", &s.sdw.fadeend, &s.sdw.fadebgn)
		sec[0].ReadF32("xshear", &s.sdw.xshear)
	}
	if reflect {
		if sec := defmap["reflection"]; len(sec) > 0 {
			var tmp int32
			if sec[0].ReadI32("intensity", &tmp) {
				s.reflection = Clamp(tmp, 0, 255)
			}
		}
	}
	var bglink *backGround
	for _, bgsec := range defmap["bg"] {
		if len(s.bg) > 0 && !s.bg[len(s.bg)-1].positionlink {
			bglink = s.bg[len(s.bg)-1]
		}
		s.bg = append(s.bg, readBackGround(bgsec, bglink,
			s.sff, s.at, s.stageprops))
	}
	bgcdef := *newBgCtrl()
	i = 0
	for i < len(lines) {
		is, name, _ := ReadIniSection(lines, &i)
		if len(name) > 0 && name[len(name)-1] == ' ' {
			name = name[:len(name)-1]
		}
		switch name {
		case "bgctrldef":
			bgcdef.bg, bgcdef.looptime = nil, -1
			if ids := is.readI32CsvForStage("ctrlid"); len(ids) > 0 &&
				(len(ids) > 1 || ids[0] != -1) {
				kishutu := make(map[int32]bool)
				for _, id := range ids {
					if kishutu[id] {
						continue
					}
					bgcdef.bg = append(bgcdef.bg, s.getBg(id)...)
					kishutu[id] = true
				}
			} else {
				bgcdef.bg = append(bgcdef.bg, s.bg...)
			}
			is.ReadI32("looptime", &bgcdef.looptime)
		case "bgctrl":
			bgc := newBgCtrl()
			*bgc = bgcdef
			if ids := is.readI32CsvForStage("ctrlid"); len(ids) > 0 {
				bgc.bg = nil
				if len(ids) > 1 || ids[0] != -1 {
					kishutu := make(map[int32]bool)
					for _, id := range ids {
						if kishutu[id] {
							continue
						}
						bgc.bg = append(bgc.bg, s.getBg(id)...)
						kishutu[id] = true
					}
				} else {
					bgc.bg = append(bgc.bg, s.bg...)
				}
			}
			bgc.read(is, len(s.bgc))
			s.bgc = append(s.bgc, *bgc)
		}
	}
	link, zlink := 0, -1
	for i, b := range s.bg {
		if b.positionlink && i > 0 {
			s.bg[i].start[0] += s.bg[link].start[0]
			s.bg[i].start[1] += s.bg[link].start[1]
		} else {
			link = i
		}
		if s.zoffsetlink >= 0 && zlink < 0 && b.id == s.zoffsetlink {
			zlink = i
			s.stageCamera.zoffset += int32(b.start[1] * s.scale[1])
		}
	}

	s.mainstage = main
	return s, nil
}
func (s *Stage) copyStageVars(src *Stage) {
	s.stageCamera.boundleft = src.stageCamera.boundleft
	s.stageCamera.boundright = src.stageCamera.boundright
	s.stageCamera.boundhigh = src.stageCamera.boundhigh
	s.stageCamera.boundlow = src.stageCamera.boundlow
	s.stageCamera.verticalfollow = src.stageCamera.verticalfollow
	s.stageCamera.floortension = src.stageCamera.floortension
	s.stageCamera.tensionhigh = src.stageCamera.tensionhigh
	s.stageCamera.tensionlow = src.stageCamera.tensionlow
	s.stageCamera.tension = src.stageCamera.tension
	s.stageCamera.startzoom = src.stageCamera.startzoom
	s.stageCamera.zoomout = src.stageCamera.zoomout
	s.stageCamera.zoomin = src.stageCamera.zoomin
	s.stageCamera.ytensionenable = src.stageCamera.ytensionenable
	s.leftbound = src.leftbound
	s.rightbound = src.rightbound
	s.stageCamera.ztopscale = src.stageCamera.ztopscale
	s.screenleft = src.screenleft
	s.screenright = src.screenright
	s.stageCamera.zoffset = src.stageCamera.zoffset
	s.zoffsetlink = src.zoffsetlink
	s.scale[0] = src.scale[0]
	s.scale[1] = src.scale[1]
	s.sdw.intensity = src.sdw.intensity
	s.sdw.color = src.sdw.color
	s.sdw.yscale = src.sdw.yscale
	s.sdw.fadeend = src.sdw.fadeend
	s.sdw.fadebgn = src.sdw.fadebgn
	s.sdw.xshear = src.sdw.xshear
	s.reflection = src.reflection
}
func (s *Stage) getBg(id int32) (bg []*backGround) {
	if id >= 0 {
		for _, b := range s.bg {
			if b.id == id {
				bg = append(bg, b)
			}
		}
	}
	return
}
func (s *Stage) runBgCtrl(bgc *bgCtrl) {
	bgc.currenttime++
	switch bgc._type {
	case BT_Anim:
		a := s.at.get(bgc.v[0])
		if a != nil {
			for i := range bgc.bg {
				masktemp := bgc.bg[i].anim.mask
				srcAlphatemp := bgc.bg[i].anim.srcAlpha
				dstAlphatemp := bgc.bg[i].anim.dstAlpha
				tiletmp := bgc.bg[i].anim.tile
				bgc.bg[i].actionno = bgc.v[0]
				bgc.bg[i].anim = *a
				bgc.bg[i].anim.tile = tiletmp
				bgc.bg[i].anim.dstAlpha = dstAlphatemp
				bgc.bg[i].anim.srcAlpha = srcAlphatemp
				bgc.bg[i].anim.mask = masktemp
			}
		}
	case BT_Visible:
		for i := range bgc.bg {
			bgc.bg[i].visible = bgc.v[0] != 0
		}
	case BT_Enable:
		for i := range bgc.bg {
			bgc.bg[i].visible, bgc.bg[i].active = bgc.v[0] != 0, bgc.v[0] != 0
		}
	case BT_PalFX:
		for i := range bgc.bg {
			bgc.bg[i].palfx.add = bgc.add
			bgc.bg[i].palfx.mul = bgc.mul
			bgc.bg[i].palfx.sinadd[0] = bgc.sinadd[0]
			bgc.bg[i].palfx.sinadd[1] = bgc.sinadd[1]
			bgc.bg[i].palfx.sinadd[2] = bgc.sinadd[2]
			bgc.bg[i].palfx.cycletime[0] = bgc.sinadd[3]
			bgc.bg[i].palfx.sinmul[0] = bgc.sinmul[0]
			bgc.bg[i].palfx.sinmul[1] = bgc.sinmul[1]
			bgc.bg[i].palfx.sinmul[2] = bgc.sinmul[2]
			bgc.bg[i].palfx.cycletime[1] = bgc.sinmul[3]
			bgc.bg[i].palfx.sincolor = bgc.sincolor[0]
			bgc.bg[i].palfx.cycletime[2] = bgc.sincolor[1]
			bgc.bg[i].palfx.sinhue = bgc.sinhue[0]
			bgc.bg[i].palfx.cycletime[3] = bgc.sinhue[1]
			bgc.bg[i].palfx.invertall = bgc.invall
			bgc.bg[i].palfx.invertblend = bgc.invblend
			bgc.bg[i].palfx.color = bgc.color
			bgc.bg[i].palfx.hue = bgc.hue
		}
	case BT_PosSet:
		for i := range bgc.bg {
			if bgc.xEnable() {
				bgc.bg[i].bga.pos[0] = bgc.x
			}
			if bgc.yEnable() {
				bgc.bg[i].bga.pos[1] = bgc.y
			}
		}
		if bgc.positionlink {
			if bgc.xEnable() {
				s.bga.pos[0] = bgc.x
			}
			if bgc.yEnable() {
				s.bga.pos[1] = bgc.y
			}
		}
	case BT_PosAdd:
		for i := range bgc.bg {
			if bgc.xEnable() {
				bgc.bg[i].bga.pos[0] += bgc.x
			}
			if bgc.yEnable() {
				bgc.bg[i].bga.pos[1] += bgc.y
			}
		}
		if bgc.positionlink {
			if bgc.xEnable() {
				s.bga.pos[0] += bgc.x
			}
			if bgc.yEnable() {
				s.bga.pos[1] += bgc.y
			}
		}
	case BT_RemapPal:
		if bgc.src[0] >= 0 && bgc.src[1] >= 0 && bgc.dst[1] >= 0 {
			// Get source pal
			si, ok := s.sff.palList.PalTable[[...]int16{int16(bgc.src[0]), int16(bgc.src[1])}]
			if !ok || si < 0 {
				return
			}
			var di int
			if bgc.dst[0] < 0 {
				// Set dest pal to source pal (remap gets reset)
				di = si
			} else {
				// Get dest pal
				di, ok = s.sff.palList.PalTable[[...]int16{int16(bgc.dst[0]), int16(bgc.dst[1])}]
				if !ok || di < 0 {
					return
				}
			}
			s.sff.palList.Remap(si, di)
		}
	case BT_SinX, BT_SinY:
		ii := Btoi(bgc._type == BT_SinY)
		if bgc.v[0] == 0 {
			bgc.v[1] = 0
		}
		// Unlike plain sin.x elements, in the SinX BGCtrl the last parameter is a time offset rather than a phase
		// https://github.com/ikemen-engine/Ikemen-GO/issues/1790
		ph := float32(bgc.v[2]) / float32(bgc.v[1])
		st := int32((ph - float32(int32(ph))) * float32(bgc.v[1]))
		if st < 0 {
			st += Abs(bgc.v[1])
		}
		for i := range bgc.bg {
			bgc.bg[i].bga.radius[ii] = bgc.x
			bgc.bg[i].bga.sinlooptime[ii] = bgc.v[1]
			bgc.bg[i].bga.sintime[ii] = st
		}
		if bgc.positionlink {
			s.bga.radius[ii] = bgc.x
			s.bga.sinlooptime[ii] = bgc.v[1]
			s.bga.sintime[ii] = st
		}
	case BT_VelSet:
		for i := range bgc.bg {
			if bgc.xEnable() {
				bgc.bg[i].bga.vel[0] = bgc.x
			}
			if bgc.yEnable() {
				bgc.bg[i].bga.vel[1] = bgc.y
			}
		}
		if bgc.positionlink {
			if bgc.xEnable() {
				s.bga.vel[0] = bgc.x
			}
			if bgc.yEnable() {
				s.bga.vel[1] = bgc.y
			}
		}
	case BT_VelAdd:
		for i := range bgc.bg {
			if bgc.xEnable() {
				bgc.bg[i].bga.vel[0] += bgc.x
			}
			if bgc.yEnable() {
				bgc.bg[i].bga.vel[1] += bgc.y
			}
		}
		if bgc.positionlink {
			if bgc.xEnable() {
				s.bga.vel[0] += bgc.x
			}
			if bgc.yEnable() {
				s.bga.vel[1] += bgc.y
			}
		}
	}
}
func (s *Stage) action() {
	link, zlink, paused := 0, -1, true
	if sys.tickFrame() && (sys.super <= 0 || !sys.superpausebg) &&
		(sys.pause <= 0 || !sys.pausebg) {
		paused = false
		s.stageTime++
		s.bgct.step(s)
		s.bga.action()
		if s.model != nil {
			s.model.step()
		}
	}
	for i, b := range s.bg {
		b.palfx.step()
		if sys.bgPalFX.enable {
			// TODO: Finish proper synthesization of bgPalFX into PalFX from bg element
			// (Right now, bgPalFX just overrides all unique parameters from BG Elements' PalFX)
			// for j := 0; j < 3; j++ {
			// if sys.bgPalFX.invertall {
			// b.palfx.eAdd[j] = -b.palfx.add[j] * (b.palfx.mul[j]/256) + 256 * (1-(b.palfx.mul[j]/256))
			// b.palfx.eMul[j] = 256
			// }
			// b.palfx.eAdd[j] = int32((float32(b.palfx.eAdd[j])) * sys.bgPalFX.eColor)
			// b.palfx.eMul[j] = int32(float32(b.palfx.eMul[j]) * sys.bgPalFX.eColor + 256*(1-sys.bgPalFX.eColor))
			// }
			// b.palfx.synthesize(sys.bgPalFX)
			b.palfx.eAdd = sys.bgPalFX.eAdd
			b.palfx.eMul = sys.bgPalFX.eMul
			b.palfx.eColor = sys.bgPalFX.eColor
			b.palfx.eHue = sys.bgPalFX.eHue
			b.palfx.eInvertall = sys.bgPalFX.eInvertall
			b.palfx.eInvertblend = sys.bgPalFX.eInvertblend
			b.palfx.eNegType = sys.bgPalFX.eNegType
		}
		if b.active && !paused {
			s.bg[i].bga.action()
			if i > 0 && b.positionlink {
				bgasinoffset0 := s.bg[link].bga.sinoffset[0]
				bgasinoffset1 := s.bg[link].bga.sinoffset[1]
				if s.hires {
					bgasinoffset0 = bgasinoffset0 / 2
					bgasinoffset1 = bgasinoffset1 / 2
				}
				s.bg[i].bga.offset[0] += bgasinoffset0
				s.bg[i].bga.offset[1] += bgasinoffset1
			} else {
				link = i
			}
			if s.zoffsetlink >= 0 && zlink < 0 && b.id == s.zoffsetlink {
				zlink = i
				s.bga.offset[1] += b.bga.offset[1]
			}
			s.bg[i].anim.Action()
		}
	}
	if s.model != nil {
		s.model.pfx.step()
		if sys.bgPalFX.enable {
			s.model.pfx.eAdd = sys.bgPalFX.eAdd
			s.model.pfx.eMul = sys.bgPalFX.eMul
			s.model.pfx.eColor = sys.bgPalFX.eColor
			s.model.pfx.eHue = sys.bgPalFX.eHue
			s.model.pfx.eInvertall = sys.bgPalFX.eInvertall
			s.model.pfx.eInvertblend = sys.bgPalFX.eInvertblend
			s.model.pfx.eNegType = sys.bgPalFX.eNegType
		}
	}
}
func (s *Stage) draw(top bool, x, y, scl float32) {
	bgscl := float32(1)
	if s.hires {
		bgscl = 0.5
	}
	yofs, pos := sys.envShake.getOffset(), [...]float32{x, y}
	scl2 := s.localscl * scl
	// This code makes the background scroll faster when surpassing boundhigh with the camera pushed down
	// through floortension and boundlow. MUGEN 1.1 doesn't look like it does this, so it was commented.
	// var extraBoundH float32
	// if sys.cam.zoomout < 1 {
	// extraBoundH = sys.cam.ExtraBoundH * ((1/scl)-1)/((1/sys.cam.zoomout)-1)
	// }
	// if pos[1] <= float32(s.stageCamera.boundlow) && pos[1] < float32(s.stageCamera.boundhigh)-extraBoundH {
	// yofs += (pos[1]-float32(s.stageCamera.boundhigh))*scl2 +
	// extraBoundH*scl
	// pos[1] = float32(s.stageCamera.boundhigh) - extraBoundH/s.localscl
	// }
	if yofs != 0 && s.stageCamera.verticalfollow > 0 {
		if yofs < 0 {
			tmp := (float32(s.stageCamera.boundhigh) - pos[1]) * scl2
			if scl > 1 {
				tmp += (sys.cam.GroundLevel() + float32(sys.gameHeight-240)) * (1/scl - 1)
			} else {
				tmp += float32(sys.gameHeight) * (1/scl - 1)
			}
			if tmp >= 0 {
			} else if yofs < tmp {
				yofs -= tmp
				pos[1] += tmp / scl2
			} else {
				pos[1] += yofs / scl2
				yofs = 0
			}
		} else {
			if -yofs >= pos[1]*scl2 {
				pos[1] += yofs / scl2
				yofs = 0
			}
		}
	}
	if !sys.cam.ZoomEnable {
		for i, p := range pos {
			pos[i] = float32(math.Ceil(float64(p - 0.5)))
		}
	}
	if !top {
		s.drawModel(pos, yofs, scl, 0)
	} else {
		s.drawModel(pos, yofs, scl, 1)
	}
	for _, b := range s.bg {
		if b.visible && b.toplayer == top && b.anim.spr != nil {
			b.draw(pos, scl, bgscl, s.localscl, s.scale, yofs, true)
		}
	}
}
func (s *Stage) reset() {
	s.sff.palList.ResetRemap()
	s.bga.clear()
	for i := range s.bg {
		s.bg[i].reset()
	}
	for i := range s.bgc {
		s.bgc[i].currenttime = 0
	}
	s.bgct.clear()
	for i := len(s.bgc) - 1; i >= 0; i-- {
		s.bgct.add(&s.bgc[i])
	}
	s.stageTime = 0
	if s.model != nil {
		s.model.reset()
	}
}

func (s *Stage) modifyBGCtrl(id int32, t, v [3]int32, x, y float32, src, dst [2]int32,
	add, mul [3]int32, sinadd [4]int32, sinmul [4]int32, sincolor [2]int32, sinhue [2]int32, invall int32, invblend int32, color float32, hue float32) {
	for i := range s.bgc {
		if id == s.bgc[i].sctrlid {
			if t[0] != IErr {
				s.bgc[i].starttime = t[0]
			}
			if t[1] != IErr {
				s.bgc[i].endtime = t[1]
			}
			if t[2] != IErr {
				s.bgc[i].looptime = t[2]
			}
			for j := 0; j < 3; j++ {
				if v[j] != IErr {
					s.bgc[i].v[j] = v[j]
				}
			}
			if !math.IsNaN(float64(x)) {
				s.bgc[i].x = x
			}
			if !math.IsNaN(float64(y)) {
				s.bgc[i].y = y
			}
			for j := 0; j < 2; j++ {
				if src[j] != IErr {
					s.bgc[i].src[j] = src[j]
				}
				if dst[j] != IErr {
					s.bgc[i].dst[j] = dst[j]
				}
			}
			var side int32 = 1
			if sinadd[3] != IErr {
				if sinadd[3] < 0 {
					sinadd[3] = -sinadd[3]
					side = -1
				}
			}
			var side2 int32 = 1
			if sinmul[3] != IErr {
				if sinmul[3] < 0 {
					sinmul[3] = -sinmul[3]
					side2 = -1
				}
			}
			var side3 int32 = 1
			if sincolor[1] != IErr {
				if sincolor[1] < 0 {
					sincolor[1] = -sincolor[1]
					side3 = -1
				}
			}
			var side4 int32 = 1
			if sinhue[1] != IErr {
				if sinhue[1] < 0 {
					sinhue[1] = -sinhue[1]
					side4 = -1
				}
			}
			for j := 0; j < 4; j++ {
				if j < 3 {
					if add[j] != IErr {
						s.bgc[i].add[j] = add[j]
					}
					if mul[j] != IErr {
						s.bgc[i].mul[j] = mul[j]
					}

				}
				if sinadd[j] != IErr {
					s.bgc[i].sinadd[j] = sinadd[j] * side
				}
				if sinmul[j] != IErr {
					s.bgc[i].sinmul[j] = sinmul[j] * side2
				}
				if j < 2 {
					if sincolor[0] != IErr {
						s.bgc[i].sincolor[j] = sincolor[j] * side3
					}
					if sinhue[0] != IErr {
						s.bgc[i].sinhue[j] = sinhue[j] * side4
					}
				}
			}
			if invall != IErr {
				s.bgc[i].invall = invall != 0
			}
			if invblend != IErr {
				s.bgc[i].invblend = invblend
			}
			if !math.IsNaN(float64(color)) {
				s.bgc[i].color = color / 256
			}
			if !math.IsNaN(float64(hue)) {
				s.bgc[i].hue = hue / 256
			}
			s.reload = true
		}
	}
}

// 3D Stage Related
// TODO: Refactor and move this to a new file?
type Model struct {
	scenes              []*Scene
	nodes               []*Node
	meshes              []*Mesh
	textures            []*GLTFTexture
	materials           []*Material
	offset              [3]float32
	rotation            [3]float32
	scale               [3]float32
	pfx                 *PalFX
	animationTimeStamps map[uint32][]float32
	animations          []*GLTFAnimation
	skins               []*Skin
	vertexBuffer        []byte
	elementBuffer       []uint32
}
type Scene struct {
	nodes []uint32
	name  string
}
type GLTFAnimationType byte

const (
	TRSTranslation = iota
	TRSScale
	TRSRotation
	MorphTargetWeight
)

type GLTFAnimationInterpolation byte

const (
	InterpolationLinear = iota
	InterpolationStep
	InterpolationCubicSpline
)

type GLTFAnimation struct {
	duration float32
	time     float32
	channels []*GLTFAnimationChannel
	samplers []*GLTFAnimationSampler
}
type GLTFAnimationChannel struct {
	path         GLTFAnimationType
	nodeIndex    uint32
	samplerIndex uint32
}
type GLTFAnimationSampler struct {
	inputIndex    uint32
	output        []float32
	interpolation GLTFAnimationInterpolation
}
type GLTFTexture struct {
	tex *Texture
}

type AlphaMode byte

const (
	AlphaModeOpaque = iota
	AlphaModeMask
	AlphaModeBlend
)

type Material struct {
	name            string
	alphaMode       AlphaMode
	alphaCutoff     float32
	textureIndex    *uint32
	baseColorFactor [4]float32
	doubleSided     bool
}
type Trans byte

const (
	TransNone = iota
	TransAdd
	TransReverseSubtract
)

type Node struct {
	meshIndex          *uint32
	transition         [3]float32
	rotation           [4]float32
	scale              [3]float32
	transformChanged   bool
	localTransform     mgl.Mat4
	worldTransform     mgl.Mat4
	childrenIndex      []uint32
	trans              Trans
	zWrite             bool
	zTest              bool
	parentIndex        *int32
	skin               *uint32
	morphTargetWeights []float32
}

type Skin struct {
	joints              []uint32
	inverseBindMatrices []float32
	texture             *GLTFTexture
}
type Mesh struct {
	name               string
	morphTargetWeights []float32
	primitives         []*Primitive
}
type PrimitiveMode byte

const (
	POINTS = iota
	LINES
	LINE_LOOP
	LINE_STRIP
	TRIANGLES
	TRIANGLE_STRIP
	TRIANGLE_FAN
)

type MorphTarget struct {
	positionOffset *uint32
	uvOffset       *uint32
	colorOffset    *uint32
}
type Primitive struct {
	numVertices         uint32
	numIndices          uint32
	vertexBufferOffset  uint32
	elementBufferOffset uint32
	materialIndex       *uint32
	useUV               bool
	useVertexColor      bool
	useJoint0           bool
	useJoint1           bool
	mode                PrimitiveMode
	morphTargets        []*MorphTarget
}

var gltfPrimitiveModeMap = map[gltf.PrimitiveMode]PrimitiveMode{
	gltf.PrimitivePoints:        POINTS,
	gltf.PrimitiveLines:         LINES,
	gltf.PrimitiveLineLoop:      LINE_LOOP,
	gltf.PrimitiveLineStrip:     LINE_STRIP,
	gltf.PrimitiveTriangles:     TRIANGLES,
	gltf.PrimitiveTriangleStrip: TRIANGLE_STRIP,
	gltf.PrimitiveTriangleFan:   TRIANGLE_FAN,
}

func loadglTFStage(filepath string) (*Model, error) {
	mdl := &Model{offset: [3]float32{0, 0, 0}, rotation: [3]float32{0, 0, 0}, scale: [3]float32{1, 1, 1}}
	doc, err := gltf.Open(filepath)
	if err != nil {
		return nil, err
	}
	var images = make([]image.Image, 0, len(doc.Images))
	for _, img := range doc.Images {
		var buffer *bytes.Buffer
		if len(img.URI) > 0 {
			if strings.HasPrefix(img.URI, "data:") {
				if strings.HasPrefix(img.URI, "data:image/png;base64,") {
					decodedData, err := base64.StdEncoding.DecodeString(img.URI[22:])
					if err != nil {
						return nil, err
					}
					buffer = bytes.NewBuffer(decodedData)
				} else {
					decodedData, err := base64.StdEncoding.DecodeString(img.URI[23:])
					if err != nil {
						return nil, err
					}
					buffer = bytes.NewBuffer(decodedData)
				}
			} else {
				if err := LoadFile(&img.URI, []string{filepath, "", sys.motifDir, "data/"}, func(filename string) error {
					data, err := os.ReadFile(filename)
					if err != nil {
						return err
					}
					buffer = bytes.NewBuffer(data)
					return nil
				}); err != nil {
					return nil, err
				}

			}
		} else {
			source, err := modeler.ReadBufferView(doc, doc.BufferViews[*img.BufferView])
			if err != nil {
				return nil, err
			}
			buffer = bytes.NewBuffer(source)
		}
		res, _, err := image.Decode(buffer)
		if err != nil {
			return nil, err
		}
		images = append(images, res)
	}
	mdl.textures = make([]*GLTFTexture, 0, len(doc.Textures))
	textureMap := map[[2]int32]*GLTFTexture{}
	for _, t := range doc.Textures {
		if t.Sampler != nil {
			if texture, ok := textureMap[[2]int32{int32(*t.Source), int32(*t.Sampler)}]; ok {
				mdl.textures = append(mdl.textures, texture)
			} else {
				texture := &GLTFTexture{}
				s := doc.Samplers[*t.Sampler]
				mag, _ := map[gltf.MagFilter]int32{
					gltf.MagUndefined: 9729,
					gltf.MagNearest:   9728,
					gltf.MagLinear:    9729,
				}[s.MagFilter]
				min, _ := map[gltf.MinFilter]int32{
					gltf.MinUndefined:            9729,
					gltf.MinNearest:              9728,
					gltf.MinLinear:               9729,
					gltf.MinNearestMipMapNearest: 9984,
					gltf.MinLinearMipMapNearest:  9985,
					gltf.MinNearestMipMapLinear:  9986,
					gltf.MinLinearMipMapLinear:   9987,
				}[s.MinFilter]
				wrapS, _ := map[gltf.WrappingMode]int32{
					gltf.WrapClampToEdge:    33071,
					gltf.WrapMirroredRepeat: 33648,
					gltf.WrapRepeat:         10497,
				}[s.WrapS]
				wrapT, _ := map[gltf.WrappingMode]int32{
					gltf.WrapClampToEdge:    33071,
					gltf.WrapMirroredRepeat: 33648,
					gltf.WrapRepeat:         10497,
				}[s.WrapT]

				img := images[*t.Source]
				rgba := image.NewRGBA(img.Bounds())
				draw.Draw(rgba, img.Bounds(), img, img.Bounds().Min, draw.Src)
				sys.mainThreadTask <- func() {
					texture.tex = newTexture(int32(img.Bounds().Max.X), int32(img.Bounds().Max.Y), 32, false)
					texture.tex.SetDataG(rgba.Pix, mag, min, wrapS, wrapT)
				}
				textureMap[[2]int32{int32(*t.Source), int32(*t.Sampler)}] = texture
				mdl.textures = append(mdl.textures, texture)
			}
		} else {
			if texture, ok := textureMap[[2]int32{int32(*t.Source), -1}]; ok {
				mdl.textures = append(mdl.textures, texture)
			} else {
				texture := &GLTFTexture{}
				mag := 9728
				min := 9728
				wrapS := 10497
				wrapT := 10497
				img := images[*t.Source]
				rgba := image.NewRGBA(img.Bounds())
				draw.Draw(rgba, img.Bounds(), img, img.Bounds().Min, draw.Src)
				sys.mainThreadTask <- func() {
					texture.tex = newTexture(int32(img.Bounds().Max.X), int32(img.Bounds().Max.Y), 32, false)
					texture.tex.SetDataG(rgba.Pix, int32(mag), int32(min), int32(wrapS), int32(wrapT))
				}
				textureMap[[2]int32{int32(*t.Source), -1}] = texture
				mdl.textures = append(mdl.textures, texture)
			}
		}

	}
	mdl.materials = make([]*Material, 0, len(doc.Materials))
	for _, m := range doc.Materials {
		material := &Material{}
		if m.PBRMetallicRoughness.BaseColorTexture != nil {
			material.textureIndex = new(uint32)
			*material.textureIndex = m.PBRMetallicRoughness.BaseColorTexture.Index
		}
		material.baseColorFactor = *m.PBRMetallicRoughness.BaseColorFactor
		material.name = m.Name
		material.alphaMode, _ = map[gltf.AlphaMode]AlphaMode{
			gltf.AlphaOpaque: AlphaModeOpaque,
			gltf.AlphaMask:   AlphaModeMask,
			gltf.AlphaBlend:  AlphaModeBlend,
		}[m.AlphaMode]
		if material.alphaMode == AlphaModeMask {
			material.alphaCutoff = m.AlphaCutoffOrDefault()
		} else {
			material.alphaCutoff = 0
		}

		material.doubleSided = m.DoubleSided
		mdl.materials = append(mdl.materials, material)
	}
	var vertexBuffer []byte
	var elementBuffer []uint32
	mdl.meshes = make([]*Mesh, 0, len(doc.Meshes))
	for _, m := range doc.Meshes {
		var mesh = &Mesh{}
		mesh.name = m.Name
		mesh.morphTargetWeights = m.Weights
		for _, p := range m.Primitives {
			var primitive = &Primitive{}
			primitive.vertexBufferOffset = uint32(len(vertexBuffer))
			primitive.elementBufferOffset = uint32(4 * len(elementBuffer))
			var posBuffer [][3]float32
			positions, err := modeler.ReadPosition(doc, doc.Accessors[p.Attributes[gltf.POSITION]], posBuffer)
			if err != nil {
				return nil, err
			}
			for _, pos := range positions {
				vertexBuffer = append(vertexBuffer, f32.Bytes(binary.LittleEndian, pos[:]...)...)
			}
			primitive.numVertices = uint32(len(positions))
			if idx, ok := p.Attributes[gltf.TEXCOORD_0]; ok {
				var uvBuffer [][2]float32
				texCoords, err := modeler.ReadTextureCoord(doc, doc.Accessors[idx], uvBuffer)
				if err != nil {
					return nil, err
				}
				if len(texCoords) > 0 {
					primitive.useUV = true
					for _, tex := range texCoords {
						vertexBuffer = append(vertexBuffer, f32.Bytes(binary.LittleEndian, tex[:]...)...)
					}
				} else {
					primitive.useUV = false
				}
			} else {
				primitive.useUV = false
			}
			var indexBuffer []uint32
			indices, err := modeler.ReadIndices(doc, doc.Accessors[*p.Indices], indexBuffer)
			if err != nil {
				return nil, err
			}
			for _, p := range indices {
				elementBuffer = append(elementBuffer, p)
			}
			primitive.numIndices = uint32(len(indices))
			if idx, ok := p.Attributes[gltf.COLOR_0]; ok {
				primitive.useVertexColor = true
				switch doc.Accessors[idx].ComponentType {
				case gltf.ComponentUbyte:
					if doc.Accessors[idx].Type == gltf.AccessorVec3 {
						var vecBuffer [][3]uint8
						vecs, err := modeler.ReadAccessor(doc, doc.Accessors[idx], vecBuffer)
						if err != nil {
							return nil, err
						}
						for _, vec := range vecs.([][3]uint8) {
							vertexBuffer = append(vertexBuffer, f32.Bytes(binary.LittleEndian, float32(vec[0])/255, float32(vec[1])/255, float32(vec[2])/255, 1)...)
						}
					} else {
						var vecBuffer [][4]uint8
						vecs, err := modeler.ReadAccessor(doc, doc.Accessors[idx], vecBuffer)
						if err != nil {
							return nil, err
						}
						for _, vec := range vecs.([][4]uint8) {
							vertexBuffer = append(vertexBuffer, f32.Bytes(binary.LittleEndian, float32(vec[0])/255, float32(vec[1])/255, float32(vec[2])/255, float32(vec[3])/255)...)
						}
					}
				case gltf.ComponentUshort:
					if doc.Accessors[idx].Type == gltf.AccessorVec3 {
						var vecBuffer [][3]uint16
						vecs, err := modeler.ReadAccessor(doc, doc.Accessors[idx], vecBuffer)
						if err != nil {
							return nil, err
						}
						for _, vec := range vecs.([][3]uint16) {
							vertexBuffer = append(vertexBuffer, f32.Bytes(binary.LittleEndian, float32(vec[0])/65535, float32(vec[1])/65535, float32(vec[2])/65535, 1)...)
						}
					} else {
						var vecBuffer [][4]uint16
						vecs, err := modeler.ReadAccessor(doc, doc.Accessors[idx], vecBuffer)
						if err != nil {
							return nil, err
						}
						for _, vec := range vecs.([][4]uint16) {
							vertexBuffer = append(vertexBuffer, f32.Bytes(binary.LittleEndian, float32(vec[0])/65535, float32(vec[1])/65535, float32(vec[2])/65535, float32(vec[3])/65535)...)
						}
					}
				case gltf.ComponentFloat:
					if doc.Accessors[idx].Type == gltf.AccessorVec3 {
						var vecBuffer [][3]float32
						vecs, err := modeler.ReadAccessor(doc, doc.Accessors[idx], vecBuffer)
						if err != nil {
							return nil, err
						}
						for _, vec := range vecs.([][3]float32) {
							vertexBuffer = append(vertexBuffer, f32.Bytes(binary.LittleEndian, vec[0], vec[1], vec[2], 1)...)
						}
					} else {
						var vecBuffer [][4]float32
						vecs, err := modeler.ReadAccessor(doc, doc.Accessors[idx], vecBuffer)
						if err != nil {
							return nil, err
						}
						for _, vec := range vecs.([][4]float32) {
							vertexBuffer = append(vertexBuffer, f32.Bytes(binary.LittleEndian, vec[:]...)...)
						}
					}
				}
			} else {
				primitive.useVertexColor = false
			}
			if idx, ok := p.Attributes[gltf.JOINTS_0]; ok {
				primitive.useJoint0 = true
				var jointBuffer [][4]uint16
				joints, err := modeler.ReadJoints(doc, doc.Accessors[idx], jointBuffer)
				if err != nil {
					return nil, err
				}
				for _, joint := range joints {
					var f [4]float32
					for j, v := range joint {
						f[j] = float32(v)
					}
					vertexBuffer = append(vertexBuffer, f32.Bytes(binary.LittleEndian, f[:]...)...)
				}
				if idx, ok := p.Attributes[gltf.WEIGHTS_0]; ok {
					var weightBuffer [][4]float32
					weights, err := modeler.ReadWeights(doc, doc.Accessors[idx], weightBuffer)
					if err != nil {
						return nil, err
					}
					for _, weight := range weights {
						vertexBuffer = append(vertexBuffer, f32.Bytes(binary.LittleEndian, weight[:]...)...)
					}
				} else {
					return nil, errors.New("Primitive attribute JOINTS_0 is specified but WEIGHTS_0 is not specified.")
				}
				if idx, ok := p.Attributes["JOINTS_1"]; ok {
					primitive.useJoint1 = true
					var jointBuffer [][4]uint16
					joints, err := modeler.ReadJoints(doc, doc.Accessors[idx], jointBuffer)
					if err != nil {
						return nil, err
					}
					for _, joint := range joints {
						var f [4]float32
						for j, v := range joint {
							f[j] = float32(v)
						}
						vertexBuffer = append(vertexBuffer, f32.Bytes(binary.LittleEndian, f[:]...)...)
					}
					primitive.useJoint1 = false
					if idx, ok := p.Attributes["WEIGHTS_1"]; primitive.useJoint1 && ok {
						var weightBuffer [][4]float32
						weights, err := modeler.ReadWeights(doc, doc.Accessors[idx], weightBuffer)
						if err != nil {
							return nil, err
						}
						for _, weight := range weights {
							vertexBuffer = append(vertexBuffer, f32.Bytes(binary.LittleEndian, weight[:]...)...)
						}
					} else if primitive.useJoint1 {
						return nil, errors.New("Primitive attribute JOINTS_1 is specified but WEIGHTS_1 is not specified.")
					}
				}
			} else {
				primitive.useJoint0 = false
			}
			for _, t := range p.Targets {
				if len(mesh.morphTargetWeights) == 0 {
					mesh.morphTargetWeights = make([]float32, len(p.Targets))
				}
				target := &MorphTarget{}
				for attr, accessor := range t {
					switch attr {
					case "POSITION":
						o := uint32(len(vertexBuffer))
						target.positionOffset = &o
						var posBuffer [][3]float32
						positions, err := modeler.ReadPosition(doc, doc.Accessors[accessor], posBuffer)
						if err != nil {
							return nil, err
						}
						for _, pos := range positions {
							vertexBuffer = append(vertexBuffer, f32.Bytes(binary.LittleEndian, pos[0], pos[1], pos[2], 0)...)
						}
					case "TEXCOORD_0":
						o := uint32(len(vertexBuffer))
						target.uvOffset = &o
						var uvBuffer [][2]float32
						texCoords, err := modeler.ReadTextureCoord(doc, doc.Accessors[accessor], uvBuffer)
						if err != nil {
							return nil, err
						}
						for _, uv := range texCoords {
							vertexBuffer = append(vertexBuffer, f32.Bytes(binary.LittleEndian, uv[0], uv[1], 0, 0)...)
						}
					case "COLOR_0":
						o := uint32(len(vertexBuffer))
						target.colorOffset = &o
						switch doc.Accessors[accessor].ComponentType {
						case gltf.ComponentUbyte:
							if doc.Accessors[accessor].Type == gltf.AccessorVec3 {
								var vecBuffer [][3]uint8
								vecs, err := modeler.ReadAccessor(doc, doc.Accessors[accessor], vecBuffer)
								if err != nil {
									return nil, err
								}
								for _, vec := range vecs.([][3]uint8) {
									vertexBuffer = append(vertexBuffer, f32.Bytes(binary.LittleEndian, float32(vec[0])/255, float32(vec[1])/255, float32(vec[2])/255, 1)...)
								}
							} else {
								var vecBuffer [][4]uint8
								vecs, err := modeler.ReadAccessor(doc, doc.Accessors[accessor], vecBuffer)
								if err != nil {
									return nil, err
								}
								for _, vec := range vecs.([][4]uint8) {
									vertexBuffer = append(vertexBuffer, f32.Bytes(binary.LittleEndian, float32(vec[0])/255, float32(vec[1])/255, float32(vec[2])/255, float32(vec[3])/255)...)
								}
							}
						case gltf.ComponentUshort:
							if doc.Accessors[accessor].Type == gltf.AccessorVec3 {
								var vecBuffer [][3]uint16
								vecs, err := modeler.ReadAccessor(doc, doc.Accessors[accessor], vecBuffer)
								if err != nil {
									return nil, err
								}
								for _, vec := range vecs.([][3]uint16) {
									vertexBuffer = append(vertexBuffer, f32.Bytes(binary.LittleEndian, float32(vec[0])/65535, float32(vec[1])/65535, float32(vec[2])/65535, 1)...)
								}
							} else {
								var vecBuffer [][4]uint16
								vecs, err := modeler.ReadAccessor(doc, doc.Accessors[accessor], vecBuffer)
								if err != nil {
									return nil, err
								}
								for _, vec := range vecs.([][4]uint16) {
									vertexBuffer = append(vertexBuffer, f32.Bytes(binary.LittleEndian, float32(vec[0])/65535, float32(vec[1])/65535, float32(vec[2])/65535, float32(vec[3])/65535)...)
								}
							}
						case gltf.ComponentFloat:
							if doc.Accessors[accessor].Type == gltf.AccessorVec3 {
								var vecBuffer [][3]float32
								vecs, err := modeler.ReadAccessor(doc, doc.Accessors[accessor], vecBuffer)
								if err != nil {
									return nil, err
								}
								for _, vec := range vecs.([][3]float32) {
									vertexBuffer = append(vertexBuffer, f32.Bytes(binary.LittleEndian, vec[0], vec[1], vec[2], 1)...)
								}
							} else {
								var vecBuffer [][4]float32
								vecs, err := modeler.ReadAccessor(doc, doc.Accessors[accessor], vecBuffer)
								if err != nil {
									return nil, err
								}
								for _, vec := range vecs.([][4]float32) {
									vertexBuffer = append(vertexBuffer, f32.Bytes(binary.LittleEndian, vec[:]...)...)
								}
							}
						}
					}
				}
				primitive.morphTargets = append(primitive.morphTargets, target)
			}
			if p.Material != nil {
				primitive.materialIndex = new(uint32)
				*primitive.materialIndex = *p.Material
			}
			primitive.mode = gltfPrimitiveModeMap[p.Mode]
			mesh.primitives = append(mesh.primitives, primitive)
		}
		mdl.meshes = append(mdl.meshes, mesh)
	}
	mdl.vertexBuffer = vertexBuffer
	mdl.elementBuffer = elementBuffer

	mdl.nodes = make([]*Node, 0, len(doc.Nodes))
	for _, n := range doc.Nodes {
		var node = &Node{}
		mdl.nodes = append(mdl.nodes, node)
		node.rotation = n.Rotation
		node.transition = n.Translation
		node.scale = n.Scale
		node.skin = n.Skin
		node.childrenIndex = n.Children
		node.morphTargetWeights = n.Weights
		if n.Mesh != nil {
			node.meshIndex = new(uint32)
			*node.meshIndex = *n.Mesh
			if len(node.morphTargetWeights) == 0 {
				node.morphTargetWeights = mdl.meshes[*node.meshIndex].morphTargetWeights
			}
		}
		node.trans = TransNone
		node.zTest = true
		node.zWrite = true
		if n.Extras != nil {
			v, ok := n.Extras.(map[string]interface{})
			if ok {
				switch v["trans"] {
				case "ADD":
					node.trans = TransAdd
				case "SUB":
					node.trans = TransReverseSubtract
				case "NONE":
					node.trans = TransNone
				}
				if v["disableZTest"] != nil && v["disableZTest"] != "0" && v["disableZTest"] != "false" {
					node.zTest = false
				}
				if v["disableZWrite"] != nil && v["disableZWrite"] != "0" && v["disableZWrite"] != "false" {
					node.zWrite = false
				}
			}

		}
		node.transformChanged = true
	}
	for i, n := range mdl.nodes {
		for _, c := range n.childrenIndex {
			p := int32(i)
			mdl.nodes[c].parentIndex = &p
		}
	}
	mdl.animationTimeStamps = map[uint32][]float32{}
	for _, a := range doc.Animations {
		anim := &GLTFAnimation{}
		mdl.animations = append(mdl.animations, anim)
		anim.duration = 0
		for _, c := range a.Channels {
			channel := &GLTFAnimationChannel{}
			channel.nodeIndex = *c.Target.Node
			channel.samplerIndex = *c.Sampler
			switch c.Target.Path {
			case gltf.TRSTranslation:
				channel.path = TRSTranslation
			case gltf.TRSScale:
				channel.path = TRSScale
			case gltf.TRSRotation:
				channel.path = TRSRotation
			case gltf.TRSWeights:
				channel.path = MorphTargetWeight
			default:
				continue
			}
			anim.channels = append(anim.channels, channel)
		}
		for _, s := range a.Samplers {
			sampler := &GLTFAnimationSampler{}
			anim.samplers = append(anim.samplers, sampler)
			if _, ok := mdl.animationTimeStamps[s.Input]; !ok {
				var timeBuffer []float32
				times, err := modeler.ReadAccessor(doc, doc.Accessors[s.Input], timeBuffer)
				if err != nil {
					return nil, err
				}
				mdl.animationTimeStamps[s.Input] = make([]float32, 0, len(times.([]float32)))
				for _, t := range times.([]float32) {
					mdl.animationTimeStamps[s.Input] = append(mdl.animationTimeStamps[s.Input], t)
				}
			}
			sampler.interpolation = GLTFAnimationInterpolation(s.Interpolation)
			sampler.inputIndex = s.Input
			if anim.duration < mdl.animationTimeStamps[s.Input][len(mdl.animationTimeStamps[s.Input])-1] {
				anim.duration = mdl.animationTimeStamps[s.Input][len(mdl.animationTimeStamps[s.Input])-1]
			}
			switch doc.Accessors[s.Output].Type {
			case gltf.AccessorScalar:
				var vecBuffer []float32
				vecs, err := modeler.ReadAccessor(doc, doc.Accessors[s.Output], vecBuffer)
				if err != nil {
					return nil, err
				}
				sampler.output = make([]float32, 0, len(vecs.([]float32)))
				for _, val := range vecs.([]float32) {
					sampler.output = append(sampler.output, val)
				}
			case gltf.AccessorVec3:
				var vecBuffer [][3]float32
				vecs, err := modeler.ReadAccessor(doc, doc.Accessors[s.Output], vecBuffer)
				if err != nil {
					return nil, err
				}
				sampler.output = make([]float32, 0, len(vecs.([][3]float32))*3)
				for _, vec := range vecs.([][3]float32) {
					sampler.output = append(sampler.output, vec[0], vec[1], vec[2])
				}
			case gltf.AccessorVec4:
				var vecBuffer [][4]float32
				vecs, err := modeler.ReadAccessor(doc, doc.Accessors[s.Output], vecBuffer)
				if err != nil {
					return nil, err
				}
				sampler.output = make([]float32, 0, len(vecs.([][4]float32))*4)
				for _, vec := range vecs.([][4]float32) {
					sampler.output = append(sampler.output, vec[0], vec[1], vec[2], vec[3])
				}
			}
		}
	}
	for _, s := range doc.Skins {
		var skin = &Skin{}
		for _, j := range s.Joints {
			skin.joints = append(skin.joints, j)
		}

		if s.InverseBindMatrices != nil {
			var matrixBuffer [][4][4]float32
			matrices, err := modeler.ReadAccessor(doc, doc.Accessors[*s.InverseBindMatrices], matrixBuffer)
			if err != nil {
				return nil, err
			}
			for _, mat := range matrices.([][4][4]float32) {
				skin.inverseBindMatrices = append(skin.inverseBindMatrices, mat[0][:]...)
				skin.inverseBindMatrices = append(skin.inverseBindMatrices, mat[1][:]...)
				skin.inverseBindMatrices = append(skin.inverseBindMatrices, mat[2][:]...)
			}
		}

		skin.texture = &GLTFTexture{}
		sys.mainThreadTask <- func() {
			skin.texture.tex = newDataTexture(3, int32(len(skin.joints)))
		}

		mdl.skins = append(mdl.skins, skin)
	}

	for _, s := range doc.Scenes {
		var scene = &Scene{}
		scene.name = s.Name
		scene.nodes = s.Nodes
		mdl.scenes = append(mdl.scenes, scene)
	}
	return mdl, nil
}
func (n *Node) getLocalTransform() (mat mgl.Mat4) {
	mat = mgl.Ident4()
	if n.transformChanged {
		mat = mgl.Translate3D(n.transition[0], n.transition[1], n.transition[2])
		mat = mat.Mul4(mgl.Quat{W: n.rotation[3], V: mgl.Vec3{n.rotation[0], n.rotation[1], n.rotation[2]}}.Mat4())
		mat = mat.Mul4(mgl.Scale3D(n.scale[0], n.scale[1], n.scale[2]))
		n.localTransform = mat
		n.transformChanged = false
	} else {
		mat = n.localTransform
	}
	return
}
func (n *Node) calculateWorldTransform(parentTransorm mgl.Mat4, nodes []*Node) {
	mat := n.getLocalTransform()
	n.worldTransform = parentTransorm.Mul4(mat)
	for _, index := range n.childrenIndex {
		(*nodes[index]).calculateWorldTransform(n.worldTransform, nodes)
	}
	return
}
func drawNode(mdl *Model, n *Node, proj, view mgl.Mat4, drawBlended bool) {
	//mat := n.getLocalTransform()
	//model = model.Mul4(mat)
	for _, index := range n.childrenIndex {
		drawNode(mdl, mdl.nodes[index], proj, view, drawBlended)
	}
	if n.meshIndex == nil {
		return
	}
	if n.skin != nil {
		mdl.skins[*n.skin].calculateSkinMatrices(n.worldTransform.Inv(), mdl.nodes)
	}
	var morphTargetWeights []struct {
		index  uint32
		weight float32
	}
	if len(n.morphTargetWeights) > 0 {
		for idx, w := range n.morphTargetWeights {
			if w != 0 {
				morphTargetWeights = append(morphTargetWeights, struct {
					index  uint32
					weight float32
				}{uint32(idx), w})
			}
		}
	}

	neg, grayscale, padd, pmul, invblend, hue := mdl.pfx.getFcPalFx(false, -int(n.trans))

	blendEq := BlendAdd
	src := BlendOne
	dst := BlendOneMinusSrcAlpha
	switch n.trans {
	case TransAdd:
		if invblend == 3 {
			src = BlendOne
			dst = BlendOne
			blendEq = BlendReverseSubtract
			neg = false
		} else {
			src = BlendOne
			dst = BlendOne
		}
	case TransReverseSubtract:
		if invblend == 3 {
			src = BlendOne
			dst = BlendOne
			neg = false
		} else {
			src = BlendOne
			dst = BlendOne
			blendEq = BlendReverseSubtract
		}
	default:
		src = BlendOne
		dst = BlendOneMinusSrcAlpha
	}
	m := mdl.meshes[*n.meshIndex]
	for _, p := range m.primitives {
		if p.materialIndex == nil {
			continue
		}
		mat := mdl.materials[*p.materialIndex]
		if ((mat.alphaMode != AlphaModeBlend && n.trans == TransNone) && drawBlended) ||
			((mat.alphaMode == AlphaModeBlend || n.trans != TransNone) && !drawBlended) {
			return
		}
		color := mdl.materials[*p.materialIndex].baseColorFactor
		modelview := view.Mul4(n.worldTransform)
		gfx.SetModelPipeline(blendEq, src, dst, n.zTest, n.zWrite, mdl.materials[*p.materialIndex].doubleSided, modelview.Det() < 0, p.useUV, p.useVertexColor, p.useJoint0, p.useJoint1, p.numVertices, p.vertexBufferOffset)

		gfx.SetModelUniformMatrix("projection", proj[:])
		gfx.SetModelUniformMatrix("modelview", modelview[:])
		if index := mat.textureIndex; index != nil {
			gfx.SetModelTexture("tex", mdl.textures[*index].tex)
			gfx.SetModelUniformI("textured", 1)
		} else {
			gfx.SetModelUniformI("textured", 0)
		}

		if n.skin != nil {
			skin := mdl.skins[*n.skin]
			gfx.SetModelTexture("jointMatrices", skin.texture.tex)
		}
		if len(morphTargetWeights) > 0 && len(p.morphTargets) >= len(morphTargetWeights) {
			var targetOffsets [8]uint32
			var targetWeights [8]float32
			targetCount := 0
			for _, t := range morphTargetWeights {
				morphTarget := p.morphTargets[t.index]
				if morphTarget.positionOffset != nil {
					targetOffsets[targetCount] = *morphTarget.positionOffset
					targetWeights[targetCount] = t.weight
					targetCount += 1
				}
			}
			positionTargetCount := targetCount
			for _, t := range morphTargetWeights {
				morphTarget := p.morphTargets[t.index]
				if morphTarget.uvOffset != nil {
					targetOffsets[targetCount] = *morphTarget.uvOffset
					targetWeights[targetCount] = t.weight
					targetCount += 1
				}
			}
			uvTargetCount := targetCount - positionTargetCount
			for _, t := range morphTargetWeights {
				morphTarget := p.morphTargets[t.index]
				if morphTarget.colorOffset != nil {
					targetOffsets[targetCount] = *morphTarget.colorOffset
					targetWeights[targetCount] = t.weight
					targetCount += 1
				}
			}
			gfx.SetModelMorphTarget(targetOffsets, targetWeights, positionTargetCount, uvTargetCount)
		} else {
			gfx.SetModelUniformFv("morphTargetWeight", make([]float32, 8))
		}

		mode := p.mode
		if sys.wireframeDraw {
			mode = 1 // Set mesh render mode to "lines"
		}
		gfx.SetModelUniformFv("add", padd[:])
		gfx.SetModelUniformFv("mult", []float32{pmul[0] * float32(sys.brightness) / 256, pmul[1] * float32(sys.brightness) / 256, pmul[2] * float32(sys.brightness) / 256})
		gfx.SetModelUniformI("neg", int(Btoi(neg)))
		gfx.SetModelUniformF("hue", hue)
		gfx.SetModelUniformF("gray", grayscale)
		gfx.SetModelUniformI("enableAlpha", int(Btoi(mat.alphaMode == AlphaModeBlend)))
		gfx.SetModelUniformF("alphaThreshold", mat.alphaCutoff)
		gfx.SetModelUniformFv("baseColorFactor", color[:])
		if n.skin != nil {
			gfx.SetModelUniformI("numJoints", len(mdl.skins[*n.skin].joints))
		}

		gfx.RenderElements(mode, int(p.numIndices), int(p.elementBufferOffset))

		gfx.ReleaseModelPipeline()

	}
}
func (s *Stage) drawModel(pos [2]float32, yofs float32, scl float32, sceneNumber int) {
	if s.model == nil || len(s.model.scenes) <= sceneNumber {
		return
	}

	drawFOV := s.stageCamera.fov * math.Pi / 180

	var posMul float32 = float32(math.Tan(float64(drawFOV)/2)) * -s.model.offset[2] / (float32(sys.scrrect[3]) / 2)

	var syo float32
	aspectCorrection := (float32(sys.cam.zoffset)*float32(sys.gameHeight)/float32(sys.cam.localcoord[1]) - (float32(sys.cam.zoffset)*s.localscl - sys.cam.aspectcorrection))
	syo = -(float32(s.stageCamera.zoffset) - float32(sys.cam.localcoord[1])/2) * (1 - scl) / scl * float32(sys.gameHeight) / float32(s.stageCamera.localcoord[1])
	offset := []float32{(pos[0]*-posMul*s.localscl*sys.widthScale + s.model.offset[0]/scl), (((pos[1]*s.localscl+sys.cam.zoomanchorcorrection+aspectCorrection)/scl+yofs/scl+syo)*posMul*sys.heightScale + s.model.offset[1]), s.model.offset[2] / scl}
	rotation := []float32{s.model.rotation[0], s.model.rotation[1], s.model.rotation[2]}
	scale := []float32{s.model.scale[0], s.model.scale[1], s.model.scale[2]}
	proj := mgl.Translate3D(0, sys.cam.yshift*scl, 0)
	proj = proj.Mul4(mgl.Perspective(drawFOV, float32(sys.scrrect[2])/float32(sys.scrrect[3]), s.stageCamera.near, s.stageCamera.far))
	view := mgl.Ident4()
	view = view.Mul4(mgl.Translate3D(offset[0], offset[1], offset[2]))
	view = view.Mul4(mgl.HomogRotate3DX(rotation[0]))
	view = view.Mul4(mgl.HomogRotate3DY(rotation[1]))
	view = view.Mul4(mgl.HomogRotate3DZ(rotation[2]))
	view = view.Mul4(mgl.Scale3D(scale[0], scale[1], scale[2]))
	scene := s.model.scenes[sceneNumber]

	for _, index := range scene.nodes {
		s.model.nodes[index].calculateWorldTransform(mgl.Ident4(), s.model.nodes)
	}
	for _, index := range scene.nodes {
		drawNode(s.model, s.model.nodes[index], proj, view, false)
	}
	for _, index := range scene.nodes {
		drawNode(s.model, s.model.nodes[index], proj, view, true)
	}
}

func (model *Model) step() {
	for _, anim := range model.animations {
		anim.time += sys.turbo / 60
		for anim.time >= anim.duration && anim.duration > 0 {
			anim.time -= anim.duration
		}
		time := 60 * float64(anim.time)
		if math.Abs(time-math.Floor(time)) < 0.001 {
			anim.time = float32(math.Floor(time) / 60)
		} else if math.Abs(float64(anim.time)-math.Ceil(float64(anim.time))) < 0.001 {
			anim.time = float32(math.Ceil(time) / 60)
		}
		if anim.time >= anim.duration && anim.duration > 0 {
			anim.time = anim.duration
		}
		for _, channel := range anim.channels {
			node := model.nodes[channel.nodeIndex]
			sampler := anim.samplers[channel.samplerIndex]
			prevIndex := 0
			for i, t := range model.animationTimeStamps[sampler.inputIndex] {
				if anim.time < t {
					prevIndex = i - 1
					break
				}
			}
			if prevIndex != -1 && sampler.interpolation != InterpolationStep && prevIndex+1 < len(model.animationTimeStamps[sampler.inputIndex]) {
				if sampler.interpolation == InterpolationLinear {
					rate := (anim.time - model.animationTimeStamps[sampler.inputIndex][prevIndex]) / (model.animationTimeStamps[sampler.inputIndex][prevIndex+1] - model.animationTimeStamps[sampler.inputIndex][prevIndex])
					switch channel.path {
					case TRSTranslation:
						for i := 0; i < 3; i++ {
							newVal := sampler.output[prevIndex*3+i]*(1-rate) + sampler.output[(prevIndex+1)*3+i]*rate
							if node.transition[i] != newVal {
								node.transition[i] = newVal
								node.transformChanged = true
							}
						}
					case TRSScale:
						for i := 0; i < 3; i++ {
							newVal := sampler.output[prevIndex*3+i]*(1-rate) + sampler.output[(prevIndex+1)*3+i]*rate
							if node.scale[i] != newVal {
								node.scale[i] = newVal
								node.transformChanged = true
							}
						}
					case TRSRotation:
						q1 := mgl.Quat{sampler.output[prevIndex*4+3], mgl.Vec3{sampler.output[prevIndex*4], sampler.output[prevIndex*4+1], sampler.output[prevIndex*4+2]}}
						q2 := mgl.Quat{sampler.output[(prevIndex+1)*4+3], mgl.Vec3{sampler.output[(prevIndex+1)*4], sampler.output[(prevIndex+1)*4+1], sampler.output[(prevIndex+1)*4+2]}}
						dotProduct := q1.Dot(q2)
						if dotProduct < 0 {
							q1 = q1.Inverse()
						}
						q := mgl.QuatSlerp(q1, q2, rate)
						if node.rotation[0] != q.X() || node.rotation[1] != q.Y() || node.rotation[2] != q.Z() || node.rotation[3] != q.W {
							node.rotation[0] = q.X()
							node.rotation[1] = q.Y()
							node.rotation[2] = q.Z()
							node.rotation[3] = q.W
							node.transformChanged = true
						}
					case MorphTargetWeight:
						for i := 0; i < len(node.morphTargetWeights); i++ {
							newVal := sampler.output[prevIndex*len(node.morphTargetWeights)+i]*(1-rate) + sampler.output[(prevIndex+1)*len(node.morphTargetWeights)+i]*rate
							node.morphTargetWeights[i] = newVal
						}
					}
				} else {
					delta := (model.animationTimeStamps[sampler.inputIndex][prevIndex+1] - model.animationTimeStamps[sampler.inputIndex][prevIndex])
					rate := (anim.time - model.animationTimeStamps[sampler.inputIndex][prevIndex]) / delta
					rateSquare := rate * rate
					rateCube := rateSquare * rate

					switch channel.path {
					case TRSTranslation:
						for i := 0; i < 3; i++ {
							newVal := (2*rateCube-3*rateSquare+1)*sampler.output[prevIndex*9+3*i+1] + delta*(rateCube-2*rateSquare+rate)*sampler.output[prevIndex*9+3*i+2] + (-2*rateCube+3*rateSquare)*sampler.output[(prevIndex+1)*9+3*i+1] + delta*(rateCube-rateSquare)*sampler.output[(prevIndex+1)*9+3*i]
							if node.transition[i] != newVal {
								node.transition[i] = newVal
								node.transformChanged = true
							}
						}
					case TRSScale:
						for i := 0; i < 3; i++ {
							newVal := (2*rateCube-3*rateSquare+1)*sampler.output[prevIndex*9+3*i+1] + delta*(rateCube-2*rateSquare+rate)*sampler.output[prevIndex*9+3*i+2] + (-2*rateCube+3*rateSquare)*sampler.output[(prevIndex+1)*9+3*i+1] + delta*(rateCube-rateSquare)*sampler.output[(prevIndex+1)*9+3*i]
							if node.scale[i] != newVal {
								node.scale[i] = newVal
								node.transformChanged = true
							}
						}
					case TRSRotation:
						q1 := mgl.Quat{sampler.output[prevIndex*4+3], mgl.Vec3{sampler.output[prevIndex*4], sampler.output[prevIndex*4+1], sampler.output[prevIndex*4+2]}}
						q2 := mgl.Quat{sampler.output[(prevIndex+1)*4+3], mgl.Vec3{sampler.output[(prevIndex+1)*4], sampler.output[(prevIndex+1)*4+1], sampler.output[(prevIndex+1)*4+2]}}
						dotProduct := q1.Dot(q2)
						if dotProduct < 0 {
							q1 = q1.Inverse()
						}
						q := mgl.Quat{(2*rateCube-3*rateSquare+1)*sampler.output[prevIndex*12+9+1] + delta*(rateCube-2*rateSquare+rate)*sampler.output[prevIndex*12+9+2] + (-2*rateCube+3*rateSquare)*sampler.output[(prevIndex+1)*12+9+1] + delta*(rateCube-rateSquare)*sampler.output[(prevIndex+1)*12+9],
							mgl.Vec3{
								(2*rateCube-3*rateSquare+1)*sampler.output[prevIndex*12+1] + delta*(rateCube-2*rateSquare+rate)*sampler.output[prevIndex*12+2] + (-2*rateCube+3*rateSquare)*sampler.output[(prevIndex+1)*12+1] + delta*(rateCube-rateSquare)*sampler.output[(prevIndex+1)*12],
								(2*rateCube-3*rateSquare+1)*sampler.output[prevIndex*12+3+1] + delta*(rateCube-2*rateSquare+rate)*sampler.output[prevIndex*12+3+2] + (-2*rateCube+3*rateSquare)*sampler.output[(prevIndex+1)*12+3+1] + delta*(rateCube-rateSquare)*sampler.output[(prevIndex+1)*12+3],
								(2*rateCube-3*rateSquare+1)*sampler.output[prevIndex*12+6+1] + delta*(rateCube-2*rateSquare+rate)*sampler.output[prevIndex*12+6+2] + (-2*rateCube+3*rateSquare)*sampler.output[(prevIndex+1)*12+6+1] + delta*(rateCube-rateSquare)*sampler.output[(prevIndex+1)*12+6],
							}}.Normalize()
						if node.rotation[0] != q.X() || node.rotation[1] != q.Y() || node.rotation[2] != q.Z() || node.rotation[3] != q.W {
							node.rotation[0] = q.X()
							node.rotation[1] = q.Y()
							node.rotation[2] = q.Z()
							node.rotation[3] = q.W
							node.transformChanged = true
						}
					case MorphTargetWeight:
						for i := 0; i < len(node.morphTargetWeights); i++ {
							newVal := (2*rateCube-3*rateSquare+1)*sampler.output[prevIndex*3*len(node.morphTargetWeights)+3*i+1] + delta*(rateCube-2*rateSquare+rate)*sampler.output[prevIndex*3*len(node.morphTargetWeights)+3*i+2] + (-2*rateCube+3*rateSquare)*sampler.output[(prevIndex+1)*3*len(node.morphTargetWeights)+3*i+1] + delta*(rateCube-rateSquare)*sampler.output[(prevIndex+1)*3*len(node.morphTargetWeights)+3*i]
							node.morphTargetWeights[i] = newVal
						}
					}
				}

			} else {
				if prevIndex == -1 {
					prevIndex = 0
				}
				switch channel.path {
				case TRSTranslation:
					for i := 0; i < 3; i++ {
						if node.transition[i] != sampler.output[prevIndex*3+i] {
							node.transition[i] = sampler.output[prevIndex*3+i]
							node.transformChanged = true
						}
					}
				case TRSScale:
					for i := 0; i < 3; i++ {
						if node.scale[i] != sampler.output[prevIndex*3+i] {
							node.scale[i] = sampler.output[prevIndex*3+i]
							node.transformChanged = true
						}
					}
				case TRSRotation:
					for i := 0; i < 4; i++ {
						if node.rotation[i] != sampler.output[prevIndex*4+i] {
							node.rotation[i] = sampler.output[prevIndex*4+i]
							node.transformChanged = true
						}
					}
				case MorphTargetWeight:
					for i := 0; i < len(node.morphTargetWeights); i++ {
						newVal := sampler.output[prevIndex*len(node.morphTargetWeights)+i]
						node.morphTargetWeights[i] = newVal
					}
				}
			}
		}
	}
}
func (skin *Skin) calculateSkinMatrices(inverseGlobalTransform mgl.Mat4, nodes []*Node) {
	matrices := make([]float32, len(skin.joints)*12)
	for i, joint := range skin.joints {
		n := nodes[joint]
		reverseBindMatrix := skin.inverseBindMatrices[i*12 : (i+1)*12]
		matrix := mgl.Ident4()
		for j, v := range reverseBindMatrix {
			matrix[j] = v
		}
		matrix = n.worldTransform.Mul4(matrix.Transpose())
		matrix = inverseGlobalTransform.Mul4(matrix).Transpose()
		for j := 0; j < 12; j++ {
			matrices[i*12+j] = matrix[j]
		}
	}
	skin.texture.tex.SetPixelData(matrices)
}
func (model *Model) reset() {
	for _, anim := range model.animations {
		anim.time = 0
	}
}
