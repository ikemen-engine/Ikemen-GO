package main

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/draw"
	_ "image/jpeg"
	"io"
	"io/fs"
	"math"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"

	_ "github.com/lukegb/dds"
	"github.com/mdouchement/hdr"
	_ "github.com/mdouchement/hdr/codec/rgbe"

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

type BgType int32

const (
	BG_Normal BgType = iota
	BG_Anim
	BG_Parallax
	BG_Video
	BG_Dummy
)

type BgVideoScaleMode int32

const (
	SM_None BgVideoScaleMode = iota
	SM_Stretch
	SM_Fit
	SM_FitWidth
	SM_FitHeight
	SM_ZoomFill
)

type BgVideoScaleFilter int32

const (
	SF_FastBilinear BgVideoScaleFilter = iota
	SF_Bilinear
	SF_Bicubic
	SF_Experimental
	SF_Neighbor
	SF_Area
	SF_Bicublin
	SF_Gauss
	SF_Sinc
	SF_Lanczos
	SF_Spline
)

type backGround struct {
	_type              BgType
	palfx              *PalFX
	anim               *Animation
	bga                bgAction
	video              bgVideo
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
	enabled            bool
	positionlink       bool
	layerno            int32
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
	rot                Rotation
	fLength            float32
	projection         Projection
	xshear             float32
}

func newBackGround(sff *Sff) *backGround {
	return &backGround{
		palfx:              newPalFX(),
		anim:               newAnimation(sff, &sff.palList),
		delta:              [...]float32{1, 1},
		zoomdelta:          [...]float32{1, math.MaxFloat32},
		xscale:             [...]float32{1, 1},
		rasterx:            [...]float32{1, 1},
		yscalestart:        100,
		scalestart:         [...]float32{1, 1},
		rot:                Rotation{0, 0, 0},
		xshear:             0,
		fLength:            2048,
		projection:         Projection_Orthographic,
		xbottomzoomdelta:   math.MaxFloat32,
		zoomscaledelta:     [...]float32{math.MaxFloat32, math.MaxFloat32},
		actionno:           -1,
		visible:            true,
		enabled:            true,
		autoresizeparallax: false,
		startrect:          [...]int32{-32768, -32768, 65535, 65535},
	}
}

func readBackGround(is IniSection, link *backGround,
	sff *Sff, at AnimationTable, sProps StageProps, def string) (*backGround, error) {
	bg := newBackGround(sff)
	typ := is["type"]
	if len(typ) == 0 {
		return bg, nil
	}
	switch typ[0] {
	case 'N', 'n':
		bg._type = BG_Normal
	case 'A', 'a':
		bg._type = BG_Anim
	case 'P', 'p':
		bg._type = BG_Parallax
	case 'V', 'v':
		bg._type = BG_Video
	case 'D', 'd':
		bg._type = BG_Dummy
	default:
		return bg, nil
	}
	var tmp int32
	is.ReadI32("layerno", &bg.layerno)
	if bg._type == BG_Video {
		path := is["path"]
		LoadFile(&path, []string{def, "", sys.motifDir, "data/", "video/"}, func(filename string) error {
			path = filename
			return nil
		})
		if len(path) != 0 {
			volume := 100
			if v, ok := is["volume"]; ok {
				volume = int(Atoi(v))
			}

			var sm BgVideoScaleMode
			if v, ok := is["scalemode"]; ok {
				switch strings.ToLower(strings.TrimSpace(v)) {
				case "none":
					sm = SM_None
				case "stretch":
					sm = SM_Stretch
				case "fit":
					sm = SM_Fit
				case "fitwidth":
					sm = SM_FitWidth
				case "fitheight":
					sm = SM_FitHeight
				case "zoomfill":
					sm = SM_ZoomFill
				default:
					return nil, Error("Invalid BG Video scale mode: " + v)
				}
			}

			var sf BgVideoScaleFilter
			if v, ok := is["scalefilter"]; ok {
				switch strings.ToLower(strings.TrimSpace(v)) {
				case "fastbilinear":
					sf = SF_FastBilinear
				case "bilinear":
					sf = SF_Bilinear
				case "bicubic":
					sf = SF_Bicubic
				case "experimental":
					sf = SF_Experimental
				case "neighbor":
					sf = SF_Neighbor
				case "area":
					sf = SF_Area
				case "bicublin":
					sf = SF_Bicublin
				case "gauss":
					sf = SF_Gauss
				case "sinc":
					sf = SF_Sinc
				case "lanczos":
					sf = SF_Lanczos
				case "spline":
					sf = SF_Spline
				default:
					return nil, Error("Invalid BG Video scale filter: " + v)
				}
			}

			var loop bool
			is.ReadBool("loop", &loop)

			if err := bg.video.Open(path, volume, sm, sf, loop); err != nil {
				return nil, err
			}
		}
	} else if bg._type != BG_Dummy {
		var hasAnim bool
		if (bg._type != BG_Normal || len(is["spriteno"]) == 0) &&
			is.ReadI32("actionno", &bg.actionno) {
			if a := at.get(bg.actionno); a != nil {
				bg.anim = a
				hasAnim = true
			}
		}
		if hasAnim {
			if bg._type == BG_Normal {
				bg._type = BG_Anim
			}
		} else {
			var g, n int32
			if is.readI32ForStage("spriteno", &g, &n) {
				bg.anim.frames = []AnimFrame{*newAnimFrame()}
				bg.anim.frames[0].Group, bg.anim.frames[0].Number = g, n
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
	is.readF32ForStage("xshear", &bg.xshear)
	is.readF32ForStage("angle", &bg.rot.angle)
	is.readF32ForStage("xangle", &bg.rot.xangle)
	is.readF32ForStage("yangle", &bg.rot.yangle)
	is.readF32ForStage("focallength", &bg.fLength)
	if str, ok := is["projection"]; ok {
		switch strings.ToLower(strings.TrimSpace(str)) {
		case "orthographic":
			bg.projection = Projection_Orthographic
		case "perspective":
			bg.projection = Projection_Perspective
		case "perspective2":
			bg.projection = Projection_Perspective2
		}
	}
	is.readF32ForStage("xbottomzoomdelta", &bg.xbottomzoomdelta)
	is.readF32ForStage("zoomscaledelta", &bg.zoomscaledelta[0], &bg.zoomscaledelta[1])
	is.readF32ForStage("zoomdelta", &bg.zoomdelta[0], &bg.zoomdelta[1])
	if bg.zoomdelta[0] != math.MaxFloat32 && bg.zoomdelta[1] == math.MaxFloat32 {
		bg.zoomdelta[1] = bg.zoomdelta[0]
	}

	// Read transparency
	if data, ok := is["trans"]; ok {
		switch strings.ToLower(data) {
		case "add":
			bg.anim.mask = 0
			bg.anim.transType = TT_add
			bg.anim.srcAlpha = 255
			bg.anim.dstAlpha = 255
		case "add1":
			bg.anim.mask = 0
			bg.anim.transType = TT_add
			bg.anim.srcAlpha = 255
			bg.anim.dstAlpha = 128
		case "addalpha":
			bg.anim.mask = 0
			bg.anim.transType = TT_add
			bg.anim.srcAlpha = 255
			bg.anim.dstAlpha = 0 // In Mugen it defaults to this before reading the alpha
		case "sub":
			bg.anim.mask = 0
			bg.anim.transType = TT_sub
			bg.anim.srcAlpha = 255
			bg.anim.dstAlpha = 255
		case "none":
			// In Mugen this does the same as Default
			// TODO: Make ikemenversion fix it
			//bg.anim.transType = TT_none
			bg.anim.transType = TT_default
			bg.anim.srcAlpha = 255
			bg.anim.dstAlpha = 0
		case "default":
			bg.anim.transType = TT_default
			bg.anim.srcAlpha = 255
			bg.anim.dstAlpha = 0
		default:
			return nil, Error("Invalid trans type: " + data)
		}
	}

	// Read alpha if applicable
	if bg.anim.transType == TT_add || bg.anim.transType == TT_sub {
		if _, ok := is["alpha"]; ok {
			s, d := int32(bg.anim.srcAlpha), int32(bg.anim.dstAlpha)
			if is.readI32ForStage("alpha", &s, &d) {
				bg.anim.srcAlpha = int16(Clamp(s, 0, 255))
				bg.anim.dstAlpha = int16(Clamp(d, 0, 255))
			}
		}
	}

	if is.readI32ForStage("tile", &bg.anim.tile.xflag, &bg.anim.tile.yflag) {
		if bg._type == BG_Parallax {
			bg.anim.tile.yflag = 0
		}
		if bg.anim.tile.xflag < 0 {
			bg.anim.tile.xflag = math.MaxInt32
		}
	}
	if bg._type == BG_Parallax {
		if !is.readI32ForStage("width", &bg.width[0], &bg.width[1]) {
			is.readF32ForStage("xscale", &bg.rasterx[0], &bg.rasterx[1])
		}
		is.ReadF32("yscalestart", &bg.yscalestart)
		is.ReadF32("yscaledelta", &bg.yscaledelta)
	} else {
		is.ReadI32("tilespacing", &bg.anim.tile.xspacing, &bg.anim.tile.yspacing)
		//bg.anim.tile.yspacing = bg.anim.tile.xspacing
		if bg.actionno < 0 && len(bg.anim.frames) > 0 {
			group := bg.anim.frames[0].Group
			number := bg.anim.frames[0].Number
			if group >= 0 && number >= 0 {
				if spr := sff.GetSprite(uint16(group), uint16(number)); spr != nil {
					bg.anim.tile.xspacing += int32(spr.Size[0])
					bg.anim.tile.yspacing += int32(spr.Size[1])
				}
			}
		} else {
			if bg.anim.tile.xspacing == 0 {
				bg.anim.tile.xflag = 0
			}
			if bg.anim.tile.yspacing == 0 {
				bg.anim.tile.yflag = 0
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
	return bg, nil
}

func (bg *backGround) reset() {
	bg.bga.clear()
	bg.bga.vel = bg.startv
	bg.bga.radius = bg.startrad
	bg.bga.sintime = bg.startsint
	bg.bga.sinlooptime = bg.startsinlt

	if bg.anim != nil {
		bg.anim.Reset()
	}

	if bg.palfx != nil {
		bg.palfx.clear()
		bg.palfx.time = -1
		bg.palfx.invertblend = -3
	}
}

// Changes the BG's animation without changing its unique parameters
func (bg *backGround) changeAnim(animNo int32, table AnimationTable) {
	// Get the new animation
	// Putting this step here prevents multiple BG elements from sharing the same animation pointer
	// It's also closer to how characters, explods and such change animations
	// https://github.com/ikemen-engine/Ikemen-GO/issues/2830
	src := table.get(animNo)
	if src == nil {
		return
	}

	// Save the previous anim's parameters that should persist
	maskTemp := bg.anim.mask
	transTypeTemp := bg.anim.transType
	srcAlphaTemp := bg.anim.srcAlpha
	dstAlphaTemp := bg.anim.dstAlpha
	tileTemp := bg.anim.tile

	// Assign new animation and action number
	bg.anim = src
	bg.actionno = animNo

	// Restore persistent parameters
	bg.anim.mask = maskTemp
	bg.anim.transType = transTypeTemp
	bg.anim.srcAlpha = srcAlphaTemp
	bg.anim.dstAlpha = dstAlphaTemp
	bg.anim.tile = tileTemp
}

func (bg backGround) draw(pos [2]float32, drawscl, bgscl, stglscl float32,
	stgscl [2]float32, shakeY float32, isStage bool) {

	// Handle parallax scaling (type = 2)
	scalestartX := bg.scalestart[0]
	if bg._type == BG_Parallax && (bg.width[0] != 0 || bg.width[1] != 0) && bg.anim.spr != nil {
		bg.xscale[0] = float32(bg.width[0]) / float32(bg.anim.spr.Size[0])
		bg.xscale[1] = float32(bg.width[1]) / float32(bg.anim.spr.Size[0])
		scalestartX = AbsF(scalestartX)
		bg.xofs = scalestartX * ((-float32(bg.width[0]) / 2) + float32(bg.anim.spr.Offset[0])*bg.xscale[0])
	}

	// Calculate raster x ratio and base x scale
	xras := (bg.rasterx[1] - bg.rasterx[0]) / bg.rasterx[0]
	xbs, dx := bg.xscale[1], MaxF(0, bg.delta[0]*bgscl)

	// Initialize local scaling factors
	var sclx_recip, sclx, scly float32 = 1, 1, 1
	lscl := [...]float32{stglscl * stgscl[0], stglscl * stgscl[1]}

	// Handle zoom scaling if zoomdelta is specified
	if bg.zoomdelta[0] != math.MaxFloat32 {
		sclx = drawscl + (1-drawscl)*(1-bg.zoomdelta[0])
		scly = drawscl + (1-drawscl)*(1-bg.zoomdelta[1])
		if !bg.autoresizeparallax {
			sclx_recip = 1 + bg.zoomdelta[0]*((1/(sclx*lscl[0])*lscl[0])-1)
		}
	} else {
		sclx = MaxF(0, drawscl+(1-drawscl)*(1-dx))
		scly = MaxF(0, drawscl+(1-drawscl)*(1-MaxF(0, bg.delta[1]*bgscl)))
	}

	// Adjust x scale and x bottom zoom if autoresizeparallax is enabled
	if sclx != 0 && bg.autoresizeparallax {
		tmp := 1 / sclx
		if bg.xbottomzoomdelta != math.MaxFloat32 {
			xbs *= MaxF(0, drawscl+(1-drawscl)*(1-bg.xbottomzoomdelta*(xbs/bg.xscale[0]))) * tmp
		} else {
			xbs *= MaxF(0, drawscl+(1-drawscl)*(1-dx*(xbs/bg.xscale[0]))) * tmp
		}
		tmp *= MaxF(0, drawscl+(1-drawscl)*(1-dx*(xras+1)))
		xras -= tmp - 1
		xbs *= tmp
	}

	// Adjust scaling based on zoomscaledelta if available
	var xs3, ys3 float32 = 1, 1
	if bg.zoomscaledelta[0] != math.MaxFloat32 {
		xs3 = (drawscl + (1-drawscl)*(1-bg.zoomscaledelta[0])) / sclx
	}
	if bg.zoomscaledelta[1] != math.MaxFloat32 {
		ys3 = (drawscl + (1-drawscl)*(1-bg.zoomscaledelta[1])) / scly
	}

	// This handles the flooring of the camera position in MUGEN versions earlier than 1.0.
	var x, yScrollPos float32
	if bg.roundpos {
		x = bg.start[0] + bg.xofs - float32(Floor(pos[0]/stgscl[0]))*bg.delta[0] + bg.bga.offset[0]
		yScrollPos = float32(Floor(pos[1]/drawscl/stgscl[1])) * bg.delta[1]
		for i := 0; i < 2; i++ {
			pos[i] = float32(math.Floor(float64(pos[i])))
		}
	} else {
		x = bg.start[0] + bg.xofs - pos[0]/stgscl[0]*bg.delta[0] + bg.bga.offset[0]
		// Hires breaks ydelta scrolling vel, so bgscl was commented from here.
		yScrollPos = (pos[1] / drawscl / stgscl[1]) * bg.delta[1] // * bgscl
	}

	y := bg.start[1] - yScrollPos + bg.bga.offset[1]

	//In MUGEN, if Boundhigh is a positive value, the reference pos of yscaledelta will change
	var positiveBoundhigh float32
	if isStage && sys.cam.boundhigh > 0 {
		positiveBoundhigh = float32(sys.cam.boundhigh) * bg.delta[1] * bgscl / drawscl / stgscl[1]
	}
	// Calculate Y scaling based on vertical scroll position and delta
	ys2 := bg.scaledelta[1] * pos[1] * bg.delta[1] * bgscl / drawscl / stgscl[1]
	ys := ((100-(pos[1]-positiveBoundhigh)*bg.yscaledelta)*bgscl/bg.yscalestart)*bg.scalestart[1] + ys2
	xs := bg.scaledelta[0] * pos[0] * bg.delta[0] * bgscl / stgscl[0]
	x *= bgscl

	// Apply stage logic if BG is part of a stage
	if isStage {
		zoff := float32(sys.cam.zoffset) * stglscl
		y = y*bgscl + ((zoff-shakeY)/scly-zoff)/stglscl/stgscl[1]
		y -= sys.cam.aspectcorrection / (scly * stglscl * stgscl[1])
		y -= sys.cam.zoomanchorcorrection / (scly * stglscl * stgscl[1])
	} else {
		y = y*bgscl + ((float32(sys.gameHeight)-shakeY)/stglscl/scly-240)/stgscl[1]
	}

	// Final scaling factors
	sclx *= lscl[0]
	scly *= stglscl * stgscl[1]

	// Calculate window scale
	var wscl [2]float32
	for i := range wscl {
		if bg.zoomdelta[i] != math.MaxFloat32 {
			wscl[i] = MaxF(0, drawscl+(1-drawscl)*(1-MaxF(0, bg.zoomdelta[i]))) * bgscl * lscl[i]
		} else {
			wscl[i] = MaxF(0, drawscl+(1-drawscl)*(1-MaxF(0, bg.windowdelta[i]*bgscl))) * bgscl * lscl[i]
		}
	}

	// Calculate window top left corner position
	rect := bg.startrect

	startrect0 := float32(rect[0]) - (pos[0])/stgscl[0]*bg.windowdelta[0] +
		(float32(sys.gameWidth)/2/sclx - float32(bg.notmaskwindow)*(float32(sys.gameWidth)/2)*(1/lscl[0]))
	startrect0 *= sys.widthScale * wscl[0]
	if !isStage && wscl[0] == 1 {
		// Screenpacks X coordinates start from left edge of screen
		startrect0 += float32(sys.gameWidth-320) / 2 * sys.widthScale
	}

	startrect1 := float32(rect[1]) - pos[1]/drawscl/stgscl[1]*bg.windowdelta[1]
	if isStage {
		zoff := float32(sys.cam.zoffset) * stglscl
		startrect1 += (zoff-shakeY)/scly - zoff/stglscl/stgscl[1]
		startrect1 -= sys.cam.aspectcorrection / scly
		startrect1 -= sys.cam.zoomanchorcorrection / scly
	}
	startrect1 *= sys.heightScale * wscl[1]
	startrect1 -= shakeY

	// Determine final window
	rect[0] = int32(math.Floor(float64(startrect0)))
	rect[1] = int32(math.Floor(float64(startrect1)))
	rect[2] = int32(math.Floor(float64(startrect0 + (float32(rect[2]) * sys.widthScale * wscl[0]) - float32(rect[0]))))
	rect[3] = int32(math.Floor(float64(startrect1 + (float32(rect[3]) * sys.heightScale * wscl[1]) - float32(rect[1]))))

	// Render background if it's within the screen area
	if rect[0] < sys.scrrect[2] && rect[1] < sys.scrrect[3] && rect[0]+rect[2] > 0 && rect[1]+rect[3] > 0 {
		if bg._type == BG_Video {
			bg.video.Tick()
			if bg.video.texture == nil {
				return
			}

			bg.anim.isVideo = true
			bg.anim.spr = newSprite()
			bg.anim.spr.Tex = bg.video.texture

			w := float32(bg.video.texture.GetWidth())
			h := float32(bg.video.texture.GetHeight())
			//if bg.video.scaleMode != SM_None {
			//	w /= sys.widthScale
			//	h /= sys.heightScale
			//}
			bg.anim.spr.Size = [2]uint16{
				uint16(math.Ceil(float64(w))),
				uint16(math.Ceil(float64(h))),
			}

			bg.anim.scale_x = 1
			bg.anim.scale_y = 1

		}

		// Xshear offset correction
		xsoffset := -bg.xshear * SignF(bg.scalestart[1]) * (float32(bg.anim.spr.Offset[1]) * scly)

		if bg.rot.angle != 0 {
			xsoffset /= bg.rot.angle
		}

		// Choose render origin: top-left for screenpack/storyboard videos, center for everything else
		var rcx float32
		if bg._type != BG_Video || isStage {
			rcx = float32(sys.gameWidth) / 2
		}

		bg.anim.Draw(&rect, x-xsoffset, y, sclx, scly,
			bg.xscale[0]*bgscl*(scalestartX+xs)*xs3,
			xbs*bgscl*(scalestartX+xs)*xs3,
			ys*ys3, xras*x/(AbsF(ys*ys3)*lscl[1]*float32(bg.anim.spr.Size[1])*bg.scalestart[1])*sclx_recip*bg.scalestart[1]-bg.xshear,
			bg.rot, rcx, bg.palfx, 1, [2]float32{1, 1}, int32(bg.projection), bg.fLength, 0, false)
	}
}

type bgCtrl struct {
	bg           []*backGround
	node         []*Node
	anim         []*GLTFAnimation
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
	return &bgCtrl{
		looptime: -1,
		x:        float32(math.NaN()),
		y:        float32(math.NaN()),
	}
}

func (bgc *bgCtrl) read(is IniSection, idx int) error {
	bgc.idx = idx
	xy := false
	srcdst := false
	palfx := false
	data := strings.ToLower(is["type"])
	switch data {
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
	default:
		return Error("Invalid BGCtrl type: " + data)
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
	return nil
}

func (bgc *bgCtrl) xEnable() bool {
	return !math.IsNaN(float64(bgc.x))
}

func (bgc *bgCtrl) yEnable() bool {
	return !math.IsNaN(float64(bgc.y))
}

type stageShadow struct {
	intensity  int32
	color      uint32
	xscale     float32
	yscale     float32
	fadeend    int32
	fadebgn    int32
	xshear     float32
	rot        Rotation
	fLength    float32
	projection Projection
	offset     [2]float32
	window     [4]float32
	ydelta     float32
	layerno    int32
}

type stagePlayer struct {
	startx, starty, startz, facing int32
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
	animTable        AnimationTable
	bg               []*backGround
	bgc              []bgCtrl
	bga              bgAction // For position linking
	sdw              stageShadow
	reflection       stageShadow
	p                [MaxPlayerNo]stagePlayer
	leftbound        float32
	rightbound       float32
	screenleft       int32
	screenright      int32
	zoffsetlink      int32
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
	partnerspacing   int32
	ikemenver        [3]uint16
	ikemenverF       float32
	mugenver         [2]uint16
	mugenverF        float32
	reload           bool
	stageprops       StageProps
	model            *Model
	topbound         float32
	botbound         float32
}

func newStage(def string) *Stage {
	s := &Stage{
		def:            def,
		leftbound:      -1000,
		rightbound:     1000,
		screenleft:     15,
		screenright:    15,
		zoffsetlink:    -1,
		autoturn:       true,
		resetbg:        true,
		localscl:       1,
		scale:          [...]float32{float32(math.NaN()), float32(math.NaN())},
		bgmratiolife:   30,
		stageCamera:    *newStageCamera(),
		constants:      make(map[string]float32),
		partnerspacing: 25,
		bgmvolume:      100,
		bgmfreqmul:     1, // Fallback value to allow music to play on legacy stages without a bgmfreqmul parameter
	}
	s.sdw.intensity = 128
	s.sdw.color = 0x000000 // https://github.com/ikemen-engine/Ikemen-GO/issues/2150
	s.sdw.xscale = 1.0
	s.sdw.ydelta = 1.0
	s.sdw.yscale = 0.4
	s.reflection.color = 0xFFFFFF
	s.reflection.xscale = 1.0
	s.reflection.ydelta = 1.0
	s.reflection.yscale = 1.0 // Default scale is 1. It's normally off because default intensity is 0
	s.p[0].startx = -70
	s.p[1].startx = 70
	s.stageprops = newStageProps()
	return s
}

func loadStage(def string, maindef bool) (*Stage, error) {
	s := newStage(def)
	str, err := LoadText(def)
	if err != nil {
		return nil, err
	}
	s.sff = &Sff{}
	lines, i := SplitAndTrim(str, "\n"), 0
	s.animTable = ReadAnimationTable(s.sff, &s.sff.palList, lines, &i)
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

	var sec []IniSection
	sectionExists := false

	// Info group
	if sec = defmap[fmt.Sprintf("%v.info", sys.cfg.Config.Language)]; len(sec) > 0 {
		sectionExists = true
	} else {
		if sec = defmap["info"]; len(sec) > 0 {
			sectionExists = true
		}
	}
	if sectionExists {
		sectionExists = false
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
		// Clear then read MugenVersion
		s.mugenver = [2]uint16{}
		s.mugenverF = 0
		if str, ok := sec[0]["mugenversion"]; ok {
			s.mugenver, s.mugenverF = parseMugenVersion(str)
		}
		// Clear then read IkemenVersion
		s.ikemenver = [3]uint16{}
		if str, ok := sec[0]["ikemenversion"]; ok {
			s.ikemenver, s.ikemenverF = parseIkemenVersion(str)
		}
		// If the MUGEN version is lower than 1.0, default to camera pixel rounding (floor)
		if s.ikemenver[0] == 0 && s.ikemenver[1] == 0 && s.mugenver[0] != 1 {
			s.stageprops.roundpos = true
		}
		// AttachedChars
		ac := 0
		for i := range sec[0] {
			if !strings.HasPrefix(i, "attachedchar") {
				continue
			}
			if suffix := strings.TrimPrefix(i, "attachedchar"); suffix != "" {
				if _, err := strconv.Atoi(suffix); err != nil {
					continue
				}
			}
			if ac >= MaxAttachedChar {
				sys.appendToConsole(fmt.Sprintf("Warning: You can only define up to %d attachedchar(s). '%s' ignored.", MaxAttachedChar, i))
				continue
			}
			if err := sec[0].LoadFile(i, []string{def, "", sys.motifDir, "data/"}, func(filename string) error {
				// Ensure slice has correct length
				for len(s.attachedchardef) <= ac {
					s.attachedchardef = append(s.attachedchardef, "")
				}
				s.attachedchardef[ac] = filename
				return nil
			}); err == nil {
				ac++
			}
		}
		// RoundXdef
		if maindef {
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

	// StageInfo group. Needs to be read before most other groups so that localcoord is known
	if sec = defmap[fmt.Sprintf("%v.stageinfo", sys.cfg.Config.Language)]; len(sec) > 0 {
		sectionExists = true
	} else {
		if sec = defmap["stageinfo"]; len(sec) > 0 {
			sectionExists = true
		}
	}
	if sectionExists {
		sectionExists = false
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
	if s.stageCamera.localcoord[0] != 320 {
		// Update default values to new localcoord. Like characters do
		coordRatio := float32(s.stageCamera.localcoord[0]) / 320
		s.leftbound *= coordRatio
		s.rightbound *= coordRatio
		s.screenleft = int32(float32(s.screenleft) * coordRatio)
		s.screenright = int32(float32(s.screenright) * coordRatio)
		s.partnerspacing = int32(float32(s.partnerspacing) * coordRatio)
		s.p[0].startx = int32(float32(s.p[0].startx) * coordRatio)
		s.p[1].startx = int32(float32(s.p[1].startx) * coordRatio)
	}

	// Constants group
	if sec = defmap[fmt.Sprintf("%v.constants", sys.cfg.Config.Language)]; len(sec) > 0 {
		sectionExists = true
	} else {
		if sec = defmap["constants"]; len(sec) > 0 {
			sectionExists = true
		}
	}
	if sectionExists {
		sectionExists = false
		for key, value := range sec[0] {
			s.constants[key] = float32(Atof(value))
		}
	}

	// Scaling group
	if sec = defmap[fmt.Sprintf("%v.scaling", sys.cfg.Config.Language)]; len(sec) > 0 {
		sectionExists = true
	} else {
		if sec = defmap["scaling"]; len(sec) > 0 {
			sectionExists = true
		}
	}
	if sectionExists {
		sectionExists = false
		if s.mugenver[0] != 1 || s.ikemenver[0] >= 1 { // mugen 1.0+ removed support for z-axis, IKEMEN-Go 1.0 adds it back
			sec[0].ReadF32("topz", &s.stageCamera.topz)
			sec[0].ReadF32("botz", &s.stageCamera.botz)
			sec[0].ReadF32("topscale", &s.stageCamera.ztopscale)
			sec[0].ReadF32("botscale", &s.stageCamera.zbotscale)
			sec[0].ReadF32("depthtoscreen", &s.stageCamera.depthtoscreen)
		}
	}

	// Bound group
	if sec = defmap[fmt.Sprintf("%v.bound", sys.cfg.Config.Language)]; len(sec) > 0 {
		sectionExists = true
	} else {
		if sec = defmap["bound"]; len(sec) > 0 {
			sectionExists = true
		}
	}
	if sectionExists {
		sectionExists = false
		sec[0].ReadI32("screenleft", &s.screenleft)
		sec[0].ReadI32("screenright", &s.screenright)
	}

	// PlayerInfo Group
	if sec = defmap[fmt.Sprintf("%v.playerinfo", sys.cfg.Config.Language)]; len(sec) > 0 {
		sectionExists = true
	} else {
		if sec = defmap["playerinfo"]; len(sec) > 0 {
			sectionExists = true
		}
	}
	if sectionExists {
		sectionExists = false
		sec[0].ReadI32("partnerspacing", &s.partnerspacing)
		for i := range s.p {
			// Defaults
			if i >= 2 {
				s.p[i].startx = s.p[i-2].startx + s.partnerspacing*int32(2*(i%2)-1) // Previous partner + partnerspacing
				s.p[i].starty = s.p[i%2].starty                                     // Same as players 1 or 2
				s.p[i].startz = s.p[i%2].startz                                     // Same as players 1 or 2
				s.p[i].facing = int32(1 - 2*(i%2))                                  // By team side
			}
			// pXstartx
			sec[0].ReadI32(fmt.Sprintf("p%dstartx", i+1), &s.p[i].startx)
			// pXstarty
			sec[0].ReadI32(fmt.Sprintf("p%dstarty", i+1), &s.p[i].starty)
			// pXstartz
			sec[0].ReadI32(fmt.Sprintf("p%dstartz", i+1), &s.p[i].startz)
			// pXfacing
			sec[0].ReadI32(fmt.Sprintf("p%dfacing", i+1), &s.p[i].facing)
		}
		sec[0].ReadF32("leftbound", &s.leftbound)
		sec[0].ReadF32("rightbound", &s.rightbound)
		sec[0].ReadF32("topbound", &s.topbound)
		sec[0].ReadF32("botbound", &s.botbound)
	}

	// Camera group
	if sec := defmap["camera"]; len(sec) > 0 {
		sec[0].ReadI32("startx", &s.stageCamera.startx)
		sec[0].ReadI32("starty", &s.stageCamera.starty)
		sec[0].ReadI32("boundleft", &s.stageCamera.boundleft)
		sec[0].ReadI32("boundright", &s.stageCamera.boundright)
		sec[0].ReadI32("boundhigh", &s.stageCamera.boundhigh)
		sec[0].ReadI32("boundlow", &s.stageCamera.boundlow)
		sec[0].ReadF32("verticalfollow", &s.stageCamera.verticalfollow)
		sec[0].ReadI32("floortension", &s.stageCamera.floortension)
		sec[0].ReadI32("tension", &s.stageCamera.tension)
		sec[0].ReadF32("tensionvel", &s.stageCamera.tensionvel)
		sec[0].ReadI32("overdrawhigh", &s.stageCamera.overdrawhigh) // TODO: not implemented
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
		sec[0].ReadF32("zoominspeed", &s.stageCamera.zoominspeed)
		sec[0].ReadF32("zoomoutspeed", &s.stageCamera.zoomoutspeed)
		sec[0].ReadF32("yscrollspeed", &s.stageCamera.yscrollspeed)
		sec[0].ReadF32("boundhighzoomdelta", &s.stageCamera.boundhighzoomdelta)
		sec[0].ReadF32("verticalfollowzoomdelta", &s.stageCamera.verticalfollowzoomdelta)
		sec[0].ReadBool("lowestcap", &s.stageCamera.lowestcap)
		sec[0].ReadF32("zoomin", &s.stageCamera.zoomin)
		sec[0].ReadF32("zoomout", &s.stageCamera.zoomout)
		if s.stageCamera.zoomin == 1 && s.stageCamera.zoomout == 1 {
			if sys.cfg.Debug.ForceStageZoomin > 0 {
				s.stageCamera.zoomin = sys.cfg.Debug.ForceStageZoomin
			}
			if sys.cfg.Debug.ForceStageZoomout > 0 {
				s.stageCamera.zoomout = sys.cfg.Debug.ForceStageZoomout
			}
		}
		anchor, _, _ := sec[0].getText("zoomanchor")
		if strings.ToLower(anchor) == "bottom" {
			s.stageCamera.zoomanchor = true
		}
		if sec[0].ReadI32("tensionlow", &s.stageCamera.tensionlow) {
			s.stageCamera.ytensionenable = true
			sec[0].ReadI32("tensionhigh", &s.stageCamera.tensionhigh)
		}
		// Camera group warnings
		// Warn when camera boundaries are smaller than player boundaries
		if int32(s.leftbound) > s.stageCamera.boundleft || int32(s.rightbound) < s.stageCamera.boundright {
			sys.appendToConsole("Warning: Stage player boundaries defined incorrectly")
		}
	}

	// Music group
	if sec = defmap[fmt.Sprintf("%v.music", sys.cfg.Config.Language)]; len(sec) > 0 {
		sectionExists = true
	} else {
		if sec = defmap["music"]; len(sec) > 0 {
			sectionExists = true
		}
	}
	if sectionExists {
		sectionExists = false
		s.bgmusic = decodeShiftJIS(sec[0]["bgmusic"])
		sec[0].ReadI32("bgmvolume", &s.bgmvolume)
		sec[0].ReadI32("bgmloopstart", &s.bgmloopstart)
		sec[0].ReadI32("bgmloopend", &s.bgmloopend)
		sec[0].ReadI32("bgmstartposition", &s.bgmstartposition)
		sec[0].ReadF32("bgmfreqmul", &s.bgmfreqmul)
		sec[0].ReadI32("bgmratio.life", &s.bgmratiolife)
		sec[0].ReadI32("bgmtrigger.life", &s.bgmtriggerlife)
		sec[0].ReadI32("bgmtrigger.alt", &s.bgmtriggeralt)
	}

	// BGDef group
	if sec = defmap[fmt.Sprintf("%v.bgdef", sys.cfg.Config.Language)]; len(sec) > 0 {
		sectionExists = true
	} else {
		if sec = defmap["bgdef"]; len(sec) > 0 {
			sectionExists = true
		}
	}
	if sectionExists {
		sectionExists = false
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
			model, err := loadglTFModel(filename)
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

	// Model group
	if sec = defmap[fmt.Sprintf("%v.model", sys.cfg.Config.Language)]; len(sec) > 0 {
		sectionExists = true
	} else {
		if sec = defmap["model"]; len(sec) > 0 {
			sectionExists = true
		}
	}
	if sectionExists {
		sectionExists = false
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
		if err = sec[0].LoadFile("environment", []string{def, "", sys.motifDir, "data/"}, func(filename string) error {
			env, err := loadEnvironment(filename)
			if err != nil {
				return err
			}
			var intensity float32
			if sec[0].ReadF32("environmentintensity", &intensity) {
				env.environmentIntensity = intensity
			}
			s.model.environment = env
			return nil
		}); err != nil {
			return nil, err
		}
	}

	// Shadow group
	if sec = defmap[fmt.Sprintf("%v.shadow", sys.cfg.Config.Language)]; len(sec) > 0 {
		sectionExists = true
	} else {
		if sec = defmap["shadow"]; len(sec) > 0 {
			sectionExists = true
		}
	}
	if sectionExists {
		sectionExists = false
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
		sec[0].ReadF32("xscale", &s.sdw.xscale)
		sec[0].ReadF32("yscale", &s.sdw.yscale)
		sec[0].readI32ForStage("fade.range", &s.sdw.fadeend, &s.sdw.fadebgn)
		sec[0].ReadF32("xshear", &s.sdw.xshear)
		sec[0].ReadF32("angle", &s.sdw.rot.angle)
		sec[0].ReadF32("xangle", &s.sdw.rot.xangle)
		sec[0].ReadF32("yangle", &s.sdw.rot.yangle)
		sec[0].ReadF32("focallength", &s.sdw.fLength)
		if str, ok := sec[0]["projection"]; ok {
			switch strings.ToLower(strings.TrimSpace(str)) {
			case "orthographic":
				s.sdw.projection = Projection_Orthographic
			case "perspective":
				s.sdw.projection = Projection_Perspective
			case "perspective2":
				s.sdw.projection = Projection_Perspective2
			}
		}
		sec[0].readF32ForStage("offset", &s.sdw.offset[0], &s.sdw.offset[1])
		sec[0].readF32ForStage("window", &s.sdw.window[0], &s.sdw.window[1], &s.sdw.window[2], &s.sdw.window[3])
		sec[0].ReadF32("ydelta", &s.sdw.ydelta)
		// Shadow group warnings
		if s.sdw.fadeend > s.sdw.fadebgn {
			sys.appendToConsole("Warning: Stage shadow fade.range defined incorrectly")
		}
	}

	// Reflection group
	if sec = defmap[fmt.Sprintf("%v.reflection", sys.cfg.Config.Language)]; len(sec) > 0 {
		sectionExists = true
	} else {
		if sec = defmap["reflection"]; len(sec) > 0 {
			sectionExists = true
		}
	}
	if sectionExists {
		sectionExists = false
		s.reflection.color = 0xFFFFFF
		var tmp int32
		//sec[0].ReadBool("reflect", &reflect) // This parameter is documented in Mugen but doesn't do anything
		if sec[0].ReadI32("intensity", &tmp) {
			s.reflection.intensity = Clamp(tmp, 0, 255)
		}
		var r, g, b int32 = 0, 0, 0
		sec[0].readI32ForStage("color", &r, &g, &b)
		r, g, b = Clamp(r, 0, 255), Clamp(g, 0, 255), Clamp(b, 0, 255)
		s.reflection.color = uint32(r<<16 | g<<8 | b)
		if sec[0].ReadI32("layerno", &tmp) {
			s.reflection.layerno = Clamp(tmp, -1, 0)
		}
		sec[0].ReadF32("xscale", &s.reflection.xscale)
		sec[0].ReadF32("yscale", &s.reflection.yscale)
		sec[0].readI32ForStage("fade.range", &s.reflection.fadeend, &s.reflection.fadebgn)
		sec[0].ReadF32("xshear", &s.reflection.xshear)
		sec[0].ReadF32("angle", &s.reflection.rot.angle)
		sec[0].ReadF32("xangle", &s.reflection.rot.xangle)
		sec[0].ReadF32("yangle", &s.reflection.rot.yangle)
		sec[0].ReadF32("focallength", &s.reflection.fLength)
		if str, ok := sec[0]["projection"]; ok {
			switch strings.ToLower(strings.TrimSpace(str)) {
			case "orthographic":
				s.reflection.projection = Projection_Orthographic
			case "perspective":
				s.reflection.projection = Projection_Perspective
			case "perspective2":
				s.reflection.projection = Projection_Perspective2
			}
		}
		sec[0].readF32ForStage("offset", &s.reflection.offset[0], &s.reflection.offset[1])
		sec[0].readF32ForStage("window", &s.reflection.window[0], &s.reflection.window[1], &s.reflection.window[2], &s.reflection.window[3])
		sec[0].ReadF32("ydelta", &s.reflection.ydelta)
	}

	// BG group
	var bglink *backGround
	for _, bgsec := range defmap["bg"] {
		if len(s.bg) > 0 && !s.bg[len(s.bg)-1].positionlink {
			bglink = s.bg[len(s.bg)-1]
		}
		bg, err := readBackGround(bgsec, bglink, s.sff, s.animTable, s.stageprops, def)
		if err != nil {
			return nil, err
		}
		s.bg = append(s.bg, bg)
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
				uniqueIDs := make(map[int32]bool)
				for _, id := range ids {
					if uniqueIDs[id] {
						continue
					}
					bgcdef.bg = append(bgcdef.bg, s.getBg(id)...)
					uniqueIDs[id] = true
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
					uniqueIDs := make(map[int32]bool)
					for _, id := range ids {
						if uniqueIDs[id] {
							continue
						}
						bgc.bg = append(bgc.bg, s.getBg(id)...)
						uniqueIDs[id] = true
					}
				} else {
					bgc.bg = append(bgc.bg, s.bg...)
				}
			}
			if err := bgc.read(is, len(s.bgc)); err != nil {
				return nil, err
			}
			s.bgc = append(s.bgc, *bgc)
		case "bgctrl3d":
			bgc := newBgCtrl()
			*bgc = bgcdef
			bgc.bg = nil
			bgc.node = []*Node{}
			bgc.anim = []*GLTFAnimation{}
			if err := bgc.read(is, len(s.bgc)); err != nil {
				return nil, err
			}
			if ids := is.readI32CsvForStage("ctrlid"); len(ids) > 0 {
				if len(ids) > 1 || ids[0] != -1 {
					uniqueIDs := make(map[int32]bool)
					for _, id := range ids {
						if uniqueIDs[id] {
							continue
						}
						bgc.node = append(bgc.node, s.get3DBg(uint32(id))...)
						bgc.anim = append(bgc.anim, s.get3DAnim(uint32(id))...)
						uniqueIDs[id] = true
					}
				}
			}
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

	s.mainstage = maindef
	return s, nil
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

func (s *Stage) get3DBg(id uint32) (nodes []*Node) {
	if id >= 0 {
		for _, n := range s.model.nodes {
			if n.id == id {
				nodes = append(nodes, n)
			}
		}
	}
	return
}

func (s *Stage) get3DAnim(id uint32) (anims []*GLTFAnimation) {
	if id >= 0 {
		for _, a := range s.model.animations {
			if a.id == id {
				anims = append(anims, a)
			}
		}
	}
	return
}

// This essentially replaces the old timeline struct
func (s *Stage) bgCtrlAction() {
	for i := range s.bgc {
		bgc := &s.bgc[i]
		if bgc.starttime < 0 || (bgc.looptime >= 0 && bgc.starttime >= bgc.looptime) {
			continue
		}

		if bgc.looptime > 0 && bgc.endtime > bgc.looptime {
			bgc.endtime = bgc.looptime
		}

		active := false
		if s.stageTime >= bgc.starttime {
			if bgc.looptime > 0 {
				duration := bgc.endtime - bgc.starttime
				if (s.stageTime-bgc.starttime)%bgc.looptime <= duration {
					active = true
				}
			} else if s.stageTime <= bgc.endtime {
				active = true
			}
		}

		if active {
			s.runBgCtrl(bgc)
		}
	}
}

func (s *Stage) runBgCtrl(bgc *bgCtrl) {
	switch bgc._type {
	case BT_Anim:
		animNo := bgc.v[0]
		for i := range bgc.bg {
			bgc.bg[i].changeAnim(animNo, s.animTable)
		}
		for i := range bgc.anim {
			bgc.anim[i].toggle(animNo != 0)
		}
	case BT_Visible:
		for i := range bgc.bg {
			bgc.bg[i].visible = bgc.v[0] != 0
		}
		for i := range bgc.node {
			bgc.node[i].visible = bgc.v[0] != 0
		}
	case BT_Enable:
		for i := range bgc.bg {
			bgc.bg[i].enabled = bgc.v[0] != 0
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
			si, ok := s.sff.palList.PalTable[[...]uint16{uint16(bgc.src[0]), uint16(bgc.src[1])}]
			if !ok || si < 0 {
				return
			}
			var di int
			if bgc.dst[0] < 0 {
				// Set dest pal to source pal (remap gets reset)
				di = si
			} else {
				// Get dest pal
				di, ok = s.sff.palList.PalTable[[...]uint16{uint16(bgc.dst[0]), uint16(bgc.dst[1])}]
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

func (s *Stage) paused() bool {
	return (sys.supertime > 0 && sys.superpausebg) || (sys.pausetime > 0 && sys.pausebg)
}

func (s *Stage) action() {
	link, zlink := 0, -1
	canStep := sys.tickFrame() && !s.paused()

	// Update animations and controllers
	if canStep {
		s.bgCtrlAction()
		s.bga.action()

		if s.model != nil {
			s.model.step(sys.turbo)
		}
	}

	// Always (every frame) sync decoder run state to global pause + Enable.
	// This prevents the decoder clock from advancing during pause.
	for i := range s.bg {
		if s.bg[i]._type == BG_Video {
			shouldPlay := s.bg[i].enabled && canStep
			// Apply visibility first so there's no frame-0 audio when Visible=0.
			s.bg[i].video.SetVisible(s.bg[i].visible)
			s.bg[i].video.SetPlaying(shouldPlay)
		}
	}

	// Update BG elements
	for i, b := range s.bg {
		b.palfx.step()

		// BGPalFX can step even if the stage is paused
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
			b.palfx.eAllowNeg = sys.bgPalFX.eAllowNeg
		}

		if b.enabled && canStep {
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

	// Update model PalFX
	if s.model != nil {
		s.model.pfx.step()
		if sys.bgPalFX.enable {
			s.model.pfx.eAdd = sys.bgPalFX.eAdd
			s.model.pfx.eMul = sys.bgPalFX.eMul
			s.model.pfx.eColor = sys.bgPalFX.eColor
			s.model.pfx.eHue = sys.bgPalFX.eHue
			s.model.pfx.eInvertall = sys.bgPalFX.eInvertall
			s.model.pfx.eInvertblend = sys.bgPalFX.eInvertblend
			s.model.pfx.eAllowNeg = sys.bgPalFX.eAllowNeg
		}
	}
}

// Currently this function only exists so that the stage update sequence is similar to others. In the future it could run more tasks
// Doing this allows characters to see "stageTime = 0"
func (s *Stage) tick() {
	if s.paused() {
		return
	}

	// Stage time must be incremented after updating BGCtrl's
	// https://github.com/ikemen-engine/Ikemen-GO/issues/2656
	s.stageTime++
}

func (s *Stage) draw(layer int32, x, y, scl float32) {
	bgscl := float32(1)
	if s.hires {
		bgscl = 0.5
	}
	ofs := sys.envShake.getOffset()
	pos := [...]float32{x, y}
	scl2 := s.localscl * scl
	// This code makes the background scroll faster when surpassing boundhigh with the camera pushed down
	// through floortension and boundlow. MUGEN 1.1 doesn't look like it does this, so it was commented.
	// var extraBoundH float32
	// if sys.cam.zoomout < 1 {
	// extraBoundH = sys.cam.ExtraBoundH * ((1/scl)-1)/((1/sys.cam.zoomout)-1)
	// }
	// if pos[1] <= float32(s.stageCamera.boundlow) && pos[1] < float32(s.stageCamera.boundhigh)-extraBoundH {
	// ofs[1] += (pos[1]-float32(s.stageCamera.boundhigh))*scl2 +
	// extraBoundH*scl
	// pos[1] = float32(s.stageCamera.boundhigh) - extraBoundH/s.localscl
	// }
	if ofs[1] != 0 && s.stageCamera.verticalfollow > 0 {
		if ofs[1] < 0 {
			tmp := (float32(s.stageCamera.boundhigh) - pos[1]) * scl2
			if scl > 1 {
				tmp += (sys.cam.GroundLevel() + float32(sys.gameHeight-240)) * (1/scl - 1)
			} else {
				tmp += float32(sys.gameHeight) * (1/scl - 1)
			}
			if tmp >= 0 {
			} else if ofs[1] < tmp {
				ofs[1] -= tmp
				pos[1] += tmp / scl2
			} else {
				pos[1] += ofs[1] / scl2
				ofs[1] = 0
			}
		} else {
			if -ofs[1] >= pos[1]*scl2 {
				pos[1] += ofs[1] / scl2
				ofs[1] = 0
			}
		}
	}
	pos[0] += ofs[0] / scl2
	if !sys.cam.ZoomEnable {
		for i, p := range pos {
			pos[i] = float32(math.Ceil(float64(p - 0.5)))
		}
	}
	s.drawModel(pos, ofs[1], scl, layer)
	for _, b := range s.bg {
		// Draw only when visible and enabled.
		if b.layerno == layer && b.visible && b.enabled && (b.anim.spr != nil || b._type == BG_Video) {
			b.draw(pos, scl, bgscl, s.localscl, s.scale, ofs[1], true)
		}
	}
	BlendReset()
}

func (s *Stage) reset() {
	s.stageTime = 0
	s.sff.palList.ResetRemap()
	s.bga.clear()
	for i := range s.bg {
		s.bg[i].reset()
		// Ensure videos start paused, then rewind.
		if s.bg[i]._type == BG_Video {
			s.bg[i].video.SetPlaying(false)
			s.bg[i].video.Reset()
		}
	}
	if s.model != nil {
		s.model.reset()
	}
	// No need to reset BGCtrl at the moment. Tied to stagetime
}

// destroy stops any background video media so the stage can be safely discarded.
func (s *Stage) destroy() {
	for _, b := range s.bg {
		if b != nil && b._type == BG_Video {
			b.video.Close()
		}
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

func (s *Stage) modifyBGCtrl3d(id uint32, t, v [3]int32) {
	for i := range s.bgc {
		for j := range s.bgc[i].node {
			if s.bgc[i].node[j].id == id {
				if t[0] != IErr {
					s.bgc[i].starttime = t[0]
				}
				if t[1] != IErr {
					s.bgc[i].endtime = t[1]
				}
				if t[2] != IErr {
					s.bgc[i].looptime = t[2]
				}

				for k := 0; k < 3; k++ {
					if v[k] != IErr {
						s.bgc[i].v[k] = v[k]
					}
				}
				s.reload = true
			}
		}
		for j := range s.bgc[i].anim {
			if s.bgc[i].anim[j].id == id {
				s.bgc[i].anim[j].time = 0
				s.bgc[i].anim[j].enabled = v[0] != 0
				s.bgc[i].anim[j].loop = 0
				s.reload = true
			}
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
	lights              []GLTFLight
	environment         *Environment
	//lightNodes           []int32
	//lightNodesForeground []int32
}

type Scene struct {
	nodes           []uint32
	name            string
	lightNodes      []uint32
	imageBasedLight *uint32
}

type LightType byte

const (
	DirectionalLight = iota
	PointLight
	SpotLight
)

type GLTFLight struct {
	direction       [3]float32
	position        [3]float32
	lightRange      GLTFAnimatableProperty // float32
	color           GLTFAnimatableProperty // [3]float32
	intensity       GLTFAnimatableProperty // float32
	innerConeAngle  GLTFAnimatableProperty // float32
	outerConeAngle  GLTFAnimatableProperty // float32
	lightType       LightType
	shadowMapNear   GLTFAnimatableProperty // float32
	shadowMapFar    GLTFAnimatableProperty // float32
	shadowMapBottom GLTFAnimatableProperty // float32
	shadowMapTop    GLTFAnimatableProperty // float32
	shadowMapLeft   GLTFAnimatableProperty // float32
	shadowMapRight  GLTFAnimatableProperty // float32
	shadowMapBias   GLTFAnimatableProperty // float32
}

type GLTFAnimationType byte

const (
	TRSTranslation = iota
	TRSScale
	TRSRotation
	MorphTargetWeight
	AnimVec2
	AnimVec3
	AnimVec4
	AnimVecElem
	AnimVec
	AnimQuat
	AnimFloat
)

type GLTFAnimationInterpolation byte

const (
	InterpolationLinear = iota
	InterpolationStep
	InterpolationCubicSpline
)

type GLTFAnimation struct {
	id             uint32
	name           string
	enabled        bool
	defaultEnabled bool
	duration       float32
	time           float32
	loopCount      int32
	loop           int32
	channels       []*GLTFAnimationChannel
	samplers       []*GLTFAnimationSampler
}

type GLTFAnimationChannel struct {
	//path         GLTFAnimationType
	target       *GLTFAnimatableProperty
	targetType   GLTFAnimationType
	elemIndex    *uint32
	nodeIndex    *uint32
	samplerIndex uint32
}

type GLTFAnimationSampler struct {
	inputIndex    uint32
	output        []float32
	interpolation GLTFAnimationInterpolation
}

type GLTFTexture struct {
	tex Texture
}

type GLTFAnimatableProperty struct {
	restValue     interface{}
	animatedValue interface{}
	isAnimated    bool
}

func (p *GLTFAnimatableProperty) rest() {
	p.isAnimated = false
}

func (p *GLTFAnimatableProperty) restAt(v interface{}) {
	p.restValue = v
}

func (p *GLTFAnimatableProperty) animate(v interface{}) {
	p.isAnimated = true
	p.animatedValue = v
}

func (p *GLTFAnimatableProperty) getValue() interface{} {
	if p.isAnimated {
		return p.animatedValue
	}
	return p.restValue
}

type AlphaMode byte

const (
	AlphaModeOpaque = iota
	AlphaModeMask
	AlphaModeBlend
)

type Material struct {
	name                          string
	alphaMode                     AlphaMode
	alphaCutoff                   GLTFAnimatableProperty // float32
	textureIndex                  *uint32
	textureOffset                 GLTFAnimatableProperty // [3]float32
	textureRotation               GLTFAnimatableProperty // [4]float32
	textureScale                  GLTFAnimatableProperty // [3]float32
	textureTransform              [9]float32
	normalMapIndex                *uint32
	normalMapOffset               GLTFAnimatableProperty // [3]float32
	normalMapRotation             GLTFAnimatableProperty // [4]float32
	normalMapScale                GLTFAnimatableProperty // [3]float32
	normalMapTransform            [9]float32
	ambientOcclusionMapIndex      *uint32
	ambientOcclusionMapOffset     GLTFAnimatableProperty // [3]float32
	ambientOcclusionMapRotation   GLTFAnimatableProperty // [4]float32
	ambientOcclusionMapScale      GLTFAnimatableProperty // [3]float32
	ambientOcclusionMapTransform  [9]float32
	metallicRoughnessMapIndex     *uint32
	metallicRoughnessMapOffset    GLTFAnimatableProperty // [3]float32
	metallicRoughnessMapRotation  GLTFAnimatableProperty // [4]float32
	metallicRoughnessMapScale     GLTFAnimatableProperty // [3]float32
	metallicRoughnessMapTransform [9]float32
	emissionMapIndex              *uint32
	emissionMapOffset             GLTFAnimatableProperty // [3]float32
	emissionMapRotation           GLTFAnimatableProperty // [4]float32
	emissionMapScale              GLTFAnimatableProperty // [3]float32
	emissionMapTransform          [9]float32
	baseColorFactor               GLTFAnimatableProperty // [4]float32
	doubleSided                   bool
	ambientOcclusion              GLTFAnimatableProperty // float32
	metallic                      GLTFAnimatableProperty // float32
	roughness                     GLTFAnimatableProperty // float32
	emission                      GLTFAnimatableProperty // [3]float32
	unlit                         bool
}

type Trans byte

const (
	TransNone = iota
	TransAdd
	TransReverseSubtract
	TransMul
)

type Node struct {
	id                 uint32
	visible            bool
	meshIndex          *uint32
	transition         GLTFAnimatableProperty // [3]float32
	rotation           GLTFAnimatableProperty // [4]float32
	scale              GLTFAnimatableProperty // [3]float32
	transformChanged   bool
	localTransform     mgl.Mat4
	worldTransform     mgl.Mat4
	normalMatrix       mgl.Mat4
	childrenIndex      []uint32
	trans              Trans
	castShadow         bool
	zWrite             bool
	zTest              bool
	parentIndex        *uint32
	lightIndex         *uint32
	lightDirection     [3]float32
	shadowMapNear      GLTFAnimatableProperty // float32
	shadowMapFar       GLTFAnimatableProperty // float32
	shadowMapBottom    GLTFAnimatableProperty // float32
	shadowMapTop       GLTFAnimatableProperty // float32
	shadowMapLeft      GLTFAnimatableProperty // float32
	shadowMapRight     GLTFAnimatableProperty // float32
	shadowMapBias      GLTFAnimatableProperty // float32
	skin               *uint32
	morphTargetWeights GLTFAnimatableProperty // []float32
	activeMorphTargets []uint32
	layerNumber        *int
	meshOutline        GLTFAnimatableProperty // float32
}

type Skin struct {
	joints              []uint32
	inverseBindMatrices []float32
	texture             *GLTFTexture
}

type Mesh struct {
	name               string
	morphTargetWeights GLTFAnimatableProperty // []float32
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
	positionIndex  *uint32
	normalIndex    *uint32
	tangentIndex   *uint32
	uvIndex        *uint32
	colorIndex     *uint32
	targetType     uint32
	offset         uint32
	positionBuffer []float32
	uvBuffer       []float32
	normalBuffer   []float32
	tangentBuffer  []float32
	colorBuffer    []float32
}

type Primitive struct {
	numVertices         uint32
	numIndices          uint32
	vertexBufferOffset  uint32
	elementBufferOffset uint32
	materialIndex       *uint32
	useUV               bool
	useNormal           bool
	useTangent          bool
	useVertexColor      bool
	useJoint0           bool
	useJoint1           bool
	useOutlineAttribute bool
	mode                PrimitiveMode
	morphTargets        []*MorphTarget
	morphTargetTexture  *GLTFTexture
	morphTargetCount    uint32
	morphTargetOffset   [4]float32
	morphTargetWeight   [8]float32
	boundingBox         BoundingBox
}

type BoundingBox struct {
	min [3]float32
	max [3]float32
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

type Environment struct {
	hdrTexture            *GLTFTexture
	cubeMapTexture        *GLTFTexture
	lambertianTexture     *GLTFTexture
	GGXTexture            *GLTFTexture
	GGXLUT                *GLTFTexture
	lambertianSampleCount int32
	GGXSampleCount        int32
	GGXLUTSampleCount     int32
	mipmapLevels          int32
	environmentIntensity  float32
}

func loadEnvironment(filepath string) (*Environment, error) {
	env := &Environment{}
	env.lambertianSampleCount = 2048
	env.GGXSampleCount = 1024
	env.GGXLUTSampleCount = 512
	env.environmentIntensity = 1
	file, err := OpenFile(filepath)
	if err != nil {
		return nil, err
	}
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	env.hdrTexture = &GLTFTexture{}
	env.cubeMapTexture = &GLTFTexture{}
	env.lambertianTexture = &GLTFTexture{}
	env.GGXTexture = &GLTFTexture{}
	env.GGXLUT = &GLTFTexture{}
	if hdrImg, ok := img.(hdr.Image); ok {
		size := img.Bounds().Max.X * img.Bounds().Max.Y * 3
		data := make([]float32, size, size)
		bounds := img.Bounds()
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				color := hdrImg.HDRAt(x, y)
				r, g, b, _ := color.HDRRGBA()
				data = append(data, float32(r), float32(g), float32(b))
			}
		}
		for i, j := 0, len(data)-3; i < j; i, j = i+3, j-3 {
			data[i], data[i+1], data[i+2], data[j], data[j+1], data[j+2] = data[j], data[j+1], data[j+2], data[i], data[i+1], data[i+2]
		}
		sys.mainThreadTask <- func() {
			if !gfx.IsModelEnabled() {
				return
			}
			lowestMipLevel := int32(4)
			env.hdrTexture.tex = gfx.newHDRTexture(int32(img.Bounds().Max.X), int32(img.Bounds().Max.Y))

			env.hdrTexture.tex.SetPixelData(data)
			env.cubeMapTexture.tex = gfx.newCubeMapTexture(256, true, 0)
			env.lambertianTexture.tex = gfx.newCubeMapTexture(256, false, 0)
			env.GGXTexture.tex = gfx.newCubeMapTexture(256, true, lowestMipLevel)
			env.GGXLUT.tex = gfx.newDataTexture(1024, 1024)

			gfx.RenderCubeMap(env.hdrTexture.tex, env.cubeMapTexture.tex)
			gfx.RenderFilteredCubeMap(0, env.cubeMapTexture.tex, env.lambertianTexture.tex, 0, env.lambertianSampleCount, 0)
			env.mipmapLevels = int32(Floor(float32(math.Log2(256)))) + 1 - lowestMipLevel
			for i := int32(0); i < env.mipmapLevels; i++ {
				roughness := float32(i) / float32((env.mipmapLevels - 1))
				gfx.RenderFilteredCubeMap(1, env.cubeMapTexture.tex, env.GGXTexture.tex, int32(i), env.GGXSampleCount, roughness)
			}
			gfx.RenderLUT(1, env.cubeMapTexture.tex, env.GGXLUT.tex, env.GGXLUTSampleCount)
		}
	}
	return env, nil
}

func loadglTFModel(filepath string) (*Model, error) {
	mdl := &Model{offset: [3]float32{0, 0, 0}, rotation: [3]float32{0, 0, 0}, scale: [3]float32{1, 1, 1}}

	isZip, zipPath, pathInZip := IsZipPath(filepath)

	var doc *gltf.Document
	var err error

	if isZip {
		// Handle resources within a ZIP file
		zipReader, errOpen := zip.OpenReader(zipPath)
		if errOpen != nil {
			return nil, fmt.Errorf("failed to open zip archive '%s': %w", zipPath, errOpen)
		}
		defer zipReader.Close()

		var fsys fs.FS = &zipReader.Reader // The zip.Reader implements fs.FS for resource resolution within the archive

		// Open the GLB/glTF file from within the zip
		glbFile, errOpen := fsys.Open(pathInZip)
		if errOpen != nil {
			return nil, fmt.Errorf("failed to open glb file '%s' in zip '%s': %w", pathInZip, zipPath, errOpen)
		}
		defer glbFile.Close()

		// Create a new decoder with the file stream and the zip archive as the file system
		decoder := gltf.NewDecoderFS(glbFile, fsys)
		doc = new(gltf.Document)
		if err = decoder.Decode(doc); err != nil {
			return nil, fmt.Errorf("failed to decode gltf from zip '%s': %w", filepath, err)
		}
	} else {
		// Handle resources from the standard file system
		f, errOpen := OpenFile(filepath) // Use IKEMEN GO's file abstraction to handle case-insensitivity etc.
		if errOpen != nil {
			return nil, errOpen
		}
		defer f.Close()

		// Use the directory of the file as the file system for resolving relative resources
		fsm := os.DirFS(path.Dir(filepath))
		decoder := gltf.NewDecoderFS(f, fsm)
		doc = new(gltf.Document)
		if err = decoder.Decode(doc); err != nil {
			return nil, fmt.Errorf("failed to decode gltf from file '%s': %w", filepath, err)
		}
	}

	if doc == nil {
		return nil, fmt.Errorf("gltf document is nil after decoding for path: %s", filepath)
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
					// Use OpenFile which respects the virtual file system (zip)
					f, err := OpenFile(filename)
					if err != nil {
						return err
					}
					defer f.Close()
					data, err := io.ReadAll(f)
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
				mag, _ := map[gltf.MagFilter]TextureSamplingParam{
					gltf.MagUndefined: TextureSamplingFilterLinear,
					gltf.MagNearest:   TextureSamplingFilterNearest,
					gltf.MagLinear:    TextureSamplingFilterLinear,
				}[s.MagFilter]
				min, _ := map[gltf.MinFilter]TextureSamplingParam{
					gltf.MinUndefined:            TextureSamplingFilterLinear,
					gltf.MinNearest:              TextureSamplingFilterLinear,
					gltf.MinLinear:               TextureSamplingFilterLinear,
					gltf.MinNearestMipMapNearest: TextureSamplingFilterNearestMipMapNearest,
					gltf.MinLinearMipMapNearest:  TextureSamplingFilterLinearMipMapNearest,
					gltf.MinNearestMipMapLinear:  TextureSamplingFilterNearestMipMapLinear,
					gltf.MinLinearMipMapLinear:   TextureSamplingFilterLinearMipMapLinear,
				}[s.MinFilter]
				wrapS, _ := map[gltf.WrappingMode]TextureSamplingParam{
					gltf.WrapClampToEdge:    TextureSamplingWrapClampToEdge,
					gltf.WrapMirroredRepeat: TextureSamplingWrapMirroredRepeat,
					gltf.WrapRepeat:         TextureSamplingWrapRepeat,
				}[s.WrapS]
				wrapT, _ := map[gltf.WrappingMode]TextureSamplingParam{
					gltf.WrapClampToEdge:    TextureSamplingWrapClampToEdge,
					gltf.WrapMirroredRepeat: TextureSamplingWrapMirroredRepeat,
					gltf.WrapRepeat:         TextureSamplingWrapRepeat,
				}[s.WrapT]

				img := images[*t.Source]
				rgba := image.NewRGBA(img.Bounds())
				draw.Draw(rgba, img.Bounds(), img, img.Bounds().Min, draw.Src)
				sys.mainThreadTask <- func() {
					texture.tex = gfx.newModelTexture(int32(img.Bounds().Max.X), int32(img.Bounds().Max.Y), 32, false)
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
				mag := TextureSamplingFilterNearest
				min := TextureSamplingFilterNearest
				wrapS := TextureSamplingWrapRepeat
				wrapT := TextureSamplingWrapRepeat
				img := images[*t.Source]
				rgba := image.NewRGBA(img.Bounds())
				draw.Draw(rgba, img.Bounds(), img, img.Bounds().Min, draw.Src)
				sys.mainThreadTask <- func() {
					texture.tex = gfx.newModelTexture(int32(img.Bounds().Max.X), int32(img.Bounds().Max.Y), 32, false)
					texture.tex.SetDataG(rgba.Pix, mag, min, wrapS, wrapT)
				}
				textureMap[[2]int32{int32(*t.Source), -1}] = texture
				mdl.textures = append(mdl.textures, texture)
			}
		}

	}
	mdl.materials = make([]*Material, 0, len(doc.Materials))
	for _, m := range doc.Materials {
		material := &Material{
			baseColorFactor:              GLTFAnimatableProperty{isAnimated: false, restValue: [4]float32{1, 1, 1, 1}},
			roughness:                    GLTFAnimatableProperty{isAnimated: false, restValue: float32(1.0)},
			metallic:                     GLTFAnimatableProperty{isAnimated: false, restValue: float32(1.0)},
			ambientOcclusion:             GLTFAnimatableProperty{isAnimated: false, restValue: float32(1.0)},
			emission:                     GLTFAnimatableProperty{isAnimated: false, restValue: [3]float32{0, 0, 0}},
			alphaCutoff:                  GLTFAnimatableProperty{isAnimated: false, restValue: float32(0.0)},
			textureOffset:                GLTFAnimatableProperty{isAnimated: false, restValue: [2]float32{0.0, 0.0}},
			textureRotation:              GLTFAnimatableProperty{isAnimated: false, restValue: float32(0.0)},
			textureScale:                 GLTFAnimatableProperty{isAnimated: false, restValue: [2]float32{1.0, 1.0}},
			normalMapOffset:              GLTFAnimatableProperty{isAnimated: false, restValue: [2]float32{0.0, 0.0}},
			normalMapRotation:            GLTFAnimatableProperty{isAnimated: false, restValue: float32(0.0)},
			normalMapScale:               GLTFAnimatableProperty{isAnimated: false, restValue: [2]float32{1.0, 1.0}},
			ambientOcclusionMapOffset:    GLTFAnimatableProperty{isAnimated: false, restValue: [2]float32{0.0, 0.0}},
			ambientOcclusionMapRotation:  GLTFAnimatableProperty{isAnimated: false, restValue: float32(0.0)},
			ambientOcclusionMapScale:     GLTFAnimatableProperty{isAnimated: false, restValue: [2]float32{1.0, 1.0}},
			metallicRoughnessMapOffset:   GLTFAnimatableProperty{isAnimated: false, restValue: [2]float32{0.0, 0.0}},
			metallicRoughnessMapRotation: GLTFAnimatableProperty{isAnimated: false, restValue: float32(0.0)},
			metallicRoughnessMapScale:    GLTFAnimatableProperty{isAnimated: false, restValue: [2]float32{1.0, 1.0}},
			emissionMapOffset:            GLTFAnimatableProperty{isAnimated: false, restValue: [2]float32{0.0, 0.0}},
			emissionMapRotation:          GLTFAnimatableProperty{isAnimated: false, restValue: float32(0.0)},
			emissionMapScale:             GLTFAnimatableProperty{isAnimated: false, restValue: [2]float32{1.0, 1.0}},
		}
		if m.PBRMetallicRoughness.BaseColorTexture != nil {
			material.textureIndex = new(uint32)
			*material.textureIndex = m.PBRMetallicRoughness.BaseColorTexture.Index
			if m.PBRMetallicRoughness.BaseColorTexture.Extensions != nil {
				if l, ok := m.PBRMetallicRoughness.BaseColorTexture.Extensions["KHR_texture_transform"]; ok {
					var ext interface{}
					err := json.Unmarshal(l.(json.RawMessage), &ext)
					if err != nil {
						return nil, err
					}
					if offset, ok := ext.(map[string]interface{})["offset"].([]interface{}); ok {
						material.textureOffset.restAt([2]float32{float32(offset[0].(float64)), float32(offset[1].(float64))})
					}
					if rotation, ok := ext.(map[string]interface{})["rotation"].(float64); ok {
						material.textureRotation.restAt(float32(rotation))
					}
					if scale, ok := ext.(map[string]interface{})["scale"].([]interface{}); ok {
						material.textureScale.restAt([2]float32{float32(scale[0].(float64)), float32(scale[1].(float64))})
					}
				}
			}
		}
		if m.NormalTexture != nil {
			material.normalMapIndex = new(uint32)
			*material.normalMapIndex = *m.NormalTexture.Index
			if m.NormalTexture.Extensions != nil {
				if l, ok := m.NormalTexture.Extensions["KHR_texture_transform"]; ok {
					var ext interface{}
					err := json.Unmarshal(l.(json.RawMessage), &ext)
					if err != nil {
						return nil, err
					}
					if offset, ok := ext.(map[string]interface{})["offset"].([]interface{}); ok {
						material.normalMapOffset.restAt([2]float32{float32(offset[0].(float64)), float32(offset[1].(float64))})
					}
					if rotation, ok := ext.(map[string]interface{})["rotation"].(float64); ok {
						material.normalMapRotation.restAt(float32(rotation))
					}
					if scale, ok := ext.(map[string]interface{})["scale"].([]interface{}); ok {
						material.normalMapScale.restAt([2]float32{float32(scale[0].(float64)), float32(scale[1].(float64))})
					}
				}
			}
		}
		if m.PBRMetallicRoughness.MetallicRoughnessTexture != nil {
			material.metallicRoughnessMapIndex = new(uint32)
			*material.metallicRoughnessMapIndex = m.PBRMetallicRoughness.MetallicRoughnessTexture.Index
			if m.PBRMetallicRoughness.MetallicRoughnessTexture.Extensions != nil {
				if l, ok := m.PBRMetallicRoughness.MetallicRoughnessTexture.Extensions["KHR_texture_transform"]; ok {
					var ext interface{}
					err := json.Unmarshal(l.(json.RawMessage), &ext)
					if err != nil {
						return nil, err
					}
					if offset, ok := ext.(map[string]interface{})["offset"].([]interface{}); ok {
						material.metallicRoughnessMapOffset.restAt([2]float32{float32(offset[0].(float64)), float32(offset[1].(float64))})
					}
					if rotation, ok := ext.(map[string]interface{})["rotation"].(float64); ok {
						material.metallicRoughnessMapRotation.restAt(float32(rotation))
					}
					if scale, ok := ext.(map[string]interface{})["scale"].([]interface{}); ok {
						material.metallicRoughnessMapScale.restAt([2]float32{float32(scale[0].(float64)), float32(scale[1].(float64))})
					}
				}
			}
		}
		if m.PBRMetallicRoughness.BaseColorFactor != nil {
			material.baseColorFactor.restAt(*m.PBRMetallicRoughness.BaseColorFactor)
		}
		if m.PBRMetallicRoughness.RoughnessFactor != nil {
			material.roughness.restAt(*m.PBRMetallicRoughness.RoughnessFactor)
		}
		if m.PBRMetallicRoughness.MetallicFactor != nil {
			material.metallic.restAt(*m.PBRMetallicRoughness.MetallicFactor)
		}

		if m.OcclusionTexture != nil {
			material.ambientOcclusionMapIndex = new(uint32)
			*material.ambientOcclusionMapIndex = *m.OcclusionTexture.Index
			if m.OcclusionTexture.Strength != nil {
				material.ambientOcclusion.restAt(*m.OcclusionTexture.Strength)
			}
			if m.OcclusionTexture.Extensions != nil {
				if l, ok := m.OcclusionTexture.Extensions["KHR_texture_transform"]; ok {
					var ext interface{}
					err := json.Unmarshal(l.(json.RawMessage), &ext)
					if err != nil {
						return nil, err
					}
					if offset, ok := ext.(map[string]interface{})["offset"].([]interface{}); ok {
						material.ambientOcclusionMapOffset.restAt([2]float32{float32(offset[0].(float64)), float32(offset[1].(float64))})
					}
					if rotation, ok := ext.(map[string]interface{})["rotation"].(float64); ok {
						material.ambientOcclusionMapRotation.restAt(float32(rotation))
					}
					if scale, ok := ext.(map[string]interface{})["scale"].([]interface{}); ok {
						material.ambientOcclusionMapScale.restAt([2]float32{float32(scale[0].(float64)), float32(scale[1].(float64))})
					}
				}
			}
		} else {
			material.ambientOcclusion.restAt(float32(0))
		}
		material.emission.restAt(m.EmissiveFactor)
		if m.EmissiveTexture != nil {
			material.emissionMapIndex = new(uint32)
			*material.emissionMapIndex = m.EmissiveTexture.Index
			if m.EmissiveTexture.Extensions != nil {
				if l, ok := m.EmissiveTexture.Extensions["KHR_texture_transform"]; ok {
					var ext interface{}
					err := json.Unmarshal(l.(json.RawMessage), &ext)
					if err != nil {
						return nil, err
					}
					if offset, ok := ext.(map[string]interface{})["offset"].([]interface{}); ok {
						material.emissionMapOffset.restAt([2]float32{float32(offset[0].(float64)), float32(offset[1].(float64))})
					}
					if rotation, ok := ext.(map[string]interface{})["rotation"].(float64); ok {
						material.emissionMapRotation.restAt(float32(rotation))
					}
					if scale, ok := ext.(map[string]interface{})["scale"].([]interface{}); ok {
						material.emissionMapScale.restAt([2]float32{float32(scale[0].(float64)), float32(scale[1].(float64))})
					}
				}
			}
		}
		material.name = m.Name
		material.alphaMode, _ = map[gltf.AlphaMode]AlphaMode{
			gltf.AlphaOpaque: AlphaModeOpaque,
			gltf.AlphaMask:   AlphaModeMask,
			gltf.AlphaBlend:  AlphaModeBlend,
		}[m.AlphaMode]
		if material.alphaMode == AlphaModeMask {
			material.alphaCutoff.restAt(m.AlphaCutoffOrDefault())
		}
		material.doubleSided = m.DoubleSided
		material.unlit = false
		if m.Extensions != nil {
			if _, ok := m.Extensions["KHR_materials_unlit"]; ok {
				material.unlit = true
			}
		}

		mdl.materials = append(mdl.materials, material)
	}
	if doc.Extensions != nil {
		if lightExtension, ok := doc.Extensions["KHR_lights_punctual"]; ok {
			var ext interface{}
			err := json.Unmarshal(lightExtension.(json.RawMessage), &ext)
			if err != nil {
				return nil, err
			}
			for _, light := range ext.(map[string]interface{})["lights"].([]interface{}) {
				params := light.(map[string]interface{})
				newLight := GLTFLight{
					intensity:       GLTFAnimatableProperty{isAnimated: false, restValue: float32(1.0)},
					color:           GLTFAnimatableProperty{restValue: [3]float32{1, 1, 1}},
					lightRange:      GLTFAnimatableProperty{isAnimated: false, restValue: float32(-1.0)},
					innerConeAngle:  GLTFAnimatableProperty{isAnimated: false, restValue: float32(0.0)},
					outerConeAngle:  GLTFAnimatableProperty{isAnimated: false, restValue: float32(math.Pi / 4)},
					shadowMapNear:   GLTFAnimatableProperty{isAnimated: false, restValue: float32(0.0)},
					shadowMapFar:    GLTFAnimatableProperty{isAnimated: false, restValue: float32(0.0)},
					shadowMapBottom: GLTFAnimatableProperty{isAnimated: false, restValue: float32(0.0)},
					shadowMapTop:    GLTFAnimatableProperty{isAnimated: false, restValue: float32(0.0)},
					shadowMapLeft:   GLTFAnimatableProperty{isAnimated: false, restValue: float32(0.0)},
					shadowMapRight:  GLTFAnimatableProperty{isAnimated: false, restValue: float32(0.0)},
					shadowMapBias:   GLTFAnimatableProperty{isAnimated: false, restValue: float32(0.0)},
				}
				lightType := params["type"].(string)
				switch lightType {
				case "point":
					newLight.lightType = PointLight
				case "spot":
					newLight.lightType = SpotLight
				case "directional":
					newLight.lightType = DirectionalLight
				}
				if intensity, ok := params["intensity"]; ok {
					newLight.intensity.restAt((float32)(intensity.(float64)))
				}
				if lightRange, ok := params["range"]; ok {
					newLight.lightRange.restAt((float32)(lightRange.(float64)))
				}
				if spot, ok := params["spot"]; ok {
					if outerConeAngle, ok := spot.(map[string]interface{})["outerConeAngle"]; ok {
						newLight.outerConeAngle.restAt((float32)(outerConeAngle.(float64)))
					}
					if innerConeAngle, ok := spot.(map[string]interface{})["innerConeAngle"]; ok {
						newLight.innerConeAngle.restAt((float32)(innerConeAngle.(float64)))
					}
				}
				if color, ok := params["color"]; ok {
					colors := color.([]interface{})
					newLight.color.restAt([3]float32{(float32)(colors[0].(float64)), (float32)(colors[1].(float64)), (float32)(colors[2].(float64))})
				}

				if extraParams, ok := params["extras"]; ok {

					v, ok := extraParams.(map[string]interface{})
					if ok {
						if v["shadowMapNear"] != nil {
							newLight.shadowMapNear.restAt((float32)(v["shadowMapNear"].(float64)))
						}
						if v["shadowMapFar"] != nil {
							newLight.shadowMapFar.restAt((float32)(v["shadowMapFar"].(float64)))
						}
						if v["shadowMapBottom"] != nil {
							newLight.shadowMapBottom.restAt((float32)(v["shadowMapBottom"].(float64)))
						}
						if v["shadowMapTop"] != nil {
							newLight.shadowMapTop.restAt((float32)(v["shadowMapTop"].(float64)))
						}
						if v["shadowMapLeft"] != nil {
							newLight.shadowMapLeft.restAt((float32)(v["shadowMapLeft"].(float64)))
						}
						if v["shadowMapRight"] != nil {
							newLight.shadowMapRight.restAt((float32)(v["shadowMapRight"].(float64)))
						}
						if v["shadowMapBias"] != nil {
							newLight.shadowMapBias.restAt((float32)(v["shadowMapBias"].(float64)))
						}
					}
				}
				mdl.lights = append(mdl.lights, newLight)
			}
		}
	}

	var vertexBuffer []byte
	var elementBuffer []uint32
	mdl.meshes = make([]*Mesh, 0, len(doc.Meshes))
	for _, m := range doc.Meshes {
		var mesh = &Mesh{}
		mesh.name = m.Name
		mesh.morphTargetWeights = GLTFAnimatableProperty{isAnimated: false, restValue: m.Weights}
		for _, p := range m.Primitives {
			var primitive = &Primitive{}
			primitive.boundingBox.min = [3]float32{math.MaxFloat32, math.MaxFloat32, math.MaxFloat32}
			primitive.boundingBox.max = [3]float32{-math.MaxFloat32, -math.MaxFloat32, -math.MaxFloat32}
			primitive.vertexBufferOffset = uint32(len(vertexBuffer))
			primitive.elementBufferOffset = uint32(4 * len(elementBuffer))
			var posBuffer [][3]float32
			positions, err := modeler.ReadPosition(doc, doc.Accessors[p.Attributes[gltf.POSITION]], posBuffer)
			if err != nil {
				return nil, err
			}
			primitive.numVertices = uint32(len(positions))

			for i := 0; i < int(primitive.numVertices); i++ {
				vertexBuffer = append(vertexBuffer, byte(i%256), byte((i>>8)%256), byte((i>>16)%256), byte((i>>32)%256))
			}

			for _, pos := range positions {
				for posIdx := range pos {
					if primitive.boundingBox.min[posIdx] > pos[posIdx] {
						primitive.boundingBox.min[posIdx] = pos[posIdx]
					}
					if primitive.boundingBox.max[posIdx] < pos[posIdx] {
						primitive.boundingBox.max[posIdx] = pos[posIdx]
					}
				}
				vertexBuffer = append(vertexBuffer, f32.Bytes(binary.LittleEndian, pos[:]...)...)
			}
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
			if idx, ok := p.Attributes[gltf.NORMAL]; ok {
				var normalBuffer [][3]float32
				normals, err := modeler.ReadNormal(doc, doc.Accessors[idx], normalBuffer)
				if err != nil {
					return nil, err
				}
				if len(normals) > 0 {
					primitive.useNormal = true
					for _, tex := range normals {
						vertexBuffer = append(vertexBuffer, f32.Bytes(binary.LittleEndian, tex[:]...)...)
					}
				} else {
					primitive.useNormal = false
				}
			} else {
				primitive.useNormal = false
			}
			if idx, ok := p.Attributes[gltf.TANGENT]; ok {
				var tangentBuffer [][4]float32
				tangents, err := modeler.ReadTangent(doc, doc.Accessors[idx], tangentBuffer)
				if err != nil {
					return nil, err
				}
				if len(tangents) > 0 {
					primitive.useTangent = true
					for _, tex := range tangents {
						vertexBuffer = append(vertexBuffer, f32.Bytes(binary.LittleEndian, tex[:]...)...)
					}
				} else {
					primitive.useTangent = false
				}
			} else {
				primitive.useTangent = false
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

			if idx, ok := p.Attributes["_OUTLINE_ATTRIBUTE"]; ok {
				primitive.useOutlineAttribute = true
				var outlineAttributeBuffer [][4]uint16
				atttributes, err := modeler.ReadAccessor(doc, doc.Accessors[idx], outlineAttributeBuffer)
				if err != nil {
					return nil, err
				}
				for _, attribute := range atttributes.([][4]float32) {
					vertexBuffer = append(vertexBuffer, f32.Bytes(binary.LittleEndian, attribute[:]...)...)
				}
			} else {
				primitive.useOutlineAttribute = false
			}
			if len(p.Targets) > 0 {
				numAttributes := 0
				for _, t := range p.Targets {
					numAttributes += len(t)
				}
				for _, t := range p.Targets {
					target := &MorphTarget{}
					for attr, accessor := range t {
						switch attr {
						case "POSITION":
							var posBuffer [][3]float32
							positions, err := modeler.ReadPosition(doc, doc.Accessors[accessor], posBuffer)
							if err != nil {
								return nil, err
							}
							target.positionBuffer = make([]float32, 0, 4*int(primitive.numVertices))
							for _, pos := range positions {
								target.positionBuffer = append(target.positionBuffer, pos[0], pos[1], pos[2], 0)
							}
						case "NORMAL":
							var posBuffer [][3]float32
							positions, err := modeler.ReadPosition(doc, doc.Accessors[accessor], posBuffer)
							if err != nil {
								return nil, err
							}
							target.normalBuffer = make([]float32, 0, 4*int(primitive.numVertices))
							for _, pos := range positions {
								target.normalBuffer = append(target.normalBuffer, pos[0], pos[1], pos[2], 0)
							}
						case "TANGENT":
							var posBuffer [][3]float32
							positions, err := modeler.ReadPosition(doc, doc.Accessors[accessor], posBuffer)
							if err != nil {
								return nil, err
							}
							target.tangentBuffer = make([]float32, 0, 4*int(primitive.numVertices))
							for _, pos := range positions {
								target.tangentBuffer = append(target.tangentBuffer, pos[0], pos[1], pos[2], 0)
							}
						case "TEXCOORD_0":
							var uvBuffer [][2]float32
							texCoords, err := modeler.ReadTextureCoord(doc, doc.Accessors[accessor], uvBuffer)
							if err != nil {
								return nil, err
							}
							target.uvBuffer = make([]float32, 0, 4*int(primitive.numVertices))
							for _, uv := range texCoords {
								target.uvBuffer = append(target.uvBuffer, uv[0], uv[1], 0, 0)
							}
						case "COLOR_0":
							target.colorBuffer = make([]float32, 0, 4*int(primitive.numVertices))
							switch doc.Accessors[accessor].ComponentType {
							case gltf.ComponentUbyte:
								if doc.Accessors[accessor].Type == gltf.AccessorVec3 {
									var vecBuffer [][3]uint8
									vecs, err := modeler.ReadAccessor(doc, doc.Accessors[accessor], vecBuffer)
									if err != nil {
										return nil, err
									}
									for _, vec := range vecs.([][3]uint8) {
										target.colorBuffer = append(target.colorBuffer, float32(vec[0])/255, float32(vec[1])/255, float32(vec[2])/255, 1)
									}
								} else {
									var vecBuffer [][4]uint8
									vecs, err := modeler.ReadAccessor(doc, doc.Accessors[accessor], vecBuffer)
									if err != nil {
										return nil, err
									}
									for _, vec := range vecs.([][4]uint8) {
										target.colorBuffer = append(target.colorBuffer, float32(vec[0])/255, float32(vec[1])/255, float32(vec[2])/255, float32(vec[3])/255)
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
										target.colorBuffer = append(target.colorBuffer, float32(vec[0])/65535, float32(vec[1])/65535, float32(vec[2])/65535, 1)
									}
								} else {
									var vecBuffer [][4]uint16
									vecs, err := modeler.ReadAccessor(doc, doc.Accessors[accessor], vecBuffer)
									if err != nil {
										return nil, err
									}
									for _, vec := range vecs.([][4]uint16) {
										target.colorBuffer = append(target.colorBuffer, float32(vec[0])/65535, float32(vec[1])/65535, float32(vec[2])/65535, float32(vec[3])/65535)
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
										target.colorBuffer = append(target.colorBuffer, vec[0], vec[1], vec[2], 1)
									}
								} else {
									var vecBuffer [][4]float32
									vecs, err := modeler.ReadAccessor(doc, doc.Accessors[accessor], vecBuffer)
									if err != nil {
										return nil, err
									}
									for _, vec := range vecs.([][4]float32) {
										target.colorBuffer = append(target.colorBuffer, vec[0], vec[1], vec[2], vec[3])
									}
								}
							}
						}
					}
					primitive.morphTargets = append(primitive.morphTargets, target)
				}
				primitive.morphTargetTexture = &GLTFTexture{}
				sys.mainThreadTask <- func() {
					dimension := int(math.Ceil(math.Pow(float64(8*primitive.numVertices), 0.5)))
					primitive.morphTargetTexture.tex = gfx.newDataTexture(int32(dimension), int32(dimension))
					//primitive.morphTargetTexture.tex.SetPixelData(targetBuffer)
				}
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
	var lightNodes []int32
	for idx, n := range doc.Nodes {
		var node = &Node{}
		mdl.nodes = append(mdl.nodes, node)
		node.visible = true
		node.rotation.restAt(n.Rotation)
		node.transition.restAt(n.Translation)
		node.scale.restAt(n.Scale)
		node.skin = n.Skin
		node.childrenIndex = n.Children
		node.morphTargetWeights = GLTFAnimatableProperty{isAnimated: false, restValue: n.Weights}
		if n.Mesh != nil {
			node.meshIndex = new(uint32)
			*node.meshIndex = *n.Mesh
			if len(n.Weights) == 0 {
				m := mdl.meshes[*n.Mesh]
				node.morphTargetWeights.restAt(m.morphTargetWeights.getValue())
			}
		}
		node.trans = TransNone
		node.castShadow = true
		node.zTest = true
		node.zWrite = true
		node.meshOutline = GLTFAnimatableProperty{isAnimated: false, restValue: float32(0.0)}
		if n.Extensions != nil {
			if l, ok := n.Extensions["KHR_lights_punctual"]; ok {
				var ext interface{}
				err := json.Unmarshal(l.(json.RawMessage), &ext)
				if err != nil {
					return nil, err
				}
				lightNodes = append(lightNodes, int32(idx))
				node.lightIndex = new(uint32)
				*node.lightIndex = (uint32)(ext.(map[string]interface{})["light"].(float64))
				node.shadowMapNear = GLTFAnimatableProperty{isAnimated: false, restValue: float32(0.0)}
				node.shadowMapFar = GLTFAnimatableProperty{isAnimated: false, restValue: float32(0.0)}
				node.shadowMapBottom = GLTFAnimatableProperty{isAnimated: false, restValue: float32(0.0)}
				node.shadowMapTop = GLTFAnimatableProperty{isAnimated: false, restValue: float32(0.0)}
				node.shadowMapLeft = GLTFAnimatableProperty{isAnimated: false, restValue: float32(0.0)}
				node.shadowMapRight = GLTFAnimatableProperty{isAnimated: false, restValue: float32(0.0)}
				node.shadowMapBias = GLTFAnimatableProperty{isAnimated: false, restValue: float32(0.0)}
			}
		}
		if n.Extras != nil {
			v, ok := n.Extras.(map[string]interface{})
			if ok {
				switch v["trans"] {
				case "ADD":
					node.trans = TransAdd
				case "SUB":
					node.trans = TransReverseSubtract
				case "MUL":
					node.trans = TransMul
				case "NONE":
					node.trans = TransNone
				}
				if v["disableZTest"] != nil && v["disableZTest"] != "0" && v["disableZTest"] != "false" {
					node.zTest = false
				}
				if v["disableZWrite"] != nil && v["disableZWrite"] != "0" && v["disableZWrite"] != "false" {
					node.zWrite = false
				}
				if v["castShadow"] != nil && (v["castShadow"] == "0" || v["castShadow"] == "false") {
					node.castShadow = false
				}
				if v["shadowMapNear"] != nil {
					node.shadowMapNear.restAt((float32)(v["shadowMapNear"].(float64)))
				}
				if v["shadowMapFar"] != nil {
					node.shadowMapFar.restAt((float32)(v["shadowMapFar"].(float64)))
				}
				if v["shadowMapBottom"] != nil {
					node.shadowMapBottom.restAt((float32)(v["shadowMapBottom"].(float64)))
				}
				if v["shadowMapTop"] != nil {
					node.shadowMapTop.restAt((float32)(v["shadowMapTop"].(float64)))
				}
				if v["shadowMapLeft"] != nil {
					node.shadowMapLeft.restAt((float32)(v["shadowMapLeft"].(float64)))
				}
				if v["shadowMapRight"] != nil {
					node.shadowMapRight.restAt((float32)(v["shadowMapRight"].(float64)))
				}
				if v["shadowMapBias"] != nil {
					node.shadowMapBias.restAt((float32)(v["shadowMapBias"].(float64)))
				}
				if v["id"] != nil {
					node.id = uint32(v["id"].(float64))
				}
				if v["layerNumber"] != nil {
					node.layerNumber = new(int)
					*node.layerNumber = int(v["layerNumber"].(float64))
				}
				if v["meshOutline"] != nil {
					node.meshOutline.restAt((float32)(v["meshOutline"].(float64)))
				}
			}
		}
		node.transformChanged = true
	}
	mdl.animationTimeStamps = map[uint32][]float32{}
	for _, a := range doc.Animations {
		anim := &GLTFAnimation{}
		mdl.animations = append(mdl.animations, anim)
		anim.duration = 0
		anim.name = a.Name
		anim.enabled = true
		anim.defaultEnabled = true
		anim.loopCount = -1
		anim.loop = 0
		for _, c := range a.Channels {
			channel := &GLTFAnimationChannel{}
			channel.nodeIndex = c.Target.Node
			channel.samplerIndex = *c.Sampler
			channel.target = nil
			if c.Target.Extensions != nil {
				if p, ok := c.Target.Extensions["KHR_animation_pointer"]; ok {
					var ext interface{}
					if err := json.Unmarshal(p.(json.RawMessage), &ext); err != nil {
						return nil, err
					}
					pointer := ext.(map[string]interface{})["pointer"].(string)
					if err := channel.parseAnimationPointer(mdl, pointer); err != nil {
						return nil, err
					}
				}
			}
			if channel.target == nil {
				switch c.Target.Path {
				case gltf.TRSTranslation:
					channel.targetType = TRSTranslation
					channel.target = &mdl.nodes[*channel.nodeIndex].transition
				case gltf.TRSScale:
					channel.targetType = TRSScale
					channel.target = &mdl.nodes[*channel.nodeIndex].scale
				case gltf.TRSRotation:
					channel.targetType = TRSRotation
					channel.target = &mdl.nodes[*channel.nodeIndex].rotation
				case gltf.TRSWeights:
					channel.targetType = MorphTargetWeight
					channel.target = &mdl.nodes[*channel.nodeIndex].morphTargetWeights
				default:

					continue
				}
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
		if a.Extras != nil {
			v, ok := a.Extras.(map[string]interface{})
			if ok {
				if v["id"] != nil {
					anim.id = uint32(v["id"].(float64))
				}
				if v["loopCount"] != nil {
					anim.loopCount = int32(v["loopCount"].(float64))
				}
				if v["enabled"] != nil {
					anim.enabled = v["enabled"] != "0" && v["enabled"] != "false"
					anim.defaultEnabled = anim.enabled
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
			skin.texture.tex = gfx.newDataTexture(6, int32(len(skin.joints)))
		}

		mdl.skins = append(mdl.skins, skin)
	}

	for _, s := range doc.Scenes {
		var scene = &Scene{}
		scene.name = s.Name
		scene.nodes = s.Nodes
		for _, n := range s.Nodes {
			scene.getSceneLight(n, mdl.nodes)
		}
		mdl.scenes = append(mdl.scenes, scene)
	}
	return mdl, nil
}

func (s *Scene) getSceneLight(n uint32, nodes []*Node) {
	node := nodes[n]
	for _, c := range node.childrenIndex {
		s.getSceneLight(c, nodes)
	}
	if node.lightIndex != nil {
		s.lightNodes = append(s.lightNodes, n)
	}
}

func (n *Node) getLocalTransform() (mat mgl.Mat4) {
	mat = mgl.Ident4()
	if n.transformChanged {
		t := n.transition.getValue().([3]float32)
		mat = mgl.Translate3D(t[0], t[1], t[2])
		r := n.rotation.getValue().([4]float32)
		mat = mat.Mul4(mgl.Quat{W: r[3], V: mgl.Vec3{r[0], r[1], r[2]}}.Mat4())
		s := n.scale.getValue().([3]float32)
		mat = mat.Mul4(mgl.Scale3D(s[0], s[1], s[2]))
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
	if n.meshIndex != nil {
		n.normalMatrix = n.worldTransform.Inv().Transpose()
	}
	if n.lightIndex != nil {
		scale := [3]float32{n.worldTransform.Col(0).Len(), n.worldTransform.Col(1).Len(), n.worldTransform.Col(2).Len()}
		mat := mgl.Ident4()
		for i := 0; i < 3; i++ {
			mat[i] = n.worldTransform[i] / scale[0]
			mat[i+4] = n.worldTransform[i+4] / scale[1]
			mat[i+8] = n.worldTransform[i+8] / scale[2]
		}
		quat := mgl.Mat4ToQuat(mat).Normalize()
		direction := mgl.Vec3{0, 0, -1}
		n.lightDirection = quat.Rotate(direction)
	}
	for _, index := range n.childrenIndex {
		(*nodes[index]).calculateWorldTransform(n.worldTransform, nodes)
	}
	return
}

func (mdl *Model) calculateTextureTransform() {
	for _, m := range mdl.materials {
		if index := m.textureIndex; index != nil {
			t := m.textureOffset.getValue().([2]float32)
			mat := mgl.Translate2D(t[0], t[1])
			r := m.textureRotation.getValue().(float32)
			mat = mat.Mul3(mgl.HomogRotate2D(r).Transpose())
			s := m.textureScale.getValue().([2]float32)
			mat = mat.Mul3(mgl.Scale2D(s[0], s[1]))
			m.textureTransform = mat
		}
		if index := m.normalMapIndex; index != nil {
			t := m.normalMapOffset.getValue().([2]float32)
			mat := mgl.Translate2D(t[0], t[1])
			r := m.normalMapRotation.getValue().(float32)
			mat = mat.Mul3(mgl.HomogRotate2D(r).Transpose())
			s := m.normalMapScale.getValue().([2]float32)
			mat = mat.Mul3(mgl.Scale2D(s[0], s[1]))
			m.normalMapTransform = mat
		}
		if index := m.ambientOcclusionMapIndex; index != nil {
			t := m.ambientOcclusionMapOffset.getValue().([2]float32)
			mat := mgl.Translate2D(t[0], t[1])
			r := m.ambientOcclusionMapRotation.getValue().(float32)
			mat = mat.Mul3(mgl.HomogRotate2D(r).Transpose())
			s := m.ambientOcclusionMapScale.getValue().([2]float32)
			mat = mat.Mul3(mgl.Scale2D(s[0], s[1]))
			m.ambientOcclusionMapTransform = mat
		}
		if index := m.metallicRoughnessMapIndex; index != nil {
			t := m.metallicRoughnessMapOffset.getValue().([2]float32)
			mat := mgl.Translate2D(t[0], t[1])
			r := m.metallicRoughnessMapRotation.getValue().(float32)
			mat = mat.Mul3(mgl.HomogRotate2D(r).Transpose())
			s := m.metallicRoughnessMapScale.getValue().([2]float32)
			mat = mat.Mul3(mgl.Scale2D(s[0], s[1]))
			m.metallicRoughnessMapTransform = mat
		}
		if index := m.emissionMapIndex; index != nil {
			t := m.emissionMapOffset.getValue().([2]float32)
			mat := mgl.Translate2D(t[0], t[1])
			r := m.emissionMapRotation.getValue().(float32)
			mat = mat.Mul3(mgl.HomogRotate2D(r).Transpose())
			s := m.emissionMapScale.getValue().([2]float32)
			mat = mat.Mul3(mgl.Scale2D(s[0], s[1]))
			m.emissionMapTransform = mat
		}
	}
}

func calculateAnimationData(mdl *Model, n *Node) {
	for _, index := range n.childrenIndex {
		calculateAnimationData(mdl, mdl.nodes[index])
	}
	if n.meshIndex == nil {
		return
	}
	if n.skin != nil {
		mdl.skins[*n.skin].calculateSkinMatrices(n.worldTransform.Inv(), mdl.nodes)
	}
	m := mdl.meshes[*n.meshIndex]
	var morphTargetWeights []struct {
		index  uint32
		weight float32
	}
	weights := n.morphTargetWeights.getValue().([]float32)
	if len(weights) == 0 {
		weights = m.morphTargetWeights.getValue().([]float32)
	}
	activeMorphTargetChanged := false
	if len(weights) > 0 {
		activeMorphTargets := make([]uint32, 0, len(weights))
		morphTargetWeights = make([]struct {
			index  uint32
			weight float32
		}, 0, len(weights))
		for idx, w := range weights {
			if w != 0 {
				morphTargetWeights = append(morphTargetWeights, struct {
					index  uint32
					weight float32
				}{uint32(idx), w})
				activeMorphTargets = append(activeMorphTargets, uint32(idx))
			}
		}
		if len(activeMorphTargets) != len(n.activeMorphTargets) {
			activeMorphTargetChanged = true
			n.activeMorphTargets = activeMorphTargets
		} else {
			for i := range activeMorphTargets {
				if activeMorphTargets[i] != n.activeMorphTargets[i] {
					activeMorphTargetChanged = true
					n.activeMorphTargets = activeMorphTargets
				}
			}
		}
	}
	for _, p := range m.primitives {
		if p.materialIndex == nil {
			continue
		}
		if len(morphTargetWeights) > 0 && len(p.morphTargets) >= len(morphTargetWeights) {
			p.morphTargetWeight = [8]float32{0, 0, 0, 0, 0, 0, 0, 0}
			if activeMorphTargetChanged {
				width := p.morphTargetTexture.tex.GetWidth()
				targetBuffer := make([]float32, 4*width*width)
				count := 0
				offset := 0
				for _, t := range morphTargetWeights {
					morphTarget := p.morphTargets[t.index]
					if len(morphTarget.positionBuffer) > 0 {
						copy(targetBuffer[offset:offset+len(morphTarget.positionBuffer)], morphTarget.positionBuffer)
						offset += len(morphTarget.positionBuffer)
						p.morphTargetWeight[count] = t.weight
						count += 1
					}
				}
				p.morphTargetOffset[0] = float32(count)
				for _, t := range morphTargetWeights {
					morphTarget := p.morphTargets[t.index]
					if len(morphTarget.normalBuffer) > 0 {
						copy(targetBuffer[offset:offset+len(morphTarget.normalBuffer)], morphTarget.normalBuffer)
						offset += len(morphTarget.normalBuffer)
						p.morphTargetWeight[count] = t.weight
						count += 1
					}
				}
				p.morphTargetOffset[1] = float32(count)
				for _, t := range morphTargetWeights {
					morphTarget := p.morphTargets[t.index]
					if len(morphTarget.tangentBuffer) > 0 {
						copy(targetBuffer[offset:offset+len(morphTarget.tangentBuffer)], morphTarget.tangentBuffer)
						offset += len(morphTarget.tangentBuffer)
						p.morphTargetWeight[count] = t.weight
						count += 1
					}
				}
				p.morphTargetOffset[2] = float32(count)
				for _, t := range morphTargetWeights {
					morphTarget := p.morphTargets[t.index]
					if len(morphTarget.uvBuffer) > 0 {
						copy(targetBuffer[offset:offset+len(morphTarget.uvBuffer)], morphTarget.uvBuffer)
						offset += len(morphTarget.uvBuffer)
						p.morphTargetWeight[count] = t.weight
						count += 1
					}
				}
				p.morphTargetOffset[3] = float32(count)
				for _, t := range morphTargetWeights {
					morphTarget := p.morphTargets[t.index]
					if len(morphTarget.colorBuffer) > 0 {
						copy(targetBuffer[offset:offset+len(morphTarget.colorBuffer)], morphTarget.colorBuffer)
						offset += len(morphTarget.colorBuffer)
						targetBuffer = append(targetBuffer, morphTarget.colorBuffer...)
						p.morphTargetWeight[count] = t.weight
						count += 1
					}
				}
				p.morphTargetCount = uint32(count)
				if len(targetBuffer) > int(4*width*width) {
					targetBuffer = targetBuffer[:4*width*width]
				}
				p.morphTargetTexture.tex.SetPixelData(targetBuffer)
			} else {
				count := 0
				for _, t := range morphTargetWeights {
					morphTarget := p.morphTargets[t.index]
					if len(morphTarget.positionBuffer) > 0 {
						p.morphTargetWeight[count] = t.weight
						count += 1
					}
				}
				p.morphTargetOffset[0] = float32(count)
				for _, t := range morphTargetWeights {
					morphTarget := p.morphTargets[t.index]
					if len(morphTarget.normalBuffer) > 0 {
						p.morphTargetWeight[count] = t.weight
						count += 1
					}
				}
				p.morphTargetOffset[1] = float32(count)
				for _, t := range morphTargetWeights {
					morphTarget := p.morphTargets[t.index]
					if len(morphTarget.tangentBuffer) > 0 {
						p.morphTargetWeight[count] = t.weight
						count += 1
					}
				}
				p.morphTargetOffset[2] = float32(count)
				for _, t := range morphTargetWeights {
					morphTarget := p.morphTargets[t.index]
					if len(morphTarget.uvBuffer) > 0 {
						p.morphTargetWeight[count] = t.weight
						count += 1
					}
				}
				p.morphTargetOffset[3] = float32(count)
				for _, t := range morphTargetWeights {
					morphTarget := p.morphTargets[t.index]
					if len(morphTarget.colorBuffer) > 0 {
						p.morphTargetWeight[count] = t.weight
						count += 1
					}
				}

			}
		} else {
			p.morphTargetCount = 0
			p.morphTargetOffset = [4]float32{0, 0, 0, 0}
			p.morphTargetWeight = [8]float32{0, 0, 0, 0, 0, 0, 0, 0}
		}
	}
}

type Plane struct {
	Normal [3]float32
	D      float32 // Distance from the origin
}

// Extract planes from the Model-View-Projection (MVP) matrix
func ExtractFrustumPlanes(MVPMatrix mgl.Mat4) [6]Plane {
	var planes [6]Plane
	// Left plane
	planes[0] = Plane{
		Normal: [3]float32{MVPMatrix.At(3, 0) + MVPMatrix.At(0, 0), MVPMatrix.At(3, 1) + MVPMatrix.At(0, 1), MVPMatrix.At(3, 2) + MVPMatrix.At(0, 2)},
		D:      MVPMatrix.At(3, 3) + MVPMatrix.At(0, 3),
	}
	// Right plane
	planes[1] = Plane{
		Normal: [3]float32{MVPMatrix.At(3, 0) - MVPMatrix.At(0, 0), MVPMatrix.At(3, 1) - MVPMatrix.At(0, 1), MVPMatrix.At(3, 2) - MVPMatrix.At(0, 2)},
		D:      MVPMatrix.At(3, 3) - MVPMatrix.At(0, 3),
	}
	// Bottom plane
	planes[2] = Plane{
		Normal: [3]float32{MVPMatrix.At(3, 0) + MVPMatrix.At(1, 0), MVPMatrix.At(3, 1) + MVPMatrix.At(1, 1), MVPMatrix.At(3, 2) + MVPMatrix.At(1, 2)},
		D:      MVPMatrix.At(3, 3) + MVPMatrix.At(1, 3),
	}
	// Top plane
	planes[3] = Plane{
		Normal: [3]float32{MVPMatrix.At(3, 0) - MVPMatrix.At(1, 0), MVPMatrix.At(3, 1) - MVPMatrix.At(1, 1), MVPMatrix.At(3, 2) - MVPMatrix.At(1, 2)},
		D:      MVPMatrix.At(3, 3) - MVPMatrix.At(1, 3),
	}
	// Near plane
	planes[4] = Plane{
		Normal: [3]float32{MVPMatrix.At(3, 0) + MVPMatrix.At(2, 0), MVPMatrix.At(3, 1) + MVPMatrix.At(2, 1), MVPMatrix.At(3, 2) + MVPMatrix.At(2, 2)},
		D:      MVPMatrix.At(3, 3) + MVPMatrix.At(2, 3),
	}
	// Far plane
	planes[5] = Plane{
		Normal: [3]float32{MVPMatrix.At(3, 0) - MVPMatrix.At(2, 0), MVPMatrix.At(3, 1) - MVPMatrix.At(2, 1), MVPMatrix.At(3, 2) - MVPMatrix.At(2, 2)},
		D:      MVPMatrix.At(3, 3) - MVPMatrix.At(2, 3),
	}

	// Normalize the planes
	for i := 0; i < 6; i++ {
		length := float32(math.Sqrt(float64(planes[i].Normal[0]*planes[i].Normal[0] + planes[i].Normal[1]*planes[i].Normal[1] + planes[i].Normal[2]*planes[i].Normal[2])))
		planes[i].Normal[0] /= length
		planes[i].Normal[1] /= length
		planes[i].Normal[2] /= length
		planes[i].D /= length
	}

	return planes
}

func isCulled(MVPMatrix mgl.Mat4, box BoundingBox) bool {
	points := [8][3]float32{
		{box.min[0], box.min[1], box.min[2]},
		{box.min[0], box.min[1], box.max[2]},
		{box.min[0], box.max[1], box.min[2]},
		{box.min[0], box.max[1], box.max[2]},
		{box.max[0], box.min[1], box.min[2]},
		{box.max[0], box.min[1], box.max[2]},
		{box.max[0], box.max[1], box.min[2]},
		{box.max[0], box.max[1], box.max[2]},
	}

	for _, point := range points {
		clipSpace := MVPMatrix.Mul4x1(mgl.Vec4{point[0], point[1], point[2], 1})
		// Check if the point is within the normalized device coordinates
		if clipSpace[0] >= -clipSpace[3] && clipSpace[0] <= clipSpace[3] &&
			clipSpace[1] >= -clipSpace[3] && clipSpace[1] <= clipSpace[3] &&
			clipSpace[2] >= -clipSpace[3] && clipSpace[2] <= clipSpace[3] {
			return false // At least one point is within the frustum
		}
	}
	planes := ExtractFrustumPlanes(MVPMatrix)
	for _, plane := range planes {
		// Find the positive vertex
		var positive [3]float32
		if plane.Normal[0] >= 0 {
			positive[0] = box.max[0]
		} else {
			positive[0] = box.min[0]
		}
		if plane.Normal[1] >= 0 {
			positive[1] = box.max[1]
		} else {
			positive[1] = box.min[1]
		}
		if plane.Normal[2] >= 0 {
			positive[2] = box.max[2]
		} else {
			positive[2] = box.min[2]
		}
		// Check if the positive vertex is outside the plane
		if plane.Normal[0]*positive[0]+plane.Normal[1]*positive[1]+plane.Normal[2]*positive[2]+plane.D < 0 {
			return true // Entire bounding box is outside the frustum
		}
	}

	return false
}

func drawNode(mdl *Model, scene *Scene, layerNumber int, defaultLayerNumber int, n *Node, camOffset [3]float32, drawBlended bool, unlit bool, viewProjMatrix mgl.Mat4, outlineConst float32) {
	//mat := n.getLocalTransform()
	//model = model.Mul4(mat)
	for _, index := range n.childrenIndex {
		drawNode(mdl, scene, layerNumber, defaultLayerNumber, mdl.nodes[index], camOffset, drawBlended, unlit, viewProjMatrix, outlineConst)
	}
	nodeLayerNumber := defaultLayerNumber
	if n.layerNumber != nil {
		nodeLayerNumber = *n.layerNumber
	}
	if n.meshIndex == nil || !n.visible || (nodeLayerNumber != layerNumber) {
		return
	}

	negTrans := (n.trans == TransReverseSubtract)
	alpha := [2]int32{255, 255}
	if n.trans == TransNone {
		alpha = [2]int32{255, 0}
	}
	neg, grayscale, padd, pmul, invblend, hue := mdl.pfx.getFcPalFx(negTrans, alpha)

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
			if invblend >= 1 {
				padd[0] = -padd[0]
				padd[1] = -padd[1]
				padd[2] = -padd[2]
			}
		} else {
			src = BlendOne
			dst = BlendOne
		}
	case TransReverseSubtract:
		if invblend == 3 {
			src = BlendOne
			dst = BlendOne
			neg = false
			if invblend >= 1 {
				padd[0] = -padd[0]
				padd[1] = -padd[1]
				padd[2] = -padd[2]
			}
		} else {
			src = BlendOne
			dst = BlendOne
			blendEq = BlendReverseSubtract
		}
	case TransMul:
		if invblend == 3 {
			//Not accurate
			src = BlendOneMinusDstColor
			dst = BlendOne
			neg = false
			blendEq = BlendReverseSubtract
		} else {
			src = BlendDstColor
			dst = BlendOneMinusSrcAlpha
		}
	default:
		src = BlendOne
		dst = BlendOneMinusSrcAlpha
	}
	m := mdl.meshes[*n.meshIndex]
	reverseCull := n.worldTransform.Det() < 0
	for _, p := range m.primitives {
		if p.materialIndex == nil {
			continue
		}
		if n.skin == nil && p.morphTargetCount == 0 && isCulled(viewProjMatrix.Mul4(n.worldTransform), p.boundingBox) {
			continue
		}

		mat := mdl.materials[*p.materialIndex]
		if ((mat.alphaMode != AlphaModeBlend && n.trans == TransNone) && drawBlended) ||
			((mat.alphaMode == AlphaModeBlend || n.trans != TransNone) && !drawBlended) {
			return
		}
		color := mdl.materials[*p.materialIndex].baseColorFactor.getValue().([4]float32)
		meshOutline := n.meshOutline.getValue().(float32)
		gfx.SetModelPipeline(blendEq, src, dst, n.zTest, n.zWrite, mdl.materials[*p.materialIndex].doubleSided, reverseCull, p.useUV, p.useNormal, p.useTangent, p.useVertexColor, p.useJoint0, p.useJoint1, p.useOutlineAttribute, p.numVertices, p.vertexBufferOffset)

		gfx.SetModelUniformMatrix("model", n.worldTransform[:])
		gfx.SetModelUniformMatrix("normalMatrix", n.normalMatrix[:])
		gfx.SetModelUniformI("numVertices", int(p.numVertices))
		//gfx.SetModelUniformF("ambientOcclusion", 1)
		gfx.SetModelUniformF("metallicRoughness", mat.metallic.getValue().(float32), mat.roughness.getValue().(float32))
		gfx.SetModelUniformF("ambientOcclusionStrength", mat.ambientOcclusion.getValue().(float32))
		gfx.SetModelUniformMatrix3("texTransform", mat.textureTransform[:])
		gfx.SetModelUniformMatrix3("normalMapTransform", mat.normalMapTransform[:])
		gfx.SetModelUniformMatrix3("metallicRoughnessMapTransform", mat.metallicRoughnessMapTransform[:])
		gfx.SetModelUniformMatrix3("ambientOcclusionMapTransform", mat.ambientOcclusionMapTransform[:])
		gfx.SetModelUniformMatrix3("emissionMapTransform", mat.emissionMapTransform[:])

		gfx.SetModelUniformF("cameraPosition", -camOffset[0], -camOffset[1], -camOffset[2])

		if n.skin != nil {
			skin := mdl.skins[*n.skin]
			gfx.SetModelTexture("jointMatrices", skin.texture.tex)
		}

		if p.morphTargetCount > 0 {
			gfx.SetModelUniformF("morphTargetOffset", p.morphTargetOffset[0], p.morphTargetOffset[1], p.morphTargetOffset[2], p.morphTargetOffset[3])
			gfx.SetModelUniformI("numTargets", int(Min(int32(p.morphTargetCount), 8)))
			gfx.SetModelTexture("morphTargetValues", p.morphTargetTexture.tex)
			gfx.SetModelUniformFv("morphTargetWeight", p.morphTargetWeight[:])
			gfx.SetModelUniformI("morphTargetTextureDimension", int(p.morphTargetTexture.tex.GetWidth()))
		} else {
			gfx.SetModelUniformFv("morphTargetWeight", make([]float32, 8))
		}
		mode := p.mode
		if sys.wireframeDisplay {
			mode = 1 // Set mesh render mode to "lines"
		}
		gfx.SetModelUniformI("unlit", int(Btoi(unlit || mat.unlit)))
		gfx.SetModelUniformFv("add", padd[:])
		gfx.SetModelUniformFv("mult", []float32{pmul[0] * sys.brightness, pmul[1] * sys.brightness, pmul[2] * sys.brightness})
		gfx.SetModelUniformI("neg", int(Btoi(neg)))
		gfx.SetModelUniformF("hue", hue)
		gfx.SetModelUniformF("gray", grayscale)
		gfx.SetModelUniformI("enableAlpha", int(Btoi(mat.alphaMode == AlphaModeBlend)))
		gfx.SetModelUniformF("alphaThreshold", mat.alphaCutoff.getValue().(float32))
		gfx.SetModelUniformFv("baseColorFactor", color[:])
		if n.skin != nil {
			gfx.SetModelUniformI("numJoints", len(mdl.skins[*n.skin].joints))
		}
		if index := mat.textureIndex; index != nil {
			gfx.SetModelTexture("tex", mdl.textures[*index].tex)
			gfx.SetModelUniformI("useTexture", 1)
		} else {
			gfx.SetModelUniformI("useTexture", 0)
		}
		if index := mat.normalMapIndex; index != nil {
			gfx.SetModelTexture("normalMap", mdl.textures[*index].tex)
			gfx.SetModelUniformI("useNormalMap", 1)
		} else {
			gfx.SetModelUniformI("useNormalMap", 0)
		}
		if index := mat.metallicRoughnessMapIndex; index != nil {
			gfx.SetModelTexture("metallicRoughnessMap", mdl.textures[*index].tex)
			gfx.SetModelUniformI("useMetallicRoughnessMap", 1)
		} else {
			gfx.SetModelUniformI("useMetallicRoughnessMap", 0)
		}
		if index := mat.ambientOcclusionMapIndex; index != nil {
			gfx.SetModelTexture("ambientOcclusionMap", mdl.textures[*index].tex)
		}
		emission := mat.emission.getValue().([3]float32)
		gfx.SetModelUniformFv("emission", emission[:])
		if index := mat.emissionMapIndex; index != nil {
			gfx.SetModelTexture("emissionMap", mdl.textures[*index].tex)
			gfx.SetModelUniformI("useEmissionMap", 1)
		} else {
			gfx.SetModelUniformI("useEmissionMap", 0)
		}
		gfx.SetModelUniformF("meshOutline", 0)
		gfx.RenderElements(mode, int(p.numIndices), int(p.elementBufferOffset))
		if meshOutline > 0 {
			gfx.SetMeshOulinePipeline(!reverseCull, meshOutline*outlineConst)
			gfx.RenderElements(mode, int(p.numIndices), int(p.elementBufferOffset))
		}

	}
}

func drawNodeShadow(mdl *Model, scene *Scene, n *Node, camOffset [3]float32, drawBlended bool, lightIndex int, numLights int, viewProjMatrices []mgl.Mat4, lightTypes []LightType) {
	//mat := n.getLocalTransform()
	//model = model.Mul4(mat)
	for _, index := range n.childrenIndex {
		drawNodeShadow(mdl, scene, mdl.nodes[index], camOffset, drawBlended, lightIndex, numLights, viewProjMatrices, lightTypes)
	}
	if n.meshIndex == nil || !n.visible {
		return
	}

	if n.trans == TransAdd || n.trans == TransReverseSubtract || n.trans == TransMul || !n.zTest || !n.zWrite || !n.castShadow {
		return
	}
	m := mdl.meshes[*n.meshIndex]
	reverseCull := n.worldTransform.Det() < 0
	for _, p := range m.primitives {
		if p.materialIndex == nil {
			continue
		}
		mat := mdl.materials[*p.materialIndex]
		if ((mat.alphaMode != AlphaModeBlend && n.trans == TransNone) && drawBlended) ||
			((mat.alphaMode == AlphaModeBlend || n.trans != TransNone) && !drawBlended) {
			return
		}
		color := mdl.materials[*p.materialIndex].baseColorFactor.getValue().([4]float32)
		if color[3] == 0 && mat.alphaMode == AlphaModeBlend {
			return
		}
		gfx.setShadowMapPipeline(mdl.materials[*p.materialIndex].doubleSided, reverseCull, p.useUV, p.useNormal, p.useTangent, p.useVertexColor, p.useJoint0, p.useJoint1, p.numVertices, p.vertexBufferOffset)

		gfx.SetShadowMapUniformMatrix("model", n.worldTransform[:])
		gfx.SetShadowMapUniformI("numVertices", int(p.numVertices))
		if n.skin != nil {
			skin := mdl.skins[*n.skin]
			gfx.SetShadowMapTexture("jointMatrices", skin.texture.tex)
		}

		if p.morphTargetCount > 0 {
			gfx.SetShadowMapUniformF("morphTargetOffset", p.morphTargetOffset[0], p.morphTargetOffset[1], p.morphTargetOffset[2], p.morphTargetOffset[3])
			gfx.SetShadowMapUniformI("numTargets", int(Min(int32(p.morphTargetCount), 8)))
			gfx.SetShadowMapTexture("morphTargetValues", p.morphTargetTexture.tex)
			gfx.SetShadowMapUniformFv("morphTargetWeight", p.morphTargetWeight[:])
			gfx.SetShadowMapUniformI("morphTargetTextureDimension", int(p.morphTargetTexture.tex.GetWidth()))
		} else {
			gfx.SetShadowMapUniformFv("morphTargetOffset", make([]float32, 4))
			gfx.SetShadowMapUniformI("numTargets", 0)
			gfx.SetShadowMapUniformFv("morphTargetWeight", make([]float32, 8))
		}
		mode := p.mode
		gfx.SetShadowMapUniformI("enableAlpha", int(Btoi(mat.alphaMode == AlphaModeBlend)))
		gfx.SetShadowMapUniformF("alphaThreshold", mat.alphaCutoff.getValue().(float32))
		gfx.SetShadowMapUniformFv("baseColorFactor", color[:])
		if n.skin != nil {
			gfx.SetShadowMapUniformI("numJoints", len(mdl.skins[*n.skin].joints))
		}
		if index := mat.textureIndex; index != nil {
			gfx.SetShadowMapTexture("tex", mdl.textures[*index].tex)
			gfx.SetShadowMapUniformI("useTexture", 1)
		} else {
			gfx.SetShadowMapUniformI("useTexture", 0)
		}
		gfx.SetShadowMapUniformMatrix3("texTransform", mat.textureTransform[:])
		for i := 0; i < numLights; i++ {
			culled := true
			if n.skin == nil && p.morphTargetCount == 0 {
				if lightTypes[i] == PointLight {
					for j := 0; j < 6; j++ {
						if !isCulled(viewProjMatrices[i*6+j].Mul4(n.worldTransform), p.boundingBox) {
							culled = false
							break
						}
					}
				} else {
					if isCulled(viewProjMatrices[i*6].Mul4(n.worldTransform), p.boundingBox) {
						continue
					} else {
						culled = false
					}
				}
			}
			if culled == false {
				gfx.SetShadowMapUniformI("layerOffset", i*6)
				gfx.SetShadowMapUniformI("lightIndex", i+lightIndex)
				gfx.RenderShadowMapElements(mode, int(p.numIndices), int(p.elementBufferOffset))
			}
		}
	}
}
func (model *Model) drawShadow(bufferIndex uint32, sceneNumber int, offset [3]float32) {
	scene := model.scenes[sceneNumber]
	gfx.prepareShadowMapPipeline(bufferIndex)
	numLights := 0
	var lightMatrices [32]mgl.Mat4
	var lightTypes [4]LightType
	for i := 0; i < 4; i++ {
		if i >= len(scene.lightNodes) {
			gfx.SetShadowMapUniformI("lightType["+strconv.Itoa(i)+"]", 0)
			continue
		}
		numLights += 1
		lightNode := model.nodes[scene.lightNodes[i]]
		light := model.lights[*lightNode.lightIndex]
		shadowMapNear := float32(0.1)
		if light.lightType == DirectionalLight {
			shadowMapNear = -20
		}
		shadowMapFar := float32(50)
		shadowMapBottom := float32(-20)
		shadowMapTop := float32(20)
		shadowMapLeft := float32(-20)
		shadowMapRight := float32(20)

		if light.lightType == DirectionalLight {
			shadowMapNear = -20
		}
		if v := light.shadowMapNear.getValue().(float32); v != 0 {
			shadowMapNear = v
		}
		if v := light.shadowMapFar.getValue().(float32); v != 0 {
			shadowMapFar = v
		}
		if v := light.shadowMapBottom.getValue().(float32); v != 0 {
			shadowMapBottom = v
		}
		if v := light.shadowMapTop.getValue().(float32); v != 0 {
			shadowMapTop = v
		}
		if v := light.shadowMapLeft.getValue().(float32); v != 0 {
			shadowMapLeft = v
		}
		if v := light.shadowMapRight.getValue().(float32); v != 0 {
			shadowMapRight = v
		}
		if v := lightNode.shadowMapNear.getValue().(float32); v != 0 {
			shadowMapNear = v
		}
		if v := lightNode.shadowMapFar.getValue().(float32); v != 0 {
			shadowMapFar = v
		}
		if v := lightNode.shadowMapBottom.getValue().(float32); v != 0 {
			shadowMapBottom = v
		}
		if v := lightNode.shadowMapTop.getValue().(float32); v != 0 {
			shadowMapTop = v
		}
		if v := lightNode.shadowMapLeft.getValue().(float32); v != 0 {
			shadowMapLeft = v
		}
		if v := lightNode.shadowMapRight.getValue().(float32); v != 0 {
			shadowMapRight = v
		}

		lightProj := gfx.PerspectiveProjectionMatrix(mgl.DegToRad(90), 1, shadowMapNear, shadowMapFar)
		if light.lightType == DirectionalLight {
			lightProj = gfx.OrthographicProjectionMatrix(shadowMapLeft, shadowMapRight, shadowMapBottom, shadowMapTop, shadowMapNear, shadowMapFar)
		} else if light.lightType == SpotLight {
			lightProj = gfx.PerspectiveProjectionMatrix(mgl.DegToRad(90), 1, shadowMapNear, shadowMapFar)
		}
		lightTypes[i] = light.lightType
		if light.lightType == PointLight {
			//var lightMatrices [6]mgl.Mat4
			lightMatrices[i*6] = lightProj.Mul4(mgl.LookAtV([3]float32{lightNode.worldTransform[12], lightNode.worldTransform[13], lightNode.worldTransform[14]}, [3]float32{lightNode.worldTransform[12] + 1, lightNode.worldTransform[13], lightNode.worldTransform[14]}, [3]float32{0, -1, 0}))
			lightMatrices[i*6+1] = lightProj.Mul4(mgl.LookAtV([3]float32{lightNode.worldTransform[12], lightNode.worldTransform[13], lightNode.worldTransform[14]}, [3]float32{lightNode.worldTransform[12] - 1, lightNode.worldTransform[13], lightNode.worldTransform[14]}, [3]float32{0, -1, 0}))
			lightMatrices[i*6+2] = lightProj.Mul4(mgl.LookAtV([3]float32{lightNode.worldTransform[12], lightNode.worldTransform[13], lightNode.worldTransform[14]}, [3]float32{lightNode.worldTransform[12], lightNode.worldTransform[13] + 1, lightNode.worldTransform[14]}, [3]float32{0, 0, 1}))
			lightMatrices[i*6+3] = lightProj.Mul4(mgl.LookAtV([3]float32{lightNode.worldTransform[12], lightNode.worldTransform[13], lightNode.worldTransform[14]}, [3]float32{lightNode.worldTransform[12], lightNode.worldTransform[13] - 1, lightNode.worldTransform[14]}, [3]float32{0, 0, -1}))
			lightMatrices[i*6+4] = lightProj.Mul4(mgl.LookAtV([3]float32{lightNode.worldTransform[12], lightNode.worldTransform[13], lightNode.worldTransform[14]}, [3]float32{lightNode.worldTransform[12], lightNode.worldTransform[13], lightNode.worldTransform[14] + 1}, [3]float32{0, -1, 0}))
			lightMatrices[i*6+5] = lightProj.Mul4(mgl.LookAtV([3]float32{lightNode.worldTransform[12], lightNode.worldTransform[13], lightNode.worldTransform[14]}, [3]float32{lightNode.worldTransform[12], lightNode.worldTransform[13], lightNode.worldTransform[14] - 1}, [3]float32{0, -1, 0}))
			for j := 0; j < 6; j++ {
				gfx.SetShadowMapUniformMatrix("lightMatrices["+strconv.Itoa(i*6+j)+"]", lightMatrices[i*6+j][:])
			}
			gfx.SetShadowMapUniformI("lights["+strconv.Itoa(i)+"].type", 2)

		} else {
			lightView := mgl.LookAtV([3]float32{lightNode.localTransform[12], lightNode.localTransform[13], lightNode.localTransform[14]}, [3]float32{lightNode.localTransform[12] + lightNode.lightDirection[0], lightNode.localTransform[13] + lightNode.lightDirection[1], lightNode.localTransform[14] + lightNode.lightDirection[2]}, [3]float32{0, 1, 0})
			lightMatrices[i*6] = lightProj.Mul4(lightView)
			gfx.SetShadowMapUniformMatrix("lightMatrices["+strconv.Itoa(i*6)+"]", lightMatrices[i*6][:])
			if light.lightType == DirectionalLight {
				gfx.SetShadowMapUniformI("lights["+strconv.Itoa(i)+"].type", 1)
			} else {
				gfx.SetShadowMapUniformI("lights["+strconv.Itoa(i)+"].type", 3)
			}
		}
		gfx.SetShadowMapUniformF("lights["+strconv.Itoa(i)+"].shadowMapFar", shadowMapFar)
		gfx.SetShadowMapUniformF("lights["+strconv.Itoa(i)+"].position", lightNode.worldTransform[12], lightNode.worldTransform[13], lightNode.worldTransform[14])
		if gfx.GetName() == "OpenGL 2.1" {
			if light.lightType == PointLight {
				gfx.SetShadowFrameCubeTexture(uint32(i))
			} else {
				gfx.SetShadowFrameTexture(uint32(i))
			}
			for _, index := range scene.nodes {
				drawNodeShadow(model, scene, model.nodes[index], offset, false, i, 1, lightMatrices[:], lightTypes[:])
			}
			for _, index := range scene.nodes {
				drawNodeShadow(model, scene, model.nodes[index], offset, true, i, 1, lightMatrices[:], lightTypes[:])
			}
			if len(model.scenes) > 1 {
				for _, index := range scene.nodes {
					drawNodeShadow(model, model.scenes[1], model.nodes[index], offset, false, i, 1, lightMatrices[:], lightTypes[:])
				}
				for _, index := range scene.nodes {
					drawNodeShadow(model, model.scenes[1], model.nodes[index], offset, true, i, 1, lightMatrices[:], lightTypes[:])
				}
			}
		}
	}
	if gfx.GetName() != "OpenGL 2.1" {
		for _, index := range scene.nodes {
			drawNodeShadow(model, scene, model.nodes[index], offset, false, 0, numLights, lightMatrices[:], lightTypes[:])
		}
		for _, index := range scene.nodes {
			drawNodeShadow(model, scene, model.nodes[index], offset, true, 0, numLights, lightMatrices[:], lightTypes[:])
		}
		if len(model.scenes) > 1 {
			for _, index := range scene.nodes {
				drawNodeShadow(model, model.scenes[1], model.nodes[index], offset, false, 0, numLights, lightMatrices[:], lightTypes[:])
			}
			for _, index := range scene.nodes {
				drawNodeShadow(model, model.scenes[1], model.nodes[index], offset, true, 0, numLights, lightMatrices[:], lightTypes[:])
			}
		}
	}
	gfx.ReleaseShadowPipeline()
}
func (model *Model) draw(bufferIndex uint32, sceneNumber int, layerNumber int, defaultLayerNumber int, offset [3]float32, proj, view, viewProjMatrix mgl.Mat4, outlineConst float32) {
	if sceneNumber < 0 || sceneNumber >= len(model.scenes) {
		return
	}
	scene := model.scenes[sceneNumber]
	for _, index := range scene.nodes {
		model.nodes[index].calculateWorldTransform(mgl.Ident4(), model.nodes)
		calculateAnimationData(model, model.nodes[index])
	}
	if len(scene.lightNodes) > 0 && layerNumber == -1 && gfx.IsShadowEnabled() {
		// Do it in another thread if possible
		if gfx.NewWorkerThread() {
			go model.drawShadow(bufferIndex, sceneNumber, offset)
		} else {
			model.drawShadow(bufferIndex, sceneNumber, offset)
		}

	}
	if model.environment != nil {
		gfx.prepareModelPipeline(bufferIndex, model.environment)
	} else {
		gfx.prepareModelPipeline(bufferIndex, nil)
	}
	gfx.SetModelUniformMatrix("projection", proj[:])
	gfx.SetModelUniformMatrix("view", view[:])

	unlit := false
	for idx := 0; idx < 4; idx++ {
		gfx.SetModelUniformF("lights["+strconv.Itoa(idx)+"].color", 0, 0, 0)
	}
	if len(scene.lightNodes) > 0 {
		for idx := 0; idx < MinI(len(scene.lightNodes), 4); idx++ {
			lightNode := model.nodes[scene.lightNodes[idx]]
			light := model.lights[*lightNode.lightIndex]
			shadowMapNear := float32(0.1)
			shadowMapFar := float32(50)
			shadowMapBottom := float32(-20)
			shadowMapTop := float32(20)
			shadowMapLeft := float32(-20)
			shadowMapRight := float32(20)
			shadowMapBias := float32(0.02)

			if light.lightType == DirectionalLight {
				shadowMapNear = -20
			}
			if v := light.shadowMapNear.getValue().(float32); v != 0 {
				shadowMapNear = v
			}
			if v := light.shadowMapFar.getValue().(float32); v != 0 {
				shadowMapFar = v
			}
			if v := light.shadowMapBottom.getValue().(float32); v != 0 {
				shadowMapBottom = v
			}
			if v := light.shadowMapTop.getValue().(float32); v != 0 {
				shadowMapTop = v
			}
			if v := light.shadowMapLeft.getValue().(float32); v != 0 {
				shadowMapLeft = v
			}
			if v := light.shadowMapRight.getValue().(float32); v != 0 {
				shadowMapRight = v
			}
			if v := light.shadowMapBias.getValue().(float32); v != 0 {
				shadowMapBias = v
			}
			if v := lightNode.shadowMapNear.getValue().(float32); v != 0 {
				shadowMapNear = v
			}
			if v := lightNode.shadowMapFar.getValue().(float32); v != 0 {
				shadowMapFar = v
			}
			if v := lightNode.shadowMapBottom.getValue().(float32); v != 0 {
				shadowMapBottom = v
			}
			if v := lightNode.shadowMapTop.getValue().(float32); v != 0 {
				shadowMapTop = v
			}
			if v := lightNode.shadowMapLeft.getValue().(float32); v != 0 {
				shadowMapLeft = v
			}
			if v := lightNode.shadowMapRight.getValue().(float32); v != 0 {
				shadowMapRight = v
			}
			if v := lightNode.shadowMapBias.getValue().(float32); v != 0 {
				shadowMapBias = v
			}
			gfx.SetModelUniformI("lights["+strconv.Itoa(idx)+"].type", int(light.lightType))
			gfx.SetModelUniformF("lights["+strconv.Itoa(idx)+"].intensity", light.intensity.getValue().(float32))
			gfx.SetModelUniformF("lights["+strconv.Itoa(idx)+"].innerConeCos", float32(math.Cos(float64(light.innerConeAngle.getValue().(float32)))))
			gfx.SetModelUniformF("lights["+strconv.Itoa(idx)+"].outerConeCos", float32(math.Cos(float64(light.outerConeAngle.getValue().(float32)))))
			gfx.SetModelUniformF("lights["+strconv.Itoa(idx)+"].range", light.lightRange.getValue().(float32))
			c := light.color.getValue().([3]float32)
			gfx.SetModelUniformF("lights["+strconv.Itoa(idx)+"].color", c[0], c[1], c[2])
			gfx.SetModelUniformF("lights["+strconv.Itoa(idx)+"].position", lightNode.worldTransform[12], lightNode.worldTransform[13], lightNode.worldTransform[14])
			gfx.SetModelUniformF("lights["+strconv.Itoa(idx)+"].shadowMapFar", shadowMapFar)
			gfx.SetModelUniformF("lights["+strconv.Itoa(idx)+"].shadowBias", shadowMapBias)
			if light.lightType != PointLight {
				gfx.SetModelUniformF("lights["+strconv.Itoa(idx)+"].direction", lightNode.lightDirection[0], lightNode.lightDirection[1], lightNode.lightDirection[2])
			}
			if light.lightType == DirectionalLight {
				lightProj := mgl.Ortho(shadowMapLeft, shadowMapRight, shadowMapBottom, shadowMapTop, shadowMapNear, shadowMapFar)
				lightView := mgl.LookAtV([3]float32{lightNode.localTransform[12], lightNode.localTransform[13], lightNode.localTransform[14]}, [3]float32{lightNode.localTransform[12] + lightNode.lightDirection[0], lightNode.localTransform[13] + lightNode.lightDirection[1], lightNode.localTransform[14] + lightNode.lightDirection[2]}, [3]float32{0, 1, 0})
				lightMatrix := lightProj.Mul4(lightView)
				gfx.SetModelUniformMatrix("lightMatrices["+strconv.Itoa(idx)+"]", lightMatrix[:])
			} else if light.lightType == SpotLight {
				lightProj := gfx.PerspectiveProjectionMatrix(mgl.DegToRad(90), 1, shadowMapNear, shadowMapFar)
				lightView := mgl.LookAtV([3]float32{lightNode.localTransform[12], lightNode.localTransform[13], lightNode.localTransform[14]}, [3]float32{lightNode.localTransform[12] + lightNode.lightDirection[0], lightNode.localTransform[13] + lightNode.lightDirection[1], lightNode.localTransform[14] + lightNode.lightDirection[2]}, [3]float32{0, 1, 0})
				lightMatrix := lightProj.Mul4(lightView)
				gfx.SetModelUniformMatrix("lightMatrices["+strconv.Itoa(idx)+"]", lightMatrix[:])
			} else {
				ident := mgl.Ident4()
				gfx.SetModelUniformMatrix("lightMatrices["+strconv.Itoa(idx)+"]", ident[:])
			}
		}
	} else if model.environment == nil {
		unlit = true
	}
	for _, index := range scene.nodes {
		drawNode(model, scene, layerNumber, defaultLayerNumber, model.nodes[index], offset, false, unlit, viewProjMatrix, outlineConst)
	}
	for _, index := range scene.nodes {
		drawNode(model, scene, layerNumber, defaultLayerNumber, model.nodes[index], offset, true, unlit, viewProjMatrix, outlineConst)
	}
	gfx.ReleaseModelPipeline()
}

func (s *Stage) drawModel(pos [2]float32, yofs float32, scl float32, layerNumber int32) {
	if s.model == nil || !gfx.IsModelEnabled() {
		return
	}
	drawFOV := s.stageCamera.fov * math.Pi / 180
	outlineConst := float32(0.003 * math.Tan(float64(drawFOV)))
	var syo float32
	scaleCorrection := float32(sys.cam.localcoord[1]) * sys.cam.localscl / float32(sys.gameHeight)
	posMul := float32(math.Tan(float64(drawFOV)/2)) * -s.model.offset[2] / (float32(sys.cam.localcoord[1]) / 2)
	aspectCorrection := (float32(sys.cam.zoffset)/float32(sys.cam.localcoord[1]) - (float32(sys.cam.zoffset)*s.localscl-sys.cam.aspectcorrection)/float32(sys.gameHeight)) * 2
	syo = -(float32(s.stageCamera.zoffset) - float32(sys.cam.localcoord[1])/2) * (1 - scl) / scl
	syo2 := -(float32(s.stageCamera.zoffset) - float32(sys.cam.localcoord[1])/2) * (1 - scaleCorrection) / float32(sys.cam.localcoord[1]) * 2
	offset := [3]float32{(pos[0]*-posMul + s.model.offset[0]/scl), (((pos[1])/scl+syo)*posMul + s.model.offset[1]), s.model.offset[2] / scl}
	rotation := [3]float32{s.model.rotation[0], s.model.rotation[1], s.model.rotation[2]}
	scale := [3]float32{s.model.scale[0], s.model.scale[1], s.model.scale[2]}
	proj := mgl.Translate3D(0, (sys.cam.zoomanchorcorrection+yofs)/float32(sys.gameHeight)*2+syo2+aspectCorrection, 0)

	// Apply aspect ratio scaling
	// TODO: In the letterbox case the model renders too low
	scaleX := scaleCorrection
	scaleY := scaleCorrection
	if sys.cfg.Video.FightAspectWidth != 0 && sys.cfg.Video.FightAspectHeight != 0 {
		aspectGame := sys.getCurrentAspect()
		aspectWindow := float32(sys.scrrect[2]) / float32(sys.scrrect[3])

		if aspectWindow > aspectGame {
			// Pillarbox - window is wider than game
			scaleX *= aspectWindow / aspectGame
		} else if aspectWindow < aspectGame {
			// Letterbox - window is taller than game
			scaleY *= aspectGame / aspectWindow
		}
	}

	proj = proj.Mul4(mgl.Scale3D(scaleX, scaleY, 1))
	proj = proj.Mul4(mgl.Translate3D(0, (sys.cam.yshift * scl), 0))
	proj = proj.Mul4(gfx.PerspectiveProjectionMatrix(drawFOV, float32(sys.scrrect[2])/float32(sys.scrrect[3]), s.stageCamera.near, s.stageCamera.far))
	view := mgl.Ident4()
	view = view.Mul4(mgl.Translate3D(offset[0], offset[1], offset[2]))
	view = view.Mul4(mgl.HomogRotate3DX(rotation[0]))
	view = view.Mul4(mgl.HomogRotate3DY(rotation[1]))
	view = view.Mul4(mgl.HomogRotate3DZ(rotation[2]))
	view = view.Mul4(mgl.Scale3D(scale[0], scale[1], scale[2]))

	if layerNumber == -1 {
		s.model.calculateTextureTransform()
		s.model.draw(0, 0, int(layerNumber), 0, [3]float32{offset[0] / scale[0], offset[1] / scale[1], offset[2] / scale[2]}, proj, view, proj.Mul4(view), outlineConst)
	} else if layerNumber == 0 {
		s.model.draw(0, 0, int(layerNumber), 0, [3]float32{offset[0] / scale[0], offset[1] / scale[1], offset[2] / scale[2]}, proj, view, proj.Mul4(view), outlineConst)
	} else if layerNumber == 1 {
		s.model.draw(0, 0, int(layerNumber), 0, [3]float32{offset[0] / scale[0], offset[1] / scale[1], offset[2] / scale[2]}, proj, view, proj.Mul4(view), outlineConst)
		s.model.draw(0, 1, int(layerNumber), 1, [3]float32{offset[0] / scale[0], offset[1] / scale[1], offset[2] / scale[2]}, proj, view, proj.Mul4(view), outlineConst)
	}
}

func (channel *GLTFAnimationChannel) parseAnimationPointer(m *Model, pointer string) error {
	channel.nodeIndex = nil
	components := strings.Split(pointer, "/")
	switch components[1] {
	case "materials":
		index, _ := strconv.Atoi(components[2])
		material := m.materials[index]
		switch components[3] {
		case "alphaCutoff":
			channel.target = &material.alphaCutoff
			channel.targetType = AnimFloat
		case "emissiveFactor":
			channel.target = &material.emission
			channel.targetType = AnimVec3
		case "normalTexture":
			switch components[4] {
			case "scale":
				return Error("invalid/unsupported JSON pointer: " + pointer)
			case "extensions":
				switch components[5] {
				case "KHR_texture_transform":
					switch components[6] {
					case "offset":
						channel.target = &material.normalMapOffset
						channel.targetType = AnimVec2
					case "scale":
						channel.target = &material.normalMapScale
						channel.targetType = AnimVec2
					case "rotation":
						channel.target = &material.normalMapRotation
						channel.targetType = AnimFloat
					default:
						return Error("invalid/unsupported JSON pointer: " + pointer)
					}
				default:
					return Error("invalid/unsupported JSON pointer: " + pointer)
				}
			default:
				return Error("invalid/unsupported JSON pointer: " + pointer)
			}
		case "emissiveTexture":
			switch components[4] {
			case "extensions":
				switch components[5] {
				case "KHR_texture_transform":
					switch components[6] {
					case "offset":
						channel.target = &material.emissionMapOffset
						channel.targetType = AnimVec2
					case "scale":
						channel.target = &material.emissionMapScale
						channel.targetType = AnimVec2
					case "rotation":
						channel.target = &material.emissionMapRotation
						channel.targetType = AnimFloat
					default:
						return Error("invalid/unsupported JSON pointer: " + pointer)
					}
				default:
					return Error("invalid/unsupported JSON pointer: " + pointer)
				}

			default:
				return Error("invalid/unsupported JSON pointer: " + pointer)
			}
		case "occlusionTexture":
			switch components[4] {
			case "strength":
				channel.target = &material.ambientOcclusion
				channel.targetType = AnimFloat
			case "extensions":
				switch components[5] {
				case "KHR_texture_transform":
					switch components[6] {
					case "offset":
						channel.target = &material.ambientOcclusionMapOffset
						channel.targetType = AnimVec2
					case "scale":
						channel.target = &material.ambientOcclusionMapScale
						channel.targetType = AnimVec2
					case "rotation":
						channel.target = &material.ambientOcclusionMapRotation
						channel.targetType = AnimFloat
					default:
						return Error("invalid/unsupported JSON pointer: " + pointer)
					}
				default:
					return Error("invalid/unsupported JSON pointer: " + pointer)
				}

			default:
				return Error("invalid/unsupported JSON pointer: " + pointer)
			}
		case "pbrMetallicRoughness":
			switch components[4] {
			case "baseColorFactor":
				channel.target = &material.baseColorFactor
				channel.targetType = AnimVec4
			case "metallicFactor":
				channel.target = &material.metallic
				channel.targetType = AnimFloat
			case "roughnessFactor":
				channel.target = &material.roughness
				channel.targetType = AnimFloat
			case "extensions":
				switch components[5] {
				case "KHR_texture_transform":
					switch components[6] {
					case "offset":
						channel.target = &material.metallicRoughnessMapOffset
						channel.targetType = AnimVec2
					case "scale":
						channel.target = &material.metallicRoughnessMapScale
						channel.targetType = AnimVec2
					case "rotation":
						channel.target = &material.metallicRoughnessMapRotation
						channel.targetType = AnimFloat
					default:
						return Error("invalid/unsupported JSON pointer: " + pointer)
					}
				default:
					return Error("invalid/unsupported JSON pointer: " + pointer)
				}
			default:
				return Error("invalid/unsupported JSON pointer: " + pointer)
			}
		default:
			return Error("invalid/unsupported JSON pointer: " + pointer)
		}
	case "meshes":
		index, _ := strconv.Atoi(components[2])
		mesh := m.meshes[index]
		switch components[3] {
		case "weights":
			channel.target = &mesh.morphTargetWeights
			if len(components) > 4 {
				elemIndex, _ := strconv.Atoi(components[4])
				channel.elemIndex = new(uint32)
				*channel.elemIndex = uint32(elemIndex)
				channel.targetType = AnimVecElem
			} else {
				channel.targetType = MorphTargetWeight
			}
		default:
			return Error("invalid/unsupported JSON pointer: " + pointer)
		}
	case "nodes":
		index, _ := strconv.Atoi(components[2])
		channel.nodeIndex = new(uint32)
		*channel.nodeIndex = uint32(index)
		node := m.nodes[index]
		switch components[3] {
		case "translation":
			channel.target = &node.transition
			channel.targetType = TRSTranslation
		case "rotation":
			channel.target = &node.rotation
			channel.targetType = TRSRotation
		case "scale":
			channel.target = &node.scale
			channel.targetType = TRSScale
		case "weights":
			channel.target = &node.morphTargetWeights
			if len(components) > 4 {
				elemIndex, _ := strconv.Atoi(components[4])
				channel.elemIndex = new(uint32)
				*channel.elemIndex = uint32(elemIndex)
				channel.targetType = AnimVecElem
			} else {
				channel.targetType = MorphTargetWeight
			}
		default:
			return Error("invalid/unsupported JSON pointer: " + pointer)
		}
	case "extensions":
		switch components[2] {
		case "KHR_lights_punctual":
			switch components[3] {
			case "lights":
				index, _ := strconv.Atoi(components[4])
				light := m.lights[index]
				switch components[5] {
				case "color":
					channel.target = &light.color
					channel.targetType = AnimVec3
				case "intensity":
					channel.target = &light.intensity
					channel.targetType = AnimFloat
				case "range":
					channel.target = &light.lightRange
					channel.targetType = AnimFloat
				case "spot":
					switch components[6] {
					case "innerConeAngle":
						channel.target = &light.innerConeAngle
						channel.targetType = AnimFloat
					case "outerConeAngle":
						channel.target = &light.outerConeAngle
						channel.targetType = AnimFloat
					default:
						return Error("invalid/unsupported JSON pointer: " + pointer)
					}
				default:
					return Error("invalid/unsupported JSON pointer: " + pointer)
				}
			default:
				return Error("invalid/unsupported JSON pointer: " + pointer)
			}

		default:
			return Error("invalid/unsupported JSON pointer: " + pointer)
		}
	default:
		return Error("invalid/unsupported JSON pointer: " + pointer)
	}
	return nil
}

func (model *Model) calculateAnimInterpolation(interpolation GLTFAnimationInterpolation, sampler *GLTFAnimationSampler, animTime float32, prevIndex int, length int) []float32 {
	if interpolation == InterpolationStep || prevIndex == -1 || len(model.animationTimeStamps[sampler.inputIndex]) == 1 {
		if prevIndex == -1 {
			prevIndex = 0
		}
		newVals := make([]float32, length)
		for i := 0; i < length; i++ {
			if interpolation == InterpolationCubicSpline {
				newVals[i] = sampler.output[prevIndex*3*length+i+length]
			} else {
				newVals[i] = sampler.output[prevIndex*length+i]
			}
		}
		return newVals
	}
	if interpolation == InterpolationLinear {
		rate := (animTime - model.animationTimeStamps[sampler.inputIndex][prevIndex]) / (model.animationTimeStamps[sampler.inputIndex][prevIndex+1] - model.animationTimeStamps[sampler.inputIndex][prevIndex])
		newVals := make([]float32, length)
		for i := 0; i < length; i++ {
			newVals[i] = sampler.output[prevIndex*length+i]*(1-rate) + sampler.output[(prevIndex+1)*length+i]*rate
		}
		return newVals
	} else {
		delta := (model.animationTimeStamps[sampler.inputIndex][prevIndex+1] - model.animationTimeStamps[sampler.inputIndex][prevIndex])
		rate := (animTime - model.animationTimeStamps[sampler.inputIndex][prevIndex]) / delta
		rateSquare := rate * rate
		rateCube := rateSquare * rate
		newVals := make([]float32, length)
		for i := 0; i < length; i++ {
			newVals[i] = (2*rateCube-3*rateSquare+1)*sampler.output[prevIndex*3*length+i+length] + delta*(rateCube-2*rateSquare+rate)*sampler.output[prevIndex*3*length+i+length*2] + (-2*rateCube+3*rateSquare)*sampler.output[(prevIndex+1)*3*length+i+length] + delta*(rateCube-rateSquare)*sampler.output[(prevIndex+1)*3*length+i]
		}
		return newVals
	}
}

func (model *Model) calculateAnimQuatInterpolation(interpolation GLTFAnimationInterpolation, sampler *GLTFAnimationSampler, animTime float32, prevIndex int) [4]float32 {
	if interpolation == InterpolationStep || prevIndex == -1 || len(model.animationTimeStamps[sampler.inputIndex]) == 1 {
		if prevIndex == -1 {
			prevIndex = 0
		}
		if interpolation == InterpolationCubicSpline {
			newVals := [4]float32{
				sampler.output[prevIndex*12+4],
				sampler.output[prevIndex*12+4+1],
				sampler.output[prevIndex*12+4+2],
				sampler.output[prevIndex*12+4+3],
			}
			return newVals
		} else {
			newVals := [4]float32{
				sampler.output[prevIndex*4],
				sampler.output[prevIndex*4+1],
				sampler.output[prevIndex*4+2],
				sampler.output[prevIndex*4+3],
			}
			return newVals
		}
	}
	if interpolation == InterpolationLinear {
		rate := (animTime - model.animationTimeStamps[sampler.inputIndex][prevIndex]) / (model.animationTimeStamps[sampler.inputIndex][prevIndex+1] - model.animationTimeStamps[sampler.inputIndex][prevIndex])
		q1 := mgl.Quat{sampler.output[prevIndex*4+3], mgl.Vec3{sampler.output[prevIndex*4], sampler.output[prevIndex*4+1], sampler.output[prevIndex*4+2]}}
		q2 := mgl.Quat{sampler.output[(prevIndex+1)*4+3], mgl.Vec3{sampler.output[(prevIndex+1)*4], sampler.output[(prevIndex+1)*4+1], sampler.output[(prevIndex+1)*4+2]}}
		dotProduct := q1.Dot(q2)
		if dotProduct < 0 {
			q1 = q1.Inverse()
		}
		q := mgl.QuatSlerp(q1, q2, rate)
		newVals := [4]float32{q.X(), q.Y(), q.Z(), q.W}
		return newVals
	} else {
		delta := (model.animationTimeStamps[sampler.inputIndex][prevIndex+1] - model.animationTimeStamps[sampler.inputIndex][prevIndex])
		rate := (animTime - model.animationTimeStamps[sampler.inputIndex][prevIndex]) / delta
		rateSquare := rate * rate
		rateCube := rateSquare * rate
		q := mgl.Quat{(2*rateCube-3*rateSquare+1)*sampler.output[prevIndex*12+3+4] + delta*(rateCube-2*rateSquare+rate)*sampler.output[prevIndex*12+3+8] + (-2*rateCube+3*rateSquare)*sampler.output[(prevIndex+1)*12+3+4] + delta*(rateCube-rateSquare)*sampler.output[(prevIndex+1)*12+3],
			mgl.Vec3{
				(2*rateCube-3*rateSquare+1)*sampler.output[prevIndex*12+4] + delta*(rateCube-2*rateSquare+rate)*sampler.output[prevIndex*12+8] + (-2*rateCube+3*rateSquare)*sampler.output[(prevIndex+1)*12+4] + delta*(rateCube-rateSquare)*sampler.output[(prevIndex+1)*12],
				(2*rateCube-3*rateSquare+1)*sampler.output[prevIndex*12+1+4] + delta*(rateCube-2*rateSquare+rate)*sampler.output[prevIndex*12+1+8] + (-2*rateCube+3*rateSquare)*sampler.output[(prevIndex+1)*12+1+4] + delta*(rateCube-rateSquare)*sampler.output[(prevIndex+1)*12+1],
				(2*rateCube-3*rateSquare+1)*sampler.output[prevIndex*12+2+4] + delta*(rateCube-2*rateSquare+rate)*sampler.output[prevIndex*12+2+8] + (-2*rateCube+3*rateSquare)*sampler.output[(prevIndex+1)*12+2+4] + delta*(rateCube-rateSquare)*sampler.output[(prevIndex+1)*12+2],
			}}.Normalize()
		newVals := [4]float32{q.X(), q.Y(), q.Z(), q.W}
		return newVals
	}
}

func (anim *GLTFAnimation) toggle(enabled bool) {
	if enabled {
		anim.enabled = enabled
		anim.time = 0
	} else if anim.enabled != enabled {
		anim.enabled = enabled
		for _, channel := range anim.channels {
			channel.target.rest()
		}
	}
}

func (model *Model) step(turbo float32) {
	for _, anim := range model.animations {
		if anim.enabled == false {
			continue
		}
		anim.time += turbo / 60
		for anim.time >= anim.duration && anim.duration > 0 && (anim.loopCount < 0 || anim.loop < anim.loopCount) {
			anim.time -= anim.duration
			anim.loop += 1
		}
		time := 60 * float64(anim.time)
		if math.Abs(time-math.Floor(time)) < 0.001 {
			anim.time = float32(math.Floor(time) / 60)
		} else if math.Abs(float64(time)-math.Ceil(float64(time))) < 0.001 {
			anim.time = float32(math.Ceil(time) / 60)
		}
		if anim.time >= anim.duration && anim.duration > 0 {
			anim.time = anim.duration
		}
		for _, channel := range anim.channels {
			sampler := anim.samplers[channel.samplerIndex]
			prevIndex := 0
			for i, t := range model.animationTimeStamps[sampler.inputIndex] {
				if anim.time > 0 && anim.time <= t {
					prevIndex = i - 1
					break
				}
			}
			switch channel.targetType {
			case TRSTranslation, TRSScale:
				node := model.nodes[*channel.nodeIndex]
				newVals := model.calculateAnimInterpolation(sampler.interpolation, sampler, anim.time, prevIndex, 3)
				oldVals := channel.target.getValue().([3]float32)
				if newVals[0] != oldVals[0] || newVals[1] != oldVals[1] || newVals[2] != oldVals[2] {
					node.transformChanged = true
				}
				channel.target.animate([3]float32{newVals[0], newVals[1], newVals[2]})
			case TRSRotation:
				node := model.nodes[*channel.nodeIndex]
				newVals := model.calculateAnimQuatInterpolation(sampler.interpolation, sampler, anim.time, prevIndex)
				oldVals := channel.target.getValue().([4]float32)
				if newVals[0] != oldVals[0] || newVals[1] != oldVals[1] || newVals[2] != oldVals[2] || newVals[3] != oldVals[3] {
					node.transformChanged = true
				}
				channel.target.animate(newVals)
			case MorphTargetWeight, AnimVec:
				newVals := model.calculateAnimInterpolation(sampler.interpolation, sampler, anim.time, prevIndex, len(channel.target.getValue().([]float32)))
				channel.target.animate(newVals)
			case AnimVec2:
				newVals := model.calculateAnimInterpolation(sampler.interpolation, sampler, anim.time, prevIndex, 2)
				channel.target.animate([2]float32{newVals[0], newVals[1]})
			case AnimVec3:
				newVals := model.calculateAnimInterpolation(sampler.interpolation, sampler, anim.time, prevIndex, 3)
				channel.target.animate([3]float32{newVals[0], newVals[1], newVals[2]})
			case AnimVec4:
				newVals := model.calculateAnimInterpolation(sampler.interpolation, sampler, anim.time, prevIndex, 4)
				channel.target.animate([4]float32{newVals[0], newVals[1], newVals[2], newVals[3]})
			case AnimVecElem:
				replaceVals := model.calculateAnimInterpolation(sampler.interpolation, sampler, anim.time, prevIndex, 1)
				newVals := channel.target.getValue().([]float32)
				newVals[*channel.elemIndex] = replaceVals[0]
				channel.target.animate(newVals)
			case AnimFloat:
				newVals := model.calculateAnimInterpolation(sampler.interpolation, sampler, anim.time, prevIndex, 1)
				channel.target.animate(newVals[0])
			case AnimQuat:
				newVals := model.calculateAnimQuatInterpolation(sampler.interpolation, sampler, anim.time, prevIndex)
				channel.target.animate(newVals)
			}
		}
	}
}

func (skin *Skin) calculateSkinMatrices(inverseGlobalTransform mgl.Mat4, nodes []*Node) {
	matrices := make([]float32, len(skin.joints)*12*2)
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
			matrices[i*24+j] = matrix[j]
		}
		normalMatrix := matrix.Transpose().Inv().Transpose()
		for j := 0; j < 12; j++ {
			matrices[i*24+12+j] = normalMatrix[j]
		}
	}
	skin.texture.tex.SetPixelData(matrices)
}

func (model *Model) reset() {
	for _, anim := range model.animations {
		anim.time = 0
		anim.enabled = anim.defaultEnabled
		anim.loop = 0
	}
	for _, node := range model.nodes {
		node.visible = true
	}
}
