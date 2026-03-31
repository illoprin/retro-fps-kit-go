package render

import (
	"fmt"

	"github.com/go-gl/gl/v4.1-core/gl"
)

// Framebuffer object implementation
type Framebuffer struct {
	ID              uint32
	Width           int32
	Height          int32
	ColorTextures   [](*Texture)
	DepthTexture    *Texture
	HasDepth        bool
	HasDepthStencil bool
}

// NewFramebuffer создаёт новый Framebuffer
func NewFramebuffer(width, height int32) (*Framebuffer, error) {
	var id uint32
	gl.GenFramebuffers(1, &id)

	if id == 0 {
		return nil, fmt.Errorf("failed to create framebuffer")
	}

	fb := &Framebuffer{
		ID:            id,
		Width:         width,
		Height:        height,
		ColorTextures: make([](*Texture), 0),
	}

	return fb, nil
}

func (fb *Framebuffer) HasColor() bool {
	return len(fb.ColorTextures) > 0
}

// NewColorAttachment generates new color attachment (bind before use)
func (fb *Framebuffer) NewColorAttachment(colorFormat TextureFormat) error {
	// create color attachment
	colorTex, err := NewFramebufferColorTexture(fb.Width, fb.Height, colorFormat)
	if err != nil {
		return fmt.Errorf("failed to create color texture: %v", err)
	}
	// attach created texture
	gl.FramebufferTexture2D(
		gl.FRAMEBUFFER,
		gl.COLOR_ATTACHMENT0+uint32(len(fb.ColorTextures)),
		gl.TEXTURE_2D, colorTex.ID,
		0,
	)
	// add color attachment to slice
	fb.ColorTextures = append(fb.ColorTextures, colorTex)
	return nil
}

// SetDrawBuffers determines buffers to color drawing (bind before use)
func (fb *Framebuffer) SetDrawBuffers(colorAttachmentIndices []int) {

	// create color attachments list
	attachmentsList := make([]uint32, len(colorAttachmentIndices))
	for i, index := range colorAttachmentIndices {
		attachmentsList[i] = gl.COLOR_ATTACHMENT0 + uint32(index)
	}
	size := int32(len(attachmentsList))

	gl.DrawBuffers(size, &attachmentsList[0])
}

// NewDepthStencilAttachment generates new depth stencil attachment (bind before use)
func (fb *Framebuffer) NewDepthStencilAttachment() error {
	// create depth texture
	depthTex, err := NewFramebufferDepthTexture(fb.Width, fb.Height)
	if err != nil {
		return fmt.Errorf("failed to create depth stencil texture: %v", err)
	}
	// change format to Depth24Stencil8
	depthTex.Config.Format = FormatDepth24Stencil8
	depthTex.bind()
	depthTex.allocateStorage()
	depthTex.unbind()

	fb.DepthTexture = depthTex
	fb.HasDepthStencil = true

	// attach created texture
	gl.FramebufferTexture2D(
		gl.FRAMEBUFFER,
		gl.DEPTH_STENCIL_ATTACHMENT,
		gl.TEXTURE_2D,
		depthTex.ID,
		0,
	)
	return nil
}

// NewDepthAttachment generates new depth texture attachment (bind before use)
func (fb *Framebuffer) NewDepthAttachment() error {
	// create depth texture
	depthTex, err := NewFramebufferDepthTexture(fb.Width, fb.Height)
	if err != nil {
		return fmt.Errorf("failed to create depth texture: %v", err)
	}
	fb.DepthTexture = depthTex
	fb.HasDepth = true

	// attach created texture
	gl.FramebufferTexture2D(
		gl.FRAMEBUFFER,
		gl.DEPTH_ATTACHMENT,
		gl.TEXTURE_2D,
		depthTex.ID,
		0,
	)
	return nil
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
		fmt.Printf("fbo %d broken, current status is %s", fb.ID, statusStr)
		return false
	}
	return true
}

// Bind framebuffer
func (fb *Framebuffer) Bind() {
	gl.BindFramebuffer(gl.FRAMEBUFFER, fb.ID)
	gl.Viewport(0, 0, fb.Width, fb.Height)
}

// Unbind framebuffer (back to main opengl framebuffer)
func (fb *Framebuffer) Unbind() {
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
}

// Delete deletes framebuffer and related textures
func (fb *Framebuffer) Delete() {
	if len(fb.ColorTextures) > 0 {
		for _, t := range fb.ColorTextures {
			t.Delete()
		}
	}
	if fb.DepthTexture != nil {
		fb.DepthTexture.Delete()
	}
	if fb.ID > 0 {
		gl.DeleteFramebuffers(1, &fb.ID)
	}
}

// Resize framebuffer color attachments
func (fb *Framebuffer) Resize(width, height int32) {
	if (fb.Width == width && fb.Height == height) || (width <= 0 || height <= 0) {
		return
	}

	fb.Width = width
	fb.Height = height

	if len(fb.ColorTextures) > 0 {
		for _, t := range fb.ColorTextures {
			t.Resize(width, height)
		}
	}
	if fb.DepthTexture != nil {
		fb.DepthTexture.Resize(width, height)
	}

	fb.Check()
}

// BlitToScreen copy framebuffer data to another framebuffer
func (fb *Framebuffer) Blit(id uint32) {
	gl.BindFramebuffer(gl.READ_FRAMEBUFFER, fb.ID)
	gl.BindFramebuffer(gl.DRAW_FRAMEBUFFER, id)
	gl.BlitFramebuffer(0, 0, fb.Width, fb.Height, 0, 0, fb.Width, fb.Height, gl.COLOR_BUFFER_BIT, gl.LINEAR)
	gl.BindFramebuffer(gl.READ_FRAMEBUFFER, id)
}

// GetDepthTextureID returns id of depth texture
func (fb *Framebuffer) GetDepthTextureID() uint32 {
	if fb.DepthTexture != nil {
		return fb.DepthTexture.ID
	}
	return 0
}
