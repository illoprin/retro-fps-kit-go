package post

import (
	mgl "github.com/go-gl/mathgl/mgl32"
	"github.com/illoprin/retro-fps-kit-go/pkg/core/window"
	"github.com/illoprin/retro-fps-kit-go/pkg/render/pipeline"
	"github.com/illoprin/retro-fps-kit-go/pkg/render/rhi"
)

type PassID string

// PassSharedResources - shared resources
// for all rendering passes
type PassSharedResources struct {
	ScreenConfig *window.ScreenConfig
	MeshQuad     *rhi.Mesh
}

// PostProcessingPass represents
// one screen space effect
type PostProcessingPass interface {
	GetColor() *rhi.Texture
	GetResultFramebuffer() *rhi.Framebuffer
	GetConfig() interface{}
	RenderPass(*pipeline.DeferredRenderResult)
	Use() bool
	ResizeCallback()
	Delete()
}

type HasProjection interface {
	SetProjectionMatrix(mgl.Mat4)
}

type HasDeltaTime interface {
	SetDeltaTime(float32)
}
