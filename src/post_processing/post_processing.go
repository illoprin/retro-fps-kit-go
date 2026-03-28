package postprocessing

import "github.com/illoprin/retro-fps-kit-go/src/render"

type PostProcessingPass interface {
	GetColor() *render.Texture
	RenderPass([]*render.Texture)
	ResizeCallback()
	Delete()
}
