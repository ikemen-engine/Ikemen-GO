package main

import (
	"github.com/go-gl/gl/v2.1/gl"
	"math"
	"unsafe"
)

var notiling = [4]int32{0, 0, 0, 0}
var mugenShader uintptr
var uniformA, uniformPal, uniformMsk int32
var mugenShaderFc uintptr
var uniformFcA, uniformNeg, uniformGray, uniformAdd, uniformMul int32
var mugenShaderFcS uintptr
var uniformFcSA, uniformColor int32

func RenderInit() {
	vertShader := "void main(void){" +
		"gl_TexCoord[0] = gl_TextureMatrix[0] * gl_MultiTexCoord0;" +
		"gl_Position = ftransform();" +
		"}\x00"
	fragShader := "uniform float a;" +
		"uniform sampler2D tex;" +
		"uniform sampler1D pal;" +
		"uniform int msk;" +
		"void main(void){" +
		"float r = texture2D(tex, gl_TexCoord[0].st).r;" +
		"vec4 c;" +
		"gl_FragColor =" +
		"int(255.0*r) == msk ? vec4(0.0)" +
		": (c = texture1D(pal, r*0.9961), vec4(c.rgb, a));" +
		"}\x00"
	fragShaderFc := "uniform float a;" +
		"uniform sampler2D tex;" +
		"uniform bool neg;" +
		"uniform float gray;" +
		"uniform vec3 add;" +
		"uniform vec3 mul;" +
		"void main(void){" +
		"vec4 c = texture2D(tex, gl_TexCoord[0].st);" +
		"if(neg) c.rgb = vec3(1.0) - c.rgb;" +
		"c.rgb += (vec3((c.r + c.g + c.b) / 3.0) - c.rgb) * gray + add;" +
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
		s, l := gl.Str(src), int32(len(src)-1)
		gl.ShaderSourceARB(shader, 1, &s, &l)
		gl.CompileShaderARB(shader)
		var ok int32
		gl.GetObjectParameterivARB(shader, gl.OBJECT_COMPILE_STATUS_ARB, &ok)
		if ok == 0 {
			chk(errLog(shader))
			panic(Error("コンパイルエラー"))
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
			panic(Error("リンクエラー"))
		}
		return
	}
	vertObj := compile(gl.VERTEX_SHADER, vertShader)
	fragObj := compile(gl.FRAGMENT_SHADER, fragShader)
	mugenShader = link(vertObj, fragObj)
	uniformA = gl.GetUniformLocationARB(mugenShader, gl.Str("a\x00"))
	uniformPal = gl.GetUniformLocationARB(mugenShader, gl.Str("pal\x00"))
	uniformMsk = gl.GetUniformLocationARB(mugenShader, gl.Str("msk\x00"))
	gl.DeleteObjectARB(fragObj)
	fragObj = compile(gl.FRAGMENT_SHADER, fragShaderFc)
	mugenShaderFc = link(vertObj, fragObj)
	uniformFcA = gl.GetUniformLocationARB(mugenShader, gl.Str("a\x00"))
	uniformNeg = gl.GetUniformLocationARB(mugenShaderFc, gl.Str("neg\x00"))
	uniformGray = gl.GetUniformLocationARB(mugenShaderFc, gl.Str("gray\x00"))
	uniformAdd = gl.GetUniformLocationARB(mugenShaderFc, gl.Str("add\x00"))
	uniformMul = gl.GetUniformLocationARB(mugenShaderFc, gl.Str("mul\x00"))
	gl.DeleteObjectARB(fragObj)
	fragObj = compile(gl.FRAGMENT_SHADER, fragShaderFcS)
	mugenShaderFcS = link(vertObj, fragObj)
	uniformFcSA = gl.GetUniformLocationARB(mugenShader, gl.Str("a\x00"))
	uniformColor = gl.GetUniformLocationARB(mugenShaderFcS, gl.Str("color\x00"))
	gl.DeleteObjectARB(fragObj)
	gl.DeleteObjectARB(vertObj)
}
func kaiten(x, y *float32, angle float64, rcx, rcy, vscl float32) {
	temp := (*y - rcy) / vscl
	length := math.Sqrt(float64((*x-rcx)*(*x-rcx) + temp*temp))
	if *x-rcx == 0 {
		if *y-rcy > 0 {
			angle += math.Pi / 2
		} else {
			angle -= math.Pi / 2
		}
		*x = rcx + float32(length*math.Cos(angle))
		*y = rcy + float32(length*math.Sin(angle))*vscl
		return
	}
	kakudo := math.Atan(float64(temp/(*x-rcx))) + angle
	if *x-rcx < 0 {
		kakudo += math.Pi
	}
	*x = rcx + float32(length*math.Cos(kakudo))
	*y = rcy + float32(length*math.Sin(kakudo))*vscl
}
func drawQuads(x1, y1, x2, y2, x3, y3, x4, y4, r, g, b, a, pers float32) {
	gl.Color4f(r, g, b, a)
	gl.Begin(gl.TRIANGLE_STRIP)
	gl.TexCoord2f(0, 1)
	gl.Vertex2f(x1, y1)
	gl.TexCoord2f(0, 0)
	gl.Vertex2f(x4, y4)
	if pers != 1 {
		n := int((1 - (pers * pers)) * AbsF(x1-x2) *
			float32(sys.scrrect[3]>>5) / (AbsF(y1-y4) + float32(sys.scrrect[3]>>5)))
		for i := 1; i < n; i++ {
			gl.TexCoord2f(float32(i)/float32(n), 1)
			gl.Vertex2f(x1+(x2-x1)*float32(i)/float32(n),
				y1+(y2-y1)*float32(i)/float32(n))
			gl.TexCoord2f(float32(i)/float32(n), 0)
			gl.Vertex2f(x4+(x3-x4)*float32(i)/float32(n),
				y4+(y3-y4)*float32(i)/float32(n))
		}
	}
	gl.TexCoord2f(1, 1)
	gl.Vertex2f(x2, y2)
	gl.TexCoord2f(1, 0)
	gl.Vertex2f(x3, y3)
	gl.End()
}
func rmTileHSub(x1, y1, x2, y2, x3, y3, x4, y4, xtw, xbw, xts, xbs float32,
	tl *[4]int32, rcx, r, g, b, a, pers float32) {
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
					drawQuads(x1d, y1, x2d, y2, x3d, y3, x4d, y4, r, g, b, a, pers)
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
			drawQuads(x1, y1, x2, y2, x3, y3, x4, y4, r, g, b, a, pers)
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
func rmTileSub(w, h uint16, x, y float32, tl *[4]int32,
	xts, xbs, ys, vs, rxadd, agl, rcx, rcy, r, g, b, a float32) {
	x1, y1 := x+rxadd*ys*float32(h), rcy+((y-ys*float32(h))-rcy)*vs
	x2, y2 := x1+xbs*float32(w), y1
	x3, y3 := x+xts*float32(w), rcy+(y-rcy)*vs
	x4, y4 := x, y3
	var pers float32
	if AbsF(xts) < AbsF(xbs) {
		pers = AbsF(xts) / AbsF(xbs)
	} else {
		pers = AbsF(xbs) / AbsF(xts)
	}
	if agl != 0 {
		kaiten(&x1, &y1, float64(agl), rcx, rcy, vs)
		kaiten(&x2, &y2, float64(agl), rcx, rcy, vs)
		kaiten(&x3, &y3, float64(agl), rcx, rcy, vs)
		kaiten(&x4, &y4, float64(agl), rcx, rcy, vs)
		drawQuads(x1, y1, x2, y2, x3, y3, x4, y4, r, g, b, a, pers)
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
					rcx, r, g, b, a, pers)
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
					(x3-x4)/float32(w), (x2-x1)/float32(w), tl, rcx, r, g, b, a, pers)
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
	xts, xbs, ys, vs, rxadd, agl float32, trans int32, rcx, rcy float32) {
	gl.MatrixMode(gl.PROJECTION)
	gl.PushMatrix()
	gl.LoadIdentity()
	gl.Ortho(0, float64(sys.scrrect[2]), 0, float64(sys.scrrect[3]), -1, 1)
	gl.MatrixMode(gl.MODELVIEW)
	gl.PushMatrix()
	gl.Translated(0, float64(sys.scrrect[3]), 0)
	switch {
	case trans == -1:
		gl.Uniform1fARB(a, 1)
		gl.BlendFunc(gl.SRC_ALPHA, gl.ONE)
		rmTileSub(size[0], size[1], x, y, tl, xts, xbs, ys, vs, rxadd,
			agl, rcx, rcy, 1, 1, 1, 1)
	case trans == -2:
		gl.Uniform1fARB(a, 1)
		gl.BlendFunc(gl.ZERO, gl.ONE_MINUS_SRC_COLOR)
		rmTileSub(size[0], size[1], x, y, tl, xts, xbs, ys, vs, rxadd,
			agl, rcx, rcy, 1, 1, 1, 1)
	case trans <= 0:
	case trans < 255:
		gl.Uniform1fARB(a, float32(trans)/255)
		gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
		rmTileSub(size[0], size[1], x, y, tl, xts, xbs, ys, vs, rxadd,
			agl, rcx, rcy, 1, 1, 1, float32(trans)/255)
	case trans < 512:
		gl.Uniform1fARB(a, 1)
		gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
		rmTileSub(size[0], size[1], x, y, tl, xts, xbs, ys, vs, rxadd,
			agl, rcx, rcy, 1, 1, 1, 1)
	default:
		src, dst := trans&0xff, trans>>10&0xff
		if dst < 255 {
			gl.Uniform1fARB(a, 1-float32(dst)/255)
			gl.BlendFunc(gl.ZERO, gl.ONE_MINUS_SRC_ALPHA)
			rmTileSub(size[0], size[1], x, y, tl, xts, xbs, ys, vs, rxadd,
				agl, rcx, rcy, 1, 1, 1, 1-float32(trans)/255)
		}
		if src > 0 {
			gl.Uniform1fARB(a, float32(src)/255)
			gl.BlendFunc(gl.SRC_ALPHA, gl.ONE)
			rmTileSub(size[0], size[1], x, y, tl, xts, xbs, ys, vs, rxadd,
				agl, rcx, rcy, 1, 1, 1, float32(trans)/255)
		}
	}
	gl.PopMatrix()
	gl.MatrixMode(gl.PROJECTION)
	gl.PopMatrix()
}
func rmInitSub(size [2]uint16, x, y *float32, tile *[4]int32, xts float32,
	ys, vs, agl *float32, window *[4]int32, rcx float32, rcy *float32) (
	tl [4]int32) {
	if *vs < 0 {
		*vs *= -1
		*ys *= -1
		*agl *= -1
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
func RenderMugenPal(tex Texture, paltex uint32, mask int32, size [2]uint16,
	x, y float32, tile *[4]int32, xts, xbs, ys, vs, rxadd, agl float32,
	trans int32, window *[4]int32, rcx, rcy float32) {
	if tex == 0 || !IsFinite(x+y+xts+xbs+ys+vs+rxadd+agl+rcx+rcy) {
		return
	}
	tl := rmInitSub(size, &x, &y, tile, xts, &ys, &vs, &agl, window, rcx, &rcy)
	gl.UseProgramObjectARB(mugenShader)
	gl.Uniform1iARB(uniformPal, 1)
	gl.Uniform1iARB(uniformMsk, mask)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, uint32(tex))
	rmMainSub(uniformA, size, x, y, &tl, xts, xbs, ys, vs, rxadd, agl,
		trans, rcx, rcy)
	gl.UseProgramObjectARB(0)
	gl.Disable(gl.SCISSOR_TEST)
	gl.Disable(gl.TEXTURE_2D)
	gl.Disable(gl.BLEND)
}
func RenderMugen(tex Texture, pal []uint32, mask int32, size [2]uint16,
	x, y float32, tile *[4]int32, xts, xbs, ys, vs, rxadd, agl float32,
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
	RenderMugenPal(tex, paltex, mask, size, x, y, tile, xts, xbs, ys, vs, rxadd,
		agl, trans, window, rcx, rcy)
	gl.DeleteTextures(1, &paltex)
	gl.Disable(gl.TEXTURE_1D)
}
func RenderMugenFc(tex Texture, size [2]uint16, x, y float32,
	tile *[4]int32, xts, xbs, ys, vs, rxadd, agl float32, trans int32,
	window *[4]int32, rcx, rcy float32, neg bool, color float32,
	padd, pmul *[3]float32) {
	if tex == 0 || !IsFinite(x+y+xts+xbs+ys+vs+rxadd+agl+rcx+rcy) {
		return
	}
	tl := rmInitSub(size, &x, &y, tile, xts, &ys, &vs, &agl, window, rcx, &rcy)
	gl.UseProgramObjectARB(mugenShaderFc)
	ineg := int32(0)
	if neg {
		ineg = 1
	}
	gl.Uniform1iARB(uniformNeg, ineg)
	gl.Uniform1fARB(uniformGray, 1-color)
	gl.Uniform3fARB(uniformAdd, (*padd)[0], (*padd)[1], (*padd)[2])
	gl.Uniform3fARB(uniformMul, (*pmul)[0], (*pmul)[1], (*pmul)[2])
	gl.BindTexture(gl.TEXTURE_2D, uint32(tex))
	rmMainSub(uniformFcA, size, x, y, &tl, xts, xbs, ys, vs, rxadd, agl,
		trans, rcx, rcy)
	gl.UseProgramObjectARB(0)
	gl.Disable(gl.SCISSOR_TEST)
	gl.Disable(gl.TEXTURE_2D)
	gl.Disable(gl.BLEND)
}
func RenderMugenFcS(tex Texture, size [2]uint16, x, y float32,
	tile *[4]int32, xts, xbs, ys, vs, rxadd, agl float32, trans int32,
	window *[4]int32, rcx, rcy float32, color uint32) {
	if tex == 0 || !IsFinite(x+y+xts+xbs+ys+vs+rxadd+agl+rcx+rcy) {
		return
	}
	tl := rmInitSub(size, &x, &y, tile, xts, &ys, &vs, &agl, window, rcx, &rcy)
	gl.UseProgramObjectARB(mugenShaderFcS)
	gl.Uniform3fARB(
		uniformColor, float32(color>>16&0xff)/255, float32(color>>8&0xff)/255,
		float32(color&0xff)/255)
	gl.BindTexture(gl.TEXTURE_2D, uint32(tex))
	rmMainSub(uniformFcSA, size, x, y, &tl, xts, xbs, ys, vs, rxadd, agl,
		trans, rcx, rcy)
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
	gl.Ortho(0, float64(sys.scrrect[2]), 0, float64(sys.scrrect[3]), -1, 1)
	gl.MatrixMode(gl.MODELVIEW)
	gl.PushMatrix()
	gl.Translated(0, float64(sys.scrrect[3]), 0)
	if trans == -1 {
		gl.BlendFunc(gl.SRC_ALPHA, gl.ONE)
		fill(1)
	} else if trans == -2 {
		gl.BlendFunc(gl.ZERO, gl.ONE_MINUS_SRC_COLOR)
		fill(1)
	} else if trans <= 0 {
	} else if trans < 255 {
		gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
		fill(float32(trans) / 256)
	} else if trans < 512 {
		gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
		fill(1)
	} else {
		src, dst := trans&0xff, trans>>10&0xff
		if dst < 255 {
			gl.BlendFunc(gl.ZERO, gl.ONE_MINUS_SRC_ALPHA)
			fill(float32(dst) / 255)
		}
		if src > 0 {
			gl.BlendFunc(gl.SRC_ALPHA, gl.ONE)
			fill(float32(src) / 255)
		}
	}
	gl.PopMatrix()
	gl.MatrixMode(gl.PROJECTION)
	gl.PopMatrix()
	gl.Disable(gl.BLEND)
}
