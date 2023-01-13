//go:build kinc

package main

import (
	"runtime"
	"unsafe"
)

/*
#include <stdlib.h> // malloc()
#include <string.h> // memcpy()

#include <kinc/graphics4/graphics.h>
#include <kinc/graphics4/indexbuffer.h>
#include <kinc/graphics4/pipeline.h>
#include <kinc/graphics4/rendertarget.h>
#include <kinc/graphics4/shader.h>
#include <kinc/graphics4/texture.h>
#include <kinc/graphics4/vertexbuffer.h>

static size_t sizeof_kinc_g4_pipeline_t = sizeof(kinc_g4_pipeline_t);
static size_t sizeof_kinc_g4_texture_t = sizeof(kinc_g4_texture_t);
static size_t sizeof_kinc_g4_vertex_structure_t = sizeof(kinc_g4_vertex_structure_t);
static size_t sizeof_kinc_g4_index_buffer_t = sizeof(kinc_g4_index_buffer_t);
static size_t sizeof_kinc_g4_vertex_buffer_t = sizeof(kinc_g4_vertex_buffer_t);

#include <kinc/log.h>
#include <kinc/io/filereader.h>
#include <stdio.h>
static kinc_g4_shader_t *load_shader(const char *filename, kinc_g4_shader_type_t shader_type) {
	kinc_file_reader_t file;
	kinc_file_reader_open(&file, filename, KINC_FILE_TYPE_ASSET);
	size_t data_size = kinc_file_reader_size(&file);
	uint8_t *data = malloc(data_size);
	kinc_file_reader_read(&file, data, data_size);
	kinc_file_reader_close(&file);
	kinc_g4_shader_t *shader = malloc(sizeof(kinc_g4_shader_t));
	kinc_g4_shader_init(shader, data, data_size, shader_type);
	free(data);
	return shader;
}
*/
import "C"

var BlendEquationLUT = map[BlendEquation]C.kinc_g4_blending_operation_t {
	BlendAdd: C.KINC_G4_BLENDOP_ADD,
	BlendReverseSubtract: C.KINC_G4_BLENDOP_REVERSE_SUBTRACT,
}

var BlendFunctionLUT = map[BlendFunc]C.kinc_g4_blending_factor_t {
	BlendOne: C.KINC_G4_BLEND_ONE,
	BlendZero: C.KINC_G4_BLEND_ZERO,
	BlendSrcAlpha: C.KINC_G4_BLEND_SOURCE_ALPHA,
	BlendOneMinusSrcAlpha: C.KINC_G4_BLEND_INV_SOURCE_ALPHA,
}

// ------------------------------------------------------------------
// Texture

type Texture struct {
	width  int32
	height int32
	depth  int32
	filter bool
	handle *C.kinc_g4_texture_t
}

var TextureFormatLUT = map[int32]C.kinc_image_format_t {
	8: C.KINC_IMAGE_FORMAT_GREY8,
	24: C.KINC_IMAGE_FORMAT_RGB24,
	32: C.KINC_IMAGE_FORMAT_RGBA32,
}

func newTexture(width, height, depth int32, filter bool) (t *Texture) {
	handle := (*C.kinc_g4_texture_t)(C.malloc(C.sizeof_kinc_g4_texture_t))
	t = &Texture{width, height, depth, filter, handle}

	C.kinc_g4_texture_init(t.handle,
		 C.int(width), C.int(height), TextureFormatLUT[depth])

	runtime.SetFinalizer(t, func (t *Texture) {
		sys.mainThreadTask <- func() {
			C.kinc_g4_texture_destroy(t.handle)
			C.free(unsafe.Pointer(t.handle))
		}
	})

	return
}

func (t *Texture) SetData(data []byte) {
	pixels := C.kinc_g4_texture_lock(t.handle)
	stride := C.kinc_g4_texture_stride(t.handle)
	rowBytes := t.width * (t.depth / 8)
	for j := int32(0); j < t.height; j++ {
		src := unsafe.Pointer(&data[j * rowBytes])
		dst := unsafe.Add(unsafe.Pointer(pixels), uintptr(j) * uintptr(stride))
		C.memcpy(dst, src, C.size_t(rowBytes))
	}
	C.kinc_g4_texture_unlock(t.handle)
}

func (t *Texture) IsValid() bool {
	return true
}

// ------------------------------------------------------------------
// Pipeline

type PipelineParams struct {
	eq BlendEquation
	src BlendFunc
	dst BlendFunc
}

type Pipeline struct {
	kp *C.kinc_g4_pipeline_t
	u map[string]C.kinc_g4_constant_location_t
	t map[string]C.kinc_g4_texture_unit_t
}

func (p *Pipeline) RegisterTextures(names ...string) {
	for _, name := range names {
		cname := C.CString(name)
		p.t[name] = C.kinc_g4_pipeline_get_texture_unit(p.kp, cname)
		C.free(unsafe.Pointer(cname))
	}
}

func (p *Pipeline) RegisterUniforms(names ...string) {
	for _, name := range names {
		cname := C.CString(name)
		p.u[name] = C.kinc_g4_pipeline_get_constant_location(p.kp, cname)
		C.free(unsafe.Pointer(cname))
	}
}

// ------------------------------------------------------------------
// Renderer

type Renderer struct {
	layout *C.kinc_g4_vertex_structure_t
	indexBuffer *C.kinc_g4_index_buffer_t
	vertexBuffer *C.kinc_g4_vertex_buffer_t
	// All pipelines use the same shaders
	vertexShader *C.kinc_g4_shader_t
	fragmentShader *C.kinc_g4_shader_t
	// The rendering pipelines
	pipelineCache map[PipelineParams]*Pipeline
	currentPipeline *Pipeline
}

func (r *Renderer) Init() {
	sys.errLog.Printf("Using Kinc library for rendering")

	r.layout = (*C.kinc_g4_vertex_structure_t)(C.malloc(C.sizeof_kinc_g4_vertex_structure_t))
	C.kinc_g4_vertex_structure_init(r.layout)
	C.kinc_g4_vertex_structure_add(r.layout, C.CString("position"), C.KINC_G4_VERTEX_DATA_F32_2X)
	C.kinc_g4_vertex_structure_add(r.layout, C.CString("uv"), C.KINC_G4_VERTEX_DATA_F32_2X)

	r.indexBuffer = (*C.kinc_g4_index_buffer_t)(C.malloc(C.sizeof_kinc_g4_index_buffer_t))
	C.kinc_g4_index_buffer_init(r.indexBuffer, 6, C.KINC_G4_INDEX_BUFFER_FORMAT_16BIT, C.KINC_G4_USAGE_STATIC)
	data := C.kinc_g4_index_buffer_lock(r.indexBuffer)
	indices := unsafe.Slice((*uint16)(unsafe.Pointer(data)), 6)
	indices[0] = 0
	indices[1] = 1
	indices[2] = 2
	indices[3] = 2
	indices[4] = 1
	indices[5] = 3
	C.kinc_g4_index_buffer_unlock(r.indexBuffer)

	r.vertexBuffer = (*C.kinc_g4_vertex_buffer_t)(C.malloc(C.sizeof_kinc_g4_vertex_buffer_t))
	C.kinc_g4_vertex_buffer_init(r.vertexBuffer, 4, r.layout, C.KINC_G4_USAGE_DYNAMIC, 0)

	r.vertexShader = C.load_shader(C.CString("sprite.vert"), C.KINC_G4_SHADER_TYPE_VERTEX)
	r.fragmentShader = C.load_shader(C.CString("sprite.frag"), C.KINC_G4_SHADER_TYPE_FRAGMENT)

	r.pipelineCache = make(map[PipelineParams]*Pipeline)
}

func (r *Renderer) Close() {
}

func (r *Renderer) BeginFrame(clear bool) {
	C.kinc_g4_begin(0)
	if clear {
		C.kinc_g4_clear(C.KINC_G4_CLEAR_COLOR, 0xff000000, 0.0, 0)
	}
}

func (r *Renderer) EndFrame() {
	C.kinc_g4_end(0)
}

func (r *Renderer) SetPipeline (eq BlendEquation, src, dst BlendFunc) {
	params := PipelineParams{ eq, src, dst }
	p, ok := r.pipelineCache[params]
	if !ok {
		// Parameters are unknown, compile a pipeline on-the-fly and cache it
		kp := (*C.kinc_g4_pipeline_t)(C.malloc(C.sizeof_kinc_g4_pipeline_t))
		C.kinc_g4_pipeline_init(kp)
		kp.vertex_shader = r.vertexShader
		kp.fragment_shader = r.fragmentShader
		kp.input_layout[0] = r.layout
		kp.input_layout[1] = nil

		kp.depth_write = false
		kp.depth_mode = C.KINC_G4_COMPARE_ALWAYS

		kp.blend_operation = BlendEquationLUT[eq]
		kp.alpha_blend_operation = C.KINC_G4_BLENDOP_ADD
		kp.blend_source = BlendFunctionLUT[src]
		kp.blend_destination = BlendFunctionLUT[dst]
		kp.alpha_blend_source = C.KINC_G4_BLEND_ZERO
		kp.alpha_blend_destination = C.KINC_G4_BLEND_ONE

		C.kinc_g4_pipeline_compile(kp)

		p = &Pipeline{ kp: kp }
		p.u = make(map[string]C.kinc_g4_constant_location_t)
		p.t = make(map[string]C.kinc_g4_texture_unit_t)
		p.RegisterUniforms("modelview", "projection", "x1x2x4x3",
			"alpha", "tint", "mask", "neg", "gray", "add", "mult", "isFlat", "isRgba", "isTrapez")
		p.RegisterTextures("pal", "tex")

		r.pipelineCache[params] = p
	}
	r.currentPipeline = p
	C.kinc_g4_set_pipeline(p.kp)
}

func (r *Renderer) ReleasePipeline() {
}

func (r *Renderer) ReadPixels(data[]uint8, width, height int) {
	sys.errLog.Printf("STUB: ReadPixels()")
}

func (r *Renderer) Scissor(x, y, width, height int32) {
	C.kinc_g4_scissor(C.int(x), C.int(y), C.int(width), C.int(height))
}

func (r *Renderer) DisableScissor() {
	C.kinc_g4_disable_scissor();
}

func (r *Renderer) SetUniformI(name string, val int) {
	loc := r.currentPipeline.u[name]
	C.kinc_g4_set_int(loc, C.int(val))
}

func (r *Renderer) SetUniformF(name string, val ...float32) {
	loc := r.currentPipeline.u[name]
	switch len(val) {
	case 1: C.kinc_g4_set_float(loc, C.float(val[0]))
	case 2: C.kinc_g4_set_float2(loc, C.float(val[0]), C.float(val[1]))
	case 3: C.kinc_g4_set_float3(loc, C.float(val[0]), C.float(val[1]), C.float(val[2]))
	case 4: C.kinc_g4_set_float4(loc, C.float(val[0]), C.float(val[1]), C.float(val[2]), C.float(val[3]))
	}
}

func (r *Renderer) SetUniformFv(name string, val []float32) {
	loc := r.currentPipeline.u[name]
	C.kinc_g4_set_floats(loc, (*C.float)(unsafe.Pointer(&val[0])), C.int(len(val)))
}

func (r *Renderer) SetUniformMatrix(name string, val []float32) {
	loc := r.currentPipeline.u[name]
	C.kinc_g4_set_matrix4(loc, (*C.kinc_matrix4x4_t)(unsafe.Pointer(&val[0])))
}

func (r *Renderer) SetTexture(name string, t *Texture) {
	unit := r.currentPipeline.t[name]
	C.kinc_g4_set_texture(unit, t.handle)
	C.kinc_g4_set_texture_addressing(unit, C.KINC_G4_TEXTURE_DIRECTION_U, C.KINC_G4_TEXTURE_ADDRESSING_CLAMP);
	C.kinc_g4_set_texture_addressing(unit, C.KINC_G4_TEXTURE_DIRECTION_V, C.KINC_G4_TEXTURE_ADDRESSING_CLAMP);
	if t.filter {
		C.kinc_g4_set_texture_minification_filter(unit, C.KINC_G4_TEXTURE_FILTER_LINEAR);
		C.kinc_g4_set_texture_magnification_filter(unit, C.KINC_G4_TEXTURE_FILTER_LINEAR);
	} else {
		C.kinc_g4_set_texture_minification_filter(unit, C.KINC_G4_TEXTURE_FILTER_POINT);
		C.kinc_g4_set_texture_magnification_filter(unit, C.KINC_G4_TEXTURE_FILTER_POINT);
	}
	C.kinc_g4_set_texture_mipmap_filter(unit, C.KINC_G4_MIPMAP_FILTER_NONE);
}

func (r *Renderer) SetVertexData(values ...float32) {
	data := C.kinc_g4_vertex_buffer_lock_all(r.vertexBuffer)
	for i := 0; i < len(values); i++ {
		dst := unsafe.Add(unsafe.Pointer(data), uintptr(i * 4))
		*(*float32)(dst) = values[i]
	}
	C.kinc_g4_vertex_buffer_unlock_all(r.vertexBuffer)
}

func (r *Renderer) RenderQuad() {
	C.kinc_g4_set_vertex_buffer(r.vertexBuffer)
	C.kinc_g4_set_index_buffer(r.indexBuffer)
	C.kinc_g4_draw_indexed_vertices()
}
