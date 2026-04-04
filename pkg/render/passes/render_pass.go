package passes

import (
	mgl "github.com/go-gl/mathgl/mgl32"
	"github.com/illoprin/retro-fps-kit-go/pkg/render/pipeline"
	"github.com/illoprin/retro-fps-kit-go/pkg/render/rhi"
)

type PostProcessingPass interface {
	GetColor() *rhi.Texture
	GetResultFramebuffer() *rhi.Framebuffer
	GetConfig() interface{}
	GetName() string
	RenderPass(*pipeline.DeferredRenderResult)
	Use() bool
	ResizeCallback()
	Delete()
}

type HasProjection interface {
	SetProjectionMatrix(m mgl.Mat4)
}
