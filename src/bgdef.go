// bgdef.go manages background definitions for stages and screenpacks.
// It parses stage `.def` files into runtime structures that interact with stage
// logic and the rendering pipeline each frame.
package main

import (
	"fmt"
	"math"
	"strings"

	mgl "github.com/go-gl/mathgl/mgl32"
)

func (bgct *bgcTimeLine) stepBGDef(s *BGDef) {
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

// BGDef bundles background layers, controllers, and rendering parameters parsed
// from stage or screenpack `.def` files. It holds pointers to the sprite file
// (SFF), animation table, and runtime state used by stage logic.
type BGDef struct {
	def          string         // [MUGEN-Compat] path to `.def` file
	localcoord   [2]float32     // [MUGEN-Compat] width,height in pixels (>0); default 320x240 [PITFALL] wrong values skew placement
	sff          *Sff           // [MUGEN-Compat] sprite file handle; nil if missing [PITFALL] nil yields invisible layers
	at           AnimationTable // [MUGEN-Compat] animation table extracted from the definition
	bg           []*backGround  // [MUGEN-Compat] ordered background layers
	bgc          []bgCtrl       // [MUGEN-Compat] background controller definitions
	bgct         bgcTimeLine    // [Ikemen-Only] runtime scheduler for controllers
	bga          bgAction       // [Ikemen-Only] global background action state
	resetbg      bool           // [Ikemen-Only] reset backgrounds between rounds
	localscl     float32        // [Ikemen-Only] pixel-to-screen scale factor
	scale        [2]float32     // [Ikemen-Only] extra scaling multipliers (1=default)
	stageprops   StageProps     // [MUGEN-Compat] camera and bound properties
	bgclearcolor [3]int32       // [MUGEN-Compat] RGB clear color (0-255); default black [PITFALL] omission may leave artifacts
	model        *Model         // [Ikemen-Only] optional 3D model reference
	sceneNumber  int32          // [Ikemen-Only] 3D scene index (>=0 to enable)
	fov          float32        // [Ikemen-Only] field of view in degrees; 0 disables perspective
	near         float32        // [Ikemen-Only] near clipping plane (>0) [PITFALL] zero causes depth issues
	far          float32        // [Ikemen-Only] far clipping plane (>near) [PITFALL] small values clip models
	modelOffset  [3]float32     // [Ikemen-Only] 3D model offset (x,y,z) in stage units
	modelScale   [3]float32     // [Ikemen-Only] 3D model scaling factors; default 1
}

// newBGDef initializes BGDef with default parameters before parsing a stage
// definition. Local coordinates default to 320x240 and scaling factors to 1.
// [PITFALL] Leaving localcoord at the default can misalign high-resolution assets.
func newBGDef(def string) *BGDef {
	s := &BGDef{def: def, localcoord: [...]float32{320, 240}, resetbg: true, localscl: 1, scale: [...]float32{1, 1}}
	s.stageprops = newStageProps()
	return s
}

// loadBGDef parses a background definition from a stage `.def` file. The file
// should contain INI-style sections such as `[info]`, `[<bgname>]`, and
// `<bgname>ctrldef`. Lines follow `key = value` format with comma-separated
// numbers for vectors. Returns an error if the file cannot be read. Stage
// dependencies include an accompanying SFF sprite file and optional 3D model.
// [PITFALL] Missing lines fall back to defaults and may cause rendering artifacts.
func loadBGDef(sff *Sff, model *Model, def string, bgname string) (*BGDef, error) {
	s := newBGDef(def)
	str, err := LoadText(def)
	if err != nil {
		return nil, err
	}
	lines, i := SplitAndTrim(str, "\n"), 0
	defmap := make(map[string][]IniSection)
	for i < len(lines) {
		is, name, _ := ReadIniSection(lines, &i)
		if i := strings.IndexAny(name, " \t"); i >= 0 {
			if name[:i] == bgname {
				defmap[bgname] = append(defmap[bgname], is)
			}
		} else {
			defmap[name] = append(defmap[name], is)
		}
	}
	i = 0
	if sec := defmap["info"]; len(sec) > 0 {
		sec[0].readF32ForStage("localcoord", &s.localcoord[0], &s.localcoord[1]) // [PITFALL] defaults to 320,240
	}
	if sec := defmap[fmt.Sprintf("%sdef", bgname)]; len(sec) > 0 {
		sec[0].readI32ForStage("bgclearcolor", &s.bgclearcolor[0], &s.bgclearcolor[1], &s.bgclearcolor[2]) // [PITFALL] default black may expose previous frame
		s.sceneNumber = -1
		sec[0].readI32ForStage("scenenumber", &s.sceneNumber)
		sec[0].readF32ForStage("fov", &s.fov)
		sec[0].readF32ForStage("near", &s.near)
		sec[0].readF32ForStage("far", &s.far)
		if offset := sec[0].readF32CsvForStage("modeloffset"); len(offset) == 3 {
			s.modelOffset = [3]float32{offset[0], offset[1], offset[2]}
		} // [PITFALL] omitted offset anchors model at origin
		if scale := sec[0].readF32CsvForStage("modelscale"); len(scale) == 3 {
			s.modelScale = [3]float32{scale[0], scale[1], scale[2]}
		} // [PITFALL] omitted scale leaves model at full size
	}
	s.sff = sff
	s.model = model
	s.at = ReadAnimationTable(s.sff, &s.sff.palList, lines, &i)
	var bglink *backGround
	for _, bgsec := range defmap[bgname] {
		if len(s.bg) > 0 && s.bg[len(s.bg)-1].positionlink {
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
		case bgname + "ctrldef":
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
		case bgname + "ctrl":
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
			bgc.read(is, len(s.bgc))
			s.bgc = append(s.bgc, *bgc)
		}
	}
	s.localscl = 240 / s.localcoord[1]
	return s, nil
}
func (s *BGDef) getBg(id int32) (bg []*backGround) {
	if id >= 0 {
		for _, b := range s.bg {
			if b.id == id {
				bg = append(bg, b)
			}
		}
	}
	return
}
func (s *BGDef) runBgCtrl(bgc *bgCtrl) {
	bgc.currenttime++
	switch bgc._type {
	case BT_Anim:
		a := s.at.get(bgc.v[0])
		if a != nil {
			for i := range bgc.bg {
				bgc.bg[i].actionno = bgc.v[0]
				bgc.bg[i].anim = *a
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
	case BT_SinX, BT_SinY:
		ii := Btoi(bgc._type == BT_SinY)
		if bgc.v[0] == 0 {
			bgc.v[1] = 0
		}
		a := float32(bgc.v[2]) / 360
		st := int32((a - float32(int32(a))) * float32(bgc.v[1]))
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
func (s *BGDef) action() {
	s.bgct.stepBGDef(s)
	s.bga.action()
	if s.model != nil {
		s.model.step(1)
	}
	link := 0
	for i, b := range s.bg {
		s.bg[i].bga.action()
		if i > 0 && b.positionlink {
			s.bg[i].bga.offset[0] += s.bg[link].bga.sinoffset[0]
			s.bg[i].bga.offset[1] += s.bg[link].bga.sinoffset[1]
		} else {
			link = i
		}
		if b.active {
			s.bg[i].anim.Action()
		}
	}
}

func (s *BGDef) draw(layer int32, x, y, scl float32) {
	// Action will only happen once per frame even though this function is called more times
	// TODO: Doing this in layer 0 is currently necessary for it to work in screenpacks, but it might be introducing a frame of delay in layer -1
	if layer == 0 {
		s.action()
	}
	if s.model != nil && s.sceneNumber >= 0 {
		if layer == 0 {
			s.model.calculateTextureTransform()
		}
		drawFOV := s.fov * math.Pi / 180
		outlineConst := float32(0.003 * math.Tan(float64(drawFOV)))
		proj := mgl.Perspective(drawFOV, float32(sys.scrrect[2])/float32(sys.scrrect[3]), s.near, s.far)
		view := mgl.Translate3D(s.modelOffset[0], s.modelOffset[1], s.modelOffset[2])
		view = view.Mul4(mgl.Scale3D(s.modelScale[0], s.modelScale[1], s.modelScale[2]))
		s.model.draw(1, int(s.sceneNumber), int(layer), 0, s.modelOffset, proj, view, proj.Mul4(view), outlineConst)
	}
	//x, y = x/s.localscl, y/s.localscl
	for _, b := range s.bg {
		if b.layerno == layer && b.visible && b.anim.spr != nil {
			b.draw([...]float32{x, y}, scl, s.localscl, 1, s.scale, 0, false)
		}
	}
}

func (s *BGDef) reset() {
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
}
