package main

import (
	"runtime"
	"strings"

	gl "github.com/fyne-io/gl-js"
)

// ------------------------------------------------------------------
// ShaderProgram

type ShaderProgram struct {
	// Program
	program gl.Program
	// Attribute locations (sprite shaders)
	aPos gl.Attrib
	aUv  gl.Attrib
	// Attribute locations (postprocess shaders)
	aVert gl.Attrib
	// Common uniforms used by most shaders
	uModelView  gl.Uniform
	uProjection gl.Uniform
	uTexture    gl.Uniform
	uAlpha      gl.Uniform
	// Additional uniforms
	u map[string]gl.Uniform
}

func newShaderProgram(vert, frag, id string) (s *ShaderProgram) {
	vertObj := compileShader(gl.VERTEX_SHADER, vert)
	fragObj := compileShader(gl.FRAGMENT_SHADER, frag)
	prog := linkProgram(vertObj, fragObj)

	s = &ShaderProgram{program: prog}
	s.aPos = gl.GetAttribLocation(s.program, "position")
	s.aUv = gl.GetAttribLocation(s.program, "uv")
	s.aVert = gl.GetAttribLocation(s.program, "VertCoord")

	s.uModelView = gl.GetUniformLocation(s.program, "modelview")
	s.uProjection = gl.GetUniformLocation(s.program, "projection")
	s.uTexture = gl.GetUniformLocation(s.program, "tex")
	s.uAlpha = gl.GetUniformLocation(s.program, "alpha")
	s.u = make(map[string]gl.Uniform)
	return
}

func (s *ShaderProgram) RegisterUniforms(names ...string) {
	for _, name := range names {
		s.u[name] = gl.GetUniformLocation(s.program, name)
	}
}

func (s *ShaderProgram) UseProgram() {
	gl.UseProgram(s.program)
}

func compileShader(shaderType gl.Enum, src string) (shader gl.Shader) {
	shader = gl.CreateShader(shaderType)
	if !strings.Contains(src, "gl_TexCoord") {
		src = "#version 100\nprecision highp float;\n" + src
	}
	gl.ShaderSource(shader, src)
	gl.CompileShader(shader)
	ok := gl.GetShaderi(shader, gl.COMPILE_STATUS)
	if ok == 0 {
		log := gl.GetShaderInfoLog(shader)
		gl.DeleteShader(shader)
		panic(Error("Shader compile error: " + log))
	}
	return
}

func linkProgram(v, f gl.Shader) (program gl.Program) {
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

// ------------------------------------------------------------------
// Texture

type Texture struct {
	handle gl.Texture
}

// Generate a new texture name
func newTexture() (t *Texture) {
	t = &Texture{gl.CreateTexture()}
	runtime.SetFinalizer(t, func (t *Texture) {
		sys.mainThreadTask <- func() {
			gl.DeleteTexture(t.handle)
		}
	})
	return
}

// Bind a texture and upload texel data to it
func (t *Texture) SetData(width, height, depth int32, filter bool, data []byte) {
	var interp int = gl.NEAREST
	if filter {
		interp = gl.LINEAR
	}

	var format gl.Enum = gl.LUMINANCE
	if depth == 24 {
		format = gl.RGB
	} else if depth == 32 {
		format = gl.RGBA
	}

	gl.BindTexture(gl.TEXTURE_2D, t.handle)
	gl.PixelStorei(gl.UNPACK_ALIGNMENT, 1)
	gl.TexImage2D(gl.TEXTURE_2D, 0, int(width), int(height), format, gl.UNSIGNED_BYTE, data)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, interp)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, interp)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
}
