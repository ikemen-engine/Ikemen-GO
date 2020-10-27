package main

import (
	"unsafe"

	"github.com/go-gl/gl/v2.1/gl"
)

var notiling = [4]int32{0, 0, 0, 0}
var mugenShader uintptr
var uniformA, uniformPal, uniformMsk, uniformPalNeg, uniformPalGray, uniformPalAdd, uniformPalMul int32
var uniformPalX1x2x4x3, uniformPalIsTrapez int32
var mugenShaderFc uintptr
var uniformFcA, uniformNeg, uniformGray, uniformAdd, uniformMul int32
var uniformX1x2x4x3, uniformIsTrapez int32
var mugenShaderFcS uintptr
var uniformFcSA, uniformColor int32
var posattLocation, uvattLocation int32
var vertexUv = [8]float32{0, 1, 1, 1, 1, 0, 0, 0}
var indices = [4]int32{1, 2, 0, 3}

// Post-processing
var fbo, fbo_texture uint32

// Clasic AA
var rbo_depth uint32

// MSAA
var fbo_f, fbo_f_texture uint32

var postShader uintptr
var postVertAttrib int32
var postTexUniform int32
var postTexSizeUniform int32
var postVertices = [8]float32{-1, -1, 1, -1, -1, 1, 1, 1}

var postShaderSelect []uintptr //[4]uintptr

func RenderInit() {
	vertShader := "attribute vec2 position;" +
		"attribute vec2 uv;" +
		"void main(void){" +
		"gl_TexCoord[0] = gl_TextureMatrix[0] * vec4(uv, 0.0, 1.0);" +
		"gl_Position = gl_ModelViewProjectionMatrix * vec4(position, 0.0, 1.0);" +
		"}\x00"
	fragShader := "uniform float a;" +
		"uniform sampler2D tex;" +
		"uniform sampler1D pal;" +
		"uniform int msk;" +
		"uniform bool neg;" +
		"uniform float gray;" +
		"uniform vec3 add;" +
		"uniform vec3 mul;" +
		"uniform vec4 x1x2x4x3;" +
		"uniform bool isTrapez;" +
		"void main(void){" +
		"vec2 texcoord = gl_TexCoord[0].st;" +
		"if(isTrapez){" +
		"float y = 1.0 - gl_TexCoord[0].t;" + // ここから台形用のテクスチャ座標計算/ Compute texture coordinates for trapezoid from here
		"float left = (x1x2x4x3[2] - x1x2x4x3[0]) * y + x1x2x4x3[0];" +
		"float right = (x1x2x4x3[3] - x1x2x4x3[1]) * y + x1x2x4x3[1];" +
		"left = (gl_FragCoord.x - left);" +
		"right = (right - gl_FragCoord.x);" +
		"texcoord[0] = left / (left + right);" + // ここまで / To this point
		"}" +
		"float r = texture2D(tex, texcoord).r;" +
		"if(int(255.25*r) == msk){" +
		"	gl_FragColor = vec4(0.0);" +
		"}else{" +
		"	vec4 c = texture1D(pal, r*0.9961);" +
		"	if(neg) c.rgb = vec3(1.0) - c.rgb;" +
		"	c.rgb += (vec3((c.r + c.g + c.b) / 3.0) - c.rgb) * gray + add;" +
		"	gl_FragColor = vec4(c.rgb * mul, c.a * a);" +
		"}" +
		"}\x00"
	fragShaderFc := "uniform float a;" +
		"uniform sampler2D tex;" +
		"uniform bool neg;" +
		"uniform float gray;" +
		"uniform vec3 add;" +
		"uniform vec3 mul;" +
		"uniform vec4 x1x2x4x3;" +
		"uniform bool isTrapez;" +
		"void main(void){" +
		"vec2 texcoord = gl_TexCoord[0].st;" +
		"if(isTrapez){" +
		"float y = 1.0 - gl_TexCoord[0].t;" + // ここから台形用のテクスチャ座標計算 / Compute texture coordinates for trapezoid from here
		"float left = (x1x2x4x3[2] - x1x2x4x3[0]) * y + x1x2x4x3[0];" +
		"float right = (x1x2x4x3[3] - x1x2x4x3[1]) * y + x1x2x4x3[1];" +
		"left = (gl_FragCoord.x - left);" +
		"right = (right - gl_FragCoord.x);" +
		"texcoord[0] = left / (left + right);" + // ここまで / To this point
		"}" +
		"vec4 c = texture2D(tex, texcoord);" +
		"if(neg) c.rgb = vec3(1.0 * c.a) - c.rgb;" +
		"c.rgb += (vec3((c.r + c.g + c.b) / 3.0) - c.rgb) * gray + add * c.a;" +
		"c.rgb *= mul;" +
		"c.a *= a;" +
		"gl_FragColor = c;" +
		"}\x00"
	fragShaderFcS := "uniform float a;" +
		"uniform sampler2D tex;" +
		"uniform vec3 color;" +
		"void main(void){" +
		"vec4 c = texture2D(tex, gl_TexCoord[0].st);" +
		"c.rgb = color * c.a;" +
		"c.a *= a;" +
		"gl_FragColor = c;" +
		"}\x00"
	errLog := func(obl uintptr) error {
		var size int32
		gl.GetObjectParameterivARB(obl, gl.INFO_LOG_LENGTH, &size)
		if size <= 0 {
			return nil
		}
		var l int32
		str := make([]byte, size+1)
		gl.GetInfoLogARB(obl, size, &l, &str[0])
		return Error(str[:l])
	}
	compile := func(shaderType uint32, src string) (shader uintptr) {
		shader = gl.CreateShaderObjectARB(shaderType)
		s, _ := gl.Strs(src)
		var l int32 = int32(len(src) - 1)
		gl.ShaderSourceARB(shader, 1, s, &l)
		gl.CompileShaderARB(shader)
		var ok int32
		gl.GetObjectParameterivARB(shader, gl.OBJECT_COMPILE_STATUS_ARB, &ok)
		if ok == 0 {
			chk(errLog(shader))
			panic(Error("Shader compile error"))
		}
		return
	}
	link := func(v, f uintptr) (program uintptr) {
		program = gl.CreateProgramObjectARB()
		gl.AttachObjectARB(program, v)
		gl.AttachObjectARB(program, f)
		gl.LinkProgramARB(program)
		var ok int32
		gl.GetObjectParameterivARB(program, gl.OBJECT_LINK_STATUS_ARB, &ok)
		if ok == 0 {
			chk(errLog(program))
			panic(Error("Link error"))
		}
		return
	}
	vertObj := compile(gl.VERTEX_SHADER, vertShader)
	fragObj := compile(gl.FRAGMENT_SHADER, fragShader)
	mugenShader = link(vertObj, fragObj)
	posattLocation = gl.GetAttribLocationARB(mugenShader, gl.Str("position\x00"))
	uvattLocation = gl.GetAttribLocationARB(mugenShader, gl.Str("uv\x00"))
	uniformA = gl.GetUniformLocationARB(mugenShader, gl.Str("a\x00"))
	uniformPal = gl.GetUniformLocationARB(mugenShader, gl.Str("pal\x00"))
	uniformMsk = gl.GetUniformLocationARB(mugenShader, gl.Str("msk\x00"))
	uniformPalNeg = gl.GetUniformLocationARB(mugenShader, gl.Str("neg\x00"))
	uniformPalGray = gl.GetUniformLocationARB(mugenShader, gl.Str("gray\x00"))
	uniformPalAdd = gl.GetUniformLocationARB(mugenShader, gl.Str("add\x00"))
	uniformPalMul = gl.GetUniformLocationARB(mugenShader, gl.Str("mul\x00"))
	uniformPalX1x2x4x3 = gl.GetUniformLocationARB(mugenShader, gl.Str("x1x2x4x3\x00"))
	uniformPalIsTrapez = gl.GetUniformLocationARB(mugenShader, gl.Str("isTrapez\x00"))
	gl.DeleteObjectARB(fragObj)
	fragObj = compile(gl.FRAGMENT_SHADER, fragShaderFc)
	mugenShaderFc = link(vertObj, fragObj)
	uniformFcA = gl.GetUniformLocationARB(mugenShaderFc, gl.Str("a\x00"))
	uniformNeg = gl.GetUniformLocationARB(mugenShaderFc, gl.Str("neg\x00"))
	uniformGray = gl.GetUniformLocationARB(mugenShaderFc, gl.Str("gray\x00"))
	uniformAdd = gl.GetUniformLocationARB(mugenShaderFc, gl.Str("add\x00"))
	uniformMul = gl.GetUniformLocationARB(mugenShaderFc, gl.Str("mul\x00"))
	uniformX1x2x4x3 = gl.GetUniformLocationARB(mugenShaderFc, gl.Str("x1x2x4x3\x00"))
	uniformIsTrapez = gl.GetUniformLocationARB(mugenShaderFc, gl.Str("isTrapez\x00"))
	gl.DeleteObjectARB(fragObj)
	fragObj = compile(gl.FRAGMENT_SHADER, fragShaderFcS)
	mugenShaderFcS = link(vertObj, fragObj)
	uniformFcSA = gl.GetUniformLocationARB(mugenShader, gl.Str("a\x00"))
	uniformColor = gl.GetUniformLocationARB(mugenShaderFcS, gl.Str("color\x00"))
	gl.DeleteObjectARB(fragObj)
	gl.DeleteObjectARB(vertObj)

	// Compile postprocessing shaders

	// Calculate total ammount of shaders loaded.
	postShaderSelect = make([]uintptr, 4+len(sys.externalShaderList))

	// [0]: ident shader(no postprocessing)
	vertObj = compile(gl.VERTEX_SHADER, identVertShader)
	fragObj = compile(gl.FRAGMENT_SHADER, identFragShader)
	postShaderSelect[0] = link(vertObj, fragObj)
	gl.DeleteObjectARB(vertObj)
	gl.DeleteObjectARB(fragObj)

	// [1]: hqx2 shader
	vertObj = compile(gl.VERTEX_SHADER, hqx2VertShader)
	fragObj = compile(gl.FRAGMENT_SHADER, hqx2FragShader)
	postShaderSelect[1] = link(vertObj, fragObj)
	gl.DeleteObjectARB(vertObj)
	gl.DeleteObjectARB(fragObj)

	// [2]: hqx4 shader
	vertObj = compile(gl.VERTEX_SHADER, hqx4VertShader)
	fragObj = compile(gl.FRAGMENT_SHADER, hqx4FragShader)
	postShaderSelect[2] = link(vertObj, fragObj)
	gl.DeleteObjectARB(vertObj)
	gl.DeleteObjectARB(fragObj)

	// [3]: scanline shader
	vertObj = compile(gl.VERTEX_SHADER, scanlineVertShader)
	fragObj = compile(gl.FRAGMENT_SHADER, scanlineFragShader)
	postShaderSelect[3] = link(vertObj, fragObj)
	gl.DeleteObjectARB(vertObj)
	gl.DeleteObjectARB(fragObj)

	// External Shaders
	for i := 0; i < len(sys.externalShaderList); i++ {
		vertObj = compile(gl.VERTEX_SHADER, sys.externalShaders[0][i])
		fragObj = compile(gl.FRAGMENT_SHADER, sys.externalShaders[1][i])
		postShaderSelect[4+i] = link(vertObj, fragObj)
		gl.DeleteObjectARB(vertObj)
		gl.DeleteObjectARB(fragObj)
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

	postVertAttrib = gl.GetAttribLocationARB(postShader, gl.Str("VertCoord\x00"))
	postTexUniform = gl.GetUniformLocationARB(postShader, gl.Str("Texture\x00"))
	postTexSizeUniform = gl.GetUniformLocationARB(postShader, gl.Str("TextureSize\x00"))

	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	gl.UseProgramObjectARB(postShader)

	if sys.multisampleAntialiasing {
		gl.BindTexture(gl.TEXTURE_2D, fbo_f_texture)
	} else {
		gl.BindTexture(gl.TEXTURE_2D, fbo_texture)
	}

	gl.Uniform1iARB(postTexUniform, 0)
	gl.Uniform2fARB(postTexSizeUniform, float32(sys.scrrect[2]), float32(sys.scrrect[3]))
	gl.EnableVertexAttribArrayARB(uint32(postVertAttrib))
	gl.VertexAttribPointerARB(uint32(postVertAttrib), 2, gl.FLOAT, false, 0, unsafe.Pointer(&postVertices[0]))
	gl.DrawArrays(gl.TRIANGLE_STRIP, 0, 4)
	gl.UseProgramObjectARB(0)
}

func drawQuads(x1, y1, x2, y2, x3, y3, x4, y4 float32, renderMode int32) {
	vertexPosition := [8]float32{x1, y1, x2, y2, x3, y3, x4, y4}
	switch renderMode {
	case 1:
		gl.Uniform4fARB(uniformPalX1x2x4x3, x1, x2, x4, x3)
	case 2:
		gl.Uniform4fARB(uniformX1x2x4x3, x1, x2, x4, x3)
	}
	gl.EnableVertexAttribArrayARB(uint32(posattLocation))
	gl.EnableVertexAttribArrayARB(uint32(uvattLocation))
	gl.VertexAttribPointerARB(uint32(posattLocation), 2, gl.FLOAT, false, 0, unsafe.Pointer(&vertexPosition[0]))
	gl.VertexAttribPointerARB(uint32(uvattLocation), 2, gl.FLOAT, false, 0, unsafe.Pointer(&vertexUv[0]))

	gl.DrawElements(gl.TRIANGLE_STRIP, 4, gl.UNSIGNED_INT, unsafe.Pointer(&indices))
}
func rmTileHSub(x1, y1, x2, y2, x3, y3, x4, y4, xtw, xbw, xts, xbs float32,
	tl *[4]int32, rcx float32, renderMode int32) {
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
					drawQuads(x1d, y1, x2d, y2, x3d, y3, x4d, y4, renderMode)
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
			drawQuads(x1, y1, x2, y2, x3, y3, x4, y4, renderMode)
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
func rmTileSub(w, h uint16, x, y float32, tl *[4]int32, renderMode int32,
	xts, xbs, ys, vs, rxadd, agl, yagl, xagl, rcx, rcy float32) {
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
		gl.Translated(float64(rcx), float64(rcy), 0)
		gl.Scaled(1, float64(vs), 1)
		gl.Rotated(float64(xagl), 1.0, 0.0, 0.0)
		gl.Rotated(float64(-yagl), 0.0, 1.0, 0.0)
		gl.Rotated(float64(agl), 0.0, 0.0, 1.0)
		gl.Translated(float64(-rcx), float64(-rcy), 0)
		drawQuads(x1, y1, x2, y2, x3, y3, x4, y4, renderMode)
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
				rmTileHSub(x1d, y1d, x2d, y2d, x3d, y3d, x4d, y4d, x3d-x4d, x2d-x1d,
					(x3d-x4d)/float32(w), (x2d-x1d)/float32(w), tl,
					rcx, renderMode)
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
				rmTileHSub(x1, y1, x2, y2, x3, y3, x4, y4, x3-x4, x2-x1,
					(x3-x4)/float32(w), (x2-x1)/float32(w), tl, rcx, renderMode)
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
func rmMainSub(a int32, size [2]uint16, x, y float32, tl *[4]int32,
	xts, xbs, ys, vs, rxadd, agl, yagl, xagl float32, renderMode, trans int32, rcx, rcy float32, neg bool, color float32,
	padd, pmul *[3]float32) {
	gl.MatrixMode(gl.PROJECTION)
	gl.PushMatrix()
	gl.LoadIdentity()
	gl.Ortho(0, float64(sys.scrrect[2]), 0, float64(sys.scrrect[3]), -65535, 65535)
	gl.MatrixMode(gl.MODELVIEW)
	gl.PushMatrix()
	gl.Translated(0, float64(sys.scrrect[3]), 0)
	switch {
	case trans == -1:
		gl.Uniform1fARB(a, 1)
		gl.BlendFunc(gl.SRC_ALPHA, gl.ONE)
		gl.BlendEquation(gl.FUNC_ADD)
		rmTileSub(size[0], size[1], x, y, tl, renderMode, xts, xbs, ys, vs, rxadd,
			agl, yagl, xagl, rcx, rcy)
	case trans == -2:
		gl.Uniform1fARB(a, 1)
		gl.BlendFunc(gl.ONE, gl.ONE)
		gl.BlendEquation(gl.FUNC_REVERSE_SUBTRACT)
		rmTileSub(size[0], size[1], x, y, tl, renderMode, xts, xbs, ys, vs, rxadd,
			agl, yagl, xagl, rcx, rcy)
	case trans <= 0:
	case trans < 255:
		gl.Uniform1fARB(a, float32(trans)/255)
		gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
		gl.BlendEquation(gl.FUNC_ADD)
		rmTileSub(size[0], size[1], x, y, tl, renderMode, xts, xbs, ys, vs, rxadd,
			agl, yagl, xagl, rcx, rcy)
	case trans < 512:
		gl.Uniform1fARB(a, 1)
		if renderMode == 1 {
			gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
		} else {
			gl.BlendFunc(gl.ONE, gl.ONE_MINUS_SRC_ALPHA)
		}
		gl.BlendEquation(gl.FUNC_ADD)
		rmTileSub(size[0], size[1], x, y, tl, renderMode, xts, xbs, ys, vs, rxadd,
			agl, yagl, xagl, rcx, rcy)
	default:
		src, dst := trans&0xff, trans>>10&0xff
		aglOver := 0
		if dst < 255 {
			gl.Uniform1fARB(a, 1-float32(dst)/255)
			gl.BlendFunc(gl.ZERO, gl.ONE_MINUS_SRC_ALPHA)
			gl.BlendEquation(gl.FUNC_ADD)
			rmTileSub(size[0], size[1], x, y, tl, renderMode, xts, xbs, ys, vs, rxadd,
				agl, yagl, xagl, rcx, rcy)
			aglOver++
		}
		if src > 0 {
			if aglOver != 0 {
				agl = 0
				yagl = 0
				xagl = 0
			}
			gl.Uniform1fARB(a, float32(src)/255)
			gl.BlendFunc(gl.SRC_ALPHA, gl.ONE)
			gl.BlendEquation(gl.FUNC_ADD)
			rmTileSub(size[0], size[1], x, y, tl, renderMode, xts, xbs, ys, vs, rxadd,
				agl, yagl, xagl, rcx, rcy)
		}
	}
	gl.PopMatrix()
	gl.MatrixMode(gl.PROJECTION)
	gl.PopMatrix()
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
	gl.Enable(gl.TEXTURE_2D)
	gl.Enable(gl.SCISSOR_TEST)
	gl.Scissor((*window)[0], sys.scrrect[3]-((*window)[1]+(*window)[3]),
		(*window)[2], (*window)[3])
	return
}

func RenderMugenPal(tex Texture, mask int32, size [2]uint16,
	x, y float32, tile *[4]int32, xts, xbs, ys, vs, rxadd, agl, yagl, xagl float32,
	trans int32, window *[4]int32, rcx, rcy float32, neg bool, color float32,
	padd, pmul *[3]float32) {
	if tex == 0 || !IsFinite(x+y+xts+xbs+ys+vs+rxadd+agl+rcx+rcy) {
		return
	}
	tl := rmInitSub(size, &x, &y, tile, xts, &ys, &vs, &agl, &yagl, &xagl, window, rcx, &rcy)
	ineg := int32(0)
	if neg {
		ineg = 1
	}
	isTrapez := int32(0)
	if AbsF(AbsF(xts)-AbsF(xbs)) > 0.001 {
		isTrapez = 1
	}
	gl.UseProgramObjectARB(mugenShader)
	gl.Uniform1iARB(uniformPal, 1)
	gl.Uniform1iARB(uniformMsk, mask)
	gl.Uniform1iARB(uniformPalNeg, ineg)
	gl.Uniform1fARB(uniformPalGray, 1-color)
	gl.Uniform3fARB(uniformPalAdd, (*padd)[0], (*padd)[1], (*padd)[2])
	gl.Uniform3fARB(uniformPalMul, (*pmul)[0], (*pmul)[1], (*pmul)[2])
	gl.Uniform1iARB(uniformPalIsTrapez, isTrapez)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, uint32(tex))
	rmMainSub(uniformA, size, x, y, &tl, xts, xbs, ys, vs, rxadd, agl, yagl, xagl,
		1, trans, rcx, rcy, neg, color, padd, pmul)
	gl.UseProgramObjectARB(0)
	gl.Disable(gl.SCISSOR_TEST)
	gl.Disable(gl.TEXTURE_2D)
	gl.Disable(gl.BLEND)
}

func RenderMugen(tex Texture, pal []uint32, mask int32, size [2]uint16,
	x, y float32, tile *[4]int32, xts, xbs, ys, vs, rxadd, agl, yagl, xagl float32,
	trans int32, window *[4]int32, rcx, rcy float32) {
	gl.Enable(gl.TEXTURE_1D)
	gl.ActiveTexture(gl.TEXTURE1)
	var paltex uint32
	gl.GenTextures(1, &paltex)
	gl.BindTexture(gl.TEXTURE_1D, paltex)
	gl.PixelStorei(gl.UNPACK_ALIGNMENT, 1)
	gl.TexImage1D(gl.TEXTURE_1D, 0, gl.RGBA, 256, 0, gl.RGBA, gl.UNSIGNED_BYTE,
		unsafe.Pointer(&pal[0]))
	gl.TexParameteri(gl.TEXTURE_1D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_1D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	RenderMugenPal(tex, mask, size, x, y, tile, xts, xbs, ys, vs, rxadd,
		agl, yagl, xagl, trans, window, rcx, rcy, false, 1, &[3]float32{0, 0, 0}, &[3]float32{1, 1, 1})
	gl.DeleteTextures(1, &paltex)
	gl.Disable(gl.TEXTURE_1D)
}

func RenderMugenFc(tex Texture, size [2]uint16, x, y float32,
	tile *[4]int32, xts, xbs, ys, vs, rxadd, agl, yagl, xagl float32, trans int32,
	window *[4]int32, rcx, rcy float32, neg bool, color float32,
	padd, pmul *[3]float32) {
	if tex == 0 || !IsFinite(x+y+xts+xbs+ys+vs+rxadd+agl+rcx+rcy) {
		return
	}
	tl := rmInitSub(size, &x, &y, tile, xts, &ys, &vs, &agl, &yagl, &xagl, window, rcx, &rcy)
	gl.UseProgramObjectARB(mugenShaderFc)
	ineg := int32(0)
	if neg {
		ineg = 1
	}
	isTrapez := int32(0)
	if AbsF(AbsF(xts)-AbsF(xbs)) > 0.001 {
		isTrapez = 1
	}
	gl.Uniform1iARB(uniformNeg, ineg)
	gl.Uniform1fARB(uniformGray, 1-color)
	gl.Uniform3fARB(uniformAdd, (*padd)[0], (*padd)[1], (*padd)[2])
	gl.Uniform3fARB(uniformMul, (*pmul)[0], (*pmul)[1], (*pmul)[2])
	gl.Uniform1iARB(uniformIsTrapez, isTrapez)
	gl.BindTexture(gl.TEXTURE_2D, uint32(tex))
	rmMainSub(uniformFcA, size, x, y, &tl, xts, xbs, ys, vs, rxadd, agl, yagl, xagl,
		2, trans, rcx, rcy, neg, color, padd, pmul)
	gl.UseProgramObjectARB(0)
	gl.Disable(gl.SCISSOR_TEST)
	gl.Disable(gl.TEXTURE_2D)
	gl.Disable(gl.BLEND)
}
func RenderMugenFcS(tex Texture, size [2]uint16, x, y float32,
	tile *[4]int32, xts, xbs, ys, vs, rxadd, agl, yagl, xagl float32, trans int32,
	window *[4]int32, rcx, rcy float32, color uint32) {
	if tex == 0 || !IsFinite(x+y+xts+xbs+ys+vs+rxadd+agl+rcx+rcy) {
		return
	}
	tl := rmInitSub(size, &x, &y, tile, xts, &ys, &vs, &agl, &yagl, &xagl, window, rcx, &rcy)
	gl.UseProgramObjectARB(mugenShaderFcS)
	gl.Uniform3fARB(
		uniformColor, float32(color>>16&0xff)/255, float32(color>>8&0xff)/255,
		float32(color&0xff)/255)
	gl.BindTexture(gl.TEXTURE_2D, uint32(tex))
	rmMainSub(uniformFcSA, size, x, y, &tl, xts, xbs, ys, vs, rxadd, agl, yagl, xagl,
		0, trans, rcx, rcy, false, 1, &[3]float32{0, 0, 0}, &[3]float32{1, 1, 1})
	gl.UseProgramObjectARB(0)
	gl.Disable(gl.SCISSOR_TEST)
	gl.Disable(gl.TEXTURE_2D)
	gl.Disable(gl.BLEND)
}
func FillRect(rect [4]int32, color uint32, trans int32) {
	r := float32(color>>16&0xff) / 255
	g := float32(color>>8&0xff) / 255
	b := float32(color&0xff) / 255
	fill := func(a float32) {
		gl.Begin(gl.QUADS)
		gl.Color4f(r, g, b, a)
		gl.Vertex2f(float32(rect[0]), -float32(rect[1]+rect[3]))
		gl.Vertex2f(float32(rect[0]+rect[2]), -float32(rect[1]+rect[3]))
		gl.Vertex2f(float32(rect[0]+rect[2]), -float32(rect[1]))
		gl.Vertex2f(float32(rect[0]), -float32(rect[1]))
		gl.End()
	}
	gl.Enable(gl.BLEND)
	gl.MatrixMode(gl.PROJECTION)
	gl.PushMatrix()
	gl.LoadIdentity()
	gl.Ortho(0, float64(sys.scrrect[2]), 0, float64(sys.scrrect[3]), -65535, 65535)
	gl.MatrixMode(gl.MODELVIEW)
	gl.PushMatrix()
	gl.Translated(0, float64(sys.scrrect[3]), 0)
	if trans == -1 {
		gl.BlendFunc(gl.SRC_ALPHA, gl.ONE)
		gl.BlendEquation(gl.FUNC_ADD)
		fill(1)
	} else if trans == -2 {
		gl.BlendFunc(gl.ZERO, gl.ONE_MINUS_SRC_COLOR)
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
	gl.PopMatrix()
	gl.MatrixMode(gl.PROJECTION)
	gl.PopMatrix()
	gl.Disable(gl.BLEND)
}

var identVertShader string = `
attribute vec2 VertCoord;
uniform vec2 TextureSize;

void main()
{
	gl_Position = vec4(VertCoord, 0.0, 1.0);
	gl_TexCoord[0].xy = (VertCoord + 1.0) / 2.0;
}` + "\x00"

var identFragShader string = `
uniform sampler2D Texture;

void main(void) {
	gl_FragColor = texture2D(Texture, gl_TexCoord[0].xy);
}` + "\x00"

var hqx2VertShader string = `
attribute vec2 VertCoord;
uniform vec2 TextureSize;

void main()
{
	gl_Position = vec4(VertCoord, 0.0, 1.0);
	
	vec2 TexCoord = (VertCoord + 1.0) / 2.0;

	float x = 0.5 * (1.0 / TextureSize.x);
    float y = 0.5 * (1.0 / TextureSize.y);
    vec2 dg1 = vec2( x, y);
    vec2 dg2 = vec2(-x, y);
    vec2 dx = vec2(x, 0.0);
	vec2 dy = vec2(0.0, y);
	
	gl_TexCoord[0].xy = TexCoord;
    gl_TexCoord[1].xy = gl_TexCoord[0].xy - dg1;
    gl_TexCoord[1].zw = gl_TexCoord[0].xy - dy;
    gl_TexCoord[2].xy = gl_TexCoord[0].xy - dg2;
    gl_TexCoord[2].zw = gl_TexCoord[0].xy + dx;
    gl_TexCoord[3].xy = gl_TexCoord[0].xy + dg1;
    gl_TexCoord[3].zw = gl_TexCoord[0].xy + dy;
    gl_TexCoord[4].xy = gl_TexCoord[0].xy + dg2;
	gl_TexCoord[4].zw = gl_TexCoord[0].xy - dx;
}` + "\x00"

var hqx2FragShader string = `
uniform sampler2D Texture;

const float mx = 0.325;      // start smoothing wt.
  const float k = -0.250;      // wt. decrease factor
  const float max_w = 0.25;    // max filter weigth
  const float min_w =-0.05;    // min filter weigth
  const float lum_add = 0.25;  // effects smoothing

  void main() {
    vec3 c00 = texture2D(Texture, gl_TexCoord[1].xy).xyz;
    vec3 c10 = texture2D(Texture, gl_TexCoord[1].zw).xyz;
    vec3 c20 = texture2D(Texture, gl_TexCoord[2].xy).xyz;
    vec3 c01 = texture2D(Texture, gl_TexCoord[4].zw).xyz;
    vec3 c11 = texture2D(Texture, gl_TexCoord[0].xy).xyz;
    vec3 c21 = texture2D(Texture, gl_TexCoord[2].zw).xyz;
    vec3 c02 = texture2D(Texture, gl_TexCoord[4].xy).xyz;
    vec3 c12 = texture2D(Texture, gl_TexCoord[3].zw).xyz;
    vec3 c22 = texture2D(Texture, gl_TexCoord[3].xy).xyz;
    vec3 dt = vec3(1.0, 1.0, 1.0);

    float md1 = dot(abs(c00 - c22), dt);
    float md2 = dot(abs(c02 - c20), dt);

    float w1 = dot(abs(c22 - c11), dt) * md2;
    float w2 = dot(abs(c02 - c11), dt) * md1;
    float w3 = dot(abs(c00 - c11), dt) * md2;
    float w4 = dot(abs(c20 - c11), dt) * md1;

    float t1 = w1 + w3;
    float t2 = w2 + w4;
    float ww = max(t1, t2) + 0.0001;

    c11 = (w1 * c00 + w2 * c20 + w3 * c22 + w4 * c02 + ww * c11) / (t1 + t2 + ww);

    float lc1 = k / (0.12 * dot(c10 + c12 + c11, dt) + lum_add);
    float lc2 = k / (0.12 * dot(c01 + c21 + c11, dt) + lum_add);

    w1 = clamp(lc1 * dot(abs(c11 - c10), dt) + mx, min_w, max_w);
    w2 = clamp(lc2 * dot(abs(c11 - c21), dt) + mx, min_w, max_w);
    w3 = clamp(lc1 * dot(abs(c11 - c12), dt) + mx, min_w, max_w);
    w4 = clamp(lc2 * dot(abs(c11 - c01), dt) + mx, min_w, max_w);

	gl_FragColor.xyz = w1 * c10 + w2 * c21 + w3 * c12 + w4 * c01 + (1.0 - w1 - w2 - w3 - w4) * c11;
}` + "\x00"

var hqx4VertShader string = `
attribute vec2 VertCoord;
uniform vec2 TextureSize;

void main()
{
	vec2 TexCoord = (VertCoord + 1.0) / 2.0;

	float x = 0.001;
    float y = 0.001;
	
	vec2 dg1 = vec2( x,y);  vec2 dg2 = vec2(-x,y);
	vec2 sd1 = dg1*0.5;     vec2 sd2 = dg2*0.5;
	vec2 ddx = vec2(x,0.0); vec2 ddy = vec2(0.0,y);
	
	gl_Position = vec4(VertCoord, 0.0, 1.0);
	
	gl_TexCoord[0].xy = TexCoord;
	gl_TexCoord[1].xy = gl_TexCoord[0].xy - sd1;
	gl_TexCoord[2].xy = gl_TexCoord[0].xy - sd2;
	gl_TexCoord[3].xy = gl_TexCoord[0].xy + sd1;
	gl_TexCoord[4].xy = gl_TexCoord[0].xy + sd2;
	gl_TexCoord[5].xy = gl_TexCoord[0].xy - dg1;
	gl_TexCoord[6].xy = gl_TexCoord[0].xy + dg1;
	gl_TexCoord[5].zw = gl_TexCoord[0].xy - dg2;
	gl_TexCoord[6].zw = gl_TexCoord[0].xy + dg2;
	gl_TexCoord[1].zw = gl_TexCoord[0].xy - ddy;
	gl_TexCoord[2].zw = gl_TexCoord[0].xy + ddx;
	gl_TexCoord[3].zw = gl_TexCoord[0].xy + ddy;
	gl_TexCoord[4].zw = gl_TexCoord[0].xy - ddx;
}` + "\x00"

var hqx4FragShader string = `
uniform sampler2D Texture;

const float mx = 1.00;      // start smoothing wt.
const float k = -1.10;      // wt. decrease factor
const float max_w = 0.75;   // max filter weigth
const float min_w = 0.03;   // min filter weigth
const float lum_add = 0.33; // effects smoothing

void main()
{
		vec3 c  = texture2D(Texture, gl_TexCoord[0].xy).xyz;
		vec3 i1 = texture2D(Texture, gl_TexCoord[1].xy).xyz;
		vec3 i2 = texture2D(Texture, gl_TexCoord[2].xy).xyz;
		vec3 i3 = texture2D(Texture, gl_TexCoord[3].xy).xyz;
		vec3 i4 = texture2D(Texture, gl_TexCoord[4].xy).xyz;
		vec3 o1 = texture2D(Texture, gl_TexCoord[5].xy).xyz;
		vec3 o3 = texture2D(Texture, gl_TexCoord[6].xy).xyz;
		vec3 o2 = texture2D(Texture, gl_TexCoord[5].zw).xyz;
		vec3 o4 = texture2D(Texture, gl_TexCoord[6].zw).xyz;
		vec3 s1 = texture2D(Texture, gl_TexCoord[1].zw).xyz;
		vec3 s2 = texture2D(Texture, gl_TexCoord[2].zw).xyz;
		vec3 s3 = texture2D(Texture, gl_TexCoord[3].zw).xyz;
		vec3 s4 = texture2D(Texture, gl_TexCoord[4].zw).xyz;
		vec3 dt = vec3(1.0,1.0,1.0);

		float ko1=dot(abs(o1-c),dt);
		float ko2=dot(abs(o2-c),dt);
		float ko3=dot(abs(o3-c),dt);
		float ko4=dot(abs(o4-c),dt);

		float k1=min(dot(abs(i1-i3),dt),max(ko1,ko3));
		float k2=min(dot(abs(i2-i4),dt),max(ko2,ko4));

		float w1 = k2; if(ko3<ko1) w1*=ko3/ko1;
		float w2 = k1; if(ko4<ko2) w2*=ko4/ko2;
		float w3 = k2; if(ko1<ko3) w3*=ko1/ko3;
		float w4 = k1; if(ko2<ko4) w4*=ko2/ko4;

		c=(w1*o1+w2*o2+w3*o3+w4*o4+0.001*c)/(w1+w2+w3+w4+0.001);

		w1 = k*dot(abs(i1-c)+abs(i3-c),dt)/(0.125*dot(i1+i3,dt)+lum_add);
		w2 = k*dot(abs(i2-c)+abs(i4-c),dt)/(0.125*dot(i2+i4,dt)+lum_add);
		w3 = k*dot(abs(s1-c)+abs(s3-c),dt)/(0.125*dot(s1+s3,dt)+lum_add);
		w4 = k*dot(abs(s2-c)+abs(s4-c),dt)/(0.125*dot(s2+s4,dt)+lum_add);

		w1 = clamp(w1+mx,min_w,max_w);
		w2 = clamp(w2+mx,min_w,max_w);
		w3 = clamp(w3+mx,min_w,max_w);
		w4 = clamp(w4+mx,min_w,max_w);

		gl_FragColor.xyz=(w1*(i1+i3)+w2*(i2+i4)+w3*(s1+s3)+w4*(s2+s4)+c)/(2.0*(w1+w2+w3+w4)+1.0);
		gl_FragColor.a = 1.0;
}` + "\x00"

var scanlineVertShader string = `
uniform vec2 TextureSize;
attribute vec2 VertCoord;

void main(void) {
	gl_Position = vec4(VertCoord, 0.0, 1.0);
	gl_TexCoord[0].xy = (VertCoord + 1.0) / 2.0;
}
` + "\x00"

var scanlineFragShader string = `
uniform sampler2D Texture;
uniform vec2 TextureSize;

void main(void) {
    vec4 rgb = texture2D(Texture, gl_TexCoord[0].xy);
    vec4 intens ;
    if (fract(gl_FragCoord.y * (0.5*4.0/3.0)) > 0.5)
        intens = vec4(0);
    else
        intens = smoothstep(0.2,0.8,rgb) + normalize(vec4(rgb.xyz, 1.0));
    float level = (4.0-gl_TexCoord[0].z) * 0.19;
    gl_FragColor = intens * (0.5-level) + rgb * 1.1 ;
}
` + "\x00"
