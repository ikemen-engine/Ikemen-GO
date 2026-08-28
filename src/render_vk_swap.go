//go:build !kinc && !android

package main

import (
	"fmt"
	"sort"
	"unsafe"

	vk "github.com/Eiton/vulkan"
)

// SwapOutTextures evicts cold DEVICE_LOCAL textures to HOST_VISIBLE memory
// to free up device-local VRAM. It iterates swappableTextures sorted by
// lastUsedFrame (coldest first), copies each texture's image data from
// DEVICE_LOCAL to a new HOST_VISIBLE image, then destroys the old resources.
// Stops when freedBytes >= budgetBytes or no more eligible textures remain.
//
// The operation is structured in three passes:
//  1. Setup pass: create new HOST_VISIBLE images and allocate memory (CPU-side).
//  2. GPU pass: batch ALL copy operations into a single command buffer submission.
//  3. Cleanup pass: create image views, destroy old resources, update texture fields.
//
// Returns nil on success (even if no textures were swapped). Returns an error
// only on catastrophic failure that prevents further operation.
func (r *Renderer_VK) SwapOutTextures(budgetBytes vk.DeviceSize) error {
	// Collect swappable textures into a slice for sorting.
	textures := make([]*Texture_VK, 0, len(r.swappableTextures))
	for t := range r.swappableTextures {
		// Defense-in-depth: skip textures marked non-swappable even if
		// they somehow remain in the map (e.g. marked after registration).
		if t.nonSwappable {
			continue
		}
		if _, ok := r.stagingImageCopyRegions[t.img]; !ok {
			textures = append(textures, t)
		}
	}

	// Sort by lastUsedFrame ascending (coldest first).
	sort.Slice(textures, func(i, j int) bool {
		return textures[i].lastUsedFrame < textures[j].lastUsedFrame
	})

	// Struct to hold per-texture swap state across passes.
	type swapEntry struct {
		texture    *Texture_VK
		newImg     vk.Image
		newAlloc   *VulkanAllocation
		format     vk.Format
		layerCount uint32
		oldSize    vk.DeviceSize
	}
	var swaps []swapEntry
	var estimatedFreed vk.DeviceSize

	// PASS 1: Setup — create new HOST_VISIBLE images and allocate memory.
	// Stop adding to swaps once estimated freed bytes reaches the budget.
	for _, t := range textures {
		if estimatedFreed >= budgetBytes {
			break
		}

		// Only swap out DEVICE_LOCAL textures.
		if t.residency != ResidentDeviceLocal {
			continue
		}

		// Do not swap out textures used in the current frame.
		if t.lastUsedFrame == r.currentFrameNumber {
			continue
		}

		// Safety: skip textures with nil resources.
		if t.img == nil || t.allocation == nil {
			continue
		}

		format := t.MapInternalFormat(Max(t.depth, 8))
		layerCount := uint32(1)
		cube := false
		usage := vk.ImageUsageFlags(vk.ImageUsageTransferSrcBit | vk.ImageUsageTransferDstBit | vk.ImageUsageSampledBit)

		// Create new HOST_VISIBLE image with LINEAR tiling.
		newImg := r.CreateImage(
			uint32(t.width), uint32(t.height), format,
			t.mipLevels, layerCount, usage,
			1, vk.ImageTilingLinear, cube,
		)

		// Allocate HOST_VISIBLE memory — no fallback, this IS the fallback tier.
		newAlloc, err := r.allocator.AllocateImageMemoryWithFallback(
			newImg,
			vk.MemoryPropertyHostVisibleBit|vk.MemoryPropertyHostCoherentBit,
			0, // no further fallback
		)
		if err != nil {
			vk.DestroyImage(r.device, newImg, nil)
			if vkDebug {
				LogMessage("[VRAM] swap-out: HOST_VISIBLE allocation failed for %dx%d texture: %v", t.width, t.height, err)
			}
			continue
		}

		swaps = append(swaps, swapEntry{
			texture:    t,
			newImg:     newImg,
			newAlloc:   newAlloc,
			format:     format,
			layerCount: layerCount,
			oldSize:    t.allocation.size,
		})
		estimatedFreed += t.allocation.size
	}

	// If no textures to swap, return early — no GPU work needed.
	if len(swaps) == 0 {
		return nil
	}

	// PASS 2: GPU — batch all copy operations into a single command buffer.
	commandBuffer := r.BeginSingleTimeCommands()

	// Collect all pre-copy barriers into one batch: transition all old images to
	// TRANSFER_SRC_OPTIMAL and all new images to TRANSFER_DST_OPTIMAL in one call.
	var preCopyBarriers []vk.ImageMemoryBarrier
	for i, s := range swaps {
		if vkDebug {
			LogMessage("[VRAM] swap-out texture %d 0x%x %x", i,
				uintptr(unsafe.Pointer(s.texture.img)), uintptr(unsafe.Pointer(s.newImg)))
		}
		// Transition old image: SHADER_READ_ONLY_OPTIMAL → TRANSFER_SRC_OPTIMAL.
		preCopyBarriers = append(preCopyBarriers, vk.ImageMemoryBarrier{
			SType:               vk.StructureTypeImageMemoryBarrier,
			OldLayout:           vk.ImageLayoutShaderReadOnlyOptimal,
			NewLayout:           vk.ImageLayoutTransferSrcOptimal,
			SrcAccessMask:       vk.AccessFlags(vk.AccessShaderReadBit),
			DstAccessMask:       vk.AccessFlags(vk.AccessTransferReadBit),
			SrcQueueFamilyIndex: vk.QueueFamilyIgnored,
			DstQueueFamilyIndex: vk.QueueFamilyIgnored,
			Image:               s.texture.img,
			SubresourceRange: vk.ImageSubresourceRange{
				AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
				BaseMipLevel:   0,
				LevelCount:     s.texture.mipLevels,
				BaseArrayLayer: 0,
				LayerCount:     s.layerCount,
			},
		})

		// Transition new image: UNDEFINED → TRANSFER_DST_OPTIMAL.
		preCopyBarriers = append(preCopyBarriers, vk.ImageMemoryBarrier{
			SType:               vk.StructureTypeImageMemoryBarrier,
			OldLayout:           vk.ImageLayoutUndefined,
			NewLayout:           vk.ImageLayoutTransferDstOptimal,
			SrcAccessMask:       vk.AccessFlags(vk.AccessNone),
			DstAccessMask:       vk.AccessFlags(vk.AccessTransferWriteBit),
			SrcQueueFamilyIndex: vk.QueueFamilyIgnored,
			DstQueueFamilyIndex: vk.QueueFamilyIgnored,
			Image:               s.newImg,
			SubresourceRange: vk.ImageSubresourceRange{
				AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
				BaseMipLevel:   0,
				LevelCount:     s.texture.mipLevels,
				BaseArrayLayer: 0,
				LayerCount:     s.layerCount,
			},
		})
	}

	vk.CmdPipelineBarrier(
		commandBuffer,
		vk.PipelineStageFlags(vk.PipelineStageFragmentShaderBit),
		vk.PipelineStageFlags(vk.PipelineStageTransferBit),
		0, 0, nil, 0, nil,
		uint32(len(preCopyBarriers)), preCopyBarriers,
	)

	// Record copy commands for every texture.
	for _, s := range swaps {
		region := vk.ImageCopy{
			SrcSubresource: vk.ImageSubresourceLayers{
				AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
				MipLevel:       0,
				BaseArrayLayer: 0,
				LayerCount:     s.layerCount,
			},
			DstSubresource: vk.ImageSubresourceLayers{
				AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
				MipLevel:       0,
				BaseArrayLayer: 0,
				LayerCount:     s.layerCount,
			},
			SrcOffset: vk.Offset3D{X: 0, Y: 0, Z: 0},
			DstOffset: vk.Offset3D{X: 0, Y: 0, Z: 0},
			Extent: vk.Extent3D{
				Width:  uint32(s.texture.width),
				Height: uint32(s.texture.height),
				Depth:  1,
			},
		}
		vk.CmdCopyImage(
			commandBuffer,
			s.texture.img, vk.ImageLayoutTransferSrcOptimal,
			s.newImg, vk.ImageLayoutTransferDstOptimal,
			1, []vk.ImageCopy{region},
		)
	}

	// Record post-copy barriers for ALL textures in one batch.
	// This ensures all copies complete before any layout transitions.
	var postCopyBarriers []vk.ImageMemoryBarrier
	for _, s := range swaps {
		// Transition new image: TRANSFER_DST_OPTIMAL → GENERAL.
		// HOST_VISIBLE images use LINEAR tiling, which requires GENERAL layout for shader sampling.
		newShaderBarrier := vk.ImageMemoryBarrier{
			SType:               vk.StructureTypeImageMemoryBarrier,
			OldLayout:           vk.ImageLayoutTransferDstOptimal,
			NewLayout:           vk.ImageLayoutGeneral,
			SrcAccessMask:       vk.AccessFlags(vk.AccessTransferWriteBit),
			DstAccessMask:       vk.AccessFlags(vk.AccessShaderReadBit),
			SrcQueueFamilyIndex: vk.QueueFamilyIgnored,
			DstQueueFamilyIndex: vk.QueueFamilyIgnored,
			Image:               s.newImg,
			SubresourceRange: vk.ImageSubresourceRange{
				AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
				BaseMipLevel:   0,
				LevelCount:     s.texture.mipLevels,
				BaseArrayLayer: 0,
				LayerCount:     s.layerCount,
			},
		}

		// Transition old image back: TRANSFER_SRC_OPTIMAL → SHADER_READ_ONLY_OPTIMAL
		// (in case it is accessed again before destruction).
		oldShaderBarrier := vk.ImageMemoryBarrier{
			SType:               vk.StructureTypeImageMemoryBarrier,
			OldLayout:           vk.ImageLayoutTransferSrcOptimal,
			NewLayout:           vk.ImageLayoutShaderReadOnlyOptimal,
			SrcAccessMask:       vk.AccessFlags(vk.AccessTransferReadBit),
			DstAccessMask:       vk.AccessFlags(vk.AccessShaderReadBit),
			SrcQueueFamilyIndex: vk.QueueFamilyIgnored,
			DstQueueFamilyIndex: vk.QueueFamilyIgnored,
			Image:               s.texture.img,
			SubresourceRange: vk.ImageSubresourceRange{
				AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
				BaseMipLevel:   0,
				LevelCount:     s.texture.mipLevels,
				BaseArrayLayer: 0,
				LayerCount:     s.layerCount,
			},
		}

		postCopyBarriers = append(postCopyBarriers, newShaderBarrier, oldShaderBarrier)
	}

	vk.CmdPipelineBarrier(
		commandBuffer,
		vk.PipelineStageFlags(vk.PipelineStageTransferBit),
		vk.PipelineStageFlags(vk.PipelineStageFragmentShaderBit),
		0, 0, nil, 0, nil,
		uint32(len(postCopyBarriers)), postCopyBarriers,
	)

	r.EndSingleTimeCommands(commandBuffer)

	// PASS 3: Cleanup — create image views, destroy old resources, update fields.
	// All swaps in the batch complete — the budget controls how many enter the batch,
	// not which ones complete.
	var freedBytes vk.DeviceSize
	for _, s := range swaps {
		// Create new image view for the HOST_VISIBLE image.
		newImageView := r.CreateImageView(s.newImg, s.format, 0, s.texture.mipLevels, s.layerCount, false)

		// Destroy old resources. If the texture hasn't been used in the last 2
		// frames, the GPU is done — destroy immediately. Otherwise, defer to
		// avoid destroying objects the in-flight command buffer may still reference.
		r.destroyOrDeferImageResources(s.texture.lastUsedFrame, s.texture.img, s.texture.imageView, s.texture.allocation)

		// Update texture fields.
		s.texture.img = s.newImg
		s.texture.imageView = newImageView
		s.texture.allocation = s.newAlloc
		s.texture.residency = ResidentHostVisible
		s.texture.hostVisibleUploaded = true

		// Remove from swappableTextures — HOST_VISIBLE textures are not eligible for further swap-out.
		delete(r.swappableTextures, s.texture)

		freedBytes += s.oldSize
	}

	return nil
}

// SwapInTexture transitions a texture back to DEVICE_LOCAL memory.
//
// Handles two cases:
//   - ResidentHostVisible: copies data from HOST_VISIBLE image to new DEVICE_LOCAL
//     image via vk.CmdCopyImage, destroys old HOST_VISIBLE resources.
//   - ResidentSwappedOut: creates new DEVICE_LOCAL image, uploads data from
//     hostBackingBuffer via staging buffer (immediate submission), frees hostBackingBuffer.
//
// After a successful transition, the texture is re-registered in swappableTextures
// so it is eligible for future swap-out.
//
// Returns nil if the texture is already ResidentDeviceLocal (no-op).
// Returns an error for ResidentEvicted textures (handled by TouchTexture reload).
// Returns an error if DEVICE_LOCAL allocation fails (texture stays in current state).
func (r *Renderer_VK) SwapInTexture(t *Texture_VK) error {
	switch t.residency {
	case ResidentDeviceLocal:
		// Already in the right state — no-op.
		return nil

	case ResidentEvicted:
		return fmt.Errorf("SwapInTexture: texture is evicted, handled by TouchTexture reload")

	case ResidentHostVisible, ResidentSwappedOut:
		// Delegate to the existing transition helper.
		if err := r.TransitionToDeviceLocal(t); err != nil {
			if vkDebug {
				LogMessage("[VRAM] swap-in to DEVICE_LOCAL failed for %dx%d texture: %v", t.width, t.height, err)
			}
			// Fallback: create a HOST_VISIBLE image so the texture has valid GPU
			// resources and can still be rendered (with degraded performance).
			// Without this, imageView would be nil and descriptor writes would use
			// an invalid/null VkImageView handle.
			if t.residency == ResidentSwappedOut && t.hostBackingBuffer != nil {
				if fallbackErr := r.swapInAsHostVisible(t); fallbackErr != nil {
					if vkDebug {
						LogMessage("[VRAM] swap-in fallback to HOST_VISIBLE also failed for %dx%d texture: %v", t.width, t.height, fallbackErr)
					}
					return fmt.Errorf("SwapInTexture: DEVICE_LOCAL failed (%v), HOST_VISIBLE fallback failed (%v)", err, fallbackErr)
				}
				if !t.nonSwappable {
					r.swappableTextures[t] = true
				}
				if vkDebug {
					LogMessage("[VRAM] swap-in texture %dx%d to HOST_VISIBLE (fallback)", t.width, t.height)
				}
				return nil
			}
			return err
		}

		// Re-register in swappableTextures — the texture is now DEVICE_LOCAL
		// and eligible for future swap-out (unless marked non-swappable).
		if !t.nonSwappable {
			r.swappableTextures[t] = true
		}

		if vkDebug {
			LogMessage("[VRAM] swap-in texture %dx%d to DEVICE_LOCAL", t.width, t.height)
		}
		return nil

	default:
		return fmt.Errorf("SwapInTexture: unknown residency state %d", t.residency)
	}
}

// EvictTexture evicts a texture from GPU memory. The texture's pixel data
// is preserved in hostBackingBuffer and residency is set to ResidentEvicted.
// TouchTexture will recreate the texture from hostBackingBuffer when it is
// next accessed.
//
// Returns an error if the texture is not in swappableTextures (render targets,
// depth buffers, shadow maps, and palette textures are not evictable).
func (r *Renderer_VK) EvictTexture(t *Texture_VK) error {
	if t.nonSwappable || !r.swappableTextures[t] {
		return fmt.Errorf("EvictTexture: texture is not swappable (render target, depth buffer, shadow map, or palette)")
	}

	switch t.residency {
	case ResidentDeviceLocal:
		if err := r.evictDeviceLocal(t); err != nil {
			return err
		}

	case ResidentHostVisible:
		if err := r.evictHostVisible(t); err != nil {
			return err
		}

	case ResidentSwappedOut:
		// Already have hostBackingBuffer — just mark as evicted.
		t.residency = ResidentEvicted
		t.needsReload = true
		if vkDebug {
			LogMessage("[VRAM] evict texture %dx%d (was SwappedOut)", t.width, t.height)
		}

	case ResidentEvicted:
		// Already evicted — no-op.

	default:
		return fmt.Errorf("EvictTexture: unknown residency state %d", t.residency)
	}

	// Remove from swappableTextures — evicted textures are no longer eligible for swap-out.
	delete(r.swappableTextures, t)

	return nil
}

// evictDeviceLocal reads back DEVICE_LOCAL texture data into hostBackingBuffer
// via a staging buffer, then destroys GPU resources.
func (r *Renderer_VK) evictDeviceLocal(t *Texture_VK) error {
	if t.img == nil || t.allocation == nil {
		return fmt.Errorf("evictDeviceLocal: texture has nil GPU resources")
	}

	bytesPerPixel := uint32(t.depth / 8)
	totalSize := int(t.width) * int(t.height) * int(bytesPerPixel)

	// Create a staging buffer for read-back.
	var stagingBufferMemory vk.DeviceMemory
	stagingBuffer, err := r.CreateBuffer(
		vk.DeviceSize(totalSize),
		vk.BufferUsageFlags(vk.BufferUsageTransferDstBit),
		vk.MemoryPropertyHostVisibleBit|vk.MemoryPropertyHostCoherentBit,
		&stagingBufferMemory,
	)
	if err != nil {
		return fmt.Errorf("evictDeviceLocal: CreateBuffer failed: %w", err)
	}

	// Read back image data via vk.CmdCopyImageToBuffer.
	commandBuffer := r.BeginSingleTimeCommands()

	// Transition image: SHADER_READ_ONLY_OPTIMAL → TRANSFER_SRC_OPTIMAL.
	srcBarrier := vk.ImageMemoryBarrier{
		SType:               vk.StructureTypeImageMemoryBarrier,
		OldLayout:           vk.ImageLayoutShaderReadOnlyOptimal,
		NewLayout:           vk.ImageLayoutTransferSrcOptimal,
		SrcAccessMask:       vk.AccessFlags(vk.AccessShaderReadBit),
		DstAccessMask:       vk.AccessFlags(vk.AccessTransferReadBit),
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
	}

	vk.CmdPipelineBarrier(
		commandBuffer,
		vk.PipelineStageFlags(vk.PipelineStageFragmentShaderBit),
		vk.PipelineStageFlags(vk.PipelineStageTransferBit),
		0, 0, nil, 0, nil,
		1, []vk.ImageMemoryBarrier{srcBarrier},
	)

	// Copy image to staging buffer.
	region := vk.BufferImageCopy{
		BufferOffset:      0,
		BufferRowLength:   0,
		BufferImageHeight: 0,
		ImageSubresource: vk.ImageSubresourceLayers{
			AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
			MipLevel:       0,
			BaseArrayLayer: 0,
			LayerCount:     1,
		},
		ImageOffset: vk.Offset3D{X: 0, Y: 0, Z: 0},
		ImageExtent: vk.Extent3D{
			Width:  uint32(t.width),
			Height: uint32(t.height),
			Depth:  1,
		},
	}

	vk.CmdCopyImageToBuffer(
		commandBuffer,
		t.img, vk.ImageLayoutTransferSrcOptimal,
		stagingBuffer,
		1, []vk.BufferImageCopy{region},
	)

	// Transition image back: TRANSFER_SRC_OPTIMAL → SHADER_READ_ONLY_OPTIMAL
	// (in case it is accessed again before destruction).
	dstBarrier := vk.ImageMemoryBarrier{
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
	}

	vk.CmdPipelineBarrier(
		commandBuffer,
		vk.PipelineStageFlags(vk.PipelineStageTransferBit),
		vk.PipelineStageFlags(vk.PipelineStageFragmentShaderBit),
		0, 0, nil, 0, nil,
		1, []vk.ImageMemoryBarrier{dstBarrier},
	)

	r.EndSingleTimeCommands(commandBuffer)

	// Map staging buffer and copy data to hostBackingBuffer.
	var data unsafe.Pointer
	if err := vk.Error(vk.MapMemory(r.device, stagingBufferMemory, 0, vk.DeviceSize(vk.WholeSize), 0, &data)); err != nil {
		vk.DestroyBuffer(r.device, stagingBuffer, nil)
		vk.FreeMemory(r.device, stagingBufferMemory, nil)
		return fmt.Errorf("evictDeviceLocal: vk.MapMemory failed: %w", err)
	}

	t.hostBackingBuffer = make([]byte, totalSize)
	copy(t.hostBackingBuffer, unsafe.Slice((*byte)(data), totalSize))

	vk.UnmapMemory(r.device, stagingBufferMemory)

	// Destroy staging buffer.
	vk.DestroyBuffer(r.device, stagingBuffer, nil)
	vk.FreeMemory(r.device, stagingBufferMemory, nil)

	// Destroy GPU resources. If the texture hasn't been used in the last 2
	// frames, the GPU is done — destroy immediately. Otherwise, defer.
	r.destroyOrDeferImageResources(t.lastUsedFrame, t.img, t.imageView, t.allocation)

	t.img = nil
	t.imageView = nil
	t.allocation = nil

	// hostBackingBuffer is kept alive for TouchTexture reload.
	t.residency = ResidentEvicted
	t.needsReload = true
	if vkDebug {
		LogMessage("[VRAM] evict texture %dx%d (was DeviceLocal, %d bytes to host)", t.width, t.height, totalSize)
	}

	return nil
}

// evictHostVisible delegates to TransitionToSwappedOut for the copy-and-destroy,
// then marks the texture as evicted.
func (r *Renderer_VK) evictHostVisible(t *Texture_VK) error {
	if err := r.TransitionToSwappedOut(t); err != nil {
		return fmt.Errorf("evictHostVisible: %w", err)
	}

	// hostBackingBuffer is already populated by TransitionToSwappedOut.
	t.residency = ResidentEvicted
	t.needsReload = true
	if vkDebug {
		LogMessage("[VRAM] evict texture %dx%d (was HostVisible)", t.width, t.height)
	}

	return nil
}

// swapInAsHostVisible creates a HOST_VISIBLE image and uploads data from
// hostBackingBuffer. Used as a fallback when DEVICE_LOCAL allocation fails
// during swap-in, ensuring the texture always has valid GPU resources.
func (r *Renderer_VK) swapInAsHostVisible(t *Texture_VK) error {
	if t.hostBackingBuffer == nil {
		return fmt.Errorf("swapInAsHostVisible: hostBackingBuffer is nil")
	}

	format := t.MapInternalFormat(Max(t.depth, 8))
	layerCount := uint32(1)
	usage := vk.ImageUsageFlags(vk.ImageUsageTransferSrcBit | vk.ImageUsageTransferDstBit | vk.ImageUsageSampledBit)

	// Create HOST_VISIBLE image with LINEAR tiling.
	newImg := r.CreateImage(
		uint32(t.width), uint32(t.height), format,
		t.mipLevels, layerCount, usage,
		1, vk.ImageTilingLinear, false,
	)

	// Allocate HOST_VISIBLE memory.
	newAlloc, err := r.allocator.AllocateImageMemoryWithFallback(
		newImg,
		vk.MemoryPropertyHostVisibleBit|vk.MemoryPropertyHostCoherentBit,
		0, // no further fallback
	)
	if err != nil {
		vk.DestroyImage(r.device, newImg, nil)
		return fmt.Errorf("swapInAsHostVisible: HOST_VISIBLE allocation failed: %w", err)
	}

	// Upload data via direct map+memcpy (HOST_VISIBLE path).
	commandBuffer := r.BeginSingleTimeCommands()

	// Transition new image: UNDEFINED → GENERAL (LINEAR tiling requires GENERAL).
	preBarrier := vk.ImageMemoryBarrier{
		SType:               vk.StructureTypeImageMemoryBarrier,
		OldLayout:           vk.ImageLayoutUndefined,
		NewLayout:           vk.ImageLayoutGeneral,
		SrcAccessMask:       vk.AccessFlags(vk.AccessNone),
		DstAccessMask:       vk.AccessFlags(vk.AccessHostWriteBit | vk.AccessTransferWriteBit),
		SrcQueueFamilyIndex: vk.QueueFamilyIgnored,
		DstQueueFamilyIndex: vk.QueueFamilyIgnored,
		Image:               newImg,
		SubresourceRange: vk.ImageSubresourceRange{
			AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
			BaseMipLevel:   0,
			LevelCount:     t.mipLevels,
			BaseArrayLayer: 0,
			LayerCount:     layerCount,
		},
	}
	vk.CmdPipelineBarrier(commandBuffer,
		vk.PipelineStageFlags(vk.PipelineStageTopOfPipeBit),
		vk.PipelineStageFlags(vk.PipelineStageHostBit|vk.PipelineStageTransferBit),
		0, 0, nil, 0, nil, 1, []vk.ImageMemoryBarrier{preBarrier})

	r.EndSingleTimeCommands(commandBuffer)

	// Map and copy data.
	var mappedData unsafe.Pointer
	if err := vk.Error(vk.MapMemory(r.device, newAlloc.deviceMemory, newAlloc.offset, vk.DeviceSize(vk.WholeSize), 0, &mappedData)); err != nil {
		r.allocator.FreeImageAllocation(newImg, newAlloc)
		vk.DestroyImage(r.device, newImg, nil)
		return fmt.Errorf("swapInAsHostVisible: vk.MapMemory failed: %w", err)
	}

	// Get subresource layout for row pitch.
	var subresourceLayout vk.SubresourceLayout
	vk.GetImageSubresourceLayout(r.device, newImg, &vk.ImageSubresource{
		AspectMask: vk.ImageAspectFlags(vk.ImageAspectColorBit),
		MipLevel:   0,
		ArrayLayer: 0,
	}, &subresourceLayout)

	rowPitch := int(subresourceLayout.RowPitch)
	bytesPerPixel := uint32(t.depth / 8)
	rowSize := int(t.width) * int(bytesPerPixel)
	srcSlice := t.hostBackingBuffer

	dstPtr := uintptr(mappedData) + uintptr(subresourceLayout.Offset)
	for y := int32(0); y < t.height; y++ {
		srcOffset := int(y) * rowSize
		dstRow := unsafe.Slice((*byte)(unsafe.Pointer(dstPtr)), rowPitch)
		copy(dstRow[:rowSize], srcSlice[srcOffset:srcOffset+rowSize])
		dstPtr += uintptr(rowPitch)
	}

	vk.UnmapMemory(r.device, newAlloc.deviceMemory)

	// Transition to GENERAL for shader sampling (LINEAR tiling).
	commandBuffer = r.BeginSingleTimeCommands()
	postBarrier := vk.ImageMemoryBarrier{
		SType:               vk.StructureTypeImageMemoryBarrier,
		OldLayout:           vk.ImageLayoutGeneral,
		NewLayout:           vk.ImageLayoutGeneral,
		SrcAccessMask:       vk.AccessFlags(vk.AccessHostWriteBit),
		DstAccessMask:       vk.AccessFlags(vk.AccessShaderReadBit),
		SrcQueueFamilyIndex: vk.QueueFamilyIgnored,
		DstQueueFamilyIndex: vk.QueueFamilyIgnored,
		Image:               newImg,
		SubresourceRange: vk.ImageSubresourceRange{
			AspectMask:     vk.ImageAspectFlags(vk.ImageAspectColorBit),
			BaseMipLevel:   0,
			LevelCount:     t.mipLevels,
			BaseArrayLayer: 0,
			LayerCount:     layerCount,
		},
	}
	vk.CmdPipelineBarrier(commandBuffer,
		vk.PipelineStageFlags(vk.PipelineStageHostBit),
		vk.PipelineStageFlags(vk.PipelineStageFragmentShaderBit),
		0, 0, nil, 0, nil, 1, []vk.ImageMemoryBarrier{postBarrier})

	r.EndSingleTimeCommands(commandBuffer)

	// Create image view.
	newImageView := r.CreateImageView(newImg, format, 0, t.mipLevels, layerCount, false)

	// Update texture fields.
	t.img = newImg
	t.imageView = newImageView
	t.allocation = newAlloc
	t.residency = ResidentHostVisible
	t.hostVisibleUploaded = true
	t.hostBackingBuffer = nil

	return nil
}

// CheckMemoryBudgetAndSwap monitors VRAM usage and triggers texture swap-out
// when usage exceeds the configured threshold.
//
// Two modes:
//   - VK_EXT_memory_budget: queries per-heap usage/budget from the driver each frame.
//   - Heuristic fallback: uses totalImageBytesAllocated / deviceLocalHeapSize.
//
// Respects cooldown (vramSwapCooldownFrames) to avoid swapping every frame.
// The budget target is calculated to free enough VRAM to get 10% headroom below
// the threshold.
func (r *Renderer_VK) CheckMemoryBudgetAndSwap() {
	if !r.vramFallbackEnabled {
		return
	}

	// Cooldown: don't swap every frame.
	if r.currentFrameNumber-r.lastSwapFrame < uint64(r.vramSwapCooldownFrames) {
		return
	}

	var usagePercent float64

	if r.budgetEnabled {
		// Use VK_EXT_memory_budget data queried in BeginFrame.
		heapIndex := r.deviceLocalHeapIndex
		if int(heapIndex) >= len(r.memoryBudget.HeapUsage) || int(heapIndex) >= len(r.memoryBudget.HeapBudget) {
			return
		}
		heapUsage := r.memoryBudget.HeapUsage[heapIndex]
		heapBudget := r.memoryBudget.HeapBudget[heapIndex]
		if heapBudget == 0 {
			return
		}
		usagePercent = float64(heapUsage) / float64(heapBudget) * 100.0
	} else {
		// Heuristic fallback: use tracked allocation total.
		if r.deviceLocalHeapSize == 0 {
			return
		}
		usagePercent = float64(r.totalImageBytesAllocated) / float64(r.deviceLocalHeapSize) * 100.0
	}

	if usagePercent <= float64(r.vramSwapThresholdPercent) {
		return
	}

	// Calculate budget target: free enough to get 10% headroom below threshold.
	overBudget := usagePercent - float64(r.vramSwapThresholdPercent)
	targetFreePercent := overBudget + 10.0
	budgetTarget := vk.DeviceSize(float64(r.deviceLocalHeapSize) * targetFreePercent / 100.0)

	if vkDebug {
		LogMessage("[VRAM] usage %.1f%% exceeds threshold %d%%, swapping out %d bytes",
			usagePercent, r.vramSwapThresholdPercent, budgetTarget)
	}

	if err := r.SwapOutTextures(budgetTarget); err != nil {
		if vkDebug {
			LogMessage("[VRAM] SwapOutTextures failed: %v", err)
		}
	}

	r.lastSwapFrame = r.currentFrameNumber
}
