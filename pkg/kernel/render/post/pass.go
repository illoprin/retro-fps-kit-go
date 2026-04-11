package post

import (
	mgl "github.com/go-gl/mathgl/mgl32"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/core/window"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/render/pipeline"
	"github.com/illoprin/retro-fps-toolkit-go/pkg/kernel/render/rhi"
)

type PassID string

// PassSharedResources - shared resources
// for all rendering passes
type PassSharedResources struct {
	ScreenConfig *window.ScreenConfig
	MeshQuad     *rhi.Mesh
}

type DebugTexture struct {
	Name    string
	Texture *rhi.Texture
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

// DebuggablePass represents pass
// that can return your fbo color attachments
type DebuggablePass interface {
	GetDebugTextures() []DebugTexture
}

type HasProjection interface {
	SetProjectionMatrix(mgl.Mat4)
}

type HasDeltaTime interface {
	SetDeltaTime(float32)
}
