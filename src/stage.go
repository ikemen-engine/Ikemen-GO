package main

import (
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"

	"gopkg.in/ini.v1"
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

func (bga *bgAction) action(updateTime bool) {
	for i := 0; i < 2; i++ {
		bga.pos[i] += bga.vel[i]
		if bga.sinlooptime[i] > 0 {
			bga.sinoffset[i] = bga.radius[i] * float32(math.Sin(
				2*math.Pi*float64(bga.sintime[i])/float64(bga.sinlooptime[i])))
			if updateTime {
				bga.sintime[i]++
				if bga.sintime[i] >= bga.sinlooptime[i] {
					bga.sintime[i] = 0
				}
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
	_type                 BgType
	palfx                 *PalFX
	anim                  *Animation
	bga                   bgAction
	video                 *bgVideo
	id                    int32
	start                 [2]float32
	xofs                  float32
	delta                 [2]float32
	width                 [2]int32
	xscale                [2]float32
	rasterx               [2]float32
	yscalestart           float32
	yscaledelta           float32
	actionno              int32
	startv                [2]float32
	startrad              [2]float32
	startsint             [2]int32
	startsinlt            [2]int32
	visible               bool
	enabled               bool
	positionlink          bool
	layerno               int32
	autoresizeparallax    bool
	autoresizeparallaxSet bool
	notmaskwindow         int32
	startrect             [4]int32
	windowdelta           [2]float32
	scalestart            [2]float32
	scaledelta            [2]float32
	zoomdelta             [2]float32
	zoomdeltaSet          bool
	zoomscaledelta        [2]float32
	xbottomzoomdelta      float32
	roundpos              bool
	rot                   Rotation
	fLength               float32
	projection            Projection
	xshear                float32
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
	sff *Sff, at AnimationTable, sProps StageProps, def string, startlayer int32) (*backGround, error) {
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
		bg.video = &bgVideo{}
	case 'D', 'd':
		bg._type = BG_Dummy
	default:
		return bg, nil
	}
	var tmp int32
	is.ReadI32("layerno", &bg.layerno)
	bg.layerno += startlayer

	if bg._type == BG_Video {
		if bg.video == nil {
			bg.video = &bgVideo{}
		}
		path := is["path"]
		LoadFile(&path, []string{def, "", sys.motif.Def, "data/", "video/"}, func(filename string) error {
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
			} else {
				return nil, Error(fmt.Sprint("Missing action: %d", bg.actionno))
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
				// To return an error for a missing sprite we'd need to read the SFF, which would slow this down
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
	if _, ok := is["autoresizeparallax"]; ok {
		bg.autoresizeparallaxSet = true
		is.ReadBool("autoresizeparallax", &bg.autoresizeparallax)
	}
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
	if is.readF32ForStage("zoomdelta", &bg.zoomdelta[0], &bg.zoomdelta[1]) {
		bg.zoomdeltaSet = true
	}
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
			if group != -1 && number != -1 {
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

	bg.visible = true
	bg.enabled = true

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
		scalestartX = Abs(scalestartX)
		bg.xofs = scalestartX * ((-float32(bg.width[0]) / 2) + float32(bg.anim.spr.Offset[0])*bg.xscale[0])
		bg.anim.isParallax = true
	}

	// Calculate raster x ratio and base x scale
	xras := (bg.rasterx[1] - bg.rasterx[0]) / bg.rasterx[0]
	xbs, dx := bg.xscale[1], Max(0, bg.delta[0]*bgscl)

	// Initialize local scaling factors
	var sclx_recip, sclx, scly float32 = 1, 1, 1
	lscl := [...]float32{stglscl * stgscl[0], stglscl * stgscl[1]}

	// Handle zoom scaling if zoomdelta is specified
	var Yzoomdelta float32 = 1
	if bg.zoomdelta[0] != math.MaxFloat32 {
		sclx = drawscl + (1-drawscl)*(1-bg.zoomdelta[0])
		scly = drawscl + (1-drawscl)*(1-bg.zoomdelta[1])
		Yzoomdelta = bg.zoomdelta[1]
		if !bg.autoresizeparallax {
			sclx_recip = 1 + bg.zoomdelta[0]*((1/(sclx*lscl[0])*lscl[0])-1)
		}
	} else {
		sclx = Max(0, drawscl+(1-drawscl)*(1-dx))
		scly = Max(0, drawscl+(1-drawscl)*(1-Max(0, bg.delta[1]*bgscl)))
		Yzoomdelta = Max(0, bg.delta[1]*bgscl)
	}

	// Adjust x scale and x bottom zoom if autoresizeparallax is enabled
	if sclx != 0 && bg.autoresizeparallax {
		tmp := 1 / sclx
		if bg.xbottomzoomdelta != math.MaxFloat32 {
			xbs *= Max(0, drawscl+(1-drawscl)*(1-bg.xbottomzoomdelta*(xbs/bg.xscale[0]))) * tmp
		} else {
			xbs *= Max(0, drawscl+(1-drawscl)*(1-dx*(xbs/bg.xscale[0]))) * tmp
		}
		tmp *= Max(0, drawscl+(1-drawscl)*(1-dx*(xras+1)))
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
		y -= (sys.cam.zoomanchorcorrection / (scly * stglscl * stgscl[1])) * Yzoomdelta
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
			wscl[i] = Max(0, drawscl+(1-drawscl)*(1-Max(0, bg.zoomdelta[i]))) * bgscl * lscl[i]
		} else {
			wscl[i] = Max(0, drawscl+(1-drawscl)*(1-Max(0, bg.windowdelta[i]*bgscl))) * bgscl * lscl[i]
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
		startrect1 -= (sys.cam.zoomanchorcorrection / scly) * Yzoomdelta
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
			if bg.video == nil {
				return
			}
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
		xsoffset := -bg.xshear * Sign(bg.scalestart[1]) * (float32(bg.anim.spr.Offset[1]) * scly)

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
			ys*ys3, xras*x/(Abs(ys*ys3)*lscl[1]*float32(bg.anim.spr.Size[1])*bg.scalestart[1])*sclx_recip*bg.scalestart[1]-bg.xshear,
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
			bgc.hue = bgc.hue / 512
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

// Return Select Stage Info
func (s *Stage) si() *SelectStage {
	return &sys.sel.stagelist[sys.sel.selectedStageNo-1]
}

type Stage struct {
	def             string
	name            string
	displayname     string
	author          string
	nameLow         string
	displaynameLow  string
	authorLow       string
	attachedchardef []string
	sff             *Sff
	animTable       AnimationTable
	bg              []*backGround
	bgc             []bgCtrl
	bga             bgAction // For position linking
	sdw             stageShadow
	reflection      stageShadow
	p               [MaxPlayerNo]stagePlayer
	leftbound       float32
	rightbound      float32
	screenleft      int32
	screenright     int32
	zoffsetlink     int32
	hires           bool
	autoturn        bool
	resetbg         bool
	debugbg         bool
	bgclearcolor    [3]int32
	localscl        float32
	scale           [2]float32
	mainstage       bool
	stageCamera     stageCamera
	stageTime       int32
	music           Music
	bgmState        BGMState
	bgmratio        float32
	constants       map[string]float32
	partnerspacing  int32
	ikemenver       [3]uint16
	ikemenverF      float32
	mugenver        [2]uint16
	mugenverF       float32
	reload          bool
	stageprops      StageProps
	model           *Model
	topbound        float32
	botbound        float32
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
		stageCamera:    *newStageCamera(),
		music:          make(Music),
		bgmratio:       0.3,
		constants:      make(map[string]float32),
		partnerspacing: 25,
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

	// Helper to get localized section or fallback to default
	// TODO: This looks cleaner than what's used in char.go. Maybe standardize it
	getSection := func(baseName string) (IniSection, string) {
		langKey := fmt.Sprintf("%v.%v", sys.cfg.Config.Language, baseName)
		if secList, ok := defmap[langKey]; ok && len(secList) > 0 {
			return secList[0], langKey
		}
		if secList, ok := defmap[baseName]; ok && len(secList) > 0 {
			return secList[0], baseName
		}
		return nil, ""
	}

	// Info group
	if sec, _ := getSection("info"); sec != nil {
		var ok bool
		s.name, ok, _ = sec.getText("name")
		if !ok {
			s.name = def
		}
		s.displayname, ok, _ = sec.getText("displayname")
		if !ok {
			s.displayname = s.name
		}
		s.author, _, _ = sec.getText("author")
		s.nameLow = strings.ToLower(s.name)
		s.displaynameLow = strings.ToLower(s.displayname)
		s.authorLow = strings.ToLower(s.author)
		// Clear then read MugenVersion
		s.mugenver = [2]uint16{}
		s.mugenverF = 0
		if str, ok := sec["mugenversion"]; ok {
			s.mugenver, s.mugenverF = ParseMugenVersion(str)
		}
		// Clear then read IkemenVersion
		s.ikemenver = [3]uint16{}
		if str, ok := sec["ikemenversion"]; ok {
			s.ikemenver, s.ikemenverF = ParseIkemenVersion(str)
		}
		// If the MUGEN version is lower than 1.0, default to camera pixel rounding (floor)
		if s.ikemenver[0] == 0 && s.ikemenver[1] == 0 && s.mugenver[0] != 1 {
			s.stageprops.roundpos = true
		}
		// AttachedChars
		ac := 0
		for i := range sec {
			if !strings.HasPrefix(i, "attachedchar") {
				continue
			}
			if suffix := strings.TrimPrefix(i, "attachedchar"); suffix != "" {
				if _, err := strconv.Atoi(suffix); err != nil {
					continue
				}
			}
			if ac >= MaxAttachedChar {
				sys.appendToConsole(s.warn() + fmt.Sprintf("Can only define up to %d attachedchar(s). '%s' ignored.", MaxAttachedChar, i))
				continue
			}
			if err := sec.LoadFile(i, []string{def, "", sys.motif.Def, "data/"}, func(filename string) error {
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
			for k, v := range sec {
				if r.MatchString(k) {
					re := regexp.MustCompile("[0-9]+")
					submatchall := re.FindAllString(k, -1)
					if len(submatchall) == 1 {
						if err := LoadFile(&v, []string{def, "", sys.motif.Def, "data/"}, func(filename string) error {
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
			sec.ReadBool("roundloop", &sys.stageLoop)
		}
	}

	// StageInfo group. Needs to be read before most other groups so that localcoord is known
	if sec, _ := getSection("stageinfo"); sec != nil {
		sec.ReadI32("zoffset", &s.stageCamera.zoffset)
		sec.ReadI32("zoffsetlink", &s.zoffsetlink)
		sec.ReadBool("hires", &s.hires)
		sec.ReadBool("autoturn", &s.autoturn)
		sec.ReadBool("resetbg", &s.resetbg)
		sec.readI32ForStage("localcoord", &s.stageCamera.localcoord[0],
			&s.stageCamera.localcoord[1])
		sec.ReadF32("xscale", &s.scale[0])
		sec.ReadF32("yscale", &s.scale[1])
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
	if sec, _ := getSection("constants"); sec != nil {
		for key, value := range sec {
			s.constants[key] = float32(Atof(value))
		}
	}

	// Scaling group
	if sec, _ := getSection("scaling"); sec != nil {
		if s.mugenver[0] != 1 || s.ikemenver[0] >= 1 { // mugen 1.0+ removed support for z-axis, IKEMEN-Go 1.0 adds it back
			sec.ReadF32("topz", &s.stageCamera.topz)
			sec.ReadF32("botz", &s.stageCamera.botz)
			sec.ReadF32("topscale", &s.stageCamera.ztopscale)
			sec.ReadF32("botscale", &s.stageCamera.zbotscale)
			sec.ReadF32("depthtoscreen", &s.stageCamera.depthtoscreen)
		}
	}

	// Bound group
	if sec, _ := getSection("bound"); sec != nil {
		sec.ReadI32("screenleft", &s.screenleft)
		sec.ReadI32("screenright", &s.screenright)
	}

	// PlayerInfo Group
	if sec, _ := getSection("playerinfo"); sec != nil {
		sec.ReadI32("partnerspacing", &s.partnerspacing)
		for i := range s.p {
			// Defaults
			if i >= 2 {
				s.p[i].startx = s.p[i-2].startx + s.partnerspacing*int32(2*(i%2)-1) // Previous partner + partnerspacing
				s.p[i].starty = s.p[i%2].starty                                     // Same as players 1 or 2
				s.p[i].startz = s.p[i%2].startz                                     // Same as players 1 or 2
				s.p[i].facing = int32(1 - 2*(i%2))                                  // By team side
			}
			// pXstartx
			sec.ReadI32(fmt.Sprintf("p%dstartx", i+1), &s.p[i].startx)
			// pXstarty
			sec.ReadI32(fmt.Sprintf("p%dstarty", i+1), &s.p[i].starty)
			// pXstartz
			sec.ReadI32(fmt.Sprintf("p%dstartz", i+1), &s.p[i].startz)
			// pXfacing
			sec.ReadI32(fmt.Sprintf("p%dfacing", i+1), &s.p[i].facing)
		}
		sec.ReadF32("leftbound", &s.leftbound)
		sec.ReadF32("rightbound", &s.rightbound)
		sec.ReadF32("topbound", &s.topbound)
		sec.ReadF32("botbound", &s.botbound)
	}

	// Camera group
	if sec, _ := getSection("camera"); sec != nil {
		sec.ReadI32("startx", &s.stageCamera.startx)
		sec.ReadI32("starty", &s.stageCamera.starty)
		sec.ReadI32("boundleft", &s.stageCamera.boundleft)
		sec.ReadI32("boundright", &s.stageCamera.boundright)
		sec.ReadI32("boundhigh", &s.stageCamera.boundhigh)
		sec.ReadI32("boundlow", &s.stageCamera.boundlow)
		sec.ReadF32("verticalfollow", &s.stageCamera.verticalfollow)
		sec.ReadI32("floortension", &s.stageCamera.floortension)
		sec.ReadI32("tension", &s.stageCamera.tension)
		sec.ReadF32("tensionvel", &s.stageCamera.tensionvel)
		sec.ReadI32("overdrawhigh", &s.stageCamera.overdrawhigh) // TODO: not implemented
		sec.ReadI32("overdrawlow", &s.stageCamera.overdrawlow)
		sec.ReadI32("cuthigh", &s.stageCamera.cuthigh)
		sec.ReadI32("cutlow", &s.stageCamera.cutlow)
		sec.ReadF32("startzoom", &s.stageCamera.startzoom)
		sec.ReadF32("fov", &s.stageCamera.fov)
		sec.ReadF32("yshift", &s.stageCamera.yshift)
		sec.ReadF32("near", &s.stageCamera.near)
		sec.ReadF32("far", &s.stageCamera.far)
		sec.ReadBool("autocenter", &s.stageCamera.autocenter)
		sec.ReadF32("yscrollspeed", &s.stageCamera.yscrollspeed)
		sec.ReadF32("verticalfollowzoomdelta", &s.stageCamera.verticalfollowzoomdelta)
		sec.ReadBool("lowestcap", &s.stageCamera.lowestcap)
		sec.ReadF32("zoomin", &s.stageCamera.zoomin)
		sec.ReadF32("zoomout", &s.stageCamera.zoomout)
		if s.stageCamera.zoomin == 1 && s.stageCamera.zoomout == 1 {
			if sys.cfg.Debug.ForceStageZoomin > 0 {
				s.stageCamera.zoomin = sys.cfg.Debug.ForceStageZoomin
			}
			if sys.cfg.Debug.ForceStageZoomout > 0 {
				s.stageCamera.zoomout = sys.cfg.Debug.ForceStageZoomout
			}
		}
		anchor, zoomanchorOk, _ := sec.getText("zoomanchor")
		if strings.ToLower(anchor) == "bottom" {
			s.stageCamera.zoomanchor = true
		}

		autoZoomExisted := sec.ReadBool("autozoom", &s.stageCamera.autoZoom)
		if sys.cfg.Debug.ForceStageAutoZoom && !autoZoomExisted &&
			(s.stageCamera.zoomin == 1 || sys.cfg.Debug.ForceStageZoomin > 0) && (s.stageCamera.zoomout == 1 || sys.cfg.Debug.ForceStageZoomout > 0) {
			s.stageCamera.autoZoom = true
		}

		if s.stageCamera.autoZoom {
			if s.stageCamera.zoomin == 1 {
				s.stageCamera.zoomin = sys.cam.LegacyZoomMax
			}
			if s.stageCamera.zoomout == 1 {
				s.stageCamera.zoomout = sys.cam.LegacyZoomMin
			}
			if !zoomanchorOk {
				s.stageCamera.zoomanchor = true
			}
			s.stageCamera.zoomindelay = 25
			s.stageCamera.zoominspeed = 0.4
			s.stageCamera.zoomoutspeed = 0.4
			s.stageCamera.boundhighzoomdelta = 1.0
		}
		sec.ReadF32("zoomindelay", &s.stageCamera.zoomindelay)
		sec.ReadF32("zoominspeed", &s.stageCamera.zoominspeed)
		sec.ReadF32("zoomoutspeed", &s.stageCamera.zoomoutspeed)
		sec.ReadF32("boundhighzoomdelta", &s.stageCamera.boundhighzoomdelta)
		if sec.ReadI32("tensionlow", &s.stageCamera.tensionlow) {
			s.stageCamera.ytensionenable = true
			sec.ReadI32("tensionhigh", &s.stageCamera.tensionhigh)
		}
		// Camera group warnings
		// Warn when camera boundaries are smaller than player boundaries
		if int32(s.leftbound) > s.stageCamera.boundleft || int32(s.rightbound) < s.stageCamera.boundright {
			sys.appendToConsole(s.warn() + "Player boundaries defined incorrectly")
		}
	}

	// Music group
	if sec, secName := getSection("music"); sec != nil {
		iniFile, err := ini.LoadSources(ini.LoadOptions{
			Insensitive:             true,
			SkipUnrecognizableLines: true,
		}, []byte(str))

		if err != nil {
			fmt.Printf("Failed to load INI file: %v\n", err)
			return nil, err
		}

		sec.ReadF32("bgmratio", &s.bgmratio)
		s.music = parseMusicSection(iniFile.Section(secName))
		s.music.DebugDump(fmt.Sprintf("Stage %s [%s]", def, secName))
	}

	// BGDef group
	if sec, _ := getSection("bgdef"); sec != nil {
		if sec.LoadFile("spr", []string{def, "", sys.motif.Def, "data/"}, func(filename string) error {
			sff, err := loadSff(filename, false, false, false)
			if err != nil {
				return err
			}
			*s.sff = *sff
			// SFF v2.01 was not available before Mugen 1.1, therefore we assume that's the minimum correct version for the stage
			if s.sff.header.Version[0] == 2 && s.sff.header.Version[2] == 1 {
				s.mugenver[0] = 1
				s.mugenver[1] = 1
			}
			return nil
		}); err != nil {
			return nil, err
		}
		if err = sec.LoadFile("model", []string{def, "", sys.motif.Def, "data/"}, func(filename string) error {
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
		sec.ReadBool("debugbg", &s.debugbg)
		sec.readI32ForStage("bgclearcolor", &s.bgclearcolor[0], &s.bgclearcolor[1], &s.bgclearcolor[2])
		sec.ReadBool("roundpos", &s.stageprops.roundpos)
	}

	// Model group
	if sec, _ := getSection("model"); sec != nil {
		if str, ok := sec["offset"]; ok {
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
		if str, ok := sec["scale"]; ok {
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
		if err = sec.LoadFile("environment", []string{def, "", sys.motif.Def, "data/"}, func(filename string) error {
			env, err := loadEnvironment(filename)
			if err != nil {
				return err
			}
			var intensity float32
			if sec.ReadF32("environmentintensity", &intensity) {
				env.environmentIntensity = intensity
			}
			s.model.environment = env
			return nil
		}); err != nil {
			return nil, err
		}
	}

	// Shadow group
	if sec, _ := getSection("shadow"); sec != nil {
		var tmp int32
		if sec.ReadI32("intensity", &tmp) {
			s.sdw.intensity = Clamp(tmp, 0, 255)
		}
		var r, g, b int32
		sec.readI32ForStage("color", &r, &g, &b)
		r, g, b = Clamp(r, 0, 255), Clamp(g, 0, 255), Clamp(b, 0, 255)
		// Disable color parameter specifically in Mugen 1.1 stages
		if s.ikemenver[0] == 0 && s.ikemenver[1] == 0 && s.mugenver[0] == 1 && s.mugenver[1] == 1 {
			r, g, b = 0, 0, 0
		}
		s.sdw.color = uint32(r<<16 | g<<8 | b)
		sec.ReadF32("xscale", &s.sdw.xscale)
		sec.ReadF32("yscale", &s.sdw.yscale)
		sec.readI32ForStage("fade.range", &s.sdw.fadeend, &s.sdw.fadebgn)
		sec.ReadF32("xshear", &s.sdw.xshear)
		sec.ReadF32("angle", &s.sdw.rot.angle)
		sec.ReadF32("xangle", &s.sdw.rot.xangle)
		sec.ReadF32("yangle", &s.sdw.rot.yangle)
		sec.ReadF32("focallength", &s.sdw.fLength)
		if str, ok := sec["projection"]; ok {
			switch strings.ToLower(strings.TrimSpace(str)) {
			case "orthographic":
				s.sdw.projection = Projection_Orthographic
			case "perspective":
				s.sdw.projection = Projection_Perspective
			case "perspective2":
				s.sdw.projection = Projection_Perspective2
			}
		}
		sec.readF32ForStage("offset", &s.sdw.offset[0], &s.sdw.offset[1])
		sec.readF32ForStage("window", &s.sdw.window[0], &s.sdw.window[1], &s.sdw.window[2], &s.sdw.window[3])
		sec.ReadF32("ydelta", &s.sdw.ydelta)
		// Shadow group warnings
		if s.sdw.fadeend > s.sdw.fadebgn {
			sys.appendToConsole(s.warn() + "Shadow fade.range defined incorrectly")
		}
	}

	// Reflection group
	if sec, _ := getSection("reflection"); sec != nil {
		s.reflection.color = 0xFFFFFF
		var tmp int32
		//sec.ReadBool("reflect", &reflect) // This parameter is documented in Mugen but doesn't do anything
		if sec.ReadI32("intensity", &tmp) {
			s.reflection.intensity = Clamp(tmp, 0, 255)
		}
		var r, g, b int32 = 0, 0, 0
		sec.readI32ForStage("color", &r, &g, &b)
		r, g, b = Clamp(r, 0, 255), Clamp(g, 0, 255), Clamp(b, 0, 255)
		s.reflection.color = uint32(r<<16 | g<<8 | b)
		if sec.ReadI32("layerno", &tmp) {
			s.reflection.layerno = Clamp(tmp, -1, 0)
		}
		sec.ReadF32("xscale", &s.reflection.xscale)
		sec.ReadF32("yscale", &s.reflection.yscale)
		sec.readI32ForStage("fade.range", &s.reflection.fadeend, &s.reflection.fadebgn)
		sec.ReadF32("xshear", &s.reflection.xshear)
		sec.ReadF32("angle", &s.reflection.rot.angle)
		sec.ReadF32("xangle", &s.reflection.rot.xangle)
		sec.ReadF32("yangle", &s.reflection.rot.yangle)
		sec.ReadF32("focallength", &s.reflection.fLength)
		if str, ok := sec["projection"]; ok {
			switch strings.ToLower(strings.TrimSpace(str)) {
			case "orthographic":
				s.reflection.projection = Projection_Orthographic
			case "perspective":
				s.reflection.projection = Projection_Perspective
			case "perspective2":
				s.reflection.projection = Projection_Perspective2
			}
		}
		sec.readF32ForStage("offset", &s.reflection.offset[0], &s.reflection.offset[1])
		sec.readF32ForStage("window", &s.reflection.window[0], &s.reflection.window[1], &s.reflection.window[2], &s.reflection.window[3])
		sec.ReadF32("ydelta", &s.reflection.ydelta)
	}

	// BG group
	var bglink *backGround
	for _, bgsec := range defmap["bg"] {
		if len(s.bg) > 0 && !s.bg[len(s.bg)-1].positionlink {
			bglink = s.bg[len(s.bg)-1]
		}
		bg, err := readBackGround(bgsec, bglink, s.sff, s.animTable, s.stageprops, def, 0)
		if err != nil {
			return nil, err
		}
		s.bg = append(s.bg, bg)
	}

	if s.stageCamera.autoZoom {
		for i := range s.bg {
			if s.bg[i]._type == BG_Parallax && !s.bg[i].autoresizeparallaxSet {
				s.bg[i].autoresizeparallax = true
			}

			if !s.bg[i].zoomdeltaSet {
				s.bg[i].zoomdelta[0] = math.MaxFloat32
			}
		}
	}

	bgcdef := *newBgCtrl()
	lnidx := 0
	for lnidx < len(lines) {
		is, name, _ := ReadIniSection(lines, &lnidx)
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
	// Handle Music
	s.music.act()

	link, zlink := 0, -1
	canStep := sys.tickFrame() && !s.paused()

	// Update animations and controllers
	if canStep {
		s.bgCtrlAction()
		s.bga.action(true)

		if s.model != nil {
			s.model.step(sys.turbo)
		}
	}

	// Always (every frame) sync decoder run state to global pause + Enable.
	// This prevents the decoder clock from advancing during pause.
	for i := range s.bg {
		if s.bg[i]._type == BG_Video && s.bg[i].video != nil {
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

		if canStep {
			s.bg[i].bga.action(b.enabled)
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
		}

		if b.enabled && canStep {
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
}

func (s *Stage) reset() {
	s.stageTime = 0
	s.sff.palList.ResetRemap()
	s.bga.clear()
	for i := range s.bg {
		s.bg[i].reset()
		// Ensure videos start paused, then rewind.
		if s.bg[i]._type == BG_Video && s.bg[i].video != nil {
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
		if b != nil && b._type == BG_Video && b.video != nil {
			b.video.Close()
		}
	}
}

func (s *Stage) warn() string {
	return fmt.Sprintf("%v: WARNING: Stage %v: ", sys.tickCount, s.name)
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
				s.bgc[i].hue = hue / 512
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
