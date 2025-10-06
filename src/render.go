package main

import (
	_ "embed"
	"math"

	mgl "github.com/go-gl/mathgl/mgl32"
	"github.com/ikemen-engine/glfont"
)

type Texture interface {
	SetData(data []byte)
	SetDataG(data []byte, mag, min, ws, wt int32)
	SetPixelData(data []float32)
	SetRGBPixelData(data []float32)
	IsValid() bool
	GetWidth() int32
	GetHeight() int32
}

type Renderer interface {
	GetName() string
	Init()
	Close()
	BeginFrame(clearColor bool)
	EndFrame()
	Await()

	IsModelEnabled() bool
	IsShadowEnabled() bool

	BlendReset()
	SetPipeline(eq BlendEquation, src, dst BlendFunc)
	ReleasePipeline()
	prepareShadowMapPipeline(bufferIndex uint32)
	setShadowMapPipeline(doubleSided, invertFrontFace, useUV, useNormal, useTangent, useVertColor, useJoint0, useJoint1 bool, numVertices, vertAttrOffset uint32)
	ReleaseShadowPipeline()
	prepareModelPipeline(bufferIndex uint32, env *Environment)
	SetModelPipeline(eq BlendEquation, src, dst BlendFunc, depthTest, depthMask, doubleSided, invertFrontFace, useUV, useNormal, useTangent, useVertColor, useJoint0, useJoint1, useOutlineAttribute bool, numVertices, vertAttrOffset uint32)
	SetMeshOulinePipeline(invertFrontFace bool, meshOutline float32)
	ReleaseModelPipeline()

	newTexture(width, height, depth int32, filter bool) (t Texture)
	newDataTexture(width, height int32) (t Texture)
	newHDRTexture(width, height int32) (t Texture)
	newCubeMapTexture(widthHeight int32, mipmap bool) (t Texture)

	ReadPixels(data []uint8, width, height int)
	Scissor(x, y, width, height int32)
	DisableScissor()

	SetUniformI(name string, val int)
	SetUniformF(name string, values ...float32)
	SetUniformFv(name string, values []float32)
	SetUniformMatrix(name string, value []float32)
	SetTexture(name string, tex Texture)
	SetModelUniformI(name string, val int)
	SetModelUniformF(name string, values ...float32)
	SetModelUniformFv(name string, values []float32)
	SetModelUniformMatrix(name string, value []float32)
	SetModelUniformMatrix3(name string, value []float32)
	SetModelTexture(name string, t Texture)
	SetShadowMapUniformI(name string, val int)
	SetShadowMapUniformF(name string, values ...float32)
	SetShadowMapUniformFv(name string, values []float32)
	SetShadowMapUniformMatrix(name string, value []float32)
	SetShadowMapTexture(name string, t Texture)
	SetShadowFrameTexture(i uint32)
	SetShadowFrameCubeTexture(i uint32)
	SetVertexData(values ...float32)
	SetModelVertexData(bufferIndex uint32, values []byte)
	SetModelIndexData(bufferIndex uint32, values ...uint32)

	RenderQuad()
	RenderElements(mode PrimitiveMode, count, offset int)
	RenderCubeMap(envTexture Texture, cubeTexture Texture)
	RenderFilteredCubeMap(distribution int32, cubeTexture Texture, filteredTexture Texture, mipmapLevel, sampleCount int32, roughness float32)
	RenderLUT(distribution int32, cubeTexture Texture, lutTexture Texture, sampleCount int32)
}

//go:embed shaders/sprite.vert.glsl
var vertShader string

//go:embed shaders/sprite.frag.glsl
var fragShader string

//go:embed shaders/model.vert.glsl
var modelVertShader string

//go:embed shaders/model.frag.glsl
var modelFragShader string

//go:embed shaders/shadow.vert.glsl
var shadowVertShader string

//go:embed shaders/shadow.frag.glsl
var shadowFragShader string

//go:embed shaders/shadow.geo.glsl
var shadowGeoShader string

//go:embed shaders/ident.vert.glsl
var identVertShader string

//go:embed shaders/ident.frag.glsl
var identFragShader string

//go:embed shaders/panoramaToCubeMap.frag.glsl
var panoramaToCubeMapFragShader string

//go:embed shaders/cubemapFiltering.frag.glsl
var cubemapFilteringFragShader string

// The global, platform-specific rendering backend
var gfx Renderer
var gfxFont glfont.FontRenderer

// Blend constants
type BlendFunc int

const (
	BlendOne = BlendFunc(iota)
	BlendZero
	BlendSrcAlpha
	BlendOneMinusSrcAlpha
	BlendDstColor
	BlendOneMinusDstColor
)

type BlendEquation int

const (
	BlendAdd = BlendEquation(iota)
	BlendReverseSubtract
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
	xflag, yflag       int32
	xspacing, yspacing int32
}

var notiling = Tiling{}

// RenderParams holds the common data for all sprite rendering functions
type RenderParams struct {
	tex            Texture // Sprite
	paltex         Texture // Palette
	size           [2]uint16
	x, y           float32 // Position
	tile           Tiling
	xts, xbs       float32 // Top and bottom X scale (as in parallax)
	ys, vs         float32 // Y scale
	rxadd          float32
	xas, yas       float32
	rot            Rotation
	tint           uint32 // Sprite tint for shadows
	blendMode      TransType
	blendAlpha     [2]int32
	mask           int32 // Mask for transparency
	pfx            *PalFX
	window         *[4]int32
	rcx, rcy       float32 // Rotation center
	projectionMode int32   // Perspective projection
	fLength        float32 // Focal length
	xOffset        float32
	yOffset        float32
}

func (rp *RenderParams) IsValid() bool {
	return rp.tex.IsValid() && rp.size[0] != 0 && rp.size[1] != 0 &&
		IsFinite(rp.x+rp.y+rp.xts+rp.xbs+rp.ys+rp.vs+rp.rxadd+rp.rot.angle+rp.rcx+rp.rcy)
}

func drawQuads(modelview mgl.Mat4, x1, y1, x2, y2, x3, y3, x4, y4 float32) {
	gfx.SetUniformMatrix("modelview", modelview[:])
	gfx.SetUniformF("x1x2x4x3", x1, x2, x4, x3) // this uniform is optional
	gfx.SetVertexData(
		x2, y2, 1, 1,
		x3, y3, 1, 0,
		x1, y1, 0, 1,
		x4, y4, 0, 0,
	)

	gfx.RenderQuad()
}

// Render a quad with optional horizontal tiling
func rmTileHSub(modelview mgl.Mat4, x1, y1, x2, y2, x3, y3, x4, y4, dy, width float32, rp RenderParams) {
	//            p3
	//    p4 o-----o-----o- - -o
	//      /      |      \     ` .
	//     /       |       \       `.
	//    o--------o--------o- - - - o
	//   p1         p2
	topdist := (x3 - x4) * (((float32(rp.tile.xspacing) + width) / rp.xas) / width)
	botdist := (x2 - x1) * (((float32(rp.tile.xspacing) + width) / rp.xas) / width)
	if AbsF(topdist) >= 0.01 {
		db := (x4 - rp.rcx) * (botdist - topdist) / AbsF(topdist)
		x1 += db
		x2 += db
	}

	// Compute left/right tiling bounds (or right/left when topdist < 0)
	xmax := float32(sys.scrrect[2])
	left, right := int32(0), int32(1)
	if rp.tile.xflag != 0 {
		if topdist >= 0.01 {
			if x1 > x2 {
				left = 1 - int32(math.Ceil(float64(MaxF(x4/topdist, x1/botdist))))
				right = int32(math.Ceil(float64(MaxF((xmax-x3)/topdist, (xmax-x2)/botdist))))
			} else {
				left = 1 - int32(math.Ceil(float64(MaxF(x3/topdist, x2/botdist))))
				right = int32(math.Ceil(float64(MaxF((xmax-x4)/topdist, (xmax-x1)/botdist))))
			}
		} else if topdist <= -0.01 {
			if x1 > x2 {
				left = 1 - int32(math.Ceil(float64(MaxF((xmax-x3)/-topdist, (xmax-x2)/-botdist))))
				right = int32(math.Ceil(float64(MaxF(x4/-topdist, x1/-botdist))))
			} else {
				left = 1 - int32(math.Ceil(float64(MaxF((xmax-x4)/-topdist, (xmax-x1)/-botdist))))
				right = int32(math.Ceil(float64(MaxF(x3/-topdist, x2/-botdist))))
			}
		}
		if rp.tile.xflag != 1 {
			left = 0
			right = Min(right, Max(rp.tile.xflag, 1))
		}
	}

	// Draw all quads in one loop
	for n := left; n < right; n++ {
		x1d, x2d := x1+float32(n)*botdist, x2+float32(n)*botdist
		x3d, x4d := x3+float32(n)*topdist, x4+float32(n)*topdist
		mat := modelview
		if !rp.rot.IsZero() {
			mat = mat.Mul4(mgl.Translate3D(rp.rcx+float32(n)*botdist, rp.rcy+dy, 0))
			//modelview = modelview.Mul4(mgl.Scale3D(1, rp.vs, 1))
			mat = mat.Mul4(mgl.Rotate3DZ(rp.rot.angle * math.Pi / 180.0).Mat4())
			mat = mat.Mul4(mgl.Translate3D(-(rp.rcx + float32(n)*botdist), -(rp.rcy + dy), 0))
		}

		drawQuads(mat, x1d, y1, x2d, y2, x3d, y3, x4d, y4)
	}
}

func rmTileSub(modelview mgl.Mat4, rp RenderParams) {
	x1, y1 := rp.x, rp.rcy+((rp.y-rp.ys*float32(rp.size[1]))-rp.rcy)*rp.vs
	x2, y2 := x1+rp.xbs*float32(rp.size[0]), y1
	x3, y3 := rp.x+rp.xts*float32(rp.size[0]), rp.rcy+(rp.y-rp.rcy)*rp.vs
	x4, y4 := rp.x, y3
	//var pers float32
	//if AbsF(rp.xts) < AbsF(rp.xbs) {
	//	pers = AbsF(rp.xts) / AbsF(rp.xbs)
	//} else {
	//	pers = AbsF(rp.xbs) / AbsF(rp.xts)
	//}
	if !rp.rot.IsZero() && rp.tile.xflag == 0 && rp.tile.yflag == 0 {

		if rp.vs != 1 {
			y1 = rp.rcy + ((rp.y - rp.ys*float32(rp.size[1])) - rp.rcy)
			y2 = y1
			y3 = rp.y
			y4 = y3
		}
		if rp.projectionMode == 0 {
			modelview = modelview.Mul4(mgl.Translate3D(rp.rcx, rp.rcy, 0))
		} else if rp.projectionMode == 1 {
			// This is the inverse of the orthographic projection matrix
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

		// Apply shear matrix before rotation
		shearMatrix := mgl.Mat4{
			1, 0, 0, 0,
			rp.rxadd, 1, 0, 0,
			0, 0, 1, 0,
			0, 0, 0, 1}
		modelview = modelview.Mul4(shearMatrix)
		modelview = modelview.Mul4(mgl.Translate3D(rp.rxadd*rp.ys*float32(rp.size[1]), 0, 0))

		modelview = modelview.Mul4(mgl.Scale3D(1, rp.vs, 1))
		modelview = modelview.Mul4(
			mgl.Rotate3DX(-rp.rot.xangle * math.Pi / 180.0).Mul3(
				mgl.Rotate3DY(rp.rot.yangle * math.Pi / 180.0)).Mul3(
				mgl.Rotate3DZ(rp.rot.angle * math.Pi / 180.0)).Mat4())
		modelview = modelview.Mul4(mgl.Translate3D(-rp.rcx, -rp.rcy, 0))

		drawQuads(modelview, x1, y1, x2, y2, x3, y3, x4, y4)
		return
	}
	if rp.tile.yflag == 1 && rp.xbs != 0 {
		x1 += rp.rxadd * rp.ys * float32(rp.size[1])
		x2 = x1 + rp.xbs*float32(rp.size[0])
		x1d, y1d, x2d, y2d, x3d, y3d, x4d, y4d := x1, y1, x2, y2, x3, y3, x4, y4
		n := 0
		var xy []float32
		for {
			x1d, y1d = x4d, y4d+rp.ys*rp.vs*((float32(rp.tile.yspacing)+float32(rp.size[1]))/rp.yas-float32(rp.size[1]))
			x2d, y2d = x3d, y1d
			x3d = x4d - rp.rxadd*rp.ys*float32(rp.size[1]) + (rp.xts/rp.xbs)*(x3d-x4d)
			y3d = y2d + rp.ys*rp.vs*float32(rp.size[1])
			x4d = x4d - rp.rxadd*rp.ys*float32(rp.size[1])
			if AbsF(y3d-y4d) < 0.01 {
				break
			}
			y4d = y3d
			if rp.ys*((float32(rp.tile.yspacing)+float32(rp.size[1]))/rp.yas) < 0 {
				if y1d <= float32(-sys.scrrect[3]) && y4d <= float32(-sys.scrrect[3]) {
					break
				}
			} else if y1d >= 0 && y4d >= 0 {
				break
			}
			n += 1
			xy = append(xy, x1d, x2d, x3d, x4d, y1d, y2d, y3d, y4d)
		}
		for {
			if len(xy) == 0 {
				break
			}
			x1d, x2d, x3d, x4d, y1d, y2d, y3d, y4d, xy = xy[len(xy)-8], xy[len(xy)-7], xy[len(xy)-6], xy[len(xy)-5], xy[len(xy)-4], xy[len(xy)-3], xy[len(xy)-2], xy[len(xy)-1], xy[:len(xy)-8]
			if (0 > y1d || 0 > y4d) &&
				(y1d > float32(-sys.scrrect[3]) || y4d > float32(-sys.scrrect[3])) {
				rmTileHSub(modelview, x1d, y1d, x2d, y2d, x3d, y3d, x4d, y4d, y1d-y1, float32(rp.size[0]), rp)
			}
		}
	}
	if rp.tile.yflag == 0 || rp.xts != 0 {
		x1 += rp.rxadd * rp.ys * float32(rp.size[1])
		x2 = x1 + rp.xbs*float32(rp.size[0])
		n := rp.tile.yflag
		oy := y1
		for {
			if rp.ys*((float32(rp.tile.yspacing)+float32(rp.size[1]))/rp.yas) > 0 {
				if y1 <= float32(-sys.scrrect[3]) && y4 <= float32(-sys.scrrect[3]) {
					break
				}
			} else if y1 >= 0 && y4 >= 0 {
				break
			}
			if (0 > y1 || 0 > y4) &&
				(y1 > float32(-sys.scrrect[3]) || y4 > float32(-sys.scrrect[3])) {
				rmTileHSub(modelview, x1, y1, x2, y2, x3, y3, x4, y4, y1-oy,
					float32(rp.size[0]), rp)
			}
			if rp.tile.yflag != 1 && n != 0 {
				n--
			}
			if n == 0 {
				break
			}
			x4, y4 = x1, y1-rp.ys*rp.vs*((float32(rp.tile.yspacing)+float32(rp.size[1]))/rp.yas-float32(rp.size[1]))
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

func rmInitSub(rp *RenderParams) {
	if rp.vs < 0 {
		rp.vs *= -1
		rp.ys *= -1
		rp.rot.angle *= -1
		rp.rot.xangle *= -1
	}
	if rp.tile.xflag == 0 {
		rp.tile.xspacing = 0
	} else if rp.tile.xspacing > 0 {
		rp.tile.xspacing -= int32(rp.size[0])
	}
	if rp.tile.yflag == 0 {
		rp.tile.yspacing = 0
	} else if rp.tile.yspacing > 0 {
		rp.tile.yspacing -= int32(rp.size[1])
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
}

func BlendReset() {
	gfx.BlendReset()
}

func RenderSprite(rp RenderParams) {
	if !rp.IsValid() {
		return
	}

	rmInitSub(&rp)

	neg, grayscale, padd, pmul, invblend, hue := false, float32(0), [3]float32{0, 0, 0}, [3]float32{1, 1, 1}, int32(0), float32(0)
	tint := [4]float32{float32(rp.tint&0xff) / 255, float32(rp.tint>>8&0xff) / 255,
		float32(rp.tint>>16&0xff) / 255, float32(rp.tint>>24&0xff) / 255}

	if rp.pfx != nil {
		neg, grayscale, padd, pmul, invblend, hue = rp.pfx.getFcPalFx(false, rp.blendAlpha)
	}

	proj := mgl.Ortho(0, float32(sys.scrrect[2]), 0, float32(sys.scrrect[3]), -65535, 65535)
	modelview := mgl.Translate3D(0, float32(sys.scrrect[3]), 0)

	gfx.Scissor(rp.window[0], rp.window[1], rp.window[2], rp.window[3])

	render := func(eq BlendEquation, src, dst BlendFunc, a float32) {
		gfx.SetPipeline(eq, src, dst)

		gfx.SetUniformMatrix("projection", proj[:])
		gfx.SetTexture("tex", rp.tex)
		if rp.paltex == nil {
			gfx.SetUniformI("isRgba", 1)
		} else {
			gfx.SetTexture("pal", rp.paltex)
			gfx.SetUniformI("isRgba", 0)
		}
		gfx.SetUniformI("mask", int(rp.mask))
		gfx.SetUniformI("isTrapez", int(Btoi(AbsF(AbsF(rp.xts)-AbsF(rp.xbs)) > 0.001)))
		gfx.SetUniformI("isFlat", 0)

		gfx.SetUniformI("neg", int(Btoi(neg)))
		gfx.SetUniformF("gray", grayscale)
		gfx.SetUniformF("hue", hue)
		gfx.SetUniformFv("add", padd[:])
		gfx.SetUniformFv("mult", pmul[:])
		gfx.SetUniformFv("tint", tint[:])
		gfx.SetUniformF("alpha", a)

		rmTileSub(modelview, rp)

		gfx.ReleasePipeline()
	}

	renderWithBlending(render, rp.blendMode, rp.blendAlpha, rp.paltex != nil, invblend, &neg, &padd, &pmul, rp.paltex == nil)
	gfx.DisableScissor()
}

func renderWithBlending(
	render func(eq BlendEquation, src, dst BlendFunc, a float32),
	blendMode TransType, blendAlpha [2]int32, correctAlpha bool, invblend int32, neg *bool, acolor *[3]float32, mcolor *[3]float32, isrgba bool) {

	blendSourceFactor := BlendSrcAlpha
	if !correctAlpha {
		blendSourceFactor = BlendOne
	}

	Blend := BlendAdd
	BlendI := BlendReverseSubtract
	if invblend >= 1 {
		Blend = BlendReverseSubtract
		BlendI = BlendAdd
	}

	src := blendAlpha[0]
	dst := blendAlpha[1]

	// Ensure proper source and destination
	// TODO: Maybe use byte everywhere
	src = Clamp(src, 0, 255)
	dst = Clamp(dst, 0, 255)

	// Force None destination to 0 just in case
	if blendMode == TT_none {
		dst = 0
	}

	switch {
	// Sub
	case blendMode == TT_sub:
		if src == 0 && dst == 255 {
			// Fully transparent. Skip render
		} else if src == 255 && dst == 255 {
			// Fast path for full subtraction
			if invblend >= 1 && acolor != nil {
				(*acolor)[0], (*acolor)[1], (*acolor)[2] = -acolor[0], -acolor[1], -acolor[2]
			}
			if invblend == 3 && neg != nil {
				*neg = false
			}
			render(BlendI, blendSourceFactor, BlendOne, 1)
		} else {
			// Full alpha range
			if dst < 255 {
				render(BlendAdd, BlendZero, BlendOneMinusSrcAlpha, 1-float32(dst)/255)
			}
			if src > 0 {
				if invblend >= 1 && acolor != nil {
					(*acolor)[0], (*acolor)[1], (*acolor)[2] = -acolor[0], -acolor[1], -acolor[2]
				}
				if invblend == 3 && neg != nil {
					*neg = false
				}
				render(BlendI, blendSourceFactor, BlendOne, float32(src)/255)
			}
		}
	// Add, None or Default
	// None takes this path because SuperPause darkens sprites through their source alpha
	// Default should normally not reach here, so this is only a fallback
	default:
		if src == 0 && dst == 255 {
			// Fully transparent. Just don't render
		} else if src == 255 && dst == 0 {
			// Fast path for fully opaque
			render(BlendAdd, blendSourceFactor, BlendOneMinusSrcAlpha, 1)
		} else if src == 255 && dst == 255 {
			// Fast path for full Add
			if invblend >= 1 && acolor != nil {
				(*acolor)[0], (*acolor)[1], (*acolor)[2] = -acolor[0], -acolor[1], -acolor[2]
			}
			if invblend == 3 && neg != nil {
				*neg = false
			}
			render(Blend, blendSourceFactor, BlendOne, 1)
		} else {
			// AddAlpha (includes Add1)
			if dst < 255 {
				render(Blend, BlendZero, BlendOneMinusSrcAlpha, 1-float32(dst)/255)
			}
			if src > 0 {
				if invblend >= 1 && dst >= 255 {
					if invblend >= 2 {
						if invblend == 3 && neg != nil {
							*neg = false
						}
						if acolor != nil {
							(*acolor)[0], (*acolor)[1], (*acolor)[2] = -acolor[0], -acolor[1], -acolor[2]
						}
					}
					Blend = BlendReverseSubtract
				} else {
					Blend = BlendAdd
				}
				if !isrgba && (invblend >= 2 || invblend <= -1) && acolor != nil && mcolor != nil && src < 255 {
					// Sum of add components
					gc := AbsF(acolor[0]) + AbsF(acolor[1]) + AbsF(acolor[2])
					v3, ml, al := MaxF((gc*255)-float32(dst+src), 512)/128, (float32(src) / 255), (float32(src+dst) / 255)
					rM, gM, bM := mcolor[0]*ml, mcolor[1]*ml, mcolor[2]*ml
					(*mcolor)[0], (*mcolor)[1], (*mcolor)[2] = rM, gM, bM
					render(Blend, blendSourceFactor, BlendOne, al*Pow(v3, 3))
				} else {
					render(Blend, blendSourceFactor, BlendOne, float32(src)/255)
				}
			}
		}

	}
}

func FillRect(rect [4]int32, color uint32, alpha [2]int32) {
	r := float32(color>>16&0xff) / 255
	g := float32(color>>8&0xff) / 255
	b := float32(color&0xff) / 255

	modelview := mgl.Translate3D(0, float32(sys.scrrect[3]), 0)
	proj := mgl.Ortho(0, float32(sys.scrrect[2]), 0, float32(sys.scrrect[3]), -65535, 65535)

	x1, y1 := float32(rect[0]), -float32(rect[1])
	x2, y2 := float32(rect[0]+rect[2]), -float32(rect[1]+rect[3])

	render := func(eq BlendEquation, src, dst BlendFunc, a float32) {
		gfx.SetPipeline(eq, src, dst)
		gfx.SetVertexData(
			x2, y2, 1, 1,
			x2, y1, 1, 0,
			x1, y2, 0, 1,
			x1, y1, 0, 0,
		)
		gfx.SetUniformMatrix("modelview", modelview[:])
		gfx.SetUniformMatrix("projection", proj[:])
		gfx.SetUniformI("isFlat", 1)
		gfx.SetUniformF("tint", r, g, b, a)
		gfx.RenderQuad()
		gfx.ReleasePipeline()
	}

	renderWithBlending(render, TT_add, alpha, true, 0, nil, nil, nil, false)
}
