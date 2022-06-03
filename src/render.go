package main

import (
	"encoding/binary"
	"fmt"
	"math"

	mgl "github.com/go-gl/mathgl/mgl32"
	gl "github.com/fyne-io/gl-js"
	"golang.org/x/mobile/exp/f32"
)

type Shader struct {
	// Program
	program     gl.Program
	// Attribute locations
	aPos        gl.Attrib
	aUv         gl.Attrib
	// Common uniforms
	uModelView  gl.Uniform
	uProjection gl.Uniform
	uTexture    gl.Uniform
	uAlpha      gl.Uniform
	// Additional uniforms
	u map[string]gl.Uniform
}

func newShader(program gl.Program) (s *Shader) {
	s = &Shader{program: program}

	s.aPos = gl.GetAttribLocation(s.program, "position")
	s.aUv = gl.GetAttribLocation(s.program, "uv")

	s.uModelView = gl.GetUniformLocation(s.program, "modelview")
	s.uProjection = gl.GetUniformLocation(s.program, "projection")
	s.uTexture = gl.GetUniformLocation(s.program, "tex")
	s.uAlpha = gl.GetUniformLocation(s.program, "alpha")
	s.u = make(map[string]gl.Uniform)

	return
}

func (s *Shader) RegisterUniforms(names ...string) {
	for _, name := range names {
		s.u[name] = gl.GetUniformLocation(s.program, name)
	}
}

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
	tex      *Texture
	paltex   *Texture
	// Size, position, tiling, scaling and rotation
	size     [2]uint16
	x, y     float32
	tile     *Tiling
	xts, xbs float32
	ys, vs   float32
	rxadd    float32
	rot      Rotation
	// Transparency and palette effects
	trans    int32
	pfx      *PalFX
	// Clipping
	window   *[4]int32
	// Rotation center
	rcx, rcy float32
	// Perspective projection
	projectionMode int32
	fLength  float32
	xOffset  float32
	yOffset  float32
}

func (rp *RenderParams) IsValid() bool {
	return rp.tex.handle.IsValid() && IsFinite(rp.x + rp.y + rp.xts + rp.xbs + rp.ys + rp.vs +
			rp.rxadd + rp.rot.angle + rp.rcx + rp.rcy)
}

var vertexUv = f32.Bytes(binary.LittleEndian, 1, 1, 1, 0, 0, 1, 0, 0)

var mainShader, flatShader *Shader

// Post-processing
var fbo gl.Framebuffer
var fbo_texture gl.Texture

// Clasic AA
var rbo_depth gl.Renderbuffer

// MSAA
var fbo_f gl.Framebuffer
var fbo_f_texture gl.Texture

var postShader gl.Program
var postVertAttrib gl.Attrib
var postTexUniform gl.Uniform
var postTexSizeUniform gl.Uniform
var postVertices = f32.Bytes(binary.LittleEndian, -1, -1, 1, -1, -1, 1, 1, 1)

var postShaderSelect []gl.Program

// Render initialization.
// Creates the default shaders, the framebuffer and enables MSAA.
func RenderInit() {
	vertShader := `
uniform mat4 modelview, projection;

attribute vec2 position;
attribute vec2 uv;

varying vec2 texcoord;

void main(void) {
	texcoord = uv;
	gl_Position = projection * (modelview * vec4(position, 0.0, 1.0));
}`

	// Main fragment shader, for RGBA and indexed sprites
	fragShader := `
uniform sampler2D tex;
uniform sampler2D pal;

uniform vec4 x1x2x4x3;
uniform vec3 add, mul;
uniform float alpha, gray;
uniform int mask;
uniform bool isRgba, isTrapez, neg;

varying vec2 texcoord;

void main(void) {
	vec2 uv = texcoord;
	if (isTrapez) {
		// ここから台形用のテクスチャ座標計算/ Compute texture coordinates for trapezoid from here
		float left = -mix(x1x2x4x3[2], x1x2x4x3[0], uv[1]);
		float right = mix(x1x2x4x3[3], x1x2x4x3[1], uv[1]);
		uv[0] = (left + gl_FragCoord.x) / (left + right); // ここまで / To this point
	}
	vec4 c = texture2D(tex, uv);
	vec3 neg_base = vec3(1.0);
	vec3 final_add = add;
	vec4 final_mul = vec4(mul, alpha);
	if (isRgba) {
		neg_base *= alpha;
		final_add *= c.a;
		final_mul.rgb *= alpha;
	} else {
		if (int(255.25*c.r) == mask) {
			c.a = 0.0;
		} else {
			c = texture2D(pal, vec2(c.r*0.9966, 0.5));
		}
	}
	if (neg) c.rgb = neg_base - c.rgb;
	c.rgb = mix(c.rgb, vec3((c.r + c.g + c.b) / 3.0), gray) + final_add;
	gl_FragColor = c * final_mul;
}`

	// “Flat” fragment shader, for shadows and plain, untextured quads
	fragShaderFlat := `
uniform sampler2D tex;
uniform vec3 color;
uniform float alpha;
uniform bool isShadow;

varying vec2 texcoord;

void main(void) {
	vec4 p = vec4(color, alpha);
	if (isShadow)
		p *= texture2D(tex, texcoord).a;
	gl_FragColor = p;
}`

	compile := func(shaderType gl.Enum, src string) (shader gl.Shader) {
		shader = gl.CreateShader(shaderType)
		gl.ShaderSource(shader, "#version 100\nprecision mediump float;\n" + src)
		gl.CompileShader(shader)
		ok := gl.GetShaderi(shader, gl.COMPILE_STATUS)
		if ok == 0 {
			log := gl.GetShaderInfoLog(shader)
			gl.DeleteShader(shader)
			panic(Error("Shader compile error: " + log))
		}
		return
	}
	link := func(v, f gl.Shader) (program gl.Program) {
		program = gl.CreateProgram()
		gl.AttachShader(program, v)
		gl.AttachShader(program, f)
		gl.LinkProgram(program)
		// Mark shaders for deletion when the program is deleted
		gl.DeleteShader(v)
		gl.DeleteShader(f)
		ok := gl.GetProgrami(program, gl.LINK_STATUS)
		if ok == 0 {
			log := gl.GetProgramInfoLog(program)
			gl.DeleteProgram(program)
			panic(Error("Link error: " + log))
		}
		return
	}

	vertObj := compile(gl.VERTEX_SHADER, vertShader)
	fragObj := compile(gl.FRAGMENT_SHADER, fragShader)
	prog := link(vertObj, fragObj)
	gl.ObjectLabel(prog, "Main Shader")
	mainShader = newShader(prog)
	mainShader.RegisterUniforms("pal", "mask", "neg", "gray", "add", "mul", "x1x2x4x3", "isRgba", "isTrapez")

	fragObj = compile(gl.FRAGMENT_SHADER, fragShaderFlat)
	prog = link(vertObj, fragObj)
	gl.ObjectLabel(prog, "Flat Shader")
	flatShader = newShader(prog)
	flatShader.RegisterUniforms("color", "isShadow")

	// Compile postprocessing shaders

	// Calculate total ammount of shaders loaded.
	postShaderSelect = make([]gl.Program, 1+len(sys.externalShaderList))

	// Ident shader (no postprocessing)
	vertObj = compile(gl.VERTEX_SHADER, identVertShader)
	fragObj = compile(gl.FRAGMENT_SHADER, identFragShader)
	postShaderSelect[0] = link(vertObj, fragObj)
	gl.ObjectLabel(postShaderSelect[0], "Identity Shader")

	// External Shaders
	for i := 0; i < len(sys.externalShaderList); i++ {
		vertObj = compile(gl.VERTEX_SHADER, sys.externalShaders[0][i])
		fragObj = compile(gl.FRAGMENT_SHADER, sys.externalShaders[1][i])
		postShaderSelect[1+i] = link(vertObj, fragObj)
		gl.ObjectLabel(postShaderSelect[1+i], fmt.Sprintf("Postprocess Shader #%v", i+1))
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
		fbo_f_texture = gl.CreateTexture()
		gl.BindTexture(gl.TEXTURE_2D, fbo_f_texture)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
		gl.TexImage2D(gl.TEXTURE_2D, 0, int(sys.scrrect[2]), int(sys.scrrect[3]), gl.RGBA, gl.UNSIGNED_BYTE, nil)
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
		gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, fbo_f_texture, 0)
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

	postShader = postShaderSelect[sys.postProcessingShader]

	postVertAttrib = gl.GetAttribLocation(postShader, "VertCoord")
	postTexUniform = gl.GetUniformLocation(postShader, "Texture")
	postTexSizeUniform = gl.GetUniformLocation(postShader, "TextureSize")

	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	gl.UseProgram(postShader)

	if sys.multisampleAntialiasing {
		gl.BindTexture(gl.TEXTURE_2D, fbo_f_texture)
	} else {
		gl.BindTexture(gl.TEXTURE_2D, fbo_texture)
	}

	gl.Uniform1i(postTexUniform, 0)
	gl.Uniform2f(postTexSizeUniform, float32(sys.scrrect[2]), float32(sys.scrrect[3]))
	vertexBuffer := gl.CreateBuffer()
	gl.ObjectLabel(vertexBuffer, "Postprocess Vertex Buffer")
	gl.BindBuffer(gl.ARRAY_BUFFER, vertexBuffer)
	gl.BufferData(gl.ARRAY_BUFFER, postVertices, gl.STATIC_DRAW)
	gl.EnableVertexAttribArray(postVertAttrib)
	gl.VertexAttribPointer(postVertAttrib, 2, gl.FLOAT, false, 0, 0)
	gl.DrawArrays(gl.TRIANGLE_STRIP, 0, 4)
	gl.DeleteBuffer(vertexBuffer)
	gl.DisableVertexAttribArray(postVertAttrib)
}

func drawQuads(s *Shader, modelview mgl.Mat4, x1, y1, x2, y2, x3, y3, x4, y4 float32) {
	gl.UniformMatrix4fv(s.uModelView, modelview[:])
	if u, ok := s.u["x1x2x4x3"]; ok {
		gl.Uniform4f(u, x1, x2, x4, x3)
	}
	vertexPosition := f32.Bytes(binary.LittleEndian, x2, y2, x3, y3, x1, y1, x4, y4)
	vertexBuffer := gl.CreateBuffer()
	gl.BindBuffer(gl.ARRAY_BUFFER, vertexBuffer)
	gl.BufferData(gl.ARRAY_BUFFER, vertexPosition, gl.STATIC_DRAW)
	gl.EnableVertexAttribArray(s.aPos)
	gl.VertexAttribPointer(s.aPos, 2, gl.FLOAT, false, 0, 0)

	uvBuffer := gl.CreateBuffer()
	gl.BindBuffer(gl.ARRAY_BUFFER, uvBuffer)
	gl.BufferData(gl.ARRAY_BUFFER, vertexUv, gl.STATIC_DRAW)
	gl.EnableVertexAttribArray(s.aUv)
	gl.VertexAttribPointer(s.aUv, 2, gl.FLOAT, false, 0, 0)

	gl.DrawArrays(gl.TRIANGLE_STRIP, 0, 4)

	gl.DeleteBuffer(vertexBuffer)
	gl.DeleteBuffer(uvBuffer)
	gl.DisableVertexAttribArray(s.aPos)
	gl.DisableVertexAttribArray(s.aUv)
}
func rmTileHSub(s *Shader, modelview mgl.Mat4, x1, y1, x2, y2, x3, y3, x4, y4, xtw, xbw, xts, xbs float32,
	tl *Tiling, rcx float32) {
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

func rmTileSub(s *Shader, modelview mgl.Mat4, rp RenderParams) {
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
				rmTileHSub(s, modelview, x1d, y1d, x2d, y2d, x3d, y3d, x4d, y4d, x3d-x4d, x2d-x1d,
					(x3d-x4d)/float32(rp.size[0]), (x2d-x1d)/float32(rp.size[0]), rp.tile, rp.rcx)
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
				rmTileHSub(s, modelview, x1, y1, x2, y2, x3, y3, x4, y4, x3-x4, x2-x1,
					(x3-x4)/float32(rp.size[0]), (x2-x1)/float32(rp.size[0]), rp.tile, rp.rcx)
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
func rmMainSub(s *Shader, rp RenderParams) {
	proj := mgl.Ortho(0, float32(sys.scrrect[2]), 0, float32(sys.scrrect[3]), -65535, 65535)
	gl.UniformMatrix4fv(s.uProjection, proj[:])

	modelview := mgl.Translate3D(0, float32(sys.scrrect[3]), 0)
	switch {
	case rp.trans == -1:
		gl.Uniform1f(s.uAlpha, 1)
		if rp.paltex != nil {
			gl.BlendFunc(gl.SRC_ALPHA, gl.ONE)
		} else {
			gl.BlendFunc(gl.ONE, gl.ONE)
		}
		gl.BlendEquation(gl.FUNC_ADD)
		rmTileSub(s, modelview, rp)
	case rp.trans == -2:
		gl.Uniform1f(s.uAlpha, 1)
		gl.BlendFunc(gl.ONE, gl.ONE)
		gl.BlendEquation(gl.FUNC_REVERSE_SUBTRACT)
		rmTileSub(s, modelview, rp)
	case rp.trans <= 0:
	case rp.trans < 255:
		gl.Uniform1f(s.uAlpha, float32(rp.trans)/255)
		if rp.paltex != nil {
			gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
		} else {
			gl.BlendFunc(gl.ONE, gl.ONE_MINUS_SRC_ALPHA)
		}
		gl.BlendEquation(gl.FUNC_ADD)
		rmTileSub(s, modelview, rp)
	case rp.trans < 512:
		gl.Uniform1f(s.uAlpha, 1)
		if rp.paltex != nil {
			gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
		} else {
			gl.BlendFunc(gl.ONE, gl.ONE_MINUS_SRC_ALPHA)
		}
		gl.BlendEquation(gl.FUNC_ADD)
		rmTileSub(s, modelview, rp)
	default:
		src, dst := rp.trans&0xff, rp.trans>>10&0xff
		aglOver := 0
		if dst < 255 {
			gl.Uniform1f(s.uAlpha, 1-float32(dst)/255)
			gl.BlendFunc(gl.ZERO, gl.ONE_MINUS_SRC_ALPHA)
			gl.BlendEquation(gl.FUNC_ADD)
			rmTileSub(s, modelview, rp)
			aglOver++
		}
		if src > 0 {
			if aglOver != 0 {
				rp.rot = Rotation{}
			}
			gl.Uniform1f(s.uAlpha, float32(src)/255)

			if rp.paltex != nil {
				gl.BlendFunc(gl.SRC_ALPHA, gl.ONE)
			} else {
				gl.BlendFunc(gl.ONE, gl.ONE)
			}
			gl.BlendEquation(gl.FUNC_ADD)
			rmTileSub(s, modelview, rp)
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
	tl := *rp.tile
	if tl.x == 0 {
		tl.sx = 0
	} else if tl.sx > 0 {
		tl.sx -= int32(rp.size[0])
	}
	if tl.y == 0 {
		tl.sy = 0
	} else if tl.sy > 0 {
		tl.sy -= int32(rp.size[1])
	}
	rp.tile = &tl
	if rp.xts >= 0 {
		rp.x *= -1
	}
	rp.x += rp.rcx
	rp.rcy *= -1
	if rp.ys < 0 {
		rp.y *= -1
	}
	rp.y += rp.rcy
	gl.Enable(gl.BLEND)
	gl.Enable(gl.SCISSOR_TEST)
	gl.Scissor(rp.window[0], sys.scrrect[3]-(rp.window[1]+rp.window[3]),
		rp.window[2], rp.window[3])
}

func RenderSprite(rp RenderParams, mask int32) {
	if !rp.IsValid() {
		return
	}
	rmInitSub(&rp)
	gl.UseProgram(mainShader.program)
	gl.Uniform1i(mainShader.uTexture, 0)
	if rp.paltex == nil {
		gl.Uniform1i(mainShader.u["isRgba"], 1)
	} else {
		gl.ActiveTexture(gl.TEXTURE1)
		gl.BindTexture(gl.TEXTURE_2D, rp.paltex.handle)
		gl.Uniform1i(mainShader.u["pal"], 1)
		gl.Uniform1i(mainShader.u["isRgba"], 0)
		gl.Uniform1i(mainShader.u["mask"], int(mask))
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
	gl.Disable(gl.SCISSOR_TEST)
	gl.Disable(gl.BLEND)
}

func RenderFlatSprite(rp RenderParams, color uint32) {
	if !rp.IsValid() {
		return
	}
	rmInitSub(&rp)
	gl.UseProgram(flatShader.program)
	gl.Uniform1i(flatShader.uTexture, 0)
	gl.Uniform3f(
		flatShader.u["color"], float32(color>>16&0xff)/255, float32(color>>8&0xff)/255,
		float32(color&0xff)/255)
	gl.Uniform1i(flatShader.u["isShadow"], 1)
	gl.BindTexture(gl.TEXTURE_2D, rp.tex.handle)
	rmMainSub(flatShader, rp)
	gl.Disable(gl.SCISSOR_TEST)
	gl.Disable(gl.BLEND)
}

func FillRect(rect [4]int32, color uint32, trans int32) {
	r := float32(color>>16&0xff) / 255
	g := float32(color>>8&0xff) / 255
	b := float32(color&0xff) / 255
	fill := func(a float32) {
		gl.Uniform1f(flatShader.uAlpha, a)
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, 4)
	}

	modelview := mgl.Translate3D(0, float32(sys.scrrect[3]), 0)
	proj := mgl.Ortho(0, float32(sys.scrrect[2]), 0, float32(sys.scrrect[3]), -65535, 65535)

	gl.Enable(gl.BLEND)

	gl.UseProgram(flatShader.program)
	gl.UniformMatrix4fv(flatShader.uModelView, modelview[:])
	gl.UniformMatrix4fv(flatShader.uProjection, proj[:])
	gl.Uniform1i(flatShader.uTexture, 0)
	gl.Uniform3f(flatShader.u["color"], r, g, b)
	gl.Uniform1i(flatShader.u["isShadow"], 0)

	x1, y1 := float32(rect[0]), -float32(rect[1])
	x2, y2 := float32(rect[0]+rect[2]), -float32(rect[1]+rect[3])
	vertexPosition := f32.Bytes(binary.LittleEndian, x2, y2, x2, y1, x1, y2, x1, y1)
	vertexBuffer := gl.CreateBuffer()
	gl.BindBuffer(gl.ARRAY_BUFFER, vertexBuffer)
	gl.BufferData(gl.ARRAY_BUFFER, vertexPosition, gl.STATIC_DRAW)
	gl.EnableVertexAttribArray(flatShader.aPos)
	gl.VertexAttribPointer(flatShader.aPos, 2, gl.FLOAT, false, 0, 0)

	if trans == -1 {
		gl.BlendFunc(gl.SRC_ALPHA, gl.ONE)
		gl.BlendEquation(gl.FUNC_ADD)
		fill(1)
	} else if trans == -2 {
		gl.BlendFunc(gl.ONE, gl.ONE)
		gl.BlendEquation(gl.FUNC_REVERSE_SUBTRACT)
		fill(1)
	} else if trans <= 0 {
	} else if trans < 255 {
		gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
		gl.BlendEquation(gl.FUNC_ADD)
		fill(float32(trans) / 256)
	} else if trans < 512 {
		gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
		gl.BlendEquation(gl.FUNC_ADD)
		fill(1)
	} else {
		src, dst := trans&0xff, trans>>10&0xff
		if dst < 255 {
			gl.BlendFunc(gl.ZERO, gl.ONE_MINUS_SRC_ALPHA)
			gl.BlendEquation(gl.FUNC_ADD)
			fill(float32(dst) / 255)
		}
		if src > 0 {
			gl.BlendFunc(gl.SRC_ALPHA, gl.ONE)
			gl.BlendEquation(gl.FUNC_ADD)
			fill(float32(src) / 255)
		}
	}
	gl.DeleteBuffer(vertexBuffer)
	gl.DisableVertexAttribArray(flatShader.aPos)
	gl.Disable(gl.BLEND)
}

var identVertShader string = `
attribute vec2 VertCoord;
uniform vec2 TextureSize;

varying vec2 texcoord;

void main()
{
	gl_Position = vec4(VertCoord, 0.0, 1.0);
	texcoord = (VertCoord + 1.0) / 2.0;
}`

var identFragShader string = `
uniform sampler2D Texture;

varying vec2 texcoord;

void main(void) {
	gl_FragColor = texture2D(Texture, texcoord);
}`
