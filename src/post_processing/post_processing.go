package postprocessing

import (
	"github.com/illoprin/retro-fps-kit-go/src/render"
	"github.com/illoprin/retro-fps-kit-go/src/renderers"
)

type PostProcessingPass interface {
	GetColor() *render.Texture
	GetResultFramebuffer() *render.Framebuffer
	GetConfig() interface{}
	GetName() string
	RenderPass(*renderers.DeferredRenderResult)
	Use() bool
	ResizeCallback()
	Delete()
}
