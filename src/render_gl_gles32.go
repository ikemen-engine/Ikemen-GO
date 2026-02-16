// This is almost identical to render_gles.go except it uses a VAO
// for GLES 3.2 which is the main version that runs on modern
// Android (ARM). Work adapted from Leon Kasovan

//go:build android

package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"runtime"
	"strings"
	"sync"
	"unsafe"

	mgl "github.com/go-gl/mathgl/mgl32"
	gl "github.com/leonkasovan/gl/v3.2/gles2"
	"github.com/veandco/go-sdl2/sdl"
	"golang.org/x/mobile/exp/f32"
)

// ------------------------------------------------------------------
// ShaderProgram_GLES32

type ShaderProgram_GLES32 struct {
	program    uint32           // OpenGL handle
	attributes map[string]int32 // Attribute name to location
	uniforms   map[string]int32 // Uniform name to location
	textures   map[string]int   // Sampler name to texture unit
	name       string           // For debugging
}

var shaderCompileMutex sync.Mutex

func (r *Renderer_GLES32) newShaderProgram(vert, frag, geo, name string, crashWhenFail bool) (s *ShaderProgram_GLES32, err error) {
	// LOCK THE THREAD HERE
	shaderCompileMutex.Lock()
	defer shaderCompileMutex.Unlock()

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	Logcat("GLES: [LOCKED] Starting: " + name)
	var vertObj, fragObj, prog uint32

	vertObj, err = r.compileShader(gl.VERTEX_SHADER, vert)
	if err != nil {
		return nil, err
	}
	Logcat("GLES: Vertex Obj created: " + name)

	fragObj, err = r.compileShader(gl.FRAGMENT_SHADER, frag)
	if err != nil {
		return nil, err
	}
	Logcat("GLES: Frag Obj created: " + name)

	// IMPORTANT: Geometry shaders are very unstable on GLES 3.2 mobile.
	// For now, let's force skip them to see if we can reach the main menu.
	if false && len(geo) > 0 {
		if geoObj, err := r.compileShader(gl.GEOMETRY_SHADER, geo); chkEX(err, "Shader compliation error on "+name+"\n", crashWhenFail) {
			return nil, err
		} else {
			if prog, err = r.linkProgram(vertObj, fragObj, geoObj); chkEX(err, "Link program error on "+name+"\n", crashWhenFail) {
				return nil, err
			}
		}
	} else {
		Logcat("GLES: Entering linkProgram...")
		prog, err = r.linkProgram(vertObj, fragObj)
		if err != nil {
			return nil, err
		}
	}

	Logcat("GLES: Program linked, creating struct...")
	s = &ShaderProgram_GLES32{program: prog, name: name}
	s.attributes = make(map[string]int32)
	s.uniforms = make(map[string]int32)
	s.textures = make(map[string]int)

	Logcat("GLES: Shader initialization complete for: " + name)
	return s, nil
}

func (r *ShaderProgram_GLES32) glStr(s string) *uint8 {
	return gl.Str(s + "\x00")
}

func (s *ShaderProgram_GLES32) RegisterAttributes(names ...string) {
	for _, name := range names {
		cstr := gl.Str(name + "\x00")
		loc := gl.GetAttribLocation(s.program, cstr)
		s.attributes[name] = loc
		Logcat(fmt.Sprintf("GLES: Attribute [%s] mapped to %d", name, loc))
	}
}

func (s *ShaderProgram_GLES32) RegisterUniforms(names ...string) {
	for _, name := range names {
		cstr := gl.Str(name + "\x00")
		loc := gl.GetUniformLocation(s.program, cstr)
		s.uniforms[name] = loc
		Logcat(fmt.Sprintf("GLES: Uniform [%s] mapped to %d", name, loc))
	}
}

func (s *ShaderProgram_GLES32) RegisterTextures(names ...string) {
	for _, name := range names {
		cstr := gl.Str(name + "\x00")
		loc := gl.GetUniformLocation(s.program, cstr)
		s.uniforms[name] = loc
		s.textures[name] = len(s.textures)
		Logcat(fmt.Sprintf("GLES: Texture [%s] mapped to %d", name, loc))
	}
}

func (r *Renderer_GLES32) compileShader(shaderType uint32, src string) (uint32, error) {
	shader := gl.CreateShader(shaderType)

	// 1. SMART HEADER INJECTION
	// GLES 3.0+ drivers REQUIRE the version to be the very first line.
	// If your file doesn't have it, we add it here.
	fullSrc := src
	if !strings.HasPrefix(strings.TrimSpace(src), "#version") {
		// Anchor to 300 es for best mobile compatibility
		header := "#version 310 es\n"
		fullSrc = header + src
	}

	// Ensure null-termination for CGO
	fullSrc = fullSrc + "\x00"

	typeName := "VERTEX"
	if shaderType == gl.FRAGMENT_SHADER {
		typeName = "FRAGMENT"
	}

	Logcat(fmt.Sprintf("GLES: Compiling %s Shader...", typeName))
	// Logcat("DEBUG SHADER SRC:\n" + fullSrc) // Keep for emergencies

	// 2. MEMORY PINNING
	csource, free := gl.Strs(fullSrc)
	defer free()

	gl.ShaderSource(shader, 1, csource, nil)
	gl.CompileShader(shader)

	// 3. STATUS CHECK
	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, (*int32)(unsafe.Pointer(&status)))

	if status == 0 {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, (*int32)(unsafe.Pointer(&logLength)))

		if logLength > 0 {
			logBytes := make([]byte, logLength)
			gl.GetShaderInfoLog(shader, logLength, nil, (*uint8)(unsafe.Pointer(&logBytes[0])))
			err := fmt.Errorf("GLES %s Shader Err: %s", typeName, string(logBytes))
			Logcat("GLES Error: " + err.Error())
			return 0, err
		}
		return 0, fmt.Errorf("GLES %s Shader Err: Unknown error", typeName)
	}

	Logcat(fmt.Sprintf("GLES: %s ready.", typeName))
	return shader, nil
}

func (r *Renderer_GLES32) linkProgram(params ...uint32) (program uint32, err error) {
	program = gl.CreateProgram()
	for _, param := range params {
		gl.AttachShader(program, param)
	}
	// if len(params) > 2 {
	// 	// Geometry Shader Params
	// 	gl.ProgramParameteri(program, gl.GEOMETRY_INPUT_TYPE, gl.TRIANGLES)
	// 	gl.ProgramParameteri(program, gl.GEOMETRY_OUTPUT_TYPE, gl.TRIANGLE_STRIP)
	// 	gl.ProgramParameteri(program, gl.GEOMETRY_VERTICES_OUT, 3*6)
	// }
	Logcat("GLES: Linking program...")
	gl.LinkProgram(program)
	// Mark shaders for deletion when the program is deleted
	for _, param := range params {
		gl.DetachShader(program, param)
		gl.DeleteShader(param)
	}

	var ok int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &ok)
	if ok == 0 {
		var size, l int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &size)
		if size > 0 {
			str := make([]byte, size+1)
			gl.GetProgramInfoLog(program, size, &l, &str[0])
			err = fmt.Errorf("Link error: %s", string(str[:l]))
		} else {
			err = fmt.Errorf("Unknown link error")
		}
		Logcat("GLES: " + err.Error())
		gl.DeleteProgram(program)
		return 0, err
	}

	Logcat("GLES: Link Successful!")
	return program, nil
}

// ------------------------------------------------------------------
// Texture_GLES32

type Texture_GLES32 struct {
	width  int32
	height int32
	depth  int32
	filter bool
	handle uint32
}

// Generate a new texture name
func (r *Renderer_GLES32) newTexture(width, height, depth int32, filter bool) (t Texture) {
	var h uint32
	gl.ActiveTexture(gl.TEXTURE0)
	gl.GenTextures(1, &h)
	t = &Texture_GLES32{width, height, depth, filter, h}
	runtime.SetFinalizer(t, func(t *Texture_GLES32) {
		sys.mainThreadTask <- func() {
			gl.DeleteTextures(1, &t.handle)
		}
	})
	return
}

func (r *Renderer_GLES32) newPaletteTexture() Texture {
	return r.newTexture(256, 1, 32, false)
}

func (r *Renderer_GLES32) newModelTexture(width, height, depth int32, filter bool) Texture {
	return r.newTexture(width, height, depth, filter)
}

func (r *Renderer_GLES32) newDataTexture(width, height int32) (t Texture) {
	var h uint32
	gl.ActiveTexture(gl.TEXTURE0)
	gl.GenTextures(1, &h)
	t = &Texture_GLES32{width, height, 32, false, h}
	runtime.SetFinalizer(t, func(t *Texture_GLES32) {
		sys.mainThreadTask <- func() {
			gl.DeleteTextures(1, &t.handle)
		}
	})
	gl.BindTexture(gl.TEXTURE_2D, h)
	//gl.TexImage2D(gl.TEXTURE_2D, 0, 32, t.width, t.height, 0, 36, gl.FLOAT, nil)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	return
}

func (r *Renderer_GLES32) newHDRTexture(width, height int32) (t Texture) {
	var h uint32
	gl.ActiveTexture(gl.TEXTURE0)
	gl.GenTextures(1, &h)
	t = &Texture_GLES32{width, height, 24, false, h}
	runtime.SetFinalizer(t, func(t *Texture_GLES32) {
		sys.mainThreadTask <- func() {
			gl.DeleteTextures(1, &t.handle)
		}
	})
	gl.BindTexture(gl.TEXTURE_2D, h)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.MIRRORED_REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.MIRRORED_REPEAT)
	return
}

func (r *Renderer_GLES32) newCubeMapTexture(widthHeight int32, mipmap bool, lowestMipLevel int32) (t Texture) {
	var h uint32
	gl.ActiveTexture(gl.TEXTURE0)
	gl.GenTextures(1, &h)
	t = &Texture_GLES32{widthHeight, widthHeight, 24, false, h}
	runtime.SetFinalizer(t, func(t *Texture_GLES32) {
		sys.mainThreadTask <- func() {
			gl.DeleteTextures(1, &t.handle)
		}
	})
	gl.BindTexture(gl.TEXTURE_CUBE_MAP, h)
	for i := 0; i < 6; i++ {
		gl.TexImage2D(uint32(gl.TEXTURE_CUBE_MAP_POSITIVE_X+i), 0, gl.RGB32F, widthHeight, widthHeight, 0, gl.RGB, gl.FLOAT, nil)
	}
	if mipmap {
		gl.TexParameteri(gl.TEXTURE_CUBE_MAP, gl.TEXTURE_MIN_FILTER, gl.LINEAR_MIPMAP_LINEAR)
		gl.GenerateMipmap(gl.TEXTURE_CUBE_MAP)
	} else {
		gl.TexParameteri(gl.TEXTURE_CUBE_MAP, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
	}

	gl.TexParameteri(gl.TEXTURE_CUBE_MAP, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
	gl.TexParameteri(gl.TEXTURE_CUBE_MAP, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_CUBE_MAP, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
	return
}

// Bind a texture and upload texel data to it
func (t *Texture_GLES32) SetData(data []byte) {
	var interp int32 = gl.NEAREST
	if t.filter {
		interp = gl.LINEAR
	}

	format := t.MapInternalFormat(Max(t.depth, 8))

	var uploadFormat uint32 = gl.RGBA
	if t.depth == 8 {
		uploadFormat = gl.RED
	}

	gl.BindTexture(gl.TEXTURE_2D, t.handle)
	gl.PixelStorei(gl.UNPACK_ALIGNMENT, 1)
	gl.PixelStorei(gl.UNPACK_ROW_LENGTH, 0)

	if data != nil {
		gl.TexImage2D(gl.TEXTURE_2D, 0, int32(format), t.width, t.height, 0, uploadFormat, gl.UNSIGNED_BYTE, unsafe.Pointer(&data[0]))
	} else {
		gl.TexImage2D(gl.TEXTURE_2D, 0, int32(format), t.width, t.height, 0, uploadFormat, gl.UNSIGNED_BYTE, nil)
	}

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, interp)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, interp)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
}

func (t *Texture_GLES32) SetSubData(data []byte, x, y, width, height, stride int32) {
	var interp int32 = gl.NEAREST
	if t.filter {
		interp = gl.LINEAR
	}

	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, t.handle)
	gl.PixelStorei(gl.UNPACK_ALIGNMENT, 1)

	format := t.MapInternalFormat(Max(t.depth, 8))
	bytesPerPixel := t.depth / 8
	if bytesPerPixel < 1 {
		bytesPerPixel = 1
	}

	// Doing this should respect both Linux and Android requirements
	if stride > 0 && stride != width*bytesPerPixel {
		// THIS IS THE FIX:
		// Tell GLES the source buffer has 'stride' pixels per row
		gl.PixelStorei(gl.UNPACK_ROW_LENGTH, stride/bytesPerPixel)
	} else {
		gl.PixelStorei(gl.UNPACK_ROW_LENGTH, 0)
	}

	if data != nil {
		gl.TexSubImage2D(gl.TEXTURE_2D, 0, x, y, width, height, uint32(format), gl.UNSIGNED_BYTE, unsafe.Pointer(&data[0]))
	} else {
		gl.TexSubImage2D(gl.TEXTURE_2D, 0, x, y, width, height, uint32(format), gl.UNSIGNED_BYTE, nil)
	}

	if err := gl.GetError(); err != 0 {
		Logcat(fmt.Sprintf("GL ERROR in SetSubDataStride: %v | w:%d h:%d s:%d", err, width, height, stride))
	}

	// We can't trust Android to do this. Best to also assert it in all other texture uploads
	gl.PixelStorei(gl.UNPACK_ROW_LENGTH, 0)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, interp)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, interp)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)

	// gl.Flush()
}

func (t *Texture_GLES32) SetDataG(data []byte, mag, min, ws, wt TextureSamplingParam) {
	format := t.MapInternalFormat(Max(t.depth, 8))

	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, t.handle)
	gl.PixelStorei(gl.UNPACK_ALIGNMENT, 1)
	gl.PixelStorei(gl.UNPACK_ROW_LENGTH, 0)
	gl.TexImage2D(gl.TEXTURE_2D, 0, int32(format), t.width, t.height, 0, format, gl.UNSIGNED_BYTE, unsafe.Pointer(&data[0]))
	gl.GenerateMipmap(gl.TEXTURE_2D)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, int32(mag))
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, int32(min))
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, int32(ws))
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, int32(wt))
}

func (t *Texture_GLES32) SetPixelData(data []float32) {
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, t.handle)
	gl.PixelStorei(gl.UNPACK_ALIGNMENT, 1)
	gl.PixelStorei(gl.UNPACK_ROW_LENGTH, 0)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA32F, t.width, t.height, 0, gl.RGBA, gl.FLOAT, unsafe.Pointer(&data[0]))
}

func (t Texture_GLES32) CopyData(src *Texture) {
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, 0) // Unbind whatever is currently bound
	srcES := (*src).(*Texture_GLES32)
	var fbo uint32
	gl.GenFramebuffers(1, &fbo)
	gl.BindFramebuffer(gl.READ_FRAMEBUFFER, fbo)
	gl.FramebufferTexture2D(gl.READ_FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, srcES.handle, 0)

	gl.BindTexture(gl.TEXTURE_2D, t.handle)
	// Copy the old texture data into the top-left of the new, larger texture
	gl.CopyTexSubImage2D(gl.TEXTURE_2D, 0, 0, 0, 0, 0, srcES.width, srcES.height)

	gl.BindFramebuffer(gl.READ_FRAMEBUFFER, 0)
	gl.DeleteFramebuffers(1, &fbo)
}

/*
// Not called anywhere
func (t *Texture_GLES32) SetRGBPixelData(data []float32) {
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, t.handle)
	gl.PixelStorei(gl.UNPACK_ALIGNMENT, 1)
	gl.PixelStorei(gl.UNPACK_ROW_LENGTH, 0)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGB32F, t.width, t.height, 0, gl.RGB, gl.FLOAT, unsafe.Pointer(&data[0]))
}
*/

// Return whether texture has a valid handle
func (t *Texture_GLES32) IsValid() bool {
	return t.width != 0 && t.height != 0 && t.handle != 0
}

func (t *Texture_GLES32) GetWidth() int32 {
	return t.width
}

func (t *Texture_GLES32) GetHeight() int32 {
	return t.height
}

func (t *Texture_GLES32) MapInternalFormat(i int32) uint32 {
	var InternalFormatLUT = map[int32]uint32{
		8:  gl.RED,
		24: gl.RGB,
		32: gl.RGBA,
	}
	return InternalFormatLUT[i]
}

// ------------------------------------------------------------------
// Renderer_GLES32

type Renderer_GLES32 struct {
	fbo         uint32
	fbo_texture uint32
	// Normal rendering
	rbo_depth uint32
	// MSAA rendering
	fbo_f         uint32
	fbo_f_texture *Texture_GLES32
	// Shadow Map
	fbo_shadow              uint32
	fbo_shadow_cube_texture uint32
	fbo_env                 uint32
	// Postprocessing FBOs
	fbo_pp         []uint32
	fbo_pp_texture []uint32
	// Post-processing shaders
	postVertBuffer   uint32
	postShaderSelect []*ShaderProgram_GLES32
	// Shader and vertex data for primitive rendering
	spriteShader *ShaderProgram_GLES32
	vertexBuffer uint32
	// Shader and index data for 3D model rendering
	shadowMapShader         *ShaderProgram_GLES32
	modelShader             *ShaderProgram_GLES32
	panoramaToCubeMapShader *ShaderProgram_GLES32
	cubemapFilteringShader  *ShaderProgram_GLES32
	modelVertexBuffer       [2]uint32
	modelIndexBuffer        [2]uint32
	vao                     uint32

	enableModel  bool
	enableShadow bool
	GLES32State
}
type GLES32State struct {
	program             uint32
	depthTest           bool
	depthMask           bool
	invertFrontFace     bool
	doubleSided         bool
	blendEnabled        bool
	blendEquation       BlendEquation
	blendSrc            BlendFunc
	blendDst            BlendFunc
	scissorRect         [4]int32
	scissorEnabled      bool
	texCacheTexHandle   []uint32 // Unit to handle. Sized per GPU
	texCacheLastUsed    []uint64 // Timer value when the slot was last used. Sized per GPU
	texCacheTimer       uint64   // Increments on every texture access
	uniformICache       map[uint32]int32
	uniformF1Cache      map[uint32]float32
	uniformF2Cache      map[uint32][2]float32
	uniformF3Cache      map[uint32][3]float32
	uniformF4Cache      map[uint32][4]float32
	useUV               bool
	useNormal           bool
	useTangent          bool
	useVertColor        bool
	useJoint0           bool
	useJoint1           bool
	useOutlineAttribute bool
}

func (r *Renderer_GLES32) GetName() string {
	return "OpenGL ES 3.2"
}

// init 3D model shader
func (r *Renderer_GLES32) InitModelShader() error {
	var err error
	if r.enableShadow {
		r.modelShader, err = r.newShaderProgram(modelVertShader, "#define ENABLE_SHADOW\n"+modelFragShader, "", "Model Shader", false)
	} else {
		r.modelShader, err = r.newShaderProgram(modelVertShader, modelFragShader, "", "Model Shader", false)
	}
	if err != nil {
		return err
	}

	r.modelShader.RegisterAttributes("inVertexId", "position", "uv", "normalIn", "tangentIn", "vertColor", "joints_0", "joints_1", "weights_0", "weights_1", "outlineAttributeIn")

	r.modelShader.RegisterUniforms("model", "view", "projection", "normalMatrix", "unlit", "baseColorFactor", "add", "mult", "useTexture", "useNormalMap", "useMetallicRoughnessMap", "useEmissionMap", "neg", "gray", "hue",
		"enableAlpha", "alphaThreshold", "numJoints", "morphTargetWeight", "morphTargetOffset", "morphTargetTextureDimension", "numTargets", "numVertices",
		"metallicRoughness", "ambientOcclusionStrength", "emission", "environmentIntensity", "mipCount", "meshOutline",
		"cameraPosition", "environmentRotation", "texTransform", "normalMapTransform", "metallicRoughnessMapTransform", "ambientOcclusionMapTransform", "emissionMapTransform",
		"lightMatrices[0]", "lightMatrices[1]", "lightMatrices[2]", "lightMatrices[3]",
		"lights[0].direction", "lights[0].range", "lights[0].color", "lights[0].intensity", "lights[0].position", "lights[0].innerConeCos", "lights[0].outerConeCos", "lights[0].type", "lights[0].shadowBias", "lights[0].shadowMapFar",
		"lights[1].direction", "lights[1].range", "lights[1].color", "lights[1].intensity", "lights[1].position", "lights[1].innerConeCos", "lights[1].outerConeCos", "lights[1].type", "lights[1].shadowBias", "lights[1].shadowMapFar",
		"lights[2].direction", "lights[2].range", "lights[2].color", "lights[2].intensity", "lights[2].position", "lights[2].innerConeCos", "lights[2].outerConeCos", "lights[2].type", "lights[2].shadowBias", "lights[2].shadowMapFar",
		"lights[3].direction", "lights[3].range", "lights[3].color", "lights[3].intensity", "lights[3].position", "lights[3].innerConeCos", "lights[3].outerConeCos", "lights[3].type", "lights[3].shadowBias", "lights[3].shadowMapFar",
	)
	r.modelShader.RegisterTextures("tex", "morphTargetValues", "jointMatrices", "normalMap", "metallicRoughnessMap", "ambientOcclusionMap", "emissionMap", "lambertianEnvSampler", "GGXEnvSampler", "GGXLUT",
		"shadowCubeMap")

	if r.enableShadow {
		r.shadowMapShader, err = r.newShaderProgram(shadowVertShader, shadowFragShader, shadowGeoShader, "Shadow Map Shader", false)
		if err != nil {
			return err
		}

		r.shadowMapShader.RegisterAttributes("inVertexId", "position", "vertColor", "uv", "joints_0", "joints_1", "weights_0", "weights_1")

		r.shadowMapShader.RegisterUniforms("model", "lightMatrices[0]", "lightMatrices[1]", "lightMatrices[2]", "lightMatrices[3]", "lightMatrices[4]", "lightMatrices[5]",
			"lightMatrices[6]", "lightMatrices[7]", "lightMatrices[8]", "lightMatrices[9]", "lightMatrices[10]", "lightMatrices[11]",
			"lightMatrices[12]", "lightMatrices[13]", "lightMatrices[14]", "lightMatrices[15]", "lightMatrices[16]", "lightMatrices[17]",
			"lightMatrices[18]", "lightMatrices[19]", "lightMatrices[20]", "lightMatrices[21]", "lightMatrices[22]", "lightMatrices[23]",
			"lights[0].type", "lights[1].type", "lights[2].type", "lights[3].type", "lights[0].position", "lights[1].position", "lights[2].position", "lights[3].position",
			"lights[0].shadowMapFar", "lights[1].shadowMapFar", "lights[2].shadowMapFar", "lights[3].shadowMapFar", "numJoints", "morphTargetWeight", "morphTargetOffset", "morphTargetTextureDimension",
			"numTargets", "numVertices", "enableAlpha", "alphaThreshold", "baseColorFactor", "useTexture", "texTransform", "layerOffset", "lightIndex")
		r.shadowMapShader.RegisterTextures("morphTargetValues", "jointMatrices", "tex")
	}
	r.panoramaToCubeMapShader, err = r.newShaderProgram(identVertShader, panoramaToCubeMapFragShader, "", "Panorama To Cubemap Shader", false)
	if err != nil {
		return err
	}
	r.panoramaToCubeMapShader.RegisterAttributes("VertCoord")
	r.panoramaToCubeMapShader.RegisterUniforms("currentFace")
	r.panoramaToCubeMapShader.RegisterTextures("panorama")

	r.cubemapFilteringShader, err = r.newShaderProgram(identVertShader, cubemapFilteringFragShader, "", "Cubemap Filtering Shader", false)
	if err != nil {
		return err
	}
	r.cubemapFilteringShader.RegisterAttributes("VertCoord")
	r.cubemapFilteringShader.RegisterUniforms("sampleCount", "distribution", "width", "currentFace", "roughness", "intensityScale", "isLUT")
	r.cubemapFilteringShader.RegisterTextures("cubeMap")
	return nil
}

// Render initialization.
// Creates the default shaders, the framebuffer and enables MSAA.
func (r *Renderer_GLES32) Init() {
	chk(gl.Init(func(name string) unsafe.Pointer {
		return eglGetProcAddress(name)
	}))
	if runtime.GOOS != "android" {
		sys.errLog.Printf("Using OpenGL %v (%v)", gl.GetString(gl.VERSION), gl.GetString(gl.RENDERER))
	} else {
		Logcat(fmt.Sprintf("Using OpenGL %v (%v)", gl.GetString(gl.VERSION), gl.GetString(gl.RENDERER)))
	}

	// Logcat("GLES: Querying Max Samples")
	// var maxSamples int32
	// gl.GetIntegerv(gl.MAX_SAMPLES, &maxSamples)
	// if sys.msaa > maxSamples {
	// 	sys.cfg.SetValueUpdate("Video.MSAA", maxSamples)
	// 	sys.msaa = maxSamples
	// }
	sys.msaa = 0
	Logcat("GLES: Past MSAA check")

	// Data buffers for rendering
	postVertData := f32.Bytes(binary.LittleEndian, -1, -1, 1, -1, -1, 1, 1, 1)

	r.enableModel = sys.cfg.Video.EnableModel
	r.enableShadow = sys.cfg.Video.EnableModelShadow

	Logcat("GLES: About to Gen VAO")
	gl.GenVertexArrays(1, &r.vao)
	gl.BindVertexArray(r.vao)
	Logcat("GLES: VAO Bound")

	gl.GenBuffers(1, &r.postVertBuffer)
	Logcat("GLES: PostVertBuffer Generated")

	gl.BindBuffer(gl.ARRAY_BUFFER, r.postVertBuffer)
	Logcat(fmt.Sprintf("GLES: Data Size: %d", len(postVertData)))

	if len(postVertData) > 0 {
		gl.BufferData(gl.ARRAY_BUFFER, len(postVertData), unsafe.Pointer(&postVertData[0]), gl.STATIC_DRAW)
		Logcat("GLES: PostVertBuffer Data Uploaded")
	} else {
		Logcat("GLES: ERROR - postVertData is empty!")
	}

	gl.GenBuffers(1, &r.vertexBuffer)
	Logcat("GLES: VertexBuffer Generated")
	gl.GenBuffers(1, &r.modelVertexBuffer[0])
	gl.GenBuffers(1, &r.modelVertexBuffer[1])
	Logcat("GLES: ModelVertexBuffers Generated")

	gl.GenBuffers(1, &r.modelIndexBuffer[0])
	gl.GenBuffers(1, &r.modelIndexBuffer[1])
	Logcat("GLES: ModelIndexBuffers Generated")

	// Sprite shader
	r.spriteShader, _ = r.newShaderProgram(vertShader, fragShader, "", "Main Shader", true)
	r.spriteShader.RegisterAttributes("position", "uv")
	r.spriteShader.RegisterUniforms("modelview", "projection", "x1x2x4x3",
		"alpha", "tint", "mask", "neg", "gray", "add", "mult", "isFlat", "isRgba", "isTrapez", "hue")
	r.spriteShader.RegisterTextures("pal", "tex")

	if r.enableModel {
		if err := r.InitModelShader(); err != nil {
			r.enableModel = false
		}
	}

	// Compile postprocessing shaders
	// Because we only have one VAO, the attributes will only be set in EndFrame()

	// Pre-allocate the shader slice to accommodate all external shaders plus the identity shader
	r.postShaderSelect = make([]*ShaderProgram_GLES32, len(sys.cfg.Video.ExternalShaders)+1)

	// External Shaders
	for i := 0; i < len(sys.cfg.Video.ExternalShaders); i++ {
		r.postShaderSelect[i], _ = r.newShaderProgram(string(sys.externalShaders[0][i])+"\x00", string(sys.externalShaders[1][i])+"\x00",
			"", fmt.Sprintf("Postprocess Shader #%v", i), true)
		r.postShaderSelect[i].RegisterAttributes("VertCoord") // "TexCoord" was registered but never used
		r.postShaderSelect[i].RegisterUniforms("Texture_GLES32", "TextureSize", "CurrentTime")
	}

	// Identity shader (no postprocessing). This should be the last one in modern OpenGL
	identShader, _ := r.newShaderProgram(identVertShader, identFragShader, "", "Identity Postprocess", true)
	identShader.RegisterAttributes("VertCoord")
	//identShader.RegisterUniforms("Texture_GLES32", "TextureSize", "CurrentTime") // None of these are used

	// It should be the last one in modern OpenGL
	r.postShaderSelect[len(r.postShaderSelect)-1] = identShader

	gl.ActiveTexture(gl.TEXTURE0)

	// create a texture for r.fbo
	gl.GenTextures(1, &r.fbo_texture)

	gl.BindTexture(gl.TEXTURE_2D, r.fbo_texture)

	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)

	// Don't change this from gl.RGBA.
	// It breaks mixing between subtractive and additive.
	Logcat("GLES: Creating RGBA Textures")
	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		gl.RGBA,
		sys.scrrect[2],
		sys.scrrect[3],
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		nil,
	)

	r.fbo_pp = make([]uint32, 2)
	r.fbo_pp_texture = make([]uint32, 2)

	// Shaders might use negative values, so
	// we specify that we want signed pixels
	// r.fbo_pp_texture
	for i := 0; i < 2; i++ {
		gl.GenTextures(1, &(r.fbo_pp_texture[i]))
		gl.BindTexture(gl.TEXTURE_2D, r.fbo_pp_texture[i])
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
		gl.TexImage2D(
			gl.TEXTURE_2D,
			0,
			gl.RGBA8_SNORM,
			sys.scrrect[2],
			sys.scrrect[3],
			0,
			gl.RGBA,
			gl.UNSIGNED_BYTE,
			nil,
		)
	}

	// done with r.fbo_texture, unbind it
	gl.BindTexture(gl.TEXTURE_2D, 0)

	//r.rbo_depth = gl.CreateRenderbuffer()
	gl.GenRenderbuffers(1, &r.rbo_depth)

	gl.BindRenderbuffer(gl.RENDERBUFFER, r.rbo_depth)
	if sys.msaa > 0 {
		//gl.RenderbufferStorage(gl.RENDERBUFFER, gl.DEPTH_COMPONENT16, int(sys.scrrect[2]), int(sys.scrrect[3]))
		gl.RenderbufferStorageMultisample(gl.RENDERBUFFER, sys.msaa, gl.DEPTH_COMPONENT16, sys.scrrect[2], sys.scrrect[3])
	} else {
		gl.RenderbufferStorage(gl.RENDERBUFFER, gl.DEPTH_COMPONENT16, sys.scrrect[2], sys.scrrect[3])
	}
	gl.BindRenderbuffer(gl.RENDERBUFFER, 0)
	if sys.msaa > 0 {
		r.fbo_f_texture = r.newTexture(sys.scrrect[2], sys.scrrect[3], 32, false).(*Texture_GLES32)
		r.fbo_f_texture.SetData(nil)
	} else {
		//r.rbo_depth = gl.CreateRenderbuffer()
		//gl.BindRenderbuffer(gl.RENDERBUFFER, r.rbo_depth)
		//gl.RenderbufferStorage(gl.RENDERBUFFER, gl.DEPTH_COMPONENT16, int(sys.scrrect[2]), int(sys.scrrect[3]))
		//gl.BindRenderbuffer(gl.RENDERBUFFER, gl.NoRenderbuffer)
	}

	// create an FBO for our r.fbo, which is then for r.fbo_texture
	gl.GenFramebuffers(1, &r.fbo)
	gl.BindFramebuffer(gl.FRAMEBUFFER, r.fbo)

	if sys.msaa > 0 {
		gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D_MULTISAMPLE, r.fbo_texture, 0)
		gl.FramebufferRenderbuffer(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, gl.RENDERBUFFER, r.rbo_depth)
		if status := gl.CheckFramebufferStatus(gl.FRAMEBUFFER); status != gl.FRAMEBUFFER_COMPLETE {
			sys.errLog.Printf("framebuffer create failed: 0x%x", status)
			fmt.Printf("framebuffer create failed: 0x%x \n", status)
		}
		gl.GenFramebuffers(1, &r.fbo_f)
		gl.BindFramebuffer(gl.FRAMEBUFFER, r.fbo_f)
		gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, r.fbo_f_texture.handle, 0)
	} else {
		gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, r.fbo_texture, 0)
		gl.FramebufferRenderbuffer(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, gl.RENDERBUFFER, r.rbo_depth)
	}

	// create our two FBOs for our postprocessing needs
	for i := 0; i < 2; i++ {
		gl.GenFramebuffers(1, &(r.fbo_pp[i]))
		gl.BindFramebuffer(gl.FRAMEBUFFER, r.fbo_pp[i])
		gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, r.fbo_pp_texture[i], 0)
	}

	// create an FBO for our model stuff
	if r.enableModel {
		if r.enableShadow {
			gl.GenFramebuffers(1, &r.fbo_shadow)
			gl.ActiveTexture(gl.TEXTURE0)
			gl.GenTextures(1, &r.fbo_shadow_cube_texture)

			gl.BindTexture(gl.TEXTURE_CUBE_MAP_ARRAY, r.fbo_shadow_cube_texture)
			gl.TexStorage3D(gl.TEXTURE_CUBE_MAP_ARRAY, 1, gl.DEPTH_COMPONENT24, 1024, 1024, 4*6)
			gl.TexParameteri(gl.TEXTURE_CUBE_MAP_ARRAY, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
			gl.TexParameteri(gl.TEXTURE_CUBE_MAP_ARRAY, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
			gl.TexParameteri(gl.TEXTURE_CUBE_MAP_ARRAY, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
			gl.TexParameteri(gl.TEXTURE_CUBE_MAP_ARRAY, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)

			gl.BindFramebuffer(gl.FRAMEBUFFER, r.fbo_shadow)

			// Actually attach the texture to the FBO
			gl.FramebufferTexture(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, r.fbo_shadow_cube_texture, 0)

			// gl.DrawBuffer(gl.NONE)
			bufs := []uint32{gl.NONE}
			gl.DrawBuffers(1, &bufs[0])
			gl.ReadBuffer(gl.NONE)
			if status := gl.CheckFramebufferStatus(gl.FRAMEBUFFER); status != gl.FRAMEBUFFER_COMPLETE {
				Logcat(fmt.Sprintf("[GLES32] framebuffer create failed: 0x%x", status))
				sys.errLog.Printf("[GLES32] framebuffer create failed: 0x%x", status)
			}
		}
		gl.GenFramebuffers(1, &r.fbo_env)
	}

	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)

	// Check hardware texture limit
	// TODO: Maybe clamp the result
	var maxTex int32
	gl.GetIntegerv(gl.MAX_TEXTURE_IMAGE_UNITS, &maxTex)

	//if r.debugMode {
	//	fmt.Printf("[GL Debug] GPU supports up to %d textures\n", maxTex)
	//}

	// Initialize sprite texture cache
	r.texCacheTexHandle = make([]uint32, maxTex)
	r.texCacheLastUsed = make([]uint64, maxTex)

	// Initialize uniform cache
	r.uniformICache = make(map[uint32]int32, 32)
	r.uniformF1Cache = make(map[uint32]float32, 32)
	r.uniformF2Cache = make(map[uint32][2]float32, 32)
	r.uniformF3Cache = make(map[uint32][3]float32, 32)
	r.uniformF4Cache = make(map[uint32][4]float32, 32)
}

func (r *Renderer_GLES32) Close() {
}

func (r *Renderer_GLES32) EnableDebug() {
	// Do nothing yet
}

func (r *Renderer_GLES32) IsModelEnabled() bool {
	return r.enableModel
}

func (r *Renderer_GLES32) IsShadowEnabled() bool {
	return r.enableShadow
}

func (r *Renderer_GLES32) BeginFrame(clearColor bool) {
	gl.BindVertexArray(r.vao)
	gl.BindFramebuffer(gl.FRAMEBUFFER, r.fbo)
	gl.Viewport(0, 0, sys.scrrect[2], sys.scrrect[3])
	if clearColor {
		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
	} else {
		gl.Clear(gl.DEPTH_BUFFER_BIT)
	}
}

func (r *Renderer_GLES32) EndFrame() {
	if len(r.fbo_pp) == 0 {
		return
	}

	x, y, width, height := int32(0), int32(0), int32(sys.scrrect[2]), int32(sys.scrrect[3])
	time := sdl.GetPerformanceCounter() // consistent time across all shaders

	if sys.msaa > 0 {
		gl.BindFramebuffer(gl.DRAW_FRAMEBUFFER, r.fbo_f)
		gl.BindFramebuffer(gl.READ_FRAMEBUFFER, r.fbo)
		gl.BlitFramebuffer(x, y, width, height, x, y, width, height, gl.COLOR_BUFFER_BIT, gl.LINEAR)
	}

	var scaleMode int32 // GL enum
	if sys.cfg.Video.WindowScaleMode {
		scaleMode = gl.LINEAR
	} else {
		scaleMode = gl.NEAREST
	}

	// set the viewport to the unscaled bounds for post-processing
	gl.Viewport(x, y, width, height)
	// clear both of our post-processing FBOs to make sure
	// nothing's there. the output is set later
	for i := 0; i < 2; i++ {
		gl.BindFramebuffer(gl.FRAMEBUFFER, r.fbo_pp[i])
		gl.Clear(gl.COLOR_BUFFER_BIT)
	}
	gl.ActiveTexture(gl.TEXTURE0) // later referred to by Texture_GL

	fbo_texture := r.fbo_texture
	if sys.msaa > 0 {
		fbo_texture = r.fbo_f_texture.handle
	}

	// Reset global state
	r.DisableScissor()
	r.SetBlending(false, 0, 0, 0)
	r.SetDepthTest(false)
	r.SetDepthMask(false)

	for i := 0; i < len(r.postShaderSelect); i++ {
		postShader := r.postShaderSelect[i]

		// tell GL we want to use our shader program
		r.ChangeProgram(postShader.program)

		// tell GL to use our vertex array object
		// this'll be where our quad is stored
		gl.BindVertexArray(r.vao)

		// this is here because it is undefined
		// behavior to write to the same FBO
		if i%2 == 0 {
			// ping! our first post-processing FBO is the output
			gl.BindFramebuffer(gl.FRAMEBUFFER, r.fbo_pp[0])
			if i == 0 {
				// first pass, use fbo_texture
				gl.BindTexture(gl.TEXTURE_2D, fbo_texture)
			} else {
				// not the first pass, use the second post-processing FBO
				gl.BindTexture(gl.TEXTURE_2D, r.fbo_pp_texture[1])
			}
		} else {
			// pong! our second post-processing FBO is the output
			gl.BindFramebuffer(gl.FRAMEBUFFER, r.fbo_pp[1])
			// our first post-processing FBO is the input
			gl.BindTexture(gl.TEXTURE_2D, r.fbo_pp_texture[0])
		}

		if i >= len(r.postShaderSelect)-1 {
			// this is the last shader,
			// so we ask GL to scale it and output it
			// to FB0, the default frame buffer that the user sees
			x, y, width, height := sys.window.GetScaledViewportSize()
			gl.Viewport(x, y, width, height)
			gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
			// clear FB0 just to make sure
			gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
		}

		// set post-processing parameters
		if loc, ok := postShader.uniforms["Texture_GLES32"]; ok && loc >= 0 {
			r.SetUniformISub(loc, 0)
		}
		if loc, ok := postShader.uniforms["TextureSize"]; ok && loc >= 0 {
			r.SetUniformFSub(loc, float32(width), float32(height))
		}
		if loc, ok := postShader.uniforms["CurrentTime"]; ok && loc >= 0 {
			r.SetUniformFSub(loc, float32(time))
		}

		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, scaleMode)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, scaleMode)

		// this actually draws the image to the FBO
		// by constructing a quad (2 tris)
		gl.BindBuffer(gl.ARRAY_BUFFER, r.postVertBuffer)

		// construct the UVs of the quad
		// VertCoord is the primary position attribute
		if loc, ok := postShader.attributes["VertCoord"]; ok && loc >= 0 {
			vLoc := uint32(loc)
			gl.EnableVertexAttribArray(vLoc)
			gl.VertexAttribPointer(vLoc, 2, gl.FLOAT, false, 0, nil)
		}

		// Some external shaders may use a separate TexCoord attribute instead of calculating it from VertCoord
		//if loc, ok := postShader.attributes["TexCoord"]; ok && loc >= 0 {
		//	tLoc := uint32(loc)
		//	gl.EnableVertexAttribArray(tLoc)
		//	gl.VertexAttribPointer(tLoc, 2, gl.FLOAT, false, 0, nil)
		//}

		// construct the quad and draw it
		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, 4)

		// Disable attributes after use
		for _, loc := range postShader.attributes {
			if loc >= 0 {
				gl.DisableVertexAttribArray(uint32(loc))
			}
		}
	}
}

func (r *Renderer_GLES32) Await() {
	gl.Finish()
}

func (r *Renderer_GLES32) MapBlendEquation(i BlendEquation) uint32 {
	var BlendEquationLUT = map[BlendEquation]uint32{
		BlendAdd:             gl.FUNC_ADD,
		BlendReverseSubtract: gl.FUNC_REVERSE_SUBTRACT,
	}
	return BlendEquationLUT[i]
}

func (r *Renderer_GLES32) MapBlendFunction(i BlendFunc) uint32 {
	var BlendFunctionLUT = map[BlendFunc]uint32{
		BlendOne:              gl.ONE,
		BlendZero:             gl.ZERO,
		BlendSrcAlpha:         gl.SRC_ALPHA,
		BlendOneMinusSrcAlpha: gl.ONE_MINUS_SRC_ALPHA,
	}
	return BlendFunctionLUT[i]
}

func (r *Renderer_GLES32) MapPrimitiveMode(i PrimitiveMode) uint32 {
	var PrimitiveModeLUT = map[PrimitiveMode]uint32{
		LINES:          gl.LINES,
		LINE_LOOP:      gl.LINE_LOOP,
		LINE_STRIP:     gl.LINE_STRIP,
		TRIANGLES:      gl.TRIANGLES,
		TRIANGLE_STRIP: gl.TRIANGLE_STRIP,
		TRIANGLE_FAN:   gl.TRIANGLE_FAN,
	}
	return PrimitiveModeLUT[i]
}

func (r *Renderer_GLES32) SetDepthTest(depthTest bool) {
	if depthTest != r.depthTest {
		r.depthTest = depthTest
		if depthTest {
			gl.Enable(gl.DEPTH_TEST)
			gl.DepthFunc(gl.LESS)
		} else {
			gl.Disable(gl.DEPTH_TEST)
		}
	}
}

// Note: This one defaults to enable so we must sync the cache early
func (r *Renderer_GLES32) SetDepthMask(depthMask bool) {
	if depthMask != r.depthMask {
		r.depthMask = depthMask
		gl.DepthMask(depthMask)
	}
}

func (r *Renderer_GLES32) SetFrontFace(invertFrontFace bool) {
	if invertFrontFace != r.invertFrontFace {
		r.invertFrontFace = invertFrontFace
		if invertFrontFace {
			gl.FrontFace(gl.CW)
		} else {
			gl.FrontFace(gl.CCW)
		}
	}
}

func (r *Renderer_GLES32) SetCullFace(doubleSided bool) {
	if doubleSided != r.doubleSided {
		r.doubleSided = doubleSided
		if !doubleSided {
			gl.Enable(gl.CULL_FACE)
			gl.CullFace(gl.BACK)
		} else {
			gl.Disable(gl.CULL_FACE)
		}
	}
}

// This should be called instead of gl.UseProgram()
func (r *Renderer_GLES32) ChangeProgram(prog uint32) {
	// Program already in use
	if r.program == prog {
		return
	}

	// Lazy release of sprite pipeline
	// We can't tell if the next thing we will draw is also a sprite, so this prevents releasing the pipeline after every single sprite
	if r.program == r.spriteShader.program {
		r.ReleasePipeline()
	}

	// Same for TTF fonts
	if r.program == gfxFont.(*FontRenderer_GLES32).shaderProgram.program {
		gfxFont.(*FontRenderer_GLES32).ReleaseFontPipeline()
	}

	// Switch program
	gl.UseProgram(prog)
	r.program = prog

	// Reset sprite texture cache
	for i := range r.texCacheTexHandle {
		r.texCacheTexHandle[i] = 0xFFFFFFFF
		r.texCacheLastUsed[i] = 0
	}
	r.texCacheTimer = 1
}

func (r *Renderer_GLES32) SetBlending(enable bool, eq BlendEquation, src, dst BlendFunc) {
	if enable != r.blendEnabled {
		if enable {
			r.blendEnabled = true
			gl.Enable(gl.BLEND)
		} else {
			r.blendEnabled = false
			gl.Disable(gl.BLEND)
		}
	}
	if enable {
		if eq != r.blendEquation {
			r.blendEquation = eq
			gl.BlendEquation(r.MapBlendEquation(eq))
		}
		if src != r.blendSrc || dst != r.blendDst {
			r.blendSrc = src
			r.blendDst = dst
			gl.BlendFunc(r.MapBlendFunction(src), r.MapBlendFunction(dst))
		}
	}
}

func (r *Renderer_GLES32) SetPipeline() {
	// Do nothing if we were already using the sprite shader
	if r.program == r.spriteShader.program {
		return
	}

	r.ChangeProgram(r.spriteShader.program)

	gl.BindVertexArray(r.vao)
	gl.BindBuffer(gl.ARRAY_BUFFER, r.vertexBuffer)

	locPos := r.spriteShader.attributes["position"]
	gl.EnableVertexAttribArray(uint32(locPos))
	gl.VertexAttribPointerWithOffset(uint32(locPos), 2, gl.FLOAT, false, 16, 0)

	locUV := r.spriteShader.attributes["uv"]
	gl.EnableVertexAttribArray(uint32(locUV))
	gl.VertexAttribPointerWithOffset(uint32(locUV), 2, gl.FLOAT, false, 16, 8)
}

func (r *Renderer_GLES32) ReleasePipeline() {
	loc := r.spriteShader.attributes["position"]
	gl.DisableVertexAttribArray(uint32(loc))
	loc = r.spriteShader.attributes["uv"]
	gl.DisableVertexAttribArray(uint32(loc))
	//gl.Disable(gl.BLEND)
}

func (r *Renderer_GLES32) prepareShadowMapPipeline(bufferIndex uint32) {
	r.ChangeProgram(r.shadowMapShader.program)
	gl.BindFramebuffer(gl.FRAMEBUFFER, r.fbo_shadow)
	gl.Viewport(0, 0, 1024, 1024)
	gl.Enable(gl.TEXTURE_2D)
	gl.Disable(gl.BLEND)
	r.SetDepthTest(true)
	r.SetDepthMask(true)
	gl.BlendEquation(gl.FUNC_ADD)
	gl.BlendFunc(gl.ONE, gl.ZERO)
	if r.invertFrontFace {
		gl.FrontFace(gl.CW)
	} else {
		gl.FrontFace(gl.CCW)
	}
	if !r.doubleSided {
		gl.Enable(gl.CULL_FACE)
		gl.CullFace(gl.BACK)
	} else {
		gl.Disable(gl.CULL_FACE)
	}
	//r.depthTest = true
	//r.depthMask = true
	r.blendEquation = BlendAdd
	r.blendSrc = BlendOne
	r.blendDst = BlendZero

	gl.BindBuffer(gl.ARRAY_BUFFER, r.modelVertexBuffer[bufferIndex])
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, r.modelIndexBuffer[bufferIndex])

	gl.FramebufferTexture(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, r.fbo_shadow_cube_texture, 0)
	gl.Clear(gl.DEPTH_BUFFER_BIT)
}

func (r *Renderer_GLES32) setShadowMapPipeline(doubleSided, invertFrontFace, useUV, useNormal, useTangent, useVertColor, useJoint0, useJoint1 bool, numVertices, vertAttrOffset uint32) {
	r.SetFrontFace(invertFrontFace)
	r.SetCullFace(doubleSided)

	loc := r.shadowMapShader.attributes["inVertexId"]
	gl.EnableVertexAttribArray(uint32(loc))
	gl.VertexAttribPointerWithOffset(uint32(loc), 1, gl.INT, false, 0, uintptr(vertAttrOffset))
	offset := vertAttrOffset + 4*numVertices

	loc = r.shadowMapShader.attributes["position"]
	gl.EnableVertexAttribArray(uint32(loc))
	gl.VertexAttribPointerWithOffset(uint32(loc), 3, gl.FLOAT, false, 0, uintptr(offset))
	offset += 12 * numVertices

	if useUV {
		r.useUV = true
		loc = r.shadowMapShader.attributes["uv"]
		gl.EnableVertexAttribArray(uint32(loc))
		gl.VertexAttribPointerWithOffset(uint32(loc), 2, gl.FLOAT, false, 0, uintptr(offset))
		offset += 8 * numVertices
	} else if r.useUV {
		r.useUV = false
		loc = r.shadowMapShader.attributes["uv"]
		gl.DisableVertexAttribArray(uint32(loc))
		gl.VertexAttrib2f(uint32(loc), 0, 0)
	}

	if useNormal {
		offset += 12 * numVertices
	}
	if useTangent {
		offset += 16 * numVertices
	}
	if useVertColor {
		loc = r.shadowMapShader.attributes["vertColor"]
		gl.EnableVertexAttribArray(uint32(loc))
		gl.VertexAttribPointerWithOffset(uint32(loc), 4, gl.FLOAT, false, 0, uintptr(offset))
		offset += 16 * numVertices
	} else {
		loc = r.shadowMapShader.attributes["vertColor"]
		gl.DisableVertexAttribArray(uint32(loc))
		gl.VertexAttrib4f(uint32(loc), 1, 1, 1, 1)
	}
	if useJoint0 {
		r.useJoint0 = true
		loc = r.shadowMapShader.attributes["joints_0"]
		gl.EnableVertexAttribArray(uint32(loc))
		gl.VertexAttribPointerWithOffset(uint32(loc), 4, gl.FLOAT, false, 0, uintptr(offset))
		offset += 16 * numVertices
		loc = r.shadowMapShader.attributes["weights_0"]
		gl.EnableVertexAttribArray(uint32(loc))
		gl.VertexAttribPointerWithOffset(uint32(loc), 4, gl.FLOAT, false, 0, uintptr(offset))
		offset += 16 * numVertices
		if useJoint1 {
			r.useJoint1 = true
			loc = r.shadowMapShader.attributes["joints_1"]
			gl.EnableVertexAttribArray(uint32(loc))
			gl.VertexAttribPointerWithOffset(uint32(loc), 4, gl.FLOAT, false, 0, uintptr(offset))
			offset += 16 * numVertices
			loc = r.shadowMapShader.attributes["weights_1"]
			gl.EnableVertexAttribArray(uint32(loc))
			gl.VertexAttribPointerWithOffset(uint32(loc), 4, gl.FLOAT, false, 0, uintptr(offset))
			offset += 16 * numVertices
		} else if r.useJoint1 {
			r.useJoint1 = false
			loc = r.shadowMapShader.attributes["joints_1"]
			gl.DisableVertexAttribArray(uint32(loc))
			gl.VertexAttrib4f(uint32(loc), 0, 0, 0, 0)
			loc = r.shadowMapShader.attributes["weights_1"]
			gl.DisableVertexAttribArray(uint32(loc))
			gl.VertexAttrib4f(uint32(loc), 0, 0, 0, 0)
		}
	} else if r.useJoint0 {
		r.useJoint0 = false
		r.useJoint1 = false
		loc = r.shadowMapShader.attributes["joints_0"]
		gl.DisableVertexAttribArray(uint32(loc))
		gl.VertexAttrib4f(uint32(loc), 0, 0, 0, 0)
		loc = r.shadowMapShader.attributes["weights_0"]
		gl.DisableVertexAttribArray(uint32(loc))
		gl.VertexAttrib4f(uint32(loc), 0, 0, 0, 0)
		loc = r.shadowMapShader.attributes["joints_1"]
		gl.DisableVertexAttribArray(uint32(loc))
		gl.VertexAttrib4f(uint32(loc), 0, 0, 0, 0)
		loc = r.shadowMapShader.attributes["weights_1"]
		gl.DisableVertexAttribArray(uint32(loc))
		gl.VertexAttrib4f(uint32(loc), 0, 0, 0, 0)
	}
}

func (r *Renderer_GLES32) ReleaseShadowPipeline() {
	loc := r.shadowMapShader.attributes["inVertexId"]
	gl.DisableVertexAttribArray(uint32(loc))
	loc = r.shadowMapShader.attributes["position"]
	gl.DisableVertexAttribArray(uint32(loc))
	loc = r.shadowMapShader.attributes["uv"]
	gl.DisableVertexAttribArray(uint32(loc))
	gl.VertexAttrib2f(uint32(loc), 0, 0)
	loc = r.shadowMapShader.attributes["vertColor"]
	gl.DisableVertexAttribArray(uint32(loc))
	gl.VertexAttrib4f(uint32(loc), 1, 1, 1, 1)
	loc = r.shadowMapShader.attributes["joints_0"]
	gl.DisableVertexAttribArray(uint32(loc))
	gl.VertexAttrib4f(uint32(loc), 0, 0, 0, 0)
	loc = r.shadowMapShader.attributes["weights_0"]
	gl.DisableVertexAttribArray(uint32(loc))
	gl.VertexAttrib4f(uint32(loc), 0, 0, 0, 0)
	loc = r.shadowMapShader.attributes["joints_1"]
	gl.DisableVertexAttribArray(uint32(loc))
	gl.VertexAttrib4f(uint32(loc), 0, 0, 0, 0)
	loc = r.shadowMapShader.attributes["weights_1"]
	gl.DisableVertexAttribArray(uint32(loc))
	gl.VertexAttrib4f(uint32(loc), 0, 0, 0, 0)
	//gl.Disable(gl.TEXTURE_2D)
	r.SetDepthMask(true)
	r.SetDepthTest(false)
	gl.Disable(gl.CULL_FACE)
	gl.Disable(gl.BLEND)
	r.useVertColor = false
	r.useJoint0 = false
	r.useJoint1 = false
}

func (r *Renderer_GLES32) prepareModelPipeline(bufferIndex uint32, env *Environment) {
	r.ChangeProgram(r.modelShader.program)
	gl.BindFramebuffer(gl.FRAMEBUFFER, r.fbo)
	gl.Viewport(0, 0, sys.scrrect[2], sys.scrrect[3])
	gl.Clear(gl.DEPTH_BUFFER_BIT)
	gl.Enable(gl.TEXTURE_2D)
	gl.Enable(gl.TEXTURE_CUBE_MAP)
	gl.Enable(gl.BLEND)

	/*
		// These should be redundant now
		if r.depthTest {
			r.SetDepthTest(true)
			gl.DepthFunc(gl.LESS)
		} else {
			r.SetDepthTest(false)
		}
		r.SetDepthMask(r.depthMask)
	*/

	if r.invertFrontFace {
		gl.FrontFace(gl.CW)
	} else {
		gl.FrontFace(gl.CCW)
	}
	if !r.doubleSided {
		gl.Enable(gl.CULL_FACE)
		gl.CullFace(gl.BACK)
	} else {
		gl.Disable(gl.CULL_FACE)
	}
	gl.BlendEquation(r.MapBlendEquation(r.blendEquation))
	gl.BlendFunc(r.MapBlendFunction(r.blendSrc), r.MapBlendFunction(r.blendDst))

	gl.BindBuffer(gl.ARRAY_BUFFER, r.modelVertexBuffer[bufferIndex])
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, r.modelIndexBuffer[bufferIndex])
	if r.enableShadow {
		loc, unit := r.modelShader.uniforms["shadowCubeMap"], r.modelShader.textures["shadowCubeMap"]
		gl.ActiveTexture((uint32(gl.TEXTURE0 + unit)))
		gl.BindTexture(gl.TEXTURE_CUBE_MAP_ARRAY, r.fbo_shadow_cube_texture)
		gl.Uniform1i(loc, int32(unit))
	}
	if env != nil {
		loc, unit := r.modelShader.uniforms["lambertianEnvSampler"], r.modelShader.textures["lambertianEnvSampler"]
		gl.ActiveTexture((uint32(gl.TEXTURE0 + unit)))
		gl.BindTexture(gl.TEXTURE_CUBE_MAP, env.lambertianTexture.tex.(*Texture_GLES32).handle)
		gl.Uniform1i(loc, int32(unit))
		loc, unit = r.modelShader.uniforms["GGXEnvSampler"], r.modelShader.textures["GGXEnvSampler"]
		gl.ActiveTexture((uint32(gl.TEXTURE0 + unit)))
		gl.BindTexture(gl.TEXTURE_CUBE_MAP, env.GGXTexture.tex.(*Texture_GLES32).handle)
		gl.Uniform1i(loc, int32(unit))
		loc, unit = r.modelShader.uniforms["GGXLUT"], r.modelShader.textures["GGXLUT"]
		gl.ActiveTexture((uint32(gl.TEXTURE0 + unit)))
		gl.BindTexture(gl.TEXTURE_2D, env.GGXLUT.tex.(*Texture_GLES32).handle)
		gl.Uniform1i(loc, int32(unit))

		loc = r.modelShader.uniforms["environmentIntensity"]
		gl.Uniform1f(loc, env.environmentIntensity)
		loc = r.modelShader.uniforms["mipCount"]
		gl.Uniform1i(loc, env.mipmapLevels)
		loc = r.modelShader.uniforms["environmentRotation"]
		rotationMatrix := mgl.Rotate3DX(math.Pi).Mul3(mgl.Rotate3DY(0.5 * math.Pi))
		rotationM := rotationMatrix[:]
		gl.UniformMatrix3fv(loc, 1, false, &rotationM[0])

	} else {
		loc, unit := r.modelShader.uniforms["lambertianEnvSampler"], r.modelShader.textures["lambertianEnvSampler"]
		gl.ActiveTexture((uint32(gl.TEXTURE0 + unit)))
		gl.BindTexture(gl.TEXTURE_CUBE_MAP, 0)
		gl.Uniform1i(loc, int32(unit))
		loc, unit = r.modelShader.uniforms["GGXEnvSampler"], r.modelShader.textures["GGXEnvSampler"]
		gl.ActiveTexture((uint32(gl.TEXTURE0 + unit)))
		gl.BindTexture(gl.TEXTURE_CUBE_MAP, 0)
		gl.Uniform1i(loc, int32(unit))
		loc, unit = r.modelShader.uniforms["GGXLUT"], r.modelShader.textures["GGXLUT"]
		gl.ActiveTexture((uint32(gl.TEXTURE0 + unit)))
		gl.BindTexture(gl.TEXTURE_2D, 0)
		gl.Uniform1i(loc, int32(unit))
		loc = r.modelShader.uniforms["environmentIntensity"]
		gl.Uniform1f(loc, 0)
	}
}

func (r *Renderer_GLES32) SetModelPipeline(eq BlendEquation, src, dst BlendFunc, depthTest, depthMask, doubleSided, invertFrontFace,
	useUV, useNormal, useTangent, useVertColor, useJoint0, useJoint1, useOutlineAttribute bool, numVertices, vertAttrOffset uint32) {
	r.SetDepthTest(depthTest)
	r.SetDepthMask(depthMask)
	r.SetFrontFace(invertFrontFace)
	r.SetCullFace(doubleSided)
	r.SetBlending(true, eq, src, dst)

	loc := r.modelShader.attributes["inVertexId"]
	gl.EnableVertexAttribArray(uint32(loc))
	gl.VertexAttribPointerWithOffset(uint32(loc), 1, gl.INT, false, 0, uintptr(vertAttrOffset))
	offset := vertAttrOffset + 4*numVertices

	loc = r.modelShader.attributes["position"]
	gl.EnableVertexAttribArray(uint32(loc))
	gl.VertexAttribPointerWithOffset(uint32(loc), 3, gl.FLOAT, false, 0, uintptr(offset))
	offset += 12 * numVertices

	if useUV {
		r.useUV = true
		loc = r.modelShader.attributes["uv"]
		gl.EnableVertexAttribArray(uint32(loc))
		gl.VertexAttribPointerWithOffset(uint32(loc), 2, gl.FLOAT, false, 0, uintptr(offset))
		offset += 8 * numVertices
	} else if r.useUV {
		r.useUV = false
		loc = r.modelShader.attributes["uv"]
		gl.DisableVertexAttribArray(uint32(loc))
		gl.VertexAttrib2f(uint32(loc), 0, 0)
	}

	if useNormal {
		r.useNormal = true
		loc = r.modelShader.attributes["normalIn"]
		gl.EnableVertexAttribArray(uint32(loc))
		gl.VertexAttribPointerWithOffset(uint32(loc), 3, gl.FLOAT, false, 0, uintptr(offset))
		offset += 12 * numVertices
	} else if r.useNormal {
		r.useNormal = false
		loc = r.modelShader.attributes["normalIn"]
		gl.DisableVertexAttribArray(uint32(loc))
		gl.VertexAttrib3f(uint32(loc), 0, 0, 0)
	}
	if useTangent {
		r.useTangent = true
		loc = r.modelShader.attributes["tangentIn"]
		gl.EnableVertexAttribArray(uint32(loc))
		gl.VertexAttribPointerWithOffset(uint32(loc), 4, gl.FLOAT, false, 0, uintptr(offset))
		offset += 16 * numVertices
	} else if r.useTangent {
		r.useTangent = false
		loc = r.modelShader.attributes["tangentIn"]
		gl.DisableVertexAttribArray(uint32(loc))
		gl.VertexAttrib4f(uint32(loc), 0, 0, 0, 0)
	}
	if useVertColor {
		r.useVertColor = true
		loc = r.modelShader.attributes["vertColor"]
		gl.EnableVertexAttribArray(uint32(loc))
		gl.VertexAttribPointerWithOffset(uint32(loc), 4, gl.FLOAT, false, 0, uintptr(offset))
		offset += 16 * numVertices
	} else if r.useVertColor {
		r.useVertColor = false
		loc = r.modelShader.attributes["vertColor"]
		gl.DisableVertexAttribArray(uint32(loc))
		gl.VertexAttrib4f(uint32(loc), 1, 1, 1, 1)
	}
	if useJoint0 {
		r.useJoint0 = true
		loc = r.modelShader.attributes["joints_0"]
		gl.EnableVertexAttribArray(uint32(loc))
		gl.VertexAttribPointerWithOffset(uint32(loc), 4, gl.FLOAT, false, 0, uintptr(offset))
		offset += 16 * numVertices
		loc = r.modelShader.attributes["weights_0"]
		gl.EnableVertexAttribArray(uint32(loc))
		gl.VertexAttribPointerWithOffset(uint32(loc), 4, gl.FLOAT, false, 0, uintptr(offset))
		offset += 16 * numVertices
		if useJoint1 {
			r.useJoint1 = true
			loc = r.modelShader.attributes["joints_1"]
			gl.EnableVertexAttribArray(uint32(loc))
			gl.VertexAttribPointerWithOffset(uint32(loc), 4, gl.FLOAT, false, 0, uintptr(offset))
			offset += 16 * numVertices
			loc = r.modelShader.attributes["weights_1"]
			gl.EnableVertexAttribArray(uint32(loc))
			gl.VertexAttribPointerWithOffset(uint32(loc), 4, gl.FLOAT, false, 0, uintptr(offset))
			offset += 16 * numVertices
		} else if r.useJoint1 {
			r.useJoint1 = false
			loc = r.modelShader.attributes["joints_1"]
			gl.DisableVertexAttribArray(uint32(loc))
			gl.VertexAttrib4f(uint32(loc), 0, 0, 0, 0)
			loc = r.modelShader.attributes["weights_1"]
			gl.DisableVertexAttribArray(uint32(loc))
			gl.VertexAttrib4f(uint32(loc), 0, 0, 0, 0)
		}
	} else if r.useJoint0 {
		r.useJoint0 = false
		r.useJoint1 = false
		loc = r.modelShader.attributes["joints_0"]
		gl.DisableVertexAttribArray(uint32(loc))
		gl.VertexAttrib4f(uint32(loc), 0, 0, 0, 0)
		loc = r.modelShader.attributes["weights_0"]
		gl.DisableVertexAttribArray(uint32(loc))
		gl.VertexAttrib4f(uint32(loc), 0, 0, 0, 0)
		loc = r.modelShader.attributes["joints_1"]
		gl.DisableVertexAttribArray(uint32(loc))
		gl.VertexAttrib4f(uint32(loc), 0, 0, 0, 0)
		loc = r.modelShader.attributes["weights_1"]
		gl.DisableVertexAttribArray(uint32(loc))
		gl.VertexAttrib4f(uint32(loc), 0, 0, 0, 0)
	}
	if useOutlineAttribute {
		r.useOutlineAttribute = true
		loc = r.modelShader.attributes["outlineAttributeIn"]
		gl.EnableVertexAttribArray(uint32(loc))
		gl.VertexAttribPointerWithOffset(uint32(loc), 4, gl.FLOAT, false, 0, uintptr(offset))
		offset += 16 * numVertices
	} else if r.useOutlineAttribute {
		r.useOutlineAttribute = false
		loc = r.modelShader.attributes["outlineAttributeIn"]
		gl.DisableVertexAttribArray(uint32(loc))
		gl.VertexAttrib4f(uint32(loc), 0, 0, 0, 0)
	}
}

func (r *Renderer_GLES32) SetMeshOulinePipeline(invertFrontFace bool, meshOutline float32) {
	r.SetFrontFace(invertFrontFace)
	r.SetDepthTest(true)
	r.SetDepthMask(true)

	loc := r.modelShader.uniforms["meshOutline"]
	gl.Uniform1f(loc, meshOutline)
}

func (r *Renderer_GLES32) ReleaseModelPipeline() {
	loc := r.modelShader.attributes["inVertexId"]
	gl.DisableVertexAttribArray(uint32(loc))
	loc = r.modelShader.attributes["position"]
	gl.DisableVertexAttribArray(uint32(loc))
	loc = r.modelShader.attributes["uv"]
	gl.DisableVertexAttribArray(uint32(loc))
	gl.VertexAttrib2f(uint32(loc), 0, 0)
	loc = r.modelShader.attributes["normalIn"]
	gl.DisableVertexAttribArray(uint32(loc))
	gl.VertexAttrib3f(uint32(loc), 0, 0, 0)
	loc = r.modelShader.attributes["tangentIn"]
	gl.DisableVertexAttribArray(uint32(loc))
	gl.VertexAttrib4f(uint32(loc), 0, 0, 0, 0)
	loc = r.modelShader.attributes["vertColor"]
	gl.DisableVertexAttribArray(uint32(loc))
	gl.VertexAttrib4f(uint32(loc), 1, 1, 1, 1)
	loc = r.modelShader.attributes["joints_0"]
	gl.DisableVertexAttribArray(uint32(loc))
	gl.VertexAttrib4f(uint32(loc), 0, 0, 0, 0)
	loc = r.modelShader.attributes["weights_0"]
	gl.DisableVertexAttribArray(uint32(loc))
	gl.VertexAttrib4f(uint32(loc), 0, 0, 0, 0)
	loc = r.modelShader.attributes["joints_1"]
	gl.DisableVertexAttribArray(uint32(loc))
	gl.VertexAttrib4f(uint32(loc), 0, 0, 0, 0)
	loc = r.modelShader.attributes["weights_1"]
	gl.DisableVertexAttribArray(uint32(loc))
	gl.VertexAttrib4f(uint32(loc), 0, 0, 0, 0)
	loc = r.modelShader.attributes["outlineAttributeIn"]
	gl.DisableVertexAttribArray(uint32(loc))
	gl.VertexAttrib4f(uint32(loc), 0, 0, 0, 0)
	//gl.Disable(gl.TEXTURE_2D)
	r.SetDepthMask(true)
	r.SetDepthTest(false)
	gl.Disable(gl.CULL_FACE)
	r.useNormal = false
	r.useTangent = false
	r.useVertColor = false
	r.useJoint0 = false
	r.useJoint1 = false
	r.useOutlineAttribute = false
}

func (r *Renderer_GLES32) ReadPixels(data []uint8, width, height int) {
	// we defer the EndFrame(), SwapBuffers(), and BeginFrame() calls that were previously below now to
	// a single spot in order to prevent the blank screenshot bug on single digit FPS
	gl.BindFramebuffer(gl.READ_FRAMEBUFFER, 0)
	gl.ReadPixels(0, 0, int32(width), int32(height), gl.RGBA, gl.UNSIGNED_BYTE, unsafe.Pointer(&data[0]))
}

func (r *Renderer_GLES32) EnableScissor(x, y, width, height int32) {
	// Flip Y to OpenGL convention
	realY := sys.scrrect[3] - (y + height)

	if r.scissorEnabled &&
		r.scissorRect[0] == x && r.scissorRect[1] == realY &&
		r.scissorRect[2] == width && r.scissorRect[3] == height {
		return
	}

	if !r.scissorEnabled {
		gl.Enable(gl.SCISSOR_TEST)
		r.scissorEnabled = true
	}

	gl.Scissor(x, realY, width, height)
	r.scissorRect = [4]int32{x, realY, width, height}
}

func (r *Renderer_GLES32) DisableScissor() {
	if r.scissorEnabled {
		gl.Disable(gl.SCISSOR_TEST)
		r.scissorEnabled = false
		r.scissorRect = [4]int32{0, 0, 0, 0}
	}
}

func (r *Renderer_GLES32) SetUniformISub(loc int32, val int32) {
	if loc < 0 {
		return
	}

	// Cached path for the sprite shader
	if r.program == r.spriteShader.program {
		key := (r.program << 16) | uint32(loc)
		if old, exists := r.uniformICache[key]; exists && old == val {
			return
		}
		r.uniformICache[key] = val
	}

	gl.Uniform1i(loc, val)
}

func (r *Renderer_GLES32) SetUniformFSub(loc int32, values ...float32) {
	if loc < 0 || len(values) == 0 {
		return
	}

	// Cached path for the sprite shader
	if r.program == r.spriteShader.program {
		key := (r.program << 16) | uint32(loc)

		switch len(values) {
		case 1:
			if old, exists := r.uniformF1Cache[key]; exists && old == values[0] {
				return
			}
			r.uniformF1Cache[key] = values[0]
		case 2:
			v2 := [2]float32{values[0], values[1]}
			if old, exists := r.uniformF2Cache[key]; exists && old == v2 {
				return
			}
			r.uniformF2Cache[key] = v2
		case 3:
			v3 := [3]float32{values[0], values[1], values[2]}
			if old, exists := r.uniformF3Cache[key]; exists && old == v3 {
				return
			}
			r.uniformF3Cache[key] = v3
		case 4:
			v4 := [4]float32{values[0], values[1], values[2], values[3]}
			if old, exists := r.uniformF4Cache[key]; exists && old == v4 {
				return
			}
			r.uniformF4Cache[key] = v4
		}
	}

	// Uncached path
	switch len(values) {
	case 1:
		gl.Uniform1f(loc, values[0])
	case 2:
		gl.Uniform2f(loc, values[0], values[1])
	case 3:
		gl.Uniform3f(loc, values[0], values[1], values[2])
	case 4:
		gl.Uniform4f(loc, values[0], values[1], values[2], values[3])
	}
}

func (r *Renderer_GLES32) SetUniformFvSub(loc int32, values []float32) {
	if loc < 0 || len(values) == 0 {
		return
	}

	switch len(values) {
	case 1, 2, 3, 4:
		r.SetUniformFSub(loc, values...)
	case 8:
		gl.Uniform4fv(loc, 2, &values[0])
	default:
		gl.Uniform1fv(loc, int32(len(values)), &values[0])
	}
}

func (r *Renderer_GLES32) SetUniformI(name string, val int) {
	loc := r.spriteShader.uniforms[name]
	r.SetUniformISub(loc, int32(val))
}

func (r *Renderer_GLES32) SetUniformF(name string, values ...float32) {
	loc := r.spriteShader.uniforms[name]
	r.SetUniformFSub(loc, values...)
}

func (r *Renderer_GLES32) SetUniformFv(name string, values []float32) {
	loc := r.spriteShader.uniforms[name]
	r.SetUniformFvSub(loc, values)
}

// Caching matrices is as expensive as direct function calls
func (r *Renderer_GLES32) SetUniformMatrix(name string, value []float32) {
	loc := r.spriteShader.uniforms[name]
	gl.UniformMatrix4fv(loc, 1, false, &value[0])
}

func (r *Renderer_GLES32) SetModelUniformI(name string, val int) {
	loc := r.modelShader.uniforms[name]
	r.SetUniformISub(loc, int32(val))
}

func (r *Renderer_GLES32) SetModelUniformF(name string, values ...float32) {
	loc := r.modelShader.uniforms[name]
	r.SetUniformFSub(loc, values...)
}

func (r *Renderer_GLES32) SetModelUniformFv(name string, values []float32) {
	loc := r.modelShader.uniforms[name]
	r.SetUniformFvSub(loc, values)
}

func (r *Renderer_GLES32) SetModelUniformMatrix(name string, value []float32) {
	loc := r.modelShader.uniforms[name]
	gl.UniformMatrix4fv(loc, 1, false, &value[0])
}

func (r *Renderer_GLES32) SetModelUniformMatrix3(name string, value []float32) {
	loc := r.modelShader.uniforms[name]
	gl.UniformMatrix3fv(loc, 1, false, &value[0])
}

func (r *Renderer_GLES32) SetShadowMapUniformI(name string, val int) {
	loc := r.shadowMapShader.uniforms[name]
	r.SetUniformISub(loc, int32(val))
}

func (r *Renderer_GLES32) SetShadowMapUniformF(name string, values ...float32) {
	loc := r.shadowMapShader.uniforms[name]
	r.SetUniformFSub(loc, values...)
}

func (r *Renderer_GLES32) SetShadowMapUniformFv(name string, values []float32) {
	loc := r.shadowMapShader.uniforms[name]
	r.SetUniformFvSub(loc, values)
}

func (r *Renderer_GLES32) SetShadowMapUniformMatrix(name string, value []float32) {
	loc := r.shadowMapShader.uniforms[name]
	gl.UniformMatrix4fv(loc, 1, false, &value[0])
}

func (r *Renderer_GLES32) SetShadowMapUniformMatrix3(name string, value []float32) {
	loc := r.shadowMapShader.uniforms[name]
	gl.UniformMatrix3fv(loc, 1, false, &value[0])
}

func (r *Renderer_GLES32) SetTextureSub(uMap map[string]int32, tMap map[string]int, name string, tex Texture) {
	t := tex.(*Texture_GLES32)
	loc := uMap[name]

	// Cached path for the sprite shader
	// Note: The cache doesn't care if a texture is "tex" or "pal"
	if r.program == r.spriteShader.program {
		// Increment reference timer
		r.texCacheTimer++

		var oldestUnit int32 = 0
		var minTime uint64 = math.MaxUint64

		// Look for a hit or the oldest slot
		for i := range r.texCacheTexHandle {
			// If we find the texture already bound, that's a hit
			if r.texCacheTexHandle[i] == t.handle {
				r.texCacheLastUsed[i] = r.texCacheTimer
				r.SetUniformISub(loc, int32(i))
				return
			}

			// While searching, track the oldest slot in case we miss
			if r.texCacheLastUsed[i] < minTime {
				minTime = r.texCacheLastUsed[i]
				oldestUnit = int32(i)
			}
		}

		// Cache miss
		gl.ActiveTexture(gl.TEXTURE0 + uint32(oldestUnit))
		gl.BindTexture(gl.TEXTURE_2D, t.handle)

		// Update cache state
		r.texCacheTexHandle[oldestUnit] = t.handle
		r.texCacheLastUsed[oldestUnit] = r.texCacheTimer

		// Update uniform
		r.SetUniformISub(loc, oldestUnit)
		return
	}

	// Uncached path
	fixedUnit := uint32(tMap[name])
	gl.ActiveTexture(gl.TEXTURE0 + fixedUnit)
	gl.BindTexture(gl.TEXTURE_2D, t.handle)
	r.SetUniformISub(loc, int32(fixedUnit))
}

func (r *Renderer_GLES32) SetTexture(name string, tex Texture) {
	r.SetTextureSub(r.spriteShader.uniforms, r.spriteShader.textures, name, tex)
}

func (r *Renderer_GLES32) SetModelTexture(name string, tex Texture) {
	r.SetTextureSub(r.modelShader.uniforms, r.modelShader.textures, name, tex)
}

func (r *Renderer_GLES32) SetShadowMapTexture(name string, tex Texture) {
	r.SetTextureSub(r.shadowMapShader.uniforms, r.shadowMapShader.textures, name, tex)
}

func (r *Renderer_GLES32) SetShadowFrameTexture(i uint32) {
	gl.FramebufferTexture(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, r.fbo_shadow_cube_texture, 0)
}

func (r *Renderer_GLES32) SetShadowFrameCubeTexture(i uint32) {
	gl.FramebufferTexture(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, r.fbo_shadow_cube_texture, 0)
}

func (r *Renderer_GLES32) SetVertexData(values ...float32) {
	data := f32.Bytes(binary.LittleEndian, values...)
	gl.BindBuffer(gl.ARRAY_BUFFER, r.vertexBuffer)
	gl.BufferData(gl.ARRAY_BUFFER, len(data), unsafe.Pointer(&data[0]), gl.STATIC_DRAW)
}

func (r *Renderer_GLES32) SetModelVertexData(bufferIndex uint32, values []byte) {
	gl.BindBuffer(gl.ARRAY_BUFFER, r.modelVertexBuffer[bufferIndex])
	gl.BufferData(gl.ARRAY_BUFFER, len(values), unsafe.Pointer(&values[0]), gl.STATIC_DRAW)
}

func (r *Renderer_GLES32) SetModelIndexData(bufferIndex uint32, values ...uint32) {
	data := new(bytes.Buffer)
	binary.Write(data, binary.LittleEndian, values)

	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, r.modelIndexBuffer[bufferIndex])
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(values)*4, unsafe.Pointer(&data.Bytes()[0]), gl.STATIC_DRAW)
}

func (r *Renderer_GLES32) RenderQuad() {
	gl.DrawArrays(gl.TRIANGLE_STRIP, 0, 4)
}

func (r *Renderer_GLES32) RenderElements(mode PrimitiveMode, count, offset int) {
	gl.DrawElementsWithOffset(r.MapPrimitiveMode(mode), int32(count), gl.UNSIGNED_INT, uintptr(offset))
}

func (r *Renderer_GLES32) RenderShadowMapElements(mode PrimitiveMode, count, offset int) {
	r.RenderElements(mode, count, offset)
}

func (r *Renderer_GLES32) RenderCubeMap(envTex Texture, cubeTex Texture) {
	envTexture := envTex.(*Texture_GLES32)
	cubeTexture := cubeTex.(*Texture_GLES32)
	textureSize := cubeTexture.width

	r.ChangeProgram(r.panoramaToCubeMapShader.program)

	gl.BindFramebuffer(gl.FRAMEBUFFER, r.fbo_env)
	gl.Viewport(0, 0, textureSize, textureSize)

	loc := r.panoramaToCubeMapShader.attributes["VertCoord"]
	gl.EnableVertexAttribArray(uint32(loc))
	gl.VertexAttribPointerWithOffset(uint32(loc), 2, gl.FLOAT, false, 0, 0)
	data := f32.Bytes(binary.LittleEndian, -1, -1, 1, -1, -1, 1, 1, 1)
	gl.BindBuffer(gl.ARRAY_BUFFER, r.vertexBuffer)
	gl.BufferData(gl.ARRAY_BUFFER, len(data), unsafe.Pointer(&data[0]), gl.STATIC_DRAW)
	loc, unit := r.panoramaToCubeMapShader.uniforms["panorama"], r.panoramaToCubeMapShader.textures["panorama"]
	gl.ActiveTexture((uint32(gl.TEXTURE0 + unit)))
	gl.BindTexture(gl.TEXTURE_2D, envTexture.handle)
	gl.Uniform1i(loc, int32(unit))
	for i := 0; i < 6; i++ {
		gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, uint32(gl.TEXTURE_CUBE_MAP_POSITIVE_X+i), cubeTexture.handle, 0)

		gl.Clear(gl.COLOR_BUFFER_BIT)
		loc := r.panoramaToCubeMapShader.uniforms["currentFace"]
		gl.Uniform1i(loc, int32(i))

		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, 4)
	}
	gl.BindFramebuffer(gl.FRAMEBUFFER, r.fbo)
	gl.BindTexture(gl.TEXTURE_CUBE_MAP, cubeTexture.handle)
	gl.GenerateMipmap(gl.TEXTURE_CUBE_MAP)
}

func (r *Renderer_GLES32) RenderFilteredCubeMap(distribution int32, cubeTex Texture, filteredTex Texture, mipmapLevel, sampleCount int32, roughness float32) {
	cubeTexture := cubeTex.(*Texture_GLES32)
	filteredTexture := filteredTex.(*Texture_GLES32)
	textureSize := filteredTexture.width
	currentTextureSize := textureSize >> mipmapLevel

	r.ChangeProgram(r.cubemapFilteringShader.program)

	gl.BindFramebuffer(gl.FRAMEBUFFER, r.fbo_env)
	gl.Viewport(0, 0, currentTextureSize, currentTextureSize)

	data := f32.Bytes(binary.LittleEndian, -1, -1, 1, -1, -1, 1, 1, 1)
	gl.BindBuffer(gl.ARRAY_BUFFER, r.vertexBuffer)
	gl.BufferData(gl.ARRAY_BUFFER, len(data), unsafe.Pointer(&data[0]), gl.STATIC_DRAW)

	loc := r.cubemapFilteringShader.attributes["VertCoord"]
	gl.EnableVertexAttribArray(uint32(loc))
	gl.VertexAttribPointerWithOffset(uint32(loc), 2, gl.FLOAT, false, 0, 0)

	loc, unit := r.cubemapFilteringShader.uniforms["cubeMap"], r.cubemapFilteringShader.textures["cubeMap"]
	gl.ActiveTexture((uint32(gl.TEXTURE0 + unit)))
	gl.BindTexture(gl.TEXTURE_CUBE_MAP, cubeTexture.handle)
	gl.Uniform1i(loc, int32(unit))
	loc = r.cubemapFilteringShader.uniforms["sampleCount"]
	gl.Uniform1i(loc, sampleCount)
	loc = r.cubemapFilteringShader.uniforms["distribution"]
	gl.Uniform1i(loc, distribution)
	loc = r.cubemapFilteringShader.uniforms["width"]
	gl.Uniform1i(loc, textureSize)
	loc = r.cubemapFilteringShader.uniforms["roughness"]
	gl.Uniform1f(loc, roughness)
	loc = r.cubemapFilteringShader.uniforms["intensityScale"]
	gl.Uniform1f(loc, 1)
	loc = r.cubemapFilteringShader.uniforms["isLUT"]
	gl.Uniform1i(loc, 0)
	for i := 0; i < 6; i++ {
		gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, uint32(gl.TEXTURE_CUBE_MAP_POSITIVE_X+i), filteredTexture.handle, mipmapLevel)

		gl.Clear(gl.COLOR_BUFFER_BIT)
		loc := r.cubemapFilteringShader.uniforms["currentFace"]
		gl.Uniform1i(loc, int32(i))

		gl.DrawArrays(gl.TRIANGLE_STRIP, 0, 4)
	}
	gl.BindFramebuffer(gl.FRAMEBUFFER, r.fbo)
}

func (r *Renderer_GLES32) RenderLUT(distribution int32, cubeTex Texture, lutTex Texture, sampleCount int32) {
	cubeTexture := cubeTex.(*Texture_GLES32)
	lutTexture := lutTex.(*Texture_GLES32)
	textureSize := lutTexture.width

	r.ChangeProgram(r.cubemapFilteringShader.program)

	gl.BindFramebuffer(gl.FRAMEBUFFER, r.fbo_env)
	gl.Viewport(0, 0, textureSize, textureSize)

	data := f32.Bytes(binary.LittleEndian, -1, -1, 1, -1, -1, 1, 1, 1)
	gl.BindBuffer(gl.ARRAY_BUFFER, r.vertexBuffer)
	gl.BufferData(gl.ARRAY_BUFFER, len(data), unsafe.Pointer(&data[0]), gl.STATIC_DRAW)

	loc := r.cubemapFilteringShader.attributes["VertCoord"]
	gl.EnableVertexAttribArray(uint32(loc))
	gl.VertexAttribPointerWithOffset(uint32(loc), 2, gl.FLOAT, false, 0, 0)

	loc, unit := r.cubemapFilteringShader.uniforms["cubeMap"], r.cubemapFilteringShader.textures["cubeMap"]
	gl.ActiveTexture((uint32(gl.TEXTURE0 + unit)))
	gl.BindTexture(gl.TEXTURE_CUBE_MAP, cubeTexture.handle)
	gl.Uniform1i(loc, int32(unit))
	loc = r.cubemapFilteringShader.uniforms["sampleCount"]
	gl.Uniform1i(loc, sampleCount)
	loc = r.cubemapFilteringShader.uniforms["distribution"]
	gl.Uniform1i(loc, distribution)
	loc = r.cubemapFilteringShader.uniforms["width"]
	gl.Uniform1i(loc, textureSize)
	loc = r.cubemapFilteringShader.uniforms["roughness"]
	gl.Uniform1f(loc, 0)
	loc = r.cubemapFilteringShader.uniforms["intensityScale"]
	gl.Uniform1f(loc, 1)
	loc = r.cubemapFilteringShader.uniforms["currentFace"]
	gl.Uniform1i(loc, 0)
	loc = r.cubemapFilteringShader.uniforms["isLUT"]
	gl.Uniform1i(loc, 1)

	gl.BindTexture(gl.TEXTURE_2D, lutTexture.handle)
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA32F, lutTexture.width, lutTexture.height, 0, gl.RGBA, gl.FLOAT, nil)

	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, lutTexture.handle, 0)
	gl.Clear(gl.COLOR_BUFFER_BIT)
	gl.DrawArrays(gl.TRIANGLE_STRIP, 0, 4)
	gl.BindFramebuffer(gl.FRAMEBUFFER, r.fbo)
}

func (r *Renderer_GLES32) PerspectiveProjectionMatrix(angle, aspect, near, far float32) mgl.Mat4 {
	return mgl.Perspective(angle, aspect, near, far)
}

func (r *Renderer_GLES32) OrthographicProjectionMatrix(left, right, bottom, top, near, far float32) mgl.Mat4 {
	ret := mgl.Ortho(left, right, bottom, top, near, far)
	return ret
}

func (r *Renderer_GLES32) NewWorkerThread() bool {
	return false
}

func (r *Renderer_GLES32) SetVSync(interval int) {
	sdl.GLSetSwapInterval(interval)
}
