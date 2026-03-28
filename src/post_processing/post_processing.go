package postprocessing

import "github.com/illoprin/retro-fps-kit-go/src/render"

type PostProcessingPass interface {
	GetColor() *render.Texture
	GetConfig() interface{}
	GetName() string
	RenderPass([]*render.Texture)
	Use() bool
	ResizeCallback()
	Delete()
}
