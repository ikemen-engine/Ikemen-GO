package main

import (
	_ "embed" // Support for go:embed resources
	"encoding/binary"
	"fmt"
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
	trans int32
	mask  int32
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
	return rp.tex.handle.IsValid() && IsFinite(rp.x+rp.y+rp.xts+rp.xbs+rp.ys+rp.vs+
		rp.rxadd+rp.rot.angle+rp.rcx+rp.rcy)
}

var vertexBuffer, uvBuffer gl.Buffer

var mainShader, flatShader *ShaderProgram

// Post-processing
var fbo gl.Framebuffer
var fbo_texture gl.Texture

// Clasic AA
var rbo_depth gl.Renderbuffer

// MSAA
var fbo_f gl.Framebuffer
var fbo_f_texture *Texture

var postVertBuffer gl.Buffer
var postShaderSelect []*ShaderProgram

//go:embed shaders/sprite.vs.glsl
var vertShader string

//go:embed shaders/sprite.fs.glsl
var fragShader string

//go:embed shaders/flat.fs.glsl
var fragShaderFlat string

//go:embed shaders/ident.vs.glsl
var identVertShader string

//go:embed shaders/ident.fs.glsl
var identFragShader string

// Render initialization.
// Creates the default shaders, the framebuffer and enables MSAA.
func RenderInit() {
	sys.errLog.Printf("Using OpenGL %v (%v)",
		gl.GetString(gl.VERSION), gl.GetString(gl.RENDERER))

	mainShader = newShaderProgram(vertShader, fragShader, "Main Shader")
	mainShader.RegisterUniforms("pal", "mask", "neg", "gray", "add", "mul", "x1x2x4x3", "isRgba", "isTrapez")

	flatShader = newShaderProgram(vertShader, fragShaderFlat, "Flat Shader")
	flatShader.RegisterUniforms("color", "isShadow")

	// Persistent data buffers for rendering
	vertexBuffer = gl.CreateBuffer()

	uvData := f32.Bytes(binary.LittleEndian, 1, 1, 1, 0, 0, 1, 0, 0)
	uvBuffer = gl.CreateBuffer()
	gl.BindBuffer(gl.ARRAY_BUFFER, uvBuffer)
	gl.BufferData(gl.ARRAY_BUFFER, uvData, gl.STATIC_DRAW)

	postVertData := f32.Bytes(binary.LittleEndian, -1, -1, 1, -1, -1, 1, 1, 1)
	postVertBuffer = gl.CreateBuffer()
	gl.BindBuffer(gl.ARRAY_BUFFER, postVertBuffer)
	gl.BufferData(gl.ARRAY_BUFFER, postVertData, gl.STATIC_DRAW)

	// Compile postprocessing shaders

	// Calculate total ammount of shaders loaded.
	postShaderSelect = make([]*ShaderProgram, 1+len(sys.externalShaderList))

	// Ident shader (no postprocessing)
	postShaderSelect[0] = newShaderProgram(identVertShader, identFragShader, "Identity Postprocess")
	postShaderSelect[0].RegisterUniforms("Texture", "TextureSize")

	// External Shaders
	for i := 0; i < len(sys.externalShaderList); i++ {
		postShaderSelect[1+i] = newShaderProgram(sys.externalShaders[0][i],
			sys.externalShaders[1][i], fmt.Sprintf("Postprocess Shader #%v", i+1))
		postShaderSelect[1+i].RegisterUniforms("Texture", "TextureSize")
	}

	if sys.multisampleAntialiasing {
		gl.Enable(gl.MULTISAMPLE)
	}

	gl.ActiveTexture(gl.TEXTURE0)
	fbo_texture = gl.CreateTexture()
	gl.ObjectLabel(fbo_texture, "Main Framebuffer Texture")

	if sys.multisampleAntialiasing {
		gl.BindTexture(gl.TEXTURE_2D_MULTISAMPLE, fbo_texture)
	} else {
		gl.BindTexture(gl.TEXTURE_2D, fbo_texture)
	}

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)

	if sys.multisampleAntialiasing {
		gl.TexImage2DMultisample(gl.TEXTURE_2D_MULTISAMPLE, 16, gl.RGBA, int(sys.scrrect[2]), int(sys.scrrect[3]), false)
	} else {
		gl.TexImage2D(gl.TEXTURE_2D, 0, int(sys.scrrect[2]), int(sys.scrrect[3]), gl.RGBA, gl.UNSIGNED_BYTE, nil)
	}

	gl.BindTexture(gl.TEXTURE_2D, gl.NoTexture)

	if sys.multisampleAntialiasing {
		fbo_f_texture = newTexture()
		fbo_f_texture.SetData(sys.scrrect[2], sys.scrrect[3], 32, false, nil)
	} else {
		rbo_depth = gl.CreateRenderbuffer()
		gl.ObjectLabel(rbo_depth, "Depth Renderbuffer")
		gl.BindRenderbuffer(gl.RENDERBUFFER, rbo_depth)
		gl.RenderbufferStorage(gl.RENDERBUFFER, gl.DEPTH_COMPONENT16, int(sys.scrrect[2]), int(sys.scrrect[3]))
		gl.BindRenderbuffer(gl.RENDERBUFFER, gl.NoRenderbuffer)
	}

	fbo = gl.CreateFramebuffer()
	gl.ObjectLabel(fbo, "Main Framebuffer")
	gl.BindFramebuffer(gl.FRAMEBUFFER, fbo)

	if sys.multisampleAntialiasing {
		gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D_MULTISAMPLE, fbo_texture, 0)

		fbo_f = gl.CreateFramebuffer()
		gl.BindFramebuffer(gl.FRAMEBUFFER, fbo_f)
		gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, fbo_f_texture.handle, 0)
	} else {
		gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, fbo_texture, 0)
		gl.FramebufferRenderbuffer(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, gl.RENDERBUFFER, rbo_depth)
	}

	if status := gl.CheckFramebufferStatus(gl.FRAMEBUFFER); status != gl.FRAMEBUFFER_COMPLETE {
		sys.errLog.Printf("framebuffer create failed: 0x%x", status)
	}

	gl.BindFramebuffer(gl.FRAMEBUFFER, gl.NoFramebuffer)
}

func bindFB() {
	gl.BindFramebuffer(gl.FRAMEBUFFER, fbo)
}

func unbindFB() {
	if sys.multisampleAntialiasing {
		gl.BindFramebuffer(gl.DRAW_FRAMEBUFFER, fbo_f)
		gl.BindFramebuffer(gl.READ_FRAMEBUFFER, fbo)
		gl.BlitFramebuffer(0, 0, int(sys.scrrect[2]), int(sys.scrrect[3]), 0, 0, int(sys.scrrect[2]), int(sys.scrrect[3]), gl.COLOR_BUFFER_BIT, gl.LINEAR)
	}

	gl.BindFramebuffer(gl.FRAMEBUFFER, gl.NoFramebuffer)

	postShader := postShaderSelect[sys.postProcessingShader]

	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	postShader.UseProgram()

	if sys.multisampleAntialiasing {
		gl.BindTexture(gl.TEXTURE_2D, fbo_f_texture.handle)
	} else {
		gl.BindTexture(gl.TEXTURE_2D, fbo_texture)
	}

	gl.Uniform1i(postShader.u["Texture"], 0)
	gl.Uniform2f(postShader.u["TextureSize"], float32(sys.scrrect[2]), float32(sys.scrrect[3]))

	gl.BindBuffer(gl.ARRAY_BUFFER, postVertBuffer)
	gl.EnableVertexAttribArray(postShader.aVert)
	gl.VertexAttribPointer(postShader.aVert, 2, gl.FLOAT, false, 0, 0)

	gl.DrawArrays(gl.TRIANGLE_STRIP, 0, 4)

	gl.DisableVertexAttribArray(postShader.aVert)
}

func drawQuads(s *ShaderProgram, modelview mgl.Mat4, x1, y1, x2, y2, x3, y3, x4, y4 float32) {
	gl.UniformMatrix4fv(s.uModelView, modelview[:])
	if u, ok := s.u["x1x2x4x3"]; ok {
		gl.Uniform4f(u, x1, x2, x4, x3)
	}
	vertexPosition := f32.Bytes(binary.LittleEndian, x2, y2, x3, y3, x1, y1, x4, y4)
	gl.BufferData(gl.ARRAY_BUFFER, vertexPosition, gl.STATIC_DRAW)

	gl.DrawArrays(gl.TRIANGLE_STRIP, 0, 4)
}

func rmTileHSub(s *ShaderProgram, modelview mgl.Mat4, x1, y1, x2, y2, x3, y3, x4, y4, width float32,
	tl Tiling, rcx float32) {
	xtw, xbw := x3-x4, x2-x1
	xts, xbs := xtw/width, xbw/width
	topdist := xtw + xts*float32(tl.sx)
	if AbsF(topdist) >= 0.01 {
		botdist := xbw + xbs*float32(tl.sx)
		db := (x4 - rcx) * (botdist - topdist) / AbsF(topdist)
		x1 += db
		x2 += db
		if tl.x == 1 {
			x1d, x2d, x3d, x4d := x1, x2, x3, x4
			for {
				x2d = x1d - xbs*float32(tl.sx)
				x3d = x4d - xts*float32(tl.sx)
				x4d = x3d - xtw
				x1d = x2d - xbw
				if topdist < 0 {
					if x1d >= float32(sys.scrrect[2]) &&
						x2d >= float32(sys.scrrect[2]) && x3d >= float32(sys.scrrect[2]) &&
						x4d >= float32(sys.scrrect[2]) {
						break
					}
				} else if x1d <= 0 && x2d <= 0 && x3d <= 0 && x4d <= 0 {
					break
				}
				if (0 < x1d || 0 < x2d) &&
					(x1d < float32(sys.scrrect[2]) || x2d < float32(sys.scrrect[2])) ||
					(0 < x3d || 0 < x4d) &&
						(x3d < float32(sys.scrrect[2]) || x4d < float32(sys.scrrect[2])) {
					drawQuads(s, modelview, x1d, y1, x2d, y2, x3d, y3, x4d, y4)
				}
			}
		}
	}
	n := tl.x
	for {
		if topdist > 0 {
			if x1 >= float32(sys.scrrect[2]) && x2 >= float32(sys.scrrect[2]) &&
				x3 >= float32(sys.scrrect[2]) && x4 >= float32(sys.scrrect[2]) {
				break
			}
		} else if x1 <= 0 && x2 <= 0 && x3 <= 0 && x4 <= 0 {
			break
		}
		if (0 < x1 || 0 < x2) &&
			(x1 < float32(sys.scrrect[2]) || x2 < float32(sys.scrrect[2])) ||
			(0 < x3 || 0 < x4) &&
				(x3 < float32(sys.scrrect[2]) || x4 < float32(sys.scrrect[2])) {
			drawQuads(s, modelview, x1, y1, x2, y2, x3, y3, x4, y4)
		}
		if tl.x != 1 && n != 0 {
			n--
		}
		if n == 0 || AbsF(topdist) < 0.01 {
			break
		}
		x4 = x3 + xts*float32(tl.sx)
		x1 = x2 + xbs*float32(tl.sx)
		x2 = x1 + xbw
		x3 = x4 + xtw
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
	gl.UniformMatrix4fv(s.uProjection, proj[:])

	modelview := mgl.Translate3D(0, float32(sys.scrrect[3]), 0)

	gl.BindBuffer(gl.ARRAY_BUFFER, uvBuffer)
	gl.EnableVertexAttribArray(s.aUv)
	gl.VertexAttribPointer(s.aUv, 2, gl.FLOAT, false, 0, 0)

	// Keep vertexBuffer bound so that it can be updated in drawQuads()
	gl.BindBuffer(gl.ARRAY_BUFFER, vertexBuffer)
	gl.EnableVertexAttribArray(s.aPos)
	gl.VertexAttribPointer(s.aPos, 2, gl.FLOAT, false, 0, 0)

	renderWithBlending(func(a float32) {
		gl.Uniform1f(s.uAlpha, a)
		rmTileSub(s, modelview, rp)
	}, rp.trans, rp.paltex != nil)

	gl.DisableVertexAttribArray(s.aPos)
	gl.DisableVertexAttribArray(s.aUv)

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
	gl.Uniform1i(mainShader.uTexture, 0)
	if rp.paltex == nil {
		gl.Uniform1i(mainShader.u["isRgba"], 1)
	} else {
		gl.ActiveTexture(gl.TEXTURE1)
		gl.BindTexture(gl.TEXTURE_2D, rp.paltex.handle)
		gl.Uniform1i(mainShader.u["pal"], 1)
		gl.Uniform1i(mainShader.u["isRgba"], 0)
		gl.Uniform1i(mainShader.u["mask"], int(rp.mask))
	}
	gl.Uniform1i(mainShader.u["isTrapez"], int(Btoi(AbsF(AbsF(rp.xts)-AbsF(rp.xbs)) > 0.001)))

	neg, grayscale, padd, pmul := false, float32(0), [3]float32{0, 0, 0}, [3]float32{1, 1, 1}
	if rp.pfx != nil {
		neg, grayscale, padd, pmul = rp.pfx.getFcPalFx(rp.trans == -2)
		if rp.trans == -2 {
			padd[0], padd[1], padd[2] = -padd[0], -padd[1], -padd[2]
		}
	}
	gl.Uniform1i(mainShader.u["neg"], int(Btoi(neg)))
	gl.Uniform1f(mainShader.u["gray"], grayscale)
	gl.Uniform3fv(mainShader.u["add"], padd[:])
	gl.Uniform3fv(mainShader.u["mul"], pmul[:])

	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, rp.tex.handle)
	rmMainSub(mainShader, rp)
}

func RenderFlatSprite(rp RenderParams, color uint32) {
	if !rp.IsValid() {
		return
	}
	rmInitSub(&rp)
	flatShader.UseProgram()
	gl.Uniform1i(flatShader.uTexture, 0)
	gl.Uniform3f(
		flatShader.u["color"], float32(color>>16&0xff)/255, float32(color>>8&0xff)/255,
		float32(color&0xff)/255)
	gl.Uniform1i(flatShader.u["isShadow"], 1)
	gl.BindTexture(gl.TEXTURE_2D, rp.tex.handle)
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

	flatShader.UseProgram()
	gl.UniformMatrix4fv(flatShader.uModelView, modelview[:])
	gl.UniformMatrix4fv(flatShader.uProjection, proj[:])
	gl.Uniform1i(flatShader.uTexture, 0)
	gl.Uniform3f(flatShader.u["color"], r, g, b)
	gl.Uniform1i(flatShader.u["isShadow"], 0)

	x1, y1 := float32(rect[0]), -float32(rect[1])
	x2, y2 := float32(rect[0]+rect[2]), -float32(rect[1]+rect[3])
	vertexPosition := f32.Bytes(binary.LittleEndian, x2, y2, x2, y1, x1, y2, x1, y1)
	gl.BindBuffer(gl.ARRAY_BUFFER, vertexBuffer)
	gl.BufferData(gl.ARRAY_BUFFER, vertexPosition, gl.STATIC_DRAW)
	gl.EnableVertexAttribArray(flatShader.aPos)
	gl.VertexAttribPointer(flatShader.aPos, 2, gl.FLOAT, false, 0, 0)

	renderWithBlending(func(a float32) {
		gl.Uniform1f(flatShader.uAlpha, a)
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, 4)
	}, trans, true)

	gl.DisableVertexAttribArray(flatShader.aPos)
}
