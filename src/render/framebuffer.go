package render

import (
	"fmt"

	"github.com/go-gl/gl/v4.1-core/gl"
)

// Framebuffer обёртка над OpenGL Framebuffer
type Framebuffer struct {
	ID              uint32
	Width           int32
	Height          int32
	ColorTexture    *Texture
	DepthTexture    *Texture
	HasColor        bool
	HasDepth        bool
	HasDepthStencil bool
}

// FramebufferConfig конфигурация для создания Framebuffer
type FramebufferConfig struct {
	Width           int32
	Height          int32
	ColorFormat     TextureFormat
	UseDepth        bool
	UseDepthStencil bool
	UseMultisample  bool
	Samples         int32
}

// DefaultFramebufferConfig возвращает конфигурацию по умолчанию
func DefaultFramebufferConfig(width, height int32) FramebufferConfig {
	return FramebufferConfig{
		Width:           width,
		Height:          height,
		ColorFormat:     FormatRGBA8,
		UseDepth:        true,
		UseDepthStencil: false,
		UseMultisample:  false,
		Samples:         4,
	}
}

// NewFramebuffer создаёт новый Framebuffer
func NewFramebuffer(config FramebufferConfig) (*Framebuffer, error) {
	var id uint32
	gl.GenFramebuffers(1, &id)
	if id == 0 {
		return nil, fmt.Errorf("failed to generate framebuffer")
	}

	fb := &Framebuffer{
		ID:     id,
		Width:  config.Width,
		Height: config.Height,
	}

	fb.Bind()

	// Создаём цветную текстуру
	if config.ColorFormat != 0 {
		colorTex, err := NewFramebufferColorTexture(config.Width, config.Height, config.ColorFormat)
		if err != nil {
			fb.Delete()
			return nil, fmt.Errorf("failed to create color texture: %v", err)
		}
		fb.ColorTexture = colorTex
		fb.HasColor = true

		if config.UseMultisample {
			// Для мультисэмплинга используем Renderbuffer
			var rbo uint32
			gl.GenRenderbuffers(1, &rbo)
			gl.BindRenderbuffer(gl.RENDERBUFFER, rbo)
			gl.RenderbufferStorageMultisample(gl.RENDERBUFFER, config.Samples, uint32(fb.ColorTexture.getInternalFormat()), config.Width, config.Height)
			gl.FramebufferRenderbuffer(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.RENDERBUFFER, rbo)
		} else {
			gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, colorTex.ID, 0)
		}
	}

	// Создаём текстуру глубины
	if config.UseDepthStencil {
		depthTex, err := NewFramebufferDepthTexture(config.Width, config.Height)
		if err != nil {
			fb.Delete()
			return nil, fmt.Errorf("failed to create depth texture: %v", err)
		}
		// Изменяем формат на Depth24Stencil8
		depthTex.Config.Format = FormatDepth24Stencil8
		depthTex.bind()
		depthTex.allocateStorage()
		depthTex.unbind()

		fb.DepthTexture = depthTex
		fb.HasDepthStencil = true

		if config.UseMultisample {
			var rbo uint32
			gl.GenRenderbuffers(1, &rbo)
			gl.BindRenderbuffer(gl.RENDERBUFFER, rbo)
			gl.RenderbufferStorageMultisample(gl.RENDERBUFFER, config.Samples, gl.DEPTH24_STENCIL8, config.Width, config.Height)
			gl.FramebufferRenderbuffer(gl.FRAMEBUFFER, gl.DEPTH_STENCIL_ATTACHMENT, gl.RENDERBUFFER, rbo)
		} else {
			gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.DEPTH_STENCIL_ATTACHMENT, gl.TEXTURE_2D, depthTex.ID, 0)
		}
	} else if config.UseDepth {
		depthTex, err := NewFramebufferDepthTexture(config.Width, config.Height)
		if err != nil {
			fb.Delete()
			return nil, fmt.Errorf("failed to create depth texture: %v", err)
		}
		fb.DepthTexture = depthTex
		fb.HasDepth = true

		if config.UseMultisample {
			var rbo uint32
			gl.GenRenderbuffers(1, &rbo)
			gl.BindRenderbuffer(gl.RENDERBUFFER, rbo)
			gl.RenderbufferStorageMultisample(gl.RENDERBUFFER, config.Samples, gl.DEPTH_COMPONENT32F, config.Width, config.Height)
			gl.FramebufferRenderbuffer(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, gl.RENDERBUFFER, rbo)
		} else {
			gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, gl.TEXTURE_2D, depthTex.ID, 0)
		}
	}

	// Проверяем завершённость Framebuffer
	if status := gl.CheckFramebufferStatus(gl.FRAMEBUFFER); status != gl.FRAMEBUFFER_COMPLETE {
		fb.Delete()
		return nil, fmt.Errorf("framebuffer is not complete: %d", status)
	}

	fb.Unbind()

	return fb, nil
}

// Bind привязывает Framebuffer
func (fb *Framebuffer) Bind() {
	gl.BindFramebuffer(gl.FRAMEBUFFER, fb.ID)
	gl.Viewport(0, 0, fb.Width, fb.Height)
}

// Unbind отвязывает Framebuffer (возвращает к экрану)
func (fb *Framebuffer) Unbind() {
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
}

// Delete удаляет Framebuffer и связанные текстуры
func (fb *Framebuffer) Delete() {
	if fb.ColorTexture != nil {
		fb.ColorTexture.Delete()
	}
	if fb.DepthTexture != nil {
		fb.DepthTexture.Delete()
	}
	gl.DeleteFramebuffers(1, &fb.ID)
}

// Resize изменяет размер Framebuffer
func (fb *Framebuffer) Resize(width, height int32) {
	if fb.Width == width && fb.Height == height {
		return
	}

	fb.Width = width
	fb.Height = height

	if fb.ColorTexture != nil {
		fb.ColorTexture.Resize(width, height)
	}
	if fb.DepthTexture != nil {
		fb.DepthTexture.Resize(width, height)
	}
}

// BlitToScreen копирует содержимое Framebuffer на экран
func (fb *Framebuffer) BlitToScreen() {
	gl.BindFramebuffer(gl.READ_FRAMEBUFFER, fb.ID)
	gl.BindFramebuffer(gl.DRAW_FRAMEBUFFER, 0)
	gl.BlitFramebuffer(0, 0, fb.Width, fb.Height, 0, 0, fb.Width, fb.Height, gl.COLOR_BUFFER_BIT, gl.LINEAR)
	gl.BindFramebuffer(gl.READ_FRAMEBUFFER, 0)
}

// GetColorTextureID возвращает ID цветной текстуры
func (fb *Framebuffer) GetColorTextureID() uint32 {
	if fb.ColorTexture != nil {
		return fb.ColorTexture.ID
	}
	return 0
}

// GetDepthTextureID возвращает ID текстуры глубины
func (fb *Framebuffer) GetDepthTextureID() uint32 {
	if fb.DepthTexture != nil {
		return fb.DepthTexture.ID
	}
	return 0
}
