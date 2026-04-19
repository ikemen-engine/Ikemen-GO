package main

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/draw"
	_ "image/jpeg"
	"io"
	"io/fs"
	"math"
	"os"
	"path"
	"strconv"
	"strings"

	_ "github.com/lukegb/dds"
	"github.com/mdouchement/hdr"
	_ "github.com/mdouchement/hdr/codec/rgbe"

	mgl "github.com/go-gl/mathgl/mgl32"
	"github.com/qmuntal/gltf"
	"github.com/qmuntal/gltf/modeler"
	"golang.org/x/mobile/exp/f32"
)

type Model struct {
	scenes              []*Scene
	nodes               []*Node
	meshes              []*Mesh
	textures            []*GLTFTexture
	materials           []*Material
	offset              [3]float32
	rotation            [3]float32
	scale               [3]float32
	pfx                 *PalFX
	animationTimeStamps map[uint32][]float32
	animations          []*GLTFAnimation
	skins               []*Skin
	vertexBuffer        []byte
	elementBuffer       []uint32
	lights              []GLTFLight
	environment         *Environment
	//lightNodes           []int32
	//lightNodesForeground []int32
}

type Scene struct {
	nodes           []uint32
	name            string
	lightNodes      []uint32
	imageBasedLight *uint32
}

type LightType byte

const (
	DirectionalLight = iota
	PointLight
	SpotLight
)

type GLTFLight struct {
	direction       [3]float32
	position        [3]float32
	lightRange      GLTFAnimatableProperty // float32
	color           GLTFAnimatableProperty // [3]float32
	intensity       GLTFAnimatableProperty // float32
	innerConeAngle  GLTFAnimatableProperty // float32
	outerConeAngle  GLTFAnimatableProperty // float32
	lightType       LightType
	shadowMapNear   GLTFAnimatableProperty // float32
	shadowMapFar    GLTFAnimatableProperty // float32
	shadowMapBottom GLTFAnimatableProperty // float32
	shadowMapTop    GLTFAnimatableProperty // float32
	shadowMapLeft   GLTFAnimatableProperty // float32
	shadowMapRight  GLTFAnimatableProperty // float32
	shadowMapBias   GLTFAnimatableProperty // float32
}

type GLTFAnimationType byte

const (
	TRSTranslation = iota
	TRSScale
	TRSRotation
	MorphTargetWeight
	AnimVec2
	AnimVec3
	AnimVec4
	AnimVecElem
	AnimVec
	AnimQuat
	AnimFloat
)

type GLTFAnimationInterpolation byte

const (
	InterpolationLinear = iota
	InterpolationStep
	InterpolationCubicSpline
)

type GLTFAnimation struct {
	id             uint32
	name           string
	enabled        bool
	defaultEnabled bool
	duration       float32
	time           float32
	loopCount      int32
	loop           int32
	channels       []*GLTFAnimationChannel
	samplers       []*GLTFAnimationSampler
}

type GLTFAnimationChannel struct {
	//path         GLTFAnimationType
	target       *GLTFAnimatableProperty
	targetType   GLTFAnimationType
	elemIndex    *uint32
	nodeIndex    *uint32
	samplerIndex uint32
}

type GLTFAnimationSampler struct {
	inputIndex    uint32
	output        []float32
	interpolation GLTFAnimationInterpolation
}

type GLTFTexture struct {
	tex Texture
}

type GLTFAnimatableProperty struct {
	restValue     interface{}
	animatedValue interface{}
	isAnimated    bool
}

func (p *GLTFAnimatableProperty) rest() {
	p.isAnimated = false
}

func (p *GLTFAnimatableProperty) restAt(v interface{}) {
	p.restValue = v
}

func (p *GLTFAnimatableProperty) animate(v interface{}) {
	p.isAnimated = true
	p.animatedValue = v
}

func (p *GLTFAnimatableProperty) getValue() interface{} {
	if p.isAnimated {
		return p.animatedValue
	}
	return p.restValue
}

type AlphaMode byte

const (
	AlphaModeOpaque = iota
	AlphaModeMask
	AlphaModeBlend
)

type Material struct {
	name                          string
	alphaMode                     AlphaMode
	alphaCutoff                   GLTFAnimatableProperty // float32
	textureIndex                  *uint32
	textureOffset                 GLTFAnimatableProperty // [3]float32
	textureRotation               GLTFAnimatableProperty // [4]float32
	textureScale                  GLTFAnimatableProperty // [3]float32
	textureTransform              [9]float32
	normalMapIndex                *uint32
	normalMapOffset               GLTFAnimatableProperty // [3]float32
	normalMapRotation             GLTFAnimatableProperty // [4]float32
	normalMapScale                GLTFAnimatableProperty // [3]float32
	normalMapTransform            [9]float32
	ambientOcclusionMapIndex      *uint32
	ambientOcclusionMapOffset     GLTFAnimatableProperty // [3]float32
	ambientOcclusionMapRotation   GLTFAnimatableProperty // [4]float32
	ambientOcclusionMapScale      GLTFAnimatableProperty // [3]float32
	ambientOcclusionMapTransform  [9]float32
	metallicRoughnessMapIndex     *uint32
	metallicRoughnessMapOffset    GLTFAnimatableProperty // [3]float32
	metallicRoughnessMapRotation  GLTFAnimatableProperty // [4]float32
	metallicRoughnessMapScale     GLTFAnimatableProperty // [3]float32
	metallicRoughnessMapTransform [9]float32
	emissionMapIndex              *uint32
	emissionMapOffset             GLTFAnimatableProperty // [3]float32
	emissionMapRotation           GLTFAnimatableProperty // [4]float32
	emissionMapScale              GLTFAnimatableProperty // [3]float32
	emissionMapTransform          [9]float32
	baseColorFactor               GLTFAnimatableProperty // [4]float32
	doubleSided                   bool
	ambientOcclusion              GLTFAnimatableProperty // float32
	metallic                      GLTFAnimatableProperty // float32
	roughness                     GLTFAnimatableProperty // float32
	emission                      GLTFAnimatableProperty // [3]float32
	unlit                         bool
}

type Trans byte

const (
	TransNone = iota
	TransAdd
	TransReverseSubtract
	TransMul
)

type Node struct {
	id                 uint32
	visible            bool
	meshIndex          *uint32
	transition         GLTFAnimatableProperty // [3]float32
	rotation           GLTFAnimatableProperty // [4]float32
	scale              GLTFAnimatableProperty // [3]float32
	transformChanged   bool
	localTransform     mgl.Mat4
	worldTransform     mgl.Mat4
	normalMatrix       mgl.Mat4
	childrenIndex      []uint32
	trans              Trans
	castShadow         bool
	zWrite             bool
	zTest              bool
	parentIndex        *uint32
	lightIndex         *uint32
	lightDirection     [3]float32
	shadowMapNear      GLTFAnimatableProperty // float32
	shadowMapFar       GLTFAnimatableProperty // float32
	shadowMapBottom    GLTFAnimatableProperty // float32
	shadowMapTop       GLTFAnimatableProperty // float32
	shadowMapLeft      GLTFAnimatableProperty // float32
	shadowMapRight     GLTFAnimatableProperty // float32
	shadowMapBias      GLTFAnimatableProperty // float32
	skin               *uint32
	morphTargetWeights GLTFAnimatableProperty // []float32
	activeMorphTargets []uint32
	layerNumber        *int
	meshOutline        GLTFAnimatableProperty // float32
}

type Skin struct {
	joints              []uint32
	inverseBindMatrices []float32
	texture             *GLTFTexture
}

type Mesh struct {
	name               string
	morphTargetWeights GLTFAnimatableProperty // []float32
	primitives         []*Primitive
}

type PrimitiveMode byte

const (
	POINTS = iota
	LINES
	LINE_LOOP
	LINE_STRIP
	TRIANGLES
	TRIANGLE_STRIP
	TRIANGLE_FAN
)

type MorphTarget struct {
	positionIndex  *uint32
	normalIndex    *uint32
	tangentIndex   *uint32
	uvIndex        *uint32
	colorIndex     *uint32
	targetType     uint32
	offset         uint32
	positionBuffer []float32
	uvBuffer       []float32
	normalBuffer   []float32
	tangentBuffer  []float32
	colorBuffer    []float32
}

type Primitive struct {
	numVertices         uint32
	numIndices          uint32
	vertexBufferOffset  uint32
	elementBufferOffset uint32
	materialIndex       *uint32
	useUV               bool
	useNormal           bool
	useTangent          bool
	useVertexColor      bool
	useJoint0           bool
	useJoint1           bool
	useOutlineAttribute bool
	mode                PrimitiveMode
	morphTargets        []*MorphTarget
	morphTargetTexture  *GLTFTexture
	morphTargetCount    uint32
	morphTargetOffset   [4]float32
	morphTargetWeight   [8]float32
	boundingBox         BoundingBox
}

type BoundingBox struct {
	min [3]float32
	max [3]float32
}

var gltfPrimitiveModeMap = map[gltf.PrimitiveMode]PrimitiveMode{
	gltf.PrimitivePoints:        POINTS,
	gltf.PrimitiveLines:         LINES,
	gltf.PrimitiveLineLoop:      LINE_LOOP,
	gltf.PrimitiveLineStrip:     LINE_STRIP,
	gltf.PrimitiveTriangles:     TRIANGLES,
	gltf.PrimitiveTriangleStrip: TRIANGLE_STRIP,
	gltf.PrimitiveTriangleFan:   TRIANGLE_FAN,
}

type Environment struct {
	hdrTexture            *GLTFTexture
	cubeMapTexture        *GLTFTexture
	lambertianTexture     *GLTFTexture
	GGXTexture            *GLTFTexture
	GGXLUT                *GLTFTexture
	lambertianSampleCount int32
	GGXSampleCount        int32
	GGXLUTSampleCount     int32
	mipmapLevels          int32
	environmentIntensity  float32
}

func loadEnvironment(filepath string) (*Environment, error) {
	env := &Environment{}
	env.lambertianSampleCount = 2048
	env.GGXSampleCount = 1024
	env.GGXLUTSampleCount = 512
	env.environmentIntensity = 1
	file, err := OpenFile(filepath)
	if err != nil {
		return nil, err
	}
	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}
	env.hdrTexture = &GLTFTexture{}
	env.cubeMapTexture = &GLTFTexture{}
	env.lambertianTexture = &GLTFTexture{}
	env.GGXTexture = &GLTFTexture{}
	env.GGXLUT = &GLTFTexture{}
	if hdrImg, ok := img.(hdr.Image); ok {
		bounds := img.Bounds()
		size := bounds.Max.X * bounds.Max.Y * 3
		data := make([]float32, 0, size)
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				color := hdrImg.HDRAt(x, y)
				r, g, b, _ := color.HDRRGBA()
				data = append(data, float32(r), float32(g), float32(b))
			}
		}
		for i, j := 0, len(data)-3; i < j; i, j = i+3, j-3 {
			data[i], data[i+1], data[i+2], data[j], data[j+1], data[j+2] = data[j], data[j+1], data[j+2], data[i], data[i+1], data[i+2]
		}
		sys.mainThreadTask <- func() {
			if !gfx.IsModelEnabled() {
				return
			}
			lowestMipLevel := int32(4)
			env.hdrTexture.tex = gfx.newHDRTexture(int32(bounds.Max.X), int32(bounds.Max.Y))

			env.hdrTexture.tex.SetPixelData(data)
			env.cubeMapTexture.tex = gfx.newCubeMapTexture(256, true, 0)
			env.lambertianTexture.tex = gfx.newCubeMapTexture(256, false, 0)
			env.GGXTexture.tex = gfx.newCubeMapTexture(256, true, lowestMipLevel)
			env.GGXLUT.tex = gfx.newDataTexture(1024, 1024)

			gfx.RenderCubeMap(env.hdrTexture.tex, env.cubeMapTexture.tex)
			gfx.RenderFilteredCubeMap(0, env.cubeMapTexture.tex, env.lambertianTexture.tex, 0, env.lambertianSampleCount, 0)
			env.mipmapLevels = int32(Floor(float32(math.Log2(256)))) + 1 - lowestMipLevel
			for i := int32(0); i < env.mipmapLevels; i++ {
				roughness := float32(i) / float32((env.mipmapLevels - 1))
				gfx.RenderFilteredCubeMap(1, env.cubeMapTexture.tex, env.GGXTexture.tex, int32(i), env.GGXSampleCount, roughness)
			}
			gfx.RenderLUT(1, env.cubeMapTexture.tex, env.GGXLUT.tex, env.GGXLUTSampleCount)
		}
	}
	return env, nil
}

func loadglTFModel(filepath string) (*Model, error) {
	mdl := &Model{offset: [3]float32{0, 0, 0}, rotation: [3]float32{0, 0, 0}, scale: [3]float32{1, 1, 1}}

	isZip, zipPath, pathInZip := IsZipPath(filepath)

	var doc *gltf.Document
	var err error

	if isZip {
		// Handle resources within a ZIP file
		zipReader, errOpen := zip.OpenReader(zipPath)
		if errOpen != nil {
			return nil, fmt.Errorf("Failed to open ZIP archive '%s': %w", zipPath, errOpen)
		}
		defer zipReader.Close()

		var fsys fs.FS = &zipReader.Reader // The zip.Reader implements fs.FS for resource resolution within the archive

		// Open the GLB/glTF file from within the zip
		glbFile, errOpen := fsys.Open(pathInZip)
		if errOpen != nil {
			return nil, fmt.Errorf("Failed to open GLB file '%s' in ZIP '%s': %w", pathInZip, zipPath, errOpen)
		}
		defer glbFile.Close()

		// Create a new decoder with the file stream and the zip archive as the file system
		decoder := gltf.NewDecoderFS(glbFile, fsys)
		doc = new(gltf.Document)
		if err = decoder.Decode(doc); err != nil {
			return nil, fmt.Errorf("failed to decode gltf from zip '%s': %w", filepath, err)
		}
	} else {
		// Handle resources from the standard file system
		f, errOpen := OpenFile(filepath) // Use IKEMEN GO's file abstraction to handle case-insensitivity etc.
		if errOpen != nil {
			return nil, errOpen
		}
		defer f.Close()

		// Use the directory of the file as the file system for resolving relative resources
		fsm := os.DirFS(path.Dir(filepath))
		decoder := gltf.NewDecoderFS(f, fsm)
		doc = new(gltf.Document)
		if err = decoder.Decode(doc); err != nil {
			return nil, fmt.Errorf("failed to decode gltf from file '%s': %w", filepath, err)
		}
	}

	if doc == nil {
		return nil, fmt.Errorf("gltf document is nil after decoding for path: %s", filepath)
	}

	var images = make([]image.Image, 0, len(doc.Images))
	for _, img := range doc.Images {
		var buffer *bytes.Buffer
		if len(img.URI) > 0 {
			if strings.HasPrefix(img.URI, "data:") {
				if strings.HasPrefix(img.URI, "data:image/png;base64,") {
					decodedData, err := base64.StdEncoding.DecodeString(img.URI[22:])
					if err != nil {
						return nil, err
					}
					buffer = bytes.NewBuffer(decodedData)
				} else {
					decodedData, err := base64.StdEncoding.DecodeString(img.URI[23:])
					if err != nil {
						return nil, err
					}
					buffer = bytes.NewBuffer(decodedData)
				}
			} else {
				if err := LoadFile(&img.URI, []string{filepath, "", sys.motif.Def, "data/"}, func(filename string) error {
					// Use OpenFile which respects the virtual file system (zip)
					f, err := OpenFile(filename)
					if err != nil {
						return err
					}
					defer f.Close()
					data, err := io.ReadAll(f)
					if err != nil {
						return err
					}
					buffer = bytes.NewBuffer(data)
					return nil
				}); err != nil {
					return nil, err
				}

			}
		} else {
			source, err := modeler.ReadBufferView(doc, doc.BufferViews[*img.BufferView])
			if err != nil {
				return nil, err
			}
			buffer = bytes.NewBuffer(source)
		}
		res, _, err := image.Decode(buffer)
		if err != nil {
			return nil, err
		}
		images = append(images, res)
	}
	mdl.textures = make([]*GLTFTexture, 0, len(doc.Textures))
	textureMap := map[[2]int32]*GLTFTexture{}
	for _, t := range doc.Textures {
		if t.Sampler != nil {
			if texture, ok := textureMap[[2]int32{int32(*t.Source), int32(*t.Sampler)}]; ok {
				mdl.textures = append(mdl.textures, texture)
			} else {
				texture := &GLTFTexture{}
				s := doc.Samplers[*t.Sampler]
				mag, _ := map[gltf.MagFilter]TextureSamplingParam{
					gltf.MagUndefined: TextureSamplingFilterLinear,
					gltf.MagNearest:   TextureSamplingFilterNearest,
					gltf.MagLinear:    TextureSamplingFilterLinear,
				}[s.MagFilter]
				min, _ := map[gltf.MinFilter]TextureSamplingParam{
					gltf.MinUndefined:            TextureSamplingFilterLinear,
					gltf.MinNearest:              TextureSamplingFilterLinear,
					gltf.MinLinear:               TextureSamplingFilterLinear,
					gltf.MinNearestMipMapNearest: TextureSamplingFilterNearestMipMapNearest,
					gltf.MinLinearMipMapNearest:  TextureSamplingFilterLinearMipMapNearest,
					gltf.MinNearestMipMapLinear:  TextureSamplingFilterNearestMipMapLinear,
					gltf.MinLinearMipMapLinear:   TextureSamplingFilterLinearMipMapLinear,
				}[s.MinFilter]
				wrapS, _ := map[gltf.WrappingMode]TextureSamplingParam{
					gltf.WrapClampToEdge:    TextureSamplingWrapClampToEdge,
					gltf.WrapMirroredRepeat: TextureSamplingWrapMirroredRepeat,
					gltf.WrapRepeat:         TextureSamplingWrapRepeat,
				}[s.WrapS]
				wrapT, _ := map[gltf.WrappingMode]TextureSamplingParam{
					gltf.WrapClampToEdge:    TextureSamplingWrapClampToEdge,
					gltf.WrapMirroredRepeat: TextureSamplingWrapMirroredRepeat,
					gltf.WrapRepeat:         TextureSamplingWrapRepeat,
				}[s.WrapT]

				img := images[*t.Source]
				rgba := image.NewRGBA(img.Bounds())
				draw.Draw(rgba, img.Bounds(), img, img.Bounds().Min, draw.Src)
				sys.mainThreadTask <- func() {
					texture.tex = gfx.newModelTexture(int32(img.Bounds().Max.X), int32(img.Bounds().Max.Y), 32, false)
					texture.tex.SetDataG(rgba.Pix, mag, min, wrapS, wrapT)
				}
				textureMap[[2]int32{int32(*t.Source), int32(*t.Sampler)}] = texture
				mdl.textures = append(mdl.textures, texture)
			}
		} else {
			if texture, ok := textureMap[[2]int32{int32(*t.Source), -1}]; ok {
				mdl.textures = append(mdl.textures, texture)
			} else {
				texture := &GLTFTexture{}
				mag := TextureSamplingFilterNearest
				min := TextureSamplingFilterNearest
				wrapS := TextureSamplingWrapRepeat
				wrapT := TextureSamplingWrapRepeat
				img := images[*t.Source]
				rgba := image.NewRGBA(img.Bounds())
				draw.Draw(rgba, img.Bounds(), img, img.Bounds().Min, draw.Src)
				sys.mainThreadTask <- func() {
					texture.tex = gfx.newModelTexture(int32(img.Bounds().Max.X), int32(img.Bounds().Max.Y), 32, false)
					texture.tex.SetDataG(rgba.Pix, mag, min, wrapS, wrapT)
				}
				textureMap[[2]int32{int32(*t.Source), -1}] = texture
				mdl.textures = append(mdl.textures, texture)
			}
		}

	}
	mdl.materials = make([]*Material, 0, len(doc.Materials))
	for _, m := range doc.Materials {
		material := &Material{
			baseColorFactor:              GLTFAnimatableProperty{isAnimated: false, restValue: [4]float32{1, 1, 1, 1}},
			roughness:                    GLTFAnimatableProperty{isAnimated: false, restValue: float32(1.0)},
			metallic:                     GLTFAnimatableProperty{isAnimated: false, restValue: float32(1.0)},
			ambientOcclusion:             GLTFAnimatableProperty{isAnimated: false, restValue: float32(1.0)},
			emission:                     GLTFAnimatableProperty{isAnimated: false, restValue: [3]float32{0, 0, 0}},
			alphaCutoff:                  GLTFAnimatableProperty{isAnimated: false, restValue: float32(0.0)},
			textureOffset:                GLTFAnimatableProperty{isAnimated: false, restValue: [2]float32{0.0, 0.0}},
			textureRotation:              GLTFAnimatableProperty{isAnimated: false, restValue: float32(0.0)},
			textureScale:                 GLTFAnimatableProperty{isAnimated: false, restValue: [2]float32{1.0, 1.0}},
			normalMapOffset:              GLTFAnimatableProperty{isAnimated: false, restValue: [2]float32{0.0, 0.0}},
			normalMapRotation:            GLTFAnimatableProperty{isAnimated: false, restValue: float32(0.0)},
			normalMapScale:               GLTFAnimatableProperty{isAnimated: false, restValue: [2]float32{1.0, 1.0}},
			ambientOcclusionMapOffset:    GLTFAnimatableProperty{isAnimated: false, restValue: [2]float32{0.0, 0.0}},
			ambientOcclusionMapRotation:  GLTFAnimatableProperty{isAnimated: false, restValue: float32(0.0)},
			ambientOcclusionMapScale:     GLTFAnimatableProperty{isAnimated: false, restValue: [2]float32{1.0, 1.0}},
			metallicRoughnessMapOffset:   GLTFAnimatableProperty{isAnimated: false, restValue: [2]float32{0.0, 0.0}},
			metallicRoughnessMapRotation: GLTFAnimatableProperty{isAnimated: false, restValue: float32(0.0)},
			metallicRoughnessMapScale:    GLTFAnimatableProperty{isAnimated: false, restValue: [2]float32{1.0, 1.0}},
			emissionMapOffset:            GLTFAnimatableProperty{isAnimated: false, restValue: [2]float32{0.0, 0.0}},
			emissionMapRotation:          GLTFAnimatableProperty{isAnimated: false, restValue: float32(0.0)},
			emissionMapScale:             GLTFAnimatableProperty{isAnimated: false, restValue: [2]float32{1.0, 1.0}},
		}
		if m.PBRMetallicRoughness.BaseColorTexture != nil {
			material.textureIndex = new(uint32)
			*material.textureIndex = m.PBRMetallicRoughness.BaseColorTexture.Index
			if m.PBRMetallicRoughness.BaseColorTexture.Extensions != nil {
				if l, ok := m.PBRMetallicRoughness.BaseColorTexture.Extensions["KHR_texture_transform"]; ok {
					var ext interface{}
					err := json.Unmarshal(l.(json.RawMessage), &ext)
					if err != nil {
						return nil, err
					}
					if offset, ok := ext.(map[string]interface{})["offset"].([]interface{}); ok {
						material.textureOffset.restAt([2]float32{float32(offset[0].(float64)), float32(offset[1].(float64))})
					}
					if rotation, ok := ext.(map[string]interface{})["rotation"].(float64); ok {
						material.textureRotation.restAt(float32(rotation))
					}
					if scale, ok := ext.(map[string]interface{})["scale"].([]interface{}); ok {
						material.textureScale.restAt([2]float32{float32(scale[0].(float64)), float32(scale[1].(float64))})
					}
				}
			}
		}
		if m.NormalTexture != nil {
			material.normalMapIndex = new(uint32)
			*material.normalMapIndex = *m.NormalTexture.Index
			if m.NormalTexture.Extensions != nil {
				if l, ok := m.NormalTexture.Extensions["KHR_texture_transform"]; ok {
					var ext interface{}
					err := json.Unmarshal(l.(json.RawMessage), &ext)
					if err != nil {
						return nil, err
					}
					if offset, ok := ext.(map[string]interface{})["offset"].([]interface{}); ok {
						material.normalMapOffset.restAt([2]float32{float32(offset[0].(float64)), float32(offset[1].(float64))})
					}
					if rotation, ok := ext.(map[string]interface{})["rotation"].(float64); ok {
						material.normalMapRotation.restAt(float32(rotation))
					}
					if scale, ok := ext.(map[string]interface{})["scale"].([]interface{}); ok {
						material.normalMapScale.restAt([2]float32{float32(scale[0].(float64)), float32(scale[1].(float64))})
					}
				}
			}
		}
		if m.PBRMetallicRoughness.MetallicRoughnessTexture != nil {
			material.metallicRoughnessMapIndex = new(uint32)
			*material.metallicRoughnessMapIndex = m.PBRMetallicRoughness.MetallicRoughnessTexture.Index
			if m.PBRMetallicRoughness.MetallicRoughnessTexture.Extensions != nil {
				if l, ok := m.PBRMetallicRoughness.MetallicRoughnessTexture.Extensions["KHR_texture_transform"]; ok {
					var ext interface{}
					err := json.Unmarshal(l.(json.RawMessage), &ext)
					if err != nil {
						return nil, err
					}
					if offset, ok := ext.(map[string]interface{})["offset"].([]interface{}); ok {
						material.metallicRoughnessMapOffset.restAt([2]float32{float32(offset[0].(float64)), float32(offset[1].(float64))})
					}
					if rotation, ok := ext.(map[string]interface{})["rotation"].(float64); ok {
						material.metallicRoughnessMapRotation.restAt(float32(rotation))
					}
					if scale, ok := ext.(map[string]interface{})["scale"].([]interface{}); ok {
						material.metallicRoughnessMapScale.restAt([2]float32{float32(scale[0].(float64)), float32(scale[1].(float64))})
					}
				}
			}
		}
		if m.PBRMetallicRoughness.BaseColorFactor != nil {
			material.baseColorFactor.restAt(*m.PBRMetallicRoughness.BaseColorFactor)
		}
		if m.PBRMetallicRoughness.RoughnessFactor != nil {
			material.roughness.restAt(*m.PBRMetallicRoughness.RoughnessFactor)
		}
		if m.PBRMetallicRoughness.MetallicFactor != nil {
			material.metallic.restAt(*m.PBRMetallicRoughness.MetallicFactor)
		}

		if m.OcclusionTexture != nil {
			material.ambientOcclusionMapIndex = new(uint32)
			*material.ambientOcclusionMapIndex = *m.OcclusionTexture.Index
			if m.OcclusionTexture.Strength != nil {
				material.ambientOcclusion.restAt(*m.OcclusionTexture.Strength)
			}
			if m.OcclusionTexture.Extensions != nil {
				if l, ok := m.OcclusionTexture.Extensions["KHR_texture_transform"]; ok {
					var ext interface{}
					err := json.Unmarshal(l.(json.RawMessage), &ext)
					if err != nil {
						return nil, err
					}
					if offset, ok := ext.(map[string]interface{})["offset"].([]interface{}); ok {
						material.ambientOcclusionMapOffset.restAt([2]float32{float32(offset[0].(float64)), float32(offset[1].(float64))})
					}
					if rotation, ok := ext.(map[string]interface{})["rotation"].(float64); ok {
						material.ambientOcclusionMapRotation.restAt(float32(rotation))
					}
					if scale, ok := ext.(map[string]interface{})["scale"].([]interface{}); ok {
						material.ambientOcclusionMapScale.restAt([2]float32{float32(scale[0].(float64)), float32(scale[1].(float64))})
					}
				}
			}
		} else {
			material.ambientOcclusion.restAt(float32(0))
		}
		material.emission.restAt(m.EmissiveFactor)
		if m.EmissiveTexture != nil {
			material.emissionMapIndex = new(uint32)
			*material.emissionMapIndex = m.EmissiveTexture.Index
			if m.EmissiveTexture.Extensions != nil {
				if l, ok := m.EmissiveTexture.Extensions["KHR_texture_transform"]; ok {
					var ext interface{}
					err := json.Unmarshal(l.(json.RawMessage), &ext)
					if err != nil {
						return nil, err
					}
					if offset, ok := ext.(map[string]interface{})["offset"].([]interface{}); ok {
						material.emissionMapOffset.restAt([2]float32{float32(offset[0].(float64)), float32(offset[1].(float64))})
					}
					if rotation, ok := ext.(map[string]interface{})["rotation"].(float64); ok {
						material.emissionMapRotation.restAt(float32(rotation))
					}
					if scale, ok := ext.(map[string]interface{})["scale"].([]interface{}); ok {
						material.emissionMapScale.restAt([2]float32{float32(scale[0].(float64)), float32(scale[1].(float64))})
					}
				}
			}
		}
		material.name = m.Name
		material.alphaMode, _ = map[gltf.AlphaMode]AlphaMode{
			gltf.AlphaOpaque: AlphaModeOpaque,
			gltf.AlphaMask:   AlphaModeMask,
			gltf.AlphaBlend:  AlphaModeBlend,
		}[m.AlphaMode]
		if material.alphaMode == AlphaModeMask {
			material.alphaCutoff.restAt(m.AlphaCutoffOrDefault())
		}
		material.doubleSided = m.DoubleSided
		material.unlit = false
		if m.Extensions != nil {
			if _, ok := m.Extensions["KHR_materials_unlit"]; ok {
				material.unlit = true
			}
		}

		mdl.materials = append(mdl.materials, material)
	}
	if doc.Extensions != nil {
		if lightExtension, ok := doc.Extensions["KHR_lights_punctual"]; ok {
			var ext interface{}
			err := json.Unmarshal(lightExtension.(json.RawMessage), &ext)
			if err != nil {
				return nil, err
			}
			for _, light := range ext.(map[string]interface{})["lights"].([]interface{}) {
				params := light.(map[string]interface{})
				newLight := GLTFLight{
					intensity:       GLTFAnimatableProperty{isAnimated: false, restValue: float32(1.0)},
					color:           GLTFAnimatableProperty{restValue: [3]float32{1, 1, 1}},
					lightRange:      GLTFAnimatableProperty{isAnimated: false, restValue: float32(-1.0)},
					innerConeAngle:  GLTFAnimatableProperty{isAnimated: false, restValue: float32(0.0)},
					outerConeAngle:  GLTFAnimatableProperty{isAnimated: false, restValue: float32(math.Pi / 4)},
					shadowMapNear:   GLTFAnimatableProperty{isAnimated: false, restValue: float32(0.0)},
					shadowMapFar:    GLTFAnimatableProperty{isAnimated: false, restValue: float32(0.0)},
					shadowMapBottom: GLTFAnimatableProperty{isAnimated: false, restValue: float32(0.0)},
					shadowMapTop:    GLTFAnimatableProperty{isAnimated: false, restValue: float32(0.0)},
					shadowMapLeft:   GLTFAnimatableProperty{isAnimated: false, restValue: float32(0.0)},
					shadowMapRight:  GLTFAnimatableProperty{isAnimated: false, restValue: float32(0.0)},
					shadowMapBias:   GLTFAnimatableProperty{isAnimated: false, restValue: float32(0.0)},
				}
				lightType := params["type"].(string)
				switch lightType {
				case "point":
					newLight.lightType = PointLight
				case "spot":
					newLight.lightType = SpotLight
				case "directional":
					newLight.lightType = DirectionalLight
				}
				if intensity, ok := params["intensity"]; ok {
					newLight.intensity.restAt((float32)(intensity.(float64)))
				}
				if lightRange, ok := params["range"]; ok {
					newLight.lightRange.restAt((float32)(lightRange.(float64)))
				}
				if spot, ok := params["spot"]; ok {
					if outerConeAngle, ok := spot.(map[string]interface{})["outerConeAngle"]; ok {
						newLight.outerConeAngle.restAt((float32)(outerConeAngle.(float64)))
					}
					if innerConeAngle, ok := spot.(map[string]interface{})["innerConeAngle"]; ok {
						newLight.innerConeAngle.restAt((float32)(innerConeAngle.(float64)))
					}
				}
				if color, ok := params["color"]; ok {
					colors := color.([]interface{})
					newLight.color.restAt([3]float32{(float32)(colors[0].(float64)), (float32)(colors[1].(float64)), (float32)(colors[2].(float64))})
				}

				if extraParams, ok := params["extras"]; ok {

					v, ok := extraParams.(map[string]interface{})
					if ok {
						if v["shadowMapNear"] != nil {
							newLight.shadowMapNear.restAt((float32)(v["shadowMapNear"].(float64)))
						}
						if v["shadowMapFar"] != nil {
							newLight.shadowMapFar.restAt((float32)(v["shadowMapFar"].(float64)))
						}
						if v["shadowMapBottom"] != nil {
							newLight.shadowMapBottom.restAt((float32)(v["shadowMapBottom"].(float64)))
						}
						if v["shadowMapTop"] != nil {
							newLight.shadowMapTop.restAt((float32)(v["shadowMapTop"].(float64)))
						}
						if v["shadowMapLeft"] != nil {
							newLight.shadowMapLeft.restAt((float32)(v["shadowMapLeft"].(float64)))
						}
						if v["shadowMapRight"] != nil {
							newLight.shadowMapRight.restAt((float32)(v["shadowMapRight"].(float64)))
						}
						if v["shadowMapBias"] != nil {
							newLight.shadowMapBias.restAt((float32)(v["shadowMapBias"].(float64)))
						}
					}
				}
				mdl.lights = append(mdl.lights, newLight)
			}
		}
	}

	var vertexBuffer []byte
	var elementBuffer []uint32
	mdl.meshes = make([]*Mesh, 0, len(doc.Meshes))
	for _, m := range doc.Meshes {
		var mesh = &Mesh{}
		mesh.name = m.Name
		mesh.morphTargetWeights = GLTFAnimatableProperty{isAnimated: false, restValue: m.Weights}
		for _, p := range m.Primitives {
			var primitive = &Primitive{}
			primitive.boundingBox.min = [3]float32{math.MaxFloat32, math.MaxFloat32, math.MaxFloat32}
			primitive.boundingBox.max = [3]float32{-math.MaxFloat32, -math.MaxFloat32, -math.MaxFloat32}
			primitive.vertexBufferOffset = uint32(len(vertexBuffer))
			primitive.elementBufferOffset = uint32(4 * len(elementBuffer))
			var posBuffer [][3]float32
			positions, err := modeler.ReadPosition(doc, doc.Accessors[p.Attributes[gltf.POSITION]], posBuffer)
			if err != nil {
				return nil, err
			}
			primitive.numVertices = uint32(len(positions))

			for i := 0; i < int(primitive.numVertices); i++ {
				vertexBuffer = append(vertexBuffer, byte(i%256), byte((i>>8)%256), byte((i>>16)%256), byte((i>>32)%256))
			}

			for _, pos := range positions {
				for posIdx := range pos {
					if primitive.boundingBox.min[posIdx] > pos[posIdx] {
						primitive.boundingBox.min[posIdx] = pos[posIdx]
					}
					if primitive.boundingBox.max[posIdx] < pos[posIdx] {
						primitive.boundingBox.max[posIdx] = pos[posIdx]
					}
				}
				vertexBuffer = append(vertexBuffer, f32.Bytes(binary.LittleEndian, pos[:]...)...)
			}
			if idx, ok := p.Attributes[gltf.TEXCOORD_0]; ok {
				var uvBuffer [][2]float32
				texCoords, err := modeler.ReadTextureCoord(doc, doc.Accessors[idx], uvBuffer)
				if err != nil {
					return nil, err
				}
				if len(texCoords) > 0 {
					primitive.useUV = true
					for _, tex := range texCoords {
						vertexBuffer = append(vertexBuffer, f32.Bytes(binary.LittleEndian, tex[:]...)...)
					}
				} else {
					primitive.useUV = false
				}
			} else {
				primitive.useUV = false
			}
			if idx, ok := p.Attributes[gltf.NORMAL]; ok {
				var normalBuffer [][3]float32
				normals, err := modeler.ReadNormal(doc, doc.Accessors[idx], normalBuffer)
				if err != nil {
					return nil, err
				}
				if len(normals) > 0 {
					primitive.useNormal = true
					for _, tex := range normals {
						vertexBuffer = append(vertexBuffer, f32.Bytes(binary.LittleEndian, tex[:]...)...)
					}
				} else {
					primitive.useNormal = false
				}
			} else {
				primitive.useNormal = false
			}
			if idx, ok := p.Attributes[gltf.TANGENT]; ok {
				var tangentBuffer [][4]float32
				tangents, err := modeler.ReadTangent(doc, doc.Accessors[idx], tangentBuffer)
				if err != nil {
					return nil, err
				}
				if len(tangents) > 0 {
					primitive.useTangent = true
					for _, tex := range tangents {
						vertexBuffer = append(vertexBuffer, f32.Bytes(binary.LittleEndian, tex[:]...)...)
					}
				} else {
					primitive.useTangent = false
				}
			} else {
				primitive.useTangent = false
			}
			var indexBuffer []uint32
			indices, err := modeler.ReadIndices(doc, doc.Accessors[*p.Indices], indexBuffer)
			if err != nil {
				return nil, err
			}
			for _, p := range indices {
				elementBuffer = append(elementBuffer, p)
			}
			primitive.numIndices = uint32(len(indices))
			if idx, ok := p.Attributes[gltf.COLOR_0]; ok {
				primitive.useVertexColor = true
				switch doc.Accessors[idx].ComponentType {
				case gltf.ComponentUbyte:
					if doc.Accessors[idx].Type == gltf.AccessorVec3 {
						var vecBuffer [][3]uint8
						vecs, err := modeler.ReadAccessor(doc, doc.Accessors[idx], vecBuffer)
						if err != nil {
							return nil, err
						}
						for _, vec := range vecs.([][3]uint8) {
							vertexBuffer = append(vertexBuffer, f32.Bytes(binary.LittleEndian, float32(vec[0])/255, float32(vec[1])/255, float32(vec[2])/255, 1)...)
						}
					} else {
						var vecBuffer [][4]uint8
						vecs, err := modeler.ReadAccessor(doc, doc.Accessors[idx], vecBuffer)
						if err != nil {
							return nil, err
						}
						for _, vec := range vecs.([][4]uint8) {
							vertexBuffer = append(vertexBuffer, f32.Bytes(binary.LittleEndian, float32(vec[0])/255, float32(vec[1])/255, float32(vec[2])/255, float32(vec[3])/255)...)
						}
					}
				case gltf.ComponentUshort:
					if doc.Accessors[idx].Type == gltf.AccessorVec3 {
						var vecBuffer [][3]uint16
						vecs, err := modeler.ReadAccessor(doc, doc.Accessors[idx], vecBuffer)
						if err != nil {
							return nil, err
						}
						for _, vec := range vecs.([][3]uint16) {
							vertexBuffer = append(vertexBuffer, f32.Bytes(binary.LittleEndian, float32(vec[0])/65535, float32(vec[1])/65535, float32(vec[2])/65535, 1)...)
						}
					} else {
						var vecBuffer [][4]uint16
						vecs, err := modeler.ReadAccessor(doc, doc.Accessors[idx], vecBuffer)
						if err != nil {
							return nil, err
						}
						for _, vec := range vecs.([][4]uint16) {
							vertexBuffer = append(vertexBuffer, f32.Bytes(binary.LittleEndian, float32(vec[0])/65535, float32(vec[1])/65535, float32(vec[2])/65535, float32(vec[3])/65535)...)
						}
					}
				case gltf.ComponentFloat:
					if doc.Accessors[idx].Type == gltf.AccessorVec3 {
						var vecBuffer [][3]float32
						vecs, err := modeler.ReadAccessor(doc, doc.Accessors[idx], vecBuffer)
						if err != nil {
							return nil, err
						}
						for _, vec := range vecs.([][3]float32) {
							vertexBuffer = append(vertexBuffer, f32.Bytes(binary.LittleEndian, vec[0], vec[1], vec[2], 1)...)
						}
					} else {
						var vecBuffer [][4]float32
						vecs, err := modeler.ReadAccessor(doc, doc.Accessors[idx], vecBuffer)
						if err != nil {
							return nil, err
						}
						for _, vec := range vecs.([][4]float32) {
							vertexBuffer = append(vertexBuffer, f32.Bytes(binary.LittleEndian, vec[:]...)...)
						}
					}
				}
			} else {
				primitive.useVertexColor = false
			}
			if idx, ok := p.Attributes[gltf.JOINTS_0]; ok {
				primitive.useJoint0 = true
				var jointBuffer [][4]uint16
				joints, err := modeler.ReadJoints(doc, doc.Accessors[idx], jointBuffer)
				if err != nil {
					return nil, err
				}
				for _, joint := range joints {
					var f [4]float32
					for j, v := range joint {
						f[j] = float32(v)
					}
					vertexBuffer = append(vertexBuffer, f32.Bytes(binary.LittleEndian, f[:]...)...)
				}
				if idx, ok := p.Attributes[gltf.WEIGHTS_0]; ok {
					var weightBuffer [][4]float32
					weights, err := modeler.ReadWeights(doc, doc.Accessors[idx], weightBuffer)
					if err != nil {
						return nil, err
					}
					for _, weight := range weights {
						vertexBuffer = append(vertexBuffer, f32.Bytes(binary.LittleEndian, weight[:]...)...)
					}
				} else {
					return nil, errors.New("Primitive attribute JOINTS_0 is specified but WEIGHTS_0 is not specified.")
				}
				if idx, ok := p.Attributes["JOINTS_1"]; ok {
					primitive.useJoint1 = true
					var jointBuffer [][4]uint16
					joints, err := modeler.ReadJoints(doc, doc.Accessors[idx], jointBuffer)
					if err != nil {
						return nil, err
					}
					for _, joint := range joints {
						var f [4]float32
						for j, v := range joint {
							f[j] = float32(v)
						}
						vertexBuffer = append(vertexBuffer, f32.Bytes(binary.LittleEndian, f[:]...)...)
					}
					primitive.useJoint1 = false
					if idx, ok := p.Attributes["WEIGHTS_1"]; primitive.useJoint1 && ok {
						var weightBuffer [][4]float32
						weights, err := modeler.ReadWeights(doc, doc.Accessors[idx], weightBuffer)
						if err != nil {
							return nil, err
						}
						for _, weight := range weights {
							vertexBuffer = append(vertexBuffer, f32.Bytes(binary.LittleEndian, weight[:]...)...)
						}
					} else if primitive.useJoint1 {
						return nil, errors.New("Primitive attribute JOINTS_1 is specified but WEIGHTS_1 is not specified.")
					}
				}
			} else {
				primitive.useJoint0 = false
			}

			if idx, ok := p.Attributes["_OUTLINE_ATTRIBUTE"]; ok {
				primitive.useOutlineAttribute = true
				var outlineAttributeBuffer [][4]uint16
				atttributes, err := modeler.ReadAccessor(doc, doc.Accessors[idx], outlineAttributeBuffer)
				if err != nil {
					return nil, err
				}
				for _, attribute := range atttributes.([][4]float32) {
					vertexBuffer = append(vertexBuffer, f32.Bytes(binary.LittleEndian, attribute[:]...)...)
				}
			} else {
				primitive.useOutlineAttribute = false
			}
			if len(p.Targets) > 0 {
				numAttributes := 0
				for _, t := range p.Targets {
					numAttributes += len(t)
				}
				for _, t := range p.Targets {
					target := &MorphTarget{}
					for attr, accessor := range t {
						switch attr {
						case "POSITION":
							var posBuffer [][3]float32
							positions, err := modeler.ReadPosition(doc, doc.Accessors[accessor], posBuffer)
							if err != nil {
								return nil, err
							}
							target.positionBuffer = make([]float32, 0, 4*int(primitive.numVertices))
							for _, pos := range positions {
								target.positionBuffer = append(target.positionBuffer, pos[0], pos[1], pos[2], 0)
							}
						case "NORMAL":
							var posBuffer [][3]float32
							positions, err := modeler.ReadPosition(doc, doc.Accessors[accessor], posBuffer)
							if err != nil {
								return nil, err
							}
							target.normalBuffer = make([]float32, 0, 4*int(primitive.numVertices))
							for _, pos := range positions {
								target.normalBuffer = append(target.normalBuffer, pos[0], pos[1], pos[2], 0)
							}
						case "TANGENT":
							var posBuffer [][3]float32
							positions, err := modeler.ReadPosition(doc, doc.Accessors[accessor], posBuffer)
							if err != nil {
								return nil, err
							}
							target.tangentBuffer = make([]float32, 0, 4*int(primitive.numVertices))
							for _, pos := range positions {
								target.tangentBuffer = append(target.tangentBuffer, pos[0], pos[1], pos[2], 0)
							}
						case "TEXCOORD_0":
							var uvBuffer [][2]float32
							texCoords, err := modeler.ReadTextureCoord(doc, doc.Accessors[accessor], uvBuffer)
							if err != nil {
								return nil, err
							}
							target.uvBuffer = make([]float32, 0, 4*int(primitive.numVertices))
							for _, uv := range texCoords {
								target.uvBuffer = append(target.uvBuffer, uv[0], uv[1], 0, 0)
							}
						case "COLOR_0":
							target.colorBuffer = make([]float32, 0, 4*int(primitive.numVertices))
							switch doc.Accessors[accessor].ComponentType {
							case gltf.ComponentUbyte:
								if doc.Accessors[accessor].Type == gltf.AccessorVec3 {
									var vecBuffer [][3]uint8
									vecs, err := modeler.ReadAccessor(doc, doc.Accessors[accessor], vecBuffer)
									if err != nil {
										return nil, err
									}
									for _, vec := range vecs.([][3]uint8) {
										target.colorBuffer = append(target.colorBuffer, float32(vec[0])/255, float32(vec[1])/255, float32(vec[2])/255, 1)
									}
								} else {
									var vecBuffer [][4]uint8
									vecs, err := modeler.ReadAccessor(doc, doc.Accessors[accessor], vecBuffer)
									if err != nil {
										return nil, err
									}
									for _, vec := range vecs.([][4]uint8) {
										target.colorBuffer = append(target.colorBuffer, float32(vec[0])/255, float32(vec[1])/255, float32(vec[2])/255, float32(vec[3])/255)
									}
								}
							case gltf.ComponentUshort:
								if doc.Accessors[accessor].Type == gltf.AccessorVec3 {
									var vecBuffer [][3]uint16
									vecs, err := modeler.ReadAccessor(doc, doc.Accessors[accessor], vecBuffer)
									if err != nil {
										return nil, err
									}
									for _, vec := range vecs.([][3]uint16) {
										target.colorBuffer = append(target.colorBuffer, float32(vec[0])/65535, float32(vec[1])/65535, float32(vec[2])/65535, 1)
									}
								} else {
									var vecBuffer [][4]uint16
									vecs, err := modeler.ReadAccessor(doc, doc.Accessors[accessor], vecBuffer)
									if err != nil {
										return nil, err
									}
									for _, vec := range vecs.([][4]uint16) {
										target.colorBuffer = append(target.colorBuffer, float32(vec[0])/65535, float32(vec[1])/65535, float32(vec[2])/65535, float32(vec[3])/65535)
									}
								}
							case gltf.ComponentFloat:
								if doc.Accessors[accessor].Type == gltf.AccessorVec3 {
									var vecBuffer [][3]float32
									vecs, err := modeler.ReadAccessor(doc, doc.Accessors[accessor], vecBuffer)
									if err != nil {
										return nil, err
									}
									for _, vec := range vecs.([][3]float32) {
										target.colorBuffer = append(target.colorBuffer, vec[0], vec[1], vec[2], 1)
									}
								} else {
									var vecBuffer [][4]float32
									vecs, err := modeler.ReadAccessor(doc, doc.Accessors[accessor], vecBuffer)
									if err != nil {
										return nil, err
									}
									for _, vec := range vecs.([][4]float32) {
										target.colorBuffer = append(target.colorBuffer, vec[0], vec[1], vec[2], vec[3])
									}
								}
							}
						}
					}
					primitive.morphTargets = append(primitive.morphTargets, target)
				}
				primitive.morphTargetTexture = &GLTFTexture{}
				sys.mainThreadTask <- func() {
					dimension := int(math.Ceil(math.Pow(float64(8*primitive.numVertices), 0.5)))
					primitive.morphTargetTexture.tex = gfx.newDataTexture(int32(dimension), int32(dimension))
					//primitive.morphTargetTexture.tex.SetPixelData(targetBuffer)
				}
			}

			if p.Material != nil {
				primitive.materialIndex = new(uint32)
				*primitive.materialIndex = *p.Material
			}
			primitive.mode = gltfPrimitiveModeMap[p.Mode]
			mesh.primitives = append(mesh.primitives, primitive)
		}
		mdl.meshes = append(mdl.meshes, mesh)
	}
	mdl.vertexBuffer = vertexBuffer
	mdl.elementBuffer = elementBuffer

	mdl.nodes = make([]*Node, 0, len(doc.Nodes))
	var lightNodes []int32
	for idx, n := range doc.Nodes {
		var node = &Node{}
		mdl.nodes = append(mdl.nodes, node)
		node.visible = true
		node.rotation.restAt(n.Rotation)
		node.transition.restAt(n.Translation)
		node.scale.restAt(n.Scale)
		node.skin = n.Skin
		node.childrenIndex = n.Children
		node.morphTargetWeights = GLTFAnimatableProperty{isAnimated: false, restValue: n.Weights}
		if n.Mesh != nil {
			node.meshIndex = new(uint32)
			*node.meshIndex = *n.Mesh
			if len(n.Weights) == 0 {
				m := mdl.meshes[*n.Mesh]
				node.morphTargetWeights.restAt(m.morphTargetWeights.getValue())
			}
		}
		node.trans = TransNone
		node.castShadow = true
		node.zTest = true
		node.zWrite = true
		node.meshOutline = GLTFAnimatableProperty{isAnimated: false, restValue: float32(0.0)}
		if n.Extensions != nil {
			if l, ok := n.Extensions["KHR_lights_punctual"]; ok {
				var ext interface{}
				err := json.Unmarshal(l.(json.RawMessage), &ext)
				if err != nil {
					return nil, err
				}
				lightNodes = append(lightNodes, int32(idx))
				node.lightIndex = new(uint32)
				*node.lightIndex = (uint32)(ext.(map[string]interface{})["light"].(float64))
				node.shadowMapNear = GLTFAnimatableProperty{isAnimated: false, restValue: float32(0.0)}
				node.shadowMapFar = GLTFAnimatableProperty{isAnimated: false, restValue: float32(0.0)}
				node.shadowMapBottom = GLTFAnimatableProperty{isAnimated: false, restValue: float32(0.0)}
				node.shadowMapTop = GLTFAnimatableProperty{isAnimated: false, restValue: float32(0.0)}
				node.shadowMapLeft = GLTFAnimatableProperty{isAnimated: false, restValue: float32(0.0)}
				node.shadowMapRight = GLTFAnimatableProperty{isAnimated: false, restValue: float32(0.0)}
				node.shadowMapBias = GLTFAnimatableProperty{isAnimated: false, restValue: float32(0.0)}
			}
		}
		if n.Extras != nil {
			v, ok := n.Extras.(map[string]interface{})
			if ok {
				switch v["trans"] {
				case "ADD":
					node.trans = TransAdd
				case "SUB":
					node.trans = TransReverseSubtract
				case "MUL":
					node.trans = TransMul
				case "NONE":
					node.trans = TransNone
				}
				if v["disableZTest"] != nil && v["disableZTest"] != "0" && v["disableZTest"] != "false" {
					node.zTest = false
				}
				if v["disableZWrite"] != nil && v["disableZWrite"] != "0" && v["disableZWrite"] != "false" {
					node.zWrite = false
				}
				if v["castShadow"] != nil && (v["castShadow"] == "0" || v["castShadow"] == "false") {
					node.castShadow = false
				}
				if v["shadowMapNear"] != nil {
					node.shadowMapNear.restAt((float32)(v["shadowMapNear"].(float64)))
				}
				if v["shadowMapFar"] != nil {
					node.shadowMapFar.restAt((float32)(v["shadowMapFar"].(float64)))
				}
				if v["shadowMapBottom"] != nil {
					node.shadowMapBottom.restAt((float32)(v["shadowMapBottom"].(float64)))
				}
				if v["shadowMapTop"] != nil {
					node.shadowMapTop.restAt((float32)(v["shadowMapTop"].(float64)))
				}
				if v["shadowMapLeft"] != nil {
					node.shadowMapLeft.restAt((float32)(v["shadowMapLeft"].(float64)))
				}
				if v["shadowMapRight"] != nil {
					node.shadowMapRight.restAt((float32)(v["shadowMapRight"].(float64)))
				}
				if v["shadowMapBias"] != nil {
					node.shadowMapBias.restAt((float32)(v["shadowMapBias"].(float64)))
				}
				if v["id"] != nil {
					node.id = uint32(v["id"].(float64))
				}
				if v["layerNumber"] != nil {
					node.layerNumber = new(int)
					*node.layerNumber = int(v["layerNumber"].(float64))
				}
				if v["meshOutline"] != nil {
					node.meshOutline.restAt((float32)(v["meshOutline"].(float64)))
				}
			}
		}
		node.transformChanged = true
	}
	mdl.animationTimeStamps = map[uint32][]float32{}
	for _, a := range doc.Animations {
		anim := &GLTFAnimation{}
		mdl.animations = append(mdl.animations, anim)
		anim.duration = 0
		anim.name = a.Name
		anim.enabled = true
		anim.defaultEnabled = true
		anim.loopCount = -1
		anim.loop = 0
		for _, c := range a.Channels {
			channel := &GLTFAnimationChannel{}
			channel.nodeIndex = c.Target.Node
			channel.samplerIndex = *c.Sampler
			channel.target = nil
			if c.Target.Extensions != nil {
				if p, ok := c.Target.Extensions["KHR_animation_pointer"]; ok {
					var ext interface{}
					if err := json.Unmarshal(p.(json.RawMessage), &ext); err != nil {
						return nil, err
					}
					pointer := ext.(map[string]interface{})["pointer"].(string)
					if err := channel.parseAnimationPointer(mdl, pointer); err != nil {
						return nil, err
					}
				}
			}
			if channel.target == nil {
				switch c.Target.Path {
				case gltf.TRSTranslation:
					channel.targetType = TRSTranslation
					channel.target = &mdl.nodes[*channel.nodeIndex].transition
				case gltf.TRSScale:
					channel.targetType = TRSScale
					channel.target = &mdl.nodes[*channel.nodeIndex].scale
				case gltf.TRSRotation:
					channel.targetType = TRSRotation
					channel.target = &mdl.nodes[*channel.nodeIndex].rotation
				case gltf.TRSWeights:
					channel.targetType = MorphTargetWeight
					channel.target = &mdl.nodes[*channel.nodeIndex].morphTargetWeights
				default:

					continue
				}
			}
			anim.channels = append(anim.channels, channel)
		}
		for _, s := range a.Samplers {
			sampler := &GLTFAnimationSampler{}
			anim.samplers = append(anim.samplers, sampler)
			if _, ok := mdl.animationTimeStamps[s.Input]; !ok {
				var timeBuffer []float32
				times, err := modeler.ReadAccessor(doc, doc.Accessors[s.Input], timeBuffer)
				if err != nil {
					return nil, err
				}
				mdl.animationTimeStamps[s.Input] = make([]float32, 0, len(times.([]float32)))
				for _, t := range times.([]float32) {
					mdl.animationTimeStamps[s.Input] = append(mdl.animationTimeStamps[s.Input], t)
				}
			}
			sampler.interpolation = GLTFAnimationInterpolation(s.Interpolation)
			sampler.inputIndex = s.Input
			if anim.duration < mdl.animationTimeStamps[s.Input][len(mdl.animationTimeStamps[s.Input])-1] {
				anim.duration = mdl.animationTimeStamps[s.Input][len(mdl.animationTimeStamps[s.Input])-1]
			}
			switch doc.Accessors[s.Output].Type {
			case gltf.AccessorScalar:
				var vecBuffer []float32
				vecs, err := modeler.ReadAccessor(doc, doc.Accessors[s.Output], vecBuffer)
				if err != nil {
					return nil, err
				}
				sampler.output = make([]float32, 0, len(vecs.([]float32)))
				for _, val := range vecs.([]float32) {
					sampler.output = append(sampler.output, val)
				}
			case gltf.AccessorVec3:
				var vecBuffer [][3]float32
				vecs, err := modeler.ReadAccessor(doc, doc.Accessors[s.Output], vecBuffer)
				if err != nil {
					return nil, err
				}
				sampler.output = make([]float32, 0, len(vecs.([][3]float32))*3)
				for _, vec := range vecs.([][3]float32) {
					sampler.output = append(sampler.output, vec[0], vec[1], vec[2])
				}
			case gltf.AccessorVec4:
				var vecBuffer [][4]float32
				vecs, err := modeler.ReadAccessor(doc, doc.Accessors[s.Output], vecBuffer)
				if err != nil {
					return nil, err
				}
				sampler.output = make([]float32, 0, len(vecs.([][4]float32))*4)
				for _, vec := range vecs.([][4]float32) {
					sampler.output = append(sampler.output, vec[0], vec[1], vec[2], vec[3])
				}
			}
		}
		if a.Extras != nil {
			v, ok := a.Extras.(map[string]interface{})
			if ok {
				if v["id"] != nil {
					anim.id = uint32(v["id"].(float64))
				}
				if v["loopCount"] != nil {
					anim.loopCount = int32(v["loopCount"].(float64))
				}
				if v["enabled"] != nil {
					anim.enabled = v["enabled"] != "0" && v["enabled"] != "false"
					anim.defaultEnabled = anim.enabled
				}
			}
		}
	}
	for _, s := range doc.Skins {
		var skin = &Skin{}
		for _, j := range s.Joints {
			skin.joints = append(skin.joints, j)
		}

		if s.InverseBindMatrices != nil {
			var matrixBuffer [][4][4]float32
			matrices, err := modeler.ReadAccessor(doc, doc.Accessors[*s.InverseBindMatrices], matrixBuffer)
			if err != nil {
				return nil, err
			}
			for _, mat := range matrices.([][4][4]float32) {
				skin.inverseBindMatrices = append(skin.inverseBindMatrices, mat[0][:]...)
				skin.inverseBindMatrices = append(skin.inverseBindMatrices, mat[1][:]...)
				skin.inverseBindMatrices = append(skin.inverseBindMatrices, mat[2][:]...)
			}
		}

		skin.texture = &GLTFTexture{}
		sys.mainThreadTask <- func() {
			skin.texture.tex = gfx.newDataTexture(6, int32(len(skin.joints)))
		}

		mdl.skins = append(mdl.skins, skin)
	}

	for _, s := range doc.Scenes {
		var scene = &Scene{}
		scene.name = s.Name
		scene.nodes = s.Nodes
		for _, n := range s.Nodes {
			scene.getSceneLight(n, mdl.nodes)
		}
		mdl.scenes = append(mdl.scenes, scene)
	}
	return mdl, nil
}

func (s *Scene) getSceneLight(n uint32, nodes []*Node) {
	node := nodes[n]
	for _, c := range node.childrenIndex {
		s.getSceneLight(c, nodes)
	}
	if node.lightIndex != nil {
		s.lightNodes = append(s.lightNodes, n)
	}
}

func (n *Node) getLocalTransform() (mat mgl.Mat4) {
	mat = mgl.Ident4()
	if n.transformChanged {
		t := n.transition.getValue().([3]float32)
		mat = mgl.Translate3D(t[0], t[1], t[2])
		r := n.rotation.getValue().([4]float32)
		mat = mat.Mul4(mgl.Quat{W: r[3], V: mgl.Vec3{r[0], r[1], r[2]}}.Mat4())
		s := n.scale.getValue().([3]float32)
		mat = mat.Mul4(mgl.Scale3D(s[0], s[1], s[2]))
		n.localTransform = mat
		n.transformChanged = false
	} else {
		mat = n.localTransform
	}
	return
}

func (n *Node) calculateWorldTransform(parentTransorm mgl.Mat4, nodes []*Node) {
	mat := n.getLocalTransform()
	n.worldTransform = parentTransorm.Mul4(mat)
	if n.meshIndex != nil {
		n.normalMatrix = n.worldTransform.Inv().Transpose()
	}
	if n.lightIndex != nil {
		scale := [3]float32{n.worldTransform.Col(0).Len(), n.worldTransform.Col(1).Len(), n.worldTransform.Col(2).Len()}
		mat := mgl.Ident4()
		for i := 0; i < 3; i++ {
			mat[i] = n.worldTransform[i] / scale[0]
			mat[i+4] = n.worldTransform[i+4] / scale[1]
			mat[i+8] = n.worldTransform[i+8] / scale[2]
		}
		quat := mgl.Mat4ToQuat(mat).Normalize()
		direction := mgl.Vec3{0, 0, -1}
		n.lightDirection = quat.Rotate(direction)
	}
	for _, index := range n.childrenIndex {
		(*nodes[index]).calculateWorldTransform(n.worldTransform, nodes)
	}
	return
}

func (mdl *Model) calculateTextureTransform() {
	for _, m := range mdl.materials {
		if index := m.textureIndex; index != nil {
			t := m.textureOffset.getValue().([2]float32)
			mat := mgl.Translate2D(t[0], t[1])
			r := m.textureRotation.getValue().(float32)
			mat = mat.Mul3(mgl.HomogRotate2D(r).Transpose())
			s := m.textureScale.getValue().([2]float32)
			mat = mat.Mul3(mgl.Scale2D(s[0], s[1]))
			m.textureTransform = mat
		}
		if index := m.normalMapIndex; index != nil {
			t := m.normalMapOffset.getValue().([2]float32)
			mat := mgl.Translate2D(t[0], t[1])
			r := m.normalMapRotation.getValue().(float32)
			mat = mat.Mul3(mgl.HomogRotate2D(r).Transpose())
			s := m.normalMapScale.getValue().([2]float32)
			mat = mat.Mul3(mgl.Scale2D(s[0], s[1]))
			m.normalMapTransform = mat
		}
		if index := m.ambientOcclusionMapIndex; index != nil {
			t := m.ambientOcclusionMapOffset.getValue().([2]float32)
			mat := mgl.Translate2D(t[0], t[1])
			r := m.ambientOcclusionMapRotation.getValue().(float32)
			mat = mat.Mul3(mgl.HomogRotate2D(r).Transpose())
			s := m.ambientOcclusionMapScale.getValue().([2]float32)
			mat = mat.Mul3(mgl.Scale2D(s[0], s[1]))
			m.ambientOcclusionMapTransform = mat
		}
		if index := m.metallicRoughnessMapIndex; index != nil {
			t := m.metallicRoughnessMapOffset.getValue().([2]float32)
			mat := mgl.Translate2D(t[0], t[1])
			r := m.metallicRoughnessMapRotation.getValue().(float32)
			mat = mat.Mul3(mgl.HomogRotate2D(r).Transpose())
			s := m.metallicRoughnessMapScale.getValue().([2]float32)
			mat = mat.Mul3(mgl.Scale2D(s[0], s[1]))
			m.metallicRoughnessMapTransform = mat
		}
		if index := m.emissionMapIndex; index != nil {
			t := m.emissionMapOffset.getValue().([2]float32)
			mat := mgl.Translate2D(t[0], t[1])
			r := m.emissionMapRotation.getValue().(float32)
			mat = mat.Mul3(mgl.HomogRotate2D(r).Transpose())
			s := m.emissionMapScale.getValue().([2]float32)
			mat = mat.Mul3(mgl.Scale2D(s[0], s[1]))
			m.emissionMapTransform = mat
		}
	}
}

func calculateAnimationData(mdl *Model, n *Node) {
	for _, index := range n.childrenIndex {
		calculateAnimationData(mdl, mdl.nodes[index])
	}
	if n.meshIndex == nil {
		return
	}
	if n.skin != nil {
		mdl.skins[*n.skin].calculateSkinMatrices(n.worldTransform.Inv(), mdl.nodes)
	}
	m := mdl.meshes[*n.meshIndex]
	var morphTargetWeights []struct {
		index  uint32
		weight float32
	}
	weights := n.morphTargetWeights.getValue().([]float32)
	if len(weights) == 0 {
		weights = m.morphTargetWeights.getValue().([]float32)
	}
	activeMorphTargetChanged := false
	if len(weights) > 0 {
		activeMorphTargets := make([]uint32, 0, len(weights))
		morphTargetWeights = make([]struct {
			index  uint32
			weight float32
		}, 0, len(weights))
		for idx, w := range weights {
			if w != 0 {
				morphTargetWeights = append(morphTargetWeights, struct {
					index  uint32
					weight float32
				}{uint32(idx), w})
				activeMorphTargets = append(activeMorphTargets, uint32(idx))
			}
		}
		if len(activeMorphTargets) != len(n.activeMorphTargets) {
			activeMorphTargetChanged = true
			n.activeMorphTargets = activeMorphTargets
		} else {
			for i := range activeMorphTargets {
				if activeMorphTargets[i] != n.activeMorphTargets[i] {
					activeMorphTargetChanged = true
					n.activeMorphTargets = activeMorphTargets
				}
			}
		}
	}
	for _, p := range m.primitives {
		if p.materialIndex == nil {
			continue
		}
		if len(morphTargetWeights) > 0 && len(p.morphTargets) >= len(morphTargetWeights) {
			p.morphTargetWeight = [8]float32{0, 0, 0, 0, 0, 0, 0, 0}
			if activeMorphTargetChanged {
				width := p.morphTargetTexture.tex.GetWidth()
				targetBuffer := make([]float32, 4*width*width)
				count := 0
				offset := 0
				for _, t := range morphTargetWeights {
					morphTarget := p.morphTargets[t.index]
					if len(morphTarget.positionBuffer) > 0 {
						copy(targetBuffer[offset:offset+len(morphTarget.positionBuffer)], morphTarget.positionBuffer)
						offset += len(morphTarget.positionBuffer)
						p.morphTargetWeight[count] = t.weight
						count += 1
					}
				}
				p.morphTargetOffset[0] = float32(count)
				for _, t := range morphTargetWeights {
					morphTarget := p.morphTargets[t.index]
					if len(morphTarget.normalBuffer) > 0 {
						copy(targetBuffer[offset:offset+len(morphTarget.normalBuffer)], morphTarget.normalBuffer)
						offset += len(morphTarget.normalBuffer)
						p.morphTargetWeight[count] = t.weight
						count += 1
					}
				}
				p.morphTargetOffset[1] = float32(count)
				for _, t := range morphTargetWeights {
					morphTarget := p.morphTargets[t.index]
					if len(morphTarget.tangentBuffer) > 0 {
						copy(targetBuffer[offset:offset+len(morphTarget.tangentBuffer)], morphTarget.tangentBuffer)
						offset += len(morphTarget.tangentBuffer)
						p.morphTargetWeight[count] = t.weight
						count += 1
					}
				}
				p.morphTargetOffset[2] = float32(count)
				for _, t := range morphTargetWeights {
					morphTarget := p.morphTargets[t.index]
					if len(morphTarget.uvBuffer) > 0 {
						copy(targetBuffer[offset:offset+len(morphTarget.uvBuffer)], morphTarget.uvBuffer)
						offset += len(morphTarget.uvBuffer)
						p.morphTargetWeight[count] = t.weight
						count += 1
					}
				}
				p.morphTargetOffset[3] = float32(count)
				for _, t := range morphTargetWeights {
					morphTarget := p.morphTargets[t.index]
					if len(morphTarget.colorBuffer) > 0 {
						copy(targetBuffer[offset:offset+len(morphTarget.colorBuffer)], morphTarget.colorBuffer)
						offset += len(morphTarget.colorBuffer)
						targetBuffer = append(targetBuffer, morphTarget.colorBuffer...)
						p.morphTargetWeight[count] = t.weight
						count += 1
					}
				}
				p.morphTargetCount = uint32(count)
				if len(targetBuffer) > int(4*width*width) {
					targetBuffer = targetBuffer[:4*width*width]
				}
				p.morphTargetTexture.tex.SetPixelData(targetBuffer)
			} else {
				count := 0
				for _, t := range morphTargetWeights {
					morphTarget := p.morphTargets[t.index]
					if len(morphTarget.positionBuffer) > 0 {
						p.morphTargetWeight[count] = t.weight
						count += 1
					}
				}
				p.morphTargetOffset[0] = float32(count)
				for _, t := range morphTargetWeights {
					morphTarget := p.morphTargets[t.index]
					if len(morphTarget.normalBuffer) > 0 {
						p.morphTargetWeight[count] = t.weight
						count += 1
					}
				}
				p.morphTargetOffset[1] = float32(count)
				for _, t := range morphTargetWeights {
					morphTarget := p.morphTargets[t.index]
					if len(morphTarget.tangentBuffer) > 0 {
						p.morphTargetWeight[count] = t.weight
						count += 1
					}
				}
				p.morphTargetOffset[2] = float32(count)
				for _, t := range morphTargetWeights {
					morphTarget := p.morphTargets[t.index]
					if len(morphTarget.uvBuffer) > 0 {
						p.morphTargetWeight[count] = t.weight
						count += 1
					}
				}
				p.morphTargetOffset[3] = float32(count)
				for _, t := range morphTargetWeights {
					morphTarget := p.morphTargets[t.index]
					if len(morphTarget.colorBuffer) > 0 {
						p.morphTargetWeight[count] = t.weight
						count += 1
					}
				}

			}
		} else {
			p.morphTargetCount = 0
			p.morphTargetOffset = [4]float32{0, 0, 0, 0}
			p.morphTargetWeight = [8]float32{0, 0, 0, 0, 0, 0, 0, 0}
		}
	}
}

type Plane struct {
	Normal [3]float32
	D      float32 // Distance from the origin
}

// Extract planes from the Model-View-Projection (MVP) matrix
func ExtractFrustumPlanes(MVPMatrix mgl.Mat4) [6]Plane {
	var planes [6]Plane
	// Left plane
	planes[0] = Plane{
		Normal: [3]float32{MVPMatrix.At(3, 0) + MVPMatrix.At(0, 0), MVPMatrix.At(3, 1) + MVPMatrix.At(0, 1), MVPMatrix.At(3, 2) + MVPMatrix.At(0, 2)},
		D:      MVPMatrix.At(3, 3) + MVPMatrix.At(0, 3),
	}
	// Right plane
	planes[1] = Plane{
		Normal: [3]float32{MVPMatrix.At(3, 0) - MVPMatrix.At(0, 0), MVPMatrix.At(3, 1) - MVPMatrix.At(0, 1), MVPMatrix.At(3, 2) - MVPMatrix.At(0, 2)},
		D:      MVPMatrix.At(3, 3) - MVPMatrix.At(0, 3),
	}
	// Bottom plane
	planes[2] = Plane{
		Normal: [3]float32{MVPMatrix.At(3, 0) + MVPMatrix.At(1, 0), MVPMatrix.At(3, 1) + MVPMatrix.At(1, 1), MVPMatrix.At(3, 2) + MVPMatrix.At(1, 2)},
		D:      MVPMatrix.At(3, 3) + MVPMatrix.At(1, 3),
	}
	// Top plane
	planes[3] = Plane{
		Normal: [3]float32{MVPMatrix.At(3, 0) - MVPMatrix.At(1, 0), MVPMatrix.At(3, 1) - MVPMatrix.At(1, 1), MVPMatrix.At(3, 2) - MVPMatrix.At(1, 2)},
		D:      MVPMatrix.At(3, 3) - MVPMatrix.At(1, 3),
	}
	// Near plane
	planes[4] = Plane{
		Normal: [3]float32{MVPMatrix.At(3, 0) + MVPMatrix.At(2, 0), MVPMatrix.At(3, 1) + MVPMatrix.At(2, 1), MVPMatrix.At(3, 2) + MVPMatrix.At(2, 2)},
		D:      MVPMatrix.At(3, 3) + MVPMatrix.At(2, 3),
	}
	// Far plane
	planes[5] = Plane{
		Normal: [3]float32{MVPMatrix.At(3, 0) - MVPMatrix.At(2, 0), MVPMatrix.At(3, 1) - MVPMatrix.At(2, 1), MVPMatrix.At(3, 2) - MVPMatrix.At(2, 2)},
		D:      MVPMatrix.At(3, 3) - MVPMatrix.At(2, 3),
	}

	// Normalize the planes
	for i := 0; i < 6; i++ {
		length := float32(math.Sqrt(float64(planes[i].Normal[0]*planes[i].Normal[0] + planes[i].Normal[1]*planes[i].Normal[1] + planes[i].Normal[2]*planes[i].Normal[2])))
		planes[i].Normal[0] /= length
		planes[i].Normal[1] /= length
		planes[i].Normal[2] /= length
		planes[i].D /= length
	}

	return planes
}

func isCulled(MVPMatrix mgl.Mat4, box BoundingBox) bool {
	points := [8][3]float32{
		{box.min[0], box.min[1], box.min[2]},
		{box.min[0], box.min[1], box.max[2]},
		{box.min[0], box.max[1], box.min[2]},
		{box.min[0], box.max[1], box.max[2]},
		{box.max[0], box.min[1], box.min[2]},
		{box.max[0], box.min[1], box.max[2]},
		{box.max[0], box.max[1], box.min[2]},
		{box.max[0], box.max[1], box.max[2]},
	}

	for _, point := range points {
		clipSpace := MVPMatrix.Mul4x1(mgl.Vec4{point[0], point[1], point[2], 1})
		// Check if the point is within the normalized device coordinates
		if clipSpace[0] >= -clipSpace[3] && clipSpace[0] <= clipSpace[3] &&
			clipSpace[1] >= -clipSpace[3] && clipSpace[1] <= clipSpace[3] &&
			clipSpace[2] >= -clipSpace[3] && clipSpace[2] <= clipSpace[3] {
			return false // At least one point is within the frustum
		}
	}
	planes := ExtractFrustumPlanes(MVPMatrix)
	for _, plane := range planes {
		// Find the positive vertex
		var positive [3]float32
		if plane.Normal[0] >= 0 {
			positive[0] = box.max[0]
		} else {
			positive[0] = box.min[0]
		}
		if plane.Normal[1] >= 0 {
			positive[1] = box.max[1]
		} else {
			positive[1] = box.min[1]
		}
		if plane.Normal[2] >= 0 {
			positive[2] = box.max[2]
		} else {
			positive[2] = box.min[2]
		}
		// Check if the positive vertex is outside the plane
		if plane.Normal[0]*positive[0]+plane.Normal[1]*positive[1]+plane.Normal[2]*positive[2]+plane.D < 0 {
			return true // Entire bounding box is outside the frustum
		}
	}

	return false
}

func drawNode(mdl *Model, scene *Scene, layerNumber int, defaultLayerNumber int, n *Node, camOffset [3]float32, drawBlended bool, unlit bool, viewProjMatrix mgl.Mat4, outlineConst float32) {
	//mat := n.getLocalTransform()
	//model = model.Mul4(mat)
	for _, index := range n.childrenIndex {
		drawNode(mdl, scene, layerNumber, defaultLayerNumber, mdl.nodes[index], camOffset, drawBlended, unlit, viewProjMatrix, outlineConst)
	}
	nodeLayerNumber := defaultLayerNumber
	if n.layerNumber != nil {
		nodeLayerNumber = *n.layerNumber
	}
	if n.meshIndex == nil || !n.visible || (nodeLayerNumber != layerNumber) {
		return
	}

	// Convert to our sprite blend modes for PalFX synthesis
	var blendMode TransType
	switch n.trans {
	case TransReverseSubtract:
		blendMode = TT_sub // The one that matters right now
	case TransAdd, TransMul:
		blendMode = TT_add
	default:
		blendMode = TT_none
	}

	alpha := [2]int32{255, 255}
	if n.trans == TransNone {
		alpha = [2]int32{255, 0}
	}

	neg, grayscale, padd, pmul, invblend, hue := mdl.pfx.getFinalPalFx(blendMode, alpha)

	blendEq := BlendAdd
	src := BlendOne
	dst := BlendOneMinusSrcAlpha
	switch n.trans {
	case TransAdd:
		if invblend == 3 {
			src = BlendOne
			dst = BlendOne
			blendEq = BlendReverseSubtract
			neg = false
			if invblend >= 1 {
				padd[0] = -padd[0]
				padd[1] = -padd[1]
				padd[2] = -padd[2]
			}
		} else {
			src = BlendOne
			dst = BlendOne
		}
	case TransReverseSubtract:
		if invblend == 3 {
			src = BlendOne
			dst = BlendOne
			neg = false
			if invblend >= 1 {
				padd[0] = -padd[0]
				padd[1] = -padd[1]
				padd[2] = -padd[2]
			}
		} else {
			src = BlendOne
			dst = BlendOne
			blendEq = BlendReverseSubtract
		}
	case TransMul:
		if invblend == 3 {
			//Not accurate
			src = BlendOneMinusDstColor
			dst = BlendOne
			neg = false
			blendEq = BlendReverseSubtract
		} else {
			src = BlendDstColor
			dst = BlendOneMinusSrcAlpha
		}
	default:
		src = BlendOne
		dst = BlendOneMinusSrcAlpha
	}
	m := mdl.meshes[*n.meshIndex]
	reverseCull := n.worldTransform.Det() < 0
	for _, p := range m.primitives {
		if p.materialIndex == nil {
			continue
		}
		if n.skin == nil && p.morphTargetCount == 0 && isCulled(viewProjMatrix.Mul4(n.worldTransform), p.boundingBox) {
			continue
		}

		mat := mdl.materials[*p.materialIndex]
		if ((mat.alphaMode != AlphaModeBlend && n.trans == TransNone) && drawBlended) ||
			((mat.alphaMode == AlphaModeBlend || n.trans != TransNone) && !drawBlended) {
			return
		}
		color := mdl.materials[*p.materialIndex].baseColorFactor.getValue().([4]float32)
		meshOutline := n.meshOutline.getValue().(float32)
		gfx.SetModelPipeline(blendEq, src, dst, n.zTest, n.zWrite, mdl.materials[*p.materialIndex].doubleSided, reverseCull, p.useUV, p.useNormal, p.useTangent, p.useVertexColor, p.useJoint0, p.useJoint1, p.useOutlineAttribute, p.numVertices, p.vertexBufferOffset)

		gfx.SetModelUniformMatrix("model", n.worldTransform[:])
		gfx.SetModelUniformMatrix("normalMatrix", n.normalMatrix[:])
		gfx.SetModelUniformI("numVertices", int(p.numVertices))
		//gfx.SetModelUniformF("ambientOcclusion", 1)
		gfx.SetModelUniformF("metallicRoughness", mat.metallic.getValue().(float32), mat.roughness.getValue().(float32))
		gfx.SetModelUniformF("ambientOcclusionStrength", mat.ambientOcclusion.getValue().(float32))
		gfx.SetModelUniformMatrix3("texTransform", mat.textureTransform[:])
		gfx.SetModelUniformMatrix3("normalMapTransform", mat.normalMapTransform[:])
		gfx.SetModelUniformMatrix3("metallicRoughnessMapTransform", mat.metallicRoughnessMapTransform[:])
		gfx.SetModelUniformMatrix3("ambientOcclusionMapTransform", mat.ambientOcclusionMapTransform[:])
		gfx.SetModelUniformMatrix3("emissionMapTransform", mat.emissionMapTransform[:])

		gfx.SetModelUniformF("cameraPosition", -camOffset[0], -camOffset[1], -camOffset[2])

		if n.skin != nil {
			skin := mdl.skins[*n.skin]
			gfx.SetModelTexture("jointMatrices", skin.texture.tex)
		}

		if p.morphTargetCount > 0 {
			gfx.SetModelUniformF("morphTargetOffset", p.morphTargetOffset[0], p.morphTargetOffset[1], p.morphTargetOffset[2], p.morphTargetOffset[3])
			gfx.SetModelUniformI("numTargets", int(Min(int32(p.morphTargetCount), 8)))
			gfx.SetModelTexture("morphTargetValues", p.morphTargetTexture.tex)
			gfx.SetModelUniformFv("morphTargetWeight", p.morphTargetWeight[:])
			gfx.SetModelUniformI("morphTargetTextureDimension", int(p.morphTargetTexture.tex.GetWidth()))
		} else {
			gfx.SetModelUniformFv("morphTargetWeight", make([]float32, 8))
		}
		mode := p.mode
		if sys.wireframeDisplay {
			mode = 1 // Set mesh render mode to "lines"
		}
		gfx.SetModelUniformI("unlit", int(Btoi(unlit || mat.unlit)))
		gfx.SetModelUniformFv("add", padd[:])
		gfx.SetModelUniformFv("mult", []float32{pmul[0] * sys.brightness, pmul[1] * sys.brightness, pmul[2] * sys.brightness})
		gfx.SetModelUniformI("neg", int(Btoi(neg)))
		gfx.SetModelUniformF("hue", hue)
		gfx.SetModelUniformF("gray", grayscale)
		gfx.SetModelUniformI("enableAlpha", int(Btoi(mat.alphaMode == AlphaModeBlend)))
		gfx.SetModelUniformF("alphaThreshold", mat.alphaCutoff.getValue().(float32))
		gfx.SetModelUniformFv("baseColorFactor", color[:])
		if n.skin != nil {
			gfx.SetModelUniformI("numJoints", len(mdl.skins[*n.skin].joints))
		}
		if index := mat.textureIndex; index != nil {
			gfx.SetModelTexture("tex", mdl.textures[*index].tex)
			gfx.SetModelUniformI("useTexture", 1)
		} else {
			gfx.SetModelUniformI("useTexture", 0)
		}
		if index := mat.normalMapIndex; index != nil {
			gfx.SetModelTexture("normalMap", mdl.textures[*index].tex)
			gfx.SetModelUniformI("useNormalMap", 1)
		} else {
			gfx.SetModelUniformI("useNormalMap", 0)
		}
		if index := mat.metallicRoughnessMapIndex; index != nil {
			gfx.SetModelTexture("metallicRoughnessMap", mdl.textures[*index].tex)
			gfx.SetModelUniformI("useMetallicRoughnessMap", 1)
		} else {
			gfx.SetModelUniformI("useMetallicRoughnessMap", 0)
		}
		if index := mat.ambientOcclusionMapIndex; index != nil {
			gfx.SetModelTexture("ambientOcclusionMap", mdl.textures[*index].tex)
		}
		emission := mat.emission.getValue().([3]float32)
		gfx.SetModelUniformFv("emission", emission[:])
		if index := mat.emissionMapIndex; index != nil {
			gfx.SetModelTexture("emissionMap", mdl.textures[*index].tex)
			gfx.SetModelUniformI("useEmissionMap", 1)
		} else {
			gfx.SetModelUniformI("useEmissionMap", 0)
		}
		gfx.SetModelUniformF("meshOutline", 0)
		gfx.RenderElements(mode, int(p.numIndices), int(p.elementBufferOffset))
		if meshOutline > 0 {
			gfx.SetMeshOutlinePipeline(!reverseCull, meshOutline*outlineConst)
			gfx.RenderElements(mode, int(p.numIndices), int(p.elementBufferOffset))
		}

	}
}

func drawNodeShadow(mdl *Model, scene *Scene, n *Node, camOffset [3]float32, drawBlended bool, lightIndex int, numLights int, viewProjMatrices []mgl.Mat4, lightTypes []LightType) {
	//mat := n.getLocalTransform()
	//model = model.Mul4(mat)
	for _, index := range n.childrenIndex {
		drawNodeShadow(mdl, scene, mdl.nodes[index], camOffset, drawBlended, lightIndex, numLights, viewProjMatrices, lightTypes)
	}
	if n.meshIndex == nil || !n.visible {
		return
	}

	if n.trans == TransAdd || n.trans == TransReverseSubtract || n.trans == TransMul || !n.zTest || !n.zWrite || !n.castShadow {
		return
	}
	m := mdl.meshes[*n.meshIndex]
	reverseCull := n.worldTransform.Det() < 0
	for _, p := range m.primitives {
		if p.materialIndex == nil {
			continue
		}
		mat := mdl.materials[*p.materialIndex]
		if ((mat.alphaMode != AlphaModeBlend && n.trans == TransNone) && drawBlended) ||
			((mat.alphaMode == AlphaModeBlend || n.trans != TransNone) && !drawBlended) {
			return
		}
		color := mdl.materials[*p.materialIndex].baseColorFactor.getValue().([4]float32)
		if color[3] == 0 && mat.alphaMode == AlphaModeBlend {
			return
		}
		gfx.setShadowMapPipeline(mdl.materials[*p.materialIndex].doubleSided, reverseCull, p.useUV, p.useNormal, p.useTangent, p.useVertexColor, p.useJoint0, p.useJoint1, p.numVertices, p.vertexBufferOffset)

		gfx.SetShadowMapUniformMatrix("model", n.worldTransform[:])
		gfx.SetShadowMapUniformI("numVertices", int(p.numVertices))
		if n.skin != nil {
			skin := mdl.skins[*n.skin]
			gfx.SetShadowMapTexture("jointMatrices", skin.texture.tex)
		}

		if p.morphTargetCount > 0 {
			gfx.SetShadowMapUniformF("morphTargetOffset", p.morphTargetOffset[0], p.morphTargetOffset[1], p.morphTargetOffset[2], p.morphTargetOffset[3])
			gfx.SetShadowMapUniformI("numTargets", int(Min(int32(p.morphTargetCount), 8)))
			gfx.SetShadowMapTexture("morphTargetValues", p.morphTargetTexture.tex)
			gfx.SetShadowMapUniformFv("morphTargetWeight", p.morphTargetWeight[:])
			gfx.SetShadowMapUniformI("morphTargetTextureDimension", int(p.morphTargetTexture.tex.GetWidth()))
		} else {
			gfx.SetShadowMapUniformFv("morphTargetOffset", make([]float32, 4))
			gfx.SetShadowMapUniformI("numTargets", 0)
			gfx.SetShadowMapUniformFv("morphTargetWeight", make([]float32, 8))
		}
		mode := p.mode
		gfx.SetShadowMapUniformI("enableAlpha", int(Btoi(mat.alphaMode == AlphaModeBlend)))
		gfx.SetShadowMapUniformF("alphaThreshold", mat.alphaCutoff.getValue().(float32))
		gfx.SetShadowMapUniformFv("baseColorFactor", color[:])
		if n.skin != nil {
			gfx.SetShadowMapUniformI("numJoints", len(mdl.skins[*n.skin].joints))
		}
		if index := mat.textureIndex; index != nil {
			gfx.SetShadowMapTexture("tex", mdl.textures[*index].tex)
			gfx.SetShadowMapUniformI("useTexture", 1)
		} else {
			gfx.SetShadowMapUniformI("useTexture", 0)
		}
		gfx.SetShadowMapUniformMatrix3("texTransform", mat.textureTransform[:])
		for i := 0; i < numLights; i++ {
			culled := true
			if n.skin == nil && p.morphTargetCount == 0 {
				if lightTypes[i] == PointLight {
					for j := 0; j < 6; j++ {
						if !isCulled(viewProjMatrices[i*6+j].Mul4(n.worldTransform), p.boundingBox) {
							culled = false
							break
						}
					}
				} else {
					if isCulled(viewProjMatrices[i*6].Mul4(n.worldTransform), p.boundingBox) {
						continue
					} else {
						culled = false
					}
				}
			}
			if culled == false {
				gfx.SetShadowMapUniformI("layerOffset", i*6)
				gfx.SetShadowMapUniformI("lightIndex", i+lightIndex)
				gfx.RenderShadowMapElements(mode, int(p.numIndices), int(p.elementBufferOffset))
			}
		}
	}
}

func (model *Model) drawShadow(bufferIndex uint32, sceneNumber int, offset [3]float32) {
	scene := model.scenes[sceneNumber]
	gfx.prepareShadowMapPipeline(bufferIndex)
	numLights := 0
	var lightMatrices [32]mgl.Mat4
	var lightTypes [4]LightType
	for i := 0; i < 4; i++ {
		if i >= len(scene.lightNodes) {
			//gfx.SetShadowMapUniformI("lightType["+strconv.Itoa(i)+"]", 0) OpenGL error
			gfx.SetShadowMapUniformI("lights["+strconv.Itoa(i)+"].type", 0)
			continue
		}
		numLights += 1
		lightNode := model.nodes[scene.lightNodes[i]]
		light := model.lights[*lightNode.lightIndex]
		shadowMapNear := float32(0.1)
		if light.lightType == DirectionalLight {
			shadowMapNear = -20
		}
		shadowMapFar := float32(50)
		shadowMapBottom := float32(-20)
		shadowMapTop := float32(20)
		shadowMapLeft := float32(-20)
		shadowMapRight := float32(20)

		if light.lightType == DirectionalLight {
			shadowMapNear = -20
		}
		if v := light.shadowMapNear.getValue().(float32); v != 0 {
			shadowMapNear = v
		}
		if v := light.shadowMapFar.getValue().(float32); v != 0 {
			shadowMapFar = v
		}
		if v := light.shadowMapBottom.getValue().(float32); v != 0 {
			shadowMapBottom = v
		}
		if v := light.shadowMapTop.getValue().(float32); v != 0 {
			shadowMapTop = v
		}
		if v := light.shadowMapLeft.getValue().(float32); v != 0 {
			shadowMapLeft = v
		}
		if v := light.shadowMapRight.getValue().(float32); v != 0 {
			shadowMapRight = v
		}
		if v := lightNode.shadowMapNear.getValue().(float32); v != 0 {
			shadowMapNear = v
		}
		if v := lightNode.shadowMapFar.getValue().(float32); v != 0 {
			shadowMapFar = v
		}
		if v := lightNode.shadowMapBottom.getValue().(float32); v != 0 {
			shadowMapBottom = v
		}
		if v := lightNode.shadowMapTop.getValue().(float32); v != 0 {
			shadowMapTop = v
		}
		if v := lightNode.shadowMapLeft.getValue().(float32); v != 0 {
			shadowMapLeft = v
		}
		if v := lightNode.shadowMapRight.getValue().(float32); v != 0 {
			shadowMapRight = v
		}

		lightProj := gfx.PerspectiveProjectionMatrix(mgl.DegToRad(90), 1, shadowMapNear, shadowMapFar)
		if light.lightType == DirectionalLight {
			lightProj = gfx.OrthographicProjectionMatrix(shadowMapLeft, shadowMapRight, shadowMapBottom, shadowMapTop, shadowMapNear, shadowMapFar)
		} else if light.lightType == SpotLight {
			lightProj = gfx.PerspectiveProjectionMatrix(mgl.DegToRad(90), 1, shadowMapNear, shadowMapFar)
		}
		lightTypes[i] = light.lightType
		if light.lightType == PointLight {
			//var lightMatrices [6]mgl.Mat4
			lightMatrices[i*6] = lightProj.Mul4(mgl.LookAtV([3]float32{lightNode.worldTransform[12], lightNode.worldTransform[13], lightNode.worldTransform[14]}, [3]float32{lightNode.worldTransform[12] + 1, lightNode.worldTransform[13], lightNode.worldTransform[14]}, [3]float32{0, -1, 0}))
			lightMatrices[i*6+1] = lightProj.Mul4(mgl.LookAtV([3]float32{lightNode.worldTransform[12], lightNode.worldTransform[13], lightNode.worldTransform[14]}, [3]float32{lightNode.worldTransform[12] - 1, lightNode.worldTransform[13], lightNode.worldTransform[14]}, [3]float32{0, -1, 0}))
			lightMatrices[i*6+2] = lightProj.Mul4(mgl.LookAtV([3]float32{lightNode.worldTransform[12], lightNode.worldTransform[13], lightNode.worldTransform[14]}, [3]float32{lightNode.worldTransform[12], lightNode.worldTransform[13] + 1, lightNode.worldTransform[14]}, [3]float32{0, 0, 1}))
			lightMatrices[i*6+3] = lightProj.Mul4(mgl.LookAtV([3]float32{lightNode.worldTransform[12], lightNode.worldTransform[13], lightNode.worldTransform[14]}, [3]float32{lightNode.worldTransform[12], lightNode.worldTransform[13] - 1, lightNode.worldTransform[14]}, [3]float32{0, 0, -1}))
			lightMatrices[i*6+4] = lightProj.Mul4(mgl.LookAtV([3]float32{lightNode.worldTransform[12], lightNode.worldTransform[13], lightNode.worldTransform[14]}, [3]float32{lightNode.worldTransform[12], lightNode.worldTransform[13], lightNode.worldTransform[14] + 1}, [3]float32{0, -1, 0}))
			lightMatrices[i*6+5] = lightProj.Mul4(mgl.LookAtV([3]float32{lightNode.worldTransform[12], lightNode.worldTransform[13], lightNode.worldTransform[14]}, [3]float32{lightNode.worldTransform[12], lightNode.worldTransform[13], lightNode.worldTransform[14] - 1}, [3]float32{0, -1, 0}))
			for j := 0; j < 6; j++ {
				gfx.SetShadowMapUniformMatrix("lightMatrices["+strconv.Itoa(i*6+j)+"]", lightMatrices[i*6+j][:])
			}
			gfx.SetShadowMapUniformI("lights["+strconv.Itoa(i)+"].type", 2)

		} else {
			lightView := mgl.LookAtV([3]float32{lightNode.localTransform[12], lightNode.localTransform[13], lightNode.localTransform[14]}, [3]float32{lightNode.localTransform[12] + lightNode.lightDirection[0], lightNode.localTransform[13] + lightNode.lightDirection[1], lightNode.localTransform[14] + lightNode.lightDirection[2]}, [3]float32{0, 1, 0})
			lightMatrices[i*6] = lightProj.Mul4(lightView)
			gfx.SetShadowMapUniformMatrix("lightMatrices["+strconv.Itoa(i*6)+"]", lightMatrices[i*6][:])
			if light.lightType == DirectionalLight {
				gfx.SetShadowMapUniformI("lights["+strconv.Itoa(i)+"].type", 1)
			} else {
				gfx.SetShadowMapUniformI("lights["+strconv.Itoa(i)+"].type", 3)
			}
		}
		gfx.SetShadowMapUniformF("lights["+strconv.Itoa(i)+"].shadowMapFar", shadowMapFar)
		gfx.SetShadowMapUniformF("lights["+strconv.Itoa(i)+"].position", lightNode.worldTransform[12], lightNode.worldTransform[13], lightNode.worldTransform[14])
	}
	for _, index := range scene.nodes {
		drawNodeShadow(model, scene, model.nodes[index], offset, false, 0, numLights, lightMatrices[:], lightTypes[:])
	}
	for _, index := range scene.nodes {
		drawNodeShadow(model, scene, model.nodes[index], offset, true, 0, numLights, lightMatrices[:], lightTypes[:])
	}
	if len(model.scenes) > 1 {
		for _, index := range scene.nodes {
			drawNodeShadow(model, model.scenes[1], model.nodes[index], offset, false, 0, numLights, lightMatrices[:], lightTypes[:])
		}
		for _, index := range scene.nodes {
			drawNodeShadow(model, model.scenes[1], model.nodes[index], offset, true, 0, numLights, lightMatrices[:], lightTypes[:])
		}
	}
	gfx.ReleaseShadowPipeline()
}

func (model *Model) draw(bufferIndex uint32, sceneNumber int, layerNumber int, defaultLayerNumber int, offset [3]float32, proj, view, viewProjMatrix mgl.Mat4, outlineConst float32) {
	if sceneNumber < 0 || sceneNumber >= len(model.scenes) {
		return
	}
	scene := model.scenes[sceneNumber]
	for _, index := range scene.nodes {
		model.nodes[index].calculateWorldTransform(mgl.Ident4(), model.nodes)
		calculateAnimationData(model, model.nodes[index])
	}
	if len(scene.lightNodes) > 0 && layerNumber == -1 && gfx.IsShadowEnabled() {
		// Do it in another thread if possible
		if gfx.NewWorkerThread() {
			SafeGo(func() {
				model.drawShadow(bufferIndex, sceneNumber, offset)
			})
		} else {
			model.drawShadow(bufferIndex, sceneNumber, offset)
		}

	}
	if model.environment != nil {
		gfx.prepareModelPipeline(bufferIndex, model.environment)
	} else {
		gfx.prepareModelPipeline(bufferIndex, nil)
	}
	gfx.SetModelUniformMatrix("projection", proj[:])
	gfx.SetModelUniformMatrix("view", view[:])

	unlit := false
	for idx := 0; idx < 4; idx++ {
		gfx.SetModelUniformF("lights["+strconv.Itoa(idx)+"].color", 0, 0, 0)
	}
	if len(scene.lightNodes) > 0 {
		for idx := 0; idx < Min(len(scene.lightNodes), 4); idx++ {
			lightNode := model.nodes[scene.lightNodes[idx]]
			light := model.lights[*lightNode.lightIndex]
			shadowMapNear := float32(0.1)
			shadowMapFar := float32(50)
			shadowMapBottom := float32(-20)
			shadowMapTop := float32(20)
			shadowMapLeft := float32(-20)
			shadowMapRight := float32(20)
			shadowMapBias := float32(0.02)

			if light.lightType == DirectionalLight {
				shadowMapNear = -20
			}
			if v := light.shadowMapNear.getValue().(float32); v != 0 {
				shadowMapNear = v
			}
			if v := light.shadowMapFar.getValue().(float32); v != 0 {
				shadowMapFar = v
			}
			if v := light.shadowMapBottom.getValue().(float32); v != 0 {
				shadowMapBottom = v
			}
			if v := light.shadowMapTop.getValue().(float32); v != 0 {
				shadowMapTop = v
			}
			if v := light.shadowMapLeft.getValue().(float32); v != 0 {
				shadowMapLeft = v
			}
			if v := light.shadowMapRight.getValue().(float32); v != 0 {
				shadowMapRight = v
			}
			if v := light.shadowMapBias.getValue().(float32); v != 0 {
				shadowMapBias = v
			}
			if v := lightNode.shadowMapNear.getValue().(float32); v != 0 {
				shadowMapNear = v
			}
			if v := lightNode.shadowMapFar.getValue().(float32); v != 0 {
				shadowMapFar = v
			}
			if v := lightNode.shadowMapBottom.getValue().(float32); v != 0 {
				shadowMapBottom = v
			}
			if v := lightNode.shadowMapTop.getValue().(float32); v != 0 {
				shadowMapTop = v
			}
			if v := lightNode.shadowMapLeft.getValue().(float32); v != 0 {
				shadowMapLeft = v
			}
			if v := lightNode.shadowMapRight.getValue().(float32); v != 0 {
				shadowMapRight = v
			}
			if v := lightNode.shadowMapBias.getValue().(float32); v != 0 {
				shadowMapBias = v
			}
			gfx.SetModelUniformI("lights["+strconv.Itoa(idx)+"].type", int(light.lightType))
			gfx.SetModelUniformF("lights["+strconv.Itoa(idx)+"].intensity", light.intensity.getValue().(float32))
			gfx.SetModelUniformF("lights["+strconv.Itoa(idx)+"].innerConeCos", float32(math.Cos(float64(light.innerConeAngle.getValue().(float32)))))
			gfx.SetModelUniformF("lights["+strconv.Itoa(idx)+"].outerConeCos", float32(math.Cos(float64(light.outerConeAngle.getValue().(float32)))))
			gfx.SetModelUniformF("lights["+strconv.Itoa(idx)+"].range", light.lightRange.getValue().(float32))
			c := light.color.getValue().([3]float32)
			gfx.SetModelUniformF("lights["+strconv.Itoa(idx)+"].color", c[0], c[1], c[2])
			gfx.SetModelUniformF("lights["+strconv.Itoa(idx)+"].position", lightNode.worldTransform[12], lightNode.worldTransform[13], lightNode.worldTransform[14])
			gfx.SetModelUniformF("lights["+strconv.Itoa(idx)+"].shadowMapFar", shadowMapFar)
			gfx.SetModelUniformF("lights["+strconv.Itoa(idx)+"].shadowBias", shadowMapBias)
			if light.lightType != PointLight {
				gfx.SetModelUniformF("lights["+strconv.Itoa(idx)+"].direction", lightNode.lightDirection[0], lightNode.lightDirection[1], lightNode.lightDirection[2])
			}
			if light.lightType == DirectionalLight {
				lightProj := mgl.Ortho(shadowMapLeft, shadowMapRight, shadowMapBottom, shadowMapTop, shadowMapNear, shadowMapFar)
				lightView := mgl.LookAtV([3]float32{lightNode.localTransform[12], lightNode.localTransform[13], lightNode.localTransform[14]}, [3]float32{lightNode.localTransform[12] + lightNode.lightDirection[0], lightNode.localTransform[13] + lightNode.lightDirection[1], lightNode.localTransform[14] + lightNode.lightDirection[2]}, [3]float32{0, 1, 0})
				lightMatrix := lightProj.Mul4(lightView)
				gfx.SetModelUniformMatrix("lightMatrices["+strconv.Itoa(idx)+"]", lightMatrix[:])
			} else if light.lightType == SpotLight {
				lightProj := gfx.PerspectiveProjectionMatrix(mgl.DegToRad(90), 1, shadowMapNear, shadowMapFar)
				lightView := mgl.LookAtV([3]float32{lightNode.localTransform[12], lightNode.localTransform[13], lightNode.localTransform[14]}, [3]float32{lightNode.localTransform[12] + lightNode.lightDirection[0], lightNode.localTransform[13] + lightNode.lightDirection[1], lightNode.localTransform[14] + lightNode.lightDirection[2]}, [3]float32{0, 1, 0})
				lightMatrix := lightProj.Mul4(lightView)
				gfx.SetModelUniformMatrix("lightMatrices["+strconv.Itoa(idx)+"]", lightMatrix[:])
			} else {
				ident := mgl.Ident4()
				gfx.SetModelUniformMatrix("lightMatrices["+strconv.Itoa(idx)+"]", ident[:])
			}
		}
	} else if model.environment == nil {
		unlit = true
	}
	for _, index := range scene.nodes {
		drawNode(model, scene, layerNumber, defaultLayerNumber, model.nodes[index], offset, false, unlit, viewProjMatrix, outlineConst)
	}
	for _, index := range scene.nodes {
		drawNode(model, scene, layerNumber, defaultLayerNumber, model.nodes[index], offset, true, unlit, viewProjMatrix, outlineConst)
	}
	gfx.ReleaseModelPipeline()
}

func (s *Stage) drawModel(pos [2]float32, yofs float32, scl float32, layerNumber int32) {
	if s.model == nil || !gfx.IsModelEnabled() {
		return
	}
	drawFOV := s.stageCamera.fov * math.Pi / 180
	outlineConst := float32(0.003 * math.Tan(float64(drawFOV)))
	var syo float32
	scaleCorrection := float32(sys.cam.localcoord[1]) * sys.cam.localscl / float32(sys.gameHeight)
	posMul := float32(math.Tan(float64(drawFOV)/2)) * -s.model.offset[2] / (float32(sys.cam.localcoord[1]) / 2)
	aspectCorrection := (float32(sys.cam.zoffset)/float32(sys.cam.localcoord[1]) - (float32(sys.cam.zoffset)*s.localscl-sys.cam.aspectcorrection)/float32(sys.gameHeight)) * 2
	syo = -(float32(s.stageCamera.zoffset) - float32(sys.cam.localcoord[1])/2) * (1 - scl) / scl
	syo2 := -(float32(s.stageCamera.zoffset) - float32(sys.cam.localcoord[1])/2) * (1 - scaleCorrection) / float32(sys.cam.localcoord[1]) * 2
	offset := [3]float32{(pos[0]*-posMul + s.model.offset[0]/scl), (((pos[1])/scl+syo)*posMul + s.model.offset[1]), s.model.offset[2] / scl}
	rotation := [3]float32{s.model.rotation[0], s.model.rotation[1], s.model.rotation[2]}
	scale := [3]float32{s.model.scale[0], s.model.scale[1], s.model.scale[2]}
	proj := mgl.Translate3D(0, (sys.cam.zoomanchorcorrection+yofs)/float32(sys.gameHeight)*2+syo2+aspectCorrection, 0)

	// Apply aspect ratio scaling
	// TODO: In the letterbox case the model renders too low
	scaleX := scaleCorrection
	scaleY := scaleCorrection
	if sys.cfg.Video.FightAspectWidth != 0 && sys.cfg.Video.FightAspectHeight != 0 {
		aspectGame := sys.getCurrentAspect()
		aspectWindow := float32(sys.scrrect[2]) / float32(sys.scrrect[3])

		if aspectWindow > aspectGame {
			// Pillarbox - window is wider than game
			scaleX *= aspectWindow / aspectGame
		} else if aspectWindow < aspectGame {
			// Letterbox - window is taller than game
			scaleY *= aspectGame / aspectWindow
		}
	}

	proj = proj.Mul4(mgl.Scale3D(scaleX, scaleY, 1))
	proj = proj.Mul4(mgl.Translate3D(0, (sys.cam.yshift * scl), 0))
	proj = proj.Mul4(gfx.PerspectiveProjectionMatrix(drawFOV, float32(sys.scrrect[2])/float32(sys.scrrect[3]), s.stageCamera.near, s.stageCamera.far))
	view := mgl.Ident4()
	view = view.Mul4(mgl.Translate3D(offset[0], offset[1], offset[2]))
	view = view.Mul4(mgl.HomogRotate3DX(rotation[0]))
	view = view.Mul4(mgl.HomogRotate3DY(rotation[1]))
	view = view.Mul4(mgl.HomogRotate3DZ(rotation[2]))
	view = view.Mul4(mgl.Scale3D(scale[0], scale[1], scale[2]))

	if layerNumber == -1 {
		s.model.calculateTextureTransform()
		s.model.draw(0, 0, int(layerNumber), 0, [3]float32{offset[0] / scale[0], offset[1] / scale[1], offset[2] / scale[2]}, proj, view, proj.Mul4(view), outlineConst)
	} else if layerNumber == 0 {
		s.model.draw(0, 0, int(layerNumber), 0, [3]float32{offset[0] / scale[0], offset[1] / scale[1], offset[2] / scale[2]}, proj, view, proj.Mul4(view), outlineConst)
	} else if layerNumber == 1 {
		s.model.draw(0, 0, int(layerNumber), 0, [3]float32{offset[0] / scale[0], offset[1] / scale[1], offset[2] / scale[2]}, proj, view, proj.Mul4(view), outlineConst)
		s.model.draw(0, 1, int(layerNumber), 1, [3]float32{offset[0] / scale[0], offset[1] / scale[1], offset[2] / scale[2]}, proj, view, proj.Mul4(view), outlineConst)
	}
}

func (channel *GLTFAnimationChannel) parseAnimationPointer(m *Model, pointer string) error {
	channel.nodeIndex = nil
	components := strings.Split(pointer, "/")
	switch components[1] {
	case "materials":
		index, _ := strconv.Atoi(components[2])
		material := m.materials[index]
		switch components[3] {
		case "alphaCutoff":
			channel.target = &material.alphaCutoff
			channel.targetType = AnimFloat
		case "emissiveFactor":
			channel.target = &material.emission
			channel.targetType = AnimVec3
		case "normalTexture":
			switch components[4] {
			case "scale":
				return Error("invalid/unsupported JSON pointer: " + pointer)
			case "extensions":
				switch components[5] {
				case "KHR_texture_transform":
					switch components[6] {
					case "offset":
						channel.target = &material.normalMapOffset
						channel.targetType = AnimVec2
					case "scale":
						channel.target = &material.normalMapScale
						channel.targetType = AnimVec2
					case "rotation":
						channel.target = &material.normalMapRotation
						channel.targetType = AnimFloat
					default:
						return Error("invalid/unsupported JSON pointer: " + pointer)
					}
				default:
					return Error("invalid/unsupported JSON pointer: " + pointer)
				}
			default:
				return Error("invalid/unsupported JSON pointer: " + pointer)
			}
		case "emissiveTexture":
			switch components[4] {
			case "extensions":
				switch components[5] {
				case "KHR_texture_transform":
					switch components[6] {
					case "offset":
						channel.target = &material.emissionMapOffset
						channel.targetType = AnimVec2
					case "scale":
						channel.target = &material.emissionMapScale
						channel.targetType = AnimVec2
					case "rotation":
						channel.target = &material.emissionMapRotation
						channel.targetType = AnimFloat
					default:
						return Error("invalid/unsupported JSON pointer: " + pointer)
					}
				default:
					return Error("invalid/unsupported JSON pointer: " + pointer)
				}

			default:
				return Error("invalid/unsupported JSON pointer: " + pointer)
			}
		case "occlusionTexture":
			switch components[4] {
			case "strength":
				channel.target = &material.ambientOcclusion
				channel.targetType = AnimFloat
			case "extensions":
				switch components[5] {
				case "KHR_texture_transform":
					switch components[6] {
					case "offset":
						channel.target = &material.ambientOcclusionMapOffset
						channel.targetType = AnimVec2
					case "scale":
						channel.target = &material.ambientOcclusionMapScale
						channel.targetType = AnimVec2
					case "rotation":
						channel.target = &material.ambientOcclusionMapRotation
						channel.targetType = AnimFloat
					default:
						return Error("invalid/unsupported JSON pointer: " + pointer)
					}
				default:
					return Error("invalid/unsupported JSON pointer: " + pointer)
				}

			default:
				return Error("invalid/unsupported JSON pointer: " + pointer)
			}
		case "pbrMetallicRoughness":
			switch components[4] {
			case "baseColorFactor":
				channel.target = &material.baseColorFactor
				channel.targetType = AnimVec4
			case "metallicFactor":
				channel.target = &material.metallic
				channel.targetType = AnimFloat
			case "roughnessFactor":
				channel.target = &material.roughness
				channel.targetType = AnimFloat
			case "extensions":
				switch components[5] {
				case "KHR_texture_transform":
					switch components[6] {
					case "offset":
						channel.target = &material.metallicRoughnessMapOffset
						channel.targetType = AnimVec2
					case "scale":
						channel.target = &material.metallicRoughnessMapScale
						channel.targetType = AnimVec2
					case "rotation":
						channel.target = &material.metallicRoughnessMapRotation
						channel.targetType = AnimFloat
					default:
						return Error("invalid/unsupported JSON pointer: " + pointer)
					}
				default:
					return Error("invalid/unsupported JSON pointer: " + pointer)
				}
			default:
				return Error("invalid/unsupported JSON pointer: " + pointer)
			}
		default:
			return Error("invalid/unsupported JSON pointer: " + pointer)
		}
	case "meshes":
		index, _ := strconv.Atoi(components[2])
		mesh := m.meshes[index]
		switch components[3] {
		case "weights":
			channel.target = &mesh.morphTargetWeights
			if len(components) > 4 {
				elemIndex, _ := strconv.Atoi(components[4])
				channel.elemIndex = new(uint32)
				*channel.elemIndex = uint32(elemIndex)
				channel.targetType = AnimVecElem
			} else {
				channel.targetType = MorphTargetWeight
			}
		default:
			return Error("invalid/unsupported JSON pointer: " + pointer)
		}
	case "nodes":
		index, _ := strconv.Atoi(components[2])
		channel.nodeIndex = new(uint32)
		*channel.nodeIndex = uint32(index)
		node := m.nodes[index]
		switch components[3] {
		case "translation":
			channel.target = &node.transition
			channel.targetType = TRSTranslation
		case "rotation":
			channel.target = &node.rotation
			channel.targetType = TRSRotation
		case "scale":
			channel.target = &node.scale
			channel.targetType = TRSScale
		case "weights":
			channel.target = &node.morphTargetWeights
			if len(components) > 4 {
				elemIndex, _ := strconv.Atoi(components[4])
				channel.elemIndex = new(uint32)
				*channel.elemIndex = uint32(elemIndex)
				channel.targetType = AnimVecElem
			} else {
				channel.targetType = MorphTargetWeight
			}
		default:
			return Error("invalid/unsupported JSON pointer: " + pointer)
		}
	case "extensions":
		switch components[2] {
		case "KHR_lights_punctual":
			switch components[3] {
			case "lights":
				index, _ := strconv.Atoi(components[4])
				light := m.lights[index]
				switch components[5] {
				case "color":
					channel.target = &light.color
					channel.targetType = AnimVec3
				case "intensity":
					channel.target = &light.intensity
					channel.targetType = AnimFloat
				case "range":
					channel.target = &light.lightRange
					channel.targetType = AnimFloat
				case "spot":
					switch components[6] {
					case "innerConeAngle":
						channel.target = &light.innerConeAngle
						channel.targetType = AnimFloat
					case "outerConeAngle":
						channel.target = &light.outerConeAngle
						channel.targetType = AnimFloat
					default:
						return Error("invalid/unsupported JSON pointer: " + pointer)
					}
				default:
					return Error("invalid/unsupported JSON pointer: " + pointer)
				}
			default:
				return Error("invalid/unsupported JSON pointer: " + pointer)
			}

		default:
			return Error("invalid/unsupported JSON pointer: " + pointer)
		}
	default:
		return Error("invalid/unsupported JSON pointer: " + pointer)
	}
	return nil
}

func (model *Model) calculateAnimInterpolation(interpolation GLTFAnimationInterpolation, sampler *GLTFAnimationSampler, animTime float32, prevIndex int, length int) []float32 {
	if interpolation == InterpolationStep || prevIndex == -1 || len(model.animationTimeStamps[sampler.inputIndex]) == 1 {
		if prevIndex == -1 {
			prevIndex = 0
		}
		newVals := make([]float32, length)
		for i := 0; i < length; i++ {
			if interpolation == InterpolationCubicSpline {
				newVals[i] = sampler.output[prevIndex*3*length+i+length]
			} else {
				newVals[i] = sampler.output[prevIndex*length+i]
			}
		}
		return newVals
	}
	if interpolation == InterpolationLinear {
		rate := (animTime - model.animationTimeStamps[sampler.inputIndex][prevIndex]) / (model.animationTimeStamps[sampler.inputIndex][prevIndex+1] - model.animationTimeStamps[sampler.inputIndex][prevIndex])
		newVals := make([]float32, length)
		for i := 0; i < length; i++ {
			newVals[i] = sampler.output[prevIndex*length+i]*(1-rate) + sampler.output[(prevIndex+1)*length+i]*rate
		}
		return newVals
	} else {
		delta := (model.animationTimeStamps[sampler.inputIndex][prevIndex+1] - model.animationTimeStamps[sampler.inputIndex][prevIndex])
		rate := (animTime - model.animationTimeStamps[sampler.inputIndex][prevIndex]) / delta
		rateSquare := rate * rate
		rateCube := rateSquare * rate
		newVals := make([]float32, length)
		for i := 0; i < length; i++ {
			newVals[i] = (2*rateCube-3*rateSquare+1)*sampler.output[prevIndex*3*length+i+length] + delta*(rateCube-2*rateSquare+rate)*sampler.output[prevIndex*3*length+i+length*2] + (-2*rateCube+3*rateSquare)*sampler.output[(prevIndex+1)*3*length+i+length] + delta*(rateCube-rateSquare)*sampler.output[(prevIndex+1)*3*length+i]
		}
		return newVals
	}
}

func (model *Model) calculateAnimQuatInterpolation(interpolation GLTFAnimationInterpolation, sampler *GLTFAnimationSampler, animTime float32, prevIndex int) [4]float32 {
	if interpolation == InterpolationStep || prevIndex == -1 || len(model.animationTimeStamps[sampler.inputIndex]) == 1 {
		if prevIndex == -1 {
			prevIndex = 0
		}
		if interpolation == InterpolationCubicSpline {
			newVals := [4]float32{
				sampler.output[prevIndex*12+4],
				sampler.output[prevIndex*12+4+1],
				sampler.output[prevIndex*12+4+2],
				sampler.output[prevIndex*12+4+3],
			}
			return newVals
		} else {
			newVals := [4]float32{
				sampler.output[prevIndex*4],
				sampler.output[prevIndex*4+1],
				sampler.output[prevIndex*4+2],
				sampler.output[prevIndex*4+3],
			}
			return newVals
		}
	}
	if interpolation == InterpolationLinear {
		rate := (animTime - model.animationTimeStamps[sampler.inputIndex][prevIndex]) / (model.animationTimeStamps[sampler.inputIndex][prevIndex+1] - model.animationTimeStamps[sampler.inputIndex][prevIndex])
		q1 := mgl.Quat{sampler.output[prevIndex*4+3], mgl.Vec3{sampler.output[prevIndex*4], sampler.output[prevIndex*4+1], sampler.output[prevIndex*4+2]}}
		q2 := mgl.Quat{sampler.output[(prevIndex+1)*4+3], mgl.Vec3{sampler.output[(prevIndex+1)*4], sampler.output[(prevIndex+1)*4+1], sampler.output[(prevIndex+1)*4+2]}}
		dotProduct := q1.Dot(q2)
		if dotProduct < 0 {
			q1 = q1.Inverse()
		}
		q := mgl.QuatSlerp(q1, q2, rate)
		newVals := [4]float32{q.X(), q.Y(), q.Z(), q.W}
		return newVals
	} else {
		delta := (model.animationTimeStamps[sampler.inputIndex][prevIndex+1] - model.animationTimeStamps[sampler.inputIndex][prevIndex])
		rate := (animTime - model.animationTimeStamps[sampler.inputIndex][prevIndex]) / delta
		rateSquare := rate * rate
		rateCube := rateSquare * rate
		q := mgl.Quat{(2*rateCube-3*rateSquare+1)*sampler.output[prevIndex*12+3+4] + delta*(rateCube-2*rateSquare+rate)*sampler.output[prevIndex*12+3+8] + (-2*rateCube+3*rateSquare)*sampler.output[(prevIndex+1)*12+3+4] + delta*(rateCube-rateSquare)*sampler.output[(prevIndex+1)*12+3],
			mgl.Vec3{
				(2*rateCube-3*rateSquare+1)*sampler.output[prevIndex*12+4] + delta*(rateCube-2*rateSquare+rate)*sampler.output[prevIndex*12+8] + (-2*rateCube+3*rateSquare)*sampler.output[(prevIndex+1)*12+4] + delta*(rateCube-rateSquare)*sampler.output[(prevIndex+1)*12],
				(2*rateCube-3*rateSquare+1)*sampler.output[prevIndex*12+1+4] + delta*(rateCube-2*rateSquare+rate)*sampler.output[prevIndex*12+1+8] + (-2*rateCube+3*rateSquare)*sampler.output[(prevIndex+1)*12+1+4] + delta*(rateCube-rateSquare)*sampler.output[(prevIndex+1)*12+1],
				(2*rateCube-3*rateSquare+1)*sampler.output[prevIndex*12+2+4] + delta*(rateCube-2*rateSquare+rate)*sampler.output[prevIndex*12+2+8] + (-2*rateCube+3*rateSquare)*sampler.output[(prevIndex+1)*12+2+4] + delta*(rateCube-rateSquare)*sampler.output[(prevIndex+1)*12+2],
			}}.Normalize()
		newVals := [4]float32{q.X(), q.Y(), q.Z(), q.W}
		return newVals
	}
}

func (anim *GLTFAnimation) toggle(enabled bool) {
	if enabled {
		anim.enabled = enabled
		anim.time = 0
	} else if anim.enabled != enabled {
		anim.enabled = enabled
		for _, channel := range anim.channels {
			channel.target.rest()
		}
	}
}

func (model *Model) step(turbo float32) {
	for _, anim := range model.animations {
		if anim.enabled == false {
			continue
		}
		anim.time += turbo / 60
		for anim.time >= anim.duration && anim.duration > 0 && (anim.loopCount < 0 || anim.loop < anim.loopCount) {
			anim.time -= anim.duration
			anim.loop += 1
		}
		time := 60 * float64(anim.time)
		if math.Abs(time-math.Floor(time)) < 0.001 {
			anim.time = float32(math.Floor(time) / 60)
		} else if math.Abs(float64(time)-math.Ceil(float64(time))) < 0.001 {
			anim.time = float32(math.Ceil(time) / 60)
		}
		if anim.time >= anim.duration && anim.duration > 0 {
			anim.time = anim.duration
		}
		for _, channel := range anim.channels {
			sampler := anim.samplers[channel.samplerIndex]
			prevIndex := 0
			for i, t := range model.animationTimeStamps[sampler.inputIndex] {
				if anim.time > 0 && anim.time <= t {
					prevIndex = i - 1
					break
				}
			}
			switch channel.targetType {
			case TRSTranslation, TRSScale:
				node := model.nodes[*channel.nodeIndex]
				newVals := model.calculateAnimInterpolation(sampler.interpolation, sampler, anim.time, prevIndex, 3)
				oldVals := channel.target.getValue().([3]float32)
				if newVals[0] != oldVals[0] || newVals[1] != oldVals[1] || newVals[2] != oldVals[2] {
					node.transformChanged = true
				}
				channel.target.animate([3]float32{newVals[0], newVals[1], newVals[2]})
			case TRSRotation:
				node := model.nodes[*channel.nodeIndex]
				newVals := model.calculateAnimQuatInterpolation(sampler.interpolation, sampler, anim.time, prevIndex)
				oldVals := channel.target.getValue().([4]float32)
				if newVals[0] != oldVals[0] || newVals[1] != oldVals[1] || newVals[2] != oldVals[2] || newVals[3] != oldVals[3] {
					node.transformChanged = true
				}
				channel.target.animate(newVals)
			case MorphTargetWeight, AnimVec:
				newVals := model.calculateAnimInterpolation(sampler.interpolation, sampler, anim.time, prevIndex, len(channel.target.getValue().([]float32)))
				channel.target.animate(newVals)
			case AnimVec2:
				newVals := model.calculateAnimInterpolation(sampler.interpolation, sampler, anim.time, prevIndex, 2)
				channel.target.animate([2]float32{newVals[0], newVals[1]})
			case AnimVec3:
				newVals := model.calculateAnimInterpolation(sampler.interpolation, sampler, anim.time, prevIndex, 3)
				channel.target.animate([3]float32{newVals[0], newVals[1], newVals[2]})
			case AnimVec4:
				newVals := model.calculateAnimInterpolation(sampler.interpolation, sampler, anim.time, prevIndex, 4)
				channel.target.animate([4]float32{newVals[0], newVals[1], newVals[2], newVals[3]})
			case AnimVecElem:
				replaceVals := model.calculateAnimInterpolation(sampler.interpolation, sampler, anim.time, prevIndex, 1)
				newVals := channel.target.getValue().([]float32)
				newVals[*channel.elemIndex] = replaceVals[0]
				channel.target.animate(newVals)
			case AnimFloat:
				newVals := model.calculateAnimInterpolation(sampler.interpolation, sampler, anim.time, prevIndex, 1)
				channel.target.animate(newVals[0])
			case AnimQuat:
				newVals := model.calculateAnimQuatInterpolation(sampler.interpolation, sampler, anim.time, prevIndex)
				channel.target.animate(newVals)
			}
		}
	}
}

func (skin *Skin) calculateSkinMatrices(inverseGlobalTransform mgl.Mat4, nodes []*Node) {
	matrices := make([]float32, len(skin.joints)*12*2)
	for i, joint := range skin.joints {
		n := nodes[joint]
		reverseBindMatrix := skin.inverseBindMatrices[i*12 : (i+1)*12]
		matrix := mgl.Ident4()
		for j, v := range reverseBindMatrix {
			matrix[j] = v
		}
		matrix = n.worldTransform.Mul4(matrix.Transpose())
		matrix = inverseGlobalTransform.Mul4(matrix).Transpose()
		for j := 0; j < 12; j++ {
			matrices[i*24+j] = matrix[j]
		}
		normalMatrix := matrix.Transpose().Inv().Transpose()
		for j := 0; j < 12; j++ {
			matrices[i*24+12+j] = normalMatrix[j]
		}
	}
	skin.texture.tex.SetPixelData(matrices)
}

func (model *Model) reset() {
	for _, anim := range model.animations {
		anim.time = 0
		anim.enabled = anim.defaultEnabled
		anim.loop = 0
	}
	for _, node := range model.nodes {
		node.visible = true
	}
}
