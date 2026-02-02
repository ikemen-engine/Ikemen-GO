//go:build !kinc && !android

package main

import (
	"container/list"
	"embed" // Support for go:embed resources
	"fmt"
	"io"
	"log"
	"math"
	"os"
	"runtime"
	"strconv"
	"sync"
	"unsafe"

	vk "github.com/Eiton/vulkan"
	mgl "github.com/go-gl/mathgl/mgl32"
	"github.com/veandco/go-sdl2/sdl"
)

var vkDebug bool

//go:embed shaders/sprite.vert.spv
//go:embed shaders/sprite.frag.spv
//go:embed shaders/font.vert.spv
//go:embed shaders/font.frag.spv
//go:embed shaders/ident.vert.spv
//go:embed shaders/ident.frag.spv
//go:embed shaders/model.vert.spv
//go:embed shaders/model.frag.spv
//go:embed shaders/shadow.vert.spv
//go:embed shaders/shadow.frag.spv
//go:embed shaders/panoramaToCubeMap.frag.spv
//go:embed shaders/cubemapFiltering.frag.spv
var staticFiles embed.FS

// ------------------------------------------------------------------
// Texture_VK

type Texture_VK struct {
	width     int32
	height    int32
	depth     int32
	filter    bool
	mipLevels uint32
	offset    [2]int32
	uvst      [4]float32
	img       vk.Image
	imageView vk.ImageView
	sampler   vk.Sampler
}

func (r *Renderer_VK) newTexture(width, height, depth int32, filter bool) Texture {
	t := &Texture_VK{width, height, depth, filter, 1, [2]int32{0, 0}, [4]float32{0, 0, 1, 1}, nil, nil, nil}
	format := t.MapInternalFormat(Max(t.depth, 8))
	t.img = r.CreateImage(uint32(t.width), uint32(t.height), format, t.mipLevels, 1, vk.ImageUsageFlags(vk.ImageUsageTransferDstBit|vk.ImageUsageSampledBit), 1, vk.ImageTilingOptimal, false)
	imageMemory := r.AllocateImageMemory(t.img, vk.MemoryPropertyDeviceLocalBit)
	t.imageView = r.CreateImageView(t.img, format, 0, 1, 1, false)

	runtime.SetFinalizer(t, func(t *Texture_VK) {
		r.destroyResourceQueues[r.destroyResourceQueueIndex] <- VulkanResource{
			VulkanResourceTypeTexture,
			[]interface{}{
				t.img, t.imageView, imageMemory,
			},
		}
	})
	return t
}
func (r *Renderer_VK) newModelTexture(width, height, depth int32, filter bool) Texture {
	t := &Texture_VK{width, height, depth, filter, 1, [2]int32{0, 0}, [4]float32{0, 0, 1, 1}, nil, nil, nil}
	format := t.MapInternalFormat(Max(t.depth, 8))
	t.mipLevels = uint32(math.Floor(math.Log2(float64(MaxI(int(width), int(height))))) + 1)
	t.img = r.CreateImage(uint32(t.width), uint32(t.height), format, t.mipLevels, 1, vk.ImageUsageFlags(vk.ImageUsageTransferSrcBit|vk.ImageUsageTransferDstBit|vk.ImageUsageSampledBit), 1, vk.ImageTilingOptimal, false)
	imageMemory := r.AllocateImageMemory(t.img, vk.MemoryPropertyDeviceLocalBit)
	t.imageView = r.CreateImageView(t.img, format, 0, t.mipLevels, 1, false)

	runtime.SetFinalizer(t, func(t *Texture_VK) {
		r.destroyResourceQueues[r.destroyResourceQueueIndex] <- VulkanResource{
			VulkanResourceTypeTexture,
			[]interface{}{
				t.img, t.imageView, imageMemory,
			},
		}
	})
	return t
}
func (r *Renderer_VK) newDataTexture(width, height int32) Texture {
	t := &Texture_VK{width, height, 32 * 4, false, 1, [2]int32{0, 0}, [4]float32{0, 0, 1, 1}, nil, nil, nil}
	format := t.MapInternalFormat(Max(t.depth, 8))
	t.img = r.CreateImage(uint32(t.width), uint32(t.height), format, t.mipLevels, 1, vk.ImageUsageFlags(vk.ImageUsageColorAttachmentBit|vk.ImageUsageSampledBit|vk.ImageUsageTransferDstBit), 1, vk.ImageTilingOptimal, false)
	imageMemory := r.AllocateImageMemory(t.img, vk.MemoryPropertyDeviceLocalBit)
	t.imageView = r.CreateImageView(t.img, format, 0, 1, 1, false)
	t.sampler = r.GetSampler(VulkanSamplerInfo{TextureSamplingFilterNearest, TextureSamplingFilterNearest, TextureSamplingWrapClampToEdge, TextureSamplingWrapClampToEdge})

	runtime.SetFinalizer(t, func(t *Texture_VK) {
		r.destroyResourceQueues[r.destroyResourceQueueIndex] <- VulkanResource{
			VulkanResourceTypeTexture,
			[]interface{}{
				t.img, t.imageView, imageMemory,
			},
		}
	})
	return t
}
func (r *Renderer_VK) newHDRTexture(width, height int32) Texture {
	t := r.newTexture(width, height, 32*3, true) //float
	t.(*Texture_VK).sampler = r.GetSampler(VulkanSamplerInfo{TextureSamplingFilterLinear, TextureSamplingFilterLinear, TextureSamplingWrapMirroredRepeat, TextureSamplingWrapMirroredRepeat})
	return t
}
func (r *Renderer_VK) newCubeMapTexture(widthHeight int32, mipmap bool, lowestMipLevel int32) Texture {
	t := &Texture_VK{widthHeight, widthHeight, 96, false, 1, [2]int32{0, 0}, [4]float32{0, 0, 1, 1}, nil, nil, nil}
	if mipmap {
		t.mipLevels = uint32(math.Floor(math.Log2(float64(widthHeight)))+1) - uint32(lowestMipLevel)
		t.sampler = r.GetSampler(VulkanSamplerInfo{TextureSamplingFilterLinear, TextureSamplingFilterLinearMipMapLinear, TextureSamplingWrapClampToEdge, TextureSamplingWrapClampToEdge})
	} else {
		t.sampler = r.GetSampler(VulkanSamplerInfo{TextureSamplingFilterLinear, TextureSamplingFilterLinear, TextureSamplingWrapClampToEdge, TextureSamplingWrapClampToEdge})

	}
	format := t.MapInternalFormat(Max(t.depth, 8))
	t.img = r.CreateImage(uint32(t.width), uint32(t.height), format, t.mipLevels, 6, vk.ImageUsageFlags(vk.ImageUsageColorAttachmentBit|vk.ImageUsageTransferSrcBit|vk.ImageUsageTransferDstBit|vk.ImageUsageSampledBit), 1, vk.ImageTilingOptimal, true)
	imageMemory := r.AllocateImageMemory(t.img, vk.MemoryPropertyDeviceLocalBit)
	t.imageView = r.CreateImageView(t.img, format, 0, t.mipLevels, 6, true)

	runtime.SetFinalizer(t, func(t *Texture_VK) {
		r.destroyResourceQueues[r.destroyResourceQueueIndex] <- VulkanResource{
			VulkanResourceTypeTexture,
			[]interface{}{
				t.img, t.imageView, imageMemory,
			},
		}
	})
	return t
}

func (r *Renderer_VK) newPaletteTexture() Texture {
	t := &Texture_VK{256, 1, 32, false, 1, [2]int32{0, 0}, [4]float32{0, 0, 1, 1}, nil, nil, nil}
	if r.palTexture.emptySlot.Len() == 0 {
		r.addPalTexture()
	}
	slot := r.palTexture.emptySlot.Remove(r.palTexture.emptySlot.Front()).([2]uint32)
	t.offset = [2]int32{int32((slot[1] / r.palTexture.size) * 256), int32(slot[1] % r.palTexture.size)}
	t.img = r.palTexture.textures[slot[0]].img
	t.imageView = r.palTexture.textures[slot[0]].imageView
	t.uvst = [4]float32{
		(float32(t.offset[0]) + 0.5) / float32(r.palTexture.size),
		(float32(t.offset[1]) + 0.5) / float32(r.palTexture.size),
		float32(256) / float32(r.palTexture.size),
		float32(1) / float32(r.palTexture.size),
	}

	runtime.SetFinalizer(t, func(t *Texture_VK) {
		r.destroyResourceQueues[r.destroyResourceQueueIndex] <- VulkanResource{
			VulkanResourceTypePaletteTexture,
			[]interface{}{
				slot,
			},
		}
	})
	return t
}

func (r *Renderer_VK) newDummyCubeMapTexture() Texture {
	t := &Texture_VK{1, 1, 8, false, 1, [2]int32{0, 0}, [4]float32{0, 0, 1, 1}, nil, nil, nil}
	format := t.MapInternalFormat(Max(t.depth, 8))
	t.img = r.CreateImage(uint32(t.width), uint32(t.height), format, t.mipLevels, 6, vk.ImageUsageFlags(vk.ImageUsageTransferDstBit|vk.ImageUsageSampledBit), 1, vk.ImageTilingOptimal, true)
	imageMemory := r.AllocateImageMemory(t.img, vk.MemoryPropertyDeviceLocalBit)
	t.imageView = r.CreateImageView(t.img, format, 0, t.mipLevels, 6, true)

	runtime.SetFinalizer(t, func(t *Texture_VK) {
		r.destroyResourceQueues[r.destroyResourceQueueIndex] <- VulkanResource{
			VulkanResourceTypeTexture,
			[]interface{}{
				t.img, t.imageView, imageMemory,
			},
		}
	})
	return t
}

func (t *Texture_VK) SetData(textureData []byte) {
	if textureData == nil {
		return
	}
	t.SetSubData(textureData, t.offset[0], t.offset[1], t.width, t.height)
}
func (t *Texture_VK) SetSubData(textureData []byte, x, y, width, height int32) {
	size := uint32(width * height * t.depth / 8)
	bufferOffset := gfx.(*Renderer_VK).CopyToStagingBuffer(size, textureData)
	imageExtent := vk.Extent3D{
		Width:  uint32(width),
		Height: uint32(height),
		Depth:  1,
	}
	imageOffset := vk.Offset3D{
		X: x,
		Y: y,
		Z: 0,
	}
	regions, ok := gfx.(*Renderer_VK).stagingImageCopyRegions[t.img]
	if !ok {
		regions = make([]vk.BufferImageCopy, 0, 1)
		gfx.(*Renderer_VK).stagingImageBarriers[0] = append(gfx.(*Renderer_VK).stagingImageBarriers[0], vk.ImageMemoryBarrier{
			SType:               vk.StructureTypeImageMemoryBarrier,
			OldLayout:           vk.ImageLayoutUndefined,
			NewLayout:           vk.ImageLayoutTransferDstOptimal,
			SrcAccessMask:       vk.AccessFlags(vk.AccessNone),
			DstAccessMask:       vk.AccessFlags(vk.AccessTransferWriteBit),
			SrcQueueFamilyIndex: vk.QueueFamilyIgnored,
			DstQueueFamilyIndex: vk.QueueFamilyIgnored,
			Image:               t.img,
			SubresourceRange: vk.ImageSubresourceRange{
				AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
				BaseMipLevel:   0,
				LevelCount:     1,
				BaseArrayLayer: 0,
				LayerCount:     1,
			},
		})
		gfx.(*Renderer_VK).stagingImageBarriers[1] = append(gfx.(*Renderer_VK).stagingImageBarriers[1], vk.ImageMemoryBarrier{
			SType:               vk.StructureTypeImageMemoryBarrier,
			OldLayout:           vk.ImageLayoutTransferDstOptimal,
			NewLayout:           vk.ImageLayoutShaderReadOnlyOptimal,
			SrcAccessMask:       vk.AccessFlags(vk.AccessTransferWriteBit),
			DstAccessMask:       vk.AccessFlags(vk.AccessShaderReadBit),
			SrcQueueFamilyIndex: vk.QueueFamilyIgnored,
			DstQueueFamilyIndex: vk.QueueFamilyIgnored,
			Image:               t.img,
			SubresourceRange: vk.ImageSubresourceRange{
				AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
				BaseMipLevel:   0,
				LevelCount:     1,
				BaseArrayLayer: 0,
				LayerCount:     1,
			},
		})
	}
	gfx.(*Renderer_VK).stagingImageCopyRegions[t.img] = append(regions, vk.BufferImageCopy{
		BufferOffset:      bufferOffset,
		BufferRowLength:   0,
		BufferImageHeight: 0,
		ImageSubresource: vk.ImageSubresourceLayers{
			AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
			MipLevel:       0,
			BaseArrayLayer: 0,
			LayerCount:     1,
		},
		ImageOffset: imageOffset,
		ImageExtent: imageExtent,
	})
}
func (t *Texture_VK) SetSubDataStride(textureData []byte, x, y, width, height, stride int32) {

}

func (t *Texture_VK) SetDataG(textureData []byte, mag, min, ws, wt TextureSamplingParam) {
	size := uint32(t.width * t.height * t.depth / 8)
	t.sampler = gfx.(*Renderer_VK).GetSampler(VulkanSamplerInfo{mag, min, ws, wt})
	bufferOffset := gfx.(*Renderer_VK).CopyToStagingBuffer(size, textureData)
	imageExtent := vk.Extent3D{
		Width:  uint32(t.width),
		Height: uint32(t.height),
		Depth:  1,
	}
	regions, ok := gfx.(*Renderer_VK).stagingImageCopyRegions[t.img]
	if !ok {
		regions = make([]vk.BufferImageCopy, 0, 1)
		gfx.(*Renderer_VK).stagingImageBarriers[0] = append(gfx.(*Renderer_VK).stagingImageBarriers[0], vk.ImageMemoryBarrier{
			SType:               vk.StructureTypeImageMemoryBarrier,
			OldLayout:           vk.ImageLayoutUndefined,
			NewLayout:           vk.ImageLayoutTransferDstOptimal,
			SrcAccessMask:       vk.AccessFlags(vk.AccessNone),
			DstAccessMask:       vk.AccessFlags(vk.AccessTransferWriteBit),
			SrcQueueFamilyIndex: vk.QueueFamilyIgnored,
			DstQueueFamilyIndex: vk.QueueFamilyIgnored,
			Image:               t.img,
			SubresourceRange: vk.ImageSubresourceRange{
				AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
				BaseMipLevel:   0,
				LevelCount:     1,
				BaseArrayLayer: 0,
				LayerCount:     1,
			},
		})
		if t.mipLevels <= 1 {
			gfx.(*Renderer_VK).stagingImageBarriers[0] = append(gfx.(*Renderer_VK).stagingImageBarriers[0], vk.ImageMemoryBarrier{
				SType:               vk.StructureTypeImageMemoryBarrier,
				OldLayout:           vk.ImageLayoutTransferDstOptimal,
				NewLayout:           vk.ImageLayoutShaderReadOnlyOptimal,
				SrcAccessMask:       vk.AccessFlags(vk.AccessTransferWriteBit),
				DstAccessMask:       vk.AccessFlags(vk.AccessShaderReadBit),
				SrcQueueFamilyIndex: vk.QueueFamilyIgnored,
				DstQueueFamilyIndex: vk.QueueFamilyIgnored,
				Image:               t.img,
				SubresourceRange: vk.ImageSubresourceRange{
					AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
					BaseMipLevel:   0,
					LevelCount:     1,
					BaseArrayLayer: 0,
					LayerCount:     1,
				},
			})
		}
	}
	gfx.(*Renderer_VK).stagingImageCopyRegions[t.img] = append(regions, vk.BufferImageCopy{
		BufferOffset:      bufferOffset,
		BufferRowLength:   0,
		BufferImageHeight: 0,
		ImageSubresource: vk.ImageSubresourceLayers{
			AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
			MipLevel:       0,
			BaseArrayLayer: 0,
			LayerCount:     1,
		},
		ImageOffset: vk.Offset3D{
			X: 0,
			Y: 0,
			Z: 0,
		},
		ImageExtent: imageExtent,
	})
	if t.mipLevels > 1 {
		commandBuffer := gfx.(*Renderer_VK).BeginSingleTimeCommands()
		barriers := []vk.ImageMemoryBarrier{
			{
				SType:               vk.StructureTypeImageMemoryBarrier,
				OldLayout:           vk.ImageLayoutTransferDstOptimal,
				NewLayout:           vk.ImageLayoutTransferSrcOptimal,
				SrcAccessMask:       vk.AccessFlags(vk.AccessTransferWriteBit),
				DstAccessMask:       vk.AccessFlags(vk.AccessTransferReadBit),
				SrcQueueFamilyIndex: vk.QueueFamilyIgnored,
				DstQueueFamilyIndex: vk.QueueFamilyIgnored,
				Image:               t.img,
				SubresourceRange: vk.ImageSubresourceRange{
					AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
					BaseMipLevel:   0,
					LevelCount:     1,
					BaseArrayLayer: 0,
					LayerCount:     1,
				},
			},
			{
				SType:               vk.StructureTypeImageMemoryBarrier,
				OldLayout:           vk.ImageLayoutUndefined,
				NewLayout:           vk.ImageLayoutTransferDstOptimal,
				SrcAccessMask:       0,
				DstAccessMask:       vk.AccessFlags(vk.AccessTransferWriteBit),
				SrcQueueFamilyIndex: vk.QueueFamilyIgnored,
				DstQueueFamilyIndex: vk.QueueFamilyIgnored,
				Image:               t.img,
				SubresourceRange: vk.ImageSubresourceRange{
					AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
					BaseMipLevel:   1,
					LevelCount:     t.mipLevels - 1,
					BaseArrayLayer: 0,
					LayerCount:     1,
				},
			},
		}

		vk.CmdPipelineBarrier(commandBuffer, vk.PipelineStageFlags(vk.PipelineStageTransferBit), vk.PipelineStageFlags(vk.PipelineStageTransferBit), 0, 0, nil, 0, nil, uint32(len(barriers)), barriers)

		for i := uint32(1); i < t.mipLevels; i++ {
			imageBlits := []vk.ImageBlit{{
				SrcSubresource: vk.ImageSubresourceLayers{
					AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
					MipLevel:       i - 1,
					BaseArrayLayer: 0,
					LayerCount:     1,
				},
				SrcOffsets: [2]vk.Offset3D{
					{
						X: 0,
						Y: 0,
						Z: 0,
					},
					{
						X: (t.width >> (i - 1)),
						Y: (t.height >> (i - 1)),
						Z: 1,
					},
				},
				DstSubresource: vk.ImageSubresourceLayers{
					AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
					MipLevel:       i,
					BaseArrayLayer: 0,
					LayerCount:     1,
				},
				DstOffsets: [2]vk.Offset3D{
					{
						X: 0,
						Y: 0,
						Z: 0,
					},
					{
						X: (t.width >> i),
						Y: (t.height >> i),
						Z: 1,
					},
				},
			}}
			vk.CmdBlitImage(commandBuffer, t.img, vk.ImageLayoutTransferSrcOptimal, t.img, vk.ImageLayoutTransferDstOptimal, uint32(len(imageBlits)), imageBlits, vk.FilterLinear)
			barriers = []vk.ImageMemoryBarrier{
				{
					SType:               vk.StructureTypeImageMemoryBarrier,
					OldLayout:           vk.ImageLayoutTransferDstOptimal,
					NewLayout:           vk.ImageLayoutTransferSrcOptimal,
					SrcAccessMask:       vk.AccessFlags(vk.AccessTransferWriteBit),
					DstAccessMask:       vk.AccessFlags(vk.AccessTransferReadBit),
					SrcQueueFamilyIndex: vk.QueueFamilyIgnored,
					DstQueueFamilyIndex: vk.QueueFamilyIgnored,
					Image:               t.img,
					SubresourceRange: vk.ImageSubresourceRange{
						AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
						BaseMipLevel:   i,
						LevelCount:     1,
						BaseArrayLayer: 0,
						LayerCount:     1,
					},
				},
			}
			vk.CmdPipelineBarrier(commandBuffer, vk.PipelineStageFlags(vk.PipelineStageTransferBit), vk.PipelineStageFlags(vk.PipelineStageTransferBit), 0, 0, nil, 0, nil, uint32(len(barriers)), barriers)
		}

		barriers = []vk.ImageMemoryBarrier{
			{
				SType:               vk.StructureTypeImageMemoryBarrier,
				OldLayout:           vk.ImageLayoutTransferSrcOptimal,
				NewLayout:           vk.ImageLayoutShaderReadOnlyOptimal,
				SrcAccessMask:       vk.AccessFlags(vk.AccessTransferReadBit),
				DstAccessMask:       vk.AccessFlags(vk.AccessShaderReadBit),
				SrcQueueFamilyIndex: vk.QueueFamilyIgnored,
				DstQueueFamilyIndex: vk.QueueFamilyIgnored,
				Image:               t.img,
				SubresourceRange: vk.ImageSubresourceRange{
					AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
					BaseMipLevel:   0,
					LevelCount:     t.mipLevels,
					BaseArrayLayer: 0,
					LayerCount:     1,
				},
			},
		}
		vk.CmdPipelineBarrier(commandBuffer, vk.PipelineStageFlags(vk.PipelineStageTransferBit), vk.PipelineStageFlags(vk.PipelineStageFragmentShaderBit), 0, 0, nil, 0, nil, uint32(len(barriers)), barriers)
		vk.EndCommandBuffer(commandBuffer)
		gfx.(*Renderer_VK).tempCommands = append(gfx.(*Renderer_VK).tempCommands, commandBuffer)
	}
}
func (t *Texture_VK) SetPixelData(textureData []float32) {
	if t.depth == 96 {
		textureData = gfx.(*Renderer_VK).rgbToRGBA(textureData)
	}
	size := uint32(len(textureData) * 4)
	const m = 0x7fffffff
	bufferOffset := gfx.(*Renderer_VK).CopyToStagingBuffer(size, (*[m]byte)(unsafe.Pointer((*sliceHeader)(unsafe.Pointer(&textureData)).Data))[:size])

	imageExtent := vk.Extent3D{
		Width:  uint32(t.width),
		Height: uint32(t.height),
		Depth:  1,
	}
	regions, ok := gfx.(*Renderer_VK).stagingImageCopyRegions[t.img]
	if !ok {
		regions = make([]vk.BufferImageCopy, 0, 1)
		gfx.(*Renderer_VK).stagingImageBarriers[0] = append(gfx.(*Renderer_VK).stagingImageBarriers[0], vk.ImageMemoryBarrier{
			SType:               vk.StructureTypeImageMemoryBarrier,
			OldLayout:           vk.ImageLayoutUndefined,
			NewLayout:           vk.ImageLayoutTransferDstOptimal,
			SrcAccessMask:       vk.AccessFlags(vk.AccessNone),
			DstAccessMask:       vk.AccessFlags(vk.AccessTransferWriteBit),
			SrcQueueFamilyIndex: vk.QueueFamilyIgnored,
			DstQueueFamilyIndex: vk.QueueFamilyIgnored,
			Image:               t.img,
			SubresourceRange: vk.ImageSubresourceRange{
				AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
				BaseMipLevel:   0,
				LevelCount:     1,
				BaseArrayLayer: 0,
				LayerCount:     1,
			},
		})
		gfx.(*Renderer_VK).stagingImageBarriers[1] = append(gfx.(*Renderer_VK).stagingImageBarriers[1], vk.ImageMemoryBarrier{
			SType:               vk.StructureTypeImageMemoryBarrier,
			OldLayout:           vk.ImageLayoutTransferDstOptimal,
			NewLayout:           vk.ImageLayoutShaderReadOnlyOptimal,
			SrcAccessMask:       vk.AccessFlags(vk.AccessTransferWriteBit),
			DstAccessMask:       vk.AccessFlags(vk.AccessShaderReadBit),
			SrcQueueFamilyIndex: vk.QueueFamilyIgnored,
			DstQueueFamilyIndex: vk.QueueFamilyIgnored,
			Image:               t.img,
			SubresourceRange: vk.ImageSubresourceRange{
				AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
				BaseMipLevel:   0,
				LevelCount:     1,
				BaseArrayLayer: 0,
				LayerCount:     1,
			},
		})
	}
	gfx.(*Renderer_VK).stagingImageCopyRegions[t.img] = append(regions, vk.BufferImageCopy{
		BufferOffset:      bufferOffset,
		BufferRowLength:   0,
		BufferImageHeight: 0,
		ImageSubresource: vk.ImageSubresourceLayers{
			AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
			MipLevel:       0,
			BaseArrayLayer: 0,
			LayerCount:     1,
		},
		ImageOffset: vk.Offset3D{
			X: 0,
			Y: 0,
			Z: 0,
		},
		ImageExtent: imageExtent,
	})
}

func (t *Texture_VK) SetCubeMapData(textureData []byte) {
	size := uint32(t.width * t.height * t.depth / 8)
	bufferOffset := gfx.(*Renderer_VK).CopyToStagingBuffer(size, textureData)
	imageExtent := vk.Extent3D{
		Width:  uint32(t.width),
		Height: uint32(t.height),
		Depth:  1,
	}
	regions, ok := gfx.(*Renderer_VK).stagingImageCopyRegions[t.img]
	if !ok {
		regions = make([]vk.BufferImageCopy, 0, 6)
		gfx.(*Renderer_VK).stagingImageBarriers[0] = append(gfx.(*Renderer_VK).stagingImageBarriers[0], vk.ImageMemoryBarrier{
			SType:               vk.StructureTypeImageMemoryBarrier,
			OldLayout:           vk.ImageLayoutUndefined,
			NewLayout:           vk.ImageLayoutTransferDstOptimal,
			SrcAccessMask:       vk.AccessFlags(vk.AccessNone),
			DstAccessMask:       vk.AccessFlags(vk.AccessTransferWriteBit),
			SrcQueueFamilyIndex: vk.QueueFamilyIgnored,
			DstQueueFamilyIndex: vk.QueueFamilyIgnored,
			Image:               t.img,
			SubresourceRange: vk.ImageSubresourceRange{
				AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
				BaseMipLevel:   0,
				LevelCount:     1,
				BaseArrayLayer: 0,
				LayerCount:     6,
			},
		})
		gfx.(*Renderer_VK).stagingImageBarriers[1] = append(gfx.(*Renderer_VK).stagingImageBarriers[1], vk.ImageMemoryBarrier{
			SType:               vk.StructureTypeImageMemoryBarrier,
			OldLayout:           vk.ImageLayoutTransferDstOptimal,
			NewLayout:           vk.ImageLayoutShaderReadOnlyOptimal,
			SrcAccessMask:       vk.AccessFlags(vk.AccessTransferWriteBit),
			DstAccessMask:       vk.AccessFlags(vk.AccessShaderReadBit),
			SrcQueueFamilyIndex: vk.QueueFamilyIgnored,
			DstQueueFamilyIndex: vk.QueueFamilyIgnored,
			Image:               t.img,
			SubresourceRange: vk.ImageSubresourceRange{
				AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
				BaseMipLevel:   0,
				LevelCount:     1,
				BaseArrayLayer: 0,
				LayerCount:     6,
			},
		})
	}
	for i := uint32(0); i < 6; i++ {
		regions = append(regions, vk.BufferImageCopy{
			BufferOffset:      bufferOffset,
			BufferRowLength:   0,
			BufferImageHeight: 0,
			ImageSubresource: vk.ImageSubresourceLayers{
				AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
				MipLevel:       0,
				BaseArrayLayer: i,
				LayerCount:     1,
			},
			ImageOffset: vk.Offset3D{
				X: 0,
				Y: 0,
				Z: 0,
			},
			ImageExtent: imageExtent,
		})
	}
	gfx.(*Renderer_VK).stagingImageCopyRegions[t.img] = regions
}

func (t Texture_VK) CopyData(sourceTexture *Texture) {
	src := (*sourceTexture).(*Texture_VK)

	imageExtent := vk.Extent3D{
		Width:  uint32(src.width),
		Height: uint32(src.height),
		Depth:  1,
	}
	imageOffset := vk.Offset3D{
		X: 0,
		Y: 0,
		Z: 0,
	}
	commandBuffer := gfx.(*Renderer_VK).BeginSingleTimeCommands()
	barriers := []vk.ImageMemoryBarrier{
		{
			SType:               vk.StructureTypeImageMemoryBarrier,
			OldLayout:           vk.ImageLayoutUndefined,
			NewLayout:           vk.ImageLayoutTransferDstOptimal,
			SrcAccessMask:       vk.AccessFlags(vk.AccessNone),
			DstAccessMask:       vk.AccessFlags(vk.AccessTransferWriteBit),
			SrcQueueFamilyIndex: vk.QueueFamilyIgnored,
			DstQueueFamilyIndex: vk.QueueFamilyIgnored,
			Image:               t.img,
			SubresourceRange: vk.ImageSubresourceRange{
				AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
				BaseMipLevel:   0,
				LevelCount:     1,
				BaseArrayLayer: 0,
				LayerCount:     1,
			},
		},
		{
			SType:               vk.StructureTypeImageMemoryBarrier,
			OldLayout:           vk.ImageLayoutUndefined,
			NewLayout:           vk.ImageLayoutTransferSrcOptimal,
			SrcAccessMask:       vk.AccessFlags(vk.AccessNone),
			DstAccessMask:       vk.AccessFlags(vk.AccessTransferReadBit),
			SrcQueueFamilyIndex: vk.QueueFamilyIgnored,
			DstQueueFamilyIndex: vk.QueueFamilyIgnored,
			Image:               src.img,
			SubresourceRange: vk.ImageSubresourceRange{
				AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
				BaseMipLevel:   0,
				LevelCount:     1,
				BaseArrayLayer: 0,
				LayerCount:     1,
			},
		},
	}
	vk.CmdPipelineBarrier(commandBuffer, vk.PipelineStageFlags(vk.PipelineStageTopOfPipeBit), vk.PipelineStageFlags(vk.PipelineStageTransferBit), 0, 0, nil, 0, nil, 1, barriers)
	region := []vk.ImageCopy{{
		SrcSubresource: vk.ImageSubresourceLayers{
			AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
			MipLevel:       0,
			BaseArrayLayer: 0,
			LayerCount:     1,
		},
		DstSubresource: vk.ImageSubresourceLayers{
			AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
			MipLevel:       0,
			BaseArrayLayer: 0,
			LayerCount:     1,
		},
		SrcOffset: imageOffset,
		DstOffset: imageOffset,
		Extent:    imageExtent,
	}}
	vk.CmdCopyImage(commandBuffer, src.img, vk.ImageLayoutTransferSrcOptimal, t.img, vk.ImageLayoutTransferDstOptimal, 1, region)
	barriers = []vk.ImageMemoryBarrier{
		{
			SType:               vk.StructureTypeImageMemoryBarrier,
			OldLayout:           vk.ImageLayoutTransferDstOptimal,
			NewLayout:           vk.ImageLayoutShaderReadOnlyOptimal,
			SrcAccessMask:       vk.AccessFlags(vk.AccessTransferWriteBit),
			DstAccessMask:       vk.AccessFlags(vk.AccessShaderReadBit),
			SrcQueueFamilyIndex: vk.QueueFamilyIgnored,
			DstQueueFamilyIndex: vk.QueueFamilyIgnored,
			Image:               t.img,
			SubresourceRange: vk.ImageSubresourceRange{
				AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
				BaseMipLevel:   0,
				LevelCount:     1,
				BaseArrayLayer: 0,
				LayerCount:     1,
			},
		},
		{
			SType:               vk.StructureTypeImageMemoryBarrier,
			OldLayout:           vk.ImageLayoutTransferSrcOptimal,
			NewLayout:           vk.ImageLayoutShaderReadOnlyOptimal,
			SrcAccessMask:       vk.AccessFlags(vk.AccessTransferReadBit),
			DstAccessMask:       vk.AccessFlags(vk.AccessShaderReadBit),
			SrcQueueFamilyIndex: vk.QueueFamilyIgnored,
			DstQueueFamilyIndex: vk.QueueFamilyIgnored,
			Image:               src.img,
			SubresourceRange: vk.ImageSubresourceRange{
				AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
				BaseMipLevel:   0,
				LevelCount:     1,
				BaseArrayLayer: 0,
				LayerCount:     1,
			},
		},
	}
	vk.CmdPipelineBarrier(commandBuffer, vk.PipelineStageFlags(vk.PipelineStageTransferBit), vk.PipelineStageFlags(vk.PipelineStageFragmentShaderBit), 0, 0, nil, 0, nil, 1, barriers)
	gfx.(*Renderer_VK).EndSingleTimeCommands(commandBuffer)
}

// Return whether texture has a valid handle
func (t *Texture_VK) IsValid() bool {
	return true
}

func (t *Texture_VK) GetWidth() int32 {
	return t.width
}

func (t *Texture_VK) GetHeight() int32 {
	return t.height
}

func (t *Texture_VK) MapInternalFormat(i int32) vk.Format {
	var InternalFormatLUT = map[int32]vk.Format{
		8:  vk.FormatR8Unorm,
		24: vk.FormatR8g8b8Unorm,
		32: vk.FormatR8g8b8a8Unorm,
		// Seems like FormatR32g32b32Sfloat is not widely supported, just use FormatR32g32b32a32Sfloat
		96:  vk.FormatR32g32b32a32Sfloat,
		128: vk.FormatR32g32b32a32Sfloat,
	}
	return InternalFormatLUT[i]
}

// ------------------------------------------------------------------
// Renderer_VK

type Renderer_VK struct {
	destroyResourceQueues     [2]chan VulkanResource
	destroyResourceQueueIndex uint32
	enableModel               bool
	enableShadow              bool
	renderShadowMap           bool
	waitGroup                 sync.WaitGroup

	gpuDevices []vk.PhysicalDevice
	gpuIndex   uint32

	setVSync                 bool
	instance                 vk.Instance
	surface                  vk.Surface
	device                   vk.Device
	queue                    vk.Queue
	descriptorPoolSize       uint32
	descriptorPool           vk.DescriptorPool
	commandPools             []vk.CommandPool
	commandBuffers           []vk.CommandBuffer
	fences                   []vk.Fence
	semaphores               []vk.Semaphore
	swapchains               []*VulkanSwapchainInfo
	mainRenderPass           *VulkanRenderInfo
	swapchainRenderPass      *VulkanRenderInfo
	postProcessingRenderPass *VulkanRenderInfo
	spriteSamplers           []vk.Sampler
	samplers                 map[VulkanSamplerInfo]vk.Sampler
	vertexBufferOffset       uintptr
	vertexBuffers            []VulkanBuffer
	stagingBuffers           [2]VulkanBuffer
	modelVertexBuffers       []VulkanBuffer
	modelIndexBuffers        []VulkanBuffer
	stagingBufferFences      [2]bool
	stagingBufferIndex       uint32
	stagingBufferOffset      uint32
	stagingImageBarriers     [2][]vk.ImageMemoryBarrier
	stagingImageCopyRegions  map[vk.Image][]vk.BufferImageCopy
	tempCommands             []vk.CommandBuffer
	usedCommands             []vk.CommandBuffer

	pipelineCache            vk.PipelineCache
	spriteProgram            *VulkanProgramInfo
	modelProgram             *VulkanProgramInfo
	shadowMapProgram         *VulkanProgramInfo
	postProcessingProgram    *VulkanProgramInfo
	panoramaToCubeMapProgram *VulkanProgramInfo
	cubemapFilteringProgram  *VulkanProgramInfo
	lutProgram               *VulkanProgramInfo
	mainRenderTarget         *VulkanRenderTargetInfo
	renderTargets            [2]*VulkanRenderTargetInfo
	palTexture               VulkanPalTexture
	shadowMapTextures        *Texture_VK
	dummyTexture             *Texture_VK
	dummyCubeTexture         *Texture_VK

	fixedSizeTextures []*Texture_VK

	maxAnisotropy                   float32
	minUniformBufferOffsetAlignment uint32
	maxImageArrayLayers             uint32
	memoryTypeMap                   map[vk.MemoryPropertyFlagBits]uint32
	allocatedImageMemory            uint64

	VKState
}

type VKState struct {
	currentProgram                *VulkanProgramInfo
	currentPipeline               vk.Pipeline
	currentShadowmMapPipeline     vk.Pipeline
	currentSpriteTexture          VulkanSpriteTexture
	currentModelTexture           VulkanModelTexture
	currentShadowMapTexture       VulkanModelTexture
	scissor                       vk.Rect2D
	spriteVertUniformBufferOffset uint32
	spriteFragUniformBufferOffset uint32
	modelVertAttrOffset           uint32
	modelVertexBufferIndex        uint32
	modelUniformBufferOffset1     uint32
	modelUniformBufferOffset2     uint32
	numShadowMapLayers            uint32
	shadowMapVertAttrOffset       uint32
	shadowMapUniformBufferOffset1 uint32
	shadowMapUniformBufferOffset2 uint32
	VulkanSpriteTexture
	VulkanModelTexture
	VulkanShadowMapTexture
	VulkanPipelineState
	VulkanShadowMapPipelineState
	VulkanSpriteProgramVertUniformBufferObject
	VulkanSpriteProgramFragUniformBufferObject
	VulkanModelProgramUniformBufferObject0
	VulkanModelProgramUniformBufferObject1
	VulkanModelProgramUniformBufferObject2
	VulkanShadowMapProgramUniformBufferObject0
	VulkanShadowMapProgramUniformBufferObject1
	VulkanShadowMapProgramUniformBufferObject2
	VulkanShadowMapProgramPushConstant
}

type VulkanResourceType int

const (
	VulkanResourceTypeTexture VulkanResourceType = iota
	VulkanResourceTypePaletteTexture
	VulkanResourceTypeBuffer
	VulkanResourceTypeFence
)

type VulkanResource struct {
	resourceType VulkanResourceType
	resources    []interface{}
}

type VulkanPalTexture struct {
	textures  []*Texture_VK
	size      uint32
	emptySlot *list.List
}

type VulkanSpriteTexture struct {
	spriteTexture *Texture_VK
	palTexture    *Texture_VK
}
type VulkanModelTexture struct {
	lambertianEnvSampler *Texture_VK
	GGXEnvSampler        *Texture_VK
	GGXLUT               *Texture_VK
	jointMatricesTexture *Texture_VK
	morphTargetTexture   *Texture_VK
	tex                  *Texture_VK
	normalMap            *Texture_VK
	metallicRoughnessMap *Texture_VK
	ambientOcclusionMap  *Texture_VK
	emissionMap          *Texture_VK
}
type VulkanShadowMapTexture struct {
	jointMatricesTexture *Texture_VK
	morphTargetTexture   *Texture_VK
	tex                  *Texture_VK
}

// Set 0
type VulkanModelProgramUniformBufferObject0 struct {
	viewMatrix           [16]float32
	projectionMatrix     [16]float32
	lightMatrices        [4][16]float32
	lights               [4]VulkanLightUniform
	environmentRotation  [12]float32 //[9]float32
	cameraPosition       [3]float32
	environmentIntensity float32
	mipCount             int32
}

// Set 1
type VulkanModelProgramUniformBufferObject1 struct {
	texTransform                  [12]float32 //[9]float32
	normalMapTransform            [12]float32 //[9]float32
	metallicRoughnessMapTransform [12]float32 //[9]float32
	ambientOcclusionMapTransform  [12]float32 //[9]float32
	emissionMapTransform          [12]float32 //[9]float32
	baseColorFactor               [4]float32
	emission                      [3]float32
	_padding                      uint32
	metallicRoughness             [2]float32
	ambientOcclusionStrength      float32
	alphaThreshold                float32
	unlit                         bool
	_padding2                     [3]bool
	enableAlpha                   bool
}

// Set 2
type VulkanModelProgramUniformBufferObject2 struct {
	modelMatrix                 [16]float32
	normalMatrix                [16]float32
	numJoints                   int32
	numTargets                  int32
	morphTargetTextureDimension int32
	numVertices                 uint32
	morphTargetWeight           [2][4]float32
	morphTargetOffset           [4]float32
	meshOutline                 float32
	modelGray                   float32
	modelHue                    float32
	_padding                    uint32
	modelAdd                    [3]float32
	_padding2                   uint32
	modelMult                   [3]float32
	_padding3                   float32
	neg                         uint32
}
type VulkanShadowMapProgramUniformBufferObject0 struct {
	lightMatrices [24][16]float32
	lights        [4]VulkanLightUniform
	layers        [24]float32
}
type VulkanShadowMapProgramUniformBufferObject1 struct {
	texTransform    [12]float32 //[9]float32
	baseColorFactor [4]float32
	alphaThreshold  float32
	enableAlpha     bool
}
type VulkanShadowMapProgramUniformBufferObject2 struct {
	numJoints                   int32
	numTargets                  int32
	morphTargetTextureDimension int32
	morphTargetWeight           [2][4]float32
	morphTargetOffset           [4]float32
}

type VulkanShadowMapProgramPushConstant struct {
	model       [16]float32
	numVertices int32
	lightIndex  int32
}

type VulkanPipelineState struct {
	VulkanModelPipelineState
	VulkanBlendState
}

type VulkanBlendState struct {
	op  BlendEquation
	src BlendFunc
	dst BlendFunc
}

type VulkanModelPipelineState struct {
	depthTest       bool
	depthMask       bool
	invertFrontFace bool
	doubleSided     bool
	useUV           bool
	primitiveMode   PrimitiveMode
	VulkanModelSpecializationConstants0
	VulkanModelSpecializationConstants1
}

type VulkanShadowMapPipelineState struct {
	invertFrontFace bool
	doubleSided     bool
	useUV           bool
	primitiveMode   PrimitiveMode
	VulkanModelSpecializationConstants0
	VulkanModelSpecializationConstants1
}
type VulkanModelSpecializationConstants0 struct {
	useJoint0           bool
	_padding0           [3]bool
	useJoint1           bool
	_padding1           [3]bool
	useNormal           bool
	_padding2           [3]bool
	useTangent          bool
	_padding3           [3]bool
	useVertColor        bool
	_padding4           [3]bool
	useOutlineAttribute bool
	_padding5           [3]bool
}
type VulkanModelSpecializationConstants1 struct {
	useTexture              bool
	_padding0               [3]bool
	useNormalMap            bool
	_padding1               [3]bool
	useMetallicRoughnessMap bool
	_padding2               [3]bool
	useEmissionMap          bool
	_padding3               [3]bool
	useShadowMap            bool
	_padding4               [3]bool
}

const VulkanVertUniformSize = int(unsafe.Sizeof(VulkanSpriteProgramVertUniformBufferObject{}))
const VulkanFragUniformSize = int(unsafe.Sizeof(VulkanSpriteProgramFragUniformBufferObject{}))
const VulkanModelUniform0Size = int(unsafe.Sizeof(VulkanModelProgramUniformBufferObject0{}))
const VulkanModelUniform1Size = int(unsafe.Sizeof(VulkanModelProgramUniformBufferObject1{}))
const VulkanModelUniform2Size = int(unsafe.Sizeof(VulkanModelProgramUniformBufferObject2{}))
const VulkanShadowMapUniform0Size = int(unsafe.Sizeof(VulkanShadowMapProgramUniformBufferObject0{}))
const VulkanShadowMapUniform1Size = int(unsafe.Sizeof(VulkanShadowMapProgramUniformBufferObject1{}))
const VulkanShadowMapUniform2Size = int(unsafe.Sizeof(VulkanShadowMapProgramUniformBufferObject2{}))

type VulkanBuffer struct {
	buffer       vk.Buffer
	bufferMemory vk.DeviceMemory
	data         unsafe.Pointer
	size         uintptr
}

type VulkanSwapchainInfo struct {
	swapchain         vk.Swapchain
	currentImageIndex uint32
	imageCount        uint32
	images            []vk.Image
	imageViews        []vk.ImageView
	framebuffers      []vk.Framebuffer

	extent vk.Extent2D
	format vk.Format
}

type VulkanProgramInfo struct {
	pipelineLayout       vk.PipelineLayout
	descriptorSetLayouts []vk.DescriptorSetLayout
	pipelines            []vk.Pipeline
	pipelineIndexMap     map[interface{}]int
	uniformBuffers       []VulkanBuffer
	uniformBufferOffset  uint32
	uniformSize          uint32
	uniformOffsetMap     map[interface{}]uint32
	vertShader           vk.ShaderModule
	fragShader           vk.ShaderModule
}

type VulkanSamplerInfo struct {
	magFilter TextureSamplingParam
	minFilter TextureSamplingParam
	wrapS     TextureSamplingParam
	wrapT     TextureSamplingParam
}

type VulkanRenderInfo struct {
	renderPass vk.RenderPass
}

type VulkanRenderTargetInfo struct {
	texture       *Texture_VK
	depthTexture  *Texture_VK
	framebuffer   vk.Framebuffer
	descriptorSet vk.DescriptorSet
}

type VulkanSpriteProgramVertUniformBufferObject struct {
	modelView  [16]float32
	projection [16]float32
}
type VulkanSpriteProgramFragUniformBufferObject struct {
	x1x2x4x3                      [4]float32
	tint                          [4]float32
	add                           [3]float32
	_padding                      uint32
	mult                          [3]float32
	alpha, gray, hue              float32
	mask                          int32
	isFlat, isRgba, isTrapez, neg uint32
}

type VulkanLightUniform struct {
	direction  [3]float32
	lightRange float32

	color     [3]float32
	intensity float32

	position     [3]float32
	innerConeCos float32

	outerConeCos float32
	lightType    int32

	shadowBias   float32
	shadowMapFar float32
}

var appInfo = &vk.ApplicationInfo{
	SType:              vk.StructureTypeApplicationInfo,
	ApiVersion:         vk.MakeVersion(1, 3, 0),
	ApplicationVersion: vk.MakeVersion(1, 0, 0),
	PApplicationName:   "Ikemen GO\x00",
	PEngineName:        "Ikemen GO\x00",
	EngineVersion:      vk.MakeVersion(0, 99, 0),
}

var vk_validationLayers = []string{
	"VK_LAYER_KHRONOS_validation\x00",
}

func (r *Renderer_VK) GetName() string {
	return "Vulkan 1.3.239"
}

func (r *Renderer_VK) NewVulkanDevice(appInfo *vk.ApplicationInfo, window uintptr) error {
	// create a Vulkan instance.
	instanceExtensions := sys.window.Window.VulkanGetInstanceExtensions()
	for i := range instanceExtensions {
		instanceExtensions[i] = fmt.Sprintf("%s\x00", instanceExtensions[i])
	}
	instanceCreateInfo := &vk.InstanceCreateInfo{
		SType:                   vk.StructureTypeInstanceCreateInfo,
		PApplicationInfo:        appInfo,
		EnabledExtensionCount:   uint32(len(instanceExtensions)),
		PpEnabledExtensionNames: instanceExtensions,
		EnabledLayerCount:       0,
		PpEnabledLayerNames:     []string{},
	}
	// This causes a crash with it in SDL and without it in GLFW on macOS. Have no idea why.
	// if runtime.GOOS == "darwin" {
	// 	instanceExtensions = append(instanceExtensions, vk.KhrPortabilityEnumerationExtensionName+"\x00")
	// 	instanceCreateInfo.PpEnabledExtensionNames = instanceExtensions
	// 	instanceCreateInfo.EnabledExtensionCount = uint32(len(instanceExtensions))
	// 	instanceCreateInfo.Flags = vk.InstanceCreateFlags(vk.InstanceCreateEnumeratePortabilityBit)
	// }
	vkDebug = sys.cfg.Video.RendererDebugMode
	if vkDebug {
		if r.checkValidationLayerSupport() {
			instanceCreateInfo.EnabledLayerCount = uint32(len(vk_validationLayers))
			instanceCreateInfo.PpEnabledLayerNames = vk_validationLayers
		} else {
			log.Println("Vulkan validation layers requested but not available")
		}
	}
	var instance vk.Instance
	err := vk.Error(vk.CreateInstance(instanceCreateInfo, nil, &instance))
	if err != nil {
		err = fmt.Errorf("vkCreateInstance failed with %s", err)
		return err
	}
	r.instance = instance
	vk.InitInstance(r.instance)

	surface, err := sys.window.VulkanCreateSurface(r.instance)
	if err != nil {
		vk.DestroyInstance(r.instance, nil)
		err = fmt.Errorf("vkCreateWindowSurface failed with %s", err)
		return err
	}

	r.surface = vk.SurfaceFromPointer(uintptr(surface))

	if r.gpuDevices, err = r.getPhysicalDevices(r.instance); err != nil {
		r.gpuDevices = nil
		vk.DestroyInstance(r.instance, nil)
		return err
	}
	//Try to use discrete GPU, use the first available GPU if discrete GPU is not available
	r.gpuIndex = 0
	for i := range r.gpuDevices {
		var gpuProperties vk.PhysicalDeviceProperties
		vk.GetPhysicalDeviceProperties(r.gpuDevices[i], &gpuProperties)
		gpuProperties.Deref()
		if gpuProperties.DeviceType == vk.PhysicalDeviceTypeDiscreteGpu {
			//gpuProperties.Limits.Deref()
			r.maxAnisotropy = gpuProperties.Limits.MaxSamplerAnisotropy
			r.minUniformBufferOffsetAlignment = uint32(gpuProperties.Limits.MinUniformBufferOffsetAlignment)
			r.maxImageArrayLayers = gpuProperties.Limits.MaxImageArrayLayers
			maxSampleFlag := int32(gpuProperties.Limits.FramebufferColorSampleCounts & gpuProperties.Limits.FramebufferDepthSampleCounts)
			maxSamples := 1
			if maxSampleFlag&int32(vk.SampleCount64Bit) > 0 {
				maxSamples = 64
			} else if maxSampleFlag&int32(vk.SampleCount32Bit) > 0 {
				maxSamples = 32
			} else if maxSampleFlag&int32(vk.SampleCount16Bit) > 0 {
				maxSamples = 16
			} else if maxSampleFlag&int32(vk.SampleCount8Bit) > 0 {
				maxSamples = 8
			} else if maxSampleFlag&int32(vk.SampleCount4Bit) > 0 {
				maxSamples = 4
			} else if maxSampleFlag&int32(vk.SampleCount2Bit) > 0 {
				maxSamples = 2
			}
			if sys.msaa > int32(maxSamples) {
				sys.cfg.SetValueUpdate("Video.MSAA", maxSamples)
				sys.msaa = int32(maxSamples)
			}
			r.gpuIndex = uint32(i)
			break
		}
	}
	queueCreateInfos := []vk.DeviceQueueCreateInfo{{
		SType:            vk.StructureTypeDeviceQueueCreateInfo,
		QueueCount:       1,
		PQueuePriorities: []float32{1.0},
	}}
	deviceExtensions := []string{
		"VK_KHR_swapchain\x00",
		"VK_KHR_push_descriptor\x00",
	}

	if runtime.GOOS == "darwin" {
		deviceExtensions = append(deviceExtensions, "VK_KHR_portability_subset\x00")
	}
	deviceCreateInfo := &vk.DeviceCreateInfo{
		SType:                   vk.StructureTypeDeviceCreateInfo,
		QueueCreateInfoCount:    uint32(len(queueCreateInfos)),
		PQueueCreateInfos:       queueCreateInfos,
		EnabledExtensionCount:   uint32(len(deviceExtensions)),
		PpEnabledExtensionNames: deviceExtensions,
	}

	sync2 := vk.PhysicalDeviceSynchronization2Features{
		SType:            vk.StructureTypePhysicalDeviceSynchronization2Features,
		Synchronization2: vk.True,
	}
	deviceCreateInfo.PNext = unsafe.Pointer(&sync2)
	multiViewFeature := vk.PhysicalDeviceMultiviewFeatures{
		SType:     vk.StructureTypePhysicalDeviceMultiviewFeatures,
		Multiview: vk.True,
	}
	sync2.PNext = unsafe.Pointer(&multiViewFeature)

	dynamicRenderFeature := vk.PhysicalDeviceDynamicRenderingFeatures{
		SType:            vk.StructureTypePhysicalDeviceDynamicRenderingFeatures,
		DynamicRendering: vk.True,
	}
	multiViewFeature.PNext = unsafe.Pointer(&dynamicRenderFeature)

	if sys.cfg.Video.EnableModelShadow {
		vulkan12Features := vk.PhysicalDeviceVulkan12Features{
			SType:                     vk.StructureTypePhysicalDeviceVulkan12Features,
			ShaderOutputViewportIndex: vk.True,
			ShaderOutputLayer:         vk.True,
		}
		dynamicRenderFeature.PNext = unsafe.Pointer(&vulkan12Features)
	}
	if vkDebug && r.checkValidationLayerSupport() {
		deviceCreateInfo.EnabledLayerCount = uint32(len(vk_validationLayers))
		deviceCreateInfo.PpEnabledLayerNames = vk_validationLayers
	}
	deviceCreateInfo.PEnabledFeatures = []vk.PhysicalDeviceFeatures{}
	if r.maxAnisotropy > 0 {
		deviceCreateInfo.PEnabledFeatures = append(deviceCreateInfo.PEnabledFeatures, vk.PhysicalDeviceFeatures{SamplerAnisotropy: vk.True})
	}
	if sys.cfg.Video.EnableModel {
		if len(deviceCreateInfo.PEnabledFeatures) > 0 {
			deviceCreateInfo.PEnabledFeatures[0].ImageCubeArray = vk.True
		} else {
			deviceCreateInfo.PEnabledFeatures = append(deviceCreateInfo.PEnabledFeatures, vk.PhysicalDeviceFeatures{ImageCubeArray: vk.True})
		}
	}
	var device vk.Device
	err = vk.Error(vk.CreateDevice(r.gpuDevices[r.gpuIndex], deviceCreateInfo, nil, &device))
	if err != nil {
		//r.gpuDevices = nil
		vk.DestroySurface(r.instance, r.surface, nil)
		vk.DestroyInstance(r.instance, nil)
		err = fmt.Errorf("vkCreateDevice failed with %s", err)
		return err
	} else {
		r.device = device
		var queue vk.Queue
		vk.GetDeviceQueue(device, 0, 0, &queue)
		r.queue = queue
	}

	return nil
}

func (r *Renderer_VK) getPhysicalDevices(instance vk.Instance) ([]vk.PhysicalDevice, error) {
	var gpuCount uint32
	err := vk.Error(vk.EnumeratePhysicalDevices(instance, &gpuCount, nil))
	if err != nil {
		err = fmt.Errorf("vkEnumeratePhysicalDevices failed with %s", err)
		return nil, err
	}
	if gpuCount == 0 {
		err = fmt.Errorf("getPhysicalDevice: no GPUs found on the system")
		return nil, err
	}
	gpuList := make([]vk.PhysicalDevice, gpuCount)
	err = vk.Error(vk.EnumeratePhysicalDevices(instance, &gpuCount, gpuList))
	if err != nil {
		err = fmt.Errorf("vkEnumeratePhysicalDevices failed with %s", err)
		return nil, err
	}
	return gpuList, nil
}

func (r *Renderer_VK) physicalDeviceType(dev vk.PhysicalDeviceType) string {
	switch dev {
	case vk.PhysicalDeviceTypeIntegratedGpu:
		return "Integrated GPU"
	case vk.PhysicalDeviceTypeDiscreteGpu:
		return "Discrete GPU"
	case vk.PhysicalDeviceTypeVirtualGpu:
		return "Virtual GPU"
	case vk.PhysicalDeviceTypeCpu:
		return "CPU"
	case vk.PhysicalDeviceTypeOther:
		return "Other"
	default:
		return "Unknown"
	}
}

func (r *Renderer_VK) GetPresentMode(vsync bool) vk.PresentMode {
	if !vsync {
		var presentModeCount uint32
		err := vk.Error(vk.GetPhysicalDeviceSurfacePresentModes(r.gpuDevices[r.gpuIndex], r.surface, &presentModeCount, nil))
		if err != nil {
			err = fmt.Errorf("vkGetPhysicalDeviceSurfacePresentModesKHR failed with %s", err)
			panic(err)
		}
		if presentModeCount > 0 {
			presentModes := make([]vk.PresentMode, presentModeCount)
			err := vk.Error(vk.GetPhysicalDeviceSurfacePresentModes(r.gpuDevices[r.gpuIndex], r.surface, &presentModeCount, presentModes))
			if err != nil {
				err = fmt.Errorf("vkGetPhysicalDeviceSurfacePresentModes failed with %s", err)
				panic(err)
			}
			for i := range presentModes {
				if presentModes[i] == vk.PresentModeMailbox {
					return vk.PresentModeMailbox
				}
				if presentModes[i] == vk.PresentModeImmediate {
					return vk.PresentModeImmediate
				}
			}
		}
	}
	return vk.PresentModeFifo
}

func (r *Renderer_VK) PrintInfo() {
	var gpuProperties vk.PhysicalDeviceProperties
	vk.GetPhysicalDeviceProperties(r.gpuDevices[r.gpuIndex], &gpuProperties)
	gpuProperties.Deref()

	log.Println("VULKAN PROPERTIES AND SURFACE CAPABILITES")
	log.Println(vk.ToString(gpuProperties.DeviceName[:]))
	log.Println(fmt.Sprintf("%x", gpuProperties.VendorID))
	log.Println(r.physicalDeviceType(gpuProperties.DeviceType))
	log.Println(len(r.gpuDevices))
	log.Println(vk.Version(gpuProperties.ApiVersion))
	log.Println(vk.Version(gpuProperties.DriverVersion))

	log.Println("VULKAN SUPPORTED DEVICE EXTENSIONS")
	var extensionCount uint32
	var extensionProperties []vk.ExtensionProperties
	vk.EnumerateDeviceExtensionProperties(r.gpuDevices[r.gpuIndex], "", &extensionCount, nil)
	extensionProperties = make([]vk.ExtensionProperties, extensionCount)
	vk.EnumerateDeviceExtensionProperties(r.gpuDevices[r.gpuIndex], "", &extensionCount, extensionProperties)
	for i := range extensionProperties {
		extensionProperties[i].Deref()
		log.Println(vk.ToString(extensionProperties[i].ExtensionName[:]))
	}
}

func (r *Renderer_VK) checkValidationLayerSupport() bool {
	var layerCount uint32
	var layerProperties []vk.LayerProperties
	vk.EnumerateInstanceLayerProperties(&layerCount, nil)
	layerProperties = make([]vk.LayerProperties, layerCount)
	err := vk.Error(vk.EnumerateInstanceLayerProperties(&layerCount, layerProperties))
	if err != nil {
		panic(err)
	}
	for _, layerName := range vk_validationLayers {
		for _, layerProperty := range layerProperties {
			layerProperty.Deref()
			s := vk.ToString(layerProperty.LayerName[:])
			if s == layerName[:len(layerName)-1] {
				return true
			}
		}
	}
	return false
}

func (r *Renderer_VK) CreateRenderTarget(renderpass vk.RenderPass, width, height uint32, numSamples int32, createDepthTexture bool) *VulkanRenderTargetInfo {
	var rt VulkanRenderTargetInfo
	rt.texture = r.CreateRenderTargetTexture(width, height, numSamples, createDepthTexture)
	if createDepthTexture {
		rt.depthTexture = r.CreateRenderTargetDepthTexture(width, height, numSamples)
		rt.framebuffer = r.CreateRenderTargetFramebuffer(width, height, []vk.ImageView{rt.texture.imageView, rt.depthTexture.imageView}, renderpass, 1)
	} else {
		rt.framebuffer = r.CreateRenderTargetFramebuffer(width, height, []vk.ImageView{rt.texture.imageView}, renderpass, 1)
	}
	return &rt
}

func (r *Renderer_VK) CreateRenderTargetTexture(width, height uint32, numSamples int32, main bool) *Texture_VK {
	t := &Texture_VK{int32(width), int32(height), 32, false, 1, [2]int32{0, 0}, [4]float32{0, 0, 1, 1}, nil, nil, nil}
	usage := vk.ImageUsageFlags(vk.ImageUsageColorAttachmentBit)
	if main {
		usage = usage | vk.ImageUsageFlags(vk.ImageUsageTransferSrcBit)
	} else {
		usage = usage | vk.ImageUsageFlags(vk.ImageUsageSampledBit) | vk.ImageUsageFlags(vk.ImageUsageTransferDstBit)
	}
	t.img = r.CreateImage(width, height, r.swapchains[0].format, 1, 1, usage, numSamples, vk.ImageTilingOptimal, false)
	imageMemory := r.AllocateImageMemory(t.img, vk.MemoryPropertyDeviceLocalBit)
	t.imageView = r.CreateImageView(t.img, r.swapchains[0].format, 0, 1, 1, false)

	runtime.SetFinalizer(t, func(t *Texture_VK) {
		r.destroyResourceQueues[r.destroyResourceQueueIndex] <- VulkanResource{
			VulkanResourceTypeTexture,
			[]interface{}{
				t.img, t.imageView, imageMemory,
			},
		}
	})
	return t
}

func (r *Renderer_VK) CreateRenderTargetDepthTexture(width, height uint32, numSamples int32) *Texture_VK {
	t := &Texture_VK{int32(width), int32(height), 32, false, 1, [2]int32{0, 0}, [4]float32{0, 0, 1, 1}, nil, nil, nil}
	usage := vk.ImageUsageFlags(vk.ImageUsageDepthStencilAttachmentBit)
	if numSamples > 1 {
		usage = usage | vk.ImageUsageFlags(vk.ImageUsageTransferSrcBit)
	}
	t.img = r.CreateImage(width, height, vk.FormatD32Sfloat, 1, 1, usage, numSamples, vk.ImageTilingOptimal, false)
	imageMemory := r.AllocateImageMemory(t.img, vk.MemoryPropertyDeviceLocalBit)
	t.imageView = r.CreateImageView(t.img, vk.FormatD32Sfloat, 0, 1, 1, false)

	runtime.SetFinalizer(t, func(t *Texture_VK) {
		r.destroyResourceQueues[r.destroyResourceQueueIndex] <- VulkanResource{
			VulkanResourceTypeTexture,
			[]interface{}{
				t.img, t.imageView, imageMemory,
			},
		}
	})
	commandBuffer := r.BeginSingleTimeCommands()
	barriers := []vk.ImageMemoryBarrier{
		{
			SType:               vk.StructureTypeImageMemoryBarrier,
			OldLayout:           vk.ImageLayoutUndefined,
			NewLayout:           vk.ImageLayoutDepthStencilAttachmentOptimal,
			SrcAccessMask:       vk.AccessFlags(vk.AccessNone),
			DstAccessMask:       vk.AccessFlags(vk.AccessDepthStencilAttachmentReadBit),
			SrcQueueFamilyIndex: vk.QueueFamilyIgnored,
			DstQueueFamilyIndex: vk.QueueFamilyIgnored,
			Image:               t.img,
			SubresourceRange: vk.ImageSubresourceRange{
				AspectMask:     vk.ImageAspectFlags(vk.ImageAspectDepthBit),
				BaseMipLevel:   0,
				LevelCount:     1,
				BaseArrayLayer: 0,
				LayerCount:     1,
			},
		},
	}
	vk.CmdPipelineBarrier(commandBuffer, vk.PipelineStageFlags(vk.PipelineStageTransferBit), vk.PipelineStageFlags(vk.PipelineStageLateFragmentTestsBit), 0, 0, nil, 0, nil, uint32(len(barriers)), barriers)
	vk.EndCommandBuffer(commandBuffer)
	r.tempCommands = append(gfx.(*Renderer_VK).tempCommands, commandBuffer)
	return t
}
func (r *Renderer_VK) CreateRenderTargetFramebuffer(width, height uint32, attachments []vk.ImageView, renderPass vk.RenderPass, layer uint32) vk.Framebuffer {
	var framebuffer vk.Framebuffer
	fbCreateInfo := vk.FramebufferCreateInfo{
		SType:           vk.StructureTypeFramebufferCreateInfo,
		RenderPass:      renderPass,
		Layers:          layer,
		AttachmentCount: uint32(len(attachments)), // 2 if has depthView
		PAttachments:    attachments,
		Width:           width,
		Height:          height,
	}
	err := vk.Error(vk.CreateFramebuffer(r.device, &fbCreateInfo, nil, &framebuffer))
	if err != nil {
		err = fmt.Errorf("vk.CreateFramebuffer failed with %s", err)
		panic(err)
	}
	return framebuffer
}

func (r *Renderer_VK) createPalTexture(size uint32) {
	r.palTexture.size = size
	r.palTexture.textures = make([]*Texture_VK, 0, 1)
	r.palTexture.emptySlot = list.New()
	r.addPalTexture()
}

func (r *Renderer_VK) addPalTexture() {
	index := uint32(len(r.palTexture.textures))
	for i := uint32(0); i < r.palTexture.size; i++ {
		r.palTexture.emptySlot.PushBack([2]uint32{index, uint32(i)})
	}
	t := &Texture_VK{int32(r.palTexture.size), int32(r.palTexture.size), 32, false, 1, [2]int32{0, 0}, [4]float32{0, 0, 1, 1}, nil, nil, nil}
	t.img = r.CreateImage(uint32(t.width), uint32(t.height), vk.FormatR8g8b8a8Unorm, 1, 1, vk.ImageUsageFlags(vk.ImageUsageTransferDstBit|vk.ImageUsageSampledBit), 1, vk.ImageTilingLinear, false)
	imageMemory := r.AllocateImageMemory(t.img, vk.MemoryPropertyDeviceLocalBit)
	t.imageView = r.CreateImageView(t.img, vk.FormatR8g8b8a8Unorm, 0, 1, 1, false)
	r.palTexture.textures = append(r.palTexture.textures, t)
	runtime.SetFinalizer(t, func(t *Texture_VK) {
		r.destroyResourceQueues[r.destroyResourceQueueIndex] <- VulkanResource{
			VulkanResourceTypeTexture,
			[]interface{}{
				t.img, t.imageView, imageMemory,
			},
		}
	})

	commandBuffer := r.BeginSingleTimeCommands()
	barrier := vk.ImageMemoryBarrier{
		SType:               vk.StructureTypeImageMemoryBarrier,
		OldLayout:           vk.ImageLayoutUndefined,
		NewLayout:           vk.ImageLayoutShaderReadOnlyOptimal,
		SrcAccessMask:       vk.AccessFlags(vk.AccessNone),
		DstAccessMask:       vk.AccessFlags(vk.AccessShaderReadBit),
		SrcQueueFamilyIndex: vk.QueueFamilyIgnored,
		DstQueueFamilyIndex: vk.QueueFamilyIgnored,
		Image:               t.img,
		SubresourceRange: vk.ImageSubresourceRange{
			AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
			BaseMipLevel:   0,
			LevelCount:     1,
			BaseArrayLayer: 0,
			LayerCount:     1,
		},
	}
	vk.CmdPipelineBarrier(commandBuffer, vk.PipelineStageFlags(vk.PipelineStageTopOfPipeBit), vk.PipelineStageFlags(vk.PipelineStageFragmentShaderBit), 0, 0, nil, 0, nil, 1, []vk.ImageMemoryBarrier{barrier})
	r.EndSingleTimeCommands(commandBuffer)
}
func (r *Renderer_VK) createShadowMapTexture(widthHeight int32) *Texture_VK {
	t := &Texture_VK{widthHeight, widthHeight, 96, false, 1, [2]int32{0, 0}, [4]float32{0, 0, 1, 1}, nil, nil, nil}
	t.sampler = r.GetSampler(VulkanSamplerInfo{TextureSamplingFilterNearest, TextureSamplingFilterNearest, TextureSamplingWrapClampToEdge, TextureSamplingWrapClampToEdge})
	format := vk.FormatD32Sfloat
	t.img = r.CreateImage(uint32(widthHeight), uint32(widthHeight), format, 1, 6*4, vk.ImageUsageFlags(vk.ImageUsageDepthStencilAttachmentBit|vk.ImageUsageSampledBit), 1, vk.ImageTilingOptimal, true)
	imageMemory := r.AllocateImageMemory(t.img, vk.MemoryPropertyDeviceLocalBit)
	t.imageView = r.CreateImageView(t.img, vk.FormatD32Sfloat, 0, 1, 6*4, true)
	commandBuffer := gfx.(*Renderer_VK).BeginSingleTimeCommands()

	barriers := []vk.ImageMemoryBarrier{
		{
			SType:               vk.StructureTypeImageMemoryBarrier,
			OldLayout:           vk.ImageLayoutUndefined,
			NewLayout:           vk.ImageLayoutShaderReadOnlyOptimal,
			SrcAccessMask:       vk.AccessFlags(vk.AccessNone),
			DstAccessMask:       vk.AccessFlags(vk.AccessShaderReadBit),
			SrcQueueFamilyIndex: vk.QueueFamilyIgnored,
			DstQueueFamilyIndex: vk.QueueFamilyIgnored,
			Image:               t.img,
			SubresourceRange: vk.ImageSubresourceRange{
				AspectMask:     vk.ImageAspectFlags(vk.ImageAspectDepthBit),
				BaseMipLevel:   0,
				LevelCount:     1,
				BaseArrayLayer: 0,
				LayerCount:     6 * 4,
			},
		},
	}
	vk.CmdPipelineBarrier(commandBuffer, vk.PipelineStageFlags(vk.PipelineStageTopOfPipeBit), vk.PipelineStageFlags(vk.PipelineStageFragmentShaderBit), 0, 0, nil, 0, nil, uint32(len(barriers)), barriers)
	vk.EndCommandBuffer(commandBuffer)
	r.tempCommands = append(gfx.(*Renderer_VK).tempCommands, commandBuffer)
	runtime.SetFinalizer(t, func(t *Texture_VK) {
		r.destroyResourceQueues[r.destroyResourceQueueIndex] <- VulkanResource{
			VulkanResourceTypeTexture,
			[]interface{}{
				t.img, t.imageView, imageMemory,
			},
		}
	})
	return t
}

func (r *Renderer_VK) CreateBuffer(size vk.DeviceSize, usage vk.BufferUsageFlags, properties vk.MemoryPropertyFlagBits, bufferMemory *vk.DeviceMemory) (vk.Buffer, error) {
	queueFamilyIdx := []uint32{0}
	bufferCreateInfo := vk.BufferCreateInfo{
		SType:                 vk.StructureTypeBufferCreateInfo,
		Size:                  size,
		Usage:                 usage,
		SharingMode:           vk.SharingModeExclusive,
		QueueFamilyIndexCount: 1,
		PQueueFamilyIndices:   queueFamilyIdx,
	}
	var buffer vk.Buffer
	err := vk.Error(vk.CreateBuffer(r.device, &bufferCreateInfo, nil, &buffer))
	if err != nil {
		err = fmt.Errorf("vk.CreateBuffer failed with %s", err)
		return nil, err
	}
	var memReq vk.MemoryRequirements
	vk.GetBufferMemoryRequirements(r.device, buffer, &memReq)
	//memReq.Deref()
	memoryTypeIndex, ok := r.memoryTypeMap[properties]
	if !ok {
		memoryTypeIndex, _ = vk.FindMemoryTypeIndex(r.gpuDevices[r.gpuIndex], memReq.MemoryTypeBits, properties)
		r.memoryTypeMap[properties] = memoryTypeIndex
	}
	allocInfo := vk.MemoryAllocateInfo{
		SType:           vk.StructureTypeMemoryAllocateInfo,
		AllocationSize:  memReq.Size,
		MemoryTypeIndex: memoryTypeIndex, // see below
	}
	err = vk.Error(vk.AllocateMemory(r.device, &allocInfo, nil, bufferMemory))
	if err != nil {
		err = fmt.Errorf("vk.AllocateMemory failed with %s", err)
		return buffer, err
	}
	err = vk.Error(vk.BindBufferMemory(r.device, buffer, *bufferMemory, 0))
	if err != nil {
		err = fmt.Errorf("vk.BindBufferMemory failed with %s", err)
		return buffer, err
	}

	return buffer, nil
}
func (r *Renderer_VK) CreateSwapchain() error {
	gpu := r.gpuDevices[r.gpuIndex]

	var surfaceCapabilities vk.SurfaceCapabilities
	err := vk.Error(vk.GetPhysicalDeviceSurfaceCapabilities(gpu, r.surface, &surfaceCapabilities))
	if err != nil {
		err = fmt.Errorf("vk.GetPhysicalDeviceSurfaceCapabilities failed with %s", err)
		return err
	}
	var formatCount uint32
	vk.GetPhysicalDeviceSurfaceFormats(gpu, r.surface, &formatCount, nil)
	formats := make([]vk.SurfaceFormat, formatCount)
	vk.GetPhysicalDeviceSurfaceFormats(gpu, r.surface, &formatCount, formats)

	chosenFormat := -1
	for i := 0; i < int(formatCount); i++ {
		//formats[i].Deref()
		if formats[i].Format == vk.FormatB8g8r8a8Unorm || formats[i].Format == vk.FormatR8g8b8a8Unorm {
			chosenFormat = i
			break
		}
	}
	if chosenFormat < 0 {
		if len(formats) > 0 {
			fmt.Printf("Choosing fallback SurfaceFormat")
			//formats[0].Deref()
			chosenFormat = 0
			fmt.Printf("Falling back on surface format: %v", formats[0].Format)
		} else {
			err := fmt.Errorf("vk.GetPhysicalDeviceSurfaceFormats not found suitable format")
			return err
		}

	}

	//surfaceCapabilities.Deref()
	imageExtent := surfaceCapabilities.CurrentExtent
	width, height := sys.window.GetSize()
	imageExtent.Width = uint32(width)
	imageExtent.Height = uint32(height)
	queueFamily := []uint32{0}
	swapchainCreateInfo := vk.SwapchainCreateInfo{
		SType:           vk.StructureTypeSwapchainCreateInfo,
		Surface:         r.surface,
		MinImageCount:   surfaceCapabilities.MinImageCount + 1,
		ImageFormat:     formats[chosenFormat].Format,
		ImageColorSpace: formats[chosenFormat].ColorSpace,
		ImageExtent:     imageExtent,
		ImageUsage:      vk.ImageUsageFlags(vk.ImageUsageColorAttachmentBit | vk.ImageUsageTransferSrcBit),
		PreTransform:    surfaceCapabilities.CurrentTransform,

		ImageArrayLayers:      1,
		ImageSharingMode:      vk.SharingModeExclusive,
		CompositeAlpha:        vk.CompositeAlphaOpaqueBit,
		QueueFamilyIndexCount: 1,
		PQueueFamilyIndices:   queueFamily,
		PresentMode:           r.GetPresentMode(sys.cfg.Video.VSync > 0),
		Clipped:               vk.False,
		OldSwapchain:          vk.NullSwapchain,
	}
	r.swapchains = make([]*VulkanSwapchainInfo, 1)
	r.swapchains[0] = &VulkanSwapchainInfo{}
	r.swapchains[0].currentImageIndex = 0
	err = vk.Error(vk.CreateSwapchain(r.device, &swapchainCreateInfo, nil, &(r.swapchains[0].swapchain)))
	if err != nil {
		err = fmt.Errorf("vk.CreateSwapchain failed with %s", err)
		return err
	}
	err = vk.Error(vk.GetSwapchainImages(r.device, r.swapchains[0].swapchain, &r.swapchains[0].imageCount, nil))
	if err != nil {
		err = fmt.Errorf("vk.GetSwapchainImages failed with %s", err)
		return err
	}
	r.swapchains[0].images = make([]vk.Image, r.swapchains[0].imageCount)
	err = vk.Error(vk.GetSwapchainImages(r.device, r.swapchains[0].swapchain, &r.swapchains[0].imageCount, r.swapchains[0].images))
	if err != nil {
		err = fmt.Errorf("vk.GetSwapchainImages failed with %s", err)
		return err
	}
	r.swapchains[0].format = formats[chosenFormat].Format
	r.swapchains[0].extent = surfaceCapabilities.CurrentExtent
	//r.swapchains[0].extent.Deref()
	// for i := range formats {
	// 	formats[i].Free()
	// }
	r.CreateSwapchainImageViews(r.swapchains[0])
	return nil
}
func (r *Renderer_VK) CreateSwapchainImageViews(swapchain *VulkanSwapchainInfo) error {
	swapchain.imageViews = make([]vk.ImageView, len(swapchain.images))
	for i := range swapchain.images {
		swapchain.imageViews[i] = r.CreateImageView(swapchain.images[i], swapchain.format, 0, 1, 1, false)
	}
	return nil
}
func (r *Renderer_VK) RecreateSwapchain() {
	vk.DeviceWaitIdle(r.device)
	r.DestroySwapchain()
	err := r.CreateSwapchain()
	if err != nil {
		panic(err)
	}
	err = r.CreateSwapChainFramebuffer(r.swapchains[0], r.swapchainRenderPass.renderPass)
	if err != nil {
		panic(err)
	}
}
func (r *Renderer_VK) CreateMainRenderPass(sweapchain *VulkanSwapchainInfo, numSamples int32, containsDepthTexture bool) (*VulkanRenderInfo, error) {
	attachmentDescriptions := []vk.AttachmentDescription{
		{
			Format:         sweapchain.format,
			Samples:        vk.SampleCountFlagBits(numSamples),
			LoadOp:         vk.AttachmentLoadOpLoad,
			StoreOp:        vk.AttachmentStoreOpStore,
			StencilLoadOp:  vk.AttachmentLoadOpDontCare,
			StencilStoreOp: vk.AttachmentStoreOpDontCare,
			InitialLayout:  vk.ImageLayoutColorAttachmentOptimal,
			FinalLayout:    vk.ImageLayoutColorAttachmentOptimal,
		},
	}
	colorAttachments := []vk.AttachmentReference{
		{
			Attachment: 0,
			Layout:     vk.ImageLayoutColorAttachmentOptimal,
		},
	}
	subpassDescriptions := []vk.SubpassDescription{{
		PipelineBindPoint:    vk.PipelineBindPointGraphics,
		ColorAttachmentCount: 1,
		PColorAttachments:    colorAttachments,
	}}
	if containsDepthTexture {
		attachmentDescriptions = append(attachmentDescriptions, vk.AttachmentDescription{
			Format:         vk.FormatD32Sfloat,
			Samples:        vk.SampleCountFlagBits(numSamples),
			LoadOp:         vk.AttachmentLoadOpLoad,
			StoreOp:        vk.AttachmentStoreOpStore,
			StencilLoadOp:  vk.AttachmentLoadOpDontCare,
			StencilStoreOp: vk.AttachmentStoreOpDontCare,
			InitialLayout:  vk.ImageLayoutDepthStencilAttachmentOptimal,
			FinalLayout:    vk.ImageLayoutDepthStencilAttachmentOptimal,
		})
		colorAttachments = append(colorAttachments, vk.AttachmentReference{
			Attachment: 1,
			Layout:     vk.ImageLayoutDepthStencilAttachmentOptimal,
		})
		subpassDescriptions[0].PDepthStencilAttachment = &vk.AttachmentReference{
			Attachment: 1,
			Layout:     vk.ImageLayoutDepthStencilAttachmentOptimal,
		}
	}
	renderPassCreateInfo := vk.RenderPassCreateInfo{
		SType:           vk.StructureTypeRenderPassCreateInfo,
		AttachmentCount: uint32(len(attachmentDescriptions)),
		PAttachments:    attachmentDescriptions,
		SubpassCount:    uint32(len(subpassDescriptions)),
		PSubpasses:      subpassDescriptions,
	}
	renderInfo := &VulkanRenderInfo{}
	err := vk.Error(vk.CreateRenderPass(r.device, &renderPassCreateInfo, nil, &renderInfo.renderPass))
	if err != nil {
		err = fmt.Errorf("vk.CreateRenderPass failed with %s", err)
		return nil, err
	}
	return renderInfo, nil
}
func (r *Renderer_VK) CreatePostProcessingRenderPass(sweapchain *VulkanSwapchainInfo) (*VulkanRenderInfo, error) {
	attachmentDescriptions := []vk.AttachmentDescription{{
		Format:         sweapchain.format,
		Samples:        vk.SampleCount1Bit,
		LoadOp:         vk.AttachmentLoadOpClear,
		StoreOp:        vk.AttachmentStoreOpStore,
		StencilLoadOp:  vk.AttachmentLoadOpDontCare,
		StencilStoreOp: vk.AttachmentStoreOpDontCare,
		InitialLayout:  vk.ImageLayoutUndefined,
		FinalLayout:    vk.ImageLayoutShaderReadOnlyOptimal,
	}}
	colorAttachments := []vk.AttachmentReference{{
		Attachment: 0,
		Layout:     vk.ImageLayoutColorAttachmentOptimal,
	}}
	subpassDescriptions := []vk.SubpassDescription{{
		PipelineBindPoint:    vk.PipelineBindPointGraphics,
		ColorAttachmentCount: 1,
		PColorAttachments:    colorAttachments,
	}}
	renderPassCreateInfo := vk.RenderPassCreateInfo{
		SType:           vk.StructureTypeRenderPassCreateInfo,
		AttachmentCount: 1,
		PAttachments:    attachmentDescriptions,
		SubpassCount:    1,
		PSubpasses:      subpassDescriptions,
	}
	renderInfo := &VulkanRenderInfo{}
	err := vk.Error(vk.CreateRenderPass(r.device, &renderPassCreateInfo, nil, &renderInfo.renderPass))
	if err != nil {
		err = fmt.Errorf("vk.CreateRenderPass failed with %s", err)
		return nil, err
	}
	return renderInfo, nil
}
func (r *Renderer_VK) CreateSwapchainRenderPass(sweapchain *VulkanSwapchainInfo) (*VulkanRenderInfo, error) {
	attachmentDescriptions := []vk.AttachmentDescription{{
		Format:         sweapchain.format,
		Samples:        vk.SampleCount1Bit,
		LoadOp:         vk.AttachmentLoadOpClear,
		StoreOp:        vk.AttachmentStoreOpStore,
		StencilLoadOp:  vk.AttachmentLoadOpDontCare,
		StencilStoreOp: vk.AttachmentStoreOpDontCare,
		InitialLayout:  vk.ImageLayoutUndefined,
		FinalLayout:    vk.ImageLayoutPresentSrc,
	}}
	colorAttachments := []vk.AttachmentReference{{
		Attachment: 0,
		Layout:     vk.ImageLayoutColorAttachmentOptimal,
	}}
	subpassDescriptions := []vk.SubpassDescription{{
		PipelineBindPoint:    vk.PipelineBindPointGraphics,
		ColorAttachmentCount: 1,
		PColorAttachments:    colorAttachments,
	}}
	renderPassCreateInfo := vk.RenderPassCreateInfo{
		SType:           vk.StructureTypeRenderPassCreateInfo,
		AttachmentCount: 1,
		PAttachments:    attachmentDescriptions,
		SubpassCount:    1,
		PSubpasses:      subpassDescriptions,
	}
	renderInfo := &VulkanRenderInfo{}
	err := vk.Error(vk.CreateRenderPass(r.device, &renderPassCreateInfo, nil, &renderInfo.renderPass))
	if err != nil {
		err = fmt.Errorf("vk.CreateRenderPass failed with %s", err)
		return nil, err
	}
	return renderInfo, nil
}

type sliceHeader struct {
	Data uintptr
	Len  int
	Cap  int
}

func (r *Renderer_VK) CreateSpriteProgram() (*VulkanProgramInfo, error) {
	program := &VulkanProgramInfo{}
	var uniformBufferMemory vk.DeviceMemory
	var err error
	minAlignment := r.minUniformBufferOffsetAlignment
	uniformSize := uint32(MaxI(int(unsafe.Sizeof(VulkanSpriteProgramVertUniformBufferObject{})), int(unsafe.Sizeof(VulkanSpriteProgramFragUniformBufferObject{}))))
	if uniformSize < minAlignment {
		uniformSize = minAlignment
	} else if uniformSize > minAlignment && minAlignment > 0 && uniformSize%minAlignment != 0 {
		uniformSize = (uniformSize/minAlignment + 1) * minAlignment
	}
	program.uniformBuffers = make([]VulkanBuffer, 1)
	program.uniformBuffers[0].size = uintptr(10000 * uniformSize)
	program.uniformSize = uniformSize
	program.uniformBuffers[0].buffer, err = r.CreateBuffer(vk.DeviceSize(program.uniformBuffers[0].size), vk.BufferUsageFlags(vk.BufferUsageTransferSrcBit|vk.BufferUsageUniformBufferBit), (vk.MemoryPropertyHostVisibleBit | vk.MemoryPropertyHostCoherentBit), &uniformBufferMemory)
	if err != nil {
		panic(err)
	}
	program.uniformBuffers[0].bufferMemory = uniformBufferMemory
	var uniformData unsafe.Pointer
	vk.MapMemory(r.device, program.uniformBuffers[0].bufferMemory, 0, vk.DeviceSize(program.uniformBuffers[0].size), 0, &uniformData)
	program.uniformBuffers[0].data = uniformData

	VertShader, err := staticFiles.ReadFile("shaders/sprite.vert.spv")
	if err != nil {
		return nil, err
	}
	VertShader2 := make([]uint32, len(VertShader)/4)
	vk.Memcopy(unsafe.Pointer((*sliceHeader)(unsafe.Pointer(&VertShader2)).Data), VertShader)
	vertShader, err := r.CreateShader(r.device, VertShader2)
	if err != nil {
		return nil, err
	}
	defer vk.DestroyShaderModule(r.device, vertShader, nil)

	FragShader, err := staticFiles.ReadFile("shaders/sprite.frag.spv")
	if err != nil {
		return nil, err
	}
	FragShader2 := make([]uint32, len(FragShader)/4)
	vk.Memcopy(unsafe.Pointer((*sliceHeader)(unsafe.Pointer(&FragShader2)).Data), FragShader)
	fragShader, err := r.CreateShader(r.device, FragShader2)
	if err != nil {
		return nil, err
	}
	defer vk.DestroyShaderModule(r.device, fragShader, nil)
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

	samplerLayoutBinding := []vk.DescriptorSetLayoutBinding{
		{
			Binding:            0,
			DescriptorCount:    1,
			DescriptorType:     vk.DescriptorTypeUniformBuffer,
			PImmutableSamplers: nil,
			StageFlags:         vk.ShaderStageFlags(vk.ShaderStageVertexBit),
		},
		{
			Binding:            1,
			DescriptorCount:    1,
			DescriptorType:     vk.DescriptorTypeUniformBuffer,
			PImmutableSamplers: nil,
			StageFlags:         vk.ShaderStageFlags(vk.ShaderStageFragmentBit),
		},
		{
			Binding:            2,
			DescriptorCount:    1,
			DescriptorType:     vk.DescriptorTypeCombinedImageSampler,
			PImmutableSamplers: nil,
			StageFlags:         vk.ShaderStageFlags(vk.ShaderStageFragmentBit),
		},
		{
			Binding:            3,
			DescriptorCount:    1,
			DescriptorType:     vk.DescriptorTypeCombinedImageSampler,
			PImmutableSamplers: nil,
			StageFlags:         vk.ShaderStageFlags(vk.ShaderStageFragmentBit),
		},
	}

	layoutInfo := vk.DescriptorSetLayoutCreateInfo{
		SType:        vk.StructureTypeDescriptorSetLayoutCreateInfo,
		Flags:        vk.DescriptorSetLayoutCreateFlags(vk.DescriptorSetLayoutCreatePushDescriptorBit),
		BindingCount: 4,
		PBindings:    samplerLayoutBinding,
	}
	var descriptorSetLayout vk.DescriptorSetLayout
	vk.CreateDescriptorSetLayout(r.device, &layoutInfo, nil, &descriptorSetLayout)
	program.descriptorSetLayouts = append(program.descriptorSetLayouts, descriptorSetLayout)
	//program.descriptorSetCache, err = v.NewDescriptorSetCache(5000, descriptorSetLayout)
	if err != nil {
		err = fmt.Errorf("vk.CreateDescriptorSetLayout failed with %s", err)
		return nil, err
	}
	pushConstantRanges := []vk.PushConstantRange{
		// palUV
		{
			StageFlags: vk.ShaderStageFlags(vk.ShaderStageFragmentBit),
			Offset:     0,
			Size:       16,
		},
	}
	pipelineLayoutCreateInfo := vk.PipelineLayoutCreateInfo{
		SType:                  vk.StructureTypePipelineLayoutCreateInfo,
		SetLayoutCount:         1,
		PSetLayouts:            []vk.DescriptorSetLayout{descriptorSetLayout},
		PushConstantRangeCount: uint32(len(pushConstantRanges)),
		PPushConstantRanges:    pushConstantRanges,
	}
	var pipelineLayout vk.PipelineLayout
	vk.CreatePipelineLayout(r.device, &pipelineLayoutCreateInfo, nil, &pipelineLayout)
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
		Topology:               vk.PrimitiveTopologyTriangleStrip,
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
	blendFunctions := []struct {
		op  vk.BlendOp
		src vk.BlendFactor
		dst vk.BlendFactor
	}{
		{
			op:  vk.BlendOpAdd,
			src: vk.BlendFactorOne,
			dst: vk.BlendFactorOneMinusSrcAlpha,
		},
		{
			op:  vk.BlendOpAdd,
			src: vk.BlendFactorOne,
			dst: vk.BlendFactorOne,
		},
		{
			op:  vk.BlendOpAdd,
			src: vk.BlendFactorOne,
			dst: vk.BlendFactorZero,
		},
		{
			op:  vk.BlendOpReverseSubtract,
			src: vk.BlendFactorOne,
			dst: vk.BlendFactorOneMinusSrcAlpha,
		},
		{
			op:  vk.BlendOpReverseSubtract,
			src: vk.BlendFactorOne,
			dst: vk.BlendFactorOne,
		},
		{
			op:  vk.BlendOpReverseSubtract,
			src: vk.BlendFactorOne,
			dst: vk.BlendFactorZero,
		},
		{
			op:  vk.BlendOpAdd,
			src: vk.BlendFactorSrcAlpha,
			dst: vk.BlendFactorOneMinusSrcAlpha,
		},
		{
			op:  vk.BlendOpAdd,
			src: vk.BlendFactorSrcAlpha,
			dst: vk.BlendFactorOne,
		},
		{
			op:  vk.BlendOpAdd,
			src: vk.BlendFactorSrcAlpha,
			dst: vk.BlendFactorZero,
		},
		{
			op:  vk.BlendOpReverseSubtract,
			src: vk.BlendFactorSrcAlpha,
			dst: vk.BlendFactorOneMinusSrcAlpha,
		},
		{
			op:  vk.BlendOpReverseSubtract,
			src: vk.BlendFactorSrcAlpha,
			dst: vk.BlendFactorOne,
		},
		{
			op:  vk.BlendOpReverseSubtract,
			src: vk.BlendFactorSrcAlpha,
			dst: vk.BlendFactorZero,
		},
		{
			op:  vk.BlendOpAdd,
			src: vk.BlendFactorZero,
			dst: vk.BlendFactorOneMinusSrcAlpha,
		},
		{
			op:  vk.BlendOpAdd,
			src: vk.BlendFactorZero,
			dst: vk.BlendFactorOne,
		},
		{
			op:  vk.BlendOpAdd,
			src: vk.BlendFactorZero,
			dst: vk.BlendFactorZero,
		},
		{
			op:  vk.BlendOpReverseSubtract,
			src: vk.BlendFactorZero,
			dst: vk.BlendFactorOneMinusSrcAlpha,
		},
		{
			op:  vk.BlendOpReverseSubtract,
			src: vk.BlendFactorZero,
			dst: vk.BlendFactorOne,
		},
		{
			op:  vk.BlendOpReverseSubtract,
			src: vk.BlendFactorZero,
			dst: vk.BlendFactorZero,
		},
	}
	attachmentStates := make([]vk.PipelineColorBlendAttachmentState, 0, len(blendFunctions))
	colorBlendStates := make([]vk.PipelineColorBlendStateCreateInfo, 0, len(blendFunctions))
	pipelineCreateInfos := make([]vk.GraphicsPipelineCreateInfo, 0, len(blendFunctions))
	pipelineIndexMap := map[interface{}]int{}
	MapBlendEquation := func(i vk.BlendOp) BlendEquation {
		var BlendEquationLUT = map[vk.BlendOp]BlendEquation{
			vk.BlendOpAdd:             BlendAdd,
			vk.BlendOpReverseSubtract: BlendReverseSubtract,
		}
		return BlendEquationLUT[i]
	}

	MapBlendFactor := func(i vk.BlendFactor) BlendFunc {
		var MapBlendFactor = map[vk.BlendFactor]BlendFunc{
			vk.BlendFactorOne:              BlendOne,
			vk.BlendFactorZero:             BlendZero,
			vk.BlendFactorSrcAlpha:         BlendSrcAlpha,
			vk.BlendFactorOneMinusSrcAlpha: BlendOneMinusSrcAlpha,
			vk.BlendFactorOneMinusDstColor: BlendOneMinusDstColor,
			vk.BlendFactorDstColor:         BlendDstColor,
		}
		return MapBlendFactor[i]
	}
	for i := range blendFunctions {
		attachmentStates = append(attachmentStates, vk.PipelineColorBlendAttachmentState{
			ColorWriteMask: vk.ColorComponentFlags(
				vk.ColorComponentRBit | vk.ColorComponentGBit |
					vk.ColorComponentBBit | vk.ColorComponentABit,
			),
			BlendEnable:         vk.True,
			SrcColorBlendFactor: blendFunctions[i].src,
			DstColorBlendFactor: blendFunctions[i].dst,
			ColorBlendOp:        blendFunctions[i].op,
			SrcAlphaBlendFactor: vk.BlendFactorOne,
			DstAlphaBlendFactor: vk.BlendFactorZero,
			AlphaBlendOp:        vk.BlendOpAdd,
		})
		colorBlendStates = append(colorBlendStates, vk.PipelineColorBlendStateCreateInfo{
			SType:           vk.StructureTypePipelineColorBlendStateCreateInfo,
			LogicOpEnable:   vk.False,
			LogicOp:         vk.LogicOpCopy,
			AttachmentCount: 1,
			PAttachments:    []vk.PipelineColorBlendAttachmentState{attachmentStates[i]},
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
			PColorBlendState:    &colorBlendStates[i],
			PDynamicState:       &dynamicState,
			Layout:              program.pipelineLayout,
			RenderPass:          r.mainRenderPass.renderPass,
			PDepthStencilState:  &depthStencilState,
		})
		pipelineIndexMap[VulkanBlendState{
			op:  MapBlendEquation(blendFunctions[i].op),
			src: MapBlendFactor(blendFunctions[i].src),
			dst: MapBlendFactor(blendFunctions[i].dst),
		}] = i
	}
	pipelines := make([]vk.Pipeline, len(pipelineCreateInfos))
	err = vk.Error(vk.CreateGraphicsPipelines(r.device,
		r.pipelineCache, uint32(len(pipelineCreateInfos)), pipelineCreateInfos, nil, pipelines))
	if err != nil {
		err = fmt.Errorf("vk.CreateGraphicsPipelines failed with %s", err)
		return nil, err
	}
	program.pipelines = pipelines
	program.pipelineIndexMap = pipelineIndexMap
	return program, nil
}

func (r *Renderer_VK) CreateShadowMapProgram() (*VulkanProgramInfo, error) {
	program := &VulkanProgramInfo{}
	var uniformBufferMemory vk.DeviceMemory
	var err error
	uniformSize := r.AlignUniformSize(uint32(MaxI(VulkanShadowMapUniform1Size, VulkanShadowMapUniform2Size)))
	program.uniformBuffers = make([]VulkanBuffer, 1)
	program.uniformBuffers[0].size = uintptr(2000 * uniformSize)
	program.uniformSize = uniformSize
	program.uniformBuffers[0].buffer, err = r.CreateBuffer(vk.DeviceSize(program.uniformBuffers[0].size), vk.BufferUsageFlags(vk.BufferUsageTransferSrcBit|vk.BufferUsageUniformBufferBit), (vk.MemoryPropertyHostVisibleBit | vk.MemoryPropertyHostCoherentBit), &uniformBufferMemory)
	if err != nil {
		panic(err)
	}
	program.uniformBuffers[0].bufferMemory = uniformBufferMemory
	var uniformData unsafe.Pointer
	vk.MapMemory(r.device, program.uniformBuffers[0].bufferMemory, 0, vk.DeviceSize(program.uniformBuffers[0].size), 0, &uniformData)
	program.uniformBuffers[0].data = uniformData

	VertShader, err := staticFiles.ReadFile("shaders/shadow.vert.spv")
	if err != nil {
		return nil, err
	}
	VertShader2 := make([]uint32, len(VertShader)/4)
	vk.Memcopy(unsafe.Pointer((*sliceHeader)(unsafe.Pointer(&VertShader2)).Data), VertShader)
	vertShader, err := r.CreateShader(r.device, VertShader2)
	if err != nil {
		return nil, err
	}

	FragShader, err := staticFiles.ReadFile("shaders/shadow.frag.spv")
	if err != nil {
		return nil, err
	}
	FragShader2 := make([]uint32, len(FragShader)/4)
	vk.Memcopy(unsafe.Pointer((*sliceHeader)(unsafe.Pointer(&FragShader2)).Data), FragShader)
	fragShader, err := r.CreateShader(r.device, FragShader2)
	if err != nil {
		return nil, err
	}
	program.vertShader = vertShader
	program.fragShader = fragShader
	uniformBindings0 := []vk.DescriptorSetLayoutBinding{
		//VulkanModelProgramUniformBufferObject0
		{
			Binding:            0,
			DescriptorCount:    1,
			DescriptorType:     vk.DescriptorTypeUniformBuffer,
			PImmutableSamplers: nil,
			StageFlags:         vk.ShaderStageFlags(vk.ShaderStageVertexBit | vk.ShaderStageFragmentBit),
		},
		//VulkanModelProgramUniformBufferObject1
		{
			Binding:            1,
			DescriptorCount:    1,
			DescriptorType:     vk.DescriptorTypeUniformBuffer,
			PImmutableSamplers: nil,
			StageFlags:         vk.ShaderStageFlags(vk.ShaderStageFragmentBit),
		},
		//VulkanModelProgramUniformBufferObject2
		{
			Binding:            2,
			DescriptorCount:    1,
			DescriptorType:     vk.DescriptorTypeUniformBuffer,
			PImmutableSamplers: nil,
			StageFlags:         vk.ShaderStageFlags(vk.ShaderStageVertexBit),
		},
		//jointMatrices
		{
			Binding:            3,
			DescriptorCount:    1,
			DescriptorType:     vk.DescriptorTypeCombinedImageSampler,
			PImmutableSamplers: nil,
			StageFlags:         vk.ShaderStageFlags(vk.ShaderStageVertexBit),
		},
		//morphTargetValues
		{
			Binding:            4,
			DescriptorCount:    1,
			DescriptorType:     vk.DescriptorTypeCombinedImageSampler,
			PImmutableSamplers: nil,
			StageFlags:         vk.ShaderStageFlags(vk.ShaderStageVertexBit),
		},
		//tex
		{
			Binding:            5,
			DescriptorCount:    1,
			DescriptorType:     vk.DescriptorTypeCombinedImageSampler,
			PImmutableSamplers: nil,
			StageFlags:         vk.ShaderStageFlags(vk.ShaderStageFragmentBit),
		},
	}
	layoutInfo := vk.DescriptorSetLayoutCreateInfo{
		SType:        vk.StructureTypeDescriptorSetLayoutCreateInfo,
		Flags:        vk.DescriptorSetLayoutCreateFlags(vk.DescriptorSetLayoutCreatePushDescriptorBit),
		BindingCount: uint32(len(uniformBindings0)),
		PBindings:    uniformBindings0,
	}
	var descriptorSetLayout vk.DescriptorSetLayout
	vk.CreateDescriptorSetLayout(r.device, &layoutInfo, nil, &descriptorSetLayout)
	program.descriptorSetLayouts = append(program.descriptorSetLayouts, descriptorSetLayout)
	if err != nil {
		err = fmt.Errorf("vk.CreateDescriptorSetLayout failed with %s", err)
		return nil, err
	}
	pushConstantRanges := []vk.PushConstantRange{
		// modelMatrix, numVertices
		{
			StageFlags: vk.ShaderStageFlags(vk.ShaderStageVertexBit),
			Offset:     0,
			Size:       68,
		},
	}
	pipelineLayoutCreateInfo := vk.PipelineLayoutCreateInfo{
		SType:                  vk.StructureTypePipelineLayoutCreateInfo,
		SetLayoutCount:         1,
		PSetLayouts:            program.descriptorSetLayouts,
		PushConstantRangeCount: uint32(len(pushConstantRanges)),
		PPushConstantRanges:    pushConstantRanges,
	}
	var pipelineLayout vk.PipelineLayout
	vk.CreatePipelineLayout(r.device, &pipelineLayoutCreateInfo, nil, &pipelineLayout)
	program.pipelines = []vk.Pipeline{}
	program.pipelineLayout = pipelineLayout
	program.pipelineIndexMap = map[interface{}]int{}
	return program, nil
}
func (r *Renderer_VK) GetShadowMapPipeline(state *VulkanShadowMapPipelineState) vk.Pipeline {
	if index, ok := r.shadowMapProgram.pipelineIndexMap[*state]; ok {
		return r.shadowMapProgram.pipelines[index]
	}
	vertSpecializationMapEntries := []vk.SpecializationMapEntry{
		{
			ConstantID: 0,
			Size:       4,
			Offset:     0,
		},
		{
			ConstantID: 1,
			Size:       4,
			Offset:     4,
		},
		{
			ConstantID: 2,
			Size:       4,
			Offset:     8,
		},
	}
	fragSpecializationMapEntries := []vk.SpecializationMapEntry{
		{
			ConstantID: 3,
			Size:       4,
			Offset:     0,
		},
	}
	vertData := struct {
		useJoint0    bool
		_padding0    [3]bool
		useJoint1    bool
		_padding1    [3]bool
		useVertColor bool
		_padding2    [3]bool
	}{useJoint0: state.VulkanModelSpecializationConstants0.useJoint0, useJoint1: state.VulkanModelSpecializationConstants0.useJoint1, useVertColor: state.VulkanModelSpecializationConstants0.useVertColor}
	vertSpecializationInfo := []vk.SpecializationInfo{
		{
			DataSize:      uint64(unsafe.Sizeof(state.VulkanModelSpecializationConstants0)),
			MapEntryCount: uint32(len(vertSpecializationMapEntries)),
			PMapEntries:   vertSpecializationMapEntries,
			PData:         unsafe.Pointer(&vertData),
		},
	}
	fragData := struct {
		useTexture bool
		_padding0  [3]bool
	}{useTexture: state.useTexture}
	fragSpecializationInfo := []vk.SpecializationInfo{
		{
			DataSize:      uint64(unsafe.Sizeof(fragData)),
			MapEntryCount: uint32(len(fragSpecializationMapEntries)),
			PMapEntries:   fragSpecializationMapEntries,
			PData:         unsafe.Pointer(&fragData),
		},
	}
	shaderStages := []vk.PipelineShaderStageCreateInfo{
		{
			SType:               vk.StructureTypePipelineShaderStageCreateInfo,
			Stage:               vk.ShaderStageVertexBit,
			Module:              r.shadowMapProgram.vertShader,
			PSpecializationInfo: vertSpecializationInfo,
			PName:               "main\x00",
		},
		{
			SType:               vk.StructureTypePipelineShaderStageCreateInfo,
			Stage:               vk.ShaderStageFragmentBit,
			Module:              r.shadowMapProgram.fragShader,
			PSpecializationInfo: fragSpecializationInfo,
			PName:               "main\x00",
		},
	}
	viewports := []vk.Viewport{{
		MinDepth: 0.0,
		MaxDepth: 1.0,
		X:        0,
		Y:        0,
		Width:    float32(r.shadowMapTextures.width),
		Height:   float32(r.shadowMapTextures.height),
	}}
	scissors := []vk.Rect2D{{
		Offset: vk.Offset2D{
			X: 0,
			Y: 0,
		},
		Extent: vk.Extent2D{
			Width:  uint32(r.shadowMapTextures.width),
			Height: uint32(r.shadowMapTextures.height),
		},
	}}
	viewportState := vk.PipelineViewportStateCreateInfo{
		SType:         vk.StructureTypePipelineViewportStateCreateInfo,
		ViewportCount: 1,
		ScissorCount:  1,
		PViewports:    viewports,
		PScissors:     scissors,
	}
	dynamicState := vk.PipelineDynamicStateCreateInfo{
		SType: vk.StructureTypePipelineDynamicStateCreateInfo,
		//DynamicStateCount: uint32(len(dynamicStates)),
		//PDynamicStates:    dynamicStates,
		DynamicStateCount: 0,
	}
	topology := map[PrimitiveMode]vk.PrimitiveTopology{
		LINES: vk.PrimitiveTopologyLineList,
		//LINE_LOOP:      vk.PrimitiveTopologyLineList,
		LINE_STRIP:     vk.PrimitiveTopologyLineStrip,
		TRIANGLES:      vk.PrimitiveTopologyTriangleList,
		TRIANGLE_STRIP: vk.PrimitiveTopologyTriangleStrip,
		TRIANGLE_FAN:   vk.PrimitiveTopologyTriangleFan,
	}[state.primitiveMode]
	inputAssemblyState := vk.PipelineInputAssemblyStateCreateInfo{
		SType:                  vk.StructureTypePipelineInputAssemblyStateCreateInfo,
		Topology:               topology,
		PrimitiveRestartEnable: vk.False,
	}
	vertexInputBindings := []vk.VertexInputBindingDescription{
		{
			Binding:   0,
			Stride:    4 * 1, // 4 = sizeof(float32)
			InputRate: vk.VertexInputRateVertex,
		},
		{
			Binding:   1,
			Stride:    4 * 3, // 4 = sizeof(float32)
			InputRate: vk.VertexInputRateVertex,
		},
	}
	vertexInputAttributes := []vk.VertexInputAttributeDescription{
		//vertexId
		{
			Binding:  0,
			Location: 0,
			Format:   vk.FormatR32Sint,
			Offset:   0,
		},
		//position
		{
			Binding:  1,
			Location: 1,
			Format:   vk.FormatR32g32b32Sfloat,
			Offset:   0,
		},
	}
	vertexBindingIndex := uint32(2)
	addVertexAttribute := func(size, location uint32, useAttribute bool) {
		format := vk.FormatR32g32Sfloat
		switch size {
		case 3:
			format = vk.FormatR32g32b32Sfloat
			break
		case 4:
			format = vk.FormatR32g32b32a32Sfloat
		}
		if useAttribute {
			vertexInputBindings = append(vertexInputBindings, vk.VertexInputBindingDescription{
				Binding:   vertexBindingIndex,
				Stride:    4 * size, // 4 = sizeof(float32)
				InputRate: vk.VertexInputRateVertex,
			})
			vertexInputAttributes = append(vertexInputAttributes, vk.VertexInputAttributeDescription{
				Binding:  vertexBindingIndex,
				Location: location,
				Format:   format,
				Offset:   0,
			})
			vertexBindingIndex += 1
		} else {
			vertexInputAttributes = append(vertexInputAttributes, vk.VertexInputAttributeDescription{
				Binding:  0,
				Location: location,
				Format:   format,
				Offset:   0,
			})
		}
	}
	addVertexAttribute(2, 2, state.useUV)
	addVertexAttribute(4, 3, state.useVertColor)
	addVertexAttribute(4, 4, state.useJoint0)
	addVertexAttribute(4, 5, state.useJoint0)
	addVertexAttribute(4, 6, state.useJoint1)
	addVertexAttribute(4, 7, state.useJoint1)

	vertexInputState := vk.PipelineVertexInputStateCreateInfo{
		SType:                           vk.StructureTypePipelineVertexInputStateCreateInfo,
		VertexBindingDescriptionCount:   uint32(len(vertexInputBindings)),
		PVertexBindingDescriptions:      vertexInputBindings,
		VertexAttributeDescriptionCount: uint32(len(vertexInputAttributes)),
		PVertexAttributeDescriptions:    vertexInputAttributes,
	}
	sampleMask := []vk.SampleMask{vk.SampleMask(vk.MaxUint32)}
	multisampleState := vk.PipelineMultisampleStateCreateInfo{
		SType:                 vk.StructureTypePipelineMultisampleStateCreateInfo,
		RasterizationSamples:  vk.SampleCount1Bit,
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
		CullMode:                vk.CullModeFlags(vk.CullModeBackBit),
		//CullMode:                vk.CullModeFlags(vk.CullModeNone),
		FrontFace:               vk.FrontFaceClockwise,
		DepthBiasEnable:         vk.False,
		DepthBiasConstantFactor: 0,
		DepthBiasClamp:          0,
		DepthBiasSlopeFactor:    0,
		LineWidth:               1,
	}
	if state.invertFrontFace {
		rasterState.FrontFace = vk.FrontFaceCounterClockwise
	}
	if state.doubleSided {
		rasterState.CullMode = vk.CullModeFlags(vk.CullModeNone)
	}
	depthStencilState := vk.PipelineDepthStencilStateCreateInfo{
		SType:                 vk.StructureTypePipelineDepthStencilStateCreateInfo,
		DepthTestEnable:       vk.True,
		DepthWriteEnable:      vk.True,
		DepthBoundsTestEnable: vk.False,
		StencilTestEnable:     vk.False,
		DepthCompareOp:        vk.CompareOpLess,
	}
	MapBlendEquation := func(i BlendEquation) vk.BlendOp {
		var BlendEquationLUT = map[BlendEquation]vk.BlendOp{
			BlendAdd:             vk.BlendOpAdd,
			BlendReverseSubtract: vk.BlendOpReverseSubtract,
		}
		return BlendEquationLUT[i]
	}

	MapBlendFactor := func(i BlendFunc) vk.BlendFactor {
		var MapBlendFactor = map[BlendFunc]vk.BlendFactor{
			BlendOne:              vk.BlendFactorOne,
			BlendZero:             vk.BlendFactorZero,
			BlendSrcAlpha:         vk.BlendFactorSrcAlpha,
			BlendOneMinusSrcAlpha: vk.BlendFactorOneMinusSrcAlpha,
			BlendOneMinusDstColor: vk.BlendFactorOneMinusDstColor,
			BlendDstColor:         vk.BlendFactorDstColor,
		}
		return MapBlendFactor[i]
	}

	blendFunctions := []struct {
		op  vk.BlendOp
		src vk.BlendFactor
		dst vk.BlendFactor
	}{
		{
			op:  MapBlendEquation(BlendAdd),
			src: MapBlendFactor(BlendOne),
			dst: MapBlendFactor(BlendZero),
		},
	}
	attachmentStates := make([]vk.PipelineColorBlendAttachmentState, 0, len(blendFunctions))
	colorBlendStates := make([]vk.PipelineColorBlendStateCreateInfo, 0, len(blendFunctions))
	pipelineCreateInfos := make([]vk.GraphicsPipelineCreateInfo, 0, len(blendFunctions))
	pipelineRenderingCreateInfo := vk.PipelineRenderingCreateInfo{
		SType:                   vk.StructureTypePipelineRenderingCreateInfo,
		ColorAttachmentCount:    0,
		DepthAttachmentFormat:   vk.FormatD32Sfloat,
		StencilAttachmentFormat: vk.FormatUndefined,
	}
	cPipelineRenderingCreateInfo, _ := pipelineRenderingCreateInfo.PassRef()
	defer pipelineRenderingCreateInfo.Free()
	for i := range blendFunctions {
		attachmentStates = append(attachmentStates, vk.PipelineColorBlendAttachmentState{
			ColorWriteMask: vk.ColorComponentFlags(
				vk.ColorComponentRBit | vk.ColorComponentGBit |
					vk.ColorComponentBBit | vk.ColorComponentABit,
			),
			BlendEnable:         vk.True,
			SrcColorBlendFactor: blendFunctions[i].src,
			DstColorBlendFactor: blendFunctions[i].dst,
			ColorBlendOp:        blendFunctions[i].op,
			SrcAlphaBlendFactor: vk.BlendFactorOne,
			DstAlphaBlendFactor: vk.BlendFactorZero,
			AlphaBlendOp:        vk.BlendOpAdd,
		})
		colorBlendStates = append(colorBlendStates, vk.PipelineColorBlendStateCreateInfo{
			SType:           vk.StructureTypePipelineColorBlendStateCreateInfo,
			LogicOpEnable:   vk.False,
			LogicOp:         vk.LogicOpCopy,
			AttachmentCount: 1,
			PAttachments:    []vk.PipelineColorBlendAttachmentState{attachmentStates[i]},
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
			PColorBlendState:    &colorBlendStates[i],
			PDynamicState:       &dynamicState,
			Layout:              r.shadowMapProgram.pipelineLayout,
			PDepthStencilState:  &depthStencilState,
			RenderPass:          nil,
			PNext:               unsafe.Pointer(cPipelineRenderingCreateInfo),
		})
	}
	pipelines := make([]vk.Pipeline, len(pipelineCreateInfos))
	err := vk.Error(vk.CreateGraphicsPipelines(r.device,
		r.pipelineCache, uint32(len(pipelineCreateInfos)), pipelineCreateInfos, nil, pipelines))
	if err != nil {
		err = fmt.Errorf("vk.CreateGraphicsPipelines failed with %s", err)
	}
	r.shadowMapProgram.pipelineIndexMap[*state] = len(r.shadowMapProgram.pipelines)
	r.shadowMapProgram.pipelines = append(r.shadowMapProgram.pipelines, pipelines...)
	return pipelines[0]
}

func (r *Renderer_VK) CreateModelProgram() (*VulkanProgramInfo, error) {
	program := &VulkanProgramInfo{}
	var uniformBufferMemory vk.DeviceMemory
	var err error
	minAlignment := r.minUniformBufferOffsetAlignment
	uniformSize := uint32(MaxI(VulkanModelUniform0Size, MaxI(VulkanModelUniform1Size, VulkanModelUniform2Size)))
	if uniformSize < minAlignment {
		uniformSize = minAlignment
	} else if uniformSize > minAlignment && minAlignment > 0 && uniformSize%minAlignment != 0 {
		uniformSize = (uniformSize/minAlignment + 1) * minAlignment
	}
	program.uniformBuffers = make([]VulkanBuffer, 1)
	program.uniformBuffers[0].size = uintptr(2000 * uniformSize)
	program.uniformSize = uniformSize
	program.uniformBuffers[0].buffer, err = r.CreateBuffer(vk.DeviceSize(program.uniformBuffers[0].size), vk.BufferUsageFlags(vk.BufferUsageTransferSrcBit|vk.BufferUsageUniformBufferBit), (vk.MemoryPropertyHostVisibleBit | vk.MemoryPropertyHostCoherentBit), &uniformBufferMemory)
	if err != nil {
		panic(err)
	}
	program.uniformBuffers[0].bufferMemory = uniformBufferMemory
	var uniformData unsafe.Pointer
	vk.MapMemory(r.device, program.uniformBuffers[0].bufferMemory, 0, vk.DeviceSize(program.uniformBuffers[0].size), 0, &uniformData)
	program.uniformBuffers[0].data = uniformData

	VertShader, err := staticFiles.ReadFile("shaders/model.vert.spv")
	if err != nil {
		return nil, err
	}
	VertShader2 := make([]uint32, len(VertShader)/4)
	vk.Memcopy(unsafe.Pointer((*sliceHeader)(unsafe.Pointer(&VertShader2)).Data), VertShader)
	vertShader, err := r.CreateShader(r.device, VertShader2)
	if err != nil {
		return nil, err
	}
	//defer vk.DestroyShaderModule(r.device, vertShader, nil)

	FragShader, err := staticFiles.ReadFile("shaders/model.frag.spv")
	if err != nil {
		return nil, err
	}
	FragShader2 := make([]uint32, len(FragShader)/4)
	vk.Memcopy(unsafe.Pointer((*sliceHeader)(unsafe.Pointer(&FragShader2)).Data), FragShader)
	fragShader, err := r.CreateShader(r.device, FragShader2)
	if err != nil {
		return nil, err
	}
	program.vertShader = vertShader
	program.fragShader = fragShader
	//defer vk.DestroyShaderModule(r.device, fragShader, nil)
	uniformBindings0 := []vk.DescriptorSetLayoutBinding{
		//VulkanModelProgramUniformBufferObject0
		{
			Binding:            0,
			DescriptorCount:    1,
			DescriptorType:     vk.DescriptorTypeUniformBuffer,
			PImmutableSamplers: nil,
			StageFlags:         vk.ShaderStageFlags(vk.ShaderStageVertexBit | vk.ShaderStageFragmentBit),
		},
		//VulkanModelProgramUniformBufferObject1
		{
			Binding:            1,
			DescriptorCount:    1,
			DescriptorType:     vk.DescriptorTypeUniformBuffer,
			PImmutableSamplers: nil,
			StageFlags:         vk.ShaderStageFlags(vk.ShaderStageFragmentBit),
		},
		//VulkanModelProgramUniformBufferObject2
		{
			Binding:            2,
			DescriptorCount:    1,
			DescriptorType:     vk.DescriptorTypeUniformBuffer,
			PImmutableSamplers: nil,
			StageFlags:         vk.ShaderStageFlags(vk.ShaderStageVertexBit | vk.ShaderStageFragmentBit),
		},
		//jointMatrices
		{
			Binding:            3,
			DescriptorCount:    1,
			DescriptorType:     vk.DescriptorTypeCombinedImageSampler,
			PImmutableSamplers: nil,
			StageFlags:         vk.ShaderStageFlags(vk.ShaderStageVertexBit),
		},
		//morphTargetValues
		{
			Binding:            4,
			DescriptorCount:    1,
			DescriptorType:     vk.DescriptorTypeCombinedImageSampler,
			PImmutableSamplers: nil,
			StageFlags:         vk.ShaderStageFlags(vk.ShaderStageVertexBit),
		},
		//lambertianEnvSampler
		{
			Binding:            5,
			DescriptorCount:    1,
			DescriptorType:     vk.DescriptorTypeCombinedImageSampler,
			PImmutableSamplers: nil,
			StageFlags:         vk.ShaderStageFlags(vk.ShaderStageFragmentBit),
		},
		//GGXEnvSampler
		{
			Binding:            6,
			DescriptorCount:    1,
			DescriptorType:     vk.DescriptorTypeCombinedImageSampler,
			PImmutableSamplers: nil,
			StageFlags:         vk.ShaderStageFlags(vk.ShaderStageFragmentBit),
		},
		//GGXLUT
		{
			Binding:            7,
			DescriptorCount:    1,
			DescriptorType:     vk.DescriptorTypeCombinedImageSampler,
			PImmutableSamplers: nil,
			StageFlags:         vk.ShaderStageFlags(vk.ShaderStageFragmentBit),
		},
		//tex
		{
			Binding:            8,
			DescriptorCount:    1,
			DescriptorType:     vk.DescriptorTypeCombinedImageSampler,
			PImmutableSamplers: nil,
			StageFlags:         vk.ShaderStageFlags(vk.ShaderStageFragmentBit),
		},
		//normalMap
		{
			Binding:            9,
			DescriptorCount:    1,
			DescriptorType:     vk.DescriptorTypeCombinedImageSampler,
			PImmutableSamplers: nil,
			StageFlags:         vk.ShaderStageFlags(vk.ShaderStageFragmentBit),
		},
		//metallicRoughnessMap
		{
			Binding:            10,
			DescriptorCount:    1,
			DescriptorType:     vk.DescriptorTypeCombinedImageSampler,
			PImmutableSamplers: nil,
			StageFlags:         vk.ShaderStageFlags(vk.ShaderStageFragmentBit),
		},
		//ambientOcclusionMap
		{
			Binding:            11,
			DescriptorCount:    1,
			DescriptorType:     vk.DescriptorTypeCombinedImageSampler,
			PImmutableSamplers: nil,
			StageFlags:         vk.ShaderStageFlags(vk.ShaderStageFragmentBit),
		},
		//emissionMap
		{
			Binding:            12,
			DescriptorCount:    1,
			DescriptorType:     vk.DescriptorTypeCombinedImageSampler,
			PImmutableSamplers: nil,
			StageFlags:         vk.ShaderStageFlags(vk.ShaderStageFragmentBit),
		},
		//shadowCubeMap
		{
			Binding:            13,
			DescriptorCount:    1,
			DescriptorType:     vk.DescriptorTypeCombinedImageSampler,
			PImmutableSamplers: nil,
			StageFlags:         vk.ShaderStageFlags(vk.ShaderStageFragmentBit),
		},
	}
	layoutInfo := vk.DescriptorSetLayoutCreateInfo{
		SType:        vk.StructureTypeDescriptorSetLayoutCreateInfo,
		Flags:        vk.DescriptorSetLayoutCreateFlags(vk.DescriptorSetLayoutCreatePushDescriptorBit),
		BindingCount: uint32(len(uniformBindings0)),
		PBindings:    uniformBindings0,
	}
	var descriptorSetLayout vk.DescriptorSetLayout
	vk.CreateDescriptorSetLayout(r.device, &layoutInfo, nil, &descriptorSetLayout)
	program.descriptorSetLayouts = append(program.descriptorSetLayouts, descriptorSetLayout)
	if err != nil {
		err = fmt.Errorf("vk.CreateDescriptorSetLayout failed with %s", err)
		return nil, err
	}
	pipelineLayoutCreateInfo := vk.PipelineLayoutCreateInfo{
		SType:                  vk.StructureTypePipelineLayoutCreateInfo,
		SetLayoutCount:         1,
		PSetLayouts:            program.descriptorSetLayouts,
		PushConstantRangeCount: 0,
		PPushConstantRanges:    nil,
	}
	var pipelineLayout vk.PipelineLayout
	vk.CreatePipelineLayout(r.device, &pipelineLayoutCreateInfo, nil, &pipelineLayout)
	program.pipelines = []vk.Pipeline{}
	program.pipelineLayout = pipelineLayout
	program.pipelineIndexMap = map[interface{}]int{}
	return program, nil
}
func (r *Renderer_VK) GetModelPipeline(state *VulkanPipelineState) vk.Pipeline {
	if index, ok := r.modelProgram.pipelineIndexMap[*state]; ok {
		return r.modelProgram.pipelines[index]
	}
	vertSpecializationMapEntries := []vk.SpecializationMapEntry{
		{
			ConstantID: 0,
			Size:       4,
			Offset:     0,
		},
		{
			ConstantID: 1,
			Size:       4,
			Offset:     4,
		},
		{
			ConstantID: 2,
			Size:       4,
			Offset:     8,
		},
		{
			ConstantID: 3,
			Size:       4,
			Offset:     12,
		},
		{
			ConstantID: 4,
			Size:       4,
			Offset:     16,
		},
		{
			ConstantID: 5,
			Size:       4,
			Offset:     20,
		},
	}
	fragSpecializationMapEntries := []vk.SpecializationMapEntry{
		{
			ConstantID: 6,
			Size:       4,
			Offset:     0,
		},
		{
			ConstantID: 7,
			Size:       4,
			Offset:     4,
		},
		{
			ConstantID: 8,
			Size:       4,
			Offset:     8,
		},
		{
			ConstantID: 9,
			Size:       4,
			Offset:     12,
		},
		{
			ConstantID: 10,
			Size:       4,
			Offset:     16,
		},
	}
	vertSpecializationInfo := []vk.SpecializationInfo{
		{
			DataSize:      uint64(unsafe.Sizeof(state.VulkanModelPipelineState.VulkanModelSpecializationConstants0)),
			MapEntryCount: uint32(len(vertSpecializationMapEntries)),
			PMapEntries:   vertSpecializationMapEntries,
			PData:         unsafe.Pointer(&state.VulkanModelPipelineState.VulkanModelSpecializationConstants0),
		},
	}
	fragSpecializationInfo := []vk.SpecializationInfo{
		{
			DataSize:      uint64(unsafe.Sizeof(state.VulkanModelPipelineState.VulkanModelSpecializationConstants1)),
			MapEntryCount: uint32(len(fragSpecializationMapEntries)),
			PMapEntries:   fragSpecializationMapEntries,
			PData:         unsafe.Pointer(&state.VulkanModelPipelineState.VulkanModelSpecializationConstants1),
		},
	}
	shaderStages := []vk.PipelineShaderStageCreateInfo{
		{
			SType:               vk.StructureTypePipelineShaderStageCreateInfo,
			Stage:               vk.ShaderStageVertexBit,
			Module:              r.modelProgram.vertShader,
			PSpecializationInfo: vertSpecializationInfo,
			PName:               "main\x00",
		},
		{
			SType:               vk.StructureTypePipelineShaderStageCreateInfo,
			Stage:               vk.ShaderStageFragmentBit,
			Module:              r.modelProgram.fragShader,
			PSpecializationInfo: fragSpecializationInfo,
			PName:               "main\x00",
		},
	}
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
	topology := map[PrimitiveMode]vk.PrimitiveTopology{
		LINES: vk.PrimitiveTopologyLineList,
		//LINE_LOOP:      vk.PrimitiveTopologyLineList,
		LINE_STRIP:     vk.PrimitiveTopologyLineStrip,
		TRIANGLES:      vk.PrimitiveTopologyTriangleList,
		TRIANGLE_STRIP: vk.PrimitiveTopologyTriangleStrip,
		TRIANGLE_FAN:   vk.PrimitiveTopologyTriangleFan,
	}[state.VulkanModelPipelineState.primitiveMode]
	inputAssemblyState := vk.PipelineInputAssemblyStateCreateInfo{
		SType:                  vk.StructureTypePipelineInputAssemblyStateCreateInfo,
		Topology:               topology,
		PrimitiveRestartEnable: vk.False,
	}
	vertexInputBindings := []vk.VertexInputBindingDescription{
		{
			Binding:   0,
			Stride:    4 * 1, // 4 = sizeof(float32)
			InputRate: vk.VertexInputRateVertex,
		},
		{
			Binding:   1,
			Stride:    4 * 3, // 4 = sizeof(float32)
			InputRate: vk.VertexInputRateVertex,
		},
	}
	vertexInputAttributes := []vk.VertexInputAttributeDescription{
		//vertexId
		{
			Binding:  0,
			Location: 0,
			Format:   vk.FormatR32Sint,
			Offset:   0,
		},
		//position
		{
			Binding:  1,
			Location: 1,
			Format:   vk.FormatR32g32b32Sfloat,
			Offset:   0,
		},
	}
	vertexBindingIndex := uint32(2)
	addVertexAttribute := func(size, location uint32, useAttribute bool) {
		format := vk.FormatR32g32Sfloat
		switch size {
		case 3:
			format = vk.FormatR32g32b32Sfloat
			break
		case 4:
			format = vk.FormatR32g32b32a32Sfloat
		}
		if useAttribute {
			vertexInputBindings = append(vertexInputBindings, vk.VertexInputBindingDescription{
				Binding:   vertexBindingIndex,
				Stride:    4 * size, // 4 = sizeof(float32)
				InputRate: vk.VertexInputRateVertex,
			})
			vertexInputAttributes = append(vertexInputAttributes, vk.VertexInputAttributeDescription{
				Binding:  vertexBindingIndex,
				Location: location,
				Format:   format,
				Offset:   0,
			})
			vertexBindingIndex += 1
		} else {
			vertexInputAttributes = append(vertexInputAttributes, vk.VertexInputAttributeDescription{
				Binding:  0,
				Location: location,
				Format:   format,
				Offset:   0,
			})
		}
	}
	addVertexAttribute(2, 2, state.useUV)
	addVertexAttribute(3, 3, state.useNormal)
	addVertexAttribute(4, 4, state.useTangent)
	addVertexAttribute(4, 5, state.useVertColor)
	addVertexAttribute(4, 6, state.useJoint0)
	addVertexAttribute(4, 7, state.useJoint0)
	addVertexAttribute(4, 8, state.useJoint1)
	addVertexAttribute(4, 9, state.useJoint1)
	addVertexAttribute(4, 10, state.useOutlineAttribute)

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
		CullMode:                vk.CullModeFlags(vk.CullModeBackBit),
		//CullMode:                vk.CullModeFlags(vk.CullModeNone),
		FrontFace:               vk.FrontFaceCounterClockwise,
		DepthBiasEnable:         vk.False,
		DepthBiasConstantFactor: 0,
		DepthBiasClamp:          0,
		DepthBiasSlopeFactor:    0,
		LineWidth:               1,
	}
	if state.VulkanModelPipelineState.invertFrontFace {
		rasterState.FrontFace = vk.FrontFaceClockwise
	}
	if state.VulkanModelPipelineState.doubleSided {
		rasterState.CullMode = vk.CullModeFlags(vk.CullModeNone)
	}
	depthStencilState := vk.PipelineDepthStencilStateCreateInfo{
		SType:                 vk.StructureTypePipelineDepthStencilStateCreateInfo,
		DepthTestEnable:       vk.True,
		DepthWriteEnable:      vk.True,
		DepthBoundsTestEnable: vk.False,
		StencilTestEnable:     vk.False,
		DepthCompareOp:        vk.CompareOpLess,
	}
	if !state.VulkanModelPipelineState.depthTest {
		depthStencilState.DepthTestEnable = vk.False
	}
	if !state.VulkanModelPipelineState.depthMask {
		depthStencilState.DepthWriteEnable = vk.False
	}
	MapBlendEquation := func(i BlendEquation) vk.BlendOp {
		var BlendEquationLUT = map[BlendEquation]vk.BlendOp{
			BlendAdd:             vk.BlendOpAdd,
			BlendReverseSubtract: vk.BlendOpReverseSubtract,
		}
		return BlendEquationLUT[i]
	}

	MapBlendFactor := func(i BlendFunc) vk.BlendFactor {
		var MapBlendFactor = map[BlendFunc]vk.BlendFactor{
			BlendOne:              vk.BlendFactorOne,
			BlendZero:             vk.BlendFactorZero,
			BlendSrcAlpha:         vk.BlendFactorSrcAlpha,
			BlendOneMinusSrcAlpha: vk.BlendFactorOneMinusSrcAlpha,
			BlendOneMinusDstColor: vk.BlendFactorOneMinusDstColor,
			BlendDstColor:         vk.BlendFactorDstColor,
		}
		return MapBlendFactor[i]
	}

	blendFunctions := []struct {
		op  vk.BlendOp
		src vk.BlendFactor
		dst vk.BlendFactor
	}{
		{
			op:  MapBlendEquation(r.VKState.VulkanBlendState.op),
			src: MapBlendFactor(r.VKState.VulkanBlendState.src),
			dst: MapBlendFactor(r.VKState.VulkanBlendState.dst),
		},
	}
	attachmentStates := make([]vk.PipelineColorBlendAttachmentState, 0, len(blendFunctions))
	colorBlendStates := make([]vk.PipelineColorBlendStateCreateInfo, 0, len(blendFunctions))
	pipelineCreateInfos := make([]vk.GraphicsPipelineCreateInfo, 0, len(blendFunctions))
	for i := range blendFunctions {
		attachmentStates = append(attachmentStates, vk.PipelineColorBlendAttachmentState{
			ColorWriteMask: vk.ColorComponentFlags(
				vk.ColorComponentRBit | vk.ColorComponentGBit |
					vk.ColorComponentBBit | vk.ColorComponentABit,
			),
			BlendEnable:         vk.True,
			SrcColorBlendFactor: blendFunctions[i].src,
			DstColorBlendFactor: blendFunctions[i].dst,
			ColorBlendOp:        blendFunctions[i].op,
			SrcAlphaBlendFactor: vk.BlendFactorOne,
			DstAlphaBlendFactor: vk.BlendFactorZero,
			AlphaBlendOp:        vk.BlendOpAdd,
		})
		colorBlendStates = append(colorBlendStates, vk.PipelineColorBlendStateCreateInfo{
			SType:           vk.StructureTypePipelineColorBlendStateCreateInfo,
			LogicOpEnable:   vk.False,
			LogicOp:         vk.LogicOpCopy,
			AttachmentCount: 1,
			PAttachments:    []vk.PipelineColorBlendAttachmentState{attachmentStates[i]},
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
			PColorBlendState:    &colorBlendStates[i],
			PDynamicState:       &dynamicState,
			Layout:              r.modelProgram.pipelineLayout,
			RenderPass:          r.mainRenderPass.renderPass,
			PDepthStencilState:  &depthStencilState,
		})
	}
	pipelines := make([]vk.Pipeline, len(pipelineCreateInfos))
	err := vk.Error(vk.CreateGraphicsPipelines(r.device,
		r.pipelineCache, uint32(len(pipelineCreateInfos)), pipelineCreateInfos, nil, pipelines))
	if err != nil {
		err = fmt.Errorf("vk.CreateGraphicsPipelines failed with %s", err)
	}
	r.modelProgram.pipelineIndexMap[*state] = len(r.modelProgram.pipelines)
	r.modelProgram.pipelines = append(r.modelProgram.pipelines, pipelines...)
	return pipelines[0]
}
func (r *Renderer_VK) CreateFullScreenShaderProgram(externalShaders [][][]byte) (*VulkanProgramInfo, error) {
	program := &VulkanProgramInfo{}
	var err error

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
	vk.CreateDescriptorSetLayout(r.device, &layoutInfo, nil, &descriptorSetLayout)
	program.descriptorSetLayouts = append(program.descriptorSetLayouts, descriptorSetLayout)
	if err != nil {
		err = fmt.Errorf("vk.CreateDescriptorSetLayout failed with %s", err)
		return nil, err
	}
	pushConstantRanges := []vk.PushConstantRange{
		{
			StageFlags: vk.ShaderStageFlags(vk.ShaderStageVertexBit | vk.ShaderStageFragmentBit),
			Offset:     0,
			Size:       4 * 3,
		},
	}
	pipelineLayoutCreateInfo := vk.PipelineLayoutCreateInfo{
		SType:                  vk.StructureTypePipelineLayoutCreateInfo,
		SetLayoutCount:         1,
		PSetLayouts:            []vk.DescriptorSetLayout{descriptorSetLayout},
		PushConstantRangeCount: 1,
		PPushConstantRanges:    pushConstantRanges,
	}
	var pipelineLayout vk.PipelineLayout
	vk.CreatePipelineLayout(r.device, &pipelineLayoutCreateInfo, nil, &pipelineLayout)
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
		Topology:               vk.PrimitiveTopologyTriangleStrip,
		PrimitiveRestartEnable: vk.False,
	}

	vertexInputBindings := []vk.VertexInputBindingDescription{{
		Binding:   0,
		Stride:    2 * 4, // 4 = sizeof(float32)
		InputRate: vk.VertexInputRateVertex,
	}}
	vertexInputAttributes := []vk.VertexInputAttributeDescription{{
		Binding:  0,
		Location: 0,
		Format:   vk.FormatR32g32Sfloat,
		Offset:   0,
	}}
	vertexInputState := vk.PipelineVertexInputStateCreateInfo{
		SType:                           vk.StructureTypePipelineVertexInputStateCreateInfo,
		VertexBindingDescriptionCount:   uint32(len(vertexInputBindings)),
		PVertexBindingDescriptions:      vertexInputBindings,
		VertexAttributeDescriptionCount: uint32(len(vertexInputAttributes)),
		PVertexAttributeDescriptions:    vertexInputAttributes,
	}
	sampleMask := []vk.SampleMask{vk.SampleMask(vk.MaxUint32)}
	multisampleState := vk.PipelineMultisampleStateCreateInfo{
		SType:                 vk.StructureTypePipelineMultisampleStateCreateInfo,
		RasterizationSamples:  vk.SampleCount1Bit,
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
	var pipelineCreateInfos []vk.GraphicsPipelineCreateInfo
	attachmentState := vk.PipelineColorBlendAttachmentState{
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
	}
	colorBlendState := vk.PipelineColorBlendStateCreateInfo{
		SType:           vk.StructureTypePipelineColorBlendStateCreateInfo,
		LogicOpEnable:   vk.False,
		LogicOp:         vk.LogicOpCopy,
		AttachmentCount: 1,
		PAttachments:    []vk.PipelineColorBlendAttachmentState{attachmentState},
	}
	if len(externalShaders) == 0 {
		pipelineCreateInfos = make([]vk.GraphicsPipelineCreateInfo, 0, 1)
	} else {
		pipelineCreateInfos = make([]vk.GraphicsPipelineCreateInfo, 0, len(externalShaders[0])+1)
		for i := range externalShaders[0] {
			VertShader2 := make([]uint32, len(externalShaders[0][i])/4)
			vk.Memcopy(unsafe.Pointer((*sliceHeader)(unsafe.Pointer(&VertShader2)).Data), externalShaders[0][i])
			vertShader, err := r.CreateShader(r.device, VertShader2)
			if err != nil {
				return nil, err
			}
			defer vk.DestroyShaderModule(r.device, vertShader, nil)

			FragShader2 := make([]uint32, len(externalShaders[1][i])/4)
			vk.Memcopy(unsafe.Pointer((*sliceHeader)(unsafe.Pointer(&FragShader2)).Data), externalShaders[1][i])
			fragShader, err := r.CreateShader(r.device, FragShader2)
			if err != nil {
				return nil, err
			}
			defer vk.DestroyShaderModule(r.device, fragShader, nil)
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
			pipelineCreateInfos = append(pipelineCreateInfos, vk.GraphicsPipelineCreateInfo{
				SType:               vk.StructureTypeGraphicsPipelineCreateInfo,
				StageCount:          2, // vert + frag
				PStages:             shaderStages,
				PVertexInputState:   &vertexInputState,
				PInputAssemblyState: &inputAssemblyState,
				PViewportState:      &viewportState,
				PRasterizationState: &rasterState,
				PMultisampleState:   &multisampleState,
				PColorBlendState:    &colorBlendState,
				PDynamicState:       &dynamicState,
				Layout:              program.pipelineLayout,
				RenderPass:          r.swapchainRenderPass.renderPass,
				PDepthStencilState:  nil,
			})
		}
	}
	VertShader, err := staticFiles.ReadFile("shaders/ident.vert.spv")
	if err != nil {
		return nil, err
	}
	VertShader2 := make([]uint32, len(VertShader)/4)
	vk.Memcopy(unsafe.Pointer((*sliceHeader)(unsafe.Pointer(&VertShader2)).Data), VertShader)
	vertShader, err := r.CreateShader(r.device, VertShader2)
	if err != nil {
		return nil, err
	}
	defer vk.DestroyShaderModule(r.device, vertShader, nil)

	FragShader, err := staticFiles.ReadFile("shaders/ident.frag.spv")
	if err != nil {
		return nil, err
	}
	FragShader2 := make([]uint32, len(FragShader)/4)
	vk.Memcopy(unsafe.Pointer((*sliceHeader)(unsafe.Pointer(&FragShader2)).Data), FragShader)
	fragShader, err := r.CreateShader(r.device, FragShader2)
	if err != nil {
		return nil, err
	}
	defer vk.DestroyShaderModule(r.device, fragShader, nil)

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
	pipelineCreateInfos = append(pipelineCreateInfos, vk.GraphicsPipelineCreateInfo{
		SType:               vk.StructureTypeGraphicsPipelineCreateInfo,
		StageCount:          2, // vert + frag
		PStages:             shaderStages,
		PVertexInputState:   &vertexInputState,
		PInputAssemblyState: &inputAssemblyState,
		PViewportState:      &viewportState,
		PRasterizationState: &rasterState,
		PMultisampleState:   &multisampleState,
		PColorBlendState:    &colorBlendState,
		PDynamicState:       &dynamicState,
		Layout:              program.pipelineLayout,
		RenderPass:          r.swapchainRenderPass.renderPass,
		PDepthStencilState:  nil,
	})
	pipelines := make([]vk.Pipeline, len(pipelineCreateInfos))
	err = vk.Error(vk.CreateGraphicsPipelines(r.device,
		r.pipelineCache, uint32(len(pipelineCreateInfos)), pipelineCreateInfos, nil, pipelines))
	if err != nil {
		err = fmt.Errorf("vk.CreateGraphicsPipelines failed with %s", err)
		return nil, err
	}
	program.pipelines = pipelines
	return program, nil
}

func (r *Renderer_VK) CreatePanoramaToCubeMapProgram() (*VulkanProgramInfo, error) {
	program := &VulkanProgramInfo{}
	var err error
	VertShader, err := staticFiles.ReadFile("shaders/ident.vert.spv")
	if err != nil {
		return nil, err
	}
	VertShader2 := make([]uint32, len(VertShader)/4)
	vk.Memcopy(unsafe.Pointer((*sliceHeader)(unsafe.Pointer(&VertShader2)).Data), VertShader)
	vertShader, err := r.CreateShader(r.device, VertShader2)
	if err != nil {
		return nil, err
	}
	defer vk.DestroyShaderModule(r.device, vertShader, nil)

	FragShader, err := staticFiles.ReadFile("shaders/panoramaToCubeMap.frag.spv")
	if err != nil {
		return nil, err
	}
	FragShader2 := make([]uint32, len(FragShader)/4)
	vk.Memcopy(unsafe.Pointer((*sliceHeader)(unsafe.Pointer(&FragShader2)).Data), FragShader)
	fragShader, err := r.CreateShader(r.device, FragShader2)
	if err != nil {
		return nil, err
	}
	defer vk.DestroyShaderModule(r.device, fragShader, nil)
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
		Flags:        vk.DescriptorSetLayoutCreateFlags(vk.DescriptorSetLayoutCreatePushDescriptorBit),
		BindingCount: 1,
		PBindings:    samplerLayoutBinding,
	}
	var descriptorSetLayout vk.DescriptorSetLayout
	vk.CreateDescriptorSetLayout(r.device, &layoutInfo, nil, &descriptorSetLayout)
	program.descriptorSetLayouts = append(program.descriptorSetLayouts, descriptorSetLayout)
	if err != nil {
		err = fmt.Errorf("vk.CreateDescriptorSetLayout failed with %s", err)
		return nil, err
	}
	pipelineLayoutCreateInfo := vk.PipelineLayoutCreateInfo{
		SType:                  vk.StructureTypePipelineLayoutCreateInfo,
		SetLayoutCount:         1,
		PSetLayouts:            []vk.DescriptorSetLayout{descriptorSetLayout},
		PushConstantRangeCount: 0,
		PPushConstantRanges:    nil,
	}
	var pipelineLayout vk.PipelineLayout
	vk.CreatePipelineLayout(r.device, &pipelineLayoutCreateInfo, nil, &pipelineLayout)
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
		Topology:               vk.PrimitiveTopologyTriangleStrip,
		PrimitiveRestartEnable: vk.False,
	}

	vertexInputBindings := []vk.VertexInputBindingDescription{{
		Binding:   0,
		Stride:    2 * 4, // 4 = sizeof(float32)
		InputRate: vk.VertexInputRateVertex,
	}}
	vertexInputAttributes := []vk.VertexInputAttributeDescription{{
		Binding:  0,
		Location: 0,
		Format:   vk.FormatR32g32Sfloat,
		Offset:   0,
	}}
	vertexInputState := vk.PipelineVertexInputStateCreateInfo{
		SType:                           vk.StructureTypePipelineVertexInputStateCreateInfo,
		VertexBindingDescriptionCount:   uint32(len(vertexInputBindings)),
		PVertexBindingDescriptions:      vertexInputBindings,
		VertexAttributeDescriptionCount: uint32(len(vertexInputAttributes)),
		PVertexAttributeDescriptions:    vertexInputAttributes,
	}
	sampleMask := []vk.SampleMask{vk.SampleMask(vk.MaxUint32)}
	multisampleState := vk.PipelineMultisampleStateCreateInfo{
		SType:                 vk.StructureTypePipelineMultisampleStateCreateInfo,
		RasterizationSamples:  vk.SampleCount1Bit,
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
	blendFunctions := []struct {
		op  vk.BlendOp
		src vk.BlendFactor
		dst vk.BlendFactor
	}{
		{
			op:  vk.BlendOpAdd,
			src: vk.BlendFactorOne,
			dst: vk.BlendFactorZero,
		},
	}
	attachmentStates := make([]vk.PipelineColorBlendAttachmentState, 0, len(blendFunctions))
	colorBlendStates := make([]vk.PipelineColorBlendStateCreateInfo, 0, len(blendFunctions))
	pipelineCreateInfos := make([]vk.GraphicsPipelineCreateInfo, 0, len(blendFunctions))
	pipelineIndexMap := map[interface{}]int{}
	MapBlendEquation := func(i vk.BlendOp) BlendEquation {
		var BlendEquationLUT = map[vk.BlendOp]BlendEquation{
			vk.BlendOpAdd:             BlendAdd,
			vk.BlendOpReverseSubtract: BlendReverseSubtract,
		}
		return BlendEquationLUT[i]
	}

	MapBlendFactor := func(i vk.BlendFactor) BlendFunc {
		var MapBlendFactor = map[vk.BlendFactor]BlendFunc{
			vk.BlendFactorOne:              BlendOne,
			vk.BlendFactorZero:             BlendZero,
			vk.BlendFactorSrcAlpha:         BlendSrcAlpha,
			vk.BlendFactorOneMinusSrcAlpha: BlendOneMinusSrcAlpha,
			vk.BlendFactorOneMinusDstColor: BlendOneMinusDstColor,
			vk.BlendFactorDstColor:         BlendDstColor,
		}
		return MapBlendFactor[i]
	}
	pipelineRenderingCreateInfo := vk.PipelineRenderingCreateInfo{
		SType:                   vk.StructureTypePipelineRenderingCreateInfo,
		ViewMask:                0b00111111,
		ColorAttachmentCount:    1,
		PColorAttachmentFormats: []vk.Format{vk.FormatR32g32b32a32Sfloat},
		DepthAttachmentFormat:   vk.FormatUndefined,
		StencilAttachmentFormat: vk.FormatUndefined,
	}
	cPipelineRenderingCreateInfo, _ := pipelineRenderingCreateInfo.PassRef()
	for i := range blendFunctions {
		attachmentStates = append(attachmentStates, vk.PipelineColorBlendAttachmentState{
			ColorWriteMask: vk.ColorComponentFlags(
				vk.ColorComponentRBit | vk.ColorComponentGBit |
					vk.ColorComponentBBit | vk.ColorComponentABit,
			),
			BlendEnable:         vk.True,
			SrcColorBlendFactor: blendFunctions[i].src,
			DstColorBlendFactor: blendFunctions[i].dst,
			ColorBlendOp:        blendFunctions[i].op,
			SrcAlphaBlendFactor: vk.BlendFactorOne,
			DstAlphaBlendFactor: vk.BlendFactorZero,
			AlphaBlendOp:        vk.BlendOpAdd,
		})
		colorBlendStates = append(colorBlendStates, vk.PipelineColorBlendStateCreateInfo{
			SType:           vk.StructureTypePipelineColorBlendStateCreateInfo,
			LogicOpEnable:   vk.False,
			LogicOp:         vk.LogicOpCopy,
			AttachmentCount: 1,
			PAttachments:    []vk.PipelineColorBlendAttachmentState{attachmentStates[i]},
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
			PColorBlendState:    &colorBlendStates[i],
			PDynamicState:       &dynamicState,
			Layout:              program.pipelineLayout,
			RenderPass:          nil,
			PDepthStencilState:  nil,
			PNext:               unsafe.Pointer(cPipelineRenderingCreateInfo),
		})
		pipelineIndexMap[VulkanBlendState{
			op:  MapBlendEquation(blendFunctions[i].op),
			src: MapBlendFactor(blendFunctions[i].src),
			dst: MapBlendFactor(blendFunctions[i].dst),
		}] = i
	}
	pipelines := make([]vk.Pipeline, len(pipelineCreateInfos))
	err = vk.Error(vk.CreateGraphicsPipelines(r.device,
		r.pipelineCache, uint32(len(pipelineCreateInfos)), pipelineCreateInfos, nil, pipelines))
	if err != nil {
		err = fmt.Errorf("vk.CreateGraphicsPipelines failed with %s", err)
		return nil, err
	}
	program.pipelines = pipelines
	program.pipelineIndexMap = pipelineIndexMap
	return program, nil
}
func (r *Renderer_VK) CreateCubemapFilteringProgram() (*VulkanProgramInfo, error) {
	program := &VulkanProgramInfo{}
	var err error
	VertShader, err := staticFiles.ReadFile("shaders/ident.vert.spv")
	if err != nil {
		return nil, err
	}
	VertShader2 := make([]uint32, len(VertShader)/4)
	vk.Memcopy(unsafe.Pointer((*sliceHeader)(unsafe.Pointer(&VertShader2)).Data), VertShader)
	vertShader, err := r.CreateShader(r.device, VertShader2)
	if err != nil {
		return nil, err
	}
	defer vk.DestroyShaderModule(r.device, vertShader, nil)

	FragShader, err := staticFiles.ReadFile("shaders/cubemapFiltering.frag.spv")
	if err != nil {
		return nil, err
	}
	FragShader2 := make([]uint32, len(FragShader)/4)
	vk.Memcopy(unsafe.Pointer((*sliceHeader)(unsafe.Pointer(&FragShader2)).Data), FragShader)
	fragShader, err := r.CreateShader(r.device, FragShader2)
	if err != nil {
		return nil, err
	}
	defer vk.DestroyShaderModule(r.device, fragShader, nil)
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
		Flags:        vk.DescriptorSetLayoutCreateFlags(vk.DescriptorSetLayoutCreatePushDescriptorBit),
		BindingCount: 1,
		PBindings:    samplerLayoutBinding,
	}
	var descriptorSetLayout vk.DescriptorSetLayout
	vk.CreateDescriptorSetLayout(r.device, &layoutInfo, nil, &descriptorSetLayout)
	program.descriptorSetLayouts = append(program.descriptorSetLayouts, descriptorSetLayout)
	if err != nil {
		err = fmt.Errorf("vk.CreateDescriptorSetLayout failed with %s", err)
		return nil, err
	}

	pushConstantRanges := []vk.PushConstantRange{
		{
			StageFlags: vk.ShaderStageFlags(vk.ShaderStageFragmentBit),
			Offset:     0,
			Size:       4 * 6,
		},
	}
	pipelineLayoutCreateInfo := vk.PipelineLayoutCreateInfo{
		SType:                  vk.StructureTypePipelineLayoutCreateInfo,
		SetLayoutCount:         1,
		PSetLayouts:            []vk.DescriptorSetLayout{descriptorSetLayout},
		PushConstantRangeCount: uint32(len(pushConstantRanges)),
		PPushConstantRanges:    pushConstantRanges,
	}
	var pipelineLayout vk.PipelineLayout
	vk.CreatePipelineLayout(r.device, &pipelineLayoutCreateInfo, nil, &pipelineLayout)
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
		Topology:               vk.PrimitiveTopologyTriangleStrip,
		PrimitiveRestartEnable: vk.False,
	}

	vertexInputBindings := []vk.VertexInputBindingDescription{{
		Binding:   0,
		Stride:    2 * 4, // 4 = sizeof(float32)
		InputRate: vk.VertexInputRateVertex,
	}}
	vertexInputAttributes := []vk.VertexInputAttributeDescription{{
		Binding:  0,
		Location: 0,
		Format:   vk.FormatR32g32Sfloat,
		Offset:   0,
	}}
	vertexInputState := vk.PipelineVertexInputStateCreateInfo{
		SType:                           vk.StructureTypePipelineVertexInputStateCreateInfo,
		VertexBindingDescriptionCount:   uint32(len(vertexInputBindings)),
		PVertexBindingDescriptions:      vertexInputBindings,
		VertexAttributeDescriptionCount: uint32(len(vertexInputAttributes)),
		PVertexAttributeDescriptions:    vertexInputAttributes,
	}
	sampleMask := []vk.SampleMask{vk.SampleMask(vk.MaxUint32)}
	multisampleState := vk.PipelineMultisampleStateCreateInfo{
		SType:                 vk.StructureTypePipelineMultisampleStateCreateInfo,
		RasterizationSamples:  vk.SampleCount1Bit,
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
	blendFunctions := []struct {
		op  vk.BlendOp
		src vk.BlendFactor
		dst vk.BlendFactor
	}{
		{
			op:  vk.BlendOpAdd,
			src: vk.BlendFactorOne,
			dst: vk.BlendFactorZero,
		},
	}
	attachmentStates := make([]vk.PipelineColorBlendAttachmentState, 0, len(blendFunctions))
	colorBlendStates := make([]vk.PipelineColorBlendStateCreateInfo, 0, len(blendFunctions))
	pipelineCreateInfos := make([]vk.GraphicsPipelineCreateInfo, 0, len(blendFunctions))
	pipelineIndexMap := map[interface{}]int{}
	MapBlendEquation := func(i vk.BlendOp) BlendEquation {
		var BlendEquationLUT = map[vk.BlendOp]BlendEquation{
			vk.BlendOpAdd:             BlendAdd,
			vk.BlendOpReverseSubtract: BlendReverseSubtract,
		}
		return BlendEquationLUT[i]
	}

	MapBlendFactor := func(i vk.BlendFactor) BlendFunc {
		var MapBlendFactor = map[vk.BlendFactor]BlendFunc{
			vk.BlendFactorOne:              BlendOne,
			vk.BlendFactorZero:             BlendZero,
			vk.BlendFactorSrcAlpha:         BlendSrcAlpha,
			vk.BlendFactorOneMinusSrcAlpha: BlendOneMinusSrcAlpha,
			vk.BlendFactorOneMinusDstColor: BlendOneMinusDstColor,
			vk.BlendFactorDstColor:         BlendDstColor,
		}
		return MapBlendFactor[i]
	}
	pipelineRenderingCreateInfo := vk.PipelineRenderingCreateInfo{
		SType:                   vk.StructureTypePipelineRenderingCreateInfo,
		ViewMask:                0b00111111,
		ColorAttachmentCount:    1,
		PColorAttachmentFormats: []vk.Format{vk.FormatR32g32b32a32Sfloat},
		DepthAttachmentFormat:   vk.FormatUndefined,
		StencilAttachmentFormat: vk.FormatUndefined,
	}
	cPipelineRenderingCreateInfo, _ := pipelineRenderingCreateInfo.PassRef()
	for i := range blendFunctions {
		attachmentStates = append(attachmentStates, vk.PipelineColorBlendAttachmentState{
			ColorWriteMask: vk.ColorComponentFlags(
				vk.ColorComponentRBit | vk.ColorComponentGBit |
					vk.ColorComponentBBit | vk.ColorComponentABit,
			),
			BlendEnable:         vk.True,
			SrcColorBlendFactor: blendFunctions[i].src,
			DstColorBlendFactor: blendFunctions[i].dst,
			ColorBlendOp:        blendFunctions[i].op,
			SrcAlphaBlendFactor: vk.BlendFactorOne,
			DstAlphaBlendFactor: vk.BlendFactorZero,
			AlphaBlendOp:        vk.BlendOpAdd,
		})
		colorBlendStates = append(colorBlendStates, vk.PipelineColorBlendStateCreateInfo{
			SType:           vk.StructureTypePipelineColorBlendStateCreateInfo,
			LogicOpEnable:   vk.False,
			LogicOp:         vk.LogicOpCopy,
			AttachmentCount: 1,
			PAttachments:    []vk.PipelineColorBlendAttachmentState{attachmentStates[i]},
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
			PColorBlendState:    &colorBlendStates[i],
			PDynamicState:       &dynamicState,
			Layout:              program.pipelineLayout,
			RenderPass:          nil,
			PDepthStencilState:  nil,
			PNext:               unsafe.Pointer(cPipelineRenderingCreateInfo),
		})
		pipelineIndexMap[VulkanBlendState{
			op:  MapBlendEquation(blendFunctions[i].op),
			src: MapBlendFactor(blendFunctions[i].src),
			dst: MapBlendFactor(blendFunctions[i].dst),
		}] = i
	}
	pipelines := make([]vk.Pipeline, len(pipelineCreateInfos))
	err = vk.Error(vk.CreateGraphicsPipelines(r.device,
		r.pipelineCache, uint32(len(pipelineCreateInfos)), pipelineCreateInfos, nil, pipelines))
	if err != nil {
		err = fmt.Errorf("vk.CreateGraphicsPipelines failed with %s", err)
		return nil, err
	}
	program.pipelines = pipelines
	program.pipelineIndexMap = pipelineIndexMap
	return program, nil
}

func (r *Renderer_VK) CreateLutProgram() (*VulkanProgramInfo, error) {
	program := &VulkanProgramInfo{}
	var err error
	VertShader, err := staticFiles.ReadFile("shaders/ident.vert.spv")
	if err != nil {
		return nil, err
	}
	VertShader2 := make([]uint32, len(VertShader)/4)
	vk.Memcopy(unsafe.Pointer((*sliceHeader)(unsafe.Pointer(&VertShader2)).Data), VertShader)
	vertShader, err := r.CreateShader(r.device, VertShader2)
	if err != nil {
		return nil, err
	}
	defer vk.DestroyShaderModule(r.device, vertShader, nil)

	FragShader, err := staticFiles.ReadFile("shaders/cubemapFiltering.frag.spv")
	if err != nil {
		return nil, err
	}
	FragShader2 := make([]uint32, len(FragShader)/4)
	vk.Memcopy(unsafe.Pointer((*sliceHeader)(unsafe.Pointer(&FragShader2)).Data), FragShader)
	fragShader, err := r.CreateShader(r.device, FragShader2)
	if err != nil {
		return nil, err
	}
	defer vk.DestroyShaderModule(r.device, fragShader, nil)
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
		Flags:        vk.DescriptorSetLayoutCreateFlags(vk.DescriptorSetLayoutCreatePushDescriptorBit),
		BindingCount: 1,
		PBindings:    samplerLayoutBinding,
	}
	var descriptorSetLayout vk.DescriptorSetLayout
	vk.CreateDescriptorSetLayout(r.device, &layoutInfo, nil, &descriptorSetLayout)
	program.descriptorSetLayouts = append(program.descriptorSetLayouts, descriptorSetLayout)
	if err != nil {
		err = fmt.Errorf("vk.CreateDescriptorSetLayout failed with %s", err)
		return nil, err
	}

	pushConstantRanges := []vk.PushConstantRange{
		{
			StageFlags: vk.ShaderStageFlags(vk.ShaderStageFragmentBit),
			Offset:     0,
			Size:       4 * 6,
		},
	}
	pipelineLayoutCreateInfo := vk.PipelineLayoutCreateInfo{
		SType:                  vk.StructureTypePipelineLayoutCreateInfo,
		SetLayoutCount:         1,
		PSetLayouts:            []vk.DescriptorSetLayout{descriptorSetLayout},
		PushConstantRangeCount: uint32(len(pushConstantRanges)),
		PPushConstantRanges:    pushConstantRanges,
	}
	var pipelineLayout vk.PipelineLayout
	vk.CreatePipelineLayout(r.device, &pipelineLayoutCreateInfo, nil, &pipelineLayout)
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
		Topology:               vk.PrimitiveTopologyTriangleStrip,
		PrimitiveRestartEnable: vk.False,
	}

	vertexInputBindings := []vk.VertexInputBindingDescription{{
		Binding:   0,
		Stride:    2 * 4, // 4 = sizeof(float32)
		InputRate: vk.VertexInputRateVertex,
	}}
	vertexInputAttributes := []vk.VertexInputAttributeDescription{{
		Binding:  0,
		Location: 0,
		Format:   vk.FormatR32g32Sfloat,
		Offset:   0,
	}}
	vertexInputState := vk.PipelineVertexInputStateCreateInfo{
		SType:                           vk.StructureTypePipelineVertexInputStateCreateInfo,
		VertexBindingDescriptionCount:   uint32(len(vertexInputBindings)),
		PVertexBindingDescriptions:      vertexInputBindings,
		VertexAttributeDescriptionCount: uint32(len(vertexInputAttributes)),
		PVertexAttributeDescriptions:    vertexInputAttributes,
	}
	sampleMask := []vk.SampleMask{vk.SampleMask(vk.MaxUint32)}
	multisampleState := vk.PipelineMultisampleStateCreateInfo{
		SType:                 vk.StructureTypePipelineMultisampleStateCreateInfo,
		RasterizationSamples:  vk.SampleCount1Bit,
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
	blendFunctions := []struct {
		op  vk.BlendOp
		src vk.BlendFactor
		dst vk.BlendFactor
	}{
		{
			op:  vk.BlendOpAdd,
			src: vk.BlendFactorOne,
			dst: vk.BlendFactorZero,
		},
	}
	attachmentStates := make([]vk.PipelineColorBlendAttachmentState, 0, len(blendFunctions))
	colorBlendStates := make([]vk.PipelineColorBlendStateCreateInfo, 0, len(blendFunctions))
	pipelineCreateInfos := make([]vk.GraphicsPipelineCreateInfo, 0, len(blendFunctions))
	pipelineIndexMap := map[interface{}]int{}
	MapBlendEquation := func(i vk.BlendOp) BlendEquation {
		var BlendEquationLUT = map[vk.BlendOp]BlendEquation{
			vk.BlendOpAdd:             BlendAdd,
			vk.BlendOpReverseSubtract: BlendReverseSubtract,
		}
		return BlendEquationLUT[i]
	}

	MapBlendFactor := func(i vk.BlendFactor) BlendFunc {
		var MapBlendFactor = map[vk.BlendFactor]BlendFunc{
			vk.BlendFactorOne:              BlendOne,
			vk.BlendFactorZero:             BlendZero,
			vk.BlendFactorSrcAlpha:         BlendSrcAlpha,
			vk.BlendFactorOneMinusSrcAlpha: BlendOneMinusSrcAlpha,
			vk.BlendFactorOneMinusDstColor: BlendOneMinusDstColor,
			vk.BlendFactorDstColor:         BlendDstColor,
		}
		return MapBlendFactor[i]
	}
	pipelineRenderingCreateInfo := vk.PipelineRenderingCreateInfo{
		SType:                   vk.StructureTypePipelineRenderingCreateInfo,
		ViewMask:                0b00000001,
		ColorAttachmentCount:    1,
		PColorAttachmentFormats: []vk.Format{vk.FormatR32g32b32a32Sfloat},
		DepthAttachmentFormat:   vk.FormatUndefined,
		StencilAttachmentFormat: vk.FormatUndefined,
	}
	cPipelineRenderingCreateInfo, _ := pipelineRenderingCreateInfo.PassRef()
	for i := range blendFunctions {
		attachmentStates = append(attachmentStates, vk.PipelineColorBlendAttachmentState{
			ColorWriteMask: vk.ColorComponentFlags(
				vk.ColorComponentRBit | vk.ColorComponentGBit |
					vk.ColorComponentBBit | vk.ColorComponentABit,
			),
			BlendEnable:         vk.True,
			SrcColorBlendFactor: blendFunctions[i].src,
			DstColorBlendFactor: blendFunctions[i].dst,
			ColorBlendOp:        blendFunctions[i].op,
			SrcAlphaBlendFactor: vk.BlendFactorOne,
			DstAlphaBlendFactor: vk.BlendFactorZero,
			AlphaBlendOp:        vk.BlendOpAdd,
		})
		colorBlendStates = append(colorBlendStates, vk.PipelineColorBlendStateCreateInfo{
			SType:           vk.StructureTypePipelineColorBlendStateCreateInfo,
			LogicOpEnable:   vk.False,
			LogicOp:         vk.LogicOpCopy,
			AttachmentCount: 1,
			PAttachments:    []vk.PipelineColorBlendAttachmentState{attachmentStates[i]},
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
			PColorBlendState:    &colorBlendStates[i],
			PDynamicState:       &dynamicState,
			Layout:              program.pipelineLayout,
			RenderPass:          nil,
			PDepthStencilState:  nil,
			PNext:               unsafe.Pointer(cPipelineRenderingCreateInfo),
		})
		pipelineIndexMap[VulkanBlendState{
			op:  MapBlendEquation(blendFunctions[i].op),
			src: MapBlendFactor(blendFunctions[i].src),
			dst: MapBlendFactor(blendFunctions[i].dst),
		}] = i
	}
	pipelines := make([]vk.Pipeline, len(pipelineCreateInfos))
	err = vk.Error(vk.CreateGraphicsPipelines(r.device,
		r.pipelineCache, uint32(len(pipelineCreateInfos)), pipelineCreateInfos, nil, pipelines))
	if err != nil {
		err = fmt.Errorf("vk.CreateGraphicsPipelines failed with %s", err)
		return nil, err
	}
	program.pipelines = pipelines
	program.pipelineIndexMap = pipelineIndexMap
	return program, nil
}
func (r *Renderer_VK) CreateDescriptorPool() {
	poolSize := []vk.DescriptorPoolSize{
		{
			Type:            vk.DescriptorTypeCombinedImageSampler,
			DescriptorCount: uint32(len(r.renderTargets)),
		},
	}
	poolInfo := vk.DescriptorPoolCreateInfo{
		SType:         vk.StructureTypeDescriptorPoolCreateInfo,
		PoolSizeCount: 1,
		PPoolSizes:    poolSize,
		MaxSets:       uint32(len(r.renderTargets)),
	}
	var descriptorPool vk.DescriptorPool
	err := vk.Error(vk.CreateDescriptorPool(r.device, &poolInfo, nil, &descriptorPool))
	r.descriptorPool = descriptorPool
	if err != nil {
		err = fmt.Errorf("vk.CreateDescriptorPool failed with %s", err)
		panic(err)
	}

	descriptorSets := make([]vk.DescriptorSet, len(r.renderTargets))
	layouts := []vk.DescriptorSetLayout{r.postProcessingProgram.descriptorSetLayouts[0], r.postProcessingProgram.descriptorSetLayouts[0]}
	allocInfo := vk.DescriptorSetAllocateInfo{
		SType:              vk.StructureTypeDescriptorSetAllocateInfo,
		DescriptorPool:     descriptorPool,
		DescriptorSetCount: uint32(len(r.renderTargets)),
		PSetLayouts:        layouts,
	}
	err = vk.Error(vk.AllocateDescriptorSets(r.device, &allocInfo, &descriptorSets[0]))
	if err != nil {
		panic(err)
	}
	imageInfo := [][]vk.DescriptorImageInfo{
		{{
			ImageLayout: vk.ImageLayoutShaderReadOnlyOptimal,
			ImageView:   r.renderTargets[0].texture.imageView,
			Sampler:     r.spriteSamplers[1],
		}},
		{{
			ImageLayout: vk.ImageLayoutShaderReadOnlyOptimal,
			ImageView:   r.renderTargets[1].texture.imageView,
			Sampler:     r.spriteSamplers[1],
		}},
	}
	descriptorWrites := []vk.WriteDescriptorSet{
		{
			SType:           vk.StructureTypeWriteDescriptorSet,
			DstSet:          descriptorSets[0],
			DstBinding:      0,
			DstArrayElement: 0,
			DescriptorCount: 1,
			DescriptorType:  vk.DescriptorTypeCombinedImageSampler,
			PImageInfo:      imageInfo[0],
		},
		{
			SType:           vk.StructureTypeWriteDescriptorSet,
			DstSet:          descriptorSets[1],
			DstBinding:      0,
			DstArrayElement: 0,
			DescriptorCount: 1,
			DescriptorType:  vk.DescriptorTypeCombinedImageSampler,
			PImageInfo:      imageInfo[1],
		},
	}
	vk.UpdateDescriptorSets(r.device, uint32(len(descriptorWrites)), descriptorWrites, 0, nil)
	r.renderTargets[0].descriptorSet = descriptorSets[0]
	r.renderTargets[1].descriptorSet = descriptorSets[1]
}
func (r *Renderer_VK) CreateShader(device vk.Device, data []uint32) (vk.ShaderModule, error) {
	var module vk.ShaderModule

	shaderModuleCreateInfo := vk.ShaderModuleCreateInfo{
		SType:    vk.StructureTypeShaderModuleCreateInfo,
		CodeSize: uint64(len(data) * 4),
		PCode:    data,
	}
	err := vk.Error(vk.CreateShaderModule(device, &shaderModuleCreateInfo, nil, &module))
	if err != nil {
		err = fmt.Errorf("vk.CreateShaderModule failed with %s", err)
		return module, err
	}
	return module, nil
}
func (r *Renderer_VK) CreateSpriteSampler() error {

	interp := vk.FilterNearest
	samplerInfo := vk.SamplerCreateInfo{
		SType:                   vk.StructureTypeSamplerCreateInfo,
		MagFilter:               interp,
		MinFilter:               interp,
		AddressModeU:            vk.SamplerAddressModeClampToEdge,
		AddressModeV:            vk.SamplerAddressModeClampToEdge,
		AddressModeW:            vk.SamplerAddressModeClampToEdge,
		AnisotropyEnable:        vk.False,
		MaxAnisotropy:           0,
		UnnormalizedCoordinates: vk.False,
		CompareEnable:           vk.False,
		CompareOp:               vk.CompareOpAlways,
		MipmapMode:              vk.SamplerMipmapModeLinear,
		MipLodBias:              0,
		MinLod:                  0,
		MaxLod:                  0,
	}
	var sampler vk.Sampler
	err := vk.Error(vk.CreateSampler(r.device, &samplerInfo, nil, &sampler))
	if err != nil {
		err = fmt.Errorf("vk.CreateSampler failed with %s", err)
		return err
	}
	r.spriteSamplers = append(r.spriteSamplers, sampler)
	interp = vk.FilterLinear
	samplerInfo2 := vk.SamplerCreateInfo{
		SType:                   vk.StructureTypeSamplerCreateInfo,
		MagFilter:               interp,
		MinFilter:               interp,
		AddressModeU:            vk.SamplerAddressModeClampToEdge,
		AddressModeV:            vk.SamplerAddressModeClampToEdge,
		AddressModeW:            vk.SamplerAddressModeClampToEdge,
		AnisotropyEnable:        vk.False,
		MaxAnisotropy:           0,
		UnnormalizedCoordinates: vk.False,
		CompareEnable:           vk.False,
		CompareOp:               vk.CompareOpAlways,
		MipmapMode:              vk.SamplerMipmapModeLinear,
		MipLodBias:              0,
		MinLod:                  0,
		MaxLod:                  0,
	}
	var sampler2 vk.Sampler
	err = vk.Error(vk.CreateSampler(r.device, &samplerInfo2, nil, &sampler2))
	if err != nil {
		err = fmt.Errorf("vk.CreateSampler failed with %s", err)
		return err
	}
	r.spriteSamplers = append(r.spriteSamplers, sampler2)
	return nil
}
func (r *Renderer_VK) GetSampler(info VulkanSamplerInfo) vk.Sampler {
	if sampler, ok := r.samplers[info]; ok {
		return sampler
	}
	samplerInfo := vk.SamplerCreateInfo{
		SType:                   vk.StructureTypeSamplerCreateInfo,
		MagFilter:               vk.FilterNearest,
		MinFilter:               vk.FilterNearest,
		AddressModeU:            vk.SamplerAddressModeClampToEdge,
		AddressModeV:            vk.SamplerAddressModeClampToEdge,
		AddressModeW:            vk.SamplerAddressModeClampToEdge,
		AnisotropyEnable:        vk.False,
		MaxAnisotropy:           0,
		UnnormalizedCoordinates: vk.False,
		CompareEnable:           vk.False,
		CompareOp:               vk.CompareOpAlways,
		MipmapMode:              vk.SamplerMipmapModeLinear,
		MipLodBias:              0,
		MinLod:                  0,
		MaxLod:                  vk.LodClampNone,
	}
	if info.magFilter == TextureSamplingFilterLinear {
		samplerInfo.MagFilter = vk.FilterLinear
	}
	if info.minFilter == TextureSamplingFilterLinear || info.minFilter == TextureSamplingFilterLinearMipMapLinear || info.minFilter == TextureSamplingFilterLinearMipMapNearest {
		samplerInfo.MinFilter = vk.FilterLinear
	}
	if info.minFilter == TextureSamplingFilterLinearMipMapNearest || info.minFilter == TextureSamplingFilterNearestMipMapNearest {
		samplerInfo.MipmapMode = vk.SamplerMipmapModeNearest
	}
	if info.minFilter == TextureSamplingFilterLinear || info.minFilter == TextureSamplingFilterNearest {
		samplerInfo.MaxLod = 0
	}
	if info.wrapS == TextureSamplingWrapMirroredRepeat {
		samplerInfo.AddressModeU = vk.SamplerAddressModeMirroredRepeat
	} else if info.wrapS == TextureSamplingWrapRepeat {
		samplerInfo.AddressModeU = vk.SamplerAddressModeRepeat
	}
	if info.wrapT == TextureSamplingWrapMirroredRepeat {
		samplerInfo.AddressModeV = vk.SamplerAddressModeMirroredRepeat
	} else if info.wrapT == TextureSamplingWrapRepeat {
		samplerInfo.AddressModeV = vk.SamplerAddressModeRepeat
	}
	var sampler vk.Sampler
	err := vk.Error(vk.CreateSampler(r.device, &samplerInfo, nil, &sampler))
	if err != nil {
		err = fmt.Errorf("vk.CreateSampler failed with %s", err)
		panic(err)
	}
	r.samplers[info] = sampler
	return sampler
}
func (r *Renderer_VK) CreateSwapChainFramebuffer(swapchain *VulkanSwapchainInfo, renderPass vk.RenderPass) error {
	var swapchainImagesCount uint32
	err := vk.Error(vk.GetSwapchainImages(r.device, swapchain.swapchain, &swapchainImagesCount, nil))
	if err != nil {
		err = fmt.Errorf("vk.GetSwapchainImages failed with %s", err)
		return err
	}
	swapchainImages := make([]vk.Image, swapchainImagesCount)
	vk.GetSwapchainImages(r.device, swapchain.swapchain, &swapchainImagesCount, swapchainImages)
	swapchain.framebuffers = make([]vk.Framebuffer, swapchain.imageCount)
	for i := range swapchain.framebuffers {
		attachments := []vk.ImageView{}
		attachments = append(attachments, swapchain.imageViews[i])
		fbCreateInfo := vk.FramebufferCreateInfo{
			SType:           vk.StructureTypeFramebufferCreateInfo,
			RenderPass:      renderPass,
			Layers:          1,
			AttachmentCount: uint32(len(attachments)), // 2 if has depthView
			PAttachments:    attachments,
			Width:           swapchain.extent.Width,
			Height:          swapchain.extent.Height,
		}
		err := vk.Error(vk.CreateFramebuffer(r.device, &fbCreateInfo, nil, &swapchain.framebuffers[i]))
		if err != nil {
			err = fmt.Errorf("vk.CreateFramebuffer failed with %s", err)
			return err
		}
	}
	return nil
}
func (r *Renderer_VK) CreateCommandPool() error {
	cmdPoolCreateInfo := vk.CommandPoolCreateInfo{
		SType:            vk.StructureTypeCommandPoolCreateInfo,
		Flags:            vk.CommandPoolCreateFlags(vk.CommandPoolCreateResetCommandBufferBit),
		QueueFamilyIndex: 0,
	}
	var commandPool1, commandPool2 vk.CommandPool
	err := vk.Error(vk.CreateCommandPool(r.device, &cmdPoolCreateInfo, nil, &commandPool1))
	if err != nil {
		err = fmt.Errorf("vk.CreateCommandPool failed with %s", err)
		return err
	}
	err = vk.Error(vk.CreateCommandPool(r.device, &cmdPoolCreateInfo, nil, &commandPool2))
	if err != nil {
		err = fmt.Errorf("vk.CreateCommandPool failed with %s", err)
		return err
	}

	r.commandPools = []vk.CommandPool{commandPool1, commandPool2}
	return nil
}
func (r *Renderer_VK) CreateCommandBuffer() error {
	commandBuffer1 := make([]vk.CommandBuffer, 1)
	commandBuffer2 := make([]vk.CommandBuffer, 1)
	cmdBufferAllocateInfo := vk.CommandBufferAllocateInfo{
		SType:              vk.StructureTypeCommandBufferAllocateInfo,
		CommandPool:        r.commandPools[0],
		Level:              vk.CommandBufferLevelPrimary,
		CommandBufferCount: 1,
	}
	err := vk.Error(vk.AllocateCommandBuffers(r.device, &cmdBufferAllocateInfo, commandBuffer1))
	if err != nil {
		err = fmt.Errorf("vk.AllocateCommandBuffers failed with %s", err)
		return err
	}
	cmdBufferAllocateInfo2 := vk.CommandBufferAllocateInfo{
		SType:              vk.StructureTypeCommandBufferAllocateInfo,
		CommandPool:        r.commandPools[1],
		Level:              vk.CommandBufferLevelPrimary,
		CommandBufferCount: 1,
	}
	//cmdBufferAllocateInfo.CommandPool = v.commandPools[1]
	err = vk.Error(vk.AllocateCommandBuffers(r.device, &cmdBufferAllocateInfo2, commandBuffer2))
	if err != nil {
		err = fmt.Errorf("vk.AllocateCommandBuffers failed with %s", err)
		return err
	}
	r.commandBuffers = []vk.CommandBuffer{commandBuffer1[0], commandBuffer2[0]}
	return nil
}
func (r *Renderer_VK) CreateSyncObjects() error {
	fenceCreateInfo := vk.FenceCreateInfo{
		SType: vk.StructureTypeFenceCreateInfo,
	}
	semaphoreCreateInfo := vk.SemaphoreCreateInfo{
		SType: vk.StructureTypeSemaphoreCreateInfo,
	}
	r.fences = make([]vk.Fence, 3)
	err := vk.Error(vk.CreateFence(r.device, &fenceCreateInfo, nil, &r.fences[0]))
	if err != nil {
		err = fmt.Errorf("vk.CreateFence failed with %s", err)
		return err
	}
	err = vk.Error(vk.CreateFence(r.device, &fenceCreateInfo, nil, &r.fences[1]))
	if err != nil {
		err = fmt.Errorf("vk.CreateFence failed with %s", err)
		return err
	}
	err = vk.Error(vk.CreateFence(r.device, &fenceCreateInfo, nil, &r.fences[2]))
	if err != nil {
		err = fmt.Errorf("vk.CreateFence failed with %s", err)
		return err
	}
	r.semaphores = make([]vk.Semaphore, 2)
	err = vk.Error(vk.CreateSemaphore(r.device, &semaphoreCreateInfo, nil, &r.semaphores[0]))
	if err != nil {
		err = fmt.Errorf("vk.CreateSemaphore failed with %s", err)
		return err
	}
	err = vk.Error(vk.CreateSemaphore(r.device, &semaphoreCreateInfo, nil, &r.semaphores[1]))
	if err != nil {
		err = fmt.Errorf("vk.CreateSemaphore failed with %s", err)
		return err
	}
	return nil
}

func (r *Renderer_VK) DestroyPipelines(program *VulkanProgramInfo) {
	for i := range program.uniformBuffers {
		vk.UnmapMemory(r.device, program.uniformBuffers[i].bufferMemory)
		vk.DestroyBuffer(r.device, program.uniformBuffers[i].buffer, nil)
		vk.FreeMemory(r.device, program.uniformBuffers[i].bufferMemory, nil)
	}
	vk.DestroyPipelineLayout(r.device, program.pipelineLayout, nil)
	for i := range program.pipelines {
		vk.DestroyPipeline(r.device, program.pipelines[i], nil)
	}
	if program.vertShader != nil {
		vk.DestroyShaderModule(r.device, program.vertShader, nil)
		vk.DestroyShaderModule(r.device, program.fragShader, nil)
	}

}
func (r *Renderer_VK) DestroySwapchain() {
	for i := range r.swapchains {
		for j := range r.swapchains[i].imageViews {
			vk.DestroyFramebuffer(r.device, r.swapchains[i].framebuffers[j], nil)
			vk.DestroyImageView(r.device, r.swapchains[i].imageViews[j], nil)
		}
		vk.DestroySwapchain(r.device, r.swapchains[i].swapchain, nil)
	}

}
func (r *Renderer_VK) DestroyRenderTarget(renderTarget *VulkanRenderTargetInfo) {
	vk.DestroyFramebuffer(r.device, renderTarget.framebuffer, nil)
	renderTarget.texture = nil
}
func (r *Renderer_VK) Destroy() {
	if r == nil {
		return
	}
	r.gpuDevices = nil
	vk.QueueWaitIdle(r.queue)
	r.shadowMapTextures = nil
	r.dummyTexture = nil
	r.dummyCubeTexture = nil
	for i := range r.palTexture.textures {
		r.palTexture.textures[i] = nil
	}

	//Save Vulkan pipeline cache to disk
	var pipelineCacheSize uint64
	err := vk.Error(vk.GetPipelineCacheData(r.device, r.pipelineCache, &pipelineCacheSize, nil))
	if err != nil {
		err = fmt.Errorf("vk.GetPipelineCacheData failed with %s", err)
		fmt.Println(err)
	} else {
		pipelineCacheData := make([]byte, pipelineCacheSize)
		err = vk.Error(vk.GetPipelineCacheData(r.device, r.pipelineCache, &pipelineCacheSize, unsafe.Pointer(&pipelineCacheData[0])))
		if err != nil {
			err = fmt.Errorf("vk.GetPipelineCacheData failed with %s", err)
			log.Println(err)
		} else {
			_ = os.MkdirAll("./cache/Vulkan", os.ModePerm)
			err = os.WriteFile("./cache/Vulkan/pipeline_cache.bin", pipelineCacheData, 0644)
			if err != nil {
				log.Println("Failed to write pipeline cache to disk:", err)
			}
		}
	}
	vk.DestroyPipelineCache(r.device, r.pipelineCache, nil)
	for i := range r.vertexBuffers {
		vk.UnmapMemory(r.device, r.vertexBuffers[i].bufferMemory)
		vk.DestroyBuffer(r.device, r.vertexBuffers[i].buffer, nil)
		vk.FreeMemory(r.device, r.vertexBuffers[i].bufferMemory, nil)
	}

	vk.UnmapMemory(r.device, r.stagingBuffers[0].bufferMemory)
	vk.DestroyBuffer(r.device, r.stagingBuffers[0].buffer, nil)
	vk.FreeMemory(r.device, r.stagingBuffers[0].bufferMemory, nil)

	vk.UnmapMemory(r.device, r.stagingBuffers[1].bufferMemory)
	vk.DestroyBuffer(r.device, r.stagingBuffers[1].buffer, nil)
	vk.FreeMemory(r.device, r.stagingBuffers[1].bufferMemory, nil)

	for i := range r.modelVertexBuffers {
		vk.DestroyBuffer(r.device, r.modelVertexBuffers[i].buffer, nil)
		vk.FreeMemory(r.device, r.modelVertexBuffers[i].bufferMemory, nil)
		vk.DestroyBuffer(r.device, r.modelIndexBuffers[i].buffer, nil)
		vk.FreeMemory(r.device, r.modelIndexBuffers[i].bufferMemory, nil)
	}

	for sampler := range r.spriteSamplers {
		vk.DestroySampler(r.device, r.spriteSamplers[sampler], nil)
	}
	for key := range r.samplers {
		vk.DestroySampler(r.device, r.samplers[key], nil)
	}
	vk.DestroySemaphore(r.device, r.semaphores[0], nil)
	vk.DestroySemaphore(r.device, r.semaphores[1], nil)
	vk.DestroyFence(r.device, r.fences[0], nil)
	vk.DestroyFence(r.device, r.fences[1], nil)
	vk.DestroyFence(r.device, r.fences[2], nil)
	vk.DestroyCommandPool(r.device, r.commandPools[0], nil)
	vk.DestroyCommandPool(r.device, r.commandPools[1], nil)
	r.DestroyPipelines(r.spriteProgram)
	if r.modelProgram != nil {
		r.DestroyPipelines(r.modelProgram)
		if r.shadowMapProgram != nil {
			r.DestroyPipelines(r.shadowMapProgram)
		}
	}
	r.DestroyPipelines(r.postProcessingProgram)
	r.DestroyPipelines(r.lutProgram)
	r.DestroyPipelines(r.cubemapFilteringProgram)
	r.DestroyPipelines(r.panoramaToCubeMapProgram)
	vk.DestroyRenderPass(r.device, r.swapchainRenderPass.renderPass, nil)
	vk.DestroyRenderPass(r.device, r.mainRenderPass.renderPass, nil)
	r.DestroyRenderTarget(r.mainRenderTarget)
	r.DestroyRenderTarget(r.renderTargets[0])
	r.DestroyRenderTarget(r.renderTargets[1])
	r.DestroySwapchain()

	runtime.GC()
	r.DestroyResources(len(r.destroyResourceQueues[r.destroyResourceQueueIndex]))
	r.destroyResourceQueueIndex = (r.destroyResourceQueueIndex + 1) % 2
	r.DestroyResources(len(r.destroyResourceQueues[r.destroyResourceQueueIndex]))

	vk.DestroySurface(r.instance, r.surface, nil)
	vk.DestroyDevice(r.device, nil)
	vk.DestroyInstance(r.instance, nil)
}
func (r *Renderer_VK) BeginSingleTimeCommands() vk.CommandBuffer {
	commandBuffers := make([]vk.CommandBuffer, 1)
	cmdBufferAllocateInfo := vk.CommandBufferAllocateInfo{
		SType:              vk.StructureTypeCommandBufferAllocateInfo,
		CommandPool:        r.commandPools[0],
		Level:              vk.CommandBufferLevelPrimary,
		CommandBufferCount: 1,
	}
	err := vk.Error(vk.AllocateCommandBuffers(r.device, &cmdBufferAllocateInfo, commandBuffers))
	if err != nil {
		err = fmt.Errorf("vk.AllocateCommandBuffers failed with %s", err)
		panic(err)
	}
	cmdBufferBeginInfo := vk.CommandBufferBeginInfo{
		SType:            vk.StructureTypeCommandBufferBeginInfo,
		Flags:            vk.CommandBufferUsageFlags(vk.CommandBufferUsageOneTimeSubmitBit),
		PInheritanceInfo: nil,
	}
	vk.BeginCommandBuffer(commandBuffers[0], &cmdBufferBeginInfo)
	return commandBuffers[0]
}
func (r *Renderer_VK) EndSingleTimeCommands(commandBuffer vk.CommandBuffer) {
	vk.EndCommandBuffer(commandBuffer)
	commandBuffers := []vk.CommandBuffer{commandBuffer}
	submitInfo := []vk.SubmitInfo{{
		SType:              vk.StructureTypeSubmitInfo,
		CommandBufferCount: 1,
		PCommandBuffers:    commandBuffers,
	}}
	vk.QueueSubmit(r.queue, 1, submitInfo, nil)
	vk.QueueWaitIdle(r.queue)
	vk.FreeCommandBuffers(r.device, r.commandPools[0], 1, commandBuffers)
}

func (r *Renderer_VK) CreateMemoryBuffers() error {
	var vertexBufferMemory vk.DeviceMemory
	var err error
	r.vertexBuffers = make([]VulkanBuffer, 1)
	r.vertexBuffers[0].size = 40000 * 4 * 4
	r.vertexBuffers[0].buffer, err = r.CreateBuffer(vk.DeviceSize(r.vertexBuffers[0].size), vk.BufferUsageFlags(vk.BufferUsageTransferSrcBit|vk.BufferUsageVertexBufferBit), (vk.MemoryPropertyHostVisibleBit | vk.MemoryPropertyHostCoherentBit), &vertexBufferMemory)
	if err != nil {
		return err
	}
	r.vertexBuffers[0].bufferMemory = vertexBufferMemory
	var vertexData unsafe.Pointer
	vk.MapMemory(r.device, r.vertexBuffers[0].bufferMemory, 0, vk.DeviceSize(r.vertexBuffers[0].size), 0, &vertexData)
	r.vertexBuffers[0].data = vertexData

	for i := 0; i < 2; i++ {
		var stagingBufferMemory vk.DeviceMemory
		r.stagingBuffers[i].size = 4096 * 4096 * 4
		r.stagingBuffers[i].buffer, err = r.CreateBuffer(vk.DeviceSize(r.stagingBuffers[i].size), vk.BufferUsageFlags(vk.BufferUsageTransferSrcBit), (vk.MemoryPropertyHostVisibleBit | vk.MemoryPropertyHostCoherentBit), &stagingBufferMemory)
		if err != nil {
			panic(err)
		}
		r.stagingBuffers[i].bufferMemory = stagingBufferMemory
		var stagingData unsafe.Pointer
		vk.MapMemory(r.device, stagingBufferMemory, 0, vk.DeviceSize(r.stagingBuffers[i].size), 0, &stagingData)
		r.stagingBuffers[i].data = stagingData
	}

	r.modelVertexBuffers = make([]VulkanBuffer, 2)
	r.modelIndexBuffers = make([]VulkanBuffer, 2)
	for i := 0; i < 2; i++ {
		r.modelVertexBuffers[i].size = 0
		r.modelIndexBuffers[i].size = 0
	}
	return nil
}

func (r *Renderer_VK) CreatePipelineCache() error {
	pipelineCacheInfo := vk.PipelineCacheCreateInfo{
		SType: vk.StructureTypePipelineCacheCreateInfo,
	}
	cacheFile, err := os.Open("./cache/Vulkan/pipeline_cache.bin")
	if err == nil {
		defer cacheFile.Close()
		cacheData, err := io.ReadAll(cacheFile)
		// go 1.21+ feature
		// pinner := runtime.Pinner{}
		// pinner.Pin(&cacheData[0])
		// defer pinner.Unpin()
		if err == nil {
			pipelineCacheInfo.InitialDataSize = uint64(len(cacheData))
			pipelineCacheInfo.PInitialData = unsafe.Pointer(&cacheData[0])
		} else {
			log.Println("Cannot read Vulkan pipeline cache file, a new one will be created")
		}
	} else {
		log.Println("Cannot open Vulkan pipeline cache file, a new one will be created")
	}
	var cache vk.PipelineCache
	err = vk.Error(vk.CreatePipelineCache(r.device, &pipelineCacheInfo, nil, &cache))
	if err != nil {
		log.Println("vk.CreatePipelineCache failed with %s", err)
		if pipelineCacheInfo.InitialDataSize > 0 {
			log.Println("Creating an empty pipeline cache instead")
			pipelineCacheInfo2 := vk.PipelineCacheCreateInfo{
				SType: vk.StructureTypePipelineCacheCreateInfo,
			}
			err = vk.Error(vk.CreatePipelineCache(r.device, &pipelineCacheInfo2, nil, &cache))
			if err != nil {
				return err
			}
		}
		return nil
	}
	r.pipelineCache = cache
	return nil
}

// Render initialization.
// Creates the default shaders, the framebuffer and enables MSAA.
func (r *Renderer_VK) Init() {
	r.allocatedImageMemory = 0
	r.enableModel = sys.cfg.Video.EnableModel
	r.enableShadow = sys.cfg.Video.EnableModelShadow
	r.memoryTypeMap = make(map[vk.MemoryPropertyFlagBits]uint32)
	r.samplers = map[VulkanSamplerInfo]vk.Sampler{}
	r.stagingBufferFences = [2]bool{false, false}
	r.stagingBufferIndex = 0
	r.stagingBufferOffset = 0
	r.stagingImageCopyRegions = make(map[vk.Image][]vk.BufferImageCopy)
	r.VKState.VulkanModelPipelineState.VulkanModelSpecializationConstants1.useShadowMap = r.enableShadow
	r.setVSync = false
	vk.SetGetInstanceProcAddr(sdl.VulkanGetVkGetInstanceProcAddr())
	err := vk.Init()
	if err != nil {
		panic(err)
	}
	wminfo, _ := sys.window.GetWMInfo()
	var osWindowHandle uintptr

	switch wminfo.Subsystem {
	case sdl.SYSWM_COCOA:
		osWindowHandle = uintptr(wminfo.GetCocoaInfo().Window)
	case sdl.SYSWM_WINDOWS:
		osWindowHandle = uintptr(wminfo.GetWindowsInfo().Window)
	case sdl.SYSWM_X11:
		x11WinID := wminfo.GetX11Info().Window
		osWindowHandle = uintptr(x11WinID)
	}
	err = r.NewVulkanDevice(appInfo, osWindowHandle)
	if err != nil {
		if len(r.gpuDevices) > 0 {
			r.PrintInfo()
		}
		panic(err)
	}
	err = r.CreateSwapchain()
	if err != nil {
		panic(err)
	}
	err = r.CreateMemoryBuffers()
	if err != nil {
		panic(err)
	}
	r.swapchainRenderPass, err = r.CreateSwapchainRenderPass(r.swapchains[0])
	if err != nil {
		panic(err)
	}
	msaa := sys.msaa
	if msaa <= 0 {
		msaa = 1
	}
	r.mainRenderPass, err = r.CreateMainRenderPass(r.swapchains[0], msaa, true)
	if err != nil {
		panic(err)
	}
	r.postProcessingRenderPass, err = r.CreatePostProcessingRenderPass(r.swapchains[0])
	if err != nil {
		panic(err)
	}
	r.CreatePipelineCache()
	r.spriteProgram, err = r.CreateSpriteProgram()
	if err != nil {
		panic(err)
	}

	r.postProcessingProgram, err = r.CreateFullScreenShaderProgram(sys.externalShaders)
	if err != nil {
		panic(err)
	}

	r.panoramaToCubeMapProgram, err = r.CreatePanoramaToCubeMapProgram()
	if err != nil {
		panic(err)
	}

	r.cubemapFilteringProgram, err = r.CreateCubemapFilteringProgram()
	if err != nil {
		panic(err)
	}
	r.lutProgram, err = r.CreateLutProgram()
	if err != nil {
		panic(err)
	}
	err = r.CreateSpriteSampler()
	if err != nil {
		panic(err)
	}
	err = r.CreateCommandPool()
	if err != nil {
		panic(err)
	}
	err = r.CreateCommandBuffer()
	if err != nil {
		panic(err)
	}
	r.renderTargets[0] = r.CreateRenderTarget(r.swapchainRenderPass.renderPass, uint32(sys.scrrect[2]), uint32(sys.scrrect[3]), 1, false)
	r.renderTargets[1] = r.CreateRenderTarget(r.swapchainRenderPass.renderPass, uint32(sys.scrrect[2]), uint32(sys.scrrect[3]), 1, false)
	r.mainRenderTarget = r.CreateRenderTarget(r.mainRenderPass.renderPass, uint32(sys.scrrect[2]), uint32(sys.scrrect[3]), msaa, true)

	err = r.CreateSyncObjects()
	if err != nil {
		panic(err)
	}
	err = r.CreateSwapChainFramebuffer(r.swapchains[0], r.swapchainRenderPass.renderPass)
	if err != nil {
		panic(err)
	}
	r.CreateDescriptorPool()
	r.VKState.scissor = vk.Rect2D{
		Extent: r.swapchains[0].extent,
		Offset: vk.Offset2D{
			X: 0, Y: 0,
		},
	}
	r.dummyTexture = r.newTexture(1, 1, 8, false).(*Texture_VK)
	r.dummyTexture.SetData([]byte{0})
	r.dummyTexture.sampler = r.spriteSamplers[0]

	r.dummyCubeTexture = r.newDummyCubeMapTexture().(*Texture_VK)
	r.dummyCubeTexture.SetCubeMapData([]byte{0})
	r.dummyCubeTexture.sampler = r.spriteSamplers[0]

	r.spriteProgram.uniformOffsetMap = map[interface{}]uint32{}
	r.createPalTexture(2048)
	r.shadowMapTextures = r.createShadowMapTexture(1024)
	if r.enableModel {
		r.modelProgram, err = r.CreateModelProgram()
		if err != nil {
			panic(err)
		}
		if r.enableShadow {
			r.shadowMapProgram, err = r.CreateShadowMapProgram()
			if err != nil {
				panic(err)
			}
		}
	}
	r.destroyResourceQueues[0] = make(chan VulkanResource, 65536)
	r.destroyResourceQueues[1] = make(chan VulkanResource, 65536)
	r.destroyResourceQueueIndex = 0
}

func (r *Renderer_VK) Close() {
	r.Destroy()
}

func (r *Renderer_VK) IsModelEnabled() bool {
	return r.enableModel
}

func (r *Renderer_VK) IsShadowEnabled() bool {
	return r.enableShadow
}

func (r *Renderer_VK) DestroyResources(queueLength int) {
	empty := false
	for i := 0; i < queueLength && !empty; i++ {
		select {
		case res := <-r.destroyResourceQueues[r.destroyResourceQueueIndex]:
			switch res.resourceType {
			case VulkanResourceTypeTexture:
				vk.DestroyImageView(r.device, res.resources[1].(vk.ImageView), nil)
				vk.DestroyImage(r.device, res.resources[0].(vk.Image), nil)
				vk.FreeMemory(r.device, res.resources[2].(vk.DeviceMemory), nil)
				break
			case VulkanResourceTypePaletteTexture:
				r.palTexture.emptySlot.PushFront(res.resources[0])
				break
			case VulkanResourceTypeBuffer:
				vk.DestroyBuffer(r.device, res.resources[0].(vk.Buffer), nil)
				vk.FreeMemory(r.device, res.resources[1].(vk.DeviceMemory), nil)
				break
			case VulkanResourceTypeFence:
				vk.WaitForFences(r.device, 1, []vk.Fence{res.resources[0].(vk.Fence)}, vk.True, 10*1000*1000*1000)
				vk.DestroyFence(r.device, res.resources[0].(vk.Fence), nil)
				break
			}
		default:
			empty = true
		}
	}
}

func (r *Renderer_VK) BeginFrame(clearColor bool) {
	sys.absTickCountF++
	now := sdl.GetPerformanceCounter()
	firstFrame := sys.prevTimestamp == 0
	diff := float32(now - sys.prevTimestamp)
	if diff*float32(sdl.GetPerformanceFrequency()) >= 1 {
		sys.gameFPS = float32(sdl.GetPerformanceFrequency()) / diff
		sys.absTickCountF = 0
		sys.prevTimestamp = now
	}
	if !firstFrame {
		vk.WaitForFences(r.device, 1, r.fences[:1], vk.True, 10*1000*1000*1000)
	}
	vk.ResetFences(r.device, 1, r.fences[:1])
	r.destroyResourceQueueIndex = (r.destroyResourceQueueIndex + 1) % 2
	if len(r.destroyResourceQueues[r.destroyResourceQueueIndex]) > 0 {
		go r.DestroyResources(len(r.destroyResourceQueues[r.destroyResourceQueueIndex]))
	}
	if len(r.usedCommands) > 0 {
		if r.stagingBufferFences[0] {
			vk.WaitForFences(r.device, 1, r.fences[1:2], vk.True, 10*1000*1000*1000)
			vk.ResetFences(r.device, 1, r.fences[1:2])
			r.stagingBufferFences[0] = false
		}
		if r.stagingBufferFences[1] {
			vk.WaitForFences(r.device, 1, r.fences[2:3], vk.True, 10*1000*1000*1000)
			vk.ResetFences(r.device, 1, r.fences[2:3])
			r.stagingBufferFences[1] = false
		}
		vk.FreeCommandBuffers(r.device, r.commandPools[0], uint32(len(r.usedCommands)), r.usedCommands)
		r.usedCommands = r.usedCommands[:0]
	}
	if r.setVSync {
		r.setVSync = false
		r.RecreateSwapchain()
	}
	res := vk.AcquireNextImage(r.device, r.swapchains[0].swapchain,
		vk.MaxUint64, r.semaphores[0], vk.NullFence, &r.swapchains[0].currentImageIndex)
	if res != vk.Success {
		if res == vk.ErrorOutOfDate || res == vk.Suboptimal {
			log.Println("[INFO] recreate swapchain")
			r.RecreateSwapchain()
			res = vk.AcquireNextImage(r.device, r.swapchains[0].swapchain,
				vk.MaxUint64, r.semaphores[0], vk.NullFence, &r.swapchains[0].currentImageIndex)
			if res != vk.Success {
				// Try continue execution if error is suboptimal
				if res == vk.Suboptimal {
					log.Println("[WARNING] vk.AcquireNextImage returned suboptimal after swapchain recreation")
				} else {
					err := fmt.Errorf("vk.AcquireNextImage failed with %s", vk.Error(res))
					panic(err)
				}
			}
		} else {
			err := fmt.Errorf("vk.AcquireNextImage failed with %s", vk.Error(res))
			panic(err)
		}
	}
	vk.ResetCommandBuffer(r.commandBuffers[0], 0)
	cmdBufferBeginInfo := vk.CommandBufferBeginInfo{
		SType:            vk.StructureTypeCommandBufferBeginInfo,
		Flags:            0,
		PInheritanceInfo: nil,
	}
	renderPassBeginInfo := vk.RenderPassBeginInfo{
		SType:       vk.StructureTypeRenderPassBeginInfo,
		RenderPass:  r.mainRenderPass.renderPass,
		Framebuffer: r.mainRenderTarget.framebuffer,
		RenderArea: vk.Rect2D{
			Offset: vk.Offset2D{
				X: 0, Y: 0,
			},
			Extent: vk.Extent2D{
				Width:  uint32(sys.scrrect[2]),
				Height: uint32(sys.scrrect[3]),
			},
		},
		ClearValueCount: 0,
		//PClearValues:    clearValues,
	}
	err := vk.Error(vk.BeginCommandBuffer(r.commandBuffers[0], &cmdBufferBeginInfo))
	if err != nil {
		err = fmt.Errorf("vk.BeginCommandBuffer failed with %s", err)
		panic(err)
	}

	imageMemoryBarrier := []vk.ImageMemoryBarrier{{
		SType:         vk.StructureTypeImageMemoryBarrier,
		DstAccessMask: vk.AccessFlags(vk.AccessColorAttachmentWriteBit),
		OldLayout:     vk.ImageLayoutUndefined,
		NewLayout:     vk.ImageLayoutColorAttachmentOptimal,
		Image:         r.mainRenderTarget.texture.img,
		SubresourceRange: vk.ImageSubresourceRange{
			AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
			BaseMipLevel:   0,
			LevelCount:     1,
			BaseArrayLayer: 0,
			LayerCount:     1,
		},
	}}
	vk.CmdPipelineBarrier(r.commandBuffers[0], vk.PipelineStageFlags(vk.PipelineStageBottomOfPipeBit), vk.PipelineStageFlags(vk.PipelineStageColorAttachmentOutputBit), 0, 0, nil, 0, nil, 1, imageMemoryBarrier)

	vk.CmdBeginRenderPass(r.commandBuffers[0], &renderPassBeginInfo, vk.SubpassContentsInline)
	if clearColor {
		clearAttachments := []vk.ClearAttachment{
			{
				AspectMask:      vk.ImageAspectFlags(vk.ImageAspectColorBit),
				ColorAttachment: 0,
				ClearValue:      vk.NewClearValue([]float32{0.0, 0.0, 0.0, 1}),
			},
			{
				AspectMask: vk.ImageAspectFlags(vk.ImageAspectDepthBit),
				ClearValue: vk.NewClearDepthStencil(1.0, 0),
			},
		}
		vk.CmdClearAttachments(r.commandBuffers[0], uint32(len(clearAttachments)), clearAttachments, 1, []vk.ClearRect{{
			Rect: vk.Rect2D{
				Offset: vk.Offset2D{
					X: 0, Y: 0,
				},
				Extent: vk.Extent2D{
					Width:  uint32(sys.scrrect[2]),
					Height: uint32(sys.scrrect[3]),
				},
			},
			BaseArrayLayer: 0,
			LayerCount:     1,
		}})
	}
	r.renderShadowMap = false
	r.VKState.currentProgram = nil
	r.vertexBufferOffset = 0
	r.spriteProgram.uniformOffsetMap = make(map[interface{}]uint32)
	r.spriteProgram.uniformBufferOffset = 0
	r.currentSpriteTexture.spriteTexture = r.dummyTexture
	r.currentSpriteTexture.palTexture = r.dummyTexture
	r.VKState.spriteTexture = r.dummyTexture
	r.VKState.palTexture = r.dummyTexture
	if r.modelProgram != nil {
		r.modelProgram.uniformOffsetMap = make(map[interface{}]uint32)
		r.modelProgram.uniformBufferOffset = 0
	}
}

func (r *Renderer_VK) EndFrame() {
	if len(r.stagingImageBarriers[0]) > 0 || len(r.tempCommands) > 0 {
		r.FlushTempCommands()
	}
	vk.CmdEndRenderPass(r.commandBuffers[0])
	imageMemoryBarrier := []vk.ImageMemoryBarrier{
		{
			SType:         vk.StructureTypeImageMemoryBarrier,
			SrcAccessMask: vk.AccessFlags(vk.AccessColorAttachmentWriteBit),
			DstAccessMask: vk.AccessFlags(vk.AccessTransferReadBit),
			OldLayout:     vk.ImageLayoutColorAttachmentOptimal,
			NewLayout:     vk.ImageLayoutTransferSrcOptimal,
			Image:         r.mainRenderTarget.texture.img,
			SubresourceRange: vk.ImageSubresourceRange{
				AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
				BaseMipLevel:   0,
				LevelCount:     1,
				BaseArrayLayer: 0,
				LayerCount:     1,
			},
		},
		{
			SType:         vk.StructureTypeImageMemoryBarrier,
			SrcAccessMask: vk.AccessFlags(vk.AccessColorAttachmentWriteBit),
			DstAccessMask: vk.AccessFlags(vk.AccessTransferWriteBit),
			OldLayout:     vk.ImageLayoutUndefined,
			NewLayout:     vk.ImageLayoutTransferDstOptimal,
			Image:         r.renderTargets[0].texture.img,
			SubresourceRange: vk.ImageSubresourceRange{
				AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
				BaseMipLevel:   0,
				LevelCount:     1,
				BaseArrayLayer: 0,
				LayerCount:     1,
			},
		},
	}
	vk.CmdPipelineBarrier(r.commandBuffers[0], vk.PipelineStageFlags(vk.PipelineStageColorAttachmentOutputBit), vk.PipelineStageFlags(vk.PipelineStageTransferBit), 0, 0, nil, 0, nil, uint32(len(imageMemoryBarrier)), imageMemoryBarrier)

	if sys.msaa > 1 {
		regions := []vk.ImageResolve{{
			SrcSubresource: vk.ImageSubresourceLayers{
				AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
				MipLevel:       0,
				BaseArrayLayer: 0,
				LayerCount:     1,
			},
			Extent: vk.Extent3D{
				Width:  uint32(r.mainRenderTarget.texture.width),
				Height: uint32(r.mainRenderTarget.texture.height),
				Depth:  1,
			},
			SrcOffset: vk.Offset3D{
				X: 0,
				Y: 0,
				Z: 0,
			},
			DstSubresource: vk.ImageSubresourceLayers{
				AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
				MipLevel:       0,
				BaseArrayLayer: 0,
				LayerCount:     1,
			},
			DstOffset: vk.Offset3D{
				X: 0,
				Y: 0,
				Z: 0,
			},
		}}
		vk.CmdResolveImage(r.commandBuffers[0], r.mainRenderTarget.texture.img, vk.ImageLayoutTransferSrcOptimal, r.renderTargets[0].texture.img, vk.ImageLayoutTransferDstOptimal, 1, regions)
	} else {
		regions := []vk.ImageCopy{{
			SrcSubresource: vk.ImageSubresourceLayers{
				AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
				MipLevel:       0,
				BaseArrayLayer: 0,
				LayerCount:     1,
			},
			SrcOffset: vk.Offset3D{
				X: 0,
				Y: 0,
				Z: 0,
			},
			DstSubresource: vk.ImageSubresourceLayers{
				AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
				MipLevel:       0,
				BaseArrayLayer: 0,
				LayerCount:     1,
			},
			DstOffset: vk.Offset3D{
				X: 0,
				Y: 0,
				Z: 0,
			},
			Extent: vk.Extent3D{
				Width:  uint32(r.mainRenderTarget.texture.width),
				Height: uint32(r.mainRenderTarget.texture.height),
				Depth:  1,
			},
		}}
		vk.CmdCopyImage(r.commandBuffers[0], r.mainRenderTarget.texture.img, vk.ImageLayoutTransferSrcOptimal, r.renderTargets[0].texture.img, vk.ImageLayoutTransferDstOptimal, 1, regions)
	}
	imageMemoryBarrier = []vk.ImageMemoryBarrier{
		{
			SType:         vk.StructureTypeImageMemoryBarrier,
			SrcAccessMask: vk.AccessFlags(vk.AccessTransferWriteBit),
			DstAccessMask: vk.AccessFlags(vk.AccessShaderReadBit),
			OldLayout:     vk.ImageLayoutTransferDstOptimal,
			NewLayout:     vk.ImageLayoutShaderReadOnlyOptimal,
			Image:         r.renderTargets[0].texture.img,
			SubresourceRange: vk.ImageSubresourceRange{
				AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
				BaseMipLevel:   0,
				LevelCount:     1,
				BaseArrayLayer: 0,
				LayerCount:     1,
			},
		},
	}
	vk.CmdPipelineBarrier(r.commandBuffers[0], vk.PipelineStageFlags(vk.PipelineStageTransferBit), vk.PipelineStageFlags(vk.PipelineStageFragmentShaderBit), 0, 0, nil, 0, nil, 1, imageMemoryBarrier)

	clearValues := []vk.ClearValue{vk.NewClearValue([]float32{0.0, 0.0, 0.0, 1})}
	r.SetVertexData(-1, -1, 1, -1, -1, 1, 1, 1)
	bufferIndex := int(r.vertexBufferOffset) / int(r.vertexBuffers[0].size)
	if r.vertexBufferOffset > 0 && int(r.vertexBufferOffset)%int(r.vertexBuffers[0].size) == 0 {
		bufferIndex -= 1
	}
	vk.CmdBindVertexBuffers(r.commandBuffers[0], 0, 1, []vk.Buffer{r.vertexBuffers[bufferIndex].buffer}, []vk.DeviceSize{0})

	//TextureSize, CurrentTime
	pushConstants := [3]float32{float32(r.renderTargets[0].texture.width), float32(r.renderTargets[0].texture.height), float32(sdl.GetPerformanceCounter())}
	vk.CmdPushConstants(r.commandBuffers[0], r.postProcessingProgram.pipelineLayout, vk.ShaderStageFlags(vk.ShaderStageVertexBit|vk.ShaderStageFragmentBit), 0, 4*3, unsafe.Pointer(&pushConstants[0]))

	for i := 0; i < len(r.postProcessingProgram.pipelines)-1; i++ {
		renderPassBeginInfo := vk.RenderPassBeginInfo{
			SType:       vk.StructureTypeRenderPassBeginInfo,
			RenderPass:  r.postProcessingRenderPass.renderPass,
			Framebuffer: r.renderTargets[(i+1)%2].framebuffer,
			RenderArea: vk.Rect2D{
				Offset: vk.Offset2D{
					X: 0, Y: 0,
				},
				Extent: r.swapchains[0].extent,
			},
			ClearValueCount: 1,
			PClearValues:    clearValues,
		}
		vk.CmdBeginRenderPass(r.commandBuffers[0], &renderPassBeginInfo, vk.SubpassContentsInline)
		vk.CmdBindPipeline(r.commandBuffers[0], vk.PipelineBindPointGraphics, r.postProcessingProgram.pipelines[i])
		vk.CmdBindDescriptorSets(r.commandBuffers[0], vk.PipelineBindPointGraphics, r.postProcessingProgram.pipelineLayout, 0, 1, []vk.DescriptorSet{r.renderTargets[i].descriptorSet}, 0, nil)
		viewports := []vk.Viewport{{
			MinDepth: 0.0,
			MaxDepth: 1.0,
			X:        0,
			Y:        0,
			Width:    float32(r.renderTargets[0].texture.width),
			Height:   float32(r.renderTargets[0].texture.height),
		}}
		scissor := vk.Rect2D{
			Offset: vk.Offset2D{
				X: 0,
				Y: 0,
			},
			Extent: vk.Extent2D{
				Width:  uint32(r.renderTargets[0].texture.width),
				Height: uint32(r.renderTargets[0].texture.height),
			},
		}
		vk.CmdSetViewport(r.commandBuffers[0], 0, 1, viewports)
		scissors := []vk.Rect2D{scissor}
		vk.CmdSetScissor(r.commandBuffers[0], 0, 1, scissors)
		vk.CmdDraw(r.commandBuffers[0], 4, 1, uint32(r.vertexBufferOffset%r.vertexBuffers[0].size)/8-4, 0)
		vk.CmdEndRenderPass(r.commandBuffers[0])
	}
	renderPassBeginInfo := vk.RenderPassBeginInfo{
		SType:       vk.StructureTypeRenderPassBeginInfo,
		RenderPass:  r.swapchainRenderPass.renderPass,
		Framebuffer: r.swapchains[0].framebuffers[r.swapchains[0].currentImageIndex],
		RenderArea: vk.Rect2D{
			Offset: vk.Offset2D{
				X: 0, Y: 0,
			},
			Extent: r.swapchains[0].extent,
		},
		ClearValueCount: 1,
		PClearValues:    clearValues,
	}
	vk.CmdBeginRenderPass(r.commandBuffers[0], &renderPassBeginInfo, vk.SubpassContentsInline)
	vk.CmdBindPipeline(r.commandBuffers[0], vk.PipelineBindPointGraphics, r.postProcessingProgram.pipelines[len(r.postProcessingProgram.pipelines)-1])
	vk.CmdBindDescriptorSets(r.commandBuffers[0], vk.PipelineBindPointGraphics, r.postProcessingProgram.pipelineLayout, 0, 1, []vk.DescriptorSet{r.renderTargets[(len(r.postProcessingProgram.pipelines)+1)%2].descriptorSet}, 0, nil)

	x, y, width, height := sys.window.GetScaledViewportSize()
	viewports := []vk.Viewport{{
		MinDepth: 0.0,
		MaxDepth: 1.0,
		X:        float32(x),
		Y:        float32(y),
		Width:    float32(width),
		Height:   float32(height),
	}}
	scissor := vk.Rect2D{
		Offset: vk.Offset2D{
			X: 0,
			Y: 0,
		},
		Extent: r.swapchains[0].extent,
	}
	vk.CmdSetViewport(r.commandBuffers[0], 0, 1, viewports)
	scissors := []vk.Rect2D{scissor}
	vk.CmdSetScissor(r.commandBuffers[0], 0, 1, scissors)
	vk.CmdDraw(r.commandBuffers[0], 4, 1, uint32(r.vertexBufferOffset%r.vertexBuffers[0].size)/8-4, 0)

	vk.CmdEndRenderPass(r.commandBuffers[0])
	vk.EndCommandBuffer(r.commandBuffers[0])

	var submitCommands []vk.CommandBuffer
	if r.renderShadowMap {
		r.waitGroup.Wait()
		submitCommands = []vk.CommandBuffer{r.commandBuffers[1], r.commandBuffers[0]}
	} else {
		submitCommands = r.commandBuffers[:1]
	}

	submitInfo := []vk.SubmitInfo{{
		SType:              vk.StructureTypeSubmitInfo,
		WaitSemaphoreCount: 1,
		PWaitSemaphores:    r.semaphores,
		PWaitDstStageMask:  []vk.PipelineStageFlags{vk.PipelineStageFlags(vk.PipelineStageColorAttachmentOutputBit)},
		CommandBufferCount: uint32(len(submitCommands)),
		PCommandBuffers:    submitCommands,
	}}

	err := vk.Error(vk.QueueSubmit(r.queue, 1, submitInfo, r.fences[0]))
	if err != nil {
		err = fmt.Errorf("vk.QueueSubmit failed with %s", err)
		panic(err)
	}

	imageIndices := []uint32{r.swapchains[0].currentImageIndex}
	presentInfo := vk.PresentInfo{
		SType:          vk.StructureTypePresentInfo,
		SwapchainCount: 1,
		PSwapchains:    []vk.Swapchain{r.swapchains[0].swapchain},
		PImageIndices:  imageIndices,
	}

	res := vk.QueuePresent(r.queue, &presentInfo)
	if res != vk.Success {
		if res == vk.ErrorOutOfDate || res == vk.Suboptimal {
			log.Println("[INFO] recreate swapchain")
			r.RecreateSwapchain()
		} else {
			err := fmt.Errorf("vk.QueuePresent failed with %s", vk.Error(res))
			panic(err)
		}
	}
}

func (r *Renderer_VK) PrintVkState() {
	fmt.Printf("Number of Uniforms: %d \n", +len(r.spriteProgram.uniformOffsetMap))
	fmt.Printf("Uniform Size: %d \n", +r.spriteProgram.uniformSize)
	fmt.Printf("Uniform Alignment: %d \n", +r.minUniformBufferOffsetAlignment)

	fmt.Printf("Vertex buffer offset: %d \n", +r.vertexBufferOffset)
	return
}

func (r *Renderer_VK) Await() {
	vk.QueueWaitIdle(r.queue)
}

func (r *Renderer_VK) SetPipeline(eq BlendEquation, src, dst BlendFunc) {
	r.VKState.VulkanPipelineState.VulkanBlendState.op = eq
	r.VKState.VulkanPipelineState.VulkanBlendState.src = src
	r.VKState.VulkanPipelineState.VulkanBlendState.dst = dst
}

func (r *Renderer_VK) prepareShadowMapPipeline(bufferIndex uint32) {
	//r.renderShadowMap = true
	vk.ResetCommandBuffer(r.commandBuffers[1], 0)
	cmdBufferBeginInfo := vk.CommandBufferBeginInfo{
		SType:            vk.StructureTypeCommandBufferBeginInfo,
		Flags:            0,
		PInheritanceInfo: nil,
	}
	vk.BeginCommandBuffer(r.commandBuffers[1], &cmdBufferBeginInfo)

	barriers := []vk.ImageMemoryBarrier{
		{
			SType:               vk.StructureTypeImageMemoryBarrier,
			OldLayout:           vk.ImageLayoutUndefined,
			NewLayout:           vk.ImageLayoutDepthStencilAttachmentOptimal,
			SrcAccessMask:       vk.AccessFlags(vk.AccessNone),
			DstAccessMask:       vk.AccessFlags(vk.AccessDepthStencilAttachmentWriteBit),
			SrcQueueFamilyIndex: vk.QueueFamilyIgnored,
			DstQueueFamilyIndex: vk.QueueFamilyIgnored,
			Image:               r.shadowMapTextures.img,
			SubresourceRange: vk.ImageSubresourceRange{
				AspectMask:     vk.ImageAspectFlags(vk.ImageAspectDepthBit),
				BaseMipLevel:   0,
				LevelCount:     1,
				BaseArrayLayer: 0,
				LayerCount:     6 * 4,
			},
		},
	}
	vk.CmdPipelineBarrier(r.commandBuffers[1], vk.PipelineStageFlags(vk.PipelineStageTopOfPipeBit), vk.PipelineStageFlags(vk.PipelineStageEarlyFragmentTestsBit), 0, 0, nil, 0, nil, uint32(len(barriers)), barriers)

	depthAttachmentInfos := []vk.RenderingAttachmentInfo{{
		SType:       vk.StructureTypeRenderingAttachmentInfo,
		ImageView:   r.shadowMapTextures.imageView,
		ImageLayout: vk.ImageLayoutDepthStencilAttachmentOptimal,
		LoadOp:      vk.AttachmentLoadOpClear,
		StoreOp:     vk.AttachmentStoreOpStore,
		ClearValue:  vk.NewClearDepthStencil(1.0, 0),
	}}
	renderInfo := vk.RenderingInfo{
		SType: vk.StructureTypeRenderingInfo,
		RenderArea: vk.Rect2D{
			Offset: vk.Offset2D{
				X: 0, Y: 0,
			},
			Extent: vk.Extent2D{
				Width:  uint32(r.shadowMapTextures.width),
				Height: uint32(r.shadowMapTextures.height),
			},
		},
		LayerCount:           6 * 4,
		ColorAttachmentCount: 0,
		PColorAttachments:    nil,
		PDepthAttachment:     depthAttachmentInfos,
	}
	vk.CmdBeginRendering(r.commandBuffers[1], &renderInfo)

	vk.CmdBindIndexBuffer(r.commandBuffers[1], r.modelIndexBuffers[bufferIndex].buffer, 0, vk.IndexTypeUint32)
	r.VKState.modelVertexBufferIndex = bufferIndex

	r.VKState.VulkanShadowMapTexture.tex = r.dummyTexture
	r.VKState.VulkanShadowMapTexture.jointMatricesTexture = r.dummyTexture
	r.VKState.VulkanShadowMapTexture.morphTargetTexture = r.dummyTexture

	r.VKState.currentShadowMapTexture.tex = nil
	r.VKState.currentShadowMapTexture.jointMatricesTexture = nil
	r.VKState.currentShadowMapTexture.morphTargetTexture = nil
	r.VKState.currentShadowmMapPipeline = nil
	r.VKState.shadowMapUniformBufferOffset1 = 0
	r.VKState.shadowMapUniformBufferOffset2 = 0

	r.shadowMapProgram.uniformOffsetMap = make(map[interface{}]uint32)
	r.shadowMapProgram.uniformBufferOffset = 0
}
func (r *Renderer_VK) setShadowMapPipeline(doubleSided, invertFrontFace, useUV, useNormal, useTangent, useVertColor, useJoint0, useJoint1 bool, numVertices, vertAttrOffset uint32) {
	r.VKState.VulkanShadowMapPipelineState.doubleSided = doubleSided
	r.VKState.VulkanShadowMapPipelineState.invertFrontFace = invertFrontFace
	r.VKState.VulkanShadowMapPipelineState.useUV = useUV
	r.VKState.VulkanShadowMapPipelineState.useNormal = useNormal
	r.VKState.VulkanShadowMapPipelineState.useTangent = useTangent
	r.VKState.VulkanShadowMapPipelineState.useVertColor = useVertColor
	r.VKState.VulkanShadowMapPipelineState.useJoint0 = useJoint0
	r.VKState.VulkanShadowMapPipelineState.useJoint1 = useJoint1
	r.VKState.VulkanShadowMapPipelineState.useVertColor = useVertColor
	r.VKState.shadowMapVertAttrOffset = vertAttrOffset
}

func (r *Renderer_VK) ReleaseShadowPipeline() {
	vk.CmdEndRendering(r.commandBuffers[1])
	barriers := []vk.ImageMemoryBarrier{
		{
			SType:               vk.StructureTypeImageMemoryBarrier,
			OldLayout:           vk.ImageLayoutDepthStencilAttachmentOptimal,
			NewLayout:           vk.ImageLayoutShaderReadOnlyOptimal,
			SrcAccessMask:       vk.AccessFlags(vk.AccessDepthStencilAttachmentWriteBit),
			DstAccessMask:       vk.AccessFlags(vk.AccessShaderReadBit),
			SrcQueueFamilyIndex: vk.QueueFamilyIgnored,
			DstQueueFamilyIndex: vk.QueueFamilyIgnored,
			Image:               r.shadowMapTextures.img,
			SubresourceRange: vk.ImageSubresourceRange{
				AspectMask:     vk.ImageAspectFlags(vk.ImageAspectDepthBit),
				BaseMipLevel:   0,
				LevelCount:     1,
				BaseArrayLayer: 0,
				LayerCount:     6 * 4,
			},
		},
	}
	vk.CmdPipelineBarrier(r.commandBuffers[1], vk.PipelineStageFlags(vk.PipelineStageLateFragmentTestsBit), vk.PipelineStageFlags(vk.PipelineStageFragmentShaderBit), 0, 0, nil, 0, nil, uint32(len(barriers)), barriers)
	vk.EndCommandBuffer(r.commandBuffers[1])
	r.waitGroup.Done()
}
func (r *Renderer_VK) prepareModelPipeline(bufferIndex uint32, env *Environment) {
	vk.CmdBindIndexBuffer(r.commandBuffers[0], r.modelIndexBuffers[bufferIndex].buffer, 0, vk.IndexTypeUint32)
	r.VKState.modelVertexBufferIndex = bufferIndex
	clearAttachments := []vk.ClearAttachment{
		{
			AspectMask: vk.ImageAspectFlags(vk.ImageAspectDepthBit),
			ClearValue: vk.NewClearDepthStencil(1.0, 0),
		},
	}
	vk.CmdClearAttachments(r.commandBuffers[0], uint32(len(clearAttachments)), clearAttachments, 1, []vk.ClearRect{{
		Rect: vk.Rect2D{
			Offset: vk.Offset2D{
				X: 0, Y: 0,
			},
			Extent: vk.Extent2D{
				Width:  uint32(sys.scrrect[2]),
				Height: uint32(sys.scrrect[3]),
			},
		},
		BaseArrayLayer: 0,
		LayerCount:     1,
	}})
	if env != nil {
		r.VKState.VulkanModelTexture.lambertianEnvSampler = env.lambertianTexture.tex.(*Texture_VK)
		r.VKState.VulkanModelTexture.GGXEnvSampler = env.GGXTexture.tex.(*Texture_VK)
		r.VKState.VulkanModelTexture.GGXLUT = env.GGXLUT.tex.(*Texture_VK)
		r.VKState.environmentIntensity = env.environmentIntensity
		r.VKState.mipCount = env.mipmapLevels
		rotation := mgl.Rotate3DX(math.Pi).Mul3(mgl.Rotate3DY(0.5 * math.Pi))
		r.Copy3x3Matrix(&r.VKState.environmentRotation, rotation[:])
	} else {
		r.VKState.environmentIntensity = 0
		r.VKState.mipCount = 0
		r.VKState.VulkanModelTexture.lambertianEnvSampler = r.dummyCubeTexture
		r.VKState.VulkanModelTexture.GGXEnvSampler = r.dummyCubeTexture
		r.VKState.VulkanModelTexture.GGXLUT = r.dummyTexture
	}
	r.VKState.VulkanModelTexture.tex = r.dummyTexture
	r.VKState.VulkanModelTexture.normalMap = r.dummyTexture
	r.VKState.VulkanModelTexture.metallicRoughnessMap = r.dummyTexture
	r.VKState.VulkanModelTexture.ambientOcclusionMap = r.dummyTexture
	r.VKState.emissionMap = r.dummyTexture
	r.VKState.VulkanModelTexture.jointMatricesTexture = r.dummyTexture
	r.VKState.VulkanModelTexture.morphTargetTexture = r.dummyTexture

	r.VKState.currentModelTexture.tex = nil
	r.VKState.currentModelTexture.normalMap = nil
	r.VKState.currentModelTexture.metallicRoughnessMap = nil
	r.VKState.currentModelTexture.ambientOcclusionMap = nil
	r.VKState.currentModelTexture.emissionMap = nil
	r.VKState.currentModelTexture.jointMatricesTexture = nil
	r.VKState.currentModelTexture.morphTargetTexture = nil
}
func (r *Renderer_VK) SetModelPipeline(eq BlendEquation, src, dst BlendFunc, depthTest, depthMask, doubleSided, invertFrontFace, useUV, useNormal, useTangent, useVertColor, useJoint0, useJoint1, useOutlineAttribute bool, numVertices, vertAttrOffset uint32) {
	r.VKState.VulkanBlendState.op = eq
	r.VKState.VulkanBlendState.src = src
	r.VKState.VulkanBlendState.dst = dst
	r.VKState.VulkanModelPipelineState.depthTest = depthTest
	r.VKState.VulkanModelPipelineState.depthMask = depthMask
	r.VKState.VulkanModelPipelineState.doubleSided = doubleSided
	r.VKState.VulkanModelPipelineState.invertFrontFace = invertFrontFace
	r.VKState.VulkanModelPipelineState.useUV = useUV
	r.VKState.VulkanModelPipelineState.useNormal = useNormal
	r.VKState.VulkanModelPipelineState.useTangent = useTangent
	r.VKState.VulkanModelPipelineState.useVertColor = useVertColor
	r.VKState.VulkanModelPipelineState.useJoint0 = useJoint0
	r.VKState.VulkanModelPipelineState.useJoint1 = useJoint1
	r.VKState.VulkanModelPipelineState.useOutlineAttribute = useOutlineAttribute
	r.VKState.VulkanModelPipelineState.useVertColor = useVertColor
	r.VKState.modelVertAttrOffset = vertAttrOffset
}
func (r *Renderer_VK) SetMeshOulinePipeline(invertFrontFace bool, meshOutline float32) {
	r.VKState.VulkanModelPipelineState.invertFrontFace = invertFrontFace
	r.VKState.VulkanModelPipelineState.depthTest = true
	r.VKState.VulkanModelPipelineState.depthMask = true
	r.VKState.VulkanModelProgramUniformBufferObject2.meshOutline = meshOutline
}
func (r *Renderer_VK) ReleaseModelPipeline() {
	//Do nothing
}

func (r *Renderer_VK) ReadPixels(data []uint8, width, height int) {
	//Make sure the rendering is finished
	vk.WaitForFences(r.device, 1, r.fences[:1], vk.True, 10*1000*1000*1000)
	cmd := r.BeginSingleTimeCommands()
	imageIndex := r.swapchains[0].currentImageIndex
	//Create a temporary texture in host visible memory
	img := r.CreateImage(r.swapchains[0].extent.Width, r.swapchains[0].extent.Height, vk.FormatR8g8b8a8Unorm, 1, 1, vk.ImageUsageFlags(vk.ImageUsageTransferDstBit), 1, vk.ImageTilingLinear, false)
	imageMemory := r.AllocateImageMemory(img, vk.MemoryPropertyHostVisibleBit|vk.MemoryPropertyHostCachedBit)
	imgBarriers := []vk.ImageMemoryBarrier{
		{
			SType:         vk.StructureTypeImageMemoryBarrier,
			OldLayout:     vk.ImageLayoutUndefined,
			NewLayout:     vk.ImageLayoutTransferDstOptimal,
			SrcAccessMask: vk.AccessFlags(vk.AccessNone),
			DstAccessMask: vk.AccessFlags(vk.AccessTransferWriteBit),
			Image:         img,
			SubresourceRange: vk.ImageSubresourceRange{
				AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
				BaseMipLevel:   0,
				LevelCount:     1,
				BaseArrayLayer: 0,
				LayerCount:     1,
			},
		},
		{
			SType:         vk.StructureTypeImageMemoryBarrier,
			OldLayout:     vk.ImageLayoutPresentSrc,
			NewLayout:     vk.ImageLayoutTransferSrcOptimal,
			SrcAccessMask: vk.AccessFlags(vk.AccessNone),
			DstAccessMask: vk.AccessFlags(vk.AccessTransferReadBit),
			Image:         r.swapchains[0].images[imageIndex],
			SubresourceRange: vk.ImageSubresourceRange{
				AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
				BaseMipLevel:   0,
				LevelCount:     1,
				BaseArrayLayer: 0,
				LayerCount:     1,
			},
		},
	}
	vk.CmdPipelineBarrier(cmd, vk.PipelineStageFlags(vk.PipelineStageBottomOfPipeBit), vk.PipelineStageFlags(vk.PipelineStageTransferBit), 0, 0, nil, 0, nil, 2, imgBarriers)

	imageBlits := []vk.ImageBlit{{
		SrcSubresource: vk.ImageSubresourceLayers{
			AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
			MipLevel:       0,
			BaseArrayLayer: 0,
			LayerCount:     1,
		},
		SrcOffsets: [2]vk.Offset3D{
			{
				X: 0,
				Y: 0,
				Z: 0,
			},
			{
				X: int32(r.swapchains[0].extent.Width),
				Y: int32(r.swapchains[0].extent.Height),
				Z: 1,
			},
		},
		DstSubresource: vk.ImageSubresourceLayers{
			AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
			MipLevel:       0,
			BaseArrayLayer: 0,
			LayerCount:     1,
		},
		DstOffsets: [2]vk.Offset3D{
			{
				X: 0,
				Y: 0,
				Z: 0,
			},
			{
				X: int32(r.swapchains[0].extent.Width),
				Y: int32(r.swapchains[0].extent.Height),
				Z: 1,
			},
		},
	}}
	vk.CmdBlitImage(cmd, r.swapchains[0].images[imageIndex], vk.ImageLayoutTransferSrcOptimal, img, vk.ImageLayoutTransferDstOptimal, uint32(len(imageBlits)), imageBlits, vk.FilterLinear)
	imgBarriers = []vk.ImageMemoryBarrier{
		{
			SType:         vk.StructureTypeImageMemoryBarrier,
			OldLayout:     vk.ImageLayoutTransferDstOptimal,
			NewLayout:     vk.ImageLayoutGeneral,
			SrcAccessMask: vk.AccessFlags(vk.AccessTransferWriteBit),
			DstAccessMask: vk.AccessFlags(vk.AccessNone),
			Image:         img,
			SubresourceRange: vk.ImageSubresourceRange{
				AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
				BaseMipLevel:   0,
				LevelCount:     1,
				BaseArrayLayer: 0,
				LayerCount:     1,
			},
		},
	}
	vk.CmdPipelineBarrier(cmd, vk.PipelineStageFlags(vk.PipelineStageTransferBit), vk.PipelineStageFlags(vk.PipelineStageTopOfPipeBit), 0, 0, nil, 0, nil, 1, imgBarriers)

	r.EndSingleTimeCommands(cmd)
	subResource := vk.ImageSubresource{
		AspectMask: vk.ImageAspectFlags(vk.ImageAspectColorBit),
		MipLevel:   0,
		ArrayLayer: 0,
	}
	var subResourceLayout vk.SubresourceLayout
	vk.GetImageSubresourceLayout(r.device, img, &subResource, &subResourceLayout)
	//subResourceLayout.Deref()
	var mappedData unsafe.Pointer
	vk.MapMemory(r.device, imageMemory, subResourceLayout.Offset, vk.DeviceSize(width*height*4), 0, &mappedData)
	const m = 0x7fffffff
	imageData := (*[m]byte)(mappedData)
	if int(subResourceLayout.RowPitch) == width*4 {
		copy(data, imageData[:width*height*4])
	} else {
		for y := 0; y < height; y++ {
			copy(data[y*width*4:(y+1)*width*4], imageData[y*int(subResourceLayout.RowPitch):y*int(subResourceLayout.RowPitch)+width*4])
		}
	}

	vk.UnmapMemory(r.device, imageMemory)
	vk.DestroyImage(r.device, img, nil)
	vk.FreeMemory(r.device, imageMemory, nil)
}

func (r *Renderer_VK) EnableScissor(x, y, width, height int32) {
	r.VKState.scissor = vk.Rect2D{
		Offset: vk.Offset2D{
			X: int32(MaxI(int(x), 0)),
			Y: int32(MaxI(int(y), 0)),
		},
		Extent: vk.Extent2D{
			Width:  uint32(width),
			Height: uint32(height),
		},
	}
}

func (r *Renderer_VK) DisableScissor() {
	r.VKState.scissor = vk.Rect2D{
		Offset: vk.Offset2D{
			X: 0,
			Y: 0,
		},
		Extent: vk.Extent2D{
			Width:  uint32(sys.scrrect[2]),
			Height: uint32(sys.scrrect[3]),
		},
	}
}
func (r *Renderer_VK) Copy3x3Matrix(dst *[12]float32, src []float32) {
	copy((*dst)[:3], src)
	copy((*dst)[4:7], src[3:])
	copy((*dst)[8:11], src[6:])
}
func (r *Renderer_VK) SetUniformMatrix(name string, value []float32) {
	switch name {
	case "modelview":
		copy(r.VKState.VulkanSpriteProgramVertUniformBufferObject.modelView[:], value)
	case "projection":
		copy(r.VKState.VulkanSpriteProgramVertUniformBufferObject.projection[:], value)
	}
}

func (r *Renderer_VK) SetUniformI(name string, value int) {
	switch name {
	case "isRgba":
		r.VKState.VulkanSpriteProgramFragUniformBufferObject.isRgba = uint32(value)
	case "isTrapez":
		r.VKState.VulkanSpriteProgramFragUniformBufferObject.isTrapez = uint32(value)
	case "isFlat":
		r.VKState.VulkanSpriteProgramFragUniformBufferObject.isFlat = uint32(value)
	case "neg":
		r.VKState.VulkanSpriteProgramFragUniformBufferObject.neg = uint32(value)
	case "mask":
		r.VKState.VulkanSpriteProgramFragUniformBufferObject.mask = int32(value)
	}
}

func (r *Renderer_VK) SetUniformF(name string, values ...float32) {
	switch name {
	case "tint":
		copy(r.VKState.VulkanSpriteProgramFragUniformBufferObject.tint[:], values)
	case "x1x2x4x3":
		copy(r.VKState.VulkanSpriteProgramFragUniformBufferObject.x1x2x4x3[:], values)
	case "gray":
		r.VKState.VulkanSpriteProgramFragUniformBufferObject.gray = values[0]
	case "hue":
		r.VKState.VulkanSpriteProgramFragUniformBufferObject.hue = values[0]
	case "alpha":
		r.VKState.VulkanSpriteProgramFragUniformBufferObject.alpha = values[0]
	}
}

func (r *Renderer_VK) SetUniformFv(name string, values []float32) {
	switch name {
	case "tint":
		copy(r.VKState.VulkanSpriteProgramFragUniformBufferObject.tint[:], values)
	case "add":
		copy(r.VKState.VulkanSpriteProgramFragUniformBufferObject.add[:], values)
	case "mult":
		copy(r.VKState.VulkanSpriteProgramFragUniformBufferObject.mult[:], values)
	}
}

func (r *Renderer_VK) SetTexture(name string, tex Texture) {
	if tex == nil {
		return
	}
	if name == "tex" {
		r.VKState.spriteTexture = tex.(*Texture_VK)
	} else if name == "pal" {
		r.VKState.palTexture = tex.(*Texture_VK)
	}

}

func (r *Renderer_VK) SetModelUniformI(name string, val int) {
	if len(name) > 7 && name[:7] == "lights[" {
		lightIndex, _ := strconv.Atoi(name[7:8])
		switch name[10:] {
		case "type":
			r.VKState.VulkanModelProgramUniformBufferObject0.lights[lightIndex].lightType = int32(val)
			break
		}
	}
	switch name {
	case "numVertices":
		r.VKState.VulkanModelProgramUniformBufferObject2.numVertices = uint32(val)
		break
	case "numTargets":
		r.VKState.VulkanModelProgramUniformBufferObject2.numTargets = int32(val)
		break
	case "morphTargetTextureDimension":
		r.VKState.VulkanModelProgramUniformBufferObject2.morphTargetTextureDimension = int32(val)
		break
	case "unlit":
		r.VKState.VulkanModelProgramUniformBufferObject1.unlit = val != 0
		break
	case "neg":
		if val != 0 {
			r.VKState.VulkanModelProgramUniformBufferObject2.neg = 1
		} else {
			r.VKState.VulkanModelProgramUniformBufferObject2.neg = 0
		}
		break
	case "enableAlpha":
		r.VKState.VulkanModelProgramUniformBufferObject1.enableAlpha = val != 0
		break
	case "numJoints":
		r.VulkanModelProgramUniformBufferObject2.numJoints = int32(val)
		break
	case "useTexture":
		r.VKState.VulkanModelPipelineState.useTexture = val != 0
		break
	case "useNormalMap":
		r.VKState.VulkanModelPipelineState.useNormalMap = val != 0
		break
	case "useMetallicRoughnessMap":
		r.VKState.VulkanModelPipelineState.useMetallicRoughnessMap = val != 0
		break
	case "useEmissionMap":
		r.VKState.VulkanModelPipelineState.useEmissionMap = val != 0
		break
	}
}

func (r *Renderer_VK) SetModelUniformF(name string, values ...float32) {
	if len(name) > 7 && name[:7] == "lights[" {
		lightIndex, _ := strconv.Atoi(name[7:8])
		switch name[10:] {
		case "color":
			copy(r.VKState.VulkanModelProgramUniformBufferObject0.lights[lightIndex].color[:], values)
			break
		case "intensity":
			r.VKState.VulkanModelProgramUniformBufferObject0.lights[lightIndex].intensity = values[0]
			break
		case "innerConeCos":
			r.VKState.VulkanModelProgramUniformBufferObject0.lights[lightIndex].innerConeCos = values[0]
			break
		case "outerConeCos":
			r.VKState.VulkanModelProgramUniformBufferObject0.lights[lightIndex].outerConeCos = values[0]
			break
		case "range":
			r.VKState.VulkanModelProgramUniformBufferObject0.lights[lightIndex].lightRange = values[0]
			break
		case "position":
			copy(r.VKState.VulkanModelProgramUniformBufferObject0.lights[lightIndex].position[:], values)
			break
		case "shadowMapFar":
			r.VKState.VulkanModelProgramUniformBufferObject0.lights[lightIndex].shadowMapFar = values[0]
			break
		case "shadowBias":
			r.VKState.VulkanModelProgramUniformBufferObject0.lights[lightIndex].shadowBias = values[0]
			break
		case "direction":
			copy(r.VKState.VulkanModelProgramUniformBufferObject0.lights[lightIndex].direction[:], values)
			break
		}
	} else {
		switch name {
		case "metallicRoughness":
			copy(r.VKState.VulkanModelProgramUniformBufferObject1.metallicRoughness[:], values)
			break
		case "ambientOcclusionStrength":
			r.VKState.VulkanModelProgramUniformBufferObject1.ambientOcclusionStrength = values[0]
			break
		case "cameraPosition":
			copy(r.VKState.VulkanModelProgramUniformBufferObject0.cameraPosition[:], values)
			break
		case "morphTargetOffset":
			copy(r.VKState.VulkanModelProgramUniformBufferObject2.morphTargetOffset[:], values)
			break
		case "gray":
			r.VKState.VulkanModelProgramUniformBufferObject2.modelGray = values[0]
		case "hue":
			r.VKState.VulkanModelProgramUniformBufferObject2.modelHue = values[0]
		case "alphaThreshold":
			r.VKState.VulkanModelProgramUniformBufferObject1.alphaThreshold = values[0]
		case "meshOutline":
			r.VKState.VulkanModelProgramUniformBufferObject2.meshOutline = values[0]
		}
	}
}
func (r *Renderer_VK) SetModelUniformFv(name string, values []float32) {
	switch name {
	case "morphTargetWeight":
		copy(r.VKState.VulkanModelProgramUniformBufferObject2.morphTargetWeight[0][:], values)
		copy(r.VKState.VulkanModelProgramUniformBufferObject2.morphTargetWeight[1][4:], values)
		break
	case "add":
		copy(r.VKState.VulkanModelProgramUniformBufferObject2.modelAdd[:], values)
		break
	case "mult":
		copy(r.VKState.VulkanModelProgramUniformBufferObject2.modelMult[:], values)
		break
	case "baseColorFactor":
		copy(r.VKState.VulkanModelProgramUniformBufferObject1.baseColorFactor[:], values)
		break
	case "emission":
		copy(r.VKState.VulkanModelProgramUniformBufferObject1.emission[:], values)
		break
	}
}

func (r *Renderer_VK) SetModelUniformMatrix(name string, values []float32) {
	if len(name) > 14 && name[:14] == "lightMatrices[" {
		lightIndex, _ := strconv.Atoi(name[14:15])
		copy(r.VKState.VulkanModelProgramUniformBufferObject0.lightMatrices[lightIndex][:], values)
	} else {
		switch name {
		case "model":
			copy(r.VKState.VulkanModelProgramUniformBufferObject2.modelMatrix[:], values)
		case "normalMatrix":
			copy(r.VKState.VulkanModelProgramUniformBufferObject2.normalMatrix[:], values)
		case "projection":
			copy(r.VKState.VulkanModelProgramUniformBufferObject0.projectionMatrix[:], values)
		case "view":
			copy(r.VKState.VulkanModelProgramUniformBufferObject0.viewMatrix[:], values)
		}
	}
}

func (r *Renderer_VK) SetModelUniformMatrix3(name string, values []float32) {
	switch name {
	case "texTransform":
		r.Copy3x3Matrix(&r.VKState.VulkanModelProgramUniformBufferObject1.texTransform, values)
	case "normalMapTransform":
		r.Copy3x3Matrix(&r.VKState.VulkanModelProgramUniformBufferObject1.normalMapTransform, values)
	case "metallicRoughnessMapTransform":
		r.Copy3x3Matrix(&r.VKState.VulkanModelProgramUniformBufferObject1.metallicRoughnessMapTransform, values)
	case "ambientOcclusionMapTransform":
		r.Copy3x3Matrix(&r.VKState.VulkanModelProgramUniformBufferObject1.ambientOcclusionMapTransform, values)
	case "emissionMapTransform":
		r.Copy3x3Matrix(&r.VKState.VulkanModelProgramUniformBufferObject1.emissionMapTransform, values)
	}
}

func (r *Renderer_VK) SetModelTexture(name string, tex Texture) {
	switch name {
	case "jointMatrices":
		r.VKState.VulkanModelTexture.jointMatricesTexture = tex.(*Texture_VK)
	case "morphTargetValues":
		r.VKState.VulkanModelTexture.morphTargetTexture = tex.(*Texture_VK)
	case "tex":
		r.VKState.VulkanModelTexture.tex = tex.(*Texture_VK)
	case "normalMap":
		r.VKState.VulkanModelTexture.normalMap = tex.(*Texture_VK)
	case "metallicRoughnessMap":
		r.VKState.VulkanModelTexture.metallicRoughnessMap = tex.(*Texture_VK)
	case "ambientOcclusionMap":
		r.VKState.VulkanModelTexture.ambientOcclusionMap = tex.(*Texture_VK)
	case "emissionMap":
		r.VKState.VulkanModelTexture.emissionMap = tex.(*Texture_VK)
	}
}

func (r *Renderer_VK) SetShadowMapUniformI(name string, val int) {
	if len(name) > 7 && name[:7] == "lights[" {
		lightIndex, _ := strconv.Atoi(name[7:8])
		switch name[10:] {
		case "type":
			r.VKState.VulkanShadowMapProgramUniformBufferObject0.lights[lightIndex].lightType = int32(val)
			break
		}
	}
	switch name {
	case "lightIndex":
		r.VKState.VulkanShadowMapProgramPushConstant.lightIndex = int32(val)
		break
	case "numVertices":
		r.VKState.VulkanShadowMapProgramPushConstant.numVertices = int32(val)
		break
	case "numTargets":
		r.VKState.VulkanShadowMapProgramUniformBufferObject2.numTargets = int32(val)
		break
	case "morphTargetTextureDimension":
		r.VKState.VulkanShadowMapProgramUniformBufferObject2.morphTargetTextureDimension = int32(val)
		break
	case "enableAlpha":
		r.VKState.VulkanShadowMapProgramUniformBufferObject1.enableAlpha = val != 0
		break
	case "numJoints":
		r.VulkanShadowMapProgramUniformBufferObject2.numJoints = int32(val)
		break
	case "useTexture":
		r.VKState.VulkanShadowMapPipelineState.useTexture = val != 0
		break
	}
}

func (r *Renderer_VK) SetShadowMapUniformF(name string, values ...float32) {
	if len(name) > 7 && name[:7] == "lights[" {
		lightIndex, _ := strconv.Atoi(name[7:8])
		switch name[10:] {
		case "position":
			copy(r.VKState.VulkanShadowMapProgramUniformBufferObject0.lights[lightIndex].position[:], values)
			break
		case "shadowMapFar":
			r.VKState.VulkanShadowMapProgramUniformBufferObject0.lights[lightIndex].shadowMapFar = values[0]
			break
		}
	} else {
		switch name {
		case "morphTargetOffset":
			copy(r.VKState.VulkanShadowMapProgramUniformBufferObject2.morphTargetOffset[:], values)
			break
		case "alphaThreshold":
			r.VKState.VulkanShadowMapProgramUniformBufferObject1.alphaThreshold = values[0]
		}
	}
}
func (r *Renderer_VK) SetShadowMapUniformFv(name string, values []float32) {
	switch name {
	case "morphTargetWeight":
		copy(r.VKState.VulkanShadowMapProgramUniformBufferObject2.morphTargetWeight[0][:], values)
		copy(r.VKState.VulkanShadowMapProgramUniformBufferObject2.morphTargetWeight[1][4:], values)
		break
	case "baseColorFactor":
		copy(r.VKState.VulkanShadowMapProgramUniformBufferObject1.baseColorFactor[:], values)
		break
	}
}
func (r *Renderer_VK) SetShadowMapUniformMatrix(name string, values []float32) {
	if len(name) > 14 && name[:14] == "lightMatrices[" {
		if name[15:16] == "]" {
			lightIndex, _ := strconv.Atoi(name[14:15])
			copy(r.VKState.VulkanShadowMapProgramUniformBufferObject0.lightMatrices[lightIndex][:], values)
		} else {
			lightIndex, _ := strconv.Atoi(name[14:16])
			copy(r.VKState.VulkanShadowMapProgramUniformBufferObject0.lightMatrices[lightIndex][:], values)
		}
	} else {
		switch name {
		case "model":
			copy(r.VKState.VulkanShadowMapProgramPushConstant.model[:], values)
		}
	}
}

func (r *Renderer_VK) SetShadowMapUniformMatrix3(name string, values []float32) {
	switch name {
	case "texTransform":
		r.Copy3x3Matrix(&r.VKState.VulkanShadowMapProgramUniformBufferObject1.texTransform, values)
		break
	}
}

func (r *Renderer_VK) SetShadowMapTexture(name string, tex Texture) {
	switch name {
	case "jointMatrices":
		r.VKState.VulkanShadowMapTexture.jointMatricesTexture = tex.(*Texture_VK)
	case "morphTargetValues":
		r.VKState.VulkanShadowMapTexture.morphTargetTexture = tex.(*Texture_VK)
	case "tex":
		r.VKState.VulkanShadowMapTexture.tex = tex.(*Texture_VK)
	}
}

func (r *Renderer_VK) SetShadowFrameTexture(i uint32) {
	//Not used in Vulkan
}

func (r *Renderer_VK) SetShadowFrameCubeTexture(i uint32) {
	//Not used in Vulkan
}

func (r *Renderer_VK) SetVertexData(values ...float32) {
	const m = 0x7fffffff
	bufferIndex := int(r.vertexBufferOffset) / int(r.vertexBuffers[0].size)
	offset := r.vertexBufferOffset % r.vertexBuffers[0].size
	if (int(offset) + len(values)*4) > int(r.vertexBuffers[0].size) {
		bufferIndex += 1
		r.vertexBufferOffset = uintptr(bufferIndex * int(r.vertexBuffers[0].size))
		offset = 0
	}
	if bufferIndex >= len(r.vertexBuffers) {
		//allocate a new buffer
		log.Println("[INFO] allocate new vertex buffer")
		var vertexBuffer VulkanBuffer
		var vertexBufferMemory vk.DeviceMemory
		var err error
		vertexBuffer.size = r.vertexBuffers[0].size
		vertexBuffer.buffer, err = r.CreateBuffer(vk.DeviceSize(vertexBuffer.size), vk.BufferUsageFlags(vk.BufferUsageTransferSrcBit|vk.BufferUsageVertexBufferBit), (vk.MemoryPropertyHostVisibleBit | vk.MemoryPropertyHostCoherentBit), &vertexBufferMemory)
		if err != nil {
			panic(err)
		}
		vertexBuffer.bufferMemory = vertexBufferMemory
		var vertexData unsafe.Pointer
		vk.MapMemory(r.device, vertexBuffer.bufferMemory, 0, vk.DeviceSize(vertexBuffer.size), 0, &vertexData)
		vertexBuffer.data = vertexData
		r.vertexBuffers = append(r.vertexBuffers, vertexBuffer)
	}
	if offset == 0 {
		vk.CmdBindVertexBuffers(r.commandBuffers[0], 0, 1, []vk.Buffer{r.vertexBuffers[bufferIndex].buffer}, []vk.DeviceSize{0})
	}
	vk.Memcopy(unsafe.Pointer(uintptr(r.vertexBuffers[bufferIndex].data)+offset), (*[m]byte)(unsafe.Pointer((*sliceHeader)(unsafe.Pointer(&values)).Data))[:len(values)*4])
	r.vertexBufferOffset += uintptr(len(values) * 4)
}

func (r *Renderer_VK) SetVertexData2(cmd vk.CommandBuffer, values ...float32) {
	const m = 0x7fffffff
	bufferIndex := int(r.vertexBufferOffset) / int(r.vertexBuffers[0].size)
	offset := r.vertexBufferOffset % r.vertexBuffers[bufferIndex].size
	if (int(offset) + len(values)*4) > int(r.vertexBuffers[0].size) {
		bufferIndex += 1
		r.vertexBufferOffset = uintptr(bufferIndex * int(r.vertexBuffers[0].size))
		offset = 0
	}
	if bufferIndex >= len(r.vertexBuffers) {
		//allocate a new buffer
		log.Println("[INFO] allocate new vertex buffer")
		var vertexBuffer VulkanBuffer
		var vertexBufferMemory vk.DeviceMemory
		var err error
		vertexBuffer.size = r.vertexBuffers[0].size
		vertexBuffer.buffer, err = r.CreateBuffer(vk.DeviceSize(vertexBuffer.size), vk.BufferUsageFlags(vk.BufferUsageTransferSrcBit|vk.BufferUsageVertexBufferBit), (vk.MemoryPropertyHostVisibleBit | vk.MemoryPropertyHostCoherentBit), &vertexBufferMemory)
		if err != nil {
			panic(err)
		}
		vertexBuffer.bufferMemory = vertexBufferMemory
		var vertexData unsafe.Pointer
		vk.MapMemory(r.device, vertexBuffer.bufferMemory, 0, vk.DeviceSize(vertexBuffer.size), 0, &vertexData)
		vertexBuffer.data = vertexData
		r.vertexBuffers = append(r.vertexBuffers, vertexBuffer)
	}
	if offset == 0 {
		vk.CmdBindVertexBuffers(cmd, 0, 1, []vk.Buffer{r.vertexBuffers[bufferIndex].buffer}, []vk.DeviceSize{0})
	}
	vk.Memcopy(unsafe.Pointer(uintptr(r.vertexBuffers[bufferIndex].data)+offset), (*[m]byte)(unsafe.Pointer((*sliceHeader)(unsafe.Pointer(&values)).Data))[:len(values)*4])
	r.vertexBufferOffset += uintptr(len(values) * 4)
}

func (r *Renderer_VK) SetModelVertexData(bufferIndex uint32, values []byte) {
	if r.modelVertexBuffers[bufferIndex].size >= 0 {
		vk.DestroyBuffer(r.device, r.modelVertexBuffers[bufferIndex].buffer, nil)
		vk.FreeMemory(r.device, r.modelVertexBuffers[bufferIndex].bufferMemory, nil)
	}
	r.modelVertexBuffers[bufferIndex].size = uintptr(len(values))
	var bufferMemory vk.DeviceMemory
	var err error
	r.modelVertexBuffers[bufferIndex].buffer, err = r.CreateBuffer(vk.DeviceSize(r.modelVertexBuffers[bufferIndex].size), vk.BufferUsageFlags(vk.BufferUsageTransferDstBit|vk.BufferUsageVertexBufferBit), vk.MemoryPropertyDeviceLocalBit, &bufferMemory)
	r.modelVertexBuffers[bufferIndex].bufferMemory = bufferMemory
	if err != nil {
		panic(err)
	}
	var stagingBufferMemory vk.DeviceMemory
	var stagingBuffer vk.Buffer
	stagingBuffer, err = r.CreateBuffer(vk.DeviceSize(len(values)), vk.BufferUsageFlags(vk.BufferUsageTransferSrcBit), (vk.MemoryPropertyHostVisibleBit | vk.MemoryPropertyHostCoherentBit), &stagingBufferMemory)
	if err != nil {
		panic(err)
	}
	var stagingData unsafe.Pointer
	vk.MapMemory(r.device, stagingBufferMemory, 0, vk.DeviceSize(len(values)), 0, &stagingData)
	vk.Memcopy(stagingData, values)
	cmd := r.BeginSingleTimeCommands()
	bufferCopy := []vk.BufferCopy{{
		SrcOffset: 0,
		DstOffset: 0,
		Size:      vk.DeviceSize(len(values)),
	}}
	vk.CmdCopyBuffer(cmd, stagingBuffer, r.modelVertexBuffers[bufferIndex].buffer, 1, bufferCopy)
	r.EndSingleTimeCommands(cmd)
	vk.UnmapMemory(r.device, stagingBufferMemory)
	vk.DestroyBuffer(r.device, stagingBuffer, nil)
	vk.FreeMemory(r.device, stagingBufferMemory, nil)
}
func (r *Renderer_VK) SetModelIndexData(bufferIndex uint32, values ...uint32) {
	if r.modelIndexBuffers[bufferIndex].size >= 0 {
		vk.DestroyBuffer(r.device, r.modelIndexBuffers[bufferIndex].buffer, nil)
		vk.FreeMemory(r.device, r.modelIndexBuffers[bufferIndex].bufferMemory, nil)
	}
	r.modelIndexBuffers[bufferIndex].size = uintptr(len(values) * 4)
	var bufferMemory vk.DeviceMemory
	var err error
	r.modelIndexBuffers[bufferIndex].buffer, err = r.CreateBuffer(vk.DeviceSize(r.modelIndexBuffers[bufferIndex].size), vk.BufferUsageFlags(vk.BufferUsageTransferDstBit|vk.BufferUsageIndexBufferBit), vk.MemoryPropertyDeviceLocalBit, &bufferMemory)
	r.modelIndexBuffers[bufferIndex].bufferMemory = bufferMemory
	if err != nil {
		panic(err)
	}
	var stagingBufferMemory vk.DeviceMemory
	var stagingBuffer vk.Buffer
	stagingBuffer, err = r.CreateBuffer(vk.DeviceSize(r.modelIndexBuffers[bufferIndex].size), vk.BufferUsageFlags(vk.BufferUsageTransferSrcBit), (vk.MemoryPropertyHostVisibleBit | vk.MemoryPropertyHostCoherentBit), &stagingBufferMemory)
	if err != nil {
		panic(err)
	}
	var stagingData unsafe.Pointer
	vk.MapMemory(r.device, stagingBufferMemory, 0, vk.DeviceSize(r.modelIndexBuffers[bufferIndex].size), 0, &stagingData)
	const m = 0x7fffffff
	vk.Memcopy(stagingData, (*[m]byte)(unsafe.Pointer((*sliceHeader)(unsafe.Pointer(&values)).Data))[:r.modelIndexBuffers[bufferIndex].size])
	cmd := r.BeginSingleTimeCommands()
	bufferCopy := []vk.BufferCopy{{
		SrcOffset: 0,
		DstOffset: 0,
		Size:      vk.DeviceSize(r.modelIndexBuffers[bufferIndex].size),
	}}
	vk.CmdCopyBuffer(cmd, stagingBuffer, r.modelIndexBuffers[bufferIndex].buffer, 1, bufferCopy)
	r.EndSingleTimeCommands(cmd)
	vk.UnmapMemory(r.device, stagingBufferMemory)
	vk.DestroyBuffer(r.device, stagingBuffer, nil)
	vk.FreeMemory(r.device, stagingBufferMemory, nil)
}

func (r *Renderer_VK) RenderQuad() {
	switchedProgram := r.VKState.currentProgram != r.spriteProgram
	r.VKState.currentProgram = r.spriteProgram
	pipelineIndex, ok := r.spriteProgram.pipelineIndexMap[r.VKState.VulkanPipelineState.VulkanBlendState]
	if !ok {
		panic("Pipeline not found")
	}
	pipeline := r.spriteProgram.pipelines[pipelineIndex]
	if switchedProgram || pipeline != r.VKState.currentPipeline {
		r.VKState.currentPipeline = pipeline
		vk.CmdBindPipeline(r.commandBuffers[0], vk.PipelineBindPointGraphics, pipeline)
	}
	if switchedProgram {
		bufferIndex := int(r.vertexBufferOffset) / int(r.vertexBuffers[0].size)
		if int(r.vertexBufferOffset)%int(r.vertexBuffers[0].size) == 0 {
			bufferIndex -= 1
		}
		vk.CmdBindVertexBuffers(r.commandBuffers[0], 0, 1, []vk.Buffer{r.vertexBuffers[bufferIndex].buffer}, []vk.DeviceSize{0})
	}
	const m = 0x7fffffff
	uniformBufferSize := uint32(r.spriteProgram.uniformBuffers[0].size)
	vertOffset, ok := r.spriteProgram.uniformOffsetMap[r.VKState.VulkanSpriteProgramVertUniformBufferObject]
	if !ok {
		bufferIndex := r.spriteProgram.uniformBufferOffset / uniformBufferSize
		if int(bufferIndex) >= len(r.spriteProgram.uniformBuffers) {
			var uniformBuffer VulkanBuffer
			var err error
			uniformBuffer.size = r.spriteProgram.uniformBuffers[0].size
			uniformBuffer.buffer, err = r.CreateBuffer(vk.DeviceSize(uniformBuffer.size), vk.BufferUsageFlags(vk.BufferUsageTransferSrcBit|vk.BufferUsageUniformBufferBit), (vk.MemoryPropertyHostVisibleBit | vk.MemoryPropertyHostCoherentBit), &uniformBuffer.bufferMemory)
			if err != nil {
				panic(err)
			}
			vk.MapMemory(r.device, uniformBuffer.bufferMemory, 0, vk.DeviceSize(uniformBuffer.size), 0, &uniformBuffer.data)
			r.spriteProgram.uniformBuffers = append(r.spriteProgram.uniformBuffers, uniformBuffer)
		}
		vertOffset = r.spriteProgram.uniformBufferOffset
		r.spriteProgram.uniformOffsetMap[r.VKState.VulkanSpriteProgramVertUniformBufferObject] = vertOffset
		vk.Memcopy(unsafe.Pointer(uintptr(r.spriteProgram.uniformBuffers[bufferIndex].data)+uintptr(vertOffset%uniformBufferSize)), (*[m]byte)(unsafe.Pointer(&r.VulkanSpriteProgramVertUniformBufferObject))[:VulkanVertUniformSize])
		r.spriteProgram.uniformBufferOffset += r.spriteProgram.uniformSize
	}
	fragOffset, ok := r.spriteProgram.uniformOffsetMap[r.VKState.VulkanSpriteProgramFragUniformBufferObject]
	if !ok {
		bufferIndex := r.spriteProgram.uniformBufferOffset / uniformBufferSize
		if int(bufferIndex) >= len(r.spriteProgram.uniformBuffers) {
			var uniformBuffer VulkanBuffer
			var err error
			uniformBuffer.size = r.spriteProgram.uniformBuffers[0].size
			uniformBuffer.buffer, err = r.CreateBuffer(vk.DeviceSize(uniformBuffer.size), vk.BufferUsageFlags(vk.BufferUsageTransferSrcBit|vk.BufferUsageUniformBufferBit), (vk.MemoryPropertyHostVisibleBit | vk.MemoryPropertyHostCoherentBit), &uniformBuffer.bufferMemory)
			if err != nil {
				panic(err)
			}
			vk.MapMemory(r.device, uniformBuffer.bufferMemory, 0, vk.DeviceSize(uniformBuffer.size), 0, &uniformBuffer.data)
			r.spriteProgram.uniformBuffers = append(r.spriteProgram.uniformBuffers, uniformBuffer)
		}
		fragOffset = r.spriteProgram.uniformBufferOffset
		r.spriteProgram.uniformOffsetMap[r.VKState.VulkanSpriteProgramFragUniformBufferObject] = fragOffset
		vk.Memcopy(unsafe.Pointer(uintptr(r.spriteProgram.uniformBuffers[bufferIndex].data)+uintptr(fragOffset%uniformBufferSize)), (*[m]byte)(unsafe.Pointer(&r.VulkanSpriteProgramFragUniformBufferObject))[:VulkanFragUniformSize])
		r.spriteProgram.uniformBufferOffset += r.spriteProgram.uniformSize
	}
	descriptorWrites := make([]vk.WriteDescriptorSet, 0, 4)
	if switchedProgram || vertOffset != r.VKState.spriteVertUniformBufferOffset {
		r.VKState.spriteVertUniformBufferOffset = vertOffset
		vertUniformInfo := []vk.DescriptorBufferInfo{{
			Buffer: r.spriteProgram.uniformBuffers[vertOffset/uniformBufferSize].buffer,
			Offset: vk.DeviceSize(vertOffset % uniformBufferSize),
			Range:  vk.DeviceSize(r.spriteProgram.uniformSize),
		}}
		descriptorWrites = append(descriptorWrites, vk.WriteDescriptorSet{
			SType:           vk.StructureTypeWriteDescriptorSet,
			DstBinding:      0,
			DstArrayElement: 0,
			DescriptorCount: 1,
			DescriptorType:  vk.DescriptorTypeUniformBuffer,
			PBufferInfo:     vertUniformInfo,
		})
	}
	if switchedProgram || fragOffset != r.VKState.spriteFragUniformBufferOffset {
		r.VKState.spriteFragUniformBufferOffset = fragOffset
		fragUniformInfo := []vk.DescriptorBufferInfo{{
			Buffer: r.spriteProgram.uniformBuffers[fragOffset/uniformBufferSize].buffer,
			Offset: vk.DeviceSize(fragOffset % uniformBufferSize),
			Range:  vk.DeviceSize(r.spriteProgram.uniformSize),
		}}
		descriptorWrites = append(descriptorWrites, vk.WriteDescriptorSet{
			SType:           vk.StructureTypeWriteDescriptorSet,
			DstBinding:      1,
			DstArrayElement: 0,
			DescriptorCount: 1,
			DescriptorType:  vk.DescriptorTypeUniformBuffer,
			PBufferInfo:     fragUniformInfo,
		})
	}
	if switchedProgram || r.VKState.spriteTexture != r.currentSpriteTexture.spriteTexture {
		r.currentSpriteTexture.spriteTexture = r.VKState.spriteTexture
		imageInfo := []vk.DescriptorImageInfo{
			{
				ImageLayout: vk.ImageLayoutShaderReadOnlyOptimal,
				ImageView:   r.VKState.spriteTexture.imageView,
				Sampler:     r.spriteSamplers[0],
			},
		}
		if r.spriteTexture.filter {
			imageInfo[0].Sampler = r.spriteSamplers[1]
		}
		descriptorWrites = append(descriptorWrites, vk.WriteDescriptorSet{
			SType:           vk.StructureTypeWriteDescriptorSet,
			DstBinding:      2,
			DstArrayElement: 0,
			DescriptorCount: 1,
			DescriptorType:  vk.DescriptorTypeCombinedImageSampler,
			PImageInfo:      imageInfo,
		})
	}
	// MacOS MoltenVK workaround: push constants need to be set every draw call
	if switchedProgram || r.VKState.palTexture != r.currentSpriteTexture.palTexture || runtime.GOOS == "darwin" {
		if switchedProgram || r.VKState.palTexture.img != r.currentSpriteTexture.palTexture.img || r.VKState.palTexture.offset[0] != r.currentSpriteTexture.palTexture.offset[0] || r.VKState.palTexture.offset[1] != r.currentSpriteTexture.palTexture.offset[1] || runtime.GOOS == "darwin" {
			uvst := r.VKState.palTexture.uvst
			vk.CmdPushConstants(r.commandBuffers[0], r.spriteProgram.pipelineLayout, vk.ShaderStageFlags(vk.ShaderStageFragmentBit), 0, 16, unsafe.Pointer(&uvst))
		}
		if switchedProgram || r.VKState.palTexture.img != r.currentSpriteTexture.palTexture.img {
			palImageInfo := []vk.DescriptorImageInfo{
				{
					ImageLayout: vk.ImageLayoutShaderReadOnlyOptimal,
					ImageView:   r.VKState.palTexture.imageView,
					Sampler:     r.spriteSamplers[0],
				},
			}
			descriptorWrites = append(descriptorWrites, vk.WriteDescriptorSet{
				SType:           vk.StructureTypeWriteDescriptorSet,
				DstBinding:      3,
				DstArrayElement: 0,
				DescriptorCount: 1,
				DescriptorType:  vk.DescriptorTypeCombinedImageSampler,
				PImageInfo:      palImageInfo,
			})
		}
		r.currentSpriteTexture.palTexture = r.VKState.palTexture
	}
	if len(descriptorWrites) > 0 {
		vk.CmdPushDescriptorSet(r.commandBuffers[0], vk.PipelineBindPointGraphics, r.spriteProgram.pipelineLayout, 0, uint32(len(descriptorWrites)), descriptorWrites)
	}
	viewports := []vk.Viewport{{
		MinDepth: 0.0,
		MaxDepth: 1.0,
		X:        0,
		Y:        0,
		Width:    float32(sys.scrrect[2]),
		Height:   float32(sys.scrrect[3]),
	}}
	scissors := []vk.Rect2D{r.VKState.scissor}
	vk.CmdSetViewport(r.commandBuffers[0], 0, 1, viewports)
	vk.CmdSetScissor(r.commandBuffers[0], 0, 1, scissors)
	vk.CmdDraw(r.commandBuffers[0], 4, 1, uint32(r.vertexBufferOffset%r.vertexBuffers[0].size)/16-4, 0)
}
func (r *Renderer_VK) RenderElements(mode PrimitiveMode, count, offset int) {
	switchedProgram := r.VKState.currentProgram != r.modelProgram
	r.VKState.currentProgram = r.modelProgram

	r.VKState.VulkanModelPipelineState.primitiveMode = mode
	pipeline := r.GetModelPipeline(&r.VKState.VulkanPipelineState)
	if switchedProgram || pipeline != r.VKState.currentPipeline {
		r.VKState.currentPipeline = pipeline
		vk.CmdBindPipeline(r.commandBuffers[0], vk.PipelineBindPointGraphics, pipeline)
	}
	vertexBuffers := make([]vk.Buffer, 0, 3)
	offsets := make([]vk.DeviceSize, 0, 3)
	bufferOffset := r.VKState.modelVertAttrOffset
	vertexAttributeUsage := []bool{true, true, r.VKState.VulkanModelPipelineState.useUV, r.VKState.VulkanModelPipelineState.useNormal, r.VKState.VulkanModelPipelineState.useTangent, r.VKState.VulkanModelPipelineState.useVertColor, r.VKState.VulkanModelPipelineState.useJoint0, r.VKState.VulkanModelPipelineState.useJoint0, r.VKState.VulkanModelPipelineState.useJoint1, r.VKState.VulkanModelPipelineState.useJoint1, r.VKState.VulkanModelPipelineState.useOutlineAttribute}
	vertexAttributeOffsets := []uint32{1, 3, 2, 3, 4, 4, 4, 4, 4, 4, 4}
	for i := 0; i < len(vertexAttributeOffsets); i++ {
		if vertexAttributeUsage[i] {
			vertexBuffers = append(vertexBuffers, r.modelVertexBuffers[r.VKState.modelVertexBufferIndex].buffer)
			offsets = append(offsets, vk.DeviceSize(bufferOffset))
			bufferOffset += uint32(r.VKState.VulkanModelProgramUniformBufferObject2.numVertices*4) * vertexAttributeOffsets[i]
		}
	}
	const m = 0x7fffffff
	vk.CmdBindVertexBuffers(r.commandBuffers[0], 0, uint32(len(vertexBuffers)), vertexBuffers, offsets)
	descriptorWrites := make([]vk.WriteDescriptorSet, 0, 13)
	if switchedProgram {
		if len(r.modelProgram.uniformOffsetMap) == 0 {
			vk.Memcopy(unsafe.Pointer(uintptr(r.modelProgram.uniformBuffers[0].data)), (*[m]byte)(unsafe.Pointer(&r.VKState.VulkanModelProgramUniformBufferObject0))[:VulkanModelUniform0Size])
			r.modelProgram.uniformOffsetMap[r.VKState.VulkanModelProgramUniformBufferObject0] = 0
			r.modelProgram.uniformBufferOffset += r.modelProgram.uniformSize
		}
		descriptorWrites = append(descriptorWrites,
			vk.WriteDescriptorSet{
				SType:           vk.StructureTypeWriteDescriptorSet,
				DstBinding:      0,
				DstArrayElement: 0,
				DescriptorCount: 1,
				DescriptorType:  vk.DescriptorTypeUniformBuffer,
				PBufferInfo: []vk.DescriptorBufferInfo{{
					Buffer: r.modelProgram.uniformBuffers[0].buffer,
					Offset: vk.DeviceSize(0),
					Range:  vk.DeviceSize(r.modelProgram.uniformSize),
				}},
			},
			vk.WriteDescriptorSet{
				SType:           vk.StructureTypeWriteDescriptorSet,
				DstBinding:      5,
				DstArrayElement: 0,
				DescriptorCount: 1,
				DescriptorType:  vk.DescriptorTypeCombinedImageSampler,
				PImageInfo: []vk.DescriptorImageInfo{{
					ImageLayout: vk.ImageLayoutShaderReadOnlyOptimal,
					ImageView:   r.VKState.lambertianEnvSampler.imageView,
					Sampler:     r.VKState.lambertianEnvSampler.sampler,
				}},
			},
			vk.WriteDescriptorSet{
				SType:           vk.StructureTypeWriteDescriptorSet,
				DstBinding:      6,
				DstArrayElement: 0,
				DescriptorCount: 1,
				DescriptorType:  vk.DescriptorTypeCombinedImageSampler,
				PImageInfo: []vk.DescriptorImageInfo{{
					ImageLayout: vk.ImageLayoutShaderReadOnlyOptimal,
					ImageView:   r.VKState.GGXEnvSampler.imageView,
					Sampler:     r.VKState.GGXEnvSampler.sampler,
				}},
			},
			vk.WriteDescriptorSet{
				SType:           vk.StructureTypeWriteDescriptorSet,
				DstBinding:      7,
				DstArrayElement: 0,
				DescriptorCount: 1,
				DescriptorType:  vk.DescriptorTypeCombinedImageSampler,
				PImageInfo: []vk.DescriptorImageInfo{{
					ImageLayout: vk.ImageLayoutShaderReadOnlyOptimal,
					ImageView:   r.VKState.GGXLUT.imageView,
					Sampler:     r.VKState.GGXLUT.sampler,
				}},
			},
			vk.WriteDescriptorSet{
				SType:           vk.StructureTypeWriteDescriptorSet,
				DstBinding:      13,
				DstArrayElement: 0,
				DescriptorCount: 1,
				DescriptorType:  vk.DescriptorTypeCombinedImageSampler,
				PImageInfo: []vk.DescriptorImageInfo{{
					ImageLayout: vk.ImageLayoutShaderReadOnlyOptimal,
					ImageView:   r.shadowMapTextures.imageView,
					Sampler:     r.shadowMapTextures.sampler,
				}},
			},
		)
	}
	uniformBufferSize := uint32(r.modelProgram.uniformBuffers[0].size)
	uniformOffset, ok := r.modelProgram.uniformOffsetMap[r.VKState.VulkanModelProgramUniformBufferObject1]
	if !ok {
		uniformOffset = r.modelProgram.uniformBufferOffset
		if int(uniformOffset/uniformBufferSize) >= len(r.modelProgram.uniformBuffers) {
			var uniformBuffer VulkanBuffer
			var err error
			uniformBuffer.size = r.modelProgram.uniformBuffers[0].size
			uniformBuffer.buffer, err = r.CreateBuffer(vk.DeviceSize(uniformBuffer.size), vk.BufferUsageFlags(vk.BufferUsageTransferSrcBit|vk.BufferUsageUniformBufferBit), (vk.MemoryPropertyHostVisibleBit | vk.MemoryPropertyHostCoherentBit), &uniformBuffer.bufferMemory)
			if err != nil {
				panic(err)
			}
			vk.MapMemory(r.device, uniformBuffer.bufferMemory, 0, vk.DeviceSize(uniformBuffer.size), 0, &uniformBuffer.data)
			r.modelProgram.uniformBuffers = append(r.modelProgram.uniformBuffers, uniformBuffer)
		}
		r.modelProgram.uniformOffsetMap[r.VKState.VulkanModelProgramUniformBufferObject1] = uniformOffset
		vk.Memcopy(unsafe.Pointer(uintptr(r.modelProgram.uniformBuffers[uniformOffset/uniformBufferSize].data)+uintptr(uniformOffset%uniformBufferSize)), (*[m]byte)(unsafe.Pointer(&r.VulkanModelProgramUniformBufferObject1))[:VulkanModelUniform1Size])
		r.modelProgram.uniformBufferOffset += r.modelProgram.uniformSize
	}
	if switchedProgram || r.VKState.modelUniformBufferOffset1 != uniformOffset {
		descriptorWrites = append(descriptorWrites,
			vk.WriteDescriptorSet{
				SType:           vk.StructureTypeWriteDescriptorSet,
				DstBinding:      1,
				DstArrayElement: 0,
				DescriptorCount: 1,
				DescriptorType:  vk.DescriptorTypeUniformBuffer,
				PBufferInfo: []vk.DescriptorBufferInfo{{
					Buffer: r.modelProgram.uniformBuffers[uniformOffset/uniformBufferSize].buffer,
					Offset: vk.DeviceSize(uniformOffset % uniformBufferSize),
					Range:  vk.DeviceSize(r.modelProgram.uniformSize),
				}},
			},
		)
		r.VKState.modelUniformBufferOffset1 = uniformOffset
	}

	uniformOffset2, ok := r.modelProgram.uniformOffsetMap[r.VKState.VulkanModelProgramUniformBufferObject2]
	if !ok {
		uniformOffset2 = r.modelProgram.uniformBufferOffset
		if int(uniformOffset2/uniformBufferSize) >= len(r.modelProgram.uniformBuffers) {
			var uniformBuffer VulkanBuffer
			var err error
			uniformBuffer.size = r.modelProgram.uniformBuffers[0].size
			uniformBuffer.buffer, err = r.CreateBuffer(vk.DeviceSize(uniformBuffer.size), vk.BufferUsageFlags(vk.BufferUsageTransferSrcBit|vk.BufferUsageUniformBufferBit), (vk.MemoryPropertyHostVisibleBit | vk.MemoryPropertyHostCoherentBit), &uniformBuffer.bufferMemory)
			if err != nil {
				panic(err)
			}
			vk.MapMemory(r.device, uniformBuffer.bufferMemory, 0, vk.DeviceSize(uniformBuffer.size), 0, &uniformBuffer.data)
			r.modelProgram.uniformBuffers = append(r.modelProgram.uniformBuffers, uniformBuffer)
		}
		r.modelProgram.uniformOffsetMap[r.VKState.VulkanModelProgramUniformBufferObject2] = uniformOffset2
		vk.Memcopy(unsafe.Pointer(uintptr(r.modelProgram.uniformBuffers[uniformOffset2/uniformBufferSize].data)+uintptr(uniformOffset2%uniformBufferSize)), (*[m]byte)(unsafe.Pointer(&r.VulkanModelProgramUniformBufferObject2))[:VulkanModelUniform2Size])
		r.modelProgram.uniformBufferOffset += r.modelProgram.uniformSize
	}

	if switchedProgram || r.VKState.modelUniformBufferOffset2 != uniformOffset2 {
		descriptorWrites = append(descriptorWrites,
			vk.WriteDescriptorSet{
				SType:           vk.StructureTypeWriteDescriptorSet,
				DstBinding:      2,
				DstArrayElement: 0,
				DescriptorCount: 1,
				DescriptorType:  vk.DescriptorTypeUniformBuffer,
				PBufferInfo: []vk.DescriptorBufferInfo{{
					Buffer: r.modelProgram.uniformBuffers[uniformOffset2/uniformBufferSize].buffer,
					Offset: vk.DeviceSize(uniformOffset2 % uniformBufferSize),
					Range:  vk.DeviceSize(r.modelProgram.uniformSize),
				}},
			},
		)
		r.VKState.modelUniformBufferOffset2 = uniformOffset2
	}
	if switchedProgram || r.VKState.currentModelTexture.jointMatricesTexture != r.VKState.VulkanModelTexture.jointMatricesTexture {
		r.VKState.currentModelTexture.jointMatricesTexture = r.VKState.VulkanModelTexture.jointMatricesTexture
		descriptorWrites = append(descriptorWrites,
			vk.WriteDescriptorSet{
				SType:           vk.StructureTypeWriteDescriptorSet,
				DstBinding:      3,
				DstArrayElement: 0,
				DescriptorCount: 1,
				DescriptorType:  vk.DescriptorTypeCombinedImageSampler,
				PImageInfo: []vk.DescriptorImageInfo{{
					ImageLayout: vk.ImageLayoutShaderReadOnlyOptimal,
					ImageView:   r.VKState.VulkanModelTexture.jointMatricesTexture.imageView,
					Sampler:     r.VKState.VulkanModelTexture.jointMatricesTexture.sampler,
				}},
			},
		)
	}
	if switchedProgram || r.VKState.currentModelTexture.morphTargetTexture != r.VKState.VulkanModelTexture.morphTargetTexture {
		r.VKState.currentModelTexture.morphTargetTexture = r.VKState.VulkanModelTexture.morphTargetTexture
		descriptorWrites = append(descriptorWrites,
			vk.WriteDescriptorSet{
				SType:           vk.StructureTypeWriteDescriptorSet,
				DstBinding:      4,
				DstArrayElement: 0,
				DescriptorCount: 1,
				DescriptorType:  vk.DescriptorTypeCombinedImageSampler,
				PImageInfo: []vk.DescriptorImageInfo{{
					ImageLayout: vk.ImageLayoutShaderReadOnlyOptimal,
					ImageView:   r.VKState.VulkanModelTexture.morphTargetTexture.imageView,
					Sampler:     r.VKState.VulkanModelTexture.morphTargetTexture.sampler,
				}},
			},
		)
	}
	if switchedProgram || r.VKState.currentModelTexture.tex != r.VKState.VulkanModelTexture.tex {
		r.VKState.currentModelTexture.tex = r.VKState.VulkanModelTexture.tex
		descriptorWrites = append(descriptorWrites,
			vk.WriteDescriptorSet{
				SType:           vk.StructureTypeWriteDescriptorSet,
				DstBinding:      8,
				DstArrayElement: 0,
				DescriptorCount: 1,
				DescriptorType:  vk.DescriptorTypeCombinedImageSampler,
				PImageInfo: []vk.DescriptorImageInfo{{
					ImageLayout: vk.ImageLayoutShaderReadOnlyOptimal,
					ImageView:   r.VKState.VulkanModelTexture.tex.imageView,
					Sampler:     r.VKState.VulkanModelTexture.tex.sampler,
				}},
			},
		)
	}
	if switchedProgram || r.VKState.currentModelTexture.normalMap != r.VKState.VulkanModelTexture.normalMap {
		r.VKState.currentModelTexture.normalMap = r.VKState.VulkanModelTexture.normalMap
		descriptorWrites = append(descriptorWrites,
			vk.WriteDescriptorSet{
				SType:           vk.StructureTypeWriteDescriptorSet,
				DstBinding:      9,
				DstArrayElement: 0,
				DescriptorCount: 1,
				DescriptorType:  vk.DescriptorTypeCombinedImageSampler,
				PImageInfo: []vk.DescriptorImageInfo{{
					ImageLayout: vk.ImageLayoutShaderReadOnlyOptimal,
					ImageView:   r.VKState.VulkanModelTexture.normalMap.imageView,
					Sampler:     r.VKState.VulkanModelTexture.normalMap.sampler,
				}},
			},
		)
	}
	if switchedProgram || r.VKState.currentModelTexture.metallicRoughnessMap != r.VKState.VulkanModelTexture.metallicRoughnessMap {
		r.VKState.currentModelTexture.metallicRoughnessMap = r.VKState.VulkanModelTexture.metallicRoughnessMap
		descriptorWrites = append(descriptorWrites,
			vk.WriteDescriptorSet{
				SType:           vk.StructureTypeWriteDescriptorSet,
				DstBinding:      10,
				DstArrayElement: 0,
				DescriptorCount: 1,
				DescriptorType:  vk.DescriptorTypeCombinedImageSampler,
				PImageInfo: []vk.DescriptorImageInfo{{
					ImageLayout: vk.ImageLayoutShaderReadOnlyOptimal,
					ImageView:   r.VKState.VulkanModelTexture.metallicRoughnessMap.imageView,
					Sampler:     r.VKState.VulkanModelTexture.metallicRoughnessMap.sampler,
				}},
			},
		)
	}
	if switchedProgram || r.VKState.currentModelTexture.ambientOcclusionMap != r.VKState.VulkanModelTexture.ambientOcclusionMap {
		r.VKState.currentModelTexture.ambientOcclusionMap = r.VKState.ambientOcclusionMap
		descriptorWrites = append(descriptorWrites,
			vk.WriteDescriptorSet{
				SType:           vk.StructureTypeWriteDescriptorSet,
				DstBinding:      11,
				DstArrayElement: 0,
				DescriptorCount: 1,
				DescriptorType:  vk.DescriptorTypeCombinedImageSampler,
				PImageInfo: []vk.DescriptorImageInfo{{
					ImageLayout: vk.ImageLayoutShaderReadOnlyOptimal,
					ImageView:   r.VKState.VulkanModelTexture.ambientOcclusionMap.imageView,
					Sampler:     r.VKState.VulkanModelTexture.ambientOcclusionMap.sampler,
				}},
			},
		)
	}
	if switchedProgram || r.VKState.currentModelTexture.emissionMap != r.VKState.VulkanModelTexture.emissionMap {
		r.VKState.currentModelTexture.emissionMap = r.VKState.emissionMap
		descriptorWrites = append(descriptorWrites,
			vk.WriteDescriptorSet{
				SType:           vk.StructureTypeWriteDescriptorSet,
				DstBinding:      12,
				DstArrayElement: 0,
				DescriptorCount: 1,
				DescriptorType:  vk.DescriptorTypeCombinedImageSampler,
				PImageInfo: []vk.DescriptorImageInfo{{
					ImageLayout: vk.ImageLayoutShaderReadOnlyOptimal,
					ImageView:   r.VKState.VulkanModelTexture.emissionMap.imageView,
					Sampler:     r.VKState.VulkanModelTexture.emissionMap.sampler,
				}},
			},
		)
	}
	if len(descriptorWrites) > 0 {
		vk.CmdPushDescriptorSet(r.commandBuffers[0], vk.PipelineBindPointGraphics, r.modelProgram.pipelineLayout, 0, uint32(len(descriptorWrites)), descriptorWrites)
	}
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
			X: 0,
			Y: 0,
		},
		Extent: vk.Extent2D{
			Width:  uint32(sys.scrrect[2]),
			Height: uint32(sys.scrrect[3]),
		},
	}}
	vk.CmdSetViewport(r.commandBuffers[0], 0, 1, viewports)
	vk.CmdSetScissor(r.commandBuffers[0], 0, 1, scissors)

	vk.CmdDrawIndexed(r.commandBuffers[0], uint32(count), 1, uint32(offset/4), 0, 0)
}

func (r *Renderer_VK) RenderShadowMapElements(mode PrimitiveMode, count, offset int) {
	index := r.VKState.VulkanShadowMapProgramPushConstant.lightIndex
	if index == 0 {
		r.VKState.primitiveMode = mode
		pipeline := r.GetShadowMapPipeline(&r.VKState.VulkanShadowMapPipelineState)
		if pipeline != r.VKState.currentShadowmMapPipeline {
			r.VKState.currentShadowmMapPipeline = pipeline
			vk.CmdBindPipeline(r.commandBuffers[1], vk.PipelineBindPointGraphics, pipeline)
		}

		vertexBuffers := make([]vk.Buffer, 0, 3)
		offsets := make([]vk.DeviceSize, 0, 3)
		bufferOffset := r.VKState.shadowMapVertAttrOffset
		vertexAttributeUsage := []bool{true, true, r.VKState.VulkanShadowMapPipelineState.useUV, r.VKState.VulkanShadowMapPipelineState.useNormal, r.VKState.VulkanShadowMapPipelineState.useTangent, r.VKState.VulkanShadowMapPipelineState.useVertColor, r.VKState.VulkanShadowMapPipelineState.useJoint0, r.VKState.VulkanShadowMapPipelineState.useJoint0, r.VKState.VulkanShadowMapPipelineState.useJoint1, r.VKState.VulkanShadowMapPipelineState.useJoint1, r.VKState.VulkanShadowMapPipelineState.useOutlineAttribute}
		ActualVertexAttributeUsage := []bool{true, true, r.VKState.VulkanShadowMapPipelineState.useUV, false, false, r.VKState.VulkanShadowMapPipelineState.useVertColor, r.VKState.VulkanShadowMapPipelineState.useJoint0, r.VKState.VulkanShadowMapPipelineState.useJoint0, r.VKState.VulkanShadowMapPipelineState.useJoint1, r.VKState.VulkanShadowMapPipelineState.useJoint1, false}
		vertexAttributeOffsets := []uint32{1, 3, 2, 3, 4, 4, 4, 4, 4, 4, 4}
		for i := 0; i < len(vertexAttributeOffsets); i++ {
			if vertexAttributeUsage[i] {
				if !ActualVertexAttributeUsage[i] {
					bufferOffset += uint32(r.VKState.VulkanShadowMapProgramPushConstant.numVertices*4) * vertexAttributeOffsets[i]
					continue
				}
				vertexBuffers = append(vertexBuffers, r.modelVertexBuffers[r.VKState.modelVertexBufferIndex].buffer)
				offsets = append(offsets, vk.DeviceSize(bufferOffset))
				bufferOffset += uint32(r.VKState.VulkanShadowMapProgramPushConstant.numVertices*4) * vertexAttributeOffsets[i]
			}
		}
		const m = 0x7fffffff
		vk.CmdBindVertexBuffers(r.commandBuffers[1], 0, uint32(len(vertexBuffers)), vertexBuffers, offsets)
		descriptorWrites := make([]vk.WriteDescriptorSet, 0, 13)
		if len(r.shadowMapProgram.uniformOffsetMap) == 0 {
			j := 0
			for i := 0; i < 4; i++ {
				if r.VKState.VulkanShadowMapProgramUniformBufferObject0.lights[i].lightType == 2 {
					for k := 0; k < 6; k++ {
						r.VKState.VulkanShadowMapProgramUniformBufferObject0.layers[j] = float32(i*6 + k)
						j += 1
					}
				} else if r.VKState.VulkanShadowMapProgramUniformBufferObject0.lights[i].lightType != 0 {
					r.VKState.VulkanShadowMapProgramUniformBufferObject0.layers[j] = float32(i * 6)
					j += 1
				}
			}
			r.VKState.numShadowMapLayers = uint32(j)
			uniformSize := r.AlignUniformSize(uint32(VulkanShadowMapUniform0Size))
			vk.Memcopy(unsafe.Pointer(uintptr(r.shadowMapProgram.uniformBuffers[0].data)), (*[m]byte)(unsafe.Pointer(&r.VKState.VulkanShadowMapProgramUniformBufferObject0))[:VulkanShadowMapUniform0Size])
			r.shadowMapProgram.uniformOffsetMap[r.VKState.VulkanShadowMapProgramUniformBufferObject0] = 0
			r.shadowMapProgram.uniformBufferOffset = uniformSize
			descriptorWrites = append(descriptorWrites,
				vk.WriteDescriptorSet{
					SType:           vk.StructureTypeWriteDescriptorSet,
					DstBinding:      0,
					DstArrayElement: 0,
					DescriptorCount: 1,
					DescriptorType:  vk.DescriptorTypeUniformBuffer,
					PBufferInfo: []vk.DescriptorBufferInfo{{
						Buffer: r.shadowMapProgram.uniformBuffers[0].buffer,
						Offset: vk.DeviceSize(0),
						Range:  vk.DeviceSize(uniformSize),
					}},
				},
			)

		}
		uniformBufferSize := uint32(r.shadowMapProgram.uniformBuffers[0].size)
		uniformOffset, ok := r.shadowMapProgram.uniformOffsetMap[r.VKState.VulkanShadowMapProgramUniformBufferObject1]
		if !ok {
			uniformOffset = r.shadowMapProgram.uniformBufferOffset
			if int(uniformOffset/uniformBufferSize) >= len(r.shadowMapProgram.uniformBuffers) || ((uniformOffset%uniformBufferSize)+r.shadowMapProgram.uniformSize) > uniformBufferSize {
				r.shadowMapProgram.uniformBufferOffset = uint32(len(r.shadowMapProgram.uniformBuffers)) * uniformBufferSize
				uniformOffset = r.shadowMapProgram.uniformBufferOffset
				var uniformBuffer VulkanBuffer
				var err error
				uniformBuffer.size = r.shadowMapProgram.uniformBuffers[0].size
				uniformBuffer.buffer, err = r.CreateBuffer(vk.DeviceSize(uniformBuffer.size), vk.BufferUsageFlags(vk.BufferUsageTransferSrcBit|vk.BufferUsageUniformBufferBit), (vk.MemoryPropertyHostVisibleBit | vk.MemoryPropertyHostCoherentBit), &uniformBuffer.bufferMemory)
				if err != nil {
					panic(err)
				}
				vk.MapMemory(r.device, uniformBuffer.bufferMemory, 0, vk.DeviceSize(uniformBuffer.size), 0, &uniformBuffer.data)
				r.shadowMapProgram.uniformBuffers = append(r.shadowMapProgram.uniformBuffers, uniformBuffer)
			}
			r.shadowMapProgram.uniformOffsetMap[r.VKState.VulkanShadowMapProgramUniformBufferObject1] = uniformOffset
			vk.Memcopy(unsafe.Pointer(uintptr(r.shadowMapProgram.uniformBuffers[r.shadowMapProgram.uniformBufferOffset/uniformBufferSize].data)+uintptr(uniformOffset%uniformBufferSize)), (*[m]byte)(unsafe.Pointer(&r.VKState.VulkanShadowMapProgramUniformBufferObject1))[:VulkanShadowMapUniform1Size])
			r.shadowMapProgram.uniformBufferOffset += r.shadowMapProgram.uniformSize
		}
		if r.VKState.shadowMapUniformBufferOffset1 != uniformOffset {
			descriptorWrites = append(descriptorWrites,
				vk.WriteDescriptorSet{
					SType:           vk.StructureTypeWriteDescriptorSet,
					DstBinding:      1,
					DstArrayElement: 0,
					DescriptorCount: 1,
					DescriptorType:  vk.DescriptorTypeUniformBuffer,
					PBufferInfo: []vk.DescriptorBufferInfo{{
						Buffer: r.shadowMapProgram.uniformBuffers[uniformOffset/uniformBufferSize].buffer,
						Offset: vk.DeviceSize(uniformOffset % uniformBufferSize),
						Range:  vk.DeviceSize(r.shadowMapProgram.uniformSize),
					}},
				},
			)
			r.VKState.shadowMapUniformBufferOffset1 = uniformOffset
		}

		uniformOffset2, ok := r.shadowMapProgram.uniformOffsetMap[r.VKState.VulkanShadowMapProgramUniformBufferObject2]
		if !ok {
			uniformOffset2 = r.shadowMapProgram.uniformBufferOffset
			if int(uniformOffset2/uniformBufferSize) >= len(r.shadowMapProgram.uniformBuffers) || ((uniformOffset%uniformBufferSize)+r.shadowMapProgram.uniformSize) > uniformBufferSize {
				r.shadowMapProgram.uniformBufferOffset = uint32(len(r.shadowMapProgram.uniformBuffers)) * uniformBufferSize
				uniformOffset = r.shadowMapProgram.uniformBufferOffset
				var uniformBuffer VulkanBuffer
				var err error
				uniformBuffer.size = r.shadowMapProgram.uniformBuffers[0].size
				uniformBuffer.buffer, err = r.CreateBuffer(vk.DeviceSize(uniformBuffer.size), vk.BufferUsageFlags(vk.BufferUsageTransferSrcBit|vk.BufferUsageUniformBufferBit), (vk.MemoryPropertyHostVisibleBit | vk.MemoryPropertyHostCoherentBit), &uniformBuffer.bufferMemory)
				if err != nil {
					panic(err)
				}
				vk.MapMemory(r.device, uniformBuffer.bufferMemory, 0, vk.DeviceSize(uniformBuffer.size), 0, &uniformBuffer.data)
				r.shadowMapProgram.uniformBuffers = append(r.shadowMapProgram.uniformBuffers, uniformBuffer)
			}
			r.shadowMapProgram.uniformOffsetMap[r.VKState.VulkanShadowMapProgramUniformBufferObject2] = uniformOffset2
			vk.Memcopy(unsafe.Pointer(uintptr(r.shadowMapProgram.uniformBuffers[r.shadowMapProgram.uniformBufferOffset/uniformBufferSize].data)+uintptr(uniformOffset2%uniformBufferSize)), (*[m]byte)(unsafe.Pointer(&r.VKState.VulkanShadowMapProgramUniformBufferObject2))[:VulkanShadowMapUniform2Size])
			r.shadowMapProgram.uniformBufferOffset += r.shadowMapProgram.uniformSize
		}

		if r.VKState.shadowMapUniformBufferOffset2 != uniformOffset2 {
			descriptorWrites = append(descriptorWrites,
				vk.WriteDescriptorSet{
					SType:           vk.StructureTypeWriteDescriptorSet,
					DstBinding:      2,
					DstArrayElement: 0,
					DescriptorCount: 1,
					DescriptorType:  vk.DescriptorTypeUniformBuffer,
					PBufferInfo: []vk.DescriptorBufferInfo{{
						Buffer: r.shadowMapProgram.uniformBuffers[uniformOffset2/uniformBufferSize].buffer,
						Offset: vk.DeviceSize(uniformOffset2 % uniformBufferSize),
						Range:  vk.DeviceSize(r.shadowMapProgram.uniformSize),
					}},
				},
			)
			r.VKState.shadowMapUniformBufferOffset2 = uniformOffset2
		}
		if r.VKState.currentShadowMapTexture.jointMatricesTexture != r.VKState.VulkanShadowMapTexture.jointMatricesTexture {
			r.VKState.currentShadowMapTexture.jointMatricesTexture = r.VKState.VulkanShadowMapTexture.jointMatricesTexture
			descriptorWrites = append(descriptorWrites,
				vk.WriteDescriptorSet{
					SType:           vk.StructureTypeWriteDescriptorSet,
					DstBinding:      3,
					DstArrayElement: 0,
					DescriptorCount: 1,
					DescriptorType:  vk.DescriptorTypeCombinedImageSampler,
					PImageInfo: []vk.DescriptorImageInfo{{
						ImageLayout: vk.ImageLayoutShaderReadOnlyOptimal,
						ImageView:   r.VKState.VulkanShadowMapTexture.jointMatricesTexture.imageView,
						Sampler:     r.VKState.VulkanShadowMapTexture.jointMatricesTexture.sampler,
					}},
				},
			)
		}
		if r.VKState.currentShadowMapTexture.morphTargetTexture != r.VKState.VulkanShadowMapTexture.morphTargetTexture {
			r.VKState.currentShadowMapTexture.morphTargetTexture = r.VKState.VulkanShadowMapTexture.morphTargetTexture
			descriptorWrites = append(descriptorWrites,
				vk.WriteDescriptorSet{
					SType:           vk.StructureTypeWriteDescriptorSet,
					DstBinding:      4,
					DstArrayElement: 0,
					DescriptorCount: 1,
					DescriptorType:  vk.DescriptorTypeCombinedImageSampler,
					PImageInfo: []vk.DescriptorImageInfo{{
						ImageLayout: vk.ImageLayoutShaderReadOnlyOptimal,
						ImageView:   r.VKState.VulkanShadowMapTexture.morphTargetTexture.imageView,
						Sampler:     r.VKState.VulkanShadowMapTexture.morphTargetTexture.sampler,
					}},
				},
			)
		}
		if r.VKState.currentShadowMapTexture.tex != r.VKState.VulkanShadowMapTexture.tex {
			r.VKState.currentShadowMapTexture.tex = r.VKState.VulkanShadowMapTexture.tex
			descriptorWrites = append(descriptorWrites,
				vk.WriteDescriptorSet{
					SType:           vk.StructureTypeWriteDescriptorSet,
					DstBinding:      5,
					DstArrayElement: 0,
					DescriptorCount: 1,
					DescriptorType:  vk.DescriptorTypeCombinedImageSampler,
					PImageInfo: []vk.DescriptorImageInfo{{
						ImageLayout: vk.ImageLayoutShaderReadOnlyOptimal,
						ImageView:   r.VKState.VulkanShadowMapTexture.tex.imageView,
						Sampler:     r.VKState.VulkanShadowMapTexture.tex.sampler,
					}},
				},
			)
		}
		if len(descriptorWrites) > 0 {
			vk.CmdPushDescriptorSet(r.commandBuffers[1], vk.PipelineBindPointGraphics, r.shadowMapProgram.pipelineLayout, 0, uint32(len(descriptorWrites)), descriptorWrites)
		}
		vertConstants := r.VKState.VulkanShadowMapProgramPushConstant
		vk.CmdPushConstants(r.commandBuffers[1], r.shadowMapProgram.pipelineLayout, vk.ShaderStageFlags(vk.ShaderStageVertexBit), 0, 68, unsafe.Pointer(&vertConstants))
		vk.CmdDrawIndexed(r.commandBuffers[1], uint32(count), r.VKState.numShadowMapLayers, uint32(offset/4), 0, 0)
	}
}

func (r *Renderer_VK) RenderCubeMap(envTex Texture, cubeTex Texture) {
	if len(r.stagingImageBarriers[0]) > 0 || len(r.tempCommands) > 0 {
		r.FlushTempCommands()
	}
	envTexture := envTex.(*Texture_VK)
	cubeTexture := cubeTex.(*Texture_VK)
	textureSize := cubeTexture.width
	r.vertexBufferOffset = 0

	cmd := r.BeginSingleTimeCommands()
	barriers := []vk.ImageMemoryBarrier{
		{
			SType:               vk.StructureTypeImageMemoryBarrier,
			OldLayout:           vk.ImageLayoutUndefined,
			NewLayout:           vk.ImageLayoutColorAttachmentOptimal,
			SrcAccessMask:       vk.AccessFlags(vk.AccessNone),
			DstAccessMask:       vk.AccessFlags(vk.AccessColorAttachmentWriteBit),
			SrcQueueFamilyIndex: vk.QueueFamilyIgnored,
			DstQueueFamilyIndex: vk.QueueFamilyIgnored,
			Image:               cubeTexture.img,
			SubresourceRange: vk.ImageSubresourceRange{
				AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
				BaseMipLevel:   0,
				LevelCount:     1,
				BaseArrayLayer: 0,
				LayerCount:     6,
			},
		},
	}
	vk.CmdPipelineBarrier(cmd, vk.PipelineStageFlags(vk.PipelineStageTopOfPipeBit), vk.PipelineStageFlags(vk.PipelineStageColorAttachmentOutputBit), 0, 0, nil, 0, nil, uint32(len(barriers)), barriers)

	colorAttachmentInfos := []vk.RenderingAttachmentInfo{{
		SType:       vk.StructureTypeRenderingAttachmentInfo,
		ImageView:   cubeTexture.imageView,
		ImageLayout: vk.ImageLayoutAttachmentOptimal,
		LoadOp:      vk.AttachmentLoadOpClear,
		StoreOp:     vk.AttachmentStoreOpStore,
		ClearValue:  vk.NewClearValue([]float32{0.0, 0.0, 0.0, 1}),
	}}
	renderInfo := vk.RenderingInfo{
		SType: vk.StructureTypeRenderingInfo,
		RenderArea: vk.Rect2D{
			Offset: vk.Offset2D{
				X: 0, Y: 0,
			},
			Extent: vk.Extent2D{
				Width:  uint32(textureSize),
				Height: uint32(textureSize),
			},
		},
		LayerCount:           6,
		ViewMask:             0b00111111,
		ColorAttachmentCount: 1,
		PColorAttachments:    colorAttachmentInfos,
	}
	vk.CmdBeginRendering(cmd, &renderInfo)
	vk.CmdBindPipeline(cmd, vk.PipelineBindPointGraphics, r.panoramaToCubeMapProgram.pipelines[0])

	vk.CmdBindVertexBuffers(cmd, 0, 1, []vk.Buffer{r.vertexBuffers[0].buffer}, []vk.DeviceSize{0})
	r.SetVertexData2(cmd, -1, -1, 1, -1, -1, 1, 1, 1)

	imageInfo := [][]vk.DescriptorImageInfo{
		{{
			ImageLayout: vk.ImageLayoutShaderReadOnlyOptimal,
			ImageView:   envTexture.imageView,
			Sampler:     envTexture.sampler,
		}},
	}
	descriptorWrites := []vk.WriteDescriptorSet{
		{
			SType:           vk.StructureTypeWriteDescriptorSet,
			DstSet:          nil,
			DstBinding:      0,
			DstArrayElement: 0,
			DescriptorCount: 1,
			DescriptorType:  vk.DescriptorTypeCombinedImageSampler,
			PImageInfo:      imageInfo[0],
		},
	}
	vk.CmdPushDescriptorSet(cmd, vk.PipelineBindPointGraphics, r.panoramaToCubeMapProgram.pipelineLayout, 0, uint32(len(descriptorWrites)), descriptorWrites)

	viewports := []vk.Viewport{{
		MinDepth: 0.0,
		MaxDepth: 1.0,
		X:        0.0,
		Y:        0.0,
		Width:    float32(cubeTexture.width),
		Height:   float32(cubeTexture.width),
	}}
	scissor := vk.Rect2D{
		Offset: vk.Offset2D{
			X: 0,
			Y: 0,
		},
		Extent: vk.Extent2D{
			Width:  uint32(textureSize),
			Height: uint32(textureSize),
		},
	}
	vk.CmdSetViewport(cmd, 0, 1, viewports)
	scissors := []vk.Rect2D{scissor}
	vk.CmdSetScissor(cmd, 0, 1, scissors)

	vk.CmdDraw(cmd, 4, 1, uint32(r.vertexBufferOffset)/8-4, 0)

	vk.CmdEndRendering(cmd)

	//Generate mip maps
	if cubeTexture.mipLevels > 1 {
		barriers := []vk.ImageMemoryBarrier{
			{
				SType:               vk.StructureTypeImageMemoryBarrier,
				OldLayout:           vk.ImageLayoutColorAttachmentOptimal,
				NewLayout:           vk.ImageLayoutTransferSrcOptimal,
				SrcAccessMask:       vk.AccessFlags(vk.AccessColorAttachmentWriteBit),
				DstAccessMask:       vk.AccessFlags(vk.AccessTransferReadBit),
				SrcQueueFamilyIndex: vk.QueueFamilyIgnored,
				DstQueueFamilyIndex: vk.QueueFamilyIgnored,
				Image:               cubeTexture.img,
				SubresourceRange: vk.ImageSubresourceRange{
					AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
					BaseMipLevel:   0,
					LevelCount:     1,
					BaseArrayLayer: 0,
					LayerCount:     6,
				},
			},
			{
				SType:               vk.StructureTypeImageMemoryBarrier,
				OldLayout:           vk.ImageLayoutUndefined,
				NewLayout:           vk.ImageLayoutTransferDstOptimal,
				SrcAccessMask:       0,
				DstAccessMask:       vk.AccessFlags(vk.AccessTransferWriteBit),
				SrcQueueFamilyIndex: vk.QueueFamilyIgnored,
				DstQueueFamilyIndex: vk.QueueFamilyIgnored,
				Image:               cubeTexture.img,
				SubresourceRange: vk.ImageSubresourceRange{
					AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
					BaseMipLevel:   1,
					LevelCount:     cubeTexture.mipLevels - 1,
					BaseArrayLayer: 0,
					LayerCount:     6,
				},
			},
		}

		vk.CmdPipelineBarrier(cmd, vk.PipelineStageFlags(vk.PipelineStageColorAttachmentOutputBit), vk.PipelineStageFlags(vk.PipelineStageTransferBit), 0, 0, nil, 0, nil, uint32(len(barriers)), barriers)

		for i := uint32(1); i < cubeTexture.mipLevels; i++ {
			imageBlits := []vk.ImageBlit{{
				SrcSubresource: vk.ImageSubresourceLayers{
					AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
					MipLevel:       i - 1,
					BaseArrayLayer: 0,
					LayerCount:     6,
				},
				SrcOffsets: [2]vk.Offset3D{
					{
						X: 0,
						Y: 0,
						Z: 0,
					},
					{
						X: (cubeTexture.width >> (i - 1)),
						Y: (cubeTexture.height >> (i - 1)),
						Z: 1,
					},
				},
				DstSubresource: vk.ImageSubresourceLayers{
					AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
					MipLevel:       i,
					BaseArrayLayer: 0,
					LayerCount:     6,
				},
				DstOffsets: [2]vk.Offset3D{
					{
						X: 0,
						Y: 0,
						Z: 0,
					},
					{
						X: (cubeTexture.width >> i),
						Y: (cubeTexture.height >> i),
						Z: 1,
					},
				},
			}}
			vk.CmdBlitImage(cmd, cubeTexture.img, vk.ImageLayoutTransferSrcOptimal, cubeTexture.img, vk.ImageLayoutTransferDstOptimal, uint32(len(imageBlits)), imageBlits, vk.FilterLinear)
			barriers = []vk.ImageMemoryBarrier{
				{
					SType:               vk.StructureTypeImageMemoryBarrier,
					OldLayout:           vk.ImageLayoutTransferDstOptimal,
					NewLayout:           vk.ImageLayoutTransferSrcOptimal,
					SrcAccessMask:       vk.AccessFlags(vk.AccessTransferWriteBit),
					DstAccessMask:       vk.AccessFlags(vk.AccessTransferReadBit),
					SrcQueueFamilyIndex: vk.QueueFamilyIgnored,
					DstQueueFamilyIndex: vk.QueueFamilyIgnored,
					Image:               cubeTexture.img,
					SubresourceRange: vk.ImageSubresourceRange{
						AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
						BaseMipLevel:   i,
						LevelCount:     1,
						BaseArrayLayer: 0,
						LayerCount:     6,
					},
				},
			}
			vk.CmdPipelineBarrier(cmd, vk.PipelineStageFlags(vk.PipelineStageTransferBit), vk.PipelineStageFlags(vk.PipelineStageTransferBit), 0, 0, nil, 0, nil, uint32(len(barriers)), barriers)
		}

		barriers = []vk.ImageMemoryBarrier{
			{
				SType:               vk.StructureTypeImageMemoryBarrier,
				OldLayout:           vk.ImageLayoutTransferSrcOptimal,
				NewLayout:           vk.ImageLayoutShaderReadOnlyOptimal,
				SrcAccessMask:       vk.AccessFlags(vk.AccessTransferReadBit),
				DstAccessMask:       vk.AccessFlags(vk.AccessShaderReadBit),
				SrcQueueFamilyIndex: vk.QueueFamilyIgnored,
				DstQueueFamilyIndex: vk.QueueFamilyIgnored,
				Image:               cubeTexture.img,
				SubresourceRange: vk.ImageSubresourceRange{
					AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
					BaseMipLevel:   0,
					LevelCount:     cubeTexture.mipLevels,
					BaseArrayLayer: 0,
					LayerCount:     6,
				},
			},
		}
		vk.CmdPipelineBarrier(cmd, vk.PipelineStageFlags(vk.PipelineStageTransferBit), vk.PipelineStageFlags(vk.PipelineStageFragmentShaderBit), 0, 0, nil, 0, nil, uint32(len(barriers)), barriers)
	} else {
		barriers := []vk.ImageMemoryBarrier{
			{
				SType:               vk.StructureTypeImageMemoryBarrier,
				OldLayout:           vk.ImageLayoutColorAttachmentOptimal,
				NewLayout:           vk.ImageLayoutShaderReadOnlyOptimal,
				SrcAccessMask:       vk.AccessFlags(vk.AccessTransferWriteBit),
				DstAccessMask:       vk.AccessFlags(vk.AccessShaderReadBit),
				SrcQueueFamilyIndex: vk.QueueFamilyIgnored,
				DstQueueFamilyIndex: vk.QueueFamilyIgnored,
				Image:               cubeTexture.img,
				SubresourceRange: vk.ImageSubresourceRange{
					AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
					BaseMipLevel:   0,
					LevelCount:     1,
					BaseArrayLayer: 0,
					LayerCount:     6,
				},
			},
		}
		vk.CmdPipelineBarrier(cmd, vk.PipelineStageFlags(vk.PipelineStageTransferBit), vk.PipelineStageFlags(vk.PipelineStageFragmentShaderBit), 0, 0, nil, 0, nil, uint32(len(barriers)), barriers)
	}
	vk.EndCommandBuffer(cmd)
	commandBuffers := []vk.CommandBuffer{cmd}
	submitInfo := []vk.SubmitInfo{{
		SType:              vk.StructureTypeSubmitInfo,
		CommandBufferCount: 1,
		PCommandBuffers:    commandBuffers,
	}}
	vk.QueueSubmit(r.queue, 1, submitInfo, nil)
	vk.QueueWaitIdle(r.queue)
}

func (r *Renderer_VK) RenderFilteredCubeMap(distribution int32, cubeTex Texture, filteredTex Texture, mipmapLevel, sampleCount int32, roughness float32) {
	cubeTexture := cubeTex.(*Texture_VK)
	filteredTexture := filteredTex.(*Texture_VK)
	textureSize := filteredTexture.width
	currentTextureSize := textureSize >> mipmapLevel
	format := filteredTexture.MapInternalFormat(Max(filteredTexture.depth, 8))

	imageView := r.CreateImageView(filteredTexture.img, format, uint32(mipmapLevel), 1, 6, true)

	defer vk.DestroyImageView(r.device, imageView, nil)

	cmd := r.BeginSingleTimeCommands()
	barriers := []vk.ImageMemoryBarrier{
		{
			SType:               vk.StructureTypeImageMemoryBarrier,
			OldLayout:           vk.ImageLayoutUndefined,
			NewLayout:           vk.ImageLayoutColorAttachmentOptimal,
			SrcAccessMask:       vk.AccessFlags(vk.AccessNone),
			DstAccessMask:       vk.AccessFlags(vk.AccessColorAttachmentWriteBit),
			SrcQueueFamilyIndex: vk.QueueFamilyIgnored,
			DstQueueFamilyIndex: vk.QueueFamilyIgnored,
			Image:               filteredTexture.img,
			SubresourceRange: vk.ImageSubresourceRange{
				AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
				BaseMipLevel:   uint32(mipmapLevel),
				LevelCount:     1,
				BaseArrayLayer: 0,
				LayerCount:     6,
			},
		},
	}
	vk.CmdPipelineBarrier(cmd, vk.PipelineStageFlags(vk.PipelineStageTopOfPipeBit), vk.PipelineStageFlags(vk.PipelineStageColorAttachmentOutputBit), 0, 0, nil, 0, nil, uint32(len(barriers)), barriers)

	colorAttachmentInfos := []vk.RenderingAttachmentInfo{{
		SType:       vk.StructureTypeRenderingAttachmentInfo,
		ImageView:   imageView,
		ImageLayout: vk.ImageLayoutAttachmentOptimal,
		LoadOp:      vk.AttachmentLoadOpClear,
		StoreOp:     vk.AttachmentStoreOpStore,
		ClearValue:  vk.NewClearValue([]float32{0.0, 0.0, 0.0, 1}),
	}}
	renderInfo := vk.RenderingInfo{
		SType: vk.StructureTypeRenderingInfo,
		RenderArea: vk.Rect2D{
			Offset: vk.Offset2D{
				X: 0, Y: 0,
			},
			Extent: vk.Extent2D{
				Width:  uint32(currentTextureSize),
				Height: uint32(currentTextureSize),
			},
		},
		LayerCount:           6,
		ViewMask:             0b00111111,
		ColorAttachmentCount: 1,
		PColorAttachments:    colorAttachmentInfos,
	}
	vk.CmdBeginRendering(cmd, &renderInfo)
	vk.CmdBindPipeline(cmd, vk.PipelineBindPointGraphics, r.cubemapFilteringProgram.pipelines[0])

	vk.CmdBindVertexBuffers(cmd, 0, 1, []vk.Buffer{r.vertexBuffers[0].buffer}, []vk.DeviceSize{0})
	r.SetVertexData2(cmd, -1, -1, 1, -1, -1, 1, 1, 1)
	imageInfo := [][]vk.DescriptorImageInfo{
		{{
			ImageLayout: vk.ImageLayoutShaderReadOnlyOptimal,
			ImageView:   cubeTexture.imageView,
			Sampler:     cubeTexture.sampler,
		}},
	}
	descriptorWrites := []vk.WriteDescriptorSet{
		{
			SType:           vk.StructureTypeWriteDescriptorSet,
			DstSet:          nil,
			DstBinding:      0,
			DstArrayElement: 0,
			DescriptorCount: 1,
			DescriptorType:  vk.DescriptorTypeCombinedImageSampler,
			PImageInfo:      imageInfo[0],
		},
	}
	vk.CmdPushDescriptorSet(cmd, vk.PipelineBindPointGraphics, r.cubemapFilteringProgram.pipelineLayout, 0, uint32(len(descriptorWrites)), descriptorWrites)
	pushConstants := struct {
		sampleCount    int32
		distribution   int32
		width          int32
		roughness      float32
		intensityScale float32
		isLUT          bool
		_padding       [3]bool
	}{sampleCount, distribution, textureSize, roughness, 1, false, [3]bool{}}

	vk.CmdPushConstants(cmd, r.cubemapFilteringProgram.pipelineLayout, vk.ShaderStageFlags(vk.ShaderStageFragmentBit), 0, 4*6, unsafe.Pointer(&pushConstants))
	viewports := []vk.Viewport{{
		MinDepth: 0.0,
		MaxDepth: 1.0,
		X:        0.0,
		Y:        0.0,
		Width:    float32(currentTextureSize),
		Height:   float32(currentTextureSize),
	}}
	scissor := vk.Rect2D{
		Offset: vk.Offset2D{
			X: 0,
			Y: 0,
		},
		Extent: vk.Extent2D{
			Width:  uint32(currentTextureSize),
			Height: uint32(currentTextureSize),
		},
	}
	vk.CmdSetViewport(cmd, 0, 1, viewports)
	scissors := []vk.Rect2D{scissor}
	vk.CmdSetScissor(cmd, 0, 1, scissors)

	vk.CmdDraw(cmd, 4, 1, uint32(r.vertexBufferOffset)/8-4, 0)

	vk.CmdEndRendering(cmd)

	barriers = []vk.ImageMemoryBarrier{
		{
			SType:               vk.StructureTypeImageMemoryBarrier,
			OldLayout:           vk.ImageLayoutColorAttachmentOptimal,
			NewLayout:           vk.ImageLayoutShaderReadOnlyOptimal,
			SrcAccessMask:       vk.AccessFlags(vk.AccessColorAttachmentWriteBit),
			DstAccessMask:       vk.AccessFlags(vk.AccessShaderReadBit),
			SrcQueueFamilyIndex: vk.QueueFamilyIgnored,
			DstQueueFamilyIndex: vk.QueueFamilyIgnored,
			Image:               filteredTexture.img,
			SubresourceRange: vk.ImageSubresourceRange{
				AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
				BaseMipLevel:   uint32(mipmapLevel),
				LevelCount:     1,
				BaseArrayLayer: 0,
				LayerCount:     6,
			},
		},
	}
	vk.CmdPipelineBarrier(cmd, vk.PipelineStageFlags(vk.PipelineStageColorAttachmentOutputBit), vk.PipelineStageFlags(vk.PipelineStageFragmentShaderBit), 0, 0, nil, 0, nil, uint32(len(barriers)), barriers)
	r.EndSingleTimeCommands(cmd)
}
func (r *Renderer_VK) RenderLUT(distribution int32, cubeTex Texture, lutTex Texture, sampleCount int32) {
	cubeTexture := cubeTex.(*Texture_VK)
	lutTexture := lutTex.(*Texture_VK)
	textureSize := lutTexture.width

	cmd := r.BeginSingleTimeCommands()
	barriers := []vk.ImageMemoryBarrier{
		{
			SType:               vk.StructureTypeImageMemoryBarrier,
			OldLayout:           vk.ImageLayoutUndefined,
			NewLayout:           vk.ImageLayoutColorAttachmentOptimal,
			SrcAccessMask:       vk.AccessFlags(vk.AccessNone),
			DstAccessMask:       vk.AccessFlags(vk.AccessColorAttachmentWriteBit),
			SrcQueueFamilyIndex: vk.QueueFamilyIgnored,
			DstQueueFamilyIndex: vk.QueueFamilyIgnored,
			Image:               lutTexture.img,
			SubresourceRange: vk.ImageSubresourceRange{
				AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
				BaseMipLevel:   0,
				LevelCount:     1,
				BaseArrayLayer: 0,
				LayerCount:     1,
			},
		},
	}
	vk.CmdPipelineBarrier(cmd, vk.PipelineStageFlags(vk.PipelineStageTopOfPipeBit), vk.PipelineStageFlags(vk.PipelineStageColorAttachmentOutputBit), 0, 0, nil, 0, nil, uint32(len(barriers)), barriers)

	colorAttachmentInfos := []vk.RenderingAttachmentInfo{{
		SType:       vk.StructureTypeRenderingAttachmentInfo,
		ImageView:   lutTexture.imageView,
		ImageLayout: vk.ImageLayoutAttachmentOptimal,
		LoadOp:      vk.AttachmentLoadOpClear,
		StoreOp:     vk.AttachmentStoreOpStore,
		ClearValue:  vk.NewClearValue([]float32{0.0, 0.0, 0.0, 1}),
	}}
	renderInfo := vk.RenderingInfo{
		SType: vk.StructureTypeRenderingInfo,
		RenderArea: vk.Rect2D{
			Offset: vk.Offset2D{
				X: 0, Y: 0,
			},
			Extent: vk.Extent2D{
				Width:  uint32(textureSize),
				Height: uint32(textureSize),
			},
		},
		LayerCount:           1,
		ViewMask:             0b00000001,
		ColorAttachmentCount: 1,
		PColorAttachments:    colorAttachmentInfos,
	}
	vk.CmdBeginRendering(cmd, &renderInfo)
	vk.CmdBindPipeline(cmd, vk.PipelineBindPointGraphics, r.lutProgram.pipelines[0])

	vk.CmdBindVertexBuffers(cmd, 0, 1, []vk.Buffer{r.vertexBuffers[0].buffer}, []vk.DeviceSize{0})
	r.SetVertexData2(cmd, -1, -1, 1, -1, -1, 1, 1, 1)
	imageInfo := [][]vk.DescriptorImageInfo{
		{{
			ImageLayout: vk.ImageLayoutShaderReadOnlyOptimal,
			ImageView:   cubeTexture.imageView,
			Sampler:     cubeTexture.sampler,
		}},
	}
	descriptorWrites := []vk.WriteDescriptorSet{
		{
			SType:           vk.StructureTypeWriteDescriptorSet,
			DstSet:          nil,
			DstBinding:      0,
			DstArrayElement: 0,
			DescriptorCount: 1,
			DescriptorType:  vk.DescriptorTypeCombinedImageSampler,
			PImageInfo:      imageInfo[0],
		},
	}
	vk.CmdPushDescriptorSet(cmd, vk.PipelineBindPointGraphics, r.lutProgram.pipelineLayout, 0, uint32(len(descriptorWrites)), descriptorWrites)
	pushConstants := struct {
		sampleCount    int32
		distribution   int32
		width          int32
		roughness      float32
		intensityScale float32
		isLUT          bool
		_padding       [3]bool
	}{sampleCount, distribution, textureSize, 0, 1, true, [3]bool{}}
	vk.CmdPushConstants(cmd, r.lutProgram.pipelineLayout, vk.ShaderStageFlags(vk.ShaderStageFragmentBit), 0, 4*6, unsafe.Pointer(&pushConstants))
	viewports := []vk.Viewport{{
		MinDepth: 0.0,
		MaxDepth: 1.0,
		X:        0.0,
		Y:        0.0,
		Width:    float32(textureSize),
		Height:   float32(textureSize),
	}}
	scissor := vk.Rect2D{
		Offset: vk.Offset2D{
			X: 0,
			Y: 0,
		},
		Extent: vk.Extent2D{
			Width:  uint32(textureSize),
			Height: uint32(textureSize),
		},
	}
	vk.CmdSetViewport(cmd, 0, 1, viewports)
	scissors := []vk.Rect2D{scissor}
	vk.CmdSetScissor(cmd, 0, 1, scissors)

	vk.CmdDraw(cmd, 4, 1, uint32(r.vertexBufferOffset)/8-4, 0)

	vk.CmdEndRendering(cmd)

	barriers = []vk.ImageMemoryBarrier{
		{
			SType:               vk.StructureTypeImageMemoryBarrier,
			OldLayout:           vk.ImageLayoutColorAttachmentOptimal,
			NewLayout:           vk.ImageLayoutShaderReadOnlyOptimal,
			SrcAccessMask:       vk.AccessFlags(vk.AccessColorAttachmentWriteBit),
			DstAccessMask:       vk.AccessFlags(vk.AccessShaderReadBit),
			SrcQueueFamilyIndex: vk.QueueFamilyIgnored,
			DstQueueFamilyIndex: vk.QueueFamilyIgnored,
			Image:               lutTexture.img,
			SubresourceRange: vk.ImageSubresourceRange{
				AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
				BaseMipLevel:   0,
				LevelCount:     1,
				BaseArrayLayer: 0,
				LayerCount:     1,
			},
		},
	}
	vk.CmdPipelineBarrier(cmd, vk.PipelineStageFlags(vk.PipelineStageColorAttachmentOutputBit), vk.PipelineStageFlags(vk.PipelineStageFragmentShaderBit), 0, 0, nil, 0, nil, uint32(len(barriers)), barriers)

	r.EndSingleTimeCommands(cmd)
}

func (r *Renderer_VK) FlushTempCommands() {
	if len(r.stagingImageBarriers[0]) == 0 && len(r.tempCommands) == 0 {
		return
	}
	commandBuffers := r.tempCommands
	if len(r.stagingImageBarriers[0]) > 0 {
		commandBuffer := r.BeginSingleTimeCommands()
		vk.CmdPipelineBarrier(commandBuffer, vk.PipelineStageFlags(vk.PipelineStageTopOfPipeBit), vk.PipelineStageFlags(vk.PipelineStageTransferBit), 0, 0, nil, 0, nil, uint32(len(r.stagingImageBarriers[0])), r.stagingImageBarriers[0])
		for img := range r.stagingImageCopyRegions {
			regions := r.stagingImageCopyRegions[img]
			vk.CmdCopyBufferToImage(commandBuffer, r.stagingBuffers[r.stagingBufferIndex].buffer, img, vk.ImageLayoutTransferDstOptimal, uint32(len(regions)), regions)
		}
		vk.CmdPipelineBarrier(commandBuffer, vk.PipelineStageFlags(vk.PipelineStageTransferBit), vk.PipelineStageFlags(vk.PipelineStageFragmentShaderBit), 0, 0, nil, 0, nil, uint32(len(r.stagingImageBarriers[1])), r.stagingImageBarriers[1])
		vk.EndCommandBuffer(commandBuffer)
		commandBuffers = append([]vk.CommandBuffer{commandBuffer}, commandBuffers...)
		r.stagingBufferOffset = 0
		r.stagingImageBarriers[0] = r.stagingImageBarriers[0][:0]
		r.stagingImageBarriers[1] = r.stagingImageBarriers[1][:0]
		r.stagingImageCopyRegions = make(map[vk.Image][]vk.BufferImageCopy)
		if r.stagingBufferFences[r.stagingBufferIndex] {
			vk.WaitForFences(r.device, 1, r.fences[1+r.stagingBufferIndex:2+r.stagingBufferIndex], vk.True, 10*1000*1000*1000)
			vk.ResetFences(r.device, 1, r.fences[1+r.stagingBufferIndex:2+r.stagingBufferIndex])
		}
		submitInfo := []vk.SubmitInfo{{
			SType:              vk.StructureTypeSubmitInfo,
			CommandBufferCount: uint32(len(commandBuffers)),
			PCommandBuffers:    commandBuffers,
		}}
		vk.QueueSubmit(r.queue, 1, submitInfo, r.fences[1+r.stagingBufferIndex])
		r.stagingBufferFences[r.stagingBufferIndex] = true
		if r.stagingBufferIndex == 0 {
			r.stagingBufferIndex = 1
		} else {
			r.stagingBufferIndex = 0
		}
	} else {
		submitInfo := []vk.SubmitInfo{{
			SType:              vk.StructureTypeSubmitInfo,
			CommandBufferCount: uint32(len(commandBuffers)),
			PCommandBuffers:    commandBuffers,
		}}
		vk.QueueSubmit(r.queue, 1, submitInfo, nil)
	}
	r.usedCommands = append(r.usedCommands, commandBuffers...)
	r.tempCommands = r.tempCommands[:0]
	return
}
func (r *Renderer_VK) CopyToStagingBuffer(size uint32, src []byte) vk.DeviceSize {
	if size > uint32(r.stagingBuffers[r.stagingBufferIndex].size) {
		r.FlushTempCommands()
		if size > uint32(r.stagingBuffers[r.stagingBufferIndex].size) {
			r.ResizeStagingBuffer(size)
		}
	}
	if size+r.stagingBufferOffset > uint32(r.stagingBuffers[r.stagingBufferIndex].size) {
		r.FlushTempCommands()
	}
	if r.stagingBufferFences[r.stagingBufferIndex] {
		vk.WaitForFences(r.device, 1, r.fences[1+r.stagingBufferIndex:2+r.stagingBufferIndex], vk.True, 10*1000*1000*1000)
		vk.ResetFences(r.device, 1, r.fences[1+r.stagingBufferIndex:2+r.stagingBufferIndex])
		r.stagingBufferFences[r.stagingBufferIndex] = false
	}
	n := vk.Memcopy(unsafe.Add(r.stagingBuffers[r.stagingBufferIndex].data, r.stagingBufferOffset), src)
	if n != int(size) {
		log.Println("[WARN] failed to copy image data")
	}
	ret := vk.DeviceSize(r.stagingBufferOffset)
	if size%16 > 0 {
		r.stagingBufferOffset += size + 16 - (size % 16)
	} else {
		r.stagingBufferOffset += size
	}
	return ret
}
func (r *Renderer_VK) ResizeStagingBuffer(size uint32) {
	r.stagingBufferFences[r.stagingBufferIndex] = false
	r.destroyResourceQueues[r.destroyResourceQueueIndex] <- VulkanResource{
		VulkanResourceTypeFence,
		[]interface{}{
			r.fences[1+r.stagingBufferIndex],
		},
	}
	var fence vk.Fence
	fenceCreateInfo := vk.FenceCreateInfo{
		SType: vk.StructureTypeFenceCreateInfo,
	}
	err := vk.Error(vk.CreateFence(r.device, &fenceCreateInfo, nil, &fence))
	if err != nil {
		panic(err)
	}
	r.fences[1+r.stagingBufferIndex] = fence
	r.destroyResourceQueues[r.destroyResourceQueueIndex] <- VulkanResource{
		VulkanResourceTypeBuffer,
		[]interface{}{
			r.stagingBuffers[r.stagingBufferIndex].buffer, r.stagingBuffers[r.stagingBufferIndex].bufferMemory,
		},
	}
	vk.UnmapMemory(r.device, r.stagingBuffers[r.stagingBufferIndex].bufferMemory)
	var stagingBufferMemory vk.DeviceMemory
	var data unsafe.Pointer
	stagingBuffer, err := r.CreateBuffer(vk.DeviceSize(size), vk.BufferUsageFlags(vk.BufferUsageTransferSrcBit), (vk.MemoryPropertyHostVisibleBit | vk.MemoryPropertyHostCoherentBit), &stagingBufferMemory)
	if err != nil {
		panic(err)
	}
	vk.MapMemory(r.device, stagingBufferMemory, 0, vk.DeviceSize(size), 0, &data)
	r.stagingBuffers[r.stagingBufferIndex].buffer = stagingBuffer
	r.stagingBuffers[r.stagingBufferIndex].bufferMemory = stagingBufferMemory
	r.stagingBuffers[r.stagingBufferIndex].data = data
	r.stagingBuffers[r.stagingBufferIndex].size = uintptr(size)
	r.stagingBufferOffset = 0
}
func (r *Renderer_VK) PerspectiveProjectionMatrix(angle, aspect, near, far float32) mgl.Mat4 {
	return mgl.Perspective(angle, aspect, near/2, far)
}

func (r *Renderer_VK) OrthographicProjectionMatrix(left, right, bottom, top, near, far float32) mgl.Mat4 {
	ret := mgl.Ortho(left, right, bottom, top, near, far)
	ret[10] *= 0.5
	ret[14] = ret[14]*0.5 + 0.5
	return ret
}

func (r *Renderer_VK) NewWorkerThread() bool {
	r.renderShadowMap = true
	r.waitGroup.Add(1)
	return true
}

func (r *Renderer_VK) rgbToRGBA(input []float32) []float32 {
	result := make([]float32, 0, len(input)/3*4)
	for i, val := range input {
		result = append(result, val)
		if (i+1)%3 == 0 {
			result = append(result, 1)
		}
	}
	return result
}

func (r *Renderer_VK) AlignUniformSize(uniformSize uint32) uint32 {
	minAlignment := r.minUniformBufferOffsetAlignment
	if uniformSize < minAlignment {
		uniformSize = minAlignment
	} else if uniformSize > minAlignment && minAlignment > 0 && uniformSize%minAlignment != 0 {
		uniformSize = (uniformSize/minAlignment + 1) * minAlignment
	}
	return uniformSize
}

func (r *Renderer_VK) CreateImage(width, height uint32, format vk.Format, mipLevel uint32, layerCount uint32, usage vk.ImageUsageFlags, numSamples int32, tiling vk.ImageTiling, cube bool) vk.Image {
	imageExtent := vk.Extent3D{
		Width:  width,
		Height: height,
		Depth:  1,
	}
	imageInfo := &vk.ImageCreateInfo{
		SType:         vk.StructureTypeImageCreateInfo,
		ImageType:     vk.ImageType2d,
		Extent:        imageExtent,
		MipLevels:     mipLevel,
		ArrayLayers:   layerCount,
		Format:        format,
		Tiling:        tiling,
		InitialLayout: vk.ImageLayoutUndefined,
		Usage:         usage,
		Samples:       vk.SampleCountFlagBits(numSamples),
		SharingMode:   vk.SharingModeExclusive,
	}
	if cube {
		imageInfo.Flags = vk.ImageCreateFlags(vk.ImageCreateCubeCompatibleBit)
	} else if layerCount > 1 {
		imageInfo.ImageType = vk.ImageType2d
	}
	var img vk.Image
	err := vk.Error(vk.CreateImage(r.device, imageInfo, nil, &img))
	if err != nil {
		panic(err)
	}
	return img
}

func (r *Renderer_VK) CreateImageView(img vk.Image, format vk.Format, baseLevel uint32, mipLevel uint32, layerCount uint32, cube bool) vk.ImageView {
	var imageView vk.ImageView
	viewCreateInfo := vk.ImageViewCreateInfo{
		SType:    vk.StructureTypeImageViewCreateInfo,
		Image:    img,
		ViewType: vk.ImageViewType2d,
		Format:   format,
		Components: vk.ComponentMapping{
			R: vk.ComponentSwizzleR,
			G: vk.ComponentSwizzleG,
			B: vk.ComponentSwizzleB,
			A: vk.ComponentSwizzleA,
		},
		SubresourceRange: vk.ImageSubresourceRange{
			AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
			BaseMipLevel:   baseLevel,
			LevelCount:     mipLevel,
			BaseArrayLayer: 0,
			LayerCount:     layerCount,
		},
	}
	if cube {
		if layerCount == 6 {
			viewCreateInfo.ViewType = vk.ImageViewTypeCube
		} else if layerCount > 6 {
			viewCreateInfo.ViewType = vk.ImageViewTypeCubeArray
		}
	} else if layerCount > 1 {
		viewCreateInfo.ViewType = vk.ImageViewType2dArray
	}
	if format == vk.FormatD32Sfloat {
		viewCreateInfo.SubresourceRange.AspectMask = vk.ImageAspectFlags(vk.ImageAspectDepthBit)
	}
	err := vk.Error(vk.CreateImageView(r.device, &viewCreateInfo, nil, &imageView))
	if err != nil {
		err = fmt.Errorf("vk.CreateImageView failed with %s", err)
		panic(err)
	}
	return imageView
}

func (r *Renderer_VK) AllocateImageMemory(img vk.Image, memoryProperty vk.MemoryPropertyFlagBits) vk.DeviceMemory {
	var memReq vk.MemoryRequirements
	vk.GetImageMemoryRequirements(r.device, img, &memReq)
	//memReq.Deref()
	memoryTypeIndex, ok := r.memoryTypeMap[memoryProperty]
	if !ok {
		memoryTypeIndex, _ = vk.FindMemoryTypeIndex(r.gpuDevices[r.gpuIndex], memReq.MemoryTypeBits, memoryProperty)
		r.memoryTypeMap[memoryProperty] = memoryTypeIndex
	}
	r.allocatedImageMemory += uint64(memReq.Size)
	allocInfo := vk.MemoryAllocateInfo{
		SType:           vk.StructureTypeMemoryAllocateInfo,
		AllocationSize:  memReq.Size,
		MemoryTypeIndex: memoryTypeIndex,
	}
	var imageMemory vk.DeviceMemory
	err := vk.Error(vk.AllocateMemory(r.device, &allocInfo, nil, &imageMemory))
	if err != nil {
		err = fmt.Errorf("vk.AllocateMemory failed with %s", err)
		panic(err)
	}
	err = vk.Error(vk.BindImageMemory(r.device, img, imageMemory, 0))
	if err != nil {
		err = fmt.Errorf("vk.BindImageMemory failed with %s", err)
		panic(err)
	}
	return imageMemory
}

func (r *Renderer_VK) SetVSync(interval int) {
	// We just ignore the interval here and force it on
	r.setVSync = true
}
