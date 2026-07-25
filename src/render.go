package main

import (
	"container/list"
	_ "embed"
	"math"

	mgl "github.com/go-gl/mathgl/mgl32"
)

type Texture interface {
	SetData(data []byte)
	SetSubData(data []byte, x, y, width, height, stride int32)
	SetDataG(data []byte, mag, min, ws, wt TextureSamplingParam)
	SetPixelData(data []float32)
	IsValid() bool
	GetWidth() int32
	GetHeight() int32
	CopyData(src *Texture)
}

type Renderer interface {
	GetName() string
	DebugInfo() string
	Init()
	Close()
	BeginFrame(clearColor bool)
	EndFrame()
	Await()

	IsModelEnabled() bool
	IsShadowEnabled() bool

	//SetPipeline()
	LoadCustomSpriteShader(shaderName string, shaderData []byte) uint32
	UnloadCustomSpriteShader(shaderName string)
	SetSpritePipeline(shaderName string)
	SetCustomUniforms(params [16]float32)
	NeedsGrabPass() bool
	ResolveBackBuffer() Texture

	EnableBlending(eq BlendEquation, src, dst BlendFunc)
	DisableBlending()

	prepareShadowMapPipeline(bufferIndex uint32)
	setShadowMapPipeline(doubleSided, invertFrontFace, useUV, useNormal, useTangent, useVertColor, useJoint0, useJoint1 bool, numVertices, vertAttrOffset uint32)
	ReleaseShadowPipeline()
	prepareModelPipeline(bufferIndex uint32, env *Environment)
	SetModelPipeline(eq BlendEquation, src, dst BlendFunc, depthTest, depthMask, doubleSided, invertFrontFace, useUV, useNormal, useTangent, useVertColor, useJoint0, useJoint1, useOutlineAttribute bool, numVertices, vertAttrOffset uint32)
	SetMeshOutlinePipeline(invertFrontFace bool, meshOutline float32)
	ReleaseModelPipeline()
	newTexture(width, height, depth int32, filter bool) (t Texture)
	newPaletteTexture() (t Texture)
	newModelTexture(width, height, depth int32, filter bool) (t Texture)
	newDataTexture(width, height int32) (t Texture)
	newHDRTexture(width, height int32) (t Texture)
	newCubeMapTexture(widthHeight int32, mipmap bool, lowestMipLevel int32) (t Texture)

	ReadPixels(data []uint8, width, height int)
	EnableScissor(x, y, width, height int32)
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
	SetShadowMapUniformMatrix3(name string, value []float32)
	SetShadowMapTexture(name string, t Texture)
	SetShadowFrameTexture(i uint32)
	SetShadowFrameCubeTexture(i uint32)
	SetVertexData(values ...float32)
	SetModelVertexData(bufferIndex uint32, values []byte)
	SetModelIndexData(bufferIndex uint32, values ...uint32)

	RenderQuad()
	RenderElements(mode PrimitiveMode, count, offset int)
	RenderShadowMapElements(mode PrimitiveMode, count, offset int)
	RenderCubeMap(envTexture Texture, cubeTexture Texture)
	RenderFilteredCubeMap(distribution int32, cubeTexture Texture, filteredTexture Texture, mipmapLevel, sampleCount int32, roughness float32)
	RenderLUT(distribution int32, cubeTexture Texture, lutTexture Texture, sampleCount int32)

	PerspectiveProjectionMatrix(angle, aspect, near, far float32) mgl.Mat4
	OrthographicProjectionMatrix(left, right, bottom, top, near, far float32) mgl.Mat4

	SetVSync(interval int)
	NewWorkerThread() bool
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

//go:embed shaders/font.frag.glsl
var fragmentFontShader string

//go:embed shaders/font.vert.glsl
var vertexFontShader string

// The global, platform-specific rendering backend
var gfx Renderer
var gfxFont FontRenderer

// Counter for unique texture cache serial numbers
var textureSerialNumber uint64

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

type TextureSamplingParam int

const (
	TextureSamplingFilterNearest = TextureSamplingParam(iota)
	TextureSamplingFilterLinear
	TextureSamplingFilterNearestMipMapNearest
	TextureSamplingFilterLinearMipMapNearest
	TextureSamplingFilterNearestMipMapLinear
	TextureSamplingFilterLinearMipMapLinear
	TextureSamplingWrapClampToEdge
	TextureSamplingWrapMirroredRepeat
	TextureSamplingWrapRepeat
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
	shader         string
	customShader   CustomShaderRenderData
}

type ShaderTexture struct {
	AnimNo int32
	Anim   *Animation
	SprNo  [2]int32
	Spr    *Sprite
}

func (st *ShaderTexture) clear() {
	st.AnimNo = -1
	st.Anim = nil
	st.SprNo = [2]int32{-1, -1}
	st.Spr = nil
}

func (st *ShaderTexture) GetTexture() Texture {
	if st.Anim != nil {
		if st.Anim.spr != nil {
			return st.Anim.spr.Tex
		}
	} else if st.Spr != nil {
		return st.Spr.Tex
	}
	return nil
}

func (st *ShaderTexture) step() {
	if st.Anim != nil {
		st.Anim.Action()
		st.Anim.UpdateSprite()
	}
}

type CustomShader struct {
	name   string
	params [16]float32
	time   int32
	sTime  float32
	tex1   ShaderTexture
	tex2   ShaderTexture
}

func (cs *CustomShader) clear() {
	cs.name = ""
	cs.params = [16]float32{}
	cs.time = 0
	cs.sTime = 0
	cs.tex1.clear()
	cs.tex2.clear()
}

type CustomShaderRenderData struct {
	name   string
	params [16]float32
	time   int32
	sTime  float32
	tex1   Texture
	tex2   Texture
}

func (rp *RenderParams) IsValid() bool {
	return rp.tex != nil && rp.tex.IsValid() && rp.size[0] != 0 && rp.size[1] != 0 &&
		IsFinite(rp.x+rp.y+rp.xts+rp.xbs+rp.ys+rp.vs+rp.rxadd+rp.rot.angle+rp.rcx+rp.rcy)
}

func drawQuads(modelview mgl.Mat4, x1, y1, x2, y2, x3, y3, x4, y4 float32) {
	gfx.SetUniformMatrix("modelview", modelview[:])
	gfx.SetUniformF("x1x2x4x3", x1, x2, x4, x3) // this uniform is optional

	// This effectively side-steps a low-level rectangle rasterization edge case,
	// that caused visible and frequent artifacts on the diagonal.
	// See: https://github.com/ikemen-engine/Ikemen-GO/issues/3583
	uvBias := float32(0.000002)

	gfx.SetVertexData(
		x2, y2, 1, 1-uvBias,
		x3, y3, 1, 0,
		x1, y1, uvBias, 1-uvBias,
		x4, y4, uvBias, 0,
	)

	gfx.RenderQuad()
}

func applyRotation(modelview mgl.Mat4, rp RenderParams) mgl.Mat4 {
	aspectGame := sys.getCurrentAspect()
	aspectWindow := float32(sys.scrrect[2]) / float32(sys.scrrect[3])

	rotMatrix := func() mgl.Mat4 {
		return mgl.Rotate3DX(-rp.rot.xangle * math.Pi / 180.0).
			Mul3(mgl.Rotate3DY(rp.rot.yangle * math.Pi / 180.0)).
			Mul3(mgl.Rotate3DZ(rp.rot.angle * math.Pi / 180.0)).
			Mat4()
	}

	if Abs(aspectGame-aspectWindow) > 0.01 {
		if aspectWindow > aspectGame {
			// Window wider: normalize X
			scaleX := aspectWindow / aspectGame
			modelview = modelview.Mul4(mgl.Scale3D(scaleX, rp.vs, 1)) // pre-scale
			modelview = modelview.Mul4(rotMatrix())                   // rotate
			modelview = modelview.Mul4(mgl.Scale3D(1/scaleX, 1, 1))   // restore
		} else {
			// Window taller: normalize Y
			scaleY := aspectGame / aspectWindow
			modelview = modelview.Mul4(mgl.Scale3D(1, scaleY*rp.vs, 1))
			modelview = modelview.Mul4(rotMatrix())
			modelview = modelview.Mul4(mgl.Scale3D(1, 1/scaleY, 1))
		}
	} else {
		// Same aspect: simple rotation
		modelview = modelview.Mul4(mgl.Scale3D(1, rp.vs, 1))
		modelview = modelview.Mul4(rotMatrix())
	}

	return modelview
}

// Builds the base projection transform depending on projectionMode
func applyProjection(modelview mgl.Mat4, rp RenderParams, n int, botdist, dy float32) mgl.Mat4 {
	if rp.projectionMode == 0 {
		// No projection, just center on pivot + tile offset
		return modelview.Mul4(mgl.Translate3D(rp.rcx+float32(n)*botdist, rp.rcy+dy, 0))
	}

	matrix := mgl.Mat4{float32(sys.scrrect[2] / 2.0), 0, 0, 0, 0, float32(sys.scrrect[3] / 2), 0, 0, 0, 0, -65535, 0, float32(sys.scrrect[2] / 2), float32(sys.scrrect[3] / 2), 0, 1}

	if rp.projectionMode == 1 {
		modelview = modelview.Mul4(mgl.Translate3D(0, -float32(sys.scrrect[3]), rp.fLength))
		modelview = modelview.Mul4(matrix)
		modelview = modelview.Mul4(mgl.Frustum(-float32(sys.scrrect[2])/2/rp.fLength, float32(sys.scrrect[2])/2/rp.fLength, -float32(sys.scrrect[3])/2/rp.fLength, float32(sys.scrrect[3])/2/rp.fLength, 1.0, 65535))
		modelview = modelview.Mul4(mgl.Translate3D(-float32(sys.scrrect[2])/2.0, float32(sys.scrrect[3])/2.0, -rp.fLength))
		return modelview.Mul4(mgl.Translate3D(rp.rcx+float32(n)*botdist, rp.rcy+dy, 0))
	}

	if rp.projectionMode == 2 {
		modelview = modelview.Mul4(mgl.Translate3D(rp.rcx-float32(sys.scrrect[2])/2.0-rp.xOffset, rp.rcy-float32(sys.scrrect[3])/2.0+rp.yOffset, rp.fLength))
		modelview = modelview.Mul4(matrix)
		modelview = modelview.Mul4(mgl.Frustum(-float32(sys.scrrect[2])/2/rp.fLength, float32(sys.scrrect[2])/2/rp.fLength, -float32(sys.scrrect[3])/2/rp.fLength, float32(sys.scrrect[3])/2/rp.fLength, 1.0, 65535))
		return modelview.Mul4(mgl.Translate3D(rp.xOffset+float32(n)*botdist, -rp.yOffset+dy, -rp.fLength))
	}

	return modelview
}

// Applies shear matrix before rotation
func applyShear(modelview mgl.Mat4, rxadd, width float32) mgl.Mat4 {
	if rxadd == 0 {
		return modelview
	}
	shearMatrix := mgl.Mat4{
		1, 0, 0, 0,
		rxadd, 1, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}
	modelview = modelview.Mul4(shearMatrix)
	return modelview.Mul4(mgl.Translate3D(rxadd*width, 0, 0))
}

// Applies the sprite rotation path to a truetype glyph quad
func transformTextQuad(x1, y1, x2, y2, x3, y3, x4, y4, rxadd float32,
	rot Rotation, projectionMode int32, fLength, rcx, rcy float32) (float32, float32, float32, float32, float32, float32, float32, float32) {
	if rot.IsZero() && rxadd == 0 {
		return x1, y1, x2, y2, x3, y3, x4, y4
	}

	sx, sy := sys.widthScale, sys.heightScale
	if sx == 0 {
		sx = 1
	}
	if sy == 0 {
		sy = 1
	}
	screenH := float32(sys.scrrect[3])

	toSpriteSpace := func(x, y float32) (float32, float32) {
		return x * sx, -y * sy
	}
	toTextSpace := func(x, y float32) (float32, float32) {
		return x / sx, (screenH - y) / sy
	}
	transformPoint := func(modelview mgl.Mat4, x, y float32) (float32, float32) {
		v := modelview.Mul4x1(mgl.Vec4{x, y, 0, 1})
		if v.W() != 0 {
			invW := 1 / v.W()
			return v.X() * invW, v.Y() * invW
		}
		return v.X(), v.Y()
	}

	// truetype quads use top-left coordinates but sprite transforms expect bottom-left ones
	blx, bly := toSpriteSpace(x3, y3)
	brx, bry := toSpriteSpace(x4, y4)
	trx, try := toSpriteSpace(x1, y1)
	tlx, tly := toSpriteSpace(x2, y2)

	rp := RenderParams{
		size:           [2]uint16{1, 1},
		tile:           notiling,
		xts:            trx - tlx,
		xbs:            brx - blx,
		ys:             try - bry,
		vs:             1,
		rxadd:          rxadd,
		xas:            1,
		yas:            1,
		rot:            rot,
		rcx:            rcx * sx,
		rcy:            -rcy * sy,
		projectionMode: projectionMode,
		fLength:        fLength,
	}

	if !rp.rot.IsZero() {
		modelview := mgl.Translate3D(0, screenH, 0)
		modelview = applyProjection(modelview, rp, 0, 1, 0)
		modelview = applyShear(modelview, rp.rxadd, rp.ys)
		modelview = applyRotation(modelview, rp)
		modelview = modelview.Mul4(mgl.Translate3D(-rp.rcx, -rp.rcy, 0))

		blx, bly = transformPoint(modelview, blx, bly)
		brx, bry = transformPoint(modelview, brx, bry)
		trx, try = transformPoint(modelview, trx, try)
		tlx, tly = transformPoint(modelview, tlx, tly)
	} else {
		if rp.rxadd != 0 {
			// Match the unrotated sprite path by shifting the bottom edge
			blx += rp.rxadd * rp.ys
			brx = blx + rp.xbs
		}
		// Sprite rendering translates quads into screen space before drawQuads
		bly += screenH
		bry += screenH
		try += screenH
		tly += screenH
	}

	x1, y1 = toTextSpace(trx, try)
	x2, y2 = toTextSpace(tlx, tly)
	x3, y3 = toTextSpace(blx, bly)
	x4, y4 = toTextSpace(brx, bry)

	return x1, y1, x2, y2, x3, y3, x4, y4
}

// Render a quad with optional horizontal tiling
func renderSpriteHTile(modelview mgl.Mat4, x1, y1, x2, y2, x3, y3, x4, y4, dy, width float32, rp RenderParams) {
	//            p3
	//    p4 o-----o-----o- - -o
	//      /      |      \     ` .
	//     /       |       \       `.
	//    o--------o--------o- - - - o
	//   p1         p2
	topdist := (x3 - x4) * (((float32(rp.tile.xspacing) + width) / rp.xas) / width)
	botdist := (x2 - x1) * (((float32(rp.tile.xspacing) + width) / rp.xas) / width)
	if Abs(topdist) >= 0.01 {
		db := (x4 - rp.rcx) * (botdist - topdist) / Abs(topdist)
		x1 += db
		x2 += db
	}

	// Compute left/right tiling bounds (or right/left when topdist < 0)
	xmax := float32(sys.scrrect[2])
	left, right := int32(0), int32(1)
	if rp.tile.xflag != 0 {
		if rp.projectionMode == 0 {
			// Original culling logic (only when no projection)
			if topdist >= 0.01 {
				if x1 > x2 {
					left = 1 - int32(math.Ceil(float64(Max(x4/topdist, x1/botdist))))
					right = int32(math.Ceil(float64(Max((xmax-x3)/topdist, (xmax-x2)/botdist))))
				} else {
					left = 1 - int32(math.Ceil(float64(Max(x3/topdist, x2/botdist))))
					right = int32(math.Ceil(float64(Max((xmax-x4)/topdist, (xmax-x1)/botdist))))
				}
			} else if topdist <= -0.01 {
				if x1 > x2 {
					left = 1 - int32(math.Ceil(float64(Max((xmax-x3)/-topdist, (xmax-x2)/-botdist))))
					right = int32(math.Ceil(float64(Max(x4/-topdist, x1/-botdist))))
				} else {
					left = 1 - int32(math.Ceil(float64(Max((xmax-x4)/-topdist, (xmax-x1)/-botdist))))
					right = int32(math.Ceil(float64(Max(x3/-topdist, x2/-botdist))))
				}
			}
			if rp.tile.xflag != 1 {
				left = 0
				right = Min(right, Max(rp.tile.xflag, 1))
			}
		} else {
			// When projection is active: skip horizontal culling (geometry distortion breaks it)
			// Instead, use a fixed symmetric range based on xflag to avoid infinite tiling
			left = 1 - rp.tile.xflag
			right = rp.tile.xflag
		}
	}

	// Draw all quads in one loop
	for n := left; n < right; n++ {
		x1d, x2d := x1+float32(n)*botdist, x2+float32(n)*botdist
		x3d, x4d := x3+float32(n)*topdist, x4+float32(n)*topdist
		mat := modelview
		if !rp.rot.IsZero() {
			mat = applyProjection(mat, rp, int(n), botdist, dy)
			mat = applyRotation(mat, rp)
			mat = mat.Mul4(mgl.Translate3D(-(rp.rcx + float32(n)*botdist), -(rp.rcy + dy), 0))
		}

		drawQuads(mat, x1d, y1, x2d, y2, x3d, y3, x4d, y4)
	}
}

func renderSpriteQuad(modelview mgl.Mat4, rp RenderParams) {
	x1, y1 := rp.x, rp.rcy+((rp.y-rp.ys*float32(rp.size[1]))-rp.rcy)*rp.vs
	x2, y2 := x1+rp.xbs*float32(rp.size[0]), y1
	x3, y3 := rp.x+rp.xts*float32(rp.size[0]), rp.rcy+(rp.y-rp.rcy)*rp.vs
	x4, y4 := rp.x, y3
	//var pers float32
	//if Abs(rp.xts) < Abs(rp.xbs) {
	//	pers = Abs(rp.xts) / Abs(rp.xbs)
	//} else {
	//	pers = Abs(rp.xbs) / Abs(rp.xts)
	//}
	if !rp.rot.IsZero() && rp.tile.xflag == 0 && rp.tile.yflag == 0 {
		// This block makes shadows ignore their own yscale when in perspective
		//if rp.vs != 1 {
		//	y1 = rp.rcy + ((rp.y - rp.ys*float32(rp.size[1])) - rp.rcy)
		//	y2 = y1
		//	y3 = rp.y
		//	y4 = y3
		//}
		modelview = applyProjection(modelview, rp, 0, 1, 0)
		modelview = applyShear(modelview, rp.rxadd, rp.ys*float32(rp.size[1]))
		modelview = applyRotation(modelview, rp)
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
			if Abs(y3d-y4d) < 0.01 {
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
				renderSpriteHTile(modelview, x1d, y1d, x2d, y2d, x3d, y3d, x4d, y4d, y1d-y1, float32(rp.size[0]), rp)
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
				renderSpriteHTile(modelview, x1, y1, x2, y2, x3, y3, x4, y4, y1-oy, float32(rp.size[0]), rp)
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
			if Abs(y1-y2) < 0.01 {
				break
			}
			y1 = y2
		}
	}
}

func initRenderSpriteQuad(rp *RenderParams) {
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

func RenderSprite(rp RenderParams) {
	if !rp.IsValid() {
		return
	}

	initRenderSpriteQuad(&rp)

	// PalFX and color setup
	spfx := ShaderPalFX{
		neg:      false,
		add:      [3]float32{0, 0, 0},
		mult:     [3]float32{1, 1, 1},
		gray:     0,
		hue:      0,
		invblend: 0,
	}
	if rp.pfx != nil {
		spfx = rp.pfx.getFinalPalFx(rp.blendMode, rp.blendAlpha)
	}

	tint := [4]float32{
		float32(rp.tint&0xff) / 255,
		float32(rp.tint>>8&0xff) / 255,
		float32(rp.tint>>16&0xff) / 255,
		float32(rp.tint>>24&0xff) / 255,
	}

	proj := gfx.OrthographicProjectionMatrix(0, float32(sys.scrrect[2]), 0, float32(sys.scrrect[3]), -65535, 65535)
	modelview := mgl.Translate3D(0, float32(sys.scrrect[3]), 0)

	// Heavy state change
	// Because renderWithBlending() sometimes needs 2 passes, we'll do most of the setup outside of render()
	gfx.SetSpritePipeline(rp.customShader.name)

	gfx.EnableScissor(rp.window[0], rp.window[1], rp.window[2], rp.window[3])

	// Static uniforms
	gfx.SetUniformMatrix("projection", proj[:])
	gfx.SetUniformI("isFlat", 0)
	gfx.SetUniformI("mask", int(rp.mask))
	gfx.SetUniformI("isTrapez", int(Btoi(Abs(Abs(rp.xts)-Abs(rp.xbs)) > 0.001)))

	gfx.SetUniformF("gray", spfx.gray)
	gfx.SetUniformF("hue", spfx.hue)
	gfx.SetUniformFv("tint", tint[:])

	if rp.paltex == nil {
		gfx.SetUniformI("isRgba", 1)
	} else {
		gfx.SetUniformI("isRgba", 0)
	}

	if rp.customShader.name != "" {
		var timeSec float32
		if sys.middleOfMatch() {
			timeSec = float32(sys.gameTime())
		} else {
			timeSec = float32(sys.frameCounter)
		}
		gfx.SetUniformF("iTime", timeSec/60.0)
		gfx.SetUniformF("sTime", rp.customShader.sTime)
		gfx.SetUniformF("iResolution", float32(sys.scrrect[2]), float32(sys.scrrect[3]))
		aspectRatio := sys.getCurrentAspect() / sys.getFightAspect()
		gfx.SetUniformF("aspectRatio", aspectRatio)

		if gfx.NeedsGrabPass() {
			grabTex := gfx.ResolveBackBuffer()
			if grabTex != nil {
				gfx.SetTexture("bgl_RenderedTexture", grabTex)
			}
		}
		if rp.customShader.tex1 != nil {
			gfx.SetTexture("tex1", rp.customShader.tex1)
		}
		if rp.customShader.tex2 != nil {
			gfx.SetTexture("tex2", rp.customShader.tex2)
		}
		gfx.SetCustomUniforms(rp.customShader.params)
	}
	// Texture binding
	gfx.SetTexture("tex", rp.tex)
	if rp.paltex != nil {
		gfx.SetTexture("pal", rp.paltex)
	}

	// Local function called for each blending pass
	renderPass := func(eq BlendEquation, src, dst BlendFunc, a float32) {
		// Lightweight state change
		gfx.EnableBlending(eq, src, dst)

		// Dynamic uniforms
		// We must include the parameters that renderWithBlending() may have changed
		gfx.SetUniformI("neg", int(Btoi(spfx.neg)))
		gfx.SetUniformFv("add", spfx.add[:])
		gfx.SetUniformFv("mult", spfx.mult[:])
		gfx.SetUniformF("alpha", a)

		renderSpriteQuad(modelview, rp)
	}

	renderWithBlending(renderPass, rp.blendMode, rp.blendAlpha, rp.paltex != nil, &spfx, rp.paltex == nil)

	gfx.DisableScissor()
}

func renderWithBlending(
	render func(eq BlendEquation, src, dst BlendFunc, a float32),
	blendMode TransType, blendAlpha [2]int32, correctAlpha bool,
	spfx *ShaderPalFX, isrgba bool) {

	blendSourceFactor := BlendSrcAlpha
	if !correctAlpha {
		blendSourceFactor = BlendOne
	}

	Blend := BlendAdd
	BlendInv := BlendReverseSubtract
	if spfx.invblend >= 1 {
		Blend = BlendReverseSubtract
		BlendInv = BlendAdd
	}

	// Convert alpha to the float the renderer uses
	src := float32(blendAlpha[0]) / 255.0
	dst := float32(blendAlpha[1]) / 255.0

	// Ensure proper source and destination
	src = Clamp(src, 0, 1)
	dst = Clamp(dst, 0, 1)

	// Force None destination to 0 just in case
	if blendMode == TT_none {
		dst = 0
	}

	// Helpers for invertblend
	// Invert PalFX add
	invertAColor := func() {
		spfx.add[0], spfx.add[1], spfx.add[2] = -spfx.add[0], -spfx.add[1], -spfx.add[2]
	}
	// Disable the "neg" uniform, which the shader uses to invert colors
	// We sometimes use this because subtractive transparency already inverts colors on its own, so it'd be a double negation
	disableNeg := func() {
		spfx.neg = false
	}

	// Proceed with the render calls
	switch {
	// Sub
	case blendMode == TT_sub:
		switch {
		case src == 0 && dst == 1:
			// Fully transparent. Skip render
		case src == 1 && dst == 1:
			// Fast path for full subtraction
			if spfx.invblend >= 1 {
				invertAColor()
			}
			if spfx.invblend == 3 {
				disableNeg()
			}
			render(BlendInv, blendSourceFactor, BlendOne, 1)
		default:
			// Full alpha range
			if dst < 1 {
				render(BlendAdd, BlendZero, BlendOneMinusSrcAlpha, 1-dst)
			}
			if src > 0 {
				if spfx.invblend >= 1 {
					invertAColor()
				}
				if spfx.invblend == 3 {
					disableNeg()
				}
				render(BlendInv, blendSourceFactor, BlendOne, src)
			}
		}
	// SubAdd
	case blendMode == TT_subadd:
		// Save original state for later restoration
		origState := *spfx

		// Helper to set PalFX parameters to grayscale for the first pass
		makeFxGrayscale := func() {
			avgAdd := (origState.add[0] + origState.add[1] + origState.add[2]) / 3
			spfx.add = [3]float32{avgAdd, avgAdd, avgAdd}
			avgMult := (origState.mult[0] + origState.mult[1] + origState.mult[2]) / 3
			spfx.mult = [3]float32{avgMult, avgMult, avgMult}
		}
		if spfx.neg {
			// With invertall we invert the passes. Add then sub
			// This is kind of like "invblend = 3"
			// But adding "invertblend" support here would force people to have to use it to get the expected results
			// We avoid "neg" entirely because it turns black edges into white auras. But maybe allowing that would be more consistent?
			disableNeg() // spfx.neg = false
			//invertAColor()

			// Pass 1: additive with gray PalFX
			if dst > 0 {
				makeFxGrayscale()
				// Set gray uniform manually because render() doesn't set it (and doesn't need to most of the time)
				gfx.SetUniformF("gray", 1.0)
				render(BlendAdd, blendSourceFactor, BlendOne, dst)
			}
			// Pass 2: subtractive with original PalFX
			if src > 0 {
				*spfx = origState
				disableNeg() // Disable neg again because of the restore
				gfx.SetUniformF("gray", spfx.gray)
				render(BlendReverseSubtract, blendSourceFactor, BlendOne, src)
			}
		} else {
			// Normal behavior
			// Pass 1: subtractive with gray PalFX
			if dst > 0 {
				makeFxGrayscale()
				gfx.SetUniformF("gray", 1.0)
				render(BlendReverseSubtract, blendSourceFactor, BlendOne, dst)
			}
			// Pass 2: additive with original PalFX
			if src > 0 {
				*spfx = origState
				gfx.SetUniformF("gray", spfx.gray)
				render(BlendAdd, blendSourceFactor, BlendOne, src)
			}
		}
	// Add, None or Default
	// None takes this path because SuperPause darkens sprites through their source alpha
	// Default should normally not reach here, so this is only a fallback
	default:
		switch {
		case src == 0 && dst == 1:
			// Fully transparent. Just don't render
		case src == 1 && dst == 0:
			// Fast path for fully opaque
			render(BlendAdd, blendSourceFactor, BlendOneMinusSrcAlpha, 1)
		case src == 1 && dst == 1:
			// Fast path for full Add
			if spfx.invblend >= 1 {
				invertAColor()
			}
			if spfx.invblend == 3 {
				disableNeg()
			}
			render(Blend, blendSourceFactor, BlendOne, 1)
		default:
			// AddAlpha (includes Add1)
			if dst < 1 {
				render(Blend, BlendZero, BlendOneMinusSrcAlpha, 1-dst)
			}
			if src > 0 {
				if spfx.invblend >= 1 && dst == 1 {
					Blend = BlendReverseSubtract
					if spfx.invblend >= 2 { // Not 1 here. TODO: Explain why in comment
						invertAColor()
					}
					if spfx.invblend == 3 {
						disableNeg()
					}
				} else {
					Blend = BlendAdd
				}
				if !isrgba && (spfx.invblend <= -1 || spfx.invblend >= 2) && src < 1 {
					// Sum of add components
					gc := Abs(spfx.add[0]) + Abs(spfx.add[1]) + Abs(spfx.add[2])
					v3, ml, al := Max(255*(gc-(src+dst)), 512)/128, src, src+dst
					rM, gM, bM := spfx.mult[0]*ml, spfx.mult[1]*ml, spfx.mult[2]*ml
					spfx.mult[0], spfx.mult[1], spfx.mult[2] = rM, gM, bM
					render(Blend, blendSourceFactor, BlendOne, al*Pow(v3, 3))
				} else {
					render(Blend, blendSourceFactor, BlendOne, src)
				}
			}
		}
	}
}

func FillRect(rect [4]int32, color uint32, alpha [2]int32, fx *PalFX) {
	r := float32(color>>16&0xff) / 255
	g := float32(color>>8&0xff) / 255
	b := float32(color&0xff) / 255

	// PalFX setup
	spfx := ShaderPalFX{
		neg:      false,
		add:      [3]float32{0, 0, 0},
		mult:     [3]float32{1, 1, 1},
		gray:     0,
		hue:      0,
		invblend: 0,
	}

	// This call is safe even if fx is nil. Defaults to just AllPalFX
	spfx = fx.getFinalPalFx(TT_add, alpha)

	modelview := mgl.Translate3D(0, float32(sys.scrrect[3]), 0)
	proj := gfx.OrthographicProjectionMatrix(0, float32(sys.scrrect[2]), 0, float32(sys.scrrect[3]), -65535, 65535)

	x1, y1 := float32(rect[0]), -float32(rect[1])
	x2, y2 := float32(rect[0]+rect[2]), -float32(rect[1]+rect[3])

	// Prepare the heavy state
	gfx.SetSpritePipeline("")

	// Set geometry
	gfx.SetVertexData(
		x2, y2, 1, 1,
		x2, y1, 1, 0,
		x1, y2, 0, 1,
		x1, y1, 0, 0,
	)

	// Static uniforms
	gfx.SetUniformMatrix("modelview", modelview[:])
	gfx.SetUniformMatrix("projection", proj[:])
	gfx.SetUniformI("isFlat", 1)
	gfx.SetUniformI("isTrapez", 0)
	gfx.SetUniformI("mask", 0)
	gfx.SetUniformI("isRgba", 1)
	gfx.SetUniformF("gray", spfx.gray)
	gfx.SetUniformF("hue", spfx.hue)

	// Alpha is determined by tint, so we reset it here
	// TODO: Maybe the shader shouldn't have a duplicate alpha component inside "tint"
	gfx.SetUniformF("alpha", 1.0)

	// Local function called for each blending pass
	renderPass := func(eq BlendEquation, src, dst BlendFunc, a float32) {
		// Update only the dynamic state
		gfx.EnableBlending(eq, src, dst)
		gfx.SetUniformF("tint", r, g, b, a)
		gfx.SetUniformI("neg", int(Btoi(spfx.neg)))
		gfx.SetUniformFv("add", spfx.add[:])
		gfx.SetUniformFv("mult", spfx.mult[:])

		gfx.RenderQuad()
	}

	renderWithBlending(renderPass, TT_add, alpha, true, &spfx, true)
}

type TextureAtlas struct {
	texture Texture
	width   int32
	height  int32
	depth   int32
	filter  bool
	resize  bool
	skyline *list.List //[][2]uint32
}

func (ta *TextureAtlas) bytesPerPixel() int32 {
	bpp := ta.depth / 8
	if bpp < 1 {
		bpp = 1
	}
	return bpp
}

func (ta *TextureAtlas) clearTexture(tex Texture, width, height int32) {
	bpp := ta.bytesPerPixel()
	clearData := make([]byte, int(width*height*bpp))
	tex.SetSubData(clearData, 0, 0, width, height, width*bpp)
}

func extrudeAtlasImage(data []byte, width, height, stride, bpp int32) ([]byte, int32, bool) {
	if width <= 0 || height <= 0 || bpp <= 0 {
		return nil, 0, false
	}
	if stride <= 0 {
		stride = width * bpp
	}
	if int(stride*height) > len(data) {
		return nil, 0, false
	}

	outWidth := width + 2
	outHeight := height + 2
	outStride := outWidth * bpp
	out := make([]byte, int(outStride*outHeight))

	for y := int32(0); y < height; y++ {
		srcRow := data[y*stride : y*stride+width*bpp]
		dstRow := out[(y+1)*outStride+1*bpp : (y+1)*outStride+1*bpp+width*bpp]
		copy(dstRow, srcRow)

		leftPx := srcRow[:bpp]
		rightPx := srcRow[(width-1)*bpp : width*bpp]
		copy(out[(y+1)*outStride:(y+1)*outStride+bpp], leftPx)
		copy(out[(y+1)*outStride+(outWidth-1)*bpp:(y+1)*outStride+outWidth*bpp], rightPx)
	}

	firstRow := out[outStride : outStride+outStride]
	lastRow := out[(outHeight-2)*outStride : (outHeight-2)*outStride+outStride]
	copy(out[0:outStride], firstRow)
	copy(out[(outHeight-1)*outStride:(outHeight-1)*outStride+outStride], lastRow)

	return out, outStride, true
}

func CreateTextureAtlas(width, height int32, depth int32, filter bool) *TextureAtlas {
	ta := &TextureAtlas{width: width, height: height, texture: gfx.newTexture(width, height, depth, filter), depth: depth, filter: filter, skyline: list.New(), resize: false}
	ta.texture.SetData(nil) // Allocate storage where supported.
	ta.clearTexture(ta.texture, width, height)
	ta.skyline.PushBack([2]int32{0, 0})
	return ta
}

func (ta *TextureAtlas) AddImage(width, height, stride int32, data []byte) ([4]float32, bool) {
	const maxWidth = 4096

	// Initial check for images larger than the atlas's current size
	if ta.resize {
		if width > ta.width || height > ta.height {
			// If the image itself is bigger than half the max size, we can't fit it reliably
			if width > maxWidth/2 || height > maxWidth/2 {
				return [4]float32{}, false
			}
			// Attempt to grow to accommodate the large image immediately
			ta.Resize(width*2, height*2)
		}
	}

	x, y, ok := ta.FindPlaceToInsert(width, height)

	// If insertion failed, we need to try resizing the atlas
	if !ok {
		if ta.resize {
			// Check if either dimension still has room to grow
			//if ta.width != maxWidth && ta.height != maxWidth {
			if ta.width < maxWidth || ta.height < maxWidth {
				// Calculate target dimensions
				newW := ta.width * 2
				newH := ta.height * 2

				// Ensure we never exceed maxWidth
				if newW > maxWidth {
					newW = maxWidth
				}
				if newH > maxWidth {
					newH = maxWidth
				}

				// Only perform the resize if the dimensions actually changed
				if newW != ta.width || newH != ta.height {
					ta.Resize(newW, newH)

					// Retry insertion with the new larger atlas
					x, y, ok = ta.FindPlaceToInsert(width, height)
				}
			}
		}

		// If it still fails, return false
		if !ok {
			return [4]float32{}, false
		}
	}

	// Otherwise upload and return
	bpp := ta.bytesPerPixel()
	paddedData, paddedStride, ok := extrudeAtlasImage(data, width, height, stride, bpp)
	if !ok {
		return [4]float32{}, false
	}
	ta.texture.SetSubData(paddedData, x-1, y-1, width+2, height+2, paddedStride)

	return [4]float32{
		float32(x) / float32(ta.width),
		float32(y) / float32(ta.height),
		float32(x+width) / float32(ta.width),
		float32(y+height) / float32(ta.height),
	}, true
}

func (ta *TextureAtlas) FindPlaceToInsert(width, height int32) (int32, int32, bool) {
	//leave 1px space
	space := int32(1)
	width += space * 2
	height += space * 2
	var bestX int32 = math.MaxInt32
	var bestY int32 = math.MaxInt32
	var bestItr *list.Element = nil
	var bestItr2 *list.Element = nil
	for itr := ta.skyline.Front(); itr != nil; itr = itr.Next() {
		x := itr.Value.([2]int32)[0]
		y := itr.Value.([2]int32)[1]
		if width > ta.width-x {
			break
		}
		if y >= bestY {
			continue
		}
		xMax := x + width
		var itr2 *list.Element
		for itr2 = itr.Next(); itr2 != nil; itr2 = itr2.Next() {
			x2 := itr2.Value.([2]int32)[0]
			y2 := itr2.Value.([2]int32)[1]
			if xMax <= x2 {
				break
			}
			if y < y2 {
				y = y2
			}
		}
		if y >= bestY || height > ta.height-y {
			continue
		}
		bestItr = itr
		bestItr2 = itr2
		bestX = x
		bestY = y
	}
	if bestItr == nil {
		return 0, 0, false
	}
	ta.skyline.InsertBefore([2]int32{bestX, bestY + height}, bestItr)

	if bestItr2 == nil && bestX+width < ta.width {
		ta.skyline.InsertBefore([2]int32{bestX + width, ta.skyline.Back().Value.([2]int32)[1]}, bestItr)
	} else if bestItr2 != nil && bestX+width < bestItr2.Value.([2]int32)[0] {
		ta.skyline.InsertBefore([2]int32{bestX + width, bestItr2.Prev().Value.([2]int32)[1]}, bestItr)
	}
	itrNext := bestItr
	for itr := bestItr; itr != bestItr2; itr = itrNext {
		itrNext = itr.Next()
		ta.skyline.Remove(itr)
	}
	bestX += space
	bestY += space
	return bestX, bestY, true
}

func (ta *TextureAtlas) Resize(width, height int32) {
	if width < ta.width || height < ta.height {
		panic("New width cannot be smaller than old width")
	}
	if height < ta.height {
		panic("New height cannot be smaller than old height")
	}
	t := gfx.newTexture(width, height, ta.depth, ta.filter)
	ta.clearTexture(t, width, height)
	t.CopyData(&ta.texture)
	ta.skyline.PushBack([2]int32{ta.width, 0})
	ta.width = width
	ta.height = height
	ta.texture = t
	return
}

// The PalFX parameters sent to the shader uniforms
type ShaderPalFX struct {
	neg      bool
	add      [3]float32
	mult     [3]float32
	gray     float32
	hue      float32
	invblend int32
}

func NewShaderPalFX() ShaderPalFX {
	return ShaderPalFX{
		neg:      false,
		add:      [3]float32{0, 0, 0},
		mult:     [3]float32{1, 1, 1},
		gray:     0,
		hue:      0,
		invblend: 0,
	}
}
