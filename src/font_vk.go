//go:build !kinc && !android

package main

import (
	"container/list"
	"fmt"
	"image"
	"image/draw"
	"io"
	"io/ioutil"
	"os"
	"runtime"
	"unsafe"

	vk "github.com/Eiton/vulkan"
	"github.com/golang/freetype"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
)

type Font_VK struct {
	fontChar    map[rune]*character
	ttf         *truetype.Font
	scale       int32
	vao         uint32
	vbo         uint32
	program     uint32
	color       color
	resolution  [2]float32
	textures    []*TextureAtlas
	descriptors []*list.Element
}
type FontRenderer_VK struct {
	device          vk.Device
	program         *VulkanProgramInfo
	descriptorPools []vk.DescriptorPool
	freeDescriptors list.List
}

func (r *FontRenderer_VK) Init(rendererVk interface{}) {
	renderer := rendererVk.(*Renderer_VK)
	r.device = renderer.device
	program := &VulkanProgramInfo{}
	var err error
	VertShader, err := staticFiles.ReadFile("shaders/font.vert.spv")
	if err != nil {
		panic(err)
	}
	VertShader2 := make([]uint32, len(VertShader)/4)
	vk.Memcopy(unsafe.Pointer((*sliceHeader)(unsafe.Pointer(&VertShader2)).Data), VertShader)
	vertShader, err := renderer.CreateShader(renderer.device, VertShader2)
	if err != nil {
		panic(err)
	}
	defer vk.DestroyShaderModule(renderer.device, vertShader, nil)

	FragShader, err := staticFiles.ReadFile("shaders/font.frag.spv")
	if err != nil {
		panic(err)
	}
	FragShader2 := make([]uint32, len(FragShader)/4)
	vk.Memcopy(unsafe.Pointer((*sliceHeader)(unsafe.Pointer(&FragShader2)).Data), FragShader)
	fragShader, err := renderer.CreateShader(renderer.device, FragShader2)
	if err != nil {
		panic(err)
	}
	defer vk.DestroyShaderModule(renderer.device, fragShader, nil)
	shaderStages := []vk.PipelineShaderStageCreateInfo{
		{
			SType:  vk.StructureTypePipelineShaderStageCreateInfo,
			Stage:  vk.ShaderStageVertexBit,
			Module: vertShader,
			PName:  "main\x00",
		},
		{
			SType:  vk.StructureTypePipelineShaderStageCreateInfo,
			Stage:  vk.ShaderStageFragmentBit,
			Module: fragShader,
			PName:  "main\x00",
		},
	}
	pushConstantRanges := []vk.PushConstantRange{
		// resolution
		{
			StageFlags: vk.ShaderStageFlags(vk.ShaderStageVertexBit),
			Offset:     4 * 4,
			Size:       4 * 2,
		},
		// textColor
		{
			StageFlags: vk.ShaderStageFlags(vk.ShaderStageFragmentBit),
			Offset:     0,
			Size:       4 * 4,
		},
	}

	samplerLayoutBinding := []vk.DescriptorSetLayoutBinding{
		{
			Binding:            0,
			DescriptorCount:    1,
			DescriptorType:     vk.DescriptorTypeCombinedImageSampler,
			PImmutableSamplers: nil,
			StageFlags:         vk.ShaderStageFlags(vk.ShaderStageFragmentBit),
		},
	}
	layoutInfo := vk.DescriptorSetLayoutCreateInfo{
		SType:        vk.StructureTypeDescriptorSetLayoutCreateInfo,
		BindingCount: 1,
		PBindings:    samplerLayoutBinding,
	}
	var descriptorSetLayout vk.DescriptorSetLayout
	vk.CreateDescriptorSetLayout(renderer.device, &layoutInfo, nil, &descriptorSetLayout)
	program.descriptorSetLayouts = append(program.descriptorSetLayouts, descriptorSetLayout)

	pipelineLayoutCreateInfo := vk.PipelineLayoutCreateInfo{
		SType:                  vk.StructureTypePipelineLayoutCreateInfo,
		SetLayoutCount:         1,
		PSetLayouts:            []vk.DescriptorSetLayout{descriptorSetLayout},
		PushConstantRangeCount: 2,
		PPushConstantRanges:    pushConstantRanges,
	}
	var pipelineLayout vk.PipelineLayout
	vk.CreatePipelineLayout(renderer.device, &pipelineLayoutCreateInfo, nil, &pipelineLayout)
	program.pipelineLayout = pipelineLayout
	dynamicStates := []vk.DynamicState{
		vk.DynamicStateViewport,
		vk.DynamicStateScissor,
	}
	viewportState := vk.PipelineViewportStateCreateInfo{
		SType:         vk.StructureTypePipelineViewportStateCreateInfo,
		ViewportCount: 1,
		ScissorCount:  1,
	}
	dynamicState := vk.PipelineDynamicStateCreateInfo{
		SType:             vk.StructureTypePipelineDynamicStateCreateInfo,
		DynamicStateCount: uint32(len(dynamicStates)),
		PDynamicStates:    dynamicStates,
	}
	inputAssemblyState := vk.PipelineInputAssemblyStateCreateInfo{
		SType:                  vk.StructureTypePipelineInputAssemblyStateCreateInfo,
		Topology:               vk.PrimitiveTopologyTriangleList,
		PrimitiveRestartEnable: vk.False,
	}
	vertexInputBindings := []vk.VertexInputBindingDescription{{
		Binding:   0,
		Stride:    4 * 4, // 4 = sizeof(float32)
		InputRate: vk.VertexInputRateVertex,
	}}
	vertexInputAttributes := []vk.VertexInputAttributeDescription{{
		Binding:  0,
		Location: 0,
		Format:   vk.FormatR32g32Sfloat,
		Offset:   0,
	}, {
		Binding:  0,
		Location: 1,
		Format:   vk.FormatR32g32Sfloat,
		Offset:   8,
	}}
	vertexInputState := vk.PipelineVertexInputStateCreateInfo{
		SType:                           vk.StructureTypePipelineVertexInputStateCreateInfo,
		VertexBindingDescriptionCount:   uint32(len(vertexInputBindings)),
		PVertexBindingDescriptions:      vertexInputBindings,
		VertexAttributeDescriptionCount: uint32(len(vertexInputAttributes)),
		PVertexAttributeDescriptions:    vertexInputAttributes,
	}
	sampleMask := []vk.SampleMask{vk.SampleMask(vk.MaxUint32)}
	msaa := sys.msaa
	if msaa <= 0 {
		msaa = 1
	}
	multisampleState := vk.PipelineMultisampleStateCreateInfo{
		SType:                 vk.StructureTypePipelineMultisampleStateCreateInfo,
		RasterizationSamples:  vk.SampleCountFlagBits(msaa),
		SampleShadingEnable:   vk.False,
		PSampleMask:           sampleMask,
		MinSampleShading:      1,
		AlphaToCoverageEnable: 0,
		AlphaToOneEnable:      0,
	}
	rasterState := vk.PipelineRasterizationStateCreateInfo{
		SType:                   vk.StructureTypePipelineRasterizationStateCreateInfo,
		DepthClampEnable:        vk.False,
		RasterizerDiscardEnable: vk.False,
		PolygonMode:             vk.PolygonModeFill,
		CullMode:                vk.CullModeFlags(vk.CullModeNone),
		FrontFace:               vk.FrontFaceClockwise,
		DepthBiasEnable:         vk.False,
		DepthBiasConstantFactor: 0,
		DepthBiasClamp:          0,
		DepthBiasSlopeFactor:    0,
		LineWidth:               1,
	}
	depthStencilState := vk.PipelineDepthStencilStateCreateInfo{
		SType:             vk.StructureTypePipelineDepthStencilStateCreateInfo,
		DepthTestEnable:   vk.False,
		DepthWriteEnable:  vk.False,
		StencilTestEnable: vk.False,
	}

	attachmentStates := make([]vk.PipelineColorBlendAttachmentState, 0, 2)
	colorBlendStates := make([]vk.PipelineColorBlendStateCreateInfo, 0, 2)
	pipelineCreateInfos := make([]vk.GraphicsPipelineCreateInfo, 0, 2)
	pipelineIndexMap := map[interface{}]int{}
	attachmentStates = append(attachmentStates, vk.PipelineColorBlendAttachmentState{
		ColorWriteMask: vk.ColorComponentFlags(
			vk.ColorComponentRBit | vk.ColorComponentGBit |
				vk.ColorComponentBBit | vk.ColorComponentABit,
		),
		BlendEnable:         vk.True,
		SrcColorBlendFactor: vk.BlendFactorOne,
		DstColorBlendFactor: vk.BlendFactorZero,
		ColorBlendOp:        vk.BlendOpAdd,
		SrcAlphaBlendFactor: vk.BlendFactorOne,
		DstAlphaBlendFactor: vk.BlendFactorZero,
		AlphaBlendOp:        vk.BlendOpAdd,
	})
	colorBlendStates = append(colorBlendStates, vk.PipelineColorBlendStateCreateInfo{
		SType:           vk.StructureTypePipelineColorBlendStateCreateInfo,
		LogicOpEnable:   vk.False,
		LogicOp:         vk.LogicOpCopy,
		AttachmentCount: 1,
		PAttachments:    []vk.PipelineColorBlendAttachmentState{attachmentStates[0]},
	})
	pipelineCreateInfos = append(pipelineCreateInfos, vk.GraphicsPipelineCreateInfo{
		SType:               vk.StructureTypeGraphicsPipelineCreateInfo,
		StageCount:          2, // vert + frag
		PStages:             shaderStages,
		PVertexInputState:   &vertexInputState,
		PInputAssemblyState: &inputAssemblyState,
		PViewportState:      &viewportState,
		PRasterizationState: &rasterState,
		PMultisampleState:   &multisampleState,
		PColorBlendState:    &colorBlendStates[0],
		PDynamicState:       &dynamicState,
		Layout:              program.pipelineLayout,
		RenderPass:          renderer.mainRenderPass.renderPass,
		PDepthStencilState:  &depthStencilState,
	})
	attachmentStates = append(attachmentStates, vk.PipelineColorBlendAttachmentState{
		ColorWriteMask: vk.ColorComponentFlags(
			vk.ColorComponentRBit | vk.ColorComponentGBit |
				vk.ColorComponentBBit | vk.ColorComponentABit,
		),
		BlendEnable:         vk.True,
		SrcColorBlendFactor: vk.BlendFactorSrcAlpha,
		DstColorBlendFactor: vk.BlendFactorOneMinusSrcAlpha,
		ColorBlendOp:        vk.BlendOpAdd,
		SrcAlphaBlendFactor: vk.BlendFactorOne,
		DstAlphaBlendFactor: vk.BlendFactorZero,
		AlphaBlendOp:        vk.BlendOpAdd,
	})
	colorBlendStates = append(colorBlendStates, vk.PipelineColorBlendStateCreateInfo{
		SType:           vk.StructureTypePipelineColorBlendStateCreateInfo,
		LogicOpEnable:   vk.False,
		LogicOp:         vk.LogicOpCopy,
		AttachmentCount: 1,
		PAttachments:    []vk.PipelineColorBlendAttachmentState{attachmentStates[1]},
	})
	pipelineCreateInfos = append(pipelineCreateInfos, vk.GraphicsPipelineCreateInfo{
		SType:               vk.StructureTypeGraphicsPipelineCreateInfo,
		StageCount:          2, // vert + frag
		PStages:             shaderStages,
		PVertexInputState:   &vertexInputState,
		PInputAssemblyState: &inputAssemblyState,
		PViewportState:      &viewportState,
		PRasterizationState: &rasterState,
		PMultisampleState:   &multisampleState,
		PColorBlendState:    &colorBlendStates[1],
		PDynamicState:       &dynamicState,
		Layout:              program.pipelineLayout,
		RenderPass:          renderer.mainRenderPass.renderPass,
		PDepthStencilState:  &depthStencilState,
	})
	pipelines := make([]vk.Pipeline, len(pipelineCreateInfos))
	err = vk.Error(vk.CreateGraphicsPipelines(renderer.device,
		renderer.pipelineCache, uint32(len(pipelineCreateInfos)), pipelineCreateInfos, nil, pipelines))
	if err != nil {
		err = fmt.Errorf("vk.CreateGraphicsPipelines failed with %s", err)
		panic(err)
	}
	program.pipelines = pipelines
	program.pipelineIndexMap = pipelineIndexMap
	r.program = program

	poolSize := []vk.DescriptorPoolSize{
		{
			Type:            vk.DescriptorTypeCombinedImageSampler,
			DescriptorCount: 1000,
		},
	}
	poolInfo := vk.DescriptorPoolCreateInfo{
		SType:         vk.StructureTypeDescriptorPoolCreateInfo,
		PoolSizeCount: 1,
		PPoolSizes:    poolSize,
		MaxSets:       1000,
	}
	var descriptorPool vk.DescriptorPool
	err = vk.Error(vk.CreateDescriptorPool(renderer.device, &poolInfo, nil, &descriptorPool))
	if err != nil {
		err = fmt.Errorf("vk.CreateDescrriptorPool failed with %s", err)
		panic(err)
	}
	r.descriptorPools = append(r.descriptorPools, descriptorPool)

	descriptorSets := make([]vk.DescriptorSet, 1000)
	layouts := make([]vk.DescriptorSetLayout, 0, 1000)
	for i := 0; i < int(1000); i++ {
		layouts = append(layouts, descriptorSetLayout)
	}
	allocInfo := vk.DescriptorSetAllocateInfo{
		SType:              vk.StructureTypeDescriptorSetAllocateInfo,
		DescriptorPool:     descriptorPool,
		DescriptorSetCount: 1000,
		PSetLayouts:        layouts,
	}
	err = vk.Error(vk.AllocateDescriptorSets(renderer.device, &allocInfo, &descriptorSets[0]))
	if err != nil {
		err = fmt.Errorf("vk.AllocateDescriptorSets failed with %s", err)
		panic(err)
	}
	for i := range descriptorSets {
		r.freeDescriptors.PushBack(descriptorSets[i])
	}
}

func (r *FontRenderer_VK) LoadFont(file string, scale int32, windowWidth int, windowHeight int) (interface{}, error) {

	fd, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer fd.Close()
	f, err := r.LoadTrueTypeFont(fd, scale, 32, 127, LeftToRight)
	f.UpdateResolution(windowWidth, windowHeight)

	runtime.SetFinalizer(f, func(f *Font_VK) {
		for i := range f.descriptors {
			r.freeDescriptors.PushBack(f.descriptors[i].Value)
		}
	})
	return f, err
}

// GenerateGlyphs builds a set of textures based on a ttf files glyphs
func (f *Font_VK) GenerateGlyphs(low, high rune) error {
	//create a freetype context for drawing
	c := freetype.NewContext()
	c.SetDPI(72)
	c.SetFont(f.ttf)
	c.SetFontSize(float64(f.scale))
	c.SetHinting(font.HintingFull)

	//create new face to measure glyph dimensions
	ttfFace := truetype.NewFace(f.ttf, &truetype.Options{
		Size:    float64(f.scale),
		DPI:     72,
		Hinting: font.HintingFull,
	})

	// Add padding to prevent cropping
	// https://github.com/ikemen-engine/Ikemen-GO/issues/3085
	padding := 2

	//make each glyph
	for ch := low; ch <= high; ch++ {
		char := new(character)

		gBnd, gAdv, ok := ttfFace.GlyphBounds(ch)
		if ok != true {
			return fmt.Errorf("ttf face glyphBounds error")
		}

		gh := int32((gBnd.Max.Y - gBnd.Min.Y) >> 6)
		gw := int32((gBnd.Max.X - gBnd.Min.X) >> 6)

		//if glyph has no dimensions set to a max value
		if gw == 0 || gh == 0 {
			gBnd = f.ttf.Bounds(fixed.Int26_6(f.scale))
			gw = int32((gBnd.Max.X - gBnd.Min.X) >> 6)
			gh = int32((gBnd.Max.Y - gBnd.Min.Y) >> 6)

			//above can sometimes yield 0 for font smaller than 48pt, 1 is minimum
			if gw == 0 || gh == 0 {
				gw = 1
				gh = 1
			}
		}

		//The glyph's ascent and descent equal -bounds.Min.Y and +bounds.Max.Y.
		gAscent := int(-gBnd.Min.Y) >> 6
		gdescent := int(gBnd.Max.Y) >> 6

		//set w,h and adv, bearing V and bearing H in char
		//char.width = int(gw)
		//char.height = int(gh)
		char.width = int(gw) + (padding * 2)
		char.height = int(gh) + (padding * 2)
		char.advance = int(gAdv)
		char.bearingV = gdescent
		//char.bearingH = (int(gBnd.Min.X) >> 6)
		char.bearingH = (int(gBnd.Min.X) >> 6) - padding

		//create image to draw glyph
		fg, bg := image.White, image.Black
		//rect := image.Rect(0, 0, int(gw), int(gh))
		rect := image.Rect(0, 0, char.width, char.height)
		rgba := image.NewRGBA(rect)
		draw.Draw(rgba, rgba.Bounds(), bg, image.ZP, draw.Src)

		//set the glyph dot
		//px := 0 - (int(gBnd.Min.X) >> 6)
		//py := (gAscent)
		px := padding - (int(gBnd.Min.X) >> 6)
		py := padding + gAscent
		pt := freetype.Pt(px, py)

		// Draw the text from mask to image
		c.SetClip(rgba.Bounds())
		c.SetDst(rgba)
		c.SetSrc(fg)
		_, err := c.DrawString(string(ch), pt)
		if err != nil {
			return err
		}

		// Generate texture
		pix := make([]byte, len(rgba.Pix)/4)
		for i := range pix {
			pix[i] = rgba.Pix[i*4]
		}
		var uv [4]float32
		textureIndex := 0
		stride := int32(rgba.Rect.Dx()) // This was added to unify desktop and Android

		for uv, ok = f.textures[textureIndex].AddImage(int32(rgba.Rect.Dx()), int32(rgba.Rect.Dy()), stride, pix); !ok; uv, ok = f.textures[textureIndex].AddImage(int32(rgba.Rect.Dx()), int32(rgba.Rect.Dy()), stride, pix) {
			textureIndex += 1
			if textureIndex >= len(f.textures) {
				f.textures = append(f.textures, CreateTextureAtlas(256, 256, 8, true))
				descriptorSet := gfxFont.(*FontRenderer_VK).freeDescriptors.Front()
				gfxFont.(*FontRenderer_VK).freeDescriptors.Remove(descriptorSet)
				f.descriptors = append(f.descriptors, descriptorSet)
				imageInfo := []vk.DescriptorImageInfo{
					{
						ImageLayout: vk.ImageLayoutShaderReadOnlyOptimal,
						ImageView:   f.textures[textureIndex].texture.(*Texture_VK).imageView,
						Sampler:     gfx.(*Renderer_VK).spriteSamplers[1],
					},
				}

				descriptorWrites := []vk.WriteDescriptorSet{
					{
						SType:           vk.StructureTypeWriteDescriptorSet,
						DstSet:          descriptorSet.Value.(vk.DescriptorSet), //
						DstBinding:      0,
						DstArrayElement: 0,
						DescriptorCount: 1,
						DescriptorType:  vk.DescriptorTypeCombinedImageSampler,
						PImageInfo:      imageInfo,
					},
				}
				vk.UpdateDescriptorSets(gfxFont.(*FontRenderer_VK).device, uint32(len(descriptorWrites)), descriptorWrites, 0, nil)
			}
		}

		//texAtlas := f.textures[textureIndex]

		// This block is no longer necessary after the padding fix and actually introduces blur
		// aw := float32(texAtlas.width)
		// ah := float32(texAtlas.height)
		// off_u := 0.5 / aw
		// off_v := 0.5 / ah
		// uv[0] += off_u
		// uv[1] += off_v
		// uv[2] -= off_u
		// uv[3] -= off_v

		char.uv = uv
		char.textureID = uint32(textureIndex)

		//add char to fontChar list
		f.fontChar[ch] = char
	}

	return nil
}

func (r *FontRenderer_VK) LoadTrueTypeFont(reader io.Reader, scale int32, low, high rune, dir Direction) (Font, error) {
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	// Read the truetype font.
	ttf, err := truetype.Parse(data)
	if err != nil {
		return nil, err
	}

	//make Font stuct type
	f := new(Font_VK)
	f.fontChar = make(map[rune]*character)
	f.ttf = ttf
	f.scale = scale
	f.SetColor(1.0, 1.0, 1.0, 1.0) //set default white
	f.textures = append(f.textures, CreateTextureAtlas(256, 256, 8, true))
	descriptorSet := r.freeDescriptors.Front()
	r.freeDescriptors.Remove(descriptorSet)
	f.descriptors = append(f.descriptors, descriptorSet)
	imageInfo := []vk.DescriptorImageInfo{
		{
			ImageLayout: vk.ImageLayoutShaderReadOnlyOptimal,
			ImageView:   f.textures[0].texture.(*Texture_VK).imageView,
			Sampler:     gfx.(*Renderer_VK).spriteSamplers[1],
		},
	}

	descriptorWrites := []vk.WriteDescriptorSet{
		{
			SType:           vk.StructureTypeWriteDescriptorSet,
			DstSet:          descriptorSet.Value.(vk.DescriptorSet), //
			DstBinding:      0,
			DstArrayElement: 0,
			DescriptorCount: 1,
			DescriptorType:  vk.DescriptorTypeCombinedImageSampler,
			PImageInfo:      imageInfo,
		},
	}
	vk.UpdateDescriptorSets(r.device, uint32(len(descriptorWrites)), descriptorWrites, 0, nil)
	err = f.GenerateGlyphs(low, high)
	if err != nil {
		return nil, err
	}
	return f, nil
}
func (f *Font_VK) SetColor(red float32, green float32, blue float32, alpha float32) {
	f.color.r = red
	f.color.g = green
	f.color.b = blue
	f.color.a = alpha
	return
}
func (f *Font_VK) UpdateResolution(windowWidth int, windowHeight int) {
	f.resolution[0] = float32(windowWidth)
	f.resolution[1] = float32(windowHeight)
	return
}
func (f *Font_VK) Printf(x, y float32, scale float32, spacingXAdd float32, align int32, blend bool, window [4]int32, fs string, argv ...interface{}) error {
	r := gfx.(*Renderer_VK)
	switchedProgram := r.VKState.currentProgram != gfxFont.(*FontRenderer_VK).program
	r.VKState.currentProgram = gfxFont.(*FontRenderer_VK).program

	indices := []rune(fmt.Sprintf(fs, argv...))

	if len(indices) == 0 {
		return nil
	}

	// Buffer to store vertex data for multiple glyphs
	vertexData := make([]float32, 0, len(indices)*24)
	//setup blending mode
	pipelineIndex := 0
	if blend {
		pipelineIndex = 1
	}
	pipeline := gfxFont.(*FontRenderer_VK).program.pipelines[pipelineIndex]
	if switchedProgram || pipeline != r.VKState.currentPipeline {
		r.VKState.currentPipeline = pipeline
		vk.CmdBindPipeline(r.commandBuffers[0], vk.PipelineBindPointGraphics, pipeline)
	}
	if switchedProgram {
		bufferIndex := int(r.vertexBufferOffset) / int(r.vertexBuffers[0].size)
		if r.vertexBufferOffset > 0 && int(r.vertexBufferOffset)%int(r.vertexBuffers[0].size) == 0 {
			bufferIndex -= 1
		}
		vk.CmdBindVertexBuffers(r.commandBuffers[0], 0, 1, []vk.Buffer{r.vertexBuffers[bufferIndex].buffer}, []vk.DeviceSize{0})
	}
	//restrict drawing to a certain part of the window
	viewports := []vk.Viewport{{
		MinDepth: 0.0,
		MaxDepth: 1.0,
		X:        0,
		Y:        0,
		Width:    float32(sys.scrrect[2]),
		Height:   float32(sys.scrrect[3]),
	}}
	scissors := []vk.Rect2D{{
		Offset: vk.Offset2D{
			X: int32(MaxI(int(window[0]), 0)),
			//Y: int32(int(sys.scrrect[3]) - int(window[3]) - MaxI(int(window[1]), 0)),
			Y: int32(MaxI(int(window[1]), 0)),
		},
		Extent: vk.Extent2D{
			Width:  uint32(window[2]),
			Height: uint32(window[3]),
		},
	}}
	vk.CmdSetViewport(r.commandBuffers[0], 0, 1, viewports)
	vk.CmdSetScissor(r.commandBuffers[0], 0, 1, scissors)
	color := f.color
	resolution := f.resolution
	vk.CmdPushConstants(r.commandBuffers[0], gfxFont.(*FontRenderer_VK).program.pipelineLayout, vk.ShaderStageFlags(vk.ShaderStageFragmentBit), 0, 4*4, unsafe.Pointer(&color))
	vk.CmdPushConstants(r.commandBuffers[0], gfxFont.(*FontRenderer_VK).program.pipelineLayout, vk.ShaderStageFlags(vk.ShaderStageVertexBit), 4*4, 4*2, unsafe.Pointer(&resolution[0]))
	if align == 0 {
		x -= f.Width(scale, spacingXAdd, fs, argv...) * 0.5
	} else if align < 0 {
		x -= f.Width(scale, spacingXAdd, fs, argv...)
	}
	spacing := spacingXAdd * scale
	renderedAny := false
	firstVertex := uint32(r.vertexBufferOffset%r.vertexBuffers[0].size) / 16
	numVerticesToDraw := uint32(0)
	descriptorSetIndex := -1

	batchSize := (int(r.vertexBuffers[0].size) - (int(r.vertexBufferOffset) % int(r.vertexBuffers[0].size))) / 4
	batchSize = batchSize - (batchSize % 24)
	// Iterate through all characters in string
	for i := range indices {

		//get rune
		runeIndex := indices[i]

		//find rune in fontChar list
		ch, ok := f.fontChar[runeIndex]

		//load missing runes in batches of 32
		if !ok {
			low := runeIndex - (runeIndex % 32)
			f.GenerateGlyphs(low, low+31)
			ch, ok = f.fontChar[runeIndex]
		}

		//skip runes that are not in font character range
		if !ok {
			//fmt.Printf("%c %d\n", runeIndex, runeIndex)
			continue
		}

		if int(ch.textureID) != descriptorSetIndex {
			if numVerticesToDraw > 0 {
				vk.CmdDraw(r.commandBuffers[0], numVerticesToDraw, 1, firstVertex, 0)
				firstVertex += numVerticesToDraw
				numVerticesToDraw = 0
			}
			descriptorSetIndex = int(ch.textureID)
			descriptorSet := f.descriptors[descriptorSetIndex]
			vk.CmdBindDescriptorSets(r.commandBuffers[0], vk.PipelineBindPointGraphics, gfxFont.(*FontRenderer_VK).program.pipelineLayout, 0, 1, []vk.DescriptorSet{descriptorSet.Value.(vk.DescriptorSet)}, 0, nil)
		}

		if renderedAny {
			x += spacing
		}

		//calculate position and size for current rune
		xpos := x + float32(ch.bearingH)*scale
		ypos := y - float32(ch.height-ch.bearingV)*scale
		w := float32(ch.width) * scale
		h := float32(ch.height) * scale

		vertexData = append(vertexData,
			xpos+w, f.resolution[1]-(ypos), ch.uv[2], ch.uv[1],
			xpos, f.resolution[1]-(ypos), ch.uv[0], ch.uv[1],
			xpos, f.resolution[1]-(ypos+h), ch.uv[0], ch.uv[3],

			xpos, f.resolution[1]-(ypos+h), ch.uv[0], ch.uv[3],
			xpos+w, f.resolution[1]-(ypos+h), ch.uv[2], ch.uv[3],
			xpos+w, f.resolution[1]-(ypos), ch.uv[2], ch.uv[1],
		)

		numVerticesToDraw += 6

		// Now advance cursors for next glyph (note that advance is number of 1/64 pixels)
		x += float32((ch.advance >> 6)) * scale // Bitshift by 6 to get value in pixels (2^6 = 64 (divide amount of 1/64th pixels by 64 to get amount of pixels))
		renderedAny = true

		if len(vertexData) >= batchSize {
			gfx.SetVertexData(vertexData...)
			vertexData = vertexData[:0]
			vk.CmdDraw(r.commandBuffers[0], numVerticesToDraw, 1, firstVertex, 0)
			firstVertex = 0
			numVerticesToDraw = 0
			batchSize = int(r.vertexBuffers[0].size) / 4
			batchSize = batchSize - (batchSize % 24)
		}
	}
	if len(vertexData) > 0 {
		gfx.SetVertexData(vertexData...)
	}
	if numVerticesToDraw > 0 {
		vk.CmdDraw(r.commandBuffers[0], numVerticesToDraw, 1, firstVertex, 0)
	}
	return nil
}
func (f *Font_VK) Width(scale float32, spacingXAdd float32, fs string, argv ...interface{}) float32 {

	var width float32

	indices := []rune(fmt.Sprintf(fs, argv...))

	if len(indices) == 0 {
		return 0
	}

	spacing := spacingXAdd * scale
	renderedAny := false

	// Iterate through all characters in string
	for i := range indices {

		//get rune
		runeIndex := indices[i]

		//find rune in fontChar list
		ch, ok := f.fontChar[runeIndex]

		//load missing runes in batches of 32
		if !ok {
			low := runeIndex & rune(32-1)
			f.GenerateGlyphs(low, low+31)
			ch, ok = f.fontChar[runeIndex]
		}

		//skip runes that are not in font character range
		if !ok {
			//fmt.Printf("%c %d\n", runeIndex, runeIndex)
			continue
		}

		if renderedAny {
			width += spacing
		}

		// Now advance cursors for next glyph (note that advance is number of 1/64 pixels)
		width += float32((ch.advance >> 6)) * scale // Bitshift by 6 to get value in pixels (2^6 = 64 (divide amount of 1/64th pixels by 64 to get amount of pixels))
		renderedAny = true
	}
	return width
}
