package main

import (
	_ "embed" // Support for go:embed resources
	"encoding/binary"
	"math"

	gl "github.com/fyne-io/gl-js"
	mgl "github.com/go-gl/mathgl/mgl32"
	"golang.org/x/mobile/exp/f32"
)

// Rotation holds rotation parameters
type Rotation struct {
	angle, xangle, yangle float32
}

func (r *Rotation) IsZero() bool {
	return r.angle == 0 && r.xangle == 0 && r.yangle == 0
}

// Tiling holds tiling parameters
type Tiling struct {
	x, y, sx, sy int32
}

var notiling = Tiling{}

// RenderParams holds the common data for all sprite rendering functions
type RenderParams struct {
	// Sprite texture and palette texture
	tex    *Texture
	paltex *Texture
	// Size, position, tiling, scaling and rotation
	size     [2]uint16
	x, y     float32
	tile     Tiling
	xts, xbs float32
	ys, vs   float32
	rxadd    float32
	rot      Rotation
	// Transparency, masking and palette effects
	tint  uint32 // Sprite tint for shadows (unused yet)
	trans int32  // Mugen transparency blending
	mask  int32  // Mask for transparency
	pfx   *PalFX
	// Clipping
	window *[4]int32
	// Rotation center
	rcx, rcy float32
	// Perspective projection
	projectionMode int32
	fLength        float32
	xOffset        float32
	yOffset        float32
}

func (rp *RenderParams) IsValid() bool {
	return rp.tex.IsValid() && IsFinite(rp.x+rp.y+rp.xts+rp.xbs+rp.ys+rp.vs+
		rp.rxadd+rp.rot.angle+rp.rcx+rp.rcy)
}

// The global rendering backend
var renderer *Renderer

var vertexBuffer gl.Buffer

var mainShader, flatShader *ShaderProgram

//go:embed shaders/sprite.vert.glsl
var vertShader string

//go:embed shaders/sprite.frag.glsl
var fragShader string

//go:embed shaders/flat.frag.glsl
var fragShaderFlat string

// Render initialization.
// Creates the default shaders, the framebuffer and enables MSAA.
func RenderInit() {
	renderer = newRenderer()

	// Sprite shaders
	mainShader = newShaderProgram(vertShader, fragShader, "Main Shader")
	mainShader.RegisterUniforms("pal", "tint", "mask", "neg", "gray", "add", "mult", "x1x2x4x3", "isRgba", "isTrapez")

	flatShader = newShaderProgram(vertShader, fragShaderFlat, "Flat Shader")
	flatShader.RegisterUniforms("color", "isShadow")

	// Persistent data buffer for rendering
	vertexBuffer = gl.CreateBuffer()
}

func drawQuads(s *ShaderProgram, modelview mgl.Mat4, x1, y1, x2, y2, x3, y3, x4, y4 float32) {
	s.UniformMatrix("modelview", modelview[:])
	s.UniformF("x1x2x4x3", x1, x2, x4, x3) // this uniform is optional
	vertexPosition := f32.Bytes(binary.LittleEndian,
		x2, y2, 1, 1,
		x3, y3, 1, 0,
		x1, y1, 0, 1,
		x4, y4, 0, 0)
	gl.BindBuffer(gl.ARRAY_BUFFER, vertexBuffer)
	gl.BufferData(gl.ARRAY_BUFFER, vertexPosition, gl.STATIC_DRAW)

	gl.DrawArrays(gl.TRIANGLE_STRIP, 0, 4)
}

// Render a quad with optional horizontal tiling
func rmTileHSub(s *ShaderProgram, modelview mgl.Mat4, x1, y1, x2, y2, x3, y3, x4, y4, width float32,
	tl Tiling, rcx float32) {
	//            p3
	//    p4 o-----o-----o- - -o
	//      /      |      \     ` .
	//     /       |       \       `.
	//    o--------o--------o- - - - o
	//   p1         p2
	topdist := (x3 - x4) * (1 + float32(tl.sx) / width)
	botdist := (x2 - x1) * (1 + float32(tl.sx) / width)
	if AbsF(topdist) >= 0.01 {
		db := (x4 - rcx) * (botdist - topdist) / AbsF(topdist)
		x1 += db
		x2 += db
	}

	// Compute left/right tiling bounds (or right/left when topdist < 0)
	xmax := float32(sys.scrrect[2])
	left, right := int32(0), int32(1)
	if topdist >= 0.01 {
		left = 1 - int32(math.Ceil(float64(MaxF(x3 / topdist, x2 / botdist))))
		right = int32(math.Ceil(float64(MaxF((xmax - x4) / topdist, (xmax - x1) / botdist))))
	} else if topdist <= -0.01 {
		left = 1 - int32(math.Ceil(float64(MaxF((xmax - x3) / -topdist, (xmax - x2) / -botdist))))
		right = int32(math.Ceil(float64(MaxF(x4 / -topdist, x1 / -botdist))))
	}

	if tl.x != 1 {
		left = 0
		right = Min(right, Max(tl.x, 1))
	}

	// Draw all quads in one loop
	for n := left; n < right; n++ {
		x1d, x2d := x1 + float32(n) * botdist, x2 + float32(n) * botdist
		x3d, x4d := x3 + float32(n) * topdist, x4 + float32(n) * topdist
		drawQuads(s, modelview, x1d, y1, x2d, y2, x3d, y3, x4d, y4)
	}
}

func rmTileSub(s *ShaderProgram, modelview mgl.Mat4, rp RenderParams) {
	x1, y1 := rp.x+rp.rxadd*rp.ys*float32(rp.size[1]), rp.rcy+((rp.y-rp.ys*float32(rp.size[1]))-rp.rcy)*rp.vs
	x2, y2 := x1+rp.xbs*float32(rp.size[0]), y1
	x3, y3 := rp.x+rp.xts*float32(rp.size[0]), rp.rcy+(rp.y-rp.rcy)*rp.vs
	x4, y4 := rp.x, y3
	//var pers float32
	//if AbsF(rp.xts) < AbsF(rp.xbs) {
	//	pers = AbsF(rp.xts) / AbsF(rp.xbs)
	//} else {
	//	pers = AbsF(rp.xbs) / AbsF(rp.xts)
	//}
	if !rp.rot.IsZero() {
		//	kaiten(&x1, &y1, float64(agl), rcx, rcy, vs)
		//	kaiten(&x2, &y2, float64(agl), rcx, rcy, vs)
		//	kaiten(&x3, &y3, float64(agl), rcx, rcy, vs)
		//	kaiten(&x4, &y4, float64(agl), rcx, rcy, vs)
		if rp.vs != 1 {
			y1 = rp.rcy + ((rp.y - rp.ys*float32(rp.size[1])) - rp.rcy)
			y2 = y1
			y3 = rp.rcy + (rp.y - rp.rcy)
			y4 = y3
		}
		if rp.projectionMode == 0 {
			modelview = modelview.Mul4(mgl.Translate3D(rp.rcx, rp.rcy, 0))
		} else if rp.projectionMode == 1 {
			//This is the inverse of the orthographic projection matrix
			matrix := mgl.Mat4{float32(sys.scrrect[2] / 2.0), 0, 0, 0, 0, float32(sys.scrrect[3] / 2), 0, 0, 0, 0, -65535, 0, float32(sys.scrrect[2] / 2), float32(sys.scrrect[3] / 2), 0, 1}
			modelview = modelview.Mul4(mgl.Translate3D(0, -float32(sys.scrrect[3]), rp.fLength))
			modelview = modelview.Mul4(matrix)
			modelview = modelview.Mul4(mgl.Frustum(-float32(sys.scrrect[2])/2/rp.fLength, float32(sys.scrrect[2])/2/rp.fLength, -float32(sys.scrrect[3])/2/rp.fLength, float32(sys.scrrect[3])/2/rp.fLength, 1.0, 65535))
			modelview = modelview.Mul4(mgl.Translate3D(-float32(sys.scrrect[2])/2.0, float32(sys.scrrect[3])/2.0, -rp.fLength))
			modelview = modelview.Mul4(mgl.Translate3D(rp.rcx, rp.rcy, 0))
		} else if rp.projectionMode == 2 {
			matrix := mgl.Mat4{float32(sys.scrrect[2] / 2.0), 0, 0, 0, 0, float32(sys.scrrect[3] / 2), 0, 0, 0, 0, -65535, 0, float32(sys.scrrect[2] / 2), float32(sys.scrrect[3] / 2), 0, 1}
			//modelview = modelview.Mul4(mgl.Translate3D(0, -float32(sys.scrrect[3]), 2048))
			modelview = modelview.Mul4(mgl.Translate3D(rp.rcx-float32(sys.scrrect[2])/2.0-rp.xOffset, rp.rcy-float32(sys.scrrect[3])/2.0+rp.yOffset, rp.fLength))
			modelview = modelview.Mul4(matrix)
			modelview = modelview.Mul4(mgl.Frustum(-float32(sys.scrrect[2])/2/rp.fLength, float32(sys.scrrect[2])/2/rp.fLength, -float32(sys.scrrect[3])/2/rp.fLength, float32(sys.scrrect[3])/2/rp.fLength, 1.0, 65535))
			modelview = modelview.Mul4(mgl.Translate3D(rp.xOffset, -rp.yOffset, -rp.fLength))
		}

		modelview = modelview.Mul4(mgl.Scale3D(1, rp.vs, 1))
		modelview = modelview.Mul4(
			mgl.Rotate3DX(-rp.rot.xangle * math.Pi / 180.0).Mul3(
				mgl.Rotate3DY(rp.rot.yangle * math.Pi / 180.0)).Mul3(
				mgl.Rotate3DZ(rp.rot.angle * math.Pi / 180.0)).Mat4())
		modelview = modelview.Mul4(mgl.Translate3D(-rp.rcx, -rp.rcy, 0))

		drawQuads(s, modelview, x1, y1, x2, y2, x3, y3, x4, y4)
		return
	}
	if rp.tile.y == 1 && rp.xbs != 0 {
		x1d, y1d, x2d, y2d, x3d, y3d, x4d, y4d := x1, y1, x2, y2, x3, y3, x4, y4
		for {
			x1d, y1d = x4d, y4d+rp.ys*rp.vs*float32(rp.tile.sy)
			x2d, y2d = x3d, y1d
			x3d = x4d - rp.rxadd*rp.ys*float32(rp.size[1]) + (rp.xts/rp.xbs)*(x3d-x4d)
			y3d = y2d + rp.ys*rp.vs*float32(rp.size[1])
			x4d = x4d - rp.rxadd*rp.ys*float32(rp.size[1])
			if AbsF(y3d-y4d) < 0.01 {
				break
			}
			y4d = y3d
			if rp.ys*(float32(rp.size[1])+float32(rp.tile.sy)) < 0 {
				if y1d <= float32(-sys.scrrect[3]) && y4d <= float32(-sys.scrrect[3]) {
					break
				}
			} else if y1d >= 0 && y4d >= 0 {
				break
			}
			if (0 > y1d || 0 > y4d) &&
				(y1d > float32(-sys.scrrect[3]) || y4d > float32(-sys.scrrect[3])) {
				rmTileHSub(s, modelview, x1d, y1d, x2d, y2d, x3d, y3d, x4d, y4d,
					float32(rp.size[0]), rp.tile, rp.rcx)
			}
		}
	}
	if rp.tile.y == 0 || rp.xts != 0 {
		n := rp.tile.y
		for {
			if rp.ys*(float32(rp.size[1])+float32(rp.tile.sy)) > 0 {
				if y1 <= float32(-sys.scrrect[3]) && y4 <= float32(-sys.scrrect[3]) {
					break
				}
			} else if y1 >= 0 && y4 >= 0 {
				break
			}
			if (0 > y1 || 0 > y4) &&
				(y1 > float32(-sys.scrrect[3]) || y4 > float32(-sys.scrrect[3])) {
				rmTileHSub(s, modelview, x1, y1, x2, y2, x3, y3, x4, y4,
					float32(rp.size[0]), rp.tile, rp.rcx)
			}
			if rp.tile.y != 1 && n != 0 {
				n--
			}
			if n == 0 {
				break
			}
			x4, y4 = x1, y1-rp.ys*rp.vs*float32(rp.tile.sy)
			x3, y3 = x2, y4
			x2 = x1 + rp.rxadd*rp.ys*float32(rp.size[1]) + (rp.xbs/rp.xts)*(x2-x1)
			y2 = y3 - rp.ys*rp.vs*float32(rp.size[1])
			x1 = x1 + rp.rxadd*rp.ys*float32(rp.size[1])
			if AbsF(y1-y2) < 0.01 {
				break
			}
			y1 = y2
		}
	}
}
func rmMainSub(s *ShaderProgram, rp RenderParams) {
	proj := mgl.Ortho(0, float32(sys.scrrect[2]), 0, float32(sys.scrrect[3]), -65535, 65535)
	s.UniformMatrix("projection", proj[:])

	modelview := mgl.Translate3D(0, float32(sys.scrrect[3]), 0)

	// Must bind buffer before enabling attributes
	gl.BindBuffer(gl.ARRAY_BUFFER, vertexBuffer)
	s.EnableAttribs()

	renderWithBlending(func(a float32) {
		s.UniformF("alpha", a)
		rmTileSub(s, modelview, rp)
	}, rp.trans, rp.paltex != nil)

	s.DisableAttribs()
	gl.Disable(gl.SCISSOR_TEST)
}

func rmInitSub(rp *RenderParams) {
	if rp.vs < 0 {
		rp.vs *= -1
		rp.ys *= -1
		rp.rot.angle *= -1
		rp.rot.xangle *= -1
	}
	if rp.tile.x == 0 {
		rp.tile.sx = 0
	} else if rp.tile.sx > 0 {
		rp.tile.sx -= int32(rp.size[0])
	}
	if rp.tile.y == 0 {
		rp.tile.sy = 0
	} else if rp.tile.sy > 0 {
		rp.tile.sy -= int32(rp.size[1])
	}
	if rp.xts >= 0 {
		rp.x *= -1
	}
	rp.x += rp.rcx
	rp.rcy *= -1
	if rp.ys < 0 {
		rp.y *= -1
	}
	rp.y += rp.rcy
	gl.Enable(gl.SCISSOR_TEST)
	gl.Scissor(rp.window[0], sys.scrrect[3]-(rp.window[1]+rp.window[3]),
		rp.window[2], rp.window[3])
}

func RenderSprite(rp RenderParams) {
	if !rp.IsValid() {
		return
	}
	rmInitSub(&rp)
	mainShader.UseProgram()
	mainShader.UniformI("tex", 0)
	if rp.paltex == nil {
		mainShader.UniformI("isRgba", 1)
	} else {
		rp.paltex.Bind(1)
		mainShader.UniformI("pal", 1)
		mainShader.UniformI("isRgba", 0)
		mainShader.UniformI("mask", int(rp.mask))
	}
	mainShader.UniformI("isTrapez", int(Btoi(AbsF(AbsF(rp.xts)-AbsF(rp.xbs)) > 0.001)))

	neg, grayscale, padd, pmul := false, float32(0), [3]float32{0, 0, 0}, [3]float32{1, 1, 1}
	if rp.pfx != nil {
		neg, grayscale, padd, pmul = rp.pfx.getFcPalFx(rp.trans == -2)
		if rp.trans == -2 {
			padd[0], padd[1], padd[2] = -padd[0], -padd[1], -padd[2]
		}
	}
	mainShader.UniformI("neg", int(Btoi(neg)))
	mainShader.UniformF("gray", grayscale)
	mainShader.UniformFv("add", padd[:])
	mainShader.UniformFv("mult", pmul[:])

	rp.tex.Bind(0)
	rmMainSub(mainShader, rp)
}

func RenderFlatSprite(rp RenderParams, color uint32) {
	if !rp.IsValid() {
		return
	}
	rmInitSub(&rp)
	flatShader.UseProgram()
	rp.tex.Bind(0)
	flatShader.UniformI("tex", 0)
	flatShader.UniformF("color",
		float32(color>>16&0xff)/255, float32(color>>8&0xff)/255, float32(color&0xff)/255)
	flatShader.UniformI("isShadow", 1)

	rmMainSub(flatShader, rp)
}

func renderWithBlending(render func(a float32), trans int32, correctAlpha bool) {
	var blendSourceFactor gl.Enum = gl.SRC_ALPHA
	if !correctAlpha {
		blendSourceFactor = gl.ONE
	}
	gl.Enable(gl.BLEND)
	switch {
	case trans == -1:
		gl.BlendFunc(blendSourceFactor, gl.ONE)
		gl.BlendEquation(gl.FUNC_ADD)
		render(1)
	case trans == -2:
		gl.BlendFunc(gl.ONE, gl.ONE)
		gl.BlendEquation(gl.FUNC_REVERSE_SUBTRACT)
		render(1)
	case trans <= 0:
	case trans < 255:
		gl.BlendFunc(blendSourceFactor, gl.ONE_MINUS_SRC_ALPHA)
		gl.BlendEquation(gl.FUNC_ADD)
		render(float32(trans) / 255)
	case trans < 512:
		gl.BlendFunc(blendSourceFactor, gl.ONE_MINUS_SRC_ALPHA)
		gl.BlendEquation(gl.FUNC_ADD)
		render(1)
	default:
		src, dst := trans&0xff, trans>>10&0xff
		if dst < 255 {
			gl.BlendFunc(gl.ZERO, gl.ONE_MINUS_SRC_ALPHA)
			gl.BlendEquation(gl.FUNC_ADD)
			render(1 - float32(dst)/255)
		}
		if src > 0 {
			gl.BlendFunc(blendSourceFactor, gl.ONE)
			gl.BlendEquation(gl.FUNC_ADD)
			render(float32(src) / 255)
		}
	}
	gl.Disable(gl.BLEND)
}

func FillRect(rect [4]int32, color uint32, trans int32) {
	r := float32(color>>16&0xff) / 255
	g := float32(color>>8&0xff) / 255
	b := float32(color&0xff) / 255

	modelview := mgl.Translate3D(0, float32(sys.scrrect[3]), 0)
	proj := mgl.Ortho(0, float32(sys.scrrect[2]), 0, float32(sys.scrrect[3]), -65535, 65535)

	x1, y1 := float32(rect[0]), -float32(rect[1])
	x2, y2 := float32(rect[0]+rect[2]), -float32(rect[1]+rect[3])
	vertexPosition := f32.Bytes(binary.LittleEndian,
		x2, y2, 1, 1,
		x2, y1, 1, 0,
		x1, y2, 0, 1,
		x1, y1, 0, 0)
	gl.BindBuffer(gl.ARRAY_BUFFER, vertexBuffer)
	gl.BufferData(gl.ARRAY_BUFFER, vertexPosition, gl.STATIC_DRAW)

	flatShader.UseProgram()
	flatShader.UniformMatrix("modelview", modelview[:])
	flatShader.UniformMatrix("projection", proj[:])
	flatShader.UniformI("tex", 0)
	flatShader.UniformF("color", r, g, b)
	flatShader.UniformI("isShadow", 0)
	flatShader.EnableAttribs()

	renderWithBlending(func(a float32) {
		flatShader.UniformF("alpha", a)
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, 4)
	}, trans, true)

	flatShader.DisableAttribs()
}
