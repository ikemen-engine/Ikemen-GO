package main

import (
	"math"
	"unsafe"

	"github.com/go-gl/gl/v2.1/gl"
	mgl "github.com/go-gl/mathgl/mgl32"
)

type Shader struct {
	// Program
	program uint32
	// Attribute locations
	aPos, aUv int32
	// Common uniforms
	uModelView, uProjection int32
	uTexture int32
	uAlpha int32
	// Additional uniforms
	u map[string]int32
}

func newShader(program uint32) (s *Shader) {
	s = &Shader{program: program}

	s.aPos = gl.GetAttribLocation(s.program, gl.Str("position\x00"))
	s.aUv = gl.GetAttribLocation(s.program, gl.Str("uv\x00"))

	s.uModelView = gl.GetUniformLocation(s.program, gl.Str("modelview\x00"))
	s.uProjection = gl.GetUniformLocation(s.program, gl.Str("projection\x00"))
	s.uTexture = gl.GetUniformLocation(s.program, gl.Str("tex\x00"))
	s.uAlpha = gl.GetUniformLocation(s.program, gl.Str("alpha\x00"))
	s.u = make(map[string]int32)

	return
}

func (s *Shader) RegisterUniforms(names ...string) {
	for _, name := range names {
		s.u[name] = gl.GetUniformLocation(s.program, gl.Str(name + "\x00"))
	}
}

var vertexUv = [8]float32{1, 1, 1, 0, 0, 1, 0, 0}
var notiling = [4]int32{0, 0, 0, 0}

var mainShader, flatShader *Shader

// Post-processing
var fbo, fbo_texture uint32

// Clasic AA
var rbo_depth uint32

// MSAA
var fbo_f, fbo_f_texture uint32

var postShader uint32
var postVertAttrib int32
var postTexUniform int32
var postTexSizeUniform int32
var postVertices = [8]float32{-1, -1, 1, -1, -1, 1, 1, 1}

var postShaderSelect []uint32

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
}` + "\x00"

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
}` + "\x00"

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
}` + "\x00"

	compile := func(shaderType uint32, src string) (shader uint32) {
		shader = gl.CreateShader(shaderType)
		s, _ := gl.Strs(src)
		var l int32 = int32(len(src) - 1)
		gl.ShaderSource(shader, 1, s, &l)
		gl.CompileShader(shader)
		var ok int32
		gl.GetShaderiv(shader, gl.COMPILE_STATUS, &ok)
		if ok == 0 {
			var err error
			var size, l int32
			gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &size)
			if size > 0 {
				str := make([]byte, size+1)
				gl.GetShaderInfoLog(shader, size, &l, &str[0])
				err = Error(str[:l])
			}
			chk(err)
			gl.DeleteShader(shader)
			panic(Error("Shader compile error"))
		}
		return
	}
	link := func(v, f uint32) (program uint32) {
		program = gl.CreateProgram()
		gl.AttachShader(program, v)
		gl.AttachShader(program, f)
		gl.LinkProgram(program)
		// Mark shaders for deletion when the program is deleted
		gl.DeleteShader(v)
		gl.DeleteShader(f)
		var ok int32
		gl.GetProgramiv(program, gl.LINK_STATUS, &ok)
		if ok == 0 {
			var err error
			var size, l int32
			gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &size)
			if size > 0 {
				str := make([]byte, size+1)
				gl.GetProgramInfoLog(program, size, &l, &str[0])
				err = Error(str[:l])
			}
			chk(err)
			gl.DeleteProgram(program)
			panic(Error("Link error"))
		}
		return
	}

	vertObj := compile(gl.VERTEX_SHADER, vertShader)
	fragObj := compile(gl.FRAGMENT_SHADER, fragShader)
	mainShader = newShader(link(vertObj, fragObj))
	mainShader.RegisterUniforms("pal", "mask", "neg", "gray", "add", "mul", "x1x2x4x3", "isRgba", "isTrapez")

	fragObj = compile(gl.FRAGMENT_SHADER, fragShaderFlat)
	flatShader = newShader(link(vertObj, fragObj))
	flatShader.RegisterUniforms("color", "isShadow")

	// Compile postprocessing shaders

	// Calculate total ammount of shaders loaded.
	postShaderSelect = make([]uint32, 1+len(sys.externalShaderList))

	// Ident shader (no postprocessing)
	vertObj = compile(gl.VERTEX_SHADER, identVertShader)
	fragObj = compile(gl.FRAGMENT_SHADER, identFragShader)
	postShaderSelect[0] = link(vertObj, fragObj)

	// External Shaders
	for i := 0; i < len(sys.externalShaderList); i++ {
		vertObj = compile(gl.VERTEX_SHADER, sys.externalShaders[0][i])
		fragObj = compile(gl.FRAGMENT_SHADER, sys.externalShaders[1][i])
		postShaderSelect[1+i] = link(vertObj, fragObj)
	}

	if sys.multisampleAntialiasing {
		gl.Enable(gl.MULTISAMPLE)
	}

	gl.ActiveTexture(gl.TEXTURE0)
	gl.GenTextures(1, &fbo_texture)

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
		gl.TexImage2DMultisample(gl.TEXTURE_2D_MULTISAMPLE, 16, gl.RGBA, sys.scrrect[2], sys.scrrect[3], false)
	} else {
		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, sys.scrrect[2], sys.scrrect[3], 0, gl.RGBA, gl.UNSIGNED_BYTE, nil)
	}

	gl.BindTexture(gl.TEXTURE_2D, 0)

	if sys.multisampleAntialiasing {
		gl.GenTextures(1, &fbo_f_texture)
		gl.BindTexture(gl.TEXTURE_2D, fbo_f_texture)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
		gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, sys.scrrect[2], sys.scrrect[3], 0, gl.RGBA, gl.UNSIGNED_BYTE, nil)
	} else {
		gl.GenRenderbuffers(1, &rbo_depth)
		gl.BindRenderbuffer(gl.RENDERBUFFER, rbo_depth)
		gl.RenderbufferStorage(gl.RENDERBUFFER, gl.DEPTH_COMPONENT16, sys.scrrect[2], sys.scrrect[3])
		gl.BindRenderbuffer(gl.RENDERBUFFER, 0)
	}

	gl.GenFramebuffers(1, &fbo)
	gl.BindFramebuffer(gl.FRAMEBUFFER, fbo)

	if sys.multisampleAntialiasing {
		gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D_MULTISAMPLE, fbo_texture, 0)

		gl.GenFramebuffers(1, &fbo_f)
		gl.BindFramebuffer(gl.FRAMEBUFFER, fbo_f)
		gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, fbo_f_texture, 0)
	} else {
		gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, fbo_texture, 0)
		gl.FramebufferRenderbuffer(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, gl.RENDERBUFFER, rbo_depth)
	}

	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
}

func bindFB() {
	gl.BindFramebuffer(gl.FRAMEBUFFER, fbo)
}

func unbindFB() {
	if sys.multisampleAntialiasing {
		gl.BindFramebuffer(gl.DRAW_FRAMEBUFFER, fbo_f)
		gl.BindFramebuffer(gl.READ_FRAMEBUFFER, fbo)
		gl.BlitFramebuffer(0, 0, sys.scrrect[2], sys.scrrect[3], 0, 0, sys.scrrect[2], sys.scrrect[3], gl.COLOR_BUFFER_BIT, gl.LINEAR)
	}

	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)

	postShader = postShaderSelect[sys.postProcessingShader]

	postVertAttrib = gl.GetAttribLocation(postShader, gl.Str("VertCoord\x00"))
	postTexUniform = gl.GetUniformLocation(postShader, gl.Str("Texture\x00"))
	postTexSizeUniform = gl.GetUniformLocation(postShader, gl.Str("TextureSize\x00"))

	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	gl.UseProgram(postShader)

	if sys.multisampleAntialiasing {
		gl.BindTexture(gl.TEXTURE_2D, fbo_f_texture)
	} else {
		gl.BindTexture(gl.TEXTURE_2D, fbo_texture)
	}

	gl.Uniform1i(postTexUniform, 0)
	gl.Uniform2f(postTexSizeUniform, float32(sys.scrrect[2]), float32(sys.scrrect[3]))
	gl.EnableVertexAttribArray(uint32(postVertAttrib))
	gl.VertexAttribPointer(uint32(postVertAttrib), 2, gl.FLOAT, false, 0, unsafe.Pointer(&postVertices[0]))
	gl.DrawArrays(gl.TRIANGLE_STRIP, 0, 4)
}

func drawQuads(s *Shader, modelview mgl.Mat4, x1, y1, x2, y2, x3, y3, x4, y4 float32) {
	vertexPosition := [8]float32{x2, y2, x3, y3, x1, y1, x4, y4}
	gl.UniformMatrix4fv(s.uModelView, 1, false, &modelview[0])
	if u, ok := s.u["x1x2x4x3"]; ok {
		gl.Uniform4f(u, x1, x2, x4, x3)
	}
	gl.EnableVertexAttribArray(uint32(s.aPos))
	gl.EnableVertexAttribArray(uint32(s.aUv))
	gl.VertexAttribPointer(uint32(s.aPos), 2, gl.FLOAT, false, 0, unsafe.Pointer(&vertexPosition[0]))
	gl.VertexAttribPointer(uint32(s.aUv), 2, gl.FLOAT, false, 0, unsafe.Pointer(&vertexUv[0]))
	gl.DrawArrays(gl.TRIANGLE_STRIP, 0, 4)
}
func rmTileHSub(s *Shader, modelview mgl.Mat4, x1, y1, x2, y2, x3, y3, x4, y4, xtw, xbw, xts, xbs float32,
	tl *[4]int32, rcx float32) {
	topdist := xtw + xts*float32((*tl)[0])
	if AbsF(topdist) >= 0.01 {
		botdist := xbw + xbs*float32((*tl)[0])
		db := (x4 - rcx) * (botdist - topdist) / AbsF(topdist)
		x1 += db
		x2 += db
		if (*tl)[2] == 1 {
			x1d, x2d, x3d, x4d := x1, x2, x3, x4
			for {
				x2d = x1d - xbs*float32((*tl)[0])
				x3d = x4d - xts*float32((*tl)[0])
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
	n := (*tl)[2]
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
		if (*tl)[2] != 1 && n != 0 {
			n--
		}
		if n == 0 || AbsF(topdist) < 0.01 {
			break
		}
		x4 = x3 + xts*float32((*tl)[0])
		x1 = x2 + xbs*float32((*tl)[0])
		x2 = x1 + xbw
		x3 = x4 + xtw
	}
}
func rmTileSub(s *Shader, modelview mgl.Mat4, w, h uint16, x, y float32, tl *[4]int32,
	xts, xbs, ys, vs, rxadd, agl, yagl, xagl, rcx, rcy float32, projectionMode int32, fLength, xOffset, yOffset float32) {
	x1, y1 := x+rxadd*ys*float32(h), rcy+((y-ys*float32(h))-rcy)*vs
	x2, y2 := x1+xbs*float32(w), y1
	x3, y3 := x+xts*float32(w), rcy+(y-rcy)*vs
	x4, y4 := x, y3
	//var pers float32
	//if AbsF(xts) < AbsF(xbs) {
	//	pers = AbsF(xts) / AbsF(xbs)
	//} else {
	//	pers = AbsF(xbs) / AbsF(xts)
	//}
	if agl != 0 || yagl != 0 || xagl != 0 {
		//	kaiten(&x1, &y1, float64(agl), rcx, rcy, vs)
		//	kaiten(&x2, &y2, float64(agl), rcx, rcy, vs)
		//	kaiten(&x3, &y3, float64(agl), rcx, rcy, vs)
		//	kaiten(&x4, &y4, float64(agl), rcx, rcy, vs)
		if vs != 1 {
			y1 = rcy + ((y - ys*float32(h)) - rcy)
			y2 = y1
			y3 = rcy + (y - rcy)
			y4 = y3
		}
		if projectionMode == 0 {
			modelview = modelview.Mul4(mgl.Translate3D(rcx, rcy, 0))
		} else if projectionMode == 1 {
			//This is the inverse of the orthographic projection matrix
			matrix := mgl.Mat4{float32(sys.scrrect[2] / 2.0), 0, 0, 0, 0, float32(sys.scrrect[3] / 2), 0, 0, 0, 0, -65535, 0, float32(sys.scrrect[2] / 2), float32(sys.scrrect[3] / 2), 0, 1}
			modelview = modelview.Mul4(mgl.Translate3D(0, -float32(sys.scrrect[3]), fLength))
			modelview = modelview.Mul4(matrix)
			modelview = modelview.Mul4(mgl.Frustum(-float32(sys.scrrect[2])/2/fLength, float32(sys.scrrect[2])/2/fLength, -float32(sys.scrrect[3])/2/fLength, float32(sys.scrrect[3])/2/fLength, 1.0, 65535))
			modelview = modelview.Mul4(mgl.Translate3D(-float32(sys.scrrect[2])/2.0, float32(sys.scrrect[3])/2.0, -fLength))
			modelview = modelview.Mul4(mgl.Translate3D(rcx, rcy, 0))
		} else if projectionMode == 2 {
			matrix := mgl.Mat4{float32(sys.scrrect[2] / 2.0), 0, 0, 0, 0, float32(sys.scrrect[3] / 2), 0, 0, 0, 0, -65535, 0, float32(sys.scrrect[2] / 2), float32(sys.scrrect[3] / 2), 0, 1}
			//modelview = modelview.Mul4(mgl.Translate3D(0, -float32(sys.scrrect[3]), 2048))
			modelview = modelview.Mul4(mgl.Translate3D(rcx-float32(sys.scrrect[2])/2.0-xOffset, rcy-float32(sys.scrrect[3])/2.0+yOffset, fLength))
			modelview = modelview.Mul4(matrix)
			modelview = modelview.Mul4(mgl.Frustum(-float32(sys.scrrect[2])/2/fLength, float32(sys.scrrect[2])/2/fLength, -float32(sys.scrrect[3])/2/fLength, float32(sys.scrrect[3])/2/fLength, 1.0, 65535))
			modelview = modelview.Mul4(mgl.Translate3D(xOffset, -yOffset, -fLength))
		}

		modelview = modelview.Mul4(mgl.Scale3D(1, vs, 1))
		modelview = modelview.Mul4(
			mgl.Rotate3DX(-xagl * math.Pi / 180.0).Mul3(
			mgl.Rotate3DY(yagl * math.Pi / 180.0)).Mul3(
			mgl.Rotate3DZ(agl * math.Pi / 180.0)).Mat4())
		modelview = modelview.Mul4(mgl.Translate3D(-rcx, -rcy, 0))

		drawQuads(s, modelview, x1, y1, x2, y2, x3, y3, x4, y4)
		return
	}
	if (*tl)[3] == 1 && xbs != 0 {
		x1d, y1d, x2d, y2d, x3d, y3d, x4d, y4d := x1, y1, x2, y2, x3, y3, x4, y4
		for {
			x1d, y1d = x4d, y4d+ys*vs*float32((*tl)[1])
			x2d, y2d = x3d, y1d
			x3d = x4d - rxadd*ys*float32(h) + (xts/xbs)*(x3d-x4d)
			y3d = y2d + ys*vs*float32(h)
			x4d = x4d - rxadd*ys*float32(h)
			if AbsF(y3d-y4d) < 0.01 {
				break
			}
			y4d = y3d
			if ys*(float32(h)+float32((*tl)[1])) < 0 {
				if y1d <= float32(-sys.scrrect[3]) && y4d <= float32(-sys.scrrect[3]) {
					break
				}
			} else if y1d >= 0 && y4d >= 0 {
				break
			}
			if (0 > y1d || 0 > y4d) &&
				(y1d > float32(-sys.scrrect[3]) || y4d > float32(-sys.scrrect[3])) {
				rmTileHSub(s, modelview, x1d, y1d, x2d, y2d, x3d, y3d, x4d, y4d, x3d-x4d, x2d-x1d,
					(x3d-x4d)/float32(w), (x2d-x1d)/float32(w), tl, rcx)
			}
		}
	}
	if (*tl)[3] == 0 || xts != 0 {
		n := (*tl)[3]
		for {
			if ys*(float32(h)+float32((*tl)[1])) > 0 {
				if y1 <= float32(-sys.scrrect[3]) && y4 <= float32(-sys.scrrect[3]) {
					break
				}
			} else if y1 >= 0 && y4 >= 0 {
				break
			}
			if (0 > y1 || 0 > y4) &&
				(y1 > float32(-sys.scrrect[3]) || y4 > float32(-sys.scrrect[3])) {
				rmTileHSub(s, modelview, x1, y1, x2, y2, x3, y3, x4, y4, x3-x4, x2-x1,
					(x3-x4)/float32(w), (x2-x1)/float32(w), tl, rcx)
			}
			if (*tl)[3] != 1 && n != 0 {
				n--
			}
			if n == 0 {
				break
			}
			x4, y4 = x1, y1-ys*vs*float32((*tl)[1])
			x3, y3 = x2, y4
			x2 = x1 + rxadd*ys*float32(h) + (xbs/xts)*(x2-x1)
			y2 = y3 - ys*vs*float32(h)
			x1 = x1 + rxadd*ys*float32(h)
			if AbsF(y1-y2) < 0.01 {
				break
			}
			y1 = y2
		}
	}
}
func rmMainSub(s *Shader, size [2]uint16, x, y float32, tl *[4]int32,
	xts, xbs, ys, vs, rxadd, agl, yagl, xagl float32, renderMode, trans int32, rcx, rcy float32, neg bool, color float32,
	padd, pmul *[3]float32, projectionMode int32, fLength, xOffset, yOffset float32) {

	proj := mgl.Ortho(0, float32(sys.scrrect[2]), 0, float32(sys.scrrect[3]), -65535, 65535)
	gl.UniformMatrix4fv(s.uProjection, 1, false, &proj[0])

	modelview := mgl.Translate3D(0, float32(sys.scrrect[3]), 0)
	switch {
	case trans == -1:
		gl.Uniform1f(s.uAlpha, 1)
		if renderMode == 1 {
			gl.BlendFunc(gl.SRC_ALPHA, gl.ONE)
		} else {
			gl.BlendFunc(gl.ONE, gl.ONE)
		}
		gl.BlendEquation(gl.FUNC_ADD)
		rmTileSub(s, modelview, size[0], size[1], x, y, tl, xts, xbs, ys, vs, rxadd,
			agl, yagl, xagl, rcx, rcy, projectionMode, fLength, xOffset, yOffset)
	case trans == -2:
		gl.Uniform1f(s.uAlpha, 1)
		gl.BlendFunc(gl.ONE, gl.ONE)
		gl.BlendEquation(gl.FUNC_REVERSE_SUBTRACT)
		rmTileSub(s, modelview, size[0], size[1], x, y, tl, xts, xbs, ys, vs, rxadd,
			agl, yagl, xagl, rcx, rcy, projectionMode, fLength, xOffset, yOffset)
	case trans <= 0:
	case trans < 255:
		gl.Uniform1f(s.uAlpha, float32(trans)/255)
		if renderMode == 1 {
			gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
		} else {
			gl.BlendFunc(gl.ONE, gl.ONE_MINUS_SRC_ALPHA)
		}
		gl.BlendEquation(gl.FUNC_ADD)
		rmTileSub(s, modelview, size[0], size[1], x, y, tl, xts, xbs, ys, vs, rxadd,
			agl, yagl, xagl, rcx, rcy, projectionMode, fLength, xOffset, yOffset)
	case trans < 512:
		gl.Uniform1f(s.uAlpha, 1)
		if renderMode == 1 {
			gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
		} else {
			gl.BlendFunc(gl.ONE, gl.ONE_MINUS_SRC_ALPHA)
		}
		gl.BlendEquation(gl.FUNC_ADD)
		rmTileSub(s, modelview, size[0], size[1], x, y, tl, xts, xbs, ys, vs, rxadd,
			agl, yagl, xagl, rcx, rcy, projectionMode, fLength, xOffset, yOffset)
	default:
		src, dst := trans&0xff, trans>>10&0xff
		aglOver := 0
		if dst < 255 {
			gl.Uniform1f(s.uAlpha, 1-float32(dst)/255)
			gl.BlendFunc(gl.ZERO, gl.ONE_MINUS_SRC_ALPHA)
			gl.BlendEquation(gl.FUNC_ADD)
			rmTileSub(s, modelview, size[0], size[1], x, y, tl, xts, xbs, ys, vs, rxadd,
				agl, yagl, xagl, rcx, rcy, projectionMode, fLength, xOffset, yOffset)
			aglOver++
		}
		if src > 0 {
			if aglOver != 0 {
				agl = 0
				yagl = 0
				xagl = 0
			}
			gl.Uniform1f(s.uAlpha, float32(src)/255)

			if renderMode == 1 {
				gl.BlendFunc(gl.SRC_ALPHA, gl.ONE)
			} else {
				gl.BlendFunc(gl.ONE, gl.ONE)
			}
			gl.BlendEquation(gl.FUNC_ADD)
			rmTileSub(s, modelview, size[0], size[1], x, y, tl, xts, xbs, ys, vs, rxadd,
				agl, yagl, xagl, rcx, rcy, projectionMode, fLength, xOffset, yOffset)
		}
	}
}
func rmInitSub(size [2]uint16, x, y *float32, tile *[4]int32, xts float32,
	ys, vs, agl, yagl, xagl *float32, window *[4]int32, rcx float32, rcy *float32) (
	tl [4]int32) {
	if *vs < 0 {
		*vs *= -1
		*ys *= -1
		*agl *= -1
		*xagl *= -1
	}
	tl = *tile
	if tl[2] == 0 {
		tl[0] = 0
	} else if tl[0] > 0 {
		tl[0] -= int32(size[0])
	}
	if tl[3] == 0 {
		tl[1] = 0
	} else if tl[1] > 0 {
		tl[1] -= int32(size[1])
	}
	if xts >= 0 {
		*x *= -1
	}
	*x += rcx
	*rcy *= -1
	if *ys < 0 {
		*y *= -1
	}
	*y += *rcy
	gl.Enable(gl.BLEND)
	gl.Enable(gl.SCISSOR_TEST)
	gl.Scissor((*window)[0], sys.scrrect[3]-((*window)[1]+(*window)[3]),
		(*window)[2], (*window)[3])
	return
}

func RenderMugenPal(tex Texture, mask int32, size [2]uint16,
	x, y float32, tile *[4]int32, xts, xbs, ys, vs, rxadd, agl, yagl, xagl float32,
	trans int32, window *[4]int32, rcx, rcy float32, neg bool, color float32,
	padd, pmul *[3]float32, projectionMode int32, fLength, xOffset, yOffset float32) {
	if tex == 0 || !IsFinite(x+y+xts+xbs+ys+vs+rxadd+agl+rcx+rcy) {
		return
	}
	tl := rmInitSub(size, &x, &y, tile, xts, &ys, &vs, &agl, &yagl, &xagl, window, rcx, &rcy)
	gl.UseProgram(mainShader.program)
	gl.Uniform1i(mainShader.uTexture, 0)
	gl.Uniform1i(mainShader.u["pal"], 1)
	gl.Uniform1i(mainShader.u["isRgba"], 0)
	gl.Uniform1i(mainShader.u["isTrapez"], Btoi(AbsF(AbsF(xts)-AbsF(xbs)) > 0.001))
	gl.Uniform1i(mainShader.u["mask"], mask)
	gl.Uniform1i(mainShader.u["neg"], Btoi(neg))
	gl.Uniform1f(mainShader.u["gray"], 1-color)
	gl.Uniform3f(mainShader.u["add"], (*padd)[0], (*padd)[1], (*padd)[2])
	gl.Uniform3f(mainShader.u["mul"], (*pmul)[0], (*pmul)[1], (*pmul)[2])
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, uint32(tex))
	rmMainSub(mainShader, size, x, y, &tl, xts, xbs, ys, vs, rxadd, agl, yagl, xagl,
		1, trans, rcx, rcy, neg, color, padd, pmul, projectionMode, fLength, xOffset, yOffset)
	gl.Disable(gl.SCISSOR_TEST)
	gl.Disable(gl.BLEND)
}

func RenderMugen(tex Texture, pal []uint32, mask int32, size [2]uint16,
	x, y float32, tile *[4]int32, xts, xbs, ys, vs, rxadd, agl, yagl, xagl float32,
	trans int32, window *[4]int32, rcx, rcy float32, projectionMode int32, fLength, xOffset, yOffset float32) {
	gl.ActiveTexture(gl.TEXTURE1)
	var paltex uint32
	gl.GenTextures(1, &paltex)
	gl.BindTexture(gl.TEXTURE_2D, paltex)
	gl.PixelStorei(gl.UNPACK_ALIGNMENT, 1)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, 256, 1, 0, gl.RGBA, gl.UNSIGNED_BYTE,
		unsafe.Pointer(&pal[0]))
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	RenderMugenPal(tex, mask, size, x, y, tile, xts, xbs, ys, vs, rxadd,
		agl, yagl, xagl, trans, window, rcx, rcy, false, 1, &[3]float32{0, 0, 0}, &[3]float32{1, 1, 1}, projectionMode, fLength, xOffset, yOffset)
	gl.DeleteTextures(1, &paltex)
}

func RenderMugenFc(tex Texture, size [2]uint16, x, y float32,
	tile *[4]int32, xts, xbs, ys, vs, rxadd, agl, yagl, xagl float32, trans int32,
	window *[4]int32, rcx, rcy float32, neg bool, color float32,
	padd, pmul *[3]float32, projectionMode int32, fLength, xOffset, yOffset float32) {
	if tex == 0 || !IsFinite(x+y+xts+xbs+ys+vs+rxadd+agl+rcx+rcy) {
		return
	}
	tl := rmInitSub(size, &x, &y, tile, xts, &ys, &vs, &agl, &yagl, &xagl, window, rcx, &rcy)
	gl.UseProgram(mainShader.program)
	gl.Uniform1i(mainShader.u["isRgba"], 1)
	gl.Uniform1i(mainShader.u["isTrapez"], Btoi(AbsF(AbsF(xts)-AbsF(xbs)) > 0.001))
	gl.Uniform1i(mainShader.u["neg"], Btoi(neg))
	gl.Uniform1f(mainShader.u["gray"], 1-color)
	gl.Uniform3f(mainShader.u["add"], (*padd)[0], (*padd)[1], (*padd)[2])
	gl.Uniform3f(mainShader.u["mul"], (*pmul)[0], (*pmul)[1], (*pmul)[2])
	gl.BindTexture(gl.TEXTURE_2D, uint32(tex))
	rmMainSub(mainShader, size, x, y, &tl, xts, xbs, ys, vs, rxadd, agl, yagl, xagl,
		2, trans, rcx, rcy, neg, color, padd, pmul, projectionMode, fLength, xOffset, yOffset)
	gl.Disable(gl.SCISSOR_TEST)
	gl.Disable(gl.BLEND)
}
func RenderMugenFcS(tex Texture, size [2]uint16, x, y float32,
	tile *[4]int32, xts, xbs, ys, vs, rxadd, agl, yagl, xagl float32, trans int32,
	window *[4]int32, rcx, rcy float32, color uint32, projectionMode int32, fLength, xOffset, yOffset float32) {
	if tex == 0 || !IsFinite(x+y+xts+xbs+ys+vs+rxadd+agl+rcx+rcy) {
		return
	}
	tl := rmInitSub(size, &x, &y, tile, xts, &ys, &vs, &agl, &yagl, &xagl, window, rcx, &rcy)
	gl.UseProgram(flatShader.program)
	gl.Uniform1i(flatShader.uTexture, 0)
	gl.Uniform3f(
		flatShader.u["color"], float32(color>>16&0xff)/255, float32(color>>8&0xff)/255,
		float32(color&0xff)/255)
	gl.Uniform1i(flatShader.u["isShadow"], 1)
	gl.BindTexture(gl.TEXTURE_2D, uint32(tex))
	rmMainSub(flatShader, size, x, y, &tl, xts, xbs, ys, vs, rxadd, agl, yagl, xagl,
		0, trans, rcx, rcy, false, 1, &[3]float32{0, 0, 0}, &[3]float32{1, 1, 1}, projectionMode, fLength, xOffset, yOffset)
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
	gl.UniformMatrix4fv(flatShader.uModelView, 1, false, &modelview[0])
	gl.UniformMatrix4fv(flatShader.uProjection, 1, false, &proj[0])
	gl.Uniform1i(flatShader.uTexture, 0)
	gl.Uniform3f(flatShader.u["color"], r, g, b)
	gl.Uniform1i(flatShader.u["isShadow"], 0)

	x1, y1 := float32(rect[0]), -float32(rect[1])
	x2, y2 := float32(rect[0]+rect[2]), -float32(rect[1]+rect[3])
	vertexPosition := [8]float32{x2, y2, x2, y1, x1, y2, x1, y1}
	gl.EnableVertexAttribArray(uint32(flatShader.aPos))
	gl.VertexAttribPointer(uint32(flatShader.aPos), 2, gl.FLOAT, false, 0, unsafe.Pointer(&vertexPosition[0]))

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
}` + "\x00"

var identFragShader string = `
uniform sampler2D Texture;

varying vec2 texcoord;

void main(void) {
	gl_FragColor = texture2D(Texture, texcoord);
}` + "\x00"
