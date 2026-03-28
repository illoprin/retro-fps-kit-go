package postprocessing

import "github.com/illoprin/retro-fps-kit-go/src/render"

type ColorGradingConfig struct {
	Gamma      float32
	Exposure   float32
	Contrast   float32
	Saturation float32
	Brightness float32
}

type ColorGradingPass struct {
	cfg       *ColorGradingConfig
	program   *render.Program
	mesh      *render.Mesh
	fbo       *render.Framebuffer
	resources []render.Resource
	screenCfg *ScreenConfig
}

func NewColorGradingPass(
	screenCfg *ScreenConfig,
	quad *render.Mesh,
	cfg *ColorGradingConfig,
) (*ColorGradingPass, error) {
	// create color framebuffer
	return nil, nil
}

// get framebuffer color
func (cg *ColorGradingPass) GetColor() *render.Texture {
	return nil
}

// RenderPass texture index 0 - color
func (cg *ColorGradingPass) RenderPass(src []*render.Texture) {
	if len(src) < 1 {
		return
	}
}

func (cg *ColorGradingPass) ResizeCallback() {
}

func (p *ColorGradingPass) Delete() {
	for _, r := range p.resources {
		if r != nil {
			r.Delete()
		}
	}
}
