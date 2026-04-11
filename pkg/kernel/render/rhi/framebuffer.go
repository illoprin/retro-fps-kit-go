// Package rhi provides high-level abstractions over OpenGL objects,
// simplifying the development of cross-platform rendering pipelines.
//
// It implements the Render Hardware Interface (RHI) concept,
// shielding the user from low-level OpenGL state management and
// providing an idiomatic Go API for common graphics tasks.
//
// Author: illoprin
//
// 2026

package rhi

import (
	"unsafe"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/core/logger"
)

// Framebuffer object implementation
type Framebuffer struct {
	ID              uint32
	Width           int32
	Height          int32
	colorTextures   [](*Texture)
	depthTexture    *Texture
	HasDepth        bool
	HasDepthStencil bool
}

// NewFramebuffer creates new fbo object
func NewFramebuffer(width, height int32) *Framebuffer {
	var id uint32
	gl.GenFramebuffers(1, &id)

	fb := &Framebuffer{
		ID:            id,
		Width:         width,
		Height:        height,
		colorTextures: make([](*Texture), 0),
	}

	logger.Infof("fbo id=%d created with size %d %d", id, width, height)

	return fb
}

func (fb *Framebuffer) HasColor() bool {
	return len(fb.colorTextures) > 0
}

// NewColorAttachment generates new color attachment (bind before use)
func (fb *Framebuffer) NewColorAttachment(format TextureFormat, filter TextureFilter) {
	// create color attachment
	colorTex := NewFramebufferColorTexture(fb.Width, fb.Height, format, filter)
	// attach created texture
	gl.FramebufferTexture2D(
		gl.FRAMEBUFFER,
		gl.COLOR_ATTACHMENT0+uint32(len(fb.colorTextures)),
		gl.TEXTURE_2D, colorTex.ID,
		0,
	)

	logger.Infof("fbo id=%d color attachment created", fb.ID)

	// add color attachment to slice
	fb.colorTextures = append(fb.colorTextures, colorTex)
}

// SetDrawBuffers determines buffers to color drawing (bind before use)
func (fb *Framebuffer) SetDrawBuffers(colorAttachmentIndices []int) {

	if colorAttachmentIndices == nil {
		gl.DrawBuffer(gl.NONE)
		return
	}

	// create color attachments list
	attachmentsList := make([]uint32, len(colorAttachmentIndices))
	for i, index := range colorAttachmentIndices {
		attachmentsList[i] = gl.COLOR_ATTACHMENT0 + uint32(index)
	}
	size := int32(len(attachmentsList))

	gl.DrawBuffers(size, &attachmentsList[0])
}

// NewDepthStencilAttachment generates new depth stencil attachment (bind before use)
func (fb *Framebuffer) NewDepthStencilAttachment() {
	// create depth texture
	depthTex := NewFramebufferDepthTexture(fb.Width, fb.Height, FormatDepth24Stencil8)

	fb.depthTexture = depthTex
	fb.HasDepthStencil = true

	// attach created texture
	gl.FramebufferTexture2D(
		gl.FRAMEBUFFER,
		gl.DEPTH_STENCIL_ATTACHMENT,
		gl.TEXTURE_2D,
		depthTex.ID,
		0,
	)

	logger.Infof("fbo id=%d depth stencil attachment created", fb.ID)

}

// NewDepthAttachment generates new depth texture attachment (bind before use)
func (fb *Framebuffer) NewDepthAttachment() {
	// create depth texture
	depthTex := NewFramebufferDepthTexture(fb.Width, fb.Height, FormatDepth24)
	fb.depthTexture = depthTex
	fb.HasDepth = true

	// attach created texture
	gl.FramebufferTexture2D(
		gl.FRAMEBUFFER,
		gl.DEPTH_ATTACHMENT,
		gl.TEXTURE_2D,
		depthTex.ID,
		0,
	)

	logger.Infof("fbo id=%d depth attachment created", fb.ID)
}

// Check framebuffer completness (bind before use)
func (fb *Framebuffer) Check() bool {
	// check framebuffer completness
	gl.BindFramebuffer(gl.FRAMEBUFFER, fb.ID)
	status := gl.CheckFramebufferStatus(gl.FRAMEBUFFER)

	var statusStr string
	switch status {
	case gl.FRAMEBUFFER_COMPLETE:
		statusStr = "FRAMEBUFFER_COMPLETE"
	case gl.FRAMEBUFFER_INCOMPLETE_ATTACHMENT:
		statusStr = "FRAMEBUFFER_INCOMPLETE_ATTACHMENT"
	case gl.FRAMEBUFFER_INCOMPLETE_MISSING_ATTACHMENT:
		statusStr = "FRAMEBUFFER_INCOMPLETE_MISSING_ATTACHMENT"
	case gl.FRAMEBUFFER_INCOMPLETE_DRAW_BUFFER:
		statusStr = "FRAMEBUFFER_INCOMPLETE_DRAW_BUFFER"
	case gl.FRAMEBUFFER_UNSUPPORTED:
		statusStr = "FRAMEBUFFER_UNSUPPORTED"
	case gl.FRAMEBUFFER_INCOMPLETE_MULTISAMPLE:
		statusStr = "FRAMEBUFFER_INCOMPLETE_MULTISAMPLE"
	default:
		statusStr = "UNKNOWN_ERROR"
	}
	if status != gl.FRAMEBUFFER_COMPLETE {
		logger.Infof("fbo id=%d broken, current status is %s", fb.ID, statusStr)
		return false
	}
	return true
}

func (f *Framebuffer) Bind() {
	gl.BindFramebuffer(gl.FRAMEBUFFER, f.ID)
}

// Bind framebuffer for drawing (glViewport set)
func (fb *Framebuffer) BindForDrawing() {
	gl.BindFramebuffer(gl.FRAMEBUFFER, fb.ID)
	gl.Viewport(0, 0, fb.Width, fb.Height)
}

// Delete deletes framebuffer and related textures
func (fb *Framebuffer) Delete() {
	if len(fb.colorTextures) > 0 {
		for _, t := range fb.colorTextures {
			t.Delete()
		}
	}
	if fb.depthTexture != nil {
		fb.depthTexture.Delete()
	}
	if fb.ID > 0 {
		gl.DeleteFramebuffers(1, &fb.ID)
	}

	logger.Infof("fbo id=%d deleted", fb.ID)
}

// Resize framebuffer color attachments
func (fb *Framebuffer) Resize(width, height int32) {
	if (fb.Width == width && fb.Height == height) ||
		(width <= 0 || height <= 0) {
		return
	}

	fb.Width = width
	fb.Height = height

	if len(fb.colorTextures) > 0 {
		for _, t := range fb.colorTextures {
			t.Resize(width, height)
		}
	}
	if fb.depthTexture != nil {
		fb.depthTexture.Resize(width, height)
	}

	fb.Check()

	logger.Infof("fbo id=%d resized to %d %d", fb.ID, width, height)

}

// GetColorTexture get color texture by attachment index
func (fb *Framebuffer) GetColorTexture(index int) *Texture {
	if index >= len(fb.colorTextures) || index < 0 {
		return nil
	}
	return fb.colorTextures[index]
}

// BlitToScreen copy framebuffer data to another framebuffer
func (fb *Framebuffer) Blit(id uint32, dstWidth, dstHeight int32, filter TextureFilter) {
	gl.BindFramebuffer(gl.READ_FRAMEBUFFER, fb.ID)
	gl.BindFramebuffer(gl.DRAW_FRAMEBUFFER, id)
	gl.BlitFramebuffer(
		0, 0, fb.Width, fb.Height, // x, y, w, h
		0, 0, dstWidth, dstHeight, // x, y, w, h
		gl.COLOR_BUFFER_BIT, uint32(GetFilter(filter)),
	)
	gl.BindFramebuffer(gl.READ_FRAMEBUFFER, id)
}

// ReadPixels read framebuffer
// color attachment to buffer
func (fb *Framebuffer) ReadPixels(
	x, y, width, height int32,
	attachmentIndex uint32, f TextureFormat,
	destination unsafe.Pointer,
) {
	if width > fb.Width || height > fb.Height ||
		x < 0 || y < 0 ||
		int(attachmentIndex) >= len(fb.colorTextures) {
		return
	}
	dataType := GetTextureDataType(f)
	format := GetFormat(f)
	gl.BindFramebuffer(gl.READ_FRAMEBUFFER, fb.ID)
	gl.ReadBuffer(gl.COLOR_ATTACHMENT0 + attachmentIndex)
	gl.ReadPixels(x, y, width, height, format, dataType, destination)
}

// GetdepthTextureID returns id of depth texture
func (fb *Framebuffer) GetDepthTexture() *Texture {
	return fb.depthTexture
}
